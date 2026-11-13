package convert

import (
	"fmt"

	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/internal/fidget"
	"gradientcontrol.com/omega/core/program"
)

// Build the given select operation (condition ? trueResult : falseResult) and return the expression associated with it.
func (d *Delegate) BuildSelect2(op *program.Operand, expr *program.Expression, operands []*program.Operand,
	conditionExpr build.Expression,
	trueResultExpr build.Expression, falseResultExpr build.Expression, resultType *program.Type,
) (build.Expression, error) {

	numType := toFidgetNumType(resultType)
	if numType == nil {
		return nil, build.ErrLoc(d.GetScope(), op.Start,
			fmt.Errorf("result type must be a number, got %v", resultType))
	}

	condNode := conditionExpr.(*TreeNode)
	trueNode := trueResultExpr.(*TreeNode)
	falseNode := falseResultExpr.(*TreeNode)

	condTrees, err := numType.ExpandTrees(condNode.trees)
	if err != nil {
		return nil, build.ErrLoc(d.GetScope(), op.Start, err)
	}
	trueTrees, err := numType.ExpandTrees(trueNode.trees)
	if err != nil {
		return nil, build.ErrLoc(d.GetScope(), op.Start, err)
	}
	falseTrees, err := numType.ExpandTrees(falseNode.trees)
	if err != nil {
		return nil, build.ErrLoc(d.GetScope(), op.Start, err)
	}

	resultTrees := make([]fidget.TreeIndex, len(condTrees))
	for i := 0; i < len(condTrees); i++ {

		// A condition is true when it is > 0, as in the WGSL and JS backends.
		// A comparison (or !, &&, || of comparisons) is already 0 or 1 and
		// weights the branches directly. Any other number is stepped at 0
		// first, the way BuildComparison builds ">": used raw, it would blend
		// the branches by its own value.
		condTree := condTrees[i]
		if !condNode.boolean {
			condTree = fidget.Max(d.zero, fidget.Compare(condTree, d.zero))
		}

		// result = and(cond, trueVal) + and(!cond, falseVal). fidget's and is a
		// branch, not a product: and(c, v) is v where c is nonzero and exactly 0
		// where it is 0, whatever v is there. Multiplied by 0, a branch not taken
		// that is infinite or NaN there (a division by zero, the root of a
		// negative) made the result NaN, as 0 * inf and 0 * NaN are NaN. Each
		// side is still kept whole, never subtracted, so trueVal keeps its
		// precision when falseVal is large, as builtin's nothing is. The two are
		// summed, not or-ed: or falls through where the branch taken is 0, on its
		// surface, and would give the condition's gradient there, not the branch's.
		resultTrees[i] = fidget.Add(fidget.And(condTree, trueTrees[i]),
			fidget.And(fidget.Not(condTree), falseTrees[i]))
	}

	return NewTreeNode(op, op.Start, resultTrees...)

}

// BuildSelect3 selects the positive, zero, or negative branch.
func (d *Delegate) BuildSelect3(op *program.Operand, expr *program.Expression, operands []*program.Operand,
	valueExpr build.Expression, positiveExpr build.Expression, zeroExpr build.Expression, negativeExpr build.Expression,
	resultType *program.Type,
) (build.Expression, error) {

	numType := toFidgetNumType(resultType)
	if numType == nil {
		return nil, build.ErrLoc(d.GetScope(), op.Start,
			fmt.Errorf("result type must be a number, got %v", resultType))
	}

	valueNode := valueExpr.(*TreeNode)
	valueTrees, err := numType.ExpandTrees(valueNode.trees)
	if err != nil {
		return nil, build.ErrLoc(d.GetScope(), op.Start, err)
	}
	positiveTrees, err := numType.ExpandTrees(positiveExpr.(*TreeNode).trees)
	if err != nil {
		return nil, build.ErrLoc(d.GetScope(), op.Start, err)
	}
	zeroTrees, err := numType.ExpandTrees(zeroExpr.(*TreeNode).trees)
	if err != nil {
		return nil, build.ErrLoc(d.GetScope(), op.Start, err)
	}
	negativeTrees, err := numType.ExpandTrees(negativeExpr.(*TreeNode).trees)
	if err != nil {
		return nil, build.ErrLoc(d.GetScope(), op.Start, err)
	}

	resultTrees := make([]fidget.TreeIndex, len(valueTrees))
	for i := 0; i < len(valueTrees); i++ {
		// Each branch masked by its own condition with and, as in BuildSelect2:
		// a branch not taken that is infinite or NaN there drops out.
		positive := fidget.Max(d.zero, fidget.Compare(valueTrees[i], d.zero))
		negative := fidget.Max(d.zero, fidget.Compare(d.zero, valueTrees[i]))
		zero := fidget.Not(fidget.Add(positive, negative))
		pos := fidget.And(positive, positiveTrees[i])
		neg := fidget.And(negative, negativeTrees[i])
		resultTrees[i] = fidget.Add(fidget.And(zero, zeroTrees[i]), fidget.Add(pos, neg))
	}

	return NewTreeNode(op, op.Start, resultTrees...)
}
