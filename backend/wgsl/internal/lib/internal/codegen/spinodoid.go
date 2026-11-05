package codegen

import (
	"fmt"

	api "gradientcontrol.com/omega/core/backend/wgsl/internal/lib"
	"gradientcontrol.com/omega/core/backend/wgsl/internal/lib/internal/wgsl"
	"gradientcontrol.com/omega/core/internal/build"
	"gradientcontrol.com/omega/core/internal/noise/spinodoid"
	"gradientcontrol.com/omega/core/program"
)

// The spinodoid in the preview (lattice Gabor; see common/noise/spinodoid).
// Its tables are baked by the fidget agent (only fidget evaluates the scale
// and orientation drivers) and bound here as ONE texture, every table a band
// of rows. Four metadata texels per slot in row zero describe its origin,
// dimensions, domain, and data offset; the data bands have variable lengths.
// One fixed parameter
// means the signatures do not depend on how many tables the build turns up,
// which is only known once it has run; the canvas matches each slot to a bake
// by the call's spinodoid.Call key (api.SpinodoidSlot).
//
// The scalar constants (lattice spacing, reach, envelope, normalization,
// level) depend only on the call's constant arguments, so they are computed
// here, in Go, from the same functions the baker uses, and emitted as
// literals.
const spinodoidTexName = "spinodoidTex"
const spinodoidTexParam = spinodoidTexName + ": texture_2d<f32>"

// spinodoid.GPUHeaderTexels metadata texels per slot must fit in the fixed-width
// header row. PackSpinodoidBinding additionally checks that all data bands fit
// the GPU's texture-height limit (spinodoid.GPUMaxRows).
const spinodoidMaxSlots = 256

func spinodoidHeaderWGSL(outside string) string {
	return fmt.Sprintf(`  let originRow = textureLoad(tex, vec2<i32>(slot * %[1]d, 0), 0);
  let sizePeriodic = textureLoad(tex, vec2<i32>(slot * %[1]d + 1, 0), 0);
  let origin = vec3<i32>(originRow.xyz);
  let size = vec3<i32>(sizePeriodic.xyz);
  let row = i32(originRow.w);
  let periodic = sizePeriodic.w > 0.5;
  if (!periodic) {
    let domainMin = textureLoad(tex, vec2<i32>(slot * %[1]d + 2, 0), 0).xyz;
    let domainMax = textureLoad(tex, vec2<i32>(slot * %[1]d + 3, 0), 0).xyz;
    if (any(q < domainMin) || any(q > domainMax)) {
      let outsideDelta = q - clamp(q, domainMin, domainMax);
      return %[2]s;
    }
  }
`, spinodoid.GPUHeaderTexels, outside)
}

// spinodoidLoopWGSL visits every lattice site within the reach of q, row by
// row (for each z the y span left inside the sphere, for each y the x span),
// so no site outside it costs a texture load; BODY sees the site's offset r,
// its squared distance d2, its wave vector kv, its phase and its tapered
// envelope k.
// Mirrors tape-core's spinodoid.rs for_sites_within / eval in f32.
func spinodoidLoopWGSL(body string) string {
	return fmt.Sprintf(`  let pi = 3.141592653589793;
  let a2 = a * a;
  let r2max = reach * reach;
  for (var z = i32(ceil((q.z - reach) / h)); z <= i32(floor((q.z + reach) / h)); z++) {
    let dz = q.z - f32(z) * h;
    let ry2 = r2max - dz * dz;
    if (ry2 < 0.0) { continue; }
    let ry = sqrt(ry2);
    var wz = z - origin.z;
    if (periodic) { wz = ((wz %% size.z) + size.z) %% size.z; }
    if (wz < 0 || wz >= size.z) { continue; }
    for (var y = i32(ceil((q.y - ry) / h)); y <= i32(floor((q.y + ry) / h)); y++) {
      let dy = q.y - f32(y) * h;
      let rx2 = ry2 - dy * dy;
      if (rx2 < 0.0) { continue; }
      let rx = sqrt(rx2);
      var wy = y - origin.y;
      if (periodic) { wy = ((wy %% size.y) + size.y) %% size.y; }
      if (wy < 0 || wy >= size.y) { continue; }
      for (var x = i32(ceil((q.x - rx) / h)); x <= i32(floor((q.x + rx) / h)); x++) {
        let r = q - vec3<f32>(f32(x), f32(y), f32(z)) * h;
        let d2 = dot(r, r);
        if (d2 >= r2max) { continue; }
        var wx = x - origin.x;
        if (periodic) { wx = ((wx %% size.x) + size.x) %% size.x; }
        if (wx < 0 || wx >= size.x) { continue; }
        let index = (wy * size.z + wz) * size.x + wx;
        let t = textureLoad(tex, vec2<i32>(index %% %[1]d, row + index / %[1]d), 0);
        let kv = t.xyz;
        let phase = dot(r, kv) + t.w;
        let g = exp(-pi * a2 * d2);
        let taper = 1.0 - d2 / r2max;
        let k = g * taper * taper;
%[2]s
      }
    }
  }
`, spinodoid.GPUWidth, body)
}

