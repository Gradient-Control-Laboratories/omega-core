package parser2

import (
	"errors"
	"fmt"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"gradientcontrol.com/omega/core/internal/scope"
	"gradientcontrol.com/omega/core/program"
)

func buildExpressions(buildScope scope.BuilderScope, expressions []*Expression) ([]*program.Operand, error) {
	operands := []*program.Operand{}
	var err error
	for _, expr := range expressions {
		var operand *program.Operand
		operand, err = buildExpression(buildScope, expr)
		if err != nil {
			// break to return the error along with what we have so far
			break
		}
		operands = append(operands, operand)
	}
	return operands, err
}

func buildExpression(buildScope scope.BuilderScope, expression *Expression) (*program.Operand, error) {
	if expression == nil {
		return nil, errors.New("missing expression")
	}

	err := buildScope.IncExprDepth()
	if err != nil {
		return nil, err
	}
	defer buildScope.DecExprDepth()

	fixTernaryAssociativity(expression)

	value, err := buildLogicOr(buildScope, expression.Value)
	if err != nil {
		return value, err
	}

	ternary := expression.Ternary
	if ternary == nil {
		return value, err
	}

	if ternary.Positive == nil {
		return value, &participle.ParseError{
			Msg: "Expected expression after question mark",
			Pos: ternary.QMark.Pos,
		}
	}

	positive, err := buildExpression(buildScope, ternary.Positive)
	if err != nil {
		return value, err
	}
	if !ternary.Colon1 {
		mark1 := ternary.Mark1
		var pos lexer.Position
		if mark1 != nil {
			pos = mark1.Pos
		} else {
			pos = ternary.QMark.Pos
		}
		return value, &participle.ParseError{
			Msg: "Expected \":\" after expression after \"?\"",
			Pos: pos,
		}
	}

	if ternary.Equal == nil {
		return program.NewExpression(
			program.Operation_SELECT, []*program.Operand{value, positive},
			value.Start, positive.End,
		), nil
	}

	equal, err := buildExpression(buildScope, ternary.Equal)
	if err != nil {
		return value, err
	}

	if ternary.Negative == nil {
		if !ternary.Colon2 && ternary.Equal != nil {
			// Optional colon 2 logic.
		}
		return program.NewExpression(
			program.Operation_SELECT, []*program.Operand{value, positive, equal},
			value.Start, equal.End,
		), nil
	}

	negative, err := buildExpression(buildScope, ternary.Negative)
	if err != nil {
		return value, err
	}

	return program.NewExpression(
		program.Operation_SELECT, []*program.Operand{value, positive, equal, negative},
		value.Start, negative.End,
	), nil
}

// fixTernaryAssociativity corrects the Abstract Syntax Tree (AST) for nested unparenthesized
// ternary expressions. Because the parser's grammar is inherently greedy and right-associative,
// when it encounters a nested ternary without parentheses (e.g., `a ? b ? c : d : e ? f : g`),
// the inner ternary branches eagerly consume subsequent branches that structurally belong to
// the outer ternary.
//
// This function recursively traverses the AST and rotates these "stolen" branches (typically
// parsed as the `Negative` branch of an inner expression) back up to the correct scope in the
// parent ternary (acting as the parent's `Equal` or `Negative` branch). It also transfers
// the associated `Colon` flags and positional markers during rotation so that formatting
// constraints (like missing colons separated by newlines) can be accurately validated later.
func fixTernaryAssociativity(expression *Expression) {
	if expression == nil || expression.Ternary == nil {
		return
	}

	fixTernaryAssociativity(expression.Ternary.Positive)

	if expression.Ternary.Positive != nil && expression.Ternary.Positive.Ternary != nil {
		pos := expression.Ternary.Positive
		if pos.Ternary.Negative != nil && expression.Ternary.Equal == nil {
			expression.Ternary.Equal = pos.Ternary.Negative
			expression.Ternary.Colon1 = pos.Ternary.Colon2
			expression.Ternary.Mark1 = pos.Ternary.Mark2
			pos.Ternary.Negative = nil
			pos.Ternary.Colon2 = false
			pos.Ternary.Mark2 = nil
		}
	}

	fixTernaryAssociativity(expression.Ternary.Equal)

	if expression.Ternary.Equal != nil && expression.Ternary.Equal.Ternary != nil {
		eq := expression.Ternary.Equal
		if eq.Ternary.Negative != nil && expression.Ternary.Negative == nil {
			expression.Ternary.Negative = eq.Ternary.Negative
			expression.Ternary.Colon2 = eq.Ternary.Colon2
			expression.Ternary.Mark2 = eq.Ternary.Mark2
			eq.Ternary.Negative = nil
			eq.Ternary.Colon2 = false
			eq.Ternary.Mark2 = nil
		}
	}

	fixTernaryAssociativity(expression.Ternary.Negative)
}

