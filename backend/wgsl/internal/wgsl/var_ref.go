package wgsl

import (
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// VarRef implements the [build.Expression] interface.
type VarRef struct {
	// Read-only immutable fields set at construction
	Expression
	VarInfo *VarInfo
}

func NewVarRef(op *program.Operand, varInfo *VarInfo) *VarRef {
	v := &VarRef{
		VarInfo: varInfo,
	}
	v.Expression = NewExpr(op, varInfo.WGSLType(), build.PrecedenceHighest)
	v.Expression.SetOuter(v)
	return v
}

func (v *VarRef) Type() *program.Type {
	if v.VarInfo != nil && v.VarInfo.Type() != nil {
		return v.VarInfo.Type()
	}
	return v.Expression.Type()
}

func (v *VarRef) ConstValues() []float64 {
	expr := v.VarInfo.Expr()
	if constExpr, ok := expr.(build.ConstExpression); ok {
		return constExpr.ConstValues()
	}
	return nil
}

func (v *VarRef) Codegen(ctx CodegenContext) string {
	if ctx != nil && v.VarInfo != nil {
		ctx.EmitVar(v.VarInfo)
	}
	return v.VarInfo.WGSLName()
}

func (v *VarRef) Reads() []string {
	if v.VarInfo == nil {
		return nil
	}
	return []string{v.VarInfo.WGSLName()}
}
