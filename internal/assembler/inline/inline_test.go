package inline_test

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/assembler/source"
	"gradientcontrol.com/omega/core/internal/assembler/inline"
	"gradientcontrol.com/omega/core/internal/closure"
	"gradientcontrol.com/omega/core/internal/collect"
	"gradientcontrol.com/omega/core/internal/eval"
	"gradientcontrol.com/omega/core/internal/link"
	"gradientcontrol.com/omega/core/internal/lower"
	internalsource "gradientcontrol.com/omega/core/internal/source"
	"gradientcontrol.com/omega/core/internal/types"
	parser "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

func runPipeline(src string) (*program.Program, error) {
	return runPipelineWithManager(src, &internalsource.MockManager{})
}

func runPipelineWithManager(src string, srcMgr source.Manager) (*program.Program, error) {
	prog, err := collect.Program([]string{src}, srcMgr, parser.ParsePackage)
	if err != nil {
		return nil, err
	}
	if err := link.References(prog, srcMgr, parser.ParsePackage); err != nil {
		return nil, err
	}
	if err := closure.Transform(prog); err != nil {
		return nil, err
	}
	if err := link.References(prog, srcMgr, parser.ParsePackage); err != nil {
		return nil, err
	}
	if err := types.Analyze(prog); err != nil {
		return nil, err
	}
	if err := lower.Transform(prog); err != nil {
		return nil, err
	}
	if err := link.References(prog, srcMgr, parser.ParsePackage); err != nil {
		return nil, err
	}
	if err := types.Analyze(prog); err != nil {
		return nil, err
	}
	if _, err := inline.Transform(prog); err != nil {
		return nil, err
	}
	if err := link.References(prog, srcMgr, parser.ParsePackage); err != nil {
		return nil, err
	}
	if err := types.Analyze(prog); err != nil {
		return nil, err
	}
	return prog, nil
}

func TestInline_SingleReturn(t *testing.T) {
	src := `
addTwo(a, b) {
	return a + b;
}
return addTwo(3, 4);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ret := prog.Primary.Files[0].Returns[0]
	// Ret should no longer be a FnCall, but an ADD expression (3 + 4)
	exprOp, ok := ret.T.(*program.Operand_Expression)
	if !ok {
		t.Fatalf("expected inlined Expression operand, got %T", ret.T)
	}
	if exprOp.Expression.Operation != program.Operation_ADD {
		t.Errorf("expected Operation_ADD, got %v", exprOp.Expression.Operation)
	}
	if len(exprOp.Expression.Operands) != 2 {
		t.Fatalf("expected 2 operands, got %d", len(exprOp.Expression.Operands))
	}
	v1 := exprOp.Expression.Operands[0].GetValue()
	v2 := exprOp.Expression.Operands[1].GetValue()
	if v1 != 3 || v2 != 4 {
		t.Errorf("expected operands 3 and 4, got %v and %v", v1, v2)
	}
}

func TestInline_WithLocalAssignments(t *testing.T) {
	src := `
calc(x) {
	y = x * 2;
	return y + 1;
}
return calc(5);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ret := prog.Primary.Files[0].Returns[0]
	exprOp, ok := ret.T.(*program.Operand_Expression)
	if !ok {
		t.Fatalf("expected inlined Expression operand, got %T", ret.T)
	}
	if exprOp.Expression.Operation != program.Operation_ADD {
		t.Errorf("expected Operation_ADD, got %v", exprOp.Expression.Operation)
	}
}

