// Package constspec specializes a function on the constant arguments that
// decide its selects.
//
// The inliner prunes a select whose condition is constant, but it inlines only
// calls whose every argument is constant. A function that takes shapes and a
// constant that picks between two constructions,
//
//	insert(iface, head, clamped, threaded, grip) {
//	  screw = remap(head == flatHead ? countersunk(...) : shcs(...), at);
//	  ...
//	}
//
// is called with bodies that read p, so it is never inlined, the condition
// reads a parameter, and both constructions ship in every call's field. This
// pass gives such a call a CLONE of the function with that argument bound: the
// parameter becomes a local holding the constant, the call drops the argument,
// and the next constant fold prunes the branch the constant does not take.
//
// Only a call that makes some select's condition constant is specialized: one
// whose condition reads nothing but parameters this call passes scalar
// constants for, constants, and locals computed from them — not p, not time,
// not another parameter. The parameters it reads are bound and every other
// stays a parameter, so a function is cloned once per distinct deciding value,
// not once per call, and never where a clone would prune nothing.
package constspec

import (
	"fmt"
	"sort"
	"strings"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/program"
)

// Transform specializes every call in prog that passes a scalar constant for a
// parameter deciding a select in the callee, and reports whether it changed
// anything. Arguments are read for constants from the ConstValues the constant
// fold stamps, so it runs after a fold.
func Transform(prog *program.Program) (bool, error) {
	if prog == nil {
		return false, nil
	}
	s := &specializer{
		fns:    map[fnKey]*fnEntry{},
		conds:  map[*program.Function][]condition{},
		clones: map[string]string{},
		taken:  map[string]bool{},
	}
	pkgs := append([]*program.Package{prog.Primary}, prog.Dependencies...)
	for i, pkg := range pkgs {
		if pkg != nil {
			s.index(pkg, pkgPathOf(pkg, i))
		}
	}
	for i, pkg := range pkgs {
		if pkg != nil {
			s.rewritePackage(pkg, pkgPathOf(pkg, i))
		}
	}
	return s.modified, nil
}

// pkgPathOf is the path references into pkg carry: none for the primary.
func pkgPathOf(pkg *program.Package, i int) string {
	if i == 0 {
		return ""
	}
	return pkg.Path
}

type fnKey struct {
	pkgPath string
	name    string
	params  int
}

type fnEntry struct {
	file *program.PackageFile
	name string
	fn   *program.Function
}

type specializer struct {
	fns   map[fnKey]*fnEntry
	conds map[*program.Function][]condition
	// clones maps a callee and its bound arguments to the clone's name.
	clones   map[string]string
	taken    map[string]bool
	modified bool
}

// index records every top-level function overload of pkg.
func (s *specializer) index(pkg *program.Package, pkgPath string) {
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, ft := range f.FnTypes {
			if ft == nil {
				continue
			}
			s.taken[pkgPath+"."+ft.Name] = true
			for _, g := range ft.Groups {
				if g == nil {
					continue
				}
				for _, fn := range g.Functions {
					if fn != nil {
						s.fns[fnKey{pkgPath, ft.Name, int(g.ParamCount)}] = &fnEntry{file: f, name: ft.Name, fn: fn}
					}
				}
			}
		}
	}
}

// rewritePackage specializes the calls in pkg's declarations, returns and
// function bodies. Clones are added as the walk goes; a clone binds every
// argument that decided anything, so its own calls are rewritten only for the
// functions it calls in turn.
func (s *specializer) rewritePackage(pkg *program.Package, pkgPath string) {
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, a := range f.Assignments {
			if a != nil {
				s.operand(a.Expression, pkgPath)
				for _, v := range a.Variables {
					for _, prop := range v.GetProperties() {
						s.operand(prop.Expression, pkgPath)
					}
				}
			}
		}
		for _, r := range f.Returns {
			s.operand(r, pkgPath)
		}
		for i := 0; i < len(f.FnTypes); i++ { // clones appended mid-walk are walked too
			s.fnType(f.FnTypes[i], pkgPath)
		}
	}
}

func (s *specializer) fnType(ft *program.FnType, pkgPath string) {
	if ft == nil {
		return
	}
	for _, g := range ft.Groups {
		if g == nil {
			continue
		}
		for _, fn := range g.Functions {
			s.function(fn, pkgPath)
		}
	}
}

