package gsleval_test

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/gsleval"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// TestEvalGslSpatialDeriv checks whether the .dp source-level derivative
// operator flows through the evalGsl query path unchanged: the query is
// spliced into the document and re-parsed through build.Build, which now runs
// property.Expand. For the sphere SDF s = |p| - 2, s.dp is the unit radial
// gradient, and it must also match the evaluator's own grad-tape gradient of s.
func TestEvalGslSpatialDeriv(t *testing.T) {
	e := gsleval.New(&source.MockManager{}, nil)

	pts := [][3]float64{{4, 0, 0}, {0, 2, 0}}
	results, err := e.EvalGsl(context.Background(), nil, testDoc, "",
		[]string{"s", "s.dp"}, pts)
	require.NoError(t, err)
	require.Equal(t, 2, len(results))

	// s.dp is a 3-vector: one group, three components.
	dp := results[1]
	require.Equal(t, 3, len(dp.Samples))

	// Unit radial gradients at the two points.
	want := [][3]float32{{1, 0, 0}, {0, 1, 0}}
	for pi := range pts {
		for c := 0; c < 3; c++ {
			got := dp.Samples[c][pi].Value
			if math.Abs(float64(got-want[pi][c])) > 1e-4 {
				t.Errorf("s.dp point %d comp %d = %v, want %v", pi, c, got, want[pi][c])
			}
		}
	}

	// Cross-check: .dp (source AD) must equal the grad-tape gradient of s.
	s := results[0]
	for pi := range pts {
		g := s.Samples[0][pi]
		grad := [3]float32{g.Dx, g.Dy, g.Dz}
		for c := 0; c < 3; c++ {
			if math.Abs(float64(dp.Samples[c][pi].Value-grad[c])) > 1e-4 {
				t.Errorf(".dp vs grad-tape point %d comp %d: %v vs %v",
					pi, c, dp.Samples[c][pi].Value, grad[c])
			}
		}
	}
}

// TestEvalGslTemporalDeriv checks whether the .dtime operator flows through the
// query path. A field translating along +x at unit rate w.r.t. the animation
// clock has d/dtime = -1 at a point ahead of it on the x axis.
func TestEvalGslTemporalDeriv(t *testing.T) {
	e := gsleval.New(&source.MockManager{}, nil)

	results, err := e.EvalGsl(context.Background(), nil, testDoc,
		"moved = length(p - [time, 0.0, 0.0]) - radius;",
		[]string{"moved.dtime"},
		[][3]float64{{4, 0, 0}})
	require.NoError(t, err)
	require.Equal(t, 1, len(results))
	got := results[0].Samples[0][0].Value
	if math.Abs(float64(got-(-1))) > 1e-4 {
		t.Errorf("moved.dtime = %v, want -1", got)
	}
}
