package property

import (
	"fmt"

	"google.golang.org/protobuf/proto"
	"gradientcontrol.com/omega/core/program"
)

// PROPERTIES CARRIED OUT OF A FUNCTION. A property is a fact about a named
// object, and the resolver reads the names a PACKAGE declares — so a function
// that builds a part and paints it,
//
//	finished(r) {
//	  s = sphere(r);
//	  s.material = material.chrome;
//	  return s;
//	}
//
// used to hand back the distance and drop the paint at the return. This pass
// carries it out. For every function whose single return is a local that
// carries a property, it writes a COMPANION function beside it — the same
// parameters and body, returning that property's expression instead of the
// local — and at every call in the primary package's declarations it gives
// the called part that property, as a call to the companion with the same
// arguments:
//
//	a = finished(1.0);    becomes   gslCarry1 = finished__material(1.0);
//	                                a = finished(1.0);  a.material = gslCarry1;
//
//	b = remap(finished(1.0), q);  becomes   gslCarry1 = finished__material(1.0);
//	                                        gslCarry2 = finished(1.0);
//	                                        gslCarry2.material = gslCarry1;
//	                                        b = remap(gslCarry2, q);
//
// A call nested in an expression is hoisted into a declaration of its own, so
// the property belongs to exactly the part the function returned, and every
// resolver and backend downstream sees an ordinary package-level property. A
// property the caller sets on the name itself wins over the carried one.
//
// THE VALUE IS A DECLARATION TOO, not a call written into the property. The
// inline pass reads declarations and not properties, and it is what folds a
// call with constant arguments that does not read p — a catalog lookup — into
// its value. Written into the property, a companion's thread.metricPitch(5.0)
// reached the backend as the whole unrolled table lookup, once per part and
// per material row.
//
// A MULTI-RETURN function carries each of its parts the same way. Every
// returned local that carries a property gets a companion per property and per
// place in the return, and a destructuring declaration of the call gives each
// name the properties of the part it receives:
//
//	pair() { a = sphere(1.0); a.color = red; b = box(...); b.color = blue; return a, b; }
//
//	x, y = pair();   becomes   gslCarry1 = pair__color__0();
//	                           gslCarry2 = pair__color__1();
//	                           x, y = pair();  x.color = gslCarry1;  y.color = gslCarry2;
//
// A multi-return call can only be destructured, so only a declaration that is
// the call is rewritten; there is no nested call to hoist.
//
// CARRIERS COMPOSE. A function that returns another carrier's part carries
// what that part carries — the local it returns is the call's result, or one
// place of a destructured call, or the return is the call itself — and its
// companion asks the inner one's:
//
//	lacquered(r) { s = finished(r); return s; }
//
//	lacquered__material(r) { s = finished(r); return finished__material(r); }
//
// so a library function can build on another without dropping its paint.
// Carriers are found to a fixed point, since a wrapper of a wrapper is one only
// once the wrapper is, and a property the returned local sets itself wins over
// the one its call carries. A part that is only an argument of the returned
// expression — remap(finished(r), q), say — carries nothing out.
//
// A PART'S STAND-IN IS CARRIED TOO, name.simple = proxy, so a library can hand
// back a fastener that knows its simpler model:
//
//	screw(d) {
//	  s = cap(d);
//	  stand = cylinder(d);
//	  stand.color = grey;
//	  s.simple = stand;
//	  return s;
//	}
//
//	a = place(screw(5.0), at);   becomes   gslCarry2 = screw__simple__color(5.0);
//	                                       gslCarry1 = screw__simple(5.0);
//	                                       gslCarry1.color = gslCarry2;
//	                                       a = place(screw(5.0), at);
//	                                       a.simple = carriedSimple(place(gslCarry1, at));
//
// The stand-in belongs to the declaration, not to a hoisted call: a part is a
// named value placed in the document, so the stand-in is the declaration's own
// expression with each call that carries one replaced by its companion, placed
// exactly as the part is. Its companion keeps the body's properties, so a
// stand-in can wear a colour and material of its own, carried out of the
// companion in turn; one the declaration sets itself is the stand-in's as well
// as the part's, as setting a part's colour replaces the library's. Give the
// stand-in its colour as a property of a local, as above, and not as an
// argument of a helper that sets it: the stand-in's distance is the helper's
// call, and evaluates every argument. builtin's
// carriedSimple marks it, so MarkParts gives it to the declaration when that
// is a part and drops it quietly when it is not, a part read through a remap
// say, where no stand-in can be shown.
//
// Only a SHAPE's stand-in is carried. A colour or material declares one too,
// its flat look, and Expand lowers that inside the function, where it stays.
// Types are not known yet, so a returned local counts as a shape when it sets
// a property only a shape takes, a box, colour, material or frame, or its call
// carries one; a function's shape stand-in on any other local is lowered as a
// value's, and CheckSimple reports it in the document and takes it off in a
// library, where the part is drawn in full. The carried stand-in is taken off the
// function's own local, so Expand never lowers it there.

