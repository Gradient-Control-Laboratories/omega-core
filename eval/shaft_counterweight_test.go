//go:build cgo

package eval

import (
	"fmt"
	"math"
	"os"
	"strings"
	"testing"
)

// Shaft crank counterweights. shaft.web takes shaft.counterweight(swingR,
// balance, pinDia, pinLen) as a seventh argument: a sector standing out of the
// web's main circle on the side opposite the pin, in the web's own planes, its
// span the one that gives it `balance` times the first moment of the throw's
// pin side (this web's hull and half the pin) about the shaft's axis.
//
// Nothing here uses the GSL's closed forms. The half-span is found by
// bisection on the exact first moment of a polygon, the distance reference is
// a polygon's, and the balance is read off the body itself, by summing its
// sections on a grid. The throw is the crank of shaft_cranks_test.go: webs of
// 25 about the axis and 18 about the pin, thrown 40 at 0.7 rad, a pin of 28.

const (
	cwPinDia  = 28.0
	cwArcStep = 2000
)

// moment is the polygon's first moment about the origin along the unit vector
// (ux, uy), the integral of x*ux + y*uy over its area; exact for a polygon,
// whichever way it winds.
func (pg poly2) moment(ux, uy float64) float64 {
	var area2, mx, my float64
	for i, a := range pg {
		b := pg[(i+1)%len(pg)]
		cross := a[0]*b[1] - b[0]*a[1]
		area2 += cross
		mx += (a[0] + b[0]) * cross
		my += (a[1] + b[1]) * cross
	}
	m := (mx*ux + my*uy) / 6
	if area2 < 0 {
		m = -m
	}
	return m
}

// arcPoints are cwArcStep+1 points on the circle of radius r about the origin,
// from angle a0 to a1.
func arcPoints(r, a0, a1 float64) [][2]float64 {
	pts := make([][2]float64, cwArcStep+1)
	for i := range pts {
		a := a0 + (a1-a0)*float64(i)/cwArcStep
		pts[i] = [2]float64{r * math.Cos(a), r * math.Sin(a)}
	}
	return pts
}

// piePolygon is the pie of radius swingR about the origin, half-span half
// either side of the direction centre: the shape the GSL's counterweight adds
// to the web, whose part inside the web's main circle is the web's already.
func piePolygon(swingR, centre, half float64) poly2 {
	return append(poly2{{0, 0}}, arcPoints(swingR, centre-half, centre+half)...)
}

// sectorPolygon is the same pie without the part inside r0: what it adds.
func sectorPolygon(r0, r1, centre, half float64) poly2 {
	return append(poly2(arcPoints(r1, centre-half, centre+half)), arcPoints(r0, centre+half, centre-half)...)
}

// pinMoment is the pin's first moment, a round of cwPinDia and length pinLen
// at the throw.
func pinMoment(pinLen float64) float64 {
	return math.Pi * cwPinDia * cwPinDia / 4 * pinLen * crankThrow
}

// weightTarget is the first moment, per unit thickness, that a web's
// counterweight has to add: balance times its hull and half the pin's, shared
// over the web's thickness.
func weightTarget(mainR, pinR, thickness, pinLen, balance float64) float64 {
	hull := crankHull(mainR, pinR).moment(math.Cos(crankAngle), math.Sin(crankAngle))
	return balance * (hull + pinMoment(pinLen)/(2*thickness))
}

// widestHalfSpan is the largest half-span that keeps the sector off the hull:
// a quarter circle, or less where pinR is the larger radius and the hull's
// tangent comes round to meet it.
func widestHalfSpan(mainR, pinR float64) float64 {
	return math.Min(math.Pi/2, math.Pi-math.Acos((mainR-pinR)/crankThrow))
}

