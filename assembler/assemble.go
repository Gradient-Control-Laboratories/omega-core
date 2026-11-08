package assembler

import (
	"fmt"
	"strings"

	"gradientcontrol.com/omega/core/assembler/meshplan"
	"gradientcontrol.com/omega/core/assembler/meshslots"
	"gradientcontrol.com/omega/core/assembler/slider"
	"gradientcontrol.com/omega/core/assembler/topofold"
	"gradientcontrol.com/omega/core/assembler/ui"
	"gradientcontrol.com/omega/core/assembler/warpfold"
	"gradientcontrol.com/omega/core/internal/assembler/annotation"
	"gradientcontrol.com/omega/core/internal/assembler/constspec"
	"gradientcontrol.com/omega/core/internal/assembler/deadcode"
	"gradientcontrol.com/omega/core/internal/assembler/inline"
	"gradientcontrol.com/omega/core/internal/assembler/property"
	"gradientcontrol.com/omega/core/internal/assembler/propval"
	internaltopofold "gradientcontrol.com/omega/core/internal/assembler/topofold"
	"gradientcontrol.com/omega/core/internal/assembler/unroll"
	internalwarpfold "gradientcontrol.com/omega/core/internal/assembler/warpfold"
	"gradientcontrol.com/omega/core/internal/closure"
	"gradientcontrol.com/omega/core/internal/collect"
	"gradientcontrol.com/omega/core/internal/eval"
	"gradientcontrol.com/omega/core/internal/link"
	"gradientcontrol.com/omega/core/internal/lower"
	"gradientcontrol.com/omega/core/internal/meshfold"
	internalslider "gradientcontrol.com/omega/core/internal/slider"
	internalsource "gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/types"
	internalui "gradientcontrol.com/omega/core/internal/ui"
	"gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
	"gradientcontrol.com/omega/core/program/textures"
)

// ParsePackageFunc parses all source files in a package into a *parser.Result.
type ParsePackageFunc func(sources []string) (*parser.Result, error)

// VariantResult holds the sliced MeshExportPlan, analytical SdfSubtrees, or
// per-variant compilation error for a single caller-supplied variant return.
type VariantResult struct {
	MeshPlan    *meshplan.MeshExportPlan `json:"meshPlan,omitempty"`
	SdfSubtrees []meshplan.SdfSubtree    `json:"sdfSubtrees,omitempty"`
	Error       string                   `json:"error,omitempty"`
}

// Result contains the assembled program AST and the assigned import slot table.
type Result struct {
	Program        *program.Program
	Slots          []meshslots.Import
	WarpedAssets   []warpfold.WarpedAsset
	MeshSdfSlots   []*program.MeshSdfSlot
	GlobalTextures *program.GlobalTextures
	Diagnostics    []program.Diagnostic
	Sliders        []slider.SliderInfo
	UIControls     []ui.Control
	Scene          *program.Scene
	MeshPlan       *meshplan.MeshExportPlan
	SdfSubtrees    []meshplan.SdfSubtree
	PreviewBodies  []meshplan.PreviewBody
	AudioReturns   []meshplan.SdfSubtree
	Variants       map[string]VariantResult
}

// Assemble collects packages directly and indirectly referenced by primarySources,
// executes compiler passes, and assembles them into an optimized Program AST.
func Assemble(primarySources []string, opts ...Option) (*Result, error) {
	if len(primarySources) == 0 {
		return nil, fmt.Errorf("no primary sources provided")
	}

	o := defaultOptions()
	for _, opt := range opts {
		if opt != nil {
			opt(&o)
		}
	}

	var srcMgr internalsource.Manager = internalsource.NewStdLibManager()
	if len(o.managers) > 0 {
		managers := make([]internalsource.Manager, 0, len(o.managers)+1)
		managers = append(managers, o.managers...)
		managers = append(managers, srcMgr)
		srcMgr = internalsource.NewMultiSourceManager(managers...)
	}

	return assembleWithOptions(primarySources, srcMgr, &o)
}

func defaultOptions() options {
	return options{
		parsePkg:              parser.ParsePackage,
		enableSlotFold:        true,
		enableDeadcodeRemoval: true,
	}
}

