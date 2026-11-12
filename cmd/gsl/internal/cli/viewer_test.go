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

// `gsl viewer ARGS` is `compile viewer ARGS`: the viewer lives in the one Fidget-linked program, as eval does.
func TestRun_ViewerRunsCompile(t *testing.T) {
	binDir := t.TempDir()
	createMockTool(t, binDir, "compile", `echo "MOCK_COMPILE:$@"`)

	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"viewer"}, "MOCK_COMPILE:viewer"},
		{[]string{"viewer", "part.gsl", "--no-open"}, "MOCK_COMPILE:viewer part.gsl --no-open"},
		{[]string{"help", "viewer"}, "MOCK_COMPILE:viewer --help"},
	} {
		var stdout, stderr bytes.Buffer
		require.NoError(t, cli.Run(context.Background(), binDir, tc.args, nil, &stdout, &stderr))
		assert.Equal(t, tc.want, strings.TrimSpace(stdout.String()))
	}

	var stdout bytes.Buffer
	cli.PrintUsage(&stdout)
	assert.Contains(t, stdout.String(), "viewer      open a GSL file")
}
