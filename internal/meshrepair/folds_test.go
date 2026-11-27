package meshrepair

import (
	"encoding/binary"
	"math"
	"testing"
)

// sphereSTL is a sphere of radius r about the origin, an octahedron subdivided level times with its vertices pushed
// out onto the sphere, as a binary STL facing outward.
func sphereSTL(r float64, level int) ([]vec, [][3]int32) {
	verts := []vec{{1, 0, 0}, {-1, 0, 0}, {0, 1, 0}, {0, -1, 0}, {0, 0, 1}, {0, 0, -1}}
	tris := [][3]int32{{0, 2, 4}, {2, 1, 4}, {1, 3, 4}, {3, 0, 4}, {2, 0, 5}, {1, 2, 5}, {3, 1, 5}, {0, 3, 5}}
	for range level {
		mid := map[[2]int32]int32{}
		half := func(a, b int32) int32 {
			key := [2]int32{min(a, b), max(a, b)}
			if m, ok := mid[key]; ok {
				return m
			}
			verts = append(verts, scale(add(verts[a], verts[b]), 0.5))
			mid[key] = int32(len(verts) - 1)
			return mid[key]
		}
		var next [][3]int32
		for _, t := range tris {
			ab, bc, ca := half(t[0], t[1]), half(t[1], t[2]), half(t[2], t[0])
			next = append(next, [3]int32{t[0], ab, ca}, [3]int32{ab, t[1], bc}, [3]int32{ca, bc, t[2]}, [3]int32{ab, bc, ca})
		}
		tris = next
	}
	for i, p := range verts {
		verts[i] = scale(p, r/norm(p))
	}
	return verts, tris
}

func writeSTL(verts []vec, tris [][3]int32) []byte {
	m := &mesh{verts: verts, tris: tris}
	return m.write(make([]byte, 80))
}

func sphereField(r float64) Field {
	return func(xs, ys, zs []float32) ([]Sample, error) {
		out := make([]Sample, len(xs))
		for i := range xs {
			p := vec{float64(xs[i]), float64(ys[i]), float64(zs[i])}
			l := norm(p)
			out[i] = Sample{Value: float32(l - r), Dx: float32(p[0] / l), Dy: float32(p[1] / l), Dz: float32(p[2] / l)}
		}
		return out, nil
	}
}

// TestFoldsRepairsPleatsOnASphere pushes a few vertices of a sphere past a neighbour, along the surface, so the
// triangles round them turn over without the mesh opening, and checks the repair turns every one back, leaves the
// moved vertices on the sphere, and keeps the triangles as they were.
func TestFoldsRepairsPleatsOnASphere(t *testing.T) {
	const r = 5.0
	verts, tris := sphereSTL(r, 4)
	m := weld(writeSTL(verts, tris), len(tris))
	for _, v := range []int32{40, 400, 900} {
		ring := m.ringOf(v)
		w := ring[0]
		p := add(m.verts[v], scale(sub(m.verts[w], m.verts[v]), 1.6))
		m.verts[v] = scale(p, r/norm(p))
	}
	folded := m.write(make([]byte, 80))

	out, stats, err := Folds(folded, sphereField(r))
	if err != nil {
		t.Fatal(err)
	}
	if stats.BackwardBefore == 0 {
		t.Fatalf("the pleats made no backward triangle: %+v", stats)
	}
	if stats.BackwardAfter != 0 {
		t.Fatalf("backward triangles left after repair: %+v", stats)
	}
	if got := int(binary.LittleEndian.Uint32(out[80:84])); got != len(tris) || len(out) != 84+50*len(tris) {
		t.Fatalf("repaired STL has %d triangles in %d bytes, want %d", got, len(out), len(tris))
	}
	after := weld(out, len(tris))
	if len(after.verts) != len(m.verts) {
		t.Fatalf("repaired mesh welds to %d vertices, want %d", len(after.verts), len(m.verts))
	}
	for i, p := range after.verts {
		if d := math.Abs(norm(p) - r); d > 1e-5 {
			t.Fatalf("vertex %d at %v is %g off the sphere", i, p, d)
		}
	}
	for i := range tris {
		for k := 0; k < 3; k++ {
			if a, b := m.tris[i][k], after.tris[i][k]; a != b {
				t.Fatalf("triangle %d corner %d is vertex %d, was %d", i, k, b, a)
			}
		}
	}
}

// TestFoldsLeavesACleanMeshAlone checks a mesh with no backward triangle comes back with its vertices where they
// were.
func TestFoldsLeavesACleanMeshAlone(t *testing.T) {
	verts, tris := sphereSTL(5, 3)
	in := writeSTL(verts, tris)
	out, stats, err := Folds(in, sphereField(5))
	if err != nil {
		t.Fatal(err)
	}
	if stats.BackwardBefore != 0 || stats.Passes != 0 || stats.Moved != 0 {
		t.Fatalf("a clean sphere was repaired: %+v", stats)
	}
	if string(out) != string(in) {
		t.Fatal("a clean sphere came back changed")
	}
}
