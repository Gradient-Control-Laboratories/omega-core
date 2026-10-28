package inline

import (
	"fmt"
	"strings"

	"gradientcontrol.com/omega/core/internal/scope"
	"gradientcontrol.com/omega/core/program"
	gsllibs "gradientcontrol.com/omega/core/stdlib"
)

func (in *inliner) countParamUses(fn *program.Function, paramName string) int {
	count := 0
	for _, a := range fn.Assignments {
		count += in.countUsesInOperand(a.Expression, paramName)
	}
	for _, r := range fn.Returns {
		count += in.countUsesInOperand(r, paramName)
	}
	return count
}

func (in *inliner) countVarUsesInRemaining(fn *program.Function, fromIndex int, varName string) int {
	count := 0
	for i := fromIndex + 1; i < len(fn.Assignments); i++ {
		count += in.countUsesInOperand(fn.Assignments[i].Expression, varName)
	}
	for _, r := range fn.Returns {
		count += in.countUsesInOperand(r, varName)
	}
	return count
}

func (in *inliner) hoistIfReused(op *program.Operand, fn *program.Function, assignIndex int, varName string, fileScope *scope.PkgFileScope) *program.Operand {
	if op == nil {
		return nil
	}
	if !in.isCopyableConst(op, fileScope) && !in.isSimpleOperand(op) && in.countVarUsesInRemaining(fn, assignIndex, varName) > 1 {
		in.anonCounter++
		tmpVar := fmt.Sprintf("gsl_tmp_%d", in.anonCounter)
		in.hoistedAssignments = append(in.hoistedAssignments, &program.Assignment{
			Start: program.CloneLocation(op.Start),
			End:   program.CloneLocation(op.End),
			Variables: []*program.Variable{
				{Name: tmpVar},
			},
			Expression: program.CloneOperand(op),
		})
		return &program.Operand{
			Start: program.CloneLocation(op.Start),
			End:   program.CloneLocation(op.End),
			T: &program.Operand_Reference{
				Reference: &program.Reference{Name: tmpVar},
			},
		}
	}
	return op
}

func (in *inliner) ensureImport(targetFile *program.PackageFile, calleePkgPath string, fileScope *scope.PkgFileScope) string {
	if targetFile == nil || calleePkgPath == "" {
		return ""
	}

	if in.currentFileScope != nil && in.currentFileScope.PkgScope() != nil && in.currentFileScope.PkgScope().Package != nil {
		if in.currentFileScope.PkgScope().Package.Path == calleePkgPath {
			return ""
		}
	}

	// 1. Check existing imports in targetFile
	for _, imp := range targetFile.Imports {
		if imp.Path == calleePkgPath && imp.PrefixOrSimpleName() != "." {
			pref := imp.PrefixOrSimpleName()
			if !in.isNameShadowed(pref, targetFile, fileScope) {
				return pref
			}
		}
	}

	// 2. Derive base prefix from package path
	pkgParts := strings.Split(calleePkgPath, "/")
	pkgBase := pkgParts[len(pkgParts)-1]

	isPrefixUsed := func(p string) bool {
		for _, imp := range targetFile.Imports {
			if imp.PrefixOrSimpleName() == p {
				return true
			}
		}
		return in.isNameShadowed(p, targetFile, fileScope)
	}

	chosenPrefix := pkgBase
	if isPrefixUsed(chosenPrefix) {
		suffix := 1
		for isPrefixUsed(fmt.Sprintf("%s_%d", pkgBase, suffix)) {
			suffix++
		}
		chosenPrefix = fmt.Sprintf("%s_%d", pkgBase, suffix)
	}

	targetFile.Imports = append(targetFile.Imports, &program.Import{
		Path:   calleePkgPath,
		Prefix: chosenPrefix,
	})
	return chosenPrefix
}

