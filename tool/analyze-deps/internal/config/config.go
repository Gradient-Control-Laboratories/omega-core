package config

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// Config represents runtime flags and options for tool/analyze-deps.
type Config struct {
	RootDir   string
	NarrowPkg string
	AllNarrow bool
	Rules     bool
	Graph     bool
	Verbose   bool
	Args      []string
}

// ParseFlags parses command line arguments into Config.
func ParseFlags(args []string) (*Config, error) {
	fs := flag.NewFlagSet("analyze-deps", flag.ContinueOnError)

	var (
		rootDir   = fs.String("root", ".", "Path to omega-core root directory")
		narrowPkg = fs.String("narrow", "", "Analyze a specific internal package for narrowing")
		allNarrow = fs.Bool("all-narrow", false, "Report all internal packages that can be narrowed")
		rules     = fs.Bool("rules", false, "Audit all packages against architectural invariants")
		graph     = fs.Bool("graph", false, "Print ASCII dependency graph between packages")
		verbose   = fs.Bool("v", false, "Verbose output showing import locations")
	)

	fs.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: tool/analyze-deps [flags] [package...]\n\nFlags:\n")
		fs.PrintDefaults()
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil, flag.ErrHelp
		}
		return nil, fmt.Errorf("flag error: %w", err)
	}

	cfg := &Config{
		RootDir:   *rootDir,
		NarrowPkg: *narrowPkg,
		AllNarrow: *allNarrow,
		Rules:     *rules,
		Graph:     *graph,
		Verbose:   *verbose,
		Args:      fs.Args(),
	}

	// Default to -rules if no specific mode selected
	if !cfg.AllNarrow && cfg.NarrowPkg == "" && !cfg.Graph && !cfg.Rules {
		cfg.Rules = true
	}

	if cfg.RootDir == "." {
		cwd, err := os.Getwd()
		if err == nil {
			cfg.RootDir = findModuleRoot(cwd)
		}
	}

	return cfg, nil
}

func findModuleRoot(startDir string) string {
	dir := startDir
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return startDir
}
