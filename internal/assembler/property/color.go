package property

import (
	"fmt"
	"sort"

	"gradientcontrol.com/omega/core/program"
)

// colorProperty is the declaration-property name that authors a shape's RGB
// color: `name.color = [r, g, b];` (a length-3 vector field in [0, 1]).
const colorProperty = "color"

// TODO: move material related functionality into separate file
// materialProperty is the declaration-property name that authors a shape's
// surface material: `name.material = material.chrome;`, a list of materialRows
// vec4 rows. Row 0 is [roughness, metallic, ior, reserved], row 1 is
// [emissive r, g, b, strength] (sRGB like color), row 2 is [transmission,
// absorptionDistance, displacement, reserved]. Independent of color: a shape
// may carry either, both, or neither, and each is blended and inherited across
// booleans on its own.
const materialProperty = "material"

// materialRows is the number of vec4 rows in a material record. Rows are
// appended, never reordered, so a document built with the material library's
// constructors keeps working when a row is added.
const materialRows = 4

// materialUnbounded is the finite stand-in for "no upper bound" in a row
// clamp, so one vec4 clamp covers a row that mixes [0,1] channels with open
// ones (ior, emissive strength, displacement in mm).
const materialUnbounded = 1e4

// fieldSpec describes one self-computing declaration property the resolver
// below blends: its name, its shape (one vector, or a list of vec4 rows), the
// value a unit falls back to when it cannot resolve, and the per-row clamp
// applied to a blend. The blend machinery is the same for every property; only
// these details differ.
type fieldSpec struct {
	name     string
	rows     int // 0: the value is one vector; K: a list of K vec4 rows
	fallback func() *program.Operand
	clampRow func(row int, v *program.Operand) *program.Operand
}

var colorSpec = fieldSpec{
	name:     colorProperty,
	rows:     0,
	fallback: func() *program.Operand { return rgbLiteral(0.5) },
	clampRow: func(_ int, v *program.Operand) *program.Operand {
		return call("clamp", v, rgbLiteral(0), rgbLiteral(1))
	},
}

var materialSpec = fieldSpec{
	name:     materialProperty,
	rows:     materialRows,
	fallback: materialDefault,
	clampRow: clampMaterialRow,
}

// materialDefault is the record a shape has when it authors no material:
// roughness 0.5, not metallic, ior 1.5, no emission, opaque, undisplaced, no
// lay. The
// viewer and the material library (material.standard) carry the same values.
func materialDefault() *program.Operand {
	return nary(program.Operation_COLLECT,
		vec4Literal(0.5, 0, 1.5, 0),
		vec4Literal(0, 0, 0, 0),
		vec4Literal(0, 0, 0, 0),
		vec4Literal(0, 0, 0, 0))
}

// clampMaterialRow bounds one blended row: roughness/metallic/transmission and
// emissive rgb to [0,1]; ior to >= 1 (a lower ior gives a nonphysical F0);
// emissive strength and absorption distance to >= 0; displacement signed; the
// lay's direction free and its anisotropy to [0,1].
func clampMaterialRow(row int, v *program.Operand) *program.Operand {
	u := materialUnbounded
	switch row {
	case 0:
		return call("clamp", v, vec4Literal(0, 0, 1, 0), vec4Literal(1, 1, u, u))
	case 1:
		return call("clamp", v, vec4Literal(0, 0, 0, 0), vec4Literal(1, 1, 1, u))
	case 2:
		return call("clamp", v, vec4Literal(0, 0, -u, 0), vec4Literal(1, u, u, u))
	default:
		return call("clamp", v, vec4Literal(-u, -u, -u, 0), vec4Literal(u, u, u, 1))
	}
}

func vec4Literal(a, b, c, d float64) *program.Operand {
	return nary(program.Operation_COLLECT, num(a), num(b), num(c), num(d))
}

// colorMarker wraps a colored shape's body as colorMarker(distance, p). It is
// transparent for the distance (gradDomain.Call returns it), but its second
// argument p is evaluated in the shape's own frame, capturing the coordinate the
// shape is sampled at. The color is then emitted as remap(colorExpr, coord), so
// the whole color — including an oracle like noise.cellularEdges that samples p
// internally — transports through remap with the shape.
const colorMarker = "gsl__color"

