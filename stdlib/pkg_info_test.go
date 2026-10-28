package gsllibs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Every registered library must be reachable by the prefix a reference uses,
// which is the last segment of its path. This is what lets the assembler
// auto-import a library the source named but never imported.
func TestPathForPrefix(t *testing.T) {
	cases := map[string]string{
		"map":       "map",
		"lattice":   "lattice",
		"beam":      "lattice/beam",
		"foam":      "lattice/foam",
		"honeycomb": "lattice/honeycomb",
		"tpms":      "lattice/tpms",
		"color":     "color",
	}
	for prefix, want := range cases {
		got, ok := PathForPrefix(prefix)
		if !ok {
			t.Errorf("PathForPrefix(%q): not found, want %q", prefix, want)
			continue
		}
		if got != want {
			t.Errorf("PathForPrefix(%q) = %q, want %q", prefix, got, want)
		}
	}
}

// builtin is implicitly dot-imported, so it is never named by a prefix. A
// non-library name must not resolve either: those need an explicit import.
func TestPathForPrefixUnknown(t *testing.T) {
	for _, prefix := range []string{"builtin", "dm", "gate", "", "lattice/beam"} {
		if path, ok := PathForPrefix(prefix); ok {
			t.Errorf("PathForPrefix(%q) = %q, want not found", prefix, path)
		}
	}
}

// The index covers every library except builtin, and each entry round-trips
// back to the path it was built from.
func TestPathForPrefixCoversAllLibraries(t *testing.T) {
	for _, lib := range All() {
		if lib.ImportPath == "builtin" {
			continue
		}
		prefix := lib.ImportPath
		if i := strings.LastIndex(prefix, "/"); i >= 0 {
			prefix = prefix[i+1:]
		}
		got, ok := PathForPrefix(prefix)
		if !ok {
			t.Errorf("library %q is unreachable: PathForPrefix(%q) not found", lib.ImportPath, prefix)
			continue
		}
		if got != lib.ImportPath {
			t.Errorf("library %q resolves to %q via prefix %q", lib.ImportPath, got, prefix)
		}
	}
}

func TestPrefixesSortedAndComplete(t *testing.T) {
	prefixes := Prefixes()

	want := 0
	for _, lib := range All() {
		if lib.ImportPath != "builtin" {
			want++
		}
	}
	if len(prefixes) != want {
		t.Errorf("Prefixes() returned %d entries, want %d", len(prefixes), want)
	}

	for i := 1; i < len(prefixes); i++ {
		if prefixes[i-1] >= prefixes[i] {
			t.Errorf("Prefixes() not sorted: %q before %q", prefixes[i-1], prefixes[i])
		}
	}

	for _, prefix := range prefixes {
		if _, ok := PathForPrefix(prefix); !ok {
			t.Errorf("Prefixes() returned %q, which PathForPrefix does not resolve", prefix)
		}
	}
}

// The last path segment is the global identity of a library: two libraries
// sharing one leaves auto-import no way to choose, so the index refuses to
// build rather than binding the prefix to whichever came first.
func TestIndexByPrefixRejectsCollision(t *testing.T) {
	_, err := indexByPrefix([]PkgInfo{
		{ImportPath: "color"},
		{ImportPath: "noise/color"},
	})
	if err == nil {
		t.Fatal("indexByPrefix accepted two libraries claiming the prefix \"color\"")
	}
	for _, want := range []string{"color", "noise/color"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not name %q", err, want)
		}
	}
}

// Nesting depth does not make a leaf name available twice.
func TestIndexByPrefixRejectsCollisionAcrossDepths(t *testing.T) {
	if _, err := indexByPrefix([]PkgInfo{
		{ImportPath: "lattice/foam"},
		{ImportPath: "noise/cell/foam"},
	}); err == nil {
		t.Fatal("indexByPrefix accepted \"lattice/foam\" and \"noise/cell/foam\"")
	}
}

// The real registry must be collision-free — this is the check that fires as
// a test failure rather than an init panic when a new library is added.
func TestRegistryHasNoPrefixCollision(t *testing.T) {
	if _, err := indexByPrefix(All()); err != nil {
		t.Fatalf("library registry has a prefix collision: %v", err)
	}
}

// TestRegistryMatchesFilesystem ensures that:
// 1. All GSL packages on disk under the package folder are registered in All().
// 2. Each library registered in All() actually exists as a directory on disk.
func TestRegistryMatchesFilesystem(t *testing.T) {
	registered := make(map[string]bool)
	for _, lib := range All() {
		registered[lib.ImportPath] = true

		// Assert that the registered package exists as a directory on disk.
		// For the "builtin" package, we check the "builtin" subdirectory.
		localPath := filepath.FromSlash(lib.ImportPath)
		info, err := os.Stat(localPath)
		if err != nil {
			if os.IsNotExist(err) {
				t.Errorf("library %q is registered in All() but the directory %q does not exist on disk", lib.ImportPath, localPath)
			} else {
				t.Errorf("failed to stat directory %q: %v", localPath, err)
			}
		} else if !info.IsDir() {
			t.Errorf("library %q is registered in All() but %q is not a directory on disk", lib.ImportPath, localPath)
		}
	}

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && filepath.Ext(path) == ".gsl" {
			dir := filepath.Dir(path)
			if dir == "." {
				return nil
			}
			importPath := filepath.ToSlash(dir)
			if !registered[importPath] {
				t.Errorf("found GSL package folder %q on disk but it is not registered in All()", importPath)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("failed to walk filesystem: %v", err)
	}
}
