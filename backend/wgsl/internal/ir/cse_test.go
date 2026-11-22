package ir

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Shaft Modeling, Phase 18: common subexpressions in the emitted WGSL.
//
// A shaft's fold evaluates every segment four times, and the section it
// carries, a gear profile or a spline, is the same expression in every copy;
// fidget's context deduplicates, the WGSL kept them all. The pass works on a
// function's statements, whose expressions are text: it parses each, merges a
// let that repeats an earlier one in scope, and names a subexpression that
// repeats within one block when naming it shortens the function. It never
// hoists out of a guard or a loop, never merges what reads a name the
// function writes (a var, an assignment's or an opaque statement's target, a
// loop index), and never takes a subexpression from the right of && or ||.

// parseExpr's tree, printed by sexpr: an identifier or literal as its text, a
// call as (call NAME args...) with NAME's template as written, a binary
// operator as (OP left right), a unary one as (uOP operand), a member as
// (. operand name), an index as ([] operand index); parentheses add no node.
func TestCSE_ParseShapes(t *testing.T) {
	for src, want := range map[string]string{
		"a * b + c":                          "(+ (* a b) c)",
		"a - b - c":                          "(- (- a b) c)",
		"a / b * c % d":                      "(% (* (/ a b) c) d)",
		"(a + b) * c":                        "(* (+ a b) c)",
		"-a.x * 2.0":                         "(* (u- (. a x)) 2.0)",
		"-100000.0 - p.z":                    "(- (u- 100000.0) (. p z))",
		"vec3<f32>(p.x, -p.y, p.z).xy":       "(. (call vec3<f32> (. p x) (u- (. p y)) (. p z)) xy)",
		"x < y && z > w || !q":               "(|| (&& (< x y) (> z w)) (u! q))",
		"a <= b == c >= d":                   "(== (<= a b) (>= c d))",
		"array<f32, 6>(1.0, 2.0)[i + 1]":     "([] (call array<f32, 6> 1.0 2.0) (+ i 1))",
		"select(a, b, c > 0.0)":              "(call select a b (> c 0.0))",
		"max(dot(edge, edge), 1e-12)":        "(call max (call dot edge edge) 1e-12)",
		"0.00000000000000006123234 * p.x":    "(* 0.00000000000000006123234 (. p x))",
		"f(1u, 2i, 0.5f) + g()":              "(+ (call f 1u 2i 0.5f) (call g))",
		"m[2][1] + s.a.b":                    "(+ ([] ([] m 2) 1) (. (. s a) b))",
		"gsl_tmp_1850.y":                     "(. gsl_tmp_1850 y)",
		"mat4x2<f32>(vec2<f32>(1.0, 2.0)).x": "(. (call mat4x2<f32> (call vec2<f32> 1.0 2.0)) x)",
	} {
		n, err := parseExpr(src)
		if err != nil {
			t.Errorf("%q: %v", src, err)
			continue
		}
		if got := n.sexpr(); got != want {
			t.Errorf("%q:\n got %s\nwant %s", src, got, want)
		}
	}
}

func TestCSE_ParseRejectsBrokenText(t *testing.T) {
	for _, src := range []string{"a +", "f(a, b", "(a", "a b", ""} {
		if _, err := parseExpr(src); err == nil {
			t.Errorf("%q parsed", src)
		}
	}
}

// Every expression the codegen has written into a golden parses, whole: a
// let's, a var's initializer and a return's.
func TestCSE_ParsesEveryGoldenExpression(t *testing.T) {
	stmt := regexp.MustCompile(`^\s*(?:(?:let|var)\s+[A-Za-z_][A-Za-z0-9_]*\s*(?::[^=]+)?=|return\s)\s*(.*);\s*$`)
	var files []string
	for _, pattern := range []string{"../codegen/testdata/*.wgsl", "../codegen/testdata/example/*.wgsl"} {
		m, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatal(err)
		}
		files = append(files, m...)
	}
	parsed := 0
	for _, f := range files {
		src, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for i, line := range strings.Split(string(src), "\n") {
			m := stmt.FindStringSubmatch(line)
			// A line holding more than one statement (a raw material
			// function's `let gsl_m = ...; return ...;`) is not one expression.
			if m == nil || strings.Contains(m[1], ";") {
				continue
			}
			if _, err := parseExpr(m[1]); err != nil {
				t.Errorf("%s:%d: %v\n  %s", filepath.Base(f), i+1, err, m[1])
				continue
			}
			parsed++
		}
	}
	if parsed < 20000 {
		t.Errorf("parsed only %d expressions from %d goldens", parsed, len(files))
	}
}

