package codegen_test

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
)

// TestSimpleNested builds a colour whose stand-in has a stand-in of its own:
// mottled dots stand in for plain polka dots, which the library stands in for
// with their flat mix. The mottled dots switch to the plain ones, and the
// plain ones to their mix, both by the Textures weight. A single-level
// stand-in, the checkerboard's, switches by it too, and there is no other.
func TestSimpleNested(t *testing.T) {
	const src = `
import "material/playful";
import "color";
dots = playful.polkaDot(4, 1, color.white, color.black);
mottled = color.mixOklab(dots, color.gray, 0.3 + 0.3 * sin(p.x * 7));
mottled.simple = playful.polkaDot(4, 1, color.white, color.black);
ball = transform.translate(sphere(10), [0, 0, 0]);
ball.color = mottled;
board = transform.translate(sphere(10), [30, 0, 0]);
board.color = playful.checkerboard(4, color.white, color.black);
return ball, board;
`
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(mockMgr))
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	out, err := renderHelper(src, res.Program)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	body := func(name string) string {
		i := strings.Index(out.WGSL, name)
		if i < 0 {
			t.Fatalf("no %s in:\n%s", name, out.WGSL)
		}
		f := out.WGSL[i:]
		return f[:strings.Index(f, "\n}\n")]
	}
	color := body("fn gsl_colorField(")
	// The mottled dots: the plain dots stand in, by the Textures weight.
	mottled := color[:strings.Index(color, "let mottled")]
	if !strings.Contains(mottled, "= playful_polkaDot(") || !strings.Contains(mottled, "if (gsl_simpleDetail > 0.0) {") {
		t.Errorf("the mottled dots do not switch to the plain ones by the Textures weight:\n%s", color)
	}
	// The plain dots: their flat mix stands in, by the same weight.
	if !strings.Contains(body("fn playful_polkaDot("), "if (gsl_simpleDetail > 0.0) {") {
		t.Errorf("the plain dots do not switch to their mix by the Textures weight")
	}
	// The checkerboard, one level: by the same weight.
	board := color[strings.Index(color, "let mottled"):]
	if !strings.Contains(board, "if (gsl_simpleDetail > 0.0) {") {
		t.Errorf("the checkerboard does not switch by the Textures weight:\n%s", board)
	}
	if strings.Contains(out.WGSL, "gsl_simpleColorDetail") {
		t.Errorf("a colour has a weight of its own")
	}
}
