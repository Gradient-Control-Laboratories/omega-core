//go:build cgo

package eval

import (
	"context"
	"math"
	"strconv"
	"testing"
)

// An audit of the bevel gear against references that owe nothing to its own
// construction. gears_bevel_test.go reads the field back at points derived the
// way bevelTeeth derives them; these tests take their expected values from
// elsewhere:
//
//   - the equal-addendum, standard-taper straight-bevel formulas (face and root
//     angles from atan(addendum / R) and atan(dedendum / R), outside diameter
//     d + 2a cos(delta), pitch apex to crown R cos(delta) - a sin(delta), and a
//     circular tooth thickness of pi m / 2 at the outer end, scaled with cone
//     distance), with every surface located by root-finding the field;
//   - the spherical involute, the conjugate bevel tooth, against which the
//     Tredgold tooth's profile is measured.
//
// Tredgold's tooth is an approximation, so the profile is held to a limit
// rather than to zero: 5 um at module 2, twice its departure when measured
// (2.3 um). A pair's phasing is TestBevelPairMesh's.

// bevelRefDoc is the reference gear of gears_bevel_test.go (20 teeth, module 2,
// pressure angle 0.35, a 45 degree pitch cone, face width 12, both recesses 2)
// with no backlash, so its tooth thickness is the theoretical one.
const bevelRefDoc = `
teeth = gears.bevelTeeth(20.0, 2.0, 0.35, 0.7853981633974483, 0.0, 0.0, 0.0, 0.25, 0.05);
g = gears.bevelGear(20.0, 2.0, 0.35, 0.7853981633974483, 0.0, 0.0, 12.0, 2.0, 2.0, 0.0, 0.0, 0.0, 0.0, 8.0, 0.25, 0.05);
return g;
`

