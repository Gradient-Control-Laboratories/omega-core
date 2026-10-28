package build

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/program"
)

type dummyDelegate struct {
	Delegate
	scope *Scope
}

func (d *dummyDelegate) GetScope() *Scope {
	return d.scope
}

func TestVisitExpression_DERIV_Error(t *testing.T) {
	b := &Builder{}
	scope := NewScope(b, "root", nil, nil, nil)
	b.delegate = &dummyDelegate{scope: scope}

	op := &program.Operand{
		Start: &program.Location{Line: 1, Column: 1},
		End:   &program.Location{Line: 1, Column: 10},
	}
	expr := &program.Expression{
		Operation: program.Operation_DERIV,
	}

	_, err := b.VisitExpression(op, expr)
	if err == nil {
		t.Fatalf("expected error when visiting unexpanded DERIV operation, got nil")
	}

	expectedMsg := "the .dtime/.dp derivative operators are only supported in top-level expressions, not inside function bodies"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Fatalf("expected error message containing %q, got %q", expectedMsg, err.Error())
	}
}
