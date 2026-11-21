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

// The profiles' ground truth: a brute-force reference that knows nothing of
// how profileXY computes its distance. Each loop's boundary is sampled into a
// polygon (a line as one segment, an arc as chords refStep apart), the
// distance is the least distance to any segment and the sign comes from
// containment, by counting crossings over every loop together, so properly
// nested loops are a solid with holes. A chord lies within refStep^2 * |k| / 8
// of its arc: under 1e-6 mm for the curvatures here.
const refStep = 0.002

// profileTolerance is how far the field may sit from the reference, in mm. The
// field is evaluated in float32; at the tens of mm these points lie at, one
// ulp is about 2e-6.
const profileTolerance = 2e-5

// refLoop is one closed loop as profileXY takes it: points, and one signed
// curvature per edge (nil for a polygon).
type refLoop struct {
	points     [][2]float64
	curvatures []float64
}

// gsl is the loop as a profileXY call.
func (l refLoop) gsl() string {
	pts := make([]string, len(l.points))
	for i, pt := range l.points {
		pts[i] = fmt.Sprintf("[%v, %v]", pt[0], pt[1])
	}
	if l.curvatures == nil {
		return fmt.Sprintf("profileXY([%s])", strings.Join(pts, ", "))
	}
	ks := make([]string, len(l.curvatures))
	for i, k := range l.curvatures {
		if k < 0 {
			ks[i] = fmt.Sprintf("neg(%v)", -k)
		} else {
			ks[i] = fmt.Sprintf("%v", k)
		}
	}
	return fmt.Sprintf("profileXY([%s], [%s])", strings.Join(pts, ", "), strings.Join(ks, ", "))
}

// polygon samples the loop's boundary: every point of its edges at most
// refStep apart along an arc, each line as its two ends.
func (l refLoop) polygon(t *testing.T) [][2]float64 {
	var out [][2]float64
	for i, a := range l.points {
		b := l.points[(i+1)%len(l.points)]
		k := 0.0
		if l.curvatures != nil {
			k = l.curvatures[i]
		}
		out = append(out, arcSamples(t, a, b, k)...)
	}
	return out
}

// arcSamples walks the edge from a toward b, b excluded: one point for a line,
// and for an arc of curvature k (positive turning left) points refStep apart in
// arc length. The arc leaves a along the chord turned by -asin(k * chord / 2).
func arcSamples(t *testing.T, a, b [2]float64, k float64) [][2]float64 {
	if k == 0 {
		return [][2]float64{a}
	}
	dx, dy := b[0]-a[0], b[1]-a[1]
	chord := math.Hypot(dx, dy)
	half := math.Asin(k * chord / 2)
	c, s := math.Cos(-half), math.Sin(-half)
	tx, ty := (dx*c-dy*s)/chord, (dx*s+dy*c)/chord
	nx, ny := -ty, tx
	length := 2 * half / k
	n := int(math.Ceil(length / refStep))
	at := func(arc float64) [2]float64 {
		along, across := math.Sin(k*arc)/k, (1-math.Cos(k*arc))/k
		return [2]float64{a[0] + along*tx + across*nx, a[1] + along*ty + across*ny}
	}
	if end := at(length); math.Hypot(end[0]-b[0], end[1]-b[1]) > 1e-9 {
		t.Fatalf("arc from %v with curvature %v ends at %v, not %v", a, k, end, b)
	}
	out := make([][2]float64, n)
	for i := range out {
		out[i] = at(length * float64(i) / float64(n))
	}
	return out
}

// refProfile is the sampled boundary of one or more loops.
type refProfile struct {
	segments [][2][2]float64
}

func newRefProfile(t *testing.T, loops ...refLoop) refProfile {
	var r refProfile
	for _, l := range loops {
		poly := l.polygon(t)
		for i, a := range poly {
			r.segments = append(r.segments, [2][2]float64{a, poly[(i+1)%len(poly)]})
		}
	}
	return r
}

