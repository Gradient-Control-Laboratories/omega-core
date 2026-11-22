package gpuparity

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/eval"
)

// The eye every ray of the march test leaves from, and the pixel scale it
// marches with: about a 1000-pixel-tall picture.
var marchEye = [3]float64{3, -4, 30}

const marchPixelScale = 0.001

// marchShader wraps a `gsl eval shader` WGSL response as a compute shader for
// gpu-eval: one ray per point, from marchEye through the point, clipped to the
// response's box, its hit written as six floats. The uniform block becomes a
// private variable the shader fills itself, since gpu-eval binds no uniforms.
func marchShader(t *testing.T, res *eval.ShaderResult) string {
	t.Helper()
	block := "@group(0) @binding(0) var<uniform> gsl_host: gsl_Host;"
	if strings.Count(res.Code, block) != 1 {
		t.Fatalf("the response does not declare its uniform block as %q", block)
	}
	var b strings.Builder
	b.WriteString("@group(0) @binding(0) var<storage, read> gsl_points: array<vec4<f32>>;\n")
	b.WriteString("@group(0) @binding(1) var<storage, read_write> gsl_out: array<f32>;\n")
	b.WriteString(strings.Replace(res.Code, block, "var<private> gsl_host: gsl_Host;", 1))
	b.WriteString("\n@compute @workgroup_size(64)\n")
	b.WriteString("fn main(@builtin(global_invocation_id) gid: vec3<u32>) {\n")
	b.WriteString("    let i = gid.x;\n")
	b.WriteString("    if (i >= arrayLength(&gsl_points)) { return; }\n")
	for _, s := range res.Sliders {
		fmt.Fprintf(&b, "    gsl_host.%s = %v;\n", s.Uniform, float32(s.Value))
	}
	fmt.Fprintf(&b, "    let eye = vec3<f32>(%v, %v, %v);\n", marchEye[0], marchEye[1], marchEye[2])
	b.WriteString("    let dir = normalize(gsl_points[i].xyz - eye);\n")
	fmt.Fprintf(&b, "    let lo = vec3<f32>(%v, %v, %v);\n", res.Region[0], res.Region[1], res.Region[2])
	fmt.Fprintf(&b, "    let hi = vec3<f32>(%v, %v, %v);\n", res.Region[3], res.Region[4], res.Region[5])
	b.WriteString("    let t1 = (lo - eye) / dir;\n    let t2 = (hi - eye) / dir;\n")
	b.WriteString("    let tMin = min(t1, t2);\n    let tMax = max(t1, t2);\n")
	b.WriteString("    let tEntry = max(max(tMin.x, tMin.y), tMin.z);\n")
	b.WriteString("    let tExit = min(min(tMax.x, tMax.y), tMax.z);\n")
	fmt.Fprintf(&b, "    let h = gsl_march(eye, dir, tEntry, tExit, %v, 0.0, 1.0, vec4<f32>(0.0));\n", marchPixelScale)
	b.WriteString("    let o = i * 6u;\n")
	b.WriteString("    gsl_out[o] = select(0.0, 1.0, h.hit);\n    gsl_out[o + 1u] = h.t;\n    gsl_out[o + 2u] = f32(h.shapeIndex);\n")
	b.WriteString("    gsl_out[o + 3u] = h.gradient.x;\n    gsl_out[o + 4u] = h.gradient.y;\n    gsl_out[o + 5u] = h.gradient.z;\n")
	b.WriteString("}\n")
	return b.String()
}

// marchTargets is an n by n grid of points on the plane z = 0, half wide each
// way, which the rays pass through.
func marchTargets(n int, half float64) [][3]float32 {
	pts := make([][3]float32, 0, n*n)
	for j := 0; j < n; j++ {
		for i := 0; i < n; i++ {
			pts = append(pts, [3]float32{
				float32(-half + 2*half*(float64(i)+0.5)/float64(n)),
				float32(-half + 2*half*(float64(j)+0.5)/float64(n)), 0})
		}
	}
	return pts
}

func marchRay(target [3]float32) (dir [3]float64) {
	var n float64
	for k := range dir {
		dir[k] = float64(target[k]) - marchEye[k]
		n += dir[k] * dir[k]
	}
	n = math.Sqrt(n)
	for k := range dir {
		dir[k] /= n
	}
	return dir
}

func marchGPU(t *testing.T, e *eval.Evaluator, doc string, pts [][3]float32) (*eval.ShaderResult, []float32) {
	t.Helper()
	bin := filepath.Join(repoRoot, "target/release/gpu-eval")
	if _, err := os.Stat(bin); err != nil {
		t.Skipf("gpu-eval is not built (tool/build-fidget builds it): %v", err)
	}
	res, err := e.Shader(t.Context(), nil, doc, "", eval.ShaderOptions{Target: eval.ShaderWGSL})
	if err != nil {
		t.Fatal(err)
	}
	return res, runGPU(t, bin, marchShader(t, res), pts, 6, false, "")
}

