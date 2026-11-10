
# GSL v2 Language Specification (future)

This is a language specification for the future programming language GSL v2 (GCL Shape Language).

This GSL v2 is built upon the design and syntax goals of GSL v1.
It contains new ideas that are not currently implemented
and represents the intended design goals for the GSL language moving forward.
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
- Resolves variables and slider properties (@range) dynamically
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

## Language Specification

This specification covers the syntax, semantics, and built-in constructs of GSL v2
representing the intended design goals for the GSL language moving forward.

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

**Implementation**
Audio support should be implemented as a pre-processor in `omega/agent/js` (`omega/agent/js/lib/lib.go`)
which transforms a single primary source file containing audio output (`audio(...)`) into two files:
- the original GSL source with the `audio(...)` expression removed entirely.
  This may result in an empty top-level return expression if audio was the only return value.
- a new audio-only GSL source with the `audio(...)` wrapper extracted for mono or stereo evaluation,
  or an empty string if no `audio(...)` expression is present.

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
* **First-Class AST Annotations**: In GSL v2, annotations are first-class constructs
  in the Abstract Syntax Tree (AST), represented directly across declarations, parameters,
  and expressions. Annotations are not desugared into synthetic function calls or external sidecars;
  they are preserved as structured AST nodes (`Annotation`) containing their directive name,
  source positions, and optional compile-time argument expressions.
* **Viewport Annotations (`@point`, `@transform`, `@spline`)**: The parser supports `@point`,
  `@transform`, and `@spline` annotations on variable assignments and inline expressions.
  As first-class AST elements, viewport annotations attach interactive UI metadata
  (point manipulators, transformation gizmos, and multi-handle spline widgets) directly
  to the AST nodes, without altering downstream shader code generation or mathematical evaluation.
  For `@spline`, unannotated raw matrix list literals automatically promote to transform handles.
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
    `tanh`, `trunc`, `noiseHash3`, `warpField`, `pixelSize`, `displacementDrawn`,
    `importMesh`, `importStep`, `importWarp`, `mesh`, `bvhSdf`, `vdbSdf`,
    `meshWarped`, `vdbSdfWarped`, `meshUnion`, `meshIntersect`, `meshDifference`,
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
[`omega/common/program/types`](../common/program/types)).
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

In GSL v2, all annotations are first-class constructs of the Abstract Syntax Tree (AST).
Rather than being treated as transient syntax, desugared into synthetic wrapper functions,
or compiled away as lossy comments, annotations are directly preserved as first-class AST nodes
attached to their syntactic targets.

Annotations communicate structured metadata from the program author to language tooling,
editor environments, and host applications—specifying interactive manipulators, recommended
parameter ranges, discrete selection values, and viewport configurations—while remaining cleanly
decoupled from mathematical evaluation and shader code generation.

#### First-Class AST Structure

An annotation begins with `@` followed by an identifier and optional parenthesized arguments:

```gsl
@annotationName(arg1, arg2, ...)
```

In the AST, every annotatable construct (top-level directives, declarations, variable assignments,
named functions, function parameters, and inline expressions) directly incorporates its associated
annotation nodes:
- **`Name`**: The annotation identifier (e.g., `scene`, `range`, `values`, `point`,
  `transform`, `spline`).
- **`Args`**: Optional sequence of compile-time argument expressions.
- **`Pos` / `EndPos`**: Exact source location spans for IDE tooling and compiler diagnostics.

Not all annotations are supported on all target types (declarations, parameters, or expressions).
Each annotation definition specifies which target types it supports.
Applying an unrecognized annotation or applying an annotation to an unsupported target produces
a compile-time error (`unsupported annotation @<name>`).

By default, each type of annotation may only be specified at most once per target
(declaration, parameter, or expression), unless explicitly allowed by the annotation definition.
Annotations may enforce further restrictions, such as requiring that arguments evaluate to
compile-time constants or restricting the target variable's or expression's type or value.

#### Supported Annotations

Annotations are supported as top-level scene directives, on variable declarations,
named functions, function parameters, and inline expressions:
- **`@scene`**: Top-level directive configuring package-level viewport and grid extents.
- **`@range`**: Numeric slider bounds and optional animation periods on variable declarations,
  function parameters, or inline expressions.
