package main

import (
	"fmt"
	"path"

	"gradientcontrol.com/omega/core/program"
	gsllibs "gradientcontrol.com/omega/core/stdlib"
)

// Operator is one @surface preset: the overload of a surface operator that
// takes only what a document supplies, the substrate material, its colour and
// the body's frame, and bakes the operator's other arguments. The Bodies panel
// wraps a body's surface with one and reads a body's stack of them back
// (doc/material_composition.md, "Design: operators as presets").
type Operator struct {
	// Call is the preset, qualified as a document writes it: "machined.turned".
	Call string `json:"call"`
	// Library is the import path of the library it comes from.
	Library string `json:"library"`
	// Arity is how many values the preset returns, read off its return list:
	// 1 (a material), 2 (material and colour) or 4 (and height and slope).
	Arity int `json:"arity"`
	// Slots are the positions of the preset's parameters.
	Slots Slots `json:"slots"`
	// Relief is whether the operator returns a height the layer has to apply
	// (a knurl's teeth): its full form returns something other than a literal
	// 0 as its third value. The panel then applies the layer's height to the
	// body's material as displacement. The operators whose relief rides in the
	// material return 0 there and need nothing applied.
	Relief bool `json:"relief,omitempty"`
	// Adjust is what the panel can open the layer up into: the operator's own
	// call with its annotated arguments lifted, as a @material's record is,
	// with ${layer} for the layer's binding and ${material}, ${color} and
	// ${frame} for the slots. Nil when the preset bakes only plain arguments.
	Adjust *Adjustable `json:"adjust,omitempty"`
}

// Slots are where a preset's parameters sit: the substrate material, always
// first; its colour and the body's frame, or -1 when the preset takes none.
type Slots struct {
	Material int `json:"material"`
	Color    int `json:"color"`
	Frame    int `json:"frame"`
}

// slotPlaceholders are the template placeholders for a preset's parameters, by
// the kind their annotation declares; the substrate's is unannotated.
var slotPlaceholders = map[string]string{"": "${material}", "color": "${color}", "frame": "${frame}"}

// ReadOperators reads every @surface preset out of the embedded libraries,
// keyed by package.symbol.
func ReadOperators() (map[string]*Operator, error) {
	out := map[string]*Operator{}
	for _, lib := range gsllibs.All() {
		pkg := program.GetStdPackage(lib.ImportPath)
		if pkg == nil {
			return nil, fmt.Errorf("%s: not a registered package", lib.ImportPath)
		}
		pkgName := path.Base(lib.ImportPath)
		for fi, file := range pkg.Files {
			if fi >= len(lib.Sources) {
				return nil, fmt.Errorf("%s: file %d has no source", lib.ImportPath, fi)
			}
			for _, a := range file.Annotations {
				if a.Kind != "surface" {
					continue
				}
				fn, err := presetOverload(file, a.Symbol)
				if err != nil {
					return nil, fmt.Errorf("%s.%s: %w", pkgName, a.Symbol, err)
				}
				op, err := operatorFor(pkg, file, lib.Sources[fi], pkgName, a.Symbol, fn)
				if err != nil {
					return nil, fmt.Errorf("%s.%s: %w", pkgName, a.Symbol, err)
				}
				op.Library = lib.ImportPath
				out[pkgName+"."+a.Symbol] = op
			}
		}
	}
	return out, nil
}

// presetOverload is the one overload of name whose every parameter is a slot.
// @surface marks the name, as every recorded annotation does, so the operator's
// full forms carry it too; the preset is told apart by what it takes.
func presetOverload(file *program.PackageFile, name string) (*program.Function, error) {
	var found *program.Function
	for _, ft := range file.FnTypes {
		if ft.Name != name {
			continue
		}
		for _, g := range ft.Groups {
			for _, fn := range g.Functions {
				if !takesOnlySlots(fn) {
					continue
				}
				if found != nil {
					return nil, fmt.Errorf("two overloads take only slots")
				}
				found = fn
			}
		}
	}
	if found == nil {
		return nil, fmt.Errorf("no overload takes only slots: the substrate first, then at most one @color and one @frame")
	}
	return found, nil
}

// takesOnlySlots: the substrate first, unannotated, then at most one @color
// and one @frame. A plain or ranged parameter anywhere else is an argument the
// document would have to supply, which a preset bakes.
func takesOnlySlots(fn *program.Function) bool {
	if len(fn.Parameters) == 0 {
		return false
	}
	seen := map[string]bool{}
	for i, p := range fn.Parameters {
		kind := paramKind(p)
		if i == 0 {
			if kind != "" {
				return false
			}
			continue
		}
		if (kind != "color" && kind != "frame") || seen[kind] {
			return false
		}
		seen[kind] = true
	}
	return true
}

