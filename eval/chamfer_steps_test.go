//go:build cgo

package eval

import (
	"fmt"
	"math"
	"testing"
)

// The Steps chamfers, checked against closed forms computed here: on a wedge of two planes the
// chamfer must be the plane through its two setback points (or through its start at its angle), and
// on a revolved part the cone through the points the description gives, which in the plane through
// the axis are straight-line constructions. Each check asks for zero at the chamfer's two ends and its
// middle, and, where the bevel decides the field, the distance to it a little off either side.

const stepsTolerance = 2e-4

// stepsDoc is a document whose field "f" is expr.
func stepsDoc(expr string) string {
	return "@scene(40)\nf = " + expr + ";\nreturn f;\n"
}

// checkStepsLine asserts the field "f" is zero at a, b and their midpoint, and reads ±delta at
// delta along `outward` (the bevel's unit normal, out of the material) from the midpoint.
func checkStepsLine(t *testing.T, doc string, a, b, outward [3]float64) {
	t.Helper()
	mid := [3]float64{(a[0] + b[0]) / 2, (a[1] + b[1]) / 2, (a[2] + b[2]) / 2}
	const delta = 0.02
	off := func(s float64) [3]float64 {
		return [3]float64{mid[0] + s*outward[0], mid[1] + s*outward[1], mid[2] + s*outward[2]}
	}
	points := [][3]float64{a, b, mid, off(delta), off(-delta)}
	want := []float64{0, 0, 0, delta, -delta}
	got := evalChamferField(t, doc, "f", points)
	for i := range points {
		if math.Abs(got[i]-want[i]) > stepsTolerance {
			t.Errorf("at %v: %.6f, want %.6f", points[i], got[i], want[i])
		}
	}
}

// stepsWedge is two planes through the z axis, the material where both are negative, between the
// directions u1 and u2 that run from the edge along each face.
type stepsWedge struct{ n1, n2, u1, u2 [3]float64 }

func newStepsWedge(interiorDegrees float64) stepsWedge {
	h := (180 - interiorDegrees) / 2 * math.Pi / 180
	return stepsWedge{
		n1: [3]float64{math.Cos(h), math.Sin(h), 0}, n2: [3]float64{math.Cos(h), -math.Sin(h), 0},
		u1: [3]float64{-math.Sin(h), math.Cos(h), 0}, u2: [3]float64{-math.Sin(h), -math.Cos(h), 0},
	}
}

func (w stepsWedge) fields() (string, string) {
	return fmt.Sprintf("dot(p, [%.12f, %.12f, 0.0])", w.n1[0], w.n1[1]), fmt.Sprintf("dot(p, [%.12f, %.12f, 0.0])", w.n2[0], w.n2[1])
}

// unit3 is a along its own length; add3, sub3, scale3 and dot3 are litrender.go's.
func unit3(a [3]float64) [3]float64 { return scale3(a, 1/math.Sqrt(dot3(a, a))) }

// towardCorner is the unit normal of the line from a to b in the plane of a, b and the point c, on
// c's side: out of the material for a cut, whose corner c it removes, so the field's outward normal;
// into it for a fill, whose corner it fills.
func towardCorner(a, b, c [3]float64) [3]float64 {
	along := unit3(sub3(b, a))
	e := sub3(c, a)
	return unit3(sub3(e, scale3(along, dot3(e, along))))
}

// On flat faces the Steps chamfer is the plane through its two setback points, at any angle between
// the faces.
func TestChamferSteps_FlatWedgeSetbacks(t *testing.T) {
	for _, interior := range []float64{60, 90, 120, 150} {
		t.Run(fmt.Sprintf("%v degrees", interior), func(t *testing.T) {
			w := newStepsWedge(interior)
			f1, f2 := w.fields()
			doc := stepsDoc(fmt.Sprintf("intersectChamferSetbacksSteps(%s, %s, 1, 2, 5)", f1, f2))
			z := [3]float64{0, 0, 0.3}
			a, b, edge := add3(scale3(w.u1, 1), z), add3(scale3(w.u2, 2), z), z
			checkStepsLine(t, doc, a, b, towardCorner(a, b, edge))
		})
	}
}

