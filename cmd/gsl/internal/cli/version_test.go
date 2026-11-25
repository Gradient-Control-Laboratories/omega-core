package cli_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/cmd/gsl/internal/cli"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/assert"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/require"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/version"
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
