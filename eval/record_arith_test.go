//go:build cgo

package eval

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// TestRecordArithmeticKeepsRows evaluates arithmetic on a material record, a
// list of three vec4 rows, where a folded part of it reaches the build with its
// values but no type: a constant difference of two records, and a record the
// inliner hoisted into a temporary because two places use it. Read as one flat
// vector, such a record indexes to a single number per row, and every row came
// out as that number repeated.
func TestRecordArithmeticKeepsRows(t *testing.T) {
	mgr := source.NewLocalManager("testpkg", "../stdlib")
	e := New(mgr, nil)

	const doc = `
import "material";
pm = material.polishedMetal;
s0 = pm[0];
lit = [[0.08, s0.y, s0.z, s0.w], pm[1], pm[2], pm[3]];
mixed = pm + (lit - pm) * p.x;
twice(a, b) {
  return [a[0] + b[0] * p.x, a[1], a[2], a[3]];
}
flakeTwice(a) {
  flake = material.withRoughness(a, 0.08);
  return twice(flake, flake);
}
hoisted = flakeTwice(material.polishedMetal);
return sphere(1);
`
	groups, err := e.EvalGsl(context.Background(), nil, doc, "", []string{"mixed", "hoisted"}, [][3]float64{{0.5, 0, 0}})
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	want := map[string][16]float64{
		// polishedMetal's surface row blended half way to roughness 0.08.
		"mixed": {0.115, 1, 1.5, 0},
		// The flake's row plus half of itself: roughness 0.12, metallic 1.5,
		// ior 2.25.
		"hoisted": {0.12, 1.5, 2.25, 0},
	}
	for i, name := range []string{"mixed", "hoisted"} {
		got := groups[i].Samples
		if len(got) != 16 {
			t.Fatalf("%s: %d components, want 16", name, len(got))
		}
		for k, w := range want[name] {
			if v := float64(got[k][0].Value); math.Abs(v-w) > 1e-5 {
				t.Errorf("%s component %d = %.5f, want %.5f", name, k, v, w)
			}
		}
	}
}
