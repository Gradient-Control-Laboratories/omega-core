package eval

import (
	"fmt"
	"math"
	"testing"

	"gradientcontrol.com/omega/core/program"
)

func val(v float64) *program.Operand {
	return &program.Operand{T: &program.Operand_Value{Value: v}}
}

func ref(name string) *program.Operand {
	return &program.Operand{T: &program.Operand_Reference{Reference: &program.Reference{Name: name}}}
}

func qref(prefix, name string) *program.Operand {
	return &program.Operand{T: &program.Operand_Reference{Reference: &program.Reference{Prefix: prefix, Name: name}}}
}

func fc(name string, args ...*program.Operand) *program.Operand {
	return &program.Operand{T: &program.Operand_FnCall{
		FnCall: &program.FnCall{
			Reference: &program.Reference{Name: name},
			Arguments: args,
		}}}
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

func assign(name string, body *program.Operand) *program.Assignment {
	return program.NewAssignment([]string{name}, body, nil, nil)
}

func approxEqual(a, b []float64) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if math.Abs(a[i]-b[i]) > 1e-9 {
			return false
		}
	}
	return true
}

func TestConstLiterals(t *testing.T) {
	got, err := Const(val(3.5), nil, nil)
	if err != nil || !approxEqual(got, []float64{3.5}) {
		t.Fatalf("scalar literal: got %v, err %v", got, err)
	}
	got, err = Const(vec(1, 2, 3), nil, nil)
	if err != nil || !approxEqual(got, []float64{1, 2, 3}) {
		t.Fatalf("vector literal: got %v, err %v", got, err)
	}
}

