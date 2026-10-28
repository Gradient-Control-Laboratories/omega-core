//! hodge's Gabor spinodoid (hodge-release spinodoid.py, spinodoid_gabor) as
//! fidget oracles over a kernel table the host bakes.
//!
//! hodge generates kernels on demand from an integer hash and a numpy-drawn
//! direction pool, and samples its scale and orientation drivers at each kernel
//! centre. The Go baker (common/noise/spinodoid, hodge.go) reproduces all of
//! that exactly and hands the result here as a table: per kernel its jitter
//! inside the cell, its wave direction, wavenumber and phase. What is left is
//! float work, done in f64 so a point matches the Go reference's Sample -- and
//! so a part this builds is one hodge can rebuild.
//!
//! The table holds the cells [-TABLE_CELLS/2, TABLE_CELLS/2)^3 and wraps
//! outside them; a lookup rebuilds each kernel's centre from the query's
//! unwrapped cell, so the wrap is seamless. Unlike the lattice table beside it
//! (super), it is always periodic and never sized to a domain: the Go baker
//! refuses a driven domain that would not fit, so a driver cannot repeat here
//! unnoticed.
//!
//! Its own slots and oracle-leaf block are separate from the lattice's, so a
//! document may read both fields. The parent's interval helpers are shared.

use std::cell::RefCell;
use std::collections::HashMap;
use std::f64::consts::PI;
use std::sync::{Arc, Mutex, OnceLock};

use fidget_core::{
    mesh_oracle::{self, MeshOracle as MeshOracleTrait},
    types::{Grad, Interval},
};

use super::{component, distance_envelope, next_down, next_up, MAX_SLOTS};

/// Reserved fidget mesh-oracle id block for hodge's tables, one 0x1000 step
/// above the lattice's (super::SPINODOID_ORACLE_ID_BASE) and clear of the
/// noise tile's and the cellular oracles' blocks.
pub const HODGE_ORACLE_ID_BASE: u32 = 0xE0_5B2_000;

/// The table's edge in kernel cells; common/noise/spinodoid.HodgeTableCells.
pub const TABLE_CELLS: i64 = 32;
const TABLE_ORIGIN: i64 = -TABLE_CELLS / 2;

/// f64s per kernel: jitter xyz, direction xyz, wavenumber, phase.
pub const KERNEL_STRIDE: usize = 8;

/// One baked table.
pub struct Table {
    kernels: Vec<f64>,
    origin: [i64; 3],
    size: [usize; 3],
    bounds: [f64; 6],
    periodic: bool,
    impulses: usize,
    cell_size: f64,
    a: f64,
    sigma: f64,
    base_beta: f64,
    level: f64,
    local: bool,
    beta_min: f64,
    beta_max: f64,
}

impl Table {
    #[allow(clippy::too_many_arguments)]
    pub fn new(
        kernels: Vec<f64>,
        origin: [i64; 3],
        size: [usize; 3],
        bounds: [f64; 6],
        periodic: bool,
        impulses: usize,
        cell_size: f64,
        a: f64,
        sigma: f64,
        base_beta: f64,
        level: f64,
        local: bool,
    ) -> Result<Table, String> {
        let cells = match size
            .iter()
            .try_fold(1usize, |n, &s| if s == 0 { None } else { n.checked_mul(s) })
        {
            Some(n) => n,
            None => return Err("hodge spinodoid table has an empty or overflowing size".into()),
        };
        if impulses == 0 || kernels.len() != cells * impulses * KERNEL_STRIDE {
            return Err(format!(
                "hodge spinodoid table has {} values, want {} cells x {} impulses x {}",
                kernels.len(),
                cells,
                impulses,
                KERNEL_STRIDE
            ));
        }
        // A periodic table is hodge's own window and nothing else; a bounded
        // one must say where it is.
        if periodic && (size != [TABLE_CELLS as usize; 3] || origin != [TABLE_ORIGIN; 3]) {
            return Err("a periodic hodge spinodoid table must be the 32-cubed window".into());
        }
        if !periodic
            && (0..3).any(|a| {
                !bounds[a].is_finite() || !bounds[a + 3].is_finite() || bounds[a] >= bounds[a + 3]
            })
        {
            return Err("invalid hodge spinodoid domain bounds".into());
        }
        if !(cell_size > 0.0 && a > 0.0 && sigma > 0.0 && base_beta > 0.0) {
            return Err("spinodoid table constants must be positive".into());
        }
        let mut beta_min = f64::INFINITY;
        let mut beta_max = f64::NEG_INFINITY;
        for k in kernels.chunks_exact(KERNEL_STRIDE) {
            beta_min = beta_min.min(k[6]);
            beta_max = beta_max.max(k[6]);
        }
        Ok(Table {
            kernels,
            origin,
            size,
            bounds,
            periodic,
            impulses,
            cell_size,
            a,
            sigma,
            base_beta,
            level,
            local,
            beta_min,
            beta_max,
        })
    }