func collectFnPrefixes(fn *program.Function) map[string]bool {
	used := make(map[string]bool)
	var visitOp func(op *program.Operand)
	visitOp = func(op *program.Operand) {
		if op == nil {
			return
		}
		switch t := op.T.(type) {
		case *program.Operand_Reference:
			if t.Reference != nil && t.Reference.Prefix != "" {
				used[t.Reference.Prefix] = true
			}
		case *program.Operand_PropRef:
			if t.PropRef != nil && t.PropRef.Reference != nil && t.PropRef.Reference.Prefix != "" {
				used[t.PropRef.Reference.Prefix] = true
			}
		case *program.Operand_FnCall:
			if t.FnCall != nil {
				if t.FnCall.Reference != nil && t.FnCall.Reference.Prefix != "" {
					used[t.FnCall.Reference.Prefix] = true
				}
				for _, arg := range t.FnCall.Arguments {
					visitOp(arg)
				}
			}
		case *program.Operand_Expression:
			if t.Expression != nil {
				for _, child := range t.Expression.Operands {
					visitOp(child)
				}
			}
		}
	}
	var visitFn func(f *program.Function)
	visitFn = func(f *program.Function) {
		if f == nil {
			return
		}
		for _, a := range f.Assignments {
			if a != nil {
				visitOp(a.Expression)
				for _, v := range a.Variables {
					if v != nil {
						for _, p := range v.Properties {
							if p != nil {
								visitOp(p.Expression)
							}
						}
					}
				}
			}
		}
		for _, ret := range f.Returns {
			visitOp(ret)
		}
		for _, ft := range f.FnTypes {
			if ft != nil {
				for _, g := range ft.Groups {
					if g != nil {
						for _, childFn := range g.Functions {
							visitFn(childFn)
						}
					}
				}
			}
		}
	}
	visitFn(fn)
	return used
}

func (in *inliner) buildPrefixMap(calleePkgPath string, targetFile *program.PackageFile, fileScope *scope.PkgFileScope, fn *program.Function) map[string]string {
	if targetFile == nil || in.progScope == nil {
		return nil
	}
	prefixMap := make(map[string]string)
	usedPrefixes := collectFnPrefixes(fn)
	if len(usedPrefixes) == 0 {
		return prefixMap
	}

	var calleePkgScope *scope.PkgScope
	if in.prog.Primary != nil && in.prog.Primary.Path == calleePkgPath {
		calleePkgScope = in.progScope.Primary
	} else if in.progScope.Dependencies != nil {
		calleePkgScope = in.progScope.Dependencies[calleePkgPath]
	}

	if calleePkgScope != nil {
		for _, file := range calleePkgScope.Files {
			for impPrefix, impRef := range file.Imports {
				if usedPrefixes[impPrefix] && impRef != nil && impRef.PkgScope() != nil && impRef.PkgScope().Package != nil {
					depPkgPath := impRef.PkgScope().Package.Path
					targetPref := in.ensureImport(targetFile, depPkgPath, fileScope)
					if targetPref != "" {
						prefixMap[impPrefix] = targetPref
					}
				}
			}
		}
	}

	for pref := range usedPrefixes {
		if sysPath, ok := gsllibs.PathForPrefix(pref); ok {
			if in.isNameShadowed(pref, targetFile, fileScope) {
				targetPref := in.ensureImport(targetFile, sysPath, fileScope)
				if targetPref != "" {
					prefixMap[pref] = targetPref
				}
			}
		}
	}

	return prefixMap
}

