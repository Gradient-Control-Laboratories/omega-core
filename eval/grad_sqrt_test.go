//go:build cgo

package eval_test

import (
	"math"
	"testing"

	"gradientcontrol.com/omega/core/eval"
	"gradientcontrol.com/omega/core/internal/source"
)

// The root of a sum of squared clamped terms is identically zero wherever every clamp is, so its derivative there
// is zero, not 0/0: fidget's gradient of sqrt takes it so, as omega's own derivative does. Deep inside a solid --
// where a shell's inner wall lies -- the exact box field and the round booleans keep a finite gradient, the face's.
func TestGradientFiniteWhereSqrtIsZero(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	queries := []string{
		"length(max(abs(p) - [20, 15, 10], 0)) + min(max(abs(p).x - 20, abs(p).y - 15, abs(p).z - 10), 0)",
		"intersectRound(p.x - 10, p.y - 10, 1)",
		"intersectRound(p.x - 10, p.y - 10, p.z - 10, 1)",
		"differenceRound(p.x - 10, 10 + p.y, 1)",
		"unionRound(p.x + 10, p.y + 10, 1)",
	}
	groups, err := e.EvalGsl(t.Context(), nil, "", "", queries, [][3]float64{{0, 0, 1}})
	if err != nil {
		t.Fatal(err)
	}
	want := []float64{-9, -10, -9, -10, 10}
	for i, group := range groups {
		s := group.Samples[0][0]
		for _, d := range []float32{s.Dx, s.Dy, s.Dz} {
			if math.IsNaN(float64(d)) || math.IsInf(float64(d), 0) {
				t.Errorf("%s: gradient (%g, %g, %g), want finite", queries[i], s.Dx, s.Dy, s.Dz)
				break
			}
		}
		if math.Abs(float64(s.Value)-want[i]) > 1e-4 {
			t.Errorf("%s = %g, want %g", queries[i], s.Value, want[i])
		}
	}
}
