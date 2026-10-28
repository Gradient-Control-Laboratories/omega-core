package codegen_test

import (
	"testing"

	api "gradientcontrol.com/omega/core/backend/wgsl/internal/lib"
	assembler "gradientcontrol.com/omega/core/internal/assembler"
	src_test "gradientcontrol.com/omega/core/internal/srctest"
)

// TestFoldLoop builds testdata/fold_loop as the viewer's compile assembles,
// its folds over constants kept for loops (fold_loop.go).
func TestFoldLoop(t *testing.T) {
	src_test.RunAllSourceTests(t, "testdata/fold_loop/*.gsl", func(t *testing.T, srcFile string) {
		src_test.RunSourceTest(t, srcFile, runFoldLoop, checkTestBuildOne, true)
	})
}

func runFoldLoop(t *testing.T, srcFile string, src []byte) (*api.RenderResult, error) {
	res, err := assembler.Assemble([]string{string(src)}, mockMgr)
	if err != nil {
		return nil, err
	}
	return renderHelper(string(src), res.Program)
}
