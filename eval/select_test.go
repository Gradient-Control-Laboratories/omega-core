//go:build cgo

package eval_test

import (
	"math"
	"testing"

	"gradientcontrol.com/omega/core/eval"
	"gradientcontrol.com/omega/core/internal/source"
)

func TestSelectThreeBranches(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	code := `choose(v) { return v ? p.y + 10 : p.y + 20 : p.y + 30; }`
	groups, err := e.EvalGsl(t.Context(), nil, code, "", []string{"choose(p.x)", "choose(1)", "choose(0)", "choose(-1)"}, [][3]float64{{1, 2, 0}, {0, 2, 0}, {-1, 2, 0}})
	if err != nil {
		t.Fatal(err)
	}
	want := [][]float32{{12, 22, 32}, {12, 12, 12}, {22, 22, 22}, {32, 32, 32}}
	for i, group := range groups {
		for j, sample := range group.Samples[0] {
			if sample.Value != want[i][j] || sample.Dy != 1 {
				t.Fatalf("query %d point %d: %+v, want %g", i, j, sample, want[i][j])
			}
		}
	}
}

func TestRampOffsetCone(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	groups, err := e.EvalGsl(t.Context(), nil, "", "", []string{"rampOffset(length(p.xy),p.z,5,65,25,60)", "rampOffset(30,30,5,65,25,60)"}, [][3]float64{{30, 0, 30}})
	if err != nil {
		t.Fatal(err)
	}
	want := (30 - (25 + 25*35.0/60)) / math.Sqrt(1+35.0*35/(60*60))
	for _, group := range groups {
		if math.Abs(float64(group.Samples[0][0].Value)-want) > 1e-4 {
			t.Fatalf("cone: %+v, want %g", group, want)
		}
	}
}
