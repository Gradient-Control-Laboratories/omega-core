package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"gradientcontrol.com/omega/core/cmd/fix/internal/cli"
)

// Run executes the fix tool with the given arguments and output writer.
func Run(args []string, stdout io.Writer) error {
	return cli.Run(args, stdout)
}

func main() {
	if err := Run(os.Args[1:], os.Stdout); err != nil {
		if !errors.Is(err, cli.ErrDiffFound) {
			fmt.Fprintf(os.Stderr, "fix: %v\n", err)
		}
		os.Exit(1)
	}
}
