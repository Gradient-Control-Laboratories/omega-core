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

// meshBounds meshes source with no --bbox: how many triangles, and the box they span.
func meshBounds(t *testing.T, source string) (int, [6]float64) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "part.gsl")
	require.NoError(t, os.WriteFile(path, []byte(source), 0644))
	out := filepath.Join(dir, "part.stl")
	var stdout, stderr bytes.Buffer
	require.NoError(t, cli.Run([]string{"-t", "stl", "--subdivision", "5", "--no-meta", "-o", out, path}, nil,
		&stdout, &stderr), stderr.String())
	data, err := os.ReadFile(out)
	require.NoError(t, err)
	return stlBounds(t, data)
}

// With no --bbox, a mesh fills the @scene box -- an edge of the scene's size centred on the origin, as the viewer
// frames it -- so a ball larger than @scene(30) is cut at +-15 (to within a small part of a cell: the mesher
// places the cut face's vertices on the field inside the box).
func TestRun_STL_DefaultBoxIsTheScene(t *testing.T) {
	n, box := meshBounds(t, "@scene(30)\nball = length(p) - 20;\nreturn ball;\n")
	assert.True(t, n > 0)
	for k := 0; k < 3; k++ {
		assert.True(t, math.Abs(box[k]+15) < 0.05, "min %d: %g, want -15", k, box[k])
		assert.True(t, math.Abs(box[k+3]-15) < 0.05, "max %d: %g, want 15", k, box[k+3])
	}
}

// A document with no @scene has no box to take: its mesh fills the box its field does (--bbox auto), so a ball far
// from the origin is meshed whole.
func TestRun_STL_DefaultBoxWithoutASceneIsAuto(t *testing.T) {
	n, box := meshBounds(t, "ball = length(p - [100, 0, 0]) - 5;\nreturn ball;\n")
	assert.True(t, n > 0)
	want := [6]float64{95, -5, -5, 105, 5, 5}
	for k := range want {
		assert.True(t, math.Abs(box[k]-want[k]) < 0.05, "bound %d: %g, want %g", k, box[k], want[k])
	}
}
