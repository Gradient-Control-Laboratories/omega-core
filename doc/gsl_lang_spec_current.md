
# GSL v1 Language Specification (current)

This is a language specification for the existing programming language GSL v1 (GCL Shape Language).

I am generating this language specification from existing work
with the goal of writing a new language specification and compiler
along with tools to convert the old languages files to the new language spec.
This spec will continue to be updated as the compiler and tools are being developed.

## Language Goals

### Why the GSL language was created and what problems it was intended to solve:
- GSL is designed as a domain-specific shape language for defining Constructive Solid
    Geometry (CSG) shapes and Signed Distance Functions (SDFs).
- It allows designers and engineers to programmatically describe
    complex, parametric, mathematical geometries.
- By expressing shapes as math functions rather than static boundary representations (B-Reps),
  models remain infinitely scalable, resolution-independent,
  and easily animatable or modifiable in real-time.
- GSL adopts a functional, Single Static Assignment (SSA) paradigm to make shape code highly
    optimized for compilation into parallel execution platforms like GPUs.

### What GSL provides
- Supports basic arithmetic, component-wise list algebra, modulo, logic, and comparison operators.
- Provides first-class named and anonymous functions, allowing closures and function passing.
- Exposes mesh/CAD STEP slot-numbered and expression-based asset imports.
- Resolves variables and slider properties (`.min`, `.max`, `.period`) dynamically
    to drive real-time parameter tweaking.
- Evaluates mathematical primitives and built-in algorithms (like spatial remapping and
    cellular noise) either on CPU or compiles them to GPU shaders.
- Supports multiple returns
- Supports audio rendering outputs
    This is an add-on to the current implementation and not part of the language proper.

### What GSL does NOT provide
- GSL is not a general-purpose programming language:
    - It lacks arbitrary loops (no `for` or `while`)
    - It lacks dynamic branching (no `if`/`else` control flow)
    - It lacks mutable states
    - It does not support recursion (direct self-recursion and mutual recursion are prohibited)
- It does not support booleans natively, but instead uses floating points:
    - `value <= 0.0` is treated as false
    - `value > 0.0` is treated as true

Compilation targets:
- **WGSL**: For WebGPU display in a browser (for high-performance real-time previews).
- **Mesh / Bytecode**: For export to other programs (STL/STEP generation compiled via Rust-based
  Fidget JIT/Manifold kernels).

---

## Sources used when generating a language specification

These sources are listed here to give you some context,
help with understanding the design decisions made in this language implementation
and allow you to verify and document the existing language specification.

The implementation is unified around the whole-program AST model defined in `program`.
The parser creates an AST directly using this program model,
and the assembler links multi-package programs into a complete program structure.
All downstream compiler passes, analyzers, and backends operate directly on this unified model.

Existing implementation:
- **Program model** in `program`:
  Protobuf schema (`program/program.proto`) and Go representations for whole programs (`Program`),
  multi-file packages (`Package`, `PackageFile`), and AST expressions (`Expr`, `Op`, `Operand`),
  with compiler and analysis subsystems across `program`, `assembler`, and `internal`:
  - `internal/types`: Type inference, type checking, and semantic validation.
  - `internal/eval`: Compile-time constant expression evaluation and constant folding.
  - `internal/assembler/property`: Property lowering, companion variables, and forward-mode
    automatic derivatives.
  - `internal/lower`: AST lowering passes for operations and map/reduce constructs.
  - `internal/closure`: Function specialization, closure analysis, and variable capture.
  - `internal/assembler/deadcode`: Dead code elimination and reachable function pruning.
  - `internal/assembler/inline`: Constant evaluation and parameter remapping for function inlining.
  - `internal/scope`: Multi-level lexical scoping for packages, functions, and variables.
  - `assembler/meshslots`: Mesh and STEP CAD import slot management.
  - `program/preamble`: Embedded definitions for primitives, built-ins, and standard library
    metadata.
- **Parser** in `parser`:
  Parses GSL v1 source directly into `program` AST structures.
  Contains:
  - `parser2` (`parser/internal/parser2`): Participle-based parser creating `program.Package` trees.
  - `identifier2` (`parser/internal/identifier2`): Validation for identifiers, keywords, and
    built-ins.
- **Assembler** in `assembler`:
  Resolves package dependencies, links imported references, lowers properties, and constructs
  complete `program.Program` bundles. An interactive WebGPU viewer and editor is located in
  `examples/viewer` (served by `cmd/serve`).
- **Common build infrastructure** in `internal/build`:
  Provides the AST visitor and code generator delegate architecture (`Builder` and `Delegate`),
  invoking type analysis, constant folding, and scope initialization before backend emission.
- **WGSL compiler** in `backend/wgsl`:
  Translates parsed GSL v1 `program.Program` structures into WebGPU shading code (WGSL)
  for real-time raymarched viewport previews in the browser.
- **Mesh / SDF compilers**:
  - `backend/mesh` and `internal/fidget`:
    Compile GSL AST into Fidget trees and bytecode for evaluation via Rust-based Fidget JIT and
    meshing kernels (`backend/tape` serializes standalone evaluation tapes).
  - `backend/mesh/manifold`:
    Provides mesh operations and CSG evaluation using the Manifold geometry library.
- **GSL v1 library packages** in `stdlib`:
  Contains standard GSL v1 packages embedded directly at build time
  (e.g., `builtin`, `color`, `ease`, `lattice`, `map`, `material`, `matrix`,
  `mechanical`, `noise`, `text`, `texture`, `transform`, `vector`).
- **Audio backend & return partitioning** in `assembler` and `backend/audio`:
  `assembler` (`assembler/returns.go`) partitions trailing `audio(...)` return expressions alongside
  geometry returns, and `backend/audio` compiles audio expressions to Web Audio JavaScript,
  supporting both mono and stereo output.

### Existing language documentation

- **Language Readmes & Guidelines**:
  - `README.md`:
    Repository architecture overview, SDK usage, CLI reference, and GSL language roadmap.
- **Auto-documentation Go code**:
  - `program/preamble/libraries.go`:
    Parses GSL v1 library sources and extracts top-level constant and function signatures
    documented with triple-slash (`///`) comments.
  - `program/preamble/primitives.go`, `builtin.go`, `pitfalls.go`, `gsl_description.go`:
    Preamble definitions for built-in shapes, operators, math functions, and pitfalls.
- **Generated viewer language documentation**:
  - `examples/viewer/docs/gsl_description.md`: Descriptive introduction to GSL v1 language
    paradigms.
  - `examples/viewer/docs/gsl_libraries.md`: API reference detailing all exported library constants
    and functions.

### Existing language testcases

- `parser/internal/parser2/parser_test.go`:
  Tests parsing standard test inputs matching `testdata/*.gsl` directly into `program`
  structures, including mesh and STEP import declarations.
- `parser/internal/parser2/scene_test.go`:
  Tests parsing scene annotations and directive configurations.
- `parser/internal/identifier2/validate_test.go`:
  Verifies variable, parameter, and function name checks (valid syntax, reserved words,
  and conflicts with primitives or built-in function names).
- `parser/import_ref_test.go`:
  Tests inline expression forms of `importMesh` and `importStep` inside GSL v1.
