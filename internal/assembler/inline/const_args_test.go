package inline_test

import (
	"testing"

	"gradientcontrol.com/omega/core/internal/assembler/inline"
	"gradientcontrol.com/omega/core/internal/closure"
	"gradientcontrol.com/omega/core/internal/collect"
	"gradientcontrol.com/omega/core/internal/link"
	"gradientcontrol.com/omega/core/internal/lower"
	"gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/types"
	parser "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

// literalChecker reports the literals under some top-level declarations as
// runtime parameters, as the assembler's checker does for sliders and UI
// markers.
type literalChecker map[*program.Operand]bool

func (c literalChecker) IsParameter(op *program.Operand) bool { return c[op] }

// inlineWithParameters runs runPipeline's passes with one inlining pass that
// knows the parameters: the literals under the declarations of the names in
// params. A nil params runs the plain Transform.
func inlineWithParameters(t *testing.T, src string, params []string) *program.Program {
	t.Helper()
	srcMgr := &source.MockManager{}
	prog, err := collect.Program([]string{src}, srcMgr, parser.ParsePackage)
	if err != nil {
		t.Fatal(err)
	}
	step := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}
	step(link.References(prog, srcMgr, parser.ParsePackage))
	step(closure.Transform(prog))
	step(link.References(prog, srcMgr, parser.ParsePackage))
	step(types.Analyze(prog))
	step(lower.Transform(prog))
	step(link.References(prog, srcMgr, parser.ParsePackage))
	step(types.Analyze(prog))
	if params == nil {
		_, err = inline.Transform(prog)
	} else {
		checker := literalChecker{}
		var mark func(op *program.Operand)
		mark = func(op *program.Operand) {
			if op == nil {
				return
			}
			if _, ok := op.T.(*program.Operand_Value); ok {
				checker[op] = true
			}
			for _, child := range op.GetExpression().GetOperands() {
				mark(child)
			}
			for _, arg := range op.GetFnCall().GetArguments() {
				mark(arg)
			}
		}
		for _, a := range prog.Primary.Files[0].Assignments {
			for _, name := range params {
				if a.Variables[0].Name == name {
					mark(a.Expression)
				}
			}
		}
		_, err = inline.TransformWithParameters(prog, checker)
	}
	step(err)
	return prog
}

// count returns how many operands under op satisfy pred.
func count(op *program.Operand, pred func(*program.Operand) bool) int {
	if op == nil {
		return 0
	}
	n := 0
	if pred(op) {
		n++
	}
	for _, child := range op.GetExpression().GetOperands() {
		n += count(child, pred)
	}
	for _, arg := range op.GetFnCall().GetArguments() {
		n += count(arg, pred)
	}
	return n
}

func isValue(v float64) func(*program.Operand) bool {
	return func(op *program.Operand) bool {
		val, ok := op.T.(*program.Operand_Value)
		return ok && val.Value == v
	}
}

func isOperation(o program.Operation) func(*program.Operand) bool {
	return func(op *program.Operand) bool {
		return op.GetExpression() != nil && op.GetExpression().Operation == o
	}
}

func isReferenceTo(name string) func(*program.Operand) bool {
	return func(op *program.Operand) bool {
		ref := op.GetReference()
		return ref != nil && ref.Name == name
	}
}

func primaryReturn(prog *program.Program) *program.Operand {
	files := prog.Primary.Files
	return files[len(files)-1].Returns[0]
}

// A constant number passed as an argument is bound as its literal, so each
// read of the parameter is the number and not a copy of the expression.
// Without the parameters known, the expression is copied to every read.
func TestTransformWithParametersBindsConstantArgumentAsLiteral(t *testing.T) {
	src := `
k(x) { return x * 2.0 + 1.0; }
h(a) { return a + a + a; }
return h(k(3.0));
`
	ret := primaryReturn(inlineWithParameters(t, src, []string{}))
	if n := count(ret, isValue(7)); n != 3 {
		t.Errorf("want each of h's three reads to be the literal 7, got %d: %s", n, ret.PrettyPrint(0))
	}
	if n := count(ret, isOperation(program.Operation_MUL)); n != 0 {
		t.Errorf("k's expression was copied: %s", ret.PrettyPrint(0))
	}

	ret = primaryReturn(inlineWithParameters(t, src, nil))
	if n := count(ret, isOperation(program.Operation_MUL)); n != 3 {
		t.Errorf("without the parameters, want k's expression at each of three reads, got %d: %s", n, ret.PrettyPrint(0))
	}
}

// An argument that reads a slider is not a constant, even through a call's
// body: the slider stays live at every read.
func TestTransformWithParametersKeepsSliderArgument(t *testing.T) {
	src := `
r = 2.0;
k(x) { return x + r; }
h(a) { return a + a; }
return h(k(1.0));
`
	ret := primaryReturn(inlineWithParameters(t, src, []string{"r"}))
	if n := count(ret, isValue(3)); n != 0 {
		t.Errorf("the slider r was evaluated into the argument: %s", ret.PrettyPrint(0))
	}
	if n := count(ret, isReferenceTo("r")); n != 2 {
		t.Errorf("want both reads of the argument to read r, got %d: %s", n, ret.PrettyPrint(0))
	}
}

// A parameter that shares its name with a top-level constant is not that
// constant: the argument d * 2.0 inside f reads f's d.
func TestTransformWithParametersKeepsEnclosingParameter(t *testing.T) {
	src := `
d = 5.0;
h(a) { return a + a; }
f(d) { return h(d * 2.0); }
return f(p.x);
`
	prog := inlineWithParameters(t, src, []string{})
	var body *program.Operand
	for _, ft := range prog.Primary.Files[0].FnTypes {
		if ft.Name == "f" {
			body = ft.Groups[0].Functions[0].Returns[0]
		}
	}
	if body == nil {
		t.Fatal("f not found")
	}
	if n := count(body, isValue(10)); n != 0 {
		t.Errorf("f's parameter d was read as the top-level d: %s", body.PrettyPrint(0))
	}
	if n := count(body, isReferenceTo("d")); n != 2 {
		t.Errorf("want both reads to read f's d, got %d: %s", n, body.PrettyPrint(0))
	}
}
