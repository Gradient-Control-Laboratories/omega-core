package parser2

import (
	"regexp"
	"strconv"
	"strings"

	"gradientcontrol.com/omega/core/program"
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
	sizeMatch := sceneSizeRE.FindStringSubmatch(source)
	if len(sizeMatch) < 2 {
		return nil
	}
	sizeStr := strings.TrimSpace(sizeMatch[1])
	// Tree documents use the same explicit millimeter unit as the grid header.
	if strings.HasSuffix(strings.ToLower(sizeStr), "mm") {
		sizeStr = strings.TrimSpace(sizeStr[:len(sizeStr)-2])
	}
	size, err := strconv.ParseFloat(sizeStr, 64)
	if err != nil || size <= 0 {
		return nil
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

	s, err := NewScene(size, grid, nil, nil)
	if err != nil {
		return nil
	}
	return s
}
