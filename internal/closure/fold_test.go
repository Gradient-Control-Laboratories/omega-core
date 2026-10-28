package closure_test

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/program"
)

// chainDepth counts how many nested calls an operand is: a reduce over N
// elements unrolls to N calls, each the accumulator argument of the next.
func chainDepth(op *program.Operand) int {
	depth := 0
	for {
		fc := callOf(op)
		if fc == nil || len(fc.Arguments) == 0 {
			return depth
		}
		depth++
		op = fc.Arguments[0]
	}
}

// foldNames returns the callee names of an unrolled reduce chain, outermost
// first.
func foldNames(op *program.Operand) []string {
	var names []string
	for {
		fc := callOf(op)
		if fc == nil || len(fc.Arguments) == 0 {
			return names
		}
		names = append(names, fc.Reference.Name)
		op = fc.Arguments[0]
	}
}

// TestClosure_FoldOverFunctionList is the shape the whole feature exists for:
// a library function that folds a list of closures its CALLER wrote. The list
// reaches it as an argument, so nothing in the fold's own file can see the
// elements — they are bound at the call site and the fold is unrolled against
// them.
func TestClosure_FoldOverFunctionList(t *testing.T) {
	src := `
seed = 100.0;
addBy(k) {
  step(acc) {
    return acc + k;
  }
  return step;
}
stack(features) {
  return reduce((acc, f) { return f(acc); }, features, seed);
}
return stack([addBy(1.0), addBy(2.0), addBy(3.0)]);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}

	file := prog.Primary.Files[0]
	if len(file.Returns) != 1 {
		t.Fatalf("expected one return, got %d", len(file.Returns))
	}

	// The call to stack is specialized and takes no arguments: the only one it
	// had was the list, and the list is gone into the body.
	fc := callOf(file.Returns[0])
	if fc == nil {
		t.Fatalf("return is not a call: %T", file.Returns[0].T)
	}
	if !strings.HasPrefix(fc.Reference.Name, "gcl_closure_stack") {
		t.Errorf("return calls %q, want a specialization of stack", fc.Reference.Name)
	}
	if len(fc.Arguments) != 0 {
		t.Errorf("specialized stack takes %d arguments, want 0", len(fc.Arguments))
	}

	specialized := fnNamed(file, fc.Reference.Name)
	if specialized == nil {
		t.Fatalf("%q was not registered on the file", fc.Reference.Name)
	}
	if got := paramNames(specialized); got != "" {
		t.Errorf("specialized stack has parameters %q, want none", got)
	}

	// Its body is three calls deep — one per element — and no reduce survives.
	if len(specialized.Returns) != 1 {
		t.Fatalf("specialized stack has %d returns, want 1", len(specialized.Returns))
	}
	if got := chainDepth(specialized.Returns[0]); got != 3 {
		t.Errorf("fold unrolled to depth %d, want 3 (one call per element)", got)
	}
	if hasOperation(specialized.Returns[0], program.Operation_REDUCE) {
		t.Error("a reduce survived in the specialized body")
	}

	// Each element got its own specialization, so the three closures cannot
	// share a capture.
	names := foldNames(specialized.Returns[0])
	seen := map[string]bool{}
	for _, n := range names {
		if seen[n] {
			t.Errorf("two elements folded through the same function %q", n)
		}
		seen[n] = true
	}
}

// libStack is a library that both MAKES the closures and FOLDS them. Its
// closure body calls a function of its own package unqualified, so a clone of
// that body into the caller's file could not resolve it; and its fold reaches
// the list through a parameter, so nothing in the library can see the
// elements. Both halves of the problem in one file.
const libStack = `
base(x) {
  return x * 2;
}
addBy(k) {
  step(acc) {
    return acc + base(k);
  }
  return step;
}
stack(features, seedVal) {
  return reduce((acc, f) { return f(acc); }, features, seedVal);
}
`

// callsTo collects every call to name anywhere in a file — top-level returns,
// assignments, and the bodies of all its functions — and returns the name each
// one's first argument references.
func callsTo(file *program.PackageFile, name string) []string {
	var found []string
	var walk func(op *program.Operand)
	walk = func(op *program.Operand) {
		if op == nil {
			return
		}
		switch v := op.T.(type) {
		case *program.Operand_FnCall:
			if v.FnCall == nil {
				return
			}
			if v.FnCall.Reference != nil && v.FnCall.Reference.Name == name && len(v.FnCall.Arguments) > 0 {
				if ref, ok := v.FnCall.Arguments[0].T.(*program.Operand_Reference); ok && ref.Reference != nil {
					found = append(found, ref.Reference.Name)
				} else {
					found = append(found, "<not a reference>")
				}
			}
			for _, arg := range v.FnCall.Arguments {
				walk(arg)
			}
		case *program.Operand_Expression:
			if v.Expression != nil {
				for _, sub := range v.Expression.Operands {
					walk(sub)
				}
			}
		}
	}
	var walkFn func(fn *program.Function)
	walkFn = func(fn *program.Function) {
		for _, a := range fn.Assignments {
			walk(a.Expression)
		}
		for _, r := range fn.Returns {
			walk(r)
		}
		for _, ft := range fn.FnTypes {
			for _, fg := range ft.Groups {
				for _, nested := range fg.Functions {
					walkFn(nested)
				}
			}
		}
	}
	for _, a := range file.Assignments {
		walk(a.Expression)
	}
	for _, r := range file.Returns {
		walk(r)
	}
	for _, ft := range file.FnTypes {
		for _, fg := range ft.Groups {
			for _, fn := range fg.Functions {
				walkFn(fn)
			}
		}
	}
	return found
}

// TestClosure_FoldOverFunctionListCrossPackage is the real case. The closures
// come from another package and capture the CALLER's variables, so they have
// to be lifted — the captures travel as leading arguments — and the fold that
// consumes them is in that other package too, so the list has to be bound at
// the call site. One lifted function serves both closures: they differ in
// their arguments, not in their function.
func TestClosure_FoldOverFunctionListCrossPackage(t *testing.T) {
	src := `
import lib "geom/lib";
gain = 1.5;
bias = 0.5;
start = 100.0;
return lib.stack([lib.addBy(gain), lib.addBy(bias)], start);
`
	prog, err := runPipelineWithManager(src, libManager(libStack))
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}

	file := prog.Primary.Files[0]
	libF := depFile(t, prog, "geom/lib")

	// addBy's body stayed in the library, as one lifted function taking its
	// capture ahead of its own parameter.
	var liftedName string
	for _, n := range synthesizedIn(libF) {
		if strings.HasPrefix(n, "gcl_closure_addBy") {
			if liftedName != "" {
				t.Errorf("two lifted functions for addBy (%s and %s); one should serve both closures", liftedName, n)
			}
			liftedName = n
		}
	}
	if liftedName == "" {
		t.Fatalf("addBy was not lifted into the library; it has %v", synthesizedIn(libF))
	}
	if got := paramNames(fnNamed(libF, liftedName)); got != "k,acc" {
		t.Errorf("lifted %s takes %q, want \"k,acc\" — the capture ahead of the closure's own parameter", liftedName, got)
	}

	// Nothing of addBy's body was cloned into the caller.
	for _, n := range synthesizedIn(file) {
		if strings.HasPrefix(n, "gcl_closure_addBy") {
			t.Errorf("the caller's file gained %q; a lifted body never leaves its own package", n)
		}
	}

	// Both closures call that one function, each with its own capture.
	got := callsTo(file, liftedName)
	if len(got) != 2 {
		t.Fatalf("found %d calls to %s, want 2: %v", len(got), liftedName, got)
	}
	if !(got[0] == "gain" && got[1] == "bias") && !(got[0] == "bias" && got[1] == "gain") {
		t.Errorf("the two calls lead with %v, want gain and bias — each closure's own capture", got)
	}

	// The fold itself is gone, and the specialized stack keeps only the
	// parameter that was not a list.
	fc := callOf(file.Returns[0])
	if fc == nil || !strings.HasPrefix(fc.Reference.Name, "gcl_closure_stack") {
		t.Fatalf("return does not call a specialization of stack")
	}
	if len(fc.Arguments) != 1 {
		t.Errorf("specialized stack takes %d arguments, want 1 (seedVal)", len(fc.Arguments))
	}
	specialized := fnNamed(file, fc.Reference.Name)
	if specialized == nil {
		t.Fatalf("%q was not registered on the caller's file", fc.Reference.Name)
	}
	if hasOperation(specialized.Returns[0], program.Operation_REDUCE) {
		t.Error("a reduce survived in the specialized body")
	}
	if got := chainDepth(specialized.Returns[0]); got != 2 {
		t.Errorf("fold unrolled to depth %d, want 2", got)
	}
}

// hasOperation reports whether op contains an expression with the given
// operation anywhere beneath it.
func hasOperation(op *program.Operand, want program.Operation) bool {
	if op == nil {
		return false
	}
	switch v := op.T.(type) {
	case *program.Operand_Expression:
		if v.Expression == nil {
			return false
		}
		if v.Expression.Operation == want {
			return true
		}
		for _, sub := range v.Expression.Operands {
			if hasOperation(sub, want) {
				return true
			}
		}
	case *program.Operand_FnCall:
		if v.FnCall == nil {
			return false
		}
		for _, arg := range v.FnCall.Arguments {
			if hasOperation(arg, want) {
				return true
			}
		}
	}
	return false
}

// TestClosure_FoldListBoundToName covers a list written to a name before it is
// folded, rather than inline at the call. The elements have to be found
// through the assignment.
func TestClosure_FoldListBoundToName(t *testing.T) {
	src := `
seed = 100.0;
addBy(k) {
  step(acc) {
    return acc + k;
  }
  return step;
}
stack(features) {
  return reduce((acc, f) { return f(acc); }, features, seed);
}
fs = [addBy(1.0), addBy(2.0)];
return stack(fs);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}

	file := prog.Primary.Files[0]
	fc := callOf(file.Returns[0])
	if fc == nil || !strings.HasPrefix(fc.Reference.Name, "gcl_closure_stack") {
		t.Fatalf("return does not call a specialization of stack")
	}
	specialized := fnNamed(file, fc.Reference.Name)
	if specialized == nil {
		t.Fatalf("%q was not registered on the file", fc.Reference.Name)
	}
	if got := chainDepth(specialized.Returns[0]); got != 2 {
		t.Errorf("fold unrolled to depth %d, want 2", got)
	}
	if hasOperation(specialized.Returns[0], program.Operation_REDUCE) {
		t.Error("a reduce survived in the specialized body")
	}
}

