// Options that make RenderLit draw the way the viewer's standard frame does,
// so the two can be compared pixel for pixel (the image-parity test in
// internal/gpuparity). All are off by default; with them off, RenderLit is
// unchanged.
package mesh

import (
	"fmt"
	"math"

	"gradientcontrol.com/omega/core/internal/fidget/internal/fidget"
	"gradientcontrol.com/omega/core/program"
)

// LitCamera places the camera the way the viewer's camera uniforms do, in
// world Z-up: a position, an orthonormal basis, and the tangents of the half
// field of view. For an orthographic camera TanFov holds the half-width and
// half-height of the view instead, and rays leave the image plane parallel.
type LitCamera struct {
	Pos, Right, Up, Fwd [3]float64
	TanFov              [2]float64
	Ortho               bool
}

// LitLinear is a lit render before display encoding.
type LitLinear struct {
	Width, Height int
	// RGB is the tone-mapped colour, three per pixel, row-major from the top
	// left: the value the viewer's shape pass writes. Zero where Shape is 0.
	RGB []float32
	// Shape is the return each pixel shows, numbered from 1 as the WGSL's
	// shapeIndex is; 0 where the ray hit nothing.
	Shape []int32
	// Glass marks pixels shaded through a transmitting material, whose
	// sampling differs between the renderers by design.
	Glass []bool
	// Depth is how far along the view direction the hit lies from where the
	// ray started, as the viewer measures it for its depth. Zero where Shape
	// is 0.
	Depth []float32
	// Radiance is an opaque hit's colour before tone mapping, linear light,
	// three per pixel as RGB is: what a caller recolouring a render works in,
	// since shading is linear in the albedo there and the tone map is not.
	// Zero where Shape is 0 or the pixel is Glass.
	Radiance []float32
}

// RenderLitLinear is RenderLit returning the image before display encoding,
// with the shape each pixel shows. The world axes are not drawn.
func RenderLitLinear(prog *program.Program, opts LitOptions) (*LitLinear, error) {
	opts.NoAxes = true
	var lin *LitLinear
	if err := renderLit(prog, opts, func(_ []byte, l *LitLinear) { lin = l }); err != nil {
		return nil, err
	}
	return lin, nil
}

// RenderLitWithLinear is RenderLit returning the linear image beside the PNG,
// from the one render.
func RenderLitWithLinear(prog *program.Program, opts LitOptions) ([]byte, *LitLinear, error) {
	var png []byte
	var lin *LitLinear
	if err := renderLit(prog, opts, func(p []byte, l *LitLinear) { png, lin = p, l }); err != nil {
		return nil, nil, err
	}
	return png, lin, nil
}

// The viewer's march constants (gsl_wgslFragment in common/web/gsl-render-graph.js).
const (
	viewGradEps  = 0.0001
	viewHitEps   = 0.001
	viewFudge    = 0.8
	viewMaxSteps = 1024
	// viewJumpSlack is how much steeper than the gradient last measured the
	// field may be between two samples and still cross zero there; past it, a
	// sign change is a step across 0 rather than a surface (litViewerMarch).
	viewJumpSlack = 4.0
)

// viewBackDim is how bright the viewer draws a surface met from behind with the
// Render panel's Back faces at Dim, its default.
const viewBackDim = 0.25

// litBackDim is a request's BackDim, where zero is the switch's default.
func litBackDim(v float64) float64 {
	if v <= 0 {
		return viewBackDim
	}
	return v
}

// viewTet is the tetrahedron the viewer takes its step gradient on.
var viewTet = [4][3]float64{{1, -1, -1}, {-1, -1, 1}, {-1, 1, -1}, {1, 1, 1}}

// litComposite is the viewer's gsl_min over every shape at once: per point,
// the shape whose surface is nearest (smallest |d|), its signed distance and
// gradient. gsl_min keeps the earlier shape only when it is strictly nearer,
// so a tie goes to the later one; a NaN never wins.
func litComposite(canvases []fidget.CanvasIndex, xs, ys, zs []float32) (d []float64, idx []int, grad [][3]float64, err error) {
	return litCompositeVars(canvases, xs, ys, zs, nil, nil)
}

