package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/eval"
)

// Manifest is what the browser panel reads: the sheet's geometry and the
// catalog, in sheet order. An entry's Index is its cell, left to right and top
// to bottom, so the panel needs no lookup table to place a swatch.
type Manifest struct {
	Image       string  `json:"image"`      // the sheet's file name, beside this manifest
	CellWidth   int     `json:"cellWidth"`  // one swatch, in pixels
	CellHeight  int     `json:"cellHeight"` //
	Cols        int     `json:"cols"`       // cells across the sheet
	Rows        int     `json:"rows"`       // cells down it, the last row short
	Gap         int     `json:"gap"`        // pixels between cells; see sheetGeometry
	SheetWidth  int     `json:"sheetWidth"` // the whole sheet, so a consumer can scale it
	SheetHeight int     `json:"sheetHeight"`
	Entries     []Entry `json:"entries"`
	// Colors are the sRGB colours each material paints, keyed by
	// package.symbol: one when it is uniform, a few spanning it when it is not
	// (see colors.go). A finish paints none and has no key.
	Colors map[string][][3]float64 `json:"colors"`
	// FinishBasis is where the finishes' recolouring renders are (basis.go).
	FinishBasis *FinishBasis `json:"finishBasis,omitempty"`
	// Large is where each entry's large image is (large.go).
	Large *LargeImages `json:"large,omitempty"`
	// Adjust is what the panel can open up on a preset, keyed by
	// package.symbol: its constructors' annotated parameters and the body as
	// a template (adjust.go). A preset with nothing adjustable has no key.
	Adjust map[string]*Adjustable `json:"adjust,omitempty"`
	// CoatBasis is where the recolouring renders of the presets with one
	// colour parameter are (coat.go), laid out as the finishes' are.
	CoatBasis *FinishBasis `json:"coatBasis,omitempty"`
	// Operators are the @surface presets the panel wraps a surface with, keyed
	// by package.symbol (operators.go).
	Operators map[string]*Operator `json:"operators,omitempty"`
}

// sheetGeometry measures the composed sheet rather than assuming how it was
// laid out. ComposeSheet puts a gap between cells, and the size of that gap is
// a constant in another module: reading it back off the image means the
// manifest stays true if that constant ever changes, instead of the panel
// quietly drawing every swatch sliced between its neighbours.
func sheetGeometry(sheet []byte, cols, rows, cellW, cellH int) (gap, width, height int, err error) {
	img, err := png.DecodeConfig(bytes.NewReader(sheet))
	if err != nil {
		return 0, 0, 0, fmt.Errorf("decode sheet: %w", err)
	}
	width, height = img.Width, img.Height
	if cols > 1 {
		gap = (width - cols*cellW) / (cols - 1)
	}
	// Whatever the gap turns out to be, the geometry has to describe the image
	// we actually hold, or a swatch is drawn from the wrong pixels.
	if want := cols*cellW + (cols-1)*gap; want != width {
		return 0, 0, 0, fmt.Errorf("sheet is %dpx wide; %d cells of %dpx with a %dpx gap make %d", width, cols, cellW, gap, want)
	}
	if want := rows*cellH + (rows-1)*gap; want != height {
		return 0, 0, 0, fmt.Errorf("sheet is %dpx tall; %d rows of %dpx with a %dpx gap make %d", height, rows, cellH, gap, want)
	}
	return gap, width, height, nil
}

