package property

import (
	"math"
	"testing"

	"gradientcontrol.com/omega/core/program"
)

// TestCompanionBoundaryMap exercises a function-internal derivative:
//
//	boundaryMap(v) { return v * v.dp; }
//	return boundaryMap(length(p));
//
// boundaryMap is tainted (its body uses .dp), so the call is replaced by a
// companion boundaryMap__d(v, v__dp) that resolves v.dp against the passed-in
// gradient. With v = length(p), v.dp = normalize(p) = p/|p|, so the result is
// |p| * p/|p| = p.
func TestCompanionBoundaryMap(t *testing.T) {
	boundaryMap := &program.Function{
		Parameters: []*program.Parameter{{Name: "v"}},
		Returns: []*program.Operand{
			nary(program.Operation_MUL, ref("v"),
				nary(program.Operation_DERIV, ref("v"), ref(program.PVarName))),
		},
	}
	pkg := &program.Package{
		Files: []*program.PackageFile{
			{
				FnTypes: []*program.FnType{
					{
						Name: "boundaryMap",
						Groups: []*program.FnGroup{
							{
								ParamCount: 1,
								Functions:  []*program.Function{boundaryMap},
							},
						},
					},
				},
				Returns: []*program.Operand{call("boundaryMap", call("length", ref(program.PVarName)))},
			},
		},
	}

	prog, err := Expand(&program.Program{Primary: pkg})
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	out := prog.Primary

	// The companion function must have been generated with a gradient parameter.
	comp := findFunction(out, "boundaryMap__d")
	if comp == nil {
		t.Fatal("companion boundaryMap__d was not generated")
	}
	wantParams := []string{"v", "v__dp0", "v__dp1", "v__dp2"}
	if len(comp.Parameters) != len(wantParams) {
		t.Fatalf("companion params = %v, want %v", comp.Parameters, wantParams)
	}
	for i, want := range wantParams {
		if comp.Parameters[i].Name != want {
			t.Fatalf("companion params = %v, want %v", comp.Parameters, wantParams)
		}
	}

	// The call site must invoke the companion with the original arg plus its
	// derivative along each axis.
	retCall, ok := out.Files[0].Returns[0].T.(*program.Operand_FnCall)
	if !ok || retCall.FnCall.Reference.Name != "boundaryMap__d" {
		t.Fatalf("return is not a companion call: %s", out.Files[0].Returns[0].PrettyPrint(0))
	}
	if len(retCall.FnCall.Arguments) != 4 {
		t.Fatalf("companion call has %d args, want 4", len(retCall.FnCall.Arguments))
	}
	arg0 := retCall.FnCall.Arguments[0] // length(p)
	// The three axis derivatives of length(p), i.e. the components of
	// normalize(p).
	dp := nary(program.Operation_COLLECT,
		retCall.FnCall.Arguments[1], retCall.FnCall.Arguments[2], retCall.FnCall.Arguments[3])

	// The companion body is v * [v__dp0, v__dp1, v__dp2]; substituting the call
	// arguments gives arg0 * dp, which must equal p at every sample point.
	body := nary(program.Operation_MUL, arg0, dp)
	for _, pt := range []point{{0.3, -1.2, 0.7, 0.4}, {-0.6, 0.9, -0.2, 1.1}, {1.3, 0.2, 0.5, 0.0}} {
		got := evalVec(t, body, pt)
		want := []float64{pt.x, pt.y, pt.z}
		if len(got) != 3 {
			t.Fatalf("expected vec3, got len %d", len(got))
		}
		for k := 0; k < 3; k++ {
			if !closeEnough(got[k], want[k]) {
				t.Errorf("at %+v component %d: got %v want %v", pt, k, got[k], want[k])
			}
		}
	}
}

func findFunction(pkg *program.Package, name string) *program.Function {
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, ft := range f.FnTypes {
			if ft != nil && ft.Name == name {
				for _, group := range ft.Groups {
					if group != nil {
						for _, fn := range group.Functions {
							if fn != nil {
								return fn
							}
						}
					}
				}
			}
		}
	}
	return nil
}

