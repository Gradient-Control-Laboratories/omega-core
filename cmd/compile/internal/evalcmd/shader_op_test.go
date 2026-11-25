package evalcmd_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/assert"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/require"
)

// A ball with a slider, in a wall of noise: sliders, a texture, one body.
const noisyBall = `@scene(100)
import "noise";
@range(10.0, 45.0) r = 30.0;
foam = max(noise.cellular(p / 10.0).x - 0.3, length(p) - r);
return foam;
`

// shader answers with the field as a library in the language asked for: the code, the two functions' names, the
// uniform block and where each slider sits in it, the textures as files, the bodies and the box.
func TestShader_AnswersForBothTargets(t *testing.T) {
	source := document(t, noisyBall)
	for _, target := range []string{"wgsl", "glsl-es300"} {
		t.Run(target, func(t *testing.T) {
			outDir := filepath.Join(t.TempDir(), "cache")
			r, code := ask(t, source, map[string]any{"op": "shader", "target": target, "outDir": outDir})
			require.Equal(t, 0, code, r)
			assert.Equal(t, target, r["target"])
			src := r["code"].(string)
			assert.Equal(t, map[string]any{"field": "gsl_field", "march": "gsl_march"}, r["entry"])
			assert.True(t, strings.Contains(src, "gsl_field(") && strings.Contains(src, "gsl_march("), "the code defines both")
			assert.False(t, strings.Contains(src, "#version") || strings.Contains(src, "@fragment") || strings.Contains(src, "void main"), "the code is a library")
			assert.True(t, strings.Contains(src, r["uniformBlock"].(string)), "the code declares the block %v", r["uniformBlock"])
			assert.Equal(t, 16.0, r["blockBytes"])
			assert.Equal(t, []any{-50.0, -50.0, -50.0, 50.0, 50.0, 50.0}, r["region"])
			assert.Equal(t, []any{map[string]any{"name": "foam", "index": 1.0}}, r["bodies"])

			sliders := r["sliders"].([]any)
			require.Len(t, sliders, 1)
			s := sliders[0].(map[string]any)
			assert.Equal(t, "r", s["name"])
			assert.Equal(t, "f32", s["type"])
			assert.Equal(t, 4.0, s["offset"])
			assert.Equal(t, 30.0, s["value"])
			assert.Equal(t, 10.0, s["min"])
			assert.Equal(t, 45.0, s["max"])
			assert.True(t, strings.HasPrefix(s["uniform"].(string), "gsl_param_"), "%v", s)

			uniforms := r["uniforms"].([]any)
			require.Len(t, uniforms, 1)
			assert.Equal(t, map[string]any{"name": "time", "type": "f32", "offset": 0.0, "value": []any{0.0}}, uniforms[0])

			textures := r["textures"].([]any)
			require.Len(t, textures, 1)
			tex := textures[0].(map[string]any)
			assert.Equal(t, "noise3d", tex["kind"])
			assert.Equal(t, "rgba8unorm", tex["format"])
			assert.Equal(t, "nearest", tex["filter"])
			assert.Equal(t, []any{64.0, 64.0, 64.0}, tex["size"])
			assert.True(t, strings.Contains(src, tex["uniform"].(string)), "the code declares %v", tex["uniform"])
			path := tex["path"].(string)
			assert.True(t, filepath.IsAbs(path) && filepath.Dir(path) == outDir, "the texture %s is under outDir", path)
			info, err := os.Stat(path)
			require.NoError(t, err)
			assert.Equal(t, int64(64*64*64*4), info.Size())
			if target == "wgsl" {
				assert.Equal(t, 1.0, tex["binding"])
			} else {
				_, has := tex["binding"]
				assert.False(t, has, "a GLSL sampler is bound by name")
			}
		})
	}
}

// A def that names a slider replaces its whole declaration, @range included, so the value is built into the field
// and the response has no slider to move.
func TestShader_DefsOverrideASlider(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "cache")
	r, code := ask(t, document(t, noisyBall), map[string]any{"op": "shader", "target": "wgsl", "outDir": outDir, "defs": "r = 20.0;"})
	require.Equal(t, 0, code, r)
	assert.Empty(t, r["sliders"])
	assert.True(t, strings.Contains(r["code"].(string), "20.0"), "the value is in the code")
}

// A document with no texture needs no outDir.
func TestShader_NoTexturesNoOutDir(t *testing.T) {
	r, code := ask(t, document(t, ball), map[string]any{"op": "shader", "target": "wgsl"})
	require.Equal(t, 0, code, r)
	assert.Equal(t, []any{}, r["textures"])
}

// A target the op does not have, or none, is a malformed request; so is an expr, since the shader is the
// document's returns; a texture with nowhere to go is a failed query.
func TestShader_MalformedRequests(t *testing.T) {
	source := document(t, ball)
	for name, tc := range map[string]struct {
		req  map[string]any
		code int
		want string
	}{
		"no target":      {map[string]any{"op": "shader"}, 2, "target"},
		"unknown target": {map[string]any{"op": "shader", "target": "hlsl"}, 2, "target"},
		"an expr":        {map[string]any{"op": "shader", "target": "wgsl", "expr": "ball"}, 2, "no expr"},
	} {
		t.Run(name, func(t *testing.T) {
			r, code := ask(t, source, tc.req)
			assert.Equal(t, tc.code, code, r)
			assert.True(t, strings.Contains(r["error"].(string), tc.want), "%v names %q", r["error"], tc.want)
		})
	}
	r, code := ask(t, document(t, noisyBall), map[string]any{"op": "shader", "target": "wgsl"})
	assert.Equal(t, 1, code, r)
	assert.True(t, strings.Contains(r["error"].(string), "outDir"), "%v", r["error"])
}
