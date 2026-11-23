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

// nested(fields): the solid bounded by surfaces that do not cross, each field
// negative on the solid's side. Held to the same reference as one loop.

// plateLoops is an outline with two holes, one square and one a circle of
// arcs, and an island inside the circle: polygon and arc loops mixed,
// counterclockwise outline and island, clockwise holes.
func plateLoops() []refLoop {
	return []refLoop{
		{points: [][2]float64{{-30, -20}, {30, -20}, {30, 20}, {-30, 20}}},
		{points: [][2]float64{{-20, -6}, {-20, 6}, {-8, 6}, {-8, -6}}},
		{
			points:     [][2]float64{{19, 0}, {12, -7}, {5, 0}, {12, 7}},
			curvatures: []float64{-1.0 / 7, -1.0 / 7, -1.0 / 7, -1.0 / 7},
		},
		{points: [][2]float64{{10, -2}, {14, -2}, {14, 2}, {10, 2}}},
	}
}

// islandLoops are two separate counterclockwise loops.
func islandLoops() []refLoop {
	return []refLoop{lShape(), {points: [][2]float64{{40, 0}, {50, 0}, {50, 10}, {40, 10}}}}
}

// nestedExpr is nested() over each loop's profileXY, in the given order.
func nestedExpr(loops []refLoop, order []int) string {
	fields := make([]string, len(order))
	for i, k := range order {
		fields[i] = loops[k].gsl()
	}
	return fmt.Sprintf("nested([%s])", strings.Join(fields, ", "))
}

func nestedCases(t *testing.T) []profileCase {
	plate, islands := plateLoops(), islandLoops()
	return []profileCase{
		{name: "plate", expr: nestedExpr(plate, []int{0, 1, 2, 3}), ref: newRefProfile(t, plate...)},
		{name: "plate in another order", expr: nestedExpr(plate, []int{3, 1, 0, 2}), ref: newRefProfile(t, plate...)},
		{name: "separate islands", expr: nestedExpr(islands, []int{0, 1}), ref: newRefProfile(t, islands...)},
	}
}

// nested over exact loops is exact: it agrees with the reference over every
// loop at once, in any order.
func TestNested_MatchesReference(t *testing.T) {
	for _, c := range nestedCases(t) {
		t.Run(c.name, func(t *testing.T) {
			points := flat(gridAround(c.ref.box(), 6, 56))
			got := evalProfileField(t, "@scene(140)\nsolid = "+c.expr+";\nreturn solid;\n", "solid", points)
			compareToReference(t, c.name, points, got, func(q [3]float64) float64 {
				return c.ref.distance([2]float64{q[0], q[1]})
			})
		})
	}
}

// Extruded, the plate is the exact solid, its surface cells are tight,
// certify proves its holes and the space above it empty, and bounds proves its
// box.
func TestNested_ExtrudedPlate(t *testing.T) {
	const low, high, gap = 0.0, 6.0, 0.25
	c := nestedCases(t)[0]
	doc := fmt.Sprintf("@scene(140)\nsolid = extrude(%s, %v, %v);\nreturn solid;\n", c.expr, low, high)
	box := c.ref.box()

	points := grid3([3]float64{box[0], box[1], low}, [3]float64{box[2], box[3], high}, 4, 24)
	compareToReference(t, "extruded plate", points, evalProfileField(t, doc, "solid", points), extrudedReference(c.ref, low, high))

	checkSurfaceCells(t, "extruded plate", doc, rimCentres(c.ref, low, high), extrudedReference(c.ref, low, high))

	e := New(&source.MockManager{}, nil)
	for where, region := range map[string][6]float64{
		"above it":        {box[0] - 5, box[1] - 5, high + gap, box[2] + 5, box[3] + 5, high + 5},
		"the square hole": {-20 + gap, -6 + gap, low - 5, -8 - gap, 6 - gap, high + 5},
		"round the island": {
			// The ring between the island and the circle is not a box; this box
			// lies in it, above the island.
			10 + gap, 2 + gap, low - 5, 14 - gap, 5, high + 5,
		},
	} {
		r, err := e.Certify(context.Background(), nil, doc, "", "solid", region, 0, 0)
		if err != nil {
			t.Fatalf("certify %s: %v", where, err)
		}
		if !r.Proven {
			t.Errorf("certify %s: want proven, got proven=%v disproven=%v indeterminate=%v in %d evals",
				where, r.Proven, r.Disproven, r.Indeterminate, r.Evals)
		}
	}

	region := [6]float64{box[0] - 10, box[1] - 10, low - 10, box[2] + 10, box[3] + 10, high + 10}
	r, err := e.Enclose(context.Background(), nil, doc, "", "solid", region, 0)
	if err != nil {
		t.Fatalf("enclose: %v", err)
	}
	if !r.Proven {
		t.Fatalf("enclose: want a proven box, got %q", r.Reason)
	}
	want := [6]float64{box[0], box[1], low, box[2], box[3], high}
	for k := 0; k < 3; k++ {
		if r.BboxMin[k] > want[k]+profileTolerance || r.BboxMax[k] < want[k+3]-profileTolerance ||
			want[k]-r.BboxMin[k] > r.CellSize[k]+profileTolerance || r.BboxMax[k]-want[k+3] > r.CellSize[k]+profileTolerance {
			t.Errorf("axis %d: proven box [%v, %v], reference [%v, %v], stand-off %v",
				k, r.BboxMin[k], r.BboxMax[k], want[k], want[k+3], r.CellSize[k])
		}
	}
}