- **`@values`**: Discrete option selectors on variable declarations, function parameters,
  or inline expressions.
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
- **Snapping**: Snapping/rounding steps are configured by the decimal places of
  the original literal (`radius = 5;` snaps to integers, `radius = 0.50;` snaps to 0.01 increments).
- **Validation Rules**:
  - Both `min` and `max` bounds are required, where `min < max`.
  - The assigned initial value must satisfy `min <= initialValue <= max`.
  - If `period` is specified, it must evaluate to a strictly positive constant (`period > 0`).
  - Range arguments can be expressions, but must evaluate to numeric constants at compile time.
  - Range arguments cannot depend on spatial coordinates `p`.
  - Assigning `@range` to non-numeric types is rejected at compile time.
  - All expression dependencies within `@range` must form a directed acyclic graph (DAG);
    cyclic dependencies produce a compile-time error.
  - Violating any of these validation rules produces a compile-time error.

A `@range(...)` annotation on a variable assignment or inline expression encourages the host application
to register a dynamic slider in the host user interface.

*Example Assignment Slider Configuration:*
```gsl
@range(0.5, 10.0, 4.0)
radius = 5.00;
```

*Example Expression Slider Configuration:*
```gsl
s = sphere(@range(0.5, 10.0) 5.00);
```

A `@range(...)` annotation on a function parameter suggests that any caller provide a value
within the specified range.
On function parameters, `@range` requires exactly 2 arguments: `@range(min, max)`.
A third `period` argument does not make sense on a parameter and produces a compile-time parse error.
For example, if a programmer wants a function parameter to be limited to values
between 0.5 and 10.0, they can add the `@range(0.5, 10.0)` annotation to the parameter.
The host application can use this information to provide feedback to the user
about the expected range of values for the parameter.
Shader code generators ignore parameter-level `@range` annotations during emission
as these annotations are suggestions for callers to the functions whose parameters
contain these annotations.

*Example Function Parameter Range Annotation:*
```gsl
someFunc(a, @range(0.5, 10.0) b, c) {
  ...
  return d;
}
```

#### Values Annotations (Discrete Options)

A top-level variable assigned a scalar numeric literal, a function parameter,
or a scalar numeric literal within an expression can be decorated with a `@values(...)` annotation.
- **`@values(v1, v2, ..., vn)`**: Registers
  a discrete UI selector allowing choice among specified scalar values ($v1, v2, \dots, vn$).
- **Restrictions**:
  - The `@values` parameters can be scalar expressions,
    but must evaluate to numeric constants at compile time.
  - At least 2 value parameters must be provided.
  - It can be applied to any top-level variable assigned a scalar literal,
    a function parameter, or a scalar literal within an expression
    (matching all contexts where `@range` is permitted).
  - It is a compile error if any of the above restrictions are not satisfied.

A `@values(...)` annotation on a variable or expression encourages the host application
to register a discrete selection UI control in the host user interface.

*Example Discrete Values Configuration:*
```gsl
@values(1.0, 2.0, 4.0, 8.0)
stepSize = 2.0;
```

*Example Expression Discrete Values Configuration:*
```gsl
s = step(@values(1.0, 2.0, 4.0, 8.0) 2.0);
```

A `@values(...)` annotation on a function parameter suggests that any caller provide a value
among the specified discrete values. For example, if a programmer wants a particular function
parameter to be limited to values among 1.0, 2.0, 4.0, or 8.0, they can add the
`@values(1.0, 2.0, 4.0, 8.0)` annotation to the parameter.
The host application can use this information to provide feedback to the user
about the expected set of values for the parameter.

*Example Function Parameter Discrete Values Annotation:*
```gsl
someFunc(a, @values(1.0, 2.0, 4.0, 8.0) b, c) {
  ...
  return d;
}
```

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
- The target being annotated must be a top-level variable or a 3-component list within an expression.
- The variable must be assigned a value.
- The value must be a 3-component list.
- All of the values in the list must be scalar numeric literals.
- It is a compile error if any of the above restrictions are not satisfied.

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
- The target being annotated must be a top-level variable or a 4x4 matrix within an expression.
- The variable must be assigned a value.
- The value must be a 4x4 matrix.
- All of the values in the matrix must be scalar numeric literals.
- It is a compile error if any of the above restrictions are not satisfied.

