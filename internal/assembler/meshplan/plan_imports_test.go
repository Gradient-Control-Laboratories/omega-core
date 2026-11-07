package meshplan_test

import (
	"math"
	"testing"

	"gradientcontrol.com/omega/core/internal/assembler/meshplan"
	parser "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

func TestSliceProgram_ImportedConstantsInMeshArgs(t *testing.T) {
	importedRes := parser.ParsePackage([]string{"myConst = 42.0;"})
	if err := importedRes.Error(); err != nil {
		t.Fatalf("ParsePackage imported failed: %v", err)
	}
	importedPkg := importedRes.Package
	importedPkg.Path = "mathMod"

	primaryRes := parser.ParsePackage([]string{`
import "mathMod";
return mesh(0, [mathMod.myConst, 0, 0]);
`})
	if err := primaryRes.Error(); err != nil {
		t.Fatalf("ParsePackage primary failed: %v", err)
	}
	primaryPkg := primaryRes.Package

	prog := &program.Program{
		Primary:      primaryPkg,
		Dependencies: []*program.Package{importedPkg},
	}

	plan, _, err := meshplan.SliceProgram(prog, "leaf-")
	if err != nil {
		t.Fatalf("unexpected slicing error: %v", err)
	}

	if plan == nil || plan.Root == nil {
		t.Fatalf("expected non-nil plan root")
	}
	if plan.Root.Kind != meshplan.MeshPlanKindAsset || plan.Root.AssetIndex != 0 {
		t.Fatalf("expected asset index 0, got %+v", plan.Root)
	}
	if plan.Root.Transform == nil || len(plan.Root.Transform.Matrix) != 16 {
		t.Fatalf("expected 16-element transform matrix, got %+v", plan.Root.Transform)
	}
	// Check translation x component is folded to 42.0
	xTrans := plan.Root.Transform.Matrix[12]
	if math.Abs(xTrans-42.0) > 1e-6 {
		t.Fatalf("expected x translation 42.0, got %f", xTrans)
	}
}
