package fix

import (
	"strings"
)

var domainFixer = Fixer{
	Name: "domain",
	Doc:  `rewrite the stdlib "map" package, renamed "domain"`,
	Help: `The domain fixer follows the rename of the standard library's "map" package
to "domain". The old name shared its spelling with the map(fn, list) primitive.

It rewrites, outside comments and strings:
  - package-qualified references:   map.rotateXY(p, a)  ->  domain.rotateXY(p, a)
  - the import path in every form:  import "map";       ->  import "domain";
                                    import m "map";     ->  import m "domain";
                                    import . "map";     ->  import . "domain";
                                    import map "map";   ->  import domain "domain";

The map(fn, list) primitive is a bare call and is left alone. map is a reserved
word, so "map." can only be the package prefix, unless the file imports another
package under the explicit name map; such a file is left unchanged.

Before:
  return remap(sphere(1), map.rotateXY(p, 0.5));

After:
  return remap(sphere(1), domain.rotateXY(p, 0.5));
`,
	Apply: applyDomainFix,
}

func applyDomainFix(src string) (string, bool, error) {
	if !strings.Contains(src, "map") {
		return src, false, nil
	}
	var toks []gslToken
	for pos := 0; ; {
		t, ok := nextGSLToken(src, pos)
		if !ok {
			break
		}
		toks = append(toks, t)
		pos = t.end
	}

	// A file that names another package map keeps its references.
	for i := 0; i+2 < len(toks); i++ {
		if toks[i].text == "import" && toks[i+1].text == "map" && toks[i+2].kind == tokString && toks[i+2].text != `"map"` {
			return src, false, nil
		}
	}

	var edits []textEdit
	for i, t := range toks {
		switch {
		case t.kind == tokString && t.text == `"map"` && isImportPath(toks, i):
			edits = append(edits, textEdit{start: t.start, end: t.end, text: `"domain"`})
			// `import map "map"` names the package by its old prefix; the
			// references are rewritten, so the name follows them.
			if i >= 2 && toks[i-2].text == "import" && toks[i-1].text == "map" {
				edits = append(edits, textEdit{start: toks[i-1].start, end: toks[i-1].end, text: "domain"})
			}
		case t.kind == tokIdent && t.text == "map" &&
			i+2 < len(toks) && toks[i+1].kind == tokDot && toks[i+1].start == t.end &&
			toks[i+2].kind == tokIdent && toks[i+2].start == toks[i+1].end &&
			(i == 0 || toks[i-1].kind != tokDot || toks[i-1].end != t.start):
			edits = append(edits, textEdit{start: t.start, end: t.end, text: "domain"})
		}
	}
	if len(edits) == 0 {
		return src, false, nil
	}
	out := applyNonOverlappingEdits(src, edits)
	return out, out != src, nil
}

// isImportPath reports whether the string token at i is an import's path:
// `import "p"`, `import name "p"` or `import . "p"`.
func isImportPath(toks []gslToken, i int) bool {
	if i >= 1 && toks[i-1].text == "import" {
		return true
	}
	return i >= 2 && toks[i-2].text == "import" && (toks[i-1].kind == tokIdent || toks[i-1].kind == tokDot)
}