const spinodoidParamsWGSL = "q: vec3<f32>, tex: texture_2d<f32>, slot: i32, h: f32, reach: f32, a: f32, sigma: f32, baseBeta: f32, level: f32, localNorm: bool"

// spinodoidWGSL is the pseudo-distance alone: what gaborSpinodoid asks for and
// what the march calls, so it skips every gradient term.
var spinodoidWGSL = "fn gsl_spinodoid(" + spinodoidParamsWGSL + ") -> f32 {\n" +
	spinodoidHeaderWGSL("length(outsideDelta)") +
	"  var acc = 0.0;\n  var wAcc = 0.0;\n  var bwAcc = 0.0;\n" +
	spinodoidLoopWGSL(`        acc += k * cos(phase);
        if (localNorm) {
          let w = k * k;
          wAcc += w;
          bwAcc += w * length(kv);
        }`) +
	`  var b = baseBeta;
  if (localNorm && wAcc > 0.0 && bwAcc > 0.0) {
    b = bwAcc / wAcc;
  }
  return (acc / sigma - level) / b;
}`

// spinodoidGradWGSL is the analytic gradient, for gaborSpinodoidGradient.
var spinodoidGradWGSL = "fn gsl_spinodoidGrad(" + spinodoidParamsWGSL + ") -> vec3<f32> {\n" +
	spinodoidHeaderWGSL("normalize(outsideDelta)") +
	"  var acc = 0.0;\n  var gAcc = vec3<f32>(0.0);\n  var wAcc = 0.0;\n  var bwAcc = 0.0;\n  var gw = vec3<f32>(0.0);\n  var gbw = vec3<f32>(0.0);\n" +
	spinodoidLoopWGSL(`        let dk = g * taper * (-2.0 * pi * a2 * taper - 4.0 / r2max);
        let s = sin(phase);
        let co = cos(phase);
        acc += k * co;
        gAcc += dk * r * co - k * s * kv;
        let beta = length(kv);
        let w = k * k;
        wAcc += w;
        bwAcc += w * beta;
        let dw = 2.0 * k * dk * r;
        gw += dw;
        gbw += dw * beta;`) +
	`  let phi = acc / sigma;
  let gphi = gAcc / sigma;
  var b = baseBeta;
  var gb = vec3<f32>(0.0);
  if (localNorm && wAcc > 0.0 && bwAcc > 0.0) {
    b = bwAcc / wAcc;
    gb = (gbw - b * gw) / wAcc;
  }
  let num = phi - level;
  return gphi / b - num * gb / (b * b);
}`

