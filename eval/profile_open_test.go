//go:build cgo

package eval

import (
	"context"
	"fmt"
	"math"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// openProfileXY: an open chain of lines and arcs whose two end edges carry on
// along their end tangents as rays, so the chain splits the plane. The field is
// the exact signed distance to that extended chain, positive to the right of
// travel, where a counterclockwise loop's outside is.

// openRayReach is how far the reference carries each ray: far beyond any point
// these tests evaluate, so the rays' far ends never come nearest.
const openRayReach = 1e4

// refChain is an open chain as openProfileXY takes it: points, and one signed
// curvature per edge, one fewer than the points (nil for straight edges).
type refChain struct {
	points     [][2]float64
	curvatures []float64
}

// gsl is the chain as an openProfileXY call.
func (c refChain) gsl() string {
	if c.curvatures == nil {
		return fmt.Sprintf("openProfileXY(%s)", pointsGSL(c.points))
	}
	ks := make([]string, len(c.curvatures))
	for i, k := range c.curvatures {
		if k < 0 {
			ks[i] = fmt.Sprintf("neg(%v)", -k)
		} else {
			ks[i] = fmt.Sprintf("%v", k)
		}
	}
	return fmt.Sprintf("openProfileXY(%s, [%s])", pointsGSL(c.points), strings.Join(ks, ", "))
}

func (c refChain) curvature(i int) float64 {
	if c.curvatures == nil {
		return 0
	}
	return c.curvatures[i]
}

// edgeTangents are the unit tangents at the two ends of the edge from a to b
// bent to curvature k: the chord turned back by asin(k * chord / 2) at a and
// forward by as much at b, as arcSamples leaves a.
func edgeTangents(a, b [2]float64, k float64) (start, end [2]float64) {
	dx, dy := b[0]-a[0], b[1]-a[1]
	chord := math.Hypot(dx, dy)
	half := math.Asin(k * chord / 2)
	turn := func(angle float64) [2]float64 {
		c, s := math.Cos(angle), math.Sin(angle)
		return [2]float64{(dx*c - dy*s) / chord, (dx*s + dy*c) / chord}
	}
	return turn(-half), turn(half)
}

// refOpen is the sampled extended chain: its segments in order from the far
// end of the start ray to the far end of the end ray, each with the distance
// along the chain at its start (negative on the start ray), and a closed
// polygon for the sign: the extended chain closed far away round its left side.
type refOpen struct {
	segments [][2][2]float64
	stations []float64
	closed   [][2]float64
}

func newRefOpen(t *testing.T, c refChain) refOpen {
	n := len(c.points)
	t0, _ := edgeTangents(c.points[0], c.points[1], c.curvature(0))
	_, te := edgeTangents(c.points[n-2], c.points[n-1], c.curvature(n-2))
	first, last := c.points[0], c.points[n-1]
	farStart := [2]float64{first[0] - openRayReach*t0[0], first[1] - openRayReach*t0[1]}
	farEnd := [2]float64{last[0] + openRayReach*te[0], last[1] + openRayReach*te[1]}

	line := [][2]float64{farStart}
	for i := 0; i+1 < n; i++ {
		line = append(line, arcSamples(t, c.points[i], c.points[i+1], c.curvature(i))...)
	}
	line = append(line, last, farEnd)

	var r refOpen
	station := -openRayReach
	for i := 0; i+1 < len(line); i++ {
		r.segments = append(r.segments, [2][2]float64{line[i], line[i+1]})
		r.stations = append(r.stations, station)
		station += math.Hypot(line[i+1][0]-line[i][0], line[i+1][1]-line[i][1])
	}

	// Close the extended chain counterclockwise round a circle twice as far
	// out, from the end ray's far end back to the start ray's: the region it
	// closes is the chain's left, its negative side.
	r.closed = append(r.closed, line...)
	radius := 2 * openRayReach
	from := math.Atan2(farEnd[1]-first[1], farEnd[0]-first[0])
	to := math.Atan2(farStart[1]-first[1], farStart[0]-first[0])
	sweep := to - from
	for sweep <= 0 {
		sweep += 2 * math.Pi
	}
	for i := 0; i <= 64; i++ {
		angle := from + sweep*float64(i)/64
		r.closed = append(r.closed, [2]float64{first[0] + radius*math.Cos(angle), first[1] + radius*math.Sin(angle)})
	}
	return r
}

// nearest is the distance from q to the extended chain and the distance along
// the chain to its nearest point.
func (r refOpen) nearest(q [2]float64) (distance, station float64) {
	distance = math.Inf(1)
	for i, s := range r.segments {
		a, b := s[0], s[1]
		ex, ey := b[0]-a[0], b[1]-a[1]
		length := math.Hypot(ex, ey)
		wx, wy := q[0]-a[0], q[1]-a[1]
		h := math.Max(0, math.Min(1, (wx*ex+wy*ey)/(length*length)))
		if d := math.Hypot(wx-h*ex, wy-h*ey); d < distance {
			distance, station = d, r.stations[i]+h*length
		}
	}
	return distance, station
}

// distance is the signed distance from q to the extended chain, negative on
// its left.
func (r refOpen) distance(q [2]float64) float64 {
	d, _ := r.nearest(q)
	inside := false
	for i, a := range r.closed {
		b := r.closed[(i+1)%len(r.closed)]
		if (a[1] > q[1]) != (b[1] > q[1]) && q[0] < a[0]+(q[1]-a[1])*(b[0]-a[0])/(b[1]-a[1]) {
			inside = !inside
		}
	}
	if inside {
		return -d
	}
	return d
}

// chainBox is the box of the chain's own segments, without the rays.
func (r refOpen) chainBox() [4]float64 {
	b := [4]float64{math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)}
	for _, s := range r.segments[1 : len(r.segments)-1] {
		for _, q := range s {
			b[0], b[1] = math.Min(b[0], q[0]), math.Min(b[1], q[1])
			b[2], b[3] = math.Max(b[2], q[0]), math.Max(b[3], q[1])
		}
	}
	return b
}

