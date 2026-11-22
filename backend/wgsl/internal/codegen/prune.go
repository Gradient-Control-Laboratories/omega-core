package codegen

import (
	"fmt"
	"regexp"

	"gradientcontrol.com/omega/core/backend/wgsl/internal/ir"
	"gradientcontrol.com/omega/core/backend/wgsl/internal/wgsl"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// A geometric feature, name.featureSize = size, is builtin's
// pruneBelow(field, size) (property/feature.go): the field everywhere but
// here, where it gives way once it is too small to see. Its size on screen is
// its size over the pixel at the point, gsl_featurePixel, which sdShape
// measures as it is entered, from the world position and the eye, so it holds
// inside every rigid remap below and is the same on every kind of ray. Across
// the band the viewer's Detail slider sets for parts (gsl_vanishView, from
// whole to gone) the feature shrinks by up to half its size, as a vanishing
// part does, and once gone it is nothing, which a union and a difference
// both pass unchanged: a bump drops out, a hole closes. It is written as a
// guard, so a feature that is gone is not evaluated:
//
//	let w = gsl_featureWeight(size);
//	var v = 100000.0;
//	if (w > 0.0) { v = field + (1.0 - w) * 0.5 * size; }
//
// With no camera set (fidget's parity harness, an export, a host with every
// rung at zero) the pixel is 0 and every feature is whole.

// pruneBelowNameRE matches builtin's pruneBelow and the copies codegen
// specializes from it.
var pruneBelowNameRE = regexp.MustCompile(`^builtin_pruneBelow(_\d+)?$`)

// pruneBelows reports whether a function is builtin's pruneBelow(field, size).
func pruneBelows(wgslName string, argCount int) bool {
	return pruneBelowNameRE.MatchString(wgslName) && argCount == 2
}

// featureWGSL declares the pixel sdShape measures and the weight a feature
// of a given size has at it: 1 whole, 0 gone, smoothstep across vanish's band.
const featureWGSL = `var<private> gsl_featurePixel: f32 = 0.0;

fn gsl_featureWeight(size: f32) -> f32 {
  if (gsl_featurePixel <= 0.0) { return 1.0; }
  return smoothstep(gsl_vanishView.y, max(gsl_vanishView.z, gsl_vanishView.y + 1e-3), size / gsl_featurePixel);
}`

// featurePixelWGSL is sdShape's first statement in a program with a feature:
// vanish's pixel at the world point p, or 0 with no camera.
const featurePixelWGSL = "select(0.0, gsl_vanishEye.w * mix(max(length(gsl_vanishEye.xyz - p), 1.0), 1.0, gsl_vanishView.x), gsl_vanishEye.w > 0.0)"

type PruneGuardExpr struct {
	wgsl.Expression
	d     *Delegate
	field wgsl.Expression
	size  wgsl.Expression
}

func NewPruneGuardExpr(op *program.Operand, d *Delegate, field, size wgsl.Expression, returnType wgsl.TypeInfo) *PruneGuardExpr {
	e := &PruneGuardExpr{d: d, field: field, size: size}
	e.Expression = wgsl.NewExpr(op, returnType, build.PrecedenceHighest)
	e.Expression.SetOuter(e)
	return e
}

func (e *PruneGuardExpr) Codegen(ctx wgsl.CodegenContext) string {
	d := e.d
	d.featureUsed = true
	block := d.currentScope().Block()
	size := d.uniqueLocalName("gsl_featureSize")
	weight := d.uniqueLocalName("gsl_featureWeight")
	drawn := d.uniqueLocalName("gsl_feature")
	block.Add(ir.NewLetStmt(size, "", e.size.Codegen(ctx), e.size.Reads()))
	block.Add(ir.NewLetStmt(weight, "", fmt.Sprintf("gsl_featureWeight(%s)", size), []string{size}))
	block.Add(ir.NewVarStmt(drawn, "", "100000.0", nil))
	fieldSrc := e.field.Codegen(ctx)
	body := ir.NewBlock()
	body.Add(ir.NewAssignStmt(drawn, fmt.Sprintf("%s + (1.0 - %s) * 0.5 * %s", fieldSrc, weight, size),
		append(append([]string{}, e.field.Reads()...), weight, size)))
	block.Add(ir.NewGuardStmt(fmt.Sprintf("%s > 0.0", weight), body, []string{weight}))
	return drawn
}

// pruneGuard is a call to pruneBelow: the guard's statements, written into the
// function being generated, and the drawn field as its value.
func (d *Delegate) pruneGuard(op *program.Operand, args []wgsl.Expression, returnType wgsl.TypeInfo) build.Expression {
	return NewPruneGuardExpr(op, d, args[0], args[1], returnType)
}
