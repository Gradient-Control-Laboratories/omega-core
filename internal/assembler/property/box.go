package property

import (
	"math"
	"strings"

	"gradientcontrol.com/omega/core/program"
)

// The bounding-box property (.box): an axis-aligned box containing a field's
// surface, the points where it is 0 or steps across 0. Every value carries a
// second box, for its solid, the region where the field is <= 0, because an
// intersection needs it: the surface of max(a, b) is where one side is on its
// surface and the other is <= 0, so it lies in a's surface inside b's solid, or
// b's surface inside a's solid, and b's surface box alone cannot say where a's
// solid ends. The surface is the solid's edge, so the solid's box bounds it
// too; a value keeps a surface box of its own only when it knows a tighter one:
//
//   - a complement, -d, has d's surface and a solid everywhere else;
//   - a declared box is the part's box, taken on trust: the renderer cuts the
//     surface to it (the mesher caps it instead), and nothing inside it is
//     evaluated, so the surface is inside the box while the solid, which the
//     cut leaves whole, is unbounded.
//
// The box is symbolic, like .second: its corners are GSL expressions of the
// same sliders, times and parameters the field uses, so a box whose shape
// depends on a slider stays right as the slider moves, and a constant box folds
// to numbers. Reading x.box gives [lo, hi], two vec3. An axis the analysis
// cannot bound reads as -nothing .. nothing; an empty box reads with lo > hi.
//
// Boxes come from named builtin primitives (sphere, box, torus, the capped
// cylinders and cones, …) and flow through the operations whose effect on the
// solid is known:
//
//	min, and the union combiners    the union of the boxes
//	max, sharp intersections        the intersection of the solids, and each
//	                                surface within the other sides' solids
//	a round intersection            the intersection of the solids (its fillet
//	                                is no side's surface)
//	a round difference              the target's solid
//	d - r, r not depending on p     grown by r (over the growth factor, below)
//	k * d, d / k, k literal         the same surface, and the same solid if k > 0
//	-d                              the same surface
//	abs(d)                          the same box (the shell {|d| <= 0} is d's surface)
//	a ternary                       the union of its branches' solids (it can
//	                                step between them anywhere)
//	a positive constant             empty (nothing is solid)
//
// A field written out in p's coordinates bounds itself the same way, which is
// how library parts are built (a head is max of a cylinder and two planes):
//
//	s * p.x + c, s a literal        a half-space, its surface the plane
//	abs(s * p.x + c)                the plane, as a field of distance to it
//	length(p - c), length(p.xy - c) the point, or the line along the axis left
//	                                out, as a field of distance to it
//
// so length(p.xy) - r is a cylinder along z, abs(p.z) - h a slab, and max of
// them the capped cylinder they cut.
//
// Everything else is unbounded, including every remap: a remapped field's box
// is cleared, since an arbitrary coordinate says nothing about where the shape
// went. The transform library says exactly where (transform.translate,
// rotateX/Y/Z, rotate, scale, affine), and its calls are read before the remap
// inside them, so a placed part keeps its box. Unbounded is always a sound
// answer, so an expression the analysis cannot follow costs precision, never
// correctness.
//
// The growth factor makes offsets exact. For every r >= 0, the solid of d - r
// lies inside d's box grown by r / growth. An exact distance field has growth
// 1; 0.5 * d has growth 0.5, because its r-offset is d's 2r-offset. It is also
// why the intersection rule survives offsets: {max(a, b) <= r} is
// {a <= r} and {b <= r}, which lies in both grown boxes, which for axis-aligned
// boxes is their intersection grown.

// bbox is a symbolic axis-aligned box. A nil lo[i] or hi[i] is unbounded on that
// side; empty marks a box known to hold nothing.
type bbox struct {
	lo, hi [3]*program.Operand
	empty  bool
}

// boxVal is the box Domain's value.
//
//	op      the value as a GSL operand when it does not depend on p (a literal,
//	        a slider, a radius); nil when it does. Primitive rules read their
//	        arguments from it.
//	box     the box of the value's solid when it is used as a field; nil is
//	        unbounded.
//	surf    the box of its surface when that is tighter than box; nil when box
//	        is the best bound on it.
//	growth  the growth factor (nil is 1), meaningful only with a box.
//	limit   the offset below which the box holds, or 0 for any offset. A
//	        positive constant has no solid only until an offset reaches it:
//	        min(d, 5) - 10 is solid everywhere. So a box built with one holds
//	        for offsets under the constant, which for nothing is every offset
//	        a part will see.
//	elems   the element values of a list, for the list forms of the combiners.
//	coord   the value as an affine function of p's coordinates, when it is one
//	        the raw forms above can read.
type boxVal struct {
	op     *program.Operand
	box    *bbox
	surf   *bbox
	growth *program.Operand
	limit  float64
	elems  []boxVal
	coord  *coordForm
}

// coordForm is a value affine in p's coordinates: the scalar s * p[axis] + off,
// or, when vec is set, the vector whose components are vec's.
type coordForm struct {
	axis  int
	scale float64
	off   *program.Operand
	vec   []coordComp
}

// coordComp is one component of a coordinate vector, sign * (p[axis] - center),
// with sign 1 or -1.
type coordComp struct {
	axis   int
	sign   float64
	center *program.Operand
}

// surface is the box of v's surface: its own when it keeps one, else its
// solid's.
func (v boxVal) surface() *bbox {
	if v.surf != nil {
		return v.surf
	}
	return v.box
}

// boxProp is the .box property over one package scope.
type boxProp struct {
	decls      map[string]*program.Operand
	multiDecls map[string]multiBinding
	funcs      map[string][]*program.Function
	imports    map[string]*namespace // qualified imports, inlined so their primitives' boxes flow through
	// own names the functions the package itself declares. A call is read as
	// a builtin primitive only when the package has not declared its own
	// function of that name, since a user's box() owes nothing to builtin's.
	own map[string]bool
	// sliders names the declarations the viewer can move (those with .min,
	// .max or .period). Their declared value is only where they start, so one
	// used as a field is not the positive constant it reads as here.
	sliders map[string]bool
}

func (b *boxProp) engine() *Engine {
	return &Engine{dom: &boxDomain{b: b}, decls: b.decls, multiDecls: b.multiDecls, funcs: b.funcs, imports: b.imports}
}

// expandBox expands a BOUNDINGBOX operand (.box) to [lo, hi]. A field the engine
// cannot evaluate reads as unbounded, which is always true of it.
func (b *boxProp) expandBox(operands []*program.Operand) (*program.Operand, error) {
	if len(operands) != 1 {
		return boxRead(nil), nil
	}
	v, err := b.engine().Eval(operands[0])
	if err != nil {
		return boxRead(nil), nil
	}
	return boxRead(v.(boxVal).surface()), nil
}

// boxRead writes a box as the GSL value [lo, hi].
func boxRead(bx *bbox) *program.Operand {
	far := ref(nothingName)
	var lo, hi [3]*program.Operand
	for i := 0; i < 3; i++ {
		switch {
		case bx != nil && bx.empty:
			lo[i], hi[i] = far, neg(far)
		default:
			lo[i], hi[i] = neg(far), far
			if bx != nil && bx.lo[i] != nil {
				lo[i] = bx.lo[i]
			}
			if bx != nil && bx.hi[i] != nil {
				hi[i] = bx.hi[i]
			}
		}
	}
	return nary(program.Operation_COLLECT,
		nary(program.Operation_COLLECT, lo[:]...),
		nary(program.Operation_COLLECT, hi[:]...))
}

// nothingName is builtin's empty solid, far from everything; an unbounded axis
// reads as plus or minus it.
const nothingName = "nothing"

type boxDomain struct {
	b *boxProp
}

// unbounded is a value with no box, carrying op when it does not depend on p.
func unbounded(op *program.Operand) boxVal { return boxVal{op: op} }

