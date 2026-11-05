package property

import (
	"fmt"
	"math"

	"gradientcontrol.com/omega/core/program"
)

const ln2 = 0.6931471805599453

// dual is the derivative Domain's value: a primal paired with its tangent
// (derivative w.r.t. the active seed direction). Both are GSL operands.
type dual struct {
	primal  *program.Operand
	tangent *program.Operand
}

// paramTangent holds the symbolic tangents of a function parameter, used when
// generating a companion function: dpAxis holds the parameter's derivative
// along each spatial axis and dt its temporal derivative (typically references
// to extra companion parameters). Any entry may be nil.
//
// Three separate per-axis tangents rather than one packed gradient, because
// only a scalar parameter's gradient can be unpacked by component: the
// derivative of a scalar along axis k is the k-th component of its gradient,
// but the derivative of a vector along axis k is itself a vector, and a packed
// gradient of a vector is a Jacobian whose k-th row is not a component. Each
// per-axis tangent instead has the parameter's own shape, whatever that is.
type paramTangent struct {
	dpAxis [3]*program.Operand
	dt     *program.Operand
}

// derivative is the forward-mode automatic-differentiation property. It expands
// the .dtime (temporal) and .dp (spatial gradient) operators, and .d(var), the
// derivative with respect to a top-level variable.
//
// params is non-empty only while generating a companion function body: it maps
// a parameter name to its symbolic tangents so a .dp / .dtime written inside the
// function resolves against the gradient passed in at the call site.
type derivative struct {
	decls      map[string]*program.Operand
	multiDecls map[string]multiBinding // names bound by a destructuring assignment
	funcs      map[string][]*program.Function
	imports    map[string]*namespace // qualified imports, inlined when differentiating
	params     map[string]*paramTangent
	own        map[string]bool // the package's own function names, for box()
	// frames expands a .u or .v met while differentiating (bar.u.dp, the
	// direction a lay is read from). Nil in a companion body, which has no
	// call site to name the shape.
	frames *frames
}

// second is the sibling second-closest property over the same scope, for
// expanding a SECONDCLOSEST met while differentiating (field.second.dp).
func (d *derivative) second() *second {
	return &second{decls: d.decls, multiDecls: d.multiDecls, funcs: d.funcs}
}

// box is the sibling bounding-box property over the same scope, for expanding a
// BOUNDINGBOX met while differentiating. A box does not depend on p, so its
// derivative is zero once expanded.
func (d *derivative) box() *boxProp {
	return &boxProp{decls: d.decls, multiDecls: d.multiDecls, funcs: d.funcs, imports: d.imports, own: d.own}
}

func (d *derivative) engine(seed map[string]*program.Operand) *Engine {
	return d.engineIn(&derivDomain{d: d, seed: seed})
}

// engineIn is engine for a Domain the caller has built, when a seed set is not
// all it needs: a .d(var) also names the variable it holds independent.
func (d *derivative) engineIn(dom *derivDomain) *Engine {
	return &Engine{dom: dom, decls: d.decls, multiDecls: d.multiDecls, funcs: d.funcs, imports: d.imports}
}

// temporal computes d(target)/d(time), preserving target's shape.
func (d *derivative) temporal(target *program.Operand) (*program.Operand, error) {
	v, err := d.engine(d.timeSeed()).Eval(target)
	if err != nil {
		return nil, err
	}
	return v.(dual).tangent, nil
}

// axisTangent computes d(target)/d(x_k), the directional derivative along one
// spatial axis, preserving target's shape.
func (d *derivative) axisTangent(target *program.Operand, k int) (*program.Operand, error) {
	v, err := d.engine(d.axisSeed(k)).Eval(target)
	if err != nil {
		return nil, err
	}
	return v.(dual).tangent, nil
}

// gradient computes the spatial gradient d(target)/d(p) of a scalar field as a
// vec3, by differentiating along each axis and collecting.
func (d *derivative) gradient(target *program.Operand) (*program.Operand, error) {
	axes := make([]*program.Operand, 3)
	for k := 0; k < 3; k++ {
		t, err := d.axisTangent(target, k)
		if err != nil {
			return nil, err
		}
		axes[k] = t
	}
	return nary(program.Operation_COLLECT, axes...), nil
}

// expandDeriv expands a DERIV operand (.dtime / .dp / .d(var)) to a derivative
// operand.
func (d *derivative) expandDeriv(operands []*program.Operand) (*program.Operand, error) {
	if len(operands) != 2 {
		return nil, fmt.Errorf("derivative expects 2 operands, got %d", len(operands))
	}
	wrtRef, ok := operands[1].T.(*program.Operand_Reference)
	if !ok {
		return nil, fmt.Errorf("derivative target variable must be a reference")
	}
	switch wrtRef.Reference.Name {
	case program.TimeVarName:
		return d.temporal(operands[0])
	case program.PVarName:
		return d.gradient(operands[0])
	default:
		return d.wrt(operands[0], wrtRef.Reference.Name)
	}
}

