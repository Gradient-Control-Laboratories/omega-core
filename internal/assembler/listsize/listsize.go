// Package listsize specializes a function on the sizes of the lists it is
// called with, where its body needs them to be constant.
//
// A reduce over a list parameter stays a loop in the generated code, as its
// function's body is shared by calls of every length. That is fine until the
// callback indexes a list by the loop counter (a matrix eliminated column by
// column, matrix.determinant), which the backends allow only with a constant
// index. Such a function, one that reduces over a parameter with a callback
// that takes the element's index, is cloned once per distinct type of the
// list arguments its calls pass; in each clone, once retyped, every indexed
// reduce runs over its collection written out, [x[0], ..., x[n - 1]], which
// the lowering unrolls with constant indices. Other functions keep their
// loops.
package listsize

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Gradient-Control-Laboratories/omega-core/program"
	"google.golang.org/protobuf/proto"
)

// maxRounds bounds the clone-retype rounds: each reaches one more level of
// calls from a clone into another function that needs sizes.
const maxRounds = 8

// Transform specializes prog as described above. retype re-links and retypes
// the program; it is called after each round of clones, so their calls and
// collections carry their own sizes. It reports whether anything changed.
func Transform(prog *program.Program, retype func() error) (bool, error) {
	if prog == nil {
		return false, nil
	}
	// Type inference keeps a top-level expression's type without looking
	// inside it again, so a call redirected beneath one would never be typed:
	// each retype starts from the top-level expressions untyped.
	inner := retype
	retype = func() error {
		clearTopLevelTypes(prog)
		return inner()
	}
	s := &specializer{clones: map[string]string{}, sized: map[*program.Function]bool{}, taken: map[string]bool{}}
	changed := false
	for round := 0; round < maxRounds; round++ {
		s.index(prog)
		cloned := s.cloneRound(prog)
		if cloned {
			if err := retype(); err != nil {
				return true, err
			}
			s.index(prog)
		}
		unrolled := s.unrollReduces()
		bound := s.bindIndexArgs()
		if !cloned && !unrolled && !bound {
			break
		}
		changed = true
		if unrolled || bound {
			if err := retype(); err != nil {
				return true, err
			}
		}
	}
	return changed, nil
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
	path string
}

type specializer struct {
	fns    map[fnKey]*fnEntry
	byName map[string]*fnEntry // pkgPath + "." + name, for callbacks
	clones map[string]string
	sized  map[*program.Function]bool
	taken  map[string]bool
	// fresh are this round's clones: their types are still the template's, so
	// their calls wait for the next round.
	fresh map[*program.Function]bool
	// idxMemo caches indexParamsOf; it is rebuilt with the index each round.
	idxMemo map[*program.Function]map[int]bool
}

func pkgs(prog *program.Program) []*program.Package {
	return append([]*program.Package{prog.Primary}, prog.Dependencies...)
}

func pkgPathOf(pkg *program.Package, i int) string {
	if i == 0 {
		return ""
	}
	return pkg.Path
}

func (s *specializer) index(prog *program.Program) {
	s.idxMemo = nil
	s.fns = map[fnKey]*fnEntry{}
	s.byName = map[string]*fnEntry{}
	for i, pkg := range pkgs(prog) {
		if pkg == nil {
			continue
		}
		path := pkgPathOf(pkg, i)
		for _, f := range pkg.Files {
			if f == nil {
				continue
			}
			for _, ft := range f.FnTypes {
				if ft == nil {
					continue
				}
				s.taken[path+"."+ft.Name] = true
				for _, g := range ft.Groups {
					for _, fn := range g.GetFunctions() {
						if fn == nil {
							continue
						}
						e := &fnEntry{file: f, name: ft.Name, fn: fn, path: path}
						s.fns[fnKey{path, ft.Name, int(g.ParamCount)}] = e
						s.byName[path+"."+ft.Name] = e
					}
				}
			}
		}
	}
}

// reduceParts are a reduce's callback and collection operands, in any of its
// forms.
func (s *specializer) reduceParts(x *program.Expression, path string, host *program.Function) (fn, coll *program.Operand) {
	ops := x.Operands
	switch len(ops) {
	case 2:
		return ops[0], ops[1]
	case 3:
		if s.callback(ops[0], path, host) != nil {
			return ops[0], ops[1] // reduce(fn, collection, initial)
		}
		return ops[1], ops[2] // reduce(initial, fn, collection)
	}
	return nil, nil
}

