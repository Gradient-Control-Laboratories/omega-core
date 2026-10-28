package mesh

import (
	"bytes"
	"image"
	"image/png"
	"math"
	"testing"

	assembler "gradientcontrol.com/omega/core/internal/assembler"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/program"
)

var testPkgMgr = source.NewLocalManager("../../../../common/source/package")

func litProgram(t *testing.T, src string) *program.Program {
	t.Helper()
	res, err := assembler.Assemble([]string{src}, testPkgMgr)
	if err != nil {
		t.Fatalf("assemble: %v", err)
	}
	return res.Program
}

// litProgramWithLibs is litProgram for sources that import the GSL standard
// libraries, resolved from the repo's package root relative to this test.
func litProgramWithLibs(t *testing.T, src string) *program.Program {
	t.Helper()
	return litProgram(t, src)
}

func decodeLit(t *testing.T, raw []byte) image.Image {
	t.Helper()
	img, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("decode png: %v", err)
	}
	return img
}

// rgbAt reads a pixel as 0..1 floats.
func rgbAt(img image.Image, x, y int) [3]float64 {
	r, g, b, _ := img.At(x, y).RGBA()
	return [3]float64{float64(r) / 65535, float64(g) / 65535, float64(b) / 65535}
}

const litBallDoc = "ball = length(p) - 3.0;\nreturn ball;\n"

func TestRenderLitShapeAndBackground(t *testing.T) {
	raw, err := RenderLit(litProgram(t, litBallDoc), LitOptions{Width: 120, Height: 80, WorldExtent: 6})
	if err != nil {
		t.Fatal(err)
	}
	img := decodeLit(t, raw)
	if got := img.Bounds().Dx(); got != 120 {
		t.Errorf("width = %d, want 120", got)
	}

	// A corner is background: the viewer's 0x1e1e1e.
	corner := rgbAt(img, 1, 1)
	for k, want := range litBackground {
		if math.Abs(corner[k]-want) > 0.02 {
			t.Errorf("corner channel %d = %.3f, want %.3f (background)", k, corner[k], want)
		}
	}
	// The center is the lit sphere, and lighting must make it brighter than
	// the background — otherwise the shading collapsed.
	center := rgbAt(img, 60, 40)
	if sum(center) <= sum(corner)+0.1 {
		t.Errorf("center %.3v not lit above background %.3v", center, corner)
	}
}

func TestRenderLitUsesAuthoredColor(t *testing.T) {
	// A pure-red color field must dominate the center pixel: the rig's lights
	// are near-white, so red stays far ahead of green and blue.
	doc := "ball = length(p) - 3.0;\nball.color = [1.0, 0.0, 0.0];\nreturn ball;\n"
	raw, err := RenderLit(litProgram(t, doc), LitOptions{Width: 120, Height: 80, WorldExtent: 6})
	if err != nil {
		t.Fatal(err)
	}
	center := rgbAt(decodeLit(t, raw), 60, 40)
	if center[0] < 0.3 {
		t.Errorf("red channel %.3f too dark for a red color field (%.3v)", center[0], center)
	}
	if center[0] <= center[1]+0.15 || center[0] <= center[2]+0.15 {
		t.Errorf("center %.3v is not dominantly red", center)
	}
}

// TestRenderLitNonFiniteGradientStillRenders locks the march's guard: a smooth
// union's gradient goes non-finite far from its surfaces, and a NaN step would
// strand every ray (the whole image came back empty before the guard).
func TestRenderLitNonFiniteGradientStillRenders(t *testing.T) {
	doc := "import . \"omega\";\n" +
		"a = sphere([2.0, 0.0, 0.0], 2.0);\n" +
		"b = sphere([-2.0, 0.0, 0.0], 2.0);\n" +
		"blend = unionRound(a, b, 1.5);\n" +
		"return blend;\n"
	raw, err := RenderLit(litProgramWithLibs(t, doc), LitOptions{Width: 120, Height: 80, WorldExtent: 8})
	if err != nil {
		t.Fatal(err)
	}
	img := decodeLit(t, raw)
	lit := 0
	for y := 0; y < 80; y++ {
		for x := 0; x < 120; x++ {
			if sum(rgbAt(img, x, y)) > sum(litBackground)+0.1 {
				lit++
			}
		}
	}
	if lit == 0 {
		t.Fatal("nothing rendered: the march stranded every ray")
	}
}

