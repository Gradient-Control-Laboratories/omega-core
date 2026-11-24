//go:build cgo

package eval

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/Gradient-Control-Laboratories/omega-core/assembler/slider"
	"github.com/Gradient-Control-Laboratories/omega-core/backend/wgsl"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/fidget"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/meshimport"
	"github.com/Gradient-Control-Laboratories/omega-core/program/float16"
	"github.com/Gradient-Control-Laboratories/omega-core/program/noise"
)

// The languages Shader answers in.
const (
	ShaderWGSL      = "wgsl"
	ShaderGLSLES300 = "glsl-es300"
)

// ShaderOptions is what Shader is asked for.
type ShaderOptions struct {
	// Target is ShaderWGSL or ShaderGLSLES300.
	Target string
	// OutDir is where the textures the field reads are written, as raw texels. It is made if it is missing, and
	// only a document that reads a texture needs it.
	OutDir string
}

// ShaderUniform is a member of the uniform block that is not a slider: time, or a value the op knows, such as
// the box a mesh's distance grid covers. Offset is in bytes from the start of the block; Value is what to put
// there, one number per component.
type ShaderUniform struct {
	Name   string
	Type   string
	Offset int
	Value  []float64
}

// ShaderSlider is a slider of the document and the member of the uniform block that carries it. A host moves the
// slider by writing a float at Offset, without asking again.
type ShaderSlider struct {
	// Name is the slider's name in the document; Uniform is the block member's.
	Name    string
	Uniform string
	Type    string
	Offset  int
	Value   float64
	Min     float64
	Max     float64
}

// ShaderTexture is a texture the field reads, written to Path as raw texels: Size[0] by Size[1] by Size[2] of
// Format, x fastest, then y, then z.
type ShaderTexture struct {
	// Uniform is what the host binds: the texture's name in WGSL, the sampler's in GLSL.
	Uniform string
	// Binding is the texture's @binding in @group(0), for WGSL; and SamplerBinding its sampler's, when it has one.
	Binding        int
	SamplerBinding int
	// SamplerUniform is the sampler's name in WGSL, when the field samples the texture with one.
	SamplerUniform string
	// Kind is "noise3d" (the hash tile) or "meshSdf" (a mesh's distance grid).
	Kind string
	// Format is "rgba8unorm" or "r16float".
	Format string
	// Filter is how the field expects to read it: "nearest" or "linear", clamped to the edge.
	Filter string
	Size   [3]int
	Path   string
}

// ShaderBody is a return of the document; Index is the shapeIndex gsl_field and gsl_march report for it.
type ShaderBody struct {
	Name  string
	Index int
}

// ShaderResult is a document's field as a library a host's shader calls.
type ShaderResult struct {
	Target string
	// Code declares the uniform block and the textures and defines FieldFn and MarchFn. It has no entry point and,
	// in GLSL, no #version line and no outputs: a host adds its own.
	Code    string
	FieldFn string
	MarchFn string
	// UniformBlock is what the host binds its buffer to: the variable in WGSL (@group(0) @binding(0)), the uniform
	// block's name in GLSL. BlockSize is the buffer's size in bytes; Uniforms and Sliders say what goes where.
	UniformBlock string
	BlockSize    int
	Uniforms     []ShaderUniform
	Sliders      []ShaderSlider
	Textures     []ShaderTexture
	Bodies       []ShaderBody
	// Region is the box the march's step budget is sized for, and the one a host should march:
	// [x0, y0, z0, x1, y1, z1] in mm. It is the document's @scene box, or without one the box its field fills.
	Region [6]float64
}

// shaderStubFn is the entry point that keeps both functions through naga, and is cut from what naga writes.
const shaderStubFn = "gsl_shaderStub"

