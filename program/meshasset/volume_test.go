package meshasset

import (
	"math"
	"testing"
)

// unitCube returns a consistently-wound (outward) closed cube spanning
// [origin, origin+size] on all axes.
func unitCube(origin, size float64) *TriMesh {
	lo, hi := origin, origin+size
	v := func(x, y, z float64) [3]float64 { return [3]float64{x, y, z} }
	quads := [][4][3]float64{
		{v(lo, lo, lo), v(lo, hi, lo), v(hi, hi, lo), v(hi, lo, lo)}, // -z
		{v(lo, lo, hi), v(hi, lo, hi), v(hi, hi, hi), v(lo, hi, hi)}, // +z
		{v(lo, lo, lo), v(hi, lo, lo), v(hi, lo, hi), v(lo, lo, hi)}, // -y
		{v(lo, hi, lo), v(lo, hi, hi), v(hi, hi, hi), v(hi, hi, lo)}, // +y
		{v(lo, lo, lo), v(lo, lo, hi), v(lo, hi, hi), v(lo, hi, lo)}, // -x
		{v(hi, lo, lo), v(hi, hi, lo), v(hi, hi, hi), v(hi, lo, hi)}, // +x
	}
	m := &TriMesh{}
	for _, q := range quads {
		m.Triangles = append(m.Triangles,
			Triangle{V0: q[0], V1: q[1], V2: q[2]},
			Triangle{V0: q[0], V1: q[2], V2: q[3]})
	}
	return m
}

func TestVolumeCube(t *testing.T) {
	if got := unitCube(0, 2).Volume(); math.Abs(got-8) > 1e-9 {
		t.Errorf("2-cube volume = %v, want 8", got)
	}
}

func TestVolumeTranslationInvariant(t *testing.T) {
	// The signed-tetrahedron sum is origin-dependent per term; the closed
	// surface must cancel that out.
	if got := unitCube(100, 2).Volume(); math.Abs(got-8) > 1e-6 {
		t.Errorf("translated 2-cube volume = %v, want 8", got)
	}
}

func TestVolumeWindingSign(t *testing.T) {
	// Reversed winding flips the signed sum; Volume returns the magnitude.
	m := unitCube(0, 3)
	for i, tr := range m.Triangles {
		m.Triangles[i] = Triangle{V0: tr.V0, V1: tr.V2, V2: tr.V1}
	}
	if got := m.Volume(); math.Abs(got-27) > 1e-9 {
		t.Errorf("reversed 3-cube volume = %v, want 27", got)
	}
}

func TestVolumeEmpty(t *testing.T) {
	if got := (&TriMesh{}).Volume(); got != 0 {
		t.Errorf("empty mesh volume = %v, want 0", got)
	}
}