// By a distance and an angle on flat faces: the plane from the start, 1.5 along face 1, at the angle
// to face 1 back toward the edge and into the material, to where it meets face 2. The angle and the
// angle between the faces sum to at most 165 degrees, short of 180, where the chamfer would run
// parallel to face 2 and never meet it.
func TestChamferSteps_FlatWedgeAngle(t *testing.T) {
	for _, interior := range []float64{60, 90, 120, 150} {
		for _, angle := range []float64{15, 30, 45, 60} {
			if interior+angle > 165 {
				continue
			}
			t.Run(fmt.Sprintf("%v degrees, %v", interior, angle), func(t *testing.T) {
				w := newStepsWedge(interior)
				f1, f2 := w.fields()
				doc := stepsDoc(fmt.Sprintf("intersectChamferAngleSteps(%s, %s, 1.5, radians(%v), 5)", f1, f2, angle))
				beta := angle * math.Pi / 180
				z := [3]float64{0, 0, -0.4}
				start := add3(scale3(w.u1, 1.5), z)
				dir := add3(scale3(w.u1, -math.Cos(beta)), scale3(w.n1, -math.Sin(beta)))
				end := add3(start, scale3(dir, -dot3(start, w.n2)/dot3(dir, w.n2)))
				checkStepsLine(t, doc, start, end, towardCorner(start, end, z))
			})
		}
	}
}

// at puts a point of the plane through the z axis, (rho, z), at the given azimuth.
func at(rho, z, azimuth float64) [3]float64 {
	return [3]float64{rho * math.Cos(azimuth), rho * math.Sin(azimuth), z}
}

// The upper half of a ball of radius 5, its rim chamfered 2 on the flat and 1 on the ball: the
// cone through the circle of radius 3 on the flat and the ball's circle a chord 1 from the rim.
func TestChamferSteps_HemisphereTwoDistances(t *testing.T) {
	doc := stepsDoc("intersectChamferSetbacksSteps(neg(p.z), length(p) - 5, 2, 1, 5)")
	phi := 2 * math.Asin(1.0/10)
	for _, az := range []float64{0, 1.3, 4} {
		a, b, rim := at(3, 0, az), at(5*math.Cos(phi), 5*math.Sin(phi), az), at(5, 0, az)
		checkStepsLine(t, doc, a, b, towardCorner(a, b, rim))
	}
}

// The same rim chamfered 1 in on the flat at 45° to it: the cone from the circle of radius 4,
// rising at 45° toward the rim's side until it meets the ball.
func TestChamferSteps_HemisphereAngle(t *testing.T) {
	doc := stepsDoc("intersectChamferAngleSteps(neg(p.z), length(p) - 5, 1, radians(45), 5)")
	// From (4, 0) along (1, 1)/√2 to the ball: |(4 + s/√2, s/√2)| = 5.
	s := (-4*math.Sqrt2 + math.Sqrt(32-4*(16-25))) / 2
	for _, az := range []float64{0, 2.2} {
		a, b, rim := at(4, 0, az), at(4+s/math.Sqrt2, s/math.Sqrt2, az), at(5, 0, az)
		checkStepsLine(t, doc, a, b, towardCorner(a, b, rim))
	}
}

// A plate of radius 8 from z = -6 to 0 and a ball of radius 5 about the origin, the groove between
// them filled from 2 out on the plate at 45°: the cone from the circle of radius 7, rising back toward
// the ball until it first meets it. And the fill stays in the groove: under the plate, where its
// bottom and the ball's lowest point are 1 apart, and past its rim, the field is the plain union's.
func TestChamferSteps_UnionFillsTheGroove(t *testing.T) {
	plate := "max(length(p.xy) - 8, abs(p.z + 3) - 3)"
	doc := stepsDoc(fmt.Sprintf("unionChamferAngleSteps(%s, length(p) - 5, 2, radians(45), 5)", plate))
	// From (7, 0) along (-1, 1)/√2 to the ball, the nearer crossing.
	b2 := -7 / math.Sqrt2
	s := -b2 - math.Sqrt(b2*b2-(49-25))
	a, b, edge := at(7, 0, 0.5), at(7-s/math.Sqrt2, s/math.Sqrt2, 0.5), at(5, 0, 0.5)
	// A fill's material is on the corner's side, so its outward normal points away from the corner.
	checkStepsLine(t, doc, a, b, scale3(towardCorner(a, b, edge), -1))

	sharp := stepsDoc(fmt.Sprintf("min(%s, length(p) - 5)", plate))
	points := [][3]float64{at(0.3, -6.3, 0), at(1, -6.5, 1), at(8.1, -1.1, 0), at(8.3, -0.6, 2)}
	got, want := evalChamferField(t, doc, "f", points), evalChamferField(t, sharp, "f", points)
	for i, q := range points {
		if math.Abs(got[i]-want[i]) > 1e-5 {
			t.Errorf("at %v: %.6f, want the plain union's %.6f", q, got[i], want[i])
		}
	}
}

