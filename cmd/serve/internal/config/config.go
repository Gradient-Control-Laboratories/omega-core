package config

import (
	"flag"
	"io"
	"strings"
)

// Config holds the runtime configuration for the serve HTTP server.
type Config struct {
	Addr        string
	ExamplesDir string
	ViewerDir   string
	SourcePaths []string
	Open        bool
	Verbose     bool
	Help        bool
}

type stringSlice []string

func (s *stringSlice) String() string {
	return strings.Join(*s, ",")
}

func (s *stringSlice) Set(val string) error {
	*s = append(*s, val)
	return nil
}

// ParseFlags parses CLI arguments and returns a populated Config.
func ParseFlags(args []string) (*Config, error) {
	cfg := &Config{}
	var sourcePaths stringSlice

	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(io.Discard)

	fs.StringVar(&cfg.Addr, "addr", "127.0.0.1:8080", "HTTP listen address (host:port)")
	fs.StringVar(&cfg.Addr, "a", "127.0.0.1:8080", "HTTP listen address (shorthand)")
	fs.StringVar(&cfg.ExamplesDir, "examples-dir", "examples", "Directory containing .gsl example files")
	fs.StringVar(&cfg.ExamplesDir, "e", "examples", "Directory containing .gsl example files (shorthand)")
	fs.StringVar(&cfg.ViewerDir, "viewer-dir", "examples/viewer", "Directory containing viewer web assets")
	fs.StringVar(&cfg.ViewerDir, "w", "examples/viewer", "Directory containing viewer web assets (shorthand)")
	fs.Var(&sourcePaths, "source-path", "Additional GSL package search directory (may be repeated)")
	fs.Var(&sourcePaths, "I", "Additional GSL package search directory (shorthand)")
	fs.BoolVar(&cfg.Open, "open", false, "Automatically open browser on startup")
	fs.BoolVar(&cfg.Open, "o", false, "Automatically open browser on startup (shorthand)")
	fs.BoolVar(&cfg.Verbose, "verbose", false, "Log compilation timings and HTTP requests")
	fs.BoolVar(&cfg.Verbose, "v", false, "Log compilation timings and HTTP requests (shorthand)")
	fs.BoolVar(&cfg.Help, "help", false, "Display command usage and exit")
	fs.BoolVar(&cfg.Help, "h", false, "Display command usage and exit (shorthand)")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if len(sourcePaths) == 0 {
		cfg.SourcePaths = []string{cfg.ExamplesDir}
	} else {
		cfg.SourcePaths = []string(sourcePaths)
	}

	return cfg, nil
}

// Usage returns the formatted CLI help text for serve.
func Usage() string {
	return `Usage: serve [flags]

Local HTTP server and interactive WebGPU viewer for the omega-core GSL SDK.

Flags:
  --addr, -a <string>         HTTP listen address (default "127.0.0.1:8080")
  --examples-dir, -e <string> Directory containing .gsl example files (default "examples")
  --viewer-dir, -w <string>   Directory containing viewer web assets (default "examples/viewer")
  --source-path, -I <string>  Additional GSL package search directory (may be repeated; default ["examples"])
  --open, -o                  Automatically open browser on startup (default false)
  --verbose, -v               Log compilation timings and HTTP requests (default false)
  --help, -h                  Display command usage and exit
`
}
