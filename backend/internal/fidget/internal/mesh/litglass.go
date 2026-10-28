// Transparency and refraction for the CPU renderer, by marching the refracted
// ray through the same fields: a port of common/web/gsl-glass.js, which the
// viewers share. That module is the original and this is the copy.
//
// A hit whose material transmits (the volume.x row) is shaded in two parts,
// exactly as gsl_shadeGlass does it. The surface itself contributes its
// specular lobes, its residual (1 - T) diffuse and its emission
// (shadeLitLinear). The transmitted part is found by refracting the view ray
// into the solid and marching it until it meets a surface: the glass's own exit,
// where the path is tinted by Beer-Lambert absorption over the thickness just
// crossed and refracted out (or reflected back in under total internal
// reflection), or a part embedded in the glass, which is shaded and, if it too
// transmits, entered. A ray that escapes shows the flat background. Everything
// is summed in linear light and tone mapped once, and a rough interface jitters
// its microfacet normal over several samples.
//
// Two things differ from the shader, both forced by evaluating in bulk rather
// than per pixel, and neither changes what is computed:
//
//   - The rays march as one wavefront. Every live (pixel, sample) ray takes its
//     step in the same bulk evaluation, bounce by bounce, so the cost is one
//     JIT pass per step over all the glass in the image rather than one per
//     ray, the way the primary march already works.
//   - The secondary march steps on the unsigned field min over shapes of |d|,
//     built for it. The primary march's union is a plain min, and inside glass
//     that is the glass's own negative distance, which says nothing about a
//     part embedded in it: a step of |d_glass| can jump straight past the part.
//     The viewer avoids this with its nearest-crossing composite, which picks
//     the shape of least |d|; the unsigned field is the same thing as a field.
//     Stepping on it cannot cross a surface, so the shader's sign-change
//     bisection has nothing to catch and is not ported.
package mesh

import (
	"fmt"
	"math"

	"gradientcontrol.com/omega/core/backend/internal/fidget/internal/convert"
	"gradientcontrol.com/omega/core/backend/internal/fidget/internal/fidget"
)

// The glass knobs, from SHADING_DEFAULTS in common/web/gsl-shading.js
// (glassSamples, glassCone, glassBounces, glassSteps). In the viewer they ride
// in a uniform; here they are the settled values.
const (
	// litGlassSamples is how many jittered refracted rays a rough interface
	// takes per pixel; a smooth one always takes 1.
	litGlassSamples = 4
	// litGlassCone scales the roughness that jitters the microfacet normal
	// (0 = no blur, 1 = the GGX lobe).
	litGlassCone = 1.0
	// litGlassBounces is how many interfaces a refracted ray may cross before
	// it is treated as escaped. One glass part costs 2 (its exit, then the
	// surface behind it) and glass seen through glass costs 4.
	litGlassBounces = 4
	// litGlassSteps is the march budget per secondary ray.
	litGlassSteps = 128
)

// The march's epsilons, from gsl_marchSecondary and gsl_shadeGlass.
const (
	litGlassHitEps  = 0.001 // floor under the pixel-footprint hit epsilon
	litGlassStepEps = 0.002 // floor under the start offset along the ray
	litGlassFudge   = 0.8   // step shortening, as in the primary march
)

// litGlassScene is what the glass path needs of the scene: the fields to march
// and to look surfaces up in, the scene box that bounds a secondary ray, the
// camera footprint that sizes its epsilons, and the shading context.
type litGlassScene struct {
	shapes           []convert.ColorShape
	renderTrees      []fidget.TreeIndex // each shape's field as rendered, displaced where it moves
	shapeCanvases    []fidget.CanvasIndex
	colorCanvases    [][3]fidget.CanvasIndex
	materialCanvases [][4][4]fidget.CanvasIndex
	lookAt           [3]float64
	boxHalf          float64
	pixelScale       float64
	camPos           [3]float64
	toView           func([3]float64) [3]float64
	// pixelSize() and each pixel's footprint at its first hit: a surface seen
	// through glass keeps the footprint of the glass it is seen through, as
	// in the viewer, where gsl_pw is set once at the front hit.
	pixelSize litPixelSize
	footprint []float64
	ground    litGround
	palette   [][3]float64 // colour of a return that authors none
	// glassSamples overrides litGlassSamples when positive (LitOptions).
	glassSamples int
}

