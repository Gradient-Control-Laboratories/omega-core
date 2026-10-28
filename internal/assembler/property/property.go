// Package property implements self-computing properties over the program Operand
// tree: quantities that are derived by propagating a per-operation rule through
// an expression, such as spatial/temporal derivatives (.dp / .dtime).
//
// Each property is a forward abstract interpretation: a Domain assigns a value
// (in some abstract sense — a derivative dual, an interval, a color) to every
// leaf and defines how each operation transfers that value to its result. The
// generic Engine in this file owns everything that is the same across
// properties — tree traversal, name/scope resolution, remap (coordinate
// change), and function evaluation — so a new property only implements its
// Domain.
//
// The result of evaluating in a Domain is ordinary GSL (program Operands encoded
// into the scalar/vector values the backends understand), so it flows through
// Fidget and WGSL unchanged.
package property

import (
	"fmt"

	"gradientcontrol.com/omega/core/program"
)

// Value is a property's abstract value for a single operand. Each Domain casts
// it to its own concrete representation (e.g. a derivative dual or an interval).
type Value any

// Domain is a forward abstract interpretation over the Operand tree. It knows
// the property-value of the leaves and how each operation and builtin transfers
// the property from operands to result. The Engine drives the traversal and
// calls these methods.
type Domain interface {
	// Literal is the value of a numeric constant.
	Literal(v float64) Value

	// ImpliedVar is the value of an implied variable: p, time, x, y, or z.
	ImpliedVar(name string) (Value, error)

	// Opaque is the value of a name the property treats as given — a slider
	// input or an unresolved global — i.e. a constant in this property.
	Opaque(name string) Value

	// Op transfers a numeric/logical operation (ADD, MUL, MIN, SIN, …) to its
	// result given the operands' values.
	Op(operation program.Operation, args []Value) (Value, error)

	// Builtin transfers a builtin function call (length, dot, normalize, …).
	Builtin(name string, args []Value) (Value, error)

	// Collect builds a vector value from component values.
	Collect(args []Value) Value

	// Access applies a swizzle (the selector operands, e.g. "xy") to a value.
	Access(base Value, selectors []*program.Operand) Value

	// Index applies a runtime index (the index operand) to a value.
	Index(base Value, index *program.Operand) Value

	// Select transfers a ternary: cond plus 2 or 3 branch values.
	Select(cond Value, branches []Value) Value

	// Named adjusts a resolved declaration's value so its value-part refers to
	// the live variable `name` (which is in scope at the emit site) rather than
	// an inlined copy — this keeps slider/uniform references live and shrinks
	// the output. It is applied only for declarations in a nameable scope
	// (top-level and companion bodies), never for inlined function locals.
	Named(name string, value Value) Value

	// Operator expands a property-specific operator (e.g. DERIV) to a plain
	// GSL operand. The Engine evaluates the result in the current Domain, so
	// operators compose. The bool is false when operation is not an operator
	// for this Domain.
	Operator(operation program.Operation, operands []*program.Operand) (*program.Operand, bool, error)
}

// CallDomain is an optional Domain extension: a Domain may intercept a
// user-function call before the Engine inlines it, returning (value, true, nil)
// to handle it or (nil, false, nil) to fall through. The color gradient uses it
// to read a shape's color at the coordinate where the shape is evaluated (so a
// procedural color transports through remap with the shape).
type CallDomain interface {
	Call(name string, args []Value) (Value, bool, error)
}

// OpaqueDomain is an optional Domain extension for treating cross-package
// (qualified) calls and references — e.g. map.triangleWave, noise.cellularEdges —
// as opaque givens rather than an error. The color gradient uses it: it only
// needs the value spliced (the backend resolves the import when it lowers) and
// does not differentiate through them.
//
// Implementing it is also the choice NOT to resolve the import: a Domain that
// declares how to treat a cross-package value keeps that treatment, while one
// that does not gets the imported function's body inlined (see Engine.imports),
// which is what a derivative needs — map.rotateZX(p, ang) depends on p, so
// splicing it with no tangent would silently zero the chain rule through it.
type OpaqueDomain interface {
	OpaqueCall(pkgName, name string, args []Value) Value
	OpaqueRef(pkgName, name string) Value
}

// QualifiedCallDomain is an optional Domain extension: a Domain may intercept a
// cross-package call (transform.translate, …) before the Engine inlines the
// library's body, returning (value, true, nil) to handle it or (nil, false,
// nil) to fall through. path is the package's import path, whatever prefix the
// document imported it under.
type QualifiedCallDomain interface {
	QualifiedCall(path, name string, args []Value) (Value, bool, error)
}

// RemapDomain is an optional Domain extension: a Domain may adjust the value of
// remap(field, q) after the Engine has evaluated the field under the rebound p.
// Without it a property passes straight through a remap, which is right for a
// derivative (the chain rule is in the rebound values) and wrong for a
// bounding box, which describes p itself and so means nothing after the
// coordinate changes.
type RemapDomain interface {
	Remap(target, coord Value) Value
}

// ArgsCallDomain is an optional Domain extension offered an unqualified call
// before any of its arguments is evaluated, with eval to evaluate the ones it
// needs. It is for a call a Domain answers from some arguments alone: the box
// takes a declared box's corners and never evaluates the shape inside it.
type ArgsCallDomain interface {
	ArgsCall(name string, args []*program.Operand, eval func(*program.Operand) (Value, error)) (Value, bool, error)
}

