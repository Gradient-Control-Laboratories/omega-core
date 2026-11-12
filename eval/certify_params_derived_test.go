//go:build cgo

package eval

import (
	"context"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// derivedSliderDoc reaches its slider only through a declaration that does
// not depend on p: grow = k * 2.0. The assembler folds that declaration at
// k's default, so unless the parametric fold keeps it dynamic, every use of
// grow is the constant 0 and the claim below holds at the default alone.
const derivedSliderDoc = `
k = 0.0;
k.min = 0.0;
k.max = 1.0;
grow = k * 2.0;
ball = length(p) - 1.0 - grow;
return ball;
`

var derivedSliderRegion = [6]float64{-1, -1, -1, 1, 1, 1}

// TestCertifyParamsDerivedDeclaration: ball >= -1.5 holds at k = 0 but fails
// at the origin for every k > 0.25, so over k in [0, 1] it must be disproven,
// with the counterexample at the high end of k, exactly as it is when the
// same expression reads k directly.
func TestCertifyParamsDerivedDeclaration(t *testing.T) {
	e := New(&source.MockManager{}, nil)
	for _, expr := range []string{"ball", "length(p) - 1.0 - k * 2.0"} {
		res, err := e.CertifyOverSliders(context.Background(), nil, derivedSliderDoc, "", expr, derivedSliderRegion, -1.5, 0)
		if err != nil {
			t.Fatal(err)
		}
		if res.Proven {
			t.Fatalf("%s: proved over k in [0, 1] a claim that fails for every k > 0.25: %+v", expr, res)
		}
		if !res.Disproven {
			t.Fatalf("%s: expected a disproof, got %+v", expr, res)
		}
		if cp := res.CounterParams["k"]; cp[0] < 0.25 {
			t.Errorf("%s: counterexample k range %v should lie above 0.25", expr, cp)
		}
	}
}
