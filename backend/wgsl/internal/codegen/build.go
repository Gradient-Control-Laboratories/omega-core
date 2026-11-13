package codegen

import (
	"fmt"
	"sort"
	"strings"

	"gradientcontrol.com/omega/core/assembler/slider"
	"gradientcontrol.com/omega/core/assembler/ui"
	"gradientcontrol.com/omega/core/backend/wgsl/internal/ir"
	"gradientcontrol.com/omega/core/backend/wgsl/internal/wgsl"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
	"gradientcontrol.com/omega/core/program/edit"
)

// emptyShapeWGSL is what sdShape returns when the program returns no shapes
// (an audio-only program, say): a distance far enough that nothing is hit,
// and shape index 0, which the viewers read as background.
const emptyShapeWGSL = "gsl_shape(1e10, 0u)"

// WGSL generates WGSL code from the given GSL source code and package.
//
//   - prog: The self-contained assembled program.
func WGSL(
	prog *program.Program,
) (string, []*program.MeshSdfSlot, error) {
	wgslSource, slots, _, _, err := WGSLWithParts(prog)
	return wgslSource, slots, err
}

// WGSLWithParts is WGSL, and also the box of each part the viewer guards whose
// box is constant and the spinodoid table keys the shader reads.
func WGSLWithParts(
	prog *program.Program,
) (string, []*program.MeshSdfSlot, [][2][3]float64, []string, error) {
	// Invariant: prog and prog.Primary are guaranteed non-nil by boundary validation in logic.Render.
	meshSdfSlots := NewMeshSdfSlotsFromProto(prog.MeshSdfSlots)

	d := &Delegate{
		prog:            prog,
		structTypeCache: make(map[string]wgsl.TypeInfo),
		module:          ir.NewModule(),
		meshSdfSlots:    meshSdfSlots,
		assignedNames:   make(map[string]bool),
		remapHelpers:    make(map[string]string),
		sliderVars:      program.ComputeSliderVars(prog.Primary),
		uiParamVars:     make(map[edit.Span]*wgsl.VarInfo),
	}
	if prog.GlobalTextures != nil {
		d.noiseUsed = prog.GlobalTextures.UsesNoiseTile
		d.spinodoidUsed = prog.GlobalTextures.UsesSpinodoid
		d.warpUsed = prog.GlobalTextures.UsesWarpField
	}
	controls, operands := ui.ExtractWithOperands(prog.Primary, prog.GetImport)
	d.uiParamOperands = operands
	for _, control := range controls {
		for _, param := range control.Params {
			if _, exists := d.uiParamVars[param.Span]; exists {
				continue
			}
			d.uiParams = append(d.uiParams, param)
			d.uiParamVars[param.Span] = nil
		}
	}
	if !hasPackageReturns(prog.Primary) {
		d.buildEmptyProgram()
		return d.wgslSource, d.meshSdfSlots.Ordered(), d.parts, d.spinodoidSlots, nil
	}
	err := build.Build(d, prog, false)
	err = build.ErrWithPkgSrc(prog, err)
	return d.wgslSource, d.meshSdfSlots.Ordered(), d.parts, d.spinodoidSlots, err
}

func hasPackageReturns(pkg *program.Package) bool {
	// Invariant: pkg and pkg.Files elements are guaranteed non-nil.
	for _, f := range pkg.Files {
		if len(f.Returns) > 0 {
			return true
		}
	}
	return false
}