// ColorFields returns, for each return of pkg, the vec3 color field of p that a
// renderer samples at the surface hit, or nil for a return that references no
// colored shape (the viewer keeps its palette there).
//
// Color is a self-computing property blended by the same principle as the
// distance: the weight of shape i's color is ∂D/∂dᵢ, the sensitivity of the
// combined field D to that shape's distance dᵢ, computed with forward-mode
// autodiff (the derivative Domain). For a translation-covariant combiner these
// weights are a partition of unity, so the color is a convex blend — hard-select
// for a sharp union (∂min/∂dᵢ ∈ {0,1}), a smooth blend for a smooth union, a
// facet for a chamfer — with no per-combinator rules. getImport resolves the
// library (omega) combinators so the engine can differentiate through them.
//
// Color is resolved compositionally: an expression's color is blended over its
// direct colored references (its "units"), each contributing its own
// recursively-resolved color field. So a named intermediate union acts as a
// single colored object — an uncolored shape grouped with a colored one inside
// that union inherits the group's color and keeps it when the group is composed
// further, instead of the whole thing flattening to the leaf colored shapes.
func ColorFields(pkg *program.Package, getImport program.GetImportFunc) ([]*program.Operand, error) {
	return propertyFields(pkg, getImport, colorSpec)
}

// MaterialFields is ColorFields for the material property: for each return of
// pkg, the field of p whose value is the shape's material record (a list of
// materialRows vec4 rows), or nil for a return that references no shape with a
// material (the renderer uses its default). Blended row by row with the same
// ∂D/∂dᵢ weights, over the material units — independent of the color units.
func MaterialFields(pkg *program.Package, getImport program.GetImportFunc) ([]*program.Operand, error) {
	return propertyFields(pkg, getImport, materialSpec)
}

func propertyFields(pkg *program.Package, getImport program.GetImportFunc, spec fieldSpec) ([]*program.Operand, error) {
	if pkg == nil || !packageHasProperty(pkg, spec.name) {
		return nil, nil
	}
	r := newResolver(pkg, getImport, spec)
	if len(pkg.Files) == 0 {
		return nil, nil
	}
	lastFile := pkg.Files[len(pkg.Files)-1]
	out := make([]*program.Operand, len(lastFile.Returns))
	for i, ret := range lastFile.Returns {
		field, err := r.fieldOf(r.unwrapRemaps(ret))
		if err != nil {
			out[i] = nil // exotic construct we can't differentiate: keep the palette
			continue
		}
		out[i] = flattenRemaps(field)
	}
	return out, nil
}

// newResolver builds the resolver for one property over pkg: its declarations
// and the dot-imported (omega) library as the differentiation namespace, with
// every call that only moves its shape read as the remap it is.
func newResolver(pkg *program.Package, getImport program.GetImportFunc, spec fieldSpec) *resolver {
	// The differentiation namespace: primary declarations plus the dot-imported
	// (omega) library.
	r := &resolver{
		spec:       spec,
		decls:      map[string]*program.Operand{},
		multiDecls: collectPackageMultiAssignments(pkg),
		overloads:  map[string][]*program.Function{},
		props:      collectPackageProperties(pkg, spec.name),
		hasMemo:    map[string]bool{},
		fieldMemo:  map[string]*program.Operand{},
	}
	if spec.name == materialProperty && !usesCall(pkg, getImport, layFunction) {
		r.constantRows = map[int]*program.Operand{materialLayRow: vec4Literal(0, 0, 0, 0)}
	}
	mergePackageNamespace(pkg, r.decls, r.overloads, nil)
	if getImport != nil {
		// Qualified imports stay opaque to the gradient (a noise oracle is a
		// given, not something to differentiate through), but a LIST a library
		// exports — a fastener size table — has to be listable when a shape's
		// body maps over it, or the map sees one opaque element. See
		// frame.listElements.
		r.imports = collectImportNamespaces(pkg, getImport)
		for _, f := range pkg.Files {
			if f == nil {
				continue
			}
			var builtinImp *program.Import
			for _, imp := range f.Imports {
				if imp.Prefix != "." {
					continue // dot-import (omega) only; qualified imports stay opaque
				}
				if imp.Path == "builtin" {
					builtinImp = imp
					continue
				}
				if ip, err := getImport(imp.Path); err == nil && ip != nil {
					mergePackageNamespace(ip, r.decls, r.overloads, nil)
				}
			}
			if builtinImp != nil {
				if ip, err := getImport(builtinImp.Path); err == nil && ip != nil {
					mergePackageNamespace(ip, r.decls, r.overloads, nil)
				}
			}
		}
	}

	// A call that only moves its shape is a remap, so the field moves with it
	// (remap_wrappers.go).
	for name, body := range r.decls {
		r.decls[name] = r.unwrapRemaps(body)
	}
	return r
}