// litGlassRay is one refracted path: a (pixel, sample) pair and the state
// gsl_shadeGlass carries through its bounce loop.
type litGlassRay struct {
	pixel      int
	sample     int
	origin     [3]float64
	dir        [3]float64
	throughput [3]float64
	camDist    float64 // camera distance of the origin, so the epsilons continue from the primary ray's
	ior        float64 // index of the medium the ray is in (1 = air)
	alive      bool
}

// litShadeGlass is gsl_shadeGlass for every transmissive hit in the image at
// once. It returns, for each such pixel, the tone-mapped colour with the
// escaped background added, in linear light; the caller encodes it. Pixels
// whose material does not transmit are not in the map and are shaded opaque.
func litShadeGlass(
	sc litGlassScene,
	hitIdx []int,
	w, h int,
	hitPos, hitNormal, baseColor [][3]float64,
	matAt []litMaterial,
	aoFactor []float64,
) (map[int][3]float64, error) {
	var glassPix []int
	for _, i := range hitIdx {
		if matAt[i][litMatVolume][0] > 0 {
			glassPix = append(glassPix, i)
		}
	}
	if len(glassPix) == 0 {
		// No transmission anywhere: build nothing and evaluate nothing, so an
		// opaque render is exactly what it was without this path.
		return nil, nil
	}

	// The field the secondary march steps on: the distance to the nearest
	// surface of any shape, whichever side of it the ray is on. Built from the
	// rendered fields, so refraction bends on displaced glass as it does in
	// the viewer.
	absUnion := fidget.Abs(sc.renderTrees[0])
	for k := 1; k < len(sc.renderTrees); k++ {
		absUnion = fidget.Min(absUnion, fidget.Abs(sc.renderTrees[k]))
	}
	absCanvas := fidget.BuildCanvas(absUnion)

	local := make(map[int][3]float64, len(glassPix))
	nSamples := make(map[int]int, len(glassPix))
	seeds := make(map[int][2]float64, len(glassPix))
	var rays []*litGlassRay
	for _, i := range glassPix {
		mat := matAt[i]
		n := hitNormal[i]
		v := normalize3(sub3(sc.camPos, hitPos[i]))

		// What the surface itself contributes: specular, residual diffuse,
		// emission.
		local[i] = shadeLitLinear(baseColor[i], n, v, mat, aoFactor[i], sc.toView, sc.ground)

		t0 := clamp(mat[litMatVolume][0], 0, 1)
		rough0 := clamp(mat[litMatSurface][0], 0, 1) * litGlassCone
		ior0 := math.Max(mat[litMatSurface][2], 1)
		noV0 := math.Max(dot3(n, v), 1e-4)
		enter0 := scale3(litTransmitWeight(noV0, mat, baseColor[i]), t0)
		camDist0 := length3(sub3(hitPos[i], sc.camPos))
		// The shader seeds its jitter with the fragment's uv.
		seed := [2]float64{
			(float64(i%w) + 0.5) / float64(w) * 1024.0,
			(float64(i/w) + 0.5) / float64(h) * 1024.0,
		}
		seeds[i] = seed

		// A smooth interface needs one ray; a rough one dithers over several.
		ns := 1
		if rough0 > 0.02 {
			ns = litGlassSamples
			if sc.glassSamples > 0 {
				ns = sc.glassSamples
			}
		}
		nSamples[i] = ns

		for s := 0; s < ns; s++ {
			// Refract into the solid about a jittered microfacet normal; a
			// facet the view ray cannot see falls back to the true normal.
			n0 := litGGXNormal(n, rough0, litHash2(seed, float64(s)))
			if dot3(v, n0) <= 0 {
				n0 = n
			}
			dir := litRefract(scale3(v, -1), n0, 1.0/ior0)
			if dot3(dir, dir) < 1e-6 {
				dir = litReflect(scale3(v, -1), n0)
			}
			// Start off the surface along the normal, not just along the ray:
			// the march declares a hit within an epsilon of a surface, and an
			// along-ray offset shrinks to nothing at grazing angles, so the ray
			// would re-find the interface it just left and burn a bounce on it.
			rays = append(rays, &litGlassRay{
				pixel:      i,
				sample:     s,
				origin:     sub3(hitPos[i], scale3(n, litSurfaceOffset(camDist0, sc.pixelScale))),
				dir:        dir,
				throughput: enter0,
				camDist:    camDist0,
				ior:        ior0,
				alive:      true,
			})
		}
	}

	radiance := make(map[int][3]float64, len(glassPix)) // scene-referred light arriving through the glass
	bgWeight := make(map[int][3]float64, len(glassPix)) // throughput of rays that escaped to the background

	for b := 0; ; b++ {
		var live []*litGlassRay
		for _, r := range rays {
			if r.alive {
				live = append(live, r)
			}
		}
		if len(live) == 0 {
			break
		}
		if b >= litGlassBounces {
			// Budget exhausted: what is still travelling is treated as escaped.
			for _, r := range live {
				bgWeight[r.pixel] = add3(bgWeight[r.pixel], r.throughput)
				r.alive = false
			}
			break
		}

		hits, err := litMarchSecondary(sc, absCanvas, live)
		if err != nil {
			return nil, err
		}

		// Gather the rays that found a surface, for the bulk lookups.
		var hitRays []*litGlassRay
		var hitPts [][3]float64
		var hitT []float64
		for k, r := range live {
			if !hits[k].hit {
				bgWeight[r.pixel] = add3(bgWeight[r.pixel], r.throughput)
				r.alive = false
				continue
			}
			hitRays = append(hitRays, r)
			hitPts = append(hitPts, hits[k].pos)
			hitT = append(hitT, hits[k].t)
		}
		if len(hitRays) == 0 {
			continue
		}

		shapeOf, normals, err := litNearestShape(sc.shapeCanvases, hitPts, hitRays)
		if err != nil {
			return nil, err
		}
		var hitFootprint []float64
		if sc.footprint != nil {
			hitFootprint = make([]float64, len(hitRays))
			for k, r := range hitRays {
				hitFootprint[k] = sc.footprint[r.pixel]
			}
		}
		colors, mats, err := litAppearance(hitPts, shapeOf, sc.shapes, sc.colorCanvases, sc.materialCanvases, sc.palette, sc.pixelSize, hitFootprint)
		if err != nil {
			return nil, err
		}

		for k, r := range hitRays {
			pos := hitPts[k]
			nRaw := normals[k]
			hColor := colors[k]
			hMat := mats[k]
			r.camDist += hitT[k]
			surfOff := litSurfaceOffset(r.camDist, sc.pixelScale)
			hT := clamp(hMat[litMatVolume][0], 0, 1)
			hRough := clamp(hMat[litMatSurface][0], 0, 1) * litGlassCone
			hIor := math.Max(hMat[litMatSurface][2], 1)
			seed := seeds[r.pixel]
			xi := litHash2([2]float64{seed[0] + float64(b) + 1, seed[1]}, float64(r.sample))

			if dot3(r.dir, nRaw) > 0 {
				// Leaving the solid: tint by the path just crossed, pay the
				// exit Fresnel, refract out or reflect back in.
				r.throughput = mul3(r.throughput, litAbsorb(hColor, hitT[k], hMat[litMatVolume][1]))
				nIn := scale3(nRaw, -1) // faces the arriving ray
				n := litGGXNormal(nIn, hRough, xi)
				if dot3(r.dir, n) >= 0 {
					n = nIn
				}
				noV := math.Max(-dot3(n, r.dir), 1e-4)
				r.throughput = mul3(r.throughput, litTransmitWeight(noV, hMat, hColor))
				out := litRefract(r.dir, n, r.ior)
				if dot3(out, out) < 1e-6 {
					out = litReflect(r.dir, n) // total internal reflection: stay inside
				} else {
					r.ior = 1
				}
				r.dir = out
				// Off the surface on whichever side the outgoing ray leaves,
				// so total internal reflection restarts inside the solid.
				r.origin = litOffsetAlong(pos, nRaw, surfOff, dot3(out, nRaw) > 0)
			} else {
				// Entering a surface: it is lit like any other, and entered
				// when it is itself transmissive.
				hView := scale3(r.dir, -1)
				radiance[r.pixel] = add3(radiance[r.pixel], mul3(r.throughput,
					shadeLitLinear(hColor, nRaw, hView, hMat, 1.0, sc.toView, sc.ground)))
				if hT <= 0 {
					r.alive = false
					continue
				}
				n := litGGXNormal(nRaw, hRough, xi)
				if dot3(hView, n) <= 0 {
					n = nRaw
				}
				noV := math.Max(dot3(n, hView), 1e-4)
				r.throughput = mul3(r.throughput, scale3(litTransmitWeight(noV, hMat, hColor), hT))
				into := litRefract(r.dir, n, r.ior/hIor)
				if dot3(into, into) < 1e-6 {
					into = litReflect(r.dir, n)
				} else {
					r.ior = hIor
				}
				r.dir = into
				r.origin = litOffsetAlong(pos, nRaw, surfOff, dot3(into, nRaw) > 0)
			}
		}
	}

	// The scene is tone mapped; the background is added after, as the shader
	// adds its display-referred bgColor. The viewer's bgColor is the canvas
	// clear colour already decoded to linear by Three, so decode it here too.
	bgLinear := litSrgbToLinear(sc.ground.background)
	out := make(map[int][3]float64, len(glassPix))
	for _, i := range glassPix {
		inv := 1.0 / float64(nSamples[i])
		c := litTonemapPbrNeutral(add3(local[i], scale3(radiance[i], inv)))
		out[i] = add3(c, mul3(scale3(bgWeight[i], inv), bgLinear))
	}
	return out, nil
}

