//go:build cgo

package eval

import (
	"fmt"
	"math"
	"os"
	"sort"
	"testing"
)

// Shaft Modeling, Phase 8: cranks. A crank pin is shaft.eccentric(dia, throw,
// angle, from, to), a round whose axis stands `throw` from the shaft's at
// `angle` (radians from +x about +z). A web is shaft.web(mainR, pinR, throw,
// angle, from, to), the convex hull of a circle of mainR about the shaft's
// axis and one of pinR about the pin. Both are exact, inside and out, as
// prisms of exact sections are. A crank is a list of them between main
// journals, and the bridges join them with no sheet on any joint.
//
// The single throw under test: main journals of 30 from 0 to 30 and from 90
// to 120, webs of 25 about the axis and 18 about the pin from 30 to 45 and
// from 75 to 90, and a pin of 28 from 45 to 75, thrown 40 at 0.7 rad.

const (
	crankThrow = 40.0
	crankAngle = 0.7
)

// crankSeg is one segment of the crank: its 2D section's signed distance,
// its planes and its GSL.
type crankSeg struct {
	section  func(x, y float64) float64
	from, to float64
	gsl      string
}

func crankPinCentre() (float64, float64) {
	return crankThrow * math.Cos(crankAngle), crankThrow * math.Sin(crankAngle)
}

func crankJournal(dia, from, to float64) crankSeg {
	return crankSeg{
		section: func(x, y float64) float64 { return math.Hypot(x, y) - dia/2 },
		from:    from, to: to,
		gsl: fmt.Sprintf("shaft.diameter(%v, %v, %v)", dia, from, to),
	}
}

func crankEccentric(dia, from, to float64) crankSeg {
	cx, cy := crankPinCentre()
	return crankSeg{
		section: func(x, y float64) float64 { return math.Hypot(x-cx, y-cy) - dia/2 },
		from:    from, to: to,
		gsl: fmt.Sprintf("shaft.eccentric(%v, %v, %v, %v, %v)", dia, crankThrow, crankAngle, from, to),
	}
}

// crankHull is the convex hull of two circles, mainR about the origin and
// pinR about the pin, as a polygon through points sampled densely enough on
// both circles that it stands within 3.1e-5 mm of the true hull (radii up to 25). It is found by
// the monotone chain, independently of any closed form.
func crankHull(mainR, pinR float64) poly2 {
	cx, cy := crankPinCentre()
	const n = 2000
	pts := make([][2]float64, 0, 2*n)
	for i := 0; i < n; i++ {
		a := 2 * math.Pi * float64(i) / n
		pts = append(pts, [2]float64{mainR * math.Cos(a), mainR * math.Sin(a)})
		pts = append(pts, [2]float64{cx + pinR*math.Cos(a), cy + pinR*math.Sin(a)})
	}
	sort.Slice(pts, func(i, j int) bool {
		if pts[i][0] != pts[j][0] {
			return pts[i][0] < pts[j][0]
		}
		return pts[i][1] < pts[j][1]
	})
	cross := func(o, a, b [2]float64) float64 {
		return (a[0]-o[0])*(b[1]-o[1]) - (a[1]-o[1])*(b[0]-o[0])
	}
	var hull [][2]float64
	for _, p := range pts {
		for len(hull) >= 2 && cross(hull[len(hull)-2], hull[len(hull)-1], p) <= 0 {
			hull = hull[:len(hull)-1]
		}
		hull = append(hull, p)
	}
	lower := len(hull) + 1
	for i := len(pts) - 2; i >= 0; i-- {
		p := pts[i]
		for len(hull) >= lower && cross(hull[len(hull)-2], hull[len(hull)-1], p) <= 0 {
			hull = hull[:len(hull)-1]
		}
		hull = append(hull, p)
	}
	return poly2(hull[:len(hull)-1])
}

func crankWeb(mainR, pinR, from, to float64) crankSeg {
	hull := crankHull(mainR, pinR)
	return crankSeg{
		section: func(x, y float64) float64 { return hull.distance([2]float64{x, y}) },
		from:    from, to: to,
		gsl: fmt.Sprintf("shaft.web(%v, %v, %v, %v, %v, %v)", mainR, pinR, crankThrow, crankAngle, from, to),
	}
}

// prismDistance is the exact signed distance to a 2D section's prism between
// two planes, given the section's exact signed distance s.
func prismDistance(s, z, from, to float64) float64 {
	axial := math.Max(from-z, z-to)
	return math.Hypot(math.Max(s, 0), math.Max(axial, 0)) + math.Min(math.Max(s, axial), 0)
}

func (c crankSeg) distance(q [3]float64) float64 {
	return prismDistance(c.section(q[0], q[1]), q[2], c.from, c.to)
}