func (d *Delegate) buildEmptyProgram() {
	// Invariant: d.prog.Primary, its Files, Assignments, and lifted assignment Expression.Start are guaranteed non-nil.
	lifted := slider.Lifted(d.prog.Primary)
	for _, f := range d.prog.Primary.Files {
		for _, a := range f.Assignments {
			if !lifted[a] {
				continue
			}
			loc := a.Expression.Start
			d.sliderParamNames = append(d.sliderParamNames, fmt.Sprintf("gsl_param_%d_%d", loc.Line, loc.Column))
		}
	}
	for _, param := range d.uiParams {
		d.sliderParamNames = append(d.sliderParamNames, param.ParamName)
	}
	var paramStrs []string
	paramStrs = append(paramStrs, "p: vec3<f32>", "time: f32")
	for _, name := range d.sliderParamNames {
		paramStrs = append(paramStrs, fmt.Sprintf("%s: f32", name))
	}
	sdShapeBlock := ir.NewBlock()
	sdShapeBlock.Add(ir.NewReturnStmt(emptyShapeWGSL, []string{emptyShapeWGSL}))
	mainFnDecl := ir.NewFuncDecl("sdShape", nil, "gsl_shape", sdShapeBlock)
	mainFnDecl.RawParams = strings.Join(paramStrs, ", ")
	d.module.AddFunction(mainFnDecl)
	d.wgslSource = d.module.WGSL()
}

func (d *Delegate) NewImpliedVars(bScope *build.Scope) ([]build.VarInfo, error) {
	// Invariant: d.prog, d.prog.Primary, its Files, Assignments, Variables, and lifted Expression.Start are guaranteed non-nil.
	d.sliderParamVars = make(map[string]*wgsl.VarInfo)
	d.sliderParamNames = nil

	var count int
	for _, f := range d.prog.Primary.Files {
		for _, a := range f.Assignments {
			if a.Range != nil {
				count++
			}
		}
	}

	impliedVars := make([]build.VarInfo, 0, count+2)

	pExpr := wgsl.NewOpaqueExpr(nil, wgsl.TypeVec3, build.PrecedenceHighest, program.PVarName)
	d.rootPVar = wgsl.NewParamWithExpr(bScope, program.PVarName, pExpr, nil, wgsl.TypeVec3)
	impliedVars = append(impliedVars, d.rootPVar)

	timeExpr := wgsl.NewOpaqueExpr(nil, wgsl.TypeF32, build.PrecedenceHighest, program.TimeVarName)
	d.rootTimeVar = wgsl.NewParamWithExpr(bScope, program.TimeVarName, timeExpr, nil, wgsl.TypeF32)
	impliedVars = append(impliedVars, d.rootTimeVar)

	// Only a range the viewer gets a slider for becomes a parameter; any
	// other keeps its literal, or sdShape is called an argument short.
	lifted := slider.Lifted(d.prog.Primary)
	for _, f := range d.prog.Primary.Files {
		for _, a := range f.Assignments {
			if !lifted[a] {
				continue
			}
			loc := a.Expression.Start
			name := fmt.Sprintf("gsl_param_%d_%d", loc.Line, loc.Column)
			expr := wgsl.NewOpaqueExpr(nil, wgsl.TypeF32, build.PrecedenceHighest, name)
			varInfo := wgsl.NewParamWithExpr(bScope, name, expr, nil, wgsl.TypeF32)
			impliedVars = append(impliedVars, varInfo)
			d.sliderParamNames = append(d.sliderParamNames, name)
			for _, v := range a.Variables {
				d.sliderParamVars[v.Name] = varInfo
			}
		}
	}

	// The viewer passes UI literals after range sliders, in extraction order.
	// Keep even unused controls in the signature so its arguments stay aligned.
	for _, param := range d.uiParams {
		name := param.ParamName
		expr := wgsl.NewOpaqueExpr(nil, wgsl.TypeF32, build.PrecedenceHighest, name)
		varInfo := wgsl.NewParamWithExpr(bScope, name, expr, nil, wgsl.TypeF32)
		impliedVars = append(impliedVars, varInfo)
		d.sliderParamNames = append(d.sliderParamNames, name)
		d.uiParamVars[param.Span] = varInfo
	}

	return impliedVars, nil
}

