package wgsl

import (
	"fmt"

	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// LogicOp implements the [build.Expression] interface.
type LogicOp struct {
	Expression
	Operands []Expression
	Symbol   string
}

func NewLogicOp(op *program.Operand, expressions []build.Expression, precedence build.Precedence, resultType *program.Type, symbol string) (*LogicOp, error) {
	wgslExprs := make([]Expression, len(expressions))
	wgslExprs[0] = bracketOtherLogic(expressions[0].(Expression).ToBool().ParenIf(precedence), precedence)
	for i := 1; i < len(expressions); i++ {
		e := expressions[i].(Expression).ToBool()
		if e.Precedence() >= precedence {
			e = Paren(e)
		}
		wgslExprs[i] = bracketOtherLogic(e, precedence)
	}

	wResultType := WGSLType(resultType, true)
	var boolType TypeInfo
	switch wResultType.Name {
	case TypeBool.Name, TypeF32.Name:
		boolType = TypeBool
	case TypeVec2Bool.Name, TypeVec2.Name:
		boolType = TypeVec2Bool
	case TypeVec3Bool.Name, TypeVec3.Name:
		boolType = TypeVec3Bool
	case TypeVec4Bool.Name, TypeVec4.Name:
		boolType = TypeVec4Bool
	default:
		return nil, fmt.Errorf("unhandled boolean operation result type: %v", resultType)
	}

	l := &LogicOp{
		Operands: wgslExprs,
		Symbol:   symbol,
	}
	l.Expression = NewExpr(op, boolType, precedence)
	l.Expression.SetOuter(l)
	return l, nil
}

func (l *LogicOp) Codegen(ctx CodegenContext) string {
	src := l.Operands[0].Codegen(ctx)
	for i := 1; i < len(l.Operands); i++ {
		src += " " + l.Symbol + " " + l.Operands[i].Codegen(ctx)
	}
	return src
}

func (l *LogicOp) Reads() []string {
	return dedupReads(l.Operands...)
}

// bracketOtherLogic brackets an operand of a logical operation that is the
// other logical operation: WGSL's grammar does not let && and || stand beside
// each other unbracketed, whatever their precedence, so a || b && c, which
// precedence alone would print, is refused (by Tint, though naga accepts it).
func bracketOtherLogic(e Expression, precedence build.Precedence) Expression {
	switch e.Precedence() {
	case build.PrecedenceLogicalAnd, build.PrecedenceLogicalOr:
		if e.Precedence() != precedence {
			return Paren(e)
		}
	}
	return e
}
