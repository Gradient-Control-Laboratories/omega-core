//! OpenVDB-backed signed distance oracle for `fidget::Op::MeshOracle` leaves.
//!
//! Wired up in two steps by the host:
//!
//!  1. Call [`register_vdb_oracle`] with a triangle soup, voxel size, and
//!     half-band width. The C++ side bakes a narrow-band level set with
//!     `openvdb::tools::meshToLevelSet` and the result is cached.
//!  2. Call [`tree_vdb_oracle`] to get a `Tree` leaf that, when evaluated,
//!     dispatches through fidget-core's mesh-oracle registry to this oracle.
//!
//! Asset-id namespacing: VDB asset ids are stored in fidget's shared
//! mesh-oracle registry with the high bit set (`asset_id | VDB_BIT`), so a
//! GSL program calling both `bvhSdf(0)` and `vdbSdf(0)` registers two
//! distinct oracles without conflict.

use core::ffi::c_ulong;
use std::collections::HashMap;
use std::os::raw::c_int;
use std::sync::{Arc, LazyLock, Mutex};
use std::time::Instant;

use cxx::UniquePtr;
use fidget::{
    context::Tree,
    mesh_oracle::{self, MeshOracle as MeshOracleTrait},
    types::{Grad, Interval},
};

use crate::openvdb_ffi::ffi;

/// High bit reserved on asset ids registered as VDB oracles. Keeps them
/// from colliding with BVH oracles registered by `mesh_oracle.rs` in the
/// shared `fidget::mesh_oracle::register` table.
pub const VDB_ASSET_BIT: u32 = 0x8000_0000;

pub struct VdbOracle {
    grid: UniquePtr<ffi::FloatGridHandle>,
    bbox_min: [f32; 3],
    bbox_max: [f32; 3],
    voxel_size: f32,
}

// SAFETY: `FloatGridHandle` wraps `openvdb::FloatGrid::Ptr` (a shared_ptr
// with thread-safe reference counting). After `from_triangles` returns we
// only call read-only sampling APIs on the grid; OpenVDB's tree accessors
// are documented as safe for concurrent readers, and we hold the handle by
// `&` everywhere except construction. The build-time invariant is that
// nobody mutates the grid after the oracle is registered.
unsafe impl Send for VdbOracle {}
unsafe impl Sync for VdbOracle {}

impl VdbOracle {
    pub fn from_triangles(
        verts: &[f32],
        tris: &[u32],
        voxel_size: f32,
        half_band_voxels: f32,
    ) -> Self {
        ffi::ensure_initialized();
        let grid = ffi::mesh_to_level_set(verts, tris, voxel_size, half_band_voxels);
        let bbox = ffi::world_bbox(&grid);
        let voxel_size = ffi::voxel_size_value(&grid);
        Self {
            grid,
            bbox_min: [bbox.min_x, bbox.min_y, bbox.min_z],
            bbox_max: [bbox.max_x, bbox.max_y, bbox.max_z],
            voxel_size,
        }
    }

    fn signed_distance(&self, p: [f32; 3]) -> f32 {
        let sample = ffi::sample_point(&self.grid, p[0], p[1], p[2]);
        // Distance from p to the bbox (0 if inside). Outside, extend the
        // SDF by the triangle inequality so the field stays Lipschitz-1
        // and monotonic — same formula the WGSL preview uses, see
        // agent/wgsl/internal/codegen/mesh_sdf.go:142-166.
        let dx = (self.bbox_min[0] - p[0]).max(0.0).max(p[0] - self.bbox_max[0]);
        let dy = (self.bbox_min[1] - p[1]).max(0.0).max(p[1] - self.bbox_max[1]);
        let dz = (self.bbox_min[2] - p[2]).max(0.0).max(p[2] - self.bbox_max[2]);
        let outside_dist = (dx * dx + dy * dy + dz * dz).sqrt();
        if outside_dist == 0.0 {
            sample
        } else {
            sample + outside_dist
        }
    }
}

impl MeshOracleTrait for VdbOracle {
    fn point(&self, x: f32, y: f32, z: f32) -> f32 {
        self.signed_distance([x, y, z])
    }

    fn interval(&self, x: Interval, y: Interval, z: Interval) -> Interval {
        // Sample at 8 corners; SDF is 1-Lipschitz so any interior point is
        // within half-diagonal of its nearest corner. Mirrors the BVH
        // oracle's interval impl in mesh_oracle.rs:351-382.
        let xs = [x.lower(), x.upper()];
        let ys = [y.lower(), y.upper()];
        let zs = [z.lower(), z.upper()];
        let mut lo = f32::INFINITY;
        let mut hi = f32::NEG_INFINITY;
        for &cx in &xs {
            for &cy in &ys {
                for &cz in &zs {
                    let d = self.signed_distance([cx, cy, cz]);
                    if d < lo {
                        lo = d;
                    }
                    if d > hi {
                        hi = d;
                    }
                }
            }
        }
        let dx = xs[1] - xs[0];
        let dy = ys[1] - ys[0];
        let dz = zs[1] - zs[0];
        let half_diag = 0.5 * (dx * dx + dy * dy + dz * dz).sqrt();
        Interval::new(lo - half_diag, hi + half_diag)
    }

