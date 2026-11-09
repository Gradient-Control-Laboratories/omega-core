package property

import (
	"fmt"

	"gradientcontrol.com/omega/core/program"
)

// LowerReturnProperties evaluates .color and .material properties across
// the program's primary package and attaches the lowered operand trees
// to the primary package's last file (ColorReturns and MaterialReturns).
func LowerReturnProperties(prog *program.Program) error {
	if err := LowerColorReturns(prog); err != nil {
		return err
	}
	return LowerMaterialReturns(prog)
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
