package wgsl

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// A compiled program is not a shader: sdShape names declarations it expects of
// its host (gsl_shape, gsl_min, the pixel footprint) and takes every slider
// and texture as an argument. Host wraps one into a module any host can call:
// those declarations, one uniform block for time and the sliders, the textures
// as bindings, and two functions,
//
//	fn gsl_field(p: vec3<f32>) -> gsl_shape
//	fn gsl_march(origin, direction, tEntry, tExit, pixelScale, ortho, stepScale, detail) -> gsl_hit
//
// gsl_march is the front face of the viewer's primary march
// (examples/viewer/render/gsl-render-graph.js). It has no entry point: a host
// adds its own, turns the hit's distance along the ray into depth, and shades.

// MarchRule is the step rule of gsl_march. The emitted text takes every number
// from it, so the text and the table cannot drift.
type MarchRule struct {
	// Fudge shortens each gradient-corrected step.
	Fudge float64
	// JumpSlack is how much steeper than the gradient last measured the field
	// may be between two samples and still cross zero there; past it a sign
	// change is a step in the field, not a surface, and is not bisected.
	JumpSlack float64
	// MinSteps to MaxSteps is the step budget, by the path's length through
	// the box against the box's width, capped at PathScaleMax, times stepScale.
	MinSteps, MaxSteps int
	PathScaleMax       float64
	// HitEps is the hit tolerance's floor and HitPixels its size in pixels.
	HitEps, HitPixels float64
	// GradEps is the gradient taps' reach: its floor, its size in pixels, and
	// its size against the point's own magnitude (about four float32 ulps).
	GradEps, GradPixels, GradUlps float64
	// A part far from the ray reads as its box while marching: past
	// VanishPixels at the far end of the ray, never under VanishHitEps hit
	// tolerances.
	VanishHitEps, VanishPixels float64
}

// FrontMarch is the rule the viewer's primary march follows.
var FrontMarch = MarchRule{
	Fudge:        0.8,
	JumpSlack:    4,
	MinSteps:     256,
	MaxSteps:     1024,
	PathScaleMax: 1.75,
	HitEps:       0.001,
	HitPixels:    0.5,
	GradEps:      0.0001,
	GradPixels:   0.025,
	GradUlps:     4e-7,
	VanishHitEps: 8,
	VanishPixels: 4,
}

// HostBlock names the uniform block and its variable.
const (
	HostBlockType = "gsl_Host"
	HostBlockVar  = "gsl_host"
	FieldFn       = "gsl_field"
	MarchFn       = "gsl_march"
)

// HostUniform is one member of the uniform block: time, or an argument of
// sdShape that is a number or a vector. Offset and Size are in bytes, by
// WGSL's layout rules, which for these types are std140's.
type HostUniform struct {
	Name   string
	Type   string
	Offset int
	Size   int
}

// HostTexture is a texture or sampler argument of sdShape, declared at
// @group(0) @binding(Binding).
type HostTexture struct {
	Name    string
	Type    string
	Binding int
}

// HostModule is a compiled program made callable.
type HostModule struct {
	WGSL      string
	Uniforms  []HostUniform
	BlockSize int
	Textures  []HostTexture
	// Parts is whether the program guards parts, so gsl_march sets the vanish
	// globals.
	Parts bool
}

var sdShapeSig = regexp.MustCompile(`\bfn sdShape\(p: vec3<f32>, time: f32([^)]*)\) -> gsl_shape\b`)

// uniformLayout is the alignment and size of the types sdShape takes by value.
var uniformLayout = map[string][2]int{
	"f32":       {4, 4},
	"vec2<f32>": {8, 8},
	"vec3<f32>": {16, 12},
	"vec4<f32>": {16, 16},
}

func isTextureType(t string) bool {
	return t == "sampler" || strings.HasPrefix(t, "texture_")
}

