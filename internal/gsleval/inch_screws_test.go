package gsleval

import (
	"context"
	"math"
	"strconv"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// TestInchRobertsonScrews checks ASME B18.6.3's square recessed machine screws
// against the tables they are read from: every inch key is a size the thread
// package has a Unified Coarse pitch for; a #6 takes a REGULAR depth #1 recess
// in a pan head and a SHORT one in a flat head, which is why the depth is a
// column of the product and not of the recess; and a #8 of each is where its
// dimensions put it — the pan head's side, its flat top and the square
// recess's wall, and the flat head's 82-degree cone.
func TestInchRobertsonScrews(t *testing.T) {
	mgr := source.NewLocalManager("testpkg", "../../stdlib")
	e := New(mgr, nil)
	ctx := context.Background()

	const doc = `
pan8 = fastener.panRobertson(0.164, 12.7);
flat8 = fastener.flatRobertson(0.164, 12.7);
return pan8;
`
	at := func(expr string, pt [3]float64) float64 {
		t.Helper()
		groups, err := e.EvalGsl(ctx, nil, doc, "", []string{expr}, [][3]float64{pt})
		if err != nil {
			t.Fatalf("eval %q: %v", expr, err)
		}
		return float64(groups[0].Samples[0][0].Value)
	}
	near := func(expr string, pt [3]float64, want, tol float64) {
		t.Helper()
		if got := at(expr, pt); math.IsNaN(got) || math.Abs(got-want) > tol {
			t.Errorf("%s at %v = %.4f, want %.4f±%.4f", expr, pt, got, want, tol)
		}
	}

	for _, size := range []float64{0.112, 0.138, 0.164, 0.19, 0.25} {
		for _, expr := range []string{"thread.uncPitch", "catalog.inchPanHeadDia", "catalog.inchFlatHeadDia"} {
			if v := at(expr+"("+ftoa4(size)+")", [3]float64{}); v <= 0 {
				t.Errorf("%s(%v) = %v: every inch screw size has to be one the thread and both heads carry", expr, size, v)
			}
		}
	}

	near("catalog.inchPanRecess(0.138)", [3]float64{}, 1, 1e-6)
	near("catalog.inchPanRecessDepth(0.138)", [3]float64{}, 2.692, 1e-3)  // 1R, .106"
	near("catalog.inchFlatRecessDepth(0.138)", [3]float64{}, 2.438, 1e-3) // 1S, .096"

	// #8 pan head: A .322" (8.179), H .120" (3.048), a #2 recess 2.840 across
	// flats, .127" (3.226) deep to the point of its pyramid.
	near("pan8", [3]float64{8.179 / 2, 0, 1.0}, 0, 1e-3)
	near("pan8", [3]float64{3.0, 0, 3.048}, 0, 1e-3)
	near("pan8", [3]float64{2.840 / 2, 0, 2.5}, 0, 1e-3)
	near("pan8", [3]float64{0, 0, 3.048 - 3.226 - 0.1}, -0.1, 1e-3)

	// #8 flat head: A .312" (7.925) at the face, the cone falling at 41 degrees
	// from the axis.
	near("flat8", [3]float64{7.925/2 - math.Tan(41*math.Pi/180), 0, -1.0}, 0, 1e-3)
}

func ftoa4(v float64) string {
	return strconv.FormatFloat(v, 'f', 4, 64)
}
