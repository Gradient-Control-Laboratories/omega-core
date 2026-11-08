// Command materialswatch builds the material browser's catalog: it reads every
// @material and @finish annotation out of the embedded GSL libraries, renders a
// sphere of each one with the CPU lit renderer, and writes the swatch sheet and
// the manifest the browser panel reads (examples/viewer/materials/).
//
// The annotations are the catalog. A library grows a surface by annotating it —
// see internal/assembler/annotation, where @material marks a function returning a
// material record and the albedo that goes with it, and @finish marks the
// material library's own presets, which carry no albedo on purpose. Nothing
// here holds a list of names. The @surface presets, which take a surface and
// return it modified, are not swatches: they go into the manifest's operators
// (operators.go) for the panel to wrap a body's surface with.
package main

import (
	"path"
	"regexp"
	"strings"

	gsllibs "gradientcontrol.com/omega/core/stdlib"
)

// Entry is one browsable surface. Kind is the annotation that marked it, so a
// consumer can tell the two apart without parsing anything: a "material" call
// returns its own colour and a "finish" has to be drawn against a stand-in.
type Entry struct {
	Kind    string `json:"kind"`    // "material" or "finish"
	Library string `json:"library"` // the library's import path, e.g. "material/bulk/stone"
	Package string `json:"package"` // the name a document calls the library by, e.g. "metal" or "stone"
	Symbol  string `json:"symbol"`  // the annotated name, e.g. "brass"
	Section string `json:"section"` // the library's own section heading, or ""
	Doc     string `json:"doc"`     // the /// block, as one paragraph
	Index   int    `json:"index"`   // this entry's cell in the swatch sheet
}

// Call is the GSL expression that yields the surface: a material is called, a
// finish is a constant and is named.
func (e Entry) Call() string {
	if e.Kind == "finish" {
		return e.Package + "." + e.Symbol
	}
	return e.Package + "." + e.Symbol + "()"
}

var (
	// An annotated declaration: "@material steel() {" or "@finish matte = …".
	// Only column zero, as the libraries' doc convention has it — an indented
	// line is inside a body.
	annotatedRE = regexp.MustCompile(`^@(material|finish)\s+([A-Za-z_][A-Za-z0-9_]*)`)
	// A section frame's rule, "// =====…". The title is the line between two
	// of them (libraries.go reads the same frames for the docs).
	sectionRuleRE  = regexp.MustCompile(`^//\s*={10,}\s*$`)
	sectionTitleRE = regexp.MustCompile(`^//\s*(\S.*?)\s*$`)
	// Every @material's doc ends by saying what it returns, which is worth
	// stating in the library and is noise in a panel where all 134 say it.
	returnsBoilerplateRE = regexp.MustCompile(`\s*Returns material and colour\.\s*$`)
)

// ReadCatalog scans every embedded library for annotated surfaces, in the order
// the libraries declare them: an entry's Index is its place in that order, and
// so its cell in the sheet, which keeps the sheet stable under a library edit
// that only appends.
func ReadCatalog() []Entry {
	var out []Entry
	for _, pkg := range gsllibs.All() {
		for _, src := range pkg.Sources {
			found := scanSource(pkg.ImportPath, src)
			for i := range found {
				found[i].Library = pkg.ImportPath
			}
			out = append(out, found...)
		}
	}
	for i := range out {
		out[i].Index = i
	}
	return out
}

// scanSource reads one library file. It tracks the section frames and the ///
// doc blocks exactly as the preamble's library docs do, so a surface's blurb in
// the panel is the one its author wrote for the docs.
func scanSource(importPath, src string) []Entry {
	var out []Entry
	var doc []string
	section := ""
	lines := strings.Split(src, "\n")
	for i, line := range lines {
		switch {
		case sectionRuleRE.MatchString(line):
			// A rule, a title, a rule. Read the title from between them and
			// leave i alone: the closing rule lands here again and finds no
			// title after it, which is harmless.
			if i+2 < len(lines) && sectionRuleRE.MatchString(lines[i+2]) {
				if m := sectionTitleRE.FindStringSubmatch(lines[i+1]); m != nil {
					section = m[1]
				}
			}
			doc = nil
		case strings.HasPrefix(line, "///"):
			doc = append(doc, strings.TrimSpace(strings.TrimPrefix(line, "///")))
		case annotatedRE.MatchString(line):
			m := annotatedRE.FindStringSubmatch(line)
			out = append(out, Entry{
				Kind: m[1],
				// A document reaches material/bulk/stone as stone.granite(…), by
				// the import path's last segment, so that is what the call is
				// written with.
				Package: path.Base(importPath),
				Symbol:  m[2],
				Section: section,
				Doc:     joinDoc(doc),
			})
			doc = nil
		case strings.TrimSpace(line) == "":
			// A blank line ends a doc block.
			doc = nil
		case strings.HasPrefix(line, "//"):
			// An internal note inside a doc block: ignored, as in the docs.
		default:
			doc = nil
		}
	}
	return out
}

// joinDoc folds a /// block into one paragraph and drops the sentence every
// @material ends with.
func joinDoc(lines []string) string {
	doc := strings.Join(lines, " ")
	doc = returnsBoilerplateRE.ReplaceAllString(doc, "")
	return strings.TrimSpace(doc)
}