- `assembler/link_test.go`:
  Tests multi-package collection, linking, type analysis, AST lowering, and dead code elimination.
- `internal/build/build_test.go`:
  Tests builder traversal, scope creation, type checking, and delegate integration.
- `internal/eval/const_test.go` and `internal/eval/fold_test.go`:
  Validates compile-time constant evaluation and AST folding.
- `internal/assembler/property/property_test.go` and
  `internal/assembler/property/derivative_test.go`:
  Tests property lowering, companion variable expansion, and forward-mode automatic derivatives.
- `internal/types/types_test.go`:
  Tests type inference and semantic validation across expressions.
- `backend/wgsl/internal/codegen/build_test.go`:
  Tests compiler code generation outputs, float parameter bindings, and precedence
  mappings to WGSL.
- `backend/wgsl/internal/codegen/const_folding_test.go`:
  Validates compiler-level constant folding of complex arithmetic operations.
- `internal/fidget/internal/convert/convert_test.go`:
  Verifies that GSL v1 AST translates to valid fidget bytecode and Rust CSG operations.
- `program/preamble/libraries_test.go`:
  Validates that triple-slash document blocks extract correctly for standard library
  packages.

---

## Language Specification

This specification covers the syntax, semantics, and built-in constructs of GSL v1
as it is currently implemented in `omega-core`.
It will be updated as the language implementation evolves.

### 1. Language Paradigm & Key Design Constraints

* **Implicit Point (`p`)**: Every shape evaluation has access to an implicit 3D coordinate
  list `p` (representing the current point in space being evaluated). `p` is defined globally
  as a reserved identifier and must not be declared as a variable, parameter, or function.
* **Implicit Time (`time`)**: Animated shapes can
  refer to `time` (a scalar `f32` representing elapsed seconds). Like `p`, it is defined globally
  as a reserved identifier and must not be declared as a variable, parameter, or function.
* **Single Static Assignment (SSA)**: Variables are immutable and can only be assigned exactly once.
  Reassignment (e.g., `x = 5; x = x + 1;`) is a compilation error.
* **No Native Booleans**: Logic evaluates floating-point scalars:
  - Value $\le 0.0$ represents **False**.
  - Value $> 0.0$ represents **True**.
    Comparisons and logical operations return `1.0` for true and `0.0` for false.
* **Multiple Return Values & Multiple Assignment**: Functions and top-level
  source files can return multiple values. The top-level returns are automatically combined using
  a union operation (`min`) to produce the final solid scene. Multiple variable assignments (e.g.,
  `a, b, c = expr;`) are supported in **exactly one situation**: capturing multiple
  return values from a function call (e.g., `a, b, c = someFunc(1, p.x, other);`). Assigning
  multiple independent values from comma-separated expressions (e.g., `a, b = 1, 2;`) is prohibited.
* **Formatting Rule**: Indentation of function blocks must be exactly **2 spaces**.
  Mixed tabulations or 4-space styles are prohibited.

#### Audio Support

Audio is an add-on to the language and is not part and should not be part of the language proper.
The `audio(...)` expression can only appear as the last element in a top-level return statement.

**Output Channel Configuration**:
- **Mono Output**: `audio(signal)` takes a single scalar expression
  representing the mono audio signal, which is routed to both left and right speaker channels.
- **Stereo Output**: `audio(leftSignal, rightSignal)` takes two scalar expressions,
  defining independent audio waveforms for the left and right speaker channels.

**Signal Bounds**: The output signal values should ideally reside in the range $[-1.0, 1.0]$.

**Time, Frequency, and Amplitude**:
- The signal is evaluated over time via the implicit `time` variable.
- **Frequency** is controlled mathematically
  by scaling the phase factor inside periodic functions (e.g. `sin(2 * pi * frequency * time)`).
- **Amplitude** is controlled by scaling the signal expression by a multiplier factor (e.g.
  `signal * amplitude`).

**Wave Shapes**: Fundamental wave shapes are synthesized using arithmetic and primitives:
- *Sine wave*: synthesized using the built-in `sin(...)` primitive.
- *Sawtooth, Triangle, and Square waves*: synthesized programmatically using combinations
  of modulo (`%`), rounding/truncation functions (`floor`, `fract`, `round`), and ternary selectors.

**Implementation**:
Audio support is implemented via top-level return partitioning in `assembler`
(`assembler/returns.go`) and Web Audio JavaScript emission in `backend/audio`:
- `assembler.Assemble` unwraps any trailing `audio(...)` expression from the primary geometry
  returns into `Result.AudioReturns` (supporting empty geometry returns when audio is the sole
  top-level return expression).
- `backend/audio` (`audio.AudioToJs`) compiles the audio variation program (constructed via
  `assembler/variation.Program`) into mono or stereo Web Audio JavaScript.

### 2. Syntax & EBNF Grammar

GSL is parsed using the `participle/v2` framework.
The grammar is defined by the following AST structures:

```ebnf
TopLevel           = SceneDirective? Import* MeshImport* StepImport* WarpImport* Declaration*
                     ( "return" Expression ( "," Expression )* ","? ";" )?
SceneDirective     = "@scene" "(" Number ( "," Number )? ")"
Import             = "import" ( Ident | "." )? String ";"
MeshImport         = "importMesh" Int String ";"
StepImport         = "importStep" Int String ";"
WarpImport         = "importWarp" Int String ";"

Declaration        = Annotation* ( NamedFunction | PropertyAssignment | Assignment )
Annotation         = "@" Ident ( "(" ( Expression ( "," Expression )* ","? )? ")" )?

NamedFunction      = FunctionName Parameters Declaration*
                     "return" Expression ( "," Expression )* ","? ";" "}"
AnonFunction       = Parameters Declaration* "return" Expression ( "," Expression )* ","? ";" "}"
Parameters         = "(" ( ParameterDecl ( "," ParameterDecl )* ","? )? ")" "{"
ParameterDecl      = Annotation? Ident

Assignment         = VarNames "=" Expression ";"
PropertyAssignment = Ident "." Ident "=" Expression ";"
VarNames           = Ident ( "," Ident )* ","?

Expression         = LogicOr Ternary?
Ternary            = "?" Expression ( ":" Expression )? ( ":" Expression )?

LogicOr            = LogicAnd ( "||" LogicAnd )*
LogicAnd           = Equality ( "&&" Equality )*
Equality           = Comparison ( ( "==" | "!=" ) Comparison )?
Comparison         = Terms ( ( ">=" | ">" | "<=>" | "<=" | "<" ) Terms )?

Terms              = Term ( ( "+" | "-" ) Term )*
Term               = Factor ( ( "*" | "/" | "%" ) Factor )*

Factor             = "!"? "-"? Value ( "." Ident )? ( "[" Expression "]" )?

Value              = ImportRef | Number | FunctionCall | QualifiedRef | Reference | AnonCall
                   | SubExpression | Collection | String | AnnotatedExpr

AnnotatedExpr      = Annotation Expression
ImportRef          = ( "importStep" | "importMesh" | "importWarp" ) String
Number             = Float | Int
QualifiedRef       = Ident "." Ident
FunctionCall       = ( Ident "." )? Ident "(" ( Expression ( "," Expression )* ","? )? ")"
Reference          = Ident
AnonCall           = AnonFunction Arguments?
Arguments          = "(" ( Expression ( "," Expression )* ","? )? ")"
SubExpression      = "(" Expression ")"
Collection         = "[" ( Expression ( "," Expression )* ","? )? "]"
String             = StringLiteral
```

