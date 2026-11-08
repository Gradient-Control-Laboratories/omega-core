package wgsltest_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/backend/wgsl/internal/wgsltest"
	"gradientcontrol.com/omega/core/internal/test/assert"
)

func TestAssertWGSLEqual_Matches(t *testing.T) {
	tmpDir := t.TempDir()
	goldenPath := filepath.Join(tmpDir, "test.wgsl")

	content := "fn foo() -> f32 {\n  return 1.0;\n}\n"
	err := os.WriteFile(goldenPath, []byte(content), 0644)
	assert.NoError(t, err)

	wgsltest.AssertWGSLEqual(t, goldenPath, content, false)
}

func TestAssertWGSLEqual_CreatesMissingGolden(t *testing.T) {
	tmpDir := t.TempDir()
	goldenPath := filepath.Join(tmpDir, "new.wgsl")

	content := "fn bar() -> f32 {\n  return 2.0;\n}"
	// Calling on non-existent golden skips and creates it.
	wgsltest.AssertWGSLEqual(t, goldenPath, content, false)

	readBack, err := os.ReadFile(goldenPath)
	assert.NoError(t, err)
	assert.Equal(t, strings.TrimSpace(content)+"\n", string(readBack))
}
