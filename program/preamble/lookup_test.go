package preamble

import (
	"strings"
	"testing"
)

func TestLookupFindsEachKindOfName(t *testing.T) {
	for _, tc := range []struct {
		name string
		want []string
	}{
		{"wood", []string{"### `material/bulk/wood`"}},
		{"material/bulk/wood", []string{"### `material/bulk/wood`"}},
		{"builtin/mesh", []string{"### `builtin/mesh`", "* **`meshUnion`**"}},
		{"length", []string{"### `length` (a primitive or built-in)"}},
		{"meshUnion", []string{"### `meshUnion` (a built-in, in builtin/mesh)"}},
		{"transform.rotate", []string{"### `transform`", "`rotate(shape, axisOrigin, axisDirection, angle)`"}},
		{"unionRound", []string{"### `builtin`", "`unionRound(radius)`", "`unionRound(a, b, radius)`"}},
		{"bcc", []string{"### `lattice`", "### `lattice/beam`", "### `lattice/foam`"}},
	} {
		got, err := Lookup(tc.name)
		if err != nil {
			t.Errorf("Lookup(%q): %v", tc.name, err)
			continue
		}
		for _, w := range tc.want {
			if !strings.Contains(got, w) {
				t.Errorf("Lookup(%q) lacks %q", tc.name, w)
			}
		}
	}
	for _, bad := range []string{"nope", "transform.nope", "nope.rotate"} {
		if _, err := Lookup(bad); err == nil {
			t.Errorf("Lookup(%q) found something", bad)
		}
	}
}

func TestGrepMatchesNamesAndDocs(t *testing.T) {
	got := Grep("countersink")
	for _, w := range []string{"### `mechanical/drill`", "- **`countersink`**", "### `mechanical/fastener`"} {
		if !strings.Contains(got, w) {
			t.Errorf("Grep(countersink) lacks %q", w)
		}
	}
	if !strings.Contains(Grep("LENGTH OF VECTOR"), "- **`length`**") {
		t.Error("Grep is not case-insensitive over built-in docs")
	}
	if Grep("zzzqqqnothing") != "" {
		t.Error("Grep matched nonsense")
	}
}

// TestFullDocsKeepEveryFormsDoc: grouping a name's forms under one entry
// keeps each form's own doc.
func TestFullDocsKeepEveryFormsDoc(t *testing.T) {
	doc, err := PackageDescriptiveText("builtin")
	if err != nil {
		t.Fatal(err)
	}
	for _, w := range []string{
		"- **`sphere`**\n",
		"  - `sphere(center, radius, pole, reference)` — ",
		"measured about `pole` from `reference`",
		"  - `unionRound(radius)` — Partial application",
	} {
		if !strings.Contains(doc, w) {
			t.Errorf("builtin docs lack %q", w)
		}
	}
}

// TestBriefIsOneLinePerName: in the index's brief a name appears on one line
// with all its forms.
func TestBriefIsOneLinePerName(t *testing.T) {
	brief := builtinBriefText()
	if n := strings.Count(brief, "- **`unionRound`**"); n != 1 {
		t.Errorf("unionRound has %d lines in the brief, want 1", n)
	}
	for _, w := range []string{"`(a, b, radius)`", "`(fields, radius)`", "`(radius)`"} {
		if !strings.Contains(brief, w) {
			t.Errorf("the brief lacks unionRound form %s", w)
		}
	}
}
