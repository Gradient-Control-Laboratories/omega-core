package wgsl

import (
	"fmt"

	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// Access implements the [build.Expression] interface.
type ArrayAccess struct {
	Expression
	Target  Expression
	Indices []int
}

func NewArrayMemberAccess(member string, op *program.Operand, target Expression) (Expression, error) {
	indices := make([]int, len(member))
	for i, c := range member {
		switch c {
		case 'x':
			indices[i] = 0
		case 'y':
			indices[i] = 1
		case 'z':
			indices[i] = 2
		case 'w':
			indices[i] = 3
		default:
			return nil, fmt.Errorf("invalid character in member: %c", c)
		}
	}

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

	a := &ArrayAccess{
		Target:  target,
		Indices: indices,
	}
	a.Expression = NewExpr(op, resultType, build.PrecedenceHighest)
	a.Expression.SetOuter(a)
	return a, nil
}

func NewArrayIndexedAccess(op *program.Operand, target Expression, index int) (Expression, error) {
	a := &ArrayAccess{
		Target:  target,
		Indices: []int{index},
	}
	a.Expression = NewExpr(op, TypeF32, build.PrecedenceHighest)
	a.Expression.SetOuter(a)
	return a, nil
}

// NewListIndexedAccess builds target[index] where target is a list of
// vectors, so the result keeps the element's vector type.
func NewListIndexedAccess(op *program.Operand, target Expression, index int, elemType TypeInfo) (Expression, error) {
	a := &ArrayAccess{
		Target:  target,
		Indices: []int{index},
	}
	a.Expression = NewExpr(op, elemType, build.PrecedenceHighest)
	a.Expression.SetOuter(a)
	return a, nil
}

func (a *ArrayAccess) Codegen(ctx CodegenContext) string {
	target := a.Target.ParenIf(build.PrecedenceHighest).Codegen(ctx)

	if len(a.Indices) == 1 {
		return fmt.Sprintf("%s[%d]", target, a.Indices[0])
	}

	// If the target is not a variable reference, assign it to a variable
	// so that it is only evaluated once.
	if _, isVarRef := a.Target.(*VarRef); !isVarRef {
		target = ctx.AssignUnique("gsl_array", target, a.Target.Reads()...)
	}

	result := fmt.Sprintf("%s[%d]", target, a.Indices[0])
	for i := 1; i < len(a.Indices); i++ {
		result = fmt.Sprintf("%s, %s[%d]", result, target, a.Indices[i])
	}
	return fmt.Sprintf("vec%d(%s)", len(a.Indices), result)
}

func (a *ArrayAccess) Reads() []string {
	return a.Target.Reads()
}
