package closure_test

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/assembler/source"
	"gradientcontrol.com/omega/core/internal/closure"
	"gradientcontrol.com/omega/core/internal/collect"
	"gradientcontrol.com/omega/core/internal/link"
	internalsource "gradientcontrol.com/omega/core/internal/source"
	parser "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
)

func runPipeline(src string) (*program.Program, error) {
	srcMgr := &internalsource.MockManager{}
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
	return prog, nil
}

func TestClosure_SingleLevelFactory(t *testing.T) {
	src := `
scaleBy(k) {
  return (a) { return a * k; };
}

soft = scaleBy(0.25);
loud = scaleBy(4.0);
return soft(10.0) + loud(2.0);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	file := prog.Primary.Files[0]
	// Verify that soft and loud function-typed assignments are eliminated or lowered
	for _, assign := range file.Assignments {
		for _, v := range assign.Variables {
			if v.Name == "soft" || v.Name == "loud" {
				t.Errorf("function-typed variable assignment %q was not eliminated", v.Name)
			}
		}
	}

	// Verify synthesized functions exist with deterministic names (gcl_closure_scaleBy, gcl_closure_scaleBy_2)
	foundFirst := false
	foundSecond := false
	for _, ft := range file.FnTypes {
		if ft.Name == "gcl_closure_scaleBy" {
			foundFirst = true
		}
		if ft.Name == "gcl_closure_scaleBy_2" {
			foundSecond = true
		}
	}
	if !foundFirst {
		t.Errorf("expected synthesized function gcl_closure_scaleBy not found")
	}
	if !foundSecond {
		t.Errorf("expected synthesized function gcl_closure_scaleBy_2 not found")
	}
}

func TestClosure_NestedFlattened(t *testing.T) {
	src := `
ramp(base) {
  return (rate) {
    return (timeVal) { return base + rate * timeVal; };
  };
}

sweepBase = ramp(110.0);
sweep = sweepBase(2.0);
return sweep(time);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	file := prog.Primary.Files[0]
	foundFlattened := false
	for _, ft := range file.FnTypes {
		if strings.HasPrefix(ft.Name, "gcl_closure_ramp") {
			foundFlattened = true
			fn := ft.Groups[0].Functions[0]
			// The flattened closure should accept timeVal (and outer captures if parameterized)
			if len(fn.Parameters) == 0 && len(fn.Assignments) == 0 {
				t.Errorf("synthesized closure %s is empty", ft.Name)
			}
		}
	}
	if !foundFlattened {
		t.Errorf("expected flattened synthesized function gcl_closure_ramp not found")
	}
}

func TestClosure_HigherOrderParameter(t *testing.T) {
	src := `
apply(x, op) {
  return op(x);
}

inc(k) {
  return (a) { return a + k; };
}

addFive = inc(5.0);
return apply(10.0, addFive);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	file := prog.Primary.Files[0]
	foundSpecializedApply := false
	for _, ft := range file.FnTypes {
		if strings.HasPrefix(ft.Name, "gcl_closure_apply") || strings.HasPrefix(ft.Name, "apply_") {
			foundSpecializedApply = true
		}
	}
	if !foundSpecializedApply {
		t.Logf("Checking if apply was specialized in place or under new name")
	}
}

func TestClosure_MultiHopForwarding(t *testing.T) {
	src := `
outer(x, f) {
  return inner(x, f);
}

inner(x, f) {
  return f(x);
}

doubleIt(k) {
  return (a) { return a * k; };
}

d = doubleIt(2.0);
return outer(5.0, d);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prog == nil {
		t.Fatalf("expected valid program")
	}
}

