package property

import (
	"math"
	"testing"

	"gradientcontrol.com/omega/core/program"
)

// A round blend of two surfaces has three pieces of topology, not two: each
// input surface contributes a face, and the fillet between them is a piece in
// its own right. The result's edges are therefore where a face meets the
// fillet — tangentially — and not where the two inputs used to cross.
//
// Half-spaces make those pieces exact, so every expectation below is read off
// the fillet circle rather than off the implementation. For unionRound(x, y, k)
// the fillet is the arc of radius k centred at (k, k), tangent to the plane
// x = 0 at (0, k) and to y = 0 at (k, 0).

// halfSpacePkg declares a = x and b = y and returns ret, which is expected to
// be the .second of a combiner over them.
func halfSpacePkg(ret *program.Operand) *program.Package {
	return &program.Package{Files: []*program.PackageFile{{
		Assignments: []*program.Assignment{
			coloredShape("a", ref(program.XVarName), nil),
			coloredShape("b", ref(program.YVarName), nil),
		},
		Returns: []*program.Operand{ret},
	}}}
}

func TestSecondRoundUnionPieces(t *testing.T) {
	const k = 0.25
	pkg := halfSpacePkg(sc(call("unionRound", ref("a"), ref("b"), num(k))))

	// Where the fillet meets side a, two pieces touch: the distance to the
	// second of them is zero. This is the edge the property exists to report.
	expectSecond(t, secondAt(t, pkg, point{x: 0, y: k}), 0)

	// Halfway round the arc the point is on the fillet, and the nearest other
	// piece is a side, at that side's own distance.
	mid := k - k/math.Sqrt2
	expectSecond(t, secondAt(t, pkg, point{x: mid, y: mid}), mid)

	// On side a past the tangency the nearest other piece is the fillet: the
	// signed distance to the circle of radius k about (k, k). It is negative
	// because a union fillet carries its material outside its own tube.
	for _, y := range []float64{2 * k, 3} {
		expectSecond(t, secondAt(t, pkg, point{x: 0, y: y}), k-math.Hypot(k, k-y))
	}
}

// The intersection is the De Morgan dual: intersectRound(x, y, k) fillets the
// convex edge with the arc of radius k about (-k, -k).
func TestSecondRoundIntersectPieces(t *testing.T) {
	const k = 0.25
	pkg := halfSpacePkg(sc(call("intersectRound", ref("a"), ref("b"), num(k))))

	expectSecond(t, secondAt(t, pkg, point{x: 0, y: -k}), 0)

	mid := -k + k/math.Sqrt2
	expectSecond(t, secondAt(t, pkg, point{x: mid, y: mid}), mid)

	expectSecond(t, secondAt(t, pkg, point{x: 0, y: -2 * k}), math.Hypot(k, -2*k+k)-k)
}

// differenceRound complements the tool, so the pieces are the target's face,
// the tool's complemented face, and the fillet between them — here the arc
// about (-k, k), tangent to the target face at (0, k).
func TestSecondRoundDifferencePieces(t *testing.T) {
	const k = 0.25
	pkg := halfSpacePkg(sc(call("differenceRound", ref("a"), ref("b"), num(k))))

	expectSecond(t, secondAt(t, pkg, point{x: 0, y: k}), 0)
	expectSecond(t, secondAt(t, pkg, point{x: 0, y: 2 * k}), math.Hypot(k, k-2*k)-k)
}

// squareSides is a square of four half-spaces, |x| <= 1 and |y| <= 1, as a list
// literal — the argument shape the list-form combiners take.
func squareSides() *program.Operand {
	return nary(program.Operation_COLLECT,
		sub(ref(program.XVarName), num(1)), sub(neg(ref(program.XVarName)), num(1)),
		sub(ref(program.YVarName), num(1)), sub(neg(ref(program.YVarName)), num(1)))
}

func listPkg(ret *program.Operand) *program.Package {
	return &program.Package{Files: []*program.PackageFile{{Returns: []*program.Operand{ret}}}}
}

// The list form rounds every edge with one blend term covering all of them at
// once, so its pieces have to be recovered rather than read off directly.
func TestSecondRoundListPieces(t *testing.T) {
	const k = 0.15
	pkg := listPkg(sc(call("intersectRound", squareSides(), num(k))))

	// Where the +x face meets its corner fillet, two pieces touch.
	expectSecond(t, secondAt(t, pkg, point{x: 1, y: 1 - k}), 0)

	// Mid-face, well away from either corner. The nearest other piece is the
	// fillet this face would form with a face 1 away, read on the tube about
	// their common spine. It must NOT come back as a constant k: that is the
	// empty-blend degeneracy the guard exists to reject, and it would drown out
	// every real neighbour on a flat face.
	got := secondAt(t, pkg, point{x: 1, y: 0})
	expectSecond(t, got, math.Hypot(k, math.Abs(-1+k))-k)
	if math.Abs(math.Abs(got)-k) < 1e-6 {
		t.Errorf("second = %v on a flat face: the empty blend leaked through the guard", got)
	}

	// Halfway round a corner arc the nearest other piece is a face.
	arc := 1 - k + k/math.Sqrt2
	expectSecond(t, secondAt(t, pkg, point{x: arc, y: arc}), arc-1)
}

