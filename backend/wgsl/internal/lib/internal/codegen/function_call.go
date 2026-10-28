package codegen

import (
	"fmt"

	"gradientcontrol.com/omega/core/backend/wgsl/internal/lib/internal/accel"
	"gradientcontrol.com/omega/core/backend/wgsl/internal/lib/internal/wgsl"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// nestedMeshDistanceWGSL stands in for a mesh() nested inside an expression,
// where the raymarch shader can't evaluate it: effectively +infinity, so a
// min or max against it collapses to the other operand.
const nestedMeshDistanceWGSL = "1e30"

func (d *Delegate) BuildFuncCall(
	fnCallOp *program.Operand, _ *program.FnCall,
	fnInfo build.FnInfo, expressions []build.Expression,
) (build.Expression, error) {
	info := fnInfo.(*wgsl.FnInfo)
	op := fnCallOp
	if info.ReturnType() != nil {
		if op == nil {
			op = &program.Operand{Type: info.ReturnType()}
		} else {
			op = &program.Operand{Start: op.Start, End: op.End, Type: info.ReturnType(), T: op.T}
		}
	}
	return d.buildFuncCall(
		op, info.WgslName(),
		info.GlobalRefs(), expressions, info.WGSLReturnType())
}

func (d *Delegate) buildFuncCall(
	fnCallOp *program.Operand,
	wgslName string,
	globalRefs []*wgsl.VarInfo,
	expressions []build.Expression,
	returnType wgsl.TypeInfo,
) (build.Expression, error) {
	if expr, ok, err := accel.BuildCall(fnCallOp, wgslName, expressions); ok {
		return expr, err
	}

	if vanishes(wgslName, len(expressions)) && len(globalRefs) == 0 {
		parts := make([]wgsl.Expression, len(expressions))
		for i, e := range expressions {
			parts[i] = e.(wgsl.Expression)
		}
		return d.vanishGuard(fnCallOp, parts, returnType), nil
	}
	// A stand-in's switches read only their four arguments, so a call that
	// also carries module state (the noise tile a noisy part needs) is
	// switched as well.
	if simplifies(wgslName, len(expressions)) {
		args := make([]wgsl.Expression, len(expressions))
		for i, e := range expressions {
			args[i] = e.(wgsl.Expression)
		}
		return d.simpleCall(fnCallOp, args, returnType), nil
	}
	if simpleValues(wgslName, len(expressions)) {
		args := make([]wgsl.Expression, len(expressions))
		for i, e := range expressions {
			args[i] = e.(wgsl.Expression)
		}
		return d.simpleValueGuard(fnCallOp, args, returnType), nil
	}
	if detailValues(wgslName, len(expressions)) {
		args := make([]wgsl.Expression, len(expressions))
		for i, e := range expressions {
			args[i] = e.(wgsl.Expression)
		}
		return d.detailGuard(fnCallOp, args, returnType), nil
	}

	args := make([]wgsl.Expression, 0, len(expressions))

	pVarIsFirstArg := false

	// Add global variables as arguments in the order they were referenced
	if len(globalRefs) > 0 {

		// Check for the current p var if it is the first global ref
		if globalRefs[0].Name() == program.PVarName {
			pVarIsFirstArg = true
		}

		for _, ref := range globalRefs {
			if ref.Type().IsFunction() {
				continue
			}
			// An inlined call's parameter is no variable: the body it is
			// inlined into reads its argument in its place, and so does a
			// fold's step function that reads it (fold_loop.go).
			if inlinedParam(ref) {
				args = append(args, ref.Expr().(wgsl.Expression))
				continue
			}
			args = append(args, wgsl.NewVarRef(nil, ref))
			d.recordRef(ref)
		}
	}

	// Add function arguments
	for _, arg := range expressions {
		if arg.Type().IsFunction() {
			continue
		}
		expr := arg.(wgsl.Expression)
		args = append(args, expr)
	}

	// Forward the bvhSdf/vdbSdf slot uniforms to every called helper.
	// Mirrors the suffix buildFunc appends to function signatures, so
	// callees that inline vdbSdf code (or call other helpers that do)
	// always have sdfTex<N>/sdfSmp<N>/sdfBMin<N>/sdfBSize<N> in scope.
	// When the program references no bvhSdf/vdbSdf, this is a no-op.
	if d.meshSdfSlots != nil {
		for _, slot := range d.meshSdfSlots.Ordered() {
			s := slot.Slot
			args = append(args,
				slotForwardArg(fmt.Sprintf("sdfTex%d", s)),
				slotForwardArg(fmt.Sprintf("sdfSmp%d", s)),
				slotForwardArg(fmt.Sprintf("sdfBMin%d", s)),
				slotForwardArg(fmt.Sprintf("sdfBSize%d", s)),
			)
		}
	}
	// Forward the noise tile to every callee, mirroring the signature suffix.
	if d.warpUsed {
		for _, n := range []string{warpTexName, warpSmpName, warpBMinName, warpBSizeName, warpRangeName, warpDimsName} {
			args = append(args, slotForwardArg(n))
		}
	}
	if d.noiseUsed {
		args = append(args, slotForwardArg(noiseTexName))
	}
	if d.spinodoidUsed {
		args = append(args, slotForwardArg(spinodoidTexName))
	}

	return wgsl.NewFuncCall(fnCallOp, wgslName, args, pVarIsFirstArg, returnType)
}

func (d *Delegate) ShouldGenerate(ref *wgsl.VarInfo) bool {
	bScope := d.GetScope()
	// Always generate the variable if it belongs to the current compiling scope,
	// or if we are compiling at the package/file level (primary or file scope),
	// which ensures that top-level constants and helper function variables are emitted.
	if ref.Scope() == bScope || bScope.IsPrimaryScope() || bScope.IsFileScope() {
		return true
	}
	// For variables defined in module/file scope, only generate them if the current
	// codegen scope is also at the module/file level. When generating inside local
	// function bodies, we do not redeclare them to avoid duplicate symbol errors
	// (since they are already globally visible in WGSL).
	if ref.Scope().IsTopLevelScope() || ref.Scope().IsFileScope() {
		return bScope.IsFileScope() || bScope.IsTopLevelScope()
	}
	// An inlined call's locals belong to no function of their own; emit their
	// `let` into whichever function body is currently being generated (the one
	// that references them). Without this, inlining inside a non-primary
	// function drops the binding and leaves the reference unresolved.
	//
	// A fold's step function is the exception: what the fold's scope and those
	// around it define reaches the step as its parameters (fold_loop.go).
	if ref.Scope().IsInlinedScope() {
		return !d.definedAroundFold(ref.Scope())
	}
	return false
}

type CellularScanExpr struct {
	wgsl.Expression
	Q       wgsl.Expression
	Octaves int
	Field   string
}

func NewCellularScanExpr(op *program.Operand, q wgsl.Expression, octaves int, field string, outType wgsl.TypeInfo) *CellularScanExpr {
	c := &CellularScanExpr{
		Q:       q,
		Octaves: octaves,
		Field:   field,
	}
	reads := append([]string{}, q.Reads()...)
	reads = append(reads, noiseTexName)
	c.Expression = wgsl.NewExprWithReads(op, outType, build.PrecedenceHighest, reads...)
	c.Expression.SetOuter(c)
	return c
}

func (c *CellularScanExpr) Codegen(ctx wgsl.CodegenContext) string {
	return fmt.Sprintf("gsl_cellularScan(%s, %d, %s).%s",
		c.Q.Codegen(ctx), c.Octaves, noiseTexName, c.Field)
}

type NoiseHash3Expr struct {
	wgsl.Expression
	Cell    wgsl.Expression
	Octaves int
}

func NewNoiseHash3Expr(op *program.Operand, cell wgsl.Expression, octaves int) *NoiseHash3Expr {
	n := &NoiseHash3Expr{
		Cell:    cell,
		Octaves: octaves,
	}
	reads := append([]string{}, cell.Reads()...)
	reads = append(reads, noiseTexName)
	n.Expression = wgsl.NewExprWithReads(op, wgsl.TypeVec3, build.PrecedenceHighest, reads...)
	n.Expression.SetOuter(n)
	return n
}

func (n *NoiseHash3Expr) Codegen(ctx wgsl.CodegenContext) string {
	c := n.Cell.Codegen(ctx)
	if n.Octaves == 1 {
		return fmt.Sprintf(
			"fract(textureLoad(%[1]s, vec3<i32>(floor(%[2]s)) & vec3<i32>(%[3]d), 0).rgb)",
			noiseTexName, c, tileWrapMask)
	}
	return fmt.Sprintf("fract("+
		"textureLoad(%[1]s, vec3<i32>(floor(%[2]s)) & vec3<i32>(%[3]d), 0).rgb + "+
		"textureLoad(%[1]s, vec3<i32>(floor((%[2]s) / %[4]d.0)) & vec3<i32>(%[3]d), 0).rgb)",
		noiseTexName, c, tileWrapMask, tileSpan)
}

// inlinedParam reports whether ref is a parameter of an inlined call, which
// stands for the expression its argument is.
func inlinedParam(ref *wgsl.VarInfo) bool {
	if ref.Scope() == nil || !ref.Scope().IsInlinedScope() || ref.CanRederive() {
		return false
	}
	_, ok := ref.Expr().(wgsl.Expression)
	return ok
}

// slotForwardArg builds a no-type wgsl.Expression that emits the literal
// `name` at the call site. The receiving function's signature will have
// the matching param (added by buildFunc), so name resolution works.
func slotForwardArg(name string) wgsl.Expression {
	return wgsl.NewOpaqueExpr(nil, wgsl.TypeF32, build.PrecedenceHighest, name, name)
}

func (d *Delegate) BuildBuiltinCall(
	op *program.Operand, fnCall *program.FnCall,
	builtin *program.BuiltinType, expressions []build.Expression,
) (build.Expression, error) {

	// parser should have already validated the number of arguments

	// bvhSdf / vdbSdf are not WGSL-native builtins; they sample a pre-baked
	// 3D distance texture supplied by the canvas agent. Route through their
	// shared emitter so the call site lowers to a textureSampleLevel expression.
	// vdbSdf rides the same preview path — both lower to a sample of the
	// dense Go-side SdfGrid. The OpenVDB-backed oracle only kicks in for
	// the fidget mesh-out path; the WGSL preview just samples the dense
	// grid the canvas extracts from the same OpenVDB level set, giving
	// users a coherent preview-vs-export experience.
	//
	// `mesh(idx)` only reaches this dispatcher when it appears INSIDE an
	// expression (top-level `mesh()` returns are stripped by canvas's
	// extractMeshDecls). Nested mesh() calls emit a `1e30` sentinel below
	// so SDF math collapses to the other operand and the rasterized mesh
	// shows up via the existing depth compositor.
	if builtin.Name == "bvhSdf" || builtin.Name == "vdbSdf" {
		return d.BuildMeshSdfCall(op, fnCall, expressions)
	}
	if program.IsGaborSpinodoid(builtin.Name) {
		return d.BuildSpinodoidCall(op, fnCall, expressions)
	}
	if builtin.Name == "mesh" {
		return wgsl.NewOpaqueExpr(op, wgsl.TypeF32, build.PrecedenceHighest, nestedMeshDistanceWGSL), nil
	}
	// meshDifference has no direct WGSL alias — lower to max(L, -R) by
	// constructing a synthetic negated argument, then dispatching as
	// `meshIntersect(L, -R)` (which is the GSL builtin already wired
	// to `max` via wgslBuiltins). Dispatching with the literal name
	// "max" doesn't work because wgslBuiltins is keyed by *GSL*
	// builtin names, not WGSL output names — that lookup misses and
	// surfaces as a "builtin max not found" error at compile time.
	// The build front-end has already folded any list operand into a
	// chain of binary meshDifference calls, so this only ever sees two.
	if builtin.Name == "meshDifference" {
		if len(expressions) != 2 {
			return nil, fmt.Errorf("meshDifference: expected 2 args, got %d", len(expressions))
		}
		left := expressions[0].(wgsl.Expression)
		right := expressions[1].(wgsl.Expression)
		negRight := wgsl.NewUnaryOp(op, "-", wgsl.Paren(right), right.WGSLType(), build.PrecedenceUnary)
		return wgsl.NewBuiltinCall(op, "meshIntersect", []build.Expression{left, negRight})
	}

	// noiseHash3(cell) -> vec3 in [0,1)^3: a nearest lookup into the baked noise
	// tile plus one supercell fold, matching common/noise.Hash3 and the Fidget
	// oracle bit-for-bit (rgba8unorm textureLoad returns channels as byte/255).
	// noiseTex is threaded in as a function parameter (see noise_tex.go).
	// warpField(p) -> vec3: the simulated displacement at a world position,
	// sampled from the baked grid (see warp_tex.go).
	if builtin.Name == "warpField" {
		return d.buildWarpFieldCall(op, expressions)
	}

	// pixelSize() -> f32: the host's per-invocation pixel footprint, a private
	// var every viewer shader declares (gsl_pixelWidthNode) and sets at a hit
	// before shading; 0 in shaders that never set it (meshing, bounds).
	if builtin.Name == program.PixelSize {
		return wgsl.NewOpaqueExpr(op, wgsl.TypeF32, build.PrecedenceHighest, "gsl_pw", "gsl_pw"), nil
	}

	// displacementDrawn() -> f32: how much of the displacement the viewer
	// draws at the hit, a private var the codegen declares 0 (dispDrawnDeclWGSL)
	// and a displacing fragment sets before shading (gsl-render-graph.js).
	if builtin.Name == program.DisplacementDrawn {
		d.dispDrawnUsed = true
		return wgsl.NewOpaqueExpr(op, wgsl.TypeF32, build.PrecedenceHighest, "gsl_dispDrawn", "gsl_dispDrawn"), nil
	}

	// The worley* builtins: the 27-neighbour scan, emitted as a helper with a
	// real WGSL `for` loop. The GSL build layer unrolls reduce at compile time,
	// so written as a library fold this became 27 inlined copies at every call
	// site — which is why a textured frame's Metal compile takes seconds (see
	// gpuparity's imageSkipExamples). As a builtin the shader gets one small loop
	// instead, and the dynamic work is unchanged. One scan serves every field, so
	// each builtin just reads its own member of the returned struct.
	if f, ok := cellularFields[builtin.Name]; ok {
		octaves, err := program.NoiseHash3Octaves(fnCall)
		if err != nil {
			return nil, err
		}
		q := expressions[0].(wgsl.Expression)
		var outType wgsl.TypeInfo = wgsl.TypeF32
		if f.vec3 {
			outType = wgsl.TypeVec3
		}
		d.cellularUsed = true
		return NewCellularScanExpr(op, q, octaves, f.field, outType), nil
	}

	// noiseHash3(cell) -> vec3 in [0,1)^3: a nearest lookup into the baked noise
	// tile, optionally folded with one supercell lookup, matching
	// common/noise.Hash3Octaves and the Fidget oracle bit-for-bit (rgba8unorm
	// textureLoad returns channels as byte/255). noiseTex is threaded in as a
	// function parameter (see noise_tex.go).
	//
	// The tile wrap is a mask, not `((v % tileSpan) + tileSpan) % tileSpan`: the
	// tile is a power of two per axis, and on two's-complement i32 the mask IS
	// the floored modulus, for negative coordinates too, so the two forms agree
	// exactly. The mask is not something the shader compiler can derive on its
	// own — signed `%` by a power of two truncates toward zero, so it cannot see
	// through the sign-correcting pair.
	if builtin.Name == program.NoiseHash3 {
		octaves, err := program.NoiseHash3Octaves(fnCall)
		if err != nil {
			return nil, err
		}
		cell := expressions[0].(wgsl.Expression)
		return NewNoiseHash3Expr(op, cell, octaves), nil
	}

	// dot of two scalars is their product — WGSL's dot() only takes vectors, so
	// lower a scalar dot to multiplication. A scalar reaches here when a lattice
	// or segment helper is called with a scalar cell/size (e.g. lineSegment's
	// dot(ba, ba) where ba came from a scalar cell); the constant case already
	// folds in the build layer, so this covers the runtime case.
	if builtin.Name == "dot" && len(expressions) == 2 &&
		expressions[0].(wgsl.Expression).WGSLType().Equal(wgsl.TypeF32) &&
		expressions[1].(wgsl.Expression).WGSLType().Equal(wgsl.TypeF32) {
		return wgsl.NewNumericOp(op, expressions, build.PrecedenceMultiplication, op.Type, "*")
	}

	return wgsl.NewBuiltinCall(op, fnCall.Reference.Name, expressions)
}
