//go:build cgo

package eval

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// intersectChamferNormalized: a bevel `depth` in from the edge, the sum of the
// two fields divided by the length of the sum of the two faces' normals rather
// than by sqrt2, offset by the depth. intersectChamfer with k = depth·|n1 + n2|
// has the same bevel plane, so the old operators below are given that k. At a
// right angle the two agree; at any other angle the bevel plane is the same
// plane (only its scale differs, so the surface is the same), but its field
// has unit gradient where intersectChamfer's has sqrt(1 + n1.n2), overstating
// the distance at an obtuse edge.

func evalChamferField(t *testing.T, doc, name string, points [][3]float64) []float64 {
	t.Helper()
	e := New(&source.MockManager{}, nil)
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

// wedgeDoc is two half-planes through the z axis meeting at an edge with the
// given interior angle (degrees), the material between them, chamfered four
// ways: the old operator, the normalized one with explicit normals, the
// normalized one reading its fields' gradients, and the union and difference
// counterparts of each.
func wedgeDoc(interior float64) string {
	half := (180 - interior) / 2 * math.Pi / 180
	// Each face's outward normal, tilted half the exterior angle off +x.
	n1 := [2]float64{math.Cos(half), math.Sin(half)}
	n2 := [2]float64{math.Cos(half), -math.Sin(half)}
	vec := func(v [2]float64) string { return fmt.Sprintf("[%.12f, %.12f, 0.0]", v[0], v[1]) }
	return fmt.Sprintf(`@scene(40)
n1 = %s;
n2 = %s;
d1 = dot(p, n1);
d2 = dot(p, n2);
old = intersectChamfer(d1, d2, 1.5 * length(n1 + n2));
normals = intersectChamferNormalized(d1, d2, 1.5, n1, n2);
gradients = intersectChamferNormalized(d1, d2, 1.5);
oldUnion = unionChamfer(d1, d2, 1.5 * length(n1 + n2));
normUnion = unionChamferNormalized(d1, d2, 1.5, n1, n2);
normUnionGradients = unionChamferNormalized(d1, d2, 1.5);
oldDifference = differenceChamfer(d1, neg(d2), 1.5 * length(n1 + n2));
normDifference = differenceChamferNormalized(d1, neg(d2), 1.5, n1, neg(n2));
normDifferenceGradients = differenceChamferNormalized(d1, neg(d2), 1.5);
return old;
`, vec(n1), vec(n2))
}

// nearCrease reports whether q is within two finite-difference steps of a crease of wedgeDoc's
// chamfer at depth 1.5, where two of its three planes (the faces d1 and d2 and the bevel) meet:
// a difference taken across a crease mixes the planes' gradients and can read above 1 though
// neither plane does. A crease is where the two largest planes are equal, and each plane has
// unit gradient, so their difference changes at most twice as fast as distance. sign is 1 for
// the intersection and the difference, which is the intersection with d2's side flipped, and
// -1 for the union, the intersection of the complements.
func nearCrease(interior float64, q [3]float64, step, sign float64) bool {
	half := (180 - interior) / 2 * math.Pi / 180
	n1 := [2]float64{math.Cos(half), math.Sin(half)}
	n2 := [2]float64{math.Cos(half), -math.Sin(half)}
	d1 := sign * (q[0]*n1[0] + q[1]*n1[1])
	d2 := sign * (q[0]*n2[0] + q[1]*n2[1])
	bevel := (d1+d2)/math.Hypot(n1[0]+n2[0], n1[1]+n2[1]) + 1.5
	planes := []float64{d1, d2, bevel}
	sort.Float64s(planes)
	return planes[2]-planes[1] < 2*2*step
}

func wedgePoints() [][3]float64 {
	var pts [][3]float64
	for x := -6.0; x <= 6.0; x += 0.37 {
		for y := -6.0; y <= 6.0; y += 0.37 {
			pts = append(pts, [3]float64{x + 0.013, y + 0.029, 0.5})
		}
	}
	return pts
}

// At a right angle, the normalized chamfer is the old one, both forms.
func TestChamferNormalized_RightAngleIsIntersectChamfer(t *testing.T) {
	doc := wedgeDoc(90)
	pts := wedgePoints()
	old := evalChamferField(t, doc, "old", pts)
	for _, name := range []string{"normals", "gradients"} {
		got := evalChamferField(t, doc, name, pts)
		for i, q := range pts {
			if math.Abs(got[i]-old[i]) > 1e-5 {
				t.Fatalf("%s at %v: %v, intersectChamfer %v", name, q, got[i], old[i])
			}
		}
	}
}

// At obtuse and acute edges, the same surface (the sign agrees everywhere,
// away from the surface itself) and, where the bevel decides the field, a
// gradient of 1. The old bevel's gradient there is sqrt(1 + n1.n2).
func TestChamferNormalized_SameSurfaceUnitGradient(t *testing.T) {
	for _, interior := range []float64{120, 150, 60} {
		t.Run(fmt.Sprintf("%v degrees", interior), func(t *testing.T) {
			doc := wedgeDoc(interior)
			pts := wedgePoints()
			const step = 0.002
			var probe [][3]float64
			for _, q := range pts {
				probe = append(probe, q, [3]float64{q[0] + step, q[1], q[2]}, [3]float64{q[0], q[1] + step, q[2]})
			}
			oldVals := evalChamferField(t, doc, "old", probe)
			for _, name := range []string{"normals", "gradients"} {
				got := evalChamferField(t, doc, name, probe)
				worstGrad, worstOld, bevelPoints := 0.0, 0.0, 0
				for i := 0; i < len(probe); i += 3 {
					if math.Abs(oldVals[i]) > 1e-3 && (got[i] > 0) != (oldVals[i] > 0) {
						t.Fatalf("%s at %v: %v against intersectChamfer's %v: a different surface", name, probe[i], got[i], oldVals[i])
					}
					if nearCrease(interior, probe[i], step, 1) {
						continue
					}
					g := math.Hypot((got[i+1]-got[i])/step, (got[i+2]-got[i])/step)
					gOld := math.Hypot((oldVals[i+1]-oldVals[i])/step, (oldVals[i+2]-oldVals[i])/step)
					worstGrad = math.Max(worstGrad, g)
					worstOld = math.Max(worstOld, gOld)
					if math.Abs(gOld-1) > 0.01 {
						bevelPoints++
					}
				}
				t.Logf("%s: steepest gradient %.4f (intersectChamfer %.4f, %d points on its bevel)", name, worstGrad, worstOld, bevelPoints)
				if worstGrad > 1.002 {
					t.Errorf("%s: gradient %.4f, want at most 1", name, worstGrad)
				}
			}
		})
	}
}

// The union and difference counterparts: the same surfaces as unionChamfer
// and differenceChamfer, and unit gradient.
func TestChamferNormalized_UnionAndDifference(t *testing.T) {
	doc := wedgeDoc(120)
	pts := wedgePoints()
	const step = 0.002
	var probe [][3]float64
	for _, q := range pts {
		probe = append(probe, q, [3]float64{q[0] + step, q[1], q[2]}, [3]float64{q[0], q[1] + step, q[2]})
	}
	pairs := map[string]string{
		"normUnion": "oldUnion", "normUnionGradients": "oldUnion",
		"normDifference": "oldDifference", "normDifferenceGradients": "oldDifference",
	}
	for name, oldName := range pairs {
		got := evalChamferField(t, doc, name, probe)
		old := evalChamferField(t, doc, oldName, probe)
		worst := 0.0
		for i := 0; i < len(probe); i += 3 {
			if math.Abs(old[i]) > 1e-3 && (got[i] > 0) != (old[i] > 0) {
				t.Fatalf("%s at %v: %v against %s's %v: a different surface", name, probe[i], got[i], oldName, old[i])
			}
			sign := 1.0
			if strings.HasPrefix(name, "normUnion") {
				sign = -1
			}
			if nearCrease(120, probe[i], step, sign) {
				continue
			}
			worst = math.Max(worst, math.Hypot((got[i+1]-got[i])/step, (got[i+2]-got[i])/step))
		}
		t.Logf("%s: steepest gradient %.4f", name, worst)
		if worst > 1.002 {
			t.Errorf("%s: gradient %.4f, want at most 1", name, worst)
		}
	}
}

// Interval bounds over cells straddling an obtuse wedge's surface, bevel
// included, are sound (they hold the value at the cell's centre) and tight
// (no wider than a few cell diagonals), for both forms. The gradient form
// bounds d1.dp and d2.dp as well as the fields.
func TestChamferNormalized_IntervalsAreSound(t *testing.T) {
	doc := wedgeDoc(120)
	const half = 0.25
	diagonal := 2 * half * math.Sqrt(3)
	var centres [][3]float64
	for _, q := range wedgePoints() {
		centres = append(centres, q)
	}
	for _, name := range []string{"normals", "gradients"} {
		values := evalChamferField(t, doc, name, centres)
		var boxes [][6]float32
		var kept []int
		for i, q := range centres {
			if math.Abs(values[i]) > half*2 {
				continue // keep cells near the surface
			}
			kept = append(kept, i)
			boxes = append(boxes, [6]float32{float32(q[0] - half), float32(q[1] - half), float32(q[2] - half),
				float32(q[0] + half), float32(q[1] + half), float32(q[2] + half)})
		}
		ctx := t.Context()
		if err := startQuery(ctx); err != nil {
			t.Fatal(err)
		}
		e := New(&source.MockManager{}, nil)
		prog, err := e.prepareQuery(ctx, nil, doc, "", []string{name})
		if err != nil {
			endQuery()
			t.Fatalf("assemble %s: %v", name, err)
		}
		results, err := EvalInterval(prog, boxes)
		endQuery()
		if err != nil {
			t.Fatalf("interval %s: %v", name, err)
		}
		widest, unsound := 0.0, 0
		for j, i := range kept {
			lo, hi := float64(results[0][j].Lower), float64(results[0][j].Upper)
			widest = math.Max(widest, (hi-lo)/diagonal)
			if values[i] < lo-1e-5 || values[i] > hi+1e-5 {
				unsound++
			}
		}
		t.Logf("%s: %d surface cells, widest %.2f diagonals, %d unsound", name, len(kept), widest, unsound)
		if unsound > 0 || widest > 4 {
			t.Errorf("%s: %d unsound cells, widest %.2f diagonals", name, unsound, widest)
		}
	}
}
