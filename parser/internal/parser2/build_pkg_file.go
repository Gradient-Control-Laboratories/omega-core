package parser2

import (
	"gradientcontrol.com/omega/core/internal/scope"
	"gradientcontrol.com/omega/core/program"
)

func buildPkgFile(pkgFileScope scope.BuilderScope, source string, topLevel *TopLevel) ([]program.Diagnostic, error) {
	diags, err := buildImports(pkgFileScope, source, topLevel.Imports)
	if err != nil {
		return diags, err
	}

	err = buildMeshImports(pkgFileScope, topLevel.MeshImports)
	if err != nil {
		return diags, err
	}

	err = buildStepImports(pkgFileScope, topLevel.StepImports)
	if err != nil {
		return diags, err
	}

	err = buildWarpImports(pkgFileScope, topLevel.WarpImports)
	if err != nil {
		return diags, err
	}

	err = buildDeclarations(pkgFileScope, topLevel.Declarations)
	if err != nil {
		return diags, err
	}

	err = buildReturnExpressions(pkgFileScope, topLevel.Return)
	if err != nil {
		return diags, err
	}

	return diags, nil
}
