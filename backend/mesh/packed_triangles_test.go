package mesh_test

import (
	"encoding/base64"
	"encoding/binary"
	"math"
	"reflect"
	"testing"

	"gradientcontrol.com/omega/core/backend/mesh"
)

func TestDecodePackedTriangles(t *testing.T) {
	var buf [36]byte
	for i := 0; i < 9; i++ {
		binary.LittleEndian.PutUint32(buf[i*4:], math.Float32bits(float32(i)))
	}
	encoded := base64.StdEncoding.EncodeToString(buf[:])

	floats, indices, err := mesh.DecodePackedTriangles(encoded)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(floats) != 9 {
		t.Fatalf("floats len = %d, want 9", len(floats))
	}
	if len(indices) != 3 {
		t.Fatalf("indices len = %d, want 3", len(indices))
	}

	for i := 0; i < 9; i++ {
		if floats[i] != float32(i) {
			t.Fatalf("floats[%d] = %v, want %v", i, floats[i], float32(i))
		}
	}
	wantIdx := []uint32{0, 1, 2}
	if !reflect.DeepEqual(indices, wantIdx) {
		t.Fatalf("indices = %v, want %v", indices, wantIdx)
	}
}
