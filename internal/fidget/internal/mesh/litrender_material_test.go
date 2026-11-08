package mesh

import (
	"math"
	"testing"
)

// The material property has to travel a long way to reach a pixel: the resolver
// lowers it per return, the build hook stores it as trees, the renderer
// evaluates twelve of them at each hit, and the BRDF reads the rows. These
// tests check the ends of that path — a material the document authors must
// change the picture in the direction the row means — rather than any
// particular shaded value, which belongs to the shader the shading is ported
// from (common/web/gsl-shading.js).

const litMatBall = "ball = length(p) - 3.0;\n"

// A shape whose material is emissive must come out far brighter than the same
// shape with no material at all. Emission is the one row added after exposure,
// so it moves the picture unmistakably and cannot be confused with a change in
// the light rig.
func TestRenderLitEmissiveMaterialIsBrighter(t *testing.T) {
	plain := litMatBall + "ball.color = [0.02, 0.02, 0.02];\nreturn ball;\n"
	glowing := "import \"material\";\n" + litMatBall +
		"ball.color = [0.02, 0.02, 0.02];\n" +
		"ball.material = material.glowing([1.0, 1.0, 1.0], 3.0);\n" +
		"return ball;\n"

	center := func(src string) [3]float64 {
		t.Helper()
		raw, err := RenderLit(litProgramWithLibs(t, src), LitOptions{Width: 120, Height: 80, WorldExtent: 6})
		if err != nil {
			t.Fatal(err)
		}
		return rgbAt(decodeLit(t, raw), 60, 40)
	}

	dark, bright := center(plain), center(glowing)
	if sum(dark) > 0.9 {
		t.Fatalf("an almost-black ball with no material rendered at %.3v; the test cannot tell emission from albedo", dark)
	}
	if sum(bright) < sum(dark)+1.0 {
		t.Errorf("emissive ball %.3v is not brighter than the same ball with no material %.3v", bright, dark)
	}
}

// Metallic is the row that changes the BRDF's shape rather than its scale: a
// metal takes its specular tint from the albedo and has no diffuse lobe at all,
// so a red metal is markedly darker than a red dielectric under a near-white
// rig. If the material never arrived, both renders would be the dielectric and
// the two pictures would be identical.
func TestRenderLitMetallicMaterialChangesTheSurface(t *testing.T) {
	doc := func(mat string) string {
		return "import \"material\";\n" + litMatBall +
			"ball.color = [0.8, 0.1, 0.1];\n" +
			"ball.material = " + mat + ";\n" +
			"return ball;\n"
	}
	center := func(src string) [3]float64 {
		t.Helper()
		raw, err := RenderLit(litProgramWithLibs(t, src), LitOptions{Width: 120, Height: 80, WorldExtent: 6})
		if err != nil {
			t.Fatal(err)
		}
		return rgbAt(decodeLit(t, raw), 60, 40)
	}

	dielectric := center(doc("material.make(0.3, 0, 1.5)"))
	metal := center(doc("material.make(0.3, 1, 1.5)"))
	if math.Abs(sum(metal)-sum(dielectric)) < 0.1 {
		t.Errorf("metallic %.3v and dielectric %.3v render the same; the material row is not reaching the BRDF", metal, dielectric)
	}
	if sum(metal) >= sum(dielectric) {
		t.Errorf("metallic %.3v is not darker than dielectric %.3v: a metal has no diffuse lobe", metal, dielectric)
	}
}