// distance is the signed distance from q to the sampled boundary, negative
// inside.
func (r refProfile) distance(q [2]float64) float64 {
	best := math.Inf(1)
	inside := false
	for _, s := range r.segments {
		a, b := s[0], s[1]
		ex, ey := b[0]-a[0], b[1]-a[1]
		wx, wy := q[0]-a[0], q[1]-a[1]
		h := math.Max(0, math.Min(1, (wx*ex+wy*ey)/(ex*ex+ey*ey)))
		best = math.Min(best, math.Hypot(wx-h*ex, wy-h*ey))
		if (a[1] > q[1]) != (b[1] > q[1]) && q[0] < a[0]+(q[1]-a[1])*ex/ey {
			inside = !inside
		}
	}
	if inside {
		return -best
	}
	return best
}

// box is the sampled boundary's [minX, minY, maxX, maxY].
func (r refProfile) box() [4]float64 {
	b := [4]float64{math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)}
	for _, s := range r.segments {
		b[0], b[1] = math.Min(b[0], s[0][0]), math.Min(b[1], s[0][1])
		b[2], b[3] = math.Max(b[2], s[0][0]), math.Max(b[3], s[0][1])
	}
	return b
}

// boundarySamples are points on the sampled boundary about spacing apart.
func (r refProfile) boundarySamples(spacing float64) [][2]float64 {
	var out [][2]float64
	carried := spacing
	for _, s := range r.segments {
		a, b := s[0], s[1]
		length := math.Hypot(b[0]-a[0], b[1]-a[1])
		for at := spacing - carried; at < length; at += spacing {
			f := at / length
			out = append(out, [2]float64{a[0] + f*(b[0]-a[0]), a[1] + f*(b[1]-a[1])})
		}
		carried = math.Mod(carried+length, spacing)
	}
	return out
}

// profileCase is one profile: its GSL expression and its reference.
type profileCase struct {
	name string
	expr string
	ref  refProfile
	// extra holds points worth checking beyond the grid: along a joint's
	// normal, where a tie between its edges would show.
	extra [][2]float64
}

func lShape() refLoop {
	return refLoop{points: [][2]float64{{0, 0}, {30, 0}, {30, 10}, {10, 10}, {10, 30}, {0, 30}}}
}

func star() refLoop {
	var pts [][2]float64
	for i := 0; i < 10; i++ {
		radius := 20.0
		if i%2 == 1 {
			radius = 8
		}
		angle := math.Pi / 2 * float64(i) / 2.5
		pts = append(pts, [2]float64{radius * math.Cos(angle), radius * math.Sin(angle)})
	}
	return refLoop{points: pts}
}

// arcL is the L of examples/profile_cylinder.gsl at its default sliders: arms
// 40 long and 14 wide, a fillet of 6 in the inside corner and a round of 5 on
// the heel.
func arcL() refLoop {
	return refLoop{
		points:     [][2]float64{{5, 0}, {40, 0}, {40, 14}, {20, 14}, {14, 20}, {14, 40}, {0, 40}, {0, 5}},
		curvatures: []float64{0, 0, 0, -1.0 / 6, 0, 0, 0, 1.0 / 5},
	}
}

// slot is a stadium with a collinear point on its lower side and quarter arcs
// for ends: every joint but none is smooth, the collinear point and the arcs'
// joints with each other and with the sides.
func slot() refLoop {
	return refLoop{
		points:     [][2]float64{{0, -5}, {10, -5}, {20, -5}, {25, 0}, {20, 5}, {0, 5}, {-5, 0}},
		curvatures: []float64{0, 0, 0.2, 0.2, 0, 0.2, 0.2},
	}
}

func squareOutline() refLoop {
	return refLoop{points: [][2]float64{{-20, -20}, {20, -20}, {20, 20}, {-20, 20}}}
}

// squareHole winds clockwise: profileXY gives its complement.
func squareHole() refLoop {
	return refLoop{points: [][2]float64{{-8, -8}, {-8, 8}, {8, 8}, {8, -8}}}
}

