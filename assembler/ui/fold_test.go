package ui_test

import (
	"testing"

	"gradientcontrol.com/omega/core/assembler/ui"
	"gradientcontrol.com/omega/core/internal/test/assert"
)

// A control's literal is a parameter, not a constant: the assembler's folding
// leaves a uiColor and a uiDirection in place as it leaves a uiPoint, however
// constant their use, so the chip and the arrow survive to the viewer. (A
// direction with no anchor, or a constant one, is the case a fold would
// otherwise take whole.)
func TestColorAndDirectionSurviveFolding(t *testing.T) {
	src := "@color tint = [0.2, 0.5, 0.9];\n" +
		"@direction along = [1, 0, 0];\n" +
		"return tint[0] + tint[1] + along[2];\n"
	controls := extract(t, src)
	assert.Equal(t, 2, len(controls))
	assert.Equal(t, ui.KindColor, controls[0].Kind)
	assert.Equal(t, []float64{0.2, 0.5, 0.9}, controls[0].Value)
	assert.Equal(t, ui.KindDirection, controls[1].Kind)
	assert.Equal(t, 3, len(controls[1].Params))
}
