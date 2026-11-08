#ifndef MANIFOLD_GLUE_H
#define MANIFOLD_GLUE_H

#include <stddef.h>
#include <stdint.h>

#ifdef __cplusplus
extern "C" {
#endif

// Internal C ABI exposed by the static glue. Rust shims (see src/lib.rs)
// re-export these under their public `manifold_*` names so the dylib has
// a clean Rust-controlled export boundary, mirroring how the fidget package
// re-exports its OpenVDB glue.
//
// The opaque handle is a heap-allocated manifold::Manifold. NULL is the
// sentinel for "operation failed" — the Rust side and the Go package both
// translate NULL into a typed error.
typedef struct ManifoldHandle ManifoldHandle;

#define MANIFOLD_OP_UNION 0
#define MANIFOLD_OP_DIFFERENCE 1
#define MANIFOLD_OP_INTERSECTION 2

ManifoldHandle* mglue_load_meshgl(const float* verts, size_t nverts,
                                  const uint32_t* tris, size_t ntris);
ManifoldHandle* mglue_boolean(const ManifoldHandle* a, const ManifoldHandle* b,
                              int op);
ManifoldHandle* mglue_transform(const ManifoldHandle* h, const float* mat3x4);
int             mglue_to_stl(const ManifoldHandle* h, uint8_t** out_ptr,
                             size_t* out_len);
void            mglue_release(ManifoldHandle* h);
void            mglue_free_buffer(uint8_t* ptr, size_t len);

#ifdef __cplusplus
}  // extern "C"
#endif

#endif  // MANIFOLD_GLUE_H