// carriedProperties are the properties a function's returned local can carry.
// The frame coordinates u and v are carried only in a program that reads one:
// in any other, eraseFrames has taken them off every function first.
var carriedProperties = []string{"color", "material", uProperty, vProperty, simplePropName}

// shapeProperties are the properties only a shape takes, which tell a shape's
// stand-in from a value's (see A PART'S STAND-IN above).
var shapeProperties = []string{"box", "color", "material", uProperty, vProperty}

// carriedSimpleName is builtin's marker for a stand-in carried out of a
// function (3d.gsl).
const carriedSimpleName = "carriedSimple"

// carrierKey identifies one overload: the package, the function name and its
// parameter count.
type carrierKey struct {
	pkgPath string
	name    string
	params  int
}

// CarryReturnProperties rewrites prog in place: companions are added to the
// packages that declare carriers, and the primary package's calls to them get
// the carried properties. It returns prog unchanged when no function carries a
// property.
func CarryReturnProperties(prog *program.Program) error {
	if prog == nil || prog.Primary == nil {
		return nil
	}
	carriers := map[carrierKey]carrier{}
	pkgs := append([]*program.Package{prog.Primary}, prog.Dependencies...)
	// The primitives declare their frames, and only a program that reads one
	// pays for them: in any other, the declarations are erased before anything
	// sees them, so it compiles exactly as it did before frames existed.
	if !packageReadsFrame(prog.Primary) {
		for _, pkg := range pkgs {
			eraseFrames(pkg)
		}
	}
	paths := make([]string, len(pkgs))
	for i, pkg := range pkgs {
		if pkg != nil && i > 0 {
			paths[i] = pkg.Path // references into the primary package carry no path
		}
	}
	// Find every carrier before writing any companion: whether a function
	// carries can depend on whether a function it calls does. A stand-in's
	// companion can carry the stand-in's own colour and material, so the
	// companions are searched in turn, until a round writes none.
	slotsOf := map[*program.Function][]map[string]*program.Operand{}
	written := map[*program.Function]bool{}
	for {
		for changed := true; changed; {
			changed = false
			for i, pkg := range pkgs {
				if pkg != nil && findCarriers(pkg, paths[i], carriers, slotsOf, written) {
					changed = true
				}
			}
		}
		added := false
		for i, pkg := range pkgs {
			if pkg != nil && addCompanions(pkg, paths[i], slotsOf, written) {
				added = true
			}
		}
		if !added {
			break
		}
	}
	if len(carriers) == 0 {
		return nil
	}
	carryCalls(prog.Primary, carriers)
	return nil
}

// carrier is what one overload carries: for each place in its return, the
// properties the part returned there carries. multi marks a multi-return
// overload, whose companions are named per place.
type carrier struct {
	multi bool
	slots [][]string
}

// findCarriers records in carriers, and in slotsOf the carried expressions
// of, each of pkg's functions that carries more than it was known to, and
// reports whether it found any. A function whose companions are written is
// settled, and is not looked at again.
func findCarriers(pkg *program.Package, pkgPath string, carriers map[carrierKey]carrier,
	slotsOf map[*program.Function][]map[string]*program.Operand, written map[*program.Function]bool) bool {
	changed := false
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, ft := range f.FnTypes {
			if ft == nil {
				continue
			}
			for _, g := range ft.Groups {
				if g == nil {
					continue
				}
				for _, fn := range g.Functions {
					if written[fn] {
						continue
					}
					slots := returnedSlots(fn, pkgPath, carriers)
					if slots == nil || countProps(slots) == countProps(slotsOf[fn]) {
						continue
					}
					slotsOf[fn] = slots
					c := carrier{multi: len(fn.Returns) > 1, slots: make([][]string, len(slots))}
					for i, props := range slots {
						for _, prop := range carriedProperties {
							if _, ok := props[prop]; ok {
								c.slots[i] = append(c.slots[i], prop)
							}
						}
					}
					carriers[carrierKey{pkgPath: pkgPath, name: ft.Name, params: int(g.ParamCount)}] = c
					changed = true
				}
			}
		}
	}
	return changed
}

