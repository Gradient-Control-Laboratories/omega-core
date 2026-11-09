package property

import (
	"fmt"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/program"
)

// materialLit builds a material record literal: three vec4 rows.
func materialLit(rough, metal, ior, emR, emG, emB, emStrength, transmission, absorption, displacement float64) *program.Operand {
	return nary(program.Operation_COLLECT,
		vec4Literal(rough, metal, ior, 0),
		vec4Literal(emR, emG, emB, emStrength),
		vec4Literal(transmission, absorption, displacement, 0),
		vec4Literal(0, 0, 0, 0))
}

// materialShape is coloredShape for the material property.
func materialShape(name string, sdf, material *program.Operand) *program.Assignment {
	v := &program.Variable{Name: name}
	if material != nil {
		v.Properties = []*program.Property{{Name: materialProperty, Expression: material}}
	}
	return &program.Assignment{
		Variables:  []*program.Variable{v},
		Expression: sdf,
	}
}

// twoMaterialPkg is two unit spheres — "ball" (chrome-like) at the origin and
// "box" (matte-like) at (3,0,0) — combined by the given return.
func twoMaterialPkg(ret *program.Operand) *program.Package {
	return &program.Package{
		Files: []*program.PackageFile{
			{
				Assignments: []*program.Assignment{
					materialShape("ball", sphereSDF(0, 0, 0, 1), materialLit(0.04, 1, 1.5, 0, 0, 0, 0, 0, 0, 0)),
					materialShape("box", sphereSDF(3, 0, 0, 1), materialLit(1.0, 0, 1.45, 1, 0.5, 0, 2, 0, 0, 0)),
				},
				Returns: []*program.Operand{ret},
			},
		},
	}
}

// materialAt runs MaterialFields, substitutes the named shapes' SDF bodies for
// their live references, and evaluates the single material field at pt as
// twelve numbers (three rows of four). The test-only numeric evaluator has no
// INDEX rule and clamps with scalar bounds, so constant indexes into literal
// records are folded here and each row's clamp is applied componentwise.
func materialAt(t *testing.T, pkg *program.Package, pt point) []float64 {
	t.Helper()
	fields, err := MaterialFields(pkg, nil)
	if err != nil {
		t.Fatalf("MaterialFields: %v", err)
	}
	if len(fields) != 1 || fields[0] == nil {
		t.Fatalf("expected one material field, got %d", len(fields))
	}
	op := flattenRemap(fields[0])
	for _, f := range pkg.Files {
		for _, a := range f.Assignments {
			if len(a.Variables) != 1 {
				continue
			}
			op = substRef(op, a.Variables[0].Name, a.Expression)
		}
	}
	op = foldIndex(op)

	// A single-unit fast path returns the record itself; a blend returns
	// COLLECT(clamp(row0…), clamp(row1…), clamp(row2…)).
	coll, ok := op.T.(*program.Operand_Expression)
	if !ok || coll.Expression.Operation != program.Operation_COLLECT || len(coll.Expression.Operands) != materialRows {
		t.Fatalf("expected a %d-row material record, got\n%s", materialRows, op.PrettyPrint(0))
	}
	var out []float64
	for _, row := range coll.Expression.Operands {
		out = append(out, evalRow(t, row, pt)...)
	}
	if len(out) != 4*materialRows {
		t.Fatalf("expected %d numbers, got %d", 4*materialRows, len(out))
	}
	return out
}

// evalRow evaluates one row, applying an outer clamp(v, lo, hi) componentwise.
func evalRow(t *testing.T, row *program.Operand, pt point) []float64 {
	t.Helper()
	if fc, ok := row.T.(*program.Operand_FnCall); ok && fc.FnCall.Reference.Name == "clamp" && len(fc.FnCall.Arguments) == 3 {
		v := evalVec(t, fc.FnCall.Arguments[0], pt)
		lo := evalVec(t, fc.FnCall.Arguments[1], pt)
		hi := evalVec(t, fc.FnCall.Arguments[2], pt)
		out := make([]float64, len(v))
		for i := range v {
			out[i] = math.Max(lo[i], math.Min(hi[i], v[i]))
		}
		return out
	}
	return evalVec(t, row, pt)
}

// foldIndex replaces INDEX(COLLECT(e0, e1, …), k) by e_k throughout the tree.
func foldIndex(op *program.Operand) *program.Operand {
	if op == nil {
		return nil
	}
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		e := t.Expression
		ops := make([]*program.Operand, len(e.Operands))
		for i, o := range e.Operands {
			ops[i] = foldIndex(o)
		}
		if e.Operation == program.Operation_INDEX && len(ops) == 2 {
			if c, ok := ops[0].T.(*program.Operand_Expression); ok && c.Expression.Operation == program.Operation_COLLECT {
				if k, ok := ops[1].T.(*program.Operand_Value); ok {
					return c.Expression.Operands[int(k.Value)]
				}
			}
		}
		return nary(e.Operation, ops...)
	case *program.Operand_FnCall:
		fc := t.FnCall
		args := make([]*program.Operand, len(fc.Arguments))
		for i, a := range fc.Arguments {
			args[i] = foldIndex(a)
		}
		return &program.Operand{T: &program.Operand_FnCall{FnCall: &program.FnCall{
			Reference: &program.Reference{Prefix: fc.Reference.Prefix, Name: fc.Reference.Name}, Arguments: args}}}
	}
	return op
}

