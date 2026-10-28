package build

import (
	"fmt"

	"gradientcontrol.com/omega/core/program"
)

func (b *Builder) visitNot(op *program.Operand, mExpr *program.Expression, expressions []Expression, operands []*program.Operand) (Expression, error) {
	expr := expressions[0]

	// parser should ensure that there is exactly one argument
	exprType := expr.Type()
	if exprType == nil && len(operands) > 0 && operands[0] != nil {
		exprType = operands[0].Type
	}
	if !exprType.IsNumeric() {
		return nil, ErrLoc(b.delegate.GetScope(), op.Start, fmt.Errorf("expected numeric or boolean argument, got %s", exprType.TypeString()))
	}

	if constExpr, ok := expr.(ConstExpression); ok {
		constValue, err := b.calcConstValueUnary(func(v float64) (float64, error) {
			if v > 0 {
				return 0, nil
			}
			return 1, nil
		})(constExpr)
		if err != nil {
			return nil, ErrLoc(b.delegate.GetScope(), operands[0].Start, err)
		}
		if constValue != nil {
			return constValue, nil
		}
	}

	return b.delegate.BuildNot(op, mExpr, operands[0], expr)
}

func (b *Builder) visitNeg(op *program.Operand, mExpr *program.Expression, expressions []Expression, operands []*program.Operand) (Expression, error) {
	expr := expressions[0]

	// parser should ensure that there is exactly one argument
	exprType := expr.Type()
	if exprType == nil && len(operands) > 0 && operands[0] != nil {
		exprType = operands[0].Type
	}
	if !exprType.IsNumeric() {
		return nil, ErrLoc(b.delegate.GetScope(), op.Start, fmt.Errorf("expected numeric or boolean argument, got %s", exprType.TypeString()))
	}

	if constExpr, ok := expr.(ConstExpression); ok {
		constValue, err := b.calcConstValueUnary(func(v float64) (float64, error) {
			return -v, nil
		})(constExpr)
		if err != nil {
			return nil, ErrLoc(b.delegate.GetScope(), operands[0].Start, err)
		}
		if constValue != nil {
			return constValue, nil
		}
	}

	return b.delegate.BuildNeg(op, mExpr, operands[0], expr)
}