func TestInline_NestedCalls(t *testing.T) {
	src := `
doubleVal(x) {
	return x * 2;
}
quadruple(x) {
	return doubleVal(doubleVal(x));
}
return quadruple(3);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ret := prog.Primary.Files[0].Returns[0]
	// Should be fully inlined (no remaining FnCalls to quadruple or doubleVal)
	if _, ok := ret.T.(*program.Operand_FnCall); ok {
		t.Errorf("expected quadruple call to be fully inlined, got FnCall")
	}
}

func TestInline_MultiReturn(t *testing.T) {
	src := `
pointNormal() {
	return [1, 2, 3], [0, 1, 0];
}
pt, n = pointNormal();
return pt + n;
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Multi-return inlining replaces assignment or resolves variables
	file := prog.Primary.Files[0]
	if len(file.Assignments) > 0 {
		for _, assign := range file.Assignments {
			if fc, ok := assign.Expression.T.(*program.Operand_FnCall); ok {
				if fc.FnCall.Reference.Name == "pointNormal" {
					t.Errorf("expected pointNormal to be inlined into separate assignments or expressions")
				}
			}
		}
	}
}

func TestInline_RecursionError(t *testing.T) {
	src := `
rec(n) {
	return rec(n - 1);
}
return rec(5);
`
	_, err := runPipeline(src)
	if err == nil {
		t.Fatalf("expected recursion error, got nil")
	}
	if !strings.Contains(err.Error(), "recurs") {
		t.Errorf("expected error mentioning recursion, got: %v", err)
	}
}