func assembleWithOptions(primarySources []string, srcMgr internalsource.Manager, o *options) (*Result, error) {
	// The mesh plan is sliced after step 7 without a step of its own: it is
	// plumbing (the preview's mesh placements, the export's leaves), not
	// something to show a person waiting on a compile.
	totalSteps := uint32(7)
	report := func(stage string, completed uint32) {
		if o.progressFn != nil {
			o.progressFn(stage, completed, totalSteps)
		}
	}

	// 1. Package Collection & Initial Linking
	// Parse the primary package and transitively collect imported packages into a Program AST,
	// then resolve and link all symbol references across package boundaries.
	report("Resolving package imports", 1)
	prog, err := collect.Program(primarySources, srcMgr, o.parsePkg)
	if err != nil {
		return nil, err
	}
	rp, err := partitionReturns(prog, primarySources, o.variants, o.enableDeadcodeRemoval)
	if err != nil {
		return nil, err
	}
	assemblyDiags, err := link.ReferencesAndDiagnostics(prog, srcMgr, o.parsePkg)
	if err != nil {
		return nil, err
	}

	// Properties a function sets on the part it returns are carried out to its
	// callers (property/carry.go), before the closure pass so the companions it
	// writes are transformed like any other function.
	if err = property.CarryReturnProperties(prog); err != nil {
		return nil, err
	}
	if err = link.References(prog, srcMgr, o.parsePkg); err != nil {
		return nil, err
	}
	rp.AttachVariants(prog, srcMgr, o.variants, o.parsePkg)

	// 2. Closure & Higher-Order Function Transformation
	// Hoist anonymous closures to package level and convert higher-order function patterns
	// to first-order representations, re-linking references afterwards.
	report("Transforming closures", 2)
	if err = closure.Transform(prog); err != nil {
		return nil, err
	}
	if err = link.References(prog, srcMgr, o.parsePkg); err != nil {
		return nil, err
	}

	// 3. Property Expansion & Initial Type Analysis
	// Hold each declaration-level annotation to its contract — @material to the
	// companion colour its library pairs it with — now that linking has brought
	// every dependency package in.
	if err = annotation.Validate(prog); err != nil {
		return nil, err
	}
	// Validate property assignment semantics, scope rules, and DAG cycles before expanding.
	report("Expanding properties & types", 3)
	if err = propval.Validate(prog); err != nil {
		return nil, err
	}
	// Expand property definitions and member accessors into explicit function calls/expressions,
	// re-link references, and perform initial type inference across all packages.
	prog, err = property.ExpandShared(prog)
	if err != nil {
		return nil, err
	}
	// Mark the document's parts for the viewer, drop declared boxes, and lower
	// .color and .material properties over the primary geometry returns only
	// (prog.Returns()[:geomCount]) so ColorReturns and MaterialReturns remain
	// 1-to-1 with geometry returns.
	prog, err = rp.WithGeometryReturnsOnly(prog, func(p *program.Program) (*program.Program, error) {
		p, err := property.MarkParts(p)
		if err != nil {
			return nil, err
		}
		property.DropDeclaredBoxes(p)
		if err := property.LowerReturnProperties(p); err != nil {
			return nil, err
		}
		return p, nil
	})
	if err != nil {
		return nil, err
	}
	if err = link.References(prog, srcMgr, o.parsePkg); err != nil {
		return nil, err
	}
	if err = types.Analyze(prog); err != nil {
		return nil, err
	}
	if err = rp.ValidateAudioTypes(prog); err != nil {
		return nil, err
	}
	// Now that types are resolved, hold an annotated pair to the material and
	// colour it promises. This is the check deferred when the annotations were
	// first recorded: it needs the returns typed, and only a zero-argument
	// function can be typed without a call site.
	if err = annotation.ValidateReturns(prog); err != nil {
		return nil, err
	}
	// A stand-in MarkParts could not give a part.
	if err = property.CheckSimple(prog); err != nil {
		return nil, err
	}

	// 4. Pre-Lowering Constant Evaluation & Structural AST Lowering
	// Fold constant expressions (such as array sizes, list concatenations, and arithmetic)
	// so structural lowering (e.g. map and reduce unrolling) can resolve static collections.
	report("Evaluating constants", 4)
	preFoldMod, err := eval.Fold(prog, newAssemblerParamChecker(prog.Primary))
	if err != nil {
		return nil, err
	}
	if preFoldMod {
		if err = link.References(prog, srcMgr, o.parsePkg); err != nil {
			return nil, err
		}
		if err = types.Analyze(prog); err != nil {
			return nil, err
		}
	}

	// Lower high-level constructs (such as map and reduce expressions) into lower-level AST nodes,
	// unroll collection operations wider than vec4 and lists into canonical Collect expressions,
	// then re-link symbols and re-run type analysis to populate types on newly generated nodes.
	report("Structural AST lowering", 4)
	if err = lower.Transform(prog); err != nil {
		return nil, err
	}
	if err = link.References(prog, srcMgr, o.parsePkg); err != nil {
		return nil, err
	}
	if err = types.Analyze(prog); err != nil {
		return nil, err
	}
	if _, err = unroll.Transform(prog); err != nil {
		return nil, err
	}
	if err = link.References(prog, srcMgr, o.parsePkg); err != nil {
		return nil, err
	}
	if err = types.Analyze(prog); err != nil {
		return nil, err
	}

	// 5. Domain-Specific Folding (Mesh Slots, Topology & Space Warps)
	// Execute optional pluggable domain folding passes for mesh asset slots, geometry/topology
	// resolution, and domain-remapping space warps before inlining so topology calls in
	// dependent packages are resolved with unambiguous local placements.
	report("Folding mesh & geometry slots", 5)
	var importSlots []meshslots.Import
	if o.enableSlotFold {
		var slotErr error
		// With the parameter checker, so the fold that resolves slot variables
		// leaves a slider live, as every other fold here does. Built for the
		// package the fold sees: the slot folding clones the primary first.
		prog, importSlots, slotErr = meshfold.FoldProgramWithSlotsChecked(prog, func(primary *program.Package) eval.ParameterChecker {
			return newAssemblerParamChecker(primary)
		})
		if slotErr != nil {
			return nil, slotErr
		}
	}

	if o.geometryResolver != nil {
		if si, ok := o.geometryResolver.(interface{ SetImportSlots([]meshslots.Import) }); ok {
			si.SetImportSlots(importSlots)
		}
		prog, err = internaltopofold.FoldProgram(prog, o.geometryResolver)
		if err != nil {
			return nil, err
		}
	}

	var warpedAssets []warpfold.WarpedAsset
	if o.warpResolver != nil {
		prog, warpedAssets, err = internalwarpfold.FoldProgram(prog, o.warpResolver)
		if err != nil {
			return nil, err
		}
	}
	if o.foldWarpFieldToZero {
		prog, err = internalwarpfold.FoldWarpFieldToZero(prog)
		if err != nil {
			return nil, err
		}
	}

	// 6. Iterative Fixed-Point Optimization
	// Iteratively run cross-package inlining, constant folding, and select branch pruning
	// until reaching a fixed point (no further AST mutations).
	report("Optimizing & inlining", 6)
	prog, err = optimizeProgram(prog, o, srcMgr)
	if err != nil {
		return nil, err
	}

	// 7. Final Dead Code Elimination
	// Sweep all packages to remove any functions, variables, or imports that became unused
	// following inlining, folding, and domain transformations.
	report("Eliminating dead code", 7)
	if o.enableDeadcodeRemoval {
		prog, err = deadcode.Remove(prog)
		if err != nil {
			return nil, err
		}
	}

	idPrefix := o.meshPlanIdPrefix
	if idPrefix == "" {
		idPrefix = "leaf-"
	}
	audioReturns, variants := rp.PopVariations(prog, idPrefix)

	sdfSlots, globalTex := textures.Extract(prog)
	// Invariant: prog is guaranteed non-nil following lowering and deadcode removal.
	prog.MeshSdfSlots = sdfSlots
	prog.GlobalTextures = globalTex

	result := &Result{
		Program:        prog,
		Slots:          importSlots,
		WarpedAssets:   warpedAssets,
		MeshSdfSlots:   sdfSlots,
		GlobalTextures: globalTex,
		Diagnostics:    assemblyDiags,
		AudioReturns:   audioReturns,
		Variants:       variants,
	}

	if o.extractMeshPlan {
		// The returns are the last primary file's, as the wgsl render names
		// them: the plan names its bodies from the same text.
		var document string
		if n := len(primarySources); n > 0 {
			document = primarySources[n-1]
		}
		plan, subtrees, planErr := meshplan.SliceDocument(prog, idPrefix, document)
		if planErr != nil {
			return nil, planErr
		}
		result.MeshPlan = plan
		result.SdfSubtrees = subtrees
		if plan.Root != nil && len(plan.MeshDecls()) > 0 {
			result.PreviewBodies = meshplan.PreviewBodies(plan, subtrees)
		}
	}

	// Invariant: prog and prog.Primary are guaranteed non-nil following lowering and deadcode removal.
	meshBboxes, decls := resolveAssemblerSpinodoidInputs(prog, importSlots, warpedAssets, sdfSlots, result.MeshPlan, o.geometryResolver)
	if err := meshplan.SetProgramSpinodoidBounds(prog, meshBboxes, decls); err != nil {
		return nil, err
	}

	if o.extractUI {
		sourceText := strings.Join(primarySources, "\n")
		result.Sliders = internalslider.Extract(sourceText, prog.Primary)
		result.UIControls = internalui.Extract(prog.Primary, prog.GetImport)
		result.Scene = prog.Scene()
	}

	return result, nil
}