// callback is the function op names, if it names one: a function nested in
// host (a lambda not yet hoisted), or one of the package's.
func (s *specializer) callback(op *program.Operand, path string, host *program.Function) *program.Function {
	ref := op.GetReference()
	if ref == nil {
		return nil
	}
	if ref.Prefix == "" && host != nil {
		if fn := nested(host, ref.Name); fn != nil {
			return fn
		}
	}
	p := ref.PkgPath
	if p == "" && ref.Prefix == "" {
		p = path
	}
	if e := s.byName[p+"."+ref.Name]; e != nil {
		return e.fn
	}
	return nil
}

// indexed reports whether the reduce x's callback indexes a list by the
// element's index, directly or through a function it passes it to.
func (s *specializer) indexed(x *program.Expression, path string, host *program.Function) (coll *program.Operand, ok bool) {
	fnOp, coll := s.reduceParts(x, path, host)
	if fnOp == nil {
		return nil, false
	}
	cb := s.callback(fnOp, path, host)
	return coll, cb != nil && len(cb.Parameters) >= 3 && s.indexParamsOf(cb, path)[2]
}

// needsSizes reports whether fn reduces over one of its parameters with an
// indexed callback.
func (s *specializer) needsSizes(fn *program.Function, path string) bool {
	params := map[string]bool{}
	for _, p := range fn.Parameters {
		params[p.Name] = true
	}
	found := false
	eachExpr(fn, func(x *program.Expression) {
		if x.Operation != program.Operation_REDUCE {
			return
		}
		if coll, ok := s.indexed(x, path, fn); ok {
			if ref := coll.GetReference(); ref != nil && ref.Prefix == "" && params[ref.Name] {
				found = true
			}
		}
	})
	return found
}

// cloneRound gives each call to a function that needs sizes a clone for the
// types of its list arguments, and reports whether it made or used any.
func (s *specializer) cloneRound(prog *program.Program) bool {
	need := map[*program.Function]bool{}
	for _, e := range s.fns {
		if !s.sized[e.fn] && s.needsSizes(e.fn, e.path) {
			need[e.fn] = true
		}
	}
	if len(need) == 0 {
		return false
	}
	s.fresh = map[*program.Function]bool{}
	changed := false
	for i, pkg := range pkgs(prog) {
		if pkg == nil {
			continue
		}
		path := pkgPathOf(pkg, i)
		for _, f := range pkg.Files {
			eachCall(f, s.fresh, func(fc *program.FnCall) {
				if s.redirect(fc, path, need) {
					changed = true
				}
			})
		}
	}
	return changed
}

func (s *specializer) redirect(fc *program.FnCall, path string, need map[*program.Function]bool) bool {
	if fc.Reference == nil {
		return false
	}
	p := fc.Reference.PkgPath
	if p == "" && fc.Reference.Prefix == "" {
		p = path
	}
	e := s.fns[fnKey{p, fc.Reference.Name, len(fc.Arguments)}]
	if e == nil || !need[e.fn] {
		return false
	}
	var sig []string
	for _, a := range fc.Arguments {
		if a.Type == nil {
			return false // not typed yet; a later round
		}
		sig = append(sig, typeString(a.Type))
	}
	key := p + "|" + e.name + "|" + strings.Join(sig, ",")
	name, ok := s.clones[key]
	if !ok {
		name = s.freshName(p, e.name)
		s.clones[key] = name
		clone := proto.Clone(e.fn).(*program.Function)
		s.sized[clone] = true
		s.fresh[clone] = true
		e.file.FnTypes = append(e.file.FnTypes, &program.FnType{
			Name:   name,
			Groups: []*program.FnGroup{{ParamCount: int32(len(clone.Parameters)), Functions: []*program.Function{clone}}},
		})
	}
	fc.Reference.Name = name
	fc.Reference.RefPath = ""
	return true
}

func typeString(t *program.Type) string {
	if t == nil {
		return "?"
	}
	s := fmt.Sprintf("%d:%d", t.Kind, t.Size)
	if t.ElemType != nil {
		s += "<" + typeString(t.ElemType) + ">"
	}
	return s
}

