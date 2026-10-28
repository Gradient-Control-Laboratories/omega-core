package build

import (
	"fmt"
	"gradientcontrol.com/omega/core/program"
	"strings"
)

func (b *Builder) visitAccess(op *program.Operand, expr *program.Expression) (Expression, error) {
	if len(expr.Operands) != 2 {
		return nil, ErrLoc(b.delegate.GetScope(), op.Start, fmt.Errorf("access operator expects 2 operands, got %d", len(expr.Operands)))
	}
	targetOp := expr.Operands[0]
	accessorOp := expr.Operands[1]

	ref, ok := accessorOp.T.(*program.Operand_Reference)
	if !ok {
		return nil, ErrLoc(b.delegate.GetScope(), accessorOp.Start, fmt.Errorf("access operator expects second operand to be a reference"))
	}
	member := ref.Reference.Name

	target, err := program.VisitOperand(targetOp, b)
	if err != nil {
		return nil, err
	}
	targetType := target.Type()
	if targetType == nil && targetOp != nil {
		targetType = targetOp.Type
	}

	if targetType.IsMatrix() {
		err := fmt.Errorf("member access is not supported on %s; use indexing instead", targetType.TypeString())
		return nil, ErrLoc(b.delegate.GetScope(), op.Start, err)
	}

	if !targetType.IsList() {
		err := fmt.Errorf("access operator requires first operand to be a collection, got %s", targetType.TypeString())
		return nil, ErrLoc(b.delegate.GetScope(), op.Start, err)
	}

	switch targetType.Size {
	case 4:
		// valid
	case 3:

		// cannot use "w" in member
		if strings.Contains(member, "w") {
			err := fmt.Errorf("cannot use \"w\" to access a collection of 3 elements")
			return nil, ErrLoc(b.delegate.GetScope(), op.Start, err)
		}
	case 2:
		// cannot use "z" or "w" in member
		if strings.Contains(member, "z") || strings.Contains(member, "w") {
			err := fmt.Errorf("cannot use \"z\" or \"w\" to access a collection of 2 elements")
			return nil, ErrLoc(b.delegate.GetScope(), op.Start, err)
		}
	case 1:
		// cannot use "y", "z", or "w" in member
		if strings.Contains(member, "y") || strings.Contains(member, "z") || strings.Contains(member, "w") {
			err := fmt.Errorf("cannot use \"y\", \"z\", or \"w\" to access a collection of 1 element")
			return nil, ErrLoc(b.delegate.GetScope(), op.Start, err)
		}
	default:
		// collections with length > 4 are valid to access using x, y, z, w
	}

	if col, ok := target.(CollectExpression); ok {
		elems := col.Elements()
		if len(elems) > 0 && len(member) == 1 {
			var idx int
			switch member {
			case "x":
				idx = 0
			case "y":
				idx = 1
			case "z":
				idx = 2
			case "w":
				idx = 3
			default:
				idx = -1
			}
			if idx >= 0 && idx < len(elems) {
				return elems[idx], nil
			}
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
				return b.delegate.BuildMemberAccess(op, expr, target, member)
			}
			if idx >= len(vals) {
				return b.delegate.BuildMemberAccess(op, expr, target, member)
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

	return b.delegate.BuildMemberAccess(op, expr, target, member)
}