// Engine evaluates operands in a Domain, resolving names against a package's
// declarations and inlining user functions.
type Engine struct {
	dom   Domain
	decls map[string]*program.Operand    // single-name assignment bodies
	funcs map[string][]*program.Function // user function declarations (including overloads)
	// multiDecls holds the names bound by a destructuring assignment
	// (a, b = f()), each paired with the return of that call it stands for.
	// Nil when the scope destructures nothing.
	multiDecls map[string]multiBinding
	// imports maps a qualified import's prefix (map, noise, color, …) to that
	// package's resolvable scope, so a Domain that does not implement
	// OpaqueDomain can differentiate through a library call instead of
	// treating it as a given. Nil when imports were not resolved.
	imports map[string]*namespace
}

// namespace is an imported package's resolvable scope: its top-level
// assignments and functions, plus whatever its own dot-imports contribute.
type namespace struct {
	path       string // the import path, for QualifiedCallDomain
	decls      map[string]*program.Operand
	multiDecls map[string]multiBinding
	funcs      map[string][]*program.Function
}

// NewEngine builds an Engine for the given Domain and package scope.
func NewEngine(dom Domain, decls map[string]*program.Operand, funcs map[string][]*program.Function) *Engine {
	return &Engine{dom: dom, decls: decls, funcs: funcs}
}

// Eval evaluates an operand to its Domain value. Implied variables (p, time,
// x, y, z) resolve through the Domain unless rebound (e.g. by remap).
func (e *Engine) Eval(op *program.Operand) (Value, error) {
	root := &frame{eng: e, vars: map[string]Value{}, decls: e.decls, multiDecls: e.multiDecls, funcs: e.funcs, evaluating: map[string]bool{}, nameable: true}
	return root.eval(op)
}

func isImpliedVar(name string) bool {
	switch name {
	case program.PVarName, program.TimeVarName, program.XVarName, program.YVarName, program.ZVarName:
		return true
	}
	return false
}

// argBinding records the operand a function parameter was bound to, and the
// frame that operand belongs to (the caller's). Two things need the parameter's
// underlying operand rather than just its evaluated Value: map/reduce, to reach
// a list's element operands (a list param carries no per-element Values), and a
// call to a function-valued parameter, to reach the function it names.
type argBinding struct {
	op    *program.Operand
	frame *frame
}

// frame is a lexical scope: bound values (implied vars, params, memoized decls),
// plus the declarations and functions visible here.
type frame struct {
	eng        *Engine
	parent     *frame
	vars       map[string]Value
	decls      map[string]*program.Operand
	multiDecls map[string]multiBinding // names bound by a destructuring assignment
	funcs      map[string][]*program.Function
	argOps     map[string]argBinding // parameter name -> its bound operand
	evaluating map[string]bool
	// nameable is true when this frame's declarations are emitted in the output
	// (top-level package scope, companion bodies), so a reference may keep the
	// name; false for inlined function bodies, whose locals must be inlined.
	nameable bool
	// shared memoizes eval by operand pointer while an operator expansion is
	// evaluated. An expansion (.second, .dp) is a DAG whose subtrees are
	// shared by pointer; walked as a tree, .second under .dp grows with every
	// path through it. Within one frame the same operand evaluates to the same
	// value, so each shared subtree is evaluated once and its value reused,
	// and the output stays a DAG for sharedLocals to name.
	shared map[*program.Operand]Value
	// rebound is true in a remap's frame, where p is the remap's coordinate.
	// There a declaration that varies with p is not named: its live reference
	// is evaluated where it is emitted, at the OUTER p, so naming it would drop
	// the remap (see lookup).
	rebound bool
}

func (f *frame) eval(op *program.Operand) (Value, error) {
	if f.shared == nil {
		return f.evalOperand(op)
	}
	if v, ok := f.shared[op]; ok {
		return v, nil
	}
	v, err := f.evalOperand(op)
	if err == nil {
		f.shared[op] = v
	}
	return v, err
}

func (f *frame) evalOperand(op *program.Operand) (Value, error) {
	if op == nil {
		return nil, fmt.Errorf("nil operand")
	}
	switch t := op.T.(type) {
	case *program.Operand_Value:
		return f.eng.dom.Literal(t.Value), nil
	case *program.Operand_Reference:
		if t.Reference.Prefix != "" {
			if od, ok := f.eng.dom.(OpaqueDomain); ok {
				return od.OpaqueRef(t.Reference.Prefix, t.Reference.Name), nil
			}
			if ns := f.eng.imports[t.Reference.Prefix]; ns != nil {
				if body, ok := ns.decls[t.Reference.Name]; ok {
					return f.namespaceFrame(ns).eval(body)
				}
			}
			return nil, fmt.Errorf("cannot evaluate cross-package reference %q.%q in property %T",
				t.Reference.Prefix, t.Reference.Name, f.eng.dom)
		}
		return f.lookup(t.Reference.Name)
	case *program.Operand_Expression:
		return f.evalExpr(t.Expression)
	case *program.Operand_FnCall:
		return f.evalCall(t.FnCall)
	case *program.Operand_PropRef:
		if t.PropRef.Reference.Prefix != "" {
			return nil, fmt.Errorf("cannot evaluate cross-package reference %q.%q in property %T",
				t.PropRef.Reference.Prefix, t.PropRef.Reference.Name, f.eng.dom)
		}
		return nil, fmt.Errorf("cannot evaluate property reference %q.%q in property %T",
			t.PropRef.Reference.Name, t.PropRef.Property, f.eng.dom)
	default:
		return nil, fmt.Errorf("cannot evaluate operand %T", op.T)
	}
}

