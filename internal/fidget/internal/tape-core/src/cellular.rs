//! Cellular (Worley) noise as a handful of fidget oracles instead of an
//! unrolled 27-neighbour scan.
//!
//! Why this exists
//! ---------------
//! Expressed in GSL, cellular noise is a 27-neighbour fold that the build layer
//! unrolls at compile time. Every evaluation therefore costs 27 hash lookups
//! lowered as 81 separate oracle leaves (three channels each), threaded through
//! generic interval arithmetic. Two things follow, and both are bad:
//!
//!   * Per point, the mesher pays 81 dynamically-dispatched oracle calls.
//!   * The per-cell-constant hash gives a degenerate interval only when the
//!     query box sits inside ONE lattice cell, so MDC cannot prune anything
//!     until the octree has refined below the noise cell size. That is a
//!     minimum depth independent of where the surface actually is, and it costs
//!     a cubic number of cells on fine noise: measured, the fold got SLOWER as
//!     the cell size shrank while these oracles get faster.
//!
//! As oracles, both problems disappear. The scan runs here in native code, and
//! because F1 and F2 are EXACTLY 1-Lipschitz — each is a distance to a point
//! set, and the k-th smallest of 1-Lipschitz functions is 1-Lipschitz — their
//! interval() is tight at any box size and prunes from the top of the octree
//! down.
//!
//! The lattice must match the GSL definition in common/source/package/noise,
//! the Go reference in common/noise, and the WGSL emitter exactly: cell c holds
//! one feature point at c + hash(c), and the scan covers the 27 cells around
//! floor(q). Coordinates here are CELL units; the caller remaps p/size into the
//! oracle and scales the result by size, which leaves the Lipschitz constant at
//! 1 in world units too.

use std::cell::RefCell;
use std::sync::Arc;

use fidget_core::{
    mesh_oracle::{self, MeshOracle as MeshOracleTrait},
    types::{Grad, Interval},
};

use crate::noise;

/// Reserved fidget mesh-oracle id block for the cellular oracles. Far above any
/// host-assigned mesh asset id, and clear of the noise tile's own block.
pub const CELLULAR_ORACLE_ID_BASE: u32 = 0xE0_15F_000;

/// What a cellular oracle leaf returns. The numbering is shared with the Go
/// convert layer and the WGSL emitter, so it is part of the contract.
pub mod component {
    /// Distance to the nearest feature point. Exactly 1-Lipschitz.
    pub const F1: u32 = 0;
    /// Distance to the second-nearest feature point. Also exactly 1-Lipschitz.
    pub const F2: u32 = 1;
    /// The nearest feature point, x/y/z. Piecewise constant.
    pub const FEATURE_X: u32 = 2;
    /// The integer lattice cell owning the nearest point, x/y/z.
    pub const NEAR_CELL_X: u32 = 5;
    /// The integer lattice cell owning the second-nearest point, x/y/z.
    pub const FAR_CELL_X: u32 = 8;
    /// The second-nearest feature point, x/y/z. Piecewise constant. The scan
    /// already has it, so F2's gradient needs no second hash to recover it.
    pub const FAR_FEATURE_X: u32 = 11;
    /// The feature point whose wall is nearest: the neighbour of the nearest
    /// point whose bisector plane with it lies closest to q, x/y/z. Piecewise
    /// constant. It carries WALL's gradient, as FEATURE carries F1's.
    pub const WALL_FEATURE_X: u32 = 14;
    /// The exact distance to the nearest Voronoi wall: the boundary of the cell
    /// q lies in. Exactly 1-Lipschitz (a distance to a set: the union of the
    /// walls). Read by its own scan, not the one above; see walls_uncached.
    pub const WALL: u32 = 17;
    /// One past the last component.
    pub const COUNT: u32 = 18;
    /// How many consecutive components a point-valued one spans.
    pub const AXES: u32 = 3;
}

/// Cells per axis in the scanned neighbourhood, and the number of candidates
/// that makes. A cell's nearest feature point can lie in any immediate
/// neighbour when the jitter fills the cell, and cannot lie further, so this is
/// both necessary and sufficient.
const NEIGHBOURS_PER_AXIS: usize = 3;
const CANDIDATES: usize = NEIGHBOURS_PER_AXIS * NEIGHBOURS_PER_AXIS * NEIGHBOURS_PER_AXIS;

/// The neighbour offsets, in the order cellularOffsets() lists them in the GSL
/// library: x fastest, then y, then z. The order does not change F1 or F2, but
/// keeping it identical lets the ports be compared candidate by candidate when
/// one of them disagrees.
const OFFSETS: [[f32; 3]; CANDIDATES] = {
    let mut out = [[0.0; 3]; CANDIDATES];
    let mut i = 0;
    while i < CANDIDATES {
        // Unpack i base NEIGHBOURS_PER_AXIS and centre it, so offsets run -1..1.
        const PLANE: usize = NEIGHBOURS_PER_AXIS * NEIGHBOURS_PER_AXIS;
        const CENTRE: f32 = (NEIGHBOURS_PER_AXIS / 2) as f32;
        out[i] = [
            (i % NEIGHBOURS_PER_AXIS) as f32 - CENTRE,
            ((i / NEIGHBOURS_PER_AXIS) % NEIGHBOURS_PER_AXIS) as f32 - CENTRE,
            (i / PLANE) as f32 - CENTRE,
        ];
        i += 1;
    }
    out
};

