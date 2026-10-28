package meshplan_test

import (
	"testing"

	"gradientcontrol.com/omega/core/internal/assembler/meshplan"
	parser "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

func sliceSource(t *testing.T, src string) *meshplan.MeshExportPlan {
	t.Helper()
	pkg, err := parser.ParsePackage([]string{src})
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	plan, _, err := meshplan.SliceProgram(&program.Program{Primary: pkg}, "leaf-")
	if err != nil {
		t.Fatalf("SliceProgram failed: %v", err)
	}
	return plan
}

// Each top-level return is a body, named by its variable; an expression
// return is body-N, a repeated name gets a suffix, and a return's own mesh
// booleans stay inside its body.
func TestMeshExportPlan_Bodies(t *testing.T) {
	plan := sliceSource(t, `
screw = sphere(1.0);
nut = box([-1.0, -1.0, -1.0], [1.0, 1.0, 1.0]);
return screw, nut, meshUnion(mesh(0), sphere(2.0)), screw;
`)
	bodies := plan.Bodies()
	want := []struct {
		index int
		name  string
		kind  string
	}{
		{1, "screw", meshplan.MeshPlanKindSdf},
		{2, "nut", meshplan.MeshPlanKindSdf},
		{3, "body-3", meshplan.MeshPlanKindBool},
		{4, "screw-2", meshplan.MeshPlanKindSdf},
	}
	if len(bodies) != len(want) {
		t.Fatalf("got %d bodies, want %d: %+v", len(bodies), len(want), bodies)
	}
	for i, w := range want {
		b := bodies[i]
		if b.ShapeIndex != w.index || b.Name != w.name || b.Root.Kind != w.kind {
			t.Errorf("body %d = {%d %q %s}, want {%d %q %s}", i, b.ShapeIndex, b.Name, b.Root.Kind, w.index, w.name, w.kind)
		}
	}
}

// A single return is one body, whatever its booleans.
func TestMeshExportPlan_BodiesSingleReturn(t *testing.T) {
	plan := sliceSource(t, `return meshUnion(mesh(0), sphere(1.0));`)
	bodies := plan.Bodies()
	if len(bodies) != 1 || bodies[0].Root != plan.Root || bodies[0].Name != "body-1" {
		t.Fatalf("got %+v, want the root as body-1", bodies)
	}
}

// KeepBodies narrows the plan to the kept returns and refuses to keep none.
func TestMeshExportPlan_KeepBodies(t *testing.T) {
	plan := sliceSource(t, `
a = sphere(1.0);
b = sphere(2.0);
c = sphere(3.0);
return a, b, c;
`)
	if plan.KeepBodies(map[int]bool{}) {
		t.Fatalf("keeping no body should fail")
	}
	if len(plan.Bodies()) != 3 {
		t.Fatalf("a failed KeepBodies changed the plan")
	}
	if !plan.KeepBodies(map[int]bool{1: true, 3: true}) {
		t.Fatalf("KeepBodies(1, 3) failed")
	}
	bodies := plan.Bodies()
	if len(bodies) != 2 || bodies[0].Name != "a" || bodies[1].Name != "c" {
		t.Fatalf("got %+v, want a and c", bodies)
	}
	leaves := map[string]*meshplan.MeshPlanNode{}
	meshplan.IndexSdfLeaves(plan.Root, leaves)
	if len(leaves) != 2 || leaves["leaf-1"] != nil {
		t.Fatalf("got leaves %v, want b's leaf-1 gone", leaves)
	}
}
