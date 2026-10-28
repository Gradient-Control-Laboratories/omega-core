package codegen

import (
	"gradientcontrol.com/omega/core/backend/wgsl/internal/lib/internal/wgsl"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

func (d *Delegate) BuildOperation(op *program.Operand, _ *program.Expression,
	symbol string, _ []*program.Operand, expressions []build.Expression,
	resultType *program.Type, precedence build.Precedence,
) (build.Expression, error) {
	if symbol == "%" {
		return wgsl.NewMod(op, expressions, resultType, d.addPredefinedFunc)
	}
	if symbol == "&&" || symbol == "||" {
		return wgsl.NewLogicOp(op, expressions, precedence, resultType, symbol)
	}
	return wgsl.NewNumericOp(op, expressions, precedence, resultType, symbol)
}

func (d *Delegate) BuildOpCall(op *program.Operand, expr *program.Expression,
	_ []*program.Operand, expressions []build.Expression,
	resultType *program.Type,
) (build.Expression, error) {
	return wgsl.NewNumericOpCall(op, expr.Operation, expressions, resultType)
}
