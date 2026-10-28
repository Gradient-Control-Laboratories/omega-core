package property

import (
	"fmt"

	"gradientcontrol.com/omega/core/program"
)

// secondInf is the sentinel "no second-closest surface" distance: large and
// finite, so it loses every comparison against a real distance and still
// renders safely in the backends. A single primitive seeds its second-closest
// to this.
//
// It is the only numeric constant these expansions need, and it must never
// reach a runtime select. Backends lower `cond ? a : b` algebraically as
// `b + cond*(a-b)`, and in f32 a real value minus 1e30 rounds to exactly -1e30,
// so a *taken* branch would evaluate to zero. Eliminate it at build time (see
// isInf) or add it; never select on it.
//
// The two guards that once wanted a small tuned constant no longer do. The
// blend piece rule needs no floor under its root, because rejecting an empty
// blend already pushes that root's argument to secondInf before the root is
// taken (see blend.second). A derivative dividing by a vector norm needs no
// epsilon, because a comparison names the zero case exactly (see nonZero).
// That is worth keeping: a constant here would also have to clear the WGSL
// backend's printable floor, since it emits with %f and anything below about
// 5e-7 becomes 0.0 in the shader while still holding in the mesher.
const secondInf = 1e30

// pair is the second-closest Domain's value: the closest primitive-surface
// distance (first) paired with the second-closest (second), both GSL operands.
// For a union `first` is the ordinary distance; at an intersection it is the
// nearer of the two binding surfaces, not the SDF (the chosen order-statistic
// reading — MIN and MAX both keep the two smallest). A single surface carries
// second == secondInf.
type pair struct {
	first  *program.Operand
	second *program.Operand
	// elems is the element pairs when this value is a list literal (e.g. the
	// [a,b,c] passed to unionRound(list, k)); nil otherwise. Only the smooth-
	// combiner interception (Call) reads it, so every other path ignores it.
	elems []pair
}

// second is the second-closest-distance self-computing property. It expands the
// `.second` operator: over the primitive surfaces reachable through min/max (and
// affine offsets), the distance to the second-nearest.
type second struct {
	decls      map[string]*program.Operand
	multiDecls map[string]multiBinding        // names bound by a destructuring assignment
	funcs      map[string][]*program.Function // incl. overloads and dot-imported (omega) functions
}

func (s *second) engine() *Engine {
	return &Engine{dom: &secondDomain{s: s}, decls: s.decls, multiDecls: s.multiDecls, funcs: s.funcs}
}

// expandSecond expands a SECONDCLOSEST operand (.second) to its second-closest
// distance field.
func (s *second) expandSecond(operands []*program.Operand) (*program.Operand, error) {
	if len(operands) != 1 {
		return nil, fmt.Errorf("second-closest expects 1 operand, got %d", len(operands))
	}
	v, err := s.engine().Eval(operands[0])
	if err != nil {
		return nil, err
	}
	return v.(pair).second, nil
}

// secondDomain evaluates a field into a {first, second} pair.
type secondDomain struct {
	s *second
}

func inf() *program.Operand { return num(secondInf) }

// isInf reports whether op is the second-closest sentinel — a value carrying no
// second surface. It distinguishes a scalar/leaf (an offset operand) from a
// genuine combiner result when propagating `.second` through an offset.
func isInf(op *program.Operand) bool {
	v, ok := op.T.(*program.Operand_Value)
	return ok && v.Value == secondInf
}

// leaf makes a single-surface (or scalar) pair: a real first, no second.
func leaf(first *program.Operand) pair { return pair{first: first, second: inf()} }

func (dd *secondDomain) Literal(v float64) Value { return leaf(num(v)) }

func (dd *secondDomain) ImpliedVar(name string) (Value, error) {
	switch name {
	case program.PVarName, program.TimeVarName, program.XVarName, program.YVarName, program.ZVarName:
		return leaf(ref(name)), nil
	}
	return nil, fmt.Errorf("not an implied variable: %q", name)
}

