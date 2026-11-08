package convert_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"gradientcontrol.com/omega/core/backend/internal/fidget/internal/convert"
	"gradientcontrol.com/omega/core/backend/internal/fidget/internal/fidget"
	"gradientcontrol.com/omega/core/backend/internal/fidget/internal/mesh"
	assembler "gradientcontrol.com/omega/core/internal/assembler"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

var mockMgr = source.NewLocalManager(
	"../../../../../stdlib",
)

func TestConvert(t *testing.T) {
	testFiles, err := filepath.Glob("testdata/*.gsl")
	if err != nil {
		t.Fatalf("Failed to find test data files: %v", err)
	}

	if len(testFiles) == 0 {
		t.Fatal("No test files *.gsl files found in testdata")
	}

	for _, testFile := range testFiles {
		testDir := filepath.Dir(testFile)
		baseName := filepath.Base(testFile)
		testName := strings.TrimSuffix(baseName, filepath.Ext(baseName))

		t.Run(testName, func(t *testing.T) {
			testOneConvert(t, testDir, testName)
		})
	}
}

type glslJSON struct {
	Inputs  map[string]interface{} `json:"inputs"`
	Helpers []string               `json:"helpers"`
	Glsl    []string               `json:"glsl"`
}

func normalizeGLSL(lines []string) []string {
	varMap := make(map[string]string)
	varCount := 0
	rVarRegex := regexp.MustCompile(`\br\d+\b`)
	normalized := make([]string, len(lines))
	for i, line := range lines {
		normalizedLine := rVarRegex.ReplaceAllStringFunc(line, func(match string) string {
			if _, ok := varMap[match]; !ok {
				varMap[match] = fmt.Sprintf("v%d", varCount)
				varCount++
			}
			return varMap[match]
		})
		normalized[i] = strings.TrimSpace(normalizedLine)
	}
	return normalized
}

func assertGLSLEquivalent(t *testing.T, expectedJSON, actualJSON string) {
	var exp, act glslJSON
	err := json.Unmarshal([]byte(expectedJSON), &exp)
	require.NoError(t, err)
	err = json.Unmarshal([]byte(actualJSON), &act)
	require.NoError(t, err)

	assert.Equal(t, exp.Inputs, act.Inputs)
	assert.Equal(t, exp.Helpers, act.Helpers)

	normExp := normalizeGLSL(exp.Glsl)
	normAct := normalizeGLSL(act.Glsl)
	assert.Equal(t, normExp, normAct)
}

