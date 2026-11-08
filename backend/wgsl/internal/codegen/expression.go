package codegen

import (
	"gradientcontrol.com/omega/core/backend/wgsl/internal/wgsl"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

func (d *Delegate) BuildMemberAccess(
	op *program.Operand, _ *program.Expression, target build.Expression, member string,
) (build.Expression, error) {
	targetExpr := target.(wgsl.Expression)
	targetType := targetExpr.WGSLType()

	if targetType.IsNumericVector() {
		return wgsl.NewVecMemberAccess(member, op, targetExpr)
	}

	// array access
	return wgsl.NewArrayMemberAccess(member, op, targetExpr)
}

func (d *Delegate) BuildIndexAccess(op *program.Operand, _ *program.Expression,
	targetOp *program.Operand, _ *program.Operand,
	target build.Expression, index int) (build.Expression, error) {

	// common build verifies that target is a collection

	targetExpr := target.(wgsl.Expression)
	targetType := targetExpr.WGSLType()

	if targetType.ElemType != nil && !targetType.ElemType.IsNumericScalar() {
		return wgsl.NewListIndexedAccess(op, targetExpr, index, *targetType.ElemType)
	}
	// list-of-vectors access yields the element's vector type
	if targetOp != nil && targetOp.Type != nil && targetOp.Type.ElemType != nil &&
		(targetOp.Type.ElemType.IsList() || targetOp.Type.ElemType.IsMatrix()) {
		elemType := wgsl.WGSLType(targetOp.Type.ElemType, false)
		return wgsl.NewListIndexedAccess(op, targetExpr, index, elemType)
	}

	// vector access
	if targetType.IsNumericVector() {
		return wgsl.NewVecIndexedAccess(op, targetExpr, index)
	}

	// array access
	return wgsl.NewArrayIndexedAccess(op, targetExpr, index)
}

func (d *Delegate) BuildVarRef(op *program.Operand, varInfo build.VarInfo) (build.Expression, error) {
	ref := varInfo.(*wgsl.VarInfo)
	d.recordRef(ref)
	return wgsl.NewVarRef(op, ref), nil
}

func (d *Delegate) BuildValue(op *program.Operand, value float64) (build.ConstExpression, error) {
	if param, ok := d.uiParam(op); ok {
		d.recordRef(param)
		return wgsl.NewVarRef(op, param), nil
	}
	return wgsl.NewValue(op, value), nil
}

func (d *Delegate) uiParam(op *program.Operand) (*wgsl.VarInfo, bool) {
	span, ok := d.uiParamOperands[op]
	return d.uiParamVars[span], ok
}

func (d *Delegate) IsParameter(op *program.Operand) bool {
	if op == nil {
		return false
	}
	if _, ok := d.uiParam(op); ok {
		return true
	}
	if ref, ok := op.T.(*program.Operand_Reference); ok && ref != nil && ref.Reference != nil {
		if ref.Reference.PkgPath != "" && ref.Reference.PkgPath != "builtin" {
			return true
		}
		if d.sliderVars[ref.Reference.Name] {
			return true
		}
	}
	return false
}
