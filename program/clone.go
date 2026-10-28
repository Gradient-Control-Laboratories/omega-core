package program

// CloneLocation returns a copy of Location.
func CloneLocation(loc *Location) *Location {
	if loc == nil {
		return nil
	}
	return &Location{
		Line:   loc.Line,
		Column: loc.Column,
		Offset: loc.Offset,
	}
}

// CloneReference returns a copy of Reference.
func CloneReference(ref *Reference) *Reference {
	if ref == nil {
		return nil
	}
	return &Reference{
		Prefix:  ref.Prefix,
		Name:    ref.Name,
		PkgPath: ref.PkgPath,
		RefPath: ref.RefPath,
	}
}

// ClonePropertyRef returns a copy of PropertyRef.
func ClonePropertyRef(pr *PropertyRef) *PropertyRef {
	if pr == nil {
		return nil
	}
	return &PropertyRef{
		Reference: CloneReference(pr.Reference),
		Property:  pr.Property,
	}
}

// CloneType returns a copy of Type.
func CloneType(t *Type) *Type {
	if t == nil {
		return nil
	}
	return &Type{
		Kind:       t.Kind,
		Size:       t.Size,
		ElemType:   CloneType(t.ElemType),
		PkgPath:    t.PkgPath,
		FnPath:     t.FnPath,
		ConstValue: CloneOperand(t.ConstValue),
	}
}

// CloneExpression returns a deep copy of Expression.
func CloneExpression(expr *Expression) *Expression {
	if expr == nil {
		return nil
	}
	var operands []*Operand
	if len(expr.Operands) > 0 {
		operands = make([]*Operand, len(expr.Operands))
		for i, op := range expr.Operands {
			operands[i] = CloneOperand(op)
		}
	}
	return &Expression{
		Operation: expr.Operation,
		Operands:  operands,
	}
}

// CloneFnCall returns a deep copy of FnCall.
func CloneFnCall(fc *FnCall) *FnCall {
	if fc == nil {
		return nil
	}
	var args []*Operand
	if len(fc.Arguments) > 0 {
		args = make([]*Operand, len(fc.Arguments))
		for i, arg := range fc.Arguments {
			args[i] = CloneOperand(arg)
		}
	}
	return &FnCall{
		Reference: CloneReference(fc.Reference),
		Arguments: args,
	}
}

// CloneOperand returns a recursive deep copy of Operand, cloning expressions,
// function calls, references, type annotations, and locations.
func CloneOperand(op *Operand) *Operand {
	if op == nil {
		return nil
	}

	cloned := &Operand{
		Start: CloneLocation(op.Start),
		End:   CloneLocation(op.End),
		Type:  CloneType(op.Type),
	}

	if len(op.ConstValues) > 0 {
		cloned.ConstValues = make([]float64, len(op.ConstValues))
		copy(cloned.ConstValues, op.ConstValues)
	}

	switch t := op.T.(type) {
	case *Operand_Value:
		cloned.T = &Operand_Value{Value: t.Value}
	case *Operand_Reference:
		cloned.T = &Operand_Reference{Reference: CloneReference(t.Reference)}
	case *Operand_Expression:
		cloned.T = &Operand_Expression{Expression: CloneExpression(t.Expression)}
	case *Operand_FnCall:
		cloned.T = &Operand_FnCall{FnCall: CloneFnCall(t.FnCall)}
	case *Operand_PropRef:
		cloned.T = &Operand_PropRef{PropRef: ClonePropertyRef(t.PropRef)}
	}

	return cloned
}

// CloneRangeAnnotation returns a copy of RangeAnnotation.
func CloneRangeAnnotation(r *Range) *Range {
	if r == nil {
		return nil
	}
	return &Range{
		Start:  CloneLocation(r.Start),
		End:    CloneLocation(r.End),
		Min:    CloneOperand(r.Min),
		Max:    CloneOperand(r.Max),
		Period: CloneOperand(r.Period),
	}
}

// CloneParameter returns a copy of Parameter.
func CloneParameter(param *Parameter) *Parameter {
	if param == nil {
		return nil
	}
	return &Parameter{
		Start:  CloneLocation(param.Start),
		Name:   param.Name,
		Type:   CloneType(param.Type),
		Range:  CloneRangeAnnotation(param.Range),
		Kind:   param.Kind,
		Anchor: param.Anchor,
	}
}

// CloneProperty returns a copy of Property.
func CloneProperty(prop *Property) *Property {
	if prop == nil {
		return nil
	}
	return &Property{
		Start:      CloneLocation(prop.Start),
		End:        CloneLocation(prop.End),
		Name:       prop.Name,
		Expression: CloneOperand(prop.Expression),
	}
}

// CloneVariable returns a copy of Variable.
func CloneVariable(v *Variable) *Variable {
	if v == nil {
		return nil
	}
	var props []*Property
	if len(v.Properties) > 0 {
		props = make([]*Property, len(v.Properties))
		for i, prop := range v.Properties {
			props[i] = CloneProperty(prop)
		}
	}
	var constVals []float64
	if len(v.ConstValues) > 0 {
		constVals = make([]float64, len(v.ConstValues))
		copy(constVals, v.ConstValues)
	}
	return &Variable{
		Name:        v.Name,
		Properties:  props,
		Type:        CloneType(v.Type),
		ConstValues: constVals,
	}
}

// CloneAssignment returns a copy of Assignment.
func CloneAssignment(assign *Assignment) *Assignment {
	if assign == nil {
		return nil
	}
	var vars []*Variable
	if len(assign.Variables) > 0 {
		vars = make([]*Variable, len(assign.Variables))
		for i, v := range assign.Variables {
			vars[i] = CloneVariable(v)
		}
	}
	return &Assignment{
		Start:      CloneLocation(assign.Start),
		End:        CloneLocation(assign.End),
		Variables:  vars,
		Expression: CloneOperand(assign.Expression),
		Range:      CloneRangeAnnotation(assign.Range),
	}
}

