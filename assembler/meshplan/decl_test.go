package meshplan

import (
	"reflect"
	"testing"

	"gradientcontrol.com/omega/core/program"
)

func TestParseMeshArgs_Identity(t *testing.T) {
	fc := &program.FnCall{
		Arguments: []*program.Operand{
			program.NewValue(0, nil, nil),
		},
	}

	decl, err := parseMeshArgs(fc, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decl.Index != 0 {
		t.Fatalf("expected index 0, got %d", decl.Index)
	}
	if len(decl.Matrix) != 0 {
		t.Fatalf("expected empty matrix for identity, got %v", decl.Matrix)
	}
}

func TestParseMeshArgs_Translation(t *testing.T) {
	fc := &program.FnCall{
		Arguments: []*program.Operand{
			program.NewValue(2, nil, nil),
			{
				ConstValues: []float64{10, 20, 30},
			},
		},
	}

	decl, err := parseMeshArgs(fc, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decl.Index != 2 {
		t.Fatalf("expected index 2, got %d", decl.Index)
	}
	expected := []float64{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 10, 20, 30, 1}
	if !reflect.DeepEqual(decl.Matrix, expected) {
		t.Fatalf("expected %v, got %v", expected, decl.Matrix)
	}
}

func TestParseMeshArgs_Matrix(t *testing.T) {
	mat := []float64{
		2, 0, 0, 0,
		0, 2, 0, 0,
		0, 0, 2, 0,
		5, 6, 7, 1,
	}
	fc := &program.FnCall{
		Arguments: []*program.Operand{
			program.NewValue(1, nil, nil),
			&program.Operand{
				ConstValues: mat,
			},
		},
	}

	decl, err := parseMeshArgs(fc, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decl.Index != 1 {
		t.Fatalf("expected index 1, got %d", decl.Index)
	}
	if !reflect.DeepEqual(decl.Matrix, mat) {
		t.Fatalf("expected %v, got %v", mat, decl.Matrix)
	}
}

func TestParseMeshArgs_PosAndRot(t *testing.T) {
	rot := []float64{
		0, 1, 0, 0,
		-1, 0, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
	fc := &program.FnCall{
		Arguments: []*program.Operand{
			program.NewValue(3, nil, nil),
			&program.Operand{
				ConstValues: []float64{1, 2, 3},
			},
			&program.Operand{
				ConstValues: rot,
			},
		},
	}

	decl, err := parseMeshArgs(fc, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if decl.Index != 3 {
		t.Fatalf("expected index 3, got %d", decl.Index)
	}
	expected := []float64{
		0, 1, 0, 0,
		-1, 0, 0, 0,
		0, 0, 1, 0,
		1, 2, 3, 1,
	}
	if !reflect.DeepEqual(decl.Matrix, expected) {
		t.Fatalf("expected %v, got %v", expected, decl.Matrix)
	}
}

func TestParseMeshArgs_Errors(t *testing.T) {
	// Wrong number of values for placement
	fc := &program.FnCall{
		Arguments: []*program.Operand{
			program.NewValue(0, nil, nil),
			&program.Operand{
				ConstValues: []float64{1, 2}, // invalid length 2
			},
		},
	}
	if _, err := parseMeshArgs(fc, nil, nil); err == nil {
		t.Fatalf("expected error for invalid placement size")
	}

	// Extra argument for matrix placement
	fcExtra := &program.FnCall{
		Arguments: []*program.Operand{
			program.NewValue(0, nil, nil),
			&program.Operand{
				ConstValues: make([]float64, 16),
			},
			program.NewValue(42, nil, nil),
		},
	}
	if _, err := parseMeshArgs(fcExtra, nil, nil); err == nil {
		t.Fatalf("expected error for extra argument with 16-element matrix")
	}
}
