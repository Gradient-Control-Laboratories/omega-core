package eval

import (
	"container/heap"
	"context"
	"fmt"
	"math"

	"gradientcontrol.com/omega/core/program"
)

const (
	// encloseBudget bounds the interval evaluations of one proof; each of the six faces has a sixth of it.
	encloseBudget = 240_000
	// encloseBatch is how many boxes one round of a search cuts in two.
	encloseBatch = 64
	// encloseSteps is how many leaf thicknesses the fitted span is cut into.
	encloseSteps = 4096
	// encloseLeafAspect is how wide a leaf may be for its thickness. Intervals loosen with a box's width where it
	// spans an edge of the part, so a wide leaf just outside an edge cannot be dropped and the box would stand
	// off by a fraction of that width.
	encloseLeafAspect = 2
	// encloseCut is where a box is cut along a side, as a share of it. Halving a region of round numbers puts cuts
	// on round coordinates, where a part's faces are: a box that only touches a face has a lower bound of zero or
	// a rounding below it and is never dropped. A cut off the half keeps boxes from landing on faces.
	encloseCut = 0.4567891
	// encloseLoose is the share of the fitted span a face may stand off a point known inside when the budget, not
	// a leaf, ended its search. A face looser than that is no answer to "how big is it", and the proof is given up.
	encloseLoose = 1.0 / 64
)

// ReasonUnbounded is why an EncloseResult is not proven when the field's intervals were not finite.
const ReasonUnbounded = "unbounded"

// EncloseResult is a box that holds every point where an expression is negative. When Proven, BboxMin and
// BboxMax are that box by interval arithmetic, and Found is false only where the region is proven to hold no
// such point; CellSize is then, on each axis, how far the box can stand off the part: the gap to a point known
// to be inside, or the search's leaf thickness where it met none. When not Proven, the BoundsResult is Bounds'
// own, a fit on a grid that can miss what is thinner than a cell, and Reason says why no proof was had.
// MinValue, MinAt and Resolution are always the grid's.
type EncloseResult struct {
	BoundsResult
	Proven bool
	Reason string
	Evals  int
}

// Enclose proves a box around the region where expr is negative, within region (zero: Bounds' default). For
// each face it searches sub-boxes best-first by how far they reach toward that face, drops each whose interval
// lower bound is at least zero, and stops when the one reaching farthest is within a leaf's thickness of a point
// known to be inside, or is a leaf it could not drop, or its share of the budget is spent: the farthest reach
// left is a sound bound at any time. Where that bound is no use, because the field's intervals are unbounded or
// too loose for the budget, it returns Bounds' sampled box, unproven.
func (e *Evaluator) Enclose(ctx context.Context, manifest map[string]string, code, defs, expr string, region [6]float64, resolution int) (*EncloseResult, error) {
	sampled, err := e.Bounds(ctx, manifest, code, defs, expr, region, resolution)
	if err != nil {
		return nil, err
	}
	out := &EncloseResult{BoundsResult: *sampled}
	if region == ([6]float64{}) {
		region = [6]float64{
			-defaultBoundsRegion, -defaultBoundsRegion, -defaultBoundsRegion,
			defaultBoundsRegion, defaultBoundsRegion, defaultBoundsRegion,
		}
	}
	if err := startQuery(ctx); err != nil {
		return nil, err
	}
	defer endQuery()
	prog, err := e.prepareQuery(ctx, manifest, code, defs, []string{expr})
	if err != nil {
		return nil, err
	}

	// The leaf thickness: a fraction of the part's span where the scan found one, of the region's where not,
	// and never below what a float32 coordinate resolves there.
	span, reach := 0.0, 0.0
	var box [6]float32
	for k := 0; k < 3; k++ {
		if sampled.Found {
			span = math.Max(span, sampled.BboxMax[k]-sampled.BboxMin[k])
		} else {
			span = math.Max(span, (region[k+3]-region[k])/16)
		}
		reach = math.Max(reach, math.Max(math.Abs(region[k]), math.Abs(region[k+3])))
		box[k] = math.Nextafter32(float32(region[k]), float32(math.Inf(-1)))
		box[k+3] = math.Nextafter32(float32(region[k+3]), float32(math.Inf(1)))
	}
	ulp := float64(math.Nextafter32(float32(reach), float32(math.Inf(1)))) - float64(float32(reach))
	leaf := float32(math.Max(span/encloseSteps, 16*ulp))

	inf := float32(math.Inf(1))
	search := &encloseSearch{ctx: ctx, prog: prog, leaf: leaf, inner: [6]float32{inf, inf, inf, -inf, -inf, -inf}}
	unproven := func() (*EncloseResult, error) {
		out.Reason = ReasonBudget
		if search.unbounded {
			out.Reason = ReasonUnbounded
		}
		return out, nil
	}
	for _, face := range [6]int{3, 0, 4, 1, 5, 2} {
		bound, state, err := search.face(box, face)
		out.Evals = search.evals
		if err != nil {
			return nil, err
		}
		switch state {
		case encloseEmpty:
			out.Proven, out.Found = true, false
			out.BboxMin, out.BboxMax = [3]float64{}, [3]float64{}
			return out, nil
		case encloseUnresolved:
			return unproven()
		}
		box[face] = bound
	}
	var gap [3]float64
	for k := 0; k < 3; k++ {
		gap[k] = float64(leaf)
		if search.inner[k] <= search.inner[k+3] {
			gap[k] = math.Max(float64(search.inner[k]-box[k]), float64(box[k+3]-search.inner[k+3]))
		} else if search.spent {
			return unproven() // the budget ended a face with no point inside to measure it against
		}
		if search.spent && gap[k] > span*encloseLoose {
			return unproven()
		}
	}
	out.Proven, out.Found = true, true
	for k := 0; k < 3; k++ {
		out.BboxMin[k], out.BboxMax[k] = float64(box[k]), float64(box[k+3])
		out.CellSize[k] = gap[k]
	}
	return out, nil
}

