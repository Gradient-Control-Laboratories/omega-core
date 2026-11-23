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

// profileEdgesXY(points, edges), openProfileEdgesXY(points, edges) and
// sweepOpenEdges(section, points, edges): profiles whose edges are each a
// line, a circular arc or a quadratic Bezier, named by edgeLine(),
// edgeArc(curvature) and edgeBezier(control).

const (
	lineEdge = iota
	arcEdge
	bezierEdge
)

// refEdge is one edge as a descriptor names it.
type refEdge struct {
	kind      int
	curvature float64
	control   [2]float64
}

func edgeL() refEdge             { return refEdge{kind: lineEdge} }
func edgeA(k float64) refEdge    { return refEdge{kind: arcEdge, curvature: k} }
func edgeB(x, y float64) refEdge { return refEdge{kind: bezierEdge, control: [2]float64{x, y}} }
func gslNumber(v float64) string {
	if v < 0 {
		return fmt.Sprintf("neg(%v)", -v)
	}
	return fmt.Sprintf("%v", v)
}

func (e refEdge) gsl() string {
	switch e.kind {
	case arcEdge:
		return fmt.Sprintf("edgeArc(%s)", gslNumber(e.curvature))
	case bezierEdge:
		return fmt.Sprintf("edgeBezier([%s, %s])", gslNumber(e.control[0]), gslNumber(e.control[1]))
	}
	return "edgeLine()"
}

func edgesGSL(edges []refEdge) string {
	out := make([]string, len(edges))
	for i, e := range edges {
		out[i] = e.gsl()
	}
	return "[" + strings.Join(out, ", ") + "]"
}

// samples walks the edge from a toward b, b excluded, refStep apart or closer
// along a curve: a line as its start, an arc by arcSamples, a Bezier at even
// steps of its parameter, as many as its control polygon's length in refSteps.
func (e refEdge) samples(t *testing.T, a, b [2]float64) [][2]float64 {
	switch e.kind {
	case arcEdge:
		return arcSamples(t, a, b, e.curvature)
	case bezierEdge:
		c := e.control
		n := int(math.Ceil((math.Hypot(c[0]-a[0], c[1]-a[1]) + math.Hypot(b[0]-c[0], b[1]-c[1])) / refStep))
		out := make([][2]float64, n)
		for i := range out {
			s := float64(i) / float64(n)
			for k := 0; k < 2; k++ {
				out[i][k] = (1-s)*(1-s)*a[k] + 2*(1-s)*s*c[k] + s*s*b[k]
			}
		}
		return out
	}
	return [][2]float64{a}
}

// tangents are the unit tangents at the edge's two ends.
func (e refEdge) tangents(a, b [2]float64) (start, end [2]float64) {
	unit := func(x, y float64) [2]float64 { l := math.Hypot(x, y); return [2]float64{x / l, y / l} }
	switch e.kind {
	case arcEdge:
		return edgeTangents(a, b, e.curvature)
	case bezierEdge:
		c := e.control
		return unit(c[0]-a[0], c[1]-a[1]), unit(b[0]-c[0], b[1]-c[1])
	}
	d := unit(b[0]-a[0], b[1]-a[1])
	return d, d
}

// edgeLoop is a closed profile of mixed edges, edge i leaving point i.
type edgeLoop struct {
	points [][2]float64
	edges  []refEdge
}

func (l edgeLoop) gsl() string {
	return fmt.Sprintf("profileEdgesXY(%s, %s)", pointsGSL(l.points), edgesGSL(l.edges))
}

// reference is the loop's sampled boundary, and the distance along the loop
// at the start of each of its segments.
func (l edgeLoop) reference(t *testing.T) (refProfile, []float64) {
	var poly [][2]float64
	for i, a := range l.points {
		poly = append(poly, l.edges[i].samples(t, a, l.points[(i+1)%len(l.points)])...)
	}
	var r refProfile
	var stations []float64
	run := 0.0
	for i, a := range poly {
		b := poly[(i+1)%len(poly)]
		r.segments = append(r.segments, [2][2]float64{a, b})
		stations = append(stations, run)
		run += math.Hypot(b[0]-a[0], b[1]-a[1])
	}
	return r, stations
}

type edgeLoopCase struct {
	name string
	loop edgeLoop
	// clear is a region outside the loop, 0.25 mm off it, for certify.
	clear [4]float64
	// curved says the loop has a Bezier off its chord, whose surface cells are
	// allowed bezierCellSpread.
	curved bool
}

