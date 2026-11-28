// Package listlen builds the lists the list builtin asks for:
//
//	list(n)      is [0, 1, ..., n - 1]
//	list(n, fn)  is map(fn, [0, 1, ..., n - 1])
//
// A list's length is part of its type, so n must be known when the document
// compiles: a literal, arithmetic on constants, a local or top-level constant,
// or a parameter that every call passes a constant for. A function whose
// parameter decides a list's length, directly or by passing it on to another
// such function, is a template: each call is given a CLONE with that
// parameter bound to the constant it passes (once per distinct value, as
// constspec does for selects), and the template itself is dropped. A call
// that does not pass a constant is an error naming the parameter.
//
// The pass runs before the closure pass, so a list(n, fn) becomes an ordinary
// map over a literal list before closures are specialized.
package listlen

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/Gradient-Control-Laboratories/omega-core/program"
	"google.golang.org/protobuf/proto"
)

// Transform builds every list(...) in prog and reports whether it changed
// anything.
func Transform(prog *program.Program) (bool, error) {
	if prog == nil {
		return false, nil
	}
	pkgs := append([]*program.Package{prog.Primary}, prog.Dependencies...)
	if !mentionsList(pkgs) {
		return false, nil
	}
	b := &builder{
		fns:     map[fnKey]*fnEntry{},
		consts:  map[string]map[string]float64{},
		lenArgs: map[*program.Function]map[int]bool{},
		clones:  map[string]string{},
		taken:   map[string]bool{},
	}
	for i, pkg := range pkgs {
		if pkg != nil {
			b.index(pkg, pkgPathOf(pkg, i))
		}
	}
	b.findLengthParams()
	for i, pkg := range pkgs {
		if pkg != nil {
			b.rewritePackage(pkg, pkgPathOf(pkg, i))
		}
	}
	if len(b.errs) > 0 {
		return b.modified, b.errs[0]
	}
	b.dropTemplates(pkgs)
	return b.modified, nil
}

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

type builder struct {
	fns map[fnKey]*fnEntry
	// consts are each package's top-level constants.
	consts map[string]map[string]float64
	// lenArgs are, for a template, the parameters that decide a list length.
	lenArgs  map[*program.Function]map[int]bool
	clones   map[string]string
	taken    map[string]bool
	errs     []error
	modified bool
}

// mentionsList reports whether any package calls list, so that a document
// without one costs a walk of nothing.
func mentionsList(pkgs []*program.Package) bool {
	found := false
	for _, pkg := range pkgs {
		if pkg == nil {
			continue
		}
		for _, f := range pkg.Files {
			walkFile(f, func(fc *program.FnCall) {
				if isList(fc) {
					found = true
				}
			})
		}
	}
	return found
}

func isList(fc *program.FnCall) bool {
	return fc != nil && fc.Reference != nil && fc.Reference.Prefix == "" && fc.Reference.Name == "list" &&
		(fc.Reference.PkgPath == "" || fc.Reference.PkgPath == "builtin")
}

// index records every top-level function overload and each package's
// top-level constants.
func (b *builder) index(pkg *program.Package, pkgPath string) {
	consts := map[string]float64{}
	b.consts[pkgPath] = consts
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, ft := range f.FnTypes {
			if ft == nil {
				continue
			}
			b.taken[pkgPath+"."+ft.Name] = true
			for _, g := range ft.Groups {
				for _, fn := range g.GetFunctions() {
					if fn != nil {
						b.fns[fnKey{pkgPath, ft.Name, int(g.ParamCount)}] = &fnEntry{file: f, name: ft.Name, fn: fn}
					}
				}
			}
		}
	}
	// Top-level constants may refer to each other in any order.
	for changed := true; changed; {
		changed = false
		for _, f := range pkg.Files {
			for _, a := range f.GetAssignments() {
				if a == nil || len(a.Variables) != 1 {
					continue
				}
				name := a.Variables[0].Name
				if _, done := consts[name]; done {
					continue
				}
				if v, ok := b.eval(a.Expression, &env{vars: consts}, pkgPath); ok {
					consts[name] = v
					changed = true
				}
			}
		}
	}
}

// env holds the constants a function body can read: its bound parameters and
// constant locals, then its enclosing function's, then the package's.
type env struct {
	vars   map[string]float64
	params map[string]int // parameter name -> index, for templates
	parent *env
}

func (e *env) lookup(name string) (float64, bool) {
	for ; e != nil; e = e.parent {
		if v, ok := e.vars[name]; ok {
			return v, true
		}
	}
	return 0, false
}