#### Spline Annotations

A top-level variable assignment or an odd-length list of 4x4 matrices within an expression
can be decorated with a `@spline` annotation.
The annotation is placed on a line immediately before the variable assignment:

```gsl
@spline
path = [
  [[1.0, 0.0, 0.0, 0.0], [0.0, 1.0, 0.0, 0.0], [0.0, 0.0, 1.0, 0.0], [0.0, 0.0, 0.0, 1.0]],
  [[1.0, 0.0, 0.0, 0.0], [0.0, 1.0, 0.0, 0.0], [0.0, 0.0, 1.0, 0.0], [5.0, 0.0, 0.0, 1.0]],
  [[1.0, 0.0, 0.0, 0.0], [0.0, 1.0, 0.0, 0.0], [0.0, 0.0, 1.0, 0.0], [10.0, 0.0, 0.0, 1.0]]
];
```

or immediately before the list expression:

```gsl
s = myFunc(@spline [start, bend, end], radius);
```

**Automatic Element Promotion**:
When `@spline` decorates a list literal containing raw 4x4 matrices,
the parser automatically wraps each unannotated matrix element in a transform manipulator.
Element promotion is idempotent: elements that are already transform handles
or references to variables (such as `@transform` definitions) are left untouched.

**Restrictions**:
- The target being annotated must be a top-level variable assignment or an inline list expression.
- The variable must be assigned a value.
- The value must be an odd-length list of at least 3 4x4 matrices (e.g., 3, 5, 7, ... matrices).
- Each matrix element in the list must resolve to a transform handle
  (either directly via `@transform` or through automatic element promotion).
- Parameter annotations with `@spline` (such as `fn(@spline s)`) are not supported.
- It is a compile error if any of the above restrictions are not satisfied.

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

### 8. Geometry and Field Properties

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
import . "omega";        // Dot import: pulls members into local scope without prefix
import "omega";          // Qualified import: elements accessed as `omega.member(...)`
import myLib "omega";    // Aliased import: elements accessed as `myLib.member(...)`
```
*Segment Rules*: Import paths cannot contain `//`, start/end with `/`, contain `.` or `..`
segments, or end with `.gsl`.

#### Asset Imports
Assets (such as STL meshes or CAD STEP files) can also be imported.
```gsl
import gMesh "gear.stl";
import gStep "gear.step";
```

It is convention although not required that asset imports appear after package imports.

**Restrictions**:
- Suffixes are restricted (`.stl` for mesh imports; `.step`, `.stp`,  or `.hodge` for STEP imports).
- Paths must be simple filenames located in the same directory the GSL file.

#### Scope Semantics, Shadowing & Import Collisions

GSL enforces structured lexical hierarchy and namespace collision constraints:

1. **Implicit `builtin` Prelude & Shadowing Warnings**:
   - The `builtin` package is implicitly dot-imported into every GSL file.
   - User declarations in local package scope are permitted to shadow `builtin` functions.
   - When a user function shadows an implicitly imported function from `builtin` with the
     same name and arity (parameter count), the compiler emits a diagnostic warning:
     `warning: function "<name>" with <N> parameters shadows implicitly imported builtin function`.
   - Shadowing warnings are generated strictly for `builtin`; no warnings are produced when
     declaring functions sharing names with other un-imported packages.

2. **Non-Dot-Imported Packages**:
   - Standard library packages (such as `vector` and `matrix`) are not in the local namespace
     unless explicitly dot-imported or qualified (`vector.addScalar`).
   - Declaring a user function sharing a name with a non-dot-imported package function is
     permitted without warning and executes with complete lexical fidelity.

