package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"gradientcontrol.com/omega/core/cmd/gsl/internal/cli"
)

// Run executes the gsl unified CLI driver using the binary directory of the running executable.
func Run(ctx context.Context, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	return cli.Run(ctx, cli.ResolveExecutableDir(), args, stdin, stdout, stderr)
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := Run(ctx, os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			fmt.Fprintf(os.Stderr, "gsl: %v\n", err)
		}
		os.Exit(cli.ExitCode(err))
	}
}
