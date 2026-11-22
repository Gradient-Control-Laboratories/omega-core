//go:build cgo

package eval

import (
	"fmt"
	"math"
	"os"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// Shaft Modeling, Phase 6: joint blends. A joint names its blend by an entry
// in the segment list between the two segments it blends, shaft.fillet(r) or
// shaft.chamfer(c). The blend is the union blend of the two capped segments,
// min'd into the fold. At a square shoulder between round segments the two
// fields are exact and orthogonal, so the fillet is the torus's arc exactly
// and the chamfer its 45-degree line: the field outside is the distance, with
// a unit gradient. Inside, near the blend, it may read deeper than the
// distance (an edge treatment's exception), but never shallow, and the rest
// of the shaft is unchanged.
//
// The shoulder under test: a diameter of 20 from 0 to 20, then 40 from 20 to
// 40. Its concave corner, in the half-section (distance from the axis, z), is
// at (10, 20).

const (
	blendSmallR = 10.0
	blendLargeR = 20.0
	blendJointZ = 20.0
	blendSize   = 2.0
)

func shoulderDoc(entry string) string {
	if entry != "" {
		entry += ", "
	}
	return fmt.Sprintf(`@scene(100)
solid = shaft.body([shaft.diameter(20.0, 0.0, 20.0), %sshaft.diameter(40.0, 20.0, 40.0)]);
return solid;
`, entry)
}

// shoulderMeridian is the shoulder's section through its axis, mirrored across
// it, with the corner replaced by `corner`: the points from the small
// diameter's side up to the large one's face, in order.
func shoulderMeridian(corner [][2]float64) poly2 {
	right := [][2]float64{{blendSmallR, 0}}
	right = append(right, corner...)
	right = append(right, [2]float64{blendLargeR, blendJointZ}, [2]float64{blendLargeR, 40})
	var pg poly2
	pg = append(pg, right...)
	for i := len(right) - 1; i >= 0; i-- {
		pg = append(pg, [2]float64{-right[i][0], right[i][1]})
	}
	return pg
}

// shoulderFillet is the fillet's arc, centred at (10 + r, 20 - r), from the
// small diameter's side round to the large one's face, in chords short enough
// that the polygon stands within 1e-7 of the arc.
func shoulderFillet(r float64) [][2]float64 {
	cx, cz := blendSmallR+r, blendJointZ-r
	const n = 4000
	out := make([][2]float64, 0, n+1)
	for i := 0; i <= n; i++ {
		phi := math.Pi - (math.Pi/2)*float64(i)/n
		out = append(out, [2]float64{cx + r*math.Cos(phi), cz + r*math.Sin(phi)})
	}
	return out
}

// shoulderChamfer is the 45-degree line from (10, 20 - c) to (10 + c, 20).
func shoulderChamfer(c float64) [][2]float64 {
	return [][2]float64{{blendSmallR, blendJointZ - c}, {blendSmallR + c, blendJointZ}}
}

func meridianReference(pg poly2) func([3]float64) float64 {
	return func(q [3]float64) float64 { return pg.distance([2]float64{math.Hypot(q[0], q[1]), q[2]}) }
}

// aroundCorner is a grid in the half-section over [rho0, rho1] x [z0, z1], at a
// few angles about the axis.
func aroundCorner(rho0, rho1, z0, z1, step float64) [][3]float64 {
	var pts [][3]float64
	for _, a := range []float64{0.3, 2.0, 3.7, 5.4} {
		for rho := rho0; rho <= rho1; rho += step {
			for z := z0; z <= z1; z += step {
				pts = append(pts, [3]float64{rho * math.Cos(a), rho * math.Sin(a), z})
			}
		}
	}
	return pts
}

// distanceToRing is how far q lies from the shoulder's corner ring.
func distanceToRing(q [3]float64) float64 {
	return math.Hypot(math.Hypot(q[0], q[1])-blendSmallR, q[2]-blendJointZ)
}

func evalBlendGrad(t *testing.T, doc string, points [][3]float64) []GradSample {
	t.Helper()
	e := New(&source.MockManager{}, nil)
	groups, err := e.EvalGsl(t.Context(), nil, doc, "", []string{"solid"}, points)
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	return groups[0].Samples[0]
}

// checkBlendIsExact: outside, the field is the distance to the blended
// meridian; everywhere clear of the surface, its sign is the meridian's.
func checkBlendIsExact(t *testing.T, label, entry string, corner [][2]float64) {
	t.Helper()
	ref := meridianReference(shoulderMeridian(corner))
	pts := aroundCorner(7, 16, 14, 24, 0.0625)
	got := evalStdShaft(t, shoulderDoc(entry), pts)
	outside, miss, wrongSign, worst := 0, 0, 0, 0.0
	for i, q := range pts {
		want := ref(q)
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

// A fillet at a square shoulder between round segments is the torus's arc:
// outside, the field is the distance to it.
func TestShaftBlends_FilletIsTheArc(t *testing.T) {
	checkBlendIsExact(t, "fillet", fmt.Sprintf("shaft.fillet(%v)", blendSize), shoulderFillet(blendSize))
}

// A chamfer at the same shoulder is its 45-degree line.
func TestShaftBlends_ChamferIsTheLine(t *testing.T) {
	checkBlendIsExact(t, "chamfer", fmt.Sprintf("shaft.chamfer(%v)", blendSize), shoulderChamfer(blendSize))
}

// Outside, within 1.5 mm of the blend, the gradient has unit length.
func TestShaftBlends_UnitGradientNearTheBlend(t *testing.T) {
	for _, c := range []struct {
		label, entry string
		corner       [][2]float64
	}{
		{"fillet", fmt.Sprintf("shaft.fillet(%v)", blendSize), shoulderFillet(blendSize)},
		{"chamfer", fmt.Sprintf("shaft.chamfer(%v)", blendSize), shoulderChamfer(blendSize)},
	} {
		ref := meridianReference(shoulderMeridian(c.corner))
		var pts [][3]float64
		for _, q := range aroundCorner(9.5, 13, 16.5, 20.5, 0.05) {
			if d := ref(q); d > 0.01 && d < 1.5 {
				pts = append(pts, q)
			}
		}
		samples := evalBlendGrad(t, shoulderDoc(c.entry), pts)
		bad, worst := 0, 0.0
		for i, s := range samples {
			g := math.Sqrt(float64(s.Dx)*float64(s.Dx) + float64(s.Dy)*float64(s.Dy) + float64(s.Dz)*float64(s.Dz))
			worst = math.Max(worst, math.Abs(g-1))
			if math.Abs(g-1) > 1e-3 {
				if bad++; bad <= 3 {
					t.Errorf("%s at %v: |grad f| = %v", c.label, pts[i], g)
				}
			}
		}
		t.Logf("%s: %d points outside within 1.5 mm, worst | |grad f| - 1 | %.2g", c.label, len(pts), worst)
	}
}

// No sheet inside the solid where the blend meets either segment or where
// the segments meet: on the joint plane and on the small diameter's side
// under the blend, every point at least 1e-3 deep reads solid, and, farther
// than the blend's size from the corner, no shallower than the body promises.
// Nearer, the blend's own material deepens the solid where the fold cannot
// see it: an edge treatment's exception (the owner's call, 2026-11-20).
func TestShaftBlends_NoSheetOnTheJoint(t *testing.T) {
	for _, c := range []struct {
		label, entry string
		corner       [][2]float64
	}{
		{"fillet", fmt.Sprintf("shaft.fillet(%v)", blendSize), shoulderFillet(blendSize)},
		{"chamfer", fmt.Sprintf("shaft.chamfer(%v)", blendSize), shoulderChamfer(blendSize)},
	} {
		ref := meridianReference(shoulderMeridian(c.corner))
		var pts [][3]float64
		for _, a := range []float64{0.3, 2.0, 3.7, 5.4} {
			for rho := 0.0; rho < blendSmallR+blendSize; rho += 0.05 {
				pts = append(pts, [3]float64{rho * math.Cos(a), rho * math.Sin(a), blendJointZ})
			}
			for z := blendJointZ - blendSize; z <= blendJointZ; z += 0.05 {
				pts = append(pts, [3]float64{blendSmallR * math.Cos(a), blendSmallR * math.Sin(a), z})
			}
		}
		got := evalStdShaft(t, shoulderDoc(c.entry), pts)
		checked, bad := 0, 0
		for i, q := range pts {
			want := ref(q)
			if want >= -1e-3 {
				continue
			}
			checked++
			if got[i] > -1e-3 || (distanceToRing(q) > blendSize && tooShallow(got[i], want)) {
				if bad++; bad <= 3 {
					t.Errorf("%s at %v: field %v where the solid is %v deep", c.label, q, got[i], -want)
				}
			}
		}
		t.Logf("%s: %d points inside on the joint and under the blend, %d shallow", c.label, checked, bad)
	}
}

// Farther than twice the blend's size from the corner ring, and wherever the
// blend leaves the true distance as it was, the blended shaft reads exactly
// as the plain one, inside and out. The second condition is the chamfer's: its
// face is the nearest surface in a strip along the corner's outward bisector,
// about 3.4 times its size long, so there the true distance does change (the
// owner's call, 2026-11-20).
func TestShaftBlends_RestUnchanged(t *testing.T) {
	plainDoc := shoulderDoc("")
	plainRef := meridianReference(shoulderMeridian([][2]float64{{blendSmallR, blendJointZ}}))
	pts := aroundCorner(0, 26, -4, 44, 0.5)
	var far [][3]float64
	for _, q := range pts {
		if distanceToRing(q) > 2*blendSize {
			far = append(far, q)
		}
	}
	plain := evalStdShaft(t, plainDoc, far)
	for _, c := range []struct {
		entry  string
		corner [][2]float64
	}{
		{fmt.Sprintf("shaft.fillet(%v)", blendSize), shoulderFillet(blendSize)},
		{fmt.Sprintf("shaft.chamfer(%v)", blendSize), shoulderChamfer(blendSize)},
	} {
		entry := c.entry
		blendRef := meridianReference(shoulderMeridian(c.corner))
		got := evalStdShaft(t, shoulderDoc(entry), far)
		bad, worst := 0, 0.0
		for i, q := range far {
			if math.Abs(blendRef(q)-plainRef(q)) > 1e-9 {
				continue
			}
			d := math.Abs(got[i] - plain[i])
			worst = math.Max(worst, d)
			if d > 1e-6 {
				if bad++; bad <= 3 {
					t.Errorf("%s at %v: blended %v, plain %v", entry, q, got[i], plain[i])
				}
			}
		}
		t.Logf("%s: %d points clear of the corner, worst change %.2g mm", entry, len(far), worst)
	}
}

// What the blend entries cost: a six-segment turned shaft compiled to WGSL
// with no blends and with a fillet at each of its five joints. Logged, as
// Phase 0's spike (a) logged the fold's growth; it fails only if the shaft
// does not compile.
func TestShaftBlends_WGSLCost(t *testing.T) {
	segs := []string{
		"shaft.diameter(20.0, 0.0, 15.0)", "shaft.diameter(25.0, 15.0, 30.0)", "shaft.diameter(32.0, 30.0, 45.0)",
		"shaft.diameter(40.0, 45.0, 55.0)", "shaft.diameter(30.0, 55.0, 75.0)", "shaft.diameter(22.0, 75.0, 90.0)",
	}
	build := func(blend bool) string {
		list := segs[0]
		for _, s := range segs[1:] {
			if blend {
				list += ", shaft.fillet(1.0)"
			}
			list += ", " + s
		}
		return "@scene(120)\nsolid = shaft.body([" + list + "]);\nreturn solid;\n"
	}
	e := New(nil, nil)
	sizes := map[bool]int{}
	for _, blend := range []bool{false, true} {
		res, err := e.Shader(t.Context(), nil, build(blend), "", ShaderOptions{Target: "wgsl", OutDir: t.TempDir()})
		if err != nil {
			t.Fatalf("blend %v: shader: %v", blend, err)
		}
		sizes[blend] = len(res.Code)
	}
	t.Logf("six segments: %d bytes of WGSL plain, %d with five fillets (+%d)", sizes[false], sizes[true], sizes[true]-sizes[false])
}

// The control: with no blend, the same check holds the plain shoulder to its
// sharp corner, so the reference and the check stand on their own.
func TestShaftBlends_PlainShoulderControl(t *testing.T) {
	checkBlendIsExact(t, "plain", "", [][2]float64{{blendSmallR, blendJointZ}})
}

// The blended example compiles to WGSL (its size logged as Phase 6's budget)
// and evaluates: solid in the journal, in a fillet's corner and in the
// collar, empty beyond the end.
func TestShaftBlends_BlendedExampleCompilesToWGSL(t *testing.T) {
	src, err := os.ReadFile("../examples/shaft_blended.gsl")
	if err != nil {
		t.Fatal(err)
	}
	e := New(nil, nil)
	res, err := e.Shader(t.Context(), nil, string(src), "", ShaderOptions{Target: "wgsl", OutDir: t.TempDir()})
	if err != nil {
		t.Fatalf("shader: %v", err)
	}
	t.Logf("examples/shaft_blended.gsl: %d bytes of WGSL", len(res.Code))
	// The first fillet's corner: (10, 25) in the half-section, 0.2 mm out
	// along its bisector, inside the 1 mm fillet but outside both segments.
	corner := 10 + 0.2/math.Sqrt2
	groups, err := e.EvalGsl(t.Context(), nil, string(src), "", []string{"blendedShaft"},
		[][3]float64{{0, 0, 10}, {corner, 0, 25 - 0.2/math.Sqrt2}, {16, 0, 75}, {0, 0, 115}})
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	got := groups[0].Samples[0]
	if got[0].Value >= 0 || got[1].Value >= 0 || got[2].Value >= 0 || got[3].Value <= 0 {
		t.Errorf("journal %v, fillet corner %v, collar %v, beyond the end %v: want solid three times, then empty",
			got[0].Value, got[1].Value, got[2].Value, got[3].Value)
	}
}
