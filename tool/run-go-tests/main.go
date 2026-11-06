package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"gradientcontrol.com/omega/core/tool/run-go-tests/internal/config"
	"gradientcontrol.com/omega/core/tool/run-go-tests/internal/runner"
	"gradientcontrol.com/omega/core/tool/run-go-tests/internal/summary"
)

func main() {
	cfg, err := config.ParseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	history := summary.LoadHistory(cfg.HistoryFile)

	runRes := runner.RunSuites(cfg.OmegaRoot, cfg.Paths, os.Stdout, os.Stderr)
	goldens := runner.CollectGoldenSizes(cfg.OmegaRoot)

	fmt.Println("")
	summary.PrintSuiteTable(os.Stdout, runRes.Results, history, runRes.TotalSec, goldens)

	allPassed := len(runRes.Failures) == 0
	summaryText := summary.FormatTotalComparison(runRes.TotalSec, goldens, history, cfg.HistoryFile, allPassed)

	if allPassed && cfg.HistoryFile != "" {
		branch, commit := runner.GitBranchAndCommit(cfg.OmegaRoot)
		rec := summary.BuildHistoryRecord(time.Now(), branch, commit, runRes.TotalSec, runRes.Results, goldens)
		if err := summary.AppendHistory(cfg.HistoryFile, rec); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to record test timing history to %s: %v\n", cfg.HistoryFile, err)
		}
	}

	if cfg.SummaryFile != "" {
		_ = os.WriteFile(cfg.SummaryFile, []byte(summaryText), 0o644)
	} else {
		fmt.Print("\n" + summaryText)
	}
	if cfg.FailuresFile != "" && len(runRes.Failures) > 0 {
		_ = os.WriteFile(cfg.FailuresFile, []byte(strings.Join(runRes.Failures, "\n")+"\n"), 0o644)
	}

	if !allPassed {
		os.Exit(1)
	}
}

