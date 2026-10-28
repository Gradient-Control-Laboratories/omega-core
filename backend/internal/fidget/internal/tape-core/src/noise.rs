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

use std::sync::{Arc, LazyLock, Mutex};

use fidget_core::{
    mesh_oracle::{self, MeshOracle as MeshOracleTrait},
    types::{Grad, Interval},
};

const TILE_SIZE: i64 = 64; // cells per axis; must match common/noise.TileSize
const CHANNELS: usize = 4; // RGBA; must match common/noise.Channels

// Reserved fidget mesh-oracle ids for the three noise channels. Far above any
// host-assigned mesh asset id (those are small import indices).
pub const NOISE_ORACLE_ID_BASE: u32 = 0xE0_15E_000;

// One-octave channels sit this far above the two-octave ones, so an oracle id
// names the channel AND the octave count. Keeping two octaves at the base means
// tapes written before the octaves argument existed still resolve unchanged.
// The offset is 4, not 3, so the ids line up with the tile's RGBA channels.
const ONE_OCTAVE_OFFSET: u32 = 4;

// The tile baked into the library at build time — the same bytes
// common/noise.Bake() ships (gen/main.go writes them here) and the GPU path
// loads as a texture. Embedding it makes the render lib self-sufficient: a
// process with no Go startup path (e.g. hodge replaying a tape) still has the
// tile, and the tape itself only needs to carry the noise leaf id. The omega
// fidget agent still calls register_noise_tile at startup, which just refreshes
// this to the exact Bake() output.
static EMBEDDED_TILE: &[u8] = include_bytes!(concat!(
    env!("CARGO_MANIFEST_DIR"),
    "/../../../../common/web/noise_tile.bin"
));

/// The active tile, shared by all channels. Defaults to the library-baked tile;
/// register_noise_tile overwrites it with the host's bake.
static NOISE_TILE: LazyLock<Mutex<Arc<Vec<u8>>>> =
    LazyLock::new(|| Mutex::new(Arc::new(EMBEDDED_TILE.to_vec())));

#[inline]
fn wrap(v: i64) -> usize {
    (((v % TILE_SIZE) + TILE_SIZE) % TILE_SIZE) as usize
}

/// Integer division rounding toward -inf (the supercell index). Matches the Go
/// reference's floorDiv.
#[inline]
fn floor_div(a: i64, b: i64) -> i64 {
    let q = a / b;
    if (a % b != 0) && ((a < 0) != (b < 0)) {
        q - 1
    } else {
        q
    }
}

/// One channel of the nearest texel at integer cell (x,y,z), as a float in
/// [0,1). Wraps the index like a repeating texture.
#[inline]
fn sample(tile: &[u8], x: i64, y: i64, z: i64, ch: usize) -> f32 {
    let i = ((wrap(z) * TILE_SIZE as usize + wrap(y)) * TILE_SIZE as usize + wrap(x)) * CHANNELS + ch;
    tile[i] as f32 / 255.0
}

#[inline]
fn fract(v: f32) -> f32 {
    v - v.floor()
}

/// The canonical per-cell hash for one channel: fine lookup + one supercell
/// fold, fract-combined. Mirrors common/noise.Hash3 exactly. `x,y,z` are
/// integer-valued cell coordinates (the GSL caller floors p before this).
fn hash_channel(tile: &[u8], x: f32, y: f32, z: f32, ch: usize) -> f32 {
    hash_channel_octaves(tile, x, y, z, ch, 2)
}

/// hash_channel over a chosen number of tile lookups — what the GSL builtin's
/// octaves argument selects. Mirrors common/noise.Hash3Octaves. One octave is
/// the bare nearest lookup: half the work, repeating every TILE_SIZE cells
/// instead of TILE_SIZE^2. It keeps the fract, because a texel byte of 255
/// samples as exactly 1.0 and the hash promises [0,1).
fn hash_channel_octaves(tile: &[u8], x: f32, y: f32, z: f32, ch: usize, octaves: u32) -> f32 {
    let xi = x.floor() as i64;
    let yi = y.floor() as i64;
    let zi = z.floor() as i64;
    let fine = sample(tile, xi, yi, zi, ch);
    if octaves <= 1 {
        return fract(fine);
    }
    let coarse = sample(
        tile,
        floor_div(xi, TILE_SIZE),
        floor_div(yi, TILE_SIZE),
        floor_div(zi, TILE_SIZE),
        ch,
    );
    fract(fine + coarse)
}

// ============ Host API ================================================

/// Install the baked tile. `bytes` must be TILE_SIZE^3 * CHANNELS long;
/// returns false (and changes nothing) otherwise.
pub fn set_tile(bytes: &[u8]) -> bool {
    let want = (TILE_SIZE * TILE_SIZE * TILE_SIZE) as usize * CHANNELS;
    if bytes.len() != want {
        return false;
    }
    *NOISE_TILE.lock().unwrap() = Arc::new(bytes.to_vec());
    true
}

