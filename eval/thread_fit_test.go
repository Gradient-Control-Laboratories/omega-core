//go:build cgo

package eval

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// TestThreadToleranceFit: ISO 965-1's fundamental deviations for an M6x1, and
// what one does. A 6g screw is built at 6 + es, and its relieved surface is the
// basic one moved in by |es|/2 at every phase: a point |es|/2 inside the BASIC
// male surface is on the 6g one, and the basic surface itself — where a 6H
// hole's flank sits — is clear of it. The displaced field is only right ON its
// surface, not a distance off it, so the surface is what is checked.
func TestThreadToleranceFit(t *testing.T) {
	mgr := source.NewLocalManager("testpkg", "../stdlib")
	e := New(mgr, nil)
	ctx := context.Background()

	const doc = `
es = thread.deviationG(1.0);
shaft6g = (length(p.xy) - (6.0 + es) / 2.0) - thread.maleRelief(6.0 + es, 1.0, 5.0, -5.0);
return shaft6g;
`
	at := func(expr string, pt [3]float64) float64 {
		t.Helper()
		groups, err := e.EvalGsl(ctx, nil, doc, "", []string{expr}, [][3]float64{pt})
		if err != nil {
			t.Fatalf("eval %q: %v", expr, err)
		}
		return float64(groups[0].Samples[0][0].Value)
	}
	for expr, want := range map[string]float64{
		"thread.deviationE(1.0)":         -0.060,
		"thread.deviationF(1.0)":         -0.040,
		"thread.deviationG(1.0)":         -0.026,
		"thread.deviationH(1.0)":         0,
		"thread.internalDeviationG(1.0)": 0.026,
		"thread.internalDeviationH(1.0)": 0,
		"thread.deviationG(1.5)":         -0.032,
	} {
		if got := at(expr, [3]float64{}); math.Abs(got-want) > 1e-6 {
			t.Errorf("%s = %.4f, want %.4f", expr, got, want)
		}
	}

	// Mid-flank and on the root flat, the basic surface at radius r.
	for _, basic := range [][2]float64{
		{3.0 - 0.2706329, 0.21875},
		{3.0 - 0.5412659, 0.5},
	} {
		r, z := basic[0], basic[1]
		if got := at("shaft6g", [3]float64{r - 0.013, 0, z}); math.Abs(got) > 3e-4 {
			t.Errorf("a 6g M6 at r %.4f, |es|/2 inside the basic surface at z %.3f, reads %.5f; want it on its surface", r-0.013, z, got)
		}
		if got := at("shaft6g", [3]float64{r, 0, z}); got <= 0 {
			t.Errorf("a 6g M6 at r %.4f, z %.3f, on the basic surface, reads %.5f; want it clear", r, z, got)
		}
	}
}
