package closure

import "gradientcontrol.com/omega/core/program"

// callbackDifferentiates reports whether the function a fold names applies a
// derivative operator (.dp, .dtime) to anything, in its own body or in a
// function it calls, transitively. Such a fold cannot wait for lower.Transform:
// the property pass, which runs between, differentiates a parameter only at a
// call whose argument it can see, and a fold's callback is called by the fold.
func (t *transformer) callbackDifferentiates(file *program.PackageFile, enclosingFn *program.Function, prefix, name string, arity int) bool {
	return t.fnDifferentiates(file, enclosingFn, prefix, name, arity, map[*program.Function]bool{})
}

func (t *transformer) fnDifferentiates(file *program.PackageFile, enclosingFn *program.Function, prefix, name string, arity int, seen map[*program.Function]bool) bool {
	calleeFile, ft := t.lookupFnTypeInScope(enclosingFn, file, prefix, name)
	if ft == nil {
		return false
	}
	if calleeFile == nil {
		calleeFile = file
	}
	for _, fg := range ft.Groups {
		if fg == nil || (arity >= 0 && int(fg.ParamCount) != arity) {
			continue
		}
		for _, fn := range fg.Functions {
			if fn == nil || seen[fn] {
				continue
			}
			seen[fn] = true
			if t.bodyDifferentiates(calleeFile, fn, seen) {
				return true
			}
		}
	}
	return false
}

func (t *transformer) bodyDifferentiates(file *program.PackageFile, fn *program.Function, seen map[*program.Function]bool) bool {
	for _, a := range fn.Assignments {
		if a != nil && t.operandDifferentiates(file, fn, a.Expression, seen) {
			return true
		}
	}
	for _, ret := range fn.Returns {
		if t.operandDifferentiates(file, fn, ret, seen) {
			return true
		}
	}
	for _, nft := range fn.FnTypes {
		if nft == nil {
			continue
		}
		for _, fg := range nft.Groups {
			for _, sub := range fg.Functions {
				if sub != nil && !seen[sub] {
					seen[sub] = true
					if t.bodyDifferentiates(file, sub, seen) {
						return true
					}
				}
			}
		}
	}
	return false
}

func (t *transformer) operandDifferentiates(file *program.PackageFile, fn *program.Function, op *program.Operand, seen map[*program.Function]bool) bool {
	if op == nil {
		return false
	}
	switch v := op.T.(type) {
	case *program.Operand_Expression:
		if v.Expression == nil {
			return false
		}
		if v.Expression.Operation == program.Operation_DERIV {
			return true
		}
		for _, sub := range v.Expression.Operands {
			if t.operandDifferentiates(file, fn, sub, seen) {
				return true
			}
		}
	case *program.Operand_FnCall:
		if v.FnCall == nil {
			return false
		}
		for _, a := range v.FnCall.Arguments {
			if t.operandDifferentiates(file, fn, a, seen) {
				return true
			}
		}
		if r := v.FnCall.Reference; r != nil {
			return t.fnDifferentiates(file, fn, r.Prefix, r.Name, len(v.FnCall.Arguments), seen)
		}
	}
	return false
}

// liftEnclosingReads gives each function of ft the names it reads (in its body
// or its nested functions' bodies) that are enclosingFn's parameters or locals
// and not its own, as trailing parameters of the same names, and returns those
// names in order: the caller passes them as arguments. A top-level fold
// (enclosingFn nil) reads only package names, which need nothing.
func liftEnclosingReads(enclosingFn *program.Function, ft *program.FnType) []string {
	if enclosingFn == nil || ft == nil {
		return nil
	}
	outer := map[string]bool{}
	for _, prm := range enclosingFn.Parameters {
		outer[prm.Name] = true
	}
	for _, a := range enclosingFn.Assignments {
		if a == nil {
			continue
		}
		for _, v := range a.Variables {
			outer[v.Name] = true
		}
	}
	var names []string
	for _, fg := range ft.Groups {
		for _, fn := range fg.Functions {
			if fn == nil {
				continue
			}
			own := map[string]bool{}
			for _, prm := range fn.Parameters {
				own[prm.Name] = true
			}
			for _, a := range fn.Assignments {
				if a == nil {
					continue
				}
				for _, v := range a.Variables {
					own[v.Name] = true
				}
			}
			seen := map[string]bool{}
			var read func(op *program.Operand)
			read = func(op *program.Operand) {
				if op == nil {
					return
				}
				switch v := op.T.(type) {
				case *program.Operand_Reference:
					if r := v.Reference; r != nil && r.Prefix == "" && outer[r.Name] && !own[r.Name] && !seen[r.Name] {
						seen[r.Name] = true
						names = append(names, r.Name)
					}
				case *program.Operand_PropRef:
					if r := v.PropRef.GetReference(); r != nil && r.Prefix == "" && outer[r.Name] && !own[r.Name] && !seen[r.Name] {
						seen[r.Name] = true
						names = append(names, r.Name)
					}
				case *program.Operand_Expression:
					for _, sub := range v.Expression.GetOperands() {
						read(sub)
					}
				case *program.Operand_FnCall:
					for _, a := range v.FnCall.GetArguments() {
						read(a)
					}
				}
			}
			var walk func(f *program.Function)
			walk = func(f *program.Function) {
				for _, a := range f.Assignments {
					if a != nil {
						read(a.Expression)
					}
				}
				for _, ret := range f.Returns {
					read(ret)
				}
				for _, nft := range f.FnTypes {
					for _, nfg := range nft.GetGroups() {
						for _, sub := range nfg.Functions {
							if sub != nil {
								walk(sub)
							}
						}
					}
				}
			}
			walk(fn)
			for _, n := range names {
				fn.Parameters = append(fn.Parameters, &program.Parameter{Name: n})
			}
			fg.ParamCount = int32(len(fn.Parameters))
		}
	}
	return names
}

