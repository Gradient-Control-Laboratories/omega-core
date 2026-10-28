package codegen

import (
	"fmt"

	"gradientcontrol.com/omega/core/backend/wgsl/internal/lib/internal/wgsl"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// BuildAssignment implements [build.Delegate].
func (d *Delegate) BuildSingleAssignment(assign *program.Assignment, op *program.Operand,
	name string, expr build.Expression) (build.VarInfo, error) {

	if assign.Range != nil {
		if paramVar, ok := d.sliderParamVars[name]; ok && paramVar != nil {
			d.recordRef(paramVar)
			expr = wgsl.NewVarRef(op, paramVar)
		}
	}

	varInfo := wgsl.NewVar(d.GetScope(), name, expr, expr.(wgsl.Expression).WGSLType())
	varInfo.SetWGSLName(d.uniqueLocalName(varInfo.WGSLName()))
	return varInfo, nil
}

// BuildMultiAssignment implements [build.Delegate].
func (d *Delegate) BuildMultiAssignment(_ *program.Assignment, _ *program.Operand,
	fnCall *program.FnCall, names []string, expr build.Expression) ([]build.VarInfo, error) {

	structType := expr.(wgsl.Expression).WGSLType()
	if !structType.IsStruct() {
		return nil, fmt.Errorf("expected struct type for %s, got %s", fnCall.Reference.Name, expr.Type())
	}

	if len(structType.Types) != len(names) {
		return nil, fmt.Errorf("expected %d types for %s", len(names), fnCall.Reference.Name)
	}

	fieldNames := structType.FieldNames

	// NewVarName only uniquifies within the declaring GSL scope, but the emitted
	// `let` lands in whatever WGSL function the destructure is inlined into, and
	// the guard that emits it (markGenerated below) is keyed by that emitted
	// name. Two destructures declared in different GSL scopes that inline into
	// the same WGSL function therefore used to collide on `gsl_result`: the
	// first reference emitted its call, the second was suppressed and its field
	// reads silently re-bound to the other struct. Reserving the emitted name
	// module-wide keeps every destructure temp distinct, so no destructure can
	// swallow another's call.
	resultName := d.GetScope().NewVarName("gsl_result")
	resultWGSLName := d.uniqueTopLevelName(resultName)
	// A struct return's temp (function.go) steps past this function's local
	// names, so record the temp there too.
	d.getLocalAssignedNamesMap()[resultWGSLName] = true
	structVar := wgsl.NewVar(d.GetScope(), resultName, expr, expr.(wgsl.Expression).WGSLType())
	structVar.SetWGSLName(resultWGSLName)
	err := d.GetScope().AddLocalVar(structVar)
	if err != nil {
		return nil, err
	}

	bScope := d.GetScope()
	varInfos := make([]build.VarInfo, len(names))
	for i, name := range names {
		field := fieldNames[i]
		structRef := wgsl.NewVarRef(nil, structVar)
		varType := structType.Types[i]
		subExpr := wgsl.NewFieldAccess(nil, structRef, field, varType)

		varInfo := wgsl.NewVar(bScope, name, subExpr, varType)
		varInfo.SetWGSLName(d.uniqueLocalName(varInfo.WGSLName()))
		varInfos[i] = varInfo
	}

	return varInfos, nil
}
