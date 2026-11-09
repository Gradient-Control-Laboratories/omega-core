package mesh

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"
)

// makeCubeSTL generates binary STL bytes for a unit cube [0..s]^3 centered on the
// given offset with outward-facing normals.
func makeCubeSTL(size float32, offset [3]float32) []byte {
	var buf bytes.Buffer
	buf.Write(make([]byte, 80)) // header
	binary.Write(&buf, binary.LittleEndian, uint32(12))

	v := func(x, y, z float32) [3]float32 {
		return [3]float32{x*size + offset[0], y*size + offset[1], z*size + offset[2]}
	}

	quad := func(v0, v1, v2, v3 [3]float32) {
		tri := func(a, b, c [3]float32) {
			buf.Write(make([]byte, 12)) // normal
			binary.Write(&buf, binary.LittleEndian, a)
			binary.Write(&buf, binary.LittleEndian, b)
			binary.Write(&buf, binary.LittleEndian, c)
			buf.Write([]byte{0, 0}) // attribute byte count
		}
		tri(v0, v1, v2)
		tri(v0, v2, v3)
	}

	// -Z, +Z
	quad(v(0, 0, 0), v(0, 1, 0), v(1, 1, 0), v(1, 0, 0))
	quad(v(0, 0, 1), v(1, 0, 1), v(1, 1, 1), v(0, 1, 1))
	// -X, +X
	quad(v(0, 0, 0), v(0, 0, 1), v(0, 1, 1), v(0, 1, 0))
	quad(v(1, 0, 0), v(1, 1, 0), v(1, 1, 1), v(1, 0, 1))
	// -Y, +Y
	quad(v(0, 0, 0), v(1, 0, 0), v(1, 0, 1), v(0, 0, 1))
	quad(v(0, 1, 0), v(0, 1, 1), v(1, 1, 1), v(1, 1, 0))

	return buf.Bytes()
}

func TestMeasureSTLUnitCube(t *testing.T) {
	const s = 10.0
	stl := makeCubeSTL(s, [3]float32{1, 2, 3})
	stats, err := MeasureSTL(stl)
	if err != nil {
		t.Fatalf("MeasureSTL: %v", err)
	}
	if stats.TriangleCount != 12 {
		t.Errorf("TriangleCount = %d, want 12", stats.TriangleCount)
	}

	wantVol := float64(s * s * s)
	if math.Abs(stats.Volume-wantVol) > 1e-4 {
		t.Errorf("Volume = %v, want %v", stats.Volume, wantVol)
	}

	wantArea := float64(6 * s * s)
	if math.Abs(stats.SurfaceArea-wantArea) > 1e-4 {
		t.Errorf("SurfaceArea = %v, want %v", stats.SurfaceArea, wantArea)
	}

	wantBbox := [6]float32{1, 2, 3, 1 + s, 2 + s, 3 + s}
	for i := range wantBbox {
		if math.Abs(float64(stats.Bbox[i]-wantBbox[i])) > 1e-4 {
			t.Errorf("Bbox[%d] = %v, want %v", i, stats.Bbox[i], wantBbox[i])
		}
	}

	wantCenter := [3]float32{1 + s/2, 2 + s/2, 3 + s/2}
	for i := range wantCenter {
		if math.Abs(float64(stats.Center[i]-wantCenter[i])) > 1e-4 {
			t.Errorf("Center[%d] = %v, want %v", i, stats.Center[i], wantCenter[i])
		}
	}
}

func TestMeasureSTLEmpty(t *testing.T) {
	var buf bytes.Buffer
	buf.Write(make([]byte, 80))
	binary.Write(&buf, binary.LittleEndian, uint32(0))
	stats, err := MeasureSTL(buf.Bytes())
	if err != nil {
		t.Fatalf("MeasureSTL: %v", err)
	}
	if !stats.Empty() {
		t.Errorf("expected Empty() = true for 0-triangle STL")
	}
}

func TestMeasureSTLTruncated(t *testing.T) {
	if _, err := MeasureSTL([]byte{0, 1, 2}); err == nil {
		t.Errorf("expected error on truncated header")
	}
	var buf bytes.Buffer
	buf.Write(make([]byte, 80))
	binary.Write(&buf, binary.LittleEndian, uint32(5))
	buf.Write(make([]byte, 10)) // fewer than 5 * 50 bytes
	if _, err := MeasureSTL(buf.Bytes()); err == nil {
		t.Errorf("expected error on truncated triangle stream")
	}
}