func (s *specializer) freshName(path, name string) string {
	for n := 1; ; n++ {
		c := fmt.Sprintf("%s__size%d", name, n)
		if !s.taken[path+"."+c] {
			s.taken[path+"."+c] = true
			return c
		}
	}
}

// regionFns are the functions specialized so far, sorted for a stable walk.
func (s *specializer) regionFns() []*fnEntry {
	var ids []string
	byID := map[string]*fnEntry{}
	for k, e := range s.fns {
		if s.sized[e.fn] {
			id := fmt.Sprintf("%s.%s/%d", k.pkgPath, k.name, k.params)
			ids = append(ids, id)
			byID[id] = e
		}
	}
	sort.Strings(ids)
	out := make([]*fnEntry, len(ids))
	for i, id := range ids {
		out[i] = byID[id]
	}
	return out
}

// unrollReduces writes, in every specialized function, each indexed reduce
// over a collection of known size as nested calls of its callback with
// literal indices: reduce(init, f, xs) is f(f(init, xs[0], 0), xs[1], 1), so
// the callback's index is a constant in each call.
func (s *specializer) unrollReduces() bool {
	changed := false
	for _, e := range s.regionFns() {
		eachOperand(e.fn, func(op *program.Operand, host *program.Function) {
			x := op.GetExpression()
			if x == nil || x.Operation != program.Operation_REDUCE || len(x.Operands) != 3 {
				return
			}
			fnOp, coll := s.reduceParts(x, e.path, host)
			cb := s.callback(fnOp, e.path, host)
			if cb == nil || len(cb.Parameters) != 3 {
				return
			}
			collType := typeOf(coll, host)
			if collType == nil || collType.Size < 1 {
				return
			}
			init := x.Operands[0]
			if fnOp == x.Operands[0] {
				init = x.Operands[2]
			}
			ref := fnOp.GetReference()
			if ref == nil {
				return
			}
			// A callback that is one return, with nothing nested, is written in
			// place, so a constant index reaches the calls it makes.
			// It must read each parameter once: a parameter read twice would copy
			// the accumulator into both places at every step, doubling the code.
			inline := len(cb.Assignments) == 0 && len(cb.Returns) == 1 && len(cb.FnTypes) == 0 && readsEachOnce(cb)
			acc := init
			for i := 0; i < int(collType.Size); i++ {
				elem := program.NewExpression(program.Operation_INDEX, []*program.Operand{
					proto.Clone(coll).(*program.Operand),
					program.NewValue(float64(i), coll.Start, coll.End),
				}, coll.Start, coll.End)
				index := program.NewValue(float64(i), op.Start, op.End)
				if inline {
					acc = substitute(cb.Returns[0], map[string]*program.Operand{
						cb.Parameters[0].Name: acc,
						cb.Parameters[1].Name: elem,
						cb.Parameters[2].Name: index,
					})
					continue
				}
				acc = &program.Operand{Start: op.Start, End: op.End, T: &program.Operand_FnCall{FnCall: &program.FnCall{
					Reference: proto.Clone(ref).(*program.Reference),
					Arguments: []*program.Operand{acc, elem, index},
				}}}
			}
			op.T = acc.T
			changed = true
		})
	}
	return changed
}

