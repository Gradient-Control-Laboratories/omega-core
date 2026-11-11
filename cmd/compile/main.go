package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"

	"gradientcontrol.com/omega/core/cmd/compile/internal/cli"
	"gradientcontrol.com/omega/core/cmd/compile/internal/evalcmd"
)

// Run executes the compile pipeline with the given arguments and I/O streams.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	return cli.Run(args, stdin, stdout, stderr)
}

func main() {
	// `compile eval` (which `gsl eval` runs) answers JSON queries of a document's field, with exit codes of its own.
	if len(os.Args) > 1 && os.Args[1] == "eval" {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		code := evalcmd.Run(ctx, os.Args[2:], os.Stdin, os.Stdout, os.Stderr)
		stop()
		os.Exit(code)
	}
	if err := Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "compile: %v\n", err)
		os.Exit(1)
	}
}
