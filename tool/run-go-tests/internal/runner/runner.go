package runner

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// SuiteResult holds the timing and pass/fail outcome of a single Go module test run.
type SuiteResult struct {
	Path    string
	Name    string
	Seconds float64
	Passed  bool
}

// RunResult aggregates the outcomes across all executed Go module test suites.
type RunResult struct {
	Results  []SuiteResult
	Failures []string
	TotalSec float64
}

// GoldenDiff describes a golden file whose byte size differs from main (or was added/removed).
type GoldenDiff struct {
	Path        string `json:"path"`
	MainBytes   int64  `json:"main_bytes"`
	ActualBytes int64  `json:"actual_bytes"`
	Status      string `json:"status,omitempty"` // "modified", "added", "removed"
}

// GoldenCategorySummary summarizes compiled golden sizes for one backend category (WGSL or Fidget GLSL).
type GoldenCategorySummary struct {
	FileCount   int
	ActualBytes int64
	MainBytes   int64
	HasMain     bool
	Diffs       []GoldenDiff
}

// GoldensReport holds the compiled size summaries for WGSL and Fidget GLSL testdata goldens.
type GoldensReport struct {
	WGSL       GoldenCategorySummary
	FidgetGLSL GoldenCategorySummary
}

// RunSuites executes `go test -count=1 ./...` in each directory path and records per-suite timings.
func RunSuites(repoRoot string, paths []string, stdout, stderr io.Writer) RunResult {
	results := make([]SuiteResult, 0, len(paths))
	var failures []string
	var totalSec float64

	for _, dir := range paths {
		absDir, err := filepath.Abs(dir)
		if err != nil {
			absDir = dir
		}
		suiteName := SuiteDisplayName(absDir, repoRoot)

		fmt.Fprintln(stdout, "")
		fmt.Fprintln(stdout, absDir)
		fmt.Fprintln(stdout, "go test -count=1 ./...")

		start := time.Now()
		cmd := exec.Command("go", "test", "-count=1", "./...")
		cmd.Dir = absDir
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		runErr := cmd.Run()
		elapsed := Round2(time.Since(start).Seconds())
		totalSec += elapsed

		passed := runErr == nil
		statusStr := "ok"
		if !passed {
			statusStr = "FAIL"
			failures = append(failures, fmt.Sprintf("cd %s; go test -count=1 ./...", dir))
		}
		fmt.Fprintf(stdout, "  -> %s: %.2fs (%s)\n", suiteName, elapsed, statusStr)

		results = append(results, SuiteResult{
			Path:    dir,
			Name:    suiteName,
			Seconds: elapsed,
			Passed:  passed,
		})
	}

	return RunResult{
		Results:  results,
		Failures: failures,
		TotalSec: Round2(totalSec),
	}
}

// SuiteDisplayName formats absDir relative to repoRoot when inside repoRoot.
func SuiteDisplayName(absDir, repoRoot string) string {
	if repoRoot != "" {
		if rel, err := filepath.Rel(repoRoot, absDir); err == nil && !strings.HasPrefix(rel, "..") {
			return filepath.ToSlash(rel)
		}
	}
	return filepath.ToSlash(absDir)
}

// CollectGoldenSizes scans WGSL and Fidget GLSL testdata directories and compares their sizes against git main.
func CollectGoldenSizes(repoRoot, wgslTestdataRel, fidgetTestdataRel string) GoldensReport {
	if repoRoot == "" {
		return GoldensReport{}
	}
	var relPaths []string
	if wgslTestdataRel != "" {
		relPaths = append(relPaths, wgslTestdataRel)
	}
	if fidgetTestdataRel != "" {
		relPaths = append(relPaths, fidgetTestdataRel)
	}
	if len(relPaths) == 0 {
		return GoldensReport{}
	}

	mainSizes, hasMain := gitMainBlobSizes(repoRoot, relPaths...)

	var report GoldensReport
	if wgslTestdataRel != "" {
		wgslDisk := scanGoldenFiles(repoRoot, wgslTestdataRel, ".wgsl", "")
		wgslMain := filterPrefixAndSuffix(mainSizes, wgslTestdataRel+"/", ".wgsl", "")
		report.WGSL = CompareGoldenCategory(wgslDisk, wgslMain, hasMain)
	}
	if fidgetTestdataRel != "" {
		fidgetDisk := scanGoldenFiles(repoRoot, fidgetTestdataRel, ".glsl", ".actual.glsl")
		fidgetMain := filterPrefixAndSuffix(mainSizes, fidgetTestdataRel+"/", ".glsl", ".actual.glsl")
		report.FidgetGLSL = CompareGoldenCategory(fidgetDisk, fidgetMain, hasMain)
	}
	return report
}

