package main

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// The generated sheet and manifest, as committed.
const outputDir = "../../examples/viewer/materials"

func readManifest(t *testing.T) (Manifest, image.Image) {
	t.Helper()
	body, err := os.ReadFile(filepath.Join(outputDir, "catalog.json"))
	if err != nil {
		t.Fatalf("read manifest (run `go run ./tool/materialswatch`): %v", err)
	}
	var m Manifest
	if err := json.Unmarshal(body, &m); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(filepath.Join(outputDir, m.Image))
	if err != nil {
		t.Fatalf("read sheet: %v", err)
	}
	sheet, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	return m, sheet
}

// The manifest's geometry has to describe the sheet that is actually there.
// The cells are not flush — ComposeSheet leaves a gap whose size is a constant
// in another module — so a panel that assumed they were would draw every
// swatch sliced between its neighbours, and nothing would look obviously
// broken enough to notice.
func TestManifestGeometryFitsTheSheet(t *testing.T) {
	m, sheet := readManifest(t)
	b := sheet.Bounds()

	if b.Dx() != m.SheetWidth || b.Dy() != m.SheetHeight {
		t.Fatalf("sheet is %dx%d, manifest says %dx%d", b.Dx(), b.Dy(), m.SheetWidth, m.SheetHeight)
	}
	if want := m.Cols*m.CellWidth + (m.Cols-1)*m.Gap; want != m.SheetWidth {
		t.Errorf("%d cells of %dpx with a %dpx gap make %d, not %d", m.Cols, m.CellWidth, m.Gap, want, m.SheetWidth)
	}
	for _, e := range m.Entries {
		col, row := e.Index%m.Cols, e.Index/m.Cols
		x := col * (m.CellWidth + m.Gap)
		y := row * (m.CellHeight + m.Gap)
		if x+m.CellWidth > m.SheetWidth || y+m.CellHeight > m.SheetHeight {
			t.Fatalf("%s at index %d lands at (%d,%d), outside the %dx%d sheet",
				e.Call(), e.Index, x, y, m.SheetWidth, m.SheetHeight)
		}
	}
}

// Every annotated surface has to be in the committed manifest: a library that
// grows a material and is not regenerated would otherwise just be missing from
// the panel, silently.
func TestCommittedManifestIsCurrent(t *testing.T) {
	m, _ := readManifest(t)
	live := ReadCatalog()

	if len(live) != len(m.Entries) {
		t.Fatalf("the libraries declare %d surfaces, the manifest holds %d — run `go run ./tool/materialswatch`",
			len(live), len(m.Entries))
	}
	for i, want := range live {
		if got := m.Entries[i]; got.Kind != want.Kind || got.Package != want.Package || got.Symbol != want.Symbol {
			t.Errorf("entry %d is %s in the manifest and %s in the libraries — run `go run ./tool/materialswatch`",
				i, got.Call(), want.Call())
		}
	}

	// What each preset can adjust is read off its constructors' parameter
	// annotations, so a library that annotates a parameter and is not
	// regenerated would leave the panel offering the old knobs.
	adjust, err := ReadAdjust()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(adjust, m.Adjust) {
		t.Errorf("the manifest's adjust records differ from the libraries' — run `go run ./tool/materialswatch -manifest-only`")
	}
}

// The strongest check available: crop a cell out of the committed sheet at the
// coordinates the manifest gives, render that same material again now, and
// require the two to be identical. That tests the geometry and the sheet's
// currency at once — if the cell holds a different material, or the right one
// drawn with different settings, the pixels differ.
func TestSheetCellsAreTheMaterialsTheManifestNames(t *testing.T) {
	m, sheet := readManifest(t)

	// A spread: a finish, a mirror metal, a transmissive material whose
	// refraction depends on the ground, and the last entry, which is the one a
	// short final row would misplace.
	for _, want := range []string{"material.chrome", "metal.gold", "glass.clear", "physical.plasticPVC"} {
		var entry Entry
		for _, e := range m.Entries {
			if e.Package+"."+e.Symbol == want {
				entry = e
				break
			}
		}
		if entry.Symbol == "" {
			t.Fatalf("%s is not in the manifest", want)
		}

		t.Run(want, func(t *testing.T) {
			raw, err := renderSwatch(entry, m.CellWidth, m.CellHeight, 1.8, "Softbox", true, true, defaultGroundGrey, defaultChecker)
			if err != nil {
				t.Fatal(err)
			}
			fresh, err := png.Decode(bytes.NewReader(raw))
			if err != nil {
				t.Fatal(err)
			}

			col, row := entry.Index%m.Cols, entry.Index/m.Cols
			originX := col * (m.CellWidth + m.Gap)
			originY := row * (m.CellHeight + m.Gap)
			for y := 0; y < m.CellHeight; y++ {
				for x := 0; x < m.CellWidth; x++ {
					got := sheet.At(originX+x, originY+y)
					exp := fresh.At(fresh.Bounds().Min.X+x, fresh.Bounds().Min.Y+y)
					gr, gg, gb, _ := got.RGBA()
					er, eg, eb, _ := exp.RGBA()
					if gr != er || gg != eg || gb != eb {
						t.Fatalf("cell (%d,%d) differs from a fresh render of %s at pixel (%d,%d): sheet %v, render %v",
							col, row, entry.Call(), x, y,
							[3]uint32{gr >> 8, gg >> 8, gb >> 8}, [3]uint32{er >> 8, eg >> 8, eb >> 8})
					}
				}
			}
		})
	}
}
