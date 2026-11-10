package parser

import (
	"container/list"
	"crypto/sha256"
	"testing"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/internal/stdpkg"
	"gradientcontrol.com/omega/core/program"
	gsllibs "gradientcontrol.com/omega/core/stdlib"
)

// A package parsed a second time by CachingParser comes from the cache, and
// what one caller does to its package never reaches the next: assembly rewrites
// the packages it is given in place.
func TestCachingParserHandsOutIndependentCopies(t *testing.T) {
	cp := NewCachingParser()
	src := "r = 2.0;\nreturn length(p) - r;\n"
	firstRes, err := cp.ParsePackage([]string{src})
	if err != nil {
		t.Fatal(err)
	}
	first := firstRes.Package
	want := proto.Clone(first).(*program.Package)

	first.Files[0].Assignments[0].Variables[0].Name = "rewritten"
	first.Files[0].Assignments = nil

	secondRes, err := cp.ParsePackage([]string{src})
	if err != nil {
		t.Fatal(err)
	}
	second := secondRes.Package
	if second == first {
		t.Fatal("the cache handed out the package it gave the first caller")
	}
	if !proto.Equal(second, want) {
		t.Fatalf("the second parse saw the first caller's edits:\n got %v\nwant %v", second, want)
	}
	if _, ok := cp.entries[stdpkg.SourcesHash([]string{src})]; !ok {
		t.Fatal("the package was not cached")
	}
}

// Standard library packages resolve via the shared lock-free program cache
// without consuming CachingParser's LRU capacity, while diverged/edited sources
// fall back to live parsing and LRU caching.
func TestCachingParserSharedStdlibFastPathAndLiveFallback(t *testing.T) {
	cp := NewCachingParser()
	if cp.size != 0 || len(cp.entries) != 0 {
		t.Fatalf("NewCachingParser initial LRU state: size=%d entries=%d, want 0, 0", cp.size, len(cp.entries))
	}

	for _, lib := range gsllibs.All() {
		want := program.GetStdPackage(lib.ImportPath)
		if want == nil {
			t.Fatalf("%s: program.GetStdPackage returned nil", lib.ImportPath)
		}
		gotRes, err := cp.ParsePackage(lib.Sources)
		if err != nil {
			t.Fatalf("%s: %v", lib.ImportPath, err)
		}
		if gotRes.Package == want {
			t.Fatalf("%s: CachingParser did not return an independent clone", lib.ImportPath)
		}
		if !proto.Equal(gotRes.Package, want) {
			t.Errorf("%s: CachingParser package differs from program.GetStdPackage", lib.ImportPath)
		}
		// Mutating returned stdlib clone must not corrupt subsequent parses.
		gotRes.Package.Files = nil
		again, err := cp.ParsePackage(lib.Sources)
		if err != nil || !proto.Equal(again.Package, want) {
			t.Fatalf("%s: mutating returned stdlib package leaked into shared cache", lib.ImportPath)
		}

		// Stateless ParsePackage also resolves stdlib packages via the shared cache.
		directRes, err := ParsePackage(lib.Sources)
		if err != nil || directRes.Package == want || !proto.Equal(directRes.Package, want) {
			t.Fatalf("%s: ParsePackage did not return an independent clone from program.GetStdPackage", lib.ImportPath)
		}
	}

	if cp.size != 0 || len(cp.entries) != 0 {
		t.Fatalf("stdlib hits consumed LRU capacity: size=%d entries=%d, want 0, 0", cp.size, len(cp.entries))
	}

	// Diverged/edited stdlib source misses the hash-matched .pb fast path and falls back to live ParsePackage.
	editedColor := []string{"customColor = [0.2, 0.4, 0.8];\n"}
	editedRes, err := cp.ParsePackage(editedColor)
	if err != nil {
		t.Fatalf("ParsePackage on edited source failed: %v", err)
	}
	if len(editedRes.Package.Files) == 0 || len(editedRes.Package.Files[0].Assignments) != 1 {
		t.Fatalf("expected live-parsed assignment on diverged source, got %v", editedRes.Package)
	}
	if _, ok := cp.entries[stdpkg.SourcesHash(editedColor)]; !ok {
		t.Fatal("expected diverged source to be stored in CachingParser LRU map")
	}
}

// A parse that fails is not cached, and a later parse reports it again.
func TestCachingParserDoesNotCacheErrors(t *testing.T) {
	cp := NewCachingParser()
	src := "r = ;\nreturn r;\n"
	res, err := cp.ParsePackage([]string{src})
	if err == nil {
		t.Fatal("expected a parse error")
	}
	if res == nil || res.Package == nil || len(res.Package.Files) != 1 || res.Package.Files[0] == nil {
		t.Fatalf("expected non-nil partial Package with 1 PackageFile, got %+v", res)
	}
	if _, ok := cp.entries[stdpkg.SourcesHash([]string{src})]; ok {
		t.Fatal("a failed parse was cached")
	}
	if _, err := cp.ParsePackage([]string{src}); err == nil {
		t.Fatal("expected the parse error again")
	}
}

// The cache keeps the most recently used packages within its byte limit.
func TestCachingParserEvictsLeastRecentlyUsed(t *testing.T) {
	c := &CachingParser{
		limit:   10,
		entries: make(map[[sha256.Size]byte]*list.Element),
		order:   list.New(),
	}
	key := func(s string) [32]byte { return stdpkg.SourcesHash([]string{s}) }
	pkg := &program.Package{}
	c.put(key("a"), pkg, 4)
	c.put(key("b"), pkg, 4)
	if c.get(key("a")) == nil {
		t.Fatal("a missing")
	}
	c.put(key("c"), pkg, 4) // over the limit: b is the least recently used
	if c.get(key("b")) != nil {
		t.Fatal("b should have been evicted")
	}
	if c.get(key("a")) == nil || c.get(key("c")) == nil {
		t.Fatal("a and c should remain")
	}
	if c.size != 8 {
		t.Fatalf("size = %d, want 8", c.size)
	}
	c.put(key("big"), pkg, 11)
	if c.get(key("big")) != nil {
		t.Fatal("a package larger than the limit should not be cached")
	}
}
