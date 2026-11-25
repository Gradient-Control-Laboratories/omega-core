//go:build cgo

package eval_test

import (
	"context"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/eval"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/source"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/assert"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/require"
)

// An L-bracket: a base 60 by 40 by 8 with a wall 6 thick standing 40 tall along one edge.
const bracketDoc = `
base = max(max(abs(p.x) - 30.0, abs(p.y) - 20.0), abs(p.z - 4.0) - 4.0);
wall = max(max(abs(p.x) - 30.0, abs(p.y - 17.0) - 3.0), abs(p.z - 20.0) - 20.0);
bracket = min(base, wall);
return bracket;
`

// encloses asserts that the proven box holds want and exceeds it by no more than slack on any side.
func encloses(t *testing.T, r *eval.EncloseResult, want [6]float64, slack float64) {
	t.Helper()
	require.True(t, r.Proven, "reason %q", r.Reason)
	require.True(t, r.Found)
	for k := 0; k < 3; k++ {
		assert.True(t, r.BboxMin[k] <= want[k] && r.BboxMin[k] >= want[k]-slack, "min %d is %v, want %v", k, r.BboxMin[k], want[k])
		assert.True(t, r.BboxMax[k] >= want[k+3] && r.BboxMax[k] <= want[k+3]+slack, "max %d is %v, want %v", k, r.BboxMax[k], want[k+3])
		assert.True(t, r.CellSize[k] <= slack, "cell %d is %v", k, r.CellSize[k])
	}
}

// The wall is thinner than a cell of the default scan, which reads the bracket as 8 tall; the proof does not.
func TestEnclose_AWallThinnerThanACellIsNotMissed(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	r, err := e.Enclose(context.Background(), nil, bracketDoc, "", "bracket", [6]float64{}, 0)
	require.NoError(t, err)
	encloses(t, r, [6]float64{-30, -20, 0, 30, 20, 40}, 0.1)
}

// A ball far smaller than a cell of the default region, and off its centre.
func TestEnclose_ASmallBallInALargeRegion(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	r, err := e.Enclose(context.Background(), nil, "s = length(p - [100.0, -50.0, 7.0]) - 2.0;\nreturn s;\n", "", "s",
		[6]float64{}, 0)
	require.NoError(t, err)
	encloses(t, r, [6]float64{98, -52, 5, 102, -48, 9}, 0.05)
}

// The box follows defs, and a region that cuts the part bounds the box.
func TestEnclose_DefsAndRegion(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	r, err := e.Enclose(context.Background(), nil, testDoc, "grown = s - 1.0;", "grown",
		[6]float64{-10, -10, 0, 10, 10, 10}, 0)
	require.NoError(t, err)
	encloses(t, r, [6]float64{-3, -3, 0, 3, 3, 3}, 0.05)
}

// A field nowhere below zero: proven empty, with the scan's nearest approach still reported.
func TestEnclose_NothingInsideIsProven(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	r, err := e.Enclose(context.Background(), nil, testDoc, "", "length(p) + 100.0",
		[6]float64{-10, -10, -10, 10, 10, 10}, 16)
	require.NoError(t, err)
	assert.True(t, r.Proven)
	assert.False(t, r.Found)
	assert.True(t, r.MinValue >= 100 && r.MinValue <= 120, "minValue %v", r.MinValue)
}

// A field whose intervals are unbounded on every box wider than a ripple: no proof, so the sampled box, and
// said to be so.
func TestEnclose_NoProofFallsBackToTheScan(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	region := [6]float64{-10, -10, -10, 10, 10, 10}
	expr := "(length(p) - 2.0) / abs(sin(p.x * 400.0) * sin(p.y * 400.0) * sin(p.z * 400.0))"
	r, err := e.Enclose(context.Background(), nil, testDoc, "", expr, region, 32)
	require.NoError(t, err)
	assert.False(t, r.Proven)
	assert.True(t, r.Reason == eval.ReasonBudget || r.Reason == eval.ReasonUnbounded, "reason %q", r.Reason)

	sampled, err := e.Bounds(context.Background(), nil, testDoc, "", expr, region, 32)
	require.NoError(t, err)
	assert.Equal(t, *sampled, r.BoundsResult)
}
