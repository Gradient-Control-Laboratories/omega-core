package fix

import (
	"strconv"
	"strings"
)

var omegaImportFixer = Fixer{
	Name: "omegaimport",
	Doc:  `remove unnecessary import . "omega" statements`,
	Help: `The omegaimport fixer removes unnecessary 'import . "omega";' statements
from GSL source files because the builtin package is implicitly imported.

Before:
  import . "omega";

  return sphere(1);

After:
  return sphere(1);
`,
	Apply: applyOmegaImportFix,
}

func applyOmegaImportFix(src string) (string, bool, error) {
	if src == "" || !strings.Contains(src, "import") {
		return src, false, nil
	}

	useCRLF := strings.Contains(src, "\r\n")
	normalized := src
	if useCRLF {
		normalized = strings.ReplaceAll(src, "\r\n", "\n")
	}

	firstStart, _, ok := findFirstOmegaDotImport(normalized)
	if !ok {
		return src, false, nil
	}

	firstLineStart := strings.LastIndexByte(normalized[:firstStart], '\n') + 1
	hadLeadingContent := strings.TrimSpace(normalized[:firstLineStart]) != ""
	hadTrailingContent := true

	curr := normalized
	for {
		stmtStart, stmtEnd, found := findFirstOmegaDotImport(curr)
		if !found {
			break
		}
		hadTrailingContent = strings.TrimSpace(curr[stmtEnd:]) != ""
		curr = removeImportSpan(curr, stmtStart, stmtEnd)
	}

	if curr != "" {
		lines := strings.Split(curr, "\n")
		if !hadLeadingContent {
			for len(lines) > 1 && strings.TrimSpace(lines[0]) == "" {
				lines = lines[1:]
			}
		}
		lines = collapseRedundantBlankLines(lines)
		if !hadTrailingContent {
			if len(lines) > 0 && lines[len(lines)-1] == "" {
				for len(lines) >= 2 && strings.TrimSpace(lines[len(lines)-2]) == "" {
					lines = append(lines[:len(lines)-2], "")
				}
				if len(lines) == 1 && lines[0] == "" {
					lines = nil
				}
			} else {
				for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
					lines = lines[:len(lines)-1]
				}
			}
		}
		curr = strings.Join(lines, "\n")
	}

	if useCRLF {
		curr = strings.ReplaceAll(curr, "\n", "\r\n")
	}
	return curr, curr != src, nil
}

func removeImportSpan(s string, stmtStart, stmtEnd int) string {
	lineStart := strings.LastIndexByte(s[:stmtStart], '\n') + 1
	lineEnd := len(s)
	if rel := strings.IndexByte(s[stmtEnd:], '\n'); rel >= 0 {
		lineEnd = stmtEnd + rel
	}

	beforeOnLine := s[lineStart:stmtStart]
	afterOnLine := s[stmtEnd:lineEnd]
	hasBefore := strings.TrimSpace(beforeOnLine) != ""
	hasAfter := strings.TrimSpace(afterOnLine) != ""

	switch {
	case !hasBefore && !hasAfter:
		if lineEnd < len(s) {
			return s[:lineStart] + s[lineEnd+1:]
		}
		if lineStart > 0 {
			return s[:lineStart-1]
		}
		return ""

	case !hasBefore && hasAfter:
		cutEnd := stmtEnd
		for cutEnd < lineEnd && (s[cutEnd] == ' ' || s[cutEnd] == '\t') {
			cutEnd++
		}
		return s[:stmtStart] + s[cutEnd:]

	case hasBefore && !hasAfter:
		cutStart := stmtStart
		for cutStart > lineStart && (s[cutStart-1] == ' ' || s[cutStart-1] == '\t') {
			cutStart--
		}
		return s[:cutStart] + s[lineEnd:]

	default:
		cutEnd := stmtEnd
		for cutEnd < lineEnd && (s[cutEnd] == ' ' || s[cutEnd] == '\t') {
			cutEnd++
		}
		return s[:stmtStart] + s[cutEnd:]
	}
}

