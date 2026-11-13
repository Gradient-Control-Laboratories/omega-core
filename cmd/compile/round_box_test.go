package main_test

import (
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/test/assert"
)

// The exact box field with round dimensions, unshelled. Dual contouring finds its faces' crossings exactly on
// them, where max(d, 0) and min(max(d.x, d.y, d.z), 0) tie with their constants and the evaluator's gradient is
// zero; the mesher evaluates it again off the tie, so every face gives a plane and the box keeps its edges and
// its volume, at any depth.
func TestRun_STL_RoundExactBoxKeepsItsEdges(t *testing.T) {
	source := "d = abs(p) - [20, 15, 10];\n" +
		"return length(max(d, 0)) + min(max(d.x, d.y, d.z), 0);\n"
	for _, depth := range []string{"7", "8", "9"} {
		closed, volume := closedVolume(stlTriangles(t, meshed(t, source, "--subdivision", depth, "--bbox", "auto")))
		assert.True(t, closed, "depth %s: closed", depth)
		want := 40.0 * 30 * 20
		assert.True(t, math.Abs(volume-want) < 0.005*want, "depth %s: volume %.1f, want %.0f", depth, volume, want)
	}
}
