//go:build cgo

package eval

import (
	"context"
	"fmt"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// Open profiles on chains beyond openCases and edgeChainCases: single edges,
// near-degenerate ones, joints of every kind, chains far from the origin,
// rotated and reversed, and Beziers that hook back past their chords. Each is
// held to the reference as openCases are; every chain keeps the condition the
// doc comments state, that neither ray crosses the chain or the other ray.

type robustOpen struct {
	name  string
	chain refChain
}

func robustOpenCases() []robustOpen {
	zigzag, s := openCases()[0].chain, openCases()[1].chain
	return []robustOpen{
		{"one line", refChain{points: [][2]float64{{0, 0}, {10, 0}}}},
		// A semicircle turns its tangent round: the end ray runs back beside it.
		{"line then semicircle", refChain{points: [][2]float64{{0, 0}, {10, 0}, {20, 0}}, curvatures: []float64{0, 0.2}}},
		{"line then near semicircle", refChain{points: [][2]float64{{0, 0}, {10, 0}, {20, 0}}, curvatures: []float64{0, 0.1999}}},
		{"tiny curvatures", refChain{points: [][2]float64{{0, 0}, {10, 0}, {20, 5}}, curvatures: []float64{1e-4, 1e-6}}},
		{"zero curvatures", refChain{points: zigzag.points, curvatures: []float64{0, 0, 0, 0}}},
		{"collinear interior", refChain{points: [][2]float64{{0, 0}, {10, 0}, {20, 0}, {30, 5}}}},
		{"collinear interior, arcs form", refChain{points: [][2]float64{{0, 0}, {10, 0}, {20, 0}, {30, 5}}, curvatures: []float64{0, 0, 0}}},
		{"short edge", refChain{points: [][2]float64{{0, 0}, {10, 0}, {10.01, 0.01}, {20, 5}}}},
		// Half a turn exactly: the rays run parallel, the edge of the condition.
		{"U turn", refChain{points: [][2]float64{{0, 10}, {0, 0}, {10, 0}, {10, 10}}}},
		{"hairpin", refChain{points: [][2]float64{{0, 0}, {20, 0}, {0, 2}}}},
		{"arc to arc, smooth", refChain{points: [][2]float64{{0, 0}, {10, 0}, {20, 0}}, curvatures: []float64{0.1, -0.1}}},
		{"arc to arc, corner left", refChain{points: [][2]float64{{0, 0}, {10, 0}, {20, 0}}, curvatures: []float64{0.1, 0.1}}},
		{"arc to arc, corner right", refChain{points: [][2]float64{{0, 0}, {10, 0}, {20, 0}}, curvatures: []float64{-0.1, -0.1}}},
		{"far from the origin", movedChain(zigzag, 0, 1000, -2000)},
		{"rotated", movedChain(s, 0.645, 3, -7)},
	}
}

// movedChain is c turned by angle about the origin, then moved by (dx, dy).
func movedChain(c refChain, angle, dx, dy float64) refChain {
	co, si := math.Cos(angle), math.Sin(angle)
	out := refChain{curvatures: c.curvatures}
	for _, q := range c.points {
		out.points = append(out.points, [2]float64{q[0]*co - q[1]*si + dx, q[0]*si + q[1]*co + dy})
	}
	return out
}

// reversedChain is c walked the other way: its points reversed, each edge's
// curvature negated, since an arc that turns left one way turns right the other.
func reversedChain(c refChain) refChain {
	var out refChain
	for i := len(c.points) - 1; i >= 0; i-- {
		out.points = append(out.points, c.points[i])
	}
	for i := len(c.curvatures) - 1; i >= 0; i-- {
		out.curvatures = append(out.curvatures, -c.curvatures[i])
	}
	return out
}

// robustPoints are a grid round the chain's box, at least 1 mm across each
// way, and a ring of far points 100 and 400 mm from its middle.
func robustPoints(box [4]float64) [][2]float64 {
	for _, k := range []int{0, 1} {
		if box[k+2]-box[k] < 1 {
			box[k], box[k+2] = box[k]-5, box[k+2]+5
		}
	}
	points := gridAround(box, 10, 48)
	cx, cy := (box[0]+box[2])/2, (box[1]+box[3])/2
	for i := 0; i < 64; i++ {
		a := 2 * math.Pi * float64(i) / 64
		for _, r := range []float64{100, 400} {
			points = append(points, [2]float64{cx + r*math.Cos(a), cy + r*math.Sin(a)})
		}
	}
	return points
}

// checkRobustField fails unless the field at every point is finite, on the
// reference's side wherever the reference is off the chain, and within
// profileTolerance of it, grown by float32's spacing at the point's distance
// from the origin.
func checkRobustField(t *testing.T, label, doc string, points [][2]float64, want func([2]float64) float64) {
	t.Helper()
	got := evalProfileField(t, doc, "solid", flat(points))
	worst, worstAt := 0.0, -1
	for i, q := range points {
		w := want(q)
		if math.IsNaN(got[i]) || math.IsInf(got[i], 0) {
			t.Fatalf("%s at %v = %v, reference %v", label, q, got[i], w)
		}
		if math.Abs(w) > 1e-3 && (got[i] < 0) != (w < 0) {
			t.Errorf("%s at %v = %v, on the other side from the reference's %v", label, q, got[i], w)
		}
		tolerance := profileTolerance + 1.2e-7*math.Max(math.Abs(q[0]), math.Abs(q[1]))
		if diff := math.Abs(got[i] - w); diff/tolerance > worst {
			worst, worstAt = diff/tolerance, i
		}
	}
	q := points[worstAt]
	t.Logf("%s: worst difference %.2g mm, %.2g of its tolerance, over %d points", label, math.Abs(got[worstAt]-want(q)), worst, len(points))
	if worst > 1 {
		t.Errorf("%s at %v = %v, reference %v", label, q, got[worstAt], want(q))
	}
}

// checkRobustFrame fails unless .u is within 1e-3 mm of the reference's
// station wherever the nearest point is unique, a little off the chain on
// both sides and along the rays. Beside an end, where the ray and the chain
// tie to within the 1e-4 mm openProfileRaysRun allows, .u may be the end's
// station instead of a point up to sqrt(2 d 1e-4) along from it, at distance d
// from the chain: that much is allowed too.
func checkRobustFrame(t *testing.T, label, doc string, ref refOpen) {
	t.Helper()
	var points [][3]float64
	for _, s := range ref.samples(1.3, 8, 8) {
		for _, d := range []float64{-0.4, 0.4} {
			points = append(points, [3]float64{s[0], s[1] + d, 0}, [3]float64{s[0] + d, s[1], 0})
		}
	}
	e := New(&source.MockManager{}, nil)
	groups, err := e.EvalGsl(context.Background(), nil, doc, "", []string{"solid.u"}, points)
	if err != nil {
		t.Fatalf("%s: eval .u: %v", label, err)
	}
	checked, worst := 0, 0.0
	for i, q := range points {
		d, want := ref.nearest([2]float64{q[0], q[1]})
		if !uniqueNearest(ref, [2]float64{q[0], q[1]}, d) {
			continue
		}
		checked++
		got := float64(groups[0].Samples[0][i].Value)
		if math.IsNaN(got) || math.Abs(got-want) > 1e-3+math.Sqrt(2*d*1e-4) {
			t.Errorf("%s: .u at %v = %v, want %v", label, q, got, want)
		}
		worst = math.Max(worst, math.Abs(got-want))
	}
	t.Logf("%s: .u at %d points, worst difference %.2g mm", label, checked, worst)
}

func TestOpenProfileXY_MoreChains(t *testing.T) {
	for _, c := range robustOpenCases() {
		t.Run(c.name, func(t *testing.T) {
			ref := newRefOpen(t, c.chain)
			doc := "@scene(120)\nsolid = " + c.chain.gsl() + ";\nreturn solid;\n"
			checkRobustField(t, c.name, doc, robustPoints(ref.chainBox()), ref.distance)
			checkRobustFrame(t, c.name, doc, ref)
		})
	}
}

// Walked the other way, a chain's field is negated and its .u runs back from
// its length, to within the tie checkRobustFrame allows: near a corner the
// two walks can settle a tie on different sides of it.
func TestOpenProfileXY_Reversed(t *testing.T) {
	for _, c := range openCases()[:2] {
		t.Run(c.name, func(t *testing.T) {
			ref := newRefOpen(t, c.chain)
			length := ref.stations[len(ref.stations)-1]
			points := flat(gridAround(ref.chainBox(), 10, 32))
			forward := "@scene(120)\nsolid = " + c.chain.gsl() + ";\nreturn solid;\n"
			backward := "@scene(120)\nsolid = " + reversedChain(c.chain).gsl() + ";\nreturn solid;\n"
			e := New(&source.MockManager{}, nil)
			there, err := e.EvalGsl(context.Background(), nil, forward, "", []string{"solid", "solid.u"}, points)
			if err != nil {
				t.Fatalf("eval: %v", err)
			}
			back, err := e.EvalGsl(context.Background(), nil, backward, "", []string{"solid", "solid.u"}, points)
			if err != nil {
				t.Fatalf("eval reversed: %v", err)
			}
			for i, q := range points {
				d, u := float64(there[0].Samples[0][i].Value), float64(there[1].Samples[0][i].Value)
				dr, ur := float64(back[0].Samples[0][i].Value), float64(back[1].Samples[0][i].Value)
				if math.Abs(d+dr) > profileTolerance {
					t.Errorf("at %v: %v forward, %v reversed", q, d, dr)
				}
				dist, _ := ref.nearest([2]float64{q[0], q[1]})
				if uniqueNearest(ref, [2]float64{q[0], q[1]}, dist) && math.Abs(u+ur-length) > 1e-3+math.Sqrt(2*dist*1e-4) {
					t.Errorf("at %v: .u %v forward, %v reversed, length %v", q, u, ur, length)
				}
			}
		})
	}
}

type robustEdges struct {
	name  string
	chain edgeChain
	// reference, when set, is a chain of the same shape whose edges the
	// reference can sample: a degenerate Bezier's straight line.
	reference *edgeChain
}

func robustEdgeCases() []robustEdges {
	return []robustEdges{
		{name: "one Bezier", chain: edgeChain{points: [][2]float64{{0, 0}, {10, 0}}, edges: []refEdge{edgeB(5, 5)}}},
		{name: "two lines", chain: edgeChain{points: [][2]float64{{0, 0}, {10, 0}, {20, 5}}, edges: []refEdge{edgeL(), edgeL()}}},
		{name: "Bezier last", chain: edgeChain{points: [][2]float64{{0, 0}, {10, 0}, {20, 10}}, edges: []refEdge{edgeL(), edgeB(20, 0)}}},
		{name: "Bezier corners left", chain: edgeChain{points: [][2]float64{{0, 0}, {10, 0}, {20, 0}}, edges: []refEdge{edgeB(5, 5), edgeB(15, 5)}}},
		{name: "Bezier corners right", chain: edgeChain{points: [][2]float64{{0, 0}, {10, 0}, {20, 0}}, edges: []refEdge{edgeB(5, -5), edgeB(15, -5)}}},
		{name: "straight Bezier last", chain: edgeChain{points: [][2]float64{{0, 0}, {10, 0}, {20, 0}}, edges: []refEdge{edgeL(), edgeB(15, 0)}}},
		{name: "Bezier U", chain: edgeChain{points: [][2]float64{{0, 10}, {0, 0}, {10, 0}, {10, 10}}, edges: []refEdge{edgeL(), edgeB(5, -6), edgeL()}}},
		{name: "arc then Bezier", chain: edgeChain{points: [][2]float64{{0, 0}, {10, 0}, {20, 5}}, edges: []refEdge{edgeA(0.1), edgeB(18, 0)}}},
		// Beziers whose end tangent turns past a right angle from the chord.
		{name: "Bezier hooks at the end", chain: edgeChain{points: [][2]float64{{0, 0}, {10, 0}, {12, 2}}, edges: []refEdge{edgeL(), edgeB(30, 0)}}},
		{name: "Bezier hooks at the start", chain: edgeChain{points: [][2]float64{{12, 2}, {10, 0}, {0, 0}}, edges: []refEdge{edgeB(30, 0), edgeL()}}},
		{name: "Bezier hooks both ways", chain: edgeChain{points: [][2]float64{{0, 0}, {10, 0}, {12, 2}, {0, 4}}, edges: []refEdge{edgeL(), edgeB(30, 0), edgeB(30, 4)}}},
		// A control on a chain's end: the ray runs along the chord.
		{
			name:      "Bezier control on the first point",
			chain:     edgeChain{points: [][2]float64{{0, 0}, {10, 0}, {20, 5}}, edges: []refEdge{edgeB(0, 0), edgeL()}},
			reference: &edgeChain{points: [][2]float64{{0, 0}, {10, 0}, {20, 5}}, edges: []refEdge{edgeL(), edgeL()}},
		},
		{
			name:      "Bezier control on the last point",
			chain:     edgeChain{points: [][2]float64{{0, 0}, {10, 0}, {20, 5}}, edges: []refEdge{edgeL(), edgeB(20, 5)}},
			reference: &edgeChain{points: [][2]float64{{0, 0}, {10, 0}, {20, 5}}, edges: []refEdge{edgeL(), edgeL()}},
		},
	}
}

func TestOpenProfileEdgesXY_MoreChains(t *testing.T) {
	for _, c := range robustEdgeCases() {
		t.Run(c.name, func(t *testing.T) {
			shape := c.chain
			if c.reference != nil {
				shape = *c.reference
			}
			ref := shape.reference(t)
			doc := "@scene(120)\nsolid = " + c.chain.gsl() + ";\nreturn solid;\n"
			checkRobustField(t, c.name, doc, robustPoints(ref.chainBox()), ref.distance)
			if c.reference == nil {
				checkRobustFrame(t, c.name, doc, ref)
			}
		})
	}
}

// A closed loop with a hooked Bezier edge, as the open chains above.
func TestProfileEdgesXY_HookedBezier(t *testing.T) {
	for _, l := range []edgeLoop{
		{points: [][2]float64{{0, 0}, {10, 0}, {12, 2}, {0, 2}}, edges: []refEdge{edgeL(), edgeB(30, 0), edgeL(), edgeL()}},
		{points: [][2]float64{{0, 0}, {10, 0}, {12, 2}, {0, 4}}, edges: []refEdge{edgeL(), edgeB(30, 0), edgeB(30, 4), edgeL()}},
	} {
		t.Run(l.gsl(), func(t *testing.T) {
			ref, _ := l.reference(t)
			doc := "@scene(120)\nsolid = " + l.gsl() + ";\nreturn solid;\n"
			checkRobustField(t, "hooked loop", doc, robustPoints([4]float64{-2, -2, 22, 6}), ref.distance)
		})
	}
}

// An open chain of Beziers: .u against the reference, cells straddling the
// extended chain sound and as tight as a closed Bezier profile's, and certify
// proving a region on its right empty.
func TestOpenProfileEdgesXY_FrameAndProves(t *testing.T) {
	clear := [][4]float64{{32, -10, 50, 24}, {-10, -10, 30, -0.25}}
	for i, c := range edgeChainCases() {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			ref := c.reference(t)
			doc := "@scene(120)\nsolid = " + c.gsl() + ";\nreturn solid;\n"
			checkRobustFrame(t, c.gsl(), doc, ref)
			checkSurfaceCellsWithin(t, c.gsl(), doc, flat(ref.samples(1, 8, 8)), bezierCellSpread, func(q [3]float64) float64 {
				return ref.distance([2]float64{q[0], q[1]})
			})
			e := New(&source.MockManager{}, nil)
			r, err := e.Certify(context.Background(), nil, doc, "", "solid",
				[6]float64{clear[i][0], clear[i][1], -1, clear[i][2], clear[i][3], 1}, 0, 0)
			if err != nil {
				t.Fatalf("certify: %v", err)
			}
			if !r.Proven {
				t.Errorf("right of the chain: want proven, got proven=%v disproven=%v indeterminate=%v in %d evals",
					r.Proven, r.Disproven, r.Indeterminate, r.Evals)
			}
			t.Logf("certify proves the right side in %d evals", r.Evals)
		})
	}
}