func (in *inliner) inlineSingleReturn(
	fc *program.FnCall,
	fn *program.Function,
	fileScope *scope.PkgFileScope,
	targetFile *program.PackageFile,
	currentPkgPath string,
	calleePkgPath string,
) (*program.Operand, error) {
	origHoistedCount := len(in.hoistedAssignments)
	success := false
	defer func() {
		if !success {
			in.hoistedAssignments = in.hoistedAssignments[:origHoistedCount]
		}
	}()

	calleePrefix := ""
	if calleePkgPath != currentPkgPath {
		calleePrefix = in.ensureImport(targetFile, calleePkgPath, fileScope)
	}
	prefixMap := in.buildPrefixMap(calleePkgPath, targetFile, fileScope, fn)

	env := make(map[string]*program.Operand)

	for i, param := range fn.Parameters {
		if i < len(fc.Arguments) {
			arg := fc.Arguments[i]
			if !in.isCopyableConst(arg, fileScope) && !in.isSimpleOperand(arg) && in.countParamUses(fn, param.Name) > 1 {
				in.anonCounter++
				tmpVar := fmt.Sprintf("gsl_tmp_%d", in.anonCounter)
				in.hoistedAssignments = append(in.hoistedAssignments, &program.Assignment{
					Start: program.CloneLocation(arg.Start),
					End:   program.CloneLocation(arg.End),
					Variables: []*program.Variable{
						{Name: tmpVar},
					},
					Expression: program.CloneOperand(arg),
				})
				env[param.Name] = &program.Operand{
					Start: program.CloneLocation(arg.Start),
					End:   program.CloneLocation(arg.End),
					T: &program.Operand_Reference{
						Reference: &program.Reference{Name: tmpVar},
					},
				}
			} else {
				env[param.Name] = in.bindArg(arg)
			}
		}
	}

	anonRenames := make(map[string]string)
	var allNestedFnTypes []*program.FnType
	if targetFile != nil && len(fn.FnTypes) > 0 {
		pkgParts := strings.Split(calleePkgPath, "/")
		pkgBase := pkgParts[len(pkgParts)-1]
		var collectFnTypes func(fts []*program.FnType)
		collectFnTypes = func(fts []*program.FnType) {
			for _, ft := range fts {
				in.anonCounter++
				newName := fmt.Sprintf("%s%s_%s_%d", program.InlinedFnPrefix, pkgBase, ft.Name, in.anonCounter)
				anonRenames[ft.Name] = newName
				allNestedFnTypes = append(allNestedFnTypes, ft)
				for _, grp := range ft.Groups {
					for _, f := range grp.Functions {
						collectFnTypes(f.FnTypes)
					}
				}
			}
		}
		collectFnTypes(fn.FnTypes)
	}

	for assignIndex, assign := range fn.Assignments {
		if len(assign.Variables) > 1 && assign.Expression != nil {
			subAssignExpr := in.substitute(assign.Expression, env, calleePrefix, anonRenames, prefixMap)
			if localFC := subAssignExpr.GetFnCall(); localFC != nil {
				nestedDecl, nestedFn, nestedPkgPath := in.findFunction(localFC, fileScope, len(localFC.Arguments))
				if nestedFn != nil && len(nestedFn.Returns) == len(assign.Variables) {
					tempAssign := &program.Assignment{
						Variables:  assign.Variables,
						Expression: subAssignExpr,
						Start:      assign.Start,
						End:        assign.End,
					}
					nestedExpanded, err := in.inlineMultiReturnAssignment(tempAssign, nestedDecl, nestedFn, nestedPkgPath, fileScope, targetFile, currentPkgPath)
					if err == nil && len(nestedExpanded) == len(assign.Variables) {
						for idx, na := range nestedExpanded {
							env[assign.Variables[idx].Name] = in.hoistIfReused(na.Expression, fn, assignIndex, assign.Variables[idx].Name, fileScope)
						}
						continue
					}
				}
			}
		}

		subExpr := in.substitute(assign.Expression, env, calleePrefix, anonRenames, prefixMap)
		if len(assign.Variables) == 1 {
			env[assign.Variables[0].Name] = in.hoistIfReused(subExpr, fn, assignIndex, assign.Variables[0].Name, fileScope)
		} else if len(assign.Variables) > 1 {
			if exprOp, ok := subExpr.T.(*program.Operand_Expression); ok && exprOp.Expression.Operation == program.Operation_COLLECT {
				for i, v := range assign.Variables {
					if i < len(exprOp.Expression.Operands) {
						env[v.Name] = in.hoistIfReused(exprOp.Expression.Operands[i], fn, assignIndex, v.Name, fileScope)
					}
				}
			}
		}
	}

	for _, assign := range fn.Assignments {
		for _, v := range assign.Variables {
			if _, ok := env[v.Name]; !ok {
				return nil, nil
			}
		}
	}

	if targetFile != nil && len(fn.FnTypes) > 0 {
		for _, ft := range fn.FnTypes {
			if in.hasUnresolvedFreeVars(ft, env, fileScope, calleePkgPath) {
				return nil, nil
			}
		}
		var processFnType func(ft *program.FnType, currentEnv map[string]*program.Operand) *program.FnType
		processFnType = func(ft *program.FnType, currentEnv map[string]*program.Operand) *program.FnType {
			clonedFt := program.CloneFnType(ft)
			if newName, ok := anonRenames[ft.Name]; ok {
				clonedFt.Name = newName
			}
			for _, grp := range clonedFt.Groups {
				for _, f := range grp.Functions {
					childEnv := make(map[string]*program.Operand, len(currentEnv))
					for k, v := range currentEnv {
						childEnv[k] = v
					}
					for _, p := range f.Parameters {
						childEnv[p.Name] = nil
					}
					for _, a := range f.Assignments {
						for _, v := range a.Variables {
							childEnv[v.Name] = nil
						}
					}
					for _, a := range f.Assignments {
						a.Expression = in.substitute(a.Expression, childEnv, calleePrefix, anonRenames, prefixMap)
					}
					for i, ret := range f.Returns {
						f.Returns[i] = in.substitute(ret, childEnv, calleePrefix, anonRenames, prefixMap)
					}
					for j, subFt := range f.FnTypes {
						f.FnTypes[j] = processFnType(subFt, childEnv)
					}
				}
			}
			return clonedFt
		}
		for _, ft := range fn.FnTypes {
			targetFile.FnTypes = append(targetFile.FnTypes, processFnType(ft, env))
		}
	}

	if len(fn.Returns) == 0 {
		return nil, nil
	}

	inlinedRet := in.substitute(fn.Returns[0], env, calleePrefix, anonRenames, prefixMap)
	success = true
	return inlinedRet, nil
}

