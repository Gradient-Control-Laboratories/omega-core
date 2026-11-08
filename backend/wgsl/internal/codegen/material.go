package codegen

import (
	"fmt"
	"strings"

	"gradientcontrol.com/omega/core/backend/wgsl/internal/ir"
	"gradientcontrol.com/omega/core/backend/wgsl/internal/wgsl"
	"gradientcontrol.com/omega/core/internal/build"
)

// materialFieldFn is the name of the generated per-return material function.
const materialFieldFn = "gsl_materialField"

// materialStructName is the host-declared WGSL struct the material function
// returns. Like gsl_shape it is declared by the viewer, not by codegen, so the
// generated module only references it:
//
//	struct gsl_material {
//	    surface: vec4<f32>,   // roughness, metallic, ior, reserved
//	    emissive: vec4<f32>,  // sRGB r, g, b, strength
//	    volume: vec4<f32>,    // transmission, absorptionDistance, displacement, reserved
//	    lay: vec4<f32>,       // lay x, y, z, anisotropy
//	    present: f32,         // 1 = authored field, 0 = default record
//	}
const materialStructName = "gsl_material"

// materialDefaultWGSL is the record returned for a return that authors no
// material, with present = 0. Mirrors property.materialDefault and the
// material library's `standard`.
const materialDefaultWGSL = "vec4<f32>(0.5, 0.0, 1.5, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), vec4<f32>(0.0, 0.0, 0.0, 0.0), 0.0"

// Ensure the delegate satisfies the optional material hook.
var _ build.MaterialReturnsBuilder = (*Delegate)(nil)

// BuildPackageMaterialReturns implements [build.MaterialReturnsBuilder]. It
// weaves a
//
//	fn gsl_materialField(p, time, gsl_materialIdx: u32, <sdShape params>) -> gsl_material
//
// into the module when the program authors any material (name.material = …).
// The function returns the material record of the winning shape (selected by
// gsl_materialIdx, the 1-based return index) with present = 1; returns with no
// material fall through to the default record with present = 0, so the viewer
// consumes the result directly (WGSL select cannot pick between structs).
//
// Same fixed signature as sdShape and gsl_colorField (plus the index) so the
// viewer's arg injection treats all three alike.
func (d *Delegate) BuildPackageMaterialReturns(materialExpressions []build.Expression) error {
	if !anyExpression(materialExpressions) {
		return nil
	}

	childSrc, caseLets, caseSrc := d.codegenFieldCases(materialFieldFn, materialExpressions)

	var sb strings.Builder
	if childSrc != "" {
		sb.WriteString(childSrc)
		sb.WriteString("\n\n")
	}
	fmt.Fprintf(&sb, "fn %s(p: vec3<f32>, time: f32, gsl_materialIdx: u32%s) -> %s {\n",
		materialFieldFn, d.fieldFnParamSuffix(), materialStructName)
	for i, src := range caseSrc {
		if src == "" {
			continue
		}
		fieldCase(&sb, "gsl_materialIdx", i+1, caseLets[i],
			fmt.Sprintf("let gsl_m = %s; return %s(gsl_m[0], gsl_m[1], gsl_m[2], gsl_m[3], 1.0);", src, materialStructName))
	}
	fmt.Fprintf(&sb, "  return %s(%s);\n", materialStructName, materialDefaultWGSL)
	sb.WriteString("}")

	d.materialFieldSrc = sb.String()
	return d.buildHeights(materialExpressions)
}

// anyExpression reports whether at least one lowered field is present.
func anyExpression(exprs []build.Expression) bool {
	for _, e := range exprs {
		if e != nil {
			return true
		}
	}
	return false
}

