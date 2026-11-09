package cli_test

import (
	"bytes"
	"context"
	"testing"

	"gradientcontrol.com/omega/core/cmd/gsl/internal/cli"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
	"gradientcontrol.com/omega/core/internal/version"
)

// gsl --version (and gsl version) says the release in three numbers, without looking for any tool binary.
func TestRun_Version(t *testing.T) {
	for _, arg := range []string{"--version", "-version", "version"} {
		var stdout, stderr bytes.Buffer
		require.NoError(t, cli.Run(context.Background(), t.TempDir(), []string{arg}, nil, &stdout, &stderr))
		assert.Equal(t, "gsl "+version.Version+"\n", stdout.String())
	}
	assert.True(t, len(version.Version) >= 5)
}
