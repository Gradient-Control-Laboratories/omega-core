package cli_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/cmd/gsl/internal/cli"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/internal/test/require"
)

func createMockTool(t *testing.T, dir, name, scriptBody string) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("shell script mock requires unix")
	}
	p := filepath.Join(dir, name)
	content := "#!/bin/sh\n" + scriptBody + "\n"
	require.NoError(t, os.WriteFile(p, []byte(content), 0755))
	return p
}

func TestRun_UsageAndHelp(t *testing.T) {
	var stdout, stderr bytes.Buffer

	// No args prints usage and returns nil
	err := cli.Run(context.Background(), "", nil, nil, &stdout, &stderr)
	require.NoError(t, err)
	assert.Contains(t, stdout.String(), "Usage: gsl <command> [arguments]")
	assert.Contains(t, stdout.String(), "compile")
	assert.Contains(t, stdout.String(), "serve")
	assert.Contains(t, stdout.String(), "fix")

	// --help and help also print usage
	for _, arg := range []string{"help", "-h", "-help", "--help"} {
		stdout.Reset()
		err := cli.Run(context.Background(), "", []string{arg}, nil, &stdout, &stderr)
		require.NoError(t, err)
		assert.Contains(t, stdout.String(), "Usage: gsl <command> [arguments]")
	}
}

func TestRun_SubcommandDelegationAndAliases(t *testing.T) {
	binDir := t.TempDir()
	createMockTool(t, binDir, "compile", `echo "MOCK_COMPILE:$@"`)
	createMockTool(t, binDir, "serve", `echo "MOCK_SERVE:$@"`)
	createMockTool(t, binDir, "fix", `echo "MOCK_FIX:$@"`)
	createMockTool(t, binDir, "customtool", `echo "MOCK_CUSTOM:$@"`)

	tests := []struct {
		args []string
		want string
	}{
		{[]string{"compile", "-t", "wgsl", "a.gsl"}, "MOCK_COMPILE:-t wgsl a.gsl"},
		{[]string{"build", "a.gsl"}, "MOCK_COMPILE:a.gsl"},
		{[]string{"c", "a.gsl"}, "MOCK_COMPILE:a.gsl"},
		{[]string{"serve", "--open"}, "MOCK_SERVE:--open"},
		{[]string{"fix", "-diff", "examples"}, "MOCK_FIX:-diff examples"},
		{[]string{"customtool", "--foo"}, "MOCK_CUSTOM:--foo"},
		{[]string{"help", "fix"}, "MOCK_FIX:--help"},
	}

	for _, tc := range tests {
		var stdout, stderr bytes.Buffer
		err := cli.Run(context.Background(), binDir, tc.args, nil, &stdout, &stderr)
		require.NoError(t, err)
		assert.Equal(t, tc.want, strings.TrimSpace(stdout.String()))
	}
}

func TestRun_ExitCodePropagationAndMissingBinary(t *testing.T) {
	binDir := t.TempDir()
	createMockTool(t, binDir, "fix", `echo "diff failed" >&2; exit 3`)

	var stdout, stderr bytes.Buffer
	err := cli.Run(context.Background(), binDir, []string{"fix", "a.gsl"}, nil, &stdout, &stderr)
	require.Error(t, err)
	assert.Equal(t, 3, cli.ExitCode(err))
	assert.Contains(t, stderr.String(), "diff failed")

	// Missing tool binary returns actionable error
	stdout.Reset()
	stderr.Reset()
	err = cli.Run(context.Background(), binDir, []string{"nonexistent_cmd"}, nil, &stdout, &stderr)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent_cmd")
	assert.Equal(t, 1, cli.ExitCode(err))
	assert.Equal(t, 0, cli.ExitCode(nil))

	// Help on missing tool binary returns error
	err = cli.Run(context.Background(), binDir, []string{"help", "nonexistent_cmd"}, nil, &stdout, &stderr)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent_cmd")

	// ResolveExecutableDir returns a non-empty path
	assert.NotEqual(t, "", cli.ResolveExecutableDir())
}

// `gsl eval ARGS` is `compile eval ARGS`: the toolchain keeps one Fidget-linked program. `gsl render ARGS` is
// `compile -t png ARGS`.
func TestRun_EvalAndRenderRunCompile(t *testing.T) {
	binDir := t.TempDir()
	createMockTool(t, binDir, "compile", `echo "MOCK_COMPILE:$@"`)

	for _, tc := range []struct {
		args []string
		want string
	}{
		{[]string{"eval", "--stdio"}, "MOCK_COMPILE:eval --stdio"},
		{[]string{"eval", "-I", "lib", "req.json"}, "MOCK_COMPILE:eval -I lib req.json"},
		{[]string{"help", "eval"}, "MOCK_COMPILE:eval --help"},
		{[]string{"render", "--views", "iso,top", "--light", "a.gsl"}, "MOCK_COMPILE:-t png --views iso,top --light a.gsl"},
		{[]string{"help", "render"}, "MOCK_COMPILE:-t png --help"},
	} {
		var stdout, stderr bytes.Buffer
		require.NoError(t, cli.Run(context.Background(), binDir, tc.args, nil, &stdout, &stderr))
		assert.Equal(t, tc.want, strings.TrimSpace(stdout.String()))
	}

	var stdout bytes.Buffer
	cli.PrintUsage(&stdout)
	assert.Contains(t, stdout.String(), "eval        query a document's field")
	assert.Contains(t, stdout.String(), "render      draw GSL sources into a PNG")
}
