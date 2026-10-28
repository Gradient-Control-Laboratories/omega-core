#[cxx::bridge(namespace = "omega_vdb")]
pub mod ffi {
    pub struct BBox {
        pub min_x: f32,
        pub min_y: f32,
        pub min_z: f32,
        pub max_x: f32,
        pub max_y: f32,
        pub max_z: f32,
    }

    // A level set's tree as plain arrays (see tape.proto's VdbLevelSet):
    // the leaves, and the value of every 8^3 block without one, over a
    // block table one block wider than the leaves on each side.
    pub struct LevelSetBlocks {
        pub voxel_size: f64,
        pub background: f32,
        pub origin_x: i32,
        pub origin_y: i32,
        pub origin_z: i32,
        pub dims_x: u32,
        pub dims_y: u32,
        pub dims_z: u32,
        pub block_leaf: Vec<i32>,
        pub block_value: Vec<f32>,
        pub leaf_values: Vec<f32>,
        // Tiles held at the root, which would put values outside the table.
        pub root_tiles: u32,
    }

    extern "Rust" {
        // Bridge invoked from C++ during long-running OpenVDB operations
        // (meshToLevelSet's Interrupter and the dense-grid extract loop).
        // The Rust side dispatches to the active progress callback set
        // by `set_progress_callback` on the calling sequence; with one
        // bake in flight at a time this is race-free, and worker
        // threads inside meshToLevelSet all forward to the same global
        // sink through an atomic load.
        fn vdb_progress_emit(completed: u32, total: u32);
    }

    unsafe extern "C++" {
        include!("openvdb_glue.h");

        type FloatGridHandle;

        fn ensure_initialized();

        // Bake a triangle soup into a narrow-band SDF level set.
        //
        // `verts` is a flat array of (x, y, z) floats; `tris` is a flat
        // array of three vertex indices per triangle. `voxel_size` is the
        // world-space cell size; `half_band_voxels` is the narrow-band
        // half-width in voxels (background magnitude is
        // `voxel_size * half_band_voxels`).
        fn mesh_to_level_set(
            verts: &[f32],
            tris: &[u32],
            voxel_size: f32,
            half_band_voxels: f32,
        ) -> UniquePtr<FloatGridHandle>;

        fn sample_point(grid: &FloatGridHandle, x: f32, y: f32, z: f32) -> f32;
        fn world_bbox(grid: &FloatGridHandle) -> BBox;
        fn voxel_size_value(grid: &FloatGridHandle) -> f32;

        // Sample the level set into a uniform dense grid sized
        // `dims_x × dims_y × dims_z` over the world-space bbox returned
        // by `world_bbox(grid)`. Output is row-major with x fastest then
        // y then z (matches meshasset.SdfGrid layout in Go). Cells outside
        // the narrow band receive the grid's background sentinel.
        fn extract_dense_grid(
            grid: &FloatGridHandle,
            dims_x: u32,
            dims_y: u32,
            dims_z: u32,
        ) -> Vec<f32>;

        // The tree's leaves and block values, for a reader that samples the
        // level set without OpenVDB.
        fn extract_level_set(grid: &FloatGridHandle) -> LevelSetBlocks;
    }
}

// =================== Progress sink ===================

use std::sync::atomic::{AtomicU32, AtomicU64, Ordering};
use std::sync::Mutex;

/// Active progress callback for the in-flight VDB bake. Stored as a
/// raw `extern "C" fn` pointer so OpenVDB's TBB worker threads (which
/// drive the meshToLevelSet Interrupter) can read it without locking.
/// The progress channel is single-writer at the C-FFI boundary —
/// `vdb_bake_lock` serializes bake invocations end-to-end so there's
/// at most one active callback at any time.
static VDB_PROGRESS_CB: AtomicU64 = AtomicU64::new(0);

/// Serializes VDB bakes that emit progress so the global callback
/// pointer can't be reassigned mid-bake by a concurrent caller.
pub static VDB_BAKE_LOCK: Mutex<()> = Mutex::new(());

/// Type of the C ABI progress callback the host wires in.
pub type VdbProgressCallback = extern "C" fn(completed: u32, total: u32);

/// Set the active progress callback for the duration of a bake. The
/// caller must hold `VDB_BAKE_LOCK` so we don't race with another
/// invocation.
pub fn set_progress_callback(cb: Option<VdbProgressCallback>) {
    let raw = match cb {
        Some(f) => f as usize as u64,
        None => 0,
    };
    VDB_PROGRESS_CB.store(raw, Ordering::Release);
}

/// Fixed denominator for the host-facing progress fraction. Each bake
/// reports `completed` in `[0, PROGRESS_SCALE]`; the browser just renders
/// `completed / PROGRESS_SCALE`. 10_000 gives sub-0.1% granularity.
pub const PROGRESS_SCALE: u32 = 10_000;

