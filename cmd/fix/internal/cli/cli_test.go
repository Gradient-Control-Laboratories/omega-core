package cli_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"gradientcontrol.com/omega/core/cmd/fix/internal/cli"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

func TestCLI_ModesAndDiff(t *testing.T) {
	var stdout, stderr bytes.Buffer
	require.NoError(t, cli.Run(nil, nil, &stdout, &stderr))
	assert.Contains(t, stdout.String(), "Usage of fix:")

	stdout.Reset()
	require.NoError(t, cli.Run([]string{"help"}, nil, &stdout, &stderr))
	assert.Contains(t, stdout.String(), "Registered fixers:")

	stdout.Reset()
	require.NoError(t, cli.Run([]string{"help", "scene"}, nil, &stdout, &stderr))
	assert.Contains(t, stdout.String(), "@scene")

	err := cli.Run([]string{"help", "nonexistent"}, nil, &stdout, &stderr)
	require.Error(t, err)

	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "test.gsl")
	require.NoError(t, os.WriteFile(f, []byte("// Header\n// Scene Size: 30\nreturn sphere(1);\n"), 0644))

	stdout.Reset()
	err = cli.Run([]string{"-diff", f}, nil, &stdout, &stderr)
	assert.True(t, errors.Is(err, cli.ErrDiffFound))
	assert.Contains(t, stdout.String(), "+ @scene(30)")

	stdout.Reset()
	require.NoError(t, cli.Run([]string{tmpDir}, nil, &stdout, &stderr))
	b, err := os.ReadFile(f)
	require.NoError(t, err)
	assert.Equal(t, "// Header\n@scene(30)\nreturn sphere(1);\n", string(b))
}
