package parser_test

import (
	"testing"

	"gradientcontrol.com/omega/core/parser"
)

func TestParsePackage_Empty(t *testing.T) {
	_, err := parser.ParsePackage(nil)
	if err == nil {
		t.Fatal("expected error for empty sources, got nil")
	}
}

func TestParsePackage_Valid(t *testing.T) {
	src := `x = 10;
return x;`
	pkg, err := parser.ParsePackage([]string{src})
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	if pkg == nil {
		t.Fatal("expected non-nil Package")
	}
}
