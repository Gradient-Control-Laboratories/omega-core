package summary

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gradientcontrol.com/omega/core/tool/run-go-tests/internal/runner"
)

// HistoryRecord represents one passing test run stored in test-local.jsonl.
type HistoryRecord struct {
	Timestamp            string              `json:"timestamp"`
	Branch               string              `json:"branch"`
	Commit               string              `json:"commit"`
	TotalGoSec           float64             `json:"total_go_sec"`
	Suites               map[string]float64  `json:"suites"`
	WGSLTotalBytes       int64               `json:"wgsl_total_bytes,omitempty"`
	FidgetGLSLTotalBytes int64               `json:"fidget_glsl_total_bytes,omitempty"`
	ChangedGoldens       []runner.GoldenDiff `json:"changed_goldens,omitempty"`
}

// LoadHistory reads prior passing test run records from a JSONL file.
func LoadHistory(path string) []HistoryRecord {
	if path == "" {
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()

	var out []HistoryRecord
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var rec HistoryRecord
		if err := json.Unmarshal([]byte(line), &rec); err == nil && rec.TotalGoSec > 0 {
			out = append(out, rec)
		}
	}
	return out
}

// AppendHistory appends a passing run record to the JSONL history file.
func AppendHistory(path string, rec HistoryRecord) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	data, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	_, err = f.Write(append(data, '\n'))
	return err
}

// BuildHistoryRecord constructs a HistoryRecord for the completed run.
func BuildHistoryRecord(now time.Time, branch, commit string, totalSec float64, results []runner.SuiteResult, goldens runner.GoldensReport) HistoryRecord {
	suites := make(map[string]float64, len(results))
	for _, r := range results {
		suites[r.Name] = r.Seconds
	}
	var changed []runner.GoldenDiff
	changed = append(changed, goldens.WGSL.Diffs...)
	changed = append(changed, goldens.FidgetGLSL.Diffs...)

	return HistoryRecord{
		Timestamp:            now.Format(time.RFC3339),
		Branch:               branch,
		Commit:               commit,
		TotalGoSec:           totalSec,
		Suites:               suites,
		WGSLTotalBytes:       goldens.WGSL.ActualBytes,
		FidgetGLSLTotalBytes: goldens.FidgetGLSL.ActualBytes,
		ChangedGoldens:       changed,
	}
}

// PrintSuiteTable writes the per-suite timing table and golden size diffs (if any) to w.
func PrintSuiteTable(w io.Writer, results []runner.SuiteResult, history []HistoryRecord, totalSec float64, goldens runner.GoldensReport) {
	var prevSuites map[string]float64
	if len(history) > 0 {
		prevSuites = history[len(history)-1].Suites
	}

	fmt.Fprintln(w, "==========================================================================")
	fmt.Fprintln(w, "Omega Go Test Suite Timings")
	fmt.Fprintln(w, "--------------------------------------------------------------------------")
	fmt.Fprintf(w, "%-32s %9s %10s %10s  %s\n", "Suite", "Time", "vs Prev", "vs Avg", "Status")
	fmt.Fprintln(w, "--------------------------------------------------------------------------")

	for _, r := range results {
		var prevPtr *float64
		if prevSuites != nil {
			if v, ok := prevSuites[r.Name]; ok {
				val := v
				prevPtr = &val
			}
		}
		avgPtr := suiteAverage(r.Name, history)
		status := "ok"
		if !r.Passed {
			status = "FAIL"
		}
		fmt.Fprintf(w, "%-32s %8.2fs %10s %10s  %s\n",
			r.Name, r.Seconds, formatDelta(r.Seconds, prevPtr), formatDelta(r.Seconds, avgPtr), status)
	}
	fmt.Fprintln(w, "--------------------------------------------------------------------------")
	fmt.Fprintf(w, "%-32s %8.2fs\n", "TOTAL Go Tests", totalSec)

	allDiffs := append([]runner.GoldenDiff(nil), goldens.WGSL.Diffs...)
	allDiffs = append(allDiffs, goldens.FidgetGLSL.Diffs...)
	if len(allDiffs) > 0 {
		fmt.Fprintln(w, "--------------------------------------------------------------------------")
		fmt.Fprintln(w, "Compiled Golden Size Changes vs main")
		fmt.Fprintf(w, "%-44s %9s %9s %16s\n", "File", "main", "Current", "Delta")
		for _, d := range allDiffs {
			shortPath := d.Path
			for _, prefix := range []string{
				runner.WGSLTestdataRel + "/",
				runner.OmegaCoreWGSLTestdataRel + "/",
				runner.FidgetTestdataRel + "/",
				runner.OmegaCoreFidgetTestdataRel + "/",
			} {
				if trimmed := strings.TrimPrefix(shortPath, prefix); trimmed != shortPath {
					shortPath = trimmed
					break
				}
			}
			fmt.Fprintf(w, "%-44s %8dB %8dB %16s\n",
				shortPath, d.MainBytes, d.ActualBytes, formatBytesDeltaPct(d.ActualBytes, d.MainBytes))
		}
	}
	fmt.Fprintln(w, "==========================================================================")
}

