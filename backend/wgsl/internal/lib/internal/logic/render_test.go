package logic

import (
	"strings"
	"testing"

	assembler "gradientcontrol.com/omega/core/internal/assembler"
	"gradientcontrol.com/omega/core/internal/assembler/variation"
	"gradientcontrol.com/omega/core/internal/slider"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/require"
	"gradientcontrol.com/omega/core/internal/ui"
	"gradientcontrol.com/omega/core/program"
)

func TestRender(t *testing.T) {
	src := "return length(p) - 1.0;"
	res, err := assembler.Assemble([]string{src}, &source.MockManager{})
	require.NoError(t, err)

	rendered, err := Render(src, res.Program)
	require.NoError(t, err)
	require.NotNil(t, rendered)
	require.NotEmpty(t, rendered.WGSL)
}

func TestRender_EmptyProgram(t *testing.T) {
	prog := &program.Program{Primary: &program.Package{}}
	rendered, err := Render("", prog)
	require.NoError(t, err)
	require.NotNil(t, rendered)
	require.Equal(t, 0, rendered.ShapeCount)
	require.Empty(t, rendered.ShapeNames)
	require.Empty(t, rendered.ShapeLocations)
	require.Contains(t, rendered.WGSL, "fn sdShape(p: vec3<f32>, time: f32) -> gsl_shape")
	require.Contains(t, rendered.WGSL, "return gsl_shape(1e10, 0u);")
}

func TestRender_ZeroReturnsWithSlidersAndUIControls(t *testing.T) {
	src := `importId = importStep "part.step";
radius = 5;
radius.min = 1;
radius.max = 10;
center = uiPoint([1, 2, 3]);
xf = uiTransform([[1, 0, 0, 0], [0, 1, 0, 0], [0, 0, 1, 0], [4, 5, 6, 1]]);
cavity = length(p - center) - radius + noiseHash3(p);
return meshDifference(mesh(importId, xf), cavity);`
	res, err := assembler.Assemble([]string{src}, &source.MockManager{})
	require.NoError(t, err)

	sliced := variation.Program(res.Program)
	rendered, err := Render(src, sliced)
	require.NoError(t, err)
	require.NotNil(t, rendered)
	require.Equal(t, 0, rendered.ShapeCount)
	require.Empty(t, rendered.ShapeNames)
	require.Empty(t, rendered.ShapeLocations)
	require.Empty(t, rendered.MeshSdfSlots)
	sliders := slider.Extract(src, sliced.Primary)
	require.Len(t, sliders, 1)
	require.Contains(t, rendered.WGSL, sliders[0].ParamName+": f32")
	controls := ui.Extract(sliced.Primary, sliced.GetImport)
	require.Len(t, controls, 2)
	for _, c := range controls {
		for _, p := range c.Params {
			require.Contains(t, rendered.WGSL, p.ParamName+": f32")
		}
	}
	require.Contains(t, rendered.WGSL, "return gsl_shape(1e10, 0u);")
	require.False(t, strings.Contains(rendered.WGSL, "noiseTex"))
}
