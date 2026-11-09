package scope

import (
	"gradientcontrol.com/omega/core/program"
)

// FnGroupScope represents the scope of a function overload group.
type FnGroupScope struct {
	parent    *FnTypeScope
	FnGroup   *program.FnGroup
	Functions []*FnScope
}

var _ Scope = (*FnGroupScope)(nil)

func (f *FnGroupScope) ContainingScope() Scope {
	if f.parent != nil {
		return f.parent.ContainingScope()
	}
	return nil
}

func (f *FnGroupScope) GetAgentScope() AgentScope {
	return nil
}

func (f *FnGroupScope) SetAgentScope(agentScope AgentScope) {}

func (f *FnGroupScope) PkgFileScope() *PkgFileScope {
	if f.parent != nil {
		return f.parent.PkgFileScope()
	}
	return nil
}

func (f *FnGroupScope) PkgScope() *PkgScope {
	if f.parent != nil {
		return f.parent.PkgScope()
	}
	return nil
}

func (f *FnGroupScope) ProgScope() *ProgScope {
	if f.parent != nil {
		return f.parent.ProgScope()
	}
	return nil
}

func (f *FnGroupScope) Assignments() []*program.Assignment {
	return nil
}

func (f *FnGroupScope) NamedElem(name string) NamedElem {
	if f.parent != nil {
		return f.parent.NamedElem(name)
	}
	return nil
}

func (f *FnGroupScope) Err(loc *program.Location, err error) error {
	return f.PkgFileScope().Err(loc, err)
}
