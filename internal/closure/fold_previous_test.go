package closure_test

import (
	"strings"
	"testing"
)

// TestClosure_FoldPassesPreviousElement: a reduce over a list of functions
// whose callback declares a fourth parameter is handed the element before
// this one, as a function it can call, so a fold can ask an entry about its
// predecessor with constant arguments. The chain unrolls one call per
// element, each with four arguments; the first step, which has no
// predecessor, is handed its own element.
func TestClosure_FoldPassesPreviousElement(t *testing.T) {
	src := `
tag(k) {
  t(want) {
    return k;
  }
  return t;
}
stack(features) {
  return reduce((acc, f, i, prev) { return acc + prev(0.0) * 10.0 + f(0.0); }, features, 0.0);
}
return stack([tag(1.0), tag(2.0), tag(3.0)]);
`
	prog, err := runPipeline(src)
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	file := prog.Primary.Files[0]
	fc := callOf(file.Returns[0])
	if fc == nil {
		t.Fatalf("return is not a call: %T", file.Returns[0].T)
	}
	if !strings.HasPrefix(fc.Reference.Name, "gcl_closure_stack") {
		t.Fatalf("return calls %q, want a specialization of stack", fc.Reference.Name)
	}
	specialized := fnNamed(file, fc.Reference.Name)
	if specialized == nil {
		t.Fatalf("no function %q in the file", fc.Reference.Name)
	}
	body := specialized.Returns[0]
	if got := chainDepth(body); got != 3 {
		t.Errorf("the fold unrolled to %d calls, want 3", got)
	}
}
