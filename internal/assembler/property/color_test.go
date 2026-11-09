package property

import (
	"testing"

	"gradientcontrol.com/omega/core/program"
)

// rgb builds an RGB color literal operand.
func rgb(r, g, b float64) *program.Operand {
	return nary(program.Operation_COLLECT, num(r), num(g), num(b))
}

// sphereSDF is length(p - center) - radius.
func sphereSDF(cx, cy, cz, radius float64) *program.Operand {
	return sub(call("length", sub(ref(program.PVarName), rgb(cx, cy, cz))), num(radius))
}

func coloredShape(name string, sdf, color *program.Operand) *program.Assignment {
	v := &program.Variable{Name: name}
	if color != nil {
		v.Properties = []*program.Property{{Name: colorProperty, Expression: color}}
	}
	return &program.Assignment{
		Variables:  []*program.Variable{v},
		Expression: sdf,
	}
}

// twoBallPkg is two unit spheres — "ball" (red) at the origin and "box" (blue)
// at (3,0,0) — combined by the given return.
func twoBallPkg(ret *program.Operand) *program.Package {
	return &program.Package{
		Files: []*program.PackageFile{
			{
				Assignments: []*program.Assignment{
					coloredShape("ball", sphereSDF(0, 0, 0, 1), rgb(1, 0, 0)),
					coloredShape("box", sphereSDF(3, 0, 0, 1), rgb(0, 0, 1)),
				},
				Returns: []*program.Operand{ret},
			},
		},
	}
}

// colorAt runs ColorFields (no imports — these returns use inline ops only),
// substitutes the named shapes' SDF bodies for their live references, and
// evaluates the single color field at pt.
func colorAt(t *testing.T, pkg *program.Package, pt point) []float64 {
	t.Helper()
	fields, err := ColorFields(pkg, nil)
	if err != nil {
		t.Fatalf("ColorFields: %v", err)
	}
	if len(fields) != 1 || fields[0] == nil {
		t.Fatalf("expected one color field, got %d", len(fields))
	}
	t.Logf("BEFORE FLATTEN: %s", fields[0].PrettyPrint(0))
	op := flattenRemap(fields[0])
	t.Logf("AFTER FLATTEN: %s", op.PrettyPrint(0))
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, a := range f.Assignments {
			if a == nil || len(a.Variables) != 1 {
				continue
			}
			op = substRef(op, a.Variables[0].Name, a.Expression)
		}
	}
	c := evalVec(t, op, pt)
	if len(c) != 3 {
		t.Fatalf("expected vec3 color, got len %d", len(c))
	}
	return c
}

// flattenRemap eliminates REMAP(field, coord) by substituting p (and x/y/z)
// with coord in the field, so the numeric evaluator (which has no REMAP rule)
// can evaluate the transported color the same way the backend lowers it.
func flattenRemap(op *program.Operand) *program.Operand {
	if op == nil {
		return nil
	}
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		e := t.Expression
		if e.Operation == program.Operation_REMAP {
			field := flattenRemap(e.Operands[0])
			coord := flattenRemap(e.Operands[1])
			field = substRef(field, program.PVarName, coord)
			field = substRef(field, program.XVarName, nary(program.Operation_ACCESS, coord, ref(program.XVarName)))
			field = substRef(field, program.YVarName, nary(program.Operation_ACCESS, coord, ref(program.YVarName)))
			field = substRef(field, program.ZVarName, nary(program.Operation_ACCESS, coord, ref(program.ZVarName)))
			return field
		}
		ops := make([]*program.Operand, len(e.Operands))
		for i, o := range e.Operands {
			ops[i] = flattenRemap(o)
		}
		return &program.Operand{T: &program.Operand_Expression{Expression: &program.Expression{Operation: e.Operation, Operands: ops}}}
	case *program.Operand_FnCall:
		fc := t.FnCall
		args := make([]*program.Operand, len(fc.Arguments))
		for i, a := range fc.Arguments {
			args[i] = flattenRemap(a)
		}
		return &program.Operand{T: &program.Operand_FnCall{FnCall: &program.FnCall{
			Reference: &program.Reference{Prefix: fc.Reference.Prefix, Name: fc.Reference.Name}, Arguments: args}}}
	}
	return op
}

