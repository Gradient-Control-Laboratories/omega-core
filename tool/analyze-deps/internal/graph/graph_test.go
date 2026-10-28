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