func (s *specializer) function(fn *program.Function, pkgPath string) {
	if fn == nil {
		return
	}
	for _, a := range fn.Assignments {
		if a != nil {
			s.operand(a.Expression, pkgPath)
		}
	}
	for _, r := range fn.Returns {
		s.operand(r, pkgPath)
	}
	for _, ft := range fn.FnTypes {
		s.fnType(ft, pkgPath)
	}
}

func (s *specializer) operand(op *program.Operand, pkgPath string) {
	if op == nil {
		return
	}
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		if t.Expression != nil {
			for _, o := range t.Expression.Operands {
				s.operand(o, pkgPath)
			}
		}
	case *program.Operand_FnCall:
		if t.FnCall != nil {
			for _, a := range t.FnCall.Arguments {
				s.operand(a, pkgPath)
			}
			s.call(t.FnCall, pkgPath)
		}
	}
}

// call redirects fc to a clone of its callee with its constant deciding
// arguments bound, if it passes any.
func (s *specializer) call(fc *program.FnCall, pkgPath string) {
	if fc.Reference == nil {
		return
	}
	path := fc.Reference.PkgPath
	if path == "" && fc.Reference.Prefix == "" {
		path = pkgPath
	}
	e := s.fns[fnKey{path, fc.Reference.Name, len(fc.Arguments)}]
	if e == nil || len(e.fn.FnTypes) > 0 {
		return
	}
	type bound struct {
		index int
		value float64
	}
	constArg := map[int]float64{}
	for i, a := range fc.Arguments {
		if v, ok := scalarConst(a); ok {
			constArg[i] = v
		}
	}
	var binds []bound
	bindSet := map[int]bool{}
	for _, cond := range s.conditions(e.fn) {
		if cond.dynamic {
			continue
		}
		all := true
		for _, i := range cond.params {
			if _, ok := constArg[i]; !ok {
				all = false
				break
			}
		}
		if !all {
			continue
		}
		for _, i := range cond.params {
			bindSet[i] = true
		}
	}
	for i := range fc.Arguments {
		if bindSet[i] {
			binds = append(binds, bound{i, constArg[i]})
		}
	}
	if len(binds) == 0 {
		return
	}
	parts := []string{path, e.name, fmt.Sprint(len(fc.Arguments))}
	for _, b := range binds {
		parts = append(parts, fmt.Sprintf("%d=%v", b.index, b.value))
	}
	key := strings.Join(parts, "|")
	name, ok := s.clones[key]
	if !ok {
		name = s.freshName(path, e.name)
		s.clones[key] = name
		clone := proto.Clone(e.fn).(*program.Function)
		bindAt := map[int]float64{}
		for _, b := range binds {
			bindAt[b.index] = b.value
		}
		var params []*program.Parameter
		var locals []*program.Assignment
		for i, prm := range clone.Parameters {
			v, isBound := bindAt[i]
			if !isBound {
				params = append(params, prm)
				continue
			}
			locals = append(locals, &program.Assignment{
				Start:      prm.Start,
				End:        prm.Start,
				Variables:  []*program.Variable{{Name: prm.Name}},
				Expression: &program.Operand{Start: prm.Start, End: prm.Start, T: &program.Operand_Value{Value: v}},
			})
		}
		clone.Parameters = params
		clone.Assignments = append(locals, clone.Assignments...)
		e.file.FnTypes = append(e.file.FnTypes, &program.FnType{
			Name:   name,
			Groups: []*program.FnGroup{{ParamCount: int32(len(params)), Functions: []*program.Function{clone}}},
		})
		s.fns[fnKey{path, name, len(params)}] = &fnEntry{file: e.file, name: name, fn: clone}
	}
	drop := map[int]bool{}
	for _, b := range binds {
		drop[b.index] = true
	}
	var args []*program.Operand
	for i, a := range fc.Arguments {
		if !drop[i] {
			args = append(args, a)
		}
	}
	fc.Arguments = args
	fc.Reference.Name = name
	fc.Reference.RefPath = ""
	s.modified = true
}

// freshName is an unused name for a clone of name in the package at path.
func (s *specializer) freshName(path, name string) string {
	for n := 1; ; n++ {
		c := fmt.Sprintf("%s__spec%d", name, n)
		if !s.taken[path+"."+c] {
			s.taken[path+"."+c] = true
			return c
		}
	}
}