// litCompositeVars is litComposite with input vars bound per point: values[k]
// holds varTrees[k]'s value at each point.
func litCompositeVars(canvases []fidget.CanvasIndex, xs, ys, zs []float32, varTrees []fidget.TreeIndex, values [][]float32) (d []float64, idx []int, grad [][3]float64, err error) {
	n := len(xs)
	d = make([]float64, n)
	idx = make([]int, n)
	grad = make([][3]float64, n)
	for j := range d {
		d[j] = math.Inf(1)
	}
	for si, c := range canvases {
		var samples []fidget.GradResult
		if len(varTrees) > 0 {
			samples, err = fidget.EvalGradVars(c, xs, ys, zs, varTrees, values)
		} else {
			samples, err = fidget.EvalGrad(c, xs, ys, zs)
		}
		if err != nil {
			return nil, nil, nil, err
		}
		for j, s := range samples {
			v := float64(s.Value)
			if math.IsNaN(v) {
				continue
			}
			if math.Abs(v) <= math.Abs(d[j]) {
				d[j] = v
				idx[j] = si
				grad[j] = [3]float64{float64(s.Dx), float64(s.Dy), float64(s.Dz)}
			}
		}
	}
	return d, idx, grad, nil
}

func litCompositeAt(canvases []fidget.CanvasIndex, pts [][3]float64) ([]float64, []int, error) {
	return litCompositeProxied(canvases, pts, litPartProxy{}, nil)
}

// litPartProxy is the input a document's parts read as their boxes beyond
// (convert's partProxy), when it has a part whose solid is in its box, and
// whether the field it marches is displaced, which widens the viewer's
// radius.
type litPartProxy struct {
	tree      fidget.TreeIndex
	ok        bool
	displaced bool
}

// litDispNearMargin is the viewer's gsl_dispNearMargin: how near a surface
// its march displaces the field.
const litDispNearMargin = 4.0

// radius is the viewer's gsl_vanishView.w for a ray leaving the scene box at
// tExit (gsl-render-graph.js): four pixels at the ray's far end and never
// under eight hit tolerances, plus displacement's reach (gsl_dispNearMargin,
// 4) where the field is displaced.
func (pp litPartProxy) radius(pixelScale, ortho, tExit float64) float64 {
	r := math.Max(viewHitEps*8, pixelScale*(math.Max(tExit, 1.0)*(1-ortho)+ortho)*4)
	if pp.displaced {
		r += litDispNearMargin
	}
	return r
}

// litCompositeProxied is litCompositeAt with each point's parts read as their
// boxes beyond radii[j], as the viewer's march reads them; with no proxy, or
// radii nil, every part is whole.
func litCompositeProxied(canvases []fidget.CanvasIndex, pts [][3]float64, pp litPartProxy, radii []float64) ([]float64, []int, error) {
	xs := make([]float32, len(pts))
	ys := make([]float32, len(pts))
	zs := make([]float32, len(pts))
	for j, p := range pts {
		xs[j], ys[j], zs[j] = float32(p[0]), float32(p[1]), float32(p[2])
	}
	var varTrees []fidget.TreeIndex
	var values [][]float32
	if pp.ok && radii != nil {
		r := make([]float32, len(pts))
		for j := range r {
			r[j] = float32(radii[j])
		}
		varTrees, values = []fidget.TreeIndex{pp.tree}, [][]float32{r}
	}
	d, idx, _, err := litCompositeVars(canvases, xs, ys, zs, varTrees, values)
	return d, idx, err
}

