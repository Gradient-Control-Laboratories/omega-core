package lib

import (
	"container/list"
	"crypto/sha256"
	"testing"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/parser/internal/parser2"
	"gradientcontrol.com/omega/core/program"
	gsllibs "gradientcontrol.com/omega/core/stdlib"
)

// A package parsed a second time by CachingParser comes from the cache, and
// what one caller does to its package never reaches the next: assembly rewrites
// the packages it is given in place.
func TestCachingParserHandsOutIndependentCopies(t *testing.T) {
	cp := NewCachingParser()
	src := "r = 2.0;\nreturn length(p) - r;\n"
	firstRes := cp.ParsePackage([]string{src})
	if err := firstRes.Error(); err != nil {
		t.Fatal(err)
	}
	first := firstRes.Package
	want := proto.Clone(first).(*program.Package)

	first.Files[0].Assignments[0].Variables[0].Name = "rewritten"
	first.Files[0].Assignments = nil

	secondRes := cp.ParsePackage([]string{src})
	if err := secondRes.Error(); err != nil {
		t.Fatal(err)
	}
	second := secondRes.Package
	if second == first {
		t.Fatal("the cache handed out the package it gave the first caller")
	}
	if !proto.Equal(second, want) {
		t.Fatalf("the second parse saw the first caller's edits:\n got %v\nwant %v", second, want)
	}
	if _, ok := cp.entries[sourcesKey([]string{src})]; !ok {
		t.Fatal("the package was not cached")
	}
}

// Every standard package cached in a default CachingParser matches a fresh parse.
func TestCachingParserCachedStdPackagesMatchFreshParse(t *testing.T) {
	cp := NewCachingParser()
	for _, lib := range gsllibs.All() {
		res := parser2.Parse(lib.Sources)
		if len(res.Errors) > 0 {
			continue
		}
		if _, ok := cp.entries[sourcesKey(lib.Sources)]; !ok {
			t.Errorf("%s: not cached in NewCachingParser", lib.ImportPath)
			continue
		}
		cached := cp.ParsePackage(lib.Sources)
		if err := cached.Error(); err != nil {
			t.Fatalf("%s: %v", lib.ImportPath, err)
		}
		if !proto.Equal(cached.Package, res.Package) {
			t.Errorf("%s: the cached package differs from a fresh parse", lib.ImportPath)
		}
	}
}

// A parse that fails is not cached, and a later parse reports it again.
func TestCachingParserDoesNotCacheErrors(t *testing.T) {
	cp := NewCachingParser()
	src := "r = ;\nreturn r;\n"
	res := cp.ParsePackage([]string{src})
	if res.Error() == nil {
		t.Fatal("expected a parse error")
	}
	if res == nil || res.Package == nil || len(res.Package.Files) != 1 || res.Package.Files[0] == nil {
		t.Fatalf("expected non-nil partial Package with 1 PackageFile, got %+v", res)
	}
	if _, ok := cp.entries[sourcesKey([]string{src})]; ok {
		t.Fatal("a failed parse was cached")
	}
	if cp.ParsePackage([]string{src}).Error() == nil {
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
	key := func(s string) [32]byte { return sourcesKey([]string{s}) }
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

// Two lists of files with the same concatenation are different packages.
func TestSourcesKeySeparatesFiles(t *testing.T) {
	if sourcesKey([]string{"ab", "c"}) == sourcesKey([]string{"a", "bc"}) {
		t.Fatal("file boundaries do not change the key")
	}
}
