package main_test

import (
	"bytes"
	"context"
	"testing"

	"gradientcontrol.com/omega/core/cmd/gsl/internal/cli"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

func TestMainRun_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := cli.Run(context.Background(), cli.ResolveExecutableDir(), []string{"help"}, nil, &stdout, &stderr)
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Usage: gsl <command> [arguments]")
}
