package wgsl

import (
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// VarInfo implements the [build.VarInfo] interface.
type VarInfo struct {
	bScope   *build.Scope
	name     string
	wgslName string
	expr     Expression
	varType  TypeInfo
	progType *program.Type
	isParam  bool
}

// Ensure that VarInfo implements the build.VarInfo interface.
var _ build.VarInfo = (*VarInfo)(nil)

func NewVar(
	bScope *build.Scope, name string, expr build.Expression, varType TypeInfo,
) *VarInfo {
	wgslName := SanitizeVarName(name)
	if bScope != nil && (bScope.IsTopLevelScope() || bScope.IsFileScope()) {
		var impScope *build.Scope
		curr := bScope
		for curr != nil {
			if curr.IsImportScope() {
				impScope = curr
				break
			}
			curr = curr.Parent()
		}
		if impScope != nil {
			// The scope name is the package path, which may contain characters
			// (e.g. '/') that are illegal in a WGSL identifier; sanitize it before
			// using it as a name prefix.
			wgslName = SanitizeIdent(impScope.Name()) + "_" + wgslName
		}
	}
	var e Expression
	var progType *program.Type
	if expr != nil {
		e = expr.(Expression)
		progType = expr.Type()
	}
	return &VarInfo{
		bScope:   bScope,
		name:     name,
		wgslName: wgslName,
		expr:     e,
		varType:  varType,
		progType: progType,
	}
}

func NewParam(bScope *build.Scope, name string, progType *program.Type, varType TypeInfo) *VarInfo {
	return &VarInfo{
		bScope:   bScope,
		name:     name,
		wgslName: SanitizeVarName(name),
		progType: progType,
		varType:  varType,
		isParam:  true,
	}
}

func NewParamWithExpr(bScope *build.Scope, name string, expr Expression, progType *program.Type, varType TypeInfo) *VarInfo {
	return &VarInfo{
		bScope:   bScope,
		name:     name,
		wgslName: SanitizeVarName(name),
		expr:     expr,
		progType: progType,
		varType:  varType,
		isParam:  true,
	}
}

func NewPVar(bScope *build.Scope, wgslName string, expr Expression) *VarInfo {
	return &VarInfo{
		bScope:   bScope,
		name:     program.PVarName,
		wgslName: wgslName,
		expr:     expr,
		varType:  expr.WGSLType(),
		isParam:  false,
	}
}

func (v *VarInfo) Name() string {
	return v.name
}

func (v *VarInfo) Scope() *build.Scope {
	return v.bScope
}

// WGSLName returns the WGSL name for the variable.
func (v *VarInfo) WGSLName() string {
	return v.wgslName
}

func (v *VarInfo) SetWGSLName(wgslName string) {
	v.wgslName = wgslName
}

func (v *VarInfo) Type() *program.Type {
	if v.progType.IsFunction() {
		return v.progType
	}
	if v.varType.ProgType != nil {
		return v.varType.ProgType
	}
	if v.progType != nil {
		return v.progType
	}
	if v.expr != nil {
		return v.expr.Type()
	}
	return nil
}

func (v *VarInfo) WGSLType() TypeInfo {
	return v.varType
}

func (v *VarInfo) Expr() build.Expression {
	return v.expr
}

// CanRederive reports whether this var carries a re-emittable definition (a
// let statement derived from an expr), as opposed to a bare parameter (p, time, function params)
// whose value only arrives through a signature. Remap uses this to decide which
// captured primary-scope vars re-derive inside the helper vs stay parameters.
func (v *VarInfo) CanRederive() bool {
	return v.expr != nil && !v.isParam
}