// flattenRemaps composes nested remap operands into a single remap per colored
// leaf: remap(remap(X, A), B) becomes remap(X, A∘B). The compositional resolver
// samples a colored shape through one remap per level it is nested under, so a
// shape under several transforms (e.g. remap(remap(ball, rotate), translate))
// produces nested inline remaps — and the backend lowers a single inline remap
// correctly but drops the OUTER coordinate of a nested one. Composing here
// restores what the pre-compositional (flat) resolver emitted. The innermost
// remap around each leaf is kept, so an oracle inside it (noise) still gets its
// implicit p bound once.
func flattenRemaps(op *program.Operand) *program.Operand {
	if op == nil {
		return nil
	}
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		e := t.Expression
		ops := make([]*program.Operand, len(e.Operands))
		for i, o := range e.Operands {
			ops[i] = flattenRemaps(o)
		}
		if e.Operation == program.Operation_REMAP {
			body, coord := ops[0], ops[1]
			if containsRemap(body) {
				return pushRemap(body, coord) // collapse: push coord into inner remaps
			}
			return nary(program.Operation_REMAP, body, coord) // innermost: keep
		}
		return nary(e.Operation, ops...)
	case *program.Operand_FnCall:
		fc := t.FnCall
		args := make([]*program.Operand, len(fc.Arguments))
		for i, a := range fc.Arguments {
			args[i] = flattenRemaps(a)
		}
		return &program.Operand{T: &program.Operand_FnCall{FnCall: &program.FnCall{
			Reference: &program.Reference{Prefix: fc.Reference.Prefix, Name: fc.Reference.Name}, Arguments: args}}}
	}
	return op
}

// pushRemap evaluates op at p = coord, flattening rather than nesting: implied
// variables become coord's components, and a REMAP(body, inner) composes as
// REMAP(body, pushRemap(inner, coord)) — the coord pushes into the inner
// coordinate while body keeps its own remap frame (its p stays inner, so an
// oracle in body is bound once). It never creates a new outer remap. ACCESS
// selectors (the x/y/z of a swizzle) are left alone; only the base is rebound.
func pushRemap(op, coord *program.Operand) *program.Operand {
	if op == nil {
		return nil
	}
	switch t := op.T.(type) {
	case *program.Operand_Reference:
		switch t.Reference.Name {
		case program.PVarName:
			return coord
		case program.XVarName:
			return nary(program.Operation_ACCESS, coord, ref(program.XVarName))
		case program.YVarName:
			return nary(program.Operation_ACCESS, coord, ref(program.YVarName))
		case program.ZVarName:
			return nary(program.Operation_ACCESS, coord, ref(program.ZVarName))
		}
		return op
	case *program.Operand_Expression:
		e := t.Expression
		switch e.Operation {
		case program.Operation_REMAP:
			return nary(program.Operation_REMAP, e.Operands[0], pushRemap(e.Operands[1], coord))
		case program.Operation_ACCESS:
			ops := make([]*program.Operand, len(e.Operands))
			ops[0] = pushRemap(e.Operands[0], coord)
			for i := 1; i < len(e.Operands); i++ {
				ops[i] = e.Operands[i] // selector, not a rebindable p
			}
			return nary(program.Operation_ACCESS, ops...)
		default:
			ops := make([]*program.Operand, len(e.Operands))
			for i, o := range e.Operands {
				ops[i] = pushRemap(o, coord)
			}
			return nary(e.Operation, ops...)
		}
	case *program.Operand_FnCall:
		fc := t.FnCall
		args := make([]*program.Operand, len(fc.Arguments))
		for i, a := range fc.Arguments {
			args[i] = pushRemap(a, coord)
		}
		return &program.Operand{T: &program.Operand_FnCall{FnCall: &program.FnCall{
			Reference: &program.Reference{Prefix: fc.Reference.Prefix, Name: fc.Reference.Name}, Arguments: args}}}
	}
	return op
}

// containsRemap reports whether op has a REMAP anywhere in its tree.
func containsRemap(op *program.Operand) bool {
	if op == nil {
		return false
	}
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		if t.Expression.Operation == program.Operation_REMAP {
			return true
		}
		for _, o := range t.Expression.Operands {
			if containsRemap(o) {
				return true
			}
		}
	case *program.Operand_FnCall:
		for _, a := range t.FnCall.Arguments {
			if containsRemap(a) {
				return true
			}
		}
	}
	return false
}

// resolver resolves each declaration's property field compositionally. decls
// and overloads are the differentiation namespace (primary declarations plus
// the dot-imported omega library); props holds the explicitly authored
// `.color` / `.material` expressions (the leaves) for the property in spec.
// hasMemo/fieldMemo memoize the two recursions below.
type resolver struct {
	spec       fieldSpec
	decls      map[string]*program.Operand
	multiDecls map[string]multiBinding // names bound by a destructuring assignment
	overloads  map[string][]*program.Function
	imports    map[string]*namespace // qualified imports, for listing a library's lists
	props      map[string]*program.Operand
	hasMemo    map[string]bool
	fieldMemo  map[string]*program.Operand
	// constantRows are rows every unit has the same constant in, which blend
	// takes as they are: a material's lay where the program gives none a lay.
	constantRows map[int]*program.Operand
}

