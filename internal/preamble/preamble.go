package preamble

import (
	"fmt"
	"sort"
	"strings"

	"gradientcontrol.com/omega/core/program"
)

func PrimitiveAndBuiltinDescriptiveText() string {
	var sb strings.Builder
	sb.WriteString("\n---\n\n## **Primitive and Built-in Functions**\n\n")
	sb.WriteString("The following are the primitive and built-in functions that are available in the GSL (GCL Shape Language).\n")
	sb.WriteString("Call them rather than writing the same code yourself.\n\n")

	// Sort built-ins for stable output
	names := make([]string, 0, len(builtinDescriptions))
	for name := range builtinDescriptions {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		sb.WriteString(fmt.Sprintf("* **`%s`**: %s\n", name, builtinDescriptions[name]))
	}

	// Sort primitives for stable output
	ops := make([]int, 0, len(primitiveDescriptions))
	for op := range primitiveDescriptions {
		ops = append(ops, int(op))
	}
	sort.Ints(ops)
	for _, op := range ops {
		desc := primitiveDescriptions[program.Operation(op)]
		if desc == skipDescription {
			continue
		}
		sb.WriteString(fmt.Sprintf("* **`%s`**: %s\n", program.Operation(op).Name(), desc))
	}

	sb.WriteString(`
---

## **Remap Guideline**

**STRICT CONSTRAINT:** Remap is a powerful function that can be used to translate, rotate, and scale shapes.

The first argument can be any expression that evaluates to a signed distance function.
If you want to pass a function as the first argument, then you must pass a function call instead.
For example,

` + "```gsl" + `
// Given
myShape = calcShape()

// This is allowed
x = remap(myShape, [1, 2, 3]);

// and this is allowed
x = remap(calcShape(), [1, 2, 3]);

// but this is NOT allowed because calcShape is a function, not a function call
x = remap(calcShape, [1, 2, 3]);

// When using an anonymous function as the first argument, 
// this is allowed because it is a function call to an anonymous function
ball = remap(() { return sphere(0.3); }(), p - [0, 0, centerZ]);

// but this is NOT allowed because it is an anonymous function, not a function call
ball = remap(() { return sphere(0.3); }, p - [0, 0, centerZ]);

Be sure to add the parentheses to the function call.
` + "```" + `

The second argument is an expression that evaluates to a vector with x, y, and z components.
The result of evaluating the second argument becomes the p value used when evaluating the signed distance function.

For example, a box centered at the origin can be moved with transform.translate,
which is a remap whose second argument is p minus the offset:

` + "```gsl\ntransform.translate(boxCentered(vector.zero, [10, 20, 30]), [1, 2, 3])\n```" + `

This would translate the box 1 mm in the positive x direction,
2 mm in the positive y direction, and 3 mm in the positive z direction.
transform.rotateX, rotateY and rotateZ turn a shape about an axis through the origin by an angle in radians.

Another example is to assign an expression to a variable and then reuse it to produce identical shapes in different places and orientations:

` + "```gsl\n" + `
baseShape = boxCentered(vector.zero, [10, 20, 30]);

shape1 = transform.translate(baseShape, [1, 2, 3]);
shape2 = transform.translate(baseShape, [-1, 2, 3]);
rotatedShape = transform.rotateZ(baseShape, radians(45));
` + "```" + `

When possible, translate, rotate, and scale shapes with the transform functions, which are remaps.
For example, instead of generating a function that returns a box at a specific location with a specific rotation and scale,
use a library shape centered at the origin, such as boxCentered, and then move it with transform.translate, rotateZ and the like.
Write the remap yourself only when the change is not one of those, such as a warp that depends on the point.
A hand-written remap clears the shape's bounding box, so declare one with name.box = [lo, hi] (see Bounding box).
`)
	return sb.String()
}

func ReservedWordsDescriptiveText() string {
	var sb strings.Builder
	sb.WriteString("\n---\n\n## **Reserved Words**\n\n")
	sb.WriteString("The following are reserved words.\n")
	sb.WriteString("Do not use these names for your own functions or parameters or variables.\n\n")

	words := program.ReservedWords()
	for _, name := range words {
		fmt.Fprintf(&sb, "`%s`, ", name)
	}
	res := strings.TrimSuffix(sb.String(), ", ")
	return res + "\n"
}
