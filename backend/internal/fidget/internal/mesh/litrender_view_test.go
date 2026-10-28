package mesh

import (
	"bytes"
	"image"
	"math"
	"testing"
)

// The camera options must be invisible to every existing caller: spelling
// the default view out explicitly renders the same bytes as leaving it zero.
func TestRenderLitExplicitDefaultViewIsIdentical(t *testing.T) {
	prog := litProgram(t, litBallDoc)
	zero, err := RenderLit(prog, LitOptions{Width: 120, Height: 80, WorldExtent: 6})
	if err != nil {
		t.Fatal(err)
	}
	explicit, err := RenderLit(prog, LitOptions{
		Width: 120, Height: 80, WorldExtent: 6,
		View: [3]float64{1, 1, 1}, Up: [3]float64{0, 0, 1},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(zero, explicit) {
		t.Fatal("explicit (1,1,1)/+Z view renders differently from the zero options")
	}
}

func countDiff(a, b image.Image, w, h int) int {
	diff := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			p, q := rgbAt(a, x, y), rgbAt(b, x, y)
			if math.Abs(p[0]-q[0])+math.Abs(p[1]-q[1])+math.Abs(p[2]-q[2]) > 0.1 {
				diff++
			}
		}
	}
	return diff
}

// A lying slab seen from above is a wide rectangle; from the front it is a
// thin bar. The two cameras must produce visibly different pictures.
func TestRenderLitViewDirectionChangesThePicture(t *testing.T) {
	doc := "boxHalf(half) {\n  return max(\n    -half.x - p.x, p.x - half.x,\n" +
		"    -half.y - p.y, p.y - half.y,\n    -half.z - p.z, p.z - half.z,\n  );\n}\n" +
		"slab = boxHalf([10.0, 3.0, 1.5]);\nreturn slab;\n"
	prog := litProgram(t, doc)
	top, err := RenderLit(prog, LitOptions{Width: 120, Height: 80, WorldExtent: 12,
		View: [3]float64{0, 0, 1}, Up: [3]float64{0, 1, 0}})
	if err != nil {
		t.Fatal(err)
	}
	front, err := RenderLit(prog, LitOptions{Width: 120, Height: 80, WorldExtent: 12,
		View: [3]float64{0, -1, 0}})
	if err != nil {
		t.Fatal(err)
	}
	if diff := countDiff(decodeLit(t, top), decodeLit(t, front), 120, 80); diff < 120*80/10 {
		t.Fatalf("top vs front differ at only %d pixels", diff)
	}
}

// A peg hanging below a flange is hidden from the corner-on camera above and
// visible from the iso-below camera — the case the feedback images were
// silently failing on.
func TestRenderLitIsoBelowSeesTheUnderside(t *testing.T) {
	doc := "boxHalf(half) {\n  return max(\n    -half.x - p.x, p.x - half.x,\n" +
		"    -half.y - p.y, p.y - half.y,\n    -half.z - p.z, p.z - half.z,\n  );\n}\n" +
		"flange = boxHalf([8.0, 8.0, 1.0]);\n" +
		"peg = max(length(p.xy) - 1.5, -p.z - 8.0, p.z + 1.0);\n" +
		"peg.color = [1.0, 0.0, 0.0];\n" +
		"return flange, peg;\n"
	prog := litProgram(t, doc)
	red := func(raw []byte) int {
		img := decodeLit(t, raw)
		n := 0
		for y := 0; y < 80; y++ {
			for x := 0; x < 120; x++ {
				c := rgbAt(img, x, y)
				if c[0] > 0.3 && c[0] > c[1]+0.15 && c[0] > c[2]+0.15 {
					n++
				}
			}
		}
		return n
	}
	above, err := RenderLit(prog, LitOptions{Width: 120, Height: 80, WorldExtent: 10})
	if err != nil {
		t.Fatal(err)
	}
	below, err := RenderLit(prog, LitOptions{Width: 120, Height: 80, WorldExtent: 10,
		View: [3]float64{1, 1, -1}, Up: [3]float64{0, 0, -1}, NoAxes: true})
	if err != nil {
		t.Fatal(err)
	}
	// From the corner above, the flange hides all but the peg's tip poking
	// past its edge in perspective; from below the peg is the subject.
	nAbove, nBelow := red(above), red(below)
	if nBelow < 40 {
		t.Errorf("iso-below camera sees only %d red pixels of the hanging peg", nBelow)
	}
	if nBelow < 5*nAbove {
		t.Errorf("iso-below sees %d red pixels vs %d from the corner above; the underside camera is not looking underneath", nBelow, nAbove)
	}
}

func TestRenderLitLightGroundAndNoAxes(t *testing.T) {
	raw, err := RenderLit(litProgram(t, litBallDoc), LitOptions{Width: 120, Height: 80, WorldExtent: 6, Light: true, NoAxes: true})
	if err != nil {
		t.Fatal(err)
	}
	img := decodeLit(t, raw)
	corner := rgbAt(img, 1, 1)
	if corner[0] < 0.85 || corner[1] < 0.85 || corner[2] < 0.85 {
		t.Errorf("corner %.3v is not a light ground", corner)
	}
	// With the overlay off, no pixel carries an axis colour.
	for y := 0; y < 80; y++ {
		for x := 0; x < 120; x++ {
			c := rgbAt(img, x, y)
			for _, ax := range [][3]float64{{0.85, 0.30, 0.30}, {0.30, 0.72, 0.30}, {0.32, 0.52, 0.90}} {
				if math.Abs(c[0]-ax[0]) < 0.02 && math.Abs(c[1]-ax[1]) < 0.02 && math.Abs(c[2]-ax[2]) < 0.02 {
					t.Fatalf("pixel (%d,%d) = %.3v is an axis colour with NoAxes set", x, y, c)
				}
			}
		}
	}
}
