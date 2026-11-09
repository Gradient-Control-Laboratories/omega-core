package property

import (
	"math"
	"testing"

	"gradientcontrol.com/omega/core/program"
)

// shape declares a plain (uncolored) sphere SDF named `name`.
func shape(name string, cx, cy, cz, r float64) *program.Assignment {
	return coloredShape(name, sphereSDF(cx, cy, cz, r), nil)
}

// sc wraps target in the SECONDCLOSEST operator (the .second accessor).
func sc(target *program.Operand) *program.Operand {
	return nary(program.Operation_SECONDCLOSEST, target)
}

// secondAt expands the package's single return (a .second field) and evaluates
// it at pt, substituting each shape's SDF body for its live reference — the same
// numeric-evaluation shape colorAt uses.
func secondAt(t *testing.T, pkg *program.Package, pt point) float64 {
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
	return evalScalar(t, op, pt)
}

// threeSpheresPkg: a@x=-2, b@x=0, c@x=3, radius 1, combined by ret. At the origin
// the distances are da=1, db=-1, dc=2, so the second-closest surface is a (1).
func threeSpheresPkg(ret *program.Operand) *program.Package {
	return &program.Package{Files: []*program.PackageFile{{
		Assignments: []*program.Assignment{
			shape("a", -2, 0, 0, 1),
			shape("b", 0, 0, 0, 1),
			shape("c", 3, 0, 0, 1),
		},
		Returns: []*program.Operand{ret},
	}}}
}

func expectSecond(t *testing.T, got, want float64) {
	t.Helper()
	if !closeEnough(got, want) {
		t.Errorf("second = %v, want %v", got, want)
	}
}

// TestSecondSharpUnion: min(a,b,c).second is the distance to the second-nearest
// surface — at the origin da=1 (a) is second after db=-1 (b).
func TestSecondSharpUnion(t *testing.T) {
	pkg := threeSpheresPkg(sc(nary(program.Operation_MIN, ref("a"), ref("b"), ref("c"))))
	expectSecond(t, secondAt(t, pkg, point{}), 1)
	// Shift the query to x=2.5 (between b and c): db=1.5, dc=-0.5, da=3.5 -> second db=1.5.
	expectSecond(t, secondAt(t, pkg, point{x: 2.5}), 1.5)
}

// TestSecondMaxIsTwoLargest: MIN keeps the two smallest, MAX the two largest —
// a 2-way max's .second is the non-binding constraint, not the SDF itself.
func TestSecondMaxIsTwoLargest(t *testing.T) {
	// Two spheres, origin: da=1, db=-1. MIN keeps the two smallest so its .second
	// is the larger (max = 1); MAX keeps the two largest so its .second is the
	// smaller (min = -1).
	twoMin := threeSpheresPkg(sc(nary(program.Operation_MIN, ref("a"), ref("b"))))
	twoMax := threeSpheresPkg(sc(nary(program.Operation_MAX, ref("a"), ref("b"))))
	expectSecond(t, secondAt(t, twoMin, point{}), 1)
	expectSecond(t, secondAt(t, twoMax, point{}), -1)

	// Three operands: the 2nd-smallest and 2nd-largest are both the middle value,
	// so min and max agree (da=1, db=-1, dc=2 -> middle = 1).
	threeMin := threeSpheresPkg(sc(nary(program.Operation_MIN, ref("a"), ref("b"), ref("c"))))
	threeMax := threeSpheresPkg(sc(nary(program.Operation_MAX, ref("a"), ref("b"), ref("c"))))
	expectSecond(t, secondAt(t, threeMax, point{}), secondAt(t, threeMin, point{}))
	expectSecond(t, secondAt(t, threeMax, point{}), 1)
}

// TestSecondNestedFlattens: min(min(a,b),c).second flattens to the 2nd-smallest
// of {a,b,c}, not treating the inner union as a single surface.
func TestSecondNestedFlattens(t *testing.T) {
	inner := nary(program.Operation_MIN, ref("a"), ref("b"))
	pkg := threeSpheresPkg(sc(nary(program.Operation_MIN, inner, ref("c"))))
	expectSecond(t, secondAt(t, pkg, point{}), 1)
}

// TestSecondSingleShape: a lone surface has no second-closest — the sentinel.
func TestSecondSingleShape(t *testing.T) {
	pkg := threeSpheresPkg(sc(ref("a")))
	if v := secondAt(t, pkg, point{}); v < 1e20 {
		t.Errorf("single-shape second = %v, want the ~1e30 sentinel", v)
	}
}

// TestSecondOffsetUnion: an affine offset above a union shifts the second-closest
// too — (min+k) adds k, (min-k) subtracts k.
func TestSecondOffsetUnion(t *testing.T) {
	union := nary(program.Operation_MIN, ref("a"), ref("b"), ref("c"))
	grow := threeSpheresPkg(sc(add(union, num(0.5))))
	erode := threeSpheresPkg(sc(sub(union, num(0.25))))
	expectSecond(t, secondAt(t, grow, point{}), 1.5)   // 1 + 0.5
	expectSecond(t, secondAt(t, erode, point{}), 0.75) // 1 - 0.25
}

