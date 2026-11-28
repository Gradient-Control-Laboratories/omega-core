package codegen

import (
	"fmt"
	"strings"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/build"
)

// toleranceFieldFn is the generated function that returns a return's working
// tolerance at a point.
const toleranceFieldFn = "gsl_toleranceField"

// Ensure the delegate draws a part at its most or least material.
var _ build.ToleranceReturnsBuilder = (*Delegate)(nil)

// BuildPackageToleranceReturns implements [build.ToleranceReturnsBuilder]: it
// generates gsl_toleranceField, the working tolerance of the return
// gsl_tolIdx (1-based) at p, a vec2 [lo, hi] of offsets of its distance, with
// the same parameters as gsl_colorField so the viewer calls it the same way.
// The viewer draws the part at its most material by adding lo to the winning
// return's distance, at its least by adding hi. Called only for a document
// that declares a working tolerance.
func (d *Delegate) BuildPackageToleranceReturns(toleranceExpressions []build.Expression) error {
	if !anyExpression(toleranceExpressions) {
		return nil
	}
	childSrc, caseLets, caseSrc := d.codegenFieldCases(toleranceFieldFn, toleranceExpressions)

	var sb strings.Builder
	if childSrc != "" {
		sb.WriteString(childSrc)
		sb.WriteString("\n\n")
	}
	fmt.Fprintf(&sb, "fn %s(p: vec3<f32>, time: f32, gsl_tolIdx: u32%s) -> vec2<f32> {\n",
		toleranceFieldFn, d.fieldFnParamSuffix())
	for i, src := range caseSrc {
		if src == "" {
			continue
		}
		fieldCase(&sb, "gsl_tolIdx", i+1, caseLets[i], fmt.Sprintf("return %s;", src))
	}
	sb.WriteString("  return vec2<f32>(0.0, 0.0);\n")
	sb.WriteString("}")
	d.toleranceFieldSrc = sb.String()
	return nil
}