// lookup resolves a name, evaluating and memoizing assignment bodies on demand.
// Implied variables and bound params/locals are found in vars; declared names
// (including the names of a destructuring assignment) are evaluated lazily;
// everything else is opaque (a constant in this property).
//
// Each frame answers for its own names, evaluated or not, before its parent
// does. A called function's frame hangs off its caller's, so looking through
// every frame's vars first let a caller's value stand in for a local the callee
// had not evaluated yet: catalog.panCrown's k = nominal read as the k of the
// function that called it. It is also what lets a remap re-derive a named field
// under its own p, rather than take the value its parent already has.
func (f *frame) lookup(name string) (Value, error) {
	for s := f; s != nil; s = s.parent {
		if v, ok := s.vars[name]; ok {
			return v, nil
		}
		body, ok := s.decls[name]
		mb, destructured := s.multiDecls[name]
		if !ok && !destructured {
			continue
		}
		if s.evaluating[name] {
			return nil, fmt.Errorf("cyclic reference through %q", name)
		}
		s.evaluating[name] = true
		var v Value
		var err error
		switch {
		case destructured && s.nameable && isQualifiedCall(mb.value) && isOpaqueDomain(f.eng.dom):
			// A DESTRUCTURED LIBRARY CALL IS A GIVEN, like every other qualified
			// call in a domain that splices them opaquely — which has no way to
			// say "return k of an opaque call" as a value. It does not need one:
			// at package scope the name is live where the output is emitted, so
			// it stands for itself, and a unit's seed still reaches it through
			// Opaque. fastener.setScrewSoftTip's body and pad are the case.
			v = f.eng.dom.Opaque(name)
		case destructured:
			v, err = s.evalReturnK(name, mb)
		default:
			v, err = s.eval(body)
		}
		delete(s.evaluating, name)
		if err != nil {
			// ONE BAD NAME, NOT ONE BAD RETURN. The property-field gradient
			// used to answer any failure by dropping the field of the whole
			// return it sat in, so a single part it could not differentiate
			// left every other part of a board on the default palette.
			// Contained here instead, the failing name is its own live value
			// with the seed as its tangent: exact when it is itself a unit, and
			// when it is a group, only the units inside it lose their weights.
			if _, ok := f.eng.dom.(*gradDomain); !ok || !s.nameable {
				return nil, err
			}
			v = f.eng.dom.Opaque(name)
		}
		// UNDER A REMAP, A NAME THAT VARIES WITH p STAYS ITS VALUE. Named turns a
		// declaration into its live reference, which the backend evaluates at
		// the outer p — so remap(ball, q) read the ball unmoved wherever ball
		// was named inside it: min(near, remap(ball, q)).second and .dp both
		// came out as if the remap were not there, where the same remap written
		// as a declaration was right. The value evaluated here, under the
		// rebound p, is the one the remap means. A name that does not vary with
		// p — a slider, a constant — is the same under any remap and is named
		// as everywhere else, so it stays live.
		if s.nameable {
			if s.rebound && s.declVariesWithP(name, map[string]bool{}) {
				v = unnamed(f.eng.dom, name, v)
			} else {
				v = f.eng.dom.Named(name, v)
			}
		}
		s.vars[name] = v
		return v, nil
	}
	if isImpliedVar(name) {
		return f.eng.dom.ImpliedVar(name)
	}
	return f.eng.dom.Opaque(name), nil
}

func (f *frame) evalAll(ops []*program.Operand) ([]Value, error) {
	out := make([]Value, len(ops))
	for i, op := range ops {
		v, err := f.eval(op)
		if err != nil {
			return nil, err
		}
		out[i] = v
	}
	return out, nil
}

func (f *frame) evalExpr(expr *program.Expression) (Value, error) {
	// Property-specific operators (e.g. DERIV) expand to a plain operand which
	// is then evaluated in the current Domain, so they compose.
	if expanded, ok, err := f.eng.dom.Operator(expr.Operation, expr.Operands); err != nil {
		return nil, err
	} else if ok {
		if f.shared == nil {
			f.shared = map[*program.Operand]Value{}
			defer func() { f.shared = nil }()
		}
		return f.eval(expanded)
	}

	switch expr.Operation {
	case program.Operation_REMAP:
		return f.evalRemap(expr)
	case program.Operation_ACCESS:
		base, err := f.eval(expr.Operands[0])
		if err != nil {
			return nil, err
		}
		return f.eng.dom.Access(base, expr.Operands[1:]), nil
	case program.Operation_INDEX:
		base, err := f.eval(expr.Operands[0])
		if err != nil {
			return nil, err
		}
		return f.eng.dom.Index(base, f.resolveIndexOperand(expr.Operands[1])), nil
	case program.Operation_COLLECT:
		args, err := f.evalAll(expr.Operands)
		if err != nil {
			return nil, err
		}
		return f.eng.dom.Collect(args), nil
	case program.Operation_SELECT:
		args, err := f.evalAll(expr.Operands)
		if err != nil {
			return nil, err
		}
		return f.eng.dom.Select(args[0], args[1:]), nil
	case program.Operation_MAP:
		elems, err := f.mapElements(expr)
		if err != nil {
			return nil, err
		}
		return f.eng.dom.Collect(elems), nil
	case program.Operation_REDUCE:
		return f.reduceElements(expr)
	}

	args, err := f.evalAll(expr.Operands)
	if err != nil {
		return nil, err
	}
	return f.eng.dom.Op(expr.Operation, args)
}

