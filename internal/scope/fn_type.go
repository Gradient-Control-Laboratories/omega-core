package scope

import (
	"gradientcontrol.com/omega/core/program"
)

// FnTypeScope represents the overloaded groups for a function name.
type FnTypeScope struct {
	parent Scope
	FnType *program.FnType
	Groups map[int32]*FnGroupScope
}

var _ Scope = (*FnTypeScope)(nil)
var _ NamedElem = (*FnTypeScope)(nil)

func (f *FnTypeScope) ContainingScope() Scope {
	return f.parent
}

func (f *FnTypeScope) GetAgentScope() AgentScope {
	return nil
}

func (f *FnTypeScope) SetAgentScope(agentScope AgentScope) {}

func (f *FnTypeScope) PkgFileScope() *PkgFileScope {
	if f.parent != nil {
		return f.parent.PkgFileScope()
	}
	return nil
}

func (f *FnTypeScope) PkgScope() *PkgScope {
	if f.parent != nil {
		return f.parent.PkgScope()
	}
	return nil
}

func (f *FnTypeScope) ProgScope() *ProgScope {
	if f.parent != nil {
		return f.parent.ProgScope()
	}
	return nil
}

func (f *FnTypeScope) Assignments() []*program.Assignment {
	return nil
}

func (f *FnTypeScope) NamedElem(name string) NamedElem {
	if f.parent != nil {
		return f.parent.NamedElem(name)
	}
	return nil
}

func (f *FnTypeScope) Name() string {
	return f.FnType.Name
}

func (f *FnTypeScope) Err(loc *program.Location, err error) error {
	return f.PkgFileScope().Err(loc, err)
}