// constant is a p-independent value. A positive literal's solid is empty, so
// min(d, nothing) keeps d's box; any other constant might be solid everywhere.
func constant(op *program.Operand) boxVal {
	if v, ok := literal(op); ok && v > 0 {
		return boxVal{op: op, box: &bbox{empty: true}, limit: v}
	}
	return boxVal{op: op}
}

func (dd *boxDomain) Literal(v float64) Value { return constant(num(v)) }

func (dd *boxDomain) ImpliedVar(name string) (Value, error) {
	switch name {
	case program.TimeVarName:
		return unbounded(ref(name)), nil
	case program.PVarName:
		comps := make([]coordComp, 3)
		for i := range comps {
			comps[i] = coordComp{axis: i, sign: 1, center: num(0)}
		}
		return coordVector(comps), nil
	case program.XVarName:
		return linear(0, 1, num(0)), nil
	case program.YVarName:
		return linear(1, 1, num(0)), nil
	case program.ZVarName:
		return linear(2, 1, num(0)), nil
	}
	return boxVal{}, nil
}

func (dd *boxDomain) Opaque(name string) Value { return unbounded(ref(name)) }

// Collect gathers a list. Components of p, each moved and turned only by a
// sign, gather into a coordinate vector, as vec2(p.x, p.y) is p.xy.
func (dd *boxDomain) Collect(args []Value) Value {
	vs := boxVals(args)
	out := boxVal{op: opsOrNil(program.Operation_COLLECT, vs), elems: vs}
	if out.op == nil && len(vs) >= 2 && len(vs) <= 3 {
		comps := make([]coordComp, len(vs))
		for i, v := range vs {
			if v.coord == nil || v.coord.vec != nil || math.Abs(v.coord.scale) != 1 {
				return out
			}
			s := v.coord.scale
			comps[i] = coordComp{axis: v.coord.axis, sign: s, center: divOp(negOp(v.coord.off), num(s))}
		}
		out.coord = &coordForm{vec: comps}
	}
	return out
}

func (dd *boxDomain) Access(base Value, selectors []*program.Operand) Value {
	b := base.(boxVal)
	if len(b.elems) > 0 && len(selectors) == 1 {
		if r, ok := selectors[0].T.(*program.Operand_Reference); ok && r.Reference != nil && len(r.Reference.Name) == 1 {
			if k := strings.Index("xyzw", r.Reference.Name); k >= 0 && k < len(b.elems) && b.elems[k].op != nil {
				return b.elems[k]
			}
		}
	}
	if b.op == nil {
		if v, ok := swizzled(b, selectors); ok {
			return v
		}
		return boxVal{}
	}
	return unbounded(nary(program.Operation_ACCESS, append([]*program.Operand{b.op}, selectors...)...))
}

// swizzled reads components of a coordinate vector: p.z is the scalar p.z, and
// (p - c).xy the vector of its first two components.
func swizzled(b boxVal, selectors []*program.Operand) (boxVal, bool) {
	if b.coord == nil || b.coord.vec == nil || len(selectors) != 1 {
		return boxVal{}, false
	}
	r, ok := selectors[0].T.(*program.Operand_Reference)
	if !ok || r.Reference == nil || len(r.Reference.Name) == 0 || len(r.Reference.Name) > 3 {
		return boxVal{}, false
	}
	name := r.Reference.Name
	comps := make([]coordComp, len(name))
	for i, ch := range name {
		k := strings.IndexRune("xyz", ch)
		if k < 0 || k >= len(b.coord.vec) {
			return boxVal{}, false
		}
		comps[i] = b.coord.vec[k]
	}
	if len(comps) == 1 {
		c := comps[0]
		return linear(c.axis, c.sign, negOp(mulOp(num(c.sign), c.center))), true
	}
	return coordVector(comps), true
}

// linear is the field s * p[axis] + off: solid on one side of the plane where
// it is 0, and s times the distance to it, so its growth is |s|.
func linear(axis int, s float64, off *program.Operand) boxVal {
	root := divOp(negOp(off), num(s))
	solid, surf := &bbox{}, &bbox{}
	if s > 0 {
		solid.hi[axis] = root
	} else {
		solid.lo[axis] = root
	}
	surf.lo[axis], surf.hi[axis] = root, root
	return boxVal{box: solid, surf: surf, growth: growthOf(s), coord: &coordForm{axis: axis, scale: s, off: off}}
}

// coordVector is a vector of p's components, with no box of its own: it is not
// a field until length reads it.
func coordVector(comps []coordComp) boxVal {
	return boxVal{coord: &coordForm{vec: comps}}
}

// growthOf is the growth of a field that is s times a distance.
func growthOf(s float64) *program.Operand {
	if a := math.Abs(s); a != 1 {
		return num(a)
	}
	return nil
}

// coordOp applies an operation to coordinate forms: moving, turning over and
// scaling a scalar or vector keeps it one, and abs of a scalar is the distance
// to its plane.
func coordOp(operation program.Operation, vs []boxVal) (boxVal, bool) {
	scalar := func(v boxVal) bool { return v.coord != nil && v.coord.vec == nil }
	vector := func(v boxVal) bool { return v.coord != nil && v.coord.vec != nil }
	constOp := func(v boxVal) bool { return v.op != nil }
	switch {
	case len(vs) == 1 && scalar(vs[0]):
		c := vs[0].coord
		switch operation {
		case program.Operation_NEG:
			return linear(c.axis, -c.scale, negOp(c.off)), true
		case program.Operation_ABS:
			root := divOp(negOp(c.off), num(c.scale))
			bx := &bbox{}
			bx.lo[c.axis], bx.hi[c.axis] = root, root
			return fieldBox(bx, growthOf(c.scale)), true
		}
	case len(vs) == 1 && vector(vs[0]) && operation == program.Operation_NEG:
		return coordVector(movedComps(vs[0].coord.vec, nil, -1)), true
	case len(vs) == 2 && scalar(vs[0]) && constOp(vs[1]):
		c, k := vs[0].coord, vs[1].op
		switch operation {
		case program.Operation_ADD:
			return linear(c.axis, c.scale, addOp(c.off, k)), true
		case program.Operation_SUB:
			return linear(c.axis, c.scale, subOp(c.off, k)), true
		case program.Operation_MUL, program.Operation_DIV:
			v, ok := literal(k)
			if !ok || v == 0 {
				return boxVal{}, false
			}
			if operation == program.Operation_DIV {
				v = 1 / v
			}
			return linear(c.axis, c.scale*v, mulOp(c.off, num(v))), true
		}
	case len(vs) == 2 && constOp(vs[0]) && scalar(vs[1]):
		k, c := vs[0].op, vs[1].coord
		switch operation {
		case program.Operation_ADD:
			return linear(c.axis, c.scale, addOp(k, c.off)), true
		case program.Operation_SUB:
			return linear(c.axis, -c.scale, subOp(k, c.off)), true
		case program.Operation_MUL:
			v, ok := literal(k)
			if !ok || v == 0 {
				return boxVal{}, false
			}
			return linear(c.axis, c.scale*v, mulOp(c.off, num(v))), true
		}
	case len(vs) == 2 && vector(vs[0]) && constOp(vs[1]):
		switch operation {
		case program.Operation_ADD:
			return coordVector(movedComps(vs[0].coord.vec, negOp(vs[1].op), 1)), true
		case program.Operation_SUB:
			return coordVector(movedComps(vs[0].coord.vec, vs[1].op, 1)), true
		}
	case len(vs) == 2 && constOp(vs[0]) && vector(vs[1]):
		switch operation {
		case program.Operation_ADD:
			return coordVector(movedComps(vs[1].coord.vec, negOp(vs[0].op), 1)), true
		case program.Operation_SUB:
			// k - v is -(v - k).
			return coordVector(movedComps(vs[1].coord.vec, vs[0].op, -1)), true
		}
	}
	return boxVal{}, false
}

