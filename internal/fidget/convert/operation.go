package convert

import (
	"fmt"

	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/internal/fidget"
	"gradientcontrol.com/omega/core/program"
)

func (d *Delegate) BuildOperation(
	op *program.Operand, expr *program.Expression, symbol string,
	operands []*program.Operand, expressions []build.Expression,
	resultType *program.Type, precedence build.Precedence,
) (build.Expression, error) {
	switch expr.Operation {

	case program.Operation_ADD:
		return buildNumericOperation(d.GetScope(), op, operands, expressions, resultType,
			func(arg1, arg2 fidget.TreeIndex) (fidget.TreeIndex, error) {
				return fidget.Add(arg1, arg2), nil
			})

	case program.Operation_SUB:
		return buildNumericOperation(d.GetScope(), op, operands, expressions, resultType,
			func(arg1, arg2 fidget.TreeIndex) (fidget.TreeIndex, error) {
				return fidget.Sub(arg1, arg2), nil
			})

	case program.Operation_MUL:
		return buildNumericOperation(d.GetScope(), op, operands, expressions, resultType,
			func(arg1, arg2 fidget.TreeIndex) (fidget.TreeIndex, error) {
				return fidget.Mul(arg1, arg2), nil
			})

	case program.Operation_DIV:
		return buildNumericOperation(d.GetScope(), op, operands, expressions, resultType,
			func(arg1, arg2 fidget.TreeIndex) (fidget.TreeIndex, error) {
				return fidget.Div(arg1, arg2), nil
			})

	case program.Operation_MODULO:
		return buildNumericOperation(d.GetScope(), op, operands, expressions, resultType,
			func(arg1, arg2 fidget.TreeIndex) (fidget.TreeIndex, error) {
				// Clamp divisor's absolute value to prevent interval evaluation
				// over 0, which yields NaN bounds and crashes the MDC mesher.
				safeDivisor := fidget.Max(fidget.Abs(arg2), d.epsilon)
				return fidget.Modulo(arg1, safeDivisor), nil
			})

	// && and || return one of their operands, so they are 0/1 exactly when
	// both operands are.
	case program.Operation_AND:
		e, err := buildNumericOperation(d.GetScope(), op, operands, expressions, resultType,
			func(arg1, arg2 fidget.TreeIndex) (fidget.TreeIndex, error) {
				return fidget.And(arg1, arg2), nil
			})
		return keepBoolean(expressions, e, err)

	case program.Operation_OR:
		e, err := buildNumericOperation(d.GetScope(), op, operands, expressions, resultType,
			func(arg1, arg2 fidget.TreeIndex) (fidget.TreeIndex, error) {
				return fidget.Or(arg1, arg2), nil
			})
		return keepBoolean(expressions, e, err)

	default:
		return nil, fmt.Errorf("unhandled operation: %v", expr.Operation)
	}
}

// keepBoolean marks the result of && or || as 0/1 valued when every operand
// is.
func keepBoolean(operands []build.Expression, e build.Expression, err error) (build.Expression, error) {
	for _, o := range operands {
		if node, ok := o.(*TreeNode); !ok || !node.boolean {
			return e, err
		}
	}
	return markBoolean(e, err)
}

