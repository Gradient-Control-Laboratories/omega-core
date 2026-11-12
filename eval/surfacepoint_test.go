//go:build cgo

package eval_test

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/eval"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// TestSurfacePoint projects seed points onto the sphere SDF s = |p| - 2 from
// testDoc. A seed inside or outside should land on radius 2 with a unit radial
// outward normal — one Newton step suffices for an exact SDF.
func TestSurfacePoint(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)

	seeds := [][3]float64{{5, 0, 0}, {0, 0.5, 0}, {3, 4, 0}}
	samples, err := e.SurfacePoint(context.Background(), nil, testDoc, "", "s", seeds, 0, 0)
	require.NoError(t, err)
	require.Equal(t, len(seeds), len(samples))

	for i, s := range samples {
		if !s.Converged {
			t.Errorf("seed %d did not converge (residual %v)", i, s.Value)
		}
		r := math.Sqrt(s.Point[0]*s.Point[0] + s.Point[1]*s.Point[1] + s.Point[2]*s.Point[2])
		if math.Abs(r-2) > 1e-3 {
			t.Errorf("seed %d landed at radius %v, want 2", i, r)
		}
		// Outward normal is radial: point / |point|.
		for c := 0; c < 3; c++ {
			want := s.Point[c] / r
			if math.Abs(s.Normal[c]-want) > 1e-3 {
				t.Errorf("seed %d normal[%d] = %v, want %v", i, c, s.Normal[c], want)
			}
		}
	}
}
