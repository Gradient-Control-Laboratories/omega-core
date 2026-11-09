package main

import (
	"fmt"
	"math"
	"path"
	"regexp"
	"sort"
	"strings"

	// Registers every embedded library as a parsed program, which is where a
	// constructor's parameter annotations live.
	"gradientcontrol.com/omega/core/internal/eval"
	_ "gradientcontrol.com/omega/core/parser"
	"gradientcontrol.com/omega/core/program"
	gsllibs "gradientcontrol.com/omega/core/stdlib"
)

// Adjustable is what the Bodies panel needs to open a preset up: the
// constructor calls its body makes, the arguments of theirs that the library
// declared adjustable, and the body itself as a template the panel fills in
// for a particular body (doc/material_composition.md, "Design: parameters on
// materials").
//
// The panel never guesses an arity or a kind. The template already
// destructures what the constructor returns, exactly as the preset's author
// wrote it, and each parameter's kind is the annotation on the constructor's
// own parameter list.
type Adjustable struct {
	// Calls are the constructors the preset calls, qualified as a document
	// writes them: "coating.powderCoat".
	Calls []string `json:"calls"`
	// Parameters are the lifted arguments, in the order the template reads
	// them.
	Parameters []Parameter `json:"parameters"`
	// Template is the preset's body with its locals and the lifted arguments
	// replaced by placeholders: ${body} for the body's name, ${<name>} for
	// the variable the panel writes for a parameter. Each line is one
	// statement, the last two assigning .material and .color.
	Template string `json:"template"`
}

// Parameter is one lifted argument: the constructor parameter's name, the
// control kind its annotation declared, and the preset's own argument as the
// default.
type Parameter struct {
	Name string `json:"name"`
	// Kind is "range", "color", "point", "direction" or "transform".
	Kind string   `json:"kind"`
	Min  *float64 `json:"min,omitempty"`
	Max  *float64 `json:"max,omitempty"`
	// Call is the constructor the parameter belongs to, as Calls names it,
	// and Arg its position in that call's arguments, so a document's
	// expansion can be read back parameter by parameter.
	Call string `json:"call"`
	Arg  int    `json:"arg"`
	// Anchor is the Name of the point parameter a direction is drawn from,
	// when the constructor declares one and it is lifted too.
	Anchor string `json:"anchor,omitempty"`
	// Default is the argument's value: one number for a range, three for a
	// colour, point or direction, sixteen for a frame.
	Default []float64 `json:"default"`
}

// ReadAdjust reads, for every @material preset, what the panel can adjust,
// keyed by package.symbol. A preset whose constructors take no annotated
// parameters has no entry: the panel shows nothing to adjust.
func ReadAdjust() (map[string]*Adjustable, error) {
	out := map[string]*Adjustable{}
	for _, lib := range gsllibs.All() {
		pkg := program.GetStdPackage(lib.ImportPath)
		if pkg == nil {
			return nil, fmt.Errorf("%s: not a registered package", lib.ImportPath)
		}
		for fi, file := range pkg.Files {
			if fi >= len(lib.Sources) {
				return nil, fmt.Errorf("%s: file %d has no source", lib.ImportPath, fi)
			}
			for _, a := range file.Annotations {
				if a.Kind != "material" {
					continue
				}
				fn := zeroArgFunction(file, a.Symbol)
				if fn == nil {
					continue
				}
				adj, err := adjustFor(pkg, file, lib.Sources[fi], path.Base(lib.ImportPath), fn)
				if err != nil {
					return nil, fmt.Errorf("%s.%s: %w", path.Base(lib.ImportPath), a.Symbol, err)
				}
				if adj != nil {
					out[path.Base(lib.ImportPath)+"."+a.Symbol] = adj
				}
			}
		}
	}
	return out, nil
}

// zeroArgFunction is the preset's function: the overload of name that takes
// nothing, which is what a @material names.
func zeroArgFunction(file *program.PackageFile, name string) *program.Function {
	for _, ft := range file.FnTypes {
		if ft.Name != name {
			continue
		}
		for _, g := range ft.Groups {
			if g.ParamCount == 0 && len(g.Functions) > 0 {
				return g.Functions[0]
			}
		}
	}
	return nil
}

// edit is a replacement of one span of the source text.
type edit struct {
	from, to int
	text     string
}

