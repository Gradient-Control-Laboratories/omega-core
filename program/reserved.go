package program

const NoiseHash3 = "noiseHash3"

// Cellular (Worley) noise as single oracles rather than an unrolled
// 27-neighbour scan.
//
// One scan knows BOTH closest feature points, so it knows both gradients:
// WorleyFeature and WorleyFarFeature are the points WorleyF1 and WorleyF2
// measure to. Each is locally constant, so each carries its distance's
// derivative the way length() carries its own: d|v| = dot(v, v')/|v|. Neither
// gradient needs a hash the scan has not already done.
const WorleyF1 = "worleyF1"
const WorleyF2 = "worleyF2"
const WorleyFeature = "worleyFeature"
const WorleyFarFeature = "worleyFarFeature"
const WorleyNearCell = "worleyNearCell"
const WorleyFarCell = "worleyFarCell"

// The exact distance to the nearest Voronoi wall (iq's two-pass bisector scan;
// https://iquilezles.org/articles/voronoilines/), and the feature point whose
// wall it is. WorleyWallFeature carries WorleyWall's derivative together with
// WorleyFeature: the wall's normal is the unit vector from one to the other.
const WorleyWall = "worleyWall"
const WorleyWallFeature = "worleyWallFeature"

// The Gabor spinodoid (hodge-release spinodoid.py, spinodoid_gabor) as one
// oracle over a kernel table baked in Go (common/noise/spinodoid), and its
// analytic gradient, which carries the field's derivative the way
// WorleyFeature carries WorleyF1's.
const GaborSpinodoid = "gaborSpinodoid"
const GaborSpinodoidGradient = "gaborSpinodoidGradient"

// The lattice spinodoid: the same field family and the same arguments, with
// the kernels on a lattice and their envelopes tapered, which makes it cheap
// enough to ray-march. It is a different realization than hodge's, so a part
// built from it is not one hodge can rebuild — which is the whole difference
// between the two. Every preview renders this one whichever the document
// names; naming it here asks for it in the bake as well.
const FastGaborSpinodoid = "fastGaborSpinodoid"
const FastGaborSpinodoidGradient = "fastGaborSpinodoidGradient"

// GaborSpinodoidArgs is every spinodoid builtin's argument count: q, period,
// theta, poles, packetWidth, impulses, pool, seed, bias, localNormalize,
// scale, orientation.
const GaborSpinodoidArgs = 12

// IsGaborSpinodoid reports whether name is one of the four spinodoid builtins,
// and FastSpinodoid whether it is one of the lattice pair.
func IsGaborSpinodoid(name string) bool {
	switch name {
	case GaborSpinodoid, GaborSpinodoidGradient, FastGaborSpinodoid, FastGaborSpinodoidGradient:
		return true
	}
	return false
}

// IsFastGaborSpinodoid reports whether name asks for the lattice field.
func IsFastGaborSpinodoid(name string) bool {
	return name == FastGaborSpinodoid || name == FastGaborSpinodoidGradient
}

// IsGaborSpinodoidGradient reports whether name is one of the gradient
// builtins, which return the field's gradient rather than its value.
func IsGaborSpinodoidGradient(name string) bool {
	return name == GaborSpinodoidGradient || name == FastGaborSpinodoidGradient
}

const PixelSize = "pixelSize"

const DisplacementDrawn = "displacementDrawn"

func IsReservedWord(name string) bool {
	return reservedMap[name]
}

func ReservedWords() []string {
	var words []string
	for word := range reservedMap {
		words = append(words, word)
	}
	return words
}