    fn wrap(c: i64) -> usize {
        (c - TABLE_ORIGIN).rem_euclid(TABLE_CELLS) as usize
    }

    /// The kernels of the unwrapped cell c, or None where a bounded table does
    /// not hold it. A periodic table wraps, which is what makes a stationary
    /// field tile; a bounded one hashed the true cell index, so it is hodge's
    /// everywhere inside its domain and holds nothing outside it.
    fn cell(&self, c: [i64; 3]) -> Option<&[f64]> {
        let mut local = [0usize; 3];
        for a in 0..3 {
            if self.periodic {
                local[a] = Self::wrap(c[a]);
                continue;
            }
            let index = c[a] - self.origin[a];
            if index < 0 || index >= self.size[a] as i64 {
                return None;
            }
            local[a] = index as usize;
        }
        let idx = (local[2] * self.size[1] + local[1]) * self.size[0] + local[0];
        let n = self.impulses * KERNEL_STRIDE;
        Some(&self.kernels[idx * n..(idx + 1) * n])
    }

    /// Whether q is inside the table's domain (always, when periodic).
    pub fn contains(&self, q: [f64; 3]) -> bool {
        self.periodic || (0..3).all(|a| q[a] >= self.bounds[a] && q[a] <= self.bounds[a + 3])
    }

    /// Distance to the finite domain and its outward unit gradient, the same
    /// conservative extension the lattice table carries: a ray marcher can
    /// reach the part instead of stepping over it.
    pub fn outside(&self, q: [f64; 3]) -> (f64, [f64; 3]) {
        let mut delta = [0.0; 3];
        let mut squared = 0.0;
        for a in 0..3 {
            delta[a] = q[a] - q[a].clamp(self.bounds[a], self.bounds[a + 3]);
            squared += delta[a] * delta[a];
        }
        let distance = squared.sqrt();
        if distance > 0.0 {
            for v in &mut delta {
                *v /= distance;
            }
        }
        (distance, delta)
    }

    /// True for a bounded table, which the oracle's interval has to treat
    /// differently outside the domain.
    pub fn is_periodic(&self) -> bool {
        self.periodic
    }

    /// The domain, as [min xyz, max xyz]; meaningless when periodic.
    pub fn domain(&self) -> [f64; 6] {
        self.bounds
    }

    fn cell_of(&self, v: f64) -> i64 {
        (v / self.cell_size).floor() as i64
    }