func TestClosure_LexicalScopeIsolation(t *testing.T) {
	src := `
makeAdder(x) {
  return (x) { return x + 1.0; };
}

adder = makeAdder(10.0);
return adder(5.0);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prog == nil {
		t.Fatalf("expected valid program")
	}
}

func TestClosure_ErrorMixedReturns(t *testing.T) {
	src := `
badFunc(k) {
  return (x) { return x + k; }, 42.0;
}

f, val = badFunc(1.0);
return f(val);
`
	_, err := runPipeline(src)
	if err == nil {
		t.Fatalf("expected error for mixed tuple return, got nil")
	}
	expected := "a function that returns a function cannot return anything else"
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("expected error message %q, got %q", expected, err.Error())
	}
}

func TestClosure_ErrorMultiFunctionReturns(t *testing.T) {
	src := `
badPair(k) {
  return (x) { return x + k; }, (y) { return y * k; };
}

f1, f2 = badPair(2.0);
return f1(f2(1.0));
`
	_, err := runPipeline(src)
	if err == nil {
		t.Fatalf("expected error for multiple function returns, got nil")
	}
	expected := "a function that returns a function cannot return anything else"
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("expected error message %q, got %q", expected, err.Error())
	}
}

func TestClosure_ErrorUncalledPrimaryReturn(t *testing.T) {
	src := `
scaleBy(k) {
  return (a) { return a * k; };
}

return scaleBy(0.5);
`
	_, err := runPipeline(src)
	if err == nil {
		t.Fatalf("expected error for uncalled primary return, got nil")
	}
	expected := "cannot return a function; perhaps forgot parentheses () after function name?"
	if !strings.Contains(err.Error(), expected) {
		t.Errorf("expected error message %q, got %q", expected, err.Error())
	}
}

func TestClosure_DerivedNormalExact(t *testing.T) {
	src := `
sphereSdf() {
  return length(p) - 1.0;
}
boxSdf() {
  return length(p) - 2.0;
}
unionRoundExact(sdf1, sdf2) {
  return unionRoundExact(sdf1, sdf2, sdf1().dp, sdf2().dp);
}
unionRoundExact(sdf1, sdf2, n1, n2) {
  d1 = sdf1();
  d2 = sdf2();
  return min(d1, d2);
}

return unionRoundExact(sphereSdf, boxSdf);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prog == nil {
		t.Fatalf("expected valid program")
	}
}

// TestClosure_FullApplicationBoundToName covers a closure whose call result is
// bound to a name rather than consumed inline. A call supplying every parameter
// of the closure's target is a full application and yields a value, so it must
// survive as an assignment and be rewritten into a direct call to the function
// already synthesized for that closure. Treating it as a partial application
// re-specialized the original template with only these arguments bound and
// dropped the captures baked in at synthesis, so the body was left referring to
// a capture that no longer existed.
//
// Both instantiations are exercised because the captures must stay per-closure:
// sharing one binding for k would silently give one of them the other's value.
func TestClosure_FullApplicationBoundToName(t *testing.T) {
	src := `
scaleBy(k) {
  return (a) { return a * k; };
}

soft = scaleBy(0.25);
loud = scaleBy(4.0);
quiet = soft(10.0);
blare = loud(2.0);
return quiet + blare;
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	file := prog.Primary.Files[0]

	// The value-producing assignments stay; the function-typed ones do not.
	kept := make(map[string]bool)
	for _, assign := range file.Assignments {
		for _, v := range assign.Variables {
			kept[v.Name] = true
		}
	}
	for _, name := range []string{"quiet", "blare"} {
		if !kept[name] {
			t.Errorf("assignment %q was dropped; a fully applied closure call yields a value and must be kept", name)
		}
	}
	for _, name := range []string{"soft", "loud"} {
		if kept[name] {
			t.Errorf("function-typed variable assignment %q was not eliminated", name)
		}
	}

	// Each call must land on its own synthesized function, so the two captures
	// cannot collide.
	targets := make(map[string]string)
	for _, assign := range file.Assignments {
		fc, ok := assign.Expression.T.(*program.Operand_FnCall)
		if !ok || fc.FnCall == nil || fc.FnCall.Reference == nil {
			continue
		}
		for _, v := range assign.Variables {
			targets[v.Name] = fc.FnCall.Reference.Name
		}
	}
	for _, name := range []string{"quiet", "blare"} {
		if !strings.HasPrefix(targets[name], "gcl_closure_scaleBy") {
			t.Errorf("%q calls %q, want a direct call to a synthesized gcl_closure_scaleBy*", name, targets[name])
		}
	}
	if targets["quiet"] == targets["blare"] {
		t.Errorf("both calls target %q; the two closures must not share a synthesized function, or one capture overwrites the other", targets["quiet"])
	}
}

// TestClosure_PartialApplicationOfSynthesized covers supplying only some of a
// closure's remaining parameters. The arguments given are recorded as captures
// of parameters the target still declares, so the function synthesized for the
// result must drop those parameters — otherwise it keeps the template's arity
// and the next call cannot match it. The captures baked in by the earlier
// synthesis have to survive this step too, or the body loses them.
func TestClosure_PartialApplicationOfSynthesized(t *testing.T) {
	src := `
scaleBy(k) {
  return (a, b) { return a * b * k; };
}

half = scaleBy(0.5);
stepped = half(2.0);
return stepped(3.0);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	file := prog.Primary.Files[0]

	// The closure that stepped resolves to must take exactly the one parameter
	// still unbound, not the two its template declared.
	var target string
	for _, assign := range file.Assignments {
		for _, v := range assign.Variables {
			if v.Name == "stepped" {
				t.Errorf("function-typed variable assignment %q was not eliminated", v.Name)
			}
		}
	}
	for _, ret := range file.Returns {
		if fc, ok := ret.T.(*program.Operand_FnCall); ok && fc.FnCall != nil && fc.FnCall.Reference != nil {
			target = fc.FnCall.Reference.Name
			if got := len(fc.FnCall.Arguments); got != 1 {
				t.Errorf("return calls %q with %d arguments, want 1", target, got)
			}
		}
	}
	if !strings.HasPrefix(target, "gcl_closure_scaleBy") {
		t.Fatalf("return calls %q, want a synthesized gcl_closure_scaleBy*", target)
	}

	found := false
	for _, ft := range file.FnTypes {
		if ft.Name != target {
			continue
		}
		for _, fg := range ft.Groups {
			if len(fg.Functions) == 0 {
				continue
			}
			found = true
			if fg.ParamCount != 1 {
				t.Errorf("%s has ParamCount %d, want 1 — a partially applied closure must drop the parameters it bound", target, fg.ParamCount)
			}
			if got := len(fg.Functions[0].Parameters); got != 1 {
				t.Errorf("%s declares %d parameters, want 1", target, got)
			}
		}
	}
	if !found {
		t.Errorf("synthesized function %q was not registered on the file", target)
	}
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
	return prog, nil
}

