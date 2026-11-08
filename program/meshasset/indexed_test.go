package meshasset

import "testing"

func TestFromIndexed(t *testing.T) {
	// Two triangles forming a 10x5 quad in z=0.
	positions := []float32{0, 0, 0, 10, 0, 0, 0, 5, 0, 10, 5, 0}
	tris := []int32{0, 1, 2, 1, 3, 2}
	m, hash := FromIndexed(positions, tris)

	if len(m.Triangles) != 2 {
		t.Fatalf("want 2 triangles, got %d", len(m.Triangles))
	}
	if m.Triangles[0].V1 != ([3]float64{10, 0, 0}) {
		t.Fatalf("tri0 V1 = %v", m.Triangles[0].V1)
	}
	if !m.Bbox.IsValid() || m.Bbox.Min != ([3]float64{0, 0, 0}) || m.Bbox.Max != ([3]float64{10, 5, 0}) {
		t.Fatalf("bbox = %+v", m.Bbox)
	}
	if hash == "" {
		t.Fatal("empty hash")
	}
	// Identical input hashes identically (registry dedup); different input differs.
	if _, h2 := FromIndexed(positions, tris); h2 != hash {
		t.Fatal("hash not stable for identical input")
	}
	if _, h3 := FromIndexed([]float32{0, 0, 0, 1, 0, 0, 0, 1, 0}, []int32{0, 1, 2}); h3 == hash {
		t.Fatal("hash collided for different input")
	}
}
