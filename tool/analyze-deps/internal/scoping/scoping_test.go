package scoping_test

import (
	"testing"

	"gradientcontrol.com/omega/core/tool/analyze-deps/internal/graph"
	"gradientcontrol.com/omega/core/tool/analyze-deps/internal/scoping"
)

func TestAnalyzePackage_CanNarrow(t *testing.T) {
	// A package internal/foo imported ONLY by packages in assembler
	g := &graph.Graph{
		ModulePrefix: "gradientcontrol.com/omega/core",
		Packages: map[string]*graph.PackageNode{
			"internal/foo": {
				ImportPath: "internal/foo",
				ImportedBy: []graph.ImportEdge{
					{FromPkg: "assembler", ToPkg: "internal/foo"},
					{FromPkg: "internal/assembler", ToPkg: "internal/foo"},
				},
			},
			"assembler": {
				ImportPath: "assembler",
			},
			"internal/assembler": {
				ImportPath: "internal/assembler",
			},
		},
	}

	report := scoping.AnalyzePackage(g, "internal/foo")
	if report == nil {
		t.Fatal("expected report, got nil")
	}
	if !report.CanNarrow {
		t.Errorf("expected CanNarrow to be true")
	}
	if report.SuggestedSubsystem != "assembler" {
		t.Errorf("expected SuggestedSubsystem assembler, got %s", report.SuggestedSubsystem)
	}
}

func TestAnalyzePackage_BlockedByMultipleSubsystems(t *testing.T) {
	// A package internal/bar imported by BOTH assembler and backend/mesh
	g := &graph.Graph{
		ModulePrefix: "gradientcontrol.com/omega/core",
		Packages: map[string]*graph.PackageNode{
			"internal/bar": {
				ImportPath: "internal/bar",
				ImportedBy: []graph.ImportEdge{
					{FromPkg: "assembler", ToPkg: "internal/bar"},
					{FromPkg: "backend/mesh", ToPkg: "internal/bar"},
				},
			},
			"assembler":    {ImportPath: "assembler"},
			"backend/mesh": {ImportPath: "backend/mesh"},
		},
	}

	report := scoping.AnalyzePackage(g, "internal/bar")
	if report == nil {
		t.Fatal("expected report, got nil")
	}
	if report.CanNarrow {
		t.Errorf("expected CanNarrow to be false because imported by multiple subsystems")
	}
	if len(report.Subsystems) != 2 {
		t.Errorf("expected 2 subsystems, got %d", len(report.Subsystems))
	}
}
