package fidget

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

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

func TestConvert(t *testing.T) {
	testFiles, err := filepath.Glob("testdata/*.gsl")
	if err != nil {
		t.Fatalf("Failed to find test data files: %v", err)
	}

	if len(testFiles) == 0 {
		t.Fatal("No test files found in parser/testdata/*.yaml")
	}

	for _, testFile := range testFiles {
		testDir := filepath.Dir(testFile)
		baseName := filepath.Base(testFile)
		testName := strings.TrimSuffix(baseName, filepath.Ext(baseName))

		if testName == "literal_to_param" {
			// tested separately
			continue
		}

		t.Run(testName, func(t *testing.T) {
			gslFile := filepath.Join(testDir, testName+".gsl")
			gslSrc, err := os.ReadFile(gslFile)
			require.NoError(t, err)

			res, err := assembler.Assemble([]string{string(gslSrc)}, assembler.SourceManager(&source.MockManager{}))
			require.NoError(t, err)

			expectedFile := filepath.Join(testDir, testName+".glsl")
			expectedOutput, err := os.ReadFile(expectedFile)
			expectedExists := true
			if err == nil {
				if string(expectedOutput) == "skip" {
					// TODO: remove this once all tests are passing
					t.Skip("Skipping test")
					return
				}
			} else if errors.Is(err, os.ErrNotExist) {
				expectedExists = false
			} else {
				t.Fatalf("Failed to read expected output: %v", err)
			}

			// Clear the global state
			Clear()

			// Create a new scope
			fgScope := NewScope(nil)
			fgScope.SetTime(0)

			if !expectedExists {
				// Mark skip so that the test will be re-run if fidget crashes
				require.NoError(t, os.WriteFile(expectedFile, []byte("skip"), 0644))
				time.Sleep(100 * time.Millisecond)
			}

			for _, dep := range res.Program.Dependencies {
				err = fgScope.EvaluatePackage(dep)
				if err != nil {
					t.Fatalf("Failed to evaluate dependency %s: %v", dep.Path, err)
				}
			}
			err = fgScope.EvaluatePackage(res.Program.Primary)
			if err != nil {
				t.Fatalf("Failed to evaluate package: %v", err)
			}
			treeIndices, f, err := fgScope.EvaluateReturns(res.Program.Primary.Files[len(res.Program.Primary.Files)-1].Returns)
			if err != nil {
				t.Fatalf("Failed to evaluate returns: %v", err)
			}

			require.Nil(t, f, "unexpected return")
			require.NotEmpty(t, treeIndices, "no returns")

			treeIndex := treeIndices[0]
			for i := 1; i < len(treeIndices); i++ {
				treeIndex = Min(treeIndex, treeIndices[i])
			}

			canvasIndex := BuildCanvas(treeIndex)
			actualOutput := CodegenGLSL(canvasIndex)

			if !expectedExists {
				// Remove the skip file so that the test will be re-run
				require.NoError(t, os.Remove(expectedFile))
			}

			actualFile := filepath.Join(testDir, testName+".actual.glsl")
			require.NoError(t, os.WriteFile(actualFile, []byte(strings.TrimSpace(actualOutput)), 0644))

			if expectedExists {
				if strings.TrimSpace(string(expectedOutput)) == actualOutput {
					// exact match
				} else {
					assertGLSLEquivalent(t, string(expectedOutput), actualOutput)
				}
			} else {
				require.NoError(t, os.WriteFile(expectedFile, []byte(strings.TrimSpace(actualOutput)), 0644))
			}
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