func resolveAssemblerSpinodoidInputs(
	prog *program.Program,
	importSlots []meshslots.Import,
	warpedAssets []warpfold.WarpedAsset,
	sdfSlots []*program.MeshSdfSlot,
	plan *meshplan.MeshExportPlan,
	geomResolver topofold.GeometryResolver,
) (map[int][6]float64, []meshplan.MeshDecl) {
	var decls []meshplan.MeshDecl
	if plan != nil && plan.Root != nil {
		decls = plan.MeshDecls()
	}
	if geomResolver == nil {
		return nil, decls
	}
	if plan == nil {
		if sliced, _, err := meshplan.SliceProgram(prog, "spinodoid-bounds::"); err == nil && sliced != nil && sliced.Root != nil {
			decls = sliced.MeshDecls()
		}
	}

	candidates := make(map[int]struct{})
	for _, imp := range importSlots {
		candidates[imp.Slot] = struct{}{}
	}
	for _, wa := range warpedAssets {
		candidates[wa.MeshSlot] = struct{}{}
		candidates[wa.DerivedSlot] = struct{}{}
	}
	for _, s := range sdfSlots {
		candidates[int(s.AssetIndex)] = struct{}{}
	}
	for _, d := range decls {
		candidates[d.Index] = struct{}{}
	}
	// Invariant: prog.Primary and prog.Dependencies entries are guaranteed non-nil.
	collectPkgImportIndices := func(pkg *program.Package) {
		for _, f := range pkg.Files {
			for _, mi := range f.MeshImports {
				candidates[int(mi.Index)] = struct{}{}
			}
			for _, si := range f.StepImports {
				candidates[int(si.Index)] = struct{}{}
			}
		}
	}
	collectPkgImportIndices(prog.Primary)
	for _, dep := range prog.Dependencies {
		collectPkgImportIndices(dep)
	}

	var meshBboxes map[int][6]float64
	for slot := range candidates {
		mg, ok := geomResolver.GetMesh(slot)
		if !ok || mg == nil {
			continue
		}
		bm, ok := mg.(interface {
			Bounds() ([3]float64, [3]float64)
		})
		if !ok {
			continue
		}
		if meshBboxes == nil {
			meshBboxes = make(map[int][6]float64)
		}
		minPt, maxPt := bm.Bounds()
		meshBboxes[slot] = [6]float64{minPt[0], minPt[1], minPt[2], maxPt[0], maxPt[1], maxPt[2]}
	}
	return meshBboxes, decls
}

