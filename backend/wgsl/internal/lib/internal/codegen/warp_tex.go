package codegen

import (
	"fmt"

	"gradientcontrol.com/omega/core/backend/wgsl/internal/lib/internal/wgsl"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/program"
)

// The warp field is a single global texture holding a simulated displacement,
// threaded into every shader function that (transitively) calls warpField.
// Mirrors the noise tile rather than the bvhSdf/vdbSdf slot machinery: those
// are per-mesh-asset, whereas a simulation produces one displacement
// field for the whole scene, so there is nothing to key slots on.
//
// Unlike the noise tile it needs a sampler (the field is interpolated, not
// fetched by exact texel) plus the bbox that maps a world position into the
// texture, and the symmetric range the components were normalized against.
// rgba16float is the filterable four-channel format; the alpha channel is
// unused.
const (
	warpTexName   = "warpTex"
	warpSmpName   = "warpSmp"
	warpBMinName  = "warpBMin"
	warpBSizeName = "warpBSize"
	warpRangeName = "warpRange"
	warpDimsName  = "warpDims"
)

// warpParams is the comma-prefixed parameter list appended to every function
// signature that can reach warpField.
const warpParams = ", " + warpTexName + ": texture_3d<f32>" +
	", " + warpSmpName + ": sampler" +
	", " + warpBMinName + ": vec3<f32>" +
	", " + warpBSizeName + ": vec3<f32>" +
	", " + warpRangeName + ": f32" +
	", " + warpDimsName + ": vec3<f32>"

// warpSampleWGSL is the tricubic sampler, emitted once per module that uses
// warpField. Tricubic rather than the hardware trilinear because the surface
// normal of a warped SDF depends on the *gradient* of the displacement, and
// trilinear is only C0 -- its gradient jumps at every cell boundary, which
// shows up as a crease on each face of the grid. A cubic B-spline is C2, so
// the normals stay smooth.
//
// The classic Sigg & Hadwiger trick: a cubic B-spline needs 4 taps per axis,
// but pairs of them can be folded into one linearly-filtered tap placed
// off-centre, so 64 fetches become 8. That is what h0/h1 are -- offset sample
// positions chosen so the hardware's own lerp produces the B-spline weighting.
//
// Note this is approximating rather than interpolating: the surface passes
// near the samples, not exactly through them. Correct for a displacement field,
// which is smooth and measured, not a set of hard constraints.
const warpSampleWGSL = `fn gsl_warpSample(pw: vec3<f32>` + warpParams + `) -> vec3<f32> {
    let uvw = (pw - ` + warpBMinName + `) / ` + warpBSizeName + `;
    // Outside the simulated box the field is zero, not the edge value: clamping
    // would drag the rest of the scene along with the boundary displacement.
    if (any(uvw < vec3<f32>(0.0)) || any(uvw > vec3<f32>(1.0))) {
        return vec3<f32>(0.0);
    }
    let t = uvw * ` + warpDimsName + ` - 0.5;
    let i = floor(t);
    let f = t - i;
    let f2 = f * f;
    let f3 = f2 * f;
    let w0 = (1.0 / 6.0) * (1.0 - 3.0 * f + 3.0 * f2 - f3);
    let w1 = (1.0 / 6.0) * (4.0 - 6.0 * f2 + 3.0 * f3);
    let w3 = (1.0 / 6.0) * f3;
    let g0 = w0 + w1;
    let g1 = 1.0 - g0;
    let h0 = (i - 0.5 + w1 / g0) / ` + warpDimsName + `;
    let h1 = (i + 1.5 + w3 / g1) / ` + warpDimsName + `;
    let c000 = textureSampleLevel(` + warpTexName + `, ` + warpSmpName + `, vec3<f32>(h0.x, h0.y, h0.z), 0.0).rgb;
    let c100 = textureSampleLevel(` + warpTexName + `, ` + warpSmpName + `, vec3<f32>(h1.x, h0.y, h0.z), 0.0).rgb;
    let c010 = textureSampleLevel(` + warpTexName + `, ` + warpSmpName + `, vec3<f32>(h0.x, h1.y, h0.z), 0.0).rgb;
    let c110 = textureSampleLevel(` + warpTexName + `, ` + warpSmpName + `, vec3<f32>(h1.x, h1.y, h0.z), 0.0).rgb;
    let c001 = textureSampleLevel(` + warpTexName + `, ` + warpSmpName + `, vec3<f32>(h0.x, h0.y, h1.z), 0.0).rgb;
    let c101 = textureSampleLevel(` + warpTexName + `, ` + warpSmpName + `, vec3<f32>(h1.x, h0.y, h1.z), 0.0).rgb;
    let c011 = textureSampleLevel(` + warpTexName + `, ` + warpSmpName + `, vec3<f32>(h0.x, h1.y, h1.z), 0.0).rgb;
    let c111 = textureSampleLevel(` + warpTexName + `, ` + warpSmpName + `, vec3<f32>(h1.x, h1.y, h1.z), 0.0).rgb;
    let x00 = mix(c100, c000, g0.x);
    let x10 = mix(c110, c010, g0.x);
    let x01 = mix(c101, c001, g0.x);
    let x11 = mix(c111, c011, g0.x);
    let y0 = mix(x10, x00, g0.y);
    let y1 = mix(x11, x01, g0.y);
    // Decode the symmetric [0,1] encoding back to millimetres.
    return (mix(y1, y0, g0.z) * 2.0 - 1.0) * ` + warpRangeName + `;
}`

// buildWarpFieldCall lowers warpField(p) to a filtered sample of the baked
// displacement grid, decoded back to millimetres.
//
// The bake stores each component normalized to [0,1] against one symmetric
// range shared by all three, so zero lands exactly on 0.5 and the direction is
// not sheared; decoding is therefore (texel*2 - 1) * range per component.
//
// Outside the baked box the result is forced to zero rather than clamped to the
// edge value. Clamping would drag everything the field does not cover — supports,
// the grid — along with the boundary displacement, which reads as the
// whole scene sliding.
type WarpFieldExpr struct {
	wgsl.Expression
	Pos wgsl.Expression
}

func NewWarpFieldExpr(op *program.Operand, pos wgsl.Expression) *WarpFieldExpr {
	w := &WarpFieldExpr{Pos: pos}
	reads := append([]string{}, pos.Reads()...)
	reads = append(reads, warpTexName, warpSmpName, warpBMinName, warpBSizeName, warpRangeName, warpDimsName)
	w.Expression = wgsl.NewExprWithReads(op, wgsl.TypeVec3, build.PrecedenceHighest, reads...)
	w.Expression.SetOuter(w)
	return w
}

func (w *WarpFieldExpr) Codegen(ctx wgsl.CodegenContext) string {
	return fmt.Sprintf("gsl_warpSample(%s, %s, %s, %s, %s, %s, %s)",
		w.Pos.Codegen(ctx), warpTexName, warpSmpName, warpBMinName,
		warpBSizeName, warpRangeName, warpDimsName)
}

func (d *Delegate) buildWarpFieldCall(
	op *program.Operand, args []build.Expression,
) (build.Expression, error) {
	pos := args[0].(wgsl.Expression)
	return NewWarpFieldExpr(op, pos), nil
}
