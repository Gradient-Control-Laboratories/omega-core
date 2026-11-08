package codegen

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"gradientcontrol.com/omega/core/backend/wgsl/internal/ir"
	"gradientcontrol.com/omega/core/backend/wgsl/internal/wgsl"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// A part can have a simpler stand-in, name.simple = proxy. The assembler wraps
// the part in builtin's simplified(part, proxy, lo, hi), inside its
// shrinkToVanish, and its colour and material fields in simplified(field,
// proxyField, lo, hi) (property.MarkParts, property.resolveField). The GSL
// body is the part's own, so fidget, the mesher and export never see the
// stand-in. The codegen writes the viewer's switch in place of each call
// instead, for a distance:
//
//	let w = gsl_simple(lo, hi);              // 1 the part, 0 the stand-in
//	var part = 1e10;
//	if (w > 0.0) { part = shape; }
//	if (w < 1.0) { part = mix(max(proxy, gsl_boxSdf(p, lo, hi)), part, w); }
//
// and for a field, row by row for a material:
//
//	var field = proxyField;
//	if (w > 0.0) { field = mix(field, partField, w); }
//
// Like the vanish weight, w depends only on the camera and the box, so every
// pixel takes the same branch, and it is constant over the part, so the mix
// of two distance bounds is one too. The stand-in is cropped to the part's box,
// so a part whose solid is in its box can still be read as the box far from it.
// A switch inside a part's guard is written inside the guard's own branch, so
// neither model is evaluated where the part is skipped or has vanished.
//
// gsl_simpleRung is the Detail slider's Simple rung in screen pixels; the part
// becomes its stand-in across the octave below it. A host that does not set
// it, and a rung of 0, keep every part whole.

// simplifiedNameRE matches builtin's simplified and the copies codegen
// specializes from it for other argument types: a colour's, a material's.
var simplifiedNameRE = regexp.MustCompile(`^builtin_simplified(_\d+)?$`)

// simplifies reports whether a function is builtin's simplified(shape, proxy,
// lo, hi).
func simplifies(wgslName string, argCount int) bool {
	return simplifiedNameRE.MatchString(wgslName) && argCount == 4
}

// simpleCall is a call to simplified: a distance's switch, or a field's blend
// for anything else.
func (d *Delegate) simpleCall(op *program.Operand, args []wgsl.Expression, returnType wgsl.TypeInfo) build.Expression {
	if returnType.IsNumericScalar() {
		return d.simpleGuard(op, args, returnType)
	}
	return d.simpleFieldGuard(op, args, returnType)
}

// simpleSwitch is a part's switch to its stand-in: the names it declares and
// the sources it reads, generated but not yet written.
type simpleSwitch struct {
	part, weight         string
	shape, proxy, lo, hi string
	p                    string
	readsShape           []string
	readsProxy           []string
	readsLo              []string
	readsHi              []string
}

func (s *simpleSwitch) emitTo(b *ir.Block) {
	weightInit := fmt.Sprintf("gsl_simple(%s, %s)", s.lo, s.hi)
	b.Add(ir.NewLetStmt(s.weight, "", weightInit, append(append([]string{}, s.readsLo...), s.readsHi...)))
	b.Add(ir.NewVarStmt(s.part, "", vanishedWGSL, nil))

	guard1Body := ir.NewBlock()
	guard1Body.Add(ir.NewAssignStmt(s.part, s.shape, s.readsShape))
	b.Add(ir.NewGuardStmt(fmt.Sprintf("%s > 0.0", s.weight), guard1Body, []string{s.weight}))

	guard2Body := ir.NewBlock()
	proxyMix := fmt.Sprintf("mix(max(%s, gsl_boxSdf(%s, %s, %s)), %s, %s)", s.proxy, s.p, s.lo, s.hi, s.part, s.weight)
	mixReads := append([]string{}, s.readsProxy...)
	mixReads = append(mixReads, s.p)
	mixReads = append(mixReads, s.readsLo...)
	mixReads = append(mixReads, s.readsHi...)
	mixReads = append(mixReads, s.part, s.weight)
	guard2Body.Add(ir.NewAssignStmt(s.part, proxyMix, mixReads))
	b.Add(ir.NewGuardStmt(fmt.Sprintf("%s < 1.0", s.weight), guard2Body, []string{s.weight}))
}

type SimpleGuardExpr struct {
	wgsl.Expression
	d     *Delegate
	shape wgsl.Expression
	proxy wgsl.Expression
	lo    wgsl.Expression
	hi    wgsl.Expression
}

