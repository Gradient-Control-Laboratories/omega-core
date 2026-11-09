package source

import (
	"errors"
	"testing"
)

func TestStdLibManagerSources(t *testing.T) {
	mgr := NewStdLibManager()

	sources, assets, err := mgr.Sources("builtin")
	if err != nil {
		t.Fatalf("expected builtin package to load, got error: %v", err)
	}
	if len(sources) == 0 {
		t.Fatalf("expected non-empty sources for builtin")
	}
	if len(assets) != 0 {
		t.Fatalf("expected zero assets for builtin, got %d", len(assets))
	}

	_, _, err = mgr.Sources("nonexistent/pkg")
	if !errors.Is(err, ErrPkgNotFound) {
		t.Fatalf("expected ErrPkgNotFound for unknown package, got %v", err)
	}
}

func TestStdLibManagerAsset(t *testing.T) {
	mgr := NewStdLibManager()

	data, err := mgr.Asset("builtin", "part.stl")
	if data != nil {
		t.Fatalf("expected nil asset data, got %v", data)
	}
	if !errors.Is(err, ErrPkgNotFound) {
		t.Fatalf("expected ErrPkgNotFound for stdlib asset, got %v", err)
	}
}

