package gpuparity

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/eval"
	"gradientcontrol.com/omega/core/internal/fidget"
	"gradientcontrol.com/omega/core/internal/fidget/convert"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/program/noise"
)

const repoRoot = "../../../.."

var (
	pkgMgr = source.NewLocalManager(filepath.Join(repoRoot, "stdlib"))

	sceneSizeRE = regexp.MustCompile(`(?mi)^\s*//\s*Scene Size:\s*([0-9.eE+-]+)`)
	sdShapeRE   = regexp.MustCompile(`(?m)^fn sdShape\(p: vec3<f32>, time: f32(.*)\) -> gsl_shape \{`)
	shapeIdxRE  = regexp.MustCompile(`gsl_shape\([^;]*?, (\d+)u\)`)
)

// The audio examples do not build in fidget (the JS path strips their audio
// returns); TestBuildExamples skips them too, so they have no golden.
var skipExamples = map[string]string{
	"Barberpole":      "audio example",
	"Vibrating Plate": "audio example",
}

// Tolerances, f32 against f32. A distance is allowed an absolute error that
// scales with the scene, since coordinates of that size lose that much
// precision, plus a relative one for large values. Colors and material rows
// are bounded quantities.
const (
	distAbsTol  = 2e-5 // times the scene half-size
	distRelTol  = 1e-4
	colorTol    = 1e-3
	materialTol = 1e-3 // relative, floor 1
)

const (
	gridPerAxis     = 12
	seedsPerShape   = 400
	newtonSteps     = 3
	nearSurfaceFrac = 0.02 // of the half-size
	jumpRadius      = 1e-4 // of the half-size; see jumpExplains
)

func TestGpuFieldParity(t *testing.T) {
	bin := filepath.Join(repoRoot, "target/release/gpu-eval")
	if _, err := os.Stat(bin); err != nil {
		t.Skipf("gpu-eval is not built (tool/build-fidget builds it): %v", err)
	}
	host := readHostWGSL(t)

	tile := noise.Bake()
	noisePath := filepath.Join(t.TempDir(), "noise_tile.bin")
	if err := os.WriteFile(noisePath, tile, 0o644); err != nil {
		t.Fatal(err)
	}

	files, err := filepath.Glob(filepath.Join(repoRoot, "internal/testdata/example/*.gsl"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no examples found: %v", err)
	}
	for _, file := range files {
		name := strings.TrimSuffix(filepath.Base(file), ".gsl")
		t.Run(name, func(t *testing.T) {
			if why, ok := skipExamples[name]; ok {
				t.Skip(why)
			}
			checkExample(t, bin, host, noisePath, file, name)
		})
	}
}

// hostWGSL is what the viewers declare around a compiled program, read from
// the modules that hold it so the test cannot drift from the page.
type hostWGSL struct {
	shapeStruct, materialStruct, minFn string
}

func readHostWGSL(t *testing.T) hostWGSL {
	t.Helper()
	read := func(file, name string) string {
		src, err := os.ReadFile(filepath.Join(repoRoot, "examples/viewer/render", file))
		if err != nil {
			t.Fatal(err)
		}
		re := regexp.MustCompile("(?s)export const " + name + " = `([^`]*)`;")
		m := re.FindSubmatch(src)
		if m == nil {
			t.Fatalf("%s: no plain template literal %s", file, name)
		}
		if strings.Contains(string(m[1]), "${") {
			t.Fatalf("%s: %s interpolates; the test cannot extract it", file, name)
		}
		return string(m[1])
	}
	return hostWGSL{
		shapeStruct:    read("gsl-host.js", "GSL_SHAPE_STRUCT_WGSL"),
		materialStruct: read("gsl-shading.js", "GSL_MATERIAL_STRUCT_WGSL"),
		minFn:          read("gsl-host.js", "GSL_MIN_WGSL"),
	}
}

func checkExample(t *testing.T, bin string, host hostWGSL, noisePath, file, name string) {
	src, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	golden, err := os.ReadFile(filepath.Join(repoRoot, "backend/wgsl/internal/codegen/testdata/example", name+".wgsl"))
	if errors.Is(err, os.ErrNotExist) {
		t.Skip("no WGSL golden")
	} else if err != nil {
		t.Fatal(err)
	}
	wgsl := string(golden)

	sig := sdShapeRE.FindStringSubmatch(wgsl)
	if sig == nil {
		t.Fatal("golden has no sdShape")
	}
	usesNoise := false
	switch sig[1] {
	case "":
	case ", noiseTex: texture_3d<f32>":
		usesNoise = true
	default:
		t.Skipf("sdShape binds textures the harness does not provide:%s", sig[1])
	}
	hasColor := strings.Contains(wgsl, "fn gsl_colorField(")
	hasMaterial := strings.Contains(wgsl, "fn gsl_materialField(")

	half := 10.0
	if m := sceneSizeRE.FindSubmatch(src); m != nil {
		if v, err := strconv.ParseFloat(string(m[1]), 64); err == nil && v > 0 {
			half = v / 2
		}
	}

	res, err := assembler.Assemble([]string{string(src)}, assembler.SourceManager(pkgMgr))
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}

	var rep *fieldReport
	err = convert.ToColorShapes(res.Program, 0, func(shapes []convert.ColorShape) error {
		n := len(shapes)
		if want := maxShapeIndex(wgsl); want != n {
			t.Skipf("fidget has %d outputs but the WGSL numbers %d shapes", n, want)
		}
		cpu := newCPUSide(shapes, hasColor, hasMaterial)
		pts := samplePoints(t, cpu, half, sampleSeed(name))
		want := cpu.eval(t, pts)
		shader := assembleShader(host, wgsl, n, usesNoise, hasColor, hasMaterial)
		got := runGPU(t, bin, shader, pts, stride(n, hasColor, hasMaterial), usesNoise, noisePath)
		rep = compare(t, cpu, pts, want, got, n, half, hasColor, hasMaterial)
		return nil
	})
	if err != nil {
		t.Fatalf("convert: %v", err)
	}
	// Painted ties are settled once convert has let go of fidget, since
	// gsleval takes it too.
	rep.finish(t, func(pts [][3]float32) []bool { return paintedTies(t, string(src), half, pts) })
}

