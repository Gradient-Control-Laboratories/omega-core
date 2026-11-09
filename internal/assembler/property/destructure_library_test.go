package property

import (
	"testing"

	"gradientcontrol.com/omega/core/program"
)

// A LIBRARY THAT RETURNS TWO SHAPES. fastener.setScrewSoftTip is the case that
// found this: a steel body and a nylon pad, multireturned so that each can carry
// its own material, which only works if a destructured library call resolves.
//
//	lib: pair() { return length(p) - 1, length(p - [3, 0, 0]) - 1; }
func pairLibrary() program.GetImportFunc {
	pair := &program.Function{Returns: []*program.Operand{
		sphereSDF(0, 0, 0, 1),
		sphereSDF(3, 0, 0, 1),
	}}
	lib := &program.Package{Files: []*program.PackageFile{{
		FnTypes: []*program.FnType{
			{Name: "pair", Groups: []*program.FnGroup{{ParamCount: 0, Functions: []*program.Function{pair}}}},
		},
	}}}
	return func(path string) (*program.Package, error) { return lib, nil }
}

// libCall is the qualified call lib.name(args...).
func libCall(name string, args ...*program.Operand) *program.Operand {
	return &program.Operand{T: &program.Operand_FnCall{FnCall: &program.FnCall{
		Reference: &program.Reference{Prefix: "lib", Name: name}, Arguments: args}}}
}

// destructureWith is `a, b = expr` with the given property on each name.
func destructureWith(expr *program.Operand, property string, values map[string]*program.Operand, names ...string) *program.Assignment {
	vs := make([]*program.Variable, len(names))
	for i, n := range names {
		vs[i] = &program.Variable{Name: n}
		if v, ok := values[n]; ok {
			vs[i].Properties = []*program.Property{{Name: property, Expression: v}}
		}
	}
	return &program.Assignment{Variables: vs, Expression: expr}
}

// fieldAt substitutes each named shape's own field for its live reference in a
// resolved property field and evaluates it at pt.
func fieldAt(t *testing.T, field *program.Operand, shapes map[string]*program.Operand, pt point) []float64 {
	t.Helper()
	op := field
	for name, sdf := range shapes {
		op = substRef(op, name, sdf)
	}
	return evalVec(t, flattenRemap(op), pt)
}

// TestColorDestructuredLibraryMultiReturn: `a, b = lib.pair()`, each name
// colored, unioned. Each half has to keep its own color. It used to lose both —
// and every other color in the same return with them — because the engine
// refused to destructure a qualified call and the resolver answered that by
// dropping the whole return's field.
func TestColorDestructuredLibraryMultiReturn(t *testing.T) {
	pkg := &program.Package{Files: []*program.PackageFile{{
		Imports: []*program.Import{{Path: "lib"}},
		Assignments: []*program.Assignment{
			destructureWith(libCall("pair"), colorProperty,
				map[string]*program.Operand{"a": rgb(1, 0, 0), "b": rgb(0, 0, 1)}, "a", "b"),
		},
		Returns: []*program.Operand{bin(program.Operation_MIN, ref("a"), ref("b"))},
	}}}
	fields, err := ColorFields(pkg, pairLibrary())
	if err != nil {
		t.Fatalf("ColorFields: %v", err)
	}
	if len(fields) != 1 || fields[0] == nil {
		t.Fatalf("expected one color field, got %v", fields)
	}
	shapes := map[string]*program.Operand{"a": sphereSDF(0, 0, 0, 1), "b": sphereSDF(3, 0, 0, 1)}
	expectColor(t, fieldAt(t, fields[0], shapes, point{x: 1}), 1, 0, 0) // on a -> red
	expectColor(t, fieldAt(t, fields[0], shapes, point{x: 2}), 0, 0, 1) // on b -> blue
}

// TestColorDestructuredLibraryUnderRemap: the fastener idiom places the pair
// first, `a, b = remap(lib.pair(), p - [10, 0, 0])`, so the destructure has to
// see through the remap to the qualified call inside it.
func TestColorDestructuredLibraryUnderRemap(t *testing.T) {
	q := sub(ref(program.PVarName), rgb(10, 0, 0))
	pkg := &program.Package{Files: []*program.PackageFile{{
		Imports: []*program.Import{{Path: "lib"}},
		Assignments: []*program.Assignment{
			destructureWith(nary(program.Operation_REMAP, libCall("pair"), q), colorProperty,
				map[string]*program.Operand{"a": rgb(1, 0, 0), "b": rgb(0, 0, 1)}, "a", "b"),
		},
		Returns: []*program.Operand{bin(program.Operation_MIN, ref("a"), ref("b"))},
	}}}
	fields, err := ColorFields(pkg, pairLibrary())
	if err != nil {
		t.Fatalf("ColorFields: %v", err)
	}
	if len(fields) != 1 || fields[0] == nil {
		t.Fatalf("expected one color field, got %v", fields)
	}
	shapes := map[string]*program.Operand{
		"a": nary(program.Operation_REMAP, sphereSDF(0, 0, 0, 1), q),
		"b": nary(program.Operation_REMAP, sphereSDF(3, 0, 0, 1), q),
	}
	expectColor(t, fieldAt(t, fields[0], shapes, point{x: 11}), 1, 0, 0)
	expectColor(t, fieldAt(t, fields[0], shapes, point{x: 12}), 0, 0, 1)
}