// A ball bored through: the bore and the ball meet in two rims, and the seed selects one. Starting
// the steps from p reflected into z >= 0 chamfers the upper rim, 2 down the bore and 1 on the ball;
// near the lower rim the field is the plain bored ball's.
func TestChamferSteps_SeedSelectsAnEdge(t *testing.T) {
	doc := stepsDoc("intersectChamferSetbacksSteps(2 - length(p.xy), length(p) - 5, 2, 1, 5, [p.x, p.y, abs(p.z)])")
	zr := math.Sqrt(21)
	theta := math.Atan2(zr, 2) - 2*math.Asin(1.0/10)
	a, b, rim := at(2, zr-2, 1), at(5*math.Cos(theta), 5*math.Sin(theta), 1), at(2, zr, 1)
	checkStepsLine(t, doc, a, b, towardCorner(a, b, rim))

	sharp := stepsDoc("max(2 - length(p.xy), length(p) - 5)")
	points := [][3]float64{at(2.05, -zr+0.2, 0), at(2.4, -zr+0.05, 1), at(2.2, -4.3, 2), at(2.6, -3.9, 3)}
	got, want := evalChamferField(t, doc, "f", points), evalChamferField(t, sharp, "f", points)
	for i, q := range points {
		if math.Abs(got[i]-want[i]) > 1e-5 {
			t.Errorf("near the lower rim at %v: %.6f, want the plain bored ball's %.6f", q, got[i], want[i])
		}
	}
}

// Far from the edge the frame means nothing (the faces' normals there may be parallel): the field
// must still be a number, and the solid there the sharp one.
func TestChamferSteps_FiniteFarFromTheEdge(t *testing.T) {
	cases := map[string][2]string{
		"intersection": {"intersectChamferAngleSteps(neg(p.z), length(p) - 5, 1, radians(45), 5)", "max(neg(p.z), length(p) - 5)"},
		"union": {"unionChamferAngleSteps(max(length(p.xy) - 8, abs(p.z + 3) - 3), length(p) - 5, 2, radians(45), 5)",
			"min(max(length(p.xy) - 8, abs(p.z + 3) - 3), length(p) - 5)"},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			var points [][3]float64
			for x := -12.0; x <= 12; x += 1.7 {
				for z := -10.0; z <= 14; z += 1.9 {
					points = append(points, [3]float64{x + 0.01, 0.37 * x, z})
				}
			}
			got, want := evalChamferField(t, stepsDoc(c[0]), "f", points), evalChamferField(t, stepsDoc(c[1]), "f", points)
			for i, q := range points {
				if math.IsNaN(got[i]) || math.IsInf(got[i], 0) {
					t.Fatalf("at %v: %v", q, got[i])
				}
				// Points within 3 of the rim's circle may be on the chamfer. Elsewhere a cut leaves the sharp
				// field as it is; a fill adds material, so the field there may be lower than the sharp
				// union's, nearer the fill, but never higher, and of the same sign.
				rho := math.Hypot(q[0], q[1])
				if math.Hypot(rho-5, q[2]) < 3 {
					continue
				}
				if name == "intersection" && math.Abs(got[i]-want[i]) > 1e-5 {
					t.Errorf("at %v, far from the rim: %.6f, want the sharp %.6f", q, got[i], want[i])
				}
				if name == "union" && (got[i] > want[i]+1e-5 || (got[i] > 0) != (want[i] > 0)) {
					t.Errorf("at %v, far from the rim: %.6f, against the sharp union's %.6f", q, got[i], want[i])
				}
			}
		})
	}
}
