package closure_test

import (
	"strings"
	"testing"
)

// A closure made by a factory call whose argument reads a parameter or local
// of the function making the call — `h(k) { return box(c, s, intersectRound(k)) }`.
// The closure is handed to a higher-order template, whose specialized clone is
// a function of its own, so `k` does not reach it unless it is passed in. These
// pin that the capture is lifted and the name it reads arrives as a parameter
// of the clone, whether the factory is another package's or the caller's own.

// libApply makes closures and applies them, in one package.
const libApply = `
scaleBy(k) {
  by(a) {
    return a * k;
  }
  return by;
}
apply(op, x) {
  return op(x);
}
wrapped(amount) {
  return apply(scaleBy(amount), 3.0);
}
wrappedSameName(k) {
  return apply(scaleBy(k), 3.0);
}
`

// TestClosure_FactoryCaptureReadsEnclosingParameter: the factory is another
// package's, so its closure was already lifted; the clone of apply still has
// to be given the caller's `k`.
func TestClosure_FactoryCaptureReadsEnclosingParameter(t *testing.T) {
	src := `
import lib "geom/lib";
wrap(k) {
  return lib.apply(lib.scaleBy(k), 3.0);
}
return wrap(2.0);
`
	prog, err := runPipelineWithManager(src, libManager(libApply))
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	file := prog.Primary.Files[0]
	fc := callOf(fnNamed(file, "wrap").Returns[0])
	if fc == nil || !strings.HasPrefix(fc.Reference.Name, "gcl_closure_apply") {
		t.Fatalf("wrap does not call a specialization of apply")
	}
	specialized := fnNamed(file, fc.Reference.Name)
	if specialized == nil {
		t.Fatalf("%q was not registered on the file", fc.Reference.Name)
	}
	if got := paramNames(specialized); got != "k,x" {
		t.Errorf("specialized apply takes %q, want \"k,x\" — the enclosing parameter its closure reads, then its own", got)
	}
	if ref := fc.Arguments[0].GetReference(); ref == nil || ref.Name != "k" {
		t.Errorf("the call leads with %v, want wrap's own k", fc.Arguments[0])
	}
}

// TestClosure_FactoryCaptureReadsEnclosingLocal: a local of the enclosing
// function, computed from its parameter.
func TestClosure_FactoryCaptureReadsEnclosingLocal(t *testing.T) {
	src := `
import lib "geom/lib";
wrap(k) {
  s = k * 2.0;
  return lib.apply(lib.scaleBy(s), 1.0);
}
return wrap(2.0);
`
	prog, err := runPipelineWithManager(src, libManager(libApply))
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	file := prog.Primary.Files[0]
	fc := callOf(fnNamed(file, "wrap").Returns[0])
	specialized := fnNamed(file, fc.Reference.Name)
	if specialized == nil {
		t.Fatalf("%q was not registered on the file", fc.Reference.Name)
	}
	if got := paramNames(specialized); got != "s,x" {
		t.Errorf("specialized apply takes %q, want \"s,x\"", got)
	}
}

// TestClosure_SamePackageFactoryCaptureReadsEnclosing: a library function
// handing its own template a closure from its own factory, built from its own
// parameter. Same package, so nothing lifted it, and the capture was baked into
// a synthesized function as `k = amount`, where amount is undefined.
func TestClosure_SamePackageFactoryCaptureReadsEnclosing(t *testing.T) {
	src := `
import lib "geom/lib";
return lib.wrapped(2.0);
`
	prog, err := runPipelineWithManager(src, libManager(libApply))
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	libFile := depFile(t, prog, "geom/lib")
	fc := callOf(fnNamed(libFile, "wrapped").Returns[0])
	specialized := fnNamed(libFile, fc.Reference.Name)
	if specialized == nil {
		t.Fatalf("%q was not registered on the library's file", fc.Reference.Name)
	}
	if got := paramNames(specialized); got != "amount,x" {
		t.Errorf("specialized apply takes %q, want \"amount,x\"", got)
	}
}

// TestClosure_SamePackageCaptureSharesItsName: as above, with the enclosing
// parameter and the factory's parameter both named k. Baked, that is `k = k`,
// which links — it reads itself — and so compiled into the wrong program
// rather than failing. It must be passed in like any other.
func TestClosure_SamePackageCaptureSharesItsName(t *testing.T) {
	src := `
import lib "geom/lib";
return lib.wrappedSameName(2.0);
`
	prog, err := runPipelineWithManager(src, libManager(libApply))
	if err != nil {
		t.Fatalf("pipeline: %v", err)
	}
	libFile := depFile(t, prog, "geom/lib")
	fc := callOf(fnNamed(libFile, "wrappedSameName").Returns[0])
	specialized := fnNamed(libFile, fc.Reference.Name)
	if specialized == nil {
		t.Fatalf("%q was not registered on the library's file", fc.Reference.Name)
	}
	if got := paramNames(specialized); got != "k,x" {
		t.Errorf("specialized apply takes %q, want \"k,x\" — k passed in, not baked as k = k", got)
	}
}

// TestClosure_UnreachedCallerOfPrunedTemplate: the same library, with nothing
// calling wrapped. Every overload of apply and scaleBy is a template and is
// pruned, so wrapped — unreached, never specialized — would name a function
// that no longer exists, and the re-link would fail every document importing
// the library.
func TestClosure_UnreachedCallerOfPrunedTemplate(t *testing.T) {
	src := `
import lib "geom/lib";
return lib.apply(lib.scaleBy(2.0), 3.0);
`
	if _, err := runPipelineWithManager(src, libManager(libApply)); err != nil {
		t.Fatalf("pipeline: %v", err)
	}
}