// TestCompanionTransitiveNeeds: a function with no derivative of its own that
// forwards locals of its parameters into a differentiating callee still needs
// its parameters' tangents — the callee's needs propagate through the call.
//
//	normalizeField(f)      { return f / length(f.dp); }
//	lengthNormalized(a, b) { s = a + b; return normalizeField(s); }
//
// Without the propagation lengthNormalized's companion took no tangents, so the
// nested normalizeField call saw zero gradients and length(f.dp) folded to 0.
func TestCompanionTransitiveNeeds(t *testing.T) {
	normalizeField := &program.Function{
		Parameters: []*program.Parameter{{Name: "f"}},
		Returns: []*program.Operand{
			div(ref("f"), call("length",
				nary(program.Operation_DERIV, ref("f"), ref(program.PVarName)))),
		},
	}
	lengthNormalized := &program.Function{
		Parameters: []*program.Parameter{{Name: "a"}, {Name: "b"}},
		Assignments: []*program.Assignment{
			assignDecl("s", add(ref("a"), ref("b"))),
		},
		Returns: []*program.Operand{call("normalizeField", ref("s"))},
	}
	pkg := &program.Package{Files: []*program.PackageFile{{
		FnTypes: []*program.FnType{
			{Name: "normalizeField", Groups: []*program.FnGroup{{ParamCount: 1, Functions: []*program.Function{normalizeField}}}},
			{Name: "lengthNormalized", Groups: []*program.FnGroup{{ParamCount: 2, Functions: []*program.Function{lengthNormalized}}}},
		},
		Returns: []*program.Operand{call("lengthNormalized",
			call("length", ref(program.PVarName)),
			nary(program.Operation_MUL, ref(program.XVarName), ref(program.XVarName)))},
	}}}

	prog, err := Expand(&program.Program{Primary: pkg})
	if err != nil {
		t.Fatalf("Expand: %v", err)
	}
	out := prog.Primary
	comp := findFunction(out, "lengthNormalized__d")
	if comp == nil {
		t.Fatal("companion lengthNormalized__d was not generated")
	}
	var names []string
	for _, p := range comp.Parameters {
		names = append(names, p.Name)
	}
	want := []string{"a", "b", "a__dp0", "a__dp1", "a__dp2", "b__dp0", "b__dp1", "b__dp2"}
	if len(names) != len(want) {
		t.Fatalf("companion params = %v, want %v", names, want)
	}
	for i := range want {
		if names[i] != want[i] {
			t.Fatalf("companion params = %v, want %v", names, want)
		}
	}

	// Numeric: lengthNormalized(|p|, x*x) = (|p| + x*x) / |grad(|p| + x*x)|.
	retCall, ok := out.Files[0].Returns[0].T.(*program.Operand_FnCall)
	if !ok || retCall.FnCall.Reference.Name != "lengthNormalized__d" {
		t.Fatalf("return is not a companion call: %s", out.Files[0].Returns[0].PrettyPrint(0))
	}
	if len(retCall.FnCall.Arguments) != 8 {
		t.Fatalf("companion call has %d args, want 8", len(retCall.FnCall.Arguments))
	}
	// s = a + b; the nested normalizeField__d receives s's axis derivatives,
	// each a__dpk + b__dpk; result = s / length(grad s). Reconstruct that from
	// the call-site arguments and check numerically.
	a, b := retCall.FnCall.Arguments[0], retCall.FnCall.Arguments[1]
	axes := make([]*program.Operand, 3)
	for k := 0; k < 3; k++ {
		axes[k] = add(retCall.FnCall.Arguments[2+k], retCall.FnCall.Arguments[5+k])
	}
	body := div(add(a, b), call("length", nary(program.Operation_COLLECT, axes...)))
	for _, pt := range []point{{0.4, -0.7, 0.5, 0}, {1.2, 0.3, -0.8, 0}} {
		got := evalScalar(t, body, pt)
		// grad(|p|) = p/|p|; grad(x*x) = (2x, 0, 0).
		r := math.Sqrt(pt.x*pt.x + pt.y*pt.y + pt.z*pt.z)
		gx, gy, gz := pt.x/r+2*pt.x, pt.y/r, pt.z/r
		want := (r + pt.x*pt.x) / math.Sqrt(gx*gx+gy*gy+gz*gz)
		if !closeEnough(got, want) {
			t.Errorf("at %+v: got %v want %v", pt, got, want)
		}
	}
}