// wrt computes d(target)/d(name) for a top-level variable, preserving target's
// shape. The variable is INDEPENDENT: its own assignment is not differentiated
// (see derivDomain.Seeded), so name' = 1 whatever it was set to, and every
// declaration written in terms of it picks up its tangent by the chain rule —
// with h = budget / r, (pi * r * r * h).d(r) differentiates through h.
//
// Only a top-level variable, and only from a top-level expression: a parameter
// would need the call site's tangent, which is the companion machinery (.dp /
// .dtime inside a function). A scalar variable gives target's shape. A vector
// variable written as a literal [..] goes through wrtVector, one seed per
// component. Types are not known when properties expand, so a vector that is
// not a literal is not recognised here: it is seeded as a scalar and fails
// later, in type analysis, as a shape error.
func (d *derivative) wrt(target *program.Operand, name string) (*program.Operand, error) {
	if d.params != nil {
		return nil, fmt.Errorf("cannot differentiate with respect to %q inside a function body; .d(%s) is only supported in top-level expressions", name, name)
	}
	body, declared := d.decls[name]
	_, destructured := d.multiDecls[name]
	if !declared && !destructured {
		return nil, fmt.Errorf("cannot differentiate with respect to %q: not a top-level variable", name)
	}
	if declared {
		if e, ok := body.T.(*program.Operand_Expression); ok && e.Expression.Operation == program.Operation_COLLECT {
			return d.wrtVector(target, name, e.Expression.Operands)
		}
	}
	return d.wrtSeeded(target, name, num(1))
}

// wrtVector computes d(target)/d(name) for a vector variable of n components,
// seeding each component in turn and collecting, as gradient does for the
// three axes: a scalar target gives a vector of n partials, a vector target a
// list of n vectors — the Jacobian by columns, column k being d(target)/d(name_k).
func (d *derivative) wrtVector(target *program.Operand, name string, components []*program.Operand) (*program.Operand, error) {
	n := len(components)
	for _, c := range components {
		if e, ok := c.T.(*program.Operand_Expression); ok && e.Expression.Operation == program.Operation_COLLECT {
			return nil, fmt.Errorf("cannot differentiate with respect to %q: it is a list of vectors, and .d takes a scalar or vector variable", name)
		}
	}
	cols := make([]*program.Operand, n)
	for k := 0; k < n; k++ {
		e := make([]*program.Operand, n)
		for i := range e {
			e[i] = num(0)
		}
		e[k] = num(1)
		t, err := d.wrtSeeded(target, name, nary(program.Operation_COLLECT, e...))
		if err != nil {
			return nil, err
		}
		cols[k] = t
	}
	return nary(program.Operation_COLLECT, cols...), nil
}

// wrtSeeded is one seed direction of a .d(var): name' = tangent, which has
// name's shape (1 for a scalar, a unit vector for one component of a vector).
func (d *derivative) wrtSeeded(target *program.Operand, name string, tangent *program.Operand) (*program.Operand, error) {
	v, err := d.engineIn(&derivDomain{d: d, seed: d.varSeed(name, tangent), wrt: name}).Eval(target)
	if err != nil {
		return nil, err
	}
	return v.(dual).tangent, nil
}

// varSeed seeds d/d(name) for a top-level variable: name' = tangent, time and
// every spatial coordinate '=0. No parameter tangents: wrt refuses companion
// bodies.
func (d *derivative) varSeed(name string, tangent *program.Operand) map[string]*program.Operand {
	return map[string]*program.Operand{
		program.TimeVarName: num(0),
		program.PVarName:    nary(program.Operation_COLLECT, num(0), num(0), num(0)),
		program.XVarName:    num(0),
		program.YVarName:    num(0),
		program.ZVarName:    num(0),
		name:                tangent,
	}
}

// timeSeed seeds d/d(time): time'=1, every spatial coordinate '=0, plus each
// parameter's temporal tangent (when generating a companion).
func (d *derivative) timeSeed() map[string]*program.Operand {
	s := map[string]*program.Operand{
		program.TimeVarName: num(1),
		program.PVarName:    nary(program.Operation_COLLECT, num(0), num(0), num(0)),
		program.XVarName:    num(0),
		program.YVarName:    num(0),
		program.ZVarName:    num(0),
	}
	for name, pt := range d.params {
		if pt.dt != nil {
			s[name] = pt.dt
		}
	}
	return s
}

