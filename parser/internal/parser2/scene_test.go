package parser2_test

import (
	"testing"

	"gradientcontrol.com/omega/core/parser/internal/parser2"
)

func TestParseSceneDirective(t *testing.T) {
	// Valid @scene(20) defaults grid to size / 10 = 2.0
	pkg, err := parser2.Parse([]string{"@scene(20)\nreturn sphere(1);"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkg.Scene == nil {
		t.Fatalf("expected package scene to be non-nil")
	}
	if pkg.Scene.Size != 20 || pkg.Scene.Grid != 2.0 {
		t.Errorf("got size=%v grid=%v, want 20 and 2.0", pkg.Scene.Size, pkg.Scene.Grid)
	}

	// Valid @scene(20, 5)
	pkg2, err := parser2.Parse([]string{"@scene(20, 5)\nreturn sphere(1);"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkg2.Scene == nil {
		t.Fatalf("expected package scene to be non-nil")
	}
	if pkg2.Scene.Size != 20 || pkg2.Scene.Grid != 5.0 {
		t.Errorf("got size=%v grid=%v, want 20 and 5.0", pkg2.Scene.Size, pkg2.Scene.Grid)
	}

	// Preceded by whitespace and comments
	pkg3, err := parser2.Parse([]string{"\n  // A helpful header comment\n  @scene(30, 3)\nreturn sphere(1);"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkg3.Scene == nil || pkg3.Scene.Size != 30 || pkg3.Scene.Grid != 3.0 {
		t.Errorf("expected scene with size 30, grid 3.0, got %+v", pkg3.Scene)
	}

	// Semicolon after @scene is rejected
	if _, err := parser2.Parse([]string{"@scene(20);\nreturn sphere(1);"}); err == nil {
		t.Errorf("expected parse error for semicolon after @scene")
	}

	// @scene after import is rejected
	if _, err := parser2.Parse([]string{"import \"math.gsl\";\n@scene(20)\nreturn sphere(1);"}); err == nil {
		t.Errorf("expected parse error for @scene after import")
	}

	// @scene after declaration is rejected
	if _, err := parser2.Parse([]string{"x = 1;\n@scene(20)\nreturn sphere(1);"}); err == nil {
		t.Errorf("expected parse error for @scene after declaration")
	}

	// @scene in file index > 0 is rejected
	file0 := "return sphere(1);"
	file1 := "@scene(20)\nfoo = 1;"
	if _, err := parser2.Parse([]string{file0, file1}); err == nil {
		t.Errorf("expected parse error for @scene in file index > 0")
	}

	// Fallback to legacy comments in file 0 when @scene is absent
	legacySrc := "// Scene Size: 40\n// Scene Grid: 4\nreturn sphere(1);"
	pkgLegacy, err := parser2.Parse([]string{legacySrc})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkgLegacy.Scene == nil {
		t.Fatalf("expected package scene from legacy comments")
	}
	if pkgLegacy.Scene.Size != 40 || pkgLegacy.Scene.Grid != 10.0 {
		t.Errorf("got size=%v grid=%v, want 40 and 10.0", pkgLegacy.Scene.Size, pkgLegacy.Scene.Grid)
	}

	// When neither @scene nor legacy comments exist, Scene is nil
	pkgNone, err := parser2.Parse([]string{"return sphere(1);"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pkgNone.Scene != nil {
		t.Errorf("expected nil Scene when absent, got %+v", pkgNone.Scene)
	}
}
