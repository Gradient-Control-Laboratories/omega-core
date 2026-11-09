package manifold

import (
	"bytes"
	"encoding/binary"
	"runtime"
	"testing"
)

// unitCubeMesh returns a triangle-soup unit cube centered on the origin
// with side length 2 (i.e. corners at ±1). Each face is two triangles
// with three independent vertices, so 36 floats × 3 axes = 36 verts and
// 12 triangles. Vertex indices are sequential 0..35; manifold's
// MeshGL::Merge() recovers shared topology before the Manifold
// constructor builds its CSG state.
func unitCubeMesh() ([]float32, []uint32) {
	// 6 faces × 2 tris × 3 verts = 36 verts. xyz stored flat.
	v := []float32{
		// -X face (normal -X)
		-1, -1, -1, -1, 1, -1, -1, 1, 1,
		-1, -1, -1, -1, 1, 1, -1, -1, 1,
		// +X face
		1, -1, -1, 1, -1, 1, 1, 1, 1,
		1, -1, -1, 1, 1, 1, 1, 1, -1,
		// -Y face
		-1, -1, -1, -1, -1, 1, 1, -1, 1,
		-1, -1, -1, 1, -1, 1, 1, -1, -1,
		// +Y face
		-1, 1, -1, 1, 1, -1, 1, 1, 1,
		-1, 1, -1, 1, 1, 1, -1, 1, 1,
		// -Z face
		-1, -1, -1, 1, -1, -1, 1, 1, -1,
		-1, -1, -1, 1, 1, -1, -1, 1, -1,
		// +Z face
		-1, -1, 1, -1, 1, 1, 1, 1, 1,
		-1, -1, 1, 1, 1, 1, 1, -1, 1,
	}
	t := make([]uint32, 36)
	for i := range t {
		t[i] = uint32(i)
	}
	return v, t
}

func TestLoadReleaseRoundtrip(t *testing.T) {
	v, tris := unitCubeMesh()
	h, err := Load(v, tris)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if !h.Valid() {
		t.Fatal("Load returned zero handle without error")
	}
	h.Release()
	if h.Valid() {
		t.Fatal("Release did not clear handle")
	}
	// Double-release must be a safe no-op.
	h.Release()
}

func TestLoadRejectsInvalid(t *testing.T) {
	cases := []struct {
		name  string
		verts []float32
		tris  []uint32
	}{
		{"empty", nil, nil},
		{"verts non-multiple of 3", []float32{1, 2, 3, 4}, []uint32{0, 1, 2}},
		{"tris non-multiple of 3", []float32{0, 0, 0, 1, 0, 0, 0, 1, 0}, []uint32{0, 1}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h, err := Load(tc.verts, tc.tris)
			if err == nil {
				h.Release()
				t.Fatal("expected error, got nil")
			}
			if h.Valid() {
				h.Release()
				t.Fatal("expected zero handle on error")
			}
		})
	}
}

func TestBooleanReleasesIndependently(t *testing.T) {
	v, tris := unitCubeMesh()
	a, err := Load(v, tris)
	if err != nil {
		t.Fatalf("Load a: %v", err)
	}
	defer a.Release()
	b, err := Load(v, tris)
	if err != nil {
		t.Fatalf("Load b: %v", err)
	}
	defer b.Release()

	out, err := Boolean(a, b, OpUnion)
	if err != nil {
		t.Fatalf("Boolean: %v", err)
	}
	if !out.Valid() {
		t.Fatal("Boolean returned zero handle")
	}
	defer out.Release()

	// Inputs must remain valid — Boolean does not consume them. We
	// confirm by reading them again (ToSTL inspects internal state).
	if _, err := ToSTL(a); err != nil {
		t.Fatalf("input a corrupted after Boolean: %v", err)
	}
	if _, err := ToSTL(b); err != nil {
		t.Fatalf("input b corrupted after Boolean: %v", err)
	}
}

func TestTransformReleasesIndependently(t *testing.T) {
	v, tris := unitCubeMesh()
	h, err := Load(v, tris)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	defer h.Release()

	// Identity-ish 3x4 column-major: cols 0..2 are basis (sX=2, sY=2,
	// sZ=2 to make it observable), col 3 is translation.
	mat := [12]float32{
		2, 0, 0, // col 0: scaled X axis
		0, 2, 0, // col 1: scaled Y axis
		0, 0, 2, // col 2: scaled Z axis
		1, 2, 3, // col 3: translation
	}
	out, err := Transform(h, mat)
	if err != nil {
		t.Fatalf("Transform: %v", err)
	}
	if !out.Valid() {
		t.Fatal("Transform returned zero handle")
	}
	defer out.Release()

	// Source must remain valid.
	if _, err := ToSTL(h); err != nil {
		t.Fatalf("input corrupted after Transform: %v", err)
	}
}

