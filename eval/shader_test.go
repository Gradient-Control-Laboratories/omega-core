//go:build cgo

package eval_test

import (
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/eval"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/fidget"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/source"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/require"
)

// shaderStub is a fragment entry point that calls both functions, so a WGSL
// response can be put through naga whole.
const shaderStub = `
@fragment
fn fs_main() -> @location(0) vec4<f32> {
    let s = gsl_field(vec3<f32>(0.0));
    let h = gsl_march(vec3<f32>(0.0), vec3<f32>(0.0, 0.0, 1.0), 0.0, 1.0, 1.0, 0.0, 1.0, vec4<f32>(0.0));
    return vec4<f32>(s.distance + h.t);
}
`

// Both targets answer with a library: the two functions and no entry point.
func TestShader_BothTargets(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	for _, target := range []string{"wgsl", "glsl-es300"} {
		t.Run(target, func(t *testing.T) {
			res, err := e.Shader(t.Context(), nil, meshBallDoc, "", eval.ShaderOptions{Target: target, OutDir: t.TempDir()})
			require.NoError(t, err)
			require.Equal(t, target, res.Target)
			require.Equal(t, "gsl_field", res.FieldFn)
			require.Equal(t, "gsl_march", res.MarchFn)
			require.Contains(t, res.Code, "gsl_field(")
			require.Contains(t, res.Code, "gsl_march(")
			for _, gone := range []string{"@fragment", "void main", "#version", "gsl_shaderStub"} {
				require.False(t, strings.Contains(res.Code, gone), gone)
			}
			require.Contains(t, res.Code, res.UniformBlock)
			require.Equal(t, [6]float64{-20, -20, -20, 20, 20, 20}, res.Region)
			require.Equal(t, []eval.ShaderBody{{Name: "ball", Index: 1}}, res.Bodies)
			require.Empty(t, res.Textures)

			require.Len(t, res.Sliders, 1)
			s := res.Sliders[0]
			require.Equal(t, "r", s.Name)
			require.Equal(t, "f32", s.Type)
			require.Equal(t, 5.0, s.Value)
			require.Equal(t, 2.0, s.Min)
			require.Equal(t, 12.0, s.Max)
			require.Equal(t, 4, s.Offset)
			require.Equal(t, 16, res.BlockSize)
			require.Len(t, res.Uniforms, 1)
			require.Equal(t, eval.ShaderUniform{Name: "time", Type: "f32", Offset: 0, Value: []float64{0}}, res.Uniforms[0])
			if target == "wgsl" {
				require.Equal(t, "gsl_host", res.UniformBlock)
				require.Contains(t, res.Code, "    "+s.Uniform+": f32,")
			}
		})
	}
}

// defs apply as they do to every query. One that replaces a slider declared with @range replaces the annotation
// too (substituteTopLevelDecls), so the value is built into the field and there is no slider to move.
func TestShader_DefsAndSliders(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	res, err := e.Shader(t.Context(), nil, meshBallDoc, "r = 7.0;", eval.ShaderOptions{Target: "wgsl"})
	require.NoError(t, err)
	require.Empty(t, res.Sliders)
	require.Contains(t, res.Code, "7.0")

	// Without the def the same document keeps its slider.
	res, err = e.Shader(t.Context(), nil, meshBallDoc, "", eval.ShaderOptions{Target: "wgsl"})
	require.NoError(t, err)
	require.Len(t, res.Sliders, 1)
	require.Equal(t, "r", res.Sliders[0].Name)
	require.Equal(t, 5.0, res.Sliders[0].Value)
	require.Equal(t, 2.0, res.Sliders[0].Min)
	require.Equal(t, 12.0, res.Sliders[0].Max)
}

// The WGSL response is a module naga parses and validates once an entry point
// is added, and the GLSL one is that module's transpile.
func TestShader_WGSLValidates(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	res, err := e.Shader(t.Context(), nil, meshBallDoc, "", eval.ShaderOptions{Target: "wgsl", OutDir: t.TempDir()})
	require.NoError(t, err)
	glsl, err := fidget.GLSLES300(res.Code+shaderStub, "fs_main")
	require.NoError(t, err)
	require.Contains(t, glsl.Source, "gsl_march(")
}

// A document with no @scene is marched over the box its field fills.
func TestShader_RegionWithoutAScene(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	res, err := e.Shader(t.Context(), nil, "return length(p) - 5.0;", "", eval.ShaderOptions{Target: "wgsl", OutDir: t.TempDir()})
	require.NoError(t, err)
	for k := 0; k < 3; k++ {
		require.True(t, res.Region[k] <= -5 && res.Region[k] > -8, "%v", res.Region)
		require.True(t, res.Region[k+3] >= 5 && res.Region[k+3] < 8, "%v", res.Region)
	}
}

