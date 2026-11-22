package ir

import (
	"regexp"
	"strings"
)

// Guards that share a condition.
//
// A form split at its pitch diameter (a thread's crests and roots, a gear's
// tips and gaps) is two features of one size, and CSE gives both guards the
// one weight, so the function holds two `if (w > 0.0) { ... }` whose bodies
// each evaluate the whole form; CSE never hoists out of a guard, so the form
// is computed twice. MergeGuards moves a later guard into an earlier one in
// the same block when the condition and header are the same text, and CSE run
// after it names the form once inside the one branch.
//
// Moving the later guard B up into the earlier guard A is legal when every
// statement S between them either
//
//   - commutes with A and with every statement kept behind before it (S reads
//     nothing they write and writes nothing they read or write), and moves
//     above A; or
//   - stays behind, below the merged guard, and commutes with B, which now
//     runs before it.
//
// B's condition must also read nothing A's body writes, or it would be read
// before the write it followed; a let or var B's body declares must not
// already be declared at the top of A's, the one scope they now share. A
// return or an opaque statement between them, or in either body, stops the
// merge, since what it does is not known here. An inline guard is left alone.
// Reads are every identifier in a statement's text, a superset; writes are a
// let's, var's or assignment's name, a loop's index and a body's writes, and
// a private global that a called function of the module writes.

// MergeGuards merges the guards of a structured function's body that share a
// condition, as the comment above describes. A raw function is left as it is.
func MergeGuards(fn *FuncDecl) {
	mergeGuards(fn, nil)
}

// MergeGuards runs the pass on every structured function of the module, a
// call to a function that writes a private global counted as writing it.
func (m *Module) MergeGuards() {
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
	fns := append(append([]*FuncDecl{}, m.Imports...), m.Functions...)
	for _, fn := range fns {
		scan(fn.RawSource)
		scan(fn.TrailingSource)
	}
	writers := globalWriters(fns, globals)
	for _, fn := range fns {
		mergeGuards(fn, writers)
	}
}

var (
	callRE       = regexp.MustCompile(`([A-Za-z_][A-Za-z0-9_]*)\s*\(`)
	assignLineRE = regexp.MustCompile(`(?m)^\s*([A-Za-z_][A-Za-z0-9_]*)\s*(?:\.[A-Za-z_][A-Za-z0-9_]*|\[[^\]]*\])*\s*=(?:[^=]|$)`)
)

// globalWriters maps each function of the module that writes a private
// global, itself or through a call, to the globals it writes.
func globalWriters(fns []*FuncDecl, globals map[string]bool) map[string]map[string]bool {
	text := make(map[string]string, len(fns))
	for _, fn := range fns {
		var b strings.Builder
		if fn.RawSource != "" {
			b.WriteString(fn.RawSource)
		} else if err := fn.Format(&b); err != nil {
			b.Reset()
		}
		text[fn.Name] = b.String()
	}
	writers := make(map[string]map[string]bool)
	for name, src := range text {
		for _, m := range assignLineRE.FindAllStringSubmatch(src, -1) {
			if globals[m[1]] {
				if writers[name] == nil {
					writers[name] = make(map[string]bool)
				}
				writers[name][m[1]] = true
			}
		}
	}
	for changed := true; changed; {
		changed = false
		for name, src := range text {
			for _, m := range callRE.FindAllStringSubmatch(src, -1) {
				for g := range writers[m[1]] {
					if writers[name] == nil {
						writers[name] = make(map[string]bool)
					}
					if !writers[name][g] {
						writers[name][g] = true
						changed = true
					}
				}
			}
		}
	}
	return writers
}

func mergeGuards(fn *FuncDecl, writers map[string]map[string]bool) {
	if fn == nil || fn.RawSource != "" || fn.Body == nil {
		return
	}
	g := &guardMerger{writers: writers}
	g.block(fn.Body)
}

type guardMerger struct {
	writers map[string]map[string]bool
}

// block merges the guards of b, then of every block nested in it.
func (g *guardMerger) block(b *Block) {
	for i := 0; i < len(b.stmts); i++ {
		a, ok := b.stmts[i].(*GuardStmt)
		if !ok || a.Inline {
			continue
		}
		for j := i + 1; j < len(b.stmts); j++ {
			later, ok := b.stmts[j].(*GuardStmt)
			if !ok || later.Inline || later.Condition != a.Condition || later.HeaderExtra != a.HeaderExtra {
				continue
			}
			if at, merged := g.merge(b, i, j); merged {
				// A may have moved down by what was hoisted above it; look for
				// more guards to merge into it from where it now stands.
				i, j = at, at
				a = b.stmts[at].(*GuardStmt)
			}
		}
	}
	for _, s := range b.stmts {
		switch s := s.(type) {
		case *GuardStmt:
			g.block(s.Body)
		case *LoopStmt:
			g.block(s.Body)
		}
	}
}

