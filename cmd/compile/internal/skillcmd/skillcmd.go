// Package skillcmd is `compile skill`, which `gsl skill` runs: the GSL
// reference, written for a model to read before it writes GSL.
package skillcmd

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/version"
	"github.com/Gradient-Control-Laboratories/omega-core/program/preamble"
)

const usage = `Usage of compile skill (gsl skill):
  gsl skill                  what nearly every program needs: the language, the core built-ins, the builtin
                             package in brief, how packages are referenced and one line per package
  gsl skill NAME...          each name's documentation: a package by import path or prefix (material/bulk/wood
                             or wood), a built-in group (builtin/mesh), a built-in (length), or a member
                             (transform.rotate, or unionRound in every package that has one)
  gsl skill --grep TEXT      every built-in and member whose name or doc contains TEXT, one line each
  gsl skill --all            everything: the language and every package in full

Flags:
      --all
        Print the whole reference
      --grep TEXT
        Search names and docs, ignoring case
  -h, --help
        Display usage instructions

The text is generated from the standard library this program embeds, so it describes exactly the
GSL it compiles. Exit codes: 0 printed, 1 --grep matched nothing, 2 a bad argument or an unknown name.
`

// footer tells the reader of the index how to read further.
const footer = `
---

## **Reading Further**

The index above names each package and built-in group. Before calling a package, read its full
documentation with ` + "`gsl skill PACKAGE...`" + ` (by import path or prefix, as in ` + "`gsl skill builtin wood`" + `);
` + "`gsl skill NAME`" + ` prints one function's every form (` + "`gsl skill unionRound`" + `, ` + "`gsl skill transform.rotate`" + `),
and ` + "`gsl skill --grep TEXT`" + ` finds functions by what their docs say. ` + "`gsl skill --all`" + ` prints everything.
`

const (
	exitPrinted   = 0
	exitNoMatches = 1
	exitBadArgs   = 2
)

// Run is `compile skill ARGS`; it returns the process's exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("skill", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	all := flags.Bool("all", false, "")
	grep := flags.String("grep", "", "")
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
	case *all && (flags.NArg() > 0 || *grep != ""):
		fmt.Fprintf(stderr, "skill: --all prints everything; name packages, pass --grep, or pass --all alone\n")
		return exitBadArgs
	case *grep != "" && flags.NArg() > 0:
		fmt.Fprintf(stderr, "skill: --grep searches everything; pass names or --grep, not both\n")
		return exitBadArgs
	case *all:
		fmt.Fprint(stdout, header+preamble.GSLAuthorSkill())
	case *grep != "":
		found := preamble.Grep(*grep)
		if found == "" {
			fmt.Fprintf(stderr, "skill: no built-in or member mentions %q\n", *grep)
			return exitNoMatches
		}
		fmt.Fprint(stdout, header+found)
	case flags.NArg() == 0:
		fmt.Fprint(stdout, header+preamble.GSLSkillIndex()+footer)
	default:
		docs := make([]string, 0, flags.NArg())
		for _, name := range flags.Args() {
			doc, err := preamble.Lookup(name)
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
