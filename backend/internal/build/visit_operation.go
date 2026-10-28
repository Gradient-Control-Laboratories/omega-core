package build

import (
	"gradientcontrol.com/omega/core/program"
)

type calcConstValueUnaryFunc func(cv ConstExpression) (ConstExpression, error)

type calcConstValueOpFunc func(i0 int, cv0 ConstExpression, i1 int, cv1 ConstExpression) (ConstExpression, error)

func (b *Builder) visitOperation(
	op *program.Operand, expr *program.Expression,
	expressions []Expression, operands []*program.Operand,
	symbol string, precedence Precedence,
	calcConstValue calcConstValueOpFunc,
) (Expression, error) {

	resultType := expressions[0].Type()
	for i := 1; i < len(expressions); i++ {
		var err error
		resultType, err = CommonType(resultType, expressions[i].Type())
		if err != nil {
			return nil, ErrLoc(b.delegate.GetScope(), operands[i].Start, err)
		}
	}
	constIndex := -1
	if constExpr, ok := expressions[0].(ConstExpression); ok && constExpr.ConstValues() != nil {
		constIndex = 0
	}

	for i := 1; i < len(expressions); i++ {
		var err error
		if calcConstValue == nil {
			continue
		}
		if constExpr, ok := expressions[i].(ConstExpression); ok && constExpr.ConstValues() != nil {
			if constIndex == -1 {
				constIndex = i
			} else {
				// Combine the two const values
				expressions[constIndex], err = calcConstValue(
					constIndex, expressions[constIndex].(ConstExpression),
					i, expressions[i].(ConstExpression))

				if err != nil {
					return nil, ErrLoc(b.delegate.GetScope(), operands[constIndex].Start, err)
				}

				// Remove the later expression from the list. The operands
				// slice aliases expr.Operands in the shared parsed model, so
				// the removal must build a new slice; an in-place append
				// would shift the model's own operand list and corrupt every
				// subsequent build of this expression (functions are rebuilt
				// per call site).
				expressions = append(expressions[:i], expressions[i+1:]...)
				operands = append(append(make([]*program.Operand, 0, len(operands)-1), operands[:i]...), operands[i+1:]...)

				// decrement i now that the element has been removed
				i--
			}
		}
	}

	if len(expressions) == 1 {
		return expressions[0], nil
	}

	return b.delegate.BuildOperation(op, expr, symbol, operands, expressions, resultType, precedence)
}

func (b *Builder) visitUnaryCall(
	op *program.Operand, expr *program.Expression,
	operands []*program.Operand, expressions []Expression,
	calcConstValue calcConstValueUnaryFunc,
) (Expression, error) {

	if calcConstValue != nil {
		if constExpr, ok := expressions[0].(ConstExpression); ok && constExpr.ConstValues() != nil {
			constValue, err := calcConstValue(constExpr)
			if err != nil {
				return nil, ErrLoc(b.delegate.GetScope(), operands[0].Start, err)
			}
			return constValue, nil
		}
	}

	returnType := expressions[0].Type()
	return b.delegate.BuildOpCall(op, expr, operands, expressions, returnType)
}

func (b *Builder) visitOpCall(
	op *program.Operand, expr *program.Expression,
	operands []*program.Operand, expressions []Expression,
	calcConstValue calcConstValueOpFunc,
) (Expression, error) {

	// parser should ensure that there are the correct number of arguments
	returnType := expressions[0].Type()
	for i := 1; i < len(expressions); i++ {
		var err error
		returnType, err = CommonType(returnType, expressions[i].Type())
		if err != nil {
			return nil, ErrLoc(b.delegate.GetScope(), operands[i].Start, err)
		}
	}
	constIndex := -1
	if constExpr, ok := expressions[0].(ConstExpression); ok && constExpr.ConstValues() != nil {
		constIndex = 0
	}

	for i := 1; i < len(expressions); i++ {
		var err error
		if calcConstValue == nil {
			continue
		}
		if constExpr, ok := expressions[i].(ConstExpression); ok && constExpr.ConstValues() != nil {
			if constIndex == -1 {

				constIndex = i
			} else {
				// Combine the two const values
				expressions[constIndex], err = calcConstValue(
					constIndex, expressions[constIndex].(ConstExpression),
					i, expressions[i].(ConstExpression))

				if err != nil {
					return nil, ErrLoc(b.delegate.GetScope(), operands[constIndex].Start, err)
				}

				// Remove the later expression from the list. See the matching
				// comment in visitOperation: operands aliases expr.Operands,
				// so removal must not mutate the shared parsed program.
				expressions = append(expressions[:i], expressions[i+1:]...)
				operands = append(append(make([]*program.Operand, 0, len(operands)-1), operands[:i]...), operands[i+1:]...)

				// decrement i now that the element has been removed
				i--
			}
		}
	}

	if len(expressions) == 1 {
		return expressions[0], nil
	}

	return b.delegate.BuildOpCall(op, expr, operands, expressions, returnType)
}

func (b *Builder) calcConstValue(opFn func(i0 int, v0 float64, i1 int, v1 float64) (float64, error)) calcConstValueOpFunc {
	return func(i0 int, cv0 ConstExpression, i1 int, cv1 ConstExpression) (ConstExpression, error) {
		vals0 := cv0.ConstValues()
		vals1 := cv1.ConstValues()
		if vals0 == nil || vals1 == nil {
			return nil, nil
		}
		count := max(len(vals0), len(vals1))
		result := make([]float64, 0, count)
		for i := 0; i < count; i++ {
			v0 := vals0[min(i, len(vals0)-1)]
			v1 := vals1[min(i, len(vals1)-1)]
			v, err := opFn(i0, v0, i1, v1)
			if err != nil {
				return nil, err
			}
			result = append(result, v)
		}

		op := cv0.Operand()
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

func (b *Builder) calcConstValueUnary(opFn func(v float64) (float64, error)) calcConstValueUnaryFunc {
	return func(cv ConstExpression) (ConstExpression, error) {
		vals := cv.ConstValues()
		if vals == nil {
			return nil, nil
		}
		result := make([]float64, 0, len(vals))
		for _, val := range vals {
			v, err := opFn(val)
			if err != nil {
				return nil, err
			}
			result = append(result, v)
		}

		op := cv.Operand()
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