    /// The pseudo-distance and its gradient at q.
    ///
    /// d = (phi - level) / beta, with phi the unit-variance kernel sum and beta
    /// the envelope-energy-weighted wavenumber (local) or the base one. The
    /// neighbourhood is hodge's: the 27 cells around floor(q / cell), last axis
    /// fastest.
    pub fn eval(&self, q: [f64; 3]) -> (f64, [f64; 3]) {
        if !self.contains(q) {
            return self.outside(q);
        }
        let cell = [self.cell_of(q[0]), self.cell_of(q[1]), self.cell_of(q[2])];
        let a2 = self.a * self.a;
        let mut acc = 0.0;
        let mut g_acc = [0.0; 3];
        let mut w_acc = 0.0;
        let mut bw_acc = 0.0;
        let mut gw = [0.0; 3];
        let mut gbw = [0.0; 3];
        for i in -1..=1 {
            for j in -1..=1 {
                for k in -1..=1 {
                    let c = [cell[0] + i, cell[1] + j, cell[2] + k];
                    let Some(cell_kernels) = self.cell(c) else {
                        continue;
                    };
                    for kn in cell_kernels.chunks_exact(KERNEL_STRIDE) {
                        let r = [
                            q[0] - (c[0] as f64 + kn[0]) * self.cell_size,
                            q[1] - (c[1] as f64 + kn[1]) * self.cell_size,
                            q[2] - (c[2] as f64 + kn[2]) * self.cell_size,
                        ];
                        let d2 = r[0] * r[0] + r[1] * r[1] + r[2] * r[2];
                        let env = (-PI * a2 * d2).exp();
                        let beta = kn[6];
                        let phase = beta * (r[0] * kn[3] + r[1] * kn[4] + r[2] * kn[5]) + kn[7];
                        let (s, co) = phase.sin_cos();
                        acc += env * co;
                        let w = env * env;
                        w_acc += w;
                        bw_acc += w * beta;
                        for ax in 0..3 {
                            // d/dx [env cos] = -2 pi a^2 r env cos - env sin beta n
                            g_acc[ax] += env * (-2.0 * PI * a2 * r[ax] * co - s * beta * kn[3 + ax]);
                            // d/dx env^2 = -4 pi a^2 r env^2
                            let dw = -4.0 * PI * a2 * r[ax] * w;
                            gw[ax] += dw;
                            gbw[ax] += dw * beta;
                        }
                    }
                }
            }
        }
        let phi = acc / self.sigma;
        let gphi = [g_acc[0] / self.sigma, g_acc[1] / self.sigma, g_acc[2] / self.sigma];
        let mut beta = self.base_beta;
        let mut gbeta = [0.0; 3];
        if self.local {
            let b = bw_acc / w_acc.max(1e-300);
            if b > 0.0 {
                beta = b;
                for ax in 0..3 {
                    gbeta[ax] = (gbw[ax] - b * gw[ax]) / w_acc;
                }
            }
        }
        let num = phi - self.level;
        let d = num / beta;
        let g = [
            gphi[0] / beta - num * gbeta[0] / (beta * beta),
            gphi[1] / beta - num * gbeta[1] / (beta * beta),
            gphi[2] / beta - num * gbeta[2] / (beta * beta),
        ];
        (d, g)
    }

    /// Distance from the box [lo, hi] to point c (0 inside it).
    fn box_distance(lo: [f64; 3], hi: [f64; 3], c: [f64; 3]) -> (f64, f64) {
        let mut near = 0.0;
        let mut far = 0.0;
        for ax in 0..3 {
            let below = lo[ax] - c[ax];
            let above = c[ax] - hi[ax];
            let n = below.max(above).max(0.0);
            let f = (c[ax] - lo[ax]).abs().max((hi[ax] - c[ax]).abs());
            near += n * n;
            far += f * f;
        }
        (near.sqrt(), far.sqrt())
    }