// TestSecondOffsetSingleShapeStaysSentinel: offsetting a lone surface creates no
// second-closest; it stays the sentinel.
func TestSecondOffsetSingleShapeStaysSentinel(t *testing.T) {
	pkg := threeSpheresPkg(sc(add(ref("a"), num(0.5))))
	if v := secondAt(t, pkg, point{}); v < 1e20 {
		t.Errorf("offset single-shape second = %v, want the sentinel", v)
	}
}

// TestSecondDestructured: shapes bound by a destructuring assignment open up —
// an intersection bound through `s, r = f()` contributes its own runner-up
// (winner's-runner-up fold), instead of reading as a single opaque surface.
func TestSecondDestructured(t *testing.T) {
	// f() { return <sphere@origin r=1>, <slab intersection max(x-2, 1-x)> }
	// The slab's solid is x in [1,2] (faces at x=1 and x=2), infinite in y/z.
	sphere := sphereSDF(0, 0, 0, 1)
	slab := bin(program.Operation_MAX,
		sub(ref(program.XVarName), num(2)),
		sub(num(1), ref(program.XVarName)))
	f := &program.Function{Returns: []*program.Operand{sphere, slab}}
	pkg := &program.Package{Files: []*program.PackageFile{{
		FnTypes: []*program.FnType{{Name: "f", Groups: []*program.FnGroup{{
			Functions: []*program.Function{f},
		}}}},
		Assignments: []*program.Assignment{multiAssign(fnCallOp("f"), "s", "r")},
		Returns:     []*program.Operand{sc(nary(program.Operation_MIN, ref("s"), ref("r")))},
	}}}

	prog, err := Expand(&program.Program{Primary: pkg})
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	out := prog.Primary
	// The expansion keeps live references to the destructured names; substitute
	// their fields (the function's returns) for numeric evaluation.
	op := flattenRemap(out.Files[0].Returns[0])
	op = substRef(op, "s", sphere)
	op = substRef(op, "r", slab)

	// Origin: ds=-1, dr=max(-2,1)=1 -> s wins; s is a lone surface (no
	// runner-up), so .second is the loser's surface: 1.
	expectSecond(t, evalScalar(t, op, point{}), 1)
	// x=1.8: ds=0.8, dr=max(-0.2,-0.8)=-0.2 -> the intersection wins; its own
	// runner-up (the non-binding face, -0.8) beats the sphere (0.8): -0.8.
	expectSecond(t, evalScalar(t, op, point{x: 1.8}), -0.8)
}

// TestSecondGradient: field.second.dp differentiates through the expanded
// order statistic — the gradient of the second-closest field is the unit
// vector of whichever surface is currently second.
func TestSecondGradient(t *testing.T) {
	union := nary(program.Operation_MIN, ref("a"), ref("b"), ref("c"))
	pkg := threeSpheresPkg(nary(program.Operation_DERIV, sc(union), ref(program.PVarName)))
	prog, err := Expand(&program.Program{Primary: pkg})
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	out := prog.Primary
	op := flattenRemap(out.Files[0].Returns[0])
	for _, f := range pkg.Files {
		for _, a := range f.Assignments {
			if a != nil && len(a.Variables) == 1 {
				op = substRef(op, a.Variables[0].Name, a.Expression)
			}
		}
	}
	// At (0,1,0): db=0 wins, the second surface is a@(-2,0,0) (da≈1.236 vs
	// dc≈2.162), so the gradient is (p - center a)/|…| = (2,1,0)/sqrt(5).
	g := evalVec(t, op, point{y: 1})
	s5 := math.Sqrt(5)
	want := []float64{2 / s5, 1 / s5, 0}
	for i := range want {
		if !closeEnough(g[i], want[i]) {
			t.Errorf("second.dp[%d] = %v, want %v (full %v)", i, g[i], want[i], g)
		}
	}
}

// TestSecondTemporal: a static field's second-closest does not move — .second
// under .dtime is zero.
func TestSecondTemporal(t *testing.T) {
	union := nary(program.Operation_MIN, ref("a"), ref("b"), ref("c"))
	pkg := threeSpheresPkg(nary(program.Operation_DERIV, sc(union), ref(program.TimeVarName)))
	prog, err := Expand(&program.Program{Primary: pkg})
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	out := prog.Primary
	op := flattenRemap(out.Files[0].Returns[0])
	for _, f := range pkg.Files {
		for _, a := range f.Assignments {
			if a != nil && len(a.Variables) == 1 {
				op = substRef(op, a.Variables[0].Name, a.Expression)
			}
		}
	}
	expectSecond(t, evalScalar(t, op, point{y: 1}), 0)
}