func (dd *secondDomain) Opaque(name string) Value { return leaf(ref(name)) }

func (dd *secondDomain) Collect(args []Value) Value {
	ps := pairs(args)
	// Keep the element pairs so a smooth list-combiner (unionRound(list, k), …)
	// can take their order-statistic; first stays the flattened vector/list for
	// any non-combiner use (a coordinate vector inside a primitive's distance).
	return pair{first: nary(program.Operation_COLLECT, firstsOf(ps)...), second: inf(), elems: ps}
}

func (dd *secondDomain) Access(base Value, selectors []*program.Operand) Value {
	b := base.(pair)
	return leaf(nary(program.Operation_ACCESS, append([]*program.Operand{b.first}, selectors...)...))
}

func (dd *secondDomain) Index(base Value, index *program.Operand) Value {
	b := base.(pair)
	return leaf(nary(program.Operation_INDEX, b.first, index))
}

// Select is a hard branch (ternary), which can act as a union: keep the chosen
// branch's pair, selecting by the condition's first (a boolean is a scalar).
func (dd *secondDomain) Select(cond Value, branches []Value) Value {
	c := cond.(pair)
	bs := pairs(branches)
	pArgs := append([]*program.Operand{c.first}, firstsOf(bs)...)
	sArgs := append([]*program.Operand{c.first}, secondsOf(bs)...)
	return pair{
		first:  nary(program.Operation_SELECT, pArgs...),
		second: nary(program.Operation_SELECT, sArgs...),
	}
}

// Named keeps the value part of a declaration live (a reference the backend
// re-derives, so a slider tracks its current value) while carrying the computed
// second-closest. Using the reference for `first` matches the ordinary SDF for a
// union or leaf; a named intersection reused inside an outer combiner is the one
// case where `first` (the SDF here) differs from the order-statistic min-fold.
func (dd *secondDomain) Named(name string, value Value) Value {
	v := value.(pair)
	return pair{first: ref(name), second: v.second, elems: v.elems}
}

func (dd *secondDomain) Operator(operation program.Operation, operands []*program.Operand) (*program.Operand, bool, error) {
	if operation == program.Operation_SURFACEU || operation == program.Operation_SURFACEV {
		// bar.u.second: a frame coordinate is a value on a surface, not a
		// distance to one, so there is no second-closest surface to find. The
		// sentinel it would otherwise read as must never reach a value.
		return nil, true, fmt.Errorf("a surface frame coordinate (.u, .v) has no second-closest surface; did you mean .second.u?")
	}
	if operation != program.Operation_SECONDCLOSEST {
		return nil, false, nil
	}
	op, err := dd.s.expandSecond(operands)
	return op, true, err
}

// OpaqueCall / OpaqueRef splice a qualified (map./noise./…) call or reference
// through unchanged as a single surface (no second-closest); the backend
// resolves the import. A map.rotateZX used as a remap coordinate is a coordinate
// given, not a surface combiner, so .second transports through it — mirroring the
// color property's opaque handling.
func (dd *secondDomain) OpaqueCall(pkgName, name string, args []Value) Value {
	return leaf(&program.Operand{T: &program.Operand_FnCall{FnCall: &program.FnCall{
		Reference: &program.Reference{Prefix: pkgName, Name: name}, Arguments: firstsOf(pairs(args))}}})
}

func (dd *secondDomain) OpaqueRef(pkgName, name string) Value {
	return leaf(&program.Operand{T: &program.Operand_Reference{Reference: &program.Reference{
		Prefix: pkgName, Name: name}}})
}

