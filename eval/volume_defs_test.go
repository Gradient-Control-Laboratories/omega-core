//go:build cgo

package eval_test

import (
	"math"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/eval"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/source"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/assert"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/require"
)

// volumeBallDoc is a ball whose radius is a slider.
const volumeBallDoc = "@scene(40)\n@range(2, 12) r = 5.0;\nball = length(p) - r;\nreturn ball;\n"

// VolumeOf brackets the document with defs applied: doubling the radius brackets eight times the volume, and the
// bracket of the document as written does not reach it.
func TestVolumeOfAppliesDefs(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	region := [6]float64{-12, -12, -12, 12, 12, 12}
	written := 4 * math.Pi * 125 / 3
	doubled := 8 * written

	got, err := e.VolumeOf(t.Context(), nil, volumeBallDoc, "r = 10.0;", "ball", region, 0.02)
	require.NoError(t, err)
	assert.True(t, got.Lower <= doubled && doubled <= got.Upper, "%+v brackets %g", got, doubled)
	assert.True(t, got.Upper-got.Lower <= doubled*0.05, "%+v is close about %g", got, doubled)

	plain, err := e.VolumeOf(t.Context(), nil, volumeBallDoc, "", "ball", region, 0.02)
	require.NoError(t, err)
	assert.True(t, plain.Lower <= written && written <= plain.Upper, "%+v brackets %g", plain, written)
	assert.True(t, plain.Upper < doubled, "%+v stays below %g without the def", plain, doubled)
}

// Volume is VolumeOf with no defs: the same bracket for the same document.
func TestVolumeIsVolumeOfWithoutDefs(t *testing.T) {
	e := eval.New(&source.MockManager{}, nil)
	region := [6]float64{-12, -12, -12, 12, 12, 12}
	a, err := e.Volume(t.Context(), volumeBallDoc, "ball", region, 0.05)
	require.NoError(t, err)
	b, err := e.VolumeOf(t.Context(), nil, volumeBallDoc, "", "ball", region, 0.05)
	require.NoError(t, err)
	assert.Equal(t, a, b)
}