func (d *Delegate) BuildOpCall(op *program.Operand, expr *program.Expression,
	operands []*program.Operand, expressions []build.Expression,
	resultType *program.Type,
) (build.Expression, error) {
	switch expr.Operation {

	case program.Operation_MIN:
		return buildNumericOperation(d.GetScope(), op, operands, expressions, resultType,
			func(arg1, arg2 fidget.TreeIndex) (fidget.TreeIndex, error) {
				return fidget.Min(arg1, arg2), nil
			})

	case program.Operation_MAX:
		return buildNumericOperation(d.GetScope(), op, operands, expressions, resultType,
			func(arg1, arg2 fidget.TreeIndex) (fidget.TreeIndex, error) {
				return fidget.Max(arg1, arg2), nil
			})

	case program.Operation_SQUARE:
		return buildUnaryOperation(d.GetScope(), op, expressions[0], resultType,
			func(arg1 fidget.TreeIndex) (fidget.TreeIndex, error) {
				return fidget.Square(arg1), nil
			})

	case program.Operation_SQRT:
		return buildUnaryOperation(d.GetScope(), op, expressions[0], resultType,
			func(arg1 fidget.TreeIndex) (fidget.TreeIndex, error) {
				// Clamp input to >= 0. Fidget interval engines evaluate all ternary branches
				// unconditionally; negative inputs to Sqrt produce NaN and panic the MDC mesher.
				clamped := fidget.Max(arg1, d.zero)
				return fidget.Sqrt(clamped), nil
			})

	case program.Operation_FLOOR:
		return buildUnaryOperation(d.GetScope(), op, expressions[0], resultType,
			func(arg1 fidget.TreeIndex) (fidget.TreeIndex, error) {
				return fidget.Floor(arg1), nil
			})

	case program.Operation_CEIL:
		return buildUnaryOperation(d.GetScope(), op, expressions[0], resultType,
			func(arg1 fidget.TreeIndex) (fidget.TreeIndex, error) {
				return fidget.Ceil(arg1), nil
			})

	case program.Operation_ROUND:
		return buildUnaryOperation(d.GetScope(), op, expressions[0], resultType,
			func(arg1 fidget.TreeIndex) (fidget.TreeIndex, error) {
				return fidget.Round(arg1), nil
			})

	case program.Operation_ABS:
		return buildUnaryOperation(d.GetScope(), op, expressions[0], resultType,
			func(arg1 fidget.TreeIndex) (fidget.TreeIndex, error) {
				return fidget.Abs(arg1), nil
			})

	case program.Operation_EXP:
		return buildUnaryOperation(d.GetScope(), op, expressions[0], resultType,
			func(arg1 fidget.TreeIndex) (fidget.TreeIndex, error) {
				return fidget.Exp(arg1), nil
			})

	case program.Operation_LN:
		return buildUnaryOperation(d.GetScope(), op, expressions[0], resultType,
			func(arg1 fidget.TreeIndex) (fidget.TreeIndex, error) {
				// Clamp to prevent Ln(negative) from conservatively producing NaN bounds
				return fidget.Ln(fidget.Max(arg1, d.epsilon)), nil
			})

	case program.Operation_POW:
		// Two paths. If the exponent is a constant integer, use Tree::pow's
		// exponentiation by squaring — exact, fast, works on negative bases.
		// Otherwise synthesize pow(b, e) = exp(e * ln(b)) with the base
		// clamped, since negative bases would produce NaN bounds and panic
		// the MDC mesher.
		if n, ok := constInt(operands[1]); ok {
			return buildUnaryOperation(d.GetScope(), op, expressions[0], resultType,
				func(base fidget.TreeIndex) (fidget.TreeIndex, error) {
					return fidget.PowInt(base, n), nil
				})
		}
		return buildBinaryOperation(d.GetScope(), op, operands, expressions, resultType,
			func(base, exponent fidget.TreeIndex) (fidget.TreeIndex, error) {
				return fidget.Exp(fidget.Mul(exponent, fidget.Ln(fidget.Max(base, d.epsilon)))), nil
			})

	case program.Operation_SIN:
		return buildUnaryOperation(d.GetScope(), op, expressions[0], resultType,
			func(arg1 fidget.TreeIndex) (fidget.TreeIndex, error) {
				return fidget.Sin(arg1), nil
			})

	case program.Operation_COS:
		return buildUnaryOperation(d.GetScope(), op, expressions[0], resultType,
			func(arg1 fidget.TreeIndex) (fidget.TreeIndex, error) {
				return fidget.Cos(arg1), nil
			})

	case program.Operation_TAN:
		return buildUnaryOperation(d.GetScope(), op, expressions[0], resultType,
			func(arg1 fidget.TreeIndex) (fidget.TreeIndex, error) {
				return fidget.Tan(arg1), nil
			})

	case program.Operation_ASIN:
		return buildUnaryOperation(d.GetScope(), op, expressions[0], resultType,
			func(arg1 fidget.TreeIndex) (fidget.TreeIndex, error) {
				// Clamp input to [-1, 1]. Fidget interval engines evaluate all ternary branches
				// unconditionally; out-of-domain inputs to Asin produce NaN and panic the MDC mesher.
				clamped := fidget.Min(fidget.Max(arg1, d.negOne), d.one)
				return fidget.Asin(clamped), nil
			})

	case program.Operation_ACOS:
		return buildUnaryOperation(d.GetScope(), op, expressions[0], resultType,
			func(arg1 fidget.TreeIndex) (fidget.TreeIndex, error) {
				// Clamp input to [-1, 1]. Fidget interval engines evaluate all ternary branches
				// unconditionally; out-of-domain inputs to Acos produce NaN and panic the MDC mesher.
				clamped := fidget.Min(fidget.Max(arg1, d.negOne), d.one)
				return fidget.Acos(clamped), nil
			})

	case program.Operation_ATAN:
		return buildUnaryOperation(d.GetScope(), op, expressions[0], resultType,
			func(arg1 fidget.TreeIndex) (fidget.TreeIndex, error) {
				return fidget.Atan(arg1), nil
			})

	case program.Operation_ATAN2:
		return buildBinaryOperation(d.GetScope(), op, operands, expressions, resultType,
			func(arg1, arg2 fidget.TreeIndex) (fidget.TreeIndex, error) {
				return fidget.Atan2(arg1, arg2), nil
			})

	default:
		return nil, fmt.Errorf("unhandled op call: %v", expr.Operation)
	}
}

// constInt returns (n, true) when the operand is a literal whose value is an
// integer in the i64 range, and (0, false) otherwise. Used to take the native
// Tree::pow fast path for integer exponents.
func constInt(operand *program.Operand) (int64, bool) {
	v, ok := operand.GetT().(*program.Operand_Value)
	if !ok {
		return 0, false
	}
	n := int64(v.Value)
	if float64(n) != v.Value {
		return 0, false
	}
	return n, true
}
