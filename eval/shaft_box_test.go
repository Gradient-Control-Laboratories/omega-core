//go:build cgo

package eval

import (
	"context"
	"fmt"
	"math"
	"testing"
)

// Shaft Modeling, Phase 10: the box and the simplified model. shaft.body
// declares a box from its end faces and each segment's reach from the axis; a
// caller's own .box replaces it. shaft.envelope is the same list with its
// detail dropped: a gear at its tip circle, a spline and a thread at their
// major diameters, runouts gone, the smooth segments, the joints' blends and
// the end holes kept, under the same box. The analysis grows every box by
// 0.01, so a box is checked to within 0.02 of what the segments give.

// boxCase is a shaft, its list, and the box its segments give: the largest
// reach from the axis, and the end faces. A reach of 0 means unbounded
// radially, which only containment is checked against.
type boxCase struct {
	name, segments string
	reach, z0, z1  float64
}

func boxCases() []boxCase {
	return []boxCase{
		{"turned", "[shaft.diameter(20.0, 0.0, 10.0), shaft.taper(20.0, 30.0, 10.0, 15.0), shaft.diameter(30.0, 15.0, 40.0), shaft.diameter(24.0, 40.0, 50.0)]", 15, 0, 50},
		{"pinion", "[shaft.spline(6.0, 16.0, 20.0, 4.0, 0.0, 20.0), shaft.diameter(26.0, 20.0, 30.0), shaft.gear(18.0, 2.0, radians(20.0), radians(15.0), 30.0, 60.0), shaft.hex(24.0, 60.0, 70.0), shaft.dFlat(20.0, 18.0, 70.0, 80.0), shaft.doubleD(20.0, 16.0, 80.0, 90.0)]", 20, 0, 90},
		{"threaded", "[shaft.thread(6.0, 1.0, 0.0, 20.0, shaft.chamferEnd(0.6), shaft.threadRunout(2.5)), shaft.diameter(6.0, 20.0, 25.0), shaft.diameter(10.0, 25.0, 30.0)]", 5, 0, 30},
		{"hobbed", "[shaft.diameter(30.0, 0.0, 20.0), shaft.gear(20.0, 2.0, radians(20.0), 0.0, 20.0, 50.0, shaft.chamferEnd(1.0), shaft.runout(80.0)), shaft.diameter(50.0, 50.0, 80.0)]", 25, 0, 80},
		{"crank", "[shaft.diameter(30.0, 0.0, 30.0), shaft.web(25.0, 18.0, 40.0, 0.7, 30.0, 45.0), shaft.eccentric(28.0, 40.0, 0.7, 45.0, 75.0), shaft.web(25.0, 18.0, 40.0, 0.7, 75.0, 90.0), shaft.diameter(30.0, 90.0, 120.0)]", 58, 0, 120},
		{"blended", "[shaft.diameter(20.0, 0.0, 20.0), shaft.fillet(2.0), shaft.diameter(40.0, 20.0, 40.0)]", 20, 0, 40},
		{"section", "[shaft.diameter(16.0, 0.0, 10.0), shaft.section(profileXY([[-5.0, -5.0], [5.0, -5.0], [5.0, 5.0], [-5.0, 5.0]]), 10.0, 20.0)]", 0, 0, 20},
	}
}

// readBox reads a named value's box, [lo, hi].
func readBox(t *testing.T, doc, name string) (lo, hi [3]float64) {
	t.Helper()
	e := New(nil, nil)
	groups, err := e.EvalGsl(context.Background(), nil, doc, "", []string{name + ".box[0]", name + ".box[1]"}, [][3]float64{{0, 0, 0}})
	if err != nil {
		t.Fatalf("eval %s.box: %v", name, err)
	}
	for k := 0; k < 3; k++ {
		lo[k] = float64(groups[0].Samples[k][0].Value)
		hi[k] = float64(groups[1].Samples[k][0].Value)
	}
	return lo, hi
}

func boxDoc(entry, segments string) string {
	return fmt.Sprintf("@scene(200)\nsolid = shaft.%s(%s);\nreturn solid;\n", entry, segments)
}