// hasField reports whether a declaration carries the property — it has an
// explicit assignment, or it transitively references a shape that does.
// Because it is transitive, a reference without the property contains no
// shape with it, so units need only look at direct references. Memoized; the
// tentative-false write also breaks any reference cycle.
func (r *resolver) hasField(name string) bool {
	if v, ok := r.hasMemo[name]; ok {
		return v
	}
	if _, has := r.props[name]; has {
		r.hasMemo[name] = true
		return true
	}
	r.hasMemo[name] = false
	body, ok := r.decls[name]
	if !ok {
		return false
	}
	refs := map[string]bool{}
	collectRefs(body, refs)
	for ref := range refs {
		if r.hasField(ref) {
			r.hasMemo[name] = true
			return true
		}
	}
	return false
}

// units are the property units of an expression: its direct references that
// carry the property. Each is the ∂/∂dᵤ this level differentiates with respect
// to — a whole named sub-object — and contributes its own recursively-resolved
// field.
func (r *resolver) units(expr *program.Operand) []string {
	refs := map[string]bool{}
	collectRefs(expr, refs)
	var u []string
	for name := range refs {
		if r.hasField(name) {
			u = append(u, name)
		}
	}
	sort.Strings(u) // deterministic output
	return u
}

// resolveField returns a declaration's field: its explicit property if it has
// one, otherwise the composed field of its body. nil for a name without it.
func (r *resolver) resolveField(name string) (*program.Operand, error) {
	if c, ok := r.fieldMemo[name]; ok {
		return c, nil
	}
	stand := simplifiedCall(r.decls[name])
	if c, has := r.props[name]; has {
		if stand != nil {
			r.fieldMemo[name] = nil // tentative (breaks cycles)
			var err error
			if c, err = r.simplifiedField(c, stand); err != nil {
				return nil, err
			}
		}
		r.fieldMemo[name] = c
		return c, nil
	}
	r.fieldMemo[name] = nil // tentative (breaks cycles)
	body, ok := r.decls[name]
	if !ok {
		return nil, nil
	}
	if stand != nil {
		// The part's own field is its full model's; the stand-in's joins it
		// below, blended only in the viewer.
		body = stand.Arguments[0]
	}
	field, err := r.fieldOf(body)
	if err != nil {
		return nil, err
	}
	if stand != nil {
		if field, err = r.simplifiedField(field, stand); err != nil {
			return nil, err
		}
	}
	r.fieldMemo[name] = field
	return field, nil
}

// simplifiedField is a part's field with its stand-in's, stand being the
// part's simplified(part, proxy, lo, hi) (MarkParts): builtin's simplified
// again, over the fields, simplified(field, proxyField, lo, hi), the part's
// field everywhere but the viewer. A part without the property takes the
// property's default, and a stand-in without it the part's, whose own .simple
// flattens it: the field is then the part's alone, since switching between two
// copies of it would only evaluate it twice.
func (r *resolver) simplifiedField(field *program.Operand, stand *program.FnCall) (*program.Operand, error) {
	proxy, err := r.fieldOf(stand.Arguments[1])
	if err != nil {
		return nil, err
	}
	if field == nil {
		field = r.spec.fallback()
	}
	if proxy == nil {
		return field, nil
	}
	return call(simplifiedName, field, proxy,
		program.CloneOperand(stand.Arguments[2]), program.CloneOperand(stand.Arguments[3])), nil
}