* **Trailing Commas**: Optional trailing commas (`","?`) are permitted
  across all comma-separated multi-item lists, including parameters, argument lists,
  return expressions, variable assignments, and list literals (`[...]`).
* **Viewport Annotations & Manipulators (`@point`, `@transform`, `@spline`, `uiSpline`)**:
  The parser supports `@point`, `@transform`, and `@spline` annotations on both variable
  assignments and inline expressions, desugaring them internally into `uiPoint(<expr>)`,
  `uiTransform(<expr>)`, and `uiSpline(<expr>)` calls.
  The `@spline` annotation includes automatic element promotion on list literals,
  wrapping unannotated raw matrix elements in `uiTransform` calls.
  Composite multi-frame curves can also be declared using the `uiSpline(<list>)` built-in function.
* **Expression Depth Limit**: The parser tracks the recursion depth of expressions. If nesting
  (such as nested paren blocks, functions, or conditionals) reaches **128**, parsing fails
  with `"maximum parser recursive depth exceeded"`.
* **Scene Directive (`@scene`) & Legacy Comments**: The `@scene(size, grid)` annotation configures
  viewport extent and grid spacing.
  It must be located at the top of the first file in a package before imports or declarations.
  If grid is omitted (`@scene(size)`), grid spacing defaults to `size / 10.0`.
  In GSL v1, legacy `// Scene Size:` and `// Scene Grid:` header comments are also supported
  as a fallback when `@scene` is absent.

### 3. Identifiers & Reserved Names

* **Valid Structure**: Must begin with a letter or underscore, followed by letters, digits,
  or underscores.
* **Reserved Prefixes**: Identifiers cannot start with `gcl_` or `gsl_`,
  nor can they contain the substrings `_gcl` or `_gsl`.
* **Implicit Reserved Identifiers**: `p` and `time` are globally reserved implicit identifiers
  and cannot be declared as variables, parameters, or functions.
* **Special Reserved Words**: `auto` is a reserved word and only permitted within `.box`
  property assignments.
* **Package Namespaces**: Package import prefixes (e.g., `import . "map"`) may use names that
  are reserved words or built-ins because qualified member access (`pkg.member`) is unambiguous;
  they still must satisfy standard structural rules and reserved prefix restrictions.
