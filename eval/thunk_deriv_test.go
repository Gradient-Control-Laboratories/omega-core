//go:build cgo

package eval

import (
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/internal/source"
)

// TestThunkDerivative differentiates through a higher-order combiner end to
// end. omega's unionRoundExact takes its two sides as 0-argument SDF thunks and
// calls them inside its body — including under a remap (n12/n21) — which is the
// form the Aerospace Bracket example uses. The property Engine could not inline
// those calls, so any .dp/.color over such a shape failed with `cannot evaluate
// call to unknown function "sdf2"`.
//
// The check is that the property-computed .dp agrees with fidget's own autodiff
// gradient of the same field.
func TestThunkDerivative(t *testing.T) {
	srcMgr := source.NewLocalManager("../stdlib")

	src := `import . "omega";
n1 = [0, 0, 1];
n2 = normalize([1, 0, 1]);
u = unionRoundExact(
  () { return p.z - 1; },
  () { return dot(p, n2) - 1; },
  n1, n2);
return u, u.dp.x, u.dp.y, u.dp.z;
`
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(srcMgr))
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}

	// Points spanning both arms of unionRoundExact's ternary: well inside each
	// half-space (the sharp branch) and near the seam (the blended branch).
	xs := []float32{0, 3, -3, 0.5, 1.5}
	ys := []float32{0, 0.5, -0.5, 0, 0.25}
	zs := []float32{0, 2, 2, 0.9, 1.1}

	out, err := EvalGrad(res.Program, xs, ys, zs)
	if err != nil {
		t.Fatalf("eval: %v", err)
	}

	for i := range xs {
		field := out[0][i]
		got := [3]float32{out[1][i].Value, out[2][i].Value, out[3][i].Value}
		want := [3]float32{field.Dx, field.Dy, field.Dz}
		for k := 0; k < 3; k++ {
			if d := got[k] - want[k]; d > 1e-4 || d < -1e-4 {
				t.Errorf("at (%v,%v,%v) axis %d: .dp=%v fidget grad=%v",
					xs[i], ys[i], zs[i], k, got[k], want[k])
			}
		}
	}
}
