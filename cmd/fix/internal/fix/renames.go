package fix

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/Gradient-Control-Laboratories/omega-core/program"
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

// simpleOperand is a name, a qualified name or a number: safe to subtract
// without parentheses.
var simpleOperand = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_.]*|[0-9]+(\.[0-9]*)?([eE][-+]?[0-9]+)?)$`)

// collectLatticeBccEdits rewrites the deprecated lattice.bcc(cellSize,
// beamRadius[, combine]) to (beam.bcc(cellSize[, combine]) - beamRadius), the
// same lattice. lattice is reached by its default prefix or by the name the
// file imports it under.
func collectLatticeBccEdits(src string, file *program.PackageFile) []textEdit {
	prefixes := map[string]bool{"lattice": true}
	for _, imp := range file.Imports {
		if imp == nil {
			continue
		}
		if imp.Path == "lattice" && imp.Prefix != "" && imp.Prefix != "." {
			prefixes[imp.Prefix] = true
		} else if imp.Prefix == "lattice" && imp.Path != "lattice" {
			delete(prefixes, "lattice")
		}
	}
	text := func(op *program.Operand) (string, bool) {
		if op == nil || op.Start == nil || op.End == nil {
			return "", false
		}
		s, e := int(op.Start.Offset), int(op.End.Offset)
		if s < 0 || e > len(src) || s > e {
			return "", false
		}
		return src[s:e], true
	}
	var edits []textEdit
	visitCallsInFile(file, func(op *program.Operand, fc *program.FnCall) {
		if fc.Reference == nil || fc.Reference.Name != "bcc" || !prefixes[fc.Reference.Prefix] {
			return
		}
		n := len(fc.Arguments)
		if n != 2 && n != 3 {
			return
		}
		call, ok := text(op)
		if !ok {
			return
		}
		// An expression argument's span is not its whole text, so split the
		// call's own text at its top-level commas.
		parts := callArguments(call)
		if len(parts) != n {
			return
		}
		radius := parts[1]
		if !simpleOperand.MatchString(radius) {
			radius = "(" + radius + ")"
		}
		args := parts[0]
		if n == 3 {
			args += ", " + parts[2]
		}
		edits = append(edits, textEdit{
			start: int(op.Start.Offset), end: int(op.End.Offset),
			text: "(beam.bcc(" + args + ") - " + radius + ")",
		})
	})
	return edits
}

// callArguments splits a call's text, name(a, b, ...), into its trimmed
// arguments at the top-level commas, or nil if it is not a call.
func callArguments(call string) []string {
	open := strings.IndexByte(call, '(')
	if open < 0 || !strings.HasSuffix(call, ")") {
		return nil
	}
	inner := call[open+1 : len(call)-1]
	var parts []string
	depth, start := 0, 0
	for i := 0; i < len(inner); i++ {
		switch inner[i] {
		case '(', '[', '{':
			depth++
		case ')', ']', '}':
			depth--
		case '"':
			for i++; i < len(inner) && inner[i] != '"'; i++ {
				if inner[i] == '\\' {
					i++
				}
			}
		case ',':
			if depth == 0 {
				parts = append(parts, strings.TrimSpace(inner[start:i]))
				start = i + 1
			}
		}
	}
	return append(parts, strings.TrimSpace(inner[start:]))
}
