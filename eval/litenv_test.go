//go:build cgo

package eval

import (
	"bytes"
	"math"
	"math/rand/v2"
	"testing"

	"gradientcontrol.com/omega/core/internal/fidget/environment"
)

// litEnvSphere draws a uniform direction on the sphere.
func litEnvSphere(rng *rand.Rand) [3]float64 {
	z := 2*rng.Float64() - 1
	phi := 2 * math.Pi * rng.Float64()
	s := math.Sqrt(1 - z*z)
	return [3]float64{s * math.Cos(phi), s * math.Sin(phi), z}
}

// litEnvSampleSG draws a direction from a spherical Gaussian of axis mu and
// sharpness lambda, the path tracer's gsl_envSamplePanel for one panel, and
// returns it with its density per steradian.
func litEnvSampleSG(rng *rand.Rand, mu [3]float64, lambda float64) ([3]float64, float64) {
	cosT := 1 + math.Log(1-rng.Float64()*(1-math.Exp(-2*lambda)))/lambda
	cosT = clamp(cosT, -1, 1)
	sinT := math.Sqrt(math.Max(0, 1-cosT*cosT))
	phi := 2 * math.Pi * rng.Float64()
	up := [3]float64{0, 0, 1}
	if math.Abs(mu[2]) >= 0.999 {
		up = [3]float64{1, 0, 0}
	}
	tx := normalize3(cross3(up, mu))
	ty := cross3(mu, tx)
	d := normalize3(add3(add3(scale3(tx, sinT*math.Cos(phi)), scale3(ty, sinT*math.Sin(phi))), scale3(mu, cosT)))
	pdf := lambda * math.Exp(lambda*(dot3(mu, d)-1)) / (2 * math.Pi * (1 - math.Exp(-2*lambda)))
	return d, pdf
}

// The step sky's diffuse, mix(ground, sky, 0.5 + 0.5 n.z), is exact: the
// cosine-weighted integral of the step sky over the hemisphere, over pi.
func TestLitEnvSkyDiffuseIsExact(t *testing.T) {
	env := &litEnv{sky: [3]float64{1, 0.5, 0.25}, ground: [3]float64{0.1, 0.2, 0.4}}
	rng := rand.New(rand.NewPCG(1, 2))
	const m = 400000
	for _, nz := range []float64{1, 0.6, 0, -0.5} {
		n := [3]float64{math.Sqrt(1 - nz*nz), 0, nz}
		var sum [3]float64
		for i := 0; i < m; i++ {
			d := litEnvSphere(rng)
			if c := dot3(n, d); c > 0 {
				sum = add3(sum, scale3(env.radiance(d), c))
			}
		}
		mc := scale3(sum, 4*math.Pi/m/math.Pi)
		got := env.diffuse(n)
		for k := 0; k < 3; k++ {
			if math.Abs(got[k]-mc[k]) > 0.01 {
				t.Errorf("n.z %v: diffuse %v, Monte Carlo %v", nz, got, mc)
				break
			}
		}
	}
}

// Hill's fitted SG irradiance against the integral it stands for, sampled
// from the SG itself so sharp lobes converge.
func TestLitSGIrradianceFit(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 4))
	const m = 200000
	n := [3]float64{0, 0, 1}
	worst := 0.0
	for _, lambda := range []float64{2, 6, 24, 150} {
		for _, c := range []float64{1, 0.5, 0, -0.3} {
			mu := [3]float64{math.Sqrt(1 - c*c), 0, c}
			sum := 0.0
			for i := 0; i < m; i++ {
				d, pdf := litEnvSampleSG(rng, mu, lambda)
				if cn := dot3(n, d); cn > 0 {
					sum += math.Exp(lambda*(dot3(mu, d)-1)) * cn / pdf
				}
			}
			mc := sum / m
			fit := litSGIrradiance(c, lambda)
			// The fit's error is absolute near the horizon, where the
			// irradiance itself goes to zero; measure it against the peak.
			peak := 2 * math.Pi * (1 - math.Exp(-2*lambda)) / lambda
			err := math.Abs(fit-mc) / peak
			worst = math.Max(worst, err)
			if err > 0.05 {
				t.Errorf("lambda %v, cos %v: fit %.5f, Monte Carlo %.5f (%.1f%% of the peak)", lambda, c, fit, mc, 100*err)
			}
		}
	}
	t.Logf("worst SG irradiance error: %.2f%% of the peak", 100*worst)
}

