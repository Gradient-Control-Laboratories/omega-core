package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"gradientcontrol.com/omega/core/cmd/compile/internal/cli"
	"gradientcontrol.com/omega/core/cmd/compile/internal/evalcmd"
	"gradientcontrol.com/omega/core/cmd/compile/internal/skillcmd"
)

func main() {
	// `compile eval` (which `gsl eval` runs) answers JSON queries of a document's field, with exit codes of its own.
	if len(os.Args) > 1 && os.Args[1] == "eval" {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		code := evalcmd.Run(ctx, os.Args[2:], os.Stdin, os.Stdout, os.Stderr)
		stop()
		os.Exit(code)
	}
	// `compile skill` (which `gsl skill` runs) prints the GSL reference for a model to read.
	if len(os.Args) > 1 && os.Args[1] == "skill" {
		os.Exit(skillcmd.Run(os.Args[2:], os.Stdout, os.Stderr))
	}
	if err := cli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "compile: %v\n", err)
		os.Exit(1)
	}
}