// FormatTotalComparison formats the end-of-script duration and golden-size comparison block.
func FormatTotalComparison(totalSec float64, goldens runner.GoldensReport, history []HistoryRecord, historyPath string, allPassed bool) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Go tests duration: %.2fs\n", totalSec)

	if len(history) > 0 {
		prev := history[len(history)-1]
		meta := fmt.Sprintf("%s@%s", nonEmpty(prev.Branch, "?"), nonEmpty(prev.Commit, "?"))
		fmt.Fprintf(&b, "  vs previous run (%s, %.2fs): %s\n",
			meta, prev.TotalGoSec, formatDeltaPct(totalSec, prev.TotalGoSec))

		var sum float64
		minVal := history[0].TotalGoSec
		maxVal := history[0].TotalGoSec
		for _, h := range history {
			sum += h.TotalGoSec
			if h.TotalGoSec < minVal {
				minVal = h.TotalGoSec
			}
			if h.TotalGoSec > maxVal {
				maxVal = h.TotalGoSec
			}
		}
		avgVal := sum / float64(len(history))
		fmt.Fprintf(&b, "  vs history (%d passing runs: avg %.2fs, min %.2fs, max %.2fs): %s\n",
			len(history), avgVal, minVal, maxVal, formatDeltaPct(totalSec, avgVal))
	} else {
		fmt.Fprintln(&b, "  vs history: baseline run (no prior passing runs recorded yet)")
	}

	if goldens.WGSL.FileCount > 0 || goldens.FidgetGLSL.FileCount > 0 {
		fmt.Fprintf(&b, "Compiled golden sizes:\n")
		if goldens.WGSL.FileCount > 0 {
			fmt.Fprintf(&b, "  WGSL codegen (%d files): %d B (%s)\n",
				goldens.WGSL.FileCount, goldens.WGSL.ActualBytes, formatCategoryVsMain(goldens.WGSL))
		}
		if goldens.FidgetGLSL.FileCount > 0 {
			fmt.Fprintf(&b, "  Fidget GLSL (%d files):  %d B (%s)\n",
				goldens.FidgetGLSL.FileCount, goldens.FidgetGLSL.ActualBytes, formatCategoryVsMain(goldens.FidgetGLSL))
		}
	}

	if allPassed && historyPath != "" {
		fmt.Fprintf(&b, "  timing & sizes recorded to: %s\n", historyPath)
	} else if !allPassed {
		fmt.Fprintln(&b, "  timing not recorded (one or more Go test suites failed)")
	}
	return b.String()
}

func formatCategoryVsMain(cat runner.GoldenCategorySummary) string {
	if !cat.HasMain {
		return "main baseline unavailable"
	}
	if len(cat.Diffs) == 0 && cat.ActualBytes == cat.MainBytes {
		return "unchanged vs main"
	}
	return fmt.Sprintf("%s vs main, %d file(s) changed",
		formatBytesDeltaPct(cat.ActualBytes, cat.MainBytes), len(cat.Diffs))
}

func suiteAverage(name string, history []HistoryRecord) *float64 {
	var sum float64
	var count int
	for _, h := range history {
		if h.Suites == nil {
			continue
		}
		if v, ok := h.Suites[name]; ok {
			sum += v
			count++
		}
	}
	if count == 0 {
		return nil
	}
	avg := sum / float64(count)
	return &avg
}

func formatDelta(curr float64, ref *float64) string {
	if ref == nil {
		return "n/a"
	}
	diff := curr - *ref
	sign := "+"
	if diff < 0 {
		sign = ""
	}
	return fmt.Sprintf("%s%.2fs", sign, diff)
}

func formatDeltaPct(curr, ref float64) string {
	if ref <= 0 {
		return "n/a"
	}
	diff := curr - ref
	pct := (diff / ref) * 100.0
	sign := "+"
	if diff < 0 {
		sign = ""
	}
	return fmt.Sprintf("%s%.2fs (%s%.1f%%)", sign, diff, sign, pct)
}

func formatBytesDeltaPct(curr, ref int64) string {
	diff := curr - ref
	sign := "+"
	if diff < 0 {
		sign = ""
	}
	if ref <= 0 {
		return fmt.Sprintf("%s%d B", sign, diff)
	}
	pct := (float64(diff) / float64(ref)) * 100.0
	return fmt.Sprintf("%s%d B (%s%.2f%%)", sign, diff, sign, pct)
}

func nonEmpty(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}