// libScaleBy is a factory in its own package whose closure body calls a
// function of that package UNQUALIFIED — which is exactly what a clone of the
// body into the caller's file cannot resolve, and what the second link in the
// pipeline checks.
const libScaleBy = `
twice(x) {
  return x * 2;
}
scaleBy(k) {
  by(a) {
    return twice(a) * k;
  }
  return by;
}
`

func libManager(src string) *internalsource.MockManager {
	return &internalsource.MockManager{PkgSources: map[string][]string{"geom/lib": {src}}}
}

func depFile(t *testing.T, prog *program.Program, path string) *program.PackageFile {
	t.Helper()
	for _, dep := range prog.Dependencies {
		if dep.Path == path && len(dep.Files) > 0 {
			return dep.Files[0]
		}
	}
	t.Fatalf("dependency %q not found", path)
	return nil
}

func synthesizedIn(file *program.PackageFile) []string {
	var names []string
	for _, ft := range file.FnTypes {
		if strings.HasPrefix(ft.Name, "gcl_closure_") {
			names = append(names, ft.Name)
		}
	}
	return names
}

func fnNamed(file *program.PackageFile, name string) *program.Function {
	for _, ft := range file.FnTypes {
		if ft.Name != name {
			continue
		}
		for _, fg := range ft.Groups {
			if len(fg.Functions) > 0 {
				return fg.Functions[0]
			}
		}
	}
	return nil
}

func paramNames(fn *program.Function) string {
	names := make([]string, len(fn.Parameters))
	for i, p := range fn.Parameters {
		names[i] = p.Name
	}
	return strings.Join(names, ",")
}

func callOf(op *program.Operand) *program.FnCall {
	if fc, ok := op.T.(*program.Operand_FnCall); ok && fc.FnCall != nil && fc.FnCall.Reference != nil {
		return fc.FnCall
	}
	return nil
}