    fn grad(&self, x: f32, y: f32, z: f32) -> Grad {
        // Central differences with eps tied to the voxel grid spacing —
        // any smaller and we just sample the same trilinear cell on both
        // sides and get zero. Floor matches the sphere-tracer epsilon.
        let eps = (self.voxel_size * 0.5).max(1e-3);
        let v = self.signed_distance([x, y, z]);
        let dxp = self.signed_distance([x + eps, y, z]);
        let dxm = self.signed_distance([x - eps, y, z]);
        let dyp = self.signed_distance([x, y + eps, z]);
        let dym = self.signed_distance([x, y - eps, z]);
        let dzp = self.signed_distance([x, y, z + eps]);
        let dzm = self.signed_distance([x, y, z - eps]);
        let gx = (dxp - dxm) / (2.0 * eps);
        let gy = (dyp - dym) / (2.0 * eps);
        let gz = (dzp - dzm) / (2.0 * eps);
        Grad::new(v, gx, gy, gz)
    }
}

// ============ FFI registry =============================================

static ORACLES: LazyLock<Mutex<HashMap<u32, Arc<VdbOracle>>>> =
    LazyLock::new(|| Mutex::new(HashMap::new()));

/// Build a `VdbOracle` from raw pointers and install it under `asset_id`.
///
/// # Safety
///
/// * `vertices_ptr` must point to `vertex_count * 3` valid f32 values.
/// * `triangles_ptr` must point to `tri_count * 3` valid u32 values.
/// * All triangle indices must be `< vertex_count`.
///
/// Returns `0` on success, `-1` on validation failure.
#[no_mangle]
pub unsafe extern "C" fn register_vdb_oracle(
    asset_id: u32,
    vertices_ptr: *const f32,
    vertex_count: usize,
    triangles_ptr: *const u32,
    tri_count: usize,
    voxel_size: f32,
    half_band_voxels: f32,
) -> c_int {
    if vertices_ptr.is_null() || triangles_ptr.is_null() {
        return -1;
    }
    if voxel_size <= 0.0 || half_band_voxels <= 0.0 {
        return -1;
    }
    let verts: &[f32] =
        unsafe { std::slice::from_raw_parts(vertices_ptr, vertex_count * 3) };
    let tris: &[u32] =
        unsafe { std::slice::from_raw_parts(triangles_ptr, tri_count * 3) };

    for i in 0..tri_count {
        for k in 0..3 {
            if tris[i * 3 + k] as usize >= vertex_count {
                return -1;
            }
        }
    }

    let oracle = Arc::new(VdbOracle::from_triangles(
        verts,
        tris,
        voxel_size,
        half_band_voxels,
    ));
    eprintln!(
        "register_vdb_oracle: asset={} verts={} tris={} voxel={} band={} bbox=({:.3},{:.3},{:.3})..({:.3},{:.3},{:.3})",
        asset_id, vertex_count, tri_count, voxel_size, half_band_voxels,
        oracle.bbox_min[0], oracle.bbox_min[1], oracle.bbox_min[2],
        oracle.bbox_max[0], oracle.bbox_max[1], oracle.bbox_max[2],
    );
    {
        let mut map = ORACLES.lock().unwrap_or_else(std::sync::PoisonError::into_inner);
        map.insert(asset_id, oracle.clone());
    }
    mesh_oracle::register(asset_id | VDB_ASSET_BIT, oracle);
    0
}

/// Returns `1` if a VDB oracle has been registered for this asset id.
#[no_mangle]
pub extern "C" fn has_vdb_oracle(asset_id: u32) -> c_int {
    let map = ORACLES.lock().unwrap_or_else(std::sync::PoisonError::into_inner);
    if map.contains_key(&asset_id) {
        1
    } else {
        0
    }
}

/// Sample the registered VDB oracle's level set into a uniform dense grid
/// and return its packed binary representation. Layout (little-endian):
///
///  - `dims[3]`: u32 — grid dims in x, y, z
///  - `bbox_min[3]`: f32 — world-space lower corner
///  - `bbox_max[3]`: f32 — world-space upper corner
///  - `values[N]`: f32 — grid samples, row-major with x fastest then y then z
///
/// Returns the malloc'd buffer pointer and writes its length into
/// `*out_len`. Returns NULL on failure (asset not registered, zero dims).
/// Caller must release with [`free_vdb_dense_bytes`].
#[no_mangle]
pub unsafe extern "C" fn extract_vdb_dense_bytes(
    asset_id: u32,
    dims_x: u32,
    dims_y: u32,
    dims_z: u32,
    out_len: *mut usize,
) -> *mut u8 {
    if out_len.is_null() {
        return std::ptr::null_mut();
    }
    unsafe { *out_len = 0 };
    if dims_x == 0 || dims_y == 0 || dims_z == 0 {
        return std::ptr::null_mut();
    }
    let oracle = {
        let map = ORACLES.lock().unwrap_or_else(std::sync::PoisonError::into_inner);
        match map.get(&asset_id) {
            Some(o) => o.clone(),
            None => return std::ptr::null_mut(),
        }
    };
    let values = ffi::extract_dense_grid(&oracle.grid, dims_x, dims_y, dims_z);
    let total = (dims_x as usize) * (dims_y as usize) * (dims_z as usize);
    if values.len() != total {
        return std::ptr::null_mut();
    }

    let header_bytes = 3 * 4 + 6 * 4; // dims (3*u32) + bbox (6*f32)
    let total_bytes = header_bytes + total * 4;
    let mut buf: Vec<u8> = Vec::with_capacity(total_bytes);
    buf.extend_from_slice(&dims_x.to_le_bytes());
    buf.extend_from_slice(&dims_y.to_le_bytes());
    buf.extend_from_slice(&dims_z.to_le_bytes());
    for &c in &oracle.bbox_min {
        buf.extend_from_slice(&c.to_le_bytes());
    }
    for &c in &oracle.bbox_max {
        buf.extend_from_slice(&c.to_le_bytes());
    }
    for v in &values {
        buf.extend_from_slice(&v.to_le_bytes());
    }
    debug_assert_eq!(buf.len(), total_bytes);

    let len = buf.len();
    let ptr = buf.as_mut_ptr();
    std::mem::forget(buf);
    unsafe { *out_len = len };
    ptr
}

