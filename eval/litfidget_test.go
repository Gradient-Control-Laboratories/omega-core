//go:build cgo

package eval

import (
	"math"
	"sort"
	"strings"
	"testing"
)

// litTwoReturnsDoc is a cube and a ball beside it, two returns.
const litTwoReturnsDoc = "boxHalf(half) {\n  return max(\n    -half.x - p.x, p.x - half.x,\n" +
	"    -half.y - p.y, p.y - half.y,\n    -half.z - p.z, p.z - half.z,\n  );\n}\n" +
	"cube = boxHalf([4.0, 4.0, 4.0]);\n" +
	"ball = length(p - [8.0, 0.0, 0.0]) - 3.0;\n" +
	"return cube, ball;\n"

// The fidget device finds the same picture as the sphere tracer: the same
// pixels hit, the same return under them and the same depth, up to fidget's
// voxels, from a perspective and from an orthographic camera.
func TestRenderLitFidgetDeviceMatchesTheTracer(t *testing.T) {
	prog := litProgram(t, litTwoReturnsDoc)
	const w, h = 120, 80
	ortho := &LitCamera{
		Pos: [3]float64{2, 30, 0}, Right: [3]float64{-1, 0, 0}, Up: [3]float64{0, 0, 1}, Fwd: [3]float64{0, -1, 0},
		TanFov: [2]float64{18, 12}, Ortho: true,
	}
	for _, tc := range []struct {
		name string
		opts LitOptions
	}{
		{"perspective", LitOptions{Width: w, Height: h, WorldExtent: 14}},
		{"orthographic", LitOptions{Width: w, Height: h, WorldExtent: 14, Camera: ortho}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			trace, err := RenderLitLinear(prog, tc.opts)
			if err != nil {
				t.Fatal(err)
			}
			opts := tc.opts
			opts.Device = "fidget"
			fid, err := RenderLitLinear(prog, opts)
			if err != nil {
				t.Fatal(err)
			}

			// onOutline: within two pixels of where the tracer's coverage, or
			// the return it hits, changes.
			onOutline := func(x, y int) bool {
				at := trace.Shape[y*w+x]
				for dy := -2; dy <= 2; dy++ {
					for dx := -2; dx <= 2; dx++ {
						u, v := x+dx, y+dy
						if u >= 0 && u < w && v >= 0 && v < h && trace.Shape[v*w+u] != at {
							return true
						}
					}
				}
				return false
			}
			var traceHits, coverageDiff, inside, both, returnDiff int
			var depthErr []float64
			for i := 0; i < w*h; i++ {
				th, fh := trace.Shape[i] > 0, fid.Shape[i] > 0
				if th {
					traceHits++
				}
				if th != fh {
					coverageDiff++
					if !onOutline(i%w, i/w) {
						inside++
					}
					continue
				}
				if !th {
					continue
				}
				both++
				if trace.Shape[i] != fid.Shape[i] {
					returnDiff++
				}
				depthErr = append(depthErr, math.Abs(float64(trace.Depth[i]-fid.Depth[i])))
			}
			if traceHits < w*h/20 {
				t.Fatalf("the tracer hit only %d pixels", traceHits)
			}
			// The outline is where the two may disagree: a pixel there holds the
			// surface for one ray (the tracer counts half a pixel's distance as a
			// hit) and misses it for one column of voxels. Inside a shape or the
			// background they must not.
			t.Logf("%d of %d hit pixels differ in coverage, all within two pixels of an outline but %d", coverageDiff, traceHits, inside)
			if inside > 0 {
				t.Errorf("%d pixels away from any outline differ in coverage", inside)
			}
			if returnDiff > both/100 {
				t.Errorf("%d of %d common hits name a different return", returnDiff, both)
			}
			sort.Float64s(depthErr)
			if med := depthErr[len(depthErr)/2]; med > 0.1 {
				t.Errorf("median depth difference %g, want under 0.1 (a scene 28 across)", med)
			}
		})
	}
}

// The fidget device is named, and it draws neither the Simple frame nor the
// viewer's march.
func TestRenderLitDeviceChecks(t *testing.T) {
	prog := litProgram(t, litBallDoc)
	cam := &LitCamera{Pos: [3]float64{10, 10, 10}, Right: [3]float64{0.7071, -0.7071, 0}, Up: [3]float64{-0.4082, -0.4082, 0.8165},
		Fwd: [3]float64{-0.5774, -0.5774, -0.5774}, TanFov: [2]float64{0.7, 0.466}}
	for _, tc := range []struct {
		opts LitOptions
		want string
	}{
		{LitOptions{Width: 40, Height: 30, WorldExtent: 6, Device: "gpu"}, "unknown lit device"},
		{LitOptions{Width: 40, Height: 30, WorldExtent: 6, Device: "fidget", Camera: cam, Simple: &LitSimple{Steps: 10}}, "Simple"},
		{LitOptions{Width: 40, Height: 30, WorldExtent: 6, Device: "fidget", Camera: cam, ViewerMarch: true}, "march"},
	} {
		_, err := RenderLitLinear(prog, tc.opts)
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("device %q: error %v, want one mentioning %q", tc.opts.Device, err, tc.want)
		}
	}
}
