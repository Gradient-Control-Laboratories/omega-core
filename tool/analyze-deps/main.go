package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"gradientcontrol.com/omega/core/tool/analyze-deps/internal/config"
	"gradientcontrol.com/omega/core/tool/analyze-deps/internal/graph"
	"gradientcontrol.com/omega/core/tool/analyze-deps/internal/rules"
	"gradientcontrol.com/omega/core/tool/analyze-deps/internal/scoping"
)

func main() {
	cfg, err := config.ParseFlags(os.Args[1:])
	if err != nil {
		if errors.Is(err, flag.ErrHelp) {
			os.Exit(0)
		}
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(2)
	}

	g, err := graph.Build(cfg.RootDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error building dependency graph: %v\n", err)
		os.Exit(1)
	}

	exitCode := 0

	// 1. Single package narrowing analysis
	if cfg.NarrowPkg != "" {
		report := scoping.AnalyzePackage(g, cfg.NarrowPkg)
		printScopingReport(report, cfg.Verbose)
		return
	}

	// 2. All packages narrowing audit
	if cfg.AllNarrow {
		reports := scoping.AuditAll(g)
		printAllScopingAudit(reports, cfg.Verbose)
		return
	}

	// 3. ASCII dependency tree / subsystem graph
	if cfg.Graph {
		printDependencyGraph(g, cfg.Verbose)
		return
	}

	// 4. Architectural Rules Audit
	if cfg.Rules {
		violations := rules.Check(g)
		if len(violations) == 0 {
			fmt.Printf("✓ Architectural Integrity Check: All packages passed 3 invariants (layer-hierarchy, pass-separation, framework-creep).\n")
		} else {
			prodCount := 0
			testCount := 0
			for _, v := range violations {
				if v.IsTest {
					testCount++
				} else {
					prodCount++
				}
			}

			fmt.Fprintf(os.Stderr, "Architectural Invariant Violations (%d total: %d prod, %d test):\n\n", len(violations), prodCount, testCount)
			for i, v := range violations {
				kind := "PROD"
				if v.IsTest {
					kind = "TEST"
				}
				fmt.Fprintf(os.Stderr, "  %d. [%s] [%s] %s\n", i+1, kind, v.Rule, v.Description)
				fmt.Fprintf(os.Stderr, "     File: %s:%d\n", v.FilePath, v.Line)
				fmt.Fprintf(os.Stderr, "     Import: %s -> %s\n\n", v.FromPkg, v.ToPkg)
			}
			exitCode = 1
		}
	}

	os.Exit(exitCode)
}

func printScopingReport(r *scoping.ScopingReport, verbose bool) {
	fmt.Printf("Scoping Analysis for: %s\n", r.Pkg)
	fmt.Printf("Status: ")
	switch {
	case r.CanNarrow:
		fmt.Printf("READY FOR NARROWING\n")
		fmt.Printf("Suggested Destination: %s\n", r.SuggestedDest)
	case r.PinnedReason != "":
		fmt.Printf("PINNED IN ROOT internal/\n")
	default:
		fmt.Printf("BLOCKED FROM NARROWING (Shared across multiple subsystems)\n")
	}
	fmt.Printf("Advice: %s\n\n", r.Advice)

	if len(r.Subsystems) > 0 {
		fmt.Printf("Importers by Subsystem (%d subsystems):\n", len(r.Subsystems))
		for _, s := range r.Subsystems {
			pkgs := r.ImportersBySubsys[s]
			fmt.Printf("  • %-12s (%d callers): %s\n", s, len(pkgs), strings.Join(pkgs, ", "))
		}
	}
}

func printAllScopingAudit(reports []*scoping.ScopingReport, verbose bool) {
	var ready []*scoping.ScopingReport
	var pinned []*scoping.ScopingReport
	var shared []*scoping.ScopingReport

	for _, r := range reports {
		switch {
		case r.CanNarrow:
			ready = append(ready, r)
		case r.PinnedReason != "":
			pinned = append(pinned, r)
		default:
			shared = append(shared, r)
		}
	}

	fmt.Printf("=== Package Scoping Audit for root internal/ (%d total packages) ===\n\n", len(reports))

	fmt.Printf("Candidates Ready for Narrower Scoping (%d):\n", len(ready))
	if len(ready) == 0 {
		fmt.Printf("  (None — all internal packages are shared across multiple subsystems, pinned, or already scoped)\n")
	} else {
		for _, r := range ready {
			fmt.Printf("  • %-24s -> %-36s (Subsystem: %s)\n", r.Pkg, r.SuggestedDest, r.SuggestedSubsystem)
		}
	}
	fmt.Println()

	if len(pinned) > 0 {
		fmt.Printf("Packages Pinned in Root internal/ (%d):\n", len(pinned))
		for _, r := range pinned {
			fmt.Printf("  • %-24s imported by: %s (because %s)\n", r.Pkg, strings.Join(r.Subsystems, ", "), r.PinnedReason)
			if verbose {
				for _, s := range r.Subsystems {
					pkgs := r.ImportersBySubsys[s]
					fmt.Printf("      - %s: %s\n", s, strings.Join(pkgs, ", "))
				}
			}
		}
		fmt.Println()
	}

	fmt.Printf("Packages Shared Across Multiple Subsystems (%d):\n", len(shared))
	for _, r := range shared {
		fmt.Printf("  • %-24s imported by: %s\n", r.Pkg, strings.Join(r.Subsystems, ", "))
		if verbose {
			for _, s := range r.Subsystems {
				pkgs := r.ImportersBySubsys[s]
				fmt.Printf("      - %s: %s\n", s, strings.Join(pkgs, ", "))
			}
		}
	}
}

func printDependencyGraph(g *graph.Graph, verbose bool) {
	if verbose {
		fmt.Println("=== Subsystem Dependency Relationships (Production + Test Imports) ===")
	} else {
		fmt.Println("=== Subsystem Dependency Relationships (Production Imports) ===")
	}
	subsysDeps := make(map[string]map[string]int)

	for _, node := range g.Packages {
		fromSubsys := graph.SubsystemOf(node.ImportPath)
		if subsysDeps[fromSubsys] == nil {
			subsysDeps[fromSubsys] = make(map[string]int)
		}
		for _, edge := range node.Imports {
			if edge.IsTest && !verbose {
				continue
			}
			if strings.HasPrefix(edge.RawImport, g.ModulePrefix) {
				toSubsys := graph.SubsystemOf(edge.ToPkg)
				subsysDeps[fromSubsys][toSubsys]++
			}
		}
	}

	var subsystems []string
	for s := range subsysDeps {
		subsystems = append(subsystems, s)
	}
	sort.Strings(subsystems)

	for _, s := range subsystems {
		fmt.Printf("%s\n", s)
		targets := subsysDeps[s]
		var targetList []string
		for t := range targets {
			targetList = append(targetList, t)
		}
		sort.Strings(targetList)
		for _, t := range targetList {
			count := targets[t]
			fmt.Printf("  └──> %-16s (%d imports)\n", t, count)
		}
	}
}