// TestClosure_MapOverFunctionList covers the other direction: many functions,
// one argument. It falls out of the same unroll, and the shaft's flange hole
// patterns will want it.
func TestClosure_MapOverFunctionList(t *testing.T) {
	src := `
addBy(k) {
  step(acc) {
    return acc + k;
  }
  return step;
}
spread(fns) {
  return map((f) { return f(10.0); }, fns);
}
return spread([addBy(1.0), addBy(2.0)]);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}

	file := prog.Primary.Files[0]
	fc := callOf(file.Returns[0])
	if fc == nil || !strings.HasPrefix(fc.Reference.Name, "gcl_closure_spread") {
		t.Fatalf("return does not call a specialization of spread")
	}
	specialized := fnNamed(file, fc.Reference.Name)
	if specialized == nil {
		t.Fatalf("%q was not registered on the file", fc.Reference.Name)
	}
	if hasOperation(specialized.Returns[0], program.Operation_MAP) {
		t.Error("a map survived in the specialized body")
	}
	// The result is a list of two direct calls, one per element.
	expr, ok := specialized.Returns[0].T.(*program.Operand_Expression)
	if !ok || expr.Expression == nil || expr.Expression.Operation != program.Operation_COLLECT {
		t.Fatalf("map did not unroll to a list, got %T", specialized.Returns[0].T)
	}
	if n := len(expr.Expression.Operands); n != 2 {
		t.Fatalf("map unrolled to %d elements, want 2", n)
	}
	seen := map[string]bool{}
	for i, el := range expr.Expression.Operands {
		elCall := callOf(el)
		if elCall == nil {
			t.Fatalf("element %d is not a call: %T", i, el.T)
		}
		if seen[elCall.Reference.Name] {
			t.Errorf("two elements mapped through the same function %q", elCall.Reference.Name)
		}
		seen[elCall.Reference.Name] = true
	}
}

// TestClosure_TemplateIsPruned checks that the function declaring the list
// parameter does not survive unspecialized. It still folds a parameter, which
// no backend can iterate, so it is a template like any other higher-order one.
func TestClosure_TemplateIsPruned(t *testing.T) {
	src := `
seed = 100.0;
addBy(k) {
  step(acc) {
    return acc + k;
  }
  return step;
}
stack(features) {
  return reduce((acc, f) { return f(acc); }, features, seed);
}
return stack([addBy(1.0), addBy(2.0)]);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	if fn := fnNamed(prog.Primary.Files[0], "stack"); fn != nil {
		t.Error("the unspecialized stack template survived; it folds a parameter no backend can iterate")
	}
}

