package fidget

import (
	"testing"

	noise "gradientcontrol.com/omega/core/program/noise"
)

// The Rust noiseHash3 oracle must reproduce the Go reference (common/noise.Hash3)
// bit-for-bit on the same baked tile — that identity is the whole basis for the
// noise matching across backends.
func TestNoiseHash3MatchesGoReference(t *testing.T) {
	if err := RegisterNoiseTile(noise.Bake()); err != nil {
		t.Fatalf("RegisterNoiseTile: %v", err)
	}

	// Spread of coordinates: origin, interior, negatives, and offsets past the
	// fine period (64) and the folded period (4096) so the supercell fold is
	// exercised on both sides.
	cells := [][3]int{
		{0, 0, 0}, {1, 2, 3}, {63, 63, 63}, {64, 0, 0}, {0, 64, 0}, {0, 0, 64},
		{-1, -1, -1}, {-7, 5, -130}, {1000, -1000, 50}, {4096, 0, 0}, {-65, 129, -4097},
	}
	for _, c := range cells {
		want := noise.Hash3(c[0], c[1], c[2])
		got := NoiseHash3(float32(c[0]), float32(c[1]), float32(c[2]))
		for k := 0; k < 3; k++ {
			if got[k] != want[k] {
				t.Errorf("cell %v ch %d: rust=%v go=%v", c, k, got[k], want[k])
			}
		}
	}
}