// CloneFunction returns a deep copy of Function.
func CloneFunction(fn *Function) *Function {
	if fn == nil {
		return nil
	}
	var params []*Parameter
	if len(fn.Parameters) > 0 {
		params = make([]*Parameter, len(fn.Parameters))
		for i, p := range fn.Parameters {
			params[i] = CloneParameter(p)
		}
	}
	var assigns []*Assignment
	if len(fn.Assignments) > 0 {
		assigns = make([]*Assignment, len(fn.Assignments))
		for i, a := range fn.Assignments {
			assigns[i] = CloneAssignment(a)
		}
	}
	var fnTypes []*FnType
	if len(fn.FnTypes) > 0 {
		fnTypes = make([]*FnType, len(fn.FnTypes))
		for i, ft := range fn.FnTypes {
			fnTypes[i] = CloneFnType(ft)
		}
	}
	var returns []*Operand
	if len(fn.Returns) > 0 {
		returns = make([]*Operand, len(fn.Returns))
		for i, ret := range fn.Returns {
			returns[i] = CloneOperand(ret)
		}
	}
	return &Function{
		Parameters:  params,
		Assignments: assigns,
		FnTypes:     fnTypes,
		Returns:     returns,
	}
}

// CloneFnGroup returns a deep copy of FnGroup.
func CloneFnGroup(fg *FnGroup) *FnGroup {
	if fg == nil {
		return nil
	}
	var functions []*Function
	if len(fg.Functions) > 0 {
		functions = make([]*Function, len(fg.Functions))
		for i, fn := range fg.Functions {
			functions[i] = CloneFunction(fn)
		}
	}
	return &FnGroup{
		ParamCount: fg.ParamCount,
		Functions:  functions,
	}
}

// CloneFnType returns a deep copy of FnType.
func CloneFnType(ft *FnType) *FnType {
	if ft == nil {
		return nil
	}
	var groups []*FnGroup
	if len(ft.Groups) > 0 {
		groups = make([]*FnGroup, len(ft.Groups))
		for i, g := range ft.Groups {
			groups[i] = CloneFnGroup(g)
		}
	}
	return &FnType{
		Name:   ft.Name,
		Groups: groups,
	}
}

// CloneScene returns a deep copy of Scene.
func CloneScene(s *Scene) *Scene {
	if s == nil {
		return nil
	}
	return &Scene{
		Start:           CloneLocation(s.Start),
		End:             CloneLocation(s.End),
		Size:            s.Size,
		Grid:            s.Grid,
		SpinodoidBounds: append([]float64(nil), s.SpinodoidBounds...),
	}
}

// CloneImport returns a copy of Import.
func CloneImport(imp *Import) *Import {
	if imp == nil {
		return nil
	}
	return &Import{
		Start:  CloneLocation(imp.Start),
		End:    CloneLocation(imp.End),
		Prefix: imp.Prefix,
		Path:   imp.Path,
	}
}

// ClonePackageFile returns a deep copy of PackageFile.
func ClonePackageFile(file *PackageFile) *PackageFile {
	if file == nil {
		return nil
	}
	var imports []*Import
	if len(file.Imports) > 0 {
		imports = make([]*Import, len(file.Imports))
		for i, imp := range file.Imports {
			imports[i] = CloneImport(imp)
		}
	}
	var assigns []*Assignment
	if len(file.Assignments) > 0 {
		assigns = make([]*Assignment, len(file.Assignments))
		for i, a := range file.Assignments {
			assigns[i] = CloneAssignment(a)
		}
	}
	var fnTypes []*FnType
	if len(file.FnTypes) > 0 {
		fnTypes = make([]*FnType, len(file.FnTypes))
		for i, ft := range file.FnTypes {
			fnTypes[i] = CloneFnType(ft)
		}
	}
	var returns []*Operand
	if len(file.Returns) > 0 {
		returns = make([]*Operand, len(file.Returns))
		for i, ret := range file.Returns {
			returns[i] = CloneOperand(ret)
		}
	}
	var colorReturns []*Operand
	if len(file.ColorReturns) > 0 {
		colorReturns = make([]*Operand, len(file.ColorReturns))
		for i, ret := range file.ColorReturns {
			colorReturns[i] = CloneOperand(ret)
		}
	}
	var materialReturns []*Operand
	if len(file.MaterialReturns) > 0 {
		materialReturns = make([]*Operand, len(file.MaterialReturns))
		for i, ret := range file.MaterialReturns {
			materialReturns[i] = CloneOperand(ret)
		}
	}
	return &PackageFile{
		Path:            file.Path,
		Source:          file.Source,
		Imports:         imports,
		Assignments:     assigns,
		FnTypes:         fnTypes,
		Returns:         returns,
		MeshImports:     file.MeshImports,
		StepImports:     file.StepImports,
		WarpImports:     file.WarpImports,
		ColorReturns:    colorReturns,
		MaterialReturns: materialReturns,
		Annotations:     file.Annotations,
	}
}

// ClonePackage returns a deep copy of Package.
func ClonePackage(pkg *Package) *Package {
	if pkg == nil {
		return nil
	}
	var files []*PackageFile
	if len(pkg.Files) > 0 {
		files = make([]*PackageFile, len(pkg.Files))
		for i, f := range pkg.Files {
			files[i] = ClonePackageFile(f)
		}
	}
	return &Package{
		VersionMajor: pkg.VersionMajor,
		VersionMinor: pkg.VersionMinor,
		Path:         pkg.Path,
		Files:        files,
		Scene:        CloneScene(pkg.Scene),
	}
}