// paintedRE finds the values a document paints (name.color = … or
// name.material = …): a return's colour and material are those of whichever of
// them is nearest.
var paintedRE = regexp.MustCompile(`(?m)^\s*([A-Za-z_]\w*)\.(?:color|material)\s*=`)

// paintedTies reports, for each point, whether the two painted values nearest
// it are as near as each other, within the distance tolerance. There the
// colour and material are those of either, and float32 rounding picks, as it
// does the composite's shape (see compare). Two parts resting on the same
// plane are that close everywhere under both, so moving the point, as
// jumpExplains does, does not separate them. Every painted value in the
// document is considered, not only those under one return.
func paintedTies(t *testing.T, code string, half float64, pts [][3]float32) []bool {
	ties := make([]bool, len(pts))
	seen := map[string]bool{}
	var names []string
	for _, m := range paintedRE.FindAllStringSubmatch(code, -1) {
		if !seen[m[1]] {
			seen[m[1]] = true
			names = append(names, m[1])
		}
	}
	if len(names) < 2 || len(pts) == 0 {
		return ties
	}
	at := make([][3]float64, len(pts))
	for i, p := range pts {
		at[i] = [3]float64{float64(p[0]), float64(p[1]), float64(p[2])}
	}
	groups, err := eval.New(pkgMgr, nil).EvalGsl(context.Background(), nil, code, "", names, at)
	if err != nil || len(groups) != len(names) {
		t.Logf("painted values not evaluated, so no mismatch is taken for a tie: %v", err)
		return ties
	}
	for i := range pts {
		best, second := math.Inf(1), math.Inf(1)
		for _, g := range groups {
			if len(g.Samples) != 1 || len(g.Samples[0]) != len(pts) {
				continue
			}
			a := math.Abs(float64(g.Samples[0][i].Value))
			if a < best {
				best, second = a, best
			} else if a < second {
				second = a
			}
		}
		ties[i] = second-best <= 2*distTolAt(float32(best), half)
	}
	return ties
}

