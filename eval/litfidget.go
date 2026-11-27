//go:build cgo

// The lit render's fidget device (LitOptions.Device "fidget"): each return is
// voxel-rendered by fidget on the CPU, the images merged keeping the nearest
// hit tagged with its return, and fidget's SSAO taken over the merged depth.
// That replaces the sphere tracer's march, shape identification, normals and
// occlusion taps; the appearance, the glass and the lighting after it are
// marchLit's own. It is the CPU twin of the viewer's Fidget device
// (examples/viewer/render/fidget-gpu.js, which runs the same steps on the GPU
// through internal/fidget/internal/wgpu-gbuf), and its camera matrix and
// normal transform are ported from there, so the two hit the same voxels.

package eval

import (
	"fmt"
	"math"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/fidget"
)

// The devices LitOptions.Device names.
const (
	litDeviceTrace  = "trace"
	litDeviceFidget = "fidget"
)

// litDevice checks opts.Device and says whether it asks for fidget.
func litDevice(opts LitOptions) (fidgetDevice bool, err error) {
	switch opts.Device {
	case "", litDeviceTrace:
		return false, nil
	case litDeviceFidget:
		if opts.Simple != nil {
			return false, fmt.Errorf("the fidget device does not draw the Simple frame")
		}
		if opts.ViewerMarch {
			return false, fmt.Errorf("the fidget device replaces the march, so it does not take the viewer's")
		}
		return true, nil
	}
	return false, fmt.Errorf("unknown lit device %q (want %q or %q)", opts.Device, litDeviceTrace, litDeviceFidget)
}

// litFidgetCamera is the camera marchLit built, as the fidget device needs it.
type litFidgetCamera struct {
	pos, right, up, fwd [3]float64
	tanFovX, tanFovY    float64
	ortho               bool
}

// litFidgetHits is what the fidget device finds at each pixel: whether it hit,
// which return, the world point, the world normal (facing the camera), the
// distance along the pixel's ray, the occlusion factor (1 for none) and the
// back-face dimming.
type litFidgetHits struct {
	isHit   []bool
	shapeOf []int
	pos     [][3]float64
	normal  [][3]float64
	t       []float64
	ao      []float64
	dim     []float64
}

