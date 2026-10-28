package scope

import (
	"fmt"

	"gradientcontrol.com/omega/core/program"
)

// ProgScope represents the top-level scope enclosing an entire GSL program.
type ProgScope struct {
	Program      *program.Program
	Primary      *PkgScope
	Dependencies map[string]*PkgScope

	agentScope AgentScope
}

func (p *ProgScope) GetAgentScope() AgentScope {
	return p.agentScope
}

func (p *ProgScope) SetAgentScope(agentScope AgentScope) {
	p.agentScope = agentScope
}

var _ Scope = (*ProgScope)(nil)

// Init parses a *program.Program AST and builds the full program scope structure.
func Init(prog *program.Program) (*ProgScope, error) {
	if prog == nil {
		return nil, fmt.Errorf("program cannot be nil")
	}

	progScope := &ProgScope{
		Program:      prog,
		Dependencies: make(map[string]*PkgScope),
	}

	// Create scope for the primary package
	primaryScope, err := initPkgScope(progScope, prog.Primary)
	if err != nil {
		return nil, err
	}
	progScope.Primary = primaryScope

	// Create scopes for dependencies
	for _, dep := range prog.Dependencies {
		depScope, err := initPkgScope(progScope, dep)
		if err != nil {
			return nil, err
		}
		progScope.Dependencies[dep.Path] = depScope
	}

	// Build PkgRefs for all imports in all PkgFileScopes
	allPkgScopes := make([]*PkgScope, 0, 1+len(progScope.Dependencies))
	allPkgScopes = append(allPkgScopes, progScope.Primary)
	for _, depScope := range progScope.Dependencies {
		allPkgScopes = append(allPkgScopes, depScope)
	}

	for _, pkgScope := range allPkgScopes {
		for _, fileScope := range pkgScope.Files {
			for _, imp := range fileScope.File.Imports {
				importedPkgScope, foundPkgScope := progScope.Dependencies[imp.Path]
				if !foundPkgScope {
					continue
				}
				impName := imp.PrefixOrSimpleName()
				pkgRef := &PkgRef{
					parent: fileScope,
					name:   impName,
					pkg:    importedPkgScope,
				}
				if impName == "." {
					fileScope.DotImports = append(fileScope.DotImports, pkgRef)
					simpleName := imp.SimpleName()
					if simpleName != "" && fileScope.Imports[simpleName] == nil {
						fileScope.Imports[simpleName] = &PkgRef{
							parent: fileScope,
							name:   simpleName,
							pkg:    importedPkgScope,
						}
					}
				} else {
					fileScope.Imports[impName] = pkgRef
				}
			}
		}
	}

	return progScope, nil
}

func (p *ProgScope) ContainingScope() Scope {
	return nil
}

func (p *ProgScope) PkgFileScope() *PkgFileScope {
	return nil
}

func (p *ProgScope) PkgScope() *PkgScope {
	return nil
}

func (p *ProgScope) ProgScope() *ProgScope {
	return p
}

func (p *ProgScope) Assignments() []*program.Assignment {
	return nil
}

func (p *ProgScope) NamedElem(name string) NamedElem {
	return nil
}

func (p *ProgScope) Err(loc *program.Location, err error) error {
	return ErrLoc("", "", loc, err)
}

// AddDependency dynamically registers a new package dependency and returns its scope.
func (p *ProgScope) AddDependency(pkg *program.Package) (*PkgScope, error) {
	if s, ok := p.Dependencies[pkg.Path]; ok {
		return s, nil
	}
	depScope, err := initPkgScope(p, pkg)
	if err != nil {
		return nil, err
	}
	p.Dependencies[pkg.Path] = depScope

	// Build PkgRefs for imports in the new dependency's files
	for _, fileScope := range depScope.Files {
		for _, imp := range fileScope.File.Imports {
			importedPkgScope, foundPkgScope := p.Dependencies[imp.Path]
			if !foundPkgScope {
				continue
			}
			impName := imp.PrefixOrSimpleName()
			pkgRef := &PkgRef{
				parent: fileScope,
				name:   impName,
				pkg:    importedPkgScope,
			}
			if impName == "." {
				fileScope.DotImports = append(fileScope.DotImports, pkgRef)
				simpleName := imp.SimpleName()
				if simpleName != "" && fileScope.Imports[simpleName] == nil {
					fileScope.Imports[simpleName] = &PkgRef{
						parent: fileScope,
						name:   simpleName,
						pkg:    importedPkgScope,
					}
				}
			} else {
				fileScope.Imports[impName] = pkgRef
			}
		}
	}
	return depScope, nil
}
