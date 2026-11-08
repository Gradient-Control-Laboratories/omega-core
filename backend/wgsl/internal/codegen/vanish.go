package codegen

import (
	"fmt"

	"gradientcontrol.com/omega/core/backend/wgsl/internal/ir"
	"gradientcontrol.com/omega/core/backend/wgsl/internal/wgsl"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// Small parts shrink to nothing in the viewer. The assembler wraps each part of
// a document, a named value that only a min or a return reads and whose box is
// bounded, in builtin's shrinkToVanish(shape, lo, hi, solid), whose GSL body is
// the shape: fidget, the mesher and export keep every part whole. The codegen
// writes the viewer's guard in place of the call instead:
//
//	let w = gsl_vanish(lo, hi);            // how large the box looks on screen
//	var part = 1e10;                        // gone
//	if (w.x > 0.0) { part = shape + w.y; }  // whole, or shrunk by w.y
//	if (abs(part) < gsl_partNear.x) { gsl_partNear = vec3(abs(part), w.zw); }
//
// The weight depends only on the camera and the box, so every pixel takes the
// same branch and a part too small to see costs nothing. Shrinking offsets the
// field by a constant, which keeps it a distance bound and leaves its normals
// alone, and it reaches half the box's smallest extent as the weight reaches 0,
// so the part has vanished by the time it is skipped.
//
// The camera arrives through two private variables the codegen declares after
// sdShape (never first: the viewers hand the module to three.js's wgslFn, which
// parses it as a function). A host that does not set them leaves them zero,
// which keeps every part whole, so the mesher, the probes and the parity tests
// see the program as written.
//
// A part whose solid is in its box (solid 1) is also skipped wherever the point
// is far from it. Its distance is then the distance to its box, which is never
// more than the distance to the part, so a march can step by it, and the part
// itself is evaluated only within gsl_vanishView.w of its box:
//
//	let box = gsl_boxDistance(p, lo, hi);
//	var part = box;                         // far: the box
//	if (gsl_vanishView.w <= 0.0 || box <= gsl_vanishView.w) { ...as above... }
//
// A surface is inside its part's box, so every hit is on the exact field, and
// the margin keeps a box's own face from reading as one. But a shadow's
// softness and ambient occlusion read the field's size, not only its zero, and
// a box is nearer than its part, so the host sets the margin only while it
// marches and leaves it 0, which turns the skip off, everywhere else. A part
// whose box holds only its surface, a declared box's crop, may be solid
// outside it, and is never skipped.
//
// Each guard also notes the part nearest the point sdShape was asked about, in
// gsl_partNear: its distance, how large it is on screen in pixels and its
// shortest edge in world units. sdShape starts it far away. The host reads it
// after a call, to turn off a small part's displacement and fade its textures
// by the same measure that shrinks it away. A group of parts is noted after the
// parts in it and at the same distance, so the nearest part itself is kept.

// shrinkToVanishWGSLName is the name codegen gives builtin's shrinkToVanish.
const shrinkToVanishWGSLName = "builtin_shrinkToVanish"

// vanishedWGSL is a part's distance once it has shrunk away: the empty
// return's (emptyShapeWGSL), far from everything.
const vanishedWGSL = "1e10"

// vanishes reports whether a function is builtin's shrinkToVanish(shape, lo, hi,
// solid).
func vanishes(wgslName string, argCount int) bool {
	return wgslName == shrinkToVanishWGSLName && argCount == 4
}

// solidPart reports whether a shrinkToVanish call says its part's solid is in
// its box: its last argument, which the assembler writes as a literal.
func solidPart(op *program.Operand) bool {
	fc := op.GetFnCall()
	return fc != nil && len(fc.Arguments) == 4 && fc.Arguments[3].GetValue() == 1
}

type VanishGuardExpr struct {
	wgsl.Expression
	d     *Delegate
	shape wgsl.Expression
	lo    wgsl.Expression
	hi    wgsl.Expression
	solid bool
}

func NewVanishGuardExpr(op *program.Operand, d *Delegate, shape, lo, hi wgsl.Expression, returnType wgsl.TypeInfo, solid bool) *VanishGuardExpr {
	v := &VanishGuardExpr{
		d:     d,
		shape: shape,
		lo:    lo,
		hi:    hi,
		solid: solid,
	}
	v.Expression = wgsl.NewExpr(op, returnType, build.PrecedenceHighest)
	v.Expression.SetOuter(v)
	return v
}

func (v *VanishGuardExpr) Codegen(ctx wgsl.CodegenContext) string {
	d := v.d
	weight := d.uniqueLocalName("gsl_partVanish")
	part := d.uniqueLocalName("gsl_part")
	loSrc, hiSrc := v.lo.Codegen(ctx), v.hi.Codegen(ctx)
	var sw *simpleSwitch

	if v.solid {
		boxLo := d.uniqueLocalName("gsl_partLo")
		boxHi := d.uniqueLocalName("gsl_partHi")
		box := d.uniqueLocalName("gsl_partBox")

		block := d.currentScope().Block()
		block.Add(ir.NewLetStmt(boxLo, "", loSrc, v.lo.Reads()))
		block.Add(ir.NewLetStmt(boxHi, "", hiSrc, v.hi.Reads()))
		boxDist := fmt.Sprintf("gsl_boxDistance(%s, %s, %s)", ctx.P().WGSLName(), boxLo, boxHi)
		block.Add(ir.NewLetStmt(box, "", boxDist, []string{ctx.P().WGSLName(), boxLo, boxHi}))
		block.Add(ir.NewVarStmt(part, "", box, []string{box}))

		var shapeSrc string
		shapeSrc, sw = d.shapeInGuard(v.shape, ctx)

		innerBlock := ir.NewBlock()
		vanishCall := fmt.Sprintf("gsl_vanish(%s, %s)", boxLo, boxHi)
		innerBlock.Add(ir.NewLetStmt(weight, "", vanishCall, []string{boxLo, boxHi}))
		innerBlock.Add(ir.NewAssignStmt(part, vanishedWGSL, nil))
		if sw != nil {
			sw.emitTo(innerBlock)
		}
		assignExpr := fmt.Sprintf("%s + %s.y", shapeSrc, weight)
		assignBody := ir.NewBlock()
		assignBody.Add(ir.NewAssignStmt(part, assignExpr, append(v.shape.Reads(), weight)))
		innerBlock.Add(ir.NewInlineGuardStmt(fmt.Sprintf("%s.x > 0.0", weight), assignBody, []string{weight}))

		partNearStmt := fmt.Sprintf("if (abs(%[2]s) < %[3]s.x) { %[3]s = vec3<f32>(abs(%[2]s), %[1]s.zw); }", weight, part, partNearWGSL)
		innerBlock.Add(ir.NewOpaqueStmt([]string{partNearStmt}, []string{weight, part, partNearWGSL}, []string{partNearWGSL}))

		guardCond := fmt.Sprintf("gsl_vanishView.w <= 0.0 || %s <= gsl_vanishView.w", box)
		block.Add(ir.NewGuardStmt(guardCond, innerBlock, []string{"gsl_vanishView", box}))

		return part
	}

	block := d.currentScope().Block()
	vanishCall := fmt.Sprintf("gsl_vanish(%s, %s)", loSrc, hiSrc)
	block.Add(ir.NewLetStmt(weight, "", vanishCall, append(append([]string{}, v.lo.Reads()...), v.hi.Reads()...)))
	block.Add(ir.NewVarStmt(part, "", vanishedWGSL, nil))

	var shapeSrc string
	shapeSrc, sw = d.shapeInGuard(v.shape, ctx)

	if sw != nil {
		sw.emitTo(block)
	}
	assignExpr := fmt.Sprintf("%s + %s.y", shapeSrc, weight)
	assignBody := ir.NewBlock()
	assignBody.Add(ir.NewAssignStmt(part, assignExpr, append(v.shape.Reads(), weight)))
	block.Add(ir.NewInlineGuardStmt(fmt.Sprintf("%s.x > 0.0", weight), assignBody, []string{weight}))

	partNearStmt := fmt.Sprintf("if (abs(%[2]s) < %[3]s.x) { %[3]s = vec3<f32>(abs(%[2]s), %[1]s.zw); }", weight, part, partNearWGSL)
	block.Add(ir.NewOpaqueStmt([]string{partNearStmt}, []string{weight, part, partNearWGSL}, []string{partNearWGSL}))

	return part
}

// vanishGuard is a call to shrinkToVanish: the guard's statements, written into
// the function being generated, and the guarded part as its value. The shape
// is generated before the if, so anything it depends on is declared outside
// it; only the shape's own expression, the part's call, is skipped.
func (d *Delegate) vanishGuard(op *program.Operand, args []wgsl.Expression, returnType wgsl.TypeInfo) build.Expression {
	d.vanishUsed = true
	d.notePart(op)
	shape, lo, hi := args[0], args[1], args[2]
	return NewVanishGuardExpr(op, d, shape, lo, hi, returnType, solidPart(op))
}

// notePart records a shrinkToVanish call's box for the viewer (api.Part),
// when both its corners are constant.
func (d *Delegate) notePart(op *program.Operand) {
	fc := op.GetFnCall()
	if fc == nil || len(fc.Arguments) != 4 {
		return
	}
	lo, okLo := constVec3(fc.Arguments[1])
	hi, okHi := constVec3(fc.Arguments[2])
	if okLo && okHi {
		d.parts = append(d.parts, [2][3]float64{lo, hi})
	}
}

// constVec3 is a vec3 operand's value, when it is a constant: folded, or a
// collect of three numbers.
func constVec3(op *program.Operand) ([3]float64, bool) {
	var v [3]float64
	if len(op.GetConstValues()) == 3 {
		copy(v[:], op.GetConstValues())
		return v, true
	}
	e := op.GetExpression()
	if e == nil || e.Operation != program.Operation_COLLECT || len(e.Operands) != 3 {
		return v, false
	}
	for i, o := range e.Operands {
		switch {
		case isValue(o):
			v[i] = o.GetValue()
		case len(o.GetConstValues()) == 1:
			v[i] = o.GetConstValues()[0]
		default:
			return v, false
		}
	}
	return v, true
}

// isValue reports whether an operand is a number literal.
func isValue(o *program.Operand) bool {
	_, ok := o.GetT().(*program.Operand_Value)
	return ok
}

// partNearWGSL is the part nearest the point sdShape was last asked about:
// its distance, its size on screen in pixels and its shortest edge in world
// units.
const partNearWGSL = "gsl_partNear"

// partNearResetWGSL starts sdShape with no part near.
const partNearResetWGSL = "  gsl_partNear = vec3<f32>(1e10, 1e10, 1e10);\n"

// vanishWGSL declares the camera the guards read and gsl_vanish, which they
// call. gsl_vanishEye is the camera's position (Z-up) and its pixelScale, the
// world size of a pixel at distance 1; gsl_vanishView is 1 for an orthographic
// camera, then the band of on-screen sizes, in pixels, across which a part
// shrinks from whole to gone. gsl_vanish returns the weight, 1 for whole and 0
// for gone, the offset that shrinks the part, the part's size on screen in
// pixels and its shortest edge in world units (both far, 1e10, when the camera
// is unset). A box's size on screen is its
// shortest edge over the size of a pixel at its nearest point, so a part goes
// once it is thin, and it shrinks by half that edge, which is all a part that
// thin has to give. gsl_vanishView.w is the margin within which a part whose
// solid is in its box is evaluated, and 0 evaluates it everywhere;
// gsl_boxDistance is a point's distance to a box, 0 inside it.
const vanishWGSL = `var<private> gsl_vanishEye: vec4<f32>;
var<private> gsl_vanishView: vec4<f32>;
var<private> gsl_partNear: vec3<f32>;

fn gsl_vanish(lo: vec3<f32>, hi: vec3<f32>) -> vec4<f32> {
  if (gsl_vanishEye.w <= 0.0) { return vec4<f32>(1.0, 0.0, 1e10, 1e10); }
  let size = hi - lo;
  let dist = length(gsl_vanishEye.xyz - clamp(gsl_vanishEye.xyz, lo, hi));
  let pixel = gsl_vanishEye.w * mix(max(dist, 1.0), 1.0, gsl_vanishView.x);
  let edge = min(size.x, min(size.y, size.z));
  let px = edge / pixel;
  let w = smoothstep(gsl_vanishView.y, max(gsl_vanishView.z, gsl_vanishView.y + 1e-3), px);
  return vec4<f32>(w, (1.0 - w) * 0.5 * edge, px, edge);
}

fn gsl_boxDistance(q: vec3<f32>, lo: vec3<f32>, hi: vec3<f32>) -> f32 {
  return length(max(max(lo - q, q - hi), vec3<f32>(0.0)));
}`
