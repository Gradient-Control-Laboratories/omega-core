package wgsl

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/require"
)

func hostOf(t *testing.T, src string) *HostModule {
	t.Helper()
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(&source.MockManager{}))
	require.NoError(t, err)
	rendered, err := Render(src, res.Program)
	require.NoError(t, err)
	host, err := Host(rendered, 8)
	require.NoError(t, err)
	return host
}

// The step rule is the viewer's, and the emitted march takes its numbers from
// the table: a rule with other numbers writes them and not the viewer's.
func TestMarchTakesItsNumbersFromTheRule(t *testing.T) {
	require.Equal(t, MarchRule{
		Fudge: 0.8, JumpSlack: 4, MinSteps: 256, MaxSteps: 1024, PathScaleMax: 1.75,
		HitEps: 0.001, HitPixels: 0.5, GradEps: 0.0001, GradPixels: 0.025, GradUlps: 4e-7,
		VanishHitEps: 8, VanishPixels: 4,
	}, FrontMarch)

	text := marchWGSL(FrontMarch, 8, false, "")
	for _, want := range []string{
		"gradMag * 0.8;",
		"<= 4.0 * prevGradMag * (t - prevT)",
		"clamp(256.0 * pathScale, 256.0, 1024.0) * stepScale",
		"i < 1024u",
		"/ 16.0, 1.0, 1.75)",
		"max(0.001, pixelWorld * 0.5)",
		"max(max(0.0001, pixelWorld * 0.025), length(p) * 4e-07)",
	} {
		require.Contains(t, text, want)
	}

	other := FrontMarch
	other.Fudge, other.JumpSlack, other.MinSteps, other.MaxSteps = 0.7, 3, 100, 200
	text = marchWGSL(other, 8, false, "")
	for _, want := range []string{"gradMag * 0.7;", "<= 3.0 * prevGradMag", "clamp(100.0 * pathScale, 100.0, 200.0)", "i < 200u"} {
		require.Contains(t, text, want)
	}
	for _, gone := range []string{"* 0.8;", "256.0", "1024"} {
		require.False(t, strings.Contains(text, gone), gone)
	}
}

// A program with parts has the march hand the part guards the ray and take it
// back; one without never mentions them.
func TestMarchSetsTheVanishGlobalsOnlyForParts(t *testing.T) {
	plain := marchWGSL(FrontMarch, 8, false, "")
	require.False(t, strings.Contains(plain, "gsl_vanish"), "gsl_vanish")
	require.False(t, strings.Contains(plain, "gsl_simpleRung"), "gsl_simpleRung")

	parts := marchWGSL(FrontMarch, 8, true, "var<private> gsl_vanishEye: vec4<f32>;")
	require.Contains(t, parts, "gsl_vanishEye = vec4<f32>(origin, pixelScale);")
	require.Contains(t, parts, "gsl_vanishView = vec4<f32>(ortho, detail.w, 4.0 * detail.w,")
	require.Contains(t, parts, "max(0.001 * 8.0, pixelScale * mix(max(tExit, 1.0), 1.0, ortho) * 4.0)")
	require.Contains(t, parts, "gsl_vanishView.w = 0.0;")
	require.False(t, strings.Contains(parts, "gsl_simpleRung"), "gsl_simpleRung")
	// The guards are off again before the gradient at the hit is taken.
	require.Less(t, strings.Index(parts, "gsl_vanishView.w = 0.0;"), strings.Index(parts, "result = gsl_hit(true"))

	standIns := marchWGSL(FrontMarch, 8, true, "var<private> gsl_simpleRung: f32;\nvar<private> gsl_detailPw: f32 = 0.0;")
	require.Contains(t, standIns, "gsl_simpleRung = detail.z;")
	require.Contains(t, standIns, "gsl_detailPw = gsl_pw;")
}

