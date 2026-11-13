package eval

import (
	"fmt"
	"strings"

	"gradientcontrol.com/omega/core/assembler/slider"
	"gradientcontrol.com/omega/core/parser"
)

// SliderRanges parses the document and returns every declared slider's [min, max] range.
func SliderRanges(code string) (map[string][2]float64, error) {
	res, err := parser.ParsePackage([]string{code})
	if err != nil {
		return nil, fmt.Errorf("parse document for sliders: %w", err)
	}
	sliders := slider.Extract(code, res.Package)
	if len(sliders) == 0 {
		return nil, fmt.Errorf("document declares no sliders (variables with .min and .max) to certify over")
	}
	params := make(map[string][2]float64, len(sliders))
	for _, s := range sliders {
		params[s.VariableName] = [2]float64{s.Min, s.Max}
	}
	return params, nil
}

// TopLevelNames returns the set of names declared at the top level of source.
// Exported for graders that need to know whether a document still declares a
// name its checks bind to.
func TopLevelNames(source string) map[string]bool {
	return topLevelNames(source)
}

// TopLevelReturnExprs returns the source text of each expression in the
// document's top-level return statement, in order — nil when the source has
// no top-level return. Graders use it to align a candidate document's returns
// with the names a case's checks were written against.
func TopLevelReturnExprs(source string) []string {
	for _, s := range topLevelStatements(source) {
		if s.name != "return" {
			continue
		}
		// The statement's text starts with what leads it -- comments, the @scene directive -- and its return
		// keyword is where leadingIdentifier found it.
		body := strings.TrimPrefix(s.text[s.keep:], "return")
		body = strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(body), ";"))
		if body == "" {
			return nil
		}
		exprs := splitTopLevelCommas(body)
		for i := range exprs {
			exprs[i] = strings.TrimSpace(exprs[i])
		}
		return exprs
	}
	return nil
}

// splitTopLevelCommas splits src on commas at bracket depth zero, treating
// strings and comments as opaque — the same conventions as
// topLevelStatements.
func splitTopLevelCommas(src string) []string {
	var parts []string
	depth := 0
	start := 0
	i, n := 0, len(src)
	for i < n {
		c := src[i]
		switch {
		case c == '/' && i+1 < n && src[i+1] == '/':
			e := strings.IndexByte(src[i:], '\n')
			if e < 0 {
				i = n
			} else {
				i += e
			}
		case c == '/' && i+1 < n && src[i+1] == '*':
			e := strings.Index(src[i+2:], "*/")
			if e < 0 {
				i = n
			} else {
				i += 2 + e + 2
			}
		case c == '"':
			j := i + 1
			for j < n && src[j] != '"' {
				if src[j] == '\\' {
					j++
				}
				j++
			}
			if j < n {
				j++
			}
			i = j
		case c == '(' || c == '[' || c == '{':
			depth++
			i++
		case c == ')' || c == ']' || c == '}':
			if depth > 0 {
				depth--
			}
			i++
		case c == ',' && depth == 0:
			parts = append(parts, src[start:i])
			i++
			start = i
		default:
			i++
		}
	}
	parts = append(parts, src[start:])
	return parts
}

// gslStmt is one top-level GSL statement plus its leading identifier.
type gslStmt struct {
	text     string // raw text, including leading whitespace and terminator
	name     string // leading identifier ("" if none)
	keep     int    // where the declaration begins in text (leadingIdentifier)
	isImport bool   // leading keyword is "import"
}