func optimizeProgram(prog *program.Program, o *options, srcMgr internalsource.Manager) (*program.Program, error) {
	const maxIterations = 1000
	for i := 0; i < maxIterations; i++ {
		if o.progressFn != nil {
			o.progressFn(fmt.Sprintf("Optimizing & inlining (pass %d)", i+1), 6, 7)
		}
		modifiedAny := false

		// 1. Inlining pass: inline functions with constant arguments, binding
		// a constant number as its literal, so it knows the parameters too
		modInline, err := inline.TransformWithParameters(prog, newAssemblerParamChecker(prog.Primary))
		if err != nil {
			return nil, err
		}
		if modInline {
			modifiedAny = true
		}

		// 2. Re-link and type-check before constant folding
		if err = link.References(prog, srcMgr, o.parsePkg); err != nil {
			return nil, err
		}
		if err = types.Analyze(prog); err != nil {
			return nil, err
		}

		// 3. Constant folding pass: fold expressions and prune dead select branches
		modFold, err := eval.Fold(prog, newAssemblerParamChecker(prog.Primary))
		if err != nil {
			return nil, err
		}
		if modFold {
			modifiedAny = true
		}

		// 3b. Specialize calls on the constants that decide a callee's selects,
		// which the fold just stamped, so the next fold prunes inside the clones
		modSpec, err := constspec.Transform(prog)
		if err != nil {
			return nil, err
		}
		if modSpec {
			modifiedAny = true
		}

		// 4. Re-link and type-check before lowering
		if err = link.References(prog, srcMgr, o.parsePkg); err != nil {
			return nil, err
		}
		if err = types.Analyze(prog); err != nil {
			return nil, err
		}

		// 5. Structural AST lowering: lower newly introduced map/reduce or collection expressions
		if err = lower.Transform(prog); err != nil {
			return nil, err
		}
		if err = link.References(prog, srcMgr, o.parsePkg); err != nil {
			return nil, err
		}
		if err = types.Analyze(prog); err != nil {
			return nil, err
		}
		modUnroll, err := unroll.Transform(prog)
		if err != nil {
			return nil, err
		}
		if modUnroll {
			modifiedAny = true
		}
		if err = link.References(prog, srcMgr, o.parsePkg); err != nil {
			return nil, err
		}
		if err = types.Analyze(prog); err != nil {
			return nil, err
		}

		// 6. Dead code elimination: prune unused declarations between rounds
		if o.enableDeadcodeRemoval {
			prog, err = deadcode.Remove(prog)
			if err != nil {
				return nil, err
			}
		}

		// Fixed-point reached: stop if no pass modified the AST
		if !modifiedAny {
			break
		}
	}
	return prog, nil
}

