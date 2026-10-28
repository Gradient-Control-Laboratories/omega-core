package preamble

import (
	"fmt"
	"regexp"
	"strings"

	gsllibs "gradientcontrol.com/omega/core/stdlib"
)

// LibrariesDescriptiveText scans every embedded GSL library, extracts each
// top-level declaration along with any leading /// doc block, and emits a
// markdown section for the AI preamble. Authors get a single source of
// truth: edit a library's .gsl, and the description an AI sees follows.
//
// Doc convention (Rust-style):
//
//   - `///` lines are public docs — extracted into the preamble and the
//     Libraries panel.
//   - Plain `//` lines are internal notes — they stay in the source for
//     library authors but never reach the preamble.
//   - A blank line ends a doc block; an internal `//` line in the
//     middle of one is ignored (doesn't break or appear).
//
// What's extracted from each library:
//   - Section headers framed as `// =====` / `// TITLE` / `// =====` —
//     used as sub-headings to group the entries that follow. (Sections
//     use plain `//` because they're visual structure, not documentation.)
//   - Top-level constants (`name = expr;` at column 0) — the right-hand
//     side is shown so a reader can tell at a glance that pi ≈ 3.14.
//   - Top-level function signatures (`name(args) { ... }` at column 0)
//     that carry a leading `///` doc block. Functions WITHOUT a `///`
//     block are treated as internal helpers and skipped — that's how a
//     library marks something as "implementation, not API." A generator
//     like cxf2gsl uses this to keep hundreds of per-glyph helpers out
//     of the user-facing docs while still exposing its `text()` entry
//     points.
//
// Indented declarations (function bodies) and stray `//` lines outside
// sections are skipped; they're either internal or pure visual structure.
//
// includeDomains selects which product-specific libraries to document: a
// library with a non-empty gsllibs.PkgInfo.Domain only appears when its domain
// is listed. The general libraries (empty Domain) always appear. Pass no
// domains for the general set.
func containsStr(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

func LibrariesDescriptiveText(includeDomains ...string) string {
	var sb strings.Builder
	sb.WriteString("\n---\n\n## **Library Packages**\n\n")
	sb.WriteString("**Call these rather than writing the same code yourself.** " +
		"The library's versions are SDF-correct, consistently named, and " +
		"tested. If you need a helper the library doesn't provide, name it " +
		"in the library's style (e.g. `cylinderHelical`, not `helicalCyl`) " +
		"so it composes cleanly with the rest.\n\n")
	sb.WriteString("### Core & Standard Packages\n\n")
	// The auto-imported set is read from the library registry rather than
	// listed here, so this text cannot drift from what the assembler resolves.
	prefixes := gsllibs.Prefixes()
	quoted := make([]string, len(prefixes))
	for i, prefix := range prefixes {
		quoted[i] = "`" + prefix + "`"
	}
	sb.WriteString("The core package `builtin` is implicitly dot-imported for every " +
		"file, exposing basic shapes and operations (like `sphere(1)`, `mix(a, b, t)`) " +
		"without any import statement. All other standard GSL packages (" +
		strings.Join(quoted, ", ") + ") " +
		"are automatically imported, allowing you to reference their members directly " +
		"using the package prefix (e.g. `color.red`, `ease.quadOut(t)`). A nested " +
		"package is referenced by the last segment of its path alone — write " +
		"`beam.bcc(10)`, not `lattice.beam.bcc(10)`.\n\n")
	sb.WriteString("To import a custom or product-specific library, put an `import` " +
		"line at the top of the file. Two forms:\n\n")
	sb.WriteString("```gsl\n" +
		"import \"path/to/pkg\";      // qualified: call as `pkg.func(1)` (prefix = last path segment)\n" +
		"import alias \"path/to/pkg\";  // aliased:   call as `alias.func(1)`\n" +
		"```\n\n")
	sb.WriteString("---\n\n")

	for _, lib := range gsllibs.All() {
		if lib.Domain != "" && !containsStr(includeDomains, lib.Domain) {
			continue
		}
		entries := extractLibraryEntries(strings.Join(lib.Sources, "\n"))
		if len(entries.constants) == 0 && len(entries.preamble) == 0 && len(entries.sections) == 0 {
			continue
		}
		fmt.Fprintf(&sb, "### `%s`\n\n", lib.ImportPath)
		writeLibraryEntries(&sb, entries)
	}

	return sb.String()
}

// SourceLibraryDescriptiveText renders the same per-library documentation
// section that LibrariesDescriptiveText emits, but for a library source
// supplied by the caller — e.g. product libraries like dm/gate that live in
// a documents tree rather than the embedded set. Returns "" when the source
// has no documented API.
func SourceLibraryDescriptiveText(importPath, source string) string {
	entries := extractLibraryEntries(source)
	if len(entries.constants) == 0 && len(entries.preamble) == 0 && len(entries.sections) == 0 {
		return ""
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "### `%s`\n\n", importPath)
	writeLibraryEntries(&sb, entries)
	return sb.String()
}

// libraryEntries is what we collect for one .gsl file before rendering.
type libraryEntries struct {
	constants []declEntry      // top-level constants, in source order
	sections  []sectionEntries // function sections, in source order
	// preamble holds function-section entries that appeared before any
	// section header. Usually empty for our libraries; kept here so a
	// function declared above any divider doesn't silently disappear.
	preamble []declEntry
}

type sectionEntries struct {
	title string
	decls []declEntry
}

type declEntry struct {
	signature string // e.g. "mix(a, b, t)" or "pi"
	value     string // assignment RHS for constants; empty for functions
	doc       string // leading // block, joined with spaces; empty if none
}

var (
	// Section headers are framed by lines that look like `// ====...`.
	// The title between them is treated as the section name.
	dividerRE = regexp.MustCompile(`^//\s*=+\s*$`)
	// A top-level declaration starts at column 0 with an identifier.
	declStartRE = regexp.MustCompile(`^([a-zA-Z_][a-zA-Z0-9_]*)\s*([(=])`)
	// A declaration-level annotation sits in front of the declaration it
	// marks (`@material steel() {`). It is stripped before the line is read
	// as a declaration, so an annotated member is documented like any other
	// — without this, annotating a library silently removes it from these
	// docs.
	annotationPrefixRE = regexp.MustCompile(`^@[a-zA-Z_][a-zA-Z0-9_]*(\([^)]*\))?\s+`)
	// Strip trailing inline comments (only handles `//`, which is what
	// our libraries use today).
	inlineCommentRE = regexp.MustCompile(`\s*//.*$`)
)

func extractLibraryEntries(src string) libraryEntries {
	var out libraryEntries
	var current *sectionEntries
	var docBuf []string
	var i int

	lines := strings.Split(src, "\n")
	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		// Section header: divider, single comment line (the title), divider.
		if dividerRE.MatchString(trimmed) && i+2 < len(lines) {
			titleLine := strings.TrimSpace(lines[i+1])
			closer := strings.TrimSpace(lines[i+2])
			if isCommentLine(titleLine) && dividerRE.MatchString(closer) {
				title := commentText(titleLine)
				out.sections = append(out.sections, sectionEntries{title: title})
				current = &out.sections[len(out.sections)-1]
				docBuf = nil
				i += 3
				continue
			}
		}

		// Lone divider line that isn't part of a header — skip it.
		if dividerRE.MatchString(trimmed) {
			docBuf = nil
			i++
			continue
		}

		// Blank line resets the doc buffer (so comments far above don't
		// stick to a later declaration).
		if trimmed == "" {
			docBuf = nil
			i++
			continue
		}

		// Public doc line — append to the buffer.
		if isPublicDocLine(trimmed) {
			docBuf = append(docBuf, publicDocText(trimmed))
			i++
			continue
		}

		// Internal comment line — skip without resetting the doc buffer
		// so a `///` block above can still attach to the function below
		// even if there's an internal `//` note in between.
		if isCommentLine(trimmed) {
			i++
			continue
		}

		// Top-level declaration at column 0? An annotation in front of it is
		// not part of the declaration, so read past it.
		declLine := line
		if !strings.HasPrefix(declLine, " ") && !strings.HasPrefix(declLine, "\t") {
			declLine = annotationPrefixRE.ReplaceAllString(declLine, "")
		}
		if m := declStartRE.FindStringSubmatch(declLine); m != nil && !strings.HasPrefix(line, " ") && !strings.HasPrefix(line, "\t") {
			name := m[1]
			kind := m[2]
			doc := strings.Join(docBuf, " ")
			docBuf = nil
			switch kind {
			case "=":
				entry := declEntry{
					signature: name,
					value:     extractAssignmentValue(declLine),
					doc:       doc,
				}
				out.constants = append(out.constants, entry)
			case "(":
				// No `///` block above it → internal helper. Skip; the
				// docs only show functions a library author marked as
				// public API.
				if doc == "" {
					break
				}
				sig := annotationPrefixRE.ReplaceAllString(extractSignatureUpToBrace(lines, i), "")
				entry := declEntry{signature: sig, doc: doc}
				if current == nil {
					out.preamble = append(out.preamble, entry)
				} else {
					current.decls = append(current.decls, entry)
				}
			}
		} else {
			// Function body, indented line, or something we don't model.
			docBuf = nil
		}
		i++
	}

	return out
}

