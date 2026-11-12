//go:build cgo

package eval_test

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/eval"
	"gradientcontrol.com/omega/core/internal/source"
)

// The frame forms of machined.turned and machined.knurled read the part's own
// frame, [bar.u, bar.v], in place of an axis kept in step with the geometry by
// hand (doc/material_composition.md, "Datums are two different things"). On a
// capped cylinder the one call has to agree with the axis forms: with turned()
// on the side, where v runs along the axis, and with faced() on the caps,
// where v runs out from it.

const machinedFrameDoc = `
steel, steelColor = metal.steel();
bar = cylinderCapped([0, 0, -6], [0, 0, 6], 4);
frameMat, frameColor, frameHeight, frameSlope = machined.turned(steel, steelColor, [bar.u, bar.v], 0.15, 0.8);
axisMat, axisColor, axisHeight, axisSlope = machined.turned(steel, steelColor, [0, 0, 0], [0, 0, 1], 0.15, 0.8);
facedMat, facedColor, facedHeight, facedSlope = machined.faced(steel, steelColor, [0, 0, 0], [0, 0, 1], 0.15, 0.8);
knurlMat, knurlColor, knurlHeight, knurlSlope = machined.knurled(steel, steelColor, [bar.u, bar.v], 8, 0.8, 1, 1);
straightMat, straightColor, straightHeight, straightSlope = machined.knurled(steel, steelColor, [bar.u, bar.v], 8, 0.8, 0, 1);
return bar;
`

// materialReads are the rows of a material record the finishes set, read back
// through the material library's accessors.
func materialReads(name string) []string {
	return []string{
		"material.lay(" + name + ")",
		"material.anisotropy(" + name + ")",
		"material.roughness(" + name + ")",
		"material.displacement(" + name + ")",
	}
}

func TestTurnedFrameFormAgreesWithTheAxisForms(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	side := [][3]float64{{4, 0, 0}, {0, 4, 0}, {2.828427, 2.828427, 1.5}, {-4, 0, -3}}
	caps := [][3]float64{{1, 0, 6}, {2, 1, 6}, {1.5, -2, -6}}

	for _, tc := range []struct {
		where string
		pts   [][3]float64
		want  string
	}{
		{"the side", side, "axisMat"},
		{"the caps", caps, "facedMat"},
	} {
		queries := append(materialReads("frameMat"), materialReads(tc.want)...)
		results, err := e.EvalGsl(context.Background(), nil, machinedFrameDoc, "", queries, tc.pts)
		if err != nil {
			t.Fatalf("eval on %s: %v", tc.where, err)
		}
		if len(results) != len(queries) {
			t.Fatalf("%d results for %d queries", len(results), len(queries))
		}
		for q := range materialReads("frameMat") {
			got, want := results[q], results[q+4]
			if len(got.Samples) != len(want.Samples) {
				t.Fatalf("%s on %s: %d components, want %d", queries[q], tc.where, len(got.Samples), len(want.Samples))
			}
			for c := range got.Samples {
				for i, pt := range tc.pts {
					g, w := float64(got.Samples[c][i].Value), float64(want.Samples[c][i].Value)
					if math.Abs(g-w) > 1e-4 {
						t.Errorf("%s at %v on %s: %g, %s gives %g", queries[q], pt, tc.where, g, queries[q+4], w)
					}
				}
			}
		}
	}
}

// The knurl counts a whole number of teeth round the part, so its height runs
// on across u's seam, where u's value jumps by a turn, and repeats every pitch
// along v.
func TestKnurledFrameFormRunsOnAcrossTheSeam(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	// u's seam on a cylinder about z lies along -x. Points a hair either side
	// of it, at a few heights, and a straight knurl a pitch apart along z.
	pts := [][3]float64{
		{-4, 0.001, 2}, {-4, -0.001, 2},
		{-4, 0.001, -1.3}, {-4, -0.001, -1.3},
		{4, 0, 0.1}, {4, 0, 0.9},
	}
	results, err := e.EvalGsl(context.Background(), nil, machinedFrameDoc, "", []string{"knurlHeight", "straightHeight"}, pts)
	if err != nil {
		t.Fatal(err)
	}
	knurl, straight := results[0].Samples[0], results[1].Samples[0]
	for _, pair := range [][2]int{{0, 1}, {2, 3}} {
		a, b := float64(knurl[pair[0]].Value), float64(knurl[pair[1]].Value)
		if math.Abs(a-b) > 5e-3 {
			t.Errorf("the knurl breaks at the seam: %g either side of it at %v, %g at %v", a, pts[pair[0]], b, pts[pair[1]])
		}
	}
	if a, b := float64(straight[4].Value), float64(straight[5].Value); math.Abs(a-b) > 1e-4 {
		t.Errorf("a straight knurl should repeat every pitch along v: %g at z=0.1, %g at z=0.9", a, b)
	}
}
