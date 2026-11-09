package program

// IsPrimitive reports whether name is a recognized primitive GSL operation.
func IsPrimitive(name string) bool {
	return PrimitiveFuncNamed(name) != nil
}

// PrimitiveFuncNamed returns the PrimitiveFunc definition for name, or nil if none exists.
func PrimitiveFuncNamed(name string) *PrimitiveFunc {
	return primitives[name]
}

// PrimitiveFunc defines a built-in primitive language operation and its argument count bounds.
type PrimitiveFunc struct {
	Name    string
	Op      Operation
	MinArgs int
	MaxArgs int
}

var primitives = map[string]*PrimitiveFunc{
	"add": {Name: "add", Op: Operation_ADD, MinArgs: 2, MaxArgs: 100},
	"sub": {Name: "sub", Op: Operation_SUB, MinArgs: 2, MaxArgs: 100},
	"mul": {Name: "mul", Op: Operation_MUL, MinArgs: 2, MaxArgs: 100},
	"div": {Name: "div", Op: Operation_DIV, MinArgs: 2, MaxArgs: 100},
	"mod": {Name: "mod", Op: Operation_MODULO, MinArgs: 2, MaxArgs: 100},

	"min": {Name: "min", Op: Operation_MIN, MinArgs: 2, MaxArgs: 100},
	"max": {Name: "max", Op: Operation_MAX, MinArgs: 2, MaxArgs: 100},

	"and": {Name: "and", Op: Operation_AND, MinArgs: 2, MaxArgs: 100},
	"or":  {Name: "or", Op: Operation_OR, MinArgs: 2, MaxArgs: 100},
	"not": {Name: "not", Op: Operation_NOT, MinArgs: 1, MaxArgs: 1},

	"square": {Name: "square", Op: Operation_SQUARE, MinArgs: 1, MaxArgs: 1},
	"sqrt":   {Name: "sqrt", Op: Operation_SQRT, MinArgs: 1, MaxArgs: 1},

	"floor": {Name: "floor", Op: Operation_FLOOR, MinArgs: 1, MaxArgs: 1},
	"ceil":  {Name: "ceil", Op: Operation_CEIL, MinArgs: 1, MaxArgs: 1},
	"round": {Name: "round", Op: Operation_ROUND, MinArgs: 1, MaxArgs: 1},

	"neg": {Name: "neg", Op: Operation_NEG, MinArgs: 1, MaxArgs: 1},
	"abs": {Name: "abs", Op: Operation_ABS, MinArgs: 1, MaxArgs: 1},
	"exp": {Name: "exp", Op: Operation_EXP, MinArgs: 1, MaxArgs: 1},
	"ln":  {Name: "ln", Op: Operation_LN, MinArgs: 1, MaxArgs: 1},
	"pow": {Name: "pow", Op: Operation_POW, MinArgs: 2, MaxArgs: 2},

	"sin":  {Name: "sin", Op: Operation_SIN, MinArgs: 1, MaxArgs: 1},
	"cos":  {Name: "cos", Op: Operation_COS, MinArgs: 1, MaxArgs: 1},
	"tan":  {Name: "tan", Op: Operation_TAN, MinArgs: 1, MaxArgs: 1},
	"asin": {Name: "asin", Op: Operation_ASIN, MinArgs: 1, MaxArgs: 1},
	"acos": {Name: "acos", Op: Operation_ACOS, MinArgs: 1, MaxArgs: 1},
	"atan": {Name: "atan", Op: Operation_ATAN, MinArgs: 1, MaxArgs: 2},

	"compare": {Name: "compare", Op: Operation_COMPARE, MinArgs: 2, MaxArgs: 2},
	"remap":   {Name: "remap", Op: Operation_REMAP, MinArgs: 2, MaxArgs: 2},
	"collect": {Name: "collect", Op: Operation_COLLECT, MinArgs: 1, MaxArgs: 100},
	"map":     {Name: "map", Op: Operation_MAP, MinArgs: 2, MaxArgs: 100},
	"reduce":  {Name: "reduce", Op: Operation_REDUCE, MinArgs: 2, MaxArgs: 100},
	"access":  {Name: "access", Op: Operation_ACCESS, MinArgs: 2, MaxArgs: 2},
	"select":  {Name: "select", Op: Operation_SELECT, MinArgs: 3, MaxArgs: 4},
}
