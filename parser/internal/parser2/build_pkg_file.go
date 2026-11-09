package parser2

import (
	"gradientcontrol.com/omega/core/internal/scope"
)

func buildPkgFile(pkgFileScope scope.BuilderScope, topLevel *TopLevel) error {
	var err error
	err = buildImports(pkgFileScope, topLevel.Imports)
	if err != nil {
		return err
	}

	err = buildMeshImports(pkgFileScope, topLevel.MeshImports)
	if err != nil {
		return err
	}

	err = buildStepImports(pkgFileScope, topLevel.StepImports)
	if err != nil {
		return err
	}

	err = buildWarpImports(pkgFileScope, topLevel.WarpImports)
	if err != nil {
		return err
	}

	err = buildDeclarations(pkgFileScope, topLevel.Declarations)
	if err != nil {
		return err
	}

	err = buildReturnExpressions(pkgFileScope, topLevel.Return)
	if err != nil {
		return err
	}

	return nil
}
