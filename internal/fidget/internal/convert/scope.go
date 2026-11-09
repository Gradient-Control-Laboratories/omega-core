package convert

import "gradientcontrol.com/omega/core/internal/build"

type Scope struct {
	parent         *Scope
	importingScope *Scope
	bScope         *build.Scope
}

func NewScope(parent *Scope, bScope *build.Scope) *Scope {
	return &Scope{
		parent: parent,
		bScope: bScope,
	}
}

func (s *Scope) IsAgentScope() {}

var _ build.AgentScope = (*Scope)(nil)
