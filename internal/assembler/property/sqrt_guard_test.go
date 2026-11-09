package property

import (
	"testing"

	"gradientcontrol.com/omega/core/program"
)

// TestSqrtGradientAtZero pins the guard on sqrt's derivative. The list form of
// intersectRound spells its fillet term sqrt(sum of squared clamped excesses)
// where the two-input form spells it length(...). Where every clamp is zero --
// anywhere strictly inside a box built from it -- the term is identically zero
// and so is its derivative, but 1/(2*sqrt(0)) is infinite and the product came
// back NaN, poisoning the whole field's gradient.
func TestSqrtGradientAtZero(t *testing.T) {
	d := &derivative{decls: map[string]*program.Operand{}, funcs: map[string][]*program.Function{}}

	// sqrt(max(x - 10, 0)) is identically zero for x well below 10, so its
	// gradient there is zero on every axis.
	clamped := nary(program.Operation_MAX, sub(ref(program.XVarName), num(10)), num(0))
	grad, err := d.gradient(un(program.Operation_SQRT, clamped))
	if err != nil {
		t.Fatalf("gradient: %v", err)
	}
	for _, pt := range []point{{0.3, -1.2, 0.7, 0.4}, {-0.6, 0.9, -0.2, 1.1}} {
		got := evalVec(t, grad, pt)
		if len(got) != 3 {
			t.Fatalf("expected vec3 gradient, got len %d", len(got))
		}
		for k := 0; k < 3; k++ {
			if got[k] != got[k] {
				t.Fatalf("at %+v axis %d: gradient is NaN", pt, k)
			}
			if !closeEnough(got[k], 0) {
				t.Errorf("at %+v axis %d: got %v want 0", pt, k, got[k])
			}
		}
	}

	// Away from zero the guard must not perturb the ordinary derivative:
	// d/dx sqrt(x) = 1/(2*sqrt(x)).
	grad, err = d.gradient(un(program.Operation_SQRT, ref(program.XVarName)))
	if err != nil {
		t.Fatalf("gradient: %v", err)
	}
	for _, pt := range []point{{4, -1.2, 0.7, 0.4}, {0.25, 0.9, -0.2, 1.1}} {
		got := evalVec(t, grad, pt)
		want := 1 / (2 * sqrtOf(pt.x))
		if !closeEnough(got[0], want) {
			t.Errorf("at %+v: got %v want %v", pt, got[0], want)
		}
	}
}

func sqrtOf(v float64) float64 {
	x := v
	for i := 0; i < 60; i++ {
		x = 0.5 * (x + v/x)
	}
	return x
}
