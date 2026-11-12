package threemf

import (
	"fmt"
	"math"
)

// PaletteColor is body index's (1-based) default color as "#rrggbb": the
// viewer's palette (examples/viewer/panels/body-palette.js, its first preset),
// so an export colors a body as the viewer draws it when the program gives it
// no color. The palette is a cosine palette a + b cos 2π(c t + d), sampled at
// a golden-ish step per body and sRGB-encoded as the viewer encodes it.
func PaletteColor(index int) string {
	if index < 1 {
		index = 1
	}
	a := [3]float64{0.5, 0.5, 0.5}
	b := [3]float64{0.5, 0.5, 0.5}
	c := [3]float64{1, 1, 1}
	d := [3]float64{0.00, 0.33, 0.67}
	t := math.Mod(1.0/3+float64(index-1)*(5.0/12), 1)
	var rgb [3]int
	for i := range rgb {
		v := a[i] + b[i]*math.Cos(2*math.Pi*(c[i]*t+d[i]))
		v = math.Max(0, math.Min(1, v))
		rgb[i] = int(math.Round(srgbEncode(v) * 255))
	}
	return fmt.Sprintf("#%02x%02x%02x", rgb[0], rgb[1], rgb[2])
}

// srgbEncode is the viewer's linear-to-sRGB encoding (render/gsl-shading.js).
func srgbEncode(v float64) float64 {
	if v <= 0.0031308 {
		return 12.92 * v
	}
	return 1.055*math.Pow(v, 1/2.4) - 0.055
}