// TestMaterialDestructuredLibraryMultiReturn: the same pair carrying materials,
// which is what the multi-material set screws are for.
func TestMaterialDestructuredLibraryMultiReturn(t *testing.T) {
	steel := materialLit(0.45, 1, 1.5, 0, 0, 0, 0, 0, 0, 0)
	nylon := materialLit(0.35, 0, 1.46, 0, 0, 0, 0, 0, 0, 0)
	pkg := &program.Package{Files: []*program.PackageFile{{
		Imports: []*program.Import{{Path: "lib"}},
		Assignments: []*program.Assignment{
			destructureWith(libCall("pair"), materialProperty,
				map[string]*program.Operand{"a": steel, "b": nylon}, "a", "b"),
		},
		Returns: []*program.Operand{bin(program.Operation_MIN, ref("a"), ref("b"))},
	}}}
	fields, err := MaterialFields(pkg, pairLibrary())
	if err != nil {
		t.Fatalf("MaterialFields: %v", err)
	}
	if len(fields) != 1 || fields[0] == nil {
		t.Fatalf("expected one material field, got %v", fields)
	}
	shapes := map[string]*program.Operand{"a": sphereSDF(0, 0, 0, 1), "b": sphereSDF(3, 0, 0, 1)}
	row0 := func(pt point) []float64 {
		op := fields[0]
		for name, sdf := range shapes {
			op = substRef(op, name, sdf)
		}
		return evalRow(t, foldIndex(nary(program.Operation_INDEX, flattenRemap(op), num(0))), pt)
	}
	expectRow(t, row0(point{x: 1}), 0, 0.45, 1, 1.5, 0)  // on a -> steel
	expectRow(t, row0(point{x: 2}), 0, 0.35, 0, 1.46, 0) // on b -> nylon
}

// TestColorOneBadUnitKeepsTheRest: a return whose shapes include one the color
// gradient cannot evaluate — here a local function that destructures a library
// call inside its own body, where there is no live name to refer to — still
// colors everything else. The failure used to take the whole return's field
// with it, so one exotic part left an entire board on the default palette.
func TestColorOneBadUnitKeepsTheRest(t *testing.T) {
	// two() { x, y = lib.pair(); return min(x, y); }
	two := &program.Function{
		Assignments: []*program.Assignment{multiAssign(libCall("pair"), "x", "y")},
		Returns:     []*program.Operand{bin(program.Operation_MIN, ref("x"), ref("y"))},
	}
	pkg := &program.Package{Files: []*program.PackageFile{{
		Imports: []*program.Import{{Path: "lib"}},
		Assignments: []*program.Assignment{
			coloredShape("u", call("two"), rgb(1, 0, 0)),
			coloredShape("c", sphereSDF(0, 6, 0, 1), rgb(0, 1, 0)),
		},
		FnTypes: []*program.FnType{
			{Name: "two", Groups: []*program.FnGroup{{ParamCount: 0, Functions: []*program.Function{two}}}},
		},
		Returns: []*program.Operand{bin(program.Operation_MIN, ref("u"), ref("c"))},
	}}}
	fields, err := ColorFields(pkg, pairLibrary())
	if err != nil {
		t.Fatalf("ColorFields: %v", err)
	}
	if len(fields) != 1 || fields[0] == nil {
		t.Fatalf("expected one color field, got %v", fields)
	}
	shapes := map[string]*program.Operand{
		"u": bin(program.Operation_MIN, sphereSDF(0, 0, 0, 1), sphereSDF(3, 0, 0, 1)),
		"c": sphereSDF(0, 6, 0, 1),
	}
	expectColor(t, fieldAt(t, fields[0], shapes, point{x: 1}), 1, 0, 0) // on u -> red
	expectColor(t, fieldAt(t, fields[0], shapes, point{y: 5}), 0, 1, 0) // on c -> green
}

// TestDerivDestructuredLibraryMultiReturn: a derivative has no opaque view of a
// library call — it differentiates through the body — so it destructures one by
// inlining the imported function and taking the name's own return.
func TestDerivDestructuredLibraryMultiReturn(t *testing.T) {
	getImport := pairLibrary()
	callPair := libCall("pair")
	pkg := &program.Package{Files: []*program.PackageFile{{
		Imports:     []*program.Import{{Path: "lib"}},
		Assignments: []*program.Assignment{multiAssign(callPair, "a", "b")},
	}}}
	d := &derivative{
		decls: map[string]*program.Operand{},
		multiDecls: map[string]multiBinding{
			"a": {value: callPair, index: 0},
			"b": {value: callPair, index: 1},
		},
		funcs:   map[string][]*program.Function{},
		imports: collectImportNamespaces(pkg, getImport),
	}
	db, err := d.gradient(ref("b"))
	if err != nil {
		t.Fatalf("gradient(b): %v", err)
	}
	// b is the unit sphere at (3, 0, 0): its gradient at (5, 0, 0) is +x.
	got := evalVec(t, db, point{x: 5})
	if len(got) != 3 || !closeEnough(got[0], 1) || !closeEnough(got[1], 0) || !closeEnough(got[2], 0) {
		t.Errorf("grad b at (5,0,0) = %v, want [1 0 0]", got)
	}
}
