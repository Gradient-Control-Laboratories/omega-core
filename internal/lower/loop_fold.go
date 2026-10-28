package lower

import (
	"fmt"
	"sort"
	"strings"

	"gradientcontrol.com/omega/core/program"
)

// loopable reports whether a reduce whose callback is fnOp can be kept for a
// loop. A backend builds an index into a list from a constant, so a list
// indexed by anything that varies from one step to the next, as ocean's
// frequencies[i] is, builds only with the reduce unrolled, where each step's
// index is a number. So no index anywhere the callback reaches may depend on
// its parameters: in its body, in the functions it calls with them, or in the
// functions it declares and reads them in. A call the search cannot follow to
// a function, with an argument that depends on them, keeps the reduce
// unrolled too.
func loopable(prog *program.Program, pkg *program.Package, file *program.PackageFile, fnStack []*program.Function, fnOp *program.Operand) bool {
	ref := fnOp.GetReference()
	if ref == nil {
		return false
	}
	fns := findFunctions(prog, pkg, file, fnStack, ref.Prefix, ref.Name)
	if len(fns) == 0 {
		return false
	}
	t := &taint{prog: prog, seen: map[string]bool{}}
	for _, f := range fns {
		varying := map[string]bool{}
		for _, p := range f.fn.Parameters {
			varying[p.Name] = true
		}
		if !t.indexesSteadily(f, varying) {
			return false
		}
	}
	return true
}

// taint follows what varies from one step of a loop to the next through the
// functions a callback reaches.
type taint struct {
	prog *program.Program
	// seen marks each function already searched with a set of varying names.
	seen map[string]bool
}

// foundFunction is a function, the package it was found in, and the functions
// it is declared in, outermost first, ending with itself.
type foundFunction struct {
	fn    *program.Function
	pkg   *program.Package
	stack []*program.Function
}

// indexesSteadily reports whether f indexes nothing by a name in varying, or
// by a local computed from one, and passes none of them on to a function that
// does. varying holds f's own varying parameters and the names around it that
// vary, which f reads by closure.
func (t *taint) indexesSteadily(f foundFunction, varying map[string]bool) bool {
	key := searchKey(f.fn, varying)
	if t.seen[key] {
		return true
	}
	t.seen[key] = true

	varying = copySet(varying)
	ok := true
	// call searches the functions prefix.name names, each parameter made
	// varying whose argument depends on a varying name, or every one when
	// allVary is set.
	call := func(prefix, name string, args []*program.Operand, allVary bool) {
		fns := findFunctions(t.prog, f.pkg, nil, f.stack, prefix, name)
		if len(fns) == 0 {
			if (allVary || mentionsAny(args, varying)) && !(prefix == "" && program.BuiltinNamed(name) != nil) {
				ok = false
			}
			return
		}
		for _, g := range fns {
			// A function declared around here reads the varying names by
			// closure; one declared elsewhere sees only its parameters.
			inner := map[string]bool{}
			if nestedIn(g, f.stack) {
				inner = copySet(varying)
			}
			for i, p := range g.fn.Parameters {
				if allVary || (i < len(args) && mentions(args[i], varying)) {
					inner[p.Name] = true
				}
			}
			if len(inner) == 0 {
				continue
			}
			if !t.indexesSteadily(g, inner) {
				ok = false
				return
			}
		}
	}
	// isFunction reports whether op names a function, passed as a value.
	isFunction := func(op *program.Operand) bool {
		r := op.GetReference()
		return r != nil && len(findFunctions(t.prog, f.pkg, nil, f.stack, r.Prefix, r.Name)) > 0
	}
	var visit func(op *program.Operand)
	visit = func(op *program.Operand) {
		if op == nil || !ok {
			return
		}
		var operands []*program.Operand
		switch e := op.T.(type) {
		case *program.Operand_Expression:
			x := e.Expression
			if x.Operation == program.Operation_INDEX && len(x.Operands) == 2 && mentions(x.Operands[1], varying) {
				ok = false
				return
			}
			operands = x.Operands
		case *program.Operand_FnCall:
			if r := e.FnCall.Reference; r != nil {
				call(r.Prefix, r.Name, e.FnCall.Arguments, false)
			}
			operands = e.FnCall.Arguments
		default:
			return
		}
		// A function passed as a value, a callback, is called with what the
		// operation hands it, which varies if anything else passed does.
		anyVary := false
		for _, o := range operands {
			anyVary = anyVary || (!isFunction(o) && mentions(o, varying))
		}
		for _, o := range operands {
			if isFunction(o) {
				r := o.GetReference()
				call(r.Prefix, r.Name, nil, anyVary)
			}
			visit(o)
		}
	}
	for _, a := range f.fn.Assignments {
		visit(a.Expression)
		for _, v := range a.Variables {
			for _, prop := range v.Properties {
				visit(prop.Expression)
			}
		}
		if mentions(a.Expression, varying) {
			for _, v := range a.Variables {
				varying[v.Name] = true
			}
		}
	}
	for _, r := range f.fn.Returns {
		visit(r)
	}
	return ok
}