// weightHalfSpan is the half-span whose sector adds the moment a counterweight
// is asked for, found by bisection, or the widest there is when that falls
// short.
func weightHalfSpan(mainR, pinR, from, to, pinLen, swingR, balance float64) float64 {
	want := weightTarget(mainR, pinR, to-from, pinLen, balance)
	centre := crankAngle + math.Pi
	ux, uy := math.Cos(centre), math.Sin(centre)
	moment := func(half float64) float64 { return sectorPolygon(mainR, swingR, centre, half).moment(ux, uy) }
	hi := widestHalfSpan(mainR, pinR)
	if moment(hi) <= want {
		return hi
	}
	lo := 0.0
	for i := 0; i < 60; i++ {
		if mid := (lo + hi) / 2; moment(mid) < want {
			lo = mid
		} else {
			hi = mid
		}
	}
	return (lo + hi) / 2
}

// crankWeightedWeb is a web with its counterweight: the hull and the pie, whose
// min is the exact distance to their union outside it.
func crankWeightedWeb(mainR, pinR, from, to, pinLen, swingR, balance float64) crankSeg {
	hull := crankHull(mainR, pinR)
	pie := piePolygon(swingR, crankAngle+math.Pi, weightHalfSpan(mainR, pinR, from, to, pinLen, swingR, balance))
	return crankSeg{
		section: func(x, y float64) float64 {
			return math.Min(hull.distance([2]float64{x, y}), pie.distance([2]float64{x, y}))
		},
		from: from, to: to,
		gsl: fmt.Sprintf("shaft.web(%v, %v, %v, %v, %v, %v, shaft.counterweight(%v, %v, %v, %v))",
			mainR, pinR, crankThrow, crankAngle, from, to, swingR, balance, cwPinDia, pinLen),
	}
}

// weightedThrow is the single throw of shaft_cranks_test.go with a
// counterweight on each web.
func weightedThrow(swingR, balance float64) []crankSeg {
	return []crankSeg{
		crankJournal(30, 0, 30),
		crankWeightedWeb(25, 18, 30, 45, 30, swingR, balance),
		crankEccentric(cwPinDia, 45, 75),
		crankWeightedWeb(25, 18, 75, 90, 30, swingR, balance),
		crankJournal(30, 90, 120),
	}
}

// cwGrid covers a throw and its counterweights, which stand out on the side
// away from the pin.
func cwGrid(z0, z1, step float64) [][3]float64 {
	var pts [][3]float64
	for x := -58.0; x <= 58; x += step {
		for y := -58.0; y <= 58; y += step {
			for z := z0; z <= z1; z += step {
				pts = append(pts, [3]float64{x, y, z})
			}
		}
	}
	return pts
}

// checkWeighted: the segment is exact outside, and on the right side
// everywhere clear of the surface.
func checkWeighted(t *testing.T, label string, seg crankSeg) {
	t.Helper()
	pts := cwGrid(seg.from-6, seg.to+6, 1.9)
	got := evalStdShaft(t, crankDoc([]crankSeg{seg}), pts)
	outside, miss, wrongSign, worst := 0, 0, 0, 0.0
	for i, q := range pts {
		want := seg.distance(q)
		if math.Abs(want) > 1e-3 && (got[i] < 0) != (want < 0) {
			if wrongSign++; wrongSign <= 3 {
				t.Errorf("%s at %v: field %v, distance %v: wrong side", label, q, got[i], want)
			}
		}
		if want <= 0 {
			continue
		}
		outside++
		d := math.Abs(got[i] - want)
		worst = math.Max(worst, d)
		if d > 1e-4 {
			if miss++; miss <= 3 {
				t.Errorf("%s at %v: field %v, distance %v", label, q, got[i], want)
			}
		}
	}
	t.Logf("%s: %d points outside, worst miss %.2g mm; %d on the wrong side", label, outside, worst, wrongSign)
}

// A web with its counterweight, alone: with the span in the middle of its
// range; asked for more than the swing radius can give, so that it stops at
// half a circle (a swing of 40 is short of the 49.7 that a balance of 1 takes
// here); and with pinR the larger radius, where the span stops at the hull.
func TestShaftCounterweight_WebIsExactOutside(t *testing.T) {
	checkWeighted(t, "balance 1, swing 52", crankWeightedWeb(25, 18, 0, 15, 30, 52, 1))
	checkWeighted(t, "balance 1.4, swing 60", crankWeightedWeb(25, 18, 0, 15, 30, 60, 1.4))
	checkWeighted(t, "balance 1, swing 40, short of it", crankWeightedWeb(25, 18, 0, 15, 30, 40, 1))
	checkWeighted(t, "pinR 25 over mainR 18, balance 1.5, swing 56", crankWeightedWeb(18, 25, 0, 15, 30, 56, 1.5))
}

