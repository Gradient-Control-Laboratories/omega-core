package convert

import (
	"fmt"

	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

func (d *Delegate) NewFnVar(name string, decl *program.FnType, enclosingScope *build.Scope) build.VarInfo {
	return NewFnVarInfo(name, decl, enclosingScope)
}

func (d *Delegate) NewParamVar(scope *build.Scope, name string, expr build.Expression, typ *program.Type) build.VarInfo {
	return VarInfoFor(scope, name, expr)
}

func (d *Delegate) BuildFunc(
	decl *program.FnType, fn *program.Function,
	pkgName string, args []build.Expression,
) (build.FnInfo, []build.VarInfo, error) {

	params := make([]build.VarInfo, len(args))
	paramTypes := make([]*program.Type, len(args))
	for i, arg := range args {
		paramName := fn.Parameters[i].Name
		varInfo := VarInfoFor(d.currBScope, paramName, arg)
		params[i] = varInfo
		paramTypes[i] = arg.Type()
	}

	return &FnInfo{
		name:       decl.Name,
		paramTypes: paramTypes,
		args:       args,
	}, params, nil
}

func (d *Delegate) BuildFuncReturn(
	fnInfo build.FnInfo,
	returnOps []*program.Operand, returnExprs []build.Expression,
) (build.FnInfo, error) {
	info := fnInfo.(*FnInfo)

	for i, expr := range returnExprs {
		_, isTreeNode := expr.(*TreeNode)
		if !isTreeNode {
			return nil, build.ErrLoc(d.GetScope(), returnOps[i].Start,
				fmt.Errorf("invalid return expression type: %T", expr))
		}
	}

	info.returnOps = returnOps
	info.returnExprs = returnExprs

	return info, nil
}

func (d *Delegate) BuildFuncCall(
	fnCallOp *program.Operand, fnCall *program.FnCall,
	fnInfo build.FnInfo, expressions []build.Expression,
) (build.Expression, error) {
	info := fnInfo.(*FnInfo)

	returnExprs := info.returnExprs
	if len(returnExprs) == 1 {
		if d.pixelSizeInput && info.name == shrinkToVanishName && len(info.args) == 4 &&
			len(fnCall.Arguments) == 4 && fnCall.Arguments[3].GetValue() == 1 {
			return d.partProxy(fnCallOp, returnExprs[0], info.args[1], info.args[2])
		}
		return returnExprs[0], nil
	}

	nodes := make([]*TreeNode, len(returnExprs))
	for i, expr := range returnExprs {
		nodes[i] = expr.(*TreeNode)
	}

	return NewMultiTreeNode(nodes...)
}

// -----------------------------------------------------------------------

// FnInfo caches the result of building a single function call
// and implements the [build.FnInfo] interface containing the result.
type FnInfo struct {
	name        string
	fn          *program.Function
	paramTypes  []*program.Type
	args        []build.Expression
	returnOps   []*program.Operand
	returnExprs []build.Expression
}

// Ensure that FnInfo implements the build.FnInfo interface.
var _ build.FnInfo = (*FnInfo)(nil)

func (f *FnInfo) Name() string {
	return f.name
}

func (f *FnInfo) Args() []build.Expression {
	return f.args
}

func (f *FnInfo) ParamTypes() []*program.Type {
	return f.paramTypes
}