// codegenFieldCases generates the source of each return's field expression
// inside a fresh scope named fnName and rooted at the primary build scope, so
// referenced field vars (e.g. a top-level `ball = sphere(1)`) re-derive as
// `let`s inside the field function rather than resolving to sdShape's
// already-generated ones. Mirrors Remap's helper setup. It returns the helper
// functions the expressions spawned (childSrc, to emit before the field
// function), and for each return its `let` lines and its expression source
// (both empty where the return has no field).
//
// Each return is built in a scope of its own, so its lets are only the ones
// it reads, and the caller writes them inside that return's case: a call
// computes the one field it is asked for. With the lets shared at the top of
// the function, every call computed every return's field, and displacement
// asks the material field at each step of the march.
func (d *Delegate) codegenFieldCases(fnName string, exprs []build.Expression) (childSrc string, caseLets, caseSrc []string) {
	prevScope := d.currentScope()
	funcCtx := NewBuildContext(d, d.rootPVar)
	caseLets = make([]string, len(exprs))
	caseSrc = make([]string, len(exprs))
	var children []string
	for i, e := range exprs {
		if e == nil {
			continue
		}
		tempScope := prevScope.copy(fnName)
		tempScope.bScope = d.primaryBScope
		d.withScopedAgent(tempScope, func() {
			caseSrc[i] = e.(wgsl.Expression).Codegen(funcCtx)
			ir.SinkStatements(tempScope.block, e.(wgsl.Expression).Reads()...)
			var bBuf strings.Builder
			_ = tempScope.block.Format(&bBuf, "  ")
			caseLets[i] = bBuf.String()
			for _, fn := range tempScope.funcs {
				var fnBuf strings.Builder
				_ = fn.Format(&fnBuf)
				if c := strings.TrimSpace(fnBuf.String()); c != "" {
					children = append(children, c)
				}
			}
		})
	}
	childSrc = strings.Join(children, "\n\n")
	return childSrc, caseLets, caseSrc
}

// codegenFieldCasesShared generates all return expressions inside a single
// shared scope, suitable for functions like gsl_heights that compute and
// return an array of all values together.
func (d *Delegate) codegenFieldCasesShared(fnName string, exprs []build.Expression) (childSrc, bodyLets string, caseSrc []string) {
	prevScope := d.currentScope()
	tempScope := prevScope.copy(fnName)
	tempScope.bScope = d.primaryBScope
	funcCtx := NewBuildContext(d, d.rootPVar)
	caseSrc = make([]string, len(exprs))
	d.withScopedAgent(tempScope, func() {
		for i, e := range exprs {
			if e == nil {
				continue
			}
			caseSrc[i] = e.(wgsl.Expression).Codegen(funcCtx)
		}
		var bBuf strings.Builder
		_ = tempScope.block.Format(&bBuf, "  ")
		bodyLets = bBuf.String()
		var childBuf strings.Builder
		for _, fn := range tempScope.funcs {
			_ = fn.Format(&childBuf)
			childBuf.WriteString("\n\n")
		}
		childSrc = strings.TrimSpace(childBuf.String())
	})
	return childSrc, bodyLets, caseSrc
}

// fieldCase writes one return's case of a field function: its lets and ret,
// the statement that returns its field, inside the case's if.
func fieldCase(sb *strings.Builder, idxName string, idx int, lets, ret string) {
	if lets == "" {
		fmt.Fprintf(sb, "  if (%s == %du) { %s }\n", idxName, idx, ret)
		return
	}
	fmt.Fprintf(sb, "  if (%s == %du) {\n", idxName, idx)
	for line := range strings.SplitSeq(strings.TrimRight(lets, "\n"), "\n") {
		sb.WriteString("  ")
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	fmt.Fprintf(sb, "    %s\n  }\n", ret)
}

// fieldFnParamSuffix is the parameter list a generated field function shares
// with sdShape after its own leading parameters: the lifted literals, the
// mesh-SDF slots and the noise tile. The index parameter comes before this
// suffix so the viewer's injectExtraArgs, which appends the suffix's arguments
// to each call site, lands them in the same positions as for sdShape.
func (d *Delegate) fieldFnParamSuffix() string {
	var sb strings.Builder
	for _, name := range d.sliderParamNames {
		fmt.Fprintf(&sb, ", %s: f32", name)
	}
	sb.WriteString(d.meshSdfSlots.SdShapeParamsSuffix())
	if d.noiseUsed {
		sb.WriteString(", " + noiseTexParam)
	}
	if d.spinodoidUsed {
		sb.WriteString(", " + spinodoidTexParam)
	}
	return sb.String()
}
