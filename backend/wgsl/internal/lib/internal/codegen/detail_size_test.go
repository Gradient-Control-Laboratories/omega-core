package codegen_test

import (
	"strings"
	"testing"

	assembler "gradientcontrol.com/omega/core/internal/assembler"
)

// TestDetailSize builds colours with fine detail of three sizes: none given,
// wood's fibre's 0.15 mm; a constant 2 mm; and one computed from the point.
// Each draws its fine detail below half its size, by the larger of the
// screen's pixel and the Detail slider's (gsl_detailPw), the constant folded
// into the bound, and the computed size is worked out ahead of the guard,
// where its test reads it, never inside the branch.
func TestDetailSize(t *testing.T) {
	const src = `
import "color";
plain = color.gray;
plain.detail = color.mixOklab(color.gray, color.white, 0.5 + 0.5 * sin(p.x * 40));
sized = color.gray;
sized.detail = color.mixOklab(color.gray, color.black, 0.5 + 0.5 * sin(p.y * 3));
sized.detailSize = 2;
grain = 0.5 + 0.25 * sin(p.z);
varied = color.gray;
varied.detail = color.mixOklab(color.gray, color.red, 0.5 + 0.5 * sin(p.x * 9));
varied.detailSize = grain;
a = transform.translate(sphere(10), [0, 0, 0]);
a.color = plain;
b = transform.translate(sphere(10), [30, 0, 0]);
b.color = sized;
c = transform.translate(sphere(10), [60, 0, 0]);
c.color = varied;
return a, b, c;
`
	res, err := assembler.Assemble([]string{src}, mockMgr)
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	out, err := renderHelper(src, res.Program)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	i := strings.Index(out.WGSL, "fn gsl_colorField(")
	if i < 0 {
		t.Fatalf("no colour field in:\n%s", out.WGSL)
	}
	field := out.WGSL[i:]
	field = field[:strings.Index(field, "\n}\n")]
	for _, bound := range []string{"if (max(gsl_pw, gsl_detailPw) < 0.075) {", "if (max(gsl_pw, gsl_detailPw) < 1.0) {", "if (max(gsl_pw, gsl_detailPw) < 0.5 * ("} {
		if !strings.Contains(field, bound) {
			t.Errorf("no guard %q in:\n%s", bound, field)
		}
	}
	// The computed size: whatever names its bound reads are declared before
	// the guard opens.
	guard := strings.Index(field, "if (max(gsl_pw, gsl_detailPw) < 0.5 * (")
	if guard < 0 {
		return
	}
	line := field[guard:]
	line = line[:strings.Index(line, "\n")]
	for _, name := range []string{"grain"} {
		if strings.Contains(line, name) && !strings.Contains(field[:guard], "let "+name) {
			t.Errorf("the guard reads %s but it is not declared ahead of it:\n%s", name, field)
		}
	}
	if !strings.Contains(line, "sin(p.z)") && !strings.Contains(line, "grain") {
		t.Errorf("the computed size is not the guard's bound:\n%s", line)
	}
}