func expectRow(t *testing.T, got []float64, row int, want ...float64) {
	t.Helper()
	for i, w := range want {
		if !closeEnough(got[4*row+i], w) {
			t.Errorf("row %d = %v, want ~%v", row, got[4*row:4*row+4], want)
			return
		}
	}
}

// TestMaterialSharpUnion: min(ball, box) hard-selects the nearer shape's whole
// record, every row.
func TestMaterialSharpUnion(t *testing.T) {
	pkg := twoMaterialPkg(bin(program.Operation_MIN, ref("ball"), ref("box")))
	near := materialAt(t, pkg, point{x: 1})
	expectRow(t, near, 0, 0.04, 1, 1.5, 0)
	expectRow(t, near, 1, 0, 0, 0, 0)
	far := materialAt(t, pkg, point{x: 2})
	expectRow(t, far, 0, 1.0, 0, 1.45, 0)
	expectRow(t, far, 1, 1, 0.5, 0, 2)
}

// TestMaterialSmoothUnionBlends: a log-sum-exp smooth-min interpolates each
// row by the softmin weights; an even mix at the seam.
func TestMaterialSmoothUnionBlends(t *testing.T) {
	a, b := ref("ball"), ref("box")
	rad := num(0.4)
	smin := neg(mul(rad, un(program.Operation_LN, add(
		un(program.Operation_EXP, div(neg(a), rad)),
		un(program.Operation_EXP, div(neg(b), rad)),
	))))
	pkg := twoMaterialPkg(smin)
	seam := materialAt(t, pkg, point{x: 1.5})
	expectRow(t, seam, 0, 0.52, 0.5, 1.475, 0)
	expectRow(t, seam, 1, 0.5, 0.25, 0, 1)
}

// TestMaterialUnassignedInherits: a shape without a material unioned with one
// that has it takes that material everywhere (the nearest, and only, unit).
func TestMaterialUnassignedInherits(t *testing.T) {
	pkg := &program.Package{
		Files: []*program.PackageFile{
			{
				Assignments: []*program.Assignment{
					materialShape("ball", sphereSDF(0, 0, 0, 1), materialLit(0.3, 1, 1.5, 0, 0, 0, 0, 0, 0, 0)),
					materialShape("flr", ref(program.ZVarName), nil),
				},
				Returns: []*program.Operand{bin(program.Operation_MIN, ref("ball"), ref("flr"))},
			},
		},
	}
	// The single unit takes the fast path: the record itself, unblended.
	over := materialAt(t, pkg, point{x: 6, y: 0, z: 0})
	expectRow(t, over, 0, 0.3, 1, 1.5, 0)
}

// TestMaterialClampsRows: a blend clamps ior up to 1 and emissive strength
// down to 0, while displacement may stay negative.
func TestMaterialClampsRows(t *testing.T) {
	pkg := &program.Package{
		Files: []*program.PackageFile{
			{
				Assignments: []*program.Assignment{
					materialShape("ball", sphereSDF(0, 0, 0, 1), materialLit(0.5, 0, 0.2, 0, 0, 0, -3, 0, 0, -0.5)),
					materialShape("box", sphereSDF(3, 0, 0, 1), materialLit(0.5, 0, 1.5, 0, 0, 0, 0, 0, 0, 0)),
				},
				Returns: []*program.Operand{bin(program.Operation_MIN, ref("ball"), ref("box"))},
			},
		},
	}
	near := materialAt(t, pkg, point{x: 1})
	expectRow(t, near, 0, 0.5, 0, 1, 0)
	expectRow(t, near, 1, 0, 0, 0, 0)
	expectRow(t, near, 2, 0, 0, -0.5, 0)
}

// TestMaterialIndependentOfColor: a package that authors color but no
// material has no material field, and vice versa.
func TestMaterialIndependentOfColor(t *testing.T) {
	colorOnly := twoBallPkg(bin(program.Operation_MIN, ref("ball"), ref("box")))
	fields, err := MaterialFields(colorOnly, nil)
	if err != nil {
		t.Fatalf("MaterialFields: %v", err)
	}
	if fields != nil {
		t.Errorf("expected no material fields for a color-only package, got %v", fields)
	}
	materialOnly := twoMaterialPkg(bin(program.Operation_MIN, ref("ball"), ref("box")))
	cfields, err := ColorFields(materialOnly, nil)
	if err != nil {
		t.Fatalf("ColorFields: %v", err)
	}
	if cfields != nil {
		t.Errorf("expected no color fields for a material-only package, got %v", cfields)
	}
}

// TestMaterialDefaultMatchesSpec pins the default record the compiler,
// library and viewers share.
func TestMaterialDefaultMatchesSpec(t *testing.T) {
	got := evalVec(t, materialDefault(), point{})
	want := []float64{0.5, 0, 1.5, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("materialDefault = %v, want %v", got, want)
	}
}
