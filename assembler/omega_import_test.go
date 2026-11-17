package assembler_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/internal/test/assert"
	"gradientcontrol.com/omega/core/program"
)

func TestAssemble_OmegaDotImportWarning_PrimaryPackage(t *testing.T) {
	src := "import . \"omega\";\nreturn sphere(length(dirX * 5));\n"
	res, err := assembler.Assemble([]string{src})
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}

	var found bool
	for _, diag := range res.Diagnostics {
		if diag.Severity == program.SeverityWarning && strings.Contains(diag.Message, `import . "omega"`) {
			found = true
			assert.Equal(t, "", diag.PackagePath)
			assert.Equal(t, `import . "omega";`, diag.Line)
			assert.Equal(t, 1, diag.LineNum)
			assert.Equal(t, 1, diag.Column)
			assert.Equal(t, len(`import . "omega";`), diag.Length)
			assert.Contains(t, diag.Message, "builtin")
			assert.Contains(t, diag.Message, "gsl fix")
			break
		}
	}
	assert.True(t, found, "expected omega dot-import warning in Result.Diagnostics, got %+v", res.Diagnostics)
}

func TestAssemble_OmegaDotImportWarning_ImportedPackage(t *testing.T) {
	root := t.TempDir()
	depDir := filepath.Join(root, "omegaused")
	if err := os.MkdirAll(depDir, 0o755); err != nil {
		t.Fatal(err)
	}
	depFile := filepath.Join(depDir, "omegaused.gsl")
	depSrc := "// Helper pkg\nimport . \"omega\";\nrad = length(dirX * 4);\n"
	if err := os.WriteFile(depFile, []byte(depSrc), 0o644); err != nil {
		t.Fatal(err)
	}

	primarySrc := "import \"omegaused\";\nreturn sphere(omegaused.rad);\n"
	res, err := assembler.Assemble([]string{primarySrc}, assembler.SourcePaths(root))
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}

	var found bool
	for _, diag := range res.Diagnostics {
		if diag.Severity == program.SeverityWarning && strings.Contains(diag.Message, `import . "omega"`) {
			found = true
			assert.Equal(t, "omegaused", diag.PackagePath)
			assert.Equal(t, `import . "omega";`, diag.Line)
			assert.Equal(t, 2, diag.LineNum)
			assert.Equal(t, 1, diag.Column)
			assert.Equal(t, len(`import . "omega";`), diag.Length)
			break
		}
	}
	assert.True(t, found, "expected imported package omega dot-import warning in Result.Diagnostics, got %+v", res.Diagnostics)
}

func TestAssemble_CleanImportsEmitNoOmegaWarning(t *testing.T) {
	src := "import \"vector\";\nreturn sphere(length(vector.dirX * 5));\n"
	res, err := assembler.Assemble([]string{src})
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}
	if len(res.Diagnostics) != 0 {
		t.Fatalf("expected 0 diagnostics, got %+v", res.Diagnostics)
	}
}