// movedComps is the vector v - by (by nil for none), times flip. A component
// sign * (p[a] - center) less k is sign * (p[a] - (center + sign * k)).
func movedComps(vec []coordComp, by *program.Operand, flip float64) []coordComp {
	out := make([]coordComp, len(vec))
	for i, c := range vec {
		center := c.center
		if by != nil {
			k := component(by, i)
			if c.sign < 0 {
				k = negOp(k)
			}
			if v, ok := literal(center); ok && v == 0 {
				center = k
			} else {
				center = addOp(center, k)
			}
		}
		out[i] = coordComp{axis: c.axis, sign: c.sign * flip, center: center}
	}
	return out
}

// distanceTo is length of a coordinate vector: the distance to the point, or
// the line along the axis the vector leaves out, where each component is 0.
// A vector that repeats an axis is not a distance, and bounds nothing.
func distanceTo(vec []coordComp) (boxVal, bool) {
	bx := &bbox{}
	for _, c := range vec {
		if bx.lo[c.axis] != nil {
			return boxVal{}, false
		}
		bx.lo[c.axis], bx.hi[c.axis] = c.center, c.center
	}
	return fieldBox(bx, nil), true
}

func (dd *boxDomain) Index(base Value, index *program.Operand) Value {
	b := base.(boxVal)
	if i, ok := literal(index); ok && i >= 0 && int(i) < len(b.elems) && float64(int(i)) == i {
		return b.elems[int(i)]
	}
	if b.op == nil {
		return boxVal{}
	}
	return unbounded(nary(program.Operation_INDEX, b.op, index))
}

// Select could take either branch, so its solid is inside the union. It can
// also step from one branch to the other wherever its condition flips, which
// is a surface for the mesher even where neither branch is 0, so it keeps no
// surface box tighter than that union of solids.
func (dd *boxDomain) Select(cond Value, branches []Value) Value {
	c := cond.(boxVal)
	bs := boxVals(branches)
	// A literal condition takes one branch, as c ? a : b does for c > 0 and
	// v ? lt : eq : gt does by v's sign; a table lookup is made of these.
	if v, ok := literal(c.op); ok {
		switch len(bs) {
		case 2:
			if v > 0 {
				return bs[0]
			}
			return bs[1]
		case 3:
			switch {
			case v < 0:
				return bs[0]
			case v == 0:
				return bs[1]
			}
			return bs[2]
		}
	}
	out := solidOnly(unionOf(bs))
	if c.op != nil {
		if op := opsOrNil(program.Operation_SELECT, append([]boxVal{c}, bs...)); op != nil {
			out.op = op
		}
	}
	return out
}

// Named keeps a p-independent declaration live by name, so a slider tracks its
// value in the box expressions. A slider keeps nothing else: whatever its
// starting value says about its solid need not hold once it moves.
func (dd *boxDomain) Named(name string, value Value) Value {
	v := value.(boxVal)
	if v.op == nil {
		return v
	}
	if dd.b.sliders[name] {
		return unbounded(ref(name))
	}
	v.op = ref(name)
	return v
}

// Operator expands a .box nested inside the field being bounded; the Engine
// evaluates the [lo, hi] it expands to, which does not depend on p.
func (dd *boxDomain) Operator(operation program.Operation, operands []*program.Operand) (*program.Operand, bool, error) {
	if operation != program.Operation_BOUNDINGBOX {
		return nil, false, nil
	}
	op, err := dd.b.expandBox(operands)
	return op, true, err
}

// Remap clears the box: the field is evaluated in another coordinate, and only
// an explicit transform says how that coordinate relates to p. A value that does
// not depend on p is the same under any remap.
func (dd *boxDomain) Remap(target, coord Value) Value {
	t := target.(boxVal)
	if t.op != nil {
		return t
	}
	return boxVal{}
}

func (dd *boxDomain) Op(operation program.Operation, args []Value) (Value, error) {
	vs := boxVals(args)
	if op := opsOrNil(operation, vs); op != nil {
		if v, ok := foldOperation(operation, vs); ok {
			return constant(num(v)), nil
		}
		return constant(op), nil
	}
	if v, ok := coordOp(operation, vs); ok {
		return v, nil
	}
	switch operation {
	case program.Operation_MIN:
		return unionOf(vs), nil
	case program.Operation_MAX:
		return intersectionOf(vs), nil
	case program.Operation_SUB:
		// d - c grows d's solid by c; c - d is a complement.
		if len(vs) == 2 && vs[0].op == nil && vs[1].op != nil {
			return offset(vs[0], vs[1].op), nil
		}
	case program.Operation_ADD:
		// d + c is d - (-c).
		if len(vs) == 2 {
			if vs[0].op == nil && vs[1].op != nil {
				return offset(vs[0], negOp(vs[1].op)), nil
			}
			if vs[1].op == nil && vs[0].op != nil {
				return offset(vs[1], negOp(vs[0].op)), nil
			}
		}
	case program.Operation_MUL:
		// k * d with k a positive literal: the same solid, its offsets scaled.
		if len(vs) == 2 {
			if vs[0].op == nil {
				return scaled(vs[0], vs[1].op, false), nil
			}
			return scaled(vs[1], vs[0].op, false), nil
		}
	case program.Operation_DIV:
		if len(vs) == 2 && vs[0].op == nil {
			return scaled(vs[0], vs[1].op, true), nil
		}
	case program.Operation_ABS:
		// {|d| <= r} is {-r <= d <= r}, inside {d <= r}; |d| is 0 where d is.
		if len(vs) == 1 && vs[0].box != nil && !vs[0].box.empty {
			return withoutOp(vs[0]), nil
		}
		if len(vs) == 1 && vs[0].box == nil {
			return boxVal{surf: vs[0].surf}, nil
		}
	case program.Operation_NEG:
		// -d is 0 where d is, and solid wherever d is not.
		if len(vs) == 1 {
			return boxVal{surf: vs[0].surface()}, nil
		}
	}
	return boxVal{}, nil
}

func (dd *boxDomain) Builtin(name string, args []Value) (Value, error) {
	vs := boxVals(args)
	if ops, ok := opsOf(vs); ok {
		if v, ok := foldBuiltin(name, vs); ok {
			return constant(num(v)), nil
		}
		return constant(call(name, ops...)), nil
	}
	if name == "length" && len(vs) == 1 && vs[0].coord != nil && vs[0].coord.vec != nil {
		if v, ok := distanceTo(vs[0].coord.vec); ok {
			return v, nil
		}
	}
	return boxVal{}, nil
}

// Call reads builtin's primitives and smooth combiners by name, before the
// Engine inlines them: inlined, a sphere is length(p) - r, which says nothing
// about where it is. A primitive whose dimensions depend on p is unbounded.
func (dd *boxDomain) Call(name string, args []Value) (Value, bool, error) {
	if dd.b.own[name] {
		return nil, false, nil
	}
	vs := boxVals(args)
	if name == "revolve" && len(vs) == 3 {
		return revolved(vs[0], vs[1], vs[2]), true, nil
	}
	if v, ok := primitiveBox(name, vs); ok {
		return v, true, nil
	}
	if v, ok := combinerBox(name, vs); ok {
		return v, true, nil
	}
	return nil, false, nil
}

