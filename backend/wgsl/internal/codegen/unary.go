package codegen

import (
	"gradientcontrol.com/omega/core/backend/wgsl/internal/wgsl"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

func (d *Delegate) BuildNot(op *program.Operand, _ *program.Expression,
	_ *program.Operand, target build.Expression,
) (build.Expression, error) {
	boolExpr := target.(wgsl.Expression).ToBool().ParenIf(build.PrecedenceUnary)
	return wgsl.NewUnaryOp(op, "!", boolExpr, boolExpr.WGSLType(), build.PrecedenceHighest), nil
}

func (d *Delegate) BuildNeg(op *program.Operand, _ *program.Expression,
	_ *program.Operand, target build.Expression,
) (build.Expression, error) {
	numExpr := target.(wgsl.Expression).ToNum().ParenIf(build.PrecedenceUnary)
	return wgsl.NewUnaryOp(op, "-", numExpr, numExpr.WGSLType(), build.PrecedenceUnary), nil
}
