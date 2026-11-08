package main

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// The Bodies panel draws a finish in the selected body's colour by arithmetic
// on two stored renders (examples/viewer/panels/finish-tint.js). That is only worth doing
// if it is the render: a finish drawn in a colour by the panel has to be what
// the renderer draws in that colour, pixel for pixel, to within the half
// floats' rounding.
func TestTintedFinishIsARender(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("Node is required to run finish-tint.js")
	}
	m, sheet := readManifest(t)
	if m.FinishBasis == nil {
		t.Fatal("the manifest has no finish basis")
	}
	raw, err := os.ReadFile(filepath.Join(outputDir, m.FinishBasis.File))
	if err != nil {
		t.Fatal(err)
	}
	px := m.FinishBasis.Width * m.FinishBasis.Height
	if want := len(m.FinishBasis.Order) * 2 * 3 * px * 2; len(raw) != want {
		t.Fatalf("basis is %d bytes, want %d for %d finishes", len(raw), want, len(m.FinishBasis.Order))
	}
	halves := func(b []byte) []uint16 {
		out := make([]uint16, len(b)/2)
		for i := range out {
			out[i] = binary.LittleEndian.Uint16(b[2*i:])
		}
		return out
	}

	// A colour far from the stand-in grey, so a recolouring that ignored it, or
	// got the channels crossed, could not pass.
	colour := [3]float64{0.85, 0.25, 0.1}
	for _, want := range []string{"material.matte", "material.plastic", "material.chrome"} {
		t.Run(want, func(t *testing.T) {
			var entry Entry
			for _, e := range m.Entries {
				if e.Package+"."+e.Symbol == want {
					entry = e
				}
			}
			at := -1
			for i, name := range m.FinishBasis.Order {
				if name == want {
					at = i
				}
			}
			if entry.Symbol == "" || at < 0 {
				t.Fatalf("%s is not in the manifest's finishes", want)
			}
			size := 3 * px * 2
			r0 := halves(raw[(2*at)*size : (2*at+1)*size])
			r1 := halves(raw[(2*at+1)*size : (2*at+2)*size])

			// The swatch's own pixels, from the sheet, for the ground.
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

			direct, _, err := renderRadiance(entry, colour, m.CellWidth, m.CellHeight, 1.8, "Softbox", true, true, defaultGroundGrey, defaultChecker)
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
				t.Errorf("tinted swatch differs from a render by up to %d levels (%d channel values differ)", worst, off)
			}
			t.Logf("%d of %d channel values differ, by at most %d", off, 3*px, worst)
		})
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// tintWithNode runs finish-tint.js's tintFinish on the arrays.
func tintWithNode(t *testing.T, r0, r1 []uint16, background []int, colour [3]float64) []int {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"type":"module"}`), 0600); err != nil {
		t.Fatal(err)
	}
	src, err := os.ReadFile("../../examples/viewer/panels/finish-tint.js")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "finish-tint.js"), src, 0600); err != nil {
		t.Fatal(err)
	}
	input, err := json.Marshal(map[string]any{"r0": r0, "r1": r1, "background": background, "rgb": colour})
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.CommandContext(t.Context(), "node", "--input-type=module", "-e", `
import { readFileSync } from 'node:fs';
const { tintFinish } = await import(process.argv[1]);
const r = JSON.parse(readFileSync(0, 'utf8'));
const out = tintFinish(Uint16Array.from(r.r0), Uint16Array.from(r.r1), Uint8ClampedArray.from(r.background), r.rgb);
console.log(JSON.stringify(Array.from(out)));
`, filepath.Join(dir, "finish-tint.js"))
	cmd.Stdin = bytes.NewReader(input)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("finish-tint.js: %v", err)
	}
	var px []int
	if err := json.Unmarshal(out, &px); err != nil {
		t.Fatal(err)
	}
	return px
}