func (d *Delegate) BuildPackage(_ *program.Package) error {
	// At this point the current scope is the primary (top-level, unnamed) scope;
	// its parent is the root scope.
	d.primaryBScope = d.GetScope()
	d.rootBScope = d.GetScope().Parent()
	d.currentScope().isFunctionScope = true

	// Initialize the root build/codegen context
	d.rootCtx.delegate = d
	d.rootCtx.PVar = d.rootPVar

	return nil
}

// BuildPackageReturns implements [build.Delegate].
func (d *Delegate) BuildPackageReturns(_ []*program.Operand, expressions []build.Expression) error {
	for _, fn := range d.rootScope().funcs {
		d.module.AddFunction(fn)
	}

	for _, impInfo := range d.imports {
		s := impInfo.scope
		if s.parent != nil && s.parent != d.rootScope() {
			s = s.parent
		}
		for _, fn := range s.funcs {
			d.module.AddFunction(fn)
		}
	}

	curr := d.currentScope()
	for curr.parent != nil && curr.parent != d.rootScope() {
		curr = curr.parent
	}
	for _, fn := range curr.funcs {
		d.module.AddFunction(fn)
	}

	if d.warpUsed {
		d.module.AddRuntimeSnippet(warpSampleWGSL)
	}

	var paramStrs []string
	paramStrs = append(paramStrs, "p: vec3<f32>", "time: f32")
	for _, name := range d.sliderParamNames {
		paramStrs = append(paramStrs, fmt.Sprintf("%s: f32", name))
	}
	rawParams := strings.Join(paramStrs, ", ")
	rawParams += d.meshSdfSlots.SdShapeParamsSuffix()
	if d.warpUsed {
		rawParams += warpParams
	}
	if d.noiseUsed {
		rawParams += ", " + noiseTexParam
	}
	if d.spinodoidUsed {
		rawParams += ", " + spinodoidTexParam
	}

	var suffix strings.Builder
	for _, name := range d.sliderParamNames {
		fmt.Fprintf(&suffix, ", %s: f32", name)
	}
	suffix.WriteString(d.meshSdfSlots.SdShapeParamsSuffix())
	if d.warpUsed {
		suffix.WriteString(warpParams)
	}
	if d.noiseUsed {
		suffix.WriteString(", " + noiseTexParam)
	}
	if d.spinodoidUsed {
		suffix.WriteString(", " + spinodoidTexParam)
	}

	// A program with a displaced return builds its returns' distances once, in
	// gsl_shapeParts, which both sdShape and the displaced composite read (see
	// displacedComposite). With one return the winner's height is its own, and
	// the viewer's plain path is already exact; so it is when every displaced
	// return carries the same height (heightsDiffer).
	displaced := d.heightsSrc != "" && len(expressions) > 1 && d.heightsDiffer() && d.displacementEvident(d.module.WGSL())

	sdShapeBlock := ir.NewBlock()
	if d.vanishUsed {
		sdShapeBlock.Add(ir.NewAssignStmt("gsl_partNear", "vec3<f32>(1e10, 1e10, 1e10)", nil))
	}

	// The context used when generating the function body
	// where p is passed in as the first argument
	// if it is directly or indirectly referenced
	funcCtx := NewBuildContext(d, d.rootPVar)

	partSrcs := make([]string, len(expressions))
	for index, expr := range expressions {
		partSrcs[index] = (expr.(wgsl.Expression)).Codegen(funcCtx)
	}

	for _, stmt := range d.currentScope().Block().Statements() {
		sdShapeBlock.Add(stmt)
	}
	d.currentScope().block = ir.NewBlock()

	if displaced {
		sdShapeBlock.Add(ir.NewReturnStmt(
			fmt.Sprintf("array<f32, %d>(%s)", len(expressions), strings.Join(partSrcs, ", ")),
			partSrcs,
		))
	} else {
		var returnExprWsglSrc string
		for index, partSrc := range partSrcs {
			letName := fmt.Sprintf("gsl_s%d", index+1)
			shapeExpr := fmt.Sprintf("gsl_shape(%s, %du)", partSrc, index+1)
			sdShapeBlock.Add(ir.NewLetStmt(letName, "", shapeExpr, []string{partSrc}))
			if index == 0 {
				returnExprWsglSrc = letName
			} else {
				returnExprWsglSrc = fmt.Sprintf("gsl_min(%s, %s)", returnExprWsglSrc, letName)
			}
		}
		if returnExprWsglSrc == "" {
			returnExprWsglSrc = emptyShapeWGSL
		}
		sdShapeBlock.Add(ir.NewReturnStmt(returnExprWsglSrc, []string{returnExprWsglSrc}))
	}

	var fnName, returnType string
	if displaced {
		fnName = shapePartsFn
		returnType = fmt.Sprintf("array<f32, %d>", len(expressions))
	} else {
		fnName = "sdShape"
		returnType = "gsl_shape"
	}
	mainFnDecl := ir.NewFuncDecl(fnName, nil, returnType, sdShapeBlock)
	mainFnDecl.RawParams = rawParams
	d.module.AddFunction(mainFnDecl)

	if displaced {
		sdShapeSrc, displacedSrc := d.displacedComposite(suffix.String(), len(expressions))
		d.module.AddFunction(ir.NewRawFuncDecl(sdShapeSrc))
		d.displacedFieldSrc = displacedSrc
	}

	// Weave the generated color function (name.color = …) in after sdShape and
	// before the struct declarations. Empty unless the program authors color.
	if d.colorFieldSrc != "" {
		d.module.AddFunction(ir.NewRawFuncDecl(d.colorFieldSrc))
	}
	if d.materialFieldSrc != "" {
		d.module.AddFunction(ir.NewRawFuncDecl(d.materialFieldSrc))
	}
	if d.cellularUsed {
		d.module.AddRuntimeSnippet(cellularWGSL)
		if d.cellularWallUsed {
			d.module.AddRuntimeSnippet(cellularWallWGSL)
		}
	}
	if d.spinodoidEmitted {
		d.module.AddRuntimeSnippet(spinodoidWGSL)
	}
	if d.spinodoidGradEmitted {
		d.module.AddRuntimeSnippet(spinodoidGradWGSL)
	}
	if d.vanishUsed {
		d.module.AddRuntimeSnippet(vanishWGSL)
	}
	if d.simpleUsed {
		d.module.AddRuntimeSnippet(simpleWGSL)
	}
	if d.simpleValueUsed {
		d.module.AddRuntimeSnippet(simpleValueDeclWGSL)
	}
	if d.dispDrawnUsed {
		d.module.AddRuntimeSnippet(dispDrawnDeclWGSL)
	}
	if d.detailUsed {
		d.module.AddRuntimeSnippet(detailPwDeclWGSL)
	}
	for _, decl := range d.foldConstDecls {
		d.module.AddRuntimeSnippet(decl)
	}
	if d.displacedFieldSrc != "" {
		d.module.AddRuntimeSnippet(d.heightsSrc)
		d.module.AddRuntimeSnippet(d.displacedFieldSrc)
	}

	// Append struct type declarations at the end
	// so that the client's wgslFn function finds function declarations first.
	// If a struct type declaration appears first, the client's wgslFn function
	// will try to parse it as a function declaration and fail
	// with the cryptic error message.
	structTypeList := make([]wgsl.TypeInfo, 0, len(d.structTypeCache))
	for _, structType := range d.structTypeCache {
		structTypeList = append(structTypeList, structType)
	}
	sort.Slice(structTypeList, func(i, j int) bool {
		return structTypeList[i].Name < structTypeList[j].Name
	})
	for _, structType := range structTypeList {
		fields := make([]ir.StructField, len(structType.FieldNames))
		for i, name := range structType.FieldNames {
			fields[i] = ir.StructField{
				Name: name,
				Type: structType.Types[i].Name,
			}
		}
		structDecl := ir.NewStructDecl(structType.Name, fields)
		d.module.AddStruct(structDecl)
	}

	d.wgslSource = d.module.WGSL()
	return nil
}
