package codegen_test

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/backend/audio/internal/codegen"
	"gradientcontrol.com/omega/core/backend/audio/internal/extract"
	assembler "gradientcontrol.com/omega/core/internal/assembler"
	"gradientcontrol.com/omega/core/internal/assembler/variation"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
	"gradientcontrol.com/omega/core/program"
)

func TestAudioToJs(t *testing.T) {
	testFiles, err := filepath.Glob("testdata/*.gsl")
	if err != nil {
		t.Fatalf("Failed to find test data files: %v", err)
	}

	if len(testFiles) == 0 {
		t.Fatal("No test files found in testdata/*.gsl")
	}

	for _, testFile := range testFiles {
		testDir := filepath.Dir(testFile)
		baseName := filepath.Base(testFile)
		testName := strings.TrimSuffix(baseName, filepath.Ext(baseName))

		t.Run(testName, func(t *testing.T) {
			gslFile := filepath.Join(testDir, testName+".gsl")
			gslSrc, err := os.ReadFile(gslFile)
			require.NoError(t, err)

			gslAudioSrc, _, err := extract.ExtractAudio(string(gslSrc))
			require.NoError(t, err)

			res, err := assembler.Assemble([]string{string(gslAudioSrc)}, &source.MockManager{})
			require.NoError(t, err)

			jsSrc, _, err := codegen.AudioToJs(res.Program)
			if err != nil {
				t.Log("Unexpected error:\n", err.Error())
				t.FailNow()
			}

			jsFilename := testName + ".js"
			jsExpectedSrc, err := os.ReadFile(filepath.Join("testdata", jsFilename))

			// Write output to file if it doesn't exist or UPDATE_GOLDEN is set
			if errors.Is(err, os.ErrNotExist) || os.Getenv("UPDATE_GOLDEN") != "" {
				os.WriteFile(filepath.Join("testdata", jsFilename), []byte(strings.TrimSpace(jsSrc)+"\n"), 0644)
			} else {
				require.NoError(t, err)
				assert.Equal(t, strings.TrimSpace(string(jsExpectedSrc)), strings.TrimSpace(jsSrc))
			}

			// Verify single-pass assembler audio variation slicing produces identical JS.
			singlePassRes, err := assembler.Assemble([]string{string(gslSrc)}, &source.MockManager{})
			require.NoError(t, err)
			require.True(t, len(singlePassRes.AudioReturns) > 0)
			audioOps := make([]*program.Operand, len(singlePassRes.AudioReturns))
			for i, ar := range singlePassRes.AudioReturns {
				audioOps[i] = ar.Return
			}
			audioProg := variation.Program(singlePassRes.Program, audioOps...)
			singlePassJs, _, err := codegen.AudioToJs(audioProg)
			require.NoError(t, err)
			assert.Equal(t, strings.TrimSpace(jsSrc), strings.TrimSpace(singlePassJs))
		})
	}
}

func TestAudioToJs_UnprunedProgram(t *testing.T) {
	src := `
		unusedVar = 999.0;
		unusedFn() { return 123.0; }
		sampleRate = 44100.0;
		return sin(time * 440.0);
	`
	res, err := assembler.Assemble([]string{src}, &source.MockManager{})
	require.NoError(t, err)

	jsSrc, _, err := codegen.AudioToJs(res.Program)
	require.NoError(t, err)
	assert.False(t, strings.Contains(jsSrc, "unusedVar"))
	assert.False(t, strings.Contains(jsSrc, "unusedFn"))
	assert.True(t, strings.Contains(jsSrc, "computeAudio"))
}