// countProps is how many properties slots carries in all.
func countProps(slots []map[string]*program.Operand) int {
	n := 0
	for _, props := range slots {
		n += len(props)
	}
	return n
}

// addCompanions writes the companion functions for pkg's carriers, whose
// carried expressions findCarriers left in slotsOf, for each carrier not yet
// written, and reports whether it wrote any. A shape's stand-in its returned
// local sets is taken off the local first, so neither the function nor a
// companion lowers it as a value's.
func addCompanions(pkg *program.Package, pkgPath string, slotsOf map[*program.Function][]map[string]*program.Operand,
	written map[*program.Function]bool) bool {
	added := false
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		companions := map[string]*program.FnType{}
		var order []string
		for _, ft := range f.FnTypes {
			if ft == nil {
				continue
			}
			for _, g := range ft.Groups {
				if g == nil {
					continue
				}
				for _, fn := range g.Functions {
					slots := slotsOf[fn]
					if slots == nil || written[fn] {
						continue
					}
					written[fn] = true
					added = true
					multi := len(fn.Returns) > 1
					for i, props := range slots {
						if _, ok := props[simplePropName]; ok {
							dropReturnedSimple(fn, i)
						}
					}
					for i, props := range slots {
						for _, prop := range carriedProperties {
							expr, ok := props[prop]
							if !ok {
								continue
							}
							name := companionFor(ft.Name, prop, i, multi)
							ct, ok := companions[name]
							if !ok {
								ct = &program.FnType{Name: name}
								companions[name] = ct
								order = append(order, name)
							}
							ct.Groups = append(ct.Groups, &program.FnGroup{
								ParamCount: g.ParamCount,
								Functions:  []*program.Function{companionOf(fn, expr, prop == simplePropName)},
							})
						}
					}
				}
			}
		}
		for _, name := range order {
			f.FnTypes = append(f.FnTypes, companions[name])
		}
	}
	return added
}

// dropReturnedSimple takes the stand-in off the local fn returns at place i,
// when the local sets one itself.
func dropReturnedSimple(fn *program.Function, i int) {
	a, place := returnedLocal(fn, fn.Returns[i])
	if a == nil {
		return
	}
	v := a.Variables[place]
	kept := v.Properties[:0]
	for _, p := range v.Properties {
		if p != nil && p.Name == simplePropName {
			continue
		}
		kept = append(kept, p)
	}
	v.Properties = kept
}

// returnedLocal is the assignment that last gives the bare name ret its value
// in fn's body, and the place of the name in it, or nil when ret is not one.
func returnedLocal(fn *program.Function, ret *program.Operand) (*program.Assignment, int) {
	r, ok := ret.T.(*program.Operand_Reference)
	if !ok || r.Reference == nil || r.Reference.Prefix != "" {
		return nil, 0
	}
	var last *program.Assignment
	place := 0
	for _, a := range fn.Assignments {
		if a == nil {
			continue
		}
		for i, v := range a.Variables {
			if v != nil && v.Name == r.Reference.Name {
				last, place = a, i
			}
		}
	}
	return last, place
}

// companionFor names the companion that returns prop for function name: of
// its return, or of the part at place i of a multi-return.
func companionFor(name, prop string, i int, multi bool) string {
	if multi {
		return fmt.Sprintf("%s__%s__%d", name, prop, i)
	}
	return name + "__" + prop
}

// returnedSlots are the carried properties of each part fn, declared in the
// package at pkgPath, returns, by place in the return, or nil when no part
// carries any.
func returnedSlots(fn *program.Function, pkgPath string, carriers map[carrierKey]carrier) []map[string]*program.Operand {
	if fn == nil || len(fn.Returns) == 0 {
		return nil
	}
	slots := make([]map[string]*program.Operand, len(fn.Returns))
	carries := false
	for i, r := range fn.Returns {
		slots[i] = returnedProperties(fn, r, pkgPath, carriers)
		carries = carries || len(slots[i]) > 0
	}
	if !carries {
		return nil
	}
	return slots
}