// reduceElements evaluates reduce(init, fn, list) or reduce(fn, list) to its Value.
func (f *frame) reduceElements(expr *program.Expression) (Value, error) {
	var initOp *program.Operand
	var fnOp *program.Operand
	var listOp *program.Operand

	if len(expr.Operands) == 3 {
		initOp, fnOp, listOp = expr.Operands[0], expr.Operands[1], expr.Operands[2]
	} else if len(expr.Operands) == 2 {
		fnOp, listOp = expr.Operands[0], expr.Operands[1]
	} else {
		return nil, fmt.Errorf("reduce must have 2 or 3 arguments, got %d", len(expr.Operands))
	}

	fn, err := f.resolveLambda(fnOp)
	if err != nil {
		return nil, err
	}
	elems, err := f.listElements(listOp)
	if err != nil {
		return nil, err
	}
	var acc Value
	start := 0
	if initOp != nil {
		if acc, err = f.eval(initOp); err != nil {
			return nil, err
		}
	} else {
		if len(elems) == 0 {
			return nil, fmt.Errorf("reduce of an empty list with no initial value")
		}
		acc, start = elems[0], 1
	}
	for i := start; i < len(elems); i++ {
		if acc, err = f.applyLambda(fn, []Value{acc, elems[i]}); err != nil {
			return nil, err
		}
	}
	return acc, nil
}

// mapElements evaluates map(fn, list) to its element Values.
func (f *frame) mapElements(expr *program.Expression) ([]Value, error) {
	if len(expr.Operands) != 2 {
		return nil, fmt.Errorf("map over %d collections not supported", len(expr.Operands)-1)
	}
	fn, err := f.resolveLambda(expr.Operands[0])
	if err != nil {
		return nil, err
	}
	src, err := f.listElements(expr.Operands[1])
	if err != nil {
		return nil, err
	}
	out := make([]Value, len(src))
	for i, s := range src {
		args := []Value{s}
		if len(fn.Parameters) == 2 { // (value, index)
			args = append(args, f.eng.dom.Literal(float64(i)))
		}
		if out[i], err = f.applyLambda(fn, args); err != nil {
			return nil, err
		}
	}
	return out, nil
}

// listElements resolves an operand to the Values of its list elements: a COLLECT
// yields its components, a map yields its mapped elements, a reference follows to
// the operand it was bound to (a fold argument or a declaration), and anything
// else is a length-one list (a scalar is a list of length one).
func (f *frame) listElements(op *program.Operand) ([]Value, error) {
	switch t := op.T.(type) {
	case *program.Operand_Reference:
		if t.Reference.Prefix != "" {
			// A qualified library list (fastener.clearanceTable) is not in any
			// frame here, and the fall-through below would evaluate it as ONE
			// opaque value — so a map over it would hand the whole table to
			// the closure, whose pr.y then reads as member access on the entire
			// list. The rows live in the import's namespace; list them there.
			if ns := f.eng.imports[t.Reference.Prefix]; ns != nil {
				if body, ok := ns.decls[t.Reference.Name]; ok {
					return f.namespaceFrame(ns).listElements(body)
				}
			}
		} else if body, fr, ok := f.resolveRef(t.Reference.Name); ok {
			return fr.listElements(body)
		}
	case *program.Operand_Expression:
		switch t.Expression.Operation {
		case program.Operation_COLLECT:
			return f.evalAll(t.Expression.Operands)
		case program.Operation_MAP:
			return f.mapElements(t.Expression)
		}
	case *program.Operand_FnCall:
		// A list returned by a function call: enumerate the elements of the
		// call's return expression, in the callee's scope. Without this the
		// fall-through below evaluates the call as ONE opaque value, so a map
		// or reduce over it hands the whole list to the closure — the same
		// hazard the qualified-reference case above guards against. The noise
		// library's cellularOffsets() is exactly this shape, which is why
		// differentiating cellular noise failed with a shape error.
		if elems, ok, err := f.callListElements(t.FnCall); err != nil {
			return nil, err
		} else if ok {
			return elems, nil
		}
	case *program.Operand_Value, *program.Operand_PropRef:
		// fall through to single-value eval
	default:
		return nil, fmt.Errorf("internal error: unhandled operand type %T in listElements; please add a new case to this switch", t)
	}
	v, err := f.eval(op)
	if err != nil {
		return nil, err
	}
	return []Value{v}, nil
}

// callListElements enumerates the elements of a list returned by a user
// function call, by inlining the call and listing its return expression in the
// callee's frame. Derived from the qualified-reference case in listElements:
// resolve the list where it lives rather than evaluating the call as one opaque
// value.
//
// ok is false when the call is not ours to resolve — a builtin, an unknown or
// multi-return function, or a call a Domain handles itself. eval and evalFnCall
// consult the Domain BEFORE imports, and this preserves that precedence: the
// colour gradient and the second-closest property intercept calls and still see
// them exactly as they did.
func (f *frame) callListElements(fc *program.FnCall) ([]Value, bool, error) {
	if fc.Reference == nil || program.IsBuiltin(fc.Reference.Name) {
		return nil, false, nil
	}
	scope := f
	if fc.Reference.Prefix != "" {
		if _, opaque := f.eng.dom.(OpaqueDomain); opaque {
			return nil, false, nil
		}
		ns := f.eng.imports[fc.Reference.Prefix]
		if ns == nil {
			return nil, false, nil
		}
		scope = f.namespaceFrame(ns)
	} else if _, intercepts := f.eng.dom.(CallDomain); intercepts {
		return nil, false, nil
	}
	fn := scope.resolveFunc(fc.Reference.Name, len(fc.Arguments))
	if fn == nil || len(fn.Returns) != 1 {
		return nil, false, nil
	}
	args, err := f.evalAll(fc.Arguments)
	if err != nil {
		return nil, false, err
	}
	// A non-list return falls through to single-value eval inside listElements,
	// which is what the old path did for it anyway.
	elems, err := f.callFrame(scope, fc, fn, args).listElements(fn.Returns[0])
	if err != nil {
		return nil, false, err
	}
	return elems, true, nil
}