// litSecondaryHit is gsl_secondaryHit: where a secondary ray met a surface.
type litSecondaryHit struct {
	pos [3]float64
	t   float64
	hit bool
}

// litMarchSecondary is gsl_marchSecondary for a batch of rays at once: a sphere
// trace with the gradient step correction, started a little way along each ray
// so it does not re-find the surface it leaves, and stopped at the scene box.
// It steps on the unsigned field (see the package note), so it needs no
// bisection.
func litMarchSecondary(sc litGlassScene, absCanvas fidget.CanvasIndex, rays []*litGlassRay) ([]litSecondaryHit, error) {
	out := make([]litSecondaryHit, len(rays))
	t := make([]float64, len(rays))
	tExit := make([]float64, len(rays))
	active := make([]bool, len(rays))
	for k, r := range rays {
		pw := sc.pixelScale * math.Max(r.camDist, 1.0)
		t[k] = math.Max(litGlassStepEps, pw*0.5)
		_, exit, hit := rayBoxAt(r.origin, r.dir, sc.lookAt, sc.boxHalf)
		tExit[k] = exit
		// Leave at the scene's bounding box.
		active[k] = hit && exit > t[k]
	}

	idx := make([]int, 0, len(rays))
	for step := 0; step < litGlassSteps; step++ {
		idx = idx[:0]
		for k := range rays {
			if active[k] {
				idx = append(idx, k)
			}
		}
		if len(idx) == 0 {
			break
		}
		xs := make([]float32, len(idx))
		ys := make([]float32, len(idx))
		zs := make([]float32, len(idx))
		for j, k := range idx {
			p := add3(rays[k].origin, scale3(rays[k].dir, t[k]))
			xs[j], ys[j], zs[j] = float32(p[0]), float32(p[1]), float32(p[2])
		}
		samples, err := fidget.EvalGrad(absCanvas, xs, ys, zs)
		if err != nil {
			return nil, fmt.Errorf("glass march eval: %w", err)
		}
		for j, k := range idx {
			s := samples[j]
			d := float64(s.Value)
			if math.IsNaN(d) || math.IsInf(d, 0) {
				// As in the primary march: a non-finite sample can never
				// resolve, so drop the ray rather than burn its budget.
				active[k] = false
				continue
			}
			pixelWorld := sc.pixelScale * math.Max(rays[k].camDist+t[k], 1.0)
			hitEps := math.Max(litGlassHitEps, pixelWorld*0.5)
			if d < hitEps {
				out[k] = litSecondaryHit{
					pos: [3]float64{float64(xs[j]), float64(ys[j]), float64(zs[j])},
					t:   t[k],
					hit: true,
				}
				active[k] = false
				continue
			}
			gradMag := length3([3]float64{float64(s.Dx), float64(s.Dy), float64(s.Dz)})
			if math.IsNaN(gradMag) || math.IsInf(gradMag, 0) || gradMag < 1 {
				gradMag = 1
			}
			t[k] += d / gradMag * litGlassFudge
			if t[k] > tExit[k] {
				active[k] = false
			}
		}
	}
	return out, nil
}