func buildLogicOr(buildScope scope.BuilderScope, logicOr *LogicOr) (*program.Operand, error) {
	if logicOr == nil {
		return nil, errors.New("missing logic or")
	}

	left, err := buildLogicAnd(buildScope, logicOr.Left)
	if err != nil || len(logicOr.Right) == 0 {
		return left, err
	}

	operands := []*program.Operand{left}

	for i := 0; i < len(logicOr.Right); i++ {
		logicAnd := logicOr.Right[i]
		var operand *program.Operand
		operand, err = buildLogicAnd(buildScope, logicAnd)
		if err != nil {
			// Return the error along with what we have so far
			break
		}
		operands = append(operands, operand)
	}

	return program.NewExpression(
		program.Operation_OR, operands,
		left.Start, operands[len(operands)-1].End,
	), nil
}

func buildLogicAnd(buildScope scope.BuilderScope, logicAnd *LogicAnd) (*program.Operand, error) {
	if logicAnd == nil {
		return nil, errors.New("missing logic and")
	}

	left, err := buildEquality(buildScope, logicAnd.Left)
	if err != nil || len(logicAnd.Right) == 0 {
		return left, err
	}

	operands := []*program.Operand{left}

	for i := 0; i < len(logicAnd.Right); i++ {
		equality := logicAnd.Right[i]
		var operand *program.Operand
		operand, err = buildEquality(buildScope, equality)
		if err != nil {
			// Return the error along with what we have so far
			break
		}
		operands = append(operands, operand)
	}

	return program.NewExpression(
		program.Operation_AND, operands,
		left.Start, operands[len(operands)-1].End,
	), nil
}

func buildEquality(buildScope scope.BuilderScope, equality *Equality) (*program.Operand, error) {
	if equality == nil {
		return nil, errors.New("missing equality")
	}

	left, err := buildComparison(buildScope, equality.Left)
	if err != nil || equality.Right == nil {
		return left, err
	}

	right, err := buildComparison(buildScope, equality.Right.Comparison)
	if err != nil {
		return left, err
	}

	return newComparison(left, equality.Right.Comparator, right)
}

func buildComparison(buildScope scope.BuilderScope, comparison *Comparison) (*program.Operand, error) {
	if comparison == nil {
		return nil, errors.New("missing comparison")
	}

	left, err := buildTerms(buildScope, comparison.Left)
	if err != nil || comparison.Right == nil {
		return left, err
	}

	right, err := buildTerms(buildScope, comparison.Right.Terms)
	if err != nil {
		return left, err
	}

	return newComparison(left, comparison.Right.Comparator, right)
}

func buildTerms(buildScope scope.BuilderScope, terms *Terms) (*program.Operand, error) {
	if terms == nil {
		return nil, errors.New("missing expression")
	}

	term, err := buildTerm(buildScope, terms.Left)
	if err != nil || terms.Right == nil {
		return term, err
	}

	operands := []*program.Operand{term}
	oper := terms.Right[0].Operator

	lastTerm := terms.Right[len(terms.Right)-1].Term.Right

	pos := terms.Left.Left.Pos
	var endPos lexer.Position
	if len(lastTerm) == 0 {
		endPos = terms.Left.Left.EndPos
	} else {
		endPos = lastTerm[len(lastTerm)-1].Factor.EndPos
	}

	for i := 0; i < len(terms.Right); i++ {
		opTerm := terms.Right[i]
		if oper != opTerm.Operator {
			expr, err := newOperation(oper, operands, pos, endPos)
			if err != nil {
				return term, err
			}
			operands = []*program.Operand{expr}
			oper = opTerm.Operator
		}

		term, err := buildTerm(buildScope, opTerm.Term)
		if err != nil {
			// Return the error along with what we have so far
			expr, exprErr := newOperation(oper, operands, pos, endPos)
			if exprErr != nil {
				return term, exprErr
			}
			return expr, err
		}

		operands = append(operands, term)
	}

	return newOperation(oper, operands, pos, endPos)
}

