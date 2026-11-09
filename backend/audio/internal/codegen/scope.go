package codegen

import (
	"strings"

	"gradientcontrol.com/omega/core/internal/build"
)

type Scope struct {
	name           string
	parent         *Scope
	importingScope *Scope
	content        strings.Builder

	// pkgPath identifies the package this scope represents. Empty for the
	// primary scope; set to imp.Path by BuildImport. Function and constant
	// names declared in non-primary scopes get mangled to "<sanitized-
	// pkgPath>_<name>" so two imports defining helper() don't collide in
	// the flat JS namespace.
	pkgPath string
}

func newScope(name string, parent *Scope) *Scope {
	return &Scope{
		name:   name,
		parent: parent,
	}
}

func (s *Scope) IsAgentScope() {}

var _ build.AgentScope = (*Scope)(nil)
