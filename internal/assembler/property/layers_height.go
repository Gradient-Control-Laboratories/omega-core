package property

import (
	"github.com/Gradient-Control-Laboratories/omega-core/program"
)

// WHICH MATERIALS HAVE A HEIGHT. A shape moves only if its material can carry
// a height, and every height comes through material.withDisplacement. So a
// material may displace when its expression reaches a call to it: directly,
// or through the functions and declarations it calls and names. Names
// resolve as the compiler resolves them, package by package — an unqualified
// name in the package that wrote it or one it dot-imports, a qualified one
// through that package's own import of the prefix — so a namesake in another
// library does not make a plain material look displaced.

// primaryScope keys the primary package among the scopes.
const primaryScope = "\x00primary"

// scope is one package's names, for resolving what its expressions call.
type scope struct {
	funcs    map[string][]*program.Function
	decls    map[string]*program.Operand
	multi    map[string]multiBinding
	prefixes map[string]string // qualified import prefix -> path
	dots     []string          // dot-imported paths
}

// indexScopes records the primary package and every package it imports,
// transitively, as scopes.
func (l *layers) indexScopes(getImport program.GetImportFunc) {
	l.scopes = map[string]*scope{}
	var visit func(key string, pkg *program.Package)
	visit = func(key string, pkg *program.Package) {
		s := &scope{
			funcs:    collectPackageFunctionsList(pkg),
			decls:    collectPackageAssignments(pkg),
			multi:    collectPackageMultiAssignments(pkg),
			prefixes: map[string]string{},
		}
		l.scopes[key] = s
		for _, f := range pkg.Files {
			if f == nil {
				continue
			}
			for _, imp := range f.Imports {
				if imp == nil {
					continue
				}
				if imp.Prefix == "." {
					s.dots = append(s.dots, imp.Path)
				} else {
					s.prefixes[imp.PrefixOrSimpleName()] = imp.Path
				}
				if _, done := l.scopes[imp.Path]; done || getImport == nil {
					continue
				}
				if ip, err := getImport(imp.Path); err == nil && ip != nil {
					visit(imp.Path, ip)
				}
			}
		}
	}
	visit(primaryScope, l.pkg)
	// The builder imports matrix and vector without an import statement.
	for _, auto := range []string{"matrix", "vector"} {
		if _, done := l.scopes[auto]; done || getImport == nil {
			continue
		}
		if ip, err := getImport(auto); err == nil && ip != nil {
			visit(auto, ip)
		}
	}
	for _, auto := range []string{"matrix", "vector"} {
		if _, ok := l.scopes[auto]; ok {
			if _, has := l.scopes[primaryScope].prefixes[auto]; !has {
				l.scopes[primaryScope].prefixes[auto] = auto
			}
		}
	}
}

// lookupFuncs finds the functions an unqualified name calls in scope key:
// the package's own, else a dot-imported package's, transitively, paired
// with the scope each was written in.
func (l *layers) lookupFuncs(key, name string, seen map[string]bool) ([]*program.Function, string) {
	s := l.scopes[key]
	if s == nil || seen[key] {
		return nil, ""
	}
	seen[key] = true
	if fns := s.funcs[name]; len(fns) > 0 {
		return fns, key
	}
	for _, d := range s.dots {
		if fns, at := l.lookupFuncs(d, name, seen); len(fns) > 0 {
			return fns, at
		}
	}
	return nil, ""
}

// lookupDecl finds the declaration an unqualified name refers to in scope
// key, as lookupFuncs does for a function.
func (l *layers) lookupDecl(key, name string, seen map[string]bool) (*program.Operand, string) {
	s := l.scopes[key]
	if s == nil || seen[key] {
		return nil, ""
	}
	seen[key] = true
	if d, ok := s.decls[name]; ok {
		return d, key
	}
	if mb, ok := s.multi[name]; ok {
		return mb.value, key
	}
	for _, d := range s.dots {
		if op, at := l.lookupDecl(d, name, seen); op != nil {
			return op, at
		}
	}
	return nil, ""
}

// target is the scope a reference or call resolves in from scope key: key
// itself for an unqualified name, the imported package for a qualified one.
func (l *layers) target(key, prefix string) string {
	if prefix == "" {
		return key
	}
	if s := l.scopes[key]; s != nil {
		if path, ok := s.prefixes[prefix]; ok {
			return path
		}
	}
	return ""
}

// mayDisplace reports whether an expression written in scope key can carry
// a height.
func (l *layers) mayDisplace(op *program.Operand, key string, visiting map[string]bool) bool {
	if op == nil {
		return false
	}
	switch t := op.T.(type) {
	case *program.Operand_Reference:
		at := l.target(key, t.Reference.Prefix)
		if at == "" {
			return false
		}
		body, in := l.lookupDecl(at, t.Reference.Name, map[string]bool{})
		if body == nil {
			return false
		}
		return l.memo("d|"+in+"|"+t.Reference.Name, visiting, func() bool {
			return l.mayDisplace(body, in, visiting)
		})
	case *program.Operand_Expression:
		for _, o := range t.Expression.Operands {
			if l.mayDisplace(o, key, visiting) {
				return true
			}
		}
		return false
	case *program.Operand_FnCall:
		fc := t.FnCall
		for _, a := range fc.Arguments {
			if l.mayDisplace(a, key, visiting) {
				return true
			}
		}
		if fc.Reference == nil {
			return false
		}
		if fc.Reference.Name == displacementFn {
			return true
		}
		at := l.target(key, fc.Reference.Prefix)
		if at == "" {
			return false
		}
		fns, in := l.lookupFuncs(at, fc.Reference.Name, map[string]bool{})
		if len(fns) == 0 {
			return false
		}
		return l.memo("f|"+in+"|"+fc.Reference.Name, visiting, func() bool {
			for _, fn := range fns {
				if l.functionDisplaces(fn, in, visiting) {
					return true
				}
			}
			return false
		})
	}
	return false
}

// memo answers f once per key, and no to a key it is already answering (a
// recursive function or declaration adds nothing its first visit does not).
func (l *layers) memo(key string, visiting map[string]bool, f func() bool) bool {
	if v, ok := l.displaceMemo[key]; ok {
		return v
	}
	if visiting[key] {
		return false
	}
	visiting[key] = true
	v := f()
	l.displaceMemo[key] = v
	return v
}

// functionDisplaces reports whether a function written in scope key can
// carry a height: anything its body assigns, sets as a property or returns.
func (l *layers) functionDisplaces(fn *program.Function, key string, visiting map[string]bool) bool {
	if fn == nil {
		return false
	}
	for _, a := range fn.Assignments {
		if a == nil {
			continue
		}
		if l.mayDisplace(a.Expression, key, visiting) {
			return true
		}
		for _, v := range a.Variables {
			for _, p := range v.Properties {
				if p != nil && l.mayDisplace(p.Expression, key, visiting) {
					return true
				}
			}
		}
	}
	for _, r := range fn.Returns {
		if l.mayDisplace(r, key, visiting) {
			return true
		}
	}
	for _, ft := range fn.FnTypes {
		if ft == nil {
			continue
		}
		for _, g := range ft.Groups {
			if g == nil {
				continue
			}
			for _, inner := range g.Functions {
				if l.functionDisplaces(inner, key, visiting) {
					return true
				}
			}
		}
	}
	return false
}
