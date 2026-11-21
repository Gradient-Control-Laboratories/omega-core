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

func TestAssertWGSLEqual_FloatTolerance(t *testing.T) {
	tmpDir := t.TempDir()
	goldenPath := filepath.Join(tmpDir, "float_tol.wgsl")

	golden := "fn foo() -> vec3<f32> {\n  return vec3<f32>(0.79895276, 0.00000000006467769, 3.1415927);\n}\n"
	assert.NoError(t, os.WriteFile(goldenPath, []byte(golden), 0644))

	// 1-2 ULP cross-arch difference near zero and in f32 tail should pass without rewriting golden.
	actual := "fn foo() -> vec3<f32> {\n  return vec3<f32>(0.79895276, 0.00000000006467782, 3.1415925);\n}\n"
	wgsltest.AssertWGSLEqual(t, goldenPath, actual, true)

	readBack, err := os.ReadFile(goldenPath)
	assert.NoError(t, err)
	assert.Equal(t, golden, string(readBack))
}