func (in *inliner) inlineMultiReturnAssignment(
	assign *program.Assignment,
	fnDecl *program.FnType,
	fn *program.Function,
	calleePkgPath string,
	fileScope *scope.PkgFileScope,
	targetFile *program.PackageFile,
	currentPkgPath string,
) ([]*program.Assignment, error) {
	if fn == nil {
		return nil, nil
	}

	fc := assign.Expression.GetFnCall()
	if fc == nil {
		return nil, nil
	}

	origHoistedCount := len(in.hoistedAssignments)
	success := false
	defer func() {
		if !success {
			in.hoistedAssignments = in.hoistedAssignments[:origHoistedCount]
		}
	}()

	calleePrefix := ""
	if calleePkgPath != currentPkgPath {
		calleePrefix = in.ensureImport(targetFile, calleePkgPath, fileScope)
	}
	prefixMap := in.buildPrefixMap(calleePkgPath, targetFile, fileScope, fn)

	callKey := fmt.Sprintf("%s/%s/%d", calleePkgPath, fnDecl.Name, len(fc.Arguments))
	if in.callStack[callKey] {
		return nil, fmt.Errorf("recursion not allowed in GSL: %q is called recursively", fnDecl.Name)
	}

	in.callStack[callKey] = true
	defer func() {
		delete(in.callStack, callKey)
	}()

	env := make(map[string]*program.Operand)
	for i, param := range fn.Parameters {
		if i < len(fc.Arguments) {
			arg := fc.Arguments[i]
			if !in.isCopyableConst(arg, fileScope) && !in.isSimpleOperand(arg) && in.countParamUses(fn, param.Name) > 1 {
				in.anonCounter++
				tmpVar := fmt.Sprintf("gsl_tmp_%d", in.anonCounter)
				in.hoistedAssignments = append(in.hoistedAssignments, &program.Assignment{
					Start: program.CloneLocation(arg.Start),
					End:   program.CloneLocation(arg.End),
					Variables: []*program.Variable{
						{Name: tmpVar},
					},
					Expression: program.CloneOperand(arg),
				})
				env[param.Name] = &program.Operand{
					Start: program.CloneLocation(arg.Start),
					End:   program.CloneLocation(arg.End),
					T: &program.Operand_Reference{
						Reference: &program.Reference{Name: tmpVar},
					},
				}
			} else {
				env[param.Name] = in.bindArg(arg)
			}
		}
	}

	anonRenames := make(map[string]string)
	if targetFile != nil && len(fn.FnTypes) > 0 {
		pkgParts := strings.Split(calleePkgPath, "/")
		pkgBase := pkgParts[len(pkgParts)-1]
		var collectFnTypes func(fts []*program.FnType)
		collectFnTypes = func(fts []*program.FnType) {
			for _, ft := range fts {
				in.anonCounter++
				newName := fmt.Sprintf("%s%s_%s_%d", program.InlinedFnPrefix, pkgBase, ft.Name, in.anonCounter)
				anonRenames[ft.Name] = newName
				for _, grp := range ft.Groups {
					for _, f := range grp.Functions {
						collectFnTypes(f.FnTypes)
					}
				}
			}
		}
		collectFnTypes(fn.FnTypes)
	}

	for assignIndex, localAssign := range fn.Assignments {
		if len(localAssign.Variables) > 1 && localAssign.Expression != nil {
			subAssignExpr := in.substitute(localAssign.Expression, env, calleePrefix, anonRenames, prefixMap)
			if localFC := subAssignExpr.GetFnCall(); localFC != nil {
				nestedDecl, nestedFn, nestedPkgPath := in.findFunction(localFC, fileScope, len(localFC.Arguments))
				if nestedFn != nil && len(nestedFn.Returns) == len(localAssign.Variables) {
					tempAssign := &program.Assignment{
						Variables:  localAssign.Variables,
						Expression: subAssignExpr,
						Start:      localAssign.Start,
						End:        localAssign.End,
					}
					nestedExpanded, err := in.inlineMultiReturnAssignment(tempAssign, nestedDecl, nestedFn, nestedPkgPath, fileScope, targetFile, currentPkgPath)
					if err == nil && len(nestedExpanded) == len(localAssign.Variables) {
						for idx, na := range nestedExpanded {
							env[localAssign.Variables[idx].Name] = in.hoistIfReused(na.Expression, fn, assignIndex, localAssign.Variables[idx].Name, fileScope)
						}
						continue
					}
				}
			}
		}

		subExpr := in.substitute(localAssign.Expression, env, calleePrefix, anonRenames, prefixMap)
		if len(localAssign.Variables) == 1 {
			if localFC := subExpr.GetFnCall(); localFC != nil {
				_, nestedFn, nestedPkgPath := in.findFunction(localFC, fileScope, len(localFC.Arguments))
				if nestedFn != nil && len(nestedFn.Returns) == 1 && in.shouldInlineCall(localFC, nestedFn, nestedPkgPath, currentPkgPath, fileScope) {
					inlinedNested, err := in.inlineSingleReturn(localFC, nestedFn, fileScope, targetFile, currentPkgPath, nestedPkgPath)
					if err == nil && inlinedNested != nil {
						subExpr = inlinedNested
					}
				}
			}
			env[localAssign.Variables[0].Name] = in.hoistIfReused(subExpr, fn, assignIndex, localAssign.Variables[0].Name, fileScope)
		} else if len(localAssign.Variables) > 1 {
			if exprOp, ok := subExpr.T.(*program.Operand_Expression); ok && exprOp.Expression.Operation == program.Operation_COLLECT {
				for i, v := range localAssign.Variables {
					if i < len(exprOp.Expression.Operands) {
						env[v.Name] = in.hoistIfReused(exprOp.Expression.Operands[i], fn, assignIndex, v.Name, fileScope)
					}
				}
			}
		}
	}

	for _, localAssign := range fn.Assignments {
		for _, v := range localAssign.Variables {
			if _, ok := env[v.Name]; !ok {
				return nil, nil
			}
		}
	}

	if targetFile != nil && len(fn.FnTypes) > 0 {
		for _, ft := range fn.FnTypes {
			if in.hasUnresolvedFreeVars(ft, env, fileScope, calleePkgPath) {
				return nil, nil
			}
		}
		var processFnType func(ft *program.FnType, currentEnv map[string]*program.Operand) *program.FnType
		processFnType = func(ft *program.FnType, currentEnv map[string]*program.Operand) *program.FnType {
			clonedFt := program.CloneFnType(ft)
			if newName, ok := anonRenames[ft.Name]; ok {
				clonedFt.Name = newName
			}
			for _, grp := range clonedFt.Groups {
				for _, f := range grp.Functions {
					childEnv := make(map[string]*program.Operand, len(currentEnv))
					for k, v := range currentEnv {
						childEnv[k] = v
					}
					for _, p := range f.Parameters {
						childEnv[p.Name] = nil
					}
					for _, a := range f.Assignments {
						for _, v := range a.Variables {
							childEnv[v.Name] = nil
						}
					}
					for _, a := range f.Assignments {
						a.Expression = in.substitute(a.Expression, childEnv, calleePrefix, anonRenames, prefixMap)
					}
					for i, ret := range f.Returns {
						f.Returns[i] = in.substitute(ret, childEnv, calleePrefix, anonRenames, prefixMap)
					}
					for j, subFt := range f.FnTypes {
						f.FnTypes[j] = processFnType(subFt, childEnv)
					}
				}
			}
			return clonedFt
		}
		for _, ft := range fn.FnTypes {
			targetFile.FnTypes = append(targetFile.FnTypes, processFnType(ft, env))
		}
	}

	result := make([]*program.Assignment, len(assign.Variables))
	for i, v := range assign.Variables {
		var retExpr *program.Operand
		if i < len(fn.Returns) {
			retExpr = in.substitute(fn.Returns[i], env, calleePrefix, anonRenames, prefixMap)
		}
		result[i] = &program.Assignment{
			Variables:  []*program.Variable{v},
			Expression: retExpr,
			Start:      assign.Start,
			End:        assign.End,
		}
	}

	success = true
	return result, nil
}

