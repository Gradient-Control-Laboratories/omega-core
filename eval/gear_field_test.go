//go:build cgo

package eval

import (
	"context"
	"math"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/source"
)

// Shaft Modeling, Phase 20: the gear's drawn field is its profile's. Phase 19
// bounded the tips to the band above the root and the gaps to the band below
// the tip circle, and in each gap the field measured to the pitch circle and
// the tip circle rather than the flanks: closed contours on the pitch circle
// and a seam along the tip circle, up to 0.9 mm short. With the tips the whole
// profile and the gaps the whole space outside it, the field is the profile's,
// max(min(disc, f), f) = f, and pruned it is still the pitch disc.

// Drawn, gears.gearProfile is the profile before Phase 19 (frozen in
// oldGearProfileDoc) to float precision, across two teeth from below the root
// to above the tip, the gaps included.
func TestGearProfile_FieldIsTheProfiles(t *testing.T) {
	e := New(&source.MockManager{}, nil)
	pts := profilePoints()
	g, err := e.EvalGsl(context.Background(), nil, oldGearProfileDoc, "", []string{"newP", "oldP"}, pts)
	if err != nil {
		t.Fatal(err)
	}
	bad, worst, worstAt := 0, 0.0, 0
	for i := range pts {
		n, o := float64(g[0].Samples[0][i].Value), float64(g[1].Samples[0][i].Value)
		if d := math.Abs(n - o); d > 1e-4 {
			bad++
			if d > worst {
				worst, worstAt = d, i
			}
		}
	}
	if bad > 0 {
		p := pts[worstAt]
		t.Errorf("%d of %d points differ from the profile's field; worst %.4f at (%.3f, %.3f): new %g, old %g",
			bad, len(pts), worst, p[0], p[1], g[0].Samples[0][worstAt].Value, g[1].Samples[0][worstAt].Value)
	}
}