var shaderStubWGSL = fmt.Sprintf(`
@fragment
fn %s() -> @location(0) vec4<f32> {
    let s = %s(vec3<f32>(0.0));
    let h = %s(vec3<f32>(0.0), vec3<f32>(0.0, 0.0, 1.0), 0.0, 1.0, 1.0, 0.0, 1.0, vec4<f32>(0.0));
    return vec4<f32>(s.distance + h.t);
}
`, shaderStubFn, wgsl.FieldFn, wgsl.MarchFn)

// Shader returns the document's field, defs applied as the other queries apply them, as a library in the target
// language: one uniform block carrying time and the sliders, the textures the field reads, and two functions,
//
//	gsl_field(p)                                                          the distance and the body at p
//	gsl_march(origin, direction, tEntry, tExit, pixelScale, ortho, stepScale, detail)
//
// gsl_march is the front face of the viewer's primary march and returns whether the ray hit, how far along, which
// body, and the field's gradient there. Shading, depth and the entry point are the host's.
//
// Both targets are one WGSL module: ShaderWGSL returns it, ShaderGLSLES300 returns it transpiled for WebGL2.
func (e *Evaluator) Shader(ctx context.Context, manifest map[string]string, code, defs string, opt ShaderOptions) (*ShaderResult, error) {
	if opt.Target != ShaderWGSL && opt.Target != ShaderGLSLES300 {
		return nil, fmt.Errorf("shader: target must be %q or %q, got %q", ShaderWGSL, ShaderGLSLES300, opt.Target)
	}
	exprs := TopLevelReturnExprs(code)
	if len(exprs) == 0 {
		return nil, fmt.Errorf("shader: the document has no top-level return")
	}
	rendered, sliders, half, grids, err := e.shaderProgram(ctx, manifest, code, defs, exprs)
	if err != nil {
		return nil, err
	}

	result := &ShaderResult{Target: opt.Target, FieldFn: wgsl.FieldFn, MarchFn: wgsl.MarchFn}
	if half > 0 {
		result.Region = [6]float64{-half, -half, -half, half, half, half}
	} else {
		if result.Region, err = e.fitBox(ctx, manifest, code, defs, exprs); err != nil {
			return nil, fmt.Errorf("shader: %w", err)
		}
		for _, v := range result.Region {
			half = math.Max(half, math.Abs(v))
		}
	}
	host, err := wgsl.Host(rendered, half)
	if err != nil {
		return nil, err
	}
	for i, name := range rendered.ShapeNames {
		result.Bodies = append(result.Bodies, ShaderBody{Name: name, Index: i + 1})
	}

	// The block's members: a slider by the parameter the compiler lifted it to, anything else by its name.
	byParam := make(map[string]slider.SliderInfo, len(sliders))
	for _, s := range sliders {
		byParam[s.ParamName] = s
	}
	values, err := shaderGridUniforms(rendered, grids)
	if err != nil {
		return nil, err
	}
	result.BlockSize = host.BlockSize
	for _, u := range host.Uniforms {
		if s, ok := byParam[u.Name]; ok {
			result.Sliders = append(result.Sliders, ShaderSlider{
				Name: s.VariableName, Uniform: u.Name, Type: u.Type, Offset: u.Offset,
				Value: s.InitialValue, Min: s.Min, Max: s.Max,
			})
			continue
		}
		value, ok := values[u.Name]
		switch {
		case u.Name == "time":
			value = []float64{0}
		case !ok:
			return nil, fmt.Errorf("shader: the field takes %s (%s), which this op cannot give a value", u.Name, u.Type)
		}
		result.Uniforms = append(result.Uniforms, ShaderUniform{Name: u.Name, Type: u.Type, Offset: u.Offset, Value: value})
	}
	if result.Textures, err = shaderTextures(host, rendered, grids, opt.OutDir); err != nil {
		return nil, err
	}

	if opt.Target == ShaderWGSL {
		result.Code = host.WGSL
		result.UniformBlock = wgsl.HostBlockVar
		return result, nil
	}
	glsl, err := fidget.GLSLES300(host.WGSL+shaderStubWGSL, shaderStubFn)
	if err != nil {
		return nil, fmt.Errorf("shader: %w", err)
	}
	if result.Code, err = glslLibrary(glsl.Source); err != nil {
		return nil, err
	}
	block, ok := glsl.UniformBlocks[wgsl.HostBlockVar]
	if !ok {
		return nil, fmt.Errorf("shader: the transpiled module has no uniform block for %s", wgsl.HostBlockVar)
	}
	result.UniformBlock = block
	for i := range result.Textures {
		t := &result.Textures[i]
		sampler, ok := glsl.Textures[t.Uniform]
		if !ok {
			return nil, fmt.Errorf("shader: the transpiled module has no sampler for %s", t.Uniform)
		}
		// GLSL has one sampler where WGSL has a texture and a sampler, bound by name and not by number.
		t.Uniform, t.SamplerUniform, t.Binding, t.SamplerBinding = sampler, "", 0, 0
	}
	return result, nil
}