// axisSeed seeds d/d(x_k) for the spatial gradient: the k-th coordinate '=1, the
// other coordinates and time '=0, plus each parameter's own derivative along
// axis k (when generating a companion).
func (d *derivative) axisSeed(k int) map[string]*program.Operand {
	e := func(i int) *program.Operand {
		if i == k {
			return num(1)
		}
		return num(0)
	}
	s := map[string]*program.Operand{
		program.TimeVarName: num(0),
		program.PVarName:    nary(program.Operation_COLLECT, e(0), e(1), e(2)),
		program.XVarName:    e(0),
		program.YVarName:    e(1),
		program.ZVarName:    e(2),
	}
	for name, pt := range d.params {
		if pt.dpAxis[k] != nil {
			s[name] = pt.dpAxis[k]
		}
	}
	return s
}

// derivDomain is the derivative Domain for one seed direction.
type derivDomain struct {
	d    *derivative
	seed map[string]*program.Operand // implied-var name -> tangent operand
	// wrt names the top-level variable a .d(var) differentiates with respect
	// to, which Seeded holds independent of its assignment. Empty for .dtime
	// and .dp, whose seeds are implied variables with no assignment to skip.
	wrt string
}

// Seeded stands in for the declaration of the variable a .d(var) is taken with
// respect to: its live reference with the seed's tangent (1, or a unit vector
// for one component of a vector variable), in place of the derivative of
// whatever it was assigned. That is what makes the variable independent.
func (dd *derivDomain) Seeded(name string) (Value, bool) {
	if dd.wrt == "" || name != dd.wrt {
		return nil, false
	}
	return dual{primal: ref(name), tangent: dd.seed[name]}, true
}

func (dd *derivDomain) Literal(v float64) Value { return dual{num(v), num(0)} }

func (dd *derivDomain) ImpliedVar(name string) (Value, error) {
	var primal *program.Operand
	switch name {
	case program.PVarName:
		primal = ref(program.PVarName)
	case program.TimeVarName:
		primal = ref(program.TimeVarName)
	case program.XVarName:
		primal = ref(program.XVarName)
	case program.YVarName:
		primal = ref(program.YVarName)
	case program.ZVarName:
		primal = ref(program.ZVarName)
	default:
		return nil, fmt.Errorf("not an implied variable: %q", name)
	}
	return dual{primal: primal, tangent: dd.seed[name]}, nil
}

// Opaque: a slider input or unresolved global is constant w.r.t. the seed —
// unless it is a function parameter with a seeded tangent (companion bodies).
func (dd *derivDomain) Opaque(name string) Value {
	tangent, ok := dd.seed[name]
	if !ok {
		tangent = num(0)
	}
	return dual{primal: ref(name), tangent: tangent}
}

func (dd *derivDomain) Collect(args []Value) Value {
	ds := duals(args)
	return dual{nary(program.Operation_COLLECT, primalsOf(ds)...), nary(program.Operation_COLLECT, tangentsOf(ds)...)}
}

// Access and Index keep a zero tangent zero rather than selecting into it. A
// constant's tangent is the literal 0 whatever its shape, and a constant list's
// is a collection of zeros, so indexing one would otherwise produce an
// expression whose shape nothing can infer — which is what broke
// differentiating a reduce over a list returned by a function, e.g. the noise
// library's cellular offsets.
func (dd *derivDomain) Access(base Value, selectors []*program.Operand) Value {
	b := base.(dual)
	pArgs := append([]*program.Operand{b.primal}, selectors...)
	primal := nary(program.Operation_ACCESS, pArgs...)
	if isZeroTangent(b.tangent) {
		return dual{primal, num(0)}
	}
	tArgs := append([]*program.Operand{b.tangent}, selectors...)
	return dual{primal, nary(program.Operation_ACCESS, tArgs...)}
}

func (dd *derivDomain) Index(base Value, index *program.Operand) Value {
	b := base.(dual)
	primal := nary(program.Operation_INDEX, b.primal, index)
	if isZeroTangent(b.tangent) {
		return dual{primal, num(0)}
	}
	return dual{primal, nary(program.Operation_INDEX, b.tangent, index)}
}

func (dd *derivDomain) Select(cond Value, branches []Value) Value {
	c := cond.(dual)
	bs := duals(branches)
	pArgs := append([]*program.Operand{c.primal}, primalsOf(bs)...)
	tArgs := append([]*program.Operand{c.primal}, tangentsOf(bs)...)
	return dual{nary(program.Operation_SELECT, pArgs...), nary(program.Operation_SELECT, tArgs...)}
}

// Named keeps the value-part of a declaration as a live reference (so a slider
// or other uniform tracks its current value in the derivative) while preserving
// the computed tangent.
func (dd *derivDomain) Named(name string, value Value) Value {
	return dual{primal: ref(name), tangent: value.(dual).tangent}
}