// TestClosure_ErrorReturnFunctionList checks that a list of functions cannot
// escape as a value. It only means something where it is folded.
func TestClosure_ErrorReturnFunctionList(t *testing.T) {
	src := `
addBy(k) {
  step(acc) {
    return acc + k;
  }
  return step;
}
one(x) {
  return x + 1;
}
two(x) {
  return x + 2;
}
return [one, two];
`
	_, err := runPipeline(src)
	if err == nil {
		t.Fatal("returning a list of functions was accepted")
	}
	if !strings.Contains(err.Error(), "cannot return a function") {
		t.Errorf("error was %q, want it to say a function cannot be returned", err)
	}
}

// TestClosure_FoldListForwarded covers a list handed on through an
// intermediate function before it is folded. The binding travels with the
// specialization, so the fold at the end still sees the elements.
func TestClosure_FoldListForwarded(t *testing.T) {
	src := `
seed = 100.0;
addBy(k) {
  step(acc) {
    return acc + k;
  }
  return step;
}
stack(features) {
  return reduce((acc, f) { return f(acc); }, features, seed);
}
forward(fs) {
  return stack(fs);
}
return forward([addBy(1.0), addBy(2.0)]);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	file := prog.Primary.Files[0]
	fc := callOf(file.Returns[0])
	if fc == nil || !strings.HasPrefix(fc.Reference.Name, "gcl_closure_forward") {
		t.Fatalf("return does not call a specialization of forward")
	}
	if len(fc.Arguments) != 0 {
		t.Errorf("specialized forward takes %d arguments, want 0", len(fc.Arguments))
	}
	for _, fn := range []string{"forward", "stack"} {
		if fnNamed(file, fn) != nil {
			t.Errorf("the unspecialized %s template survived", fn)
		}
	}
}

// TestClosure_ErrorFunctionListThroughCall covers a list that cannot be
// resolved at all — it comes back from a call, so there is no point at which
// the compiler sees its elements.
func TestClosure_ErrorFunctionListThroughCall(t *testing.T) {
	src := `
seed = 100.0;
addBy(k) {
  step(acc) {
    return acc + k;
  }
  return step;
}
stack(features) {
  return reduce((acc, f) { return f(acc); }, features, seed);
}
pick(n) {
  return [addBy(n), addBy(n + 1.0)];
}
return stack(pick(1.0));
`
	_, err := runPipeline(src)
	if err == nil {
		t.Fatal("a fold over an unresolvable list of functions was accepted")
	}
	if !strings.Contains(err.Error(), "has to be visible where it is folded") {
		t.Errorf("error was %q, want it to say the list has to be visible where it is folded", err)
	}
}

// TestClosure_FoldSingleFunction covers a fold of ONE function. A collection
// literal needs two or more elements, so [f] does not parse; the singleton is
// spelled as the bare function, which the language already treats as a list of
// length one.
func TestClosure_FoldSingleFunction(t *testing.T) {
	src := `
seed = 100.0;
addBy(k) {
  step(acc) {
    return acc + k;
  }
  return step;
}
stack(features) {
  return reduce((acc, f) { return f(acc); }, features, seed);
}
return stack(addBy(7.0));
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	file := prog.Primary.Files[0]
	fc := callOf(file.Returns[0])
	if fc == nil || !strings.HasPrefix(fc.Reference.Name, "gcl_closure_stack") {
		t.Fatalf("return does not call a specialization of stack")
	}
	specialized := fnNamed(file, fc.Reference.Name)
	if specialized == nil {
		t.Fatalf("%q was not registered on the file", fc.Reference.Name)
	}
	if got := chainDepth(specialized.Returns[0]); got != 1 {
		t.Errorf("fold unrolled to depth %d, want 1", got)
	}
	if hasOperation(specialized.Returns[0], program.Operation_REDUCE) {
		t.Error("a reduce survived in the specialized body")
	}
}

