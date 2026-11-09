package property

import (
	"fmt"
	"strings"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/program"
)

// shrinkToVanishName is builtin's marker for a part (3d.gsl): the part itself
// everywhere but the viewer, which shrinks it to nothing when it is small on
// screen.
const shrinkToVanishName = "shrinkToVanish"

// MarkParts wraps each part of the primary package in shrinkToVanish(part, lo,
// hi, solid), with [lo, hi] the box of its solid, and solid 1, or of its
// surface when its solid is unbounded, as a declared box's is, and solid 0. A part is a top-level value of one name,
// or one name a multi-value call is destructured into, whose every reference is a return or an operand of a min (or a builtin
// unionSharp of two or three) that is itself one, all the way up to a return,
// and whose box is bounded on every axis. So a part is always
// evaluated where it is placed, in the document's own coordinates: a value
// also read anywhere else, a remap above all, is not one, and neither is
// anything reached only through it.
//
// It returns prog itself when there is no part; otherwise a copy, so the
// caller's (possibly cached) program is untouched.
//
// A part with a simpler stand-in, name.simple = proxy, is wrapped in
// simplified(part, proxy, lo, hi) inside that, and its simple property is
// dropped. Any other top-level value's is a colour's or a material's, and is
// written as simpleValue(value, stand), as Expand writes one inside a
// function; one on a shape that is not a part is an error CheckSimple
// reports, once the program is known to be otherwise sound. A stand-in
// carried out of a function, carriedSimple(proxy) (CarryReturnProperties), is
// a part's like any other, and dropped from anything else: the document did
// not write it, and a fastener read through a remap simply has none.
func MarkParts(prog *program.Program) (*program.Program, error) {
	if prog == nil || prog.Primary == nil {
		return prog, nil
	}
	parts := findParts(prog)
	if len(parts) == 0 && !topLevelSimple(prog.Primary) {
		return prog, nil
	}

	cloned := proto.Clone(prog).(*program.Program)
	taken := declaredNames(cloned.Primary)
	for _, f := range cloned.Primary.Files {
		if f == nil {
			continue
		}
		var out []*program.Assignment
		for _, a := range f.Assignments {
			out = append(out, a)
			if a == nil {
				continue
			}
			if len(a.Variables) == 1 && a.Variables[0] != nil {
				if part, ok := parts[a.Variables[0].Name]; ok {
					a.Expression = shrinkToVanish(simplified(a.Expression, a.Variables[0], part), part)
				} else {
					dropCarriedSimple(a.Variables[0])
				}
				continue
			}
			// A destructured part is renamed where it is destructured, and
			// declared again under its own name, wrapped, with its properties:
			// g1, g2 = gearPair(…) becomes gsl_part_g1, g2 = gearPair(…) and
			// g1 = shrinkToVanish(gsl_part_g1, …).
			for _, v := range a.Variables {
				if v == nil {
					continue
				}
				part, ok := parts[v.Name]
				if !ok {
					dropCarriedSimple(v)
					continue
				}
				hidden := freshName("gsl_part_"+v.Name, taken)
				named := &program.Variable{Name: v.Name, Properties: v.Properties}
				out = append(out, &program.Assignment{
					Start:      a.Start,
					End:        a.End,
					Variables:  []*program.Variable{named},
					Expression: shrinkToVanish(simplified(ref(hidden), named, part), part),
				})
				v.Name, v.Properties = hidden, nil
			}
		}
		if err := lowerSimpleValues(out, standNests(cloned, f, out)); err != nil {
			return nil, err
		}
		out, err := placeStandIns(out)
		if err != nil {
			return nil, err
		}
		f.Assignments = out
	}
	return cloned, nil
}

