//go:build cgo

package eval

import (
	"fmt"
	"math"
	"os"
	"testing"
)

// Shaft Modeling, Phase 4: end treatments. A non-round segment folds in as its
// round blank and cuts its teeth or gaps afterwards (spec 3.4). Past an end
// with a cutter runout the cut runs on into the neighbour, its floor rising
// from the root along the cutter's arc, centred at (root + Rc, the end plane)
// in the half-section; within the cutter's reach the neighbour is cut to the
// segment's profile. A chamfered end chamfers the blank before the teeth.

// radialSurface finds, for each z, the radius at which the field along the
// ray from the axis at angle `angle` first turns from solid to empty, by
// bisection between the axis and rMax, all z at once.
func radialSurface(t *testing.T, doc string, angle float64, zs []float64, rMax float64) []float64 {
	t.Helper()
	lo, hi := make([]float64, len(zs)), make([]float64, len(zs))
	for i := range zs {
		hi[i] = rMax
	}
	c, s := math.Cos(angle), math.Sin(angle)
	for iter := 0; iter < 30; iter++ {
		pts := make([][3]float64, len(zs))
		for i, z := range zs {
			r := (lo[i] + hi[i]) / 2
			pts[i] = [3]float64{r * c, r * s, z}
		}
		f := evalShaftNamed(t, doc, "solid", pts)
		for i := range zs {
			if f[i] < 0 {
				lo[i] = (lo[i] + hi[i]) / 2
			} else {
				hi[i] = (lo[i] + hi[i]) / 2
			}
		}
	}
	out := make([]float64, len(zs))
	for i := range zs {
		out[i] = (lo[i] + hi[i]) / 2
	}
	return out
}

// gearGapAndTooth finds the angles of a gap's centre and a tooth's centre in
// the first pitch of the section the shaft's gear draws, by reading
// gears.gearProfile round its pitch circle.
func gearGapAndTooth(t *testing.T, numTeeth, module float64) (gap, tooth float64) {
	t.Helper()
	doc := fmt.Sprintf("@scene(140)\nprofile = gears.gearProfile(%v, %v, radians(20.0), 0.0, 0.25, %v);\nreturn profile;\n",
		numTeeth, module, 0.05*module)
	pitchR := numTeeth * module / 2
	period := 2 * math.Pi / numTeeth
	const n = 3600
	pts := make([][3]float64, n)
	for i := range pts {
		a := period * float64(i) / n
		pts[i] = [3]float64{pitchR * math.Cos(a), pitchR * math.Sin(a), 0}
	}
	f := evalShaftNamed(t, doc, "profile", pts)
	hiV, loV := math.Inf(-1), math.Inf(1)
	for i, v := range f {
		a := period * float64(i) / n
		if v > hiV {
			hiV, gap = v, a
		}
		if v < loV {
			loV, tooth = v, a
		}
	}
	return gap, tooth
}

// cutterFloor is the floor of a runout at distance d past the end plane: the
// cutter's arc from the root, clipped to the neighbour's radius.
func cutterFloor(root, cutterR, neighbourR, d float64) float64 {
	if d >= cutterR {
		return neighbourR
	}
	return math.Min(neighbourR, root+cutterR-math.Sqrt(cutterR*cutterR-d*d))
}

func checkRadii(t *testing.T, what string, zs, got []float64, want func(z float64) float64, tol float64) {
	t.Helper()
	bad := 0
	worst := 0.0
	for i, z := range zs {
		miss := math.Abs(got[i] - want(z))
		worst = math.Max(worst, miss)
		if miss > tol {
			if bad++; bad <= 3 {
				t.Errorf("%s at z %.2f: surface at r %.4f, want %.4f", what, z, got[i], want(z))
			}
		}
	}
	t.Logf("%s: worst miss %.2g mm over %d heights", what, worst, len(zs))
}

func spanZs(from, to, step float64) []float64 {
	var zs []float64
	for z := from + step/2; z < to; z += step {
		zs = append(zs, z)
	}
	return zs
}

