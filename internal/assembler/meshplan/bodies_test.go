package meshplan_test

import (
	"testing"

	"gradientcontrol.com/omega/core/internal/assembler/meshplan"
	parser "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

func sliceSource(t *testing.T, src string) *meshplan.MeshExportPlan {
	t.Helper()
	res, err := parser.ParsePackage([]string{src})
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	plan, _, err := meshplan.SliceProgram(&program.Program{Primary: res.Package}, "leaf-")
	if err != nil {
		t.Fatalf("SliceProgram failed: %v", err)
	}
	return plan
}

// Each top-level return is a body, named by its variable or the function it
// calls (program.OperandName), a repeated name gets a suffix, and a return's
// own mesh booleans stay inside its body.
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
		{3, "meshUnion", meshplan.MeshPlanKindBool},
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
	if len(bodies) != 1 || bodies[0].Root != plan.Root || bodies[0].Name != "meshUnion" {
		t.Fatalf("got %+v, want the root as meshUnion", bodies)
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

// With the document's source, a return that names nothing is named by its
// text, as the Bodies panel names it, and repeats of any name get a suffix.
func TestSliceDocument_NamesExpressionsByText(t *testing.T) {
	src := `
a = sphere(1.0);
b = sphere(2.0);
return min(a, b), a, min(a, b);
`
	res, err := parser.ParsePackage([]string{src})
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	plan, _, err := meshplan.SliceDocument(&program.Program{Primary: res.Package}, "leaf-", src)
	if err != nil {
		t.Fatalf("SliceDocument failed: %v", err)
	}
	var got []string
	for _, b := range plan.Bodies() {
		got = append(got, b.Name)
	}
	want := []string{"min(a, b)", "a", "min(a, b)-2"}
	if len(got) != len(want) {
		t.Fatalf("got %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got %q, want %q", got, want)
		}
	}
}
