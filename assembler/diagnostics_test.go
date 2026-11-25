package assembler_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/assembler"
	"github.com/Gradient-Control-Laboratories/omega-core/assembler/source"
	internalsource "github.com/Gradient-Control-Laboratories/omega-core/internal/source"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/test/assert"
	"github.com/Gradient-Control-Laboratories/omega-core/program"
)

func newTestMockManager() source.Manager {
	return &internalsource.MockManager{
		PkgSources: map[string][]string{
			"pkgA": {`
x = 10.0;
foo(a, b) {
    return a + b;
}
return x;
`},
			"pkgB": {`
x = 20.0;
foo(a, b) {
    return a * b;
}
return x;
`},
		},
	}
}

func TestUserFunctionShadowingBuiltinWarning(t *testing.T) {
	src := `
min(a, b) {
    return a + b;
}
return min(1.0, 2.0);
`
	mgr := newTestMockManager()
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(mgr))
	assert.NoError(t, err)

	var found bool
	for _, diag := range res.Diagnostics {
		if diag.Severity == program.SeverityWarning &&
			strings.Contains(diag.Message, `function "min" with 2 parameters shadows implicitly imported builtin function`) {
			found = true
			break
		}
	}
	assert.True(t, found, "expected warning diagnostic in res.Diagnostics for shadowing builtin function")
}

func TestUserFunctionNoWarningForNonBuiltin(t *testing.T) {
	src := `
customAdd(a, b) {
    return a + b;
}
return customAdd(1.0, 2.0);
`
	mgr := newTestMockManager()
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(mgr))
	assert.NoError(t, err)

	for _, diag := range res.Diagnostics {
		if strings.Contains(diag.Message, "shadows") {
			t.Errorf("unexpected shadowing warning: %v", diag)
		}
	}
}

func TestDotImportVariableCollisionWarning(t *testing.T) {
	src := `
import . "pkgA";
import . "pkgB";
return pkgA.x + pkgB.x;
`
	mgr := newTestMockManager()
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(mgr))
	assert.NoError(t, err)

	var found bool
	for _, diag := range res.Diagnostics {
		if diag.Severity == program.SeverityWarning &&
			strings.Contains(diag.Message, `variable "x" declared in both import "pkgA" and "pkgB"`) {
			found = true
			break
		}
	}
	assert.True(t, found, "expected warning in res.Diagnostics for colliding variable in dot-imports")
}

func TestUnqualifiedCollidingVariableAmbiguityError(t *testing.T) {
	src := `
import . "pkgA";
import . "pkgB";
return x;
`
	mgr := newTestMockManager()
	_, err := assembler.Assemble([]string{src}, assembler.SourceManager(mgr))
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), `ambiguous reference to variable "x": defined in import "pkgA" and "pkgB"`),
		"unexpected error: %v", err)
}

func TestDotImportFunctionCollisionWarning(t *testing.T) {
	src := `
import . "pkgA";
import . "pkgB";
return pkgA.foo(1.0, 2.0) + pkgB.foo(3.0, 4.0);
`
	mgr := newTestMockManager()
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(mgr))
	assert.NoError(t, err)

	var found bool
	for _, diag := range res.Diagnostics {
		if diag.Severity == program.SeverityWarning &&
			strings.Contains(diag.Message, `function "foo" with 2 parameters declared in both import "pkgA" and "pkgB"`) {
			found = true
			break
		}
	}
	assert.True(t, found, "expected warning in res.Diagnostics for colliding function in dot-imports")
}

func TestUnqualifiedCollidingFunctionAmbiguityError(t *testing.T) {
	src := `
import . "pkgA";
import . "pkgB";
return foo(1.0, 2.0);
`
	mgr := newTestMockManager()
	_, err := assembler.Assemble([]string{src}, assembler.SourceManager(mgr))
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), `ambiguous call to "foo" with 2 arguments: defined in import "pkgA" and "pkgB"`),
		"unexpected error: %v", err)
}

func TestQualifiedCollidingVariableAndFunctionSuccess(t *testing.T) {
	src := `
import . "pkgA";
import . "pkgB";
return pkgA.x + pkgB.x + pkgA.foo(1.0, 2.0) + pkgB.foo(3.0, 4.0);
`
	mgr := newTestMockManager()
	_, err := assembler.Assemble([]string{src}, assembler.SourceManager(mgr))
	assert.NoError(t, err)
}

func TestExplicitDotImportLocalDeclarationCollisionError(t *testing.T) {
	t.Run("variable_collision", func(t *testing.T) {
		src := `
import . "pkgA";
x = 42.0;
return x;
`
		mgr := newTestMockManager()
		_, err := assembler.Assemble([]string{src}, assembler.SourceManager(mgr))
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), `variable or function named "x" already defined`),
			"unexpected error: %v", err)
	})

	t.Run("function_collision", func(t *testing.T) {
		src := `
import . "pkgA";
foo(a, b) {
    return a;
}
return foo(1.0, 2.0);
`
		mgr := newTestMockManager()
		_, err := assembler.Assemble([]string{src}, assembler.SourceManager(mgr))
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), `variable or function named "foo" already defined`),
			"unexpected error: %v", err)
	})
}

func TestAssembler_PropagatesParserDiagnostics(t *testing.T) {
	src := `// Scene Size: 20
return sphere(5);
`
	res, err := assembler.Assemble([]string{src}, assembler.SourceManager(newTestMockManager()))
	assert.NoError(t, err)

	var found bool
	for _, diag := range res.Diagnostics {
		if diag.Severity == program.SeverityWarning &&
			strings.Contains(diag.Message, "@scene") &&
			strings.Contains(diag.Message, "gsl fix") {
			found = true
			assert.Equal(t, 1, diag.LineNum)
			break
		}
	}
	assert.True(t, found, "expected parser legacy scene warning in assembler Result.Diagnostics, got %+v", res.Diagnostics)
}

func TestAssembler_PropagatesImportedPackageParserDiagnostics(t *testing.T) {
	root := t.TempDir()
	depDir := filepath.Join(root, "legacydep")
	if err := os.MkdirAll(depDir, 0o755); err != nil {
		t.Fatal(err)
	}
	depFile := filepath.Join(depDir, "legacydep.gsl")
	depSrc := "// Scene Size: 30\nrad = 6.0;\n"
	if err := os.WriteFile(depFile, []byte(depSrc), 0o644); err != nil {
		t.Fatal(err)
	}

	primarySrc := "import \"legacydep\";\nreturn sphere(legacydep.rad);\n"
	res, err := assembler.Assemble([]string{primarySrc}, assembler.SourcePaths(root))
	if err != nil {
		t.Fatalf("Assemble failed: %v", err)
	}

	var found bool
	for _, diag := range res.Diagnostics {
		if diag.Severity == program.SeverityWarning && strings.Contains(diag.Message, "@scene") {
			found = true
			assert.Equal(t, "legacydep", diag.PackagePath)
			assert.Equal(t, "// Scene Size: 30", diag.Line)
			assert.Equal(t, 1, diag.LineNum)
			break
		}
	}
	assert.True(t, found, "expected imported package parser warning in assembler Result.Diagnostics, got %+v", res.Diagnostics)
}