// revolved is the box of revolve(profile, axisPoint, axisDirection)
// (sweep.gsl): the profile's box in x and y, turned about the line through
// axisPoint along axisDirection in the sketch plane. revolve reads the profile
// at (along, out), the distance along the line and the distance out from it
// taken on the right of its direction, so the solid lies inside the cylinder
// about the line that spans the box's reach along it and reaches as far out as
// the box does on that side; the result is that cylinder's box. A point's
// distance to the cylinder is no more than its (along, out) is from the
// profile's box, so the profile's growth holds. A profile unbounded in x or y,
// or an axis that depends on p, leaves the result unbounded.
func revolved(profile, axisPoint, axisDirection boxVal) boxVal {
	b := profile.box
	if profile.op != nil || axisPoint.op == nil || axisDirection.op == nil || b == nil {
		return boxVal{}
	}
	if b.empty {
		return boxVal{box: b, growth: profile.growth, limit: profile.limit}
	}
	if b.lo[0] == nil || b.hi[0] == nil || b.lo[1] == nil || b.hi[1] == nil {
		return boxVal{}
	}
	ax, ay := planar(axisPoint.op, 0), planar(axisPoint.op, 1)
	dx, dy := planar(axisDirection.op, 0), planar(axisDirection.op, 1)
	norm := sqrtOp(addOp(mulOp(dx, dx), mulOp(dy, dy)))
	dx, dy = divOp(dx, norm), divOp(dy, norm)
	var alongLo, alongHi *program.Operand
	out := num(0)
	for _, x := range []*program.Operand{b.lo[0], b.hi[0]} {
		for _, y := range []*program.Operand{b.lo[1], b.hi[1]} {
			ox, oy := subOp(x, ax), subOp(y, ay)
			along := addOp(mulOp(ox, dx), mulOp(oy, dy))
			if alongLo == nil {
				alongLo, alongHi = along, along
			} else {
				alongLo, alongHi = minOp(alongLo, along), maxOp(alongHi, along)
			}
			out = maxOp(out, subOp(mulOp(ox, dy), mulOp(oy, dx)))
		}
	}
	bx := &bbox{}
	for i, d := range []*program.Operand{dx, dy} {
		a := ax
		if i == 1 {
			a = ay
		}
		// The cylinder's discs stand across the line, so they reach
		// out * sqrt(1 - d[i]^2) along axis i: out times the other component.
		across := mulOp(out, absOp(dy))
		if i == 1 {
			across = mulOp(out, absOp(dx))
		}
		e0, e1 := mulOp(d, alongLo), mulOp(d, alongHi)
		bx.lo[i] = subOp(addOp(a, minOp(e0, e1)), across)
		bx.hi[i] = addOp(addOp(a, maxOp(e0, e1)), across)
	}
	bx.lo[2], bx.hi[2] = negOp(out), out
	return boxVal{box: bx, growth: profile.growth, limit: profile.limit}
}

// planar is component i of a point of the sketch plane, a list of two.
func planar(v *program.Operand, i int) *program.Operand {
	if e, ok := v.T.(*program.Operand_Expression); ok &&
		e.Expression.Operation == program.Operation_COLLECT && len(e.Expression.Operands) == 2 {
		return e.Expression.Operands[i]
	}
	return nary(program.Operation_ACCESS, v, ref(axisNames[i]))
}

// primitiveBox is the box of a builtin primitive (3d.gsl). Every one here is an
// exact distance outside its box, so its growth is 1, except the octahedron,
// whose field is a plane through abs(p) and reads 1/sqrt(3) of the distance
// past a vertex.
func primitiveBox(name string, vs []boxVal) (boxVal, bool) {
	ops, ok := opsOf(vs)
	switch {
	case name == "sphere" && len(vs) == 1:
		if !ok {
			return boxVal{}, true
		}
		r := ops[0]
		return fieldBox(around(zeroVec(), r), nil), true
	case name == "sphere" && len(vs) == 2:
		if !ok {
			return boxVal{}, true
		}
		return fieldBox(around(ops[0], ops[1]), nil), true
	case name == "torus" && len(vs) == 2:
		if !ok {
			return boxVal{}, true
		}
		outer := addOp(ops[0], ops[1])
		bx := &bbox{}
		for i := 0; i < 3; i++ {
			e := outer
			if i == 2 {
				e = ops[1]
			}
			bx.lo[i], bx.hi[i] = negOp(e), e
		}
		return fieldBox(bx, nil), true
	case name == "box" && len(vs) == 2:
		if !ok {
			return boxVal{}, true
		}
		return fieldBox(between(ops[0], ops[1], nil), nil), true
	case name == "boxCentered" && len(vs) == 2:
		if !ok {
			return boxVal{}, true
		}
		bx := &bbox{}
		for i := 0; i < 3; i++ {
			c := component(ops[0], i)
			h := mulOp(absOp(component(ops[1], i)), num(0.5))
			bx.lo[i], bx.hi[i] = subOp(c, h), addOp(c, h)
		}
		return fieldBox(bx, nil), true
	case (name == "cylinderCapped" || name == "cylinderRounded") && len(vs) == 3:
		if !ok {
			return boxVal{}, true
		}
		return fieldBox(between(ops[0], ops[1], ops[2]), nil), true
	case name == "lineSegment" && len(vs) == 2:
		if !ok {
			return boxVal{}, true
		}
		return fieldBox(between(ops[0], ops[1], nil), nil), true
	case (name == "coneCapped" || name == "coneRounded") && len(vs) == 4:
		if !ok {
			return boxVal{}, true
		}
		return fieldBox(between(ops[0], ops[1], maxOp(ops[2], ops[3])), nil), true
	case (name == "cylinderX" || name == "cylinderY" || name == "cylinderZ") && len(vs) == 2:
		if !ok {
			return boxVal{}, true
		}
		return fieldBox(slab(ops[0], ops[1], axisOf(name)), nil), true
	case (name == "axisX" || name == "axisY" || name == "axisZ") && len(vs) == 1:
		if !ok {
			return boxVal{}, true
		}
		return fieldBox(slab(ops[0], num(0), axisOf(name)), nil), true
	case name == "octahedron" && len(vs) == 1:
		if !ok {
			return boxVal{}, true
		}
		return fieldBox(around(zeroVec(), ops[0]), num(1/math.Sqrt(3))), true
	}
	return boxVal{}, false
}

// ArgsCall answers cropToBox(shape, lo, hi), which a declared name.box applies
// with the box grown by boxPad, from its corners alone: the declaration is the
// part's box, and the shape inside it, however many primitives it is built of,
// is never evaluated for it. The renderer cuts the surface to the box, so the
// surface is inside it whatever the shape is. The solid is left unbounded: the
// cut keeps the shape's sign outside the box (the mesher caps it instead), so
// an open shape's solid reaches past it, and only the shape could say where.
// Corners that depend on p bound nothing.
//
// It answers declaredBox, a standard library part's declaration, the same way,
// except that the solid is in the box too: nothing is cut, and the box is
// proven to hold the part.
func (dd *boxDomain) ArgsCall(name string, args []*program.Operand, eval func(*program.Operand) (Value, error)) (Value, bool, error) {
	if (name != cropToBoxName && name != declaredBoxName) || len(args) != 3 || dd.b.own[name] {
		return nil, false, nil
	}
	lo, err := eval(args[1])
	if err != nil {
		return nil, true, err
	}
	hi, err := eval(args[2])
	if err != nil {
		return nil, true, err
	}
	l, h := lo.(boxVal), hi.(boxVal)
	if l.op == nil || h.op == nil {
		return boxVal{}, true, nil
	}
	if name == declaredBoxName {
		return boxVal{box: between(l.op, h.op, ref(boxPadName))}, true, nil
	}
	return boxVal{surf: between(l.op, h.op, ref(boxPadName))}, true, nil
}

// boxPadName is builtin's padding for a declared box's crop.
const boxPadName = "boxPad"

// combinerBox is the box of a builtin smooth combiner (booleans.gsl), read from
// its sides. Every intersection here is at least the sharp max of its sides, so
// it only removes material; a round union adds material only within k of every
// side, so its solid is inside the union grown by k. A fillet is a surface where
// no side is 0, so none of them keeps a surface box of its own.
func combinerBox(name string, vs []boxVal) (boxVal, bool) {
	sides, k, ok := combinerSides(name, vs)
	if !ok {
		return boxVal{}, false
	}
	switch name {
	case "intersectRound", "intersectChamfer":
		return solidOnly(intersectionOf(sides)), true
	case "intersectExponential":
		// The peak plus k / e times the log of a sum of at least 1: the sharp
		// max or more only while k is not negative.
		if v, ok := literal(k); !ok || v < 0 {
			return boxVal{}, true
		}
		return solidOnly(intersectionOf(sides)), true
	case "differenceRound":
		return solidOnly(sides[0]), true
	case "unionRound":
		if k == nil {
			return boxVal{}, true
		}
		u := unionOf(sides)
		return offset(u, maxOp(k, num(0))), true
	}
	return boxVal{}, false
}