/// Everything one scan of the neighbourhood yields.
#[derive(Clone, Copy)]
pub struct Scan {
    pub f1: f32,
    pub f2: f32,
    pub feature: [f32; 3],
    pub near_cell: [f32; 3],
    pub far_cell: [f32; 3],
    pub far_feature: [f32; 3],
}

impl Scan {
    /// The axis of a point-valued component, given the component that starts it.
    fn axis(c: u32, start: u32) -> usize {
        (c - start) as usize
    }

    fn component(&self, c: u32) -> f32 {
        use component::{AXES, F1, F2, FAR_CELL_X, FAR_FEATURE_X, FEATURE_X, NEAR_CELL_X};
        match c {
            F1 => self.f1,
            F2 => self.f2,
            _ if c < FEATURE_X + AXES => self.feature[Self::axis(c, FEATURE_X)],
            _ if c < NEAR_CELL_X + AXES => self.near_cell[Self::axis(c, NEAR_CELL_X)],
            _ if c < FAR_CELL_X + AXES => self.far_cell[Self::axis(c, FAR_CELL_X)],
            _ if c < FAR_FEATURE_X + AXES => self.far_feature[Self::axis(c, FAR_FEATURE_X)],
            _ => f32::NAN,
        }
    }
}

/// Scan the 3x3x3 neighbourhood around floor(q).
///
/// The arithmetic is `pf - (o + h)` with `pf = q - floor(q)`, exactly as the GSL
/// fold writes it, and NOT the algebraically equal `q - (cell + h)`. Every term
/// there is O(1), so the hash keeps all its bits; forming the absolute feature
/// point first would round h into a coordinate of magnitude |q| and throw away
/// its low bits, which at |q| ~ 1000 cells costs about 1e-4 of a cell and would
/// show up as the three ports disagreeing.
///
/// Squared distances throughout, rooted at the end: sqrt is monotonic, so the
/// two winners are the same ones the GSL fold picks.
fn scan_uncached(tile: &[u8], q: [f32; 3], octaves: u32) -> Scan {
    let pi = [q[0].floor(), q[1].floor(), q[2].floor()];
    let pf = [q[0] - pi[0], q[1] - pi[1], q[2] - pi[2]];
    let mut best = f32::INFINITY;
    let mut second = f32::INFINITY;
    let mut best_o = [0.0f32; 3];
    let mut best_h = [0.0f32; 3];
    let mut second_o = [0.0f32; 3];
    // The runner-up's hash is carried alongside its offset for the same reason
    // the winner's is: knowing both closest points means knowing BOTH gradients,
    // so F2's derivative must not have to hash the far cell a second time to
    // recover a point this scan already visited.
    let mut second_h = [0.0f32; 3];
    for o in OFFSETS.iter() {
        let h = noise::hash3_from_tile(tile, pi[0] + o[0], pi[1] + o[1], pi[2] + o[2], octaves);
        let d = [
            pf[0] - (o[0] + h[0]),
            pf[1] - (o[1] + h[1]),
            pf[2] - (o[2] + h[2]),
        ];
        let d2 = d[0] * d[0] + d[1] * d[1] + d[2] * d[2];
        if d2 < best {
            second = best;
            second_o = best_o;
            second_h = best_h;
            best = d2;
            best_o = *o;
            best_h = h;
        } else if d2 < second {
            second = d2;
            second_o = *o;
            second_h = h;
        }
    }
    let cell = |o: [f32; 3]| [pi[0] + o[0], pi[1] + o[1], pi[2] + o[2]];
    let point = |c: [f32; 3], h: [f32; 3]| [c[0] + h[0], c[1] + h[1], c[2] + h[2]];
    let near_cell = cell(best_o);
    let far_cell = cell(second_o);
    Scan {
        f1: best.sqrt(),
        f2: second.sqrt(),
        // Absolute feature points: these are the only parts that carry |q|
        // magnitude, and only derivative expressions ask for them.
        feature: point(near_cell, best_h),
        near_cell,
        far_cell,
        far_feature: point(far_cell, second_h),
    }
}

/// How many cells the wall pass may reach from floor(q) on each side, per axis:
/// a 5x5x5 window, the 3-D form of iq's second pass
/// (https://iquilezles.org/articles/voronoilines/). The pass visits only the cells
/// inside the radius F1 + 2 * wall that the 27 candidates already bound, which
/// has never been seen past 1.5 cells, so this cap is a guard on the loop rather
/// than a limit on the answer.
const WALL_REACH: f32 = 2.0;

/// Everything the wall pass yields.
#[derive(Clone, Copy)]
pub struct Walls {
    pub wall: f32,
    pub wall_feature: [f32; 3],
}

