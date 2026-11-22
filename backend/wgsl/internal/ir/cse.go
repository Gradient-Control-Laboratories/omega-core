package ir

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

// Common subexpressions.
//
// The assembler inlines every call it can, and a function called twice with
// arguments that differ in a constant is inlined twice: a shaft's fold asks
// each segment four times, capped and extended, and the section it carries (a
// gear profile, a spline, a helix's turn) does not depend on the extension,
// so the four copies hold the same expression. Fidget's context deduplicates;
// nothing here did. This pass merges them in a structured function's
// statements, whose expressions are text the codegen has already written.
//
// It parses each expression, numbers its values (two subtrees with the same
// operation on the same values have the same number, and a let's name has its
// expression's), and then applies two rules:
//
//   - A let whose value an earlier let in scope already holds is dropped and
//     read as that let; any subexpression an in-scope let holds is read as
//     the let, where its name is the shorter text.
//   - A subexpression whose value repeats within one block is named, by a let
//     before its first use, when naming it shortens the function, the largest
//     first; where a later let of the same block already holds it, that let
//     moves up to the first use instead.
//
// What it never does: hoist out of a guard or a loop body (a guard keeps
// conditional work conditional, the stand-in's detail for one, and a loop's
// body runs once per element), merge what reads a name the function or the
// module writes (a var, a loop index, an assignment's, an opaque statement's
// or a guard header's target, a private global), or name a subexpression from
// the right of && or ||, which runs only when the left allows it. Opaque
// statements and guard headers are left as they are, and a let they read is
// never dropped. A replacement only ever puts a name in place of a span of
// text, so nothing else is reprinted.

// EliminateCommonSubexpressions merges the repeated expressions of a
// structured function's body, as the comment above describes. A raw function
// is left as it is, and so is a function with any expression the parser does
// not read.
func EliminateCommonSubexpressions(fn *FuncDecl) {
	eliminateCommonSubexpressions(fn, nil)
}

// EliminateCommonSubexpressions runs the pass on every structured function of
// the module, its private globals counted as written everywhere: a function
// it calls may write one.
func (m *Module) EliminateCommonSubexpressions() {
	globals := make(map[string]bool)
	scan := func(src string) {
		for _, match := range privateVarRE.FindAllStringSubmatch(src, -1) {
			globals[match[1]] = true
		}
	}
	for _, s := range m.Predefined {
		scan(s)
	}
	for _, s := range m.RuntimeSnippets {
		scan(s)
	}
	for _, fn := range append(append([]*FuncDecl{}, m.Imports...), m.Functions...) {
		scan(fn.RawSource)
		scan(fn.TrailingSource)
	}
	for _, fn := range m.Imports {
		eliminateCommonSubexpressions(fn, globals)
	}
	for _, fn := range m.Functions {
		eliminateCommonSubexpressions(fn, globals)
	}
}

var (
	privateVarRE = regexp.MustCompile(`var\s*<\s*private\s*>\s*([A-Za-z_][A-Za-z0-9_]*)`)
	identRE      = regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_]*`)
	targetRE     = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)\s*(?:\.[A-Za-z_][A-Za-z0-9_]*|\[[^\]]*\])*\s*=(?:[^=]|$)`)
)

func eliminateCommonSubexpressions(fn *FuncDecl, globals map[string]bool) {
	if fn == nil || fn.RawSource != "" || fn.Body == nil {
		return
	}
	c := &cseState{
		mutable:   make(map[string]bool),
		protected: make(map[string]bool),
		taken:     make(map[string]bool),
		alias:     make(map[string]string),
		nameVN:    make(map[string]int),
		intern:    make(map[string]int),
		parsed:    make(map[string]*exprNode),
	}
	for g := range globals {
		c.mutable[g] = true
	}
	for _, p := range fn.Params {
		c.taken[p.Name] = true
	}
	for _, id := range identRE.FindAllString(fn.RawParams, -1) {
		c.taken[id] = true
	}
	if !c.collect(fn.Body) {
		return
	}
	c.block(fn.Body, nil, 1, false)
}

