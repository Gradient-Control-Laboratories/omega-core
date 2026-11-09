package wgsltest

import (
	"errors"
	"os"
	"strings"
	"testing"
	"unicode"

	"gradientcontrol.com/omega/core/internal/test/assert"
)

// AssertWGSLEqual verifies that actualWGSL matches the golden file at goldenPath.
// If the golden file is missing or UPDATE_GOLDEN is set, it creates or overwrites the golden.
// If exact string comparison fails, it compares top-level declarations (functions, structs, vars).
// If all declarations match and only declaration order differs, it auto-updates the golden file
// when autoUpdate is true (or passes if autoUpdate is false).
func AssertWGSLEqual(t *testing.T, goldenPath string, actualWGSL string, autoUpdate bool) {
	t.Helper()

	actualTrimmed := strings.TrimSpace(actualWGSL)
	actualBytes := []byte(actualTrimmed + "\n")
	goldenBytes, err := os.ReadFile(goldenPath)

	if errors.Is(err, os.ErrNotExist) || os.Getenv("UPDATE_GOLDEN") != "" {
		if err == nil && strings.TrimSpace(string(goldenBytes)) != actualTrimmed {
			logGoldenSizeChange(t, goldenPath, len(goldenBytes), len(actualBytes))
		}
		err = os.WriteFile(goldenPath, actualBytes, 0644)
		assert.NoError(t, err)
		if os.Getenv("UPDATE_GOLDEN") == "" {
			t.Skipf("Generated expected WGSL file: %s", goldenPath)
		}
		return
	}
	assert.NoError(t, err)

	expectedTrimmed := strings.TrimSpace(string(goldenBytes))
	if expectedTrimmed == actualTrimmed {
		return
	}

	logGoldenSizeChange(t, goldenPath, len(goldenBytes), len(actualBytes))

	// Declaration-aware fallback comparison
	expDecls, expOk := parseDeclarations(expectedTrimmed)
	actDecls, actOk := parseDeclarations(actualTrimmed)

	if expOk && actOk && declarationsEqual(expDecls, actDecls) {
		if autoUpdate {
			err = os.WriteFile(goldenPath, actualBytes, 0644)
			assert.NoError(t, err)
			t.Logf("Auto-updated golden file %s due to canonical declaration reordering", goldenPath)
		}
		return
	}

	assert.Equal(t, expectedTrimmed, actualTrimmed)
}

func logGoldenSizeChange(t *testing.T, path string, oldBytes, newBytes int) {
	t.Helper()
	diff := newBytes - oldBytes
	var pct float64
	if oldBytes > 0 {
		pct = (float64(diff) / float64(oldBytes)) * 100.0
	}
	t.Logf("golden size change for %s: %d B -> %d B (%+d B, %+.2f%%)", path, oldBytes, newBytes, diff, pct)
}

func declarationsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, va := range a {
		vb, ok := b[k]
		if !ok {
			return false
		}
		if normalizeBody(va) != normalizeBody(vb) {
			return false
		}
	}
	return true
}

func normalizeBody(s string) string {
	lines := strings.Split(strings.TrimSpace(s), "\n")
	var norm []string
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		if trimmed != "" {
			norm = append(norm, trimmed)
		}
	}
	return strings.Join(norm, "\n")
}

func parseDeclarations(src string) (map[string]string, bool) {
	decls := make(map[string]string)
	i := 0
	n := len(src)

	skipWhitespaceAndComments := func() {
		for i < n {
			if unicode.IsSpace(rune(src[i])) {
				i++
				continue
			}
			if i+1 < n && src[i] == '/' && src[i+1] == '/' {
				for i < n && src[i] != '\n' {
					i++
				}
				continue
			}
			if i+1 < n && src[i] == '/' && src[i+1] == '*' {
				i += 2
				for i+1 < n && !(src[i] == '*' && src[i+1] == '/') {
					i++
				}
				if i+1 < n {
					i += 2
				}
				continue
			}
			break
		}
	}

	for {
		skipWhitespaceAndComments()
		if i >= n {
			break
		}

		start := i
		if strings.HasPrefix(src[i:], "fn ") || strings.HasPrefix(src[i:], "struct ") {
			isFn := strings.HasPrefix(src[i:], "fn ")
			var kind string
			if isFn {
				kind = "fn"
				i += 3
			} else {
				kind = "struct"
				i += 7
			}
			for i < n && unicode.IsSpace(rune(src[i])) {
				i++
			}
			nameStart := i
			for i < n && (src[i] == '_' || unicode.IsLetter(rune(src[i])) || unicode.IsDigit(rune(src[i]))) {
				i++
			}
			name := src[nameStart:i]
			if name == "" {
				return nil, false
			}

			// Find opening brace '{'
			for i < n && src[i] != '{' {
				i++
			}
			if i >= n {
				return nil, false
			}

			braceDepth := 0
			for i < n {
				if src[i] == '{' {
					braceDepth++
				} else if src[i] == '}' {
					braceDepth--
					if braceDepth == 0 {
						i++
						break
					}
				}
				i++
			}
			if braceDepth != 0 {
				return nil, false
			}
			body := strings.TrimSpace(src[start:i])
			decls[kind+":"+name] = body
		} else if strings.HasPrefix(src[i:], "var") {
			for i < n && src[i] != ';' {
				i++
			}
			if i < n && src[i] == ';' {
				i++
			}
			stmt := strings.TrimSpace(src[start:i])
			parts := strings.Fields(stmt)
			varName := ""
			for _, p := range parts {
				if strings.Contains(p, ":") {
					varName = strings.TrimSuffix(p, ":")
					break
				}
			}
			if varName == "" && len(parts) > 1 {
				varName = parts[1]
			}
			decls["var:"+varName] = stmt
		} else {
			lineStart := i
			for i < n && src[i] != '\n' {
				i++
			}
			line := strings.TrimSpace(src[lineStart:i])
			if line != "" {
				decls["raw:"+line] = line
			}
		}
	}

	return decls, len(decls) > 0
}
