package ui

import (
	"gradientcontrol.com/omega/core/assembler/ui"
	"gradientcontrol.com/omega/core/program/edit"
	"gradientcontrol.com/omega/core/internal/eval"
	"gradientcontrol.com/omega/core/program"
)

func Extract(pkg *program.Package, getImport program.GetImportFunc) []ui.Control {
	controls, _ := ExtractWithOperands(pkg, getImport)
	return controls
}

// ExtractWithOperands also binds editable literals to their actual AST operands.
// Inlining can copy a control, or put an unrelated imported literal at the same
// source span. Every control copy shares its parameter; other literals do not.
func ExtractWithOperands(pkg *program.Package, getImport program.GetImportFunc) ([]ui.Control, map[*program.Operand]edit.Span) {
	operands := make(map[*program.Operand]edit.Span)
	if pkg == nil {
		return nil, operands
	}
	var out []ui.Control
	// Invariant: pkg.Files, FnTypes, Groups, and Functions are guaranteed non-nil by the parser.
	for _, f := range pkg.Files {
		for _, assign := range f.Assignments {
			collect(assign.Expression, pkg, getImport, &out, operands)
		}
		for _, fnType := range f.FnTypes {
			collectFnType(fnType, pkg, getImport, &out, operands)
		}
		for _, ret := range f.Returns {
			collect(ret, pkg, getImport, &out, operands)
		}
	}
	return out, operands
}

func collectFunction(fn *program.Function, pkg *program.Package, getImport program.GetImportFunc, out *[]ui.Control, operands map[*program.Operand]edit.Span) {
	pkg = &program.Package{Path: pkg.Path, Files: append([]*program.PackageFile{{Assignments: fn.Assignments, FnTypes: fn.FnTypes}}, pkg.Files...)}
	for _, assign := range fn.Assignments {
		collect(assign.Expression, pkg, getImport, out, operands)
	}
	for _, ret := range fn.Returns {
		collect(ret, pkg, getImport, out, operands)
	}
	for _, fnType := range fn.FnTypes {
		collectFnType(fnType, pkg, getImport, out, operands)
	}
}

func collectFnType(fnType *program.FnType, pkg *program.Package, getImport program.GetImportFunc, out *[]ui.Control, operands map[*program.Operand]edit.Span) {
	for _, group := range fnType.Groups {
		for _, fn := range group.Functions {
			collectFunction(fn, pkg, getImport, out, operands)
		}
	}
}

func collect(op *program.Operand, pkg *program.Package, getImport program.GetImportFunc, out *[]ui.Control, operands map[*program.Operand]edit.Span) {
	if op == nil {
		return
	}
	switch v := op.T.(type) {
	case *program.Operand_FnCall:
		fc := v.FnCall
		if fc == nil {
			return
		}
		if c, ok := controlFromCall(fc, pkg, getImport, operands); ok {
			duplicate := false
			for _, existing := range *out {
				if existing.Kind == c.Kind && existing.Span == c.Span {
					duplicate = true
					break
				}
			}
			if !duplicate {
				*out = append(*out, c)
			}
		}
		for _, arg := range fc.Arguments {
			collect(arg, pkg, getImport, out, operands)
		}
	case *program.Operand_Expression:
		if v.Expression == nil {
			return
		}
		for _, arg := range v.Expression.Operands {
			collect(arg, pkg, getImport, out, operands)
		}
	}
}

