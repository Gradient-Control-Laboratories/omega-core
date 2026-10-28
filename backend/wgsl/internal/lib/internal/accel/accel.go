package accel

import (
	"strings"

	"gradientcontrol.com/omega/core/backend/wgsl/internal/lib/internal/wgsl"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// Lookup finds the OpTemplate for a given mangled WGSL function name.
func Lookup(wgslName string) (OpTemplate, bool) {
	if pkg, rest, ok := strings.Cut(wgslName, "_"); ok {
		fnName, _, _ := strings.Cut(rest, "_")
		if tmpl, ok := libraryOps[pkg+"."+fnName]; ok {
			return tmpl, true
		}
	}
	return OpTemplate{}, false
}

// IsAccelerated returns true if the function has hardware acceleration and valid argument types.
func IsAccelerated(wgslName string, paramTypes []wgsl.TypeInfo) bool {
	opTemplate, ok := Lookup(wgslName)
	if !ok {
		return false
	}
	_, ok = opTemplate.TypeDeriver(paramTypes)
	return ok
}

// BuildCall builds a WGSL expression for an accelerated standard library function call.
func BuildCall(
	fnCallOp *program.Operand,
	wgslName string,
	expressions []build.Expression,
) (build.Expression, bool, error) {
	opTemplate, ok := Lookup(wgslName)
	if !ok {
		return nil, false, nil
	}

	argTypes := make([]wgsl.TypeInfo, len(expressions))
	for i, e := range expressions {
		if we, ok := e.(wgsl.Expression); ok {
			argTypes[i] = we.WGSLType()
		} else {
			argTypes[i] = wgsl.WGSLType(e.Type(), false)
		}
	}

	outType, ok := opTemplate.TypeDeriver(argTypes)
	if !ok {
		return nil, false, nil
	}

	switch opTemplate.Kind {
	case OpKindOperator:
		expr, err := wgsl.NewNumericOp(fnCallOp, expressions, opTemplate.Precedence, outType, opTemplate.Symbol)
		return expr, true, err
	case OpKindBuiltin:
		wgslOutType := wgsl.WGSLType(outType, false)
		return wgsl.NewCustomBuiltinCall(fnCallOp, opTemplate.Builtin, expressions, wgslOutType), true, nil
	default:
		return nil, false, nil
	}
}