func expectColor(t *testing.T, got []float64, wr, wg, wb float64) {
	t.Helper()
	if !closeEnough(got[0], wr) || !closeEnough(got[1], wg) || !closeEnough(got[2], wb) {
		t.Errorf("color = [%v %v %v], want ~[%v %v %v]", got[0], got[1], got[2], wr, wg, wb)
	}
}

// TestColorSharpUnion: min(ball, box) — ∂min/∂dᵢ ∈ {0,1} gives a hard select of
// the nearer surface's color.
func TestColorSharpUnion(t *testing.T) {
	pkg := twoBallPkg(bin(program.Operation_MIN, ref("ball"), ref("box")))
	expectColor(t, colorAt(t, pkg, point{x: 1}), 1, 0, 0) // near ball -> red
	expectColor(t, colorAt(t, pkg, point{x: 2}), 0, 0, 1) // near box  -> blue
}

// TestColorSharpIntersection: max(ball, box) selects the binding (farther)
// surface's color.
func TestColorSharpIntersection(t *testing.T) {
	pkg := twoBallPkg(bin(program.Operation_MAX, ref("ball"), ref("box")))
	// At x=1: ball=0, box=1; the max binds on box -> blue.
	expectColor(t, colorAt(t, pkg, point{x: 1}), 0, 0, 1)
}

// TestColorSmoothUnionBlends: a log-sum-exp smooth-min (the form the omega
// library uses) blends the two colors by its softmin weights ∂D/∂dᵢ, which are
// a genuine convex partition of unity. Even mix at the seam; red-dominant toward
// ball. (An IQ *polynomial* smin, not in the library, has sensitivities outside
// [0,1] — the gradient method assumes the library's monotone combiners.)
func TestColorSmoothUnionBlends(t *testing.T) {
	a, b := ref("ball"), ref("box")
	rad := num(0.4)
	// smin = -rad·ln(exp(-a/rad) + exp(-b/rad)); ∂/∂a = softmin(a) ∈ [0,1].
	smin := neg(mul(rad, un(program.Operation_LN, add(
		un(program.Operation_EXP, div(neg(a), rad)),
		un(program.Operation_EXP, div(neg(b), rad)),
	))))
	pkg := twoBallPkg(smin)

	// Equidistant seam (x=1.5): even mix.
	expectColor(t, colorAt(t, pkg, point{x: 1.5}), 0.5, 0, 0.5)
	// Toward ball (x=1.2): red dominant, some blue, summing to ~1.
	c := colorAt(t, pkg, point{x: 1.2})
	if c[0] <= 0.6 || c[2] >= 0.4 || !closeEnough(c[0]+c[2], 1) {
		t.Errorf("near-ball color = %v, want red-dominant convex blend", c)
	}
}

// TestColorReduceUnion: a local list-union built with reduce (exercising the
// engine's fold unrolling) differentiates to the same hard select as min.
func TestColorReduceUnion(t *testing.T) {
	// lambda(acc, v) { return min(acc, v); }
	lambda := &program.Function{
		Parameters: []*program.Parameter{{Name: "acc"}, {Name: "v"}},
		Returns:    []*program.Operand{bin(program.Operation_MIN, ref("acc"), ref("v"))},
	}
	// unionList(list) { return reduce(lambda, list); }  (no-initial fold)
	unionList := &program.Function{
		Parameters: []*program.Parameter{{Name: "list"}},
		FnTypes: []*program.FnType{
			{
				Name: "lambda",
				Groups: []*program.FnGroup{
					{
						ParamCount: 2,
						Functions:  []*program.Function{lambda},
					},
				},
			},
		},
		Returns: []*program.Operand{nary(program.Operation_REDUCE,
			ref("lambda"),
			ref("list"),
		)},
	}
	pkg := twoBallPkg(call("unionList", nary(program.Operation_COLLECT, ref("ball"), ref("box"))))
	// Add unionList to the main file's FnTypes
	pkg.Files[0].FnTypes = append(pkg.Files[0].FnTypes, &program.FnType{
		Name: "unionList",
		Groups: []*program.FnGroup{
			{
				ParamCount: 1,
				Functions:  []*program.Function{unionList},
			},
		},
	})

	expectColor(t, colorAt(t, pkg, point{x: 1}), 1, 0, 0)
	expectColor(t, colorAt(t, pkg, point{x: 2}), 0, 0, 1)
}

