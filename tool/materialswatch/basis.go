package main

import (
	"encoding/binary"
	"fmt"
	"math"

	"gradientcontrol.com/omega/core/assembler"
	fidgetlib "gradientcontrol.com/omega/core/internal/fidget/lib"
)

// The finishes' recolouring basis. A finish has no colour, so the Bodies panel
// draws its swatch in the selected body's colour (examples/viewer/panels/finish-tint.js).
// Before tone mapping a render is linear in the albedo, so two renders of each
// finish — albedo 0 and albedo 1, radiance kept — give its swatch in any
// colour. They are stored as half floats, three per pixel, finish after finish
// in catalog order, the albedo-0 render then the albedo-1 render.

// FinishBasis is the manifest's description of the basis file.
type FinishBasis struct {
	File   string   `json:"file"`   // beside the manifest
	Width  int      `json:"width"`  // pixels, the sheet's cell
	Height int      `json:"height"` //
	Order  []string `json:"order"`  // package.symbol, the finishes in the file's order
}

// basisAlbedos are the two renders' colours.
var basisAlbedos = [2][3]float64{{0, 0, 0}, {1, 1, 1}}

// renderRadiance renders a finish in albedo and returns its PNG and its
// radiance before tone mapping.
func renderRadiance(e Entry, albedo [3]float64, width, height int, extent float64, env string, light, ground bool, groundGrey, checker float64) ([]byte, []float32, error) {
	return renderRadianceDoc(swatchDocIn(e, albedo, ground, groundGrey, checker), width, height, extent, env, light)
}

// renderRadianceDoc renders a swatch document and returns its PNG and its
// radiance before tone mapping.
func renderRadianceDoc(doc string, width, height int, extent float64, env string, light bool) ([]byte, []float32, error) {
	res, err := assembler.Assemble([]string{doc})
	if err != nil {
		return nil, nil, fmt.Errorf("assemble: %w", err)
	}
	return fidgetlib.RenderLitSizedViewRadiance(res.Program, width, height, extent*swatchRadius, 0, fidgetlib.LitView{
		Light:        light,
		NoAxes:       true,
		Environment:  env,
		GlassSamples: swatchGlassSamples,
	})
}

// renderBasis appends a finish's two renders to out, as half floats.
func renderBasis(out []byte, e Entry, width, height int, extent float64, env string, light, ground bool, groundGrey, checker float64) ([]byte, error) {
	for _, albedo := range basisAlbedos {
		_, radiance, err := renderRadiance(e, albedo, width, height, extent, env, light, ground, groundGrey, checker)
		if err != nil {
			return nil, err
		}
		for _, v := range radiance {
			out = binary.LittleEndian.AppendUint16(out, floatToHalf(v))
		}
	}
	return out, nil
}

// floatToHalf rounds a float32 to the nearest IEEE 754 half. Radiance is
// non-negative and far under the half's 65504; values below its smallest
// normal go to the subnormals.
func floatToHalf(f float32) uint16 {
	if f != f {
		return 0x7e00
	}
	sign := uint16(0)
	if f < 0 {
		sign, f = 0x8000, -f
	}
	if f >= 65520 {
		return sign | 0x7c00
	}
	if f < 6.103515625e-05 { // subnormal: f / 2^-24, rounded
		return sign | uint16(math.RoundToEven(float64(f)*16777216))
	}
	exp := math.Floor(math.Log2(float64(f)))
	mant := float64(f)/math.Pow(2, exp) - 1
	m := math.RoundToEven(mant * 1024)
	if m == 1024 {
		m, exp = 0, exp+1
	}
	return sign | uint16(exp+15)<<10 | uint16(m)
}
