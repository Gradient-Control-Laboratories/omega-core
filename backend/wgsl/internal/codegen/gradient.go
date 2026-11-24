package codegen

import (
	"fmt"
	"strings"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/build"
)

// gradFieldFn is the name of the generated per-return analytic gradient.
const gradFieldFn = "gsl_gradField"

// Ensure the delegate satisfies the optional gradient hook.
var _ build.GradientReturnsBuilder = (*Delegate)(nil)

// BuildPackageGradientReturns implements [build.GradientReturnsBuilder]. It
// weaves a
//
//	fn gsl_gradField(p, time, gsl_gradIdx: u32, <sdShape params>) -> vec4<f32>
//
// into the module when assembly lowered gradients (assembler.WithAnalyticGradient).
// The function returns the spatial gradient of the winning return's distance
// (selected by gsl_gradIdx, the 1-based return index) as vec4(grad, 1); a
// return the derivative could not go through falls through to vec4(0), so the
// viewer takes central differences there instead (w = 0 = "no gradient").
//
// Same fixed signature as sdShape and gsl_colorField (plus the index) so the
// viewer's arg injection treats it alike.
func (d *Delegate) BuildPackageGradientReturns(gradientExpressions []build.Expression) error {
	if !anyExpression(gradientExpressions) {
		return nil
	}

	childSrc, caseLets, caseSrc := d.codegenFieldCases(gradFieldFn, gradientExpressions)

	var sb strings.Builder
	if childSrc != "" {
		sb.WriteString(childSrc)
		sb.WriteString("\n\n")
	}
	fmt.Fprintf(&sb, "fn %s(p: vec3<f32>, time: f32, gsl_gradIdx: u32%s) -> vec4<f32> {\n",
		gradFieldFn, d.fieldFnParamSuffix())
	for i, src := range caseSrc {
		if src == "" {
			continue
		}
		fieldCase(&sb, "gsl_gradIdx", i+1, caseLets[i], fmt.Sprintf("return vec4<f32>(%s, 1.0);", src))
	}
	sb.WriteString("  return vec4<f32>(0.0, 0.0, 0.0, 0.0);\n")
	sb.WriteString("}")

	d.gradFieldSrc = sb.String()
	return nil
}