// Host wraps res.WGSL as described above. extent is half the width of the box
// a host marches, which the step budget is measured against.
func Host(res *Result, extent float64) (*HostModule, error) {
	if !(extent > 0) {
		return nil, fmt.Errorf("wgsl host: the march box's half-width must be positive, got %g", extent)
	}
	sig := sdShapeSig.FindStringSubmatch(res.WGSL)
	if sig == nil {
		return nil, fmt.Errorf("wgsl host: the module has no sdShape(p, time, ...)")
	}
	out := &HostModule{Parts: declares(res.WGSL, "gsl_vanishEye")}
	uniforms := []HostUniform{{Name: "time", Type: "f32"}}
	args := []string{"p", HostBlockVar + ".time"}
	for _, param := range strings.Split(sig[1], ",") {
		param = strings.TrimSpace(param)
		if param == "" {
			continue
		}
		name, typ, ok := strings.Cut(param, ":")
		name, typ = strings.TrimSpace(name), strings.TrimSpace(typ)
		switch {
		case !ok:
			return nil, fmt.Errorf("wgsl host: sdShape parameter %q has no type", param)
		case isTextureType(typ):
			out.Textures = append(out.Textures, HostTexture{Name: name, Type: typ, Binding: 1 + len(out.Textures)})
			args = append(args, name)
		case uniformLayout[typ] != [2]int{}:
			uniforms = append(uniforms, HostUniform{Name: name, Type: typ})
			args = append(args, HostBlockVar+"."+name)
		default:
			return nil, fmt.Errorf("wgsl host: sdShape parameter %s has type %s, which a host cannot bind", name, typ)
		}
	}
	offset, align := 0, 16
	for i := range uniforms {
		l := uniformLayout[uniforms[i].Type]
		offset = (offset + l[0] - 1) / l[0] * l[0]
		uniforms[i].Offset, uniforms[i].Size = offset, l[1]
		offset += l[1]
	}
	out.Uniforms = uniforms
	out.BlockSize = (offset + align - 1) / align * align

	var b strings.Builder
	b.WriteString(hostShapeStructWGSL)
	b.WriteString(hostMaterialStructWGSL)
	b.WriteString(hostStateWGSL)
	b.WriteString(hostMinWGSL)
	fmt.Fprintf(&b, "\nstruct %s {\n", HostBlockType)
	for _, u := range uniforms {
		fmt.Fprintf(&b, "    %s: %s,\n", u.Name, u.Type)
	}
	fmt.Fprintf(&b, "}\n@group(0) @binding(0) var<uniform> %s: %s;\n", HostBlockVar, HostBlockType)
	for _, t := range out.Textures {
		fmt.Fprintf(&b, "@group(0) @binding(%d) var %s: %s;\n", t.Binding, t.Name, t.Type)
	}
	b.WriteString("\n")
	b.WriteString(res.WGSL)
	fmt.Fprintf(&b, "\n\nfn %s(p: vec3<f32>) -> gsl_shape {\n    return sdShape(%s);\n}\n", FieldFn, strings.Join(args, ", "))
	b.WriteString(marchWGSL(FrontMarch, extent, out.Parts, res.WGSL))
	out.WGSL = b.String()
	return out, nil
}

// declares reports whether the module declares the private variable name.
func declares(wgsl, name string) bool {
	return regexp.MustCompile(`\bvar<private> ` + regexp.QuoteMeta(name) + `\b`).MatchString(wgsl)
}

// What sdShape expects of its host. The two structs and gsl_min are the
// viewer's, word for word (examples/viewer/render/gsl-host.js and
// gsl-shading.js); host_test.go holds them to it.
const hostShapeStructWGSL = `
struct gsl_shape {
    distance: f32,    // the distance from the shape
    shapeIndex: u32,  // the index of the shape or zero if none
}
`

const hostMaterialStructWGSL = `
struct gsl_material {
    surface: vec4<f32>,   // roughness, metallic, ior, reserved
    emissive: vec4<f32>,  // sRGB r, g, b, strength
    volume: vec4<f32>,    // transmission, absorptionDistance, displacement, reserved
    lay: vec4<f32>,       // lay x, y, z, anisotropy
    present: f32,         // 1 = authored field, 0 = default record
}
`