// returnedProperties are the carried properties of one returned operand: a
// bare name, assigned in fn's body, that has them — set on it, or carried by
// the call it was assigned from — or a call to a carrier.
func returnedProperties(fn *program.Function, ret *program.Operand, pkgPath string, carriers map[carrierKey]carrier) map[string]*program.Operand {
	out := map[string]*program.Operand{}
	// carriedOut adds what slot i of the carrier call fc carries, under
	// whatever is already set.
	carriedOut := func(fc *program.FnCall, i int) {
		c, ok := carrierIn(fc, pkgPath, carriers)
		if !ok || i >= len(c.slots) {
			return
		}
		for _, prop := range c.slots[i] {
			if _, set := out[prop]; !set {
				out[prop] = companionCall(fc, prop, i, c.multi)
			}
		}
	}
	if fc, ok := ret.T.(*program.Operand_FnCall); ok {
		carriedOut(fc.FnCall, 0)
		return nonEmpty(out)
	}
	// The last assignment to the name is the part returned.
	last, place := returnedLocal(fn, ret)
	if last == nil {
		return nil
	}
	shape := false
	for _, p := range last.Variables[place].Properties {
		for _, want := range carriedProperties {
			if p.Name == want {
				out[want] = p.Expression
			}
		}
		for _, only := range shapeProperties {
			shape = shape || p.Name == only
		}
	}
	if fc, ok := last.Expression.T.(*program.Operand_FnCall); ok {
		carriedOut(fc.FnCall, place)
		if c, ok := carrierIn(fc.FnCall, pkgPath, carriers); ok && place < len(c.slots) {
			shape = shape || len(c.slots[place]) > 0
		}
	}
	// What a carrier's part carries makes it a shape: a colour carries nothing.
	if !shape {
		delete(out, simplePropName)
	}
	return nonEmpty(out)
}

// nonEmpty is m, or nil when it holds nothing.
func nonEmpty(m map[string]*program.Operand) map[string]*program.Operand {
	if len(m) == 0 {
		return nil
	}
	return m
}

// carrierIn is what the overload fc calls carries, fc being written in the
// package at pkgPath: an unqualified call there names that package's own
// function.
func carrierIn(fc *program.FnCall, pkgPath string, carriers map[carrierKey]carrier) (carrier, bool) {
	if fc == nil || fc.Reference == nil {
		return carrier{}, false
	}
	path := fc.Reference.PkgPath
	if path == "" && fc.Reference.Prefix == "" {
		path = pkgPath
	}
	c, ok := carriers[carrierKey{pkgPath: path, name: fc.Reference.Name, params: len(fc.Arguments)}]
	return c, ok
}

// companionOf is fn with its return replaced by expr: the same parameters and
// body, so expr reads the same locals, with the properties taken off the body
// so the companion carries nothing of its own. A stand-in's companion, keep,
// leaves them on, so the stand-in it returns carries its own colour and
// material out in turn.
func companionOf(fn *program.Function, expr *program.Operand, keep bool) *program.Function {
	c := &program.Function{}
	for _, p := range fn.Parameters {
		c.Parameters = append(c.Parameters, proto.Clone(p).(*program.Parameter))
	}
	for _, a := range fn.Assignments {
		ca := proto.Clone(a).(*program.Assignment)
		if !keep {
			for _, v := range ca.Variables {
				v.Properties = nil
			}
		}
		c.Assignments = append(c.Assignments, ca)
	}
	for _, ft := range fn.FnTypes {
		c.FnTypes = append(c.FnTypes, proto.Clone(ft).(*program.FnType))
	}
	c.Returns = []*program.Operand{proto.Clone(expr).(*program.Operand)}
	return c
}

