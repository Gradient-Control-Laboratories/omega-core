package parser_test

import (
	"testing"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

func TestParsePackage_Empty(t *testing.T) {
	res := parser.ParsePackage(nil)
	if res.Error() == nil {
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
	res := parser.ParsePackage([]string{src})
	if err := res.Error(); err != nil {
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
	first := cp.ParsePackage([]string{src})
	if err := first.Error(); err != nil {
		t.Fatalf("CachingParser.ParsePackage failed: %v", err)
	}
	want := proto.Clone(first.Package).(*program.Package)
	first.Package.Files[0].Assignments = nil

	second := cp.ParsePackage([]string{src})
	if err := second.Error(); err != nil {
		t.Fatalf("CachingParser.ParsePackage second call failed: %v", err)
	}
	if !proto.Equal(second.Package, want) {
		t.Fatalf("expected unmodified clone from cache, got %+v", second.Package)
	}
}

func TestResult_Error(t *testing.T) {
	var nilRes *parser.Result
	if nilRes.Error() != nil {
		t.Fatalf("expected nil for nil Result, got %v", nilRes.Error())
	}

	emptyRes := &parser.Result{Package: &program.Package{}}
	if emptyRes.Error() != nil {
		t.Fatalf("expected nil for empty Errors, got %v", emptyRes.Error())
	}

	singleRes := &parser.Result{
		Package: &program.Package{},
		Errors:  []program.Error{{LineNum: 1, Column: 1, Message: "first err"}},
	}
	if err := singleRes.Error(); err == nil || err.Error() != "1:1: first err" {
		t.Fatalf("unexpected single error: %v", err)
	}

	multiRes := &parser.Result{
		Package: &program.Package{},
		Errors: []program.Error{
			{LineNum: 1, Column: 1, Message: "first err"},
			{LineNum: 2, Column: 3, Message: "second err"},
		},
	}
	if err := multiRes.Error(); err == nil || err.Error() != "1:1: first err\n2:3: second err" {
		t.Fatalf("unexpected joined error: %v", err)
	}
}
