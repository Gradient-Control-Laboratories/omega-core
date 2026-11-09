package codegen

import (
	"strings"

	"gradientcontrol.com/omega/core/internal/build"
)

// Delegate implements build.Delegate
type Delegate struct {
	gslSource       string
	jsSource        string
	complexityScore int
	varMap          map[string]build.VarInfo
	rootBScope      *build.Scope
	currBScope      *build.Scope

	builtFns map[string]bool // which functions have been code-generated
	fnDefs   strings.Builder // accumulated JS function definitions
	topDecls strings.Builder // accumulated top-level const declarations across primary + imports
	fnStack  []*fnFrame      // stack of in-progress function compilations
}

// Ensure that Delegate implements the build.Delegate interface.
var _ build.Delegate = (*Delegate)(nil)

func (d *Delegate) currentScope() *Scope {
	return d.currBScope.GetAgentScope().(*Scope)
}

// toJSIdentifier converts GSL identifier to valid JavaScript identifier
func toJSIdentifier(name string) string {
	// For now, simple mapping - could be enhanced
	if name == "time" {
		return "gsl_time"
	}
	return name
}

// pkgQualifiedName mangles `name` to "<sanitized-pkgPath>_<name>" when
// `bScope` belongs to an imported package, leaving primary-pkg names
// untouched. Used so two imports declaring the same identifier (e.g.
// helper() in pkg_a/ and pkg_b/) emit distinct JS symbols.
//
// Sanitization: '/' and '.' become '_'. The simple mapping is good enough
// for the import paths we currently use (e.g. "omega" → "omega"); if
// richer nested paths show up we can swap in a stricter scheme.
func (d *Delegate) pkgQualifiedName(bScope *build.Scope, name string) string {
	if bScope == nil {
		return name
	}
	agentScope, ok := bScope.GetAgentScope().(*Scope)
	if !ok || agentScope == nil {
		return name
	}
	path := ""
	curr := agentScope
	for curr != nil {
		if curr.pkgPath != "" {
			path = curr.pkgPath
			break
		}
		curr = curr.parent
	}
	if path == "" {
		return name
	}
	prefix := sanitizePkgPath(path)
	return prefix + "_" + name
}

func sanitizePkgPath(p string) string {
	b := make([]byte, 0, len(p))
	for i := 0; i < len(p); i++ {
		c := p[i]
		switch c {
		case '/', '.', '-':
			b = append(b, '_')
		default:
			b = append(b, c)
		}
	}
	return string(b)
}

func (d *Delegate) GetScope() *build.Scope {
	return d.currBScope
}

func (d *Delegate) SetScope(bScope *build.Scope) *build.Scope {
	prev := d.currBScope
	d.currBScope = bScope

	s := d.ensureAgentScope(bScope)
	if prev != nil {
		if prevScope, ok := prev.GetAgentScope().(*Scope); ok {
			s.importingScope = prevScope
		}
	}

	return prev
}

func (d *Delegate) ensureAgentScope(bScope *build.Scope) *Scope {
	if existing, ok := bScope.GetAgentScope().(*Scope); ok && existing != nil {
		return existing
	}

	var parentScope *Scope
	if bScope.Parent() != nil {
		parentScope = d.ensureAgentScope(bScope.Parent())
	}

	s := newScope(bScope.Name(), parentScope)
	bScope.SetAgentScope(s)
	return s
}
