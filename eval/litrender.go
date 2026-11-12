//go:build cgo

// Lit CPU rendering: a sphere-traced preview that reproduces the viewers'
// shading — the same three-light rig in linear light, the metallic-roughness
// BRDF, hemisphere ambient scaled by Quilez ambient occlusion, and the
// document's own authored appearance fields (name.color = …, name.material =
// …) — without a GPU or a browser.
//
// The march and the shading run in Go over bulk tape evaluations (one
// evaluation per live ray per step, through the JIT), so the picture comes
// from exactly the field the geometry tools query and certify.
//
// The shading below is a port of gsl_shadeSurface in
// common/web/gsl-shading.js, which the three viewers share; that module is the
// original and this is the copy, so when it is retuned or a material row is
// added, the constants and the BRDF here are what must follow. Transparency and
// refraction are ported likewise from common/web/gsl-glass.js, in litglass.go.
// One renderer feature the viewers have is deliberately not ported — see the
// note on litMaterial.
package eval

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"sync"

	"gradientcontrol.com/omega/core/internal/fidget"
	"gradientcontrol.com/omega/core/internal/fidget/convert"
	"gradientcontrol.com/omega/core/program"
	"gradientcontrol.com/omega/core/program/noise"
)

// RenderLitSized renders prog on the CPU at the given dimensions and world extent.
func RenderLitSized(
	prog *program.Program,
	width, height int,
	worldExtent, timeAt float64,
) ([]byte, error) {
	return RenderLit(prog, LitOptions{
		Width:       width,
		Height:      height,
		WorldExtent: worldExtent,
		Time:        timeAt,
	})
}

// RenderLitSizedView renders prog on the CPU with a chosen camera and ground.
func RenderLitSizedView(
	prog *program.Program,
	width, height int,
	worldExtent, timeAt float64,
	view LitView,
) ([]byte, error) {
	return RenderLit(prog, LitOptions{
		Width:           width,
		Height:          height,
		WorldExtent:     worldExtent,
		Time:            timeAt,
		View:            view.Dir,
		Up:              view.Up,
		Light:           view.Light,
		NoAxes:          view.NoAxes,
		Environment:     view.Environment,
		EnvironmentTurn: view.EnvironmentTurn,
		GlassSamples:    view.GlassSamples,
	})
}

// RenderLitSizedViewRadiance is RenderLitSizedView returning each pixel's linear radiance beside the PNG.
func RenderLitSizedViewRadiance(
	prog *program.Program,
	width, height int,
	worldExtent, timeAt float64,
	view LitView,
) ([]byte, []float32, error) {
	png, lin, err := RenderLitWithLinear(prog, LitOptions{
		Width:           width,
		Height:          height,
		WorldExtent:     worldExtent,
		Time:            timeAt,
		View:            view.Dir,
		Up:              view.Up,
		Light:           view.Light,
		NoAxes:          view.NoAxes,
		Environment:     view.Environment,
		EnvironmentTurn: view.EnvironmentTurn,
		GlassSamples:    view.GlassSamples,
	})
	if err != nil {
		return nil, nil, err
	}
	return png, lin.Radiance, nil
}

// litGround is one choice of background and ambient hemisphere, and the
// analytic environment that replaces the hemisphere when a preset chooses one
// (nil for the viewer's own lighting).
type litGround struct {
	background, sky, ground [3]float64
	env                     *litEnv
}

// The calibration knobs, from SHADING_DEFAULTS in common/web/gsl-shading.js.
// In the viewer these ride in a uniform so they can be tuned in the running
// app; here they are the settled values.
const (
	litLightScale      = 2.0  // key-light radiance (the Lambert lobe divides by pi)
	litAmbientScale    = 1.25 // hemisphere fill, diffuse and specular
	litExposure        = 1.75 // applied before emission and tone mapping
	litWarmLightHeight = -2.0 // view-space height of the warm waist light

	// The shoulder lights are fixed in the shader, not knobs.
	litCoolLightOffset = 0.5 // horizontal shoulder separation (view X)
	litCoolLightHeight = 0.6 // height above eye for the shoulder lights (view Y)
)

// Shading constants, transcribed from gsl_shadeLinear.
var (
	litCool1Dir  = normalize3([3]float64{-litCoolLightOffset, litCoolLightHeight, 0.5})
	litCool2Dir  = normalize3([3]float64{litCoolLightOffset, litCoolLightHeight, 0.5})
	litWarmDir   = normalize3([3]float64{0.0, litWarmLightHeight, 1.0})
	litCoolCol   = scale3([3]float64{0.618, 0.684, 0.855}, 0.75*litLightScale) // cool blue-white
	litWarmCol   = scale3([3]float64{0.600, 0.420, 0.200}, litLightScale)      // warm amber, waist
	litSkyCol    = [3]float64{0.15, 0.13, 0.12}
	litGroundCol = [3]float64{0.06, 0.07, 0.1}
	// Background 0x1e1e1e, the viewer's canvas clear color.
	litBackground = [3]float64{30.0 / 255.0, 30.0 / 255.0, 30.0 / 255.0}

	// litDark is the viewer's ground: the constants above, as one value.
	litDark = litGround{background: litBackground, sky: litSkyCol, ground: litGroundCol}
	// litLight is the studio ground for presentation renders: a near-white
	// canvas and an ambient hemisphere bright enough that faces turned away
	// from the key lights still read as surface rather than silhouette.
	litLight = litGround{
		background: [3]float64{0.93, 0.93, 0.94},
		sky:        [3]float64{0.32, 0.31, 0.30},
		ground:     [3]float64{0.16, 0.17, 0.20},
	}
)

// litPalette is the fallback color for a shape that authors no color field,
// standing in for the viewer's per-shape UI palette.
var litPalette = [][3]float64{
	{0.85, 0.85, 0.88},
	{0.80, 0.55, 0.45},
	{0.55, 0.70, 0.85},
	{0.70, 0.75, 0.55},
	{0.75, 0.60, 0.75},
	{0.60, 0.75, 0.72},
}

var (
	litNoiseOnce sync.Once
	litNoiseErr  error
)

const (
	litMaxSteps = 192
	litFudge    = 0.8
	litAOSteps  = 5
)

// The rows of a material record, indexed as the compiler lowered them and as
// the gsl_material struct in common/web/gsl-shading.js names them. Rows are
// appended, never reordered.
const (
	litMatSurface  = 0 // roughness, metallic, ior, reserved
	litMatEmissive = 1 // sRGB r, g, b, strength
	litMatVolume   = 2 // transmission, absorptionDistance, displacement, reserved
	litMatLay      = 3 // lay x, y, z, anisotropy
)