3. **Explicit Dot-Imports, Warnings & Ambiguity Errors**:
   - Explicitly dot-importing a package pulls all its exported members directly into the local
     file and package namespace.
   - Declaring a local variable or function with the same name as an explicitly dot-imported
     symbol produces a compile-time collision error:
     `variable or function named "<name>" already defined`.
   - If two or more explicitly dot-imported packages declare a variable with the same name,
     the compiler emits a diagnostic warning on the dot-import:
     `warning: variable "<name>" declared in both import "<pkgA>" and "<pkgB>"`.
   - If user code makes an unqualified reference to a colliding variable name, the compiler
     reports a compile-time ambiguity error at the reference site:
     `ambiguous reference to variable "<name>": defined in import "<pkgA>" and "<pkgB>"`.
   - If two or more explicitly dot-imported packages declare a function with the same name and
     arity (parameter count), the compiler emits a diagnostic warning on the dot-import:
     `warning: function "<name>" with <N> parameters declared in both import "<pkgA>" and "<pkgB>"`.
   - If user code makes an unqualified call to a colliding function name and arity, the compiler
     reports a compile-time ambiguity error at the call site:
     `ambiguous call to "<name>" with <N> arguments: defined in import "<pkgA>" and "<pkgB>"`.
   - Qualifying a variable reference or function call with its explicit package prefix
     (e.g. `<pkgA>.<name>` or `<pkgA>.<name>(...)`) resolves the ambiguity.

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

## Migration from GSL v1 to GSL v2

For a detailed list of differences between GSL v1 and GSL v2, mechanical syntactic conversion
guidelines, and current migration status, see [gsl_migration.md](gsl_migration.md).

---

## Implementation Notes

This section details implementation-specific details
that are outside the formal language specification but used in the current implementation:

- **Scalar Precision**: Scalars are implemented as 32-bit single-precision floating-point
  numbers (`f32`) when compiled for GPU shader models (such as WGSL) or native JIT bytecode targets.

- **atan2(y, x)**: If the compiler detects `atan2(y, x)`
  it should report an error along with a suggestion that `atan(y,x)` should be used instead.

- **and(...), or(...),
  and not(...)**: If the compiler detects function calls to `and(...)`, `or(...)`, or `not(...)`, it
  should report a compilation error along with a suggestion to use `&&`, `||`, or `!` instead (e.g.
  replace `not(x)` with `!x`, `and(a, b)` with `a && b`, `or(a, b)` with `a || b`).

- For language elements of GSL v1 that are removed in GSL v2, the compiler should
  report an error along with a suggestion to change the language element to the GSL v2 equivalent.

- There should be an api for requesting that an current in progress compilation be canceled

---

## Proposed Implementation & Design Decisions

This section explores ideas, enhancements,
and design questions for the next iteration of the GSL v2 language compiler.

### 1. Introduction of Native Booleans
* **Idea**: GSL v1 lacks native boolean types, treating $\le 0.0$ as false and $> 0.0$ as true.
  We could introduce a native `bool` type to prevent numerical logic errors
  and optimize WGSL code generation (where WGSL natively uses `bool` for conditions).
* **Questions**:
  - Should the language parser strictly enforce boolean types for ternaries (`cond ? a : b`),
    or should we allow implicit conversion from `float` to `bool`?
  - If we introduce `bool`, do we preserve operators like `and(a, b)` as numeric functions,
    or do we introduce short-circuiting logical operators (`&&`, `||`)?
* **Answer**: Implicit conversion from scalar to boolean.

### 2. Easing SSA Verbosity with Local Mutability
* **Idea**: GSL v1 is strictly Single Static Assignment (SSA) at the syntax level. Reassignments
  are forbidden. We could introduce local mutability (e.g. `var x = 1.0; x = x + 2.0;`) within
  function bodies or script scope. The compiler can automatically desugar this into SSA form (e.g.,
  standard renaming: `x_0 = 1.0; x_1 = x_0 + 2.0;`) before generating WGSL or bytecode.
* **Questions**:
  - Should we require
    an explicit keyword like `var` to declare mutable variables to distinguish them from constants?
  - Does local mutability make code significantly more readable for complex geometry computations,
    or does it run contrary to the functional, declarative nature of the shape engine?
* **Answer**: No, keep v2 the same as v1 in this regard so as to keep the complexity down.

### 3. Loop and Iteration Constructs
* **Idea**: Currently, repetition is restricted to `map` and `reduce` over small,
  constant-sized lists. We could add bounded loop constructs (e.g.,
  `for (i = 0 .. N)` or `while`) where `N` must be a compile-time constant.
  This allows the compiler to unroll loops completely during compilation to GPU/WGSL shader models.
* **Questions**:
  - Is a bounded `for` loop (with compile-time constant limits) sufficient,
    or is there a need for general-purpose, dynamic loops?
  - If dynamic loops are introduced, how should we handle execution limits
    and prevent GPU hangs (e.g., safety recursion bounds or runtime loop count checks)?
