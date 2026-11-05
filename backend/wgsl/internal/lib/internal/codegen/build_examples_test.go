package codegen_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	api "gradientcontrol.com/omega/core/backend/wgsl/internal/lib"
	"gradientcontrol.com/omega/core/backend/wgsl/internal/lib/internal/wgsltest"
	assembler "gradientcontrol.com/omega/core/internal/assembler"
	src_test "gradientcontrol.com/omega/core/internal/srctest"
)

func TestBuildExamples(t *testing.T) {
	src_test.RunAllSourceTests(t, "../../../../../../internal/testdata/example/*.gsl", testBuildExample)
}

func testBuildExample(t *testing.T, srcFile string) {
	if strings.HasSuffix(srcFile, "Barberpole.gsl") || strings.HasSuffix(srcFile, "Vibrating Plate.gsl") {
		t.Skip("Skipping audio example")
		return
	}
	src_test.RunSourceTest(t, srcFile, runTestBuildExample, checkTestBuildExample, true)
}

func runTestBuildExample(t *testing.T, srcFile string, src []byte) (*api.RenderResult, error) {
	// As the viewer's compile assembles them (webai's compile.go).
	res, err := assembler.Assemble([]string{string(src)}, mockMgr)
	if err != nil {
		return nil, err
	}
	return renderHelper(string(src), res.Program)
}

func checkTestBuildExample(t *testing.T, srcFile string, src []byte, result *api.RenderResult) error {
	if result == nil {
		return fmt.Errorf("no result returned")
	}
	baseName := filepath.Base(srcFile)
	testName := strings.TrimSuffix(baseName, filepath.Ext(baseName))
	wgslFilename := testName + ".wgsl"
	goldenDir := "testdata/example"
	if err := os.MkdirAll(goldenDir, 0755); err != nil {
		return err
	}
	goldenPath := filepath.Join(goldenDir, wgslFilename)
	wgsltest.AssertWGSLEqual(t, goldenPath, result.WGSL, true)
	return nil
}