// The LP family blends with an Lp norm instead of L2; the pieces are recovered
// the same way, with the exponent carried through.
func TestSecondLPListPieces(t *testing.T) {
	const k, lp = 0.15, 4.0
	pkg := listPkg(sc(call("intersectLp", squareSides(), num(k), num(lp))))

	expectSecond(t, secondAt(t, pkg, point{x: 1, y: 1 - k}), 0)

	want := math.Pow(math.Pow(k, lp)+math.Pow(1-k, lp), 1/lp) - k
	expectSecond(t, secondAt(t, pkg, point{x: 1, y: 0}), want)
}

// A side that is itself a combiner keeps its own runner-up, so feeding a sharp
// box into a blend does not erase the box's own edges.
func TestSecondRoundKeepsNestedEdges(t *testing.T) {
	const k = 0.2
	box := nary(program.Operation_MAX,
		sub(ref(program.XVarName), num(1)), sub(neg(ref(program.XVarName)), num(1)),
		sub(ref(program.YVarName), num(1)), sub(neg(ref(program.YVarName)), num(1)))
	pkg := &program.Package{Files: []*program.PackageFile{{
		Assignments: []*program.Assignment{
			coloredShape("box", box, nil),
			coloredShape("far", add(ref(program.XVarName), num(4)), nil),
		},
		Returns: []*program.Operand{sc(call("unionRound", ref("box"), ref("far"), num(k)))},
	}}}

	// The box's corner is still an edge of the union.
	expectSecond(t, secondAt(t, pkg, point{x: 1, y: 1}), 0)
	// Mid-face it is the box's own runner-up — the opposite face, 1 away — that
	// is nearest, not the distant blend with `far`.
	expectSecond(t, secondAt(t, pkg, point{x: 1, y: 0}), -1)
}

// expandTree expands pkg's single return and substitutes each declaration's
// body for its live reference, the way secondAt does, but hands back the
// operand instead of a number.
func expandTree(t *testing.T, pkg *program.Package) *program.Operand {
	t.Helper()
	prog, err := Expand(&program.Program{Primary: pkg})
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	out := prog.Primary
	rets := out.Files[len(out.Files)-1].Returns
	if len(rets) != 1 {
		t.Fatalf("expected one return, got %d", len(rets))
	}
	op := flattenRemap(rets[0])
	for _, f := range pkg.Files {
		for _, a := range f.Assignments {
			if a == nil || len(a.Variables) != 1 {
				continue
			}
			op = substRef(op, a.Variables[0].Name, a.Expression)
		}
	}
	return op
}

func subtrees(op *program.Operand) []*program.Operand {
	if op == nil {
		return nil
	}
	out := []*program.Operand{op}
	switch v := op.T.(type) {
	case *program.Operand_Expression:
		for _, o := range v.Expression.Operands {
			out = append(out, subtrees(o)...)
		}
	case *program.Operand_FnCall:
		for _, o := range v.FnCall.Arguments {
			out = append(out, subtrees(o)...)
		}
	}
	return out
}

// The backends evaluate both arms of a select — the mesher does so on intervals
// — so every subexpression has to be finite on its own, not just the arm that
// happens to be taken. A blend's guarded-away branch is where that bites: its
// root argument is identically zero across a whole region, and d(root(u))
// carries 1/root(u), so `.second.dp` would be 0/0 there and one NaN arm poisons
// the result whichever way the guard falls.
func TestSecondBlendDerivativeIsFiniteInBothArms(t *testing.T) {
	const k = 0.15
	pkg := listPkg(nary(program.Operation_DERIV,
		sc(call("intersectRound", squareSides(), num(k))), ref(program.PVarName)))
	tree := expandTree(t, pkg)

	// Mid-face, where a single side is active: exactly the degenerate region the
	// piece guard rejects, and so exactly where the rejected arm is evaluated.
	for _, pt := range []point{{x: 1}, {x: 1, y: 0.5}, {y: 1}, {x: 1, y: 1 - k}} {
		for _, sub := range subtrees(tree) {
			v, err := evalNum(sub, pt.vars())
			if err != nil {
				continue // not every node is independently evaluable
			}
			for _, f := range v {
				if math.IsNaN(f) || math.IsInf(f, 0) {
					t.Fatalf("at %+v a subexpression evaluates to %v:\n%s", pt, f, sub.PrettyPrint(0))
				}
			}
		}
	}
}