// CompareGoldenCategory computes total bytes and per-file diffs between current disk files and main.
func CompareGoldenCategory(diskSizes, mainSizes map[string]int64, hasMain bool) GoldenCategorySummary {
	var summary GoldenCategorySummary
	summary.FileCount = len(diskSizes)
	summary.HasMain = hasMain

	seen := make(map[string]bool, len(diskSizes)+len(mainSizes))
	for k, v := range diskSizes {
		summary.ActualBytes += v
		seen[k] = true
	}
	for k, v := range mainSizes {
		summary.MainBytes += v
		seen[k] = true
	}

	if !hasMain {
		return summary
	}

	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		act, inDisk := diskSizes[k]
		mainB, inMain := mainSizes[k]
		switch {
		case inDisk && inMain && act != mainB:
			summary.Diffs = append(summary.Diffs, GoldenDiff{
				Path:        k,
				MainBytes:   mainB,
				ActualBytes: act,
				Status:      "modified",
			})
		case inDisk && !inMain:
			summary.Diffs = append(summary.Diffs, GoldenDiff{
				Path:        k,
				MainBytes:   0,
				ActualBytes: act,
				Status:      "added",
			})
		case !inDisk && inMain:
			summary.Diffs = append(summary.Diffs, GoldenDiff{
				Path:        k,
				MainBytes:   mainB,
				ActualBytes: 0,
				Status:      "removed",
			})
		}
	}
	return summary
}

func scanGoldenFiles(repoRoot, relDir, ext, actualExt string) map[string]int64 {
	out := make(map[string]int64)
	rootDir := filepath.Join(repoRoot, relDir)
	_ = filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		name := d.Name()
		if !strings.HasSuffix(name, ext) {
			return nil
		}
		if actualExt != "" && strings.HasSuffix(name, actualExt) {
			return nil
		}
		measurePath := path
		if actualExt != "" {
			actualCandidate := strings.TrimSuffix(path, ext) + actualExt
			if info, aErr := os.Stat(actualCandidate); aErr == nil && !info.IsDir() {
				measurePath = actualCandidate
			}
		}
		info, sErr := os.Stat(measurePath)
		if sErr != nil {
			return nil
		}
		rel, rErr := filepath.Rel(repoRoot, path)
		if rErr != nil {
			rel = path
		}
		out[filepath.ToSlash(rel)] = info.Size()
		return nil
	})
	return out
}

func filterPrefixAndSuffix(all map[string]int64, prefix, suffix, excludeSuffix string) map[string]int64 {
	out := make(map[string]int64)
	for k, v := range all {
		if !strings.HasPrefix(k, prefix) || !strings.HasSuffix(k, suffix) {
			continue
		}
		if excludeSuffix != "" && strings.HasSuffix(k, excludeSuffix) {
			continue
		}
		out[k] = v
	}
	return out
}

func gitMainBlobSizes(repoRoot string, relPaths ...string) (map[string]int64, bool) {
	args := append([]string{"ls-tree", "-r", "-l", "main", "--"}, relPaths...)
	cmd := exec.Command("git", args...)
	cmd.Dir = repoRoot
	raw, err := cmd.Output()
	if err != nil {
		return nil, false
	}
	return ParseGitLsTreeSizes(raw), true
}

// ParseGitLsTreeSizes parses `git ls-tree -r -l` output into a map of relative path -> byte size.
func ParseGitLsTreeSizes(raw []byte) map[string]int64 {
	sizes := make(map[string]int64)
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	for scanner.Scan() {
		line := scanner.Text()
		tabIdx := strings.IndexByte(line, '\t')
		if tabIdx <= 0 {
			continue
		}
		meta := strings.Fields(line[:tabIdx])
		path := strings.TrimSpace(line[tabIdx+1:])
		if len(meta) < 4 || path == "" {
			continue
		}
		sz, err := strconv.ParseInt(meta[3], 10, 64)
		if err != nil {
			continue
		}
		sizes[filepath.ToSlash(path)] = sz
	}
	return sizes
}

// GitBranchAndCommit returns the current git branch and short commit hash for dir.
func GitBranchAndCommit(dir string) (string, string) {
	workDir := dir
	if workDir == "" {
		workDir = "."
	}
	branch := runGit(workDir, "rev-parse", "--abbrev-ref", "HEAD")
	commit := runGit(workDir, "rev-parse", "--short", "HEAD")
	return nonEmpty(branch, "unknown"), nonEmpty(commit, "unknown")
}

func runGit(dir string, args ...string) string {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func nonEmpty(s, fallback string) string {
	if s == "" {
		return fallback
	}
	return s
}

// Round2 rounds v to 2 decimal places.
func Round2(v float64) float64 {
	return math.Round(v*100) / 100
}
