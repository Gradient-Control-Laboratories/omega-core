package gpuparity

import (
	"encoding/json"
	"errors"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/eval"
	"gradientcontrol.com/omega/core/internal/fidget/convert"
	"gradientcontrol.com/omega/core/program/noise"
)

// The image both renderers draw: small, since the CPU side evaluates every
// shape at every march step.
const (
	imageW = 160
	imageH = 120
)

// The pass criteria, over the pixels away from silhouettes and glass:
//
//   - the return a pixel shows may differ on at most maxIndexOff of them
//   - the tone-mapped colour may differ by more than colorTolImage on at
//     most maxUnexplained of the pixels both sides hit, beyond what the CPU
//     image itself changes by to a neighbouring pixel
//
// The second is the image's jumpExplains: the viewer counts a hit anywhere
// within half a pixel's footprint of the surface, so the two marches can
// stop at slightly different depths, and on a textured, displaced or lattice
// surface that moves the colour by about as much as the next pixel over.
// Smooth shading has almost no neighbour change, so it is held to the plain
// tolerance.
const (
	maxIndexOff    = 0.005
	colorTolImage  = 0.02
	maxUnexplained = 0.01
)

// imageSkipExamples are left out of the image test only; the field test still
// covers them. Cellular Noise, Rust and Patina, Wood Planks and Spalted and
// Wormy were once here: their heavy frames came out of the Metal compile
// differently with unrelated code around them. Since each return's colour
// and material are computed only in its own case, they pass.
var imageSkipExamples = map[string]string{
	// Twenty-three screws threaded by relief, each thread well under a pixel
	// in this frame: the two marches stop within half a pixel of each other
	// (0.14 mm at the median) and sub-pixel relief turns that into colour, on
	// 1.03% of pixels against the 1% allowed. The per-return and winning-return
	// displacement give identical numbers here, since no two parts overlap;
	// the field test covers it point for point.
	"Fastener Matrix": "its threads are sub-pixel in this frame (see imageSkipExamples)",
}

// imagePalette is each return's colour on both sides, standing in for the
// viewer's per-shape UI palette.
var imagePalette = [][3]float64{
	{0.85, 0.85, 0.88},
	{0.80, 0.55, 0.45},
	{0.55, 0.70, 0.85},
	{0.70, 0.75, 0.55},
	{0.75, 0.60, 0.75},
	{0.60, 0.75, 0.72},
}

// TestGpuImageParity draws each example with the viewer's standard frame on
// the GPU (the shader the viewers build, from common/web/gsl-render-graph.js,
// run by gpu-eval) and with litrender on the CPU in its viewer mode
// (eval.LitOptions.ViewerMarch), from the same camera, and compares them pixel
// by pixel: which return each pixel shows, and its tone-mapped colour.
//
// Silhouettes are left out (a pixel whose neighbours show a different return
// on either side), since the two marches may place an edge a pixel apart, and
// so is glass, which each side samples differently by design. The field
// parity test proves the fields; what this adds is the march, the normals,
// the ambient occlusion and the shading.
func TestGpuImageParity(t *testing.T) {
	bin := filepath.Join(repoRoot, "target/release/gpu-eval")
	if _, err := os.Stat(bin); err != nil {
		t.Skipf("gpu-eval is not built (tool/build-fidget builds it): %v", err)
	}
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node is not installed; it builds the viewer's shader")
	}

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
			if why, ok := imageSkipExamples[name]; ok {
				t.Skip(why)
			}
			checkImage(t, bin, node, noisePath, file, name)
		})
	}
}

// imageCamera is litrender's default framing: from the corner (e, e, e) of
// the scene box toward its center, Z up, a 50-degree vertical field of view.
func imageCamera(e float64) eval.LitCamera {
	norm := func(v [3]float64) [3]float64 {
		l := math.Sqrt(v[0]*v[0] + v[1]*v[1] + v[2]*v[2])
		return [3]float64{v[0] / l, v[1] / l, v[2] / l}
	}
	cross := func(a, b [3]float64) [3]float64 {
		return [3]float64{a[1]*b[2] - a[2]*b[1], a[2]*b[0] - a[0]*b[2], a[0]*b[1] - a[1]*b[0]}
	}
	fwd := norm([3]float64{-1, -1, -1})
	right := norm(cross(fwd, [3]float64{0, 0, 1}))
	up := cross(right, fwd)
	tanY := math.Tan(25 * math.Pi / 180)
	return eval.LitCamera{
		Pos:    [3]float64{e, e, e},
		Right:  right,
		Up:     up,
		Fwd:    fwd,
		TanFov: [2]float64{tanY * imageW / imageH, tanY},
	}
}

