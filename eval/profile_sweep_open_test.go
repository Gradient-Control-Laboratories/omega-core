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

// sweepOpen(section, points[, curvatures]): the section swept along an open
// chain, x across it (positive on the right of travel) and y up as z, and
// capped flat where the chain ends.

type openSweepCase struct {
	name  string
	chain refChain
	// exact says the section stays inside the chain's tube, so the field is
	// the distance to the solid and not only a bound on it.
	exact bool
}

func openSweepCases() []openSweepCase {
	cases := openCases()
	return []openSweepCase{
		{"S of arcs", cases[1].chain, true},
		{"zigzag, sharp corners", cases[0].chain, false},
	}
}

func (c openSweepCase) doc() string {
	call := strings.Replace(c.chain.gsl(), "openProfileXY(", "sweepOpen(moulding, ", 1)
	return fmt.Sprintf("@scene(140)\nmoulding = %s;\nsolid = %s;\nreturn solid;\n", moulding().gsl(), call)
}

// openEnds are the chain's two ends and the unit tangents of travel there.
func openEnds(c refChain) (first, firstTangent, last, lastTangent [2]float64) {
	n := len(c.points)
	firstTangent, _ = edgeTangents(c.points[0], c.points[1], c.curvature(0))
	_, lastTangent = edgeTangents(c.points[n-2], c.points[n-1], c.curvature(n-2))
	return c.points[0], firstTangent, c.points[n-1], lastTangent
}

// across is q's distance across the line through at along tangent, positive
// on the right of travel.
func across(q, at, tangent [2]float64) float64 {
	return (q[0]-at[0])*tangent[1] - (q[1]-at[1])*tangent[0]
}

