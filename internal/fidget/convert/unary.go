package convert

import (
	"github.com/Gradient-Control-Laboratories/omega-core/internal/build"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/fidget"
	"github.com/Gradient-Control-Laboratories/omega-core/program"
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
