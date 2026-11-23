//go:build cgo

package eval

import (
	"context"
	"math"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/source"
)

// A one-element list is a list to the type check, but fidget builds it as one
// tree that reports a number; indexing it at 0 still reads its element, on the
// CPU as in WGSL, and past 0 is an error.
func TestOneElementListIndexes(t *testing.T) {
	for _, c := range []struct {
		name, doc string
		at        [3]float64
		want      float64
	}{
		// Read at run time inside a reduce: 2 (0.1 x + 2) at x = 3.
		{"in a reduce", "f(points, curvatures) {\n" +
			"  count = reduce([0, 0], (sum, point) { return sum + 1; }, points).x;\n" +
			"  return reduce(0, (acc, start, index) { return acc + curvatures[0] * p.x + count; }, points);\n" +
			"}\n" +
			"solid = f([[0, 0], [10, 0]], [0.1]);\nreturn solid;\n", [3]float64{3, 0, 0}, 4.6},
		// Arithmetic on it first: |p| - 1 + 2 x at x = 2.
		{"after arithmetic", "l = [p.x];\nm = l * 2;\nsolid = sphere(1) + m[0];\nreturn solid;\n", [3]float64{2, 0, 0}, 5},
	} {
		t.Run(c.name, func(t *testing.T) {
			got := evalProfileField(t, c.doc, "solid", [][3]float64{c.at})
			if math.Abs(got[0]-c.want) > 1e-5 {
				t.Errorf("solid at %v = %v, want %v", c.at, got[0], c.want)
			}
		})
	}
	e := New(&source.MockManager{}, nil)
	if _, err := e.EvalGsl(context.Background(), nil, "l = [p.x];\nsolid = l[1];\nreturn solid;\n", "", []string{"solid"},
		[][3]float64{{1, 0, 0}}); err == nil {
		t.Error("l[1] of a one-element list evaluated; want an error")
	}
}
