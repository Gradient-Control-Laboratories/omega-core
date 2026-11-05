package parser2

import (
	"fmt"

	"strconv"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"

	"gradientcontrol.com/omega/core/internal/scope"
	"gradientcontrol.com/omega/core/program"
)

func buildCollection(buildScope scope.BuilderScope, expressions []*Expression, pos, endPos lexer.Position) (*program.Operand, error) {
	operands := []*program.Operand{}
	var err error
	for _, expr := range expressions {
		if str, ok := extractStringLiteral(expr); ok {
			unquoted, err := strconv.Unquote(str.Value)
			if err != nil {
				return nil, &participle.ParseError{
					Pos: str.Pos,
					Msg: err.Error(),
				}
			}
			if len(unquoted) == 0 {
				return nil, &participle.ParseError{
					Pos: str.Pos,
					Msg: "empty strings are not allowed",
				}
			}
			runes := []rune(unquoted)
			for _, r := range runes {
				operands = append(operands,
					program.NewValue(float64(r), posToLocation(str.Pos), posToLocation(str.EndPos)))
			}
		} else {
			var operand *program.Operand
			operand, err = buildExpression(buildScope, expr)
			if err != nil {
				// break to return the error along with what we have so far
				break
			}
			operands = append(operands, operand)
		}
	}
	if err != nil {
		return nil, err
	}
	if len(operands) < 1 {
		return nil, &participle.ParseError{
			Pos: pos,
			Msg: "collections cannot be empty",
		}
	}
	return program.NewExpression(
		program.Operation_COLLECT, operands,
		posToLocation(pos), posToLocation(endPos),
	), nil
}

func newAccess(target *program.Operand, name string, endPos lexer.Position) *program.Operand {
	return program.NewAccess(target, name, posToLocation(endPos))
}

// newDeriv builds a DERIV operand for a .dtime / .dp postfix accessor. The
// second operand references the variable to differentiate with respect to:
// "time" for .dtime, "p" for .dp. The DERIV converts 1:1 to the model DERIV,
// which the autodiff pass expands into a concrete derivative expression
// before backend lowering.
func newDeriv(target *program.Operand, accessor string, endPos lexer.Position) *program.Operand {
	wrt := program.TimeVarName
	if accessor == program.DPAccessor {
		wrt = program.PVarName
	}
	endLoc := posToLocation(endPos)
	refStart := program.NewLoc(
		int32(endPos.Line),
		int32(endPos.Column-len(accessor)),
		int32(endPos.Offset-len(accessor)),
	)
	wrtRef := program.NewReference("", wrt, refStart, endLoc)
	return program.NewExpression(
		program.Operation_DERIV,
		[]*program.Operand{target, wrtRef},
		target.Start, endLoc,
	)
}

// newDerivWrt builds a DERIV operand for the derivative of target with respect
// to a named variable: the postfix `.d(var)`, or `name.d(var)`, which reaches
// the build as a qualified call whose prefix is not an import. As for .dtime
// and .dp the second operand references the variable; the autodiff pass
// expands the DERIV before backend lowering, and is where the name is checked
// to be a top-level scalar variable.
func newDerivWrt(buildScope scope.BuilderScope, target *program.Operand, args []*Expression, pos, endPos lexer.Position) (*program.Operand, error) {
	if len(args) != 1 {
		return nil, &participle.ParseError{
			Pos: pos,
			Msg: fmt.Sprintf(".d takes one argument, the variable to differentiate with respect to, not %d", len(args)),
		}
	}
	wrt, err := buildExpression(buildScope, args[0])
	if err != nil {
		return nil, err
	}
	if r, ok := wrt.T.(*program.Operand_Reference); !ok || r.Reference.Prefix != "" {
		return nil, &participle.ParseError{
			Pos: args[0].Pos,
			Msg: ".d takes the name of a variable to differentiate with respect to",
		}
	}
	return program.NewExpression(
		program.Operation_DERIV,
		[]*program.Operand{target, wrt},
		target.Start, posToLocation(endPos),
	), nil
}

// newSecond builds a SECONDCLOSEST operand for a .second postfix accessor: the
// second-closest primitive-surface distance of the target field. It converts
// 1:1 to the model SECONDCLOSEST, which the self-computing-property pass expands
// into a concrete field before backend lowering.
func newSecond(target *program.Operand, endPos lexer.Position) *program.Operand {
	return program.NewExpression(
		program.Operation_SECONDCLOSEST,
		[]*program.Operand{target},
		target.Start, posToLocation(endPos),
	)
}

// newBox builds a BOUNDINGBOX operand for a .box postfix accessor: the bounding
// box of the target field's solid, which the self-computing-property pass
// expands to [lo, hi] before backend lowering.
func newBox(target *program.Operand, endPos lexer.Position) *program.Operand {
	return program.NewExpression(
		program.Operation_BOUNDINGBOX,
		[]*program.Operand{target},
		target.Start, posToLocation(endPos),
	)
}

// newSurfaceFrame builds a SURFACEU or SURFACEV operand for a .u or .v postfix
// accessor: where a point sits on the target field's own surface, which the
// self-computing-property pass expands to the target's frame coordinate before
// backend lowering.
func newSurfaceFrame(target *program.Operand, accessor string, endPos lexer.Position) *program.Operand {
	operation := program.Operation_SURFACEU
	if accessor == program.VAccessor {
		operation = program.Operation_SURFACEV
	}
	return program.NewExpression(
		operation,
		[]*program.Operand{target},
		target.Start, posToLocation(endPos),
	)
}

func newComparison(left *program.Operand,
	comparator Comparator, right *program.Operand,
) (*program.Operand, error) {
	operation, err := comparatorToOperation(comparator)
	if err != nil {
		return nil, err
	}
	return program.NewExpression(
		operation, []*program.Operand{left, right},
		left.Start, right.End,
	), nil
}

func comparatorToOperation(comparator Comparator) (program.Operation, error) {
	switch comparator {

	case CompGT:
		return program.Operation_GT, nil
	case CompGE:
		return program.Operation_GE, nil

	case CompEQ:
		return program.Operation_EQ, nil
	case CompNE:
		return program.Operation_NE, nil

	case CompLE:
		return program.Operation_LE, nil
	case CompLT:
		return program.Operation_LT, nil

	case CompCompare:
		return program.Operation_COMPARE, nil

	default:
		return program.Operation_UNSPECIFIED_OPERAND,
			fmt.Errorf("unhandled comparator: %v", comparator)
	}
}

func newOperation(oper Operator, operands []*program.Operand, pos, endPos lexer.Position) (*program.Operand, error) {
	operation, err := operatorToOperation(oper)
	if err != nil {
		return nil, err
	}

	return program.NewExpression(
		operation, operands,
		posToLocation(pos), posToLocation(endPos),
	), nil
}

func operatorToOperation(oper Operator) (program.Operation, error) {
	var operation program.Operation
	switch oper {
	case OpAdd:
		operation = program.Operation_ADD
	case OpSub:
		operation = program.Operation_SUB
	case OpMul:
		operation = program.Operation_MUL
	case OpDiv:
		operation = program.Operation_DIV
	case OpMod:
		operation = program.Operation_MODULO
	default:
		return 0, fmt.Errorf("invalid operator: %v", oper)
	}
	return operation, nil
}

func posToLocation(pos lexer.Position) *program.Location {
	return &program.Location{
		Line:   int32(pos.Line),
		Column: int32(pos.Column),
		Offset: int32(pos.Offset),
	}
}
