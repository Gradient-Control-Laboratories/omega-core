package lower_test

import (
	"testing"

	"gradientcontrol.com/omega/core/internal/closure"
	"gradientcontrol.com/omega/core/internal/collect"
	"gradientcontrol.com/omega/core/internal/link"
	"gradientcontrol.com/omega/core/internal/lower"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/types"
	parser "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

// keptReduces lowers src with LoopFolds and counts the reduces left in its
// functions.
func buildTestProgram(t *testing.T, src string, lowerStep func(prog *program.Program) error) *program.Program {
	t.Helper()
	srcMgr := &source.MockManager{}
	prog, err := collect.Program([]string{src}, srcMgr, parser.ParsePackage)
	if err != nil {
		t.Fatal(err)
	}
	for _, step := range []func() error{
		func() error { return link.References(prog, srcMgr, parser.ParsePackage) },
		func() error { return closure.Transform(prog) },
		func() error { return link.References(prog, srcMgr, parser.ParsePackage) },
		func() error { return types.Analyze(prog) },
		func() error { return lowerStep(prog) },
	} {
		if err := step(); err != nil {
			t.Fatal(err)
		}
	}
	return prog
}

func countReduces(prog *program.Program) int {
	count := 0
	var visit func(op *program.Operand)
	visit = func(op *program.Operand) {
		if op == nil {
			return
		}
		if e := op.GetExpression(); e != nil {
			if e.Operation == program.Operation_REDUCE {
				count++
			}
			for _, sub := range e.Operands {
				visit(sub)
			}
		}
		if fc := op.GetFnCall(); fc != nil {
			for _, arg := range fc.Arguments {
				visit(arg)
			}
		}
	}
	var walk func(fts []*program.FnType)
	walk = func(fts []*program.FnType) {
		for _, ft := range fts {
			for _, g := range ft.Groups {
				for _, fn := range g.Functions {
					for _, a := range fn.Assignments {
						visit(a.Expression)
					}
					for _, r := range fn.Returns {
						visit(r)
					}
					walk(fn.FnTypes)
				}
			}
		}
	}
	for _, f := range prog.Primary.Files {
		walk(f.FnTypes)
	}
	return count
}

func keptReduces(t *testing.T, src string) int {
	t.Helper()
	prog := buildTestProgram(t, src, func(p *program.Program) error {
		return lower.TransformWith(p, lower.Options{LoopFolds: lower.LoopFoldThreshold})
	})
	return countReduces(prog)
}

func TestLoopFolds(t *testing.T) {
	cases := []struct {
		name string
		src  string
		kept int
	}{
		{"constants with an initial value", `
scan(q) {
  return reduce([1000, 0], (best, i) {
    d = length(q - [i, 0, 0]);
    return d < best.x ? [d, i] : best;
  }, [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]);
}
return sphere(p, 10 + scan(p).x);`, 1},
		{"vectors from a named list", `
scan(q) {
  offsets = [[0, 0, 0], [1, 0, 0], [0, 1, 0], [1, 1, 0], [0, 0, 1], [1, 0, 1], [0, 1, 1], [1, 1, 1]];
  return reduce(0.0, (acc, o) { return acc + length(q - o); }, offsets);
}
return sphere(p, scan(p));`, 1},
		{"fewer than the threshold", `
scan(q) {
  return reduce(0.0, (acc, i) { return acc + length(q - [i, 0, 0]); }, [0, 1, 2, 3, 4, 5, 6]);
}
return sphere(p, scan(p));`, 0},
		{"no initial value", `
scan(q) {
  return reduce((acc, i) { return acc + i * q.x; }, [0, 1, 2, 3, 4, 5, 6, 7, 8]);
}
return sphere(p, scan(p));`, 0},
		{"outside a function", `
s = reduce(0.0, (acc, i) { return acc + i * p.x; }, [0, 1, 2, 3, 4, 5, 6, 7, 8]);
return sphere(p, s);`, 0},
		{"a list indexed by the element", `
scan(q) {
  ks = [1, 2, 3, 4, 5, 6, 7, 8, 9];
  return reduce(0.0, (acc, i) { return acc + ks[i] * q.x; }, [0, 1, 2, 3, 4, 5, 6, 7, 8]);
}
return sphere(p, scan(p));`, 0},
		{"a list indexed through a called function", `
pick(i) {
  ks = [1, 2, 3, 4, 5, 6, 7, 8, 9];
  return ks[i];
}
scan(q) {
  return reduce(0.0, (acc, i) { return acc + pick(i) * q.x; }, [0, 1, 2, 3, 4, 5, 6, 7, 8]);
}
return sphere(p, scan(p));`, 0},
		{"a list indexed by a constant", `
scan(q) {
  ks = [1, 2, 3, 4, 5, 6, 7, 8, 9];
  return reduce(0.0, (acc, i) { return acc + ks[2] * i * q.x; }, [0, 1, 2, 3, 4, 5, 6, 7, 8]);
}
return sphere(p, scan(p));`, 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := keptReduces(t, c.src); got != c.kept {
				t.Errorf("kept %d reduces, want %d", got, c.kept)
			}
		})
	}
}

func TestUniversalTransformKeepsReduces(t *testing.T) {
	src := `
scan(q) {
  return reduce([1000, 0], (best, i) {
    d = length(q - [i, 0, 0]);
    return d < best.x ? [d, i] : best;
  }, [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]);
}
return sphere(p, 10 + scan(p).x);`

	prog := buildTestProgram(t, src, lower.Transform)
	if got := countReduces(prog); got != 1 {
		t.Errorf("lower.Transform kept %d reduces, want 1", got)
	}
}

func TestUnrollReduce(t *testing.T) {
	src := `
scan(q) {
  return reduce([1000, 0], (best, i) {
    d = length(q - [i, 0, 0]);
    return d < best.x ? [d, i] : best;
  }, [0, 1, 2, 3, 4, 5, 6, 7, 8, 9]);
}
return sphere(p, 10 + scan(p).x);`

	prog := buildTestProgram(t, src, func(p *program.Program) error {
		return lower.TransformWith(p, lower.Options{LoopFolds: lower.LoopFoldThreshold})
	})
	if got := countReduces(prog); got != 1 {
		t.Fatalf("expected 1 reduce before UnrollReduce, got %d", got)
	}

	if err := lower.UnrollReduce(prog); err != nil {
		t.Fatalf("UnrollReduce failed: %v", err)
	}

	if got := countReduces(prog); got != 0 {
		t.Errorf("after UnrollReduce, remaining reduces = %d, want 0", got)
	}
}
