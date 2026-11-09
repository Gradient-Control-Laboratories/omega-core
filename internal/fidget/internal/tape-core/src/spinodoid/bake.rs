//! The spinodoid bakes, from a field's arguments: the lattice table (super)
//! and hodge's manufacturing table (super::hodge), each the Go baker's exactly
//! (common/noise/spinodoid: kernels.go Bake and hodge.go HodgeBake, with the
//! hash, pool and layouts they share). This is what lets a tape reader rebuild
//! a spinodoid from the arguments the tape carries, rather than being handed a
//! table.
//!
//! The integer work (the hash, numpy's pool draw) is reproduced bit for bit, so
//! a table's kernels are the Go table's. The scalar constants go through libm
//! (log, exp, cbrt), whose last bit may differ from Go's own math package; the
//! tests pin both against values the Go baker printed.

use std::f64::consts::PI;

use super::numpy::Pcg64;

/// hodge's cutoff: the envelope value at the kernel's reach.
const ENVELOPE_CUTOFF: f64 = 0.01;
const DEFAULT_POOL: usize = 4096;
/// common/noise/spinodoid.MaxTableSites.
const MAX_TABLE_SITES: usize = 1 << 20;

/// Scale driver: the wavelength multiplier (> 0) at every kernel centre.
pub type ScaleDriver<'a> = &'a dyn Fn(&[[f64; 3]]) -> Result<Vec<f64>, String>;
/// Orientation driver: a row-major rotation, applied as R . n, at every kernel
/// centre.
pub type OrientationDriver<'a> = &'a dyn Fn(&[[f64; 3]]) -> Result<Vec<[[f64; 3]; 3]>, String>;

/// A spinodoid's arguments: common/noise/spinodoid.Params.
pub struct Params<'a> {
    /// Cone half-angles, one per pole, in radians.
    pub theta: Vec<f64>,
    /// Cone axes; empty is the Cartesian basis.
    pub poles: Vec<[f64; 3]>,
    pub period: f64,
    pub packet_width: f64,
    pub impulses: usize,
    /// Direction pool size; 0 means hodge's 4096.
    pub pool: usize,
    pub seed: i64,
    pub scale: Option<ScaleDriver<'a>>,
    pub orientation: Option<OrientationDriver<'a>>,
    /// The finite world region a driven field is baked over.
    pub domain: Option<[f64; 6]>,
    /// Keeps the stationary table even with (constant) drivers set.
    pub periodic: bool,
}

const SPLITMIX_GAMMA: u64 = 0x9E3779B97F4A7C15;
const SPLITMIX_M1: u64 = 0xBF58476D1CE4E5B9;
const SPLITMIX_M2: u64 = 0x94D049BB133111EB;

fn splitmix64(x: u64) -> u64 {
    let mut z = x.wrapping_add(SPLITMIX_GAMMA);
    z = (z ^ (z >> 30)).wrapping_mul(SPLITMIX_M1);
    z = (z ^ (z >> 27)).wrapping_mul(SPLITMIX_M2);
    z ^ (z >> 31)
}

/// hodge's _cell_uniform for one cell: a uniform in [0, 1) with 53 random
/// bits. Negative cells and seeds wrap two's-complement, as in Go.
fn cell_uniform(cell: [i64; 3], stream: usize, seed: i64) -> f64 {
    let key = (cell[0] as u64).wrapping_mul(0x9E3779B97F4A7C15)
        ^ (cell[1] as u64).wrapping_mul(0xC2B2AE3D27D4EB4F)
        ^ (cell[2] as u64).wrapping_mul(0x165667B19E3779F9)
        ^ (stream as u64).wrapping_mul(0x27D4EB2F165667C5)
        ^ seed as u64;
    (splitmix64(key) >> 11) as f64 * (1.0 / 9007199254740992.0)
}

/// ugf.normalize's eps: shorter vectors pass through unchanged.
const NORM_EPS: f64 = 1e-12;

