//! The spinodoid (lattice Gabor; see common/noise/spinodoid) as fidget
//! oracles over a table the host bakes: one tapered Gabor kernel per lattice
//! site, with its direction, wavenumber and phase. Evaluated in f64, so a point
//! matches the Go reference's Sample.
//!
//! Each registered table is a slot; a slot has COMPONENTS oracle leaves, the
//! pseudo-distance d and the three components of its analytic gradient, which
//! share one memoized evaluation.

use std::cell::RefCell;
use std::collections::HashMap;
use std::f64::consts::PI;
use std::sync::{
    atomic::{AtomicU64, Ordering},
    Arc, Mutex, OnceLock,
};

use fidget_core::{
    mesh_oracle::{self, MeshOracle as MeshOracleTrait},
    types::{Grad, Interval},
};

/// hodge's own field, over its own slots and leaf ids (hodge.rs). omega's
/// builds bind only the lattice here; hodge's is for a tape reader that
/// registers hodge's kernels itself (tape.rs register_hodge_spinodoid).
pub mod hodge;

/// Both tables baked from a field's arguments, for a tape reader that rebuilds
/// a spinodoid from its payload.
pub mod bake;
mod numpy;
mod ziggurat_tables;

/// Reserved fidget mesh-oracle id block for spinodoid tables, clear of the
/// noise tile's and the cellular oracles' blocks.
pub const SPINODOID_ORACLE_ID_BASE: u32 = 0xE0_5B1_000;

/// The table's edge in lattice sites; common/noise/spinodoid.TableSites.
pub const TABLE_SITES: i64 = 64;
const TABLE_ORIGIN: i64 = -TABLE_SITES / 2;
const MAX_TABLE_SITES: usize = 1 << 20;

/// f64s per site: direction xyz, wavenumber, phase.
pub const SITE_STRIDE: usize = 5;

/// What a leaf returns: the pseudo-distance, or one gradient component.
pub mod component {
    pub const DISTANCE: u32 = 0;
    pub const GRAD_X: u32 = 1;
    pub const COUNT: u32 = 4;
}

/// How many tables can be registered at once.
pub const MAX_SLOTS: u32 = 256;

static NEXT_TABLE_ID: AtomicU64 = AtomicU64::new(1);

