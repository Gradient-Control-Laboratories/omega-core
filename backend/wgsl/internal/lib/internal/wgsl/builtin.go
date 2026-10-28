package wgsl

import (
	"errors"
	"fmt"
	"strings"

	"gradientcontrol.com/omega/core/internal/build"
)

type Builtin struct {
	WGSLName  string
	Overloads []*BuiltinOverload
}

type BuiltinOverload struct {
	ParamTypes []TypeInfo
	ReturnType TypeInfo
}

// Keep this list in sync with the builtinsMap in common/model/builtin.go
var wgslBuiltins = map[string]*Builtin{
	"acosh": {
		WGSLName: "acosh",
		Overloads: []*BuiltinOverload{
			{ParamTypes: []TypeInfo{TypeF32}, ReturnType: TypeF32},
			{ParamTypes: []TypeInfo{TypeVec2}, ReturnType: TypeVec2},
			{ParamTypes: []TypeInfo{TypeVec3}, ReturnType: TypeVec3},
			{ParamTypes: []TypeInfo{TypeVec4}, ReturnType: TypeVec4},
		},
	},
	"asinh": {
		WGSLName: "asinh",
		Overloads: []*BuiltinOverload{
			{ParamTypes: []TypeInfo{TypeF32}, ReturnType: TypeF32},
			{ParamTypes: []TypeInfo{TypeVec2}, ReturnType: TypeVec2},
			{ParamTypes: []TypeInfo{TypeVec3}, ReturnType: TypeVec3},
			{ParamTypes: []TypeInfo{TypeVec4}, ReturnType: TypeVec4},
		},
	},
	"atanh": {
		WGSLName: "atanh",
		Overloads: []*BuiltinOverload{
			{ParamTypes: []TypeInfo{TypeF32}, ReturnType: TypeF32},
			{ParamTypes: []TypeInfo{TypeVec2}, ReturnType: TypeVec2},
			{ParamTypes: []TypeInfo{TypeVec3}, ReturnType: TypeVec3},
			{ParamTypes: []TypeInfo{TypeVec4}, ReturnType: TypeVec4},
		},
	},
	"clamp": {
		WGSLName: "clamp",
		Overloads: []*BuiltinOverload{
			{ParamTypes: []TypeInfo{TypeF32, TypeF32, TypeF32}, ReturnType: TypeF32},
			{ParamTypes: []TypeInfo{TypeVec2, TypeVec2, TypeVec2}, ReturnType: TypeVec2},
			{ParamTypes: []TypeInfo{TypeVec3, TypeVec3, TypeVec3}, ReturnType: TypeVec3},
			{ParamTypes: []TypeInfo{TypeVec4, TypeVec4, TypeVec4}, ReturnType: TypeVec4},
		},
	},
	"cosh": {
		WGSLName: "cosh",
		Overloads: []*BuiltinOverload{
			{ParamTypes: []TypeInfo{TypeF32}, ReturnType: TypeF32},
			{ParamTypes: []TypeInfo{TypeVec2}, ReturnType: TypeVec2},
			{ParamTypes: []TypeInfo{TypeVec3}, ReturnType: TypeVec3},
			{ParamTypes: []TypeInfo{TypeVec4}, ReturnType: TypeVec4},
		},
	},
	"cross": {
		WGSLName: "cross",
		Overloads: []*BuiltinOverload{
			{ParamTypes: []TypeInfo{TypeVec3, TypeVec3}, ReturnType: TypeVec3},
		},
	},
	"degrees": {
		WGSLName: "degrees",
		Overloads: []*BuiltinOverload{
			{ParamTypes: []TypeInfo{TypeF32}, ReturnType: TypeF32},
			{ParamTypes: []TypeInfo{TypeVec2}, ReturnType: TypeVec2},
			{ParamTypes: []TypeInfo{TypeVec3}, ReturnType: TypeVec3},
			{ParamTypes: []TypeInfo{TypeVec4}, ReturnType: TypeVec4},
		},
	},
	"distance": {
		WGSLName: "distance",
		Overloads: []*BuiltinOverload{
			{ParamTypes: []TypeInfo{TypeF32, TypeF32}, ReturnType: TypeF32},
			{ParamTypes: []TypeInfo{TypeVec2, TypeVec2}, ReturnType: TypeF32},
			{ParamTypes: []TypeInfo{TypeVec3, TypeVec3}, ReturnType: TypeF32},
			{ParamTypes: []TypeInfo{TypeVec4, TypeVec4}, ReturnType: TypeF32},
		},
	},
	"dot": {
		WGSLName: "dot",
		Overloads: []*BuiltinOverload{
			{ParamTypes: []TypeInfo{TypeVec2, TypeVec2}, ReturnType: TypeF32},
			{ParamTypes: []TypeInfo{TypeVec3, TypeVec3}, ReturnType: TypeF32},
			{ParamTypes: []TypeInfo{TypeVec4, TypeVec4}, ReturnType: TypeF32},
		},
	},
	"exp2": {
		WGSLName: "exp2",
		Overloads: []*BuiltinOverload{
			{ParamTypes: []TypeInfo{TypeF32}, ReturnType: TypeF32},
			{ParamTypes: []TypeInfo{TypeVec2}, ReturnType: TypeVec2},
			{ParamTypes: []TypeInfo{TypeVec3}, ReturnType: TypeVec3},
			{ParamTypes: []TypeInfo{TypeVec4}, ReturnType: TypeVec4},
		},
	},
	"fract": {
		WGSLName: "fract",
		Overloads: []*BuiltinOverload{
			{ParamTypes: []TypeInfo{TypeF32}, ReturnType: TypeF32},
			{ParamTypes: []TypeInfo{TypeVec2}, ReturnType: TypeVec2},
			{ParamTypes: []TypeInfo{TypeVec3}, ReturnType: TypeVec3},
			{ParamTypes: []TypeInfo{TypeVec4}, ReturnType: TypeVec4},
		},
	},
	"length": {
		WGSLName: "length",
		Overloads: []*BuiltinOverload{
			{ParamTypes: []TypeInfo{TypeVec2}, ReturnType: TypeF32},
			{ParamTypes: []TypeInfo{TypeVec3}, ReturnType: TypeF32},
			{ParamTypes: []TypeInfo{TypeVec4}, ReturnType: TypeF32},
		},
	},
	"normalize": {
		WGSLName: "normalize",
		Overloads: []*BuiltinOverload{
			{ParamTypes: []TypeInfo{TypeVec2}, ReturnType: TypeVec2},
			{ParamTypes: []TypeInfo{TypeVec3}, ReturnType: TypeVec3},
			{ParamTypes: []TypeInfo{TypeVec4}, ReturnType: TypeVec4},
		},
	},
	"radians": {
		WGSLName: "radians",
		Overloads: []*BuiltinOverload{
			{ParamTypes: []TypeInfo{TypeF32}, ReturnType: TypeF32},
			{ParamTypes: []TypeInfo{TypeVec2}, ReturnType: TypeVec2},
			{ParamTypes: []TypeInfo{TypeVec3}, ReturnType: TypeVec3},
			{ParamTypes: []TypeInfo{TypeVec4}, ReturnType: TypeVec4},
		},
	},
	"sign": {
		WGSLName: "sign",
		Overloads: []*BuiltinOverload{
			{ParamTypes: []TypeInfo{TypeF32}, ReturnType: TypeF32},
			{ParamTypes: []TypeInfo{TypeVec2}, ReturnType: TypeVec2},
			{ParamTypes: []TypeInfo{TypeVec3}, ReturnType: TypeVec3},
			{ParamTypes: []TypeInfo{TypeVec4}, ReturnType: TypeVec4},
		},
	},
	"sinh": {
		WGSLName: "sinh",
		Overloads: []*BuiltinOverload{
			{ParamTypes: []TypeInfo{TypeF32}, ReturnType: TypeF32},
			{ParamTypes: []TypeInfo{TypeVec2}, ReturnType: TypeVec2},
			{ParamTypes: []TypeInfo{TypeVec3}, ReturnType: TypeVec3},
			{ParamTypes: []TypeInfo{TypeVec4}, ReturnType: TypeVec4},
		},
	},
	"tanh": {
		WGSLName: "tanh",
		Overloads: []*BuiltinOverload{
			{ParamTypes: []TypeInfo{TypeF32}, ReturnType: TypeF32},
			{ParamTypes: []TypeInfo{TypeVec2}, ReturnType: TypeVec2},
			{ParamTypes: []TypeInfo{TypeVec3}, ReturnType: TypeVec3},
			{ParamTypes: []TypeInfo{TypeVec4}, ReturnType: TypeVec4},
		},
	},
	"trunc": {
		WGSLName: "trunc",
		Overloads: []*BuiltinOverload{
			{ParamTypes: []TypeInfo{TypeF32}, ReturnType: TypeF32},
			{ParamTypes: []TypeInfo{TypeVec2}, ReturnType: TypeVec2},
			{ParamTypes: []TypeInfo{TypeVec3}, ReturnType: TypeVec3},
			{ParamTypes: []TypeInfo{TypeVec4}, ReturnType: TypeVec4},
		},
	},
	// Mesh-CSG operators are export-only; the WGSL preview lowers them
	// to the equivalent SDF distance-field math so the user sees a
	// continuous visualization while authoring. meshDifference has no
	// direct WGSL alias (we want max(L, -R)), so it gets a custom
	// emitter in codegen/function_call.go and is intentionally absent
	// from this map.
	"meshUnion": {
		WGSLName: "min",
		Overloads: []*BuiltinOverload{
			{ParamTypes: []TypeInfo{TypeF32, TypeF32}, ReturnType: TypeF32},
		},
	},
	"meshIntersect": {
		WGSLName: "max",
		Overloads: []*BuiltinOverload{
			{ParamTypes: []TypeInfo{TypeF32, TypeF32}, ReturnType: TypeF32},
		},
	},
}

