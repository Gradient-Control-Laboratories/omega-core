package eval

import (
	"errors"
	"fmt"

	"gradientcontrol.com/omega/core/internal/scope"
	"gradientcontrol.com/omega/core/program"
)

// ParameterChecker identifies dynamic parameter operands to prevent folding them.
type ParameterChecker interface {
	IsParameter(op *program.Operand) bool
}

// ErrParameter is returned by constant evaluation when the value depends on a
// literal the ParameterChecker has lifted to a runtime parameter.
var ErrParameter = errors.New("operand depends on a runtime parameter")

// Fold runs a constant folding pass over the entire program, evaluating all constant
// expressions and assignments and populating their ConstValues fields. Under a
// checker it also collapses each scalar constant it has evaluated into its literal.
// It returns true if any modifications (such as select pruning, constant folding
// or collapsing) were made.
func Fold(prog *program.Program, checker ParameterChecker) (bool, error) {
	progScope, err := scope.Init(prog)
	if err != nil {
		return false, err
	}

	getImport := func(path string) (*program.Package, error) {
		if prog.Primary != nil && prog.Primary.Path == path {
			return prog.Primary, nil
		}
		for _, dep := range prog.Dependencies {
			if dep.Path == path {
				return dep, nil
			}
		}
		return nil, fmt.Errorf("package %q not found in program dependencies", path)
	}

	// Helper to find the package scope
	findPkgScope := func(pkg *program.Package) *scope.PkgScope {
		if pkg == prog.Primary {
			return progScope.Primary
		}
		return progScope.Dependencies[pkg.Path]
	}

	modifiedAny := false

	// Helper to evaluate a package
	foldPkg := func(pkg *program.Package) error {
		if pkg == nil {
			return nil
		}
		pkgScope := findPkgScope(pkg)
		visitor := &foldVisitor{
			progScope: progScope,
			pkgScope:  pkgScope,
			pkg:       pkg,
			getImport: getImport,
			checker:   checker,
			cache:     make(map[*program.Operand]value),
		}

		for _, file := range pkg.Files {
			// Evaluate assignments in the file
			for _, assign := range file.Assignments {
				visitor.keepRoot = nil
				if assign.Range != nil {
					visitor.keepRoot = assign.Expression
				}
				if _, err := program.VisitOperand(assign.Expression, visitor); err != nil {
					return err
				}
				if len(assign.Variables) == 1 {
					if len(assign.Expression.ConstValues) > 0 {
						if !sliceEqual(assign.Variables[0].ConstValues, assign.Expression.ConstValues) {
							visitor.modified = true
						}
						assign.Variables[0].ConstValues = assign.Expression.ConstValues
					} else {
						if len(assign.Variables[0].ConstValues) > 0 {
							visitor.modified = true
						}
						assign.Variables[0].ConstValues = nil
					}
				} else if len(assign.Variables) > 1 {
					ctx := &constCtx{
						progScope:  progScope,
						cur:        pkgScope,
						getImport:  getImport,
						evaluating: make(map[evalKey]bool),
						calls:      make(map[callKey]bool),
						checker:    checker,
					}
					vVal, err := ctx.operand(assign.Expression)
					if err == nil {
						elements := vVal.elements()
						if len(elements) == len(assign.Variables) {
							for i, v := range assign.Variables {
								if !sliceEqual(v.ConstValues, elements[i].floats()) {
									visitor.modified = true
								}
								v.ConstValues = elements[i].floats()
							}
						}
					} else if errors.Is(err, ErrParameter) {
						for _, v := range assign.Variables {
							if len(v.ConstValues) > 0 {
								visitor.modified = true
							}
							v.ConstValues = nil
						}
					}
				}
				// Also evaluate property assignments if any
				for _, v := range assign.Variables {
					for _, prop := range v.Properties {
						if _, err := program.VisitOperand(prop.Expression, visitor); err != nil {
							return err
						}
					}
				}
			}

			// Evaluate functions in the file
			for _, fnType := range file.FnTypes {
				if err := visitor.visitFnType(fnType); err != nil {
					return err
				}
			}

			// Evaluate returns in the file
			for _, ret := range file.Returns {
				if _, err := program.VisitOperand(ret, visitor); err != nil {
					return err
				}
			}
			for _, ret := range file.ColorReturns {
				if _, err := program.VisitOperand(ret, visitor); err != nil {
					return err
				}
			}
			for _, ret := range file.MaterialReturns {
				if _, err := program.VisitOperand(ret, visitor); err != nil {
					return err
				}
			}
		}
		if visitor.modified {
			modifiedAny = true
		}
		return nil
	}

	if err := foldPkg(prog.Primary); err != nil {
		return false, err
	}

	for _, depPkg := range prog.Dependencies {
		if err := foldPkg(depPkg); err != nil {
			return false, err
		}
	}

	return modifiedAny, nil
}