impl VdbOracle {
    /// The level set as a tape carries it, for a reader that samples it
    /// without OpenVDB (tape-core's level_set). Errors if the tree holds
    /// root tiles, whose values would fall outside the block table.
    pub fn level_set(&self) -> Result<tape_core::tape_proto::VdbLevelSet, String> {
        let blocks = ffi::extract_level_set(&self.grid);
        if blocks.root_tiles != 0 {
            return Err(format!("the level set holds {} root tiles", blocks.root_tiles));
        }
        Ok(tape_core::tape_proto::VdbLevelSet {
            voxel_size: blocks.voxel_size,
            background: blocks.background,
            active_bbox: vec![
                self.bbox_min[0], self.bbox_min[1], self.bbox_min[2],
                self.bbox_max[0], self.bbox_max[1], self.bbox_max[2],
            ],
            block_origin: vec![blocks.origin_x, blocks.origin_y, blocks.origin_z],
            block_dims: vec![blocks.dims_x, blocks.dims_y, blocks.dims_z],
            block_leaf: blocks.block_leaf,
            block_value: blocks.block_value,
            leaf_values: blocks.leaf_values,
        })
    }
}

/// The level set registered under `asset_id` (see [`VdbOracle::level_set`]),
/// as an encoded `tape.VdbLevelSet` message. Returns NULL when no VDB oracle
/// is registered there or its tree cannot be written out. Free the buffer with
/// [`free_vdb_dense_bytes`].
///
/// # Safety
///
/// `out_len` must point to a writable `usize`.
#[no_mangle]
pub unsafe extern "C" fn extract_vdb_level_set(asset_id: u32, out_len: *mut usize) -> *mut u8 {
    use prost::Message;
    if out_len.is_null() {
        return std::ptr::null_mut();
    }
    unsafe { *out_len = 0 };
    let oracle = {
        let map = ORACLES.lock().unwrap_or_else(std::sync::PoisonError::into_inner);
        match map.get(&asset_id) {
            Some(o) => o.clone(),
            None => return std::ptr::null_mut(),
        }
    };
    let message = match oracle.level_set() {
        Ok(m) => m,
        Err(e) => {
            eprintln!("extract_vdb_level_set: asset {}: {}", asset_id, e);
            return std::ptr::null_mut();
        }
    };
    // A boxed slice's capacity is exactly its length, which the free assumes.
    let buf = message.encode_to_vec().into_boxed_slice();
    let len = buf.len();
    let ptr = Box::into_raw(buf) as *mut u8;
    unsafe { *out_len = len };
    ptr
}

/// Frees a buffer returned by [`extract_vdb_dense_bytes`] or
/// [`extract_vdb_level_set`]. Safe to call with `(NULL, 0)`.
///
/// # Safety
///
/// `ptr` must be NULL or a value returned from one of those, and `len` must
/// match the length reported at allocation time.
#[no_mangle]
pub unsafe extern "C" fn free_vdb_dense_bytes(ptr: *mut u8, len: usize) {
    if ptr.is_null() || len == 0 {
        return;
    }
    unsafe {
        let _ = Vec::from_raw_parts(ptr, len, len);
    }
}

/// Self-calibrating cost ratio between the bake's two phases:
/// `(seconds per phase-1 work unit) / (seconds per phase-2 work unit)`.
/// Both work-unit counts are computed up front from the mesh and the
/// requested voxel density, but the *relative* cost of the two
/// algorithms is not — it depends on the implementation and the host
/// CPU — so every bake records its actual phase durations and folds the
/// observed ratio into this EWMA. The first bake after startup uses the
/// seed; subsequent bakes track reality. `samples == 0` means "still on
/// the seed".
struct VdbCostRatio {
    ratio: f64,
    samples: u32,
}

/// Seeded at ~30:1 (a phase-1 narrow-band voxel is taken to cost roughly
/// 30× a phase-2 dense sample); replaced by the first real measurement.
static VDB_COST_RATIO: Mutex<VdbCostRatio> =
    Mutex::new(VdbCostRatio { ratio: 30.0, samples: 0 });

/// Sum of triangle areas — the dominant driver of meshToLevelSet cost,
/// since the narrow band it builds scales with surface area. O(tris),
/// negligible next to the bake itself.
fn mesh_surface_area(verts: &[f32], tris: &[u32]) -> f64 {
    let vertex = |i: u32| -> [f64; 3] {
        let base = i as usize * 3;
        [
            verts[base] as f64,
            verts[base + 1] as f64,
            verts[base + 2] as f64,
        ]
    };
    let mut area = 0.0_f64;
    for tri in tris.chunks_exact(3) {
        let a = vertex(tri[0]);
        let b = vertex(tri[1]);
        let c = vertex(tri[2]);
        let e1 = [b[0] - a[0], b[1] - a[1], b[2] - a[2]];
        let e2 = [c[0] - a[0], c[1] - a[1], c[2] - a[2]];
        let cross = [
            e1[1] * e2[2] - e1[2] * e2[1],
            e1[2] * e2[0] - e1[0] * e2[2],
            e1[0] * e2[1] - e1[1] * e2[0],
        ];
        area += 0.5
            * (cross[0] * cross[0] + cross[1] * cross[1] + cross[2] * cross[2])
                .sqrt();
    }
    area
}

