package fix

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

var (
	sceneSizeLineRE  = regexp.MustCompile(`(?i)^\s*//\s*Scene\s+Size:[^\S\n]*(.*)$`)
	sceneGridLineRE  = regexp.MustCompile(`(?i)^\s*//\s*Scene\s+Grid:[^\S\n]*(.*)$`)
	mmGridRE         = regexp.MustCompile(`(?i)^\s*[-+0-9.eE]+\s*mm\b`)
	sceneDirectiveRE = regexp.MustCompile(`^\s*@scene\s*\(`)
)

var sceneFixer = Fixer{
	Name: "scene",
	Doc:  "replace legacy // Scene Size and // Scene Grid comments with @scene directive",
	Help: `The scene fixer replaces legacy GSL v1 "// Scene Size:" and "// Scene Grid:"
header comments with the top-level "@scene(...)" directive.

When the resolved grid spacing equals the default (size / 10), single-argument
"@scene(size)" is emitted; otherwise "@scene(size, grid)" is emitted.

Before:
  // Scene Size: 20
  // Scene Grid: 10

  return sphere(1);

After:
  @scene(20)

  return sphere(1);
`,
	Apply: applySceneFix,
}

func applySceneFix(src string) (string, bool, error) {
	if src == "" {
		return src, false, nil
	}

	useCRLF := strings.Contains(src, "\r\n")
	normalized := src
	if useCRLF {
		normalized = strings.ReplaceAll(src, "\r\n", "\n")
	}

	lines := strings.Split(normalized, "\n")
	hasDirective := hasTopLevelSceneDirective(lines)

	sizeVal, hasSize, gridVal, hasGrid := parseLegacySceneComments(lines)
	hasAnySceneComment := hasLegacySceneCommentLine(lines)

	if hasDirective {
		if !hasAnySceneComment {
			return src, false, nil
		}
		cleaned := removeSceneCommentLines(lines)
		out := joinAndRestoreEndings(cleaned, useCRLF)
		return out, out != src, nil
	}

	if !hasSize {
		return src, false, nil
	}

	directive := formatSceneDirective(sizeVal, gridVal, hasGrid)
	firstCodeIdx := findFirstCodeLineIndex(lines)
	firstSceneIdx := findFirstSceneCommentIndex(lines)

	var resultLines []string
	if firstCodeIdx == -1 || firstSceneIdx < firstCodeIdx {
		replaced := false
		for _, line := range lines {
			if isSceneCommentLine(line) {
				if !replaced {
					resultLines = append(resultLines, directive)
					replaced = true
				}
				continue
			}
			resultLines = append(resultLines, line)
		}
	} else {
		// Scene comment appeared after the first code token; insert @scene before firstCodeIdx.
		for i, line := range lines {
			if i == firstCodeIdx {
				resultLines = append(resultLines, directive, "")
			}
			if isSceneCommentLine(line) {
				continue
			}
			resultLines = append(resultLines, line)
		}
	}

	resultLines = collapseRedundantBlankLines(resultLines)
	out := joinAndRestoreEndings(resultLines, useCRLF)
	return out, out != src, nil
}

func hasTopLevelSceneDirective(lines []string) bool {
	inBlockComment := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if inBlockComment {
			if idx := strings.Index(trimmed, "*/"); idx != -1 {
				inBlockComment = false
				trimmed = strings.TrimSpace(trimmed[idx+2:])
			} else {
				continue
			}
		}
		if strings.HasPrefix(trimmed, "/*") {
			if !strings.Contains(trimmed, "*/") {
				inBlockComment = true
			}
			continue
		}
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}
		return sceneDirectiveRE.MatchString(trimmed)
	}
	return false
}