// checkBox: every point of the solid lies in the box, and where the segments
// give a reach the box is that reach and those faces.
func checkBox(t *testing.T, entry string, c boxCase) {
	t.Helper()
	doc := boxDoc(entry, c.segments)
	lo, hi := readBox(t, doc, "solid")
	if c.reach > 0 {
		want := [2][3]float64{{-c.reach, -c.reach, c.z0}, {c.reach, c.reach, c.z1}}
		for k := 0; k < 3; k++ {
			if math.Abs(lo[k]-want[0][k]) > 0.02 || math.Abs(hi[k]-want[1][k]) > 0.02 {
				t.Errorf("%s %s: box %v .. %v, want %v .. %v", entry, c.name, lo, hi, want[0], want[1])
				break
			}
		}
	}
	span := math.Max(c.reach, 25) + 5
	var pts [][3]float64
	for x := -span; x <= span; x += span / 23 {
		for y := -span; y <= span; y += span / 21 {
			for z := c.z0 - 3; z <= c.z1+3; z += (c.z1 - c.z0 + 6) / 37 {
				pts = append(pts, [3]float64{x, y, z})
			}
		}
	}
	f := evalShaftNamed(t, doc, "solid", pts)
	inside, out := 0, 0
	for i, q := range pts {
		if f[i] >= 0 {
			continue
		}
		inside++
		for k := 0; k < 3; k++ {
			if q[k] < lo[k] || q[k] > hi[k] {
				if out++; out <= 3 {
					t.Errorf("%s %s: solid point %v outside its box %v .. %v", entry, c.name, q, lo, hi)
				}
				break
			}
		}
	}
	t.Logf("%s %s: box %.2f .. %.2f; %d solid points, %d outside", entry, c.name, lo, hi, inside, out)
}

// The body's box holds the shaft, and is the segments' reach and end faces.
func TestShaftBox_BodyBoxHoldsTheShaft(t *testing.T) {
	for _, c := range boxCases() {
		checkBox(t, "body", c)
	}
}

// The envelope declares the same box, and holds itself.
func TestShaftBox_EnvelopeBoxHoldsIt(t *testing.T) {
	for _, c := range boxCases() {
		checkBox(t, "envelope", c)
	}
}

// With end holes, the box is the same: holes take material away.
func TestShaftBox_HolesKeepTheBox(t *testing.T) {
	doc := `@scene(200)
solid = shaft.body([shaft.diameter(20.0, 0.0, 20.0), shaft.diameter(30.0, 20.0, 40.0)],
                   drill.centerHole(3.15, 6.7, 8.0), drill.centerHole(3.15, 6.7, 8.0));
return solid;
`
	lo, hi := readBox(t, doc, "solid")
	want := [2][3]float64{{-15, -15, 0}, {15, 15, 40}}
	for k := 0; k < 3; k++ {
		if math.Abs(lo[k]-want[0][k]) > 0.02 || math.Abs(hi[k]-want[1][k]) > 0.02 {
			t.Errorf("box %v .. %v, want %v .. %v", lo, hi, want[0], want[1])
			break
		}
	}
}

// A caller's own box replaces the shaft's.
func TestShaftBox_OverrideReplaces(t *testing.T) {
	doc := `@scene(200)
solid = shaft.body([shaft.diameter(20.0, 0.0, 20.0), shaft.diameter(30.0, 20.0, 40.0)]);
solid.box = [[-1.0, -2.0, -3.0], [1.0, 2.0, 3.0]];
return solid;
`
	lo, hi := readBox(t, doc, "solid")
	want := [2][3]float64{{-1, -2, -3}, {1, 2, 3}}
	for k := 0; k < 3; k++ {
		if math.Abs(lo[k]-want[0][k]) > 0.02 || math.Abs(hi[k]-want[1][k]) > 0.02 {
			t.Errorf("overridden box %v .. %v, want %v .. %v", lo, hi, want[0], want[1])
			break
		}
	}
}

// The envelope holds the detailed shaft, by sign on a dense grid: wherever the
// detail is solid the envelope is, and wherever the envelope is empty the
// detail is (sampled: the two share surfaces, which intervals cannot
// separate). By sign, not by value: a helical gear's field is a bound outside,
// reading below the envelope's exact cylinder (the owner's call, 2026-11-20).
func TestShaftBox_EnvelopeHoldsTheDetail(t *testing.T) {
	for _, c := range boxCases() {
		if c.name == "section" {
			continue
		}
		doc := fmt.Sprintf("@scene(200)\ndetail = shaft.body(%s);\nenv = shaft.envelope(%s);\nreturn detail;\n", c.segments, c.segments)
		span := math.Max(c.reach, 25) + 3
		var pts [][3]float64
		for x := -span; x <= span; x += span / 31 {
			for y := -span; y <= span; y += span / 29 {
				for z := c.z0 - 2; z <= c.z1+2; z += (c.z1 - c.z0 + 4) / 53 {
					pts = append(pts, [3]float64{x, y, z})
				}
			}
		}
		detail := evalShaftNamed(t, doc, "detail", pts)
		env := evalShaftNamed(t, doc, "env", pts)
		bad := 0
		for i, q := range pts {
			if (detail[i] < -1e-4 && env[i] >= 0) || (env[i] > 1e-4 && detail[i] <= 0) {
				if bad++; bad <= 3 {
					t.Errorf("%s at %v: envelope %v, detail %v: the detail pokes out", c.name, q, env[i], detail[i])
				}
			}
		}
		t.Logf("%s: %d points, %d where the detail pokes out of the envelope", c.name, len(pts), bad)
	}
}

