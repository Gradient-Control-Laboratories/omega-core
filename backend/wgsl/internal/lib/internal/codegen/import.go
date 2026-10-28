package codegen

import (
	"fmt"

	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// BuildImport implements [build.Delegate].
func (d *Delegate) BuildImport(imp *program.Import, pkg *program.Package) (build.ImportInfo, error) {

	s := d.currentScope()
	s.pkgPath = imp.Path
	s.name = imp.PrefixOrSimpleName()

	return newImportInfo(imp, pkg, s), nil
}

// EndBuildImport implements [build.Delegate].
func (d *Delegate) EndBuildImport(info build.ImportInfo) error {
	impInfo := info.(*ImportInfo)
	imp := impInfo.imp
	s := impInfo.scope

	curScope := s.importingScope

	prefixOrSimpleName := imp.PrefixOrSimpleName()
	cacheKey := prefixOrSimpleName
	if cacheKey == "." {
		cacheKey += imp.Path
	}
	if curScope.prefixImpCache == nil {
		curScope.prefixImpCache = make(map[string]*Scope)
	} else if _, hasImp := curScope.prefixImpCache[cacheKey]; hasImp {
		return fmt.Errorf("import with prefix %q already exists", prefixOrSimpleName)
	}
	curScope.prefixImpCache[cacheKey] = s

	d.imports = append(d.imports, impInfo)

	return nil
}