// mentions reports whether op reads any name in names.
func mentions(op *program.Operand, names map[string]bool) bool {
	if op == nil {
		return false
	}
	switch t := op.T.(type) {
	case *program.Operand_Reference:
		return t.Reference.Prefix == "" && names[t.Reference.Name]
	case *program.Operand_PropRef:
		r := t.PropRef.GetReference()
		return r != nil && r.Prefix == "" && names[r.Name]
	case *program.Operand_Expression:
		return mentionsAny(t.Expression.Operands, names)
	case *program.Operand_FnCall:
		return mentionsAny(t.FnCall.Arguments, names)
	}
	return false
}

func mentionsAny(ops []*program.Operand, names map[string]bool) bool {
	for _, op := range ops {
		if mentions(op, names) {
			return true
		}
	}
	return false
}

// nestedIn reports whether g is declared inside a function of stack.
func nestedIn(g foundFunction, stack []*program.Function) bool {
	if len(g.stack) < 2 {
		return false
	}
	outer := g.stack[len(g.stack)-2]
	for _, s := range stack {
		if s == outer {
			return true
		}
	}
	return false
}

func copySet(s map[string]bool) map[string]bool {
	c := make(map[string]bool, len(s))
	for k, v := range s {
		c[k] = v
	}
	return c
}

func searchKey(fn *program.Function, varying map[string]bool) string {
	names := make([]string, 0, len(varying))
	for k := range varying {
		names = append(names, k)
	}
	sort.Strings(names)
	return fmt.Sprintf("%p|%s", fn, strings.Join(names, "|"))
}

// findFunctions returns every function named name visible from stack in pkg
// (file, when given, is searched first) or in the builtin package every
// package imports, or else in the package prefix names.
func findFunctions(prog *program.Program, pkg *program.Package, file *program.PackageFile, stack []*program.Function, prefix, name string) []foundFunction {
	var found []foundFunction
	collect := func(fnTypes []*program.FnType, in *program.Package, outer []*program.Function) {
		for _, ft := range fnTypes {
			if ft == nil || ft.Name != name {
				continue
			}
			for _, g := range ft.Groups {
				for _, f := range g.Functions {
					st := append(append([]*program.Function(nil), outer...), f)
					found = append(found, foundFunction{fn: f, pkg: in, stack: st})
				}
			}
		}
	}
	deps := func(match func(path string) bool) {
		if prog == nil {
			return
		}
		for _, dep := range prog.Dependencies {
			if dep != nil && match(dep.Path) {
				for _, f := range dep.Files {
					collect(f.FnTypes, dep, nil)
				}
			}
		}
	}
	if prefix != "" {
		deps(func(path string) bool {
			simpleName := path[strings.LastIndex(path, "/")+1:]
			return path == prefix || simpleName == prefix || strings.HasSuffix(path, "/"+prefix)
		})
		return found
	}
	for i := len(stack) - 1; i >= 0 && len(found) == 0; i-- {
		if stack[i] != nil {
			collect(stack[i].FnTypes, pkg, stack[:i+1])
		}
	}
	if len(found) == 0 && file != nil {
		collect(file.FnTypes, pkg, nil)
	}
	if len(found) == 0 && pkg != nil {
		for _, f := range pkg.Files {
			collect(f.FnTypes, pkg, nil)
		}
	}
	if len(found) == 0 {
		deps(func(path string) bool { return path == "builtin" })
	}
	return found
}
