package summary

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gradientcontrol.com/omega/core/tool/run-go-tests/internal/runner"
)

func TestHistoryRoundTripAndSummaryFormatting(t *testing.T) {
	tmpDir := t.TempDir()
	histFile := filepath.Join(tmpDir, "test-local.jsonl")

	goldens1 := runner.GoldensReport{
		WGSL:       runner.GoldenCategorySummary{FileCount: 214, ActualBytes: 1500000, MainBytes: 1500000, HasMain: true},
		FidgetGLSL: runner.GoldenCategorySummary{FileCount: 34, ActualBytes: 300000, MainBytes: 300000, HasMain: true},
	}

	rec1 := BuildHistoryRecord(time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC), "main", "abc1234", 50.0, []runner.SuiteResult{
		{Name: "agent/assembler", Seconds: 10.0, Passed: true},
		{Name: "webai/agent/canvas", Seconds: 40.0, Passed: true},
	}, goldens1)
	if err := AppendHistory(histFile, rec1); err != nil {
		t.Fatalf("AppendHistory(rec1): %v", err)
	}

	rec2 := BuildHistoryRecord(time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC), "feat", "def5678", 60.0, []runner.SuiteResult{
		{Name: "agent/assembler", Seconds: 12.0, Passed: true},
		{Name: "webai/agent/canvas", Seconds: 48.0, Passed: true},
	}, goldens1)
	if err := AppendHistory(histFile, rec2); err != nil {
		t.Fatalf("AppendHistory(rec2): %v", err)
	}

	loaded := LoadHistory(histFile)
	if len(loaded) != 2 {
		t.Fatalf("len(loaded) = %d, want 2", len(loaded))
	}

	current := []runner.SuiteResult{
		{Name: "agent/assembler", Seconds: 11.0, Passed: true},
		{Name: "webai/agent/canvas", Seconds: 43.0, Passed: true},
	}
	goldensCurr := runner.GoldensReport{
		WGSL: runner.GoldenCategorySummary{
			FileCount:   214,
			ActualBytes: 1499500,
			MainBytes:   1500000,
			HasMain:     true,
			Diffs: []runner.GoldenDiff{{
				Path:        "agent/wgsl/internal/codegen/testdata/bike.wgsl",
				MainBytes:   14500,
				ActualBytes: 14000,
				Status:      "modified",
			}},
		},
		FidgetGLSL: runner.GoldenCategorySummary{
			FileCount:   34,
			ActualBytes: 300000,
			MainBytes:   300000,
			HasMain:     true,
		},
	}

	var tableBuf bytes.Buffer
	PrintSuiteTable(&tableBuf, current, loaded, 54.0, goldensCurr, "agent/wgsl/internal/codegen/testdata", "agent/fidget/internal/convert/testdata")
	tableStr := tableBuf.String()
	if !strings.Contains(tableStr, "agent/assembler") || !strings.Contains(tableStr, "-1.00s") {
		t.Fatalf("unexpected suite table:\n%s", tableStr)
	}
	if !strings.Contains(tableStr, "bike.wgsl") || !strings.Contains(tableStr, "-500 B") {
		t.Fatalf("expected golden diff in table:\n%s", tableStr)
	}

	summaryStr := FormatTotalComparison(54.0, goldensCurr, loaded, histFile, true)
	if !strings.Contains(summaryStr, "feat@def5678, 60.00s): -6.00s (-10.0%)") {
		t.Errorf("summary missing previous run comparison:\n%s", summaryStr)
	}
	if !strings.Contains(summaryStr, "2 passing runs: avg 55.00s, min 50.00s, max 60.00s): -1.00s (-1.8%)") {
		t.Errorf("summary missing historical avg/min/max comparison:\n%s", summaryStr)
	}
	if !strings.Contains(summaryStr, "WGSL codegen (214 files): 1499500 B (-500 B (-0.03%) vs main, 1 file(s) changed)") {
		t.Errorf("summary missing WGSL golden comparison:\n%s", summaryStr)
	}
	if !strings.Contains(summaryStr, "Fidget GLSL (34 files):  300000 B (unchanged vs main)") {
		t.Errorf("summary missing Fidget GLSL golden comparison:\n%s", summaryStr)
	}
}
