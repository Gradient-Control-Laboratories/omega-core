package codegen_test

import (
	"os"
	"regexp"
	"strings"
	"testing"

	assembler "gradientcontrol.com/omega/core/internal/assembler"
)

// TestDestructureTempsDistinct builds an example whose colour and material
// fields each inline several destructures declared in different GSL scopes:
// Material - Coated's chipped paint, destructured in the document, and the
// galvanized sheets' spangle cells, destructured inside the library's
// galvanized, whose stand-ins inline its body. The let a destructure emits is
// guarded by its temp's name, so two sharing a name emit one call and bind the
// other's fields to its struct: the shader then reads a material record as a
// vec3 and fails to validate, or, where the two structs match, silently takes
// the wrong values. Every function must declare each temp it reads, once.
func TestDestructureTempsDistinct(t *testing.T) {
	src, err := os.ReadFile("../../../../common/web/example/Material - Coated.gsl")
	if err != nil {
		t.Fatal(err)
	}
	res, err := assembler.Assemble([]string{string(src)}, mockMgr)
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	out, err := renderHelper(string(src), res.Program)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	checked := 0
	for _, fn := range strings.Split(out.WGSL, "\nfn ")[1:] {
		name := fn[:strings.Index(fn, "(")]
		declared := map[string]bool{}
		for _, m := range regexp.MustCompile(`(?:let|var) (gsl_result\w*)\b`).FindAllStringSubmatch(fn, -1) {
			if declared[m[1]] {
				t.Errorf("%s declares %s twice", name, m[1])
			}
			declared[m[1]] = true
		}
		for _, m := range regexp.MustCompile(`\b(gsl_result\w*)\.v_`).FindAllStringSubmatch(fn, -1) {
			if !declared[m[1]] {
				t.Errorf("%s reads %s, declared in another function or not at all", name, m[1])
			}
		}
		if name == "gsl_colorField" || name == "gsl_materialField" {
			checked++
			// Where two temps share a name every read still finds a
			// declaration, the wrong one; what shows is the call that
			// was never emitted. Each field needs the paint's call and
			// both sheets' spangle calls.
			if n := strings.Count(fn, "= coating_chippedPaint("); n != 1 {
				t.Errorf("%s emits the paint's call %d times, want 1", name, n)
			}
			if n := strings.Count(fn, "= noise_cellularCell(p, 1.0"); n != 2 {
				t.Errorf("%s emits %d spangle calls, want 2, one for each sheet", name, n)
			}
		}
	}
	if checked != 2 {
		t.Errorf("found %d of the colour and material fields, want both", checked)
	}
}
