package main_test

import (
	"bytes"
	"context"
	"testing"

	main "gradientcontrol.com/omega/core/cmd/gsl"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

func TestMainRun_Help(t *testing.T) {
	var stdout, stderr bytes.Buffer
	err := main.Run(context.Background(), []string{"help"}, nil, &stdout, &stderr)
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Usage: gsl <command> [arguments]")
}
