package cli_test

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/cmd/gsl/internal/cli"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

// `gsl export ARGS` is `compile export ARGS`: the export finds the viewer's files as `compile viewer` does.
func TestRun_ExportRunsCompile(t *testing.T) {
	binDir := t.TempDir()
	createMockTool(t, binDir, "compile", `echo "MOCK_COMPILE:$@"`)

	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"export", "part.gsl"}, "MOCK_COMPILE:export part.gsl"},
		{[]string{"export", "a.gsl", "b.gsl", "-o", "site"}, "MOCK_COMPILE:export a.gsl b.gsl -o site"},
		{[]string{"help", "export"}, "MOCK_COMPILE:export --help"},
	} {
		var stdout, stderr bytes.Buffer
		require.NoError(t, cli.Run(context.Background(), binDir, tc.args, nil, &stdout, &stderr))
		assert.Equal(t, tc.want, strings.TrimSpace(stdout.String()))
	}

	var stdout bytes.Buffer
	cli.PrintUsage(&stdout)
	assert.Contains(t, stdout.String(), "export      write GSL files and the WebGPU viewer")
}