// The sky's specular fit against the fraction of the split-sum GGX lobe
// (N = V = R, weighted by N.L) that lies above the horizon.
func TestLitEnvSkySpecularFit(t *testing.T) {
	rng := rand.New(rand.NewPCG(5, 6))
	const m = 100000
	env := &litEnv{sky: [3]float64{1, 1, 1}}
	worst := 0.0
	for _, r := range []float64{0.05, 0.1, 0.2, 0.3, 0.5, 0.7, 1.0} {
		a := r * r
		for _, rz := range []float64{-0.9, -0.6, -0.3, -0.1, 0, 0.1, 0.3, 0.6, 0.9} {
			R := [3]float64{math.Sqrt(1 - rz*rz), 0, rz}
			// A frame about R: ex is perpendicular to it in the XZ plane.
			ex := [3]float64{rz, 0, -math.Sqrt(1 - rz*rz)}
			ey := [3]float64{0, 1, 0}
			above, total := 0.0, 0.0
			for i := 0; i < m; i++ {
				u1, u2 := rng.Float64(), rng.Float64()
				ct := math.Sqrt((1 - u2) / (1 + (a*a-1)*u2))
				st := math.Sqrt(1 - ct*ct)
				ph := 2 * math.Pi * u1
				h := add3(add3(scale3(ex, st*math.Cos(ph)), scale3(ey, st*math.Sin(ph))), scale3(R, ct))
				l := sub3(scale3(h, 2*dot3(h, R)), R)
				w := math.Max(dot3(l, R), 0)
				total += w
				if l[2] > 0 {
					above += w
				}
			}
			mc := above / total
			got := env.skySpecular(R, r)[0]
			err := math.Abs(got - mc)
			worst = math.Max(worst, err)
			if err > 0.05 {
				t.Errorf("roughness %v, R.z %v: fit %.3f, Monte Carlo %.3f", r, rz, got, mc)
			}
		}
	}
	t.Logf("worst sky specular error: %.3f", worst)
}

// The panels' specular (the warped SG product) against the BRDF integrated
// over the panel. The approximation reads bright in every case measured here,
// by 2% to 40% (most often 10% to 30%, the most at a grazing view with the
// panel off the mirror direction), in the viewers too, whose WGSL this copies.
// The bounds hold it there, so a change for the worse shows.
func TestLitEnvPanelSpecularFit(t *testing.T) {
	rng := rand.New(rand.NewPCG(7, 8))
	const m = 2000000
	n := [3]float64{0, 0, 1}
	f0 := [3]float64{0.04, 0.04, 0.04}
	worst := 0.0
	for _, r := range []float64{0.3, 0.5, 0.8} {
		a := r * r
		a2 := a * a
		for _, vang := range []float64{20, 60} {
			v := [3]float64{math.Sin(vang * math.Pi / 180), 0, math.Cos(vang * math.Pi / 180)}
			R := litReflect(scale3(v, -1), n)
			noV := dot3(n, v)
			for _, off := range []float64{0, 15} {
				o := off * math.Pi / 180
				mu := normalize3(add3(scale3(R, math.Cos(o)), scale3([3]float64{0, 1, 0}, math.Sin(o))))
				for _, lambda := range []float64{6, 24} {
					env := &litEnv{panels: []litPanel{{axis: mu, lambda: lambda, color: [3]float64{1, 1, 1}}}}
					sum := 0.0
					for i := 0; i < m; i++ {
						l := litEnvSphere(rng)
						noL := dot3(n, l)
						if noL <= 0 {
							continue
						}
						h := normalize3(add3(l, v))
						brdf := litDistributionGGX(dot3(h, n), a2) * litVisibilitySmithGGX(noV, noL, a2) *
							litFresnelSchlick(clamp(dot3(h, v), 0, 1), f0)[0]
						sum += math.Exp(lambda*(dot3(mu, l)-1)) * brdf * noL
					}
					mc := sum * 4 * math.Pi / m
					got := env.panelSpecular(n, v, f0, r)[0]
					ratio := got / mc
					worst = math.Max(worst, math.Abs(ratio-1))
					if ratio < 0.95 || ratio > 1.45 {
						t.Errorf("roughness %v, view %v deg, panel %v deg off R, lambda %v: fit %.5f, Monte Carlo %.5f (x%.2f)",
							r, vang, off, lambda, got, mc, ratio)
					}
				}
			}
		}
	}
	t.Logf("worst panel specular ratio error: %.0f%%", 100*worst)
}

