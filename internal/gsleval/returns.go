package gsleval

import "strings"

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
