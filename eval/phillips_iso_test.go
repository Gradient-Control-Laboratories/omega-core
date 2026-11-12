//go:build cgo

package eval

import (
	"context"
	"fmt"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// TestPhillipsWingThicknessFollows holds drill.phillipsAt to the one ISO 4757
// Table 1 dimension it does not take: the wing thickness f. The wings are what
// the four ribs leave between them, so f has to come back out of b, e, g and
// the angles — measured, as the drawing measures it, across the wing's end
// where the wing-end cone meets the bottom.
//
// It comes back within 0.02 of Table 1's range for every number: inside it for
// PH1, PH2 and PH4, and under it by 0.014 for PH0 and 0.016 for PH3. The slack
// is that, stated, and not room for a different construction — reading b as a
// wing width, the old model, gives a PH2 wing of 1.47.
func TestPhillipsWingThicknessFollows(t *testing.T) {
	mgr := source.NewLocalManager("testpkg", "../stdlib")
	e := New(mgr, nil)
	ctx := context.Background()

	const doc = `
return sphere(1.0);
`
	// Table 1's f, min and max, by recess number.
	table := [][2]float64{{0.31, 0.36}, {0.51, 0.56}, {0.66, 0.74}, {0.79, 0.86}, {1.19, 1.27}}
	const slack = 0.02
	const step = 0.001

	pts := make([][3]float64, 1000)
	for i := range pts {
		pts[i] = [3]float64{0, float64(i) * step, 0}
	}
	for no, want := range table {
		n := fmt.Sprintf("%d.0", no)
		// One rib's field with its ridge meeting the bottom at z = 0, sampled
		// across the wing on +x at the wing's end: x = g/2, at the height the
		// bottom reaches g.
		expr := "remap(drill.phillipsRibs(catalog.phillipsCore(" + n + "), catalog.phillipsFacet(" + n + "), " +
			"catalog.phillipsRidge(" + n + "), catalog.phillipsFlank(" + n + "), catalog.phillipsRoot(" + n + "), 0.0), " +
			"[p.x + catalog.phillipsPoint(" + n + ") / 2.0, p.y, " +
			"p.z + (catalog.phillipsPoint(" + n + ") - catalog.phillipsCore(" + n + ")) / 2.0 * tan(radians(28.0))])"
		groups, err := e.EvalGsl(ctx, nil, doc, "", []string{expr}, pts)
		if err != nil {
			t.Fatalf("PH%d: %v", no, err)
		}
		f := -1.0
		for i, s := range groups[0].Samples[0] {
			if s.Value < 0 {
				f = 2 * float64(i) * step
				break
			}
		}
		if f < want[0]-slack || f > want[1]+slack {
			t.Errorf("PH%d wing thickness f = %.3f, want %.2f to %.2f (within %.2f)", no, f, want[0], want[1], slack)
		}
	}
}

// TestPhillipsGaugeSeatsWithinPenetration seats ISO 4757's own type H gauge
// (Table 2) in each ISO 7045 pan head's recess and checks it against the one
// number ISO 7045 gives for doing exactly that: the penetration. The recess
// takes its form from ISO 4757 and its m from ISO 7045; the penetration is in
// neither of the inputs, so this is a check and not a restatement.
//
// At the MINIMUM penetration the gauge has to be clear of the screw — proven,
// by certify over the recess and the head around it. By the MAXIMUM it has to
// have touched. Every size touches inside its range but M3, which lands 0.024
// past its 1.8; the 0.03 on the maximum is that, stated.
func TestPhillipsGaugeSeatsWithinPenetration(t *testing.T) {
	mgr := source.NewLocalManager("testpkg", "../stdlib")
	e := New(mgr, nil)
	ctx := context.Background()

	// The gauge's flanks run parallel to the ribs', a few hundredths off, over
	// whole faces — the hardest thing certify is asked, and worth the budget.
	e.SetCertifyBudget(20000000)

	const pastMax = 0.03
	for _, nd := range []float64{1.6, 2.0, 2.5, 3.0, 3.5, 4.0, 5.0, 6.0, 8.0, 10.0} {
		size := fmt.Sprintf("nd = %v;\nreturn sphere(1.0);\n", nd)
		groups, err := e.EvalGsl(ctx, nil, size, "", []string{
			"catalog.panPhillipsDia(nd)", "fastener.panPhillipsReference(nd)",
			"catalog.panPhillipsDepth(nd)", "catalog.panHeadHeight(nd)"}, [][3]float64{{0, 0, 0}})
		if err != nil {
			t.Fatalf("M%v: %v", nd, err)
		}
		m := float64(groups[0].Samples[0][0].Value)
		zRef := float64(groups[1].Samples[0][0].Value)
		penMax := float64(groups[2].Samples[0][0].Value)
		k := float64(groups[3].Samples[0][0].Value)
		// The recess and the head around it, up to the top of the head: the
		// gauge's cone and the recess's are parallel above the reference plane.
		h := m/2 + (k-zRef)*0.4986 + 0.3
		region := [6]float64{-h, -h, zRef - penMax - 0.3, h, h, k + 0.1}

		for _, deep := range []bool{false, true} {
			pen := "catalog.panPhillipsDepthMin(nd)"
			if deep {
				pen = fmt.Sprintf("catalog.panPhillipsDepth(nd) + %v", pastMax)
			}
			doc := fmt.Sprintf(`
nd = %v;
ph = fastener.panPhillips(nd, 10.0);
gauge = fastener.panPhillipsGauge(nd, %s);
return ph;
`, nd, pen)
			if !deep {
				r, err := e.Certify(ctx, nil, doc, "", "max(gauge, ph)", region, 0, 1e-4)
				if err != nil {
					t.Fatalf("M%v: %v", nd, err)
				}
				if !r.Proven {
					t.Errorf("M%v: gauge at the minimum penetration is not proven clear: disproven %v, %s, counter box %v",
						nd, r.Disproven, r.Reason, r.CounterBox)
				}
				continue
			}
			r, err := e.Extremum(ctx, nil, doc, "", "max(gauge, ph)", region, "min", 1e-3, 1)
			if err != nil {
				t.Fatalf("M%v: %v", nd, err)
			}
			if !(r.Min < 0) {
				t.Errorf("M%v: gauge has not touched %.2f past the maximum penetration: min %.4f at %v",
					nd, pastMax, r.Min, r.MinAt)
			}
		}
	}
}