// Transmission takes its share out of the diffuse lobe before the refracted
// ray adds back what lies behind. Behind this ball is only the dark canvas, so
// a fully transmissive red ball must lose its red body colour rather than keep
// it.
func TestRenderLitTransmissionDimsTheDiffuseLobe(t *testing.T) {
	doc := func(transmission string) string {
		return "import \"material\";\n" + litMatBall +
			"ball.color = [0.9, 0.1, 0.1];\n" +
			"ball.material = material.translucent(material.make(0.1, 0, 1.5), " + transmission + ", 0.0);\n" +
			"return ball;\n"
	}
	center := func(src string) [3]float64 {
		t.Helper()
		raw, err := RenderLit(litProgramWithLibs(t, src), LitOptions{Width: 120, Height: 80, WorldExtent: 6})
		if err != nil {
			t.Fatal(err)
		}
		return rgbAt(decodeLit(t, raw), 60, 40)
	}

	opaque, clear := center(doc("0.0")), center(doc("1.0"))
	if opaque[0] <= opaque[1]+0.15 {
		t.Fatalf("opaque ball %.3v is not dominantly red; the test cannot measure the diffuse lobe", opaque)
	}
	if clear[0]-clear[1] >= opaque[0]-opaque[1] {
		t.Errorf("transmission did not dim the red diffuse lobe: clear %.3v vs opaque %.3v", clear, opaque)
	}
}

// The defining behaviour of the glass path: a part inside clear glass is seen
// through it. A red core sits inside a clear sphere; the centre ray enters the
// sphere along its normal, so refraction does not bend it off the core, and
// the refracted march must find the core and shade it red. The same sphere made
// opaque hides the core, which is what separates "seen through" from "the
// sphere happens to be red".
func TestRenderLitSeesThroughGlass(t *testing.T) {
	doc := func(prelude, shellMaterial string) string {
		return "import \"glass\";\nimport \"material\";\n" +
			prelude +
			"shell = length(p) - 4.0;\n" +
			"shell.material = " + shellMaterial + ";\n" +
			"core = length(p) - 1.5;\n" +
			"core.color = [1.0, 0.0, 0.0];\n" +
			"return shell, core;\n"
	}
	center := func(src string) [3]float64 {
		t.Helper()
		raw, err := RenderLit(litProgramWithLibs(t, src), LitOptions{Width: 120, Height: 80, WorldExtent: 8})
		if err != nil {
			t.Fatal(err)
		}
		return rgbAt(decodeLit(t, raw), 60, 40)
	}

	red := func(c [3]float64) bool { return c[0] > c[1]+0.15 && c[0] > c[2]+0.15 }

	through := center(doc("clearMat, clearColor = glass.clear();\n", "clearMat"))
	if !red(through) {
		t.Errorf("the red core is not seen through clear glass: centre %.3v", through)
	}
	opaque := center(doc("", "material.standard"))
	if red(opaque) {
		t.Errorf("an opaque shell shows the red core it should hide: centre %.3v", opaque)
	}
}

// Displacement moves the rendered surface itself, not just its shading. A
// constant height of 1 on a radius-3 ball must render a radius-4 ball, so its
// silhouette covers about (4/3)^2 = 1.78 times the pixels of the same ball
// with the zero-height preset. A constant height is the case worth locking:
// only a height known to be zero may skip the displaced field, so a constant
// that is not zero has to move the surface.
func TestRenderLitDisplacementMovesTheSurface(t *testing.T) {
	doc := func(mat string) string {
		return "import \"material\";\n" + litMatBall +
			"ball.material = " + mat + ";\n" +
			"return ball;\n"
	}
	covered := func(src string) int {
		t.Helper()
		raw, err := RenderLit(litProgramWithLibs(t, src), LitOptions{Width: 120, Height: 80, WorldExtent: 8, NoAxes: true})
		if err != nil {
			t.Fatal(err)
		}
		img := decodeLit(t, raw)
		n := 0
		for y := 0; y < 80; y++ {
			for x := 0; x < 120; x++ {
				c := rgbAt(img, x, y)
				if math.Abs(c[0]-litBackground[0])+math.Abs(c[1]-litBackground[1])+math.Abs(c[2]-litBackground[2]) > 0.02 {
					n++
				}
			}
		}
		return n
	}

	flat := covered(doc("material.standard"))
	raised := covered(doc("material.withDisplacement(material.standard, 1.0)"))
	if flat == 0 {
		t.Fatal("the undisplaced ball rendered no pixels")
	}
	if ratio := float64(raised) / float64(flat); ratio < 1.5 || ratio > 2.1 {
		t.Errorf("a height of 1 on a radius-3 ball covers %.2fx the pixels of the undisplaced ball, want about 1.78 (%d vs %d)", ratio, raised, flat)
	}
}