// litFidgetRender voxel-renders every shape canvas, merges them, takes the
// SSAO and turns fidget's pixels into world hits. origins and dirs are the
// pixels' rays, lookAt and boxHalf the scene box the trace device clips its
// rays to; fidget's columns span the geometry inside that box along the view
// (union is the union of the returns, which the frame scan reads). pixelSize, when the
// program reads it, is bound to the footprint of a pixel at the box's centre.
func litFidgetRender(
	union fidget.CanvasIndex,
	shapeCanvases []fidget.CanvasIndex,
	cam litFidgetCamera,
	w, h int,
	origins, dirs [][3]float64,
	lookAt [3]float64,
	boxHalf, pixelScale, backDim float64,
	pixelSize litPixelSize,
	progress func(g float64),
) (*litFidgetHits, error) {
	d := max(w, h)
	// The columns span the geometry, as the frame scan finds it, widened by
	// one of the scan's cells, inside the scene box.
	var fit *[2][3]float64
	if box, ok := occupiedBox(union); ok {
		var b [2][3]float64
		for k := 0; k < 3; k++ {
			pad := float64(box[1][k]-box[0][k]) / frameScanRes
			b[0][k] = float64(box[0][k]) - pad
			b[1][k] = float64(box[1][k]) + pad
		}
		fit = &b
	}
	near, far := litFidgetDepthRange(cam, lookAt, boxHalf, fit)
	m := litFidgetWorldToModel(cam, w, h, d, near, far)
	var m32 [16]float32
	for i, v := range m {
		m32[i] = float32(v)
	}

	var varTrees []fidget.TreeIndex
	var varValues []float32
	if pixelSize.ok {
		dist := 1.0
		if !cam.ortho {
			dist = math.Max(length3(sub3(lookAt, cam.pos)), 1)
		}
		varTrees = []fidget.TreeIndex{pixelSize.tree}
		varValues = []float32{float32(pixelScale * dist)}
	}

	n := w * h
	merged := &fidget.VoxelFrame{Width: w, Height: h, Depth: d, Z: make([]uint32, n), Normal: make([]float32, 3*n)}
	index := make([]int, n)
	for i := range index {
		index[i] = -1
	}
	for k, canvas := range shapeCanvases {
		f, err := fidget.RenderVoxels(canvas, w, h, d, m32, varTrees, varValues)
		if err != nil {
			return nil, fmt.Errorf("fidget render of return %d: %w", k+1, err)
		}
		// The nearest hit wins (fidget's depth grows toward the viewer); a tie
		// keeps the earlier return, as the GPU merge does.
		for i, z := range f.Z {
			if z > merged.Z[i] {
				merged.Z[i] = z
				copy(merged.Normal[3*i:3*i+3], f.Normal[3*i:3*i+3])
				index[i] = k
			}
		}
		if progress != nil {
			progress(float64(k+1) / float64(len(shapeCanvases)))
		}
	}
	occlusion, err := fidget.VoxelSSAO(merged)
	if err != nil {
		return nil, err
	}

	hits := &litFidgetHits{
		isHit: make([]bool, n), shapeOf: make([]int, n), pos: make([][3]float64, n),
		normal: make([][3]float64, n), t: make([]float64, n), ao: make([]float64, n), dim: make([]float64, n),
	}
	k := 2 / float64(min(w, h, d))
	for py := 0; py < h; py++ {
		for px := 0; px < w; px++ {
			i := py*w + px
			hits.ao[i], hits.dim[i], hits.t[i] = 1, 1, math.Inf(1)
			if index[i] < 0 {
				continue
			}
			s := [3]float64{(float64(px) - float64(w)/2) * k, -(float64(py) - float64(h)/2 + 1) * k, (float64(merged.Z[i]) - float64(d)/2) * k}
			jac, world := litFidgetJacobian(m, s)
			// fidget's gradient is in its pixels, y down; in s it is that over k,
			// y flipped.
			gs := [3]float64{float64(merged.Normal[3*i]) / k, -float64(merged.Normal[3*i+1]) / k, float64(merged.Normal[3*i+2]) / k}
			nw := litFidgetSolveTranspose(jac, gs)
			if math.IsNaN(nw[0]+nw[1]+nw[2]) || math.IsInf(nw[0]+nw[1]+nw[2], 0) || dot3(nw, nw) == 0 {
				nw = scale3(dirs[i], -1)
			}
			nw = normalize3(nw)
			// A face seen from behind: the normal turned to the camera and the
			// colour dimmed, as the viewer's Fidget device lights it.
			if dot3(dirs[i], nw) > 0 {
				nw = scale3(nw, -1)
				hits.dim[i] = backDim
			}
			hits.isHit[i] = true
			hits.shapeOf[i] = index[i]
			hits.pos[i] = world
			hits.normal[i] = nw
			hits.t[i] = dot3(sub3(world, origins[i]), dirs[i])
			if occ := float64(occlusion[i]); !math.IsNaN(occ) && !math.IsInf(occ, 0) && occ >= 0 {
				hits.ao[i] = occ
			}
		}
	}
	return hits, nil
}

// litFidgetDepthRange is where fidget's voxel columns start and stop along the
// view: the depths of the scene box's corners. Perspective slices are uniform
// in inverse depth, so a range reaching back to the camera would crowd every
// slice beside it; a camera inside the range starts a hundredth of the way to
// its far side, as the viewer's does.
func litFidgetDepthRange(cam litFidgetCamera, lookAt [3]float64, boxHalf float64, fit *[2][3]float64) (near, far float64) {
	nearD, farD := math.Inf(1), math.Inf(-1)
	for _, sx := range []float64{-1, 1} {
		for _, sy := range []float64{-1, 1} {
			for _, sz := range []float64{-1, 1} {
				c := [3]float64{lookAt[0] + sx*boxHalf, lookAt[1] + sy*boxHalf, lookAt[2] + sz*boxHalf}
				dd := dot3(sub3(c, cam.pos), cam.fwd)
				nearD = math.Min(nearD, dd)
				farD = math.Max(farD, dd)
			}
		}
	}
	if fit != nil {
		fn, ff := math.Inf(1), math.Inf(-1)
		for _, x := range []float64{fit[0][0], fit[1][0]} {
			for _, y := range []float64{fit[0][1], fit[1][1]} {
				for _, z := range []float64{fit[0][2], fit[1][2]} {
					dd := dot3(sub3([3]float64{x, y, z}, cam.pos), cam.fwd)
					fn = math.Min(fn, dd)
					ff = math.Max(ff, dd)
				}
			}
		}
		nearD, farD = math.Max(nearD, fn), math.Min(farD, ff)
	}
	far = math.Max(farD, 1e-6)
	near = nearD
	if !cam.ortho {
		near = math.Max(nearD, far*0.01)
	}
	return near, far
}

