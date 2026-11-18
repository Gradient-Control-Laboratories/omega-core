//go:build cgo

package eval

import (
	"context"
	"fmt"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// What lets `gsl eval` prove things about a profile: interval bounds over
// cells that straddle its surface are sound and tight, and certify proves the
// space just off the part is empty.

// cellHalf is the half-width of a surface cell, in mm.
const cellHalf = 0.25

// maxCellSpread is how wide a surface cell's interval may be, as a multiple of
// the cell's diagonal: the most a distance can change across the cell. A pool
// left bounded by ±nothing (1e5) is wider by orders of magnitude.
const maxCellSpread = 4.0

// intervalOver bounds the named field of doc over each box.
func intervalOver(t *testing.T, doc, name string, boxes [][6]float32) []IntervalSample {
	t.Helper()
	ctx := context.Background()
	if err := startQuery(ctx); err != nil {
		t.Fatal(err)
	}
	defer endQuery()
	e := New(&source.MockManager{}, nil)
	prog, err := e.prepareQuery(ctx, nil, doc, "", []string{name})
	if err != nil {
		t.Fatalf("assemble %s: %v", name, err)
	}
	results, err := EvalInterval(prog, boxes)
	if err != nil {
		t.Fatalf("interval %s: %v", name, err)
	}
	return results[0]
}

// cellAround is the cube of half-width cellHalf about q.
func cellAround(q [3]float64) [6]float32 {
	return [6]float32{
		float32(q[0] - cellHalf), float32(q[1] - cellHalf), float32(q[2] - cellHalf),
		float32(q[0] + cellHalf), float32(q[1] + cellHalf), float32(q[2] + cellHalf),
	}
}

// checkSurfaceCells bounds the field over a cell about each point and fails if
// a bound misses the reference value at the cell's centre or is wider than
// maxCellSpread diagonals. It logs the widest spread either way.
func checkSurfaceCells(t *testing.T, label, doc string, centres [][3]float64, want func([3]float64) float64) {
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
		value := want(q)
		if value < lo-profileTolerance || value > hi+profileTolerance || spread > maxCellSpread {
			if failures++; failures <= 3 {
				t.Errorf("%s: cell about %v bounds [%v, %v], reference at its centre %v, spread %.2f diagonals",
					label, q, lo, hi, value, spread)
			}
		}
	}
	t.Logf("%s: %d surface cells, widest %.2f diagonals", label, len(centres), widest)
}

// Cells straddling each profile's boundary bound it soundly and tightly.
func TestProfileXY_SurfaceCellsAreTight(t *testing.T) {
	for _, c := range profileCases(t) {
		t.Run(c.name, func(t *testing.T) {
			doc := "@scene(120)\nsolid = " + c.expr + ";\nreturn solid;\n"
			checkSurfaceCells(t, c.name, doc, flat(c.ref.boundarySamples(1)), func(q [3]float64) float64 {
				return c.ref.distance([2]float64{q[0], q[1]})
			})
		})
	}
}

// rimCentres are points on an extrusion's boundary: round its loop at the two
// rims and half way up, and across its top cap on a grid, kept to where the
// top is.
func rimCentres(ref refProfile, low, high float64) [][3]float64 {
	var out [][3]float64
	for _, q := range ref.boundarySamples(1.5) {
		for _, z := range []float64{low, (low + high) / 2, high} {
			out = append(out, [3]float64{q[0], q[1], z})
		}
	}
	for _, q := range gridAround(ref.box(), 0, 16) {
		if ref.distance(q) < 0 {
			out = append(out, [3]float64{q[0], q[1], high})
		}
	}
	return out
}

// An extrusion's rims and caps bound soundly and tightly, sharp and rounded.
// The rounded rim is where #45 met a pool bounded by ±nothing: intersectRound's
// sharp - max(sharp + radius, 0) could not prove above the cap.
func TestExtrude_SurfaceCellsAreTight(t *testing.T) {
	const low, high = 0.0, 12.0
	for _, c := range profileCases(t) {
		t.Run(c.name, func(t *testing.T) {
			centres := rimCentres(c.ref, low, high)
			sharp := fmt.Sprintf("@scene(120)\nsolid = extrude(%s, %v, %v);\nreturn solid;\n", c.expr, low, high)
			checkSurfaceCells(t, c.name+" sharp", sharp, centres, extrudedReference(c.ref, low, high))
			// A rounded rim lies inside the sharp one, so the sharp distance is a
			// lower bound on it; check the bound against the field itself.
			rounded := fmt.Sprintf("@scene(120)\nsolid = extrude(%s, %v, %v, intersectRound(2));\nreturn solid;\n", c.expr, low, high)
			values := evalProfileField(t, rounded, "solid", centres)
			index := make(map[[3]float64]float64, len(centres))
			for i, q := range centres {
				index[q] = values[i]
			}
			checkSurfaceCells(t, c.name+" rounded", rounded, centres, func(q [3]float64) float64 { return index[q] })
		})
	}
}

// certify proves the space just off an extrusion empty, above its cap and
// beside it, with sharp rims and rounded.
func TestExtrude_CertifiesClearance(t *testing.T) {
	const low, high, gap = 0.0, 12.0, 0.25
	for _, c := range profileCases(t) {
		t.Run(c.name, func(t *testing.T) {
			box := c.ref.box()
			regions := map[string][6]float64{
				"above the cap": {box[0] - 5, box[1] - 5, high + gap, box[2] + 5, box[3] + 5, high + 5},
				"beside it":     {box[2] + gap, box[1] - 5, low - 5, box[2] + 5, box[3] + 5, high + 5},
			}
			for _, blend := range []string{"", ", intersectRound(2)"} {
				doc := fmt.Sprintf("@scene(120)\nsolid = extrude(%s, %v, %v%s);\nreturn solid;\n", c.expr, low, high, blend)
				for where, region := range regions {
					e := New(&source.MockManager{}, nil)
					r, err := e.Certify(context.Background(), nil, doc, "", "solid", region, 0, 0)
					if err != nil {
						t.Fatalf("certify %s%s: %v", where, blend, err)
					}
					if !r.Proven {
						t.Errorf("%s%s, %s: want proven, got proven=%v disproven=%v indeterminate=%v [%v, %v] in %d evals",
							c.name, blend, where, r.Proven, r.Disproven, r.Indeterminate, r.RegionLower, r.RegionUpper, r.Evals)
					}
				}
			}
		})
	}
}
