package preamble

import (
	"strings"
	"testing"
)

func TestExtractLibraryEntries_SectionHeader(t *testing.T) {
	src := `// =========================================================
// UTILITIES
// =========================================================

/// Linear interpolation.
mix(a, b, t) {
  return a + (b - a) * t;
}
`
	got := extractLibraryEntries(src)
	if len(got.sections) != 1 {
		t.Fatalf("expected 1 section, got %d", len(got.sections))
	}
	if got.sections[0].title != "UTILITIES" {
		t.Errorf("section title: got %q want UTILITIES", got.sections[0].title)
	}
	if len(got.sections[0].decls) != 1 || got.sections[0].decls[0].signature != "mix(a, b, t)" {
		t.Errorf("decls: %+v", got.sections[0].decls)
	}
}

func TestExtractLibraryEntries_ConstantWithRHS(t *testing.T) {
	src := `pi = atan(1) * 4;
zero = [0, 0, 0];
`
	got := extractLibraryEntries(src)
	if len(got.constants) != 2 {
		t.Fatalf("expected 2 constants, got %d", len(got.constants))
	}
	if got.constants[0].signature != "pi" || got.constants[0].value != "atan(1) * 4" {
		t.Errorf("pi: %+v", got.constants[0])
	}
	if got.constants[1].signature != "zero" || got.constants[1].value != "[0, 0, 0]" {
		t.Errorf("zero: %+v", got.constants[1])
	}
}

func TestExtractLibraryEntries_FunctionDoc(t *testing.T) {
	src := `// =========================================================
// 3D PRIMITIVES
// =========================================================

/// Sphere at origin with the given radius. Uses length(p).
sphere(radius) {
  return length(p) - radius;
}

// Internal note about the cube implementation — not for users.

box(a, b) {
  return 0;
}
`
	got := extractLibraryEntries(src)
	// box has no `///` block so it's treated as internal and dropped;
	// only sphere shows up.
	if len(got.sections) != 1 || len(got.sections[0].decls) != 1 {
		t.Fatalf("section/decl count wrong: %+v", got.sections)
	}
	sphere := got.sections[0].decls[0]
	if sphere.signature != "sphere(radius)" {
		t.Errorf("sphere sig: %q", sphere.signature)
	}
	if !strings.Contains(sphere.doc, "Sphere at origin") {
		t.Errorf("sphere doc: %q", sphere.doc)
	}
}

// Internal `//` notes between a `///` block and a function don't break
// the attachment — the public doc still reaches the function.
func TestExtractLibraryEntries_InternalNoteDoesNotBreakDoc(t *testing.T) {
	src := `// MISC
// =========================================================

/// Public summary.
/// More public detail.
// Internal implementation note — should NOT appear in preamble.
foo(x) {
  return x;
}
`
	got := extractLibraryEntries(src)
	// foo lacks a preceding `// =====` so it ends up in preamble.
	// Look for foo in either sections or preamble.
	var found bool
	check := func(d declEntry) {
		if d.signature != "foo(x)" {
			return
		}
		found = true
		if !strings.Contains(d.doc, "Public summary") {
			t.Errorf("missing public summary: %q", d.doc)
		}
		if !strings.Contains(d.doc, "More public detail") {
			t.Errorf("missing public detail: %q", d.doc)
		}
		if strings.Contains(d.doc, "Internal implementation") {
			t.Errorf("internal note leaked into preamble: %q", d.doc)
		}
	}
	for _, s := range got.sections {
		for _, d := range s.decls {
			check(d)
		}
	}
	for _, d := range got.preamble {
		check(d)
	}
	if !found {
		t.Errorf("foo not found: %+v", got)
	}
}

// A function with only bare `//` notes (no `///`) is treated as internal
// and dropped from the extracted entries entirely.
func TestExtractLibraryEntries_BareCommentNotExtracted(t *testing.T) {
	src := `// MISC
// =========================================================

// This is an internal note.
// Spanning multiple lines.
hidden(x) {
  return x;
}
`
	got := extractLibraryEntries(src)
	for _, s := range got.sections {
		for _, d := range s.decls {
			if d.signature == "hidden(x)" {
				t.Errorf("internal function leaked into extracted entries: %+v", d)
			}
		}
	}
	for _, d := range got.preamble {
		if d.signature == "hidden(x)" {
			t.Errorf("internal function leaked into preamble: %+v", d)
		}
	}
}

