package deadcode

import (
	"errors"
	"fmt"

	"gradientcontrol.com/omega/core/internal/scope"
	"gradientcontrol.com/omega/core/program"
)

// Remove takes a linked program and removes all dead code,
// returning a new program.
func Remove(origProg *program.Program) (*program.Program, error) {
	progScope, err := scope.Init(origProg)
	if err != nil {
		return nil, err
	}

	newProg := &program.Program{
		VersionMajor: origProg.VersionMajor,
		VersionMinor: origProg.VersionMinor,
	}

	var outOfScope []scope.NamedElem

	// Resolve/remove dead code in the primary package
	newProg.Primary, outOfScope, err = removePkg(
		progScope.Primary,
		progScope.Primary.Files[len(progScope.Primary.Files)-1].File.Returns,
		outOfScope,
	)
	if err != nil {
		return nil, err
	}

	// Fixed-point iteration to collect all reachable symbols across all dependencies.
	// Since dependencies can have cyclic references (e.g. builtin referencing vector
	// and vector referencing builtin), we must accumulate all referenced symbols
	// until no new references are discovered.
	allReferenced := make(map[any]scope.NamedElem)
	for _, el := range outOfScope {
		allReferenced[uniqueElemKey(el)] = el
	}

	for {
		oldSize := len(allReferenced)
		currList := make([]scope.NamedElem, 0, len(allReferenced))
		for _, el := range allReferenced {
			currList = append(currList, el)
		}

		for _, depPkg := range origProg.Dependencies {
			var filteredList []scope.NamedElem
			for _, el := range currList {
				pPath, _ := scope.ResolvePaths(progScope.Primary, el)
				if pPath == depPkg.Path {
					filteredList = append(filteredList, el)
				}
			}

			var newOutOfScope []scope.NamedElem
			_, newOutOfScope, err = removePkg(
				progScope.Dependencies[depPkg.Path],
				nil,
				filteredList,
			)
			if err != nil {
				return nil, err
			}
			for _, el := range newOutOfScope {
				allReferenced[uniqueElemKey(el)] = el
			}
		}

		if len(allReferenced) == oldSize {
			break
		}
	}

	// Do the final pass to construct the new program's dependencies using the
	// complete set of reachable symbols.
	finalList := make([]scope.NamedElem, 0, len(allReferenced))
	for _, el := range allReferenced {
		finalList = append(finalList, el)
	}

	for _, depPkg := range origProg.Dependencies {
		var filteredList []scope.NamedElem
		for _, el := range finalList {
			pPath, _ := scope.ResolvePaths(progScope.Primary, el)
			if pPath == depPkg.Path {
				filteredList = append(filteredList, el)
			}
		}

		var newPkg *program.Package
		newPkg, _, err = removePkg(
			progScope.Dependencies[depPkg.Path],
			nil,
			filteredList,
		)
		if err != nil {
			return nil, err
		}
		if newPkg != nil {
			newProg.Dependencies = append(newProg.Dependencies, newPkg)
		}
	}

	return newProg, nil
}

func removePkg(pkgScope *scope.PkgScope, returnExpr []*program.Operand, outOfScope []scope.NamedElem) (*program.Package, []scope.NamedElem, error) {
	resolvedFns := make(map[*scope.FnScope]bool)
	fnCache := make(map[*scope.FnScope]fnResult)
	files := make([]*program.PackageFile, 0, len(pkgScope.Files))

	// Resolve each file in the primary package in reverse file order
	for fileIndex := len(pkgScope.Files) - 1; fileIndex >= 0; fileIndex-- {
		fileScope := pkgScope.Files[fileIndex]

		var newFile *program.PackageFile
		var err error
		newFile, outOfScope, err = removeFile(fileScope, returnExpr, outOfScope, resolvedFns, fnCache)
		returnExpr = nil

		if err != nil {
			return nil, outOfScope, err
		}
		files = append([]*program.PackageFile{newFile}, files...)
	}

	return &program.Package{
		VersionMajor: pkgScope.Package.VersionMajor,
		VersionMinor: pkgScope.Package.VersionMinor,
		Path:         pkgScope.Package.Path,
		Files:        files,
		Scene:        program.CloneScene(pkgScope.Package.Scene),
	}, outOfScope, nil
}

