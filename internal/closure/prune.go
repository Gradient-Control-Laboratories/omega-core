package closure

import (
	"gradientcontrol.com/omega/core/program"
)

func (t *transformer) pruneHigherOrderTemplates(files []*program.PackageFile) {
	vanished := t.pruneFunctions(files, func(file *program.PackageFile, fn *program.Function) bool {
		return t.isHigherOrderTemplate(file, fn)
	})

	// Call sites are specialized only in functions the transform reached, so
	// an unreached library function still calls the templates just pruned.
	// When that removed every overload of a name, the dangling call fails the
	// re-link, though nothing will ever run the caller. Prune such callers
	// too, until removing one empties no further name.
	for len(vanished) > 0 {
		vanished = t.pruneFunctions(files, func(file *program.PackageFile, fn *program.Function) bool {
			return !t.processedFns[fn] && t.referencesFnType(file, fn, vanished)
		})
	}
}

// pruneFunctions removes the top-level functions for which prune reports true
// and returns the function types it emptied, whose names no longer resolve.
func (t *transformer) pruneFunctions(files []*program.PackageFile, prune func(*program.PackageFile, *program.Function) bool) map[*program.FnType]bool {
	vanished := make(map[*program.FnType]bool)
	for _, file := range files {
		var remainingFnTypes []*program.FnType
		for _, ft := range file.FnTypes {
			var remainingGroups []*program.FnGroup
			for _, fg := range ft.Groups {
				var remainingFns []*program.Function
				for _, fn := range fg.Functions {
					if !prune(file, fn) {
						remainingFns = append(remainingFns, fn)
					}
				}
				if len(remainingFns) > 0 {
					fg.Functions = remainingFns
					remainingGroups = append(remainingGroups, fg)
				}
			}
			if len(remainingGroups) > 0 {
				ft.Groups = remainingGroups
				remainingFnTypes = append(remainingFnTypes, ft)
			} else {
				vanished[ft] = true
			}
		}
		file.FnTypes = remainingFnTypes
	}
	return vanished
}

// referencesFnType reports whether fn, or a function nested in it, names one
// of fnTypes, by call or as a value. A name bound by a parameter or an
// assignment in an enclosing body shadows the function of that name.
func (t *transformer) referencesFnType(file *program.PackageFile, fn *program.Function, fnTypes map[*program.FnType]bool) bool {
	return t.fnReferencesFnType(file, fn, fn, nil, fnTypes)
}

func (t *transformer) fnReferencesFnType(file *program.PackageFile, top, fn *program.Function, outer map[string]bool, fnTypes map[*program.FnType]bool) bool {
	bound := make(map[string]bool, len(outer)+len(fn.Parameters)+len(fn.Assignments))
	for name := range outer {
		bound[name] = true
	}
	for _, p := range fn.Parameters {
		bound[p.Name] = true
	}
	for _, assign := range fn.Assignments {
		for _, v := range assign.Variables {
			bound[v.Name] = true
		}
	}

	names := func(ref *program.Reference) bool {
		if ref == nil || (ref.Prefix == "" && bound[ref.Name]) {
			return false
		}
		_, ft := t.lookupFnTypeInScope(top, file, ref.Prefix, ref.Name)
		return ft != nil && fnTypes[ft]
	}
	var walk func(op *program.Operand) bool
	walk = func(op *program.Operand) bool {
		if op == nil {
			return false
		}
		switch v := op.T.(type) {
		case *program.Operand_Reference:
			return names(v.Reference)
		case *program.Operand_FnCall:
			if v.FnCall == nil {
				return false
			}
			if names(v.FnCall.Reference) {
				return true
			}
			for _, arg := range v.FnCall.Arguments {
				if walk(arg) {
					return true
				}
			}
		case *program.Operand_Expression:
			if v.Expression == nil {
				return false
			}
			for _, sub := range v.Expression.Operands {
				if walk(sub) {
					return true
				}
			}
		}
		return false
	}

	for _, assign := range fn.Assignments {
		if walk(assign.Expression) {
			return true
		}
	}
	for _, ret := range fn.Returns {
		if walk(ret) {
			return true
		}
	}
	for _, nft := range fn.FnTypes {
		for _, nfg := range nft.Groups {
			for _, nfn := range nfg.Functions {
				if t.fnReferencesFnType(file, top, nfn, bound, fnTypes) {
					return true
				}
			}
		}
	}
	return false
}

