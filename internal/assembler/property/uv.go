package property

import (
	"fmt"

	"gradientcontrol.com/omega/core/program"
)

// uProperty and vProperty are the declaration properties that author a shape's
// own surface frame: `name.u = …;` and `name.v = …;`.
//
// They differ from color and material in who writes them. A document's author
// writes a colour; the PRIMITIVES write u and v, because a cylinder already
// knows the angle about its axis and the distance along it. A material that
// needs a frame — a turned lay, a wood grain's pith — can then read the frame
// of the shape it is on, instead of being handed one at the call site that has
// to be kept in step with the geometry by hand. Nothing keeps those two in step
// today, and they have already drifted: Material - Coated.gsl turns one
// material about an axis through the origin and puts it on four bars that are
// not there.
//
// Linear coordinates are millimetres and angular ones radians. The angle is the
// ANGLE and never the arc length: multiplying by the radius adds a radial term
// through the product rule, so a direction read from the gradient comes out
// wrong (measured at 0.38 to 1.06 against the exact answer, doc/material_composition.md).
//
// A frame coordinate's VALUE is discontinuous where the parameterisation wraps
// — u jumps by 2 pi across a cylinder's seam — while its GRADIENT is not. A
// material that differentiates u is therefore safe everywhere off the axis; one
// that indexes marks by u's value meets the seam.
const (
	uProperty = "u"
	vProperty = "v"
)

// uvDefault is the frame coordinate of a shape that authors none. Zero is the
// only honest answer: the shape has no frame, and a reader gets a constant —
// whose gradient is zero, so a lay taken from it has no direction — rather than
// a direction invented here that would look deliberate.
func uvDefault() *program.Operand { return num(0) }

// uvUnclamped keeps a blended frame coordinate as it comes. Unlike a colour,
// which is clamped into [0, 1], and a material row, whose channels have
// meanings to bound, a frame coordinate has no range: an angle is already
// bounded by how it is built and a length along a shape is unbounded.
func uvUnclamped(_ int, v *program.Operand) *program.Operand { return v }

var uSpec = fieldSpec{
	name:     uProperty,
	rows:     0,
	fallback: uvDefault,
	clampRow: uvUnclamped,
}

var vSpec = fieldSpec{
	name:     vProperty,
	rows:     0,
	fallback: uvDefault,
	clampRow: uvUnclamped,
}

// UFields is ColorFields for the u coordinate: for each return of pkg, the
// field of p whose value is the shape's u, or nil for a return that references
// no shape carrying one.
//
// It blends over the same ∂D/∂dᵢ weights as colour, which is what makes a frame
// behave sensibly under a boolean: a sharp union hard-selects the nearer
// shape's frame, so at a box's edge the frame flips between faces rather than
// averaging into a diagonal that belongs to neither. Only a smooth or rounded
// combiner blends, and there the two faces' coordinates are kept orthogonal by
// how the primitives choose them, so the average stays a rotation rather than
// collapsing.
func UFields(pkg *program.Package, getImport program.GetImportFunc) ([]*program.Operand, error) {
	return propertyFields(pkg, getImport, uSpec)
}

// VFields is UFields for the v coordinate, resolved independently: a shape may
// carry either, both or neither.
func VFields(pkg *program.Package, getImport program.GetImportFunc) ([]*program.Operand, error) {
	return propertyFields(pkg, getImport, vSpec)
}

// frames expands the frame reads .u and .v (SURFACEU, SURFACEV): name.u is the
// u of the shape name, resolved by the same resolver as its colour, so a read
// at the call site — bar.material = turned(..., bar.u, ...) — hands a material
// the frame of the shape it is on, and nothing has to keep the two in step.
//
// A read resolves on its own target and nothing else, so it never blends one
// shape's frame into another's: a material on bar reads bar's frame, and the
// materials, not the frames, are what a boolean then blends. A read of a
// composite (min(a, b).u) blends its parts' frames over the colour weights,
// as UFields does for a return.
//
// bar.second.u is the frame of the second-closest surface. The read expands
// the target's .second first, so the weights it resolves over are those of the
// second-closest distance — max(a, b) for two shapes — and select the surface
// that is second, not the one that binds. Resolved over bar instead, the read
// would answer for the binding surface while asking for the other one.
// bar.u.second is refused: a frame coordinate has no second-closest surface.
type frames struct {
	pkg       *program.Package
	getImport program.GetImportFunc
	sec       *second   // expands a .second in a read's target
	u, v      *resolver // built on first read
}

func newFrames(pkg *program.Package, getImport program.GetImportFunc, sec *second) *frames {
	return &frames{pkg: pkg, getImport: getImport, sec: sec}
}

