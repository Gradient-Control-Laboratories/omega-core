//go:build cgo

// The Simple render tier on the CPU: the spec in common/web/gsl-simple.js,
// whose Simple frame the viewers draw on the GPU. The caller passes the
// spec's constants (the browser sends them with each request), so
// gsl-simple.js stays their one copy; the march, the composite and the
// headlight below follow its description step for step.
package eval

import (
	"fmt"
	"math"

	"gradientcontrol.com/omega/core/internal/fidget"
)

// litSimpleRender draws the Simple frame: a plain sphere trace on march (the
// min of |d| over the returns), then, at each hit, the composite's return and
// its tetrahedral normal, lit by a headlight on the return's palette colour
// and dimmed to backDim where the ray meets it from behind.
func litSimpleRender(
	march fidget.CanvasIndex,
	canvases []fidget.CanvasIndex,
	cam LitCamera,
	w, h int,
	extent float64,
	palette [][3]float64,
	spec LitSimple,
	backDim float64,
) (*LitLinear, error) {
	if len(palette) == 0 {
		return nil, fmt.Errorf("the Simple frame needs a palette")
	}
	n := w * h
	ortho := 0.0
	if cam.Ortho {
		ortho = 1
	}
	pixelScale := 2 * cam.TanFov[1] / float64(h)
	pixelWorld := func(t float64) float64 { return pixelScale * (math.Max(t, 1)*(1-ortho) + ortho) }

	// The rays, clipped to the scene box as the WGSL clips them.
	origins := make([][3]float64, n)
	dirs := make([][3]float64, n)
	ts := make([]float64, n)
	exits := make([]float64, n)
	var live []int
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*w + x
			ndcX := (float64(x)+0.5)/float64(w)*2 - 1
			ndcY := 1 - (float64(y)+0.5)/float64(h)*2
			var d [3]float64
			for k := 0; k < 3; k++ {
				off := ndcX*cam.TanFov[0]*cam.Right[k] + ndcY*cam.TanFov[1]*cam.Up[k]
				origins[i][k] = cam.Pos[k] + off*ortho
				d[k] = cam.Fwd[k] + off*(1-ortho)
			}
			dirs[i] = normalize3(d)
			entry, exit := math.Inf(-1), math.Inf(1)
			for k := 0; k < 3; k++ {
				inv := 1 / dirs[i][k]
				t1 := (-extent - origins[i][k]) * inv
				t2 := (extent - origins[i][k]) * inv
				entry = math.Max(entry, math.Min(t1, t2))
				exit = math.Min(exit, math.Max(t1, t2))
			}
			if entry <= exit && exit > 0 {
				ts[i] = math.Max(entry, 0)
				exits[i] = exit
				live = append(live, i)
			}
		}
	}

	// The march: every live ray in one bulk evaluation per step.
	var hits []int
	for step := 0; step < spec.Steps && len(live) > 0; step++ {
		next := live[:0:0]
		for _, i := range live {
			if !(ts[i] > exits[i]) {
				next = append(next, i)
			}
		}
		live = next
		if len(live) == 0 {
			break
		}
		xs := make([]float32, len(live))
		ys := make([]float32, len(live))
		zs := make([]float32, len(live))
		for j, i := range live {
			xs[j] = float32(origins[i][0] + dirs[i][0]*ts[i])
			ys[j] = float32(origins[i][1] + dirs[i][1]*ts[i])
			zs[j] = float32(origins[i][2] + dirs[i][2]*ts[i])
		}
		samples, err := fidget.EvalGrad(march, xs, ys, zs)
		if err != nil {
			return nil, fmt.Errorf("simple march eval: %w", err)
		}
		next = live[:0:0]
		for j, i := range live {
			d := float64(samples[j].Value)
			if d < math.Max(spec.HitFloor, pixelWorld(ts[i])*spec.HitPixel) {
				hits = append(hits, i)
				continue
			}
			ts[i] += d * spec.Fudge
			next = append(next, i)
		}
		live = next
	}

	// At each hit: the point and its four tetrahedral neighbours, through the
	// composite, for the return and the normal.
	pts := make([][3]float64, 0, 5*len(hits))
	for _, i := range hits {
		p := [3]float64{origins[i][0] + dirs[i][0]*ts[i], origins[i][1] + dirs[i][1]*ts[i], origins[i][2] + dirs[i][2]*ts[i]}
		e := math.Max(spec.NormalFloor, pixelWorld(ts[i])*spec.NormalPixel)
		pts = append(pts, p)
		for _, k := range viewTet {
			pts = append(pts, [3]float64{p[0] + k[0]*e, p[1] + k[1]*e, p[2] + k[2]*e})
		}
	}
	d, idx, err := litCompositeAt(canvases, pts)
	if err != nil {
		return nil, fmt.Errorf("simple composite eval: %w", err)
	}

	lin := &LitLinear{
		Width:  w,
		Height: h,
		RGB:    make([]float32, 3*n),
		Shape:  make([]int32, n),
		Glass:  make([]bool, n),
		Depth:  make([]float32, n),
	}
	for hi, i := range hits {
		var g [3]float64
		for k := 0; k < 4; k++ {
			g = add3(g, scale3(viewTet[k], d[5*hi+1+k]))
		}
		nrm := normalize3(g)
		shape := idx[5*hi] + 1
		base := palette[(shape-1)%len(palette)]
		lit := spec.Ambient + (1-spec.Ambient)*math.Abs(dot3(nrm, dirs[i]))
		if dot3(nrm, dirs[i]) > 0 {
			lit *= backDim
		}
		for c := 0; c < 3; c++ {
			lin.RGB[3*i+c] = float32(base[c] * lit)
		}
		lin.Shape[i] = int32(shape)
		lin.Depth[i] = float32(ts[i] * dot3(dirs[i], cam.Fwd))
	}
	return lin, nil
}