// The plate's .u is its nearest loop's own.
func TestNested_FrameIsTheNearestLoops(t *testing.T) {
	plate := plateLoops()
	var doc strings.Builder
	doc.WriteString("@scene(140)\n")
	names := make([]string, len(plate))
	for i, l := range plate {
		names[i] = fmt.Sprintf("loop%d", i)
		fmt.Fprintf(&doc, "%s = %s;\n", names[i], l.gsl())
	}
	fmt.Fprintf(&doc, "solid = nested([%s]);\nreturn solid;\n", strings.Join(names, ", "))

	refs := make([]refProfile, len(plate))
	for i, l := range plate {
		refs[i] = newRefProfile(t, l)
	}
	// Points whose nearest loop is nearer than the next by a clear margin.
	var points [][3]float64
	var nearest []int
	for _, q := range gridAround(newRefProfile(t, plate...).box(), 3, 30) {
		best, second, at := math.Inf(1), math.Inf(1), -1
		for i, r := range refs {
			d := math.Abs(r.distance(q))
			if d < best {
				best, second, at = d, best, i
			} else if d < second {
				second = d
			}
		}
		if second-best > 0.5 {
			points = append(points, [3]float64{q[0], q[1], 0})
			nearest = append(nearest, at)
		}
	}
	exprs := []string{"solid.u"}
	for _, n := range names {
		exprs = append(exprs, n+".u")
	}
	e := New(&source.MockManager{}, nil)
	groups, err := e.EvalGsl(context.Background(), nil, doc.String(), "", exprs, points)
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	failures := 0
	for i, q := range points {
		got := float64(groups[0].Samples[0][i].Value)
		want := float64(groups[1+nearest[i]].Samples[0][i].Value)
		if math.Abs(got-want) > 1e-4 {
			if failures++; failures <= 3 {
				t.Errorf("solid.u at %v = %v, want loop%d's %v", q, got, nearest[i], want)
			}
		}
	}
	t.Logf("%d points checked", len(points))
}

// Generic: a 3D box with a spherical void, each exact, nested is the exact
// hollow solid.
func TestNested_SolidWithAVoid(t *testing.T) {
	const doc = `@scene(60)
corner = abs(p) - [15, 10, 8];
block = length(max(corner, 0)) + min(max(corner.x, max(corner.y, corner.z)), 0);
hollow = neg(length(p - [3, 0, 0]) - 5);
solid = nested([block, hollow]);
return solid;
`
	blockAt := func(q [3]float64) float64 {
		cx, cy, cz := math.Abs(q[0])-15, math.Abs(q[1])-10, math.Abs(q[2])-8
		out := math.Sqrt(math.Pow(math.Max(cx, 0), 2) + math.Pow(math.Max(cy, 0), 2) + math.Pow(math.Max(cz, 0), 2))
		return out + math.Min(math.Max(cx, math.Max(cy, cz)), 0)
	}
	want := func(q [3]float64) float64 {
		b := blockAt(q)
		s := math.Sqrt(math.Pow(q[0]-3, 2)+q[1]*q[1]+q[2]*q[2]) - 5
		switch {
		case b >= 0:
			return b // outside the block, which holds the void
		case s <= 0:
			return -s // in the void
		default:
			return -math.Min(-b, s)
		}
	}
	points := grid3([3]float64{-15, -10, -8}, [3]float64{15, 10, 8}, 4, 20)
	compareToReference(t, "block with a void", points, evalProfileField(t, doc, "solid", points), want)
}

// The winding rule, pinned: a counterclockwise loop standing in the material,
// where a hole's clockwise winding belongs, is not checked. Where it is the
// nearest loop nested reports its own distance: right in sign inside it, but
// positive just outside it, in material, so the field there is void and jumps
// sign, without passing zero, where another loop becomes the nearer.
func TestNested_MisWoundLoopIsPinned(t *testing.T) {
	outline := refLoop{points: [][2]float64{{-20, -20}, {20, -20}, {20, 20}, {-20, 20}}}
	hole := refLoop{points: [][2]float64{{-8, -8}, {-8, 8}, {8, 8}, {8, -8}}}
	misWound := refLoop{points: [][2]float64{{11, 11}, {17, 11}, {17, 17}, {11, 17}}}
	doc := fmt.Sprintf("@scene(60)\nsolid = nested([%s, %s, %s]);\nreturn solid;\n", outline.gsl(), hole.gsl(), misWound.gsl())
	own := newRefProfile(t, misWound)
	points := [][3]float64{{12, 12, 0}, {14, 14, 0}, {18, 14, 0}, {14, 10, 0}}
	got := evalProfileField(t, doc, "solid", points)
	for i, q := range points {
		want := own.distance([2]float64{q[0], q[1]})
		if math.Abs(got[i]-want) > profileTolerance {
			t.Errorf("at %v nested = %v, want the mis-wound loop's own %v", q, got[i], want)
		}
	}
	// (18, 14) and (14, 10) lie in the material, yet read positive.
	for _, i := range []int{2, 3} {
		if got[i] <= 0 {
			t.Errorf("at %v nested = %v: the pinned behaviour reads positive in the material there", points[i], got[i])
		}
	}
}
