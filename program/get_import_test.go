package program

import (
	"testing"

	"google.golang.org/protobuf/proto"
)

func TestGetImportAndGetStdPackage_Standalone(t *testing.T) {
	depPkg := &Package{Path: "dep_test"}
	prog := &Program{
		Dependencies: []*Package{depPkg},
	}

	// 1. Explicit program dependency resolves directly.
	pkg, err := prog.GetImport("dep_test")
	if err != nil {
		t.Fatalf("unexpected error resolving dependency: %v", err)
	}
	if pkg != depPkg {
		t.Errorf("expected depPkg pointer, got %v", pkg)
	}

	// 2. Embedded standard library package resolves standalone without importing parser.
	vecShared, err := prog.GetImport("vector")
	if err != nil {
		t.Fatalf("unexpected error resolving stdlib \"vector\": %v", err)
	}
	if vecShared == nil || len(vecShared.Files) == 0 {
		t.Fatalf("expected populated \"vector\" Package, got %v", vecShared)
	}
	snapshot := proto.Clone(vecShared).(*Package)

	// 3. GetStdPackage returns independent clones so caller mutations never leak.
	clone1 := GetStdPackage("vector")
	if clone1 == nil {
		t.Fatal("GetStdPackage(\"vector\") returned nil")
	}
	if clone1 == vecShared {
		t.Fatal("GetStdPackage(\"vector\") returned the shared canonical pointer instead of a clone")
	}
	clone1.Files = nil

	clone2 := GetStdPackage("vector")
	if !proto.Equal(clone2, snapshot) {
		t.Fatal("mutating first GetStdPackage clone affected subsequent GetStdPackage call")
	}
	if !proto.Equal(vecShared, snapshot) {
		t.Fatal("mutating GetStdPackage clone affected shared GetImport Package")
	}

	// 4. Unknown package returns nil / error.
	if got := GetStdPackage("missing"); got != nil {
		t.Errorf("GetStdPackage(\"missing\") = %v, want nil", got)
	}
	if _, err := prog.GetImport("missing"); err == nil {
		t.Error("expected error resolving missing package, got nil")
	}
}
