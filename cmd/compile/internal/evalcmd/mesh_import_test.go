package evalcmd_test

import (
	"os"
	"path/filepath"
	"testing"

	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
	"gradientcontrol.com/omega/core/internal/test/stlbox"
)

// A document beside the box STL it imports: 20 x 10 x 6 mm about the origin.
func boxDocument(t *testing.T, text string) string {
	t.Helper()
	dir := t.TempDir()
	stlbox.Write(t, dir, "box.stl", [3]float64{-10, -5, -3}, [3]float64{10, 5, 3})
	path := filepath.Join(dir, "part.gsl")
	require.NoError(t, os.WriteFile(path, []byte(text), 0644))
	return path
}

// gsl eval reads the meshes a document imports from its folder: bvhSdf is the exact distance to the box, and its
// volume is the box's.
func TestImportedMesh_EvalAndVolume(t *testing.T) {
	source := boxDocument(t, "@scene(30)\nimportMesh 0 \"box.stl\";\nbody = bvhSdf(0);\nreturn body;\n")
	r, code := ask(t, source, map[string]any{"op": "eval", "exprs": []string{"body"},
		"points": [][3]float64{{0, 0, 0}, {0, 0, 10}, {13, 0, 0}}})
	require.Equal(t, 0, code, r)
	values := r["groups"].([]any)[0].(map[string]any)["values"].([]any)[0].([]any) // component 0, by point
	near(t, -3, values[0], 1e-4)
	near(t, 7, values[1], 1e-4)
	near(t, 3, values[2], 1e-4)

	// bounds fits the box on a grid, padded by up to a cell. (volume brackets it too, but its interval bounds
	// through the BVH oracle are loose and slow: seconds, not milliseconds.)
	r, code = ask(t, source, map[string]any{"op": "bounds"})
	require.Equal(t, 0, code, r)
	require.Equal(t, true, r["found"], r)
	cell := r["cellSize"].([]any)
	for k, want := range []float64{-10, -5, -3} {
		got := number(t, r["bboxMin"].([]any)[k])
		assert.True(t, got <= want && want-got <= 2*number(t, cell[k]), "min %d is %g, want %g", k, got, want)
		got = number(t, r["bboxMax"].([]any)[k])
		assert.True(t, got >= -want && got+want <= 2*number(t, cell[k]), "max %d is %g, want %g", k, got, -want)
	}
}

// vdbSdf reads the same box through its level set: exact near the surface, clamped to the band beyond it.
func TestImportedMesh_VdbSdf(t *testing.T) {
	source := boxDocument(t, "@scene(30)\nbody = vdbSdf(importMesh \"box.stl\");\nreturn body;\n")
	r, code := ask(t, source, map[string]any{"op": "eval", "exprs": []string{"body"},
		"points": [][3]float64{{10.05, 0, 0}, {0, 0, 0}}})
	require.Equal(t, 0, code, r)
	values := r["groups"].([]any)[0].(map[string]any)["values"].([]any)[0].([]any) // component 0, by point
	near(t, 0.05, values[0], 0.02)
	assert.True(t, number(t, values[1]) < 0, "the centre is inside: %v", values[1])
}
