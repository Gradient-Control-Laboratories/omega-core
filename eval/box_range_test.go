//go:build cgo

package eval

import (
	"math"
	"testing"
)

// A slider declared with @range is a slider to .box, as one declared with .min
// and .max is (TestBoundingBox): its declared value is only where it starts, so
// used as a field it is not the positive constant it reads as, and a box that
// depends on it reads at its value. The box rule used to find sliders by their
// properties alone, so an @range variable was taken for a constant and
// min(a, r) was bounded by a, a box that is wrong once r moves.
func TestBoundingBox_RangeSlider(t *testing.T) {
	const u = unboundedAxis
	code := `
a = sphere([1, 2, 3], 0.5);
@range(0.1, 2.0) r = 0.5;
return a;
`
	cases := []struct {
		expr string
		want [6]float64
	}{
		{"sphere(r)", [6]float64{-0.5, -0.5, -0.5, 0.5, 0.5, 0.5}},
		{"min(a, r)", [6]float64{-u, -u, -u, u, u, u}},
		{"max(a, r)", [6]float64{0.5, 1.5, 2.5, 1.5, 2.5, 3.5}},
	}
	exprs := make([]string, len(cases))
	for i, c := range cases {
		exprs[i] = c.expr
	}
	got := boxEval(t, newBoxEvaluator(), code, exprs)
	for i, c := range cases {
		for k := 0; k < 6; k++ {
			if d := got[i][k] - c.want[k]; math.Abs(d) > 2e-5*math.Max(1, math.Abs(c.want[k])) {
				t.Errorf("%s.box = %v, want %v", c.expr, got[i], c.want)
				break
			}
		}
	}
}