// Call intercepts the omega smooth *round* and *LP* boolean combiners. They are
// De Morgan wrappers around a sharp intersect plus blend arithmetic; folding
// through that body collapses the second-closest, so instead read the pieces of
// topology the combiner actually produces — the input surfaces AND the blended
// region between them — and report the nearest piece the point is not already
// on. Unrecognized calls fall through to normal inlining.
func (dd *secondDomain) Call(name string, args []Value) (Value, bool, error) {
	ps := pairs(args)
	b, ok := blendOf(name, ps)
	if !ok {
		return nil, false, nil
	}
	// first stays the combiner's own SDF: the actual rounded boundary.
	return pair{first: call(name, firstsOf(ps)...), second: b.secondOrSharp()}, true, nil
}

// secondOrSharp is second, except that a literal zero radius is the sharp
// combiner and takes its order statistic. The library spells a plain box as
// intersectRound(0), and at k = 0 the piece formula breaks down: with every
// term clamped the point sits in the sharp intersect, which at zero radius is
// the faces themselves, so re-adding face j names a piece |dⱼ| away. A hit
// point that lands a rounding error inside a face then reports that face as
// its own runner-up, and the face reads as an edge.
func (b blend) secondOrSharp() *program.Operand {
	if !isZeroLit(b.k) {
		return b.second()
	}
	return orderStat(b.sides, !b.union).second
}

// blend describes a recognized smooth boolean combiner: its surface operands,
// the fillet radius k, and the norm exponent lp (nil for the round family,
// which is L2). union records that the sides arrive in the union frame and must
// be De Morgan'd into the intersect frame the arithmetic below is written in.
type blend struct {
	sides []pair
	k     *program.Operand
	lp    *program.Operand
	union bool
}

// blendOf matches a round/LP combiner and pulls out its surfaces, fillet radius
// and exponent (tool negated for a difference), or ok=false if the call is not
// one — or is a list form whose elements were not preserved.
func blendOf(name string, args []pair) (blend, bool) {
	switch name {
	case "unionRound", "intersectRound":
		u := name == "unionRound"
		switch len(args) {
		case 2: // (list, k)
			if e, ok := listElems(args[0]); ok {
				return blend{sides: e, k: args[1].first, union: u}, true
			}
		case 3: // (d1, d2, k)
			return blend{sides: args[:2], k: args[2].first, union: u}, true
		case 4: // (d1, d2, d3, k)
			return blend{sides: args[:3], k: args[3].first, union: u}, true
		}
	case "unionLp", "intersectLp":
		if len(args) == 3 { // (list, k, lp)
			if e, ok := listElems(args[0]); ok {
				return blend{sides: e, k: args[1].first, lp: args[2].first, union: name == "unionLp"}, true
			}
		}
	case "differenceRound":
		if len(args) == 3 { // (target, tool, k)
			return blend{sides: []pair{args[0], negLeaf(args[1])}, k: args[2].first}, true
		}
	case "differenceLp":
		if len(args) == 4 { // (target, tool, k, lp)
			return blend{sides: []pair{args[0], negLeaf(args[1])}, k: args[2].first, lp: args[3].first}, true
		}
	}
	return blend{}, false
}

// pow raises a non-negative operand to the combiner's norm exponent; root
// inverts it. The round family is L2, where SQUARE/SQRT say so directly.
func (b blend) pow(x *program.Operand) *program.Operand {
	if b.lp == nil {
		return un(program.Operation_SQUARE, x)
	}
	return bin(program.Operation_POW, x, b.lp)
}

func (b blend) root(x *program.Operand) *program.Operand {
	if b.lp == nil {
		return sqrt(x)
	}
	return bin(program.Operation_POW, x, div(num(1), b.lp))
}

