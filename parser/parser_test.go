package parser_test

import (
	"testing"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

func TestParsePackage_Empty(t *testing.T) {
	res, err := parser.ParsePackage(nil)
	if err == nil {
		t.Fatal("expected error for empty sources, got nil")
	}
	if res == nil || res.Package == nil || len(res.Package.Files) != 1 || res.Package.Files[0] == nil {
		t.Fatalf("expected non-nil Result and Package with 1 PackageFile, got %+v", res)
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

func TestCachingParser(t *testing.T) {
	cp := parser.NewCachingParser()
	src := `x = 10;
return x;`
	first, err := cp.ParsePackage([]string{src})
	if err != nil {
		t.Fatalf("CachingParser.ParsePackage failed: %v", err)
	}
	want := proto.Clone(first.Package).(*program.Package)
	first.Package.Files[0].Assignments = nil

	second, err := cp.ParsePackage([]string{src})
	if err != nil {
		t.Fatalf("CachingParser.ParsePackage second call failed: %v", err)
	}
	if !proto.Equal(second.Package, want) {
		t.Fatalf("expected unmodified clone from cache, got %+v", second.Package)
	}
}