impl Walls {
    fn component(&self, c: u32) -> f32 {
        use component::{AXES, WALL, WALL_FEATURE_X};
        match c {
            WALL => self.wall,
            _ if (WALL_FEATURE_X..WALL_FEATURE_X + AXES).contains(&c) => {
                self.wall_feature[(c - WALL_FEATURE_X) as usize]
            }
            _ => f32::NAN,
        }
    }
}

/// The exact distance from q to the nearest Voronoi wall, by iq's two passes.
///
/// q lies in the cell of its nearest feature point a, and every wall it can
/// reach first is a face of that cell: the bisector plane between a and some
/// neighbour c. A cell is convex, so the distance to its boundary is the least
/// distance to those planes, and that least distance is the distance to the
/// union of all walls. F2 - F1 is not it: that difference is ~twice the
/// distance beside a wall and drifts with how far apart the two points are.
///
/// The plane between a and c lies (|q-c|^2 - |q-a|^2) / (2|a-c|) from q, which
/// is iq's dot(0.5*(a+c) - q, normalize(c-a)) written with the squared
/// distances the first pass already has.
///
/// Pass 1 is scan_uncached's 27-cell scan, keeping every candidate, so a is
/// the same point worleyFeature names (same arithmetic, same order, same
/// strict comparison). Pass 2 takes the planes against those 27 first. A
/// further point c can only own a nearer plane if |q-c| < F1 + 2*wall, since
/// |a-c| <= |q-c| + F1 bounds the plane's distance below by (|q-c| - F1) / 2,
/// so it then visits the cells of the window that radius reaches beyond the 27.
/// Most points need none of them. Over 200k random points this matched a
/// brute force over 9x9x9 cells exactly; planes against only the 27 overshot it
/// at 1 point in 12k, by up to 0.05 cell.
///
/// All arithmetic is relative to q, as in scan_uncached, so the hash keeps its
/// low bits at large |q|.
fn walls_uncached(tile: &[u8], q: [f32; 3], octaves: u32) -> Walls {
    let pi = [q[0].floor(), q[1].floor(), q[2].floor()];
    let pf = [q[0] - pi[0], q[1] - pi[1], q[2] - pi[2]];
    let away = |o: [f32; 3], h: [f32; 3]| {
        [pf[0] - (o[0] + h[0]), pf[1] - (o[1] + h[1]), pf[2] - (o[2] + h[2])]
    };
    let norm2 = |d: [f32; 3]| d[0] * d[0] + d[1] * d[1] + d[2] * d[2];

    // Pass 1: the nearest point, keeping every candidate's offset q - point.
    let mut d = [[0.0f32; 3]; CANDIDATES];
    let mut d2 = [0.0f32; CANDIDATES];
    let mut hs = [[0.0f32; 3]; CANDIDATES];
    let mut best = f32::INFINITY;
    let mut a = 0;
    for (i, o) in OFFSETS.iter().enumerate() {
        let h = noise::hash3_from_tile(tile, pi[0] + o[0], pi[1] + o[1], pi[2] + o[2], octaves);
        d[i] = away(*o, h);
        d2[i] = norm2(d[i]);
        hs[i] = h;
        if d2[i] < best {
            best = d2[i];
            a = i;
        }
    }

    // The distance from q to the plane between a and a point whose offset is dc.
    let da = d[a];
    let plane = |dc: [f32; 3], dc2: f32| {
        (dc2 - best) / (2.0 * norm2([dc[0] - da[0], dc[1] - da[1], dc[2] - da[2]]).sqrt())
    };

    // Pass 2a: the planes against the 27.
    let mut wall = f32::INFINITY;
    let mut wall_o = OFFSETS[a];
    let mut wall_h = hs[a];
    for i in 0..CANDIDATES {
        if i == a {
            continue;
        }
        let w = plane(d[i], d2[i]);
        if w < wall {
            wall = w;
            wall_o = OFFSETS[i];
            wall_h = hs[i];
        }
    }

    // Pass 2b: the cells beyond the 27 that the bound F1 + 2*wall reaches.
    let reach = best.sqrt() + 2.0 * wall;
    let lo = |k: usize| (pf[k] - reach).floor().max(-WALL_REACH);
    let hi = |k: usize| (pf[k] + reach).floor().min(WALL_REACH);
    let mut oz = lo(2);
    while oz <= hi(2) {
        let mut oy = lo(1);
        while oy <= hi(1) {
            let mut ox = lo(0);
            while ox <= hi(0) {
                let o = [ox, oy, oz];
                if o.iter().any(|v| v.abs() > 1.0) {
                    let h = noise::hash3_from_tile(tile, pi[0] + ox, pi[1] + oy, pi[2] + oz, octaves);
                    let dc = away(o, h);
                    let w = plane(dc, norm2(dc));
                    if w < wall {
                        wall = w;
                        wall_o = o;
                        wall_h = h;
                    }
                }
                ox += 1.0;
            }
            oy += 1.0;
        }
        oz += 1.0;
    }

    Walls {
        wall,
        wall_feature: [
            pi[0] + wall_o[0] + wall_h[0],
            pi[1] + wall_o[1] + wall_h[1],
            pi[2] + wall_o[2] + wall_h[2],
        ],
    }
}

