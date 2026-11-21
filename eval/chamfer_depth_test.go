//go:build cgo

package eval

import (
	"fmt"
	"math"
	"testing"
)

// A chamfer by depth, setback and width (the Normalized chamfers and their two conversions),
// checked on a wedge of planes through the z axis against the closed forms: with φ the angle
// between the faces through the material, a bevel set back s along each face is s·cos(φ/2) deep
// and s·2sin(φ/2) wide, so a width w is (w/2)·cot(φ/2) deep.

// chamferWedge is wedgeDoc's wedge: faces d1 = p·n1 and d2 = p·n2 through the z axis, the
// material where both are negative, between the directions u1 and u2 that run from the edge
// along each face. bisector is the bevel's outward normal, (n1 + n2) normalized.
type chamferWedge struct {
	n1, n2, u1, u2, bisector [3]float64
	phi                      float64 // the angle between the faces, through the material
}

func newChamferWedge(interiorDegrees float64) chamferWedge {
	half := (180 - interiorDegrees) / 2 * math.Pi / 180
	return chamferWedge{
		n1:       [3]float64{math.Cos(half), math.Sin(half), 0},
		n2:       [3]float64{math.Cos(half), -math.Sin(half), 0},
		u1:       [3]float64{-math.Sin(half), math.Cos(half), 0},
		u2:       [3]float64{-math.Sin(half), -math.Cos(half), 0},
		bisector: [3]float64{1, 0, 0},
		phi:      interiorDegrees * math.Pi / 180,
	}
}

// doc chamfers the wedge with depth `depthExpr` (GSL, which may use n1, n2, d1 and d2), as
// "normals" (normals passed) and "gradients" (taken from the fields), and returns the depth
// itself as "depth".
func (w chamferWedge) doc(depthExpr string) string {
	vec := func(v [3]float64) string { return fmt.Sprintf("[%.12f, %.12f, 0.0]", v[0], v[1]) }
	return fmt.Sprintf(`@scene(40)
n1 = %s;
n2 = %s;
d1 = dot(p, n1);
d2 = dot(p, n2);
depth = %s;
normals = intersectChamferNormalized(d1, d2, depth, n1, n2);
gradients = intersectChamferNormalized(d1, d2, depth);
return normals;
`, vec(w.n1), vec(w.n2), depthExpr)
}

// checkBevel asserts the bevel through a and b, the points where it meets the two faces: zero
// at both and at their midpoint, and, along the bisector from the midpoint, a field that reads
// the distance (unit gradient).
func checkBevel(t *testing.T, doc string, a, b, bisector [3]float64) {
	t.Helper()
	mid := chamferScale(chamferAdd(a, b), 0.5)
	const delta = 0.05
	points := [][3]float64{a, b, mid, chamferAdd(mid, chamferScale(bisector, delta)), chamferAdd(mid, chamferScale(bisector, -delta))}
	want := []float64{0, 0, 0, delta, -delta}
	for _, name := range []string{"normals", "gradients"} {
		got := evalChamferField(t, doc, name, points)
		for i := range points {
			if math.Abs(got[i]-want[i]) > 2e-5 {
				t.Errorf("%s at %v: %.7f, want %.7f", name, points[i], got[i], want[i])
			}
		}
	}
}

func TestChamferDepth_IsTheBevelsDistanceFromTheEdge(t *testing.T) {
	for _, interior := range []float64{60, 90, 120, 150} {
		t.Run(fmt.Sprintf("%v degrees", interior), func(t *testing.T) {
			w := newChamferWedge(interior)
			const depth = 1.5
			doc := w.doc(fmt.Sprintf("%v", depth))
			// The bevel meets each face depth/cos(φ/2) from the edge.
			s := depth / math.Cos(w.phi/2)
			checkBevel(t, doc, chamferScale(w.u1, s), chamferScale(w.u2, s), w.bisector)
		})
	}
}