// carryCalls gives every call to a carrier in pkg's declarations and returns
// the carried properties, hoisting a nested call into a declaration of its own.
func carryCalls(pkg *program.Package, carriers map[carrierKey]carrier) {
	taken := map[string]bool{}
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, a := range f.Assignments {
			for _, v := range a.Variables {
				taken[v.Name] = true
			}
		}
	}
	n := 0
	fresh := func() string {
		for {
			n++
			name := fmt.Sprintf("gslCarry%d", n)
			if !taken[name] {
				taken[name] = true
				return name
			}
		}
	}

	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		var out []*program.Assignment
		// carried declares the value of prop for the call fc and returns the
		// property that reads it.
		carried := func(at *program.Assignment, fc *program.FnCall, prop string, i int, multi bool) *program.Property {
			name := fresh()
			out = append(out, &program.Assignment{
				Start:      at.Start,
				End:        at.End,
				Variables:  []*program.Variable{{Name: name}},
				Expression: companionCall(fc, prop, i, multi),
			})
			return &program.Property{
				Start:      at.Start,
				End:        at.End,
				Name:       prop,
				Expression: &program.Operand{T: &program.Operand_Reference{Reference: &program.Reference{Name: name}}},
			}
		}
		// hoist declares a carrier call nested in at's expression as a value of
		// its own, with what it carries but the properties in skip.
		hoist := func(at *program.Assignment, skip map[string]bool) func(*program.FnCall, []string) *program.Operand {
			return func(fc *program.FnCall, props []string) *program.Operand {
				name := fresh()
				v := &program.Variable{Name: name}
				for _, prop := range props {
					if !skip[prop] {
						v.Properties = append(v.Properties, carried(at, fc, prop, 0, false))
					}
				}
				out = append(out, &program.Assignment{
					Start:      at.Start,
					End:        at.End,
					Variables:  []*program.Variable{v},
					Expression: &program.Operand{T: &program.Operand_FnCall{FnCall: fc}},
				})
				return &program.Operand{T: &program.Operand_Reference{Reference: &program.Reference{Name: name}}}
			}
		}
		// standIn gives v, declared by at, the stand-in stand, its carrier
		// calls hoisted with what they carry but what the document sets on v
		// itself, own, so a colour it gives the part is its stand-in's too.
		standIn := func(at *program.Assignment, v *program.Variable, own map[string]bool, stand *program.Operand) {
			stand = hoistCalls(stand, carriers, hoist(at, own))
			v.Properties = append(v.Properties, &program.Property{
				Start:      at.Start,
				End:        at.End,
				Name:       simplePropName,
				Expression: call(carriedSimpleName, stand),
			})
		}
		for _, a := range f.Assignments {
			if a == nil {
				out = append(out, a)
				continue
			}
			// A declaration that IS the call keeps its names: the properties go
			// straight onto them, each the part at its place in the return,
			// unless the caller set them there already.
			if fc, ok := a.Expression.T.(*program.Operand_FnCall); ok {
				if c, ok := carrierOf(fc.FnCall, carriers); ok && len(a.Variables) == len(c.slots) {
					for i, arg := range fc.FnCall.Arguments {
						fc.FnCall.Arguments[i] = hoistCalls(arg, carriers, hoist(a, nil))
					}
					for i, v := range a.Variables {
						own := propertyNames(v)
						for _, prop := range c.slots[i] {
							if own[prop] {
								continue
							}
							if prop == simplePropName {
								standIn(a, v, own, companionCall(fc.FnCall, prop, i, c.multi))
								continue
							}
							v.Properties = append(v.Properties, carried(a, fc.FnCall, prop, i, c.multi))
						}
					}
					out = append(out, a)
					continue
				}
			}
			// A part placed by an expression, place(screw(5.0), at), has as its
			// stand-in the same expression with the stand-ins put in.
			if len(a.Variables) == 1 && a.Variables[0] != nil && !hasProperty(a.Variables[0], simplePropName) {
				if stand, ok := withStandIns(a.Expression, carriers); ok {
					standIn(a, a.Variables[0], propertyNames(a.Variables[0]), stand)
				}
			}
			a.Expression = hoistCalls(a.Expression, carriers, hoist(a, nil))
			out = append(out, a)
		}
		ret := &program.Assignment{}
		if len(f.Assignments) > 0 {
			ret = f.Assignments[len(f.Assignments)-1]
		}
		for i, r := range f.Returns {
			f.Returns[i] = hoistCalls(r, carriers, hoist(ret, nil))
		}
		f.Assignments = out
	}
}