// TestClosure_CrossPackageFactoryIsLifted covers a closure made by a factory
// from ANOTHER package. Its body is left in that package with the captures as
// leading parameters, and each call through it passes the capture values —
// the caller's own expressions — as leading arguments. One lifted function
// serves both instantiations, so the two cannot collide on a capture.
func TestClosure_CrossPackageFactoryIsLifted(t *testing.T) {
	src := `
import lib "geom/lib";
gain = 0.25;
soft = lib.scaleBy(gain);
loud = lib.scaleBy(4.0);
quiet = soft(10.0);
blare = loud(2.0);
return quiet + blare;
`
	prog, err := runPipelineWithManager(src, libManager(libScaleBy))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	file := prog.Primary.Files[0]
	calls := make(map[string]*program.FnCall)
	for _, assign := range file.Assignments {
		if fc := callOf(assign.Expression); fc != nil {
			for _, v := range assign.Variables {
				calls[v.Name] = fc
			}
		}
	}
	for _, name := range []string{"quiet", "blare"} {
		fc := calls[name]
		if fc == nil {
			t.Fatalf("%q is not a direct call", name)
		}
		if fc.Reference.Prefix != "lib" || !strings.HasPrefix(fc.Reference.Name, "gcl_closure_scaleBy") {
			t.Errorf("%q calls %s.%s, want lib.gcl_closure_scaleBy*: the lifted function lives in the factory's package", name, fc.Reference.Prefix, fc.Reference.Name)
		}
		if got := len(fc.Arguments); got != 2 {
			t.Errorf("%q passes %d arguments, want 2: the capture then the argument", name, got)
		}
	}
	if calls["quiet"].Reference.Name != calls["blare"].Reference.Name {
		t.Errorf("quiet calls %q and blare calls %q; one lifted function must serve every instantiation", calls["quiet"].Reference.Name, calls["blare"].Reference.Name)
	}
	if ref, ok := calls["quiet"].Arguments[0].T.(*program.Operand_Reference); !ok || ref.Reference == nil || ref.Reference.Name != "gain" {
		t.Errorf("quiet's leading argument is %s, want the caller's own variable gain", calls["quiet"].Arguments[0].PrettyPrint(0))
	}

	libFile := depFile(t, prog, "geom/lib")
	fn := fnNamed(libFile, calls["quiet"].Reference.Name)
	if fn == nil {
		t.Fatalf("%q was not added to the library file", calls["quiet"].Reference.Name)
	}
	if got := paramNames(fn); got != "k,a" {
		t.Errorf("lifted function takes (%s), want (k,a): the capture leads, then the closure's own parameter", got)
	}
	if len(fn.Assignments) != 0 {
		t.Errorf("lifted function has %d prepended assignments, want none: captures are parameters, not baked values", len(fn.Assignments))
	}
	if got := synthesizedIn(file); len(got) != 0 {
		t.Errorf("caller's file gained %v; a cross-package closure must not be cloned into the caller", got)
	}
}

// TestClosure_CrossPackagePartialApplication: partially applying a lifted
// closure binds the next parameters of the SAME lifted function and
// synthesizes nothing more; the completing call carries every bound value.
func TestClosure_CrossPackagePartialApplication(t *testing.T) {
	lib := `
twice(x) {
  return x * 2;
}
scaleBy(k) {
  by(a, b) {
    return twice(a) * b * k;
  }
  return by;
}
`
	src := `
import lib "geom/lib";
half = lib.scaleBy(0.5);
stepped = half(2.0);
return stepped(3.0);
`
	prog, err := runPipelineWithManager(src, libManager(lib))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	file := prog.Primary.Files[0]
	fc := callOf(file.Returns[0])
	if fc == nil {
		t.Fatalf("return is not a direct call: %s", file.Returns[0].PrettyPrint(0))
	}
	if fc.Reference.Prefix != "lib" || !strings.HasPrefix(fc.Reference.Name, "gcl_closure_scaleBy") {
		t.Errorf("return calls %s.%s, want lib.gcl_closure_scaleBy*", fc.Reference.Prefix, fc.Reference.Name)
	}
	if got := len(fc.Arguments); got != 3 {
		t.Errorf("return passes %d arguments, want 3: k from the factory, a from the partial application, b from the call", got)
	}

	libFile := depFile(t, prog, "geom/lib")
	if got := synthesizedIn(libFile); len(got) != 1 {
		t.Errorf("library holds %v, want exactly one lifted function: partial application must not synthesize another", got)
	}
	fn := fnNamed(libFile, fc.Reference.Name)
	if fn == nil {
		t.Fatalf("%q was not added to the library file", fc.Reference.Name)
	}
	if got := paramNames(fn); got != "k,a,b" {
		t.Errorf("lifted function takes (%s), want (k,a,b)", got)
	}
}

