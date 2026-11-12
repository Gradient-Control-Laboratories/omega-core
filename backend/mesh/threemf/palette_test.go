package threemf

import (
	"testing"

	"gradientcontrol.com/omega/core/internal/test/assert"
)

// The colors examples/viewer/panels/body-palette.js gives bodies 1-7 and 12
// (rgbToHex(paletteColor(i)), its first preset), so an export colors a body as
// the viewer draws it.
func TestPaletteColorMatchesViewer(t *testing.T) {
	for i, want := range map[int]string{
		1: "#8987ff", 2: "#bcf846", 3: "#e100e2", 4: "#49f7ba",
		5: "#ff8b8b", 6: "#49baf7", 7: "#e1e200", 12: "#f7bd4c",
	} {
		assert.Equal(t, want, PaletteColor(i), "body %d", i)
	}
	assert.Equal(t, PaletteColor(1), PaletteColor(0), "no index is body 1")
}