fn dot3(a: [f64; 3], b: [f64; 3]) -> f64 {
    a[0] * b[0] + a[1] * b[1] + a[2] * b[2]
}

/// hodge's ugf.normalize for one vector.
fn normalize(v: [f64; 3]) -> [f64; 3] {
    let n = dot3(v, v).sqrt();
    if !(n > NORM_EPS) {
        return v;
    }
    [v[0] / n, v[1] / n, v[2] / n]
}

/// The draw's own v / np.linalg.norm(v): no eps guard.
fn normalize_row(v: [f64; 3]) -> [f64; 3] {
    let n = dot3(v, v).sqrt();
    [v[0] / n, v[1] / n, v[2] / n]
}

/// hodge's spinodoid_waves without the phases: common/noise/spinodoid.Waves.
fn waves(theta: &[f64], n_waves: usize, seed: u64, poles: &[[f64; 3]]) -> Result<Vec<[f64; 3]>, String> {
    let poles: Vec<[f64; 3]> = if poles.is_empty() {
        vec![[1.0, 0.0, 0.0], [0.0, 1.0, 0.0], [0.0, 0.0, 1.0]]
    } else {
        poles.iter().map(|&p| normalize(p)).collect()
    };
    if theta.len() != poles.len() {
        return Err(format!(
            "theta has {} angles but {} poles were given",
            theta.len(),
            poles.len()
        ));
    }
    if theta.iter().any(|&t| t < 0.0 || t > 0.5 * PI + 1e-12) {
        return Err("theta must lie in [0, pi/2] radians".into());
    }
    if !theta.iter().any(|&t| t > 0.0) {
        return Err("at least one cone angle must be > 0".into());
    }
    if n_waves < 1 {
        return Err(format!("n_waves must be >= 1, got {}", n_waves));
    }
    let cos_theta: Vec<f64> = theta.iter().map(|t| t.cos()).collect();

    let mut g = Pcg64::new(seed);
    let mut kept: Vec<[f64; 3]> = Vec::with_capacity(n_waves);
    let (mut n_have, mut n_drawn) = (0usize, 0usize);
    let mut batch = (4 * n_waves).max(1024);
    while n_have < n_waves {
        let mut accepted = 0;
        for _ in 0..batch {
            let mut v = [0.0; 3];
            for c in v.iter_mut() {
                *c = g.standard_normal();
            }
            let n = normalize_row(v);
            for (k, &p) in poles.iter().enumerate() {
                if dot3(n, p).abs() > cos_theta[k] {
                    kept.push(n);
                    accepted += 1;
                    break;
                }
            }
        }
        n_drawn += batch;
        if accepted > 0 {
            n_have += accepted;
        } else if n_drawn > 1_000_000 {
            return Err(format!("cone ODF {:?} rad is too narrow to sample; widen theta", theta));
        }
        let rate = (n_have as f64 / n_drawn as f64).max(1e-6);
        let next = (n_waves as f64 - n_have as f64) / rate * 1.3 + 64.0;
        batch = next.max(1024.0).min((1 << 22) as f64) as usize;
    }
    kept.truncate(n_waves);
    Ok(kept)
}

/// The checks and pool both bakes share, and the clamped theta.
fn pool_directions(p: &Params) -> Result<Vec<[f64; 3]>, String> {
    if !(p.period > 0.0) {
        return Err(format!("period must be > 0, got {}", p.period));
    }
    if !(p.packet_width > 0.0) {
        return Err(format!("packet_width must be > 0, got {}", p.packet_width));
    }
    if p.impulses < 1 {
        return Err(format!("impulses must be >= 1, got {}", p.impulses));
    }
    if p.seed < 0 {
        return Err(format!("seed must be >= 0, got {}", p.seed));
    }
    let pool = if p.pool == 0 { DEFAULT_POOL } else { p.pool };
    // The pool is hodge's draw, in radians. Allow pi/2 to arrive a float32
    // rounding high, and clamp it back.
    let mut theta = Vec::with_capacity(p.theta.len());
    for &t in &p.theta {
        if !(t >= 0.0 && t <= PI / 2.0 + 1e-6) {
            return Err(format!("theta must lie in [0, pi/2] radians, got {}", t));
        }
        theta.push(t.min(PI / 2.0));
    }
    waves(&theta, pool, p.seed as u64, &p.poles)
}