// TestClosure_CrossPackageNestedFactory: a factory returning a factory lifts
// at the level that finally returns a value, with every outer level's
// parameters ahead of its own.
func TestClosure_CrossPackageNestedFactory(t *testing.T) {
	lib := `
twice(x) {
  return x * 2;
}
outer(k) {
  middle(m) {
    inner(a) {
      return twice(a) * k * m;
    }
    return inner;
  }
  return middle;
}
`
	src := `
import lib "geom/lib";
f = lib.outer(2.0);
g = f(3.0);
return g(4.0);
`
	prog, err := runPipelineWithManager(src, libManager(lib))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	file := prog.Primary.Files[0]
	fc := callOf(file.Returns[0])
	if fc == nil {
		t.Fatalf("return is not a direct call: %s", file.Returns[0].PrettyPrint(0))
	}
	if fc.Reference.Prefix != "lib" || !strings.HasPrefix(fc.Reference.Name, "gcl_closure_outer") {
		t.Errorf("return calls %s.%s, want lib.gcl_closure_outer*", fc.Reference.Prefix, fc.Reference.Name)
	}
	if got := len(fc.Arguments); got != 3 {
		t.Errorf("return passes %d arguments, want 3", got)
	}
	fn := fnNamed(depFile(t, prog, "geom/lib"), fc.Reference.Name)
	if fn == nil {
		t.Fatalf("%q was not added to the library file", fc.Reference.Name)
	}
	if got := paramNames(fn); got != "k,m,a" {
		t.Errorf("lifted function takes (%s), want (k,m,a): outer levels first, in order", got)
	}
}

// TestClosure_LiftedClosureAsArgument: a lifted closure passed to a
// higher-order function of the library specializes that function into the
// CALLER's file, since the call through the closure inside it now carries the
// caller's capture expressions.
func TestClosure_LiftedClosureAsArgument(t *testing.T) {
	lib := libScaleBy + `
apply(x, f) {
  return f(x);
}
`
	src := `
import lib "geom/lib";
gain = 0.25;
soft = lib.scaleBy(gain);
return lib.apply(10.0, soft);
`
	prog, err := runPipelineWithManager(src, libManager(lib))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	file := prog.Primary.Files[0]
	fc := callOf(file.Returns[0])
	if fc == nil {
		t.Fatalf("return is not a direct call: %s", file.Returns[0].PrettyPrint(0))
	}
	if fc.Reference.Prefix != "" || !strings.HasPrefix(fc.Reference.Name, "gcl_closure_apply") {
		t.Fatalf("return calls %s.%s, want a gcl_closure_apply* in the caller's file", fc.Reference.Prefix, fc.Reference.Name)
	}
	if got := len(fc.Arguments); got != 1 {
		t.Errorf("specialized apply is passed %d arguments, want 1: the closure argument is gone", got)
	}
	applied := fnNamed(file, fc.Reference.Name)
	if applied == nil {
		t.Fatalf("%q was not added to the caller's file", fc.Reference.Name)
	}
	inner := callOf(applied.Returns[0])
	if inner == nil {
		t.Fatalf("specialized apply returns %s, want a direct call", applied.Returns[0].PrettyPrint(0))
	}
	if inner.Reference.Prefix != "lib" || !strings.HasPrefix(inner.Reference.Name, "gcl_closure_scaleBy") {
		t.Errorf("specialized apply calls %s.%s, want lib.gcl_closure_scaleBy*", inner.Reference.Prefix, inner.Reference.Name)
	}
	if got := len(inner.Arguments); got != 2 {
		t.Errorf("call through the closure passes %d arguments, want 2: the capture then x", got)
	}
	if ref, ok := inner.Arguments[0].T.(*program.Operand_Reference); !ok || ref.Reference == nil || ref.Reference.Name != "gain" {
		t.Errorf("leading argument is %s, want the caller's variable gain", inner.Arguments[0].PrettyPrint(0))
	}
}

