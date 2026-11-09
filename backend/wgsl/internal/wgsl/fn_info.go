package wgsl

import (
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// FnInfo implements the [build.FnInfo] interface.
type FnInfo struct {
	name       string
	wgslName   string
	paramNames []string

	globalRefs     []*VarInfo
	paramTypes     []TypeInfo
	progParamTypes []*program.Type
	returnType     TypeInfo
	progReturnType *program.Type
}

// Ensure that FnInfo implements the build.FnInfo interface.
var _ build.FnInfo = (*FnInfo)(nil)

func NewFnInfo(
	name string, wgslName string,
	paramNames []string, args []build.Expression,
) *FnInfo {
	wgslName, paramTypes, progParamTypes := buildWgslSignature(wgslName, args)
	return &FnInfo{
		name:           name,
		wgslName:       wgslName,
		paramNames:     paramNames,
		paramTypes:     paramTypes,
		progParamTypes: progParamTypes,
	}
}

func buildWgslSignature(wgslName string, args []build.Expression) (string, []TypeInfo, []*program.Type) {
	paramTypes := make([]TypeInfo, 0, len(args))
	progParamTypes := make([]*program.Type, 0, len(args))
	for _, arg := range args {
		argType := arg.Type()
		if argType != nil && argType.FnPath != "" {
			wgslName += "_" + argType.FnPath
		}
		progParamTypes = append(progParamTypes, argType)
		if w, ok := arg.(Expression); ok {
			paramTypes = append(paramTypes, w.WGSLType())
		} else {
			paramTypes = append(paramTypes, WGSLType(argType, false))
		}
	}
	return wgslName, paramTypes, progParamTypes
}

func (f *FnInfo) Name() string {
	return f.name
}

func (f *FnInfo) WgslName() string {
	return SanitizeFnName(f.wgslName)
}

// SetWgslName overrides the generated WGSL function name. Used to re-uniquify
// a name after argument-driven suffixes have been appended.
func (f *FnInfo) SetWgslName(name string) {
	f.wgslName = name
}

func (f *FnInfo) ParamNames() []string {
	return f.paramNames
}

func (f *FnInfo) GlobalRefs() []*VarInfo {
	return f.globalRefs
}

func (f *FnInfo) ParamTypes() []*program.Type {
	return f.progParamTypes
}

func (f *FnInfo) WGSLParamTypes() []TypeInfo {
	return f.paramTypes
}

func (f *FnInfo) ReturnType() *program.Type {
	return f.progReturnType
}

func (f *FnInfo) WGSLReturnType() TypeInfo {
	return f.returnType
}

func (f *FnInfo) Built(globalRefs []*VarInfo, returnType TypeInfo, progReturnType *program.Type) {
	f.globalRefs = globalRefs
	f.returnType = returnType
	f.progReturnType = progReturnType
}