func removeFile(fileScope *scope.PkgFileScope, returnExprs []*program.Operand, outOfScope []scope.NamedElem, resolvedFns map[*scope.FnScope]bool, fnCache map[*scope.FnScope]fnResult) (*program.PackageFile, []scope.NamedElem, error) {
	linker := NewRemoveVisitor(fileScope, fileScope.File.Assignments)
	var err error

	if returnExprs != nil {
		roots := append([]*program.Operand{}, returnExprs...)
		roots = append(roots, fileScope.File.ColorReturns...)
		roots = append(roots, fileScope.File.MaterialReturns...)
		outOfScope, err = linker.ResolveExpr(roots...)
	} else {
		outOfScope, err = linker.ResolveNamed(outOfScope...)
	}
	if err != nil {
		return nil, outOfScope, err
	}

	newFnTypes := emptyFnTypes(fileScope.File.FnTypes)

	outOfScope, err = removeNested(
		linker,
		fileScope, outOfScope,
		func(fnName string, paramCount int32) bool {
			return hasFn(newFnTypes, fnName, paramCount)
		},
		func(fnName string, fn *program.Function) {
			addFn(newFnTypes, fn, fnName)
		},
		resolvedFns,
		fnCache,
	)
	if err != nil {
		return nil, outOfScope, err
	}

	outOfScope, err = linker.ResolveNamed(outOfScope...)
	if err != nil {
		return nil, outOfScope, err
	}

	newFnTypes = removeEmptyFnTypes(newFnTypes)

	newFile := &program.PackageFile{
		Path:            fileScope.File.Path,
		Source:          fileScope.File.Source,
		Imports:         fileScope.File.Imports,
		Assignments:     linker.ResolvedAssignments(),
		FnTypes:         newFnTypes,
		Returns:         returnExprs,
		MeshImports:     fileScope.File.MeshImports,
		StepImports:     fileScope.File.StepImports,
		WarpImports:     fileScope.File.WarpImports,
		ColorReturns:    fileScope.File.ColorReturns,
		MaterialReturns: fileScope.File.MaterialReturns,
		// Annotations describe the source, not the reachable code, so they
		// survive deadcoding: a material browser wants every library material,
		// not only the ones the document being assembled happens to use.
		Annotations: fileScope.File.Annotations,
	}

	return newFile, outOfScope, nil
}

func removeFn(fnScope *scope.FnScope, resolvedFns map[*scope.FnScope]bool, fnCache map[*scope.FnScope]fnResult) (*program.Function, []scope.NamedElem, error) {
	if res, ok := fnCache[fnScope]; ok {
		return res.fn, res.outOfScope, res.err
	}

	returnExprs := fnScope.Function.Returns

	linker := NewRemoveVisitor(fnScope, fnScope.Function.Assignments)
	outOfScope, err := linker.ResolveExpr(returnExprs...)
	if err != nil {
		fnCache[fnScope] = fnResult{fn: nil, outOfScope: outOfScope, err: err}
		return nil, outOfScope, err
	}

	newFnTypes := emptyFnTypes(fnScope.Function.FnTypes)

	outOfScope, err = removeNested(
		linker,
		fnScope, outOfScope,
		func(fnName string, paramCount int32) bool {
			return hasFn(newFnTypes, fnName, paramCount)
		},
		func(fnName string, fn *program.Function) {
			addFn(newFnTypes, fn, fnName)
		},
		resolvedFns,
		fnCache,
	)
	if err != nil {
		fnCache[fnScope] = fnResult{fn: nil, outOfScope: outOfScope, err: err}
		return nil, outOfScope, err
	}

	outOfScope, err = linker.ResolveNamed(outOfScope...)
	if err != nil {
		fnCache[fnScope] = fnResult{fn: nil, outOfScope: outOfScope, err: err}
		return nil, outOfScope, err
	}

	newFnTypes = removeEmptyFnTypes(newFnTypes)

	newFn := &program.Function{
		Parameters:  fnScope.Function.Parameters,
		Assignments: linker.ResolvedAssignments(),
		FnTypes:     newFnTypes,
		Returns:     returnExprs,
	}

	fnCache[fnScope] = fnResult{fn: newFn, outOfScope: outOfScope, err: nil}
	return newFn, outOfScope, nil
}

