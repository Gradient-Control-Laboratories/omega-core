package types

import (
	"errors"
	"fmt"
	"testing"

	parser "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

func init() {
	program.RegisterStdPackage("builtin", &program.Package{
		Path: "builtin",
		Files: []*program.PackageFile{
			{
				Path: "builtin.gsl",
			},
		},
	})
}

func parseProgram(t *testing.T, src string) *program.Program {
	t.Helper()
	res := parser.ParsePackage([]string{src})
	if err := res.Error(); err != nil {
		t.Fatalf("ParsePackage failed: %v", err)
	}
	return &program.Program{Primary: res.Package}
}

func val(v float64) *program.Operand {
	return &program.Operand{T: &program.Operand_Value{Value: v}}
}

func fnCall(name string, args ...*program.Operand) *program.Operand {
	return program.NewFnCall("", name, args, nil, nil)
}

func expr(op program.Operation, args ...*program.Operand) *program.Operand {
	return &program.Operand{T: &program.Operand_Expression{Expression: &program.Expression{Operation: op, Operands: args}}}
}

func vec(xs ...float64) *program.Operand {
	args := make([]*program.Operand, len(xs))
	for i, x := range xs {
		args[i] = val(x)
	}
	return expr(program.Operation_COLLECT, args...)
}

func TestAnalyze_MismatchedBinaryOps(t *testing.T) {
	// Invalid addition: vec2 + vec3
	invalidAdd := expr(program.Operation_ADD, vec(1.0, 2.0), vec(3.0, 4.0, 5.0))
	pkg := &program.Package{
		Files: []*program.PackageFile{
			{
				Assignments: []*program.Assignment{
					program.NewAssignment([]string{"x"}, invalidAdd, nil, nil),
				},
				Returns: []*program.Operand{val(1.0)},
			},
		},
	}
	prog := &program.Program{Primary: pkg}

	err := Analyze(prog)
	if err == nil {
		t.Fatal("expected type analysis error for vec2 + vec3, got nil")
	}
}

func TestAnalyze_ValidBroadcasting(t *testing.T) {
	// Valid addition: scalar + vec2 -> vec2
	validAdd := expr(program.Operation_ADD, val(1.0), vec(2.0, 3.0))
	pkg := &program.Package{
		Files: []*program.PackageFile{
			{
				Assignments: []*program.Assignment{
					program.NewAssignment([]string{"x"}, validAdd, nil, nil),
				},
				Returns: []*program.Operand{val(1.0)},
			},
		},
	}
	prog := &program.Program{Primary: pkg}

	err := Analyze(prog)
	if err != nil {
		t.Fatalf("unexpected type analysis error for broadcasting: %v", err)
	}

	if validAdd.Type == nil {
		t.Fatal("expected type to be populated on validAdd operand")
	}
	if validAdd.Type.Size != 2 || validAdd.Type.ElemType == nil || validAdd.Type.ElemType.Size != 1 {
		t.Errorf("expected type to be vec2, got size=%d", validAdd.Type.Size)
	}
}

func TestAnalyze_SwizzleBounds(t *testing.T) {
	// Invalid swizzle: vec2.z
	target := vec(1.0, 2.0)
	refOp := &program.Operand{T: &program.Operand_Reference{Reference: &program.Reference{Name: "z"}}}
	invalidAccess := expr(program.Operation_ACCESS, target, refOp)

	pkg := &program.Package{
		Files: []*program.PackageFile{
			{
				Assignments: []*program.Assignment{
					program.NewAssignment([]string{"x"}, invalidAccess, nil, nil),
				},
				Returns: []*program.Operand{val(1.0)},
			},
		},
	}
	prog := &program.Program{Primary: pkg}

	err := Analyze(prog)
	if err == nil {
		t.Fatal("expected swizzle out of bounds error, got nil")
	}
}

func TestAnalyze_MatrixVecMultiply(t *testing.T) {
	// Matrix = list of column vectors
	col1 := vec(1.0, 2.0)
	col2 := vec(3.0, 4.0)
	matrix := expr(program.Operation_COLLECT, col1, col2)

	vector := vec(5.0, 6.0)
	multiply := expr(program.Operation_MUL, matrix, vector)

	pkg := &program.Package{
		Files: []*program.PackageFile{
			{
				Assignments: []*program.Assignment{
					program.NewAssignment([]string{"x"}, multiply, nil, nil),
				},
				Returns: []*program.Operand{val(1.0)},
			},
		},
	}
	prog := &program.Program{Primary: pkg}

	err := Analyze(prog)
	if err != nil {
		t.Fatalf("unexpected error for matrix-vector multiplication: %v", err)
	}

	if multiply.Type == nil {
		t.Fatal("expected type to be populated on multiply operand")
	}
	if multiply.Type.Size != 2 || multiply.Type.ElemType == nil || multiply.Type.ElemType.Size != 1 {
		t.Errorf("expected result to be vec2, got size=%d", multiply.Type.Size)
	}
}

func TestAnalyze_Indexing(t *testing.T) {
	// Indexing: vec2[1]
	target := vec(10.0, 20.0)
	index := val(1.0)
	indexing := expr(program.Operation_INDEX, target, index)

	pkg := &program.Package{
		Files: []*program.PackageFile{
			{
				Assignments: []*program.Assignment{
					program.NewAssignment([]string{"x"}, indexing, nil, nil),
				},
				Returns: []*program.Operand{val(1.0)},
			},
		},
	}
	prog := &program.Program{Primary: pkg}

	err := Analyze(prog)
	if err != nil {
		t.Fatalf("unexpected error for indexing: %v", err)
	}

	if indexing.Type == nil {
		t.Fatal("expected type to be populated on indexing operand")
	}
	if indexing.Type.Size != 1 || indexing.Type.ElemType != nil {
		t.Errorf("expected indexing result to be scalar, got size=%d", indexing.Type.Size)
	}
}

// TestAnalyze_PropertyExpressionIsTyped locks the step that types property
// expressions. A list inside `name.color = …` went untyped because the
// analyzer only visited assignments and returns, so the constant fold stamped
// the list as flat values and the fidget backend rebuilt a list of rows as one
// long vector: indexing it gave a scalar, and material.withRoughness failed on
// the CPU renderer while the same call in an ordinary assignment worked.
func TestAnalyze_PropertyExpressionIsTyped(t *testing.T) {
	rows := expr(program.Operation_COLLECT, vec(0.1, 0.9, 0.1), vec(0.0, 0.0, 0.0))
	firstRow := expr(program.Operation_INDEX, rows, val(0.0))

	ball := program.NewAssignment([]string{"ball"}, val(1.0), nil, nil)
	ball.Variables[0].AddPropertyAssignment("color", firstRow, nil, nil)

	pkg := &program.Package{
		Files: []*program.PackageFile{
			{
				Assignments: []*program.Assignment{ball},
				Returns:     []*program.Operand{val(1.0)},
			},
		},
	}
	prog := &program.Program{Primary: pkg}

	if err := Analyze(prog); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rows.Type == nil {
		t.Fatal("the list inside a property expression was left untyped")
	}
	if rows.Type.Size != 2 || rows.Type.ElemType == nil || rows.Type.ElemType.Size != 3 {
		t.Errorf("expected a list of two vec3 rows, got size=%d elem=%v", rows.Type.Size, rows.Type.ElemType)
	}
	if firstRow.Type == nil || firstRow.Type.Size != 3 {
		t.Errorf("expected indexing the list to give a vec3 row, got %v", firstRow.Type)
	}
}

func TestAnalyze_MinMaxNotOperations(t *testing.T) {
	// min(vec2, vec2)
	minOp := expr(program.Operation_MIN, vec(10.0, 20.0), vec(5.0, 30.0))
	// max(vec3, scalar)
	maxOp := expr(program.Operation_MAX, vec(1.0, 2.0, 3.0), val(5.0))
	// not(vec2)
	notOp := expr(program.Operation_NOT, vec(1.0, 0.0))

	pkg := &program.Package{
		Files: []*program.PackageFile{
			{
				Assignments: []*program.Assignment{
					program.NewAssignment([]string{"a"}, minOp, nil, nil),
					program.NewAssignment([]string{"b"}, maxOp, nil, nil),
					program.NewAssignment([]string{"c"}, notOp, nil, nil),
				},
				Returns: []*program.Operand{val(1.0)},
			},
		},
	}
	prog := &program.Program{Primary: pkg}

	err := Analyze(prog)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if minOp.Type == nil {
		t.Fatal("expected type to be populated on minOp")
	}
	if minOp.Type.Size != 2 {
		t.Errorf("expected min(vec2, vec2) to return vec2, got size=%d", minOp.Type.Size)
	}

	if maxOp.Type == nil {
		t.Fatal("expected type to be populated on maxOp")
	}
	if maxOp.Type.Size != 3 {
		t.Errorf("expected max(vec3, scalar) to return vec3, got size=%d", maxOp.Type.Size)
	}

	if notOp.Type == nil {
		t.Fatal("expected type to be populated on notOp")
	}
	if notOp.Type.Size != 2 {
		t.Errorf("expected not(vec2) to return vec2, got size=%d", notOp.Type.Size)
	}
}

func TestAnalyze_SwizzleValid(t *testing.T) {
	// vec4.xyz
	target1 := vec(1.0, 2.0, 3.0, 4.0)
	refOp1 := &program.Operand{T: &program.Operand_Reference{Reference: &program.Reference{Name: "xyz"}}}
	swizzle1 := expr(program.Operation_ACCESS, target1, refOp1)

	// vec2.xxxx
	target2 := vec(5.0, 6.0)
	refOp2 := &program.Operand{T: &program.Operand_Reference{Reference: &program.Reference{Name: "xxxx"}}}
	swizzle2 := expr(program.Operation_ACCESS, target2, refOp2)

	pkg := &program.Package{
		Files: []*program.PackageFile{
			{
				Assignments: []*program.Assignment{
					program.NewAssignment([]string{"x"}, swizzle1, nil, nil),
					program.NewAssignment([]string{"y"}, swizzle2, nil, nil),
				},
				Returns: []*program.Operand{val(1.0)},
			},
		},
	}
	prog := &program.Program{Primary: pkg}

	err := Analyze(prog)
	if err != nil {
		t.Fatalf("unexpected swizzle error: %v", err)
	}

	if swizzle1.Type == nil || swizzle1.Type.Size != 3 {
		t.Errorf("expected vec4.xyz to type as vec3, got size=%d", swizzle1.Type.Size)
	}
	if swizzle2.Type == nil || swizzle2.Type.Size != 4 {
		t.Errorf("expected vec2.xxxx to type as vec4, got size=%d", swizzle2.Type.Size)
	}
}

func TestAnalyze_SelectValid(t *testing.T) {
	// cond ? vec2 : vec2
	cond := val(1.0)
	branches := expr(program.Operation_SELECT, cond, vec(1.0, 2.0), vec(3.0, 4.0))

	pkg := &program.Package{
		Files: []*program.PackageFile{
			{
				Assignments: []*program.Assignment{
					program.NewAssignment([]string{"x"}, branches, nil, nil),
				},
				Returns: []*program.Operand{val(1.0)},
			},
		},
	}
	prog := &program.Program{Primary: pkg}

	err := Analyze(prog)
	if err != nil {
		t.Fatalf("unexpected select error: %v", err)
	}

	if branches.Type == nil || branches.Type.Size != 2 {
		t.Errorf("expected ternary select to type as vec2, got size=%d", branches.Type.Size)
	}
}

func TestAnalyze_SelectMismatched(t *testing.T) {
	// cond ? vec2 : vec3
	cond := val(1.0)
	branches := expr(program.Operation_SELECT, cond, vec(1.0, 2.0), vec(3.0, 4.0, 5.0))

	pkg := &program.Package{
		Files: []*program.PackageFile{
			{
				Assignments: []*program.Assignment{
					program.NewAssignment([]string{"x"}, branches, nil, nil),
				},
				Returns: []*program.Operand{val(1.0)},
			},
		},
	}
	prog := &program.Program{Primary: pkg}

	err := Analyze(prog)
	if err == nil {
		t.Fatal("expected select type mismatch error, got nil")
	}
}

func TestAnalyze_NamedPrimaryPackage(t *testing.T) {
	// Function declared in named primary package
	fn := &program.Function{
		Parameters: []*program.Parameter{
			{Name: "a"},
		},
		Returns: []*program.Operand{
			expr(program.Operation_ADD, &program.Operand{T: &program.Operand_Reference{Reference: &program.Reference{Name: "a"}}}, val(1.0)),
		},
	}
	fnType := &program.FnType{
		Name: "my_fn",
		Groups: []*program.FnGroup{
			{
				ParamCount: 1,
				Functions:  []*program.Function{fn},
			},
		},
	}

	call := &program.Operand{
		T: &program.Operand_FnCall{
			FnCall: &program.FnCall{
				Reference: &program.Reference{Name: "my_fn"},
				Arguments: []*program.Operand{val(2.0)},
			},
		},
	}

	pkg := &program.Package{
		Path: "my/custom/module",
		Files: []*program.PackageFile{
			{
				FnTypes: []*program.FnType{fnType},
				Assignments: []*program.Assignment{
					program.NewAssignment([]string{"y"}, call, nil, nil),
				},
				Returns: []*program.Operand{val(1.0)},
			},
		},
	}
	prog := &program.Program{Primary: pkg}

	err := Analyze(prog)
	if err != nil {
		t.Fatalf("Analyze failed for named primary package: %v", err)
	}

	if prog.Primary.Path != "my/custom/module" {
		t.Errorf("expected primary package path preserved as 'my/custom/module', got %q", prog.Primary.Path)
	}

	if call.Type == nil || call.Type.Size != 1 {
		t.Errorf("expected call return type scalar, got %v", call.Type)
	}
}

func TestAnalyze_SingleElementList(t *testing.T) {
	src := `
		s = [5.0];
		idx0 = s[0];
		idx1 = s[1];
		sum = s + [2.0];
		prod = s * 3.0;
		xVal = s.x;
		xxVal = s.xx;
		cmp = s == [2.0];
		return 1.0;
	`
	prog := parseProgram(t, src)
	if err := Analyze(prog); err != nil {
		t.Fatalf("Analyze failed for single-element list expressions: %v", err)
	}

	assigns := make(map[string]*program.Type)
	for _, a := range prog.Primary.Files[0].Assignments {
		if len(a.Variables) > 0 {
			assigns[a.Variables[0].Name] = a.Variables[0].Type
		}
	}

	// Verify types
	if tVal := assigns["s"]; !tVal.IsList() || tVal.Size != 1 || !tVal.ElemType.IsScalar() {
		t.Errorf("expected s to be [1]float64, got %s", tVal.TypeString())
	}
	if tVal := assigns["idx0"]; !tVal.IsScalar() {
		t.Errorf("expected idx0 to be scalar, got %s", tVal.TypeString())
	}
	if tVal := assigns["idx1"]; !tVal.IsScalar() {
		t.Errorf("expected idx1 to be scalar, got %s", tVal.TypeString())
	}
	if tVal := assigns["sum"]; !tVal.IsList() || tVal.Size != 1 {
		t.Errorf("expected sum to be [1]float64, got %s", tVal.TypeString())
	}
	if tVal := assigns["prod"]; !tVal.IsList() || tVal.Size != 1 {
		t.Errorf("expected prod to be [1]float64, got %s", tVal.TypeString())
	}
	if tVal := assigns["xVal"]; !tVal.IsScalar() {
		t.Errorf("expected xVal to be scalar, got %s", tVal.TypeString())
	}
	if tVal := assigns["xxVal"]; !tVal.IsList() || tVal.Size != 2 {
		t.Errorf("expected xxVal to be [2]float64, got %s", tVal.TypeString())
	}
	if tVal := assigns["cmp"]; !tVal.IsList() || tVal.Size != 1 {
		t.Errorf("expected cmp to be [1]float64, got %s", tVal.TypeString())
	}

	// Test invalid swizzle .y on 1-element list
	errProg1 := parseProgram(t, "s = [5.0]; return s.y;")
	if err := Analyze(errProg1); err == nil {
		t.Fatal("expected error for s.y on [1]float64, got nil")
	}

	// Test invalid comparison between list and scalar
	errProg2 := parseProgram(t, "s = [5.0]; return s == 5.0;")
	if err := Analyze(errProg2); err == nil {
		t.Fatal("expected error comparing list with scalar, got nil")
	}
}

func TestAnalyze_CrossAndBuiltinVec3(t *testing.T) {
	v3 := vec(1.0, 2.0, 3.0)
	s := val(1.0)
	single := vec(5.0)
	v2 := vec(1.0, 2.0)
	col1 := vec(1.0, 0.0, 0.0)
	col2 := vec(0.0, 1.0, 0.0)
	col3 := vec(0.0, 0.0, 1.0)
	mat3 := expr(program.Operation_COLLECT, col1, col2, col3)

	runAnalyze := func(exprs ...*program.Operand) error {
		var assigns []*program.Assignment
		for i, e := range exprs {
			assigns = append(assigns, program.NewAssignment([]string{fmt.Sprintf("v%d", i)}, e, nil, nil))
		}
		pkg := &program.Package{
			Files: []*program.PackageFile{
				{
					Assignments: assigns,
					Returns:     []*program.Operand{val(1.0)},
				},
			},
		}
		return Analyze(&program.Program{Primary: pkg})
	}

	// Valid cases
	validCross1 := fnCall("cross", v3, v3)
	validCross2 := fnCall("cross", v3, s)
	validUIPoint1 := fnCall("uiPoint", v3)
	validUIPoint2 := fnCall("uiPoint", s)
	if err := runAnalyze(validCross1, validCross2, validUIPoint1, validUIPoint2); err != nil {
		t.Fatalf("expected valid cross and uiPoint calls to pass, got: %v", err)
	}

	// Invalid cross with 1-element list
	if err := runAnalyze(fnCall("cross", single, v3)); err == nil {
		t.Fatal("expected error for cross([5.0], vec3), got nil")
	}

	// Invalid cross with vec2
	if err := runAnalyze(fnCall("cross", v2, v3)); err == nil {
		t.Fatal("expected error for cross(vec2, vec3), got nil")
	}

	// Invalid cross with 3x3 matrix
	if err := runAnalyze(mat3, fnCall("cross", mat3, mat3)); err == nil {
		t.Fatal("expected error for cross(mat3x3, mat3x3), got nil")
	}

	// Invalid uiPoint with 1-element list
	if err := runAnalyze(fnCall("uiPoint", single)); err == nil {
		t.Fatal("expected error for uiPoint([5.0]), got nil")
	}

	// Invalid uiPoint with 3x3 matrix
	if err := runAnalyze(mat3, fnCall("uiPoint", mat3)); err == nil {
		t.Fatal("expected error for uiPoint(mat3x3), got nil")
	}
}

func TestAnalyze_InternalErrors(t *testing.T) {
	// 1. Unspecified operation returns *program.InternalError
	unspecifiedExpr := program.NewExpression(program.Operation_UNSPECIFIED_OPERAND, nil, nil, nil)
	prog1 := &program.Program{
		Primary: &program.Package{
			Files: []*program.PackageFile{
				{
					Returns: []*program.Operand{unspecifiedExpr},
				},
			},
		},
	}
	err1 := Analyze(prog1)
	if err1 == nil {
		t.Fatal("expected error for UNSPECIFIED_OPERAND, got nil")
	}
	var internalErr1 *program.InternalError
	if !errors.As(err1, &internalErr1) {
		t.Fatalf("expected *program.InternalError, got %T (%v)", err1, err1)
	}

	// 2. Unexpanded Operation_DERIV returns *program.InternalError
	derivExpr := program.NewExpression(program.Operation_DERIV, []*program.Operand{val(1.0)}, nil, nil)
	prog2 := &program.Program{
		Primary: &program.Package{
			Files: []*program.PackageFile{
				{
					Returns: []*program.Operand{derivExpr},
				},
			},
		},
	}
	err2 := Analyze(prog2)
	if err2 == nil {
		t.Fatal("expected error for unexpanded DERIV, got nil")
	}
	var internalErr2 *program.InternalError
	if !errors.As(err2, &internalErr2) {
		t.Fatalf("expected *program.InternalError, got %T (%v)", err2, err2)
	}
}
