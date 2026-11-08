// Thin C-ABI wrapper around manifold::Manifold for the Go package. We expose
// only the four operations the export pipeline actually uses: load a mesh,
// transform it, boolean it against another, and serialize to binary STL.
//
// Memory contract — read this before touching anything in this file:
//
//   - Handles are heap-allocated `manifold::Manifold` instances. Every
//     entry point that returns a non-null ManifoldHandle* hands ownership
//     to the caller; the caller MUST eventually pair each with exactly
//     one mglue_release. Returning nullptr means we already cleaned up
//     internally — no further free is required (or possible).
//
//   - mglue_to_stl writes to a malloc'd buffer; the caller pairs it with
//     exactly one mglue_free_buffer. On failure the function leaves
//     *out_ptr / *out_len untouched and returns negative.
//
//   - Every `extern "C"` entry point catches all exceptions and converts
//     them to nullptr / negative-rc returns. C++ exceptions MUST NOT
//     escape into Rust/Go — that would be UB. The destructors of
//     stack-local objects (MeshGL, etc.) run during stack unwinding so
//     no allocations are leaked when an exception is caught.
//
//   - All size_t arithmetic is bounds-checked before use. A malformed
//     input that would overflow `nverts * 3`, `ntris * 3`, or the STL
//     byte count is rejected up-front rather than allocated against.

#include "manifold_glue.h"

#include <algorithm>
#include <cmath>
#include <cstdlib>
#include <cstring>
#include <limits>
#include <new>
#include <stdexcept>
#include <vector>

#include "manifold/manifold.h"

using manifold::Manifold;
using manifold::MeshGL;
using manifold::OpType;

namespace {

// Hard ceilings to refuse pathological inputs before they hit the
// allocator. A billion triangles is already ~50 GB of binary STL and
// ~150 GB of MeshGL state — well past anything we'd want to compute on
// a server even in the worst-case real workload. These ceilings exist
// only to prevent malformed payloads from crashing the process or
// triggering OOM-killer.
constexpr size_t kMaxVerts = static_cast<size_t>(1) << 30;  // 1 G verts
constexpr size_t kMaxTris  = static_cast<size_t>(1) << 30;  // 1 G tris

// Returns true and stores the product in `out` if a*b doesn't overflow
// size_t; returns false otherwise. Used to bounds-check buffer sizes
// before allocating.
inline bool mul_no_overflow(size_t a, size_t b, size_t* out) {
  if (a == 0 || b == 0) {
    *out = 0;
    return true;
  }
  if (a > std::numeric_limits<size_t>::max() / b) {
    return false;
  }
  *out = a * b;
  return true;
}

inline bool add_no_overflow(size_t a, size_t b, size_t* out) {
  if (a > std::numeric_limits<size_t>::max() - b) {
    return false;
  }
  *out = a + b;
  return true;
}

// Treat the opaque ManifoldHandle as an alias for manifold::Manifold under
// the hood. This avoids an extra allocation on every op while keeping the
// public C ABI opaque.
inline Manifold* as_inner(ManifoldHandle* h) {
  return reinterpret_cast<Manifold*>(h);
}
inline const Manifold* as_inner(const ManifoldHandle* h) {
  return reinterpret_cast<const Manifold*>(h);
}
inline ManifoldHandle* as_handle(Manifold* m) {
  return reinterpret_cast<ManifoldHandle*>(m);
}

OpType to_op_type(int op) {
  switch (op) {
    case MANIFOLD_OP_UNION:
      return OpType::Add;
    case MANIFOLD_OP_DIFFERENCE:
      return OpType::Subtract;
    case MANIFOLD_OP_INTERSECTION:
      return OpType::Intersect;
    default:
      // Caller passed a junk op code; default to union so the result is at
      // least geometrically defined. The Go side validates before getting
      // here, but defense-in-depth costs nothing.
      return OpType::Add;
  }
}

// Compute the unit normal of a triangle given three vertices. Returns the
// zero vector for degenerate (collinear) triangles — STL readers tolerate
// zero-length normals (they recompute from vertex order), so this is a
// reasonable fallback that avoids NaN propagation.
void tri_normal(const float* a, const float* b, const float* c, float out[3]) {
  float ux = b[0] - a[0], uy = b[1] - a[1], uz = b[2] - a[2];
  float vx = c[0] - a[0], vy = c[1] - a[1], vz = c[2] - a[2];
  float nx = uy * vz - uz * vy;
  float ny = uz * vx - ux * vz;
  float nz = ux * vy - uy * vx;
  float len = std::sqrt(nx * nx + ny * ny + nz * nz);
  if (len > 0.0f) {
    out[0] = nx / len;
    out[1] = ny / len;
    out[2] = nz / len;
  } else {
    out[0] = 0.0f;
    out[1] = 0.0f;
    out[2] = 0.0f;
  }
}

}  // namespace