func removeNested(
	linker *RemoveVisitor,
	currScope scope.Scope,
	outOfScope []scope.NamedElem,
	hasFn func(fnName string, paramCount int32) bool,
	addFn func(fnName string, fn *program.Function),
	resolvedFns map[*scope.FnScope]bool,
	fnCache map[*scope.FnScope]fnResult,
) ([]scope.NamedElem, error) {
	outOfScope = dedup(outOfScope)
	seen := make(map[any]bool, len(outOfScope))
	for _, el := range outOfScope {
		seen[uniqueElemKey(el)] = true
	}

	index := 0
	for index < len(outOfScope) {
		elem := outOfScope[index]

		if elem.ContainingScope() == currScope {
			switch elem := elem.(type) {
			case *scope.Var:
				outOfScope = append(outOfScope[:index], outOfScope[index+1:]...)
				delete(seen, uniqueElemKey(elem))

				moreOutOfScope, err := linker.ResolveNamed(elem)
				for _, el := range moreOutOfScope {
					key := uniqueElemKey(el)
					if !seen[key] {
						seen[key] = true
						outOfScope = append(outOfScope, el)
					}
				}
				if err != nil {
					return outOfScope, err
				}
				continue

			case *scope.Call:
				outOfScope = append(outOfScope[:index], outOfScope[index+1:]...)
				delete(seen, uniqueElemKey(elem))

				fnName := elem.Name()
				fnScope := elem.CalledFnScope
				paramCount := int32(len(fnScope.Function.Parameters))
				if hasFn(fnName, paramCount) {
					continue
				}
				if resolvedFns[fnScope] {
					continue
				}
				resolvedFns[fnScope] = true

				linkedFn, moreOutOfScope, err := removeFn(fnScope, resolvedFns, fnCache)
				for _, el := range moreOutOfScope {
					key := uniqueElemKey(el)
					if !seen[key] {
						seen[key] = true
						outOfScope = append(outOfScope, el)
					}
				}
				if err != nil {
					return outOfScope, err
				}

				addFn(fnName, linkedFn)
				continue

			case *scope.FnTypeScope:
				outOfScope = append(outOfScope[:index], outOfScope[index+1:]...)
				delete(seen, uniqueElemKey(elem))

				for _, fnGroupScope := range elem.Groups {
					for _, fnScope := range fnGroupScope.Functions {
						fnName := elem.FnType.Name
						if hasFn(fnName, fnGroupScope.FnGroup.ParamCount) {
							continue
						}
						if resolvedFns[fnScope] {
							continue
						}
						resolvedFns[fnScope] = true

						linkedFn, moreOutOfScope, err := removeFn(fnScope, resolvedFns, fnCache)
						for _, el := range moreOutOfScope {
							key := uniqueElemKey(el)
							if !seen[key] {
								seen[key] = true
								outOfScope = append(outOfScope, el)
							}
						}
						if err != nil {
							return outOfScope, err
						}

						addFn(fnName, linkedFn)
					}
				}
				continue
			}
		}
		index = index + 1
	}
	return outOfScope, nil
}

var errIsPropRef = errors.New("is property reference")

type RemoveVisitor struct {
	currScope   scope.Scope
	assignments []*program.Assignment
	todo        []*program.Operand
	refCounts   map[string]int
	outOfScope  []scope.NamedElem
}

func NewRemoveVisitor(currScope scope.Scope, assignments []*program.Assignment) *RemoveVisitor {
	return &RemoveVisitor{
		currScope:   currScope,
		assignments: assignments,
		refCounts:   make(map[string]int),
		outOfScope:  make([]scope.NamedElem, 0, 20),
	}
}

func (v *RemoveVisitor) ResolveNamed(elems ...scope.NamedElem) ([]scope.NamedElem, error) {
	elems = dedup(elems)
	var todo []*program.Operand
	v.outOfScope = make([]scope.NamedElem, 0, len(elems))

	for _, elem := range elems {
		if elem.ContainingScope() == v.currScope {
			switch e := elem.(type) {
			case *scope.Var:
				if v.refCounts[e.Var.Name] == 0 {
					assign := e.Parent.Assign
					todo = append(todo, assign.Expression)
					for _, assignedVar := range assign.Variables {
						v.refCounts[assignedVar.Name]++
						for _, prop := range assignedVar.Properties {
							todo = append(todo, prop.Expression)
						}
					}
				}
				continue
			case *scope.Param:
				continue
			}
		}
		v.outOfScope = append(v.outOfScope, elem)
	}

	return v.ResolveExpr(todo...)
}