func TestChamferDepth_FromSetback(t *testing.T) {
	for _, interior := range []float64{60, 90, 120, 150} {
		t.Run(fmt.Sprintf("%v degrees", interior), func(t *testing.T) {
			w := newChamferWedge(interior)
			const setback = 1.5
			doc := w.doc(fmt.Sprintf("chamferDepthFromSetback(%v, n1, n2)", setback))
			depth := evalChamferField(t, doc, "depth", [][3]float64{{0, 0, 0}})[0]
			if want := setback * math.Cos(w.phi/2); math.Abs(depth-want) > 1e-6 {
				t.Errorf("depth %.7f, want s·cos(φ/2) = %.7f", depth, want)
			}
			checkBevel(t, doc, chamferScale(w.u1, setback), chamferScale(w.u2, setback), w.bisector)
		})
	}
}

func TestChamferDepth_FromWidth(t *testing.T) {
	for _, interior := range []float64{60, 90, 120, 150} {
		t.Run(fmt.Sprintf("%v degrees", interior), func(t *testing.T) {
			w := newChamferWedge(interior)
			const width = 2.0
			doc := w.doc(fmt.Sprintf("chamferDepthFromWidth(%v, n1, n2)", width))
			depth := evalChamferField(t, doc, "depth", [][3]float64{{0, 0, 0}})[0]
			if want := width / 2 / math.Tan(w.phi/2); math.Abs(depth-want) > 1e-6 {
				t.Errorf("depth %.7f, want (w/2)·cot(φ/2) = %.7f", depth, want)
			}
			// A bevel `width` across meets each face width/(2sin(φ/2)) from the edge.
			s := width / (2 * math.Sin(w.phi/2))
			a, b := chamferScale(w.u1, s), chamferScale(w.u2, s)
			if got := math.Hypot(a[0]-b[0], a[1]-b[1]); math.Abs(got-width) > 1e-9 {
				t.Fatalf("the test's own points are %.9f apart, want %v", got, width)
			}
			checkBevel(t, doc, a, b, w.bisector)
		})
	}
}

// The union counterpart fills the empty wedge between two solids: its bevel meets each solid's
// face, outside the other, `setback` from the edge, by the same conversion.
func TestChamferDepth_UnionFromSetback(t *testing.T) {
	for _, interior := range []float64{60, 90, 120} {
		t.Run(fmt.Sprintf("%v degrees", interior), func(t *testing.T) {
			w := newChamferWedge(interior)
			const setback = 1.5
			vec := func(v [3]float64) string { return fmt.Sprintf("[%.12f, %.12f, 0.0]", v[0], v[1]) }
			// Two solids, d1 <= 0 and d2 <= 0, whose union leaves the wedge d1 > 0, d2 > 0
			// empty; its faces run from the edge along -u1 and -u2.
			doc := fmt.Sprintf(`@scene(40)
n1 = %s;
n2 = %s;
d1 = dot(p, n1);
d2 = dot(p, n2);
depth = chamferDepthFromSetback(%v, n1, n2);
normals = unionChamferNormalized(d1, d2, depth, n1, n2);
gradients = unionChamferNormalized(d1, d2, depth);
return normals;
`, vec(w.n1), vec(w.n2), setback)
			// The bevel's outward normal points out of the filled material, into the wedge.
			checkBevel(t, doc, chamferScale(w.u1, -setback), chamferScale(w.u2, -setback), w.bisector)
		})
	}
}

// docWith is the wedge with "normals" and "gradients" given as GSL expressions in n1, n2, d1 and
// d2, for the chamfers that have no three-argument form.
func (w chamferWedge) docWith(normals, gradients string) string {
	vec := func(v [3]float64) string { return fmt.Sprintf("[%.12f, %.12f, 0.0]", v[0], v[1]) }
	return fmt.Sprintf(`@scene(40)
n1 = %s;
n2 = %s;
d1 = dot(p, n1);
d2 = dot(p, n2);
normals = %s;
gradients = %s;
return normals;
`, vec(w.n1), vec(w.n2), normals, gradients)
}

// setbacksNormal is the outward normal of a bevel set back s1 along the first face and s2 along
// the second: (s1·n1 + s2·n2) normalized.
func setbacksNormal(w chamferWedge, s1, s2 float64) [3]float64 {
	v := chamferAdd(chamferScale(w.n1, s1), chamferScale(w.n2, s2))
	return chamferScale(v, 1/math.Sqrt(v[0]*v[0]+v[1]*v[1]+v[2]*v[2]))
}