func (dd *derivDomain) Operator(operation program.Operation, operands []*program.Operand) (*program.Operand, bool, error) {
	switch operation {
	case program.Operation_DERIV:
		op, err := dd.d.expandDeriv(operands)
		return op, true, err
	case program.Operation_SECONDCLOSEST:
		// A .second nested under a derivative (field.second.dp): expand it to
		// its plain min/max/select field with the second property over the same
		// scope, then the Engine evaluates that expansion in this Domain — so
		// the derivative differentiates straight through the order statistic.
		op, err := dd.d.second().expandSecond(operands)
		return op, true, err
	case program.Operation_BOUNDINGBOX:
		op, err := dd.d.box().expandBox(operands)
		return op, true, err
	case program.Operation_SURFACEU, program.Operation_SURFACEV:
		// A frame read nested under a derivative (bar.u.dp): expand it to the
		// frame's own expression of p, which the Engine then differentiates in
		// this Domain, so .u.dp is the gradient of the coordinate itself.
		if dd.d.frames == nil {
			return nil, true, fmt.Errorf("the .u and .v operators are not supported inside a function body; read the frame at the call site and pass it in")
		}
		op, err := dd.d.frames.expand(operation, operands)
		return op, true, err
	}
	return nil, false, nil
}

func (dd *derivDomain) Op(operation program.Operation, args []Value) (Value, error) {
	ds := duals(args)
	switch operation {
	case program.Operation_ADD:
		// Through add/sub rather than nary, so zero tangents drop instead of
		// surviving as (0 + 0) or (0 - 0) — shapes nothing downstream can use.
		return dual{nary(program.Operation_ADD, primalsOf(ds)...), add(tangentsOf(ds)...)}, nil
	case program.Operation_SUB:
		return dual{nary(program.Operation_SUB, primalsOf(ds)...), subAll(tangentsOf(ds))}, nil
	case program.Operation_MUL:
		return foldDual(ds, mulDual), nil
	case program.Operation_DIV:
		return foldDual(ds, divDual), nil
	case program.Operation_NEG:
		return dual{neg(ds[0].primal), neg(ds[0].tangent)}, nil
	case program.Operation_ABS:
		return dual{un(program.Operation_ABS, ds[0].primal), mul(call("sign", ds[0].primal), ds[0].tangent)}, nil
	case program.Operation_SQUARE:
		return dual{un(program.Operation_SQUARE, ds[0].primal), mul(mul(num(2), ds[0].primal), ds[0].tangent)}, nil
	case program.Operation_SQRT:
		// Guarded like the vector norm below, and for the same combiner: the
		// list form of intersectRound spells its fillet term sqrt(sum of
		// squared clamped excesses) where the two- and three-input forms spell
		// it length(...). Where every clamp is zero the term is identically
		// zero, so its derivative is zero -- but unguarded the quotient is
		// 0/0, and one NaN poisons the whole field's .dp.
		p := un(program.Operation_SQRT, ds[0].primal)
		return dual{p, div(ds[0].tangent, mul(num(2), nonZero(p)))}, nil
	case program.Operation_EXP:
		p := un(program.Operation_EXP, ds[0].primal)
		return dual{p, mul(p, ds[0].tangent)}, nil
	case program.Operation_LN:
		return dual{un(program.Operation_LN, ds[0].primal), div(ds[0].tangent, ds[0].primal)}, nil
	case program.Operation_SIN:
		return dual{un(program.Operation_SIN, ds[0].primal), mul(un(program.Operation_COS, ds[0].primal), ds[0].tangent)}, nil
	case program.Operation_COS:
		return dual{un(program.Operation_COS, ds[0].primal), mul(neg(un(program.Operation_SIN, ds[0].primal)), ds[0].tangent)}, nil
	case program.Operation_TAN:
		return dual{un(program.Operation_TAN, ds[0].primal), div(ds[0].tangent, un(program.Operation_SQUARE, un(program.Operation_COS, ds[0].primal)))}, nil
	case program.Operation_ASIN:
		return dual{un(program.Operation_ASIN, ds[0].primal), div(ds[0].tangent, sqrt(sub(num(1), un(program.Operation_SQUARE, ds[0].primal))))}, nil
	case program.Operation_ACOS:
		return dual{un(program.Operation_ACOS, ds[0].primal), neg(div(ds[0].tangent, sqrt(sub(num(1), un(program.Operation_SQUARE, ds[0].primal)))))}, nil
	case program.Operation_ATAN:
		return dual{un(program.Operation_ATAN, ds[0].primal), div(ds[0].tangent, add(num(1), un(program.Operation_SQUARE, ds[0].primal)))}, nil
	case program.Operation_ATAN2:
		// atan2(y, x): d = (x·y' − y·x') / (x² + y²)
		y, x := ds[0], ds[1]
		denom := add(un(program.Operation_SQUARE, x.primal), un(program.Operation_SQUARE, y.primal))
		numer := sub(mul(x.primal, y.tangent), mul(y.primal, x.tangent))
		return dual{nary(program.Operation_ATAN2, y.primal, x.primal), div(numer, denom)}, nil
	case program.Operation_POW:
		return powDual(ds[0], ds[1]), nil
	case program.Operation_MODULO:
		// (a mod b)' = a' − floor(a/b)·b'  (the floor term has zero derivative a.e.)
		a, b := ds[0], ds[1]
		t := sub(a.tangent, mul(un(program.Operation_FLOOR, div(a.primal, b.primal)), b.tangent))
		return dual{nary(program.Operation_MODULO, a.primal, b.primal), t}, nil
	case program.Operation_MIN:
		return foldDualBalanced(ds, minDual), nil
	case program.Operation_MAX:
		return foldDualBalanced(ds, maxDual), nil
	case program.Operation_FLOOR, program.Operation_CEIL, program.Operation_ROUND,
		program.Operation_GT, program.Operation_GE, program.Operation_EQ, program.Operation_NE,
		program.Operation_LE, program.Operation_LT, program.Operation_COMPARE,
		program.Operation_AND, program.Operation_OR, program.Operation_NOT:
		// Piecewise-constant or boolean results: zero derivative a.e.
		return dual{nary(operation, primalsOf(ds)...), num(0)}, nil
	case program.Operation_MAP, program.Operation_REDUCE:
		return nil, fmt.Errorf("cannot differentiate %s", operation.Name())
	}
	return nil, fmt.Errorf("cannot differentiate operation %s", operation.Name())
}

