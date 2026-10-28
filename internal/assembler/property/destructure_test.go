package property

import (
	"math"
	"testing"

	"gradientcontrol.com/omega/core/program"
)

// fnCallOp is a call operand to a same-package function.
func fnCallOp(name string, args ...*program.Operand) *program.Operand {
	return call(name, args...)
}

// multiAssign is `names... = expr` — a destructuring assignment. vars carry no
// properties; use multiAssignVars to attach them.
func multiAssign(expr *program.Operand, names ...string) *program.Assignment {
	vs := make([]*program.Variable, len(names))
	for i, n := range names {
		vs[i] = &program.Variable{Name: n}
	}
	return &program.Assignment{Variables: vs, Expression: expr}
}

// TestDestructureDeriv differentiates through names bound by a destructuring
// assignment: `a, b = f()` where f returns two fields. Before multiDecls such a
// name resolved as an opaque given, so its derivative was silently zero.
func TestDestructureDeriv(t *testing.T) {
	// f() { return x*x*time, sin(time); }
	f := &program.Function{Returns: []*program.Operand{
		mul(mul(ref(program.XVarName), ref(program.XVarName)), ref(program.TimeVarName)),
		un(program.Operation_SIN, ref(program.TimeVarName)),
	}}
	callF := fnCallOp("f")
	d := &derivative{
		decls: map[string]*program.Operand{},
		multiDecls: map[string]multiBinding{
			"a": {value: callF, index: 0},
			"b": {value: callF, index: 1},
		},
		funcs: map[string][]*program.Function{"f": {f}},
	}

	da, err := d.temporal(ref("a"))
	if err != nil {
		t.Fatalf("temporal(a): %v", err)
	}
	db, err := d.temporal(ref("b"))
	if err != nil {
		t.Fatalf("temporal(b): %v", err)
	}
	for _, pt := range []point{{x: 0.7, time: 0.4}, {x: -1.3, time: 1.1}} {
		// d(x*x*time)/dtime = x*x; d(sin(time))/dtime = cos(time).
		if got, want := evalScalar(t, da, pt), pt.x*pt.x; !closeEnough(got, want) {
			t.Errorf("da at %+v: got %v want %v", pt, got, want)
		}
		if got, want := evalScalar(t, db, pt), math.Cos(pt.time); !closeEnough(got, want) {
			t.Errorf("db at %+v: got %v want %v", pt, got, want)
		}
	}
}

// TestDestructureExpand runs the destructure through the package-level Expand
// pass: the DERIV of a destructured name expands via the file's multi-variable
// assignment.
func TestDestructureExpand(t *testing.T) {
	f := &program.Function{Returns: []*program.Operand{
		mul(ref(program.TimeVarName), num(3)),
		mul(ref(program.TimeVarName), ref(program.TimeVarName)),
	}}
	pkg := &program.Package{Files: []*program.PackageFile{{
		FnTypes: []*program.FnType{{Name: "f", Groups: []*program.FnGroup{{
			Functions: []*program.Function{f},
		}}}},
		Assignments: []*program.Assignment{multiAssign(fnCallOp("f"), "a", "b")},
		Returns: []*program.Operand{
			nary(program.Operation_DERIV, ref("b"), ref(program.TimeVarName)),
		},
	}}}
	prog, err := Expand(&program.Program{Primary: pkg})
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	out := prog.Primary
	got := out.Files[0].Returns[0]
	for _, pt := range []point{{time: 0.5}, {time: -2.0}} {
		// d(time*time)/dtime = 2*time.
		if want := 2 * pt.time; !closeEnough(evalScalar(t, got, pt), want) {
			t.Errorf("at %+v: got %v want %v", pt, evalScalar(t, got, pt), want)
		}
	}
}

