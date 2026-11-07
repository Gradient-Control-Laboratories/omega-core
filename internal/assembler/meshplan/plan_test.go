package meshplan_test

import (
	"encoding/json"
	"reflect"
	"testing"

	"gradientcontrol.com/omega/core/internal/assembler/meshplan"
	"gradientcontrol.com/omega/core/internal/assembler/variation"
	parser "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

func TestMeshExportPlan_JSONRoundTrip(t *testing.T) {
	plan := &meshplan.MeshExportPlan{
		Root: &meshplan.MeshPlanNode{
			Kind: meshplan.MeshPlanKindBool,
			Op:   meshplan.MeshPlanOpUnion,
			Left: &meshplan.MeshPlanNode{
				Kind:       meshplan.MeshPlanKindAsset,
				AssetIndex: 0,
				Transform: &meshplan.MeshTransform{
					Matrix: []float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 10, 20, 30, 1},
				},
			},
			Right: &meshplan.MeshPlanNode{
				Kind:      meshplan.MeshPlanKindSdf,
				SubtreeID: "test-subtree-1",
				StlUrl:    "/downloads/test-subtree-1.stl",
			},
		},
	}

	data, err := json.Marshal(plan)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded meshplan.MeshExportPlan
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if !reflect.DeepEqual(plan, &decoded) {
		t.Fatalf("round-trip mismatch:\nwant: %+v\ngot:  %+v", plan, &decoded)
	}

	// Verify AssetIndex=0 is preserved in JSON (no omitempty dropping index 0).
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("unmarshal raw error: %v", err)
	}
	rootMap, ok := raw["root"].(map[string]interface{})
	if !ok {
		t.Fatalf("root is not map: %v", raw["root"])
	}
	leftMap, ok := rootMap["left"].(map[string]interface{})
	if !ok {
		t.Fatalf("left is not map: %v", rootMap["left"])
	}
	assetIdx, exists := leftMap["assetIndex"]
	if !exists {
		t.Fatalf("expected assetIndex to exist in JSON for index 0")
	}
	if assetIdx != float64(0) {
		t.Fatalf("expected assetIndex to be 0, got %v", assetIdx)
	}
}

func TestSdfSubtree_Model(t *testing.T) {
	ret := &program.Operand{}
	subtree := meshplan.SdfSubtree{
		ID:     "sub-123",
		Return: ret,
	}
	if subtree.ID != "sub-123" || subtree.Return != ret {
		t.Fatalf("unexpected SdfSubtree fields: %+v", subtree)
	}
}

func TestSliceProgram_Empty(t *testing.T) {
	// Nil program
	plan, subtrees, err := meshplan.SliceProgram(nil, "test::")
	if err != nil || plan == nil || plan.Root != nil || len(subtrees) != 0 {
		t.Fatalf("expected empty plan for nil program, got plan=%v subtrees=%d err=%v", plan, len(subtrees), err)
	}

	// Program with empty files
	prog := &program.Program{
		Primary: &program.Package{
			Files: []*program.PackageFile{},
		},
	}
	plan, subtrees, err = meshplan.SliceProgram(prog, "test::")
	if err != nil || plan == nil || plan.Root != nil || len(subtrees) != 0 {
		t.Fatalf("expected empty plan for empty package, got plan=%v subtrees=%d err=%v", plan, len(subtrees), err)
	}
}

func TestSliceProgram_PureSDF(t *testing.T) {
	res, err := parser.ParsePackage([]string{"return sphere(1.0);"})
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	prog := &program.Program{
		Primary:      res.Package,
		Dependencies: []*program.Package{{VersionMajor: 1}},
	}

	plan, subtrees, err := meshplan.SliceProgram(prog, "leaf-")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan == nil || plan.Root == nil {
		t.Fatalf("expected non-nil plan root")
	}
	if plan.Root.Kind != meshplan.MeshPlanKindSdf || plan.Root.SubtreeID != "leaf-0" {
		t.Fatalf("expected sdf root with id leaf-0, got %+v", plan.Root)
	}
	if len(subtrees) != 1 {
		t.Fatalf("expected 1 subtree, got %d", len(subtrees))
	}
	if subtrees[0].ID != "leaf-0" {
		t.Fatalf("expected subtree id leaf-0, got %s", subtrees[0].ID)
	}
	if subtrees[0].Return == nil || len(variation.Program(prog, subtrees[0].Return).Dependencies) != 1 {
		t.Fatalf("expected subtree program to carry parent dependencies")
	}
}