/// Bake a triangle soup into an OpenVDB level set, register the
/// resulting oracle under `asset_id`, sample it onto a uniform dense
/// grid, and return the packed bytes — all under a single progress
/// callback that ticks per-mille completion (0..=1000) across both
/// phases:
///
///  - `meshToLevelSet` reports through the OpenVDB Interrupter at
///    known checkpoints inside its narrow-band voxelization (covers
///    0..=500).
///  - The dense-grid extract emits one tick per Z-slice plus a final
///    1000 (covers 501..=1000).
///
/// Concurrent invocations serialize on a global mutex so the callback
/// pointer is unique to the active bake. The mesh-out path keeps using
/// [`register_vdb_oracle`] (no progress) — this entry point exists for
/// the canvas-side preview IPC where we want a marching progress bar.
///
/// # Safety
///
/// Pointer / count contract matches [`register_vdb_oracle`]; `progress_cb`
/// may be null to skip progress emission.
#[no_mangle]
pub unsafe extern "C" fn bake_and_extract_vdb_with_progress(
    asset_id: u32,
    vertices_ptr: *const f32,
    vertex_count: usize,
    triangles_ptr: *const u32,
    tri_count: usize,
    voxel_size: f32,
    half_band_voxels: f32,
    dims_x: u32,
    dims_y: u32,
    dims_z: u32,
    progress_cb: Option<crate::openvdb_ffi::VdbProgressCallback>,
    out_len: *mut usize,
) -> *mut u8 {
    if out_len.is_null() {
        return std::ptr::null_mut();
    }
    unsafe { *out_len = 0 };
    if vertices_ptr.is_null() || triangles_ptr.is_null() {
        return std::ptr::null_mut();
    }
    if voxel_size <= 0.0 || half_band_voxels <= 0.0 {
        return std::ptr::null_mut();
    }
    if dims_x == 0 || dims_y == 0 || dims_z == 0 {
        return std::ptr::null_mut();
    }
    let verts: &[f32] =
        unsafe { std::slice::from_raw_parts(vertices_ptr, vertex_count * 3) };
    let tris: &[u32] =
        unsafe { std::slice::from_raw_parts(triangles_ptr, tri_count * 3) };
    for i in 0..tri_count {
        for k in 0..3 {
            if tris[i * 3 + k] as usize >= vertex_count {
                return std::ptr::null_mut();
            }
        }
    }

    let _bake_guard = crate::openvdb_ffi::VDB_BAKE_LOCK.lock().unwrap_or_else(std::sync::PoisonError::into_inner);
    crate::openvdb_ffi::set_progress_callback(progress_cb);

    // Size the two bake phases from the actual mesh and the GSL-requested
    // voxel density, so the progress bar advances in proportion to real
    // work. Phase 1 (meshToLevelSet) cost tracks the narrow-band voxel
    // count — surface_area · half_band / voxel_size² — and phase 2 (the
    // dense extract) cost tracks the dense grid voxel count. A phase-1
    // voxel is far more expensive to process than a phase-2 sample, and
    // that ratio is a property of the algorithm + hardware rather than
    // anything derivable from the inputs, so it is self-calibrated — see
    // VDB_COST_RATIO. With the phases weighted, set_progress_phase /
    // vdb_progress_emit remap each phase's local progress into one honest
    // global bar.
    let phase1_units = (mesh_surface_area(verts, tris) * half_band_voxels as f64
        / (voxel_size as f64 * voxel_size as f64))
        .max(1.0);
    let phase2_units = (dims_x as f64 * dims_y as f64 * dims_z as f64).max(1.0);
    let cost_ratio = VDB_COST_RATIO.lock().unwrap_or_else(std::sync::PoisonError::into_inner).ratio;
    let phase1_cost = phase1_units * cost_ratio;
    let phase2_cost = phase2_units;
    let total_cost = phase1_cost + phase2_cost;
    let scale = crate::openvdb_ffi::PROGRESS_SCALE;
    // Clamp so each phase keeps at least a sliver of the bar even when
    // one phase is negligible — keeps the bar monotonic and non-stuck.
    let phase1_span = ((phase1_cost / total_cost) * scale as f64)
        .round()
        .clamp(1.0, (scale - 1) as f64) as u32;

    crate::openvdb_ffi::set_progress_phase(0, phase1_span);
    let phase1_start = Instant::now();
    let oracle = Arc::new(VdbOracle::from_triangles(
        verts,
        tris,
        voxel_size,
        half_band_voxels,
    ));
    let phase1_secs = phase1_start.elapsed().as_secs_f64();
    {
        let mut map = ORACLES.lock().unwrap_or_else(std::sync::PoisonError::into_inner);
        map.insert(asset_id, oracle.clone());
    }
    mesh_oracle::register(asset_id | VDB_ASSET_BIT, oracle.clone());

    crate::openvdb_ffi::set_progress_phase(phase1_span, scale - phase1_span);
    let phase2_start = Instant::now();
    let values = ffi::extract_dense_grid(&oracle.grid, dims_x, dims_y, dims_z);
    let phase2_secs = phase2_start.elapsed().as_secs_f64();
    crate::openvdb_ffi::set_progress_callback(None);

    // Fold the observed per-unit cost ratio back into the calibration so
    // the next bake's phase split is sharper. The first sample replaces
    // the seed outright; later ones EWMA in at 0.5.
    let calibrated_ratio = {
        let mut cal = VDB_COST_RATIO.lock().unwrap_or_else(std::sync::PoisonError::into_inner);
        let observed =
            (phase1_secs / phase1_units) / (phase2_secs.max(1e-9) / phase2_units);
        if observed.is_finite() && observed > 0.0 {
            cal.ratio = if cal.samples == 0 {
                observed
            } else {
                0.5 * cal.ratio + 0.5 * observed
            }
            .clamp(0.1, 100_000.0);
            cal.samples += 1;
        }
        cal.ratio
    };
    eprintln!(
        "vdb-bake: phase1={:.3}s ({:.0} units) phase2={:.3}s ({:.0} units) \
         cost_ratio={:.1} (asset {}, grid {}×{}×{})",
        phase1_secs,
        phase1_units,
        phase2_secs,
        phase2_units,
        calibrated_ratio,
        asset_id,
        dims_x,
        dims_y,
        dims_z,
    );

    let total = (dims_x as usize) * (dims_y as usize) * (dims_z as usize);
    if values.len() != total {
        return std::ptr::null_mut();
    }

    let header_bytes = 3 * 4 + 6 * 4;
    let total_bytes = header_bytes + total * 4;
    let mut buf: Vec<u8> = Vec::with_capacity(total_bytes);
    buf.extend_from_slice(&dims_x.to_le_bytes());
    buf.extend_from_slice(&dims_y.to_le_bytes());
    buf.extend_from_slice(&dims_z.to_le_bytes());
    for &c in &oracle.bbox_min {
        buf.extend_from_slice(&c.to_le_bytes());
    }
    for &c in &oracle.bbox_max {
        buf.extend_from_slice(&c.to_le_bytes());
    }
    for v in &values {
        buf.extend_from_slice(&v.to_le_bytes());
    }
    debug_assert_eq!(buf.len(), total_bytes);

    let len = buf.len();
    let ptr = buf.as_mut_ptr();
    std::mem::forget(buf);
    unsafe { *out_len = len };
    ptr
}

