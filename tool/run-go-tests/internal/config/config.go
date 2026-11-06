package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds the parsed command-line options and target module paths for run-go-tests.
type Config struct {
	OmegaRoot    string
	HistoryFile  string
	SummaryFile  string
	FailuresFile string
	Paths        []string
}

// ParseFlags parses command-line arguments and resolves default paths for OmegaRoot and HistoryFile.
func ParseFlags(args []string) (*Config, error) {
	fs := flag.NewFlagSet("run-go-tests", flag.ContinueOnError)

	var cfg Config
	fs.StringVar(&cfg.OmegaRoot, "omega-root", "", "Root directory of the omega repository (used to shorten suite names)")
	fs.StringVar(&cfg.HistoryFile, "history-file", "", "Path to persistent JSONL timing history file outside the repo")
	fs.StringVar(&cfg.SummaryFile, "summary-file", "", "Optional path to write the end-of-script duration & comparison summary")
	fs.StringVar(&cfg.FailuresFile, "failures-file", "", "Optional path to write failed suite commands (one per line)")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	cfg.Paths = fs.Args()
	if len(cfg.Paths) == 0 {
		return nil, fmt.Errorf("usage: run-go-tests [flags] <dir> [<dir> ...]")
	}

	if cfg.OmegaRoot == "" {
		cfg.OmegaRoot = DetectOmegaRoot(cfg.Paths[0])
	}
	if cfg.OmegaRoot != "" {
		if abs, err := filepath.Abs(cfg.OmegaRoot); err == nil {
			cfg.OmegaRoot = abs
		}
	}

	if cfg.HistoryFile == "" && cfg.OmegaRoot != "" {
		repoName := filepath.Base(cfg.OmegaRoot)
		cfg.HistoryFile = filepath.Join(filepath.Dir(cfg.OmegaRoot), ".test-history", repoName+"-test.jsonl")
	}

	return &cfg, nil
}

// DetectOmegaRoot walks up from samplePath looking for the directory containing
// go.work (multi-module workspace) or the outermost go.mod (single-module repo).
func DetectOmegaRoot(samplePath string) string {
	abs, err := filepath.Abs(samplePath)
	if err != nil {
		return ""
	}
	dir := abs
	outerGoMod := ""
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.work")); err == nil {
			return dir
		}
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			outerGoMod = dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return outerGoMod
		}
		dir = parent
	}
}