func TestSliceProgram_DiscreteMesh(t *testing.T) {
	res, err := parser.ParsePackage([]string{"return mesh(0, [10, 20, 30]);"})
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	prog := &program.Program{Primary: res.Package}

	plan, subtrees, err := meshplan.SliceProgram(prog, "leaf-")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan == nil || plan.Root == nil {
		t.Fatalf("expected non-nil plan root")
	}
	if plan.Root.Kind != meshplan.MeshPlanKindAsset || plan.Root.AssetIndex != 0 {
		t.Fatalf("expected asset root with index 0, got %+v", plan.Root)
	}
	if plan.Root.Transform == nil || len(plan.Root.Transform.Matrix) != 16 {
		t.Fatalf("expected 16-element transform matrix, got %+v", plan.Root.Transform)
	}
	if len(subtrees) != 0 {
		t.Fatalf("expected 0 subtrees for pure mesh, got %d", len(subtrees))
	}
}

func TestSliceProgram_HybridCSG(t *testing.T) {
	res, err := parser.ParsePackage([]string{"return meshUnion(mesh(0), sphere(1.0));"})
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	prog := &program.Program{Primary: res.Package}

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
	if plan.Root.Left == nil || plan.Root.Left.Kind != meshplan.MeshPlanKindAsset {
		t.Fatalf("expected left asset child, got %+v", plan.Root.Left)
	}
	if plan.Root.Right == nil || plan.Root.Right.Kind != meshplan.MeshPlanKindSdf {
		t.Fatalf("expected right sdf child, got %+v", plan.Root.Right)
	}
	if len(subtrees) != 1 {
		t.Fatalf("expected 1 subtree, got %d", len(subtrees))
	}
}

func TestSliceProgram_AdjacentSDFMerge(t *testing.T) {
	res, err := parser.ParsePackage([]string{"return meshUnion(sphere(1.0), cylinder(2.0));"})
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	prog := &program.Program{Primary: res.Package}

	plan, subtrees, err := meshplan.SliceProgram(prog, "leaf-")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if plan == nil || plan.Root == nil {
		t.Fatalf("expected non-nil plan root")
	}
	// The adjacent SDF leaves must be collapsed into a single SdfLeaf!
	if plan.Root.Kind != meshplan.MeshPlanKindSdf {
		t.Fatalf("expected collapsed sdf leaf root, got kind=%s", plan.Root.Kind)
	}
	if len(subtrees) != 1 {
		t.Fatalf("expected 1 merged subtree, got %d", len(subtrees))
	}
	retExpr := subtrees[0].Return.GetExpression()
	if retExpr == nil || retExpr.Operation != program.Operation_MIN {
		t.Fatalf("expected merged return expression to be MIN, got %v", retExpr)
	}
}

func TestUnionPlanNodes(t *testing.T) {
	nodeA := &meshplan.MeshPlanNode{Kind: meshplan.MeshPlanKindAsset, AssetIndex: 1}
	nodeB := &meshplan.MeshPlanNode{Kind: meshplan.MeshPlanKindAsset, AssetIndex: 2}

	// root is nil
	res := meshplan.UnionPlanNodes(nil, nodeA)
	if res != nodeA {
		t.Fatalf("expected nodeA, got %v", res)
	}

	// root is non-nil
	res = meshplan.UnionPlanNodes(nodeA, nodeB)
	if res.Kind != meshplan.MeshPlanKindBool || res.Op != meshplan.MeshPlanOpUnion || res.Left != nodeA || res.Right != nodeB {
		t.Fatalf("expected union of nodeA and nodeB, got %+v", res)
	}
}

func TestIndexSdfLeaves(t *testing.T) {
	plan := &meshplan.MeshPlanNode{
		Kind: meshplan.MeshPlanKindBool,
		Op:   meshplan.MeshPlanOpUnion,
		Left: &meshplan.MeshPlanNode{
			Kind:      meshplan.MeshPlanKindSdf,
			SubtreeID: "leaf-0",
		},
		Right: &meshplan.MeshPlanNode{
			Kind:      meshplan.MeshPlanKindSdf,
			SubtreeID: "leaf-1",
		},
	}

	leaves := make(map[string]*meshplan.MeshPlanNode)
	meshplan.IndexSdfLeaves(plan, leaves)
	if len(leaves) != 2 || leaves["leaf-0"] == nil || leaves["leaf-1"] == nil {
		t.Fatalf("expected 2 indexed leaves, got %v", leaves)
	}
}