// A one-entry memo. A call like noise.cellularCell wants five components of the
// SAME scan at the same point, and each is its own oracle leaf, so without this
// one evaluation would scan the neighbourhood five times. scan_uncached is pure,
// so reusing a result for identical arguments is always safe; a miss just does
// the work. Thread-local, so no locking and no sharing between eval threads.
thread_local! {
    static MEMO: RefCell<Option<([f32; 3], u32, usize, Scan)>> = const { RefCell::new(None) };
}

fn scan(tile: &Arc<Vec<u8>>, q: [f32; 3], octaves: u32) -> Scan {
    // The tile pointer is part of the key: two oracles registered against
    // different bakes must not read each other's cached answer.
    let tag = Arc::as_ptr(tile) as usize;
    MEMO.with(|m| {
        if let Some((cq, co, ct, s)) = *m.borrow() {
            // Bit equality, not approximate: the memo must never answer for a
            // different point.
            if ct == tag
                && co == octaves
                && cq[0].to_bits() == q[0].to_bits()
                && cq[1].to_bits() == q[1].to_bits()
                && cq[2].to_bits() == q[2].to_bits()
            {
                return s;
            }
        }
        let s = scan_uncached(tile, q, octaves);
        *m.borrow_mut() = Some((q, octaves, tag, s));
        s
    })
}

// The wall pass's own one-entry memo, on the same terms as MEMO: WALL and the
// three WALL_FEATURE axes are four leaves at the same point. Kept apart from
// MEMO so the F1 family never pays for a second pass it does not read.
thread_local! {
    static WALL_MEMO: RefCell<Option<([f32; 3], u32, usize, Walls)>> = const { RefCell::new(None) };
}

fn walls(tile: &Arc<Vec<u8>>, q: [f32; 3], octaves: u32) -> Walls {
    let tag = Arc::as_ptr(tile) as usize;
    WALL_MEMO.with(|m| {
        if let Some((cq, co, ct, w)) = *m.borrow() {
            if ct == tag
                && co == octaves
                && cq[0].to_bits() == q[0].to_bits()
                && cq[1].to_bits() == q[1].to_bits()
                && cq[2].to_bits() == q[2].to_bits()
            {
                return w;
            }
        }
        let w = walls_uncached(tile, q, octaves);
        *m.borrow_mut() = Some((q, octaves, tag, w));
        w
    })
}

/// True for the components the wall pass answers rather than the F1 scan.
fn is_wall_component(c: u32) -> bool {
    c >= component::WALL_FEATURE_X && c <= component::WALL
}

/// One component at q, from whichever pass yields it.
fn component_at(tile: &Arc<Vec<u8>>, q: [f32; 3], octaves: u32, c: u32) -> f32 {
    if is_wall_component(c) {
        walls(tile, q, octaves).component(c)
    } else {
        scan(tile, q, octaves).component(c)
    }
}

/// The oracle. Holds the tile it was registered with, so re-baking does not
/// disturb live trees — the same contract as the noise-tile oracle.
struct CellularOracle {
    tile: Arc<Vec<u8>>,
    comp: u32,
    octaves: u32,
}

impl CellularOracle {
    /// True for the components that are distances, and so 1-Lipschitz.
    fn is_distance(&self) -> bool {
        self.comp == component::F1 || self.comp == component::F2 || self.comp == component::WALL
    }
}

impl MeshOracleTrait for CellularOracle {
    fn point(&self, x: f32, y: f32, z: f32) -> f32 {
        component_at(&self.tile, [x, y, z], self.octaves, self.comp)
    }

    fn interval(&self, x: Interval, y: Interval, z: Interval) -> Interval {
        if !self.is_distance() {
            // Feature points and cell indices jump as q crosses a Voronoi wall,
            // so they are not Lipschitz and cannot be bracketed by sampling.
            // Bound them structurally instead: the winning cell is always one of
            // the 27 around floor(q), so its index on this axis lies within
            // [floor(lo) - 1, floor(hi) + 1], and a feature point sits that
            // index plus a jitter in [0, 1) — hence the +2. Sound for any box.
            // The wall feature can come from anywhere in the wall pass's window,
            // WALL_REACH cells out rather than 1.
            let reach = if is_wall_component(self.comp) { WALL_REACH } else { 1.0 };
            let a = [x, y, z][((self.comp - component::FEATURE_X) % 3) as usize];
            return Interval::new(a.lower().floor() - reach, a.upper().floor() + reach + 1.0);
        }
        // Sample the 8 corners of the query box and widen by half the diagonal.
        // F1, F2 and WALL are 1-Lipschitz, and every point of an axis-aligned box is
        // within half the diagonal of its nearest corner, so this brackets the
        // true range soundly. Half, not the full diagonal: the full one is a
        // wider-than-needed bound that makes the octree descend into cells with
        // no surface in them (the mesh oracle learned the same lesson).
        //
        // This is the whole point of these oracles. The 81-leaf form could only
        // report [0, 1) per hash leaf once a box straddled a cell boundary, so
        // nothing upstream could prune; here the bound tightens smoothly as the
        // box shrinks, at every depth.
        let xs = [x.lower(), x.upper()];
        let ys = [y.lower(), y.upper()];
        let zs = [z.lower(), z.upper()];
        let mut lo = f32::INFINITY;
        let mut hi = f32::NEG_INFINITY;
        for &cx in &xs {
            for &cy in &ys {
                for &cz in &zs {
                    let v = component_at(&self.tile, [cx, cy, cz], self.octaves, self.comp);
                    if v < lo {
                        lo = v;
                    }
                    if v > hi {
                        hi = v;
                    }
                }
            }
        }
        let dx = xs[1] - xs[0];
        let dy = ys[1] - ys[0];
        let dz = zs[1] - zs[0];
        let half_diag = 0.5 * (dx * dx + dy * dy + dz * dz).sqrt();
        // A distance is never negative however wide the box gets.
        Interval::new((lo - half_diag).max(0.0), hi + half_diag)
    }

