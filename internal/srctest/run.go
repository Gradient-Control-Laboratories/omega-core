package src_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"gradientcontrol.com/omega/core/program"
)

// Find all files matching [glob] in the current directory.
// For each file, run a subtest using the [testFunc] function.
func RunAllSourceTests(t *testing.T, glob string, testFunc func(*testing.T, string)) {
	targets, err := filepath.Glob(glob)
	if err != nil {
		t.Fatalf("Failed to find test files: %v", err)
	}
	fileCount := 0
	for _, target := range targets {
		t.Run(filepath.Base(target), func(t *testing.T) {
			testFunc(t, target)
		})
		fileCount++

	}
	if fileCount == 0 {
		t.Fatalf("No test files found for glob %s", glob)
	}
	t.Logf("%d test files in %s", fileCount, glob)
}

// Read the content of [srcFile] and pass it to the [testFunc] function.
// If the first line of [srcFile] starts with "// ERROR:"
// then assert that [testFunc] has returned an error,
// check the error message, and return
//
// Otherwise if the [testFunc] function returns an error,
// then then call t.Errorf to report the error and return.
// Call the [checkResult] function to check the result of testFunc
// and report any error returned by checkResult.
//
// If [testFunc] returns nil, then
// scan each line of the source file and check for comments that contain "ERROR:"
//
// For each line that contains "ERROR:"
// 1. Scan the line for the number after "ERROR:"
// 2. If the line starts with "//", then uncomment the line, otherwise comment out the line
// 3. Run a subtest named "error-<number>"
// 4. In that subtest, call [testFunc] with the modified source code.
// 5. If [testFunc] does not return an error, then call t.Errorf to report the missing error and return.
//
//  6. If [testFunc] returns an error, then
//     6a. If the expected error message is an empty string,
//     then update that line in the [srcFile] with the actual error message.
//     6b. If the expected error message is not an empty string
//     and the actual error message does not contain the expected error message,
//     then call t.Errorf to report the error.
func RunSourceTest[R any](t *testing.T, srcFile string,
	// The function to run with the source code
	testFunc func(t *testing.T, srcFile string, src []byte) (R, error),
	// The function to check the result of testFunc
	checkResult func(t *testing.T, srcFile string, src []byte, result R) error,
	// If true, then verify the error message
	verifyErrMsg bool,
) {
	src, err := os.ReadFile(srcFile)
	if err != nil {
		t.Fatalf("Failed to read source file %s: %v", srcFile, err)
	}

	result, err := testFunc(t, srcFile, src)

	// If the first line of the source file starts with "// ERROR:"
	// then assert that an error was returned
	lines := strings.Split(string(src), "\n")
	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "// ERROR:") {
		_, expectedMsg, _ := sourceWithError(t, lines, 0, strings.Index(lines[0], "ERROR:"))
		if err == nil {
			t.Errorf("Expected error %q in %s, but got no error", expectedMsg, srcFile)
		} else if !strings.Contains(err.Error(), expectedMsg) {
			t.Errorf("Expected error %q in %s, got %q", expectedMsg, srcFile, err.Error())
		}
		return
	}

	if err == nil {
		err = checkResult(t, srcFile, src, result)
	}
	if err != nil {
		t.Fatalf("Unexpected error in %s\n  %v", srcFile, err)
	}

	// Avoid race condition when multiple subtests try to update the same source file
	updateSourceMutex := sync.Mutex{}

	baseName := filepath.Base(srcFile)
	for lineIdx, line := range lines {
		if !strings.Contains(line, "ERROR:") {
			continue
		}

		errorIdx := strings.Index(line, "ERROR:")
		errorNum, expectedMsg, newSource := sourceWithError(t, lines, lineIdx, errorIdx)

		updateSource := func(actualMsg string, targetLine int) {
			updateSourceMutex.Lock()
			defer updateSourceMutex.Unlock()

			actualMsg = strings.ReplaceAll(actualMsg, "\n", " ")
			updatedLine := line[:errorIdx] + fmt.Sprintf("ERROR: %s %s", errorNum, actualMsg)
			lines[targetLine-1] = updatedLine // 1 based error line number
			if err := os.WriteFile(srcFile, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
				t.Fatalf("Failed to update source file: %v", err)
			}
		}

		t.Run(fmt.Sprintf("error-%s", errorNum), func(t *testing.T) {
			_, err := testFunc(t, srcFile, []byte(newSource))
			if err == nil {
				t.Errorf("Expected error on line %d matching %q, but got no errors", lineIdx+1, expectedMsg)
				return
			}

			targetLineIdx := lineIdx + 1
			actualMsg := err.Error()
			if buildErr, isBuildErr := err.(*program.Error); isBuildErr {
				actualMsg = buildErr.Message
			}

			if expectedMsg == "" {
				updateSource(actualMsg, targetLineIdx)
				t.Logf("Updated %s line %d with actual error message", baseName, targetLineIdx)
			} else if verifyErrMsg && !strings.Contains(strings.ReplaceAll(actualMsg, "\n", " "), expectedMsg) {
				t.Errorf("Expected error matching %q, got %q", expectedMsg, actualMsg)
			}

			if buildErr, isBuildErr := err.(*program.Error); isBuildErr {
				assertErrorLineNum(t, buildErr, lines, targetLineIdx)
			}
		})
	}
}

// Extract the error number and expected error message from the given line.
// Return the error number, expected error message, and the source code with the error line uncommented.
func sourceWithError(t *testing.T, lines []string, lineIdx int, errorIdx int) (string, string, string) {
	line := lines[lineIdx]
	rest := strings.TrimSpace(line[errorIdx+6:]) // +6 for "ERROR:"
	parts := strings.SplitN(rest, " ", 2)
	if len(parts) == 0 {
		t.Fatalf("Skipping malformed ERROR line: %s", line)
	}
	errorNum := parts[0]
	expectedMsg := ""
	if len(parts) > 1 {
		expectedMsg = strings.TrimSpace(parts[1])
	}

	originalLine := lines[lineIdx]
	if strings.HasPrefix(strings.TrimSpace(originalLine), "//") {
		// Uncomment the line
		commentIdx := strings.Index(originalLine, "//")
		lines[lineIdx] = originalLine[:commentIdx] + strings.TrimSpace(originalLine[commentIdx+2:errorIdx])
	} else {
		// Comment out the line
		lines[lineIdx] = "// " + originalLine
	}
	newSource := strings.Join(lines, "\n")
	lines[lineIdx] = originalLine
	return errorNum, expectedMsg, newSource
}

// Assert that the error is on the expected line.
func assertErrorLineNum(t *testing.T, buildErr *program.Error, lines []string, targetLine int) {
	// Find the next line of code
	nextLine := targetLine
	for nextLine < len(lines) {
		line := strings.TrimSpace(lines[nextLine])
		if line != "" && !strings.HasPrefix(line, "//") {
			break
		}
		nextLine++
	}
	nextLine++ // 1-based error line number

	if buildErr.LineNum < targetLine || buildErr.LineNum > nextLine {
		t.Errorf("Expected error on line %d, got error on line %d message %q", targetLine, buildErr.LineNum, buildErr.Message)
	}
}