// resolveRef follows a name to the operand it stands for and the frame in which
// that operand is valid: a fold-argument binding (resolved in the caller's
// frame) or a declaration (resolved in its own frame).
func (f *frame) resolveRef(name string) (*program.Operand, *frame, bool) {
	for s := f; s != nil; s = s.parent {
		if b, ok := s.argOps[name]; ok {
			return b.op, b.frame, true
		}
		if body, ok := s.decls[name]; ok {
			return body, s, true
		}
	}
	return nil, nil, false
}

// resolveLambda resolves a fold's function operand (a reference to a hoisted
// local lambda, or an overloaded function) to its declaration.
func (f *frame) resolveLambda(op *program.Operand) (*program.Function, error) {
	ref, ok := op.T.(*program.Operand_Reference)
	if !ok {
		return nil, fmt.Errorf("fold function must be a reference, got %T", op.T)
	}
	for s := f; s != nil; s = s.parent {
		if fns, ok := s.funcs[ref.Reference.Name]; ok && len(fns) > 0 {
			return fns[0], nil
		}
	}
	return nil, fmt.Errorf("unknown fold function %q", ref.Reference.Name)
}

// resolveIndexOperand walks an index operand back to the operand it names.
//
// Index takes an *operand* rather than a Value, because a backend indexes a
// collection by a compile-time constant and the operand is what gets emitted.
// That is right at package scope, where the name it references is emitted
// alongside it — and wrong inside an INLINED function body, whose locals are
// not emitted at all (see frame.nameable). Such a reference reaches the output
// naming something that does not exist there, and type analysis rejects it.
//
// So follow the binding: a parameter back through argOps to the argument
// written at the call site, a local back through decls to its body, until a
// literal falls out or a nameable frame owns the name. A table lookup that
// takes its column as a parameter — at(rows, key, col, halfWidth) — is the
// case that found this.
func (f *frame) resolveIndexOperand(op *program.Operand) *program.Operand {
	fr := f
	// A binding chain is short; the bound guards against a cycle rather than
	// standing for any real depth.
	for depth := 0; depth < 32; depth++ {
		ref, ok := op.T.(*program.Operand_Reference)
		if !ok || ref.Reference == nil || ref.Reference.Prefix != "" {
			return op
		}
		name := ref.Reference.Name
		next, nextFrame := (*program.Operand)(nil), fr
		for s := fr; s != nil; s = s.parent {
			if b, bound := s.argOps[name]; bound {
				next, nextFrame = b.op, b.frame
				break
			}
			if body, declared := s.decls[name]; declared {
				if s.nameable {
					// The declaration is emitted too, so the name resolves.
					return op
				}
				next, nextFrame = body, s
				break
			}
		}
		if next == nil {
			return op
		}
		op, fr = next, nextFrame
	}
	return op
}

// applyLambda evaluates a lambda body with its parameters bound to the given
// Values. The child frame's parent is f, so the lambda closes over the enclosing
// scope's locals.
func (f *frame) applyLambda(fn *program.Function, args []Value) (Value, error) {
	if len(fn.Returns) != 1 {
		return nil, fmt.Errorf("fold lambda must have exactly one return")
	}
	child := &frame{
		eng:        f.eng,
		parent:     f,
		vars:       map[string]Value{},
		decls:      collectAssignments(fn.Assignments),
		multiDecls: collectMultiAssignments(fn.Assignments),
		funcs:      collectFunctions(fn.FnTypes),
		argOps:     map[string]argBinding{},
		evaluating: map[string]bool{},
	}
	for i, param := range fn.Parameters {
		if i < len(args) {
			child.vars[param.Name] = args[i]
		}
	}
	return child.eval(fn.Returns[0])
}

// evalRemap evaluates remap(field, q): the field is evaluated with the point p
// rebound to q's value, so the property propagates through the coordinate
// transform via the chain rule (handled by the Domain's own rules).
func (f *frame) evalRemap(expr *program.Expression) (Value, error) {
	child, err := f.remapFrame(expr)
	if err != nil {
		return nil, err
	}
	v, err := child.eval(expr.Operands[0])
	if err != nil {
		return nil, err
	}
	return child.remapped(v), nil
}

// remapped hands a value evaluated in a remap's child frame to the Domain's
// Remap, when it has one, with the coordinate the frame rebound p to.
func (f *frame) remapped(v Value) Value {
	if rd, ok := f.eng.dom.(RemapDomain); ok {
		return rd.Remap(v, f.vars[program.PVarName])
	}
	return v
}