// openSwept is the flat-capped sweep from the references: the least of the
// wall's distance where q's nearest point of the extended chain is on the
// chain, and each cap's, the distance to the section standing in the plane
// square to the chain at that end; negative inside the wall beside the chain.
func openSwept(c refChain, path refOpen, section refProfile) func([3]float64) float64 {
	first, t0, last, te := openEnds(c)
	length := path.stations[len(path.stations)-1]
	return func(q [3]float64) float64 {
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
}

func (c openSweepCase) grid(path refOpen, n int) [][3]float64 {
	b := path.chainBox()
	return grid3([3]float64{b[0] - 4, b[1] - 4, 0}, [3]float64{b[2] + 4, b[3] + 4, 6}, 5, n)
}

// sweepOpen is that formula, on a grid round the chain and past its ends.
func TestSweepOpen_IsTheFormula(t *testing.T) {
	for _, c := range openSweepCases() {
		t.Run(c.name, func(t *testing.T) {
			path, section := newRefOpen(t, c.chain), newRefProfile(t, moulding())
			points := c.grid(path, 22)
			compareToReference(t, c.name, points, evalProfileField(t, c.doc(), "solid", points), openSwept(c.chain, path, section))
		})
	}
}

// openSweptCloud samples the solid's surface itself: its wall, for each point
// (x, y) of the section's outline the level set where the extended chain's
// distance is x, traced by marching squares at spacing h at height y and kept
// where its nearest point is on the chain; and its two caps, the section's
// inside on a grid of spacing h standing in each end's plane, apart.
func openSweptCloud(c refChain, path refOpen, section refProfile, h float64) (wall, caps [][3]float64) {
	b := path.chainBox()
	x0, y0 := b[0]-6, b[1]-6
	nx, ny := int((b[2]-b[0]+12)/h)+1, int((b[3]-b[1]+12)/h)+1
	field := make([][]float64, nx)
	stations := make([][]float64, nx)
	for i := range field {
		field[i], stations[i] = make([]float64, ny), make([]float64, ny)
		for j := range field[i] {
			q := [2]float64{x0 + float64(i)*h, y0 + float64(j)*h}
			field[i][j] = path.distance(q)
			_, stations[i][j] = path.nearest(q)
		}
	}
	length := path.stations[len(path.stations)-1]
	keep := func(q [2]float64, station, height float64) {
		if station >= 0 && station <= length {
			wall = append(wall, [3]float64{q[0], q[1], height})
		}
	}
	for _, s := range section.boundarySamples(h) {
		level, height := s[0], s[1]
		for i := 0; i < nx; i++ {
			for j := 0; j < ny; j++ {
				a := field[i][j] - level
				if i+1 < nx {
					if e := field[i+1][j] - level; (a < 0) != (e < 0) {
						f := a / (a - e)
						keep([2]float64{x0 + (float64(i)+f)*h, y0 + float64(j)*h}, stations[i][j]+f*(stations[i+1][j]-stations[i][j]), height)
					}
				}
				if j+1 < ny {
					if e := field[i][j+1] - level; (a < 0) != (e < 0) {
						f := a / (a - e)
						keep([2]float64{x0 + float64(i)*h, y0 + (float64(j)+f)*h}, stations[i][j]+f*(stations[i][j+1]-stations[i][j]), height)
					}
				}
			}
		}
	}
	first, t0, last, te := openEnds(c)
	sb := section.box()
	for x := sb[0]; x <= sb[2]; x += h {
		for y := sb[1]; y <= sb[3]; y += h {
			if section.distance([2]float64{x, y}) > 0 {
				continue
			}
			for _, end := range [][2][2]float64{{first, t0}, {last, te}} {
				at, tangent := end[0], end[1]
				caps = append(caps, [3]float64{at[0] + x*tangent[1], at[1] - x*tangent[0], y})
			}
		}
	}
	return wall, caps
}

// Against the solid's surface sampled on its own, where the documented
// condition holds, at points whose nearest point of the extended chain is on
// the chain and at points beyond a cap that the cap is nearest: never more
// than its distance, and equal to it, to within the sampling, where the
// section stays inside the chain's tube. Beyond a cap where another part of
// the sweep is nearer than the cap, the field reads that cap's distance and
// is over: logged, not asserted (sweepOpen's doc comment).
func TestSweepOpen_AgainstTheSweptSolid(t *testing.T) {
	const h = 0.2
	nearest := func(q [3]float64, cloud [][3]float64) float64 {
		best := math.Inf(1)
		for _, s := range cloud {
			best = math.Min(best, math.Hypot(math.Hypot(q[0]-s[0], q[1]-s[1]), q[2]-s[2]))
		}
		return best
	}
	for _, c := range openSweepCases() {
		t.Run(c.name, func(t *testing.T) {
			path, section := newRefOpen(t, c.chain), newRefProfile(t, moulding())
			wall, caps := openSweptCloud(c.chain, path, section, h)
			length := path.stations[len(path.stations)-1]
			points := c.grid(path, 10)
			got := evalProfileField(t, c.doc(), "solid", points)
			worstOver, worstUnder, overAt, underAt := 0.0, 0.0, [3]float64{}, [3]float64{}
			outside, outsideOver, outsideAt := 0, 0.0, [3]float64{}
			for i, q := range points {
				toWall, toCaps := nearest(q, wall), nearest(q, caps)
				diff := math.Abs(got[i]) - math.Min(toWall, toCaps)
				if _, station := path.nearest([2]float64{q[0], q[1]}); (station < 0 || station > length) && toCaps > toWall+h {
					// Beyond a cap, with another part of the sweep nearer.
					outside++
					if diff > outsideOver {
						outsideOver, outsideAt = diff, q
					}
					continue
				}
				if diff > worstOver {
					worstOver, overAt = diff, q
				}
				if -diff > worstUnder {
					worstUnder, underAt = -diff, q
				}
			}
			t.Logf("%s: %d wall and %d cap points; over the surface's distance by at most %.3g mm (at %v), under it by at most %.3g mm (at %v)",
				c.name, len(wall), len(caps), worstOver, overAt, worstUnder, underAt)
			t.Logf("%s: %d points beyond a cap with another part of the sweep nearer, over by at most %.3g mm (at %v)",
				c.name, outside, outsideOver, outsideAt)
			if worstOver > h {
				t.Errorf("%s: the field exceeds the distance to the swept solid by %.3g mm at %v", c.name, worstOver, overAt)
			}
			if c.exact && worstUnder > h {
				t.Errorf("%s: inside the tube the field falls short of the distance by %.3g mm at %v", c.name, worstUnder, underAt)
			}
		})
	}
}

// Its surface cells, on the wall and on both caps, bound it soundly; certify
// proves the space above it and beyond its first cap empty; bounds, when it
// proves a box, holds the solid. Cells are allowed sweepCellSpread diagonals
// and certify sweepCertifyBudget evaluations, as sweep's are. On a chain with
// sharp corners only soundness is asserted and the widest cell logged: near
// a corner's bisector the pools' sign test is undecided over a cell, and the
// cell bounds both signs (9.00 diagonals on the zigzag, 2026-11-18).
func TestSweepOpen_Proves(t *testing.T) {
	for _, c := range openSweepCases() {
		t.Run(c.name, func(t *testing.T) {
			path, section := newRefOpen(t, c.chain), newRefProfile(t, moulding())
			want := openSwept(c.chain, path, section)
			var centres [][3]float64
			length := path.stations[len(path.stations)-1]
			for _, s := range section.boundarySamples(1.5) {
				for _, q := range path.samples(4, 0, 0) {
					// Offset each chain point across by the section's x, along
					// the extended chain's own gradient there.
					e := 1e-4
					gx := (path.distance([2]float64{q[0] + e, q[1]}) - path.distance([2]float64{q[0] - e, q[1]})) / (2 * e)
					gy := (path.distance([2]float64{q[0], q[1] + e}) - path.distance([2]float64{q[0], q[1] - e})) / (2 * e)
					l := math.Hypot(gx, gy)
					centre := [2]float64{q[0] + s[0]*gx/l, q[1] + s[0]*gy/l}
					if _, station := path.nearest(centre); station >= 0 && station <= length {
						centres = append(centres, [3]float64{centre[0], centre[1], s[1]})
					}
				}
			}
			first, t0, last, te := openEnds(c.chain)
			for _, s := range [][2]float64{{0, 3}, {2, 2}, {-2, 4}, {3, 1}} {
				for _, end := range [][2][2]float64{{first, t0}, {last, te}} {
					at, tangent := end[0], end[1]
					centres = append(centres, [3]float64{at[0] + s[0]*tangent[1], at[1] - s[0]*tangent[0], s[1]})
				}
			}
			limit := sweepCellSpread
			if !c.exact {
				limit = math.Inf(1)
			}
			checkSurfaceCellsWithin(t, c.name, c.doc(), centres, limit, want)

			b := path.chainBox()
			e := New(&source.MockManager{}, nil)
			e.SetCertifyBudget(sweepCertifyBudget)
			regions := map[string][6]float64{
				"above it":             {b[0] - 10, b[1] - 10, 6.25, b[2] + 10, b[3] + 10, 12},
				"beyond its first cap": {first[0] - 10, first[1] - 10, -2, first[0] - 0.25, first[1] + 10, 8},
			}
			for where, region := range regions {
				r, err := e.Certify(context.Background(), nil, c.doc(), "", "solid", region, 0, 0)
				if err != nil {
					t.Fatalf("certify %s: %v", where, err)
				}
				if !r.Proven {
					t.Errorf("%s: want proven, got proven=%v disproven=%v in %d evals", where, r.Proven, r.Disproven, r.Evals)
				}
				t.Logf("%s: certify proves the space %s in %d evals", c.name, where, r.Evals)
			}
			en, err := e.Enclose(context.Background(), nil, c.doc(), "", "solid",
				[6]float64{b[0] - 15, b[1] - 15, -10, b[2] + 15, b[3] + 15, 16}, 0)
			if err != nil {
				t.Fatalf("enclose: %v", err)
			}
			// Enclose's budget is fixed (backlog M16): logged, not asserted.
			if !en.Proven {
				t.Logf("%s: bounds does not prove the box within its fixed budget: %q", c.name, en.Reason)
				return
			}
			solid := [6]float64{b[0] - 4, b[1] - 4, 0, b[2] + 4, b[3] + 4, 6}
			for _, q := range c.grid(path, 12) {
				if want(q) < 0 {
					for k := 0; k < 3; k++ {
						if q[k] < en.BboxMin[k]-0.05 || q[k] > en.BboxMax[k]+0.05 {
							t.Errorf("proven box %v to %v misses the solid at %v", en.BboxMin, en.BboxMax, q)
						}
					}
				}
			}
			t.Logf("%s: proven box %v to %v (the solid lies within %v)", c.name, en.BboxMin, en.BboxMax, solid)
		})
	}
}

// Its frame: .u the chain's openProfileXY .u, the distance along it, and .v
// the section's .u read at that profile's distance, as sweep reads it.
func TestSweepOpen_Frame(t *testing.T) {
	c := openSweepCases()[0]
	path := newRefOpen(t, c.chain)
	points := c.grid(path, 6)
	doc := strings.Replace(c.doc(), "return solid;", "route = "+c.chain.gsl()+";\nreturn solid;", 1)
	e := New(&source.MockManager{}, nil)
	groups, err := e.EvalGsl(context.Background(), nil, doc, "",
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

// Its box is the chain's, without its rays, grown by the section's reach
// across it, the greater of its two sides, and in z the section's height; an
// arc chain's box is grown for the bulge as a closed arc profile's is.
func TestSweepOpen_Box(t *testing.T) {
	cases := []struct {
		name, path string
		want       []float64
	}{
		{"lines", "[[0, 0], [40, 0], [40, 35]]", []float64{-4, -4, 0, 44, 39, 6}},
		// An L whose bend is a quarter arc of radius 10, which stays within
		// its points' box: the chain's exact box is 40 by 35 (C7).
		{"arcs", "[[0, 0], [30, 0], [40, 10], [40, 35]], [0, 0.1, 0]", []float64{-4, -4, 0, 44, 39, 6}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			doc := fmt.Sprintf("@scene(140)\nmoulding = profileXY([[-3, 0], [4, 0], [4, 6], [-3, 6]]);\nsolid = sweepOpen(moulding, %s);\nreturn solid;\n", c.path)
			e := New(&source.MockManager{}, nil)
			groups, err := e.EvalGsl(context.Background(), nil, doc, "", []string{"solid.box"}, [][3]float64{{0, 0, 0}})
			if err != nil {
				t.Fatalf("eval: %v", err)
			}
			for k, w := range c.want {
				// The library pads a profile's box by 0.01 mm each way.
				if got := float64(groups[0].Samples[k][0].Value); math.Abs(got-w) > 0.05 {
					t.Errorf("solid.box component %d = %v, want %v", k, got, w)
				}
			}
		})
	}
}