func TestExtractLibraryEntries_IgnoreFunctionBodyLines(t *testing.T) {
	// Lines inside a function body must not be picked up as top-level
	// declarations even when they happen to start with `name(`.
	src := `// =========================================================
// MISC
// =========================================================

/// outer doc.
outer(x) {
  inner(x);
  y = 1;
  return y;
}
`
	got := extractLibraryEntries(src)
	if len(got.sections) != 1 || len(got.sections[0].decls) != 1 {
		t.Fatalf("got %d sections with decls %v", len(got.sections), got.sections)
	}
	if got.sections[0].decls[0].signature != "outer(x)" {
		t.Errorf("sig: %q", got.sections[0].decls[0].signature)
	}
}

func TestExtractLibraryEntries_MultiLineSignature(t *testing.T) {
	src := `// MISC
// =========================================================
// (Missing opening divider — intentionally — to test the bare-divider
// path doesn't crash anything.)

/// Pair of meshing gears.
gearPair(numTeeth1, numTeeth2,
         moduleVal, pressureAngle,
         thickness) {
  return 0;
}
`
	got := extractLibraryEntries(src)
	// gearPair has no preceding `// =====` so it sits in the implicit
	// preamble — that's fine, the test is mainly that the signature
	// extractor joined the wrapped lines.
	for _, s := range got.sections {
		for _, d := range s.decls {
			if strings.HasPrefix(d.signature, "gearPair") {
				if !strings.Contains(d.signature, "thickness") {
					t.Errorf("wrapped sig truncated: %q", d.signature)
				}
				return
			}
		}
	}
	for _, d := range got.preamble {
		if strings.HasPrefix(d.signature, "gearPair") && strings.Contains(d.signature, "thickness") {
			return
		}
	}
	t.Errorf("gearPair not found: %+v", got)
}

func TestLibrariesDescriptiveText_NonEmpty(t *testing.T) {
	out := LibrariesDescriptiveText()
	if !strings.Contains(out, "## **Library Packages**") {
		t.Errorf("missing top header in:\n%s", out)
	}
	if !strings.Contains(out, "### `builtin`") {
		t.Errorf("missing builtin header")
	}
	if !strings.Contains(out, "**`pi = atan(1) * 4`**") {
		t.Errorf("missing pi constant entry")
	}
}

// The text library exposes a small public surface (fontStroke,
// fontSpacing, two left() overloads); the ~540 internal per-glyph and
// dispatch helpers must not leak into the generated docs.
func TestLibrariesDescriptiveText_TextOnlyPublicAPI(t *testing.T) {
	out := LibrariesDescriptiveText()
	if !strings.Contains(out, "### `text`") {
		t.Fatalf("missing text header")
	}
	want := []string{
		"`fontStroke()`",
		"`fontSpacing()`",
		"`left(codes, position, size, thickness, spacing)`",
		"`left(codes, position, size)`",
	}
	for _, w := range want {
		if !strings.Contains(out, w) {
			t.Errorf("text section missing %s", w)
		}
	}
	// Internal helpers must stay out — a per-glyph leak would balloon the
	// preamble with hundreds of bare signatures.
	forbidden := []string{
		"glyph_0021",
		"glyph_dispatch_",
		"glyphWidth_dispatch_",
		"segmentDist(",
		"arcDist(",
	}
	for _, f := range forbidden {
		if strings.Contains(out, f) {
			t.Errorf("internal helper leaked into preamble: %s", f)
		}
	}
}

func TestLibrariesDescriptiveText_GeneralLibrariesWithDomain(t *testing.T) {
	general := LibrariesDescriptiveText()
	withDomain := LibrariesDescriptiveText("product")

	for _, body := range []string{general, withDomain} {
		if !strings.Contains(body, "### `builtin`") || !strings.Contains(body, "### `ease`") {
			t.Errorf("general libraries (builtin, ease) must appear regardless of domain")
		}
	}
}