type cseState struct {
	mutable   map[string]bool   // names written: never merged
	protected map[string]bool   // names read by text the pass does not parse: never dropped
	taken     map[string]bool   // every name in the function: a new one avoids them
	alias     map[string]string // a dropped let's name -> the let read in its place
	nameVN    map[string]int    // a let's name -> its value's number
	intern    map[string]int    // a value's key -> its number
	nextVN    int
	nextName  int
	parsed    map[string]*exprNode
}

// collect records what the function writes, what unparsed text reads and
// every name it uses, and checks that every expression parses.
func (c *cseState) collect(b *Block) bool {
	for _, s := range b.stmts {
		if src, ok := siteExpr(s); ok {
			n := c.parse(src)
			if n == nil {
				return false
			}
			c.takeNames(n)
		}
		switch st := s.(type) {
		case *LetStmt:
			c.taken[st.Name] = true
		case *VarStmt:
			c.taken[st.Name] = true
			c.mutable[st.Name] = true
		case *AssignStmt:
			if m := targetRE.FindStringSubmatch(st.Target + " = 0"); m != nil {
				c.mutable[m[1]] = true
			}
			for _, id := range identRE.FindAllString(st.Target, -1) {
				c.taken[id] = true
			}
		case *OpaqueStmt:
			for _, w := range st.Writes() {
				if id := identRE.FindString(w); id != "" {
					c.mutable[id] = true
				}
			}
			for _, r := range st.Reads() {
				for _, id := range identRE.FindAllString(r, -1) {
					c.protected[id] = true
				}
			}
			for _, line := range st.Lines {
				for _, id := range identRE.FindAllString(line, -1) {
					c.protected[id] = true
					c.taken[id] = true
				}
			}
		case *GuardStmt:
			for _, part := range strings.Split(st.HeaderExtra, ";") {
				part = strings.TrimSpace(part)
				if m := targetRE.FindStringSubmatch(part); m != nil && m[1] != "let" && m[1] != "var" {
					c.mutable[m[1]] = true
				}
			}
			for _, id := range identRE.FindAllString(st.HeaderExtra, -1) {
				c.protected[id] = true
				c.taken[id] = true
			}
			if !c.collect(st.Body) {
				return false
			}
		case *LoopStmt:
			c.taken[st.Index] = true
			c.mutable[st.Index] = true
			if !c.collect(st.Body) {
				return false
			}
		}
	}
	return true
}

func (c *cseState) takeNames(n *exprNode) {
	if n.isIdent() {
		c.taken[n.op] = true
	}
	for _, k := range n.kids {
		c.takeNames(k)
	}
}

func (c *cseState) parse(src string) *exprNode {
	if n, ok := c.parsed[src]; ok {
		return n
	}
	n, err := parseExpr(src)
	if err != nil {
		n = nil
	}
	c.parsed[src] = n
	return n
}

