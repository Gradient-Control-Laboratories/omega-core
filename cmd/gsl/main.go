package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"gradientcontrol.com/omega/core/cmd/gsl/internal/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := cli.Run(ctx, cli.ResolveExecutableDir(), os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			fmt.Fprintf(os.Stderr, "gsl: %v\n", err)
		}
		os.Exit(cli.ExitCode(err))
	}
}