type assemblerParamChecker struct {
	// Read-only immutable fields set at construction
	sliderVars    map[string]bool
	paramLiterals map[*program.Operand]bool
}

func newAssemblerParamChecker(primary *program.Package) *assemblerParamChecker {
	c := &assemblerParamChecker{
		sliderVars:    program.ComputeSliderVars(primary),
		paramLiterals: make(map[*program.Operand]bool),
	}
	if primary == nil {
		return c
	}
	var markLiterals func(op *program.Operand)
	markLiterals = func(op *program.Operand) {
		if op == nil {
			return
		}
		switch t := op.T.(type) {
		case *program.Operand_Value:
			c.paramLiterals[op] = true
		case *program.Operand_Expression:
			for _, child := range t.Expression.GetOperands() {
				markLiterals(child)
			}
		case *program.Operand_FnCall:
			for _, arg := range t.FnCall.GetArguments() {
				markLiterals(arg)
			}
		}
	}
	var findMarkers func(op *program.Operand)
	findMarkers = func(op *program.Operand) {
		if op == nil {
			return
		}
		switch t := op.T.(type) {
		case *program.Operand_Expression:
			for _, child := range t.Expression.GetOperands() {
				findMarkers(child)
			}
		case *program.Operand_FnCall:
			if ref := t.FnCall.GetReference(); ref != nil {
				switch ref.Name {
				case "uiPoint", "uiTransform", "uiSpline", "uiColor", "uiDirection":
					for _, arg := range t.FnCall.GetArguments() {
						markLiterals(arg)
					}
					return
				}
			}
			for _, arg := range t.FnCall.GetArguments() {
				findMarkers(arg)
			}
		}
	}
	var findInFnType func(ft *program.FnType)
	findInFnType = func(ft *program.FnType) {
		for _, g := range ft.Groups {
			for _, fn := range g.Functions {
				for _, a := range fn.Assignments {
					findMarkers(a.Expression)
				}
				for _, ret := range fn.Returns {
					findMarkers(ret)
				}
				for _, nested := range fn.FnTypes {
					findInFnType(nested)
				}
			}
		}
	}
	for _, f := range primary.Files {
		for _, a := range f.Assignments {
			if a.Range != nil {
				markLiterals(a.Expression)
			}
			findMarkers(a.Expression)
			for _, v := range a.Variables {
				for _, prop := range v.Properties {
					findMarkers(prop.Expression)
				}
			}
		}
		for _, ft := range f.FnTypes {
			findInFnType(ft)
		}
		for _, ret := range f.Returns {
			findMarkers(ret)
		}
	}
	return c
}

