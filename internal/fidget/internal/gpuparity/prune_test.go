package gpuparity

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/eval"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/source"
)

// Shaft Modeling, Phase 19: a pruned feature, run on the GPU. The viewer's
// WGSL measures the pixel at a point from the world position and the eye, and
// a feature (name.featureSize) too small at that pixel gives way: a hole
// closes, a gear's teeth leave its pitch cylinder. The part itself, larger,
// stays. The band is vanish's, set here as the viewer's Detail slider sets it:
// 1 to 2 pixels.

// pruneShader wraps a `gsl eval shader` response as a compute shader that sets
// the camera the guards read, eye and pixel scale (w <= 0 for none), and
// writes the field at each point.
func pruneShader(t *testing.T, res *eval.ShaderResult, eye [3]float64, pixel float64) string {
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
	fmt.Fprintf(&b, "    gsl_vanishEye = vec4<f32>(%v, %v, %v, %v);\n", eye[0], eye[1], eye[2], pixel)
	b.WriteString("    gsl_vanishView = vec4<f32>(0.0, 1.0, 2.0, 0.0);\n")
	fmt.Fprintf(&b, "    gsl_out[i] = %s(gsl_points[i].xyz).distance;\n", res.FieldFn)
	b.WriteString("}\n")
	return b.String()
}

func pruneGPU(t *testing.T, doc string, eye [3]float64, pixel float64, pts [][3]float32) []float32 {
	t.Helper()
	bin := filepath.Join(repoRoot, "target/release/gpu-eval")
	if _, err := os.Stat(bin); err != nil {
		t.Skipf("gpu-eval is not built (tool/build-fidget builds it): %v", err)
	}
	e := eval.New(&source.MockManager{}, nil)
	res, err := e.Shader(t.Context(), nil, doc, "", eval.ShaderOptions{Target: eval.ShaderWGSL})
	if err != nil {
		t.Fatal(err)
	}
	return runGPU(t, bin, pruneShader(t, res, eye, pixel), pts, 1, false, "")
}

// A plate 40 x 40 x 10 with a hole 1 across, a feature. With a pixel scale of
// 0.001, from 50 away the hole is 20 pixels and stays; from 1500 it is 0.67
// pixels and closes, while the plate, 10 thick, is 6.7 pixels and stays.
const prunePlateDoc = "plate = box([-20.0, -20.0, -5.0], [20.0, 20.0, 5.0]);\nhole = length(p.xy) - 0.5;\nhole.featureSize = 1.0;\nplated = max(plate, neg(hole));\nreturn plated;\n"

func TestGpuPrune_HoleClosesWhenSmall(t *testing.T) {
	pts := [][3]float32{{0, 0, 0}}
	for _, c := range []struct {
		name  string
		eye   [3]float64
		pixel float64
		open  bool
	}{
		{"near", [3]float64{0, 0, 50}, 0.001, true},
		{"far", [3]float64{0, 0, 1500}, 0.001, false},
		{"no camera", [3]float64{0, 0, 1500}, 0, true},
	} {
		d := float64(pruneGPU(t, prunePlateDoc, c.eye, c.pixel, pts)[0])
		if c.open && math.Abs(d-0.5) > 1e-3 {
			t.Errorf("%s: the hole's middle reads %g, want 0.5 (open)", c.name, d)
		}
		if !c.open && d > -4 {
			t.Errorf("%s: the hole's middle reads %g, want the plate's -5 (closed)", c.name, d)
		}
	}
}

// A 60-tooth module-2 gear 20 thick: its teeth, pi m / 2 = 3.14 across, are
// 0.63 pixels from 5000 away at a pixel scale of 0.001 and give way, leaving
// the pitch cylinder, radius 60: at mid-face the field is r - 60 at any angle.
const pruneGearDoc = "g = max(gears.gearProfile(60.0, 2.0, 0.3490658503988659, 0.0, 0.25, 0.1), abs(p.z - 10.0) - 10.0);\nreturn g;\n"

func TestGpuPrune_TeethLeaveThePitchCylinder(t *testing.T) {
	var pts [][3]float32
	var radii []float64
	for _, r := range []float64{58.8, 59.4, 60.6, 61.2} {
		for j := 0; j < 12; j++ {
			a := 2 * math.Pi / 60 * float64(j) / 12
			pts = append(pts, [3]float32{float32(r * math.Cos(a)), float32(r * math.Sin(a)), 10})
			radii = append(radii, r)
		}
	}
	far := pruneGPU(t, pruneGearDoc, [3]float64{0, 0, 5010}, 0.001, pts)
	for i, d := range far {
		if want := radii[i] - 60; math.Abs(float64(d)-want) > 2e-3 {
			t.Errorf("far, at r %g: %g, want the pitch cylinder's %g", radii[i], d, want)
		}
	}
	near := pruneGPU(t, pruneGearDoc, [3]float64{0, 0, 60}, 0.001, pts)
	differs := false
	for i, d := range near {
		if math.Abs(float64(d)-(radii[i]-60)) > 0.1 {
			differs = true
		}
	}
	if !differs {
		t.Errorf("near, the teeth were pruned too: every point reads the pitch cylinder")
	}
}