// builder gathers a preset's lifted parameters and the edits that turn its
// body into the template.
type builder struct {
	pkg     *program.Package
	file    *program.PackageFile
	src     string
	pkgName string
	fn      *program.Function

	calls  []string
	params []Parameter
	edits  []edit
	// lifted maps a constructor parameter's name within one call to the
	// lifted Parameter's name, for anchors.
	liftedNames map[string]bool
}

func adjustFor(pkg *program.Package, file *program.PackageFile, src, pkgName string, fn *program.Function) (*Adjustable, error) {
	if len(fn.Returns) != 2 {
		return nil, nil
	}
	return buildAdjust(pkg, file, src, pkgName, fn, nil)
}

// buildAdjust is adjustFor over a preset's body: a @material preset's when
// slots is nil, a @surface preset's otherwise. slots maps the operator preset's
// parameters to their placeholders, ${material}, ${color} and ${frame}, which
// the panel fills with what the layer's call already reads. An operator's
// returns are bound to the layer, ${layer}, ${layer}Color, ${layer}Height and
// ${layer}Slope in the order the operators return them, and the first two are
// assigned to the body's properties.
func buildAdjust(pkg *program.Package, file *program.PackageFile, src, pkgName string, fn *program.Function, slots map[string]string) (*Adjustable, error) {
	// A body with its own functions or property statements is more than this
	// template carries; such a preset is left unadjustable rather than
	// half-copied.
	if len(fn.FnTypes) > 0 {
		return nil, nil
	}
	for _, as := range fn.Assignments {
		for _, v := range as.Variables {
			if len(v.Properties) > 0 {
				return nil, nil
			}
		}
	}

	b := &builder{pkg: pkg, file: file, src: src, pkgName: pkgName, fn: fn, liftedNames: map[string]bool{}}
	for _, as := range fn.Assignments {
		if err := b.walk(as.Expression); err != nil {
			return nil, err
		}
	}
	for _, r := range fn.Returns {
		if err := b.walk(r); err != nil {
			return nil, err
		}
	}
	if len(b.params) == 0 {
		return nil, nil
	}

	// The locals: the ones the return reads are the surface and its colour
	// (and an operator's height and slope), named as the panel names them; the
	// rest take the body's, or the layer's, name as a prefix.
	prefix := "${body}"
	returnNames := []string{"${body}Surface", "${body}SurfaceColor"}
	if slots != nil {
		prefix = "${layer}"
		returnNames = []string{"${layer}", "${layer}Color", "${layer}Height", "${layer}Slope"}
	}
	renames := map[string]string{}
	for _, as := range fn.Assignments {
		for _, v := range as.Variables {
			renames[v.Name] = prefix + strings.ToUpper(v.Name[:1]) + v.Name[1:]
		}
	}
	for i, r := range fn.Returns {
		ref, ok := r.T.(*program.Operand_Reference)
		local := ok && ref.Reference.Prefix == ""
		if local {
			_, local = renames[ref.Reference.Name]
		}
		if local && i < len(returnNames) {
			renames[ref.Reference.Name] = returnNames[i]
		} else if slots != nil {
			// An operator preset destructures its operator's result and
			// returns the locals, so its arity is read off the return list; a
			// return that forwards a call is refused rather than guessed at.
			return nil, fmt.Errorf("return %d is not a local the body binds", i)
		}
	}
	for param, placeholder := range slots {
		renames[param] = placeholder
	}

	var lines []string
	for _, as := range fn.Assignments {
		if as.Start == nil || as.End == nil {
			return nil, fmt.Errorf("an assignment has no span")
		}
		lines = append(lines, b.render(int(as.Start.Offset), int(as.End.Offset), renames)+";")
	}
	props := []string{"material", "color"}
	for i, r := range fn.Returns {
		if i >= len(props) {
			break
		}
		if r.Start == nil || r.End == nil {
			return nil, fmt.Errorf("a return has no span")
		}
		lines = append(lines, fmt.Sprintf("${body}.%s = %s;", props[i], b.render(int(r.Start.Offset), int(r.End.Offset), renames)))
	}
	return &Adjustable{Calls: b.calls, Parameters: b.params, Template: strings.Join(lines, "\n")}, nil
}

