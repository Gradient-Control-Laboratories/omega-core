package convert

import (
	"fmt"

	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

func (d *Delegate) BuildSingleAssignment(
	decl *program.Assignment, op *program.Operand, name string, expr build.Expression,
) (build.VarInfo, error) {
	// A requested parameter builds as an input var instead of its literal —
	// the substrate for interval evaluation over the parameter's range.
	if pv := d.paramAssignment(name, expr); pv != nil {
		return NewTreeVarInfo(d.currBScope, name, pv.(*TreeNode)), nil
	}

	node, isNode := expr.(*TreeNode)
	if isNode {
		return NewTreeVarInfo(d.currBScope, name, node), nil
	}

	fnVarExpr, isFnVarExpr := expr.(*FnVarExpr)
	if isFnVarExpr {
		return NewFnVarInfo(name, fnVarExpr.fnVarInfo.fnDecl, fnVarExpr.fnVarInfo.enclosingScope), nil
	}

	return nil, build.ErrLoc(d.GetScope(), op.Start,
		fmt.Errorf("expected tree node for %s, got %T", name, expr))
}

func (d *Delegate) BuildMultiAssignment(
	decl *program.Assignment, op *program.Operand,
	fnCall *program.FnCall, names []string, expr build.Expression,
) ([]build.VarInfo, error) {

	multiNode, ok := expr.(*MultiTreeNode)
	if !ok {
		return nil, build.ErrLoc(d.GetScope(), op.Start,
			fmt.Errorf("expected multi tree node for %s, got %T", fnCall.Reference.Name, expr))
	}

	nodes := multiNode.Nodes()
	if len(nodes) != len(names) {
		return nil, build.ErrLoc(d.GetScope(), op.Start,
			fmt.Errorf("expected %d children for %s, got %d", len(names), fnCall.Reference.Name, len(nodes)))
	}

	varInfos := make([]build.VarInfo, len(names))
	for i, name := range names {
		varInfos[i] = NewTreeVarInfo(d.currBScope, name, nodes[i])
	}
	return varInfos, nil
}