/// Which cells or sites a table holds: common/noise/spinodoid.Layout.
struct Layout {
    origin: [i64; 3],
    size: [usize; 3],
    bounds: [f64; 6],
    periodic: bool,
}

impl Layout {
    fn count(&self) -> usize {
        self.size[0] * self.size[1] * self.size[2]
    }

    fn periodic(edge: i64) -> Layout {
        Layout {
            origin: [-edge / 2; 3],
            size: [edge as usize; 3],
            bounds: [0.0; 6],
            periodic: true,
        }
    }
}

fn check_domain_axis(lo: f64, hi: f64) -> Result<(), String> {
    if !lo.is_finite() || !hi.is_finite() || hi <= lo {
        return Err("spinodoid domain must have finite increasing bounds on every axis".into());
    }
    Ok(())
}

const DOMAIN_TOO_LARGE: &str =
    "graded spinodoid domain is too large; reduce the scene/domain extent or increase the base period";

/// The scalar constants of a lattice table: common/noise/spinodoid.Constants.
struct LatticeConstants {
    reach: f64,
    spacing: f64,
    a: f64,
    sigma: f64,
    base_beta: f64,
}

/// common/noise/spinodoid.TableConstants: Sigma from the kernel density and
/// the tapered kernel's energy, by Simpson's rule.
fn lattice_constants(period: f64, packet_width: f64, impulses: usize) -> LatticeConstants {
    let a = 1.0 / (packet_width * period);
    let reach = (-ENVELOPE_CUTOFF.ln() / PI).sqrt() / a;
    let spacing = reach / (impulses as f64).cbrt();
    const STEPS: usize = 4096;
    let f = |r: f64| {
        let taper = 1.0 - r * r / (reach * reach);
        let k = (-PI * a * a * r * r).exp() * taper * taper;
        k * k * 4.0 * PI * r * r
    };
    let dr = reach / STEPS as f64;
    let mut sum = f(0.0) + f(reach);
    for i in 1..STEPS {
        let w = if i % 2 == 1 { 4.0 } else { 2.0 };
        sum += w * f(i as f64 * dr);
    }
    let energy = sum * dr / 3.0;
    LatticeConstants {
        reach,
        spacing,
        a,
        sigma: (0.5 * energy / (spacing * spacing * spacing)).sqrt(),
        base_beta: 2.0 * PI / period,
    }
}

/// common/noise/spinodoid.NewLayout.
fn lattice_layout(c: &LatticeConstants, domain: Option<[f64; 6]>, periodic: bool) -> Result<Layout, String> {
    let domain = match domain {
        Some(d) if !periodic => d,
        _ => return Ok(Layout::periodic(super::TABLE_SITES)),
    };
    let mut l = Layout {
        origin: [0; 3],
        size: [0; 3],
        bounds: domain,
        periodic: false,
    };
    let mut count = 1usize;
    for a in 0..3 {
        let (lo, hi) = (domain[a], domain[a + 3]);
        check_domain_axis(lo, hi)?;
        let first = ((lo - c.reach) / c.spacing).floor();
        let last = ((hi + c.reach) / c.spacing).ceil();
        let limit = (1i64 << 29) as f64;
        if first.abs() > limit || last.abs() > limit || last - first + 1.0 > MAX_TABLE_SITES as f64 {
            return Err(DOMAIN_TOO_LARGE.into());
        }
        l.origin[a] = first as i64;
        l.size[a] = (last - first) as usize + 1;
        if l.size[a] > MAX_TABLE_SITES / count {
            return Err(format!(
                "graded spinodoid domain exceeds {} kernel sites; reduce the scene/domain extent or increase the base period",
                MAX_TABLE_SITES
            ));
        }
        count *= l.size[a];
    }
    Ok(l)
}

