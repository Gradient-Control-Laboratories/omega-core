package build

import (
	"fmt"
	"gradientcontrol.com/omega/core/program"
	"strings"
)

// VisitPropRef implements [program.OperandVisitor].
func (b *Builder) VisitPropRef(op *program.Operand, propRef *program.PropertyRef) (Expression, error) {
	res, err := b.visitPropRefImpl(op, propRef)
	if err != nil {
		return nil, err
	}
	if len(op.ConstValues) > 0 {
		return b.buildConstExpressionWithType(op, op.ConstValues, op.Type)
	}
	return res, nil
}

func (b *Builder) visitPropRefImpl(op *program.Operand, propRef *program.PropertyRef) (Expression, error) {
	refOp := &program.Operand{T: &program.Operand_Reference{Reference: propRef.Reference}}
	target, err := program.VisitOperand(refOp, b)
	if err != nil {
		return nil, err
	}
	targetType := target.Type()
	if targetType == nil && refOp.Type != nil {
		targetType = refOp.Type
	}

	if targetType.IsMatrix() {
		err := fmt.Errorf("member access is not supported on %s; use indexing instead", targetType.TypeString())
		return nil, ErrLoc(b.delegate.GetScope(), op.Start, err)
	}

	if !targetType.IsList() {
		err := fmt.Errorf("access operator requires first operand to be a collection, got %s", targetType.TypeString())
		return nil, ErrLoc(b.delegate.GetScope(), op.Start, err)
	}

	member := propRef.Property

	switch targetType.Size {
	case 4:
		// valid
	case 3:

		if strings.Contains(member, "w") {
			err := fmt.Errorf("cannot use \"w\" to access a collection of 3 elements")
			return nil, ErrLoc(b.delegate.GetScope(), op.Start, err)
		}
	case 2:
		if strings.Contains(member, "z") || strings.Contains(member, "w") {
			err := fmt.Errorf("cannot use \"z\" or \"w\" to access a collection of 2 elements")
			return nil, ErrLoc(b.delegate.GetScope(), op.Start, err)
		}
	case 1:
		if strings.Contains(member, "y") || strings.Contains(member, "z") || strings.Contains(member, "w") {
			err := fmt.Errorf("cannot use \"y\", \"z\", or \"w\" to access a collection of 1 element")
			return nil, ErrLoc(b.delegate.GetScope(), op.Start, err)
		}
	}

	if ce, ok := target.(ConstExpression); ok && ce.ConstValues() != nil {
		vals := ce.ConstValues()
		result := make([]float64, 0, len(member))
		for _, ch := range member {
			var idx int
			switch ch {
			case 'x':
				idx = 0
			case 'y':
				idx = 1
			case 'z':
				idx = 2
			case 'w':
				idx = 3
			default:
				return b.delegate.BuildMemberAccess(op, nil, target, member)
			}
			if idx >= len(vals) {
				return b.delegate.BuildMemberAccess(op, nil, target, member)
			}
			result = append(result, vals[idx])
		}

		if len(result) == 1 {
			return b.delegate.BuildValue(op, result[0])
		}

		childExprs := make([]Expression, len(result))
		for i, val := range result {
			child, err := b.delegate.BuildValue(op, val)
			if err != nil {
				return nil, err
			}
			childExprs[i] = child
		}
		return b.delegate.BuildConstCollect(op, nil, nil, childExprs, result)
	}

	return b.delegate.BuildMemberAccess(op, nil, target, member)
}
