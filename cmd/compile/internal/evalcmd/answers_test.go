package evalcmd_test

import (
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/assert"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/require"
)

// An L-bracket whose wall, 6 thick and 40 tall, is thinner than a cell of bounds' default scan.
const bracket = "@scene(100)\n" +
	"base = max(max(abs(p.x) - 30.0, abs(p.y) - 20.0), abs(p.z - 4.0) - 4.0);\n" +
	"wall = max(max(abs(p.x) - 30.0, abs(p.y - 17.0) - 3.0), abs(p.z - 20.0) - 20.0);\n" +
	"bracket = min(base, wall);\nreturn bracket;\n"

// bounds answers with a box it has proven to hold the part, and says so.
func TestBounds_IsProvenAndHoldsAThinWall(t *testing.T) {
	r, code := ask(t, document(t, bracket), map[string]any{"op": "bounds"})
	require.Equal(t, 0, code, r)
	assert.Equal(t, true, r["found"])
	assert.Equal(t, true, r["proven"], r)
	_, has := r["reason"]
	assert.False(t, has, "a proven box gives no reason: %v", r)
	for k, want := range []float64{-30, -20, 0} {
		got := number(t, r["bboxMin"].([]any)[k])
		assert.True(t, got <= want && want-got <= 0.1, "min %d is %g, want %g", k, got, want)
	}
	for k, want := range []float64{30, 20, 40} {
		got := number(t, r["bboxMax"].([]any)[k])
		assert.True(t, got >= want && got-want <= 0.1, "max %d is %g, want %g", k, got, want)
	}
}

// Where the field's intervals prove nothing, the box is the scan's, and bounds says it is not proven and why.
func TestBounds_SaysWhenItIsOnlySampled(t *testing.T) {
	r, code := ask(t, document(t, ball), map[string]any{"op": "bounds", "region": []float64{-20, -20, -20, 20, 20, 20},
		"expr": "ball / abs(sin(p.x * 400.0) * sin(p.y * 400.0) * sin(p.z * 400.0))"})
	require.Equal(t, 0, code, r)
	assert.Equal(t, false, r["proven"], r)
	reason, _ := r["reason"].(string)
	assert.True(t, reason == "budget" || reason == "unbounded", "reason %q", reason)
	assert.Equal(t, true, r["found"])
}

// volume says whether its bracket met the tolerance asked.
func TestVolume_SaysItConverged(t *testing.T) {
	r, code := ask(t, document(t, ball), map[string]any{"op": "volume", "region": []float64{-12, -12, -12, 12, 12, 12}})
	require.Equal(t, 0, code, r)
	assert.Equal(t, true, r["converged"], r)
	_, has := r["reason"]
	assert.False(t, has, "a converged volume gives no reason: %v", r)
}

// shell's bias moves the wall as its comment says: -1 inward, +1 outward, 0 centred on the surface.
func TestShell_BiasRunsAsDocumented(t *testing.T) {
	source := document(t, ball)
	at := [][3]float64{{19, 0, 0}, {20, 0, 0}, {21, 0, 0}}
	wall := func(bias string) []any {
		r, code := ask(t, source, map[string]any{"op": "eval", "points": at,
			"exprs": []string{"shell(length(p) - 20.0, 2.0, " + bias + ")"}})
		require.Equal(t, 0, code, r)
		return r["groups"].([]any)[0].(map[string]any)["values"].([]any)[0].([]any)
	}
	inward, outward, centred := wall("-1.0"), wall("1.0"), wall("0.0")
	near(t, -1, inward[0], 1e-5) // the middle of a wall from r = 18 to 20
	near(t, 1, inward[2], 1e-5)
	near(t, 1, outward[0], 1e-5) // the middle of a wall from r = 20 to 22
	near(t, -1, outward[2], 1e-5)
	near(t, -1, centred[1], 1e-5) // the middle of a wall from r = 19 to 21
}

// eval with no exprs evaluates the document's returns, as the other ops default their expr to them.
func TestEval_DefaultsToTheReturns(t *testing.T) {
	two := "@scene(40)\na = length(p - [5, 0, 0]) - 2;\nb = length(p + [5, 0, 0]) - 2;\nreturn a, b;\n"
	r, code := ask(t, document(t, two), map[string]any{"op": "eval", "points": [][3]float64{{5, 0, 0}}})
	require.Equal(t, 0, code, r)
	groups := r["groups"].([]any)
	require.Len(t, groups, 2)
	assert.Equal(t, "a", groups[0].(map[string]any)["expr"])
	near(t, -2, groups[0].(map[string]any)["values"].([]any)[0].([]any)[0], 1e-5)
	near(t, 8, groups[1].(map[string]any)["values"].([]any)[0].([]any)[0], 1e-5)
}