const (
	encloseBounded = iota
	encloseEmpty
	encloseUnresolved
)

// encloseNode is a sub-box in a face's search that its interval could not rule out: key is how far it reaches
// toward the face, lower its interval's lower bound, inside whether the middle of its side away from the face is
// in the part, and kept is set once it is a leaf.
type encloseNode struct {
	box    [6]float32
	key    float32
	size   float32
	lower  float32
	inside bool
	kept   bool
}

// encloseHeap orders nodes farthest-reaching first and, among equals, smallest first, so the search dives to a
// leaf on the frontier before it widens.
type encloseHeap []encloseNode

func (h encloseHeap) Len() int { return len(h) }
func (h encloseHeap) Less(i, j int) bool {
	if h[i].key != h[j].key {
		return h[i].key > h[j].key
	}
	return h[i].size < h[j].size
}
func (h encloseHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h *encloseHeap) Push(x any)   { *h = append(*h, x.(encloseNode)) }
func (h *encloseHeap) Pop() any {
	old := *h
	n := old[len(old)-1]
	*h = old[:len(old)-1]
	return n
}

// encloseSearch is the state the six face searches share: inner is the box of every point so far found inside
// the part, which bounds each face from within; spent says the budget ended a face's search.
type encloseSearch struct {
	ctx       context.Context
	prog      *program.Program
	leaf      float32
	inner     [6]float32
	evals     int
	spent     bool
	unbounded bool
}

func unboundedBelow(lower float32) bool { return lower != lower || math.IsInf(float64(lower), 0) }

// intervals is the field's interval over each box.
func (s *encloseSearch) intervals(boxes [][6]float32) ([]IntervalSample, error) {
	if len(boxes) == 0 {
		return nil, nil
	}
	intervals, err := EvalInterval(s.prog, boxes)
	if err != nil {
		return nil, err
	}
	if len(intervals) != 1 || len(intervals[0]) != len(boxes) {
		return nil, fmt.Errorf("bounds expression must be scalar")
	}
	s.evals += len(boxes)
	return intervals[0], nil
}

