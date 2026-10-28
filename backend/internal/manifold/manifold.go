// Package manifold is the cgo bridge to the manifold_glue Rust dylib
// (which itself wraps the vendored manifold C++ library). It exposes a
// safe-ish Go API: load a mesh, transform it, boolean it against another,
// serialize to binary STL bytes.
//
// Memory contract — read this before calling anything in this package:
//
//   - A non-zero Handle returned from Load / Boolean / Transform owns a
//     C++ allocation. The caller MUST eventually call Release on it
//     exactly once. Forgetting to release a handle leaks ~the entire
//     Manifold (vertices, triangles, CSG tree state) until the agent
//     process exits.
//
//   - Handles are pointer-sized values. Copying a Handle (e.g. passing
//     by value) creates a second alias to the SAME C++ allocation;
//     calling Release on both copies is a double-free. The compose-tree
//     walker below threads each handle through a single owner so this
//     can't happen there, but if you add new call sites, audit them.
//
//   - Boolean and Transform never consume their inputs — they return a
//     fresh handle. The caller still owns and must release every input
//     handle in addition to the result.
//
//   - ToSTL allocates a Go-side []byte by copying out of a malloc'd
//     buffer that this package frees before returning. The caller
//     receives a normal Go slice with no manual lifetime management.
//
// The cgo header mirrors the public Rust ABI in
// agent/manifold/internal/rust/src/lib.rs. Keep the two in sync; rebuild
// the dylib when the Rust signatures change.
package manifold

/*
#cgo LDFLAGS: -L../../../../target/release -lmanifold_glue
#include <stdint.h>
#include <stdlib.h>

typedef struct ManifoldHandle ManifoldHandle;

ManifoldHandle* manifold_load_meshgl(const float* verts, size_t nverts,
                                     const uint32_t* tris, size_t ntris);
ManifoldHandle* manifold_boolean(const ManifoldHandle* a,
                                 const ManifoldHandle* b, int op);
ManifoldHandle* manifold_transform(const ManifoldHandle* h,
                                   const float* mat3x4);
int             manifold_to_stl(const ManifoldHandle* h, uint8_t** out_ptr,
                                size_t* out_len);
void            manifold_release(ManifoldHandle* h);
void            manifold_free_buffer(uint8_t* ptr, size_t len);
*/
import "C"
import (
	"errors"
	"fmt"
	"unsafe"
)

// Op selects the boolean operation Compose runs at an internal plan
// node. Values match the C ABI's MANIFOLD_OP_* constants — keep them
// in sync.
type Op int

const (
	OpUnion        Op = 0
	OpDifference   Op = 1
	OpIntersection Op = 2
)

// Handle is an opaque pointer to a heap-allocated manifold::Manifold on
// the C++ side. The zero-value Handle is invalid; allocate via Load and
// release via Handle.Release. Compose ops return new Handles — they do
// not consume their inputs.
type Handle struct {
	ptr *C.ManifoldHandle
}

// Valid reports whether the handle points at an allocated manifold.
// Defensive guard for callers that may have already released the handle
// or hit a NULL-returning op upstream.
func (h Handle) Valid() bool { return h.ptr != nil }

// Release frees the underlying C++ allocation. Safe to call on a
// zero-value or already-released handle (the C++ side no-ops on NULL).
// After Release the handle should not be used; subsequent calls into the
// FFI would fail with NULL inputs and propagate as errors.
func (h *Handle) Release() {
	if h == nil || h.ptr == nil {
		return
	}
	C.manifold_release(h.ptr)
	h.ptr = nil
}

