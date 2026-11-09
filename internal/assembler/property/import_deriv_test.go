package property

import (
	"testing"

	"gradientcontrol.com/omega/core/program"
)

// libPkg is a stand-in for a qualified library (map, noise, …): one scalar
// function and one coordinate transform, plus a constant.
func libPkg() *program.Package {
	twice := &program.Function{
		Parameters: []*program.Parameter{{Name: "v"}},
		Returns:    []*program.Operand{mul(ref("v"), num(2))},
	}
	// swapXY(v) { return [v.y, v.x, v.z]; }
	swapXY := &program.Function{
		Parameters: []*program.Parameter{{Name: "v"}},
		Returns: []*program.Operand{nary(program.Operation_COLLECT,
			nary(program.Operation_ACCESS, ref("v"), ref(program.YVarName)),
			nary(program.Operation_ACCESS, ref("v"), ref(program.XVarName)),
			nary(program.Operation_ACCESS, ref("v"), ref(program.ZVarName)))},
	}
	return &program.Package{Files: []*program.PackageFile{{
		Assignments: []*program.Assignment{assignDecl("half", num(0.5))},
		FnTypes: []*program.FnType{
			{Name: "twice", Groups: []*program.FnGroup{{ParamCount: 1, Functions: []*program.Function{twice}}}},
			{Name: "swapXY", Groups: []*program.FnGroup{{ParamCount: 1, Functions: []*program.Function{swapXY}}}},
		},
	}}}
}

func libImport(path string) (*program.Package, error) { return libPkg(), nil }

// qualCall builds a `lib.name(args...)` call operand.
func qualCall(name string, args ...*program.Operand) *program.Operand {
	return &program.Operand{T: &program.Operand_FnCall{FnCall: &program.FnCall{
		Reference: &program.Reference{Prefix: "lib", Name: name}, Arguments: args}}}
}

// importPkg wraps returns in a package that imports "lib" qualified.
func importPkg(rets ...*program.Operand) *program.Package {
	return &program.Package{Files: []*program.PackageFile{{
		Imports: []*program.Import{{Path: "lib"}},
		Returns: rets,
	}}}
}

// TestDerivThroughQualifiedImport differentiates through a qualified library
// call — the aerospace bracket's map.rotateZX(p, ang) shape. A library function
// is ordinary GSL, so the chain rule must run through its body: splicing it as
// an opaque given (what a Domain implementing OpaqueDomain asks for) would zero
// the tangent of anything reached through it.
func TestDerivThroughQualifiedImport(t *testing.T) {
	// f = lib.twice(x*x) = 2x²  ->  df/dp = (4x, 0, 0)
	direct := nary(program.Operation_DERIV,
		qualCall("twice", mul(ref(program.XVarName), ref(program.XVarName))),
		ref(program.PVarName))

	// A library call used as a REMAP coordinate, like the bracket's rotation:
	// remap(x, lib.swapXY(p)) is y, so its gradient is (0, 1, 0).
	remapped := nary(program.Operation_DERIV,
		nary(program.Operation_REMAP, ref(program.XVarName), qualCall("swapXY", ref(program.PVarName))),
		ref(program.PVarName))

	libDep := libPkg()
	libDep.Path = "lib"
	prog, err := Expand(&program.Program{
		Primary:      importPkg(direct, remapped),
		Dependencies: []*program.Package{libDep},
	})
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	out := prog.Primary
	pt := point{x: 0.6, y: -1.1, z: 0.35}
	for i, want := range [][]float64{{4 * 0.6, 0, 0}, {0, 1, 0}} {
		got := evalVec(t, flattenRemap(out.Files[0].Returns[i]), pt)
		for k := range want {
			if !closeEnough(got[k], want[k]) {
				t.Errorf("return %d component %d: got %v want %v (full %v)", i, k, got[k], want[k], got)
			}
		}
	}
}

// TestQualifiedImportRef resolves a qualified constant (color.darkGray's shape)
// while differentiating: it is a given, so its tangent is zero, but its value
// must still come through.
func TestQualifiedImportRef(t *testing.T) {
	// f = x * lib.half  ->  df/dp = (0.5, 0, 0)
	ret := nary(program.Operation_DERIV,
		mul(ref(program.XVarName), &program.Operand{T: &program.Operand_Reference{
			Reference: &program.Reference{Prefix: "lib", Name: "half"}}}),
		ref(program.PVarName))
	libDep := libPkg()
	libDep.Path = "lib"
	prog, err := Expand(&program.Program{
		Primary:      importPkg(ret),
		Dependencies: []*program.Package{libDep},
	})
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	out := prog.Primary
	got := evalVec(t, flattenRemap(out.Files[0].Returns[0]), point{x: 0.6, y: 0.2})
	for k, want := range []float64{0.5, 0, 0} {
		if !closeEnough(got[k], want) {
			t.Errorf("component %d: got %v want %v (full %v)", k, got[k], want, got)
		}
	}
}
