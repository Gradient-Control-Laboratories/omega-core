package program

import (
	"fmt"
	"strings"
)

func (o *Operand) PrettyPrint(indent int) string {
	switch t := o.T.(type) {
	case *Operand_Value:
		return fmt.Sprintf("%v", t.Value)
	case *Operand_Reference:
		return t.Reference.PrettyPrint(indent)
	case *Operand_Expression:
		return t.Expression.PrettyPrint(indent)
	case *Operand_FnCall:
		return t.FnCall.PrettyPrint(indent)
	case *Operand_PropRef:
		return t.PropRef.PrettyPrint(indent)
	}
	return fmt.Sprintf("%v", o)
}

func (r *Reference) PrettyPrint(indent int) string {
	var sb strings.Builder
	if r.Prefix != "" {
		sb.WriteString(r.Prefix)
		sb.WriteString(".")
	}
	sb.WriteString(r.Name)
	return sb.String()
}

func (fc *FnCall) PrettyPrint(indent int) string {
	var sb strings.Builder
	sb.WriteString(fc.Reference.PrettyPrint(indent))
	sb.WriteString("(")
	if len(fc.Arguments) > 0 {
		sb.WriteString(fc.Arguments[0].PrettyPrint(indent))
		for i := 1; i < len(fc.Arguments); i++ {
			sb.WriteString(", ")
			sb.WriteString(fc.Arguments[i].PrettyPrint(indent))
		}
	}
	sb.WriteString(")")
	return sb.String()
}

func (pr *PropertyRef) PrettyPrint(indent int) string {
	var sb strings.Builder
	sb.WriteString(pr.Reference.PrettyPrint(indent))
	sb.WriteString(".")
	sb.WriteString(pr.Property)
	return sb.String()
}

func (e *Expression) PrettyPrint(indent int) string {
	switch e.Operation {

	// basic

	case Operation_ADD:
		return prettyPrintBasic(indent, "+", e.Operands)
	case Operation_SUB:
		return prettyPrintBasic(indent, "-", e.Operands)
	case Operation_MUL:
		return prettyPrintBasic(indent, "*", e.Operands)
	case Operation_DIV:
		return prettyPrintBasic(indent, "/", e.Operands)

	// multi

	case Operation_MIN:
		return prettyPrintOp(indent, "min", e.Operands)
	case Operation_MAX:
		return prettyPrintOp(indent, "max", e.Operands)
	case Operation_AND:
		return prettyPrintOp(indent, "and", e.Operands)
	case Operation_OR:
		return prettyPrintOp(indent, "or", e.Operands)

	// unary

	case Operation_SQUARE:
		return prettyPrintOp(indent, "square", e.Operands)
	case Operation_FLOOR:
		return prettyPrintOp(indent, "floor", e.Operands)
	case Operation_CEIL:
		return prettyPrintOp(indent, "ceil", e.Operands)
	case Operation_ROUND:
		return prettyPrintOp(indent, "round", e.Operands)
	case Operation_SQRT:
		return prettyPrintOp(indent, "sqrt", e.Operands)
	case Operation_NEG:
		return prettyPrintOp(indent, "neg", e.Operands)
	case Operation_SIN:
		return prettyPrintOp(indent, "sin", e.Operands)
	case Operation_COS:
		return prettyPrintOp(indent, "cos", e.Operands)
	case Operation_TAN:
		return prettyPrintOp(indent, "tan", e.Operands)
	case Operation_ASIN:
		return prettyPrintOp(indent, "asin", e.Operands)
	case Operation_ACOS:
		return prettyPrintOp(indent, "acos", e.Operands)
	case Operation_ATAN:
		return prettyPrintOp(indent, "atan", e.Operands)
	case Operation_EXP:
		return prettyPrintOp(indent, "exp", e.Operands)
	case Operation_LN:
		return prettyPrintOp(indent, "ln", e.Operands)
	case Operation_NOT:
		return prettyPrintOp(indent, "not", e.Operands)
	case Operation_ABS:
		return prettyPrintOp(indent, "abs", e.Operands)

	case Operation_DERIV:
		return prettyPrintOp(indent, "op:DERIV", e.Operands)
	case Operation_SECONDCLOSEST:
		return prettyPrintOp(indent, "op:SECONDCLOSEST", e.Operands)
	case Operation_BOUNDINGBOX:
		return prettyPrintOp(indent, "op:BOUNDINGBOX", e.Operands)
	case Operation_SURFACEU:
		return prettyPrintOp(indent, "op:SURFACEU", e.Operands)
	case Operation_SURFACEV:
		return prettyPrintOp(indent, "op:SURFACEV", e.Operands)

	// binary

	case Operation_COMPARE:
		return prettyPrintOp(indent, "compare", e.Operands)
	case Operation_MODULO:
		return prettyPrintOp(indent, "modulo", e.Operands)
	case Operation_ATAN2:
		return prettyPrintOp(indent, "atan2", e.Operands)
	case Operation_POW:
		return prettyPrintOp(indent, "pow", e.Operands)

	// comparison

	case Operation_GT:
		return prettyPrintBasic(indent, ">", e.Operands)
	case Operation_GE:
		return prettyPrintBasic(indent, ">=", e.Operands)
	case Operation_EQ:
		return prettyPrintBasic(indent, "==", e.Operands)
	case Operation_NE:
		return prettyPrintBasic(indent, "!=", e.Operands)
	case Operation_LE:
		return prettyPrintBasic(indent, "<=", e.Operands)
	case Operation_LT:
		return prettyPrintBasic(indent, "<", e.Operands)

	// custom / other

	case Operation_MAP:
		return prettyPrintOp(indent, "map", e.Operands)
	case Operation_REDUCE:
		return prettyPrintOp(indent, "reduce", e.Operands)
	case Operation_REMAP:
		return prettyPrintOp(indent, "remap", e.Operands)
	case Operation_COLLECT:
		return prettyPrintOp(indent, "collect", e.Operands)
	case Operation_ACCESS:
		return prettyPrintOp(indent, "access", e.Operands)
	case Operation_SELECT:
		return prettyPrintOp(indent, "select", e.Operands)
	case Operation_INDEX:
		return prettyPrintOp(indent, "index", e.Operands)
	}

	return fmt.Sprintf("%v", e)
}