// placeStandIns moves each value with a stand-in, a part or a colour or
// material, to just after the last declaration its stand-in reads, when that
// comes later: a document reads its names in order, and bolt.simple =
// boltProxy is usually written after both. A value read in between cannot
// move, and is an error.
func placeStandIns(assigns []*program.Assignment) ([]*program.Assignment, error) {
	for i := 0; i < len(assigns); i++ {
		a := assigns[i]
		if a == nil || len(a.Variables) != 1 || a.Variables[0] == nil {
			continue
		}
		stand := standOf(a.Expression)
		if stand == nil {
			continue
		}
		reads := map[string]int{}
		countRefs(stand, reads)
		last := i
		for j := i + 1; j < len(assigns); j++ {
			if assigns[j] == nil {
				continue
			}
			for _, v := range assigns[j].Variables {
				if v != nil && reads[v.Name] > 0 {
					last = j
				}
			}
		}
		if last == i {
			continue
		}
		name := a.Variables[0].Name
		for j := i + 1; j <= last; j++ {
			between := map[string]int{}
			countAssignRefs(assigns[j:j+1], between)
			if between[name] > 0 {
				return nil, fmt.Errorf("%s.simple: %s is read before its stand-in is declared; declare the stand-in before %s is first used", name, name, name)
			}
		}
		moved := append([]*program.Assignment{}, assigns[:i]...)
		moved = append(moved, assigns[i+1:last+1]...)
		moved = append(moved, a)
		assigns = append(moved, assigns[last+1:]...)
		i-- // the assignment now at i has not been looked at
	}
	return assigns, nil
}

// shrinkToVanish wraps a part's expression in builtin's marker.
func shrinkToVanish(expr *program.Operand, part partBox) *program.Operand {
	solid := 0.0
	if part.solid {
		solid = 1
	}
	return call(shrinkToVanishName, expr,
		nary(program.Operation_COLLECT, part.box.lo[:]...),
		nary(program.Operation_COLLECT, part.box.hi[:]...), num(solid))
}

// simplePropName is the property that gives a part its simpler stand-in, and
// simplifiedName builtin's marker for it (3d.gsl).
const (
	simplePropName = "simple"
	simplifiedName = "simplified"
)

// simplifiedCall is the simplified(part, proxy, lo, hi) inside a part's
// shrinkToVanish, or nil when the value is not a part with a stand-in.
func simplifiedCall(op *program.Operand) *program.FnCall {
	outer := op.GetFnCall()
	if !namedCall(outer, shrinkToVanishName, 4) {
		return nil
	}
	inner := outer.Arguments[0].GetFnCall()
	if !namedCall(inner, simplifiedName, 4) {
		return nil
	}
	return inner
}

// simplified wraps a part's expression in builtin's simplified marker when its
// variable has a simple property, which it takes off the variable, and returns
// the expression as it is otherwise.
func simplified(expr *program.Operand, v *program.Variable, part partBox) *program.Operand {
	var proxy *program.Operand
	kept := v.Properties[:0]
	for _, prop := range v.Properties {
		if prop != nil && prop.Name == simplePropName {
			proxy = prop.Expression
			continue
		}
		kept = append(kept, prop)
	}
	v.Properties = kept
	if proxy == nil {
		return expr
	}
	if fc := proxy.GetFnCall(); namedCall(fc, carriedSimpleName, 1) {
		proxy = fc.Arguments[0]
	}
	return call(simplifiedName, expr, proxy,
		nary(program.Operation_COLLECT, part.box.lo[:]...),
		nary(program.Operation_COLLECT, part.box.hi[:]...))
}

// dropCarriedSimple takes a stand-in carried out of a function off v, which
// is not a part.
func dropCarriedSimple(v *program.Variable) {
	kept := v.Properties[:0]
	for _, prop := range v.Properties {
		if prop != nil && prop.Name == simplePropName && namedCall(prop.Expression.GetFnCall(), carriedSimpleName, 1) {
			continue
		}
		kept = append(kept, prop)
	}
	v.Properties = kept
}

