package main

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// The large images: each entry drawn again at the size the grid's hover
// preview shows it, three times the swatch, so a material can be looked at
// rather than recognised. One PNG per entry, fetched only when hovered, so the
// grid's load is the sheet alone. A finish also gets its two recolouring
// renders at this size (basis.go), gzipped — most of a render is the ground,
// where the radiance is zero — so its preview can be drawn in the body's
// colour too.

// LargeImages is the manifest's description of them. An entry's image is
// <Dir>/<package>.<symbol>.png, and a finish's renders are
// <Dir>/<package>.<symbol>.basis.gz, laid out as the sheet's basis file is.
type LargeImages struct {
	Dir    string `json:"dir"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}

func largeName(e Entry) string { return e.Package + "." + e.Symbol }

// writeLarge renders every entry's large image, and every finish's large
// basis, into outDir/large.
func writeLarge(outDir string, entries []Entry, size int, extent float64, env string, light, ground bool, groundGrey, checker float64) (*LargeImages, error) {
	dir := filepath.Join(outDir, "large")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, err
	}
	started := time.Now()
	for i, e := range entries {
		png, err := renderSwatch(e, size, size, extent, env, light, ground, groundGrey, checker)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Call(), err)
		}
		if err := os.WriteFile(filepath.Join(dir, largeName(e)+".png"), png, 0o644); err != nil {
			return nil, err
		}
		if e.Kind == "finish" {
			raw, err := renderBasis(nil, e, size, size, extent, env, light, ground, groundGrey, checker)
			if err != nil {
				return nil, fmt.Errorf("%s basis: %w", e.Call(), err)
			}
			var gz bytes.Buffer
			w, _ := gzip.NewWriterLevel(&gz, gzip.BestCompression)
			if _, err := w.Write(raw); err != nil {
				return nil, err
			}
			if err := w.Close(); err != nil {
				return nil, err
			}
			if err := os.WriteFile(filepath.Join(dir, largeName(e)+".basis.gz"), gz.Bytes(), 0o644); err != nil {
				return nil, err
			}
		}
		fmt.Fprintf(os.Stderr, "\rlarge %d/%d %-32s", i+1, len(entries), largeName(e))
	}
	fmt.Fprintf(os.Stderr, "\rrendered %d large images in %s\n", len(entries), time.Since(started).Round(time.Second))
	return &LargeImages{Dir: "large", Width: size, Height: size}, nil
}
