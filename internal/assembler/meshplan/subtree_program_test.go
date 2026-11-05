package meshplan_test

import (
	"testing"

	"gradientcontrol.com/omega/core/internal/assembler/meshplan"
	parser "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

// A leaf names its return; SubtreeProgram rebuilds the program that evaluates
// it over the sliced program: the same assignments with the mesh-typed ones
// pruned, the given returns in place of the program's own, the dependencies
// shared, and the texture slots the WGSL backend reads set. The sliced
// program itself is left as it was.
func TestSubtreeProgram(t *testing.T) {
	pkg, err := parser.ParsePackage([]string{`
m = mesh(0);
s = vdbSdf(1, 64, 5);
b = bvhSdf(2);
return m, s, b;
`})
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	dep := &program.Package{VersionMajor: 1}
	prog := &program.Program{Primary: pkg, Dependencies: []*program.Package{dep}}

	_, subtrees, err := meshplan.SliceProgram(prog, "leaf-")
	if err != nil {
		t.Fatalf("SliceProgram failed: %v", err)
	}
	if len(subtrees) != 2 {
		t.Fatalf("got %d subtrees, want one per analytical return", len(subtrees))
	}
	returns := []*program.Operand{subtrees[0].Return, subtrees[1].Return}
	sub := meshplan.SubtreeProgram(prog, returns...)

	if len(sub.Primary.Files) != 1 {
		t.Fatalf("got %d files, want the primary's one", len(sub.Primary.Files))
	}
	f := sub.Primary.Files[0]
	if len(f.Returns) != 2 || f.Returns[0] != returns[0] || f.Returns[1] != returns[1] {
		t.Errorf("returns %v, want the two given", f.Returns)
	}
	names := map[string]bool{}
	for _, a := range f.Assignments {
		if len(a.Variables) > 0 && a.Variables[0] != nil {
			names[a.Variables[0].Name] = true
		}
	}
	if names["m"] || !names["s"] || !names["b"] {
		t.Errorf("assignments %v: want the mesh-typed m pruned and s, b kept", names)
	}
	if len(sub.Dependencies) != 1 || sub.Dependencies[0].VersionMajor != 1 {
		t.Errorf("dependencies %v, want the program's dependency preserved", sub.Dependencies)
	}
	if len(sub.MeshSdfSlots) != 2 ||
		sub.MeshSdfSlots[0].AssetIndex != 1 ||
		sub.MeshSdfSlots[0].Kind != program.MeshSdfKind_MESH_SDF_KIND_VDB ||
		sub.MeshSdfSlots[0].VoxelDivisor != 64 ||
		sub.MeshSdfSlots[0].HalfBandVoxels != 5 ||
		sub.MeshSdfSlots[1].AssetIndex != 2 ||
		sub.MeshSdfSlots[1].Kind != program.MeshSdfKind_MESH_SDF_KIND_BVH {
		t.Errorf("mesh SDF slots %v, want VDB slot for vdbSdf(1, 64, 5) and BVH slot for bvhSdf(2)", sub.MeshSdfSlots)
	}
	if got := prog.Primary.Files[0].Returns; len(got) != 3 {
		t.Errorf("the sliced program's returns changed: %d, want 3", len(got))
	}
}

func TestSubtreeProgramPrunesDeadDeclarations(t *testing.T) {
	pkg, err := parser.ParsePackage([]string{`
usedFn(r) {
	return length(p) - r;
}
unusedFn(r) {
	return length(p) - r * 2.0;
}
usedSdf = vdbSdf(1, 64, 5) + usedFn(2.0);
unusedSdf = bvhSdf(2) + unusedFn(3.0);
return usedSdf, unusedSdf;
`})
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	prog := &program.Program{Primary: pkg}

	_, subtrees, err := meshplan.SliceProgram(prog, "leaf-")
	if err != nil {
		t.Fatalf("SliceProgram failed: %v", err)
	}
	if len(subtrees) != 2 {
		t.Fatalf("got %d subtrees, want 2", len(subtrees))
	}

	sub := meshplan.SubtreeProgram(prog, subtrees[0].Return)
	f := sub.Primary.Files[0]

	names := map[string]bool{}
	for _, a := range f.Assignments {
		for _, v := range a.Variables {
			if v != nil {
				names[v.Name] = true
			}
		}
	}
	if !names["usedSdf"] || names["unusedSdf"] {
		t.Errorf("assignments = %v, want only usedSdf retained and unusedSdf pruned", names)
	}

	fnNames := map[string]bool{}
	for _, ft := range f.FnTypes {
		if ft != nil {
			fnNames[ft.Name] = true
		}
	}
	if !fnNames["usedFn"] || fnNames["unusedFn"] {
		t.Errorf("functions = %v, want only usedFn retained and unusedFn pruned", fnNames)
	}

	if len(sub.MeshSdfSlots) != 1 || sub.MeshSdfSlots[0].AssetIndex != 1 {
		t.Errorf("MeshSdfSlots = %v, want only slot 1 (from usedSdf) and slot 2 pruned", sub.MeshSdfSlots)
	}
}
