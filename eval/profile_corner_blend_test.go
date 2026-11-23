//go:build cgo

package eval

import (
	"context"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/source"
)

// cornerBlend(): a corner whose two edges are trimmed back to half the
// shorter one and joined by a piece whose field is the profile's blend
// operator applied to the two edges' lines (its dual at a reflex corner).
// Outside those pieces the profile is exact; inside them it is the operator.

// blendCornerIndices are the U's corners given cornerBlend() here: a convex
// one and a reflex one. The rest are an exact fillet, a chamfer and sharp.
var blendCornerIndices = []int{2, 5}

func uBlendCorners() ([]string, []refCorner) {
	gsl := []string{"cornerFillet(3)", "cornerChamfer(2)", "cornerBlend()", "cornerSharp()",
		"cornerFillet(3)", "cornerBlend()", "cornerSharp()", "cornerSharp()"}
	// The reference outside the blended pieces: the blended corners sharp.
	ref := []refCorner{{filletCorner, 3}, {chamferCorner, 2}, {sharpCorner, 0}, {sharpCorner, 0},
		{filletCorner, 3}, {sharpCorner, 0}, {sharpCorner, 0}, {sharpCorner, 0}}
	return gsl, ref
}

var blendOperators = []string{"intersectRound(1.5)", "intersectChamfer(1.5)", "intersectLp(1.5, 4)"}

// appliedBlend is the operator named by blend applied to fields, in its list
// form: GSL cannot call the result of a call.
func appliedBlend(blend, fields string) string {
	name, args, _ := strings.Cut(strings.TrimSuffix(blend, ")"), "(")
	return fmt.Sprintf("%s(%s, %s)", name, fields, args)
}

// blendDoc is the U with blended corners as `solid`, and each blended
// corner's two edge lines as lineIn<i> and lineOut<i>: signed distances,
// positive on the right of the edge's direction, outside a counterclockwise
// loop.
func blendDoc(blend string) string {
	gsl, _ := uBlendCorners()
	zeros := strings.TrimSuffix(strings.Repeat("0, ", len(uShape)), ", ")
	var b strings.Builder
	fmt.Fprintf(&b, "@scene(80)\nsolid = profileXY(%s, [%s], [%s], %s);\n", pointsGSL(uShape), zeros, strings.Join(gsl, ", "), blend)
	n := len(uShape)
	for _, i := range blendCornerIndices {
		prev, v, next := uShape[(i+n-1)%n], uShape[i], uShape[(i+1)%n]
		fmt.Fprintf(&b, "lineIn%d = %s;\nlineOut%d = %s;\n", i, lineGSL(prev, v), i, lineGSL(v, next))
	}
	b.WriteString("return solid;\n")
	return b.String()
}

// lineGSL is the signed distance to the line from a to b, positive on its
// right.
func lineGSL(a, b [2]float64) string {
	dx, dy := b[0]-a[0], b[1]-a[1]
	l := math.Hypot(dx, dy)
	return fmt.Sprintf("(p.x - %v) * %v - (p.y - %v) * %v", a[0], dy/l, a[1], dx/l)
}

// blendTrim is half the shorter edge at corner i of the U.
func blendTrim(i int) float64 {
	n := len(uShape)
	prev, v, next := uShape[(i+n-1)%n], uShape[i], uShape[(i+1)%n]
	return math.Min(math.Hypot(v[0]-prev[0], v[1]-prev[1]), math.Hypot(next[0]-v[0], next[1]-v[1])) / 2
}

// nearestOnBoundary is the point of the reference's sampled boundary nearest q.
func nearestOnBoundary(r refProfile, q [2]float64) [2]float64 {
	best, at := math.Inf(1), [2]float64{}
	for _, s := range r.segments {
		a, b := s[0], s[1]
		ex, ey := b[0]-a[0], b[1]-a[1]
		h := math.Max(0, math.Min(1, ((q[0]-a[0])*ex+(q[1]-a[1])*ey)/(ex*ex+ey*ey)))
		c := [2]float64{a[0] + h*ex, a[1] + h*ey}
		if d := math.Hypot(q[0]-c[0], q[1]-c[1]); d < best {
			best, at = d, c
		}
	}
	return at
}

