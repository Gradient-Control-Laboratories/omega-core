package codegen_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/backend/wgsl/internal/codegen"
	"gradientcontrol.com/omega/core/backend/wgsl/internal/wgsltest"
	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/require"
)

func TestBuildUniform(t *testing.T) {
	testFile := "testdata/literal_to_param.gsl"
	testDir := filepath.Dir(testFile)
	baseName := filepath.Base(testFile)
	testName := strings.TrimSuffix(baseName, filepath.Ext(baseName))

	gslFile := filepath.Join(testDir, testName+".gsl")
	gslSrc, err := os.ReadFile(gslFile)
	require.NoError(t, err)

	res, err := assembler.Assemble([]string{string(gslSrc)}, assembler.SourceManager(&source.MockManager{}))
	require.NoError(t, err)

	wgslSrc, _, err := codegen.WGSL(res.Program)
	if err != nil {
		t.Log("Unexpected error:\n", err.Error())
		t.FailNow()
	}

	wgslFilename := testName + ".wgsl"
	wgslPath := filepath.Join("testdata", wgslFilename)
	wgsltest.AssertWGSLEqual(t, wgslPath, wgslSrc, true)
}
