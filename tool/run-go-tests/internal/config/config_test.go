package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseFlagsDefaultsAndValidation(t *testing.T) {
	if _, err := ParseFlags(nil); err == nil {
		t.Fatal("expected error when no paths are provided")
	}

	tmp := t.TempDir()
	omegaDir := filepath.Join(tmp, "omega")
	pkgDir := filepath.Join(omegaDir, "tool", "mdpath")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(omegaDir, "go.work"), []byte("go 1.24\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := ParseFlags([]string{pkgDir})
	if err != nil {
		t.Fatalf("ParseFlags failed: %v", err)
	}
	if cfg.OmegaRoot != omegaDir {
		t.Errorf("OmegaRoot = %q, want %q", cfg.OmegaRoot, omegaDir)
	}
	wantHist := filepath.Join(tmp, ".test-history", "omega-test.jsonl")
	if cfg.HistoryFile != wantHist {
		t.Errorf("HistoryFile = %q, want %q", cfg.HistoryFile, wantHist)
	}

	// Single-module repo (like omega-core) with go.mod at root and no go.work.
	coreDir := filepath.Join(tmp, "omega-core")
	corePkgDir := filepath.Join(coreDir, "parser")
	if err := os.MkdirAll(corePkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(coreDir, "go.mod"), []byte("module gradientcontrol.com/omega/core\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	coreCfg, err := ParseFlags([]string{corePkgDir})
	if err != nil {
		t.Fatalf("ParseFlags(omega-core) failed: %v", err)
	}
	if coreCfg.OmegaRoot != coreDir {
		t.Errorf("coreCfg.OmegaRoot = %q, want %q", coreCfg.OmegaRoot, coreDir)
	}
	wantCoreHist := filepath.Join(tmp, ".test-history", "omega-core-test.jsonl")
	if coreCfg.HistoryFile != wantCoreHist {
		t.Errorf("coreCfg.HistoryFile = %q, want %q", coreCfg.HistoryFile, wantCoreHist)
	}
}
