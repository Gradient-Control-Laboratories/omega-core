package gpuparity

import (
	"math"
	"testing"
)

// remap(field, coords) evaluates field with p := coords. With coords a
// constant point the field is read at that one point, wherever it is asked
// about, as the CPU reads it; the WGSL once wrote it as a translation, p - c,
// so the GPU drew the shape moved where the CPU and every export read a
// constant. Found by drill.hole's size probe (Phase 19). The probe, 4 at
// (0, 0, 5), is added to a ball less 4, so the field is the ball's, |p| - 2,
// only when it is read as the constant it is.
func TestGpuRemapToAConstantPoint(t *testing.T) {
	doc := "probe = remap(length(p) - 1.0, [0.0, 0.0, 5.0]);\nball = length(p) - 2.0 + probe - 4.0;\nreturn ball;\n"
	pts := [][3]float32{{0, 0, 1.5}, {1, 0, 0}, {0, -3, 1}}
	got := pruneGPU(t, doc, [3]float64{0, 0, 50}, 0, pts)
	for i, d := range got {
		p := pts[i]
		want := math.Sqrt(float64(p[0]*p[0]+p[1]*p[1]+p[2]*p[2])) - 2
		if math.Abs(float64(d)-want) > 1e-4 {
			t.Errorf("at %v: %g, want %g, the ball's", p, d, want)
		}
	}
}