// litNearestShape returns, for each secondary hit, the shape whose surface it
// is on (the least |d|, the viewer's nearest-crossing composite) and that
// shape's outward normal there, the normalized gradient of its own field. A
// non-finite gradient falls back to facing the arriving ray, as the primary
// march's normal does.
func litNearestShape(shapeCanvases []fidget.CanvasIndex, pts [][3]float64, rays []*litGlassRay) ([]int, [][3]float64, error) {
	shapeOf := make([]int, len(pts))
	normals := make([][3]float64, len(pts))
	best := make([]float64, len(pts))
	for j := range best {
		best[j] = math.Inf(1)
	}
	xs, ys, zs := make([]float32, len(pts)), make([]float32, len(pts)), make([]float32, len(pts))
	for j, p := range pts {
		xs[j], ys[j], zs[j] = float32(p[0]), float32(p[1]), float32(p[2])
	}
	for si, canvas := range shapeCanvases {
		samples, err := fidget.EvalGrad(canvas, xs, ys, zs)
		if err != nil {
			return nil, nil, fmt.Errorf("glass shape eval: %w", err)
		}
		for j, s := range samples {
			if v := math.Abs(float64(s.Value)); v < best[j] {
				best[j] = v
				shapeOf[j] = si
				normals[j] = [3]float64{float64(s.Dx), float64(s.Dy), float64(s.Dz)}
			}
		}
	}
	for j := range normals {
		g := normals[j]
		if l := length3(g); l == 0 || math.IsNaN(l) || math.IsInf(l, 0) {
			normals[j] = scale3(rays[j].dir, -1)
			continue
		}
		normals[j] = normalize3(g)
	}
	return shapeOf, normals, nil
}

