package codegen_test

import (
	"regexp"
	"strings"
	"testing"

	assembler "gradientcontrol.com/omega/core/internal/assembler"
)

// TestSimpleNested builds a colour whose stand-in has a stand-in of its own:
// mottled dots stand in for plain polka dots, which the library stands in for
// with their flat mix. The mottled dots switch to the plain ones, and the
// plain ones to their mix, both by the Textures weight. A single-level
// stand-in, the checkerboard's, switches by it too, and there is no other.
// The plain dots are the dot plane remapped into their direction's frame, so
// the shader holds them as specialised gsl_remap calls rather than a
// playful_polkaDot function.
func TestSimpleNested(t *testing.T) {
	const src = `
import "material/collection/playful";
import "color";
dots = playful.polkaDot(4, 0.25, [0, 0, 1], color.white, color.black);
mottled = color.mixOklab(dots, color.gray, 0.3 + 0.3 * sin(p.x * 7));
mottled.simple = playful.polkaDot(4, 0.25, [0, 0, 1], color.white, color.black);
ball = transform.translate(sphere(10), [0, 0, 0]);
ball.color = mottled;
board = transform.translate(sphere(10), [30, 0, 0]);
board.color = playful.checkerboard(4, color.white, color.black);
return ball, board;
`
	res, err := assembler.Assemble([]string{src}, mockMgr)
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
	// Before the mottled value: the plain dots and the stand-in's own dots,
	// each a gsl_remap call switched to its flat mix by the Textures weight,
	// then the mottled dots switched to the plain ones by it, reading the
	// dots' switched value. Three switches, no more.
	mottled := color[:strings.Index(color, "let mottled")]
	if strings.Count(mottled, "gsl_remap") < 2 || strings.Count(mottled, "if (gsl_simpleDetail > 0.0) {") != 3 {
		t.Errorf("the plain dots and the mottled dots do not each switch by the Textures weight:\n%s", color)
	}
	if !regexp.MustCompile(`let dots = gsl_simpleValue(_\d+)?;`).MatchString(mottled) {
		t.Errorf("the mottled dots do not read the plain dots' switched value:\n%s", color)
	}
	// The checkerboard, one level: by the same weight.
	board := color[strings.Index(color, "let mottled"):]
	if strings.Count(board, "if (gsl_simpleDetail > 0.0) {") != 1 {
		t.Errorf("the checkerboard does not switch by the Textures weight, once:\n%s", board)
	}
	if strings.Contains(out.WGSL, "gsl_simpleColorDetail") {
		t.Errorf("a colour has a weight of its own")
	}
}