// Away from the blended pieces the profile is exact: wherever the nearest
// boundary point of the unblended reference, and the point itself, stand off
// every blended corner by more than its trim.
func TestProfileXY_BlendCornersLeaveTheRestExact(t *testing.T) {
	_, refCorners := uBlendCorners()
	ref := newRefProfile(t, blendedLoop(uShape, refCorners))
	var points [][3]float64
	for _, q := range gridAround(ref.box(), 5, 52) {
		near := nearestOnBoundary(ref, q)
		clear := true
		for _, i := range blendCornerIndices {
			v, trim := uShape[i], blendTrim(i)
			if math.Hypot(near[0]-v[0], near[1]-v[1]) <= trim+0.05 || math.Hypot(q[0]-v[0], q[1]-v[1]) <= trim+0.05 {
				clear = false
			}
		}
		if clear {
			points = append(points, [3]float64{q[0], q[1], 0})
		}
	}
	for _, blend := range blendOperators {
		t.Run(blend, func(t *testing.T) {
			got := evalProfileField(t, blendDoc(blend), "solid", points)
			compareToReference(t, blend, points, got, func(q [3]float64) float64 {
				return ref.distance([2]float64{q[0], q[1]})
			})
		})
	}
}

// Near a blended corner, inside its piece, the profile is the operator on the
// corner's two edge lines (the intersection blend at the convex corner, its
// dual, the union blend, at the reflex one), held to no less than the distance
// to the triangle of the corner and its two trims, which holds the blend's
// curve: the operator's sign, the greater of the two magnitudes.
func TestProfileXY_BlendCornerIsTheOperator(t *testing.T) {
	n := len(uShape)
	for _, blend := range blendOperators {
		t.Run(blend, func(t *testing.T) {
			for _, i := range blendCornerIndices {
				prev, v, next := uShape[(i+n-1)%n], uShape[i], uShape[(i+1)%n]
				convex := (v[0]-prev[0])*(next[1]-v[1])-(v[1]-prev[1])*(next[0]-v[0]) > 0
				operator := appliedBlend(blend, fmt.Sprintf("[lineIn%d, lineOut%d]", i, i))
				if !convex {
					operator = "neg(" + appliedBlend(blend, fmt.Sprintf("[neg(lineIn%d), neg(lineOut%d)]", i, i)) + ")"
				}
				// Points within a third of the trim of the corner, on both sides.
				var points [][3]float64
				r := blendTrim(i) / 3
				for _, f := range [][2]float64{{0, 0}, {1, 0}, {0, 1}, {-1, 0}, {0, -1}, {0.7, 0.7}, {-0.7, 0.7}, {0.7, -0.7}, {-0.7, -0.7}} {
					points = append(points, [3]float64{v[0] + r*f[0], v[1] + r*f[1], 0})
				}
				e := New(&source.MockManager{}, nil)
				groups, err := e.EvalGsl(context.Background(), nil, blendDoc(blend), "", []string{"solid", operator}, points)
				if err != nil {
					t.Fatalf("eval: %v", err)
				}
				trim := blendTrim(i)
				in, out := unit(v, prev), unit(next, v)
				a := [2]float64{v[0] - trim*in[0], v[1] - trim*in[1]}
				b := [2]float64{v[0] + trim*out[0], v[1] + trim*out[1]}
				bounded := 0
				for k, q := range points {
					got, operator := float64(groups[0].Samples[0][k].Value), float64(groups[1].Samples[0][k].Value)
					triangle := triangleDistance([2]float64{q[0], q[1]}, a, v, b)
					want := math.Copysign(math.Max(math.Abs(operator), triangle), operator)
					if triangle > math.Abs(operator) {
						bounded++
					}
					if math.Abs(got-want) > profileTolerance {
						t.Errorf("corner %d at %v: solid %v, want %v (operator %v, triangle %v)", i, q, got, want, operator, triangle)
					}
				}
				t.Logf("corner %d: %d of %d points held up by the triangle", i, bounded, len(points))
			}
		})
	}
}

