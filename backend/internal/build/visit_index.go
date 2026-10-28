package build

import (
	"fmt"

	"gradientcontrol.com/omega/core/program"
)

func (b *Builder) visitIndex(op *program.Operand, expr *program.Expression, operands []*program.Operand, expressions []Expression) (Expression, error) {
	targetExpr := expressions[0]
	indexExpr := expressions[1]

	targetType := targetExpr.Type()
	if targetType == nil && len(operands) > 0 && operands[0] != nil {
		targetType = operands[0].Type
	}
	indexType := indexExpr.Type()
	if indexType == nil && len(operands) > 1 && operands[1] != nil {
		indexType = operands[1].Type
	}

	if targetType == nil || (!targetType.IsList() && !targetType.IsMatrix()) {
		return nil, ErrLoc(b.delegate.GetScope(), operands[0].Start,
			fmt.Errorf("target must be a collection, got %s", targetType.TypeString()))
	}

	constIndexExpr, ok := indexExpr.(ConstExpression)
	if !ok {
		return nil, ErrLoc(b.delegate.GetScope(), operands[1].Start,
			fmt.Errorf("index must be a constant expression"))
	}
	values := constIndexExpr.ConstValues()

	if len(values) != 1 {
		return nil, ErrLoc(b.delegate.GetScope(), operands[1].Start,
			fmt.Errorf("index must be a const scalar, got %s", indexType.TypeString()))
	}

	f := values[0]
	index := int(f)
	if f != float64(index) {
		return nil, ErrLoc(b.delegate.GetScope(), operands[1].Start,
			fmt.Errorf("index must be an integer, got %f", f))
	}

	if col, ok := targetExpr.(CollectExpression); ok {
		elems := col.Elements()
		if len(elems) > 0 {
			if index >= 0 && index < len(elems) {
				return elems[index], nil
			}
			return nil, ErrLoc(b.delegate.GetScope(), operands[1].Start,
				fmt.Errorf("index must be in range [0, %d), got %d", len(elems), index))
		}
	}

	targetLen := int(targetType.Size)
	if index < 0 || index >= targetLen {
		return nil, ErrLoc(b.delegate.GetScope(), operands[1].Start,
			fmt.Errorf("index must be in range [0, %d), got %d", targetLen, index))
	}

	// A list of vectors indexes to a vector, so it cannot use the flat
	// const-values shortcut.
	if !targetType.IsMatrix() {
		if constTargetExpr, ok := targetExpr.(ConstExpression); ok {
			constValues := constTargetExpr.ConstValues()
			if index < len(constValues) {
				return b.delegate.BuildValue(operands[0], constValues[index])
			}
		}
	}

	return b.delegate.BuildIndexAccess(op, expr, operands[0], operands[1], targetExpr, index)
}
