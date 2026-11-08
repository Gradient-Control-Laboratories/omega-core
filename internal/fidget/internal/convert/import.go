package convert

import (
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

type ImportInfo struct {
	imp   *program.Import
	pkg   *program.Package
	scope *Scope
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

// BuildImport implements [build.Delegate].
func (d *Delegate) BuildImport(imp *program.Import, pkg *program.Package) (build.ImportInfo, error) {
	s := d.currBScope.GetAgentScope().(*Scope)
	return &ImportInfo{
		imp:   imp,
		pkg:   pkg,
		scope: s,
	}, nil
}

// EndBuildImport implements [build.Delegate].
func (d *Delegate) EndBuildImport(info build.ImportInfo) error {
	return nil
}
