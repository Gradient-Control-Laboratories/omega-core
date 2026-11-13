package graph_test

import (
	"os"
	"path/filepath"
	"testing"

	"gradientcontrol.com/omega/core/tool/analyze-deps/internal/graph"
)

func TestBuild_SimpleModule(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "dep-graph-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create go.mod
	goMod := "module gradientcontrol.com/omega/core\n\ngo 1.22\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	// Create pkgA
	pkgADir := filepath.Join(tmpDir, "internal", "pkga")
	if err := os.MkdirAll(pkgADir, 0755); err != nil {
		t.Fatalf("failed to create pkga dir: %v", err)
	}
	fileA := `package pkga
func Hello() string { return "hello" }
`
	if err := os.WriteFile(filepath.Join(pkgADir, "a.go"), []byte(fileA), 0644); err != nil {
		t.Fatalf("failed to write a.go: %v", err)
	}

	// Create pkgB importing pkgA
	pkgBDir := filepath.Join(tmpDir, "internal", "pkgb")
	if err := os.MkdirAll(pkgBDir, 0755); err != nil {
		t.Fatalf("failed to create pkgb dir: %v", err)
	}
	fileB := `package pkgb
import "gradientcontrol.com/omega/core/internal/pkga"
func Use() string { return pkga.Hello() }
`
	if err := os.WriteFile(filepath.Join(pkgBDir, "b.go"), []byte(fileB), 0644); err != nil {
		t.Fatalf("failed to write b.go: %v", err)
	}

	g, err := graph.Build(tmpDir)
	if err != nil {
		t.Fatalf("graph.Build failed: %v", err)
	}

	if g.ModulePrefix != "gradientcontrol.com/omega/core" {
		t.Errorf("expected module prefix gradientcontrol.com/omega/core, got %s", g.ModulePrefix)
	}

	nodeA := g.Packages["internal/pkga"]
	if nodeA == nil {
		t.Fatalf("expected node for internal/pkga")
	}

	nodeB := g.Packages["internal/pkgb"]
	if nodeB == nil {
		t.Fatalf("expected node for internal/pkgb")
	}

	if len(nodeB.Imports) != 1 {
		t.Errorf("expected nodeB to have 1 import, got %d", len(nodeB.Imports))
	} else if nodeB.Imports[0].ToPkg != "internal/pkga" {
		t.Errorf("expected nodeB import to be internal/pkga, got %s", nodeB.Imports[0].ToPkg)
	}

	if len(nodeA.ImportedBy) != 1 {
		t.Errorf("expected nodeA to be imported by 1 package, got %d", len(nodeA.ImportedBy))
	} else if nodeA.ImportedBy[0].FromPkg != "internal/pkgb" {
		t.Errorf("expected nodeA importedBy to be internal/pkgb, got %s", nodeA.ImportedBy[0].FromPkg)
	}
}

func TestBuild_SkipsTargetDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	goMod := "module gradientcontrol.com/omega/core\n\ngo 1.22\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	targetPkg := filepath.Join(tmpDir, "target", "ignoredpkg")
	if err := os.MkdirAll(targetPkg, 0755); err != nil {
		t.Fatalf("failed to create target pkg dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(targetPkg, "ignored.go"), []byte("package ignoredpkg\n"), 0644); err != nil {
		t.Fatalf("failed to write ignored.go: %v", err)
	}

	g, err := graph.Build(tmpDir)
	if err != nil {
		t.Fatalf("graph.Build failed: %v", err)
	}
	if _, exists := g.Packages["target/ignoredpkg"]; exists {
		t.Fatalf("expected target/ignoredpkg to be skipped by graph.Build")
	}
}

func TestBuild_ExtractsKeepHereReasonAndExternalTestFlag(t *testing.T) {
	tmpDir := t.TempDir()
	goMod := "module gradientcontrol.com/omega/core\n\ngo 1.22\n"
	if err := os.WriteFile(filepath.Join(tmpDir, "go.mod"), []byte(goMod), 0644); err != nil {
		t.Fatalf("failed to write go.mod: %v", err)
	}

	verDir := filepath.Join(tmpDir, "internal", "version")
	if err := os.MkdirAll(verDir, 0755); err != nil {
		t.Fatalf("failed to create version dir: %v", err)
	}
	verSrc := `// Package version holds the release string.
//
// Keep internal/version here because pyproject.toml and wheels.yml
// reference this path.
package version

// Keep other/pkg here because mismatched package paths must be ignored.
var Version = "0.1.1"
`
	if err := os.WriteFile(filepath.Join(verDir, "version.go"), []byte(verSrc), 0644); err != nil {
		t.Fatalf("failed to write version.go: %v", err)
	}

	progDir := filepath.Join(tmpDir, "program")
	if err := os.MkdirAll(progDir, 0755); err != nil {
		t.Fatalf("failed to create program dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(progDir, "a_ext_test.go"), []byte(`package program_test
import "gradientcontrol.com/omega/core/parser"
var _ = parser.Result{}
`), 0644); err != nil {
		t.Fatalf("failed to write a_ext_test.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(progDir, "z_prog.go"), []byte("package program\n"), 0644); err != nil {
		t.Fatalf("failed to write z_prog.go: %v", err)
	}

	g, err := graph.Build(tmpDir)
	if err != nil {
		t.Fatalf("graph.Build failed: %v", err)
	}

	verNode := g.Packages["internal/version"]
	if verNode == nil {
		t.Fatal("expected internal/version node")
	}
	wantReason := "pyproject.toml and wheels.yml reference this path"
	if verNode.KeepHereReason != wantReason {
		t.Errorf("KeepHereReason = %q, want %q", verNode.KeepHereReason, wantReason)
	}

	progNode := g.Packages["program"]
	if progNode == nil {
		t.Fatal("expected program node")
	}
	if progNode.Name != "program" {
		t.Errorf("expected canonical package name \"program\", got %q", progNode.Name)
	}
	if len(progNode.Imports) != 1 || !progNode.Imports[0].IsTest || !progNode.Imports[0].IsExternalTest {
		t.Errorf("expected IsTest && IsExternalTest edge, got %+v", progNode.Imports)
	}
}
