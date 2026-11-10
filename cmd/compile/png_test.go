package main_test

import (
	"bytes"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	main "gradientcontrol.com/omega/core/cmd/compile"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

const pngBox = "@scene(60)\nplate = boxCentered([0, 0, 2.5], [40, 30, 5], intersectRound(1));\nreturn plate;\n"

// decoded reads a PNG and counts its distinct colours: a picture of a part has many, an empty canvas few.
func decoded(t *testing.T, data []byte) (image.Rectangle, int) {
	t.Helper()
	img, err := png.Decode(bytes.NewReader(data))
	require.NoError(t, err)
	b := img.Bounds()
	colours := map[[4]uint32]bool{}
	for y := b.Min.Y; y < b.Max.Y; y += 4 {
		for x := b.Min.X; x < b.Max.X; x += 4 {
			r, g, bl, a := img.At(x, y).RGBA()
			colours[[4]uint32{r >> 8, g >> 8, bl >> 8, a >> 8}] = true
		}
	}
	return b, len(colours)
}

// One view is one panel of the size asked; the field is drawn, so the image is not a blank canvas.
func TestRun_PNG_OneView(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "plate.gsl")
	require.NoError(t, os.WriteFile(source, []byte(pngBox), 0644))
	out := filepath.Join(dir, "plate.png")
	var stdout, stderr bytes.Buffer
	require.NoError(t, main.Run([]string{"-t", "png", "--size", "160x120", "--light", "--no-meta", "-o", out, source},
		nil, &stdout, &stderr))
	data, err := os.ReadFile(out)
	require.NoError(t, err)
	b, colours := decoded(t, data)
	assert.Equal(t, 160, b.Dx())
	assert.Equal(t, 120, b.Dy())
	assert.True(t, colours > 20)
}

// Several views tile two across into one sheet, a direction among the presets.
func TestRun_PNG_ViewsTileTwoAcross(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "plate.gsl")
	require.NoError(t, os.WriteFile(source, []byte(pngBox), 0644))
	var stdout, stderr bytes.Buffer
	require.NoError(t, main.Run([]string{"-t", "png", "--views", "iso,top,-1,1,0.6", "--size", "100x80", "--no-meta",
		"-o", "-", source}, nil, &stdout, &stderr))
	b, _ := decoded(t, stdout.Bytes())
	assert.Equal(t, 2*100+4, b.Dx()) // ComposeSheet leaves 4 px between panels
	assert.Equal(t, 2*80+4, b.Dy())
}

// The default output is the source's name with .png.
func TestRun_PNG_DefaultOutputPath(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "plate.gsl")
	require.NoError(t, os.WriteFile(source, []byte(pngBox), 0644))
	var stdout, stderr bytes.Buffer
	require.NoError(t, main.Run([]string{"-t", "png", "--size", "64x48", "--no-meta", source}, nil, &stdout, &stderr))
	_, err := os.Stat(filepath.Join(dir, "plate.png"))
	require.NoError(t, err)
}

func TestRun_PNG_Refusals(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "plate.gsl")
	require.NoError(t, os.WriteFile(source, []byte(pngBox), 0644))
	for _, c := range []struct {
		args []string
		says string
	}{
		{[]string{"--size", "640"}, "invalid --size"},
		{[]string{"--size", "0x480"}, "invalid --size"},
		{[]string{"--extent", "-1"}, "invalid --extent"},
		{[]string{"--views", "sideways"}, "--views"},
	} {
		var stdout, stderr bytes.Buffer
		args := append([]string{"-t", "png", "--no-meta", "-o", filepath.Join(dir, "x.png")}, c.args...)
		err := main.Run(append(args, source), nil, &stdout, &stderr)
		require.True(t, err != nil)
		assert.True(t, strings.Contains(err.Error(), c.says))
	}
}
