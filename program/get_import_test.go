package program

import (
	"testing"
)

func TestGetImport(t *testing.T) {
	// Standard package setup
	stdPkg := &Package{Path: "std_test"}
	RegisterStdPackage("std_test", stdPkg)

	depPkg := &Package{Path: "dep_test"}
	prog := &Program{
		Dependencies: []*Package{depPkg},
	}

	// 1. Test resolving dependency package
	pkg, err := prog.GetImport("dep_test")
	if err != nil {
		t.Fatalf("unexpected error resolving dependency: %v", err)
	}
	if pkg != depPkg {
		t.Errorf("expected to resolve depPkg, got %v", pkg)
	}

	// 2. Test resolving standard package fallback
	pkg, err = prog.GetImport("std_test")
	if err != nil {
		t.Fatalf("unexpected error resolving stdPkg: %v", err)
	}
	if pkg != stdPkg {
		t.Errorf("expected to resolve stdPkg, got %v", pkg)
	}

	// 3. Test resolving non-existent package
	_, err = prog.GetImport("missing")
	if err == nil {
		t.Error("expected error resolving missing package, got nil")
	}
}
