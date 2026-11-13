package main

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/tool/analyze-deps/internal/graph"
	"gradientcontrol.com/omega/core/tool/analyze-deps/internal/scoping"
)

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe failed: %v", err)
	}
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = orig
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("io.Copy failed: %v", err)
	}
	r.Close()
	return buf.String()
}

func TestPrintScopingAndGraphOutputs(t *testing.T) {
	pinned := &scoping.ScopingReport{
		Pkg:          "internal/version",
		CanNarrow:    false,
		PinnedReason: "pyproject.toml references this path",
		Advice:       "Pinned in root internal/ because pyproject.toml references this path.",
		Subsystems:   []string{"cmd"},
		ImportersBySubsys: map[string][]string{
			"cmd": {"cmd/gsl/internal/cli"},
		},
	}
	shared := &scoping.ScopingReport{
		Pkg:        "internal/lower",
		CanNarrow:  false,
		Advice:     "Cannot narrow directly",
		Subsystems: []string{"assembler", "backend"},
		ImportersBySubsys: map[string][]string{
			"assembler": {"assembler"},
			"backend":   {"backend/wgsl"},
		},
	}

	singleOut := captureStdout(t, func() {
		printScopingReport(pinned, false)
	})
	if !strings.Contains(singleOut, "PINNED IN ROOT internal/") {
		t.Errorf("expected PINNED IN ROOT status in single report, got:\n%s", singleOut)
	}

	auditOut := captureStdout(t, func() {
		printAllScopingAudit([]*scoping.ScopingReport{pinned, shared}, true)
	})
	if !strings.Contains(auditOut, "Packages Pinned in Root internal/ (1):") {
		t.Errorf("expected pinned section in audit output, got:\n%s", auditOut)
	}
	if !strings.Contains(auditOut, "Packages Shared Across Multiple Subsystems (1):") {
		t.Errorf("expected shared section in audit output, got:\n%s", auditOut)
	}

	g := &graph.Graph{
		ModulePrefix: "gradientcontrol.com/omega/core",
		Packages: map[string]*graph.PackageNode{
			"program": {
				ImportPath: "program",
				Imports: []graph.ImportEdge{
					{
						FromPkg:        "program",
						ToPkg:          "parser",
						RawImport:      "gradientcontrol.com/omega/core/parser",
						IsTest:         true,
						IsExternalTest: true,
					},
				},
			},
			"parser": {
				ImportPath: "parser",
				Imports: []graph.ImportEdge{
					{
						FromPkg:   "parser",
						ToPkg:     "program",
						RawImport: "gradientcontrol.com/omega/core/program",
						IsTest:    false,
					},
				},
			},
		},
	}

	prodGraph := captureStdout(t, func() {
		printDependencyGraph(g, false)
	})
	if strings.Contains(prodGraph, "└──> parser") {
		t.Errorf("expected default -graph to omit test-only program->parser edge, got:\n%s", prodGraph)
	}

	verboseGraph := captureStdout(t, func() {
		printDependencyGraph(g, true)
	})
	if !strings.Contains(verboseGraph, "└──> parser") {
		t.Errorf("expected -graph -v to include test-only program->parser edge, got:\n%s", verboseGraph)
	}
}