// Every returned shape is visible and none is singled out: a host that wants
// either changes its own copy. gsl_pw is the pixel footprint a program's
// pixelSize() reads, which gsl_march sets at a hit, and gsl_pwScale its scale.
const hostStateWGSL = `
var<private> gsl_fieldFilterIdx: u32 = 0u;
var<private> gsl_pw: f32 = 0.0;
var<private> gsl_pwScale: f32 = 1.0;

fn gsl_isShapeVisible(shapeIndex: u32) -> bool {
    return true;
}
`

const hostMinWGSL = `
fn gsl_min( a : gsl_shape, b : gsl_shape ) -> gsl_shape {
    if (gsl_fieldFilterIdx != 0u) {
        if (b.shapeIndex == gsl_fieldFilterIdx) { return b; }
        return a;
    }
    var ra = a;
    var rb = b;
    if (a.shapeIndex > 0u && !gsl_isShapeVisible(a.shapeIndex)) { ra = gsl_shape(1e9, 0u); }
    if (b.shapeIndex > 0u && !gsl_isShapeVisible(b.shapeIndex)) { rb = gsl_shape(1e9, 0u); }
    // Nearest crossing, not nearest signed distance: compare |d| so every
    // returned shape keeps its own zero set instead of being folded into one
    // solid. The first crossing along a ray is the union boundary whenever the
    // camera is outside every solid, so closed geometry is unaffected; open
    // surfaces stop trimming each other (a half-space field like p.x claims
    // half of all space as its "interior", which min() would swallow). Only
    // the comparison uses |d| -- the winner's signed distance passes through
    // untouched, so normals, AO and the backface march still see a sign.
    if (abs(ra.distance) < abs(rb.distance)) { return ra; } else { return rb; }
}
`

// wgslFloat writes v as a WGSL f32 literal, always with a decimal point or an
// exponent.
func wgslFloat(v float64) string {
	s := strconv.FormatFloat(v, 'g', -1, 64)
	if !strings.ContainsAny(s, ".e") {
		s += ".0"
	}
	return s
}