// bezierCellSpread is how wide a surface cell's interval may be, in cell
// diagonals, on a loop with a curved Bezier edge: the closed-form root's
// interval is loose right at the curve. Measured 2026-11-18: 9.18 on the tight
// Bezier, 5.93 on the mixed loop, 2.41 on the lens (the owner set 10); lines and
// arcs keep maxCellSpread.
const bezierCellSpread = 10.0

func edgeLoopCases() []edgeLoopCase {
	return []edgeLoopCase{
		// Two Bezier edges, out below and out above: a lens with sharp ends.
		{
			name:   "Bezier lens",
			loop:   edgeLoop{points: [][2]float64{{0, 0}, {30, 0}}, edges: []refEdge{edgeB(15, -12), edgeB(15, 12)}},
			clear:  [4]float64{-5, 6.25, 35, 15},
			curved: true,
		},
		// A line, a Bezier leaving it tangentially and arriving tangent to the
		// next line, a Bezier dented in between sharp corners, an arc, a line.
		{
			name: "mixed",
			loop: edgeLoop{
				points: [][2]float64{{0, 0}, {30, 0}, {40, 15}, {40, 30}, {20, 30}, {0, 30}},
				edges:  []refEdge{edgeL(), edgeB(40, 0), edgeL(), edgeB(30, 20), edgeA(0.05), edgeL()},
			},
			clear:  [4]float64{40.25, -5, 50, 35},
			curved: true,
		},
		// examples/profile_bezier.gsl's plate at its default sliders: a
		// Bezier whose control polygon turns sharply at the control, so the
		// parabola it lies on swings back past it into the plate, and a dent.
		{
			name: "tight Bezier",
			loop: edgeLoop{
				points: [][2]float64{{0, 0}, {34, 0}, {44, 15}, {44, 30}, {16, 30}, {0, 30}},
				edges:  []refEdge{edgeL(), edgeB(44, 12), edgeL(), edgeB(22, 22), edgeA(0.04), edgeL()},
			},
			clear:  [4]float64{44.25, -5, 55, 35},
			curved: true,
		},
		// Beziers whose controls lie on their chords, at the middle and a
		// quarter of the way: straight edges, the square they draw.
		{
			name:  "straight Beziers",
			loop:  edgeLoop{points: [][2]float64{{0, 0}, {20, 0}, {20, 20}, {0, 20}}, edges: []refEdge{edgeB(10, 0), edgeB(20, 5), edgeL(), edgeL()}},
			clear: [4]float64{20.25, -5, 30, 25},
		},
	}
}

// profileEdgesXY is exact against the reference on a grid round each loop.
func TestProfileEdgesXY_MatchesReference(t *testing.T) {
	for _, c := range edgeLoopCases() {
		t.Run(c.name, func(t *testing.T) {
			ref, _ := c.loop.reference(t)
			points := flat(gridAround(ref.box(), 6, 48))
			got := evalProfileField(t, "@scene(120)\nsolid = "+c.loop.gsl()+";\nreturn solid;\n", "solid", points)
			compareToReference(t, c.name, points, got, func(q [3]float64) float64 { return ref.distance([2]float64{q[0], q[1]}) })
		})
	}
}

// With lines and arcs alone it is profileXY(points, curvatures).
func TestProfileEdgesXY_LinesAndArcsAreProfileXY(t *testing.T) {
	for _, loop := range []refLoop{arcL(), slot(), lShape()} {
		edges := make([]refEdge, len(loop.points))
		for i := range edges {
			if loop.curvatures != nil && loop.curvatures[i] != 0 {
				edges[i] = edgeA(loop.curvatures[i])
			} else {
				edges[i] = edgeL()
			}
		}
		ref := newRefProfile(t, loop)
		points := flat(gridAround(ref.box(), 6, 32))
		doc := fmt.Sprintf("@scene(120)\nwant = %s;\nsolid = profileEdgesXY(%s, %s);\nreturn solid;\n", loop.gsl(), pointsGSL(loop.points), edgesGSL(edges))
		got := evalProfileField(t, doc, "solid", points)
		want := evalProfileField(t, doc, "want", points)
		for i, q := range points {
			if math.Abs(got[i]-want[i]) > profileTolerance {
				t.Fatalf("%s at %v: profileEdgesXY %v, profileXY %v", loop.gsl(), q, got[i], want[i])
			}
		}
	}
}

