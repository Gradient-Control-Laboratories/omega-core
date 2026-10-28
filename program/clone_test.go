package program_test

import (
	"testing"

	"gradientcontrol.com/omega/core/program"
)

func TestCloneOperand_Value(t *testing.T) {
	locStart := program.NewLoc(1, 1, 0)
	locEnd := program.NewLoc(1, 5, 4)
	valOp := program.NewValue(42.5, locStart, locEnd)
	valOp.ConstValues = []float64{42.5}
	valOp.Type = program.NewScalarType()

	cloned := program.CloneOperand(valOp)
	if cloned == nil {
		t.Fatalf("expected non-nil cloned operand")
	}
	if cloned == valOp {
		t.Fatalf("expected cloned pointer to be distinct from original")
	}
	if cloned.GetValue() != 42.5 {
		t.Errorf("expected value 42.5, got %v", cloned.GetValue())
	}
	if cloned.Start.Line != 1 || cloned.End.Column != 5 {
		t.Errorf("expected location to be preserved")
	}
	if len(cloned.ConstValues) != 1 || cloned.ConstValues[0] != 42.5 {
		t.Errorf("expected ConstValues to be cloned")
	}
	// Mutate original ConstValues and ensure clone is unchanged
	valOp.ConstValues[0] = 999.0
	if cloned.ConstValues[0] != 42.5 {
		t.Errorf("cloned ConstValues was mutated when original changed")
	}
}

func TestCloneOperand_Expression(t *testing.T) {
	op1 := program.NewValue(10, nil, nil)
	op2 := program.NewReference("", "x", nil, nil)
	exprOp := program.NewExpression(program.Operation_ADD, []*program.Operand{op1, op2}, nil, nil)
	exprOp.Type = program.NewScalarType()

	cloned := program.CloneOperand(exprOp)
	if cloned == nil {
		t.Fatalf("expected non-nil cloned expression")
	}
	expr := cloned.GetExpression()
	if expr == nil {
		t.Fatalf("expected cloned expression")
	}
	if expr.Operation != program.Operation_ADD {
		t.Errorf("expected Operation_ADD, got %v", expr.Operation)
	}
	if len(expr.Operands) != 2 {
		t.Fatalf("expected 2 operands, got %d", len(expr.Operands))
	}
	if expr.Operands[0] == op1 || expr.Operands[1] == op2 {
		t.Errorf("expected sub-operands to be deep-cloned")
	}
	if expr.Operands[0].GetValue() != 10 {
		t.Errorf("expected operand 0 value 10, got %v", expr.Operands[0].GetValue())
	}
	if expr.Operands[1].GetReference().Name != "x" {
		t.Errorf("expected operand 1 reference 'x', got %v", expr.Operands[1].GetReference().Name)
	}
}

func TestCloneOperand_FnCall(t *testing.T) {
	arg := program.NewValue(3.14, nil, nil)
	fnCallOp := program.NewFnCall("math", "sin", []*program.Operand{arg}, nil, nil)

	cloned := program.CloneOperand(fnCallOp)
	if cloned == nil {
		t.Fatalf("expected non-nil cloned fnCall")
	}
	fc := cloned.GetFnCall()
	if fc == nil {
		t.Fatalf("expected cloned FnCall")
	}
	if fc.Reference.Prefix != "math" || fc.Reference.Name != "sin" {
		t.Errorf("expected math.sin, got %s.%s", fc.Reference.Prefix, fc.Reference.Name)
	}
	if len(fc.Arguments) != 1 {
		t.Fatalf("expected 1 argument, got %d", len(fc.Arguments))
	}
	if fc.Arguments[0] == arg {
		t.Errorf("expected argument operand to be deep-cloned")
	}
}

func TestCloneOperand_Nil(t *testing.T) {
	if program.CloneOperand(nil) != nil {
		t.Errorf("expected CloneOperand(nil) to return nil")
	}
}

func TestCloneRangeAnnotation(t *testing.T) {
	r := &program.Range{
		Start:  program.NewLoc(1, 1, 0),
		End:    program.NewLoc(1, 15, 14),
		Min:    program.NewValue(1.0, nil, nil),
		Max:    program.NewValue(10.0, nil, nil),
		Period: program.NewValue(2.0, nil, nil),
	}

	cloned := program.CloneRangeAnnotation(r)
	if cloned == nil {
		t.Fatalf("expected non-nil cloned range")
	}
	if cloned == r {
		t.Fatalf("expected cloned range to be distinct pointer")
	}
	if cloned.Start == r.Start || cloned.End == r.End {
		t.Fatalf("expected location pointers to be cloned")
	}
	if cloned.Min == r.Min || cloned.Max == r.Max || cloned.Period == r.Period {
		t.Fatalf("expected operand pointers to be cloned")
	}
	if minVal, _ := cloned.MinConst(); minVal != 1.0 {
		t.Errorf("expected MinConst 1.0, got %v", minVal)
	}

	if program.CloneRangeAnnotation(nil) != nil {
		t.Errorf("expected CloneRangeAnnotation(nil) to return nil")
	}
}

func TestCloneParameter(t *testing.T) {
	param := &program.Parameter{
		Start: program.NewLoc(2, 5, 20),
		Name:  "radius",
		Type:  program.NewScalarType(),
		Range: &program.Range{
			Start: program.NewLoc(2, 1, 16),
			End:   program.NewLoc(2, 4, 19),
			Min:   program.NewValue(0.5, nil, nil),
			Max:   program.NewValue(5.0, nil, nil),
		},
	}

	cloned := program.CloneParameter(param)
	if cloned == nil {
		t.Fatalf("expected non-nil cloned parameter")
	}
	if cloned == param {
		t.Fatalf("expected distinct parameter pointer")
	}
	if cloned.Start == param.Start {
		t.Errorf("expected Start location to be cloned")
	}
	if cloned.Name != "radius" {
		t.Errorf("expected name 'radius', got %s", cloned.Name)
	}
	if cloned.Range == param.Range || cloned.Range == nil {
		t.Errorf("expected RangeAnnotation to be deep cloned")
	}

	if program.CloneParameter(nil) != nil {
		t.Errorf("expected CloneParameter(nil) to return nil")
	}
}

func TestCloneAssignment_WithRange(t *testing.T) {
	assign := &program.Assignment{
		Start:      program.NewLoc(1, 1, 0),
		End:        program.NewLoc(1, 7, 6),
		Variables:  []*program.Variable{{Name: "radius"}},
		Expression: program.NewValue(5.0, nil, nil),
		Range: &program.Range{
			Start: program.NewLoc(1, 1, 0),
			End:   program.NewLoc(1, 10, 9),
			Min:   program.NewValue(1.0, nil, nil),
			Max:   program.NewValue(10.0, nil, nil),
		},
	}

	cloned := program.CloneAssignment(assign)
	if cloned == nil {
		t.Fatalf("expected non-nil cloned assignment")
	}
	if cloned == assign {
		t.Fatalf("expected distinct assignment pointer")
	}
	if cloned.Range == assign.Range || cloned.Range == nil {
		t.Errorf("expected Range to be deep cloned")
	}
	if minVal, _ := cloned.Range.MinConst(); minVal != 1.0 {
		t.Errorf("expected cloned range MinConst 1.0, got %v", minVal)
	}

	if program.CloneAssignment(nil) != nil {
		t.Errorf("expected CloneAssignment(nil) to return nil")
	}
}
