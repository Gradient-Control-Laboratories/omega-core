package codegen_test

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/internal/test/require"
)

func TestUiWrappersStripFromWGSL(t *testing.T) {
	src := `
import "vector";

center = uiPoint([1, 2, 3]);
xf = uiTransform([[1, 0, 0, 0], [0, 1, 0, 0], [0, 0, 1, 0], [3, 4, 5, 1]]);
path = uiSpline([xf, xf, xf]);

return length(vector.multiplyPoint(path[2], p) - center);
`
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(mockMgr))
	require.NoError(t, err)

	render, err := renderHelper(src, res.Program)
	require.NoError(t, err)
	require.False(t, strings.Contains(render.WGSL, "uiPoint"))
	require.False(t, strings.Contains(render.WGSL, "uiTransform"))
	require.False(t, strings.Contains(render.WGSL, "uiSpline"))
}

func TestAnnotationWrappersStripFromWGSL(t *testing.T) {
	src := `
import "vector";

@point
center = [1, 2, 3];

@transform
xf = [[1, 0, 0, 0], [0, 1, 0, 0], [0, 0, 1, 0], [3, 4, 5, 1]];

@spline
path = [xf, xf, xf];

return length(vector.multiplyPoint(path[2], p) - center);
`
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(mockMgr))
	require.NoError(t, err)

	render, err := renderHelper(src, res.Program)
	require.NoError(t, err)
	require.NotEmpty(t, render.WGSL)
	require.False(t, strings.Contains(render.WGSL, "uiPoint"))
	require.False(t, strings.Contains(render.WGSL, "uiTransform"))
	require.False(t, strings.Contains(render.WGSL, "uiSpline"))
}

func TestPromotedSplineWrappersStripFromWGSL(t *testing.T) {
	src := `
import "vector";

@spline
path = [
  [[1, 0, 0, 0], [0, 1, 0, 0], [0, 0, 1, 0], [1, 2, 3, 1]],
  [[1, 0, 0, 0], [0, 1, 0, 0], [0, 0, 1, 0], [4, 5, 6, 1]],
  [[1, 0, 0, 0], [0, 1, 0, 0], [0, 0, 1, 0], [7, 8, 9, 1]]
];

return length(vector.multiplyPoint(path[2], p));
`
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(mockMgr))
	require.NoError(t, err)

	render, err := renderHelper(src, res.Program)
	require.NoError(t, err)
	require.NotEmpty(t, render.WGSL)
	require.False(t, strings.Contains(render.WGSL, "uiPoint"))
	require.False(t, strings.Contains(render.WGSL, "uiTransform"))
	require.False(t, strings.Contains(render.WGSL, "uiSpline"))
}
