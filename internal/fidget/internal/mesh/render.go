package mesh

import (
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sort"
	"strings"

	"gradientcontrol.com/omega/core/internal/fidget/internal/fidget"
	"gradientcontrol.com/omega/core/program"
)

type glslCompilationResult struct {
	Inputs  map[string]float64 `json:"inputs"`
	Helpers []string           `json:"helpers"`
	Glsl    []string           `json:"glsl"`
}

type GlslResult struct {
	Inputs  map[string]*fidget.Input
	GlslSrc string
}

func GenerateRender(prog *program.Program) ([]byte, *GlslResult, error) {
	return generateRender(prog, false)
}

// GenerateRenderFramed renders like GenerateRender but frames the camera on
// the geometry: a two-pass grid scan of the union field finds the box where
// it is negative, and the view centers there with its half-height set from
// the box's bounding radius. When the scan finds nothing negative inside
// ±frameScanHalfWidth (an empty scene, or a field the eval rejects), it
// falls back to GenerateRender's fixed camera.
func GenerateRenderFramed(prog *program.Program) ([]byte, *GlslResult, error) {
	return generateRender(prog, true)
}

func generateRender(prog *program.Program, framed bool) ([]byte, *GlslResult, error) {
	fidget.Lock()
	defer fidget.Unlock()
	defer fidget.Clear()

	fidget.Clear()

	fgScope := fidget.NewScope(nil)

	for _, dep := range prog.Dependencies {
		err := fgScope.EvaluatePackage(dep)
		if err != nil {
			return nil, nil, err
		}
	}
	err := fgScope.EvaluatePackage(prog.Primary)
	if err != nil {
		return nil, nil, err
	}

	treeIndices, f, err := fgScope.EvaluateReturns(prog.Primary.Files[len(prog.Primary.Files)-1].Returns)
	if err != nil {
		return nil, nil, err
	}
	if f != nil {
		return nil, nil, fmt.Errorf("unexpected return: %v", f)
	}
	if len(treeIndices) == 0 {
		return nil, nil, fmt.Errorf("expected at least 1 return, but got 0")
	}
	treeIndex := treeIndices[0]
	for i := 1; i < len(treeIndices); i++ {
		treeIndex = fidget.Min(treeIndex, treeIndices[i])
	}

	canvasIndex := fidget.BuildCanvas(treeIndex)

	// GLSL codegen
	encodedJson := fidget.CodegenGLSL(canvasIndex)
	log.Printf(" --- json ---\n%s", encodedJson)

	result := glslCompilationResult{}
	err = json.NewDecoder(strings.NewReader(encodedJson)).Decode(&result)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decode json: %v", err)
	}
	log.Printf(" --- decoded ---\n%v", result)

	glslSrc := ""

	// Build a map of the referenced inputs
	inputs := make(map[string]*fidget.Input, 10)
	inputNames := []string{}
	for name := range result.Inputs {
		treeIndices, ok := fgScope.NodeNamed(name)
		if !ok {
			log.Printf("Failed to find tree index for input: %s", name)
			continue
		}
		if len(treeIndices) != 1 {
			log.Printf("Found %d tree indices for input: %s", len(treeIndices), name)
			continue
		}
		treeIndex := treeIndices[0]
		input := fgScope.InputFor(treeIndex)
		if input == nil {
			log.Printf("Failed to find input: %s", name)
			continue
		}
		inputs[name] = input
		inputNames = append(inputNames, name)
	}
	sort.Strings(inputNames)
	for _, name := range inputNames {
		glslSrc += fmt.Sprintf("uniform float %s;\n", name)
	}

	helpers := strings.Join(result.Helpers, "\n")
	if len(helpers) > 0 {
		glslSrc += "\n// Fidget helpers\n"
		glslSrc += helpers + "\n"
	}

	glslSrc += "\nColorImplicit buildShape(vec3 iP) {\n  "
	glslSrc += strings.Join(result.Glsl, "\n  ")
	glslSrc += "\n}\n"

	log.Printf(" --- glsl ---\n%s", glslSrc)

	// render PNG
	centerX, centerY, centerZ, scale := float32(0), float32(0), float32(0), float32(100)
	if framed {
		if center, s, ok := frameCanvas(canvasIndex); ok {
			centerX, centerY, centerZ, scale = center[0], center[1], center[2], s
		}
	}
	pngBytes, err := fidget.RenderPng(canvasIndex, centerX, centerY, centerZ, scale, 3.6, 1.2)
	if err != nil {
		log.Printf("failed to render png: %v", err)
		// fall through to return GLSL without PNG
	}

	return pngBytes, &GlslResult{
		Inputs:  inputs,
		GlslSrc: glslSrc,
	}, nil
}

// frameScanRes and frameScanHalfWidth control the camera-framing scan: two
// passes of a coarse grid over the union field, a cell counting as occupied
// where its center's value is negative.
const (
	frameScanRes       = 24
	frameScanHalfWidth = float32(1024)
)

