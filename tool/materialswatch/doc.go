package main

import (
	"fmt"
	"math"
)

// The swatch scene. A sphere 10 cm across at the origin, framed corner-on,
// over a plane just below it. The size is a part's, in mm, because a material's
// features are: a granite's 4 mm grains or a marble's 25 mm veins drawn on a
// sphere a millimetre across fill it with one flat colour. Every other length
// in the scene — the frame, the drop to the plane, the checker — is in radii,
// so the scene is the same picture at any size.
//
// The plane is not decoration. A swatch has to show what a material DOES, and
// three of the rows only show themselves in what the surface gathers from
// elsewhere: a metal is its reflection, a transmissive material is what lies
// behind it, and a clearcoat is a second highlight. A sphere alone in the void
// renders clear glass as a faint rim and an absorption distance as nothing at
// all. So every swatch gets the same neutral ground to reflect and refract,
// and all 142 are drawn in the one scene — a swatch's job is to be comparable
// with the swatch beside it, not to flatter its own material.
const (
	swatchRadius = 50.0
	// groundDrop puts the plane below the sphere by half a radius (it is in
	// radii, as the checker and the frame are): close
	// enough to be gathered by the lower hemisphere, far enough that the
	// sphere reads as a sphere and not as a cap.
	groundDrop = 1.5
	// defaultGroundGrey is dark enough to stay behind the material and light
	// enough to carry a reflection. It is the swatch's only free parameter and
	// the same for every entry; -ground-grey sweeps it.
	defaultGroundGrey = 0.35
	// standInGrey is the albedo a @finish is drawn against. A finish carries
	// no colour on purpose (internal/assembler/annotation), so the panel has to
	// supply one to show it at all; a neutral grey shows roughness and
	// metalness without implying an albedo the finish does not have.
	standInGrey = 0.7
	// defaultChecker is the ground's checker cell, in radii. Swept at the swatch's own
	// 96px: finer than this and the board reads as noise behind 142
	// thumbnails, coarser and it competes with the sphere in front of it, and
	// a clear sphere stops reading as a lens because too few squares fall
	// inside it to show the inversion.
	defaultChecker = 0.3
	// swatchGlassSamples is the refracted rays per pixel a rough
	// transmitting material takes in a swatch: fifty times the viewer's
	// four, since a swatch is rendered once and frosted glass is grain at
	// four and still visibly grainy at twenty in the large preview.
	swatchGlassSamples = 200
)

// swatchDoc is the GSL document that renders one entry's swatch. A material is
// destructured into its pair, which is the contract the libraries now carry; a
// finish is assigned on its own and takes the stand-in grey.
func swatchDoc(e Entry, ground bool, groundGrey, checker float64) string {
	return swatchDocIn(e, [3]float64{standInGrey, standInGrey, standInGrey}, ground, groundGrey, checker)
}

// swatchDocIn is swatchDoc with a finish drawn in albedo (sRGB) rather than the
// stand-in grey: the basis renders a finish is recoloured from (basis.go), and
// the direct render a recolouring is checked against.
func swatchDocIn(e Entry, albedo [3]float64, ground bool, groundGrey, checker float64) string {
	var doc string
	switch e.Kind {
	case "finish":
		doc = fmt.Sprintf(""+
			"ball = sphere(%g);\n"+
			"ball.material = %s;\n"+
			"ball.color = [%g, %g, %g];\n",
			swatchRadius, e.Call(), albedo[0], albedo[1], albedo[2])
	default:
		doc = fmt.Sprintf(""+
			"swatchMat, swatchColor = %s;\n"+
			"ball = sphere(%g);\n"+
			"ball.material = swatchMat;\n"+
			"ball.color = swatchColor;\n",
			e.Call(), swatchRadius)
	}
	return swatchTail(doc, ground, groundGrey, checker)
}

// swatchTail finishes a swatch document whose sphere is `ball`: the ground
// below it when there is one, and the return.
func swatchTail(doc string, ground bool, groundGrey, checker float64) string {
	if !ground {
		return doc + "return ball;\n"
	}
	return doc + fmt.Sprintf(""+
		"swatchGround = p.z + %g;\n"+
		"swatchGround.material = material.matte;\n"+
		"swatchGround.color = %s;\n"+
		"return ball, swatchGround;\n",
		groundDrop*swatchRadius, groundColor(groundGrey, checker*swatchRadius))
}

// groundColor is the plane's albedo: a flat grey, or a checkerboard of that
// grey against twice it when checker is a cell size.
//
// A checker is what a transmissive material needs to show itself. A flat ground
// refracted is still flat — a clear sphere over it reads as a dark ball with a
// bright rim, and tells you nothing about how strongly it bends. A board's
// squares come through bent, inverted and magnified, which is the one thing a
// swatch can show about an index of refraction. playful.checkerboard averages
// each square over the pixel, so the cells that fall below the swatch's
// resolution settle to the two colours' mix instead of aliasing.
func groundColor(grey, checker float64) string {
	if checker <= 0 {
		return fmt.Sprintf("[%g, %g, %g]", grey, grey, grey)
	}
	light := math.Min(1, 2*grey)
	return fmt.Sprintf("playful.checkerboard(%g, [%g, %g, %g], [%g, %g, %g])",
		checker, grey, grey, grey, light, light, light)
}
