package mesh_test

import (
	"testing"

	assembler "gradientcontrol.com/omega/core/internal/assembler"
	"gradientcontrol.com/omega/core/internal/fidget/internal/mesh"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

func TestTapeReduceInvariance(t *testing.T) {
	src := `
r = reduce(0.0, (acc, x) {
  return acc + x;
}, [1.0, 2.0, 3.0, 4.0]);

s = sphere(10.0 + r);
return s;
`
	res, err := assembler.Assemble([]string{src}, &source.MockManager{})
	require.NoError(t, err)

	tape1, names1, err := mesh.GenerateTape(res.Program)
	require.NoError(t, err)
	assert.True(t, len(tape1) > 0)
	assert.Equal(t, 1, len(names1))

	// Re-generating tape must be deterministic and byte-for-byte identical.
	tape2, names2, err := mesh.GenerateTape(res.Program)
	require.NoError(t, err)
	assert.Equal(t, tape1, tape2)
	assert.Equal(t, names1, names2)
}

func TestTapeReduceCellScanInvariance(t *testing.T) {
	src := `
offsets = [[-1.0, 0.0], [0.0, 0.0], [1.0, 0.0]];
d = reduce(100.0, (acc, o) {
  dist = length(p.xy - o);
  return min(acc, dist);
}, offsets);

s = unionSharp(sphere(5.0), d - 2.0);
return s;
`
	res, err := assembler.Assemble([]string{src}, &source.MockManager{})
	require.NoError(t, err)

	tape, names, err := mesh.GenerateTape(res.Program)
	require.NoError(t, err)
	assert.True(t, len(tape) > 0)
	assert.Equal(t, 1, len(names))
}
