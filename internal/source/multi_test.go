package source

import (
	"bytes"
	"errors"
	"testing"
)

type errManager struct {
	srcErr   error
	assetErr error
}

func (e *errManager) Sources(path string) ([]string, []string, error) {
	return nil, nil, e.srcErr
}

func (e *errManager) Asset(pkg, filename string) ([]byte, error) {
	return nil, e.assetErr
}

func TestMultiSourceManagerSources(t *testing.T) {
	first := &MockManager{
		PkgSources: map[string][]string{
			"custom/pkg": {"first = 1;"},
		},
		AssetNames: map[string][]string{
			"custom/pkg": {"mesh.stl"},
		},
	}
	second := &MockManager{
		PkgSources: map[string][]string{
			"custom/pkg": {"second = 2;"},
			"other/pkg":  {"other = 3;"},
		},
	}
	stdlib := NewStdLibManager()

	multi := NewMultiSourceManager(nil, first, second, stdlib)

	// First manager takes precedence on "custom/pkg".
	srcs, assets, err := multi.Sources("custom/pkg")
	if err != nil {
		t.Fatalf("unexpected error loading custom/pkg: %v", err)
	}
	if len(srcs) != 1 || srcs[0] != "first = 1;" {
		t.Fatalf("expected first manager sources, got %v", srcs)
	}
	if len(assets) != 1 || assets[0] != "mesh.stl" {
		t.Fatalf("expected first manager assets, got %v", assets)
	}

	// Falls through to second manager for "other/pkg".
	srcs, _, err = multi.Sources("other/pkg")
	if err != nil {
		t.Fatalf("unexpected error loading other/pkg: %v", err)
	}
	if len(srcs) != 1 || srcs[0] != "other = 3;" {
		t.Fatalf("expected second manager sources, got %v", srcs)
	}

	// Falls through to trailing StdLibManager for "builtin".
	srcs, _, err = multi.Sources("builtin")
	if err != nil || len(srcs) == 0 {
		t.Fatalf("expected builtin from trailing StdLibManager, got srcs=%d, err=%v", len(srcs), err)
	}

	// Returns ErrPkgNotFound when no manager has the package.
	_, _, err = multi.Sources("missing/pkg")
	if !errors.Is(err, ErrPkgNotFound) {
		t.Fatalf("expected ErrPkgNotFound for missing/pkg, got %v", err)
	}
}

func TestMultiSourceManagerAssets(t *testing.T) {
	localEmpty := NewLocalManager(t.TempDir())
	mockWithAsset := &MockManager{
		Assets: map[string]map[string][]byte{
			"pkg/a": {"part.stl": []byte("stl-bytes")},
		},
	}
	stdlib := NewStdLibManager()

	multi := NewMultiSourceManager(localEmpty, mockWithAsset, stdlib)

	// LocalManager returns (nil, nil) on missing asset; MultiSourceManager continues to mockWithAsset.
	data, err := multi.Asset("pkg/a", "part.stl")
	if err != nil {
		t.Fatalf("unexpected error loading asset: %v", err)
	}
	if !bytes.Equal(data, []byte("stl-bytes")) {
		t.Fatalf("expected stl-bytes, got %q", data)
	}

	// Missing asset across all managers returns ErrPkgNotFound.
	_, err = multi.Asset("pkg/a", "missing.stl")
	if !errors.Is(err, ErrPkgNotFound) {
		t.Fatalf("expected ErrPkgNotFound for missing asset, got %v", err)
	}
}

func TestMultiSourceManagerPropagatesNonNotFoundError(t *testing.T) {
	ioErr := errors.New("disk I/O failure")
	failing := &errManager{srcErr: ioErr, assetErr: ioErr}
	stdlib := NewStdLibManager()

	multi := NewMultiSourceManager(failing, stdlib)

	if _, _, err := multi.Sources("builtin"); !errors.Is(err, ioErr) {
		t.Fatalf("expected disk I/O failure to propagate from Sources, got %v", err)
	}
	if _, err := multi.Asset("builtin", "file.stl"); !errors.Is(err, ioErr) {
		t.Fatalf("expected disk I/O failure to propagate from Asset, got %v", err)
	}
}