// second reports the distance to the second-nearest piece of the combiner's
// topology. A blend of n surfaces has more pieces than the n it was handed:
// each input surface contributes a face, and the fillets between them are
// pieces in their own right. So a point on a face has a fillet as its nearest
// neighbour, and a point on a fillet has a face — the edges of the result are
// where a face meets a fillet, not where the two inputs used to cross.
//
// Which piece the point is ON is named exactly by the combiner's own
// arithmetic. Every one of these combiners is min(-k, sharp) + (Σ cᵢ^lp)^(1/lp)
// with cᵢ = max(dᵢ + k, 0), so the clamped terms pick it out: S = {i : cᵢ > 0}
// is the blend the point currently sits in (a lone face when |S| = 1, the sharp
// intersect when S is empty).
//
// The neighbouring pieces are S with one index taken out or put back, and both
// are the same edit: term j swaps its clamped contribution cⱼ for the part that
// was clamped away, uⱼ = min(dⱼ + k, 0). For j in S that drops face j, naming
// the lower-order blend the point would cross into; for j outside S it adds
// face j, naming the blend they would form — which is why the unclamped |uⱼ|
// belongs there, growing as face j recedes. One expression covers both, so this
// costs O(n) and not the O(n²) of writing out every pair.
//
// Extending each piece to a global field is what makes this tractable, and is
// also where it approximates: a fillet extends to the whole tube of radius k
// about its spine, so distance to it is read on that tube rather than on the
// bounded patch. At the tangency curves — the edges this exists to find — the
// two agree. Note those curves are where a face meets a fillet TANGENTIALLY, so
// the gap between pieces is second order and a threshold ε selects a band of
// width ~sqrt(2kε) rather than ~ε. That is the geometry of a smooth edge, not
// slack here; `.second.dp` reads near-parallel across it, which is what tells a
// blend edge apart from a sharp one.
func (b blend) second() *program.Operand {
	n := len(b.sides)
	// Written in the intersect frame throughout; a union is the De Morgan dual,
	// so negate going in and coming back out.
	clamped := make([]*program.Operand, n)
	cut := make([]*program.Operand, n)
	for i, s := range b.sides {
		d := s.first
		if b.union {
			d = neg(d)
		}
		off := add(d, b.k)
		clamped[i] = b.pow(bin(program.Operation_MAX, off, num(0)))
		cut[i] = b.pow(neg(bin(program.Operation_MIN, off, num(0))))
	}
	sum := add(clamped...)

	cands := make([]*program.Operand, 0, 2*n)
	for j := range b.sides {
		raw := add(sub(sum, clamped[j]), cut[j])
		// raw == 0 means the edit empties the blend, which is not a piece: face
		// j is the only active one, so its neighbours are named by the other
		// indices instead. Unrejected, a flat face reports a piece a constant k
		// away and drowns out the real one.
		//
		// Push it out to the sentinel BEFORE the root rather than discarding the
		// answer after it, which does two jobs with one term. The root's
		// derivative carries 1/root, and raw is identically zero across a whole
		// region, so a root taken over raw would make `.second.dp` 0/0 there —
		// this way the root never sees zero and needs no floor of its own.
		//
		// Added, never selected. The mesher lowers `cond ? a : b` to
		// `b + cond*(a-b)`, so a sentinel branch destroys the other one: with
		// b = secondInf, `v - 1e30` rounds to exactly -1e30 in f32 and a taken
		// branch evaluates to 1e30 - 1e30 = 0. Multiplying by a comparison that
		// is exactly 0 when the piece is real leaves raw untouched.
		swapped := add(raw, mul(sub(num(1), bin(program.Operation_GT, raw, num(0))), inf()))
		v := sub(b.root(swapped), b.k)
		if b.union {
			v = neg(v)
		}
		cands = append(cands, v)
	}
	// A side that is itself a combiner keeps its own runner-up, so the edges of
	// a shape fed into a blend still register.
	for _, s := range b.sides {
		if !isInf(s.second) {
			cands = append(cands, s.second)
		}
	}
	return nearestToSurface(cands)
}

