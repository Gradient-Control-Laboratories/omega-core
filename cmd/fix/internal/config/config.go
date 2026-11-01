package config

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"gradientcontrol.com/omega/core/cmd/fix/internal/fix"
)

// Mode represents the execution mode determined by CLI arguments.
type Mode int

const (
	// ModeUsage prints the brief usage summary and exits.
	ModeUsage Mode = iota
	// ModeHelp prints the full help text including registered fixers and flags.
	ModeHelp
	// ModeFixerHelp prints detailed help for a specific named fixer.
	ModeFixerHelp
	// ModeRun executes the active fixers on the specified target paths.
	ModeRun
)

// Config holds parsed CLI flags, execution mode, and selected fixers for cmd/fix.
type Config struct {
	Mode         Mode
	HelpTopic    string
	Diff         bool
	Paths        []string
	ActiveFixers []fix.Fixer
}

// ParseFlags parses CLI arguments following Go's cmd/fix conventions.
func ParseFlags(args []string) (*Config, error) {
	if len(args) == 0 {
		return &Config{Mode: ModeUsage}, nil
	}

	if args[0] == "help" {
		if len(args) == 1 {
			return &Config{Mode: ModeHelp}, nil
		}
		if len(args) > 2 {
			return nil, fmt.Errorf("usage: fix help [fixer]")
		}
		topic := args[1]
		if _, ok := fix.ByName(topic); !ok {
			return nil, fmt.Errorf("unknown fixer %q", topic)
		}
		return &Config{
			Mode:      ModeFixerHelp,
			HelpTopic: topic,
		}, nil
	}

	fs := flag.NewFlagSet("fix", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	var diff bool
	fs.BoolVar(&diff, "diff", false, "with -fix, don't update the files, but print a unified diff")

	allFixers := fix.All()
	fixerFlags := make(map[string]*bool, len(allFixers))
	for _, f := range allFixers {
		val := false
		fixerFlags[f.Name] = &val
		fs.BoolVar(&val, f.Name, false, "enable "+f.Name+" fixer")
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return &Config{Mode: ModeHelp}, nil
		}
		return nil, err
	}

	explicitlySet := make(map[string]bool)
	fs.Visit(func(fl *flag.Flag) {
		if _, isFixer := fixerFlags[fl.Name]; isFixer {
			explicitlySet[fl.Name] = true
		}
	})

	paths := fs.Args()
	if len(paths) == 0 {
		return &Config{
			Mode: ModeUsage,
			Diff: diff,
		}, nil
	}

	if paths[0] == "help" {
		if len(paths) == 1 {
			return &Config{Mode: ModeHelp}, nil
		}
		topic := paths[1]
		if _, ok := fix.ByName(topic); !ok {
			return nil, fmt.Errorf("unknown fixer %q", topic)
		}
		return &Config{
			Mode:      ModeFixerHelp,
			HelpTopic: topic,
		}, nil
	}

	activeFixers := selectFixers(allFixers, fixerFlags, explicitlySet)

	return &Config{
		Mode:         ModeRun,
		Diff:         diff,
		Paths:        paths,
		ActiveFixers: activeFixers,
	}, nil
}

func selectFixers(allFixers []fix.Fixer, fixerFlags map[string]*bool, explicitlySet map[string]bool) []fix.Fixer {
	if len(explicitlySet) == 0 {
		return allFixers
	}

	anyExplicitTrue := false
	for name := range explicitlySet {
		if *fixerFlags[name] {
			anyExplicitTrue = true
			break
		}
	}

	var active []fix.Fixer
	for _, f := range allFixers {
		set := explicitlySet[f.Name]
		val := *fixerFlags[f.Name]
		if anyExplicitTrue {
			if set && val {
				active = append(active, f)
			}
		} else {
			// Only -<name>=false flags were specified; run all fixers not explicitly disabled.
			if !set || val {
				active = append(active, f)
			}
		}
	}
	return active
}