// block runs both rules on b, whose enclosing blocks' lets in scope at its
// start are env (a value's number -> the let holding it), then on the blocks
// nested in it.
func (c *cseState) block(b *Block, env map[int]string, depth int, inline bool) {
	// Rule 1: in order, renaming dropped lets, reading what a let in scope
	// holds as that let, and dropping a let that repeats one.
	scope := copyEnv(env)
	kept := make([]Stmt, 0, len(b.stmts))
	for _, s := range b.stmts {
		src, ok := siteExpr(s)
		if !ok {
			kept = append(kept, s)
			continue
		}
		src = c.renamed(src)
		root := c.parse(src)
		c.number(root)
		if let, isLet := s.(*LetStmt); isLet {
			value := unparen(root)
			if value.kind != kAtom && !value.fresh && !c.protected[let.Name] {
				if name, held := scope[value.vn]; held {
					c.alias[let.Name] = name
					c.nameVN[let.Name] = value.vn
					continue
				}
			}
		}
		var edits []cseEdit
		c.readHeld(root, nil, scope, &edits)
		if len(edits) > 0 {
			src = applyEdits(src, edits)
			root = c.parse(src)
			c.number(root)
		}
		if src != mustSiteExpr(s) {
			setSiteExpr(s, src, root)
		}
		if let, isLet := s.(*LetStmt); isLet {
			c.define(let.Name, root, scope)
		}
		kept = append(kept, s)
	}
	b.stmts = kept

	// Rule 2: name what repeats within this block.
	c.hoist(b, depth, inline)

	// Then the nested blocks, each with the lets in scope where it stands.
	scope = copyEnv(env)
	for _, s := range b.stmts {
		switch st := s.(type) {
		case *LetStmt:
			root := c.parse(st.Expr)
			c.number(root)
			c.define(st.Name, root, scope)
		case *GuardStmt:
			c.block(st.Body, scope, depth+1, st.Inline && st.Body.Len() == 1)
		case *LoopStmt:
			c.block(st.Body, scope, depth+1, false)
		}
	}
}

// define records a let: its name has its value's number, and a value worth
// reading by name joins the scope unless a let there already holds it.
func (c *cseState) define(name string, root *exprNode, scope map[int]string) {
	value := unparen(root)
	c.nameVN[name] = value.vn
	if value.kind == kAtom || value.fresh {
		return
	}
	if _, held := scope[value.vn]; !held {
		scope[value.vn] = name
	}
}

// renamed puts each dropped let's replacement in place of its name.
func (c *cseState) renamed(src string) string {
	if len(c.alias) == 0 {
		return src
	}
	root := c.parse(src)
	var edits []cseEdit
	var walk func(n *exprNode)
	walk = func(n *exprNode) {
		if n.isIdent() {
			if to, ok := c.alias[n.op]; ok {
				edits = append(edits, cseEdit{n.start, n.end, to})
			}
			return
		}
		for _, k := range n.kids {
			walk(k)
		}
	}
	walk(root)
	if len(edits) == 0 {
		return src
	}
	return applyEdits(src, edits)
}

// readHeld replaces, outermost first, each subexpression whose value a let in
// scope holds by that let's name, where the name is the shorter text.
func (c *cseState) readHeld(n, paren *exprNode, scope map[int]string, edits *[]cseEdit) {
	if n.kind != kAtom && n.kind != kParen && !n.fresh {
		if name, held := scope[n.vn]; held {
			span := n
			if paren != nil {
				span = paren
			}
			if len(name) < span.end-span.start {
				*edits = append(*edits, cseEdit{span.start, span.end, name})
				return
			}
		}
	}
	var inner *exprNode
	if n.kind == kParen {
		inner = n
	}
	for _, k := range n.kids {
		c.readHeld(k, inner, scope, edits)
	}
}

type cseOcc struct {
	site    int       // index into the block's statements
	node    *exprNode // the subexpression
	span    *exprNode // what a name replaces: the node, or the parentheses round it
	counted bool      // not on the right of && or ||
	letRoot bool      // a let's whole value
}

