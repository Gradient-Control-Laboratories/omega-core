package main_test

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"testing"

	"gradientcontrol.com/omega/core/cmd/compile/internal/cli"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// --bbox auto finds the box of the document's returns when a comment stands above its return.
func TestRun_STL_BboxAutoAfterACommentedReturn(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "part.gsl")
	require.NoError(t, os.WriteFile(path, []byte("ball = length(p - [100, 0, 0]) - 5;\n// the part\nreturn ball;\n"), 0644))
	out := filepath.Join(dir, "part.stl")
	var stdout, stderr bytes.Buffer
	require.NoError(t, cli.Run([]string{"-t", "stl", "--subdivision", "5", "--bbox", "auto", "--no-meta", "-o", out,
		path}, nil, &stdout, &stderr), stderr.String())
	data, err := os.ReadFile(out)
	require.NoError(t, err)
	n, box := stlBounds(t, data)
	assert.True(t, n > 0)
	want := [6]float64{95, -5, -5, 105, 5, 5}
	for k := range want {
		assert.True(t, math.Abs(box[k]-want[k]) < 0.05, "bound %d: %g, want %g", k, box[k], want[k])
	}
}
