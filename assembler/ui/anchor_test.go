package ui_test

import (
	"testing"

	"gradientcontrol.com/omega/core/assembler/ui"
	"gradientcontrol.com/omega/core/internal/test/assert"
)

// @direction(anchor) lowers to uiDirection(dir, anchor), and the anchor —
// a @point variable or a constant — reaches the browser evaluated, so the
// arrow is drawn from it. The anchor is not editable: only the direction's
// three components are the control's params.
func TestDirectionAnchor(t *testing.T) {
	src := "@point center = [10, 0, 5];\n" +
		"@direction(center) axis = [0, 0, 1];\n" +
		"@direction([1, 2, 3]) other = [0, 1, 0];\n" +
		"@direction bare = [1, 0, 0];\n" +
		"return center[0] + axis[2] + other[1] + bare[0];\n"
	controls := extract(t, src)
	assert.Equal(t, 4, len(controls))
	assert.Equal(t, ui.KindPoint, controls[0].Kind)

	assert.Equal(t, ui.KindDirection, controls[1].Kind)
	assert.Equal(t, []float64{10, 0, 5}, controls[1].Anchor)
	assert.Equal(t, "[0, 0, 1]", textAt(src, controls[1].Span))
	assert.Equal(t, 3, len(controls[1].Params))

	assert.Equal(t, []float64{1, 2, 3}, controls[2].Anchor)
	assert.Equal(t, 0, len(controls[3].Anchor))
}
