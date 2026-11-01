package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"gradientcontrol.com/omega/core/cmd/fix/internal/cli"
)

// Run executes the fix tool with the given arguments and I/O streams.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	return cli.Run(args, stdin, stdout, stderr)
}

func main() {
	if err := Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		if !errors.Is(err, cli.ErrDiffFound) {
			fmt.Fprintf(os.Stderr, "fix: %v\n", err)
		}
		os.Exit(1)
	}
}
