package rules

import (
	"fmt"
	"strings"

	"gradientcontrol.com/omega/core/tool/analyze-deps/internal/graph"
)

// Violation represents an architectural invariant violation.
type Violation struct {
	Rule        string // "layer-hierarchy", "pass-separation", "framework-creep"
	FromPkg     string
	ToPkg       string
	FilePath    string
	Line        int
	IsTest      bool
	Description string
}

// Check validates all packages in the graph against the three core invariants.
func Check(g *graph.Graph) []Violation {
	var violations []Violation

	for _, node := range g.Packages {
		for _, edge := range node.Imports {
			// Rule 3: Prohibition of Framework Creep (stirigma)
			if strings.HasPrefix(edge.RawImport, "gradientcontrol.com/stirigma/") {
				violations = append(violations, Violation{
					Rule:        "framework-creep",
					FromPkg:     edge.FromPkg,
					ToPkg:       edge.RawImport,
					FilePath:    edge.FilePath,
					Line:        edge.Line,
					IsTest:      edge.IsTest,
					Description: fmt.Sprintf("package imports prohibited Stirigma framework package: %s", edge.RawImport),
				})
			}

			// Rule 2: Compiler Pass Separation of Concerns
			if isLoweringPass(edge.FromPkg) && isEvalPass(edge.ToPkg) {
				violations = append(violations, Violation{
					Rule:        "pass-separation",
					FromPkg:     edge.FromPkg,
					ToPkg:       edge.ToPkg,
					FilePath:    edge.FilePath,
					Line:        edge.Line,
					IsTest:      edge.IsTest,
					Description: fmt.Sprintf("structural lowering pass %s must not import constant evaluation %s", edge.FromPkg, edge.ToPkg),
				})
			}

			// Rule 1: Layer Hierarchy & Subsystem Isolation
			fromLayer := layerOf(edge.FromPkg)
			toLayer := layerOf(edge.ToPkg)

			if fromLayer > 0 && toLayer > 0 && fromLayer < toLayer {
				violations = append(violations, Violation{
					Rule:        "layer-hierarchy",
					FromPkg:     edge.FromPkg,
					ToPkg:       edge.ToPkg,
					FilePath:    edge.FilePath,
					Line:        edge.Line,
					IsTest:      edge.IsTest,
					Description: fmt.Sprintf("lower layer %s cannot import higher layer %s", edge.FromPkg, edge.ToPkg),
				})
			}

			// Backend Target Isolation
			fromBackendTarget := backendTargetOf(edge.FromPkg)
			toBackendTarget := backendTargetOf(edge.ToPkg)
			if fromBackendTarget != "" && toBackendTarget != "" && fromBackendTarget != toBackendTarget {
				violations = append(violations, Violation{
					Rule:        "layer-hierarchy",
					FromPkg:     edge.FromPkg,
					ToPkg:       edge.ToPkg,
					FilePath:    edge.FilePath,
					Line:        edge.Line,
					IsTest:      edge.IsTest,
					Description: fmt.Sprintf("backend target %s cannot import sibling backend target %s (use backend/common or backend/internal)", edge.FromPkg, edge.ToPkg),
				})
			}
		}
	}

	return violations
}

func layerOf(pkg string) int {
	clean := strings.Trim(pkg, "/")
	if clean == "program" || strings.HasPrefix(clean, "program/") {
		return 1
	}
	if clean == "parser" || strings.HasPrefix(clean, "parser/") {
		return 2
	}
	if clean == "assembler" || strings.HasPrefix(clean, "assembler/") || strings.HasPrefix(clean, "internal/assembler") {
		return 3
	}
	if clean == "backend" || strings.HasPrefix(clean, "backend/") {
		return 4
	}
	return 0 // internal shared, stdlib, external
}

func backendTargetOf(pkg string) string {
	clean := strings.Trim(pkg, "/")
	for _, target := range []string{"backend/audio", "backend/wgsl", "backend/mesh", "backend/tape"} {
		if clean == target || strings.HasPrefix(clean, target+"/") {
			return target
		}
	}
	return ""
}

func isLoweringPass(pkg string) bool {
	clean := strings.Trim(pkg, "/")
	return clean == "internal/lower" ||
		clean == "internal/assembler/unroll" ||
		clean == "internal/assembler/topofold" ||
		strings.HasSuffix(clean, "/lower") ||
		strings.HasSuffix(clean, "/unroll") ||
		strings.HasSuffix(clean, "/topofold")
}

func isEvalPass(pkg string) bool {
	clean := strings.Trim(pkg, "/")
	return clean == "internal/eval" || strings.HasSuffix(clean, "/eval")
}
