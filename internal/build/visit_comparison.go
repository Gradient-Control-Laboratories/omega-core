package build

import (
	"gradientcontrol.com/omega/core/program"
)

type calcConstComparisonFunc func(v0, v1 float64) float64

func (b *Builder) visitComparison(
	op *program.Operand, expr *program.Expression,
	symbol string, expressions []Expression, operands []*program.Operand, precedence Precedence,
	calcConstComparison calcConstComparisonFunc,
) (Expression, error) {

	// parser should ensure that there are 2 expressions... e.g. expr1 < expr2

	expr1 := expressions[0]
	expr2 := expressions[1]

	resultType, err := CommonType(expr1.Type(), expr2.Type())
	if err != nil {
		return nil, ErrLoc(b.delegate.GetScope(), operands[0].Start, err)
	}

	if cv1, ok1 := expr1.(ConstExpression); ok1 && cv1.ConstValues() != nil {
		if cv2, ok2 := expr2.(ConstExpression); ok2 && cv2.ConstValues() != nil {
			vals1 := cv1.ConstValues()
			vals2 := cv2.ConstValues()
			count := max(len(vals1), len(vals2))
			result := make([]float64, count)
			for i := 0; i < count; i++ {
				v1 := vals1[min(i, len(vals1)-1)]
				v2 := vals2[min(i, len(vals2)-1)]
				result[i] = calcConstComparison(v1, v2)
			}

			if len(result) == 1 {
				return b.delegate.BuildValue(op, result[0])
			}

			childExprs := make([]Expression, len(result))
			for i, val := range result {
				child, err := b.delegate.BuildValue(op, val)
				if err != nil {
					return nil, err
				}
				childExprs[i] = child
			}
			return b.delegate.BuildConstCollect(op, nil, nil, childExprs, result)
		}
	}

	return b.delegate.BuildComparison(op, expr, symbol, operands, expressions, resultType, precedence)
}

func (b *Builder) visitCompare(op *program.Operand, expr *program.Expression,
	operands []*program.Operand, expressions []Expression) (Expression, error) {

	// parser should ensure that there are 2 expressions... e.g. expr1 <=> expr2

	expr1 := expressions[0]
	expr2 := expressions[1]

	resultType, err := CommonType(expr1.Type(), expr2.Type())
	if err != nil {
		return nil, ErrLoc(b.delegate.GetScope(), operands[0].Start, err)
	}

	if cv1, ok1 := expr1.(ConstExpression); ok1 && cv1.ConstValues() != nil {
		if cv2, ok2 := expr2.(ConstExpression); ok2 && cv2.ConstValues() != nil {
			vals1 := cv1.ConstValues()
			vals2 := cv2.ConstValues()
			count := max(len(vals1), len(vals2))
			result := make([]float64, count)
			for i := 0; i < count; i++ {
				v1 := vals1[min(i, len(vals1)-1)]
				v2 := vals2[min(i, len(vals2)-1)]
				var res float64
				if v1 < v2 {
					res = -1.0
				} else if v1 > v2 {
					res = 1.0
				} else {
					res = 0.0
				}
				result[i] = res
			}

			if len(result) == 1 {
				return b.delegate.BuildValue(op, result[0])
			}

			childExprs := make([]Expression, len(result))
			for i, val := range result {
				child, err := b.delegate.BuildValue(op, val)
				if err != nil {
					return nil, err
				}
				childExprs[i] = child
			}
			return b.delegate.BuildConstCollect(op, nil, nil, childExprs, result)
		}
	}

	return b.delegate.BuildCompare(op, expr, operands, expr1, expr2, resultType)
}
