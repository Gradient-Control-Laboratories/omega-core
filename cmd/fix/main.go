package main

import (
	"errors"
	"fmt"
	"os"

	"gradientcontrol.com/omega/core/cmd/fix/internal/cli"
)

func main() {
	if err := cli.Run(os.Args[1:], os.Stdout); err != nil {
		if !errors.Is(err, cli.ErrDiffFound) {
			fmt.Fprintf(os.Stderr, "fix: %v\n", err)
		}
		os.Exit(1)
	}
}