func TestInline_CrossPackage(t *testing.T) {
	srcMgr := internalsource.NewLocalManager("../../../stdlib")
	src := `
import font "text";
return font.fontSpacing();
`
	prog, err := runPipelineWithManager(src, srcMgr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ret := prog.Primary.Files[0].Returns[0]
	// fontSpacing() returns a constant 0.3333, which inlines into an Operand_Value (or expression)
	if _, ok := ret.T.(*program.Operand_FnCall); ok {
		t.Errorf("expected font.fontSpacing() to be inlined, but found Operand_FnCall: %s", ret.PrettyPrint(0))
	}
}

func TestInline_CrossPackageQualification(t *testing.T) {
	srcMgr := internalsource.NewLocalManager("../../../stdlib")
	src := `
import font "text";
return font.left("A", [0, 0], 1, 0.04, 0.3);
`
	prog, err := runPipelineWithManager(src, srcMgr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ret := prog.Primary.Files[0].Returns[0]
	// font.left should be inlined at the top-level return
	if fc, ok := ret.T.(*program.Operand_FnCall); ok && fc.FnCall.Reference.Name == "left" {
		t.Errorf("expected font.left to be inlined, but top return is still font.left: %s", ret.PrettyPrint(0))
	}
}

func TestInline_NestedAnonFunctions(t *testing.T) {
	src := `
outer(k) {
	helper(x) {
		return x + k;
	}
	return helper(10);
}
return outer(5);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ret := prog.Primary.Files[0].Returns[0]
	if fc, ok := ret.T.(*program.Operand_FnCall); ok && fc.FnCall.Reference.Name == "outer" {
		t.Errorf("expected outer to be inlined, got FnCall to outer")
	}
}

func TestInline_CrossPackageMultiReturn(t *testing.T) {
	srcMgr := &internalsource.MockManager{
		PkgSources: map[string][]string{
			"geom/helpers": {
				`
splitShape(size) {
	s1 = size * 2;
	s2 = size * 3;
	return s1, s2;
}
`,
			},
		},
	}
	src := `
import helpers "geom/helpers";
val1, val2 = helpers.splitShape(10);
return val1 + val2;
`
	prog, err := runPipelineWithManager(src, srcMgr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	file := prog.Primary.Files[0]
	for _, assign := range file.Assignments {
		if fc, ok := assign.Expression.T.(*program.Operand_FnCall); ok {
			if fc.FnCall.Reference.Name == "splitShape" {
				t.Errorf("expected helpers.splitShape to be inlined into individual variable assignments, got FnCall")
			}
		}
	}
}

// TestInline_HoistedLambdaNeverCapturesACallerLocal: a library closure passed
// through a library higher-order function is baked with its capture as a
// LOCAL — `major = 2` — and its body calls a function whose table fold
// captures its argument. Constant folding knows that local's value, so the
// argument reads as constant; but inlining the call would hoist the fold's
// lambda to file level, outside the function that holds `major`, and write
// `major` into it. The call must stay a call, and the program must link. It
// is drill.tap's closure consulting catalog.pitch.
func TestInline_HoistedLambdaNeverCapturesACallerLocal(t *testing.T) {
	tab := `
table = [[1.0, 10.0], [2.0, 20.0]];
onRow(k, key) {
  return select(abs(k - key) < 0.0001, 1.0, 0.0);
}
lookup(n) {
  k = n;
  return reduce((acc, v) { return acc + v; },
                map((r) { return r[1] * onRow(k, r[0]); }, table));
}
`
	lib := `
import tab "geom/tab";
make(major) {
  cut(z) {
    return tab.lookup(major) + z;
  }
  return cut;
}
apply(f, z) {
  return f(z);
}
`
	src := `
import lib "geom/lib";
return lib.apply(lib.make(2.0), p.z);
`
	mgr := &internalsource.MockManager{PkgSources: map[string][]string{"geom/tab": {tab}, "geom/lib": {lib}}}
	prog, err := runPipelineWithManager(src, mgr)
	if err != nil {
		t.Fatal(err)
	}
	// The local's value is known only once constant folding has run, so the
	// hoist happens on a LATER round of the assembler's optimize loop, which
	// these rounds reproduce.
	for round := 0; round < 3; round++ {
		if _, err := eval.Fold(prog, nil); err != nil {
			t.Fatal(err)
		}
		if _, err := inline.Transform(prog); err != nil {
			t.Fatal(err)
		}
		if err := link.References(prog, mgr, parser.ParsePackage); err != nil {
			t.Fatalf("round %d: a closure whose body folds a table through a captured parameter must link: %v", round, err)
		}
		if err := types.Analyze(prog); err != nil {
			t.Fatal(err)
		}
		if err := lower.Transform(prog); err != nil {
			t.Fatal(err)
		}
		if err := link.References(prog, mgr, parser.ParsePackage); err != nil {
			t.Fatal(err)
		}
		if err := types.Analyze(prog); err != nil {
			t.Fatal(err)
		}
	}
	for _, dep := range prog.Dependencies {
		for _, file := range dep.Files {
			for _, ft := range file.FnTypes {
				if !strings.HasPrefix(ft.Name, program.InlinedFnPrefix) {
					continue
				}
				for _, grp := range ft.Groups {
					for _, fn := range grp.Functions {
						if strings.Contains(fn.PrettyPrint(0), "major") {
							t.Errorf("%s hoisted %s, which names the caller's local major:\n%s",
								dep.Path, ft.Name, fn.PrettyPrint(0))
						}
					}
				}
			}
		}
	}
}

func TestInline_PreservesHardwareAcceleratedPackages(t *testing.T) {
	src := `
import "matrix";
import "vector";
m = [[1.0, 0.0], [0.0, 1.0]];
v = [1.0, 2.0];
resM = matrix.scale(2.0, m);
resV = vector.addScalar(v, 3.0);
return resM;
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	file := prog.Primary.Files[0]
	var foundMatrixScale, foundVectorAddScalar bool
	for _, assign := range file.Assignments {
		if fc, ok := assign.Expression.T.(*program.Operand_FnCall); ok && fc.FnCall.Reference != nil {
			if fc.FnCall.Reference.Prefix == "matrix" && fc.FnCall.Reference.Name == "scale" {
				foundMatrixScale = true
			}
			if fc.FnCall.Reference.Prefix == "vector" && fc.FnCall.Reference.Name == "addScalar" {
				foundVectorAddScalar = true
			}
		}
	}
	if !foundVectorAddScalar {
		t.Errorf("expected vector.addScalar to remain un-inlined FnCall")
	}
	if !foundMatrixScale {
		t.Errorf("expected matrix.scale to remain un-inlined FnCall")
	}
}

func TestInline_FunctionCallingReduceWithPNotMisclassifiedAsConst(t *testing.T) {
	src := `
mix(a, b, t) {
  return a + (b - a) * t;
}

waveSum(freq) {
  return reduce(0.0, (acc, i) {
    return acc + cos(p.x * freq + i);
  }, [0, 1, 2, 3, 4, 5, 6, 7]);
}

computeWall(freq) {
  w = waveSum(freq);
  return mix([1.0, 0.0, 0.0], [0.0, 1.0, 0.0], w);
}

res = computeWall(2.0);
return res;
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify that mix was NOT inlined with constant arguments into computeWall / res
	file := prog.Primary.Files[0]
	foundMixCall := false
	for _, assign := range file.Assignments {
		t.Logf("assign: %s = %s", assign.Variables[0].Name, assign.Expression.PrettyPrint(0))
		if assign.Expression != nil {
			if fc := assign.Expression.GetFnCall(); fc != nil && fc.Reference != nil && fc.Reference.Name == "mix" {
				foundMixCall = true
			}
		}
	}
	for _, ft := range file.FnTypes {
		for _, grp := range ft.Groups {
			for _, fn := range grp.Functions {
				t.Logf("fn %s: %s", ft.Name, fn.PrettyPrint(0))
				for _, assign := range fn.Assignments {
					if assign.Expression != nil {
						if fc := assign.Expression.GetFnCall(); fc != nil && fc.Reference != nil && fc.Reference.Name == "mix" {
							foundMixCall = true
						}
					}
				}
				for _, ret := range fn.Returns {
					if fc := ret.GetFnCall(); fc != nil && fc.Reference != nil && fc.Reference.Name == "mix" {
						foundMixCall = true
					}
				}
			}
		}
	}
	if !foundMixCall {
		t.Errorf("expected call to 'mix' to NOT be inlined because argument depends on p via reduce")
	}
}

func TestInline_MixOklabNotConstInlinedWhenArgumentCallsReduceWithP(t *testing.T) {
	src := `
mixOklab(c1, c2, t) {
  return c1 * (1.0 - t) + c2 * t;
}

fade(x, a) {
  return x * a;
}

latticeNoise(scale) {
  return reduce([0.0, 0.0], (acc, o) {
    return [acc.x + cos(p.x * scale + o.x), acc.y + sin(p.y * scale + o.y)];
  }, [[1.0, 0.0], [0.0, 1.0]]);
}

edgesNoise(size) {
  lat = latticeNoise(size);
  return lat.y - lat.x;
}

cellSkin(c1, c2, size) {
  edges = edgesNoise(size);
  wall = fade(edges, 0.2);
  return mixOklab(c1, c2, wall);
}

res = cellSkin([0.0, 1.0, 1.0], [1.0, 1.0, 1.0], 5.0);
return res;
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify that mixOklab was NOT inlined with constant arguments
	file := prog.Primary.Files[0]
	foundMixOklabCall := false
	for _, assign := range file.Assignments {
		if assign.Expression != nil {
			if fc := assign.Expression.GetFnCall(); fc != nil && fc.Reference != nil && fc.Reference.Name == "mixOklab" {
				foundMixOklabCall = true
			}
		}
	}
	for _, ft := range file.FnTypes {
		for _, grp := range ft.Groups {
			for _, fn := range grp.Functions {
				for _, assign := range fn.Assignments {
					if assign.Expression != nil {
						if fc := assign.Expression.GetFnCall(); fc != nil && fc.Reference != nil && fc.Reference.Name == "mixOklab" {
							foundMixOklabCall = true
						}
					}
				}
				for _, ret := range fn.Returns {
					if fc := ret.GetFnCall(); fc != nil && fc.Reference != nil && fc.Reference.Name == "mixOklab" {
						foundMixOklabCall = true
					}
				}
			}
		}
	}
	if !foundMixOklabCall {
		t.Errorf("expected call to 'mixOklab' to NOT be inlined because argument depends on p via reduce")
	}
}