func (in *inliner) substitute(op *program.Operand, env map[string]*program.Operand, calleePrefix string, anonRenames map[string]string, prefixMap map[string]string) *program.Operand {
	if op == nil {
		return nil
	}

	switch t := op.T.(type) {
	case *program.Operand_Reference:
		if t.Reference.Prefix == "" {
			if env != nil {
				if repl, ok := env[t.Reference.Name]; ok {
					if repl != nil {
						return program.CloneOperand(repl)
					}
					// Local parameter or variable: do NOT qualify with calleePrefix
					return program.CloneOperand(op)
				}
			}
			if anonRenames != nil {
				if newName, ok := anonRenames[t.Reference.Name]; ok {
					cloned := program.CloneOperand(op)
					cloned.GetReference().Name = newName
					cloned.GetReference().Prefix = ""
					if cloned.Type != nil {
						cloned.Type.FnPath = newName
						cloned.Type.PkgPath = ""
					}
					return cloned
				}
			}
			if calleePrefix != "" && !in.isBuiltinOrSpecialName(t.Reference.Name) {
				cloned := program.CloneOperand(op)
				cloned.GetReference().Prefix = calleePrefix
				return cloned
			}
		} else if env != nil && env[t.Reference.Prefix] != nil {
			repl := env[t.Reference.Prefix]
			propIdx := -1
			switch t.Reference.Name {
			case "x", "r":
				propIdx = 0
			case "y", "g":
				propIdx = 1
			case "z", "b":
				propIdx = 2
			case "w", "a":
				propIdx = 3
			}
			if propIdx >= 0 {
				if exprOp, ok := repl.T.(*program.Operand_Expression); ok && exprOp.Expression.Operation == program.Operation_COLLECT {
					if propIdx < len(exprOp.Expression.Operands) {
						return program.CloneOperand(exprOp.Expression.Operands[propIdx])
					}
				}
				if len(repl.ConstValues) > propIdx {
					return &program.Operand{
						Start:       program.CloneLocation(op.Start),
						End:         program.CloneLocation(op.End),
						T:           &program.Operand_Value{Value: repl.ConstValues[propIdx]},
						ConstValues: []float64{repl.ConstValues[propIdx]},
					}
				}
			}
			return &program.Operand{
				Start: program.CloneLocation(op.Start),
				End:   program.CloneLocation(op.End),
				T: &program.Operand_Expression{
					Expression: &program.Expression{
						Operation: program.Operation_ACCESS,
						Operands: []*program.Operand{
							program.CloneOperand(repl),
							{
								Start: program.CloneLocation(op.Start),
								End:   program.CloneLocation(op.End),
								T: &program.Operand_Reference{
									Reference: &program.Reference{Name: t.Reference.Name},
								},
							},
						},
					},
				},
			}
		} else if prefixMap != nil {
			if mapped, ok := prefixMap[t.Reference.Prefix]; ok && mapped != "" {
				cloned := program.CloneOperand(op)
				cloned.GetReference().Prefix = mapped
				return cloned
			}
		}
		return program.CloneOperand(op)

	case *program.Operand_Expression:
		cloned := &program.Operand{
			Start:       program.CloneLocation(op.Start),
			End:         program.CloneLocation(op.End),
			Type:        program.CloneType(op.Type),
			ConstValues: op.ConstValues,
		}
		newOperands := make([]*program.Operand, len(t.Expression.Operands))
		for i, child := range t.Expression.Operands {
			if t.Expression.Operation == program.Operation_ACCESS && i > 0 {
				newOperands[i] = program.CloneOperand(child)
			} else {
				newOperands[i] = in.substitute(child, env, calleePrefix, anonRenames, prefixMap)
			}
		}
		cloned.T = &program.Operand_Expression{
			Expression: &program.Expression{
				Operation: t.Expression.Operation,
				Operands:  newOperands,
			},
		}
		return cloned

	case *program.Operand_FnCall:
		cloned := &program.Operand{
			Start:       program.CloneLocation(op.Start),
			End:         program.CloneLocation(op.End),
			Type:        program.CloneType(op.Type),
			ConstValues: op.ConstValues,
		}
		newArgs := make([]*program.Operand, len(t.FnCall.Arguments))
		for i, arg := range t.FnCall.Arguments {
			newArgs[i] = in.substitute(arg, env, calleePrefix, anonRenames, prefixMap)
		}
		ref := program.CloneReference(t.FnCall.Reference)
		// UI markers edit the file that declared them. Across an import they
		// are identities, not controls at coincident offsets in the caller.
		// Substituted caller arguments keep their own markers and locations.
		if calleePrefix != "" && ref != nil && ref.Prefix == "" && len(newArgs) >= 1 {
			switch ref.Name {
			case "uiPoint", "uiTransform", "uiSpline", "uiColor", "uiDirection":
				return newArgs[0]
			}
		}
		if ref != nil && ref.Prefix == "" {
			if env != nil {
				if _, ok := env[ref.Name]; ok {
					cloned.T = &program.Operand_FnCall{
						FnCall: &program.FnCall{
							Reference: ref,
							Arguments: newArgs,
						},
					}
					return cloned
				}
			}
			if anonRenames != nil {
				if newName, ok := anonRenames[ref.Name]; ok {
					ref.Name = newName
					ref.Prefix = ""
					if cloned.Type != nil {
						cloned.Type.FnPath = newName
						cloned.Type.PkgPath = ""
					}
				} else if calleePrefix != "" && !in.isBuiltinOrSpecialName(ref.Name) {
					ref.Prefix = calleePrefix
				}
			} else if calleePrefix != "" && !in.isBuiltinOrSpecialName(ref.Name) {
				ref.Prefix = calleePrefix
			}
		} else if ref != nil && ref.Prefix != "" {
			if prefixMap != nil {
				if mapped, ok := prefixMap[ref.Prefix]; ok && mapped != "" {
					ref.Prefix = mapped
				}
			}
		}
		cloned.T = &program.Operand_FnCall{
			FnCall: &program.FnCall{
				Reference: ref,
				Arguments: newArgs,
			},
		}
		return cloned

	case *program.Operand_PropRef:
		cloned := &program.Operand{
			Start:       program.CloneLocation(op.Start),
			End:         program.CloneLocation(op.End),
			Type:        program.CloneType(op.Type),
			ConstValues: op.ConstValues,
		}
		if t.PropRef != nil && t.PropRef.Reference != nil && t.PropRef.Reference.Prefix == "" {
			if env != nil {
				if repl, ok := env[t.PropRef.Reference.Name]; ok && repl != nil {
					propIdx := -1
					switch t.PropRef.Property {
					case "x", "r":
						propIdx = 0
					case "y", "g":
						propIdx = 1
					case "z", "b":
						propIdx = 2
					case "w", "a":
						propIdx = 3
					}
					if propIdx >= 0 {
						if exprOp, ok := repl.T.(*program.Operand_Expression); ok && exprOp.Expression.Operation == program.Operation_COLLECT {
							if propIdx < len(exprOp.Expression.Operands) {
								return program.CloneOperand(exprOp.Expression.Operands[propIdx])
							}
						}
						if len(repl.ConstValues) > propIdx {
							return &program.Operand{
								Start:       program.CloneLocation(op.Start),
								End:         program.CloneLocation(op.End),
								T:           &program.Operand_Value{Value: repl.ConstValues[propIdx]},
								ConstValues: []float64{repl.ConstValues[propIdx]},
							}
						}
					}
				}
			}
		}
		cloned.T = &program.Operand_PropRef{
			PropRef: program.ClonePropertyRef(t.PropRef),
		}
		return cloned

	default:
		return program.CloneOperand(op)
	}
}

