package program_test

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/program"
)

func TestDiagnostic_String(t *testing.T) {
	d := program.Diagnostic{
		Severity:    program.SeverityWarning,
		PackagePath: "mypkg",
		FilePath:    "mypkg/test.gsl",
		Line:        "fn add(a, b) { return a + b; }",
		LineNum:     1,
		Column:      4,
		Length:      3,
		Message:     `function "add" with 2 parameters shadows implicitly imported builtin function`,
	}

	str := d.String()
	if !strings.Contains(str, "warning:") {
		t.Errorf("expected warning: in string, got %s", str)
	}
	if !strings.Contains(str, "mypkg") {
		t.Errorf("expected package path in string, got %s", str)
	}
	if !strings.Contains(str, "^^^") {
		t.Errorf("expected underline caret in string, got %s", str)
	}
}

func TestDiagnosticSeverity_String(t *testing.T) {
	tests := []struct {
		sev  program.DiagnosticSeverity
		want string
	}{
		{program.SeverityInfo, "info"},
		{program.SeverityWarning, "warning"},
		{program.SeverityError, "error"},
		{program.DiagnosticSeverity(99), "severity(99)"},
	}
	for _, tc := range tests {
		if tc.sev.String() != tc.want {
			t.Errorf("expected %s, got %s", tc.want, tc.sev.String())
		}
	}
}
