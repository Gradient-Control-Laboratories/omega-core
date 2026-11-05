//! A port of the slice of numpy (v2.2.6) that hodge's spinodoid_waves draws
//! from: np.random.default_rng(seed) — a SeedSequence feeding PCG64 — and the
//! Generator's standard_normal (a 256-layer ziggurat). The direction pool is
//! drawn by rejection from these normals, so reproducing hodge's pool means
//! reproducing numpy's stream bit for bit. This mirrors the Go port
//! (common/noise/spinodoid/numpy.go) line for line; Rust never fuses x*y + z
//! into an FMA, so the products need no guarding here.

use super::ziggurat_tables::{ZIG_FI, ZIG_KI, ZIG_NOR_INV_R, ZIG_NOR_R, ZIG_WI};

// SeedSequence constants (numpy/random/bit_generator.pyx).
const SS_POOL_SIZE: usize = 4;
const SS_INIT_A: u32 = 0x43b0d7e5;
const SS_MULT_A: u32 = 0x931e8875;
const SS_INIT_B: u32 = 0x8b51f9dd;
const SS_MULT_B: u32 = 0x58f38ded;
const SS_MIX_MULT_L: u32 = 0xca01f9dd;
const SS_MIX_MULT_R: u32 = 0x4973f715;
const SS_XSHIFT: u32 = 16;

fn ss_hashmix(mut value: u32, hash_const: &mut u32) -> u32 {
    value ^= *hash_const;
    *hash_const = hash_const.wrapping_mul(SS_MULT_A);
    value = value.wrapping_mul(*hash_const);
    value ^= value >> SS_XSHIFT;
    value
}

fn ss_mix(x: u32, y: u32) -> u32 {
    let mut result = SS_MIX_MULT_L.wrapping_mul(x).wrapping_sub(SS_MIX_MULT_R.wrapping_mul(y));
    result ^= result >> SS_XSHIFT;
    result
}

/// SeedSequence(seed).generate_state(4, np.uint64) for a non-negative integer
/// seed with no spawn key — the four words PCG64 seeds itself from.
fn seed_sequence_state(seed: u64) -> [u64; 4] {
    // _int_to_uint32_array: lowest word first, and 0 is one zero word.
    let mut entropy = vec![seed as u32];
    if seed >> 32 != 0 {
        entropy.push((seed >> 32) as u32);
    }

    // mix_entropy.
    let mut pool = [0u32; SS_POOL_SIZE];
    let mut hash_const = SS_INIT_A;
    for (i, p) in pool.iter_mut().enumerate() {
        *p = ss_hashmix(entropy.get(i).copied().unwrap_or(0), &mut hash_const);
    }
    for src in 0..SS_POOL_SIZE {
        for dst in 0..SS_POOL_SIZE {
            if src != dst {
                pool[dst] = ss_mix(pool[dst], ss_hashmix(pool[src], &mut hash_const));
            }
        }
    }
    for &e in entropy.iter().skip(SS_POOL_SIZE) {
        for p in pool.iter_mut() {
            *p = ss_mix(*p, ss_hashmix(e, &mut hash_const));
        }
    }

    // generate_state(4, uint64): eight uint32 words cycled from the pool, then
    // viewed little-endian as four uint64.
    let mut words = [0u32; 8];
    let mut hash_b = SS_INIT_B;
    for (i, w) in words.iter_mut().enumerate() {
        let mut v = pool[i % SS_POOL_SIZE];
        v ^= hash_b;
        hash_b = hash_b.wrapping_mul(SS_MULT_B);
        v = v.wrapping_mul(hash_b);
        v ^= v >> SS_XSHIFT;
        *w = v;
    }
    let mut out = [0u64; 4];
    for (i, o) in out.iter_mut().enumerate() {
        *o = words[2 * i] as u64 | (words[2 * i + 1] as u64) << 32;
    }
    out
}

// PCG64's default 128-bit multiplier (numpy/random/src/pcg64/pcg64.h).
const PCG_MULT: u128 = (2549297995355413924u128 << 64) | 4865540595714422341u128;

/// numpy's PCG64: a 128-bit LCG with the XSL-RR output function.
pub(super) struct Pcg64 {
    state: u128,
    inc: u128,
}

impl Pcg64 {
    fn step(&mut self) {
        self.state = self.state.wrapping_mul(PCG_MULT).wrapping_add(self.inc);
    }

    /// np.random.default_rng(seed).
    pub(super) fn new(seed: u64) -> Pcg64 {
        let s = seed_sequence_state(seed);
        let init = (s[0] as u128) << 64 | s[1] as u128;
        let seq = (s[2] as u128) << 64 | s[3] as u128;
        let mut g = Pcg64 { state: 0, inc: (seq << 1) | 1 };
        g.step();
        g.state = g.state.wrapping_add(init);
        g.step();
        g
    }

    fn next_u64(&mut self) -> u64 {
        self.step();
        let hi = (self.state >> 64) as u64;
        let lo = self.state as u64;
        (hi ^ lo).rotate_right((hi >> 58) as u32)
    }

    /// numpy's uint64_to_double: 53 random bits in [0, 1).
    fn next_double(&mut self) -> f64 {
        (self.next_u64() >> 11) as f64 * (1.0 / 9007199254740992.0)
    }

    /// numpy's random_standard_normal
    /// (numpy/random/src/distributions/distributions.c), the ziggurat that
    /// Generator.normal() with loc 0 and scale 1 reduces to.
    pub(super) fn standard_normal(&mut self) -> f64 {
        loop {
            let mut r = self.next_u64();
            let idx = (r & 0xff) as usize;
            r >>= 8;
            let sign = r & 0x1;
            let rabs = (r >> 1) & 0x000f_ffff_ffff_ffff;
            let mut x = rabs as f64 * ZIG_WI[idx];
            if sign & 0x1 != 0 {
                x = -x;
            }
            if rabs < ZIG_KI[idx] {
                return x;
            }
            if idx == 0 {
                loop {
                    let xx = -ZIG_NOR_INV_R * (-self.next_double()).ln_1p();
                    let yy = -(-self.next_double()).ln_1p();
                    if yy + yy > xx * xx {
                        if (rabs >> 8) & 0x1 != 0 {
                            return -(ZIG_NOR_R + xx);
                        }
                        return ZIG_NOR_R + xx;
                    }
                }
            }
            if (ZIG_FI[idx - 1] - ZIG_FI[idx]) * self.next_double() + ZIG_FI[idx] < (-0.5 * x * x).exp() {
                return x;
            }
        }
    }
}
