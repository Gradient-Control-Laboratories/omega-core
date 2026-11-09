package codegen

import (
	"fmt"
	"strings"

	"gradientcontrol.com/omega/core/backend/wgsl/internal/accel"
	"gradientcontrol.com/omega/core/backend/wgsl/internal/ir"
	"gradientcontrol.com/omega/core/backend/wgsl/internal/wgsl"
	"gradientcontrol.com/omega/core/internal"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

func (d *Delegate) BuildFunc(
	fnDecl *program.FnType, fn *program.Function,
	pkgName string, args []build.Expression,
) (build.FnInfo, []build.VarInfo, error) {

	fnName := fnDecl.Name
	declBScope := d.GetScope().Parent()
	parent := d.findEnclosingScope(fnDecl, pkgName)
	if parent == nil {
		return nil, nil, fmt.Errorf("missing %q function declaration", fnName)
	}

	current := d.currentScope()
	current.isFunctionScope = true
	if strings.HasPrefix(fnName, "gsl_") {
		current.parent = parent
	} else if d.isLocalScope(declBScope) {
		current.parent = d.ensureAgentScope(declBScope)
	} else {
		current.parent = parent
	}

	// Build the parameters for this invocation
	params := make([]build.VarInfo, len(fn.Parameters))
	paramNames := make([]string, len(fn.Parameters))
	for i, arg := range args {
		paramName := fn.Parameters[i].Name
		paramNames[i] = paramName
		var varType wgsl.TypeInfo
		if we, ok := arg.(wgsl.Expression); ok {
			varType = we.WGSLType()
		} else {
			varType = wgsl.WGSLType(arg.Type(), false)
		}
		param := wgsl.NewParam(d.GetScope(), paramName, arg.Type(), varType)
		params[i] = param
	}

	fnWgslName := d.buildFnName(fnName, declBScope)
	fnInfo := wgsl.NewFnInfo(fnName, fnWgslName, paramNames, args)

	// NewFnInfo appends a _<lambdaArg> suffix to the base name for each function
	// argument, so distinct specializations of the same function get distinct
	// names. buildFnName only uniquified the base, so this fuller name still has
	// to be made unique among top-level functions — otherwise a specialization
	// such as remapEdges(profile=gsl_anon) -> omega_remapEdges_gsl_anon collides
	// with the callee's own hoisted lambda of that same name, emitting two WGSL
	// functions with one name (invalid shader). Only re-uniquify when a suffix
	// was actually added, so plain (lambda-free) functions keep their names.
	if fnInfo.WgslName() != fnWgslName {
		fnInfo.SetWgslName(d.uniqueTopLevelName(fnInfo.WgslName()))
	}

	return fnInfo, params, nil
}

func (d *Delegate) BuildFuncReturn(
	fnInfo build.FnInfo,
	returnOps []*program.Operand, returnExprs []build.Expression,
) (build.FnInfo, error) {

	current := d.currentScope()

	// Determine return type
	fnName := fnInfo.Name()
	returnType := wgsl.TypeFor(returnExprs, d.structTypeCache)

	// Cache the function info
	wgslFnInfo := fnInfo.(*wgsl.FnInfo)

	globalRefs := current.globalRefs
	wgslName := wgslFnInfo.WgslName()
	paramNames := wgslFnInfo.ParamNames()
	paramTypes := wgslFnInfo.WGSLParamTypes()

	mergedRefs, _, err := d.buildFunc(fnName, wgslName, globalRefs, paramNames, paramTypes, returnType, returnExprs)
	if err != nil {
		return nil, err
	}

	var progReturnType *program.Type
	if len(returnExprs) == 1 && returnExprs[0] != nil {
		progReturnType = returnExprs[0].Type()
	}
	wgslFnInfo.Built(mergedRefs, returnType, progReturnType)

	return fnInfo, nil
}

func isRemapHelper(name string) bool {
	return strings.HasPrefix(name, "gsl_") && strings.Contains(name, "remap")
}

func (d *Delegate) buildFunc(
	fnName string,
	wgslName string,
	globalRefs []*wgsl.VarInfo,
	paramNames []string,
	paramTypes []wgsl.TypeInfo,
	returnType wgsl.TypeInfo,
	returnExprs []build.Expression,
) ([]*wgsl.VarInfo, string, error) {
	current := d.currentScope()
	assignedNames := d.getLocalAssignedNamesMap()
	for _, pName := range paramNames {
		assignedNames[wgsl.SanitizeVarName(pName)] = true
	}
	if d.rootPVar != nil {
		assignedNames[d.rootPVar.WGSLName()] = true
	}
	if d.rootTimeVar != nil {
		assignedNames[d.rootTimeVar.WGSLName()] = true
	}

	// The context used when generating the function body
	// where p is passed in as the first argument
	// if it is directly or indirectly referenced
	funcCtx := NewBuildContext(d, d.rootPVar)

	// Build the function return statement
	// which recursively builds the rest of the function body
	isStructType := returnType.IsStruct()
	var sbReturn strings.Builder
	if isStructType {
		// Destructure temps (BuildMultiAssignment) draw their emitted name from
		// the module-wide namespace, so a function holding both a destructure and
		// a struct return could otherwise land on one name twice. Step past what
		// that namespace has handed out. Checking is enough without reserving:
		// returns are built last, so this function's own destructures are already
		// emitted and their names registered.
		returnName := d.uniqueLocalName("gsl_result")
		vals := make([]string, len(returnType.FieldNames))
		for i := range returnType.FieldNames {
			vals[i] = (returnExprs[i].(wgsl.Expression)).Codegen(funcCtx)
		}
		fmt.Fprintf(&sbReturn, "  var %s : %s;\n", returnName, returnType.Name)
		current.block.Add(ir.NewVarStmt(returnName, returnType.Name, "", nil))
		for i, fieldName := range returnType.FieldNames {
			fmt.Fprintf(&sbReturn, "  %s.%s = %s;\n",
				returnName, fieldName, vals[i])
			current.block.Add(ir.NewAssignStmt(fmt.Sprintf("%s.%s", returnName, fieldName), vals[i], returnExprs[i].(wgsl.Expression).Reads()))
		}
		fmt.Fprintf(&sbReturn, "  return %s;\n", returnName)
		current.block.Add(ir.NewReturnStmt(returnName, []string{returnName}))
	} else if cropsToBox(wgslName, paramNames) {
		pName := wgsl.SanitizeVarName(paramNames[0])
		cappedVal := (returnExprs[0].(wgsl.Expression)).Codegen(funcCtx)
		cappedReads := returnExprs[0].(wgsl.Expression).Reads()
		cappedLocal := d.uniqueLocalName("gsl_capped")
		boxRet := boxCutReturn(cappedLocal, cappedVal, pName)
		sbReturn.WriteString(boxRet)
		current.block.Add(ir.NewLetStmt(cappedLocal, "", cappedVal, cappedReads))
		current.block.Add(ir.NewOpaqueStmt([]string{fmt.Sprintf("if (%s) { return %s; }", boxCapOverride, cappedLocal)}, []string{boxCapOverride, cappedLocal}, nil))
		current.block.Add(ir.NewOpaqueStmt([]string{fmt.Sprintf("if (%s <= max(%s, 0.0)) { return %s; }", cappedLocal, pName, pName)}, []string{cappedLocal, pName}, nil))
		finalRet := fmt.Sprintf("select(-1.0, 1.0, %[1]s >= 0.0) * max(abs(%[1]s), %[2]s)", pName, cappedLocal)
		current.block.Add(ir.NewReturnStmt(finalRet, []string{pName, cappedLocal}))
	} else {
		retVal := (returnExprs[0].(wgsl.Expression)).Codegen(funcCtx)
		reads := returnExprs[0].(wgsl.Expression).Reads()
		if len(reads) == 0 && retVal != "" {
			reads = []string{retVal}
		}
		fmt.Fprintf(&sbReturn, "  return %s;\n", retVal)
		current.block.Add(ir.NewReturnStmt(retVal, reads))
	}

	// Merge the passed globalRefs and the ones recorded during Codegen.
	// We want to keep the order, but if rootPVar is present, it must be first.
	var finalGlobalRefs []*wgsl.VarInfo
	seen := make(map[string]bool)

	addRef := func(g *wgsl.VarInfo) {
		if seen[g.Name()] {
			return
		}
		seen[g.Name()] = true
		if g.Name() == program.PVarName {
			// Prepend p to keep it first
			finalGlobalRefs = append([]*wgsl.VarInfo{g}, finalGlobalRefs...)
		} else {
			finalGlobalRefs = append(finalGlobalRefs, g)
		}
	}

	for _, g := range globalRefs {
		addRef(g)
	}
	for _, g := range current.globalRefs {
		addRef(g)
	}

	var params []*ir.Param
	var paramStrs []string
	for _, g := range finalGlobalRefs {
		t := g.WGSLType()
		if t.Name != "" {
			paramStrs = append(paramStrs, fmt.Sprintf("%s: %s", g.WGSLName(), t.Name))
			params = append(params, &ir.Param{Name: g.WGSLName(), Type: t.Name})
		} else {
			if !g.Type().IsFunction() {
				err := fmt.Errorf("global ref of type %v", g.Type())
				return nil, "", internal.Err(err)
			}
		}
	}
	for i, argName := range paramNames {
		t := paramTypes[i]
		if t.Name != "" {
			sanitized := wgsl.SanitizeVarName(argName)
			paramStrs = append(paramStrs, fmt.Sprintf("%s: %s", sanitized, t.Name))
			params = append(params, &ir.Param{Name: sanitized, Type: t.Name})
		}
	}

	if returnType.Name == "" {
		return nil, "", fmt.Errorf("cannot return empty type from function %s", fnName)
	}

	// Every helper function the codegen synthesizes (gsl_remap_X, user
	// function definitions, etc.) inherits the bvhSdf/vdbSdf slot
	// uniforms in its signature when the program has any. Without this,
	// any helper whose body inlines vdbSdf code (e.g. `gsl_remap` over
	// a top-level `meshDist = vdbSdf(0)`) ends up referencing sdfBMin0
	// without it being in scope. Helpers that don't actually use the
	// uniforms get them as unused parameters — fine, WGSL elides them.
	rawParams := strings.Join(paramStrs, ", ")
	hadParams := len(paramStrs) > 0
	if d.meshSdfSlots != nil && d.meshSdfSlots.Len() > 0 {
		suffix := d.meshSdfSlots.SdShapeParamsSuffix()
		if !hadParams {
			// Strip the leading ", " — no preceding params to bridge to.
			suffix = strings.TrimPrefix(suffix, ", ")
		}
		rawParams += suffix
		hadParams = true
	}

	// Thread the noise tile into every helper, like the mesh-SDF slots, so a
	// callee inlining noiseHash3 has noiseTex in scope. Unused ones get it as an
	// elided parameter.
	if d.warpUsed {
		rawParams += warpParams
	}
	if d.noiseUsed {
		if hadParams {
			rawParams += ", " + noiseTexParam
		} else {
			rawParams += noiseTexParam
		}
		hadParams = true
	}
	if d.spinodoidUsed {
		if hadParams {
			rawParams += ", " + spinodoidTexParam
		} else {
			rawParams += spinodoidTexParam
		}
		hadParams = true
	}

	// Find the top-most scope (package/primary scope) below the root scope.
	// Since WGSL does not support nested function declarations, all generated
	// functions must be hoisted and appended to the package-level
	// funcs so they are declared at the module scope of the shader.
	parent := current.parent
	for parent.parent != nil && parent.parent != d.rootScope() {
		parent = parent.parent
	}

	// Sinking and local dead-code pruning via IR
	ir.SinkStatements(current.block)

	fnDecl := ir.NewFuncDecl(wgslName, params, returnType.Name, current.block)
	fnDecl.RawParams = rawParams
	if cropsToBox(wgslName, paramNames) {
		fnDecl.TrailingSource = boxCapDecl()
	}
	current.block = ir.NewBlock()

	// shrinkToVanish is never called: every call is its guard (vanish.go).
	if vanishes(wgslName, len(paramNames)) {
		return finalGlobalRefs, wgslName, nil
	}
	// Nor are simplified and simpleValue: every call is its switch
	// (simple.go). Their parameters may include module state after their own.
	if simplifiedNameRE.MatchString(wgslName) || simpleValueNameRE.MatchString(wgslName) || detailValueNameRE.MatchString(wgslName) {
		return finalGlobalRefs, wgslName, nil
	}

	if accel.IsAccelerated(wgslName, paramTypes) {
		return finalGlobalRefs, wgslName, nil
	}

	if isRemapHelper(wgslName) {
		var buf strings.Builder
		_ = fnDecl.Format(&buf)
		key := buf.String()[len("fn ")+len(wgslName):]
		if canonical, exists := d.remapHelpers[key]; exists {
			delete(d.assignedNames, wgslName)
			return finalGlobalRefs, canonical, nil
		}
		d.remapHelpers[key] = wgslName
	}

	// Append hoisted functions and this function declaration to the parent scope
	if len(current.funcs) > 0 {
		parent.funcs = append(parent.funcs, current.funcs...)
		current.funcs = nil
	}
	parent.funcs = append(parent.funcs, fnDecl)

	return finalGlobalRefs, wgslName, nil
}

func (d *Delegate) NewParamVar(scope *build.Scope, name string, expr build.Expression, typ *program.Type) build.VarInfo {
	var wgslExpr wgsl.Expression
	if expr != nil {
		if we, ok := expr.(wgsl.Expression); ok {
			wgslExpr = we
		}
	}
	return wgsl.NewParamWithExpr(scope, name, wgslExpr, typ, wgsl.WGSLType(typ, false))
}
