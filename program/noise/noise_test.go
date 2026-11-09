package noise

import "testing"

// Bake is deterministic and correctly sized.
func TestBakeDeterministic(t *testing.T) {
	a, b := Bake(), Bake()
	want := TileSize * TileSize * TileSize * Channels
	if len(a) != want {
		t.Fatalf("tile size = %d, want %d", len(a), want)
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("bake not deterministic at %d: %d vs %d", i, a[i], b[i])
		}
	}
}

// The hash bytes should be well spread, not clustered — a sanity check on the
// finalizer, not a strict randomness test. Mean per channel near 127.5.
func TestBakeDistribution(t *testing.T) {
	b := Bake()
	var sum [Channels]int64
	n := TileSize * TileSize * TileSize
	for i := 0; i < n; i++ {
		for c := 0; c < Channels; c++ {
			sum[c] += int64(b[i*Channels+c])
		}
	}
	for c := 0; c < Channels; c++ {
		mean := float64(sum[c]) / float64(n)
		if mean < 120 || mean > 135 {
			t.Errorf("channel %d mean %.2f far from 127.5", c, mean)
		}
	}
}

// Hash3 stays in [0,1) on every channel, including for negative coordinates.
func TestHash3Range(t *testing.T) {
	for _, c := range [][3]int{{0, 0, 0}, {5, 9, 13}, {-1, -7, -130}, {1000, -1000, 50}} {
		h := Hash3(c[0], c[1], c[2])
		for k := 0; k < 3; k++ {
			if h[k] < 0 || h[k] >= 1 {
				t.Errorf("Hash3%v[%d] = %f out of [0,1)", c, k, h[k])
			}
		}
	}
}

// The supercell fold sets the period to exactly TileSize*TileSize: a +TileSize
// shift changes the value (period > TileSize), and a +TileSize*TileSize shift
// repeats it. Checked along each axis.
func TestHash3FoldPeriod(t *testing.T) {
	base := Hash3(0, 0, 0)
	period := TileSize * TileSize

	type shift struct {
		name    string
		at      [3]int
		repeats bool
	}
	for _, s := range []shift{
		{"+TileSize x", [3]int{TileSize, 0, 0}, false},
		{"+TileSize y", [3]int{0, TileSize, 0}, false},
		{"+TileSize z", [3]int{0, 0, TileSize}, false},
		{"+period x", [3]int{period, 0, 0}, true},
		{"+period y", [3]int{0, period, 0}, true},
		{"+period z", [3]int{0, 0, period}, true},
	} {
		h := Hash3(s.at[0], s.at[1], s.at[2])
		if equal3(h, base) != s.repeats {
			t.Errorf("%s: repeats=%v, want %v (base=%v got=%v)", s.name, equal3(h, base), s.repeats, base, h)
		}
	}
}

func equal3(a, b [3]float32) bool { return a[0] == b[0] && a[1] == b[1] && a[2] == b[2] }