// The emitted gsl_march, run on the GPU, meets a ball where the ray meets it
// in closed form, reports the body, and returns the field's gradient there.
func TestGpuMarchMeetsABall(t *testing.T) {
	const radius = 5.0
	doc := "@scene(40)\n@range(2.0, 12.0) r = 5.0;\nball = length(p) - r;\nreturn ball;\n"
	pts := marchTargets(24, 8)
	_, out := marchGPU(t, eval.New(pkgMgr, nil), doc, pts)

	hits := 0
	for i, target := range pts {
		dir := marchRay(target)
		// |eye + t dir| = radius.
		b := marchEye[0]*dir[0] + marchEye[1]*dir[1] + marchEye[2]*dir[2]
		c := marchEye[0]*marchEye[0] + marchEye[1]*marchEye[1] + marchEye[2]*marchEye[2] - radius*radius
		disc := b*b - c
		hit, tGPU := out[6*i] > 0.5, float64(out[6*i+1])
		// Within a pixel or so of the silhouette either answer is right.
		if math.Abs(disc) < 0.5 {
			continue
		}
		if want := disc > 0; hit != want {
			t.Errorf("ray %d: hit = %v, want %v (discriminant %g)", i, hit, want, disc)
			continue
		}
		if !hit {
			continue
		}
		hits++
		tWant := -b - math.Sqrt(disc)
		// The march stops a hit tolerance short of the surface, measured along the normal.
		if math.Abs(tGPU-tWant) > 0.1 {
			t.Errorf("ray %d: t = %g, want %g", i, tGPU, tWant)
		}
		if out[6*i+2] != 1 {
			t.Errorf("ray %d: shapeIndex = %g, want 1", i, out[6*i+2])
		}
		var dot, gn, pn float64
		for k := 0; k < 3; k++ {
			p := marchEye[k] + tGPU*dir[k]
			g := float64(out[6*i+3+k])
			dot, gn, pn = dot+p*g, gn+g*g, pn+p*p
		}
		if cos := dot / math.Sqrt(gn*pn); cos < 0.999 || math.Abs(math.Sqrt(gn)-1) > 0.01 {
			t.Errorf("ray %d: gradient %v is not the unit radial (cos %g, length %g)", i, out[6*i+3:6*i+6], cos, math.Sqrt(gn))
		}
	}
	if hits < 50 {
		t.Fatalf("only %d rays met the ball", hits)
	}
}

// On a gyroid whose gradient is about three times a distance field's, the
// emitted march meets the surface where a fine sampling of the field along the
// same ray first finds it: the gradient-corrected step does not tunnel.
func TestGpuMarchMeetsAGyroid(t *testing.T) {
	doc := `@scene(40)
g = 3.0 * (sin(p.x) * cos(p.y) + sin(p.y) * cos(p.z) + sin(p.z) * cos(p.x));
slab = max(g, length(p) - 12.0);
return slab;
`
	e := eval.New(pkgMgr, nil)
	pts := marchTargets(12, 10)
	res, out := marchGPU(t, e, doc, pts)

	// The reference: the field every `step` along each ray through the box, on the CPU.
	const step = 0.004
	half := res.Region[3]
	type span struct{ first, count int }
	spans := make([]span, len(pts))
	var samples [][3]float64
	var ts []float64
	for i, target := range pts {
		dir := marchRay(target)
		tEntry, tExit := math.Inf(-1), math.Inf(1)
		for k := 0; k < 3; k++ {
			a, b := (-half-marchEye[k])/dir[k], (half-marchEye[k])/dir[k]
			tEntry, tExit = math.Max(tEntry, math.Min(a, b)), math.Min(tExit, math.Max(a, b))
		}
		spans[i].first = len(samples)
		for tt := math.Max(tEntry, 0); tt < tExit; tt += step {
			samples = append(samples, [3]float64{marchEye[0] + tt*dir[0], marchEye[1] + tt*dir[1], marchEye[2] + tt*dir[2]})
			ts = append(ts, tt)
		}
		spans[i].count = len(samples) - spans[i].first
	}
	groups, err := e.EvalGsl(t.Context(), nil, doc, "", []string{"slab"}, samples)
	if err != nil {
		t.Fatal(err)
	}
	field := groups[0].Samples[0]

	hits, grazing := 0, 0
	for i := range pts {
		tWant, least := math.NaN(), math.Inf(1)
		for j := spans[i].first; j < spans[i].first+spans[i].count; j++ {
			v := float64(field[j].Value)
			least = math.Min(least, v)
			if v <= 0 {
				tWant = ts[j]
				break
			}
		}
		hit, tGPU := out[6*i] > 0.5, float64(out[6*i+1])
		// A ray that only comes within a hit tolerance of the surface may count as either.
		if math.IsNaN(tWant) && least < 0.1 {
			grazing++
			continue
		}
		if want := !math.IsNaN(tWant); hit != want {
			t.Errorf("ray %d: hit = %v, want %v (least field value %g)", i, hit, want, least)
			continue
		}
		if !hit {
			continue
		}
		hits++
		if math.Abs(tGPU-tWant) > 0.1 {
			t.Errorf("ray %d: t = %g, want %g: the march %s", i, tGPU, tWant,
				map[bool]string{true: "stepped through the surface", false: "stopped short"}[tGPU > tWant])
		}
	}
	if hits < 40 {
		t.Fatalf("only %d rays met the gyroid (%d grazing)", hits, grazing)
	}
}