// TestDestructureColor resolves colors on shapes bound by a destructuring
// assignment: `s, r = twoShapes()`, each with its own .color, combined by
// min(s, r). Inside each sphere the blend weight picks that sphere's color.
// Exercises collectPackageColors (properties on a multi-variable assignment)
// and the gradient engine resolving the names through multiDecls.
func TestDestructureColor(t *testing.T) {
	// twoShapes() { return <sphere@origin>, <sphere@(3,0,0)>; }
	twoShapes := &program.Function{Returns: []*program.Operand{
		sphereSDF(0, 0, 0, 1),
		sphereSDF(3, 0, 0, 1),
	}}
	assign := multiAssign(fnCallOp("twoShapes"), "s", "r")
	assign.Variables[0].Properties = []*program.Property{{Name: colorProperty, Expression: rgb(1, 0, 0)}}
	assign.Variables[1].Properties = []*program.Property{{Name: colorProperty, Expression: rgb(0, 0, 1)}}
	pkg := &program.Package{Files: []*program.PackageFile{{
		FnTypes: []*program.FnType{{Name: "twoShapes", Groups: []*program.FnGroup{{
			Functions: []*program.Function{twoShapes},
		}}}},
		Assignments: []*program.Assignment{assign},
		Returns:     []*program.Operand{bin(program.Operation_MIN, ref("s"), ref("r"))},
	}}}

	fields, err := ColorFields(pkg, nil)
	if err != nil {
		t.Fatalf("ColorFields: %v", err)
	}
	if len(fields) != 1 || fields[0] == nil {
		t.Fatalf("expected one color field, got %v", fields)
	}
	// Substitute the destructured names' fields (the function's returns) so the
	// numeric evaluator can run the emitted refs.
	op := flattenRemap(fields[0])
	op = substRef(op, "s", sphereSDF(0, 0, 0, 1))
	op = substRef(op, "r", sphereSDF(3, 0, 0, 1))

	red := evalVec(t, op, point{})      // inside s
	blue := evalVec(t, op, point{x: 3}) // inside r
	if !closeEnough(red[0], 1) || !closeEnough(red[2], 0) {
		t.Errorf("inside s: got %v, want red", red)
	}
	if !closeEnough(blue[0], 0) || !closeEnough(blue[2], 1) {
		t.Errorf("inside r: got %v, want blue", blue)
	}
}

// TestDestructureRemap: a destructuring assignment whose right-hand side is a
// remap of a multi-return call — `a, b = remap(f(), q)`, the aerospace
// bracket's `clevisShape, horizPlane = remap(arcShape(), clevisRot)`. Remap
// distributes over the returns, so each name binds its own return evaluated
// under the rebound coordinate.
func TestDestructureRemap(t *testing.T) {
	// f() { return x, y; }  under q = p - [1, 2, 0]
	f := &program.Function{Returns: []*program.Operand{
		ref(program.XVarName),
		ref(program.YVarName),
	}}
	coord := nary(program.Operation_COLLECT,
		sub(ref(program.XVarName), num(1)),
		sub(ref(program.YVarName), num(2)),
		ref(program.ZVarName))
	remapped := nary(program.Operation_REMAP, fnCallOp("f"), coord)

	pkg := &program.Package{Files: []*program.PackageFile{{
		FnTypes: []*program.FnType{{Name: "f", Groups: []*program.FnGroup{{
			Functions: []*program.Function{f},
		}}}},
		Assignments: []*program.Assignment{multiAssign(remapped, "a", "b")},
		// d(a)/dp = d(x-1)/dp = (1,0,0); d(b)/dp = d(y-2)/dp = (0,1,0).
		Returns: []*program.Operand{
			nary(program.Operation_DERIV, ref("a"), ref(program.PVarName)),
			nary(program.Operation_DERIV, ref("b"), ref(program.PVarName)),
		},
	}}}

	prog, err := Expand(&program.Program{Primary: pkg})
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	out := prog.Primary
	pt := point{x: 0.4, y: -0.7, z: 0.2}
	for i, want := range [][]float64{{1, 0, 0}, {0, 1, 0}} {
		got := evalVec(t, flattenRemap(out.Files[0].Returns[i]), pt)
		for k := range want {
			if !closeEnough(got[k], want[k]) {
				t.Errorf("return %d component %d: got %v want %v (full %v)", i, k, got[k], want[k], got)
			}
		}
	}
}
