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

// TestKeepApartLeavesTwoVerticesOffOnePoint checks a vertex whose new place is another vertex's stays where it was:
// both a vertex sent onto one that is not moving, and the second of two moving vertices sent to the same new point.
// Welding by position would otherwise join them and open the mesh.
func TestKeepApartLeavesTwoVerticesOffOnePoint(t *testing.T) {
	verts, tris := sphereSTL(5, 1)
	m := weld(writeSTL(verts, tris), len(tris))
	r := &repair{mesh: m, field: sphereField(5)}
	free := vec{0.1, 0.2, 5}
	order := []int32{0, 1, 2}
	pts := []vec{m.verts[3], free, free} // 0 onto vertex 3, which stays; 1 and 2 onto one new point
	r.keepApart(order, pts)
	if pts[0] != m.verts[0] {
		t.Fatalf("vertex 0 sent onto vertex 3 moved to %v", pts[0])
	}
	if pts[1] != free {
		t.Fatalf("vertex 1, the first sent to a free point, did not move there: %v", pts[1])
	}
	if pts[2] != m.verts[2] {
		t.Fatalf("vertex 2, sent to the point vertex 1 took, moved to %v", pts[2])
	}
}

// TestKeepApartLeavesNoTwoVerticesOnOnePoint checks a vertex kept where it is holds its place ahead of one sent
// there: vertex 0 sent onto vertex 1's place while vertex 1, also moving, is sent back to where it is (as one already
// on the zero set is, when a neighbour's walk ends on the same point). And a chain: vertex 2 and vertex 3 sent to one
// free point, vertex 3 kept where it is, and vertex 4 sent to vertex 3's place, which keeping vertex 3 takes. No two
// may end on one point.
func TestKeepApartLeavesNoTwoVerticesOnOnePoint(t *testing.T) {
	verts, tris := sphereSTL(5, 1)
	m := weld(writeSTL(verts, tris), len(tris))
	r := &repair{mesh: m, field: sphereField(5)}
	free := vec{0.1, 0.2, 5}
	order := []int32{0, 1, 2, 3, 4}
	pts := []vec{m.verts[1], m.verts[1], free, free, m.verts[3]}
	r.keepApart(order, pts)
	want := []vec{m.verts[0], m.verts[1], free, m.verts[3], m.verts[4]}
	for k := range order {
		if pts[k] != want[k] {
			t.Fatalf("vertex %d ended at %v, want %v", order[k], pts[k], want[k])
		}
	}
	seen := make(map[[3]float32]int32)
	for v := range m.verts {
		p := m.verts[v]
		for k, u := range order {
			if u == int32(v) {
				p = pts[k]
			}
		}
		if other, ok := seen[key32(p)]; ok {
			t.Fatalf("vertices %d and %d end on one point %v", other, v, p)
		}
		seen[key32(p)] = int32(v)
	}
}

// TestFoldsLeavesTrianglesOffTheZeroSetAlone checks a triangle where the field jumps across zero is not taken for a
// fold: a sphere whose field, over its top cap, is -8 with a gradient pointing in, as a face cut by a choice between
// branches reads, comes back as it was given.
func TestFoldsLeavesTrianglesOffTheZeroSetAlone(t *testing.T) {
	const r = 5.0
	verts, tris := sphereSTL(r, 3)
	in := writeSTL(verts, tris)
	sphere := sphereField(r)
	field := func(xs, ys, zs []float32) ([]Sample, error) {
		out, err := sphere(xs, ys, zs)
		for i := range out {
			if zs[i] > 4 {
				out[i] = Sample{Value: -8, Dx: -out[i].Dx, Dy: -out[i].Dy, Dz: -out[i].Dz}
			}
		}
		return out, err
	}
	out, stats, err := Folds(in, field)
	if err != nil {
		t.Fatal(err)
	}
	if stats.BackwardBefore != 0 || stats.Passes != 0 || stats.Moved != 0 {
		t.Fatalf("triangles off the zero set were taken for folds: %+v", stats)
	}
	if string(out) != string(in) {
		t.Fatal("a sphere with triangles off the zero set came back changed")
	}
}

// TestKeepNearHoldsAVertexSentPastItsEdges checks a vertex whose new place is farther than its edges' mean length
// stays where it was, and one sent less far moves.
func TestKeepNearHoldsAVertexSentPastItsEdges(t *testing.T) {
	verts, tris := sphereSTL(5, 2)
	m := weld(writeSTL(verts, tris), len(tris))
	r := &repair{mesh: m, field: sphereField(5)}
	edge := norm(sub(m.verts[m.ringOf(0)[0]], m.verts[0]))
	near := add(m.verts[0], vec{0, 0.2 * edge, 0})
	far := add(m.verts[1], vec{0, 3 * edge, 0})
	order := []int32{0, 1}
	pts := []vec{near, far}
	r.keepNear(order, pts)
	if pts[0] != near {
		t.Fatalf("vertex 0, sent a fifth of an edge, did not move there: %v", pts[0])
	}
	if pts[1] != m.verts[1] {
		t.Fatalf("vertex 1, sent three edges, moved to %v", pts[1])
	}
}

// TestFoldsKeepsTheMeshWhenRepairsGrowItsArea checks no pass that makes the mesh's area larger is kept: a pleated
// sphere of radius 5 repaired against a sphere of radius 5.3, onto which every pass walks the moved vertices out,
// comes back as it was given. 5.3 is near enough that the moves stay within an edge (keepNear lets them through) and
// far enough that they add more area than righting the pleats gives back.
func TestFoldsKeepsTheMeshWhenRepairsGrowItsArea(t *testing.T) {
	verts, tris := sphereSTL(5, 4)
	m := weld(writeSTL(verts, tris), len(tris))
	for _, v := range []int32{40, 400, 900} {
		w := m.ringOf(v)[0]
		p := add(m.verts[v], scale(sub(m.verts[w], m.verts[v]), 1.6))
		m.verts[v] = scale(p, 5/norm(p))
	}
	in := m.write(make([]byte, 80))
	out, stats, err := Folds(in, sphereField(5.3))
	if err != nil {
		t.Fatal(err)
	}
	if stats.BackwardBefore == 0 {
		t.Fatalf("the pleats made no backward triangle: %+v", stats)
	}
	if stats.Passes != 0 || stats.Moved != 0 {
		t.Fatalf("a pass that grew the area was kept: %+v", stats)
	}
	if string(out) != string(in) {
		t.Fatal("every pass grew the area, yet the mesh came back changed")
	}
}

// TestFoldsReturnsTheMeshWhenNoPassHelps checks a mesh no pass improves comes back as it was given: a sphere whose
// every triangle faces in, which moving vertices cannot turn round.
func TestFoldsReturnsTheMeshWhenNoPassHelps(t *testing.T) {
	verts, tris := sphereSTL(5, 2)
	for i := range tris {
		tris[i][1], tris[i][2] = tris[i][2], tris[i][1]
	}
	in := writeSTL(verts, tris)
	out, stats, err := Folds(in, sphereField(5))
	if err != nil {
		t.Fatal(err)
	}
	if stats.BackwardBefore != len(tris) || stats.BackwardAfter != len(tris) || stats.Passes != 0 || stats.Moved != 0 {
		t.Fatalf("an inside-out sphere was changed: %+v", stats)
	}
	if string(out) != string(in) {
		t.Fatal("no pass helped, yet the mesh came back changed")
	}
}
