package closure_test

import (
	"strings"
	"testing"

	"gradientcontrol.com/omega/core/internal/source"
)

// A list of functions written INSIDE a function, whose elements read that
// function's parameters or locals. The fold's specialized clone is made where
// the list's elements resolve, but a clone is a function of its own, so
// without help the enclosing function's names do not reach it: the elements
// come out naming `d` in a body that has no `d`. These pin that the names the
// elements need arrive as parameters of the clone.

// TestClosure_FoldListReadsEnclosingParameter: a document function wrapping a
// library's fold with its own parameter — `h(d) { return drill.hole(drill.thru(d), 0) }`.
// The elements' factory is another package's, so its closures are lifted; a
// same-package closure with a non-literal capture is baked, which is a separate
// gap and not this one.
func TestClosure_FoldListReadsEnclosingParameter(t *testing.T) {
	src := `
import lib "geom/lib";
wrap(d) {
  return lib.stack([lib.addBy(d), lib.addBy(2.0)], 100.0);
}
return wrap(7.0);
`
	prog, err := runPipelineWithManager(src, libManager(libStack))
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	file := prog.Primary.Files[0]
	fc := callOf(fnNamed(file, "wrap").Returns[0])
	if fc == nil || !strings.HasPrefix(fc.Reference.Name, "gcl_closure_stack") {
		t.Fatalf("wrap does not call a specialization of stack")
	}
	specialized := fnNamed(file, fc.Reference.Name)
	if specialized == nil {
		t.Fatalf("%q was not registered on the file", fc.Reference.Name)
	}
	if got := paramNames(specialized); got != "d,seedVal" {
		t.Errorf("specialized stack takes %q, want \"d,seedVal\" — the enclosing parameter its elements read, then its own", got)
	}
	if len(fc.Arguments) != 2 {
		t.Fatalf("the call passes %d arguments, want 2", len(fc.Arguments))
	}
	if ref := fc.Arguments[0].GetReference(); ref == nil || ref.Name != "d" {
		t.Errorf("the call leads with %v, want wrap's own d", fc.Arguments[0])
	}
}

// TestClosure_FoldListReadsEnclosingLocal: a local of the enclosing function
// alongside a parameter, both read by the elements.
func TestClosure_FoldListReadsEnclosingLocal(t *testing.T) {
	src := `
import lib "geom/lib";
wrap(d) {
  e = d * 2.0;
  return lib.stack([lib.addBy(e), lib.addBy(d)], 1.0);
}
return wrap(7.0);
`
	prog, err := runPipelineWithManager(src, libManager(libStack))
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	file := prog.Primary.Files[0]
	fc := callOf(fnNamed(file, "wrap").Returns[0])
	specialized := fnNamed(file, fc.Reference.Name)
	if specialized == nil {
		t.Fatalf("%q was not registered on the file", fc.Reference.Name)
	}
	if got := paramNames(specialized); got != "e,d,seedVal" {
		t.Errorf("specialized stack takes %q, want \"e,d,seedVal\" — enclosing names in order of use, then its own", got)
	}
}

// TestClosure_FoldListEnclosingNameCollides: the enclosing function's name is
// the same as one of the clone's own parameters. The elements' reference is
// renamed, so the clone's own seedVal is still its own.
func TestClosure_FoldListEnclosingNameCollides(t *testing.T) {
	src := `
import lib "geom/lib";
wrap(seedVal) {
  return lib.stack([lib.addBy(seedVal), lib.addBy(1.0)], 5.0);
}
return wrap(7.0);
`
	prog, err := runPipelineWithManager(src, libManager(libStack))
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	file := prog.Primary.Files[0]
	fc := callOf(fnNamed(file, "wrap").Returns[0])
	specialized := fnNamed(file, fc.Reference.Name)
	if specialized == nil {
		t.Fatalf("%q was not registered on the file", fc.Reference.Name)
	}
	if got := paramNames(specialized); got != "seedVal_enc,seedVal" {
		t.Errorf("specialized stack takes %q, want \"seedVal_enc,seedVal\"", got)
	}
	if ref := fc.Arguments[0].GetReference(); ref == nil || ref.Name != "seedVal" {
		t.Errorf("the call leads with %v, want wrap's own seedVal", fc.Arguments[0])
	}
	if v := fc.Arguments[1].GetValue(); v != 5.0 {
		t.Errorf("the call's second argument is %v, want the literal seed 5", fc.Arguments[1])
	}
}

// libWrap is a LIBRARY function that folds another library's list with its own
// parameter — fastener building a drill.hole from a screw size.
const libWrap = `
import lib "geom/lib";
wrap(d) {
  return lib.stack([lib.addBy(d), lib.addBy(1.0)], 0.0);
}
`

// TestClosure_FoldListInLibraryReadsItsParameter: the cross-package case, a
// library wrapping another library's fold, called from a document.
func TestClosure_FoldListInLibraryReadsItsParameter(t *testing.T) {
	src := `
import w "geom/wrap";
return w.wrap(7.0);
`
	mgr := &source.MockManager{PkgSources: map[string][]string{
		"geom/lib":  {libStack},
		"geom/wrap": {libWrap},
	}}
	prog, err := runPipelineWithManager(src, mgr)
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	wrapF := depFile(t, prog, "geom/wrap")
	fc := callOf(fnNamed(wrapF, "wrap").Returns[0])
	if fc == nil || !strings.HasPrefix(fc.Reference.Name, "gcl_closure_stack") {
		t.Fatalf("wrap does not call a specialization of stack")
	}
	specialized := fnNamed(wrapF, fc.Reference.Name)
	if specialized == nil {
		t.Fatalf("%q was not registered in wrap's package, where its elements resolve", fc.Reference.Name)
	}
	if !strings.Contains(paramNames(specialized), "d") {
		t.Errorf("specialized stack takes %q, want wrap's d among them", paramNames(specialized))
	}
}