// marchWGSL is gsl_gradient and gsl_march by rule r, for a box of half-width
// extent. With parts it hands the part guards the ray before the first sample
// and takes it back after the last; module is searched for the rungs and
// footprints the program declares, which the march sets when they are there.
func marchWGSL(r MarchRule, extent float64, parts bool, module string) string {
	f := wgslFloat
	vanishOn, vanishOff, hitState := "", "", ""
	if parts {
		vanishOn = fmt.Sprintf(`
    // While it marches, a part far from the ray reads as its box: past %[1]s
    // pixels at the far end of the ray and never under %[2]s hit tolerances.
    // The gradient at the hit reads every part whole.
    gsl_vanishEye = vec4<f32>(origin, pixelScale);
    gsl_vanishView = vec4<f32>(ortho, detail.w, 4.0 * detail.w,
        max(%[3]s * %[2]s, pixelScale * mix(max(tExit, 1.0), 1.0, ortho) * %[1]s));`,
			f(r.VanishPixels), f(r.VanishHitEps), f(r.HitEps))
		if declares(module, "gsl_simpleRung") {
			vanishOn += "\n    gsl_simpleRung = detail.z;"
		}
		vanishOff = "\n    gsl_vanishView.w = 0.0;"
		if declares(module, "gsl_detailPw") {
			hitState = "\n        gsl_detailPw = gsl_pw;"
		}
	}
	return fmt.Sprintf(`
struct gsl_hit {
    hit: bool,           // whether the ray met a surface between tEntry and tExit
    t: f32,              // the distance along the ray, in the units of direction's length
    shapeIndex: u32,     // the return it met, counted from 1
    gradient: vec3<f32>, // the field's gradient there, not normalized
}

// The field's gradient at p by four tetrahedral taps of reach ge. gsl_march
// takes every gradient through this one function.
fn gsl_gradient(p: vec3<f32>, ge: f32) -> vec3<f32> {
    let taps = array<vec3<f32>, 4>(
        vec3<f32>(1.0, -1.0, -1.0), vec3<f32>(-1.0, -1.0, 1.0),
        vec3<f32>(-1.0, 1.0, -1.0), vec3<f32>(1.0, 1.0, 1.0));
    var g = vec3<f32>(0.0);
    for (var k = 0u; k < 4u; k = k + 1u) {
        g = g + taps[k] * %[1]s(p + taps[k] * ge).distance;
    }
    return g / (4.0 * ge);
}

// The first surface along origin + t * direction for t in [tEntry, tExit],
// which the host has clipped to its box. direction is a unit vector.
// pixelScale is the world size of a pixel one unit from the eye, or of every
// pixel when ortho is 1. stepScale multiplies the step budget. detail is the
// part ladder's rungs in pixels (displacement, textures, simplified, vanish);
// only a program with parts reads it, and zeros keep every part whole.
fn %[2]s(origin: vec3<f32>, direction: vec3<f32>, tEntry: f32, tExit: f32,
        pixelScale: f32, ortho: f32, stepScale: f32, detail: vec4<f32>) -> gsl_hit {
    var result = gsl_hit(false, tExit, 0u, vec3<f32>(0.0));
    if (!(tEntry <= tExit && tExit > 0.0)) { return result; }
    let pathScale = clamp(max(tExit - max(tEntry, 0.0), 0.0) / %[3]s, 1.0, %[4]s);
    let maxSteps = u32(clamp(%[5]s * pathScale, %[5]s, %[6]s) * stepScale);%[13]s
    var t = max(tEntry, 0.0);
    var prevT = t;
    // Only read once a step has been taken (crossed needs t > prevT), by which
    // time it holds that step's sample.
    var prevSd: gsl_shape;
    var prevGradMag = 1.0;
    var hitSd = gsl_shape(0.0, 0u);
    for (var i = 0u; i < %[7]du; i = i + 1u) {
        if (i >= maxSteps || t > tExit) { break; }
        let p = origin + direction * t;
        let sd = %[1]s(p);
        let pixelWorld = pixelScale * mix(max(t, 1.0), 1.0, ortho);
        if (abs(sd.distance) < max(%[8]s, pixelWorld * %[9]s)) {
            hitSd = sd;
            break;
        }
        // An overshoot is a sign change of the same shape between two samples,
        // unless the samples are too far from zero for the field to have
        // crossed it between them: then the field steps across zero without
        // passing through it, and the march steps on.
        let crossed = sd.shapeIndex == prevSd.shapeIndex
            && (sd.distance < 0.0) != (prevSd.distance < 0.0)
            && t > prevT
            && abs(prevSd.distance) + abs(sd.distance) <= %[10]s * prevGradMag * (t - prevT);
        if (crossed) {
            t = (prevT + t) * 0.5;
        } else {
            // Divide the step by the gradient's magnitude, so a field that is
            // not 1-Lipschitz is not stepped through.
            let ge = max(max(%[11]s, pixelWorld * %[12]s), length(p) * %[15]s);
            let gradMag = max(length(gsl_gradient(p, ge)), 1.0);
            prevT = t;
            prevSd = sd;
            prevGradMag = gradMag;
            t = t + abs(sd.distance) / gradMag * %[16]s;
        }
    }%[14]s
    if (hitSd.shapeIndex > 0u) {
        let p = origin + direction * t;
        let pixelWorld = pixelScale * mix(max(t, 1.0), 1.0, ortho);
        gsl_pw = gsl_pwScale * pixelWorld;%[17]s
        let ge = max(max(%[11]s, pixelWorld * %[12]s), length(p) * %[15]s);
        result = gsl_hit(true, t, hitSd.shapeIndex, gsl_gradient(p, ge));
    }
    return result;
}
`,
		FieldFn, MarchFn,
		f(2*extent), f(r.PathScaleMax), f(float64(r.MinSteps)), f(float64(r.MaxSteps)), r.MaxSteps,
		f(r.HitEps), f(r.HitPixels), f(r.JumpSlack), f(r.GradEps), f(r.GradPixels),
		vanishOn, vanishOff, f(r.GradUlps), f(r.Fudge), hitState)
}