// callPrefixes returns the prefix of every call to a gcl_closure_ name whose
// suffix starts with `stem`, anywhere in a file.
func callPrefixes(file *program.PackageFile, stem string) []string {
	var found []string
	var walk func(op *program.Operand)
	walk = func(op *program.Operand) {
		if op == nil {
			return
		}
		switch v := op.T.(type) {
		case *program.Operand_FnCall:
			if v.FnCall == nil {
				return
			}
			if r := v.FnCall.Reference; r != nil && strings.HasPrefix(r.Name, "gcl_closure_"+stem) {
				found = append(found, r.Prefix)
			}
			for _, arg := range v.FnCall.Arguments {
				walk(arg)
			}
		case *program.Operand_Expression:
			if v.Expression != nil {
				for _, sub := range v.Expression.Operands {
					walk(sub)
				}
			}
		}
	}
	var walkFn func(fn *program.Function)
	walkFn = func(fn *program.Function) {
		for _, a := range fn.Assignments {
			walk(a.Expression)
		}
		for _, r := range fn.Returns {
			walk(r)
		}
		for _, ft := range fn.FnTypes {
			for _, fg := range ft.Groups {
				for _, nested := range fg.Functions {
					walkFn(nested)
				}
			}
		}
	}
	for _, r := range file.Returns {
		walk(r)
	}
	for _, ft := range file.FnTypes {
		for _, fg := range ft.Groups {
			for _, fn := range fg.Functions {
				walkFn(fn)
			}
		}
	}
	return found
}

