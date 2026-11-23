package parser2

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/Gradient-Control-Laboratories/omega-core/program"
)

var (
	sceneSizeRE = regexp.MustCompile(`(?mi)^\s*//\s*Scene Size:[^\S\n]*(.*)$`)
	sceneGridRE = regexp.MustCompile(`(?mi)^\s*//\s*Scene Grid:[^\S\n]*(.*)$`)
	mmGridRE    = regexp.MustCompile(`(?i)^\s*[-+0-9.eE]+\s*mm\b`)
)

// fromLegacyComments inspects the source text for legacy `// Scene Size:` and
// `// Scene Grid:` comments. Returns nil if no `// Scene Size:` comment is found
// or if parsing fails.
func fromLegacyComments(source string) *program.Scene {
	sc, _ := fromLegacyCommentsWithDiag(source)
	return sc
}

// fromLegacyCommentsWithDiag parses legacy `// Scene Size:` / `// Scene Grid:`
// comments and, when successful, returns a warning diagnostic pointing at the
// legacy `// Scene Size:` comment.
func fromLegacyCommentsWithDiag(source string) (*program.Scene, *program.Diagnostic) {
	loc := sceneSizeRE.FindStringSubmatchIndex(source)
	if len(loc) < 4 {
		return nil, nil
	}
	sizeStr := strings.TrimSpace(source[loc[2]:loc[3]])
	// Tree documents use the same explicit millimeter unit as the grid header.
	if strings.HasSuffix(strings.ToLower(sizeStr), "mm") {
		sizeStr = strings.TrimSpace(sizeStr[:len(sizeStr)-2])
	}
	size, err := strconv.ParseFloat(sizeStr, 64)
	if err != nil || size <= 0 {
		return nil, nil
	}

	var grid *float64
	gridMatch := sceneGridRE.FindStringSubmatch(source)
	if len(gridMatch) >= 2 {
		gridText := strings.TrimSpace(gridMatch[1])
		if mmGridRE.MatchString(gridText) {
			// Explicit millimeter spacing e.g. "5 mm"
			valStr := strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(gridText), "mm"))
			if val, err := strconv.ParseFloat(valStr, 64); err == nil && val > 0 {
				grid = &val
			}
		} else {
			// Legacy division count e.g. "10" -> spacing = size / count
			if count, err := strconv.ParseFloat(gridText, 64); err == nil && count > 0 {
				spacing := size / count
				grid = &spacing
			}
		}
	}

	s, err := newScene(size, grid, nil, nil)
	if err != nil {
		return nil, nil
	}

	diag := legacySceneDiagnostic(source, loc[0])
	return s, &diag
}

func legacySceneDiagnostic(source string, matchStart int) program.Diagnostic {
	commentStart := matchStart
	for commentStart < len(source) && (source[commentStart] == ' ' || source[commentStart] == '\t') {
		commentStart++
	}
	lineStart := strings.LastIndexByte(source[:commentStart], '\n') + 1
	lineEnd := len(source)
	if rel := strings.IndexByte(source[commentStart:], '\n'); rel >= 0 {
		lineEnd = commentStart + rel
	}
	line := strings.TrimRight(source[lineStart:lineEnd], "\r")
	lineNum := 1 + strings.Count(source[:commentStart], "\n")
	col := commentStart - lineStart + 1
	length := max(len(line)-(commentStart-lineStart), 1)
	return program.Diagnostic{
		Severity: program.SeverityWarning,
		Line:     line,
		LineNum:  lineNum,
		Column:   col,
		Length:   length,
		Message:  "legacy `// Scene Size:` comment is deprecated; use `@scene(...)` instead and run `gsl fix` to update",
	}
}