// remapFrame builds the child scope a remap evaluates its target in: p (and
// x/y/z) rebound to the coordinate operand's value, everything else inherited.
func (f *frame) remapFrame(expr *program.Expression) (*frame, error) {
	q, err := f.eval(expr.Operands[1])
	if err != nil {
		return nil, err
	}
	child := &frame{
		eng:    f.eng,
		parent: f,
		vars: map[string]Value{
			program.PVarName: q,
			program.XVarName: f.eng.dom.Access(q, []*program.Operand{ref(program.XVarName)}),
			program.YVarName: f.eng.dom.Access(q, []*program.Operand{ref(program.YVarName)}),
			program.ZVarName: f.eng.dom.Access(q, []*program.Operand{ref(program.ZVarName)}),
		},
		// Inherit the enclosing scope so a named field referenced inside the
		// remap re-derives under the rebound p (fresh vars, so it is not read
		// from the enclosing frame's original-p memo).
		decls:      f.decls,
		multiDecls: f.multiDecls,
		funcs:      f.funcs,
		argOps:     f.argOps,
		evaluating: map[string]bool{},
		nameable:   f.nameable,
		rebound:    true,
	}
	return child, nil
}

// UnnamedDomain is an optional Domain extension for a Domain whose Named does
// more than name the value: the colour gradient seeds a unit's tangent at Named
// and records the coordinate the unit is sampled at. A declaration that varies
// with p inside a remap is not named (see lookup); Unnamed is where such a
// Domain still marks it, keeping the value it was evaluated to.
type UnnamedDomain interface {
	Unnamed(name string, value Value) Value
}

func unnamed(dom Domain, name string, v Value) Value {
	if ud, ok := dom.(UnnamedDomain); ok {
		return ud.Unnamed(name, v)
	}
	return v
}

// declVariesWithP reports whether the declaration name varies with p: its body
// reads p (or x, y, z), calls a function — whose body may read p — reads a
// qualified name, or names a declaration that does. A slider or a constant
// does not. A name bound by a destructuring assignment is taken to vary, the
// safe answer here: not naming a constant only inlines it.
func (f *frame) declVariesWithP(name string, seen map[string]bool) bool {
	if name == program.PVarName || name == program.XVarName || name == program.YVarName || name == program.ZVarName {
		return true
	}
	if seen[name] {
		return false
	}
	seen[name] = true
	if body, ok := f.decls[name]; ok {
		return f.operandVariesWithP(body, seen)
	}
	_, destructured := f.multiDecls[name]
	return destructured
}

func (f *frame) operandVariesWithP(op *program.Operand, seen map[string]bool) bool {
	if op == nil {
		return false
	}
	switch t := op.T.(type) {
	case *program.Operand_Reference:
		if t.Reference == nil {
			return false
		}
		if t.Reference.Prefix != "" {
			return true
		}
		return f.declVariesWithP(t.Reference.Name, seen)
	case *program.Operand_Expression:
		for _, sub := range t.Expression.Operands {
			if f.operandVariesWithP(sub, seen) {
				return true
			}
		}
	case *program.Operand_FnCall:
		return true
	}
	return false
}

func ref(name string) *program.Operand {
	return &program.Operand{
		T: &program.Operand_Reference{
			Reference: &program.Reference{Name: name},
		},
	}
}

func (f *frame) evalCall(fc *program.FnCall) (Value, error) {
	if fc.Reference == nil {
		return nil, fmt.Errorf("nil reference in function call")
	}
	if ad, ok := f.eng.dom.(ArgsCallDomain); ok && fc.Reference.Prefix == "" {
		if v, handled, err := ad.ArgsCall(fc.Reference.Name, fc.Arguments, f.eval); err != nil {
			return nil, err
		} else if handled {
			return v, nil
		}
	}
	args, err := f.evalAll(fc.Arguments)
	if err != nil {
		return nil, err
	}
	if fc.Reference.Prefix != "" {
		if qd, ok := f.eng.dom.(QualifiedCallDomain); ok {
			path := fc.Reference.Prefix
			if ns := f.eng.imports[fc.Reference.Prefix]; ns != nil && ns.path != "" {
				path = ns.path
			}
			if v, handled, err := qd.QualifiedCall(path, fc.Reference.Name, args); err != nil {
				return nil, err
			} else if handled {
				return v, nil
			}
		}
		if od, ok := f.eng.dom.(OpaqueDomain); ok {
			return od.OpaqueCall(fc.Reference.Prefix, fc.Reference.Name, args), nil
		}
		ns := f.eng.imports[fc.Reference.Prefix]
		if ns == nil {
			return nil, fmt.Errorf("cannot evaluate cross-package call %q.%q: package %q is not resolvable here",
				fc.Reference.Prefix, fc.Reference.Name, fc.Reference.Prefix)
		}
		scope := f.namespaceFrame(ns)
		fn := scope.resolveFunc(fc.Reference.Name, len(args))
		if fn == nil {
			return nil, fmt.Errorf("cannot evaluate call to unknown function %q.%q with %d arguments",
				fc.Reference.Prefix, fc.Reference.Name, len(args))
		}
		if len(fn.Returns) != 1 {
			return nil, fmt.Errorf("cannot evaluate multi-return function %q.%q in a property",
				fc.Reference.Prefix, fc.Reference.Name)
		}
		return f.inlineReturnIn(scope, fc, fn, args, 0)
	}
	if program.IsBuiltin(fc.Reference.Name) {
		return f.eng.dom.Builtin(fc.Reference.Name, args)
	}
	if cd, ok := f.eng.dom.(CallDomain); ok {
		if v, handled, err := cd.Call(fc.Reference.Name, args); err != nil {
			return nil, err
		} else if handled {
			return v, nil
		}
	}
	return f.inlineCall(fc, args)
}

