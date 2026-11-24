package wgsl

import (
	"fmt"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/build"
	"github.com/Gradient-Control-Laboratories/omega-core/program"
)

// Square implements the [build.Expression] interface.
type Square struct {
	Expression
	Target Expression
}

func NewSquare(op *program.Operand, target build.Expression, resultType *program.Type) (Expression, error) {
	wgslExpr := target.(Expression).ToNum()
	rType := WGSLType(resultType, false)
	if rType.Name == "" {
		rType = TypeF32
	}

	s := &Square{
		Target: wgslExpr,
	}
	s.Expression = NewExpr(op, rType, build.PrecedenceHighest)
	s.Expression.SetOuter(s)
	return s, nil
}

func (s *Square) Codegen(ctx CodegenContext) string {
	return fmt.Sprintf("pow(%s, 2.0)", s.Target.Codegen(ctx))
}

func (s *Square) Reads() []string {
	return s.Target.Reads()
}