// hoistToFile moves ft from enclosingFn's nested functions to file's, where
// lookups by name then find it. A lambda already at file level stays.
func (t *transformer) hoistToFile(file *program.PackageFile, enclosingFn *program.Function, ft *program.FnType) {
	if enclosingFn == nil || file == nil {
		return
	}
	for i, nft := range enclosingFn.FnTypes {
		if nft == ft {
			enclosingFn.FnTypes = append(enclosingFn.FnTypes[:i], enclosingFn.FnTypes[i+1:]...)
			file.FnTypes = append(file.FnTypes, ft)
			if t.fileFunctions[file] == nil {
				t.fileFunctions[file] = map[string]*program.FnType{}
			}
			t.fileFunctions[file][ft.Name] = ft
			return
		}
	}
}

// paramFoldedDifferentiating reports whether fn folds its parameter paramName
// with a function that differentiates (callbackDifferentiates), or passes it
// to a function whose matching parameter is, transitively. A list of fields
// bound to such a parameter has to reach the fold as elements this pass can
// see, as a list of functions does, so the call is specialized on it.
func (t *transformer) paramFoldedDifferentiating(file *program.PackageFile, fn *program.Function, paramName string, seen map[listParamKey]bool) bool {
	if fn == nil {
		return false
	}
	key := listParamKey{fn: fn, param: paramName}
	if seen[key] {
		return false
	}
	seen[key] = true
	isParam := func(op *program.Operand) bool {
		ref, ok := op.T.(*program.Operand_Reference)
		return ok && ref.Reference != nil && ref.Reference.Prefix == "" && ref.Reference.Name == paramName
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
				if cbOp, collOp, _, ok := t.foldParts(file, fn, v.Expression); ok && collOp != nil && isParam(collOp) {
					if prefix, name, arity, ok := t.callbackArity(file, fn, cbOp); ok && t.callbackDifferentiates(file, fn, prefix, name, arity) {
						found = true
						return
					}
				}
			}
			for _, sub := range v.Expression.Operands {
				walk(sub)
			}
		case *program.Operand_FnCall:
			if v.FnCall == nil {
				return
			}
			for i, arg := range v.FnCall.Arguments {
				if isParam(arg) && v.FnCall.Reference != nil {
					calleeFile, ft := t.lookupFnTypeInScope(fn, file, v.FnCall.Reference.Prefix, v.FnCall.Reference.Name)
					if calleeFile == nil {
						calleeFile = file
					}
					if ft != nil {
						for _, fg := range ft.Groups {
							if int(fg.ParamCount) != len(v.FnCall.Arguments) || len(fg.Functions) == 0 {
								continue
							}
							callee := fg.Functions[0]
							if i < len(callee.Parameters) && t.paramFoldedDifferentiating(calleeFile, callee, callee.Parameters[i].Name, seen) {
								found = true
								return
							}
						}
					}
				}
				walk(arg)
			}
		}
	}
	for _, a := range fn.Assignments {
		if a != nil {
			walk(a.Expression)
		}
	}
	for _, ret := range fn.Returns {
		walk(ret)
	}
	return found
}

func calleeFileOr(calleeFile, file *program.PackageFile) *program.PackageFile {
	if calleeFile != nil {
		return calleeFile
	}
	return file
}

// functionReads reports whether fn, or a function nested in it, reads the
// unqualified name.
func functionReads(fn *program.Function, name string) bool {
	var read func(op *program.Operand) bool
	read = func(op *program.Operand) bool {
		if op == nil {
			return false
		}
		switch v := op.T.(type) {
		case *program.Operand_Reference:
			return v.Reference != nil && v.Reference.Prefix == "" && v.Reference.Name == name
		case *program.Operand_PropRef:
			r := v.PropRef.GetReference()
			return r != nil && r.Prefix == "" && r.Name == name
		case *program.Operand_Expression:
			for _, sub := range v.Expression.GetOperands() {
				if read(sub) {
					return true
				}
			}
		case *program.Operand_FnCall:
			for _, a := range v.FnCall.GetArguments() {
				if read(a) {
					return true
				}
			}
		}
		return false
	}
	for _, a := range fn.Assignments {
		if a != nil && read(a.Expression) {
			return true
		}
	}
	for _, ret := range fn.Returns {
		if read(ret) {
			return true
		}
	}
	for _, nft := range fn.FnTypes {
		for _, fg := range nft.GetGroups() {
			for _, sub := range fg.Functions {
				if sub != nil && functionReads(sub, name) {
					return true
				}
			}
		}
	}
	return false
}
