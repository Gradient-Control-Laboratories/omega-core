package inline

import (
	"gradientcontrol.com/omega/core/program"
)

// bindArg returns what a parameter bound to arg reads at each of its uses.
//
// A constant argument is copied to every place the callee reads it, and
// every callee it is passed on to copies it again. When it is a constant
// number the fold reduces each copy to its literal afterwards, but the copies
// are built first: Fastener Cosmetics' thread pitch, a catalog lookup of
// 1,619 operands, reached its readers 136 times, and one inlining pass built
// 17 MB of program that the next fold reduced to 350 KB. So a constant number
// is bound as the literal the fold would make of it: same type, value and
// source span. Anything the fold would leave as written is copied as before.
func (in *inliner) bindArg(arg *program.Operand) *program.Operand {
	if lit := in.literalOf(arg); lit != nil {
		return lit
	}
	return program.CloneOperand(arg)
}

// literalOf returns op as the literal the fold would collapse it to, or nil.
// It mirrors the fold's rule (eval's collapsible), and is stricter where the
// inliner knows less: op names no local of a function being transformed,
// since the evaluator would read a top-level declaration of the same name
// in its place, and no UI marker or import call lies anywhere under it.
func (in *inliner) literalOf(op *program.Operand) *program.Operand {
	if in.evaluator == nil || in.inUIMarker > 0 || op == nil {
		return nil
	}
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		if t.Expression == nil || t.Expression.Operation == program.Operation_COLLECT {
			return nil
		}
	case *program.Operand_FnCall:
		if t.FnCall == nil || t.FnCall.Reference == nil {
			return nil
		}
	default:
		return nil
	}
	if t := op.Type; t != nil && (t.ElemType != nil || t.IsList() || t.IsMatrix() || t.IsFunction()) {
		return nil
	}
	if !in.isHoistableConst(op) || in.callsPOrTime(op, in.currentFileScope) || hasMarkerOrImport(op) {
		return nil
	}
	val, ok := in.evaluator.Scalar(op, in.currentPkgScope)
	if !ok {
		return nil
	}
	return &program.Operand{
		Start:       program.CloneLocation(op.Start),
		End:         program.CloneLocation(op.End),
		Type:        program.CloneType(op.Type),
		T:           &program.Operand_Value{Value: val},
		ConstValues: []float64{val},
	}
}

// isUIMarker reports whether fc is a UI marker, whose literal arguments the
// viewer binds by operand.
func isUIMarker(fc *program.FnCall) bool {
	if fc == nil || fc.Reference == nil {
		return false
	}
	switch fc.Reference.Name {
	case "uiPoint", "uiTransform", "uiSpline", "uiColor", "uiDirection":
		return true
	}
	return false
}

// hasMarkerOrImport reports whether a UI marker or an import call lies
// anywhere under op.
func hasMarkerOrImport(op *program.Operand) bool {
	if op == nil {
		return false
	}
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		for _, child := range t.Expression.GetOperands() {
			if hasMarkerOrImport(child) {
				return true
			}
		}
	case *program.Operand_FnCall:
		if isUIMarker(t.FnCall) {
			return true
		}
		if ref := t.FnCall.GetReference(); ref != nil && (ref.Name == "importStep" || ref.Name == "importMesh") {
			return true
		}
		for _, arg := range t.FnCall.GetArguments() {
			if hasMarkerOrImport(arg) {
				return true
			}
		}
	}
	return false
}