// CheckSimple reports a top-level stand-in MarkParts wrote as a value's,
// simpleValue(value, stand), whose value is a distance: a shape that is not a
// part, whose stand-in would have no box to be chosen by. It runs after type
// analysis, which tells a distance from a colour, so an error in the program
// itself is reported first, and is not mistaken for this.
//
// It reports one Expand wrote inside a function of the document too: a shape's
// stand-in that CarryReturnProperties could not carry out, on a local the
// function does not return, or one that sets nothing only a shape takes (see A
// PART'S STAND-IN in carry.go). One in a library's function is taken off
// instead, leaving the shape: a library is used as it is written, with or
// without the boxes that let its stand-ins out, and a part with no stand-in is
// only drawn in full.
func CheckSimple(prog *program.Program) error {
	if prog == nil || prog.Primary == nil {
		return nil
	}
	for _, f := range prog.Primary.Files {
		if f == nil {
			continue
		}
		for _, a := range f.Assignments {
			if name := shapeStandIn(a, false); name != "" {
				return fmt.Errorf("%s.simple: %s is not a part; only a named value that only a min or a return reads, and whose box is bounded, can have a simpler stand-in", name, name)
			}
		}
	}
	for _, f := range prog.Primary.Files {
		for _, ft := range f.GetFnTypes() {
			if name := fnTypeShapeStandIn(ft, false); name != "" {
				return fmt.Errorf("%s.simple: a shape's stand-in leaves a function only on the local it returns, and only when that local also declares a box, colour or material", name)
			}
		}
	}
	for _, dep := range prog.Dependencies {
		for _, f := range dep.GetFiles() {
			for _, ft := range f.GetFnTypes() {
				fnTypeShapeStandIn(ft, true)
			}
		}
	}
	return nil
}

// shapeStandIn is the name a declaration gives a distance with a value's
// stand-in, simpleValue(value, stand), or "". With drop, the declaration is
// given the distance alone.
func shapeStandIn(a *program.Assignment, drop bool) string {
	if a == nil || len(a.Variables) != 1 || a.Variables[0] == nil {
		return ""
	}
	fc := a.Expression.GetFnCall()
	if !namedCall(fc, simpleValueName, 2) && !namedCall(fc, simpleTextureName, 2) {
		return ""
	}
	if t := fc.Arguments[0].GetType(); t != nil && t.Size == 1 && t.ElemType == nil {
		if drop {
			a.Expression = fc.Arguments[0]
		}
		return a.Variables[0].Name
	}
	return ""
}

// fnTypeShapeStandIn is shapeStandIn over every function of ft, nested ones
// too: the first name it finds, or "", and with drop every one of them.
func fnTypeShapeStandIn(ft *program.FnType, drop bool) string {
	first := ""
	for _, g := range ft.GetGroups() {
		for _, fn := range g.GetFunctions() {
			for _, a := range fn.GetAssignments() {
				if name := shapeStandIn(a, drop); name != "" && first == "" {
					first = name
				}
			}
			for _, nested := range fn.GetFnTypes() {
				if name := fnTypeShapeStandIn(nested, drop); name != "" && first == "" {
					first = name
				}
			}
		}
	}
	return first
}

// topLevelSimple reports whether a top-level declaration of the package sets
// name.simple.
func topLevelSimple(pkg *program.Package) bool {
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, a := range f.Assignments {
			if a == nil {
				continue
			}
			for _, v := range a.Variables {
				if v != nil && hasProp(v, simplePropName) {
					return true
				}
			}
		}
	}
	return false
}

// standOf is the stand-in a declaration's expression carries: a part's
// simplified or a value's simpleValue or simpleTexture, or nil.
func standOf(op *program.Operand) *program.Operand {
	if fc := simplifiedCall(op); fc != nil {
		return fc.Arguments[1]
	}
	if fc := op.GetFnCall(); namedCall(fc, simpleValueName, 2) || namedCall(fc, simpleTextureName, 2) {
		return fc.Arguments[1]
	}
	return nil
}

