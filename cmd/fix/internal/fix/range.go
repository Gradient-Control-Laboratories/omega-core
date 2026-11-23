package fix

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

// maxRangeLine is the longest line the range fixer writes the annotation on the
// same line as its assignment; a longer one goes on the line above.
const maxRangeLine = 100

var rangeFixer = Fixer{
	Name: "range",
	Doc:  "replace .min/.max/.period slider properties with an @range annotation",
	Help: `The range fixer replaces a top-level variable's "name.min", "name.max" and
optional "name.period" property assignments with one "@range(min, max[, period])"
annotation on the variable, using the same expressions.

The bounds are copied exactly as written, so a bound that reads other variables
(Slider Mechanism's slidePos) keeps reading them, and the annotation goes in front
of the assignment even when a bound reads a variable declared below it. The
annotation shares the assignment's line when the result fits in 100 columns, and
takes the line above it otherwise.

A variable is left as it is when the rewrite could not keep everything it says:
it has an explicit @range already, it is one of several variables a single
assignment declares, or one of its property statements holds a comment inside the
expression. Other properties of the variable (.color, .box, ...) stay where they
are, and a comment after a removed property stays on its own line.

Before:
  slidePos = 20.2;
  slidePos.min = abs(slideRadius - baseRadius) - slideOffset;
  slidePos.max = slideRadius + baseRadius - slideOffset;
  slidePos.period = 10;

After:
  @range(abs(slideRadius - baseRadius) - slideOffset, slideRadius + baseRadius - slideOffset, 10)
  slidePos = 20.2;
`,
	Apply: applyRangeFix,
}

// rangeEdit replaces src[start:end] with text.
type rangeEdit struct {
	start, end int
	text       string
}

func applyRangeFix(src string) (string, bool, error) {
	if src == "" {
		return src, false, nil
	}

	useCRLF := strings.Contains(src, "\r\n")
	normalized := src
	if useCRLF {
		normalized = strings.ReplaceAll(src, "\r\n", "\n")
	}

	res, err := parser.ParsePackage([]string{normalized})
	if err != nil || res.Package == nil || len(res.Package.Files) == 0 || res.Package.Files[0] == nil {
		return src, false, nil
	}

	var edits []rangeEdit
	var converted []string
	for _, a := range res.Package.Files[0].Assignments {
		es, name, ok := rangeEditsFor(normalized, a)
		if !ok {
			continue
		}
		edits = append(edits, es...)
		converted = append(converted, name)
	}
	if len(edits) == 0 {
		return src, false, nil
	}

	sort.SliceStable(edits, func(i, j int) bool { return edits[i].start < edits[j].start })
	edits = trimBlankRuns(normalized, edits)

	// Apply from the end of the file so each edit's offsets still hold.
	out := normalized
	for i := len(edits) - 1; i >= 0; i-- {
		e := edits[i]
		out = out[:e.start] + e.text + out[e.end:]
	}

	// The rewrite must leave each variable with an explicit range and none of the
	// properties it replaced; anything else is a bug here, not a file to skip.
	if err := checkRangeFix(out, converted); err != nil {
		return src, false, err
	}

	if useCRLF {
		out = strings.ReplaceAll(out, "\n", "\r\n")
	}
	return out, out != src, nil
}

// rangeEditsFor is the edits that turn one assignment's .min/.max/.period
// statements into an @range annotation, and the variable they are for. ok is
// false for an assignment this fixer leaves alone.
func rangeEditsFor(src string, a *program.Assignment) (edits []rangeEdit, name string, ok bool) {
	if a == nil || a.Start == nil || len(a.Variables) != 1 || a.Variables[0] == nil {
		return nil, "", false
	}
	if a.Range != nil {
		return nil, "", false
	}
	v := a.Variables[0]

	props := map[string]*program.Property{}
	for _, p := range v.Properties {
		if p != nil && (p.Name == "min" || p.Name == "max" || p.Name == "period") {
			props[p.Name] = p
		}
	}
	if props["min"] == nil || props["max"] == nil {
		return nil, "", false
	}

	// The bound is the text after the property's "=", because an operand's own
	// span stops at its first operand ("a - b" reads as "a").
	var args []string
	var stmts []*program.Property
	for _, n := range []string{"min", "max", "period"} {
		p := props[n]
		if p == nil {
			continue
		}
		arg, ok := propertyArgument(src, v.Name, p)
		if !ok {
			return nil, "", false
		}
		args = append(args, arg)
		stmts = append(stmts, p)
	}

	annotation := "@range(" + strings.Join(args, ", ") + ")"
	assignStart := int(a.Start.Offset)
	if assignStart < 0 || assignStart > len(src) {
		return nil, "", false
	}
	lineStart := strings.LastIndexByte(src[:assignStart], '\n') + 1
	lineEnd := len(src)
	if i := strings.IndexByte(src[assignStart:], '\n'); i >= 0 {
		lineEnd = assignStart + i
	}
	indent := src[lineStart:assignStart]
	if strings.TrimSpace(indent) == "" && len(indent)+len(annotation)+1+len(src[assignStart:lineEnd]) > maxRangeLine && !strings.Contains(annotation, "\n") {
		edits = append(edits, rangeEdit{assignStart, assignStart, annotation + "\n" + indent})
	} else {
		edits = append(edits, rangeEdit{assignStart, assignStart, annotation + " "})
	}

	for _, p := range stmts {
		edits = append(edits, removeStatement(src, int(p.Start.Offset), int(p.End.Offset)))
	}
	return edits, v.Name, true
}

