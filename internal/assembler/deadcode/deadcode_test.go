package deadcode_test

import (
	"testing"

	"gradientcontrol.com/omega/core/assembler/source"
	"gradientcontrol.com/omega/core/internal/assembler/deadcode"
	"gradientcontrol.com/omega/core/internal/collect"
	"gradientcontrol.com/omega/core/internal/eval"
	"gradientcontrol.com/omega/core/internal/link"
	internalsource "gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/types"
	parserlib "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

func assembleUnpruned(primarySources []string, srcMgr source.Manager) (*program.Program, error) {
	prog, err := collect.Program(primarySources, srcMgr, parserlib.ParsePackage)
	if err != nil {
		return nil, err
	}
	err = link.References(prog, srcMgr, parserlib.ParsePackage)
	if err != nil {
		return nil, err
	}
	err = types.Analyze(prog)
	if err != nil {
		return nil, err
	}
	_, err = eval.Fold(prog, nil)
	if err != nil {
		return nil, err
	}
	return prog, nil
}

func TestRemoveDeadCodeSingle(t *testing.T) {
	src := `
		y = 10.0;
		x = 42.0;
		unused() {
			return 100.0;
		}
		used() {
			return 200.0;
		}
		return x + used();
	`

	pkgMgr := &internalsource.MockManager{}

	// Assemble compiles, links, and returns the unpruned program
	prog, err := assembleUnpruned([]string{src}, pkgMgr)
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}

	// Verify that y is present before Remove
	hasY := false
	for _, file := range prog.Primary.Files {
		for _, assign := range file.Assignments {
			for _, v := range assign.Variables {
				if v.Name == "y" {
					hasY = true
				}
			}
		}
	}
	if !hasY {
		t.Fatal("expected 'y' to be present before dead code removal")
	}

	// Run dead code removal
	pruned, err := deadcode.Remove(prog)
	if err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	// Check that y is removed, and x is kept
	hasYAfter := false
	hasXAfter := false
	for _, file := range pruned.Primary.Files {
		for _, assign := range file.Assignments {
			for _, v := range assign.Variables {
				if v.Name == "y" {
					hasYAfter = true
				}
				if v.Name == "x" {
					hasXAfter = true
				}
			}
		}
	}

	if hasYAfter {
		t.Error("expected assignment 'y' to be removed")
	}
	if !hasXAfter {
		t.Error("expected assignment 'x' to be kept")
	}

	// Check that unused() is removed, and used() is kept
	hasUnused := false
	hasUsed := false
	for _, file := range pruned.Primary.Files {
		for _, fnType := range file.FnTypes {
			if fnType.Name == "unused" {
				hasUnused = true
			}
			if fnType.Name == "used" {
				hasUsed = true
			}
		}
	}

	if hasUnused {
		t.Error("expected function 'unused' to be removed")
	}
	if !hasUsed {
		t.Error("expected function 'used' to be kept")
	}
}

func TestDeadcode_PreservesColorAndMaterialReturnsAndRoots(t *testing.T) {
	loc := &program.Location{Line: 1, Column: 1}

	assignColor := &program.Assignment{
		Variables:  []*program.Variable{{Name: "usedByColor"}},
		Expression: program.NewValue(42.0, loc, loc),
	}
	assignMat := &program.Assignment{
		Variables:  []*program.Variable{{Name: "usedByMat"}},
		Expression: program.NewValue(84.0, loc, loc),
	}
	assignDead := &program.Assignment{
		Variables:  []*program.Variable{{Name: "deadVar"}},
		Expression: program.NewValue(999.0, loc, loc),
	}

	file := &program.PackageFile{
		Path:            "main.gsl",
		Assignments:     []*program.Assignment{assignColor, assignMat, assignDead},
		Returns:         []*program.Operand{program.NewValue(1.0, loc, loc)},
		ColorReturns:    []*program.Operand{program.NewReference("", "usedByColor", loc, loc)},
		MaterialReturns: []*program.Operand{program.NewReference("", "usedByMat", loc, loc)},
	}
	prog := &program.Program{
		Primary: &program.Package{
			Path:  "main",
			Files: []*program.PackageFile{file},
		},
	}

	if err := link.References(prog, nil, nil); err != nil {
		t.Fatalf("link.References failed: %v", err)
	}

	cleaned, err := deadcode.Remove(prog)
	if err != nil {
		t.Fatalf("deadcode.Remove failed: %v", err)
	}

	cleanedFile := cleaned.Primary.Files[0]
	assignNames := make(map[string]bool)
	for _, a := range cleanedFile.Assignments {
		for _, v := range a.Variables {
			assignNames[v.Name] = true
		}
	}
	if !assignNames["usedByColor"] {
		t.Error("expected usedByColor to be preserved")
	}
	if !assignNames["usedByMat"] {
		t.Error("expected usedByMat to be preserved")
	}
	if assignNames["deadVar"] {
		t.Error("expected deadVar to be removed")
	}
	if len(cleanedFile.ColorReturns) != 1 {
		t.Errorf("expected 1 ColorReturn, got %d", len(cleanedFile.ColorReturns))
	}
	if len(cleanedFile.MaterialReturns) != 1 {
		t.Errorf("expected 1 MaterialReturn, got %d", len(cleanedFile.MaterialReturns))
	}
}
