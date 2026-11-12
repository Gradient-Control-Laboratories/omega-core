// Package skillcmd is `compile skill`, which `gsl skill` runs: the GSL
// reference, written for a model to read before it writes GSL.
package skillcmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"gradientcontrol.com/omega/core/cmd/internal/version"
	"gradientcontrol.com/omega/core/program/preamble"
)

const usage = `Usage of compile skill (gsl skill):
  gsl skill                  the language, the built-ins, how packages are referenced and one line per package
  gsl skill PACKAGE...       those packages' full documentation, by import path or prefix (material/bulk/wood or wood)
  gsl skill --all            everything: the language and every package in full (about 470 KB)

Flags:
      --all
        Print the whole reference
  -h, --help
        Display usage instructions

The text is generated from the standard library this program embeds, so it describes exactly the
GSL it compiles. Exit codes: 0 printed, 2 a bad argument or an unknown package.
`

// footer tells the reader of the index how to read further.
const footer = `
---

## **Reading Further**

The package index above names each package. Before calling a package, read its full documentation
with ` + "`gsl skill PACKAGE...`" + ` (by import path or prefix, as in ` + "`gsl skill builtin wood`" + `).
` + "`gsl skill --all`" + ` prints every package at once.
`

const (
	exitPrinted = 0
	exitBadArgs = 2
)

// Run is `compile skill ARGS`; it returns the process's exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("skill", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	all := flags.Bool("all", false, "")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			fmt.Fprint(stdout, usage)
			return exitPrinted
		}
		fmt.Fprintf(stderr, "skill: %v\n\n%s", err, usage)
		return exitBadArgs
	}
	header := fmt.Sprintf("<!-- GSL reference from gsl %s -->\n", version.Version)
	switch {
	case *all && flags.NArg() > 0:
		fmt.Fprintf(stderr, "skill: --all prints every package; name packages or pass --all, not both\n")
		return exitBadArgs
	case *all:
		fmt.Fprint(stdout, header+preamble.GSLAuthorSkill())
	case flags.NArg() == 0:
		fmt.Fprint(stdout, header+preamble.GSLSkillIndex()+footer)
	default:
		docs := make([]string, 0, flags.NArg())
		for _, name := range flags.Args() {
			doc, err := preamble.PackageDescriptiveText(name)
			if err != nil {
				fmt.Fprintf(stderr, "skill: %v\n", err)
				return exitBadArgs
			}
			docs = append(docs, doc)
		}
		fmt.Fprint(stdout, header+strings.Join(docs, "\n"))
	}
	return exitPrinted
}
