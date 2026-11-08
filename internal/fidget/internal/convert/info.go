package convert

import (
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

func VarInfoFor(scope *build.Scope, name string, expr build.Expression) build.VarInfo {
	if node, ok := expr.(*TreeNode); ok {
		return NewTreeVarInfo(scope, name, node)
	}
	if fnExpr, ok := expr.(*FnVarExpr); ok {
		return &FnVarInfo{
			name:           name,
			fnDecl:         fnExpr.fnVarInfo.fnDecl,
			enclosingScope: fnExpr.fnVarInfo.enclosingScope,
			fnExpr:         fnExpr,
		}
	}
	panic("expression is not a tree node or function variable expression")
}

// -----------------------------------------------------------------------

type TreeVarInfo struct {
	bScope *build.Scope
	name   string
	node   *TreeNode
}

// Ensure that TreeVarInfo implements the build.VarInfo interface.
var _ build.VarInfo = (*TreeVarInfo)(nil)

func NewTreeVarInfo(bScope *build.Scope, name string, node *TreeNode) *TreeVarInfo {
	return &TreeVarInfo{bScope: bScope, name: name, node: node}
}

func (v *TreeVarInfo) Name() string {
	return v.name
}

func (v *TreeVarInfo) Scope() *build.Scope {
	return v.bScope
}

func (v *TreeVarInfo) Expr() build.Expression {
	return v.node
}

func (v *TreeVarInfo) Type() *program.Type {
	return v.node.Type()
}

func (v *TreeVarInfo) Node() *TreeNode {
	return v.node
}

// -----------------------------------------------------------------------

type FnVarInfo struct {
	name           string
	fnDecl         *program.FnType
	enclosingScope *build.Scope
	fnExpr         *FnVarExpr
}

// Ensure that FnVarInfo implements the build.VarInfo interface.
var _ build.VarInfo = (*FnVarInfo)(nil)

func NewFnVarInfo(name string, decl *program.FnType, enclosingScope *build.Scope) *FnVarInfo {
	expr := &FnVarExpr{name: name}
	info := &FnVarInfo{name: name, fnDecl: decl, enclosingScope: enclosingScope, fnExpr: expr}
	expr.fnVarInfo = info
	return info
}

func (v *FnVarInfo) Name() string {
	return v.name
}

func (v *FnVarInfo) Scope() *build.Scope {
	return v.enclosingScope
}

func (v *FnVarInfo) Expr() build.Expression {
	return v.fnExpr
}

func (v *FnVarInfo) Type() *program.Type {
	return program.NewFunctionType("", v.name)
}

// -----------------------------------------------------------------------

type FnVarExpr struct {
	name      string
	op        *program.Operand
	fnVarInfo *FnVarInfo
}

// Ensure that FnVarExpr implements the build.Expression interface.
var _ build.Expression = (*FnVarExpr)(nil)

func (e *FnVarExpr) Operand() *program.Operand {
	return e.op
}

func (e *FnVarExpr) Start() *program.Location {
	return e.op.Start
}

func (e *FnVarExpr) Precedence() build.Precedence {
	return build.PrecedenceHighest
}

func (e *FnVarExpr) VarInfo() *FnVarInfo {
	return e.fnVarInfo
}

func (e *FnVarExpr) Type() *program.Type {
	return program.NewFunctionType("", e.name)
}