/// Read the drivers once at every kernel centre and finish the kernels'
/// directions and wavenumbers, as both bakes do.
fn apply_drivers(
    p: &Params,
    centers: &[[f64; 3]],
    dirs: &[[f64; 3]],
    reject_infinite_scale: bool,
) -> Result<Vec<([f64; 3], f64)>, String> {
    let n = centers.len();
    let rots = match p.orientation {
        Some(f) => {
            let r = f(centers)?;
            if r.len() != n {
                return Err(format!(
                    "orientation returned {} rotations for {} kernel centres",
                    r.len(),
                    n
                ));
            }
            Some(r)
        }
        None => None,
    };
    let scales = match p.scale {
        Some(f) => {
            let s = f(centers)?;
            if s.len() != n {
                return Err(format!("scale returned {} values for {} kernel centres", s.len(), n));
            }
            Some(s)
        }
        None => None,
    };
    let mut out = Vec::with_capacity(n);
    for i in 0..n {
        let mut dir = dirs[i];
        if let Some(r) = &rots {
            let r = r[i];
            dir = [dot3(r[0], dir), dot3(r[1], dir), dot3(r[2], dir)];
        }
        let mut s = 1.0;
        if let Some(sc) = &scales {
            s = sc[i];
            if !(s > 0.0) || (reject_infinite_scale && s.is_infinite()) {
                return Err(format!("scale must be > 0, got {} at {:?}", s, centers[i]));
            }
        }
        out.push((normalize(dir), 2.0 * PI / (p.period * s)));
    }
    Ok(out)
}

fn wrap_site(c: i64) -> i64 {
    (c + super::TABLE_SITES / 2).rem_euclid(super::TABLE_SITES)
}

/// The lattice table for `p`, leveled at `level`: common/noise/spinodoid.Bake.
pub fn lattice(p: &Params, level: f64, local: bool) -> Result<super::Table, String> {
    let (sites, layout, c) = lattice_sites(p)?;
    super::Table::new(
        sites,
        layout.origin,
        layout.size,
        layout.bounds,
        layout.periodic,
        c.spacing,
        c.reach,
        c.a,
        c.sigma,
        c.base_beta,
        level,
        local,
    )
}

/// The lattice table's sites, in Go's Flat order, and what locates them.
fn lattice_sites(p: &Params) -> Result<(Vec<f64>, Layout, LatticeConstants), String> {
    let base_dirs = pool_directions(p)?;
    let c = lattice_constants(p.period, p.packet_width, p.impulses);
    let layout = lattice_layout(&c, p.domain, p.periodic || (p.scale.is_none() && p.orientation.is_none()))?;
    let n = layout.count();
    let mut centers = vec![[0.0; 3]; n];
    let mut dirs = vec![[0.0; 3]; n];
    let mut gammas = vec![0.0; n];
    let origin = -super::TABLE_SITES / 2;
    for z in 0..layout.size[2] {
        for y in 0..layout.size[1] {
            for x in 0..layout.size[0] {
                let site = [
                    x as i64 + layout.origin[0],
                    y as i64 + layout.origin[1],
                    z as i64 + layout.origin[2],
                ];
                let i = (z * layout.size[1] + y) * layout.size[0] + x;
                let mut random_site = [0i64; 3];
                for a in 0..3 {
                    centers[i][a] = site[a] as f64 * c.spacing;
                    // The world-anchored random realization; only the driver
                    // sampling gains an unwrapped domain.
                    random_site[a] = origin + wrap_site(site[a]);
                }
                let idx = (cell_uniform(random_site, 0, p.seed) * base_dirs.len() as f64) as i64;
                let idx = idx.clamp(0, base_dirs.len() as i64 - 1) as usize;
                dirs[i] = base_dirs[idx];
                gammas[i] = cell_uniform(random_site, 1, p.seed) * 2.0 * PI;
            }
        }
    }
    let finished = apply_drivers(p, &centers, &dirs, true)?;
    let mut sites = Vec::with_capacity(n * super::SITE_STRIDE);
    for (i, (dir, beta)) in finished.into_iter().enumerate() {
        sites.extend_from_slice(&[dir[0], dir[1], dir[2], beta, gammas[i]]);
    }
    Ok((sites, layout, c))
}