/// Build a `Tree` leaf that evaluates to the VDB oracle's signed distance
/// at the implicit `(x, y, z)`. Pushes onto `TREES` and returns its index.
///
/// The `(x-x) + (y-y) + (z-z)` trick keeps `Var::X/Y/Z` in the tape's
/// variable map; without it, fidget-mesh's octree builder writes per-axis
/// intervals to nonexistent slots and produces empty meshes silently. See
/// the matching note in `mesh_oracle.rs:677-684`.
#[no_mangle]
pub extern "C" fn tree_vdb_oracle(asset_id: u32) -> c_ulong {
    let (x, y, z) = Tree::axes();
    let zero_via_axes = (x.clone() - x) + (y.clone() - y) + (z.clone() - z);
    let tree = Tree::mesh_oracle(asset_id | VDB_ASSET_BIT) + zero_via_axes;
    let mut trees = crate::TREES.lock().unwrap_or_else(std::sync::PoisonError::into_inner);
    trees.push(tree);
    (trees.len() - 1) as c_ulong
}

#[cfg(test)]
mod tests {
    use super::*;

    fn cube_mesh() -> (Vec<f32>, Vec<u32>) {
        let v = vec![
            -1.0, -1.0, -1.0, 1.0, -1.0, -1.0, 1.0, 1.0, -1.0, -1.0, 1.0, -1.0, // z=-1
            -1.0, -1.0, 1.0, 1.0, -1.0, 1.0, 1.0, 1.0, 1.0, -1.0, 1.0, 1.0, // z=+1
        ];
        let t = vec![
            0, 2, 1, 0, 3, 2, // -z
            4, 5, 6, 4, 6, 7, // +z
            0, 1, 5, 0, 5, 4, // -y
            2, 3, 7, 2, 7, 6, // +y
            0, 4, 7, 0, 7, 3, // -x
            1, 2, 6, 1, 6, 5, // +x
        ];
        (v, t)
    }

    #[test]
    fn cube_signed_distance() {
        let (v, t) = cube_mesh();
        let oracle = VdbOracle::from_triangles(&v, &t, 0.05, 3.0);

        // Inside (origin): negative.
        assert!(oracle.point(0.0, 0.0, 0.0) < 0.0);
        // On a face: near zero.
        assert!(oracle.point(1.0, 0.0, 0.0).abs() < 0.1);
        // Just outside: small positive.
        let outside_near = oracle.point(1.05, 0.0, 0.0);
        assert!(outside_near > 0.0 && outside_near < 0.2);
        // Far outside: distance grows monotonically thanks to the
        // bbox-distance extension term.
        let far = oracle.point(5.0, 0.0, 0.0);
        assert!(far > outside_near, "far {far} should exceed near {outside_near}");
    }