// The whole crank with a counterweight on each web is exact outside against
// the min of its capped prisms, as the plain crank is.
func TestShaftCounterweight_CrankMatchesReference(t *testing.T) {
	segs := weightedThrow(52, 1)
	ref := crankReference(segs)
	pts := cwGrid(-5, 125, 2.5)
	got := evalStdShaft(t, crankDoc(segs), pts)
	outside, miss, wrongSign, worst := 0, 0, 0, 0.0
	for i, q := range pts {
		want := ref(q)
		if math.Abs(want) > 1e-3 && (got[i] < 0) != (want < 0) {
			if wrongSign++; wrongSign <= 3 {
				t.Errorf("at %v: field %v, distance %v: wrong side", q, got[i], want)
			}
		}
		if want <= 0 {
			continue
		}
		outside++
		d := math.Abs(got[i] - want)
		worst = math.Max(worst, d)
		if d > 1e-4 {
			if miss++; miss <= 3 {
				t.Errorf("at %v: field %v, distance %v", q, got[i], want)
			}
		}
	}
	t.Logf("weighted crank: %d points outside, worst miss %.2g mm; %d on the wrong side", outside, worst, wrongSign)
}

// Outside the weighted crank, within 3 mm of it, the gradient has unit length.
func TestShaftCounterweight_UnitGradientOutside(t *testing.T) {
	segs := weightedThrow(52, 1)
	ref := crankReference(segs)
	var pts [][3]float64
	for _, q := range cwGrid(-3, 123, 2.3) {
		if d := ref(q); d > 0.05 && d < 3 {
			pts = append(pts, q)
		}
	}
	samples := evalBlendGrad(t, crankDoc(segs), pts)
	bad, worst := 0, 0.0
	for i, s := range samples {
		g := math.Sqrt(float64(s.Dx)*float64(s.Dx) + float64(s.Dy)*float64(s.Dy) + float64(s.Dz)*float64(s.Dz))
		worst = math.Max(worst, math.Abs(g-1))
		if math.Abs(g-1) > 1e-3 {
			if bad++; bad <= 3 {
				t.Errorf("at %v: |grad f| = %v", pts[i], g)
			}
		}
	}
	t.Logf("weighted crank: %d points outside within 3 mm, worst | |grad f| - 1 | %.2g", len(pts), worst)
}

// No sheet on the weighted crank's joints: on each joint plane every point
// inside both sections reads solid, no shallower than the body promises
// against a lower bound of the true depth. A joint's inside-both region lies
// in the smaller segment: the journal, or the pin.
func TestShaftCounterweight_NoSheetOnJoints(t *testing.T) {
	segs := weightedThrow(52, 1)
	doc := crankDoc(segs)
	cx, cy := crankPinCentre()
	var pts [][3]float64
	var depth []float64
	for j := 0; j+1 < len(segs); j++ {
		below, above := segs[j], segs[j+1]
		z := below.to
		for x := -17.0; x <= 47; x += 0.5 {
			for y := -17.0; y <= 42; y += 0.5 {
				if math.Hypot(x, y) > 16 && math.Hypot(x-cx, y-cy) > 15 {
					continue
				}
				dBelow, dAbove := below.section(x, y), above.section(x, y)
				if dBelow > -0.05 || dAbove > -0.05 {
					continue
				}
				lb := math.Min(math.Min(-dBelow, -dAbove), math.Min(z-below.from, above.to-z))
				pts = append(pts, [3]float64{x, y, z})
				depth = append(depth, lb)
			}
		}
	}
	got := evalStdShaft(t, doc, pts)
	bad := 0
	for i, q := range pts {
		if got[i] > -1e-3 || tooShallow(got[i], -depth[i]) {
			if bad++; bad <= 3 {
				t.Errorf("joint point %v reads %v where the solid is at least %v deep", q, got[i], depth[i])
			}
		}
	}
	t.Logf("%d joint points inside both sides, %d shallow", len(pts), bad)
}