func TestConstArithmetic(t *testing.T) {
	cases := []struct {
		name string
		op   *program.Operand
		want []float64
	}{
		{"add", expr(program.Operation_ADD, val(1), val(2), val(3)), []float64{6}},
		{"sub", expr(program.Operation_SUB, val(10), val(3), val(2)), []float64{5}},
		{"mul", expr(program.Operation_MUL, val(2), val(3), val(4)), []float64{24}},
		{"div", expr(program.Operation_DIV, val(24), val(2), val(3)), []float64{4}},
		{"neg", expr(program.Operation_NEG, val(5)), []float64{-5}},
		{"sqrt", expr(program.Operation_SQRT, val(9)), []float64{3}},
		{"square", expr(program.Operation_SQUARE, val(4)), []float64{16}},
		{"vec+scalar", expr(program.Operation_ADD, vec(1, 2, 3), val(10)), []float64{11, 12, 13}},
		{"vec+vec", expr(program.Operation_ADD, vec(1, 2, 3), vec(10, 20, 30)), []float64{11, 22, 33}},
		{"single_list", vec(5), []float64{5}},
		{"single_list_index", expr(program.Operation_INDEX, vec(5), val(0)), []float64{5}},
		{"single_list_neg", expr(program.Operation_NEG, vec(5)), []float64{-5}},
		{"single_list_abs", expr(program.Operation_ABS, vec(-5)), []float64{5}},
		{"single_list_sqrt", expr(program.Operation_SQRT, vec(9)), []float64{3}},
		{"single_list_sin", expr(program.Operation_SIN, vec(0)), []float64{0}},
		{"single_list+scalar", expr(program.Operation_ADD, vec(5), val(10)), []float64{15}},
		{"single_list+single_list", expr(program.Operation_ADD, vec(5), vec(10)), []float64{15}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Const(tc.op, nil, nil)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if !approxEqual(got, tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestConstBuiltinCalls(t *testing.T) {
	cases := []struct {
		name string
		op   *program.Operand
		want []float64
	}{
		{"radians(90)", fc("radians", val(90)), []float64{math.Pi / 2}},
		{"degrees(pi)", fc("degrees", val(math.Pi)), []float64{180}},
		{"length(3,4)", fc("length", vec(3, 4)), []float64{5}},
		{"length([5])", fc("length", vec(5)), []float64{5}},
		{"normalize([5])", fc("normalize", vec(5)), []float64{1}},
		{"normalize([-5])", fc("normalize", vec(-5)), []float64{-1}},
		{"dot([1,2,3],[4,5,6])", fc("dot", vec(1, 2, 3), vec(4, 5, 6)), []float64{32}},
		{"dot([2],[3])", fc("dot", vec(2), vec(3)), []float64{6}},
		{"cross x,y", fc("cross", vec(1, 0, 0), vec(0, 1, 0)), []float64{0, 0, 1}},
		{"clamp(5, 0, 3)", fc("clamp", val(5), val(0), val(3)), []float64{3}},
		{"distance", fc("distance", vec(0, 0, 0), vec(3, 4, 0)), []float64{5}},
		{"distance([2],[5])", fc("distance", vec(2), vec(5)), []float64{3}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Const(tc.op, nil, nil)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if !approxEqual(got, tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestConstReferences(t *testing.T) {
	assignments := []*program.Assignment{
		assign("a", val(10)),
		assign("b", expr(program.Operation_MUL, val(2), ref("a"))), // 20
		assign("angle", fc("radians", val(90))),                    // pi/2
		assign("rot", vec(0, 0, 1)),                                // [0,0,1]
		assign("rotScaled", expr(program.Operation_MUL, ref("rot"), val(2))),
	}

	cases := []struct {
		name string
		op   *program.Operand
		want []float64
	}{
		{"a", ref("a"), []float64{10}},
		{"b", ref("b"), []float64{20}},
		{"angle", ref("angle"), []float64{math.Pi / 2}},
		{"rot", ref("rot"), []float64{0, 0, 1}},
		{"rotScaled", ref("rotScaled"), []float64{0, 0, 2}},
		{"a + b", expr(program.Operation_ADD, ref("a"), ref("b")), []float64{30}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Const(tc.op, assignments, nil)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if !approxEqual(got, tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func TestConstRejects(t *testing.T) {
	cases := []struct {
		name        string
		op          *program.Operand
		assignments []*program.Assignment
	}{
		{"undefined ref", ref("missing"), nil},
		{"p reference", ref("p"), nil},
		{"user function", fc("myFunc", val(1)), nil},
		{"cyclic reference", ref("a"), []*program.Assignment{assign("a", ref("a"))}},
		{"cross non-vec3", fc("cross", vec(1, 2), vec(3, 4)), nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Const(tc.op, tc.assignments, nil); err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}

func TestConstSelect(t *testing.T) {
	// 2-way select
	gotTrue, err := Const(expr(program.Operation_SELECT, val(1), val(10), val(20)), nil, nil)
	if err != nil || !approxEqual(gotTrue, []float64{10}) {
		t.Fatalf("select true: got %v, err %v", gotTrue, err)
	}
	gotFalse, err := Const(expr(program.Operation_SELECT, val(0), val(10), val(20)), nil, nil)
	if err != nil || !approxEqual(gotFalse, []float64{20}) {
		t.Fatalf("select false: got %v, err %v", gotFalse, err)
	}

	// Vector select
	gotVec, err := Const(expr(program.Operation_SELECT, vec(1, 0), vec(10, 20), vec(30, 40)), nil, nil)
	if err != nil || !approxEqual(gotVec, []float64{10, 40}) {
		t.Fatalf("vector select: got %v, err %v", gotVec, err)
	}
}

func pkg(imports []*program.Import, assignments []*program.Assignment, fnTypes []*program.FnType) *program.Package {
	return &program.Package{
		Files: []*program.PackageFile{
			{
				Imports:     imports,
				Assignments: assignments,
				FnTypes:     fnTypes,
			},
		},
	}
}

func dotImport(path string) *program.Import {
	return &program.Import{Prefix: ".", Path: path}
}

// mockImports returns a GetImportFunc backed by a map, counting loads so
// tests can assert laziness and memoization.
func mockImports(pkgs map[string]*program.Package, loads map[string]int) program.GetImportFunc {
	return func(path string) (*program.Package, error) {
		if loads != nil {
			loads[path] = loads[path] + 1
		}
		p, ok := pkgs[path]
		if !ok {
			return nil, fmt.Errorf("package not found")
		}
		return p, nil
	}
}

func TestConstPkgImports(t *testing.T) {
	omega := pkg(nil, []*program.Assignment{
		assign("pi", expr(program.Operation_MUL, expr(program.Operation_ATAN, val(1)), val(4))),
		assign("pi2", expr(program.Operation_MUL, ref("pi"), val(2))),
	}, nil)
	getImport := mockImports(map[string]*program.Package{"omega": omega}, nil)

	cases := []struct {
		name string
		op   *program.Operand
		pkg  *program.Package
		want []float64
	}{
		{"dot-imported const", ref("pi"), pkg([]*program.Import{dotImport("omega")}, nil, nil), []float64{math.Pi}},
		{"imported const referencing its own package", ref("pi2"),
			pkg([]*program.Import{dotImport("omega")}, nil, nil), []float64{2 * math.Pi}},
		{"expression over imported const", expr(program.Operation_DIV, ref("pi"), val(2)),
			pkg([]*program.Import{dotImport("omega")}, nil, nil), []float64{math.Pi / 2}},
		{"vector with imported const", vec3Op(ref("pi"), expr(program.Operation_DIV, ref("pi"), val(2)), val(1)),
			pkg([]*program.Import{dotImport("omega")}, nil, nil), []float64{math.Pi, math.Pi / 2, 1}},
		{"local shadows imported", ref("pi"),
			pkg([]*program.Import{dotImport("omega")}, []*program.Assignment{assign("pi", val(3))}, nil), []float64{3}},
		{"prefixed import qreference", qref("omega", "pi"),
			pkg([]*program.Import{{Path: "omega"}}, nil, nil), []float64{math.Pi}},
		{"aliased import qreference", qref("o", "pi"),
			pkg([]*program.Import{{Prefix: "o", Path: "omega"}}, nil, nil), []float64{math.Pi}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ConstPkg(tc.op, tc.pkg, getImport)
			if err != nil {
				t.Fatalf("err: %v", err)
			}
			if !approxEqual(got, tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

func vec3Op(xs ...*program.Operand) *program.Operand {
	return expr(program.Operation_COLLECT, xs...)
}

func TestConstPkgImportLazyAndMemoized(t *testing.T) {
	loads := map[string]int{}
	omega := pkg(nil, []*program.Assignment{assign("pi", val(3.14))}, nil)
	getImport := mockImports(map[string]*program.Package{"omega": omega}, loads)
	module := pkg([]*program.Import{dotImport("omega")}, []*program.Assignment{assign("local", val(5))}, nil)

	// A constant that never touches the import must not load it.
	if _, err := ConstPkg(ref("local"), module, getImport); err != nil {
		t.Fatalf("err: %v", err)
	}
	if loads["omega"] != 0 {
		t.Fatalf("expected lazy import load, got %d loads", loads["omega"])
	}

	// Two references through the import in one evaluation load it once.
	op := expr(program.Operation_ADD, ref("pi"), ref("pi"))
	if _, err := ConstPkg(op, module, getImport); err != nil {
		t.Fatalf("err: %v", err)
	}
	if loads["omega"] != 1 {
		t.Fatalf("expected memoized import load, got %d loads", loads["omega"])
	}
}

func TestConstPkgImportErrors(t *testing.T) {
	omega := pkg(nil, []*program.Assignment{assign("pi", val(3.14))}, nil)
	getImport := mockImports(map[string]*program.Package{"omega": omega}, nil)

	cases := []struct {
		name      string
		op        *program.Operand
		pkg       *program.Package
		getImport program.GetImportFunc
	}{
		{"unknown name across imports", ref("tau"),
			pkg([]*program.Import{dotImport("omega")}, nil, nil), getImport},
		{"unknown prefix", qref("other", "pi"),
			pkg([]*program.Import{{Path: "omega"}}, nil, nil), getImport},
		{"missing package", ref("pi"),
			pkg([]*program.Import{dotImport("missing")}, nil, nil), getImport},
		{"nil getImport with dot import", ref("pi"),
			pkg([]*program.Import{dotImport("omega")}, nil, nil), nil},
		{"cyclic reference across packages", ref("a"),
			pkg([]*program.Import{dotImport("cyc")}, nil, nil), mockImports(map[string]*program.Package{
				"cyc": pkg(nil, []*program.Assignment{assign("a", ref("a"))}, nil),
			}, nil)},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ConstPkg(tc.op, tc.pkg, tc.getImport); err == nil {
				t.Fatalf("expected error, got nil")
			}
		})
	}
}

func fnTypeDecl(name string, params []string, locals []*program.Assignment, ret *program.Operand) *program.FnType {
	fnParams := make([]*program.Parameter, len(params))
	for i, p := range params {
		fnParams[i] = &program.Parameter{Name: p}
	}
	return &program.FnType{
		Name: name,
		Groups: []*program.FnGroup{
			{
				ParamCount: int32(len(params)),
				Functions: []*program.Function{
					{
						Parameters:  fnParams,
						Assignments: locals,
						Returns:     []*program.Operand{ret},
					},
				},
			},
		},
	}
}

func acc(target *program.Operand, component int) *program.Operand {
	return expr(program.Operation_ACCESS, target, val(float64(component)))
}

func rotateXYDecl() *program.FnType {
	return fnTypeDecl("rotateXY", []string{"v", "ang"},
		[]*program.Assignment{
			assign("ca", expr(program.Operation_COS, ref("ang"))),
			assign("sa", expr(program.Operation_SIN, ref("ang"))),
		},
		vec3Op(
			expr(program.Operation_ADD,
				expr(program.Operation_MUL, acc(ref("v"), 1), ref("ca")),
				expr(program.Operation_MUL, acc(ref("v"), 2), ref("sa"))),
			expr(program.Operation_ADD,
				expr(program.Operation_MUL, expr(program.Operation_NEG, acc(ref("v"), 1)), ref("sa")),
				expr(program.Operation_MUL, acc(ref("v"), 2), ref("ca"))),
			acc(ref("v"), 3),
		))
}

func TestConstUserFunction(t *testing.T) {
	fnTypes := []*program.FnType{rotateXYDecl()}

	got, err := Const(fc("rotateXY", vec(1.5, 0, 0), val(0)), nil, fnTypes)
	if err != nil || !approxEqual(got, []float64{1.5, 0, 0}) {
		t.Fatalf("rotateXY(v, 0): got %v, err %v", got, err)
	}
	got, err = Const(fc("rotateXY", vec(1.5, 0, 0), expr(program.Operation_DIV, expr(program.Operation_MUL, expr(program.Operation_ATAN, val(1)), val(4)), val(2))), nil, fnTypes)
	if err != nil || !approxEqual(got, []float64{0, -1.5, 0}) {
		t.Fatalf("rotateXY(v, pi/2): got %v, err %v", got, err)
	}
}

func TestConstMapScalar(t *testing.T) {
	fnTypes := []*program.FnType{
		fnTypeDecl("doubleVal", []string{"value"}, nil,
			expr(program.Operation_MUL, ref("value"), val(2))),
		fnTypeDecl("scaleByIndex", []string{"value", "index"}, nil,
			expr(program.Operation_MUL, ref("value"), expr(program.Operation_ADD, ref("index"), val(1)))),
	}
	assignments := []*program.Assignment{
		assign("doubled", expr(program.Operation_MAP, ref("doubleVal"), vec(1, 2, 3))),
		assign("scaled", expr(program.Operation_MAP, ref("scaleByIndex"), vec(1, 2, 3))),
	}

	got, err := Const(ref("doubled"), assignments, fnTypes)
	if err != nil || !approxEqual(got, []float64{2, 4, 6}) {
		t.Fatalf("map scalar: got %v, err %v", got, err)
	}

	// A two-parameter function receives (value, index).
	got, err = Const(ref("scaled"), assignments, fnTypes)
	if err != nil || !approxEqual(got, []float64{1, 4, 9}) {
		t.Fatalf("map (value, index): got %v, err %v", got, err)
	}

	// Indexing a scalar map result picks the element.
	got, err = Const(expr(program.Operation_INDEX, ref("doubled"), val(1)), assignments, fnTypes)
	if err != nil || !approxEqual(got, []float64{4}) {
		t.Fatalf("map scalar index: got %v, err %v", got, err)
	}
}

func TestConstMapIndexVector(t *testing.T) {
	pi := math.Pi
	fnTypes := []*program.FnType{
		rotateXYDecl(),
		fnTypeDecl("placeAt", []string{"ang"}, nil,
			fc("rotateXY", expr(program.Operation_MUL, ref("dirX"), ref("shift")), ref("ang"))),
	}
	assignments := []*program.Assignment{
		assign("dirX", vec(1, 0, 0)),
		assign("shift", val(4)),
		assign("angles", vec(0, pi/2, pi)),
		assign("radials", expr(program.Operation_MAP, ref("placeAt"), ref("angles"))),
	}

	cases := []struct {
		index int
		want  []float64
	}{
		{0, []float64{4, 0, 0}},
		{1, []float64{0, -4, 0}},
		{2, []float64{-4, 0, 0}},
	}
	for _, tc := range cases {
		got, err := Const(expr(program.Operation_INDEX, ref("radials"), val(float64(tc.index))), assignments, fnTypes)
		if err != nil {
			t.Fatalf("radials[%d]: err %v", tc.index, err)
		}
		if !approxEqual(got, tc.want) {
			t.Fatalf("radials[%d]: got %v, want %v", tc.index, got, tc.want)
		}
	}

	// The flat value of a vector-producing map folds to its elements'
	// values in order (shaped values made this supported).
	got, err := Const(ref("radials"), assignments, fnTypes)
	if err != nil {
		t.Fatalf("vector-producing map flat value: err %v", err)
	}
	if !approxEqual(got, []float64{4, 0, 0, 0, -4, 0, -4, 0, 0}) {
		t.Fatalf("vector-producing map flat value: got %v", got)
	}
}

func multiAssign(names []string, body *program.Operand) *program.Assignment {
	return program.NewAssignment(names, body, nil, nil)
}

func multiFnTypeDecl(name string, params []string, locals []*program.Assignment, rets ...*program.Operand) *program.FnType {
	fnParams := make([]*program.Parameter, len(params))
	for i, p := range params {
		fnParams[i] = &program.Parameter{Name: p}
	}
	return &program.FnType{
		Name: name,
		Groups: []*program.FnGroup{
			{
				ParamCount: int32(len(params)),
				Functions: []*program.Function{
					{
						Parameters:  fnParams,
						Assignments: locals,
						Returns:     rets,
					},
				},
			},
		},
	}
}

func TestConstMultiReturnDestructure(t *testing.T) {
	fnTypes := []*program.FnType{
		multiFnTypeDecl("build", nil,
			[]*program.Assignment{
				assign("sdf", expr(program.Operation_ADD, ref("p"), val(1))),
				assign("trans", vec(60, 0, -40)),
			},
			ref("sdf"), ref("trans")),
	}
	assignments := []*program.Assignment{
		multiAssign([]string{"shape", "trans"}, fc("build")),
	}

	// The constant return resolves even though its sibling return is not const.
	got, err := Const(ref("trans"), assignments, fnTypes)
	if err != nil || !approxEqual(got, []float64{60, 0, -40}) {
		t.Fatalf("destructured const return: got %v, err %v", got, err)
	}

	// Using it in an expression folds too.
	got, err = Const(expr(program.Operation_INDEX, ref("trans"), val(0)), assignments, fnTypes)
	if err != nil || !approxEqual(got, []float64{60}) {
		t.Fatalf("destructured const return index: got %v, err %v", got, err)
	}

	// The non-constant sibling still fails to fold.
	if _, err := Const(ref("shape"), assignments, fnTypes); err == nil {
		t.Fatalf("expected error folding a non-constant destructured return")
	}
}

func TestConstMultiReturnDestructureImported(t *testing.T) {
	shapes := pkg(nil, nil, []*program.FnType{
		multiFnTypeDecl("branch", []string{"scale"},
			[]*program.Assignment{
				assign("field", expr(program.Operation_MUL, ref("p"), ref("scale"))),
				assign("dims", expr(program.Operation_MUL, vec(10, 20, 30), ref("scale"))),
			},
			ref("field"), ref("dims")),
	})
	getImport := mockImports(map[string]*program.Package{"shapes": shapes}, nil)

	module := pkg([]*program.Import{dotImport("shapes")}, []*program.Assignment{
		multiAssign([]string{"mainPart", "dims"}, fc("branch", val(2))),
	}, nil)

	got, err := ConstPkg(ref("dims"), module, getImport)
	if err != nil || !approxEqual(got, []float64{20, 40, 60}) {
		t.Fatalf("imported destructured const return: got %v, err %v", got, err)
	}

	// Indexing the imported constant return folds too.
	got, err = ConstPkg(expr(program.Operation_INDEX, ref("dims"), val(1)), module, getImport)
	if err != nil || !approxEqual(got, []float64{40}) {
		t.Fatalf("imported destructured const return index: got %v, err %v", got, err)
	}

	// The p-dependent sibling still cannot fold.
	if _, err := ConstPkg(ref("mainPart"), module, getImport); err == nil {
		t.Fatalf("expected error folding a p-dependent destructured return")
	}
}

func TestConstIndexVector(t *testing.T) {
	got, err := Const(expr(program.Operation_INDEX, vec(7, 8, 9), val(2)), nil, nil)
	if err != nil || !approxEqual(got, []float64{9}) {
		t.Fatalf("index literal vector: got %v, err %v", got, err)
	}

	assignments := []*program.Assignment{assign("v", vec(7, 8, 9))}
	got, err = Const(expr(program.Operation_INDEX, ref("v"), val(0)), assignments, nil)
	if err != nil || !approxEqual(got, []float64{7}) {
		t.Fatalf("index referenced vector: got %v, err %v", got, err)
	}

	if _, err := Const(expr(program.Operation_INDEX, vec(7, 8, 9), val(3)), nil, nil); err == nil {
		t.Fatalf("expected out-of-range error")
	}
}

func TestConstImportedFunction(t *testing.T) {
	omega := pkg(nil, []*program.Assignment{assign("dirX", vec(1, 0, 0))}, []*program.FnType{rotateXYDecl()})
	getImport := mockImports(map[string]*program.Package{"omega": omega}, nil)

	module := pkg([]*program.Import{dotImport("omega")},
		[]*program.Assignment{
			assign("shift", val(4)),
			assign("angles", vec(0, math.Pi/2)),
			assign("radials", expr(program.Operation_MAP, ref("placeAt"), ref("angles"))),
		},
		[]*program.FnType{
			fnTypeDecl("placeAt", []string{"ang"}, nil,
				fc("rotateXY", expr(program.Operation_MUL, ref("dirX"), ref("shift")), ref("ang"))),
		},
	)

	got, err := ConstPkg(expr(program.Operation_INDEX, ref("radials"), val(1)), module, getImport)
	if err != nil {
		t.Fatalf("radials[1]: err %v", err)
	}
	if !approxEqual(got, []float64{0, -4, 0}) {
		t.Fatalf("radials[1]: got %v, want [0 -4 0]", got)
	}
}

func TestConstUserFunctionErrors(t *testing.T) {
	fnTypes := []*program.FnType{
		fnTypeDecl("recur", []string{"x"}, nil, fc("recur", ref("x"))),
		fnTypeDecl("usesP", []string{"x"}, nil, expr(program.Operation_ADD, ref("p"), ref("x"))),
	}

	if _, err := Const(fc("recur", val(1)), nil, fnTypes); err == nil {
		t.Fatalf("expected recursion error")
	}
	if _, err := Const(fc("usesP", val(1)), nil, fnTypes); err == nil {
		t.Fatalf("expected error for p reference inside function")
	}
	if _, err := Const(fc("rotateXY", val(1)), nil, nil); err == nil {
		t.Fatalf("expected unknown-function error")
	}
}

func TestConstMatrixMultiply(t *testing.T) {
	identity := vec(
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	)
	matrixA := vec(
		1, 2, 3, 4,
		5, 6, 7, 8,
		9, 10, 11, 12,
		13, 14, 15, 16,
	)

	op := &program.Operand{T: &program.Operand_FnCall{
		FnCall: &program.FnCall{
			Reference: &program.Reference{Prefix: "matrix", Name: "multiply"},
			Arguments: []*program.Operand{identity, matrixA},
		},
	}}

	pkg := &program.Package{}
	got, err := ConstPkg(op, pkg, nil)
	if err != nil {
		t.Fatalf("matrix.multiply failed: %v", err)
	}
	want := []float64{
		1, 2, 3, 4,
		5, 6, 7, 8,
		9, 10, 11, 12,
		13, 14, 15, 16,
	}
	if !approxEqual(got, want) {
		t.Fatalf("matrix.multiply got %v, want %v", got, want)
	}

	// 2x2 case
	got2, err := ConstPkg(&program.Operand{T: &program.Operand_FnCall{
		FnCall: &program.FnCall{
			Reference: &program.Reference{Prefix: "matrix", Name: "multiply"},
			Arguments: []*program.Operand{vec(1, 0, 0, 1), vec(2, 3, 4, 5)},
		},
	}}, pkg, nil)
	if err != nil || !approxEqual(got2, []float64{2, 3, 4, 5}) {
		t.Fatalf("2x2 matrix.multiply: got %v, err %v", got2, err)
	}

	// 3x3 case
	got3, err := ConstPkg(&program.Operand{T: &program.Operand_FnCall{
		FnCall: &program.FnCall{
			Reference: &program.Reference{Prefix: "matrix", Name: "multiply"},
			Arguments: []*program.Operand{vec(1, 0, 0, 0, 1, 0, 0, 0, 1), vec(2, 3, 4, 5, 6, 7, 8, 9, 10)},
		},
	}}, pkg, nil)
	if err != nil || !approxEqual(got3, []float64{2, 3, 4, 5, 6, 7, 8, 9, 10}) {
		t.Fatalf("3x3 matrix.multiply: got %v, err %v", got3, err)
	}

	opErr := &program.Operand{T: &program.Operand_FnCall{
		FnCall: &program.FnCall{
			Reference: &program.Reference{Prefix: "matrix", Name: "multiply"},
			Arguments: []*program.Operand{vec(1, 2), vec(3, 4)},
		},
	}}
	if _, err := ConstPkg(opErr, pkg, nil); err == nil {
		t.Fatalf("expected error for non-square/invalid matrix dimensions")
	}
}

func TestConst_UserFunctionShadowing(t *testing.T) {
	userFnTypes := []*program.FnType{
		fnTypeDecl("multiply", []string{"a", "b"}, nil, expr(program.Operation_ADD, expr(program.Operation_MUL, ref("a"), ref("b")), val(100.0))),
		fnTypeDecl("addScalar", []string{"v", "s"}, nil, expr(program.Operation_ADD, expr(program.Operation_ADD, ref("v"), ref("s")), val(100.0))),
		fnTypeDecl("addVector", []string{"a", "b"}, nil, expr(program.Operation_ADD, expr(program.Operation_ADD, ref("a"), ref("b")), val(100.0))),
		fnTypeDecl("subtractVector", []string{"a", "b"}, nil, expr(program.Operation_ADD, expr(program.Operation_SUB, ref("a"), ref("b")), val(100.0))),
		fnTypeDecl("multiplyScalar", []string{"v", "s"}, nil, expr(program.Operation_ADD, expr(program.Operation_MUL, ref("v"), ref("s")), val(100.0))),
	}

	t.Run("unqualified_calls_execute_user_function", func(t *testing.T) {
		got, err := Const(fc("multiply", val(2), val(3)), nil, userFnTypes)
		if err != nil {
			t.Fatalf("multiply user function failed: %v", err)
		}
		if !approxEqual(got, []float64{106.0}) {
			t.Errorf("expected user function result [106], got %v", got)
		}

		got, err = Const(fc("addScalar", vec(1, 2), val(5)), nil, userFnTypes)
		if err != nil {
			t.Fatalf("addScalar user function failed: %v", err)
		}
		if !approxEqual(got, []float64{106.0, 107.0}) {
			t.Errorf("expected user function result [106, 107], got %v", got)
		}

		got, err = Const(fc("addVector", vec(1, 2), vec(3, 4)), nil, userFnTypes)
		if err != nil {
			t.Fatalf("addVector user function failed: %v", err)
		}
		if !approxEqual(got, []float64{104.0, 106.0}) {
			t.Errorf("expected user function result [104, 106], got %v", got)
		}

		got, err = Const(fc("subtractVector", vec(10, 20), vec(3, 4)), nil, userFnTypes)
		if err != nil {
			t.Fatalf("subtractVector user function failed: %v", err)
		}
		if !approxEqual(got, []float64{107.0, 116.0}) {
			t.Errorf("expected user function result [107, 116], got %v", got)
		}

		got, err = Const(fc("multiplyScalar", vec(2, 3), val(4)), nil, userFnTypes)
		if err != nil {
			t.Fatalf("multiplyScalar user function failed: %v", err)
		}
		if !approxEqual(got, []float64{108.0, 112.0}) {
			t.Errorf("expected user function result [108, 112], got %v", got)
		}
	})

	t.Run("qualified_calls_execute_standard_library_fast_paths", func(t *testing.T) {
		got, err := Const(qfc("vector", "addScalar", vec(1, 2), val(5)), nil, userFnTypes)
		if err != nil {
			t.Fatalf("vector.addScalar fast path failed: %v", err)
		}
		if !approxEqual(got, []float64{6.0, 7.0}) {
			t.Errorf("expected fast-path result [6, 7], got %v", got)
		}

		got, err = Const(qfc("matrix", "multiply", vec(1, 0, 0, 1), vec(2, 3, 4, 5)), nil, userFnTypes)
		if err != nil {
			t.Fatalf("matrix.multiply fast path failed: %v", err)
		}
		if !approxEqual(got, []float64{2.0, 3.0, 4.0, 5.0}) {
			t.Errorf("expected fast-path result [2, 3, 4, 5], got %v", got)
		}
	})
}