// merge moves the guard at j into the guard at i where that is legal, and
// reports the merged guard's new index.
func (g *guardMerger) merge(b *Block, i, j int) (int, bool) {
	a, later := b.stmts[i].(*GuardStmt), b.stmts[j].(*GuardStmt)
	aR, aW, okA := g.effects(a)
	bR, bW, okB := g.effects(later)
	if !okA || !okB {
		return i, false
	}
	if overlaps(identSet(later.Condition+" "+later.HeaderExtra), aW) {
		return i, false
	}
	declared := make(map[string]bool)
	for _, s := range a.Body.stmts {
		if name := declaredName(s); name != "" {
			declared[name] = true
		}
	}
	for _, s := range later.Body.stmts {
		if name := declaredName(s); name != "" && declared[name] {
			return i, false
		}
	}
	var hoisted, behind []Stmt
	behindR, behindW := make(map[string]bool), make(map[string]bool)
	for _, s := range b.stmts[i+1 : j] {
		sR, sW, ok := g.effects(s)
		if !ok {
			return i, false
		}
		commutesA := !overlaps(sR, aW) && !overlaps(sW, aR) && !overlaps(sW, aW)
		commutesBehind := !overlaps(sR, behindW) && !overlaps(sW, behindR) && !overlaps(sW, behindW)
		if commutesA && commutesBehind {
			hoisted = append(hoisted, s)
			continue
		}
		if overlaps(bR, sW) || overlaps(bW, sR) || overlaps(bW, sW) {
			return i, false
		}
		behind = append(behind, s)
		union(behindR, sR)
		union(behindW, sW)
	}
	body := NewBlock()
	body.stmts = append(append(body.stmts, a.Body.stmts...), later.Body.stmts...)
	merged := NewGuardStmtWithHeader(a.Condition, a.HeaderExtra, body, a.HeaderReads())
	out := make([]Stmt, 0, len(b.stmts)-1)
	out = append(out, b.stmts[:i]...)
	out = append(out, hoisted...)
	at := len(out)
	out = append(out, merged)
	out = append(out, behind...)
	out = append(out, b.stmts[j+1:]...)
	b.stmts = out
	return at, true
}

// effects returns what a statement reads (every identifier in its text) and
// writes, and false for a statement whose effects are not known here: a
// return or an opaque statement, at any depth.
func (g *guardMerger) effects(s Stmt) (reads, writes map[string]bool, ok bool) {
	reads, writes = make(map[string]bool), make(map[string]bool)
	if !g.collect(s, reads, writes) {
		return nil, nil, false
	}
	return reads, writes, true
}

func (g *guardMerger) collect(s Stmt, reads, writes map[string]bool) bool {
	text := func(src string) {
		union(reads, identSet(src))
		for _, m := range callRE.FindAllStringSubmatch(src, -1) {
			union(writes, g.writers[m[1]])
		}
	}
	switch s := s.(type) {
	case *LetStmt:
		text(s.Expr)
		writes[s.Name] = true
	case *VarStmt:
		text(s.InitExpr)
		writes[s.Name] = true
	case *AssignStmt:
		text(s.Target + " " + s.Expr)
		if m := identRE.FindString(s.Target); m != "" {
			writes[m] = true
		}
	case *GuardStmt:
		text(s.Condition + " " + s.HeaderExtra)
		for _, inner := range s.Body.stmts {
			if !g.collect(inner, reads, writes) {
				return false
			}
		}
	case *LoopStmt:
		writes[s.Index] = true
		reads[s.Index] = true
		for _, inner := range s.Body.stmts {
			if !g.collect(inner, reads, writes) {
				return false
			}
		}
	default:
		return false
	}
	return true
}

// declaredName is the name a let or var declares, or "".
func declaredName(s Stmt) string {
	switch s := s.(type) {
	case *LetStmt:
		return s.Name
	case *VarStmt:
		return s.Name
	}
	return ""
}

func identSet(src string) map[string]bool {
	set := make(map[string]bool)
	for _, id := range identRE.FindAllString(src, -1) {
		set[id] = true
	}
	return set
}

func overlaps(a, b map[string]bool) bool {
	if len(a) > len(b) {
		a, b = b, a
	}
	for k := range a {
		if b[k] {
			return true
		}
	}
	return false
}

func union(dst, src map[string]bool) {
	for k := range src {
		dst[k] = true
	}
}
