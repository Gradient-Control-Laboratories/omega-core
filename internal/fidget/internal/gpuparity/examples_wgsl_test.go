package gpuparity

import (
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/backend/wgsl"
	"gradientcontrol.com/omega/core/program/noise"
)

// The profile examples' WGSL, compiled as the viewer compiles it and run once
// through gpu-eval, whose naga validates it. The CPU tests evaluate the same
// programs but never compile WGSL, so a program WGSL refuses, a list indexed
// through a function's parameter or || beside && without brackets, passes them
// and fails only in the viewer. naga is not the browser's Tint, so this is a
// floor, not a proof. The values are not compared (TestGpuFieldParity does that
// for the goldens); only that each runs and is finite.

// sliderParamRE finds a slider's parameter in sdShape's signature: the codegen
// names it by the line and column of its declared value.
var sliderParamRE = regexp.MustCompile(`gsl_param_(\d+)_(\d+): f32`)

// leadingNumberRE is a number at the start of a slider's declared value.
var leadingNumberRE = regexp.MustCompile(`^-?[0-9]*\.?[0-9]+(?:[eE][-+]?[0-9]+)?`)

// sliderDefaults are the values to hand sdShape after time, one per slider
// parameter of sig, each read from src at the line and column its name
// records, or false when one is not a plain number there.
func sliderDefaults(sig, src string) ([]string, bool) {
	lines := strings.Split(src, "\n")
	var out []string
	for _, m := range sliderParamRE.FindAllStringSubmatch(sig, -1) {
		line, _ := strconv.Atoi(m[1])
		col, _ := strconv.Atoi(m[2])
		if line < 1 || line > len(lines) || col < 1 || col > len(lines[line-1]) {
			return nil, false
		}
		v := leadingNumberRE.FindString(lines[line-1][col-1:])
		if v == "" {
			return nil, false
		}
		f, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, false
		}
		out = append(out, strconv.FormatFloat(f, 'f', -1, 64)+wgslFloatSuffix(f))
	}
	return out, true
}

// wgslFloatSuffix makes an integral value read as an f32 literal.
func wgslFloatSuffix(f float64) string {
	if f == math.Trunc(f) {
		return ".0"
	}
	return ""
}

func TestProfileExamplesWGSLRunsOnTheGPU(t *testing.T) {
	bin := filepath.Join(repoRoot, "target/release/gpu-eval")
	if _, err := os.Stat(bin); err != nil {
		t.Skipf("gpu-eval is not built (tool/build-fidget builds it): %v", err)
	}
	host := readHostWGSL(t)
	noisePath := filepath.Join(t.TempDir(), "noise_tile.bin")
	if err := os.WriteFile(noisePath, noise.Bake(), 0o644); err != nil {
		t.Fatal(err)
	}
	files, err := filepath.Glob(filepath.Join(repoRoot, "examples/profile_*.gsl"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no profile examples found: %v", err)
	}
	rng := rand.New(rand.NewSource(1))
	pts := make([][3]float32, 4096)
	for i := range pts {
		pts[i] = [3]float32{float32(rng.Float64()*160 - 80), float32(rng.Float64()*160 - 80), float32(rng.Float64()*60 - 20)}
	}
	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".gsl")
		t.Run(name, func(t *testing.T) {
			src, err := os.ReadFile(file)
			if err != nil {
				t.Fatal(err)
			}
			res, err := assembler.Assemble([]string{string(src)}, assembler.SourceManager(pkgMgr))
			if err != nil {
				t.Fatalf("assemble: %v", err)
			}
			out, err := wgsl.Render(string(src), res.Program)
			if err != nil {
				t.Fatalf("render WGSL: %v", err)
			}
			sig := sdShapeRE.FindStringSubmatch(out.WGSL)
			if sig == nil {
				t.Fatal("no sdShape in the WGSL")
			}
			rest := sliderParamRE.ReplaceAllString(sig[1], "")
			usesNoise := strings.Contains(rest, "noiseTex: texture_3d<f32>")
			if strings.Contains(strings.ReplaceAll(rest, "noiseTex: texture_3d<f32>", ""), ":") {
				t.Skipf("sdShape binds what the harness does not provide:%s", sig[1])
			}
			args, ok := sliderDefaults(sig[0], string(src))
			if !ok {
				t.Skip("a slider's declared value is not a plain number")
			}
			// sdShape's parameters after time in their order: the sliders, and
			// the noise texture where it takes one.
			var call []string
			for _, param := range strings.Split(strings.TrimPrefix(sig[1], ", "), ", ") {
				switch {
				case param == "":
				case strings.HasPrefix(param, "noiseTex"):
					call = append(call, "gsl_noise")
				default:
					call = append(call, args[0])
					args = args[1:]
				}
			}
			extra := ""
			if len(call) > 0 {
				extra = ", " + strings.Join(call, ", ")
			}
			var b strings.Builder
			b.WriteString("@group(0) @binding(0) var<storage, read> gsl_points: array<vec4<f32>>;\n")
			b.WriteString("@group(0) @binding(1) var<storage, read_write> gsl_out: array<f32>;\n")
			if usesNoise {
				b.WriteString("@group(0) @binding(2) var gsl_noise: texture_3d<f32>;\n")
			}
			b.WriteString("var<private> gsl_fieldFilterIdx: u32 = 0u;\nvar<private> gsl_pw: f32 = 0.0;\n")
			b.WriteString("fn gsl_isShapeVisible(shapeIndex: u32) -> bool { return true; }\n")
			b.WriteString(host.shapeStruct + host.materialStruct + host.minFn)
			b.WriteString(out.WGSL)
			b.WriteString("\n@compute @workgroup_size(64)\nfn main(@builtin(global_invocation_id) gid: vec3<u32>) {\n")
			b.WriteString("    let i = gid.x;\n    if (i >= arrayLength(&gsl_points)) { return; }\n")
			b.WriteString("    gsl_out[i] = sdShape(gsl_points[i].xyz, 0.0" + extra + ").distance;\n}\n")
			got := runGPU(t, bin, b.String(), pts, 1, usesNoise, noisePath)
			bad := 0
			for _, v := range got {
				if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
					bad++
				}
			}
			if bad > 0 {
				t.Errorf("%d of %d values are not finite", bad, len(got))
			}
		})
	}
}