func TestChamferSetbacks_MeetsEachFaceAtItsSetback(t *testing.T) {
	for _, interior := range []float64{60, 90, 120, 150} {
		for _, s := range [][2]float64{{1, 2}, {2.5, 0.5}} {
			t.Run(fmt.Sprintf("%v degrees, %v and %v", interior, s[0], s[1]), func(t *testing.T) {
				w := newChamferWedge(interior)
				doc := w.docWith(
					fmt.Sprintf("intersectChamferSetbacks(d1, d2, %v, %v, n1, n2)", s[0], s[1]),
					fmt.Sprintf("intersectChamferSetbacks(d1, d2, %v, %v, d1.dp, d2.dp)", s[0], s[1]))
				checkBevel(t, doc, chamferScale(w.u1, s[0]), chamferScale(w.u2, s[1]), setbacksNormal(w, s[0], s[1]))
			})
		}
	}
}

// Equal setbacks are the Normalized chamfer at chamferDepthFromSetback's depth, everywhere.
func TestChamferSetbacks_EqualIsNormalizedFromSetback(t *testing.T) {
	for _, interior := range []float64{60, 90, 120, 150} {
		t.Run(fmt.Sprintf("%v degrees", interior), func(t *testing.T) {
			w := newChamferWedge(interior)
			const setback = 1.5
			doc := w.docWith(
				fmt.Sprintf("intersectChamferSetbacks(d1, d2, %v, %v, n1, n2)", setback, setback),
				fmt.Sprintf("intersectChamferNormalized(d1, d2, chamferDepthFromSetback(%v, n1, n2), n1, n2)", setback))
			var points [][3]float64
			for x := -4.0; x <= 4; x += 0.5 {
				for y := -4.0; y <= 4; y += 0.5 {
					points = append(points, [3]float64{x + 0.013, y + 0.029, 0})
				}
			}
			setbacks := evalChamferField(t, doc, "normals", points)
			normalized := evalChamferField(t, doc, "gradients", points)
			for i := range points {
				if math.Abs(setbacks[i]-normalized[i]) > 1e-5 {
					t.Errorf("at %v: setbacks %.7f, normalized %.7f", points[i], setbacks[i], normalized[i])
				}
			}
		})
	}
}

// The union counterpart's bevel meets each solid's face, outside the other, at that solid's
// setback; the difference counterpart's meets the target's face at the target's setback and the
// cut's at the tool's.
func TestChamferSetbacks_UnionAndDifference(t *testing.T) {
	const s1, s2 = 1.0, 2.0
	for _, interior := range []float64{60, 90, 120} {
		w := newChamferWedge(interior)
		t.Run(fmt.Sprintf("union, %v degrees", interior), func(t *testing.T) {
			doc := w.docWith(
				fmt.Sprintf("unionChamferSetbacks(d1, d2, %v, %v, n1, n2)", s1, s2),
				fmt.Sprintf("unionChamferSetbacks(d1, d2, %v, %v, d1.dp, d2.dp)", s1, s2))
			checkBevel(t, doc, chamferScale(w.u1, -s1), chamferScale(w.u2, -s2), setbacksNormal(w, s1, s2))
		})
		t.Run(fmt.Sprintf("difference, %v degrees", interior), func(t *testing.T) {
			// The tool is d2 >= 0, its outward normal -n2; cutting it leaves the wedge.
			doc := w.docWith(
				fmt.Sprintf("differenceChamferSetbacks(d1, neg(d2), %v, %v, n1, neg(n2))", s1, s2),
				fmt.Sprintf("differenceChamferSetbacks(d1, neg(d2), %v, %v, d1.dp, neg(d2.dp))", s1, s2))
			checkBevel(t, doc, chamferScale(w.u1, s1), chamferScale(w.u2, s2), setbacksNormal(w, s1, s2))
		})
	}
}

func chamferAdd(a, b [3]float64) [3]float64 {
	return [3]float64{a[0] + b[0], a[1] + b[1], a[2] + b[2]}
}

func chamferScale(a [3]float64, s float64) [3]float64 {
	return [3]float64{a[0] * s, a[1] * s, a[2] * s}
}
