package meshplan_test

import (
	"testing"

	"gradientcontrol.com/omega/core/internal/assembler/meshplan"
	parser "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

func parseProgram(t *testing.T, src string) *program.Program {
	t.Helper()
	pkg, err := parser.ParsePackage([]string{src})
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	return &program.Program{Primary: pkg}
}

func TestSliceProgram_UnrolledListCollect(t *testing.T) {
	src := `
meshes = [mesh(0, [1, 0, 0]), mesh(1, [2, 0, 0]), mesh(2, [3, 0, 0])];
return meshUnion(meshes);
`
	prog := parseProgram(t, src)
	plan, subtrees, err := meshplan.SliceProgram(prog, "leaf-")
	if err != nil {
		t.Fatalf("unexpected slicing error: %v", err)
	}

	if plan == nil || plan.Root == nil {
		t.Fatalf("expected non-nil plan root")
	}

	// 3 meshes folded left into a chain:
	//   union(union(m0, m1), m2)
	root := plan.Root
	if root.Kind != meshplan.MeshPlanKindBool || root.Op != meshplan.MeshPlanOpUnion {
		t.Fatalf("expected bool union root, got %+v", root)
	}
	if root.Right == nil || root.Right.Kind != meshplan.MeshPlanKindAsset || root.Right.AssetIndex != 2 {
		t.Fatalf("expected right child to be mesh asset 2, got %+v", root.Right)
	}
	left := root.Left
	if left == nil || left.Kind != meshplan.MeshPlanKindBool || left.Op != meshplan.MeshPlanOpUnion {
		t.Fatalf("expected left child to be bool union, got %+v", left)
	}
	if left.Left == nil || left.Left.Kind != meshplan.MeshPlanKindAsset || left.Left.AssetIndex != 0 {
		t.Fatalf("expected left.left to be mesh asset 0, got %+v", left.Left)
	}
	if left.Right == nil || left.Right.Kind != meshplan.MeshPlanKindAsset || left.Right.AssetIndex != 1 {
		t.Fatalf("expected left.right to be mesh asset 1, got %+v", left.Right)
	}

	if len(subtrees) != 0 {
		t.Fatalf("expected 0 subtrees for pure mesh unrolled list, got %d", len(subtrees))
	}
}

func TestSliceProgram_DifferenceWithList(t *testing.T) {
	src := `
base = [mesh(0), sphere(1.0)];
return meshDifference(base, sphere(0.5));
`
	prog := parseProgram(t, src)
	plan, subtrees, err := meshplan.SliceProgram(prog, "leaf-")
	if err != nil {
		t.Fatalf("unexpected slicing error: %v", err)
	}

	if plan == nil || plan.Root == nil {
		t.Fatalf("expected non-nil plan root")
	}

	// Difference over base list [mesh(0), sphere(1.0)] - sphere(0.5):
	// Base is union(mesh(0), sphere(1.0))
	// Then subtract sphere(0.5)
	if plan.Root.Kind != meshplan.MeshPlanKindBool || plan.Root.Op != meshplan.MeshPlanOpSubtract {
		t.Fatalf("expected bool subtract root, got %+v", plan.Root)
	}
	if len(subtrees) == 0 {
		t.Fatalf("expected at least 1 SDF subtree for difference with spheres")
	}
}