// scalarConst is op's value if it is one constant scalar.
func scalarConst(op *program.Operand) (float64, bool) {
	if op == nil {
		return 0, false
	}
	if v, ok := op.T.(*program.Operand_Value); ok {
		return v.Value, true
	}
	if len(op.ConstValues) == 1 {
		return op.ConstValues[0], true
	}
	return 0, false
}

// condition is what one select condition in a function reads: the indices of
// the parameters, and whether it reads anything else that varies — p, time, or
// a name that is neither a parameter, a local nor a folded constant.
type condition struct {
	params  []int
	dynamic bool
}

// conditions are the select conditions of fn's body, each with what it reads
// through the locals it is computed from.
func (s *specializer) conditions(fn *program.Function) []condition {
	if cs, ok := s.conds[fn]; ok {
		return cs
	}
	defs := map[string]*program.Operand{}
	for _, a := range fn.Assignments {
		if a != nil && len(a.Variables) == 1 {
			defs[a.Variables[0].Name] = a.Expression
		}
	}
	paramAt := map[string]int{}
	for i, prm := range fn.Parameters {
		paramAt[prm.Name] = i
	}
	var conds []*program.Operand
	for _, a := range fn.Assignments {
		if a != nil {
			walkSelects(a.Expression, func(c *program.Operand) { conds = append(conds, c) })
		}
	}
	for _, r := range fn.Returns {
		walkSelects(r, func(c *program.Operand) { conds = append(conds, c) })
	}
	var out []condition
	for _, c := range conds {
		var cond condition
		seen := map[string]bool{}
		var visit func(op *program.Operand)
		visit = func(op *program.Operand) {
			if op == nil || cond.dynamic {
				return
			}
			if len(op.ConstValues) > 0 {
				return
			}
			switch t := op.T.(type) {
			case *program.Operand_Value:
			case *program.Operand_Reference:
				r := t.Reference
				if r == nil || r.Prefix != "" {
					cond.dynamic = true
					return
				}
				if seen[r.Name] {
					return
				}
				seen[r.Name] = true
				if i, ok := paramAt[r.Name]; ok {
					cond.params = append(cond.params, i)
				} else if def, ok := defs[r.Name]; ok {
					visit(def)
				} else {
					cond.dynamic = true // p, time, or a declaration that is not folded
				}
			case *program.Operand_Expression:
				if t.Expression != nil {
					for _, o := range t.Expression.Operands {
						visit(o)
					}
				}
			case *program.Operand_FnCall:
				// A call may read p, time or the pixel size without naming them
				// (sphere(r), pixelSize()); only pure arithmetic of its arguments
				// is known not to.
				if t.FnCall == nil || t.FnCall.Reference == nil || t.FnCall.Reference.Prefix != "" ||
					!pureBuiltins[t.FnCall.Reference.Name] {
					cond.dynamic = true
					return
				}
				for _, a := range t.FnCall.Arguments {
					visit(a)
				}
			default:
				cond.dynamic = true
			}
		}
		visit(c)
		sort.Ints(cond.params)
		out = append(out, cond)
	}
	s.conds[fn] = out
	return out
}

// pureBuiltins are the native builtins that are functions of their arguments
// alone. The rest of program.IsBuiltin's names read meshes, fields or the UI.
var pureBuiltins = map[string]bool{
	"acosh": true, "asinh": true, "atanh": true, "clamp": true, "cosh": true, "cross": true,
	"degrees": true, "distance": true, "dot": true, "exp2": true, "fract": true, "length": true,
	"normalize": true, "radians": true, "sign": true, "sinh": true, "tanh": true, "trunc": true,
}

// walkSelects calls cond with the condition of every select beneath op.
func walkSelects(op *program.Operand, cond func(*program.Operand)) {
	if op == nil {
		return
	}
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		if t.Expression == nil {
			return
		}
		if t.Expression.Operation == program.Operation_SELECT && len(t.Expression.Operands) >= 3 {
			cond(t.Expression.Operands[0])
		}
		for _, o := range t.Expression.Operands {
			walkSelects(o, cond)
		}
	case *program.Operand_FnCall:
		if t.FnCall != nil {
			for _, a := range t.FnCall.Arguments {
				walkSelects(a, cond)
			}
		}
	}
}