// samples are points on the extended chain about spacing apart, from `before`
// mm back along the start ray to `after` mm out along the end ray.
func (r refOpen) samples(spacing, before, after float64) [][2]float64 {
	last := len(r.segments) - 1
	end := r.stations[last] + after
	var out [][2]float64
	i := 0
	for at := -before; at <= end; at += spacing {
		for i < last && r.stations[i+1] <= at {
			i++
		}
		a, b := r.segments[i][0], r.segments[i][1]
		f := (at - r.stations[i]) / math.Hypot(b[0]-a[0], b[1]-a[1])
		out = append(out, [2]float64{a[0] + f*(b[0]-a[0]), a[1] + f*(b[1]-a[1])})
	}
	return out
}

type openCase struct {
	name  string
	chain refChain
	// clear is a region on the chain's right, kept 0.25 mm off it, that
	// certify should prove empty: [minX, minY, maxX, maxY].
	clear [4]float64
}

func openCases() []openCase {
	return []openCase{
		// Straight edges turning left and right.
		{
			name:  "zigzag",
			chain: refChain{points: [][2]float64{{0, 0}, {20, 0}, {30, 15}, {45, 5}, {60, 10}}},
			clear: [4]float64{-10, -15, 70, -0.25},
		},
		// A line, a quarter arc left, a line and a quarter arc right, every
		// joint smooth: the chain ends heading the way it began.
		{
			name:  "S of arcs",
			chain: refChain{points: [][2]float64{{0, 0}, {20, 0}, {30, 10}, {30, 25}, {40, 35}}, curvatures: []float64{0, 0.1, 0, -0.1}},
			clear: [4]float64{32, -10, 50, 30},
		},
		// An arc first, so the start ray is the arc's tangent, then a sharp
		// right turn.
		{
			name:  "arc first",
			chain: refChain{points: [][2]float64{{0, 0}, {10, 10}, {25, 10}}, curvatures: []float64{0.1, 0}},
			clear: [4]float64{-10, -10, 30, -0.25},
		},
	}
}