// indexParamsOf are the parameters of fn that index a list or vector, in its
// body or nested functions, or that it passes to a function's such parameter.
func (s *specializer) indexParamsOf(fn *program.Function, path string) map[int]bool {
	if s.idxMemo == nil {
		s.idxMemo = map[*program.Function]map[int]bool{}
	}
	if out, ok := s.idxMemo[fn]; ok {
		return out
	}
	out := map[int]bool{}
	s.idxMemo[fn] = out // a call back into fn while it is being worked out reads none
	at := map[string]int{}
	for i, p := range fn.Parameters {
		at[p.Name] = i
	}
	paramOf := func(op *program.Operand) (int, bool) {
		if ref := op.GetReference(); ref != nil && ref.Prefix == "" {
			i, ok := at[ref.Name]
			return i, ok
		}
		return 0, false
	}
	eachOperand(fn, func(op *program.Operand, host *program.Function) {
		if x := op.GetExpression(); x != nil && x.Operation == program.Operation_INDEX && len(x.Operands) >= 2 {
			if i, ok := paramOf(x.Operands[1]); ok {
				out[i] = true
			}
		}
		fc := op.GetFnCall()
		if fc == nil || fc.Reference == nil {
			return
		}
		var callee *program.Function
		if fc.Reference.Prefix == "" {
			callee = nested(host, fc.Reference.Name)
		}
		if callee == nil {
			p := fc.Reference.PkgPath
			if p == "" && fc.Reference.Prefix == "" {
				p = path
			}
			if e := s.fns[fnKey{p, fc.Reference.Name, len(fc.Arguments)}]; e != nil {
				callee = e.fn
			}
		}
		if callee == nil || callee == fn {
			return
		}
		for j := range s.indexParamsOf(callee, path) {
			if j < len(fc.Arguments) {
				if i, ok := paramOf(fc.Arguments[j]); ok {
					out[i] = true
				}
			}
		}
	})
	return out
}

// readsEachOnce reports whether cb's return reads each of its parameters at
// most once.
func readsEachOnce(cb *program.Function) bool {
	count := map[string]int{}
	var walk func(op *program.Operand)
	walk = func(op *program.Operand) {
		if op == nil {
			return
		}
		switch t := op.T.(type) {
		case *program.Operand_Reference:
			if t.Reference != nil && t.Reference.Prefix == "" {
				count[t.Reference.Name]++
			}
		case *program.Operand_Expression:
			if t.Expression != nil {
				for _, o := range t.Expression.Operands {
					walk(o)
				}
			}
		case *program.Operand_FnCall:
			if t.FnCall != nil {
				for _, a := range t.FnCall.Arguments {
					walk(a)
				}
			}
		}
	}
	for _, r := range cb.Returns {
		walk(r)
	}
	for _, p := range cb.Parameters {
		if count[p.Name] > 1 {
			return false
		}
	}
	return true
}

// constEnv is the constant locals a call in fn can read: the ones assigned a
// literal number, the parameters a clone bound among them.
func constEnv(fn *program.Function) map[string]float64 {
	env := map[string]float64{}
	for _, a := range fn.Assignments {
		if a != nil && len(a.Variables) == 1 {
			if v, ok := a.Expression.GetT().(*program.Operand_Value); ok {
				env[a.Variables[0].Name] = v.Value
			}
		}
	}
	return env
}

// bindIndexArgs gives each call, from a specialized function, that passes a
// constant for a parameter its callee indexes with, a clone of the callee
// with that parameter bound, as constspec does for selects. The clone joins
// the specialized functions, so its own calls and reduces are reached in turn.
func (s *specializer) bindIndexArgs() bool {
	changed := false
	for _, e := range s.regionFns() {
		outer := constEnv(e.fn)
		eachCallIn(e.fn, func(fc *program.FnCall, host *program.Function) {
			if fc.Reference == nil {
				return
			}
			p := fc.Reference.PkgPath
			if p == "" && fc.Reference.Prefix == "" {
				p = e.path
			}
			callee := s.fns[fnKey{p, fc.Reference.Name, len(fc.Arguments)}]
			if callee == nil {
				return
			}
			env := outer
			if host != e.fn {
				env = map[string]float64{}
				for k, v := range outer {
					env[k] = v
				}
				for k, v := range constEnv(host) {
					env[k] = v
				}
			}
			bound := map[int]float64{}
			for i := range s.indexParamsOf(callee.fn, p) {
				switch a := fc.Arguments[i].GetT().(type) {
				case *program.Operand_Value:
					bound[i] = a.Value
				case *program.Operand_Reference:
					if a.Reference != nil && a.Reference.Prefix == "" {
						if v, ok := env[a.Reference.Name]; ok {
							bound[i] = v
						}
					}
				}
			}
			if len(bound) == 0 {
				return
			}
			var idx []int
			for i := range bound {
				idx = append(idx, i)
			}
			sort.Ints(idx)
			parts := []string{"bind", p, callee.name, fmt.Sprint(len(fc.Arguments))}
			for _, i := range idx {
				parts = append(parts, fmt.Sprintf("%d=%v", i, bound[i]))
			}
			key := strings.Join(parts, "|")
			name, ok := s.clones[key]
			if !ok {
				name = s.freshName(p, callee.name)
				s.clones[key] = name
				clone := proto.Clone(callee.fn).(*program.Function)
				var params []*program.Parameter
				var locals []*program.Assignment
				for i, prm := range clone.Parameters {
					v, isBound := bound[i]
					if !isBound {
						params = append(params, prm)
						continue
					}
					locals = append(locals, &program.Assignment{
						Start: prm.Start, End: prm.Start,
						Variables:  []*program.Variable{{Name: prm.Name}},
						Expression: program.NewValue(v, prm.Start, prm.Start),
					})
				}
				clone.Parameters = params
				clone.Assignments = append(locals, clone.Assignments...)
				s.sized[clone] = true
				callee.file.FnTypes = append(callee.file.FnTypes, &program.FnType{
					Name:   name,
					Groups: []*program.FnGroup{{ParamCount: int32(len(params)), Functions: []*program.Function{clone}}},
				})
			}
			var args []*program.Operand
			for i, a := range fc.Arguments {
				if _, isBound := bound[i]; !isBound {
					args = append(args, a)
				}
			}
			fc.Arguments = args
			fc.Reference.Name = name
			fc.Reference.RefPath = ""
			changed = true
		})
	}
	return changed
}

