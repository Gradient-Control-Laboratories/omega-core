//go:build cgo

package eval

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// A declared box that reads names assigned after its shape. The box is applied
// as a crop on the shape's own assignment, which came before those names; a
// function whose body was spliced into its caller in order (any document
// function called with constants, or profileXY with a blend made from a
// slider) then read them before they were defined.
func TestBox_ReadsNamesAssignedAfterItsShape(t *testing.T) {
	cases := map[string]string{
		"two names": `two(a) {
  return a, a + 1;
}
f(a) {
  form = length(p.xy - [a, 0]) - 3;
  lo, hi = two(a);
  form.box = [[lo - 9, neg(9), neg(nothing)], [hi + 9, 9, nothing]];
  return form;
}
solid = f(1);
return solid;
`,
		"one name each": `f(a) {
  form = length(p.xy - [a, 0]) - 3;
  lo = a - 9;
  hi = a + 9;
  form.box = [[lo, neg(9), neg(nothing)], [hi, 9, nothing]];
  return form;
}
solid = f(1);
return solid;
`,
		"with a blend from a slider": `@range(0.5, 3) r = 1.0;
two(a) {
  return a, a + 1;
}
f(a, blend) {
  form = blend([length(p.xy - [a, 0]) - 3, neg(1000)]);
  lo, hi = two(a);
  form.box = [[lo - 9, neg(9), neg(nothing)], [hi + 9, 9, nothing]];
  return form;
}
solid = f(1, intersectRound(r));
return solid;
`,
	}
	for name, body := range cases {
		t.Run(name, func(t *testing.T) {
			e := New(&source.MockManager{}, nil)
			groups, err := e.EvalGsl(context.Background(), nil, "@scene(60)\n"+body, "", []string{"solid"}, [][3]float64{{1, 0, 0}, {5, 0, 0}})
			if err != nil {
				t.Fatalf("eval: %v", err)
			}
			for i, want := range []float64{-3, 1} {
				if got := float64(groups[0].Samples[0][i].Value); math.Abs(got-want) > 1e-4 {
					t.Errorf("solid at point %d = %v, want %v", i, got, want)
				}
			}
		})
	}
}

// A shape read between its own assignment and a name its box reads after it
// cannot be moved past that name; it is refused with a message saying so.
func TestBox_RefusesAShapeReadBeforeTheNamesItsBoxReads(t *testing.T) {
	const body = `f(a) {
  form = length(p.xy - [a, 0]) - 3;
  lo = a - 9 + 0 * form;
  form.box = [[lo, neg(9), neg(nothing)], [a + 9, 9, nothing]];
  return form;
}
solid = f(1);
return solid;
`
	e := New(&source.MockManager{}, nil)
	_, err := e.EvalGsl(context.Background(), nil, "@scene(60)\n"+body, "", []string{"solid"}, [][3]float64{{1, 0, 0}})
	if err == nil {
		t.Fatal("want an error: lo reads form, and form's box reads lo")
	}
	t.Logf("refused: %v", err)
}
