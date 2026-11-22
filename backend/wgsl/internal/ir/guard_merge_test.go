package ir

import (
	"strings"
	"testing"
)

// Shaft Modeling, Phase 21: guards that share a condition merge. A form split
// at its pitch diameter (a thread, a gear) is two features of one size, the
// crests and the roots, and CSE gives both guards the one weight; each body
// evaluates the whole form, and since CSE never hoists out of a guard, a
// visible thread computed its form twice. MergeGuards moves a later guard
// whose condition is the same text into the earlier one, hoisting what lies
// between above it where that is independent, and leaves both where it is
// not; CSE then names the form once inside the one branch.

func mergeFormat(t *testing.T, fn *FuncDecl, cse bool) string {
	t.Helper()
	MergeGuards(fn)
	if cse {
		EliminateCommonSubexpressions(fn)
	}
	var b strings.Builder
	if err := fn.Format(&b); err != nil {
		t.Fatal(err)
	}
	return b.String()
}

func assign(target, expr string) Stmt { return NewAssignStmt(target, expr, nil) }
func decl(name, init string) Stmt     { return NewVarStmt(name, "", init, nil) }

// Two guards on one weight merge into the first; what lies between and does
// not read the first guard's writes moves above it.
func TestMergeGuards_SameConditionMerges(t *testing.T) {
	got := mergeFormat(t, cseFn(
		let("w", "weight(0.5)"),
		decl("a", "100000.0"),
		guard("w > 0.0", assign("a", gearCall+" + 1.0")),
		let("t", "p.x * 2.0"),
		decl("b", "100000.0"),
		guard("w > 0.0", assign("b", "-"+gearCall+" + t")),
		ret("max(min(t, a), -b)")), false)
	checkFn(t, got, wantFn(
		"  let w = weight(0.5);",
		"  var a = 100000.0;",
		"  let t = p.x * 2.0;",
		"  var b = 100000.0;",
		"  if (w > 0.0) {",
		"    a = "+gearCall+" + 1.0;",
		"    b = -"+gearCall+" + t;",
		"  }",
		"  return max(min(t, a), -b);"))
}

// Merged, CSE names the form once inside the one branch.
func TestMergeGuards_FormEvaluatedOnce(t *testing.T) {
	got := mergeFormat(t, cseFn(
		let("w", "weight(0.5)"),
		decl("a", "100000.0"),
		guard("w > 0.0", assign("a", gearCall+" + 1.0")),
		decl("b", "100000.0"),
		guard("w > 0.0", assign("b", "-"+gearCall+" + 2.0")),
		ret("max(a, -b)")), true)
	if n := strings.Count(got, "if ("); n != 1 {
		t.Errorf("%d guards, want 1:\n%s", n, got)
	}
	if n := strings.Count(got, gearCall); n != 1 {
		t.Errorf("the form is written %d times, want once:\n%s", n, got)
	}
}

// Three guards on one weight become one.
func TestMergeGuards_ThreeBecomeOne(t *testing.T) {
	got := mergeFormat(t, cseFn(
		let("w", "weight(0.5)"),
		decl("a", "0.0"), guard("w > 0.0", assign("a", "f(p)")),
		decl("b", "0.0"), guard("w > 0.0", assign("b", "g(p)")),
		decl("c", "0.0"), guard("w > 0.0", assign("c", "h(p)")),
		ret("a + b + c")), false)
	checkFn(t, got, wantFn(
		"  let w = weight(0.5);",
		"  var a = 0.0;",
		"  var b = 0.0;",
		"  var c = 0.0;",
		"  if (w > 0.0) {",
		"    a = f(p);",
		"    b = g(p);",
		"    c = h(p);",
		"  }",
		"  return a + b + c;"))
}

// A statement between that reads the first guard's write stays after it, and
// the second guard still merges when it does not read that statement.
func TestMergeGuards_DependentStatementStaysBehind(t *testing.T) {
	got := mergeFormat(t, cseFn(
		let("w", "weight(0.5)"),
		decl("a", "0.0"),
		guard("w > 0.0", assign("a", "f(p)")),
		let("c", "a + 1.0"),
		decl("b", "0.0"),
		guard("w > 0.0", assign("b", "g(p)")),
		ret("c + b")), false)
	checkFn(t, got, wantFn(
		"  let w = weight(0.5);",
		"  var a = 0.0;",
		"  var b = 0.0;",
		"  if (w > 0.0) {",
		"    a = f(p);",
		"    b = g(p);",
		"  }",
		"  let c = a + 1.0;",
		"  return c + b;"))
}