// shaderProgram assembles the document with defs applied and renders its WGSL, holding the query slot while the
// meshes it reads are registered and baked. half is half the @scene box's edge, or 0 without one.
func (e *Evaluator) shaderProgram(ctx context.Context, manifest map[string]string, code, defs string, exprs []string) (rendered *wgsl.Result, sliders []slider.SliderInfo, half float64, grids map[int]gridOf, err error) {
	if err := startQuery(ctx); err != nil {
		return nil, nil, 0, nil, err
	}
	defer endQuery()
	prog, source, uses, err := e.assembleQuery(ctx, manifest, code, defs, exprs)
	if err != nil {
		return nil, nil, 0, nil, err
	}
	if rendered, err = wgsl.Render(source, prog); err != nil {
		return nil, nil, 0, nil, fmt.Errorf("shader: %w", err)
	}
	if len(rendered.MeshSdfSlots) > 0 {
		if len(uses) == 0 {
			return nil, nil, 0, nil, fmt.Errorf("shader: the document reads a mesh's distance field, and this evaluator was not given the mesh files (LoadMeshesFrom)")
		}
		baked, err := meshimport.Grids(uses)
		if err != nil {
			return nil, nil, 0, nil, fmt.Errorf("shader: %w", err)
		}
		grids = make(map[int]gridOf, len(rendered.MeshSdfSlots))
		for _, slot := range rendered.MeshSdfSlots {
			g := baked[slot.AssetIndex]
			if g == nil {
				return nil, nil, 0, nil, fmt.Errorf("shader: no distance grid for mesh %d", slot.AssetIndex)
			}
			size := g.Bbox.Size()
			grids[slot.Slot] = gridOf{dims: g.Dims, lo: g.Bbox.Min, size: size, values: g.Values}
		}
	}
	if scene := prog.Primary.Scene; scene != nil {
		half = scene.GetSize() / 2
	}
	return rendered, slider.Extract(source, prog.Primary), half, grids, nil
}

// gridOf is a mesh's distance grid as the shader's slot reads it.
type gridOf struct {
	dims   [3]int
	lo     [3]float64
	size   [3]float64
	values []float32
}

// The arguments sdShape takes for mesh slot N (backend/wgsl's codegen names them).
var (
	shaderGridTex  = regexp.MustCompile(`^sdfTex(\d+)$`)
	shaderGridSmp  = regexp.MustCompile(`^sdfSmp(\d+)$`)
	shaderGridLo   = "sdfBMin"
	shaderGridSize = "sdfBSize"
)

// shaderGridUniforms is the value of each uniform that says where a mesh's distance grid sits.
func shaderGridUniforms(rendered *wgsl.Result, grids map[int]gridOf) (map[string][]float64, error) {
	values := make(map[string][]float64, 2*len(grids))
	for slot, g := range grids {
		values[shaderGridLo+strconv.Itoa(slot)] = g.lo[:]
		values[shaderGridSize+strconv.Itoa(slot)] = g.size[:]
	}
	return values, nil
}

