package assembler_test

import (
	"strings"
	"testing"

	"github.com/Gradient-Control-Laboratories/omega-core/assembler"
)

// assembleErr assembles src and returns its error text, failing the test when
// it assembles.
func assembleErr(t *testing.T, src string) string {
	t.Helper()
	_, err := assembler.Assemble([]string{src})
	if err == nil {
		t.Fatalf("expected an error for:\n%s", src)
	}
	return err.Error()
}

func TestUndefinedNameSuggestsWithSignatureAndDoc(t *testing.T) {
	msg := assembleErr(t, "return cylindrZ(vector.zero, 1);")
	for _, want := range []string{
		"undefined identifier: cylindrZ; did you mean:",
		"cylinderZ(center, radius) — ",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("missing %q in:\n%s", want, msg)
		}
	}
}

func TestUndefinedLocalSuggestsLocal(t *testing.T) {
	msg := assembleErr(t, "radius = 2;\nreturn sphere(radus);")
	if !strings.Contains(msg, "did you mean:\n  radius") {
		t.Errorf("no suggestion of radius in:\n%s", msg)
	}
}

func TestPackageMemberSuggestsMember(t *testing.T) {
	msg := assembleErr(t, "return transform.translat(sphere(1), [1, 0, 0]);")
	for _, want := range []string{
		`name "translat" not found in package "transform"; did you mean:`,
		"transform.translate(shape, offset)",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("missing %q in:\n%s", want, msg)
		}
	}
}

func TestPackagePrefixSuggestsPackage(t *testing.T) {
	msg := assembleErr(t, "return transfrom.translate(sphere(1), [1, 0, 0]);")
	if !strings.Contains(msg, "transform (package)") {
		t.Errorf("no suggestion of the transform package in:\n%s", msg)
	}
}

func TestShortNamesAreNotGuessed(t *testing.T) {
	msg := assembleErr(t, "return y;")
	if strings.Contains(msg, "did you mean") {
		t.Errorf("a one-letter name got a guess:\n%s", msg)
	}
}

func TestAnnotationCallsAreNotSuggested(t *testing.T) {
	msg := assembleErr(t, "return transfrom.translate(sphere(1), [1, 0, 0]);")
	if strings.Contains(msg, "uiTransform") {
		t.Errorf("suggested the call an annotation compiles to:\n%s", msg)
	}
}

func TestAtan2CallNamesAtan(t *testing.T) {
	msg := assembleErr(t, "theta = atan2(p.y, p.x);\nreturn length(p) - 1 - 0.1 * theta;")
	if !strings.Contains(msg, "'atan2' is a reserved word, not a GSL function; use atan(y, x)") {
		t.Errorf("no atan(y, x) in:\n%s", msg)
	}
	if strings.Contains(msg, "atan2_") {
		t.Errorf("suggested renaming a call:\n%s", msg)
	}
}

func TestRemapRejectsATwoComponentPoint(t *testing.T) {
	msg := assembleErr(t, "return remap(sphere(1), p.xy);")
	for _, want := range []string{"must be a vec3, not 2 component(s)", "[p.x, p.y, 0]"} {
		if !strings.Contains(msg, want) {
			t.Errorf("missing %q in:\n%s", want, msg)
		}
	}
}

// A parameter may share a built-in's name: a call of that name still reaches
// the built-in, as a primitive's (max) always did.
func TestParameterNamedAfterABuiltin(t *testing.T) {
	for _, src := range []string{
		"f(length) { return length(p) - length; }\nreturn f(2);",
		"f(dot) { return dot(p, p) - dot; }\nreturn f(2);",
		"f(max) { return max(length(p) - 1, max); }\nreturn f(2);",
	} {
		if _, err := assembler.Assemble([]string{src}); err != nil {
			t.Errorf("%v\n%s", err, src)
		}
	}
}

// A function that calls a parameter needs a function for it; passing a value
// used to prune the function silently and report it undefined.
func TestFunctionParameterGivenAValue(t *testing.T) {
	msg := assembleErr(t, "g(op) { return op(1) + length(p); }\nreturn g(2);")
	for _, want := range []string{"g calls its parameter op as a function", "2 is not one"} {
		if !strings.Contains(msg, want) {
			t.Errorf("missing %q in:\n%s", want, msg)
		}
	}
	if strings.Contains(msg, "undefined identifier: g") {
		t.Errorf("still reported as undefined:\n%s", msg)
	}
}

// The documented higher-order pattern still specializes.
func TestFunctionParameterGivenABlend(t *testing.T) {
	src := "joined(blend) { return blend([sphere(1), sphere([2, 0, 0], 1)]); }\nreturn joined(unionRound(0.5));"
	if _, err := assembler.Assemble([]string{src}); err != nil {
		t.Errorf("%v", err)
	}
}
