// Package suggest turns "undefined identifier" and "name not found in package"
// into errors that name the closest names in scope, each with its signature
// and a one-line doc, so the error is the documentation the reader needs.
package suggest

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Gradient-Control-Laboratories/omega-core/internal/scope"
	"github.com/Gradient-Control-Laboratories/omega-core/program"
	"github.com/Gradient-Control-Laboratories/omega-core/program/preamble"
	gsllibs "github.com/Gradient-Control-Laboratories/omega-core/stdlib"
)

// maxSuggestions is how many names an error offers.
const maxSuggestions = 3

// annotationCalls are what @point, @transform, @color, @direction and @spline
// compile to. A document writes the annotation, so they are never suggested.
var annotationCalls = map[string]bool{
	"uiPoint": true, "uiTransform": true, "uiColor": true, "uiDirection": true, "uiSpline": true,
}

// Undefined is the error for a name that resolves to nothing in s: "undefined
// identifier: name", followed by up to three close names visible from s.
func Undefined(s scope.Scope, name string) error {
	msg := "undefined identifier: " + name
	cands := visible(s)
	return fmt.Errorf("%s%s", msg, list(closest(name, keys(cands)), func(c string) string {
		return describe(c, "", cands[c])
	}))
}

// NotInPackage is the error for prefix.name when the package has no member
// name: up to three close members of that package.
func NotInPackage(pkg *scope.PkgRef, name string) error {
	msg := fmt.Sprintf("name %q not found in package %q", name, pkg.Name())
	cands := map[string]entry{}
	if ps := pkg.PkgScopePtr(); ps != nil {
		path := ""
		if ps.Package != nil {
			path = ps.Package.Path
		}
		for _, f := range ps.Files {
			addFile(cands, f.File, path)
		}
	}
	return fmt.Errorf("%s%s", msg, list(closest(name, keys(cands)), func(c string) string {
		return describe(c, pkg.Name(), cands[c])
	}))
}

// entry is what is known about a candidate name.
type entry struct {
	fnType  *program.FnType // a function declared in GSL, for its signatures
	pkgPath string          // the library it comes from, for its doc
	builtin bool            // a built-in or primitive, documented by preamble
	pkg     string          // a package prefix: its import path
}

// visible collects every name a reference in s could resolve to: the
// enclosing functions' parameters, locals and nested functions, the file's
// and package's declarations, dot-imported library members, built-ins and
// primitives, and the standard package prefixes.
func visible(s scope.Scope) map[string]entry {
	cands := map[string]entry{}
	seenFiles := map[*program.PackageFile]bool{}
	for cur := s; cur != nil; cur = cur.ContainingScope() {
		switch v := cur.(type) {
		case *scope.FnScope:
			if fn := v.Function; fn != nil {
				for _, p := range fn.Parameters {
					addName(cands, p.Name, entry{})
				}
				for _, a := range fn.Assignments {
					for _, vr := range a.Variables {
						addName(cands, vr.Name, entry{})
					}
				}
				for _, ft := range fn.FnTypes {
					addName(cands, ft.Name, entry{fnType: ft})
				}
			}
		case *scope.PkgFileScope:
			if !seenFiles[v.File] {
				seenFiles[v.File] = true
				addFile(cands, v.File, "")
			}
			for _, imp := range v.DotImports {
				if ps := imp.PkgScopePtr(); ps != nil && ps.Package != nil {
					for _, f := range ps.Files {
						addFile(cands, f.File, ps.Package.Path)
					}
				}
			}
			for prefix, imp := range v.Imports {
				path := prefix
				if ps := imp.PkgScopePtr(); ps != nil && ps.Package != nil {
					path = ps.Package.Path
				}
				addName(cands, prefix, entry{pkg: path})
			}
		case *scope.PkgScope:
			for _, f := range v.Files {
				if !seenFiles[f.File] {
					seenFiles[f.File] = true
					addFile(cands, f.File, "")
				}
			}
		}
	}
	for _, b := range program.Builtins() {
		if annotationCalls[b] {
			continue
		}
		addName(cands, b, entry{builtin: true})
	}
	for op, name := range program.Operation_name {
		if preamble.BuiltinDoc(program.Operation(op).Name()) != "" {
			addName(cands, strings.ToLower(name), entry{builtin: true})
		}
	}
	for _, prefix := range gsllibs.Prefixes() {
		path, _ := gsllibs.PathForPrefix(prefix)
		addName(cands, prefix, entry{pkg: path})
	}
	return cands
}

