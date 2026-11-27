package wgsl

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/assembler"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/source"
	"github.com/Gradient-Control-Laboratories/omega-core/program/noise/spinodoid"
)

// The reference spinodoid's WGSL, run on the GPU over the table a preview
// binds, against the Go reference table it was packed from: the value against
// ReferenceGaborTable.Distance, the gradient against central differences of
// it. The table is baked here in Go, from the same arguments the document
// passes, and packed by PackSpinodoidBinding as a preview packs it. It runs
// through gpu-eval (GPU_EVAL, or target/release/gpu-eval) and skips without
// it or a GPU adapter.

const spinodoidGPUUpright = "upright = [[1, 0, 0], [0, 1, 0], [0, 0, 1]];\n"

func referenceParams() spinodoid.Params {
	return spinodoid.Params{
		Period:      6,
		Theta:       []float64{math.Pi / 2, math.Pi / 2, math.Pi / 2},
		Poles:       [][3]float64{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}},
		PacketWidth: 3,
		Impulses:    8,
		Pool:        4096,
	}
}

func TestReferenceSpinodoidOnTheGPUMatchesTheGoTable(t *testing.T) {
	bin := gpuEvalBinary(t)

	stationary := referenceParams()
	graded := referenceParams()
	graded.Seed = 3
	graded.Scale = func(cs [][3]float64) ([]float64, error) {
		out := make([]float64, len(cs))
		for i, c := range cs {
			out[i] = 1 + 0.004*c[0]
		}
		return out, nil
	}
	box := [6]float64{-20, -20, -20, 20, 20, 20}
	graded.Domain = &box

	cases := []struct {
		name   string
		src    string
		p      spinodoid.Params
		level  float64
		half   float64
		value  bool
		weight [3]float64
	}{
		{
			name:  "stationary",
			src:   "return gaborSpinodoid(p, 6, [pi / 2, pi / 2, pi / 2], upright, 3, 8, 4096, 0, 0.3, 1, 1, upright, [[0, 0, 0], [0, 0, 0]]);",
			p:     stationary,
			level: 0.3, half: 40, value: true,
		},
		{
			name:  "graded",
			src:   "return gaborSpinodoid(p, 6, [pi / 2, pi / 2, pi / 2], upright, 3, 8, 4096, 3, 0, 1, 1 + 0.004 * p.x, upright, [[-20, -20, -20], [20, 20, 20]]);",
			p:     graded,
			level: 0, half: 26, value: true,
		},
		{
			name:  "gradient",
			src:   "return dot(gaborSpinodoidGradient(p, 6, [pi / 2, pi / 2, pi / 2], upright, 3, 8, 4096, 0, 0.3, 1, 1, upright, [[0, 0, 0], [0, 0, 0]]), [1, 2, 4]);",
			p:     stationary,
			level: 0.3, half: 40,
			weight: [3]float64{1, 2, 4},
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			table, err := spinodoid.ReferenceGaborBake(c.p)
			if err != nil {
				t.Fatal(err)
			}
			src := spinodoidGPUUpright + c.src
			res, err := assembler.Assemble([]string{src}, assembler.SourceManager(&source.MockManager{}))
			if err != nil {
				t.Fatalf("assemble: %v", err)
			}
			out, err := Render(src, res.Program)
			if err != nil {
				t.Fatalf("render: %v", err)
			}
			if len(out.SpinodoidSlots) != 1 {
				t.Fatalf("want one spinodoid slot, got %d", len(out.SpinodoidSlots))
			}
			texels, dims := packOneTable(t, out.SpinodoidSlots[0], spinodoid.Baked{Reference: table, ContentID: c.name})

			rng := rand.New(rand.NewSource(7))
			pts := make([][3]float32, 2048)
			for i := range pts {
				for a := range 3 {
					pts[i][a] = float32((rng.Float64()*2 - 1) * c.half)
				}
			}
			got := runSpinodoidShader(t, bin, out.WGSL, pts, texels, dims)

			worst, at := 0.0, 0
			for i, q32 := range pts {
				q := [3]float64{float64(q32[0]), float64(q32[1]), float64(q32[2])}
				var want float64
				if c.value {
					want = table.Distance(q, c.level, true)
				} else {
					const h = 1e-4
					for a := range 3 {
						hi, lo := q, q
						hi[a] += h
						lo[a] -= h
						d := (table.Distance(hi, c.level, true) - table.Distance(lo, c.level, true)) / (2 * h)
						want += c.weight[a] * d
					}
				}
				if e := math.Abs(float64(got[i]) - want); e > worst {
					worst, at = e, i
				}
			}
			// order one, more for the gradient's sum of seven.
			tol := 1e-4
			if !c.value {
				tol = 1e-3
			}
			if worst > tol {
				t.Errorf("worst difference %g at %v (GPU %v), want <= %g", worst, pts[at], got[at], tol)
			}
			t.Logf("%d points, worst difference %g", len(pts), worst)
		})
	}
}

