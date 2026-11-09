package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// PrintUsage writes the unified gsl CLI help message to w.
func PrintUsage(w io.Writer) {
	fmt.Fprintf(w, `GSL is a tool for managing and compiling Geometry Shading Language (GSL) source code.

Usage: gsl <command> [arguments]

The commands are:

    compile     compile GSL sources to target outputs
    serve       launch local HTTP server and interactive WebGPU viewer
    fix         refactor GSL source files to current syntax
    help        display help for gsl or a specific subcommand

Use "gsl help <command>" or "gsl <command> --help" for more information about a command.
`)
}

// MapSubcommand resolves command aliases (e.g. "build" or "c" -> "compile") to the target binary name.
func MapSubcommand(subcmd string) string {
	switch subcmd {
	case "compile", "build", "c":
		return "compile"
	case "serve":
		return "serve"
	case "fix":
		return "fix"
	default:
		return subcmd
	}
}

func isKnownTool(target string) bool {
	switch target {
	case "compile", "serve", "fix":
		return true
	default:
		return false
	}
}

// ResolveExecutableDir returns the directory containing the currently running gsl binary.
func ResolveExecutableDir() string {
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	if resolved, err := filepath.EvalSymlinks(exe); err == nil {
		exe = resolved
	}
	return filepath.Dir(exe)
}

func resolveToolBinary(binDir, target string) (string, error) {
	binName := target
	if runtime.GOOS == "windows" && filepath.Ext(binName) == "" {
		binName += ".exe"
	}

	if binDir != "" {
		candidate := filepath.Join(binDir, binName)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}

	if isKnownTool(target) {
		if path, err := exec.LookPath(binName); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("unknown command or missing tool binary %q (run ./tool/build-cli.sh to build toolchain binaries)", target)
}

// ExitCode extracts the process exit status from err, returning 0 if err is nil and 1 for non-exec errors.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return exitErr.ExitCode()
	}
	return 1
}

// Run dispatches the requested gsl subcommand to the corresponding sibling tool binary.
func Run(ctx context.Context, binDir string, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	if len(args) == 0 {
		PrintUsage(stdout)
		return nil
	}

	subcmd := args[0]
	switch subcmd {
	case "-h", "-help", "--help":
		PrintUsage(stdout)
		return nil
	case "help":
		if len(args) == 1 {
			PrintUsage(stdout)
			return nil
		}
		target := MapSubcommand(args[1])
		binPath, err := resolveToolBinary(binDir, target)
		if err != nil {
			return err
		}
		return execSubprocess(ctx, binPath, []string{"--help"}, stdin, stdout, stderr)
	}

	target := MapSubcommand(subcmd)
	binPath, err := resolveToolBinary(binDir, target)
	if err != nil {
		return err
	}

	return execSubprocess(ctx, binPath, args[1:], stdin, stdout, stderr)
}

func execSubprocess(ctx context.Context, binPath string, subArgs []string, stdin io.Reader, stdout, stderr io.Writer) error {
	cmd := exec.CommandContext(ctx, binPath, subArgs...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}
