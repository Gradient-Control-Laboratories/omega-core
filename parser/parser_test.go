package parser_test

import (
	"testing"

	"gradientcontrol.com/omega/core/parser"
)

func TestParsePackage_Empty(t *testing.T) {
	res, err := parser.ParsePackage(nil)
	if err == nil {
		t.Fatal("expected error for empty sources, got nil")
	}
	if res == nil || res.Package == nil || len(res.Package.Files) != 1 || res.Package.Files[0] == nil {
		t.Fatalf("expected non-nil Result and Package with 1 PackageFile, got %+v", res)
	}
	if len(res.Errors) != 1 {
		t.Fatalf("expected 1 structured error in Result.Errors, got %+v", res.Errors)
	}
}

func TestParsePackage_Valid(t *testing.T) {
	src := `x = 10;
return x;`
	res, err := parser.ParsePackage([]string{src})
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	if res == nil || res.Package == nil {
		t.Fatal("expected non-nil Result and Package")
	}
}