func buildTerm(buildScope scope.BuilderScope, term *Term) (*program.Operand, error) {
	if term == nil {
		return nil, errors.New("missing term")
	}

	factor, err := buildFactor(buildScope, term.Left)
	if err != nil || term.Right == nil {
		return factor, err
	}

	pos := term.Left.Pos
	endPos := term.Right[len(term.Right)-1].Factor.EndPos

	operands := []*program.Operand{factor}
	oper := term.Right[0].Operator

	for i := 0; i < len(term.Right); i++ {
		opFactor := term.Right[i]
		if oper != opFactor.Operator {
			expr, err := newOperation(oper, operands, pos, endPos)
			if err != nil {
				return factor, err
			}
			operands = []*program.Operand{expr}
			oper = opFactor.Operator
		}

		factor, err := buildFactor(buildScope, opFactor.Factor)
		if err != nil {
			// Return the error along with what we have so far
			expr, exprErr := newOperation(oper, operands, pos, endPos)
			if exprErr != nil {
				return expr, exprErr
			}
			return expr, err
		}

		operands = append(operands, factor)
	}

	return newOperation(oper, operands, pos, endPos)
}

func buildFactor(buildScope scope.BuilderScope, factor *Factor) (*program.Operand, error) {
	if factor == nil {
		return nil, errors.New("missing factor")
	}

	value, err := buildValue(buildScope, factor.Value)
	if err != nil {
		return value, err
	}

	if factor.Access != "" {
		switch factor.Access {
		case program.DTimeAccessor, program.DPAccessor:
			// Forward-mode derivative: .dtime is d/d(time), .dp is the
			// spatial gradient d/d(p). Expanded by the autodiff pass.
			value = newDeriv(value, factor.Access, factor.EndPos)
		case program.SecondAccessor:
			// Second-closest primitive-surface distance. Expanded by the
			// self-computing-property pass.
			value = newSecond(value, factor.EndPos)
		case program.BoxAccessor:
			// Bounding box of the field's solid. Expanded by the
			// self-computing-property pass.
			value = newBox(value, factor.EndPos)
		case program.UAccessor, program.VAccessor:
			// Surface frame coordinate of the field (.u, .v). Expanded by the
			// self-computing-property pass.
			value = newSurfaceFrame(value, factor.Access, factor.EndPos)
		default:
			isAccessor, err := IsAccessor(factor.Access)
			if err != nil {
				return value, &participle.ParseError{
					Pos: lexer.Position{Line: factor.EndPos.Line, Column: factor.EndPos.Column - len(factor.Access)},
					Msg: err.Error(),
				}
			}
			if !isAccessor {
				return value, &participle.ParseError{
					Pos: lexer.Position{Line: factor.EndPos.Line, Column: factor.EndPos.Column - len(factor.Access)},
					Msg: fmt.Sprintf("invalid accessor: %s", factor.Access),
				}
			}
			value = newAccess(value, factor.Access, factor.EndPos)
		}
	}

	if factor.Index != nil {
		index, err := buildExpression(buildScope, factor.Index)
		if err != nil {
			return value, err
		}
		value = program.NewExpression(
			program.Operation_INDEX, []*program.Operand{value, index},
			posToLocation(factor.Pos), posToLocation(factor.EndPos),
		)
	}

	if factor.Negative {
		switch value.T.(type) {
		case *program.Operand_Value:
			value.T.(*program.Operand_Value).Value *= -1
			// Update start position to cover the leading negative sign,
			// allowing the editor UI to rewrite the full literal span.
			value.Start = posToLocation(factor.Pos)
		default:
			value = program.NewExpression(
				program.Operation_NEG, []*program.Operand{value},
				posToLocation(factor.Pos), value.End,
			)
		}
	}

	if !factor.Not {
		return value, nil
	}

	return program.NewExpression(
		program.Operation_NOT, []*program.Operand{value},
		posToLocation(factor.Pos), value.End,
	), nil
}
