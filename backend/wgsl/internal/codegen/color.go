package codegen

import (
	"fmt"
	"strings"

	"gradientcontrol.com/omega/core/internal/build"
)

// colorFieldFn is the name of the generated per-return color function.
const colorFieldFn = "gsl_colorField"

// Ensure the delegate satisfies the optional color hook.
var _ build.ColorReturnsBuilder = (*Delegate)(nil)

// BuildPackageColorReturns implements [build.ColorReturnsBuilder]. It weaves a
//
//	fn gsl_colorField(p, time, <sdShape params>, gsl_colorIdx: u32) -> vec4<f32>
//
// into the module when the program authors any color (name.color = …). The
// function returns the color field of the winning shape (selected by
// gsl_colorIdx, the 1-based return index) as vec4(rgb, 1); returns with no color
// fall through to vec4(0) so the viewer keeps its palette (alpha 0 = "no field").
//
// It has the same fixed signature as sdShape (plus the index) so the viewer can
// call it with the same arguments. The body re-derives each field's distance
// under the shared p — the same self-contained codegen sdShape and remap
// produce — by building in a fresh scope rooted at the primary build scope.
func (d *Delegate) BuildPackageColorReturns(colorExpressions []build.Expression) error {
	if !anyExpression(colorExpressions) {
		return nil
	}

	childSrc, caseLets, caseSrc := d.codegenFieldCases(colorFieldFn, colorExpressions)

	var sb strings.Builder
	if childSrc != "" {
		sb.WriteString(childSrc)
		sb.WriteString("\n\n")
	}

	// gsl_colorIdx comes right after time, before the lifted-literal / mesh-SDF /
	// noise suffix, so the viewer's injectExtraArgs (which appends that suffix to
	// each call site) lands the extra args in the same positions as sdShape.
	fmt.Fprintf(&sb, "fn %s(p: vec3<f32>, time: f32, gsl_colorIdx: u32%s) -> vec4<f32> {\n",
		colorFieldFn, d.fieldFnParamSuffix())

	for i, src := range caseSrc {
		if src == "" {
			continue
		}
		fieldCase(&sb, "gsl_colorIdx", i+1, caseLets[i], fmt.Sprintf("return vec4<f32>(%s, 1.0);", src))
	}
	sb.WriteString("  return vec4<f32>(0.0, 0.0, 0.0, 0.0);\n")
	sb.WriteString("}")

	d.colorFieldSrc = sb.String()
	return nil
}
