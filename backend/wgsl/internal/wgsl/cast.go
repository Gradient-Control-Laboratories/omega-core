package wgsl

import (
	"fmt"

	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// CastExpr represents a typed WGSL constructor / cast expression (e.g. vec3<f32>(x)).
type CastExpr struct {
	// Read-only immutable fields set at construction
	Expression
	TargetType TypeInfo
	Sub        Expression
}

func NewCastExpr(op *program.Operand, targetType TypeInfo, sub Expression) *CastExpr {
	c := &CastExpr{
		TargetType: targetType,
		Sub:        sub,
	}
	c.Expression = NewExpr(op, targetType, build.PrecedenceHighest)
	c.Expression.SetOuter(c)
	return c
}

func (c *CastExpr) Codegen(ctx CodegenContext) string {
	return fmt.Sprintf("%s(%s)", c.TargetType.Name, c.Sub.Codegen(ctx))
}

func (c *CastExpr) Reads() []string {
	return c.Sub.Reads()
}
