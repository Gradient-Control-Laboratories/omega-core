package annotation

import "fmt"

// ValidateTarget checks if the given annotation is permitted on the specified syntactic target.
func ValidateTarget(name string, target Target) error {
	switch name {
	case Range:
		// @range is permitted on variable assignments, parameters, and inline expressions.
		if target == TargetAssignment || target == TargetParameter || target == TargetExpression {
			return nil
		}
		return fmt.Errorf("@%s annotation cannot be applied to %s", name, target)

	case Values:
		// @values is permitted on variable assignments only.
		if target == TargetAssignment {
			return nil
		}
		return fmt.Errorf("@%s annotation can only be applied to variable assignments", name)

	case Point, Transform, Color, Direction:
		// Viewport annotations are permitted on variable assignments and inline
		// expressions, where they give the value a control. On a parameter
		// they declare what kind of argument the function takes, recorded
		// rather than lowered, so a tool writing a call knows which control
		// each argument wants.
		if target == TargetAssignment || target == TargetExpression || target == TargetParameter {
			return nil
		}
		return fmt.Errorf("@%s annotation cannot be applied to %s", name, target)

	case Frame:
		// A frame is the body's own [body.u, body.v], read off the body a
		// call is written for; there is no control to give a variable one.
		if target == TargetParameter {
			return nil
		}
		return fmt.Errorf("@%s annotation can only be applied to function parameters", name)

	case Spline:
		// A spline is a list of transforms with its own overlay; nothing
		// writes one as a single argument.
		if target == TargetAssignment || target == TargetExpression {
			return nil
		}
		if target == TargetParameter {
			return fmt.Errorf("unsupported parameter annotation @%s", name)
		}
		return fmt.Errorf("@%s annotation cannot be applied to %s", name, target)

	case Material, Finish:
		// Recorded annotations mark a browsable symbol rather than wrapping an
		// expression, so a function declaration is as valid a target as an
		// assignment.
		if target == TargetAssignment || target == TargetFunction {
			return nil
		}
		if target == TargetParameter {
			return fmt.Errorf("unsupported parameter annotation @%s", name)
		}
		return fmt.Errorf("@%s annotation cannot be applied to %s", name, target)

	case Surface:
		// An operator takes a surface and returns the modified one, which
		// only a function can do.
		if target == TargetFunction {
			return nil
		}
		if target == TargetParameter {
			return fmt.Errorf("unsupported parameter annotation @%s", name)
		}
		return fmt.Errorf("@%s annotation can only be applied to function declarations", name)

	case Scene:
		return fmt.Errorf("@scene is a top-level package directive, not an element annotation")

	default:
		return fmt.Errorf("unsupported annotation @%s", name)
	}
}

// ValidateArgCount verifies the number of arguments passed to the annotation.
func ValidateArgCount(name string, count int) error {
	switch name {
	case Point, Transform, Spline, Material, Finish, Color, Frame, Surface:
		if count > 0 {
			return fmt.Errorf("@%s annotation does not take arguments", name)
		}
		return nil

	case Direction:
		// A direction has no place of its own; the optional argument is the
		// point its arrow is drawn from, a constant or a @point variable, or
		// on a parameter the sibling parameter that anchors it.
		if count > 1 {
			return fmt.Errorf("@direction annotation takes at most one argument: @direction(anchor)")
		}
		return nil

	case Range:
		if count < 2 || count > 3 {
			return fmt.Errorf("@range annotation requires 2 or 3 arguments: @range(min, max[, period])")
		}
		return nil

	case Values:
		if count < 2 {
			return fmt.Errorf("@values annotation requires at least 2 arguments")
		}
		return nil

	default:
		return fmt.Errorf("unsupported annotation @%s", name)
	}
}
