package codegen_test

import (
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/assembler"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/source"
)

// TestNonFiniteConstantIsAnError: a constant that folds to infinity or NaN, a
// division or a mod by zero, is reported where it is written, instead of
// reaching the shader as +Inf.0 or NaN.0, which WGSL does not accept.
func TestNonFiniteConstantIsAnError(t *testing.T) {
	for _, c := range []struct{ name, src, value string }{
		{"division by zero", "k = 0;\nreturn length(p) - 1 / k;\n", "+Inf"},
		{"mod by zero", "k = 0;\nreturn length(p) - mod(1, k);\n", "NaN"},
		{"in a list", "k = 0;\nv = [1, 1 / k];\nreturn length(p) - v.y;\n", "+Inf"},
	} {
		t.Run(c.name, func(t *testing.T) {
			res, err := assembler.Assemble([]string{c.src}, assembler.SourceManager(&source.MockManager{}))
			if err == nil {
				_, err = renderHelper(c.src, res.Program)
			}
			if err == nil || !strings.Contains(err.Error(), "evaluates to "+c.value) {
				t.Fatalf("want an error that the constant evaluates to %s, got %v", c.value, err)
			}
		})
	}
}
