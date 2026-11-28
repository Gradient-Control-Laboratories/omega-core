package wgsl_test

import (
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/assembler"
	"github.com/Gradient-Control-Laboratories/omega-core/backend/wgsl"
)

// The generic determinant and inverse reach the shader for a matrix that
// varies with p, pivot choices and all.
func TestMatrixGenericAtRunTime(t *testing.T) {
	src := `m = [[1, p.x * 0.1, 0], [p.y * 0.1, 1, 0], [0, 0, 1 + p.z * 0.01]];
inv = matrix.inverse(m);
return length(p) - 2 * matrix.determinant(m) - inv[0][0];`
	res, err := assembler.Assemble([]string{src})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := wgsl.Render(src, res.Program); err != nil {
		t.Fatal(err)
	}
}
