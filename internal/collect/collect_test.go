package collect

import (
	"fmt"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/program"
)

func parseMockPackage(sources []string) (*program.Package, error) {
	if len(sources) == 0 {
		return nil, fmt.Errorf("empty sources")
	}
	files := make([]*program.PackageFile, 0, len(sources))
	for _, src := range sources {
		var imports []*program.Import
		if strings.Contains(src, `import "dep1"`) {
			imports = append(imports, &program.Import{Path: "dep1"})
		}
		if strings.Contains(src, `import "dep2"`) {
			imports = append(imports, &program.Import{Path: "dep2"})
		}
		if strings.Contains(src, `import "dep3"`) {
			imports = append(imports, &program.Import{Path: "dep3"})
		}
		files = append(files, &program.PackageFile{
			Source:  src,
			Imports: imports,
			Returns: []*program.Operand{
				{T: &program.Operand_Value{Value: 1.0}},
			},
		})
	}
	return &program.Package{Files: files}, nil
}

func TestCollect_Ordering(t *testing.T) {
	// Setup dependency package sources:
	// primary imports dep1, dep2
	// dep1 imports dep3
	// dep2 imports dep3
	// dep3 has no imports

	dep3Source := `return 3.0;`
	dep1Source := `import "dep3"; return 1.0;`
	dep2Source := `import "dep3"; return 2.0;`
	primarySource := `import "dep1"; import "dep2"; return 42.0;`

	srcMgr := &source.MockManager{
		PkgSources: map[string][]string{
			"dep1": {dep1Source},
			"dep2": {dep2Source},
			"dep3": {dep3Source},
		},
	}

	prog, err := Program([]string{primarySource}, srcMgr, parseMockPackage)
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	// Verify order constraint:
	// "a package earlier in the list cannot be imported by a package later in the list"
	// This means that for any i < j, dependencies[j] cannot import dependencies[i].
	deps := prog.Dependencies
	var customDeps []*program.Package
	for _, dep := range deps {
		if dep.Path == "dep1" || dep.Path == "dep2" || dep.Path == "dep3" {
			customDeps = append(customDeps, dep)
		}
	}
	if len(customDeps) != 3 {
		t.Fatalf("expected 3 custom dependencies, got %d", len(customDeps))
	}

	// Build a map of paths to their indices in the returned dependencies list
	indices := make(map[string]int)
	for i, dep := range deps {
		indices[dep.Path] = i
	}

	// Check importing relations:
	// dep1 imports dep3 -> index(dep1) < index(dep3)
	// dep2 imports dep3 -> index(dep2) < index(dep3)
	idx1, ok1 := indices["dep1"]
	idx2, ok2 := indices["dep2"]
	idx3, ok3 := indices["dep3"]

	if !ok1 || !ok2 || !ok3 {
		t.Fatalf("missing one or more packages in dependencies: dep1=%t, dep2=%t, dep3=%t", ok1, ok2, ok3)
	}

	if idx1 >= idx3 {
		t.Errorf("expected dep1 to be earlier than dep3, but index(dep1)=%d >= index(dep3)=%d", idx1, idx3)
	}
	if idx2 >= idx3 {
		t.Errorf("expected dep2 to be earlier than dep3, but index(dep2)=%d >= index(dep3)=%d", idx2, idx3)
	}
}

