package scope

import (
	"fmt"
	"strings"

	"gradientcontrol.com/omega/core/program"
)

// PkgFileScope represents the scope of a single file in a GSL package.
type PkgFileScope struct {
	Pkg        *PkgScope
	File       *program.PackageFile
	Imports    map[string]*PkgRef
	DotImports []*PkgRef
	Variables  map[string]*Var
	Functions  map[string]*FnTypeScope

	exprDepth int

	agentScope AgentScope
}

func (p *PkgFileScope) GetAgentScope() AgentScope {
	return p.agentScope
}

func (p *PkgFileScope) SetAgentScope(agentScope AgentScope) {
	p.agentScope = agentScope
}

var _ Scope = (*PkgFileScope)(nil)
var _ BuilderScope = (*PkgFileScope)(nil)

func (p *PkgFileScope) ContainingScope() Scope {
	return p.Pkg
}

func (p *PkgFileScope) PkgFileScope() *PkgFileScope {
	return p
}

func (p *PkgFileScope) PkgScope() *PkgScope {
	return p.Pkg
}

func (p *PkgFileScope) ProgScope() *ProgScope {
	if p.Pkg != nil {
		return p.Pkg.Prog
	}
	return nil
}

func (p *PkgFileScope) Assignments() []*program.Assignment {
	return p.File.Assignments
}

func (p *PkgFileScope) NamedElem(name string) NamedElem {
	if v, ok := p.Variables[name]; ok {
		return v
	}
	if fn, ok := p.Functions[name]; ok {
		return fn
	}
	if imp, ok := p.Imports[name]; ok {
		return imp
	}
	var builtinDotImp *PkgRef
	for _, dotImp := range p.DotImports {
		if dotImp.pkg != nil && dotImp.pkg.Package != nil && dotImp.pkg.Package.Path == "builtin" {
			builtinDotImp = dotImp
			continue
		}
		if dotImp.pkg != nil {
			elem := dotImp.pkg.namedElemInPkg(name)
			if elem != nil {
				return elem
			}
		}
	}
	if builtinDotImp != nil && builtinDotImp.pkg != nil {
		elem := builtinDotImp.pkg.namedElemInPkg(name)
		if elem != nil {
			return elem
		}
	}
	if p.Pkg != nil {
		return p.Pkg.NamedElem(name)
	}
	return nil
}

func (p *PkgFileScope) Err(loc *program.Location, err error) error {
	pkgName := ""
	if p.Pkg != nil {
		pkgName = p.Pkg.Name()
	}
	filePath := ""
	if p.File != nil {
		filePath = p.File.Path
	}
	return ErrLoc(pkgName, filePath, loc, err)
}

// AddImportRef dynamically registers an imported package scope reference (for linking).
func (p *PkgFileScope) AddImportRef(name string, pkgScope *PkgScope) {
	pkgRef := &PkgRef{
		parent: p,
		name:   name,
		pkg:    pkgScope,
	}
	p.Imports[name] = pkgRef
}

// -------------------------------------------------------------
// BuilderScope implementation
// -------------------------------------------------------------

func (p *PkgFileScope) verifyUndefinedLocally(name string) error {
	if name == program.PVarName || name == program.TimeVarName {
		return fmt.Errorf("%s cannot be reassigned", name)
	}
	if _, isImported := p.Imports[name]; isImported {
		return fmt.Errorf("%q is an existing import prefix", name)
	}
	if _, isAssigned := p.Variables[name]; isAssigned {
		return fmt.Errorf("%q is an existing variable", name)
	}
	if _, isFunc := p.Functions[name]; isFunc {
		return fmt.Errorf("%q is an existing function", name)
	}
	if p.Pkg != nil {
		if p.Pkg.FindVarInOtherFiles(name, p) != nil {
			return fmt.Errorf("%q is an existing variable", name)
		}
		if p.Pkg.FindFnTypeInOtherFiles(name, p) != nil {
			return fmt.Errorf("%q is an existing function", name)
		}
	}
	return nil
}

func (p *PkgFileScope) AddAssignment(names []string, expr *program.Operand, start, end *program.Location, rng *program.Range) error {
	for _, name := range names {
		if err := p.verifyUndefinedLocally(name); err != nil {
			return err
		}
	}

	newAssignment := program.NewAssignment(names, expr, start, end)
	newAssignment.Range = rng
	p.File.Assignments = append(p.File.Assignments, newAssignment)

	assignScope := &AssignScope{Parent: p, Assign: newAssignment}
	for _, v := range newAssignment.Variables {
		p.Variables[v.Name] = &Var{Parent: assignScope, Var: v}
	}
	return nil
}

