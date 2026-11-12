//go:build cgo

package eval

import (
	"context"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// TestSimpleKeepsThePart evaluates a part with a simpler stand-in, bolt.simple,
// and the same part without one. The stand-in is the viewer's alone: fidget,
// and with it the mesher and export, draw the part as written.
func TestSimpleKeepsThePart(t *testing.T) {
	mgr := source.NewLocalManager("testpkg", "../stdlib")
	e := New(mgr, nil)

	const part = `
bolt = boxCentered([0, 0, 0], [2, 2, 4]) - 0.2;
boltProxy = sphere(2);
plate = box([-10, -10, -3], [10, 10, -2]);
`
	const ret = "return min(bolt, plate);\n"
	points := [][3]float64{{0, 0, 0}, {1.5, 0, 0}, {0, 1.1, 2.1}, {3, 3, 3}, {0.5, -0.5, -2.5}}
	eval := func(doc string) []float64 {
		groups, err := e.EvalGsl(context.Background(), nil, doc, "", []string{"bolt"}, points)
		if err != nil {
			t.Fatalf("eval: %v", err)
		}
		var out []float64
		for _, s := range groups[0].Samples {
			for _, v := range s {
				out = append(out, float64(v.Value))
			}
		}
		return out
	}
	want := eval(part + ret)
	got := eval(part + "bolt.simple = boltProxy;\n" + ret)
	if len(got) != len(want) {
		t.Fatalf("%d values, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("value %d = %v with a stand-in, %v without", i, got[i], want[i])
		}
	}
}
