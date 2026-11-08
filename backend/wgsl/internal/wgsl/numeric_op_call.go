package wgsl

import (
	"fmt"

	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// NumericOpCall implements the [build.Expression] interface.
type NumericOpCall struct {
	Expression
	Name string
	Args []Expression
}

func NewNumericOpCall(op *program.Operand, exprOp program.Operation, expressions []build.Expression, resultType *program.Type) (Expression, error) {
	if exprOp == program.Operation_SQUARE {
		return NewSquare(op, expressions[0], resultType)
	}

	var name string
	switch exprOp {
	case program.Operation_MIN:
		name = "min"
	case program.Operation_MAX:
		name = "max"
	case program.Operation_SQRT:
		name = "sqrt"
	case program.Operation_FLOOR:
		name = "floor"
	case program.Operation_CEIL:
		name = "ceil"
	case program.Operation_ROUND:
		name = "round"
	case program.Operation_ABS:
		name = "abs"
	case program.Operation_EXP:
		name = "exp"
	case program.Operation_LN:
		name = "log"
	case program.Operation_POW:
		name = "pow"
	case program.Operation_SIN:
		name = "sin"
	case program.Operation_COS:
		name = "cos"
	case program.Operation_TAN:
		name = "tan"
	case program.Operation_ASIN:
		name = "asin"
	case program.Operation_ACOS:
		name = "acos"
	case program.Operation_ATAN:
		name = "atan"
	case program.Operation_ATAN2:
		name = "atan2"
	default:
		return nil, fmt.Errorf("unhandled operation: %s", exprOp)
	}

	rType := WGSLType(resultType, false)
	if rType.Name == "" {
		rType = TypeF32
	}

	// Apparently numeric consts are considered to be abstract values
	// and cannot be automatically converted to vectors if they don't have a type.
	// In that case, we need to convert them to the correct type.
	wgslExprs := make([]Expression, len(expressions))
	for i, expression := range expressions {
		argExpr, err := expression.(Expression).ToNum().ToCommon(rType)
		if err != nil {
			return nil, err
		}
		wgslExprs[i] = argExpr
	}

	n := &NumericOpCall{
		Name: name,
		Args: wgslExprs,
	}
	n.Expression = NewExpr(op, rType, build.PrecedenceHighest)
	n.Expression.SetOuter(n)
	return n, nil
}

func (n *NumericOpCall) Codegen(ctx CodegenContext) string {
	if len(n.Args) == 1 {
		return fmt.Sprintf("%s(%s)", n.Name, n.Args[0].Codegen(ctx))
	}

	src := fmt.Sprintf("%s(%s, %s)", n.Name, n.Args[0].Codegen(ctx), n.Args[1].Codegen(ctx))
	for i := 2; i < len(n.Args); i++ {
		src = fmt.Sprintf("%s(%s, %s)", n.Name, src, n.Args[i].Codegen(ctx))
	}
	return src
}

func (n *NumericOpCall) Reads() []string {
	return dedupReads(n.Args...)
}