func (p *PkgFileScope) SetParam(index int, start *program.Location, rng *program.Range) {
	panic("SetParam cannot be called on PkgFileScope")
}

func (p *PkgFileScope) SetParamKind(index int, kind, anchor string) {
	panic("SetParamKind cannot be called on PkgFileScope")
}

var validAssignablePropNames = []string{"min", "max", "period", "color", "material", "box", "simple", "detail", "detailSize", "u", "v"}

// invalidPropertyError is the error for a property name outside
// validAssignablePropNames, listing the valid names so the message tracks the
// list.
func invalidPropertyError(propName string) error {
	quoted := make([]string, len(validAssignablePropNames))
	for i, name := range validAssignablePropNames {
		quoted[i] = fmt.Sprintf("%q", name)
	}
	return fmt.Errorf("only %s properties may be assigned, got %q", strings.Join(quoted, ", "), propName)
}

func (p *PkgFileScope) AddPropertyAssignment(varName, propName string, expr *program.Operand, start, end *program.Location) error {
	variable, ok := p.Variables[varName]
	if !ok {
		return fmt.Errorf("variable %q is not defined in this scope", varName)
	}

	found := false
	for _, valid := range validAssignablePropNames {
		if propName == valid {
			found = true
			break
		}
	}
	if !found {
		return invalidPropertyError(propName)
	}

	for _, prop := range variable.Var.Properties {
		if prop.Name == propName {
			return fmt.Errorf("property %s.%s already assigned", varName, propName)
		}
	}

	variable.Var.AddPropertyAssignment(propName, expr, start, end)
	if propName == "min" || propName == "max" || propName == "period" {
		synthesizeRangeAnnotation(variable)
	}
	return nil
}

// AddAnnotation records a declaration-level annotation on the file. The
// annotated declaration itself is built separately and unchanged, so the only
// thing this adds is the note that the symbol carries the annotation.
func (p *PkgFileScope) AddAnnotation(kind, symbol string, start, end *program.Location) error {
	for _, existing := range p.File.Annotations {
		if existing.Kind == kind && existing.Symbol == symbol {
			return fmt.Errorf("%q is already annotated @%s", symbol, kind)
		}
	}
	p.File.Annotations = append(p.File.Annotations, &program.Annotation{
		Kind: kind, Symbol: symbol,
		Start: start, End: end,
	})
	return nil
}

func synthesizeRangeAnnotation(variable *Var) {
	assign := variable.Parent.Assign
	if assign.Range != nil && !assign.Range.IsSynthetic() {
		return
	}

	var minProp, maxProp, periodProp *program.Property
	for _, prop := range variable.Var.Properties {
		switch prop.Name {
		case "min":
			minProp = prop
		case "max":
			maxProp = prop
		case "period":
			periodProp = prop
		}
	}

	if minProp == nil || maxProp == nil {
		return
	}

	loc := assign.Start
	if assign.Expression != nil && assign.Expression.Start != nil {
		loc = assign.Expression.Start
	}

	var periodOp *program.Operand
	if periodProp != nil {
		periodOp = periodProp.Expression
	}

	assign.Range = &program.Range{
		Start:  loc,
		End:    loc,
		Min:    minProp.Expression,
		Max:    maxProp.Expression,
		Period: periodOp,
	}
}

func (p *PkgFileScope) AddImport(imp *program.Import) error {
	for _, existing := range p.File.Imports {
		if imp.Path == existing.Path {
			// Legacy omega compatibility gives vector both dot and named prefixes.
			if (imp.Path == "builtin" || imp.Path == "vector") && imp.PrefixOrSimpleName() != existing.PrefixOrSimpleName() {
				continue
			}
			return fmt.Errorf("Cannot import the same package more than once: %s", imp.Path)
		}
	}

	prefix := imp.PrefixOrSimpleName()
	if prefix != "." {
		if _, isImported := p.Imports[prefix]; isImported && (imp.Path == "builtin" || imp.Path == "vector") {
			// Already registered import prefix for legacy builtin/vector compatibility.
		} else if err := p.verifyUndefinedLocally(prefix); err != nil {
			return err
		}
		// Register it locally as a prefix reference
		p.Imports[prefix] = &PkgRef{
			parent: p,
			name:   prefix,
		}
	} else {
		simpleName := imp.SimpleName()
		if simpleName != "" && simpleName != "." {
			if _, exists := p.Imports[simpleName]; !exists {
				p.Imports[simpleName] = &PkgRef{
					parent: p,
					name:   simpleName,
				}
			}
		}
	}

	p.File.AddImport(imp)
	return nil
}