/// One baked table: a lattice of kernel sites at `spacing`, each a tapered
/// Gabor kernel reaching `reach` (see common/noise/spinodoid).
pub struct Table {
    id: u64,
    sites: Vec<f64>,
    origin: [i64; 3],
    size: [usize; 3],
    bounds: [f64; 6],
    periodic: bool,
    spacing: f64,
    reach: f64,
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
        sites: Vec<f64>,
        origin: [i64; 3],
        size: [usize; 3],
        bounds: [f64; 6],
        periodic: bool,
        spacing: f64,
        reach: f64,
        a: f64,
        sigma: f64,
        base_beta: f64,
        level: f64,
        local: bool,
    ) -> Result<Table, String> {
        let count = size
            .iter()
            .try_fold(1usize, |n, &s| if s == 0 { None } else { n.checked_mul(s) })
            .ok_or("invalid spinodoid table dimensions")?;
        if count > MAX_TABLE_SITES
            || (periodic && (size != [TABLE_SITES as usize; 3] || origin != [TABLE_ORIGIN; 3]))
        {
            return Err("spinodoid table dimensions exceed the supported layout".into());
        }
        if !periodic
            && (0..3).any(|a| {
                !bounds[a].is_finite() || !bounds[a + 3].is_finite() || bounds[a] >= bounds[a + 3]
            })
        {
            return Err("invalid spinodoid domain bounds".into());
        }
        if sites.len() != count * SITE_STRIDE {
            return Err(format!(
                "spinodoid table has {} values, want {} sites x {}",
                sites.len(),
                count,
                SITE_STRIDE
            ));
        }
        if !(spacing > 0.0 && reach > 0.0 && a > 0.0 && sigma > 0.0 && base_beta > 0.0) {
            return Err("spinodoid table constants must be positive".into());
        }
        let mut beta_min = f64::INFINITY;
        let mut beta_max = f64::NEG_INFINITY;
        for s in sites.chunks_exact(SITE_STRIDE) {
            beta_min = beta_min.min(s[3]);
            beta_max = beta_max.max(s[3]);
        }
        Ok(Table {
            id: NEXT_TABLE_ID.fetch_add(1, Ordering::Relaxed),
            sites,
            origin,
            size,
            bounds,
            periodic,
            spacing,
            reach,
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
        (c - TABLE_ORIGIN).rem_euclid(TABLE_SITES) as usize
    }

    fn contains(&self, q: [f64; 3]) -> bool {
        self.periodic || (0..3).all(|a| q[a] >= self.bounds[a] && q[a] <= self.bounds[a + 3])
    }

    fn outside(&self, q: [f64; 3]) -> (f64, [f64; 3]) {
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

    fn site(&self, x: i64, y: i64, z: i64) -> Option<&[f64]> {
        let coords = [x, y, z];
        let mut local = [0usize; 3];
        for a in 0..3 {
            if self.periodic {
                local[a] = Self::wrap(coords[a]);
            } else {
                let index = coords[a] - self.origin[a];
                if index < 0 || index >= self.size[a] as i64 {
                    return None;
                }
                local[a] = index as usize;
            }
        }
        let i = (local[2] * self.size[1] + local[1]) * self.size[0] + local[0];
        Some(&self.sites[i * SITE_STRIDE..(i + 1) * SITE_STRIDE])
    }

    /// Every site within `radius` of `c`, row by row: for each z the y span
    /// left inside the sphere, for each y the x span.
    fn for_sites_within(&self, c: [f64; 3], radius: f64, mut f: impl FnMut(i64, i64, i64)) {
        let h = self.spacing;
        let r2 = radius * radius;
        let z0 = ((c[2] - radius) / h).ceil() as i64;
        let z1 = ((c[2] + radius) / h).floor() as i64;
        for z in z0..=z1 {
            let dz = c[2] - z as f64 * h;
            let ry2 = r2 - dz * dz;
            if ry2 < 0.0 {
                continue;
            }
            let ry = ry2.sqrt();
            let y0 = ((c[1] - ry) / h).ceil() as i64;
            let y1 = ((c[1] + ry) / h).floor() as i64;
            for y in y0..=y1 {
                let dy = c[1] - y as f64 * h;
                let rx2 = ry2 - dy * dy;
                if rx2 < 0.0 {
                    continue;
                }
                let rx = rx2.sqrt();
                let x0 = ((c[0] - rx) / h).ceil() as i64;
                let x1 = ((c[0] + rx) / h).floor() as i64;
                for x in x0..=x1 {
                    f(x, y, z);
                }
            }
        }
    }

    /// The pseudo-distance and its gradient at q: d = (phi - level) / beta,
    /// phi the unit-variance sum of the tapered kernels in reach, beta their
    /// energy-weighted wavenumber (local) or the base one.
    pub fn eval(&self, q: [f64; 3]) -> (f64, [f64; 3]) {
        if !self.contains(q) {
            return self.outside(q);
        }
        let a2 = self.a * self.a;
        let r2max = self.reach * self.reach;
        let h = self.spacing;
        let mut acc = 0.0;
        let mut g_acc = [0.0; 3];
        let mut w_acc = 0.0;
        let mut bw_acc = 0.0;
        let mut gw = [0.0; 3];
        let mut gbw = [0.0; 3];
        self.for_sites_within(q, self.reach, |x, y, z| {
            let r = [
                q[0] - x as f64 * h,
                q[1] - y as f64 * h,
                q[2] - z as f64 * h,
            ];
            let d2 = r[0] * r[0] + r[1] * r[1] + r[2] * r[2];
            if d2 >= r2max {
                return;
            }
            let Some(s) = self.site(x, y, z) else {
                return;
            };
            let g = (-PI * a2 * d2).exp();
            let taper = 1.0 - d2 / r2max;
            let k = g * taper * taper;
            let dk = g * taper * (-2.0 * PI * a2 * taper - 4.0 / r2max);
            let beta = s[3];
            let (sn, co) = (beta * (r[0] * s[0] + r[1] * s[1] + r[2] * s[2]) + s[4]).sin_cos();
            acc += k * co;
            let w = k * k;
            w_acc += w;
            bw_acc += w * beta;
            for ax in 0..3 {
                g_acc[ax] += dk * r[ax] * co - k * sn * beta * s[ax];
                let dw = 2.0 * k * dk * r[ax];
                gw[ax] += dw;
                gbw[ax] += dw * beta;
            }
        });
        let phi = acc / self.sigma;
        let gphi = [
            g_acc[0] / self.sigma,
            g_acc[1] / self.sigma,
            g_acc[2] / self.sigma,
        ];
        let mut beta = self.base_beta;
        let mut gbeta = [0.0; 3];
        if self.local && w_acc > 0.0 {
            let b = bw_acc / w_acc;
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

    /// Distance from the box [lo, hi] to point c (0 inside it), and to its
    /// farthest corner.
    fn box_distance(lo: [f64; 3], hi: [f64; 3], c: [f64; 3]) -> (f64, f64) {
        let mut near = 0.0;
        let mut far = 0.0;
        for ax in 0..3 {
            let n = (lo[ax] - c[ax]).max(c[ax] - hi[ax]).max(0.0);
            let f = (c[ax] - lo[ax]).abs().max((hi[ax] - c[ax]).abs());
            near += n * n;
            far += f * f;
        }
        (near.sqrt(), far.sqrt())
    }

    /// A bound on |phi| over the box [lo, hi], the wavenumber range the local
    /// normalization can take there, and a Lipschitz constant of d over it.
    /// The field is smooth everywhere (the taper takes each kernel to zero
    /// with zero slope at its reach), so the Lipschitz bound holds for any box.
    fn bound(&self, lo: [f64; 3], hi: [f64; 3]) -> (f64, f64, f64, f64) {
        let a2 = self.a * self.a;
        let r2max = self.reach * self.reach;
        let h = self.spacing;
        let c = [
            (lo[0] + hi[0]) * 0.5,
            (lo[1] + hi[1]) * 0.5,
            (lo[2] + hi[2]) * 0.5,
        ];
        let half_diag = 0.5
            * ((hi[0] - lo[0]).powi(2) + (hi[1] - lo[1]).powi(2) + (hi[2] - lo[2]).powi(2)).sqrt();
        // Past a table's worth of sites the walk only repeats the same sites.
        let span = 2.0 * (self.reach + half_diag) / h;
        if span > TABLE_SITES as f64 {
            return (
                f64::INFINITY,
                self.beta_min.min(self.base_beta),
                self.beta_max.max(self.base_beta),
                f64::INFINITY,
            );
        }
        let mut env_sum = 0.0;
        let mut grad_sum = 0.0;
        let mut r_far_max: f64 = 0.0;
        let mut b_lo = f64::INFINITY;
        let mut b_hi = f64::NEG_INFINITY;
        self.for_sites_within(c, self.reach + half_diag, |x, y, z| {
            let centre = [x as f64 * h, y as f64 * h, z as f64 * h];
            let (near, far) = Self::box_distance(lo, hi, centre);
            if near >= self.reach {
                return;
            }
            let Some(s) = self.site(x, y, z) else {
                return;
            };
            let g = (-PI * a2 * near * near).exp();
            let taper = 1.0 - near * near / r2max;
            env_sum += g * taper * taper;
            // |grad(k cos)| <= |dk/dr| + k beta; each factor of |dk/dr| is
            // largest at the near or the far end, so their product bounds it.
            let far_r = far.min(self.reach);
            grad_sum += g * taper * far_r * (2.0 * PI * a2 * taper + 4.0 / r2max)
                + g * taper * taper * s[3];
            r_far_max = r_far_max.max(far_r);
            b_lo = b_lo.min(s[3]);
            b_hi = b_hi.max(s[3]);
        });
        let phi_max = env_sum / self.sigma;
        if !self.local {
            b_lo = self.base_beta;
            b_hi = self.base_beta;
        } else {
            b_lo = b_lo.min(self.base_beta);
            b_hi = b_hi.max(self.base_beta);
        }
        let g_phi = grad_sum / self.sigma;
        // |grad beta_local| <= max |grad w| / w * (beta_max - beta_min), and
        // |grad w| / w = 2 |dk/dr| / k <= 2 r (2 pi a^2 + 4 / (R^2 taper)); the
        // taper term is unbounded near the reach, so fall back to infinity when
        // the box's kernels come close to it.
        let g_beta = if self.local && b_hi > b_lo {
            let taper_min = 1.0 - r_far_max * r_far_max / r2max;
            if taper_min > 1e-3 {
                2.0 * r_far_max * (2.0 * PI * a2 + 4.0 / (r2max * taper_min)) * (b_hi - b_lo)
            } else {
                f64::INFINITY
            }
        } else {
            0.0
        };
        let lip = g_phi / b_lo + (phi_max + self.level.abs()) * g_beta / (b_lo * b_lo);
        (phi_max, b_lo, b_hi, lip)
    }
}

/// Interval of (phi - level) / beta for |phi| <= phi_max and beta in [b_lo, b_hi].
fn distance_envelope(phi_max: f64, level: f64, b_lo: f64, b_hi: f64) -> (f64, f64) {
    let n_lo = -phi_max - level;
    let n_hi = phi_max - level;
    let lo = if n_lo >= 0.0 {
        n_lo / b_hi
    } else {
        n_lo / b_lo
    };
    let hi = if n_hi >= 0.0 {
        n_hi / b_lo
    } else {
        n_hi / b_hi
    };
    (lo, hi)
}

// One-entry memo: the four leaves of a slot ask for the same evaluation at the
// same point. Pure, so reuse for bit-identical arguments is always safe.
// Keyed by the table's monotonic `id` rather than its heap address so re-baking
// into a recycled allocation cannot hit a stale memoized sample.
thread_local! {
    static MEMO: RefCell<Option<([u32; 3], u64, (f64, [f64; 3]))>> = const { RefCell::new(None) };
}

fn eval_memo(table: &Arc<Table>, x: f32, y: f32, z: f32) -> (f64, [f64; 3]) {
    let tag = table.id;
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

struct SpinodoidOracle {
    table: Arc<Table>,
    comp: u32,
}

impl MeshOracleTrait for SpinodoidOracle {
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
            if !self.table.contains(lo) || !self.table.contains(hi) {
                // The finite domain has an outside distance extension. A box
                // crossing its boundary cannot use an interior Lipschitz bound.
                return Interval::new(f32::NEG_INFINITY, f32::INFINITY);
            }
        }
        let (phi_max, b_lo, b_hi, lip) = self.table.bound(lo, hi);
        if self.comp != component::DISTANCE {
            // The gradient is only asked for by a .dp chain; bound it by the
            // Lipschitz constant of d, which bounds |grad d|.
            return if lip.is_finite() {
                Interval::new(-(lip as f32), lip as f32)
            } else {
                Interval::new(f32::NEG_INFINITY, f32::INFINITY)
            };
        }
        let (mut out_lo, mut out_hi) = distance_envelope(phi_max, self.table.level, b_lo, b_hi);
        if lip.is_finite() {
            let c = [
                (lo[0] + hi[0]) * 0.5,
                (lo[1] + hi[1]) * 0.5,
                (lo[2] + hi[2]) * 0.5,
            ];
            let (dc, _) = self.table.eval(c);
            let half_diag = 0.5
                * ((hi[0] - lo[0]).powi(2) + (hi[1] - lo[1]).powi(2) + (hi[2] - lo[2]).powi(2))
                    .sqrt();
            out_lo = out_lo.max(dc - lip * half_diag);
            out_hi = out_hi.min(dc + lip * half_diag);
        }
        Interval::new(next_down(out_lo as f32), next_up(out_hi as f32))
    }

    fn grad(&self, x: f32, y: f32, z: f32) -> Grad {
        let (d, g) = eval_memo(&self.table, x, y, z);
        if self.comp == component::DISTANCE {
            return Grad::new(d as f32, g[0] as f32, g[1] as f32, g[2] as f32);
        }
        // The gradient's own derivative (a Hessian row) is not carried.
        Grad::new(
            g[(self.comp - component::GRAD_X) as usize] as f32,
            0.0,
            0.0,
            0.0,
        )
    }
}

fn next_up(v: f32) -> f32 {
    if v.is_nan() || v == f32::INFINITY {
        return v;
    }
    if v == 0.0 {
        return f32::from_bits(1);
    }
    let b = v.to_bits();
    f32::from_bits(if v > 0.0 { b + 1 } else { b - 1 })
}

fn next_down(v: f32) -> f32 {
    -next_up(-v)
}

fn tables() -> &'static Mutex<HashMap<u32, Arc<Table>>> {
    static TABLES: OnceLock<Mutex<HashMap<u32, Arc<Table>>>> = OnceLock::new();
    TABLES.get_or_init(|| Mutex::new(HashMap::new()))
}

/// The oracle leaf id for one component of the table in `slot`.
pub fn oracle_id(slot: u32, comp: u32) -> u32 {
    SPINODOID_ORACLE_ID_BASE + slot * component::COUNT + comp
}

/// The (slot, component) a lattice spinodoid leaf id names, or None if the id
/// is not one of ours — the inverse of [`oracle_id`], for a caller inspecting
/// a tape's oracle leaves.
pub fn leaf_of_id(id: u32) -> Option<(u32, u32)> {
    let off = id.wrapping_sub(SPINODOID_ORACLE_ID_BASE);
    (off < MAX_SLOTS * component::COUNT).then(|| (off / component::COUNT, off % component::COUNT))
}

/// Install a table in `slot` and register its component oracles. Re-registering
/// a slot replaces it; live trees keep the table they were built against.
pub fn register(slot: u32, table: Table) -> Result<(), String> {
    if slot >= MAX_SLOTS {
        return Err(format!("spinodoid slot {} out of range", slot));
    }
    let table = Arc::new(table);
    for comp in 0..component::COUNT {
        mesh_oracle::register(
            oracle_id(slot, comp),
            Arc::new(SpinodoidOracle {
                table: table.clone(),
                comp,
            }),
        );
    }
    tables().lock().unwrap().insert(slot, table);
    Ok(())
}

/// Install a table under the four leaf ids starting at `first_id`, whichever
/// block they are in: a tape reader binds the table it baked to the ids the
/// tape's leaves carry.
pub fn register_leaves(first_id: u32, table: Table) {
    let table = Arc::new(table);
    for comp in 0..component::COUNT {
        mesh_oracle::register(
            first_id + comp,
            Arc::new(SpinodoidOracle {
                table: table.clone(),
                comp,
            }),
        );
    }
}

/// True once `slot` holds a table.
pub fn has(slot: u32) -> bool {
    tables().lock().unwrap().contains_key(&slot)
}
