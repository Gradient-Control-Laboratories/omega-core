package property

import (
	"fmt"
	"sort"
	"strings"

	"github.com/Gradient-Control-Laboratories/omega-core/program"
)

// LAYERS: DISPLACEMENT WHERE THE MATERIAL IS SET.
//
// A material's height moves a surface for display (material.withDisplacement).
// A renderer used to apply it to each return as a whole, d - h, after every
// boolean in the return: a coat set on a blank that is then cut displaced the
// cut face too, and a coat modelled as a thin shell thickened or pinched shut
// instead of riding the surface beneath it.
//
// Here each named shape that sets a material with a height moves its own
// surface, before the booleans that use it, by the height its material adds
// over the material it would otherwise have (its inherited material, the blend
// of its children's):
//
//	inc(v)   = height(v.material) - height(inherited(v))
//	under(v) = v's body, with every shape in it moved the same way
//	disp(v)  = under(v) - displacementScale() * inc(v)
//
// For paint laid with material.levelled over the material beneath, inc is the
// film's build less the relief it levels, so the paint rides the wood and
// stacks coat over coat; a cut taken after it stays a plane, its cutter never
// having moved; and paint laid over the cut raises the cut face and the outer
// faces alike. A shape that sets a material with no height stops the chain:
// its children are drawn undisplaced, as a return with such a material always
// was.
//
// Shading follows the same layers. Where a shape's layer has thickness (inc >
// 0) and a point lies inside the solid beneath it (under <= 0), the point shows
// the inherited material and colour, not the shape's own: a cut through four
// coats shows four bands, and the wood beneath them.
//
// A return that is one such shape with nothing displaced inside it is left to
// the renderer as before (its DisplacedReturns entry is nil), so a document
// that sets its materials on its returns compiles exactly as it did.

const (
	// The synthetic declarations: a shape's displaced distance, the distance
	// of what is beneath its layer, and its layer's height.
	layerDispPrefix     = "gsl__disp__"
	layerUnderPrefix    = "gsl__under__"
	layerIncPrefix      = "gsl__inc__"
	layerSurfPrefix     = "gsl__surf__"
	layerSurfGradPrefix = "gsl__surfgrad__"
	layerTopPrefix      = "gsl__top__"
	layerSurfReadPrefix = "gsl__surfread__"

	// layerThick is the least layer thickness (mm) the shading switch takes
	// as a layer; layerInside is how far above the surface beneath (mm) a
	// point still reads as on it, the march's hit tolerance with room to
	// spare.
	layerThick  = 2e-3
	layerInside = 1e-3
	// layerUndrawn is how far inside the undisplaced surface (mm) a point
	// must be, at least, to read as beneath a coat where nothing is displaced
	// (beneath).
	layerUndrawn = 0.01

	// displacementFn is the library function every height comes through.
	displacementFn = "withDisplacement"
)

// layers is the layer analysis of a package: which shapes move, the
// declarations their displaced distances need, and the switch each moving
// shape's colour and material take.
type layers struct {
	pkg   *program.Package
	decls map[string]*program.Operand // the package's single-name declarations, as written
	mats  map[string]*program.Operand // the authored .material expressions
	plain *resolver                   // today's material resolution, for heights

	// scopes is the primary package and every package it imports, by path,
	// for mayDisplace (layers_height.go).
	scopes map[string]*scope

	displaceMemo map[string]bool // mayDisplace of a function or declaration
	containsMemo map[string]bool // contains

	// displacing is every shape that sets a material with a height.
	displacing map[string]bool
	// emitted names each shape whose displaced declarations have been made,
	// and synth holds them in order.
	emitted map[string]bool
	synth   []*program.Assignment
	// inTree is every displacing shape inside a return displaced in its
	// booleans: the shapes whose colour and material switch by layer.
	inTree map[string]bool
	// returns is each geometry return's displaced distance, nil where the
	// renderer moves the return as a whole.
	returns []*program.Operand

	// deriv differentiates a chain's bottom shape for the closest point its
	// coats read their heights at, and surfaces records, by chain, whether
	// that point's declaration was made (surface).
	deriv    *derivative
	surfaces map[string]bool
	// reads are the values read at surface points, packed by packReads, and
	// surfAt the index in synth just after each point's declaration.
	reads  []surfRead
	surfAt map[string]int
}