// fieldOf computes an expression's property field, or nil if it references no
// shape carrying the property. It blends the expression's units by ∂D/∂dᵤ and
// fills the remainder with the nearest unit's value — the same gradient blend
// used at every level. The compositionality is entirely in the choice of units
// (each a whole named sub-object, opaque here) and their recursively-resolved
// fields, so a named intermediate union keeps its own value when composed
// further instead of flattening to the leaf shapes.
func (r *resolver) fieldOf(expr *program.Operand) (*program.Operand, error) {
	units := r.units(expr)
	if len(units) == 0 {
		return nil, nil
	}
	// A single unit with no coordinate change: the blend is the unit's value
	// exactly — gradientBlend over one unit is g·c + max(0,1−g)·c normalized,
	// which is c for any weight — so skip the weight gradient entirely. This
	// keeps a derived scalar built from one colored shape
	// (edges = f(combined, combined.second)) from differentiating through its
	// whole expression just to reproduce the unit's color. A remap anywhere in
	// the expression falls through to the full path, which captures the
	// coordinate the field must be sampled at.
	if len(units) == 1 && !containsRemap(expr) {
		c, err := r.resolveField(units[0])
		if err != nil {
			return nil, err
		}
		if c == nil {
			c = r.spec.fallback() // a unit always resolves; stay safe if not
		}
		return c, nil
	}
	// At this level the units are opaque, colorMarker-wrapped leaves (so each
	// unit's sample coordinate probes for transport); every other declaration
	// inlines normally.
	levelDecls := make(map[string]*program.Operand, len(r.decls))
	for k, v := range r.decls {
		levelDecls[k] = v
	}
	for _, u := range units {
		body, ok := r.decls[u]
		if !ok {
			// A destructured name (a, b = f()) has no per-name body to wrap: it
			// resolves through the engine's multi-assignment binding, keeping its
			// seed tangent via Named, but skips the coordinate probe (no remap
			// transport for such a unit).
			continue
		}
		levelDecls[u] = call(colorMarker, body, ref(program.PVarName))
	}

	gs := make([]*program.Operand, len(units))
	coords := map[string]*program.Operand{}
	for i, u := range units {
		g, err := shapeGradient(expr, u, units, levelDecls, r.multiDecls, r.overloads, r.imports, coords)
		if err != nil {
			return nil, err
		}
		gs[i] = g
	}
	cs := make([]*program.Operand, len(units))
	// WHEN THE WEIGHTS PARTITION UNITY THE FILL IS DEAD CODE. rest is
	// max(1 - Σg, 0), so Σg ≡ 1 makes it identically zero and the whole
	// nearest-color expression is multiplied by nothing — while still being
	// emitted, still carrying a remap per unit, and still costing every pixel
	// that evaluates the field. In a three-sphere union 8 of the 11 generated
	// remap functions existed only inside that zero. Skipping it here is not an
	// optimization of the blend; it is declining to write down a term that
	// cannot contribute.
	unitSet := make(map[string]bool, len(units))
	for _, u := range units {
		unitSet[u] = true
	}
	exact := partitionsUnity(expr, unitSet, r.decls, 0)

	// ds is the units' distances, for the fill's nearest-unit chain. When the
	// fill is elided they are not built at all: the chain is the only thing
	// that reads them.
	var ds []*program.Operand
	if !exact {
		ds = make([]*program.Operand, len(units))
	}
	for i, u := range units {
		c, err := r.resolveField(u)
		if err != nil {
			return nil, err
		}
		if c == nil {
			c = r.spec.fallback() // a unit always resolves; stay safe if not
		}
		cs[i] = c
		if exact {
			continue
		}
		d := ref(u)
		if q, ok := coords[u]; ok {
			d = nary(program.Operation_REMAP, d, q) // the unit's distance in its own frame, for the nearest-unit fill
		}
		ds[i] = d
	}
	return r.blend(gs, cs, ds, units, coords, exact), nil
}

// blend forms the gradient blend of the units' values. A one-vector property
// (color) blends the values directly. A row property (material) blends each
// row on its own — INDEX(value, k) per unit, the shared weights, the nearest
// fill over that row — and collects the rows back into a list, so the backend
// lowers each row as a vec4 and never needs arithmetic on lists. A unit whose
// coordinate was captured under a remap samples its value in its own frame:
// the remap wraps the vector (or the indexed row), matching the distance in ds.
func (r *resolver) blend(gs, cs, ds []*program.Operand, units []string, coords map[string]*program.Operand, exact bool) *program.Operand {
	inFrame := func(v *program.Operand, u string) *program.Operand {
		if q, ok := coords[u]; ok {
			return nary(program.Operation_REMAP, v, q)
		}
		return v
	}
	if r.spec.rows == 0 {
		vals := make([]*program.Operand, len(units))
		for i, u := range units {
			vals[i] = inFrame(cs[i], u)
		}
		if exact {
			return r.spec.clampRow(0, exactBlend(gs, vals))
		}
		return r.spec.clampRow(0, gradientBlend(gs, vals, nearestValue(ds, vals)))
	}
	rows := make([]*program.Operand, r.spec.rows)
	for k := range rows {
		if c, ok := r.constantRows[k]; ok {
			rows[k] = c
			continue
		}
		vals := make([]*program.Operand, len(units))
		for i, u := range units {
			vals[i] = inFrame(nary(program.Operation_INDEX, cs[i], num(float64(k))), u)
		}
		if exact {
			rows[k] = r.spec.clampRow(k, exactBlend(gs, vals))
		} else {
			rows[k] = r.spec.clampRow(k, gradientBlend(gs, vals, nearestValue(ds, vals)))
		}
	}
	return nary(program.Operation_COLLECT, rows...)
}