// A hobbed gear, 20 teeth of module 2 (root 17.5, tip 22), from z 20 to 50
// between a journal of 30 and a shoulder of 50, run out at its top with a hob
// of 80. In a gap's mid-plane the floor rises along the hob's arc into the
// shoulder and breaks out at its diameter, 23.3 mm on; in a tooth's plane the
// shoulder is cut to the tip circle until the arc rises above it. The bottom
// end is a plane: the journal is whole.
func TestShaftEnds_HobRunoutFollowsTheCutter(t *testing.T) {
	doc := `@scene(140)
solid = shaft.body([shaft.diameter(30.0, 0.0, 20.0),
                    shaft.gear(20.0, 2.0, radians(20.0), 0.0, 20.0, 50.0, shaft.planeEnd(), shaft.runout(80.0)),
                    shaft.diameter(50.0, 50.0, 80.0)]);
return solid;
`
	gap, tooth := gearGapAndTooth(t, 20, 2)
	const root, tip, journal, shoulder, rc = 17.5, 22.0, 15.0, 25.0, 40.0
	zs := spanZs(0, 80, 0.5)
	want := func(inGap bool) func(z float64) float64 {
		return func(z float64) float64 {
			switch {
			case z < 20:
				return journal
			case z < 50:
				if inGap {
					return root
				}
				return tip
			}
			floor := cutterFloor(root, rc, shoulder, z-50)
			if inGap {
				return floor
			}
			return math.Min(shoulder, math.Max(tip, floor))
		}
	}
	checkRadii(t, "gap", zs, radialSurface(t, doc, gap, zs, 30), want(true), 1e-3)
	checkRadii(t, "tooth", zs, radialSurface(t, doc, tooth, zs, 30), want(false), 1e-3)
	checkRunoutIsVoid(t, doc, gap, root, rc, shoulder, 50)
}

// checkRunoutIsVoid: in a gap's mid-plane, points of the runout band (above
// the cutter's floor and inside the neighbour's diameter, at least 0.5 mm
// clear of both) read empty, a positive field, and not a zero sheet: a surface
// test that counts f < 0 as solid cannot tell -0 from a void.
func checkRunoutIsVoid(t *testing.T, doc string, gap, root, rc, neighbourR, end float64) {
	t.Helper()
	c, s := math.Cos(gap), math.Sin(gap)
	var pts [][3]float64
	for d := 0.5; d < rc; d += 0.5 {
		floor := cutterFloor(root, rc, neighbourR, d)
		for r := floor + 0.5; r < neighbourR-0.5; r += 0.25 {
			pts = append(pts, [3]float64{r * c, r * s, end + d})
		}
	}
	if len(pts) == 0 {
		t.Fatal("no points in the runout band")
	}
	f := evalShaftNamed(t, doc, "solid", pts)
	bad := 0
	for i, q := range pts {
		if f[i] <= 0.1 {
			if bad++; bad <= 3 {
				t.Errorf("runout band at %v reads %v, not a void", q, f[i])
			}
		}
	}
	t.Logf("runout band: %d points, %d not reading empty", len(pts), bad)
}

// A milled spline, 6 teeth 4 wide from a minor of 16 to a major of 20, from z
// 10 to 30 between a round of 14 and a shoulder of 26, run out at its top
// with a disc cutter of 60. The same checks, at a gap's centre (30 degrees)
// and a tooth's (0).
func TestShaftEnds_SplineRunoutFollowsTheCutter(t *testing.T) {
	doc := `@scene(140)
solid = shaft.body([shaft.diameter(14.0, 0.0, 10.0),
                    shaft.spline(6.0, 16.0, 20.0, 4.0, 10.0, 30.0, shaft.planeEnd(), shaft.runout(60.0)),
                    shaft.diameter(26.0, 30.0, 50.0)]);
return solid;
`
	const minor, major, below, shoulder, rc = 8.0, 10.0, 7.0, 13.0, 30.0
	zs := spanZs(0, 50, 0.5)
	want := func(inGap bool) func(z float64) float64 {
		return func(z float64) float64 {
			switch {
			case z < 10:
				return below
			case z < 30:
				if inGap {
					return minor
				}
				return major
			}
			floor := cutterFloor(minor, rc, shoulder, z-30)
			if inGap {
				return floor
			}
			return math.Min(shoulder, math.Max(major, floor))
		}
	}
	checkRadii(t, "gap", zs, radialSurface(t, doc, math.Pi/6, zs, 16), want(true), 1e-3)
	checkRadii(t, "tooth", zs, radialSurface(t, doc, 0, zs, 16), want(false), 1e-3)
	checkRunoutIsVoid(t, doc, math.Pi/6, minor, rc, shoulder, 30)
}