    #[test]
    fn interval_brackets_point() {
        let (v, t) = cube_mesh();
        let oracle = VdbOracle::from_triangles(&v, &t, 0.05, 3.0);
        let ix = Interval::new(-0.5, 0.5);
        let iy = Interval::new(-0.5, 0.5);
        let iz = Interval::new(-0.5, 0.5);
        let bounds = oracle.interval(ix, iy, iz);
        let center = oracle.point(0.0, 0.0, 0.0);
        assert!(
            bounds.lower() <= center && center <= bounds.upper(),
            "interval {bounds:?} must bracket center sample {center}"
        );
    }

    fn thin_slab_mesh() -> (Vec<f32>, Vec<u32>) {
        let h = 0.334_f32;
        let w = 3.5_f32;
        let v: Vec<f32> = vec![
            -w, -w, -h, w, -w, -h, w, w, -h, -w, w, -h, // bottom face
            -w, -w, h, w, -w, h, w, w, h, -w, w, h, // top face
        ];
        let t: Vec<u32> = vec![
            0, 2, 1, 0, 3, 2, // -z
            4, 5, 6, 4, 6, 7, // +z
            0, 1, 5, 0, 5, 4, // -y
            3, 7, 6, 3, 6, 2, // +y
            0, 4, 7, 0, 7, 3, // -x
            1, 2, 6, 1, 6, 5, // +x
        ];
        (v, t)
    }

    /// Canary for the empty-octree trap: without the `(x-x)+(y-y)+(z-z)`
    /// axis-preservation trick on the tree leaf, fidget-mesh's octree
    /// builder writes per-axis intervals to nonexistent variable slots and
    /// produces an empty mesh silently. Mirrors mesh_oracle.rs:854.
    #[test]
    fn thin_slab_offset_produces_non_empty_octree() {
        use fidget::{gui::View3, vm::VmShape, Context};
        use fidget_mesh::{Octree, Settings as MeshSettings};
        use nalgebra::Vector3;

        let (verts, tris) = thin_slab_mesh();
        unsafe {
            let rc = register_vdb_oracle(
                17,
                verts.as_ptr(),
                verts.len() / 3,
                tris.as_ptr(),
                tris.len() / 3,
                0.05,
                3.0,
            );
            assert_eq!(rc, 0, "register_vdb_oracle failed");
        }

        let (x, y, z) = Tree::axes();
        let zero_via_axes = (x.clone() - x) + (y.clone() - y) + (z.clone() - z);
        let augmented = Tree::mesh_oracle(17 | VDB_ASSET_BIT) + zero_via_axes;
        let tree = augmented - Tree::constant(1.0);
        let mut ctx = Context::new();
        let node = ctx.import(&tree);
        assert!(
            ctx.contains_mesh_oracle(node),
            "tree must contain MeshOracle"
        );

        let shape = VmShape::new(&ctx, node).expect("VmShape::new");
        let view = View3::from_center_and_scale(Vector3::new(0.0_f32, 0.0, 0.0), 3.0_f32);
        let settings = MeshSettings {
            depth: 5,
            world_to_model: view.world_to_model(),
            ..Default::default()
        };
        let vars = fidget::shape::ShapeVars::<f32>::new();
        let bound = shape.bind(&vars).expect("bind");
        let octree = Octree::build(&bound, &settings).expect("Octree::build");
        let mesh = octree.walk_dual();
        eprintln!(
            "thin_slab vdb: {} verts, {} tris",
            mesh.vertices.len(),
            mesh.triangles.len()
        );
        assert!(
            !mesh.triangles.is_empty(),
            "expected non-empty mesh for inflated thin slab"
        );
    }