// BuildSpinodoidCall lowers gaborSpinodoid / gaborSpinodoidGradient to
// gsl_spinodoid over the call's table band.
func (d *Delegate) BuildSpinodoidCall(
	op *program.Operand, fnCall *program.FnCall, args []build.Expression,
) (build.Expression, error) {
	name := fnCall.Reference.Name
	constant := func(i int, what string) ([]float64, error) {
		ce, ok := args[i].(build.ConstExpression)
		if !ok || ce.ConstValues() == nil {
			return nil, fmt.Errorf("%s: %s must be a constant: it keys the baked kernel table", name, what)
		}
		return ce.ConstValues(), nil
	}
	scalar := func(i int, what string) (float64, error) {
		v, err := constant(i, what)
		if err != nil {
			return 0, err
		}
		if len(v) != 1 {
			return 0, fmt.Errorf("%s: %s must be a scalar", name, what)
		}
		return v[0], nil
	}

	var call spinodoid.Call
	var err error
	if call.Period, err = scalar(1, "period"); err != nil {
		return nil, err
	}
	if call.Theta, err = constant(2, "theta"); err != nil {
		return nil, err
	}
	if call.Poles, err = constant(3, "poles"); err != nil {
		return nil, err
	}
	if call.PacketWidth, err = scalar(4, "packetWidth"); err != nil {
		return nil, err
	}
	impulses, err := scalar(5, "impulses")
	if err != nil {
		return nil, err
	}
	call.Impulses = int(impulses)
	if call.Impulses < 1 || float64(call.Impulses) != impulses {
		return nil, fmt.Errorf("%s: impulses must be a whole number >= 1, got %g", name, impulses)
	}
	pool, err := scalar(6, "pool")
	if err != nil {
		return nil, err
	}
	call.Pool = int(pool)
	seed, err := scalar(7, "seed")
	if err != nil {
		return nil, err
	}
	call.Seed = int64(seed)
	if call.Bias, err = scalar(8, "bias"); err != nil {
		return nil, err
	}
	local, err := scalar(9, "localNormalize")
	if err != nil {
		return nil, err
	}
	call.Local = local != 0
	// Without a field scale every kernel has the same wavenumber, so the
	// energy-weighted one is the base one: the shader skips the weights.
	weighLocally := call.Local
	if ce, ok := args[10].(build.ConstExpression); ok && ce.ConstValues() != nil {
		weighLocally = false
	}
	call.Chain = build.KeyedCallSite(d.GetScope(), op)

	// The bias is the level itself (spinodoid.CheckBias).
	level := call.Bias
	if err := spinodoid.CheckBias(level); err != nil {
		return nil, fmt.Errorf("%s: %v", name, err)
	}
	if !(call.Period > 0) || !(call.PacketWidth > 0) {
		return nil, fmt.Errorf("%s: period and packetWidth must be > 0", name)
	}
	consts := spinodoid.TableConstants(call.Period, call.PacketWidth, call.Impulses)
	// A constant scale other than 1 sets every kernel's wavenumber; that, not
	// the period's, is then the one to divide by.
	baseBeta := consts.BaseBeta
	if ce, ok := args[10].(build.ConstExpression); ok && ce.ConstValues() != nil && call.Local {
		if s := ce.ConstValues()[0]; s > 0 {
			baseBeta = consts.BaseBeta / s
		}
	}

	// gaborSpinodoid and fastGaborSpinodoid are both the lattice field, here
	// and in the fidget bake, so the key it asks the baker for is the lattice
	// one.
	call.Fast = true
	key := call.Key()
	slot, ok := d.spinodoidSlotByKey[key]
	if !ok {
		if len(d.spinodoidSlots) >= spinodoidMaxSlots {
			return nil, fmt.Errorf("%s: the preview binds at most %d spinodoid tables", name, spinodoidMaxSlots)
		}
		if d.spinodoidSlotByKey == nil {
			d.spinodoidSlotByKey = map[string]int{}
		}
		slot = len(d.spinodoidSlots)
		d.spinodoidSlotByKey[key] = slot
		d.spinodoidSlots = append(d.spinodoidSlots, api.SpinodoidSlot{Slot: slot, Key: key})
	}
	q := args[0].(wgsl.Expression)
	fn := "gsl_spinodoid"
	var outType wgsl.TypeInfo = wgsl.TypeF32
	if program.IsGaborSpinodoidGradient(name) {
		fn = "gsl_spinodoidGrad"
		outType = wgsl.TypeVec3
		d.spinodoidGradEmitted = true
	} else {
		d.spinodoidEmitted = true
	}
	tableArgs := fmt.Sprintf("%s, %d, %s, %s, %s, %s, %s, %s, %t",
		spinodoidTexName, slot,
		wgsl.CodegenValue(consts.Spacing), wgsl.CodegenValue(consts.Reach), wgsl.CodegenValue(consts.A),
		wgsl.CodegenValue(consts.Sigma), wgsl.CodegenValue(baseBeta), wgsl.CodegenValue(level), weighLocally)
	return NewSpinodoidExpr(op, outType, fn, q, tableArgs), nil
}

// SpinodoidExpr is a call of a spinodoid helper at a query point: the helper
// name, the point, and the table arguments after it (already WGSL). It reads
// the point's variables and the spinodoid texture.
type SpinodoidExpr struct {
	wgsl.Expression
	fn   string
	q    wgsl.Expression
	args string
}

func NewSpinodoidExpr(op *program.Operand, outType wgsl.TypeInfo, fn string, q wgsl.Expression, args string) *SpinodoidExpr {
	e := &SpinodoidExpr{fn: fn, q: q, args: args}
	reads := append([]string{}, q.Reads()...)
	reads = append(reads, spinodoidTexName)
	e.Expression = wgsl.NewExprWithReads(op, outType, build.PrecedenceHighest, reads...)
	e.Expression.SetOuter(e)
	return e
}

func (e *SpinodoidExpr) Codegen(ctx wgsl.CodegenContext) string {
	return fmt.Sprintf("%s(%s, %s)", e.fn, e.q.Codegen(ctx), e.args)
}
