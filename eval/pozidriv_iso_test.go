//go:build cgo

package eval

import (
	"context"
	"fmt"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// TestPozidrivGaugeSeatsWithinPenetration seats ISO 4757's type Z gauge
// (Table 4, read from IS 7478) in each ISO 7045 pan head's type Z recess and
// checks it against the penetration ISO 7045 gives for exactly that — a number
// in neither of the recess's inputs. At the MINIMUM the gauge has to be clear
// of the screw, proven by certify; by the MAXIMUM it has to have touched.
// Every size does both, with no slack.
func TestPozidrivGaugeSeatsWithinPenetration(t *testing.T) {
	mgr := source.NewLocalManager("testpkg", "../stdlib")
	e := New(mgr, nil)
	ctx := context.Background()

	// The gauge's wings run parallel to the recess's a few hundredths off,
	// over whole faces — worth the budget.
	e.SetCertifyBudget(20000000)

	for _, nd := range []float64{1.6, 2.0, 2.5, 3.0, 3.5, 4.0, 5.0, 6.0, 8.0, 10.0} {
		size := fmt.Sprintf("nd = %v;\nreturn sphere(1.0);\n", nd)
		groups, err := e.EvalGsl(ctx, nil, size, "", []string{
			"catalog.panPozidrivDia(nd)", "fastener.panPozidrivReference(nd)",
			"catalog.panPozidrivDepth(nd)", "catalog.panHeadHeight(nd)"}, [][3]float64{{0, 0, 0}})
		if err != nil {
			t.Fatalf("M%v: %v", nd, err)
		}
		m := float64(groups[0].Samples[0][0].Value)
		zRef := float64(groups[1].Samples[0][0].Value)
		penMax := float64(groups[2].Samples[0][0].Value)
		k := float64(groups[3].Samples[0][0].Value)
		h := m/2 + (k-zRef)*0.4986 + 0.3
		region := [6]float64{-h, -h, zRef - penMax - 0.8, h, h, k + 0.1}

		for _, deep := range []bool{false, true} {
			pen := "catalog.panPozidrivDepthMin(nd)"
			if deep {
				pen = "catalog.panPozidrivDepth(nd)"
			}
			doc := fmt.Sprintf(`
nd = %v;
pz = fastener.panPozidriv(nd, 10.0);
gauge = fastener.panPozidrivGauge(nd, %s);
return pz;
`, nd, pen)
			if !deep {
				r, err := e.Certify(ctx, nil, doc, "", "max(gauge, pz)", region, 0, 1e-4)
				if err != nil {
					t.Fatalf("M%v: %v", nd, err)
				}
				if !r.Proven {
					t.Errorf("M%v: gauge at the minimum penetration is not proven clear: disproven %v, %s, counter box %v",
						nd, r.Disproven, r.Reason, r.CounterBox)
				}
				continue
			}
			r, err := e.Extremum(ctx, nil, doc, "", "max(gauge, pz)", region, "min", 1e-3, 1)
			if err != nil {
				t.Fatalf("M%v: %v", nd, err)
			}
			if !(r.Min < 0) {
				t.Errorf("M%v: gauge has not touched by the maximum penetration: min %.4f at %v", nd, r.Min, r.MinAt)
			}
		}
	}
}
