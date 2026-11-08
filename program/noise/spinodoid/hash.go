package spinodoid

// hodge's stateless kernel hash (spinodoid.py _splitmix64 / _cell_uniform):
// a SplitMix64 finalizer over a key mixed from the integer cell, a stream id
// and the seed. Integer-only, so it lives here rather than in any backend; the
// backends read its output from the baked kernel table.

const (
	splitmixGamma = 0x9E3779B97F4A7C15
	splitmixM1    = 0xBF58476D1CE4E5B9
	splitmixM2    = 0x94D049BB133111EB
)

func splitmix64(x uint64) uint64 {
	z := x + splitmixGamma
	z = (z ^ (z >> 30)) * splitmixM1
	z = (z ^ (z >> 27)) * splitmixM2
	return z ^ (z >> 31)
}

// cellUniform is _cell_uniform(cells, stream, seed) for one cell: a uniform in
// [0, 1) with 53 random bits. Negative cells and seeds wrap two's-complement
// into uint64, as numpy's int64 -> uint64 cast and Python's % 2**64 both do.
func cellUniform(cell [3]int64, stream int, seed int64) float64 {
	key := uint64(cell[0])*0x9E3779B97F4A7C15 ^
		uint64(cell[1])*0xC2B2AE3D27D4EB4F ^
		uint64(cell[2])*0x165667B19E3779F9 ^
		uint64(stream)*0x27D4EB2F165667C5 ^
		uint64(seed)
	return float64(splitmix64(key)>>11) * (1.0 / 9007199254740992.0)
}
