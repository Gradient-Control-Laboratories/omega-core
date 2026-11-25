package codegen_test

import (
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/assembler"
)

// WGSL's grammar does not let && and || stand beside each other unbracketed,
// whatever their precedence; Tint refuses a || b && c though naga accepts it.
// So an && inside an || is bracketed, and an || inside an &&, as the GSL wrote
// them or not.
func TestLogicOps_MixedAreBracketed(t *testing.T) {
	cases := []struct{ gsl, want string }{
		{"inside = p.x > 0 || (p.y > 0 && p.z > 0);", "p.x > 0.0 || (p.y > 0.0 && p.z > 0.0)"},
		{"inside = (p.x > 0 && p.y > 0) || p.z > 0;", "(p.x > 0.0 && p.y > 0.0) || p.z > 0.0"},
		{"inside = p.x > 0 && (p.y > 0 || p.z > 0);", "p.x > 0.0 && (p.y > 0.0 || p.z > 0.0)"},
		{"inside = p.x > 0 || p.y > 0 || p.z > 0;", "p.x > 0.0 || p.y > 0.0 || p.z > 0.0"},
	}
	for _, c := range cases {
		src := "@scene(40)\n" + c.gsl + "\nsolid = inside ? length(p) - 3 : length(p) - 2;\nreturn solid;\n"
		res, err := assembler.Assemble([]string{src}, assembler.SourceManager(mockMgr))
		if err != nil {
			t.Fatalf("%s: assemble: %v", c.gsl, err)
		}
		out, err := renderHelper(src, res.Program)
		if err != nil {
			t.Fatalf("%s: render: %v", c.gsl, err)
		}
		if !strings.Contains(out.WGSL, "let inside = "+c.want+";") {
			t.Errorf("%s: want %q in the WGSL, got:\n%s", c.gsl, c.want, out.WGSL)
		}
	}
}
