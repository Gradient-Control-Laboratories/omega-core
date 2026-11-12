//go:build cgo

package eval

import (
	"context"
	"fmt"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// TestFastenerJointsHaveNoSheet: where a screw is built from two solids — head
// on shank, thread on plain run, shoulder on stud, crown on flange, point on
// shank — the two used to be bounded at one plane from either side, so inside
// the metal their min read exactly zero across that plane: a surface of no
// thickness. Each joint is probed on its plane, well inside both pieces, and
// has to read inside.
func TestFastenerJointsHaveNoSheet(t *testing.T) {
	mgr := source.NewLocalManager("testpkg", "../stdlib")
	e := New(mgr, nil)
	ctx := context.Background()

	type probe struct {
		expr string
		pt   [3]float64
	}
	// A quarter of the diameter off the axis at 20 degrees, clear of every
	// drive's arms and inside the thread's minor diameter.
	off := func(d, z float64) [3]float64 {
		r := 0.25 * d
		return [3]float64{r * math.Cos(20*math.Pi/180), r * math.Sin(20*math.Pi/180), z}
	}
	var probes []probe
	for _, d := range []float64{3.0, 6.0} {
		for _, f := range []string{"shcs", "shcsThreaded", "shcsThreadedSin", "shcsTorx", "button", "hexScrew",
			"panPhillips", "panPozidriv", "panTorx"} {
			probes = append(probes, probe{fmt.Sprintf("fastener.%s(%v, 12.0)", f, d), off(d, 0)})
		}
	}
	probes = append(probes,
		probe{"fastener.panRobertson(0.164, 12.0)", off(4.17, 0)},
		probe{"fastener.partialShank(6.0, 1.0, 30.0, 12.0)", off(6, -18)},
		probe{"fastener.shoulderShank(8.0, 10.0, 6.0, 1.0, 10.0)", off(6, -10)},
		probe{"fastener.flangeHead(12.0, 5.0, 15.0, 1.5)", off(12, 1.5)},
		probe{"min(fastener.plainShank(6.0, 10.0), fastener.dogPoint(3.5, -10.0, 2.5))", off(3.5, -10)},
		probe{"min(fastener.plainShank(6.0, 10.0), fastener.softTip(4.9, -10.0, 1.5))", off(4.9, -10)},
		// Across the flutes, which run in the +-x half-planes: inside one, the
		// plane is the flute's own ceiling, a real face.
		probe{"min(fastener.plainShank(6.0, 10.0), fastener.selfDrillPoint(4.9, -10.0, 4.0, radians(118.0), 1.0))", [3]float64{0, 0.25 * 4.9, -10}},
	)

	const doc = "return sphere(1.0);\n"
	for _, pr := range probes {
		groups, err := e.EvalGsl(ctx, nil, doc, "", []string{pr.expr}, [][3]float64{pr.pt})
		if err != nil {
			t.Fatalf("%s: %v", pr.expr, err)
		}
		if got := float64(groups[0].Samples[0][0].Value); !(got < -0.05) {
			t.Errorf("%s at %v, on a joint plane inside the metal, reads %.4f; want inside by more than 0.05", pr.expr, pr.pt, got)
		}
	}
}