/// The scalar constants of hodge's table: common/noise/spinodoid.HodgeConstants.
struct HodgeConstants {
    cell_size: f64,
    a: f64,
    sigma: f64,
    base_beta: f64,
}

fn hodge_constants(period: f64, packet_width: f64, impulses: usize) -> HodgeConstants {
    let a0 = 1.0 / (packet_width * period);
    let cell_size = (-ENVELOPE_CUTOFF.ln() / PI).sqrt() / a0;
    let rho_k = impulses as f64 / (cell_size * cell_size * cell_size);
    HodgeConstants {
        cell_size,
        a: a0,
        sigma: (0.5 * rho_k * (2.0 * a0 * a0).powf(-1.5)).sqrt(),
        base_beta: 2.0 * PI / period,
    }
}

/// common/noise/spinodoid's newHodgeLayout: cells rather than sites, a
/// one-cell pad, and true cell indices (no wrap) when bounded.
fn hodge_layout(c: &HodgeConstants, domain: Option<[f64; 6]>, periodic: bool) -> Result<Layout, String> {
    let domain = match domain {
        Some(d) if !periodic => d,
        _ => return Ok(Layout::periodic(super::hodge::TABLE_CELLS)),
    };
    let mut l = Layout {
        origin: [0; 3],
        size: [0; 3],
        bounds: domain,
        periodic: false,
    };
    for a in 0..3 {
        let (lo, hi) = (domain[a], domain[a + 3]);
        check_domain_axis(lo, hi)?;
        let first = (lo / c.cell_size).floor() - 1.0;
        let last = (hi / c.cell_size).floor() + 1.0;
        let limit = (1i64 << 29) as f64;
        if first.abs() > limit || last.abs() > limit || last - first + 1.0 > MAX_TABLE_SITES as f64 {
            return Err(DOMAIN_TOO_LARGE.into());
        }
        l.origin[a] = first as i64;
        l.size[a] = (last - first) as usize + 1;
    }
    Ok(l)
}

/// hodge's manufacturing table for `p`, leveled at `level`:
/// common/noise/spinodoid.HodgeBake.
pub fn manufacturing(p: &Params, level: f64, local: bool) -> Result<super::hodge::Table, String> {
    let (kernels, layout, c) = manufacturing_kernels(p)?;
    super::hodge::Table::new(
        kernels,
        layout.origin,
        layout.size,
        layout.bounds,
        layout.periodic,
        p.impulses,
        c.cell_size,
        c.a,
        c.sigma,
        c.base_beta,
        level,
        local,
    )
}

