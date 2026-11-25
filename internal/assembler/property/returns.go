package property

import (
	"fmt"

	"github.com/Gradient-Control-Laboratories/omega-core/program"
)

// LowerReturnProperties evaluates .color and .material properties across
// the program's primary package and attaches the lowered operand trees
// to the primary package's last file (ColorReturns and MaterialReturns).
//
// Where a shape inside a return sets a material with a height, it also lowers
// the return's displaced distance (DisplacedReturns) and the declarations it
// reads, and switches each layered shape's colour and material by layer: see
// layers.go.
func LowerReturnProperties(prog *program.Program) error {
	if len(prog.Primary.Files) == 0 {
		return nil
	}
	lay := analyzeLayers(prog)
	colorFields, err := layeredPropertyFields(prog.Primary, prog.GetImport, colorSpec, lay)
	if err != nil {
		return fmt.Errorf("lower color returns: %w", err)
	}
	matFields, err := layeredPropertyFields(prog.Primary, prog.GetImport, materialSpec, lay)
	if err != nil {
		return fmt.Errorf("lower material returns: %w", err)
	}
	lastFile := prog.Primary.Files[len(prog.Primary.Files)-1]
	if colorFields != nil {
		lastFile.ColorReturns = colorFields
	}
	if matFields != nil {
		lastFile.MaterialReturns = matFields
	}
	lay.attach(prog)
	return nil
}

// LowerColorReturns evaluates .color properties across the program's primary
// package and attaches the lowered color operand trees to the primary
// package's last file ColorReturns.
func LowerColorReturns(prog *program.Program) error {
	if len(prog.Primary.Files) == 0 {
		return nil
	}
	colorFields, err := ColorFields(prog.Primary, prog.GetImport)
	if err != nil {
		return fmt.Errorf("lower color returns: %w", err)
	}
	if colorFields != nil {
		lastFile := prog.Primary.Files[len(prog.Primary.Files)-1]
		lastFile.ColorReturns = colorFields
	}
	return nil
}

// LowerMaterialReturns evaluates .material properties across the program's primary
// package and attaches the lowered material operand trees to the primary
// package's last file MaterialReturns.
func LowerMaterialReturns(prog *program.Program) error {
	if len(prog.Primary.Files) == 0 {
		return nil
	}
	matFields, err := MaterialFields(prog.Primary, prog.GetImport)
	if err != nil {
		return fmt.Errorf("lower material returns: %w", err)
	}
	if matFields != nil {
		lastFile := prog.Primary.Files[len(prog.Primary.Files)-1]
		lastFile.MaterialReturns = matFields
	}
	return nil
}