// hoist applies rule 2 to one block until nothing repeating is worth naming.
func (c *cseState) hoist(b *Block, depth int, inline bool) {
	for {
		srcs := make([]string, len(b.stmts))
		occs := make(map[int][]cseOcc)
		var order []int
		for i, s := range b.stmts {
			src, ok := siteExpr(s)
			if !ok {
				continue
			}
			srcs[i] = src
			root := c.parse(src)
			c.number(root)
			_, isLet := s.(*LetStmt)
			value := unparen(root)
			var walk func(n, parent *exprNode, short bool)
			walk = func(n, parent *exprNode, short bool) {
				if n.kind != kAtom && n.kind != kParen && !n.fresh && n.hasIdent {
					span := n
					if parent != nil && parent.kind == kParen {
						span = parent
					}
					if _, seen := occs[n.vn]; !seen {
						order = append(order, n.vn)
					}
					occs[n.vn] = append(occs[n.vn], cseOcc{site: i, node: n, span: span, counted: !short, letRoot: isLet && n == value})
				}
				for k, kid := range n.kids {
					walk(kid, n, short || (n.kind == kBinary && k == 1 && (n.op == "&&" || n.op == "||")))
				}
			}
			walk(root, nil, false)
		}

		best, bestLen, bestMove := -1, 0, -1
		for _, vn := range order {
			list := occs[vn]
			first := -1
			counted := 0
			for k, o := range list {
				if o.counted {
					counted++
					if first < 0 {
						first = k
					}
				}
			}
			if counted < 2 {
				continue
			}
			lead := list[first]
			textLen := lead.node.end - lead.node.start
			if textLen <= bestLen {
				continue
			}
			move := -1
			for k, o := range list {
				if o.counted && o.letRoot {
					move = k
					break
				}
			}
			before, after := 0, 0
			if move >= 0 {
				name := b.stmts[list[move].site].(*LetStmt).Name
				for k, o := range list {
					if o.counted && k != move {
						before += o.span.end - o.span.start
						after += len(name)
					}
				}
			} else {
				name := c.peekName()
				for _, o := range list {
					if o.counted {
						before += o.span.end - o.span.start
						after += len(name)
					}
				}
				// The new line, indented, and, in a guard written on one
				// line, the braces and indentation it then spreads over.
				after += 2*depth + len("let ") + len(name) + len(" = ") + textLen + len(";\n")
				if inline && len(b.stmts) == 1 {
					after += 2*depth + 2*(depth-1)
				}
			}
			if after < before {
				best, bestLen, bestMove = vn, textLen, move
			}
		}
		if best < 0 {
			return
		}

		list := occs[best]
		firstSite := len(b.stmts)
		for _, o := range list {
			if o.counted && o.site < firstSite {
				firstSite = o.site
			}
		}
		var name string
		var moved Stmt
		if bestMove >= 0 {
			moved = b.stmts[list[bestMove].site]
			name = moved.(*LetStmt).Name
		} else {
			lead := list[0]
			for _, o := range list {
				if o.counted {
					lead = o
					break
				}
			}
			name = c.takeName()
			text := srcs[lead.site][lead.node.start:lead.node.end]
			c.nameVN[name] = best
			moved = NewLetStmt(name, "", text, nil)
			setSiteExpr(moved, text, c.parse(text))
		}
		edits := make(map[int][]cseEdit)
		for k, o := range list {
			if k == bestMove || o.site < firstSite {
				continue
			}
			edits[o.site] = append(edits[o.site], cseEdit{o.span.start, o.span.end, name})
		}
		for site, es := range edits {
			src := applyEdits(srcs[site], es)
			setSiteExpr(b.stmts[site], src, c.parse(src))
		}
		if bestMove >= 0 {
			from := list[bestMove].site
			if from != firstSite {
				// The let moves above whatever lies between: if it reads a
				// name declared there, it takes the first use's text, the same
				// value spelled in names already declared where it now stands.
				declared := map[string]bool{}
				for _, s := range b.stmts[firstSite:from] {
					switch d := s.(type) {
					case *LetStmt:
						declared[d.Name] = true
					case *VarStmt:
						declared[d.Name] = true
					}
				}
				for _, id := range identsOf(c.parse(mustSiteExpr(moved))) {
					if declared[id] {
						lead := list[0]
						for _, o := range list {
							if o.counted && o.site == firstSite {
								lead = o
								break
							}
						}
						text := srcs[lead.site][lead.node.start:lead.node.end]
						setSiteExpr(moved, text, c.parse(text))
						break
					}
				}
				b.stmts = append(b.stmts[:from], b.stmts[from+1:]...)
				b.stmts = append(b.stmts[:firstSite], append([]Stmt{moved}, b.stmts[firstSite:]...)...)
			}
		} else {
			b.stmts = append(b.stmts[:firstSite], append([]Stmt{moved}, b.stmts[firstSite:]...)...)
		}
	}
}

