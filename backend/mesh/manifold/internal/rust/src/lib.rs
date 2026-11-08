//! Public C ABI for the manifold package's dylib. This file is the export
//! boundary that the Go package calls into via cgo.
//!
//! All of the heavy lifting happens in `src/cpp/manifold_glue.cpp`, which
//! exposes a flat `mglue_*` C API around `manifold::Manifold`. The Rust
//! shims below re-export each `mglue_*` function under a stable
//! `manifold_*` name, matching the spirit of the fidget package (Rust as
//! the symbol-export boundary, C++ as the implementation).
//!
//! The shims are unsafe wrappers: every input pointer is forwarded as-is
//! and the C++ side does the validation. Callers (the Go package) are
//! expected to honor the documented ownership rules — release every
//! returned handle exactly once, free returned STL buffers via
//! `manifold_free_buffer`.

use std::os::raw::{c_int, c_void};

// Opaque type used purely for type-checking on the Go/Rust side. The C++
// glue casts to/from `manifold::Manifold*` internally.
#[repr(C)]
pub struct ManifoldHandle {
    _private: [u8; 0],
}

extern "C" {
    fn mglue_load_meshgl(
        verts: *const f32,
        nverts: usize,
        tris: *const u32,
        ntris: usize,
    ) -> *mut c_void;
    fn mglue_boolean(a: *const c_void, b: *const c_void, op: c_int) -> *mut c_void;
    fn mglue_transform(h: *const c_void, mat3x4: *const f32) -> *mut c_void;
    fn mglue_to_stl(
        h: *const c_void,
        out_ptr: *mut *mut u8,
        out_len: *mut usize,
    ) -> c_int;
    fn mglue_release(h: *mut c_void);
    fn mglue_free_buffer(ptr: *mut u8, len: usize);
}

#[no_mangle]
pub unsafe extern "C" fn manifold_load_meshgl(
    verts: *const f32,
    nverts: usize,
    tris: *const u32,
    ntris: usize,
) -> *mut ManifoldHandle {
    mglue_load_meshgl(verts, nverts, tris, ntris) as *mut ManifoldHandle
}

#[no_mangle]
pub unsafe extern "C" fn manifold_boolean(
    a: *const ManifoldHandle,
    b: *const ManifoldHandle,
    op: c_int,
) -> *mut ManifoldHandle {
    mglue_boolean(a as *const c_void, b as *const c_void, op) as *mut ManifoldHandle
}

#[no_mangle]
pub unsafe extern "C" fn manifold_transform(
    h: *const ManifoldHandle,
    mat3x4: *const f32,
) -> *mut ManifoldHandle {
    mglue_transform(h as *const c_void, mat3x4) as *mut ManifoldHandle
}

#[no_mangle]
pub unsafe extern "C" fn manifold_to_stl(
    h: *const ManifoldHandle,
    out_ptr: *mut *mut u8,
    out_len: *mut usize,
) -> c_int {
    mglue_to_stl(h as *const c_void, out_ptr, out_len)
}

#[no_mangle]
pub unsafe extern "C" fn manifold_release(h: *mut ManifoldHandle) {
    mglue_release(h as *mut c_void)
}

#[no_mangle]
pub unsafe extern "C" fn manifold_free_buffer(ptr: *mut u8, len: usize) {
    mglue_free_buffer(ptr, len)
}