// combinerSides splits a combiner's arguments into its sides and fillet radius,
// for the forms (d1, d2, k), (d1, d2, d3, k), (list, k) and, for a difference,
// (target, tool, k).
func combinerSides(name string, vs []boxVal) ([]boxVal, *program.Operand, bool) {
	switch name {
	case "intersectExponential":
		if len(vs) == 2 && len(vs[0].elems) > 0 {
			return vs[0].elems, vs[1].op, true
		}
	case "intersectRound", "unionRound", "intersectChamfer":
		switch len(vs) {
		case 2:
			if len(vs[0].elems) == 0 {
				return nil, nil, false
			}
			return vs[0].elems, vs[1].op, true
		case 3, 4:
			return vs[:len(vs)-1], vs[len(vs)-1].op, true
		}
	case "differenceRound":
		if len(vs) == 3 {
			return vs[:1], vs[2].op, true
		}
	}
	return nil, nil, false
}

// fieldBox is a field value with the given box and growth (nil is 1).
func fieldBox(bx *bbox, growth *program.Operand) boxVal {
	return boxVal{box: bx, growth: growth}
}

// withoutOp keeps a value's boxes and growth but not its operand, as when it is
// one side of a combiner whose result depends on p.
func withoutOp(v boxVal) boxVal {
	return boxVal{box: v.box, surf: v.surf, growth: v.growth, limit: v.limit}
}

// solidOnly keeps a value's solid box, and drops the tighter surface box it may
// have, for a field whose surface is no longer its own.
func solidOnly(v boxVal) boxVal { return boxVal{box: v.box, growth: v.growth, limit: v.limit} }

// unionOf is the box of min over the values: the union of their boxes, which is
// unbounded if any is. The growth is the least of the sides'. min is 0 only
// where a side is, so its surface is inside the union of theirs.
func unionOf(vs []boxVal) boxVal {
	surf := surfaceUnion(vs)
	var out *bbox
	limit := 0.0
	for _, v := range vs {
		if v.box == nil {
			return boxVal{surf: surf}
		}
		limit = leastLimit(limit, v.limit)
		if v.box.empty {
			continue
		}
		if out == nil {
			cp := *v.box
			out = &cp
			continue
		}
		for i := 0; i < 3; i++ {
			out.lo[i] = pairOrNil(out.lo[i], v.box.lo[i], minOp)
			out.hi[i] = pairOrNil(out.hi[i], v.box.hi[i], maxOp)
		}
	}
	if out == nil {
		return boxVal{box: &bbox{empty: true}, surf: surf, limit: limit}
	}
	return boxVal{box: out, surf: surf, growth: leastGrowth(vs), limit: limit}
}

// intersectionOf is the box of max over the values: the intersection of their
// boxes. An unbounded side leaves the others' bounds. So does a positive
// constant: max(d, v) has no solid below v and is d's solid above it, so d's
// box holds at every offset.
func intersectionOf(vs []boxVal) boxVal {
	surf := surfaceIntersection(vs)
	var out *bbox
	limit := 0.0
	for _, v := range vs {
		if v.box == nil || v.box.empty {
			continue
		}
		limit = leastLimit(limit, v.limit)
		if out == nil {
			cp := *v.box
			out = &cp
			continue
		}
		for i := 0; i < 3; i++ {
			out.lo[i] = eitherOrBoth(out.lo[i], v.box.lo[i], maxOp)
			out.hi[i] = eitherOrBoth(out.hi[i], v.box.hi[i], minOp)
		}
	}
	if out == nil {
		return boxVal{surf: surf}
	}
	return boxVal{box: out, surf: surf, growth: leastGrowth(vs), limit: limit}
}

// surfaceUnion is the surface box of min over the values, or nil when none of
// them knows a surface tighter than its solid, so that the union of the solids
// already says it.
func surfaceUnion(vs []boxVal) *bbox {
	if !anySurface(vs) {
		return nil
	}
	var out *bbox
	for _, v := range vs {
		s := v.surface()
		if s == nil {
			return nil
		}
		if s.empty {
			continue
		}
		if out == nil {
			cp := *s
			out = &cp
			continue
		}
		for i := 0; i < 3; i++ {
			out.lo[i] = pairOrNil(out.lo[i], s.lo[i], minOp)
			out.hi[i] = pairOrNil(out.hi[i], s.hi[i], maxOp)
		}
	}
	if out == nil {
		return &bbox{empty: true}
	}
	return out
}

// surfaceIntersection is the surface box of max over the values: max is 0 only
// where one side is 0 and every other side is solid, so it is the union, over
// the sides, of each one's surface inside the others' solids. It is nil when
// none of them knows a surface tighter than its solid, and the intersection of
// the solids already says it.
func surfaceIntersection(vs []boxVal) *bbox {
	if !anySurface(vs) {
		return nil
	}
	parts := make([]boxVal, len(vs))
	for i, v := range vs {
		part := v.surface()
		for j, w := range vs {
			if j != i {
				part = intersectBoxes(part, w.box)
			}
		}
		parts[i] = boxVal{box: part}
	}
	return unionOf(parts).box
}

// anySurface reports whether any value keeps a surface box of its own.
func anySurface(vs []boxVal) bool {
	for _, v := range vs {
		if v.surf != nil {
			return true
		}
	}
	return false
}

// intersectBoxes is the intersection of two boxes, where nil (unbounded) yields
// to the other and empty absorbs.
func intersectBoxes(a, b *bbox) *bbox {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	if a.empty || b.empty {
		return &bbox{empty: true}
	}
	out := &bbox{}
	for i := 0; i < 3; i++ {
		out.lo[i] = eitherOrBoth(a.lo[i], b.lo[i], maxOp)
		out.hi[i] = eitherOrBoth(a.hi[i], b.hi[i], minOp)
	}
	return out
}

// offset is the box of d - c: grown by c / growth when c is positive, and kept
// when it is not, since shrinking an underestimated field's solid is not safe to
// assume. The offset also spends c of the box's limit. Its surface is d's level
// set at c, which only the solid's box bounds.
func offset(d boxVal, c *program.Operand) boxVal {
	if d.box == nil || d.box.empty {
		return boxVal{}
	}
	amount := c
	limit := d.limit
	if v, ok := literal(c); ok {
		if v <= 0 {
			return solidOnly(d)
		}
		if limit > 0 {
			if limit -= v; limit <= 0 {
				return boxVal{}
			}
		}
	} else {
		if limit > 0 {
			return boxVal{}
		}
		amount = maxOp(c, num(0))
	}
	if d.growth != nil {
		amount = divOp(amount, d.growth)
	}
	bx := &bbox{}
	for i := 0; i < 3; i++ {
		if d.box.lo[i] != nil {
			bx.lo[i] = subOp(d.box.lo[i], amount)
		}
		if d.box.hi[i] != nil {
			bx.hi[i] = addOp(d.box.hi[i], amount)
		}
	}
	return boxVal{box: bx, growth: d.growth, limit: limit}
}

// scaled is the box of k * d (or d / k): the same solid when k is a positive
// literal, with the growth multiplied (or divided) by k. A negative k turns the
// solid inside out, and one of unknown sign might; any nonzero k keeps the
// surface.
func scaled(d boxVal, k *program.Operand, divide bool) boxVal {
	v, ok := literal(k)
	if !ok || v == 0 {
		return boxVal{}
	}
	if v < 0 {
		return boxVal{surf: d.surface()}
	}
	if d.box == nil || d.box.empty {
		return boxVal{surf: d.surf}
	}
	limit := d.limit
	g := d.growth
	if g == nil {
		g = num(1)
	}
	if divide {
		g, limit = divOp(g, k), limit/v
	} else {
		g, limit = mulOp(g, k), limit*v
	}
	return boxVal{box: d.box, surf: d.surf, growth: g, limit: limit}
}

