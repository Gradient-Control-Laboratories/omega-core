package preamble

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Gradient-Control-Laboratories/omega-core/program"
	gsllibs "github.com/Gradient-Control-Laboratories/omega-core/stdlib"
)

// Lookup returns the documentation a reader asks for by name, finest first
// match: a library by import path or prefix ("material/bulk/wood", "wood"),
// a specialist built-in group ("builtin/mesh"), a primitive or built-in
// ("length"), a library member qualified by its prefix ("transform.rotate"),
// or a library member by its bare name ("unionRound"), every library that
// declares one.
func Lookup(name string) (string, error) {
	if doc, err := PackageDescriptiveText(name); err == nil {
		return doc, nil
	}
	if doc, ok := builtinGroupText(name); ok {
		return doc, nil
	}
	if doc, ok := builtinText(name); ok {
		return doc, nil
	}
	if prefix, member, ok := strings.Cut(name, "."); ok {
		path := prefix
		if p, found := gsllibs.PathForPrefix(prefix); found {
			path = p
		}
		for _, lib := range gsllibs.All() {
			if lib.ImportPath == path {
				if doc, found := memberText(lib, member); found {
					return doc, nil
				}
			}
		}
		return "", fmt.Errorf("no member %q in package %q", member, prefix)
	}
	var docs []string
	for _, lib := range gsllibs.All() {
		if doc, found := memberText(lib, name); found {
			docs = append(docs, doc)
		}
	}
	if len(docs) == 0 {
		return "", fmt.Errorf("no library, built-in group or member named %q: `gsl skill --grep %s` searches names and docs", name, name)
	}
	return strings.Join(docs, "\n"), nil
}

// builtinText documents one primitive or built-in, and names its group when
// it has one.
func builtinText(name string) (string, bool) {
	desc, ok := builtinDescriptions[name]
	if !ok {
		for op, d := range primitiveDescriptions {
			if d != skipDescription && program.Operation(op).Name() == name {
				desc, ok = d, true
				break
			}
		}
	}
	if !ok {
		return "", false
	}
	where := "a primitive or built-in"
	if g, grouped := builtinGroupOf[name]; grouped {
		where = "a built-in, in " + g
	}
	return fmt.Sprintf("### `%s` (%s)\n\n* **`%s`**: %s\n", name, where, name, desc), true
}

// memberText documents every form of one member of lib, under the library's
// heading, or reports that lib has no documented member of that name.
func memberText(lib gsllibs.PkgInfo, name string) (string, bool) {
	e := extractLibraryEntries(strings.Join(lib.Sources, "\n"))
	var found libraryEntries
	for _, c := range e.constants {
		if c.signature == name {
			found.constants = append(found.constants, c)
		}
	}
	collect := func(decls []declEntry) {
		for _, d := range decls {
			if declName(d.signature) == name {
				found.preamble = append(found.preamble, d)
			}
		}
	}
	collect(e.preamble)
	for _, s := range e.sections {
		collect(s.decls)
	}
	if len(found.constants) == 0 && len(found.preamble) == 0 {
		return "", false
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "### `%s`\n\n", lib.ImportPath)
	writeLibraryEntries(&sb, found)
	return sb.String(), true
}

// Grep returns every primitive, built-in and library member whose name or doc
// contains query, ignoring case: one line each, grouped by library, the forms
// and the first sentence of the doc. "" when nothing matches.
func Grep(query string) string {
	q := strings.ToLower(query)
	matches := func(name, doc string) bool {
		return strings.Contains(strings.ToLower(name), q) || strings.Contains(strings.ToLower(doc), q)
	}
	var sb strings.Builder

	var builtins []string
	for name, desc := range builtinDescriptions {
		if !annotationBuiltins[name] && matches(name, desc) {
			builtins = append(builtins, name)
		}
	}
	for op, desc := range primitiveDescriptions {
		if name := program.Operation(op).Name(); desc != skipDescription && matches(name, desc) {
			builtins = append(builtins, name)
		}
	}
	if len(builtins) > 0 {
		sort.Strings(builtins)
		sb.WriteString("### primitives and built-ins\n\n")
		for _, name := range builtins {
			desc := builtinDescriptions[name]
			if desc == "" {
				for op, d := range primitiveDescriptions {
					if program.Operation(op).Name() == name {
						desc = d
					}
				}
			}
			fmt.Fprintf(&sb, "- **`%s`** — %s\n", name, firstSentence(desc))
		}
		sb.WriteString("\n")
	}

	for _, lib := range gsllibs.All() {
		e := extractLibraryEntries(strings.Join(lib.Sources, "\n"))
		var hit libraryEntries
		for _, c := range e.constants {
			if matches(c.signature, c.doc) {
				hit.constants = append(hit.constants, c)
			}
		}
		// A name matches as a whole: every form of it is listed when any
		// form's name or doc matches.
		var all []declEntry
		all = append(all, e.preamble...)
		for _, s := range e.sections {
			all = append(all, s.decls...)
		}
		order, groups := groupByName(all)
		for _, name := range order {
			for _, d := range groups[name] {
				if matches(name, d.doc) {
					hit.preamble = append(hit.preamble, groups[name]...)
					break
				}
			}
		}
		if len(hit.constants) == 0 && len(hit.preamble) == 0 {
			continue
		}
		fmt.Fprintf(&sb, "### `%s`\n\n", lib.ImportPath)
		sb.WriteString(compactEntriesText(hit))
	}
	return sb.String()
}

// compactEntriesText renders entries in brief, as compactLibraryText does a
// whole library.
func compactEntriesText(e libraryEntries) string {
	var sb strings.Builder
	for _, c := range e.constants {
		fmt.Fprintf(&sb, "- `%s = %s`", c.signature, c.value)
		if c.doc != "" {
			sb.WriteString(" — " + firstSentence(c.doc))
		}
		sb.WriteString("\n")
	}
	order, groups := groupByName(e.preamble)
	for _, name := range order {
		var forms []string
		doc := ""
		for _, d := range groups[name] {
			forms = append(forms, "`"+strings.TrimPrefix(d.signature, name)+"`")
			if doc == "" && d.doc != "" {
				doc = firstSentence(d.doc)
			}
		}
		fmt.Fprintf(&sb, "- **`%s`** %s", name, strings.Join(forms, " · "))
		if doc != "" {
			sb.WriteString(" — " + doc)
		}
		sb.WriteString("\n")
	}
	sb.WriteString("\n")
	return sb.String()
}
