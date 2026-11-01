package meshasset

import (
	"math"
	"testing"
)

func TestAreaCube(t *testing.T) {
	// 2-cube: 6 faces x 2x2 = 24.
	if got := unitCube(0, 2).Area(); math.Abs(got-24) > 1e-9 {
		t.Errorf("2-cube area = %v, want 24", got)
	}
}

func TestAreaWindingIndependent(t *testing.T) {
	m := unitCube(0, 3)
	for i, tr := range m.Triangles {
		m.Triangles[i] = Triangle{V0: tr.V0, V1: tr.V2, V2: tr.V1}
	}
	if got := m.Area(); math.Abs(got-54) > 1e-9 {
		t.Errorf("reversed 3-cube area = %v, want 54", got)
	}
}

func TestAreaEmpty(t *testing.T) {
	if got := (&TriMesh{}).Area(); got != 0 {
		t.Errorf("empty mesh area = %v, want 0", got)
	}
}