// distTolAt is the distance tolerance at a value d, for a scene of half-size
// half.
func distTolAt(d float32, half float64) float64 {
	return distAbsTol*half + distRelTol*math.Abs(float64(d))
}

// maxShapeIndex is the largest shapeIndex sdShape assigns, which is the
// number of returns it composites.
func maxShapeIndex(wgsl string) int {
	start := sdShapeRE.FindStringIndex(wgsl)
	body := wgsl[start[1]:]
	if end := strings.Index(body, "\n}\n"); end >= 0 {
		body = body[:end]
	}
	best := 0
	for _, m := range shapeIdxRE.FindAllStringSubmatch(body, -1) {
		if v, err := strconv.Atoi(m[1]); err == nil && v > best {
			best = v
		}
	}
	return best
}

// Output layout per point, in f32:
//
//	0        composite distance
//	1        composite shapeIndex
//	2..2+n   each shape's own distance (gsl_fieldFilterIdx = k)
//	then     4 per shape: gsl_colorField rgba          (when there is one)
//	then     13 per shape: material rows + present     (when there is one)
func stride(n int, hasColor, hasMaterial bool) int {
	s := 2 + n
	if hasColor {
		s += 4 * n
	}
	if hasMaterial {
		s += 17 * n
	}
	return s
}

func assembleShader(host hostWGSL, golden string, n int, usesNoise, hasColor, hasMaterial bool) string {
	var b strings.Builder
	b.WriteString("@group(0) @binding(0) var<storage, read> gsl_points: array<vec4<f32>>;\n")
	b.WriteString("@group(0) @binding(1) var<storage, read_write> gsl_out: array<f32>;\n")
	extra := ""
	if usesNoise {
		b.WriteString("@group(0) @binding(2) var gsl_noise: texture_3d<f32>;\n")
		extra = ", gsl_noise"
	}
	b.WriteString("var<private> gsl_fieldFilterIdx: u32 = 0u;\n")
	// pixelSize() reads the pixel footprint; 0 is full detail, as in fidget.
	b.WriteString("var<private> gsl_pw: f32 = 0.0;\n")
	b.WriteString("fn gsl_isShapeVisible(shapeIndex: u32) -> bool { return true; }\n")
	b.WriteString(host.shapeStruct)
	b.WriteString(host.materialStruct)
	b.WriteString(host.minFn)
	b.WriteString(golden)
	b.WriteString("\n@compute @workgroup_size(64)\n")
	b.WriteString("fn main(@builtin(global_invocation_id) gid: vec3<u32>) {\n")
	b.WriteString("    let i = gid.x;\n")
	b.WriteString("    if (i >= arrayLength(&gsl_points)) { return; }\n")
	b.WriteString("    let p = gsl_points[i].xyz;\n")
	fmt.Fprintf(&b, "    let o = i * %du;\n", stride(n, hasColor, hasMaterial))
	fmt.Fprintf(&b, "    let s = sdShape(p, 0.0%s);\n", extra)
	b.WriteString("    gsl_out[o] = s.distance;\n")
	b.WriteString("    gsl_out[o + 1u] = f32(s.shapeIndex);\n")
	for k := 1; k <= n; k++ {
		fmt.Fprintf(&b, "    gsl_fieldFilterIdx = %du;\n", k)
		fmt.Fprintf(&b, "    gsl_out[o + %du] = sdShape(p, 0.0%s).distance;\n", 1+k, extra)
	}
	b.WriteString("    gsl_fieldFilterIdx = 0u;\n")
	off := 2 + n
	if hasColor {
		for k := 1; k <= n; k++ {
			fmt.Fprintf(&b, "    let c%d = gsl_colorField(p, 0.0, %du%s);\n", k, k, extra)
			for c, comp := range []string{"x", "y", "z", "w"} {
				fmt.Fprintf(&b, "    gsl_out[o + %du] = c%d.%s;\n", off+4*(k-1)+c, k, comp)
			}
		}
		off += 4 * n
	}
	if hasMaterial {
		for k := 1; k <= n; k++ {
			fmt.Fprintf(&b, "    let m%d = gsl_materialField(p, 0.0, %du%s);\n", k, k, extra)
			base := off + 17*(k-1)
			for r, row := range []string{"surface", "emissive", "volume", "lay"} {
				for c, comp := range []string{"x", "y", "z", "w"} {
					fmt.Fprintf(&b, "    gsl_out[o + %du] = m%d.%s.%s;\n", base+4*r+c, k, row, comp)
				}
			}
			fmt.Fprintf(&b, "    gsl_out[o + %du] = m%d.present;\n", base+16, k)
		}
	}
	b.WriteString("}\n")
	return b.String()
}

