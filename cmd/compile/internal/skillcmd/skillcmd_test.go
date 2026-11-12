package skillcmd

import (
	"bytes"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/program/preamble"
)

func run(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := Run(args, &stdout, &stderr)
	return stdout.String(), stderr.String(), code
}

func TestIndexNamesEveryPackageAndHowToReadOn(t *testing.T) {
	out, errOut, code := run(t)
	if code != exitPrinted || errOut != "" {
		t.Fatalf("exit %d, stderr %q", code, errOut)
	}
	for _, name := range preamble.PackageNames() {
		if !strings.Contains(out, "- **`"+name+"`**") {
			t.Errorf("the index has no line for %s", name)
		}
	}
	for _, want := range []string{"<!-- GSL reference from gsl ", "## **Primitive and Built-in Functions**", "### Package Index", "gsl skill PACKAGE..."} {
		if !strings.Contains(out, want) {
			t.Errorf("the index lacks %q", want)
		}
	}
	// The point of the index: a model can read it whole.
	if full := preamble.GSLAuthorSkill(); len(out) > len(full)/4 {
		t.Errorf("the index is %d bytes, over a quarter of the full reference's %d", len(out), len(full))
	}
}

func TestPackagesByPathOrPrefix(t *testing.T) {
	byPrefix, _, code := run(t, "wood")
	if code != exitPrinted {
		t.Fatalf("exit %d", code)
	}
	byPath, _, _ := run(t, "material/bulk/wood")
	if byPrefix != byPath {
		t.Error("a package named by prefix and by import path printed differently")
	}
	two, _, _ := run(t, "builtin", "wood")
	if !strings.Contains(two, "### `builtin`") || !strings.Contains(two, "### `material/bulk/wood`") {
		t.Error("naming two packages did not print both")
	}
}

func TestAllIsTheWholeReference(t *testing.T) {
	out, _, code := run(t, "--all")
	if code != exitPrinted || !strings.HasSuffix(out, preamble.GSLAuthorSkill()) {
		t.Fatalf("exit %d; --all is not GSLAuthorSkill", code)
	}
}

func TestBadArguments(t *testing.T) {
	for _, args := range [][]string{{"nope"}, {"--all", "wood"}, {"--bogus"}} {
		if _, errOut, code := run(t, args...); code != exitBadArgs || errOut == "" {
			t.Errorf("%v: exit %d, stderr %q; want %d and a message", args, code, errOut, exitBadArgs)
		}
	}
	if out, _, code := run(t, "--help"); code != exitPrinted || !strings.Contains(out, "Usage of compile skill") {
		t.Errorf("--help: exit %d", code)
	}
}
