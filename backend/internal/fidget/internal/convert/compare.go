package convert

import (
	"fmt"

	"gradientcontrol.com/omega/core/backend/internal/fidget/internal/fidget"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

func (d *Delegate) BuildCompare(
	op *program.Operand, expr *program.Expression, operands []*program.Operand,
	expr1, expr2 build.Expression, resultType *program.Type,
) (build.Expression, error) {

	numType := toFidgetNumType(resultType)
	if numType == nil {
		return nil, build.ErrLoc(d.GetScope(), operands[0].Start, fmt.Errorf("expected numeric type, got %s", resultType))
	}

	node1 := expr1.(*TreeNode)
	node2 := expr2.(*TreeNode)

	trees1, err := numType.ExpandTrees(node1.trees)
	if err != nil {
		return nil, build.ErrLoc(d.GetScope(), operands[0].Start, err)
	}
	trees2, err := numType.ExpandTrees(node2.trees)
	if err != nil {
		return nil, build.ErrLoc(d.GetScope(), operands[1].Start, err)
	}

	resultTrees := make([]fidget.TreeIndex, len(trees1))
	for i := 0; i < len(trees1); i++ {
		resultTrees[i], err = func(arg1, arg2 fidget.TreeIndex) (fidget.TreeIndex, error) {
			// fidget.Compare evaluates spatial intervals continuously, outputting [-1, 1].
			// When arg1 < arg2 it evaluates to -1.
			// When arg1 > arg2 it evaluates to 1.
			// When arg1 == arg2 it evaluates exactly to 0.
			// We rely on this continuous -1 to 1 mapping when building comparison conditions
			// for the MDC mesher because sharp Boolean thresholds cause interval expansion breakdown.
			return fidget.Compare(arg1, arg2), nil
		}(trees1[i], trees2[i])
		if err != nil {
			return nil, build.ErrLoc(d.GetScope(), operands[0].Start, err)
		}
	}

	return NewTreeNode(op, op.Start, resultTrees...)
}