// The presets are laid out for the viewer's starting view; the environment
// turns with the camera's azimuth, plus Turn, and the sky stays up.
func TestLitEnvironmentTurnsWithTheCamera(t *testing.T) {
	p, err := environment.Lookup("Horizon")
	if err != nil {
		t.Fatal(err)
	}
	want := normalize3(p.Panels[0].Axis)
	// The starting view's right vector, from (1, 1, 1) looking in.
	start := normalize3(cross3(normalize3([3]float64{-1, -1, -1}), [3]float64{0, 0, 1}))
	env, err := litEnvironment("Horizon", 0, start)
	if err != nil {
		t.Fatal(err)
	}
	if got := env.panels[0].axis; length3(sub3(got, want)) > 1e-9 {
		t.Errorf("starting view: sun axis %v, preset %v", got, want)
	}
	// Turning the camera 90 degrees about Z turns the sun with it; so does
	// Turn, on its own.
	turned := [3]float64{-start[1], start[0], start[2]}
	quarter := [3]float64{-want[1], want[0], want[2]}
	for _, c := range []struct {
		name  string
		turn  float64
		right [3]float64
	}{{"camera", 0, turned}, {"Turn", 90, start}} {
		env, err := litEnvironment("Horizon", c.turn, c.right)
		if err != nil {
			t.Fatal(err)
		}
		if got := env.panels[0].axis; length3(sub3(got, quarter)) > 1e-9 {
			t.Errorf("%s turned 90 degrees: sun axis %v, want %v", c.name, got, quarter)
		}
	}
	if env, err := litEnvironment("Viewer", 0, start); err != nil || env != nil {
		t.Errorf("Viewer: env %v, err %v; want the viewer's own lighting (nil)", env, err)
	}
	if _, err := litEnvironment("no such preset", 0, start); err == nil {
		t.Error("an unknown preset resolved")
	}
}

// The empty name and Viewer render the viewer's own lighting exactly; each
// analytic preset lights the scene differently, and without NaNs.
func TestRenderLitEnvironmentPresets(t *testing.T) {
	prog := litProgram(t, litBallDoc)
	render := func(name string) ([]byte, *LitLinear) {
		t.Helper()
		opts := LitOptions{Width: 120, Height: 80, WorldExtent: 6, Environment: name}
		var png []byte
		var lin *LitLinear
		if err := renderLit(prog, opts, func(p []byte, l *LitLinear) { png, lin = p, l }); err != nil {
			t.Fatalf("%q: %v", name, err)
		}
		return png, lin
	}
	base, _ := render("")
	if viewer, _ := render("Viewer"); !bytes.Equal(base, viewer) {
		t.Error("the Viewer preset renders differently from no preset")
	}
	all, err := environment.Presets()
	if err != nil {
		t.Fatal(err)
	}
	for name, p := range all {
		if p.Mode == 0 {
			continue
		}
		png, lin := render(name)
		if bytes.Equal(png, base) {
			t.Errorf("%s renders the same as the viewer's own lighting", name)
		}
		for i, v := range lin.RGB {
			if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
				t.Fatalf("%s: pixel %d channel %d is %v", name, i/3, i%3, v)
			}
		}
	}
	if _, err := RenderLit(prog, LitOptions{Width: 40, Height: 30, WorldExtent: 6, Environment: "no such preset"}); err == nil {
		t.Error("rendering with an unknown preset succeeded")
	}
}