// TestColorProceduralField: a spatially varying color propagates as a field.
func TestColorProceduralField(t *testing.T) {
	varying := nary(program.Operation_COLLECT, un(program.Operation_ABS, ref(program.XVarName)), num(0), num(0))
	pkg := &program.Package{
		Files: []*program.PackageFile{
			{
				Assignments: []*program.Assignment{coloredShape("ball", sphereSDF(0, 0, 0, 1), varying)},
				Returns:     []*program.Operand{ref("ball")},
			},
		},
	}
	expectColor(t, colorAt(t, pkg, point{x: 0.5}), 0.5, 0, 0)
	expectColor(t, colorAt(t, pkg, point{x: -0.75}), 0.75, 0, 0)
}

// TestColorUncoloredRegionInheritsComposed: a colored ball unioned with an
// uncolored plane shows the ball's color everywhere — on the ball, and over the
// plane, which inherits the color of the colored object it is composed with (the
// ball is the only, hence nearest, colored shape). No flat default gray.
func TestColorUncoloredRegionInheritsComposed(t *testing.T) {
	pkg := &program.Package{
		Files: []*program.PackageFile{
			{
				Assignments: []*program.Assignment{
					coloredShape("ball", sphereSDF(0, 0, 0, 1), rgb(1, 0, 0)),
					coloredShape("flr", ref(program.ZVarName), nil), // plane z=0, uncolored
				},
				Returns: []*program.Operand{bin(program.Operation_MIN, ref("ball"), ref("flr"))},
			},
		},
	}
	// Inside the ball (ball nearer): red.
	expectColor(t, colorAt(t, pkg, point{x: 0, y: 0, z: 0.5}), 1, 0, 0)
	// Out over the plane region (far from ball, plane nearer): inherits the
	// ball's red — the color it is composed with — not a default gray.
	expectColor(t, colorAt(t, pkg, point{x: 6, y: 0, z: 0}), 1, 0, 0)
}

