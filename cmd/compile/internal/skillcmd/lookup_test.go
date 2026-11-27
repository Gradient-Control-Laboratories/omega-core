package skillcmd

import (
	"strings"
	"testing"
)

func TestNamesAreLookedUp(t *testing.T) {
	out, errOut, code := run(t, "transform.rotate", "builtin/mesh")
	if code != exitPrinted || errOut != "" {
		t.Fatalf("exit %d, stderr %q", code, errOut)
	}
	for _, w := range []string{"`rotate(shape, axisOrigin, axisDirection, angle)`", "### `builtin/mesh`"} {
		if !strings.Contains(out, w) {
			t.Errorf("output lacks %q", w)
		}
	}
}

func TestGrep(t *testing.T) {
	out, errOut, code := run(t, "--grep", "countersink")
	if code != exitPrinted || errOut != "" || !strings.Contains(out, "- **`countersink`**") {
		t.Errorf("--grep countersink: exit %d, stderr %q", code, errOut)
	}
	if _, errOut, code := run(t, "--grep", "zzzqqqnothing"); code != exitNoMatches || errOut == "" {
		t.Errorf("--grep with no match: exit %d, stderr %q; want %d and a message", code, errOut, exitNoMatches)
	}
	for _, args := range [][]string{{"--grep", "x", "wood"}, {"--all", "--grep", "x"}} {
		if _, errOut, code := run(t, args...); code != exitBadArgs || errOut == "" {
			t.Errorf("%v: exit %d, stderr %q; want %d and a message", args, code, errOut, exitBadArgs)
		}
	}
}
