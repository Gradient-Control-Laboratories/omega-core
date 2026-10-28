package scope

import (
	"gradientcontrol.com/omega/core/program"
)

// PkgScope represents the scope of an entire GSL package.
type PkgScope struct {
	Prog    *ProgScope
	Package *program.Package
	Files   []*PkgFileScope

	meshPaths   map[string]bool
	meshImports map[int]*program.MeshImport
	stepPaths   map[string]bool
	stepImports map[int]*program.StepImport
	warpPaths   map[string]bool
	warpImports map[int]*program.WarpImport

	agentScope AgentScope
}

func (p *PkgScope) GetAgentScope() AgentScope {
	return p.agentScope
}

func (p *PkgScope) SetAgentScope(agentScope AgentScope) {
	p.agentScope = agentScope
}

func NewPkgScope(prog *ProgScope, pkg *program.Package) *PkgScope {
	return &PkgScope{
		Prog:        prog,
		Package:     pkg,
		meshPaths:   make(map[string]bool),
		meshImports: make(map[int]*program.MeshImport),
		stepPaths:   make(map[string]bool),
		stepImports: make(map[int]*program.StepImport),
		warpPaths:   make(map[string]bool),
		warpImports: make(map[int]*program.WarpImport),
	}
}

func (p *PkgScope) NewPkgFileScope(file *program.PackageFile) *PkgFileScope {
	fileScope := &PkgFileScope{
		Pkg:       p,
		File:      file,
		Variables: make(map[string]*Var),
		Functions: make(map[string]*FnTypeScope),
		Imports:   make(map[string]*PkgRef),
	}
	p.Files = append(p.Files, fileScope)
	return fileScope
}

func (p *PkgScope) ContainingScope() Scope {
	return nil
}

func (p *PkgScope) PkgFileScope() *PkgFileScope {
	return nil
}

func (p *PkgScope) PkgScope() *PkgScope {
	return p
}

func (p *PkgScope) ProgScope() *ProgScope {
	return p.Prog
}

func (p *PkgScope) Assignments() []*program.Assignment {
	return nil
}

func (p *PkgScope) NamedElem(name string) NamedElem {
	elem := p.namedElemInPkg(name)
	if elem != nil {
		return elem
	}
	return nil
}

func (p *PkgScope) Err(loc *program.Location, err error) error {
	return ErrLoc(p.Package.Path, "", loc, err)
}

func (p *PkgScope) namedElemInPkg(name string) NamedElem {
	for _, pkgFile := range p.Files {
		if v, ok := pkgFile.Variables[name]; ok {
			return v
		}
		if fn, ok := pkgFile.Functions[name]; ok {
			return fn
		}
	}
	return nil
}

// FindVarInOtherFiles searches for a variable named name declared in any file of this package
// other than the excluded file.
func (p *PkgScope) FindVarInOtherFiles(name string, exclude *PkgFileScope) *Var {
	for _, pkgFile := range p.Files {
		if pkgFile == exclude {
			continue
		}
		if v, ok := pkgFile.Variables[name]; ok {
			return v
		}
	}
	return nil
}

// FindFnTypeInOtherFiles searches for a function named name declared in any file of this package
// other than the excluded file.
func (p *PkgScope) FindFnTypeInOtherFiles(name string, exclude *PkgFileScope) *FnTypeScope {
	for _, pkgFile := range p.Files {
		if pkgFile == exclude {
			continue
		}
		if fn, ok := pkgFile.Functions[name]; ok {
			return fn
		}
	}
	return nil
}

func (p *PkgScope) Name() string {
	return p.Package.Path
}

// ReloadDeclarations copies functions and assignments from freshPkg into the package and its scopes.
func (p *PkgScope) ReloadDeclarations(freshPkg *program.Package) {
	if p == nil || freshPkg == nil || len(freshPkg.Files) == 0 || len(p.Files) == 0 {
		return
	}
	targetFileScope := p.Files[0]
	targetFile := targetFileScope.File

	for _, freshFile := range freshPkg.Files {
		for _, assign := range freshFile.Assignments {
			exists := false
			for _, v := range assign.Variables {
				if _, ok := targetFileScope.Variables[v.Name]; ok {
					exists = true
					break
				}
			}
			if !exists {
				targetFile.Assignments = append(targetFile.Assignments, assign)
				assignScope := &AssignScope{Parent: targetFileScope, Assign: assign}
				for _, v := range assign.Variables {
					targetFileScope.Variables[v.Name] = &Var{Parent: assignScope, Var: v}
				}
			}
		}

		for _, fnType := range freshFile.FnTypes {
			if _, ok := targetFileScope.Functions[fnType.Name]; !ok {
				targetFile.FnTypes = append(targetFile.FnTypes, fnType)
				typeScope := &FnTypeScope{
					parent: targetFileScope,
					FnType: fnType,
					Groups: make(map[int32]*FnGroupScope),
				}
				for _, group := range fnType.Groups {
					groupScope := &FnGroupScope{parent: typeScope, FnGroup: group}
					for _, fn := range group.Functions {
						fnScope := &FnScope{parent: groupScope, Function: fn}
						groupScope.Functions = append(groupScope.Functions, fnScope)
					}
					typeScope.Groups[group.ParamCount] = groupScope
				}
				targetFileScope.Functions[fnType.Name] = typeScope
			}
		}
	}
}
