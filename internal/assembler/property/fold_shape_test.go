package property

import (
	"fmt"
	"testing"

	"gradientcontrol.com/omega/core/program"
)

// nodeCount is how many operand nodes an expression is written with — the
// measure that matters downstream, where nothing shares a subexpression and a
// value is re-emitted once per textual mention.
func nodeCount(op *program.Operand) int {
	if op == nil {
		return 0
	}
	n := 1
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		for _, o := range t.Expression.Operands {
			n += nodeCount(o)
		}
	case *program.Operand_FnCall:
		for _, a := range t.FnCall.Arguments {
			n += nodeCount(a)
		}
	}
	return n
}

// rampMin is min over n lines of distinct slope, so the argmin — and therefore
// the derivative — changes with t and a finite difference can check it.
func rampMin(n int) *program.Operand {
	arms := make([]*program.Operand, n)
	for i := 0; i < n; i++ {
		arms[i] = add(mul(num(float64(i+1)), ref(program.TimeVarName)), num(float64(n-i)))
	}
	return nary(program.Operation_MIN, arms...)
}

// TestFoldMinDerivativeIsCorrect pins the VALUE of an n-ary min's derivative
// against a finite difference. Reassociating the fold must not move it: IEEE754
// min is exactly associative, so only the textual shape and the tie-break
// change, and none of these points is a tie.
func TestFoldMinDerivativeIsCorrect(t *testing.T) {
	d := newDerivative()
	for _, n := range []int{2, 3, 5, 8, 16} {
		t.Run(fmt.Sprintf("n=%d", n), func(t *testing.T) {
			target := rampMin(n)
			deriv, err := d.temporal(target)
			if err != nil {
				t.Fatalf("temporal: %v", err)
			}
			for _, pt := range []point{{time: -1.7}, {time: -0.4}, {time: 0.35}, {time: 1.9}} {
				analytic := evalScalar(t, deriv, pt)
				numeric := fdTime(t, target, pt)
				if !closeEnough(analytic, numeric) {
					t.Errorf("at %+v: analytic=%v finite-diff=%v", pt, analytic, numeric)
				}
			}
		})
	}
}

// TestFoldMinDerivativeGrowth pins the SHAPE. minDual mentions both operands'
// primals — once in the comparison and once in the min — so a left-leaning
// chain mentions an accumulator of size k again at step k+1 and the tangent is
// Θ(n²). Folded over a balanced tree it is Θ(n log n).
//
// The bound is a ratio between two sizes of the same expression, so it does not
// become a golden: quadratic gives (32/8)² = 16x, n log n gives 4·(5/3) ≈ 6.7x.
func TestFoldMinDerivativeGrowth(t *testing.T) {
	d := newDerivative()
	const small, large = 8, 32

	dSmall, err := d.temporal(rampMin(small))
	if err != nil {
		t.Fatalf("temporal(%d): %v", small, err)
	}
	dLarge, err := d.temporal(rampMin(large))
	if err != nil {
		t.Fatalf("temporal(%d): %v", large, err)
	}
	nSmall, nLarge := nodeCount(dSmall), nodeCount(dLarge)
	t.Logf("n=%d: %d nodes", small, nSmall)
	t.Logf("n=%d: %d nodes", large, nLarge)

	const maxRatio = 10
	if nLarge > maxRatio*nSmall {
		t.Errorf("quadrupling the arms of a min multiplied the derivative's nodes by %.1fx (%d -> %d).\n"+
			"A left-leaning fold mentions its accumulator again at every step, which is quadratic; "+
			"a balanced one is n log n.", float64(nLarge)/float64(nSmall), nSmall, nLarge)
	}
}