type gslTokenKind int

const (
	tokOther gslTokenKind = iota
	tokIdent
	tokDot
	tokString
	tokSemicolon
)

type gslToken struct {
	kind  gslTokenKind
	start int
	end   int
	text  string
}

func findFirstOmegaDotImport(src string) (int, int, bool) {
	pos := 0
	prevKind := tokOther
	for {
		t0, ok := nextGSLToken(src, pos)
		if !ok {
			return 0, 0, false
		}
		if prevKind != tokDot && t0.kind == tokIdent && t0.text == "import" {
			if t1, ok := nextGSLToken(src, t0.end); ok && t1.kind == tokDot {
				if t2, ok := nextGSLToken(src, t1.end); ok && t2.kind == tokString {
					if unquoted, err := strconv.Unquote(t2.text); err == nil && unquoted == "omega" {
						if t3, ok := nextGSLToken(src, t2.end); ok && t3.kind == tokSemicolon {
							return t0.start, t3.end, true
						}
					}
				}
			}
		}
		prevKind = t0.kind
		pos = t0.end
	}
}

func nextGSLToken(src string, pos int) (gslToken, bool) {
	for pos < len(src) {
		c := src[pos]
		if c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			pos++
			continue
		}
		if c == '/' && pos+1 < len(src) {
			if src[pos+1] == '/' {
				pos += 2
				for pos < len(src) && src[pos] != '\n' {
					pos++
				}
				continue
			}
			if src[pos+1] == '*' {
				pos += 2
				endIdx := strings.Index(src[pos:], "*/")
				if endIdx < 0 {
					return gslToken{}, false
				}
				pos += endIdx + 2
				continue
			}
		}
		break
	}
	if pos >= len(src) {
		return gslToken{}, false
	}

	start := pos
	c := src[pos]
	switch {
	case c == '"':
		pos++
		for pos < len(src) {
			if src[pos] == '\\' {
				pos += 2
				continue
			}
			if src[pos] == '"' {
				pos++
				return gslToken{kind: tokString, start: start, end: pos, text: src[start:pos]}, true
			}
			pos++
		}
		return gslToken{}, false

	case isIdentStart(c):
		pos++
		for pos < len(src) && isIdentPart(src[pos]) {
			pos++
		}
		return gslToken{kind: tokIdent, start: start, end: pos, text: src[start:pos]}, true

	case c == '.':
		if pos+1 < len(src) && isDigit(src[pos+1]) {
			pos = consumeNumber(src, pos)
			return gslToken{kind: tokOther, start: start, end: pos, text: src[start:pos]}, true
		}
		return gslToken{kind: tokDot, start: start, end: start + 1, text: "."}, true

	case isDigit(c):
		pos = consumeNumber(src, pos)
		return gslToken{kind: tokOther, start: start, end: pos, text: src[start:pos]}, true

	case c == ';':
		return gslToken{kind: tokSemicolon, start: start, end: start + 1, text: ";"}, true

	default:
		return gslToken{kind: tokOther, start: start, end: start + 1, text: src[start : start+1]}, true
	}
}

func isIdentStart(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_'
}

func isIdentPart(c byte) bool {
	return isIdentStart(c) || isDigit(c)
}

func isDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func consumeNumber(src string, pos int) int {
	for pos < len(src) && isDigit(src[pos]) {
		pos++
	}
	if pos < len(src) && src[pos] == '.' {
		pos++
		for pos < len(src) && isDigit(src[pos]) {
			pos++
		}
	}
	if pos < len(src) && (src[pos] == 'e' || src[pos] == 'E') {
		exp := pos + 1
		if exp < len(src) && (src[exp] == '+' || src[exp] == '-') {
			exp++
		}
		if exp < len(src) && isDigit(src[exp]) {
			pos = exp + 1
			for pos < len(src) && isDigit(src[pos]) {
				pos++
			}
		}
	}
	return pos
}
