package eval

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"
)

func TestParseView(t *testing.T) {
	v, err := ParseView("iso-below")
	if err != nil {
		t.Fatal(err)
	}
	if v.Dir != [3]float64{1, 1, -1} || v.Up != [3]float64{0, 0, -1} {
		t.Errorf("iso-below = %+v", v)
	}
	v, err = ParseView(" 0, 2, 0.5 ")
	if err != nil {
		t.Fatal(err)
	}
	if v.Dir != [3]float64{0, 2, 0.5} || v.Up != [3]float64{} {
		t.Errorf("direction view = %+v", v)
	}
	for _, bad := range []string{"", "sideways", "1,2", "0,0,0"} {
		if _, err := ParseView(bad); err == nil {
			t.Errorf("ParseView(%q) accepted", bad)
		}
	}
}

func solidPNG(t *testing.T, w, h int, c color.RGBA) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, c)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestComposeSheet(t *testing.T) {
	red := solidPNG(t, 30, 20, color.RGBA{200, 0, 0, 255})
	blue := solidPNG(t, 30, 20, color.RGBA{0, 0, 200, 255})
	raw, err := ComposeSheet([][]byte{red, blue, red}, 2)
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	// Two columns, two rows, a 4-pixel gap between cells.
	if got, want := img.Bounds().Dx(), 30*2+4; got != want {
		t.Errorf("width %d, want %d", got, want)
	}
	if got, want := img.Bounds().Dy(), 20*2+4; got != want {
		t.Errorf("height %d, want %d", got, want)
	}
	at := func(x, y int) color.RGBA {
		r, g, b, a := img.At(x, y).RGBA()
		return color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
	}
	if c := at(5, 5); c.R != 200 || c.B != 0 {
		t.Errorf("panel 0 pixel = %+v, want red", c)
	}
	if c := at(34+5, 5); c.B != 200 || c.R != 0 {
		t.Errorf("panel 1 pixel = %+v, want blue", c)
	}
	if c := at(5, 24+5); c.R != 200 {
		t.Errorf("panel 2 pixel = %+v, want red", c)
	}
	// The empty last cell carries the ground colour of the first panel.
	if c := at(34+5, 24+5); c.R != 200 || c.B != 0 {
		t.Errorf("blank cell = %+v, want the first panel's ground", c)
	}
	if _, err := ComposeSheet(nil, 2); err == nil {
		t.Error("empty sheet accepted")
	}
}