// litMaterial is the CPU mirror of the viewer's gsl_material struct: one vec4
// row per vec4 of the GSL name.material list.
//
// Every row is rendered: roughness, metallic, ior and emission by shadeLit,
// transmission and its absorption distance by the refracted march in
// litglass.go, and displacement by the field the march steps on
// (litDisplaced).
type litMaterial [4][4]float64

// litDefaultMaterial is the record a shape has when the document authors no
// material: roughness 0.5, not metallic, ior 1.5, no emission, opaque,
// undisplaced. It is the viewer's present = 0 case, and must match
// GSL_MATERIAL_DEFAULT_WGSL, property.materialDefault and material.standard.
var litDefaultMaterial = litMaterial{{0.5, 0.0, 1.5, 0.0}, {}, {}, {}}

// RenderLit sphere-traces the program and returns a PNG shaded like the
// viewer. Colors come from the document's color fields where it authors
// them, and from a fallback palette where it does not.
func RenderLit(prog *program.Program, opts LitOptions) ([]byte, error) {
	if opts.Simple != nil {
		return nil, fmt.Errorf("RenderLit: the Simple frame is drawn through RenderLitLinear only")
	}
	var png []byte
	if err := renderLit(prog, opts, func(p []byte, _ *LitLinear) { png = p }); err != nil {
		return nil, err
	}
	return png, nil
}

// renderLit renders and hands the PNG and the linear image to done.
func renderLit(prog *program.Program, opts LitOptions, done func([]byte, *LitLinear)) error {
	if opts.Width <= 0 {
		opts.Width = 900
	}
	if opts.Height <= 0 {
		opts.Height = 600
	}
	// The noiseHash3 oracle samples a fixed baked tile that survives
	// fidget.Clear(); a document using noise evaluates to NaN without it.
	litNoiseOnce.Do(func() {
		if err := fidget.RegisterNoiseTile(noise.Bake()); err != nil {
			litNoiseErr = err
		}
	})
	if litNoiseErr != nil {
		return fmt.Errorf("register noise tile: %w", litNoiseErr)
	}
	if opts.Camera != nil && opts.WorldExtent <= 0 {
		return fmt.Errorf("a lit camera needs the scene extent")
	}
	if _, err := litEnvironment(opts.Environment, 0, [3]float64{1, 0, 0}); err != nil {
		return err
	}

	if opts.Simple != nil && opts.Camera == nil {
		return fmt.Errorf("the Simple frame needs a camera")
	}

	err := convert.ToColorShapesWithPixelSize(prog, opts.Time, func(shapes []convert.ColorShape, pixelSizeTree fidget.TreeIndex, hasPixelSize bool, partProxyTree fidget.TreeIndex, hasPartProxy bool) error {
		// Color draws without the program's materials, and Simple without its
		// colours too.
		if opts.NoMaterials || opts.Simple != nil {
			for i := range shapes {
				shapes[i].HasMaterial = false
				if opts.Simple != nil {
					shapes[i].HasColor = false
				}
			}
		}
		// The field each shape is rendered with: its distance, moved by its
		// material's displacement where it has one (litDisplaced).
		renderTrees := make([]fidget.TreeIndex, len(shapes))
		for i, s := range shapes {
			renderTrees[i] = litDisplaced(s)
		}

		// The union drives the march; per-shape canvases identify which shape
		// a ray hit, and the color canvases paint it.
		unionTree := renderTrees[0]
		for i := 1; i < len(renderTrees); i++ {
			unionTree = fidget.Min(unionTree, renderTrees[i])
		}
		union := fidget.BuildCanvas(unionTree)
		shapeCanvases := make([]fidget.CanvasIndex, len(shapes))
		colorCanvases := make([][3]fidget.CanvasIndex, len(shapes))
		materialCanvases := make([][4][4]fidget.CanvasIndex, len(shapes))
		for i, s := range shapes {
			shapeCanvases[i] = fidget.BuildCanvas(renderTrees[i])
			if s.HasColor {
				for c := 0; c < 3; c++ {
					colorCanvases[i][c] = fidget.BuildCanvas(s.Color[c])
				}
			}
			if s.HasMaterial {
				for r := range s.Material {
					for c := range s.Material[r] {
						materialCanvases[i][r][c] = fidget.BuildCanvas(s.Material[r][c])
					}
				}
			}
		}

		if opts.Simple != nil {
			marchTree := fidget.Abs(renderTrees[0])
			for i := 1; i < len(renderTrees); i++ {
				marchTree = fidget.Min(marchTree, fidget.Abs(renderTrees[i]))
			}
			lin, err := litSimpleRender(fidget.BuildCanvas(marchTree), shapeCanvases, *opts.Camera,
				opts.Width, opts.Height, opts.WorldExtent, opts.Colors, *opts.Simple, litBackDim(opts.BackDim))
			if err != nil {
				return err
			}
			done(nil, lin)
			return nil
		}

		// The camera: a declared extent keeps the viewer's convention — look
		// at the origin from (e, e, e) with the scene box that size. With no
		// declared extent, frame the geometry itself: center the look-at on
		// its bounding box and fit its bounding radius to the 50-degree
		// vertical field of view (camera distance e*sqrt(3), so e = 1.45r
		// leaves ~17% margin). The earlier fallback reused frameCanvas's
		// half-height as if it were a camera distance, which cropped anything
		// taller than it was wide — a correctly stood-up part rendered as a
		// clipped ambiguous wall.
		extent := opts.WorldExtent
		var lookAt [3]float64
		boxHalf := extent
		if extent <= 0 && opts.Camera == nil {
			if center, scale, ok := frameCanvas(union); ok {
				radius := float64(scale) / 1.25 // frameCanvas pads its half-height
				extent = 1.45 * radius
				boxHalf = 1.15 * radius
				for k := 0; k < 3; k++ {
					lookAt[k] = float64(center[k])
				}
			}
			if extent <= 0 {
				extent, boxHalf = 10, 10
			}
		}

		pixelSize := litPixelSize{tree: pixelSizeTree, ok: hasPixelSize}
		// The viewer widens its parts' radius by displacement's reach wherever
		// the shader has a material field (displaceActive).
		proxy := litPartProxy{tree: partProxyTree, ok: hasPartProxy}
		for _, s := range shapes {
			proxy.displaced = proxy.displaced || s.HasMaterial
		}
		img, lin, err := marchLit(union, renderTrees, shapeCanvases, colorCanvases, materialCanvases, shapes, pixelSize, proxy, extent, boxHalf, lookAt, opts)
		if err != nil {
			return err
		}
		png, err := encodePNG(img)
		if err != nil {
			return err
		}
		done(png, lin)
		return nil
	})
	return err
}