// leastLimit is the tighter of two limits, where 0 is none.
func leastLimit(a, b float64) float64 {
	if a == 0 {
		return b
	}
	if b == 0 {
		return a
	}
	return math.Min(a, b)
}

// leastGrowth is the smallest growth among the sides that have a box.
func leastGrowth(vs []boxVal) *program.Operand {
	var out *program.Operand
	any := false
	for _, v := range vs {
		if v.box == nil || v.box.empty {
			continue
		}
		g := v.growth
		if g == nil {
			g = num(1)
		}
		if !any {
			out, any = g, true
			continue
		}
		out = minOp(out, g)
	}
	if v, ok := literal(out); ok && v == 1 {
		return nil
	}
	return out
}

// around is the box c ± r.
func around(c, r *program.Operand) *bbox {
	bx := &bbox{}
	for i := 0; i < 3; i++ {
		ci := component(c, i)
		bx.lo[i], bx.hi[i] = subOp(ci, r), addOp(ci, r)
	}
	return bx
}

// between is the box spanning points a and b, padded by r (nil for none).
func between(a, b, r *program.Operand) *bbox {
	bx := &bbox{}
	for i := 0; i < 3; i++ {
		ai, bi := component(a, i), component(b, i)
		lo, hi := minOp(ai, bi), maxOp(ai, bi)
		if r != nil {
			lo, hi = subOp(lo, r), addOp(hi, r)
		}
		bx.lo[i], bx.hi[i] = lo, hi
	}
	return bx
}

// slab is the box of an infinite cylinder of radius r through c along axis:
// unbounded along it, c ± r across it.
func slab(c, r *program.Operand, axis int) *bbox {
	bx := around(c, r)
	bx.lo[axis], bx.hi[axis] = nil, nil
	return bx
}

func axisOf(name string) int {
	switch name[len(name)-1] {
	case 'X':
		return 0
	case 'Y':
		return 1
	}
	return 2
}

// --- transforms -------------------------------------------------------------

// transformPath is the import path of the shape-transform library, whose
// calls carry a box through the remap inside them.
const transformPath = "transform"

// QualifiedCall reads the transform library's calls before the Engine inlines
// them: inlined, each is a remap, which would clear the box.
func (dd *boxDomain) QualifiedCall(path, name string, args []Value) (Value, bool, error) {
	if path != transformPath {
		return nil, false, nil
	}
	vs := boxVals(args)
	if len(vs) == 0 || vs[0].op != nil {
		return nil, false, nil
	}
	shape, rest := vs[0], vs[1:]
	restOps, ok := opsOf(rest)
	if !ok {
		return boxVal{}, true, nil
	}
	switch {
	case name == "translate" && len(rest) == 1:
		return translated(shape, restOps[0]), true, nil
	case name == "rotateX" && len(rest) == 1:
		return turned(shape, 1, 2, restOps[0]), true, nil
	case name == "rotateY" && len(rest) == 1:
		return turned(shape, 2, 0, restOps[0]), true, nil
	case name == "rotateZ" && len(rest) == 1:
		return turned(shape, 0, 1, restOps[0]), true, nil
	case name == "rotate" && len(rest) == 3:
		return rotated(shape, restOps[0], restOps[1], restOps[2]), true, nil
	case name == "scale" && len(rest) == 1:
		return scaledAbout(shape, restOps[0]), true, nil
	case name == "affine" && len(rest) == 1:
		return carried(shape, restOps[0]), true, nil
	}
	return nil, false, nil
}

// mapBox carries a box through a transform's action f on boxes, which returns
// nil when the result is unbounded. Unbounded and empty boxes stay as they are.
func mapBox(bx *bbox, f func(*bbox) *bbox) *bbox {
	if bx == nil || bx.empty {
		return bx
	}
	return f(bx)
}

// translated is the box of a shape moved by offset: the box moved with it.
func translated(v boxVal, offset *program.Operand) boxVal {
	move := func(b *bbox) *bbox {
		bx := &bbox{}
		for i := 0; i < 3; i++ {
			o := component(offset, i)
			if b.lo[i] != nil {
				bx.lo[i] = addOp(b.lo[i], o)
			}
			if b.hi[i] != nil {
				bx.hi[i] = addOp(b.hi[i], o)
			}
		}
		return bx
	}
	surf := mapBox(v.surf, move)
	if v.box == nil {
		return boxVal{surf: surf}
	}
	return boxVal{box: mapBox(v.box, move), surf: surf, growth: v.growth, limit: v.limit}
}

// turned is the box of a shape turned by angle in the plane of axes i and j,
// counterclockwise from i toward j: the axis-aligned hull of the turned box.
// The third axis keeps its bounds; the two turned axes are unbounded if either
// was. The hull of a box grown by r grows by up to r(|cos| + |sin|), which
// divides the growth.
func turned(v boxVal, i, j int, angle *program.Operand) boxVal {
	c, s := cosOp(angle), sinOp(angle)
	ac, as := absOp(c), absOp(s)
	turn := func(b *bbox) *bbox {
		bx := *b
		if bx.lo[i] == nil || bx.hi[i] == nil || bx.lo[j] == nil || bx.hi[j] == nil {
			bx.lo[i], bx.hi[i], bx.lo[j], bx.hi[j] = nil, nil, nil, nil
			return &bx
		}
		ci, hi := centerHalf(b, i)
		cj, hj := centerHalf(b, j)
		ni := subOp(mulOp(c, ci), mulOp(s, cj))
		nj := addOp(mulOp(s, ci), mulOp(c, cj))
		wi := addOp(mulOp(ac, hi), mulOp(as, hj))
		wj := addOp(mulOp(as, hi), mulOp(ac, hj))
		bx.lo[i], bx.hi[i] = subOp(ni, wi), addOp(ni, wi)
		bx.lo[j], bx.hi[j] = subOp(nj, wj), addOp(nj, wj)
		return &bx
	}
	surf := mapBox(v.surf, turn)
	if v.box == nil {
		return boxVal{surf: surf}
	}
	if v.box.empty {
		return boxVal{box: v.box, surf: surf, growth: v.growth, limit: v.limit}
	}
	return boxVal{box: turn(v.box), surf: surf, growth: divGrowth(v.growth, addOp(ac, as)), limit: v.limit}
}

// rotated is the box of a shape turned by angle about the line through origin
// along direction (Rodrigues' rotation, right-handed).
func rotated(v boxVal, origin, direction, angle *program.Operand) boxVal {
	k := normalizeOp(direction)
	kx, ky, kz := component(k, 0), component(k, 1), component(k, 2)
	kv := [3]*program.Operand{kx, ky, kz}
	c, s := cosOp(angle), sinOp(angle)
	oneMinusC := subOp(num(1), c)
	cross := [3][3]*program.Operand{
		{num(0), negOp(kz), ky},
		{kz, num(0), negOp(kx)},
		{negOp(ky), kx, num(0)},
	}
	var a [3][3]*program.Operand
	for r := 0; r < 3; r++ {
		for col := 0; col < 3; col++ {
			e := mulOp(mulOp(oneMinusC, kv[r]), kv[col])
			e = addOp(e, mulOp(s, cross[r][col]))
			if r == col {
				e = addOp(e, c)
			}
			a[r][col] = e
		}
	}
	// p = origin + R (q - origin) = R q + (origin - R origin).
	var t [3]*program.Operand
	for r := 0; r < 3; r++ {
		ro := num(0)
		for col := 0; col < 3; col++ {
			ro = addOp(ro, mulOp(a[r][col], component(origin, col)))
		}
		t[r] = subOp(component(origin, r), ro)
	}
	return linearly(v, a, t, nil)
}