// The envelope has no teeth, threads or splines: a gear, a thread and a
// spline in a row are, outside, exactly their tip and major cylinders (the
// min of the capped cylinders' distances), and on the right side everywhere.
func TestShaftBox_EnvelopeIsTheSmoothCylinders(t *testing.T) {
	segs := "[shaft.gear(20.0, 2.0, radians(20.0), 0.0, 0.0, 30.0), shaft.thread(6.0, 1.0, 30.0, 50.0), shaft.spline(6.0, 16.0, 20.0, 4.0, 50.0, 70.0)]"
	doc := fmt.Sprintf("@scene(200)\nenv = shaft.envelope(%s);\nreturn env;\n", segs)
	cyl := func(q [3]float64, r, z0, z1 float64) float64 {
		return prismDistance(math.Hypot(q[0], q[1])-r, q[2], z0, z1)
	}
	ref := func(q [3]float64) float64 {
		return math.Min(cyl(q, 22, 0, 30), math.Min(cyl(q, 3, 30, 50), cyl(q, 10, 50, 70)))
	}
	var pts [][3]float64
	for x := -26.0; x <= 26; x += 0.83 {
		for y := -26.0; y <= 26; y += 1.17 {
			for z := -3.0; z <= 73; z += 1.31 {
				pts = append(pts, [3]float64{x, y, z})
			}
		}
	}
	got := evalShaftNamed(t, doc, "env", pts)
	outside, miss, wrong, worst := 0, 0, 0, 0.0
	for i, q := range pts {
		want := ref(q)
		if math.Abs(want) > 1e-3 && (got[i] < 0) != (want < 0) {
			if wrong++; wrong <= 3 {
				t.Errorf("at %v: envelope %v, cylinders %v: wrong side", q, got[i], want)
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
				t.Errorf("at %v: envelope %v, cylinders %v", q, got[i], want)
			}
		}
	}
	t.Logf("envelope: %d points outside, worst miss %.2g mm, %d on the wrong side", outside, worst, wrong)
}

// The envelope keeps the joints' blends and the end holes: for a shaft of
// smooth segments, filleted and centre-drilled, it is the body exactly.
func TestShaftBox_EnvelopeKeepsBlendsAndHoles(t *testing.T) {
	segs := "[shaft.diameter(20.0, 0.0, 20.0), shaft.fillet(2.0), shaft.diameter(40.0, 20.0, 40.0)]"
	doc := fmt.Sprintf(`@scene(200)
body = shaft.body(%s, drill.centerHole(3.15, 6.7, 8.0), drill.centerHole(3.15, 6.7, 8.0));
env = shaft.envelope(%s, drill.centerHole(3.15, 6.7, 8.0), drill.centerHole(3.15, 6.7, 8.0));
return body;
`, segs, segs)
	var pts [][3]float64
	for x := -23.0; x <= 23; x += 0.61 {
		for z := -3.0; z <= 43; z += 0.29 {
			pts = append(pts, [3]float64{x, 0.4 * x, z})
		}
	}
	body := evalShaftNamed(t, doc, "body", pts)
	env := evalShaftNamed(t, doc, "env", pts)
	bad, worst := 0, 0.0
	for i, q := range pts {
		d := math.Abs(body[i] - env[i])
		worst = math.Max(worst, d)
		if d > 1e-6 {
			if bad++; bad <= 3 {
				t.Errorf("at %v: body %v, envelope %v", q, body[i], env[i])
			}
		}
	}
	t.Logf("smooth shaft: %d points, worst difference %.2g mm", len(pts), worst)
}

// The point of the simplified model: a claim about the pinion shaft, that it
// stays within 0.3 mm of its gear's tip circle, proves on the envelope, and
// with fewer interval evaluations than on the detailed shaft.
func TestShaftBox_EnvelopeProvesFaster(t *testing.T) {
	segs := boxCases()[1].segments
	doc := fmt.Sprintf(`@scene(200)
detail = shaft.body(%s);
env = shaft.envelope(%s);
detailClear = max(detail, 20.3 - length(p.xy));
envClear = max(env, 20.3 - length(p.xy));
return detail;
`, segs, segs)
	region := [6]float64{-22, -22, -2, 22, 22, 92}
	e := New(nil, nil)
	envR, err := e.Certify(context.Background(), nil, doc, "", "envClear", region, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	detailR, err := e.Certify(context.Background(), nil, doc, "", "detailClear", region, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !envR.Proven {
		t.Errorf("the claim should prove on the envelope, got %+v", envR)
	}
	if envR.Evals >= detailR.Evals {
		t.Errorf("the envelope took %d evaluations, the detailed shaft %d", envR.Evals, detailR.Evals)
	}
	t.Logf("envelope: proven %v in %d evaluations; detail: proven %v in %d (%s)", envR.Proven, envR.Evals, detailR.Proven, detailR.Evals, detailR.Reason)
}