// planeMoments integrates the body's section at height z on a grid of cell h
// (midpoint sums): its area and its first moments along the pin's direction
// and across it, about the shaft's axis. The body here is prisms, so a
// section stands for its segment between its planes.
func planeMoments(t *testing.T, doc string, zs []float64, h float64) (mu, mv []float64) {
	t.Helper()
	cu, su := math.Cos(crankAngle), math.Sin(crankAngle)
	var pts [][3]float64
	for _, z := range zs {
		for x := -58 + h/2; x < 58; x += h {
			for y := -58 + h/2; y < 58; y += h {
				pts = append(pts, [3]float64{x, y, z})
			}
		}
	}
	got := evalStdShaft(t, doc, pts)
	mu, mv = make([]float64, len(zs)), make([]float64, len(zs))
	for i, q := range pts {
		if got[i] < 0 {
			k := i / (len(pts) / len(zs))
			mu[k] += (q[0]*cu + q[1]*su) * h * h
			mv[k] += (-q[0]*su + q[1]*cu) * h * h
		}
	}
	return mu, mv
}

// throwMoment is a throw's first moment along the pin: the sum, over the
// segments zs sample, of each section's moment times its thickness.
func throwMoment(mu, thickness []float64) float64 {
	sum := 0.0
	for i := range mu {
		sum += mu[i] * thickness[i]
	}
	return sum
}

// The first moment of the whole body, summed from its own sections, falls
// as (1 - balance) of the bare throw's, and across the pin it stays zero. At 1
// the throw is balanced; at 1.25 it is as far out the other way as 0.75 is out
// this way. A balance out of the swing radius's reach delivers what the half
// circle holds.
func TestShaftCounterweight_BalancesTheThrow(t *testing.T) {
	// the segments' mid-planes and thicknesses: web, pin, web.
	zs := []float64{37.5, 60, 82.5}
	th := []float64{15, 30, 15}
	const h = 0.25

	// The bare throw's moment, from polygons and the pin's formula, not the GSL.
	hull := crankHull(25, 18).moment(math.Cos(crankAngle), math.Sin(crankAngle))
	bare := 2*15*hull + pinMoment(30)

	// A swing of 52 reaches a balance of 1 (it takes 49.7) but not 1.25 (53.1),
	// so 1.25 is also asked of 56.
	for _, c := range []struct{ f, swing float64 }{{0, 52}, {0.5, 52}, {1, 52}, {1.25, 56}} {
		mu, mv := planeMoments(t, crankDoc(weightedDocSegs(c.swing, c.f)), zs, h)
		got := throwMoment(mu, th)
		across := throwMoment(mv, th)
		if c.f == 0 {
			if math.Abs(got-bare) > 0.003*bare {
				t.Errorf("bare throw: summed from the body %.0f, from polygons %.0f", got, bare)
			}
		}
		want := (1 - c.f) * bare
		t.Logf("balance %.2f, swing %v: moment along the pin %.0f (want %.0f, bare throw %.0f); across %.0f", c.f, c.swing, got, want, bare, across)
		if math.Abs(got-want) > 0.005*bare {
			t.Errorf("balance %v: moment %.0f, want %.0f, within 0.5%% of the bare throw's %.0f", c.f, got, want, bare)
		}
		if math.Abs(across) > 0.005*bare {
			t.Errorf("balance %v: moment across the pin %.0f, want 0", c.f, across)
		}
	}

	// Asked for more than the swing radius can give, the counterweight is the
	// half circle, and that is what it delivers: the bare throw less two
	// half-circle sectors of 15 mm. Here 1.25 at a swing of 52.
	centre := crankAngle + math.Pi
	half := sectorPolygon(25, 52, centre, math.Pi/2).moment(math.Cos(centre), math.Sin(centre))
	want := bare - 2*15*half
	mu, _ := planeMoments(t, crankDoc(weightedDocSegs(52, 1.25)), zs, h)
	got := throwMoment(mu, th)
	t.Logf("balance 1.25, swing 52, out of reach: moment %.0f, half circles leave %.0f", got, want)
	if math.Abs(got-want) > 0.005*bare {
		t.Errorf("out of reach: moment %.0f, the half circles leave %.0f", got, want)
	}
}

