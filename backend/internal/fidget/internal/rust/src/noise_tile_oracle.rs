//! Noise-tile oracle for the GSL `noiseHash3` builtin.
//!
//! The host (Go, common/noise) bakes a TILE_SIZE^3 RGBA8 tile with a real
//! integer hash and ships the bytes here via `register_noise_tile`. At
//! evaluation time we only do float work — a nearest (integer-indexed) lookup
//! plus one "supercell" fold — so the result is identical to the Go
//! `noise.Hash3` reference and to the WGSL shader, and stays bounded for MDC.
//!
//! `noiseHash3(cell)` lowers (in the Go convert layer) to three of these
//! oracle leaves — one per RGB channel — whose x,y,z inputs are remapped to the
//! cell coordinate. Each leaf evaluates one channel.
//
// The tile, the hash and the oracle itself live in tape-core (noise.rs), which
// the browser build shares; this file is the C ABI the Go side calls, and the
// tree builder that pushes onto TREES.

use std::os::raw::{c_int, c_ulong};

use fidget::context::Tree;
use tape_core::noise;

// Reserved fidget mesh-oracle ids for the three noise channels (tape-core).
// Re-exported for the rest of the crate; tree_noise_channel itself now asks
// noise::noise_oracle_id, which encodes the octave count alongside the channel.
#[allow(unused_imports)]
pub use tape_core::noise::NOISE_ORACLE_ID_BASE;

// ============ FFI =====================================================

/// Install the baked tile. `len` must be TILE_SIZE^3 * CHANNELS. Returns 0 on
/// success, -1 on a null pointer or wrong length.
///
/// # Safety
/// `bytes` must point to `len` valid bytes.
#[no_mangle]
pub unsafe extern "C" fn register_noise_tile(bytes: *const u8, len: usize) -> c_int {
    if bytes.is_null() {
        return -1;
    }
    let slice = unsafe { std::slice::from_raw_parts(bytes, len) };
    if noise::set_tile(slice) { 0 } else { -1 }
}

/// Evaluate all three RGB channels of the hash at cell (x,y,z) into `out[0..3]`.
/// Uses the active tile (host bake if registered, else the library-baked tile).
/// Used for host-side parity testing against common/noise.Hash3.
///
/// # Safety
/// `out` must point to at least 3 writable f32s.
#[no_mangle]
pub unsafe extern "C" fn noise_hash3(x: f32, y: f32, z: f32, out: *mut f32) {
    let h = noise::hash3(x, y, z);
    for ch in 0..3 {
        unsafe { *out.add(ch) = h[ch] };
    }
}

/// noise_hash3 over a chosen octave count (1 or 2), for host-side parity testing
/// of the one-octave hash against common/noise.Hash3Octaves.
///
/// # Safety
/// `out` must point to at least 3 writable f32s.
#[no_mangle]
pub unsafe extern "C" fn noise_hash3_octaves(x: f32, y: f32, z: f32, octaves: u32, out: *mut f32) {
    let h = noise::hash3_octaves(x, y, z, octaves);
    for ch in 0..3 {
        unsafe { *out.add(ch) = h[ch] };
    }
}

/// Build a `Tree` leaf evaluating channel `channel` (0..3) of the noise hash at
/// the implicit (x,y,z) — the convert layer remaps those to the cell — over
/// `octaves` tile lookups (1 or 2; see noise::hash_channel_octaves). Snapshots
/// the active tile (host bake if registered, else the library-baked tile).
/// Returns the TREES index, or u64::MAX if the channel is out of range.
#[no_mangle]
pub extern "C" fn tree_noise_channel(channel: u32, octaves: u32) -> c_ulong {
    if noise::register_channel(channel, octaves).is_err() {
        return u64::MAX;
    }
    let id = noise::noise_oracle_id(channel, octaves);
    // The (v - v) triplet is zero but keeps Var::X/Y/Z in the tape so the
    // octree feeds real per-axis intervals to interval() (see tree_mesh_oracle).
    let (x, y, z) = Tree::axes();
    let zero_via_axes = (x.clone() - x) + (y.clone() - y) + (z.clone() - z);
    let tree = Tree::mesh_oracle(id) + zero_via_axes;
    let mut trees = crate::TREES.lock().unwrap_or_else(std::sync::PoisonError::into_inner);
    trees.push(tree);
    (trees.len() - 1) as c_ulong
}
