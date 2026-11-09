package spinodoid

import (
	"math"
	"math/bits"
)

// A port of the slice of numpy (v2.2.6) that hodge's spinodoid_waves draws
// from: np.random.default_rng(seed) — a SeedSequence feeding PCG64 — and the
// Generator's standard_normal (a 256-layer ziggurat) and uniform. The direction
// pool is drawn by rejection from these normals, so reproducing hodge's pool
// means reproducing numpy's stream bit for bit; nothing here is "a PCG64", it is
// numpy's PCG64, seeded numpy's way.
//
// Products that feed an addition are wrapped in float64(): the Go spec lets the
// compiler fuse x*y + z into one FMA (arm64 does), and a fused result is not
// the value numpy computes.

// SeedSequence constants (numpy/random/bit_generator.pyx).
const (
	ssPoolSize = 4
	ssInitA    = 0x43b0d7e5
	ssMultA    = 0x931e8875
	ssInitB    = 0x8b51f9dd
	ssMultB    = 0x58f38ded
	ssMixMultL = 0xca01f9dd
	ssMixMultR = 0x4973f715
	ssXShift   = 16
)

func ssHashmix(value uint32, hashConst *uint32) uint32 {
	value ^= *hashConst
	*hashConst *= ssMultA
	value *= *hashConst
	value ^= value >> ssXShift
	return value
}

func ssMix(x, y uint32) uint32 {
	result := ssMixMultL*x - ssMixMultR*y
	result ^= result >> ssXShift
	return result
}

// seedSequenceState is SeedSequence(seed).generate_state(4, np.uint64) for a
// non-negative integer seed with no spawn key — the four words PCG64 seeds
// itself from.
func seedSequenceState(seed uint64) [4]uint64 {
	// _int_to_uint32_array: lowest word first, and 0 is one zero word.
	entropy := []uint32{uint32(seed)}
	if seed>>32 != 0 {
		entropy = append(entropy, uint32(seed>>32))
	}

	// mix_entropy.
	var pool [ssPoolSize]uint32
	hashConst := uint32(ssInitA)
	for i := range pool {
		if i < len(entropy) {
			pool[i] = ssHashmix(entropy[i], &hashConst)
		} else {
			pool[i] = ssHashmix(0, &hashConst)
		}
	}
	for src := range pool {
		for dst := range pool {
			if src != dst {
				pool[dst] = ssMix(pool[dst], ssHashmix(pool[src], &hashConst))
			}
		}
	}
	for src := ssPoolSize; src < len(entropy); src++ {
		for dst := range pool {
			pool[dst] = ssMix(pool[dst], ssHashmix(entropy[src], &hashConst))
		}
	}

	// generate_state(4, uint64): eight uint32 words cycled from the pool, then
	// viewed little-endian as four uint64.
	var words [8]uint32
	hashB := uint32(ssInitB)
	for i := range words {
		v := pool[i%ssPoolSize]
		v ^= hashB
		hashB *= ssMultB
		v *= hashB
		v ^= v >> ssXShift
		words[i] = v
	}
	var out [4]uint64
	for i := range out {
		out[i] = uint64(words[2*i]) | uint64(words[2*i+1])<<32
	}
	return out
}

// PCG64's default 128-bit multiplier (numpy/random/src/pcg64/pcg64.h).
const (
	pcgMultHigh = 2549297995355413924
	pcgMultLow  = 4865540595714422341
)

// pcg64 is numpy's PCG64: a 128-bit LCG with the XSL-RR output function.
type pcg64 struct {
	stateHigh, stateLow uint64
	incHigh, incLow     uint64
}

func (g *pcg64) step() {
	hi, lo := bits.Mul64(g.stateLow, pcgMultLow)
	hi += g.stateHigh*pcgMultLow + g.stateLow*pcgMultHigh
	var carry uint64
	g.stateLow, carry = bits.Add64(lo, g.incLow, 0)
	g.stateHigh, _ = bits.Add64(hi, g.incHigh, carry)
}

// newGenerator is np.random.default_rng(seed).
func newGenerator(seed uint64) *pcg64 {
	s := seedSequenceState(seed)
	initHigh, initLow := s[0], s[1]
	seqHigh, seqLow := s[2], s[3]
	g := &pcg64{
		incHigh: seqHigh<<1 | seqLow>>63,
		incLow:  seqLow<<1 | 1,
	}
	g.step()
	var carry uint64
	g.stateLow, carry = bits.Add64(g.stateLow, initLow, 0)
	g.stateHigh, _ = bits.Add64(g.stateHigh, initHigh, carry)
	g.step()
	return g
}

func (g *pcg64) nextUint64() uint64 {
	g.step()
	return bits.RotateLeft64(g.stateHigh^g.stateLow, -int(g.stateHigh>>58))
}

// nextDouble is numpy's uint64_to_double: 53 random bits in [0, 1).
func (g *pcg64) nextDouble() float64 {
	return float64(g.nextUint64()>>11) * (1.0 / 9007199254740992.0)
}

// uniform is Generator.uniform(low, high) for scalar bounds.
func (g *pcg64) uniform(low, high float64) float64 {
	span := high - low
	return low + float64(span*g.nextDouble())
}

// standardNormal is numpy's random_standard_normal
// (numpy/random/src/distributions/distributions.c), the ziggurat that
// Generator.normal() with loc 0 and scale 1 reduces to.
func (g *pcg64) standardNormal() float64 {
	for {
		r := g.nextUint64()
		idx := r & 0xff
		r >>= 8
		sign := r & 0x1
		rabs := (r >> 1) & 0x000fffffffffffff
		x := float64(rabs) * zigWi[idx]
		if sign&0x1 != 0 {
			x = -x
		}
		if rabs < zigKi[idx] {
			return x
		}
		if idx == 0 {
			for {
				xx := float64(-zigNorInvR * math.Log1p(-g.nextDouble()))
				yy := -math.Log1p(-g.nextDouble())
				if yy+yy > float64(xx*xx) {
					if (rabs>>8)&0x1 != 0 {
						return -(zigNorR + xx)
					}
					return zigNorR + xx
				}
			}
		}
		if float64((zigFi[idx-1]-zigFi[idx])*g.nextDouble())+zigFi[idx] < math.Exp(float64(-0.5*x)*x) {
			return x
		}
	}
}