// TestClosure_FoldLiteralCapturesKeepPrefix covers the other half of the
// cross-package case. Closures whose captures are closed literals still BAKE
// into the library's own file — that is what keeps every existing program
// compiling to what it did — but the fold's body lives in the caller, so the
// calls into the library have to carry the prefix that reaches it.
func TestClosure_FoldLiteralCapturesKeepPrefix(t *testing.T) {
	src := `
import lib "geom/lib";
start = 100.0;
return lib.stack([lib.addBy(1.0), lib.addBy(2.0)], start);
`
	prog, err := runPipelineWithManager(src, libManager(libStack))
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	file := prog.Primary.Files[0]
	libF := depFile(t, prog, "geom/lib")

	// Literal captures bake, so the bodies are in the library, not the caller.
	if n := len(synthesizedIn(libF)); n == 0 {
		t.Fatalf("nothing was synthesized into the library")
	}
	for _, n := range synthesizedIn(file) {
		if strings.HasPrefix(n, "gcl_closure_addBy") {
			t.Errorf("the caller's file gained %q; a literal capture bakes into the definer", n)
		}
	}
	// And the calls to them, made from a body that ended up in the caller,
	// carry the prefix.
	got := callPrefixes(file, "addBy")
	if len(got) != 2 {
		t.Fatalf("found %d calls to a synthesized addBy, want 2: %v", len(got), got)
	}
	for _, prefix := range got {
		if prefix != "lib" {
			t.Errorf("a call to the library's baked closure carries prefix %q, want \"lib\"", prefix)
		}
	}
}
