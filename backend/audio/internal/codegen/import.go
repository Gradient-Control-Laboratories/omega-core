package codegen

import (
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// BuildImport implements [build.Delegate]. JS audio output now supports
// imports — function declarations from the imported package are visited
// in the import scope and accumulated into d.fnDefs alongside primary-pkg
// functions; lookups go through the common build framework's scope chain.
// pkgPath on the agent scope is used to disambiguate identifier names
// (functions and top-level constants) across packages.
func (d *Delegate) BuildImport(imp *program.Import, pkg *program.Package) (build.ImportInfo, error) {
	d.currentScope().pkgPath = imp.Path
	return &ImportInfo{imp: imp, pkg: pkg}, nil
}

// EndBuildImport implements [build.Delegate].
func (d *Delegate) EndBuildImport(info build.ImportInfo) error {
	return nil
}

// ImportInfo is the JS-side carrier for build.ImportInfo. The common build
// framework already wires up scope-based name resolution for imported
// declarations, so the agent just exposes the import + package handles
// required by the interface.
type ImportInfo struct {
	imp *program.Import
	pkg *program.Package
}

var _ build.ImportInfo = (*ImportInfo)(nil)

func (i *ImportInfo) Import() *program.Import   { return i.imp }
func (i *ImportInfo) Package() *program.Package { return i.pkg }