/// The current phase's slice of the `[0, PROGRESS_SCALE]` global bake
/// timeline. The orchestrator sizes these from the measured per-phase
/// cost (see `bake_and_extract_vdb_with_progress`) and calls
/// `set_progress_phase` between the two phases; `vdb_progress_emit` then
/// maps the phase-local fraction it gets from C++ into `[BASE, BASE+SPAN]`.
/// Sizing the slices by cost is what makes the global bar advance roughly
/// linearly in wall time — so the browser's elapsed/fraction ETA is honest
/// and the frontend needs no estimation logic of its own.
static PROGRESS_PHASE_BASE: AtomicU32 = AtomicU32::new(0);
static PROGRESS_PHASE_SPAN: AtomicU32 = AtomicU32::new(PROGRESS_SCALE);

/// Point `vdb_progress_emit` at a new phase's slice of the global bar.
/// Called by the orchestrator on the bake thread between the two C++
/// calls — no progress tick is in flight across that boundary, so plain
/// atomic stores are sufficient.
pub fn set_progress_phase(base: u32, span: u32) {
    PROGRESS_PHASE_BASE.store(base, Ordering::Release);
    PROGRESS_PHASE_SPAN.store(span, Ordering::Release);
}

/// Invoked from C++ each time OpenVDB ticks an Interrupter or the extract
/// loop finishes a Z-slice. C++ reports *phase-local* progress
/// (`completed / total` in `[0, 1]`); we remap it into the current
/// phase's slice of the global `[0, PROGRESS_SCALE]` bar and forward that
/// to whatever callback the host registered for the current bake.
pub fn vdb_progress_emit(completed: u32, total: u32) {
    let raw = VDB_PROGRESS_CB.load(Ordering::Acquire);
    if raw == 0 {
        return;
    }
    let base = u64::from(PROGRESS_PHASE_BASE.load(Ordering::Acquire));
    let span = u64::from(PROGRESS_PHASE_SPAN.load(Ordering::Acquire));
    let local = if total > 0 {
        u64::from(completed.min(total)) * span / u64::from(total)
    } else {
        0
    };
    let global = (base + local).min(u64::from(PROGRESS_SCALE)) as u32;
    let cb: VdbProgressCallback = unsafe { std::mem::transmute(raw as usize) };
    cb(global, PROGRESS_SCALE);
}

#[cfg(test)]
mod tests {
    use super::ffi;

    #[test]
    fn cube_level_set_signs() {
        ffi::ensure_initialized();

        // Single 2×2×2 cube centered at origin (12 triangles).
        let v: Vec<f32> = vec![
            -1.0, -1.0, -1.0, 1.0, -1.0, -1.0, 1.0, 1.0, -1.0, -1.0, 1.0, -1.0, // z=-1
            -1.0, -1.0, 1.0, 1.0, -1.0, 1.0, 1.0, 1.0, 1.0, -1.0, 1.0, 1.0, // z=+1
        ];
        let t: Vec<u32> = vec![
            0, 2, 1, 0, 3, 2, // -z
            4, 5, 6, 4, 6, 7, // +z
            0, 1, 5, 0, 5, 4, // -y
            2, 3, 7, 2, 7, 6, // +y
            0, 4, 7, 0, 7, 3, // -x
            1, 2, 6, 1, 6, 5, // +x
        ];
        let grid = ffi::mesh_to_level_set(&v, &t, 0.05, 3.0);
        assert!(!grid.is_null());

        // Inside (origin): negative; outside (far away): positive (background).
        let inside = ffi::sample_point(&grid, 0.0, 0.0, 0.0);
        let outside = ffi::sample_point(&grid, 5.0, 0.0, 0.0);
        assert!(inside < 0.0, "origin should be inside cube, got {inside}");
        assert!(outside > 0.0, "far point should be outside, got {outside}");
    }

    #[test]
    fn extract_dense_grid_round_trip() {
        ffi::ensure_initialized();
        let v: Vec<f32> = vec![
            -1.0, -1.0, -1.0, 1.0, -1.0, -1.0, 1.0, 1.0, -1.0, -1.0, 1.0, -1.0,
            -1.0, -1.0, 1.0, 1.0, -1.0, 1.0, 1.0, 1.0, 1.0, -1.0, 1.0, 1.0,
        ];
        let t: Vec<u32> = vec![
            0, 2, 1, 0, 3, 2, 4, 5, 6, 4, 6, 7, 0, 1, 5, 0, 5, 4,
            2, 3, 7, 2, 7, 6, 0, 4, 7, 0, 7, 3, 1, 2, 6, 1, 6, 5,
        ];
        let grid = ffi::mesh_to_level_set(&v, &t, 0.05, 3.0);
        let dense = ffi::extract_dense_grid(&grid, 16, 16, 16);
        assert_eq!(dense.len(), 16 * 16 * 16);
        // Center cell (8,8,8) should be solidly inside (negative).
        let center = dense[8 + 8 * 16 + 8 * 16 * 16];
        assert!(center < 0.0, "center cell should be inside cube, got {center}");
    }
}