// shaderTextures writes each texture the field reads under outDir and says what it is. A texture this op has no
// data for is an error, not a binding left empty.
func shaderTextures(host *wgsl.HostModule, rendered *wgsl.Result, grids map[int]gridOf, outDir string) ([]ShaderTexture, error) {
	samplers := make(map[string]wgsl.HostTexture)
	for _, t := range host.Textures {
		if m := shaderGridSmp.FindStringSubmatch(t.Name); m != nil && t.Type == "sampler" {
			samplers[m[1]] = t
		}
	}
	var out []ShaderTexture
	for _, t := range host.Textures {
		if t.Type == "sampler" {
			if shaderGridSmp.MatchString(t.Name) {
				continue
			}
			return nil, fmt.Errorf("shader: the field takes the sampler %s, which this op does not provide yet", t.Name)
		}
		tex := ShaderTexture{Uniform: t.Name, Binding: t.Binding}
		var data []byte
		switch m := shaderGridTex.FindStringSubmatch(t.Name); {
		case t.Name == "noiseTex" && t.Type == "texture_3d<f32>":
			tex.Kind, tex.Format, tex.Filter = "noise3d", "rgba8unorm", "nearest"
			tex.Size = [3]int{noise.TileSize, noise.TileSize, noise.TileSize}
			data = noise.Bake()
		case m != nil && t.Type == "texture_3d<f32>":
			slot, _ := strconv.Atoi(m[1])
			g, ok := grids[slot]
			if !ok {
				return nil, fmt.Errorf("shader: no distance grid for %s", t.Name)
			}
			tex.Kind, tex.Format, tex.Filter, tex.Size = "meshSdf", "r16float", "linear", g.dims
			data = make([]byte, 0, 2*len(g.values))
			for _, v := range g.values {
				data = binary.LittleEndian.AppendUint16(data, float16.FromFloat32(v))
			}
			if s, ok := samplers[m[1]]; ok {
				tex.SamplerUniform, tex.SamplerBinding = s.Name, s.Binding
			}
		default:
			return nil, fmt.Errorf("shader: the field reads the texture %s (%s), which this op does not write yet", t.Name, t.Type)
		}
		if outDir == "" {
			return nil, fmt.Errorf("shader: the field reads the texture %s, so the request needs an outDir to write it to", t.Name)
		}
		if err := os.MkdirAll(outDir, 0o755); err != nil {
			return nil, fmt.Errorf("shader: %w", err)
		}
		tex.Path = filepath.Join(outDir, t.Name+".bin")
		if err := os.WriteFile(tex.Path, data, 0o644); err != nil {
			return nil, fmt.Errorf("shader: %w", err)
		}
		out = append(out, tex)
	}
	return out, nil
}

var (
	glslVersionLine = regexp.MustCompile(`(?m)^#version [^\n]*\n`)
	glslStubOutput  = regexp.MustCompile(`(?m)^layout\(location = 0\) out vec4 \w+;\n`)
)

// glslLibrary cuts what makes naga's output a program from it: the #version line, the stub entry point's output
// and main(), which naga writes last. What is left declares and defines, and a host adds its own program around it.
func glslLibrary(src string) (string, error) {
	main := strings.LastIndex(src, "\nvoid main() {")
	if main < 0 {
		return "", fmt.Errorf("shader: the transpiled module has no main() to cut")
	}
	src = src[:main+1]
	if !glslVersionLine.MatchString(src) {
		return "", fmt.Errorf("shader: the transpiled module has no #version line to cut")
	}
	src = glslVersionLine.ReplaceAllString(src, "")
	src = glslStubOutput.ReplaceAllString(src, "")
	for _, fn := range []string{wgsl.FieldFn + "(", wgsl.MarchFn + "("} {
		if !strings.Contains(src, fn) {
			return "", fmt.Errorf("shader: the transpiled module lost %s", strings.TrimSuffix(fn, "("))
		}
	}
	return src, nil
}
