// Package stlbox writes a box as a binary STL, for tests that import a mesh
// whose volume and extent they know.
package stlbox

import (
	"os"
	"path/filepath"
	"testing"

	"gradientcontrol.com/omega/core/program/meshasset"
)

// Bytes is the box from lo to hi as binary STL: twelve triangles wound
// counter-clockwise seen from outside, so their normals point out.
func Bytes(lo, hi [3]float64) []byte {
	corner := func(i, j, k int) [3]float64 {
		pick := func(bit int, axis int) float64 {
			if bit == 0 {
				return lo[axis]
			}
			return hi[axis]
		}
		return [3]float64{pick(i, 0), pick(j, 1), pick(k, 2)}
	}
	centre := [3]float64{(lo[0] + hi[0]) / 2, (lo[1] + hi[1]) / 2, (lo[2] + hi[2]) / 2}
	// Each face as its four corners in order around it.
	faces := [6][4][3]int{
		{{0, 0, 0}, {0, 0, 1}, {0, 1, 1}, {0, 1, 0}},
		{{1, 0, 0}, {1, 1, 0}, {1, 1, 1}, {1, 0, 1}},
		{{0, 0, 0}, {1, 0, 0}, {1, 0, 1}, {0, 0, 1}},
		{{0, 1, 0}, {0, 1, 1}, {1, 1, 1}, {1, 1, 0}},
		{{0, 0, 0}, {0, 1, 0}, {1, 1, 0}, {1, 0, 0}},
		{{0, 0, 1}, {1, 0, 1}, {1, 1, 1}, {0, 1, 1}},
	}
	mesh := &meshasset.TriMesh{Bbox: meshasset.Bbox{Min: lo, Max: hi}}
	for _, f := range faces {
		var q [4][3]float64
		for n, c := range f {
			q[n] = corner(c[0], c[1], c[2])
		}
		for _, tri := range [2][3][3]float64{{q[0], q[1], q[2]}, {q[0], q[2], q[3]}} {
			if !outward(tri, centre) {
				tri[1], tri[2] = tri[2], tri[1]
			}
			mesh.Triangles = append(mesh.Triangles, meshasset.Triangle{V0: tri[0], V1: tri[1], V2: tri[2]})
		}
	}
	data, err := meshasset.MarshalBinarySTL(mesh)
	if err != nil {
		panic(err) // a box always has triangles
	}
	return data
}

// outward reports whether the triangle's right-hand normal points away from
// centre.
func outward(t [3][3]float64, centre [3]float64) bool {
	var u, v, n, out [3]float64
	for k := 0; k < 3; k++ {
		u[k] = t[1][k] - t[0][k]
		v[k] = t[2][k] - t[0][k]
		out[k] = (t[0][k]+t[1][k]+t[2][k])/3 - centre[k]
	}
	n[0] = u[1]*v[2] - u[2]*v[1]
	n[1] = u[2]*v[0] - u[0]*v[2]
	n[2] = u[0]*v[1] - u[1]*v[0]
	return n[0]*out[0]+n[1]*out[1]+n[2]*out[2] > 0
}

// Write writes the box to dir/name and returns its path.
func Write(t testing.TB, dir, name string, lo, hi [3]float64) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, Bytes(lo, hi), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
