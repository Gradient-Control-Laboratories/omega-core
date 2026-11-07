package parser2_test

import (
	"testing"

	"gradientcontrol.com/omega/core/parser/internal/parser2"
)

func TestParseSceneDirective(t *testing.T) {
	// Valid @scene(20) defaults grid to size / 10 = 2.0
	res := parser2.Parse([]string{"@scene(20)\nreturn sphere(1);"})
	if len(res.Errors) > 0 {
		t.Fatalf("unexpected parse errors: %v", res.Errors)
	}
	if res.Package.Scene == nil {
		t.Fatalf("expected package scene to be non-nil")
	}
	if res.Package.Scene.Size != 20 || res.Package.Scene.Grid != 2.0 {
		t.Errorf("got size=%v grid=%v, want 20 and 2.0", res.Package.Scene.Size, res.Package.Scene.Grid)
	}

	// Valid @scene(20, 5)
	res2 := parser2.Parse([]string{"@scene(20, 5)\nreturn sphere(1);"})
	if len(res2.Errors) > 0 {
		t.Fatalf("unexpected parse errors: %v", res2.Errors)
	}
	if res2.Package.Scene == nil {
		t.Fatalf("expected package scene to be non-nil")
	}
	if res2.Package.Scene.Size != 20 || res2.Package.Scene.Grid != 5.0 {
		t.Errorf("got size=%v grid=%v, want 20 and 5.0", res2.Package.Scene.Size, res2.Package.Scene.Grid)
	}

	// Preceded by whitespace and comments
	res3 := parser2.Parse([]string{"\n  // A helpful header comment\n  @scene(30, 3)\nreturn sphere(1);"})
	if len(res3.Errors) > 0 {
		t.Fatalf("unexpected parse errors: %v", res3.Errors)
	}
	if res3.Package.Scene == nil || res3.Package.Scene.Size != 30 || res3.Package.Scene.Grid != 3.0 {
		t.Errorf("expected scene with size 30, grid 3.0, got %+v", res3.Package.Scene)
	}

	// Semicolon after @scene is rejected
	resSemi := parser2.Parse([]string{"@scene(20);\nreturn sphere(1);"})
	if resSemi == nil || len(resSemi.Errors) == 0 {
		t.Errorf("expected parse error for semicolon after @scene")
	}

	// @scene after import is rejected
	resAfterImport := parser2.Parse([]string{"import \"math.gsl\";\n@scene(20)\nreturn sphere(1);"})
	if resAfterImport == nil || len(resAfterImport.Errors) == 0 {
		t.Errorf("expected parse error for @scene after import")
	}

	// @scene after declaration is rejected
	resAfterDecl := parser2.Parse([]string{"x = 1;\n@scene(20)\nreturn sphere(1);"})
	if resAfterDecl == nil || len(resAfterDecl.Errors) == 0 {
		t.Errorf("expected parse error for @scene after declaration")
	}

	// @scene in file index > 0 is rejected
	file0 := "return sphere(1);"
	file1 := "@scene(20)\nfoo = 1;"
	resMulti := parser2.Parse([]string{file0, file1})
	if resMulti == nil || len(resMulti.Errors) == 0 {
		t.Errorf("expected parse error for @scene in file index > 0")
	}

	// Fallback to legacy comments in file 0 when @scene is absent
	legacySrc := "// Scene Size: 40\n// Scene Grid: 4\nreturn sphere(1);"
	resLegacy := parser2.Parse([]string{legacySrc})
	if len(resLegacy.Errors) > 0 {
		t.Fatalf("unexpected parse errors: %v", resLegacy.Errors)
	}
	if resLegacy.Package.Scene == nil {
		t.Fatalf("expected package scene from legacy comments")
	}
	if resLegacy.Package.Scene.Size != 40 || resLegacy.Package.Scene.Grid != 10.0 {
		t.Errorf("got size=%v grid=%v, want 40 and 10.0", resLegacy.Package.Scene.Size, resLegacy.Package.Scene.Grid)
	}

	// When neither @scene nor legacy comments exist, Scene is nil
	resNone := parser2.Parse([]string{"return sphere(1);"})
	if len(resNone.Errors) > 0 {
		t.Fatalf("unexpected parse errors: %v", resNone.Errors)
	}
	if resNone.Package.Scene != nil {
		t.Errorf("expected nil Scene when absent, got %+v", resNone.Package.Scene)
	}
}