// scaledAbout is the box of a shape scaled by factor about the origin. The field
// scales with it, so offsets grow the box as before; a limit on them scales too.
func scaledAbout(v boxVal, factor *program.Operand) boxVal {
	k, isLiteral := literal(factor)
	if isLiteral && k <= 0 {
		return boxVal{}
	}
	scale := func(b *bbox) *bbox {
		bx := &bbox{}
		for i := 0; i < 3; i++ {
			if b.lo[i] != nil && b.hi[i] != nil {
				lo, hi := mulOp(b.lo[i], factor), mulOp(b.hi[i], factor)
				bx.lo[i], bx.hi[i] = minOp(lo, hi), maxOp(lo, hi)
			}
		}
		return bx
	}
	surf := mapBox(v.surf, scale)
	if v.box == nil {
		return boxVal{surf: surf}
	}
	if v.box.empty {
		return boxVal{box: v.box, surf: surf, growth: v.growth, limit: v.limit}
	}
	limit := v.limit
	if isLiteral {
		limit *= k
	} else if limit > 0 {
		return boxVal{surf: surf}
	}
	return boxVal{box: scale(v.box), surf: surf, growth: v.growth, limit: limit}
}

// carried is the box of a shape carried by the 4x4 affine matrix m, whose
// columns are m[0..3]: the linear block A (A[r][c] = m[c][r]) and the
// translation m[3]. The field is scaled by m's smallest scale.
func carried(v boxVal, m *program.Operand) boxVal {
	var a [3][3]*program.Operand
	var t [3]*program.Operand
	for c := 0; c < 3; c++ {
		col := element(m, c)
		for r := 0; r < 3; r++ {
			a[r][c] = element(col, r)
		}
	}
	tc := element(m, 3)
	for r := 0; r < 3; r++ {
		t[r] = element(tc, r)
	}
	var scale *program.Operand
	for c := 0; c < 3; c++ {
		n := num(0)
		for r := 0; r < 3; r++ {
			n = addOp(n, mulOp(a[r][c], a[r][c]))
		}
		l := sqrtOp(n)
		if scale == nil {
			scale = l
		} else {
			scale = minOp(scale, l)
		}
	}
	return linearly(v, a, t, scale)
}

// linearly is the box of a shape carried by p = A q + t: centers map through A,
// half-widths through |A|. The field is scaled by fieldScale (nil for 1). Every
// axis must be bounded, since A may mix them all.
func linearly(v boxVal, a [3][3]*program.Operand, t [3]*program.Operand, fieldScale *program.Operand) boxVal {
	carry := func(b *bbox) *bbox {
		var cen, half [3]*program.Operand
		for i := 0; i < 3; i++ {
			if b.lo[i] == nil || b.hi[i] == nil {
				return nil
			}
			cen[i], half[i] = centerHalf(b, i)
		}
		bx := &bbox{}
		for r := 0; r < 3; r++ {
			c := t[r]
			w := num(0)
			for col := 0; col < 3; col++ {
				c = addOp(c, mulOp(a[r][col], cen[col]))
				w = addOp(w, mulOp(absOp(a[r][col]), half[col]))
			}
			bx.lo[r], bx.hi[r] = subOp(c, w), addOp(c, w)
		}
		return bx
	}
	surf := mapBox(v.surf, carry)
	if v.box == nil {
		return boxVal{surf: surf}
	}
	if v.box.empty {
		return boxVal{box: v.box, surf: surf, growth: v.growth, limit: v.limit}
	}
	bx := carry(v.box)
	if bx == nil {
		return boxVal{surf: surf}
	}
	limit := v.limit
	if fieldScale != nil && limit > 0 {
		k, ok := literal(fieldScale)
		if !ok {
			return boxVal{surf: surf}
		}
		limit *= k
	}
	var rowSum *program.Operand
	for r := 0; r < 3; r++ {
		sum := num(0)
		for col := 0; col < 3; col++ {
			sum = addOp(sum, absOp(a[r][col]))
		}
		if rowSum == nil {
			rowSum = sum
		} else {
			rowSum = maxOp(rowSum, sum)
		}
	}
	g := divGrowth(v.growth, rowSum)
	if fieldScale != nil {
		g = mulGrowth(g, fieldScale)
	}
	return boxVal{box: bx, surf: surf, growth: g, limit: limit}
}

func centerHalf(bx *bbox, i int) (*program.Operand, *program.Operand) {
	return mulOp(addOp(bx.lo[i], bx.hi[i]), num(0.5)), mulOp(subOp(bx.hi[i], bx.lo[i]), num(0.5))
}

// element is the i-th entry of a list or vector operand, read straight out of a
// literal when it is one.
func element(v *program.Operand, i int) *program.Operand {
	if e, ok := v.T.(*program.Operand_Expression); ok &&
		e.Expression.Operation == program.Operation_COLLECT && i < len(e.Expression.Operands) {
		return e.Expression.Operands[i]
	}
	return nary(program.Operation_INDEX, v, num(float64(i)))
}

func divGrowth(g, by *program.Operand) *program.Operand {
	if g == nil {
		g = num(1)
	}
	return nilIfOne(divOp(g, by))
}

func mulGrowth(g, by *program.Operand) *program.Operand {
	if g == nil {
		g = num(1)
	}
	return nilIfOne(mulOp(g, by))
}

func nilIfOne(g *program.Operand) *program.Operand {
	if v, ok := literal(g); ok && v == 1 {
		return nil
	}
	return g
}

func sqrtOp(a *program.Operand) *program.Operand {
	if v, ok := literal(a); ok && v >= 0 {
		return num(math.Sqrt(v))
	}
	return sqrt(a)
}

func cosOp(a *program.Operand) *program.Operand {
	if v, ok := literal(a); ok {
		return num(math.Cos(v))
	}
	return un(program.Operation_COS, a)
}

func sinOp(a *program.Operand) *program.Operand {
	if v, ok := literal(a); ok {
		return num(math.Sin(v))
	}
	return un(program.Operation_SIN, a)
}

// normalizeOp is a unit vector along v, computed here when v is a vector
// literal.
func normalizeOp(v *program.Operand) *program.Operand {
	if e, ok := v.T.(*program.Operand_Expression); ok &&
		e.Expression.Operation == program.Operation_COLLECT && len(e.Expression.Operands) == 3 {
		var xs [3]float64
		all := true
		for i, o := range e.Expression.Operands {
			x, ok := literal(o)
			if !ok {
				all = false
				break
			}
			xs[i] = x
		}
		if l := math.Sqrt(xs[0]*xs[0] + xs[1]*xs[1] + xs[2]*xs[2]); all && l > 0 {
			return nary(program.Operation_COLLECT, num(xs[0]/l), num(xs[1]/l), num(xs[2]/l))
		}
	}
	return call("normalize", v)
}

// --- literal folding --------------------------------------------------------

// A box's corners are GSL expressions of the values the field is built from,
// and a library part's values come from table lookups: catalog.panHeadDia(6.0)
// is a sum over every row of a clamp of a sign of the distance to its key.
// Left as expressions they multiply through every transform and bury the
// program in arithmetic the backends fold much later, so arithmetic on
// literals, and the pure builtins lookups use, fold here.

