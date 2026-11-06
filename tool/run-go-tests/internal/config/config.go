package config

import (
	"flag"
	"fmt"
	"path/filepath"
)

// Config holds the parsed command-line options and target module paths for run-go-tests.
type Config struct {
	RepoRoot       string
	HistoryFile    string
	WGSLTestdata   string
	FidgetTestdata string
	SummaryFile    string
	FailuresFile   string
	Paths          []string
}

// ParseFlags parses and validates command-line arguments for run-go-tests.
func ParseFlags(args []string) (*Config, error) {
	fs := flag.NewFlagSet("run-go-tests", flag.ContinueOnError)

	var cfg Config
	fs.StringVar(&cfg.RepoRoot, "repo-root", "", "Root directory of the repository (required; used to shorten suite names)")
	fs.StringVar(&cfg.HistoryFile, "history-file", "", "Path to persistent JSONL timing history file outside the repo (required)")
	fs.StringVar(&cfg.WGSLTestdata, "wgsl-testdata", "", "Optional relative path under repo-root to WGSL golden testdata directory")
	fs.StringVar(&cfg.FidgetTestdata, "fidget-testdata", "", "Optional relative path under repo-root to Fidget GLSL golden testdata directory")
	fs.StringVar(&cfg.SummaryFile, "summary-file", "", "Optional path to write the end-of-script duration & comparison summary")
	fs.StringVar(&cfg.FailuresFile, "failures-file", "", "Optional path to write failed suite commands (one per line)")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if cfg.RepoRoot == "" {
		return nil, fmt.Errorf("missing required --repo-root flag")
	}
	if abs, err := filepath.Abs(cfg.RepoRoot); err == nil {
		cfg.RepoRoot = abs
	}

	if cfg.HistoryFile == "" {
		return nil, fmt.Errorf("missing required --history-file flag")
	}

	cfg.Paths = fs.Args()
	if len(cfg.Paths) == 0 {
		return nil, fmt.Errorf("usage: run-go-tests --repo-root <dir> --history-file <file> [flags] <dir> [<dir> ...]")
	}

	return &cfg, nil
}