func (p *PkgFileScope) AddMeshImport(meshImport *program.MeshImport) error {
	if p.Pkg != nil {
		err := p.Pkg.AddMeshImport(meshImport)
		if err != nil {
			return err
		}
	}
	p.File.MeshImports = append(p.File.MeshImports, meshImport)
	return nil
}

func (p *PkgFileScope) AddStepImport(stepImport *program.StepImport) error {
	if p.Pkg != nil {
		err := p.Pkg.AddStepImport(stepImport)
		if err != nil {
			return err
		}
	}
	p.File.StepImports = append(p.File.StepImports, stepImport)
	return nil
}

func (p *PkgFileScope) AddWarpImport(warpImport *program.WarpImport) error {
	if p.Pkg != nil {
		err := p.Pkg.AddWarpImport(warpImport)
		if err != nil {
			return err
		}
	}
	p.File.WarpImports = append(p.File.WarpImports, warpImport)
	return nil
}

func (p *PkgFileScope) IncExprDepth() error {
	if p.exprDepth >= 127 {
		return fmt.Errorf("expression depth limit exceeded")
	}
	p.exprDepth++
	return nil
}

func (p *PkgFileScope) DecExprDepth() {
	p.exprDepth--
}

func (p *PkgFileScope) NewFnScope(fnName string, params []string) (BuilderScope, error) {
	fnTypes, fnScope, err := p.addFunction(p.File.FnTypes, fnName, params)
	if err != nil {
		return nil, err
	}
	p.File.FnTypes = fnTypes
	return fnScope, nil
}

func (p *PkgFileScope) SetReturnOperands(operands []*program.Operand) {
	p.File.Returns = operands
}

func (p *PkgFileScope) addFunction(
	fnTypes []*program.FnType, fnName string, paramNames []string,
) ([]*program.FnType, *FnScope, error) {
	seen := make(map[string]bool)
	for _, paramName := range paramNames {
		if seen[paramName] {
			return fnTypes, nil, fmt.Errorf("duplicate parameter name %q", paramName)
		}
		seen[paramName] = true
	}

	if fnName == program.PVarName || fnName == program.TimeVarName {
		return fnTypes, nil, fmt.Errorf("%s cannot be reassigned", fnName)
	}
	if _, isImported := p.Imports[fnName]; isImported {
		return fnTypes, nil, fmt.Errorf("%q is an existing import prefix", fnName)
	}
	if _, isVar := p.Variables[fnName]; isVar {
		return fnTypes, nil, fmt.Errorf("%q is an existing variable", fnName)
	}
	if p.Pkg != nil && p.Pkg.FindVarInOtherFiles(fnName, p) != nil {
		return fnTypes, nil, fmt.Errorf("%q is an existing variable", fnName)
	}

	fnTypeScope, ok := p.Functions[fnName]
	if !ok {
		fnTypeScope = &FnTypeScope{
			parent: p,
			FnType: &program.FnType{Name: fnName},
			Groups: make(map[int32]*FnGroupScope),
		}
		if p.Pkg != nil {
			for _, file := range p.Pkg.Files {
				if file == p {
					continue
				}
				if otherFn, exists := file.Functions[fnName]; exists {
					for k, v := range otherFn.Groups {
						fnTypeScope.Groups[k] = v
					}
				}
			}
		}
		p.Functions[fnName] = fnTypeScope
	}

	paramCount := len(paramNames)
	if _, ok := fnTypeScope.Groups[int32(paramCount)]; ok {
		return fnTypes, nil, fmt.Errorf(
			"function %q with %d parameters is already defined", fnName, paramCount,
		)
	}

	fnTypes, fn := program.AddFunction(fnTypes, fnName, paramNames)
	// Sync the FnType fields since program.AddFunction might have modified/reallocated it
	for _, ft := range fnTypes {
		if ft.Name == fnName {
			fnTypeScope.FnType = ft
		}
	}

	groupScope := &FnGroupScope{
		parent:  fnTypeScope,
		FnGroup: &program.FnGroup{ParamCount: int32(paramCount)},
	}
	fnScope := &FnScope{
		parent:   groupScope,
		Function: fn,
	}
	groupScope.Functions = append(groupScope.Functions, fnScope)
	fnTypeScope.Groups[int32(paramCount)] = groupScope

	if p.Pkg != nil {
		for _, file := range p.Pkg.Files {
			if file == p {
				continue
			}
			if otherFn, exists := file.Functions[fnName]; exists {
				otherFn.Groups[int32(paramCount)] = groupScope
			}
		}
	}

	return fnTypes, fnScope, nil
}