func (c *cseState) peekName() string {
	for n := c.nextName + 1; ; n++ {
		name := fmt.Sprintf("gsl_cse_%d", n)
		if !c.taken[name] {
			return name
		}
	}
}

func (c *cseState) takeName() string {
	for {
		c.nextName++
		name := fmt.Sprintf("gsl_cse_%d", c.nextName)
		if !c.taken[name] {
			c.taken[name] = true
			return name
		}
	}
}

// number gives every node of the tree its value's number. A name the
// function writes reads a fresh number at every occurrence, so nothing that
// reads it equals anything; a let's name reads its value's.
func (c *cseState) number(n *exprNode) {
	n.fresh, n.hasIdent = false, false
	switch n.kind {
	case kAtom:
		if !n.isIdent() {
			n.vn = c.key("lit:" + n.op)
			return
		}
		n.hasIdent = true
		if c.mutable[n.op] {
			c.nextVN++
			n.vn = -c.nextVN
			n.fresh = true
			return
		}
		if vn, ok := c.nameVN[n.op]; ok {
			n.vn = vn
			return
		}
		n.vn = c.key("id:" + n.op)
		return
	case kParen:
		k := n.kids[0]
		c.number(k)
		n.vn, n.fresh, n.hasIdent = k.vn, k.fresh, k.hasIdent
		return
	}
	var key strings.Builder
	fmt.Fprintf(&key, "%d:%s", n.kind, n.op)
	for _, k := range n.kids {
		c.number(k)
		fmt.Fprintf(&key, ",%d", k.vn)
		n.fresh = n.fresh || k.fresh
		n.hasIdent = n.hasIdent || k.hasIdent
	}
	if n.fresh {
		c.nextVN++
		n.vn = -c.nextVN
		return
	}
	n.vn = c.key(key.String())
}

func (c *cseState) key(k string) int {
	if vn, ok := c.intern[k]; ok {
		return vn
	}
	c.nextVN++
	c.intern[k] = c.nextVN
	return c.nextVN
}

func copyEnv(env map[int]string) map[int]string {
	out := make(map[int]string, len(env))
	for k, v := range env {
		out[k] = v
	}
	return out
}

func unparen(n *exprNode) *exprNode {
	for n.kind == kParen {
		n = n.kids[0]
	}
	return n
}

type cseEdit struct {
	start, end int
	text       string
}

func applyEdits(src string, edits []cseEdit) string {
	sort.Slice(edits, func(i, j int) bool { return edits[i].start > edits[j].start })
	for _, e := range edits {
		src = src[:e.start] + e.text + src[e.end:]
	}
	return src
}

// siteExpr is the expression a statement evaluates in its own block, if it
// has one: a let's, a var's initializer, an assignment's, a return's or a
// guard's condition.
func siteExpr(s Stmt) (string, bool) {
	switch st := s.(type) {
	case *LetStmt:
		return st.Expr, true
	case *VarStmt:
		return st.InitExpr, st.InitExpr != ""
	case *AssignStmt:
		return st.Expr, true
	case *ReturnStmt:
		return st.Expr, true
	case *GuardStmt:
		return st.Condition, true
	}
	return "", false
}

func mustSiteExpr(s Stmt) string {
	src, _ := siteExpr(s)
	return src
}

