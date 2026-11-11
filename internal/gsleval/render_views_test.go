package gsleval_test

import (
	"bytes"
	"context"
	"image/png"
	"testing"

	"gradientcontrol.com/omega/core/internal/fidget/lib"
	"gradientcontrol.com/omega/core/internal/gsleval"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

const renderDoc = "@scene(40)\n@range(5, 15) r = 10.0;\nball = length(p) - r;\nreturn ball;\n"

func views(t *testing.T, s string) []lib.LitView {
	t.Helper()
	v, err := lib.ParseViews(s)
	require.NoError(t, err)
	return v
}

// Render from presets is RenderViews from the same cameras: one renderer behind both.
func TestRenderViewsDrawsWhatRenderDraws(t *testing.T) {
	e := gsleval.New(&source.MockManager{}, nil)
	ctx := context.Background()
	byName, err := e.Render(ctx, nil, renderDoc, "", "", []string{"iso", "top"}, 96, 72)
	require.NoError(t, err)
	byView, err := e.RenderViews(ctx, nil, renderDoc, "", "", views(t, "iso,top"), 96, 72, 0, 0)
	require.NoError(t, err)
	assert.True(t, bytes.Equal(byName, byView), "the same cameras draw the same picture")
}

// RenderViews takes what `compile -t png` takes: a direction, a light ground, an extent; and defs, as Render does.
func TestRenderViewsTakesDirectionsGroundExtentAndDefs(t *testing.T) {
	e := gsleval.New(&source.MockManager{}, nil)
	ctx := context.Background()
	draw := func(defs string, extent float64, light bool) []byte {
		t.Helper()
		v := views(t, "-1,1,0.6")
		v[0].Light = light
		data, err := e.RenderViews(ctx, nil, renderDoc, defs, "", v, 96, 72, extent, 0)
		require.NoError(t, err)
		img, err := png.Decode(bytes.NewReader(data))
		require.NoError(t, err)
		assert.Equal(t, 96, img.Bounds().Dx())
		return data
	}
	dark := draw("", 30, false)
	assert.False(t, bytes.Equal(dark, draw("", 30, true)), "the light ground changes the picture")
	assert.False(t, bytes.Equal(dark, draw("r = 5.0;", 30, false)), "at a fixed extent a smaller ball draws smaller")

	_, err := e.RenderViews(ctx, nil, renderDoc, "", "", nil, 0, 72, 0, 0)
	assert.Error(t, err)
	_, err = e.RenderViews(ctx, nil, renderDoc, "", "", nil, 96, 72, -1, 0)
	assert.Error(t, err)
}
