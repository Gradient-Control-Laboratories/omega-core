package common

import (
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// Builder implements AST traversal and code emission for backends.
type Builder = build.Builder

// Delegate specifies code generator callbacks during AST traversal.
type Delegate = build.Delegate

// Scope represents a lexical symbol scope in the builder.
type Scope = build.Scope

// Expression represents an emitted expression node.
type Expression = build.Expression

// Build traverses the program AST and emits code via the specified delegate.
func Build(delegate Delegate, prog *program.Program, buildFnEveryTime bool) error {
	return build.Build(delegate, prog, buildFnEveryTime)
}