// inlineCall inlines a user function: bind its parameters to the argument
// values and evaluate its body in a child scope.
func (f *frame) inlineCall(fc *program.FnCall, args []Value) (Value, error) {
	fn := f.resolveFunc(fc.Reference.Name, len(args))
	if fn == nil {
		return nil, fmt.Errorf("cannot evaluate call to unknown function %q with %d arguments", fc.Reference.Name, len(args))
	}
	if len(fn.Returns) != 1 {
		return nil, fmt.Errorf("cannot evaluate multi-return function %q in a property", fc.Reference.Name)
	}
	return f.inlineReturn(fc, fn, args, 0)
}

// inlineReturn binds fn's parameters to the argument values and evaluates one of
// its returns in that child scope: return 0 for an ordinary call, the name's own
// index for a destructuring assignment.
func (f *frame) inlineReturn(fc *program.FnCall, fn *program.Function, args []Value, index int) (Value, error) {
	return f.inlineReturnIn(f, fc, fn, args, index)
}

// inlineReturnIn is inlineReturn with the body's lexical scope given separately
// from the frame the argument operands belong to: an imported function resolves
// names in its own package, while its arguments were written at the call site.
func (f *frame) inlineReturnIn(scope *frame, fc *program.FnCall, fn *program.Function, args []Value, index int) (Value, error) {
	if len(args) != len(fn.Parameters) {
		return nil, fmt.Errorf("function %q expects %d arguments, got %d", fc.Reference.Name, len(fn.Parameters), len(args))
	}
	if index >= len(fn.Returns) {
		return nil, fmt.Errorf("function %q returns %d values, cannot take return %d", fc.Reference.Name, len(fn.Returns), index+1)
	}
	return f.callFrame(scope, fc, fn, args).eval(fn.Returns[index])
}

// callFrame is the child scope a call's body evaluates in: fn's locals and
// nested functions, with its parameters bound to the argument values (and to
// the argument operands, in the caller's frame, for the fold machinery).
func (f *frame) callFrame(scope *frame, fc *program.FnCall, fn *program.Function, args []Value) *frame {
	child := &frame{
		eng:        f.eng,
		parent:     scope,
		vars:       map[string]Value{},
		decls:      collectAssignments(fn.Assignments),
		multiDecls: collectMultiAssignments(fn.Assignments),
		funcs:      collectFunctions(fn.FnTypes),
		argOps:     make(map[string]argBinding, len(fn.Parameters)),
		evaluating: map[string]bool{},
	}
	for i, param := range fn.Parameters {
		child.vars[param.Name] = args[i]
		child.argOps[param.Name] = argBinding{op: fc.Arguments[i], frame: f}
	}
	return child
}

// namespaceFrame is the lexical scope of an imported package: its declarations
// and functions, parented to the calling frame so the implied variables (p and
// friends, which a remap may have rebound) still resolve to the call site's.
func (f *frame) namespaceFrame(ns *namespace) *frame {
	return &frame{
		eng:        f.eng,
		parent:     f,
		vars:       map[string]Value{},
		decls:      ns.decls,
		multiDecls: ns.multiDecls,
		funcs:      ns.funcs,
		evaluating: map[string]bool{},
	}
}

// evalReturnK evaluates one name of a destructuring assignment (a, b = f()) by
// inlining the call and taking that name's return. Without it the name resolves
// as an opaque given, which flattens whatever the property is propagating: a
// shape bound this way reads as a single surface, a color on it goes unseen.
func (f *frame) evalReturnK(name string, mb multiBinding) (Value, error) {
	return f.returnK(name, mb.value, mb.index)
}

// returnK evaluates one value of a multi-value operand: a call to a multi-return
// function, or a remap of one — remap distributes over the returns, so
// `a, b = remap(f(), q)` binds each of f's returns under the rebound coordinate
// (the aerospace bracket's `clevisShape, horizPlane = remap(arcShape(), rot)`).
func (f *frame) returnK(name string, op *program.Operand, index int) (Value, error) {
	switch t := op.T.(type) {
	case *program.Operand_FnCall:
		fc := t.FnCall
		if fc.Reference == nil {
			return nil, fmt.Errorf("nil reference in destructuring assignment of %q", name)
		}
		if fc.Reference.Prefix != "" {
			// A domain that splices library calls opaquely has no value for one
			// return of one; lookup names it instead at package scope, so this
			// is only reached from inside a function body.
			if isOpaqueDomain(f.eng.dom) {
				return nil, fmt.Errorf("cannot destructure the cross-package call %q.%q inside a function body in property %T",
					fc.Reference.Prefix, fc.Reference.Name, f.eng.dom)
			}
			// Every other domain differentiates through the library the way
			// evalCall does, taking the name's own return of the inlined body.
			args, err := f.evalAll(fc.Arguments)
			if err != nil {
				return nil, err
			}
			ns := f.eng.imports[fc.Reference.Prefix]
			if ns == nil {
				return nil, fmt.Errorf("cannot destructure cross-package call %q.%q: package %q is not resolvable here",
					fc.Reference.Prefix, fc.Reference.Name, fc.Reference.Prefix)
			}
			scope := f.namespaceFrame(ns)
			fn := scope.resolveFunc(fc.Reference.Name, len(args))
			if fn == nil {
				return nil, fmt.Errorf("cannot evaluate call to unknown function %q.%q with %d arguments",
					fc.Reference.Prefix, fc.Reference.Name, len(args))
			}
			return f.inlineReturnIn(scope, fc, fn, args, index)
		}
		args, err := f.evalAll(fc.Arguments)
		if err != nil {
			return nil, err
		}
		fn := f.resolveFunc(fc.Reference.Name, len(args))
		if fn == nil {
			return nil, fmt.Errorf("cannot evaluate call to unknown function %q with %d arguments", fc.Reference.Name, len(args))
		}
		return f.inlineReturn(fc, fn, args, index)

	case *program.Operand_Expression:
		if t.Expression.Operation == program.Operation_REMAP && len(t.Expression.Operands) == 2 {
			child, err := f.remapFrame(t.Expression)
			if err != nil {
				return nil, err
			}
			v, err := child.returnK(name, t.Expression.Operands[0], index)
			if err != nil {
				return nil, err
			}
			return child.remapped(v), nil
		}
	case *program.Operand_Value, *program.Operand_Reference, *program.Operand_PropRef:
		// fall through to error
	default:
		return nil, fmt.Errorf("internal error: unhandled operand type %T in returnK; please add a new case to this switch", t)
	}
	return nil, fmt.Errorf("destructuring assignment of %q must call a multi-return function (or remap one), got %T",
		name, op.T)
}