// exactBlend is gradientBlend for the case where the weights are known to sum
// to 1: the convex blend Σ gᵢ·vᵢ, with no leftover weight to fill and no
// normalization to apply. Identical in value to gradientBlend with the same
// weights — rest is 0 and denom is 1 there — and without the fill's cost.
func exactBlend(gs, vals []*program.Operand) *program.Operand {
	numer := mul(gs[0], vals[0])
	for i := 1; i < len(gs); i++ {
		numer = add(numer, mul(gs[i], vals[i]))
	}
	return numer
}

// partitionsUnity reports whether the combiner weights ∂D/∂dᵤ necessarily sum
// to exactly 1 at every point, which decides whether the nearest-color fill can
// contribute anything at all.
//
// It is true exactly when D is a tree of SHARP min/max whose every leaf is one
// of the colored units. A sharp min or max passes its whole derivative to
// whichever branch is active and none to the others, so the weights are one-hot
// and sum to 1; remap only changes where the branches are sampled, not which
// one wins.
//
// BOTH HALVES MATTER, and the second is the easy one to forget. An UNCOLORED
// leaf in the same union — min(coloredA, plainB) — takes the derivative for
// itself wherever it is nearest, and the sum over the colored units is 0 there,
// not 1. That is precisely the case the fill exists for: the region belongs to
// no colored shape and inherits the nearest one's color. So a single uncolored
// leaf anywhere in the tree makes the fill live, and this must answer false.
//
// Everything else answers false as well: a smooth or rounded union spreads the
// derivative across branches without summing to 1, a subtraction flips a sign,
// and arithmetic on a distance is not a combiner at all. False only costs the
// fill, which is what the code did unconditionally before.
func partitionsUnity(op *program.Operand, units map[string]bool,
	decls map[string]*program.Operand, depth int) bool {
	if op == nil || depth > 64 {
		return false
	}
	switch t := op.T.(type) {
	case *program.Operand_Reference:
		name := t.Reference.Name
		if t.Reference.Prefix != "" {
			return false // a qualified name is opaque here
		}
		if units[name] {
			return true
		}
		body, ok := decls[name]
		if !ok {
			return false // p, time, a parameter: not a colored leaf
		}
		return partitionsUnity(body, units, decls, depth+1)
	case *program.Operand_Expression:
		e := t.Expression
		switch e.Operation {
		case program.Operation_MIN, program.Operation_MAX:
			if len(e.Operands) == 0 {
				return false
			}
			for _, o := range e.Operands {
				if !partitionsUnity(o, units, decls, depth+1) {
					return false
				}
			}
			return true
		case program.Operation_REMAP:
			// The coordinate is not a distance and carries no weight.
			return len(e.Operands) > 0 &&
				partitionsUnity(e.Operands[0], units, decls, depth+1)
		}
		return false
	}
	return false
}

// shapeGradient computes ∂ret/∂(shape target) as a field of p, treating each
// colored shape's whole field as an independent variable (seed 1 on target, 0 on
// the others, 0 on p/time) and differentiating everything else — combiners
// included. The colored shapes stay in decls so they re-derive under any remap;
// gradDomain injects the seed as their tangent while keeping that re-derived
// primal, so both the weight and the shape distance follow the transform.
func shapeGradient(ret *program.Operand, target string, colored []string,
	decls map[string]*program.Operand, multiDecls map[string]multiBinding,
	overloads map[string][]*program.Function, imports map[string]*namespace,
	coords map[string]*program.Operand) (*program.Operand, error) {
	seed := map[string]*program.Operand{
		program.TimeVarName: num(0),
		program.PVarName:    nary(program.Operation_COLLECT, num(0), num(0), num(0)),
		program.XVarName:    num(0),
		program.YVarName:    num(0),
		program.ZVarName:    num(0),
	}
	coloredSet := make(map[string]bool, len(colored))
	for _, name := range colored {
		coloredSet[name] = true
		if name == target {
			seed[name] = num(1)
		} else {
			seed[name] = num(0)
		}
	}
	deriv := &derivative{decls: decls, funcs: map[string][]*program.Function{}, imports: imports}
	dom := &gradDomain{derivDomain: &derivDomain{d: deriv, seed: seed}, colored: coloredSet, coords: coords}
	eng := &Engine{dom: dom, decls: decls, multiDecls: multiDecls, funcs: overloads, imports: imports}
	v, err := eng.Eval(ret)
	if err != nil {
		return nil, err
	}
	d, ok := v.(dual)
	if !ok {
		return nil, fmt.Errorf("color gradient of a non-scalar return")
	}
	return d.tangent, nil
}