// nearestToSurface picks the candidate closest to zero and keeps its sign: the
// pieces are ranked by how far the point is from each, and a piece behind the
// point is as real as one in front.
//
// How often each candidate is named matters more than it looks. Nothing shares
// subexpressions downstream, so a candidate is re-emitted once per textual
// occurrence, and each of these carries the combiner's whole clamped sum. A
// pairwise fold names its running winner in both the comparison and a branch,
// doubling at every step; even splitting into a smallest-non-negative and a
// largest-negative fold names each candidate eight times once the two folds are
// compared. Taking the magnitude separately from the sign costs four: the
// magnitude fold names each candidate once and is itself named twice, and the
// non-negative fold names each candidate twice and is itself named once.
//
// A tie in magnitude between a positive and a negative candidate resolves
// positive. The two differ only in sign, so anything reading a distance from
// this is unaffected.
func nearestToSurface(cands []*program.Operand) *program.Operand {
	if len(cands) == 1 {
		return cands[0]
	}
	mags := make([]*program.Operand, len(cands))
	ahead := make([]*program.Operand, len(cands))
	for i, v := range cands {
		m := un(program.Operation_ABS, v)
		mags[i] = m
		// Only a piece in front competes here, so push the ones behind out by
		// the same additive penalty rather than selecting a sentinel over them.
		ahead[i] = add(m, mul(bin(program.Operation_LT, v, num(0)), inf()))
	}
	nearest := nary(program.Operation_MIN, mags...)
	// The nearest piece is behind the point exactly when no piece in front of it
	// is as close. Both arms of this select are small constants, so lowering it
	// as b + cond*(a-b) is exact.
	behind := bin(program.Operation_LT, nearest, nary(program.Operation_MIN, ahead...))
	return mul(nearest, nary(program.Operation_SELECT, behind, num(-1), num(1)))
}

func listElems(p pair) ([]pair, bool) {
	if p.elems == nil {
		return nil, false
	}
	return p.elems, true
}

// negLeaf negates a surface's distance (De Morgan complement, for difference);
// the result is one surface, so it carries no second.
func negLeaf(p pair) pair { return leaf(neg(p.first)) }

func (dd *secondDomain) Builtin(name string, args []Value) (Value, error) {
	ps := pairs(args)
	// worleyF1 is the one builtin that already knows its own second-closest. Its
	// field IS the distance to the nearest of a cloud of feature points, so each
	// feature point is a surface in exactly the sense .second means, and the
	// second-nearest of them is worleyF2 — which the oracle reads off the same
	// 27-cell scan, so asking for both costs one scan rather than two.
	//
	// Without this the candidates are invisible: they are not reachable through
	// min/max at this level (the scan is a builtin, and before it was a builtin
	// the fold hid them behind noiseHash3 and length), so .second reported the
	// no-second sentinel for every cellular field. It reaches the public
	// noise.cellular(size) only because scalePair carries the `* size` that
	// converts cell units to world units.
	if name == program.WorleyF1 {
		firsts := firstsOf(ps)
		return pair{
			first:  call(program.WorleyF1, firsts...),
			second: call(program.WorleyF2, firsts...),
		}, nil
	}
	// Other builtins (length, dot, clamp, …) appear inside a single primitive's
	// distance computation: the result is one surface, so it has no second-closest.
	return leaf(call(name, firstsOf(ps)...)), nil
}

func (dd *secondDomain) Op(operation program.Operation, args []Value) (Value, error) {
	ps := pairs(args)
	switch operation {
	case program.Operation_MIN:
		// Union: the two smallest surface distances.
		return orderStat(ps, false), nil
	case program.Operation_MAX:
		// Intersection: the two largest — the binding constraint and the next one.
		// A 2-way max (e.g. cylinderCapped) then gives the non-binding side rather
		// than the SDF itself; a 3-way max (boxCentered) still gives the middle
		// face (for three operands the 2nd-largest and 2nd-smallest coincide).
		return orderStat(ps, true), nil
	case program.Operation_ADD:
		return foldPair(ps, addPair), nil
	case program.Operation_SUB:
		return foldPair(ps, subPair), nil
	case program.Operation_MUL:
		if scaled, ok := scalePair(ps); ok {
			return scaled, nil
		}
		return leaf(nary(program.Operation_MUL, firstsOf(ps)...)), nil
	case program.Operation_DIV:
		// Dividing by constants is the same similarity as multiplying by their
		// reciprocal, but only the FIRST operand may carry the field: a / b / c
		// scales a, while b / a does not scale anything (a is in a denominator).
		if scaled, ok := divPair(ps); ok {
			return scaled, nil
		}
		return leaf(nary(program.Operation_DIV, firstsOf(ps)...)), nil
	case program.Operation_NEG:
		// A negated surface is still one surface (used by difference: max(a, -b)).
		return leaf(neg(ps[0].first)), nil
	default:
		// Any other arithmetic (sqrt, trig, modulo, comparisons, …) is internal to
		// a primitive's distance or is not a surface combiner: keep first, drop the
		// second-closest.
		return leaf(nary(operation, firstsOf(ps)...)), nil
	}
}