func (dd *derivDomain) Builtin(name string, args []Value) (Value, error) {
	ds := duals(args)
	switch name {
	case "length":
		v := ds[0]
		l := call("length", v.primal)
		if isZeroTangent(v.tangent) {
			return dual{l, num(0)}, nil
		}
		return dual{l, div(call("dot", v.primal, v.tangent), nonZero(l))}, nil
	case program.WorleyF1, program.WorleyF2:
		// F1 is the distance from q to its nearest feature point, so it
		// differentiates exactly as length() does — dot(q - feat, q') / F1 —
		// with the feature point supplied by cellularFeature. That point is
		// locally constant (it only changes when q crosses a Voronoi wall,
		// where F1 is not differentiable anyway), so it contributes no term of
		// its own, the way the lattice hash does not.
		//
		// The result is the TRUE analytic gradient, a unit vector away from the
		// feature point. Differentiating the unrolled 27-way min could only pick
		// a branch and inherit its rounding.
		v := ds[0]
		rest := make([]*program.Operand, 0, len(ds)-1)
		for _, d := range ds[1:] {
			rest = append(rest, d.primal)
		}
		f1 := call(name, append([]*program.Operand{v.primal}, rest...)...)
		if isZeroTangent(v.tangent) {
			return dual{f1, num(0)}, nil
		}
		// Which point the distance measures to. One scan knows both, so each
		// distance's gradient is a direct read: worleyFeature for F1,
		// worleyFarFeature for F2. Recovering the far point by re-hashing its
		// cell would work but would pay a hash the scan already did, and would
		// state "a cell's feature point is cell + hash(cell)" in a second place.
		point := program.WorleyFeature
		if name == program.WorleyF2 {
			point = program.WorleyFarFeature
		}
		feat := call(point, append([]*program.Operand{v.primal}, rest...)...)
		return dual{f1, div(call("dot", sub(v.primal, feat), v.tangent), nonZero(f1))}, nil
	case program.WorleyFeature, program.WorleyFarFeature,
		program.WorleyNearCell, program.WorleyFarCell:
		// The winning cells and their feature points: piecewise constant in q —
		// they change only as q crosses a Voronoi wall, where the distances are
		// not differentiable anyway — so a zero tangent, exactly like
		// noiseHash3's.
		primals := make([]*program.Operand, len(ds))
		for i, d := range ds {
			primals[i] = d.primal
		}
		return dual{call(name, primals...), num(0)}, nil
	case "distance":
		a, b := ds[0], ds[1]
		l := call("distance", a.primal, b.primal)
		if isZeroTangent(a.tangent) && isZeroTangent(b.tangent) {
			return dual{l, num(0)}, nil
		}
		return dual{l, div(call("dot", sub(a.primal, b.primal), sub(a.tangent, b.tangent)), l)}, nil
	case "dot":
		a, b := ds[0], ds[1]
		return dual{call("dot", a.primal, b.primal),
			add(vecTerm("dot", a.primal, b.tangent), vecTerm("dot", a.tangent, b.primal))}, nil
	case "cross":
		a, b := ds[0], ds[1]
		return dual{call("cross", a.primal, b.primal),
			add(vecTerm("cross", a.primal, b.tangent), vecTerm("cross", a.tangent, b.primal))}, nil
	case "normalize":
		v := ds[0]
		l := call("length", v.primal)
		if isZeroTangent(v.tangent) {
			return dual{call("normalize", v.primal), num(0)}, nil
		}
		// v'/|v| − v·(v·v')/|v|³
		t := sub(div(v.tangent, l), div(mul(v.primal, call("dot", v.primal, v.tangent)), mul(l, mul(l, l))))
		return dual{call("normalize", v.primal), t}, nil
	case "clamp":
		x, lo, hi := ds[0], ds[1], ds[2]
		inner := nary(program.Operation_SELECT, bin(program.Operation_GT, x.primal, hi.primal), hi.tangent, x.tangent)
		t := nary(program.Operation_SELECT, bin(program.Operation_LT, x.primal, lo.primal), lo.tangent, inner)
		return dual{call("clamp", x.primal, lo.primal, hi.primal), t}, nil
	case "sign", "trunc":
		return dual{call(name, ds[0].primal), num(0)}, nil
	case "noiseHash3":
		// A nearest-sample lookup into the baked noise tile, keyed by an
		// integer lattice cell: locally constant in p, so the tangent is
		// exactly zero, the way FLOOR's and trunc's are. Scalar zero like
		// every other zero tangent here, even though the primal is a vec3;
		// add and sub drop a literal zero rather than trying to match shapes.
		//
		// The optional octaves argument is carried through untouched: it is a
		// literal the backends read, not a value with a derivative, and
		// dropping it here would silently rebuild the call at two octaves.
		primals := make([]*program.Operand, len(ds))
		for i, d := range ds {
			primals[i] = d.primal
		}
		return dual{call(name, primals...), num(0)}, nil
	case program.GaborSpinodoid, program.FastGaborSpinodoid:
		// The kernel table is baked from the constant arguments and from the
		// drivers sampled at kernel centres, not at q, so q is the only
		// argument the field varies with: d' = dot(grad(q), q'). Each field
		// carries its own gradient builtin.
		gradient := program.GaborSpinodoidGradient
		if name == program.FastGaborSpinodoid {
			gradient = program.FastGaborSpinodoidGradient
		}
		primals := make([]*program.Operand, len(ds))
		for i, d := range ds {
			primals[i] = d.primal
		}
		value := call(name, primals...)
		if isZeroTangent(ds[0].tangent) {
			return dual{value, num(0)}, nil
		}
		return dual{value, call("dot", call(gradient, primals...), ds[0].tangent)}, nil
	case "fract":
		return dual{call("fract", ds[0].primal), ds[0].tangent}, nil
	case program.PixelSize, program.DisplacementDrawn:
		// The pixel footprint at the shading point, or the displacement drawn
		// there: a per-pixel constant, not a function of p, so its tangent is
		// zero.
		return dual{call(name), num(0)}, nil
	case "exp2":
		x := ds[0]
		return dual{call("exp2", x.primal), mul(mul(num(ln2), call("exp2", x.primal)), x.tangent)}, nil
	case "sinh":
		x := ds[0]
		return dual{call("sinh", x.primal), mul(call("cosh", x.primal), x.tangent)}, nil
	case "cosh":
		x := ds[0]
		return dual{call("cosh", x.primal), mul(call("sinh", x.primal), x.tangent)}, nil
	case "tanh":
		x := ds[0]
		return dual{call("tanh", x.primal), mul(sub(num(1), un(program.Operation_SQUARE, call("tanh", x.primal))), x.tangent)}, nil
	case "asinh":
		x := ds[0]
		return dual{call("asinh", x.primal), div(x.tangent, sqrt(add(un(program.Operation_SQUARE, x.primal), num(1))))}, nil
	case "acosh":
		x := ds[0]
		return dual{call("acosh", x.primal), div(x.tangent, sqrt(sub(un(program.Operation_SQUARE, x.primal), num(1))))}, nil
	case "atanh":
		x := ds[0]
		return dual{call("atanh", x.primal), div(x.tangent, sub(num(1), un(program.Operation_SQUARE, x.primal)))}, nil
	case "radians":
		return dual{call("radians", ds[0].primal), mul(num(math.Pi/180), ds[0].tangent)}, nil
	case "degrees":
		return dual{call("degrees", ds[0].primal), mul(num(180/math.Pi), ds[0].tangent)}, nil
	}
	return nil, fmt.Errorf("cannot differentiate builtin %q", name)
}