func NewSimpleGuardExpr(op *program.Operand, d *Delegate, shape, proxy, lo, hi wgsl.Expression, returnType wgsl.TypeInfo) *SimpleGuardExpr {
	s := &SimpleGuardExpr{
		d:     d,
		shape: shape,
		proxy: proxy,
		lo:    lo,
		hi:    hi,
	}
	s.Expression = wgsl.NewExpr(op, returnType, build.PrecedenceHighest)
	s.Expression.SetOuter(s)
	return s
}

func (s *SimpleGuardExpr) Codegen(ctx wgsl.CodegenContext) string {
	d := s.d
	deferred := d.deferSimple
	d.deferSimple = false
	sw := &simpleSwitch{
		part:       d.uniqueLocalName("gsl_simplePart"),
		weight:     d.uniqueLocalName("gsl_simpleWeight"),
		lo:         s.lo.Codegen(ctx),
		hi:         s.hi.Codegen(ctx),
		shape:      s.shape.Codegen(ctx),
		proxy:      s.proxy.Codegen(ctx),
		p:          ctx.P().WGSLName(),
		readsShape: s.shape.Reads(),
		readsProxy: s.proxy.Reads(),
		readsLo:    s.lo.Reads(),
		readsHi:    s.hi.Reads(),
	}
	if deferred {
		d.pendingSimple = sw
		return sw.part
	}
	sw.emitTo(d.currentScope().Block())
	return sw.part
}

// simpleGuard is a call to simplified: the switch's statements and the part it
// draws as its value. Inside a part's guard (deferSimple) the statements are
// handed to the guard, in pendingSimple, to write inside its own branch.
func (d *Delegate) simpleGuard(op *program.Operand, args []wgsl.Expression, returnType wgsl.TypeInfo) build.Expression {
	d.simpleUsed = true
	shape, proxy, lo, hi := args[0], args[1], args[2], args[3]
	return NewSimpleGuardExpr(op, d, shape, proxy, lo, hi, returnType)
}

// shapeInGuard generates a part's shape for its guard (vanishGuard) and
// returns the switch to its stand-in when the shape is one, for the guard to
// write before the shape's value is read.
func (d *Delegate) shapeInGuard(shape wgsl.Expression, ctx wgsl.CodegenContext) (string, *simpleSwitch) {
	d.deferSimple = true
	src := shape.Codegen(ctx)
	d.deferSimple = false
	s := d.pendingSimple
	d.pendingSimple = nil
	return src, s
}

type SimpleFieldGuardExpr struct {
	wgsl.Expression
	d        *Delegate
	field    wgsl.Expression
	proxy    wgsl.Expression
	lo       wgsl.Expression
	hi       wgsl.Expression
	rows     int
	typeName string
}

func NewSimpleFieldGuardExpr(op *program.Operand, d *Delegate, field, proxy, lo, hi wgsl.Expression, returnType wgsl.TypeInfo, rows int, typeName string) *SimpleFieldGuardExpr {
	e := &SimpleFieldGuardExpr{
		d:        d,
		field:    field,
		proxy:    proxy,
		lo:       lo,
		hi:       hi,
		rows:     rows,
		typeName: typeName,
	}
	e.Expression = wgsl.NewExpr(op, returnType, build.PrecedenceHighest)
	e.Expression.SetOuter(e)
	return e
}

func (e *SimpleFieldGuardExpr) Codegen(ctx wgsl.CodegenContext) string {
	d := e.d
	weight := d.uniqueLocalName("gsl_simpleWeight")
	value := d.uniqueLocalName("gsl_simpleField")
	loSrc, hiSrc := e.lo.Codegen(ctx), e.hi.Codegen(ctx)
	proxySrc := e.proxy.Codegen(ctx)
	fieldSrc := e.field.Codegen(ctx)

	block := d.currentScope().Block()
	weightInit := fmt.Sprintf("gsl_simple(%s, %s)", loSrc, hiSrc)
	block.Add(ir.NewLetStmt(weight, "", weightInit, append(append([]string{}, e.lo.Reads()...), e.hi.Reads()...)))
	block.Add(ir.NewVarStmt(value, "", proxySrc, e.proxy.Reads()))
	guardBody := ir.NewBlock()

	if e.rows == 0 {
		mixExpr := fmt.Sprintf("mix(%s, %s, %s)", value, fieldSrc, weight)
		guardBody.Add(ir.NewAssignStmt(value, mixExpr, append(append([]string{value}, e.field.Reads()...), weight)))
	} else {
		full := d.uniqueLocalName("gsl_simpleFull")
		guardBody.Add(ir.NewLetStmt(full, "", fieldSrc, e.field.Reads()))
		blended := make([]string, e.rows)
		for i := range blended {
			blended[i] = fmt.Sprintf("mix(%s[%d], %s[%d], %s)", value, i, full, i, weight)
		}
		fullMix := fmt.Sprintf("%s(%s)", e.typeName, strings.Join(blended, ", "))
		guardBody.Add(ir.NewAssignStmt(value, fullMix, []string{value, full, weight}))
	}
	block.Add(ir.NewGuardStmt(fmt.Sprintf("%s > 0.0", weight), guardBody, []string{weight}))
	return value
}