// analyzeLayers builds the layer analysis of prog's primary package: the
// displaced distance of each return and the declarations it reads. It
// returns nil when no shape sets a material with a height inside a return.
func analyzeLayers(prog *program.Program) *layers {
	pkg := prog.Primary
	if pkg == nil || len(pkg.Files) == 0 || !packageHasProperty(pkg, materialProperty) {
		return nil
	}
	l := &layers{
		pkg:          pkg,
		decls:        collectPackageAssignments(pkg),
		mats:         collectPackageProperties(pkg, materialProperty),
		plain:        newResolver(pkg, prog.GetImport, materialSpec),
		displaceMemo: map[string]bool{},
		containsMemo: map[string]bool{},
		displacing:   map[string]bool{},
		emitted:      map[string]bool{},
		inTree:       map[string]bool{},
	}
	l.indexScopes(prog.GetImport)
	exp := &expansion{getImport: prog.GetImport, deps: prog.Dependencies, scopes: map[*program.Package]*pkgScope{}}
	s := exp.scopeFor(pkg)
	l.deriv = &derivative{
		decls:      s.decls,
		multiDecls: s.multiDecls,
		funcs:      s.funcsList,
		imports:    s.imports,
		own:        ownFunctionNames(pkg),
	}

	names := make([]string, 0, len(l.mats))
	for name := range l.mats {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if _, single := l.decls[name]; !single {
			continue // a destructured name has no body to move
		}
		if l.mayDisplace(l.mats[name], primaryScope, map[string]bool{}) {
			l.displacing[name] = true
		}
	}
	if len(l.displacing) == 0 {
		return nil
	}

	last := pkg.Files[len(pkg.Files)-1]
	l.returns = make([]*program.Operand, len(last.Returns))
	any := false
	for i, ret := range last.Returns {
		if !l.containsOp(ret) || l.wholeReturn(ret) {
			continue
		}
		l.returns[i] = l.subst(ret)
		any = true
	}
	if !any {
		return nil
	}
	return l
}

// wholeReturn reports whether ret is one displacing shape with nothing moved
// inside it, placed by nothing but moves (remap, a transform, a part's
// stand-in): moving that shape before the moves is moving the return after
// them, which the renderer does as a whole exactly as before.
func (l *layers) wholeReturn(ret *program.Operand) bool {
	op := l.plain.unwrapRemaps(ret)
	for depth := 0; op != nil && depth < 64; depth++ {
		if stand := simplifiedCall(op); stand != nil {
			op = stand.Arguments[0]
			continue
		}
		switch t := op.T.(type) {
		case *program.Operand_Reference:
			name := t.Reference.Name
			if t.Reference.Prefix != "" {
				return false
			}
			if l.displacing[name] {
				return !l.containsOp(l.decls[name])
			}
			body, ok := l.plain.decls[name]
			if !ok {
				return false
			}
			op = body
		case *program.Operand_Expression:
			if t.Expression.Operation != program.Operation_REMAP || len(t.Expression.Operands) == 0 {
				return false
			}
			op = t.Expression.Operands[0]
		default:
			return false
		}
	}
	return false
}

// contains reports whether a declaration has a moving surface in it: it sets
// a material with a height, or sets none and refers to a shape that has one.
// A shape whose material has no height stops the chain.
func (l *layers) contains(name string) bool {
	if v, ok := l.containsMemo[name]; ok {
		return v
	}
	l.containsMemo[name] = false // tentative (breaks cycles)
	body, ok := l.decls[name]
	if !ok {
		return false
	}
	var v bool
	if _, authored := l.mats[name]; authored {
		v = l.displacing[name]
	} else {
		v = l.containsOp(body)
	}
	l.containsMemo[name] = v
	return v
}

