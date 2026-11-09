package assembler_test

import (
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/assembler/meshplan"
	"gradientcontrol.com/omega/core/assembler/variation"
	"gradientcontrol.com/omega/core/program"
)

// fidget has no mesh() builtin -- mesh() is a scene placement, not an SDF -- so
// a module reaching it with one intact fails the whole export or simulation
// with "unsupported builtin: mesh". These check the module assembler.Assemble
// hands over is clean, which is the property that actually broke, rather than
// the intermediate bookkeeping.

func preparedModule(t *testing.T, src string) (*program.Package, *assembler.Result) {
	t.Helper()
	res, err := assembler.Assemble([]string{src}, assembler.ExtractMeshPlan("leaf-"))
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	if len(res.SdfSubtrees) > 0 {
		return variation.Program(res.Program, res.SdfSubtrees[0].Return).Primary, res
	}
	return &program.Package{}, res
}

// mentionsMeshCall looks for a call to mesh() specifically -- not meshUnion,
// not an importMesh path -- since that is the one fidget has no builtin for.
func mentionsMeshCall(pkg *program.Package) bool {
	found := false
	var operand func(op *program.Operand)
	operand = func(op *program.Operand) {
		if op == nil || found {
			return
		}
		switch v := op.T.(type) {
		case *program.Operand_FnCall:
			fc := v.FnCall
			if fc == nil || fc.Reference == nil {
				return
			}
			if fc.Reference.Prefix == "" && fc.Reference.Name == "mesh" {
				found = true
				return
			}
			for _, a := range fc.Arguments {
				operand(a)
			}
		case *program.Operand_Expression:
			if v.Expression != nil {
				for _, o := range v.Expression.Operands {
					operand(o)
				}
			}
		}
	}
	var fnType func(ft *program.FnType)
	fnType = func(ft *program.FnType) {
		if ft == nil {
			return
		}
		for _, g := range ft.Groups {
			for _, fn := range g.Functions {
				for _, a := range fn.Assignments {
					operand(a.Expression)
				}
				for _, sub := range fn.FnTypes {
					fnType(sub)
				}
				for _, r := range fn.Returns {
					operand(r)
				}
			}
		}
	}
	for _, file := range pkg.Files {
		for _, a := range file.Assignments {
			operand(a.Expression)
		}
		for _, ft := range file.FnTypes {
			fnType(ft)
		}
		for _, r := range file.Returns {
			operand(r)
		}
	}
	return found
}

// returnCount totals the package's returns across its files, standing in for
// the flat pkg.Returns the old package shape had.
func returnCount(pkg *program.Package) int {
	n := 0
	for _, file := range pkg.Files {
		n += len(file.Returns)
	}
	return n
}

func TestPreparedModuleCarriesNoBareMeshPlacement(t *testing.T) {
	cases := map[string]string{
		"placed by call":             "importMesh 0 \"p.stl\";\nreturn mesh(0, [1,2,3]);\n",
		"placed by name":             "importMesh 0 \"p.stl\";\nmeshOut = mesh(0, [1,2,3]);\nreturn meshOut;\n",
		"placed by name, second":     "importMesh 0 \"p.stl\";\nmeshOut = mesh(0, [1,2,3]);\nreturn 1, meshOut;\n",
		"placed by name, unused too": "importMesh 0 \"p.stl\";\nmeshOut = mesh(0, [1,2,3]);\nspare = mesh(0, [9,9,9]);\nreturn meshOut;\n",
	}
	for name, src := range cases {
		pkg, res := preparedModule(t, src)
		if mentionsMeshCall(pkg) {
			t.Errorf("%s: the prepared module still carries a mesh() call; fidget will reject it", name)
		}
		if res.MeshPlan == nil || res.MeshPlan.Root == nil {
			t.Errorf("%s: expected mesh plan root", name)
		}
	}
}

// The other half: a mesh composed into a boolean must be preserved in the mesh plan
// for Hodge / Manifold CSG evaluation.
func TestPreparedModuleKeepsComposedMesh(t *testing.T) {
	_, res := preparedModule(t, `importMesh 0 "p.stl";
return meshUnion(mesh(0, [1,2,3]), mesh(0, [4,5,6]));
`)
	if res.MeshPlan == nil || res.MeshPlan.Root == nil {
		t.Fatal("expected mesh plan for composed meshUnion")
	}
	if res.MeshPlan.Root.Kind != meshplan.MeshPlanKindBool || res.MeshPlan.Root.Op != meshplan.MeshPlanOpUnion {
		t.Errorf("expected boolean union root in mesh plan, got %+v", res.MeshPlan.Root)
	}
}

// An implicit return alongside a placement keeps its own slot, so shape indices
// and the SDF the simulation runs on both stay intact.
func TestPreparedModuleKeepsImplicitReturns(t *testing.T) {
	pkg, res := preparedModule(t, `importMesh 0 "p.stl";
meshOut = mesh(0, [1,2,3]);
solid = 1;
return solid, meshOut;
`)
	if returnCount(pkg) != 1 {
		t.Fatalf("expected the implicit return to remain alone, got %d", returnCount(pkg))
	}
	if mentionsMeshCall(pkg) {
		t.Error("the placement was not lifted out")
	}
	if res.MeshPlan == nil || res.MeshPlan.Root == nil {
		t.Error("expected mesh plan to capture placed mesh")
	}
}

