//go:build cgo

package eval

import (
	"math"
	"os"
	"testing"
)

// Shaft Modeling, Phase 3: segments whose section is not round. hex, dFlat,
// doubleD and the male spline are exact sections, held to a reference sampled
// densely from the same lines and arcs; the gear is a bound, held to
// gears.gearProfile and to a Lipschitz check.

// sectionArc appends points on the circle of radius r from angle a0 to a1
// (a1 > a0), at most sectionStep apart, each end included when asked: a chord
// that short lies within r * (step / r)^2 / 8 of its arc, under 1e-7 mm here.
// An end the loop already holds is left out, so no edge has length zero; an
// end it does not, a flat's corner, has to be in, or the straight edge after
// it starts a step along the arc.
func sectionArc(pg poly2, r, a0, a1 float64, withStart, withEnd bool) poly2 {
	const sectionStep = 0.002
	n := int(math.Ceil((a1 - a0) * r / sectionStep))
	for i := 0; i <= n; i++ {
		if (i == 0 && !withStart) || (i == n && !withEnd) {
			continue
		}
		a := a0 + (a1-a0)*float64(i)/float64(n)
		pg = append(pg, [2]float64{r * math.Cos(a), r * math.Sin(a)})
	}
	return pg
}

// hexSection has its flats on +-x, as drill.hexXY's do: corners at 30 + 60k
// degrees.
func hexSection(acrossFlats float64) poly2 {
	r := acrossFlats / math.Sqrt(3)
	var pg poly2
	for k := 0; k < 6; k++ {
		a := math.Pi/6 + math.Pi/3*float64(k)
		pg = append(pg, [2]float64{r * math.Cos(a), r * math.Sin(a)})
	}
	return pg
}

// dFlatSection is a disc of diameter dia with one flat on +x, acrossFlat from
// the flat to the far side.
func dFlatSection(dia, acrossFlat float64) poly2 {
	r := dia / 2
	c := acrossFlat - r
	a0 := math.Atan2(math.Sqrt(r*r-c*c), c)
	return sectionArc(nil, r, a0, 2*math.Pi-a0, true, true)
}

// doubleDSection is a disc of diameter dia with flats on +-x, acrossFlats apart.
func doubleDSection(dia, acrossFlats float64) poly2 {
	r := dia / 2
	a := acrossFlats / 2
	a0 := math.Atan2(math.Sqrt(r*r-a*a), a)
	pg := sectionArc(nil, r, a0, math.Pi-a0, true, true)
	return sectionArc(pg, r, math.Pi+a0, 2*math.Pi-a0, true, true)
}

// splineSection is a straight-sided male spline: count teeth `width` across,
// on +x and every 2pi/count, from the minor circle out to the major.
func splineSection(count int, minor, major, width float64) poly2 {
	rm, rM, h := minor/2, major/2, width/2
	aM := math.Asin(h / rM)
	am := math.Asin(h / rm)
	pitch := 2 * math.Pi / float64(count)
	var pg poly2
	for k := 0; k < count; k++ {
		phi := pitch * float64(k)
		rot := func(x, y float64) [2]float64 {
			return [2]float64{x*math.Cos(phi) - y*math.Sin(phi), x*math.Sin(phi) + y*math.Cos(phi)}
		}
		// Up the leading flank from the minor circle, round the tip arc (its
		// ends are the flanks' tops), down the trailing flank, and round the
		// root arc to the next tooth (its ends are the flanks' feet).
		pg = append(pg, rot(math.Sqrt(rm*rm-h*h), -h))
		pg = sectionArc(pg, rM, phi-aM, phi+aM, true, true)
		pg = append(pg, rot(math.Sqrt(rm*rm-h*h), h))
		pg = sectionArc(pg, rm, phi+am, phi+pitch-am, false, false)
	}
	return pg
}

type sectionCase struct {
	name  string
	gsl   string
	pg    poly2
	exact bool // exact inside too: a single exact prism; the spline is a union
}

