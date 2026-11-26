package main_test

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/protobuf/proto"

	"github.com/Gradient-Control-Laboratories/omega-core/backend/tape"
	"github.com/Gradient-Control-Laboratories/omega-core/cmd/compile/internal/cli"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/assert"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/require"
)

// meshBounds meshes source with no --bbox: how many triangles, and the box they span.
func meshBounds(t *testing.T, source string) (int, [6]float64) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "part.gsl")
	require.NoError(t, os.WriteFile(path, []byte(source), 0644))
	out := filepath.Join(dir, "part.stl")
	var stdout, stderr bytes.Buffer
	require.NoError(t, cli.Run([]string{"-t", "stl", "--algo", "mdc", "--subdivision", "5", "--no-meta", "-o", out, path}, nil,
		&stdout, &stderr), stderr.String())
	data, err := os.ReadFile(out)
	require.NoError(t, err)
	return stlBounds(t, data)
}

// With no --bbox and a returned value with no box of its own (a remap clears .box), a mesh fills the @scene box --
// an edge of the scene's size centred on the origin, as the viewer frames it -- so a ball larger than @scene(30) is
// cut at +-15 (to within a small part of a cell: the mesher places the cut face's vertices on the field inside the
// box).
func TestRun_STL_DefaultBoxIsTheScene(t *testing.T) {
	n, box := meshBounds(t, "@scene(30)\nball = remap(sphere(20), [p.x, p.y, p.z]);\nreturn ball;\n")
	assert.True(t, n > 0)
	for k := 0; k < 3; k++ {
		assert.True(t, math.Abs(box[k]+15) < 0.05, "min %d: %g, want -15", k, box[k])
		assert.True(t, math.Abs(box[k+3]-15) < 0.05, "max %d: %g, want 15", k, box[k+3])
	}
}

// A document with no @scene, returning a value with no box of its own, has no box to take: its mesh fills the box
// its field does (--bbox auto), so a ball far from the origin is meshed whole.
func TestRun_STL_DefaultBoxWithoutASceneIsAuto(t *testing.T) {
	n, box := meshBounds(t, "ball = remap(sphere([100, 0, 0], 5), [p.x, p.y, p.z]);\nreturn ball;\n")
	assert.True(t, n > 0)
	want := [6]float64{95, -5, -5, 105, 5, 5}
	for k := range want {
		assert.True(t, math.Abs(box[k]-want[k]) < 0.05, "bound %d: %g, want %g", k, box[k], want[k])
	}
}

// With no --bbox, a returned value whose .box is bounded is meshed in that box, ahead of @scene: a plate reaching
// past @scene(30)'s +-15 is meshed whole.
func TestRun_STL_DefaultBoxIsTheReturnedBox(t *testing.T) {
	n, box := meshBounds(t, "@scene(30)\nplate = box([0, 0, 0], [40, 30, 5]);\nreturn plate;\n")
	assert.True(t, n > 0)
	want := [6]float64{0, 0, 0, 40, 30, 5}
	for k := range want {
		assert.True(t, math.Abs(box[k]-want[k]) < 1e-3, "bound %d: %g, want %g", k, box[k], want[k])
	}
}

// A box declared on the returned value is the mesh's box exactly as declared, not grown by the crop's boxPad, so a
// part thinner than boxPad is meshed in a box its own size: a slab 1e-4 thick.
func TestRun_STL_DefaultBoxIsTheDeclaredBox(t *testing.T) {
	n, box := meshBounds(t, "slab = box([-1, -1, 0], [1, 1, 0.0001]);\nslab.box = [[-1, -1, 0], [1, 1, 0.0001]];\nreturn slab;\n")
	assert.True(t, n > 0)
	assert.True(t, math.Abs(box[2]) < 1e-5 && math.Abs(box[5]-0.0001) < 1e-5, "z %g..%g, want 0..0.0001", box[2], box[5])
}