// isCommentLine returns true for any `//`-prefixed line (including `///`
// public docs). Used by section header detection where the title line
// is a plain `//` comment.
func isCommentLine(s string) bool {
	return strings.HasPrefix(s, "//")
}

// isPublicDocLine returns true only for `///`-prefixed public docs. A
// fourth slash (`////`) is treated as a divider-ish internal line so
// authors can still draw separators inside doc blocks without leaking
// them into the preamble.
func isPublicDocLine(s string) bool {
	return strings.HasPrefix(s, "///") && !strings.HasPrefix(s, "////")
}

// publicDocText strips the `///` prefix and any leading space.
func publicDocText(s string) string {
	s = strings.TrimPrefix(s, "///")
	return strings.TrimSpace(s)
}

// commentText strips the `// ` (or `//`) prefix from a comment line and
// trims surrounding space.
func commentText(s string) string {
	s = strings.TrimPrefix(s, "//")
	return strings.TrimSpace(s)
}

// extractAssignmentValue returns the right-hand side of `name = value;`
// with any trailing semicolon and inline comment stripped.
func extractAssignmentValue(line string) string {
	idx := strings.Index(line, "=")
	if idx < 0 {
		return ""
	}
	rhs := strings.TrimSpace(line[idx+1:])
	rhs = inlineCommentRE.ReplaceAllString(rhs, "")
	rhs = strings.TrimSuffix(rhs, ";")
	return strings.TrimSpace(rhs)
}