func TestMeshTransformFromDecl_And_ApplyMat4Col(t *testing.T) {
	// Identity from empty
	xf := meshplan.MeshTransformFromDecl(nil)
	if len(xf.Matrix) != 16 || xf.Matrix[0] != 1 || xf.Matrix[15] != 1 {
		t.Fatalf("expected 16-element identity matrix, got %v", xf.Matrix)
	}

	// Translation matrix: translate by [10, 20, 30]
	transMat := []float64{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		10, 20, 30, 1,
	}
	pt := meshplan.ApplyMat4Col(transMat, [3]float64{1, 2, 3})
	expected := [3]float64{11, 22, 33}
	if !reflect.DeepEqual(pt, expected) {
		t.Fatalf("expected %v, got %v", expected, pt)
	}
}

func TestMeshExportPlan_MeshDecls(t *testing.T) {
	// Pure SDF plan
	sdfPlan := &meshplan.MeshExportPlan{
		Root: &meshplan.MeshPlanNode{
			Kind:      meshplan.MeshPlanKindSdf,
			SubtreeID: "leaf-0",
		},
	}
	if decls := sdfPlan.MeshDecls(); len(decls) != 0 {
		t.Fatalf("expected 0 decls for pure SDF plan, got %d", len(decls))
	}

	// Hybrid CSG tree with nested asset leaves
	hybridPlan := &meshplan.MeshExportPlan{
		Root: &meshplan.MeshPlanNode{
			Kind: meshplan.MeshPlanKindBool,
			Op:   meshplan.MeshPlanOpUnion,
			Left: &meshplan.MeshPlanNode{
				Kind:       meshplan.MeshPlanKindAsset,
				AssetIndex: 0,
				ShapeIndex: 1,
				Transform:  &meshplan.MeshTransform{Matrix: []float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 5, 0, 0, 1}},
			},
			Right: &meshplan.MeshPlanNode{
				Kind: meshplan.MeshPlanKindBool,
				Op:   meshplan.MeshPlanOpSubtract,
				Left: &meshplan.MeshPlanNode{
					Kind:      meshplan.MeshPlanKindSdf,
					SubtreeID: "leaf-0",
				},
				Right: &meshplan.MeshPlanNode{
					Kind:       meshplan.MeshPlanKindAsset,
					AssetIndex: 1,
					ShapeIndex: 2,
				},
			},
		},
	}

	decls := hybridPlan.MeshDecls()
	if len(decls) != 2 {
		t.Fatalf("expected 2 mesh decls, got %d", len(decls))
	}
	if decls[0].Index != 0 || decls[0].ShapeIndex != 1 || len(decls[0].Matrix) != 16 {
		t.Fatalf("unexpected decl[0]: %+v", decls[0])
	}
	if decls[1].Index != 1 || decls[1].ShapeIndex != 2 || len(decls[1].Matrix) != 0 {
		t.Fatalf("unexpected decl[1]: %+v", decls[1])
	}
}

func TestMeshExportPlan_HasImplicit(t *testing.T) {
	// Asset-only plan
	assetPlan := &meshplan.MeshExportPlan{
		Root: &meshplan.MeshPlanNode{
			Kind:       meshplan.MeshPlanKindAsset,
			AssetIndex: 0,
		},
	}
	if assetPlan.HasImplicit() {
		t.Fatalf("expected false for asset-only plan")
	}

	// Pure SDF plan
	sdfPlan := &meshplan.MeshExportPlan{
		Root: &meshplan.MeshPlanNode{
			Kind:      meshplan.MeshPlanKindSdf,
			SubtreeID: "leaf-0",
		},
	}
	if !sdfPlan.HasImplicit() {
		t.Fatalf("expected true for pure SDF plan")
	}

	// Hybrid tree
	hybridPlan := &meshplan.MeshExportPlan{
		Root: &meshplan.MeshPlanNode{
			Kind: meshplan.MeshPlanKindBool,
			Op:   meshplan.MeshPlanOpUnion,
			Left: &meshplan.MeshPlanNode{
				Kind:       meshplan.MeshPlanKindAsset,
				AssetIndex: 0,
			},
			Right: &meshplan.MeshPlanNode{
				Kind:      meshplan.MeshPlanKindSdf,
				SubtreeID: "leaf-0",
			},
		},
	}
	if !hybridPlan.HasImplicit() {
		t.Fatalf("expected true for hybrid plan")
	}
}
