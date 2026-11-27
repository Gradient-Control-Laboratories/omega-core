package fix

import (
	"strconv"
	"strings"
)

// renamedMembers maps a standard library import path to the members that
// were renamed in it, old name to new.
var renamedMembers = map[string]map[string]string{
	"lattice/honeycomb": {
		"squareHoneycomb":           "rectangular",
		"squareDiagonalHoneycomb":   "rectangularDiagonal",
		"squareHoneycombReinforced": "rectangularReinforced",
		"hexagonalHoneycomb":        "hexagonal",
		"triangularHoneycomb":       "triangular",
		"octetHoneycomb":            "octet",
	},
}

// renameMembers rewrites references to renamed standard library members,
// outside comments and strings. A member is reached through its package's
// prefix: the last segment of its path unless the file imports it under
// another name, or bare when the file dot-imports it.
func renameMembers(src string) string {
	var toks []gslToken
	for pos := 0; ; {
		t, ok := nextGSLToken(src, pos)
		if !ok {
			break
		}
		toks = append(toks, t)
		pos = t.end
	}

	// prefixes[path] holds the names the file reaches path by; "" is a
	// dot-import. A prefix the file binds to some other path is not one.
	prefixes := map[string]map[string]bool{}
	bound := map[string]string{}
	for path := range renamedMembers {
		seg := path[strings.LastIndex(path, "/")+1:]
		prefixes[path] = map[string]bool{seg: true}
	}
	for i, t := range toks {
		if t.kind != tokString || !isImportPath(toks, i) {
			continue
		}
		path, err := strconv.Unquote(t.text)
		if err != nil {
			continue
		}
		name := path[strings.LastIndex(path, "/")+1:]
		if toks[i-1].text != "import" {
			name = toks[i-1].text
			if name == "." {
				name = ""
			}
		}
		bound[name] = path
		if p, ok := prefixes[path]; ok {
			p[name] = true
		}
	}

	var edits []textEdit
	for path, members := range renamedMembers {
		for prefix := range prefixes[path] {
			if b, ok := bound[prefix]; ok && b != path {
				continue
			}
			for i, t := range toks {
				newName, ok := members[t.text]
				if t.kind != tokIdent || !ok {
					continue
				}
				afterDot := i >= 1 && toks[i-1].kind == tokDot && toks[i-1].end == t.start
				switch {
				case prefix == "" && !afterDot:
					edits = append(edits, textEdit{start: t.start, end: t.end, text: newName})
				case prefix != "" && afterDot && i >= 2 && toks[i-2].kind == tokIdent &&
					toks[i-2].text == prefix && toks[i-2].end == toks[i-1].start:
					edits = append(edits, textEdit{start: t.start, end: t.end, text: newName})
				}
			}
		}
	}
	if len(edits) == 0 {
		return src
	}
	return applyNonOverlappingEdits(src, edits)
}
