package scoping

import (
	"fmt"
	"sort"
	"strings"

	"gradientcontrol.com/omega/core/tool/analyze-deps/internal/graph"
)

// ScopingReport contains diagnostic recommendations for narrowing an internal package.
type ScopingReport struct {
	Pkg                string
	CanNarrow          bool
	PinnedReason       string
	SuggestedSubsystem string
	SuggestedDest      string
	Subsystems         []string
	ImportersBySubsys  map[string][]string
	Blockers           []string
	Advice             string
}

// AnalyzePackage evaluates whether a package in root internal/ can be scoped more narrowly.
func AnalyzePackage(g *graph.Graph, pkg string) *ScopingReport {
	cleanPkg := strings.Trim(pkg, "/")
	node := g.Packages[cleanPkg]

	report := &ScopingReport{
		Pkg:               cleanPkg,
		ImportersBySubsys: make(map[string][]string),
	}

	if node == nil {
		report.Advice = fmt.Sprintf("package %s not found in module", cleanPkg)
		return report
	}

	// Gather unique external importers (ignoring self-imports)
	seen := make(map[string]bool)
	subsysMap := make(map[string]map[string]bool)

	for _, edge := range node.ImportedBy {
		fromPkg := edge.FromPkg
		if fromPkg == cleanPkg || strings.HasPrefix(fromPkg, cleanPkg+"/") {
			continue // Self or child package import
		}
		if seen[fromPkg] {
			continue
		}
		seen[fromPkg] = true

		subsys := SubsystemForScoping(fromPkg)
		if subsysMap[subsys] == nil {
			subsysMap[subsys] = make(map[string]bool)
		}
		subsysMap[subsys][fromPkg] = true
	}

	var subsystems []string
	for subsys, pkgs := range subsysMap {
		subsystems = append(subsystems, subsys)
		var list []string
		for p := range pkgs {
			list = append(list, p)
		}
		sort.Strings(list)
		report.ImportersBySubsys[subsys] = list
	}
	sort.Strings(subsystems)
	report.Subsystems = subsystems

	baseName := cleanPkg
	if strings.HasPrefix(cleanPkg, "internal/") {
		baseName = strings.TrimPrefix(cleanPkg, "internal/")
	}

	if len(subsystems) == 0 {
		report.CanNarrow = true
		report.SuggestedSubsystem = "unreferenced"
		report.Advice = fmt.Sprintf("package %s has zero external importers (potentially dead code or candidate for removal)", cleanPkg)
		return report
	}

	if node.KeepHereReason != "" {
		report.CanNarrow = false
		report.PinnedReason = node.KeepHereReason
		report.Blockers = append(report.Blockers, fmt.Sprintf("pinned (%s)", node.KeepHereReason))
		report.Advice = fmt.Sprintf("Pinned in root internal/ because %s.", node.KeepHereReason)
		return report
	}

	if len(subsystems) == 1 {
		singleSubsys := subsystems[0]
		report.CanNarrow = true
		report.SuggestedSubsystem = singleSubsys
		switch singleSubsys {
		case "assembler":
			report.SuggestedDest = "internal/assembler/" + baseName
		case "parser":
			report.SuggestedDest = "parser/internal/" + baseName
		case "backend":
			// Check if all importers are from a specific backend target
			specificTarget := detectSpecificBackendTarget(report.ImportersBySubsys["backend"])
			if specificTarget != "" {
				report.SuggestedDest = specificTarget + "/internal/" + baseName
			} else {
				report.SuggestedDest = "backend/internal/" + baseName
			}
		default:
			report.SuggestedDest = singleSubsys + "/internal/" + baseName
		}
		report.Advice = fmt.Sprintf("Ready for narrowing: all %d importers belong to %s. Move to %s",
			len(seen), singleSubsys, report.SuggestedDest)
		return report
	}

	// Multiple subsystems import this package
	report.CanNarrow = false
	for _, s := range subsystems {
		report.Blockers = append(report.Blockers, fmt.Sprintf("%s (%d packages: %s)", s, len(report.ImportersBySubsys[s]), strings.Join(report.ImportersBySubsys[s], ", ")))
	}
	report.Advice = fmt.Sprintf("Cannot narrow directly: imported across %d subsystems: %s. To narrow, decouple callers via public facade or interface extraction.",
		len(subsystems), strings.Join(subsystems, ", "))

	return report
}

// AuditAll evaluates all packages in root internal/ and returns reports for narrowing candidates.
func AuditAll(g *graph.Graph) []*ScopingReport {
	var results []*ScopingReport

	var pkgs []string
	for p := range g.Packages {
		if isRootInternal(p) {
			pkgs = append(pkgs, p)
		}
	}
	sort.Strings(pkgs)

	for _, p := range pkgs {
		rep := AnalyzePackage(g, p)
		results = append(results, rep)
	}

	return results
}

func isRootInternal(pkg string) bool {
	clean := strings.Trim(pkg, "/")
	if !strings.HasPrefix(clean, "internal/") {
		return false
	}
	parts := strings.Split(clean, "/")
	// Packages like internal/assembler/* are already scoped to assembler,
	// and nested */internal/* packages are already privately scoped to their parent.
	if len(parts) >= 2 && parts[1] == "assembler" {
		return false
	}
	for _, part := range parts[1:] {
		if part == "internal" {
			return false
		}
	}
	return true
}

// SubsystemForScoping normalizes importer package to its primary subsystem.
func SubsystemForScoping(pkg string) string {
	clean := strings.Trim(pkg, "/")
	if strings.HasPrefix(clean, "parser") {
		return "parser"
	}
	if strings.HasPrefix(clean, "assembler") || strings.HasPrefix(clean, "internal/assembler") {
		return "assembler"
	}
	if strings.HasPrefix(clean, "backend") {
		return "backend"
	}
	if clean == "eval" || strings.HasPrefix(clean, "eval/") {
		return "eval"
	}
	if clean == "program" || strings.HasPrefix(clean, "program/") {
		return "program"
	}
	if clean == "stdlib" || strings.HasPrefix(clean, "stdlib/") {
		return "stdlib"
	}
	if clean == "cmd" || strings.HasPrefix(clean, "cmd/") {
		return "cmd"
	}
	if clean == "tool" || strings.HasPrefix(clean, "tool/") {
		return "tool"
	}
	return "internal"
}

func detectSpecificBackendTarget(importers []string) string {
	if len(importers) == 0 {
		return ""
	}
	var target string
	for _, imp := range importers {
		var cur string
		for _, t := range []string{"backend/wgsl", "backend/audio", "backend/mesh", "backend/tape"} {
			if strings.HasPrefix(imp, t) {
				cur = t
				break
			}
		}
		if cur == "" {
			return "" // Non-target backend importer (e.g. backend/common or backend/internal)
		}
		if target == "" {
			target = cur
		} else if target != cur {
			return "" // Multiple different backend targets
		}
	}
	return target
}