func (in *inliner) substituteFn(fn *program.Function, env map[string]*program.Operand, calleePrefix string, anonRenames map[string]string, prefixMap map[string]string) {
	nestedEnv := make(map[string]*program.Operand, len(env))
	for k, v := range env {
		nestedEnv[k] = v
	}
	for _, p := range fn.Parameters {
		nestedEnv[p.Name] = nil
	}
	for _, a := range fn.Assignments {
		for _, v := range a.Variables {
			nestedEnv[v.Name] = nil
		}
	}
	for _, a := range fn.Assignments {
		a.Expression = in.substitute(a.Expression, nestedEnv, calleePrefix, anonRenames, prefixMap)
	}
	for i, ret := range fn.Returns {
		fn.Returns[i] = in.substitute(ret, nestedEnv, calleePrefix, anonRenames, prefixMap)
	}
	for _, subFt := range fn.FnTypes {
		in.substituteFnType(subFt, nestedEnv, calleePrefix, anonRenames, prefixMap)
	}
}

func (in *inliner) substituteFnType(fnType *program.FnType, env map[string]*program.Operand, calleePrefix string, anonRenames map[string]string, prefixMap map[string]string) {
	for _, grp := range fnType.Groups {
		for _, fn := range grp.Functions {
			in.substituteFn(fn, env, calleePrefix, anonRenames, prefixMap)
		}
	}
}