    /// A conservative bound on d over the box [lo, hi], and on |grad d| when
    /// the box sits inside one kernel cell (None otherwise, where the field's
    /// kernel set changes and it is not continuous).
    ///
    /// phi is bounded by its envelopes: every kernel any point of the box can
    /// see lies in the cells within one of the box's cell range, and each
    /// contributes at most exp(-pi a^2 near^2) in magnitude (or nothing, where
    /// it falls out of a point's neighbourhood). The local wavenumber is an
    /// energy-weighted mean of those kernels' wavenumbers, so it lies between
    /// their extremes.
    fn bound(&self, lo: [f64; 3], hi: [f64; 3]) -> (f64, f64, f64, Option<f64>) {
        let a2 = self.a * self.a;
        let c_lo = [self.cell_of(lo[0]), self.cell_of(lo[1]), self.cell_of(lo[2])];
        let c_hi = [self.cell_of(hi[0]), self.cell_of(hi[1]), self.cell_of(hi[2])];
        let one_cell = c_lo == c_hi;
        let mut env_sum = 0.0;
        let mut grad_sum = 0.0;
        let mut r_far_max: f64 = 0.0;
        let mut b_lo = f64::INFINITY;
        let mut b_hi = f64::NEG_INFINITY;
        // A box spanning the whole table sees every kernel; past that the
        // per-kernel walk only repeats the same kernels, so cap the walk and
        // fall back to the table-wide extremes.
        let span = (0..3).map(|ax| c_hi[ax] - c_lo[ax] + 3).product::<i64>();
        if span > TABLE_CELLS * TABLE_CELLS * TABLE_CELLS {
            return (f64::INFINITY, self.beta_min, self.beta_max, None);
        }
        for cx in c_lo[0] - 1..=c_hi[0] + 1 {
            for cy in c_lo[1] - 1..=c_hi[1] + 1 {
                for cz in c_lo[2] - 1..=c_hi[2] + 1 {
                    let c = [cx, cy, cz];
                    let Some(cell_kernels) = self.cell(c) else {
                        continue;
                    };
                    for kn in cell_kernels.chunks_exact(KERNEL_STRIDE) {
                        let centre = [
                            (c[0] as f64 + kn[0]) * self.cell_size,
                            (c[1] as f64 + kn[1]) * self.cell_size,
                            (c[2] as f64 + kn[2]) * self.cell_size,
                        ];
                        let (near, far) = Self::box_distance(lo, hi, centre);
                        let env = (-PI * a2 * near * near).exp();
                        env_sum += env;
                        b_lo = b_lo.min(kn[6]);
                        b_hi = b_hi.max(kn[6]);
                        if one_cell {
                            // |grad(env cos)| <= env (2 pi a^2 r + beta), and
                            // env r is largest where r = 1 / (a sqrt(2 pi)).
                            let r_peak = 1.0 / (self.a * (2.0 * PI).sqrt());
                            let r_at = r_peak.clamp(near, far);
                            let peak = (-PI * a2 * r_at * r_at).exp() * 2.0 * PI * a2 * r_at;
                            grad_sum += peak + env * kn[6];
                            r_far_max = r_far_max.max(far);
                        }
                    }
                }
            }
        }
        let phi_max = env_sum / self.sigma;
        if !self.local {
            b_lo = self.base_beta;
            b_hi = self.base_beta;
        } else {
            // Where no kernel's energy registers, hodge falls back to the base
            // wavenumber.
            b_lo = b_lo.min(self.base_beta);
            b_hi = b_hi.max(self.base_beta);
        }
        let lip = if one_cell {
            let g_phi = grad_sum / self.sigma;
            // |grad beta_local| <= 4 pi a^2 r_far (beta_max - beta_min).
            let g_beta = if self.local { 4.0 * PI * a2 * r_far_max * (b_hi - b_lo) } else { 0.0 };
            Some(g_phi / b_lo + (phi_max + self.level.abs()) * g_beta / (b_lo * b_lo))
        } else {
            None
        };
        (phi_max, b_lo, b_hi, lip)
    }
}

impl Table {
    /// The level the field is thresholded at.
    pub fn level(&self) -> f64 {
        self.level
    }
}

// One-entry memo: the four leaves of a slot ask for the same evaluation at the
// same point. Pure, so reuse for bit-identical arguments is always safe.
thread_local! {
    static MEMO: RefCell<Option<([u32; 3], usize, (f64, [f64; 3]))>> = const { RefCell::new(None) };
}

fn eval_memo(table: &Arc<Table>, x: f32, y: f32, z: f32) -> (f64, [f64; 3]) {
    let tag = Arc::as_ptr(table) as usize;
    let key = [x.to_bits(), y.to_bits(), z.to_bits()];
    MEMO.with(|m| {
        if let Some((k, t, v)) = *m.borrow() {
            if k == key && t == tag {
                return v;
            }
        }
        let v = table.eval([x as f64, y as f64, z as f64]);
        *m.borrow_mut() = Some((key, tag, v));
        v
    })
}

struct HodgeOracle {
    table: Arc<Table>,
    comp: u32,
}

impl MeshOracleTrait for HodgeOracle {
    fn point(&self, x: f32, y: f32, z: f32) -> f32 {
        let (d, g) = eval_memo(&self.table, x, y, z);
        match self.comp {
            component::DISTANCE => d as f32,
            c => g[(c - component::GRAD_X) as usize] as f32,
        }
    }