// standNests reports whether a stand-in declared in the document has a
// stand-in of its own: a value of the document whose declaration is one or
// such a call, or a call to a function of the document or of an import whose
// single return is one. Expand has lowered every function's stand-ins by now, so a
// library texture that declares one returns a value declared simpleValue(…).
func standNests(prog *program.Program, f *program.PackageFile, assigns []*program.Assignment) func(*program.Operand) bool {
	isStandIn := func(op *program.Operand, decls map[string]*program.Operand) bool {
		if r := op.GetReference(); r != nil && r.Prefix == "" && decls != nil {
			op = decls[r.Name]
		}
		fc := op.GetFnCall()
		return namedCall(fc, simpleValueName, 2) || namedCall(fc, simpleTextureName, 2)
	}
	returnsStandIn := func(fn *program.Function) bool {
		if fn == nil || len(fn.Returns) != 1 {
			return false
		}
		locals := map[string]*program.Operand{}
		for _, a := range fn.Assignments {
			if a != nil && len(a.Variables) == 1 && a.Variables[0] != nil {
				locals[a.Variables[0].Name] = a.Expression
			}
		}
		return isStandIn(fn.Returns[0], locals)
	}
	functions := func(pkg *program.Package, name string, args int) []*program.Function {
		var out []*program.Function
		for _, pf := range pkg.GetFiles() {
			for _, ft := range pf.GetFnTypes() {
				if ft == nil || ft.Name != name {
					continue
				}
				for _, g := range ft.Groups {
					for _, fn := range g.GetFunctions() {
						if fn != nil && len(fn.Parameters) == args {
							out = append(out, fn)
						}
					}
				}
			}
		}
		return out
	}
	imported := func(prefix string) *program.Package {
		for _, imp := range f.GetImports() {
			alias := imp.Prefix
			if alias == "" {
				alias = imp.Path[strings.LastIndex(imp.Path, "/")+1:]
			}
			if alias != prefix {
				continue
			}
			for _, dep := range prog.Dependencies {
				if dep != nil && dep.Path == imp.Path {
					return dep
				}
			}
		}
		return nil
	}
	var nests func(stand *program.Operand, seen map[string]bool) bool
	nests = func(stand *program.Operand, seen map[string]bool) bool {
		// A value of the document nests when its declaration is a stand-in or
		// a call that returns one, as dots = playful.polkaDot(…) does.
		if r := stand.GetReference(); r != nil && r.Prefix == "" {
			if seen[r.Name] {
				return false
			}
			seen[r.Name] = true
			for _, a := range assigns {
				if a != nil && len(a.Variables) == 1 && a.Variables[0] != nil && a.Variables[0].Name == r.Name {
					return isStandIn(a.Expression, nil) || hasProp(a.Variables[0], simplePropName) || nests(a.Expression, seen)
				}
			}
			return false
		}
		fc := stand.GetFnCall()
		if fc == nil || fc.Reference == nil {
			return false
		}
		pkg := prog.Primary
		if fc.Reference.Prefix != "" {
			if pkg = imported(fc.Reference.Prefix); pkg == nil {
				return false
			}
		}
		for _, fn := range functions(pkg, fc.Reference.Name, len(fc.Arguments)) {
			if returnsStandIn(fn) {
				return true
			}
		}
		return false
	}
	return func(stand *program.Operand) bool {
		return nests(stand, map[string]bool{})
	}
}

// namedCall reports whether fc calls the unprefixed name with args arguments.
func namedCall(fc *program.FnCall, name string, args int) bool {
	return fc != nil && fc.Reference != nil && fc.Reference.Prefix == "" &&
		fc.Reference.Name == name && len(fc.Arguments) == args
}

func hasProp(v *program.Variable, name string) bool {
	for _, prop := range v.Properties {
		if prop != nil && prop.Name == name {
			return true
		}
	}
	return false
}

// declaredNames is every name the package declares at the top level.
func declaredNames(pkg *program.Package) map[string]bool {
	names := map[string]bool{}
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, a := range f.Assignments {
			if a == nil {
				continue
			}
			for _, v := range a.Variables {
				if v != nil {
					names[v.Name] = true
				}
			}
		}
		for _, ft := range f.FnTypes {
			if ft != nil {
				names[ft.Name] = true
			}
		}
	}
	return names
}

