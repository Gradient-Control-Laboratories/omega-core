package ui_test

import (
	"testing"

	"gradientcontrol.com/omega/core/assembler/ui"
	"gradientcontrol.com/omega/core/internal/test/assert"
)

// uiColor and uiDirection are vec3 controls like uiPoint, three editable
// literals each, reported under their own kinds so the browser can draw a
// colour chip and a direction arrow rather than a point handle. The @color
// and @direction annotations lower to them.
func TestExtractColorAndDirection(t *testing.T) {
	src := "@color coat = [0.8, 0.1, 0.1];\n" +
		"@direction up = [0, 0, 1];\n" +
		"return coat[0] + up[2];\n"
	controls := extract(t, src)
	assert.Equal(t, 2, len(controls))

	assert.Equal(t, ui.KindColor, controls[0].Kind)
	assert.Equal(t, []float64{0.8, 0.1, 0.1}, controls[0].Value)
	assert.Equal(t, "[0.8, 0.1, 0.1]", textAt(src, controls[0].Span))
	assert.Equal(t, 3, len(controls[0].Params))

	assert.Equal(t, ui.KindDirection, controls[1].Kind)
	assert.Equal(t, []float64{0, 0, 1}, controls[1].Value)
	assert.Equal(t, "[0, 0, 1]", textAt(src, controls[1].Span))
	assert.Equal(t, 3, len(controls[1].Params))
}
