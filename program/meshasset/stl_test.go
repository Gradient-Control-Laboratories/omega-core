package meshasset

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"
)

// buildSTL synthesizes a minimal binary STL blob for tests.
func buildSTL(tris []Triangle) []byte {
	var buf bytes.Buffer
	buf.Write(make([]byte, 80)) // header
	_ = binary.Write(&buf, binary.LittleEndian, uint32(len(tris)))
	for _, t := range tris {
		// normal = 0,0,0
		_ = binary.Write(&buf, binary.LittleEndian, float32(0))
		_ = binary.Write(&buf, binary.LittleEndian, float32(0))
		_ = binary.Write(&buf, binary.LittleEndian, float32(0))
		for _, v := range [...][3]float64{t.V0, t.V1, t.V2} {
			_ = binary.Write(&buf, binary.LittleEndian, float32(v[0]))
			_ = binary.Write(&buf, binary.LittleEndian, float32(v[1]))
			_ = binary.Write(&buf, binary.LittleEndian, float32(v[2]))
		}
		_ = binary.Write(&buf, binary.LittleEndian, uint16(0))
	}
	return buf.Bytes()
}

func TestParseBinarySTLBasic(t *testing.T) {
	tris := []Triangle{
		{V0: [3]float64{0, 0, 0}, V1: [3]float64{1, 0, 0}, V2: [3]float64{0, 1, 0}},
		{V0: [3]float64{1, 1, 2}, V1: [3]float64{-1, 0, 0}, V2: [3]float64{2, -1, 3}},
	}
	data := buildSTL(tris)
	mesh, hash, err := ParseBinarySTL(data)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(mesh.Triangles) != 2 {
		t.Fatalf("want 2 triangles, got %d", len(mesh.Triangles))
	}
	if hash == "" || len(hash) != 64 {
		t.Fatalf("want 64-char hex hash, got %q", hash)
	}

	// bbox should span the vertex extents
	wantMin := [3]float64{-1, -1, 0}
	wantMax := [3]float64{2, 1, 3}
	for i := 0; i < 3; i++ {
		if math.Abs(mesh.Bbox.Min[i]-wantMin[i]) > 1e-5 {
			t.Errorf("bbox.Min[%d] = %v, want %v", i, mesh.Bbox.Min[i], wantMin[i])
		}
		if math.Abs(mesh.Bbox.Max[i]-wantMax[i]) > 1e-5 {
			t.Errorf("bbox.Max[%d] = %v, want %v", i, mesh.Bbox.Max[i], wantMax[i])
		}
	}
	if !mesh.Bbox.IsValid() {
		t.Errorf("expected valid bbox")
	}
}

func TestMarshalBinarySTLRoundTrip(t *testing.T) {
	tris := []Triangle{
		{V0: [3]float64{0, 0, 0}, V1: [3]float64{1, 0, 0}, V2: [3]float64{0, 1, 0}},
	}
	data, err := MarshalBinarySTL(&TriMesh{Triangles: tris})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	mesh, _, err := ParseSTL(data)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(mesh.Triangles) != len(tris) || mesh.Triangles[0] != tris[0] {
		t.Fatalf("round trip mismatch: %#v", mesh.Triangles)
	}
}

func TestParseBinarySTLRejectsShort(t *testing.T) {
	if _, _, err := ParseBinarySTL([]byte("short")); err == nil {
		t.Fatalf("expected error on short input")
	}
}

func TestParseBinarySTLRejectsTruncated(t *testing.T) {
	data := buildSTL([]Triangle{
		{V0: [3]float64{0, 0, 0}, V1: [3]float64{1, 0, 0}, V2: [3]float64{0, 1, 0}},
	})
	// Chop off the last triangle's bytes.
	if _, _, err := ParseBinarySTL(data[:len(data)-10]); err == nil {
		t.Fatalf("expected error on truncated input")
	}
}

func TestParseBinarySTLRejectsEmpty(t *testing.T) {
	data := buildSTL(nil)
	if _, _, err := ParseBinarySTL(data); err == nil {
		t.Fatalf("expected error on zero-triangle input")
	}
}

func TestHashStableAcrossCalls(t *testing.T) {
	data := buildSTL([]Triangle{
		{V0: [3]float64{0, 0, 0}, V1: [3]float64{1, 0, 0}, V2: [3]float64{0, 1, 0}},
	})
	h1 := HashBytes(data)
	h2 := HashBytes(data)
	if h1 != h2 {
		t.Fatalf("hash should be stable: %q vs %q", h1, h2)
	}
}

func TestBboxPadded(t *testing.T) {
	b := Bbox{Min: [3]float64{0, 0, 0}, Max: [3]float64{10, 10, 10}}
	p := b.Padded(0.2) // 20% on each side = +2 each side
	want := Bbox{Min: [3]float64{-2, -2, -2}, Max: [3]float64{12, 12, 12}}
	if p != want {
		t.Fatalf("padded = %+v, want %+v", p, want)
	}
}

func TestParseASCIISTLBasic(t *testing.T) {
	src := `solid test
  facet normal 0 0 0
    outer loop
      vertex 0 0 0
      vertex 1 0 0
      vertex 0 1 0
    endloop
  endfacet
  facet normal 0 0 0
    outer loop
      vertex 2 2 3
      vertex -1 0 0
      vertex 1 -1 2
    endloop
  endfacet
endsolid test
`
	mesh, hash, err := ParseSTL([]byte(src))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(mesh.Triangles) != 2 {
		t.Fatalf("want 2 triangles, got %d", len(mesh.Triangles))
	}
	if hash == "" || len(hash) != 64 {
		t.Fatalf("want 64-char hex hash, got %q", hash)
	}
	wantMin := [3]float64{-1, -1, 0}
	wantMax := [3]float64{2, 2, 3}
	for i := 0; i < 3; i++ {
		if math.Abs(mesh.Bbox.Min[i]-wantMin[i]) > 1e-9 {
			t.Errorf("bbox.Min[%d] = %v, want %v", i, mesh.Bbox.Min[i], wantMin[i])
		}
		if math.Abs(mesh.Bbox.Max[i]-wantMax[i]) > 1e-9 {
			t.Errorf("bbox.Max[%d] = %v, want %v", i, mesh.Bbox.Max[i], wantMax[i])
		}
	}
}

func TestParseSTLAutoDetectsBinary(t *testing.T) {
	tris := []Triangle{{V0: [3]float64{0, 0, 0}, V1: [3]float64{1, 0, 0}, V2: [3]float64{0, 1, 0}}}
	data := buildSTL(tris)
	mesh, _, err := ParseSTL(data)
	if err != nil {
		t.Fatalf("ParseSTL on binary input: %v", err)
	}
	if len(mesh.Triangles) != 1 {
		t.Fatalf("want 1 triangle, got %d", len(mesh.Triangles))
	}
}

func TestParseASCIISTLCaseAndWhitespace(t *testing.T) {
	src := `SOLID mixed
  FACET NORMAL  0.0  0.0  1.0
     OUTER LOOP
        vertex   0.5    0.5    0.0
        VERTEX   1.5    0.5    0.0
        vertex   0.5    1.5    0.0
     ENDLOOP
  ENDFACET
ENDSOLID`
	mesh, _, err := ParseSTL([]byte(src))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	if len(mesh.Triangles) != 1 {
		t.Fatalf("want 1 triangle, got %d", len(mesh.Triangles))
	}
}
