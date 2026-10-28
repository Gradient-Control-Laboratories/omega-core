package convert

import (
	"gradientcontrol.com/omega/core/backend/internal/fidget/internal/fidget"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

func (d *Delegate) BuildNeg(
	op *program.Operand, expr *program.Expression,
	targetOp *program.Operand, target build.Expression,
) (build.Expression, error) {
	return buildUnaryOperation(d.GetScope(), op, target, target.Type(),
		func(arg1 fidget.TreeIndex) (fidget.TreeIndex, error) {
			return fidget.Neg(arg1), nil
		})
}

func (d *Delegate) BuildNot(
	op *program.Operand, expr *program.Expression,
	targetOp *program.Operand, target build.Expression,
) (build.Expression, error) {
	return markBoolean(buildUnaryOperation(d.GetScope(), op, target, target.Type(),
		func(arg1 fidget.TreeIndex) (fidget.TreeIndex, error) {
			return fidget.Not(arg1), nil
		}))
}
