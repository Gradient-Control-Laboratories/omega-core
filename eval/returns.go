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
		name, _ := leadingIdentifier(s.text)
		if name != "return" {
			continue
		}
		body := strings.TrimSpace(s.text)
		body = strings.TrimPrefix(body, "return")
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
		name, isImport := leadingIdentifier(text)
		stmts = append(stmts, gslStmt{text: text, name: name, isImport: isImport})
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

// leadingIdentifier returns the first identifier in a statement (skipping
// leading whitespace and comments) and whether it is the `import` keyword.
func leadingIdentifier(stmt string) (name string, isImport bool) {
	i, n := 0, len(stmt)
	for i < n {
		c := stmt[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i++
		case c == '/' && i+1 < n && stmt[i+1] == '/':
			e := strings.IndexByte(stmt[i:], '\n')
			if e < 0 {
				return "", false
			}
			i += e
		case c == '/' && i+1 < n && stmt[i+1] == '*':
			e := strings.Index(stmt[i+2:], "*/")
			if e < 0 {
				return "", false
			}
			i += 2 + e + 2
		default:
			goto ident
		}
	}
ident:
	start := i
	for i < n {
		c := stmt[i]
		if c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || (i > start && c >= '0' && c <= '9') {
			i++
		} else {
			break
		}
	}
	name = stmt[start:i]
	return name, name == "import"
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