// litSurfaceOffset is gsl_surfaceOffset: how far off a surface a secondary ray
// starts, along the normal. Twice the distance within which the march declares
// a hit at this pixel footprint, so the ray cannot re-find the interface it
// just left. Sub-pixel by construction, which is also why a wall thinner than
// this is punched through.
func litSurfaceOffset(camDist, pixelScale float64) float64 {
	pixelWorld := pixelScale * math.Max(camDist, 1.0)
	return 2.0 * math.Max(litGlassHitEps, pixelWorld*0.5)
}

// litOffsetAlong moves p off a surface by off along the normal n, onto the
// side the outgoing ray leaves by: outward when outward is true.
func litOffsetAlong(p, n [3]float64, off float64, outward bool) [3]float64 {
	if outward {
		return add3(p, scale3(n, off))
	}
	return sub3(p, scale3(n, off))
}

// litTransmitWeight is gsl_transmitWeight: the fraction of light a
// transmissive interface passes rather than reflects, one minus the
// roughness-aware Schlick Fresnel with F0 from the index of refraction.
// Symmetric, so the same term serves the exit interface. Total internal
// reflection is handled geometrically by the caller.
func litTransmitWeight(noV float64, mat litMaterial, baseColor [3]float64) [3]float64 {
	albedo := litSrgbToLinear(baseColor)
	roughness := clamp(mat[litMatSurface][0], 0.045, 1.0)
	metallic := clamp(mat[litMatSurface][1], 0.0, 1.0)
	ior := math.Max(mat[litMatSurface][2], 1.0)
	f0Dielectric := ((ior - 1.0) / (ior + 1.0)) * ((ior - 1.0) / (ior + 1.0))
	f0 := mix3([3]float64{f0Dielectric, f0Dielectric, f0Dielectric}, albedo, metallic)
	return sub3([3]float64{1, 1, 1}, litFresnelSchlickRoughness(noV, f0, roughness))
}

