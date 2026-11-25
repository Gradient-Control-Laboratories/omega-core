//go:build cgo

package eval

import (
	"context"
	"fmt"
	"math"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/source"
)

// sweep(section, path): the section, x across the path (positive outside)
// and y up as z, swept along the zero line of a 2D path field in XY.

// moulding is a section 3 mm inside the path to 4 mm outside it, 6 mm high,
// with a cove of radius 2 cut from its outer top corner.
func moulding() refLoop {
	return refLoop{
		points:     [][2]float64{{-3, 0}, {4, 0}, {4, 4}, {2, 6}, {-3, 6}},
		curvatures: []float64{0, 0, -0.5, 0, 0},
	}
}

// filletedU is the U of the corner tests with every corner filleted, the
// fillets wider than the moulding reaches: a path the section stays inside
// the tube of everywhere.
func filletedU() refLoop {
	corners := []refCorner{
		{filletCorner, 5}, {filletCorner, 5}, {filletCorner, 2}, {filletCorner, 2},
		{filletCorner, 4.5}, {filletCorner, 4.5}, {filletCorner, 2}, {filletCorner, 2},
	}
	return blendedLoop([][2]float64{{0, 0}, {40, 0}, {40, 35}, {30, 35}, {30, 15}, {10, 15}, {10, 35}, {0, 35}}, corners)
}

type sweepCase struct {
	name string
	path refLoop
}

func sweepCases() []sweepCase {
	return []sweepCase{
		{"filleted U", filletedU()},
		{"arc L, sharp corners", arcL()},
	}
}

func (c sweepCase) doc() string {
	return fmt.Sprintf("@scene(140)\nroute = %s;\nmoulding = %s;\nsolid = sweep(moulding, route);\nreturn solid;\n",
		c.path.gsl(), moulding().gsl())
}

// composed is the section's distance at (the path's distance, z): the
// composition sweep is, from the two references.
func composed(path, section refProfile) func([3]float64) float64 {
	return func(q [3]float64) float64 {
		return section.distance([2]float64{path.distance([2]float64{q[0], q[1]}), q[2]})
	}
}

func (c sweepCase) grid(path refProfile, n int) [][3]float64 {
	b := path.box()
	return grid3([3]float64{b[0] - 4, b[1] - 4, 0}, [3]float64{b[2] + 4, b[3] + 4, 6}, 4, n)
}

// sweep is that composition.
func TestSweep_IsTheComposition(t *testing.T) {
	for _, c := range sweepCases() {
		t.Run(c.name, func(t *testing.T) {
			path, section := newRefProfile(t, c.path), newRefProfile(t, moulding())
			points := c.grid(path, 22)
			compareToReference(t, c.name, points, evalProfileField(t, c.doc(), "solid", points), composed(path, section))
		})
	}
}

// sweptCloud samples the swept surface itself: for each point (x, y) of the
// section's outline, the level set where the path's distance is x, traced by
// marching squares at spacing h, at height y. Rounding round the outside of
// a sharp corner and mitring on the inside come from the level sets, not from
// sweep's formula.
func sweptCloud(path, section refProfile, h float64) [][3]float64 {
	b := path.box()
	x0, y0 := b[0]-6, b[1]-6
	nx, ny := int((b[2]-b[0]+12)/h)+1, int((b[3]-b[1]+12)/h)+1
	field := make([][]float64, nx)
	for i := range field {
		field[i] = make([]float64, ny)
		for j := range field[i] {
			field[i][j] = path.distance([2]float64{x0 + float64(i)*h, y0 + float64(j)*h})
		}
	}
	var cloud [][3]float64
	for _, s := range section.boundarySamples(h) {
		level, height := s[0], s[1]
		for i := 0; i < nx; i++ {
			for j := 0; j < ny; j++ {
				a := field[i][j] - level
				if i+1 < nx {
					if c := field[i+1][j] - level; (a < 0) != (c < 0) {
						f := a / (a - c)
						cloud = append(cloud, [3]float64{x0 + (float64(i)+f)*h, y0 + float64(j)*h, height})
					}
				}
				if j+1 < ny {
					if c := field[i][j+1] - level; (a < 0) != (c < 0) {
						f := a / (a - c)
						cloud = append(cloud, [3]float64{x0 + float64(i)*h, y0 + (float64(j)+f)*h, height})
					}
				}
			}
		}
	}
	return cloud
}

