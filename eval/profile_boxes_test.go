//go:build cgo

package eval

import (
	"context"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/source"
)

// A profile's box is exact: each edge's own box, its ends and, for an arc, the
// extremes of its circle that lie on it, for a Bezier its extremum along each
// axis. The reference box is the sampled boundary's, so this is the true box
// to within the sampling; the library pads a box by 0.01 mm each way.

// boxMargin is how far an evaluated box may sit from the reference's: the
// 0.01 mm padding and the sampling.
const boxMargin = 0.02

// evalBox is the named value's box in doc: [lo x, lo y, lo z, hi x, hi y, hi z].
func evalBox(t *testing.T, doc, name string) [6]float64 {
	t.Helper()
	e := New(&source.MockManager{}, nil)
	groups, err := e.EvalGsl(context.Background(), nil, doc, "", []string{name + ".box"}, [][3]float64{{0, 0, 0}})
	if err != nil {
		t.Fatalf("eval %s.box: %v", name, err)
	}
	var b [6]float64
	for k := range b {
		b[k] = float64(groups[0].Samples[k][0].Value)
	}
	return b
}

// checkBoxXY fails unless got's x and y bounds hold want's and sit within
// boxMargin of them.
func checkBoxXY(t *testing.T, label string, got [6]float64, want [4]float64) {
	t.Helper()
	for k, axis := range []string{"x", "y"} {
		lo, hi := got[k], got[k+3]
		if lo > want[k]+1e-6 || hi < want[k+2]-1e-6 {
			t.Errorf("%s: box in %s [%v, %v] misses the profile's [%v, %v]", label, axis, lo, hi, want[k], want[k+2])
		}
		if want[k]-lo > boxMargin || hi-want[k+2] > boxMargin {
			t.Errorf("%s: box in %s [%v, %v] is loose about the profile's [%v, %v] by %.3g and %.3g mm",
				label, axis, lo, hi, want[k], want[k+2], want[k]-lo, hi-want[k+2])
		}
	}
}

// Closed profiles of lines and arcs, a concave arc among them.
func TestProfileBoxes_ArcProfilesAreExact(t *testing.T) {
	for name, loop := range map[string]refLoop{"arc L": arcL(), "slot": slot(), "moulding": moulding(), "star": star()} {
		t.Run(name, func(t *testing.T) {
			ref := newRefProfile(t, loop)
			checkBoxXY(t, name, evalBox(t, "@scene(120)\nsolid = "+loop.gsl()+";\nreturn solid;\n", "solid"), ref.box())
		})
	}
}

// Closed profiles with Bezier edges.
func TestProfileBoxes_BezierProfilesAreExact(t *testing.T) {
	for _, c := range edgeLoopCases() {
		t.Run(c.name, func(t *testing.T) {
			ref, _ := c.loop.reference(t)
			checkBoxXY(t, c.name, evalBox(t, "@scene(120)\nsolid = "+c.loop.gsl()+";\nreturn solid;\n", "solid"), ref.box())
		})
	}
}

// A corner form's box is its loop's without the corners, which holds every
// fillet, chamfer and blend: exact for the polygon, and holding the filleted
// loop's own.
func TestProfileBoxes_CornerFormsAreTheirPolygons(t *testing.T) {
	corners := make([]refCorner, len(uShape))
	for i := range corners {
		corners[i] = refCorner{filletCorner, 3}
	}
	polygon := newRefProfile(t, refLoop{points: uShape})
	got := evalBox(t, "@scene(120)\nsolid = "+cornersExpr(uShape, corners)+";\nreturn solid;\n", "solid")
	checkBoxXY(t, "filleted U", got, polygon.box())
	filleted := newRefProfile(t, blendedLoop(uShape, corners)).box()
	if got[0] > filleted[0] || got[1] > filleted[1] || got[3] < filleted[2] || got[4] < filleted[3] {
		t.Errorf("box %v misses the filleted loop's %v", got, filleted)
	}
}

// An open sweep's box is its chain's exact box, without the rays, grown by the
// section's reach across and between its lowest and highest: here a
// rectangle 3 left and 4 right of the chain and 6 high.
func TestProfileBoxes_OpenSweepsAlongCurves(t *testing.T) {
	section := "profileXY([[-3, 0], [4, 0], [4, 6], [-3, 6]])"
	arcChain := openCases()[1].chain
	bezierChain := edgeChainCases()[0]
	cases := []struct {
		name, call string
		chain      [4]float64
	}{
		{"arcs", strings.Replace(arcChain.gsl(), "openProfileXY(", "sweepOpen(section, ", 1), newRefOpen(t, arcChain).chainBox()},
		{"Beziers", fmt.Sprintf("sweepOpenEdges(section, %s, %s)", pointsGSL(bezierChain.points), edgesGSL(bezierChain.edges)), bezierChain.reference(t).chainBox()},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := evalBox(t, fmt.Sprintf("@scene(140)\nsection = %s;\nsolid = %s;\nreturn solid;\n", section, c.call), "solid")
			want := [4]float64{c.chain[0] - 4, c.chain[1] - 4, c.chain[2] + 4, c.chain[3] + 4}
			checkBoxXY(t, c.name, got, want)
			if math.Abs(got[2]) > boxMargin || math.Abs(got[5]-6) > boxMargin {
				t.Errorf("%s: box in z [%v, %v], want [0, 6]", c.name, got[2], got[5])
			}
		})
	}
}