func (c *assemblerParamChecker) IsParameter(op *program.Operand) bool {
	if c.paramLiterals[op] {
		return true
	}
	if ref, ok := op.T.(*program.Operand_Reference); ok && ref != nil && ref.Reference != nil {
		return c.sliderVars[ref.Reference.Name]
	}
	return false
}

// AnalyzePackage parses source into an un-inlined *program.Package and stamps
// evaluated constant float64 values onto top-level assignments, local function
// expressions/returns, and call arguments while preserving the original AST
// structure for template comparisons and bounds extraction.
func AnalyzePackage(sourceText string, getImport program.GetImportFunc) (*program.Package, error) {
	res, err := parser.ParsePackage([]string{sourceText})
	if err != nil {
		return nil, err
	}
	pkg := res.Package
	stampPackageConstants(pkg, getImport)
	return pkg, nil
}

func stampPackageConstants(pkg *program.Package, getImport program.GetImportFunc) {
	// Invariant: pkg and its parsed AST slices (Files, Assignments, FnTypes, Groups, Functions)
	// are guaranteed non-nil by parser.ParsePackage in AnalyzePackage.
	var stampOperand func(op *program.Operand, scope *program.Package)
	var stampFnType func(ft *program.FnType, scope *program.Package)

	stampOperand = func(op *program.Operand, scope *program.Package) {
		if op == nil {
			return
		}
		if len(op.ConstValues) == 0 {
			if val, err := eval.ConstPkg(op, scope, getImport); err == nil && len(val) > 0 {
				op.ConstValues = val
			}
		}
		switch t := op.T.(type) {
		case *program.Operand_Expression:
			expr := t.Expression
			if expr == nil {
				return
			}
			operands := expr.Operands
			switch expr.Operation {
			case program.Operation_ACCESS:
				operands = operands[:min(1, len(operands))]
			case program.Operation_MAP:
				if len(operands) > 1 {
					operands = operands[1:]
				}
			}
			for _, child := range operands {
				stampOperand(child, scope)
			}
		case *program.Operand_FnCall:
			if t.FnCall == nil {
				return
			}
			for _, arg := range t.FnCall.Arguments {
				stampOperand(arg, scope)
			}
		}
	}

	stampFnType = func(ft *program.FnType, scope *program.Package) {
		for _, g := range ft.Groups {
			for _, fn := range g.Functions {
				local := &program.Package{
					Path:  scope.Path,
					Files: append([]*program.PackageFile{{Assignments: fn.Assignments, FnTypes: fn.FnTypes}}, scope.Files...),
				}
				for _, a := range fn.Assignments {
					stampOperand(a.Expression, local)
				}
				for _, r := range fn.Returns {
					stampOperand(r, local)
				}
				for _, nested := range fn.FnTypes {
					stampFnType(nested, local)
				}
			}
		}
	}

	for _, f := range pkg.Files {
		for _, a := range f.Assignments {
			stampOperand(a.Expression, pkg)
		}
		for _, ft := range f.FnTypes {
			stampFnType(ft, pkg)
		}
		for _, r := range f.Returns {
			stampOperand(r, pkg)
		}
	}
}