// containsOp reports whether an expression refers to a declaration with a
// moving surface in it. A part's stand-in (simplified's second argument) is
// another drawing of the same part, moved with it as a whole, so only the
// part is looked in.
func (l *layers) containsOp(op *program.Operand) bool {
	refs := map[string]bool{}
	collectPartRefs(op, refs)
	for name := range refs {
		if l.contains(name) {
			return true
		}
	}
	return false
}

// collectPartRefs is collectRefs leaving out each part's stand-in.
func collectPartRefs(op *program.Operand, out map[string]bool) {
	if op == nil {
		return
	}
	if stand := simplifiedCall(op); stand != nil {
		collectPartRefs(stand.Arguments[0], out)
		return
	}
	switch t := op.T.(type) {
	case *program.Operand_Reference:
		if t.Reference.Prefix == "" {
			out[t.Reference.Name] = true
		}
	case *program.Operand_Expression:
		for _, o := range t.Expression.Operands {
			collectPartRefs(o, out)
		}
	case *program.Operand_FnCall:
		for _, a := range t.FnCall.Arguments {
			collectPartRefs(a, out)
		}
	}
}

// subst is op with every reference to a declaration that has a moving
// surface in it replaced by that declaration's displaced distance.
func (l *layers) subst(op *program.Operand) *program.Operand {
	if op == nil {
		return nil
	}
	switch t := op.T.(type) {
	case *program.Operand_Reference:
		if t.Reference.Prefix == "" && l.contains(t.Reference.Name) {
			l.emit(t.Reference.Name)
			return withSpan(ref(layerDispPrefix+t.Reference.Name), op)
		}
		return program.CloneOperand(op)
	case *program.Operand_Expression:
		e := t.Expression
		ops := make([]*program.Operand, len(e.Operands))
		for i, o := range e.Operands {
			if e.Operation == program.Operation_ACCESS && i > 0 {
				ops[i] = program.CloneOperand(o) // a selector, not a value
				continue
			}
			ops[i] = l.subst(o)
		}
		return withSpan(nary(e.Operation, ops...), op)
	case *program.Operand_FnCall:
		fc := t.FnCall
		stand := simplifiedCall(op) != nil
		args := make([]*program.Operand, len(fc.Arguments))
		for i, a := range fc.Arguments {
			if stand && i > 0 {
				// The stand-in and its rung are drawn as they are.
				args[i] = program.CloneOperand(a)
				continue
			}
			args[i] = l.subst(a)
		}
		out := &program.Operand{T: &program.Operand_FnCall{FnCall: &program.FnCall{
			Reference: &program.Reference{Prefix: fc.Reference.Prefix, Name: fc.Reference.Name},
			Arguments: args,
		}}}
		return withSpan(out, op)
	}
	return program.CloneOperand(op)
}

// withSpan gives out op's source span, so an error in a displaced distance
// points at the expression it came from.
func withSpan(out, op *program.Operand) *program.Operand {
	out.Start, out.End, out.Type = op.Start, op.End, op.Type
	return out
}

// emit makes a declaration's displaced distance, and for a displacing shape
// the distance beneath its layer and its layer's height, once.
func (l *layers) emit(name string) {
	if l.emitted[name] {
		return
	}
	l.emitted[name] = true
	body := l.decls[name]
	if !l.displacing[name] {
		l.declare(layerDispPrefix+name, l.subst(body))
		return
	}
	l.inTree[name] = true
	l.declare(layerUnderPrefix+name, l.subst(body))
	// A chain of coats each laid on the name beneath it, down to one
	// undisplaced shape, telescopes: the increments sum to the top coat's
	// height, so its surface is that shape's distance less that one height,
	// read at the shape's closest point. The march then evaluates one
	// material at one point, not every coat's twice; each coat's height is
	// declared once (layerTopPrefix), and its increment, read only by the
	// colour and material switch, is its height less the one beneath.
	if root, _, aliases := l.chainRoot(name); aliases {
		if at := l.surface(name); at != nil {
			l.declare(layerTopPrefix+name, l.readAt(at, height(program.CloneOperand(l.mats[name]))))
			inc := ref(layerTopPrefix + name)
			if beneath := aliasOf(body); beneath != "" && l.displacing[beneath] {
				inc = sub(inc, ref(layerTopPrefix+beneath))
			}
			l.declare(layerIncPrefix+name, inc)
			l.declare(layerDispPrefix+name, sub(program.CloneOperand(root),
				mul(call(program.DisplacementScale), ref(layerTopPrefix+name))))
			return
		}
	}
	l.declare(layerIncPrefix+name, l.increment(name))
	l.declare(layerDispPrefix+name, sub(ref(layerUnderPrefix+name),
		mul(call(program.DisplacementScale), ref(layerIncPrefix+name))))
}

