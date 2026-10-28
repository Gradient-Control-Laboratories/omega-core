package main_test

import (
	"os/exec"
	"strings"
	"testing"
)

func TestAnalyzeDepsCLI_Help(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "-h")
	out, _ := cmd.CombinedOutput()
	output := string(out)

	if !strings.Contains(output, "-narrow") || !strings.Contains(output, "-all-narrow") {
		t.Errorf("expected help output to describe -narrow and -all-narrow flags, got:\n%s", output)
	}
}

func TestAnalyzeDepsCLI_AllNarrow(t *testing.T) {
	cmd := exec.Command("go", "run", ".", "-all-narrow")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("expected -all-narrow to succeed, failed with: %v\nOutput: %s", err, string(out))
	}
	output := string(out)
	if !strings.Contains(output, "Package Scoping Audit") {
		t.Errorf("expected audit header in output, got:\n%s", output)
	}
}