// With plane ends, the default, every Phase 3 segment keeps its surface: the
// field's sign matches the reference prism's, alone and between a smaller
// and a larger neighbour, wherever the reference is clear of zero.
func TestShaftEnds_PlaneEndsKeepTheSurface(t *testing.T) {
	type stackCase struct {
		name string
		gsl  string
		// section is the reference: the material at a point in the shaft's span
		// of the non-round segment, and the rounds' radii either side of it.
		section       poly2
		below, above  float64 // neighbour radii; 0 for none
		from, to, end float64
	}
	cases := []stackCase{
		{"hex alone", "shaft.hex(20.0, 0.0, 30.0)", hexSection(20), 0, 0, 0, 30, 30},
		{"d-flat alone", "shaft.dFlat(16.0, 13.0, 0.0, 30.0)", dFlatSection(16, 13), 0, 0, 0, 30, 30},
		{"double-d alone", "shaft.doubleD(16.0, 12.0, 0.0, 30.0)", doubleDSection(16, 12), 0, 0, 0, 30, 30},
		{"spline alone", "shaft.spline(6.0, 16.0, 20.0, 4.0, 0.0, 30.0)", splineSection(6, 16, 20, 4), 0, 0, 0, 30, 30},
		{"hex between", "[shaft.diameter(18.0, -10.0, 0.0), shaft.hex(20.0, 0.0, 30.0), shaft.diameter(26.0, 30.0, 40.0)]",
			hexSection(20), 9, 13, 0, 30, 30},
		{"spline between", "[shaft.diameter(14.0, -10.0, 0.0), shaft.spline(6.0, 16.0, 20.0, 4.0, 0.0, 30.0), shaft.diameter(26.0, 30.0, 40.0)]",
			splineSection(6, 16, 20, 4), 7, 13, 0, 30, 30},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			doc := "@scene(140)\nsolid = shaft.body(" + c.gsl + ");\nreturn solid;\n"
			lo, hi := c.from, c.to
			if c.below > 0 {
				lo = c.from - 10
			}
			if c.above > 0 {
				hi = c.to + 10
			}
			points := shaftGrid3([3]float64{-14, -14, lo}, [3]float64{14, 14, hi}, 2, 28)
			got := evalStdShaft(t, doc, points)
			checked, bad := 0, 0
			for i, q := range points {
				var ref float64
				switch {
				case q[2] < c.from && c.below > 0:
					ref = shaftExtruded(math.Hypot(q[0], q[1])-c.below, q[2], c.from-10, c.from)
				case q[2] > c.to && c.above > 0:
					ref = shaftExtruded(math.Hypot(q[0], q[1])-c.above, q[2], c.to, c.to+10)
				default:
					ref = shaftExtruded(c.section.distance([2]float64{q[0], q[1]}), q[2], c.from, c.to)
				}
				if math.Abs(ref) < 1e-3 || math.Abs(math.Abs(q[2]-c.from)) < 1e-3 || math.Abs(q[2]-c.to) < 1e-3 {
					continue
				}
				checked++
				if (got[i] < 0) != (ref < 0) {
					if bad++; bad <= 3 {
						t.Errorf("at %v: field %v, reference %v", q, got[i], ref)
					}
				}
			}
			t.Logf("%d points checked, %d on the wrong side", checked, bad)
		})
	}
}

// The gear's own plane ends keep its surface too: across its span the sign is
// gears.gearProfile's, and past a plane end beside a larger round the round
// is whole.
func TestShaftEnds_GearPlaneEndsKeepTheSurface(t *testing.T) {
	doc := `@scene(140)
solid = shaft.body([shaft.diameter(12.0, -10.0, 0.0),
                    shaft.gear(17.0, 1.0, radians(20.0), 0.0, 0.0, 15.0),
                    shaft.diameter(24.0, 15.0, 25.0)]);
profile = gears.gearProfile(17.0, 1.0, radians(20.0), 0.0, 0.25, 0.05);
return solid, profile;
`
	points := shaftGrid3([3]float64{-12, -12, -10}, [3]float64{12, 12, 25}, 0, 40)
	solid := evalShaftNamed(t, doc, "solid", points)
	profile := evalShaftNamed(t, doc, "profile", points)
	checked, bad := 0, 0
	for i, q := range points {
		var ref float64
		switch {
		case q[2] < -1e-3:
			ref = math.Hypot(q[0], q[1]) - 6
		case q[2] > 15+1e-3:
			ref = math.Hypot(q[0], q[1]) - 12
		case q[2] > 1e-3 && q[2] < 15-1e-3:
			ref = profile[i]
		default:
			continue
		}
		if math.Abs(ref) < 1e-3 || q[2] < -10+1e-3 || q[2] > 25-1e-3 {
			continue
		}
		checked++
		if (solid[i] < 0) != (ref < 0) {
			if bad++; bad <= 3 {
				t.Errorf("at %v: field %v, reference %v", q, solid[i], ref)
			}
		}
	}
	t.Logf("%d points checked, %d on the wrong side", checked, bad)
}