// simpleFieldGuard is a call to simplified over a field: the blend's statements,
// written into the function being generated, and the drawn field as its value.
// A type it cannot blend keeps the part's field.
func (d *Delegate) simpleFieldGuard(op *program.Operand, args []wgsl.Expression, returnType wgsl.TypeInfo) build.Expression {
	field, proxy, lo, hi := args[0], args[1], args[2], args[3]
	rows, ok := blendRows(returnType)
	if !ok {
		return field
	}
	d.simpleUsed = true
	return NewSimpleFieldGuardExpr(op, d, field, proxy, lo, hi, returnType, rows, returnType.Name)
}

// Fine detail, name.detail = fine of features name.detailSize = size mm
// across, is builtin's detailValue(coarse, fine, size): fine where a pixel is
// under half the size, where fadeDetail(size) has not yet faded it to its
// coarse look, coarse elsewhere. Its GSL body is fine, so fidget and export,
// whose pixel is 0, take fine. The codegen writes it as a guard, so fine's
// work is done only where it is drawn:
//
//	var v = coarse;
//	if (max(gsl_pw, gsl_detailPw) < 0.5 * size) { let o = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);
//	  v = fine;
//	  gsl_pw = o;
//	}
//
// a constant size folded into the bound, and sinkIntoGuards moves the lets
// only fine reads inside the branch, after the line that opens it. The pixel
// is the larger of pixelSize()'s and gsl_detailPw, the viewer's Textures rung
// (gsl-shading.js), which a host that does not set it leaves 0: so the rung
// turns fine detail off in a colour too, whose own patterns read the screen's
// pixel and go flat only by its stand-in, and fine is evaluated at the pixel it
// was switched by.

// detailValueNameRE matches builtin's detailValue and the copies codegen
// specializes from it for other argument types.
var detailValueNameRE = regexp.MustCompile(`^builtin_detailValue(_\d+)?$`)

// detailPixelShare is the share of a feature's size below which a pixel
// draws its fine detail. TestDetailPixel holds it to detailValue's in
// builtin/base.gsl.
const detailPixelShare = 0.5

// detailPwDeclWGSL declares the pixel fine detail is switched by beyond the
// screen's: 0 until a viewer with a Detail slider sets it.
const detailPwDeclWGSL = "var<private> gsl_detailPw: f32 = 0.0;"

// detailValues reports whether a function is builtin's detailValue(coarse,
// fine, size).
func detailValues(wgslName string, argCount int) bool {
	return detailValueNameRE.MatchString(wgslName) && argCount == 3
}

// detailBound is the pixel below which fine detail of the given size is drawn:
// a number when the size is one, else the product written out.
func detailBound(sizeSrc string) string {
	if size, err := strconv.ParseFloat(sizeSrc, 64); err == nil {
		return wgsl.CodegenValue(detailPixelShare * size)
	}
	return fmt.Sprintf("%s * (%s)", wgsl.CodegenValue(detailPixelShare), sizeSrc)
}

type DetailGuardExpr struct {
	wgsl.Expression
	d      *Delegate
	coarse wgsl.Expression
	fine   wgsl.Expression
	size   wgsl.Expression
}

func NewDetailGuardExpr(op *program.Operand, d *Delegate, coarse, fine, size wgsl.Expression, returnType wgsl.TypeInfo) *DetailGuardExpr {
	e := &DetailGuardExpr{
		d:      d,
		coarse: coarse,
		fine:   fine,
		size:   size,
	}
	e.Expression = wgsl.NewExpr(op, returnType, build.PrecedenceHighest)
	e.Expression.SetOuter(e)
	return e
}