    /// Guards against the gradient-stretching bug: when `remap()`
    /// substitutes the oracle's sample point with a non-uniform-scaled
    /// expression like `[p.x, p.y, 0.5 * p.z]`, the gradient evaluator
    /// must apply the chain rule. Without it the oracle returns a
    /// gradient in *local* coords which the rest of the pipeline treats
    /// as *world* coords — MDC then positions Hermite vertices using
    /// stale gradients and the triangulation distorts increasingly with
    /// distance from the origin.
    #[test]
    fn remap_grad_applies_chain_rule() {
        use fidget::{
            shape::EzShape, types::Grad, vm::VmShape, Context,
        };

        let (verts, tris) = cube_mesh();
        unsafe {
            let rc = register_vdb_oracle(
                81,
                verts.as_ptr(),
                verts.len() / 3,
                tris.as_ptr(),
                tris.len() / 3,
                0.05,
                3.0,
            );
            assert_eq!(rc, 0);
        }

        let oracle = Tree::mesh_oracle(81 | VDB_ASSET_BIT);

        // Remapped tree: `oracle(p.x, p.y, 0.5 * p.z)`.
        // After remap the local SDF is f_local(x, y, z) = cube(x, y, 0.5z),
        // i.e. the cube stretched 2× in world-Z. World-space gradient
        // is df/dz_world = df/dz_local * 0.5 (chain rule). Without the
        // chain rule applied in `RegOp::MeshOracle`, the world gradient
        // would equal the local gradient, which is exactly the bug that
        // produced "increasingly stretched triangles" in MDC output.
        let half_z = Tree::z() * 0.5;
        let zero_via_axes = (Tree::x() - Tree::x())
            + (Tree::y() - Tree::y())
            + (Tree::z() - Tree::z());
        let remapped =
            oracle.clone().remap_xyz(Tree::x(), Tree::y(), half_z)
                + zero_via_axes.clone();
        let local = oracle + zero_via_axes;

        let mut ctx = Context::new();
        let remapped_node = ctx.import(&remapped);
        let local_node = ctx.import(&local);

        let remapped_shape =
            VmShape::new(&ctx, remapped_node).expect("VmShape::new remapped");
        let local_shape =
            VmShape::new(&ctx, local_node).expect("VmShape::new local");

        let remapped_tape =
            remapped_shape.ez_grad_slice_tape();
        let local_tape =
            local_shape.ez_grad_slice_tape();
        let mut eval = VmShape::new_grad_slice_eval();

        // Sample at world (0.5, 0, 1.5). Local sample point after
        // remap = (0.5, 0, 0.75) — inside the unit cube, closest face
        // is +z at distance 0.25, so df/dz_local ≈ 1 and the other
        // two partials ≈ 0.
        let world_xs = [Grad::new(0.5, 1.0, 0.0, 0.0)];
        let world_ys = [Grad::new(0.0, 0.0, 1.0, 0.0)];
        let world_zs = [Grad::new(1.5, 0.0, 0.0, 1.0)];
        let remapped_g =
            eval.eval(&remapped_tape, &world_xs, &world_ys, &world_zs).unwrap()[0];

        // Sample the un-remapped oracle at the corresponding local
        // point so we can compare the bare local gradient against the
        // chain-rule result.
        let local_xs = [Grad::new(0.5, 1.0, 0.0, 0.0)];
        let local_ys = [Grad::new(0.0, 0.0, 1.0, 0.0)];
        let local_zs = [Grad::new(0.75, 0.0, 0.0, 1.0)];
        let local_g = eval
            .eval(&local_tape, &local_xs, &local_ys, &local_zs)
            .unwrap()[0];

        let expected_dzw = 0.5 * local_g.dz;
        eprintln!(
            "remapped: dx={} dy={} dz={}; local: dx={} dy={} dz={}; expected dz_w={}",
            remapped_g.dx,
            remapped_g.dy,
            remapped_g.dz,
            local_g.dx,
            local_g.dy,
            local_g.dz,
            expected_dzw
        );
        // Tolerance accounts for the oracle's finite-difference step
        // jitter; the dominant relationship is dz_world ≈ 0.5 *
        // dz_local. Without the chain rule we'd see dz_world ≈ 1.0 *
        // dz_local — well outside this tolerance.
        let err = (remapped_g.dz - expected_dzw).abs();
        assert!(
            err < 0.1,
            "chain rule not applied: remapped df/dz_world = {}, \
             expected 0.5 * local df/dz_local = {} (err = {})",
            remapped_g.dz,
            expected_dzw,
            err
        );
    }

    /// Mirrors how the canvas pipeline drives MDC after the FFI rewrite:
    /// instead of `View3::from_center_and_scale` we build an anisotropic
    /// `world_to_model` matrix straight from the meshing bbox. The VDB
    /// oracle must still be visited and produce a non-empty mesh — if it
    /// silently isn't, the user sees a "vdbSdf doesn't enter the mesh
    /// calculation" symptom.
    #[test]
    fn anisotropic_bbox_matrix_meshes_vdb_oracle() {
        use crate::world_to_model_from_bbox;
        use fidget::{vm::VmShape, Context};
        use fidget_mesh::{Octree, Settings as MeshSettings};

        let (verts, tris) = cube_mesh();
        unsafe {
            let rc = register_vdb_oracle(
                73,
                verts.as_ptr(),
                verts.len() / 3,
                tris.as_ptr(),
                tris.len() / 3,
                0.05,
                3.0,
            );
            assert_eq!(rc, 0, "register_vdb_oracle failed");
        }

        let (x, y, z) = Tree::axes();
        let zero_via_axes = (x.clone() - x) + (y.clone() - y) + (z.clone() - z);
        let tree = Tree::mesh_oracle(73 | VDB_ASSET_BIT) + zero_via_axes;
        let mut ctx = Context::new();
        let node = ctx.import(&tree);

        let shape = VmShape::new(&ctx, node).expect("VmShape::new");
        // Anisotropic bbox: tall and asymmetric around origin. The cube
        // mesh sits at [-1, 1]³ — the asymmetric region is deliberately
        // wider on +x to confirm the matrix translates correctly.
        let bbox_min = [-3.0_f32, -2.0, -2.0];
        let bbox_max = [4.0_f32, 2.0, 2.0];
        let world_to_model = world_to_model_from_bbox(bbox_min, bbox_max);
        let settings = MeshSettings {
            depth: 5,
            world_to_model,
            ..Default::default()
        };
        let vars = fidget::shape::ShapeVars::<f32>::new();
        let bound = shape.bind(&vars).expect("bind");
        let octree = Octree::build(&bound, &settings).expect("Octree::build");
        let mesh = octree.walk_dual();
        eprintln!(
            "anisotropic vdb mdc: {} verts, {} tris",
            mesh.vertices.len(),
            mesh.triangles.len()
        );
        assert!(
            !mesh.triangles.is_empty(),
            "anisotropic bbox MDC must not silently miss the VDB surface"
        );
    }

    #[test]
    fn registry_round_trip() {
        let (v, t) = cube_mesh();
        unsafe {
            let rc = register_vdb_oracle(
                42,
                v.as_ptr(),
                v.len() / 3,
                t.as_ptr(),
                t.len() / 3,
                0.05,
                3.0,
            );
            assert_eq!(rc, 0);
        }
        assert_eq!(has_vdb_oracle(42), 1);
        assert_eq!(has_vdb_oracle(43), 0);
    }
}