func findFirstCodeLineIndex(lines []string) int {
	inBlockComment := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if inBlockComment {
			if idx := strings.Index(trimmed, "*/"); idx != -1 {
				inBlockComment = false
				trimmed = strings.TrimSpace(trimmed[idx+2:])
			} else {
				continue
			}
		}
		if strings.HasPrefix(trimmed, "/*") {
			if !strings.Contains(trimmed, "*/") {
				inBlockComment = true
			}
			continue
		}
		if trimmed == "" || strings.HasPrefix(trimmed, "//") {
			continue
		}
		return i
	}
	return -1
}

func findFirstSceneCommentIndex(lines []string) int {
	for i, line := range lines {
		if isSceneCommentLine(line) {
			return i
		}
	}
	return -1
}

func hasLegacySceneCommentLine(lines []string) bool {
	return findFirstSceneCommentIndex(lines) != -1
}

func isSceneCommentLine(line string) bool {
	return sceneSizeLineRE.MatchString(line) || sceneGridLineRE.MatchString(line)
}

func parseLegacySceneComments(lines []string) (size float64, hasSize bool, grid float64, hasGrid bool) {
	var rawGrid string
	var foundGridLine bool

	for _, line := range lines {
		if !hasSize {
			if m := sceneSizeLineRE.FindStringSubmatch(line); len(m) >= 2 {
				sizeStr := strings.TrimSpace(m[1])
				if strings.HasSuffix(strings.ToLower(sizeStr), "mm") {
					sizeStr = strings.TrimSpace(sizeStr[:len(sizeStr)-2])
				}
				if v, err := strconv.ParseFloat(sizeStr, 64); err == nil && v > 0 {
					size = v
					hasSize = true
				}
			}
		}
		if !foundGridLine {
			if m := sceneGridLineRE.FindStringSubmatch(line); len(m) >= 2 {
				rawGrid = strings.TrimSpace(m[1])
				foundGridLine = true
			}
		}
	}

	if !hasSize {
		return 0, false, 0, false
	}

	if foundGridLine {
		if mmGridRE.MatchString(rawGrid) {
			valStr := strings.TrimSpace(rawGrid)
			if strings.HasSuffix(strings.ToLower(valStr), "mm") {
				valStr = strings.TrimSpace(valStr[:len(valStr)-2])
			}
			if val, err := strconv.ParseFloat(valStr, 64); err == nil && val > 0 {
				grid = val
				hasGrid = true
			}
		} else {
			if count, err := strconv.ParseFloat(rawGrid, 64); err == nil && count > 0 {
				grid = size / count
				hasGrid = true
			}
		}
	}

	return size, true, grid, hasGrid
}

func formatSceneDirective(size, grid float64, hasGrid bool) string {
	sizeStr := strconv.FormatFloat(size, 'f', -1, 64)
	defaultGrid := size / 10.0
	if !hasGrid || math.Abs(grid-defaultGrid) < 1e-9 {
		return fmt.Sprintf("@scene(%s)", sizeStr)
	}
	gridStr := strconv.FormatFloat(grid, 'f', -1, 64)
	return fmt.Sprintf("@scene(%s, %s)", sizeStr, gridStr)
}

func removeSceneCommentLines(lines []string) []string {
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		if isSceneCommentLine(line) {
			continue
		}
		out = append(out, line)
	}
	return collapseRedundantBlankLines(out)
}

func collapseRedundantBlankLines(lines []string) []string {
	out := make([]string, 0, len(lines))
	prevBlank := false
	for i, line := range lines {
		isBlank := strings.TrimSpace(line) == ""
		// Preserve a single trailing empty element from strings.Split when file ends with '\n'.
		if i == len(lines)-1 && line == "" {
			out = append(out, line)
			continue
		}
		if isBlank && prevBlank {
			continue
		}
		out = append(out, line)
		prevBlank = isBlank
	}
	return out
}

func joinAndRestoreEndings(lines []string, useCRLF bool) string {
	joined := strings.Join(lines, "\n")
	if useCRLF {
		return strings.ReplaceAll(joined, "\n", "\r\n")
	}
	return joined
}