// eval is op's value if it is a compile-time scalar constant.
func (b *builder) eval(op *program.Operand, e *env, pkgPath string) (float64, bool) {
	if op == nil {
		return 0, false
	}
	switch t := op.T.(type) {
	case *program.Operand_Value:
		return t.Value, true
	case *program.Operand_Reference:
		r := t.Reference
		if r == nil {
			return 0, false
		}
		if r.Prefix == "" {
			return e.lookup(r.Name)
		}
		if path := r.PkgPath; path != "" {
			v, ok := b.consts[path][r.Name]
			return v, ok
		}
	case *program.Operand_Expression:
		x := t.Expression
		if x == nil {
			return 0, false
		}
		args := make([]float64, len(x.Operands))
		for i, o := range x.Operands {
			v, ok := b.eval(o, e, pkgPath)
			if !ok {
				return 0, false
			}
			args[i] = v
		}
		return apply(x.Operation, args)
	}
	return 0, false
}

func apply(op program.Operation, a []float64) (float64, bool) {
	switch {
	case op == program.Operation_ADD && len(a) == 2:
		return a[0] + a[1], true
	case op == program.Operation_SUB && len(a) == 2:
		return a[0] - a[1], true
	case op == program.Operation_MUL && len(a) == 2:
		return a[0] * a[1], true
	case op == program.Operation_DIV && len(a) == 2 && a[1] != 0:
		return a[0] / a[1], true
	case op == program.Operation_MODULO && len(a) == 2 && a[1] != 0:
		return a[0] - a[1]*math.Floor(a[0]/a[1]), true
	case op == program.Operation_NEG && len(a) == 1:
		return -a[0], true
	case op == program.Operation_FLOOR && len(a) == 1:
		return math.Floor(a[0]), true
	case op == program.Operation_CEIL && len(a) == 1:
		return math.Ceil(a[0]), true
	case op == program.Operation_ROUND && len(a) == 1:
		return math.Round(a[0]), true
	case op == program.Operation_ABS && len(a) == 1:
		return math.Abs(a[0]), true
	case op == program.Operation_MIN && len(a) >= 1:
		m := a[0]
		for _, v := range a[1:] {
			m = math.Min(m, v)
		}
		return m, true
	case op == program.Operation_MAX && len(a) >= 1:
		m := a[0]
		for _, v := range a[1:] {
			m = math.Max(m, v)
		}
		return m, true
	}
	return 0, false
}

// paramDeps are the parameters op reads, directly or through the locals in
// locals (each local's own parameters).
func paramDeps(op *program.Operand, params map[string]int, locals map[string]map[int]bool, out map[int]bool) {
	walkOperand(op, func(o *program.Operand) {
		ref, ok := o.T.(*program.Operand_Reference)
		if !ok || ref.Reference == nil || ref.Reference.Prefix != "" {
			return
		}
		name := ref.Reference.Name
		if deps, ok := locals[name]; ok {
			for i := range deps {
				out[i] = true
			}
			return
		}
		if i, ok := params[name]; ok {
			out[i] = true
		}
	}, nil)
}

// findLengthParams marks, for every function, the parameters that decide the
// length of a list in its body or of one in a function it passes them to.
func (b *builder) findLengthParams() {
	for changed := true; changed; {
		changed = false
		for key, e := range b.fns {
			fn := e.fn
			params := map[string]int{}
			for i, p := range fn.Parameters {
				params[p.Name] = i
			}
			locals := map[string]map[int]bool{}
			for _, a := range fn.Assignments {
				if a == nil {
					continue
				}
				deps := map[int]bool{}
				paramDeps(a.Expression, params, locals, deps)
				for _, v := range a.Variables {
					locals[v.Name] = deps
				}
			}
			mark := func(deps map[int]bool) {
				for i := range deps {
					if b.lenArgs[fn] == nil {
						b.lenArgs[fn] = map[int]bool{}
					}
					if !b.lenArgs[fn][i] {
						b.lenArgs[fn][i] = true
						changed = true
					}
				}
			}
			walkFunction(fn, func(fc *program.FnCall) {
				if isList(fc) && len(fc.Arguments) >= 1 {
					deps := map[int]bool{}
					paramDeps(fc.Arguments[0], params, locals, deps)
					mark(deps)
					return
				}
				if callee := b.callee(fc, key.pkgPath); callee != nil {
					for i := range b.lenArgs[callee.fn] {
						if i < len(fc.Arguments) {
							deps := map[int]bool{}
							paramDeps(fc.Arguments[i], params, locals, deps)
							mark(deps)
						}
					}
				}
			})
		}
	}
}

