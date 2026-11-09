// Package boxsweep checks every box the examples have: where a named value's
// box bounds anything, interval arithmetic proves the value has no surface
// outside it. The proofs grow with every example that gets a box, so the
// examples' values are split across the shard packages below this one
// (shard0 ...), each its own test binary under its own timeout, apart from
// gsleval's.
package boxsweep

import (
	"context"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	fidgetapi "gradientcontrol.com/omega/core/internal/fidget"
	"gradientcontrol.com/omega/core/internal/gsleval"
	"gradientcontrol.com/omega/core/internal/source"
)

// unboundedAxis is what an axis .box cannot bound reads as: builtin's nothing.
const unboundedAxis = 1e5

// Sweep checks every named value in the examples of one shard against its box:
// where the box bounds anything, interval arithmetic proves the value has no
// zero outside it (grown by a hair, since a surface may touch it), so no
// surface escapes. A disproof fails the test; a proof the interval search
// cannot finish within its budget is logged, not failed. The examples' named
// values, the examples sorted and each one's values sorted, are dealt to shards
// round-robin: shard takes those whose index modulo shards is shard. Dealing
// values rather than whole examples spreads a large example over every shard,
// and with the query cache each value costs about one assembly wherever it is.
//
// With GSL_BOX_COVERAGE set it also reports how many of each example's fields
// (scalar values that vary with p) got a box, which costs compiling every one;
// a shard counts only its own values, so the shards' reports add up.
func Sweep(t *testing.T, shard, shards int) {
	if testing.Short() {
		t.Skip("interval proofs over every example")
	}
	files, err := filepath.Glob(filepath.Join(repoRoot(), "internal/testdata/example/*.gsl"))
	if err != nil || len(files) == 0 {
		t.Fatalf("no examples: %v", err)
	}
	e := newBoxEvaluator()
	e.SetCertifyBudget(20000)
	// A value is proved over each side of its box with the same query.
	e.CacheQueries(4)
	ctx := context.Background()
	coverage := os.Getenv("GSL_BOX_COVERAGE") != ""

	var bounded, proven, indeterminate, fields int
	var perExample []string
	dealt := 0
	for _, file := range files {
		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		code := string(src)
		ex := filepath.Base(file)
		var names []string
		for _, n := range valueNames(code) {
			if dealt%shards == shard {
				names = append(names, n)
			}
			dealt++
		}
		if len(names) == 0 {
			continue
		}
		t.Run(strings.TrimSuffix(ex, ".gsl"), func(t *testing.T) {
			boxes := namedBoxes(ctx, e, code, names)

			exBounded := 0
			for _, n := range names {
				bx, ok := boxes[n]
				if !ok || !anyBounded(bx) {
					continue
				}
				for _, slab := range outsideSlabs(bx) {
					res, err := certifyOneSign(ctx, e, code, n, slab)
					if err != nil {
						t.Errorf("%s: certify %s outside %v: %v", ex, n, bx, err)
						break
					}
					if res.Disproven {
						t.Errorf("%s: %s has surface outside its box %v, in %v", ex, n, bx, res.CounterBox)
						break
					}
					if res.Indeterminate {
						indeterminate++
					} else {
						proven++
					}
				}
				if len(outsideSlabs(bx)) > 0 {
					bounded++
					exBounded++
				}
			}
			if coverage {
				exFields := len(scalarFields(ctx, e, code, names))
				fields += exFields
				perExample = append(perExample, fmt.Sprintf("%s %d/%d", ex, exBounded, exFields))
			}
		})
	}
	t.Logf("%d named values with a box; outside-the-box proofs: %d proven, %d indeterminate",
		bounded, proven, indeterminate)
	if coverage {
		t.Logf("%d of %d fields have a box; per example: %s", bounded, fields, strings.Join(perExample, ", "))
	}
}

// repoRoot is the repository root, found from this file, so every shard
// package reads the same examples and libraries wherever it runs from.
func repoRoot() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "../../..")
}

// namedBoxes reads each name's .box, all at once, and one at a time only when
// the batch fails. A name whose box cannot be read is left out.
func namedBoxes(ctx context.Context, e *gsleval.Evaluator, code string, names []string) map[string][6]float64 {
	out := map[string][6]float64{}
	query := func(ns []string) bool {
		qs := make([]string, len(ns))
		for i, n := range ns {
			qs[i] = n + ".box"
		}
		groups, err := e.EvalGsl(ctx, nil, code, "", qs, [][3]float64{{0, 0, 0}})
		if err != nil || len(groups) != len(ns) {
			return false
		}
		for i, n := range ns {
			if bx, ok := boxOf(groups[i].Samples); ok {
				out[n] = bx
			}
		}
		return true
	}
	if len(names) == 0 || query(names) {
		return out
	}
	for _, n := range names {
		query([]string{n})
	}
	return out
}

