package lib

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/program"
)

// TestSelectOnConstantArgumentIsPruned: a function that takes fields and a
// constant choosing between two constructions is not inlined (its fields read
// p), but each call is specialized on the constant (constspec), so the fold
// prunes the construction the constant does not choose and no select ships.
func TestSelectOnConstantArgumentIsPruned(t *testing.T) {
	src := `
roundHead = 1.0;
flatHead = 2.0;
capped(head, body, w) {
  cap = head == flatHead ? box([0.0 - w, 0.0 - w, -0.5], [w, w, 0.5]) : sphere(w);
  s = max(body, neg(cap));
  return s, cap;
}
a, capA = capped(flatHead, length(p) - 3.0, 1.0);
b, capB = capped(roundHead, length(p - [5.0, 0.0, 0.0]) - 3.0, 1.5);
return min(a, b), min(capA, capB);
`
	res, err := Assemble([]string{src}, source.NewLocalManager("../../../common/source/package"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n := countSelects(res.Program); n != 0 {
		t.Errorf("%d selects survive assembly; each call's construction should be chosen at build time", n)
	}
	var clones []string
	for _, f := range res.Program.Primary.Files {
		for _, ft := range f.FnTypes {
			if strings.HasPrefix(ft.Name, "capped__spec") {
				clones = append(clones, ft.Name)
			}
		}
	}
	if len(clones) > 2 {
		t.Errorf("clones %v: one per head at most", clones)
	}
}

// countSelects is the number of select expressions anywhere in prog.
func countSelects(prog *program.Program) int {
	n := 0
	var op func(*program.Operand)
	op = func(o *program.Operand) {
		if o == nil {
			return
		}
		switch t := o.T.(type) {
		case *program.Operand_Expression:
			if t.Expression.Operation == program.Operation_SELECT {
				n++
			}
			for _, sub := range t.Expression.Operands {
				op(sub)
			}
		case *program.Operand_FnCall:
			for _, a := range t.FnCall.Arguments {
				op(a)
			}
		}
	}
	var fn func(*program.Function)
	fn = func(f *program.Function) {
		for _, a := range f.Assignments {
			op(a.Expression)
		}
		for _, r := range f.Returns {
			op(r)
		}
		for _, ft := range f.FnTypes {
			for _, g := range ft.Groups {
				for _, inner := range g.Functions {
					fn(inner)
				}
			}
		}
	}
	for _, pkg := range append([]*program.Package{prog.Primary}, prog.Dependencies...) {
		for _, f := range pkg.Files {
			for _, a := range f.Assignments {
				op(a.Expression)
			}
			for _, r := range f.Returns {
				op(r)
			}
			for _, ft := range f.FnTypes {
				for _, g := range ft.Groups {
					for _, inner := range g.Functions {
						fn(inner)
					}
				}
			}
		}
	}
	return n
}

// TestSelectOnVaryingConditionIsNotSpecialized: a condition that still varies
// once the constants are bound — here it reads pixelSize(), which changes at
// render time without being an argument — prunes nothing, so the call keeps
// the function as written rather than gaining a clone.
func TestSelectOnVaryingConditionIsNotSpecialized(t *testing.T) {
	src := `
banded(x, period) {
  width = pixelSize();
  exact = fract(x / period) < 0.5 ? 0 : 1;
  return width > 0.0001 * period ? 0.5 : exact;
}
return length(p) - 3.0 - banded(p.x, 20.0);
`
	res, err := Assemble([]string{src}, source.NewLocalManager("../../../common/source/package"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, f := range res.Program.Primary.Files {
		for _, ft := range f.FnTypes {
			if strings.HasPrefix(ft.Name, "banded__spec") {
				t.Errorf("%s was cloned, but no select in banded becomes constant", ft.Name)
			}
		}
	}
}