// freshName is base, or base with a number after it, whichever taken does not
// hold yet, and takes it.
func freshName(base string, taken map[string]bool) string {
	name := base
	for i := 2; taken[name]; i++ {
		name = fmt.Sprintf("%s_%d", base, i)
	}
	taken[name] = true
	return name
}

// partBox is a part's box, and whether its solid is inside it as well as its
// surface.
type partBox struct {
	box   *bbox
	solid bool
}

// findParts returns the box of each part's solid, or of its surface where
// only that is bounded, by name.
func findParts(prog *program.Program) map[string]partBox {
	pkg := prog.Primary
	values := map[string]*program.Operand{}
	// A destructured name (g1 in g1, g2 = gearPair(…)) has no expression of
	// its own to walk into; it is a part or not like any other value, by where
	// it is read.
	destructured := map[string]bool{}
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, a := range f.Assignments {
			if a == nil {
				continue
			}
			if len(a.Variables) == 1 && a.Variables[0] != nil {
				values[a.Variables[0].Name] = a.Expression
				continue
			}
			for _, v := range a.Variables {
				if v != nil {
					destructured[v.Name] = true
				}
			}
		}
	}
	if len(values) == 0 && len(destructured) == 0 {
		return nil
	}

	refs := map[string]int{}
	countPackageRefs(pkg, refs)
	own := ownFunctionNames(pkg)

	// Walk down from the returns through min, expanding only values every
	// reference to which the walk reaches, until that set settles. It only
	// shrinks: expanding fewer values reaches fewer references.
	var clean map[string]bool
	for {
		reached := map[string]int{}
		expanded := map[string]bool{}
		var walk func(op *program.Operand)
		walk = func(op *program.Operand) {
			if op == nil {
				return
			}
			switch t := op.T.(type) {
			case *program.Operand_Reference:
				r := t.Reference
				if r != nil && r.Prefix == "" && destructured[r.Name] {
					reached[r.Name]++
					return
				}
				if r == nil || r.Prefix != "" || values[r.Name] == nil {
					return
				}
				reached[r.Name]++
				if (clean == nil || clean[r.Name]) && !expanded[r.Name] {
					expanded[r.Name] = true
					walk(values[r.Name])
				}
			case *program.Operand_Expression:
				if t.Expression.Operation == program.Operation_MIN {
					for _, o := range t.Expression.Operands {
						walk(o)
					}
				}
			case *program.Operand_FnCall:
				fc := t.FnCall
				if fc.Reference != nil && fc.Reference.Prefix == "" && fc.Reference.Name == "unionSharp" &&
					!own["unionSharp"] && (len(fc.Arguments) == 2 || len(fc.Arguments) == 3) {
					for _, a := range fc.Arguments {
						walk(a)
					}
				}
			}
		}
		for _, f := range pkg.Files {
			if f == nil {
				continue
			}
			for _, ret := range f.Returns {
				walk(ret)
			}
		}
		next := map[string]bool{}
		for name, n := range reached {
			if n == refs[name] {
				next[name] = true
			}
		}
		if clean != nil && len(next) == len(clean) {
			break
		}
		clean = next
	}
	if len(clean) == 0 {
		return nil
	}

	// Read every part's box in one evaluation, as the elements of a list, so
	// what they share is evaluated once; one at a time only if that fails.
	names := make([]string, 0, len(clean))
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, a := range f.Assignments {
			if a == nil {
				continue
			}
			for _, v := range a.Variables {
				if v != nil && clean[v.Name] {
					names = append(names, v.Name)
				}
			}
		}
	}
	b := partsBoxProp(prog)
	vals := make([]boxVal, len(names))
	ok := make([]bool, len(names))
	list := make([]*program.Operand, len(names))
	for i, name := range names {
		list[i] = ref(name)
	}
	if v, err := b.engine().Eval(nary(program.Operation_COLLECT, list...)); err == nil && len(v.(boxVal).elems) == len(names) {
		for i, e := range v.(boxVal).elems {
			vals[i], ok[i] = e, true
		}
	} else {
		for i, name := range names {
			if v, err := b.engine().Eval(ref(name)); err == nil {
				vals[i], ok[i] = v.(boxVal), true
			}
		}
	}
	out := map[string]partBox{}
	known := documentNames(prog)
	for i, name := range names {
		if !ok[i] || vals[i].op != nil {
			continue
		}
		if bx := vals[i].box; boundedBox(bx) && boxResolves(bx, known) {
			out[name] = partBox{box: bx, solid: true}
		} else if bx := vals[i].surface(); boundedBox(bx) && boxResolves(bx, known) {
			out[name] = partBox{box: bx}
		}
	}
	return out
}

