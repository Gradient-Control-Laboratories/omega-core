package main

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"

	"gradientcontrol.com/omega/core/eval"
	"gradientcontrol.com/omega/core/internal/source"
)

// The colours a material paints, for the Bodies panel's colour chip: a body
// given a material takes its colour from the material's pair, which the panel
// cannot evaluate, so the manifest carries it. They are read off the colour
// field itself, unlit, over the same unit sphere the swatch draws — the
// albedo, not the shaded swatch, since the chip stands for the colour an
// author would pick.
//
// A material whose colour is the same everywhere records that one colour. One
// whose colour varies records a few that span it, darkest to lightest, and the
// chip draws them as a gradient. Every material the libraries annotate today is
// uniform; the spread is checked rather than assumed, so a mottled one added
// later is shown as mottled rather than as whatever its first sample was.

const (
	colorSamples = 64
	// uniformSpread is the largest channel range, over the samples, still read
	// as one colour: under one 8-bit step either side.
	uniformSpread = 1.0 / 255
)

// colorEvaluator is shared by every entry: building one registers process-wide
// state (the noise tile), and queries serialise on it anyway. It is built on
// first use, so a -list run never pays for it.
var colorEvaluator = sync.OnceValue(func() *eval.Evaluator {
	return eval.New(&source.MockManager{}, nil)
})

// surfaceColors evaluates a material's colour over the swatch sphere. A
// finish carries no colour, so it has none.
func surfaceColors(e Entry) ([][3]float64, error) {
	if e.Kind != "material" {
		return nil, nil
	}
	doc := fmt.Sprintf("swatchMat, swatchColor = %s;\nball = sphere(%g);\nreturn ball;\n", e.Call(), swatchRadius)
	groups, err := colorEvaluator().EvalGsl(context.Background(), nil, doc, "", []string{"swatchColor"}, spherePoints(colorSamples, swatchRadius))
	if err != nil {
		return nil, fmt.Errorf("colour: %w", err)
	}
	if len(groups) != 1 || len(groups[0].Samples) != 3 {
		return nil, fmt.Errorf("colour is not a vec3")
	}
	comps := groups[0].Samples
	samples := make([][3]float64, len(comps[0]))
	for i := range samples {
		for c := 0; c < 3; c++ {
			samples[i][c] = float64(comps[c][i].Value)
		}
	}
	return spanColors(samples), nil
}

// spanColors reduces the samples to the colours the chip shows: their mean
// when they are one colour, else the 10th, 50th and 90th percentile by
// luminance, which span the range without its outliers.
func spanColors(samples [][3]float64) [][3]float64 {
	spread := 0.0
	for c := 0; c < 3; c++ {
		lo, hi := math.Inf(1), math.Inf(-1)
		for _, s := range samples {
			lo, hi = math.Min(lo, s[c]), math.Max(hi, s[c])
		}
		spread = math.Max(spread, hi-lo)
	}
	if spread <= uniformSpread {
		var mean [3]float64
		for _, s := range samples {
			for c := 0; c < 3; c++ {
				mean[c] += s[c] / float64(len(samples))
			}
		}
		return [][3]float64{roundColor(mean)}
	}
	sorted := append([][3]float64(nil), samples...)
	luma := func(s [3]float64) float64 { return 0.2126*s[0] + 0.7152*s[1] + 0.0722*s[2] }
	sort.SliceStable(sorted, func(i, j int) bool { return luma(sorted[i]) < luma(sorted[j]) })
	var out [][3]float64
	for _, q := range []float64{0.1, 0.5, 0.9} {
		out = append(out, roundColor(sorted[int(q*float64(len(sorted)-1)+0.5)]))
	}
	return out
}

// roundColor keeps three decimals, which round-trip every 8-bit channel and
// keep the manifest readable.
func roundColor(c [3]float64) [3]float64 {
	for i := range c {
		c[i] = math.Round(math.Max(0, math.Min(1, c[i]))*1000) / 1000
	}
	return c
}

// spherePoints spreads n points evenly over a sphere (a Fibonacci lattice).
func spherePoints(n int, radius float64) [][3]float64 {
	pts := make([][3]float64, n)
	golden := math.Pi * (3 - math.Sqrt(5))
	for i := range pts {
		z := 1 - 2*(float64(i)+0.5)/float64(n)
		r := math.Sqrt(1 - z*z)
		t := float64(i) * golden
		pts[i] = [3]float64{radius * r * math.Cos(t), radius * r * math.Sin(t), radius * z}
	}
	return pts
}
