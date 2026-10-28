package scene

import (
	"fmt"
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

// New constructs and validates a new Scene protobuf message.
// If grid is nil, grid spacing defaults to size / 10.0.
func New(size float64, grid *float64, start, end *program.Location) (*program.Scene, error) {
	if size <= 0 {
		return nil, fmt.Errorf("scene size must be greater than zero, got %v", size)
	}

	var g float64
	if grid != nil {
		if *grid <= 0 {
			return nil, fmt.Errorf("scene grid spacing must be greater than zero, got %v", *grid)
		}
		g = *grid
	} else {
		g = size / 10.0
	}

	return &program.Scene{
		Start: start,
		End:   end,
		Size:  size,
		Grid:  g,
	}, nil
}

// FromLegacyComments inspects the source text for legacy `// Scene Size:` and
// `// Scene Grid:` comments. Returns nil if no `// Scene Size:` comment is found
// or if parsing fails.
func FromLegacyComments(source string) *program.Scene {
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

	s, err := New(size, grid, nil, nil)
	if err != nil {
		return nil
	}
	return s
}

// ResolveWorldExtent returns the worldExtent (half-size of the scene bounding box)
// based on the scene definition, or falls back to the requested fallback or default 10.0.
func ResolveWorldExtent(s *program.Scene, fallback float64) float64 {
	if s != nil && s.Size > 0 {
		return s.Size / 2.0
	}
	if fallback > 0 {
		return fallback
	}
	return 10.0
}