// addFile adds a file's top-level variables and functions; pkgPath is the
// library they belong to, "" for the document's own.
func addFile(cands map[string]entry, file *program.PackageFile, pkgPath string) {
	if file == nil {
		return
	}
	for _, a := range file.Assignments {
		for _, vr := range a.Variables {
			addName(cands, vr.Name, entry{pkgPath: pkgPath})
		}
	}
	for _, ft := range file.FnTypes {
		if ft.IsAnonFn() {
			continue
		}
		addName(cands, ft.Name, entry{fnType: ft, pkgPath: pkgPath})
	}
}

// addName records a candidate; the innermost binding of a name wins.
func addName(cands map[string]entry, name string, e entry) {
	if name == "" || strings.HasPrefix(name, "gcl_") || strings.HasPrefix(name, "_") {
		return
	}
	if _, ok := cands[name]; !ok {
		cands[name] = e
	}
}

func keys(m map[string]entry) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// closest returns up to maxSuggestions candidates near name: the same name in
// another case, then by edit distance within a third of the name's length
// (at least 1, at most 3), shorter distances first. A name under three
// letters is only matched in another case: one edit away from it is almost
// every other short name.
func closest(name string, cands []string) []string {
	limit := len(name) / 3
	if limit < 1 {
		limit = 1
	}
	if len(name) < 3 {
		limit = 0
	}
	if limit > 3 {
		limit = 3
	}
	lower := strings.ToLower(name)
	type scored struct {
		name string
		d    int
	}
	var hits []scored
	for _, c := range cands {
		if c == name {
			continue
		}
		d := distance(lower, strings.ToLower(c))
		if strings.EqualFold(c, name) {
			d = 0
		}
		if d <= limit {
			hits = append(hits, scored{c, d})
		}
	}
	sort.Slice(hits, func(i, j int) bool {
		if hits[i].d != hits[j].d {
			return hits[i].d < hits[j].d
		}
		return hits[i].name < hits[j].name
	})
	var out []string
	for i := 0; i < len(hits) && i < maxSuggestions; i++ {
		out = append(out, hits[i].name)
	}
	return out
}

// distance is the optimal-string-alignment edit distance: insertions,
// deletions, substitutions and swaps of neighbours each cost 1.
func distance(a, b string) int {
	ra, rb := []rune(a), []rune(b)
	d := make([][]int, len(ra)+1)
	for i := range d {
		d[i] = make([]int, len(rb)+1)
		d[i][0] = i
	}
	for j := range d[0] {
		d[0][j] = j
	}
	for i := 1; i <= len(ra); i++ {
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			d[i][j] = min(d[i-1][j]+1, d[i][j-1]+1, d[i-1][j-1]+cost)
			if i > 1 && j > 1 && ra[i-1] == rb[j-2] && ra[i-2] == rb[j-1] {
				d[i][j] = min(d[i][j], d[i-2][j-2]+1)
			}
		}
	}
	return d[len(ra)][len(rb)]
}

// list formats the suggestions as indented lines after "; did you mean:", or
// "" when there are none.
func list(names []string, line func(string) string) string {
	if len(names) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("; did you mean:")
	for _, n := range names {
		sb.WriteString("\n  " + line(n))
	}
	return sb.String()
}

// describe renders one suggestion: its signatures, then a dash and its doc
// when there is one. prefix qualifies a package member ("transform").
func describe(name, prefix string, e entry) string {
	qualified := name
	if prefix != "" {
		qualified = prefix + "." + name
	}
	head := qualified
	doc := ""
	switch {
	case e.pkg != "":
		head = qualified + " (package)"
		if e.pkg != name {
			head = qualified + " (package " + e.pkg + ")"
		}
		doc = preamble.PackageSummary(e.pkg)
		doc = preamble.FirstSentence(doc)
	case e.builtin:
		doc = preamble.BuiltinDoc(name)
	case e.fnType != nil:
		head = strings.Join(signatures(qualified, e.fnType), ", ")
		if e.pkgPath != "" {
			doc = preamble.MemberDoc(e.pkgPath, name)
		}
	default:
		if e.pkgPath != "" {
			doc = preamble.MemberDoc(e.pkgPath, name)
		}
	}
	if doc == "" {
		return head
	}
	return head + " — " + doc
}

// signatures lists a function's overloads, fewest parameters first.
func signatures(name string, ft *program.FnType) []string {
	groups := append([]*program.FnGroup(nil), ft.Groups...)
	sort.Slice(groups, func(i, j int) bool { return groups[i].ParamCount < groups[j].ParamCount })
	var out []string
	for _, g := range groups {
		if len(g.Functions) == 0 {
			continue
		}
		var params []string
		for _, p := range g.Functions[0].Parameters {
			params = append(params, p.Name)
		}
		out = append(out, fmt.Sprintf("%s(%s)", name, strings.Join(params, ", ")))
	}
	if len(out) == 0 {
		return []string{name}
	}
	return out
}