type foldVisitor struct {
	progScope *scope.ProgScope
	pkgScope  *scope.PkgScope
	pkg       *program.Package
	getImport program.GetImportFunc
	checker   ParameterChecker
	modified  bool
	localVals map[string][]float64
	// localConst is localVals as constant-evaluation values, changed in the
	// same places localVals is. evalConst runs once per expression node, and
	// building this map there instead cost a full copy of every known local
	// per node: quadratic in a body's size, which a body of thousands of
	// shared locals (a property expansion written down as names) could not
	// finish. Nothing constant evaluation does writes to it.
	localConst map[string]value
	params     map[string]bool
	locals     map[string]bool
	inFunction bool
	cache      map[*program.Operand]value

	// keepRoot is the assignment root a slider edits in place (a ranged
	// assignment's expression); it is stamped but never collapsed.
	keepRoot *program.Operand
	// inUIMarker counts the UI marker calls (uiPoint, uiTransform, uiSpline)
	// whose arguments are being visited; their literals the UI binds by
	// operand, so nothing inside is collapsed.
	inUIMarker int
}

func (v *foldVisitor) visitFnType(fnType *program.FnType) error {
	if fnType == nil {
		return nil
	}
	for _, group := range fnType.Groups {
		for _, fn := range group.Functions {
			if err := v.visitFunction(fn); err != nil {
				return err
			}
		}
	}
	return nil
}

func (v *foldVisitor) visitFunction(fn *program.Function) error {
	if fn == nil {
		return nil
	}
	prevLocalVals := v.localVals
	prevLocalConst := v.localConst
	prevParams := v.params
	prevLocals := v.locals
	prevInFunction := v.inFunction
	v.inFunction = true

	v.localVals = make(map[string][]float64)
	v.localConst = make(map[string]value)
	if prevLocalVals != nil {
		for k, val := range prevLocalVals {
			v.localVals[k] = val
			v.localConst[k] = constValueOf(val, nil)
		}
	}
	v.params = make(map[string]bool)
	if prevParams != nil {
		for k := range prevParams {
			v.params[k] = true
		}
	}
	for _, p := range fn.Parameters {
		v.params[p.Name] = true
	}
	v.locals = make(map[string]bool)
	if prevLocals != nil {
		for k := range prevLocals {
			v.locals[k] = true
		}
	}
	for _, a := range fn.Assignments {
		for _, varElem := range a.Variables {
			v.locals[varElem.Name] = true
		}
	}

	defer func() {
		v.localVals = prevLocalVals
		v.localConst = prevLocalConst
		v.params = prevParams
		v.locals = prevLocals
		v.inFunction = prevInFunction
	}()

	for _, nestedFnType := range fn.FnTypes {
		if err := v.visitFnType(nestedFnType); err != nil {
			return err
		}
	}
	for _, assign := range fn.Assignments {
		if _, err := program.VisitOperand(assign.Expression, v); err != nil {
			return err
		}
		if len(assign.Variables) == 1 {
			if len(assign.Expression.ConstValues) > 0 {
				if !sliceEqual(assign.Variables[0].ConstValues, assign.Expression.ConstValues) {
					v.modified = true
				}
				assign.Variables[0].ConstValues = assign.Expression.ConstValues
				v.localVals[assign.Variables[0].Name] = assign.Expression.ConstValues
				v.localConst[assign.Variables[0].Name] = constValueOf(assign.Expression.ConstValues, assign.Variables[0].Type)
			} else {
				if len(assign.Variables[0].ConstValues) > 0 {
					v.modified = true
				}
				assign.Variables[0].ConstValues = nil
				if v.localVals != nil {
					delete(v.localVals, assign.Variables[0].Name)
					delete(v.localConst, assign.Variables[0].Name)
				}
			}
		}
		for _, varElem := range assign.Variables {
			for _, prop := range varElem.Properties {
				if _, err := program.VisitOperand(prop.Expression, v); err != nil {
					return err
				}
			}
		}
	}
	for _, ret := range fn.Returns {
		if _, err := program.VisitOperand(ret, v); err != nil {
			return err
		}
	}
	return nil
}

