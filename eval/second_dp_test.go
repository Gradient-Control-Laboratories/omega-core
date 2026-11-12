//go:build cgo

package eval_test

import (
	"context"
	"math"
	"testing"
	"time"

	"gradientcontrol.com/omega/core/eval"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// chamferedSlabDoc is one chamfered box. intersectChamfer's inlined body
// reuses each face operand by pointer, so its .second expansion is a DAG.
const chamferedSlabDoc = `
slab = box([-14.0, -7.0, 4.0], [14.0, 7.0, 15.0], intersectChamfer(0.5));
return slab;
`

// TestEvalGslSecondGradientStaysSmall: .second.dp of a chamfered box must
// assemble in seconds. Walking the .second DAG as a tree under .dp took 38 s
// for this one box and 25 minutes for a gripper base of a few, stalling every
// verifyDetail wall probe that asked for it.
func TestEvalGslSecondGradientStaysSmall(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)

	type result struct {
		value [3]float32
		err   error
	}
	done := make(chan result, 1)
	go func() {
		groups, err := e.EvalGsl(context.Background(), nil, chamferedSlabDoc, "",
			[]string{"slab.second.dp"}, [][3]float64{{0, 0, 10}})
		if err != nil {
			done <- result{err: err}
			return
		}
		var v [3]float32
		for c := 0; c < 3; c++ {
			v[c] = groups[0].Samples[c][0].Value
		}
		done <- result{value: v}
	}()

	select {
	case r := <-done:
		require.NoError(t, r.err)
		// The value the unshared expansion produces too, only slowly: the fix
		// changes how long the gradient takes, not what it is.
		want := [3]float32{0, 0, 1}
		for c := range want {
			if math.Abs(float64(r.value[c]-want[c])) > 1e-4 {
				t.Errorf("slab.second.dp[%d] = %v, want %v", c, r.value[c], want[c])
			}
		}
	case <-time.After(15 * time.Second):
		t.Fatal("slab.second.dp did not assemble within 15 s: the .second expansion is being walked as a tree under .dp")
	}
}