func prettyPrintBasic(indent int, operation string, operand []*Operand) string {
	result := strings.Builder{}
	result.WriteString("(")
	result.WriteString(operand[0].PrettyPrint(indent))
	for i := 1; i < len(operand); i++ {
		result.WriteString(fmt.Sprintf(" %s ", operation))
		result.WriteString(operand[i].PrettyPrint(indent))
	}
	result.WriteString(")")
	return result.String()
}

func prettyPrintOp(indent int, name string, operands []*Operand) string {
	result := strings.Builder{}
	result.WriteString(name)
	result.WriteString("(")
	if len(operands) > 0 {
		result.WriteString(operands[0].PrettyPrint(indent))
		for i := 1; i < len(operands); i++ {
			result.WriteString(", ")
			result.WriteString(operands[i].PrettyPrint(indent))
		}
	}
	result.WriteString(")")
	return result.String()
}

func (f *Function) PrettyPrint(indent int) string {
	result := strings.Builder{}
	result.WriteString("(")
	parameters := f.Parameters
	for i, param := range parameters {
		if i > 0 {
			result.WriteString(", ")
		}
		if param.Range != nil {
			minStr := param.Range.Min.PrettyPrint(indent)
			maxStr := param.Range.Max.PrettyPrint(indent)
			if param.Range.Period != nil {
				periodStr := param.Range.Period.PrettyPrint(indent)
				result.WriteString(fmt.Sprintf("@range(%s, %s, %s) ", minStr, maxStr, periodStr))
			} else {
				result.WriteString(fmt.Sprintf("@range(%s, %s) ", minStr, maxStr))
			}
		}
		if param.Kind != "" {
			result.WriteString("@" + param.Kind)
			if param.Anchor != "" {
				result.WriteString("(" + param.Anchor + ")")
			}
			result.WriteString(" ")
		}
		result.WriteString(param.Name)
	}
	result.WriteString(") {\n")

	var body strings.Builder
	assignsStr := printAssignments(indent+1, f.Assignments)
	body.WriteString(assignsStr)

	fnTypesStr := printFnTypes(indent+1, f.FnTypes)
	if fnTypesStr != "" {
		if body.Len() > 0 {
			body.WriteString(fnTypesStr)
		} else {
			body.WriteString(strings.TrimPrefix(fnTypesStr, "\n"))
		}
	}

	preceded := body.Len() > 0
	body.WriteString(printReturns(indent+1, f.Returns, preceded))

	result.WriteString(body.String())
	result.WriteString(strings.Repeat(" ", indent*3))
	result.WriteString("}")
	return result.String()
}

func (a *Assignment) PrettyPrint(indent int) string {
	result := strings.Builder{}
	if a.Range != nil && !a.Range.IsSynthetic() {
		minStr := a.Range.Min.PrettyPrint(indent)
		maxStr := a.Range.Max.PrettyPrint(indent)
		if a.Range.Period != nil {
			periodStr := a.Range.Period.PrettyPrint(indent)
			result.WriteString(fmt.Sprintf("@range(%s, %s, %s)\n", minStr, maxStr, periodStr))
		} else {
			result.WriteString(fmt.Sprintf("@range(%s, %s)\n", minStr, maxStr))
		}
		result.WriteString(strings.Repeat(" ", indent*3))
	}
	vars := a.Variables
	if len(vars) > 0 {
		result.WriteString(vars[0].Name)
		for i := 1; i < len(vars); i++ {
			result.WriteString(", ")
			result.WriteString(vars[i].Name)
		}
	}
	if a.Expression != nil {
		result.WriteString(" = ")
		result.WriteString(a.Expression.PrettyPrint(indent))
	}
	for _, v := range vars {
		for _, prop := range v.Properties {
			result.WriteString("\n")
			result.WriteString(strings.Repeat(" ", indent*3))
			result.WriteString(v.Name)
			result.WriteString(".")
			result.WriteString(prop.Name)
			result.WriteString(" = ")
			result.WriteString(prop.Expression.PrettyPrint(indent))
		}
	}
	return result.String()
}

