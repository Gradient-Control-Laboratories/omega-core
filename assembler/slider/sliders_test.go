package slider_test

import (
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/assembler/slider"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/assert"
)

// runExtract parses src and returns the SliderInfo entries the wgsl
// agent would surface to the client. Bypasses Render so the test is
// independent of the WGSL codegen.
func runExtract(t *testing.T, src string) []sliderUnderTest {
	t.Helper()
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(&source.MockManager{}))
	assert.NoError(t, err)
	infos := slider.Extract(src, res.Program.Primary)
	out := make([]sliderUnderTest, len(infos))
	for i, s := range infos {
		out[i] = sliderUnderTest{
			Name:        s.VariableName,
			Min:         s.Min,
			Max:         s.Max,
			Initial:     s.InitialValue,
			Decimals:    s.Decimals,
			ParamName:   s.ParamName,
			StartLine:   s.Span.StartLine,
			StartColumn: s.Span.StartColumn,
			EndLine:     s.Span.EndLine,
			EndColumn:   s.Span.EndColumn,
			Period:      s.Period,
		}
	}
	return out
}

type sliderUnderTest struct {
	Name        string
	Min, Max    float64
	Initial     float64
	Decimals    int
	ParamName   string
	StartLine   int
	StartColumn int
	EndLine     int
	EndColumn   int
	Period      float64
}

// A variable with both .min and .max whose RHS is a plain integer
// literal surfaces as a slider; the literal's source span is
// 1-based, end-exclusive.
func TestExtractSliders_LiteralInt(t *testing.T) {
	src := "radius = 5;\n" +
		"radius.min = 1;\n" +
		"radius.max = 10;\n" +
		"return radius;\n"
	sliders := runExtract(t, src)
	assert.Equal(t, 1, len(sliders))
	got := sliders[0]
	assert.Equal(t, "radius", got.Name)
	assert.Equal(t, 1.0, got.Min)
	assert.Equal(t, 10.0, got.Max)
	assert.Equal(t, 5.0, got.Initial)
	// "5" has no decimal point → 0 decimal places → integer step.
	assert.Equal(t, 0, got.Decimals)
	// "5" sits at line 1, columns 10..11 (1-based, end-exclusive).
	assert.Equal(t, 1, got.StartLine)
	assert.Equal(t, 10, got.StartColumn)
	assert.Equal(t, 1, got.EndLine)
	assert.Equal(t, 11, got.EndColumn)
	// ParamName encodes the operand's actual position so codegen can
	// match it during literal-to-parameter lifting.
	assert.Equal(t, "gsl_param_1_10", got.ParamName)
}

// Decimal literal: the author wrote N digits after the decimal point,
// so the slider rounds to that precision (step = 10^-N).
func TestExtractSliders_LiteralDecimal(t *testing.T) {
	src := "radius = 0.25;\n" +
		"radius.min = 0;\n" +
		"radius.max = 1;\n" +
		"return radius;\n"
	sliders := runExtract(t, src)
	assert.Equal(t, 1, len(sliders))
	assert.Equal(t, 0.25, sliders[0].Initial)
	// "0.25" has 2 digits after the decimal → step 0.01.
	assert.Equal(t, 2, sliders[0].Decimals)
}

// Trailing zeros after the decimal declare the author's intended
// precision, so `5.00` rounds at hundredths even though the value
// could be written as just `5`.
func TestExtractSliders_TrailingDecimalZero(t *testing.T) {
	src := "radius = 5.00;\n" +
		"radius.min = 0;\n" +
		"radius.max = 10;\n" +
		"return radius;\n"
	sliders := runExtract(t, src)
	assert.Equal(t, 1, len(sliders))
	assert.Equal(t, 5.0, sliders[0].Initial)
	// "5.00" — 2 digits after the decimal → step 0.01.
	assert.Equal(t, 2, sliders[0].Decimals)
	// UI rewrite span covers the full "5.00" (cols 10..14) so a
	// commit can overwrite trailing zeros in addition to the digit.
	assert.Equal(t, 10, sliders[0].StartColumn)
	assert.Equal(t, 14, sliders[0].EndColumn)
}

// `0.50` exercises the leading-zero-with-trailing-zero case — the
// authoring intent is "tenths precision with a trailing zero showing
// two decimals", and the slider should round to 0.01 step.
func TestExtractSliders_LeadingZeroTrailingZero(t *testing.T) {
	src := "radius = 0.50;\n" +
		"radius.min = 0;\n" +
		"radius.max = 1;\n" +
		"return radius;\n"
	sliders := runExtract(t, src)
	assert.Equal(t, 1, len(sliders))
	assert.Equal(t, 0.5, sliders[0].Initial)
	assert.Equal(t, 2, sliders[0].Decimals)
	// Span covers all 4 chars of "0.50" (cols 10..14).
	assert.Equal(t, 10, sliders[0].StartColumn)
	assert.Equal(t, 14, sliders[0].EndColumn)
}

// Integer literals (no decimal point) get integer step, regardless of
// trailing zeros: `100` → 0 decimals → step 1. The slider doesn't
// interpret trailing-integer-zeros as "tens precision" because the
// author can opt into that explicitly by adding a decimal point.
func TestExtractSliders_IntegerTrailingZero(t *testing.T) {
	src := "n = 100;\n" +
		"n.min = 0;\n" +
		"n.max = 1000;\n" +
		"return n;\n"
	sliders := runExtract(t, src)
	assert.Equal(t, 1, len(sliders))
	assert.Equal(t, 100.0, sliders[0].Initial)
	assert.Equal(t, 0, sliders[0].Decimals)
}

