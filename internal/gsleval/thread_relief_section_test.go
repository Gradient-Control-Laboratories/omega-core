package gsleval

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// TestMaleReliefBoundedTowardTheAxis: a section cut through a relieved screw
// must not show the helix pressed into the whole face. The male relief is zero
// from a form depth under the root inward — at 1.9 on an M6, with the root at
// 2.46 — and still the full profile depth at the root itself, so the cylinder
// renders as it did. Before the bound it read the full form depth at the axis.
func TestMaleReliefBoundedTowardTheAxis(t *testing.T) {
	mgr := source.NewLocalManager("testpkg", "../../stdlib")
	e := New(mgr, nil)
	ctx := context.Background()

	const doc = `
rel = thread.maleRelief(6.0, 1.0, 5.0, -5.0);
return sphere(1.0);
`
	at := func(pt [3]float64) float64 {
		t.Helper()
		groups, err := e.EvalGsl(ctx, nil, doc, "", []string{"rel"}, [][3]float64{pt})
		if err != nil {
			t.Fatalf("eval: %v", err)
		}
		return float64(groups[0].Samples[0][0].Value)
	}
	// On the root flat, z 0.5: the full form depth, as before.
	root := 3.0 - 0.5412659
	if got := at([3]float64{root, 0, 0.5}); math.Abs(got+0.5412659) > 1e-4 {
		t.Errorf("relief on the root = %.5f, want -0.54127: the cylinder must render as it did", got)
	}
	// On a section's face, inside the root: nothing from 1.917 inward.
	for _, r := range []float64{0.0, 0.5, 1.5, 1.9} {
		if got := at([3]float64{r, 0, 0.5}); got != 0 {
			t.Errorf("relief at r %.2f, well inside the root, = %.5f, want 0", r, got)
		}
	}
	// Between, it fades at the full rate: at r 2.2 it is 2.2 - 1.917.
	if got := at([3]float64{2.2, 0, 0.5}); math.Abs(got+(2.2-(3.0-2*0.5412659))) > 1e-4 {
		t.Errorf("relief at r 2.2 = %.5f, want %.5f", got, -(2.2 - (3.0 - 2*0.5412659)))
	}
}
