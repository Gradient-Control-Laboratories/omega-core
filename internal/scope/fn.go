package scope

import (
	"fmt"

	"gradientcontrol.com/omega/core/program"
)

// FnScope represents the scope inside a GSL function body.
type FnScope struct {
	parent     *FnGroupScope
	Function   *program.Function
	Variables  map[string]*Var
	Functions  map[string]*FnTypeScope
	Parameters map[string]*Param

	agentScope AgentScope
}

func (f *FnScope) GetAgentScope() AgentScope {
	return f.agentScope
}

func (f *FnScope) SetAgentScope(agentScope AgentScope) {
	f.agentScope = agentScope
}

var _ Scope = (*FnScope)(nil)
var _ BuilderScope = (*FnScope)(nil)

func (f *FnScope) ContainingScope() Scope {
	if f.parent != nil && f.parent.parent != nil {
		return f.parent.parent.parent
	}
	return nil
}

func (f *FnScope) PkgFileScope() *PkgFileScope {
	if f.parent != nil {
		return f.parent.PkgFileScope()
	}
	return nil
}

func (f *FnScope) PkgScope() *PkgScope {
	if f.parent != nil {
		return f.parent.PkgScope()
	}
	return nil
}

func (f *FnScope) ProgScope() *ProgScope {
	if f.parent != nil {
		return f.parent.ProgScope()
	}
	return nil
}

func (f *FnScope) Assignments() []*program.Assignment {
	return f.Function.Assignments
}

func (f *FnScope) init() {
	if f.Variables != nil {
		return
	}

	f.Variables = make(map[string]*Var)
	f.Functions = make(map[string]*FnTypeScope)
	f.Parameters = make(map[string]*Param)

	// Add parameters
	for _, param := range f.Function.Parameters {
		f.Parameters[param.Name] = &Param{
			Parent: f,
			Param:  param,
		}
	}

	// Add local variables from assignments
	for _, assign := range f.Function.Assignments {
		assignScope := &AssignScope{Parent: f, Assign: assign}
		for _, v := range assign.Variables {
			f.Variables[v.Name] = &Var{Parent: assignScope, Var: v}
		}
	}

	// Traverse nested functions
	for _, cFnType := range f.Function.FnTypes {
		cFnTypeScope := &FnTypeScope{
			parent: f,
			FnType: cFnType,
			Groups: make(map[int32]*FnGroupScope),
		}
		for _, cFnGroup := range cFnType.Groups {
			cFnGroupScope := &FnGroupScope{
				parent:  cFnTypeScope,
				FnGroup: cFnGroup,
			}
			for _, cFn := range cFnGroup.Functions {
				cFnScope := &FnScope{parent: cFnGroupScope, Function: cFn}
				cFnGroupScope.Functions = append(cFnGroupScope.Functions, cFnScope)
			}
			cFnTypeScope.Groups[cFnGroup.ParamCount] = cFnGroupScope
		}
		f.Functions[cFnType.Name] = cFnTypeScope
	}
}

func (f *FnScope) NamedElem(name string) NamedElem {
	f.init()
	if p, ok := f.Parameters[name]; ok {
		return p
	}
	if v, ok := f.Variables[name]; ok {
		return v
	}
	if fn, ok := f.Functions[name]; ok {
		return fn
	}
	if f.ContainingScope() != nil {
		return f.ContainingScope().NamedElem(name)
	}
	return nil
}

func (f *FnScope) CallFrom(fc *program.FnCall) NamedElem {
	return &Call{
		Parent:        f.ContainingScope(),
		CalledFnScope: f,
	}
}

func (f *FnScope) Err(loc *program.Location, err error) error {
	return f.PkgFileScope().Err(loc, err)
}

// -------------------------------------------------------------
// BuilderScope implementation
// -------------------------------------------------------------

func (f *FnScope) verifyUndefinedLocally(name string) error {
	if name == program.PVarName || name == program.TimeVarName {
		return fmt.Errorf("%s cannot be reassigned", name)
	}
	if _, isParam := f.Parameters[name]; isParam {
		return fmt.Errorf("%q is an existing parameter", name)
	}
	if _, isVar := f.Variables[name]; isVar {
		return fmt.Errorf("%q is an existing variable", name)
	}
	if _, isFunc := f.Functions[name]; isFunc {
		return fmt.Errorf("%q is an existing function", name)
	}
	return nil
}

