package fidget

import (
	"testing"

	noise "gradientcontrol.com/omega/core/internal/noise"
)

// The one-octave hash — what noiseHash3(cell, 1) lowers to, and what the
// one-octave oracle evaluates at a point — must reproduce the Go reference
// (common/noise.Hash3Octaves) bit-for-bit on the same baked tile, exactly as the
// two-octave hash does. Same basis: the backends agree only because every port
// computes the identical float sequence.
func TestNoiseHash3OneOctaveMatchesGoReference(t *testing.T) {
	if err := RegisterNoiseTile(noise.Bake()); err != nil {
		t.Fatalf("RegisterNoiseTile: %v", err)
	}

	// The same spread the two-octave parity test uses: origin, interior,
	// negatives, and offsets past the fine period (64) and the folded period
	// (4096), so the one-octave path is checked where the fold would have
	// changed the answer.
	cells := [][3]int{
		{0, 0, 0}, {1, 2, 3}, {63, 63, 63}, {64, 0, 0}, {0, 64, 0}, {0, 0, 64},
		{-1, -1, -1}, {-7, 5, -130}, {1000, -1000, 50}, {4096, 0, 0}, {-65, 129, -4097},
	}
	for _, c := range cells {
		want := noise.Hash3Octaves(c[0], c[1], c[2], 1)
		got := NoiseHash3Octaves(float32(c[0]), float32(c[1]), float32(c[2]), 1)
		for k := 0; k < 3; k++ {
			if got[k] != want[k] {
				t.Errorf("cell %v ch %d: rust=%v go=%v", c, k, got[k], want[k])
			}
		}
	}
}

// Asking for two octaves through the octaves entry point must be the same hash
// the original no-octaves call computes — otherwise the default form and the
// explicit form would quietly disagree.
func TestNoiseHash3TwoOctavesMatchesDefault(t *testing.T) {
	if err := RegisterNoiseTile(noise.Bake()); err != nil {
		t.Fatalf("RegisterNoiseTile: %v", err)
	}
	for _, c := range [][3]int{{0, 0, 0}, {1, 2, 3}, {-7, 5, -130}, {4096, 0, 0}} {
		want := NoiseHash3(float32(c[0]), float32(c[1]), float32(c[2]))
		got := NoiseHash3Octaves(float32(c[0]), float32(c[1]), float32(c[2]), 2)
		if got != want {
			t.Errorf("cell %v: octaves=2 %v, default %v", c, got, want)
		}
	}
}

// The octave count is a period choice, and that is the whole reason to offer it:
// one octave repeats every TileSize cells, two every TileSize*TileSize. Assert
// both halves, so a change that quietly folded (or stopped folding) is caught.
func TestOneOctavePeriodIsTileSize(t *testing.T) {
	if err := RegisterNoiseTile(noise.Bake()); err != nil {
		t.Fatalf("RegisterNoiseTile: %v", err)
	}
	const tile = noise.TileSize
	base := NoiseHash3Octaves(1, 2, 3, 1)

	// One octave repeats at +TileSize...
	if got := NoiseHash3Octaves(1+tile, 2, 3, 1); got != base {
		t.Errorf("one octave should repeat every %d cells: %v vs %v", tile, got, base)
	}
	// ...where two octaves do not, which is what the fold buys.
	two := NoiseHash3Octaves(1, 2, 3, 2)
	if got := NoiseHash3Octaves(1+tile, 2, 3, 2); got == two {
		t.Errorf("two octaves should not repeat at +%d cells: %v", tile, got)
	}
	// Two octaves repeat at +TileSize^2.
	if got := NoiseHash3Octaves(1+tile*tile, 2, 3, 2); got != two {
		t.Errorf("two octaves should repeat every %d cells: %v vs %v", tile*tile, got, two)
	}
}

// Every value the hash returns must land in [0,1), at either octave count. One
// octave is the case that needs saying: a texel byte of 255 samples as exactly
// 1.0, and only the fract in hash_channel_octaves keeps it in range.
func TestHash3OctavesInUnitRange(t *testing.T) {
	if err := RegisterNoiseTile(noise.Bake()); err != nil {
		t.Fatalf("RegisterNoiseTile: %v", err)
	}
	for _, octaves := range []uint32{1, 2} {
		for x := -3; x < 3; x++ {
			for y := 0; y < noise.TileSize; y++ {
				for z := 0; z < noise.TileSize; z++ {
					h := NoiseHash3Octaves(float32(x), float32(y), float32(z), octaves)
					for k := 0; k < 3; k++ {
						if h[k] < 0 || h[k] >= 1 {
							t.Fatalf("octaves=%d cell %d,%d,%d ch %d: %v outside [0,1)",
								octaves, x, y, z, k, h[k])
						}
					}
				}
			}
		}
	}
}
