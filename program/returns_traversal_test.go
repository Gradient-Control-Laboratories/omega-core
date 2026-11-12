package program_test

import (
	"testing"

	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/program"
	"gradientcontrol.com/omega/core/program/refs"
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

func TestRefs_ColorAndMaterialReturns(t *testing.T) {
	loc := &program.Location{Line: 1, Column: 1}
	colorMeshCall := program.NewFnCall(
		"",
		"bvhSdf",
		[]*program.Operand{program.NewValue(2.0, loc, loc)},
		loc,
		loc,
	)
	matMeshCall := program.NewFnCall(
		"",
		"mesh",
		[]*program.Operand{program.NewValue(3.0, loc, loc)},
		loc,
		loc,
	)

	pkg := &program.Package{
		Path: "main",
		Files: []*program.PackageFile{{
			Path:            "main.gsl",
			Returns:         []*program.Operand{program.NewValue(0.0, loc, loc)},
			ColorReturns:    []*program.Operand{colorMeshCall},
			MaterialReturns: []*program.Operand{matMeshCall},
		}},
	}

	meshRefs := refs.CollectMeshReferences(pkg)
	assert.Equal(t, 2, len(meshRefs))
	assert.Equal(t, 2, meshRefs[0].Index)
	assert.Equal(t, 3, meshRefs[1].Index)
}
