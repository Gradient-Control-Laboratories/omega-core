package convert

import (
	"fmt"

	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/internal/fidget/internal/fidget"
	"gradientcontrol.com/omega/core/program"
)

func (d *Delegate) BuildMemberAccess(
	op *program.Operand, expr *program.Expression,
	target build.Expression, member string,
) (build.Expression, error) {
	targetTrees := target.(*TreeNode).trees
	resultTrees := make([]fidget.TreeIndex, len(member))
	for i, ch := range member {
		switch ch {
		case 'x':
			resultTrees[i] = targetTrees[0]
		case 'y':
			resultTrees[i] = targetTrees[1]
		case 'z':
			resultTrees[i] = targetTrees[2]
		case 'w':
			resultTrees[i] = targetTrees[3]
		default:
			return nil, build.ErrLoc(d.GetScope(), op.Start, fmt.Errorf("invalid member: %s", member))
		}
	}
	return NewTreeNode(op, op.Start, resultTrees...)
}

func (d *Delegate) BuildIndexAccess(
	op *program.Operand, expr *program.Expression,
	targetOp *program.Operand, indexOp *program.Operand,
	target build.Expression, index int,
) (build.Expression, error) {
	targetTrees := target.(*TreeNode).trees

	// A list of vectors stores its trees flattened in element-major order,
	// so indexing slices out the element's trees.
	if target.Type().IsMatrix() {
		elemLen := int(target.Type().ElemType.Size)
		start := index * elemLen
		return NewTreeNode(op, op.Start, targetTrees[start:start+elemLen]...)
	}

	return NewTreeNode(op, op.Start, targetTrees[index])
}