// aliasOf is the name an expression is, or "" when it is not a bare name.
func aliasOf(op *program.Operand) string {
	if r, ok := op.T.(*program.Operand_Reference); ok && r.Reference.Prefix == "" {
		return r.Reference.Name
	}
	return ""
}

// increment is the height a displacing shape's material adds over the one it
// inherits.
//
// Both heights are read at the closest point of the surface the layers stand
// on (surface), not at p: a coat stands off that surface by its build, and a
// relief that varies through the volume — wood's pores — would otherwise
// drift with the standoff, so each coat would level a different pattern from
// the one beneath it.
//
// Only a coat laid over a named shape stands off a surface; a finish set on a
// shape's own expression (a face of a max) sits on it, so is read at p.
func (l *layers) increment(name string) *program.Operand {
	var at *program.Operand
	if aliasOf(l.decls[name]) != "" {
		at = l.surface(name)
	}
	onSurface := func(h *program.Operand) *program.Operand {
		if at == nil {
			return h
		}
		return l.readAt(at, h)
	}
	own := height(program.CloneOperand(l.mats[name]))
	inherited, err := l.plain.fieldOf(l.plain.unwrapRemaps(l.decls[name]))
	if err != nil || inherited == nil {
		return onSurface(own)
	}
	// One remap of the difference, not two of the heights: the coat's
	// material is usually built from the one beneath, so one evaluation at
	// the surface point serves both.
	return onSurface(sub(own, height(flattenRemaps(inherited))))
}

// surface is a reference to the closest point, from p, of the undisplaced
// shape a layered shape's coats stand on — the body under its chain of coats
// (oak under coat1 under coat2 …), all of which share it — written once as a
// declaration: p - d * g / |g|^2, d that shape's distance and g its gradient,
// exact for a distance and one Newton step for a bound. nil when the
// gradient cannot be taken, and the heights are read at p.
func (l *layers) surface(name string) *program.Operand {
	root, key, _ := l.chainRoot(name)
	declName := layerSurfPrefix + key
	if l.surfaces == nil {
		l.surfaces = map[string]bool{}
	}
	if ok, done := l.surfaces[key]; done {
		if !ok {
			return nil
		}
		return ref(declName)
	}
	l.surfaces[key] = false
	if l.deriv == nil {
		return nil
	}
	g, err := returnGradient(l.deriv, program.CloneOperand(root))
	if err != nil || g == nil {
		return nil
	}
	gradName := layerSurfGradPrefix + key
	l.declare(gradName, g)
	gg := call("dot", ref(gradName), ref(gradName))
	step := div(mul(program.CloneOperand(root), ref(gradName)), nary(program.Operation_MAX, gg, num(1e-12)))
	l.declare(declName, sub(ref(program.PVarName), step))
	l.surfaces[key] = true
	if l.surfAt == nil {
		l.surfAt = map[string]int{}
	}
	l.surfAt[declName] = len(l.synth)
	return ref(declName)
}

// surfRead is one value read at a surface point: expr evaluated there, to be
// written into slot when the reads are packed (packReads).
type surfRead struct {
	at   string // the surface point's declaration
	expr *program.Operand
	slot *program.Operand
}