// propertyArgument is the expression of a `name.prop = expr;` statement, as
// written.
func propertyArgument(src, name string, p *program.Property) (string, bool) {
	if p.Start == nil || p.End == nil {
		return "", false
	}
	start, end := int(p.Start.Offset), int(p.End.Offset)
	if start < 0 || end > len(src) || start >= end {
		return "", false
	}
	re := regexp.MustCompile(`(?s)^` + regexp.QuoteMeta(name) + `\s*\.\s*` + regexp.QuoteMeta(p.Name) + `\s*=\s*(.*?)\s*;$`)
	m := re.FindStringSubmatch(src[start:end])
	if m == nil || m[1] == "" || strings.Contains(m[1], "//") || strings.Contains(m[1], "/*") {
		return "", false
	}
	return m[1], true
}

// removeStatement is the edit that removes the statement at src[start:end]. A
// statement alone on its lines takes them with it, one that shares a line takes
// only itself and the space after it, and a comment after it stays.
func removeStatement(src string, start, end int) rangeEdit {
	lineStart := strings.LastIndexByte(src[:start], '\n') + 1
	lineEnd := len(src)
	if i := strings.IndexByte(src[end:], '\n'); i >= 0 {
		lineEnd = end + i
	}
	before := src[lineStart:start]
	after := src[end:lineEnd]

	switch {
	case strings.TrimSpace(before) == "" && strings.TrimSpace(after) == "":
		stop := lineEnd
		if stop < len(src) {
			stop++ // the newline
		}
		return rangeEdit{lineStart, stop, ""}
	case strings.TrimSpace(before) == "" && strings.HasPrefix(strings.TrimSpace(after), "//"):
		return rangeEdit{start, end + (len(after) - len(strings.TrimLeft(after, " \t"))), ""}
	default:
		stop := end + (len(after) - len(strings.TrimLeft(after, " \t")))
		return rangeEdit{start, stop, ""}
	}
}

// trimBlankRuns merges the removals that touch, so a block of removed lines is
// judged as one, and extends a removed block of whole lines over the blank line
// after it when a blank line comes before it too: the block sat between two
// blanks, and they should not become two. edits must be sorted by start.
func trimBlankRuns(src string, edits []rangeEdit) []rangeEdit {
	var merged []rangeEdit
	for _, e := range edits {
		if n := len(merged); n > 0 && e.text == "" && merged[n-1].text == "" && merged[n-1].end == e.start {
			merged[n-1].end = e.end
			continue
		}
		merged = append(merged, e)
	}
	for i, e := range merged {
		// Statements that shared a line with nothing but each other leave an empty
		// line behind; take it with them.
		if e.text == "" && e.end > 0 && src[e.end-1] != '\n' {
			lineStart := strings.LastIndexByte(src[:e.start], '\n') + 1
			lineEnd := len(src)
			if j := strings.IndexByte(src[e.end:], '\n'); j >= 0 {
				lineEnd = e.end + j
			}
			if strings.TrimSpace(src[lineStart:e.start]) == "" && strings.TrimSpace(src[e.end:lineEnd]) == "" {
				merged[i].start = lineStart
				merged[i].end = min(lineEnd+1, len(src))
				e = merged[i]
			}
		}
		wholeLines := e.text == "" && (e.start == 0 || src[e.start-1] == '\n') && (e.end == len(src) || src[e.end-1] == '\n')
		if !wholeLines || !blankLineBefore(src, e.start) || !blankLineAt(src, e.end) {
			continue
		}
		if next := strings.IndexByte(src[e.end:], '\n'); next >= 0 {
			merged[i].end = e.end + next + 1
		} else {
			merged[i].end = len(src)
		}
	}
	return merged
}

// blankLineBefore reports whether the line ending just before offset (the start
// of a line) is blank.
func blankLineBefore(src string, lineStart int) bool {
	if lineStart == 0 {
		return false
	}
	prevStart := strings.LastIndexByte(src[:lineStart-1], '\n') + 1
	return strings.TrimSpace(src[prevStart:lineStart-1]) == ""
}

// blankLineAt reports whether the line starting at offset is blank.
func blankLineAt(src string, lineStart int) bool {
	end := len(src)
	if i := strings.IndexByte(src[lineStart:], '\n'); i >= 0 {
		end = lineStart + i
	}
	return lineStart < len(src) && strings.TrimSpace(src[lineStart:end]) == ""
}

// checkRangeFix verifies out: every variable in converted has an explicit
// @range and none of the .min/.max/.period properties it replaced.
func checkRangeFix(out string, converted []string) error {
	res, err := parser.ParsePackage([]string{out})
	if err != nil || res.Package == nil || len(res.Package.Files) == 0 {
		return fmt.Errorf("range: the rewritten source no longer parses: %v", err)
	}
	want := map[string]bool{}
	for _, n := range converted {
		want[n] = true
	}
	for _, a := range res.Package.Files[0].Assignments {
		if a == nil || len(a.Variables) != 1 || !want[a.Variables[0].Name] {
			continue
		}
		name := a.Variables[0].Name
		if a.Range == nil {
			return fmt.Errorf("range: %s has no @range after the rewrite", name)
		}
		for _, p := range a.Variables[0].Properties {
			if p != nil && (p.Name == "min" || p.Name == "max" || p.Name == "period") {
				return fmt.Errorf("range: %s.%s is still assigned after the rewrite", name, p.Name)
			}
		}
		delete(want, name)
	}
	for n := range want {
		return fmt.Errorf("range: %s is missing after the rewrite", n)
	}
	return nil
}