func TestCollect_PrimaryWithoutReturnsAndErrorPaths(t *testing.T) {
	t.Run("primary without returns succeeds", func(t *testing.T) {
		pkg := &program.Package{
			Path: "lib/module",
			Files: []*program.PackageFile{
				{Source: "fn helper() { return 1; }"},
			},
		}
		prog, err := Package(pkg, &source.MockManager{}, parseMockPackage)
		if err != nil {
			t.Fatalf("expected primary package without returns to succeed, got: %v", err)
		}
		if len(prog.Returns()) != 0 {
			t.Fatalf("expected 0 returns, got %d", len(prog.Returns()))
		}
	})

	t.Run("missing primary sources and empty package errors", func(t *testing.T) {
		if _, err := Program(nil, &source.MockManager{}, parseMockPackage); err == nil {
			t.Fatal("expected error for empty primarySources")
		}
		if _, err := Package(nil, &source.MockManager{}, parseMockPackage); err == nil {
			t.Fatal("expected error for nil primaryPkg")
		}
		if _, err := Package(&program.Package{}, &source.MockManager{}, parseMockPackage); err == nil {
			t.Fatal("expected error for primaryPkg with no files")
		}
	})

	t.Run("self import and cyclic import errors", func(t *testing.T) {
		selfPkg := &program.Package{
			Path: "dep1",
			Files: []*program.PackageFile{
				{Imports: []*program.Import{{Path: "dep1"}}},
			},
		}
		if _, err := Package(selfPkg, &source.MockManager{}, parseMockPackage); err == nil || !strings.Contains(err.Error(), "cannot import itself") {
			t.Fatalf("expected self-import error, got: %v", err)
		}

		srcMgr := &source.MockManager{
			PkgSources: map[string][]string{
				"dep1": {`import "dep2"; return 1.0;`},
				"dep2": {`import "dep1"; return 2.0;`},
			},
		}
		if _, err := Program([]string{`import "dep1"; return 0.0;`}, srcMgr, parseMockPackage); err == nil || !strings.Contains(err.Error(), "cyclic package import") {
			t.Fatalf("expected cyclic import error, got: %v", err)
		}
	})
}

func TestCollect_SourceHashAndProgramSourceHash(t *testing.T) {
	srcMgr := &source.MockManager{
		PkgSources: map[string][]string{
			"trees/test": {`import "dep1"; import "dep2"; return 42.0;`},
			"dep1":       {`import "dep3"; return 1.0;`},
			"dep2":       {`import "dep3"; return 2.0;`},
			"dep3":       {`return 3.0;`},
			"unrelated":  {`return 999.0;`},
		},
	}

	hash1, err := SourceHash("trees/test", srcMgr, parseMockPackage)
	if err != nil {
		t.Fatalf("SourceHash failed: %v", err)
	}
	if len(hash1) != 64 {
		t.Fatalf("expected 64-char sha256 hex digest, got %q", hash1)
	}

	prog, err := Program(srcMgr.PkgSources["trees/test"], srcMgr, parseMockPackage)
	if err != nil {
		t.Fatalf("Program failed: %v", err)
	}
	if got := ProgramSourceHash(prog); got != hash1 {
		t.Fatalf("ProgramSourceHash = %q, want %q", got, hash1)
	}

	// Modifying an unreferenced package must not change the digest.
	srcMgr.PkgSources["unrelated"] = []string{`return -1.0;`}
	hashUnrelated, err := SourceHash("trees/test", srcMgr, parseMockPackage)
	if err != nil || hashUnrelated != hash1 {
		t.Fatalf("unreferenced package changed digest: got %q, want %q (err=%v)", hashUnrelated, hash1, err)
	}

	// Modifying a transitive dependency must change the digest.
	srcMgr.PkgSources["dep3"] = []string{`return 30.0;`}
	hashChanged, err := SourceHash("trees/test", srcMgr, parseMockPackage)
	if err != nil || hashChanged == hash1 {
		t.Fatalf("transitive dependency edit did not change digest: %q (err=%v)", hashChanged, err)
	}

	// Changing file order within a package must change the digest.
	srcMgr.PkgSources["trees/test"] = []string{`a = 1;`, `import "dep1"; return a;`}
	hashOrderA, err := SourceHash("trees/test", srcMgr, parseMockPackage)
	if err != nil {
		t.Fatalf("SourceHash order A failed: %v", err)
	}
	srcMgr.PkgSources["trees/test"] = []string{`import "dep1"; return a;`, `a = 1;`}
	hashOrderB, err := SourceHash("trees/test", srcMgr, parseMockPackage)
	if err != nil {
		t.Fatalf("SourceHash order B failed: %v", err)
	}
	if hashOrderA == hashOrderB {
		t.Fatal("package file order was not preserved in SourceHash")
	}

	if got := ProgramSourceHash(nil); got != "" {
		t.Fatalf("expected empty hash for nil program, got %q", got)
	}
}
