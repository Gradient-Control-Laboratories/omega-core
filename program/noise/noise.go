// Package noise is the canonical source of truth for the procedural-noise
// hash used by the GSL noise library (common/source/package/noise).
//
// No rendering backend (WGSL, Fidget, const-fold) has integer or bitwise math,
// so a per-cell pseudo-random hash cannot be computed at evaluation time. The
// resolution: bake a fixed TileSize^3 RGBA8 table HERE (in Go, where real
// integer hashing is available), ship the bytes to every backend, and at
// evaluation time do only float work — a nearest (integer-indexed) lookup plus
// one "supercell" fold. That lookup+fold is float-only and uses no
// interpolation, so it is bit-identical across backends and the resulting
// Worley/gradient fields stay Lipschitz (meshable in Fidget).
//
// Bake() defines the bytes; Hash3() defines the lookup+fold the WGSL shader and
// the Rust Fidget oracle must replicate exactly. Both are exercised by tests so
// the three ports cannot silently drift.
package noise

import "math"

const (
	// TileSize is the cells-per-axis of the baked tile. A single nearest
	// lookup therefore repeats every TileSize cells; the supercell fold in
	// Hash3 raises the period to TileSize*TileSize.
	TileSize = 64
	// Channels is the per-cell byte count (RGBA): RGB carry a feature/gradient
	// vector, A is a spare scalar (cell id / weight / 4th axis).
	Channels = 4
)

// cellHash returns Channels well-mixed pseudo-random bytes for an integer cell,
// via a splitmix64-style finalizer so all bits (hence all 4 bytes) avalanche
// independently. This is the offline integer hash the backends cannot run.
func cellHash(x, y, z int) (byte, byte, byte, byte) {
	h := uint64(uint32(x))*0x9E3779B97F4A7C15 +
		uint64(uint32(y))*0xC2B2AE3D27D4EB4F +
		uint64(uint32(z))*0x165667B19E3779F9
	h ^= h >> 30
	h *= 0xBF58476D1CE4E5B9
	h ^= h >> 27
	h *= 0x94D049BB133111EB
	h ^= h >> 31
	return byte(h), byte(h >> 8), byte(h >> 16), byte(h >> 24)
}

// Bake produces the TileSize^3 RGBA8 tile, x fastest then y then z (the layout
// a Data3DTexture / row-major 3D buffer expects). len == TileSize^3 * Channels.
func Bake() []byte {
	out := make([]byte, TileSize*TileSize*TileSize*Channels)
	i := 0
	for z := 0; z < TileSize; z++ {
		for y := 0; y < TileSize; y++ {
			for x := 0; x < TileSize; x++ {
				out[i], out[i+1], out[i+2], out[i+3] = cellHash(x, y, z)
				i += 4
			}
		}
	}
	return out
}

// tile is the baked table, computed once; Hash3 reads it.
var tile = Bake()

// wrap maps an integer coordinate into [0, TileSize) with true modulo
// (Go's % keeps the sign of the dividend, which would mis-wrap negatives).
func wrap(v int) int { return ((v % TileSize) + TileSize) % TileSize }

// floorDiv is integer division rounding toward -inf (the supercell index).
func floorDiv(a, b int) int {
	q := a / b
	if (a%b != 0) && ((a < 0) != (b < 0)) {
		q--
	}
	return q
}

// sample reads one texel as Channels floats in [0,1), nearest (no interpolation),
// wrapping the index. The backends do exactly this with textureLoad.
func sample(x, y, z int) [Channels]float32 {
	i := ((wrap(z)*TileSize+wrap(y))*TileSize + wrap(x)) * Channels
	return [Channels]float32{
		float32(tile[i]) / 255, float32(tile[i+1]) / 255,
		float32(tile[i+2]) / 255, float32(tile[i+3]) / 255,
	}
}

// fract returns x - floor(x), in float32 to match the backends' f32 math.
func fract(x float32) float32 { return x - float32(math.Floor(float64(x))) }

// Hash3 is the canonical per-cell hash: a fine nearest lookup combined with a
// coarse "supercell" lookup via fract(fine+coarse). Returns RGB in [0,1)^3 (A is
// spare). The fold raises the repeat period from TileSize to TileSize*TileSize
// while staying float-only and interpolation-free, so WGSL and the Fidget Rust
// oracle can reproduce it bit-for-bit. Worley reads this as its feature offset.
func Hash3(x, y, z int) [3]float32 {
	return Hash3Octaves(x, y, z, 2)
}

// Hash3Octaves is Hash3 over a chosen number of tile lookups — the form the
// GSL builtin's optional octaves argument selects. 2 is Hash3 itself: fine plus
// supercell, repeating every TileSize*TileSize cells. 1 is the bare nearest
// lookup, half the texture traffic and repeating every TileSize cells, which a
// 27-neighbour cellular scan pays 27 times over.
//
// One octave still passes through fract. A texel byte of 255 samples as exactly
// 1.0, which is outside the [0,1) the hash promises (and would put a cellular
// feature point on its cell's far face); fract maps that one byte to 0.0 and
// leaves the other 255 alone. Every port must do the same or the backends drift.
func Hash3Octaves(x, y, z, octaves int) [3]float32 {
	fine := sample(x, y, z)
	if octaves <= 1 {
		return [3]float32{fract(fine[0]), fract(fine[1]), fract(fine[2])}
	}
	coarse := sample(floorDiv(x, TileSize), floorDiv(y, TileSize), floorDiv(z, TileSize))
	return [3]float32{
		fract(fine[0] + coarse[0]),
		fract(fine[1] + coarse[1]),
		fract(fine[2] + coarse[2]),
	}
}