// --- pair folds & helpers ---------------------------------------------------

func pairs(args []Value) []pair {
	out := make([]pair, len(args))
	for i, a := range args {
		out[i] = a.(pair)
	}
	return out
}

func firstsOf(ps []pair) []*program.Operand {
	out := make([]*program.Operand, len(ps))
	for i, p := range ps {
		out[i] = p.first
	}
	return out
}

func secondsOf(ps []pair) []*program.Operand {
	out := make([]*program.Operand, len(ps))
	for i, p := range ps {
		out[i] = p.second
	}
	return out
}

func foldPair(ps []pair, combine func(a, b pair) pair) Value {
	acc := ps[0]
	for i := 1; i < len(ps); i++ {
		acc = combine(acc, ps[i])
	}
	return acc
}

// orderStat returns the closest/second-closest over the operands of a sharp
// combiner: the two smallest for a union (MIN), the two largest for an
// intersection (MAX). Each operand contributes its surface distance (first) and,
// when it is itself a nested combiner, its runner-up (a finite second); the +∞
// sentinel of a plain surface carries no second. A sharp combiner always has ≥ 2
// operands.
//
// A runner-up belongs to the surface it came from, so only the WINNING operand's
// competes for the second slot. Merging every operand's runner-up into one flat
// candidate list instead lets a losing one displace the winner: an intersection's
// runner-up sits behind its own surface (negative inside the solid), so a box in
// a union would win the min-fold with its second-nearest face and report its own
// surface as the second-closest — `.second` would equal the SDF over the whole
// box, not just along its edges.
func orderStat(ps []pair, largest bool) pair {
	acc := ps[0]
	for _, b := range ps[1:] {
		acc = mergeStat(acc, b, largest)
	}
	return acc
}

// mergeStat folds one operand into the running order statistic. The second is
// the nearer of the loser's surface and the winner's own runner-up, picked by
// which of the two surfaces wins — the same ternary the derivative uses to carry
// a min/max branch's tangent. When neither side has a runner-up it collapses to
// the two-operand statistic, so leaf-only combiners stay branch-free.
func mergeStat(a, b pair, largest bool) pair {
	pick, anti, wins := program.Operation_MIN, program.Operation_MAX, program.Operation_LT
	if largest {
		pick, anti, wins = program.Operation_MAX, program.Operation_MIN, program.Operation_GT
	}
	first := bin(pick, a.first, b.first)
	if isInf(a.second) && isInf(b.second) {
		return pair{first: first, second: bin(anti, a.first, b.first)}
	}
	// a wins: its runner-up competes with b's surface, and b's runner-up (behind
	// b's own surface) is dropped. Symmetrically when b wins.
	aWins, bWins := b.first, a.first
	if !isInf(a.second) {
		aWins = bin(pick, aWins, a.second)
	}
	if !isInf(b.second) {
		bWins = bin(pick, bWins, b.second)
	}
	cond := bin(wins, a.first, b.first)
	return pair{first: first, second: nary(program.Operation_SELECT, cond, aWins, bWins)}
}