// --- dual folds & helpers ---------------------------------------------------

// vecTerm builds one term of a vector builtin's product rule, or the literal
// zero when the tangent factor is zero. A zero tangent is a scalar whatever the
// primal's shape, so emitting dot(vec3, 0) would be a shape error; and add
// drops the zero, so the term disappears rather than costing anything.
func vecTerm(name string, a, b *program.Operand) *program.Operand {
	if isZeroTangent(a) || isZeroTangent(b) {
		return num(0)
	}
	return call(name, a, b)
}

func duals(args []Value) []dual {
	out := make([]dual, len(args))
	for i, a := range args {
		out[i] = a.(dual)
	}
	return out
}

func primalsOf(ds []dual) []*program.Operand {
	out := make([]*program.Operand, len(ds))
	for i, d := range ds {
		out[i] = d.primal
	}
	return out
}

func tangentsOf(ds []dual) []*program.Operand {
	out := make([]*program.Operand, len(ds))
	for i, d := range ds {
		out[i] = d.tangent
	}
	return out
}

func foldDual(ds []dual, combine func(a, b dual) dual) Value {
	acc := ds[0]
	for i := 1; i < len(ds); i++ {
		acc = combine(acc, ds[i])
	}
	return acc
}

// scaleTerm builds one term of a product rule, or the literal zero when either
// factor is zero. Like vecTerm, this is load-bearing: a zero tangent is a
// scalar whatever the primal's shape, so `vec3 * 0` would survive as a
// vec3-shaped expression that is not recognisably zero, and adding it to a
// scalar `0 * 2` from the other term is a shape error. Perlin noise hits this
// on `cellHash3(cell) * 2 - 1`.
func scaleTerm(x, y *program.Operand) *program.Operand {
	if isZeroTangent(x) || isZeroTangent(y) {
		return num(0)
	}
	return mul(x, y)
}