func runGPU(t *testing.T, bin, shader string, pts [][3]float32, stride int, usesNoise bool, noisePath string) []float32 {
	t.Helper()
	dir := t.TempDir()
	shaderPath := filepath.Join(dir, "shader.wgsl")
	pointsPath := filepath.Join(dir, "points.f32")
	outPath := filepath.Join(dir, "out.f32")
	if err := os.WriteFile(shaderPath, []byte(shader), 0o644); err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 16*len(pts))
	for i, p := range pts {
		for c := 0; c < 3; c++ {
			binary.LittleEndian.PutUint32(buf[16*i+4*c:], math.Float32bits(p[c]))
		}
	}
	if err := os.WriteFile(pointsPath, buf, 0o644); err != nil {
		t.Fatal(err)
	}
	args := []string{"--shader", shaderPath, "--points", pointsPath, "--stride", strconv.Itoa(stride), "--out", outPath}
	if usesNoise {
		args = append(args, "--noise", noisePath)
	}
	out, err := exec.Command(bin, args...).CombinedOutput()
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) && exitErr.ExitCode() == 3 {
		t.Skipf("no GPU adapter: %s", out)
	}
	if err != nil {
		t.Fatalf("gpu-eval: %v\n%s\nshader kept at %s", err, out, keepShader(t, shader))
	}
	raw, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(raw) != 4*stride*len(pts) {
		t.Fatalf("gpu-eval wrote %d bytes, want %d", len(raw), 4*stride*len(pts))
	}
	vals := make([]float32, len(raw)/4)
	for i := range vals {
		vals[i] = math.Float32frombits(binary.LittleEndian.Uint32(raw[4*i:]))
	}
	return vals
}

// keepShader saves a failing shader outside the test's temp dir, which is
// removed when the test ends, so it can be read afterwards.
func keepShader(t *testing.T, shader string) string {
	f, err := os.CreateTemp("", "gpuparity-*.wgsl")
	if err != nil {
		return "(could not save: " + err.Error() + ")"
	}
	defer f.Close()
	_, _ = f.WriteString(shader)
	return f.Name()
}

// cpuSide is the fidget half: one canvas per tree the WGSL side outputs.
type cpuSide struct {
	dist     []fidget.CanvasIndex
	color    [][3]fidget.CanvasIndex
	hasColor []bool
	mat      [][4][4]fidget.CanvasIndex
	hasMat   []bool

	wantColor, wantMaterial bool
}

func newCPUSide(shapes []convert.ColorShape, wantColor, wantMaterial bool) *cpuSide {
	c := &cpuSide{wantColor: wantColor, wantMaterial: wantMaterial}
	for _, s := range shapes {
		c.dist = append(c.dist, fidget.BuildCanvas(s.Tree))
		var col [3]fidget.CanvasIndex
		if s.HasColor {
			for i := range col {
				col[i] = fidget.BuildCanvas(s.Color[i])
			}
		}
		c.color = append(c.color, col)
		c.hasColor = append(c.hasColor, s.HasColor)
		var m [4][4]fidget.CanvasIndex
		if s.HasMaterial {
			for r := range m {
				for k := range m[r] {
					m[r][k] = fidget.BuildCanvas(s.Material[r][k])
				}
			}
		}
		c.mat = append(c.mat, m)
		c.hasMat = append(c.hasMat, s.HasMaterial)
	}
	return c
}

func split(pts [][3]float32) (xs, ys, zs []float32) {
	xs = make([]float32, len(pts))
	ys = make([]float32, len(pts))
	zs = make([]float32, len(pts))
	for i, p := range pts {
		xs[i], ys[i], zs[i] = p[0], p[1], p[2]
	}
	return
}

