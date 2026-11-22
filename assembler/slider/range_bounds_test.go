package slider_test

import (
	"testing"

	"gradientcontrol.com/omega/core/internal/test/assert"
)

// A bound written in @range that reads a plain constant keeps its slider and its
// period, as the same bounds written as .min/.max/.period do (see
// TestExtractSliders_ComputedBounds). The assembler's passes used to walk a
// variable's properties but not the bounds of an explicit @range, so
// `slideOffset`, read only by the bounds, was inlined away and the bounds could
// no longer fold: the slider vanished without an error.
func TestExtractSliders_RangeComputedBounds(t *testing.T) {
	src := "@range(40, 80) slideRadius = 60.00;\n" +
		"@range(20, 60) baseRadius = 40.00;\n" +
		"slideOffset = 40;\n" +
		"@range(abs(slideRadius - baseRadius) - slideOffset, slideRadius + baseRadius - slideOffset, 10)\n" +
		"slidePos = 20.2;\n" +
		"return slideRadius + baseRadius + slidePos;\n"
	sliders := runExtract(t, src)
	assert.Equal(t, 3, len(sliders))
	got := sliders[2]
	assert.Equal(t, "slidePos", got.Name)
	assert.Equal(t, -20.0, got.Min)
	assert.Equal(t, 60.0, got.Max)
	assert.Equal(t, 10.0, got.Period)
	assert.Equal(t, 20.2, got.Initial)
}

// A bound may read a constant that is itself computed, and one declared below the
// variable it bounds.
func TestExtractSliders_RangeBoundsReadComputedAndLaterConstants(t *testing.T) {
	src := "@range(lo, hi * 2) x = 5;\n" +
		"lo = 1;\n" +
		"hi = 10 + 2;\n" +
		"return sphere(x);\n"
	sliders := runExtract(t, src)
	assert.Equal(t, 1, len(sliders))
	assert.Equal(t, "x", sliders[0].Name)
	assert.Equal(t, 1.0, sliders[0].Min)
	assert.Equal(t, 24.0, sliders[0].Max)
}

// The same bounds written as .min/.max/.period properties keep their slider when
// they read a constant that is itself computed. The parser hands the synthesized
// range the properties' own operands, but a pass that folds a property's
// expression to 24 swaps in a new operand and leaves the range with the old tree,
// which reads `hi`, a variable dead-code elimination had removed because nothing
// live reads it any more: the slider vanished without an error.
func TestExtractSliders_PropertyBoundsReadComputedConstants(t *testing.T) {
	src := "hi = 10 + 2;\n" +
		"x = 5;\n" +
		"x.min = 0;\n" +
		"x.max = hi * 2;\n" +
		"x.period = hi / 4;\n" +
		"return sphere(x);\n"
	sliders := runExtract(t, src)
	assert.Equal(t, 1, len(sliders))
	assert.Equal(t, "x", sliders[0].Name)
	assert.Equal(t, 0.0, sliders[0].Min)
	assert.Equal(t, 24.0, sliders[0].Max)
	assert.Equal(t, 3.0, sliders[0].Period)
}