func gpuEvalBinary(t *testing.T) string {
	t.Helper()
	bin := os.Getenv("GPU_EVAL")
	if bin == "" {
		bin = filepath.Join("..", "..", "target", "release", "gpu-eval")
	}
	if _, err := os.Stat(bin); err != nil {
		t.Skipf("gpu-eval is not built (tool/build-fidget builds it, or set GPU_EVAL): %v", err)
	}
	return bin
}

// packOneTable packs one baked table into the preview's texture, as a
// preview does: the table in its GPU layout under the slot's key, then
// PackSpinodoidBinding.
func packOneTable(t *testing.T, slot SpinodoidSlot, baked spinodoid.Baked) ([]byte, [2]int) {
	t.Helper()
	bt := baked.GPUTable(slot.Key)
	raw, err := base64.StdEncoding.DecodeString(bt.TexelsB64)
	if err != nil {
		t.Fatal(err)
	}
	binding, err := PackSpinodoidBinding([]SpinodoidSlot{slot},
		map[string]spinodoid.BakedTable{slot.Key: bt},
		map[string][]byte{bt.ContentID: raw}, "")
	if err != nil {
		t.Fatal(err)
	}
	texels, err := base64.StdEncoding.DecodeString(binding.TexelsB64)
	if err != nil {
		t.Fatal(err)
	}
	return texels, binding.Dims
}

var spinodoidSdShapeRE = regexp.MustCompile(`(?m)^fn sdShape\(p: vec3<f32>, time: f32(.*)\) -> gsl_shape \{`)

// runSpinodoidShader wraps the document's WGSL in a compute shader that
// writes sdShape's distance at each point, with the spinodoid texture bound.
func runSpinodoidShader(t *testing.T, bin, wgsl string, pts [][3]float32, texels []byte, dims [2]int) []float32 {
	t.Helper()
	sig := spinodoidSdShapeRE.FindStringSubmatch(wgsl)
	if sig == nil {
		t.Fatal("no sdShape in the WGSL")
	}
	if sig[1] != ", spinodoidTex: texture_2d<f32>" {
		t.Fatalf("sdShape takes %q, want only the spinodoid texture", sig[1])
	}
	read := func(file, name string) string {
		src, err := os.ReadFile(filepath.Join("..", "..", "examples", "viewer", "render", file))
		if err != nil {
			t.Fatal(err)
		}
		m := regexp.MustCompile("(?s)export const " + name + " = `([^`]*)`;").FindSubmatch(src)
		if m == nil {
			t.Fatalf("%s: no template literal %s", file, name)
		}
		return string(m[1])
	}
	var b strings.Builder
	b.WriteString("@group(0) @binding(0) var<storage, read> gsl_points: array<vec4<f32>>;\n")
	b.WriteString("@group(0) @binding(1) var<storage, read_write> gsl_out: array<f32>;\n")
	b.WriteString("@group(0) @binding(3) var gsl_tex: texture_2d<f32>;\n")
	b.WriteString("var<private> gsl_fieldFilterIdx: u32 = 0u;\nvar<private> gsl_pw: f32 = 0.0;\n")
	b.WriteString("fn gsl_isShapeVisible(shapeIndex: u32) -> bool { return true; }\n")
	b.WriteString(read("gsl-host.js", "GSL_SHAPE_STRUCT_WGSL"))
	b.WriteString(read("gsl-shading.js", "GSL_MATERIAL_STRUCT_WGSL"))
	b.WriteString(read("gsl-host.js", "GSL_MIN_WGSL"))
	b.WriteString(wgsl)
	b.WriteString("\n@compute @workgroup_size(64)\nfn main(@builtin(global_invocation_id) gid: vec3<u32>) {\n")
	b.WriteString("    let i = gid.x;\n    if (i >= arrayLength(&gsl_points)) { return; }\n")
	b.WriteString("    gsl_out[i] = sdShape(gsl_points[i].xyz, 0.0, gsl_tex).distance;\n}\n")

	dir := t.TempDir()
	shader := filepath.Join(dir, "shader.wgsl")
	points := filepath.Join(dir, "points.f32")
	tex := filepath.Join(dir, "tex.f32")
	outPath := filepath.Join(dir, "out.f32")
	if err := os.WriteFile(shader, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 16*len(pts))
	for i, p := range pts {
		for a := range 3 {
			binary.LittleEndian.PutUint32(buf[16*i+4*a:], math.Float32bits(p[a]))
		}
	}
	if err := os.WriteFile(points, buf, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(tex, texels, 0o644); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(bin, "--shader", shader, "--points", points, "--stride", "1", "--out", outPath,
		"--texture2d", tex, "--texture2d-dims", strconv.Itoa(dims[0])+","+strconv.Itoa(dims[1]))
	msg, err := cmd.CombinedOutput()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 3 {
		t.Skipf("no GPU adapter: %s", msg)
	}
	if err != nil {
		t.Fatalf("gpu-eval: %v\n%s", err, msg)
	}
	raw, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) != 4*len(pts) {
		t.Fatalf("gpu-eval wrote %d bytes, want %d", len(raw), 4*len(pts))
	}
	vals := make([]float32, len(pts))
	for i := range vals {
		vals[i] = math.Float32frombits(binary.LittleEndian.Uint32(raw[4*i:]))
	}
	return vals
}