// gradDomain is derivDomain specialized for the color gradient. At a colored
// shape it injects the seed as the tangent (the shape's field is the
// differentiation variable) and keeps the re-derived primal — the remap the
// shape sits under is applied externally, so its bare name would sample the
// wrong coordinate. Every other declaration keeps a live reference (like
// derivDomain), so sliders stay live and a remap baked into a decl's own
// definition still lowers correctly.
type gradDomain struct {
	*derivDomain
	colored map[string]bool
	pending *program.Operand            // coordinate captured by the last colorMarker call
	coords  map[string]*program.Operand // shape name -> the coordinate it is sampled at (after remap)
}

// Call intercepts the colorMarker: it passes the distance through unchanged and
// stashes the coordinate probe (p evaluated in the shape's frame, so remapped).
func (g *gradDomain) Call(name string, args []Value) (Value, bool, error) {
	if name == colorMarker && len(args) == 2 {
		g.pending = args[1].(dual).primal
		return args[0], true, nil
	}
	return nil, false, nil
}

// OpaqueCall / OpaqueRef reconstruct a qualified (map./noise./…) call or
// reference as its own operand with no tangent: its value is spliced through and
// the backend resolves the import, while it contributes nothing to the gradient
// (a colored shape's own tangent is injected at Named, not from its body).
func (g *gradDomain) OpaqueCall(pkgName, name string, args []Value) Value {
	primals := make([]*program.Operand, len(args))
	for i, a := range args {
		primals[i] = a.(dual).primal
	}
	return dual{primal: &program.Operand{T: &program.Operand_FnCall{FnCall: &program.FnCall{
		Reference: &program.Reference{Prefix: pkgName, Name: name}, Arguments: primals}}}, tangent: num(0)}
}

func (g *gradDomain) OpaqueRef(pkgName, name string) Value {
	return dual{primal: &program.Operand{T: &program.Operand_Reference{Reference: &program.Reference{
		Prefix: pkgName, Name: name}}}, tangent: num(0)}
}

// Unnamed is Named for a declaration a remap keeps as its value: a unit still
// takes its seed as its tangent and records the coordinate it is sampled at,
// and its primal is the value evaluated under the remap's p.
func (g *gradDomain) Unnamed(name string, value Value) Value {
	d := value.(dual)
	if !g.colored[name] {
		return d
	}
	if g.pending != nil {
		g.coords[name] = g.pending
		g.pending = nil
	}
	return dual{primal: d.primal, tangent: g.derivDomain.seed[name]}
}

func (g *gradDomain) Named(name string, value Value) Value {
	d := value.(dual)
	if g.colored[name] {
		if g.pending != nil {
			g.coords[name] = g.pending
			g.pending = nil
		}
		primal := d.primal
		if q, probed := g.coords[name]; probed && isPVar(q) {
			// NAME THE UNIT RATHER THAN INLINING IT. The unit is sampled at the
			// enclosing p, so its name denotes exactly the distance the weight
			// needs, and the backend emits a named declaration once as a `let`
			// and refers to it thereafter.
			//
			// This is the difference between a cubic color field and a quadratic
			// one. Each unit's weight is a chain of selects whose condition at
			// level k compares the running min of the first k distances against
			// the next, so a unit's distance appears O(n) times in its own weight
			// and O(n²) times across all n of them. Spliced in, every one of
			// those is a full copy of the shape's field; named, it is an
			// identifier. Twelve colored spheres went from 48603 bytes of WGSL to
			// 11917.
			//
			// Only when the probe came back as the bare p. A unit read under a
			// remap written in the EXPRESSION — min(remap(ball, p - off), box) —
			// no longer reaches Named: it varies with p, so the engine keeps it as
			// its value under the remap and hands it to Unnamed instead. Every
			// unit probed here was probed at p, and the guard states why a bare
			// name is the right distance rather than deciding anything.
			primal = ref(name)
		}
		return dual{primal: primal, tangent: g.derivDomain.seed[name]}
	}
	return dual{primal: ref(name), tangent: d.tangent}
}

// gradientBlend forms the blend (Σ gᵢ·cᵢ + rest·fill) / max(Σg, 1), where rest =
// max(0, 1 − Σg) is the weight attributed to no unit. That leftover weight is
// painted with fill — the value of the nearest unit the region composes with —
// so a shape without the property takes on the value of the shapes it is
// combined with rather than a flat default. For a partition of unity (Σg = 1)
// this is Σ gᵢ·cᵢ; for a non-covariant blend (Σg > 1, e.g. a chamfer facet)
// the max(Σg,1) normalizes it. The caller clamps the result per property.
func gradientBlend(gs, cs []*program.Operand, fill *program.Operand) *program.Operand {
	total := gs[0]
	numer := mul(gs[0], cs[0])
	for i := 1; i < len(gs); i++ {
		total = add(total, gs[i])
		numer = add(numer, mul(gs[i], cs[i]))
	}
	rest := bin(program.Operation_MAX, sub(num(1), total), num(0))
	numer = add(numer, mul(rest, fill))
	denom := bin(program.Operation_MAX, total, num(1))
	return div(numer, denom)
}

