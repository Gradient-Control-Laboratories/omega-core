package ui

import (
	"gradientcontrol.com/omega/core/internal/edit"
	"gradientcontrol.com/omega/core/internal/eval"
	"gradientcontrol.com/omega/core/program"
)

const (
	KindPoint     = "point"
	KindTransform = "transform"
	KindSpline    = "spline"
	KindColor     = "color"
	KindDirection = "direction"
)

type Control struct {
	Kind     string    `json:"kind"`
	Span     edit.Span `json:"span"`
	Value    []float64 `json:"value"`
	Params   []Param   `json:"params,omitempty"`
	Controls []Control `json:"controls,omitempty"`
	// Anchor is where a direction's arrow is drawn from, uiDirection's
	// second argument evaluated as a constant; nil draws it from the origin.
	Anchor []float64 `json:"anchor,omitempty"`
}

type Param struct {
	ParamName string    `json:"paramName"`
	Span      edit.Span `json:"span"`
	Value     float64   `json:"value"`
	operand   *program.Operand
}

func Extract(pkg *program.Package, getImport program.GetImportFunc) []Control {
	controls, _ := ExtractWithOperands(pkg, getImport)
	return controls
}

// ExtractWithOperands also binds editable literals to their actual AST operands.
// Inlining can copy a control, or put an unrelated imported literal at the same
// source span. Every control copy shares its parameter; other literals do not.
func ExtractWithOperands(pkg *program.Package, getImport program.GetImportFunc) ([]Control, map[*program.Operand]edit.Span) {
	operands := make(map[*program.Operand]edit.Span)
	if pkg == nil {
		return nil, operands
	}
	var out []Control
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
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

func collectFunction(fn *program.Function, pkg *program.Package, getImport program.GetImportFunc, out *[]Control, operands map[*program.Operand]edit.Span) {
	if fn == nil {
		return
	}
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

func collectFnType(fnType *program.FnType, pkg *program.Package, getImport program.GetImportFunc, out *[]Control, operands map[*program.Operand]edit.Span) {
	if fnType == nil {
		return
	}
	for _, group := range fnType.Groups {
		if group == nil {
			continue
		}
		for _, fn := range group.Functions {
			collectFunction(fn, pkg, getImport, out, operands)
		}
	}
}

func collect(op *program.Operand, pkg *program.Package, getImport program.GetImportFunc, out *[]Control, operands map[*program.Operand]edit.Span) {
	if op == nil {
		return
	}
	switch v := op.T.(type) {
	case *program.Operand_FnCall:
		fc := v.FnCall
		if fc == nil {
			return
		}
		if c, ok := controlFromCall(fc, pkg, getImport); ok {
			collectParamOperands(c, operands)
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

func collectParamOperands(control Control, operands map[*program.Operand]edit.Span) {
	for _, param := range control.Params {
		operands[param.operand] = param.Span
	}
	for _, child := range control.Controls {
		collectParamOperands(child, operands)
	}
}

func controlFromCall(fc *program.FnCall, pkg *program.Package, getImport program.GetImportFunc) (Control, bool) {
	if fc == nil || fc.Reference == nil || fc.Reference.Prefix != "" || len(fc.Arguments) == 0 {
		return Control{}, false
	}
	kind := ""
	want := 0
	var anchor []float64
	switch fc.Reference.Name {
	case "uiPoint":
		kind, want = KindPoint, 3
	case "uiTransform":
		kind, want = KindTransform, 16
	case "uiColor":
		kind, want = KindColor, 3
	case "uiDirection":
		kind, want = KindDirection, 3
		if len(fc.Arguments) == 2 {
			if v, err := eval.ConstPkg(fc.Arguments[1], pkg, getImport); err == nil && len(v) == 3 {
				anchor = append([]float64(nil), v...)
			}
		}
	case "uiSpline":
		return splineControl(fc.Arguments[0], pkg, getImport)
	default:
		return Control{}, false
	}
	arg := fc.Arguments[0]
	value := arg.ConstValues
	if len(value) == 0 {
		var err error
		value, err = eval.ConstPkg(arg, pkg, getImport)
		if err != nil {
			return Control{}, false
		}
	}
	if len(value) != want {
		return Control{}, false
	}
	params := literalParams(editableOperand(arg, pkg))
	if len(params) != want {
		params = nil
	}
	return Control{
		Kind:   kind,
		Span:   editableSpan(arg, pkg),
		Value:  append([]float64(nil), value...),
		Params: params,
		Anchor: anchor,
	}, true
}

func splineControl(arg *program.Operand, pkg *program.Package, getImport program.GetImportFunc) (Control, bool) {
	list := editableOperand(arg, pkg)
	expr := list.GetExpression()
	if expr == nil || expr.Operation != program.Operation_COLLECT || len(expr.Operands) < 3 || len(expr.Operands)%2 != 1 {
		return Control{}, false
	}
	c := Control{Kind: KindSpline, Span: spanOf(list)}
	for _, op := range expr.Operands {
		fc := editableOperand(op, pkg).GetFnCall()
		if fc == nil || fc.Reference == nil || fc.Reference.Name != "uiTransform" {
			return Control{}, false
		}
		child, ok := controlFromCall(fc, pkg, getImport)
		if !ok {
			return Control{}, false
		}
		c.Controls = append(c.Controls, child)
		c.Value = append(c.Value, child.Value...)
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
		assign := assignmentFor(name, pkg)
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

func assignmentFor(name string, pkg *program.Package) *program.Operand {
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, a := range f.Assignments {
			if a == nil || len(a.Variables) != 1 || a.Variables[0] == nil {
				continue
			}
			if a.Variables[0].Name == name {
				return a.Expression
			}
		}
	}
	return nil
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

func literalParams(op *program.Operand) []Param {
	var out []Param
	collectLiteralParams(op, &out)
	return out
}

func collectLiteralParams(op *program.Operand, out *[]Param) {
	if op == nil {
		return
	}
	switch v := op.T.(type) {
	case *program.Operand_Value:
		span := spanOf(op)
		*out = append(*out, Param{ParamName: span.VarName(), Span: span, Value: v.Value, operand: op})
	case *program.Operand_Expression:
		if v.Expression == nil || v.Expression.Operation != program.Operation_COLLECT {
			return
		}
		for _, child := range v.Expression.Operands {
			collectLiteralParams(child, out)
		}
	}
}