// face finds how far the part can reach toward one face of root (0 to 2 the low faces of x, y, z; 3 to 5 the
// high): the coordinate beyond which every point is proven outside.
//
// With each box it also evaluates one point, the middle of the box's side away from the face. A box whose point
// is in the part straddles the surface, and is cut toward the face. Any other is most likely outside with an
// interval too loose to say so, and is cut along its longest side, since intervals tighten as a box shrinks.
func (s *encloseSearch) face(root [6]float32, face int) (float32, int, error) {
	axis, low := face%3, face < 3
	started := s.evals
	key := func(box [6]float32) float32 {
		if low {
			return -box[axis]
		}
		return box[axis+3]
	}
	// within is how far the part is known to reach toward the face, as a key.
	within := func() float32 {
		if low {
			return -s.inner[axis]
		}
		return s.inner[axis+3]
	}
	bound := func(key float32) (float32, int, error) {
		if low {
			return -key, encloseBounded, nil
		}
		return key, encloseBounded, nil
	}
	// nodes makes a node of each box its interval does not rule out, with its point looked up.
	nodes := func(boxes [][6]float32, lower []IntervalSample) ([]encloseNode, error) {
		var live []encloseNode
		var points [][6]float32
		for i, box := range boxes {
			if lower[i].Lower >= 0 {
				continue // no point of the part in this box
			}
			at := [3]float32{(box[0] + box[3]) / 2, (box[1] + box[4]) / 2, (box[2] + box[5]) / 2}
			at[axis] = box[axis]
			if low {
				at[axis] = box[axis+3]
			}
			points = append(points, [6]float32{at[0], at[1], at[2], at[0], at[1], at[2]})
			live = append(live, encloseNode{box: box, key: key(box), lower: lower[i].Lower,
				size: (box[3] - box[0]) + (box[4] - box[1]) + (box[5] - box[2])})
		}
		values, err := s.intervals(points)
		if err != nil {
			return nil, err
		}
		for i := range live {
			if values[i].Upper < 0 {
				live[i].inside = true
				for k := 0; k < 3; k++ {
					s.inner[k], s.inner[k+3] = min(s.inner[k], points[i][k]), max(s.inner[k+3], points[i][k])
				}
			}
		}
		return live, nil
	}

	lower, err := s.intervals([][6]float32{root})
	if err != nil {
		return 0, encloseUnresolved, err
	}
	first, err := nodes([][6]float32{root}, lower)
	if err != nil {
		return 0, encloseUnresolved, err
	}
	queue := &encloseHeap{}
	for _, n := range first {
		heap.Push(queue, n)
	}
	for {
		if err := s.ctx.Err(); err != nil {
			return 0, encloseUnresolved, err
		}
		if queue.Len() == 0 {
			return 0, encloseEmpty, nil
		}
		top := (*queue)[0]
		if top.kept || top.key-within() <= s.leaf {
			return bound(top.key)
		}
		if s.evals-started >= encloseBudget/6 {
			if unboundedBelow(top.lower) {
				s.unbounded = true
				return 0, encloseUnresolved, nil
			}
			s.spent = true
			return bound(top.key)
		}

		// The farthest-reaching boxes, each cut in two, the parts evaluated together.
		var boxes [][6]float32
		for queue.Len() > 0 && len(boxes) < 2*encloseBatch && !(*queue)[0].kept {
			n := heap.Pop(queue).(encloseNode)
			split := s.cutAxis(n, axis)
			mid := float32(0)
			if split >= 0 {
				mid = n.box[split] + (n.box[split+3]-n.box[split])*encloseCut
				if !(mid > n.box[split] && mid < n.box[split+3]) {
					split = -1 // too small for a float32 to cut
				}
			}
			if split < 0 {
				// A leaf that could not be dropped bounds the part, unless its interval says nothing at all:
				// an unbounded one would stand at the region's own face and prove no more than the region.
				if !n.inside && unboundedBelow(n.lower) {
					s.unbounded = true
					return 0, encloseUnresolved, nil
				}
				n.kept = true
				heap.Push(queue, n)
				continue
			}
			near, far := n.box, n.box
			near[split+3], far[split] = mid, mid
			boxes = append(boxes, near, far)
		}
		lower, err := s.intervals(boxes)
		if err != nil {
			return 0, encloseUnresolved, err
		}
		children, err := nodes(boxes, lower)
		if err != nil {
			return 0, encloseUnresolved, err
		}
		for _, n := range children {
			heap.Push(queue, n)
		}
	}
}

// cutAxis is the axis to cut a node along in a search toward axis, or -1 for a leaf. A node known to hold a
// point of the part is cut toward the face until it is a leaf's thickness. Any other is cut along its longest
// side, until it is a leaf's thickness toward the face and no more than encloseLeafAspect leaves across it.
func (s *encloseSearch) cutAxis(n encloseNode, axis int) int {
	best, longest := -1, float32(0)
	if thick := n.box[axis+3] - n.box[axis]; thick > s.leaf {
		best, longest = axis, thick
	}
	if n.inside {
		return best
	}
	for _, k := range [2]int{(axis + 1) % 3, (axis + 2) % 3} {
		if wide := n.box[k+3] - n.box[k]; wide > s.leaf*encloseLeafAspect && wide > longest {
			best, longest = k, wide
		}
	}
	return best
}
