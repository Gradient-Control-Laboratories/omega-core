package program_test

import (
	"testing"

	"gradientcontrol.com/omega/core/internal/assembler/deadcode"
	"gradientcontrol.com/omega/core/internal/assembler/warpfold"
	"gradientcontrol.com/omega/core/internal/eval"
	"gradientcontrol.com/omega/core/internal/link"
	"gradientcontrol.com/omega/core/internal/meshfold"
	"gradientcontrol.com/omega/core/internal/refs"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/types"
	"gradientcontrol.com/omega/core/program"
)

func TestAutoImport_ColorAndMaterialReturns(t *testing.T) {
	loc := &program.Location{Line: 1, Column: 1}
	colorCall := program.NewFnCall("color", "rgb", nil, loc, loc)
	matCall := program.NewFnCall("material", "chrome", nil, loc, loc)

	pkg := &program.Package{
		Path: "main",
		Files: []*program.PackageFile{{
			Path:            "main.gsl",
			Returns:         []*program.Operand{program.NewValue(1.0, loc, loc)},
			ColorReturns:    []*program.Operand{colorCall},
			MaterialReturns: []*program.Operand{matCall},
		}},
	}

	program.UpdateImports(pkg, "main")

	var importedPaths []string
	for _, imp := range pkg.Files[0].Imports {
		importedPaths = append(importedPaths, imp.Path)
	}

	assert.Contains(t, importedPaths, "color")
	assert.Contains(t, importedPaths, "material")
}

func TestClone_ColorAndMaterialReturns(t *testing.T) {
	loc := &program.Location{Line: 1, Column: 1}
	cOp := program.NewValue(0.5, loc, loc)
	mOp := program.NewValue(1.5, loc, loc)

	file := &program.PackageFile{
		Path:            "test.gsl",
		Returns:         []*program.Operand{program.NewValue(1.0, loc, loc)},
		ColorReturns:    []*program.Operand{cOp},
		MaterialReturns: []*program.Operand{mOp},
	}
	pkg := &program.Package{
		Path:  "pkg",
		Files: []*program.PackageFile{file},
	}

	clonedPkg := program.ClonePackage(pkg)
	assert.Equal(t, 1, len(clonedPkg.Files))
	clonedFile := clonedPkg.Files[0]

	assert.Equal(t, 1, len(clonedFile.ColorReturns))
	assert.Equal(t, 1, len(clonedFile.MaterialReturns))
	assert.True(t, clonedFile.ColorReturns[0] != cOp)
	assert.True(t, clonedFile.MaterialReturns[0] != mOp)
	assert.Equal(t, 0.5, clonedFile.ColorReturns[0].GetValue())
	assert.Equal(t, 1.5, clonedFile.MaterialReturns[0].GetValue())
}

func TestDeadcode_PreservesColorAndMaterialReturnsAndRoots(t *testing.T) {
	loc := &program.Location{Line: 1, Column: 1}

	// var usedByColor = 42.0
	// var usedByMat = 84.0
	// var deadVar = 999.0
	// return 1.0 (neither uses deadVar, usedByColor, nor usedByMat)
	// colorReturns: [usedByColor]
	// materialReturns: [usedByMat]
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

	retOp := program.NewValue(1.0, loc, loc)
	colorOp := program.NewReference("", "usedByColor", loc, loc)
	matOp := program.NewReference("", "usedByMat", loc, loc)

	file := &program.PackageFile{
		Path:            "main.gsl",
		Assignments:     []*program.Assignment{assignColor, assignMat, assignDead},
		Returns:         []*program.Operand{retOp},
		ColorReturns:    []*program.Operand{colorOp},
		MaterialReturns: []*program.Operand{matOp},
	}
	prog := &program.Program{
		Primary: &program.Package{
			Path:  "main",
			Files: []*program.PackageFile{file},
		},
	}

	err := link.References(prog, nil, nil)
	assert.NoError(t, err)

	cleaned, err := deadcode.Remove(prog)
	assert.NoError(t, err)

	cleanedFile := cleaned.Primary.Files[0]
	// deadVar must be removed, but usedByColor and usedByMat must be preserved
	var assignNames []string
	for _, a := range cleanedFile.Assignments {
		for _, v := range a.Variables {
			assignNames = append(assignNames, v.Name)
		}
	}
	assert.Contains(t, assignNames, "usedByColor")
	assert.Contains(t, assignNames, "usedByMat")
	assert.False(t, containsString(assignNames, "deadVar"))

	// Verify ColorReturns and MaterialReturns are preserved on cleanedFile
	assert.Equal(t, 1, len(cleanedFile.ColorReturns))
	assert.Equal(t, 1, len(cleanedFile.MaterialReturns))
}