// setConst records a fold verdict on op. A nil val (the operand depends on a
// runtime parameter) clears any value stamped by an earlier fold: the
// assembler folds with a nil checker, so its stamps must not survive a fold
// that knows about parameters.
//
// A scalar verdict also collapses the operand: a composite the fold has fully
// evaluated becomes its literal, keeping its type, value and source span. The
// subtree is otherwise dead weight that every later pass walks and every
// consumer ships; a catalog lookup inlined at each use left Fastener Cosmetics
// 99% evaluated constant arithmetic, 23 MB of program for 70 KB of shader.
// Only a fold that knows the parameters collapses: a stamp made without them
// is cleared by the next fold, but a collapse cannot be undone.
//
// The verdict lives on the operand alone. Its type used to carry a second
// copy, Type.constValue, which nothing read and every clone of the operand
// copied: more than half the operands an inlining pass built were those
// copies.
func (v *foldVisitor) setConst(op *program.Operand, val []float64) {
	if !sliceEqual(op.ConstValues, val) {
		v.modified = true
	}
	op.ConstValues = val
	if len(val) == 1 && v.collapsible(op) {
		op.T = &program.Operand_Value{Value: val[0]}
		v.modified = true
	}
}

// collapsible reports whether a scalar verdict on op may replace it with its
// literal: a composite of scalar type, under a fold that knows the
// parameters, that is not the literal a slider rewrites in place (a ranged
// assignment's root), not inside a UI marker's argument, and not an import
// call, which later passes resolve by name.
func (v *foldVisitor) collapsible(op *program.Operand) bool {
	if v.checker == nil || op == v.keepRoot || v.inUIMarker > 0 {
		return false
	}
	// Analyze does not type every operand an inlined body brings in, nor set
	// Kind on every scalar it does type, so the verdict's one value is the
	// test of scalarness: a typed list, matrix or function is not a scalar,
	// nor is a one-element list written out.
	if t := op.Type; t != nil && (t.ElemType != nil || t.IsList() || t.IsMatrix() || t.IsFunction()) {
		return false
	}
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		return t.Expression != nil && t.Expression.Operation != program.Operation_COLLECT
	case *program.Operand_FnCall:
		if t.FnCall == nil || t.FnCall.Reference == nil {
			return false
		}
		switch t.FnCall.Reference.Name {
		case "importStep", "importMesh", "uiPoint", "uiTransform", "uiSpline", "uiColor", "uiDirection":
			return false
		}
		return true
	}
	return false
}

func setConst(op *program.Operand, val []float64) {
	op.ConstValues = val
}

func (v *foldVisitor) VisitValue(op *program.Operand, value float64) (any, error) {
	if v.checker != nil && v.checker.IsParameter(op) {
		v.setConst(op, nil)
		return nil, nil
	}
	v.setConst(op, []float64{value})
	return nil, nil
}

func (v *foldVisitor) VisitReference(op *program.Operand, ref *program.Reference) (any, error) {
	if v.checker != nil && v.checker.IsParameter(op) {
		v.setConst(op, nil)
		return nil, nil
	}
	if ref.Prefix == "" && v.inFunction {
		if v.params[ref.Name] || v.locals[ref.Name] {
			if v.localVals != nil {
				if vals, ok := v.localVals[ref.Name]; ok && len(vals) > 0 {
					v.setConst(op, vals)
					return nil, nil
				}
			}
			v.setConst(op, nil)
			return nil, nil
		}
	}
	val, err := v.evalConst(op)
	if err == nil {
		v.setConst(op, val)
	} else if errors.Is(err, ErrParameter) {
		v.setConst(op, nil)
	}
	return nil, nil
}

