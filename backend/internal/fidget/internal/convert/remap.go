package convert

import (
	"fmt"

	"gradientcontrol.com/omega/core/backend/internal/fidget/internal/fidget"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

func (d *Delegate) Remap(
	remapOp *program.Operand,
	targetOp *program.Operand,
	pExpr build.Expression,
	targetExpr build.Expression,
) (build.Expression, error) {

	pTrees, err := Vec3Type.ExpandTrees(pExpr.(*TreeNode).trees)
	if err != nil {
		return nil, build.ErrLoc(d.GetScope(), pExpr.Start(), err)
	}
	x := pTrees[0]
	y := pTrees[1]
	z := pTrees[2]

	if multiNode, isMultiNode := targetExpr.(*MultiTreeNode); isMultiNode {
		result := make([]*TreeNode, len(multiNode.Nodes()))
		for i, node := range multiNode.Nodes() {
			result[i], err = d.remapNode(remapOp, node.Operand(), toFidgetNumType(node.Type()), node, x, y, z)
			if err != nil {
				return nil, err
			}
		}
		return NewMultiTreeNode(result...)
	}
	if targetNode, ok := targetExpr.(*TreeNode); ok {
		if numType := toFidgetNumType(targetNode.Type()); numType != nil {
			return d.remapNode(remapOp, targetOp, numType, targetNode, x, y, z)
		}
	}

	return nil, build.ErrLoc(d.GetScope(), targetOp.Start,
		fmt.Errorf("expected numeric type, got %s", targetExpr.Type()))
}

func (d *Delegate) remapNode(
	remapOp *program.Operand, targetOp *program.Operand,
	numType FidgetNumType, targetExpr *TreeNode,
	x fidget.TreeIndex, y fidget.TreeIndex, z fidget.TreeIndex,
) (*TreeNode, error) {
	targetTrees, err := numType.ExpandTrees(targetExpr.trees)
	if err != nil {
		return nil, build.ErrLoc(d.GetScope(), targetOp.Start, err)
	}
	if err := d.preserveRemappedSpinodoids(targetTrees); err != nil {
		return nil, build.ErrLoc(d.GetScope(), targetOp.Start, err)
	}

	resultTrees := make([]fidget.TreeIndex, len(targetTrees))
	for i := 0; i < len(targetTrees); i++ {
		resultTrees[i] = fidget.Remap(targetTrees[i], x, y, z)
	}
	return NewTreeNode(remapOp, remapOp.Start, resultTrees...)
}
