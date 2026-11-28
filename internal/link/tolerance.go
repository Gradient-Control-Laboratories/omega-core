package link

import "github.com/Gradient-Control-Laboratories/omega-core/program"

// propagatedTolerance is what y.tolerance reads when y declares none: its
// tolerance to first order over the working tolerances the package's top-level
// declarations carry (x.tolerance = tolerance.limits(lo, hi)), as offsets of
// y's value, [bias - spread, bias + spread], where
//
//	bias   = sum of dy/dx * (lo + hi) / 2
//	spread = sum of |dy/dx| * (hi - lo) / 2
//
// over each toleranced x. dy/dx is a DERIV, expanded as every derivative is:
// for a shape it varies over the part, face by face, so a shape's tolerance is
// a field, and in a query it is taken of the field the condition reads.
func propagatedTolerance(target *program.Reference, pkg *program.Package, start, end *program.Location) *program.Operand {
	var bias, spread []*program.Operand
	for _, name := range tolerancedNames(pkg) {
		if target.Prefix == "" && name == target.Name {
			continue
		}
		deriv := func() *program.Operand {
			return program.NewExpression(program.Operation_DERIV, []*program.Operand{
				program.NewReference(target.Prefix, target.Name, start, end),
				program.NewReference("", name, start, end),
			}, start, end)
		}
		limit := func(i float64) *program.Operand {
			return program.NewExpression(program.Operation_INDEX, []*program.Operand{
				program.NewPropertyRef(name, "tolerance", start, end),
				program.NewValue(i, start, end),
			}, start, end)
		}
		half := func(o *program.Operand) *program.Operand {
			return program.NewExpression(program.Operation_DIV, []*program.Operand{o, program.NewValue(2, start, end)}, start, end)
		}
		mid := half(program.NewExpression(program.Operation_ADD, []*program.Operand{limit(0), limit(1)}, start, end))
		width := half(program.NewExpression(program.Operation_SUB, []*program.Operand{limit(1), limit(0)}, start, end))
		bias = append(bias, program.NewExpression(program.Operation_MUL, []*program.Operand{deriv(), mid}, start, end))
		spread = append(spread, program.NewExpression(program.Operation_MUL, []*program.Operand{
			program.NewExpression(program.Operation_ABS, []*program.Operand{deriv()}, start, end), width,
		}, start, end))
	}
	if len(bias) == 0 {
		return program.NewExpression(program.Operation_COLLECT, []*program.Operand{
			program.NewValue(0, start, end), program.NewValue(0, start, end),
		}, start, end)
	}
	sum := func(terms []*program.Operand) *program.Operand {
		total := terms[0]
		for _, t := range terms[1:] {
			total = program.NewExpression(program.Operation_ADD, []*program.Operand{total, t}, start, end)
		}
		return total
	}
	b, s := sum(bias), sum(spread)
	return program.NewExpression(program.Operation_COLLECT, []*program.Operand{
		program.NewExpression(program.Operation_SUB, []*program.Operand{b, s}, start, end),
		program.NewExpression(program.Operation_ADD, []*program.Operand{program.CloneOperand(b), program.CloneOperand(s)}, start, end),
	}, start, end)
}

// tolerancedNames lists pkg's top-level declarations that carry a working
// tolerance, in order.
func tolerancedNames(pkg *program.Package) []string {
	if pkg == nil {
		return nil
	}
	var names []string
	seen := map[string]bool{}
	for _, f := range pkg.Files {
		for _, a := range f.Assignments {
			if len(a.Variables) != 1 || a.Variables[0] == nil {
				continue
			}
			v := a.Variables[0]
			for _, p := range v.Properties {
				if p != nil && p.Name == "tolerance" && !seen[v.Name] {
					seen[v.Name] = true
					names = append(names, v.Name)
				}
			}
		}
	}
	return names
}