func operatorFor(pkg *program.Package, file *program.PackageFile, src, pkgName, name string, fn *program.Function) (*Operator, error) {
	arity := len(fn.Returns)
	if arity != 1 && arity != 2 && arity != 4 {
		return nil, fmt.Errorf("returns %d values; an operator returns 1, 2 or 4", arity)
	}
	slots := Slots{Material: 0, Color: -1, Frame: -1}
	placeholders := map[string]string{}
	for i, p := range fn.Parameters {
		kind := paramKind(p)
		placeholders[p.Name] = slotPlaceholders[kind]
		switch kind {
		case "color":
			slots.Color = i
		case "frame":
			slots.Frame = i
		}
	}
	adjust, err := buildAdjust(pkg, file, src, pkgName, fn, placeholders)
	if err != nil {
		return nil, err
	}
	relief, err := returnsRelief(pkg, name, fn)
	if err != nil {
		return nil, err
	}
	return &Operator{Call: pkgName + "." + name, Arity: arity, Slots: slots, Relief: relief, Adjust: adjust}, nil
}

// returnsRelief reports whether the operator the preset destructures returns
// a height to apply: a four-value operator whose full form, the one the
// preset's body calls, returns anything but the literal 0 third.
func returnsRelief(pkg *program.Package, name string, preset *program.Function) (bool, error) {
	if len(preset.Returns) != 4 {
		return false, nil
	}
	for _, as := range preset.Assignments {
		fc := as.Expression.GetFnCall()
		if fc == nil || fc.Reference.Prefix != "" || fc.Reference.Name != name {
			continue
		}
		full := findFunction(pkg, name, len(fc.Arguments))
		if full == nil || len(full.Returns) != 4 {
			return false, fmt.Errorf("the preset calls %s with %d arguments, which is not a four-value operator of this package", name, len(fc.Arguments))
		}
		return !returnsZero(pkg, full, full.Returns[2], 0), nil
	}
	return false, fmt.Errorf("the preset's body does not call %s", name)
}

// returnsZero reports whether op, a return of fn, is the literal 0: itself,
// through a local of fn bound to one, or through a local bound by
// destructuring a call of this package to the same return of the callee, as
// the coats' levelled forms return the height of the plain forms they call.
func returnsZero(pkg *program.Package, fn *program.Function, op *program.Operand, depth int) bool {
	if op == nil || depth > 8 {
		return false
	}
	if _, literal := op.T.(*program.Operand_Value); literal {
		return op.GetValue() == 0
	}
	ref := op.GetReference()
	if ref == nil || ref.Prefix != "" {
		return false
	}
	for _, as := range fn.Assignments {
		for i, v := range as.Variables {
			if v.Name != ref.Name {
				continue
			}
			if len(as.Variables) == 1 {
				return returnsZero(pkg, fn, as.Expression, depth+1)
			}
			fc := as.Expression.GetFnCall()
			if fc == nil || fc.Reference.Prefix != "" {
				return false
			}
			callee := findFunction(pkg, fc.Reference.Name, len(fc.Arguments))
			if callee == nil || i >= len(callee.Returns) {
				return false
			}
			return returnsZero(pkg, callee, callee.Returns[i], depth+1)
		}
	}
	return false
}

// ProbeDoc is a document that applies op to steel on a sphere, with the
// sphere's own frame for a frame slot, destructured to op's arity: what the
// panel's Wrap writes, so compiling it proves the preset, its arity and its
// slots at generation time rather than in the panel.
func ProbeDoc(op *Operator) string {
	n := 1
	for _, i := range []int{op.Slots.Color, op.Slots.Frame} {
		if i >= n {
			n = i + 1
		}
	}
	args := make([]string, n)
	args[op.Slots.Material] = "m"
	if op.Slots.Color >= 0 {
		args[op.Slots.Color] = "c"
	}
	if op.Slots.Frame >= 0 {
		args[op.Slots.Frame] = "[s.u, s.v]"
	}
	binds := []string{"a", "b", "h", "sl"}[:op.Arity]
	doc := "s = sphere(5);\nm, c = metal.steel();\n"
	doc += joinComma(binds) + " = " + op.Call + "(" + joinComma(args) + ");\n"
	doc += "s.material = a;\n"
	if op.Arity >= 2 {
		doc += "s.color = b;\n"
	}
	return doc + "return s;\n"
}

func joinComma(xs []string) string {
	out := ""
	for i, x := range xs {
		if i > 0 {
			out += ", "
		}
		out += x
	}
	return out
}
