//go:build cgo

package eval

import (
	"math"
	"testing"
)

// A vertex whose gradient is not finite, or is zero, takes the area-weighted average of its triangles' facet
// normals; every other vertex takes its gradient, made unit.
func TestVertexNormalsFallBackToFacets(t *testing.T) {
	// Two triangles in the plane z = 0, wound to face +z, sharing the edge 0-2.
	positions := []float32{0, 0, 0, 1, 0, 0, 1, 1, 0, 0, 1, 0}
	indices := []uint32{0, 1, 2, 0, 2, 3}
	nan := float32(math.NaN())
	gradients := [][3]float32{{0, 3, 0}, {nan, 0, 0}, {0, 0, 0}, {0, 0, -2}}
	got := vertexNormals(gradients, positions, indices)
	want := []float32{0, 1, 0, 0, 0, 1, 0, 0, 1, 0, 0, -1}
	if len(got) != len(want) {
		t.Fatalf("%d floats, want %d", len(got), len(want))
	}
	for i := range want {
		if math.Abs(float64(got[i]-want[i])) > 1e-6 {
			t.Fatalf("normals %v, want %v", got, want)
		}
	}
}
