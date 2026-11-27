package preamble

import (
	"fmt"
	"strings"

	gsllibs "github.com/Gradient-Control-Laboratories/omega-core/stdlib"
)

// The full library docs run to hundreds of kilobytes, more than a model should
// read before writing a line. The skill index carries what nearly every
// program needs and names the rest; a reader loads the rest by name (Lookup)
// once it knows which it needs.

// GSLSkillIndex returns the GSL author skill a model reads first: the language
// description, the primitives and core built-ins with the specialist groups
// named, the builtin package in brief, how packages are imported and
// referenced with one line per package, the rule for reserved words, and the
// common pitfalls. The domains argument is ignored and will be removed: no
// library sets a domain.
func GSLSkillIndex(domains ...string) string {
	return GSLDescription +
		"\n\n" + coreBuiltinDescriptiveText() +
		"\n\n" + builtinBriefText() +
		"\n\n" + PackageIndexText() +
		"\n\n" + reservedWordsRuleText() +
		"\n\n" + CommonPitfalls
}

// PackageIndexText lists every embedded library on one line: its import path,
// the prefix its members are referenced by, how many documented members it
// has, and its summary (see PackageSummary). It opens with the same
// explanation of imports and prefixes as LibrariesDescriptiveText. The domains
// argument is ignored and will be removed: no library sets a domain.
func PackageIndexText(domains ...string) string {
	var sb strings.Builder
	sb.WriteString("\n---\n\n## **Library Packages**\n\n")
	sb.WriteString(librariesLead)
	writeLibrariesIntro(&sb)
	sb.WriteString("### Package Index\n\n")
	sb.WriteString("Each line is a package, the prefix its members are referenced by, " +
		"how many documented members it has, and what it holds. A package's full " +
		"documentation lists every member with its signature; read it before " +
		"calling the package.\n\n")
	for _, lib := range gsllibs.All() {
		prefix := "no prefix"
		if lib.ImportPath != "builtin" {
			prefix = "`" + packagePrefix(lib.ImportPath) + ".`"
		}
		fmt.Fprintf(&sb, "- **`%s`** (%s, %d members)", lib.ImportPath, prefix, memberCount(lib))
		if summary := packageSummary(lib); summary != "" {
			sb.WriteString(": " + summary)
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// PackageDescriptiveText returns the full documentation of one embedded
// library, named by its import path ("material/bulk/wood") or its prefix
// ("wood"): its summary, then the section LibrariesDescriptiveText writes for
// it.
func PackageDescriptiveText(name string) (string, error) {
	path := name
	if p, ok := gsllibs.PathForPrefix(name); ok {
		path = p
	}
	for _, lib := range gsllibs.All() {
		if lib.ImportPath != path {
			continue
		}
		var sb strings.Builder
		fmt.Fprintf(&sb, "### `%s`\n\n", lib.ImportPath)
		if summary := packageSummary(lib); summary != "" {
			sb.WriteString(summary + "\n\n")
		}
		writeLibraryEntries(&sb, extractLibraryEntries(strings.Join(lib.Sources, "\n")))
		return sb.String(), nil
	}
	return "", fmt.Errorf("no library %q: name one by import path or prefix (%s)", name, strings.Join(PackageNames(), ", "))
}

// PackageNames returns every embedded library's import path, in documentation
// order.
func PackageNames() []string {
	libs := gsllibs.All()
	names := make([]string, len(libs))
	for i, lib := range libs {
		names[i] = lib.ImportPath
	}
	return names
}

// PackageSummary returns a library's summary: the `///` block at the very top
// of its first source, ended by a blank line. Nothing but blank lines may come
// before it. A blank line detaches a doc block from the declaration below, so
// the summary never documents a member. "" when the library has none.
func PackageSummary(importPath string) string {
	for _, lib := range gsllibs.All() {
		if lib.ImportPath == importPath {
			return packageSummary(lib)
		}
	}
	return ""
}

func packageSummary(lib gsllibs.PkgInfo) string {
	if len(lib.Sources) == 0 {
		return ""
	}
	var doc []string
	for _, line := range strings.Split(lib.Sources[0], "\n") {
		trimmed := strings.TrimSpace(line)
		switch {
		case isPublicDocLine(trimmed):
			doc = append(doc, publicDocText(trimmed))
		case trimmed == "" && len(doc) == 0:
			continue
		case trimmed == "":
			return strings.Join(doc, " ")
		default:
			// A doc block that runs into code documents that code, not the
			// package.
			return ""
		}
	}
	return ""
}

// memberCount is how many documented members a library has: the entries its
// full documentation lists.
func memberCount(lib gsllibs.PkgInfo) int {
	e := extractLibraryEntries(strings.Join(lib.Sources, "\n"))
	n := len(e.constants) + len(e.preamble)
	for _, s := range e.sections {
		n += len(s.decls)
	}
	return n
}

// packagePrefix is the prefix a library's members are referenced by: the last
// segment of its import path.
func packagePrefix(importPath string) string {
	if i := strings.LastIndex(importPath, "/"); i >= 0 {
		return importPath[i+1:]
	}
	return importPath
}