    fn grad(&self, x: f32, y: f32, z: f32) -> Grad {
        if self.comp == component::WALL {
            // The wall's normal, pointing from the nearest wall into q's cell:
            // the unit vector from the wall feature c toward the nearest point
            // a. Exact wherever one wall is nearest; the distance has a ridge
            // where two are equally near, as F1 has across a wall.
            let w = walls(&self.tile, [x, y, z], self.octaves);
            let a = scan(&self.tile, [x, y, z], self.octaves).feature;
            let n = [a[0] - w.wall_feature[0], a[1] - w.wall_feature[1], a[2] - w.wall_feature[2]];
            let len = (n[0] * n[0] + n[1] * n[1] + n[2] * n[2]).sqrt();
            return Grad::new(w.wall, n[0] / len, n[1] / len, n[2] / len);
        }
        let s = scan(&self.tile, [x, y, z], self.octaves);
        let v = s.component(self.comp);
        if !self.is_distance() {
            // Piecewise constant: zero spatial gradient, as the per-cell-constant
            // noise hash gets.
            return Grad::new(v, 0.0, 0.0, 0.0);
        }
        // Analytic and exact for F1: the unit vector from the nearest feature
        // point toward q. Sharp across cell walls, where differentiating the
        // unrolled min could only pick a branch.
        //
        // F2 gets the same treatment about the second-nearest point. It is
        // correct wherever F2 is differentiable, which is everywhere except the
        // walls where the second-nearest changes — the same set where F1's own
        // gradient is undefined.
        // Both closest points come straight off the scan, so both gradients are
        // exact and neither costs an extra hash.
        let target = if self.comp == component::F1 {
            s.feature
        } else {
            s.far_feature
        };
        if v <= 0.0 {
            // Exactly on a feature point: a cone apex, where the gradient is
            // undefined. Report zero rather than a NaN direction.
            return Grad::new(v, 0.0, 0.0, 0.0);
        }
        Grad::new(
            v,
            (x - target[0]) / v,
            (y - target[1]) / v,
            (z - target[2]) / v,
        )
    }
}

// Oracle id layout, relative to CELLULAR_ORACLE_ID_BASE: component-major, then
// octave count. An id names both, so a tape stays self-describing without
// carrying the octaves literal anywhere in its body.
fn octave_slot(octaves: u32) -> u32 {
    if octaves <= 1 {
        0
    } else {
        1
    }
}

/// The oracle leaf id for one component at an octave count.
pub fn oracle_id(comp: u32, octaves: u32) -> u32 {
    CELLULAR_ORACLE_ID_BASE + comp * 2 + octave_slot(octaves)
}

/// If `id` names a cellular oracle leaf, return its (component, octaves).
pub fn leaf_of_id(id: u32) -> Option<(u32, u32)> {
    let slot = id.wrapping_sub(CELLULAR_ORACLE_ID_BASE);
    if slot >= component::COUNT * 2 {
        return None;
    }
    Some((slot / 2, if slot % 2 == 0 { 1 } else { 2 }))
}