// The level set a tape carries, read back by tape-core without OpenVDB, has
// to answer exactly as the oracle it was written from: the browser meshes it
// and its STL must match the server's.
#[cfg(test)]
mod level_set_parity {
    use super::*;
    use tape_core::level_set::LevelSetOracle;

    // A box big enough that OpenVDB fills its inside with tiles rather than
    // leaves, off the origin so index coordinates go negative.
    fn box_mesh(c: [f32; 3], h: [f32; 3]) -> (Vec<f32>, Vec<u32>) {
        let mut v = Vec::new();
        for k in 0..8 {
            let s = [(k & 1) as f32 * 2.0 - 1.0, ((k >> 1) & 1) as f32 * 2.0 - 1.0, ((k >> 2) & 1) as f32 * 2.0 - 1.0];
            for a in 0..3 {
                v.push(c[a] + s[a] * h[a]);
            }
        }
        let t = vec![
            0, 2, 1, 1, 2, 3, 4, 5, 6, 5, 7, 6, // -z, +z
            0, 1, 4, 1, 5, 4, 2, 6, 3, 3, 6, 7, // -y, +y
            0, 4, 2, 2, 4, 6, 1, 3, 5, 3, 7, 5, // -x, +x
        ];
        (v, t)
    }

    fn assert_same(a: f32, b: f32, what: &str) {
        assert!(a.to_bits() == b.to_bits(), "{what}: server {a} vs tape {b}");
    }

    #[test]
    fn tape_level_set_answers_as_the_oracle() {
        let (v, t) = box_mesh([-3.3, 1.7, 0.4], [9.0, 7.5, 2.2]);
        let oracle = VdbOracle::from_triangles(&v, &t, 0.05, 3.0);
        let message = oracle.level_set().expect("level set");
        let inside_tiles = message.block_leaf.iter().zip(&message.block_value)
            .filter(|(l, v)| **l < 0 && **v < 0.0)
            .count();
        assert!(inside_tiles > 0, "the box's inside should hold tiles, not only leaves");
        let read = LevelSetOracle::new(&message).expect("read level set");

        let mut s = 0x2545_F491_4F6C_DD1Du64;
        let mut rnd = || {
            s ^= s << 13;
            s ^= s >> 7;
            s ^= s << 17;
            (s >> 40) as f32 / (1u64 << 24) as f32
        };
        // Through the box, its band and past the grid on every side.
        let lo = [-13.5f32, -6.9, -2.9];
        let hi = [6.9f32, 10.3, 3.7];
        for _ in 0..20000 {
            let p: [f32; 3] = core::array::from_fn(|a| lo[a] + rnd() * (hi[a] - lo[a]));
            assert_same(oracle.point(p[0], p[1], p[2]), read.point(p[0], p[1], p[2]), "point");
            let (g, h) = (oracle.grad(p[0], p[1], p[2]), read.grad(p[0], p[1], p[2]));
            for (a, b) in [(g.v, h.v), (g.dx, h.dx), (g.dy, h.dy), (g.dz, h.dz)] {
                assert_same(a, b, "grad");
            }
            let w = rnd() * 0.7;
            let (x, y, z) = (
                Interval::new(p[0], p[0] + w),
                Interval::new(p[1], p[1] + w),
                Interval::new(p[2], p[2] + w),
            );
            let (i, j) = (oracle.interval(x, y, z), read.interval(x, y, z));
            assert_same(i.lower(), j.lower(), "interval lower");
            assert_same(i.upper(), j.upper(), "interval upper");
        }
    }

    // The same field meshed through either oracle is the same STL.
    #[test]
    fn tape_level_set_meshes_as_the_oracle() {
        use crate::world_to_model_from_bbox;
        use fidget::{vm::VmShape, Context};
        use fidget_mesh::{Octree, Settings as MeshSettings};

        let (v, t) = box_mesh([0.6, -0.4, 0.2], [1.3, 0.9, 0.7]);
        let oracle = Arc::new(VdbOracle::from_triangles(&v, &t, 0.05, 3.0));
        let read = Arc::new(LevelSetOracle::new(&oracle.level_set().expect("level set")).expect("read"));
        let key = 91 | VDB_ASSET_BIT;
        let stl = |o: Arc<dyn MeshOracleTrait>| {
            mesh_oracle::register(key, o);
            let (x, y, z) = Tree::axes();
            let tree = Tree::mesh_oracle(key) + ((x.clone() - x) + (y.clone() - y) + (z.clone() - z));
            let mut ctx = Context::new();
            let node = ctx.import(&tree);
            let shape = VmShape::new(&ctx, node).expect("VmShape::new");
            let settings = MeshSettings {
                depth: 6,
                world_to_model: world_to_model_from_bbox([-1.2, -1.8, -1.0], [2.4, 1.0, 1.4]),
                threads: None,
                ..Default::default()
            };
            let vars = fidget::shape::ShapeVars::<f32>::new();
            let bound = shape.bind(&vars).expect("bind");
            let mesh = Octree::build(&bound, &settings).expect("Octree::build").walk_dual();
            let mut out = Vec::new();
            mesh.write_stl(&mut out).expect("write_stl");
            out
        };
        let server = stl(oracle);
        let tape = stl(read);
        assert!(server.len() > 84 + 50 * 100, "the box should mesh ({} bytes)", server.len());
        assert!(server == tape, "STLs differ ({} vs {} bytes)", server.len(), tape.len());
    }
}
