package main_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	main "gradientcontrol.com/omega/core/cmd/compile"
	"gradientcontrol.com/omega/core/cmd/compile/internal/evalcmd"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// `compile eval`'s render draws, byte for byte, the picture `compile -t png` draws with the same options: two
// returns, each in its own colour, from a preset and a direction on the light ground.
func TestEvalRenderIsCompilePNG(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "two.gsl")
	require.NoError(t, os.WriteFile(source, []byte("@scene(40)\n@range(5, 15) r = 10.0;\nball = length(p) - r;\n"+
		"cube = max(abs(p.x - 14) - 3, abs(p.y) - 3, abs(p.z) - 3);\nreturn ball, cube;\n"), 0644))

	viaCompile := filepath.Join(dir, "compile.png")
	var stdout, stderr bytes.Buffer
	require.NoError(t, main.Run([]string{"-t", "png", "--views", "iso,-1,1,0.6", "--size", "120x90", "--light",
		"--no-meta", "-o", viaCompile, source}, nil, &stdout, &stderr), stderr.String())

	viaEval := filepath.Join(dir, "eval.png")
	request, err := json.Marshal(map[string]any{"op": "render", "source": source, "out": viaEval,
		"views": "iso,-1,1,0.6", "width": 120, "height": 90, "light": true})
	require.NoError(t, err)
	stdout.Reset()
	code := evalcmd.Run(context.Background(), []string{"-"}, bytes.NewReader(request), &stdout, &stderr)
	require.Equal(t, 0, code, stdout.String())

	a, err := os.ReadFile(viaCompile)
	require.NoError(t, err)
	b, err := os.ReadFile(viaEval)
	require.NoError(t, err)
	assert.True(t, bytes.Equal(a, b), "the same picture")
}