func TestToSTLProducesValidBinary(t *testing.T) {
	v, tris := unitCubeMesh()
	h, err := Load(v, tris)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	defer h.Release()

	stl, err := ToSTL(h)
	if err != nil {
		t.Fatalf("ToSTL: %v", err)
	}
	if len(stl) < 84 {
		t.Fatalf("STL too short: %d bytes", len(stl))
	}
	// Binary STL: 80-byte header, then uint32 tri count.
	count := binary.LittleEndian.Uint32(stl[80:84])
	if count == 0 {
		t.Fatal("STL reports zero triangles")
	}
	expected := 84 + int(count)*50
	if len(stl) != expected {
		t.Fatalf("STL length %d does not match header (84 + %d * 50 = %d)",
			len(stl), count, expected)
	}
	// A unit cube ought to produce 12 triangles after Merge() recovers
	// shared verts. Some manifold builds may add internal subdivisions
	// during construction; a slightly higher count is acceptable as
	// long as it's bounded and consistent.
	if count > 64 {
		t.Fatalf("unit cube STL has %d triangles, expected ~12", count)
	}
	// Header is mostly zero from our memset.
	if !bytes.Equal(stl[:80], make([]byte, 80)) {
		t.Errorf("STL header not zeroed")
	}
}

// TestStressLoopForLeaks runs many full Load/Boolean/Transform/ToSTL
// cycles back-to-back. Run under a leak detector (`leaks`, valgrind,
// macOS Instruments) to verify the FFI doesn't leak across iterations:
//
//	go test -c -o manifold.test ./internal/manifold
//	MallocStackLogging=1 leaks --atExit -- ./manifold.test \
//	    -test.run TestStressLoopForLeaks
//
// As of writing, macOS produces exactly one fixed 32-byte leak
// (xpc_date_t from libxpc) regardless of iteration count — that's an
// Apple framework artifact, not ours. Anything per-iteration shows up
// as N copies of the same allocation at the same call site, which is
// what we're guarding against here.
//
// The test itself only checks that the iterations don't return an
// error; the leak detector validates allocation balance. CI can
// optionally watch RSS via runtime.MemStats but that only catches
// Go-side leaks.
func TestStressLoopForLeaks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping stress loop in -short")
	}
	const iterations = 1000
	v, tris := unitCubeMesh()
	mat := [12]float32{1, 0, 0, 0, 1, 0, 0, 0, 1, 0, 0, 0}

	// Capture Go heap stats before and after to spot Go-side regressions.
	// C++ heap leaks must be observed externally.
	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)

	for i := 0; i < iterations; i++ {
		a, err := Load(v, tris)
		if err != nil {
			t.Fatalf("iter %d Load a: %v", i, err)
		}
		b, err := Load(v, tris)
		if err != nil {
			a.Release()
			t.Fatalf("iter %d Load b: %v", i, err)
		}
		out, err := Boolean(a, b, OpUnion)
		if err != nil {
			a.Release()
			b.Release()
			t.Fatalf("iter %d Boolean: %v", i, err)
		}
		transformed, err := Transform(out, mat)
		if err != nil {
			a.Release()
			b.Release()
			out.Release()
			t.Fatalf("iter %d Transform: %v", i, err)
		}
		stl, err := ToSTL(transformed)
		if err != nil {
			a.Release()
			b.Release()
			out.Release()
			transformed.Release()
			t.Fatalf("iter %d ToSTL: %v", i, err)
		}
		if len(stl) < 84 {
			t.Fatalf("iter %d: STL too short: %d", i, len(stl))
		}
		a.Release()
		b.Release()
		out.Release()
		transformed.Release()
	}

	runtime.GC()
	runtime.ReadMemStats(&after)
	// 8 MiB growth across 200 iterations on the Go heap signals a
	// regression in the cgo wrapper's slice handling. The C++ heap is
	// not visible here.
	const goHeapTolerance uint64 = 8 << 20
	if after.HeapAlloc > before.HeapAlloc+goHeapTolerance {
		t.Errorf("Go heap grew by %d bytes across %d iterations (tolerance %d)",
			after.HeapAlloc-before.HeapAlloc, iterations, goHeapTolerance)
	}
}
