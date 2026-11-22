//go:build cgo

package eval

import (
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// A reduce over a list of functions can take the element before this one as a
// fourth callback parameter, the first step being handed its own: for tags 1,
// 2 and 3, acc + 10 * prev + this runs 0 + 10 + 1 = 11, 11 + 10 + 2 = 23 and
// 23 + 20 + 3 = 46.
func TestFold_PreviousElement(t *testing.T) {
	doc := `
tag(k) {
  t(want) {
    return k;
  }
  return t;
}
stack(features) {
  return reduce((acc, f, i, prev) { return acc + prev(0.0) * 10.0 + f(0.0); }, features, 0.0);
}
solid = stack([tag(1.0), tag(2.0), tag(3.0)]);
return solid;
`
	e := New(&source.MockManager{}, nil)
	groups, err := e.EvalGsl(t.Context(), nil, doc, "", []string{"solid"}, [][3]float64{{0, 0, 0}})
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	if got := groups[0].Samples[0][0].Value; got != 46 {
		t.Errorf("the fold reads %v, want 46", got)
	}
}
