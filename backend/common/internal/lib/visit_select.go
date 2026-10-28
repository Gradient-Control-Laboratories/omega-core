package build

import (
	"gradientcontrol.com/omega/core/program"
)

func (b *Builder) visitSelect(op *program.Operand, expr *program.Expression, operands []*program.Operand, expressions []Expression) (Expression, error) {
	// parser should ensure that the select expression has 3 or 4 arguments
	var err error
	resultType := expressions[1].Type()
	for i := 2; i < len(expressions); i++ {
		resultType, err = CommonType(resultType, expressions[i].Type())
		if err != nil {
			return nil, ErrLoc(b.delegate.GetScope(), operands[i].Start, err)
		}
	}

	conditionExpr := expressions[0]
	conditionType := conditionExpr.Type()

	// result type must be compatible with condition type
	resultType, err = CommonType(conditionType, resultType)
	if err != nil {
		return nil, ErrLoc(b.delegate.GetScope(), operands[0].Start, err)
	}

	if len(expressions) == 3 {
		// condition ? arg1 : arg2
		return b.delegate.BuildSelect2(op, expr, operands,
			conditionExpr,
			expressions[1], expressions[2], resultType)
	}

	// value ? arg1 : arg2 : arg3
	return b.delegate.BuildSelect3(op, expr, operands,
		conditionExpr,
		expressions[1], expressions[2], expressions[3], resultType)
}
