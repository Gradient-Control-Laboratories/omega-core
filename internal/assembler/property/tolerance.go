package property

import "github.com/Gradient-Control-Laboratories/omega-core/program"

// lowerToleranceReturns sets each geometry return's working tolerance
// (Program.ToleranceReturns): the vec2 [lo, hi] of offsets its distance takes
// to first order over the primary package's toleranced declarations, as a
// value that declares none reads its .tolerance (link's propagatedTolerance):
//
//	bias   = sum of d(ret)/dx * (lo + hi) / 2
//	spread = sum of |d(ret)/dx| * (hi - lo) / 2
//	[lo, hi] = [bias - spread, bias + spread]
//
// It does nothing for a document that declares no tolerance, so every other
// document compiles as before. The derivatives are taken of each return as
// built.
func lowerToleranceReturns(prog *program.Program) error {
	pkg := prog.Primary
	last := pkg.Files[len(pkg.Files)-1]
	limits := tolerancedLimits(pkg)
	if len(limits) == 0 || len(last.Returns) == 0 {
		return nil
	}
	tols := make([]*program.Operand, len(last.Returns))
	for i, ret := range last.Returns {
		if ret == nil {
			continue
		}
		var bias, spread []*program.Operand
		for _, l := range limits {
			deriv := func() *program.Operand {
				return nary(program.Operation_DERIV, program.CloneOperand(ret), ref(l.name))
			}
			limit := func(k float64) *program.Operand {
				return nary(program.Operation_INDEX, program.CloneOperand(l.expr), num(k))
			}
			mid := div(add(limit(0), limit(1)), num(2))
			half := div(sub(limit(1), limit(0)), num(2))
			bias = append(bias, mul(deriv(), mid))
			spread = append(spread, mul(un(program.Operation_ABS, deriv()), half))
		}
		b, s := add(bias...), add(spread...)
		tols[i] = nary(program.Operation_COLLECT, sub(b, s), add(program.CloneOperand(b), program.CloneOperand(s)))
	}
	if err := expandReturnDerivs(prog, tols); err != nil {
		return err
	}
	last.ToleranceReturns = tols
	return nil
}

// tolerated is a toleranced declaration: its name and its limits' expression.
type tolerated struct {
	name string
	expr *program.Operand
}

// tolerancedLimits lists pkg's top-level declarations that carry a working
// tolerance, with the expression each assigns it, in order.
func tolerancedLimits(pkg *program.Package) []tolerated {
	var out []tolerated
	seen := map[string]bool{}
	for _, f := range pkg.Files {
		for _, a := range f.Assignments {
			if len(a.Variables) != 1 || a.Variables[0] == nil {
				continue
			}
			v := a.Variables[0]
			for _, p := range v.Properties {
				if p != nil && p.Name == "tolerance" && p.Expression != nil && !seen[v.Name] {
					seen[v.Name] = true
					out = append(out, tolerated{name: v.Name, expr: p.Expression})
				}
			}
		}
	}
	return out
}