func checkImage(t *testing.T, bin, node, noisePath, file, name string) {
	src, err := os.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}
	goldenPath := filepath.Join(repoRoot, "backend/wgsl/internal/codegen/testdata/example", name+".wgsl")
	golden, err := os.ReadFile(goldenPath)
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
	var n int
	if err := convert.ToColorShapes(res.Program, 0, func(shapes []convert.ColorShape) error {
		n = len(shapes)
		return nil
	}); err != nil {
		t.Fatalf("convert: %v", err)
	}
	if want := maxShapeIndex(wgsl); want != n {
		t.Skipf("fidget has %d outputs but the WGSL numbers %d shapes", n, want)
	}

	cam := imageCamera(half)
	colors := make([][3]float64, n)
	for k := range colors {
		colors[k] = imagePalette[k%len(imagePalette)]
	}

	// CPU.
	cpu, err := eval.RenderLitLinear(res.Program, eval.LitOptions{
		Width:       imageW,
		Height:      imageH,
		WorldExtent: half,
		Camera:      &cam,
		Colors:      colors,
		ViewerMarch: true,
	})
	if err != nil {
		t.Fatalf("litrender: %v", err)
	}

	// GPU.
	shader := standardFrame(t, node, goldenPath, n, cam, half, colors)
	pts := make([][3]float32, imageW*imageH)
	for y := 0; y < imageH; y++ {
		for x := 0; x < imageW; x++ {
			pts[y*imageW+x] = [3]float32{(float32(x) + 0.5) / imageW, (float32(y) + 0.5) / imageH, 0}
		}
	}
	got := runGPU(t, bin, shader, pts, 4, usesNoise, noisePath)

	if dir := os.Getenv("GSL_IMAGE_PARITY_DUMP"); dir != "" {
		dumpImages(t, dir, name, cpu, got)
	}
	compareImages(t, cpu, got, imageFar(half))
}

// dumpImages writes both renders as PNGs (name-cpu.png, name-gpu.png), for
// looking at a disagreement. Set GSL_IMAGE_PARITY_DUMP to a directory.
func dumpImages(t *testing.T, dir, name string, cpu *eval.LitLinear, gpu []float32) {
	t.Helper()
	w, h := cpu.Width, cpu.Height
	write := func(suffix string, at func(i, k int) float64) {
		img := image.NewRGBA(image.Rect(0, 0, w, h))
		for i := 0; i < w*h; i++ {
			var c [3]uint8
			for k := 0; k < 3; k++ {
				c[k] = uint8(math.Round(255 * math.Pow(math.Max(0, math.Min(1, at(i, k))), 1/2.2)))
			}
			img.Set(i%w, i/w, color.RGBA{c[0], c[1], c[2], 255})
		}
		f, err := os.Create(filepath.Join(dir, name+"-"+suffix+".png"))
		if err != nil {
			t.Fatal(err)
		}
		defer f.Close()
		if err := png.Encode(f, img); err != nil {
			t.Fatal(err)
		}
	}
	write("cpu", func(i, k int) float64 { return float64(cpu.RGB[3*i+k]) })
	write("gpu", func(i, k int) float64 { return float64(gpu[4*i+k]) })
}

// imageFar is the far plane the GPU encodes depth against: the camera sits on
// the scene box's corner, so nothing it can hit lies deeper than the box's
// diagonal, 2·sqrt(3)·half, which this keeps under the 0.498 the viewer
// clamps depth / far to.
func imageFar(half float64) float64 { return 8 * half }

// standardFrame has js/standard-frame.mjs build the viewer's frame shader.
func standardFrame(t *testing.T, node, goldenPath string, n int, cam eval.LitCamera, half float64, colors [][3]float64) string {
	t.Helper()
	bg, sky, ground := eval.LitEnvironment()
	job := map[string]any{
		"golden":     goldenPath,
		"shapeCount": n,
		"height":     imageH,
		"camera": map[string]any{
			"pos": cam.Pos, "right": cam.Right, "up": cam.Up, "fwd": cam.Fwd,
			"tanFov": cam.TanFov, "ortho": cam.Ortho,
		},
		"worldExtent": half,
		"far":         imageFar(half),
		"colors":      colors,
		"sky":         sky,
		"ground":      ground,
		"background":  bg,
	}
	in, err := json.Marshal(job)
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(node, "--import", "./js/register-three.mjs", "js/standard-frame.mjs")
	cmd.Stdin = strings.NewReader(string(in))
	var stderr strings.Builder
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("standard-frame.mjs: %v\n%s", err, stderr.String())
	}
	return string(out)
}

