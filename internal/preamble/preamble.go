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

For example, assuming you have a box function
that returns a box centered at the origin with a given width, depth, and height,
then the following remap call would translate the box to a new location:

` + "```gsl\nremap(box(width, depth, height), [p.x - 1, p.y - 2, p.z - 3])\n```" + `

This would translate the box 1 mm in the positive x direction,
2 mm in the positive y direction, and 3 mm in the positive z direction.
Various other expressions using sin and cos could be used as part of the 2nd remap argument to rotate the box.

Another example is to assign an expression to a variable and then use remap on it to produce identical shapes in different orientations:

` + "```gsl\n" + `
baseShape = box(width, depth, height)

shape1 = remap(baseShape, [p.x - 1, p.y - 2, p.z - 3])
shape2 = remap(baseShape, [p.x + 1, p.y - 2, p.z - 3])
rotatedShape = remap(baseShape, [p.x * cos(radians(45)) + p.y * sin(radians(45)), -p.x * sin(radians(45)) + p.y * cos(radians(45)), p.z])
` + "```" + `

When possible remap should be used to translate, rotate, and scale shapes.
For example, instead of generating a function that returns a box at a specific location with a specific rotation and scale,
generate a simple box function that defines a box centered at the origin with a given width, depth, and height,
and then use remap to translate, rotate, and scale the box as needed.
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