/// hodge's table's kernels, in Go's Flat order, and what locates them.
fn manufacturing_kernels(p: &Params) -> Result<(Vec<f64>, Layout, HodgeConstants), String> {
    let base_dirs = pool_directions(p)?;
    let c = hodge_constants(p.period, p.packet_width, p.impulses);
    let layout = hodge_layout(&c, p.domain, p.periodic || (p.scale.is_none() && p.orientation.is_none()))?;
    let cells = layout.count();
    if cells > MAX_TABLE_SITES / p.impulses {
        return Err(format!(
            "a graded hodge spinodoid over this domain needs {} kernels, past the {} limit; \
             reduce the domain extent, raise the base period, or lower impulses",
            cells * p.impulses,
            MAX_TABLE_SITES
        ));
    }
    let n = cells * p.impulses;
    let mut jitters = vec![[0.0; 3]; n];
    let mut centers = vec![[0.0; 3]; n];
    let mut dirs = vec![[0.0; 3]; n];
    let mut gammas = vec![0.0; n];
    // First the hashed part of every kernel, in table order: jitter, pool
    // direction and phase.
    for z in 0..layout.size[2] {
        for y in 0..layout.size[1] {
            for x in 0..layout.size[0] {
                let cell = [
                    x as i64 + layout.origin[0],
                    y as i64 + layout.origin[1],
                    z as i64 + layout.origin[2],
                ];
                let base = ((z * layout.size[1] + y) * layout.size[0] + x) * p.impulses;
                for k in 0..p.impulses {
                    let stream = 8 * k;
                    let i = base + k;
                    for a in 0..3 {
                        jitters[i][a] = cell_uniform(cell, stream + a, p.seed);
                        centers[i][a] = (cell[a] as f64 + jitters[i][a]) * c.cell_size;
                    }
                    let idx = (cell_uniform(cell, stream + 3, p.seed) * base_dirs.len() as f64) as i64;
                    let idx = idx.clamp(0, base_dirs.len() as i64 - 1) as usize;
                    gammas[i] = cell_uniform(cell, stream + 4, p.seed) * (2.0 * PI);
                    dirs[i] = base_dirs[idx];
                }
            }
        }
    }
    // Then the drivers, once per kernel centre, as hodge calls them.
    let finished = apply_drivers(p, &centers, &dirs, false)?;
    let mut kernels = Vec::with_capacity(n * super::hodge::KERNEL_STRIDE);
    for (i, (dir, beta)) in finished.into_iter().enumerate() {
        let j = jitters[i];
        kernels.extend_from_slice(&[j[0], j[1], j[2], dir[0], dir[1], dir[2], beta, gammas[i]]);
    }
    Ok((kernels, layout, c))
}

#[cfg(test)]
mod tests {
    //! Pinned to the Go baker (common/noise/spinodoid Bake and HodgeBake):
    //! the reference values below are what it printed for the same arguments.
    //! The kernels must match bit for bit, through an FNV-1a fingerprint of
    //! their f64 bits in Flat order; the constants and distances must match to
    //! libm rounding.
    use super::*;

    fn fnv(xs: &[f64]) -> u64 {
        let mut h: u64 = 14695981039346656037;
        for x in xs {
            for b in x.to_bits().to_le_bytes() {
                h ^= b as u64;
                h = h.wrapping_mul(1099511628211);
            }
        }
        h
    }

    const CARTESIAN: [[f64; 3]; 3] = [[1.0, 0.0, 0.0], [0.0, 1.0, 0.0], [0.0, 0.0, 1.0]];
    const POINTS: [[f64; 3]; 5] = [
        [0.0, 0.0, 0.0],
        [2.5, -3.75, 6.125],
        [-18.5, 11.25, 4.0],
        [13.0, 7.5, -19.25],
        [30.0, 0.0, 0.0],
    ];

    fn scale(c: &[[f64; 3]]) -> Result<Vec<f64>, String> {
        Ok(c.iter().map(|p| 1.0 + 0.004 * p[0]).collect())
    }

    fn rotation(c: &[[f64; 3]]) -> Result<Vec<[[f64; 3]; 3]>, String> {
        Ok(c.iter()
            .map(|p| {
                let a = 0.05 * p[1];
                [[a.cos(), -a.sin(), 0.0], [a.sin(), a.cos(), 0.0], [0.0, 0.0, 1.0]]
            })
            .collect())
    }

