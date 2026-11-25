package codegen

import (
	"github.com/Gradient-Control-Laboratories/omega-core/backend/wgsl/internal/wgsl"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/build"
	"github.com/Gradient-Control-Laboratories/omega-core/program"
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