// A document that reads the noise tile lists it and gets it as a file of raw
// texels.
func TestShader_NoiseTexture(t *testing.T) {
	dir := t.TempDir()
	e := eval.New(source.NewLocalManager("../stdlib"), nil)
	doc := "@scene(100)\nimport \"noise\";\nreturn max(noise.cellular(p / 10.0).x - 0.3, length(p) - 30.0);\n"
	for _, target := range []string{"wgsl", "glsl-es300"} {
		t.Run(target, func(t *testing.T) {
			res, err := e.Shader(t.Context(), nil, doc, "", eval.ShaderOptions{Target: target, OutDir: dir})
			require.NoError(t, err)
			require.Len(t, res.Textures, 1)
			tex := res.Textures[0]
			require.Equal(t, "noise3d", tex.Kind)
			require.Equal(t, "rgba8unorm", tex.Format)
			require.Equal(t, [3]int{64, 64, 64}, tex.Size)
			require.Equal(t, "nearest", tex.Filter)
			require.Contains(t, res.Code, tex.Uniform)
			require.Equal(t, dir, filepath.Dir(tex.Path))
			info, err := os.Stat(tex.Path)
			require.NoError(t, err)
			require.Equal(t, int64(64*64*64*4), info.Size())
			if target == "wgsl" {
				require.Equal(t, "noiseTex", tex.Uniform)
				require.Equal(t, 1, tex.Binding)
			}
		})
	}
}

// stlTetrahedron is a binary STL of a tetrahedron reaching 10 mm along each axis.
func stlTetrahedron() []byte {
	v := [4][3]float32{{0, 0, 0}, {10, 0, 0}, {0, 10, 0}, {0, 0, 10}}
	faces := [4][3]int{{0, 2, 1}, {0, 1, 3}, {0, 3, 2}, {1, 2, 3}}
	out := make([]byte, 80)
	out = binary.LittleEndian.AppendUint32(out, uint32(len(faces)))
	for _, f := range faces {
		out = append(out, make([]byte, 12)...)
		for _, i := range f {
			for _, c := range v[i] {
				out = binary.LittleEndian.AppendUint32(out, math.Float32bits(c))
			}
		}
		out = append(out, 0, 0)
	}
	return out
}

// A document that reads a mesh's distance field gets the grid as a file, and
// the box it covers as uniforms with their values.
func TestShader_MeshSdfTexture(t *testing.T) {
	meshes, out := t.TempDir(), t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(meshes, "tet.stl"), stlTetrahedron(), 0o644))
	e := eval.New(&source.MockManager{}, nil)
	e.LoadMeshesFrom(meshes)
	doc := "@scene(40)\nimportMesh 0 \"tet.stl\";\nreturn bvhSdf(0);\n"
	res, err := e.Shader(t.Context(), nil, doc, "", eval.ShaderOptions{Target: "glsl-es300", OutDir: out})
	require.NoError(t, err)
	require.Len(t, res.Textures, 1)
	tex := res.Textures[0]
	require.Equal(t, "meshSdf", tex.Kind)
	require.Equal(t, "r16float", tex.Format)
	require.Equal(t, "linear", tex.Filter)
	require.Contains(t, res.Code, "sampler3D "+tex.Uniform)
	info, err := os.Stat(tex.Path)
	require.NoError(t, err)
	require.Equal(t, int64(tex.Size[0]*tex.Size[1]*tex.Size[2]*2), info.Size())
	require.True(t, info.Size() > 0)

	// time, then the grid's corner and size.
	require.Len(t, res.Uniforms, 3)
	lo, size := res.Uniforms[1], res.Uniforms[2]
	require.Equal(t, "vec3<f32>", lo.Type)
	require.Len(t, lo.Value, 3)
	require.Len(t, size.Value, 3)
	for k := 0; k < 3; k++ {
		require.True(t, lo.Value[k] <= 0 && lo.Value[k]+size.Value[k] >= 10, "%v %v", lo.Value, size.Value)
	}
}

func TestShader_Refusals(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	for name, tc := range map[string]struct {
		doc  string
		opt  eval.ShaderOptions
		want string
	}{
		"no target":      {meshBallDoc, eval.ShaderOptions{OutDir: "x"}, "target"},
		"unknown target": {meshBallDoc, eval.ShaderOptions{Target: "hlsl", OutDir: "x"}, "target"},
		"no returns":     {"a = 1.0;", eval.ShaderOptions{Target: "wgsl", OutDir: "x"}, "return"},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := e.Shader(t.Context(), nil, tc.doc, "", tc.opt)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.want)
		})
	}
}

// A large document with parts and materials answers for both targets.
func TestShader_CorexyPrinter(t *testing.T) {
	doc, err := os.ReadFile("../examples/corexy_printer.gsl")
	require.NoError(t, err)
	e := eval.New(source.NewLocalManager("../stdlib"), nil)
	for _, target := range []string{"wgsl", "glsl-es300"} {
		t.Run(target, func(t *testing.T) {
			res, err := e.Shader(t.Context(), nil, string(doc), "", eval.ShaderOptions{Target: target, OutDir: t.TempDir()})
			require.NoError(t, err)
			require.Contains(t, res.Code, "gsl_vanishEye")
			require.True(t, len(res.Bodies) > 1)
			require.True(t, len(res.Sliders) > 0)
			if target == "wgsl" {
				_, err := fidget.GLSLES300(res.Code+shaderStub, "fs_main")
				require.NoError(t, err)
			}
		})
	}
}
