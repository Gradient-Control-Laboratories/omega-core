package rules_test

import (
	"testing"

	"gradientcontrol.com/omega/core/tool/analyze-deps/internal/graph"
	"gradientcontrol.com/omega/core/tool/analyze-deps/internal/rules"
)

func TestCheckRules_LayerHierarchy(t *testing.T) {
	g := &graph.Graph{
		ModulePrefix: "gradientcontrol.com/omega/core",
		Packages: map[string]*graph.PackageNode{
			"program": {
				ImportPath: "program",
				Imports: []graph.ImportEdge{
					{
						FromPkg:  "program",
						ToPkg:    "assembler",
						FilePath: "program/bad.go",
						Line:     10,
					},
				},
			},
		},
	}

	violations := rules.Check(g)
	if len(violations) == 0 {
		t.Fatal("expected violation for program importing assembler")
	}
	if violations[0].Rule != "layer-hierarchy" {
		t.Errorf("expected rule layer-hierarchy, got %s", violations[0].Rule)
	}
}

func TestCheckRules_BackendIsolation(t *testing.T) {
	g := &graph.Graph{
		ModulePrefix: "gradientcontrol.com/omega/core",
		Packages: map[string]*graph.PackageNode{
			"backend/audio": {
				ImportPath: "backend/audio",
				Imports: []graph.ImportEdge{
					{
						FromPkg:  "backend/audio",
						ToPkg:    "backend/wgsl",
						FilePath: "backend/audio/bad.go",
						Line:     5,
					},
				},
			},
		},
	}

	violations := rules.Check(g)
	if len(violations) == 0 {
		t.Fatal("expected violation for backend/audio importing backend/wgsl")
	}
	if violations[0].Rule != "layer-hierarchy" {
		t.Errorf("expected rule layer-hierarchy, got %s", violations[0].Rule)
	}
}

func TestCheckRules_PassSeparation(t *testing.T) {
	g := &graph.Graph{
		ModulePrefix: "gradientcontrol.com/omega/core",
		Packages: map[string]*graph.PackageNode{
			"internal/lower": {
				ImportPath: "internal/lower",
				Imports: []graph.ImportEdge{
					{
						FromPkg:  "internal/lower",
						ToPkg:    "internal/eval",
						FilePath: "internal/lower/bad.go",
						Line:     12,
					},
				},
			},
		},
	}

	violations := rules.Check(g)
	if len(violations) == 0 {
		t.Fatal("expected violation for internal/lower importing internal/eval")
	}
	if violations[0].Rule != "pass-separation" {
		t.Errorf("expected rule pass-separation, got %s", violations[0].Rule)
	}
}

func TestCheckRules_FrameworkCreep(t *testing.T) {
	g := &graph.Graph{
		ModulePrefix: "gradientcontrol.com/omega/core",
		Packages: map[string]*graph.PackageNode{
			"assembler": {
				ImportPath: "assembler",
				Imports: []graph.ImportEdge{
					{
						FromPkg:   "assembler",
						ToPkg:     "gradientcontrol.com/stirigma/common/errlog",
						RawImport: "gradientcontrol.com/stirigma/common/errlog",
						FilePath:  "assembler/bad.go",
						Line:      8,
					},
				},
			},
		},
	}

	violations := rules.Check(g)
	if len(violations) == 0 {
		t.Fatal("expected violation for importing stirigma")
	}
	if violations[0].Rule != "framework-creep" {
		t.Errorf("expected rule framework-creep, got %s", violations[0].Rule)
	}
}

func TestCheckRules_ProgramExternalTestAllowsParserOnly(t *testing.T) {
	g := &graph.Graph{
		ModulePrefix: "gradientcontrol.com/omega/core",
		Packages: map[string]*graph.PackageNode{
			"program": {
				ImportPath: "program",
				Imports: []graph.ImportEdge{
					{
						FromPkg:        "program",
						ToPkg:          "parser",
						FilePath:       "program/range_test.go",
						Line:           6,
						IsTest:         true,
						IsExternalTest: true,
					},
					{
						FromPkg:        "program",
						ToPkg:          "parser",
						FilePath:       "program/bad.go",
						Line:           6,
						IsTest:         false,
						IsExternalTest: false,
					},
					{
						FromPkg:        "program",
						ToPkg:          "internal/assembler/deadcode",
						FilePath:       "program/bad_test.go",
						Line:           7,
						IsTest:         true,
						IsExternalTest: true,
					},
				},
			},
		},
	}

	violations := rules.Check(g)
	if len(violations) != 2 {
		t.Fatalf("expected 2 violations (prod program->parser and external test program->assembler), got %d: %+v", len(violations), violations)
	}
}