// marchLit is the renderer proper: build the camera, sphere-trace every pixel
// against the union field, then shade the hits.
func marchLit(
	union fidget.CanvasIndex,
	renderTrees []fidget.TreeIndex,
	shapeCanvases []fidget.CanvasIndex,
	colorCanvases [][3]fidget.CanvasIndex,
	materialCanvases [][4][4]fidget.CanvasIndex,
	shapes []convert.ColorShape,
	pixelSize litPixelSize,
	proxy litPartProxy,
	extent, boxHalf float64,
	lookAt [3]float64,
	opts LitOptions,
) (*image.RGBA, *LitLinear, error) {
	w, h := opts.Width, opts.Height
	aspect := float64(w) / float64(h)

	// Camera: the viewer's corner-on framing by default — position
	// lookAt + (e, e, e), looking at lookAt, Z up, 50-degree vertical field
	// of view. lookAt is the origin under a declared extent, matching the
	// viewer. A declared View keeps the same camera distance, e·sqrt(3), along
	// its own direction; a declared Up reorients the image.
	viewDir := [3]float64{1, 1, 1}
	if opts.View != [3]float64{} {
		viewDir = opts.View
	}
	worldUp := [3]float64{0, 0, 1}
	if opts.Up != [3]float64{} {
		worldUp = opts.Up
	}
	var camPos [3]float64
	if opts.View == [3]float64{} {
		camPos = [3]float64{lookAt[0] + extent, lookAt[1] + extent, lookAt[2] + extent}
	} else {
		unit := normalize3(viewDir)
		dist := extent * math.Sqrt(3)
		for k := 0; k < 3; k++ {
			camPos[k] = lookAt[k] + unit[k]*dist
		}
	}
	fwd := normalize3([3]float64{-viewDir[0] * extent, -viewDir[1] * extent, -viewDir[2] * extent})
	rightRaw := cross3(fwd, worldUp)
	if dot3(rightRaw, rightRaw) < 1e-12 {
		// Looking straight along the up direction: the image has no
		// preferred rotation, so take +Y as the viewer would.
		rightRaw = cross3(fwd, [3]float64{0, 1, 0})
	}
	right := normalize3(rightRaw)
	up := cross3(right, fwd)
	tanFovY := math.Tan(50.0 * math.Pi / 180.0 / 2.0)
	tanFovX := tanFovY * aspect
	ortho := 0.0
	if c := opts.Camera; c != nil {
		camPos, right, up, fwd = c.Pos, c.Right, c.Up, c.Fwd
		tanFovX, tanFovY = c.TanFov[0], c.TanFov[1]
		if c.Ortho {
			ortho = 1
		}
	}
	ground := litDark
	if opts.Light {
		ground = litLight
	}
	env, err := litEnvironment(opts.Environment, opts.EnvironmentTurn, right)
	if err != nil {
		return nil, nil, err
	}
	ground.env = env
	palette := litPalette
	if opts.Colors != nil {
		palette = opts.Colors
	}

	// zUpToView: the lights are defined in view space, so shading needs the
	// world normal expressed in the camera basis (x right, y up, z toward
	// the viewer).
	toView := func(v [3]float64) [3]float64 {
		return [3]float64{dot3(v, right), dot3(v, up), -dot3(v, fwd)}
	}

	n := w * h
	origins := make([][3]float64, n)
	dirs := make([][3]float64, n)
	tv := make([]float64, n)
	tExit := make([]float64, n)
	active := make([]bool, n)
	// pixelScale mirrors the viewer's: the world size of one pixel at unit
	// distance, which sets the hit epsilon so detail resolves at any zoom.
	pixelScale := 2.0 * tanFovY / float64(h)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			i := y*w + x
			ndcX := (float64(x)+0.5)/float64(w)*2.0 - 1.0
			ndcY := 1.0 - (float64(y)+0.5)/float64(h)*2.0
			d := [3]float64{}
			for k := 0; k < 3; k++ {
				d[k] = fwd[k] + ndcX*tanFovX*right[k] + ndcY*tanFovY*up[k]
			}
			origins[i] = camPos
			if ortho > 0 {
				// Parallel rays from the image plane.
				for k := 0; k < 3; k++ {
					origins[i][k] += ndcX*tanFovX*right[k] + ndcY*tanFovY*up[k]
				}
				d = fwd
			}
			dirs[i] = normalize3(d)
			enter, exit, hit := rayBoxAt(origins[i], dirs[i], lookAt, boxHalf)
			active[i] = hit
			tv[i] = math.Max(enter, 0)
			tExit[i] = exit
		}
	}

	hitPos := make([][3]float64, n)
	hitNormal := make([][3]float64, n)
	isHit := make([]bool, n)
	surfT := make([]float64, n)
	for i := range surfT {
		surfT[i] = math.Inf(1)
	}

	// ViewerMarch replaces the march and the shape identification below with
	// the viewer's (litview.go); the loop then runs no steps.
	marchSteps := litMaxSteps
	viewShape := make([]int, n)
	if opts.ViewerMarch {
		marchSteps = 0
		// The plain fields, which the viewer marches until it is near a
		// surface, where any shape is displaced.
		var cheapCanvases []fidget.CanvasIndex
		for k, s := range shapes {
			if renderTrees[k] != s.Tree {
				cheapCanvases = make([]fidget.CanvasIndex, len(shapes))
				for m := range shapes {
					cheapCanvases[m] = fidget.BuildCanvas(shapes[m].Tree)
				}
				break
			}
		}
		if err := litViewerMarch(shapeCanvases, cheapCanvases, origins, dirs, tv, tExit, active, pixelScale, ortho, boxHalf, proxy,
			isHit, surfT, hitPos, viewShape); err != nil {
			return nil, nil, err
		}
	}

	// March: every step evaluates the union field at all still-live rays in
	// one bulk call, so the cost per step is one JIT pass, not one per pixel.
	idx := make([]int, 0, n)
	for step := 0; step < marchSteps; step++ {
		idx = idx[:0]
		for i := 0; i < n; i++ {
			if active[i] {
				idx = append(idx, i)
			}
		}
		if len(idx) == 0 {
			break
		}
		xs := make([]float32, len(idx))
		ys := make([]float32, len(idx))
		zs := make([]float32, len(idx))
		for j, i := range idx {
			xs[j] = float32(origins[i][0] + dirs[i][0]*tv[i])
			ys[j] = float32(origins[i][1] + dirs[i][1]*tv[i])
			zs[j] = float32(origins[i][2] + dirs[i][2]*tv[i])
		}
		samples, err := fidget.EvalGrad(union, xs, ys, zs)
		if err != nil {
			return nil, nil, fmt.Errorf("march eval: %w", err)
		}
		for j, i := range idx {
			s := samples[j]
			d := float64(s.Value)
			// A non-finite sample would make every comparison below false and
			// the step NaN, so the ray would neither hit nor terminate — it
			// would burn the whole step budget. Drop it instead.
			if math.IsNaN(d) || math.IsInf(d, 0) {
				active[i] = false
				continue
			}
			// The hit epsilon grows with distance so a pixel's worth of
			// surface always counts as a hit — the viewer's rule.
			pixelWorld := pixelScale * math.Max(tv[i], 1.0)
			hitEps := math.Max(1e-4, pixelWorld*0.5)
			if math.Abs(d) < hitEps {
				isHit[i] = true
				surfT[i] = tv[i]
				active[i] = false
				hitPos[i] = [3]float64{float64(xs[j]), float64(ys[j]), float64(zs[j])}
				g := [3]float64{float64(s.Dx), float64(s.Dy), float64(s.Dz)}
				if math.IsNaN(g[0]+g[1]+g[2]) || math.IsInf(g[0]+g[1]+g[2], 0) {
					// Face the camera rather than shading from a NaN normal.
					g = [3]float64{-dirs[i][0], -dirs[i][1], -dirs[i][2]}
				}
				hitNormal[i] = normalize3(g)
				continue
			}
			// Normalizing the step by the gradient magnitude keeps the march
			// sound for fields that are not 1-Lipschitz (TPMS, warped
			// remaps), where a raw distance step would tunnel through. A
			// smooth combinator's gradient can go non-finite far from its
			// surface even where the value is exact (unionRound at range);
			// fall back to 1, the plain distance step, rather than poisoning
			// the ray's position with a NaN it can never recover from.
			gradMag := length3([3]float64{float64(s.Dx), float64(s.Dy), float64(s.Dz)})
			if math.IsNaN(gradMag) || math.IsInf(gradMag, 0) || gradMag < 1 {
				gradMag = 1
			}
			tv[i] += d / gradMag * litFudge
			if tv[i] > tExit[i] {
				active[i] = false
			}
		}
	}

	// Identify the shape under each hit: the argmin over the per-shape
	// fields, which is what the union's min selected.
	hitIdx := make([]int, 0, n)
	for i := 0; i < n; i++ {
		if isHit[i] {
			hitIdx = append(hitIdx, i)
		}
	}
	shapeOf := make([]int, n)
	dim := make([]float64, n)
	for i := range dim {
		dim[i] = 1
	}
	behind := make([]bool, n)
	if opts.ViewerMarch {
		copy(shapeOf, viewShape)
		if err := litViewerNormals(shapeCanvases, hitIdx, hitPos, dirs, surfT, pixelScale, ortho, litBackDim(opts.BackDim), hitNormal, dim, behind); err != nil {
			return nil, nil, err
		}
	}
	if len(hitIdx) > 0 && len(shapes) > 1 && !opts.ViewerMarch {
		hx, hy, hz := gatherPoints(hitPos, hitIdx)
		best := make([]float64, len(hitIdx))
		for j := range best {
			best[j] = math.Inf(1)
		}
		for si, canvas := range shapeCanvases {
			samples, err := fidget.EvalGrad(canvas, hx, hy, hz)
			if err != nil {
				return nil, nil, fmt.Errorf("shape id eval: %w", err)
			}
			for j, s := range samples {
				if v := float64(s.Value); v < best[j] {
					best[j] = v
					shapeOf[hitIdx[j]] = si
				}
			}
		}
	}

	// The march meets the full relief, as the viewer's does by default (its
	// Fine relief "At hit"), but the viewer sets pixelSize() to the hit's
	// footprint before it takes the normal and the occlusion, so relief too
	// fine to see shades faded rather than aliasing. Take the normal again
	// with the footprint bound, and the occlusion taps below with it too.
	var hitFP []float32
	if pixelSize.ok && len(hitIdx) > 0 && !opts.ViewerMarch {
		hitFP = make([]float32, len(hitIdx))
		for j, i := range hitIdx {
			hitFP[j] = float32(pixelScale * (math.Max(surfT[i], 1.0)*(1-ortho) + ortho))
		}
		hx, hy, hz := gatherPoints(hitPos, hitIdx)
		samples, err := fidget.EvalGradVars(union, hx, hy, hz, []fidget.TreeIndex{pixelSize.tree}, [][]float32{hitFP})
		if err != nil {
			return nil, nil, fmt.Errorf("hit normal eval: %w", err)
		}
		for j, i := range hitIdx {
			g := [3]float64{float64(samples[j].Dx), float64(samples[j].Dy), float64(samples[j].Dz)}
			if !math.IsNaN(g[0]+g[1]+g[2]) && !math.IsInf(g[0]+g[1]+g[2], 0) && dot3(g, g) > 0 {
				hitNormal[i] = normalize3(g)
			}
		}
	}

	// Ambient occlusion, Quilez: five taps along the normal, evaluated for
	// every hit at once per tap.
	aoFactor := make([]float64, n)
	for i := range aoFactor {
		aoFactor[i] = 1
	}
	if len(hitIdx) > 0 {
		ao := make([]float64, len(hitIdx))
		aoScale := 2.0
		for step := 0; step < litAOSteps; step++ {
			hStep := 0.01 + 0.12*float64(step)/float64(litAOSteps-1)
			xs := make([]float32, len(hitIdx))
			ys := make([]float32, len(hitIdx))
			zs := make([]float32, len(hitIdx))
			for j, i := range hitIdx {
				xs[j] = float32(hitPos[i][0] + hitNormal[i][0]*hStep)
				ys[j] = float32(hitPos[i][1] + hitNormal[i][1]*hStep)
				zs[j] = float32(hitPos[i][2] + hitNormal[i][2]*hStep)
			}
			var samples []fidget.GradResult
			var err error
			if hitFP != nil {
				samples, err = fidget.EvalGradVars(union, xs, ys, zs, []fidget.TreeIndex{pixelSize.tree}, [][]float32{hitFP})
			} else {
				samples, err = fidget.EvalGrad(union, xs, ys, zs)
			}
			if err != nil {
				return nil, nil, fmt.Errorf("ao eval: %w", err)
			}
			if opts.ViewerMarch {
				// The viewer's AO taps read the composite, as its march does.
				cd, _, _, err := litComposite(shapeCanvases, xs, ys, zs)
				if err != nil {
					return nil, nil, fmt.Errorf("ao eval: %w", err)
				}
				for j := range samples {
					samples[j].Value = float32(cd[j])
				}
			}
			for j, s := range samples {
				// A face met from behind is seen from inside its solid; the
				// field negated reads the space the camera is in as open, as
				// the viewer's AO does.
				v := float64(s.Value)
				if behind[hitIdx[j]] {
					v = -v
				}
				d := math.Max(0, v)
				ao[j] += (hStep - d) * aoScale
			}
			aoScale *= 0.95
		}
		for j, i := range hitIdx {
			aoFactor[i] = clamp(1.0-3.0*ao[j], 0, 1)
		}
	}

	// Surface appearance at every hit: the authored color and material where
	// the shape has them, the palette and the default record where it does
	// not.
	baseColor := make([][3]float64, n)
	matAt := make([]litMaterial, n)
	// Each hit pixel's footprint, the world size of one pixel at its depth,
	// which the viewer sets as gsl_pw at the hit and pixelSize() reads.
	footprint := make([]float64, n)
	for _, i := range hitIdx {
		footprint[i] = pixelScale * (math.Max(surfT[i], 1.0)*(1-ortho) + ortho)
	}
	if len(hitIdx) > 0 {
		pts := make([][3]float64, len(hitIdx))
		hitShape := make([]int, len(hitIdx))
		hitFootprint := make([]float64, len(hitIdx))
		for j, i := range hitIdx {
			pts[j] = hitPos[i]
			hitShape[j] = shapeOf[i]
			hitFootprint[j] = footprint[i]
		}
		cols, mats, err := litAppearance(pts, hitShape, shapes, colorCanvases, materialCanvases, palette, pixelSize, hitFootprint)
		if err != nil {
			return nil, nil, err
		}
		for j, i := range hitIdx {
			baseColor[i] = cols[j]
			matAt[i] = mats[j]
		}
	}

	// Glass: a hit whose material transmits is shaded by marching the refracted
	// ray (see litglass.go), which replaces its opaque shading below.
	glass, err := litShadeGlass(litGlassScene{
		glassSamples:     opts.GlassSamples,
		shapes:           shapes,
		renderTrees:      renderTrees,
		shapeCanvases:    shapeCanvases,
		colorCanvases:    colorCanvases,
		materialCanvases: materialCanvases,
		lookAt:           lookAt,
		boxHalf:          boxHalf,
		pixelScale:       pixelScale,
		camPos:           camPos,
		toView:           toView,
		ground:           ground,
		palette:          palette,
		pixelSize:        pixelSize,
		footprint:        footprint,
	}, hitIdx, w, h, hitPos, hitNormal, baseColor, matAt, aoFactor)
	if err != nil {
		return nil, nil, err
	}

	// World-axis overlay: thin capsules along +x, +y, +z from the world
	// origin, occluded by the geometry. Without a frame cue, a render of a
	// free-floating part carries no orientation information at all — a lying
	// slab and a standing slab framed to fit look alike — which was the
	// audit finding that motivated this renderer's framing fix.
	axisLen := boxHalf
	axisR := boxHalf * 0.008
	axisCols := [3][3]float64{{0.85, 0.30, 0.30}, {0.30, 0.72, 0.30}, {0.32, 0.52, 0.90}}
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	lin := &LitLinear{Width: w, Height: h, RGB: make([]float32, 3*n), Shape: make([]int32, n), Glass: make([]bool, n),
		Depth: make([]float32, n), Radiance: make([]float32, 3*n)}
	setLinear := func(i int, c [3]float64) {
		for k := 0; k < 3; k++ {
			lin.RGB[3*i+k] = float32(c[k])
		}
		lin.Shape[i] = int32(shapeOf[i] + 1)
		lin.Depth[i] = float32(dot3(sub3(hitPos[i], origins[i]), fwd))
	}
	for i := 0; i < n; i++ {
		var rgb [3]float64
		if g, ok := glass[i]; ok {
			setLinear(i, g)
			lin.Glass[i] = true
			rgb = litEncodeDisplay(g)
		} else if isHit[i] {
			// shadeLit, with the radiance kept before it is tone-mapped.
			radiance := shadeLitLinear(baseColor[i], hitNormal[i],
				normalize3(sub3(camPos, hitPos[i])), matAt[i], aoFactor[i], toView, ground)
			if dim[i] != 1 {
				// gsl_shadeSurface's dimFactor: the radiance, not the display
				// value, is scaled.
				radiance = scale3(radiance, dim[i])
			}
			c := litTonemapPbrNeutral(radiance)
			setLinear(i, c)
			for k := 0; k < 3; k++ {
				lin.Radiance[3*i+k] = float32(radiance[k])
			}
			rgb = litEncodeDisplay(c)
		} else {
			rgb = ground.background
		}
		bestT := surfT[i]
		for k := 0; k < 3 && !opts.NoAxes; k++ {
			var end [3]float64
			end[k] = axisLen
			if t := capIntersect(origins[i], dirs[i], [3]float64{}, end, axisR); t > 0 && t < bestT {
				bestT = t
				rgb = axisCols[k]
			}
		}
		img.Set(i%w, i/w, color.RGBA{
			R: toByte(rgb[0]), G: toByte(rgb[1]), B: toByte(rgb[2]), A: 255,
		})
	}
	return img, lin, nil
}