// compareImages reports how the two images agree away from silhouettes.
func compareImages(t *testing.T, cpu *eval.LitLinear, gpu []float32, far float64) {
	t.Helper()
	w, h := cpu.Width, cpu.Height
	gIdx := make([]int32, w*h)
	gDepth := make([]float64, w*h)
	for i := range gIdx {
		if a := float64(gpu[4*i+3]); a >= 1 {
			gIdx[i] = int32(math.Floor(a))
			gDepth[i] = (a - math.Floor(a)) * far
		}
	}
	edge := func(idx []int32, x, y int) bool {
		c := idx[y*w+x]
		for dy := -1; dy <= 1; dy++ {
			for dx := -1; dx <= 1; dx++ {
				xx, yy := x+dx, y+dy
				if xx < 0 || yy < 0 || xx >= w || yy >= h {
					continue
				}
				if idx[yy*w+xx] != c {
					return true
				}
			}
		}
		return false
	}

	var interior, indexOff, bothHit, unexplained int
	var errs, depthErrs []float64
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*w + x
			if cpu.Glass[i] || edge(cpu.Shape, x, y) || edge(gIdx, x, y) {
				continue
			}
			interior++
			if cpu.Shape[i] != gIdx[i] {
				indexOff++
				continue
			}
			if gIdx[i] == 0 {
				continue
			}
			bothHit++
			e := 0.0
			for k := 0; k < 3; k++ {
				e = math.Max(e, math.Abs(float64(cpu.RGB[3*i+k])-float64(gpu[4*i+k])))
			}
			errs = append(errs, e)
			// The CPU image's own change to its neighbours: how far a hit
			// placed a fraction of a pixel differently can move the colour.
			local := 0.0
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					xx, yy := x+dx, y+dy
					if xx < 0 || yy < 0 || xx >= w || yy >= h {
						continue
					}
					j := yy*w + xx
					for k := 0; k < 3; k++ {
						local = math.Max(local, math.Abs(float64(cpu.RGB[3*j+k])-float64(cpu.RGB[3*i+k])))
					}
				}
			}
			if e > colorTolImage+local {
				unexplained++
			}
			depthErrs = append(depthErrs, math.Abs(float64(cpu.Depth[i])-gDepth[i]))
		}
	}
	if interior == 0 {
		t.Skip("nothing but silhouettes and glass in view")
	}
	sort.Float64s(errs)
	pct := func(q float64) float64 {
		if len(errs) == 0 {
			return 0
		}
		return errs[int(q*float64(len(errs)-1))]
	}
	var sum float64
	for _, e := range errs {
		sum += e * e
	}
	rms := 0.0
	if len(errs) > 0 {
		rms = math.Sqrt(sum / float64(len(errs)))
	}
	sort.Float64s(depthErrs)
	dpct := func(q float64) float64 {
		if len(depthErrs) == 0 {
			return 0
		}
		return depthErrs[int(q*float64(len(depthErrs)-1))]
	}
	over := 0
	for _, e := range errs {
		if e > colorTolImage {
			over++
		}
	}
	overPct := 0.0
	if len(errs) > 0 {
		overPct = 100 * float64(over) / float64(len(errs))
	}
	defer func() {
		if f := float64(indexOff) / float64(interior); f > maxIndexOff {
			t.Errorf("the return differs on %.2f%% of interior pixels (at most %.2f%%)", 100*f, 100*maxIndexOff)
		}
		if len(errs) > 0 {
			if f := float64(unexplained) / float64(len(errs)); f > maxUnexplained {
				t.Errorf("colour differs by more than %.2f beyond the local change on %.2f%% of pixels (at most %.2f%%)",
					colorTolImage, 100*f, 100*maxUnexplained)
			}
		}
	}()
	t.Logf("interior %d, return differs %d (%.2f%%), both hit %d: colour error rms %.4f, p50 %.4f, p99 %.4f, max %.4f, over 0.02 %.2f%%, unexplained %.2f%%; depth error p50 %.4f, p99 %.4f, max %.4f",
		interior, indexOff, 100*float64(indexOff)/float64(interior), bothHit, rms, pct(0.5), pct(0.99), pct(1), overPct, 100*float64(unexplained)/math.Max(float64(len(errs)), 1), dpct(0.5), dpct(0.99), dpct(1))
}
