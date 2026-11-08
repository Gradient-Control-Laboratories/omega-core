package gsleval_test

import (
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/gsleval"
	"gradientcontrol.com/omega/core/internal/source"
)

func TestVolumeBounds(t *testing.T) {
	e := gsleval.New(&source.MockManager{}, nil)
	for _, test := range []struct {
		expression string
		volume     float64
	}{
		{"length(p) - 5", 4 * math.Pi * 125 / 3},
		{"max(length(p) - 5, -(length(p) - 4))", 4 * math.Pi * (125 - 64) / 3},
		{"max(abs(p.x)-4, abs(p.y)-3, abs(p.z)-2)", 8 * 6 * 4},
	} {
		bounds, err := e.Volume(t.Context(), "", test.expression, [6]float64{-6, -6, -6, 6, 6, 6}, 0.02)
		if err != nil {
			t.Fatal(err)
		}
		if bounds.Lower > test.volume || bounds.Upper < test.volume || bounds.Upper-bounds.Lower > test.volume*0.05 {
			t.Fatalf("%s: %g not closely bracketed by %+v", test.expression, test.volume, bounds)
		}
	}
}
