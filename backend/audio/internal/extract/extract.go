package extract

import (
	"fmt"
	"strings"
)

// ExtractAudio breaks the source into lines, finds the last return statement,
// and extracts the audio expression if present.
func ExtractAudio(gslSource string) (string, string, error) {
	lines := strings.Split(gslSource, "\n")
	returnLineIdx := -1
	for i := len(lines) - 1; i >= 0; i-- {
		trimmed := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trimmed, "return") {
			if len(trimmed) == 6 {
				returnLineIdx = i
				break
			}
			nextChar := trimmed[6]
			if nextChar == ' ' || nextChar == '\t' || nextChar == '(' || nextChar == ';' || nextChar == ',' {
				returnLineIdx = i
				break
			}
		}
	}

	if returnLineIdx == -1 {
		return "", "", nil
	}

	// Calculate the starting position of the return statement line in the whole source
	lineStartIdx := 0
	for idx := 0; idx < returnLineIdx; idx++ {
		lineStartIdx += len(lines[idx]) + 1
	}
	returnOffset := strings.Index(lines[returnLineIdx], "return")
	returnStartIdx := lineStartIdx + returnOffset

	// Search for the closing semicolon terminating the return statement
	semiOffset := strings.Index(gslSource[returnStartIdx:], ";")
	if semiOffset == -1 {
		return "", "", nil
	}
	semiIdx := returnStartIdx + semiOffset

	returnStmt := gslSource[returnStartIdx : semiIdx+1]

	// Look for "audio(" inside the return statement
	audioStartOffset := strings.Index(returnStmt, "audio(")
	if audioStartOffset == -1 {
		return "", "", nil
	}

	// Find the matching closing parenthesis for "audio("
	audioEndOffset := -1
	depth := 0
	for i := audioStartOffset + len("audio("); i < len(returnStmt); i++ {
		char := returnStmt[i]
		if char == '(' {
			depth++
		} else if char == ')' {
			if depth == 0 {
				audioEndOffset = i
				break
			}
			depth--
		}
	}

	if audioEndOffset == -1 {
		return "", "", fmt.Errorf("malformed audio(...) expression in return statement")
	}

	// Verify audio(...) is the last element (the remaining string before the semicolon contains only whitespace)
	remaining := returnStmt[audioEndOffset+1 : len(returnStmt)-1]
	if strings.TrimSpace(remaining) != "" {
		return "", "", fmt.Errorf("audio(...) expression is not the last element in the return statement")
	}

	// Extract arguments inside audio(...)
	argsStr := returnStmt[audioStartOffset+len("audio(") : audioEndOffset]
	if strings.TrimSpace(argsStr) == "" {
		return "", "", fmt.Errorf("audio(...) expression does not contain either 1 or 2 arguments")
	}

	// Split arguments respecting nested parentheses and brackets
	args := splitArgs(argsStr)
	for i := range args {
		args[i] = strings.TrimSpace(args[i])
	}

	if len(args) < 1 || len(args) > 2 {
		return "", "", fmt.Errorf("audio(...) expression does not contain either 1 or 2 arguments")
	}

	// Construct audioSource
	audioSource := gslSource[:returnStartIdx] + "return " + strings.Join(args, ", ") + ";" + gslSource[semiIdx+1:]

	// Construct otherSource
	prefix := returnStmt[6:audioStartOffset]
	trimmedPrefix := strings.TrimSpace(prefix)
	if before, ok := strings.CutSuffix(trimmedPrefix, ","); ok {
		trimmedPrefix = before
	}
	trimmedPrefix = strings.TrimSpace(trimmedPrefix)

	var otherSource string
	if trimmedPrefix == "" {
		otherSource = ""
	} else {
		otherSource = gslSource[:returnStartIdx] + "return " + trimmedPrefix + ";" + gslSource[semiIdx+1:]
	}

	return audioSource, otherSource, nil
}

// splitArgs splits a string by comma at depth 0 of parentheses/brackets
func splitArgs(s string) []string {
	var args []string
	var current strings.Builder
	depth := 0
	bracketDepth := 0
	for i := 0; i < len(s); i++ {
		char := s[i]
		if char == '(' {
			depth++
			current.WriteByte(char)
		} else if char == ')' {
			depth--
			current.WriteByte(char)
		} else if char == '[' {
			bracketDepth++
			current.WriteByte(char)
		} else if char == ']' {
			bracketDepth--
			current.WriteByte(char)
		} else if char == ',' && depth == 0 && bracketDepth == 0 {
			args = append(args, current.String())
			current.Reset()
		} else {
			current.WriteByte(char)
		}
	}
	args = append(args, current.String())
	return args
}
