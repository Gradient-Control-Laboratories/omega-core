//! The two Gabor spinodoids as fidget oracles, each over a table the host
//! bakes, each with its own slots and leaf ids:
//!
//! - [`fast_gabor`]: the fast spinodoid (fastGaborSpinodoid), one tapered
//!   kernel per lattice site.
//! - [`reference_gabor`]: the reference spinodoid (gaborSpinodoid), kernels
//!   jittered inside the cells of its grid, untapered.
//!
//! Each registered table is a slot; a slot has COMPONENTS oracle leaves, the
//! pseudo-distance d and the three components of its analytic gradient, which
//! share one memoized evaluation. [`bake`] builds either table from a field's
//! arguments, for a tape reader that rebuilds a spinodoid from its payload.

pub mod bake;
pub mod fast_gabor;
mod numpy;
pub mod reference_gabor;
mod ziggurat_tables;

/// What a leaf returns: the pseudo-distance, or one gradient component.
pub mod component {
    pub const DISTANCE: u32 = 0;
    pub const GRAD_X: u32 = 1;
    pub const COUNT: u32 = 4;
}

/// How many tables of each kind can be registered at once.
pub const MAX_SLOTS: u32 = 256;

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