// eachOperand calls visit on every operand in fn's body, innermost last, with
// the function (fn or one nested in it) whose body holds it.
func eachOperand(fn *program.Function, visit func(op *program.Operand, host *program.Function)) {
	var walk func(op *program.Operand, host *program.Function)
	walk = func(op *program.Operand, host *program.Function) {
		if op == nil {
			return
		}
		switch t := op.T.(type) {
		case *program.Operand_Expression:
			if t.Expression != nil {
				for _, o := range t.Expression.Operands {
					walk(o, host)
				}
			}
		case *program.Operand_FnCall:
			if t.FnCall != nil {
				for _, a := range t.FnCall.Arguments {
					walk(a, host)
				}
			}
		}
		visit(op, host)
	}
	var fnWalk func(f *program.Function)
	fnWalk = func(f *program.Function) {
		for _, a := range f.Assignments {
			if a != nil {
				walk(a.Expression, f)
			}
		}
		for _, r := range f.Returns {
			walk(r, f)
		}
		for _, ft := range f.FnTypes {
			for _, g := range ft.GetGroups() {
				for _, inner := range g.GetFunctions() {
					if inner != nil {
						fnWalk(inner)
					}
				}
			}
		}
	}
	fnWalk(fn)
}

// eachCallIn calls visit on every call in fn's body, with its host function.
func eachCallIn(fn *program.Function, visit func(fc *program.FnCall, host *program.Function)) {
	eachOperand(fn, func(op *program.Operand, host *program.Function) {
		if fc := op.GetFnCall(); fc != nil {
			visit(fc, host)
		}
	})
}

// eachExpr calls visit on every expression in fn's body.
func eachExpr(fn *program.Function, visit func(*program.Expression)) {
	var walk func(op *program.Operand)
	walk = func(op *program.Operand) {
		if op == nil {
			return
		}
		switch t := op.T.(type) {
		case *program.Operand_Expression:
			if t.Expression != nil {
				visit(t.Expression)
				for _, o := range t.Expression.Operands {
					walk(o)
				}
			}
		case *program.Operand_FnCall:
			if t.FnCall != nil {
				for _, a := range t.FnCall.Arguments {
					walk(a)
				}
			}
		}
	}
	for _, a := range fn.Assignments {
		if a != nil {
			walk(a.Expression)
		}
	}
	for _, r := range fn.Returns {
		walk(r)
	}
	for _, ft := range fn.FnTypes {
		for _, g := range ft.GetGroups() {
			for _, inner := range g.GetFunctions() {
				if inner != nil {
					eachExpr(inner, visit)
				}
			}
		}
	}
}