func sectionCases() []sectionCase {
	return []sectionCase{
		{"hex", "shaft.hex(20.0, 0.0, 30.0)", hexSection(20), true},
		{"d-flat", "shaft.dFlat(16.0, 13.0, 0.0, 30.0)", dFlatSection(16, 13), true},
		{"double-d", "shaft.doubleD(16.0, 12.0, 0.0, 30.0)", doubleDSection(16, 12), true},
		{"spline", "shaft.spline(6.0, 16.0, 20.0, 4.0, 0.0, 30.0)", splineSection(6, 16, 20, 4), false},
		{"section of a hex", "shaft.section(" + hexProfileGSL(20) + ", 0.0, 30.0)", hexPolygon(20), true},
	}
}

// Each section, alone, against the reference: exact outside, and inside exact
// for a single prism and a bound for the spline's union of teeth and disc.
func TestShaftSections_MatchReference(t *testing.T) {
	for _, c := range sectionCases() {
		t.Run(c.name, func(t *testing.T) {
			points := shaftGrid3([3]float64{-12, -12, 0}, [3]float64{12, 12, 30}, 4, 24)
			got := evalStdShaft(t, "@scene(140)\nsolid = shaft.body("+c.gsl+");\nreturn solid;\n", points)
			worstOut, worstIn, failures := 0.0, 0.0, 0
			for i, q := range points {
				want := shaftExtruded(c.pg.distance([2]float64{q[0], q[1]}), q[2], 0, 30)
				var bad bool
				if want > 0 || c.exact {
					miss := math.Abs(got[i] - want)
					if want > 0 {
						worstOut = math.Max(worstOut, miss)
					} else {
						worstIn = math.Max(worstIn, miss)
					}
					bad = miss > shaftTolerance
				} else {
					worstIn = math.Max(worstIn, got[i]-want)
					bad = got[i] < want-shaftTolerance || (want < -shaftTolerance && got[i] >= 0)
				}
				if bad {
					if failures++; failures <= 3 {
						t.Errorf("at %v: field %v, reference %v", q, got[i], want)
					}
				}
			}
			t.Logf("outside worst miss %.3g mm; inside worst %.3g mm; %d failures over %d points", worstOut, worstIn, failures, len(points))
		})
	}
}

// A hex beside a round between its flats and its corners, in both orders,
// through the real package: exact outside, solid on the joint plane.
func TestShaftSections_MixedConvexity(t *testing.T) {
	hex := hexSection(20)
	for _, hexBelow := range []bool{true, false} {
		name := map[bool]string{true: "hex below", false: "hex above"}[hexBelow]
		t.Run(name, func(t *testing.T) {
			hexLo, hexHi, rLo, rHi := 0.0, 20.0, 20.0, 40.0
			segs := "[shaft.hex(20.0, 0.0, 20.0), shaft.diameter(22.0, 20.0, 40.0)]"
			if !hexBelow {
				hexLo, hexHi, rLo, rHi = 20, 40, 0, 20
				segs = "[shaft.diameter(22.0, 0.0, 20.0), shaft.hex(20.0, 20.0, 40.0)]"
			}
			ref := func(q [3]float64) float64 {
				return math.Min(shaftExtruded(hex.distance([2]float64{q[0], q[1]}), q[2], hexLo, hexHi),
					shaftExtruded(math.Hypot(q[0], q[1])-11, q[2], rLo, rHi))
			}
			points := shaftGrid3([3]float64{-12, -12, 0}, [3]float64{12, 12, 40}, 4, 24)
			points = append(points, jointPoints([]float64{20}, []float64{10})...)
			got := evalStdShaft(t, "@scene(140)\nsolid = shaft.body("+segs+");\nreturn solid;\n", points)
			worst, failures := 0.0, 0
			for i, q := range points {
				want := ref(q)
				if want > 0 {
					miss := math.Abs(got[i] - want)
					worst = math.Max(worst, miss)
					if miss > shaftTolerance {
						if failures++; failures <= 3 {
							t.Errorf("outside at %v: field %v, reference %v", q, got[i], want)
						}
					}
				} else if want < -1e-3 && got[i] > -1e-3 {
					if failures++; failures <= 3 {
						t.Errorf("inside at %v: field %v, reference %v (not solid)", q, got[i], want)
					}
				}
			}
			t.Logf("outside worst miss %.3g mm; %d failures over %d points", worst, failures, len(points))
		})
	}
}

