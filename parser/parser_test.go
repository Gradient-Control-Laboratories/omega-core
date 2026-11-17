package parser_test

import (
	"strings"
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

func TestParsePackage_LegacySceneCommentWarning(t *testing.T) {
	src := "// Part: P-1\n  // Scene Size: 40\n// Scene Grid: 10\nreturn sphere(5);"

	res, err := parser.ParsePackage([]string{src})
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	if res.Package.Scene == nil || res.Package.Scene.Size != 40 {
		t.Fatalf("expected parsed legacy Scene Size 40, got %+v", res.Package.Scene)
	}
	if len(res.Diagnostics) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d (%+v)", len(res.Diagnostics), res.Diagnostics)
	}
	d := res.Diagnostics[0]
	if d.Severity != program.SeverityWarning {
		t.Errorf("Severity = %v, want SeverityWarning", d.Severity)
	}
	if d.LineNum != 2 || d.Column != 3 {
		t.Errorf("location = %d:%d, want 2:3", d.LineNum, d.Column)
	}
	if d.Line != "  // Scene Size: 40" {
		t.Errorf("Line = %q, want %q", d.Line, "  // Scene Size: 40")
	}
	if d.Length != len("// Scene Size: 40") {
		t.Errorf("Length = %d, want %d", d.Length, len("// Scene Size: 40"))
	}
	if !strings.Contains(d.Message, "@scene") || !strings.Contains(d.Message, "gsl fix") {
		t.Errorf("expected @scene and gsl fix guidance in message, got %q", d.Message)
	}

	// Verify CachingParser preserves diagnostics across both miss and cache hit.
	cp := parser.NewCachingParser()
	first, err := cp.ParsePackage([]string{src})
	if err != nil {
		t.Fatalf("CachingParser first call failed: %v", err)
	}
	if len(first.Diagnostics) != 1 {
		t.Fatalf("CachingParser first call len(Diagnostics) = %d, want 1", len(first.Diagnostics))
	}
	second, err := cp.ParsePackage([]string{src})
	if err != nil {
		t.Fatalf("CachingParser second call failed: %v", err)
	}
	if len(second.Diagnostics) != 1 || second.Diagnostics[0] != d {
		t.Fatalf("CachingParser second call Diagnostics = %+v, want [%+v]", second.Diagnostics, d)
	}

	// Verify modern @scene directive emits zero diagnostics.
	modernRes, err := parser.ParsePackage([]string{"@scene(40, 4)\nreturn sphere(5);"})
	if err != nil {
		t.Fatalf("ParsePackage(@scene) failed: %v", err)
	}
	if len(modernRes.Diagnostics) != 0 {
		t.Errorf("expected 0 diagnostics for @scene, got %+v", modernRes.Diagnostics)
	}
}
