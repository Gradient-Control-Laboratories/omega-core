package codegen

import (
	"gradientcontrol.com/omega/core/backend/wgsl/internal/ir"
	"gradientcontrol.com/omega/core/backend/wgsl/internal/wgsl"
)

type BuildContext struct {
	delegate *Delegate
	PVar     *wgsl.VarInfo
}

// Ensure that BuildContext implements the wgsl.CodegenContext interface.
var _ wgsl.CodegenContext = (*BuildContext)(nil)

func NewBuildContext(delegate *Delegate, pVar *wgsl.VarInfo) BuildContext {
	return BuildContext{delegate: delegate, PVar: pVar}
}

func (c BuildContext) P() *wgsl.VarInfo {
	return c.PVar
}

func (c BuildContext) WithP(p *wgsl.VarInfo) wgsl.CodegenContext {
	return BuildContext{delegate: c.delegate, PVar: p}
}

func (c BuildContext) EmitVar(ref *wgsl.VarInfo) {
	if c.delegate.ShouldGenerate(ref) && c.delegate.markGenerated(ref.WGSLName()) {
		c.delegate.emitVar(ref, c)
	}
}

func (c BuildContext) Root() wgsl.CodegenContext {
	return c.delegate.rootCtx
}

func (c BuildContext) AssignUnique(prefix string, target string, reads ...string) string {
	name := c.delegate.uniqueTopLevelName(prefix)
	c.delegate.currentScope().AddStmt(ir.NewLetStmt(name, "", target, reads))
	return name
}