// expand returns the expansion of a frame read: operation is SURFACEU or
// SURFACEV and operands its one target. A shape that carries no frame reads as
// uvDefault, whose gradient is zero, so a lay taken from it has no direction.
func (f *frames) expand(operation program.Operation, operands []*program.Operand) (*program.Operand, error) {
	spec, which := uSpec, uProperty
	if operation == program.Operation_SURFACEV {
		spec, which = vSpec, vProperty
	}
	if len(operands) != 1 {
		return nil, fmt.Errorf(".%s expects 1 operand, got %d", which, len(operands))
	}
	if !packageHasProperty(f.pkg, spec.name) {
		return spec.fallback(), nil
	}
	target, err := f.expandSeconds(operands[0])
	if err != nil {
		return nil, err
	}
	r := f.resolverFor(operation, spec)
	field, err := r.fieldOf(r.unwrapRemaps(target))
	if err != nil {
		return nil, fmt.Errorf("cannot resolve the .%s frame: %w", which, err)
	}
	if field == nil {
		return spec.fallback(), nil
	}
	return flattenRemaps(field), nil
}

func (f *frames) resolverFor(operation program.Operation, spec fieldSpec) *resolver {
	if operation == program.Operation_SURFACEV {
		if f.v == nil {
			f.v = newResolver(f.pkg, f.getImport, spec)
		}
		return f.v
	}
	if f.u == nil {
		f.u = newResolver(f.pkg, f.getImport, spec)
	}
	return f.u
}

// expandSeconds replaces each .second in op with the field it expands to, so
// a frame read of it resolves over the second-closest distance.
func (f *frames) expandSeconds(op *program.Operand) (*program.Operand, error) {
	if op == nil || f.sec == nil {
		return op, nil
	}
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		if t.Expression.Operation == program.Operation_SECONDCLOSEST {
			return f.sec.expandSecond(t.Expression.Operands)
		}
		operands := make([]*program.Operand, len(t.Expression.Operands))
		for i, sub := range t.Expression.Operands {
			expanded, err := f.expandSeconds(sub)
			if err != nil {
				return nil, err
			}
			operands[i] = expanded
		}
		return &program.Operand{Start: op.Start, End: op.End, T: &program.Operand_Expression{
			Expression: &program.Expression{Operation: t.Expression.Operation, Operands: operands},
		}}, nil
	case *program.Operand_FnCall:
		args := make([]*program.Operand, len(t.FnCall.Arguments))
		for i, a := range t.FnCall.Arguments {
			expanded, err := f.expandSeconds(a)
			if err != nil {
				return nil, err
			}
			args[i] = expanded
		}
		return &program.Operand{Start: op.Start, End: op.End, T: &program.Operand_FnCall{
			FnCall: &program.FnCall{Reference: t.FnCall.Reference, Arguments: args},
		}}, nil
	}
	return op, nil
}

// packageReadsFrame reports whether anything in pkg reads a frame (.u or .v):
// a declaration, a property, a return, or any function body. A program that
// reads none has its frame declarations erased before they cost anything (see
// eraseFrames).
func packageReadsFrame(pkg *program.Package) bool {
	if pkg == nil {
		return false
	}
	for _, f := range pkg.Files {
		if f == nil {
			continue
		}
		for _, a := range f.Assignments {
			if assignmentReadsFrame(a) {
				return true
			}
		}
		for _, r := range f.Returns {
			if operandReadsFrame(r) {
				return true
			}
		}
		for _, ft := range f.FnTypes {
			if fnTypeReadsFrame(ft) {
				return true
			}
		}
	}
	return false
}

func assignmentReadsFrame(a *program.Assignment) bool {
	if a == nil {
		return false
	}
	if operandReadsFrame(a.Expression) {
		return true
	}
	for _, v := range a.Variables {
		if v == nil {
			continue
		}
		for _, p := range v.Properties {
			if p != nil && operandReadsFrame(p.Expression) {
				return true
			}
		}
	}
	return false
}

func fnTypeReadsFrame(ft *program.FnType) bool {
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
				if assignmentReadsFrame(a) {
					return true
				}
			}
			for _, r := range fn.Returns {
				if operandReadsFrame(r) {
					return true
				}
			}
			for _, nested := range fn.FnTypes {
				if fnTypeReadsFrame(nested) {
					return true
				}
			}
		}
	}
	return false
}

func operandReadsFrame(op *program.Operand) bool {
	if op == nil {
		return false
	}
	switch t := op.T.(type) {
	case *program.Operand_Expression:
		if t.Expression.Operation == program.Operation_SURFACEU || t.Expression.Operation == program.Operation_SURFACEV {
			return true
		}
		for _, sub := range t.Expression.Operands {
			if operandReadsFrame(sub) {
				return true
			}
		}
	case *program.Operand_FnCall:
		for _, a := range t.FnCall.Arguments {
			if operandReadsFrame(a) {
				return true
			}
		}
	case *program.Operand_Value, *program.Operand_Reference, *program.Operand_PropRef:
	default:
		panic(fmt.Sprintf("internal error: unhandled operand type %T in operandReadsFrame; please add a new case to this switch", t))
	}
	return false
}