// weightedDocSegs is weightedThrow as bare GSL segments, so a balance of 0
// can ride the same document.
func weightedDocSegs(swingR, balance float64) []crankSeg {
	return []crankSeg{
		crankJournal(30, 0, 30),
		{gsl: fmt.Sprintf("shaft.web(25.0, 18.0, %v, %v, 30.0, 45.0, shaft.counterweight(%v, %v, %v, 30.0))", crankThrow, crankAngle, swingR, balance, cwPinDia)},
		crankEccentric(cwPinDia, 45, 75),
		{gsl: fmt.Sprintf("shaft.web(25.0, 18.0, %v, %v, 75.0, 90.0, shaft.counterweight(%v, %v, %v, 30.0))", crankThrow, crankAngle, swingR, balance, cwPinDia)},
		crankJournal(30, 90, 120),
	}
}

// Webs of 15 and 25 either side of the pin balance statically at 1, since each
// counterweight cancels its own web's hull and half the pin whatever the
// thickness, but not the couple: that depends on where the webs stand, and
// they are not equally far from the pin's centre plane. The couple about that
// plane is half the pin's moment times the webs' offset from symmetry:
// -pinM/2 * (z1 + z2 - 2 zc), the webs' mid-planes z1 and z2.
func TestShaftCounterweight_StaticBalanceNeedsNoSymmetry(t *testing.T) {
	segs := []crankSeg{
		crankJournal(30, 0, 30),
		{gsl: fmt.Sprintf("shaft.web(25.0, 18.0, %v, %v, 30.0, 45.0, shaft.counterweight(52.0, 1.0, %v, 30.0))", crankThrow, crankAngle, cwPinDia)},
		crankEccentric(cwPinDia, 45, 75),
		{gsl: fmt.Sprintf("shaft.web(25.0, 18.0, %v, %v, 75.0, 100.0, shaft.counterweight(52.0, 1.0, %v, 30.0))", crankThrow, crankAngle, cwPinDia)},
		crankJournal(30, 100, 130),
	}
	zs := []float64{37.5, 60, 87.5}
	th := []float64{15, 30, 25}
	mu, mv := planeMoments(t, crankDoc(segs), zs, 0.25)
	static := throwMoment(mu, th)
	across := throwMoment(mv, th)
	couple := 0.0
	for i := range zs {
		couple += mu[i] * th[i] * (zs[i] - 60)
	}
	hull := crankHull(25, 18).moment(math.Cos(crankAngle), math.Sin(crankAngle))
	bare := 15*hull + 25*hull + pinMoment(30)
	wantCouple := -pinMoment(30) / 2 * (37.5 + 87.5 - 2*60)
	t.Logf("static %.0f, across %.0f (bare throw %.0f); couple %.0f, predicted %.0f", static, across, bare, couple, wantCouple)
	if math.Abs(static) > 0.005*bare || math.Abs(across) > 0.005*bare {
		t.Errorf("not balanced statically: along %.0f, across %.0f, bare throw %.0f", static, across, bare)
	}
	if math.Abs(couple-wantCouple) > 0.02*math.Abs(wantCouple) {
		t.Errorf("couple %.0f, predicted %.0f", couple, wantCouple)
	}
}

