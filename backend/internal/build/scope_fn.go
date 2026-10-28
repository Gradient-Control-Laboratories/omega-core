package build

import (
	"gradientcontrol.com/omega/core/program"
)

// Returns true if the types of the arguments match the types of the parameters.
func MatchArgTypes(args []Expression, paramTypes []*program.Type) bool {
	if len(args) != len(paramTypes) {
		return false
	}
	for i, arg := range args {
		argType := arg.Type()
		if argType == nil && arg.Operand() != nil {
			argType = arg.Operand().Type
		}
		if !argType.Equal(paramTypes[i]) {
			return false
		}
	}
	return true
}

// A function can be inlined at a call site when its body is a sequence of
// simple single-name assignments followed by a single return expression.
// tryInlineFuncCall folds the assignments into bindings one at a time,
// bailing out if one does not fold to a constant.
func isInlinableFunction(fn *program.Function) bool {
	if fn == nil || len(fn.Returns) != 1 {
		return false
	}
	for _, fnType := range fn.FnTypes {
		// Ignore anonymous lambda/closure helper declarations (prefixed with "gsl_anon")
		// that are hoisted into the function body during parsing/compilation. These
		// are static closure helper functions, not dynamic local variables or sequential
		// imperatives, so they do not prevent inlining the outer function.
		if !fnType.IsAnonFn() {
			return false
		}
	}
	for _, assign := range fn.Assignments {
		// Only single variable assignments can be inlined
		if len(assign.Variables) != 1 {
			return false
		}
	}

	return true
}

// isMultiReturnInlinable reports whether a multi-return function's body is a
// sequence of assignments (single- or multi-name) followed by its return list —
// the shape inlineMultiFuncCall can fold at a multi-value assignment. Unlike
// isInlinableFunction it permits multi-name assignments (e.g. a nested
// `shaft, holes = debossed(...)`), since the whole body is rebuilt at the call
// site rather than folded operand-by-operand.
func isMultiReturnInlinable(fn *program.Function) bool {
	if fn == nil || len(fn.Returns) < 2 {
		return false
	}
	for _, fnType := range fn.FnTypes {
		// Ignore anonymous lambda/closure helper declarations (prefixed with "gsl_anon")
		// that are hoisted into the function body during parsing/compilation. These
		// are static closure helper functions, not dynamic local variables or sequential
		// imperatives, so they do not prevent inlining the outer function.
		if !fnType.IsAnonFn() {
			return false
		}
	}
	return true
}

// -------------------------------------------------------------------------------

// FnInfo holds information about a cached function.
type FnInfo interface {
	Name() string
	ParamTypes() []*program.Type
}
