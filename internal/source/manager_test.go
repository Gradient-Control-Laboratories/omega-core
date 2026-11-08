package source

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestLocalManagerLoadsLocalSourcesAndAssets(t *testing.T) {
	root := t.TempDir()
	pkgDir := filepath.Join(root, "custom/pkg")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatalf("mkdir pkg: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "pkg.gsl"), []byte("x = 1;\n"), 0o644); err != nil {
		t.Fatalf("write gsl: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "part.stl"), []byte("stl-bytes"), 0o644); err != nil {
		t.Fatalf("write asset: %v", err)
	}

	mgr := NewLocalManager(root)
	sources, assets, err := mgr.Sources("custom/pkg")
	if err != nil {
		t.Fatalf("Sources(custom/pkg) error: %v", err)
	}
	if len(sources) != 1 || sources[0] != "x = 1;\n" {
		t.Fatalf("unexpected sources: %v", sources)
	}
	if len(assets) != 1 || assets[0] != "part.stl" {
		t.Fatalf("unexpected assets: %v", assets)
	}

	data, err := mgr.Asset("custom/pkg", "part.stl")
	if err != nil {
		t.Fatalf("Asset(custom/pkg, part.stl) error: %v", err)
	}
	if string(data) != "stl-bytes" {
		t.Fatalf("unexpected asset data %q", string(data))
	}
}

func TestLocalManagerDoesNotLoadEmbeddedStdLib(t *testing.T) {
	mgr := NewLocalManager(t.TempDir())
	_, _, err := mgr.Sources("builtin")
	if !errors.Is(err, ErrPkgNotFound) {
		t.Fatalf("expected ErrPkgNotFound for embedded stdlib package 'builtin' on LocalManager, got: %v", err)
	}
}
