package fix

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

var (
	sceneSizeLineRE = regexp.MustCompile(`(?i)^\s*//\s*Scene\s+Size:[^\S\n]*(.*)$`)
	sceneGridLineRE = regexp.MustCompile(`(?i)^\s*//\s*Scene\s+Grid:[^\S\n]*(.*)$`)
)

var sceneFixer = Fixer{
	Name: "scene",
	Doc:  "replace legacy // Scene Size and // Scene Grid comments with @scene directive",
	Help: `The scene fixer replaces legacy "// Scene Size:" and "// Scene Grid:"
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

	pkg, _ := parser.ParsePackage([]string{normalized})
	if pkg.Scene == nil {
		return src, false, nil
	}

	lines := strings.Split(normalized, "\n")
	firstSceneIdx := findFirstSceneCommentIndex(lines)

	// pkg.Scene.Start != nil indicates an explicit @scene(...) directive was parsed.
	if pkg.Scene.Start != nil {
		if firstSceneIdx == -1 {
			return src, false, nil
		}
		cleaned := removeSceneCommentLines(lines)
		out := joinAndRestoreEndings(cleaned, useCRLF)
		return out, out != src, nil
	}

	directive := formatSceneDirective(pkg.Scene.Size, pkg.Scene.Grid)
	firstCodeIdx := firstStatementLineIndex(pkg.Files[0], lines)

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

func firstStatementLineIndex(file *program.PackageFile, lines []string) int {
	minLine := int32(0)
	consider := func(loc *program.Location) {
		if minLine == 0 || loc.Line < minLine {
			minLine = loc.Line
		}
	}

	for _, imp := range file.Imports {
		consider(imp.Start)
	}
	for _, imp := range file.MeshImports {
		consider(imp.Start)
	}
	for _, imp := range file.StepImports {
		consider(imp.Start)
	}
	for _, imp := range file.WarpImports {
		consider(imp.Start)
	}
	for _, ann := range file.Annotations {
		consider(ann.Start)
	}
	for _, assign := range file.Assignments {
		consider(assign.Start)
	}
	for _, ret := range file.Returns {
		consider(ret.Start)
	}

	if minLine > 0 {
		return int(minLine) - 1
	}
	return findFirstCodeLineIndex(lines)
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

func isSceneCommentLine(line string) bool {
	return sceneSizeLineRE.MatchString(line) || sceneGridLineRE.MatchString(line)
}

func formatSceneDirective(size, grid float64) string {
	sizeStr := strconv.FormatFloat(size, 'f', -1, 64)
	defaultGrid := size / 10.0
	if math.Abs(grid-defaultGrid) < 1e-9 {
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