// setSiteExpr rewrites a statement's expression and what it reads.
func setSiteExpr(s Stmt, src string, root *exprNode) {
	reads := identsOf(root)
	switch st := s.(type) {
	case *LetStmt:
		st.Expr, st.reads = src, reads
	case *VarStmt:
		st.InitExpr, st.reads = src, reads
	case *AssignStmt:
		st.Expr, st.reads = src, reads
	case *ReturnStmt:
		st.Expr, st.reads = src, reads
	case *GuardStmt:
		st.Condition = src
		seen := make(map[string]bool)
		for _, r := range st.reads {
			seen[r] = true
		}
		for _, r := range reads {
			if !seen[r] {
				seen[r] = true
				st.reads = append(st.reads, r)
			}
		}
	}
}

func identsOf(root *exprNode) []string {
	var out []string
	seen := make(map[string]bool)
	var walk func(n *exprNode)
	walk = func(n *exprNode) {
		if n.isIdent() {
			if !seen[n.op] {
				seen[n.op] = true
				out = append(out, n.op)
			}
			return
		}
		for _, k := range n.kids {
			walk(k)
		}
	}
	if root != nil {
		walk(root)
	}
	return out
}

// ---------------------------------------------------------------------------
// The expression parser: the WGSL the codegen writes. Literals, identifiers,
// calls (a typed constructor's template kept as written), unary and binary
// operators at WGSL's precedence, member and index access, parentheses.

type nodeKind int

const (
	kAtom nodeKind = iota
	kCall
	kBinary
	kUnary
	kMember
	kIndex
	kParen
)

// exprNode is a parsed expression: its kind, its operator (an atom's text, a
// call's callee, a member's name), its operands and its span in the source.
type exprNode struct {
	kind       nodeKind
	op         string
	kids       []*exprNode
	start, end int

	vn       int
	fresh    bool
	hasIdent bool
}

func (n *exprNode) isIdent() bool {
	return n.kind == kAtom && len(n.op) > 0 && isIdentStart(n.op[0])
}

// sexpr prints the tree for tests.
func (n *exprNode) sexpr() string {
	switch n.kind {
	case kAtom:
		return n.op
	case kParen:
		return n.kids[0].sexpr()
	case kCall:
		parts := []string{"call", n.op}
		for _, k := range n.kids {
			parts = append(parts, k.sexpr())
		}
		return "(" + strings.Join(parts, " ") + ")"
	case kBinary:
		return "(" + n.op + " " + n.kids[0].sexpr() + " " + n.kids[1].sexpr() + ")"
	case kUnary:
		return "(u" + n.op + " " + n.kids[0].sexpr() + ")"
	case kMember:
		return "(. " + n.kids[0].sexpr() + " " + n.op + ")"
	case kIndex:
		return "([] " + n.kids[0].sexpr() + " " + n.kids[1].sexpr() + ")"
	}
	return "?"
}

type tokKind int

const (
	tIdent tokKind = iota
	tNum
	tPunct
	tEOF
)

type token struct {
	kind tokKind
	text string
	pos  int
}

var twoCharPuncts = map[string]bool{"<<": true, ">>": true, "<=": true, ">=": true, "==": true, "!=": true, "&&": true, "||": true}

