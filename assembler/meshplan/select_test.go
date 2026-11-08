package meshplan_test

import (
	"encoding/json"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/assembler/meshplan"
	parser "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

func TestSelectPrimaryAndBodySdfProgram(t *testing.T) {
	res, err := parser.ParsePackage([]string{`
ball = sphere(1.0);
cube = box([-1.0, -1.0, -1.0], [1.0, 1.0, 1.0]);
return mesh(0), ball, cube;
`})
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	prog := &program.Program{Primary: res.Package}
	plan, subtrees, err := meshplan.SliceProgram(prog, "leaf-")
	if err != nil {
		t.Fatalf("SliceProgram failed: %v", err)
	}

	primaryProg, decls := meshplan.SelectPrimarySdfProgram(prog, plan, subtrees)
	if len(decls) != 1 || decls[0].Index != 0 {
		t.Fatalf("SelectPrimarySdfProgram decls = %+v, want 1 decl at index 0", decls)
	}
	primaryRets := primaryProg.Primary.Files[0].Returns
	if len(primaryRets) != 1 || primaryRets[0].GetReference().GetName() != "ball" {
		t.Fatalf("SelectPrimarySdfProgram returns = %v, want [ball]", primaryRets)
	}

	for body, want := range map[int]string{2: "ball", 3: "cube"} {
		subProg, err := meshplan.SelectBodySdfProgram(prog, plan, subtrees, body)
		if err != nil {
			t.Fatalf("body %d: %v", body, err)
		}
		returns := subProg.Primary.Files[0].Returns
		if len(returns) != 1 || returns[0].GetReference().GetName() != want {
			t.Errorf("body %d returns %v, want %s alone", body, returns, want)
		}
	}
	if _, err := meshplan.SelectBodySdfProgram(prog, plan, subtrees, 1); err == nil || !strings.Contains(err.Error(), "meshes") {
		t.Errorf("mesh body: err %v, want a refusal naming meshes", err)
	}
	if _, err := meshplan.SelectBodySdfProgram(prog, plan, subtrees, 9); err == nil {
		t.Errorf("missing body: want an error")
	}
}

func TestSdfSubtreeJSONAndLeafPrograms(t *testing.T) {
	res, err := parser.ParsePackage([]string{`
ball = bvhSdf(2);
return meshUnion(mesh(0), ball);
`})
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	prog := &program.Program{Primary: res.Package}
	_, subtrees, err := meshplan.SliceProgram(prog, "leaf-")
	if err != nil {
		t.Fatalf("SliceProgram failed: %v", err)
	}

	encoded, err := json.Marshal(subtrees)
	if err != nil {
		t.Fatalf("json.Marshal(subtrees) failed: %v", err)
	}
	var decoded []meshplan.SdfSubtree
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("json.Unmarshal(subtrees) failed: %v", err)
	}
	if len(decoded) != 1 || decoded[0].ID != subtrees[0].ID || decoded[0].Return.GetReference().GetName() != "ball" {
		t.Fatalf("decoded subtrees = %+v, want 1 subtree returning ball", decoded)
	}

	leaves, err := meshplan.LeafPrograms(prog, decoded)
	if err != nil {
		t.Fatalf("LeafPrograms failed: %v", err)
	}
	if len(leaves) != 1 || leaves[0].ID != subtrees[0].ID {
		t.Fatalf("leaves = %+v, want 1 leaf matching %q", leaves, subtrees[0].ID)
	}
	leafProg, err := program.DecodeProgram(leaves[0].Bytes)
	if err != nil {
		t.Fatalf("DecodeProgram(leaf.Bytes) failed: %v", err)
	}
	if len(leafProg.MeshSdfSlots) != 1 || leafProg.MeshSdfSlots[0].AssetIndex != 2 {
		t.Errorf("leafProg.MeshSdfSlots = %+v, want 1 BVH slot for asset 2", leafProg.MeshSdfSlots)
	}
}
