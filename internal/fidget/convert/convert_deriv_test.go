package convert_test

import (
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/internal/fidget"
	"gradientcontrol.com/omega/core/internal/fidget/convert"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// TestConvertDerivatives confirms the .dtime / .dp derivative operators are
// expanded by the autodiff pass and lower to Fidget trees without error. It
// drives tree construction only (ToTrees), not meshing: the analytic gradient
// of an SDF is singular at the body's center, which the volume mesher cannot
// contour — a property of analytic SDF gradients, not of tree construction.
func TestConvertDerivatives(t *testing.T) {
	src := `r = 2 + sin(time);
f = length(p) - r;
vel = f.dtime;
gradMag = length(f.dp);
return f, vel, gradMag;
`
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(&source.MockManager{}))
	require.NoError(t, err)

	var treeCount int

	err = convert.ToTrees(res.Program, func(treeIndices []fidget.TreeIndex, _ []string) error {
		treeCount = len(treeIndices)
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 3, treeCount)
}

// TestConvertCompanion confirms a function-internal derivative (a .dp on a
// parameter) lowers through Fidget: boundaryMap is replaced by a companion that
// receives the argument's gradient.
func TestConvertCompanion(t *testing.T) {
	src := `plane(c, n) {
  return dot(p - c, normalize(n));
}
boundaryMap(v) {
  return v * v.dp;
}
zero = [0, 0, 0];
a = p.x;
b = plane(zero, [1, 1, 0]);
ab = abs(remap(a, p - boundaryMap(b)));
body = abs(max(a, b));
return body, ab;
`
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(&source.MockManager{}))
	require.NoError(t, err)

	var treeCount int

	err = convert.ToTrees(res.Program, func(treeIndices []fidget.TreeIndex, _ []string) error {
		treeCount = len(treeIndices)
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 2, treeCount)
}