func (v *foldVisitor) VisitExpression(op *program.Operand, expr *program.Expression) (any, error) {
	if expr.Operation == program.Operation_SELECT && len(expr.Operands) >= 3 {
		condOp := expr.Operands[0]
		if _, err := program.VisitOperand(condOp, v); err != nil {
			return nil, err
		}
		if len(condOp.ConstValues) > 0 {
			var selectedBranch *program.Operand
			if len(condOp.ConstValues) == 1 {
				c := condOp.ConstValues[0]
				if len(expr.Operands) == 3 {
					// 2-way select: condition ? trueBranch : falseBranch
					if c != 0 {
						selectedBranch = expr.Operands[1]
					} else {
						selectedBranch = expr.Operands[2]
					}
				} else if len(expr.Operands) == 4 {
					// 3-way select: condition > 0 ? pos : condition == 0 ? zero : neg
					if c > 0 {
						selectedBranch = expr.Operands[1]
					} else if c == 0 {
						selectedBranch = expr.Operands[2]
					} else {
						selectedBranch = expr.Operands[3]
					}
				}
			} else {
				// Vector condition: check if all non-zero or all zero
				allNonZero := true
				allZero := true
				for _, c := range condOp.ConstValues {
					if c == 0 {
						allNonZero = false
					} else {
						allZero = false
					}
				}
				if allNonZero {
					selectedBranch = expr.Operands[1]
				} else if allZero {
					if len(expr.Operands) == 3 {
						selectedBranch = expr.Operands[2]
					} else if len(expr.Operands) == 4 {
						selectedBranch = expr.Operands[2]
					}
				}
			}

			if selectedBranch != nil {
				v.modified = true
				// Replace op with selectedBranch and prune dead branches
				op.T = selectedBranch.T
				op.ConstValues = selectedBranch.ConstValues
				if selectedBranch.Type != nil {
					op.Type = selectedBranch.Type
				}
				// Recurse and fold only the active branch
				_, err := program.VisitOperand(op, v)
				return nil, err
			}
		}
	}

	hasNonConstChild := false
	for i, operand := range expr.Operands {
		if expr.Operation == program.Operation_ACCESS && i > 0 {
			continue
		}
		_, err := program.VisitOperand(operand, v)
		if err != nil {
			return nil, err
		}
		// For ACCESS, the second operand onwards can be a swizzle reference which is non-const
		isSwizzleRef := expr.Operation == program.Operation_ACCESS && i > 0 && operand.GetReference() != nil
		if len(operand.ConstValues) == 0 && !isSwizzleRef {
			hasNonConstChild = true
		}
	}
	if hasNonConstChild {
		v.setConst(op, nil)
		return nil, nil
	}
	if v.checker != nil && v.checker.IsParameter(op) {
		return nil, nil
	}
	val, err := v.evalConst(op)
	if err == nil {
		v.setConst(op, val)
	} else if errors.Is(err, ErrParameter) {
		v.setConst(op, nil)
	}
	return nil, nil
}

func (v *foldVisitor) VisitFnCall(op *program.Operand, fc *program.FnCall) (any, error) {
	hasNonConstChild := false
	isImportCall := fc.Reference != nil && (fc.Reference.Name == "importStep" || fc.Reference.Name == "importMesh")
	if fc.Reference != nil {
		switch fc.Reference.Name {
		case "uiPoint", "uiTransform", "uiSpline", "uiColor", "uiDirection":
			v.inUIMarker++
			defer func() { v.inUIMarker-- }()
		}
	}
	for _, arg := range fc.Arguments {
		_, err := program.VisitOperand(arg, v)
		if err != nil {
			return nil, err
		}
		if len(arg.ConstValues) == 0 && !isImportCall {
			hasNonConstChild = true
		}
	}
	if hasNonConstChild {
		v.setConst(op, nil)
		return nil, nil
	}
	if v.checker != nil && v.checker.IsParameter(op) {
		return nil, nil
	}
	val, err := v.evalConst(op)
	if err == nil {
		v.setConst(op, val)
	} else if errors.Is(err, ErrParameter) {
		v.setConst(op, nil)
	}
	return nil, nil
}

func (v *foldVisitor) VisitPropRef(op *program.Operand, pr *program.PropertyRef) (any, error) {
	return nil, nil
}

func (v *foldVisitor) evalConst(op *program.Operand) ([]float64, error) {
	ctx := &constCtx{
		progScope:  v.progScope,
		cur:        v.pkgScope,
		getImport:  v.getImport,
		evaluating: make(map[evalKey]bool),
		calls:      make(map[callKey]bool),
		cache:      v.cache,
		checker:    v.checker,
	}
	if len(v.localVals) > 0 {
		ctx.frames = append(ctx.frames, &frame{vals: v.localConst, owner: v.pkgScope})
	}
	val, err := ctx.operand(op)
	if err != nil {
		return nil, err
	}
	return val.floats(), nil
}

// constValueOf is a local's folded floats as a constant-evaluation value.
func constValueOf(f []float64, t *program.Type) value {
	if t != nil && t.ElemType != nil {
		elemLen := 1
		for elem := t.ElemType; elem != nil; elem = elem.ElemType {
			elemLen *= max(1, int(elem.Size))
		}
		numElems := int(t.Size)
		if len(f) == numElems*elemLen {
			elems := make([]value, numElems)
			for i := 0; i < numElems; i++ {
				chunk := f[i*elemLen : (i+1)*elemLen]
				elems[i] = constValueOf(chunk, t.ElemType)
			}
			return listVal(elems)
		}
	}
	if len(f) == 1 {
		return scalarVal(f[0])
	}
	return vectorVal(f)
}

func sliceEqual(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