// foldDualBalanced is foldDual over a balanced tree instead of a left-leaning
// chain. ONLY FOR AN EXACTLY ASSOCIATIVE COMBINER — min and max, and nothing
// else. Division is not associative at all (a/b/c/d is a/(b·c·d) left-folded
// and a·d/(b·c) balanced), and multiplication is associative in the reals but
// not in f32, so both keep the chain.
//
// What this buys is not fewer operations — it is fewer OCCURRENCES. Nothing
// downstream shares a subexpression, so a value is re-emitted once per textual
// mention, and minDual mentions both of its operands' primals: once in the
// comparison and once in the min. Chained, the accumulator at step k is a tree
// of size k that gets mentioned again at step k+1, so one seed's tangent is
// Θ(n²) and the color field, which takes one seed per colored unit, is Θ(n³).
// Balanced, the same tangent is Θ(n log n).
//
// Measured on the examples, with the distance field byte-identical in both:
// Drill.wgsl 106212 → 76659 bytes, Fasteners.wgsl 46118 → 36589, and the
// synthesized remaps 297 → 209 and 246 → 190 with no distinct body lost —
// every one that went was a duplicate of one that stayed.
//
// TIES BREAK DIFFERENTLY, which is the whole of the behavior change. min's
// VALUE is unaffected: IEEE754 min is exactly associative, so the primal is
// identical. But minDual's condition is a strict <, so where several shapes are
// at exactly equal distance, which one contributes its tangent depends on the
// association. That set has measure zero and the previous order was itself
// arbitrary, but it is a change, not a pure rewrite.
func foldDualBalanced(ds []dual, combine func(a, b dual) dual) Value {
	cur := ds
	for len(cur) > 1 {
		next := make([]dual, 0, (len(cur)+1)/2)
		for i := 0; i+1 < len(cur); i += 2 {
			next = append(next, combine(cur[i], cur[i+1]))
		}
		if len(cur)%2 == 1 {
			next = append(next, cur[len(cur)-1]) // odd one out rides to the next round
		}
		cur = next
	}
	return cur[0]
}

func mulDual(a, b dual) dual {
	return dual{mul(a.primal, b.primal), add(scaleTerm(a.tangent, b.primal), scaleTerm(a.primal, b.tangent))}
}

func divDual(a, b dual) dual {
	num1 := sub(scaleTerm(a.tangent, b.primal), scaleTerm(a.primal, b.tangent))
	if isZeroTangent(num1) {
		return dual{div(a.primal, b.primal), num(0)}
	}
	return dual{div(a.primal, b.primal), div(num1, mul(b.primal, b.primal))}
}

func minDual(a, b dual) dual {
	cond := bin(program.Operation_LT, a.primal, b.primal)
	return dual{bin(program.Operation_MIN, a.primal, b.primal), nary(program.Operation_SELECT, cond, a.tangent, b.tangent)}
}

func maxDual(a, b dual) dual {
	cond := bin(program.Operation_GT, a.primal, b.primal)
	return dual{bin(program.Operation_MAX, a.primal, b.primal), nary(program.Operation_SELECT, cond, a.tangent, b.tangent)}
}

func powDual(b, ex dual) dual {
	p := bin(program.Operation_POW, b.primal, ex.primal)
	if lit, ok := ex.primal.T.(*program.Operand_Value); ok {
		// Constant exponent k: d(bᵏ) = k·bᵏ⁻¹·b'
		k := lit.Value
		return dual{p, mul(mul(num(k), bin(program.Operation_POW, b.primal, num(k-1))), b.tangent)}
	}
	// General: d(bᵉ) = bᵉ·(e'·ln(b) + e·b'/b)
	t := mul(p, add(mul(ex.tangent, un(program.Operation_LN, b.primal)), div(mul(ex.primal, b.tangent), b.primal)))
	return dual{p, t}
}

