package main_test

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"testing"

	"gradientcontrol.com/omega/core/cmd/compile/internal/cli"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// stlTriangles reads a binary STL's triangles.
func stlTriangles(t *testing.T, data []byte) [][3][3]float32 {
	t.Helper()
	require.True(t, len(data) >= 84)
	n := int(binary.LittleEndian.Uint32(data[80:84]))
	tris := make([][3][3]float32, n)
	for i := 0; i < n; i++ {
		for v := 0; v < 3; v++ {
			for k := 0; k < 3; k++ {
				off := 84 + 50*i + 12 + 12*v + 4*k
				tris[i][v][k] = math.Float32frombits(binary.LittleEndian.Uint32(data[off:]))
			}
		}
	}
	return tris
}

// closedVolume welds the triangles' corners by exact position and reports whether every edge is shared by exactly
// two triangles -- a closed, manifold surface -- and the volume it encloses.
func closedVolume(tris [][3][3]float32) (bool, float64) {
	edges := map[[2][3]float32]int{}
	volume := 0.0
	for _, tri := range tris {
		for v := 0; v < 3; v++ {
			a, b := tri[v], tri[(v+1)%3]
			if a == b {
				return false, 0 // a triangle with two corners in one place has no area
			}
			if b[0] < a[0] || (b[0] == a[0] && (b[1] < a[1] || (b[1] == a[1] && b[2] < a[2]))) {
				a, b = b, a
			}
			edges[[2][3]float32{a, b}]++
		}
		p, q, r := tri[0], tri[1], tri[2]
		volume += (float64(p[0])*(float64(q[1])*float64(r[2])-float64(q[2])*float64(r[1])) -
			float64(p[1])*(float64(q[0])*float64(r[2])-float64(q[2])*float64(r[0])) +
			float64(p[2])*(float64(q[0])*float64(r[1])-float64(q[1])*float64(r[0]))) / 6
	}
	for _, uses := range edges {
		if uses != 2 {
			return false, volume
		}
	}
	return true, volume
}

func meshed(t *testing.T, source string, args ...string) []byte {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "part.gsl")
	require.NoError(t, os.WriteFile(path, []byte(source), 0644))
	out := filepath.Join(dir, "part.stl")
	var stdout, stderr bytes.Buffer
	require.NoError(t, cli.Run(append(append([]string{"-t", "stl", "--algo", "mdc", "--no-meta", "-o", out}, args...),
		path), nil, &stdout, &stderr))
	data, err := os.ReadFile(out)
	require.NoError(t, err)
	return data
}

// A box shelled 2 mm with the exact box field: its inner wall lies where every component of d is negative, where
// length(max(d, 0)) is sqrt(0). Its gradient there is zero, not 0/0, so the wall meshes closed, with no slivers,
// and the shell's volume is the two boxes' difference.
func TestRun_STL_ShellOfExactBoxIsClosed(t *testing.T) {
	source := "d = abs(p) - [20.123, 15.0771, 10.0517];\n" +
		"outer = length(max(d, 0)) + min(max(d.x, d.y, d.z), 0);\n" +
		"t = 2.0;\n" +
		"return max(outer, -(outer + t));\n"
	closed, volume := closedVolume(stlTriangles(t, meshed(t, source, "--subdivision", "8", "--bbox", "auto")))
	assert.True(t, closed)
	a, b, c := 2*20.123, 2*15.0771, 2*10.0517
	want := a*b*c - (a-4)*(b-4)*(c-4)
	assert.True(t, math.Abs(volume-want) < 0.005*want)
}

// The same shell with round dimensions: its inner wall lies exactly where min(max(d.x, d.y, d.z), 0) ties, whose
// gradient there is the constant's, zero. A zero gradient names no plane, so the mesh keeps no NaN vertex.
func TestRun_STL_ShellWithRoundDimensionsIsClosed(t *testing.T) {
	source := "d = abs(p) - [20, 15, 10];\n" +
		"outer = length(max(d, 0)) + min(max(d.x, d.y, d.z), 0);\n" +
		"return max(outer, -(outer + 2));\n"
	closed, volume := closedVolume(stlTriangles(t, meshed(t, source, "--subdivision", "8", "--bbox", "auto")))
	assert.True(t, closed)
	want := 40.0*30*20 - 36.0*26*16
	assert.True(t, math.Abs(volume-want) < 0.005*want)
}