// TestColorCompositionalNesting: color is resolved compositionally, so a named
// intermediate union acts as a colored unit. `inner = min(aa, bb)` groups a red
// shape with an uncolored one; `outer = min(inner, cc)` unions that with a blue
// shape. bb's region must inherit aa's red *through inner*, even though the blue
// cc is nearer — the grouping carries the color. The same three shapes unioned
// flat (no `inner`) instead let bb inherit the nearest colored shape (cc), which
// is exactly what the nesting overrides.
func TestColorCompositionalNesting(t *testing.T) {
	red, blue := rgb(1, 0, 0), rgb(0, 0, 1)
	assign := func(name string, body *program.Operand) *program.Assignment {
		return &program.Assignment{
			Variables:  []*program.Variable{{Name: name}},
			Expression: body,
		}
	}
	// aa (red) at x=0, uncolored bb at x=5, cc (blue) at x=8; radius 1 each.
	// Declarations are in reverse-topological order (a decl before the ones it
	// references) so colorAt's single-pass substitution resolves the nesting.
	nested := &program.Package{
		Files: []*program.PackageFile{
			{
				Assignments: []*program.Assignment{
					assign("outer", bin(program.Operation_MIN, ref("inner"), ref("cc"))),
					assign("inner", bin(program.Operation_MIN, ref("aa"), ref("bb"))),
					coloredShape("cc", sphereSDF(8, 0, 0, 1), blue),
					coloredShape("aa", sphereSDF(0, 0, 0, 1), red),
					coloredShape("bb", sphereSDF(5, 0, 0, 1), nil), // uncolored
				},
				Returns: []*program.Operand{ref("outer")},
			},
		},
	}
	// Inside bb but off its center (a sphere center is a 0/0 in the SDF gradient,
	// an interior point never sampled for color in practice): bb binds; daa≈4.0,
	// dcc≈2.0 so cc is nearer. The grouping puts bb with aa inside `inner`, so it
	// inherits aa's red, not cc.
	expectColor(t, colorAt(t, nested, point{x: 5, z: 0.5}), 1, 0, 0)
	expectColor(t, colorAt(t, nested, point{x: 0, z: 0.5}), 1, 0, 0) // aa region: red
	expectColor(t, colorAt(t, nested, point{x: 8, z: 0.5}), 0, 0, 1) // cc region: blue

	// Contrast: the same three shapes unioned flat (no named `inner` group) —
	// bb inherits the *nearest* colored shape, cc (dcc≈2.0 < daa≈4.0), i.e. blue.
	flat := &program.Package{
		Files: []*program.PackageFile{
			{
				Assignments: []*program.Assignment{
					coloredShape("aa", sphereSDF(0, 0, 0, 1), red),
					coloredShape("bb", sphereSDF(5, 0, 0, 1), nil),
					coloredShape("cc", sphereSDF(8, 0, 0, 1), blue),
				},
				Returns: []*program.Operand{bin(program.Operation_MIN, bin(program.Operation_MIN, ref("aa"), ref("bb")), ref("cc"))},
			},
		},
	}
	expectColor(t, colorAt(t, flat, point{x: 5, z: 0.5}), 0, 0, 1) // blue — flattened to nearest
}

// TestColorRemapTransport: a procedural color moves with its shape under remap.
// ball is colored by a local gradient r = abs(p.x)*0.5 (0 at its center) and
// translated to x=2 via remap(ball, p-[2,0,0]); the color must follow, so at the
// translated center the local coordinate — not world space — drives the color.
func TestColorRemapTransport(t *testing.T) {
	varying := nary(program.Operation_COLLECT, mul(un(program.Operation_ABS, ref(program.XVarName)), num(0.5)), num(0), num(0))
	off := nary(program.Operation_COLLECT, num(2), num(0), num(0))
	remapped := nary(program.Operation_REMAP, ref("ball"), sub(ref(program.PVarName), off))
	pkg := &program.Package{
		Files: []*program.PackageFile{
			{
				Assignments: []*program.Assignment{
					coloredShape("ball", sphereSDF(0, 0, 0, 1), varying),
					coloredShape("box", sphereSDF(20, 0, 0, 1), rgb(0, 0, 1)), // far away; ball wins near x=2
				},
				Returns: []*program.Operand{bin(program.Operation_MIN, remapped, ref("box"))},
			},
		},
	}
	// At the translated center (x=2) the local coord is 0 -> r ~ 0 (not world-space 1.0).
	expectColor(t, colorAt(t, pkg, point{x: 2}), 0, 0, 0)
	// At x=2.5 the local coord is 0.5 -> r = 0.25.
	expectColor(t, colorAt(t, pkg, point{x: 2.5}), 0.25, 0, 0)
}

// TestColorNoColorNoField: a package with no color property emits no field.
func TestColorNoColorNoField(t *testing.T) {
	pkg := &program.Package{
		Files: []*program.PackageFile{
			{
				Assignments: []*program.Assignment{coloredShape("ball", sphereSDF(0, 0, 0, 1), nil)},
				Returns:     []*program.Operand{bin(program.Operation_MIN, ref("ball"), sphereSDF(3, 0, 0, 1))},
			},
		},
	}
	fields, err := ColorFields(pkg, nil)
	if err != nil {
		t.Fatalf("ColorFields: %v", err)
	}
	if fields != nil {
		t.Errorf("expected no color fields, got %v", fields)
	}
}