// litViewerMarch is the viewer's primary march, ray by ray in lockstep:
// step |d| over the tetrahedral gradient's magnitude (at least 1) times 0.8,
// hit within max(0.001, half a pixel), and bisect back when the same shape's
// distance changes sign between steps. The step budget grows with the ray's
// path through the scene box as the viewer's does. Fills isHit, surfT,
// hitPos and shapeOf; normals are taken afterwards (litViewerNormals).
func litViewerMarch(
	canvases, cheap []fidget.CanvasIndex,
	origins, dirs [][3]float64,
	tStart, tExit []float64,
	active []bool,
	pixelScale, ortho, extent float64,
	proxy litPartProxy,
	isHit []bool, surfT []float64, hitPos [][3]float64, shapeOf []int,
) error {
	n := len(origins)
	t := make([]float64, n)
	prevT := make([]float64, n)
	prevD := make([]float64, n)
	prevIdx := make([]int, n)
	prevGrad := make([]float64, n)
	maxSteps := make([]int, n)
	radius := make([]float64, n)
	pos := func(i int, tt float64) [3]float64 {
		return [3]float64{origins[i][0] + dirs[i][0]*tt, origins[i][1] + dirs[i][1]*tt, origins[i][2] + dirs[i][2]*tt}
	}
	pixelWorld := func(tt float64) float64 {
		return pixelScale * (math.Max(tt, 1.0)*(1-ortho) + ortho)
	}

	var live []int
	for i := 0; i < n; i++ {
		if !active[i] {
			continue
		}
		t[i] = math.Max(tStart[i], 0)
		prevT[i] = t[i]
		prevGrad[i] = 1
		path := clamp(math.Max(tExit[i]-t[i], 0)/math.Max(2*extent, 1e-6), 1.0, 1.75)
		maxSteps[i] = int(float32(clamp(256*path, 256, 1024)))
		radius[i] = proxy.radius(pixelScale, ortho, tExit[i])
		live = append(live, i)
	}
	// Each point's radius, for the rays in order: once per point, or per
	// tetrahedral tap.
	radiiOf := func(rays []int, per int) []float64 {
		r := make([]float64, 0, len(rays)*per)
		for _, i := range rays {
			for k := 0; k < per; k++ {
				r = append(r, radius[i])
			}
		}
		return r
	}
	// near marks a ray whose last centre was within displacement's reach
	// (gsl_dispNear): its taps are displaced too.
	near := make([]bool, n)
	// evalAt is the composite the viewer's march reads at pts, the rays' own
	// points (per 1) or their tetrahedral taps (per 4). Where the field is
	// displaced the viewer reads the plain composite, and displaces it only
	// within gsl_dispNearMargin of a surface, by the centre's distance, so far
	// from one, where a part reads as its box, the box is not displaced either.
	evalAt := func(pts [][3]float64, rays []int, per int) ([]float64, []int, error) {
		radii := radiiOf(rays, per)
		if cheap == nil {
			return litCompositeProxied(canvases, pts, proxy, radii)
		}
		d, idx, err := litCompositeProxied(cheap, pts, proxy, radii)
		if err != nil {
			return nil, nil, err
		}
		var sub [][3]float64
		var subRadii []float64
		var at []int
		for j := range pts {
			i := rays[j/per]
			if per == 1 {
				near[i] = math.Abs(d[j]) < litDispNearMargin
			}
			if near[i] {
				sub = append(sub, pts[j])
				subRadii = append(subRadii, radii[j])
				at = append(at, j)
			}
		}
		if len(sub) == 0 {
			return d, idx, nil
		}
		dd, di, err := litCompositeProxied(canvases, sub, proxy, subRadii)
		if err != nil {
			return nil, nil, err
		}
		for k, j := range at {
			d[j], idx[j] = dd[k], di[k]
		}
		return d, idx, nil
	}
	// prevSd: the composite where the ray enters.
	{
		pts := make([][3]float64, len(live))
		for j, i := range live {
			pts[j] = pos(i, t[i])
		}
		d, idx, err := evalAt(pts, live, 1)
		if err != nil {
			return fmt.Errorf("march eval: %w", err)
		}
		for j, i := range live {
			prevD[i], prevIdx[i] = d[j], idx[j]
		}
	}

	for step := 0; step < viewMaxSteps && len(live) > 0; step++ {
		next := live[:0:0]
		for _, i := range live {
			if step < maxSteps[i] && t[i] <= tExit[i] {
				next = append(next, i)
			}
		}
		live = next
		if len(live) == 0 {
			break
		}
		pts := make([][3]float64, len(live))
		for j, i := range live {
			pts[j] = pos(i, t[i])
		}
		d, idx, err := evalAt(pts, live, 1)
		if err != nil {
			return fmt.Errorf("march eval: %w", err)
		}

		// Rays that neither hit nor crossed take a gradient step.
		var stepping []int
		var ges []float64
		var tetPts [][3]float64
		next = live[:0:0]
		for j, i := range live {
			hitEps := math.Max(viewHitEps, pixelWorld(t[i])*0.5)
			if math.Abs(d[j]) < hitEps {
				isHit[i] = true
				surfT[i] = t[i]
				hitPos[i] = pts[j]
				shapeOf[i] = idx[j]
				continue
			}
			next = append(next, i)
			// A sign change of the same shape is an overshoot to bisect back
			// from, unless the samples are too far from zero for the field
			// to cross it between them: then it steps across 0 there (a cut
			// part's hollow, a ternary switching sides), and the ray steps on.
			crossed := idx[j] == prevIdx[i] && (d[j] < 0) != (prevD[i] < 0) && t[i] > prevT[i] &&
				math.Abs(prevD[i])+math.Abs(d[j]) <= viewJumpSlack*prevGrad[i]*(t[i]-prevT[i])
			if crossed {
				t[i] = (prevT[i] + t[i]) * 0.5
				continue
			}
			// Never under about four float32 ulps of the point itself, as the
			// viewer's (gsl-render-graph.js).
			ge := math.Max(math.Max(viewGradEps, pixelWorld(t[i])*0.025), length3(pts[j])*4e-7)
			stepping = append(stepping, j)
			ges = append(ges, ge)
			for _, k := range viewTet {
				tetPts = append(tetPts, [3]float64{pts[j][0] + k[0]*ge, pts[j][1] + k[1]*ge, pts[j][2] + k[2]*ge})
			}
		}
		if len(stepping) > 0 {
			tapRays := make([]int, len(stepping))
			for s, j := range stepping {
				tapRays[s] = live[j]
			}
			td, _, err := evalAt(tetPts, tapRays, 4)
			if err != nil {
				return fmt.Errorf("march gradient eval: %w", err)
			}
			for s, j := range stepping {
				i := live[j]
				var g [3]float64
				for k := 0; k < 4; k++ {
					g = add3(g, scale3(viewTet[k], td[s*4+k]))
				}
				gradMag := math.Max(length3(g)/(4*ges[s]), 1.0)
				prevT[i] = t[i]
				prevD[i], prevIdx[i] = d[j], idx[j]
				prevGrad[i] = gradMag
				t[i] += math.Abs(d[j]) / gradMag * viewFudge
			}
		}
		live = next
	}
	return nil
}