// isOpaqueDomain reports whether dom splices qualified calls as givens rather
// than differentiating through them.
func isOpaqueDomain(dom Domain) bool {
	_, ok := dom.(OpaqueDomain)
	return ok
}

// isQualifiedCall reports whether op is a call to another package's function,
// or a remap of one — the two forms a destructuring assignment binds.
func isQualifiedCall(op *program.Operand) bool {
	switch t := op.T.(type) {
	case *program.Operand_FnCall:
		return t.FnCall.Reference != nil && t.FnCall.Reference.Prefix != ""
	case *program.Operand_Expression:
		if t.Expression.Operation == program.Operation_REMAP && len(t.Expression.Operands) == 2 {
			return isQualifiedCall(t.Expression.Operands[0])
		}
	}
	return false
}

// resolveFunc resolves a call name, at the given arity, to the function it
// names: one declared in lexical scope, an arity-keyed overload, or — for a
// higher-order combiner such as omega's unionRoundExact, whose sdf1/sdf2
// parameters are SDF thunks that its body calls — the function that parameter
// was bound to at the call site. A thunk can be handed down through several
// combiners, so the binding is followed as far as it goes.
func (f *frame) resolveFunc(name string, arity int) *program.Function {
	scope := f
	seen := map[string]bool{}
	for !seen[name] {
		seen[name] = true
		for s := scope; s != nil; s = s.parent {
			if g, ok := s.funcs[name]; ok {
				for _, candidate := range g {
					if len(candidate.Parameters) == arity {
						return candidate
					}
				}
			}
		}
		// Not a function declared anywhere in scope: the name may be a
		// parameter bound to a function reference. Follow that binding out
		// to the frame that supplied the argument and resolve the referenced
		// name there.
		b, ok := scope.lookupArg(name)
		if !ok {
			return nil
		}
		ref, ok := funcRefName(b.op)
		if !ok {
			return nil
		}
		name, scope = ref, b.frame
	}
	return nil
}

// lookupArg finds the argument binding of a parameter name in this scope chain.
func (f *frame) lookupArg(name string) (argBinding, bool) {
	for s := f; s != nil; s = s.parent {
		if b, ok := s.argOps[name]; ok {
			return b, true
		}
	}
	return argBinding{}, false
}

// funcRefName reports the name a function-reference operand refers to.
func funcRefName(op *program.Operand) (string, bool) {
	if op == nil {
		return "", false
	}
	if t, ok := op.T.(*program.Operand_Reference); ok {
		return t.Reference.Name, true
	}
	return "", false
}

// --- package scope helpers --------------------------------------------------

func collectAssignments(assigns []*program.Assignment) map[string]*program.Operand {
	m := make(map[string]*program.Operand)
	for _, a := range assigns {
		if a == nil || len(a.Variables) != 1 {
			continue
		}
		m[a.Variables[0].Name] = a.Expression
	}
	return m
}

// multiBinding records that a name was bound by a destructuring assignment
// (e.g. `a, b = f()`): value is the shared right-hand side — a call to a
// multi-return function, or a remap of one — and index is the name's position
// in the return list. It mirrors the constant evaluator's binding of the same
// name, which only resolves the bare-call form.
type multiBinding struct {
	value *program.Operand
	index int
}

// collectMultiAssignments maps each name of a destructuring assignment to that
// call and the name's position in its return list; single-name assignments are
// collectAssignments' job. Nil when there are none.
func collectMultiAssignments(assigns []*program.Assignment) map[string]multiBinding {
	var m map[string]multiBinding
	for _, a := range assigns {
		if a == nil || len(a.Variables) < 2 {
			continue
		}
		if m == nil {
			m = make(map[string]multiBinding)
		}
		for i, v := range a.Variables {
			m[v.Name] = multiBinding{value: a.Expression, index: i}
		}
	}
	return m
}

func collectFunctions(fnTypes []*program.FnType) map[string][]*program.Function {
	m := make(map[string][]*program.Function)
	for _, ft := range fnTypes {
		if ft == nil {
			continue
		}
		for _, group := range ft.Groups {
			if group == nil {
				continue
			}
			for _, fn := range group.Functions {
				if fn != nil {
					m[ft.Name] = append(m[ft.Name], fn)
				}
			}
		}
	}
	return m
}