// --- operand construction helpers -------------------------------------------

func num(v float64) *program.Operand {
	return &program.Operand{
		T: &program.Operand_Value{Value: v},
	}
}

func call(name string, xs ...*program.Operand) *program.Operand {
	return &program.Operand{
		T: &program.Operand_FnCall{
			FnCall: &program.FnCall{
				Reference: &program.Reference{Name: name},
				Arguments: xs,
			},
		},
	}
}

func nary(o program.Operation, xs ...*program.Operand) *program.Operand {
	return &program.Operand{
		T: &program.Operand_Expression{
			Expression: &program.Expression{
				Operation: o,
				Operands:  xs,
			},
		},
	}
}

func un(o program.Operation, a *program.Operand) *program.Operand {
	return nary(o, a)
}

func bin(o program.Operation, a, b *program.Operand) *program.Operand {
	return nary(o, a, b)
}

// isZeroLit reports whether an operand is the literal 0, which is how every
// zero tangent in this file is spelled.
func isZeroLit(a *program.Operand) bool {
	v, ok := a.T.(*program.Operand_Value)
	return ok && v.Value == 0
}

// isZeroTangent reports whether an operand is structurally zero: the literal 0,
// or a collection of zeros, which is what a constant vector or list of vectors
// differentiates to.
func isZeroTangent(a *program.Operand) bool {
	if isZeroLit(a) {
		return true
	}
	e, ok := a.T.(*program.Operand_Expression)
	if !ok || e.Expression.Operation != program.Operation_COLLECT || len(e.Expression.Operands) == 0 {
		return false
	}
	for _, x := range e.Expression.Operands {
		if !isZeroTangent(x) {
			return false
		}
	}
	return true
}

// add and sub drop a literal zero rather than emitting it. That is the usual
// sparse-tangent simplification, and here it is load-bearing rather than
// cosmetic: a zero tangent is a scalar whatever the primal's shape, so
// subtracting one from a vector tangent would otherwise be a shape error. It
// is why differentiating anything built on floor or the noise hash works at
// all. It also shrinks every emitted derivative.
func add(xs ...*program.Operand) *program.Operand {
	kept := make([]*program.Operand, 0, len(xs))
	for _, x := range xs {
		if !isZeroLit(x) {
			kept = append(kept, x)
		}
	}
	if len(kept) == 0 {
		return num(0)
	}
	if len(kept) == 1 {
		return kept[0]
	}
	return nary(program.Operation_ADD, kept...)
}

// subAll folds a left-associated chain of subtractions through sub, so a zero
// term drops at each step.
func subAll(xs []*program.Operand) *program.Operand {
	if len(xs) == 0 {
		return num(0)
	}
	acc := xs[0]
	for _, x := range xs[1:] {
		acc = sub(acc, x)
	}
	return acc
}

func sub(a, b *program.Operand) *program.Operand {
	if isZeroLit(b) {
		return a
	}
	if isZeroLit(a) {
		return neg(b)
	}
	return bin(program.Operation_SUB, a, b)
}
func mul(a, b *program.Operand) *program.Operand { return bin(program.Operation_MUL, a, b) }
func div(a, b *program.Operand) *program.Operand { return bin(program.Operation_DIV, a, b) }
func neg(a *program.Operand) *program.Operand    { return un(program.Operation_NEG, a) }
func sqrt(a *program.Operand) *program.Operand   { return un(program.Operation_SQRT, a) }

// nonZero returns a norm unchanged, or 1 where it is zero, so a derivative that
// divides by it stays finite at the origin — the apex of the cone, where the
// norm is not differentiable.
//
// The numerator vanishes there too, so the quotient reads 0 rather than 0/0,
// which is the right answer whenever the norm is zero across a region and its
// derivative is genuinely zero. omega's round combiners do exactly that: their
// fillet term is the length of the clamped excesses, and beyond k from every
// input both clamps are identically zero, so the term contributes nothing and
// its derivative should be nothing. Unguarded it was NaN, and that poisoned the
// whole field's .dp even though the field's own gradient there is a clean 1.
//
// A comparison rather than an epsilon. Comparisons evaluate to exactly 0 or 1,
// so a nonzero norm passes through bit-for-bit while a zero one becomes 1 — an
// epsilon would instead perturb every small norm, and being a constant it would
// also have to clear the WGSL backend's printable floor (it emits with %f, so
// anything below about 5e-7 becomes 0.0 in the shader and the guard silently
// vanishes there while still holding in the mesher).
//
// Written as arithmetic, not a select, because backends evaluate both arms of a
// select — the mesher does so on intervals — and the unguarded denominator in
// one arm would reach the result whichever way the branch fell.
func nonZero(l *program.Operand) *program.Operand {
	return add(l, sub(num(1), bin(program.Operation_GT, l, num(0))))
}
