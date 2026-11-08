package codegen

import (
	"gradientcontrol.com/omega/core/backend/wgsl/internal/wgsl"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

func (d *Delegate) BuildSelect2(op *program.Operand, _ *program.Expression, _ []*program.Operand,
	conditionExpr build.Expression,
	trueExpr build.Expression, falseExpr build.Expression, resultType *program.Type,
) (build.Expression, error) {

	return wgsl.NewSelect2(op, conditionExpr, trueExpr, falseExpr, resultType)
}

func (d *Delegate) BuildSelect3(op *program.Operand, _ *program.Expression, _ []*program.Operand,
	valueExpr build.Expression,
	ltZeroResultExpr build.Expression, eqZeroResultExpr build.Expression, gtZeroResultExpr build.Expression,
	resultType *program.Type) (build.Expression, error) {

	return wgsl.NewSelect3(op, valueExpr, ltZeroResultExpr, eqZeroResultExpr, gtZeroResultExpr, resultType, d.addPredefinedFunc)
}