// extractSignatureUpToBrace joins the function's `name(args)` signature,
// handling the case where it wraps across lines before the `{`. Stops at
// the first `{`.
func extractSignatureUpToBrace(lines []string, start int) string {
	var sb strings.Builder
	for i := start; i < len(lines); i++ {
		line := lines[i]
		if idx := strings.Index(line, "{"); idx >= 0 {
			sb.WriteString(strings.TrimSpace(line[:idx]))
			break
		}
		if i > start {
			sb.WriteByte(' ')
		}
		sb.WriteString(strings.TrimSpace(line))
	}
	return strings.TrimSpace(sb.String())
}

func writeLibraryEntries(sb *strings.Builder, e libraryEntries) {
	if len(e.constants) > 0 {
		sb.WriteString("#### Constants\n\n")
		for _, c := range e.constants {
			if c.doc != "" {
				fmt.Fprintf(sb, "- **`%s = %s`** — %s\n", c.signature, c.value, c.doc)
			} else {
				fmt.Fprintf(sb, "- **`%s = %s`**\n", c.signature, c.value)
			}
		}
		sb.WriteString("\n")
	}
	if len(e.preamble) > 0 {
		for _, d := range e.preamble {
			if d.doc != "" {
				fmt.Fprintf(sb, "- **`%s`** — %s\n", d.signature, d.doc)
			} else {
				fmt.Fprintf(sb, "- **`%s`**\n", d.signature)
			}
		}
		sb.WriteString("\n")
	}
	for _, s := range e.sections {
		if len(s.decls) == 0 {
			continue
		}
		fmt.Fprintf(sb, "#### %s\n\n", s.title)
		for _, d := range s.decls {
			if d.doc != "" {
				fmt.Fprintf(sb, "- **`%s`** — %s\n", d.signature, d.doc)
			} else {
				fmt.Fprintf(sb, "- **`%s`**\n", d.signature)
			}
		}
		sb.WriteString("\n")
	}
}