// topLevelStatements splits GSL source into top-level statements. A
// statement ends at a top-level ';' (assignments, returns) or at the '}'
// that closes a top-level '{...}' block (function declarations). Comments
// and strings are opaque. Concatenating the statement texts reproduces the
// source exactly, so callers can drop statements and re-emit the rest
// verbatim.
func topLevelStatements(source string) []gslStmt {
	var stmts []gslStmt
	var buf strings.Builder
	flush := func() {
		text := buf.String()
		buf.Reset()
		if strings.TrimSpace(text) == "" {
			if text != "" && len(stmts) > 0 {
				stmts[len(stmts)-1].text += text
			} else if text != "" {
				stmts = append(stmts, gslStmt{text: text})
			}
			return
		}
		name, keep, isImport := leadingIdentifier(text)
		stmts = append(stmts, gslStmt{text: text, name: name, keep: keep, isImport: isImport})
	}

	depth := 0
	i, n := 0, len(source)
	for i < n {
		c := source[i]
		switch {
		case c == '/' && i+1 < n && source[i+1] == '/':
			end := strings.IndexByte(source[i:], '\n')
			if end < 0 {
				end = n - i
			}
			buf.WriteString(source[i : i+end])
			i += end
		case c == '/' && i+1 < n && source[i+1] == '*':
			end := strings.Index(source[i+2:], "*/")
			stop := n
			if end >= 0 {
				stop = i + 2 + end + 2
			}
			buf.WriteString(source[i:stop])
			i = stop
		case c == '"':
			j := i + 1
			for j < n && source[j] != '"' {
				if source[j] == '\\' {
					j++
				}
				j++
			}
			if j < n {
				j++
			}
			buf.WriteString(source[i:j])
			i = j
		case c == '{' || c == '[' || c == '(':
			depth++
			buf.WriteByte(c)
			i++
		case c == '}' || c == ']' || c == ')':
			if depth > 0 {
				depth--
			}
			buf.WriteByte(c)
			i++
			if c == '}' && depth == 0 {
				flush()
			}
		case c == ';' && depth == 0:
			buf.WriteByte(c)
			i++
			flush()
		default:
			buf.WriteByte(c)
			i++
		}
	}
	flush()
	return stmts
}

// leadingIdentifier returns the first identifier in a statement and whether
// it is the `import` keyword, skipping leading whitespace, comments, the
// `@scene(...)` directive (which ends without a `;`, so it leads the
// statement after it) and the declaration's own annotations (`@range(5, 15)
// r = 10.0;`). A statement that sets a property (`r.min = 5.0;`) declares
// nothing: its name is "". keep is where the declaration begins, its
// annotations included: what precedes it (whitespace, comments, `@scene`)
// stays when the declaration is replaced.
func leadingIdentifier(stmt string) (name string, keep int, isImport bool) {
	i, n := 0, len(stmt)
	annotated := -1 // where the declaration's first annotation starts
	for i < n {
		c := stmt[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i++
		case c == '/' && i+1 < n && stmt[i+1] == '/':
			e := strings.IndexByte(stmt[i:], '\n')
			if e < 0 {
				return "", n, false
			}
			i += e
		case c == '/' && i+1 < n && stmt[i+1] == '*':
			e := strings.Index(stmt[i+2:], "*/")
			if e < 0 {
				return "", n, false
			}
			i += 2 + e + 2
		case c == '@':
			at := i
			var directive string
			directive, i = skipAnnotation(stmt, i)
			if directive != "scene" && annotated < 0 {
				annotated = at
			}
		default:
			goto ident
		}
	}
ident:
	start := i
	i = identifierEnd(stmt, i)
	name = stmt[start:i]
	keep = start
	if annotated >= 0 {
		keep = annotated
	}
	if name == "import" {
		return name, keep, true // `import . "omega";` is a dot import, not a property
	}
	if name != "" && i < n && stmt[i] == '.' {
		return "", keep, false
	}
	return name, keep, false
}

// identifierEnd is the end of the identifier starting at i (i when none does).
func identifierEnd(s string, i int) int {
	start := i
	for i < len(s) {
		c := s[i]
		if c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || (i > start && c >= '0' && c <= '9') {
			i++
		} else {
			break
		}
	}
	return i
}

// skipAnnotation reads the `@name` or `@name(args)` at s[i]: its name and the
// index after it. Strings in the arguments are opaque.
func skipAnnotation(s string, i int) (string, int) {
	end := identifierEnd(s, i+1)
	name := s[i+1 : end]
	j := end
	for j < len(s) && (s[j] == ' ' || s[j] == '\t') {
		j++
	}
	if j >= len(s) || s[j] != '(' {
		return name, end
	}
	depth := 0
	for j < len(s) {
		switch s[j] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return name, j + 1
			}
		case '"':
			for j++; j < len(s) && s[j] != '"'; j++ {
				if s[j] == '\\' {
					j++
				}
			}
		}
		j++
	}
	return name, len(s)
}

// topLevelNames is the set of names declared at the top level of source.
func topLevelNames(source string) map[string]bool {
	names := make(map[string]bool)
	for _, s := range topLevelStatements(source) {
		if s.name != "" && !s.isImport {
			names[s.name] = true
		}
	}
	return names
}
