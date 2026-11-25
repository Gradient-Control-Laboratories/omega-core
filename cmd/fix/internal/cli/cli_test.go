package cli_test

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/cmd/fix/internal/cli"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/assert"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/require"
)

func TestCLI_ModesAndDiff(t *testing.T) {
	var stdout bytes.Buffer
	require.NoError(t, cli.Run(nil, &stdout))
	assert.Contains(t, stdout.String(), "Usage of fix:")

	stdout.Reset()
	require.NoError(t, cli.Run([]string{"help"}, &stdout))
	assert.Contains(t, stdout.String(), "Registered fixers:")

	stdout.Reset()
	require.NoError(t, cli.Run([]string{"help", "scene"}, &stdout))
	assert.Contains(t, stdout.String(), "@scene")

	stdout.Reset()
	require.NoError(t, cli.Run([]string{"help", "omegaimport"}, &stdout))
	assert.Contains(t, stdout.String(), `import . "omega"`)

	err := cli.Run([]string{"help", "nonexistent"}, &stdout)
	require.Error(t, err)

	tmpDir := t.TempDir()
	f := filepath.Join(tmpDir, "test.gsl")
	require.NoError(t, os.WriteFile(f, []byte("// Header\n// Scene Size: 30\nimport . \"omega\";\nreturn sphere(1);\n"), 0644))

	stdout.Reset()
	err = cli.Run([]string{"-diff", f}, &stdout)
	assert.True(t, errors.Is(err, cli.ErrDiffFound))
	assert.Contains(t, stdout.String(), "+ @scene(30)")
	assert.Contains(t, stdout.String(), `- import . "omega";`)

	stdout.Reset()
	require.NoError(t, cli.Run([]string{tmpDir}, &stdout))
	b, err := os.ReadFile(f)
	require.NoError(t, err)
	assert.Equal(t, "// Header\n@scene(30)\nreturn sphere(1);\n", string(b))
}
