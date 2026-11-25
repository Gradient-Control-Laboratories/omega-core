package parser2_test

import (
	"testing"

<<<<<<< HEAD
	"gradientcontrol.com/omega/core/parser/internal/parser2"
	"gradientcontrol.com/omega/core/program"
=======
	"github.com/Gradient-Control-Laboratories/omega-core/parser/internal/parser2"
>>>>>>> main
)

func TestParseSceneDirective(t *testing.T) {
	// Valid @scene(20) defaults grid to size / 10 = 2.0
	pkg, diags, err := parser2.Parse([]string{"@scene(20)\nreturn sphere(1);"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	if pkg.Scene == nil {
		t.Fatalf("expected package scene to be non-nil")
	}
	if pkg.Scene.Size != 20 || pkg.Scene.Grid != 2.0 {
		t.Errorf("got size=%v grid=%v, want 20 and 2.0", pkg.Scene.Size, pkg.Scene.Grid)
	}

	// Valid @scene(20, 5)
	pkg2, diags, err := parser2.Parse([]string{"@scene(20, 5)\nreturn sphere(1);"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	if pkg2.Scene == nil {
		t.Fatalf("expected package scene to be non-nil")
	}
	if pkg2.Scene.Size != 20 || pkg2.Scene.Grid != 5.0 {
		t.Errorf("got size=%v grid=%v, want 20 and 5.0", pkg2.Scene.Size, pkg2.Scene.Grid)
	}

	// Preceded by whitespace and comments
	pkg3, diags, err := parser2.Parse([]string{"\n  // A helpful header comment\n  @scene(30, 3)\nreturn sphere(1);"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	if pkg3.Scene == nil || pkg3.Scene.Size != 30 || pkg3.Scene.Grid != 3.0 {
		t.Errorf("expected scene with size 30, grid 3.0, got %+v", pkg3.Scene)
	}

	// Semicolon after @scene is rejected
	if _, diags, err := parser2.Parse([]string{"@scene(20);\nreturn sphere(1);"}); err == nil {
		t.Errorf("expected parse error for semicolon after @scene")
	} else if len(diags) != 0 {
		t.Errorf("unexpected diagnostics: %+v", diags)
	}

	// @scene after import is rejected
	if _, diags, err := parser2.Parse([]string{"import \"math.gsl\";\n@scene(20)\nreturn sphere(1);"}); err == nil {
		t.Errorf("expected parse error for @scene after import")
	} else if len(diags) != 0 {
		t.Errorf("unexpected diagnostics: %+v", diags)
	}

	// @scene after declaration is rejected
	if _, diags, err := parser2.Parse([]string{"x = 1;\n@scene(20)\nreturn sphere(1);"}); err == nil {
		t.Errorf("expected parse error for @scene after declaration")
	} else if len(diags) != 0 {
		t.Errorf("unexpected diagnostics: %+v", diags)
	}

	// @scene in file index > 0 is rejected
	file0 := "return sphere(1);"
	file1 := "@scene(20)\nfoo = 1;"
	if _, diags, err := parser2.Parse([]string{file0, file1}); err == nil {
		t.Errorf("expected parse error for @scene in file index > 0")
	} else if len(diags) != 0 {
		t.Errorf("unexpected diagnostics: %+v", diags)
	}

	// Fallback to legacy comments in file 0 when @scene is absent
	legacySrc := "// Scene Size: 40\n// Scene Grid: 4\nreturn sphere(1);"
	pkgLegacy, diags, err := parser2.Parse([]string{legacySrc})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 1 || diags[0].Severity != program.SeverityWarning {
		t.Fatalf("expected 1 warning diagnostic for legacy scene comment, got %+v", diags)
	}
	if pkgLegacy.Scene == nil {
		t.Fatalf("expected package scene from legacy comments")
	}
	if pkgLegacy.Scene.Size != 40 || pkgLegacy.Scene.Grid != 10.0 {
		t.Errorf("got size=%v grid=%v, want 40 and 10.0", pkgLegacy.Scene.Size, pkgLegacy.Scene.Grid)
	}

	// When neither @scene nor legacy comments exist, Scene is nil
	pkgNone, diags, err := parser2.Parse([]string{"return sphere(1);"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("unexpected diagnostics: %+v", diags)
	}
	if pkgNone.Scene != nil {
		t.Errorf("expected nil Scene when absent, got %+v", pkgNone.Scene)
	}
}