// eachCall calls visit on every call in the file.
func eachCall(f *program.PackageFile, skip map[*program.Function]bool, visit func(*program.FnCall)) {
	if f == nil {
		return
	}
	var walk func(op *program.Operand)
	walk = func(op *program.Operand) {
		if op == nil {
			return
		}
		switch t := op.T.(type) {
		case *program.Operand_Expression:
			if t.Expression != nil {
				for _, o := range t.Expression.Operands {
					walk(o)
				}
			}
		case *program.Operand_FnCall:
			if t.FnCall != nil {
				for _, a := range t.FnCall.Arguments {
					walk(a)
				}
				visit(t.FnCall)
			}
		}
	}
	var fnWalk func(fn *program.Function)
	fnWalk = func(fn *program.Function) {
		if fn == nil || skip[fn] {
			return
		}
		for _, a := range fn.Assignments {
			if a != nil {
				walk(a.Expression)
			}
		}
		for _, r := range fn.Returns {
			walk(r)
		}
		for _, ft := range fn.FnTypes {
			for _, g := range ft.GetGroups() {
				for _, inner := range g.GetFunctions() {
					fnWalk(inner)
				}
			}
		}
	}
	for _, a := range f.Assignments {
		if a != nil {
			walk(a.Expression)
		}
	}
	for _, r := range f.Returns {
		walk(r)
	}
	for i := 0; i < len(f.FnTypes); i++ {
		for _, g := range f.FnTypes[i].GetGroups() {
			for _, fn := range g.GetFunctions() {
				fnWalk(fn)
			}
		}
	}
}

// nested is the function named name declared inside fn, at any depth.
func nested(fn *program.Function, name string) *program.Function {
	for _, ft := range fn.FnTypes {
		for _, g := range ft.GetGroups() {
			for _, inner := range g.GetFunctions() {
				if inner == nil {
					continue
				}
				if ft.Name == name {
					return inner
				}
				if f := nested(inner, name); f != nil {
					return f
				}
			}
		}
	}
	return nil
}

// substitute is a copy of op with each reference to a name in args replaced by
// a copy of its argument.
func substitute(op *program.Operand, args map[string]*program.Operand) *program.Operand {
	out := proto.Clone(op).(*program.Operand)
	var walk func(o *program.Operand)
	walk = func(o *program.Operand) {
		if o == nil {
			return
		}
		switch t := o.T.(type) {
		case *program.Operand_Reference:
			if r := t.Reference; r != nil && r.Prefix == "" {
				if a, ok := args[r.Name]; ok {
					o.T = proto.Clone(a).(*program.Operand).T
					o.Type = a.Type
				}
			}
		case *program.Operand_Expression:
			if t.Expression != nil {
				for _, x := range t.Expression.Operands {
					walk(x)
				}
			}
		case *program.Operand_FnCall:
			if t.FnCall != nil {
				for _, x := range t.FnCall.Arguments {
					walk(x)
				}
			}
		}
	}
	walk(out)
	return out
}

// typeOf is op's type: its own, or, for a name, the type its parameter or
// local in host was given.
func typeOf(op *program.Operand, host *program.Function) *program.Type {
	if op.Type != nil {
		return op.Type
	}
	ref := op.GetReference()
	if ref == nil || ref.Prefix != "" || host == nil {
		return nil
	}
	for _, p := range host.Parameters {
		if p.Name == ref.Name {
			return p.Type
		}
	}
	for _, a := range host.Assignments {
		for _, v := range a.GetVariables() {
			if v.Name == ref.Name {
				return v.Type
			}
		}
	}
	return nil
}

// clearTopLevelTypes drops the types of every top-level declaration and
// return expression in prog, operands nested in them included.
func clearTopLevelTypes(prog *program.Program) {
	var clear func(op *program.Operand)
	clear = func(op *program.Operand) {
		if op == nil {
			return
		}
		op.Type = nil
		switch t := op.T.(type) {
		case *program.Operand_Expression:
			if t.Expression != nil {
				for _, o := range t.Expression.Operands {
					clear(o)
				}
			}
		case *program.Operand_FnCall:
			if t.FnCall != nil {
				for _, a := range t.FnCall.Arguments {
					clear(a)
				}
			}
		}
	}
	for _, pkg := range pkgs(prog) {
		if pkg == nil {
			continue
		}
		for _, f := range pkg.Files {
			if f == nil {
				continue
			}
			for _, a := range f.Assignments {
				if a != nil {
					clear(a.Expression)
				}
			}
			for _, r := range f.Returns {
				clear(r)
			}
		}
	}
}