/// All three RGB channels of the hash at cell (x,y,z), from the active tile
/// (host bake if registered, else the library-baked tile).
pub fn hash3(x: f32, y: f32, z: f32) -> [f32; 3] {
    hash3_octaves(x, y, z, 2)
}

/// A snapshot of the active tile, for an oracle that wants to hold its own copy
/// rather than lock on every evaluation.
pub(crate) fn tile_snapshot() -> Arc<Vec<u8>> {
    NOISE_TILE.lock().unwrap().clone()
}

/// All three channels from a tile the caller already holds. The cellular oracle
/// hashes 27 cells per evaluation, so it must not take the tile mutex 27 times.
pub(crate) fn hash3_from_tile(tile: &[u8], x: f32, y: f32, z: f32, octaves: u32) -> [f32; 3] {
    [0, 1, 2].map(|ch| hash_channel_octaves(tile, x, y, z, ch, octaves))
}

/// hash3 over a chosen number of tile lookups (see hash_channel_octaves).
pub fn hash3_octaves(x: f32, y: f32, z: f32, octaves: u32) -> [f32; 3] {
    let tile = NOISE_TILE.lock().unwrap().clone();
    [0, 1, 2].map(|ch| hash_channel_octaves(&tile, x, y, z, ch, octaves))
}

// ============ Oracle ==================================================

/// A fidget oracle that evaluates one channel of the noise hash. Holds the tile
/// it was registered with, so re-baking doesn't disturb live trees.
struct NoiseTileOracle {
    tile: Arc<Vec<u8>>,
    channel: usize,
}

/// The same oracle over a single tile lookup, the octaves=1 form. A separate
/// type rather than a field on NoiseTileOracle so the two-octave oracle — the
/// one every existing tape and test refers to — keeps its exact shape.
struct NoiseTileOracleOneOctave {
    tile: Arc<Vec<u8>>,
    channel: usize,
}

/// The per-cell-constant interval both oracles return, given a way to read the
/// value at a cell.
///
/// The hash is per-cell-constant. When the query box falls within a single cell
/// on every axis (floor(lo) == floor(hi)), the value is exactly one number —
/// return it as a degenerate interval. This is the key to fast meshing: inside a
/// cell the feature points are fixed, so the surrounding cellular() distance is
/// locally smooth and MDC's length/min interval propagation becomes tight enough
/// to prune. Only boxes straddling a cell boundary fall back to the full [0,1),
/// and those are exactly where the field is non-smooth and MDC should refine.
fn cell_constant_interval(
    x: Interval,
    y: Interval,
    z: Interval,
    at: impl Fn(f32, f32, f32) -> f32,
) -> Interval {
    let cx = x.lower().floor();
    let cy = y.lower().floor();
    let cz = z.lower().floor();
    if cx == x.upper().floor() && cy == y.upper().floor() && cz == z.upper().floor() {
        let v = at(cx, cy, cz);
        return Interval::new(v, v);
    }
    Interval::new(0.0, 1.0)
}

impl MeshOracleTrait for NoiseTileOracleOneOctave {
    fn point(&self, x: f32, y: f32, z: f32) -> f32 {
        hash_channel_octaves(&self.tile, x, y, z, self.channel, 1)
    }

    fn interval(&self, x: Interval, y: Interval, z: Interval) -> Interval {
        cell_constant_interval(x, y, z, |cx, cy, cz| {
            hash_channel_octaves(&self.tile, cx, cy, cz, self.channel, 1)
        })
    }

    fn grad(&self, x: f32, y: f32, z: f32) -> Grad {
        // Per-cell-constant: zero spatial gradient, as in NoiseTileOracle.
        Grad::new(
            hash_channel_octaves(&self.tile, x, y, z, self.channel, 1),
            0.0,
            0.0,
            0.0,
        )
    }
}

impl MeshOracleTrait for NoiseTileOracle {
    fn point(&self, x: f32, y: f32, z: f32) -> f32 {
        hash_channel(&self.tile, x, y, z, self.channel)
    }

    fn interval(&self, x: Interval, y: Interval, z: Interval) -> Interval {
        cell_constant_interval(x, y, z, |cx, cy, cz| {
            hash_channel(&self.tile, cx, cy, cz, self.channel)
        })
    }

    fn grad(&self, x: f32, y: f32, z: f32) -> Grad {
        // Per-cell-constant: zero spatial gradient.
        Grad::new(hash_channel(&self.tile, x, y, z, self.channel), 0.0, 0.0, 0.0)
    }
}

