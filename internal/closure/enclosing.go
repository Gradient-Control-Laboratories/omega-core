package closure

import (
	"sort"

	"gradientcontrol.com/omega/core/program"
)

// cloneListArgs copies every list bound at a call, so the renaming below
// touches only the copy the clone is given — a named list may still be read
// elsewhere.
func cloneListArgs(listArgs map[string][]*program.Operand) map[string][]*program.Operand {
	out := make(map[string][]*program.Operand, len(listArgs))
	for k, els := range listArgs {
		out[k] = cloneOperands(els)
	}
	return out
}

// enclosingNamesForList finds the parameters and locals of enclosingFn that
// the elements of a list bound into clonedFn read, and turns each into a
// parameter of the clone and an argument of the call: the parameter first, in
// order of first use, and the argument the enclosing function's own name for
// it. A name the clone already uses for something of its own is renamed in the
// elements, so the two cannot be confused.
//
// Function values are left alone: a list element that names a closure is
// resolved as a function by the fold, not passed as a value.
func (t *transformer) enclosingNamesForList(file *program.PackageFile, enclosingFn, clonedFn *program.Function, listArgs map[string][]*program.Operand, fc *program.FnCall) ([]*program.Parameter, []*program.Operand) {
	scope := t.enclosingValueNames(file, enclosingFn)

	taken := make(map[string]bool)
	for _, p := range clonedFn.Parameters {
		taken[p.Name] = true
	}
	for _, a := range clonedFn.Assignments {
		for _, v := range a.Variables {
			taken[v.Name] = true
		}
	}
	for _, ft := range clonedFn.FnTypes {
		taken[ft.Name] = true
	}

	keys := make([]string, 0, len(listArgs))
	for k := range listArgs {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var order []string
	first := make(map[string]*program.Operand)
	for _, k := range keys {
		for _, el := range listArgs[k] {
			walkBareReferences(el, func(op *program.Operand, ref *program.Reference) {
				if !scope[ref.Name] {
					return
				}
				if _, seen := first[ref.Name]; !seen {
					first[ref.Name] = op
					order = append(order, ref.Name)
				}
			})
		}
	}

	var params []*program.Parameter
	var args []*program.Operand
	for _, name := range order {
		inner := name
		for taken[inner] {
			inner = inner + "_enc"
		}
		taken[inner] = true
		if inner != name {
			for _, k := range keys {
				for _, el := range listArgs[k] {
					walkBareReferences(el, func(_ *program.Operand, ref *program.Reference) {
						if ref.Name == name {
							ref.Name = inner
						}
					})
				}
			}
		}
		params = append(params, &program.Parameter{Name: inner})
		at := first[name]
		args = append(args, program.NewReference("", name, at.Start, at.End))
	}
	return params, args
}

// enclosingValueNames is the set of names enclosingFn binds to values — its
// parameters and the locals that do not hold a function — which an expression
// written in its body may read but no other function can.
func (t *transformer) enclosingValueNames(file *program.PackageFile, enclosingFn *program.Function) map[string]bool {
	scope := make(map[string]bool)
	for _, p := range enclosingFn.Parameters {
		scope[p.Name] = true
	}
	for _, a := range enclosingFn.Assignments {
		if t.isFunctionValue(a.Expression, file, enclosingFn) {
			continue
		}
		for _, v := range a.Variables {
			scope[v.Name] = true
		}
	}
	for _, ft := range enclosingFn.FnTypes {
		delete(scope, ft.Name)
	}
	return scope
}

// bindsValueName reports whether some function the descent is lexically inside
// binds name to a value — a parameter or a local. A nested function reads the
// names of every function around it, not only its immediate parent, so the
// whole chain is asked, innermost first: the nearest binding is the one in
// effect. A frame that declares name as a nested FUNCTION ends the walk with
// false, because at that point the name is a function and not a value.
//
// The caller's own enclosingFn is asked first and separately. It is normally
// the top of the stack, but a caller reached outside the descent (an unreached
// library template, say) has no stack, and its parameters still bind.
//
// This is what keeps a bare reference from being mistaken for a package
// function of the same name. Without it, `ball(r) { clear = r * 2;
// inner() { return sphere(clear); } ... }` reads `clear` in a program that
// links glass as glass.clear and bakes the name into sphere, which then cannot
// resolve it.
func (t *transformer) bindsValueName(enclosingFn *program.Function, name string) bool {
	frames := make([]*program.Function, 0, len(t.fnStack)+1)
	if enclosingFn != nil {
		frames = append(frames, enclosingFn)
	}
	for i := len(t.fnStack) - 1; i >= 0; i-- {
		if t.fnStack[i] != enclosingFn {
			frames = append(frames, t.fnStack[i])
		}
	}
	for _, fn := range frames {
		for _, ft := range fn.FnTypes {
			if ft.Name == name {
				return false
			}
		}
		for _, p := range fn.Parameters {
			if p.Name == name {
				return true
			}
		}
		for _, a := range fn.Assignments {
			for _, v := range a.Variables {
				if v != nil && v.Name == name {
					return true
				}
			}
		}
	}
	return false
}

// capturesReadEnclosing reports whether any capture's value reads a name
// enclosingFn binds to a value, so it cannot be baked into a function of its
// own.
func (t *transformer) capturesReadEnclosing(file *program.PackageFile, enclosingFn *program.Function, captures []captureBinding) bool {
	scope := t.enclosingValueNames(file, enclosingFn)
	found := false
	for _, c := range captures {
		walkBareReferences(c.value, func(_ *program.Operand, ref *program.Reference) {
			if scope[ref.Name] {
				found = true
			}
		})
	}
	return found
}

// transformUnreachedTemplateCallers transforms each function of a dependency
// file that the reachable walk did not, is not itself a template, and calls one.
func (t *transformer) transformUnreachedTemplateCallers(file *program.PackageFile) error {
	for _, ft := range file.FnTypes {
		for _, fg := range ft.Groups {
			for _, fn := range fg.Functions {
				if t.processedFns[fn] || t.isHigherOrderTemplate(file, fn) {
					continue
				}
				if !t.callsTemplate(file, fn) {
					continue
				}
				if err := t.transformFunction(file, fn, make(map[string]*closureVal)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// callsTemplate reports whether fn's body folds a LIST OF FUNCTIONS through a
// higher-order template — drill.hole([...]) — which is the one call that
// cannot survive the template's pruning. A template called with an ordinary
// function argument is left to the paths that already handle it: transforming
// those early reorders what the backends emit.
func (t *transformer) callsTemplate(file *program.PackageFile, fn *program.Function) bool {
	found := false
	var visit func(op *program.Operand)
	visit = func(op *program.Operand) {
		if op == nil || found {
			return
		}
		switch v := op.T.(type) {
		case *program.Operand_FnCall:
			fc := v.FnCall
			if fc == nil {
				return
			}
			for _, arg := range fc.Arguments {
				visit(arg)
			}
			if fc.Reference == nil {
				return
			}
			calleeFile, calleeFt := t.lookupFnTypeInScope(fn, file, fc.Reference.Prefix, fc.Reference.Name)
			if calleeFt == nil {
				return
			}
			// A list literal of functions, or a lone function where the callee
			// folds a list: a scalar is a list of length one.
			hasList := false
			for i, arg := range fc.Arguments {
				if t.functionListElements(file, fn, arg) != nil {
					hasList = true
					break
				}
				if t.isFunctionValue(arg, file, fn) {
					if pf, pfn, pname := t.calleeParamAt(file, fc, i); pfn != nil && t.paramIsFunctionList(pf, pfn, pname) {
						hasList = true
						break
					}
				}
			}
			if !hasList {
				return
			}
			for _, cg := range calleeFt.Groups {
				if cg.ParamCount == int32(len(fc.Arguments)) && len(cg.Functions) > 0 &&
					t.isHigherOrderTemplate(calleeFile, cg.Functions[0]) {
					found = true
					return
				}
			}
		case *program.Operand_Expression:
			if v.Expression != nil {
				for _, sub := range v.Expression.Operands {
					visit(sub)
				}
			}
		}
	}
	for _, a := range fn.Assignments {
		visit(a.Expression)
	}
	for _, r := range fn.Returns {
		visit(r)
	}
	return found
}

// walkBareReferences calls fn on every unprefixed variable reference beneath
// op — the ones that resolve in the scope the expression was written in.
// Function names being called are not variables and are skipped.
func walkBareReferences(op *program.Operand, fn func(*program.Operand, *program.Reference)) {
	if op == nil {
		return
	}
	switch v := op.T.(type) {
	case *program.Operand_Reference:
		if v.Reference != nil && v.Reference.Prefix == "" {
			fn(op, v.Reference)
		}
	case *program.Operand_PropRef:
		if v.PropRef != nil && v.PropRef.Reference != nil && v.PropRef.Reference.Prefix == "" {
			fn(op, v.PropRef.Reference)
		}
	case *program.Operand_Expression:
		if v.Expression != nil {
			for _, sub := range v.Expression.Operands {
				walkBareReferences(sub, fn)
			}
		}
	case *program.Operand_FnCall:
		if v.FnCall != nil {
			for _, arg := range v.FnCall.Arguments {
				walkBareReferences(arg, fn)
			}
		}
	}
}
