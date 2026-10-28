package property

import (
	"testing"

	"gradientcontrol.com/omega/core/program"
)

// A library function that differentiates what it is handed, called qualified
// from a document, as machined.turned reads the gradient of the frame
// coordinate it is given. Its companion has to be generated in the library's
// own package, where its body resolves, and the call rewritten to reach it
// there with the argument's tangents (doc/material_composition.md, "Datums are
// two different things").

// gradLib is a library with grad(v) { return v.dp; } and a second overload
// grad(v, s) { return s * v.dp; }, both differentiating their first parameter.
func gradLib() *program.Package {
	grad1 := &program.Function{
		Parameters: []*program.Parameter{{Name: "v"}},
		Returns:    []*program.Operand{nary(program.Operation_DERIV, ref("v"), ref(program.PVarName))},
	}
	grad2 := &program.Function{
		Parameters: []*program.Parameter{{Name: "v"}, {Name: "s"}},
		Returns: []*program.Operand{mul(ref("s"),
			nary(program.Operation_DERIV, ref("v"), ref(program.PVarName)))},
	}
	return &program.Package{Path: "lib", Files: []*program.PackageFile{{
		FnTypes: []*program.FnType{{Name: "grad", Groups: []*program.FnGroup{
			{ParamCount: 1, Functions: []*program.Function{grad1}},
			{ParamCount: 2, Functions: []*program.Function{grad2}},
		}}},
	}}}
}

func qualifiedCall(prefix, name string, args ...*program.Operand) *program.Operand {
	return &program.Operand{T: &program.Operand_FnCall{FnCall: &program.FnCall{
		Reference: &program.Reference{Prefix: prefix, Name: name}, Arguments: args}}}
}

// callTo requires op to be a call of name through prefix with argc arguments.
func callTo(t *testing.T, op *program.Operand, prefix, name string, argc int) *program.FnCall {
	t.Helper()
	fc, ok := op.T.(*program.Operand_FnCall)
	if !ok || fc.FnCall.Reference.Prefix != prefix || fc.FnCall.Reference.Name != name {
		t.Fatalf("not a call to %s.%s: %s", prefix, name, op.PrettyPrint(0))
	}
	if len(fc.FnCall.Arguments) != argc {
		t.Fatalf("%s.%s called with %d arguments, want %d", prefix, name, len(fc.FnCall.Arguments), argc)
	}
	return fc.FnCall
}

// lib.grad(x * x) from the document: the companion goes into lib, the call
// reaches it qualified, and the tangents it passes are those of x * x.
func TestQualifiedCallReachesACompanionInItsOwnPackage(t *testing.T) {
	prog, err := Expand(&program.Program{
		Primary:      importPkg(qualifiedCall("lib", "grad", mul(ref(program.XVarName), ref(program.XVarName)))),
		Dependencies: []*program.Package{gradLib()},
	})
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}

	// Two overloads of grad, so the companions carry their arity.
	if findFunction(prog.Primary, "grad__d1") != nil {
		t.Errorf("the companion was woven into the document, not the library")
	}
	comp := findFunction(prog.Dependencies[0], "grad__d1")
	if comp == nil {
		t.Fatal("companion grad__d1 was not generated in lib")
	}
	wantParams := []string{"v", "v__dp0", "v__dp1", "v__dp2"}
	for i, want := range wantParams {
		if i >= len(comp.Parameters) || comp.Parameters[i].Name != want {
			t.Fatalf("companion params = %v, want %v", comp.Parameters, wantParams)
		}
	}

	fc := callTo(t, prog.Primary.Files[0].Returns[0], "lib", "grad__d1", 4)
	pt := point{x: 0.6, y: -1.1, z: 0.35}
	for k, want := range []float64{2 * 0.6, 0, 0} {
		got := evalVec(t, fc.Arguments[1+k], pt)
		if len(got) != 1 || !closeEnough(got[0], want) {
			t.Errorf("tangent %d: got %v want %v", k, got, want)
		}
	}
}

// The overload taking two arguments gets its own companion, named apart.
func TestQualifiedCallPicksTheOverloadByArity(t *testing.T) {
	prog, err := Expand(&program.Program{
		Primary: importPkg(
			qualifiedCall("lib", "grad", ref(program.XVarName)),
			qualifiedCall("lib", "grad", ref(program.YVarName), num(3))),
		Dependencies: []*program.Package{gradLib()},
	})
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	callTo(t, prog.Primary.Files[0].Returns[0], "lib", "grad__d1", 4)
	fc := callTo(t, prog.Primary.Files[0].Returns[1], "lib", "grad__d2", 5)
	comp := findFunction(prog.Dependencies[0], "grad__d2")
	if comp == nil {
		t.Fatal("companion grad__d2 was not generated in lib")
	}
	if n := len(comp.Parameters); n != 5 {
		t.Fatalf("grad__d2 has %d parameters, want 5 (v, s and v's three tangents)", n)
	}
	// The tangent of y along y is 1, along x and z 0.
	pt := point{x: 0.2, y: 0.4, z: -0.7}
	for k, want := range []float64{0, 1, 0} {
		got := evalVec(t, fc.Arguments[2+k], pt)
		if len(got) != 1 || !closeEnough(got[0], want) {
			t.Errorf("tangent %d: got %v want %v", k, got, want)
		}
	}
}

// A document function that only forwards its parameter into lib.grad is
// tainted through the qualified call: it gets a companion of its own, whose
// body calls lib's companion with the tangents it now carries.
func TestQualifiedCallTaintsTheCaller(t *testing.T) {
	through := &program.Function{
		Parameters: []*program.Parameter{{Name: "a"}},
		Returns:    []*program.Operand{qualifiedCall("lib", "grad", ref("a"))},
	}
	primary := importPkg(call("through", mul(ref(program.XVarName), ref(program.XVarName))))
	primary.Files[0].FnTypes = []*program.FnType{{Name: "through", Groups: []*program.FnGroup{
		{ParamCount: 1, Functions: []*program.Function{through}},
	}}}
	prog, err := Expand(&program.Program{
		Primary:      primary,
		Dependencies: []*program.Package{gradLib()},
	})
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	comp := findFunction(prog.Primary, "through__d")
	if comp == nil {
		t.Fatal("through was not given a companion, though its qualified callee differentiates")
	}
	if n := len(comp.Parameters); n != 4 {
		t.Fatalf("through__d has %d parameters, want a and its three tangents", n)
	}
	inner := callTo(t, comp.Returns[0], "lib", "grad__d1", 4)
	for k := 0; k < 3; k++ {
		r, ok := inner.Arguments[1+k].T.(*program.Operand_Reference)
		if !ok || r.Reference.Name != dpParamName("a", k) {
			t.Errorf("tangent %d passed on is %s, want %s", k, inner.Arguments[1+k].PrettyPrint(0), dpParamName("a", k))
		}
	}
	callTo(t, prog.Primary.Files[0].Returns[0], "", "through__d", 4)
}