// singleThrow is the crank under test.
func singleThrow() []crankSeg {
	return []crankSeg{
		crankJournal(30, 0, 30),
		crankWeb(25, 18, 30, 45),
		crankEccentric(28, 45, 75),
		crankWeb(25, 18, 75, 90),
		crankJournal(30, 90, 120),
	}
}

// crankReference is the min of the segments' exact distances: the exact
// distance to their union outside it, and its sign everywhere.
func crankReference(segs []crankSeg) func([3]float64) float64 {
	return func(q [3]float64) float64 {
		d := math.Inf(1)
		for _, s := range segs {
			d = math.Min(d, s.distance(q))
		}
		return d
	}
}

func crankDoc(segs []crankSeg) string {
	list := ""
	for i, s := range segs {
		if i > 0 {
			list += ", "
		}
		list += s.gsl
	}
	return "@scene(200)\nsolid = shaft.body([" + list + "]);\nreturn solid;\n"
}

// crankGrid covers the crank: the axis's journals, the webs and the pin.
func crankGrid(z0, z1, step float64) [][3]float64 {
	var pts [][3]float64
	for x := -30.0; x <= 52; x += step {
		for y := -30.0; y <= 48; y += step {
			for z := z0; z <= z1; z += step {
				pts = append(pts, [3]float64{x, y, z})
			}
		}
	}
	return pts
}

// checkLoneSegment: a lone segment is exact inside and out.
func checkLoneSegment(t *testing.T, label string, seg crankSeg) {
	t.Helper()
	pts := crankGrid(seg.from-6, seg.to+6, 1.7)
	got := evalStdShaft(t, crankDoc([]crankSeg{seg}), pts)
	bad, worst := 0, 0.0
	for i, q := range pts {
		want := seg.distance(q)
		d := math.Abs(got[i] - want)
		worst = math.Max(worst, d)
		if d > 1e-4 {
			if bad++; bad <= 3 {
				t.Errorf("%s at %v: field %v, distance %v", label, q, got[i], want)
			}
		}
	}
	t.Logf("%s: %d points, worst miss %.2g mm", label, len(pts), worst)
}

// The crank pin alone is exact, inside and out.
func TestShaftCranks_EccentricIsExact(t *testing.T) {
	checkLoneSegment(t, "eccentric", crankEccentric(28, 0, 30))
}

// The web alone is exact, inside and out.
func TestShaftCranks_WebIsExact(t *testing.T) {
	checkLoneSegment(t, "web", crankWeb(25, 18, 0, 15))
}

// The whole crank: exact outside, and on the right side everywhere clear of
// the surface.
func TestShaftCranks_CrankMatchesReference(t *testing.T) {
	segs := singleThrow()
	ref := crankReference(segs)
	pts := crankGrid(-5, 125, 2.3)
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
	t.Logf("crank: %d points outside, worst miss %.2g mm; %d on the wrong side", outside, worst, wrongSign)
}

