package convert_test

import (
	"os"
	"testing"

	assembler "gradientcontrol.com/omega/core/internal/assembler"
	"gradientcontrol.com/omega/core/internal/fidget/internal/convert"
	"gradientcontrol.com/omega/core/internal/fidget/internal/fidget"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// Verifies the generated pen-stroke font builds through the fidget (STL) path:
// the branchless glyph(code) dispatch plus the map/reduce left() renderer all
// convert to a tree. (Meshing is skipped — this only checks conversion.)
func TestFontConverts(t *testing.T) {
	// Compile-time const-fold + per-iter specialization means only the four
	// glyphs in "HI42" actually emit, so using the full text font here is
	// still fast.
	const fontPath = "../../../../stdlib/text/text.gsl"
	font, err := os.ReadFile(fontPath)
	if err != nil {
		t.Skipf("font not generated at %s: %v", fontPath, err)
	}

	// A string literal at the origin, extruded to a solid plate.
	src := string(font) + "\nreturn max(left(\"HI42\", [0, 0], 1), abs(p.z) - 1);\n"

	res, err := assembler.Assemble([]string{src}, &source.MockManager{})
	require.NoError(t, err)

	var trees []fidget.TreeIndex

	err = convert.ToTrees(res.Program, func(t []fidget.TreeIndex, _ []string) error {
		trees = t
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, 1, len(trees))
}