// documentNames is every name the primary package can read unprefixed: its
// own declarations, and those of builtin and its dot imports.
func documentNames(prog *program.Program) map[string]bool {
	names := declaredNames(prog.Primary)
	add := func(pkg *program.Package) {
		if pkg == nil {
			return
		}
		for n := range declaredNames(pkg) {
			names[n] = true
		}
	}
	for _, dep := range prog.Dependencies {
		if dep != nil && (dep.Path == "builtin" || dep.Path == "") {
			add(dep)
		}
	}
	for _, f := range prog.Primary.Files {
		if f == nil {
			continue
		}
		for _, imp := range f.Imports {
			if imp.Prefix != "." {
				continue
			}
			for _, dep := range prog.Dependencies {
				if dep != nil && dep.Path == imp.Path {
					add(dep)
				}
			}
		}
	}
	return names
}

// assetBuiltins read a mesh asset, or the topology of one, named relative to
// the package the call is written in.
var assetBuiltins = map[string]bool{
	"mesh": true, "importStep": true, "importMesh": true, "importWarp": true, "bvhSdf": true, "vdbSdf": true,
	"facePoint": true, "edgePoint": true, "vertexPoint": true, "circleCenter": true,
}

// boxResolves reports whether every name a box's corners read is one the
// document resolves (documentNames, or a native builtin), and none of them
// reads a mesh asset. A box can be built from another package's names, a STEP
// face in an imported branch say; written into the document as the part's
// mark, it would read names the document does not have, or an asset the
// document's package does not hold, so such a value is left unmarked.
func boxResolves(bx *bbox, known map[string]bool) bool {
	var ok func(op *program.Operand) bool
	ok = func(op *program.Operand) bool {
		if op == nil {
			return true
		}
		switch t := op.T.(type) {
		case *program.Operand_Reference:
			r := t.Reference
			return r != nil && r.Prefix == "" && (known[r.Name] || program.IsBuiltin(r.Name))
		case *program.Operand_PropRef:
			return false
		case *program.Operand_Expression:
			operands := t.Expression.Operands
			if t.Expression.Operation == program.Operation_ACCESS && len(operands) > 0 {
				// The selectors are swizzle names (x, xy), not values.
				operands = operands[:1]
			}
			for _, o := range operands {
				if !ok(o) {
					return false
				}
			}
		case *program.Operand_FnCall:
			fc := t.FnCall
			if fc.Reference == nil || fc.Reference.Prefix != "" || assetBuiltins[fc.Reference.Name] ||
				!(known[fc.Reference.Name] || program.IsBuiltin(fc.Reference.Name)) {
				return false
			}
			for _, a := range fc.Arguments {
				if !ok(a) {
					return false
				}
			}
		}
		return true
	}
	for i := 0; i < 3; i++ {
		if !ok(bx.lo[i]) || !ok(bx.hi[i]) {
			return false
		}
	}
	return true
}

// boundedBox reports whether a box holds something and is bounded on every
// axis.
func boundedBox(bx *bbox) bool {
	if bx == nil || bx.empty {
		return false
	}
	for i := 0; i < 3; i++ {
		if bx.lo[i] == nil || bx.hi[i] == nil {
			return false
		}
	}
	return true
}