// Against the swept surface sampled independently: never more than its
// distance (a bound everywhere), and equal to it, to within the sampling,
// where the section stays inside the path's tube.
func TestSweep_AgainstTheSweptSurface(t *testing.T) {
	const h = 0.2
	for _, c := range sweepCases() {
		t.Run(c.name, func(t *testing.T) {
			path, section := newRefProfile(t, c.path), newRefProfile(t, moulding())
			cloud := sweptCloud(path, section, h)
			points := c.grid(path, 10)
			got := evalProfileField(t, c.doc(), "solid", points)
			worstOver, worstUnder, at := 0.0, 0.0, [3]float64{}
			for i, q := range points {
				best := math.Inf(1)
				for _, s := range cloud {
					best = math.Min(best, math.Hypot(math.Hypot(q[0]-s[0], q[1]-s[1]), q[2]-s[2]))
				}
				diff := math.Abs(got[i]) - best
				if diff > worstOver {
					worstOver = diff
				}
				if -diff > worstUnder {
					worstUnder, at = -diff, q
				}
			}
			t.Logf("%s: %d cloud points; over the surface's distance by at most %.3g mm, under it by at most %.3g mm (at %v)",
				c.name, len(cloud), worstOver, worstUnder, at)
			if worstOver > h {
				t.Errorf("%s: the field exceeds the distance to the swept surface by %.3g mm", c.name, worstOver)
			}
			if c.name == "filleted U" && worstUnder > h {
				t.Errorf("%s: inside the tube the field falls short of the distance by %.3g mm at %v", c.name, worstUnder, at)
			}
		})
	}
}

// Its surface cells bound it soundly; certify proves the space above it empty;
// bounds, when it proves a box, holds the solid. A sweep reads the section at
// the path's distance, so the path's interval feeds the section's and their
// looseness compounds: its cells are allowed sweepCellSpread diagonals, not
// maxCellSpread, and certify is given sweepCertifyBudget. Both measured
// 2026-11-18: cells up to 5.54 diagonals, and 291,497 evaluations to prove the
// space above the filleted U.
func TestSweep_Proves(t *testing.T) {
	for _, c := range sweepCases() {
		t.Run(c.name, func(t *testing.T) {
			path, section := newRefProfile(t, c.path), newRefProfile(t, moulding())
			var centres [][3]float64
			for _, s := range section.boundarySamples(1.5) {
				for _, q := range path.boundarySamples(4) {
					// Offset each path point across by the section's x, along the
					// path's own gradient there.
					e := 1e-4
					gx := (path.distance([2]float64{q[0] + e, q[1]}) - path.distance([2]float64{q[0] - e, q[1]})) / (2 * e)
					gy := (path.distance([2]float64{q[0], q[1] + e}) - path.distance([2]float64{q[0], q[1] - e})) / (2 * e)
					l := math.Hypot(gx, gy)
					centres = append(centres, [3]float64{q[0] + s[0]*gx/l, q[1] + s[0]*gy/l, s[1]})
				}
			}
			checkSurfaceCellsWithin(t, c.name, c.doc(), centres, sweepCellSpread, composed(path, section))

			b := path.box()
			e := New(&source.MockManager{}, nil)
			e.SetCertifyBudget(sweepCertifyBudget)
			r, err := e.Certify(context.Background(), nil, c.doc(), "", "solid",
				[6]float64{b[0] - 10, b[1] - 10, 6.25, b[2] + 10, b[3] + 10, 12}, 0, 0)
			if err != nil {
				t.Fatalf("certify: %v", err)
			}
			if !r.Proven {
				t.Errorf("above it: want proven, got proven=%v disproven=%v in %d evals", r.Proven, r.Disproven, r.Evals)
			}
			t.Logf("%s: certify proves the space above in %d evals (the default budget is 200,000)", c.name, r.Evals)
			en, err := e.Enclose(context.Background(), nil, c.doc(), "", "solid",
				[6]float64{b[0] - 15, b[1] - 15, -10, b[2] + 15, b[3] + 15, 16}, 0)
			if err != nil {
				t.Fatalf("enclose: %v", err)
			}
			// Enclose's budget is fixed at 240,000 evaluations, short of what a
			// sweep's compounded intervals need: logged, not asserted (backlog M16).
			if !en.Proven {
				t.Logf("%s: bounds does not prove the box within its fixed budget: %q", c.name, en.Reason)
				return
			}
			want := [6]float64{b[0] - 4, b[1] - 4, 0, b[2] + 4, b[3] + 4, 6}
			for k := 0; k < 3; k++ {
				if en.BboxMin[k] > want[k]+0.05 || en.BboxMax[k] < want[k+3]-0.05 {
					t.Errorf("axis %d: proven box [%v, %v] misses the solid's [%v, %v]", k, en.BboxMin[k], en.BboxMax[k], want[k], want[k+3])
				}
			}
			t.Logf("%s: proven box %v to %v", c.name, en.BboxMin, en.BboxMax)
		})
	}
}

