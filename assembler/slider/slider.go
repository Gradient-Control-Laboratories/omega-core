package slider

import "gradientcontrol.com/omega/core/program/edit"

// SliderInfo describes a top-level variable declared with both a `.min` and `.max` property
// whose assignment is a simple numeric literal.
//
// The editor renders a slider per entry. The slider's value drives a WGSL uniform.
// The wgsl agent lifts the literal into a top-level `sdShape` parameter
// with the name stored in the ParamName field so the host can update it without recompiling.
// The editor literal at LiteralStart is also rewritten so the source persists the new value.
// TODO: Merge into UI controls
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