// nearestValue is the value of the unit with the smallest distance — what a
// region without the property inherits from the unit nearest it in the
// composition. A select chain over the present units; with a single unit it is
// just that unit's value. ds are the units' composed distances, remap-wrapped
// to match cs.
//
// THE RUNNING DISTANCE IS A MIN, NOT A SELECT, and that is the whole cost of
// this function. select(dᵢ < best, dᵢ, best) and min(best, dᵢ) are the same
// number, but the select NAMES best twice — once in the comparison and once as
// the false arm — so the accumulated expression DOUBLES on every colored shape.
// Written that way the color field is Θ(2ⁿ) in the number of colored objects:
// eight colored shapes in a flat union emitted 389 synthesized remaps where
// this emits 50, and a board of thirty parts reached 2.7 MB of WGSL. min names
// it once, which is the difference between exponential and quadratic.
//
// Anything rewriting this loop has to keep that property. The residual
// quadratic is bestC: each cond carries the running distance, so the chain is
// Θ(n²) — worth fixing separately, and nothing like the same order of problem.
func nearestValue(ds, cs []*program.Operand) *program.Operand {
	bestD, bestC := ds[0], cs[0]
	for i := 1; i < len(ds); i++ {
		cond := bin(program.Operation_LT, ds[i], bestD)
		bestC = nary(program.Operation_SELECT, cond, cs[i], bestC)
		bestD = bin(program.Operation_MIN, bestD, ds[i])
	}
	return bestC
}

// isPVar reports whether op is the bare coordinate p — an identity probe, so
// the shape is sampled where the expression is.
func isPVar(op *program.Operand) bool {
	if op == nil {
		return false
	}
	t, ok := op.T.(*program.Operand_Reference)
	return ok && t.Reference.Prefix == "" && t.Reference.Name == program.PVarName
}

func rgbLiteral(v float64) *program.Operand {
	return nary(program.Operation_COLLECT, num(v), num(v), num(v))
}

// --- package scanning -------------------------------------------------------

func packageHasProperty(pkg *program.Package, name string) bool {
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, a := range f.Assignments {
			if a == nil {
				continue
			}
			for _, v := range a.Variables {
				if v == nil {
					continue
				}
				for _, p := range v.Properties {
					if p.Name == name {
						return true
					}
				}
			}
		}
	}
	return false
}

func collectPackageProperties(pkg *program.Package, name string) map[string]*program.Operand {
	m := map[string]*program.Operand{}
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, a := range f.Assignments {
			if a == nil {
				continue
			}
			// Every variable, including the names of a destructuring assignment
			// (a, b = f()) — a property set on a destructured name counts.
			for _, v := range a.Variables {
				for _, p := range v.Properties {
					if p.Name == name {
						m[v.Name] = p.Expression
					}
				}
			}
		}
	}
	return m
}

func mergePackageNamespace(pkg *program.Package,
	decls map[string]*program.Operand, overloads map[string][]*program.Function, funcsMap map[string]*program.Function) {
	if pkg == nil {
		return
	}
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, a := range f.Assignments {
			if a == nil || len(a.Variables) != 1 {
				continue
			}
			name := a.Variables[0].Name
			if _, exists := decls[name]; !exists {
				decls[name] = a.Expression
			}
		}
		for _, ft := range f.FnTypes {
			if ft == nil {
				continue
			}
			for _, group := range ft.Groups {
				if group == nil {
					continue
				}
				for _, fn := range group.Functions {
					if fn != nil {
						overloads[ft.Name] = append(overloads[ft.Name], fn)
						if funcsMap != nil {
							funcsMap[ft.Name] = fn
						}
					}
				}
			}
		}
	}
}

func collectRefs(op *program.Operand, out map[string]bool) {
	if op == nil {
		return
	}
	switch t := op.T.(type) {
	case *program.Operand_Reference:
		if t.Reference.Prefix == "" {
			out[t.Reference.Name] = true
		}
	case *program.Operand_Expression:
		for _, o := range t.Expression.Operands {
			collectRefs(o, out)
		}
	case *program.Operand_FnCall:
		for _, a := range t.FnCall.Arguments {
			collectRefs(a, out)
		}
	}
}
