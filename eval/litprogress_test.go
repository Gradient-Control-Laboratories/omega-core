//go:build cgo

package eval

import (
	"math"
	"testing"
)

// litProgressCamera looks at the origin from (6, 6, 6), as the default view
// does, written out for the viewer's march.
func litProgressCamera() *LitCamera {
	a, b := math.Sqrt(0.5), math.Sqrt(1.0/3.0)
	return &LitCamera{
		Pos:    [3]float64{6, 6, 6},
		Right:  [3]float64{-a, a, 0},
		Up:     [3]float64{-a * b, -a * b, 2 * a * b},
		Fwd:    [3]float64{-b, -b, -b},
		TanFov: [2]float64{math.Tan(25*math.Pi/180) * 96.0 / 64.0, math.Tan(25 * math.Pi / 180)},
	}
}

// Progress reports the share of a render's work as it goes: inside [0, 1],
// never back, more than once and not only at the end, and finishing at 1.
func TestLitProgressRisesToOne(t *testing.T) {
	var shares []float64
	_, err := RenderLitLinear(litProgram(t, litBallDoc), LitOptions{
		Width: 96, Height: 64, WorldExtent: 6, Camera: litProgressCamera(), ViewerMarch: true,
		Progress: func(s float64) { shares = append(shares, s) },
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(shares) < 3 {
		t.Fatalf("%d progress reports, want several along the way: %v", len(shares), shares)
	}
	for i, s := range shares {
		if s < 0 || s > 1 {
			t.Errorf("report %d is %g, outside [0, 1]", i, s)
		}
		if i > 0 && s < shares[i-1] {
			t.Errorf("report %d is %g, back from %g", i, s, shares[i-1])
		}
	}
	if last := shares[len(shares)-1]; last != 1 {
		t.Errorf("the last report is %g, want 1", last)
	}
	if shares[0] >= 0.5 {
		t.Errorf("the first report is %g: progress should start low", shares[0])
	}
}

// Asking for progress does not change the image.
func TestLitProgressDoesNotChangeTheImage(t *testing.T) {
	opts := LitOptions{Width: 96, Height: 64, WorldExtent: 6, Camera: litProgressCamera(), ViewerMarch: true}
	prog := litProgram(t, litBallDoc)
	plain, err := RenderLitLinear(prog, opts)
	if err != nil {
		t.Fatal(err)
	}
	opts.Progress = func(float64) {}
	watched, err := RenderLitLinear(prog, opts)
	if err != nil {
		t.Fatal(err)
	}
	if len(plain.RGB) != len(watched.RGB) {
		t.Fatalf("%d and %d values", len(plain.RGB), len(watched.RGB))
	}
	for i := range plain.RGB {
		if plain.RGB[i] != watched.RGB[i] {
			t.Fatalf("value %d is %g with progress and %g without", i, watched.RGB[i], plain.RGB[i])
		}
	}
}

// A render that does not march as the viewer does reports only that it is done.
func TestLitProgressWithoutViewerMarchReportsDone(t *testing.T) {
	var shares []float64
	_, err := RenderLitLinear(litProgram(t, litBallDoc), LitOptions{
		Width: 96, Height: 64, WorldExtent: 6,
		Progress: func(s float64) { shares = append(shares, s) },
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(shares) != 1 || shares[0] != 1 {
		t.Errorf("reports %v, want just [1]", shares)
	}
}
