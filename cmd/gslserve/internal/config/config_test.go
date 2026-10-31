package config_test

import (
	"reflect"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/cmd/gslserve/internal/config"
)

func TestParseFlags_Defaults(t *testing.T) {
	cfg, err := config.ParseFlags(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Addr != "127.0.0.1:8080" {
		t.Errorf("expected default Addr '127.0.0.1:8080', got %q", cfg.Addr)
	}
	if cfg.ExamplesDir != "examples" {
		t.Errorf("expected default ExamplesDir 'examples', got %q", cfg.ExamplesDir)
	}
	if cfg.ViewerDir != "examples/viewer" {
		t.Errorf("expected default ViewerDir 'examples/viewer', got %q", cfg.ViewerDir)
	}
	if !reflect.DeepEqual(cfg.SourcePaths, []string{"examples"}) {
		t.Errorf("expected default SourcePaths ['examples'], got %v", cfg.SourcePaths)
	}
	if cfg.Open {
		t.Errorf("expected default Open false")
	}
	if cfg.Verbose {
		t.Errorf("expected default Verbose false")
	}
	if cfg.Help {
		t.Errorf("expected default Help false")
	}
}

func TestParseFlags_LongAndShort(t *testing.T) {
	args := []string{
		"--addr", "0.0.0.0:9090",
		"-e", "/tmp/examples",
		"-w", "/tmp/viewer",
		"-I", "/tmp/lib1",
		"--source-path", "/tmp/lib2",
		"-o",
		"--verbose",
	}
	cfg, err := config.ParseFlags(args)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Addr != "0.0.0.0:9090" {
		t.Errorf("expected Addr '0.0.0.0:9090', got %q", cfg.Addr)
	}
	if cfg.ExamplesDir != "/tmp/examples" {
		t.Errorf("expected ExamplesDir '/tmp/examples', got %q", cfg.ExamplesDir)
	}
	if cfg.ViewerDir != "/tmp/viewer" {
		t.Errorf("expected ViewerDir '/tmp/viewer', got %q", cfg.ViewerDir)
	}
	wantPaths := []string{"/tmp/lib1", "/tmp/lib2"}
	if !reflect.DeepEqual(cfg.SourcePaths, wantPaths) {
		t.Errorf("expected SourcePaths %v, got %v", wantPaths, cfg.SourcePaths)
	}
	if !cfg.Open {
		t.Errorf("expected Open true")
	}
	if !cfg.Verbose {
		t.Errorf("expected Verbose true")
	}
}

func TestParseFlags_ShortFlagsAndHelp(t *testing.T) {
	cfg, err := config.ParseFlags([]string{"-a", "localhost:3000", "-v", "-h"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Addr != "localhost:3000" {
		t.Errorf("expected Addr 'localhost:3000', got %q", cfg.Addr)
	}
	if !cfg.Verbose {
		t.Errorf("expected Verbose true")
	}
	if !cfg.Help {
		t.Errorf("expected Help true")
	}
}

func TestParseFlags_InvalidFlag(t *testing.T) {
	_, err := config.ParseFlags([]string{"--nonexistent-flag"})
	if err == nil {
		t.Fatalf("expected error on unknown flag")
	}
}

func TestUsage(t *testing.T) {
	u := config.Usage()
	for _, flagName := range []string{"--addr", "--examples-dir", "--viewer-dir", "--source-path", "--open", "--verbose", "--help"} {
		if !strings.Contains(u, flagName) {
			t.Errorf("expected Usage() to mention %s, got:\n%s", flagName, u)
		}
	}
}