// The test's long expression: a gear profile's call, as a shaft's fold
// repeats it, long enough that naming it always shortens the function.
const gearCall = "gearProfile(vec3<f32>(p.x * c + p.y * s, p.y * c - p.x * s, p.z), n, m, 0.25)"

func cseFn(stmts ...Stmt) *FuncDecl {
	b := NewBlock()
	for _, s := range stmts {
		b.Add(s)
	}
	return NewFuncDecl("f", []*Param{{Name: "p", Type: "vec3<f32>"}}, "f32", b)
}

func let(name, expr string) Stmt { return NewLetStmt(name, "", expr, nil) }
func ret(expr string) Stmt       { return NewReturnStmt(expr, nil) }
func guard(cond string, body ...Stmt) Stmt {
	b := NewBlock()
	for _, s := range body {
		b.Add(s)
	}
	return NewGuardStmt(cond, b, nil)
}
func loop(body ...Stmt) Stmt {
	b := NewBlock()
	for _, s := range body {
		b.Add(s)
	}
	return NewLoopStmt("i", 3, b)
}

func cseFormat(t *testing.T, fn *FuncDecl) string {
	t.Helper()
	EliminateCommonSubexpressions(fn)
	var b strings.Builder
	if err := fn.Format(&b); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

func wantFn(lines ...string) string {
	return "fn f(p: vec3<f32>) -> f32 {\n" + strings.Join(lines, "\n") + "\n}\n"
}

func checkFn(t *testing.T, got, want string) {
	t.Helper()
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

// Rule 1: a let that repeats an earlier one is dropped and read as it.
func TestCSE_DuplicateLetIsMerged(t *testing.T) {
	got := cseFormat(t, cseFn(
		let("a", "length(p.xy) - r / 2.0"),
		let("b", "length(p.xy) - r / 2.0"),
		ret("max(a, b)")))
	checkFn(t, got, wantFn(
		"  let a = length(p.xy) - r / 2.0;",
		"  return max(a, a);"))
}

// Merging one let makes the next repeat too: a helix's twist and its cosine.
func TestCSE_MergesCascade(t *testing.T) {
	got := cseFormat(t, cseFn(
		let("t1", "(p.z - 34.0) * k"),
		let("c1", "cos(t1)"),
		let("t2", "(p.z - 34.0) * k"),
		let("c2", "cos(t2)"),
		ret("c1 + c2")))
	checkFn(t, got, wantFn(
		"  let t1 = (p.z - 34.0) * k;",
		"  let c1 = cos(t1);",
		"  return c1 + c1;"))
}

// A subexpression a let in scope already names is read as that let.
func TestCSE_SubexpressionReadsItsLet(t *testing.T) {
	got := cseFormat(t, cseFn(
		let("a", gearCall),
		let("b", gearCall+" * 2.0"),
		ret("a + b")))
	checkFn(t, got, wantFn(
		"  let a = "+gearCall+";",
		"  let b = a * 2.0;",
		"  return a + b;"))
}

// Rule 2: the section a fold repeats inside two lets is named once, before
// its first use, and the lets read the name.
func TestCSE_RepeatedSubexpressionIsNamed(t *testing.T) {
	fn := cseFn(
		let("s1", "prism("+gearCall+", 0.0, 0.0)"),
		let("s2", "prism("+gearCall+", 100000.0, 0.0)"),
		ret("min(s1, s2)"))
	got := cseFormat(t, fn)
	checkFn(t, got, wantFn(
		"  let gsl_cse_1 = "+gearCall+";",
		"  let s1 = prism(gsl_cse_1, 0.0, 0.0);",
		"  let s2 = prism(gsl_cse_1, 100000.0, 0.0);",
		"  return min(s1, s2);"))
	// The rewritten lets read the name, so sinking and pruning still see it.
	s1 := fn.Body.Statements()[1].(*LetStmt)
	if !containsString(s1.Reads(), "gsl_cse_1") {
		t.Errorf("s1 reads %v, not gsl_cse_1", s1.Reads())
	}
}

// A short repeat is left: naming it would lengthen the function.
func TestCSE_ShortRepeatIsLeft(t *testing.T) {
	got := cseFormat(t, cseFn(
		let("a", "x / 2.0 + 1.0"),
		let("b", "x / 2.0 - 1.0"),
		ret("a * b")))
	checkFn(t, got, wantFn(
		"  let a = x / 2.0 + 1.0;",
		"  let b = x / 2.0 - 1.0;",
		"  return a * b;"))
}

// A name is never taken twice: a parameter called gsl_cse_1 moves the new
// one on.
func TestCSE_NameAvoidsCollision(t *testing.T) {
	b := NewBlock()
	b.Add(let("s1", "prism("+gearCall+", 0.0)"))
	b.Add(let("s2", "prism("+gearCall+", 1.0)"))
	b.Add(ret("min(s1, s2) + gsl_cse_1"))
	fn := NewFuncDecl("f", []*Param{{Name: "p", Type: "vec3<f32>"}, {Name: "gsl_cse_1", Type: "f32"}}, "f32", b)
	got := cseFormat(t, fn)
	want := "fn f(p: vec3<f32>, gsl_cse_1: f32) -> f32 {\n" +
		"  let gsl_cse_2 = " + gearCall + ";\n" +
		"  let s1 = prism(gsl_cse_2, 0.0);\n" +
		"  let s2 = prism(gsl_cse_2, 1.0);\n" +
		"  return min(s1, s2) + gsl_cse_1;\n}\n"
	checkFn(t, got, want)
}

// The name goes before its first use, never above an opaque early return: a
// box cut's `if (...) { return ...; }` keeps skipping the field.
func TestCSE_NameStaysBelowAnEarlyReturn(t *testing.T) {
	got := cseFormat(t, cseFn(
		let("capped", "length(p) - 1.0"),
		NewOpaqueStmt([]string{"if (capped > 1.0) { return capped; }"}, []string{"capped"}, nil),
		let("s1", "prism("+gearCall+", 0.0)"),
		let("s2", "prism("+gearCall+", 1.0)"),
		ret("min(s1, s2)")))
	checkFn(t, got, wantFn(
		"  let capped = length(p) - 1.0;",
		"  if (capped > 1.0) { return capped; }",
		"  let gsl_cse_1 = "+gearCall+";",
		"  let s1 = prism(gsl_cse_1, 0.0);",
		"  let s2 = prism(gsl_cse_1, 1.0);",
		"  return min(s1, s2);"))
}

// Two guards that each compute it keep their own: naming it above them would
// compute what neither guard may need.
func TestCSE_SiblingGuardsKeepTheirOwn(t *testing.T) {
	got := cseFormat(t, cseFn(
		NewVarStmt("v", "", "0.0", nil),
		guard("w > 0.0", let("u", gearCall), NewAssignStmt("v", "u", nil)),
		guard("w < 0.0", let("u2", gearCall), NewAssignStmt("v", "u2", nil)),
		ret("v")))
	checkFn(t, got, wantFn(
		"  var v = 0.0;",
		"  if (w > 0.0) {",
		"    let u = "+gearCall+";",
		"    v = u;",
		"  }",
		"  if (w < 0.0) {",
		"    let u2 = "+gearCall+";",
		"    v = u2;",
		"  }",
		"  return v;"))
}

// Inside a guard, what a let above it names is read as that let.
func TestCSE_GuardReadsTheLetAboveIt(t *testing.T) {
	got := cseFormat(t, cseFn(
		let("big", gearCall),
		NewVarStmt("v", "", "big", nil),
		guard("w > 0.0", NewAssignStmt("v", gearCall+" * 2.0", nil)),
		ret("v")))
	checkFn(t, got, wantFn(
		"  let big = "+gearCall+";",
		"  var v = big;",
		"  if (w > 0.0) {",
		"    v = big * 2.0;",
		"  }",
		"  return v;"))
}

// Nothing leaves a loop's body for the block around it.
func TestCSE_LoopBodyKeepsItsOwn(t *testing.T) {
	got := cseFormat(t, cseFn(
		let("a", gearCall+" + 1.0"),
		NewVarStmt("acc", "", "0.0", nil),
		loop(NewAssignStmt("acc", "acc + "+gearCall, nil)),
		ret("a + acc")))
	checkFn(t, got, wantFn(
		"  let a = "+gearCall+" + 1.0;",
		"  var acc = 0.0;",
		"  for (var i = 0; i < 3; i = i + 1) {",
		"    acc = acc + "+gearCall+";",
		"  }",
		"  return a + acc;"))
}

// Within one loop body the rule holds as in any block.
func TestCSE_LoopBodyNamesItsOwnRepeat(t *testing.T) {
	got := cseFormat(t, cseFn(
		NewVarStmt("acc", "", "0.0", nil),
		loop(NewAssignStmt("acc", "acc + "+gearCall, nil),
			NewAssignStmt("acc", "acc * "+gearCall, nil)),
		ret("acc")))
	checkFn(t, got, wantFn(
		"  var acc = 0.0;",
		"  for (var i = 0; i < 3; i = i + 1) {",
		"    let gsl_cse_1 = "+gearCall+";",
		"    acc = acc + gsl_cse_1;",
		"    acc = acc * gsl_cse_1;",
		"  }",
		"  return acc;"))
}

// What reads a var is never merged: its value can change between the two.
func TestCSE_VarReadsAreNotMerged(t *testing.T) {
	call := "gearProfile(vec3<f32>(v * c + p.y * s, p.y * c - v * s, p.z), n, m, 0.25)"
	got := cseFormat(t, cseFn(
		NewVarStmt("v", "", "1.0", nil),
		let("a", call),
		NewAssignStmt("v", "2.0", nil),
		let("b", call),
		ret("a + b")))
	checkFn(t, got, wantFn(
		"  var v = 1.0;",
		"  let a = "+call+";",
		"  v = 2.0;",
		"  let b = "+call+";",
		"  return a + b;"))
}

// Nor what reads a name an opaque statement writes.
func TestCSE_OpaqueWritesAreNotMerged(t *testing.T) {
	call := "gearProfile(vec3<f32>(q * c + p.y * s, p.y * c - q * s, p.z), n, m, 0.25)"
	got := cseFormat(t, cseFn(
		let("a", call),
		NewOpaqueStmt([]string{"q = q + 1.0;"}, []string{"q"}, []string{"q"}),
		let("b", call),
		ret("a + b")))
	checkFn(t, got, wantFn(
		"  let a = "+call+";",
		"  q = q + 1.0;",
		"  let b = "+call+";",
		"  return a + b;"))
}

// The right of && or || runs only when the left allows it, so nothing is
// named from there.
func TestCSE_ShortCircuitRightIsNotHoisted(t *testing.T) {
	got := cseFormat(t, cseFn(
		let("a", "x > 0.0 && "+gearCall+" > 0.0"),
		let("b", "y > 0.0 || "+gearCall+" > 1.0"),
		ret("select(0.0, 1.0, a || b)")))
	checkFn(t, got, wantFn(
		"  let a = x > 0.0 && "+gearCall+" > 0.0;",
		"  let b = y > 0.0 || "+gearCall+" > 1.0;",
		"  return select(0.0, 1.0, a || b);"))
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
