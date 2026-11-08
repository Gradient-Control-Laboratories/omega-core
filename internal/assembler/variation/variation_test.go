package variation_test

import (
	"testing"

	"gradientcontrol.com/omega/core/internal/assembler/meshplan"
	"gradientcontrol.com/omega/core/internal/assembler/variation"
	parser "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

func TestVariationProgram(t *testing.T) {
	src := `
m = mesh(0);
s = vdbSdf(1, 64, 5);
b = bvhSdf(2);
return m, s, b;
`
	res, err := parser.ParsePackage([]string{src})
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	pkg := res.Package
	pkg.Files[0].Source = src
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
	sub := variation.Program(prog, returns...)

	if len(sub.Primary.Files) != 1 {
		t.Fatalf("got %d files, want the primary's one", len(sub.Primary.Files))
	}
	f := sub.Primary.Files[0]
	if f.Source != src {
		t.Errorf("Source = %q, want %q preserved", f.Source, src)
	}
	if len(sub.Returns()) != 2 || sub.Returns()[0] != returns[0] || sub.Returns()[1] != returns[1] {
		t.Errorf("returns %v, want the two given", sub.Returns())
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
	if got := prog.Returns(); len(got) != 3 {
		t.Errorf("the sliced program's returns changed: %d, want 3", len(got))
	}
}

func TestVariationProgramPrunesDeadDeclarations(t *testing.T) {
	res, err := parser.ParsePackage([]string{`
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
	prog := &program.Program{Primary: res.Package}

	_, subtrees, err := meshplan.SliceProgram(prog, "leaf-")
	if err != nil {
		t.Fatalf("SliceProgram failed: %v", err)
	}
	if len(subtrees) != 2 {
		t.Fatalf("got %d subtrees, want 2", len(subtrees))
	}

	sub := variation.Program(prog, subtrees[0].Return)
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

func TestVariationProgramEmptyReturnsRetainsUIControls(t *testing.T) {
	res, err := parser.ParsePackage([]string{`
r = 10;
r.min = 1;
r.max = 20;
pt = uiPoint([1, 2, 3]);
unused = 42;
m = mesh(0);
return m;
`})
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	pkg := res.Package
	// Simulate property.ExpandShared setting a.Range on r
	for _, a := range pkg.Files[0].Assignments {
		if len(a.Variables) > 0 && a.Variables[0].Name == "r" {
			a.Range = &program.Range{
				Min: program.NewValue(1, nil, nil),
				Max: program.NewValue(20, nil, nil),
			}
		}
	}
	prog := &program.Program{Primary: pkg}

	sub := variation.Program(prog)
	if len(sub.Returns()) != 0 {
		t.Fatalf("got %d returns, want 0", len(sub.Returns()))
	}
	names := map[string]bool{}
	for _, a := range sub.Primary.Files[0].Assignments {
		for _, v := range a.Variables {
			names[v.Name] = true
		}
	}
	if !names["r"] || !names["pt"] {
		t.Errorf("assignments = %v, want UI control roots r and pt retained", names)
	}
	if names["unused"] || names["m"] {
		t.Errorf("assignments = %v, want unused and mesh-typed m pruned", names)
	}
}

func TestVariationPrunePrimary(t *testing.T) {
	src := `
geomFn(r) {
	return length(p) - r;
}
audioFn(t) {
	return sin(t);
}
colorHelper = [1, 0, 0];
matHelper = [[1,0,0,0],[0,1,0,0],[0,0,1,0],[0,0,0,1]];
freq = 440;
s = vdbSdf(1, 64, 5) + geomFn(5);
a = audioFn(freq * time);
return s;
`
	res, err := parser.ParsePackage([]string{src})
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	pkg := res.Package
	pkg.Files[0].Source = src
	// Simulate freq having a Range (slider), and ColorReturns / MaterialReturns referencing colorHelper / matHelper.
	for _, a := range pkg.Files[0].Assignments {
		if len(a.Variables) > 0 && a.Variables[0].Name == "freq" {
			a.Range = &program.Range{
				Min: program.NewValue(100, nil, nil),
				Max: program.NewValue(880, nil, nil),
			}
		}
	}
	pkg.Files[0].ColorReturns = []*program.Operand{program.NewReference("", "colorHelper", nil, nil)}
	pkg.Files[0].MaterialReturns = []*program.Operand{program.NewReference("", "matHelper", nil, nil)}
	prog := &program.Program{Primary: pkg}

	pruned := variation.PrunePrimary(prog)
	f := pruned.Primary.Files[0]
	if f.Source != src {
		t.Errorf("Source = %q, want %q preserved", f.Source, src)
	}
	if len(pruned.ColorReturns()) != 1 || len(pruned.MaterialReturns()) != 1 {
		t.Errorf("ColorReturns=%d, MaterialReturns=%d, want 1 each preserved", len(pruned.ColorReturns()), len(pruned.MaterialReturns()))
	}

	names := map[string]bool{}
	for _, a := range f.Assignments {
		for _, v := range a.Variables {
			names[v.Name] = true
		}
	}
	if !names["s"] || !names["colorHelper"] || !names["matHelper"] || !names["freq"] {
		t.Errorf("assignments = %v, want s, colorHelper, matHelper, and slider freq retained", names)
	}
	if names["a"] {
		t.Errorf("assignments = %v, want audio-only variable 'a' pruned", names)
	}

	fnNames := map[string]bool{}
	for _, ft := range f.FnTypes {
		fnNames[ft.Name] = true
	}
	if !fnNames["geomFn"] || fnNames["audioFn"] {
		t.Errorf("functions = %v, want geomFn retained and audioFn pruned", fnNames)
	}
	if len(pruned.MeshSdfSlots) != 1 || pruned.MeshSdfSlots[0].AssetIndex != 1 {
		t.Errorf("MeshSdfSlots = %v, want slot 1 extracted", pruned.MeshSdfSlots)
	}
}