func (c openCase) doc() string {
	return "@scene(120)\nsolid = " + c.chain.gsl() + ";\nreturn solid;\n"
}

// endNormals are points along the normal lines at the chain's two ends, where
// a ray meets the chain's first or last edge: a tie between them would show.
func (c openCase) endNormals() [][2]float64 {
	n := len(c.chain.points)
	t0, _ := edgeTangents(c.chain.points[0], c.chain.points[1], c.chain.curvature(0))
	_, te := edgeTangents(c.chain.points[n-2], c.chain.points[n-1], c.chain.curvature(n-2))
	var out [][2]float64
	for _, end := range []struct{ at, t [2]float64 }{{c.chain.points[0], t0}, {c.chain.points[n-1], te}} {
		for _, d := range []float64{-6, -3, -1, -0.25, 0.25, 1, 3, 6} {
			for _, nudge := range []float64{-1e-3, 0, 1e-3} {
				out = append(out, [2]float64{
					end.at[0] + d*end.t[1] + nudge*end.t[0],
					end.at[1] - d*end.t[0] + nudge*end.t[1],
				})
			}
		}
	}
	return out
}

// openProfileXY is exact against the reference on a grid round the chain,
// into the rays, and along the normals at its ends.
func TestOpenProfileXY_MatchesReference(t *testing.T) {
	for _, c := range openCases() {
		t.Run(c.name, func(t *testing.T) {
			ref := newRefOpen(t, c.chain)
			points := flat(append(gridAround(ref.chainBox(), 10, 48), c.endNormals()...))
			compareToReference(t, c.name, points, evalProfileField(t, c.doc(), "solid", points), func(q [3]float64) float64 {
				return ref.distance([2]float64{q[0], q[1]})
			})
		})
	}
}

// A piece of a closed profile, taken as an open chain, has the closed
// profile's distance and sign beside each of its edges: positive on the right
// of travel is the counterclockwise loop's outside.
func TestOpenProfileXY_PieceOfAClosedProfile(t *testing.T) {
	loop := arcL()
	piece := refChain{points: loop.points[1:6], curvatures: loop.curvatures[1:5]}
	var points [][2]float64
	for i := 0; i+1 < len(piece.points); i++ {
		a, b := piece.points[i], piece.points[i+1]
		k := piece.curvature(i)
		tangent, _ := edgeTangents(a, b, k)
		mid := [2]float64{(a[0] + b[0]) / 2, (a[1] + b[1]) / 2}
		if k != 0 {
			// The arc's middle sample, and its tangent there: the start
			// tangent turned by the arc's turn so far.
			samples := arcSamples(t, a, b, k)
			m := len(samples) / 2
			arc := 2 * math.Asin(k*math.Hypot(b[0]-a[0], b[1]-a[1])/2) / k * float64(m) / float64(len(samples))
			mid = samples[m]
			c, s := math.Cos(k*arc), math.Sin(k*arc)
			tangent = [2]float64{tangent[0]*c - tangent[1]*s, tangent[0]*s + tangent[1]*c}
		}
		// Across the edge, right of travel and left.
		for _, d := range []float64{-1, -0.5, 0.5, 1} {
			points = append(points, [2]float64{mid[0] + d*tangent[1], mid[1] - d*tangent[0]})
		}
	}
	q := flat(points)
	open := evalProfileField(t, "@scene(120)\nsolid = "+piece.gsl()+";\nreturn solid;\n", "solid", q)
	closed := evalProfileField(t, "@scene(120)\nsolid = "+loop.gsl()+";\nreturn solid;\n", "solid", q)
	for i := range q {
		if math.Abs(open[i]-closed[i]) > profileTolerance {
			t.Errorf("beside the piece at %v: open %v, closed %v", q[i], open[i], closed[i])
		}
	}
}

