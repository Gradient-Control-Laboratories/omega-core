package graph

import (
	"bufio"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ImportEdge represents a single import relationship between packages.
type ImportEdge struct {
	FromPkg   string
	ToPkg     string
	RawImport string
	FilePath  string
	Line      int
	IsTest    bool
}

// PackageNode represents a Go package in the module.
type PackageNode struct {
	ImportPath string // Relative to module root, e.g. "internal/lower"
	FullImport string // Full module path, e.g. "gradientcontrol.com/omega/core/internal/lower"
	Dir        string // Relative filesystem directory
	Name       string // Go package clause name
	Imports    []ImportEdge
	ImportedBy []ImportEdge
}

// Graph represents the directed dependency graph of the module.
type Graph struct {
	ModulePrefix string
	Packages     map[string]*PackageNode // Key is relative package path
}

// Build traverses rootDir and constructs the dependency graph.
func Build(rootDir string) (*Graph, error) {
	modulePrefix, err := readModulePrefix(rootDir)
	if err != nil {
		return nil, err
	}

	g := &Graph{
		ModulePrefix: modulePrefix,
		Packages:     make(map[string]*PackageNode),
	}

	fset := token.NewFileSet()

	err = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			base := info.Name()
			if base == ".git" || base == "bin" || base == "testdata" {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") {
			return nil
		}

		relDir, err := filepath.Rel(rootDir, filepath.Dir(path))
		if err != nil {
			return err
		}
		relDir = filepath.Clean(relDir)
		if relDir == "." {
			relDir = ""
		}

		// Parse file imports only
		node, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return fmt.Errorf("failed to parse %s: %w", path, err)
		}

		pkgRel := relDir
		pkgNode := g.Packages[pkgRel]
		if pkgNode == nil {
			fullImport := g.ModulePrefix
			if pkgRel != "" {
				fullImport = g.ModulePrefix + "/" + pkgRel
			}
			pkgNode = &PackageNode{
				ImportPath: pkgRel,
				FullImport: fullImport,
				Dir:        relDir,
				Name:       node.Name.Name,
				Imports:    nil,
				ImportedBy: nil,
			}
			g.Packages[pkgRel] = pkgNode
		}

		relFile, _ := filepath.Rel(rootDir, path)
		isTest := strings.HasSuffix(path, "_test.go")

		for _, imp := range node.Imports {
			rawPath, err := strconv.Unquote(imp.Path.Value)
			if err != nil {
				rawPath = strings.Trim(imp.Path.Value, `"`)
			}

			targetPkg := rawPath
			if strings.HasPrefix(rawPath, g.ModulePrefix+"/") {
				targetPkg = strings.TrimPrefix(rawPath, g.ModulePrefix+"/")
			} else if rawPath == g.ModulePrefix {
				targetPkg = ""
			}

			line := fset.Position(imp.Pos()).Line
			edge := ImportEdge{
				FromPkg:   pkgRel,
				ToPkg:     targetPkg,
				RawImport: rawPath,
				FilePath:  relFile,
				Line:      line,
				IsTest:    isTest,
			}
			pkgNode.Imports = append(pkgNode.Imports, edge)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed scanning packages: %w", err)
	}

	// Build reverse ImportedBy index
	for _, node := range g.Packages {
		for _, edge := range node.Imports {
			targetNode := g.Packages[edge.ToPkg]
			if targetNode != nil {
				targetNode.ImportedBy = append(targetNode.ImportedBy, edge)
			}
		}
	}

	return g, nil
}

func readModulePrefix(rootDir string) (string, error) {
	goModPath := filepath.Join(rootDir, "go.mod")
	f, err := os.Open(goModPath)
	if err != nil {
		return "", fmt.Errorf("failed to open go.mod at %s: %w", goModPath, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module ")), nil
		}
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return "", fmt.Errorf("module declaration not found in %s", goModPath)
}

// SubsystemOf returns the high-level compiler subsystem for a package:
// "program", "parser", "assembler", "backend", "internal", "stdlib", "common", "tool", or "root".
func SubsystemOf(pkg string) string {
	clean := filepath.Clean(pkg)
	if clean == "." || clean == "" {
		return "root"
	}
	parts := strings.Split(clean, "/")
	if len(parts) == 0 {
		return "root"
	}
	first := parts[0]
	switch first {
	case "program", "parser", "assembler", "stdlib", "common", "tool":
		return first
	case "backend":
		if len(parts) > 1 {
			// Subsystem could be backend target (backend/wgsl, backend/audio, backend/mesh, backend/tape)
			// or backend infrastructure (backend/common, backend/internal)
			return parts[0] + "/" + parts[1]
		}
		return "backend"
	case "internal":
		if len(parts) > 1 && parts[1] == "assembler" {
			return "assembler"
		}
		return "internal"
	default:
		return first
	}
}