extern "C" {

ManifoldHandle* mglue_load_meshgl(const float* verts, size_t nverts,
                                  const uint32_t* tris, size_t ntris) {
  if (verts == nullptr || tris == nullptr || nverts == 0 || ntris == 0) {
    return nullptr;
  }
  if (nverts > kMaxVerts || ntris > kMaxTris) {
    return nullptr;
  }
  // Bounds-check the (verts + nverts*3) and (tris + ntris*3) arithmetic
  // before we hand it to std::vector::assign — otherwise an overflow
  // would yield a wrapped pointer and UB inside libc++.
  size_t vert_floats = 0;
  size_t tri_indices = 0;
  if (!mul_no_overflow(nverts, 3, &vert_floats) ||
      !mul_no_overflow(ntris, 3, &tri_indices)) {
    return nullptr;
  }
  // The Manifold heap allocation lives across the try/catch. If `new`
  // succeeds and we later decide to discard, `delete` reclaims it on
  // every code path below. If `new` itself throws (constructor failure),
  // C++'s exception machinery deallocates the raw memory before the
  // catch block runs — no leak.
  Manifold* m = nullptr;
  try {
    MeshGL mesh;
    mesh.numProp = 3;
    // assign() may throw bad_alloc on huge inputs; the MeshGL stack
    // local destructs and frees any partial vector storage during
    // unwinding.
    mesh.vertProperties.assign(verts, verts + vert_floats);
    mesh.triVerts.assign(tris, tris + tri_indices);
    // Recover topology if the input came from a triangle soup with
    // duplicated corner vertices (the STL upload path always does — every
    // STL face stores its three corners independently).
    mesh.Merge();
    m = new Manifold(mesh);
    if (m->Status() != Manifold::Error::NoError) {
      delete m;
      return nullptr;
    }
    return as_handle(m);
  } catch (const std::exception&) {
    delete m;
    return nullptr;
  } catch (...) {
    delete m;
    return nullptr;
  }
}

ManifoldHandle* mglue_boolean(const ManifoldHandle* a,
                              const ManifoldHandle* b, int op) {
  if (a == nullptr || b == nullptr) {
    return nullptr;
  }
  // Pre-declare so the catch block can clean up if `new` succeeded but a
  // later statement threw. If Boolean(...) itself throws BEFORE `new`
  // runs, `result` stays nullptr and `delete nullptr` is a no-op.
  Manifold* result = nullptr;
  try {
    // Boolean returns a Manifold by value; the rvalue is moved into the
    // heap allocation. Manifold's CSG tree is lazy — no surface mesh is
    // computed here, just a tree node — so this is cheap and uses bounded
    // memory regardless of input size.
    result = new Manifold(as_inner(a)->Boolean(*as_inner(b), to_op_type(op)));
    if (result->Status() != Manifold::Error::NoError) {
      delete result;
      return nullptr;
    }
    return as_handle(result);
  } catch (const std::exception&) {
    delete result;
    return nullptr;
  } catch (...) {
    delete result;
    return nullptr;
  }
}

ManifoldHandle* mglue_transform(const ManifoldHandle* h,
                                const float* mat3x4) {
  if (h == nullptr || mat3x4 == nullptr) {
    return nullptr;
  }
  Manifold* result = nullptr;
  try {
    // mat3x4 is column-major: 4 columns × 3 rows = 12 floats. Promote to
    // doubles to match manifold's mat3x4 = la::mat<double, 3, 4>.
    manifold::mat3x4 m{
        manifold::vec3{mat3x4[0], mat3x4[1], mat3x4[2]},
        manifold::vec3{mat3x4[3], mat3x4[4], mat3x4[5]},
        manifold::vec3{mat3x4[6], mat3x4[7], mat3x4[8]},
        manifold::vec3{mat3x4[9], mat3x4[10], mat3x4[11]},
    };
    result = new Manifold(as_inner(h)->Transform(m));
    if (result->Status() != Manifold::Error::NoError) {
      delete result;
      return nullptr;
    }
    return as_handle(result);
  } catch (const std::exception&) {
    delete result;
    return nullptr;
  } catch (...) {
    delete result;
    return nullptr;
  }
}

int mglue_to_stl(const ManifoldHandle* h, uint8_t** out_ptr,
                 size_t* out_len) {
  if (h == nullptr || out_ptr == nullptr || out_len == nullptr) {
    return -1;
  }
  // Allocated in the try block; freed in the catch block (or on
  // unsuccessful return paths inside the try) so an exception thrown
  // mid-serialization can't strand the buffer.
  uint8_t* buf = nullptr;
  try {
    // GetMeshGL returns by value; its internal vector storage is owned
    // by `mesh` and freed when this scope exits.
    MeshGL mesh = as_inner(h)->GetMeshGL();
    const size_t ntris = mesh.NumTri();
    // Binary STL spec stores the triangle count as uint32. Refuse to
    // emit an STL whose advertised count would silently truncate — the
    // file would be unreadable rather than just bigger.
    if (ntris > std::numeric_limits<uint32_t>::max()) {
      return -4;
    }
    if (ntris > kMaxTris) {
      return -4;
    }
    // Binary STL: 80 byte header + uint32 tri count + per-tri 50 bytes
    // (12 normal + 36 verts + 2 attr). All three multiplications are
    // bounds-checked because a malformed Manifold could in principle
    // expose a triangle count near SIZE_MAX/50 and crash the process on
    // a wrapped allocation.
    size_t per_tri_total = 0;
    if (!mul_no_overflow(ntris, 50, &per_tri_total)) {
      return -4;
    }
    size_t bytes = 0;
    if (!add_no_overflow(per_tri_total, 84, &bytes)) {
      return -4;
    }
    buf = static_cast<uint8_t*>(std::malloc(bytes));
    if (buf == nullptr) {
      return -2;
    }
    std::memset(buf, 0, 80);
    uint32_t tri_count = static_cast<uint32_t>(ntris);
    std::memcpy(buf + 80, &tri_count, 4);
    uint8_t* w = buf + 84;
    const float* verts = mesh.vertProperties.data();
    const auto* idx = mesh.triVerts.data();
    // None of the operations in the loop body throw — std::memcpy and
    // arithmetic on POD floats are noexcept. So once we're past the
    // malloc the buffer can't be stranded by an exception.
    for (size_t t = 0; t < ntris; ++t) {
      const float* a = verts + 3 * idx[3 * t + 0];
      const float* b = verts + 3 * idx[3 * t + 1];
      const float* c = verts + 3 * idx[3 * t + 2];
      float n[3];
      tri_normal(a, b, c, n);
      std::memcpy(w, n, 12);
      std::memcpy(w + 12, a, 12);
      std::memcpy(w + 24, b, 12);
      std::memcpy(w + 36, c, 12);
      w[48] = 0;
      w[49] = 0;
      w += 50;
    }
    *out_ptr = buf;
    *out_len = bytes;
    // Ownership has transferred to the caller — clear our local pointer
    // so the catch block can't double-free if a later throw-able
    // statement is added above the return.
    buf = nullptr;
    return 0;
  } catch (const std::exception&) {
    std::free(buf);
    return -3;
  } catch (...) {
    std::free(buf);
    return -3;
  }
}

void mglue_release(ManifoldHandle* h) {
  if (h == nullptr) {
    return;
  }
  delete as_inner(h);
}

void mglue_free_buffer(uint8_t* ptr, size_t /*len*/) {
  std::free(ptr);
}

}  // extern "C"