// frameCanvas fits a view to the canvas's negative region: the occupied
// box's center, with the view half-height (RenderPng's scale) at 1.25x the
// box's bounding radius — the radius covers every projected extent under
// any yaw/pitch, and the view is wider than tall. ok is false when nothing
// negative lies within the scan region or the eval fails.
func frameCanvas(canvasIndex fidget.CanvasIndex) (center [3]float32, scale float32, ok bool) {
	coarse := [2][3]float32{
		{-frameScanHalfWidth, -frameScanHalfWidth, -frameScanHalfWidth},
		{frameScanHalfWidth, frameScanHalfWidth, frameScanHalfWidth},
	}
	box, found := scanNegativeBox(canvasIndex, coarse)
	if !found {
		return center, 0, false
	}
	// Refine: rescan the occupied box (inflated by one of its cells) at the
	// same resolution until the box stops shrinking. A shape much smaller
	// than the initial cell needs several passes to converge — each pass
	// shrinks the cell by roughly the occupied fraction of the region.
	diag := func(b [2][3]float32) float32 {
		var d2 float32
		for k := 0; k < 3; k++ {
			d := b[1][k] - b[0][k]
			d2 += d * d
		}
		return float32(math.Sqrt(float64(d2)))
	}
	for pass := 0; pass < 4; pass++ {
		inflated := box
		for k := 0; k < 3; k++ {
			cell := (box[1][k] - box[0][k]) / frameScanRes
			inflated[0][k] -= cell
			inflated[1][k] += cell
		}
		refined, foundRefined := scanNegativeBox(canvasIndex, inflated)
		if !foundRefined {
			break
		}
		shrunk := diag(refined) < 0.75*diag(box)
		box = refined
		if !shrunk {
			break
		}
	}
	var radius2 float32
	for k := 0; k < 3; k++ {
		center[k] = 0.5 * (box[0][k] + box[1][k])
		half := 0.5 * (box[1][k] - box[0][k])
		radius2 += half * half
	}
	scale = 1.25 * float32(math.Sqrt(float64(radius2)))
	if !(scale > 0) {
		return center, 0, false
	}
	return center, scale, true
}

// scanNegativeBox samples the canvas at frameScanRes^3 cell centers over the
// region and returns the union box of occupied cells, inflated by a half
// cell per side. A cell is occupied when its value is negative; only when NO
// cell is negative does the scan fall back to the cell half-diagonal
// threshold, which catches a shape smaller than one coarse cell — and that
// fallback is discarded as uninformative when it marks most of the grid (a
// clamped field's constant background sits below the half-diagonal
// everywhere). The caller holds the fidget lock.
func scanNegativeBox(canvasIndex fidget.CanvasIndex, region [2][3]float32) (box [2][3]float32, found bool) {
	var cell [3]float32
	for k := 0; k < 3; k++ {
		cell[k] = (region[1][k] - region[0][k]) / frameScanRes
	}
	halfDiag := 0.5 * float32(math.Sqrt(float64(cell[0]*cell[0]+cell[1]*cell[1]+cell[2]*cell[2])))
	n := frameScanRes * frameScanRes * frameScanRes
	xs := make([]float32, 0, n)
	ys := make([]float32, 0, n)
	zs := make([]float32, 0, n)
	for iz := 0; iz < frameScanRes; iz++ {
		z := region[0][2] + (float32(iz)+0.5)*cell[2]
		for iy := 0; iy < frameScanRes; iy++ {
			y := region[0][1] + (float32(iy)+0.5)*cell[1]
			for ix := 0; ix < frameScanRes; ix++ {
				xs = append(xs, region[0][0]+(float32(ix)+0.5)*cell[0])
				ys = append(ys, y)
				zs = append(zs, z)
			}
		}
	}
	samples, err := fidget.EvalGrad(canvasIndex, xs, ys, zs)
	if err != nil {
		log.Printf("frame scan eval: %v", err)
		return box, false
	}

	inf := float32(math.Inf(1))
	newBox := func() [2][3]float32 {
		return [2][3]float32{{inf, inf, inf}, {-inf, -inf, -inf}}
	}
	grow := func(b *[2][3]float32, c [3]float32) {
		for k := 0; k < 3; k++ {
			if c[k]-0.5*cell[k] < b[0][k] {
				b[0][k] = c[k] - 0.5*cell[k]
			}
			if c[k]+0.5*cell[k] > b[1][k] {
				b[1][k] = c[k] + 0.5*cell[k]
			}
		}
	}
	negative, nearBand := newBox(), newBox()
	negFound, nearCount := false, 0
	for i, s := range samples {
		c := [3]float32{xs[i], ys[i], zs[i]}
		if s.Value < 0 {
			negFound = true
			grow(&negative, c)
		} else if s.Value < halfDiag {
			nearCount++
			grow(&nearBand, c)
		}
	}
	switch {
	case negFound:
		return negative, true
	case nearCount > 0 && nearCount < n*4/5:
		return nearBand, true
	}
	return box, false
}