func isIdentStart(ch byte) bool {
	return ch == '_' || (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

func isDigit(ch byte) bool { return ch >= '0' && ch <= '9' }

func isIdentChar(ch byte) bool { return isIdentStart(ch) || isDigit(ch) }

func isHexDigit(ch byte) bool {
	return isDigit(ch) || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

func lexExpr(src string) ([]token, error) {
	var toks []token
	for i := 0; i < len(src); {
		ch := src[i]
		switch {
		case ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r':
			i++
		case isIdentStart(ch):
			j := i + 1
			for j < len(src) && isIdentChar(src[j]) {
				j++
			}
			toks = append(toks, token{tIdent, src[i:j], i})
			i = j
		case isDigit(ch) || (ch == '.' && i+1 < len(src) && isDigit(src[i+1])):
			j := scanNumber(src, i)
			toks = append(toks, token{tNum, src[i:j], i})
			i = j
		default:
			if i+1 < len(src) && twoCharPuncts[src[i:i+2]] {
				toks = append(toks, token{tPunct, src[i : i+2], i})
				i += 2
				continue
			}
			if strings.IndexByte("()[],.<>+-*/%!~&|^", ch) < 0 {
				return nil, fmt.Errorf("unexpected %q at %d", ch, i)
			}
			toks = append(toks, token{tPunct, src[i : i+1], i})
			i++
		}
	}
	return append(toks, token{tEOF, "", len(src)}), nil
}

func scanNumber(s string, i int) int {
	j := i
	if s[j] == '0' && j+1 < len(s) && (s[j+1] == 'x' || s[j+1] == 'X') {
		j += 2
		for j < len(s) && (isHexDigit(s[j]) || s[j] == '.') {
			j++
		}
		if j < len(s) && (s[j] == 'p' || s[j] == 'P') {
			j++
			if j < len(s) && (s[j] == '+' || s[j] == '-') {
				j++
			}
			for j < len(s) && isDigit(s[j]) {
				j++
			}
		}
	} else {
		for j < len(s) && isDigit(s[j]) {
			j++
		}
		if j < len(s) && s[j] == '.' {
			j++
			for j < len(s) && isDigit(s[j]) {
				j++
			}
		}
		if j < len(s) && (s[j] == 'e' || s[j] == 'E') {
			k := j + 1
			if k < len(s) && (s[k] == '+' || s[k] == '-') {
				k++
			}
			if k < len(s) && isDigit(s[k]) {
				j = k
				for j < len(s) && isDigit(s[j]) {
					j++
				}
			}
		}
	}
	if j < len(s) && strings.IndexByte("uifh", s[j]) >= 0 {
		j++
	}
	return j
}

var binaryLevel = map[string]int{
	"||": 1, "&&": 2, "|": 3, "^": 4, "&": 5,
	"==": 6, "!=": 6,
	"<": 7, ">": 7, "<=": 7, ">=": 7,
	"<<": 8, ">>": 8,
	"+": 9, "-": 9,
	"*": 10, "/": 10, "%": 10,
}

type exprParser struct {
	src  string
	toks []token
	i    int
}

// parseExpr parses one WGSL expression as the codegen emits it.
func parseExpr(src string) (*exprNode, error) {
	toks, err := lexExpr(src)
	if err != nil {
		return nil, err
	}
	p := &exprParser{src: src, toks: toks}
	n, err := p.binary(1)
	if err != nil {
		return nil, err
	}
	if t := p.peek(); t.kind != tEOF {
		return nil, fmt.Errorf("unexpected %q at %d", t.text, t.pos)
	}
	return n, nil
}

func (p *exprParser) peek() token { return p.toks[p.i] }

func (p *exprParser) isPunct(text string) bool {
	t := p.toks[p.i]
	return t.kind == tPunct && t.text == text
}

func (p *exprParser) expect(text string) (token, error) {
	t := p.toks[p.i]
	if t.kind != tPunct || t.text != text {
		if t.kind == tEOF {
			return t, fmt.Errorf("expected %q at end", text)
		}
		return t, fmt.Errorf("expected %q at %d, found %q", text, t.pos, t.text)
	}
	p.i++
	return t, nil
}

func (p *exprParser) binary(minLevel int) (*exprNode, error) {
	left, err := p.unary()
	if err != nil {
		return nil, err
	}
	for {
		t := p.peek()
		level, ok := binaryLevel[t.text]
		if t.kind != tPunct || !ok || level < minLevel {
			return left, nil
		}
		p.i++
		right, err := p.binary(level + 1)
		if err != nil {
			return nil, err
		}
		left = &exprNode{kind: kBinary, op: t.text, kids: []*exprNode{left, right}, start: left.start, end: right.end}
	}
}

func (p *exprParser) unary() (*exprNode, error) {
	t := p.peek()
	if t.kind == tPunct && strings.Contains("-!~&*", t.text) && len(t.text) == 1 {
		p.i++
		operand, err := p.unary()
		if err != nil {
			return nil, err
		}
		return &exprNode{kind: kUnary, op: t.text, kids: []*exprNode{operand}, start: t.pos, end: operand.end}, nil
	}
	return p.postfix()
}

func (p *exprParser) postfix() (*exprNode, error) {
	n, err := p.primary()
	if err != nil {
		return nil, err
	}
	for {
		switch {
		case p.isPunct("."):
			p.i++
			t := p.peek()
			if t.kind != tIdent {
				return nil, fmt.Errorf("expected a member name at %d", t.pos)
			}
			p.i++
			n = &exprNode{kind: kMember, op: t.text, kids: []*exprNode{n}, start: n.start, end: t.pos + len(t.text)}
		case p.isPunct("["):
			p.i++
			index, err := p.binary(1)
			if err != nil {
				return nil, err
			}
			closing, err := p.expect("]")
			if err != nil {
				return nil, err
			}
			n = &exprNode{kind: kIndex, kids: []*exprNode{n, index}, start: n.start, end: closing.pos + 1}
		default:
			return n, nil
		}
	}
}

func (p *exprParser) primary() (*exprNode, error) {
	t := p.peek()
	switch t.kind {
	case tNum:
		p.i++
		return &exprNode{kind: kAtom, op: t.text, start: t.pos, end: t.pos + len(t.text)}, nil
	case tIdent:
		p.i++
		callee := t.text
		if p.isPunct("<") {
			if end, ok := p.template(); ok {
				callee = p.src[t.pos:end]
			}
		}
		if p.isPunct("(") {
			return p.call(callee, t.pos)
		}
		if callee != t.text {
			return nil, fmt.Errorf("a template without a call at %d", t.pos)
		}
		return &exprNode{kind: kAtom, op: t.text, start: t.pos, end: t.pos + len(t.text)}, nil
	case tPunct:
		if t.text == "(" {
			p.i++
			inner, err := p.binary(1)
			if err != nil {
				return nil, err
			}
			closing, err := p.expect(")")
			if err != nil {
				return nil, err
			}
			return &exprNode{kind: kParen, kids: []*exprNode{inner}, start: t.pos, end: closing.pos + 1}, nil
		}
	case tEOF:
		return nil, fmt.Errorf("unexpected end")
	}
	return nil, fmt.Errorf("unexpected %q at %d", t.text, t.pos)
}

// template reads a template list, `<` to its matching `>`, when one stands
// here and a call follows it; it returns the source offset past the `>`.
func (p *exprParser) template() (int, bool) {
	depth := 0
	for j := p.i; j < len(p.toks); j++ {
		t := p.toks[j]
		switch {
		case t.kind == tPunct && t.text == "<":
			depth++
		case t.kind == tPunct && (t.text == ">" || t.text == ">>"):
			depth -= len(t.text)
			if depth < 0 {
				return 0, false
			}
			if depth == 0 {
				next := p.toks[j+1]
				if next.kind != tPunct || next.text != "(" {
					return 0, false
				}
				p.i = j + 1
				return t.pos + len(t.text), true
			}
		case t.kind == tIdent || t.kind == tNum || (t.kind == tPunct && t.text == ","):
		default:
			return 0, false
		}
	}
	return 0, false
}

func (p *exprParser) call(callee string, start int) (*exprNode, error) {
	p.i++ // (
	n := &exprNode{kind: kCall, op: callee, start: start}
	if p.isPunct(")") {
		n.end = p.peek().pos + 1
		p.i++
		return n, nil
	}
	for {
		arg, err := p.binary(1)
		if err != nil {
			return nil, err
		}
		n.kids = append(n.kids, arg)
		if p.isPunct(",") {
			p.i++
			continue
		}
		closing, err := p.expect(")")
		if err != nil {
			return nil, err
		}
		n.end = closing.pos + 1
		return n, nil
	}
}