func (e *DetailGuardExpr) Codegen(ctx wgsl.CodegenContext) string {
	d := e.d
	d.detailUsed = true
	bound := detailBound(e.size.Codegen(ctx))
	drawn := d.uniqueLocalName("gsl_detailValue")
	outer := d.uniqueLocalName("gsl_detailOuterPw")
	block := d.currentScope().Block()
	coarseSrc := e.coarse.Codegen(ctx)
	block.Add(ir.NewVarStmt(drawn, "", coarseSrc, e.coarse.Reads()))
	fineSrc := e.fine.Codegen(ctx)

	guardBody := ir.NewBlock()
	guardBody.Add(ir.NewAssignStmt(drawn, fineSrc, e.fine.Reads()))
	guardBody.Add(ir.NewAssignStmt("gsl_pw", outer, []string{outer}))
	headerExtra := fmt.Sprintf("let %s = gsl_pw; gsl_pw = max(gsl_pw, gsl_detailPw);", outer)
	block.Add(ir.NewGuardStmtWithHeader(fmt.Sprintf("max(gsl_pw, gsl_detailPw) < %s", bound), headerExtra, guardBody, e.size.Reads()))

	return drawn
}

// detailGuard is a call to detailValue: the guard's statements, written into
// the function being generated, and the drawn value as its value.
func (d *Delegate) detailGuard(op *program.Operand, args []wgsl.Expression, returnType wgsl.TypeInfo) build.Expression {
	coarse, fine, size := args[0], args[1], args[2]
	return NewDetailGuardExpr(op, d, coarse, fine, size, returnType)
}

// arrayTypeRE reads an array type's element type and length, and
// matrixTypeRE a matrix type's columns and rows: a material is a mat4x4<f32>.
var (
	// TODO: this is hacky code. Replace this with something that uses
	// the assembler's type information rather than relying on string parsing.
	arrayTypeRE  = regexp.MustCompile(`^array<(.+), (\d+)>$`)
	matrixTypeRE = regexp.MustCompile(`^mat(\d)x(\d)<f32>$`)
)

// blendRows is how a switch blends a value of type t: by one mix of the
// whole, 0, for a scalar or a vector, and otherwise by one mix for each
// element of an array or column of a matrix, a material's four. ok is false
// for a type it cannot blend.
func blendRows(t wgsl.TypeInfo) (rows int, ok bool) {
	if m := arrayTypeRE.FindStringSubmatch(t.Name); m != nil {
		rows, _ = strconv.Atoi(m[2])
		return rows, true
	}
	if m := matrixTypeRE.FindStringSubmatch(t.Name); m != nil {
		rows, _ = strconv.Atoi(m[1])
		return rows, true
	}
	return 0, t.IsScalarOrVector()
}

// simpleWGSL declares the rung the switches read, gsl_simple, their weight,
// and gsl_boxSdf, the signed distance to a box they crop a stand-in by. It
// reads gsl_vanish's camera, which every part declares (vanishWGSL).
const simpleWGSL = `var<private> gsl_simpleRung: f32;

fn gsl_simple(lo: vec3<f32>, hi: vec3<f32>) -> f32 {
  if (gsl_vanishEye.w <= 0.0 || gsl_simpleRung <= 0.0) { return 1.0; }
  let px = gsl_vanish(lo, hi).z;
  return smoothstep(0.5 * gsl_simpleRung, gsl_simpleRung, px);
}

fn gsl_boxSdf(q: vec3<f32>, lo: vec3<f32>, hi: vec3<f32>) -> f32 {
  let d = abs(q - 0.5 * (lo + hi)) - 0.5 * (hi - lo);
  return length(max(d, vec3<f32>(0.0))) + min(max(d.x, max(d.y, d.z)), 0.0);
}`

// A colour or material can have a stand-in too, name.simple = stand, a
// texture's flat look above all. Expand and MarkParts write builtin's
// simpleValue(value, stand), whose GSL body is the value, and the codegen
// writes the viewer's switch in its place:
//
//	var v = stand;
//	if (gsl_simpleDetail > 0.0) { v = mix(v, value, gsl_simpleDetail); }
//
// for a colour, and row by row for a material. The weight is the value's: 1
// draws it, 0 the stand-in, without evaluating the value. A colour and a
// material share it, so a texture that hands back both goes flat in both at
// once. The codegen declares it after sdShape, 1 until a host sets it, so
// everything but the viewer draws the value. The viewer sets it at a hit from
// the size of the part it hit, by the Detail slider's Textures rung, blending
// across the octave below the rung (partDetailWGSL, gsl-shading.js).
//
// builtin's simpleTexture(value, stand) is a stand-in with a stand-in of its
// own, a simpler texture before the flat look. It is the same switch: both
// steps read the one weight, so the simpler texture shows only while the
// value blends to its flat look.

