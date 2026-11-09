package property_test

import (
	"fmt"
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/internal/assembler/property"
	"gradientcontrol.com/omega/core/program"
)

// mockValue wraps a string representation for testing the abstract engine.
type mockValue string

// testDomain is a simple Domain implementation for testing Engine mechanisms.
type testDomain struct{}

func (d testDomain) Literal(v float64) property.Value {
	return mockValue(fmt.Sprintf("%g", v))
}

func (d testDomain) ImpliedVar(name string) (property.Value, error) {
	return mockValue(name), nil
}

func (d testDomain) Opaque(name string) property.Value {
	return mockValue("opaque(" + name + ")")
}

func (d testDomain) Op(op program.Operation, args []property.Value) (property.Value, error) {
	strArgs := make([]string, len(args))
	for i, a := range args {
		strArgs[i] = string(a.(mockValue))
	}
	return mockValue(fmt.Sprintf("%s(%s)", op, strings.Join(strArgs, ", "))), nil
}

func (d testDomain) Builtin(name string, args []property.Value) (property.Value, error) {
	strArgs := make([]string, len(args))
	for i, a := range args {
		strArgs[i] = string(a.(mockValue))
	}
	return mockValue(fmt.Sprintf("%s(%s)", name, strings.Join(strArgs, ", "))), nil
}

func (d testDomain) Access(base property.Value, members []*program.Operand) property.Value {
	return mockValue(fmt.Sprintf("%s.access", base))
}

func (d testDomain) Index(base property.Value, index *program.Operand) property.Value {
	return mockValue(fmt.Sprintf("%s.index", base))
}

func (d testDomain) Collect(args []property.Value) property.Value {
	return mockValue("collect")
}

func (d testDomain) Select(cond property.Value, cases []property.Value) property.Value {
	return mockValue("select")
}

func (d testDomain) Named(name string, v property.Value) property.Value {
	return v
}

func (d testDomain) Operator(op program.Operation, operands []*program.Operand) (*program.Operand, bool, error) {
	return nil, false, nil
}

func TestEngine_Literal(t *testing.T) {
	eng := property.NewEngine(testDomain{}, nil, nil)
	op := &program.Operand{T: &program.Operand_Value{Value: 3.14}}

	val, err := eng.Eval(op)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val.(mockValue)) != "3.14" {
		t.Fatalf("expected '3.14', got %q", val)
	}
}

func TestEngine_CyclicReference(t *testing.T) {
	// Setup cyclic declarations: x = y, y = x
	decls := map[string]*program.Operand{
		"x": {T: &program.Operand_Reference{Reference: &program.Reference{Name: "y"}}},
		"y": {T: &program.Operand_Reference{Reference: &program.Reference{Name: "x"}}},
	}
	eng := property.NewEngine(testDomain{}, decls, nil)
	op := &program.Operand{T: &program.Operand_Reference{Reference: &program.Reference{Name: "x"}}}

	_, err := eng.Eval(op)
	if err == nil {
		t.Fatalf("expected cyclic reference error, got nil")
	}
	if !strings.Contains(err.Error(), "cyclic reference through") {
		t.Fatalf("expected error message containing 'cyclic reference through', got %q", err.Error())
	}
}

func TestEngine_FunctionInlining(t *testing.T) {
	funcs := map[string][]*program.Function{
		"double": {
			{
				Parameters: []*program.Parameter{{Name: "v"}},
				Returns: []*program.Operand{
					{
						T: &program.Operand_Expression{
							Expression: &program.Expression{
								Operation: program.Operation_ADD,
								Operands: []*program.Operand{
									{T: &program.Operand_Reference{Reference: &program.Reference{Name: "v"}}},
									{T: &program.Operand_Reference{Reference: &program.Reference{Name: "v"}}},
								},
							},
						},
					},
				},
			},
		},
	}
	eng := property.NewEngine(testDomain{}, nil, funcs)
	op := &program.Operand{
		T: &program.Operand_FnCall{
			FnCall: &program.FnCall{
				Reference: &program.Reference{Name: "double"},
				Arguments: []*program.Operand{
					{T: &program.Operand_Value{Value: 5.0}},
				},
			},
		},
	}

	val, err := eng.Eval(op)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "ADD(5, 5)"
	if string(val.(mockValue)) != expected {
		t.Fatalf("expected %q, got %q", expected, val)
	}
}
