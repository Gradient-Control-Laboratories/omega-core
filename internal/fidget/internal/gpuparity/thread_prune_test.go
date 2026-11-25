package gpuparity

import (
	"math"
	"testing"
)

// Shaft Modeling, Phase 20: a pruned thread, run on the GPU. An M6 x 1's
// crests and roots are half a pitch, 0.5, across; at a pixel scale of 0.001
// they are 0.5 pixels from 1000 away and give way, leaving the cylinder at the
// pitch diameter, 6 - 0.6495, radius 2.67525: the field is r - 2.67525 at any
// angle and height. From 20 away they are 25 pixels and stay.

const threadPitchR = (6.0 - 0.6495) / 2

func threadRing() ([][3]float32, []float64) {
	var pts [][3]float32
	var radii []float64
	for _, r := range []float64{2.45, 2.6, 2.75, 2.9} {
		for j := 0; j < 8; j++ {
			a := 2 * math.Pi * float64(j) / 8
			z := -0.5 + float64(j)/8
			pts = append(pts, [3]float32{float32(r * math.Cos(a)), float32(r * math.Sin(a)), float32(z)})
			radii = append(radii, r)
		}
	}
	return pts, radii
}

func TestGpuPrune_ThreadLeavesThePitchCylinder(t *testing.T) {
	pts, radii := threadRing()
	for _, c := range []struct {
		name string
		doc  string
		sign float64 // +1 for a screw, -1 for a hole: the field outside the pitch cylinder
	}{
		{"male", "s = max(thread.male(6.0, 1.0), abs(p.z) - 10.0);\nreturn s;\n", 1},
		{"maleSin", "s = max(thread.maleSin(6.0, 1.0), abs(p.z) - 10.0);\nreturn s;\n", 1},
		{"female", "s = max(box([-10.0, -10.0, -10.0], [10.0, 10.0, 10.0]), neg(thread.female(6.0, 1.0)));\nreturn s;\n", -1},
	} {
		far := pruneGPU(t, c.doc, [3]float64{0, 0, 1000}, 0.001, pts)
		for i, d := range far {
			if want := c.sign * (radii[i] - threadPitchR); math.Abs(float64(d)-want) > 2e-3 {
				t.Errorf("%s far, at r %g: %g, want the pitch cylinder's %g", c.name, radii[i], d, want)
			}
		}
		near := pruneGPU(t, c.doc, [3]float64{0, 0, 20}, 0.001, pts)
		differs := false
		for i, d := range near {
			if math.Abs(float64(d)-c.sign*(radii[i]-threadPitchR)) > 0.05 {
				differs = true
			}
		}
		if !differs {
			t.Errorf("%s near: the thread was pruned too, every point reads the pitch cylinder", c.name)
		}
	}
}