// walk finds the constructor calls under op and lifts their annotated
// arguments. An argument that is lifted is replaced whole, so nothing under
// it is walked; one that is not is walked for calls of its own.
func (b *builder) walk(op *program.Operand) error {
	if op == nil {
		return nil
	}
	switch v := op.T.(type) {
	case *program.Operand_FnCall:
		fc := v.FnCall
		callee, qualified := b.resolve(fc.Reference, len(fc.Arguments))
		if qualified != "" && fc.Reference.Prefix == "" && op.Start != nil {
			// A same-package call is written qualified in a document.
			b.edits = append(b.edits, edit{from: int(op.Start.Offset), to: int(op.Start.Offset), text: packagePlaceholder + "."})
		}
		if callee != nil && qualified != "" {
			lifted := map[string]string{}
			var liftedParams []*program.Parameter
			var liftedArgs []*program.Operand
			for i, arg := range fc.Arguments {
				if i >= len(callee.Parameters) {
					break
				}
				p := callee.Parameters[i]
				kind := paramKind(p)
				if kind == "" {
					continue
				}
				values, ok := b.constant(arg, kind)
				if !ok {
					continue
				}
				name := b.uniqueName(p.Name)
				lifted[p.Name] = name
				param := Parameter{Name: name, Kind: kind, Default: values, Call: qualified, Arg: i}
				if kind == "range" {
					lo, hi, err := b.rangeBounds(p)
					if err != nil {
						return fmt.Errorf("%s(%s): %w", fc.Reference.Name, p.Name, err)
					}
					param.Min, param.Max = &lo, &hi
				}
				liftedParams = append(liftedParams, p)
				liftedArgs = append(liftedArgs, arg)
				b.params = append(b.params, param)
				b.edits = append(b.edits, edit{from: int(arg.Start.Offset), to: int(arg.End.Offset), text: "${" + name + "}"})
			}
			// A direction's anchor names a sibling parameter; it carries over
			// only when that sibling was lifted too, else the arrow is drawn
			// from the origin.
			for i, p := range liftedParams {
				if p.Anchor != "" {
					if anchor, ok := lifted[p.Anchor]; ok {
						b.params[len(b.params)-len(liftedParams)+i].Anchor = anchor
					}
				}
			}
			if len(liftedParams) > 0 {
				b.calls = appendUnique(b.calls, qualified)
			}
			for _, arg := range fc.Arguments {
				isLifted := false
				for _, l := range liftedArgs {
					if l == arg {
						isLifted = true
					}
				}
				if !isLifted {
					if err := b.walk(arg); err != nil {
						return err
					}
				}
			}
			return nil
		}
		for _, arg := range fc.Arguments {
			if err := b.walk(arg); err != nil {
				return err
			}
		}
	case *program.Operand_Expression:
		for _, o := range v.Expression.Operands {
			if err := b.walk(o); err != nil {
				return err
			}
		}
	case *program.Operand_Reference:
		// A same-package constant is written qualified in a document too.
		if v.Reference.Prefix == "" && op.Start != nil && b.isPackageName(v.Reference.Name) {
			b.edits = append(b.edits, edit{from: int(op.Start.Offset), to: int(op.Start.Offset), text: packagePlaceholder + "."})
		}
	}
	return nil
}

// paramKind is the control a constructor parameter declares: its annotation's
// kind, or "range" when it carries a @range, or "" for a plain parameter.
func paramKind(p *program.Parameter) string {
	if p.Kind != "" {
		return p.Kind
	}
	if p.Range != nil && !p.Range.IsSynthetic() {
		return "range"
	}
	return ""
}

// constant evaluates an argument when it is a compile-time constant of the
// size its kind wants.
func (b *builder) constant(arg *program.Operand, kind string) ([]float64, bool) {
	if arg.Start == nil || arg.End == nil {
		return nil, false
	}
	values, err := eval.ConstPkg(arg, b.pkg, stdImport)
	if err != nil {
		return nil, false
	}
	want := map[string]int{"range": 1, "color": 3, "point": 3, "direction": 3, "transform": 16}[kind]
	if len(values) != want {
		return nil, false
	}
	out := make([]float64, len(values))
	for i, v := range values {
		out[i] = math.Round(v*1e6) / 1e6
	}
	return out, true
}

func (b *builder) rangeBounds(p *program.Parameter) (float64, float64, error) {
	lo, err := eval.ConstPkg(p.Range.Min, b.pkg, stdImport)
	if err != nil || len(lo) != 1 {
		return 0, 0, fmt.Errorf("range min is not a constant")
	}
	hi, err := eval.ConstPkg(p.Range.Max, b.pkg, stdImport)
	if err != nil || len(hi) != 1 {
		return 0, 0, fmt.Errorf("range max is not a constant")
	}
	return lo[0], hi[0], nil
}

