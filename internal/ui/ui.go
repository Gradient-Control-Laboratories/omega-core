package ui

import (
	"gradientcontrol.com/omega/core/assembler/ui"
	"gradientcontrol.com/omega/core/program"
	"gradientcontrol.com/omega/core/program/edit"
)

func Extract(pkg *program.Package, getImport program.GetImportFunc) []ui.Control {
	return ui.Extract(pkg, getImport)
}

func ExtractWithOperands(pkg *program.Package, getImport program.GetImportFunc) ([]ui.Control, map[*program.Operand]edit.Span) {
	return ui.ExtractWithOperands(pkg, getImport)
}