// The blended profile is continuous and 1-Lipschitz round its blended
// corners: no step where a piece meets the edges beyond its trims, and no
// slope steeper than a distance's.
func TestProfileXY_BlendCornersAreLipschitz(t *testing.T) {
	const h = 0.05
	for _, blend := range blendOperators {
		t.Run(blend, func(t *testing.T) {
			var points [][3]float64
			for _, i := range blendCornerIndices {
				v, reach := uShape[i], blendTrim(i)+1
				for x := -reach; x <= reach; x += 0.25 {
					for y := -reach; y <= reach; y += 0.25 {
						q := [3]float64{v[0] + x + 0.0123, v[1] + y + 0.0071, 0}
						points = append(points, q, [3]float64{q[0] + h, q[1], 0}, [3]float64{q[0], q[1] + h, 0})
					}
				}
			}
			got := evalProfileField(t, blendDoc(blend), "solid", points)
			steepest := 0.0
			for k := 0; k < len(points); k += 3 {
				for _, j := range []int{k + 1, k + 2} {
					slope := math.Abs(got[j]-got[k]) / h
					steepest = math.Max(steepest, slope)
					if slope > 1+1e-3+profileTolerance/h {
						t.Errorf("slope %.4f between %v and %v", slope, points[k], points[j])
					}
				}
			}
			t.Logf("%s: steepest slope %.4f over %d pairs", blend, steepest, len(points)*2/3)
		})
	}
}

// Its surface cells stay sound and bounded with each operator.
func TestProfileXY_BlendCornersSurfaceCells(t *testing.T) {
	_, refCorners := uBlendCorners()
	ref := newRefProfile(t, blendedLoop(uShape, refCorners))
	for _, blend := range blendOperators {
		t.Run(blend, func(t *testing.T) {
			doc := blendDoc(blend)
			centres := flat(ref.boundarySamples(0.75))
			values := evalProfileField(t, doc, "solid", centres)
			index := make(map[[3]float64]float64, len(centres))
			for k, q := range centres {
				index[q] = values[k]
			}
			checkSurfaceCellsOfHalf(t, blend, doc, centres, 0.15, func(q [3]float64) float64 { return index[q] })
		})
	}
}

// Far from its blended corners the profile keeps the sharp profile's sign,
// over a grid reaching well past the part: the blend of two lines must not
// claim a point that belongs to another part of the loop.
func TestProfileXY_BlendCornersKeepTheSignFarOff(t *testing.T) {
	_, refCorners := uBlendCorners()
	ref := newRefProfile(t, blendedLoop(uShape, refCorners))
	var points [][3]float64
	for _, q := range gridAround([4]float64{-60, -60, 90, 85}, 0, 70) {
		clear := true
		for _, i := range blendCornerIndices {
			v := uShape[i]
			if math.Hypot(q[0]-v[0], q[1]-v[1]) <= blendTrim(i)*math.Sqrt2+0.1 {
				clear = false
			}
		}
		if clear && math.Abs(ref.distance(q)) > 0.01 {
			points = append(points, [3]float64{q[0], q[1], 0})
		}
	}
	for _, blend := range blendOperators {
		t.Run(blend, func(t *testing.T) {
			got := evalProfileField(t, blendDoc(blend), "solid", points)
			wrong := 0
			for k, q := range points {
				want := ref.distance([2]float64{q[0], q[1]})
				if (got[k] < 0) != (want < 0) {
					if wrong++; wrong <= 3 {
						t.Errorf("at %v solid %v, sharp reference %v: opposite signs", q, got[k], want)
					}
				}
			}
			t.Logf("%s: %d points, %d with the wrong sign", blend, len(points), wrong)
		})
	}
}

// unit is the direction from a to b.
func unit(b, a [2]float64) [2]float64 {
	l := math.Hypot(b[0]-a[0], b[1]-a[1])
	return [2]float64{(b[0] - a[0]) / l, (b[1] - a[1]) / l}
}

// triangleDistance is the distance from q to the triangle a, b, c: 0 inside.
func triangleDistance(q, a, b, c [2]float64) float64 {
	cross := func(o, u, w [2]float64) float64 { return (u[0]-o[0])*(w[1]-o[1]) - (u[1]-o[1])*(w[0]-o[0]) }
	s1, s2, s3 := cross(a, b, q), cross(b, c, q), cross(c, a, q)
	if (s1 >= 0 && s2 >= 0 && s3 >= 0) || (s1 <= 0 && s2 <= 0 && s3 <= 0) {
		return 0
	}
	segment := func(u, w [2]float64) float64 {
		ex, ey := w[0]-u[0], w[1]-u[1]
		h := math.Max(0, math.Min(1, ((q[0]-u[0])*ex+(q[1]-u[1])*ey)/(ex*ex+ey*ey)))
		return math.Hypot(q[0]-u[0]-h*ex, q[1]-u[1]-h*ey)
	}
	return math.Min(segment(a, b), math.Min(segment(b, c), segment(c, a)))
}