func TestRenderLitEmptySceneIsAllBackground(t *testing.T) {
	// A field that is positive everywhere has no surface to hit: every pixel
	// is the background or the world-axis overlay, never stray shading.
	raw, err := RenderLit(litProgram(t, "far = length(p) + 100.0;\nreturn far;\n"),
		LitOptions{Width: 40, Height: 30, WorldExtent: 5})
	if err != nil {
		t.Fatal(err)
	}
	img := decodeLit(t, raw)
	isAxis := func(c [3]float64) bool {
		for _, ax := range [][3]float64{{0.85, 0.30, 0.30}, {0.30, 0.72, 0.30}, {0.32, 0.52, 0.90}} {
			if math.Abs(c[0]-ax[0]) < 0.02 && math.Abs(c[1]-ax[1]) < 0.02 && math.Abs(c[2]-ax[2]) < 0.02 {
				return true
			}
		}
		return false
	}
	for y := 0; y < 30; y++ {
		for x := 0; x < 40; x++ {
			got := rgbAt(img, x, y)
			if math.Abs(got[0]-litBackground[0]) > 0.02 && !isAxis(got) {
				t.Fatalf("pixel (%d,%d) = %.3v, want background or an axis", x, y, got)
			}
		}
	}
}

func sum(v [3]float64) float64 { return v[0] + v[1] + v[2] }

// TestRenderLitAutoFrameFitsTallGeometry locks the audit finding that
// motivated the perspective-correct auto-frame: a stood-up 20-long slab was
// rendered clipped top and bottom, an ambiguous wall filling the frame. With
// the fit correct, the whole part sits inside the image: the top and bottom
// pixel rows are background (or the axis overlay), never part surface.
func TestRenderLitAutoFrameFitsTallGeometry(t *testing.T) {
	standing := "boxHalf(half) {\n  return max(\n    -half.x - p.x, p.x - half.x,\n" +
		"    -half.y - p.y, p.y - half.y,\n    -half.z - p.z, p.z - half.z,\n  );\n}\n" +
		"slab = remap(boxHalf([10.0, 3.0, 1.5]), p.zyx);\nreturn slab;\n"
	raw, err := RenderLit(litProgram(t, standing), LitOptions{Width: 120, Height: 80})
	if err != nil {
		t.Fatal(err)
	}
	img := decodeLit(t, raw)
	surface := func(c [3]float64) bool {
		// Palette-lit surface: noticeably brighter than the dark background
		// and not one of the saturated axis colors.
		if sum(c) < sum(litBackground)+0.15 {
			return false
		}
		return !(c[0] > c[1]+0.3 || c[2] > c[0]+0.3 || c[1] > c[0]+0.25)
	}
	for _, y := range []int{0, 79} {
		for x := 0; x < 120; x++ {
			if surface(rgbAt(img, x, y)) {
				t.Fatalf("row %d pixel %d is part surface — geometry clipped by the frame", y, x)
			}
		}
	}
	lit := 0
	for y := 0; y < 80; y++ {
		for x := 0; x < 120; x++ {
			if surface(rgbAt(img, x, y)) {
				lit++
			}
		}
	}
	if lit < 120*80/20 {
		t.Fatalf("part occupies %d pixels — auto-frame lost the geometry", lit)
	}
}

// TestRenderLitOrientationIsVisible: the lying and standing slab must render
// visibly differently under auto-framing — the property the feedback images
// need before a model can read orientation from them at all.
func TestRenderLitOrientationIsVisible(t *testing.T) {
	base := "boxHalf(half) {\n  return max(\n    -half.x - p.x, p.x - half.x,\n" +
		"    -half.y - p.y, p.y - half.y,\n    -half.z - p.z, p.z - half.z,\n  );\n}\n"
	lying := base + "slab = boxHalf([10.0, 3.0, 1.5]);\nreturn slab;\n"
	standing := base + "slab = remap(boxHalf([10.0, 3.0, 1.5]), p.zyx);\nreturn slab;\n"
	imgs := make([]image.Image, 2)
	for i, doc := range []string{lying, standing} {
		raw, err := RenderLit(litProgram(t, doc), LitOptions{Width: 120, Height: 80})
		if err != nil {
			t.Fatal(err)
		}
		imgs[i] = decodeLit(t, raw)
	}
	diff := 0
	for y := 0; y < 80; y++ {
		for x := 0; x < 120; x++ {
			a, b := rgbAt(imgs[0], x, y), rgbAt(imgs[1], x, y)
			if math.Abs(a[0]-b[0])+math.Abs(a[1]-b[1])+math.Abs(a[2]-b[2]) > 0.1 {
				diff++
			}
		}
	}
	if diff < 120*80/10 {
		t.Fatalf("lying vs standing differ at only %d pixels — orientation is not visible", diff)
	}
}
