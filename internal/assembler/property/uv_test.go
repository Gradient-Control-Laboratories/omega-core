package property

import (
	"math"
	"testing"

	"gradientcontrol.com/omega/core/program"
)

// A shape's frame coordinates ride the same resolver as its colour, so these
// check the behaviour that matters for a frame rather than re-testing the
// blend: that u and v resolve independently, that a sharp boolean SELECTS a
// frame instead of averaging two that do not belong together, and that a shape
// carrying no frame reads as zero rather than as some invented direction.

// framedShape is a shape carrying u, v, or both.
func framedShape(name string, sdf, u, v *program.Operand) *program.Assignment {
	variable := &program.Variable{Name: name}
	if u != nil {
		variable.Properties = append(variable.Properties, &program.Property{Name: uProperty, Expression: u})
	}
	if v != nil {
		variable.Properties = append(variable.Properties, &program.Property{Name: vProperty, Expression: v})
	}
	return &program.Assignment{Variables: []*program.Variable{variable}, Expression: sdf}
}

// twoFramedPkg is two unit spheres, "near" at the origin and "far" at (3,0,0),
// each with its own frame, combined by the given return.
func twoFramedPkg(ret *program.Operand, nearU, nearV, farU, farV *program.Operand) *program.Package {
	return &program.Package{
		Files: []*program.PackageFile{{
			Assignments: []*program.Assignment{
				framedShape("near", sphereSDF(0, 0, 0, 1), nearU, nearV),
				framedShape("far", sphereSDF(3, 0, 0, 1), farU, farV),
			},
			Returns: []*program.Operand{ret},
		}},
	}
}

// frameAt evaluates the one resolved field of the given property at pt, the way
// colorAt does for colour.
func frameAt(t *testing.T, pkg *program.Package, which string, pt point) float64 {
	t.Helper()
	var fields []*program.Operand
	var err error
	if which == uProperty {
		fields, err = UFields(pkg, nil)
	} else {
		fields, err = VFields(pkg, nil)
	}
	if err != nil {
		t.Fatalf("%sFields: %v", which, err)
	}
	if len(fields) != 1 || fields[0] == nil {
		t.Fatalf("expected one %s field, got %d", which, len(fields))
	}
	op := flattenRemap(fields[0])
	for _, f := range pkg.Files {
		for _, a := range f.Assignments {
			if a == nil || len(a.Variables) != 1 {
				continue
			}
			op = substRef(op, a.Variables[0].Name, a.Expression)
		}
	}
	got := evalVec(t, op, pt)
	if len(got) != 1 {
		t.Fatalf("%s resolved to %d components, want a scalar", which, len(got))
	}
	return got[0]
}

// A sharp union hands the nearer shape's frame over whole. This is the property
// a frame needs most: at a box's edge the two faces' coordinates mean different
// things, and an average of them belongs to neither face.
func TestFrameSharpUnionSelectsRatherThanBlends(t *testing.T) {
	pkg := twoFramedPkg(bin(program.Operation_MIN, ref("near"), ref("far")), num(10), num(20), num(30), num(40))

	if got := frameAt(t, pkg, uProperty, point{x: 1}); math.Abs(got-10) > 1e-6 {
		t.Errorf("u near the first shape = %v, want its own 10", got)
	}
	if got := frameAt(t, pkg, uProperty, point{x: 2}); math.Abs(got-30) > 1e-6 {
		t.Errorf("u near the second shape = %v, want its own 30", got)
	}
}

// u and v resolve separately, so a shape may carry either, both or neither —
// and v must not pick up u's value, which a shared carrier would let happen.
func TestFrameCoordinatesResolveIndependently(t *testing.T) {
	pkg := twoFramedPkg(bin(program.Operation_MIN, ref("near"), ref("far")), num(10), num(20), num(30), num(40))

	if got := frameAt(t, pkg, vProperty, point{x: 1}); math.Abs(got-20) > 1e-6 {
		t.Errorf("v near the first shape = %v, want its own 20, not u's", got)
	}
	if got := frameAt(t, pkg, vProperty, point{x: 2}); math.Abs(got-40) > 1e-6 {
		t.Errorf("v near the second shape = %v, want its own 40", got)
	}
}

// A shape carrying no frame INHERITS its group's, exactly as an uncoloured
// shape inherits a colour (see ColorFields). This is deliberate — frames ride
// the colour resolver unchanged rather than earning a rule of their own — and
// it is the one place a frame can still be wrong: a sphere unioned onto a
// cylinder reads the cylinder's angle field, which means nothing on a sphere.
//
// Selection alone would have been enough for the case that looks like it needs
// inheritance: trimming a framed bar with a frameless slab keeps the bar's
// frame on the bar's surface, because the bar is what binds there. So the cost
// of this choice is confined to a frameless shape UNIONED into a framed group.
// If that ever bites, the fix is to let a fieldSpec say whether an operand
// without the property joins the blend carrying the fallback, and this test is
// what changes.
func TestFrameIsInheritedAcrossAUnion(t *testing.T) {
	pkg := twoFramedPkg(bin(program.Operation_MIN, ref("near"), ref("far")), num(10), num(20), nil, nil)

	if got := frameAt(t, pkg, uProperty, point{x: 2}); math.Abs(got-10) > 1e-6 {
		t.Errorf("u over the frameless shape = %v, want the group's 10 by inheritance", got)
	}
}

// A package whose shapes carry no frame at all resolves to nothing, so a
// program that never uses one pays for none of this — the same way a program
// that gives no material keeps a constant record.
func TestNoFrameAnywhereResolvesToNoField(t *testing.T) {
	pkg := twoFramedPkg(bin(program.Operation_MIN, ref("near"), ref("far")), nil, nil, nil, nil)

	fields, err := UFields(pkg, nil)
	if err != nil {
		t.Fatalf("UFields: %v", err)
	}
	if fields != nil {
		t.Errorf("a package with no u resolved %d fields, want none", len(fields))
	}
}