func FindBuiltin(name string, args []build.Expression) (*Builtin, *BuiltinOverload, error) {
	builtin, ok := wgslBuiltins[name]
	if !ok {
		return nil, nil, fmt.Errorf("builtin %s not found", name)
	}

	argTypes := make([]TypeInfo, len(args))
	for i, arg := range args {
		if w, ok := arg.(Expression); ok {
			argTypes[i] = w.WGSLType()
		} else {
			argTypes[i] = WGSLType(arg.Type(), false)
		}
	}
	for _, overload := range builtin.Overloads {
		if overload.hasCommonParamTypes(argTypes) {
			return builtin, overload, nil
		}
	}

	sb := strings.Builder{}
	sb.WriteString("builtin ")
	sb.WriteString(name)
	sb.WriteString(" not found for arguments:\n  ")
	sb.WriteString(buildArgsSig(args))
	sb.WriteString("\nrequires arguments of type:\n")
	for _, overload := range builtin.Overloads {
		sb.WriteString("  ")
		sb.WriteString(overload.buildParamSig())
		sb.WriteString("\n")
	}

	return nil, nil, errors.New(sb.String())
}

func (overload *BuiltinOverload) hasCommonParamTypes(argTypes []TypeInfo) bool {
	if len(overload.ParamTypes) != len(argTypes) {
		return false
	}
	for i, argType := range argTypes {
		if !overload.ParamTypes[i].Equal(argType) && !canPromote(argType, overload.ParamTypes[i]) {
			return false
		}
	}
	return true
}

func canPromote(from, to TypeInfo) bool {
	if from.Equal(to) {
		return true
	}
	if from.IsNumericScalar() && to.IsNumericVector() {
		return true
	}
	if from.IsBoolScalar() && to.IsBoolVector() {
		return true
	}
	return false
}

func buildArgsSig(arguments []build.Expression) string {
	sb := strings.Builder{}
	separator := ""
	for _, arg := range arguments {
		sb.WriteString(separator)
		sb.WriteString(arg.Type().String())
		separator = ", "
	}
	return sb.String()
}

func (overload *BuiltinOverload) buildParamSig() string {
	paramTypeNames := make([]string, len(overload.ParamTypes))
	for i, paramType := range overload.ParamTypes {
		paramTypeNames[i] = paramType.String()
	}
	return strings.Join(paramTypeNames, ", ")
}
