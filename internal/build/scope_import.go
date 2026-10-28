package build

import (
	"fmt"

	"gradientcontrol.com/omega/core/program"
)

type ImportInfo interface {
	Import() *program.Import
	Package() *program.Package
}

// Return the scope for the import with the given alias or prefix in this scope
// or any of its parent scopes. Otherwise return nil.
func (s *Scope) GetNamedImportScope(name string) *Scope {
	if s.namedImports != nil {
		if scope, ok := s.namedImports[name]; ok {
			return scope
		}
	}
	if s.parent != nil {
		return s.parent.GetNamedImportScope(name)
	}
	return nil
}

func (s *Scope) addImport(imp *program.Import, impScope *Scope) error {
	if !s.isFileScope {
		return fmt.Errorf("only file scopes can have imports")
	}
	name := imp.PrefixOrSimpleName()
	if name != "." {

		// Parser has already validated that there are no import prefix conflicts.

		if s.namedImports == nil {
			s.namedImports = make(map[string]*Scope)
		}
		s.namedImports[name] = impScope
		return nil
	}

	if s.namedImports == nil {
		s.namedImports = make(map[string]*Scope)
	}
	simpleName := imp.SimpleName()
	if simpleName != "" && s.namedImports[simpleName] == nil {
		s.namedImports[simpleName] = impScope
	}

	s.dotImports = append(s.dotImports, impScope)
	return nil
}
