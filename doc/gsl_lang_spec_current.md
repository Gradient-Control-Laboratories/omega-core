# GSL v1: Compiler Notes

**The language is documented by `gsl skill`.** Its text is generated from
[program/preamble](../program/preamble) (`GSLDescription`, the built-in and primitive descriptions, the
pitfalls) and from the standard library's `///` comments, and the documentation test builds every example
in it. Read that for syntax, semantics and the library. `gsl skill` prints the index, `gsl skill NAME` one
package, group or member, and `gsl skill --grep TEXT` searches.

This file keeps what that description leaves out on purpose: how the compiler parses, types and lowers the
language. [compiler-arch.md](./compiler-arch.md) gives the pass order;
[gsl_lang_spec_future.md](./gsl_lang_spec_future.md) designs what is not built yet.

## Targets

- **WGSL**, for the WebGPU preview in a browser.
- **Fidget trees and tapes**, for meshing, export and the CPU evaluators (the Rust Fidget JIT and
  Manifold kernels behind cgo).
- **Web Audio JavaScript**, for a document's `audio(...)` return (below).

## Grammar

GSL is parsed with `participle/v2` ([parser/internal/parser2](../parser/internal/parser2)). The AST
structures amount to this grammar:

```ebnf
TopLevel           = ( SceneDirective DurationDirective? | DurationDirective SceneDirective? )?
                     CaptionDirective* Import* MeshImport* StepImport* WarpImport*
                     ( Declaration | CaptionDirective )*
                     ( "return" Expression ( "," Expression )* ","? ";" )?
SceneDirective     = "@scene" "(" Number ( "," Number )? ")"
DurationDirective  = "@duration" "(" Number ")"
CaptionDirective   = "@caption" "(" String "," Number "," Number ( "," Number )? ")"
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

Factor             = "!"? "-"? Value Modifier*
Modifier           = "." Ident Arguments? | "[" Expression "]"

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

- Every comma-separated list takes an optional trailing comma.
- The parser counts expression depth and fails at 128 ([internal/scope](../internal/scope)).
- The grammar is right-associative for nested ternaries (`a ? b : c ? d : e`); the AST builder rotates the
  branches into place (`fixTernaryAssociativity`).
- There is no unary `+`, and assignment is a statement, not an operator.
- The parser desugars a few forms: `min(list)` and `max(list)` become `builtin.unionSharp` and
  `builtin.intersectSharp`; `atan(y, x)` becomes `Operation_ATAN2`.

## Annotations

The parser handles an annotation in one of three ways. It attaches slider metadata (`@range`). It lowers
the annotated expression into a `ui*` built-in (`@point`, `@direction`, `@color`, `@transform`, `@spline`).
Or it records declaration or parameter metadata on the AST (`@material`, `@finish`, `@surface`, `@frame`,
and control annotations on function parameters, which are recorded and never lowered). No annotation
changes what a program computes.

- `@point e` becomes `uiPoint(e)`, `@direction e` becomes `uiDirection(e)`, `@direction(anchor) e` becomes
  `uiDirection(e, anchor)`, `@color e` becomes `uiColor(e)` and `@transform e` becomes `uiTransform(e)`.
- `@spline path = e;` becomes `path = uiSpline(e);`. On a list literal, the elements that are raw
  matrices are wrapped in `uiTransform`. This is idempotent: elements that are already `uiTransform`
  calls, or references, are left alone.
- The `ui*` calls are identities. UI extraction (`assembler.WithExtractUI`) registers their controls,
  and shader generation strips them to their arguments.
- `@material`, `@finish` and `@surface` go into `PackageFile.Annotations`. They mark a name, covering every
  overload of an annotated function, and never change what the declaration compiles to.

## Types

`types.Analyze` ([internal/types](../internal/types)) gives every expression, variable and parameter a
`program.Type`. It types a function at each call, from that call's argument types; recursion is caught
here.

- Vectors of 2 to 4 components map to WGSL `vec2` to `vec4`. Longer lists are unrolled into scalars or
  arrays, with swizzles limited to `x` through `w`.
- Matrices of up to 4 by 4 map to WGSL `matCxR`; larger ones are unrolled.
- `internal/lower` turns `+ - * /` and unary minus on vectors and matrices into calls to the `vector` and
  `matrix` helpers, and unrolls `map` and `reduce` over lists of known length.
- Two passes specialize functions before lowering. `internal/assembler/listlen`, ahead of the closure
  pass, builds `list(n)` and `list(n, fn)`, cloning a function per constant length it is called with.
  `internal/assembler/listsize`, after typing, clones a function whose fold carries a matrix and indexes
  it by the element index, per list size, so the fold unrolls with constant indices.

## Audio

[assembler/returns.go](../assembler/returns.go) moves a trailing `audio(...)` top-level return out of the
geometry returns into `Result.AudioReturns`; geometry may be empty when audio is the only return.
`backend/audio` (`audio.AudioToJs`) compiles that variation program
(`assembler/variation.Program`) into mono or stereo Web Audio JavaScript.

## Meshes and Tables

- `bvhSdf(index)` is the exact signed distance to an imported mesh, from a bounding-volume hierarchy over
  its triangles, signed by the generalized winding number.
- `vdbSdf(index[, voxelDivisor[, halfBandVoxels]])` samples a pre-baked OpenVDB narrow-band level set.
  `voxelDivisor`, a positive integer literal with default 256, sets the resolution along the mesh's
  longest box axis. `halfBandVoxels`, a positive literal with default 3, sets the band.
- `noiseHash3(coord)` reads a pre-baked 64³ noise tile with supercell folding.
- The spinodoid builtins key a kernel table baked at compile time. A graded table is baked over the call's
  `box`; a stationary one repeats.