* **Answer**: No, keep v2 the same as v1 in this regard, at least for now.
  Need time to really think about the implications of loops.
  
### 4. Overcoming the Recursive Depth Limit
* **Idea**: The parser limits recursion depth to 127.
  If an expression tree is deeply nested (which is common in large CSG hierarchies), parsing fails.
  We can redesign the parser's expression grammar or use an iterative
  parser tree builder to process expressions of arbitrary depth without blowing Go's stack.
* **Questions**:
  - Is the depth limit of 127 a limitation of participle recursion,
    or is it a security safety mechanism to prevent stack overflow?
  - Should we handle deep hierarchies by encouraging procedural decomposition,
    or should the parser natively support arbitrary depth?
* **Answer**: The recursive depth limit is a limitation of the current parser/codegen implementation
  and not a language feature.
  Base the new limit on the capabilities of the new compiler implementation.

### 5. First-Class Functions and the `remap` Constraint
* **Idea**: In GSL v1, the first argument to `remap(shapeExpression, coordinateList)` must
  be a concrete shape function *call*, not a bare function identifier (e.g. `remap(sphere(1), p)`
  is valid, but `remap(sphere, p)` is invalid). We can support first-class function identifiers
  so that passing the function reference itself without immediate evaluation works naturally.
* **Questions**:
  - If we allow passing bare function identifiers, how
    should the compiler resolve overloaded functions (since GSL v2 supports function overloading)?
  - How do we handle closures and captures in WGSL codegen if functions are passed dynamically?
* **Answer**: No, keep v2 the same as v1 in this regard because the function may have parameters
  that need to be evaluated before passing the result to remap. May revisit this in the future,
  but needs more thought and research before considering this change.

### 6. Annotations as replacement for Slider Properties and Viewport built-ins
* **Idea**: Instead of using dot property assignments (`name.min`, `name.max`, `name.period`) or a
  dedicated `slider` keyword, we can use annotations to attach UI metadata directly to assignments.
  Similarly, viewport built-ins like `uiPoint`, `uiTransform`, and `uiSpline`
  can be converted into annotations rather than special wrapper functions.

  *Example Annotation Usage:*
  ```gsl
  @range(0.5, 10.0, 4.0)
  radius = 5.00;

  @point
  pos = [1.0, 2.0, 3.0];

  @transform
  m = [[1.0, 0.0, 0.0, 0.0],
       [0.0, 1.0, 0.0, 0.0],
       [0.0, 0.0, 1.0, 0.0],
       [1.0, 2.0, 3.0, 1.0]];

  @spline
  path = [m, m, m];

  // Parameter annotations in function declarations
  myShape(@range(0.1, 2.0) radius, center) {
    return sphere(radius) + center;
  }
  ```
* **Existing Properties to Convert**:
  - Slider accessors (`name.min`, `name.max`, `name.period`) -> `@range(min, max, [period])`
* **Existing Built-ins to Convert**:
  - `uiPoint(expr)` -> `@point` preceding the variable assignment
  - `uiTransform(expr)` -> `@transform` preceding the variable assignment
  - `uiSpline(expr)` -> `@spline` preceding the variable assignment
* **Questions**:
  - Should annotations be completely ignored by the code generation backend,
    passing through only as metadata captured in the program AST?
  - Can annotations be applied to functions (e.g.
    `@name`, `@description` or `@doc` to replace triple-slash `///` comments)?
  - How should the compiler handle conflicting or duplicate annotations on the same target (e.g.,
    both `@min` and `@range` on the same assignment/parameter)?
* **Answer**:
  - Yes, annotations are first-class constructs in the program AST, preserved directly across
    declarations, parameters, and expressions for tool and UI consumption, while ignored by the
    shader code generation backend.

### 7. Multiple Assignments in the Same Statement
* **Idea**: Support multiple assignments in a single statement (e.g., `a, b = 1, 2;`).
* **Answer**: Multiple variable assignments (e.g.,
  `a, b, c = expr;`) are supported in **exactly one situation**: capturing multiple return
  values from a function call (e.g., `a, b, c = someFunc(1, p.x, other);`). Assigning multiple
  independent values from comma-separated expressions (e.g., `a, b = 1, 2;`) is not supported.
