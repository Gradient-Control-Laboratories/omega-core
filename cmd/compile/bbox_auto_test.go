package main_test

import (
	"bytes"
	"encoding/binary"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	main "gradientcontrol.com/omega/core/cmd/compile"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
	"gradientcontrol.com/omega/core/internal/version"
)

// stlBounds reads a binary STL's triangle count and the box its vertices span.
func stlBounds(t *testing.T, data []byte) (int, [6]float64) {
	t.Helper()
	require.True(t, len(data) >= 84)
	n := int(binary.LittleEndian.Uint32(data[80:84]))
	box := [6]float64{math.Inf(1), math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1), math.Inf(-1)}
	for i := 0; i < n; i++ {
		for v := 0; v < 3; v++ {
			for k := 0; k < 3; k++ {
				off := 84 + 50*i + 12 + 12*v + 4*k
				x := float64(math.Float32frombits(binary.LittleEndian.Uint32(data[off:])))
				box[k] = math.Min(box[k], x)
				box[k+3] = math.Max(box[k+3], x)
			}
		}
	}
	return n, box
}

// A part reaching past its @scene box is meshed whole with --bbox auto, and clipped by the @scene box without it.
func TestRun_STL_BboxAuto(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "plate.gsl")
	require.NoError(t, os.WriteFile(source, []byte("@scene(60)\nplate = box([0, 0, 0], [40, 30, 5]);\nreturn plate;\n"), 0644))

	out := filepath.Join(dir, "auto.stl")
	var stdout, stderr bytes.Buffer
	require.NoError(t, main.Run([]string{"-t", "stl", "--algo", "mdc", "--subdivision", "6", "--bbox", "auto", "--no-meta",
		"-o", out, source}, nil, &stdout, &stderr))
	data, err := os.ReadFile(out)
	require.NoError(t, err)
	n, box := stlBounds(t, data)
	assert.True(t, n > 0)
	want := [6]float64{0, 0, 0, 40, 30, 5}
	for k := range want {
		assert.True(t, math.Abs(box[k]-want[k]) < 1e-3)
	}

	clipped := filepath.Join(dir, "default.stl")
	require.NoError(t, main.Run([]string{"-t", "stl", "--algo", "mdc", "--subdivision", "6", "--no-meta", "-o", clipped,
		source}, nil, &stdout, &stderr))
	data, err = os.ReadFile(clipped)
	require.NoError(t, err)
	_, box = stlBounds(t, data)
	assert.True(t, box[3] <= 30.001) // @scene(60)'s box, -30..30, cuts it at x = 30
}

func TestRun_STL_BboxAutoNothingInside(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "empty.gsl")
	require.NoError(t, os.WriteFile(source, []byte("return 1.0;\n"), 0644))
	var stdout, stderr bytes.Buffer
	err := main.Run([]string{"-t", "stl", "--bbox", "auto", "--no-meta", "-o", filepath.Join(dir, "x.stl"), source},
		nil, &stdout, &stderr)
	require.True(t, err != nil)
	assert.True(t, strings.Contains(err.Error(), "--bbox auto: the field is nowhere inside"))
}

func TestRun_Version(t *testing.T) {
	var stdout, stderr bytes.Buffer
	require.NoError(t, main.Run([]string{"--version"}, nil, &stdout, &stderr))
	assert.Equal(t, "compile "+version.Version+"\n", stdout.String())
}