// The male spline seats in drill's female spline of the same size: they
// touch (the overlap's deepest point is about 0) and never interfere (it is
// never below -0.002), the fit test the driver bits use.
func TestShaftSections_SplineFitsDrillSpline(t *testing.T) {
	doc := `@scene(140)
male = shaft.body(shaft.spline(6.0, 16.0, 20.0, 4.0, 0.0, 30.0));
hole = drill.hole([drill.thru(16.0), drill.spline(6.0, 20.0, 4.0)], 40.0);
overlap = max(male, neg(hole));
return male, hole, overlap;
`
	points := shaftGrid3([3]float64{-11, -11, 2}, [3]float64{11, 11, 28}, 0, 28)
	got := evalShaftNamed(t, doc, "overlap", points)
	deepest := math.Inf(1)
	for _, v := range got {
		deepest = math.Min(deepest, v)
	}
	t.Logf("overlap's deepest point %.4g mm", deepest)
	if deepest < -0.002 {
		t.Errorf("male spline interferes with the female by %.4g mm", -deepest)
	}
	if deepest > 0.05 {
		t.Errorf("male spline does not reach the female: nearest approach %.4g mm", deepest)
	}
}

func evalShaftNamed(t *testing.T, doc, name string, points [][3]float64) []float64 {
	t.Helper()
	e := New(nil, nil)
	groups, err := e.EvalGsl(t.Context(), nil, doc, "", []string{name}, points)
	if err != nil {
		t.Fatalf("eval %s: %v", name, err)
	}
	out := make([]float64, len(points))
	for i := range points {
		out[i] = float64(groups[0].Samples[0][i].Value)
	}
	return out
}

// A spur gear segment reads gears.gearProfile across its span, clear of its
// ends: the same section the gears package draws.
func TestShaftSections_SpurGearIsGearProfile(t *testing.T) {
	doc := `@scene(140)
solid = shaft.body(shaft.gear(17.0, 1.0, radians(20.0), 0.0, 0.0, 30.0));
profile = gears.gearProfile(17.0, 1.0, radians(20.0), 0.0, 0.25, 0.05);
return solid, profile;
`
	var points [][3]float64
	for _, q := range shaftGrid3([3]float64{-10, -10, 15}, [3]float64{10, 10, 15}, 0, 30) {
		points = append(points, [3]float64{q[0], q[1], 15})
	}
	solid := evalShaftNamed(t, doc, "solid", points)
	profile := evalShaftNamed(t, doc, "profile", points)
	for i, q := range points {
		if math.Abs(profile[i]) < 10 && math.Abs(solid[i]-profile[i]) > shaftTolerance {
			t.Fatalf("at %v: gear segment %v, gearProfile %v", q, solid[i], profile[i])
		}
	}
}

// gearSlope is the steepest gradient of a gear segment's field over sampled
// points within rhoMax of the axis, by forward differences along x, y and z
// together. A twist adds a z component on top of the in-plane gradient, so a
// difference along any one direction alone would miss the stretch.
func gearSlope(t *testing.T, helixAngle string, rhoMax float64) float64 {
	t.Helper()
	doc := "@scene(140)\nsolid = shaft.body(shaft.gear(17.0, 1.0, radians(20.0), " + helixAngle + ", 0.0, 30.0));\nreturn solid;\n"
	base := shaftGrid3([3]float64{-11, -11, 2}, [3]float64{11, 11, 28}, 0, 24)
	var points [][3]float64
	const step = 0.005
	for _, q := range base {
		if math.Hypot(q[0], q[1]) > rhoMax {
			continue
		}
		points = append(points, q,
			[3]float64{q[0] + step, q[1], q[2]}, [3]float64{q[0], q[1] + step, q[2]}, [3]float64{q[0], q[1], q[2] + step})
	}
	got := evalStdShaft(t, doc, points)
	worst := 0.0
	for i := 0; i < len(points); i += 4 {
		gx, gy, gz := (got[i+1]-got[i])/step, (got[i+2]-got[i])/step, (got[i+3]-got[i])/step
		worst = math.Max(worst, math.Sqrt(gx*gx+gy*gy+gz*gz))
	}
	return worst
}

