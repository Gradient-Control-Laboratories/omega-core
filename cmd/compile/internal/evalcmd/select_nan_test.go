package evalcmd_test

import (
	"testing"

	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// z is 0 everywhere, but not a constant the compiler can fold, so 1 / z is infinite and z / z is NaN at every point.
const zeroField = "@scene(40)\nz = p.x - p.x;\nball = length(p) - 10;\nreturn ball;\n"

// values is an eval group's scalar values, one per point (nil where the value is not finite).
func values(t *testing.T, group any) []any {
	t.Helper()
	return group.(map[string]any)["values"].([]any)[0].([]any)
}

// A ternary takes one branch: a branch not taken that is infinite or NaN there leaves the result alone. This holds when
// the condition varies over space, not only when it folds to a constant.
func TestEval_ATernaryIgnoresTheBranchItDoesNotTake(t *testing.T) {
	r, code := ask(t, document(t, zeroField), map[string]any{"op": "eval", "exprs": []string{
		"(p.x > 0) ? p.x : 1 / z",
		"(p.x > 0) ? z / z : p.y",
		"p.x ? 1 / z : 5 : p.y - 1",
		"p.x ? 7 : z / z : 1 / z",
	}, "points": [][3]float64{{1, 2, 0}, {-1, 2, 0}, {0, 2, 0}}})
	require.Equal(t, 0, code, r)
	groups := r["groups"].([]any)

	// cond ? a : b
	got := values(t, groups[0])
	near(t, 1, got[0], 0)
	assert.Equal(t, nil, got[1], "the branch taken is infinite at x = -1")
	got = values(t, groups[1])
	assert.Equal(t, nil, got[0], "the branch taken is NaN at x = 1")
	near(t, 2, got[1], 0)

	// v ? positive : zero : negative
	got = values(t, groups[2])
	assert.Equal(t, nil, got[0])
	near(t, 1, got[1], 0)
	near(t, 5, got[2], 0)
	got = values(t, groups[3])
	near(t, 7, got[0], 0)
	assert.Equal(t, nil, got[1])
	assert.Equal(t, nil, got[2])
}

// Where the branch taken is 0 -- on its surface -- its gradient is still the result's.
func TestEval_ATernaryKeepsTheGradientOfAZeroBranch(t *testing.T) {
	r, code := ask(t, document(t, zeroField), map[string]any{"op": "eval", "exprs": []string{
		"(p.y > 0) ? p.x : 2",
		"p.y ? p.x - 3 : 0 : 9",
	}, "points": [][3]float64{{0, 1, 0}, {3, 1, 0}}})
	require.Equal(t, 0, code, r)
	for i, g := range r["groups"].([]any) {
		group := g.(map[string]any)
		point := []int{0, 1}[i]
		near(t, 0, group["values"].([]any)[0].([]any)[point], 0)
		grad := group["gradients"].([]any)[0].([]any)[point].([]any)
		for k, want := range []float64{1, 0, 0} {
			near(t, want, grad[k], 0)
		}
	}
}

// unionRoundExact switches to the sharp union away from the seam. On the line through two balls' centres, outside
// both, its other branch divides by a zero normal difference: the field there is still the sharp union's distance.
func TestEval_UnionRoundExactOnTheAxisOutsideBothBalls(t *testing.T) {
	doc := "@scene(60)\na = length(p - [-6, 0, 0]) - 8;\nb = length(p - [6, 0, 0]) - 8;\n" +
		"u = unionRoundExact(() { return a; }, () { return b; });\nreturn u;\n"
	r, code := ask(t, document(t, doc), map[string]any{"op": "eval", "exprs": []string{"u"},
		"points": [][3]float64{{20, 0, 0}, {-25, 0, 0}}})
	require.Equal(t, 0, code, r)
	got := values(t, r["groups"].([]any)[0])
	near(t, 6, got[0], 1e-5)
	near(t, 11, got[1], 1e-5)
}
