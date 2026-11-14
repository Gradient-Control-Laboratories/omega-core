package fidget_test

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/internal/fidget"
)

// One uniform block, one 3D texture read by textureLoad, one function.
const glslTestShader = `
struct Host { extent: f32, time: f32 }
@group(0) @binding(0) var<uniform> host: Host;
@group(0) @binding(1) var sdfTex: texture_3d<f32>;

fn field(p: vec3<f32>) -> f32 {
    let texel = textureLoad(sdfTex, vec3<i32>(p * host.extent), 0);
    return texel.r + host.time;
}

@fragment
fn fs_main(@builtin(position) at: vec4<f32>) -> @location(0) vec4<f32> {
    return vec4<f32>(field(at.xyz));
}
`

// The wrapper returns the fragment shader as GLSL ES 3.00 and, for each WGSL
// global a WebGL2 host binds, the name naga gave it.
func TestGLSLES300(t *testing.T) {
	out, err := fidget.GLSLES300(glslTestShader, "fs_main")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"#version 300 es", "sampler3D", "texelFetch", "float field("} {
		if !strings.Contains(out.Source, want) {
			t.Errorf("GLSL lacks %q:\n%s", want, out.Source)
		}
	}
	block, ok := out.UniformBlocks["host"]
	if !ok || len(out.UniformBlocks) != 1 {
		t.Fatalf("uniform blocks = %v, want one, for host", out.UniformBlocks)
	}
	if !strings.Contains(out.Source, "uniform "+block+" {") {
		t.Errorf("GLSL declares no uniform block %q", block)
	}
	sampler, ok := out.Textures["sdfTex"]
	if !ok || len(out.Textures) != 1 {
		t.Fatalf("textures = %v, want one, for sdfTex", out.Textures)
	}
	if !strings.Contains(out.Source, "sampler3D "+sampler+";") {
		t.Errorf("GLSL declares no sampler %q", sampler)
	}
}

// A module naga refuses is an error that carries naga's message.
func TestGLSLES300_Refusals(t *testing.T) {
	for name, tc := range map[string]struct{ wgsl, entry, want string }{
		"does not parse":      {"fn f( {", "fs_main", "expected identifier"},
		"no such entry point": {glslTestShader, "nope", ""},
		"empty":               {"", "fs_main", ""},
	} {
		t.Run(name, func(t *testing.T) {
			out, err := fidget.GLSLES300(tc.wgsl, tc.entry)
			if err == nil {
				t.Fatalf("no error; got:\n%s", out.Source)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q lacks %q", err, tc.want)
			}
		})
	}
}
