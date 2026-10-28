package lib

import (
	"testing"

	assembler "gradientcontrol.com/omega/core/internal/assembler"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/require"
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
