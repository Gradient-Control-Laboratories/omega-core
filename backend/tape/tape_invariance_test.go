package tape_test

import (
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/assembler"
	"github.com/Gradient-Control-Laboratories/omega-core/backend/tape"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/source"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/assert"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/require"
)

func TestTapeReduceInvariance(t *testing.T) {
	src := `
r = reduce(0.0, (acc, x) {
  return acc + x;
}, [1.0, 2.0, 3.0, 4.0]);

s = sphere(10.0 + r);
return s;
`
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(&source.MockManager{}))
	require.NoError(t, err)

	tape1, names1, _, err := tape.GenerateTapeWithSpinodoids(res.Program)
	require.NoError(t, err)
	assert.True(t, len(tape1) > 0)
	assert.Equal(t, 1, len(names1))

	// Re-generating tape must be deterministic and byte-for-byte identical.
	tape2, names2, _, err := tape.GenerateTapeWithSpinodoids(res.Program)
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
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(&source.MockManager{}))
	require.NoError(t, err)

	tapeBytes, names, _, err := tape.GenerateTapeWithSpinodoids(res.Program)
	require.NoError(t, err)
	assert.True(t, len(tapeBytes) > 0)
	assert.Equal(t, 1, len(names))
}
