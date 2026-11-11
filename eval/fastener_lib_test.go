//go:build cgo

package eval

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// TestFastenerLibrary checks the fastener package end-to-end: catalog
// dimensions read back from the field at true ISO values, a screw fits its
// clearance bore with a certified margin, and the threaded display form
// stays within the smooth nominal.
func TestFastenerLibrary(t *testing.T) {
	mgr := source.NewLocalManager("testpkg", "../stdlib")
	e := New(mgr, nil)
	ctx := context.Background()

	const doc = `
s = fastener.shcs(3.0, 8.0);
st = fastener.shcsThreaded(3.0, 8.0);
csk = fastener.countersunk(4.0, 10.0);
n = fastener.nut(3.0);
w = fastener.washer(3.0);
pin = fastener.dowel(3.0, 12.0);
fit = max(s, -fastener.clearanceHole(3.0, fastener.fitMedium));
cskFit = max(csk, -fastener.countersunkHole(4.0, fastener.fitMedium));
return s;
`

	eval := func(expr string, pt [3]float64) float64 {
		t.Helper()
		groups, err := e.EvalGsl(ctx, nil, doc, "", []string{expr}, [][3]float64{pt})
		if err != nil {
			t.Fatalf("eval %q: %v", expr, err)
		}
		return float64(groups[0].Samples[0][0].Value)
	}
	near := func(expr string, pt [3]float64, want, tol float64) {
		t.Helper()
		got := eval(expr, pt)
		if math.IsNaN(got) || math.Abs(got-want) > tol {
			t.Errorf("%s at %v = %.4f, want %.3f±%.3f", expr, pt, got, want, tol)
		}
	}

	// ISO 4762 M3: head Ø5.5 (surface at r 2.75), shank Ø3 (surface at
	// r 1.5), hex socket across flats 2.5 (wall 1.25 from the axis),
	// socket floor ISO 4762's t min, 1.3, below the head top at z 3.
	near("s", [3]float64{2.75, 0, 1.5}, 0, 1e-3)
	near("s", [3]float64{1.5, 0, -4}, 0, 1e-3)
	near("s", [3]float64{0, 0, 2.95}, 1.25, 1e-3)
	near("s", [3]float64{0, 0, 1.5}, -0.2, 1e-3)
	// The screw sits in its Ø3.4 clearance bore with 0.2 radial margin.
	near("fit", [3]float64{1.5, 0, -4}, 0.2, 1e-3)
	// The threaded solid is contained in the smooth nominal (its field is
	// conservatively scaled, so compare containment by sign, not value):
	// nowhere inside the threaded form is outside the smooth form.
	for _, pt := range [][3]float64{{1.5, 0, -4}, {1.49, 0, -6.3}, {1.2, 0, -5}, {0, 0, -7.9}} {
		s, st := eval("s", pt), eval("st", pt)
		if st < -1e-3 && s > 1e-3 {
			t.Errorf("threaded form exceeds nominal at %v: smooth %.4f, threaded %.4f", pt, s, st)
		}
	}
	// And the thread genuinely cuts: at a root phase (z = -4.25 puts
	// cos(2πz/pitch) at -1 on the +x side), radius 1.45 is void in the
	// threaded form (root radius 1.35) while the smooth shank is solid.
	if s, st := eval("s", [3]float64{1.45, 0, -4.25}), eval("st", [3]float64{1.45, 0, -4.25}); s > -1e-3 || st < 1e-3 {
		t.Errorf("thread root not cut: smooth %.4f (want <0), threaded %.4f (want >0)", s, st)
	}
	// ISO 10642 M4 head: Ø8.96 rim flush at z 0; the head seats in its
	// countersink without interpenetration (contact by design, so ≥ -eps).
	near("csk", [3]float64{4.48, 0, 0}, 0, 2e-3)
	for _, pt := range [][3]float64{{4.48, 0, 0}, {3.0, 0, -1.4}, {2.0, 0, -5}, {2.25, 0, -5}} {
		if v := eval("cskFit", pt); v < -1e-3 {
			t.Errorf("countersunk screw interpenetrates its seat at %v: %.4f", pt, v)
		}
	}
	// ISO 4032 M3 nut: across flats 5.5, height 2.4; ISO 7089 M3 washer:
	// Ø7 × 0.5, bore Ø3.2; dowel Ø3 × 12 centered.
	near("n", [3]float64{2.75, 0, 1.2}, 0, 1e-3)
	near("n", [3]float64{0, 0, 1.2}, 1.2294, 1e-3)
	near("w", [3]float64{3.5, 0, 0.25}, 0, 1e-3)
	near("w", [3]float64{2.6, 0, 0.6}, 0.1, 1e-3)
	near("w", [3]float64{0, 0, 0.25}, 1.6, 1e-3)
	near("pin", [3]float64{1.5, 0, 0}, 0, 1e-3)
	near("pin", [3]float64{0, 0, 6.5}, 0.5, 1e-3)

	// Certified fit: along the whole shank, nowhere does the M3 screw come
	// within 0.16 of its clearance bore's wall (sound interval proof, not
	// a sample). The region stops below the seat plane — head-on-surface
	// contact at z = 0 is designed contact, not a clearance violation.
	r, err := e.Certify(ctx, nil, doc, "", "fit", [6]float64{-4, -4, -10, 4, 4, -0.5}, 0.08, 0.01)
	if err != nil {
		t.Fatalf("certify fit: %v", err)
	}
	if !r.Proven {
		t.Errorf("screw-in-bore fit not proven: disproven=%v indeterminate=%v (worst lower %.4f)",
			r.Disproven, r.Indeterminate, r.WorstLower)
	}
}
