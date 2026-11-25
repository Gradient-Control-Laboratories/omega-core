package codegen

import (
	"github.com/Gradient-Control-Laboratories/omega-core/backend/wgsl/internal/wgsl"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/build"
	"github.com/Gradient-Control-Laboratories/omega-core/program"
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