func (b *builder) callee(fc *program.FnCall, pkgPath string) *fnEntry {
	if fc == nil || fc.Reference == nil || isList(fc) {
		return nil
	}
	path := fc.Reference.PkgPath
	if path == "" && fc.Reference.Prefix == "" {
		path = pkgPath
	}
	return b.fns[fnKey{path, fc.Reference.Name, len(fc.Arguments)}]
}

// rewritePackage specializes the calls to templates in pkg and builds every
// list, walking clones as they are added. Templates themselves are skipped:
// their clones are walked instead, and they are dropped afterwards.
func (b *builder) rewritePackage(pkg *program.Package, pkgPath string) {
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		top := &env{vars: b.consts[pkgPath]}
		for _, a := range f.Assignments {
			if a == nil {
				continue
			}
			b.operand(a.Expression, top, pkgPath)
			for _, v := range a.Variables {
				for _, prop := range v.GetProperties() {
					b.operand(prop.Expression, top, pkgPath)
				}
			}
		}
		for _, r := range f.Returns {
			b.operand(r, top, pkgPath)
		}
		for i := 0; i < len(f.FnTypes); i++ { // clones appended mid-walk are walked too
			ft := f.FnTypes[i]
			for _, g := range ft.GetGroups() {
				for _, fn := range g.GetFunctions() {
					if fn != nil && len(b.lenArgs[fn]) == 0 {
						b.function(fn, top, pkgPath)
					}
				}
			}
		}
	}
}

// function walks fn's body with its constant locals in scope.
func (b *builder) function(fn *program.Function, parent *env, pkgPath string) {
	e := &env{vars: map[string]float64{}, parent: parent}
	for _, a := range fn.Assignments {
		if a == nil {
			continue
		}
		b.operand(a.Expression, e, pkgPath)
		if len(a.Variables) == 1 {
			if v, ok := b.eval(a.Expression, e, pkgPath); ok {
				e.vars[a.Variables[0].Name] = v
			}
		}
	}
	for _, r := range fn.Returns {
		b.operand(r, e, pkgPath)
	}
	// Nested functions and lambdas see this body's constants.
	for _, ft := range fn.FnTypes {
		for _, g := range ft.GetGroups() {
			for _, inner := range g.GetFunctions() {
				if inner != nil {
					b.function(inner, e, pkgPath)
				}
			}
		}
	}
}

// operand rewrites the calls under op, innermost first.
func (b *builder) operand(op *program.Operand, e *env, pkgPath string) {
	if op == nil {
		return
	}
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		if t.Expression != nil {
			for _, o := range t.Expression.Operands {
				b.operand(o, e, pkgPath)
			}
		}
	case *program.Operand_FnCall:
		fc := t.FnCall
		if fc == nil {
			return
		}
		for _, a := range fc.Arguments {
			b.operand(a, e, pkgPath)
		}
		if isList(fc) {
			b.buildList(op, fc, e, pkgPath)
			return
		}
		b.specialize(op, fc, e, pkgPath)
	}
}

// buildList replaces op, a call list(n[, fn]), with the list it asks for.
func (b *builder) buildList(op *program.Operand, fc *program.FnCall, e *env, pkgPath string) {
	n, ok := b.eval(fc.Arguments[0], e, pkgPath)
	if !ok {
		b.errs = append(b.errs, locate(op, fmt.Errorf("list's length must be known when the document compiles: a literal, a constant, or a parameter every call passes a constant for")))
		return
	}
	if n < 1 || n != math.Trunc(n) {
		b.errs = append(b.errs, locate(op, fmt.Errorf("list's length must be a whole number of at least 1, not %v", n)))
		return
	}
	items := make([]*program.Operand, int(n))
	for i := range items {
		items[i] = program.NewValue(float64(i), op.Start, op.End)
	}
	built := program.NewExpression(program.Operation_COLLECT, items, op.Start, op.End)
	if len(fc.Arguments) == 2 {
		built = program.NewExpression(program.Operation_MAP, []*program.Operand{fc.Arguments[1], built}, op.Start, op.End)
	}
	op.T = built.T
	b.modified = true
}