func testOneConvert(t *testing.T, testDir, testName string) {
	gslFile := filepath.Join(testDir, testName+".gsl")
	gslSrc, err := os.ReadFile(gslFile)
	require.NoError(t, err)

	res, err := assembler.Assemble([]string{string(gslSrc)}, mockMgr)
	require.NoError(t, err)

	expectedFile := filepath.Join(testDir, testName+".glsl")
	expectedOutput, err := os.ReadFile(expectedFile)
	expectedExists := true
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			expectedExists = false
		} else {
			t.Fatalf("Failed to read expected output: %v", err)
		}
	}

	const skipTestFlag = "skip test"
	if string(expectedOutput) == skipTestFlag {
		t.Skip("Skipping test")
		return
	}
	if !expectedExists {
		// Mark skip so that the test will not be re-run if fidget crashes
		require.NoError(t, os.WriteFile(expectedFile, []byte(skipTestFlag), 0644))
		time.Sleep(100 * time.Millisecond)
	}

	var exitTestErr = errors.New("exit test")
	const skipComparisonFlag = "skip comparison"

	var actualOutput string
	// prog is already assembled above

	err = convert.ToTrees(res.Program,
		func(treeIndices []fidget.TreeIndex, _ []string) error {

			if string(expectedOutput) == skipComparisonFlag {
				t.Log("Skipping comparison")
				return exitTestErr
			}
			if !expectedExists {
				// Mark skip so that the test comparison will not be re-run if fidget crashes
				require.NoError(t, os.WriteFile(expectedFile, []byte(skipComparisonFlag), 0644))
				time.Sleep(100 * time.Millisecond)
			}

			treeIndex := treeIndices[0]
			for i := 1; i < len(treeIndices); i++ {
				treeIndex = fidget.Min(treeIndex, treeIndices[i])
			}

			canvasIndex := fidget.BuildCanvas(treeIndex)

			actualOutput = fidget.CodegenGLSL(canvasIndex)

			const skipStlFlag = "// skip stl"
			if strings.HasPrefix(string(gslSrc), skipStlFlag) {
				t.Log("Skipping stl generation")
				return nil
			}

			// Call GenerateSTL instead with each algorithm. The mesher takes a
			// world-space bbox + octree subdivision level; the old
			// center/scale [-5, 5]^3 region is the equivalent bbox below.
			bbox := [6]float32{-5, -5, -5, 5, 5, 5}
			_, err = mesh.GenerateSTL(treeIndices, 6, bbox, "marching-cubes",
				func(completed, total uint32, status string) { /* no-op */ })
			require.NoError(t, err)

			_, err = mesh.GenerateSTL(treeIndices, 6, bbox, "mdc",
				func(completed, total uint32, status string) { /* no-op */ })
			require.NoError(t, err)

			return nil
		})
	if err == exitTestErr {
		return
	}
	if err != nil {
		errMsg := "// ERROR: " + strings.TrimSpace(strings.Split(err.Error(), "\n")[0])
		firstLine := strings.TrimSpace(strings.Split(string(gslSrc), "\n")[0])
		if strings.HasPrefix(firstLine, errMsg) {
			t.Log("Error marked as expected: " + errMsg)
			return
		}

		t.Log("To mark error, add\n" + errMsg +
			"\n   to beginning of " + gslFile + "\n   found:\n" + firstLine)
		t.Errorf("unexpected error: %v", err)
		return
	}

	if !expectedExists {
		// Remove the skip file so that the test will be re-run
		require.NoError(t, os.Remove(expectedFile))
	}

	actualTrimmed := strings.TrimSpace(actualOutput)
	actualFile := filepath.Join(testDir, testName+".actual.glsl")
	require.NoError(t, os.WriteFile(actualFile, []byte(actualTrimmed), 0644))

	if expectedExists {
		expectedTrimmed := strings.TrimSpace(string(expectedOutput))
		if len(expectedTrimmed) != len(actualTrimmed) || expectedTrimmed != actualTrimmed {
			diff := len(actualTrimmed) - len(expectedTrimmed)
			var pct float64
			if len(expectedTrimmed) > 0 {
				pct = (float64(diff) / float64(len(expectedTrimmed))) * 100.0
			}
			t.Logf("golden size change for %s: %d B -> %d B (%+d B, %+.2f%%)",
				expectedFile, len(expectedTrimmed), len(actualTrimmed), diff, pct)
		}
		if os.Getenv("UPDATE_GOLDEN") != "" {
			require.NoError(t, os.WriteFile(expectedFile, []byte(actualTrimmed), 0644))
		} else if expectedTrimmed != actualTrimmed {
			assertGLSLEquivalent(t, string(expectedOutput), actualOutput)
		}
	} else {
		require.NoError(t, os.WriteFile(expectedFile, []byte(actualTrimmed), 0644))
	}
}

func TestConvert_UnprunedProgram(t *testing.T) {
	src := `
		unusedVar = 999.0;
		unusedFn() { return 123.0; }
		r = 10.0;
		return sphere(p, r);
	`
	res, err := assembler.Assemble([]string{src}, mockMgr)
	require.NoError(t, err)

	err = convert.ToTrees(res.Program, func(treeIndices []fidget.TreeIndex, outputNames []string) error {
		assert.Equal(t, 1, len(treeIndices))
		return nil
	})
	require.NoError(t, err)
}

func TestConvert_SingleElementList(t *testing.T) {
	src := `
		rad = [10.0];
		return sphere(p, rad[0]);
	`
	res, err := assembler.Assemble([]string{src}, mockMgr)
	require.NoError(t, err)

	err = convert.ToTrees(res.Program, func(treeIndices []fidget.TreeIndex, outputNames []string) error {
		assert.Equal(t, 1, len(treeIndices))
		return nil
	})
	require.NoError(t, err)
}