// litPixelSize is the input var pixelSize() was built as, when the program
// calls it (convert.ToColorShapesWithPixelSize).
type litPixelSize struct {
	tree fidget.TreeIndex
	ok   bool
}

// litDisplaced is the field a shape is rendered with: gsl_sdDisplaced
// (common/web/gsl-displace.js), d'(p) = d(p) - h(p), where h is the height in
// the material's volume row. Positive height pushes the surface outward. The
// viewer's displacementScale knob defaults to 1, so the height is subtracted
// unscaled.
//
// Everything downstream follows from the one substitution: the march, the
// normal (the gradient is that of d - h, so relief shades with no separate
// bump term), the ambient occlusion taps, shape identification and the glass
// path all see the moved surface. Subtracting a height breaks the Lipschitz
// bound, which the march already absorbs by dividing its step by the
// gradient's magnitude. Render only, as in the viewer: meshing never comes
// through here, so what is manufactured does not move.
//
// Two differences from the viewer, both in this renderer's favour. The viewer
// subtracts the height of whichever shape wins the undisplaced composite, so
// two returns with different heights can show a seam where they meet; here
// each shape's own height moves its own field before the union, so there is
// none. And the viewer switches displacement on only when the material field
// references p, a cost shortcut that also drops a constant non-zero height
// (material.withDisplacement(m, 0.5) inflates nothing there); here the only
// skip is a height known to be the constant zero, which is every preset.
func litDisplaced(s convert.ColorShape) fidget.TreeIndex {
	if !s.HasMaterial || !s.Displaced {
		return s.Tree
	}
	return fidget.Sub(s.Tree, s.Material[litMatVolume][2])
}

