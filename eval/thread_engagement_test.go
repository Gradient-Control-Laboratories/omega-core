//go:build cgo

package eval

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// engagementDoc is an M6 cap screw in a 12 mm plate tapped M6x1 right
// through, both at maximum material: the shank at the major diameter, the
// bore at the basic minor. They overlap as designed, across the annulus
// between the two over the plate's thickness, and each joint below differs
// from `ok` in one way that makes that overlap a collision.
const engagementDoc = `
screw = fastener.shcs(6.0, 20.0);
plate = max(box([-10.0, -10.0, -12.0], [10.0, 10.0, 0.0]),
            -drill.hole(drill.tap(6.0, 1.0, 14.0), 0.0));
plateZone = thread.femaleZone(6.0, 1.0, 0.0, -12.0);
ok = max(screw, plate, thread.engaged(thread.maleZone(6.0, 1.0, 0.0, -20.0), plateZone, 0.1));
shortThread = max(screw, plate, thread.engaged(thread.maleZone(6.0, 1.0, -8.0, -20.0), plateZone, 0.1));
big = fastener.shcs(8.0, 20.0);
wrongSize = max(big, plate, thread.engaged(thread.maleZone(8.0, 1.25, 0.0, -20.0), plateZone, 0.1));
bare = max(screw, plate);
nut = fastener.nut(3.0);
return ok;
`

// engagementRegion stops 0.3 below the seating plane: the head resting on the
// plate is contact of its own, and not what these proofs are about.
var engagementRegion = [6]float64{-8, -8, -14, 8, 8, -0.3}

// TestThreadEngagementProvesTheJoint: a screw in the hole tapped for it
// interferes, and the interference is valid exactly where both threads are
// cut. The joint proves; a thread that stops short of the plate leaves a
// plain shank in the tapped bore, an M8 goes into an M6 hole, and a joint
// with no engagement at all is just two solids overlapping — each is
// disproven, with the collision where it is.
func TestThreadEngagementProvesTheJoint(t *testing.T) {
	mgr := source.NewLocalManager("testpkg", "../stdlib")
	e := New(mgr, nil)
	ctx := context.Background()

	r, err := e.Certify(ctx, nil, engagementDoc, "", "ok", engagementRegion, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !r.Proven {
		t.Fatalf("the M6 joint should prove within its engagement, got %+v", r)
	}

	for _, c := range []struct {
		expr   string
		zLo    float64 // the collision's z range
		zHi    float64
		rLo    float64 // and its radial range
		rHi    float64
		reason string
	}{
		// Above z = -8 the screw has no thread, and its plain shank fills the
		// annulus the plate's thread was cut into.
		{"shortThread", -8.0, 0, 2.45, 3.0, "a plain shank in a tapped bore"},
		// An M8 shank reaches past the M6 hole's major diameter.
		{"wrongSize", -12.0, 0, 3.0, 4.0, "an M8 screw in an M6 hole"},
		{"bare", -12.0, 0, 2.45, 3.0, "no engagement allowed at all"},
	} {
		r, err := e.Certify(ctx, nil, engagementDoc, "", c.expr, engagementRegion, 0, 0)
		if err != nil {
			t.Fatal(err)
		}
		if !r.Disproven {
			t.Errorf("%s (%s) should be disproven, got %+v", c.expr, c.reason, r)
			continue
		}
		b := r.CounterBox
		z := (b[2] + b[5]) / 2
		rad := math.Hypot((b[0]+b[3])/2, (b[1]+b[4])/2)
		slack := boxRadius(b)
		if z < c.zLo-slack || z > c.zHi+slack || rad < c.rLo-slack || rad > c.rHi+slack {
			t.Errorf("%s: collision box %v (r %.3f, z %.3f) is not where %s puts it: r in [%.2f, %.2f], z in [%.1f, %.1f]",
				c.expr, b, rad, z, c.reason, c.rLo, c.rHi, c.zLo, c.zHi)
		}
	}
}

// TestTappedHoleAndNutAtMaximumMaterial: the smooth tapped hole and the nut's
// bore stand at the basic minor diameter, the material threading starts
// from, not at the major the thread reaches — M6x1 at 4.917, the nut's M3x0.5
// at 2.459.
func TestTappedHoleAndNutAtMaximumMaterial(t *testing.T) {
	mgr := source.NewLocalManager("testpkg", "../stdlib")
	e := New(mgr, nil)
	ctx := context.Background()

	at := func(expr string, pt [3]float64) float64 {
		t.Helper()
		groups, err := e.EvalGsl(ctx, nil, engagementDoc, "", []string{expr}, [][3]float64{pt})
		if err != nil {
			t.Fatalf("eval %q: %v", expr, err)
		}
		return float64(groups[0].Samples[0][0].Value)
	}
	minor6 := (6.0 - 2*0.5412659*1.0) / 2
	if v := at("plate", [3]float64{minor6, 0, -6}); math.Abs(v) > 1e-3 {
		t.Errorf("the M6x1 tapped bore's wall should be at r %.4f, the plate reads %.4f there", minor6, v)
	}
	minor3 := (3.0 - 2*0.5412659*0.5) / 2
	if v := at("nut", [3]float64{minor3, 0, 1.2}); math.Abs(v) > 1e-3 {
		t.Errorf("the M3 nut's bore should be at r %.4f, the nut reads %.4f there", minor3, v)
	}
}
