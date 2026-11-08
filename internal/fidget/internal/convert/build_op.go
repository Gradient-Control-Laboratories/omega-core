package convert

import (
	"fmt"

	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/internal/fidget/internal/fidget"
	"gradientcontrol.com/omega/core/program"
)

func buildUnaryOperation(
	scope *build.Scope,
	op *program.Operand,
	expression build.Expression,
	resultType *program.Type,
	// The function that creates the underlying fidget expression.
	calc func(arg1 fidget.TreeIndex) (fidget.TreeIndex, error),
) (build.Expression, error) {

	numType := toFidgetNumType(resultType)
	if numType == nil {
		return nil, build.ErrLoc(scope, op.Start,
			fmt.Errorf("expected numeric type, got %s", resultType))
	}

	node := expression.(*TreeNode)

	trees, err := numType.ExpandTrees(node.trees)
	if err != nil {
		return nil, build.ErrLoc(scope, op.Start, err)
	}

	resultTrees := make([]fidget.TreeIndex, len(trees))
	for i := 0; i < len(trees); i++ {
		resultTrees[i], err = calc(trees[i])
		if err != nil {
			return nil, build.ErrLoc(scope, op.Start, err)
		}
	}

	return NewTreeNode(op, op.Start, resultTrees...)
}

func buildBinaryOperation(
	scope *build.Scope,
	op *program.Operand,
	operands []*program.Operand,
	expressions []build.Expression,
	resultType *program.Type,
	// The function that creates the underlying fidget expression.
	calc func(arg1, arg2 fidget.TreeIndex) (fidget.TreeIndex, error),
) (build.Expression, error) {
	if len(expressions) != 2 {
		return nil, build.ErrLoc(scope, op.Start,
			fmt.Errorf("expected exactly two expressions, got %d", len(expressions)))
	}

	numType := toFidgetNumType(resultType)
	if numType == nil {
		return nil, build.ErrLoc(scope, operands[0].Start,
			fmt.Errorf("expected numeric type, got %s", resultType))
	}

	node1 := expressions[0].(*TreeNode)
	node2 := expressions[1].(*TreeNode)

	trees1, err := numType.ExpandTrees(node1.trees)
	if err != nil {
		return nil, build.ErrLoc(scope, operands[0].Start, err)
	}
	trees2, err := numType.ExpandTrees(node2.trees)
	if err != nil {
		return nil, build.ErrLoc(scope, operands[1].Start, err)
	}

	resultTrees := make([]fidget.TreeIndex, len(trees1))
	for i := 0; i < len(trees1); i++ {
		resultTrees[i], err = calc(trees1[i], trees2[i])
		if err != nil {
			return nil, build.ErrLoc(scope, operands[0].Start, err)
		}
	}
	return NewTreeNode(op, op.Start, resultTrees...)
}

func buildNumericOperation(
	scope *build.Scope,
	op *program.Operand,
	operands []*program.Operand,
	expressions []build.Expression,
	resultType *program.Type,
	// The function that creates the underlying fidget expression.
	calc func(arg1 fidget.TreeIndex, arg2 fidget.TreeIndex) (fidget.TreeIndex, error),
) (build.Expression, error) {
	if len(expressions) == 0 {
		return nil, build.ErrLoc(scope, op.Start,
			fmt.Errorf("expected at least one expression, got %d", len(expressions)))
	}

	numType := toFidgetNumType(resultType)
	if numType == nil {
		return nil, build.ErrLoc(scope, operands[0].Start,
			fmt.Errorf("expected numeric type, got %s", resultType))
	}

	node := expressions[0].(*TreeNode)
	trees, err := numType.ExpandTrees(node.trees)
	if err != nil {
		return nil, build.ErrLoc(scope, operands[0].Start, err)
	}

	for i := 1; i < len(expressions); i++ {
		otherNode := expressions[i].(*TreeNode)
		otherTrees, err := numType.ExpandTrees(otherNode.trees)
		if err != nil {
			return nil, build.ErrLoc(scope, operands[i].Start, err)
		}
		resultTrees := make([]fidget.TreeIndex, len(trees))
		for i := 0; i < len(trees); i++ {
			resultTrees[i], err = calc(trees[i], otherTrees[i])
			if err != nil {
				return nil, build.ErrLoc(scope, operands[i].Start, err)
			}
		}
		trees = resultTrees
	}

	return NewTreeNode(op, op.Start, trees...)
}
