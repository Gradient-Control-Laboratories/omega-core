package main

import (
	"fmt"
	"io"
	"os"

	"gradientcontrol.com/omega/core/cmd/compile/internal/cli"
)

// Run executes the compile pipeline with the given arguments and I/O streams.
func Run(args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	return cli.Run(args, stdin, stdout, stderr)
}

func main() {
	if err := Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "compile: %v\n", err)
		os.Exit(1)
	}
}
