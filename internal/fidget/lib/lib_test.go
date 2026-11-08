package lib

import (
	"testing"

	assembler "gradientcontrol.com/omega/core/internal/assembler"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/require"
)

func TestMesh(t *testing.T) {
	res, err := assembler.Assemble([]string{"return length(p) - 1.0;"}, &source.MockManager{})
	require.NoError(t, err)

	bbox := [6]float32{-2, -2, -2, 2, 2, 2}
	stl, err := Mesh(res.Program, 4, bbox, "marching-cubes", func(completed, total uint32, status string) {})
	require.NoError(t, err)
	require.NotEmpty(t, stl)
}
