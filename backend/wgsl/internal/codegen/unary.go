package codegen

import (
	"github.com/Gradient-Control-Laboratories/omega-core/backend/wgsl/internal/wgsl"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/build"
	"github.com/Gradient-Control-Laboratories/omega-core/program"
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