// A helix adds no stretch: a helical gear segment's field changes no faster
// than the spur gear's it twists. The twist alone would stretch it by up to
// sqrt(1 + (tipRadius * twistRate)^2), which gearSection divides out.
//
// The spur's own slope is gears.gearProfile's, which is not yet a bound: its
// involute flanks read the string length past the tooth's tip, beyond their
// normal fans, where it overstates the distance (1.56 measured). That is fixed
// apart, in gears; once it is, the gear segment is a bound and this can ask
// for 1 outright.
func TestShaftSections_HelixAddsNoStretch(t *testing.T) {
	// Inside the tip circle, where the teeth are and the twist moves them
	// fastest; beyond it gearProfile's own overstatement sets the slope.
	spur := gearSlope(t, "0.0", 9.4)
	helical := gearSlope(t, "radians(30.0)", 9.4)
	t.Logf("steepest sampled slope: spur %.4f, helical %.4f", spur, helical)
	if helical > spur*1.01 {
		t.Errorf("the helix stretches the gear's field: %.4f against the spur's %.4f", helical, spur)
	}
}

// A gear between two diameters that lie between its root and tip circles:
// on both joint planes, wherever both sides are solid, the shaft is solid.
func TestShaftSections_GearMixedConvexity(t *testing.T) {
	// 17 teeth of module 1: tip diameter 19, root 14.5. The rounds are 17.
	doc := `@scene(140)
solid = shaft.body([shaft.diameter(17.0, 0.0, 10.0), shaft.gear(17.0, 1.0, radians(20.0), 0.0, 10.0, 25.0), shaft.diameter(17.0, 25.0, 35.0)]);
profile = gears.gearProfile(17.0, 1.0, radians(20.0), 0.0, 0.25, 0.05);
return solid, profile;
`
	var points [][3]float64
	for _, z := range []float64{10, 25} {
		for x := -8.4; x <= 8.4; x += 0.35 {
			for y := -8.4; y <= 8.4; y += 0.35 {
				points = append(points, [3]float64{x, y, z})
			}
		}
	}
	solid := evalShaftNamed(t, doc, "solid", points)
	profile := evalShaftNamed(t, doc, "profile", points)
	checked, bad := 0, 0
	for i, q := range points {
		insideBoth := profile[i] < -0.05 && math.Hypot(q[0], q[1]) < 8.45
		if !insideBoth {
			continue
		}
		checked++
		if solid[i] > -1e-3 {
			if bad++; bad <= 3 {
				t.Errorf("joint point %v inside the gear and the round reads %v", q, solid[i])
			}
		}
	}
	t.Logf("%d joint points inside both sides, %d not solid", checked, bad)
	if checked == 0 {
		t.Fatal("no joint point lies inside both sides")
	}
}

// The pinion example compiles to WGSL (its size logged as Phase 3's budget)
// and evaluates: solid through the pinion, the hex and the spline's core.
func TestShaftSections_PinionExampleCompilesToWGSL(t *testing.T) {
	src, err := os.ReadFile("../examples/shaft_pinion.gsl")
	if err != nil {
		t.Fatal(err)
	}
	e := New(nil, nil)
	res, err := e.Shader(t.Context(), nil, string(src), "", ShaderOptions{Target: "wgsl", OutDir: t.TempDir()})
	if err != nil {
		t.Fatalf("shader: %v", err)
	}
	t.Logf("examples/shaft_pinion.gsl: %d bytes of WGSL", len(res.Code))
	groups, err := e.EvalGsl(t.Context(), nil, string(src), "", []string{"pinionShaft"},
		[][3]float64{{4, 0, 44}, {5, 0, 75}, {5, 0, 9}, {0, 0, 100}})
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	got := groups[0].Samples[0]
	if got[0].Value >= 0 || got[1].Value >= 0 || got[2].Value >= 0 || got[3].Value <= 0 {
		t.Errorf("pinion %v, hex %v, spline %v, beyond the end %v: want solid three times, then empty",
			got[0].Value, got[1].Value, got[2].Value, got[3].Value)
	}
}