// Its .u is the arc length along the loop from its first point to the nearest
// point, checked a little off the boundary where that point is unambiguous.
func TestProfileEdgesXY_Frame(t *testing.T) {
	for _, c := range edgeLoopCases() {
		t.Run(c.name, func(t *testing.T) {
			ref, stations := c.loop.reference(t)
			var points [][3]float64
			for _, s := range ref.boundarySamples(1.3) {
				for _, d := range []float64{-0.4, 0.4} {
					points = append(points, [3]float64{s[0], s[1] + d, 0}, [3]float64{s[0] + d, s[1], 0})
				}
			}
			e := New(&source.MockManager{}, nil)
			groups, err := e.EvalGsl(context.Background(), nil, "@scene(120)\nsolid = "+c.loop.gsl()+";\nreturn solid;\n", "", []string{"solid.u"}, points)
			if err != nil {
				t.Fatalf("eval: %v", err)
			}
			checked, worst := 0, 0.0
			for i, q := range points {
				want, unique := loopStation(ref, stations, [2]float64{q[0], q[1]})
				if !unique {
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

// loopStation is the distance along the loop to q's nearest point, and whether
// every point within 1e-3 mm of the nearest distance lies within 0.05 mm of it
// along the loop (the first point's station counted as both 0 and the length).
func loopStation(r refProfile, stations []float64, q [2]float64) (float64, bool) {
	best, at := math.Inf(1), 0.0
	type near struct{ d, s float64 }
	var all []near
	for i, seg := range r.segments {
		a, b := seg[0], seg[1]
		ex, ey := b[0]-a[0], b[1]-a[1]
		l := math.Hypot(ex, ey)
		h := math.Max(0, math.Min(1, ((q[0]-a[0])*ex+(q[1]-a[1])*ey)/(l*l)))
		d := math.Hypot(q[0]-a[0]-h*ex, q[1]-a[1]-h*ey)
		all = append(all, near{d, stations[i] + h*l})
		if d < best {
			best, at = d, stations[i]+h*l
		}
	}
	total := stations[len(stations)-1] + math.Hypot(r.segments[len(r.segments)-1][1][0]-r.segments[len(r.segments)-1][0][0],
		r.segments[len(r.segments)-1][1][1]-r.segments[len(r.segments)-1][0][1])
	for _, n := range all {
		if n.d <= best+1e-3 {
			gap := math.Abs(n.s - at)
			gap = math.Min(gap, total-gap)
			if gap > 0.05 {
				return at, false
			}
		}
	}
	if total-at < 0.05 || at < 0.05 {
		return at, false
	}
	return at, true
}

// Cells straddling each loop bound it soundly and tightly, and certify proves
// a region outside it empty.
func TestProfileEdgesXY_Proves(t *testing.T) {
	for _, c := range edgeLoopCases() {
		t.Run(c.name, func(t *testing.T) {
			ref, _ := c.loop.reference(t)
			doc := "@scene(120)\nsolid = " + c.loop.gsl() + ";\nreturn solid;\n"
			limit := maxCellSpread
			if c.curved {
				limit = bezierCellSpread
			}
			checkSurfaceCellsWithin(t, c.name, doc, flat(ref.boundarySamples(1)), limit, func(q [3]float64) float64 {
				return ref.distance([2]float64{q[0], q[1]})
			})
			e := New(&source.MockManager{}, nil)
			r, err := e.Certify(context.Background(), nil, doc, "", "solid",
				[6]float64{c.clear[0], c.clear[1], -1, c.clear[2], c.clear[3], 1}, 0, 0)
			if err != nil {
				t.Fatalf("certify: %v", err)
			}
			if !r.Proven {
				t.Errorf("outside: want proven, got proven=%v disproven=%v [%v, %v] in %d evals",
					r.Proven, r.Disproven, r.RegionLower, r.RegionUpper, r.Evals)
			}
			t.Logf("%s: certify proves the region outside in %d evals", c.name, r.Evals)
		})
	}
}

// edgeChain is an open chain of mixed edges, one fewer than the points.
type edgeChain struct {
	points [][2]float64
	edges  []refEdge
}

func (c edgeChain) gsl() string {
	return fmt.Sprintf("openProfileEdgesXY(%s, %s)", pointsGSL(c.points), edgesGSL(c.edges))
}

func (c edgeChain) ends() (first, firstTangent, last, lastTangent [2]float64) {
	n := len(c.points)
	firstTangent, _ = c.edges[0].tangents(c.points[0], c.points[1])
	_, lastTangent = c.edges[n-2].tangents(c.points[n-2], c.points[n-1])
	return c.points[0], firstTangent, c.points[n-1], lastTangent
}

// reference is the extended chain sampled, as newRefOpen samples a refChain.
func (c edgeChain) reference(t *testing.T) refOpen {
	first, t0, last, te := c.ends()
	farStart := [2]float64{first[0] - openRayReach*t0[0], first[1] - openRayReach*t0[1]}
	farEnd := [2]float64{last[0] + openRayReach*te[0], last[1] + openRayReach*te[1]}
	line := [][2]float64{farStart}
	for i := 0; i+1 < len(c.points); i++ {
		line = append(line, c.edges[i].samples(t, c.points[i], c.points[i+1])...)
	}
	line = append(line, last, farEnd)
	var r refOpen
	station := -openRayReach
	for i := 0; i+1 < len(line); i++ {
		r.segments = append(r.segments, [2][2]float64{line[i], line[i+1]})
		r.stations = append(r.stations, station)
		station += math.Hypot(line[i+1][0]-line[i][0], line[i+1][1]-line[i][1])
	}
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

func edgeChainCases() []edgeChain {
	return []edgeChain{
		// A line, a Bezier bend left, a line and a Bezier bend right, smooth.
		{points: [][2]float64{{0, 0}, {20, 0}, {30, 10}, {30, 25}, {40, 35}},
			edges: []refEdge{edgeL(), edgeB(30, 0), edgeL(), edgeB(30, 35)}},
		// A Bezier first, so the start ray is its tangent, then a sharp turn
		// right and an arc.
		{points: [][2]float64{{0, 0}, {10, 10}, {25, 10}, {35, 0}},
			edges: []refEdge{edgeB(10, 0), edgeL(), edgeA(-0.08)}},
	}
}

// openProfileEdgesXY is exact against the extended chain's reference.
func TestOpenProfileEdgesXY_MatchesReference(t *testing.T) {
	for i, c := range edgeChainCases() {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			ref := newRefOpenFrom(t, c)
			points := flat(gridAround(ref.chainBox(), 10, 40))
			got := evalProfileField(t, "@scene(120)\nsolid = "+c.gsl()+";\nreturn solid;\n", "solid", points)
			compareToReference(t, c.gsl(), points, got, func(q [3]float64) float64 { return ref.distance([2]float64{q[0], q[1]}) })
		})
	}
}

func newRefOpenFrom(t *testing.T, c edgeChain) refOpen { return c.reference(t) }

// With lines and arcs alone, the open forms are openProfileXY and sweepOpen.
func TestOpenProfileEdgesXY_LinesAndArcsAreOpenProfileXY(t *testing.T) {
	chain := openCases()[1].chain
	edges := make([]refEdge, len(chain.curvatures))
	for i, k := range chain.curvatures {
		edges[i] = edgeA(k)
		if k == 0 {
			edges[i] = edgeL()
		}
	}
	ref := newRefOpen(t, chain)
	points := c3grid(ref)
	doc := fmt.Sprintf("@scene(140)\nmoulding = %s;\nwant = %s;\nsolid = openProfileEdgesXY(%s, %s);\n"+
		"wantSweep = %s;\nsweepEdges = sweepOpenEdges(moulding, %s, %s);\nreturn solid;\n",
		moulding().gsl(), chain.gsl(), pointsGSL(chain.points), edgesGSL(edges),
		strings.Replace(chain.gsl(), "openProfileXY(", "sweepOpen(moulding, ", 1), pointsGSL(chain.points), edgesGSL(edges))
	for _, pair := range [][2]string{{"solid", "want"}, {"sweepEdges", "wantSweep"}} {
		got := evalProfileField(t, doc, pair[0], points)
		want := evalProfileField(t, doc, pair[1], points)
		for i, q := range points {
			if math.Abs(got[i]-want[i]) > profileTolerance {
				t.Fatalf("%s at %v = %v, %s %v", pair[0], q, got[i], pair[1], want[i])
			}
		}
	}
}

func c3grid(ref refOpen) [][3]float64 {
	b := ref.chainBox()
	return grid3([3]float64{b[0] - 4, b[1] - 4, 0}, [3]float64{b[2] + 4, b[3] + 4, 6}, 5, 14)
}

// sweepOpenEdges along a chain with Bezier edges is sweepOpen's formula from
// the references, its frame is the chain's, and its box holds the hull.
func TestSweepOpenEdges_FormulaFrameBox(t *testing.T) {
	c := edgeChainCases()[0]
	path, section := c.reference(t), newRefProfile(t, moulding())
	first, t0, last, te := c.ends()
	length := path.stations[len(path.stations)-1]
	formula := func(q [3]float64) float64 {
		xy := [2]float64{q[0], q[1]}
		_, station := path.nearest(xy)
		onChain := station >= 0 && station <= length
		wall := section.distance([2]float64{path.distance(xy), q[2]})
		firstDepth := -((q[0]-first[0])*t0[0] + (q[1]-first[1])*t0[1])
		lastDepth := (q[0]-last[0])*te[0] + (q[1]-last[1])*te[1]
		reach := math.Min(
			math.Hypot(firstDepth, math.Max(section.distance([2]float64{across(xy, first, t0), q[2]}), 0)),
			math.Hypot(lastDepth, math.Max(section.distance([2]float64{across(xy, last, te), q[2]}), 0)))
		if onChain {
			reach = math.Min(reach, math.Abs(wall))
			if wall < 0 {
				return -reach
			}
		}
		return reach
	}
	doc := fmt.Sprintf("@scene(140)\nmoulding = %s;\nroute = %s;\nsolid = sweepOpenEdges(moulding, %s, %s);\nreturn solid;\n",
		moulding().gsl(), c.gsl(), pointsGSL(c.points), edgesGSL(c.edges))
	points := c3grid(path)
	compareToReference(t, "sweepOpenEdges", points, evalProfileField(t, doc, "solid", points), formula)

	e := New(&source.MockManager{}, nil)
	groups, err := e.EvalGsl(context.Background(), nil, doc, "",
		[]string{"solid.u", "route.u", "solid.v", "remap(moulding.u, [route, p.z, 0])"}, points[:40])
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	for k, name := range []string{"u", "v"} {
		for i := range points[:40] {
			got, want := float64(groups[2*k].Samples[0][i].Value), float64(groups[2*k+1].Samples[0][i].Value)
			if math.Abs(got-want) > 1e-4 {
				t.Fatalf("solid.%s at %v = %v, want %v", name, points[i], got, want)
			}
		}
	}
	// The box holds the hull of the points and controls, grown by the
	// section's reach (4) and between its lowest and highest (0, 6): a
	// rectangle section, whose own box is exact (an arc profile's is padded).
	boxDoc := fmt.Sprintf("@scene(140)\nsection = profileXY([[-3, 0], [4, 0], [4, 6], [-3, 6]]);\nsolid = sweepOpenEdges(section, %s, %s);\nreturn solid;\n",
		pointsGSL(c.points), edgesGSL(c.edges))
	boxes, err := e.EvalGsl(context.Background(), nil, boxDoc, "", []string{"solid.box"}, [][3]float64{{0, 0, 0}})
	if err != nil {
		t.Fatalf("eval box: %v", err)
	}
	lo, hi := [2]float64{math.Inf(1), math.Inf(1)}, [2]float64{math.Inf(-1), math.Inf(-1)}
	for _, p := range c.points {
		lo, hi = [2]float64{math.Min(lo[0], p[0]), math.Min(lo[1], p[1])}, [2]float64{math.Max(hi[0], p[0]), math.Max(hi[1], p[1])}
	}
	for _, ed := range c.edges {
		if ed.kind == bezierEdge {
			p := ed.control
			lo, hi = [2]float64{math.Min(lo[0], p[0]), math.Min(lo[1], p[1])}, [2]float64{math.Max(hi[0], p[0]), math.Max(hi[1], p[1])}
		}
	}
	want := []float64{lo[0] - 4, lo[1] - 4, 0, hi[0] + 4, hi[1] + 4, 6}
	for k, w := range want {
		if got := float64(boxes[0].Samples[k][0].Value); math.Abs(got-w) > 0.05 {
			t.Errorf("solid.box component %d = %v, want %v", k, got, w)
		}
	}
}
