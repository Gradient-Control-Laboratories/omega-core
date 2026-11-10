package main_test

import (
	"bytes"
	"math"
	"os"
	"path/filepath"
	"testing"

	main "gradientcontrol.com/omega/core/cmd/compile"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// A thin slab tilted 65 degrees -- a phone stand's back -- is thinner than the coarse grid's 32 mm cells: two of the
// grid's samples land inside it, and the grid fit alone, which only looks within a cell of those, stops at z = 64 of
// its 96. --bbox auto proves each face of its box by interval arithmetic and moves it out until nothing lies beyond,
// so the slab is meshed whole.
func TestRun_STL_BboxAutoKeepsAThinTiltedSlabWhole(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "slab.gsl")
	require.NoError(t, os.WriteFile(source, []byte(
		"a = 65.0 * 3.141592653589793 / 180.0;\n"+
			"v = (p.y + 22) * cos(a) + (p.z - 10) * sin(a);\n"+
			"s = -(p.y + 22) * sin(a) + (p.z - 10) * cos(a);\n"+
			"slab = max(abs(v - 45) - 50, abs(s + 2) - 2, abs(p.x) - 38);\n"+
			"return slab;\n"), 0644))
	out := filepath.Join(dir, "slab.stl")
	var stdout, stderr bytes.Buffer
	require.NoError(t, main.Run([]string{"-t", "stl", "--algo", "mdc", "--subdivision", "7", "--bbox", "auto", "--no-meta",
		"-o", out, source}, nil, &stdout, &stderr))
	data, err := os.ReadFile(out)
	require.NoError(t, err)
	n, box := stlBounds(t, data)
	assert.True(t, n > 0)
	// The slab runs from 5 mm below (y, z) = (-22, 10) to 95 mm above it along the slope, 4 mm thick behind it.
	sa, ca := math.Sin(65*math.Pi/180), math.Cos(65*math.Pi/180)
	want := [6]float64{-38, -22 - 5*ca, 10 - 5*sa - 4*ca, 38, -22 + 95*ca + 4*sa, 10 + 95*sa}
	for k := range want {
		assert.True(t, math.Abs(box[k]-want[k]) < 0.5)
	}
}