// litViewerNormals is gsl_calcNormal on the composite: central differences
// with a step of a twentieth of a pixel (at least 1e-4), then the viewer's
// facing-away rule: a surface the ray meets from behind is shaded with the
// normal flipped toward the camera and dimmed to backDim, and marked behind.
func litViewerNormals(
	canvases []fidget.CanvasIndex,
	hitIdx []int,
	hitPos, dirs [][3]float64,
	surfT []float64,
	pixelScale, ortho, backDim float64,
	hitNormal [][3]float64, dim []float64, behind []bool,
) error {
	if len(hitIdx) == 0 {
		return nil
	}
	pts := make([][3]float64, 0, 6*len(hitIdx))
	for _, i := range hitIdx {
		e := math.Max(viewGradEps, pixelScale*(math.Max(surfT[i], 1.0)*(1-ortho)+ortho)*0.05)
		e = math.Max(e, 1e-6)
		p := hitPos[i]
		for k := 0; k < 3; k++ {
			var o [3]float64
			o[k] = e
			pts = append(pts, add3(p, o), sub3(p, o))
		}
	}
	d, _, err := litCompositeAt(canvases, pts)
	if err != nil {
		return fmt.Errorf("normal eval: %w", err)
	}
	for j, i := range hitIdx {
		b := j * 6
		n := normalize3([3]float64{d[b] - d[b+1], d[b+2] - d[b+3], d[b+4] - d[b+5]})
		if math.IsNaN(n[0] + n[1] + n[2]) {
			n = scale3(dirs[i], -1)
		}
		if dot3(dirs[i], n) > 0 {
			n = scale3(n, -1)
			dim[i] = backDim
			behind[i] = true
		}
		hitNormal[i] = n
	}
	return nil
}

// LitEnvironment is the background and ambient hemisphere RenderLit shades
// with by default (the viewer's dark canvas), for a renderer that has to
// match it.
func LitEnvironment() (background, sky, ground [3]float64) {
	return litDark.background, litDark.sky, litDark.ground
}
