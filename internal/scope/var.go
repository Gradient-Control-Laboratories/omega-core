package scope

import (
	"gradientcontrol.com/omega/core/program"
)

// AssignScope links an assignment statement to its parent scope.
type AssignScope struct {
	Parent Scope
	Assign *program.Assignment
}

// Var represents a variable defined in an assignment statement.
type Var struct {
	Parent *AssignScope
	Var    *program.Variable
}

var _ NamedElem = (*Var)(nil)

func (v *Var) Name() string {
	return v.Var.Name
}

func (v *Var) ContainingScope() Scope {
	return v.Parent.Parent
}

// Param represents a parameter defined in a function signature.
type Param struct {
	Parent *FnScope
	Param  *program.Parameter
}

var _ NamedElem = (*Param)(nil)

func (p *Param) Name() string {
	return p.Param.Name
}

func (p *Param) ContainingScope() Scope {
	return p.Parent
}