// specialize redirects fc, a call to a template, to a clone with its
// length-deciding arguments bound to the constants it passes.
func (b *builder) specialize(op *program.Operand, fc *program.FnCall, e *env, pkgPath string) {
	callee := b.callee(fc, pkgPath)
	if callee == nil || len(b.lenArgs[callee.fn]) == 0 {
		return
	}
	path := fc.Reference.PkgPath
	if path == "" && fc.Reference.Prefix == "" {
		path = pkgPath
	}
	var idx []int
	for i := range b.lenArgs[callee.fn] {
		idx = append(idx, i)
	}
	sort.Ints(idx)
	bound := map[int]float64{}
	for _, i := range idx {
		v, ok := b.eval(fc.Arguments[i], e, pkgPath)
		if !ok {
			b.errs = append(b.errs, locate(op, fmt.Errorf("%s needs a constant for its parameter %s, which decides the length of a list; this call passes one only known at run time",
				callee.name, callee.fn.Parameters[i].Name)))
			return
		}
		bound[i] = v
	}
	parts := []string{path, callee.name, fmt.Sprint(len(fc.Arguments))}
	for _, i := range idx {
		parts = append(parts, fmt.Sprintf("%d=%v", i, bound[i]))
	}
	key := strings.Join(parts, "|")
	name, ok := b.clones[key]
	if !ok {
		name = b.freshName(path, callee.name)
		b.clones[key] = name
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
				Start:      prm.Start,
				End:        prm.Start,
				Variables:  []*program.Variable{{Name: prm.Name}},
				Expression: program.NewValue(v, prm.Start, prm.Start),
			})
		}
		clone.Parameters = params
		clone.Assignments = append(locals, clone.Assignments...)
		callee.file.FnTypes = append(callee.file.FnTypes, &program.FnType{
			Name:   name,
			Groups: []*program.FnGroup{{ParamCount: int32(len(params)), Functions: []*program.Function{clone}}},
		})
		b.fns[fnKey{path, name, len(params)}] = &fnEntry{file: callee.file, name: name, fn: clone}
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
	b.modified = true
}

func (b *builder) freshName(path, name string) string {
	for n := 1; ; n++ {
		c := fmt.Sprintf("%s__len%d", name, n)
		if !b.taken[path+"."+c] {
			b.taken[path+"."+c] = true
			return c
		}
	}
}

// dropTemplates removes every template: each call to one now calls a clone.
func (b *builder) dropTemplates(pkgs []*program.Package) {
	for _, pkg := range pkgs {
		if pkg == nil {
			continue
		}
		for _, f := range pkg.Files {
			if f == nil {
				continue
			}
			var kept []*program.FnType
			for _, ft := range f.FnTypes {
				var groups []*program.FnGroup
				for _, g := range ft.GetGroups() {
					var fns []*program.Function
					for _, fn := range g.GetFunctions() {
						if len(b.lenArgs[fn]) == 0 {
							fns = append(fns, fn)
						}
					}
					if len(fns) > 0 {
						g.Functions = fns
						groups = append(groups, g)
					}
				}
				if len(groups) > 0 {
					ft.Groups = groups
					kept = append(kept, ft)
				}
			}
			if len(kept) != len(f.FnTypes) {
				f.FnTypes = kept
				b.modified = true
			}
		}
	}
}

func locate(op *program.Operand, err error) error {
	if op != nil && op.Start != nil {
		return fmt.Errorf("%d:%d: %w", op.Start.Line, op.Start.Column, err)
	}
	return err
}

// walkFile calls visit on every call in the file, function bodies included.
func walkFile(f *program.PackageFile, visit func(*program.FnCall)) {
	if f == nil {
		return
	}
	for _, a := range f.Assignments {
		if a != nil {
			walkOperand(a.Expression, nil, visit)
		}
	}
	for _, r := range f.Returns {
		walkOperand(r, nil, visit)
	}
	for _, ft := range f.FnTypes {
		for _, g := range ft.GetGroups() {
			for _, fn := range g.GetFunctions() {
				walkFunction(fn, visit)
			}
		}
	}
}

func walkFunction(fn *program.Function, visit func(*program.FnCall)) {
	if fn == nil {
		return
	}
	for _, a := range fn.Assignments {
		if a != nil {
			walkOperand(a.Expression, nil, visit)
		}
	}
	for _, r := range fn.Returns {
		walkOperand(r, nil, visit)
	}
	for _, ft := range fn.FnTypes {
		for _, g := range ft.GetGroups() {
			for _, inner := range g.GetFunctions() {
				walkFunction(inner, visit)
			}
		}
	}
}

// walkOperand calls each operand visitor on op and below, and visit on calls.
func walkOperand(op *program.Operand, each func(*program.Operand), visit func(*program.FnCall)) {
	if op == nil {
		return
	}
	if each != nil {
		each(op)
	}
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		if t.Expression != nil {
			for _, o := range t.Expression.Operands {
				walkOperand(o, each, visit)
			}
		}
	case *program.Operand_FnCall:
		if t.FnCall != nil {
			if visit != nil {
				visit(t.FnCall)
			}
			for _, a := range t.FnCall.Arguments {
				walkOperand(a, each, visit)
			}
		}
	}
}
