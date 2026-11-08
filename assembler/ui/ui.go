package ui

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"gradientcontrol.com/omega/core/program/edit"
)

const (
	KindPoint     = "point"
	KindTransform = "transform"
	KindSpline    = "spline"
	KindColor     = "color"
	KindDirection = "direction"
)

type Control struct {
	Kind     string    `json:"kind"`
	Span     edit.Span `json:"span"`
	Value    []float64 `json:"value"`
	Params   []Param   `json:"params,omitempty"`
	Controls []Control `json:"controls,omitempty"`
	// Anchor is where a direction's arrow is drawn from, uiDirection's
	// second argument evaluated as a constant; nil draws it from the origin.
	Anchor []float64 `json:"anchor,omitempty"`
}

type Param struct {
	ParamName string    `json:"paramName"`
	Span      edit.Span `json:"span"`
	Value     float64   `json:"value"`
}

// RewriteSplineControls replaces non-literal uiSpline control expressions in source
// with JSON-marshaled 4x4 matrix literals, returning the rewritten source along
// with the number of spline controls and unique control matrices visited.
func RewriteSplineControls(source string, controls []Control) (string, int, int, error) {
	type replacement struct {
		start, end int
		text       string
	}
	var replacements []replacement
	lines := strings.SplitAfter(source, "\n")
	offset := func(line, column int) (int, error) {
		if line < 1 || line > len(lines) || column < 1 {
			return 0, fmt.Errorf("invalid gate control location")
		}
		runes := []rune(lines[line-1])
		if column > len(runes)+1 {
			return 0, fmt.Errorf("invalid gate control column")
		}
		n := len(string(runes[:column-1]))
		for _, prefix := range lines[:line-1] {
			n += len(prefix)
		}
		return n, nil
	}
	seen := map[int]bool{}
	splines := 0
	for _, spline := range controls {
		if spline.Kind != KindSpline {
			continue
		}
		splines++
		for _, control := range spline.Controls {
			start, err := offset(control.Span.StartLine, control.Span.StartColumn)
			if err != nil {
				return "", 0, 0, err
			}
			end, err := offset(control.Span.EndLine, control.Span.EndColumn)
			if err != nil {
				return "", 0, 0, err
			}
			if end < start || len(control.Value) != 16 {
				return "", 0, 0, fmt.Errorf("invalid gate control matrix")
			}
			if seen[start] {
				continue
			}
			seen[start] = true
			if len(control.Params) == 16 {
				continue
			}
			var matrix [4][4]float64
			for i := range matrix {
				copy(matrix[i][:], control.Value[4*i:4*i+4])
			}
			literal, err := json.Marshal(matrix)
			if err != nil {
				return "", 0, 0, err
			}
			replacements = append(replacements, replacement{start, end, string(literal)})
		}
	}
	sort.Slice(replacements, func(i, j int) bool { return replacements[i].start > replacements[j].start })
	for _, replacement := range replacements {
		source = source[:replacement.start] + replacement.text + source[replacement.end:]
	}
	return source, splines, len(seen), nil
}