// Load builds a manifold from a flat triangle soup. `vertices` is
// xyz triplets (length divisible by 3); `triangles` is uvec3 triplets of
// indices into `vertices` (length divisible by 3). Returns an error if
// the input is malformed or the C++ Manifold constructor rejects the
// mesh as non-recoverable (non-manifold past the merge tolerance).
func Load(vertices []float32, triangles []uint32) (Handle, error) {
	if len(vertices)%3 != 0 {
		return Handle{}, fmt.Errorf("manifold.Load: vertices length %d not divisible by 3", len(vertices))
	}
	if len(triangles)%3 != 0 {
		return Handle{}, fmt.Errorf("manifold.Load: triangles length %d not divisible by 3", len(triangles))
	}
	nverts := len(vertices) / 3
	ntris := len(triangles) / 3
	if nverts == 0 || ntris == 0 {
		return Handle{}, fmt.Errorf("manifold.Load: empty mesh (verts=%d tris=%d)", nverts, ntris)
	}
	ptr := C.manifold_load_meshgl(
		(*C.float)(unsafe.Pointer(&vertices[0])),
		C.size_t(nverts),
		(*C.uint32_t)(unsafe.Pointer(&triangles[0])),
		C.size_t(ntris),
	)
	if ptr == nil {
		return Handle{}, errors.New("manifold.Load: rust returned NULL (mesh non-manifold or degenerate)")
	}
	return Handle{ptr: ptr}, nil
}

// Boolean returns a fresh handle holding (a OP b). Inputs are not
// consumed; callers retain ownership of `a` and `b` and must release
// them separately (in addition to releasing the result). Returns an
// error if either input is invalid or the C++ op rejects the
// configuration (e.g. the result would be non-manifold).
func Boolean(a, b Handle, op Op) (Handle, error) {
	if !a.Valid() || !b.Valid() {
		return Handle{}, errors.New("manifold.Boolean: invalid input handle")
	}
	ptr := C.manifold_boolean(a.ptr, b.ptr, C.int(op))
	if ptr == nil {
		return Handle{}, errors.New("manifold.Boolean: rust returned NULL (op failed)")
	}
	return Handle{ptr: ptr}, nil
}

// Transform applies a 3x4 column-major affine transform and returns a
// new handle. The input is not consumed. `mat3x4` must be exactly 12
// floats laid out column-major: cols 0..2 are the rotation/scale basis
// and col 3 is the translation.
func Transform(h Handle, mat3x4 [12]float32) (Handle, error) {
	if !h.Valid() {
		return Handle{}, errors.New("manifold.Transform: invalid input handle")
	}
	ptr := C.manifold_transform(h.ptr, (*C.float)(unsafe.Pointer(&mat3x4[0])))
	if ptr == nil {
		return Handle{}, errors.New("manifold.Transform: rust returned NULL (op failed)")
	}
	return Handle{ptr: ptr}, nil
}

// ToSTL serializes the manifold to binary STL bytes. The returned
// slice is a heap-copy of the C-allocated buffer; this function frees
// the C buffer before returning, so callers don't have to track
// lifetimes.
//
// The buffer is copied via unsafe.Slice rather than C.GoBytes because
// C.GoBytes takes a `C.int` length, which truncates on STLs larger
// than ~2 GB. unsafe.Slice uses Go's `int` (64-bit on 64-bit hosts),
// so the copy stays correct up to platform memory limits.
func ToSTL(h Handle) ([]byte, error) {
	if !h.Valid() {
		return nil, errors.New("manifold.ToSTL: invalid handle")
	}
	var outPtr *C.uint8_t
	var outLen C.size_t
	rc := C.manifold_to_stl(h.ptr, &outPtr, &outLen)
	if rc != 0 {
		return nil, fmt.Errorf("manifold.ToSTL: rust returned %d", int(rc))
	}
	// Defensive: free whatever the C side returned, even if outLen is
	// zero. malloc(0) is implementation-defined and may return a real
	// allocation we'd otherwise leak. C.manifold_free_buffer is a
	// no-op on a NULL pointer.
	defer C.manifold_free_buffer(outPtr, outLen)
	if outPtr == nil {
		return nil, errors.New("manifold.ToSTL: rust returned NULL buffer")
	}
	if outLen == 0 {
		// Empty STL is technically valid (header + 0 tri count). Return
		// an empty Go slice rather than nil so the caller can
		// distinguish "no data" from "error".
		return []byte{}, nil
	}
	src := unsafe.Slice((*byte)(unsafe.Pointer(outPtr)), int(outLen))
	out := make([]byte, len(src))
	copy(out, src)
	return out, nil
}