func evalCanvas(t *testing.T, c fidget.CanvasIndex, pts [][3]float32) []fidget.GradResult {
	t.Helper()
	xs, ys, zs := split(pts)
	r, err := fidget.EvalGrad(c, xs, ys, zs)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// sampleSeed seeds an example's sample points with the length of its name,
// or, for an example renamed since, of the name it had, so a rename leaves its
// points where they were.
func sampleSeed(name string) int64 {
	if old, ok := renamedExamples[name]; ok {
		return int64(len(old))
	}
	return int64(len(name))
}

// renamedExamples maps each renamed example to the name it had before.
var renamedExamples = map[string]string{
	"Material - Basics":                        "Materials",
	"Material - Glass":                         "Glass",
	"Material - Displacement":                  "Displacement",
	"Material - Physical":                      "Physical Materials",
	"Material - Metals (Noble and Transition)": "Physical Metals - Noble and Transition",
	"Material - Metals (Light and Reactive)":   "Physical Metals - Light and Reactive",
	"Material - Plastics and Gems":             "Physical Plastics and Gems",
	"Material - Surfaces":                      "Physical Surfaces",
	"Material - Liquids":                       "Physical Liquids",
	"Material - Marble and Granite":            "Marble and Granite",
	"Material - Dark Granite":                  "Dark Granite",
	"Material - Concrete":                      "Concrete Test Cube",
	"Material - Brick":                         "Brick Wall",
	"Material - Canyon Strata":                 "Canyon Strata",
	"Material - Cracked Mud":                   "Cracked Mud",
	"Material - Rust and Patina":               "Rusted Cannonball",
	"Material - Wood Planks":                   "Wood Planks",
	"Material - Wood Species":                  "Wood Species",
	"Material - Wood Finishes":                 "Wood Finishes",
	"Material - Knotty Planks":                 "Knotty Planks",
	"Material - Spalted and Wormy":             "Spalted and Wormy",
	"Material - Sea and River":                 "Sea and River",
	"Material - Lava and Ice":                  "Lava and Ice",
	"Material - Rain and Mud":                  "Rain and Mud",
	"Material - Sand Ripples":                  "Sand Ripples",
	"Material - Playful in Time":               "Playful in Time",
	"Material - Library":                       "Texture Library",
	"Material Test - Granite":                  "Texture Test - Granite",
	"Material Test - Lava":                     "Texture Test - Lava",
	"Material Test - Marble":                   "Texture Test - Marble",
	"Material Test - No Displacement":          "Texture Test - No Displacement",
	"Material Test - Playful":                  "Texture Test - Playful",
	"Material Test - Water":                    "Texture Test - Water",
	"Material Test - Weathering":               "Texture Test - Weathering",
	"Material Test - Wood":                     "Texture Test - Wood",
}

// samplePoints is a jittered grid over the scene box plus points pulled onto
// each shape's surface by a few Newton steps along fidget's gradient, where a
// sign or a branch is most likely to disagree.
func samplePoints(t *testing.T, cpu *cpuSide, half float64, seed int64) [][3]float32 {
	rng := rand.New(rand.NewSource(seed))
	var pts [][3]float32
	cell := 2 * half / gridPerAxis
	for i := 0; i < gridPerAxis; i++ {
		for j := 0; j < gridPerAxis; j++ {
			for k := 0; k < gridPerAxis; k++ {
				pts = append(pts, [3]float32{
					float32(-half + (float64(i)+rng.Float64())*cell),
					float32(-half + (float64(j)+rng.Float64())*cell),
					float32(-half + (float64(k)+rng.Float64())*cell),
				})
			}
		}
	}
	for _, c := range cpu.dist {
		seeds := make([][3]float32, seedsPerShape)
		for i := range seeds {
			for a := 0; a < 3; a++ {
				seeds[i][a] = float32((rng.Float64()*2 - 1) * half)
			}
		}
		for step := 0; step < newtonSteps; step++ {
			r := evalCanvas(t, c, seeds)
			for i, g := range r {
				g2 := g.Dx*g.Dx + g.Dy*g.Dy + g.Dz*g.Dz
				if g2 < 1e-12 || !finite(g.Value) || !finite(g2) {
					continue
				}
				s := g.Value / g2
				seeds[i][0] -= s * g.Dx
				seeds[i][1] -= s * g.Dy
				seeds[i][2] -= s * g.Dz
			}
		}
		r := evalCanvas(t, c, seeds)
		for i, g := range r {
			p := seeds[i]
			inBox := math.Abs(float64(p[0])) <= half && math.Abs(float64(p[1])) <= half && math.Abs(float64(p[2])) <= half
			if inBox && finite(g.Value) && math.Abs(float64(g.Value)) < nearSurfaceFrac*half {
				pts = append(pts, p)
			}
		}
	}
	return pts
}

func finite(v float32) bool {
	f := float64(v)
	return !math.IsNaN(f) && !math.IsInf(f, 0)
}

// cpuValues are fidget's answers, indexed [shape][point].
type cpuValues struct {
	dist  [][]float32
	color [][][3]float32
	mat   [][][16]float32

	hasColor, hasMat []bool
}

func (v cpuValues) colorAuthored(k int) bool    { return v.hasColor[k] }
func (v cpuValues) materialAuthored(k int) bool { return v.hasMat[k] }

func (c *cpuSide) eval(t *testing.T, pts [][3]float32) cpuValues {
	v := cpuValues{hasColor: c.hasColor, hasMat: c.hasMat}
	for k := range c.dist {
		d := evalCanvas(t, c.dist[k], pts)
		row := make([]float32, len(pts))
		for i := range d {
			row[i] = d[i].Value
		}
		v.dist = append(v.dist, row)

		cols := make([][3]float32, len(pts))
		if c.wantColor && c.hasColor[k] {
			for ch := 0; ch < 3; ch++ {
				r := evalCanvas(t, c.color[k][ch], pts)
				for i := range r {
					cols[i][ch] = r[i].Value
				}
			}
		}
		v.color = append(v.color, cols)

		mats := make([][16]float32, len(pts))
		if c.wantMaterial && c.hasMat[k] {
			for row := 0; row < 4; row++ {
				for comp := 0; comp < 4; comp++ {
					r := evalCanvas(t, c.mat[k][row][comp], pts)
					for i := range r {
						mats[i][4*row+comp] = r[i].Value
					}
				}
			}
		}
		v.mat = append(v.mat, mats)
	}
	return v
}

// worst tracks, for one quantity, the points out of tolerance and the one
// furthest out. Checks report err/tol, so 1 is the limit.
type worst struct {
	what      string
	count     int
	explained int
	peak      float64 // the largest err/tol seen, in tolerance or not
	ratio     float64
	at        [3]float32
	cpu, gpu  float32
	held      []heldPoint // mismatches that may be painted ties (addOrHold)
	tied      int         // held mismatches finish found to be ties
}

// add records one point. explain, when given, is asked only for a point out
// of tolerance: if fidget's own field takes the WGSL value within a hair of
// the point (see jumpExplains), the two sides sit on opposite sides of a
// discontinuity and the point is counted as explained, not failed.
func (w *worst) add(ratio float64, p [3]float32, cpu, gpu float32, explain func() bool) {
	if ratio <= 1 {
		w.peak = math.Max(w.peak, ratio)
		return
	}
	if explain != nil && explain() {
		w.explained++
		return
	}
	w.count++
	if math.IsNaN(ratio) || ratio > w.ratio {
		w.ratio, w.at, w.cpu, w.gpu = ratio, p, cpu, gpu
	}
}

// addOrHold is add for a colour or material, whose mismatch may yet be a tie
// between painted values: one that jumpExplains does not explain is held for
// finish to settle, rather than counted.
func (w *worst) addOrHold(ratio float64, p [3]float32, cpu, gpu float32, explain func() bool) {
	if ratio <= 1 {
		w.add(ratio, p, cpu, gpu, nil)
		return
	}
	if explain != nil && explain() {
		w.explained++
		return
	}
	w.held = append(w.held, heldPoint{ratio, p, cpu, gpu})
}

// heldPoint is a mismatch addOrHold has set aside.
type heldPoint struct {
	ratio    float64
	at       [3]float32
	cpu, gpu float32
}

func (w *worst) report(t *testing.T, total int) {
	if w.count == 0 {
		return
	}
	t.Errorf("%s: %d of %d points out of tolerance; worst at %v: fidget %.7g, WGSL %.7g (%.3g x tolerance)",
		w.what, w.count, total, w.at, w.cpu, w.gpu, w.ratio)
}

// maxChannel is the channel of one point furthest out of its tolerance, as
// err/tol, and its index.
func maxChannel(got, want []float32, tol func(float32) float64) (float64, int) {
	worstRatio, worstC := 0.0, 0
	for c := range got {
		r := math.Abs(float64(got[c]-want[c])) / tol(want[c])
		if math.IsNaN(r) {
			return r, c
		}
		if r > worstRatio {
			worstRatio, worstC = r, c
		}
	}
	return worstRatio, worstC
}

// jumpExplains reports whether fidget's canvas c takes a value within tol of
// gpu somewhere on a tiny cross around p. A field that jumps (nearest-unit
// color fill, a select, a floor) can be split by a float32 difference in the
// quantity that decides the jump; the two backends then disagree by the size
// of the jump, and both are right. A genuine disagreement does not go away
// when the point moves by jumpRadius.
func jumpExplains(t *testing.T, c fidget.CanvasIndex, p [3]float32, gpu float32, tol func(float32) float64, half float64) bool {
	e := float32(jumpRadius * half)
	for mult := float32(1.0); mult <= 8.0; mult *= 2.0 {
		var cross [][3]float32
		em := e * mult
		for a := 0; a < 3; a++ {
			for _, s := range []float32{-em, em} {
				q := p
				q[a] += s
				cross = append(cross, q)
			}
		}
		vals := evalCanvas(t, c, cross)
		for _, r := range vals {
			if math.Abs(float64(r.Value-gpu)) <= tol(r.Value) {
				return true
			}
		}
		for a := 0; a < 3; a++ {
			v0 := vals[2*a].Value
			v1 := vals[2*a+1].Value
			minV := float32(math.Min(float64(v0), float64(v1))) - float32(tol(gpu))
			maxV := float32(math.Max(float64(v0), float64(v1))) + float32(tol(gpu))
			if gpu >= minV && gpu <= maxV {
				return true
			}
		}
	}
	return false
}

// compare checks every point, and returns what it found for finish to report,
// once the colour and material mismatches it held have been settled.
func compare(t *testing.T, cpu *cpuSide, pts [][3]float32, want cpuValues, got []float32, n int, half float64, hasColor, hasMaterial bool) *fieldReport {
	st := stride(n, hasColor, hasMaterial)
	distTol := func(d float32) float64 {
		return distAbsTol*half + distRelTol*math.Abs(float64(d))
	}
	colorTolFn := func(float32) float64 { return colorTol }
	matTol := func(w float32) float64 { return materialTol * math.Max(1, math.Abs(float64(w))) }

	dist := make([]worst, n)
	colors := make([]worst, n)
	mats := make([]worst, n)
	colorFlag := make([]worst, n)
	matFlag := make([]worst, n)
	for k := 0; k < n; k++ {
		dist[k].what = fmt.Sprintf("shape %d distance", k+1)
		colors[k].what = fmt.Sprintf("shape %d color", k+1)
		mats[k].what = fmt.Sprintf("shape %d material", k+1)
		colorFlag[k].what = fmt.Sprintf("shape %d has-color flag", k+1)
		matFlag[k].what = fmt.Sprintf("shape %d material-present flag", k+1)
	}
	composite := worst{what: "composite shape index"}

	for i, p := range pts {
		out := got[i*st : (i+1)*st]

		// The composite picks the nearest crossing, |d|, over the shapes.
		// Where two shapes are within tolerance of each other the choice is
		// a tie either side may break.
		best, bestAbs, second := 0, math.Inf(1), math.Inf(1)
		for k := 0; k < n; k++ {
			a := math.Abs(float64(want.dist[k][i]))
			if a < bestAbs {
				second = bestAbs
				best, bestAbs = k, a
			} else if a < second {
				second = a
			}
		}
		gotIdx := int(out[1] + 0.5)
		if gotIdx != best+1 && second-bestAbs > 2*distTol(float32(bestAbs)) {
			composite.add(math.Inf(1), p, float32(best+1), out[1], nil)
		}

		for k := 0; k < n; k++ {
			wd, gd := want.dist[k][i], out[2+k]
			dist[k].add(math.Abs(float64(gd-wd))/distTol(wd), p, wd, gd, func() bool {
				return jumpExplains(t, cpu.dist[k], p, gd, distTol, half)
			})
		}

		off := 2 + n
		if hasColor {
			for k := 0; k < n; k++ {
				g := out[off+4*k : off+4*k+4]
				authored := g[3] > 0.5
				if authored != want.colorAuthored(k) {
					colorFlag[k].add(math.Inf(1), p, b2f(want.colorAuthored(k)), g[3], nil)
					continue
				}
				if !authored {
					continue
				}
				r, c := maxChannel(g[:3], want.color[k][i][:], colorTolFn)
				colors[k].addOrHold(r, p, want.color[k][i][c], g[c], func() bool {
					return jumpExplains(t, cpu.color[k][c], p, g[c], colorTolFn, half)
				})
			}
			off += 4 * n
		}
		if hasMaterial {
			for k := 0; k < n; k++ {
				g := out[off+17*k : off+17*k+17]
				present := g[16] > 0.5
				if present != want.materialAuthored(k) {
					matFlag[k].add(math.Inf(1), p, b2f(want.materialAuthored(k)), g[16], nil)
					continue
				}
				if !present {
					continue
				}
				r, c := maxChannel(g[:16], want.mat[k][i][:], matTol)
				mats[k].addOrHold(r, p, want.mat[k][i][c], g[c], func() bool {
					return jumpExplains(t, cpu.mat[k][c/4][c%4], p, g[c], matTol, half)
				})
			}
		}
	}

	return &fieldReport{pts: len(pts), n: n, hasColor: hasColor, hasMaterial: hasMaterial,
		composite: composite, dist: dist, colors: colors, mats: mats, colorFlag: colorFlag, matFlag: matFlag}
}

// fieldReport is what compare found at every point.
type fieldReport struct {
	pts, n                                 int
	hasColor, hasMaterial                  bool
	composite                              worst
	dist, colors, mats, colorFlag, matFlag []worst
}

// finish settles the held colour and material mismatches, asking ties which of
// their points sit on a tie between painted values (paintedTies), counts the
// rest as failures, and reports.
func (r *fieldReport) finish(t *testing.T, ties func(pts [][3]float32) []bool) {
	var held [][3]float32
	for _, ws := range [][]worst{r.colors, r.mats} {
		for k := range ws {
			for _, h := range ws[k].held {
				held = append(held, h.at)
			}
		}
	}
	var tie []bool
	if len(held) > 0 {
		tie = ties(held)
	}
	i := 0
	for _, ws := range [][]worst{r.colors, r.mats} {
		for k := range ws {
			w := &ws[k]
			for _, h := range w.held {
				if tie[i] {
					w.tied++
				} else {
					w.add(h.ratio, h.at, h.cpu, h.gpu, nil)
				}
				i++
			}
			w.held = nil
		}
	}

	explained, tied := 0, 0
	var peakDist, peakColor, peakMat float64
	r.composite.report(t, r.pts)
	for k := 0; k < r.n; k++ {
		for _, w := range []*worst{&r.dist[k], &r.colorFlag[k], &r.colors[k], &r.matFlag[k], &r.mats[k]} {
			w.report(t, r.pts)
			explained += w.explained
			tied += w.tied
		}
		peakDist = math.Max(peakDist, r.dist[k].peak)
		peakColor = math.Max(peakColor, r.colors[k].peak)
		peakMat = math.Max(peakMat, r.mats[k].peak)
	}
	t.Logf("%d points, %d shapes, color %v, material %v; closest to tolerance: distance %.2g, color %.2g, material %.2g; %d mismatches sit on a jump in the field, %d on a tie between painted values",
		r.pts, r.n, r.hasColor, r.hasMaterial, peakDist, peakColor, peakMat, explained, tied)
}

func b2f(b bool) float32 {
	if b {
		return 1
	}
	return 0
}