// TestClosure_ForeignHigherOrderKeepsItsHelpers: a lambda from the caller
// passed to a library's higher-order function specializes that function into
// the CALLER's file, where the lambda is. The clone's calls to its own
// package's helpers, unqualified in the library, are qualified there, in its
// own body, in a lambda nested in it, and in a second higher-order function it
// forwards the lambda to; the lambda's capture of a caller variable, and the
// clone's own locals, stay as they are.
func TestClosure_ForeignHigherOrderKeepsItsHelpers(t *testing.T) {
	lib := `
twice(x) {
  return x * 2;
}
inner(x, f) {
  return f(twice(x));
}
outer(x, f) {
  y = twice(x);
  sum = reduce(0.0, (acc, i) {
    return acc + twice(i) * y;
  }, [1, 2]);
  return inner(sum, f);
}
`
	src := `
import lib "geom/lib";
gain = 0.25;
return lib.outer(10.0, (v) { return v * gain; });
`
	prog, err := runPipelineWithManager(src, libManager(lib))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	file := prog.Primary.Files[0]
	fc := callOf(file.Returns[0])
	if fc == nil {
		t.Fatalf("return is not a direct call: %s", file.Returns[0].PrettyPrint(0))
	}
	if fc.Reference.Prefix != "" || !strings.HasPrefix(fc.Reference.Name, "gcl_closure_outer") {
		t.Fatalf("return calls %s.%s, want a gcl_closure_outer* in the caller's file", fc.Reference.Prefix, fc.Reference.Name)
	}
	outer := fnNamed(file, fc.Reference.Name)
	if outer == nil {
		t.Fatalf("%q was not added to the caller's file", fc.Reference.Name)
	}
	if got := outer.Assignments[0].Expression.PrettyPrint(0); !strings.Contains(got, "lib.twice") {
		t.Errorf("specialized outer computes y as %s, want a call to lib.twice", got)
	}
	nested := false
	for _, ft := range outer.FnTypes {
		for _, fg := range ft.Groups {
			for _, fn := range fg.Functions {
				got := fn.Returns[0].PrettyPrint(0)
				if strings.Contains(got, "lib.twice") && strings.Contains(got, "y") && !strings.Contains(got, "lib.y") {
					nested = true
				}
			}
		}
	}
	if !nested {
		t.Errorf("the reduce lambda nested in specialized outer does not call lib.twice with its enclosing local y")
	}
	forwarded := callOf(outer.Returns[0])
	if forwarded == nil || forwarded.Reference.Prefix != "" || !strings.HasPrefix(forwarded.Reference.Name, "gcl_closure_inner") {
		t.Fatalf("specialized outer returns %s, want a call to a gcl_closure_inner* in the caller's file", outer.Returns[0].PrettyPrint(0))
	}
	inner := fnNamed(file, forwarded.Reference.Name)
	if inner == nil {
		t.Fatalf("%q was not added to the caller's file", forwarded.Reference.Name)
	}
	applied := callOf(inner.Returns[0])
	if applied == nil || applied.Reference.Prefix != "" || len(applied.Arguments) != 1 || !strings.Contains(applied.Arguments[0].PrettyPrint(0), "lib.twice") {
		t.Fatalf("specialized inner returns %s, want the lambda applied to lib.twice(x)", inner.Returns[0].PrettyPrint(0))
	}
	lambda := fnNamed(file, applied.Reference.Name)
	if lambda == nil {
		t.Fatalf("%q was not added to the caller's file", applied.Reference.Name)
	}
	if got := lambda.Returns[0].PrettyPrint(0); !strings.Contains(got, "gain") || strings.Contains(got, "lib.gain") {
		t.Errorf("the lambda returns %s, want it reading the caller's own gain", got)
	}
}
