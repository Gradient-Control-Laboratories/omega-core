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
	src := sources[0]
	var imports []*program.Import

	// Parse simple import declarations in the mock sources
	if strings.Contains(src, `import "dep1"`) {
		imports = append(imports, &program.Import{Path: "dep1"})
	}
	if strings.Contains(src, `import "dep2"`) {
		imports = append(imports, &program.Import{Path: "dep2"})
	}
	if strings.Contains(src, `import "dep3"`) {
		imports = append(imports, &program.Import{Path: "dep3"})
	}

	return &program.Package{
		Files: []*program.PackageFile{
			{
				Source:  src,
				Imports: imports,
				Returns: []*program.Operand{
					{T: &program.Operand_Value{Value: 1.0}},
				},
			},
		},
	}, nil
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

func TestCollect_NamedPrimaryPackage(t *testing.T) {
	depSource := `return 1.0;`
	primarySource := `import "dep1"; return 42.0;`

	srcMgr := &source.MockManager{
		PkgSources: map[string][]string{
			"dep1": {depSource},
		},
	}

	parseNamedMock := func(sources []string) (*program.Package, error) {
		pkg, err := parseMockPackage(sources)
		if err != nil {
			return nil, err
		}
		if sources[0] == primarySource {
			pkg.Path = "custom/main_pkg"
		}
		return pkg, nil
	}

	prog, err := Program([]string{primarySource}, srcMgr, parseNamedMock)
	if err != nil {
		t.Fatalf("Collect failed: %v", err)
	}

	if prog.Primary == nil || prog.Primary.Path != "custom/main_pkg" {
		t.Fatalf("expected primary package path 'custom/main_pkg', got %q", prog.Primary.GetPath())
	}

	foundDep1 := false
	for _, dep := range prog.Dependencies {
		if dep.Path == "dep1" {
			foundDep1 = true
			break
		}
	}
	if !foundDep1 {
		t.Fatalf("expected dependency 'dep1' in %v", prog.Dependencies)
	}
}
