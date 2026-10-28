package wgsl

import (
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// OpaqueExpr represents a pre-formatted expression or literal token with known reads.
type OpaqueExpr struct {
	// Read-only immutable fields set at construction
	Expression
	src   string
	reads []string
}

func NewOpaqueExpr(op *program.Operand, exprType TypeInfo, precedence build.Precedence, src string, reads ...string) *OpaqueExpr {
	o := &OpaqueExpr{
		src:   src,
		reads: reads,
	}
	o.Expression = NewExprWithReads(op, exprType, precedence, reads...)
	o.Expression.SetOuter(o)
	return o
}

func (o *OpaqueExpr) Codegen(ctx CodegenContext) string {
	return o.src
}

func (o *OpaqueExpr) Reads() []string {
	return append([]string(nil), o.reads...)
}
