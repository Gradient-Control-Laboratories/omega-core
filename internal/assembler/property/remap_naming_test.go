package property

import (
	"math"
	"testing"

	"gradientcontrol.com/omega/core/program"
)

// A field that places a named shape with an inline remap — min(near,
// remap(ball, q)) — reads the ball where the remap puts it, in every property
// the engine expands. Evaluated in the remap's frame, the named ball used to
// come back as its live reference `ball`, which the backend evaluates at the
// OUTER p, so .second and .dp read the ball unmoved, unlike the same remap
// written as a declaration (placed = remap(ball, q)), which was right. A
// slider read there is the same under any remap and stays live.

// shiftX is the coordinate that puts a shape at (dx, 0, 0): p - [dx, 0, 0].
func shiftX(dx float64) *program.Operand {
	return sub(ref(program.PVarName), rgb(dx, 0, 0))
}

func remapOf(target, coord *program.Operand) *program.Operand {
	return nary(program.Operation_REMAP, target, coord)
}

// nearAndPlacedBall is a unit sphere "near" at the origin and a unit sphere
// "ball" declared at the origin, placed at x = 3 by an inline remap, combined
// by ret.
func nearAndPlacedBall(ret *program.Operand, extra ...*program.Assignment) *program.Package {
	assigns := append([]*program.Assignment{shape("near", 0, 0, 0, 1), shape("ball", 0, 0, 0, 1)}, extra...)
	return &program.Package{Files: []*program.PackageFile{{
		Assignments: assigns,
		Returns:     []*program.Operand{ret},
	}}}
}

func TestSecondReadsAnInlineRemapOfANamedShapeWhereItPlacesIt(t *testing.T) {
	pkg := nearAndPlacedBall(sc(nary(program.Operation_MIN, ref("near"), remapOf(ref("ball"), shiftX(3)))))

	// At x = 0.5 near binds (-0.5) and the placed ball is second, 1.5 away;
	// read unmoved it would be -0.5. At x = -1 near is on its surface (0) and
	// the placed ball is 3 away; unmoved it would be 0.
	expectSecond(t, secondThroughRemaps(t, pkg, point{x: 0.5}), 1.5)
	expectSecond(t, secondThroughRemaps(t, pkg, point{x: -1}), 3)
}

// The same field with the placement declared, which was already right: the
// inline form must now agree with it.
func TestSecondOfAnInlineRemapMatchesTheDeclaredOne(t *testing.T) {
	declared := &program.Assignment{
		Variables:  []*program.Variable{{Name: "placed"}},
		Expression: remapOf(ref("ball"), shiftX(3)),
	}
	inline := nearAndPlacedBall(sc(nary(program.Operation_MIN, ref("near"), remapOf(ref("ball"), shiftX(3)))))
	// placed is declared ahead of ball, so secondAt — which substitutes each
	// declaration's body in order — resolves the ball that placed's body names.
	named := &program.Package{Files: []*program.PackageFile{{
		Assignments: []*program.Assignment{shape("near", 0, 0, 0, 1), declared, shape("ball", 0, 0, 0, 1)},
		Returns:     []*program.Operand{sc(nary(program.Operation_MIN, ref("near"), ref("placed")))},
	}}}
	for _, pt := range []point{{x: 0.5}, {x: -1}, {x: 2.5, y: 0.4}, {x: 1.5, z: -0.7}} {
		if a, b := secondThroughRemaps(t, inline, pt), secondThroughRemaps(t, named, pt); !closeEnough(a, b) {
			t.Errorf("at %v: inline remap gives %v, declared placement %v", pt, a, b)
		}
	}
}

// A slider read inside the remapped shape stays a live reference: it does not
// depend on p, so the remap is not wrapped round it, and moving the slider
// still moves the field.
func TestSecondThroughARemapKeepsASliderLive(t *testing.T) {
	rad := &program.Assignment{
		Variables:  []*program.Variable{{Name: "rad", Properties: []*program.Property{{Name: "min", Expression: num(0.5)}}}},
		Expression: num(1),
	}
	ballOfRad := sub(call("length", ref(program.PVarName)), ref("rad"))
	pkg := nearAndPlacedBall(sc(nary(program.Operation_MIN, ref("near"), remapOf(ballOfRad, shiftX(3)))), rad)

	prog, err := Expand(&program.Program{Primary: pkg})
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	out := prog.Primary.Files[0].Returns[0]
	if remapsName(out, "rad") {
		t.Errorf("the slider rad is read through a remap; it does not depend on p and must stay a live reference")
	}
	if !readsName(out, "rad") {
		t.Errorf("the expansion no longer reads the slider rad at all")
	}
	expectSecond(t, secondThroughRemaps(t, pkg, point{x: 0.5}), 1.5)
}

