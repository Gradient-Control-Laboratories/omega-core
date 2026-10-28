package parser2

import (
	"errors"

	"gradientcontrol.com/omega/core/internal/scope"
	"gradientcontrol.com/omega/core/program"
)

func buildPkgFile(pkgScope *scope.PkgScope, topLevel *TopLevel) (*program.PackageFile, error) {
	if topLevel == nil {
		return nil, errors.New("missing declarations")
	}

	pkgFile := &program.PackageFile{}
	pkgFileScope := pkgScope.NewPkgFileScope(pkgFile)

	var err error
	err = buildImports(pkgFileScope, topLevel.Imports)
	if err != nil {
		return pkgFile, err
	}

	err = buildMeshImports(pkgFileScope, topLevel.MeshImports)
	if err != nil {
		return pkgFile, err
	}

	err = buildStepImports(pkgFileScope, topLevel.StepImports)
	if err != nil {
		return pkgFile, err
	}

	err = buildWarpImports(pkgFileScope, topLevel.WarpImports)
	if err != nil {
		return pkgFile, err
	}

	err = buildDeclarations(pkgFileScope, topLevel.Declarations)
	if err != nil {
		return pkgFile, err
	}

	err = buildReturnExpressions(pkgFileScope, topLevel.Return)
	if err != nil {
		return pkgFile, err
	}

	return pkgFile, err
}