    struct Reference {
        lattice_fnv: u64,
        lattice_size: [usize; 3],
        lattice_origin: [i64; 3],
        /// reach, spacing, a, sigma, base_beta
        lattice_consts: [u64; 5],
        hodge_fnv: u64,
        hodge_size: [usize; 3],
        hodge_origin: [i64; 3],
        /// cell_size, a, sigma, base_beta
        hodge_consts: [u64; 4],
        level: f64,
    }

    fn close_bits(got: f64, want_bits: u64, what: &str) {
        let want = f64::from_bits(want_bits);
        let ulps = (got.to_bits() as i64 - want_bits as i64).abs();
        assert!(ulps <= 4, "{what}: {got:e} vs Go {want:e} ({ulps} ulps)");
    }

    /// `driven` drops the bit-for-bit kernel check: a driver's own arithmetic
    /// is not the bake's (Go fuses `1 + 0.004 x` into an FMA on arm64, and its
    /// cos/sin are its own), so there the distances carry the comparison.
    fn check(p: &Params, r: &Reference, distances: &[(f64, f64)], driven: bool) {
        let (sites, layout, c) = lattice_sites(p).expect("lattice");
        assert_eq!(layout.size, r.lattice_size);
        assert_eq!(layout.origin, r.lattice_origin);
        assert!(driven || fnv(&sites) == r.lattice_fnv, "lattice sites are not Go's bit for bit");
        for (got, (bits, what)) in [c.reach, c.spacing, c.a, c.sigma, c.base_beta]
            .into_iter()
            .zip(r.lattice_consts.iter().zip(["reach", "spacing", "a", "sigma", "base_beta"]))
        {
            close_bits(got, *bits, what);
        }

        let (kernels, layout, c) = manufacturing_kernels(p).expect("manufacturing");
        assert_eq!(layout.size, r.hodge_size);
        assert_eq!(layout.origin, r.hodge_origin);
        assert!(driven || fnv(&kernels) == r.hodge_fnv, "hodge kernels are not Go's bit for bit");
        for (got, (bits, what)) in [c.cell_size, c.a, c.sigma, c.base_beta]
            .into_iter()
            .zip(r.hodge_consts.iter().zip(["cell_size", "a", "sigma", "base_beta"]))
        {
            close_bits(got, *bits, what);
        }

        let lt = lattice(p, r.level, true).expect("lattice table");
        let ht = manufacturing(p, r.level, true).expect("hodge table");
        // Summation order in the evaluators leaves ~5e-15.
        for (q, &(want_l, want_h)) in POINTS.iter().zip(distances) {
            let (got_l, _) = lt.eval(*q);
            let (got_h, _) = ht.eval(*q);
            assert!((got_l - want_l).abs() < 1e-12, "lattice at {q:?}: {got_l} vs Go {want_l}");
            assert!((got_h - want_h).abs() < 1e-12, "hodge at {q:?}: {got_h} vs Go {want_h}");
        }
    }

    #[test]
    fn stationary_matches_go() {
        let p = Params {
            theta: vec![PI / 2.0; 3],
            poles: CARTESIAN.to_vec(),
            period: 6.0,
            packet_width: 3.0,
            impulses: 8,
            pool: 4096,
            seed: 0,
            scale: None,
            orientation: None,
            domain: None,
            periodic: true,
        };
        let r = Reference {
            lattice_fnv: 0x78b8d0d05b216866,
            lattice_size: [64; 3],
            lattice_origin: [-32; 3],
            lattice_consts: [
                0x4035cb0d340ae5ee,
                0x4025cb0d340ae5ee,
                0x3fac71c71c71c71c,
                0x3fe546db0c54c63b,
                0x3ff0c152382d7365,
            ],
            hodge_fnv: 0xba3b91b09bfffde6,
            hodge_size: [32; 3],
            hodge_origin: [-16; 3],
            hodge_consts: [0x4035cb0d340ae5ee, 0x3fac71c71c71c71c, 0x3fec90a9d4feb3b6, 0x3ff0c152382d7365],
            level: 0.3,
        };
        check(
            &p,
            &r,
            &[
                (0.52092486296795693, -0.41810563283014013),
                (1.0646893390705905, -1.0595184216823339),
                (-0.02383445563949289, 0.18173290695875885),
                (-2.2602626099138581, -0.7298651731876189),
                (-0.95940502258342408, -0.69037351183471163),
            ],
            false,
        );
    }

