package main

import (
	"strings"
	"testing"
)

// plated's presets bind a local named for the package, and calling a
// constructor of the package qualifies it with that name. The rename of the
// local stops at the local: the template calls plated.nickel, not a member of
// the renamed variable.
func TestAdjustKeepsThePackageOfACall(t *testing.T) {
	op := readOperators(t)["plated.nickel"]
	if op == nil || op.Adjust == nil {
		t.Fatal("plated.nickel has no adjust record")
	}
	if want := "= plated.nickel(${material}, ${bright});"; !strings.Contains(op.Adjust.Template, want) {
		t.Errorf("operator template:\n%s\nwant it to contain %s", op.Adjust.Template, want)
	}

	adj := readAdjust(t)["plated.zincPlatedSteel"]
	if adj == nil {
		t.Fatal("plated.zincPlatedSteel has no adjust record")
	}
	if want := "= plated.zinc(${body}SteelMat, ${bright});"; !strings.Contains(adj.Template, want) {
		t.Errorf("preset template:\n%s\nwant it to contain %s", adj.Template, want)
	}
}
