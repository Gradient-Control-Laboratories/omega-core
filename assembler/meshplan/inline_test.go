package meshplan

import (
	"testing"

	"gradientcontrol.com/omega/core/program"
)

func TestInlineFnCall_SingleReturn(t *testing.T) {
	// Function: mySphere(r) { return sphere(r); }
	fn := &program.Function{
		Parameters: []*program.Parameter{
			{Name: "r"},
		},
		Returns: []*program.Operand{
			program.NewFnCall(
				"", "sphere",
				[]*program.Operand{
					program.NewReference("", "r", nil, nil),
				},
				nil, nil,
			),
		},
	}
	fnType := &program.FnType{
		Name: "mySphere",
		Groups: []*program.FnGroup{
			{
				ParamCount: 1,
				Functions:  []*program.Function{fn},
			},
		},
	}
	pkg := &program.Package{
		Files: []*program.PackageFile{
			{
				FnTypes: []*program.FnType{fnType},
			},
		},
	}

	call := &program.FnCall{
		Reference: &program.Reference{Name: "mySphere"},
		Arguments: []*program.Operand{
			program.NewValue(10, nil, nil),
		},
	}

	inlined := inlineFnCall(call, pkg)
	if inlined == nil {
		t.Fatalf("expected inlined operand, got nil")
	}

	fc := inlined.GetFnCall()
	if fc == nil || fc.Reference == nil || fc.Reference.Name != "sphere" {
		t.Fatalf("expected sphere fncall, got %v", inlined)
	}
	if len(fc.Arguments) != 1 {
		t.Fatalf("expected 1 argument, got %d", len(fc.Arguments))
	}
	val, ok := fc.Arguments[0].T.(*program.Operand_Value)
	if !ok || val.Value != 10 {
		t.Fatalf("expected value 10, got %v", fc.Arguments[0].T)
	}
}

func TestInlineFnCall_NotInlineable(t *testing.T) {
	// Function with assignment cannot be inlined
	fn := &program.Function{
		Assignments: []*program.Assignment{
			{
				Variables:  []*program.Variable{{Name: "x"}},
				Expression: program.NewValue(1, nil, nil),
			},
		},
		Returns: []*program.Operand{
			program.NewReference("", "x", nil, nil),
		},
	}
	fnType := &program.FnType{
		Name: "fnWithAssign",
		Groups: []*program.FnGroup{
			{
				ParamCount: 0,
				Functions:  []*program.Function{fn},
			},
		},
	}
	pkg := &program.Package{
		Files: []*program.PackageFile{
			{
				FnTypes: []*program.FnType{fnType},
			},
		},
	}

	call := &program.FnCall{
		Reference: &program.Reference{Name: "fnWithAssign"},
	}

	if inlined := inlineFnCall(call, pkg); inlined != nil {
		t.Fatalf("expected nil for non-inlineable function, got %v", inlined)
	}
}
