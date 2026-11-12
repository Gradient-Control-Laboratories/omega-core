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

func TestAuditAll_SkipsNestedInternalAndRecognizesEvalSubsystem(t *testing.T) {
	g := &graph.Graph{
		ModulePrefix: "gradientcontrol.com/omega/core",
		Packages: map[string]*graph.PackageNode{
			"internal/fidget/internal/gpuparity": {
				ImportPath: "internal/fidget/internal/gpuparity",
			},
			"internal/env": {
				ImportPath: "internal/env",
				ImportedBy: []graph.ImportEdge{
					{FromPkg: "eval", ToPkg: "internal/env"},
				},
			},
			"eval": {ImportPath: "eval"},
		},
	}

	reports := scoping.AuditAll(g)
	if len(reports) != 1 {
		t.Fatalf("expected 1 root internal report (skipping nested internal/), got %d", len(reports))
	}
	if reports[0].Pkg != "internal/env" || !reports[0].CanNarrow || reports[0].SuggestedDest != "eval/internal/env" {
		t.Errorf("unexpected report: %+v", reports[0])
	}
}
