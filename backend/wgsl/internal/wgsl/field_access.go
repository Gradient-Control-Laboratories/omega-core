package wgsl

import (
	"fmt"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/build"
	"github.com/Gradient-Control-Laboratories/omega-core/program"
)

// FieldAccess models member access on a struct expression (e.g. gsl_result.v_1).
type FieldAccess struct {
	Expression
	Target Expression
	Field  string
}

func NewFieldAccess(op *program.Operand, target Expression, field string, fieldType TypeInfo) *FieldAccess {
	f := &FieldAccess{
		Target: target,
		Field:  field,
	}
	f.Expression = NewExpr(op, fieldType, build.PrecedenceHighest)
	f.Expression.SetOuter(f)
	return f
}

func (f *FieldAccess) Codegen(ctx CodegenContext) string {
	return fmt.Sprintf("%s.%s", f.Target.Codegen(ctx), f.Field)
}

func (f *FieldAccess) Reads() []string {
	return f.Target.Reads()
}