// shaft.minSwingRadius is the radius at which the sector is half a circle and
// the moment is the one asked for: the polygon's half-circle sector of that
// radius adds exactly the target moment.
func TestShaftCounterweight_MinSwingRadius(t *testing.T) {
	cases := []struct{ mainR, pinR, from, to, pinLen, balance float64 }{
		{25, 18, 40, 55, 30, 1},
		{25, 18, 40, 55, 30, 0.6},
		{25, 18, 30, 60, 30, 1.3},
	}
	for _, c := range cases {
		doc := fmt.Sprintf("@scene(200)\nsolid = shaft.minSwingRadius(%v, %v, %v, %v, %v, %v, %v, %v) + 0.0 * p.x;\nreturn solid;\n",
			c.mainR, c.pinR, crankThrow, c.from, c.to, c.balance, cwPinDia, c.pinLen)
		got := evalStdShaft(t, doc, [][3]float64{{0, 0, 0}})[0]
		// the radius whose half-circle sector adds the target, by bisection
		want := weightTarget(c.mainR, c.pinR, c.to-c.from, c.pinLen, c.balance)
		centre := crankAngle + math.Pi
		ux, uy := math.Cos(centre), math.Sin(centre)
		lo, hi := c.mainR, 4*c.mainR
		for i := 0; i < 60; i++ {
			mid := (lo + hi) / 2
			if sectorPolygon(c.mainR, mid, centre, widestHalfSpan(c.mainR, c.pinR)).moment(ux, uy) < want {
				lo = mid
			} else {
				hi = mid
			}
		}
		t.Logf("%+v: minSwingRadius %.5f, polygons %.5f", c, got, (lo+hi)/2)
		if math.Abs(got-(lo+hi)/2) > 1e-3 {
			t.Errorf("%+v: shaft.minSwingRadius %v, polygons %v", c, got, (lo+hi)/2)
		}
	}
}

// A counterweight of balance 0 adds nothing, and neither does one whose swing
// radius lies within the web's main circle: the field outside is the plain
// web's, the six-argument form's.
func TestShaftCounterweight_AddsNothingWhenThereIsNothingToAdd(t *testing.T) {
	plain := fmt.Sprintf("@scene(200)\nsolid = shaft.body([shaft.web(25.0, 18.0, %v, %v, 0.0, 15.0)]);\nreturn solid;\n", crankThrow, crankAngle)
	pts := cwGrid(-4, 19, 2.1)
	want := evalStdShaft(t, plain, pts)
	for _, c := range []struct {
		name          string
		swing, weight float64
	}{{"balance 0", 52, 0}, {"swing inside the main circle", 20, 1}} {
		doc := fmt.Sprintf("@scene(200)\nsolid = shaft.body([shaft.web(25.0, 18.0, %v, %v, 0.0, 15.0, shaft.counterweight(%v, %v, %v, 30.0))]);\nreturn solid;\n",
			crankThrow, crankAngle, c.swing, c.weight, cwPinDia)
		got := evalStdShaft(t, doc, pts)
		bad := 0
		for i, q := range pts {
			if want[i] > 0 && math.Abs(got[i]-want[i]) > 1e-9 || (got[i] < 0) != (want[i] < 0) {
				if bad++; bad <= 3 {
					t.Errorf("%s at %v: field %v, plain web %v", c.name, q, got[i], want[i])
				}
			}
		}
		t.Logf("%s: %d points, %d differ from the plain web", c.name, len(pts), bad)
	}
}

// exampleCrank is the shipped examples/shaft_crank.gsl with its sliders moved
// by defs, and the document text, evaluated as its `crank` in the shaft's own
// frame.
func exampleCrank(t *testing.T, src, defs string, pts [][3]float64) []float64 {
	t.Helper()
	e := New(nil, nil)
	groups, err := e.EvalGsl(t.Context(), nil, src, defs, []string{"crank"}, pts)
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	out := make([]float64, len(pts))
	for i := range pts {
		out[i] = float64(groups[0].Samples[0][i].Value)
	}
	return out
}