// foldOperation computes an operation whose operands are all literals.
func foldOperation(operation program.Operation, vs []boxVal) (float64, bool) {
	xs := make([]float64, len(vs))
	for i, v := range vs {
		x, ok := literal(v.op)
		if !ok {
			return 0, false
		}
		xs[i] = x
	}
	b2i := func(b bool) float64 {
		if b {
			return 1
		}
		return 0
	}
	one := func(f func(float64) float64) (float64, bool) {
		if len(xs) != 1 {
			return 0, false
		}
		return f(xs[0]), true
	}
	two := func(f func(a, b float64) float64) (float64, bool) {
		if len(xs) != 2 {
			return 0, false
		}
		return f(xs[0], xs[1]), true
	}
	fold := func(f func(a, b float64) float64) (float64, bool) {
		if len(xs) == 0 {
			return 0, false
		}
		acc := xs[0]
		for _, x := range xs[1:] {
			acc = f(acc, x)
		}
		return acc, true
	}
	switch operation {
	case program.Operation_ADD:
		return fold(func(a, b float64) float64 { return a + b })
	case program.Operation_SUB:
		return two(func(a, b float64) float64 { return a - b })
	case program.Operation_MUL:
		return fold(func(a, b float64) float64 { return a * b })
	case program.Operation_DIV:
		return two(func(a, b float64) float64 { return a / b })
	case program.Operation_MIN:
		return fold(math.Min)
	case program.Operation_MAX:
		return fold(math.Max)
	case program.Operation_NEG:
		return one(func(a float64) float64 { return -a })
	case program.Operation_ABS:
		return one(math.Abs)
	case program.Operation_SQUARE:
		return one(func(a float64) float64 { return a * a })
	case program.Operation_SQRT:
		return one(math.Sqrt)
	case program.Operation_FLOOR:
		return one(math.Floor)
	case program.Operation_CEIL:
		return one(math.Ceil)
	case program.Operation_EXP:
		return one(math.Exp)
	case program.Operation_LN:
		return one(math.Log)
	case program.Operation_POW:
		return two(math.Pow)
	case program.Operation_SIN:
		return one(math.Sin)
	case program.Operation_COS:
		return one(math.Cos)
	case program.Operation_TAN:
		return one(math.Tan)
	case program.Operation_ASIN:
		return one(math.Asin)
	case program.Operation_ACOS:
		return one(math.Acos)
	case program.Operation_ATAN:
		return one(math.Atan)
	case program.Operation_ATAN2:
		return two(math.Atan2)
	case program.Operation_GT:
		return two(func(a, b float64) float64 { return b2i(a > b) })
	case program.Operation_GE:
		return two(func(a, b float64) float64 { return b2i(a >= b) })
	case program.Operation_LT:
		return two(func(a, b float64) float64 { return b2i(a < b) })
	case program.Operation_LE:
		return two(func(a, b float64) float64 { return b2i(a <= b) })
	case program.Operation_EQ:
		return two(func(a, b float64) float64 { return b2i(a == b) })
	case program.Operation_NE:
		return two(func(a, b float64) float64 { return b2i(a != b) })
	}
	return 0, false
}

// foldBuiltin computes a pure builtin whose arguments are all literals.
func foldBuiltin(name string, vs []boxVal) (float64, bool) {
	xs := make([]float64, len(vs))
	for i, v := range vs {
		x, ok := literal(v.op)
		if !ok {
			return 0, false
		}
		xs[i] = x
	}
	switch {
	case name == "sign" && len(xs) == 1:
		switch {
		case xs[0] > 0:
			return 1, true
		case xs[0] < 0:
			return -1, true
		}
		return 0, true
	case name == "clamp" && len(xs) == 3:
		return math.Min(math.Max(xs[0], xs[1]), xs[2]), true
	case name == "fract" && len(xs) == 1:
		return xs[0] - math.Floor(xs[0]), true
	case name == "trunc" && len(xs) == 1:
		return math.Trunc(xs[0]), true
	case name == "radians" && len(xs) == 1:
		return xs[0] * math.Pi / 180, true
	case name == "degrees" && len(xs) == 1:
		return xs[0] * 180 / math.Pi, true
	case name == "exp2" && len(xs) == 1:
		return math.Exp2(xs[0]), true
	case name == "sinh" && len(xs) == 1:
		return math.Sinh(xs[0]), true
	case name == "cosh" && len(xs) == 1:
		return math.Cosh(xs[0]), true
	case name == "tanh" && len(xs) == 1:
		return math.Tanh(xs[0]), true
	}
	return 0, false
}

// --- operand helpers --------------------------------------------------------

func boxVals(args []Value) []boxVal {
	out := make([]boxVal, len(args))
	for i, a := range args {
		out[i] = a.(boxVal)
	}
	return out
}

// opsOf returns every value's operand, or false if any depends on p.
func opsOf(vs []boxVal) ([]*program.Operand, bool) {
	ops := make([]*program.Operand, len(vs))
	for i, v := range vs {
		if v.op == nil {
			return nil, false
		}
		ops[i] = v.op
	}
	return ops, true
}

// opsOrNil is the operation over the values' operands, or nil if any depends
// on p.
func opsOrNil(operation program.Operation, vs []boxVal) *program.Operand {
	ops, ok := opsOf(vs)
	if !ok {
		return nil
	}
	return nary(operation, ops...)
}

func literal(op *program.Operand) (float64, bool) {
	if op == nil {
		return 0, false
	}
	v, ok := op.T.(*program.Operand_Value)
	if !ok {
		return 0, false
	}
	return v.Value, true
}

var axisNames = [3]string{program.XVarName, program.YVarName, program.ZVarName}

// component is the i-th component of a vec3 operand, read straight out of a
// vector literal when it is one. A scalar literal broadcasts, as it does in the
// primitive's own p - center.
func component(v *program.Operand, i int) *program.Operand {
	if _, ok := literal(v); ok {
		return v
	}
	if e, ok := v.T.(*program.Operand_Expression); ok &&
		e.Expression.Operation == program.Operation_COLLECT && len(e.Expression.Operands) == 3 {
		return e.Expression.Operands[i]
	}
	return nary(program.Operation_ACCESS, v, ref(axisNames[i]))
}

func zeroVec() *program.Operand {
	return nary(program.Operation_COLLECT, num(0), num(0), num(0))
}

// pairOrNil combines two bounds, where nil (unbounded) absorbs.
func pairOrNil(a, b *program.Operand, f func(a, b *program.Operand) *program.Operand) *program.Operand {
	if a == nil || b == nil {
		return nil
	}
	return f(a, b)
}

// eitherOrBoth combines two bounds, where nil (unbounded) yields to the other.
func eitherOrBoth(a, b *program.Operand, f func(a, b *program.Operand) *program.Operand) *program.Operand {
	if a == nil {
		return b
	}
	if b == nil {
		return a
	}
	return f(a, b)
}

// The arithmetic below folds literals as it goes, so a constant box reads as
// numbers rather than an expression the backends fold later.

func foldBin(a, b *program.Operand, f func(x, y float64) float64, o program.Operation) *program.Operand {
	if x, ok := literal(a); ok {
		if y, ok := literal(b); ok {
			return num(f(x, y))
		}
	}
	return bin(o, a, b)
}

func addOp(a, b *program.Operand) *program.Operand {
	if v, ok := literal(b); ok && v == 0 {
		return a
	}
	return foldBin(a, b, func(x, y float64) float64 { return x + y }, program.Operation_ADD)
}

func subOp(a, b *program.Operand) *program.Operand {
	if v, ok := literal(b); ok && v == 0 {
		return a
	}
	return foldBin(a, b, func(x, y float64) float64 { return x - y }, program.Operation_SUB)
}

func mulOp(a, b *program.Operand) *program.Operand {
	return foldBin(a, b, func(x, y float64) float64 { return x * y }, program.Operation_MUL)
}

func divOp(a, b *program.Operand) *program.Operand {
	if v, ok := literal(b); ok && v == 1 {
		return a
	}
	return foldBin(a, b, func(x, y float64) float64 { return x / y }, program.Operation_DIV)
}

func minOp(a, b *program.Operand) *program.Operand {
	return foldBin(a, b, math.Min, program.Operation_MIN)
}

func maxOp(a, b *program.Operand) *program.Operand {
	return foldBin(a, b, math.Max, program.Operation_MAX)
}

func negOp(a *program.Operand) *program.Operand {
	if v, ok := literal(a); ok {
		return num(-v)
	}
	return neg(a)
}

func absOp(a *program.Operand) *program.Operand {
	if v, ok := literal(a); ok {
		return num(math.Abs(v))
	}
	return un(program.Operation_ABS, a)
}
