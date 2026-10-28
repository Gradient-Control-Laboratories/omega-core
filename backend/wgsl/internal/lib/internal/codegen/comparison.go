package codegen

import (
	"gradientcontrol.com/omega/core/backend/wgsl/internal/lib/internal/wgsl"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

func (d *Delegate) BuildComparison(
	op *program.Operand, expr *program.Expression,
	symbol string, _ []*program.Operand, expressions []build.Expression,
	resultType *program.Type, precedence build.Precedence,
) (build.Expression, error) {
	return wgsl.NewComparison(op, expr.Operation, symbol, expressions, resultType, precedence)
}

func (d *Delegate) BuildCompare(
	op *program.Operand, _ *program.Expression,
	_ []*program.Operand, expr1 build.Expression, expr2 build.Expression,
	resultType *program.Type,
) (build.Expression, error) {
	return wgsl.NewCompare(op, expr1, expr2, resultType, d.addPredefinedFunc)
}