// isHigherOrderTemplate reports whether a function is an unspecialized higher-order
// template declaration that cannot be executed directly by downstream backends.
//
// A "higher-order template" is a function declaration that either:
//  1. Returns a function value (i.e. a closure factory like `scale(k) { return (x) { return x * k; }; }`), or
//  2. Accepts one or more functions/combiners as formal parameters (e.g. `box(sz, combine)` or `apply(x, op)`).
//
// Because target backends (WGSL, JS audio, Fidget CSG/GLSL) lack dynamic function pointers
// and runtime closure allocations, these template functions serve solely as blueprints.
// During transformation, every call site targeting a template is monomorphized into a
// concrete, first-order synthesized function (`gcl_closure_*`). Once all call sites are
// rewritten, the unspecialized template definitions are pruned from the final AST.
func (t *transformer) isHigherOrderTemplate(file *program.PackageFile, fn *program.Function) bool {
	if fn == nil {
		return false
	}
	if t.functionReturnsFunction(fn) {
		return true
	}
	for _, p := range fn.Parameters {
		if t.paramIsFunction(file, fn, p.Name, make(map[*program.Function]bool)) {
			return true
		}
		if t.paramIsFunctionList(file, fn, p.Name) {
			return true
		}
	}
	return false
}

// paramIsFunctionList reports whether paramName is the COLLECTION of a fold
// whose callback CALLS its element — the only shape in which a list of
// functions is consumed, and the only one decidable from the declaration
// alone. Such a function is a template like any other higher-order one: it is
// specialized at each call site against that site's list, and the
// unspecialized declaration, still folding a parameter no backend can iterate,
// cannot be executed.
//
// It is deliberately narrower than "a parameter that is folded": a fold whose
// callback merely passes the element along, as builtin's intersectSharp and
// text's left both do, is an ordinary fold over values and is left alone.
func (t *transformer) paramIsFunctionList(file *program.PackageFile, fn *program.Function, paramName string) bool {
	if fn == nil {
		return false
	}
	key := listParamKey{fn: fn, param: paramName}
	if cached, ok := t.listParamCache[key]; ok {
		return cached
	}
	found := false
	var walk func(op *program.Operand)
	walk = func(op *program.Operand) {
		if found || op == nil {
			return
		}
		switch v := op.T.(type) {
		case *program.Operand_Expression:
			if v.Expression == nil {
				return
			}
			if v.Expression.Operation == program.Operation_MAP || v.Expression.Operation == program.Operation_REDUCE {
				if t.foldCallsElement(file, fn, v.Expression, paramName) {
					found = true
					return
				}
			}
			for _, sub := range v.Expression.Operands {
				walk(sub)
			}
		case *program.Operand_FnCall:
			if v.FnCall == nil {
				return
			}
			for _, arg := range v.FnCall.Arguments {
				walk(arg)
			}
		}
	}
	for _, assign := range fn.Assignments {
		walk(assign.Expression)
	}
	for _, ret := range fn.Returns {
		walk(ret)
	}
	t.listParamCache[key] = found
	return found
}

// foldCallsElement reports whether expr folds paramName with a callback that
// calls the element it is handed, rather than passing it on as a value.
func (t *transformer) foldCallsElement(file *program.PackageFile, fn *program.Function, expr *program.Expression, paramName string) bool {
	cbOp, collOp, _, ok := t.foldParts(file, fn, expr)
	if !ok {
		return false
	}
	ref, ok := collOp.T.(*program.Operand_Reference)
	if !ok || ref.Reference == nil || ref.Reference.Prefix != "" || ref.Reference.Name != paramName {
		return false
	}
	cbRef, ok := cbOp.T.(*program.Operand_Reference)
	if !ok || cbRef.Reference == nil {
		return false
	}
	_, cbFt := t.lookupFnTypeInScope(fn, file, cbRef.Reference.Prefix, cbRef.Reference.Name)
	if cbFt == nil {
		return false
	}
	// reduce hands the element to the callback's second parameter, map to its
	// first; the optional trailing index parameter never holds it.
	elemIdx := 1
	if expr.Operation == program.Operation_MAP {
		elemIdx = 0
	}
	for _, fg := range cbFt.Groups {
		if len(fg.Functions) == 0 {
			continue
		}
		cb := fg.Functions[0]
		if elemIdx < len(cb.Parameters) && t.fnDirectlyCallsParam(cb, cb.Parameters[elemIdx].Name) {
			return true
		}
	}
	return false
}

func (t *transformer) paramIsFunction(file *program.PackageFile, fn *program.Function, paramName string, visited map[*program.Function]bool) bool {
	if fn == nil || visited[fn] {
		return false
	}
	visited[fn] = true

	// Check if paramName is directly called inside fn or any nested functions
	if t.fnDirectlyCallsParam(fn, paramName) {
		return true
	}

	// Check if paramName is passed to an argument position of another function that treats it as a function
	for _, assign := range fn.Assignments {
		if t.operandPassesParamToFunction(file, assign.Expression, paramName, visited) {
			return true
		}
	}
	for _, ret := range fn.Returns {
		if t.operandPassesParamToFunction(file, ret, paramName, visited) {
			return true
		}
	}
	return false
}

func (t *transformer) fnDirectlyCallsParam(fn *program.Function, paramName string) bool {
	params := map[string]bool{paramName: true}
	for _, assign := range fn.Assignments {
		if t.operandDirectlyCallsParam(assign.Expression, params) {
			return true
		}
	}
	for _, ret := range fn.Returns {
		if t.operandDirectlyCallsParam(ret, params) {
			return true
		}
	}
	for _, nft := range fn.FnTypes {
		for _, nfg := range nft.Groups {
			for _, nfn := range nfg.Functions {
				if t.fnDirectlyCallsParam(nfn, paramName) {
					return true
				}
			}
		}
	}
	return false
}