* **Reserved Words**: The compiler enforces a strict list of reserved words
  which cannot be declared as variables, parameters, or functions. The full list includes:
  - Primitives: `add`, `sub`, `mul`, `div`, `modulo`, `min`, `max`, `and`, `or`, `not`, `square`,
    `sqrt`, `floor`, `ceil`, `round`, `neg`, `abs`, `exp`, `ln`, `pow`, `sin`, `cos`, `tan`,
    `asin`, `acos`, `atan`, `atan2`, `gt`, `ge`, `eq`, `ne`, `le`, `lt`, `compare`, `remap`,
    `collect`, `map`, `reduce`, `access`, `select`, `audio`, `auto`
  - Built-in GSL functions: `acosh`, `asinh`, `atanh`, `clamp`, `cosh`, `cross`, `degrees`,
    `distance`, `dot`, `exp2`, `fract`, `length`, `normalize`, `radians`, `sign`, `sinh`,
    `tanh`, `trunc`, `noiseHash3`, `warpField`, `pixelSize`, `displacementDrawn`, `uiPoint`,
    `uiTransform`, `uiSpline`, `importMesh`, `importStep`, `importWarp`, `mesh`, `bvhSdf`,
    `vdbSdf`, `meshWarped`, `vdbSdfWarped`, `meshUnion`, `meshIntersect`, `meshDifference`,
    `facePoint`, `facePointNormal`, `edgePoint`, `edgePointNormal`, `vertexPoint`,
    `vertexPointNormal`, `circleCenter`, `circleCenterNormal`, `meshVolume`, `meshArea`,
    `meshMaxThickness`, `meshThickestPoint`, `meshThickestPoints`
  - Future constructs: `oneof`, `closest`, `furthest`, `open`, `close`, `shape`, `public`, `private`
  - Variable decl tags from other systems:
    `var`, `final`, `const`, `type`, `func`, `typedef`, `enum`
  - Class structure terms: `class`, `struct`, `interface`, `abstract`, `extends`, `implements`,
    `mixin`, `with`, `new`, `this`, `super`, `factory`, `operator`, `get`, `set`, `static`,
    `covariant`, `extension`
  - Control flow: `if`, `else`, `switch`, `case`, `default`, `for`, `while`, `do`, `break`,
    `continue`, `return`, `goto`, `fallthrough`
  - Async terms: `async`, `await`, `yield`, `yield*`, `go`, `chan`, `defer`
  - Error catching: `try`, `catch`, `on`, `throw`, `finally`, `rethrow`
  - Package rules: `package`, `import`, `export`, `library`, `part`, `show`, `hide`, `deferred`
  - Types: `dynamic`, `void`, `true`, `false`, `null`, `is`, `as`, `in`, `range`, `assert`,
    `external`
  - Common foreign languages (Python, C#, C++, Haskell) and graphics shading (GLSL, WGSL)
    built-in type/keyword sequences (e.g. `f32`, `i32`, `u32`, `vec2`, `vec3`, `vec4`, `discard`,
    `workgroup`, etc.). *Note*: `and`, `or`, and `not` remain reserved words for compatibility
    and diagnostics, while `&&`, `||`, and `!` are used as expressions.

### 4. Types

While GSL syntax omits explicit type annotations, the compiler enforces static type analysis
and semantic inference across the entire AST at compile time (`types.Analyze` in
[`internal/types`](../internal/types)).
Every expression, variable, and parameter is assigned a concrete compile-time type (`program.Type`).
Type errors (such as dimension mismatches, illegal swizzles, or recursion)
are caught as compile-time errors during assembly.

GSL is structured around fundamental mathematical and geometric types:

* **Scalars**: 64-bit floating-point numbers (`float64`).
  - Numeric literals (both integers and floats) represent scalar values.
  - Logical operations and comparisons also evaluate to scalar values (`1.0` for true,
    `0.0` for false).
* **Booleans**: There are no booleans in the language proper.
  Scalars are interpreted as boolean values when required by conditions or selectors:
  values greater than `0.0` are considered true, and values less than or equal to `0.0` are false.
* **Lists (Vectors)**: Fixed-length ordered collections of scalar components.
  - Constructed using bracket literals (e.g., `[x]` or `[x, y, z]`) or the `collect(...)` primitive.
  - The length of a list must be known at compile time.
  - **Sizing Tiers**:
    - **Single-Element Lists ($N=1$)**: Distinct 1D containers (`[x]`, `[x,]`, `collect(x)`).
      They do not automatically coerce to scalars.
      Indexing requires $i = 0$ (`v[0]`); any non-zero constant index produces a compile-time error.
      Swizzling allows `.x` only; referencing `.y`, `.z`, or `.w` produces a compile-time error.
      Multiplication with matrices (`[x] * mat` or `mat * [x]`) is prohibited.
    - **Hardware Vectors ($N \in [2, 4]$)**: Represent 2D, 3D, and 4D vectors mapping directly
      to native GPU vector types (`vec2`, `vec3`, `vec4` in WGSL).
      Support swizzling components (`x`, `y`, `z`, `w` up to length $N$) in any combination.
    - **Wide Lists ($N > 4$)**: Arbitrary-length fixed collections unrolled by the compiler
      into individual scalar operations or array structures.
      Swizzling is restricted to the first 4 components (`x`, `y`, `z`, `w`).
      Indexing supports `v[i]` for $0 \le i < N$.
  - Strings are represented as collections of rune scalar values.
* **Matrices**: Column-major nested collections of column vectors,
  typically representing 2D, 3D, or 4D spatial transformations.
  - Constructed using nested list literals (e.g.,
    `[[1, 0, 0, 0], [0, 1, 0, 0], [0, 0, 1, 0], [x, y, z, 1]]`
    where each inner list represents a column vector) or `collect(col0, col1, ...)`.
  - Dimensions: A matrix must contain $C \ge 2$ column vectors, each of uniform length $R \ge 1$.
    Single-column collections are treated as lists, not matrices.
    Matrix dimensions ($C \times R$) must be known at compile time.
  - **Sizing Tiers**:
    - **Hardware Matrices ($C, R \in [2, 4]$)**: Direct hardware representation mapping to GPU
      matrix types (e.g., `mat2x2`, `mat3x3`, `mat4x4` in WGSL).
    - **Wide Matrices ($C > 4$ or $R > 4$)**: High-dimensional matrices unrolled by the compiler
      into scalar and vector operations.
  - **Indexing**: Accessing a matrix with a single index `mat[c]` returns the column vector at column $c$
    ($0 \le c < C$).
    Accessing with a second index `mat[c][r]` yields the scalar component at column $c$, row $r$.
  - **No Member Access (Swizzling Prohibited)**: Property member access (e.g., `mat.x` or `mat.xy`)
    is prohibited on matrices and produces a compile-time error.
    Column indexing (`mat[c]`) and element indexing (`mat[c][r]`) must be used instead.
* **Functions / Closures**: First-class named and anonymous functions that can accept arguments,
  capture outer scopes, and be passed as parameters to other functions.
  Recursion (direct self-recursion or mutual recursion) is prohibited in GSL; all function call
  graphs must form a directed acyclic graph (DAG).

**Type Promotion & Operations**

* **Scalar Promotion (Broadcasting)**: Scalars are automatically broadcast to lists
  or matrices of matching dimensions when operated alongside lists or matrices.
  - A scalar $s$ operated alongside an $N$-component list broadcasts to each component
    (e.g., `[1, 2, 3] * 2.0` yields `[2, 4, 6]`).
  - A scalar $s$ operated alongside an $M \times N$ matrix broadcasts to every element of the matrix
    (e.g., `mat4x4 * 2.0` scales every entry).
* **Matrix Operations**:
  - **Matrix-Vector Multiplication**: Multiplying a matrix by a vector (`matrix * vector`)
    computes algebraic matrix-vector multiplication ($M \times v$).
    The number of columns in the matrix must match the vector length, producing a transformed vector
    whose length equals the row count of the matrix.
    Multiplying a matrix by a single-element list (`matrix * [x]`) is prohibited.
  - **Matrix-Matrix Multiplication**: Multiplying two matrices (`matA * matB`) computes algebraic
    matrix multiplication, requiring the column count of `matA` to match the row count of `matB`.
    Chained matrix products must be explicitly parenthesized (e.g., `(matA * matB) * matC`).
  - **Vector-Matrix Multiplication**: Expressions of the form `vector * matrix` are not supported
    (they require a transposed row vector).
  - **Matrix Arithmetic**: Addition (`+`) and subtraction (`-`) are permitted only between two
    matrices of identical dimensions ($C \times R$).
    Adding or subtracting a matrix with a scalar or list produces a compile-time type error.
  - **Matrix Negation**: Unary negation (`-mat`) negates every entry in the matrix.
* **No List-to-Matrix Promotion**: Lists are not automatically promoted to matrices.
  Attempting component-wise operations (such as `+` or `-`) between a list and a matrix produces
  a compile-time type mismatch error.
* **Relational Comparisons on Collections**:
  - Direct comparisons between a list or matrix and a scalar (e.g., `list == 1.0` or `mat < 0.0`)
    are prohibited and produce a compile-time type mismatch error.
  - Relational comparisons between two lists or two matrices of identical dimensions
    (`<`, `<=`, `>`, `>=`, `==`, `!=`, `<=>`) are evaluated component-wise.
    They return a boolean collection (list or matrix of `1.0` / `0.0` values) of the same shape.

### 5. Operators, Swizzling, & Precedence

* **Spaceship Operator (`<=>`)**: Performs a three-way comparison. `a <=> b` returns `1.0` if
  $a > b$, `0.0` if $a == b$, and `-1.0` if $a < b$.
* **Vector Swizzling**: Accessed via dot property syntax. Components can be read using `x`, `y`,
  `z`, and `w` up to 4 components in any combination (e.g. `.xy`, `.zyx`, `.xxxx`).
  Swizzling is supported on lists and scalars, but is prohibited on matrices.
* **Derivative Properties**: Accessed via dot property syntax. `.dp` computes the spatial gradient
  or surface normal with respect to `p`, and `.dtime` computes the time derivative with respect
  to `time`.
* **Indexing**: Accesses component scalars in lists using single brackets (e.g. `v[0]` for $x$,
  `v[1]` for $y$), or column vectors in matrices (`mat[0]`).
  Constant indices out of bounds produce a compile-time error.
* **Operator Precedence**:

| Precedence | Operators | Description |
| :--- | :--- | :--- |
| 1 (Highest) | `()`, `f()`, `x`, `p.x`, `v.dp`, `v.dtime`, `v[0]`, `[...]` | Calls, access, swizzling, derivatives, indexing, collections |
| 2 | `!`, `-` | Unary logical NOT, unary negation |
| 3 | `*`, `/`, `%` | Multiplication, division, modulo |
| 4 | `+`, `-` | Addition, subtraction |
| 5 | `<`, `<=`, `>`, `>=`, `<=>` | Relational comparisons, spaceship |
| 6 | `==`, `!=` | Relational equalities |
| 7 | `&&` | Logical AND |
| 8 | `\|\|` | Logical OR |
| 9 (Lowest) | `?:`, `?::` | Conditional ternary and extended three-way ternary |

* **Non-Associative Comparisons**: Relational comparisons (`<`, `<=`, `>`, `>=`, `<=>`)
  and equalities (`==`, `!=`) are non-associative.
  They cannot be chained without explicit parenthesization (e.g., `(a < b) < c`).

### 6. Ternary & Extended Elvis Operators

* **Standard Conditional**: `condition ? positiveExpression : negativeExpression`
  - If `condition` $> 0.0$ (True), evaluates to `positiveExpression`.
  - Otherwise (False), evaluates to `negativeExpression`.
* **Extended Three-Way Conditional**:
  `value ? positiveExpression : zeroExpression : negativeExpression`
  - If `value` $> 0.0$, evaluates to `positiveExpression`.
  - If `value` $== 0.0$, evaluates to `zeroExpression`.
  - If `value` $< 0.0$, evaluates to `negativeExpression`.
* **Nested Ternaries & Associativity**: The parser grammar supports nested unparenthesized
  ternary expressions (e.g. `a ? b : c ? d : e`). Because the parser grammar is right-associative,
  the AST builder automatically applies associativity correction (`fixTernaryAssociativity`)
  to rotate nested branches into their expected parent scopes.
  Parenthesizing nested conditionals or decomposing them into intermediate variables is recommended
  for readability.

### 7. Annotations

Annotations are optional compile-time only metadata declarations that begin with `@` followed by an identifier.
Annotations are parsed and stored as metadata in the AST, but are otherwise ignored by the compiler.
They are hints from the author of a GSL program to the host application about suggested ways
that the program could be changed, how functions should be called, or how to provide a better user
experience.

An annotation begins with `@` followed by an identifier and optional parenthesized arguments:

```gsl
@annotationName(arg1, arg2, ...)
```

Not all annotations are supported on all target types (declarations, parameters, or expressions).
Each annotation definition specifies which target types it supports.
Applying an unrecognized annotation or applying an annotation to an unsupported target produces
a compile-time error (`unsupported annotation @<name>`).

By default, each type of annotation may only be specified at most once per target
(declaration, parameter, or expression), unless explicitly allowed by the annotation definition.
Annotations may enforce further restrictions, such as requiring that arguments evaluate to
compile-time constants or restricting the target variable's or expression's type or value.

#### Supported Annotations

In the current GSL v1 implementation, annotations are supported as top-level scene directives,
preceding top-level variable declarations, preceding function parameters, and inline expressions:
- **`@scene`**: Top-level directive configuring package-level viewport and grid extents.
- **`@range`**: Numeric slider bounds and optional animation periods on variable declarations,
  inline numeric literals, or parameter bounds suggestions.
- **`@point`**: 3D position controls on 3-component variable declarations
  or inline expressions.
- **`@transform`**: 3D spatial transformation gizmos on 4x4 matrix variable declarations
  or inline expressions.
- **`@spline`**: Multi-handle 3D spline curve manipulators on variable declarations
  or inline expressions.

#### Range Annotations (Sliders)

A top-level variable assigned a scalar numeric literal, a function parameter,
or a scalar numeric literal within an expression can be decorated with a `@range(...)` annotation:
- **`@range(min, max)`**: Registers a slider with the given minimum and maximum bounds.
- **`@range(min, max, period)`**: Registers
  an animated slider oscillating sinusoidally between `min` and `max` over `period` seconds.
  Allowed only on variable declarations and inline expressions.
- **Function Parameters**: On function parameters, `@range` requires exactly 2 arguments:
  `@range(min, max)`. A `period` argument is not permitted on parameters and produces a parse error.
  Parameter range annotations serve as suggestions for callers and are ignored during shader emission.
- **Validation Rules**:
  - Both `min` and `max` bounds are required, where `min < max`.
  - On variable declarations, the assigned initial value must satisfy `min <= initialValue <= max`.
  - If `period` is specified (variable declarations only), it must evaluate to a strictly positive constant (`period > 0`).
  - Range arguments can be expressions, but must evaluate to numeric constants at compile time.
  - Range arguments cannot depend on spatial coordinates `p`.
  - Assigning `@range` to non-numeric types is rejected at compile time.
  - All expression dependencies within `@range` must form a directed acyclic graph (DAG);
    cyclic dependencies produce a compile-time error.
  - Violating any of these validation rules produces a compile-time error.

A `@range(...)` annotation on a variable assignment encourages the host application
to register a dynamic slider in the host user interface.

*Example Assignment Slider Configuration:*
```gsl
@range(0.5, 10.0, 4.0)
radius = 5.00;
```

*Example Function Parameter Range Annotation:*
```gsl
scaleRadius(@range(0.5, 10.0) r) {
  return r * 2.0;
}
```

In addition to explicit `@range(...)` annotations, GSL v1 supports defining slider metadata using
dot-syntax property assignments (e.g., `name.min`, `name.max`, `name.period`), which the compiler
automatically synthesizes into canonical AST range annotations. See section 8 below.

#### Point Annotations

A top-level variable or a 3-component list within an expression
can be decorated with a `@point` annotation.
The annotation is placed before the variable assignment:

```gsl
@point
pos = [1.0, 2.0, 3.0];
```
or immediately before the 3-component list expression:

```gsl
s = sphere(p - @point [1.0, 2.0, 3.0]);
```

**Restrictions**:
- The decorated expression must evaluate to a 3-component coordinate vector.

**Lowering**:
The parser automatically rewrites `@point <expr>` into `uiPoint(<expr>)`.
In the editor and host UI, this registers an interactive 3D point handle that allows
direct spatial manipulation of the vector.

#### Transform Annotations

A top-level variable assignment or a 4x4 matrix within an expression
can be decorated with a `@transform` annotation.
The annotation is placed on a line immediately before the variable assignment:

```gsl
@transform
m = [[1.0, 0.0, 0.0, 0.0],
     [0.0, 1.0, 0.0, 0.0],
     [0.0, 0.0, 1.0, 0.0],
     [3.0, 4.0, 5.0, 1.0]];
```

or immediately before the 4x4 matrix expression:

```gsl
s = myFunc(@transform [[1.0, 0.0, 0.0, 0.0],
                       [0.0, 1.0, 0.0, 0.0],
                       [0.0, 0.0, 1.0, 0.0],
                       [3.0, 4.0, 5.0, 1.0]],
            sphere(1.0));
```

**Restrictions**:
- The decorated expression must evaluate to a $4 \times 4$ transformation matrix
  (a list of 4 column vectors, each of length 4).

**Lowering**:

The parser automatically rewrites `@transform <expr>` into `uiTransform(<expr>)`.
In the editor and host UI, this registers an interactive 3D spatial transformation gizmo.

#### Spline Viewport Manipulator (@spline, uiSpline)

In GSL v1, interactive multi-handle spline curves are declared
using the `@spline` annotation or the built-in function `uiSpline`.
A spline manipulates an odd-length list of $4 \times 4$ transformation matrices
representing keyframe and tangent handles (e.g., start, bend, end)
in a coordinated multi-frame spline widget in the host UI and 3D viewport.

##### `@spline` Annotation on Variable Assignments

A variable assignment can be decorated with `@spline`:

```gsl
@transform
start = frame(startPoint, startNormal);
@transform
bend = frame(bendPoint, bendNormal);
@transform
end = frame(endPoint, endTangent);

@spline
path = [start, bend, end];
```

The parser desugars `@spline path = <expr>;` into `path = uiSpline(<expr>);`.

##### Automatic Element Promotion

When `@spline` decorates a list literal containing raw $4 \times 4$ matrices,
the parser automatically wraps each unannotated matrix element in `uiTransform`:

```gsl
@spline
path = [
  [[1.0, 0.0, 0.0, 0.0], [0.0, 1.0, 0.0, 0.0], [0.0, 0.0, 1.0, 0.0], [0.0, 0.0, 0.0, 1.0]],
  [[1.0, 0.0, 0.0, 0.0], [0.0, 1.0, 0.0, 0.0], [0.0, 0.0, 1.0, 0.0], [5.0, 0.0, 0.0, 1.0]],
  [[1.0, 0.0, 0.0, 0.0], [0.0, 1.0, 0.0, 0.0], [0.0, 0.0, 1.0, 0.0], [10.0, 0.0, 0.0, 1.0]]
];
```

Desugars into:

```gsl
path = uiSpline([
  uiTransform([[1.0, 0.0, 0.0, 0.0], ...]),
  uiTransform([[1.0, 0.0, 0.0, 0.0], ...]),
  uiTransform([[1.0, 0.0, 0.0, 0.0], ...])
]);
```

Element promotion is idempotent: elements that are already calls to `uiTransform`
or references to variables (such as `@transform` definitions) are left untouched.

##### Inline Expression Annotations

The `@spline` annotation can also decorate inline expressions,
such as within function arguments or nested constructs:

```gsl
gate.splineSection(@spline [start, bend, end], radius);
```

This desugars directly into:

```gsl
gate.splineSection(uiSpline([start, bend, end]), radius);
```

##### Direct `uiSpline` Built-in Function

Splines may also be constructed directly using the `uiSpline` built-in function:

```gsl
path = uiSpline([start, bend, end]);
```

**Constraints**:
- `uiSpline` and `@spline` require an argument evaluating to an odd-length list of at least 3
  $4 \times 4$ matrices (e.g., 3, 5, 7, ... matrices).
- Each matrix element in the list must resolve to a `uiTransform` handle
  (either directly, via `@transform`, or through automatic element promotion).
- Violating the odd-count or matrix-type requirement produces a compile-time type error
  (`"uiSpline expects an odd list of at least three 4x4 matrices, got <type>"`).
- In GSL v1, parameter annotations with `@spline` (e.g. `fn(@spline s)`) are unsupported.

**Semantics & Lowering**:
- **Identity Evaluation**: At compile time and runtime, `uiSpline` is an identity pass-through.
  It returns the input list unchanged without modifying mathematical evaluation.
  Indexing expressions such as `path[0]`, `path[1]`, and `path[2]`
  access individual handle matrices.
- **UI Registration**: The compiler UI extraction pass registers a composite `spline` control.
  The host viewer draws a visible spline curve connecting the frames above scene geometry
  and binds each matrix handle for interactive manipulation.
- **Shader Stripping**: Calls to `uiSpline` (along with `uiPoint` and `uiTransform`)
  are stripped during shader code generation and lowered to their underlying list expressions.

#### Scene Annotation

The `@scene` annotation configures the document viewport size and grid density:

```gsl
@scene(20)
@scene(20, 2)
```

* **Position**: Must be placed at the very top of the first file in a package,
  strictly preceding any imports or declarations.
* **Arguments**: Takes positional numeric constants: `@scene(size)` or `@scene(size, grid)`.
  Arguments must be strictly positive (`size > 0`, `grid > 0`).
* **Default Grid**: If `grid` is omitted, it defaults to `size / 10.0`.
* **Syntax**: Must not end with a semicolon.
* **Legacy Comments**: In GSL v1, legacy header comments (`// Scene Size:` and `// Scene Grid:`)
  are also supported as a fallback when `@scene` is omitted.

### 8. Variables, Sliders, and Properties

A top-level variable assigned a single numeric literal can be decorated with property values to
register it as a dynamic slider in the host user interface:
- **`name.min`**: Minimum value constraint.
- **`name.max`**: Maximum value constraint.
- **`name.period`**: Animation clock driver period in seconds. The host drives the value
  sinusoidally between `min` and `max` using this timeframe.
- **Snapping**: Snapping/rounding steps are configured by the decimal places of the original
  literal (`radius = 5;` snaps to integers, `radius = 0.50;` snaps to 0.01 increments).

*Example Slider Configuration:*
```gsl
radius = 5.00;
radius.min = 0.5;
radius.max = 10.0;
radius.period = 4.0; // Oscillates over 4 seconds
```

#### Slider Validation Rules
Slider properties (`.min`, `.max`, `.period`) are restricted strictly to top-level file scope.
Assigning slider properties to function-local variables, parameters, or non-numeric types
is rejected at compile time.
A variable with any slider property must declare both `.min` and `.max`, where `min < max`.
If `.period` is assigned, it must evaluate to a strictly positive constant (`period > 0`).

#### Property-to-Annotation Canonicalization
When `.min` and `.max` properties are defined on a top-level variable and no explicit `@range`
annotation was declared in the source, the compiler's scope builder automatically synthesizes
a canonical AST `Range` annotation on the defining `Assignment`.
This allows downstream compiler stages (Assembler, constant-folding reachability, WGSL uniform emission)
to operate uniformly on `Assignment.Range` without requiring dual-path handling for properties.

#### Geometry and Field Properties
Variables representing 3D geometry can be configured with metadata properties:
- **`name.box`**: Bounding box corners `[lo, hi]` where each corner is a 3D coordinate.
  Constant extents must satisfy `lo < hi` across all three axes (`x`, `y`, `z`).
  Bounding box coordinates cannot depend on spatial coordinates `p`.
  The reserved keyword `auto` may be assigned (`shape.box = auto;`) or used within coordinates
  (`shape.box = [[-10, -10, auto], [10, 10, auto + 2]];`) to automatically compute analytical
  extents from the geometry.
  `shape.box` may also be evaluated in expressions to query bounding box extents.
- **`name.color`**: Color components `[r, g, b]` or `[r, g, b, a]` normalized to `[0.0, 1.0]`.
- **`name.material`**: Material parameter collection.

#### Property Lifecycle and Lifecycle Constraints
Properties are declaration metadata and cannot be read in general expressions (except `.box`).
All property dependencies must form a directed acyclic graph (DAG).
Cyclic dependencies among declarations and properties produce a compile-time error.

### 9. Imports & Package System

#### Package Imports
Imports let a file access other package modules. They must sit at the top of the file:
```gsl
import . "vector";       // Dot import: pulls members into local scope without prefix
import "vector";         // Qualified import: elements accessed as `vector.member(...)`
import myLib "vector";   // Aliased import: elements accessed as `myLib.member(...)`
```
*Segment Rules*: Import paths cannot contain `//`, start/end with `/`, contain `.` or `..`
segments, or end with `.gsl`.

#### Asset Imports
Assets (such as STL meshes or CAD STEP files) can be imported into compile-time slots or
resolved inline.
1. **Statement Imports (Pre-allocated slots)**:
   - `importMesh Index Path;` - e.g., `importMesh 0 "gear.stl";`
   - `importStep Index Path;` - e.g., `importStep 0 "gear.step";`
   *Constraint*: Index must be a non-negative integer literal.
2. **Expression Imports (Compiler-assigned slots)**:
   - `importMesh("gear.stl")` / `importStep("gear.step")`
   - Can be assigned directly to variables (e.g. `g = importMesh "gear.stl";`) and resolve
     automatically.

**Restrictions**:
- Suffixes are restricted (`.stl` for mesh imports; `.step`, `.stp`,  or `.hodge` for STEP imports).
- Paths must be simple filenames located in the same directory the GSL file.

#### Scope Semantics, Shadowing & Import Collisions

GSL enforces structured lexical hierarchy and namespace collision constraints:

1. **Implicit `builtin` Prelude**:
   - The `builtin` package is implicitly dot-imported into every GSL file.
   - User declarations in local package scope are permitted to shadow `builtin` functions
     without emitting any compiler warning.

2. **Non-Dot-Imported Packages**:
   - Standard library packages (such as `vector` and `matrix`) are not in the local namespace
     unless explicitly dot-imported or qualified (`vector.addScalar`).
   - Declaring a user function sharing a name with a non-dot-imported package function is
     permitted without warning and executes as a user function call.

3. **Explicit Dot-Imports & Collisions**:
   - Explicitly dot-importing a package pulls all its exported members directly into the local
     file and package namespace.
   - Declaring a local variable or function with the same name as an explicitly dot-imported
     symbol produces a compile-time collision error:
     `variable or function named "<name>" already defined`.
   - If two or more explicitly dot-imported packages declare a variable with the same name,
     a compile-time collision error is reported:
     `variable "<name>" in import "<pkgB>" already defined by import "<pkgA>"`.
   - If two or more explicitly dot-imported packages declare a function with the same name
     and arity, no collision error or warning is reported; unqualified calls resolve to the
     first dot-imported package declaring that function.

4. **Package-Wide Symbol Uniqueness & Function Overloading**:
   - All source files comprising a package are parsed and validated together.
   - Declaring a package-level variable that shares a name with a variable or function in
     any other file in the package produces a compile-time error: `"<name>" is an existing variable`
     or `"<name>" is an existing function`.
   - Declaring a function that shares a name with a variable in another file produces a
     compile-time error: `"<name>" is an existing variable`.
   - Declaring a function that shares a name and parameter count with a function in another
     file in the package produces a compile-time error:
     `function "<name>" with <N> parameters is already defined`.
   - Declaring a function in another file with the same name but distinct parameter count is
     accepted as a valid multi-file overload, and the overload groups are merged across files.

### 10. Primitives & Built-in Functions

#### Single Argument Numeric Primitives and Built-ins

The following single argument numeric primitives and built-ins share these common constraints:
- Must be called with exactly 1 argument.
- The argument must be numeric (scalars, lists, or matrices).
  - For scalars, returns the result of applying the primitive or built-in to the scalar value.
  - For lists and matrices, returns a list or matrix of applying the primitive or built-in
    component-wise to each element (except `length(v)`, which returns vector magnitude).
- The return value has the same type and dimensions as the argument (except `length(v)`).

**Numeric**
- `square(a)`, `sqrt(a)`: Squares / Square roots.
- `floor(a)`, `ceil(a)`, `round(a)`: Rounding primitives.
- `abs(a)`, `neg(a)`: Absolute value and negation.
- `exp(a)`, `ln(a)`: Exponential and natural log.
- `sin(a)`, `cos(a)`, `tan(a)`, `asin(a)`, `acos(a)`: Trigonometry.
- `atan(y)`: The single argument version of the arc tangent primitive.
- `acosh(e)`, `asinh(e)`, `atanh(e)`: Inverse hyperbolic trigonometry.
- `cosh(e)`, `sinh(e)`, `tanh(e)`: Hyperbolic trigonometry.
- `degrees(e)`, `radians(e)`: Angle unit conversions.
- `fract(e)`, `trunc(e)`: Fractional and integer truncation.
- `exp2(e)`: Computes $2^e$.
- `sign(e)`: Signum function.
- `length(v)`: Euclidean length (vector magnitude) primitive.
  - *Scalar*: Returns the absolute value $|v|$.
  - *List*: Returns a scalar representing the Euclidean length $\sqrt{\sum v_i^2}$.
    For single-element lists ($N=1$), returns $|v_0|$.
  - *Matrix*: Calling `length(v)` on a matrix is prohibited and produces a compile-time error.
- `normalize(a)`: Normalization primitive.
  - *Scalar*: The normal of a scalar $s$ is the scalar itself ($s$).
  - *List*: For a list vector $\mathbf{v}$, returns the unit vector $\mathbf{v} / \|\mathbf{v}\|$
    where $\|\mathbf{v}\| = \sqrt{\sum v_i^2}$. If $\|\mathbf{v}\| = 0$, returns the zero vector.
    For single-element lists ($N=1$), returns `[sign(a[0])]`.
  - *Matrix*: Normalizes
    across all elements in the matrix $M$ by dividing every element by the scalar magnitude of all
    matrix elements combined (Frobenius norm $\|M\|_F = \sqrt{\sum_{i,j} M_{ij}^2}$): $M / \|M\|_F$.
    If $\|M\|_F = 0$, returns the zero matrix.

#### Multi-Argument Numeric Primitives and Built-ins

The following multi-argument numeric primitives and built-ins share these common constraints:
- Arguments must be numeric (scalars, lists, or matrices)
- Arguments must have matching types and dimensions, where type promotion may be applied as needed.
  - For scalars, returns the result of applying the operation to the scalar values.
  - For lists and matrices, returns a list
    or matrix of applying the operation component-wise to corresponding elements in each position.
- The return value has the same type and dimensions as the arguments.

**Numeric**
- `min(a, b, ...)`: Returns the minimum (Constructive Solid Geometry Union; at least 2 arguments).
- `max(a, b, ...)`:
  Returns the maximum (Constructive Solid Geometry Intersection; at least 2 arguments).
- `pow(base, exponent)`:
  Exponential power (exactly 2 arguments; operators `**` or `^` are not supported).
- `atan(y, x)`: Two-argument arc tangent (maps internally to `Operation_ATAN2`).
- `clamp(e, min, max)`: Clamps value to bounds (exactly 3 arguments).

**Vectors**
- `cross(e1, e2)`: Vector cross product.
  - *Constraints*: Must be called with exactly 2 arguments.
    Both arguments must be 3-component lists (after type promotion if needed).
    Calling `cross` on lists with $N \ne 3$, or matrices produces a compile-time error.
  - *Return Value*:
    Returns a 3-component list representing the cross product $\mathbf{e_1} \times \mathbf{e_2}$.
- `dot(e1, e2)`: Vector dot product.
  - *Constraints*: Must be called with exactly 2 arguments.
    Both arguments must be scalars or lists of matching length (after type promotion if needed).
    Calling `dot` on matrices is prohibited and produces a compile-time error.
  - *Return Value*: Returns a
    scalar representing the inner product $\mathbf{e_1} \cdot \mathbf{e_2} = \sum e1_i \cdot e2_i$.
- `distance(e1, e2)`: Euclidean distance.
  - *Constraints*: Must be called with exactly 2 arguments.
    Both arguments must be scalars or lists of matching length (after type promotion if needed).
    Calling `distance` on matrices is prohibited and produces a compile-time error.
  - *Return Value*:
    Returns a scalar representing the Euclidean distance $\sqrt{\sum (e1_i - e2_i)^2}$.

#### Other Numeric Primitives and Built-in Functions

- `collect(...)`: Accumulates 1 or more scalars into a list,
  or accumulates 2 or more lists into a matrix.
  - *Constraints*:
    - Must be called with 1 or more arguments.
    - For 1 argument, the argument must be a scalar, producing a 1-element list.
    - For 2 or more arguments, elements must either all be scalars
      or all be lists of the same length.
    - The return value is a list of the scalar arguments or a matrix containing the list arguments.
- `map(fn, numeric)`: Applies `fn(value)` or `fn(value, index)` to each element.
  - *Constraints*:
    - Must be called with exactly 2 arguments.
    - The first argument must be a reference to a function of the form `fn(value)`
      or `fn(value, index)`.
      - If the function takes one parameter, it corresponds to the element value.
      - If the function takes two parameters, the first parameter corresponds to the element value
        and the second parameter corresponds to the element index.
      - For lists, `index` is a 0-indexed scalar integer (`0.0`, `1.0`, etc.).
      - For matrices, `index` is passed
        as a 2-element list `[col, row]` representing 0-indexed column and row indices.
    - The second argument must be a scalar, list, or matrix.
    - The length of lists and dimensions of matrices must be known at compile time.
    - The return value has the same type and dimensions as the second argument.
- `reduce(fn, numeric [, initial])`: Folds `fn(accumulator, value)`
  over a list to produce a single scalar, or over a matrix to produce a single list.
  - *Constraints*:
    - Must be called with 2 or 3 arguments.
    - The first argument must be a reference to a function of the form `fn(accumulator, value)`
      or `fn(accumulator, value, index)`.
      - For lists, `index` is a 0-indexed scalar integer (`0.0`, `1.0`, etc.).
      - For matrices, `index` is passed
        as a 2-element list `[col, row]` representing 0-indexed column and row indices.
    - The second argument (`numeric`) must be a list or matrix.
    - The length of lists and dimensions of matrices must be known at compile time.
    - The optional third argument (`initial`) is the starting accumulator value.
      When omitted, the accumulator is seeded with the first element of the list/matrix.
    - Returns a scalar when reducing a list, or a list when reducing a matrix.
- `remap(shapeExpression, coordinateList)`: Translates, rotates, and scales shapes.
  Re-evaluates `shapeExpression` with `p` replaced by `coordinateList`.
  - *Constraints*:
    - The first argument must be a shape expression or function *call*.
      Passing a bare function identifier is prohibited (e.g.
      use `remap(box(), p)` or an anonymous call `remap(() { return sphere(1); }(), p)`).
      The compiler should detect this condition and produce an error that includes a suggestion to
      change the function reference to a function call by adding "()" after the function identifier.
    - The second argument must evaluate to a list of 3 scalars.

#### Mesh Related Primitives and Built-in Functions

The first argument (`index`) of these functions has the following common constraint:
- Must be a non-negative integer literal
  or a variable/expression bound to an imported asset slot (e.g. `importMesh` / `importStep`).

**Functions**
- `mesh(index [, placement])`: Instantiates an imported mesh in the scene.
  - *Constraints*: The optional `placement` argument must be a 3D position vector (`[x, y, z]`)
    or a $4 \times 4$ transformation matrix.
  - *Operation*: Positions, rotates, and scales the specified imported mesh asset in 3D space.
  - *Result*: Returns a shape expression representing the instantiated mesh.
- `bvhSdf(index)`: Signed distance to an imported mesh via BVH acceleration.
  - *Constraints*: Must be called with exactly 1 argument.
  - *Operation*: Evaluates the exact
    signed distance from point `p` to the imported mesh asset using a Bounding Volume Hierarchy
    (BVH) triangle-set distance oracle with Generalized Winding Number (GWN) sign determination.
  - *Result*: Returns a **scalar** representing the signed distance value.
- `vdbSdf(index [, voxelDivisor, halfBandVoxels])`:
  Signed distance to an imported mesh via baked OpenVDB level sets.
  - *Constraints*: Requires between 1 and 3 arguments.
    Optional `voxelDivisor` must be a positive integer literal (defaults to `256`).
    Optional `halfBandVoxels` must be a positive numeric literal (defaults to `3.0`).
  - *Operation*: Samples the pre-baked
    OpenVDB narrow-band level set texture of the specified mesh asset relative to point `p`.
    `voxelDivisor` defines resolution along the mesh's longest bounding box axis.
  - *Result*: Returns a **scalar** representing the signed distance value.

#### Mesh Boolean Operations (Export-Only)

- `meshUnion(a, b, ...)`: Mesh-CSG Union.
  - *Constraints*: Must be called with 2 or more mesh arguments.
  - *Operation*:
    Performs a constructive solid geometry (CSG) boolean union on the input mesh operands.
  - *Result*: Returns a combined mesh object (export-only operation).
- `meshIntersect(a, b, ...)`: Mesh-CSG Intersection.
  - *Constraints*: Must be called with 2 or more mesh arguments.
  - *Operation*:
    Performs a constructive solid geometry (CSG) boolean intersection on the input mesh operands.
  - *Result*: Returns an intersected mesh object (export-only operation).
- `meshDifference(a, b)`: Mesh-CSG Difference.
  - *Constraints*: Must be called with exactly 2 mesh arguments (`a` and `b`).
  - *Operation*: Performs a constructive
    solid geometry (CSG) boolean difference ($a \setminus b$) subtracting mesh `b` from mesh `a`.
  - *Result*: Returns a subtracted mesh object (export-only operation).

#### Miscellaneous Primitives and Built-in Functions

- `noiseHash3(coord)`: 3D procedural cell noise primitive.
  - *Constraints*: Must be called with exactly 1 argument.
    `coord` must evaluate to a 3-component vector (`vec3`).
  - *Operation*: Performs a nearest-neighbor
    lookup into a pre-baked $64 \times 64 \times 64$ noise texture tile with supercell folding.
  - *Result*:
    Returns a 3-component list (`vec3`) with component values in the normalized range $[0.0, 1.0)$.

#### Viewport Manipulator and UI Built-in Functions

- `uiPoint(coord)`: Viewport 3D point manipulator.
  - *Constraints*: Must be called with exactly 1 argument.
    `coord` must evaluate to a 3-component vector (`vec3`).
  - *Operation*: Identity pass-through in mathematical evaluation.
    Registers an interactive 3D point manipulator handle in the host editor.
  - *Result*: Returns `coord` unchanged.
- `uiTransform(matrix)`: Viewport 3D spatial transformation gizmo.
  - *Constraints*: Must be called with exactly 1 argument.
    `matrix` must evaluate to a $4 \times 4$ transformation matrix.
  - *Operation*: Identity pass-through in mathematical evaluation.
    Registers an interactive 3D translation/rotation/scale gizmo in the host editor.
  - *Result*: Returns `matrix` unchanged.
- `uiSpline(matrices)`: Viewport multi-handle spline curve manipulator.
  - *Constraints*: Must be called with exactly 1 argument.
    `matrices` must evaluate to an odd-length list of at least three $4 \times 4$ matrices
    where each element resolves to a `uiTransform` handle.
  - *Operation*: Identity pass-through in mathematical evaluation.
    Registers a composite spline curve widget connecting
    the transformation frames in the host editor.
  - *Result*: Returns `matrices` unchanged.