// readAt is expr evaluated at the surface point at refers to. Every value
// read at one point is evaluated in one remap (packReads), up to three to a
// vector: each remap is a function of its own to the GPU, which does not share
// work between two, so reading a coat's height and the one beneath it in two
// would evaluate the material beneath twice.
func (l *layers) readAt(at, expr *program.Operand) *program.Operand {
	slot := &program.Operand{}
	l.reads = append(l.reads, surfRead{at: aliasOf(at), expr: expr, slot: slot})
	return slot
}

// packReads writes the reads at each surface point as remaps of up to three
// values at once, declared just after the point, and fills each read's slot
// with its component.
func (l *layers) packReads() {
	if len(l.reads) == 0 {
		return
	}
	// The outermost coats are made last, and they are what the march reads:
	// packed newest first, they share the first remap at their point, and
	// the coats beneath, read only to shade, fill the rest.
	byAt := map[string][]surfRead{}
	var order []string
	for i := len(l.reads) - 1; i >= 0; i-- {
		r := l.reads[i]
		if _, seen := byAt[r.at]; !seen {
			order = append(order, r.at)
		}
		byAt[r.at] = append(byAt[r.at], r)
	}
	inserts := map[int][]*program.Assignment{}
	comps := []string{program.XVarName, program.YVarName, program.ZVarName}
	for _, at := range order {
		reads := byAt[at]
		for g := 0; g*3 < len(reads); g++ {
			chunk := reads[g*3 : min(len(reads), g*3+3)]
			name := fmt.Sprintf("%s%s_%d", layerSurfReadPrefix, strings.TrimPrefix(at, layerSurfPrefix), g)
			var expr *program.Operand
			if len(chunk) == 1 {
				expr = nary(program.Operation_REMAP, chunk[0].expr, ref(at))
				chunk[0].slot.T = ref(name).T
			} else {
				vals := make([]*program.Operand, len(chunk))
				for i, r := range chunk {
					vals[i] = r.expr
					r.slot.T = nary(program.Operation_ACCESS, ref(name), ref(comps[i])).T
				}
				expr = nary(program.Operation_REMAP, nary(program.Operation_COLLECT, vals...), ref(at))
			}
			idx := l.surfAt[at]
			inserts[idx] = append(inserts[idx], &program.Assignment{
				Variables:  []*program.Variable{{Name: name}},
				Expression: expr,
			})
		}
	}
	out := make([]*program.Assignment, 0, len(l.synth)+len(l.reads))
	for i := 0; i <= len(l.synth); i++ {
		out = append(out, inserts[i]...)
		if i < len(l.synth) {
			out = append(out, l.synth[i])
		}
	}
	l.synth = out
}

// chainRoot is the undisplaced body at the bottom of a layered shape's chain
// of coats, each coat a name for the one beneath with a material laid over
// it, and a name to key its surface by. aliases is true when the chain is
// names all the way down to an undisplaced named shape, the case whose
// increments telescope (emit).
//
// A coat laid over a shape that is not itself a coat — paint over a cut —
// reads its heights at the closest point of the chain inside that shape when
// there is exactly one: the coats beneath read theirs there too, so the
// renderer evaluates their materials at one point once rather than at two.
// Such a coat does not telescope (aliases false). Where the shape holds
// several chains, or none, it is its own root.
func (l *layers) chainRoot(name string) (root *program.Operand, key string, aliases bool) {
	key = name
	body := l.decls[name]
	crossed := false
	for depth := 0; depth < 64; depth++ {
		r, ok := body.T.(*program.Operand_Reference)
		if !ok || r.Reference.Prefix != "" {
			return body, key, false
		}
		next, isDecl := l.decls[r.Reference.Name]
		if !isDecl {
			return body, key, false
		}
		key = r.Reference.Name
		if !l.displacing[r.Reference.Name] {
			if !l.contains(r.Reference.Name) {
				return body, key, !crossed
			}
			inner := l.soleChain(next)
			if inner == "" {
				return body, key, false
			}
			crossed = true
			body = ref(inner)
			continue
		}
		body = next
	}
	return body, key, false
}

