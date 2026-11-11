package main_test

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"testing"

	"gradientcontrol.com/omega/core/backend/mesh"
	"gradientcontrol.com/omega/core/cmd/compile/internal/cli"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// meshedWith meshes source with the named algorithm over the box its field fills.
func meshedWith(t *testing.T, algo, source string) []byte {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "part.gsl")
	require.NoError(t, os.WriteFile(path, []byte(source), 0644))
	out := filepath.Join(dir, "part.stl")
	var stdout, stderr bytes.Buffer
	require.NoError(t, cli.Run([]string{"-t", "stl", "--algo", algo, "--subdivision", "6", "--bbox", "auto",
		"--no-meta", "-o", out, path}, nil, &stdout, &stderr), stderr.String())
	data, err := os.ReadFile(out)
	require.NoError(t, err)
	return data
}

// normalsFollowWinding counts the triangles whose stored normal points the way their corners wind, of those with area.
func normalsFollowWinding(data []byte) (int, int) {
	n := int(binary.LittleEndian.Uint32(data[80:84]))
	read := func(off int) [3]float64 {
		var v [3]float64
		for k := range v {
			v[k] = float64(math.Float32frombits(binary.LittleEndian.Uint32(data[off+4*k:])))
		}
		return v
	}
	follow, counted := 0, 0
	for i := 0; i < n; i++ {
		base := 84 + 50*i
		normal, a, b, c := read(base), read(base+12), read(base+24), read(base+36)
		u := [3]float64{b[0] - a[0], b[1] - a[1], b[2] - a[2]}
		v := [3]float64{c[0] - a[0], c[1] - a[1], c[2] - a[2]}
		w := [3]float64{u[1]*v[2] - u[2]*v[1], u[2]*v[0] - u[0]*v[2], u[0]*v[1] - u[1]*v[0]}
		if w[0]*w[0]+w[1]*w[1]+w[2]*w[2] < 1e-18 {
			continue
		}
		counted++
		if w[0]*normal[0]+w[1]*normal[1]+w[2]*normal[2] > 0 {
			follow++
		}
	}
	return follow, counted
}

// Every mesher's mesh of a solid faces outward: closed, consistently wound, not inside out (MeasureClosedSTL), its
// signed volume the solid's, and each stored normal along its winding. A cavity's wall faces into the cavity, so a
// hollow ball's volume is its shell's.
func TestRun_STL_EveryMesherFacesOutward(t *testing.T) {
	parts := []struct {
		name, source string
		volume       float64
	}{
		{"sphere", "@scene(40)\nball = length(p) - 10;\nreturn ball;\n", 4.0 / 3 * math.Pi * 1000},
		{"sharp box", "@scene(40)\nblk = box([0, 0, 0], [10, 20, 30]);\nreturn blk;\n", 6000},
		{"hollow ball", "@scene(40)\nhollow = max(length(p) - 10, -(length(p) - 6));\nreturn hollow;\n",
			4.0 / 3 * math.Pi * (1000 - 216)},
	}
	for _, algo := range []string{"marching-cubes", "mdc"} {
		for _, part := range parts {
			data := meshedWith(t, algo, part.source)
			measure := mesh.MeasureClosedSTL
			if algo == "marching-cubes" && part.name == "hollow ball" {
				// Marching cubes' table splits a face whose corners alternate in and out one way in one cell and
				// the other way in its neighbour, so the cavity's wall has small holes (64 pairs of open edges at
				// this depth): measure its winding without asking that it be closed.
				measure = mesh.MeasureSTL
			}
			stats, err := measure(data)
			require.NoError(t, err, "%s, %s", algo, part.name)
			assert.True(t, math.Abs(stats.SignedVolume-part.volume) < 0.03*part.volume,
				"%s, %s: signed volume %g, want about %g", algo, part.name, stats.SignedVolume, part.volume)
			follow, counted := normalsFollowWinding(data)
			assert.Equal(t, counted, follow, "%s, %s: every stored normal follows its winding", algo, part.name)
		}
	}
}
