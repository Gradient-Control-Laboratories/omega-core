//go:build cgo

package eval

import (
	"context"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// Two discs drawn by a reduce whose step reads the point and its own index: the
// second disc's radius is 2 + 1. The gradient (.dp) is built by stepping the
// reduce again, which has to hand the step its index as the program does;
// profileXY's pools are written this way, so a profile's .dp depends on it.
const reduceIndexDoc = `@scene(40)
discs = reduce([nothing, 0], (best, c, index) {
  return [min(best.x, length(p.xy - c) - (2 + index)), 0];
}, [[0, 0], [10, 0]]).x;
return max(discs, abs(p.z) - 1);
`

func TestEvalGsl_GradientOfAReduceThatReadsItsIndex(t *testing.T) {
	e := New(&source.MockManager{}, nil)
	groups, err := e.EvalGsl(context.Background(), nil, reduceIndexDoc, "", []string{"discs", "discs.dp"}, [][3]float64{{13.5, 0, 0}, {0, 2.5, 0}})
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	// 0.5 outside the second disc (radius 3 about x = 10), and the first (radius 2 about the origin).
	for i, want := range []float64{0.5, 0.5} {
		if got := float64(groups[0].Samples[0][i].Value); math.Abs(got-want) > 1e-4 {
			t.Errorf("discs at point %d = %v, want %v", i, got, want)
		}
	}
	// The gradient points away from each disc's centre.
	wantDp := [][2]float64{{1, 0}, {0, 1}}
	for i, w := range wantDp {
		gx, gy := float64(groups[1].Samples[0][i].Value), float64(groups[1].Samples[1][i].Value)
		if math.Abs(gx-w[0]) > 1e-4 || math.Abs(gy-w[1]) > 1e-4 {
			t.Errorf("discs.dp at point %d = (%v, %v), want (%v, %v)", i, gx, gy, w[0], w[1])
		}
	}
}

// profileXY is the library function that reads its index in a reduce; its .dp
// compiles and is the outward normal on an edge.
func TestEvalGsl_GradientOfAProfile(t *testing.T) {
	e := New(&source.MockManager{}, nil)
	doc := "@scene(60)\ntri = profileXY([[0, 0], [20, 0], [0, 20]]);\nreturn tri;\n"
	groups, err := e.EvalGsl(context.Background(), nil, doc, "", []string{"tri.dp"}, [][3]float64{{5, -1, 0}})
	if err != nil {
		t.Fatalf("eval: %v", err)
	}
	gx, gy := float64(groups[0].Samples[0][0].Value), float64(groups[0].Samples[1][0].Value)
	if math.Abs(gx) > 1e-4 || math.Abs(gy+1) > 1e-4 {
		t.Errorf("tri.dp below its bottom edge = (%v, %v), want (0, -1)", gx, gy)
	}
}