/// Register the oracle for one component at `octaves` lookups against the active
/// tile, so `Tree::mesh_oracle(oracle_id(..))` resolves.
pub fn register(comp: u32, octaves: u32) -> Result<(), String> {
    if comp >= component::COUNT {
        return Err(format!("cellular component {} out of range", comp));
    }
    if octaves != 1 && octaves != 2 {
        return Err(format!("cellular octaves {} out of range", octaves));
    }
    mesh_oracle::register(
        oracle_id(comp, octaves),
        Arc::new(CellularOracle {
            tile: noise::tile_snapshot(),
            comp,
            octaves,
        }),
    );
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;

    fn oracle(comp: u32) -> CellularOracle {
        CellularOracle {
            tile: noise::tile_snapshot(),
            comp,
            octaves: 1,
        }
    }

    // F1 and F2 must both be 1-Lipschitz: that property is what makes
    // interval() sound, so check it directly rather than trusting the argument.
    #[test]
    fn distances_are_one_lipschitz() {
        let tile = noise::tile_snapshot();
        let pts = [
            [0.3, 0.7, 1.2],
            [4.9, -2.1, 0.05],
            [-7.5, 11.25, 3.75],
            [63.9, 64.1, -0.5],
            [0.31, 0.72, 1.19],
        ];
        for a in pts.iter() {
            for b in pts.iter() {
                let sa = scan_uncached(&tile, *a, 1);
                let sb = scan_uncached(&tile, *b, 1);
                let dist = ((a[0] - b[0]).powi(2) + (a[1] - b[1]).powi(2) + (a[2] - b[2]).powi(2))
                    .sqrt();
                for (va, vb, which) in [(sa.f1, sb.f1, "F1"), (sa.f2, sb.f2, "F2")] {
                    assert!(
                        (va - vb).abs() <= dist + 1e-4,
                        "{} not 1-Lipschitz: |{} - {}| > {} between {:?} and {:?}",
                        which,
                        va,
                        vb,
                        dist,
                        a,
                        b
                    );
                }
            }
        }
    }

    // F2 must never be less than F1, by definition.
    #[test]
    fn f2_is_at_least_f1() {
        let tile = noise::tile_snapshot();
        for i in 0..200 {
            let t = i as f32 * 0.37;
            let q = [t, -t * 0.5 + 1.3, t * 0.25 - 4.0];
            let s = scan_uncached(&tile, q, 1);
            assert!(s.f2 >= s.f1, "F2 {} < F1 {} at {:?}", s.f2, s.f1, q);
        }
    }

    // The interval must contain the value at every point of the box, for both
    // the Lipschitz components and the structurally-bounded ones.
    #[test]
    fn interval_brackets_every_component() {
        for comp in 0..component::COUNT {
            let o = oracle(comp);
            for &(lo, hi) in &[
                ([0.1f32, 0.1, 0.1], [0.9f32, 0.9, 0.9]),
                ([0.5f32, 0.5, 0.5], [1.5f32, 1.5, 1.5]),
                ([-2.0f32, 3.0, -1.0], [4.0f32, 9.0, 5.0]),
            ] {
                let iv = o.interval(
                    Interval::new(lo[0], hi[0]),
                    Interval::new(lo[1], hi[1]),
                    Interval::new(lo[2], hi[2]),
                );
                let n = 6;
                for i in 0..=n {
                    for j in 0..=n {
                        for k in 0..=n {
                            let t = |a: f32, b: f32, s: i32| a + (b - a) * (s as f32 / n as f32);
                            let p = [t(lo[0], hi[0], i), t(lo[1], hi[1], j), t(lo[2], hi[2], k)];
                            let v = o.point(p[0], p[1], p[2]);
                            assert!(
                                v >= iv.lower() - 1e-5 && v <= iv.upper() + 1e-5,
                                "component {}: {} outside [{}, {}] at {:?}",
                                comp,
                                v,
                                iv.lower(),
                                iv.upper(),
                                p
                            );
                        }
                    }
                }
            }
        }
    }

    // A small box must give a USEFUL interval for the distances, not the unit
    // interval the hash-leaf form fell back to. This is the regression these
    // oracles exist to prevent.
    #[test]
    fn distance_interval_is_tight_on_a_small_box() {
        for comp in [component::F1, component::F2] {
            let iv = oracle(comp).interval(
                Interval::new(0.40, 0.42),
                Interval::new(0.30, 0.32),
                Interval::new(0.70, 0.72),
            );
            assert!(
                iv.upper() - iv.lower() < 0.1,
                "component {}: interval [{}, {}] too wide for a 0.02 box",
                comp,
                iv.lower(),
                iv.upper()
            );
        }
    }

    // The cell a point belongs to must be the integer cell containing its
    // feature point, and must be one of the 27 scanned.
    #[test]
    fn near_cell_contains_the_feature_point() {
        let tile = noise::tile_snapshot();
        for i in 0..100 {
            let t = i as f32 * 0.53;
            let q = [t - 10.0, t * 0.5, -t * 0.25];
            let s = scan_uncached(&tile, q, 1);
            for axis in 0..3 {
                let inside = s.feature[axis] - s.near_cell[axis];
                assert!(
                    (0.0..1.0).contains(&inside),
                    "feature {:?} not inside cell {:?} on axis {}",
                    s.feature,
                    s.near_cell,
                    axis
                );
                let off = s.near_cell[axis] - q[axis].floor();
                assert!(
                    (-1.0..=1.0).contains(&off),
                    "near cell {:?} outside the 27 scanned around {:?}",
                    s.near_cell,
                    q
                );
            }
        }
    }

    // The analytic gradient of F1 must be a unit vector and agree with a central
    // difference.
    #[test]
    fn f1_grad_is_unit_and_matches_finite_difference() {
        let o = oracle(component::F1);
        for p in [[0.33f32, 0.61, 0.17], [2.7, -1.4, 5.2], [-3.1, 0.25, 8.8]] {
            let g = o.grad(p[0], p[1], p[2]);
            let len = (g.dx * g.dx + g.dy * g.dy + g.dz * g.dz).sqrt();
            assert!((len - 1.0).abs() < 1e-3, "gradient not unit: {}", len);

            let h = 1e-3;
            for (axis, got) in [(0, g.dx), (1, g.dy), (2, g.dz)] {
                let mut a = p;
                let mut b = p;
                a[axis] -= h;
                b[axis] += h;
                let want =
                    (o.point(b[0], b[1], b[2]) - o.point(a[0], a[1], a[2])) / (2.0 * h);
                assert!(
                    (got - want).abs() < 1e-2,
                    "axis {}: analytic {} vs finite difference {}",
                    axis,
                    got,
                    want
                );
            }
        }
    }

    // Knowing both closest points means knowing both gradients: F2 must be as
    // exact as F1, a unit vector away from the SECOND-nearest point. Checked
    // against a central difference, and against the stored far feature point.
    #[test]
    fn f2_grad_is_unit_and_points_away_from_the_far_feature() {
        let tile = noise::tile_snapshot();
        let o = oracle(component::F2);
        for p in [[0.33f32, 0.61, 0.17], [2.7, -1.4, 5.2], [-3.1, 0.25, 8.8]] {
            let g = o.grad(p[0], p[1], p[2]);
            let len = (g.dx * g.dx + g.dy * g.dy + g.dz * g.dz).sqrt();
            assert!((len - 1.0).abs() < 1e-3, "F2 gradient not unit: {}", len);

            // It must point away from the far feature point specifically.
            let s = scan_uncached(&tile, p, 1);
            let want = [
                (p[0] - s.far_feature[0]) / s.f2,
                (p[1] - s.far_feature[1]) / s.f2,
                (p[2] - s.far_feature[2]) / s.f2,
            ];
            for (axis, (got, w)) in [(0, (g.dx, want[0])), (1, (g.dy, want[1])), (2, (g.dz, want[2]))]
            {
                assert!(
                    (got - w).abs() < 1e-4,
                    "axis {}: F2 gradient {} is not away from the far feature ({})",
                    axis,
                    got,
                    w
                );
            }

            let h = 1e-3;
            for (axis, got) in [(0, g.dx), (1, g.dy), (2, g.dz)] {
                let mut a = p;
                let mut b = p;
                a[axis] -= h;
                b[axis] += h;
                let want = (o.point(b[0], b[1], b[2]) - o.point(a[0], a[1], a[2])) / (2.0 * h);
                assert!(
                    (got - want).abs() < 1e-2,
                    "axis {}: F2 analytic {} vs finite difference {}",
                    axis,
                    got,
                    want
                );
            }
        }
    }

    // The far feature point must lie in the far cell, the way the near one lies
    // in the near cell.
    #[test]
    fn far_feature_lies_in_the_far_cell() {
        let tile = noise::tile_snapshot();
        for i in 0..100 {
            let t = i as f32 * 0.41;
            let q = [t - 5.0, -t * 0.3, t * 0.7 - 2.0];
            let s = scan_uncached(&tile, q, 1);
            for axis in 0..3 {
                let inside = s.far_feature[axis] - s.far_cell[axis];
                assert!(
                    (0.0..1.0).contains(&inside),
                    "far feature {:?} not inside far cell {:?} on axis {}",
                    s.far_feature,
                    s.far_cell,
                    axis
                );
            }
        }
    }

    // The wall distance by brute force: the nearest point over a 9x9x9 block,
    // then the nearest plane against every other point in it. Four cells out
    // is far past anything the two-pass scan could miss.
    fn brute_wall(tile: &[u8], q: [f32; 3]) -> f32 {
        let pi = [q[0].floor(), q[1].floor(), q[2].floor()];
        let mut pts = Vec::new();
        for z in -4..=4 {
            for y in -4..=4 {
                for x in -4..=4 {
                    let c = [pi[0] + x as f32, pi[1] + y as f32, pi[2] + z as f32];
                    let h = noise::hash3_from_tile(tile, c[0], c[1], c[2], 1);
                    pts.push([
                        (c[0] + h[0] - q[0]) as f64,
                        (c[1] + h[1] - q[1]) as f64,
                        (c[2] + h[2] - q[2]) as f64,
                    ]);
                }
            }
        }
        let n2 = |v: [f64; 3]| v[0] * v[0] + v[1] * v[1] + v[2] * v[2];
        let a = (0..pts.len())
            .min_by(|&i, &j| n2(pts[i]).partial_cmp(&n2(pts[j])).unwrap())
            .unwrap();
        let mut wall = f64::INFINITY;
        for (i, c) in pts.iter().enumerate() {
            if i != a {
                let ac = [c[0] - pts[a][0], c[1] - pts[a][1], c[2] - pts[a][2]];
                wall = wall.min((n2(*c) - n2(pts[a])) / (2.0 * n2(ac).sqrt()));
            }
        }
        wall as f32
    }

    fn sample_points(n: usize) -> Vec<[f32; 3]> {
        (0..n)
            .map(|i| {
                let t = i as f32;
                [
                    (t * 0.618_034).fract() * 128.0 - 64.0,
                    (t * 0.414_214).fract() * 128.0 - 64.0,
                    (t * 0.732_051).fract() * 128.0 - 64.0,
                ]
            })
            .collect()
    }

    // The two-pass wall distance must be the exact distance to the nearest
    // wall, which a brute force over a much larger block decides.
    #[test]
    fn wall_matches_brute_force() {
        let tile = noise::tile_snapshot();
        for q in sample_points(4000) {
            let got = walls_uncached(&tile, q, 1).wall;
            let want = brute_wall(&tile, q);
            assert!(
                (got - want).abs() < 1e-4,
                "wall {} but brute force {} at {:?}",
                got,
                want,
                q
            );
        }
    }

    // The wall is a distance to a set, so 1-Lipschitz. It is also at least
    // (F2 - F1)/2: on the nearest wall F1 = F2, and the d it takes to get there
    // changes each of them by at most d.
    #[test]
    fn wall_is_one_lipschitz_and_bounded_by_f2_minus_f1() {
        let tile = noise::tile_snapshot();
        let pts = sample_points(300);
        for (i, a) in pts.iter().enumerate() {
            let wa = walls_uncached(&tile, *a, 1).wall;
            let s = scan_uncached(&tile, *a, 1);
            assert!(wa >= 0.0, "negative wall {} at {:?}", wa, a);
            assert!(
                wa >= 0.5 * (s.f2 - s.f1) - 1e-5,
                "wall {} below (F2 - F1)/2 = {} at {:?}",
                wa,
                0.5 * (s.f2 - s.f1),
                a
            );
            // A near neighbour as well as the far ones, so the bound is tested
            // where it is tight.
            let near = [a[0] + 0.013, a[1] - 0.021, a[2] + 0.017];
            for b in pts[i + 1..].iter().take(20).chain(std::iter::once(&near)) {
                let wb = walls_uncached(&tile, *b, 1).wall;
                let dist = ((a[0] - b[0]).powi(2) + (a[1] - b[1]).powi(2) + (a[2] - b[2]).powi(2)).sqrt();
                assert!(
                    (wa - wb).abs() <= dist + 1e-4,
                    "wall not 1-Lipschitz: |{} - {}| > {} between {:?} and {:?}",
                    wa,
                    wb,
                    dist,
                    a,
                    b
                );
            }
        }
    }

    // The wall's gradient must be a unit vector and agree with a central
    // difference wherever one wall is clearly nearest.
    #[test]
    fn wall_grad_is_unit_and_matches_finite_difference() {
        let o = oracle(component::WALL);
        let mut checked = 0;
        for p in sample_points(200) {
            let g = o.grad(p[0], p[1], p[2]);
            let len = (g.dx * g.dx + g.dy * g.dy + g.dz * g.dz).sqrt();
            assert!((len - 1.0).abs() < 1e-3, "wall gradient not unit: {}", len);
            let h = 1e-3;
            let fd: Vec<f32> = (0..3)
                .map(|axis| {
                    let mut a = p;
                    let mut b = p;
                    a[axis] -= h;
                    b[axis] += h;
                    (o.point(b[0], b[1], b[2]) - o.point(a[0], a[1], a[2])) / (2.0 * h)
                })
                .collect();
            // Skip points within a step of a ridge, where two walls tie and
            // the difference straddles both.
            let fd_len = (fd[0] * fd[0] + fd[1] * fd[1] + fd[2] * fd[2]).sqrt();
            if (fd_len - 1.0).abs() > 1e-2 {
                continue;
            }
            for (axis, got) in [(0, g.dx), (1, g.dy), (2, g.dz)] {
                assert!(
                    (got - fd[axis]).abs() < 1e-2,
                    "axis {}: wall gradient {} vs finite difference {} at {:?}",
                    axis,
                    got,
                    fd[axis],
                    p
                );
            }
            checked += 1;
        }
        assert!(checked > 150, "only {} points away from a ridge", checked);
    }

    // The memo must never answer for a different point or a different octave
    // count — a stale hit would be a silently wrong field.
    #[test]
    fn memo_does_not_confuse_points_or_octaves() {
        let tile = noise::tile_snapshot();
        let a = [1.25f32, 2.5, -3.75];
        let b = [1.26f32, 2.5, -3.75];
        for oct in [1u32, 2] {
            // Prime the memo with a, then ask for b, then a again.
            let sa1 = scan(&tile, a, oct).f1;
            let sb = scan(&tile, b, oct).f1;
            let sa2 = scan(&tile, a, oct).f1;
            assert_eq!(sa1, sa2, "memo lost a value it should have kept");
            assert_eq!(sa1, scan_uncached(&tile, a, oct).f1);
            assert_eq!(sb, scan_uncached(&tile, b, oct).f1);
        }
        // Same point, different octaves, must not share an entry.
        assert_eq!(scan(&tile, a, 1).f1, scan_uncached(&tile, a, 1).f1);
        assert_eq!(scan(&tile, a, 2).f1, scan_uncached(&tile, a, 2).f1);
        assert_eq!(scan(&tile, a, 1).f1, scan_uncached(&tile, a, 1).f1);
    }
}