// litAppearance looks up the surface colour and material at each point, given
// the shape each point lies on: the authored color and material fields where
// that shape has them, the fallback palette and the default record where it
// does not (the viewer's alpha-0 color and present = 0 material). Each field
// row is one bulk evaluation over only that shape's points, so a document with
// no authored fields costs nothing here. The primary hits and the glass path's
// secondary hits both come through it, so a surface seen through glass is
// coloured exactly as it is seen directly.
func litAppearance(
	pts [][3]float64,
	shapeOf []int,
	shapes []convert.ColorShape,
	colorCanvases [][3]fidget.CanvasIndex,
	materialCanvases [][4][4]fidget.CanvasIndex,
	palette [][3]float64,
	pixelSize litPixelSize,
	footprint []float64,
) ([][3]float64, []litMaterial, error) {
	colors := make([][3]float64, len(pts))
	mats := make([]litMaterial, len(pts))
	for j := range pts {
		colors[j] = palette[shapeOf[j]%len(palette)]
		mats[j] = litDefaultMaterial
	}
	for si, s := range shapes {
		if !s.HasColor && !s.HasMaterial {
			continue
		}
		var sub []int
		for j := range pts {
			if shapeOf[j] == si {
				sub = append(sub, j)
			}
		}
		if len(sub) == 0 {
			continue
		}
		sx, sy, sz := gatherPoints(pts, sub)
		// Where the program reads pixelSize(), bind it to each point's
		// footprint; elsewhere it keeps its default, 0.
		eval := func(c fidget.CanvasIndex) ([]fidget.GradResult, error) {
			if !pixelSize.ok || footprint == nil {
				return fidget.EvalGrad(c, sx, sy, sz)
			}
			fp := make([]float32, len(sub))
			for k, j := range sub {
				fp[k] = float32(footprint[j])
			}
			return fidget.EvalGradVars(c, sx, sy, sz, []fidget.TreeIndex{pixelSize.tree}, [][]float32{fp})
		}

		if s.HasColor {
			var rgb [3][]float64
			finite := make([]bool, len(sub))
			for k := range finite {
				finite[k] = true
			}
			for c := 0; c < 3; c++ {
				samples, err := eval(colorCanvases[si][c])
				if err != nil {
					return nil, nil, fmt.Errorf("color eval: %w", err)
				}
				rgb[c] = make([]float64, len(sub))
				for k, smp := range samples {
					v := float64(smp.Value)
					if math.IsNaN(v) || math.IsInf(v, 0) {
						// A blend weight can go degenerate far from the
						// surfaces it mixes (a smooth union's exponentials
						// underflow at range), leaving no usable color here.
						// Read that as "no field" and keep the palette, which
						// is how the viewer treats an absent color, rather
						// than painting the surface black.
						finite[k] = false
						continue
					}
					rgb[c][k] = clamp(v, 0, 1)
				}
			}
			for k, j := range sub {
				if finite[k] {
					colors[j] = [3]float64{rgb[0][k], rgb[1][k], rgb[2][k]}
				}
			}
		}

		if s.HasMaterial {
			vals := make([]litMaterial, len(sub))
			finite := make([]bool, len(sub))
			for k := range vals {
				vals[k] = litDefaultMaterial
				finite[k] = true
			}
			for r := range s.Material {
				for c := range s.Material[r] {
					samples, err := eval(materialCanvases[si][r][c])
					if err != nil {
						return nil, nil, fmt.Errorf("material eval: %w", err)
					}
					for k, smp := range samples {
						v := float64(smp.Value)
						if math.IsNaN(v) || math.IsInf(v, 0) {
							// A degenerate blend weight leaves no usable
							// material here, the same way it leaves no usable
							// color. Read it as "no field" and keep the default
							// record rather than shading from a NaN, which
							// would poison every lobe at once.
							finite[k] = false
							continue
						}
						vals[k][r][c] = v
					}
				}
			}
			for k, j := range sub {
				if finite[k] {
					mats[j] = vals[k]
				}
			}
		}
	}
	return colors, mats, nil
}