    fn interval(&self, x: Interval, y: Interval, z: Interval) -> Interval {
        let lo = [x.lower() as f64, y.lower() as f64, z.lower() as f64];
        let hi = [x.upper() as f64, y.upper() as f64, z.upper() as f64];
        if !self.table.periodic {
            // Wholly outside the domain the field is the distance to it, whose
            // interval over the box is [nearest, farthest]. Mirrors the lattice
            // oracle, so a bounded field of either kind marches the same.
            if (0..3).any(|a| hi[a] < self.table.bounds[a] || lo[a] > self.table.bounds[a + 3]) {
                if self.comp != component::DISTANCE {
                    return Interval::new(-1.0, 1.0);
                }
                let mut lower_squared = 0.0;
                let mut upper_squared = 0.0;
                for a in 0..3 {
                    let min = self.table.bounds[a];
                    let max = self.table.bounds[a + 3];
                    let near = (min - hi[a]).max(lo[a] - max).max(0.0);
                    let far = (lo[a] - lo[a].clamp(min, max))
                        .abs()
                        .max((hi[a] - hi[a].clamp(min, max)).abs());
                    lower_squared += near * near;
                    upper_squared += far * far;
                }
                return Interval::new(
                    next_down(lower_squared.sqrt() as f32),
                    next_up(upper_squared.sqrt() as f32),
                );
            }
            // A box straddling the boundary has the extension on one side and
            // the field on the other, so no interior bound covers it.
            if !self.table.contains(lo) || !self.table.contains(hi) {
                return Interval::new(f32::NEG_INFINITY, f32::INFINITY);
            }
        }
        let (phi_max, b_lo, b_hi, lip) = self.table.bound(lo, hi);
        if self.comp != component::DISTANCE {
            // The gradient is only asked for by a .dp chain; bound it by the
            // Lipschitz constant where there is one, and leave it open where
            // the kernel set changes.
            return match lip {
                Some(l) => Interval::new(-(l as f32), l as f32),
                None => Interval::new(f32::NEG_INFINITY, f32::INFINITY),
            };
        }
        let (mut out_lo, mut out_hi) = distance_envelope(phi_max, self.table.level, b_lo, b_hi);
        if let Some(l) = lip {
            // Inside one cell the field is smooth: the centre value plus the
            // Lipschitz bound over half the diagonal, intersected with the
            // envelope.
            let c = [(lo[0] + hi[0]) * 0.5, (lo[1] + hi[1]) * 0.5, (lo[2] + hi[2]) * 0.5];
            let (dc, _) = self.table.eval(c);
            let half_diag = 0.5
                * ((hi[0] - lo[0]).powi(2) + (hi[1] - lo[1]).powi(2) + (hi[2] - lo[2]).powi(2)).sqrt();
            out_lo = out_lo.max(dc - l * half_diag);
            out_hi = out_hi.min(dc + l * half_diag);
        }
        // Round outward: the f64 bound must still hold as f32.
        Interval::new(
            next_down(out_lo as f32),
            next_up(out_hi as f32),
        )
    }

    fn grad(&self, x: f32, y: f32, z: f32) -> Grad {
        let (d, g) = eval_memo(&self.table, x, y, z);
        if self.comp == component::DISTANCE {
            return Grad::new(d as f32, g[0] as f32, g[1] as f32, g[2] as f32);
        }
        // The gradient's own derivative (a Hessian row) is not carried.
        Grad::new(g[(self.comp - component::GRAD_X) as usize] as f32, 0.0, 0.0, 0.0)
    }
}

fn tables() -> &'static Mutex<HashMap<u32, Arc<Table>>> {
    static TABLES: OnceLock<Mutex<HashMap<u32, Arc<Table>>>> = OnceLock::new();
    TABLES.get_or_init(|| Mutex::new(HashMap::new()))
}

/// The oracle leaf id for one component of the table in `slot`.
pub fn oracle_id(slot: u32, comp: u32) -> u32 {
    HODGE_ORACLE_ID_BASE + slot * component::COUNT + comp
}

/// The (slot, component) a hodge spinodoid leaf id names, or None if the id is
/// not one of ours — the inverse of [`oracle_id`], for a caller inspecting a
/// tape's oracle leaves.
pub fn leaf_of_id(id: u32) -> Option<(u32, u32)> {
    let off = id.wrapping_sub(HODGE_ORACLE_ID_BASE);
    (off < MAX_SLOTS * component::COUNT).then(|| (off / component::COUNT, off % component::COUNT))
}

/// Install a hodge table in `slot` and register its component oracles. Re-registering
/// a slot replaces it; live trees keep the table they were built against.
pub fn register(slot: u32, table: Table) -> Result<(), String> {
    if slot >= MAX_SLOTS {
        return Err(format!("hodge spinodoid slot {} out of range", slot));
    }
    let table = Arc::new(table);
    for comp in 0..component::COUNT {
        mesh_oracle::register(
            oracle_id(slot, comp),
            Arc::new(HodgeOracle {
                table: table.clone(),
                comp,
            }),
        );
    }
    tables().lock().unwrap().insert(slot, table);
    Ok(())
}

/// True once `slot` holds a table.
pub fn has(slot: u32) -> bool {
    tables().lock().unwrap().contains_key(&slot)
}