func profileCases(t *testing.T) []profileCase {
	var alongJoints [][2]float64
	for y := -12.0; y <= 12; y += 0.5 {
		for _, x := range []float64{0, 10, 20} {
			alongJoints = append(alongJoints, [2]float64{x, y})
		}
	}
	for _, d := range []float64{-4, -2, -0.5, 0.5, 2, 4, 8} {
		for _, at := range [][2]float64{{20, 0}, {0, 0}} {
			for _, dir := range [][2]float64{{1, -1}, {1, 1}} {
				s := d / math.Sqrt2
				if at[0] == 0 {
					s = -s
				}
				alongJoints = append(alongJoints, [2]float64{at[0] + s*dir[0], at[1] + s*dir[1]})
			}
		}
	}
	return []profileCase{
		{name: "L", expr: lShape().gsl(), ref: newRefProfile(t, lShape())},
		{name: "star", expr: star().gsl(), ref: newRefProfile(t, star())},
		{
			name: "square with a hole",
			expr: fmt.Sprintf("max(%s, %s)", squareOutline().gsl(), squareHole().gsl()),
			ref:  newRefProfile(t, squareOutline(), squareHole()),
		},
		{name: "arc L", expr: arcL().gsl(), ref: newRefProfile(t, arcL())},
		{name: "slot", expr: slot().gsl(), ref: newRefProfile(t, slot()), extra: alongJoints},
	}
}

// gridAround is an n by n grid over the box grown by margin, its spacing
// irrational to the profiles' coordinates so few points land on a tie.
func gridAround(box [4]float64, margin float64, n int) [][2]float64 {
	var out [][2]float64
	x0, y0 := box[0]-margin, box[1]-margin
	dx, dy := (box[2]-box[0]+2*margin)/float64(n-1), (box[3]-box[1]+2*margin)/float64(n-1)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			out = append(out, [2]float64{x0 + (float64(i)+0.1*math.Sqrt2)*dx, y0 + (float64(j)+0.1*math.Pi)*dy})
		}
	}
	return out
}

// evalProfileField evaluates the named field of doc at points.
func evalProfileField(t *testing.T, doc, name string, points [][3]float64) []float64 {
	t.Helper()
	e := New(&source.MockManager{}, nil)
	groups, err := e.EvalGsl(context.Background(), nil, doc, "", []string{name}, points)
	if err != nil {
		t.Fatalf("eval %s: %v", name, err)
	}
	out := make([]float64, len(points))
	for i := range points {
		out[i] = float64(groups[0].Samples[0][i].Value)
	}
	return out
}

// compareToReference fails at the worst point if any is off by more than
// profileTolerance, and logs the worst difference either way.
func compareToReference(t *testing.T, label string, points [][3]float64, got []float64, want func([3]float64) float64) {
	t.Helper()
	worst, worstAt := 0.0, -1
	for i, q := range points {
		if diff := math.Abs(got[i] - want(q)); diff > worst {
			worst, worstAt = diff, i
		}
	}
	t.Logf("%s: worst difference %.2g mm over %d points", label, worst, len(points))
	if worst > profileTolerance {
		q := points[worstAt]
		t.Errorf("%s at %v = %v, reference %v (off by %.3g)", label, q, got[worstAt], want(q), worst)
	}
}

func flat(points [][2]float64) [][3]float64 {
	out := make([][3]float64, len(points))
	for i, q := range points {
		out[i] = [3]float64{q[0], q[1], 0}
	}
	return out
}

// Today's profileXY, polygon and arc forms, is exact: it agrees with the
// reference on a grid over each profile and, for the slot, along the normals
// of its smooth joints.
func TestProfileXY_MatchesReference(t *testing.T) {
	for _, c := range profileCases(t) {
		t.Run(c.name, func(t *testing.T) {
			points := flat(append(gridAround(c.ref.box(), 6, 48), c.extra...))
			got := evalProfileField(t, "@scene(120)\nsolid = "+c.expr+";\nreturn solid;\n", "solid", points)
			compareToReference(t, c.name, points, got, func(q [3]float64) float64 {
				return c.ref.distance([2]float64{q[0], q[1]})
			})
		})
	}
}

