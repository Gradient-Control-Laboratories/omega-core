package wgsl

import (
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// WithPExpr overrides the P variable in the context when generating code.
type WithPExpr struct {
	// Read-only immutable fields set at construction
	Expression
	Target Expression
	PVar   *VarInfo
	reads  []string
}

func NewWithPExpr(op *program.Operand, target Expression, pVar *VarInfo, reads []string) *WithPExpr {
	w := &WithPExpr{
		Target: target,
		PVar:   pVar,
		reads:  reads,
	}
	w.Expression = NewExprWithReads(op, target.WGSLType(), build.PrecedenceHighest, reads...)
	w.Expression.SetOuter(w)
	return w
}

func (w *WithPExpr) Codegen(ctx CodegenContext) string {
	funcCtx := ctx.WithP(w.PVar)
	return w.Target.Codegen(funcCtx)
}

func (w *WithPExpr) Reads() []string {
	return append([]string(nil), w.reads...)
}
