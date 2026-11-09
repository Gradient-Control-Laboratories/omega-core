package annotation

// Known annotation names in GSL.
const (
	Range     = "range"
	Values    = "values"
	Point     = "point"
	Transform = "transform"
	Spline    = "spline"
	Scene     = "scene"

	// Color and Direction are viewport annotations like Point: a vec3 that a
	// control edits, a colour chip and a direction arrow. On a parameter, any
	// of the control kinds declares what the argument is, so a tool writing a
	// call can give each argument its control.
	Color     = "color"
	Direction = "direction"

	// Frame is a parameter kind alone: the body's own frame, `[body.u,
	// body.v]`, which a tool writing the call reads off the body rather than
	// asks for, so there is no control to give a variable one.
	Frame = "frame"

	// Material and Finish are RECORDED rather than lowered: they are noted on
	// the package file and the declaration they mark is built unchanged. That
	// is what lets them mark a function as readily as a constant, which a
	// lowering cannot, and why they need no target builtin.
	Material = "material"
	Finish   = "finish"

	// Surface is recorded too. It marks an operator preset: an overload of a
	// function that takes a surface first and returns the modified one, taking
	// only what a document supplies and baking the operator's other arguments
	// (doc/material_composition.md, "Design: operators as presets").
	Surface = "surface"
)

// Target represents the syntactic element an annotation is attached to.
type Target int

const (
	TargetAssignment Target = iota + 1
	TargetParameter
	TargetExpression

	// TargetFunction is a named function declaration. Only a recorded
	// annotation can mark one; the viewport annotations wrap an expression and
	// have nothing to wrap here.
	TargetFunction
)

// String returns a human-readable representation of the target.
func (t Target) String() string {
	switch t {
	case TargetAssignment:
		return "variable assignments"
	case TargetParameter:
		return "function parameters"
	case TargetExpression:
		return "expressions"
	case TargetFunction:
		return "function declarations"
	default:
		return "unsupported target"
	}
}

// IsSupported returns true if the given identifier is a recognized annotation.
func IsSupported(name string) bool {
	switch name {
	case Range, Values, Point, Transform, Spline, Scene, Material, Finish, Color, Direction, Frame, Surface:
		return true
	default:
		return false
	}
}

// TargetBuiltin returns the internal GSL builtin function name that corresponds to
// the given viewport annotation, or empty string if not a viewport annotation.
func TargetBuiltin(name string) string {
	switch name {
	case Point:
		return "uiPoint"
	case Transform:
		return "uiTransform"
	case Spline:
		return "uiSpline"
	case Color:
		return "uiColor"
	case Direction:
		return "uiDirection"
	default:
		return ""
	}
}

// RequiresElementPromotion returns true if decorating a list literal with this annotation
// should automatically promote child elements by wrapping them in uiTransform.
func RequiresElementPromotion(name string) bool {
	return name == Spline
}