// Its .u is the distance along the chain from its first point to the nearest
// point of the extended chain: negative along the start ray, past the chain's
// length along the end ray. Checked a little off the chain on both sides, and
// on the rays, where the nearest point is unambiguous.
func TestOpenProfileXY_Frame(t *testing.T) {
	for _, c := range openCases() {
		t.Run(c.name, func(t *testing.T) {
			ref := newRefOpen(t, c.chain)
			var points [][3]float64
			for _, s := range ref.samples(1.3, 8, 8) {
				for _, d := range []float64{-0.4, 0.4} {
					points = append(points, [3]float64{s[0], s[1] + d, 0}, [3]float64{s[0] + d, s[1], 0})
				}
			}
			e := New(&source.MockManager{}, nil)
			groups, err := e.EvalGsl(context.Background(), nil, c.doc(), "", []string{"solid.u"}, points)
			if err != nil {
				t.Fatalf("eval: %v", err)
			}
			checked, worst := 0, 0.0
			for i, q := range points {
				d, want := ref.nearest([2]float64{q[0], q[1]})
				// Skip points whose nearest point is not unique to within the
				// offset: a corner's normal cone or a bend's centre.
				if !uniqueNearest(ref, [2]float64{q[0], q[1]}, d) {
					continue
				}
				checked++
				got := float64(groups[0].Samples[0][i].Value)
				worst = math.Max(worst, math.Abs(got-want))
				if math.Abs(got-want) > 1e-3 {
					t.Errorf("solid.u at %v = %v, want %v", q, got, want)
				}
			}
			t.Logf("%s: .u at %d points, worst difference %.2g mm", c.name, checked, worst)
		})
	}
}

// uniqueNearest reports whether every point of the extended chain within 1e-3
// mm of q's distance d lies within 0.05 mm of the chain's arc length of the
// nearest one: a point with two separated nearest points has no single .u.
func uniqueNearest(r refOpen, q [2]float64, d float64) bool {
	lo, hi := math.Inf(1), math.Inf(-1)
	for i, s := range r.segments {
		a, b := s[0], s[1]
		ex, ey := b[0]-a[0], b[1]-a[1]
		length := math.Hypot(ex, ey)
		wx, wy := q[0]-a[0], q[1]-a[1]
		h := math.Max(0, math.Min(1, (wx*ex+wy*ey)/(length*length)))
		if math.Hypot(wx-h*ex, wy-h*ey) <= d+1e-3 {
			station := r.stations[i] + h*length
			lo, hi = math.Min(lo, station), math.Max(hi, station)
		}
	}
	return hi-lo < 0.05
}

// Cells straddling the extended chain, the rays included, bound it soundly and
// tightly, and certify proves a region on its right empty.
func TestOpenProfileXY_Proves(t *testing.T) {
	for _, c := range openCases() {
		t.Run(c.name, func(t *testing.T) {
			ref := newRefOpen(t, c.chain)
			checkSurfaceCells(t, c.name, c.doc(), flat(ref.samples(1, 8, 8)), func(q [3]float64) float64 {
				return ref.distance([2]float64{q[0], q[1]})
			})
			e := New(&source.MockManager{}, nil)
			r, err := e.Certify(context.Background(), nil, c.doc(), "", "solid",
				[6]float64{c.clear[0], c.clear[1], -1, c.clear[2], c.clear[3], 1}, 0, 0)
			if err != nil {
				t.Fatalf("certify: %v", err)
			}
			if !r.Proven {
				t.Errorf("right of the chain: want proven, got proven=%v disproven=%v indeterminate=%v [%v, %v] in %d evals",
					r.Proven, r.Disproven, r.Indeterminate, r.RegionLower, r.RegionUpper, r.Evals)
			}
			t.Logf("%s: certify proves the right side in %d evals", c.name, r.Evals)
		})
	}
}