// The module declares what sdShape expects, the sliders as one uniform block
// laid out as WGSL lays it out, and the two functions, with no entry point.
func TestHost(t *testing.T) {
	host := hostOf(t, `r = 5;
r.min = 1;
r.max = 10;
h = 2;
h.min = 1;
h.max = 4;
return max(length(p) - r, abs(p.z) - h);`)

	require.False(t, host.Parts)
	require.Empty(t, host.Textures)
	require.Len(t, host.Uniforms, 3)
	require.Equal(t, HostUniform{Name: "time", Type: "f32", Offset: 0, Size: 4}, host.Uniforms[0])
	require.Equal(t, 4, host.Uniforms[1].Offset)
	require.Equal(t, 8, host.Uniforms[2].Offset)
	require.Equal(t, 16, host.BlockSize)
	for _, u := range host.Uniforms[1:] {
		require.True(t, strings.HasPrefix(u.Name, "gsl_param_"), u.Name)
		require.Contains(t, host.WGSL, "    "+u.Name+": f32,\n")
	}
	for _, want := range []string{
		"struct gsl_shape {", "struct gsl_material {", "fn gsl_min(", "fn gsl_isShapeVisible(",
		"var<private> gsl_fieldFilterIdx: u32", "var<private> gsl_pw: f32", "var<private> gsl_pwScale: f32",
		"@group(0) @binding(0) var<uniform> gsl_host: gsl_Host;",
		"fn gsl_field(p: vec3<f32>) -> gsl_shape {\n    return sdShape(p, gsl_host.time, gsl_host." + host.Uniforms[1].Name + ", gsl_host." + host.Uniforms[2].Name + ");",
		"fn gsl_march(origin: vec3<f32>, direction: vec3<f32>, tEntry: f32, tExit: f32,",
		"fn gsl_gradient(",
	} {
		require.Contains(t, host.WGSL, want)
	}
	for _, gone := range []string{"@fragment", "@vertex", "@compute", "gsl_vanish"} {
		require.False(t, strings.Contains(host.WGSL, gone), gone)
	}
}

// A vector argument is aligned to 16 bytes, and a number after it takes the
// fourth slot; textures and samplers become bindings, in order.
func TestHost_LayoutAndTextures(t *testing.T) {
	res := &Result{WGSL: "fn sdShape(p: vec3<f32>, time: f32, a: f32, lo: vec3<f32>, b: f32, noiseTex: texture_3d<f32>, sdfSmp0: sampler, uv: vec2<f32>) -> gsl_shape {\n  return gsl_shape(a + b, 1u);\n}"}
	host, err := Host(res, 1)
	require.NoError(t, err)
	require.Equal(t, []HostUniform{
		{Name: "time", Type: "f32", Offset: 0, Size: 4},
		{Name: "a", Type: "f32", Offset: 4, Size: 4},
		{Name: "lo", Type: "vec3<f32>", Offset: 16, Size: 12},
		{Name: "b", Type: "f32", Offset: 28, Size: 4},
		{Name: "uv", Type: "vec2<f32>", Offset: 32, Size: 8},
	}, host.Uniforms)
	require.Equal(t, 48, host.BlockSize)
	require.Equal(t, []HostTexture{
		{Name: "noiseTex", Type: "texture_3d<f32>", Binding: 1},
		{Name: "sdfSmp0", Type: "sampler", Binding: 2},
	}, host.Textures)
	require.Contains(t, host.WGSL, "@group(0) @binding(1) var noiseTex: texture_3d<f32>;")
	require.Contains(t, host.WGSL, "@group(0) @binding(2) var sdfSmp0: sampler;")
	require.Contains(t, host.WGSL, "return sdShape(p, gsl_host.time, gsl_host.a, gsl_host.lo, gsl_host.b, noiseTex, sdfSmp0, gsl_host.uv);")
}

func TestHost_Refusals(t *testing.T) {
	_, err := Host(&Result{WGSL: "fn other() {}"}, 1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no sdShape")
	_, err = Host(&Result{WGSL: "fn sdShape(p: vec3<f32>, time: f32, m: mat4x4<f32>) -> gsl_shape {}"}, 1)
	require.Error(t, err)
	require.Contains(t, err.Error(), "cannot bind")
	_, err = Host(&Result{WGSL: "fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape {}"}, 0)
	require.Error(t, err)
	require.Contains(t, err.Error(), "must be positive")
}

// The structs and gsl_min are the viewer's own text, so a program reads the
// same host in both.
func TestHostDeclarationsAreTheViewers(t *testing.T) {
	read := func(file, name string) string {
		t.Helper()
		data, err := os.ReadFile("../../examples/viewer/render/" + file)
		require.NoError(t, err)
		m := regexp.MustCompile("(?s)export const " + name + " = `(.*?)`;").FindSubmatch(data)
		require.NotNil(t, m, name)
		return string(m[1])
	}
	require.Equal(t, read("gsl-host.js", "GSL_SHAPE_STRUCT_WGSL"), hostShapeStructWGSL)
	require.Equal(t, read("gsl-host.js", "GSL_MIN_WGSL"), hostMinWGSL)
	require.Equal(t, read("gsl-shading.js", "GSL_MATERIAL_STRUCT_WGSL"), hostMaterialStructWGSL)
}
