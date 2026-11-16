package evalcmd_test

import (
	"testing"

	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// volume stops at the timeLimit asked, in seconds, says time ended it, and refuses a negative one.
func TestVolume_StopsAtItsTimeLimit(t *testing.T) {
	source := document(t, ball)
	r, code := ask(t, source, map[string]any{"op": "volume", "region": []float64{-12, -12, -12, 12, 12, 12},
		"relativeTolerance": 1e-6, "timeLimit": 1e-9})
	require.Equal(t, 0, code, r)
	assert.Equal(t, false, r["converged"], r)
	assert.Equal(t, "time", r["reason"], r)

	r, code = ask(t, source, map[string]any{"op": "volume", "region": []float64{-12, -12, -12, 12, 12, 12},
		"timeLimit": -1})
	assert.Equal(t, 2, code, r)
}
