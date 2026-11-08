package lib

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
)

// ComposeSheet tiles PNG panels into one PNG, cols across, in the order
// given, on a background matched to the first panel's top-left pixel so a
// light or dark render keeps its ground. Every panel is placed at its own
// size; a short last row is left blank. There is no lettering — a caller
// that attaches a sheet names its panels in the text beside it.
func ComposeSheet(panels [][]byte, cols int) ([]byte, error) {
	if len(panels) == 0 {
		return nil, fmt.Errorf("no panels")
	}
	if cols <= 0 {
		cols = len(panels)
	}
	imgs := make([]image.Image, len(panels))
	cellW, cellH := 0, 0
	for i, raw := range panels {
		img, err := png.Decode(bytes.NewReader(raw))
		if err != nil {
			return nil, fmt.Errorf("panel %d: %w", i, err)
		}
		imgs[i] = img
		if w := img.Bounds().Dx(); w > cellW {
			cellW = w
		}
		if h := img.Bounds().Dy(); h > cellH {
			cellH = h
		}
	}
	rows := (len(panels) + cols - 1) / cols
	const gap = 4
	sheet := image.NewRGBA(image.Rect(0, 0, cols*cellW+(cols-1)*gap, rows*cellH+(rows-1)*gap))
	bg := imgs[0].At(imgs[0].Bounds().Min.X, imgs[0].Bounds().Min.Y)
	r, g, b, _ := bg.RGBA()
	draw.Draw(sheet, sheet.Bounds(), &image.Uniform{color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), 255}}, image.Point{}, draw.Src)
	for i, img := range imgs {
		x := (i % cols) * (cellW + gap)
		y := (i / cols) * (cellH + gap)
		dst := image.Rect(x, y, x+img.Bounds().Dx(), y+img.Bounds().Dy())
		draw.Draw(sheet, dst, img, img.Bounds().Min, draw.Src)
	}
	var out bytes.Buffer
	if err := png.Encode(&out, sheet); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}
