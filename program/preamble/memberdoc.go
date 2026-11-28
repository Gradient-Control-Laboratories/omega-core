package preamble

import (
	"strings"
	"sync"

	"github.com/Gradient-Control-Laboratories/omega-core/program"
	gsllibs "github.com/Gradient-Control-Laboratories/omega-core/stdlib"
)

// The compiler quotes these when it suggests a name: an error that says what
// the suggested function does is the doc a model or a person needs at that
// moment, without a second lookup.

var (
	memberDocsOnce sync.Once
	memberDocs     map[string]map[string]string // import path -> member -> first sentence
)

// MemberDoc returns the first sentence of the `///` doc of member name in the
// embedded library importPath, or "" when it has none. For an overloaded
// function it is the first documented overload's.
func MemberDoc(importPath, name string) string {
	memberDocsOnce.Do(func() {
		memberDocs = make(map[string]map[string]string)
		for _, lib := range gsllibs.All() {
			docs := make(map[string]string)
			e := extractLibraryEntries(strings.Join(lib.Sources, "\n"))
			add := func(d declEntry) {
				name := d.signature
				if i := strings.Index(name, "("); i >= 0 {
					name = name[:i]
				}
				if _, seen := docs[name]; !seen && d.doc != "" {
					docs[name] = FirstSentence(d.doc)
				}
			}
			for _, c := range e.constants {
				add(c)
			}
			for _, d := range e.preamble {
				add(d)
			}
			for _, s := range e.sections {
				for _, d := range s.decls {
					add(d)
				}
			}
			memberDocs[lib.ImportPath] = docs
		}
	})
	return memberDocs[importPath][name]
}

// BuiltinDoc returns the first sentence of a built-in's or primitive's
// description, or "" when name is neither or is not described.
func BuiltinDoc(name string) string {
	if d, ok := builtinDescriptions[name]; ok {
		return FirstSentence(d)
	}
	for op, d := range primitiveDescriptions {
		if d != skipDescription && program.Operation(op).Name() == name {
			return FirstSentence(d)
		}
	}
	return ""
}

// FirstSentence returns text up to and including its first full stop that is
// followed by a space, or all of it when there is none.
func FirstSentence(text string) string {
	text = strings.TrimSpace(text)
	if i := strings.Index(text, ". "); i >= 0 {
		return text[:i+1]
	}
	return text
}