func TestEvalFold_ColorAndMaterialReturns(t *testing.T) {
	loc := &program.Location{Line: 1, Column: 1}
	colorExpr := program.NewExpression(
		program.Operation_ADD,
		[]*program.Operand{
			program.NewValue(1.5, loc, loc),
			program.NewValue(2.5, loc, loc),
		},
		loc,
		loc,
	)
	matExpr := program.NewExpression(
		program.Operation_MUL,
		[]*program.Operand{
			program.NewValue(3.0, loc, loc),
			program.NewValue(4.0, loc, loc),
		},
		loc,
		loc,
	)

	prog := &program.Program{
		Primary: &program.Package{
			Path: "main",
			Files: []*program.PackageFile{{
				Path:            "main.gsl",
				Returns:         []*program.Operand{program.NewValue(0.0, loc, loc)},
				ColorReturns:    []*program.Operand{colorExpr},
				MaterialReturns: []*program.Operand{matExpr},
			}},
		},
		Dependencies: []*program.Package{{
			Path: "builtin",
		}},
	}

	err := link.References(prog, nil, nil)
	assert.NoError(t, err)
	err = types.Analyze(prog)
	assert.NoError(t, err)

	mod, err := eval.Fold(prog, nil)
	assert.NoError(t, err)
	assert.True(t, mod)

	file := prog.Primary.Files[0]
	assert.Equal(t, []float64{4.0}, file.ColorReturns[0].ConstValues)
	assert.Equal(t, []float64{12.0}, file.MaterialReturns[0].ConstValues)
}

func TestMeshfoldAndRefs_ColorAndMaterialReturns(t *testing.T) {
	loc := &program.Location{Line: 1, Column: 1}
	stepCall := program.NewFnCall(
		"",
		"importStep",
		[]*program.Operand{program.NewValue(float64('p'), loc, loc)},
		loc,
		loc,
	)

	meshCall := program.NewFnCall(
		"",
		"mesh",
		[]*program.Operand{program.NewValue(3.0, loc, loc)},
		loc,
		loc,
	)

	file := &program.PackageFile{
		Path:            "main.gsl",
		Returns:         []*program.Operand{program.NewValue(0.0, loc, loc)},
		ColorReturns:    []*program.Operand{stepCall},
		MaterialReturns: []*program.Operand{meshCall},
	}

	prog := &program.Program{
		Primary: &program.Package{
			Path:  "main",
			Files: []*program.PackageFile{file},
		},
		Dependencies: []*program.Package{{
			Path: "builtin",
		}},
	}

	meshRefs := refs.CollectMeshReferences(prog.Primary)
	assert.Equal(t, 1, len(meshRefs))
	assert.Equal(t, 3, meshRefs[0].Index)

	folded, slots, err := meshfold.FoldProgramWithSlots(prog)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(slots))
	// The importStep call in ColorReturns should be folded to a numeric slot value
	val, ok := folded.Primary.Files[0].ColorReturns[0].T.(*program.Operand_Value)
	assert.True(t, ok)
	assert.Equal(t, float64(slots[0].Slot), val.Value)
}

func TestWarpfold_ColorAndMaterialReturns(t *testing.T) {
	loc := &program.Location{Line: 1, Column: 1}
	warpCall := program.NewFnCall(
		"",
		"importWarp",
		[]*program.Operand{program.NewValue(float64('s'), loc, loc)},
		loc,
		loc,
	)

	prog := &program.Program{
		Primary: &program.Package{
			Path: "main",
			Files: []*program.PackageFile{{
				Path:            "main.gsl",
				Returns:         []*program.Operand{program.NewValue(0.0, loc, loc)},
				MaterialReturns: []*program.Operand{warpCall},
			}},
		},
	}

	assert.True(t, warpfold.HasWarpCall(prog.Primary))
}

func containsString(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
