package slider

import (
	"gradientcontrol.com/omega/core/internal/edit"
	"gradientcontrol.com/omega/core/internal/eval"
	"gradientcontrol.com/omega/core/program"
)

// SliderInfo describes a top-level variable declared with both a `.min` and `.max` property
// whose assignment is a simple numeric literal.
//
// The editor renders a slider per entry. The slider's value drives a WGSL uniform.
// The wgsl agent lifts the literal into a top-level `sdShape` parameter
// with the name stored in the ParamName field so the host can update it without recompiling.
// The editor literal at LiteralStart is also rewritten so the source persists the new value.
type SliderInfo struct {
	VariableName string    `json:"variableName"`
	Min          float64   `json:"min"`
	Max          float64   `json:"max"`
	InitialValue float64   `json:"initialValue"`
	Decimals     int       `json:"decimals"`         // Decimal-place count of the source literal (0 for "5", 1 for "0.5", 2 for "0.50") — drives slider rounding step (10^-decimals)
	ParamName    string    `json:"paramName"`        // Codegen-emitted WGSL parameter name (gsl_param_<line>_<col>) — matches the lifted literal
	Span         edit.Span `json:"span"`             // Full span of the literal token in source; on each drag we overwrite this range
	Period       float64   `json:"period,omitempty"` // Optional .period (seconds): when > 0, the host drives the uniform as a sinusoid through [min, max] with the given period; 0/absent means static
}

// Extract walks the top-level declarations and returns one
// SliderInfo per variable that satisfies all of:
//   - has a range annotation (explicit @range or synthesized from .min/.max)
//   - is a simple assignment to a single name
//   - the assignment expression is a plain numeric literal whose source
//     span we can locate exactly (so the UI can rewrite the literal in
//     place on each drag without disturbing other source structure)
//
// Anything more elaborate on the RHS (`x = 27 + 3` or a function call)
// is skipped; the slider UI has nowhere unambiguous to commit the new
// value back to.
func Extract(source string, pkg *program.Package) []SliderInfo {
	if pkg == nil {
		return nil
	}
	assignments, fnTypes := topLevel(pkg)
	var sliders []SliderInfo
	for _, a := range assignments {
		if s, ok := sliderFromAssignment(source, a, assignments, fnTypes); ok {
			sliders = append(sliders, s)
		}
	}
	return sliders
}

// Lifted returns the assignments Extract surfaces as sliders. The wgsl
// codegen lifts exactly these into sdShape parameters and the viewer passes
// one argument per slider, so the two must agree: a parameter without a
// slider is an sdShape call one argument short.
func Lifted(pkg *program.Package) map[*program.Assignment]bool {
	lifted := map[*program.Assignment]bool{}
	if pkg == nil {
		return lifted
	}
	assignments, fnTypes := topLevel(pkg)
	for _, a := range assignments {
		if _, ok := sliderFromAssignment("", a, assignments, fnTypes); ok {
			lifted[a] = true
		}
	}
	return lifted
}

// topLevel gathers the package's declarations, the name scope a bound
// computed from other constants is folded against.
func topLevel(pkg *program.Package) ([]*program.Assignment, []*program.FnType) {
	var assignments []*program.Assignment
	var fnTypes []*program.FnType
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		assignments = append(assignments, f.Assignments...)
		fnTypes = append(fnTypes, f.FnTypes...)
	}
	return assignments, fnTypes
}

func sliderFromAssignment(source string, a *program.Assignment,
	assignments []*program.Assignment, fnTypes []*program.FnType) (SliderInfo, bool) {
	if len(a.Variables) != 1 || a.Range == nil || a.Expression == nil {
		return SliderInfo{}, false
	}
	v := a.Variables[0]

	minVal, okMin := a.Range.MinConst()
	if !okMin {
		minVal, okMin = foldBound(a.Range.Min, assignments, fnTypes)
	}
	maxVal, okMax := a.Range.MaxConst()
	if !okMax {
		maxVal, okMax = foldBound(a.Range.Max, assignments, fnTypes)
	}
	if !okMin || !okMax {
		return SliderInfo{}, false
	}

	// Only literal-valued assignments can be scrubbed in place.
	op := a.Expression
	if op.Start == nil || op.End == nil {
		return SliderInfo{}, false
	}
	valOp, isValue := op.T.(*program.Operand_Value)
	if !isValue || valOp == nil {
		return SliderInfo{}, false
	}

	// span matches the parser's operand span exactly.
	span := edit.Span{
		StartLine:   int(op.Start.Line),
		StartColumn: int(op.Start.Column),
		EndLine:     int(op.End.Line),
		EndColumn:   int(op.End.Column),
	}

	// .period is optional. A positive value puts the slider into
	// animation mode (host drives the uniform as a sinusoid through
	// [min, max] with the given period in seconds). Zero or negative
	// is treated the same as not setting the property at all.
	var period float64
	p, okPeriod := a.Range.PeriodConst()
	if !okPeriod && a.Range.Period != nil {
		p, okPeriod = foldBound(a.Range.Period, assignments, fnTypes)
	}
	if okPeriod && p > 0 {
		period = p
	}

	return SliderInfo{
		VariableName: v.Name,
		Min:          minVal,
		Max:          maxVal,
		InitialValue: valOp.Value,
		Decimals:     literalDecimals(source, op.Start, op.End),
		ParamName:    span.VarName(),
		Span:         span,
		Period:       period,
	}, true
}

// foldBound reads a range bound that isn't a literal. A bound may be derived
// from other constants — Slider Mechanism's
// `slidePos.min = abs(slideRadius - baseRadius) - slideOffset` — so fold it
// against the package's declarations. A bound reaching p or time still fails
// to fold, and the variable is not a slider.
func foldBound(op *program.Operand,
	assignments []*program.Assignment, fnTypes []*program.FnType) (float64, bool) {
	vals, err := eval.Const(op, assignments, fnTypes)
	if err != nil || len(vals) != 1 {
		return 0, false
	}
	return vals[0], true
}

// literalDecimals reads the source literal's text and returns the
// count of digits AFTER the decimal point. That count drives the
// slider's rounding step (10^-decimals): `5` (no dot) → 0 → step 1;
// `0.5` → 1 → step 0.1; `2.00` → 2 → step 0.01; `0.025` → 3 → step
// 0.001. Trailing zeros after the decimal are significant — they
// declare the author's intended precision.
//
// Falls back to 0 (integer step) if the span can't be resolved.
func literalDecimals(source string, start, end *program.Location) int {
	if start == nil || end == nil {
		return 0
	}
	s := int(start.Offset)
	e := int(end.Offset)
	if s < 0 || e <= s || e > len(source) {
		return 0
	}
	text := source[s:e]
	dot := -1
	for i := 0; i < len(text); i++ {
		if text[i] == '.' {
			dot = i
			break
		}
	}
	if dot < 0 {
		return 0
	}
	count := 0
	for i := dot + 1; i < len(text); i++ {
		c := text[i]
		if c < '0' || c > '9' {
			break
		}
		count++
	}
	return count
}
