//go:build cgo

package eval

import (
	"context"
	"math"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/source"
)

// TestMapBendXZ checks map.bendXZ where a drawing of the bend would measure
// it: a plate on x = 0 lands on the arc it should, a rung across the plate
// lands on the arc's radius an arc length along, the sign of the curvature
// picks the side, and a curvature of 0 leaves the plate where it was.
func TestMapBendXZ(t *testing.T) {
	mgr := source.NewLocalManager("testpkg", "../stdlib")
	e := New(mgr, nil)
	ctx := context.Background()

	// A plate 2 thick on the plane x = 0, and a rung 1 thick across it at
	// z = 50. Bent to a radius of 100, the arc's centre is at x = +100.
	const doc = `
plate = abs(p.x) - 1;
rung = abs(p.z - 50) - 0.5;
bent = remap(plate, map.bendXZ(p, 1 / 100.0));
bentRung = remap(rung, map.bendXZ(p, 1 / 100.0));
bentBack = remap(plate, map.bendXZ(p, -1 / 100.0));
flat = remap(plate, map.bendXZ(p, 0));
plateAway = abs(p.x - 10) - 1;
moved = remap(plateAway, map.bendXZ(p - [10, 0, 20], 1 / 100.0) + [10, 0, 20]);
return bent;
`

	near := func(expr string, pt [3]float64, want, tol float64) {
		t.Helper()
		groups, err := e.EvalGsl(ctx, nil, doc, "", []string{expr}, [][3]float64{pt})
		if err != nil {
			t.Fatalf("eval %q: %v", expr, err)
		}
		got := float64(groups[0].Samples[0][0].Value)
		if math.IsNaN(got) || math.Abs(got-want) > tol {
			t.Errorf("%s at %v = %.5f, want %.4f±%.4f", expr, pt, got, want, tol)
		}
	}

	// The point of the bent space at `radius` from the arc's centre and
	// `angle` round it from the origin.
	onArc := func(radius, angle float64) [3]float64 {
		return [3]float64{100 - radius*math.Cos(angle), 0, radius * math.Sin(angle)}
	}

	// The plate's middle is the arc of radius 100, its faces the arcs of 99
	// and 101, all the way round to a radian either side.
	for _, angle := range []float64{-1, -0.5, 0, 0.3, 1} {
		near("bent", onArc(100, angle), -1, 1e-4)
		near("bent", onArc(99, angle), 0, 1e-4)
		near("bent", onArc(101, angle), 0, 1e-4)
	}
	// y passes through.
	near("bent", [3]float64{100 - 99*math.Cos(0.3), 37, 99 * math.Sin(0.3)}, 0, 1e-4)

	// The rung at z = 50 lies along the radius half a radian round, since
	// lengths along the plate are kept: at every depth through the plate.
	for _, radius := range []float64{95, 100, 105} {
		near("bentRung", onArc(radius, 0.5), -0.5, 1e-4)
	}
	// Half a millimetre along the plate either way is the rung's face.
	near("bentRung", onArc(100, 0.505), 0, 1e-4)
	near("bentRung", onArc(100, 0.495), 0, 1e-4)

	// A negative curvature bends the other way: the mirror image in x.
	back := onArc(99, 0.4)
	near("bentBack", [3]float64{-back[0], 0, back[2]}, 0, 1e-4)

	// No curvature is no bend, with nothing dividing by it.
	near("flat", [3]float64{1, 0, 80}, 0, 1e-4)
	near("flat", [3]float64{-3, 5, -40}, 2, 1e-4)

	// A plate through another point, bent about that point: the same arc,
	// moved there.
	at := onArc(101, 0.2)
	near("moved", [3]float64{at[0] + 10, 0, at[2] + 20}, 0, 1e-4)
}