// Outside the crank, within 3 mm of it, the gradient has unit length.
func TestShaftCranks_UnitGradientOutside(t *testing.T) {
	segs := singleThrow()
	ref := crankReference(segs)
	var pts [][3]float64
	for _, q := range crankGrid(-3, 123, 1.9) {
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
	t.Logf("crank: %d points outside within 3 mm, worst | |grad f| - 1 | %.2g", len(pts), worst)
}

// No sheet on any of the crank's four joints: on each joint plane, every
// point inside both sections reads solid, and no shallower than the body
// promises against a lower bound of the true depth, the largest ball about the
// point that both prisms hold.
func TestShaftCranks_NoSheetOnJoints(t *testing.T) {
	segs := singleThrow()
	doc := crankDoc(segs)
	var pts [][3]float64
	var depth []float64
	for j := 0; j+1 < len(segs); j++ {
		below, above := segs[j], segs[j+1]
		z := below.to
		for x := -30.0; x <= 52; x += 0.5 {
			for y := -30.0; y <= 48; y += 0.5 {
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

// A fillet of 2 where the pin rises from the web is the exact arc. On the
// pin's outer side, within 50 degrees of the throw, the web's edge is the
// pin-end circle of 18, so about the pin's axis the crank there is a revolved
// staircase: the web to 18 below z = 45, the pin to 14 above, and the fillet's
// arc centred at (16, 47) between them.
func TestShaftCranks_PinFilletIsTheArc(t *testing.T) {
	doc := crankDoc([]crankSeg{
		crankJournal(30, 0, 30),
		crankWeb(25, 18, 30, 45),
		{gsl: "shaft.fillet(2.0)"},
		crankEccentric(28, 45, 75),
		crankWeb(25, 18, 75, 90),
		crankJournal(30, 90, 120),
	})
	right := [][2]float64{{18, 30}, {18, 45}}
	const n = 4000
	for i := 0; i <= n; i++ {
		phi := 1.5*math.Pi - (math.Pi/2)*float64(i)/n
		right = append(right, [2]float64{16 + 2*math.Cos(phi), 47 + 2*math.Sin(phi)})
	}
	right = append(right, [2]float64{14, 75})
	var meridian poly2
	meridian = append(meridian, right...)
	for i := len(right) - 1; i >= 0; i-- {
		meridian = append(meridian, [2]float64{-right[i][0], right[i][1]})
	}
	cx, cy := crankPinCentre()
	var pts [][3]float64
	var want []float64
	for _, da := range []float64{-0.8, -0.3, 0.2, 0.7} {
		a := crankAngle + da
		for rho := 12.5; rho <= 17.5; rho += 0.0625 {
			for z := 42.0; z <= 48; z += 0.0625 {
				d := meridian.distance([2]float64{rho, z})
				if d <= 0 || d > 1.2 {
					continue
				}
				pts = append(pts, [3]float64{cx + rho*math.Cos(a), cy + rho*math.Sin(a), z})
				want = append(want, d)
			}
		}
	}
	got := evalStdShaft(t, doc, pts)
	bad, worst := 0, 0.0
	for i, q := range pts {
		d := math.Abs(got[i] - want[i])
		worst = math.Max(worst, d)
		if d > 1e-4 {
			if bad++; bad <= 3 {
				t.Errorf("at %v: field %v, distance %v", q, got[i], want[i])
			}
		}
	}
	t.Logf("pin fillet: %d points outside within 1.2 mm, worst miss %.2g mm", len(pts), worst)
}

// The controls, which stand without the crank's segments. The hull polygon,
// with equal radii, is the capsule, whose distance has a closed form.
func TestShaftCranks_HullControl(t *testing.T) {
	const r = 20.0
	hull := crankHull(r, r)
	cx, cy := crankPinCentre()
	worst := 0.0
	for x := -40.0; x <= 70; x += 0.7 {
		for y := -40.0; y <= 70; y += 0.7 {
			h := math.Max(0, math.Min(1, (x*cx+y*cy)/(cx*cx+cy*cy)))
			want := math.Hypot(x-h*cx, y-h*cy) - r
			worst = math.Max(worst, math.Abs(hull.distance([2]float64{x, y})-want))
		}
	}
	if worst > 1e-4 {
		t.Errorf("the hull polygon misses the capsule by %.2g mm", worst)
	}
	t.Logf("hull polygon against the capsule: worst miss %.2g mm", worst)
}

// The reference, the grid and the check hold a coaxial stack the shaft
// package already builds to the same standard.
func TestShaftCranks_CoaxialControl(t *testing.T) {
	segs := []crankSeg{crankJournal(30, 0, 30), crankJournal(50, 30, 45), crankJournal(30, 45, 75)}
	ref := crankReference(segs)
	pts := crankGrid(-5, 80, 2.3)
	got := evalStdShaft(t, crankDoc(segs), pts)
	outside, worst := 0, 0.0
	for i, q := range pts {
		want := ref(q)
		if math.Abs(want) > 1e-3 && (got[i] < 0) != (want < 0) {
			t.Fatalf("at %v: field %v, distance %v: wrong side", q, got[i], want)
		}
		if want > 0 {
			outside++
			worst = math.Max(worst, math.Abs(got[i]-want))
		}
	}
	if worst > 1e-4 {
		t.Errorf("coaxial stack: worst miss %.2g mm outside", worst)
	}
	t.Logf("coaxial stack: %d points outside, worst miss %.2g mm", outside, worst)
}

// The crank example compiles to WGSL (its size logged as Phase 8's budget)
// and evaluates: solid in a main journal, a web and the pin, empty beyond the
// end and in the space the throw leaves opposite the pin.
func TestShaftCranks_CrankExampleCompilesToWGSL(t *testing.T) {
	src, err := os.ReadFile("../examples/shaft_crank.gsl")
	if err != nil {
		t.Fatal(err)
	}
	e := New(nil, nil)
	res, err := e.Shader(t.Context(), nil, string(src), "", ShaderOptions{Target: "wgsl", OutDir: t.TempDir()})
	if err != nil {
		t.Fatalf("shader: %v", err)
	}
	t.Logf("examples/shaft_crank.gsl: %d bytes of WGSL", len(res.Code))
	groups, err := e.EvalGsl(t.Context(), nil, string(src), "", []string{"crank"},
		[][3]float64{{0, 0, 20}, {20, 0, 47}, {40, 0, 70}, {0, 0, 145}, {-20, 0, 70}})
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	got := groups[0].Samples[0]
	if got[0].Value >= 0 || got[1].Value >= 0 || got[2].Value >= 0 || got[3].Value <= 0 || got[4].Value <= 0 {
		t.Errorf("journal %v, web %v, pin %v, beyond the end %v, opposite the pin %v: want solid three times, then empty twice",
			got[0].Value, got[1].Value, got[2].Value, got[3].Value, got[4].Value)
	}
}
