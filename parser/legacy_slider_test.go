package parser_test

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

// A variable that sets its slider with .min/.max/.period gets one warning, on
// the first of those statements, that points at @range and `gsl fix -range`.
func TestParsePackage_LegacySliderProperties(t *testing.T) {
	src := "r = 5;\nr.min = 1;\nr.max = 10;\nr.period = 4;\nreturn sphere(r);\n"

	res, err := parser.ParsePackage([]string{src})
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	if len(res.Diagnostics) != 1 {
		t.Fatalf("expected 1 diagnostic for a slider set by three properties, got %d (%+v)", len(res.Diagnostics), res.Diagnostics)
	}
	d := res.Diagnostics[0]
	if d.Severity != program.SeverityWarning {
		t.Errorf("Severity = %v, want SeverityWarning", d.Severity)
	}
	if d.LineNum != 2 || d.Column != 1 {
		t.Errorf("location = %d:%d, want 2:1", d.LineNum, d.Column)
	}
	if d.Line != "r.min = 1;" {
		t.Errorf("Line = %q, want %q", d.Line, "r.min = 1;")
	}
	if d.Length != len("r.min") {
		t.Errorf("Length = %d, want %d", d.Length, len("r.min"))
	}
	for _, want := range []string{"@range", "gsl fix -range", "`r`"} {
		if !strings.Contains(d.Message, want) {
			t.Errorf("expected %q in message, got %q", want, d.Message)
		}
	}

	// CachingParser keeps the warning across a miss and a hit.
	cp := parser.NewCachingParser()
	for i, call := range []string{"first", "second"} {
		got, err := cp.ParsePackage([]string{src})
		if err != nil {
			t.Fatalf("CachingParser %s call failed: %v", call, err)
		}
		if len(got.Diagnostics) != 1 || got.Diagnostics[0] != d {
			t.Fatalf("CachingParser call %d Diagnostics = %+v, want [%+v]", i+1, got.Diagnostics, d)
		}
	}
}

// Each variable warns once, at its earliest slider statement wherever it sits,
// and properties that are not slider limits warn about nothing.
func TestParsePackage_LegacySliderProperties_PerVariable(t *testing.T) {
	src := "a = 1; b = 2;\nb.max = 3; a.min = 0;\nb.min = 1; a.max = 4;\n" +
		"c = [1, 0, 0];\nc.color = [1, 0, 0];\nreturn sphere(a + b);\n"

	res, err := parser.ParsePackage([]string{src})
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	if len(res.Diagnostics) != 2 {
		t.Fatalf("expected 2 diagnostics (a and b), got %d (%+v)", len(res.Diagnostics), res.Diagnostics)
	}
	got := map[string]program.Diagnostic{}
	for _, d := range res.Diagnostics {
		switch {
		case strings.Contains(d.Message, "`a`"):
			got["a"] = d
		case strings.Contains(d.Message, "`b`"):
			got["b"] = d
		}
	}
	if d := got["a"]; d.LineNum != 2 || d.Column != 12 {
		t.Errorf("a's warning at %d:%d, want 2:12 (a.min, after \"b.max = 3; \")", d.LineNum, d.Column)
	}
	if d := got["b"]; d.LineNum != 2 || d.Column != 1 {
		t.Errorf("b's warning at %d:%d, want 2:1 (b.max, the first of b's)", d.LineNum, d.Column)
	}
}

// @range says the same and warns about nothing.
func TestParsePackage_RangeAnnotationDoesNotWarn(t *testing.T) {
	res, err := parser.ParsePackage([]string{"@range(1, 10, 4) r = 5;\nreturn sphere(r);\n"})
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	if len(res.Diagnostics) != 0 {
		t.Errorf("expected 0 diagnostics for @range, got %+v", res.Diagnostics)
	}
}
