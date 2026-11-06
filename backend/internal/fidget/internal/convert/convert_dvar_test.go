package convert_test

import (
	"testing"

	"gradientcontrol.com/omega/core/backend/internal/fidget/internal/convert"
	"gradientcontrol.com/omega/core/backend/internal/fidget/internal/fidget"
	assembler "gradientcontrol.com/omega/core/internal/assembler"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// TestConvertDerivWrtVariable confirms the .d(var) operator — on a scalar
// slider through a dependent (h in terms of r), and on a literal vector
// variable, whose partials collect into a vec3 — is expanded by the autodiff
// pass and lowers to Fidget trees without error. Tree construction only, as
// TestConvertDerivatives: the surface speed field is not a solid to contour.
func TestConvertDerivWrtVariable(t *testing.T) {
	src := `r = 3.0;
h = 2 * r;
can = max(length(p.xy) - r, abs(p.z) - h / 2);
speed = -can.d(r) / length(can.dp);
c = [1, 2, 3];
g = length(p - c).d(c);
return can, speed, length(g);
`
	res, err := assembler.Assemble([]string{src}, &source.MockManager{})
	require.NoError(t, err)

	var treeCount int

	err = convert.ToTrees(res.Program, func(treeIndices []fidget.TreeIndex, _ []string) error {
		treeCount = len(treeIndices)
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 3, treeCount)
}