// litAbsorb is gsl_absorb: Beer-Lambert absorption over a path through a
// tinted medium, in the glTF KHR_materials_volume form. At absorptionDistance
// the transmitted colour is the authored colour, and it darkens as a power of
// the path beyond that. absorptionDistance <= 0 means no absorption.
func litAbsorb(baseColor [3]float64, thickness, absorptionDistance float64) [3]float64 {
	if absorptionDistance <= 0 {
		return [3]float64{1, 1, 1}
	}
	tint := litSrgbToLinear(baseColor)
	e := math.Max(thickness, 0) / absorptionDistance
	var out [3]float64
	for k := 0; k < 3; k++ {
		out[k] = math.Pow(math.Max(tint[k], 1e-4), e)
	}
	return out
}

// litGGXNormal is gsl_ggxNormal: a microfacet normal drawn from the GGX
// distribution around n (alpha = roughness squared) for the uniform pair xi, so
// a rough interface refracts and reflects into a lobe instead of a line.
// Roughness 0 returns n.
func litGGXNormal(n [3]float64, roughness float64, xi [2]float64) [3]float64 {
	a := roughness * roughness
	if a < 1e-4 {
		return n
	}
	phi := 2 * math.Pi * xi[0]
	cosTheta := math.Sqrt((1.0 - xi[1]) / (1.0 + (a*a-1.0)*xi[1]))
	sinTheta := math.Sqrt(math.Max(0, 1.0-cosTheta*cosTheta))
	up := [3]float64{1, 0, 0}
	if math.Abs(n[2]) < 0.999 {
		up = [3]float64{0, 0, 1}
	}
	tx := normalize3(cross3(up, n))
	ty := cross3(n, tx)
	return normalize3(add3(add3(
		scale3(tx, sinTheta*math.Cos(phi)),
		scale3(ty, sinTheta*math.Sin(phi))),
		scale3(n, cosTheta)))
}

// litHash2 is gsl_hash2: two uniform numbers in [0, 1) from a 2-D seed and a
// sample index (Hoskins' hash without sine), for the per-pixel jitter. The
// shader runs it in f32 and this in f64, so the two draw different numbers
// from the same seed; only their distribution matters.
func litHash2(p [2]float64, s float64) [2]float64 {
	h := [3]float64{
		litFract(p[0] * 0.1031),
		litFract(p[1] * 0.1030),
		litFract(s * 0.0973),
	}
	d := h[0]*(h[1]+33.33) + h[1]*(h[2]+33.33) + h[2]*(h[0]+33.33)
	h = [3]float64{h[0] + d, h[1] + d, h[2] + d}
	return [2]float64{
		litFract((h[0] + h[1]) * h[2]),
		litFract((h[0] + h[2]) * h[1]),
	}
}

func litFract(x float64) float64 { return x - math.Floor(x) }

// litRefract is WGSL's refract: the incident direction i bent through the
// surface with normal n at the ratio of indices eta, or the zero vector under
// total internal reflection.
func litRefract(i, n [3]float64, eta float64) [3]float64 {
	cosI := dot3(n, i)
	k := 1.0 - eta*eta*(1.0-cosI*cosI)
	if k < 0 {
		return [3]float64{}
	}
	return sub3(scale3(i, eta), scale3(n, eta*cosI+math.Sqrt(k)))
}
