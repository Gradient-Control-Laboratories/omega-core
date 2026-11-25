package closure_test

import (
	"strings"
	"testing"

	internalsource "github.com/Gradient-Control-Laboratories/omega-core/internal/source"
	"github.com/Gradient-Control-Laboratories/omega-core/program"
)

// A list of functions FORWARDED: a function that does not fold its parameter
// itself but hands it on to one that does. Shaft Modeling's shaft passes the
// hole lists its caller wrote straight into drill.hole. The list is still only
// visible at the outermost call, so the forwarding function has to be
// specialized against it exactly as a folding one is, and the fold it forwards
// to then finds the elements bound to the clone.

// innermostFold follows the call in op through hops specialized functions on
// file, each a call whose callee's return is the next, and returns the operand
// the last one returns: the unrolled fold.
func innermostFold(t *testing.T, file *program.PackageFile, op *program.Operand, hops int) *program.Operand {
	t.Helper()
	for i := 0; i < hops; i++ {
		fc := callOf(op)
		if fc == nil {
			t.Fatalf("hop %d: not a call: %T", i, op.T)
		}
		if !strings.HasPrefix(fc.Reference.Name, "gcl_closure_") {
			t.Fatalf("hop %d calls %q, want a specialization", i, fc.Reference.Name)
		}
		fn := fnNamed(file, fc.Reference.Name)
		if fn == nil {
			t.Fatalf("hop %d: %q was not registered on the file", i, fc.Reference.Name)
		}
		if len(fn.Returns) != 1 {
			t.Fatalf("hop %d: %q has %d returns, want 1", i, fc.Reference.Name, len(fn.Returns))
		}
		op = fn.Returns[0]
	}
	return op
}

// checkUnrolled fails unless op is a fold unrolled to depth calls, with no
// reduce left in it.
func checkUnrolled(t *testing.T, op *program.Operand, depth int) {
	t.Helper()
	if got := chainDepth(op); got != depth {
		t.Errorf("fold unrolled to depth %d, want %d", got, depth)
	}
	if hasOperation(op, program.Operation_REDUCE) {
		t.Error("a reduce survived in the specialized body")
	}
}

const forwardSamePackage = `
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
wrap(features) {
  return stack(features);
}
`

// One hop, one package: wrap forwards its list to stack's fold, which unrolls one
// call per element (TestClosure_FoldListForwarded checks the specialization).
func TestClosure_FoldListForwardedUnrolls(t *testing.T) {
	prog, err := runPipeline(forwardSamePackage + "return wrap([addBy(1.0), addBy(2.0), addBy(3.0)]);\n")
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	file := prog.Primary.Files[0]
	checkUnrolled(t, innermostFold(t, file, file.Returns[0], 2), 3)
}

// A lone function forwarded folds as a list of one, as it does when it is
// handed straight to the fold.
func TestClosure_FoldListForwardedSingleFunction(t *testing.T) {
	prog, err := runPipeline(forwardSamePackage + "return wrap(addBy(1.0));\n")
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	file := prog.Primary.Files[0]
	checkUnrolled(t, innermostFold(t, file, file.Returns[0], 2), 1)
}

// Two hops: outer forwards to wrap, which forwards to stack.
func TestClosure_FoldListForwardedTwice(t *testing.T) {
	src := forwardSamePackage + `
outer(features) {
  return wrap(features);
}
return outer([addBy(1.0), addBy(2.0)]);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	file := prog.Primary.Files[0]
	checkUnrolled(t, innermostFold(t, file, file.Returns[0], 3), 2)
}

// A forwarded list beside an ordinary parameter, which keeps its place.
func TestClosure_FoldListForwardedBesideAValue(t *testing.T) {
	src := forwardSamePackage + `
wrapScaled(scale, features) {
  return scale * stack(features);
}
return wrapScaled(2.0, [addBy(1.0), addBy(2.0)]);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	file := prog.Primary.Files[0]
	fc := callOf(file.Returns[0])
	if fc == nil || !strings.HasPrefix(fc.Reference.Name, "gcl_closure_wrapScaled") {
		t.Fatalf("return does not call a specialization of wrapScaled")
	}
	if len(fc.Arguments) != 1 {
		t.Errorf("specialized wrapScaled takes %d arguments, want 1 (the scale)", len(fc.Arguments))
	}
}

// libForward is libStack (fold_test.go) with a function of the same package
// that forwards to its fold.
const libForward = libStack + `
relay(features, seedVal) {
  return stack(features, seedVal);
}
`

// Across a package: the document hands a list to a library function that
// forwards it to that library's own fold.
func TestClosure_FoldListForwardedInLibrary(t *testing.T) {
	src := `
import lib "geom/lib";
return lib.relay([lib.addBy(1.0), lib.addBy(2.0)], 100.0);
`
	prog, err := runPipelineWithManager(src, libManager(libForward))
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	file := prog.Primary.Files[0]
	checkUnrolled(t, innermostFold(t, file, file.Returns[0], 2), 2)
}

// From one package into another's fold: the shaft-into-drill shape. The
// document's list goes to geom/relay, which hands it to geom/lib's fold.
func TestClosure_FoldListForwardedAcrossPackages(t *testing.T) {
	relay := `
import lib "geom/lib";
handOn(features) {
  return lib.stack(features, 100.0);
}
`
	m := &internalsource.MockManager{PkgSources: map[string][]string{
		"geom/lib":   {libStack},
		"geom/relay": {relay},
	}}
	src := `
import lib "geom/lib";
import relay "geom/relay";
return relay.handOn([lib.addBy(1.0), lib.addBy(2.0), lib.addBy(3.0)]);
`
	prog, err := runPipelineWithManager(src, m)
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	file := prog.Primary.Files[0]
	checkUnrolled(t, innermostFold(t, file, file.Returns[0], 2), 3)
}

// relayLib forwards to geom/lib's fold from another package: alone, beside a
// list, and under a remap of the fold's result.
const relayLib = `
import lib "geom/lib";
one(features) {
  return lib.stack(features, 100.0);
}
two(a, b) {
  return lib.stack(a, 100.0) + lib.stack(b, 1.0);
}
remapped(features) {
  return remap(lib.stack(features, 0.0), [p.x, -p.y, 0.0 - p.z]);
}
`

func relayManager() *internalsource.MockManager {
	return &internalsource.MockManager{PkgSources: map[string][]string{
		"geom/lib":   {libStack},
		"geom/relay": {relayLib},
	}}
}

// A LONE function forwarded across packages folds as a list of one, as a list
// forwarded does and as a lone function handed straight to the fold does: the
// end of drill.hole(bottom) inside shaft, given drill.centerHole(...) alone.
func TestClosure_FoldListForwardedSingleFunctionAcrossPackages(t *testing.T) {
	cases := map[string]struct {
		call  string
		hops  int
		depth int
	}{
		"alone":         {"relay.one(lib.addBy(1.0))", 2, 1},
		"under a remap": {"relay.remapped(lib.addBy(1.0))", 0, 0},
		"beside a list": {"relay.two(lib.addBy(1.0), [lib.addBy(3.0), lib.addBy(4.0)])", 0, 0},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			src := "import lib \"geom/lib\";\nimport relay \"geom/relay\";\nreturn " + c.call + ";\n"
			prog, err := runPipelineWithManager(src, relayManager())
			if err != nil {
				t.Fatalf("pipeline: %v", err)
			}
			if c.hops > 0 {
				file := prog.Primary.Files[0]
				checkUnrolled(t, innermostFold(t, file, file.Returns[0], c.hops), c.depth)
			}
		})
	}
}
