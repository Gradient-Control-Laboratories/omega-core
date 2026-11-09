package build

import (
	"fmt"
	"sort"
	"strconv"

	"gradientcontrol.com/omega/core/program"
)

// VarInfo holds information about a cached variable.
type VarInfo interface {
	Name() string
	Scope() *Scope
	Expr() Expression
	Type() *program.Type
}

// Return the variable with the given name if it exists in this scope
// or any of its parent scopes. Otherwise return nil.
func (s *Scope) GetVar(name string) VarInfo {
	if v := s.getLexicalVar(name); v != nil {
		return v
	}
	curr := s
	for curr != nil {
		if v := curr.getDotImportedVar(name); v != nil {
			return v
		}
		curr = curr.parent
	}
	return nil
}

// Return the variable with the given name if it exists as a lexical variable
// (local, parameter, or function) in this scope or any of its parent scopes.
// Lexical variables always take precedence over dot-imported variables.
func (s *Scope) getLexicalVar(name string) VarInfo {
	if v := s.GetLocalVar(name); v != nil {
		return v
	}
	if s.parent != nil {
		return s.parent.getLexicalVar(name)
	}
	return nil
}

// Return the variable with the given name if it exists in this scope.
// If this is a package-level scope, then also return the variable if it exists
// in any of the dot-imported packages.  Otherwise return nil.
func (s *Scope) GetLocalNamespaceVar(name string) VarInfo {
	varInfo := s.GetLocalVar(name)
	if varInfo != nil {
		return varInfo
	}
	return s.getDotImportedVar(name)
}

// Return the variable with the given name if it exists in this scope
// excluding any dot-imported packages.  Otherwise return nil.
func (s *Scope) GetLocalVar(name string) VarInfo {
	if varInfo := s.localVars[name]; varInfo != nil {
		return varInfo
	}
	if varInfo := s.paramVars[name]; varInfo != nil {
		return varInfo
	}
	if fnVar := s.fnVars[name]; fnVar != nil {
		return fnVar
	}
	return nil
}

// Return true if a variable or function with the given name exists in this scope
// excluding any dot-imported packages.
func (s *Scope) IsLocalVar(name string) bool {
	return s.GetLocalVar(name) != nil
}

// Return a list of the parameters in this scope
func (s *Scope) ParamVars() []VarInfo {
	result := make([]VarInfo, 0, len(s.paramVars))
	for _, v := range s.paramVars {
		result = append(result, v)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name() < result[j].Name()
	})
	return result
}

// Return the variable with the given name if it exists in any of the
// dot-imported packages.  Otherwise return nil.
func (s *Scope) getDotImportedVar(name string) VarInfo {
	// Track matches from the implicit "builtin" scope but defer returning them.
	// This ensures that other dot-imported packages take precedence and can shadow built-ins.
	var builtinVar VarInfo
	for _, impScope := range s.dotImports {
		if impScope.name == "builtin" {
			if varInfo := impScope.localVars[name]; varInfo != nil {
				builtinVar = varInfo
			} else if fnVar := impScope.fnVars[name]; fnVar != nil {
				builtinVar = fnVar
			}
			continue
		}
		if varInfo := impScope.localVars[name]; varInfo != nil {
			return varInfo
		}
		if fnVar := impScope.fnVars[name]; fnVar != nil {
			return fnVar
		}
	}
	if builtinVar != nil {
		return builtinVar
	}
	return nil
}

// Return a new unique var name for the current scope
// derived from the given name
func (s *Scope) NewVarName(name string) string {
	if s.parent != nil && s.IsFileScope() {
		return s.parent.NewVarName(name)
	}

	result := name
	for i := 2; s.GetLocalNamespaceVar(result) != nil; i++ {
		result = name + "_" + strconv.Itoa(i)
	}
	return result
}

func (s *Scope) AddLocalVar(varInfo VarInfo) error {
	name := varInfo.Name()

	// If this is a file-level scope, promote the variable to the package-level scope
	if s.parent != nil && s.IsFileScope() {
		return s.parent.AddLocalVar(varInfo)
	}

	// Invariant: Upstream parser enforces unique declarations across all files in a package.

	if s.localVars == nil {
		s.localVars = make(map[string]VarInfo)
	}
	s.localVars[name] = varInfo
	return nil
}

func (s *Scope) addParam(varInfo VarInfo) error {
	if s.IsFileScope() {
		return fmt.Errorf("file scopes cannot have parameters")
	}

	// Parameter names are validated by the parser

	if s.paramVars == nil {
		s.paramVars = make(map[string]VarInfo)
	}
	s.paramVars[varInfo.Name()] = varInfo
	return nil
}