var reservedMap = map[string]bool{
	// --- Primitives ---
	"add":     true,
	"sub":     true,
	"mul":     true,
	"div":     true,
	"modulo":  true,
	"min":     true,
	"max":     true,
	"and":     true,
	"or":      true,
	"not":     true,
	"square":  true,
	"sqrt":    true,
	"floor":   true,
	"ceil":    true,
	"round":   true,
	"neg":     true,
	"abs":     true,
	"exp":     true,
	"ln":      true,
	"pow":     true,
	"sin":     true,
	"cos":     true,
	"tan":     true,
	"asin":    true,
	"acos":    true,
	"atan":    true,
	"atan2":   true,
	"gt":      true,
	"ge":      true,
	"eq":      true,
	"ne":      true,
	"le":      true,
	"lt":      true,
	"compare": true,
	"remap":   true,
	"collect": true,
	"map":     true,
	"reduce":  true,
	"access":  true,
	"select":  true,

	"audio": true, // marker for audio output

	"auto": true, // marker for box value default

	// --- Built-in GSL Functions ---
	"acosh":     true,
	"asinh":     true,
	"atanh":     true,
	"clamp":     true,
	"cosh":      true,
	"cross":     true,
	"degrees":   true,
	"distance":  true,
	"dot":       true,
	"erfinv":    true,
	"exp2":      true,
	"fract":     true,
	"length":    true,
	"normalize": true,
	"radians":   true,
	"sign":      true,
	"sinh":      true,
	"tanh":      true,
	"trunc":     true,

	// --- Built-in Noise Functions ---
	NoiseHash3:                 true,
	WorleyF1:                   true,
	WorleyF2:                   true,
	WorleyFeature:              true,
	WorleyFarFeature:           true,
	WorleyNearCell:             true,
	WorleyFarCell:              true,
	WorleyWall:                 true,
	WorleyWallFeature:          true,
	GaborSpinodoid:             true,
	GaborSpinodoidGradient:     true,
	FastGaborSpinodoid:         true,
	FastGaborSpinodoidGradient: true,
	"warpField":                true,
	PixelSize:                  true,

	// --- Built-in Displacement Functions ---
	DisplacementDrawn: true,

	// --- Built-in UI Functions ---
	"uiPoint":     true,
	"uiTransform": true,
	"uiSpline":    true,
	"uiColor":     true,
	"uiDirection": true,

	// --- Built-in Mesh Functions ---
	"importMesh":     true,
	"importStep":     true,
	"importWarp":     true,
	"mesh":           true,
	"bvhSdf":         true,
	"vdbSdf":         true,
	"meshWarped":     true,
	"vdbSdfWarped":   true,
	"meshUnion":      true,
	"meshIntersect":  true,
	"meshDifference": true,

	// --- Built-in Topology-Reference Functions ---
	"facePoint":          true,
	"facePointNormal":    true,
	"edgePoint":          true,
	"edgePointNormal":    true,
	"vertexPoint":        true,
	"vertexPointNormal":  true,
	"circleCenter":       true,
	"circleCenterNormal": true,

	// --- Built-in Mesh-Property Functions ---
	"meshVolume":         true,
	"meshArea":           true,
	"meshMaxThickness":   true,
	"meshThickestPoint":  true,
	"meshThickestPoints": true,

	// --- Future GSL ---
	"oneof":    true,
	"closest":  true,
	"furthest": true,

	"open":  true,
	"close": true,

	"shape": true,

	// --- GSL visibility ---
	"public":  true,
	"private": true,

	// --- Declarations and Definitions ---
	"var":     true, // (Dart/Go) Declares a variable.
	"final":   true, // (Dart) Declares a single-assignment variable.
	"const":   true, // (Dart/Go) Declares a constant.
	"type":    true, // (Go) Declares a new type.
	"func":    true, // (Go) Declares a function.
	"typedef": true, // (Dart) Creates a type alias.
	"enum":    true, // (Dart) Declares an enumerated type.

	// --- Class, Object-Oriented, and Structural Programming ---
	"class":      true, // (Dart) Declares a class.
	"struct":     true, // (Go) Declares a structure type.
	"interface":  true, // (Dart/Go) Declares an interface.
	"abstract":   true, // (Dart) Declares a class that cannot be instantiated.
	"extends":    true, // (Dart) Indicates class inheritance.
	"implements": true, // (Dart) Specifies implementation of an interface.
	"mixin":      true, // (Dart) Allows code reuse across hierarchies.
	"with":       true, // (Dart) Applies mixins to a class.
	"new":        true, // (Dart) Creates an instance (often optional).
	"this":       true, // (Dart) Refers to the current instance.
	"super":      true, // (Dart) Refers to the superclass.
	"factory":    true, // (Dart) Declares a factory constructor.
	"operator":   true, // (Dart) Declares an operator overload.
	"get":        true, // (Dart) Declares a getter.
	"set":        true, // (Dart) Declares a setter.
	"static":     true, // (Dart) Declares a class-level member.
	"covariant":  true, // (Dart) Relaxes type checking.
	"extension":  true, // (Dart) Adds functionality to existing types.

	// --- Control Flow ---
	"if":          true, // (Dart/Go) Conditional execution.
	"else":        true, // (Dart/Go) Alternative execution path.
	"switch":      true, // (Dart/Go) Multi-way conditional.
	"case":        true, // (Dart/Go) A branch within a switch.
	"default":     true, // (Dart/Go) Default branch for switch/select.
	"for":         true, // (Dart/Go) Looping construct.
	"while":       true, // (Dart) Loop based on a condition at the start.
	"do":          true, // (Dart) Loop executes at least once.
	"break":       true, // (Dart/Go) Exits a loop or switch.
	"continue":    true, // (Dart/Go) Skips to the next loop iteration.
	"return":      true, // (Dart/Go) Exits a function.
	"goto":        true, // (Go) Transfers control (use sparingly).
	"fallthrough": true, // (Go) Executes the next case in a switch.

	// --- Asynchronous Operations and Concurrency ---
	"async":  true, // (Dart) Declares an async function.
	"await":  true, // (Dart) Pauses execution for a Future.
	"yield":  true, // (Dart) Emits values in generators.
	"yield*": true, // (Dart) Delegates to another generator.
	"go":     true, // (Go) Starts a new goroutine.
	"chan":   true, // (Go) Declares a channel type.
	"defer":  true, // (Go) Schedules a function call on exit.

	// --- Exception and Error Handling ---
	"try":     true, // (Dart) Defines a block to test for errors.
	"catch":   true, // (Dart) Defines a block to handle errors.
	"on":      true, // (Dart) Catches specific exception types.
	"throw":   true, // (Dart) Throws an exception.
	"finally": true, // (Dart) Executes code regardless of errors.
	"rethrow": true, // (Dart) Rethrows a caught exception.

	// --- Packages, Libraries, and Visibility ---
	"package":  true, // (Go) Specifies the package.
	"import":   true, // (Dart/Go) Imports libraries/packages.
	"export":   true, // (Dart) Exports a library.
	"library":  true, // (Dart) Declares a library.
	"part":     true, // (Dart) Declares part of a library.
	"show":     true, // (Dart) Specifies members to import.
	"hide":     true, // (Dart) Specifies members to hide on import.
	"deferred": true, // (Dart) Allows lazy loading.

	// --- Types and Type Operations ---
	"dynamic": true, // (Dart) Indicates a dynamic type.
	"void":    true, // (Dart) Indicates no return value.
	"true":    true, // (Dart) Boolean true value.
	"false":   true, // (Dart) Boolean false value.
	"null":    true, // (Dart) Represents null.
	"is":      true, // (Dart) Checks an object's type.
	"as":      true, // (Dart) Performs a type cast.
	"in":      true, // (Dart) Used in for-in loops.
	"range":   true, // (Go) Used in for loops to iterate.

	// --- Other/Miscellaneous ---
	"assert":   true, // (Dart) Checks a condition for debugging.
	"external": true, // (Dart) Marks an externally implemented function.

	// --- Python Keywords (Not in Dart/Go) ---
	"def":      true, // (Python) Defines a function.
	"del":      true, // (Python) Deletes a reference.
	"elif":     true, // (Python) Else-if condition.
	"except":   true, // (Python) Catches exceptions.
	"from":     true, // (Python) Specific import from a module.
	"global":   true, // (Python) Declares a global variable.
	"lambda":   true, // (Python) Creates an anonymous function.
	"match":    true, // (Python) Structural pattern matching.
	"nonlocal": true, // (Python) Declares a non-local variable.
	"None":     true, // (Python) Represents absence of value.
	"pass":     true, // (Python) Null operation.
	"raise":    true, // (Python) Raises an exception.

	// --- C# Keywords (Not in Dart/Go) ---
	"bool":       true, // (C#) Boolean data type.
	"byte":       true, // (C#) 8-bit unsigned integer.
	"checked":    true, // (C#) Enables overflow checking.
	"decimal":    true, // (C#) 128-bit decimal type.
	"delegate":   true, // (C#) Declares a method reference type.
	"event":      true, // (C#) Declares an event.
	"explicit":   true, // (C#) Declares explicit type conversion.
	"fixed":      true, // (C#) Pins a variable (unsafe code).
	"foreach":    true, // (C#) Iterates over a collection.
	"implicit":   true, // (C#) Declares implicit type conversion.
	"internal":   true, // (C#) Access modifier (same assembly).
	"lock":       true, // (C#) Acquires a mutual-exclusion lock.
	"object":     true, // (C#) Base type for all other types.
	"out":        true, // (C#) Passes argument by reference (must be assigned).
	"override":   true, // (C#) Modifies an inherited method.
	"params":     true, // (C#) Variable number of arguments.
	"protected":  true, // (C#) Access modifier (class and derived).
	"readonly":   true, // (C#) Field set only in constructor.
	"ref":        true, // (C#) Passes argument by reference.
	"sbyte":      true, // (C#) 8-bit signed integer.
	"sealed":     true, // (C#) Prevents class inheritance.
	"short":      true, // (C#) 16-bit signed integer.
	"sizeof":     true, // (C#) Gets the size of a type.
	"stackalloc": true, // (C#) Allocates on the stack.
	"string":     true, // (C#) String data type.
	"typeof":     true, // (C#) Gets the Type object.
	"uint":       true, // (C#) 32-bit unsigned integer.
	"ulong":      true, // (C#) 64-bit unsigned integer.
	"unchecked":  true, // (C#) Disables overflow checking.
	"unsafe":     true, // (C#) Denotes unsafe code.
	"ushort":     true, // (C#) 16-bit unsigned integer.
	"using":      true, // (C#) Imports namespace / resource disposal.
	"virtual":    true, // (C#) Declares an overridable method.

	// --- C++ Keywords (Not in Dart/Go) ---
	"asm": true, // (C++) Inserts inline assembly.
	// "bool" is already listed from C#
	"constexpr":    true, // (C++) Compile-time constant expression.
	"decltype":     true, // (C++) Inspects declared type.
	"delete":       true, // (C++) Deallocates memory.
	"dynamic_cast": true, // (C++) Runtime type cast.
	// "explicit" is already listed from C#
	"friend":    true, // (C++) Grants access to private members.
	"mutable":   true, // (C++) Allows member modification in const method.
	"namespace": true, // (C++) Defines a declarative region.
	"nullptr":   true, // (C++) Null pointer literal.
	// "override" is already listed from C#
	// "private" is already listed from C#
	// "protected" is already listed from C#
	// "public" is already listed from C#
	"reinterpret_cast": true, // (C++) Low-level type cast.
	"static_cast":      true, // (C++) Type cast.
	"template":         true, // (C++) Declares a template.
	"typename":         true, // (C++) Refers to a type in a template.
	// "using" is already listed from C#
	// "virtual" is already listed from C#
	"wchar_t": true, // (C++) Wide character type.

	// --- C Keywords (Not in Dart/Go) ---
	"char":     true, // (C) Character data type.
	"double":   true, // (C) Double-precision float.
	"extern":   true, // (C) Declares identifier defined elsewhere.
	"float":    true, // (C) Single-precision float.
	"int":      true, // (C) Integer data type.
	"long":     true, // (C) Long integer.
	"register": true, // (C) Requests storage in register.
	// "short" is already listed from C#
	"signed": true, // (C) Signed integer modifier.
	// "sizeof" is already listed from C#
	"union":    true, // (C) Type with shared memory for members.
	"unsigned": true, // (C) Unsigned integer modifier.
	"volatile": true, // (C) Indicates variable may be modified externally.

	// --- Haskell Keywords (Not in Dart/Go) ---
	"data":     true, // (Haskell) Defines an algebraic data type.
	"deriving": true, // (Haskell) Automatic derivation of typeclass instances.
	"foreign":  true, // (Haskell) Foreign Function Interface (FFI).
	"infix":    true, // (Haskell) Specifies operator fixity.
	"infixl":   true, // (Haskell) Specifies operator fixity (left).
	"infixr":   true, // (Haskell) Specifies operator fixity (right).
	"instance": true, // (Haskell) Defines a typeclass instance.
	"let":      true, // (Haskell) Binds variables locally.
	"module":   true, // (Haskell) Declares a module.
	"of":       true, // (Haskell) Used in 'case...of' expressions.
	"then":     true, // (Haskell) Used in 'if...then...else'.
	"where":    true, // (Haskell) Defines local bindings.

	// --- Graphics Shading Languages (GLSL / WGSL) ---
	"alias":              true, // (WGSL) Creates a type alias.
	"align":              true, // (WGSL) Attribute for memory alignment.
	"attribute":          true, // (GLSL) Vertex shader input (older).
	"binding":            true, // (WGSL) Attribute for resource binding.
	"bvec2":              true, // (GLSL) 2D boolean vector.
	"bvec3":              true, // (GLSL) 3D boolean vector.
	"bvec4":              true, // (GLSL) 4D boolean vector.
	"buffer":             true, // (GLSL) Shader storage buffer.
	"builtin":            true, // (WGSL) Attribute for built-in variable.
	"centroid":           true, // (GLSL) Interpolation modifier.
	"compute":            true, // (WGSL) Attribute for compute shader stage.
	"continuing":         true, // (WGSL) Block in a loop.
	"diagnostic":         true, // (WGSL) Controls compiler diagnostics.
	"discard":            true, // (GLSL/WGSL) Rejects fragment/pixel.
	"dmat2":              true, // (GLSL) 2x2 double matrix.
	"dmat3":              true, // (GLSL) 3x3 double matrix.
	"dmat4":              true, // (GLSL) 4x4 double matrix.
	"dvec2":              true, // (GLSL) 2D double vector.
	"dvec3":              true, // (GLSL) 3D double vector.
	"dvec4":              true, // (GLSL) 4D double vector.
	"enable":             true, // (WGSL) Enables a language extension.
	"f16":                true, // (WGSL) 16-bit float type.
	"f32":                true, // (WGSL) 32-bit float type.
	"fn":                 true, // (WGSL) Declares a function.
	"fragment":           true, // (WGSL) Attribute for fragment shader stage.
	"group":              true, // (WGSL) Attribute for resource binding group.
	"highp":              true, // (GLSL) High precision qualifier.
	"i32":                true, // (WGSL) 32-bit signed integer type.
	"id":                 true, // (WGSL) Attribute for compute shader invocation ID.
	"image2D":            true, // (GLSL) 2D image for read/write.
	"iimage2D":           true, // (GLSL) 2D signed integer image.
	"inout":              true, // (GLSL) Input/output parameter.
	"interpolate":        true, // (WGSL) Attribute for interpolation.
	"isampler2D":         true, // (GLSL) 2D signed integer sampler.
	"ivec2":              true, // (GLSL) 2D signed integer vector.
	"ivec3":              true, // (GLSL) 3D signed integer vector.
	"ivec4":              true, // (GLSL) 4D signed integer vector.
	"layout":             true, // (GLSL) Specifies memory layout/binding.
	"location":           true, // (WGSL) Attribute for location.
	"loop":               true, // (WGSL) Unconditional loop.
	"lowp":               true, // (GLSL) Low precision qualifier.
	"mat2":               true, // (GLSL) 2x2 float matrix.
	"mat2x3":             true, // (GLSL) 2x3 float matrix.
	"mat2x4":             true, // (GLSL) 2x4 float matrix.
	"mat3":               true, // (GLSL) 3x3 float matrix.
	"mat3x2":             true, // (GLSL) 3x2 float matrix.
	"mat3x4":             true, // (GLSL) 3x4 float matrix.
	"mat4":               true, // (GLSL) 4x4 float matrix.
	"mat4x2":             true, // (GLSL) 4x2 float matrix.
	"mat4x3":             true, // (GLSL) 4x3 float matrix.
	"mediump":            true, // (GLSL) Medium precision qualifier.
	"noperspective":      true, // (GLSL) Disables perspective-correct interpolation.
	"patch":              true, // (GLSL) Used in tessellation shaders.
	"precision":          true, // (GLSL) Specifies default precision.
	"sample":             true, // (GLSL) Per-sample interpolation.
	"sampler1D":          true, // (GLSL) 1D texture sampler.
	"sampler2D":          true, // (GLSL) 2D texture sampler.
	"sampler2DShadow":    true, // (GLSL) 2D depth texture sampler.
	"sampler3D":          true, // (GLSL) 3D texture sampler.
	"samplerCube":        true, // (GLSL) Cube map texture sampler.
	"sampler_comparison": true, // (WGSL) Depth-comparison sampler.
	"shared":             true, // (GLSL) Shared compute workgroup variable.
	"smooth":             true, // (GLSL) Default interpolation.
	"storage":            true, // (WGSL) Storage class for buffers.
	"subroutine":         true, // (GLSL) Shader "function pointer".
	"texture_1d":         true, // (WGSL) 1D texture type.
	"texture_2d":         true, // (WGSL) 2D texture type.
	"texture_3d":         true, // (WGSL) 3D texture type.
	"texture_cube":       true, // (WGSL) Cube map texture type.
	"texture_depth_2d":   true, // (WGSL) 2D depth texture type.
	"texture_storage_2d": true, // (WGSL) 2D storage texture.
	"u32":                true, // (WGSL) 32-bit unsigned integer type.
	"uimage2D":           true, // (GLSL) 2D unsigned integer image.
	"usampler2D":         true, // (GLSL) 2D unsigned integer sampler.
	"uvec2":              true, // (GLSL) 2D unsigned integer vector.
	"uvec3":              true, // (GLSL) 3D unsigned integer vector.
	"uvec4":              true, // (GLSL) 4D unsigned integer vector.
	"varying":            true, // (GLSL) Vertex to fragment variable (older).
	"vec2":               true, // (GLSL) 2D float vector.
	"vec3":               true, // (GLSL) 3D float vector.
	"vec4":               true, // (GLSL) 4D float vector.
	"vertex":             true, // (WGSL) Attribute for vertex shader stage.
	"workgroup":          true, // (WGSL) Storage class for workgroup.
	"workgroup_size":     true, // (WGSL) Attribute for compute workgroup size.

	// (WGSL Future Reserved)
	"atomic":  true, // (WGSL Future)
	"bitcast": true, // (WGSL Future)
	"i8":      true, // (WGSL Future)
	"i16":     true, // (WGSL Future)
	"i64":     true, // (WGSL Future)
	"u8":      true, // (WGSL Future)
	"u16":     true, // (WGLSL Future)
	"u64":     true, // (WGSL Future)
	"mat2x3h": true, // (WGSL Future)
	"mat4x4f": true, // (WGSL Future)
	"vec2i":   true, // (WGSL Future)
	"vec3u":   true, // (WGSL Future)
}
