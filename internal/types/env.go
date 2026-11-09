package types

import (
	"gradientcontrol.com/omega/core/program"
)

// TypeEnv represents a nested symbol table mapping names to GSL types.
type TypeEnv struct {
	parent *TypeEnv
	vars   map[string]*program.Type
	fns    map[string]*program.FnType
}

func newEnv(parent *TypeEnv) *TypeEnv {
	return &TypeEnv{
		parent: parent,
		vars:   make(map[string]*program.Type),
		fns:    make(map[string]*program.FnType),
	}
}

func (e *TypeEnv) getVar(name string) *program.Type {
	if t, ok := e.vars[name]; ok {
		return t
	}
	if e.parent != nil {
		return e.parent.getVar(name)
	}
	return nil
}

func (e *TypeEnv) getFn(name string) *program.FnType {
	if f, ok := e.fns[name]; ok {
		return f
	}
	if e.parent != nil {
		return e.parent.getFn(name)
	}
	return nil
}

func (e *TypeEnv) setVar(name string, t *program.Type) {
	e.vars[name] = t
}

type FileImports struct {
	named map[string]*TypeEnv
	dots  []*TypeEnv
}

type pkgVarDef struct {
	assignment *program.Assignment
	file       *program.PackageFile
	index      int
}
