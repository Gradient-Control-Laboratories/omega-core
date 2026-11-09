package property

import (
	"fmt"
	"testing"

	"gradientcontrol.com/omega/core/program"
)

// Whether a program lays a material is read by reachability, not by scanning
// every function a library holds: the material library's own one-argument
// withLay(m) preset calls withLay, and a document that never reaches it must
// not pay for a blended lay row (doc/material_composition.md, "Design:
// operators as presets").

// layLib is a library in the shape of the material library around withLay:
// withLay(m, direction, anisotropy) builds the row itself and calls nothing,
// withLay(m) is the preset that calls it, coat(m) calls it too, make() calls
// nothing, and the constant `laid` is built by a call to it.
func layLib() *program.Package {
	fn := func(body *program.Operand, params ...string) *program.Function {
		ps := make([]*program.Parameter, len(params))
		for i, p := range params {
			ps[i] = &program.Parameter{Name: p}
		}
		return &program.Function{Parameters: ps, Returns: []*program.Operand{body}}
	}
	return &program.Package{Path: "lib", Files: []*program.PackageFile{{
		FnTypes: []*program.FnType{
			{Name: "withLay", Groups: []*program.FnGroup{
				{ParamCount: 3, Functions: []*program.Function{fn(ref("m"), "m", "direction", "anisotropy")}},
				{ParamCount: 1, Functions: []*program.Function{fn(call("withLay", ref("m"), num(1), num(0.8)), "m")}},
			}},
			{Name: "coat", Groups: []*program.FnGroup{
				{ParamCount: 1, Functions: []*program.Function{fn(call("withLay", ref("m"), num(0), num(0.5)), "m")}},
			}},
			{Name: "make", Groups: []*program.FnGroup{
				{ParamCount: 0, Functions: []*program.Function{fn(num(1))}},
			}},
		},
		Assignments: []*program.Assignment{{
			Variables:  []*program.Variable{{Name: "laid"}},
			Expression: call("withLay", num(1), num(0), num(0.5)),
		}},
	}}}
}

func layImport(path string) (*program.Package, error) {
	if path == "lib" {
		return layLib(), nil
	}
	return nil, fmt.Errorf("no package %q", path)
}

func qualRef(name string) *program.Operand {
	return &program.Operand{T: &program.Operand_Reference{Reference: &program.Reference{Prefix: "lib", Name: name}}}
}

// A document that calls only lib.make() reaches no withLay, although the
// library holds a preset that calls one.
func TestUsesCallIgnoresAnUnreachedPreset(t *testing.T) {
	doc := importPkg(qualCall("make"))
	if usesCall(doc, layImport, layFunction) {
		t.Error("lib.withLay(m), which nothing calls, was counted as a lay")
	}
}

// A function the document declares and never calls does not count either.
func TestUsesCallIgnoresAnUnreachedLocalFunction(t *testing.T) {
	doc := importPkg(qualCall("make"))
	doc.Files[0].FnTypes = []*program.FnType{{Name: "unused", Groups: []*program.FnGroup{
		{ParamCount: 1, Functions: []*program.Function{{
			Parameters: []*program.Parameter{{Name: "m"}},
			Returns:    []*program.Operand{qualCall("withLay", ref("m"), num(0), num(1))},
		}}},
	}}}
	if usesCall(doc, layImport, layFunction) {
		t.Error("a local function nothing calls was counted as a lay")
	}
}

// The calls that do reach one: by name, through the preset, through another
// library function, and through a constant the library builds with one.
func TestUsesCallFollowsEveryPath(t *testing.T) {
	cases := map[string]*program.Operand{
		"direct call":          qualCall("withLay", num(1), num(0), num(1)),
		"the preset":           qualCall("withLay", num(1)),
		"a library function":   qualCall("coat", num(1)),
		"a library constant":   qualRef("laid"),
		"inside an expression": mul(num(2), qualCall("coat", num(1))),
	}
	for name, ret := range cases {
		if !usesCall(importPkg(ret), layImport, layFunction) {
			t.Errorf("%s: the lay was not found", name)
		}
	}
}

// Without an import resolver only the document itself is read, as before.
func TestUsesCallWithoutImportsReadsTheDocumentAlone(t *testing.T) {
	if usesCall(importPkg(qualCall("coat", num(1))), nil, layFunction) {
		t.Error("followed an import with no resolver")
	}
	if !usesCall(importPkg(qualCall("withLay", num(1))), nil, layFunction) {
		t.Error("a direct call in the document was missed")
	}
}