// bevelRefCrossing finds where expr changes sign along the curve at(s) for s
// in [s0, s1], by sampling it densely, then again across the bracketing step.
func bevelRefCrossing(t *testing.T, e *Evaluator, expr string, at func(s float64) [3]float64, s0, s1 float64) float64 {
	t.Helper()
	ctx := context.Background()
	const n = 400
	for pass := 0; pass < 2; pass++ {
		pts := make([][3]float64, n+1)
		for i := range pts {
			pts[i] = at(s0 + (s1-s0)*float64(i)/n)
		}
		g, err := e.EvalGsl(ctx, nil, bevelRefDoc, "", []string{expr}, pts)
		if err != nil {
			t.Fatalf("eval %q: %v", expr, err)
		}
		found := false
		for i := 0; i < n; i++ {
			a := float64(g[0].Samples[0][i].Value)
			b := float64(g[0].Samples[0][i+1].Value)
			if (a < 0) != (b < 0) {
				lo, hi := s0+(s1-s0)*float64(i)/n, s0+(s1-s0)*float64(i+1)/n
				if pass == 1 {
					return lo + (hi-lo)*a/(a-b)
				}
				s0, s1 = lo, hi
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("%s: no sign change on [%g, %g]", expr, s0, s1)
		}
	}
	return math.NaN()
}

// TestBevelReferenceDimensions locates the gear's surfaces by root-finding and
// compares them with the standard straight-bevel formulas.
func TestBevelReferenceDimensions(t *testing.T) {
	e := bevelEvaluator(t)
	const (
		z     = 20.0
		m     = 2.0
		delta = math.Pi / 4
		face  = 12.0
	)
	sinD, cosD := math.Sin(delta), math.Cos(delta)
	outer := z * m / (2 * sinD)
	inner := outer - face
	addendum, dedendum := m, (1+bevelClearance)*m
	toothAt := math.Pi / z

	// Face and root angles, at both ends: the tip on a tooth's centreline, the
	// root on a space's. Standard taper puts both cones' apexes at the pitch
	// apex, so each angle must be the same at every cone distance.
	wantFace := delta + math.Atan(addendum/outer)
	wantRoot := delta - math.Atan(dedendum/outer)
	for _, cone := range []float64{outer, inner} {
		tag := "@cone " + strconv.FormatFloat(cone, 'f', 2, 64) + " "
		tip := bevelRefCrossing(t, e, "teeth", func(th float64) [3]float64 { return bevelPoint(cone, th, toothAt) },
			delta, delta+2*m/outer)
		root := bevelRefCrossing(t, e, "teeth", func(th float64) [3]float64 { return bevelPoint(cone, th, 0) },
			delta-2*m/outer, delta)
		if math.Abs(tip-wantFace) > 5e-4 {
			t.Errorf("%sface angle %.6f rad, standard %.6f", tag, tip, wantFace)
		}
		if math.Abs(root-wantRoot) > 5e-4 {
			t.Errorf("%sroot angle %.6f rad, standard %.6f", tag, root, wantRoot)
		}
		t.Logf("%sface angle %.6f (standard %.6f, diff %+.2e), root angle %.6f (standard %.6f, diff %+.2e)",
			tag, tip, wantFace, tip-wantFace, root, wantRoot, root-wantRoot)

		// Circular tooth thickness on the pitch cone: from the flank crossing
		// between the space centred on azimuth 0 and the tooth beside it.
		flank := bevelRefCrossing(t, e, "teeth", func(phi float64) [3]float64 { return bevelPoint(cone, delta, phi) },
			0, toothAt)
		thickness := 2 * (toothAt - flank) * cone * sinD
		wantThickness := math.Pi * m / 2 * cone / outer
		if math.Abs(thickness-wantThickness) > 2e-3 {
			t.Errorf("%scircular tooth thickness %.5f, standard %.5f", tag, thickness, wantThickness)
		}
		t.Logf("%scircular tooth thickness %.5f (standard %.5f, diff %+.2e)",
			tag, thickness, wantThickness, thickness-wantThickness)
	}

	// The crown: where the tip cone meets the back cone. The back cone is
	// located on the faced gear along two lines parallel to the pitch element,
	// at heights h above it on a tooth's centreline; it must be normal to the
	// pitch element, and the crown it makes with the face cone gives the
	// outside diameter and the pitch apex to crown.
	onLine := func(h float64) func(s float64) [3]float64 {
		return func(s float64) [3]float64 {
			r := s*sinD + h*cosD
			zz := s*cosD - h*sinD
			return [3]float64{r * math.Cos(toothAt), r * math.Sin(toothAt), zz}
		}
	}
	h1, h2 := 0.3*m, 0.8*m
	s1 := bevelRefCrossing(t, e, "g", onLine(h1), outer-2, outer+2)
	s2 := bevelRefCrossing(t, e, "g", onLine(h2), outer-2, outer+2)
	backTilt := math.Atan((s2 - s1) / (h2 - h1)) // 0 when normal to the pitch element
	if math.Abs(backTilt) > 1e-3 {
		t.Errorf("back cone is tilted %.5f rad from normal to the pitch element", backTilt)
	}
	// The back cone line, in (r, z): through the point at h1 along direction
	// (s2 - s1, h2 - h1) in the pitch element's frame.
	p1 := [2]float64{s1*sinD + h1*cosD, s1*cosD - h1*sinD}
	p2 := [2]float64{s2*sinD + h2*cosD, s2*cosD - h2*sinD}
	// The face cone line, from the apex at the measured face angle (taken at
	// the outer end, where the crown is).
	faceAngle := bevelRefCrossing(t, e, "teeth", func(th float64) [3]float64 { return bevelPoint(outer, th, toothAt) },
		delta, delta+2*m/outer)
	fr, fz := math.Sin(faceAngle), math.Cos(faceAngle)
	// Solve p1 + u (p2 - p1) = L (fr, fz).
	dr, dz := p2[0]-p1[0], p2[1]-p1[1]
	det := dr*fz - dz*fr
	u := (p1[1]*fr - p1[0]*fz) / det
	crownR, crownZ := p1[0]+u*dr, p1[1]+u*dz
	wantOutside := z*m + 2*addendum*cosD
	wantApexToCrown := outer*cosD - addendum*sinD
	if math.Abs(2*crownR-wantOutside) > 1e-2 {
		t.Errorf("outside diameter %.5f, standard %.5f", 2*crownR, wantOutside)
	}
	if math.Abs(crownZ-wantApexToCrown) > 1e-2 {
		t.Errorf("pitch apex to crown %.5f, standard %.5f", crownZ, wantApexToCrown)
	}
	t.Logf("back cone tilt %+.2e rad; outside diameter %.5f (standard %.5f, diff %+.2e); apex to crown %.5f (standard %.5f, diff %+.2e)",
		backTilt, 2*crownR, wantOutside, 2*crownR-wantOutside, crownZ, wantApexToCrown, crownZ-wantApexToCrown)
}

// sphericalInvolute returns the point at roll t of the spherical involute of
// the base cone of half-angle baseAngle, on the unit sphere, starting on the
// base circle at azimuth 0: the trace of a point on a great-circle plane that
// rolls without slipping on the base cone, which turns through an arc of
// t sin(baseAngle) along the plane as it rolls through azimuth t.
func sphericalInvolute(baseAngle, t float64) [3]float64 {
	sb, cb := math.Sin(baseAngle), math.Cos(baseAngle)
	s := t * sb
	e := [3]float64{sb * math.Cos(t), sb * math.Sin(t), cb}
	c := [3]float64{-math.Sin(t), math.Cos(t), 0}
	return [3]float64{
		math.Cos(s)*e[0] - math.Sin(s)*c[0],
		math.Cos(s)*e[1] - math.Sin(s)*c[1],
		math.Cos(s)*e[2] - math.Sin(s)*c[2],
	}
}

// TestBevelReferenceProfile measures how far the Tredgold tooth departs from
// the spherical involute, the conjugate tooth of the same pitch cone and
// pressure angle: its base cone has sin(base) = sin(delta) cos(alpha). The
// involute is placed by the standard tooth thickness, crossing the pitch cone
// half a tooth thickness from the tooth's centreline, and the field read along
// it from just above the base circle (or root) to just below the tip is the
// departure, positive where the Tredgold tooth is the thinner, and it must stay
// within 5 um. The field is a function of direction alone, so the departure at
// the inner end is this one scaled by cone distance.
func TestBevelReferenceProfile(t *testing.T) {
	e := bevelEvaluator(t)
	ctx := context.Background()
	const (
		z     = 20.0
		m     = 2.0
		alpha = bevelPressure
		delta = math.Pi / 4
	)
	outer := z * m / (2 * math.Sin(delta))
	base := math.Asin(math.Sin(delta) * math.Cos(alpha))
	rootAngle := delta - (1+bevelClearance)*m/outer
	tipAngle := delta + m/outer

	polar := func(q [3]float64) float64 { return math.Acos(q[2]) }
	azimuth := func(q [3]float64) float64 { return math.Atan2(q[1], q[0]) }
	// The roll at which the involute reaches polar angle th: cos(th) =
	// cos(s) cos(base) on it, with s = t sin(base).
	rollAt := func(th float64) float64 {
		return math.Acos(math.Cos(th)/math.Cos(base)) / math.Sin(base)
	}
	// Turn the involute so it crosses the pitch cone at the standard flank:
	// the tooth on azimuth pi/z is pi m/2 thick at the outer end, an angle of
	// pi/(2z) either side of its centreline. Its azimuth grows with polar
	// angle, as that tooth's lower flank does.
	turn := math.Pi/(2*z) - azimuth(sphericalInvolute(base, rollAt(delta)))

	low := math.Max(base, rootAngle) + 0.5/outer
	high := tipAngle - 0.15/outer
	const n = 24
	pts := make([][3]float64, n+1)
	heights := make([]float64, n+1)
	for i := range pts {
		th := low + (high-low)*float64(i)/n
		q := sphericalInvolute(base, rollAt(th))
		phi := azimuth(q) + turn
		pts[i] = bevelPoint(outer, polar(q), phi)
		heights[i] = (th - delta) * outer
	}
	g, err := e.EvalGsl(ctx, nil, bevelRefDoc, "", []string{"teeth"}, pts)
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	worst := 0.0
	for i := range pts {
		v := float64(g[0].Samples[0][i].Value)
		if math.IsNaN(v) {
			t.Fatalf("field is NaN on the reference involute at height %.3f", heights[i])
		}
		if math.Abs(v) > math.Abs(worst) {
			worst = v
		}
		t.Logf("height %+.3f mm from the pitch cone: departure %+.5f mm", heights[i], v)
	}
	if math.Abs(worst) > 5e-3 {
		t.Errorf("the tooth departs %.5f mm from the spherical involute, more than 0.005", worst)
	}
	t.Logf("base cone %.5f rad, root %.5f, tip %.5f; worst departure %+.5f mm at the outer end (%.2f%% of the module)",
		base, rootAngle, tipAngle, worst, 100*math.Abs(worst)/m)
}