// The second guard reads a statement that reads the first guard's write: it
// cannot move above that statement, so nothing merges.
func TestMergeGuards_DependentBodyDoesNotMerge(t *testing.T) {
	stmts := []Stmt{
		let("w", "weight(0.5)"),
		decl("a", "0.0"),
		guard("w > 0.0", assign("a", "f(p)")),
		let("c", "a + 1.0"),
		decl("b", "0.0"),
		guard("w > 0.0", assign("b", "c * 2.0")),
		ret("c + b"),
	}
	want := mergeFormat(t, cseFn(stmts...), false)
	checkFn(t, want, wantFn(
		"  let w = weight(0.5);",
		"  var a = 0.0;",
		"  if (w > 0.0) {",
		"    a = f(p);",
		"  }",
		"  let c = a + 1.0;",
		"  var b = 0.0;",
		"  if (w > 0.0) {",
		"    b = c * 2.0;",
		"  }",
		"  return c + b;"))
}

// A statement between that reads what the second guard writes would see the
// new value once the guard moved above it: nothing merges.
func TestMergeGuards_ReaderOfSecondWriteBlocks(t *testing.T) {
	got := mergeFormat(t, cseFn(
		let("w", "weight(0.5)"),
		decl("a", "0.0"),
		decl("b", "1.0"),
		guard("w > 0.0", assign("a", "f(p)")),
		let("c", "b + a"),
		guard("w > 0.0", assign("b", "g(p)")),
		ret("c + b")), false)
	if n := strings.Count(got, "if ("); n != 2 {
		t.Errorf("%d guards, want both kept:\n%s", n, got)
	}
}

// Different conditions, the stand-in switch's w > 0 and w < 1, never merge;
// nor do guards whose condition is written between them.
func TestMergeGuards_DifferentOrRewrittenConditionKept(t *testing.T) {
	for name, fn := range map[string]*FuncDecl{
		"different": cseFn(
			let("w", "weight(0.5)"),
			decl("a", "0.0"), guard("w > 0.0", assign("a", "f(p)")),
			decl("b", "0.0"), guard("w < 1.0", assign("b", "g(p)")),
			ret("a + b")),
		"rewritten": cseFn(
			decl("w", "weight(0.5)"),
			decl("a", "0.0"), guard("w > 0.0", assign("a", "f(p)")),
			assign("w", "weight(2.0)"),
			decl("b", "0.0"), guard("w > 0.0", assign("b", "g(p)")),
			ret("a + b")),
	} {
		if n := strings.Count(mergeFormat(t, fn, false), "if ("); n != 2 {
			t.Errorf("%s: %d guards, want both kept", name, n)
		}
	}
}

// An inline guard, `if (c) { stmt; }` on one line, is left as it is.
func TestMergeGuards_InlineGuardsKept(t *testing.T) {
	inline := func(target, expr string) Stmt {
		b := NewBlock()
		b.Add(assign(target, expr))
		return NewInlineGuardStmt("w > 0.0", b, nil)
	}
	got := mergeFormat(t, cseFn(
		let("w", "weight(0.5)"),
		decl("a", "0.0"), inline("a", "f(p)"),
		decl("b", "0.0"), inline("b", "g(p)"),
		ret("a + b")), false)
	if n := strings.Count(got, "if ("); n != 2 {
		t.Errorf("%d guards, want the two inline ones kept:\n%s", n, got)
	}
}

// Guards nested in a guard's body merge there too.
func TestMergeGuards_NestedBodies(t *testing.T) {
	got := mergeFormat(t, cseFn(
		let("w", "weight(0.5)"),
		decl("a", "0.0"),
		decl("b", "0.0"),
		guard("near",
			guard("w > 0.0", assign("a", "f(p)")),
			guard("w > 0.0", assign("b", "g(p)"))),
		ret("a + b")), false)
	if n := strings.Count(got, "if (w > 0.0)"); n != 1 {
		t.Errorf("%d inner guards, want 1:\n%s", n, got)
	}
}