// withStandIns is a copy of op with every call beneath it that carries a
// stand-in replaced by a call of its stand-in's companion, and whether there
// was one. Calls inside other calls' arguments are reached first, so a call
// that carries one inside another that does is the inner's stand-in in the
// outer's.
func withStandIns(op *program.Operand, carriers map[carrierKey]carrier) (*program.Operand, bool) {
	found := false
	var visit func(op *program.Operand) *program.Operand
	visit = func(op *program.Operand) *program.Operand {
		switch t := op.T.(type) {
		case *program.Operand_Expression:
			if t.Expression != nil {
				for i, o := range t.Expression.Operands {
					t.Expression.Operands[i] = visit(o)
				}
			}
		case *program.Operand_FnCall:
			fc := t.FnCall
			if fc == nil {
				return op
			}
			for i, a := range fc.Arguments {
				fc.Arguments[i] = visit(a)
			}
			if c, ok := carrierOf(fc, carriers); ok && !c.multi && hasSlot(c.slots[0], simplePropName) {
				found = true
				return companionCall(fc, simplePropName, 0, false)
			}
		}
		return op
	}
	out := visit(proto.Clone(op).(*program.Operand))
	return out, found
}

// hasSlot reports whether props holds prop.
func hasSlot(props []string, prop string) bool {
	for _, p := range props {
		if p == prop {
			return true
		}
	}
	return false
}

// hoistCalls replaces every carrier call beneath op, innermost first, with the
// reference hoist returns for it. Calls inside other calls' arguments are
// reached; function bodies are not.
func hoistCalls(op *program.Operand, carriers map[carrierKey]carrier,
	hoist func(*program.FnCall, []string) *program.Operand) *program.Operand {
	if op == nil {
		return nil
	}
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		if t.Expression != nil {
			for i, o := range t.Expression.Operands {
				t.Expression.Operands[i] = hoistCalls(o, carriers, hoist)
			}
		}
	case *program.Operand_FnCall:
		fc := t.FnCall
		if fc == nil {
			return op
		}
		for i, a := range fc.Arguments {
			fc.Arguments[i] = hoistCalls(a, carriers, hoist)
		}
		if props := carriedBy(fc, carriers); len(props) > 0 {
			return hoist(fc, props)
		}
	}
	return op
}

// carrierOf is what the called overload carries, if anything.
func carrierOf(fc *program.FnCall, carriers map[carrierKey]carrier) (carrier, bool) {
	if fc == nil || fc.Reference == nil {
		return carrier{}, false
	}
	c, ok := carriers[carrierKey{pkgPath: fc.Reference.PkgPath, name: fc.Reference.Name, params: len(fc.Arguments)}]
	return c, ok
}

// carriedBy is the properties a single-return overload carries, if any, but
// its stand-in, which belongs to the declaration the call is placed by (see
// withStandIns). A multi-return call is never nested in an expression, so it
// has none here.
func carriedBy(fc *program.FnCall, carriers map[carrierKey]carrier) []string {
	c, ok := carrierOf(fc, carriers)
	if !ok || c.multi {
		return nil
	}
	var props []string
	for _, p := range c.slots[0] {
		if p != simplePropName {
			props = append(props, p)
		}
	}
	return props
}

// companionCall is a call of fc's companion for prop, of the part at place i
// of a multi-return, with fc's arguments.
func companionCall(fc *program.FnCall, prop string, i int, multi bool) *program.Operand {
	args := make([]*program.Operand, len(fc.Arguments))
	for i, a := range fc.Arguments {
		args[i] = proto.Clone(a).(*program.Operand)
	}
	return &program.Operand{T: &program.Operand_FnCall{FnCall: &program.FnCall{
		Reference: &program.Reference{
			Prefix:  fc.Reference.Prefix,
			Name:    companionFor(fc.Reference.Name, prop, i, multi),
			PkgPath: fc.Reference.PkgPath,
		},
		Arguments: args,
	}}}
}

// propertyNames is the set of properties v sets itself.
func propertyNames(v *program.Variable) map[string]bool {
	names := map[string]bool{}
	for _, p := range v.Properties {
		if p != nil {
			names[p.Name] = true
		}
	}
	return names
}

// hasProperty reports whether v sets prop itself.
func hasProperty(v *program.Variable, prop string) bool {
	for _, p := range v.Properties {
		if p.Name == prop {
			return true
		}
	}
	return false
}

// eraseFrames takes the frame declarations (.u, .v) off every function in pkg,
// for a program that reads none. A function names the part it returns only so
// it can declare the part's frame — form = E; form.u = ...; return form; — so
// where a returned local carried nothing but its frame, the return is put back
// to the expression itself. That is the whole of the cost a declaration adds,
// and it goes with the declaration: the function compiles as if it had been
// written return E. A local that keeps another property, is read anywhere else
// in the function, or was named without a frame, is left as written.
func eraseFrames(pkg *program.Package) {
	if pkg == nil {
		return
	}
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, ft := range f.FnTypes {
			eraseFnTypeFrames(ft)
		}
	}
}

