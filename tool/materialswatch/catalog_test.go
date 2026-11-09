package main

import (
	"strings"
	"testing"
)

// The catalog is read out of the libraries rather than listed here, so the
// tests check that the scan agrees with what the libraries declare — not that
// it finds some particular count, which would have to be edited every time a
// material is added.

func TestScanReadsAnnotatedSurfaces(t *testing.T) {
	src := strings.Join([]string{
		"// ===========================================================================",
		"// Ferrous",
		"// ===========================================================================",
		"",
		"/// Mild steel, machined satin. Returns material and colour.",
		"@material steel() {",
		"  return finish(0.45), [0.77, 0.78, 0.78];",
		"}",
		"",
		"// an internal note, not documentation",
		"helper(x) {",
		"  return x;",
		"}",
		"",
		"/// Fully diffuse: chalk, unglazed clay, flat paint.",
		"@finish matte = make(1.0, 0, 1.45);",
		"",
	}, "\n")

	got := scanSource("metal", src)
	if len(got) != 2 {
		t.Fatalf("scanned %d entries, want 2: %+v", len(got), got)
	}

	if got[0] != (Entry{
		Kind:    "material",
		Package: "metal",
		Symbol:  "steel",
		Section: "Ferrous",
		Doc:     "Mild steel, machined satin.",
	}) {
		t.Errorf("material entry: %+v", got[0])
	}
	if got[1] != (Entry{
		Kind:    "finish",
		Package: "metal",
		Symbol:  "matte",
		Section: "Ferrous",
		Doc:     "Fully diffuse: chalk, unglazed clay, flat paint.",
	}) {
		t.Errorf("finish entry: %+v", got[1])
	}
}

// A doc block is only the /// lines immediately above the declaration. A blank
// line between them means the prose belongs to whatever came before, and must
// not be attached to the next surface.
func TestBlankLineEndsTheDocBlock(t *testing.T) {
	src := "/// Prose about the section.\n\n@material steel() {\n}\n"
	got := scanSource("metal", src)
	if len(got) != 1 {
		t.Fatalf("scanned %d entries, want 1", len(got))
	}
	if got[0].Doc != "" {
		t.Errorf("doc across a blank line = %q, want empty", got[0].Doc)
	}
}

// An indented @material is inside a function body, not a declaration.
func TestIndentedAnnotationIsNotADeclaration(t *testing.T) {
	if got := scanSource("metal", "  @material steel() {\n"); len(got) != 0 {
		t.Errorf("scanned %d entries from an indented line, want 0", len(got))
	}
}

// Every surface the libraries annotate has to reach the catalog, and each has
// to be addressable: a kind the panel knows, a package, and a symbol.
func TestCatalogCoversTheLibraries(t *testing.T) {
	entries := ReadCatalog()
	if len(entries) == 0 {
		t.Fatal("the catalog is empty")
	}

	seen := map[string]bool{}
	byKind := map[string]int{}
	for _, e := range entries {
		if e.Kind != "material" && e.Kind != "finish" {
			t.Errorf("%s: unknown kind %q", e.Call(), e.Kind)
		}
		if e.Package == "" || e.Symbol == "" {
			t.Errorf("entry with no address: %+v", e)
		}
		key := e.Package + "." + e.Symbol
		if seen[key] {
			t.Errorf("%s appears twice in the catalog", key)
		}
		seen[key] = true
		byKind[e.Kind]++
	}
	if byKind["material"] == 0 || byKind["finish"] == 0 {
		t.Errorf("want both kinds in the catalog, got %v", byKind)
	}

	// The libraries this panel was built for, so a package losing its
	// annotations entirely is a failure rather than a quietly shorter grid.
	for _, want := range []string{"metal.brass", "glass.clear", "material.standard"} {
		if !seen[want] {
			t.Errorf("%s is not in the catalog", want)
		}
	}
}

// An entry's index is its cell in the sheet, so it must be its position in the
// catalog exactly — an off-by-one here draws every swatch as its neighbour.
func TestIndexIsThePositionInTheSheet(t *testing.T) {
	for i, e := range ReadCatalog() {
		if e.Index != i {
			t.Fatalf("entry %d (%s) has index %d", i, e.Call(), e.Index)
		}
	}
}

// A material is called and a finish is named: getting this backwards produces a
// document that does not compile, or one that silently drops the colour.
func TestCallFormFollowsTheKind(t *testing.T) {
	if got := (Entry{Kind: "material", Package: "metal", Symbol: "brass"}).Call(); got != "metal.brass()" {
		t.Errorf("material call = %q", got)
	}
	if got := (Entry{Kind: "finish", Package: "material", Symbol: "matte"}).Call(); got != "material.matte" {
		t.Errorf("finish call = %q", got)
	}
}