// A chamfered tooth end: the blank's corner at the gear's top end is cut at
// 45 degrees by `c` before the teeth, so a tooth's tip stands at tip - (c - d)
// at distance d < c from the end plane, and at the tip further in. The gaps,
// below the chamfer, are unchanged.
func TestShaftEnds_ChamferedToothEnd(t *testing.T) {
	doc := `@scene(140)
solid = shaft.body([shaft.diameter(30.0, 0.0, 20.0),
                    shaft.gear(20.0, 2.0, radians(20.0), 0.0, 20.0, 50.0, shaft.planeEnd(), shaft.chamferEnd(1.0)),
                    shaft.diameter(30.0, 50.0, 80.0)]);
return solid;
`
	gap, tooth := gearGapAndTooth(t, 20, 2)
	zs := spanZs(45, 50, 0.125)
	checkRadii(t, "tooth", zs, radialSurface(t, doc, tooth, zs, 30), func(z float64) float64 {
		return 22 - math.Max(0, z-49)
	}, 1e-3)
	checkRadii(t, "gap", zs, radialSurface(t, doc, gap, zs, 30), func(z float64) float64 { return 17.5 }, 1e-3)
}

// The hobbed example compiles to WGSL (its size logged as Phase 4's budget)
// and evaluates: solid in the gear's and the spline's cores and in the
// shoulder past the hob's reach, empty beyond the end.
func TestShaftEnds_HobbedExampleCompilesToWGSL(t *testing.T) {
	src, err := os.ReadFile("../examples/shaft_hobbed.gsl")
	if err != nil {
		t.Fatal(err)
	}
	e := New(nil, nil)
	res, err := e.Shader(t.Context(), nil, string(src), "", ShaderOptions{Target: "wgsl", OutDir: t.TempDir()})
	if err != nil {
		t.Fatalf("shader: %v", err)
	}
	t.Logf("examples/shaft_hobbed.gsl: %d bytes of WGSL", len(res.Code))
	groups, err := e.EvalGsl(t.Context(), nil, string(src), "", []string{"hobbedShaft"},
		[][3]float64{{15, 0, 65}, {5, 0, 10}, {24, 0, 104}, {0, 0, 125}})
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	got := groups[0].Samples[0]
	if got[0].Value >= 0 || got[1].Value >= 0 || got[2].Value >= 0 || got[3].Value <= 0 {
		t.Errorf("gear %v, spline %v, shoulder %v, beyond the end %v: want solid three times, then empty",
			got[0].Value, got[1].Value, got[2].Value, got[3].Value)
	}
}

// No sheet inside the solid on either joint plane of the hobbed gear: every
// point inside the material on both sides reads solid.
func TestShaftEnds_NoSheetOnJoints(t *testing.T) {
	doc := `@scene(140)
solid = shaft.body([shaft.diameter(30.0, 0.0, 20.0),
                    shaft.gear(20.0, 2.0, radians(20.0), 0.0, 20.0, 50.0, shaft.planeEnd(), shaft.runout(80.0)),
                    shaft.diameter(50.0, 50.0, 80.0)]);
profile = gears.gearProfile(20.0, 2.0, radians(20.0), 0.0, 0.25, 0.1);
return solid, profile;
`
	var points [][3]float64
	for _, z := range []float64{20, 50} {
		for x := -24.0; x <= 24; x += 0.4 {
			for y := -24.0; y <= 24; y += 0.4 {
				points = append(points, [3]float64{x, y, z})
			}
		}
	}
	solid := evalShaftNamed(t, doc, "solid", points)
	profile := evalShaftNamed(t, doc, "profile", points)
	checked, bad := 0, 0
	for i, q := range points {
		r := math.Hypot(q[0], q[1])
		var inside bool
		if q[2] == 20 {
			inside = r < 15-0.05 // the journal; the gear's root disc is larger
		} else {
			inside = profile[i] < -0.05 // the gear; the shoulder is larger
		}
		if !inside {
			continue
		}
		checked++
		if solid[i] > -1e-3 {
			if bad++; bad <= 3 {
				t.Errorf("joint point %v inside both sides reads %v", q, solid[i])
			}
		}
	}
	t.Logf("%d joint points inside both sides, %d not solid", checked, bad)
	if checked == 0 {
		t.Fatal("no joint point inside both sides")
	}
}