// A signed-literal RHS: the parser folds `-` into the Operand_Value,
// and correctly sets the start position at the leading `-`.
func TestExtractSliders_LiteralNegative(t *testing.T) {
	src := "offset = -3;\n" +
		"offset.min = -10;\n" +
		"offset.max = 10;\n" +
		"return offset;\n"
	sliders := runExtract(t, src)
	assert.Equal(t, 1, len(sliders))
	got := sliders[0]
	assert.Equal(t, -3.0, got.Initial)
	assert.Equal(t, 10, got.StartColumn)
	assert.Equal(t, 12, got.EndColumn)
	// ParamName uses the operand's actual start column (10) since the
	// parser now correctly points to the leading `-`.
	assert.Equal(t, "gsl_param_1_10", got.ParamName)
}

// Without min/max properties the variable is just a regular
// declaration and produces no slider entry.
func TestExtractSliders_NoProperties(t *testing.T) {
	src := "radius = 5;\n" +
		"return radius;\n"
	sliders := runExtract(t, src)
	assert.Equal(t, 0, len(sliders))
}

// Only one of the two required properties is set — rejected by property validation.
func TestExtractSliders_OnlyMin(t *testing.T) {
	src := "radius = 5;\n" +
		"radius.min = 0;\n" +
		"return radius;\n"
	_, err := assembler.Assemble([]string{src}, assembler.SourceManager(&source.MockManager{}))
	assert.Error(t, err)
}

// An expression RHS (`27 + 3`) has no single literal to rewrite, so
// the slider UI has nowhere to commit changes — skip.
func TestExtractSliders_ExpressionRHS(t *testing.T) {
	src := "radius = 27 + 3;\n" +
		"radius.min = 0;\n" +
		"radius.max = 100;\n" +
		"return radius;\n"
	sliders := runExtract(t, src)
	assert.Equal(t, 0, len(sliders))
}

// .period turns a slider into an animated one — the host drives a
// sinusoid through [min, max] with the given period (seconds). It's
// optional; min/max are still required.
func TestExtractSliders_Period(t *testing.T) {
	src := "radius = 5;\n" +
		"radius.min = 1;\n" +
		"radius.max = 10;\n" +
		"radius.period = 4.0;\n" +
		"return radius;\n"
	sliders := runExtract(t, src)
	assert.Equal(t, 1, len(sliders))
	assert.Equal(t, 4.0, sliders[0].Period)
}

// A zero or negative period is rejected by property validation.
func TestExtractSliders_PeriodNonPositive(t *testing.T) {
	src := "radius = 5;\n" +
		"radius.min = 1;\n" +
		"radius.max = 10;\n" +
		"radius.period = 0;\n" +
		"return radius;\n"
	_, err := assembler.Assemble([]string{src}, assembler.SourceManager(&source.MockManager{}))
	assert.Error(t, err)
}

// .period without .min or .max is rejected by property validation.
func TestExtractSliders_PeriodWithoutMinMax(t *testing.T) {
	src := "radius = 5;\n" +
		"radius.period = 2.0;\n" +
		"return radius;\n"
	_, err := assembler.Assemble([]string{src}, assembler.SourceManager(&source.MockManager{}))
	assert.Error(t, err)
}

func TestExtractSliders_ExplicitRange(t *testing.T) {
	src := "@range(1.0, 10.0)\n" +
		"radius = 5.0;\n" +
		"return radius;\n"
	sliders := runExtract(t, src)
	assert.Equal(t, 1, len(sliders))
	assert.Equal(t, "radius", sliders[0].Name)
	assert.Equal(t, 1.0, sliders[0].Min)
	assert.Equal(t, 10.0, sliders[0].Max)
	assert.Equal(t, 5.0, sliders[0].Initial)
	assert.Equal(t, 1, sliders[0].Decimals)
	assert.Equal(t, "gsl_param_2_10", sliders[0].ParamName)
}

func TestExtractSliders_ExplicitRangeWithPeriod(t *testing.T) {
	src := "@range(1.0, 10.0, 3.5)\n" +
		"radius = 5.0;\n" +
		"return radius;\n"
	sliders := runExtract(t, src)
	assert.Equal(t, 1, len(sliders))
	assert.Equal(t, "radius", sliders[0].Name)
	assert.Equal(t, 1.0, sliders[0].Min)
	assert.Equal(t, 10.0, sliders[0].Max)
	assert.Equal(t, 3.5, sliders[0].Period)
}

// A bound computed from other constants folds, as Slider Mechanism's
// `slidePos.min = abs(slideRadius - baseRadius) - slideOffset` does. Skipping
// it dropped the slider and its .period animation.
func TestExtractSliders_ComputedBounds(t *testing.T) {
	src := "slideRadius = 6.00;\n" +
		"slideRadius.min = 4;\n" +
		"slideRadius.max = 8;\n" +
		"baseRadius = 4.00;\n" +
		"slideOffset = 4;\n" +
		"slidePos = 2.02;\n" +
		"slidePos.min = abs(slideRadius - baseRadius) - slideOffset;\n" +
		"slidePos.max = slideRadius + baseRadius - slideOffset;\n" +
		"slidePos.period = 10;\n" +
		"return slideRadius + slidePos;\n"
	sliders := runExtract(t, src)
	assert.Equal(t, 2, len(sliders))
	got := sliders[1]
	assert.Equal(t, "slidePos", got.Name)
	assert.Equal(t, -2.0, got.Min)
	assert.Equal(t, 6.0, got.Max)
	assert.Equal(t, 10.0, got.Period)
	assert.Equal(t, "gsl_param_6_12", got.ParamName)
}
