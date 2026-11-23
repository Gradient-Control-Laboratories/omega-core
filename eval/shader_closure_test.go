//go:build cgo

package eval_test

import (
	"testing"

	"gradientcontrol.com/omega/core/eval"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// A slider read by a field that a closure's remap re-derives: unionRoundExact
// evaluates each of its closures at a moved point, inside its own function, so
// the helper that re-derives `a - r` there needs the slider handed down to it.
const closureSliderDoc = `@scene(40)
@range(1, 5) r = 2.0;
a = sphere(8);
b = boxCentered([6, 0, 0], [8, 8, 8]);
blob = unionRoundExact(a, b, r, a.dp, b.dp);
return blob;
`

// The module must define every name it reads: the GLSL target is a transpile of
// the WGSL, which refuses one that does not.
func TestShader_SliderInsideAClosuresRemap(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	for _, target := range []string{"wgsl", "glsl-es300"} {
		t.Run(target, func(t *testing.T) {
			res, err := e.Shader(t.Context(), nil, closureSliderDoc, "", eval.ShaderOptions{Target: target, OutDir: t.TempDir()})
			require.NoError(t, err)
			require.Len(t, res.Sliders, 1)
			require.Equal(t, "r", res.Sliders[0].Name)
		})
	}
}
