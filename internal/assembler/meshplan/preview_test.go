package meshplan_test

import (
	"reflect"
	"testing"

	"gradientcontrol.com/omega/core/internal/assembler/meshplan"
	parser "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

func TestPreviewBodiesMapsReturnsToSlotsAndMeshes(t *testing.T) {
	res := parser.ParsePackage([]string{`
ball = sphere(1.0);
return mesh(0), ball, meshUnion(mesh(0), sphere(2.0));
`})
	if err := res.Error(); err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	plan, subtrees, err := meshplan.SliceProgram(&program.Program{Primary: res.Package}, "leaf-")
	if err != nil {
		t.Fatalf("SliceProgram failed: %v", err)
	}
	got := meshplan.PreviewBodies(plan, subtrees)
	want := []meshplan.PreviewBody{
		{Name: "mesh", Mesh: true},
		{Name: "ball", Slots: []int{1}},
		{Name: "meshUnion", Slots: []int{2}, Mesh: true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestAdditiveSdfSubtreesAndPreviewProgram(t *testing.T) {
	t.Run("drops subtracted and doubly subtracted SDF leaves", func(t *testing.T) {
		res := parser.ParsePackage([]string{`
solid = box([-2.0, -2.0, -2.0], [2.0, 2.0, 2.0]);
cavity = sphere(1.0);
plug = box([-3.0, -0.5, -0.5], [3.0, 0.5, 0.5]);
return meshDifference([mesh(0), solid], meshDifference([mesh(0), cavity], plug));
`})
		if err := res.Error(); err != nil {
			t.Fatalf("ParsePackage failed: %v", err)
		}
		prog := &program.Program{Primary: res.Package}
		plan, subtrees, err := meshplan.SliceProgram(prog, "leaf-")
		if err != nil {
			t.Fatalf("SliceProgram failed: %v", err)
		}
		additive := meshplan.AdditiveSdfSubtrees(plan, subtrees)
		if len(additive) != 1 || additive[0].Return.GetFnCall().GetReference().GetName() != "box" {
			t.Fatalf("AdditiveSdfSubtrees = %+v, want [box] only", additive)
		}

		previewProg := meshplan.PreviewProgram(prog, plan, subtrees)
		rets := previewProg.Primary.Files[0].Returns
		if len(rets) != 1 || rets[0].GetFnCall().GetReference().GetName() != "box" {
			t.Fatalf("PreviewProgram returns = %v, want [box] only", rets)
		}
	})

	t.Run("preserves original program when no meshes are present", func(t *testing.T) {
		res := parser.ParsePackage([]string{`
red = sphere(1.0);
blue = sphere(2.0);
return red, blue;
`})
		if err := res.Error(); err != nil {
			t.Fatalf("ParsePackage failed: %v", err)
		}
		prog := &program.Program{Primary: res.Package}
		plan, subtrees, err := meshplan.SliceProgram(prog, "leaf-")
		if err != nil {
			t.Fatalf("SliceProgram failed: %v", err)
		}
		previewProg := meshplan.PreviewProgram(prog, plan, subtrees)
		if previewProg != prog {
			t.Fatalf("expected PreviewProgram to return prog unchanged when no meshes exist")
		}
	})

	t.Run("prunes subtracted SDF declarations while preserving slider and UI control assignments", func(t *testing.T) {
		res := parser.ParsePackage([]string{`
uiTransform(m) {
	return m;
}
radius = 2.0;
radius.min = 1.0;
radius.max = 5.0;
xf = uiTransform([[1, 0, 0, 0], [0, 1, 0, 0], [0, 0, 1, 0], [4, 5, 6, 1]]);
cavity = length(p) - radius + noiseHash3(p)[0];
solid = length(p) - 2.0;
return meshDifference([mesh(0, xf), solid], cavity);
`})
		if err := res.Error(); err != nil {
			t.Fatalf("ParsePackage failed: %v", err)
		}
		pkg := res.Package
		// Synthesize Range on radius as lower.SynthesizeRanges does in the assembler.
		for _, a := range pkg.Files[0].Assignments {
			if len(a.Variables) == 1 && a.Variables[0].Name == "radius" {
				a.Range = &program.Range{
					Min: &program.Operand{T: &program.Operand_Value{Value: 1.0}},
					Max: &program.Operand{T: &program.Operand_Value{Value: 5.0}},
				}
			}
		}
		prog := &program.Program{Primary: pkg}
		plan, subtrees, err := meshplan.SliceProgram(prog, "leaf-")
		if err != nil {
			t.Fatalf("SliceProgram failed: %v", err)
		}

		previewProg := meshplan.PreviewProgram(prog, plan, subtrees)
		f := previewProg.Primary.Files[0]
		names := map[string]bool{}
		for _, a := range f.Assignments {
			for _, v := range a.Variables {
				if v != nil {
					names[v.Name] = true
				}
			}
		}
		if !names["radius"] || !names["xf"] || names["cavity"] {
			t.Fatalf("PreviewProgram assignments = %v, want radius and xf kept and cavity pruned", names)
		}
		if len(f.Returns) != 1 {
			t.Fatalf("PreviewProgram returns = %d, want 1 (solid)", len(f.Returns))
		}
		if previewProg.GlobalTextures != nil && previewProg.GlobalTextures.UsesNoiseTile {
			t.Fatalf("expected subtracted cavity's noiseHash3 to be pruned from GlobalTextures")
		}
	})
}

func TestSubtreesInPlan(t *testing.T) {
	res := parser.ParsePackage([]string{`
ball = sphere(1.0);
cube = box([-1.0, -1.0, -1.0], [1.0, 1.0, 1.0]);
return mesh(0), ball, cube;
`})
	if err := res.Error(); err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	plan, subtrees, err := meshplan.SliceProgram(&program.Program{Primary: res.Package}, "leaf-")
	if err != nil {
		t.Fatalf("SliceProgram failed: %v", err)
	}
	if !plan.KeepBodies(map[int]bool{3: true}) {
		t.Fatal("KeepBodies(3) returned false")
	}
	kept := meshplan.SubtreesInPlan(plan, subtrees)
	if len(kept) != 1 || kept[0].Return.GetReference().GetName() != "cube" {
		t.Fatalf("SubtreesInPlan = %+v, want [cube] only", kept)
	}
}