// grid3 is an n^3 grid over the box grown by margin, offset as gridAround's.
func grid3(lo, hi [3]float64, margin float64, n int) [][3]float64 {
	var out [][3]float64
	var step [3]float64
	for k := range step {
		step[k] = (hi[k] - lo[k] + 2*margin) / float64(n-1)
	}
	offsets := [3]float64{0.1 * math.Sqrt2, 0.1 * math.Pi, 0.1 * math.E}
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			for l := 0; l < n; l++ {
				idx := [3]int{i, j, l}
				var q [3]float64
				for k := range q {
					q[k] = lo[k] - margin + (float64(idx[k])+offsets[k])*step[k]
				}
				out = append(out, q)
			}
		}
	}
	return out
}

// extrudedReference is the exact distance to the profile kept between two
// planes with square rims: the profile's distance and the slab's combined as a
// box's are.
func extrudedReference(ref refProfile, low, high float64) func([3]float64) float64 {
	return func(q [3]float64) float64 {
		d := ref.distance([2]float64{q[0], q[1]})
		h := math.Abs(q[2]-(low+high)/2) - (high-low)/2
		outside := math.Hypot(math.Max(d, 0), math.Max(h, 0))
		return outside + math.Min(math.Max(d, h), 0)
	}
}

// extrude of an exact profile, with sharp rims, is the exact extruded solid.
func TestExtrude_MatchesReference(t *testing.T) {
	const low, high = 0.0, 12.0
	for _, c := range profileCases(t) {
		t.Run(c.name, func(t *testing.T) {
			box := c.ref.box()
			points := grid3([3]float64{box[0], box[1], low}, [3]float64{box[2], box[3], high}, 5, 22)
			doc := fmt.Sprintf("@scene(120)\nsolid = extrude(%s, %v, %v);\nreturn solid;\n", c.expr, low, high)
			compareToReference(t, c.name, points, evalProfileField(t, doc, "solid", points), extrudedReference(c.ref, low, high))
		})
	}
}

// revolve of an exact profile lying to the right of its axis is the exact
// solid of revolution: the profile's distance read at the point's distance from
// the axis and its distance along it, laid back onto the sketch.
func TestRevolve_MatchesReference(t *testing.T) {
	axes := map[string][2][2]float64{
		"L":                  {{-10, 0}, {0, 1}},
		"star":               {{-25, 0}, {0, 1}},
		"square with a hole": {{-30, 0}, {0, 1}},
		"arc L":              {{-10, 0}, {0, 1}},
		"slot":               {{0, 10}, {1, 0}},
	}
	for _, c := range profileCases(t) {
		t.Run(c.name, func(t *testing.T) {
			axis := axes[c.name]
			at, dir := axis[0], axis[1]
			norm := math.Hypot(dir[0], dir[1])
			ux, uy := dir[0]/norm, dir[1]/norm
			// The solid's reach: the farthest boundary point from the axis.
			box := c.ref.box()
			reach := 0.0
			for _, corner := range [][2]float64{{box[0], box[1]}, {box[2], box[1]}, {box[0], box[3]}, {box[2], box[3]}} {
				wx, wy := corner[0]-at[0], corner[1]-at[1]
				reach = math.Max(reach, math.Abs(wx*uy-wy*ux))
			}
			lo := [3]float64{math.Min(box[0], at[0]-reach), math.Min(box[1], at[1]-reach), -reach}
			hi := [3]float64{math.Max(box[2], at[0]+reach), math.Max(box[3], at[1]+reach), reach}
			points := grid3(lo, hi, 4, 22)
			doc := fmt.Sprintf("@scene(200)\nsolid = revolve(%s, [%v, %v], [%v, %v]);\nreturn solid;\n",
				c.expr, at[0], at[1], dir[0], dir[1])
			compareToReference(t, c.name, points, evalProfileField(t, doc, "solid", points), func(q [3]float64) float64 {
				wx, wy := q[0]-at[0], q[1]-at[1]
				along := wx*ux + wy*uy
				radial := math.Hypot(wx*uy-wy*ux, q[2])
				return c.ref.distance([2]float64{at[0] + along*ux + radial*uy, at[1] + along*uy - radial*ux})
			})
		})
	}
}