func (ft *FnType) PrettyPrint(indent int) string {
	result := strings.Builder{}
	for _, group := range ft.Groups {
		for _, fn := range group.Functions {
			result.WriteString(strings.Repeat(" ", indent*3))
			result.WriteString(ft.Name)
			result.WriteString(fn.PrettyPrint(indent))
			result.WriteString("\n")
		}
	}
	return result.String()
}

func (p *PackageFile) PrettyPrint(indent int) string {
	var sb strings.Builder

	importsStr := printImports(indent, p.Imports) +
		printMeshImports(indent, p.MeshImports) +
		printStepImports(indent, p.StepImports)
	sb.WriteString(importsStr)

	assignsStr := printAssignments(indent, p.Assignments)
	if assignsStr != "" {
		if sb.Len() > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(assignsStr)
	}

	fnTypesStr := printFnTypes(indent, p.FnTypes)
	if fnTypesStr != "" {
		if sb.Len() > 0 {
			sb.WriteString(fnTypesStr)
		} else {
			sb.WriteString(strings.TrimPrefix(fnTypesStr, "\n"))
		}
	}

	preceded := sb.Len() > 0
	sb.WriteString(printReturns(indent, p.Returns, preceded))

	return sb.String()
}

func (p *Package) PrettyPrint(indent int) string {
	result := strings.Builder{}
	if p.Scene != nil {
		result.WriteString(strings.Repeat(" ", indent*3))
		result.WriteString(fmt.Sprintf("@scene(%g, %g)\n\n", p.Scene.Size, p.Scene.Grid))
	}
	for i, file := range p.Files {
		if i > 0 {
			result.WriteString("\n")
		}
		result.WriteString(file.PrettyPrint(indent))
	}
	return result.String()
}

func (p *Program) PrettyPrint(indent int) string {
	if p.Primary == nil {
		return ""
	}
	result := strings.Builder{}
	result.WriteString(p.Primary.PrettyPrint(indent))
	for _, dep := range p.Dependencies {
		result.WriteString(fmt.Sprintf("\n\n// ==================\n// Package: %s\n\n", dep.Path))
		result.WriteString(dep.PrettyPrint(indent))
	}
	return result.String()
}

func printImports(indent int, imports []*Import) string {
	result := strings.Builder{}
	for _, imp := range imports {
		result.WriteString(strings.Repeat(" ", indent*3))
		result.WriteString("import ")
		if imp.Prefix != "" {
			result.WriteString(imp.Prefix)
			result.WriteString(" ")
		}
		result.WriteString("\"")
		result.WriteString(imp.Path)
		result.WriteString("\";\n")
	}
	return result.String()
}

func printMeshImports(indent int, meshImports []*MeshImport) string {
	result := strings.Builder{}
	for _, imp := range meshImports {
		result.WriteString(strings.Repeat(" ", indent*3))
		result.WriteString(fmt.Sprintf("importMesh %d \"%s\";\n", imp.Index, imp.Path))
	}
	return result.String()
}

func printStepImports(indent int, stepImports []*StepImport) string {
	result := strings.Builder{}
	for _, imp := range stepImports {
		result.WriteString(strings.Repeat(" ", indent*3))
		result.WriteString(fmt.Sprintf("importStep %d \"%s\";\n", imp.Index, imp.Path))
	}
	return result.String()
}

func printAssignments(indent int, assignments []*Assignment) string {
	result := strings.Builder{}
	for _, assign := range assignments {
		result.WriteString(strings.Repeat(" ", indent*3))
		result.WriteString(assign.PrettyPrint(indent))
		result.WriteString("\n")
	}
	return result.String()
}

func printFnTypes(indent int, fnTypes []*FnType) string {
	result := strings.Builder{}
	for _, ft := range fnTypes {
		result.WriteString("\n")
		result.WriteString(ft.PrettyPrint(indent))
	}
	return result.String()
}

func printReturns(indent int, returns []*Operand, preceded bool) string {
	result := strings.Builder{}
	if len(returns) > 0 {
		if preceded {
			result.WriteString("\n")
		}
		result.WriteString(strings.Repeat(" ", indent*3))
		result.WriteString("return ")
		result.WriteString(returns[0].PrettyPrint(indent))
		for i := 1; i < len(returns); i++ {
			result.WriteString(", ")
			result.WriteString(returns[i].PrettyPrint(indent))
		}
		result.WriteString(";\n")
	}
	return result.String()
}