func eraseFnTypeFrames(ft *program.FnType) {
	if ft == nil {
		return
	}
	for _, g := range ft.Groups {
		if g == nil {
			continue
		}
		for _, fn := range g.Functions {
			if fn == nil {
				continue
			}
			eraseFunctionFrames(fn)
			for _, nested := range fn.FnTypes {
				eraseFnTypeFrames(nested)
			}
		}
	}
}

func eraseFunctionFrames(fn *program.Function) {
	framed := map[string]bool{}
	for _, a := range fn.Assignments {
		if a == nil {
			continue
		}
		for _, v := range a.Variables {
			if v == nil {
				continue
			}
			kept := v.Properties[:0]
			for _, p := range v.Properties {
				if p != nil && (p.Name == uProperty || p.Name == vProperty) {
					framed[v.Name] = true
					continue
				}
				kept = append(kept, p)
			}
			v.Properties = kept
		}
	}
	if len(fn.Returns) != 1 {
		return
	}
	ret, ok := fn.Returns[0].T.(*program.Operand_Reference)
	if !ok || ret.Reference == nil || ret.Reference.Prefix != "" || !framed[ret.Reference.Name] {
		return
	}
	name := ret.Reference.Name
	at := -1
	for i, a := range fn.Assignments {
		if a == nil || len(a.Variables) != 1 || a.Variables[0].Name != name {
			continue
		}
		if at >= 0 || len(a.Variables[0].Properties) > 0 {
			return // assigned twice, or still carrying something
		}
		at = i
	}
	if at < 0 {
		return
	}
	for i, a := range fn.Assignments {
		if i != at && a != nil && assignmentReads(a, name) {
			return
		}
	}
	for _, nested := range fn.FnTypes {
		if fnTypeReads(nested, name) {
			return
		}
	}
	fn.Returns[0] = fn.Assignments[at].Expression
	fn.Assignments = append(fn.Assignments[:at], fn.Assignments[at+1:]...)
}

// assignmentReads reports whether a's expression or properties refer to name.
func assignmentReads(a *program.Assignment, name string) bool {
	if operandReads(a.Expression, name) {
		return true
	}
	for _, v := range a.Variables {
		if v == nil {
			continue
		}
		for _, p := range v.Properties {
			if p != nil && operandReads(p.Expression, name) {
				return true
			}
		}
	}
	return false
}

// fnTypeReads reports whether any function of ft refers to name, as a nested
// function may read its enclosing function's locals.
func fnTypeReads(ft *program.FnType, name string) bool {
	if ft == nil {
		return false
	}
	for _, g := range ft.Groups {
		if g == nil {
			continue
		}
		for _, fn := range g.Functions {
			if fn == nil {
				continue
			}
			for _, a := range fn.Assignments {
				if a != nil && assignmentReads(a, name) {
					return true
				}
			}
			for _, r := range fn.Returns {
				if operandReads(r, name) {
					return true
				}
			}
			for _, nested := range fn.FnTypes {
				if fnTypeReads(nested, name) {
					return true
				}
			}
		}
	}
	return false
}

// operandReads reports whether op refers to the unqualified name.
func operandReads(op *program.Operand, name string) bool {
	if op == nil {
		return false
	}
	switch t := op.T.(type) {
	case *program.Operand_Reference:
		return t.Reference != nil && t.Reference.Prefix == "" && t.Reference.Name == name
	case *program.Operand_PropRef:
		return t.PropRef != nil && t.PropRef.Reference != nil && t.PropRef.Reference.Prefix == "" && t.PropRef.Reference.Name == name
	case *program.Operand_Expression:
		for _, sub := range t.Expression.Operands {
			if operandReads(sub, name) {
				return true
			}
		}
	case *program.Operand_FnCall:
		if t.FnCall.Reference != nil && t.FnCall.Reference.Prefix == "" && t.FnCall.Reference.Name == name {
			return true
		}
		for _, a := range t.FnCall.Arguments {
			if operandReads(a, name) {
				return true
			}
		}
	case *program.Operand_Value:
	default:
		panic(fmt.Sprintf("internal error: unhandled operand type %T in operandReads; please add a new case to this switch", t))
	}
	return false
}
