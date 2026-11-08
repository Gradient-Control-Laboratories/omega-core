package gsleval

import (
	"context"
	"math"
	"strconv"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// The reference bevel gear these tests read: 20 teeth, module 2 at the outer
// end, a 45 degree pitch cone. Everything else follows from those — the outer
// cone distance is not a free parameter.
const (
	bevelTeethCount = 20.0
	bevelModule     = 2.0
	bevelPressure   = 0.35
	bevelPitchAngle = math.Pi / 4
	bevelClearance  = 0.25
)

// bevelOuterCone is the cone distance of the large end, z * m / (2 sin(delta));
// bevelInnerCone is the small end, a face width in along the cone element.
var (
	bevelOuterCone = bevelTeethCount * bevelModule / (2 * math.Sin(bevelPitchAngle))
	bevelInnerCone = bevelOuterCone - 12.0
)

// bevelDoc exercises the bevel additions to the gears library: the bare
// conical tooth field, the faced-off gear, and pairs at several ratios and
// spiral angles. sameA/sameB is a deliberately WRONG pair — two gears of the
// same hand, which bevelPair never builds — kept so TestBevelPairMesh can show
// what the opposite-hand rule is worth.
const bevelDoc = `
phase = 0.0;
recess = 2.0;
hubDiam = 0.0;
hubHeight = 0.0;
hubChamfer = 0.0;
teeth = gears.bevelTeeth(20.0, 2.0, 0.35, 0.7853981633974483, 0.0, 0.0, 0.0, 0.25, 0.05);
g = gears.bevelGear(20.0, 2.0, 0.35, 0.7853981633974483, 0.0, 0.0, 12.0, recess, recess, 0.0, hubDiam, hubHeight, hubChamfer, 8.0, 0.25, 0.05);
straightA, straightB = gears.bevelPair(20.0, 20.0, 2.0, 0.35, 1.5707963267948966, 0.0, 12.0, 1.0, 1.0, 0.1, 14.0, 14.0, 4.0, 0.4, 8.0, 8.0, 0.25, 0.05, phase);
spiralA, spiralB = gears.bevelPair(20.0, 20.0, 2.0, 0.35, 1.5707963267948966, 0.6, 12.0, 1.0, 1.0, 0.1, 14.0, 14.0, 4.0, 0.4, 8.0, 8.0, 0.25, 0.05, phase);
ratioA, ratioB = gears.bevelPair(15.0, 25.0, 2.0, 0.35, 1.5707963267948966, 0.0, 9.0, 1.0, 1.0, 0.1, 11.0, 16.0, 3.0, 0.4, 6.0, 8.0, 0.25, 0.05, phase);
sameA = gears.bevelGear(20.0, 2.0, 0.35, 0.7853981633974483, 0.6, 2 * pi * phase / 20.0, 12.0, 1.0, 1.0, 0.1, 14.0, 4.0, 0.4, 8.0, 0.25, 0.05);
sameLocal = gears.bevelGear(20.0, 2.0, 0.35, 0.7853981633974483, 0.6, pi * 19.0 / 20.0 - 2 * pi * phase / 20.0, 12.0, 1.0, 1.0, 0.1, 14.0, 4.0, 0.4, 8.0, 0.25, 0.05);
sameB = remap(sameLocal, [neg(p.z), p.y, p.x]);
return min(straightA, straightB);
`

func bevelEvaluator(t *testing.T) *Evaluator {
	t.Helper()
	mgr := source.NewLocalManager("testpkg", "../../stdlib")
	return New(mgr, nil)
}

// bevelPoint returns the point at cone distance `cone`, polar angle `theta`
// from the +Z axis, and azimuth `phi` — the natural coordinates of a bevel
// gear, whose whole tooth pattern is a function of (theta, phi) alone.
func bevelPoint(cone, theta, phi float64) [3]float64 {
	r := cone * math.Sin(theta)
	return [3]float64{r * math.Cos(phi), r * math.Sin(phi), cone * math.Cos(theta)}
}

// TestBevelGearDimensions reads the bevel proportions back out of the field at
// two different cone distances. The second distance is the point: a bevel
// gear's teeth converge on the apex, so every feature must sit at the same
// ANGLE and scale its distance linearly with cone distance. A construction
// that merely looked like a bevel gear at the large end would fail the 20.0
// column while passing the 25.0 one.
func TestBevelGearDimensions(t *testing.T) {
	e := bevelEvaluator(t)
	ctx := context.Background()
	at := func(expr string, pt [3]float64) float64 {
		t.Helper()
		g, err := e.EvalGsl(ctx, nil, bevelDoc, "", []string{expr}, [][3]float64{pt})
		if err != nil {
			t.Fatalf("eval %q: %v", expr, err)
		}
		return float64(g[0].Samples[0][0].Value)
	}
	near := func(label string, got, want, tol float64) {
		t.Helper()
		if math.IsNaN(got) || math.Abs(got-want) > tol {
			t.Errorf("%s = %.5f, want %.5f±%.5f", label, got, want, tol)
		}
	}

	// gearProfile centers a tooth SPACE on angle zero, so the bevel built from
	// it puts a space on azimuth zero and a tooth half an angular pitch over.
	spaceAt := 0.0
	toothAt := math.Pi / bevelTeethCount
	flankAt := math.Pi / (2 * bevelTeethCount)

	// Addendum is one module and dedendum (1 + clearance) modules, both
	// measured as arc length at the outer cone distance.
	tipAngle := bevelPitchAngle + bevelModule/bevelOuterCone
	rootAngle := bevelPitchAngle - (1+bevelClearance)*bevelModule/bevelOuterCone

	// The field at a space center is the distance to the nearest flank, which
	// for a tooth of circular thickness pi*m/2 lies pi*m/4*cos(alpha) away in
	// the developed plane — mapped back onto the cone by cone*sin(arc/outer).
	halfThickness := math.Pi * bevelModule / 4 * math.Cos(bevelPressure)

	for _, cone := range []float64{25.0, 20.0} {
		tag := "@cone " + strconv.FormatFloat(cone, 'f', 1, 64) + " "

		near(tag+"flank at pitch cone", at("teeth", bevelPoint(cone, bevelPitchAngle, flankAt)), 0, 1e-3)
		near(tag+"tip cone", at("teeth", bevelPoint(cone, tipAngle, toothAt)), 0, 2e-3)
		near(tag+"root cone", at("teeth", bevelPoint(cone, rootAngle, spaceAt)), 0, 2e-3)

		// Signs: solid under a tooth, empty in a space, at the pitch cone.
		if v := at("teeth", bevelPoint(cone, bevelPitchAngle, toothAt)); v >= 0 {
			t.Errorf("%stooth center at the pitch cone is empty (%.5f)", tag, v)
		}
		near(tag+"space center at the pitch cone",
			at("teeth", bevelPoint(cone, bevelPitchAngle, spaceAt)),
			cone*math.Sin(halfThickness/bevelOuterCone), 1e-3)
	}

	// The bore is open all the way through the gear.
	heelZ, toeZ := bevelTipCorner(bevelOuterCone)[1], bevelTipCorner(bevelInnerCone)[1]
	for _, z := range []float64{heelZ - 0.5, 0.5 * (heelZ + toeZ), toeZ + 0.5} {
		if v := at("g", [3]float64{0, 0, z}); v <= 0 {
			t.Errorf("bore is not open at z = %.3f (%.5f)", z, v)
		}
	}
	// And there is web between the two ends to hold the bore.
	if v := at("g", [3]float64{7, 0, 0.5 * (heelZ + toeZ)}); v >= 0 {
		t.Errorf("web at mid-face is empty (%.5f)", v)
	}
}

// bevelTipCorner returns the (radius, z) of the outer tooth-tip corner at the
// end whose cone distance is `cone` — where the tip cone crosses that end's
// cone normal to the pitch cone. This is the datum both recesses are measured
// from, so it must not move when a recess changes.
func bevelTipCorner(cone float64) [2]float64 {
	addendumAngle := bevelModule / bevelOuterCone
	tipAngle := bevelPitchAngle + addendumAngle
	rho := cone / math.Cos(addendumAngle)
	return [2]float64{rho * math.Sin(tipAngle), rho * math.Cos(tipAngle)}
}

// TestBevelGearEnds checks the blank's ends, which are not two planes: each
// tooth end runs out on a cone normal to the pitch cone (the back cone at the
// heel, the front cone at the toe) and each of those is cut by a plane set a
// recess along the axis from that end's outer tooth-tip corner.
//
// The two caps cannot be booleaned alike, and that is what this test pins
// down. Both cones open toward the apex, so the back cone's inward
// continuation passes behind the blank while the front cone's runs through
// it — at the bore radius the front cone demands a z ABOVE the back face, so
// intersecting it would punch a hole around the bore. Hence heel = cone AND
// plane, toe = cone OR plane. TestBevelGearDimensions' web check is what
// fails if that is ever "simplified" to two intersections.
func TestBevelGearEnds(t *testing.T) {
	e := bevelEvaluator(t)
	ctx := context.Background()
	at := func(defs string, pts [][3]float64) []float64 {
		t.Helper()
		g, err := e.EvalGsl(ctx, nil, bevelDoc, defs, []string{"g"}, pts)
		if err != nil {
			t.Fatalf("eval: %v", err)
		}
		out := make([]float64, len(pts))
		for i := range pts {
			out[i] = float64(g[0].Samples[0][i].Value)
		}
		return out
	}
	// Points are taken on a tooth centerline, half an angular pitch off the
	// space that gearProfile centers on zero.
	onTooth := func(r, z float64) [3]float64 {
		phi := math.Pi / bevelTeethCount
		return [3]float64{r * math.Cos(phi), r * math.Sin(phi), z}
	}
	sinD, cosD := math.Sin(bevelPitchAngle), math.Cos(bevelPitchAngle)
	heel, toe := bevelTipCorner(bevelOuterCone), bevelTipCorner(bevelInnerCone)

	// kSmall rounds the tooth-end corner, pulling the tip corner itself off
	// the surface by exactly k/sqrt(2) at the right angle it turns through.
	const kSmall = 0.05
	chamfer := kSmall / math.Sqrt2

	for _, recess := range []float64{0.0, 2.0} {
		tag := "recess " + strconv.FormatFloat(recess, 'f', 1, 64) + ": "
		defs := "recess = " + strconv.FormatFloat(recess, 'f', 6, 64) + ";"

		// The flat faces sit a recess in +Z from their own tip corner: at the
		// heel that stands the back face proud of the teeth, at the toe it
		// sets the face back so the teeth project past it.
		faces := at(defs, [][3]float64{
			onTooth(7, heel[1]+recess),
			onTooth(7, toe[1]+recess),
		})
		for i, name := range []string{"heel flat face", "toe flat face"} {
			if math.Abs(faces[i]) > 2e-3 {
				t.Errorf("%s%s off the surface by %.5f", tag, name, faces[i])
			}
		}

		// The recess must NOT move the teeth: both tip corners stay on their
		// end cone at every recess, short only of the chamfer.
		corners := at(defs, [][3]float64{
			onTooth(heel[0], heel[1]),
			onTooth(toe[0], toe[1]),
		})
		for i, name := range []string{"heel tip corner", "toe tip corner"} {
			if math.Abs(corners[i]-chamfer) > 3e-3 {
				t.Errorf("%s%s reads %.5f, want the chamfer %.5f — the recess moved the teeth",
					tag, name, corners[i], chamfer)
			}
		}
	}

	// With a recess open, the end cones are the real tooth-end surfaces:
	// sample each halfway between where its plane stops binding and the tip
	// corner, on the cone, and require the field to vanish there.
	const recess = 2.0
	defs := "recess = " + strconv.FormatFloat(recess, 'f', 6, 64) + ";"
	heelCross := (bevelOuterCone - (heel[1]+recess)*cosD) / sinD
	toeCross := (bevelInnerCone - (toe[1]+recess)*cosD) / sinD
	rHeel := 0.5 * (heelCross + heel[0])
	rToe := 0.5 * (toeCross + toe[0])
	cones := at(defs, [][3]float64{
		onTooth(rHeel, (bevelOuterCone-rHeel*sinD)/cosD),
		onTooth(rToe, (bevelInnerCone-rToe*sinD)/cosD),
	})
	for i, name := range []string{"back cone", "front cone"} {
		if math.Abs(cones[i]) > 2e-3 {
			t.Errorf("%s is not the tooth-end surface at recess %.1f: field reads %.5f",
				name, recess, cones[i])
		}
	}
	// Independent of everything above: ending the teeth on the back cone is
	// what gives a bevel gear the outside radius the tables give it,
	// z*m/2 + m*cos(delta). The corner assertions already showed the solid
	// reaches that corner, so agreeing here ties the field to the standard
	// formula rather than to this construction's own derivation. The residual
	// is the spherical addendum, which is a shade over the nominal module.
	wantOuter := bevelTeethCount*bevelModule/2 + bevelModule*cosD
	if math.Abs(heel[0]-wantOuter) > 5e-3 {
		t.Errorf("outside radius %.4f, want the textbook %.4f", heel[0], wantOuter)
	}
	t.Logf("tip corners: heel (r %.4f, z %.4f), toe (r %.4f, z %.4f); cones sampled at r %.3f and %.3f",
		heel[0], heel[1], toe[0], toe[1], rHeel, rToe)
	t.Logf("outside radius %.4f vs textbook z*m/2 + m*cos(delta) = %.4f", heel[0], wantOuter)
}

// TestBevelGearStandoff checks the hub. A bevel gear's hub belongs on the
// heel, so unlike gear() — which measures its standoff from a mid-plane a
// bevel gear does not have — the height here is what the hub projects past
// the back face, and the cylinder starts at the toe face so it can never add
// material inside the blank.
func TestBevelGearStandoff(t *testing.T) {
	e := bevelEvaluator(t)
	ctx := context.Background()
	at := func(defs string, pts [][3]float64) []float64 {
		t.Helper()
		g, err := e.EvalGsl(ctx, nil, bevelDoc, defs, []string{"g"}, pts)
		if err != nil {
			t.Fatalf("eval: %v", err)
		}
		out := make([]float64, len(pts))
		for i := range pts {
			out[i] = float64(g[0].Samples[0][i].Value)
		}
		return out
	}

	const (
		recess   = 2.0
		hubDiam  = 14.0
		hubHigh  = 4.0
		chamfer  = 0.4
		clearOf  = 9.0 // a radius outside the hub, still on the back face
		underHub = 5.5 // a radius inside the hub footprint
	)
	backFace := bevelTipCorner(bevelOuterCone)[1] + recess
	hubTop := backFace + hubHigh
	hubRadius := hubDiam / 2

	hub := "hubDiam = 14.0; hubHeight = 4.0; hubChamfer = 0.4;"
	v := at(hub, [][3]float64{
		{underHub, 0, hubTop},        // the hub's top face
		{hubRadius, 0, backFace + 2}, // its cylindrical side, mid-projection
		{underHub, 0, hubTop + 1},    // clear above it
		{underHub, 0, hubTop - 1},    // inside it
		{clearOf, 0, backFace},       // back face, outside the hub footprint
		{0, 0, hubTop - 1},           // bore, through the hub
	})
	for i, c := range []struct {
		name string
		want float64
		tol  float64
	}{
		{"hub top face", 0, 2e-3},
		{"hub side", 0, 2e-3},
		{"back face clear of the hub", 0, 2e-3},
	} {
		idx := []int{0, 1, 4}[i]
		if math.Abs(v[idx]-c.want) > c.tol {
			t.Errorf("%s reads %.5f, want %.5f", c.name, v[idx], c.want)
		}
	}
	if v[2] <= 0 {
		t.Errorf("point above the hub is solid (%.5f)", v[2])
	}
	if v[3] >= 0 {
		t.Errorf("point inside the hub is empty (%.5f)", v[3])
	}
	if v[5] <= 0 {
		t.Errorf("bore is not open through the hub (%.5f)", v[5])
	}

	// Height and chamfer both zero is the way to ask for no hub, and it has to
	// leave the blank alone.
	none := at("hubDiam = 14.0; hubHeight = 0.0; hubChamfer = 0.0;", [][3]float64{
		{underHub, 0, backFace}, {clearOf, 0, backFace},
	})
	for i, name := range []string{"under the footprint", "clear of it"} {
		if math.Abs(none[i]) > 2e-3 {
			t.Errorf("no-hub back face %s reads %.5f, want 0", name, none[i])
		}
	}

	// With a chamfer but no height the hub's top is flush with the back face,
	// and blending two coincident planes stands the face proud by the usual
	// chamfer/sqrt(2). Documented on bevelGear; pinned here so it stays a
	// known quantity rather than a surprise.
	flush := at("hubDiam = 14.0; hubHeight = 0.0; hubChamfer = 0.4;", [][3]float64{
		{underHub, 0, backFace}, {clearOf, 0, backFace},
	})
	if want := -chamfer / math.Sqrt2; math.Abs(flush[0]-want) > 5e-3 {
		t.Errorf("flush hub blend reads %.5f under the footprint, want %.5f", flush[0], want)
	}
	if math.Abs(flush[1]) > 2e-3 {
		t.Errorf("flush hub blend disturbed the back face clear of it (%.5f)", flush[1])
	}
	t.Logf("hub: back face z %.4f, top z %.4f, radius %.1f; flush-blend bulge %.5f",
		backFace, hubTop, hubRadius, flush[0])
}

// TestBevelSpiralAngle measures the angle the tooth trace actually makes with
// the cone element, rather than trusting the shear that produced it. A
// constant-angle (logarithmic) spiral has the flank at a fixed offset in
// ln(coneDist), so sin(pitchAngle) * dPhi / dLnCone is exactly tan of the
// requested spiral angle. The two cone distances are deliberately close: the
// tooth pattern repeats every pi/numTeeth in azimuth, so a coarser step would
// let the measured shift alias onto a neighbouring tooth and read a plausible
// but wrong angle.
func TestBevelSpiralAngle(t *testing.T) {
	e := bevelEvaluator(t)
	ctx := context.Background()

	const (
		coneNear = 27.5
		coneFar  = 28.0
		phiLo    = -0.20
		phiStep  = 0.0025
		phiCount = 161
	)

	cases := []struct {
		name   string
		spiral float64
	}{
		{"right hand", 0.6},
		{"left hand", -0.5},
		{"straight", 0.0},
	}

	for _, tc := range cases {
		defs := "probe = gears.bevelTeeth(20.0, 2.0, 0.35, 0.7853981633974483, " +
			strconv.FormatFloat(tc.spiral, 'f', 6, 64) + ", 0.0, 0.0, 0.25, 0.05);"

		var points [][3]float64
		for _, cone := range []float64{coneNear, coneFar} {
			for i := 0; i < phiCount; i++ {
				points = append(points, bevelPoint(cone, bevelPitchAngle, phiLo+phiStep*float64(i)))
			}
		}
		g, err := e.EvalGsl(ctx, nil, bevelDoc, defs, []string{"probe"}, points)
		if err != nil {
			t.Fatalf("%s: eval: %v", tc.name, err)
		}
		// Samples is indexed [component][point], and probe is a scalar.
		vals := make([]float64, len(points))
		for i := range points {
			vals[i] = float64(g[0].Samples[0][i].Value)
		}

		// The flank crossing nearest azimuth zero, on each cone distance.
		crossing := func(seg []float64) (float64, bool) {
			best, found := 0.0, false
			for i := 0; i+1 < len(seg); i++ {
				a, b := seg[i], seg[i+1]
				if a < 0 && b >= 0 {
					phi := phiLo + phiStep*(float64(i)+(0-a)/(b-a))
					if !found || math.Abs(phi) < math.Abs(best) {
						best, found = phi, true
					}
				}
			}
			return best, found
		}
		phiNear, okNear := crossing(vals[:phiCount])
		phiFar, okFar := crossing(vals[phiCount:])
		if !okNear || !okFar {
			t.Fatalf("%s: no flank crossing found near azimuth zero", tc.name)
		}

		tanBeta := math.Sin(bevelPitchAngle) * (phiFar - phiNear) / (math.Log(coneFar) - math.Log(coneNear))
		got := math.Atan(tanBeta)
		if math.Abs(got-tc.spiral) > 1e-3 {
			t.Errorf("%s: tooth trace is at %.5f rad to the cone element, want %.5f",
				tc.name, got, tc.spiral)
		}
		t.Logf("%s: requested %+.4f rad, measured %+.5f rad", tc.name, tc.spiral, got)
	}
}

// TestBevelPairMesh is the phase check, and the reason bevelPair's constant
// and its opposite-hand rule can be trusted. Across a full tooth pitch of
// rotation the mesh clearance must stay positive AND essentially constant:
// rolling contact cannot make the gap depend on phase, so a wrong phase or a
// wrong tooth ratio shows up as variation long before it shows up as
// interference. The same-hand case is the control — an identical pair that
// differs only in gear 2's hand buries its teeth more than half a module deep
// at every phase.
func TestBevelPairMesh(t *testing.T) {
	e := bevelEvaluator(t)
	ctx := context.Background()

	// Both gears live within a cone distance of the origin, whichever way
	// their axes point.
	region := [6]float64{-32, -32, -32, 32, 32, 32}

	cases := []struct {
		name    string
		expr    string
		spread  float64
		meshing bool
	}{
		{"straight 20:20 at 90 degrees", "max(straightA, straightB)", 0.005, true},
		{"spiral 20:20 at 90 degrees", "max(spiralA, spiralB)", 0.005, true},
		{"straight 15:25 at 90 degrees", "max(ratioA, ratioB)", 0.005, true},
		{"same-hand spiral 20:20 (wrong)", "max(sameA, sameB)", 0, false},
	}

	const steps = 5
	for _, tc := range cases {
		minGap, maxGap := math.Inf(1), math.Inf(-1)
		for i := 0; i < steps; i++ {
			phase := float64(i) / float64(steps-1) // one full tooth pitch
			defs := "phase = " + strconv.FormatFloat(phase, 'f', 6, 64) + ";"
			r, err := e.Extremum(ctx, nil, bevelDoc, defs, tc.expr, region, "min", 0, 0)
			if err != nil {
				t.Fatalf("%s: extremum at phase %.3f: %v", tc.name, phase, err)
			}
			if tc.meshing && r.Min <= 0 {
				t.Errorf("%s: gears interfere at phase %.3f: gap %.5f", tc.name, phase, r.Min)
			}
			if !tc.meshing && r.Min >= 0 {
				t.Errorf("%s: expected interference at phase %.3f, got clearance %.5f",
					tc.name, phase, r.Min)
			}
			minGap, maxGap = math.Min(minGap, r.Min), math.Max(maxGap, r.Min)
		}
		if tc.meshing {
			if spread := maxGap - minGap; spread > tc.spread {
				t.Errorf("%s: mesh clearance varies with phase (%.5f … %.5f, spread %.5f) — the pair is not rolling in step",
					tc.name, minGap, maxGap, spread)
			}
			t.Logf("%s: clearance %.5f … %.5f across one tooth pitch", tc.name, minGap, maxGap)
		} else {
			t.Logf("%s: interference %.5f … %.5f across one tooth pitch", tc.name, minGap, maxGap)
		}
	}
}

// bevelToothDoc builds the same pair twice: through the library's bevelTooth
// factory, and straight from bevelPair. The factory is defined in gears and
// called from here, so its closure crosses a package boundary, and the values
// it captures are this document's VARIABLES rather than literals — the case
// that synthesizing the closure into either file alone cannot express. A
// second factory at a coarser module checks that the one lifted function the
// two share still gives each its own captures.
const bevelToothDoc = `
shaftAngle = 1.5707963267948966;
faceWidth = 12.0;
phase = 0.0;
tooth = gears.bevelTooth(2.0, 0.35, shaftAngle, 0.6, faceWidth, 2.0, 2.0, 0.1, 4.0, 0.4, 0.25, 0.05);
coarse = gears.bevelTooth(3.0, 0.35, shaftAngle, 0.6, faceWidth, 2.0, 2.0, 0.1, 4.0, 0.4, 0.25, 0.05);
rot1 = 2 * pi * phase / 20.0;
rot2 = pi * 19.0 / 20.0 - 2 * pi * phase / 20.0;
f1 = tooth(20.0, 20.0, 1.0, rot1, 14.0, 8.0);
f2local = tooth(20.0, 20.0, -1.0, rot2, 14.0, 8.0);
f2 = remap(f2local, [neg(p.z), p.y, p.x]);
c1 = coarse(20.0, 20.0, 1.0, rot1, 14.0, 8.0);
f1.color = [0.8, 0.2, 0.2];
f2.color = [0.2, 0.4, 0.8];
d1, d2 = gears.bevelPair(20.0, 20.0, 2.0, 0.35, shaftAngle, 0.6, faceWidth, 2.0, 2.0, 0.1, 14.0, 14.0, 4.0, 0.4, 8.0, 8.0, 0.25, 0.05, phase);
return f1, f2;
`

// TestBevelToothFactory pins the library factory to bevelPair: the two gears
// it makes must match bevelPair's EXACTLY, not within a tolerance, since they
// are meant to be the same evaluation reached by another route.
func TestBevelToothFactory(t *testing.T) {
	e := bevelEvaluator(t)
	ctx := context.Background()

	var pts [][3]float64
	coords := []float64{-22, -11, -3, 0, 3, 11, 22}
	for _, x := range coords {
		for _, y := range coords {
			for _, z := range coords {
				pts = append(pts, [3]float64{x, y, z})
			}
		}
	}
	exprs := []string{"f1 - d1", "f2 - d2", "f1", "c1 - f1"}
	g, err := e.EvalGsl(ctx, nil, bevelToothDoc, "", exprs, pts)
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	maxAbs := func(k int) float64 {
		m := 0.0
		for _, s := range g[k].Samples[0] {
			m = math.Max(m, math.Abs(float64(s.Value)))
		}
		return m
	}
	if d := maxAbs(0); d != 0 {
		t.Errorf("gear 1 from bevelTooth differs from bevelPair by up to %.3e", d)
	}
	if d := maxAbs(1); d != 0 {
		t.Errorf("gear 2 from bevelTooth differs from bevelPair by up to %.3e", d)
	}
	// The comparison means nothing if the field is trivial: the gear has to be
	// solid somewhere and empty somewhere in the grid.
	lo, hi := math.Inf(1), math.Inf(-1)
	for _, s := range g[2].Samples[0] {
		lo, hi = math.Min(lo, float64(s.Value)), math.Max(hi, float64(s.Value))
	}
	if !(lo < 0 && hi > 0) {
		t.Errorf("gear 1 ranges %.4f … %.4f over the grid; want both signs", lo, hi)
	}
	// Two instantiations of the factory share one lifted function and must
	// still differ: the module is a capture, not a constant of the function.
	if d := maxAbs(3); d == 0 {
		t.Errorf("a module-3 gear from the same factory is identical to the module-2 one; the captures are not reaching the shared function")
	}
}