func (v *RemoveVisitor) ResolveExpr(expressions ...*program.Operand) ([]scope.NamedElem, error) {
	for {
		for _, expr := range expressions {
			_, err := program.VisitOperand(expr, v)
			if err != nil {
				return dedup(v.outOfScope), err
			}
		}

		expressions = v.todo
		v.todo = nil

		if len(expressions) == 0 {
			break
		}
	}
	return dedup(v.outOfScope), nil
}

func (v *RemoveVisitor) ResolvedAssignments() []*program.Assignment {
	resolved := make([]*program.Assignment, 0, len(v.assignments))
	for _, assign := range v.assignments {
		for _, assignedVar := range assign.Variables {
			if v.refCounts[assignedVar.Name] > 0 {
				resolved = append(resolved, assign)
				break
			}
		}
	}
	return resolved
}

var _ program.OperandVisitor[any] = (*RemoveVisitor)(nil)

func (v *RemoveVisitor) VisitValue(op *program.Operand, value float64) (any, error) {
	return nil, nil
}

func (v *RemoveVisitor) VisitReference(op *program.Operand, ref *program.Reference) (any, error) {
	elem, err := v.resolveRef(op, ref)
	if err == errIsPropRef {
		prefix := ref.Prefix
		name := ref.Name

		propRef := &program.PropertyRef{
			Reference: &program.Reference{
				Prefix: "",
				Name:   prefix,
			},
			Property: name,
		}

		op.T = &program.Operand_PropRef{PropRef: propRef}
		_, err = program.VisitOperand(op, v)
		if err != nil {
			return nil, err
		}

		baseElem, _ := v.resolveRef(op, propRef.Reference)
		if baseElem != nil {
			return nil, v.currScope.Err(op.Start, fmt.Errorf("property references are not supported yet"))
		}
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if elem == nil {
		return nil, nil
	}

	pkgPath, refPath := scope.ResolvePaths(v.currScope, elem)
	ref.PkgPath = pkgPath
	ref.RefPath = refPath

	switch elem := elem.(type) {
	case *scope.Var:
		if elem.ContainingScope() == v.currScope {
			varName := elem.Name()
			if v.refCounts[varName] == 0 {
				v.refCounts[varName]++
				assign := elem.Parent.Assign
				v.todo = append(v.todo, assign.Expression)
				for _, assignedVar := range assign.Variables {
					v.refCounts[assignedVar.Name]++
					for _, prop := range assignedVar.Properties {
						v.todo = append(v.todo, prop.Expression)
					}
				}
			} else {
				v.refCounts[varName]++
			}
			return nil, nil
		}
	case *scope.Param:
		return nil, nil
	}
	v.outOfScope = append(v.outOfScope, elem)

	return nil, nil
}

func (v *RemoveVisitor) VisitExpression(op *program.Operand, expr *program.Expression) (any, error) {
	if expr.Operation == program.Operation_ACCESS {
		return program.VisitOperand(expr.Operands[0], v)
	}
	v.todo = append(v.todo, expr.Operands...)
	return nil, nil
}

func (v *RemoveVisitor) VisitFnCall(op *program.Operand, fc *program.FnCall) (any, error) {
	elem, err := v.resolveRef(op, fc.Reference)
	if err != nil {
		return nil, err
	}
	v.todo = append(v.todo, fc.Arguments...)
	if elem == nil {
		return nil, nil
	}

	pkgPath, refPath := scope.ResolvePaths(v.currScope, elem)
	fc.Reference.PkgPath = pkgPath
	fc.Reference.RefPath = refPath

	tScope, isFnType := elem.(*scope.FnTypeScope)
	if isFnType {
		gScope := tScope.Groups[int32(len(fc.Arguments))]
		if gScope != nil && len(gScope.Functions) > 0 {
			fnScope := gScope.Functions[0]
			elem = fnScope.CallFrom(fc)
			v.outOfScope = append(v.outOfScope, elem)
			return nil, nil
		}
	}

	v.outOfScope = append(v.outOfScope, elem)
	return nil, nil
}

func (v *RemoveVisitor) VisitPropRef(op *program.Operand, pr *program.PropertyRef) (any, error) {
	elem, err := v.resolveRef(op, pr.Reference)
	if err != nil {
		return nil, err
	}
	if elem == nil {
		return nil, nil
	}

	pkgPath, refPath := scope.ResolvePaths(v.currScope, elem)
	pr.Reference.PkgPath = pkgPath
	pr.Reference.RefPath = refPath

	if elem.ContainingScope() == v.currScope {
		v.refCounts[elem.Name()]++
	} else {
		v.outOfScope = append(v.outOfScope, elem)
	}

	return nil, v.currScope.Err(op.Start, fmt.Errorf("property references are not supported yet"))
}

func (v *RemoveVisitor) resolveRef(op *program.Operand, ref *program.Reference) (scope.NamedElem, error) {
	if ref.Prefix != "" {
		if ref.Prefix == program.PVarName || ref.Prefix == program.TimeVarName {
			return nil, errIsPropRef
		}

		prefElem := v.currScope.NamedElem(ref.Prefix)
		if prefElem == nil {
			return nil, v.currScope.Err(op.Start, fmt.Errorf("undefined identifier: %s", ref.Prefix))
		}
		pkgRef, isPkgRef := prefElem.(*scope.PkgRef)
		if !isPkgRef {
			return nil, errIsPropRef
		}

		elem := pkgRef.NamedElem(ref.Name)
		if elem == nil {
			return nil, v.currScope.Err(op.Start, fmt.Errorf("name %q not found in package %q", ref.Name, pkgRef.Name()))
		}
		return elem, nil
	} else {
		if program.IsBuiltin(ref.Name) || ref.Name == program.PVarName || ref.Name == program.TimeVarName {
			return nil, nil
		}
	}

	elem := v.currScope.NamedElem(ref.Name)
	if elem == nil {
		return nil, v.currScope.Err(op.Start, fmt.Errorf("undefined identifier: %s", ref.Name))
	}
	return elem, nil
}

func addFn(fnTypes []*program.FnType, nestedFn *program.Function, fnName string) {
	paramCount := int32(len(nestedFn.Parameters))
	for _, fnType := range fnTypes {
		if fnType.Name == fnName {
			for _, fnGroup := range fnType.Groups {
				if fnGroup.ParamCount == paramCount {
					fnGroup.Functions = append(fnGroup.Functions, nestedFn)
					return
				}
			}
		}
	}
}

func hasFn(fnTypes []*program.FnType, fnName string, paramCount int32) bool {
	for _, fnType := range fnTypes {
		if fnType.Name == fnName {
			for _, fnGroup := range fnType.Groups {
				if fnGroup.ParamCount == paramCount {
					return len(fnGroup.Functions) > 0
				}
			}
		}
	}
	return false
}

func emptyFnTypes(origFnTypes []*program.FnType) []*program.FnType {
	newFnTypes := make([]*program.FnType, 0, len(origFnTypes))
	for _, orig := range origFnTypes {
		fnType := &program.FnType{Name: orig.Name}
		for _, origGroup := range orig.Groups {
			fnGroup := &program.FnGroup{ParamCount: origGroup.ParamCount}
			fnType.Groups = append(fnType.Groups, fnGroup)
		}
		newFnTypes = append(newFnTypes, fnType)
	}
	return newFnTypes
}

func removeEmptyFnTypes(fnTypes []*program.FnType) []*program.FnType {
	for index := len(fnTypes) - 1; index >= 0; index-- {
		fnTypes[index].Groups = removeEmptyFnGroups(fnTypes[index].Groups)
		if len(fnTypes[index].Groups) == 0 {
			fnTypes = append(fnTypes[:index], fnTypes[index+1:]...)
		}
	}
	return fnTypes
}

func removeEmptyFnGroups(fnGroups []*program.FnGroup) []*program.FnGroup {
	for index := len(fnGroups) - 1; index >= 0; index-- {
		if len(fnGroups[index].Functions) == 0 {
			fnGroups = append(fnGroups[:index], fnGroups[index+1:]...)
		}
	}
	return fnGroups
}

func dedup(elems []scope.NamedElem) []scope.NamedElem {
	seen := make(map[any]bool, len(elems))
	unique := make([]scope.NamedElem, 0, len(elems))
	for _, el := range elems {
		key := uniqueElemKey(el)
		if !seen[key] {
			seen[key] = true
			unique = append(unique, el)
		}
	}
	return unique
}

func uniqueElemKey(el scope.NamedElem) any {
	if c, ok := el.(*scope.Call); ok {
		return c.CalledFnScope
	}
	return el
}

type fnResult struct {
	fn         *program.Function
	outOfScope []scope.NamedElem
	err        error
}