// -------------------------------------------------------------
// PkgRef implementation
// -------------------------------------------------------------

// PkgRef represents a prefix reference to an imported package.
type PkgRef struct {
	parent *PkgFileScope
	name   string
	pkg    *PkgScope
}

var _ Scope = (*PkgRef)(nil)

func (p *PkgRef) ContainingScope() Scope {
	return p.parent
}

func (p *PkgRef) GetAgentScope() AgentScope {
	return nil
}

func (p *PkgRef) SetAgentScope(agentScope AgentScope) {}

func (p *PkgRef) PkgFileScope() *PkgFileScope {
	return p.parent
}

func (p *PkgRef) PkgScope() *PkgScope {
	return p.pkg
}

func (p *PkgRef) ProgScope() *ProgScope {
	if p.pkg != nil {
		return p.pkg.Prog
	}
	return nil
}

func (p *PkgRef) Assignments() []*program.Assignment {
	return nil
}

func (p *PkgRef) Name() string {
	return p.name
}

func (p *PkgRef) NamedElem(name string) NamedElem {
	if p.pkg != nil {
		return p.pkg.NamedElem(name)
	}
	return nil
}

func (p *PkgRef) Err(loc *program.Location, err error) error {
	return p.parent.Err(loc, err)
}

func (p *PkgRef) PkgScopePtr() *PkgScope {
	return p.pkg
}

func (p *PkgRef) SetPkgScopePtr(pkg *PkgScope) {
	p.pkg = pkg
}

// PkgScope helper methods for import management (called by PkgFileScope)
func (s *PkgScope) AddMeshImport(meshImport *program.MeshImport) error {
	if s.meshPaths == nil {
		s.meshPaths = make(map[string]bool)
		s.meshImports = make(map[int]*program.MeshImport)
	}
	if s.meshPaths[meshImport.Path] {
		return fmt.Errorf("mesh path %s is already imported", meshImport.Path)
	}
	if s.meshImports[int(meshImport.Index)] != nil {
		return fmt.Errorf("mesh slot %d is already imported", meshImport.Index)
	}
	s.meshPaths[meshImport.Path] = true
	s.meshImports[int(meshImport.Index)] = meshImport
	return nil
}

func (s *PkgScope) AddStepImport(stepImport *program.StepImport) error {
	if s.stepPaths == nil {
		s.stepPaths = make(map[string]bool)
		s.stepImports = make(map[int]*program.StepImport)
	}
	if s.stepPaths[stepImport.Path] {
		return fmt.Errorf("step path %s is already imported", stepImport.Path)
	}
	if s.stepImports[int(stepImport.Index)] != nil {
		return fmt.Errorf("mesh slot %d is already imported", stepImport.Index)
	}
	s.stepPaths[stepImport.Path] = true
	s.stepImports[int(stepImport.Index)] = stepImport
	return nil
}

func (s *PkgScope) AddWarpImport(warpImport *program.WarpImport) error {
	if s.warpPaths == nil {
		s.warpPaths = make(map[string]bool)
		s.warpImports = make(map[int]*program.WarpImport)
	}
	if s.warpPaths[warpImport.Path] {
		return fmt.Errorf("warp path %s is already imported", warpImport.Path)
	}
	if s.warpImports[int(warpImport.Index)] != nil {
		return fmt.Errorf("warp slot %d is already imported", warpImport.Index)
	}
	s.warpPaths[warpImport.Path] = true
	s.warpImports[int(warpImport.Index)] = warpImport
	return nil
}

func initPkgScope(progScope *ProgScope, pkg *program.Package) (*PkgScope, error) {
	pkgScope := NewPkgScope(progScope, pkg)

	for _, file := range pkg.Files {
		fileScope := pkgScope.NewPkgFileScope(file)

		// Add package-level variables from assignments in this file
		for _, assign := range file.Assignments {
			assignScope := &AssignScope{Parent: fileScope, Assign: assign}
			for _, v := range assign.Variables {
				fileScope.Variables[v.Name] = &Var{Parent: assignScope, Var: v}
			}
		}

		// Add package-level functions declared in this file
		for _, fnType := range file.FnTypes {
			typeScope := &FnTypeScope{
				parent: fileScope,
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
			fileScope.Functions[fnType.Name] = typeScope
		}
	}

	return pkgScope, nil
}