// simpleValueNameRE matches builtin's simpleValue and simpleTexture and the
// copies codegen specializes from them for other argument types.
var simpleValueNameRE = regexp.MustCompile(`^builtin_simple(Value|Texture)(_\d+)?$`)

// simpleValues reports whether a function is builtin's simpleValue(value,
// stand) or simpleTexture(value, stand).
func simpleValues(wgslName string, argCount int) bool {
	return simpleValueNameRE.MatchString(wgslName) && argCount == 2
}

// simpleDetailWGSL is a value's weight: 1 draws it, 0 its stand-in.
const simpleDetailWGSL = "gsl_simpleDetail"

// simpleValueDeclWGSL declares it, whole until a host says otherwise.
const simpleValueDeclWGSL = "var<private> gsl_simpleDetail: f32 = 1.0;"

type SimpleValueGuardExpr struct {
	wgsl.Expression
	d        *Delegate
	value    wgsl.Expression
	stand    wgsl.Expression
	rows     int
	typeName string
}

func NewSimpleValueGuardExpr(op *program.Operand, d *Delegate, value, stand wgsl.Expression, returnType wgsl.TypeInfo, rows int, typeName string) *SimpleValueGuardExpr {
	e := &SimpleValueGuardExpr{
		d:        d,
		value:    value,
		stand:    stand,
		rows:     rows,
		typeName: typeName,
	}
	e.Expression = wgsl.NewExpr(op, returnType, build.PrecedenceHighest)
	e.Expression.SetOuter(e)
	return e
}

func (e *SimpleValueGuardExpr) Codegen(ctx wgsl.CodegenContext) string {
	d := e.d
	scope := d.currentScope()
	standSrc := e.stand.Codegen(ctx)
	valueSrc := e.value.Codegen(ctx)
	key := standSrc + "\x00" + valueSrc
	if drawn, ok := d.simpleValues[scope][key]; ok {
		return drawn
	}
	drawn := d.uniqueLocalName("gsl_simpleValue")
	if d.simpleValues == nil {
		d.simpleValues = map[*Scope]map[string]string{}
	}
	if d.simpleValues[scope] == nil {
		d.simpleValues[scope] = map[string]string{}
	}
	d.simpleValues[scope][key] = drawn

	block := scope.Block()
	block.Add(ir.NewVarStmt(drawn, "", standSrc, e.stand.Reads()))
	guardBody := ir.NewBlock()

	if e.rows == 0 {
		mixExpr := fmt.Sprintf("mix(%s, %s, %s)", drawn, valueSrc, simpleDetailWGSL)
		guardBody.Add(ir.NewAssignStmt(drawn, mixExpr, append(append([]string{drawn}, e.value.Reads()...), simpleDetailWGSL)))
		block.Add(ir.NewGuardStmt(fmt.Sprintf("%s > 0.0", simpleDetailWGSL), guardBody, []string{simpleDetailWGSL}))
	} else {
		full := d.uniqueLocalName("gsl_simpleFull")
		guardBody.Add(ir.NewLetStmt(full, "", valueSrc, e.value.Reads()))
		blended := make([]string, e.rows)
		for i := range blended {
			blended[i] = fmt.Sprintf("mix(%s[%d], %s[%d], %s)", drawn, i, full, i, simpleDetailWGSL)
		}
		fullMix := fmt.Sprintf("%s(%s)", e.typeName, strings.Join(blended, ", "))
		guardBody.Add(ir.NewAssignStmt(drawn, fullMix, []string{drawn, full, simpleDetailWGSL}))
		block.Add(ir.NewGuardStmt(fmt.Sprintf("%s > 0.0", simpleDetailWGSL), guardBody, []string{simpleDetailWGSL}))
	}
	return drawn
}

// simpleValueGuard is a call to simpleValue or simpleTexture: the switch's
// statements, written into the function being generated, and the drawn value
// as its value. A type it cannot blend keeps the value.
func (d *Delegate) simpleValueGuard(op *program.Operand, args []wgsl.Expression, returnType wgsl.TypeInfo) build.Expression {
	value, stand := args[0], args[1]
	typeName := returnType.Name
	rows, ok := blendRows(returnType)
	if !ok {
		return value
	}
	d.simpleValueUsed = true
	return NewSimpleValueGuardExpr(op, d, value, stand, returnType, rows, typeName)
}
