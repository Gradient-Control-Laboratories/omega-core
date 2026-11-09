package build

import (
	"fmt"

	"gradientcontrol.com/omega/core/program"
)

func (b *Builder) visitCollect(op *program.Operand, mExpr *program.Expression,
	operands []*program.Operand, expressions []Expression) (Expression, error) {

	elemVecType, i, err := collectElemVecType(operands, expressions)
	if err != nil {
		return nil, ErrLoc(b.delegate.GetScope(), operands[i].Start, err)
	}

	// A list of vectors is never constant-folded, so it skips the const path.
	if elemVecType != nil {
		return b.delegate.BuildCollect(op, mExpr, operands, expressions)
	}

	return b.buildCollect(op, mExpr, operands, expressions)
}

// collectElemVecType validates the element types of a prospective collection:
// the elements must either all be scalars or all be vectors of the same
// length. It returns the first element's vector type when the elements are
// vectors (so the collection is a list of vectors), or nil when they are
// scalars. On a mismatch it returns the offending element's index along with
// the error.
func collectElemVecType(operands []*program.Operand, expressions []Expression) (*program.Type, int, error) {
	var elemVecType *program.Type
	var seenFirst bool
	for i, expr := range expressions {
		exprType := expr.Type()
		if exprType == nil && i < len(operands) && operands[i] != nil {
			exprType = operands[i].Type
		}
		if exprType == nil {
			continue
		}
		if exprType.IsFunction() {
			err := fmt.Errorf("collection can only contain numeric elements, got %s", exprType.TypeString())
			return nil, i, err
		}
		isVecType := exprType.IsList() || exprType.IsMatrix()
		if !seenFirst {
			seenFirst = true
			if isVecType {
				elemVecType = exprType
			}
			continue
		}
		if isVecType != (elemVecType != nil) {
			err := fmt.Errorf("collection elements must be all scalars or all vectors")
			return nil, i, err
		}
		if isVecType && elemVecType != nil && exprType.Size != elemVecType.Size {
			err := fmt.Errorf("collection elements must be vectors of the same length")
			return nil, i, err
		}
	}
	return elemVecType, 0, nil
}

func (b *Builder) buildCollect(
	op *program.Operand, mExpr *program.Expression, operands []*program.Operand, expressions []Expression,
) (Expression, error) {
	constValues := make([]float64, 0, len(expressions))

	for _, expr := range expressions {
		if constExpr, ok := expr.(ConstExpression); ok {
			cv := constExpr.ConstValues()
			if len(cv) == 1 {
				// numeric expressions only have one value
				constValues = append(constValues, cv[0])
				continue
			}
		}

		// If not const, then build standard collection
		return b.delegate.BuildCollect(op, mExpr, operands, expressions)
	}

	// If all expressions in the collection are constant, we can build a constant expression
	return b.delegate.BuildConstCollect(op, mExpr, operands, expressions, constValues)
}