// litFidgetWorldToModel takes fidget's screen coordinates to world space: a
// pixel (px, py) at voxel depth pz sits at ((px - W/2) k, -(py - H/2 + 1) k,
// (pz - D/2) k), k = 2 / min(W, H, D), with z toward the viewer. Those are
// rescaled to the camera's [-1, 1] square at pixel centres, then taken along
// the view: linearly for an orthographic camera, and for a perspective one by
// a projective map whose w is the inverse depth, near at z = 1 and far at
// z = -1. Column-major, as fidget reads it; worldToModel in fidget-gpu.js.
func litFidgetWorldToModel(cam litFidgetCamera, w, h, d int, near, far float64) [16]float64 {
	W, H, D := float64(w), float64(h), float64(d)
	k := 2 / float64(min(w, h, d))
	tx, ty := cam.tanFovX, cam.tanFovY
	R, U, F, pos := cam.right, cam.up, cam.fwd, cam.pos
	// N maps (u, v, w, 1) to homogeneous world space, columns first.
	var N [4][4]float64
	N[0] = [4]float64{tx * R[0], tx * R[1], tx * R[2], 0}
	N[1] = [4]float64{ty * U[0], ty * U[1], ty * U[2], 0}
	if cam.ortho {
		half := (far - near) / 2
		N[2] = [4]float64{-half * F[0], -half * F[1], -half * F[2], 0}
		N[3] = [4]float64{pos[0] + (near+half)*F[0], pos[1] + (near+half)*F[1], pos[2] + (near+half)*F[2], 1}
	} else {
		a := (1/near - 1/far) / 2
		b := (1/near + 1/far) / 2
		N[2] = [4]float64{a * pos[0], a * pos[1], a * pos[2], a}
		N[3] = [4]float64{F[0] + b*pos[0], F[1] + b*pos[1], F[2] + b*pos[2], b}
	}
	// S maps fidget's screen coordinates to (u, v, w): u = 2 sx / (W k) + 1/W,
	// v = 2 sy / (H k) + 1/H, w = 2 sz / (D k).
	su, sv, sw := 2/(W*k), 2/(H*k), 2/(D*k)
	var m [16]float64
	for r := 0; r < 4; r++ {
		m[0+r] = N[0][r] * su
		m[4+r] = N[1][r] * sv
		m[8+r] = N[2][r] * sw
		m[12+r] = N[0][r]/W + N[1][r]/H + N[3][r]
	}
	return m
}

// litFidgetJacobian is d(world)/d(s) at the screen point s, for world =
// (m s).xyz / (m s).w, and the world point itself.
func litFidgetJacobian(m [16]float64, s [3]float64) (jac [3][3]float64, world [3]float64) {
	var p [4]float64
	for r := 0; r < 4; r++ {
		p[r] = m[r]*s[0] + m[4+r]*s[1] + m[8+r]*s[2] + m[12+r]
	}
	wq := p[3]
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			jac[i][j] = (m[4*j+i] - (p[i]/wq)*m[4*j+3]) / wq
		}
		world[i] = p[i] / wq
	}
	return jac, world
}

// litFidgetSolveTranspose returns x with J^T x = g: the world gradient from
// the screen-space one, by Cramer's rule.
func litFidgetSolveTranspose(jac [3][3]float64, g [3]float64) [3]float64 {
	var a [3][3]float64
	for r := 0; r < 3; r++ {
		for c := 0; c < 3; c++ {
			a[r][c] = jac[c][r]
		}
	}
	det := func(M [3][3]float64) float64 {
		return M[0][0]*(M[1][1]*M[2][2]-M[1][2]*M[2][1]) -
			M[0][1]*(M[1][0]*M[2][2]-M[1][2]*M[2][0]) +
			M[0][2]*(M[1][0]*M[2][1]-M[1][1]*M[2][0])
	}
	dt := det(a)
	var x [3]float64
	for c := 0; c < 3; c++ {
		b := a
		for r := 0; r < 3; r++ {
			b[r][c] = g[r]
		}
		x[c] = det(b) / dt
	}
	return x
}