// uniqueName keeps parameter names distinct across the calls one preset
// makes: a second `cellSize` becomes `cellSize2`.
func (b *builder) uniqueName(name string) string {
	out := name
	for n := 2; b.liftedNames[out] || out == "body"; n++ {
		out = fmt.Sprintf("%s%d", name, n)
	}
	b.liftedNames[out] = true
	return out
}

// resolve finds the function a call names, with the qualified name a document
// writes it by. A builtin, or a call this tool cannot place, resolves to nil.
func (b *builder) resolve(ref *program.Reference, argc int) (*program.Function, string) {
	if ref.Prefix == "" {
		if fn := findFunction(b.pkg, ref.Name, argc); fn != nil {
			return fn, b.pkgName + "." + ref.Name
		}
		return nil, ""
	}
	for _, imp := range b.file.Imports {
		prefix := imp.Prefix
		if prefix == "" {
			prefix = path.Base(imp.Path)
		}
		if prefix != ref.Prefix {
			continue
		}
		pkg := program.GetStdPackage(imp.Path)
		if pkg == nil {
			return nil, ""
		}
		if fn := findFunction(pkg, ref.Name, argc); fn != nil {
			return fn, ref.Prefix + "." + ref.Name
		}
	}
	return nil, ""
}

func findFunction(pkg *program.Package, name string, argc int) *program.Function {
	for _, file := range pkg.Files {
		for _, ft := range file.FnTypes {
			if ft.Name != name {
				continue
			}
			for _, g := range ft.Groups {
				if int(g.ParamCount) == argc && len(g.Functions) > 0 {
					return g.Functions[0]
				}
			}
		}
	}
	return nil
}

// isPackageName reports whether name is a top-level constant or function of
// the preset's own package, which a document has to qualify. A local of the
// preset's body is not.
func (b *builder) isPackageName(name string) bool {
	for _, as := range b.fn.Assignments {
		for _, v := range as.Variables {
			if v.Name == name {
				return false
			}
		}
	}
	for _, file := range b.pkg.Files {
		for _, as := range file.Assignments {
			for _, v := range as.Variables {
				if v.Name == name {
					return true
				}
			}
		}
		for _, ft := range file.FnTypes {
			if ft.Name == name {
				return true
			}
		}
	}
	return false
}

// render is the source between from and to with the edits inside it applied
// and the locals renamed.
func (b *builder) render(from, to int, renames map[string]string) string {
	var inside []edit
	for _, e := range b.edits {
		if e.from >= from && e.to <= to {
			inside = append(inside, e)
		}
	}
	sort.Slice(inside, func(i, j int) bool { return inside[i].from > inside[j].from })
	text := b.src[from:to]
	for _, e := range inside {
		text = text[:e.from-from] + e.text + text[e.to-from:]
	}
	return strings.ReplaceAll(renameLocals(text, renames), packagePlaceholder, b.pkgName)
}

// packagePlaceholder stands for the package a same-package call is qualified
// with until the locals are renamed, since a local can share its name: plated's
// presets bind a local `plated`.
const packagePlaceholder = "${package}"

var placeholderRE = regexp.MustCompile(`\$\{[^}]*\}`)

// renameLocals renames the preset's locals in text, leaving the placeholders
// already written alone.
func renameLocals(text string, renames map[string]string) string {
	if len(renames) == 0 {
		return text
	}
	var sb strings.Builder
	last := 0
	for _, m := range placeholderRE.FindAllStringIndex(text, -1) {
		sb.WriteString(renameWords(text[last:m[0]], renames))
		sb.WriteString(text[m[0]:m[1]])
		last = m[1]
	}
	sb.WriteString(renameWords(text[last:], renames))
	return sb.String()
}

// wordRE takes a member or qualified name with its dot, so `metal.steel` is
// never read as the local `steel`.
var wordRE = regexp.MustCompile(`\.?[A-Za-z_]\w*`)

func renameWords(text string, renames map[string]string) string {
	return wordRE.ReplaceAllStringFunc(text, func(w string) string {
		if strings.HasPrefix(w, ".") {
			return w
		}
		if r, ok := renames[w]; ok {
			return r
		}
		return w
	})
}

func appendUnique(list []string, s string) []string {
	for _, have := range list {
		if have == s {
			return list
		}
	}
	return append(list, s)
}

// stdImport resolves an import path to its embedded package, for the constant
// evaluator.
func stdImport(importPath string) (*program.Package, error) {
	if pkg := program.GetStdPackage(importPath); pkg != nil {
		return pkg, nil
	}
	return nil, fmt.Errorf("package %q is not embedded", importPath)
}