// addPair is a + b. When one side is a scalar/leaf (second == ∞) and the other a
// combiner, the offset shifts the combiner's second-closest too; two combiners
// (or two leaves) share no second, so it stays ∞.
// A positive scale factor must be strictly greater than this to be carried. At
// or below it the sign is wrong or the scale is degenerate: a negative factor
// turns a field inside out, which swaps the roles of MIN and MAX and so the
// meaning of "second closest" (NEG bails to a leaf for the same reason), and a
// zero factor collapses every surface onto one value.
const smallestCarriedScale = 0.0

// constScale reports op's value when it is a numeric literal safe to scale a
// {first, second} pair by.
//
// It must be a literal, decided at build time, not a runtime sign test. The
// secondInf sentinel forbids the latter: backends lower `cond ? a : b`
// algebraically as `b + cond*(a-b)`, and in f32 a real value minus 1e30 rounds
// to exactly -1e30, so a taken branch would evaluate to zero (see secondInf).
// A `scale > 0 ? second*scale : inf` select is precisely that trap. The cost of
// the restriction is that a field scaled by a slider keeps no second-closest.
func constScale(op *program.Operand) (float64, bool) {
	v, ok := op.T.(*program.Operand_Value)
	if !ok || v.Value <= smallestCarriedScale {
		return 0, false
	}
	return v.Value, true
}

// scalePair carries a {first, second} pair through multiplication by positive
// constants: a uniform similarity, under which two surfaces at d1 and d2 move to
// scale*d1 and scale*d2, so both components scale by the same factor. This is
// exactly as sound as the additive offset addPair already carries, and it is what
// lets a library work in normalized cell units and scale its result back to world
// units — `lattice(p / size) * size` — without losing .second on the way out.
//
// It applies only when every operand but one is such a constant and that one
// operand carries a finite second. Anything else keeps the conservative leaf:
// multiplying two fields together gives a product whose magnitude is not a
// distance to anything, and a position-dependent factor is not an isometry at
// all, so neither has a second-closest surface to speak of.
func scalePair(ps []pair) (pair, bool) {
	fieldAt := -1
	scale := 1.0
	for i, p := range ps {
		if v, ok := constScale(p.first); ok {
			scale *= v
			continue
		}
		if fieldAt >= 0 {
			return pair{}, false // more than one non-constant operand
		}
		fieldAt = i
	}
	if fieldAt < 0 || isInf(ps[fieldAt].second) {
		return pair{}, false
	}
	return pair{
		first:  nary(program.Operation_MUL, firstsOf(ps)...),
		second: mul(ps[fieldAt].second, num(scale)),
	}, true
}

// divPair is scalePair for division: the field must be the numerator, and every
// divisor a positive constant, so the whole expression is still a similarity.
func divPair(ps []pair) (pair, bool) {
	if len(ps) < 2 || isInf(ps[0].second) {
		return pair{}, false
	}
	scale := 1.0
	for _, p := range ps[1:] {
		v, ok := constScale(p.first)
		if !ok {
			return pair{}, false
		}
		scale /= v
	}
	return pair{
		first:  nary(program.Operation_DIV, firstsOf(ps)...),
		second: mul(ps[0].second, num(scale)),
	}, true
}

func addPair(a, b pair) pair {
	first := add(a.first, b.first)
	switch {
	case !isInf(a.second) && isInf(b.second):
		return pair{first: first, second: add(a.second, b.first)}
	case isInf(a.second) && !isInf(b.second):
		return pair{first: first, second: add(a.first, b.second)}
	default:
		return pair{first: first, second: inf()}
	}
}

// subPair is a - b. The second-closest survives only when the minuend a is the
// combiner (union - scalar, the erode/shell case); b's own second is dropped.
func subPair(a, b pair) pair {
	first := sub(a.first, b.first)
	if !isInf(a.second) {
		return pair{first: first, second: sub(a.second, b.first)}
	}
	return pair{first: first, second: inf()}
}
