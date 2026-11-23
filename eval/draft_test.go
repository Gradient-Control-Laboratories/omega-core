//go:build cgo

package eval

import (
	"fmt"
	"math"
	"testing"
)

// The draft package: a face turned about its hinge, a circle drafted into a cone, and the draft
// of a point, checked against the closed forms of the surfaces they make.

// checkDraftZeros asserts the field "f" of doc is zero at each point.
func checkDraftZeros(t *testing.T, doc string, points [][3]float64) {
	t.Helper()
	got := evalChamferField(t, doc, "f", points)
	for i, q := range points {
		if math.Abs(got[i]) > 1e-5 {
			t.Errorf("at %v: %.7f, want 0 (on the drafted surface)", q, got[i])
		}
	}
}

// checkDraftUnitGradient asserts, by central differences, that the field "f" of doc has unit
// gradient at each point.
func checkDraftUnitGradient(t *testing.T, doc string, points [][3]float64) {
	t.Helper()
	const step = 1e-3
	var probe [][3]float64
	for _, q := range points {
		for axis := 0; axis < 3; axis++ {
			a, b := q, q
			a[axis] += step
			b[axis] -= step
			probe = append(probe, a, b)
		}
	}
	got := evalChamferField(t, doc, "f", probe)
	for i, q := range points {
		var g float64
		for axis := 0; axis < 3; axis++ {
			d := (got[6*i+2*axis] - got[6*i+2*axis+1]) / (2 * step)
			g += d * d
		}
		if g = math.Sqrt(g); math.Abs(g-1) > 1e-3 {
			t.Errorf("at %v: gradient %.5f, want 1", q, g)
		}
	}
}

// A box's side x = 20, the material at x < 20, drafted about the neutral plane z = 0: the plane
// through the hinge line (x = 20, z = 0) leaning `angle` from +z, inward as it rises for a
// positive angle, so at height h it is at x = 20 - h·tan(angle).
func TestDraft_HorizontalCurveTurnsAFaceAboutItsHinge(t *testing.T) {
	for _, degrees := range []float64{5, -5, 30} {
		t.Run(fmt.Sprintf("%v degrees", degrees), func(t *testing.T) {
			angle := degrees * math.Pi / 180
			doc := fmt.Sprintf("@scene(40)\nf = draft.horizontalCurve(p.x - 20, p.z, radians(%v));\nreturn f;\n", degrees)
			var zeros, others [][3]float64
			for _, h := range []float64{-4, 0, 3, 10} {
				for _, y := range []float64{-7, 0, 5} {
					zeros = append(zeros, [3]float64{20 - h*math.Tan(angle), y, h})
					others = append(others, [3]float64{20 - h*math.Tan(angle) + 1.3, y, h})
				}
			}
			checkDraftZeros(t, doc, zeros)
			checkDraftUnitGradient(t, doc, others)
			// Off the surface, the field is the distance to the drafted plane, outside positive.
			got := evalChamferField(t, doc, "f", others)
			for i, q := range others {
				if want := 1.3 * math.Cos(angle); math.Abs(got[i]-want) > 1e-5 {
					t.Errorf("at %v: %.7f, want %.7f, 1.3 out along x", q, got[i], want)
				}
			}
		})
	}
}

// A circle of radius 10 at z = 2 drafted by `angle`: the cone whose radius at height z is
// 10 - (z - 2)·tan(angle), as the section of a boss or a tapered hole.
func TestDraft_HorizontalCurveOfACircleIsACone(t *testing.T) {
	for _, degrees := range []float64{5, -20, 45} {
		t.Run(fmt.Sprintf("%v degrees", degrees), func(t *testing.T) {
			angle := degrees * math.Pi / 180
			doc := fmt.Sprintf("@scene(40)\nf = draft.horizontalCurve(length(p.xy) - 10, p.z - 2, radians(%v));\nreturn f;\n", degrees)
			var zeros, others [][3]float64
			for _, z := range []float64{-3, 2, 6} {
				r := 10 - (z-2)*math.Tan(angle)
				for _, a := range []float64{0.3, 2.1, 4.4} {
					zeros = append(zeros, [3]float64{r * math.Cos(a), r * math.Sin(a), z})
					others = append(others, [3]float64{(r + 2) * math.Cos(a), (r + 2) * math.Sin(a), z})
				}
			}
			checkDraftZeros(t, doc, zeros)
			checkDraftUnitGradient(t, doc, others)
		})
	}
}

// The draft of a point: the cone with its apex at the point whose side leans `angle` from the
// pull, opening against it for a positive angle; here with a pull off the axes.
func TestDraft_PointIsACone(t *testing.T) {
	center := [3]float64{1, 2, 3}
	pull := [3]float64{1 / math.Sqrt(3), 1 / math.Sqrt(3), 1 / math.Sqrt(3)}
	// Two unit vectors perpendicular to the pull.
	e1 := [3]float64{1 / math.Sqrt(2), -1 / math.Sqrt(2), 0}
	e2 := [3]float64{1 / math.Sqrt(6), 1 / math.Sqrt(6), -2 / math.Sqrt(6)}
	at := func(down, out, turn float64) [3]float64 {
		// down along -pull from the apex, out from the axis in the direction turn.
		var q [3]float64
		for i := range q {
			q[i] = center[i] - down*pull[i] + out*(math.Cos(turn)*e1[i]+math.Sin(turn)*e2[i])
		}
		return q
	}
	for _, degrees := range []float64{10, 45} {
		t.Run(fmt.Sprintf("%v degrees", degrees), func(t *testing.T) {
			angle := degrees * math.Pi / 180
			doc := fmt.Sprintf("@scene(40)\nf = draft.point([1, 2, 3], normalize([1, 1, 1]), radians(%v));\nreturn f;\n", degrees)
			var zeros, others [][3]float64
			for _, down := range []float64{1, 4} {
				for _, turn := range []float64{0, 1.9, 4} {
					zeros = append(zeros, at(down, down*math.Tan(angle), turn))
					others = append(others, at(down, down*math.Tan(angle)+1, turn))
				}
			}
			checkDraftZeros(t, doc, zeros)
			checkDraftUnitGradient(t, doc, others)
			// On the axis below the apex the point is inside the cone, its distance to the side
			// down·sin(angle).
			got := evalChamferField(t, doc, "f", [][3]float64{at(3, 0, 0)})[0]
			if want := -3 * math.Sin(angle); math.Abs(got-want) > 1e-5 {
				t.Errorf("on the axis 3 below the apex: %.7f, want %.7f", got, want)
			}
		})
	}
}
