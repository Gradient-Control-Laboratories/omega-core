package parser_test

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

func TestParsePackage_OmegaDotImportWarning(t *testing.T) {
	src := "@scene(20)\n  import . \"omega\";\nreturn sphere(5);\n"

	res, err := parser.ParsePackage([]string{src})
	if err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
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
	if d.Line != `  import . "omega";` {
		t.Errorf("Line = %q, want %q", d.Line, `  import . "omega";`)
	}
	if d.Length != len(`import . "omega";`) {
		t.Errorf("Length = %d, want %d", d.Length, len(`import . "omega";`))
	}
	if !strings.Contains(d.Message, `import . "omega"`) || !strings.Contains(d.Message, "builtin") || !strings.Contains(d.Message, "gsl fix") {
		t.Errorf("unexpected warning message: %q", d.Message)
	}

	// Verify CachingParser preserves diagnostics across miss and cache hit.
	cp := parser.NewCachingParser()
	first, err := cp.ParsePackage([]string{src})
	if err != nil {
		t.Fatalf("CachingParser first call failed: %v", err)
	}
	if len(first.Diagnostics) != 1 || first.Diagnostics[0] != d {
		t.Fatalf("CachingParser first call Diagnostics = %+v, want [%+v]", first.Diagnostics, d)
	}
	second, err := cp.ParsePackage([]string{src})
	if err != nil {
		t.Fatalf("CachingParser second call failed: %v", err)
	}
	if len(second.Diagnostics) != 1 || second.Diagnostics[0] != d {
		t.Fatalf("CachingParser second call Diagnostics = %+v, want [%+v]", second.Diagnostics, d)
	}

	// Verify non-dot omega import and standard imports emit zero diagnostics.
	for _, cleanSrc := range []string{
		"@scene(20)\nimport \"omega\";\nreturn omega.sphere(5);\n",
		"@scene(20)\nimport \"vector\";\nreturn sphere(5);\n",
	} {
		cleanRes, err := parser.ParsePackage([]string{cleanSrc})
		if err != nil {
			t.Fatalf("ParsePackage(%q) failed: %v", cleanSrc, err)
		}
		if len(cleanRes.Diagnostics) != 0 {
			t.Errorf("expected 0 diagnostics for %q, got %+v", cleanSrc, cleanRes.Diagnostics)
		}
	}
}
