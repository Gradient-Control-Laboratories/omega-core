//go:build cgo

// Environment light presets for litrender: the Go copy of the analytic
// environments in common/web/gsl-environment.js (GSL_ENVIRONMENT_WGSL), which
// the viewers' Light row chooses between. That module is the original and this
// is the copy; the presets themselves are shared, one JSON file both read
// (common/web/environment).
//
// An analytic environment is a step sky (one colour above the horizon, one
// below) plus up to four spherical-Gaussian panels, and a strength for the
// viewer's rig lights. It replaces shadeLitLinear's ambient hemisphere with
// closed forms: the step sky's exact diffuse, its specular blurred by roughness
// with a fit, the panels' fitted irradiance and their specular as a warped SG
// product. The Viewer preset is not analytic: it is the viewer's own lighting,
// litDark or litLight, shaded as before.
//
// The environment turns with the camera about the vertical, as it does in the
// viewers (followCamera), and like them shows only in reflections: the
// background stays flat, and a ray that leaves through glass still shows it.
package eval

import (
	"math"

	"gradientcontrol.com/omega/core/internal/fidget/environment"
)

// litEnv is an analytic environment, turned for one camera.
type litEnv struct {
	rig         float64    // the rig lights' strength
	sky, ground [3]float64 // the step sky, linear radiance
	panels      []litPanel
}

// litPanel is one spherical-Gaussian panel: a unit world axis, a sharpness and
// a linear RGB amplitude.
type litPanel struct {
	axis   [3]float64
	lambda float64
	color  [3]float64
}

// litEnvironment resolves the named preset for a camera whose right vector
// (world, Z up) is right, turned turn degrees further, as the viewers' Turn
// slider does. The empty name and the Viewer preset return nil: the viewer's
// own lighting.
func litEnvironment(name string, turn float64, right [3]float64) (*litEnv, error) {
	if name == "" {
		return nil, nil
	}
	p, err := environment.Lookup(name)
	if err != nil {
		return nil, err
	}
	if p.Mode == 0 {
		return nil, nil
	}
	// followCamera: the presets are laid out for the viewer's starting view,
	// camera azimuth 45 degrees, whose right vector heads 135 degrees. The
	// viewer keeps its last turn when looking straight down; a render has no
	// last turn, so it takes none.
	camTurn := 0.0
	if math.Hypot(right[0], right[1]) >= 1e-6 {
		camTurn = math.Atan2(right[1], right[0])*180/math.Pi - 90 - 45
	}
	th := (turn + camTurn) * math.Pi / 180
	cs, sn := math.Cos(th), math.Sin(th)
	env := &litEnv{rig: p.Rig, sky: p.Sky, ground: p.Ground}
	for _, pn := range p.Panels {
		l := length3(pn.Axis)
		if l == 0 {
			l = 1
		}
		x, y, z := pn.Axis[0]/l, pn.Axis[1]/l, pn.Axis[2]/l
		env.panels = append(env.panels, litPanel{
			axis:   [3]float64{cs*x - sn*y, sn*x + cs*y, z},
			lambda: pn.Sharpness,
			color:  pn.Color,
		})
	}
	return env, nil
}

// radiance is gsl_envRadiance: the light arriving from direction d.
func (e *litEnv) radiance(d [3]float64) [3]float64 {
	L := e.ground
	if d[2] > 0 {
		L = e.sky
	}
	for _, p := range e.panels {
		L = add3(L, scale3(p.color, math.Exp(p.lambda*(dot3(p.axis, d)-1))))
	}
	return L
}

// diffuse is gsl_envDiffuse: the light a Lambertian surface with normal n
// reflects per unit albedo, irradiance over pi, from the sky and the panels.
func (e *litEnv) diffuse(n [3]float64) [3]float64 {
	E := mix3(e.ground, e.sky, n[2]*0.5+0.5)
	for _, p := range e.panels {
		E = add3(E, scale3(p.color, litSGIrradiance(dot3(p.axis, n), p.lambda)/math.Pi))
	}
	return E
}

// skySpecular is gsl_envSkySpecular: the step sky seen along r through a GGX
// lobe of the given roughness, before the environment BRDF. The fraction of
// the lobe above the horizon is fitted as 0.5 + 0.5 x / (|x| + r^2 (1 - |x|)),
// x = r.z, which is the old linear ramp at roughness 1.
func (e *litEnv) skySpecular(r [3]float64, roughness float64) [3]float64 {
	x := clamp(r[2], -1, 1)
	c := roughness * roughness
	t := 0.5 + 0.5*x/math.Max(math.Abs(x)+c*(1-math.Abs(x)), 1e-6)
	return mix3(e.ground, e.sky, t)
}

// panelSpecular is gsl_envPanelSpecular: each panel's specular reflection. The
// GGX lobe (alpha = roughness squared) as an SG around n is warped to the
// reflection direction (Wang et al. 2009), convolved with the panel, and shaded
// at the warped axis; Fresnel, visibility and the cosine are inside.
func (e *litEnv) panelSpecular(n, v, f0 [3]float64, roughness float64) [3]float64 {
	a := roughness * roughness
	a2 := math.Max(a*a, 1e-7)
	noV := math.Max(dot3(n, v), 1e-4)
	r := litReflect(scale3(v, -1), n)
	lambdaW := (2.0 / a2) / (4.0 * math.Max(noV, 0.1))
	ampW := 1.0 / (math.Pi * a2)
	noL := math.Max(dot3(n, r), 0)
	h := normalize3(add3(r, v))
	f := litFresnelSchlick(math.Max(dot3(r, h), 0), f0)
	shade := scale3(f, litVisibilitySmithGGX(noV, noL, a2)*noL)
	var s [3]float64
	for _, p := range e.panels {
		s = add3(s, scale3(p.color, litSGInner(r, lambdaW, p.axis, p.lambda)))
	}
	return mul3(scale3(s, ampW), shade)
}

// litSGInner is gsl_sgInner: the integral over the sphere of the product of
// two unit-amplitude spherical Gaussians.
func litSGInner(axis1 [3]float64, lambda1 float64, axis2 [3]float64, lambda2 float64) float64 {
	um := length3(add3(scale3(axis1, lambda1), scale3(axis2, lambda2)))
	expo := math.Exp(um - lambda1 - lambda2)
	return 2 * math.Pi * expo * (1 - math.Exp(-2*um)) / math.Max(um, 1e-6)
}

// litSGIrradiance is gsl_sgIrradiance: the irradiance a unit-amplitude SG of
// sharpness lambda delivers to a surface whose normal makes cosine muDotN with
// its axis (Hill's fit, as given by Pettineo, "SG Series" part 3).
func litSGIrradiance(muDotN, lambda float64) float64 {
	l := math.Max(lambda, 0.1)
	const c0 = 0.36
	const c1 = 1.0 / (4.0 * 0.36)
	eml := math.Exp(-l)
	em2l := eml * eml
	rl := 1 / l
	scale := 1 + 2*em2l - rl
	bias := (eml-em2l)*rl - em2l
	x := math.Sqrt(math.Max(1-scale, 1e-8))
	x0 := c0 * muDotN
	x1 := c1 * x
	n := x0 + x1
	y := clamp(muDotN, 0, 1)
	if math.Abs(x0) <= x1 {
		y = n * n / x
	}
	integral := 2 * math.Pi * (1 - em2l) * rl
	return (scale*y + bias) * integral
}
