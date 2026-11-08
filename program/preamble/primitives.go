package preamble

import (
	pb "gradientcontrol.com/omega/core/program"
)

const skipDescription = "skip"

// Keep this in sync with the primitives in the common model
var primitiveDescriptions = map[pb.Operation]string{
	pb.Operation_UNSPECIFIED_OPERAND: skipDescription,

	pb.Operation_ADD:    skipDescription,
	pb.Operation_SUB:    skipDescription,
	pb.Operation_MUL:    skipDescription,
	pb.Operation_DIV:    skipDescription,
	pb.Operation_MODULO: skipDescription,

	pb.Operation_MIN:    "Takes 2 or more arguments and returns the minimum of the arguments (Union).",
	pb.Operation_MAX:    "Takes 2 or more arguments and returns the maximum of the arguments (Intersection).",
	pb.Operation_AND:    "Takes 2 or more arguments and returns the logical AND of the arguments.",
	pb.Operation_OR:     "Takes 2 or more arguments and returns the logical OR of the arguments.",
	pb.Operation_NOT:    "Takes one argument and returns the logical NOT of the argument.",
	pb.Operation_SQUARE: "Takes one argument and returns the square of the argument.",
	pb.Operation_SQRT:   "Takes one argument and returns the square root of the argument.",
	pb.Operation_FLOOR:  "Takes one argument and rounds down to the nearest integer.",
	pb.Operation_CEIL:   "Takes one argument and rounds up to the nearest integer.",
	pb.Operation_ROUND:  "Takes one argument and rounds to the nearest integer.",
	pb.Operation_NEG:    "Takes one argument and returns the negation of the argument.",
	pb.Operation_ABS:    "Takes one argument and returns the absolute value of the argument.",
	pb.Operation_EXP:    "Takes one argument and returns e raised to the power of the argument.",
	pb.Operation_LN:     "Takes one argument and returns the natural logarithm of the argument.",
	pb.Operation_POW:    "Takes two arguments, base and exponent, and returns base raised to the power of exponent.",
	pb.Operation_SIN:    "Takes one argument and returns the sine of the argument.",
	pb.Operation_COS:    "Takes one argument and returns the cosine of the argument.",
	pb.Operation_TAN:    "Takes one argument and returns the tangent of the argument.",
	pb.Operation_ASIN:   "Takes one argument and returns the arc sine of the argument.",
	pb.Operation_ACOS:   "Takes one argument and returns the arc cosine of the argument.",
	pb.Operation_ATAN: "Takes one or two arguments. " +
		"If one argument, then returns the arc tangent of the argument. " +
		"If two arguments, then returns the arc tangent of the first argument divided by the second argument.",

	// Returns the arc tangent of the first argument divided by the second argument.
	// This is the two argument version of atan, but does not appear in the GSL language.
	// The parser generates Operation_ATAN calls when it encounters atan with one argument
	//                  and Operation_ATAN2 calls when it encounters atan with two arguments.
	pb.Operation_ATAN2: skipDescription,

	pb.Operation_GT:      skipDescription,
	pb.Operation_GE:      skipDescription,
	pb.Operation_EQ:      skipDescription,
	pb.Operation_NE:      skipDescription,
	pb.Operation_LE:      skipDescription,
	pb.Operation_LT:      skipDescription,
	pb.Operation_COMPARE: skipDescription,

	pb.Operation_REMAP: "Calls the first argument, a function, with a p value equal to the second argument, a vector with x, y, and z components.",

	// Internal marker for the .dtime / .dp derivative accessors. Expanded by
	// the autodiff pass before lowering and never appears as a source word.
	pb.Operation_DERIV: skipDescription,

	pb.Operation_COLLECT: "Collects the arguments and returns a collection with the arguments as elements: 2 to 4 scalars form a vector, more form a longer list, and vectors of the same length form a list of vectors.",
	pb.Operation_MAP:     "Applies the first argument, a function taking (value, index), to each element of the second argument, a collection, and returns a new collection of the results. A function that returns a vector produces a list of vectors, which can be indexed, mapped, or reduced.",
	pb.Operation_REDUCE:  "Folds the third argument, a collection, with the second argument, a function taking (accumulator, value), starting from the first argument, the initial accumulator, and returns the final accumulator. The accumulator may be a vector.",

	pb.Operation_ACCESS: skipDescription,
	pb.Operation_SELECT: skipDescription,
	pb.Operation_INDEX:  skipDescription,

	// Internal marker for the .second accessor (second-closest distance).
	// Expanded by the self-computing-property pass before lowering and never
	// appears as a source word.
	pb.Operation_SECONDCLOSEST: skipDescription,

	// Internal marker for the .box accessor (bounding box), expanded the same
	// way.
	pb.Operation_BOUNDINGBOX: skipDescription,

	// Internal markers for the .u and .v accessors (surface frame
	// coordinates), expanded the same way.
	pb.Operation_SURFACEU: skipDescription,
	pb.Operation_SURFACEV: skipDescription,
}