// partsBoxProp is the box property over the primary package's scope, with its
// dependencies, builtin and its dot imports merged in, as Expand builds it.
func partsBoxProp(prog *program.Program) *boxProp {
	pkg := prog.Primary
	getImport := func(path string) (*program.Package, error) {
		for _, dep := range prog.Dependencies {
			if dep != nil && (dep.Path == path || (path == "builtin" && (dep.Path == "builtin" || dep.Path == ""))) {
				return dep, nil
			}
		}
		return prog.GetImport(path)
	}
	decls := collectPackageAssignments(pkg)
	multiDecls := collectPackageMultiAssignments(pkg)
	funcsList := collectPackageFunctionsList(pkg)
	funcsMap := collectPackageFunctions(pkg)
	for _, dep := range prog.Dependencies {
		if dep != nil {
			mergePackageNamespace(dep, decls, funcsList, funcsMap)
		}
	}
	if ip, err := getImport("builtin"); err == nil && ip != nil {
		mergePackageNamespace(ip, decls, funcsList, funcsMap)
	}
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, imp := range f.Imports {
			if imp.Prefix == "." {
				if ip, err := getImport(imp.Path); err == nil && ip != nil {
					mergePackageNamespace(ip, decls, funcsList, funcsMap)
				}
			}
		}
	}
	return &boxProp{decls: decls, multiDecls: multiDecls, funcs: funcsList,
		imports: collectImportNamespaces(pkg, getImport), own: ownFunctionNames(pkg), sliders: sliderNames(pkg)}
}

// countPackageRefs counts every unprefixed reference to a name anywhere in the
// package: declarations, their properties, returns and function bodies. A
// function's own local of the same name counts too, which only ever leaves a
// value out.
func countPackageRefs(pkg *program.Package, refs map[string]int) {
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		countAssignRefs(f.Assignments, refs)
		for _, ret := range f.Returns {
			countRefs(ret, refs)
		}
		for _, ft := range f.FnTypes {
			countFnTypeRefs(ft, refs)
		}
	}
}

func countFnTypeRefs(ft *program.FnType, refs map[string]int) {
	if ft == nil {
		return
	}
	for _, group := range ft.Groups {
		if group == nil {
			continue
		}
		for _, fn := range group.Functions {
			if fn == nil {
				continue
			}
			countAssignRefs(fn.Assignments, refs)
			for _, ret := range fn.Returns {
				countRefs(ret, refs)
			}
			for _, nested := range fn.FnTypes {
				countFnTypeRefs(nested, refs)
			}
		}
	}
}

func countAssignRefs(assigns []*program.Assignment, refs map[string]int) {
	for _, a := range assigns {
		if a == nil {
			continue
		}
		countRefs(a.Expression, refs)
		for _, v := range a.Variables {
			if v == nil {
				continue
			}
			for _, prop := range v.Properties {
				if prop != nil {
					countRefs(prop.Expression, refs)
				}
			}
		}
	}
}

func countRefs(op *program.Operand, refs map[string]int) {
	if op == nil {
		return
	}
	switch t := op.T.(type) {
	case *program.Operand_Reference:
		if t.Reference != nil && t.Reference.Prefix == "" {
			refs[t.Reference.Name]++
		}
	case *program.Operand_PropRef:
		if t.PropRef != nil && t.PropRef.Reference != nil && t.PropRef.Reference.Prefix == "" {
			refs[t.PropRef.Reference.Name]++
		}
	case *program.Operand_Expression:
		for _, o := range t.Expression.Operands {
			countRefs(o, refs)
		}
	case *program.Operand_FnCall:
		if t.FnCall.Reference != nil && t.FnCall.Reference.Prefix == "" {
			refs[t.FnCall.Reference.Name]++
		}
		for _, a := range t.FnCall.Arguments {
			countRefs(a, refs)
		}
	}
}
