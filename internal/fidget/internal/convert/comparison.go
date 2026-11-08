package convert

import (
	"fmt"

	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/internal/fidget/internal/fidget"
	"gradientcontrol.com/omega/core/program"
)

func (d *Delegate) BuildComparison(
	op *program.Operand, expr *program.Expression, symbol string, operands []*program.Operand,
	expressions []build.Expression, resultType *program.Type, precedence build.Precedence,
) (build.Expression, error) {
	return markBoolean(d.buildComparison(op, symbol, operands, expressions, resultType))
}

// markBoolean flags a node as 0/1 valued (TreeNode.boolean).
func markBoolean(e build.Expression, err error) (build.Expression, error) {
	if node, ok := e.(*TreeNode); ok && err == nil {
		node.boolean = true
	}
	return e, err
}

func (d *Delegate) buildComparison(
	op *program.Operand, symbol string, operands []*program.Operand,
	expressions []build.Expression, resultType *program.Type,
) (build.Expression, error) {
	// fidget.Compare evaluates spatial intervals continuously, outputting [-1, 1].
	// When arg1 < arg2 it evaluates to -1.
	// When arg1 > arg2 it evaluates to 1.
	// When arg1 == arg2 it evaluates exactly to 0.
	// Consequently, we wrap these comparative functions in fidget.Max(d.zero, ...) bounds!
	// This clamps any negatively-evaluated spatial region (false) specifically to 0.0,
	// while letting positively-evaluated regions (true) transition cleanly toward 1.0.
	// This uniquely restricts boolean condition trees to the tightly stable [0, 1] domain,
	// entirely eliminating the mathematical boundary tearing which causes the MDC mesher
	// to encounter unconstrained EDGE_SEARCH_SIZE panics during ternary interpolations.
	switch symbol {
	case ">":
		return buildBinaryOperation(d.GetScope(), op, operands, expressions, resultType,
			func(arg1, arg2 fidget.TreeIndex) (fidget.TreeIndex, error) {
				return fidget.Max(d.zero, fidget.Compare(arg1, arg2)), nil
			})
	case ">=":
		return buildBinaryOperation(d.GetScope(), op, operands, expressions, resultType,
			func(arg1, arg2 fidget.TreeIndex) (fidget.TreeIndex, error) {
				// The complement of "<", so equality is true, as in WGSL.
				return fidget.Sub(d.one, fidget.Max(d.zero, fidget.Compare(arg2, arg1))), nil
			})
	case "==":
		return buildBinaryOperation(d.GetScope(), op, operands, expressions, resultType,
			func(arg1, arg2 fidget.TreeIndex) (fidget.TreeIndex, error) {
				// 1 - abs(comp) yields 1 when arg1 == arg2, dropping off to 0 linearly.
				return fidget.Max(d.zero, fidget.Sub(d.one, fidget.Abs(fidget.Compare(arg1, arg2)))), nil
			})
	case "!=":
		return buildBinaryOperation(d.GetScope(), op, operands, expressions, resultType,
			func(arg1, arg2 fidget.TreeIndex) (fidget.TreeIndex, error) {
				// abs(comp) yields 0 when arg1 == arg2, evaluating to 1 linearly elsewhere.
				return fidget.Max(d.zero, fidget.Abs(fidget.Compare(arg1, arg2))), nil
			})
	case "<=":
		return buildBinaryOperation(d.GetScope(), op, operands, expressions, resultType,
			func(arg1, arg2 fidget.TreeIndex) (fidget.TreeIndex, error) {
				// The complement of ">", so equality is true, as in WGSL.
				return fidget.Sub(d.one, fidget.Max(d.zero, fidget.Compare(arg1, arg2))), nil
			})
	case "<":
		return buildBinaryOperation(d.GetScope(), op, operands, expressions, resultType,
			func(arg1, arg2 fidget.TreeIndex) (fidget.TreeIndex, error) {
				return fidget.Max(d.zero, fidget.Compare(arg2, arg1)), nil
			})
	default:
		return nil, fmt.Errorf("unhandled comparison: %s", symbol)
	}
}
