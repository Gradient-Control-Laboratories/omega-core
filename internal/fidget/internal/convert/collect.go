package convert

import (
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/internal/fidget/internal/fidget"
	"gradientcontrol.com/omega/core/program"
)

func (d *Delegate) BuildCollect(
	collectOp *program.Operand,
	expr *program.Expression,
	operands []*program.Operand,
	expressions []build.Expression,
) (build.Expression, error) {

	// Vector elements gather into a list of vectors with the trees stored
	// flattened in element-major order. The common build ensures the elements
	// are uniform vectors of the same length.
	if expressions[0].Type().IsList() {
		elemLen := int(expressions[0].Type().Size)
		trees := make([]fidget.TreeIndex, 0, len(expressions)*elemLen)
		for i := 0; i < len(expressions); i++ {
			node := expressions[i].(*TreeNode)
			nodeTrees, err := toFidgetNumType(node.Type()).ExpandTrees(node.trees)
			if err != nil {
				return nil, err
			}
			trees = append(trees, nodeTrees...)
		}
		node, err := NewListVecTreeNode(collectOp, collectOp.Start, elemLen, trees...)
		if err != nil {
			return nil, err
		}
		node.elements = expressions
		return node, nil
	}

	trees := make([]fidget.TreeIndex, len(expressions))
	for i := 0; i < len(expressions); i++ {
		node := expressions[i].(*TreeNode)
		trees[i] = node.trees[0]
	}

	t, err := NewTreeNode(collectOp, collectOp.Start, trees...)
	if err != nil {
		return nil, err
	}
	t.elements = expressions

	return t, nil
}

func (d *Delegate) BuildConstCollect(
	collectOp *program.Operand,
	expr *program.Expression,
	operands []*program.Operand,
	expressions []build.Expression,
	constValues []float64,
) (build.ConstExpression, error) {
	t, err := d.BuildCollect(collectOp, expr, operands, expressions)
	if err != nil {
		return nil, err
	}
	node := t.(*TreeNode)
	node.constValues = constValues
	return node, nil
}
