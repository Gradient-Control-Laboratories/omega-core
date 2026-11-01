package lib

import (
	"testing"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/internal/meshslots"
	"gradientcontrol.com/omega/core/parser/internal/parser2"
	"gradientcontrol.com/omega/core/program"
	gsllibs "gradientcontrol.com/omega/core/stdlib"
)

// A package parsed a second time comes from the cache, and what one caller
// does to its package never reaches the next: assembly rewrites the packages
// it is given in place.
func TestParsePackageCacheHandsOutIndependentCopies(t *testing.T) {
	src := "r = 2.0;\nreturn length(p) - r;\n"
	first, err := ParsePackage([]string{src})
	if err != nil {
		t.Fatal(err)
	}
	want := proto.Clone(first).(*program.Package)

	first.Files[0].Assignments[0].Variables[0].Name = "rewritten"
	first.Files[0].Assignments = nil

	second, err := ParsePackage([]string{src})
	if err != nil {
		t.Fatal(err)
	}
	if second == first {
		t.Fatal("the cache handed out the package it gave the first caller")
	}
	if !proto.Equal(second, want) {
		t.Fatalf("the second parse saw the first caller's edits:\n got %v\nwant %v", second, want)
	}
	if _, ok := parsed.entries[sourcesKey([]string{src})]; !ok {
		t.Fatal("the package was not cached")
	}
}

// Every standard package, cached at init, is the package a fresh parse gives.
func TestParsePackageCachedStdPackagesMatchFreshParse(t *testing.T) {
	for _, lib := range gsllibs.All() {
		res, err := parser2.Parse(lib.Sources)
		if err != nil || len(res.Errors) > 0 {
			continue
		}
		if _, ok := parsed.entries[sourcesKey(lib.Sources)]; !ok {
			t.Errorf("%s: not cached at init", lib.ImportPath)
			continue
		}
		cached, err := ParsePackage(lib.Sources)
		if err != nil {
			t.Fatalf("%s: %v", lib.ImportPath, err)
		}
		if !proto.Equal(cached, res.Package) {
			t.Errorf("%s: the cached package differs from a fresh parse", lib.ImportPath)
		}
	}
}

// A parse that fails is not cached, and a later parse reports it again.
func TestParsePackageDoesNotCacheErrors(t *testing.T) {
	src := "r = ;\nreturn r;\n"
	if _, err := ParsePackage([]string{src}); err == nil {
		t.Fatal("expected a parse error")
	}
	if _, ok := parsed.entries[sourcesKey([]string{src})]; ok {
		t.Fatal("a failed parse was cached")
	}
	if _, err := ParsePackage([]string{src}); err == nil {
		t.Fatal("expected the parse error again")
	}
}

// The cache keeps the most recently used packages within its byte limit.
func TestParseCacheEvictsLeastRecentlyUsed(t *testing.T) {
	c := newParseCache(10)
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

func TestParsePackageReturnsPartialPackageOnErrorAndGuaranteesFileZero(t *testing.T) {
	t.Run("NilAndEmptySourcesGuaranteePackageAndFileZero", func(t *testing.T) {
		pkgNil, err := ParsePackage(nil)
		if err != nil {
			t.Fatalf("unexpected error for nil sources: %v", err)
		}
		if pkgNil == nil || len(pkgNil.Files) != 1 || pkgNil.Files[0] == nil {
			t.Fatalf("expected non-nil Package with 1 non-nil PackageFile, got %+v", pkgNil)
		}

		pkgEmpty, err := ParsePackage([]string{})
		if err != nil {
			t.Fatalf("unexpected error for empty sources: %v", err)
		}
		if pkgEmpty == nil || len(pkgEmpty.Files) != 1 || pkgEmpty.Files[0] == nil {
			t.Fatalf("expected non-nil Package with 1 non-nil PackageFile, got %+v", pkgEmpty)
		}
	})

	t.Run("PartialPackageReturnedWithErrorAndNotCached", func(t *testing.T) {
		src := `@scene(25)
import "dm/gate";
importStep 0 "stmt.step";
importId = importStep "expr.stp";
broken = ;
`
		pkg, err := ParsePackage([]string{src})
		if err == nil {
			t.Fatal("expected parse error for trailing syntax error")
		}
		if pkg == nil {
			t.Fatal("expected non-nil partial Package alongside error")
		}
		if pkg.Scene == nil || pkg.Scene.Size != 25 {
			t.Fatalf("expected Scene Size 25, got %+v", pkg.Scene)
		}
		if len(pkg.Files) != 1 || pkg.Files[0] == nil {
			t.Fatalf("expected 1 PackageFile, got %+v", pkg.Files)
		}
		file := pkg.Files[0]
		if len(file.Imports) != 1 || file.Imports[0].Path != "dm/gate" {
			t.Fatalf("expected dm/gate import, got %+v", file.Imports)
		}
		if len(file.StepImports) != 1 || file.StepImports[0].Path != "stmt.step" {
			t.Fatalf("expected stmt.step import, got %+v", file.StepImports)
		}
		slots, _ := meshslots.Assign(pkg, nil)
		if len(slots) != 1 || slots[0].Path != "expr.stp" {
			t.Fatalf("expected meshslots.Assign to find expr.stp, got %+v", slots)
		}
		if _, ok := parsed.entries[sourcesKey([]string{src})]; ok {
			t.Fatal("failed parse must not be cached")
		}
	})
}
