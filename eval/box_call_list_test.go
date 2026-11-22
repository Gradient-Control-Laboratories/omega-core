//go:build cgo

package eval

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// A declared box that reduces over a list a function call returned. The box
// pass lists a call's elements by inlining it, but it used to stop short in
// any domain that intercepts calls (the box domain is one) even for a call
// the domain does not handle, so the reduce was handed the whole list as one
// element: abs(k) of [0, 0, 0] is a list, and the box failed to assemble.
// profileXY's arc box reads its curvatures this way.
func TestBox_ReducesOverAListACallReturned(t *testing.T) {
	cases := map[string]string{
		"a list literal": `zeros(points) {
  return [0, 0, 0];
}
pts = [[0, 0], [20, 0], [0, 20]];
solid = profileXY(pts, zeros(pts));
return solid;
`,
		"a map": `every(points, v) {
  return map((point) { return v; }, points);
}
pts = [[0, 0], [20, 0], [0, 20]];
solid = profileXY(pts, every(pts, 0));
return solid;
`,
		"a user box": `zeros(points) {
  return [0, 0, 0];
}
widest(ks) {
  form = length(p.xy) - 5;
  form.box = [[neg(9), neg(9), neg(nothing)], [reduce(0, (m, k) { return max(m, abs(k)); }, ks) + 9, 9, nothing]];
  return form;
}
solid = widest(zeros([[0, 0], [20, 0], [0, 20]]));
return solid;
`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			e := New(&source.MockManager{}, nil)
			groups, err := e.EvalGsl(context.Background(), nil, "@scene(60)\n"+body, "", []string{"solid"}, [][3]float64{{5, 5, 0}, {30, 30, 0}})
			if err != nil {
				t.Fatalf("eval: %v", err)
			}
			want := []float64{math.Hypot(5, 5) - 5, math.Hypot(30, 30) - 5}
			if name != "a user box" {
				// The triangle (0,0), (20,0), (0,20): 5 inside from its two legs at
				// (5, 5); at (30, 30), 30*sqrt(2) - 10*sqrt(2) from its hypotenuse.
				want = []float64{-5, 20 * math.Sqrt2}
			}
			for i, w := range want {
				if got := float64(groups[0].Samples[0][i].Value); math.Abs(got-w) > 1e-4 {
					t.Errorf("solid at point %d = %v, want %v", i, got, w)
				}
			}
		})
	}
}