// A tape carries the box a mesh of its document would cover: the returned value's .box.
func TestRun_Tape_StampsTheReturnedBox(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "plate.gsl")
	require.NoError(t, os.WriteFile(path, []byte("plate = box([0, 0, 0], [40, 30, 5]);\nreturn plate;\n"), 0644))
	out := filepath.Join(dir, "plate.tape")
	var stdout, stderr bytes.Buffer
	require.NoError(t, cli.Run([]string{"-t", "tape", "-o", out, path}, nil, &stdout, &stderr), stderr.String())
	data, err := os.ReadFile(out)
	require.NoError(t, err)
	var tp tape.Tape
	require.NoError(t, proto.Unmarshal(data, &tp))
	require.True(t, len(tp.Bbox) == 6, "bbox %v", tp.Bbox)
	want := [6]float64{0, 0, 0, 40, 30, 5}
	for k := range want {
		assert.True(t, math.Abs(float64(tp.Bbox[k])-want[k]) < 1e-5, "bound %d: %g, want %g", k, tp.Bbox[k], want[k])
	}
}

// meshBoundsRepaired is meshBounds with the folds repaired (mdc-repair).
func meshBoundsRepaired(t *testing.T, source string) (int, [6]float64) {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "part.gsl")
	require.NoError(t, os.WriteFile(path, []byte(source), 0644))
	out := filepath.Join(dir, "part.stl")
	var stdout, stderr bytes.Buffer
	require.NoError(t, cli.Run([]string{"-t", "stl", "--algo", "mdc-repair", "--subdivision", "5", "--no-meta", "-o", out, path}, nil,
		&stdout, &stderr), stderr.String())
	data, err := os.ReadFile(out)
	require.NoError(t, err)
	return stlBounds(t, data)
}

// TestRun_STL_DefaultBoxIsTheScene with the folds repaired (mdc-repair).
func TestRun_STL_DefaultBoxIsTheScene_MDCRepair(t *testing.T) {
	n, box := meshBoundsRepaired(t, "@scene(30)\nball = remap(sphere(20), [p.x, p.y, p.z]);\nreturn ball;\n")
	assert.True(t, n > 0)
	for k := 0; k < 3; k++ {
		assert.True(t, math.Abs(box[k]+15) < 0.05, "min %d: %g, want -15", k, box[k])
		assert.True(t, math.Abs(box[k+3]-15) < 0.05, "max %d: %g, want 15", k, box[k+3])
	}
}

// TestRun_STL_DefaultBoxWithoutASceneIsAuto with the folds repaired (mdc-repair).
func TestRun_STL_DefaultBoxWithoutASceneIsAuto_MDCRepair(t *testing.T) {
	n, box := meshBoundsRepaired(t, "ball = remap(sphere([100, 0, 0], 5), [p.x, p.y, p.z]);\nreturn ball;\n")
	assert.True(t, n > 0)
	want := [6]float64{95, -5, -5, 105, 5, 5}
	for k := range want {
		assert.True(t, math.Abs(box[k]-want[k]) < 0.05, "bound %d: %g, want %g", k, box[k], want[k])
	}
}

// TestRun_STL_DefaultBoxIsTheReturnedBox with the folds repaired (mdc-repair).
func TestRun_STL_DefaultBoxIsTheReturnedBox_MDCRepair(t *testing.T) {
	n, box := meshBoundsRepaired(t, "@scene(30)\nplate = box([0, 0, 0], [40, 30, 5]);\nreturn plate;\n")
	assert.True(t, n > 0)
	want := [6]float64{0, 0, 0, 40, 30, 5}
	for k := range want {
		assert.True(t, math.Abs(box[k]-want[k]) < 1e-3, "bound %d: %g, want %g", k, box[k], want[k])
	}
}

// TestRun_STL_DefaultBoxIsTheDeclaredBox with the folds repaired (mdc-repair).
func TestRun_STL_DefaultBoxIsTheDeclaredBox_MDCRepair(t *testing.T) {
	n, box := meshBoundsRepaired(t, "slab = box([-1, -1, 0], [1, 1, 0.0001]);\nslab.box = [[-1, -1, 0], [1, 1, 0.0001]];\nreturn slab;\n")
	assert.True(t, n > 0)
	assert.True(t, math.Abs(box[2]) < 1e-5 && math.Abs(box[5]-0.0001) < 1e-5, "z %g..%g, want 0..0.0001", box[2], box[5])
}