func controlFromCall(fc *program.FnCall, pkg *program.Package, getImport program.GetImportFunc, operands map[*program.Operand]edit.Span) (ui.Control, bool) {
	if fc == nil || fc.Reference == nil || fc.Reference.Prefix != "" || len(fc.Arguments) == 0 {
		return ui.Control{}, false
	}
	kind := ""
	want := 0
	var anchor []float64
	switch fc.Reference.Name {
	case "uiPoint":
		kind, want = ui.KindPoint, 3
	case "uiTransform":
		kind, want = ui.KindTransform, 16
	case "uiColor":
		kind, want = ui.KindColor, 3
	case "uiDirection":
		kind, want = ui.KindDirection, 3
		if len(fc.Arguments) == 2 {
			if v, err := eval.ConstPkg(fc.Arguments[1], pkg, getImport); err == nil && len(v) == 3 {
				anchor = append([]float64(nil), v...)
			}
		}
	case "uiSpline":
		return splineControl(fc.Arguments[0], pkg, getImport, operands)
	default:
		return ui.Control{}, false
	}
	arg := fc.Arguments[0]
	value := arg.ConstValues
	if len(value) == 0 {
		var err error
		value, err = eval.ConstPkg(arg, pkg, getImport)
		if err != nil {
			return ui.Control{}, false
		}
	}
	if len(value) != want {
		return ui.Control{}, false
	}
	params, ops := literalParams(editableOperand(arg, pkg))
	if len(params) != want {
		params = nil
	} else {
		for i, op := range ops {
			operands[op] = params[i].Span
		}
	}
	return ui.Control{
		Kind:   kind,
		Span:   editableSpan(arg, pkg),
		Value:  append([]float64(nil), value...),
		Params: params,
		Anchor: anchor,
	}, true
}

func splineControl(arg *program.Operand, pkg *program.Package, getImport program.GetImportFunc, operands map[*program.Operand]edit.Span) (ui.Control, bool) {
	list := editableOperand(arg, pkg)
	expr := list.GetExpression()
	if expr == nil || expr.Operation != program.Operation_COLLECT || len(expr.Operands) < 3 || len(expr.Operands)%2 != 1 {
		return ui.Control{}, false
	}
	c := ui.Control{Kind: ui.KindSpline, Span: spanOf(list)}
	childOperands := make(map[*program.Operand]edit.Span)
	for _, op := range expr.Operands {
		fc := editableOperand(op, pkg).GetFnCall()
		if fc == nil || fc.Reference == nil || fc.Reference.Name != "uiTransform" {
			return ui.Control{}, false
		}
		child, ok := controlFromCall(fc, pkg, getImport, childOperands)
		if !ok {
			return ui.Control{}, false
		}
		c.Controls = append(c.Controls, child)
		c.Value = append(c.Value, child.Value...)
	}
	for op, span := range childOperands {
		operands[op] = span
	}
	return c, true
}

func editableSpan(op *program.Operand, pkg *program.Package) edit.Span {
	return spanOf(editableOperand(op, pkg))
}

func editableOperand(op *program.Operand, pkg *program.Package) *program.Operand {
	seen := map[string]bool{}
	for name := getReferenceName(op); name != "" && !seen[name]; name = getReferenceName(op) {
		seen[name] = true
		assign := program.TopLevelAssignment(pkg, name)
		if assign == nil {
			break
		}
		op = assign
	}
	return op
}

func getReferenceName(op *program.Operand) string {
	if op == nil {
		return ""
	}
	refOp, ok := op.T.(*program.Operand_Reference)
	if !ok || refOp == nil || refOp.Reference == nil || refOp.Reference.Prefix != "" {
		return ""
	}
	return refOp.Reference.Name
}

func spanOf(op *program.Operand) edit.Span {
	if op == nil || op.Start == nil || op.End == nil {
		return edit.Span{}
	}
	return edit.Span{
		StartLine:   int(op.Start.Line),
		StartColumn: int(op.Start.Column),
		EndLine:     int(op.End.Line),
		EndColumn:   int(op.End.Column),
	}
}

func literalParams(op *program.Operand) ([]ui.Param, []*program.Operand) {
	var out []ui.Param
	var ops []*program.Operand
	collectLiteralParams(op, &out, &ops)
	return out, ops
}

func collectLiteralParams(op *program.Operand, out *[]ui.Param, ops *[]*program.Operand) {
	if op == nil {
		return
	}
	switch v := op.T.(type) {
	case *program.Operand_Value:
		span := spanOf(op)
		*out = append(*out, ui.Param{ParamName: span.VarName(), Span: span, Value: v.Value})
		*ops = append(*ops, op)
	case *program.Operand_Expression:
		if v.Expression == nil || v.Expression.Operation != program.Operation_COLLECT {
			return
		}
		for _, child := range v.Expression.Operands {
			collectLiteralParams(child, out, ops)
		}
	}
}