func (t *transformer) operandDirectlyCallsParam(op *program.Operand, params map[string]bool) bool {
	if op == nil {
		return false
	}
	switch v := op.T.(type) {
	case *program.Operand_FnCall:
		if v.FnCall != nil && v.FnCall.Reference != nil && v.FnCall.Reference.Prefix == "" {
			if params[v.FnCall.Reference.Name] {
				return true
			}
		}
		if v.FnCall != nil {
			for _, arg := range v.FnCall.Arguments {
				if t.operandDirectlyCallsParam(arg, params) {
					return true
				}
			}
		}
	case *program.Operand_Expression:
		if v.Expression != nil {
			for _, sub := range v.Expression.Operands {
				if t.operandDirectlyCallsParam(sub, params) {
					return true
				}
			}
		}
	case *program.Operand_PropRef:
		if v.PropRef != nil && v.PropRef.Reference != nil && v.PropRef.Reference.Prefix == "" {
			if params[v.PropRef.Reference.Name] {
				return true
			}
		}
	}
	return false
}

func (t *transformer) operandPassesParamToFunction(file *program.PackageFile, op *program.Operand, paramName string, visited map[*program.Function]bool) bool {
	if op == nil {
		return false
	}
	switch v := op.T.(type) {
	case *program.Operand_FnCall:
		if v.FnCall != nil {
			for i, arg := range v.FnCall.Arguments {
				if argRef, ok := arg.T.(*program.Operand_Reference); ok && argRef.Reference != nil && argRef.Reference.Prefix == "" {
					if argRef.Reference.Name == paramName {
						// Look up callee function in file's scope
						calleeFile, calleeFt := t.lookupFnType(file, v.FnCall.Reference.Prefix, v.FnCall.Reference.Name)
						if calleeFt != nil {
							for _, fg := range calleeFt.Groups {
								if fg.ParamCount == int32(len(v.FnCall.Arguments)) && len(fg.Functions) > 0 {
									calleeFn := fg.Functions[0]
									if i < len(calleeFn.Parameters) {
										calleeParamName := calleeFn.Parameters[i].Name
										targetFile := calleeFile
										if targetFile == nil {
											targetFile = file
										}
										if t.paramIsFunction(targetFile, calleeFn, calleeParamName, visited) ||
											t.paramIsFunctionList(targetFile, calleeFn, calleeParamName) {
											return true
										}
									}
								}
							}
						}
					}
				}
				if collectHasName(arg, paramName) {
					// paramName travels inside a LIST. The callee's parameter
					// is then a list of functions rather than a function, so
					// that is the predicate to ask about.
					if calleeFile, calleeFn, calleeParam := t.calleeParamAt(file, v.FnCall, i); calleeFn != nil {
						if t.paramIsFunctionList(calleeFile, calleeFn, calleeParam) {
							return true
						}
					}
				}
				if t.operandPassesParamToFunction(file, arg, paramName, visited) {
					return true
				}
			}
		}
	case *program.Operand_Expression:
		if v.Expression != nil {
			for _, sub := range v.Expression.Operands {
				if t.operandPassesParamToFunction(file, sub, paramName, visited) {
					return true
				}
			}
		}
	}
	return false
}

// calleeParamAt resolves the callee of fc in file's scope and returns the file
// it lives in, the function itself, and the name of its i-th parameter.
func (t *transformer) calleeParamAt(file *program.PackageFile, fc *program.FnCall, i int) (*program.PackageFile, *program.Function, string) {
	if fc == nil || fc.Reference == nil {
		return nil, nil, ""
	}
	calleeFile, calleeFt := t.lookupFnType(file, fc.Reference.Prefix, fc.Reference.Name)
	if calleeFt == nil {
		return nil, nil, ""
	}
	for _, fg := range calleeFt.Groups {
		if fg.ParamCount != int32(len(fc.Arguments)) || len(fg.Functions) == 0 {
			continue
		}
		calleeFn := fg.Functions[0]
		if i >= len(calleeFn.Parameters) {
			continue
		}
		target := calleeFile
		if target == nil {
			target = file
		}
		return target, calleeFn, calleeFn.Parameters[i].Name
	}
	return nil, nil, ""
}

// collectHasName reports whether op is a list with a bare reference to name
// among its elements.
func collectHasName(op *program.Operand, name string) bool {
	if op == nil {
		return false
	}
	expr, ok := op.T.(*program.Operand_Expression)
	if !ok || expr.Expression == nil || expr.Expression.Operation != program.Operation_COLLECT {
		return false
	}
	for _, el := range expr.Expression.Operands {
		if ref, ok := el.T.(*program.Operand_Reference); ok && ref.Reference != nil &&
			ref.Reference.Prefix == "" && ref.Reference.Name == name {
			return true
		}
	}
	return false
}
