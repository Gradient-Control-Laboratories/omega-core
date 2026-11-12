//go:build cgo

package eval

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/program/noise"
)

// TestNoiseHash3LiteralComponent evaluates noiseHash3 with a constant cell
// component, as the wood library's noiseHash3([columnX, columnY, 23]) has.
// The constant used to reach the noise oracle as a literal input, which fidget
// could not place in a register, and the panic took down the agent that asked.
// Each channel must match the Go reference hash of the same cell.
func TestNoiseHash3LiteralComponent(t *testing.T) {
	e := New(&source.MockManager{}, nil)
	prog := "h = noiseHash3([floor(p.x), floor(p.y), 23]);\nreturn h;"
	points := [][3]float64{{0.5, 0.5, 0}, {3.25, -2.5, 7}, {-40.5, 17.5, -3}}
	groups, err := e.EvalGsl(context.Background(), nil, prog, "", []string{"h"}, points)
	if err != nil {
		t.Fatalf("EvalGsl: %v", err)
	}
	if len(groups) != 1 || len(groups[0].Samples) != 3 {
		t.Fatalf("want one vec3 group, got %+v", groups)
	}
	for i, pt := range points {
		want := noise.Hash3(int(math.Floor(pt[0])), int(math.Floor(pt[1])), 23)
		for c := 0; c < 3; c++ {
			if got := groups[0].Samples[c][i].Value; math.Abs(float64(got-want[c])) > 1e-6 {
				t.Errorf("point %v channel %d: got %v, want %v", pt, c, got, want[c])
			}
		}
	}
}
