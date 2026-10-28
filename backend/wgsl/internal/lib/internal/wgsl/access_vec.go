package wgsl

import (
	"fmt"

	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// VecAccess implements the [build.Expression] interface.
type VecAccess struct {
	Expression
	Target Expression
	Member string
}

func NewVecMemberAccess(member string, op *program.Operand, target Expression) (Expression, error) {
	var resultType TypeInfo
	switch len(member) {
	case 1:
		resultType = TypeF32
	case 2:
		resultType = TypeVec2
	case 3:
		resultType = TypeVec3
	case 4:
		resultType = TypeVec4
	default:
		// Parser ensures that member is 1 to 4 characters long
		return nil, fmt.Errorf("access operator cannot be more than 4 characters long")
	}

	a := &VecAccess{
		Target: target,
		Member: member,
	}
	a.Expression = NewExpr(op, resultType, build.PrecedenceHighest)
	a.Expression.SetOuter(a)
	return a, nil
}

func NewVecIndexedAccess(op *program.Operand, target Expression, index int) (Expression, error) {
	var member string
	switch index {
	case 0:
		member = "x"
	case 1:
		member = "y"
	case 2:
		member = "z"
	case 3:
		member = "w"
	default:
		// Parser ensures that index is 0 to 3
		return nil, fmt.Errorf("index must be between 0 and 3")
	}
	return NewVecMemberAccess(member, op, target)
}

func (a *VecAccess) Codegen(ctx CodegenContext) string {
	return fmt.Sprintf("%s.%s", a.Target.ParenIf(build.PrecedenceHighest).Codegen(ctx), a.Member)
}

func (a *VecAccess) Reads() []string {
	return a.Target.Reads()
}