func (f *FnScope) AddAssignment(names []string, expr *program.Operand, start, end *program.Location, rng *program.Range) error {
	f.init()
	for _, name := range names {
		if err := f.verifyUndefinedLocally(name); err != nil {
			return err
		}
	}

	newAssignment := program.NewAssignment(names, expr, start, end)
	newAssignment.Range = rng
	f.Function.Assignments = append(f.Function.Assignments, newAssignment)

	assignScope := &AssignScope{Parent: f, Assign: newAssignment}
	for _, v := range newAssignment.Variables {
		f.Variables[v.Name] = &Var{Parent: assignScope, Var: v}
	}
	return nil
}

func (f *FnScope) SetParam(index int, start *program.Location, rng *program.Range) {
	if f.Function != nil && index >= 0 && index < len(f.Function.Parameters) {
		p := f.Function.Parameters[index]
		p.Start = start
		p.Range = rng
	}
}

func (f *FnScope) SetParamKind(index int, kind, anchor string) {
	if f.Function != nil && index >= 0 && index < len(f.Function.Parameters) {
		p := f.Function.Parameters[index]
		p.Kind = kind
		p.Anchor = anchor
	}
}

func (f *FnScope) AddPropertyAssignment(varName, propName string, expr *program.Operand, start, end *program.Location) error {
	f.init()
	variable, ok := f.Variables[varName]
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

func (f *FnScope) AddAnnotation(kind, symbol string, start, end *program.Location) error {
	return fmt.Errorf("@%s can only annotate a top-level declaration", kind)
}

func (f *FnScope) AddImport(imp *program.Import) error {
	return fmt.Errorf("cannot add import in function scope")
}

func (f *FnScope) AddMeshImport(meshImport *program.MeshImport) error {
	return fmt.Errorf("cannot add mesh import in function scope")
}

func (f *FnScope) AddStepImport(stepImport *program.StepImport) error {
	return fmt.Errorf("cannot add step import in function scope")
}

func (f *FnScope) AddWarpImport(warpImport *program.WarpImport) error {
	return fmt.Errorf("cannot add warp import in function scope")
}

func (f *FnScope) IncExprDepth() error {
	return f.PkgFileScope().IncExprDepth()
}

func (f *FnScope) DecExprDepth() {
	f.PkgFileScope().DecExprDepth()
}

func (f *FnScope) NewFnScope(fnName string, params []string) (BuilderScope, error) {
	f.init()
	fnTypes, fnScope, err := f.addFunction(f.Function.FnTypes, fnName, params)
	if err != nil {
		return nil, err
	}
	f.Function.FnTypes = fnTypes
	return fnScope, nil
}

func (f *FnScope) SetReturnOperands(operands []*program.Operand) {
	f.Function.Returns = operands
}

func (f *FnScope) addFunction(
	fnTypes []*program.FnType, fnName string, paramNames []string,
) ([]*program.FnType, *FnScope, error) {
	seen := make(map[string]bool)
	for _, paramName := range paramNames {
		if seen[paramName] {
			return fnTypes, nil, fmt.Errorf("duplicate parameter name %q", paramName)
		}
		seen[paramName] = true
	}

	fnTypeScope, ok := f.Functions[fnName]
	if !ok {
		if err := f.verifyUndefinedLocally(fnName); err != nil {
			return fnTypes, nil, err
		}
		fnTypeScope = &FnTypeScope{
			parent: f,
			FnType: &program.FnType{Name: fnName},
			Groups: make(map[int32]*FnGroupScope),
		}
		f.Functions[fnName] = fnTypeScope
	}

	paramCount := len(paramNames)
	if _, ok := fnTypeScope.Groups[int32(paramCount)]; ok {
		return fnTypes, nil, fmt.Errorf("function %q with %d parameters is already defined", fnName, paramCount)
	}

	fnTypes, fn := program.AddFunction(fnTypes, fnName, paramNames)
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

	return fnTypes, fnScope, nil
}
