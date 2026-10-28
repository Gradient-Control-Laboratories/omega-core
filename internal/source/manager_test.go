package source

import "testing"

func TestLocalManagerLoadsEmbeddedPackage(t *testing.T) {
	sources, assets, err := NewLocalManager(t.TempDir()).Sources("builtin")
	if err != nil {
		t.Fatalf("load builtin: %v", err)
	}
	if len(sources) == 0 || len(assets) != 0 {
		t.Fatalf("builtin sources = %d, assets = %d", len(sources), len(assets))
	}
}
