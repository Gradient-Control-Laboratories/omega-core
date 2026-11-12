package main_test

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"testing"

	main "gradientcontrol.com/omega/core/cmd/compile"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
	"gradientcontrol.com/omega/core/internal/test/stlbox"
)

// importedBox writes a 20 x 10 x 6 mm box STL beside a document and returns the document's path.
func importedBox(t *testing.T, source string) string {
	t.Helper()
	dir := t.TempDir()
	stlbox.Write(t, dir, "box.stl", [3]float64{-10, -5, -3}, [3]float64{10, 5, 3})
	path := filepath.Join(dir, "part.gsl")
	require.NoError(t, os.WriteFile(path, []byte(source), 0644))
	return path
}

func compileFile(t *testing.T, path string, args ...string) ([]byte, error) {
	t.Helper()
	out := filepath.Join(t.TempDir(), "out")
	var stdout, stderr bytes.Buffer
	err := main.Run(append(append([]string{"--no-meta", "-o", out}, args...), path), nil, &stdout, &stderr)
	if err != nil {
		return nil, err
	}
	data, readErr := os.ReadFile(out)
	require.NoError(t, readErr)
	return data, nil
}

// Meshing a document that reads an imported box through bvhSdf or vdbSdf gives a mesh of the box's volume, 1200 mm³.
// The box is found beside the document, by either form of importMesh. Only the volume is checked: dual contouring
// leaves a few degenerate triangles and open edges at the sharp edges of the exact BVH field, which an analytic box
// does not, and that is the mesher's, not the import's.
func TestRun_STL_ImportedMeshThroughEachOracle(t *testing.T) {
	box := []string{"-t", "stl", "--algo", "mdc", "--subdivision", "6", "--bbox", "-12,-7,-5,12,7,5"}
	for name, source := range map[string]string{
		"bvhSdf, numbered": "importMesh 0 \"box.stl\";\nreturn bvhSdf(0);\n",
		"bvhSdf, inline":   "return bvhSdf(importMesh \"box.stl\");\n",
		"vdbSdf":           "importMesh 2 \"box.stl\";\nreturn vdbSdf(2);\n",
		"both on one slot": "importMesh 0 \"box.stl\";\nreturn max(bvhSdf(0), vdbSdf(0));\n",
	} {
		t.Run(name, func(t *testing.T) {
			data, err := compileFile(t, importedBox(t, source), box...)
			require.NoError(t, err)
			volume := signedVolume(stlTriangles(t, data))
			assert.True(t, math.Abs(volume-1200) < 0.02*1200, "volume %g, want 1200", volume)
		})
	}
}

// With no box named and no @scene, --bbox auto finds the field's box by evaluating it, which loads the mesh too.
func TestRun_STL_ImportedMeshBboxAuto(t *testing.T) {
	data, err := compileFile(t, importedBox(t, "importMesh 0 \"box.stl\";\nreturn bvhSdf(0);\n"),
		"-t", "stl", "--algo", "mdc", "--subdivision", "6")
	require.NoError(t, err)
	volume := signedVolume(stlTriangles(t, data))
	assert.True(t, math.Abs(volume-1200) < 0.02*1200, "volume %g, want 1200", volume)
}

func TestRun_PNG_ImportedMesh(t *testing.T) {
	data, err := compileFile(t, importedBox(t, "@scene(30)\nimportMesh 0 \"box.stl\";\nreturn bvhSdf(0);\n"),
		"-t", "png", "--size", "64x48")
	require.NoError(t, err)
	rect, _ := decoded(t, data)
	assert.Equal(t, 64, rect.Dx())
}

// A tape carries the mesh it reads, so it evaluates without the STL.
func TestRun_Tape_CarriesTheImportedMesh(t *testing.T) {
	path := importedBox(t, "importMesh 0 \"box.stl\";\nreturn bvhSdf(0);\n")
	with, err := compileFile(t, path, "-t", "tape")
	require.NoError(t, err)
	assert.True(t, len(with) > 12*9*4, "the tape (%d bytes) holds the box's triangles", len(with))
}

func TestRun_ImportedMeshFailures(t *testing.T) {
	_, err := compileFile(t, importedBox(t, "@scene(30)\nimportMesh 0 \"gone.stl\";\nreturn bvhSdf(0);\n"), "-t", "stl")
	require.Error(t, err)
	assert.Contains(t, err.Error(), `importMesh "gone.stl"`)

	_, err = compileFile(t, importedBox(t, "@scene(30)\nreturn bvhSdf(5);\n"), "-t", "stl")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no importMesh declares slot 5")

	// wgsl leaves the textures to its host, so it compiles without the file.
	_, err = compileFile(t, importedBox(t, "importMesh 0 \"gone.stl\";\nreturn bvhSdf(0);\n"), "-t", "wgsl")
	require.NoError(t, err)
}

// signedVolume is the volume the triangles enclose, by the divergence theorem; a degenerate triangle adds nothing.
func signedVolume(tris [][3][3]float32) float64 {
	volume := 0.0
	for _, tri := range tris {
		p, q, r := tri[0], tri[1], tri[2]
		volume += (float64(p[0])*(float64(q[1])*float64(r[2])-float64(q[2])*float64(r[1])) -
			float64(p[1])*(float64(q[0])*float64(r[2])-float64(q[2])*float64(r[0])) +
			float64(p[2])*(float64(q[0])*float64(r[1])-float64(q[1])*float64(r[0]))) / 6
	}
	return volume
}