// shadeLit is gsl_shadeSurface: the tone-mapped colour for one hit. The viewer
// dims a face seen from behind by viewBackDim; this renderer shades only front faces,
// so that factor is 1 and the multiply is not written.
func shadeLit(
	base, normal, pos, camPos [3]float64,
	mat litMaterial,
	aoFactor float64,
	toView func([3]float64) [3]float64,
	ground litGround,
) [3]float64 {
	viewDirWorld := normalize3(sub3(camPos, pos))
	return litTonemapPbrNeutral(shadeLitLinear(base, normal, viewDirWorld, mat, aoFactor, toView, ground))
}

// shadeLitLinear is gsl_shadeLinear: the scene-referred radiance leaving the
// surface along viewDirWorld, direct and ambient light through the
// metallic-roughness BRDF, exposed, plus emission. Not tone mapped, so the
// glass path can sum it along a refracted ray before tone mapping once. It
// takes the direction rather than the camera because a surface seen through
// glass is viewed along the ray that reached it, not from the camera.
func shadeLitLinear(
	base, normal, viewDirWorld [3]float64,
	mat litMaterial,
	aoFactor float64,
	toView func([3]float64) [3]float64,
	ground litGround,
) [3]float64 {
	// Material, in linear light.
	albedo := litSrgbToLinear(base)
	emissive := scale3(
		litSrgbToLinear([3]float64{mat[litMatEmissive][0], mat[litMatEmissive][1], mat[litMatEmissive][2]}),
		mat[litMatEmissive][3])
	roughness := clamp(mat[litMatSurface][0], 0.045, 1.0) // floor keeps the GGX peak finite
	alpha2 := roughness * roughness * roughness * roughness
	metallic := clamp(mat[litMatSurface][1], 0.0, 1.0)
	ior := math.Max(mat[litMatSurface][2], 1.0)
	f0Dielectric := ((ior - 1.0) / (ior + 1.0)) * ((ior - 1.0) / (ior + 1.0))
	f0 := mix3([3]float64{f0Dielectric, f0Dielectric, f0Dielectric}, albedo, metallic)
	transmission := clamp(mat[litMatVolume][0], 0.0, 1.0)
	diffuseColor := scale3(albedo, (1.0-metallic)*(1.0-transmission))

	// The lights are defined in view space, so shade there.
	n := normalize3(toView(normal))
	v := normalize3(toView(viewDirWorld))
	noV := math.Max(dot3(n, v), 1e-4)
	// The material's lay, in the view frame for the rig and in the world's
	// for the reflection (litlay.go).
	layDir := [3]float64{mat[litMatLay][0], mat[litMatLay][1], mat[litMatLay][2]}
	layView, layAmount := litLayFrame(n, toView(layDir), mat[litMatLay][3])
	layWorld, _ := litLayFrame(normal, layDir, mat[litMatLay][3])

	// Direct light from the rig, at the strength an environment preset gives
	// it (litenv.go).
	coolCol, warmCol := litCoolCol, litWarmCol
	if env := ground.env; env != nil {
		coolCol = scale3(coolCol, env.rig)
		warmCol = scale3(warmCol, env.rig)
	}
	direct := litDirectLightLay(n, v, noV, litCool1Dir, coolCol, diffuseColor, f0, alpha2, roughness, layView, layAmount)
	direct = add3(direct, litDirectLightLay(n, v, noV, litCool2Dir, coolCol, diffuseColor, f0, alpha2, roughness, layView, layAmount))
	direct = add3(direct, litDirectLightLay(n, v, noV, litWarmDir, warmCol, diffuseColor, f0, alpha2, roughness, layView, layAmount))

	// Hemisphere ambient: sky above, ground below (world Z-up). Diffuse along
	// the normal; specular along the reflection, weighted by the environment
	// BRDF so metals do not go black. An environment preset replaces the
	// hemisphere with its own sky, blurred by roughness, and adds its panels.
	hemiN := scale3(mix3(ground.ground, ground.sky, normal[2]*0.5+0.5), litAmbientScale)
	r := litReflect(scale3(viewDirWorld, -1), litLayBentNormal(normal, viewDirWorld, layWorld, layAmount))
	hemiR := scale3(mix3(ground.ground, ground.sky, r[2]*0.5+0.5), litAmbientScale)
	var panels [3]float64
	if env := ground.env; env != nil {
		hemiN = env.diffuse(normal)
		hemiR = env.skySpecular(r, roughness)
		panels = env.panelSpecular(normal, viewDirWorld, f0, roughness)
	}
	fa := litFresnelSchlickRoughness(noV, f0, roughness)
	ambientDiffuse := mul3(hemiN, mul3(diffuseColor, sub3([3]float64{1, 1, 1}, fa)))
	ambientSpecular := add3(mul3(hemiR, litEnvBRDFApprox(f0, roughness, noV)), panels)
	ambient := scale3(add3(ambientDiffuse, ambientSpecular), aoFactor)

	return add3(scale3(add3(direct, ambient), litExposure), emissive)
}