// remapsName reports whether op holds remap(name, ...) anywhere.
func remapsName(op *program.Operand, name string) bool {
	if op == nil {
		return false
	}
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		if t.Expression.Operation == program.Operation_REMAP && len(t.Expression.Operands) == 2 {
			if r := t.Expression.Operands[0].GetReference(); r != nil && r.Prefix == "" && r.Name == name {
				return true
			}
		}
		for _, sub := range t.Expression.Operands {
			if remapsName(sub, name) {
				return true
			}
		}
	case *program.Operand_FnCall:
		for _, a := range t.FnCall.Arguments {
			if remapsName(a, name) {
				return true
			}
		}
	}
	return false
}

// readsName reports whether op refers to the unqualified name anywhere.
func readsName(op *program.Operand, name string) bool {
	if op == nil {
		return false
	}
	switch t := op.T.(type) {
	case *program.Operand_Reference:
		return t.Reference.Prefix == "" && t.Reference.Name == name
	case *program.Operand_Expression:
		for _, sub := range t.Expression.Operands {
			if readsName(sub, name) {
				return true
			}
		}
	case *program.Operand_FnCall:
		for _, a := range t.FnCall.Arguments {
			if readsName(a, name) {
				return true
			}
		}
	}
	return false
}

// secondThroughRemaps is secondAt for a field that holds remaps of names:
// every declaration is substituted until none is left, and only then are the
// remaps flattened, so a remap of a name is sampled where it points.
func secondThroughRemaps(t *testing.T, pkg *program.Package, pt point) float64 {
	t.Helper()
	prog, err := Expand(&program.Program{Primary: pkg})
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	rets := prog.Primary.Files[len(prog.Primary.Files)-1].Returns
	if len(rets) != 1 {
		t.Fatalf("expected one return, got %d", len(rets))
	}
	op := rets[0]
	for range pkg.Files[0].Assignments {
		for _, a := range pkg.Files[0].Assignments {
			if a != nil && len(a.Variables) == 1 {
				op = substRef(op, a.Variables[0].Name, a.Expression)
			}
		}
	}
	return evalScalar(t, flattenRemap(op), pt)
}

// The gradient through the same inline remap is the declared one's. Read
// unmoved, the ball ties exactly with near — both are unit spheres at the
// origin — and the tie went to the ball, so where near binds, as at
// (-0.6, 0.2, 0.5), the gradient came out as the placed ball's instead.
func TestGradientThroughAnInlineRemapMatchesTheDeclaredOne(t *testing.T) {
	declared := &program.Assignment{
		Variables:  []*program.Variable{{Name: "placed"}},
		Expression: remapOf(ref("ball"), shiftX(3)),
	}
	inline := nearAndPlacedBall(dpOf(nary(program.Operation_MIN, ref("near"), remapOf(ref("ball"), shiftX(3)))))
	named := &program.Package{Files: []*program.PackageFile{{
		Assignments: []*program.Assignment{shape("near", 0, 0, 0, 1), declared, shape("ball", 0, 0, 0, 1)},
		Returns:     []*program.Operand{dpOf(nary(program.Operation_MIN, ref("near"), ref("placed")))},
	}}}

	at := point{x: -0.6, y: 0.2, z: 0.5}
	d := math.Sqrt(0.6*0.6 + 0.2*0.2 + 0.5*0.5)
	want := []float64{-0.6 / d, 0.2 / d, 0.5 / d}
	if got := evalVec(t, throughRemaps(t, inline), at); !closeVec(got, want) {
		t.Errorf("gradient at %v through the inline remap = %v, want near's %v", at, got, want)
	}
	for _, pt := range []point{at, {x: 0.5}, {x: 2.5, y: 0.4}, {x: 3.2, y: 0.3, z: -0.2}} {
		if a, b := evalVec(t, throughRemaps(t, inline), pt), evalVec(t, throughRemaps(t, named), pt); !closeVec(a, b) {
			t.Errorf("at %v: inline remap gradient %v, declared placement %v", pt, a, b)
		}
	}
}

// throughRemaps expands pkg's single return and substitutes every declaration
// until none is left, then flattens the remaps, ready to evaluate.
func throughRemaps(t *testing.T, pkg *program.Package) *program.Operand {
	t.Helper()
	prog, err := Expand(&program.Program{Primary: pkg})
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	op := prog.Primary.Files[len(prog.Primary.Files)-1].Returns[0]
	for range pkg.Files[0].Assignments {
		for _, a := range pkg.Files[0].Assignments {
			if a != nil && len(a.Variables) == 1 {
				op = substRef(op, a.Variables[0].Name, a.Expression)
			}
		}
	}
	return flattenRemap(op)
}

func closeVec(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !closeEnough(a[i], b[i]) {
			return false
		}
	}
	return true
}
