package link_test

import (
	"strings"
	"testing"
	"time"

	"gradientcontrol.com/omega/core/assembler/source"
	"gradientcontrol.com/omega/core/internal/collect"
	"gradientcontrol.com/omega/core/internal/link"
	internalsource "gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/test/assert"
	parser "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

func newMockManager() source.Manager {
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

func parseAndLink(t *testing.T, src string, mgr source.Manager) ([]program.Diagnostic, error) {
	t.Helper()
	prog, err := collect.Program([]string{src}, mgr, parser.ParsePackage)
	if err != nil {
		return nil, err
	}
	return link.ReferencesAndDiagnostics(prog, mgr, parser.ParsePackage)
}

func TestDotImportVariableCollisionWarning(t *testing.T) {
	src := `
import . "pkgA";
import . "pkgB";
return pkgA.x + pkgB.x;
`
	mgr := newMockManager()
	diags, err := parseAndLink(t, src, mgr)
	assert.NoError(t, err)

	var found bool
	for _, diag := range diags {
		if diag.Severity == program.SeverityWarning &&
			strings.Contains(diag.Message, `variable "x" declared in both import "pkgA" and "pkgB"`) {
			found = true
			break
		}
	}
	assert.True(t, found, "expected warning diagnostic for variable collision")
}

func TestDotImportFunctionCollisionWarning(t *testing.T) {
	src := `
import . "pkgA";
import . "pkgB";
return pkgA.foo(1.0, 2.0) + pkgB.foo(3.0, 4.0);
`
	mgr := newMockManager()
	diags, err := parseAndLink(t, src, mgr)
	assert.NoError(t, err)

	var found bool
	for _, diag := range diags {
		if diag.Severity == program.SeverityWarning &&
			strings.Contains(diag.Message, `function "foo" with 2 parameters declared in both import "pkgA" and "pkgB"`) {
			found = true
			break
		}
	}
	assert.True(t, found, "expected warning diagnostic for function collision")
}

func TestBuiltinFunctionShadowingWarning(t *testing.T) {
	src := `
min(a, b) {
    return a + b;
}
return min(1.0, 2.0);
`
	mgr := newMockManager()
	diags, err := parseAndLink(t, src, mgr)
	assert.NoError(t, err)

	var found bool
	for _, diag := range diags {
		if diag.Severity == program.SeverityWarning &&
			strings.Contains(diag.Message, `function "min" with 2 parameters shadows implicitly imported builtin function`) {
			found = true
			break
		}
	}
	assert.True(t, found, "expected warning diagnostic for shadowing builtin function")
}

func TestBuiltinFunctionNoWarningForNonBuiltin(t *testing.T) {
	src := `
customFn(a, b) {
    return a + b;
}
return customFn(1.0, 2.0);
`
	mgr := newMockManager()
	diags, err := parseAndLink(t, src, mgr)
	assert.NoError(t, err)

	for _, diag := range diags {
		if strings.Contains(diag.Message, "shadows") {
			t.Errorf("unexpected shadowing warning: %v", diag)
		}
	}
}

func TestLocalDeclarationCollisionError(t *testing.T) {
	t.Run("variable_collision", func(t *testing.T) {
		src := `
import . "pkgA";
x = 42.0;
return x;
`
		mgr := newMockManager()
		_, err := parseAndLink(t, src, mgr)
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
		mgr := newMockManager()
		_, err := parseAndLink(t, src, mgr)
		assert.Error(t, err)
		assert.True(t, strings.Contains(err.Error(), `variable or function named "foo" already defined`),
			"unexpected error: %v", err)
	})
}

func TestAmbiguousReferenceError(t *testing.T) {
	src := `
import . "pkgA";
import . "pkgB";
return x;
`
	mgr := newMockManager()
	_, err := parseAndLink(t, src, mgr)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), `ambiguous reference to variable "x": defined in import "pkgA" and "pkgB"`),
		"unexpected error: %v", err)
}

func TestAmbiguousCallError(t *testing.T) {
	src := `
import . "pkgA";
import . "pkgB";
return foo(1.0, 2.0);
`
	mgr := newMockManager()
	_, err := parseAndLink(t, src, mgr)
	assert.Error(t, err)
	assert.True(t, strings.Contains(err.Error(), `ambiguous call to "foo" with 2 arguments: defined in import "pkgA" and "pkgB"`),
		"unexpected error: %v", err)
}

func TestReferencesDelegationDiscardsDiagnostics(t *testing.T) {
	src := `
import . "pkgA";
import . "pkgB";
return pkgA.x + pkgB.x;
`
	mgr := newMockManager()
	prog, err := collect.Program([]string{src}, mgr, parser.ParsePackage)
	assert.NoError(t, err)

	// Calling link.References succeeds with no error even when warnings are generated
	err = link.References(prog, mgr, parser.ParsePackage)
	assert.NoError(t, err)
}

func TestFastExitSingleOrZeroDotImports(t *testing.T) {
	src := `
x = 1.0;
return x;
`
	mgr := newMockManager()
	prog, err := collect.Program([]string{src}, mgr, parser.ParsePackage)
	assert.NoError(t, err)

	start := time.Now()
	for i := 0; i < 1000; i++ {
		_, _ = link.ReferencesAndDiagnostics(prog, mgr, parser.ParsePackage)
	}
	elapsed := time.Since(start)
	// 1000 runs should complete very quickly (< 1 second under normal load)
	assert.True(t, elapsed < 2*time.Second, "expected fast execution for single dot-import program")
}
