package property

import (
	"testing"

	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/program"
)

func TestLowerReturnProperties_BothColorAndMaterial(t *testing.T) {
	pkg := &program.Package{
		Path: "main",
		Files: []*program.PackageFile{
			{
				Assignments: []*program.Assignment{
					coloredShape("ball", sphereSDF(0, 0, 0, 1), rgb(1, 0, 0)),
					materialShape("ball", sphereSDF(0, 0, 0, 1), materialLit(0.1, 1, 1.5, 0, 0, 0, 0, 0, 0, 0)),
				},
				Returns: []*program.Operand{ref("ball")},
			},
		},
	}
	prog := &program.Program{Primary: pkg}

	err := LowerReturnProperties(prog)
	assert.NoError(t, err)

	lastFile := prog.Primary.Files[0]
	assert.Equal(t, 1, len(lastFile.ColorReturns))
	assert.True(t, lastFile.ColorReturns[0] != nil)

	assert.Equal(t, 1, len(lastFile.MaterialReturns))
	assert.True(t, lastFile.MaterialReturns[0] != nil)
}

func TestLowerReturnProperties_Neither(t *testing.T) {
	pkg := &program.Package{
		Path: "main",
		Files: []*program.PackageFile{
			{
				Assignments: []*program.Assignment{
					{
						Variables:  []*program.Variable{{Name: "plain"}},
						Expression: sphereSDF(0, 0, 0, 1),
					},
				},
				Returns: []*program.Operand{ref("plain")},
			},
		},
	}
	prog := &program.Program{Primary: pkg}

	err := LowerReturnProperties(prog)
	assert.NoError(t, err)

	lastFile := prog.Primary.Files[0]
	assert.Equal(t, 0, len(lastFile.ColorReturns))
	assert.Equal(t, 0, len(lastFile.MaterialReturns))
}

func TestLowerReturnProperties_OnlyColor(t *testing.T) {
	pkg := twoBallPkg(ref("ball"))
	prog := &program.Program{Primary: pkg}

	err := LowerReturnProperties(prog)
	assert.NoError(t, err)

	lastFile := prog.Primary.Files[0]
	assert.Equal(t, 1, len(lastFile.ColorReturns))
	assert.True(t, lastFile.ColorReturns[0] != nil)
	assert.Equal(t, 0, len(lastFile.MaterialReturns))
}

func TestLowerReturnProperties_OnlyMaterial(t *testing.T) {
	pkg := twoMaterialPkg(ref("ball"))
	prog := &program.Program{Primary: pkg}

	err := LowerReturnProperties(prog)
	assert.NoError(t, err)

	lastFile := prog.Primary.Files[0]
	assert.Equal(t, 0, len(lastFile.ColorReturns))
	assert.Equal(t, 1, len(lastFile.MaterialReturns))
	assert.True(t, lastFile.MaterialReturns[0] != nil)
}
