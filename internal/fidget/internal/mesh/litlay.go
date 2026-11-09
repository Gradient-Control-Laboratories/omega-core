package mesh

import "math"

// The CPU twin of the viewer's lay shading (gsl_layFrame and its neighbours in
// common/web/gsl-shading.js): a material's lay, grooves too fine to see
// running along a direction, stretches the specular lobe across them.

// litLayFrame is gsl_layFrame: the unit tangent along the lay across the
// surface, in the space of n, and the anisotropy, 0 where the material has
// none or the lay meets the surface head on.
func litLayFrame(n, lay [3]float64, anisotropy float64) ([3]float64, float64) {
	amount := clamp(anisotropy, 0, 1)
	across := sub3(lay, scale3(n, dot3(lay, n)))
	len2 := dot3(across, across)
	if amount <= 0 || len2 < 1e-10 {
		return [3]float64{1, 0, 0}, 0
	}
	return scale3(across, 1/math.Sqrt(len2)), amount
}

// litLayAlphas is gsl_layAlphas: the lobe's widths along the lay and across it.
func litLayAlphas(roughness, amount float64) (float64, float64) {
	alpha := roughness * roughness
	aspect := math.Sqrt(1 - 0.9*amount)
	return math.Max(alpha*aspect, 0.002), math.Max(alpha/aspect, 0.002)
}

// litDistributionGGXAniso is gsl_distributionGGXAniso.
func litDistributionGGXAniso(noH, toH, boH, aT, aB float64) float64 {
	d := toH*toH/(aT*aT) + boH*boH/(aB*aB) + noH*noH
	return 1 / (math.Pi * aT * aB * d * d)
}

// litVisibilitySmithGGXAniso is gsl_visibilitySmithGGXAniso.
func litVisibilitySmithGGXAniso(noV, noL, toV, boV, toL, boL, aT, aB float64) float64 {
	lambdaV := noL * length3([3]float64{aT * toV, aB * boV, noV})
	lambdaL := noV * length3([3]float64{aT * toL, aB * boL, noL})
	return 0.5 / math.Max(lambdaV+lambdaL, 1e-5)
}

// litDirectLightLay is gsl_directLightLay: litDirectLight where the material
// has no lay, the anisotropic lobe along it where it has one.
func litDirectLightLay(n, v [3]float64, noV float64, l, radiance, diffuseColor, f0 [3]float64, alpha2, roughness float64, t [3]float64, amount float64) [3]float64 {
	if amount <= 0 {
		return litDirectLight(n, v, noV, l, radiance, diffuseColor, f0, alpha2)
	}
	noL := dot3(n, l)
	if noL <= 0 {
		return [3]float64{}
	}
	b := cross3(n, t)
	aT, aB := litLayAlphas(roughness, amount)
	h := normalize3(add3(l, v))
	d := litDistributionGGXAniso(math.Max(dot3(n, h), 0), dot3(t, h), dot3(b, h), aT, aB)
	vis := litVisibilitySmithGGXAniso(noV, noL, dot3(t, v), dot3(b, v), dot3(t, l), dot3(b, l), aT, aB)
	f := litFresnelSchlick(math.Max(dot3(v, h), 0), f0)
	specular := scale3(f, d*vis)
	diffuse := scale3(mul3(sub3([3]float64{1, 1, 1}, f), diffuseColor), 1.0/math.Pi)
	return scale3(mul3(add3(diffuse, specular), radiance), noL)
}

// litLayBentNormal is gsl_layBentNormal: the normal the reflection is taken
// about, bent toward the plane the lobe stretches in.
func litLayBentNormal(n, v, t [3]float64, amount float64) [3]float64 {
	if amount <= 0 {
		return n
	}
	stretched := cross3(cross3(t, v), t)
	return normalize3(add3(scale3(n, 1-amount), scale3(stretched, amount)))
}
