package main

import (
	"bytes"
	"encoding/binary"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// A coat's swatch recoloured from its basis has to be the render of the
// preset in that coat colour, as a finish's is (basis_test.go): the Bodies
// panel draws an adjusted body's thumbnail from it.
func TestTintedCoatIsARender(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("Node is required to run finish-tint.js")
	}
	m, sheet := readManifest(t)
	if m.CoatBasis == nil {
		t.Fatal("the manifest has no coat basis (run `go run ./tool/materialswatch -manifest-only -coat-basis`)")
	}
	raw, err := os.ReadFile(filepath.Join(outputDir, m.CoatBasis.File))
	if err != nil {
		t.Fatal(err)
	}
	px := m.CoatBasis.Width * m.CoatBasis.Height
	if want := len(m.CoatBasis.Order) * 2 * 3 * px * 2; len(raw) != want {
		t.Fatalf("coat basis is %d bytes, want %d for %d coats", len(raw), want, len(m.CoatBasis.Order))
	}
	halves := func(b []byte) []uint16 {
		out := make([]uint16, len(b)/2)
		for i := range out {
			out[i] = binary.LittleEndian.Uint16(b[2*i:])
		}
		return out
	}

	// A colour far from the preset's own red, so a recolouring that ignored
	// it could not pass.
	colour := [3]float64{0.2, 0.5, 0.9}
	want := "coating.redPowderCoat"
	var entry Entry
	for _, e := range m.Entries {
		if e.Package+"."+e.Symbol == want {
			entry = e
		}
	}
	at := -1
	for i, name := range m.CoatBasis.Order {
		if name == want {
			at = i
		}
	}
	if entry.Symbol == "" || at < 0 {
		t.Fatalf("%s is not in the manifest's coats", want)
	}
	size := 3 * px * 2
	r0 := halves(raw[(2*at)*size : (2*at+1)*size])
	r1 := halves(raw[(2*at+1)*size : (2*at+2)*size])

	col, row := entry.Index%m.Cols, entry.Index/m.Cols
	x0, y0 := col*(m.CellWidth+m.Gap), row*(m.CellHeight+m.Gap)
	background := make([]int, 4*px)
	for y := 0; y < m.CellHeight; y++ {
		for x := 0; x < m.CellWidth; x++ {
			r, g, b, a := sheet.At(x0+x, y0+y).RGBA()
			i := 4 * (y*m.CellWidth + x)
			background[i], background[i+1], background[i+2], background[i+3] = int(r>>8), int(g>>8), int(b>>8), int(a>>8)
		}
	}
	tinted := tintWithNode(t, r0, r1, background, colour)

	settings := swatchSettings{extent: 1.8, env: "Softbox", light: true, ground: true, groundGrey: defaultGroundGrey, checker: defaultChecker}
	direct, _, err := renderRadianceDoc(coatDoc(m.Adjust[want], colour, settings), m.CellWidth, m.CellHeight, settings.extent, settings.env, settings.light)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(direct))
	if err != nil {
		t.Fatal(err)
	}
	worst, off := 0, 0
	for y := 0; y < m.CellHeight; y++ {
		for x := 0; x < m.CellWidth; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			got := tinted[4*(y*m.CellWidth+x):]
			for k, v := range []int{int(r >> 8), int(g >> 8), int(b >> 8)} {
				if d := abs(got[k] - v); d > 0 {
					off++
					worst = max(worst, d)
				}
			}
		}
	}
	if worst > 1 {
		t.Errorf("tinted coat differs from a render by up to %d levels (%d channel values differ)", worst, off)
	}
	t.Logf("%d of %d channel values differ, by at most %d", off, 3*px, worst)
}

// The coat document is the preset with its parameters as variables, the
// colour one at the basis albedo and the rest at their defaults.
func TestCoatDocFillsThePreset(t *testing.T) {
	adj := readAdjust(t)["coating.redPowderCoat"]
	if adj == nil {
		t.Fatal("coating.redPowderCoat has no adjust record")
	}
	doc := coatDoc(adj, [3]float64{0, 0, 0}, swatchSettings{ground: false})
	for _, want := range []string{
		"swatchCoatColor = [0, 0, 0];",
		"swatchGlossUnits = 85;",
		"swatchPeel = 0.7;",
		"ball = sphere(50);",
		"ballSurface, ballSurfaceColor, ballHeight, ballSlope = coating.powderCoat(swatchCoatColor, swatchGlossUnits, swatchPeel);",
		"ball.material = ballSurface;",
		"ball.color = ballSurfaceColor;",
		"return ball;",
	} {
		if !strings.Contains(doc, want) {
			t.Errorf("missing %q in:\n%s", want, doc)
		}
	}
}

// Only a preset with exactly one colour parameter has a basis: granite's
// three minerals and lava's crust and glow are not one albedo.
func TestCoatBasisIsForOneColourOnly(t *testing.T) {
	adjust := readAdjust(t)
	for name, want := range map[string]int{
		"coating.redPowderCoat":    0,
		"printed.whiteNylonPowder": 0,
		"stone.greyGranite":        -1,
		"natural.glowingLava":      -1,
		"weathering.rustySteel":    -1,
	} {
		adj := adjust[name]
		if adj == nil {
			t.Fatalf("%s has no adjust record", name)
		}
		if got := coatColorIndex(adj); got != want {
			t.Errorf("%s: colour index %d, want %d", name, got, want)
		}
	}
}