// scalarFields keeps the names whose value is a scalar that varies with p: a
// field. A constant, a vector or a list has no solid. It asks for every name at
// once, and one at a time only when the batch fails.
func scalarFields(ctx context.Context, e *gsleval.Evaluator, code string, names []string) []string {
	pts := [][3]float64{{0.31, -0.17, 0.23}, {-1.9, 2.3, 1.1}}
	isField := func(g fidgetapi.GradGroup) bool {
		if len(g.Samples) != 1 || len(g.Samples[0]) != len(pts) {
			return false
		}
		for _, s := range g.Samples[0] {
			if s.Dx != 0 || s.Dy != 0 || s.Dz != 0 {
				return true
			}
		}
		return g.Samples[0][0].Value != g.Samples[0][1].Value
	}
	var out []string
	if groups, err := e.EvalGsl(ctx, nil, code, "", names, pts); err == nil && len(groups) == len(names) {
		for i, n := range names {
			if isField(groups[i]) {
				out = append(out, n)
			}
		}
		return out
	}
	for _, n := range names {
		g, err := e.EvalGsl(ctx, nil, code, "", []string{n}, pts)
		if err == nil && len(g) == 1 && isField(g[0]) {
			out = append(out, n)
		}
	}
	return out
}

// valueNames lists the document's top-level value declarations, sorted,
// leaving out its functions and its return: neither is a value to query.
func valueNames(code string) []string {
	set := map[string]bool{}
	for _, st := range topLevelStatements(code) {
		if st.name == "" || st.isImport || st.name == "return" {
			continue
		}
		rest := strings.TrimSpace(st.text)
		if i := strings.Index(rest, st.name); i >= 0 {
			rest = strings.TrimSpace(rest[i+len(st.name):])
		}
		if strings.HasPrefix(rest, "(") {
			continue
		}
		set[st.name] = true
	}
	out := make([]string, 0, len(set))
	for n := range set {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}

func newBoxEvaluator() *gsleval.Evaluator {
	return gsleval.New(source.NewLocalManager(filepath.Join(repoRoot(), "stdlib")), nil)
}

// boxOf reads a [lo, hi] result at the first point. Samples are indexed by
// component, then by point.
func boxOf(samples [][]fidgetapi.GradSample) ([6]float64, bool) {
	var bx [6]float64
	if len(samples) != 6 {
		return bx, false
	}
	for k := range bx {
		bx[k] = float64(samples[k][0].Value)
	}
	return bx, true
}

// certifyOneSign proves that expr has no zero in slab. A slab is connected, and
// a field with no zero in it keeps one sign there: >= 0, as a solid inside its
// box is, or <= 0, as a complement is. It proves the first and tries the second
// only when the first is disproven, so a disproof means both signs are there,
// and a zero between them.
func certifyOneSign(ctx context.Context, e *gsleval.Evaluator, code, expr string, slab [6]float64) (*gsleval.CertifyResult, error) {
	res, err := e.Certify(ctx, nil, code, "", expr, slab, 0, 0)
	if err != nil || !res.Disproven {
		return res, err
	}
	return e.Certify(ctx, nil, code, "", "-("+expr+")", slab, 0, 0)
}

func finiteAxis(v float64) bool { return math.Abs(v) < unboundedAxis/2 }

func anyBounded(bx [6]float64) bool {
	for _, v := range bx {
		if finiteAxis(v) {
			return true
		}
	}
	return false
}

// outsideSlabs covers a search region minus the box, grown by a hair, with one
// slab per bounded side. The region reaches well past the box on every side,
// and is capped where an axis is unbounded.
func outsideSlabs(bx [6]float64) [][6]float64 {
	return outsideSlabsBy(bx, 0)
}

// outsideSlabsBy is outsideSlabs with the box grown by margin as well as the
// hair.
func outsideSlabsBy(bx [6]float64, margin float64) [][6]float64 {
	reach := 1.0
	for _, v := range bx {
		if finiteAxis(v) {
			reach = math.Max(reach, math.Abs(v))
		}
	}
	reach = 2*reach + 10
	hair := 1e-3*reach + margin
	var region [6]float64
	for k := 0; k < 3; k++ {
		region[k], region[k+3] = -reach, reach
	}
	var out [][6]float64
	for k := 0; k < 3; k++ {
		if lo := bx[k]; finiteAxis(lo) && lo-hair > -reach {
			s := region
			s[k+3] = lo - hair
			out = append(out, s)
		}
		if hi := bx[k+3]; finiteAxis(hi) && hi+hair < reach {
			s := region
			s[k] = hi + hair
			out = append(out, s)
		}
	}
	return out
}

// The statement splitter below is gsleval's (gsleval.go), copied because it is
// unexported there: valueNames reads a document's top-level names with it.

// gslStmt is one top-level GSL statement plus its leading identifier.
type gslStmt struct {
	text     string // raw text, including leading whitespace and terminator
	name     string // leading identifier ("" if none)
	isImport bool   // leading keyword is "import"
}

// topLevelStatements splits GSL source into top-level statements. A
// statement ends at a top-level ';' (assignments, returns) or at the '}'
// that closes a top-level '{...}' block (function declarations). Comments
// and strings are opaque. Concatenating the statement texts reproduces the
// source exactly, so callers can drop statements and re-emit the rest
// verbatim.
func topLevelStatements(source string) []gslStmt {
	var stmts []gslStmt
	var buf strings.Builder
	flush := func() {
		text := buf.String()
		buf.Reset()
		if strings.TrimSpace(text) == "" {
			// Pure whitespace/comments between statements: attach to the
			// previous statement so nothing is lost on re-emit.
			if text != "" && len(stmts) > 0 {
				stmts[len(stmts)-1].text += text
			} else if text != "" {
				stmts = append(stmts, gslStmt{text: text})
			}
			return
		}
		name, isImport := leadingIdentifier(text)
		stmts = append(stmts, gslStmt{text: text, name: name, isImport: isImport})
	}

	depth := 0
	i, n := 0, len(source)
	for i < n {
		c := source[i]
		switch {
		case c == '/' && i+1 < n && source[i+1] == '/':
			end := strings.IndexByte(source[i:], '\n')
			if end < 0 {
				end = n - i
			}
			buf.WriteString(source[i : i+end])
			i += end
		case c == '/' && i+1 < n && source[i+1] == '*':
			end := strings.Index(source[i+2:], "*/")
			stop := n
			if end >= 0 {
				stop = i + 2 + end + 2
			}
			buf.WriteString(source[i:stop])
			i = stop
		case c == '"':
			j := i + 1
			for j < n && source[j] != '"' {
				if source[j] == '\\' {
					j++
				}
				j++
			}
			if j < n {
				j++
			}
			buf.WriteString(source[i:j])
			i = j
		case c == '{' || c == '[' || c == '(':
			depth++
			buf.WriteByte(c)
			i++
		case c == '}' || c == ']' || c == ')':
			if depth > 0 {
				depth--
			}
			buf.WriteByte(c)
			i++
			if c == '}' && depth == 0 {
				flush()
			}
		case c == ';' && depth == 0:
			buf.WriteByte(c)
			i++
			flush()
		default:
			buf.WriteByte(c)
			i++
		}
	}
	flush()
	return stmts
}

// leadingIdentifier returns the first identifier in a statement (skipping
// leading whitespace and comments) and whether it is the `import` keyword.
func leadingIdentifier(stmt string) (name string, isImport bool) {
	i, n := 0, len(stmt)
	for i < n {
		c := stmt[i]
		switch {
		case c == ' ' || c == '\t' || c == '\n' || c == '\r':
			i++
		case c == '/' && i+1 < n && stmt[i+1] == '/':
			e := strings.IndexByte(stmt[i:], '\n')
			if e < 0 {
				return "", false
			}
			i += e
		case c == '/' && i+1 < n && stmt[i+1] == '*':
			e := strings.Index(stmt[i+2:], "*/")
			if e < 0 {
				return "", false
			}
			i += 2 + e + 2
		default:
			goto ident
		}
	}
ident:
	start := i
	for i < n {
		c := stmt[i]
		if c == '_' || c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || (i > start && c >= '0' && c <= '9') {
			i++
		} else {
			break
		}
	}
	name = stmt[start:i]
	return name, name == "import"
}