func main() {
	outDir := flag.String("out", "examples/viewer/materials", "directory for the sheet and the manifest")
	width := flag.Int("width", 96, "swatch width in pixels")
	height := flag.Int("height", 96, "swatch height in pixels")
	cols := flag.Int("cols", 12, "swatches across the sheet")
	extent := flag.Float64("extent", 1.8, "scene half-size, in sphere radii")
	env := flag.String("env", "Softbox", "environment light preset (the viewers' Light row)")
	light := flag.Bool("light", true, "render on a light studio ground; -light=false uses the viewer's dark canvas")
	groundGrey := flag.Float64("ground-grey", defaultGroundGrey, "albedo of the plane below the sphere (see doc.go)")
	checker := flag.Float64("checker", defaultChecker, "checkerboard cell size on the ground, in sphere radii; 0 is a flat grey")
	noGround := flag.Bool("no-ground", false, "drop the plane below the sphere (see doc.go for why it is there)")
	large := flag.Int("large", 240, "size in pixels of each entry's large image, the grid's hover preview; 0 writes none")
	only := flag.String("only", "", "render only entries whose package.symbol contains one of these comma-separated substrings, for a sample")
	list := flag.Bool("list", false, "print the catalog and render nothing")
	manifestOnly := flag.Bool("manifest-only", false, "rewrite catalog.json from the libraries against the committed sheet, rendering nothing; refuses if the surfaces no longer match the sheet")
	thumbsOnly := flag.Bool("thumbs-only", false, "render only the operators' thumbnails (thumbs.go)")
	coatBasis := flag.Bool("coat-basis", false, "with -manifest-only, render the coat basis as well (coat.go); the sheet and the finishes stay as committed")
	flag.Parse()

	settings := swatchSettings{extent: *extent, env: *env, light: *light, ground: !*noGround, groundGrey: *groundGrey, checker: *checker}
	entries := ReadCatalog()
	adjust, err := ReadAdjust()
	if err != nil {
		fmt.Fprintln(os.Stderr, "adjust:", err)
		os.Exit(1)
	}
	operators, err := ReadOperators()
	if err != nil {
		fmt.Fprintln(os.Stderr, "operators:", err)
		os.Exit(1)
	}
	if *thumbsOnly {
		if err := writeThumbs(*outDir, operators, *extent, *env, *light); err != nil {
			fmt.Fprintln(os.Stderr, "thumbs:", err)
			os.Exit(1)
		}
		return
	}
	if *manifestOnly {
		var coat *swatchSettings
		if *coatBasis {
			coat = &settings
		}
		if err := rewriteManifest(*outDir, entries, adjust, operators, coat); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}
	if *only != "" {
		var kept []Entry
		for _, e := range entries {
			for _, want := range strings.Split(*only, ",") {
				if strings.Contains(e.Package+"."+e.Symbol, strings.TrimSpace(want)) {
					kept = append(kept, e)
					break
				}
			}
		}
		entries = kept
		// An entry's index is its cell in the sheet, so a sampled run has to
		// renumber: the sheet it writes holds only these, and a manifest
		// carrying the full catalog's positions would draw every swatch as
		// some other material.
		for i := range entries {
			entries[i].Index = i
		}
	}
	if len(entries) == 0 {
		fmt.Fprintln(os.Stderr, "no annotated surfaces found")
		os.Exit(1)
	}
	if *list {
		for _, e := range entries {
			fmt.Printf("%3d  %-8s %-28s %-24s %s\n", e.Index, e.Kind, e.Call(), e.Section, e.Doc)
		}
		fmt.Fprintf(os.Stderr, "%d entries\n", len(entries))
		return
	}

	panels := make([][]byte, 0, len(entries))
	colors := map[string][][3]float64{}
	started := time.Now()
	for i, e := range entries {
		png, err := renderSwatch(e, *width, *height, *extent, *env, *light, !*noGround, *groundGrey, *checker)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", e.Call(), err)
			os.Exit(1)
		}
		panels = append(panels, png)
		if e.Kind == "material" {
			if colors[e.Package+"."+e.Symbol], err = surfaceColors(e); err != nil {
				fmt.Fprintf(os.Stderr, "%s: %v\n", e.Call(), err)
				os.Exit(1)
			}
		}
		fmt.Fprintf(os.Stderr, "\r%d/%d %-32s", i+1, len(entries), e.Package+"."+e.Symbol)
	}
	fmt.Fprintf(os.Stderr, "\rrendered %d swatches in %s\n", len(entries), time.Since(started).Round(time.Millisecond))

	basis := &FinishBasis{File: "finish-basis.bin", Width: *width, Height: *height}
	var basisBytes []byte
	for _, e := range entries {
		if e.Kind != "finish" {
			continue
		}
		var err error
		if basisBytes, err = renderBasis(basisBytes, e, *width, *height, *extent, *env, *light, !*noGround, *groundGrey, *checker); err != nil {
			fmt.Fprintf(os.Stderr, "%s basis: %v\n", e.Call(), err)
			os.Exit(1)
		}
		basis.Order = append(basis.Order, e.Package+"."+e.Symbol)
	}

	sheet, err := eval.ComposeSheet(panels, *cols)
	if err != nil {
		fmt.Fprintln(os.Stderr, "compose:", err)
		os.Exit(1)
	}
	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	sheetPath := filepath.Join(*outDir, "swatches.png")
	if err := os.WriteFile(sheetPath, sheet, 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if *only == "" {
		if err := writeThumbs(*outDir, operators, *extent, *env, *light); err != nil {
			fmt.Fprintln(os.Stderr, "thumbs:", err)
			os.Exit(1)
		}
	}
	var largeImages *LargeImages
	if *large > 0 {
		var err error
		if largeImages, err = writeLarge(*outDir, entries, *large, *extent, *env, *light, !*noGround, *groundGrey, *checker); err != nil {
			fmt.Fprintln(os.Stderr, "large:", err)
			os.Exit(1)
		}
	}
	if len(basis.Order) > 0 {
		if err := os.WriteFile(filepath.Join(*outDir, basis.File), basisBytes, 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	} else {
		basis = nil
	}
	coat, coatBytes, err := buildCoatBasis(entries, adjust, *width, *height, settings)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if coat != nil {
		if err := os.WriteFile(filepath.Join(*outDir, coat.File), coatBytes, 0o644); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	// Just the file name: the panel resolves it against this manifest's own
	// module URL, so the pair moves together and neither has to know where the
	// app mounts the viewer.
	rows := (len(entries) + *cols - 1) / *cols
	gap, sheetW, sheetH, err := sheetGeometry(sheet, *cols, rows, *width, *height)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	manifest := Manifest{
		Image:       "swatches.png",
		CellWidth:   *width,
		CellHeight:  *height,
		Cols:        *cols,
		Rows:        rows,
		Gap:         gap,
		SheetWidth:  sheetW,
		SheetHeight: sheetH,
		Entries:     entries,
		Colors:      colors,
		FinishBasis: basis,
		Large:       largeImages,
		Adjust:      adjust,
		CoatBasis:   coat,
		Operators:   operators,
	}
	manifestPath, err := writeManifest(*outDir, manifest)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "wrote %s (%d KB) and %s\n", sheetPath, len(sheet)/1024, manifestPath)
}

func writeManifest(outDir string, manifest Manifest) (string, error) {
	body, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return "", err
	}
	manifestPath := filepath.Join(outDir, "catalog.json")
	if err := os.WriteFile(manifestPath, append(body, '\n'), 0o644); err != nil {
		return "", err
	}
	return manifestPath, nil
}

