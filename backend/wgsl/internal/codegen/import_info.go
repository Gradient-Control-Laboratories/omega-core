package codegen

import (
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

type ImportInfo struct {
	imp   *program.Import
	pkg   *program.Package
	scope *Scope
}

func newImportInfo(imp *program.Import, pkg *program.Package, s *Scope) *ImportInfo {
	return &ImportInfo{imp: imp, pkg: pkg, scope: s}
}

// Ensure that ImportInfo implements the build.ImportInfo interface.
var _ build.ImportInfo = (*ImportInfo)(nil)

// Import implements [build.ImportInfo].
func (i *ImportInfo) Import() *program.Import {
	return i.imp
}

// Package implements [build.ImportInfo].
func (i *ImportInfo) Package() *program.Package {
	return i.pkg
}