// soleChain is the one name in an expression (leaving out stand-ins) with a
// moving surface in it, or "" when there are none or several.
func (l *layers) soleChain(op *program.Operand) string {
	refs := map[string]bool{}
	collectPartRefs(op, refs)
	found := ""
	for name := range refs {
		if !l.contains(name) {
			continue
		}
		if found != "" {
			return ""
		}
		found = name
	}
	return found
}

// height is a material record's displacement, its volume row's third
// component.
func height(m *program.Operand) *program.Operand {
	return nary(program.Operation_ACCESS, nary(program.Operation_INDEX, m, num(2)), ref(program.ZVarName))
}

func (l *layers) declare(name string, expr *program.Operand) {
	l.synth = append(l.synth, &program.Assignment{
		Start:      expr.Start,
		End:        expr.End,
		Variables:  []*program.Variable{{Name: name}},
		Expression: expr,
	})
}

// beneath is the weight, 0 or 1, of the inherited value at a point of a
// layered shape: 1 wherever its layer has no thickness, a coat masked off
// there, and 1 inside the solid beneath its layer, so a cut shows what is
// beneath rather than the coat's record.
//
// Inside is read off the solid beneath as drawn. With the layers drawn it is
// within layerInside of that surface. Where nothing is displaced
// (displacementScale() 0: the Detail slider or the displacement knob off, a
// part too small for relief, meshing, export) every coat lies on the
// undisplaced surface, so a point is beneath a coat only when it is clearly
// inside that surface, by layerUndrawn or two pixels, whichever is more: the
// outer surface keeps the top coat, and a cut, well inside, still shows the
// layers it goes through. nil for a shape whose colour and material do not
// switch.
func (l *layers) beneath(name string) *program.Operand {
	if l == nil || !l.inTree[name] {
		return nil
	}
	if aliasOf(l.decls[name]) == "" {
		// A finish set on a shape's own expression (a face of a max) is that
		// face's look whatever its height: it gives way only inside the solid
		// beneath, and only where it has thickness.
		inside := bin(program.Operation_LT, ref(layerUnderPrefix+name), num(layerInside))
		thick := bin(program.Operation_GT, mul(call(program.DisplacementScale), ref(layerIncPrefix+name)), num(layerThick))
		return nary(program.Operation_SELECT, bin(program.Operation_AND, thick, inside), num(1), num(0))
	}
	drawn := bin(program.Operation_GT, call(program.DisplacementScale), num(0))
	undrawn := neg(nary(program.Operation_MAX, num(layerUndrawn), mul(num(2), call(program.PixelSize))))
	edge := nary(program.Operation_SELECT, drawn, num(layerInside), undrawn)
	inside := bin(program.Operation_LT, ref(layerUnderPrefix+name), edge)
	thin := bin(program.Operation_LE, ref(layerIncPrefix+name), num(layerThick))
	return nary(program.Operation_SELECT, bin(program.Operation_OR, thin, inside), num(1), num(0))
}

// switchField is own where w is 0 and inherited where it is 1, row by row
// for a material.
func switchField(spec fieldSpec, w, own, inherited *program.Operand) *program.Operand {
	pick := func(a, b *program.Operand) *program.Operand {
		return nary(program.Operation_SELECT, bin(program.Operation_GT, program.CloneOperand(w), num(0.5)), b, a)
	}
	if spec.rows == 0 {
		return pick(own, inherited)
	}
	rows := make([]*program.Operand, spec.rows)
	for k := range rows {
		rows[k] = pick(nary(program.Operation_INDEX, program.CloneOperand(own), num(float64(k))),
			nary(program.Operation_INDEX, program.CloneOperand(inherited), num(float64(k))))
	}
	return nary(program.Operation_COLLECT, rows...)
}

// attach adds the displaced declarations to the primary package's last file
// and records the displaced returns.
func (l *layers) attach(prog *program.Program) {
	if l == nil {
		return
	}
	l.packReads()
	last := prog.Primary.Files[len(prog.Primary.Files)-1]
	last.Assignments = append(last.Assignments, l.synth...)
	last.DisplacedReturns = l.returns
}
