package config_test

import (
	"testing"

	"gradientcontrol.com/omega/core/tool/analyze-deps/internal/config"
)

func TestParseFlags_Defaults(t *testing.T) {
	cfg, err := config.ParseFlags([]string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.Rules {
		t.Errorf("expected Rules to default to true when no specific mode is given")
	}
	if cfg.AllNarrow {
		t.Errorf("expected AllNarrow to default to false")
	}
	if cfg.NarrowPkg != "" {
		t.Errorf("expected NarrowPkg to be empty")
	}
	if cfg.Graph {
		t.Errorf("expected Graph to default to false")
	}
	if cfg.Verbose {
		t.Errorf("expected Verbose to default to false")
	}
}

func TestParseFlags_ExplicitModes(t *testing.T) {
	cfg, err := config.ParseFlags([]string{"-all-narrow", "-v"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !cfg.AllNarrow {
		t.Errorf("expected AllNarrow to be true")
	}
	if !cfg.Verbose {
		t.Errorf("expected Verbose to be true")
	}

	cfgNarrow, err := config.ParseFlags([]string{"-narrow", "internal/meshplan"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfgNarrow.NarrowPkg != "internal/meshplan" {
		t.Errorf("expected NarrowPkg to be internal/meshplan, got %s", cfgNarrow.NarrowPkg)
	}
}
