package program

// IsSynthetic returns true if the RangeAnnotation was synthesized in the parser
// from v1 .min/.max properties rather than written explicitly in the source.
func (r *Range) IsSynthetic() bool {
	return r.Start.Line == r.End.Line && r.Start.Column == r.End.Column
}

func operandConst(op *Operand) (float64, bool) {
	if len(op.ConstValues) == 1 {
		return op.ConstValues[0], true
	}
	if valOp, ok := op.T.(*Operand_Value); ok {
		return valOp.Value, true
	}
	return 0, false
}

// MinConst returns the evaluated scalar constant value of r.Min, if resolved.
func (r *Range) MinConst() (float64, bool) {
	return operandConst(r.Min)
}

// MaxConst returns the evaluated scalar constant value of r.Max, if resolved.
func (r *Range) MaxConst() (float64, bool) {
	return operandConst(r.Max)
}

// PeriodConst returns the evaluated scalar constant value of r.Period, if resolved.
func (r *Range) PeriodConst() (float64, bool) {
	if r.Period == nil {
		return 0, false
	}
	return operandConst(r.Period)
}

// ComputeSliderVars traverses top-level assignments in pkg and returns a set
// containing all slider variables (having Range != nil) and all top-level variables
// whose expressions transitively depend on them.
func ComputeSliderVars(pkg *Package) map[string]bool {
	sliderVars := make(map[string]bool)
	if pkg == nil {
		return sliderVars
	}

	var assignments []*Assignment
	for _, f := range pkg.Files {
		assignments = append(assignments, f.Assignments...)
	}

	for _, a := range assignments {
		if a.Range == nil {
			continue
		}
		for _, v := range a.Variables {
			sliderVars[v.Name] = true
		}
	}

	if len(sliderVars) == 0 {
		return sliderVars
	}

	changed := true
	for changed {
		changed = false
		for _, a := range assignments {
			allMarked := true
			for _, v := range a.Variables {
				if !sliderVars[v.Name] {
					allMarked = false
					break
				}
			}
			if allMarked {
				continue
			}

			if expressionReferencesSlider(a.Expression, sliderVars) {
				for _, v := range a.Variables {
					if !sliderVars[v.Name] {
						sliderVars[v.Name] = true
						changed = true
					}
				}
			}
		}
	}

	return sliderVars
}

func expressionReferencesSlider(op *Operand, sliderVars map[string]bool) bool {
	v := &sliderRefVisitor{
		sliderVars: sliderVars,
		visited:    make(map[*Operand]bool),
	}
	res, _ := VisitOperand(op, v)
	return res
}

type sliderRefVisitor struct {
	sliderVars map[string]bool
	visited    map[*Operand]bool
}

func (v *sliderRefVisitor) VisitValue(*Operand, float64) (bool, error) {
	return false, nil
}

func (v *sliderRefVisitor) VisitReference(_ *Operand, ref *Reference) (bool, error) {
	if ref.PkgPath == "" || ref.PkgPath == "builtin" {
		if v.sliderVars[ref.Name] {
			return true, nil
		}
	}
	return false, nil
}

func (v *sliderRefVisitor) VisitExpression(op *Operand, expr *Expression) (bool, error) {
	if v.visited[op] {
		return false, nil
	}
	v.visited[op] = true
	for _, child := range expr.Operands {
		found, err := VisitOperand(child, v)
		if err != nil || found {
			return true, err
		}
	}
	return false, nil
}

func (v *sliderRefVisitor) VisitFnCall(op *Operand, fc *FnCall) (bool, error) {
	if v.visited[op] {
		return false, nil
	}
	v.visited[op] = true
	for _, arg := range fc.Arguments {
		found, err := VisitOperand(arg, v)
		if err != nil || found {
			return true, err
		}
	}
	return false, nil
}

func (v *sliderRefVisitor) VisitPropRef(*Operand, *PropertyRef) (bool, error) {
	return false, nil
}
