package config

import (
	"path/filepath"
	"testing"
)

func TestParseFlagsValidation(t *testing.T) {
	if _, err := ParseFlags(nil); err == nil {
		t.Fatal("expected error when no flags or paths are provided")
	}

	tmp := t.TempDir()
	repoDir := filepath.Join(tmp, "omega")
	histFile := filepath.Join(tmp, ".test-history", "omega-test.jsonl")
	pkgDir := filepath.Join(repoDir, "tool", "mdpath")

	if _, err := ParseFlags([]string{"--history-file", histFile, pkgDir}); err == nil {
		t.Fatal("expected error when --repo-root is missing")
	}
	if _, err := ParseFlags([]string{"--repo-root", repoDir, pkgDir}); err == nil {
		t.Fatal("expected error when --history-file is missing")
	}
	if _, err := ParseFlags([]string{"--repo-root", repoDir, "--history-file", histFile}); err == nil {
		t.Fatal("expected error when no suite paths are provided")
	}

	cfg, err := ParseFlags([]string{
		"--repo-root", repoDir,
		"--history-file", histFile,
		"--wgsl-testdata", "agent/wgsl/internal/codegen/testdata",
		"--fidget-testdata", "agent/fidget/internal/convert/testdata",
		pkgDir,
	})
	if err != nil {
		t.Fatalf("ParseFlags failed: %v", err)
	}
	if cfg.RepoRoot != repoDir {
		t.Errorf("RepoRoot = %q, want %q", cfg.RepoRoot, repoDir)
	}
	if cfg.HistoryFile != histFile {
		t.Errorf("HistoryFile = %q, want %q", cfg.HistoryFile, histFile)
	}
	if cfg.WGSLTestdata != "agent/wgsl/internal/codegen/testdata" {
		t.Errorf("WGSLTestdata = %q, want agent/wgsl/internal/codegen/testdata", cfg.WGSLTestdata)
	}
	if cfg.FidgetTestdata != "agent/fidget/internal/convert/testdata" {
		t.Errorf("FidgetTestdata = %q, want agent/fidget/internal/convert/testdata", cfg.FidgetTestdata)
	}
}