// litDirectLight is gsl_directLight: one directional light's contribution,
// Lambert diffuse scaled by what the Fresnel term leaves plus the GGX / Smith /
// Schlick specular lobe.
func litDirectLight(n, v [3]float64, noV float64, l, radiance, diffuseColor, f0 [3]float64, alpha2 float64) [3]float64 {
	noL := dot3(n, l)
	if noL <= 0 {
		return [3]float64{}
	}
	h := normalize3(add3(l, v))
	noH := math.Max(dot3(n, h), 0)
	voH := math.Max(dot3(v, h), 0)
	d := litDistributionGGX(noH, alpha2)
	vis := litVisibilitySmithGGX(noV, noL, alpha2)
	f := litFresnelSchlick(voH, f0)
	specular := scale3(f, d*vis)
	diffuse := scale3(mul3(sub3([3]float64{1, 1, 1}, f), diffuseColor), 1.0/math.Pi)
	return scale3(mul3(add3(diffuse, specular), radiance), noL)
}

// litSrgbToLinear is the exact sRGB electro-optical transfer function, decoding
// an authored colour to the linear light the BRDF works in.
func litSrgbToLinear(c [3]float64) [3]float64 {
	var out [3]float64
	for k := 0; k < 3; k++ {
		if c[k] <= 0.04045 {
			out[k] = c[k] / 12.92
		} else {
			out[k] = math.Pow((c[k]+0.055)/1.055, 2.4)
		}
	}
	return out
}

// litEncodeDisplay is the sRGB opto-electronic transfer function: the encode
// Three performs at the canvas. The viewers shade into a LINEAR render target
// and the composite pass encodes on the way out (outputColorSpace is
// SRGBColorSpace there), so writing the shaded values straight into a PNG here
// would produce a picture markedly darker than the viewer draws. The
// background and the axis overlay are authored as display values already and
// are written unencoded.
func litEncodeDisplay(c [3]float64) [3]float64 {
	var out [3]float64
	for k := 0; k < 3; k++ {
		v := clamp(c[k], 0, 1)
		if v <= 0.0031308 {
			out[k] = v * 12.92
		} else {
			out[k] = 1.055*math.Pow(v, 1.0/2.4) - 0.055
		}
	}
	return out
}

// litDistributionGGX is the GGX / Trowbridge-Reitz normal distribution.
func litDistributionGGX(noH, alpha2 float64) float64 {
	f := noH*noH*(alpha2-1.0) + 1.0
	return alpha2 / (math.Pi * f * f)
}

// litVisibilitySmithGGX is the height-correlated Smith visibility for GGX,
// already divided by 4 NoL NoV.
func litVisibilitySmithGGX(noV, noL, alpha2 float64) float64 {
	ggxV := noL * math.Sqrt(noV*noV*(1.0-alpha2)+alpha2)
	ggxL := noV * math.Sqrt(noL*noL*(1.0-alpha2)+alpha2)
	return 0.5 / math.Max(ggxV+ggxL, 1e-5)
}

// litFresnelSchlick is Schlick's Fresnel approximation.
func litFresnelSchlick(voH float64, f0 [3]float64) [3]float64 {
	f := math.Pow(1.0-voH, 5.0)
	return add3(f0, scale3(sub3([3]float64{1, 1, 1}, f0), f))
}

// litFresnelSchlickRoughness is Schlick's Fresnel with the roughness-aware
// ceiling, for ambient light.
func litFresnelSchlickRoughness(noV float64, f0 [3]float64, roughness float64) [3]float64 {
	f := math.Pow(1.0-noV, 5.0)
	var out [3]float64
	for k := 0; k < 3; k++ {
		out[k] = f0[k] + (math.Max(1.0-roughness, f0[k])-f0[k])*f
	}
	return out
}

