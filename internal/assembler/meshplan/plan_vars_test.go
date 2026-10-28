package meshplan_test

import (
	"testing"

	"gradientcontrol.com/omega/core/internal/assembler/meshplan"
	parser "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

func TestSliceProgram_VariableBindingAndReferenceChain(t *testing.T) {
	src := `
m = mesh(0, [10, 20, 30]);
refM = m;
s = sphere(1.0);
return meshUnion(refM, s);
`
	pkg, err := parser.ParsePackage([]string{src})
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	prog := &program.Program{Primary: pkg}

	plan, subtrees, err := meshplan.SliceProgram(prog, "leaf-")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if plan == nil || plan.Root == nil {
		t.Fatalf("expected non-nil plan root")
	}
	if plan.Root.Kind != meshplan.MeshPlanKindBool || plan.Root.Op != meshplan.MeshPlanOpUnion {
		t.Fatalf("expected bool union root, got %+v", plan.Root)
	}
	if plan.Root.Left == nil || plan.Root.Left.Kind != meshplan.MeshPlanKindAsset || plan.Root.Left.AssetIndex != 0 {
		t.Fatalf("expected resolved left child to be asset index 0, got %+v", plan.Root.Left)
	}
	if plan.Root.Right == nil || plan.Root.Right.Kind != meshplan.MeshPlanKindSdf {
		t.Fatalf("expected right child to be sdf, got %+v", plan.Root.Right)
	}

	if len(subtrees) != 1 {
		t.Fatalf("expected 1 subtree, got %d", len(subtrees))
	}
	// Check top-level assignments in synthetic subtree: m and refM must be pruned; s must remain!
	assignments := meshplan.SubtreeProgram(prog, subtrees[0].Return).Primary.Files[0].Assignments
	names := make(map[string]bool)
	for _, a := range assignments {
		if len(a.Variables) > 0 && a.Variables[0] != nil {
			names[a.Variables[0].Name] = true
		}
	}
	if names["m"] || names["refM"] {
		t.Fatalf("expected mesh-typed assignments m and refM to be pruned, got %v", names)
	}
	if !names["s"] {
		t.Fatalf("expected pure SDF assignment s to be kept, got %v", names)
	}
}