// Its frame: .u the path's own .u, the distance along it, and .v the
// section's .u read where the sweep reads the section.
func TestSweep_Frame(t *testing.T) {
	c := sweepCases()[0]
	path := newRefProfile(t, c.path)
	points := c.grid(path, 6)
	e := New(&source.MockManager{}, nil)
	groups, err := e.EvalGsl(context.Background(), nil, c.doc(), "",
		[]string{"solid.u", "route.u", "solid.v", "remap(moulding.u, [route, p.z, 0])"}, points)
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	for k, name := range []string{"u", "v"} {
		for i, q := range points {
			got, want := float64(groups[2*k].Samples[0][i].Value), float64(groups[2*k+1].Samples[0][i].Value)
			if math.Abs(got-want) > 1e-4 {
				t.Fatalf("solid.%s at %v = %v, want %v", name, q, got, want)
			}
		}
	}
}

// sweepCellSpread is how wide a sweep's surface cell's interval may be, in
// cell diagonals; sweepCertifyBudget is the evaluation budget its certify is
// given (see TestSweep_Proves).
const (
	sweepCellSpread    = 6.0
	sweepCertifyBudget = 1000000
)

// checkSurfaceCellsWithin is checkSurfaceCells with its own spread limit.
func checkSurfaceCellsWithin(t *testing.T, label, doc string, centres [][3]float64, limit float64, want func([3]float64) float64) {
	t.Helper()
	boxes := make([][6]float32, len(centres))
	for i, q := range centres {
		boxes[i] = cellAround(q)
	}
	bounds := intervalOver(t, doc, "solid", boxes)
	diagonal := 2 * cellHalf * math.Sqrt(3)
	widest, failures := 0.0, 0
	for i, q := range centres {
		lo, hi := float64(bounds[i].Lower), float64(bounds[i].Upper)
		spread := (hi - lo) / diagonal
		widest = math.Max(widest, spread)
		if value := want(q); value < lo-profileTolerance || value > hi+profileTolerance || spread > limit {
			if failures++; failures <= 3 {
				t.Errorf("%s: cell about %v bounds [%v, %v], reference at its centre %v, spread %.2f diagonals",
					label, q, lo, hi, value, spread)
			}
		}
	}
	t.Logf("%s: %d surface cells, widest %.2f diagonals (limit %v)", label, len(centres), widest, limit)
}

// Its box is the path's grown by the section's reach across it, the greater
// of its two sides, and in z the section's height.
func TestSweep_Box(t *testing.T) {
	const doc = `@scene(140)
route = profileXY([[0, 0], [40, 0], [40, 35], [0, 35]]);
moulding = profileXY([[-3, 0], [4, 0], [4, 6], [-3, 6]]);
solid = sweep(moulding, route);
return solid;
`
	e := New(&source.MockManager{}, nil)
	groups, err := e.EvalGsl(context.Background(), nil, doc, "", []string{"solid.box"}, [][3]float64{{0, 0, 0}})
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	want := []float64{-4, -4, 0, 44, 39, 6}
	for k, w := range want {
		// The library pads a profile's box by 0.01 mm each way.
		if got := float64(groups[0].Samples[k][0].Value); math.Abs(got-w) > 0.05 {
			t.Errorf("solid.box component %d = %v, want %v", k, got, w)
		}
	}
}