    // A seed past 32 bits (two words of SeedSequence entropy), unnormalized
    // poles, narrow cones, and a pool that is not hodge's default.
    #[test]
    fn wide_seed_and_poles_match_go() {
        let p = Params {
            theta: vec![PI / 2.0, PI / 4.0, PI / 3.0],
            poles: vec![[2.0, 0.0, 0.0], [0.0, 1.0, 1.0], [1.0, -1.0, 3.0]],
            period: 4.0,
            packet_width: 2.5,
            impulses: 6,
            pool: 1000,
            seed: 5_000_000_000,
            scale: None,
            orientation: None,
            domain: None,
            periodic: true,
        };
        let r = Reference {
            lattice_fnv: 0xedd6f1cb72b6cc18,
            lattice_size: [64; 3],
            lattice_origin: [-32; 3],
            lattice_consts: [
                0x402836f239d3385d,
                0x401aa6d2b7e30e85,
                0x3fb999999999999a,
                0x3fe26d1d88dd3244,
                0x3ff921fb54442d18,
            ],
            hodge_fnv: 0xc73f979d3ef6cd1a,
            hodge_size: [32; 3],
            hodge_origin: [-16; 3],
            hodge_consts: [0x402836f239d3385d, 0x3fb999999999999a, 0x3fe8bcf3c8916dd5, 0x3ff921fb54442d18],
            level: -0.4,
        };
        check(
            &p,
            &r,
            &[
                (-0.048578372843710901, 0.76385597832884289),
                (-0.45043473012482022, 0.97663283826805702),
                (-0.27644150562501746, 0.94208121025429192),
                (0.4788681539837249, 0.86130789315520395),
                (1.4222547055223609, -0.22384152597932225),
            ],
            false,
        );
    }

    // Field drivers over a bounded domain: both tables hold true cells, and
    // the point at x = 30 is outside, where both read the distance to it.
    #[test]
    fn driven_matches_go() {
        let p = Params {
            theta: vec![PI / 2.0, PI / 2.0, PI / 6.0],
            poles: CARTESIAN.to_vec(),
            period: 6.0,
            packet_width: 3.0,
            impulses: 8,
            pool: 4096,
            seed: 3,
            scale: Some(&scale),
            orientation: Some(&rotation),
            domain: Some([-20.0, -20.0, -20.0, 20.0, 20.0, 20.0]),
            periodic: false,
        };
        let r = Reference {
            lattice_fnv: 0x9a97bd75f299641a,
            lattice_size: [9; 3],
            lattice_origin: [-4; 3],
            lattice_consts: [
                0x4035cb0d340ae5ee,
                0x4025cb0d340ae5ee,
                0x3fac71c71c71c71c,
                0x3fe546db0c54c63b,
                0x3ff0c152382d7365,
            ],
            hodge_fnv: 0xa865fe642fe4b68b,
            hodge_size: [4; 3],
            hodge_origin: [-2; 3],
            hodge_consts: [0x4035cb0d340ae5ee, 0x3fac71c71c71c71c, 0x3fec90a9d4feb3b6, 0x3ff0c152382d7365],
            level: 0.25,
        };
        check(
            &p,
            &r,
            &[
                (-1.7870893793673694, -1.9002615102383427),
                (0.86395264880181899, 0.44457195544755918),
                (-0.22853634702410267, -0.28181459383269619),
                (0.22820537611315125, 0.065064948031209421),
                (10.0, 10.0),
            ],
            true,
        );
    }
}
