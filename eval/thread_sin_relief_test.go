//go:build cgo

package eval

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// TestThreadSinRelief checks the cosmetic sinusoidal reliefs the way
// TestThreadRelief checks the ISO ones: subtracting the height from the smooth
// part must draw exactly the thread maleSin and femaleSin are fields of, so the
// displaced field agrees in sign with the true one wherever the true one is
// clear of its surface, and the height is zero beyond its bounds.
func TestThreadSinRelief(t *testing.T) {
	mgr := source.NewLocalManager("testpkg", "../stdlib")
	e := New(mgr, nil)
	ctx := context.Background()

	const doc = `
rel = thread.maleSinRelief(6.0, 1.0, 5.0, -5.0);
shaftDisp = (length(p.xy) - 3.0) - rel;
mal = thread.maleSin(6.0, 1.0);
partMin = (thread.minorDia(6.0, 1.0) / 2.0 - length(p.xy)) - thread.femaleSinRelief(6.0, 1.0, 5.0, -5.0);
voidMin = neg(thread.femaleSin(6.0, 1.0));
femMin = thread.femaleSinRelief(6.0, 1.0, 5.0, -5.0);
return mal;
`
	sample := func(expr string, pts [][3]float64) []float64 {
		t.Helper()
		groups, err := e.EvalGsl(ctx, nil, doc, "", []string{expr}, pts)
		if err != nil {
			t.Fatalf("eval %q: %v", expr, err)
		}
		out := make([]float64, len(pts))
		for i := range pts {
			out[i] = float64(groups[0].Samples[0][i].Value)
		}
		return out
	}

	// ON THE PROFILE: nothing on the crest, the whole form depth (0.5413 for a
	// pitch of 1) half a pitch along, and half of it a quarter pitch along.
	for _, c := range []struct {
		pt   [3]float64
		want float64
	}{
		{[3]float64{3.0, 0, 0}, 0},
		{[3]float64{2.4587, 0, 0.5}, -0.5412659},
		{[3]float64{2.7294, 0, 0.25}, -0.2706329},
	} {
		if got := sample("rel", [][3]float64{c.pt})[0]; math.Abs(got-c.want) > 1e-4 {
			t.Errorf("rel at %v = %.5f, want %.5f", c.pt, got, c.want)
		}
	}

	// THE SAME THREAD, external and internal.
	var lattice [][3]float64
	for i := 0; i < 12; i++ {
		for j := 0; j < 12; j++ {
			for k := 0; k < 12; k++ {
				r := 2.0 + 1.3*float64(i)/11
				th := 2 * math.Pi * float64(j) / 12
				z := -3.0 + 6.0*float64(k)/11
				lattice = append(lattice, [3]float64{r * math.Cos(th), r * math.Sin(th), z})
			}
		}
	}
	for _, pair := range [][2]string{{"shaftDisp", "mal"}, {"partMin", "voidMin"}} {
		disp, truth := sample(pair[0], lattice), sample(pair[1], lattice)
		n, bad := 0, 0
		var badAt [3]float64
		for i := range lattice {
			if math.Abs(truth[i]) < 2e-3 {
				continue
			}
			n++
			if (disp[i] < 0) != (truth[i] < 0) {
				bad++
				badAt = lattice[i]
			}
		}
		if bad > 0 {
			t.Errorf("%s disagrees in sign with %s at %d of %d points, e.g. %v",
				pair[0], pair[1], bad, n, badAt)
		}
		if n < len(lattice)/2 {
			t.Errorf("%s: only %d of %d lattice points were clear of the surface", pair[0], n, len(lattice))
		}
	}

	// BOUNDED, as the ISO reliefs are.
	for _, expr := range []string{"rel", "femMin"} {
		for _, pt := range [][3]float64{{2.6, 0, 5.2}, {2.6, 0, -5.2}, {4.2, 0, 0.5}, {2.6, 0, 5.0}} {
			if got := sample(expr, [][3]float64{pt})[0]; math.Abs(got) > 1e-6 {
				t.Errorf("%s at %v = %.6f, want 0 outside its bounds", expr, pt, got)
			}
		}
	}
}
