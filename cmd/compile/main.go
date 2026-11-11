package main

import (
	"fmt"
	"os"

	"gradientcontrol.com/omega/core/cmd/compile/internal/cli"
)

func main() {
	if err := cli.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintf(os.Stderr, "compile: %v\n", err)
		os.Exit(1)
	}
}