// rewriteManifest refreshes the committed manifest's catalog, adjust records
// and operators without rendering: the sheet, the colours, the finish basis
// and the large images are kept as they are. That is only sound while the
// surfaces are the ones the sheet holds, in the same order, so a library that
// has gained or reordered a surface is refused and needs a full render. With
// coat set, the coat basis is rendered afresh at the sheet's cell size, which
// is cheap: two renders per preset with one colour.
func rewriteManifest(outDir string, entries []Entry, adjust map[string]*Adjustable, operators map[string]*Operator, coat *swatchSettings) error {
	raw, err := os.ReadFile(filepath.Join(outDir, "catalog.json"))
	if err != nil {
		return fmt.Errorf("read the committed manifest: %w", err)
	}
	var manifest Manifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		return fmt.Errorf("committed manifest: %w", err)
	}
	if len(entries) != len(manifest.Entries) {
		return fmt.Errorf("the libraries declare %d surfaces, the sheet holds %d: render in full", len(entries), len(manifest.Entries))
	}
	for i, e := range entries {
		if have := manifest.Entries[i]; have.Kind != e.Kind || have.Package != e.Package || have.Symbol != e.Symbol {
			return fmt.Errorf("surface %d is %s in the sheet and %s in the libraries: render in full", i, have.Call(), e.Call())
		}
	}
	manifest.Entries = entries
	manifest.Adjust = adjust
	manifest.Operators = operators
	if coat != nil {
		basis, raw, err := buildCoatBasis(entries, adjust, manifest.CellWidth, manifest.CellHeight, *coat)
		if err != nil {
			return err
		}
		if basis != nil {
			if err := os.WriteFile(filepath.Join(outDir, basis.File), raw, 0o644); err != nil {
				return err
			}
		}
		manifest.CoatBasis = basis
	}
	manifestPath, err := writeManifest(outDir, manifest)
	if err != nil {
		return err
	}
	coats := 0
	if manifest.CoatBasis != nil {
		coats = len(manifest.CoatBasis.Order)
	}
	fmt.Fprintf(os.Stderr, "rewrote %s: %d surfaces, %d adjustable, %d with a coat basis, %d operators\n", manifestPath, len(entries), len(adjust), coats, len(operators))
	return nil
}

// renderSwatch assembles one swatch document and draws it. The libraries
// resolve from the embedded copy, so there is no source root to pass. extent
// and checker are in sphere radii (see doc.go).
func renderSwatch(e Entry, width, height int, extent float64, env string, light, ground bool, groundGrey, checker float64) ([]byte, error) {
	res, err := assembler.Assemble([]string{swatchDoc(e, ground, groundGrey, checker)})
	if err != nil {
		return nil, fmt.Errorf("assemble: %w", err)
	}
	return eval.RenderLitSizedView(res.Program, width, height, extent*swatchRadius, 0, eval.LitView{
		Light:        light,
		NoAxes:       true,
		Environment:  env,
		GlassSamples: swatchGlassSamples,
	})
}