// litEnvBRDFApprox is Karis' analytic fit of the split-sum environment BRDF:
// the scale and bias applied to F0 for a given roughness and viewing angle.
func litEnvBRDFApprox(f0 [3]float64, roughness, noV float64) [3]float64 {
	c0 := [4]float64{-1.0, -0.0275, -0.572, 0.022}
	c1 := [4]float64{1.0, 0.0425, 1.04, -0.04}
	var r [4]float64
	for k := 0; k < 4; k++ {
		r[k] = roughness*c0[k] + c1[k]
	}
	a004 := math.Min(r[0]*r[0], math.Exp2(-9.28*noV))*r[0] + r[1]
	abX := -1.04*a004 + r[2]
	abY := 1.04*a004 + r[3]
	var out [3]float64
	for k := 0; k < 3; k++ {
		out[k] = f0[k]*abX + abY
	}
	return out
}

// litTonemapPbrNeutral is the Khronos PBR Neutral tone mapper: identity up to
// 0.8, then a hue-preserving shoulder that desaturates toward white as it
// clips.
func litTonemapPbrNeutral(c [3]float64) [3]float64 {
	const startCompression = 0.8 - 0.04
	const desaturation = 0.15
	x := math.Min(c[0], math.Min(c[1], c[2]))
	offset := 0.04
	if x < 0.08 {
		offset = x - 6.25*x*x
	}
	color := [3]float64{c[0] - offset, c[1] - offset, c[2] - offset}
	peak := math.Max(color[0], math.Max(color[1], color[2]))
	if peak < startCompression {
		return color
	}
	d := 1.0 - startCompression
	newPeak := 1.0 - d*d/(peak+d-startCompression)
	color = scale3(color, newPeak/peak)
	g := 1.0 - 1.0/(desaturation*(peak-newPeak)+1.0)
	return mix3(color, [3]float64{newPeak, newPeak, newPeak}, g)
}

// rayBoxAt intersects the ray with the cube of half-size half centered at c —
// the scene box, which bounds the march.
func rayBoxAt(origin, dir, c [3]float64, half float64) (enter, exit float64, hit bool) {
	enter, exit = math.Inf(-1), math.Inf(1)
	for k := 0; k < 3; k++ {
		lo, hi := c[k]-half, c[k]+half
		if math.Abs(dir[k]) < 1e-12 {
			if origin[k] < lo || origin[k] > hi {
				return 0, 0, false
			}
			continue
		}
		t0 := (lo - origin[k]) / dir[k]
		t1 := (hi - origin[k]) / dir[k]
		if t0 > t1 {
			t0, t1 = t1, t0
		}
		enter = math.Max(enter, t0)
		exit = math.Min(exit, t1)
	}
	return enter, exit, exit > math.Max(enter, 0)
}

// capIntersect returns the ray parameter where the ray first meets the
// capsule from a to b with radius r, or -1 on a miss (Quilez's construction).
func capIntersect(ro, rd, a, b [3]float64, r float64) float64 {
	ba := sub3(b, a)
	oa := sub3(ro, a)
	baba := dot3(ba, ba)
	bard := dot3(ba, rd)
	baoa := dot3(ba, oa)
	rdoa := dot3(rd, oa)
	oaoa := dot3(oa, oa)
	qa := baba - bard*bard
	if math.Abs(qa) < 1e-12 {
		qa = 1e-12
	}
	qb := baba*rdoa - baoa*bard
	qc := baba*oaoa - baoa*baoa - r*r*baba
	h := qb*qb - qa*qc
	if h < 0 {
		return -1
	}
	t := (-qb - math.Sqrt(h)) / qa
	y := baoa + t*bard
	if y > 0 && y < baba {
		return t
	}
	// End caps: a sphere at whichever end the body solution missed past.
	oc := oa
	if y > 0 {
		oc = sub3(ro, b)
	}
	sb := dot3(rd, oc)
	sc := dot3(oc, oc) - r*r
	h2 := sb*sb - sc
	if h2 > 0 {
		return -sb - math.Sqrt(h2)
	}
	return -1
}

func gatherPoints(pts [][3]float64, idx []int) (xs, ys, zs []float32) {
	xs = make([]float32, len(idx))
	ys = make([]float32, len(idx))
	zs = make([]float32, len(idx))
	for j, i := range idx {
		xs[j] = float32(pts[i][0])
		ys[j] = float32(pts[i][1])
		zs[j] = float32(pts[i][2])
	}
	return xs, ys, zs
}

func encodePNG(img *image.RGBA) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func toByte(v float64) uint8 {
	return uint8(clamp(v, 0, 1)*255 + 0.5)
}

func clamp(v, lo, hi float64) float64 {
	return math.Min(math.Max(v, lo), hi)
}

func dot3(a, b [3]float64) float64 { return a[0]*b[0] + a[1]*b[1] + a[2]*b[2] }

func add3(a, b [3]float64) [3]float64 {
	return [3]float64{a[0] + b[0], a[1] + b[1], a[2] + b[2]}
}

func sub3(a, b [3]float64) [3]float64 {
	return [3]float64{a[0] - b[0], a[1] - b[1], a[2] - b[2]}
}

// mul3 is the componentwise product, WGSL's a * b on two vectors.
func mul3(a, b [3]float64) [3]float64 {
	return [3]float64{a[0] * b[0], a[1] * b[1], a[2] * b[2]}
}

func scale3(v [3]float64, s float64) [3]float64 {
	return [3]float64{v[0] * s, v[1] * s, v[2] * s}
}

// mix3 is WGSL's mix: a at t = 0, b at t = 1.
func mix3(a, b [3]float64, t float64) [3]float64 {
	return [3]float64{
		a[0] + (b[0]-a[0])*t,
		a[1] + (b[1]-a[1])*t,
		a[2] + (b[2]-a[2])*t,
	}
}

// litReflect is WGSL's reflect: the incident direction mirrored in the plane of
// the normal.
func litReflect(i, n [3]float64) [3]float64 {
	return sub3(i, scale3(n, 2.0*dot3(n, i)))
}

func cross3(a, b [3]float64) [3]float64 {
	return [3]float64{
		a[1]*b[2] - a[2]*b[1],
		a[2]*b[0] - a[0]*b[2],
		a[0]*b[1] - a[1]*b[0],
	}
}

func length3(v [3]float64) float64 { return math.Sqrt(dot3(v, v)) }

func normalize3(v [3]float64) [3]float64 {
	l := length3(v)
	if l == 0 {
		return v
	}
	return [3]float64{v[0] / l, v[1] / l, v[2] / l}
}
