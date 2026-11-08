package main

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// The hover preview's large images are the materials the manifest names them
// after, rendered as the generator renders them now: a metal, and frosted
// glass, whose rough transmission is what the swatches' extra samples are for.
func TestLargeImagesAreTheirMaterials(t *testing.T) {
	m, _ := readManifest(t)
	if m.Large == nil {
		t.Fatal("the manifest has no large images")
	}
	for _, want := range []string{"metal.gold", "glass.frosted"} {
		t.Run(want, func(t *testing.T) {
			var entry Entry
			for _, e := range m.Entries {
				if e.Package+"."+e.Symbol == want {
					entry = e
				}
			}
			if entry.Symbol == "" {
				t.Fatalf("%s is not in the manifest", want)
			}
			committed, err := os.ReadFile(filepath.Join(outputDir, m.Large.Dir, largeName(entry)+".png"))
			if err != nil {
				t.Fatal(err)
			}
			fresh, err := renderSwatch(entry, m.Large.Width, m.Large.Height, 1.8, "Softbox", true, true, defaultGroundGrey, defaultChecker)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(committed, fresh) {
				t.Errorf("the committed large image differs from a fresh render (run `go run ./tool/materialswatch`)")
			}
		})
	}
}

// A finish's large preview, drawn in a colour from its gzipped renders, is the
// renderer's own render in that colour, as the swatch-sized one is
// (TestTintedFinishIsARender).
func TestLargeTintedFinishIsARender(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("Node is required to run finish-tint.js")
	}
	m, _ := readManifest(t)
	if m.Large == nil {
		t.Fatal("the manifest has no large images")
	}
	var entry Entry
	for _, e := range m.Entries {
		if e.Package+"."+e.Symbol == "material.chrome" {
			entry = e
		}
	}
	w, h := m.Large.Width, m.Large.Height
	base := filepath.Join(outputDir, m.Large.Dir, largeName(entry))

	gz, err := os.Open(base + ".basis.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer gz.Close()
	zr, err := gzip.NewReader(gz)
	if err != nil {
		t.Fatal(err)
	}
	raw, err := io.ReadAll(zr)
	if err != nil {
		t.Fatal(err)
	}
	if want := 2 * 3 * w * h * 2; len(raw) != want {
		t.Fatalf("large basis is %d bytes, want %d", len(raw), want)
	}
	halves := make([]uint16, len(raw)/2)
	for i := range halves {
		halves[i] = binary.LittleEndian.Uint16(raw[2*i:])
	}

	grey, err := os.ReadFile(base + ".png")
	if err != nil {
		t.Fatal(err)
	}
	greyImg, err := png.Decode(bytes.NewReader(grey))
	if err != nil {
		t.Fatal(err)
	}
	background := make([]int, 4*w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, a := greyImg.At(x, y).RGBA()
			i := 4 * (y*w + x)
			background[i], background[i+1], background[i+2], background[i+3] = int(r>>8), int(g>>8), int(b>>8), int(a>>8)
		}
	}
	colour := [3]float64{0.15, 0.55, 0.3}
	tinted := tintWithNode(t, halves[:3*w*h], halves[3*w*h:], background, colour)

	direct, _, err := renderRadiance(entry, colour, w, h, 1.8, "Softbox", true, true, defaultGroundGrey, defaultChecker)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(direct))
	if err != nil {
		t.Fatal(err)
	}
	worst := 0
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			got := tinted[4*(y*w+x):]
			for k, v := range []int{int(r >> 8), int(g >> 8), int(b >> 8)} {
				worst = max(worst, abs(got[k]-v))
			}
		}
	}
	if worst > 1 {
		t.Errorf("large tinted chrome differs from a render by up to %d levels", worst)
	}
}