// The shipped example keeps its throw balanced as the throw slider moves: its
// counterweights are sized by shaft.minSwingRadius plus a margin, and a balance
// of 1 takes the moment along the pin (+x at the example's angle of 0) to
// nothing at every throw, where a balance of 0 leaves it out toward the pin.
func TestShaftCounterweight_ExampleStaysBalancedAsTheThrowMoves(t *testing.T) {
	raw, err := os.ReadFile("../examples/shaft_crank.gsl")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	// the webs' and the pin's mid-planes and thicknesses in the example
	zs := []float64{47.5, 70, 92.5}
	th := []float64{15, 30, 15}
	const h, span = 0.25, 68.0
	moment := func(defs string) float64 {
		var pts [][3]float64
		for _, z := range zs {
			for x := -span + h/2; x < span; x += h {
				for y := -span + h/2; y < span; y += h {
					pts = append(pts, [3]float64{x, y, z})
				}
			}
		}
		got := exampleCrank(t, src, defs, pts)
		per := len(pts) / len(zs)
		sum := 0.0
		for i, q := range pts {
			if got[i] < 0 {
				sum += q[0] * h * h * th[i/per]
			}
		}
		return sum
	}
	for _, throw := range []float64{30, 40, 50} {
		bare := moment(fmt.Sprintf("crankThrow = %v; balance = 0.0;", throw))
		balanced := moment(fmt.Sprintf("crankThrow = %v; balance = 1.0;", throw))
		t.Logf("throw %v: moment along the pin %.0f bare, %.0f at balance 1 (%.3f%% of it)", throw, bare, balanced, 100*balanced/bare)
		if bare <= 0 {
			t.Errorf("throw %v: the bare throw's moment is %.0f, want it toward the pin", throw, bare)
		}
		if math.Abs(balanced) > 0.005*bare {
			t.Errorf("throw %v: moment %.0f at balance 1, want within 0.5%% of the bare throw's %.0f", throw, balanced, bare)
		}
	}
}

// The example's fillets, where each journal and the pin rise from the webs,
// add nothing out on a counterweight: with them stripped from the document the
// field is the same on the counterweights' faces and edges, clear of the
// journals and the pin.
func TestShaftCounterweight_ExampleFilletsLeaveTheCounterweightAlone(t *testing.T) {
	raw, err := os.ReadFile("../examples/shaft_crank.gsl")
	if err != nil {
		t.Fatal(err)
	}
	src := string(raw)
	bare := strings.ReplaceAll(src, "   shaft.fillet(2.0),\n", "")
	if strings.Count(src, "shaft.fillet(2.0)") != 4 || strings.Contains(bare, "shaft.fillet") {
		t.Fatalf("the example's fillets are not the four lines this test strips")
	}
	// the counterweights stand out toward -x at the example's angle of 0; the
	// pin and the journals are more than 25 mm from these points.
	var pts [][3]float64
	for x := -64.0; x <= -27; x += 1.7 {
		for y := -34.0; y <= 34; y += 1.7 {
			for _, z := range []float64{38, 40, 41.3, 47.5, 53.7, 55, 57, 83, 85, 86.3, 92.5, 98.7, 100, 102} {
				pts = append(pts, [3]float64{x, y, z})
			}
		}
	}
	with := exampleCrank(t, src, "", pts)
	without := exampleCrank(t, bare, "", pts)
	bad, worst := 0, 0.0
	for i, q := range pts {
		d := math.Abs(with[i] - without[i])
		worst = math.Max(worst, d)
		if d > 1e-6 {
			if bad++; bad <= 3 {
				t.Errorf("at %v: %v with the fillets, %v without", q, with[i], without[i])
			}
		}
	}
	t.Logf("%d points on the counterweights, %d differ, worst %.2g mm", len(pts), bad, worst)
}

// The body's box holds the counterweight: its reach grows to the swing radius
// when that is more than the web's own 58, and stays 58 when it is less.
func TestShaftCounterweight_BoxHoldsTheSwing(t *testing.T) {
	segs := func(swing float64) string {
		return fmt.Sprintf("[shaft.diameter(30.0, 0.0, 30.0), shaft.web(25.0, 18.0, 40.0, 0.7, 30.0, 45.0, shaft.counterweight(%v, 1.0, 28.0, 30.0)), shaft.eccentric(28.0, 40.0, 0.7, 45.0, 75.0), shaft.web(25.0, 18.0, 40.0, 0.7, 75.0, 90.0, shaft.counterweight(%v, 1.0, 28.0, 30.0)), shaft.diameter(30.0, 90.0, 120.0)]", swing, swing)
	}
	for _, c := range []boxCase{
		{"swing 62", segs(62), 62, 0, 120},
		{"swing 52", segs(52), 58, 0, 120},
	} {
		checkBox(t, "body", c)
		checkBox(t, "envelope", c)
	}
}
