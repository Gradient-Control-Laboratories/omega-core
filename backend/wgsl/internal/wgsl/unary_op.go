package wgsl

import (
	"strings"

	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// UnaryOp represents a unary operation such as negation (-) or logical not (!).
type UnaryOp struct {
	// Read-only immutable fields set at construction
	Expression
	Op  string
	Sub Expression
}

func NewUnaryOp(op *program.Operand, oper string, sub Expression, resultType TypeInfo, precedence build.Precedence) *UnaryOp {
	u := &UnaryOp{
		Op:  oper,
		Sub: sub,
	}
	u.Expression = NewExpr(op, resultType, precedence)
	u.Expression.SetOuter(u)
	return u
}

func (u *UnaryOp) Codegen(ctx CodegenContext) string {
	src := u.Sub.Codegen(ctx)
	if u.Op == "-" && strings.HasPrefix(src, "-") {
		src = "(" + src + ")"
	}
	return u.Op + src
}

func (u *UnaryOp) Reads() []string {
	return u.Sub.Reads()
}
