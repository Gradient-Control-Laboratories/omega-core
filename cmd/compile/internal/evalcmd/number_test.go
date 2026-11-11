package evalcmd

import (
	"encoding/json"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// A value that is not finite is written as null (encoding/json refuses NaN and infinities); a float32 is written at
// its own precision, not as the float64 it widens to.
func TestNumbers_NonFiniteIsNullAndFloat32KeepsItsDigits(t *testing.T) {
	data, err := json.Marshal([]any{f32(float32(math.NaN())), f64(math.Inf(-1)), f32(0.1), f64(0.1), [3]f32{1, 2, 3}})
	require.NoError(t, err)
	assert.Equal(t, `[null,null,0.1,0.1,[1,2,3]]`, string(data))
}