// ============ Tape replay =============================================
//
// An omega tape serializes noise-tile oracle leaves as plain MeshOracle ids; the
// tile itself is library-baked (EMBEDDED_TILE), not carried in the tape. These
// helpers let the tape deserializer spot a noise leaf and register its channel
// oracle in a consumer process (e.g. hodge) without the C-ABI / TREES dance.

/// If `id` names a noise-channel oracle leaf, return its channel (0..3) and the
/// octave count the id encodes. A tape is self-describing this way: the octaves
/// argument is a compile-time literal, so it lands in the leaf id rather than
/// anywhere in the tape body.
pub fn noise_channel_of_id(id: u32) -> Option<(u32, u32)> {
    let slot = id.wrapping_sub(NOISE_ORACLE_ID_BASE);
    if slot < 3 {
        return Some((slot, 2));
    }
    let one = slot.wrapping_sub(ONE_OCTAVE_OFFSET);
    (one < 3).then_some((one, 1))
}

/// The oracle leaf id for a channel at an octave count — the inverse of
/// noise_channel_of_id. Out-of-range inputs are rejected by the callers below.
pub fn noise_oracle_id(channel: u32, octaves: u32) -> u32 {
    if octaves <= 1 {
        NOISE_ORACLE_ID_BASE + ONE_OCTAVE_OFFSET + channel
    } else {
        NOISE_ORACLE_ID_BASE + channel
    }
}

/// Register the mesh-oracle leaf for `channel` (0..3) at `octaves` lookups using
/// the active tile, so a `Tree::mesh_oracle(noise_oracle_id(..))` resolves on
/// replay. Mirrors the registration half of `tree_noise_channel`.
pub fn register_channel(channel: u32, octaves: u32) -> Result<(), String> {
    if channel >= 3 {
        return Err(format!("noise channel {} out of range", channel));
    }
    let tile = NOISE_TILE.lock().unwrap().clone();
    let id = noise_oracle_id(channel, octaves);
    let oracle: Arc<dyn MeshOracleTrait> = if octaves <= 1 {
        Arc::new(NoiseTileOracleOneOctave {
            tile,
            channel: channel as usize,
        })
    } else {
        Arc::new(NoiseTileOracle {
            tile,
            channel: channel as usize,
        })
    };
    mesh_oracle::register(id, oracle);
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    // A synthetic tile so the test doesn't depend on the host's bake: every
    // byte is its flattened index mod 256, which still exercises wrap, the
    // supercell fold, and the period.
    fn synthetic_tile() -> Vec<u8> {
        let n = (TILE_SIZE * TILE_SIZE * TILE_SIZE) as usize * CHANNELS;
        (0..n).map(|i| i as u8).collect()
    }

    #[test]
    fn hash_in_unit_range() {
        let tile = synthetic_tile();
        for &(x, y, z) in &[(0.0, 0.0, 0.0), (5.0, 9.0, 13.0), (-1.0, -7.0, -130.0)] {
            for ch in 0..3 {
                let h = hash_channel(&tile, x, y, z, ch);
                assert!((0.0..1.0).contains(&h), "hash {} out of [0,1)", h);
            }
        }
    }

    #[test]
    fn interval_tight_within_a_cell_loose_across() {
        let oracle = NoiseTileOracle {
            tile: std::sync::Arc::new(synthetic_tile()),
            channel: 0,
        };
        // A box inside one cell collapses to the exact value (degenerate).
        let within = oracle.interval(
            Interval::new(3.1, 3.9),
            Interval::new(7.0, 7.2),
            Interval::new(-2.9, -2.1),
        );
        let exact = hash_channel(&oracle.tile, 3.0, 7.0, -3.0, 0);
        assert_eq!(within.lower(), exact);
        assert_eq!(within.upper(), exact);
        // A box straddling a cell boundary falls back to [0,1].
        let across = oracle.interval(
            Interval::new(3.5, 4.5),
            Interval::new(7.0, 7.0),
            Interval::new(-3.0, -3.0),
        );
        assert_eq!(across.lower(), 0.0);
        assert_eq!(across.upper(), 1.0);
    }

    #[test]
    fn fold_sets_period_to_tilesize_squared() {
        let tile = synthetic_tile();
        let period = (TILE_SIZE * TILE_SIZE) as f32;
        for ch in 0..3 {
            let base = hash_channel(&tile, 0.0, 0.0, 0.0, ch);
            // +TileSize changes the value (period > TileSize)...
            assert_ne!(base, hash_channel(&tile, TILE_SIZE as f32, 0.0, 0.0, ch));
            // ...+TileSize^2 repeats it.
            assert_eq!(base, hash_channel(&tile, period, 0.0, 0.0, ch));
        }
    }
}
