# Specification: Backlog

## 1. Overview

This specification serves as the self-contained master improvement backlog for `omega-core`
(`module gradientcontrol.com/omega/core`). It consolidates planned Geometric Specification Language
(GSL) evolution from [README.md](../../../README.md) and
[doc/gsl_lang_spec_future.md](../../../doc/gsl_lang_spec_future.md) alongside compiler, evaluator,
backend, native Rust/C++ FFI, interactive viewer, CLI toolchain, and code-hygiene improvements.

Items are organized by engineering scope into three tiers so future work can be scheduled
predictably:

- **Large Initiatives (`L1`–`L10`):** Multi-pass language, compiler, codegen, runtime, or CLI
  initiatives that warrant spinning off dedicated Conductor tracks (`spec.md` + `plan.md`) when
  scheduled.
- **Medium Improvements (`M1`–`M14`):** Bounded feature completions, phased syntax retirements,
  compiler hot-path optimizations, diagnostics enhancements, viewer capabilities, and developer
  tooling tasks that can either spin off focused tracks or execute as self-contained single-phase
  tracks.
- **Small Targeted Cleanups (`S1`–`S21`):** Localized refactors, invariant tightening, dead-code
  removals, public/internal API deduplication, and documentation/comment hygiene tasks suitable for
  direct execution or small cleanup batches.

## 2. Core Engineering Principles

All compiler, parser, evaluator, and backend work scheduled from this backlog must uphold two
foundational engineering principles:

1. **Fast Happy Path, Rich Cold Error Path:**
   - Valid, error-free `.gsl` files must compile with minimal latency and heap churn: avoid
     reflection-driven `proto.Clone` calls, redundant whole-program AST copies, unconditional pass
     re-runs when an AST pass makes zero modifications, and eager diagnostic string formatting or
     symbol candidate searches on valid inputs.
   - Conversely, when an error or warning *does* occur (the cold path), spend extra cycles to emit
     high-signal, actionable diagnostics optimized for both human engineers and AI GSL authors —
     including inlined call-site provenance chains, edit-distance `"did you mean ...?"` symbol
     suggestions, 1-based parameter indices with declared parameter names and types, candidate
     overload/built-in signatures, and concrete fix hints.
2. **Boundary Validation & Clean Internal Invariants (`clean-go-test-guards`):**
   - Validate untrusted caller inputs defensively at public SDK entry points (`parser`,
     `assembler`, `eval`, `backend/wgsl`, `backend/mesh`, `backend/tape`, `backend/audio`).
   - Inside internal compiler and codegen packages (`internal/*`, `backend/*/internal/*`), rely
     directly on structural AST and scope invariants established by upstream passes
     (`parser.ParsePackage`, `collect.Program`, `scope.Init`, `link.References`, `types.Analyze`).
   - Eliminate test-induced defensive `nil` skips, unused internal parameters, and dead or
     test-only helpers from production `.go` files, keeping unit test fixtures invariant-compliant
     rather than weakening internal domain guarantees.

## 3. Phased Language Evolution Lifecycle

All breaking or surface-level GSL language changes in `omega-core` follow the four-stage rollout
defined in [README.md](../../../README.md):

1. **Add Capability:** Introduce the new syntax, AST representation, and semantic validation while
   preserving existing behavior.
2. **Automate Migration (`gsl fix`):** Add a deterministic source refactoring rule in `cmd/fix` to
   upgrade `.gsl` files automatically and migrate repository `examples/` and test fixtures.
3. **Emit Compiler Warnings:** Report actionable `program.Diagnostic` warnings pointing out legacy
   constructs and suggesting the modern replacement.
4. **Retire Legacy Syntax:** Remove legacy grammar/desugaring shims and emit targeted compile-time
   migration errors if removed syntax is encountered.

## 4. Large Initiatives

- [ ] **L1. First-Class AST Annotations & Unified UI Control Extraction**
  - **Current State:** `@material`, `@finish`, and `@surface` are stored as structured `PackageFile`
    annotations and `@range` populates `program.Range`, whereas `@point`, `@transform`, `@spline`,
    `@color`, and `@direction` desugar in `parser/internal/parser2/build_annotation.go` into
    synthetic `uiPoint`, `uiTransform`, `uiSpline`, `uiColor`, and `uiDirection` `FnCall` nodes.
    Inline `@range(min, max) <literal>` validates and discards its range without registering a
    slider, `Parameter.Range` is validated in `internal/assembler/propval` without surfacing to UI
    consumers, and `assembler/slider` (`Result.Sliders`) runs separately from `assembler/ui`
    (`Result.UIControls`).
  - **Goals:**
    - Represent all annotations natively as structured AST metadata in `program/program.proto`
      across declarations, functions, parameters, and inline expressions (resolving the `TODO` in
      `program/program.proto`: `"review how annotations are manifest in the program tree"`).
    - Register dynamic sliders for inline `@range` expressions and expose function-parameter
      `@range` metadata to callers.
    - Enforce literal-element constraints on `@point` (3 scalar literals) and `@transform` (4x4
      scalar literals) so interactive gizmos can rewrite source spans deterministically, and support
      optional `@spline` curve configuration parameters.
    - Unify `assembler/slider` and `assembler/ui` into a cohesive UI control extraction pipeline
      (resolving the `TODO` in `assembler/slider/slider.go`: `"Merge into UI controls"`).

- [ ] **L2. Unified Asset Import Syntax (`import <name> "<asset>";`) & Typed Asset Handles**
  - **Current State:** External `.stl`, `.step`/`.stp`, and `.warp` assets use slot-numbered
    statements (`importMesh <slot> "<path>";`, `importStep <slot> "<path>";`,
    `importWarp <slot> "<path>";`) or inline `importMesh`/`importStep`/`importWarp` expressions.
  - **Goals:**
    - Add unified `import <name> "<asset>";` declarations in `parser/internal/parser2`
      (`grammar.go`, `build_imports.go`) classified by standard file extensions (`.stl`, `.step`,
      `.stp`, `.warp`).
    - Bind `<name>` as a typed asset handle across `program`, `internal/scope`, `internal/types`,
      `internal/meshfold`, `internal/assembler/topofold`, `internal/assembler/warpfold`, and
      `assembler/meshplan`, assigning internal numeric slots automatically inside the compiler.
    - Add a `cmd/fix` migration rule, migrate `examples/hybrid_bracket.gsl` and
      `examples/viewer/panels/mesh-import.js`, emit compiler deprecation warnings on
      `importMesh`/`importStep`/`importWarp`, and retire numbered-slot and inline import forms.

- [ ] **L3. Compilation Cancellation API (`context.Context`)**
  - **Current State:** Entry points across `parser`, `assembler`, `eval`, and `backend/*` (`wgsl`,
    `mesh`, `tape`, `audio`) do not accept `context.Context`, preventing host applications from
    aborting superseded or timed-out compilations and high-resolution mesh builds.
  - **Goals:**
    - Introduce `context.Context` support (via `WithContext(ctx)` options or context-accepting entry
      points) across `assembler.Assemble`, `eval`, `backend/wgsl`, `backend/mesh` (including
      `internal/fidget` octree `progress_cb` cancellation and `backend/mesh/manifold`),
      `backend/tape`, and `backend/audio`.
    - Check `ctx.Err()` between compiler passes, inside bounded fixed-point inlining/unrolling/eval
      loops, and inside native meshing loops, and plumb HTTP request contexts through
      `cmd/serve/internal/compiler`.

- [ ] **L4. WGSL SDF Codegen & Raymarching Performance Optimizations**
  - **Current State:** `backend/wgsl/internal/codegen` emits flat per-part AABB guard chains
    (`vanish.go`) that mutate `var<private> gsl_partNear` and compute `gsl_vanish` screen-space
    projections on every `sdShape(p)` sphere-trace step, evaluates procedural displacement
    `gsl_heights` (`displace.go`) unconditionally in empty space, expands general 3x3 adjugate
    inverses (`stdlib/matrix/matrix.gsl`) for rigid rotations, evaluates helical trigonometry
    (`stdlib/mechanical/thread/thread.gsl`) outside thread radial bounds, and threads unused global
    texture/slot parameters into every synthesized WGSL helper (`function.go`).
  - **Goals:**
    - Group spatially proximate parts and repeated clusters into a 2-level hierarchical AABB BVH in
      `backend/wgsl/internal/codegen/vanish.go`.
    - Make `sdShape` pure and side-effect-free by moving nearest-part detail lookup out of sphere
      tracing into a post-hit shading helper (`gsl_hitPart(p)`), and hoist `gsl_vanish` visibility
      computation out of the inner raymarch loop.
    - Gate `gsl_heights` surface displacement behind `abs(baseDist) <= maxDisplacementBound`.
    - Specialize `matrix.inverse` to matrix transpose on orthogonal rotation transforms, guard
      helical thread groove math in `stdlib/mechanical/thread/thread.gsl` by radial distance, and
      prune unused global texture/uniform parameters from emitted WGSL helpers.

- [ ] **L5. Single-Pass Vectorized Spatial Derivatives (`.dp`) & Jacobian Chain Rule**
  - **Current State:** `internal/assembler/property/derivative.go` evaluates `.dp` via three
    separate scalar forward-mode automatic-differentiation passes (`k = 0..2`), tripling AST
    traversal, duplicating shared primal subexpressions, scalarizing companion function tangents,
    and failing with `"cannot differentiate builtin"` on discrete/mesh leaves (`bvhSdf`, `vdbSdf`,
    `mesh`).
  - **Goals:**
    - Replace 3-pass scalar AD with single-pass vectorized forward-mode AD seeding `p` with the 3x3
      identity matrix (`matrix.identity3()`).
    - Carry `vec3` gradients for scalar primals and column-major 3x3 Jacobians for `vec3`
      expressions, propagating derivatives across `REMAP` and vector-valued functions via
      `vector.multiply(matrix.transpose3(J), grad)`.
    - Return zero tangents for non-differentiable discrete/mesh oracle built-ins, matching existing
      noise hash leaves.

- [ ] **L6. Unified Looped AST Lowering (`map`, Two-Argument `reduce`, `reduceCSG`, & Fusion)**
  - **Current State:** `internal/lower/map_reduce.go` eagerly unrolls all `map(fn, coll)` and
    2-argument `reduce(fn, coll)` calls into `Operation_COLLECT` trees, `internal/lower/lower.go`
    maintains a bespoke `reduceCSG` path alongside an unread `Options.LoopFolds` field, and
    `internal/assembler/unroll/unroll.go` expands all `> 4`-wide list/matrix ops inside non-inlined
    functions.
  - **Goals:**
    - Preserve loop nodes (`N >= LoopFoldThreshold`) for pure homogeneous `map` (`build.MapLooper`
      emitting WGSL `for` loops) and 2-argument `reduce(fn, coll)` (`acc = coll[0]`, loop `1..N-1`).
    - Unify `reduceCSG` (`meshUnion`, `meshIntersect`) onto `Operation_REDUCE` loop lowering.
    - Fuse `reduce(combine, map(transform, coll) [, init])` patterns to avoid intermediate array
      allocations in emitted WGSL, preserve indexed loops for `> 4`-wide operations inside
      non-inlined user functions, and wire or remove `lower.Options.LoopFolds`.

- [ ] **L7. Shared Backend Builder (`internal/build`) Simplification & Scope Unification**
  - **Current State:** `internal/build/build.go` unconditionally re-runs `types.Analyze(prog)` and
    `eval.Fold(prog, pc)` on programs already processed by `assembler.Assemble`, and maintains a
    parallel `*build.Scope` hierarchy (`scope.go`, `scope_fn.go`, `scope_import.go`, `scope_var.go`)
    coupled to `internal/scope` through untyped `AgentScope` downcasts in `backend/wgsl` and
    `backend/audio`.
  - **Goals:**
    - Unify backend symbol and lexical scope resolution onto `internal/scope.ProgScope`, removing
      duplicate `*build.Scope` trees and `AgentScope` type assertions.
    - Preserve `Operand.Type` invariants across all assembler passes so `internal/build` skips
      redundant type analysis and constant folding on assembled programs, and streamline the
      `build.Delegate` visitor interface.

- [ ] **L8. Native Rust Fidget/Manifold FFI Error Handling, Panic Safety & Memory Budgets**
  - **Current State:** `internal/fidget/internal/rust/src/` (`lib.rs`, `cellular_oracle.rs`,
    `spinodoid_oracle.rs`) contains unchecked `&trees[*tree_index]` indexing and `.unwrap()` calls,
    high-subdivision octree meshing lacks a cell/allocation cap to prevent process-aborting OOMs,
    and `backend/mesh/manifold` returns bare `NULL` pointers without structured C++ failure reasons.
  - **Goals:**
    - Replace unchecked slice indexing and `.unwrap()` calls across
      `internal/fidget/internal/rust/src/` with bounds-checked `Result` error propagation (including
      both `&trees[*tree_index]` `TODO`s in `internal/fidget/internal/rust/src/lib.rs` inside
      `canvas_render` and `canvas_mesh_stl_inner`).
    - Enforce configurable octree node/memory budgets inside Fidget meshing so runaway subdivisions
      return recoverable Go errors before triggering native OOM `abort()`.
    - Ensure all C++/Rust `mglue_*` FFI entry points in `backend/mesh/manifold` catch exceptions
      and report structured failure diagnostics (non-manifold merge failure, self-intersection, or
      allocation failure).

- [ ] **L9. Canonical GSL Source Formatter (`cmd/fmt` / `gsl fmt`)**
  - **Current State:** `program/pretty_print.go` is an AST debug printer that drops comments and
    prints lowered internal constructs; the `gsl` CLI (`cmd/gsl`) has no source formatter.
  - **Goals:**
    - Build a comment-preserving CST/token GSL formatter and expose `cmd/fmt`
      (`gsl fmt [-l] [-w] [-d] [path ...]`) wired into `cmd/gsl/main.go` and `tool/build-cli.sh`,
      preserving `//` and `///` doc comments, blank-line grouping, `@` annotations, and multi-line
      collection formatting.

- [ ] **L10. Hand-Written Byte Scanner & Direct-to-AST Parser (`parser/internal/parser2`)**
  - **Current State:** `parser/internal/parser2/parser.go` drives lexing through regex-based
    `lexer.MustSimple` (8 compiled regular expressions) and builds grammar trees via
    reflection-heavy `participle.MustBuild[PackageFile]` with `participle.UseLookahead(3)` plus
    backtracking union dispatch (`participle.Union[Declaration]` and `participle.Union[Value]`).
    Every expression atom — even a bare numeric literal or variable reference — allocates an
    8-to-9-node grammar pointer chain (`Expression` -> `LogicOr` -> `LogicAnd` -> `Equality` ->
    `Comparison` -> `Terms` -> `Term` -> `Factor` -> `Number`) in
    `parser/internal/parser2/grammar.go` along with heap `*program.Location` structs
    (`posToLocation`) before `build_expr.go` lowers it to `*program.Operand` and discards the
    intermediate grammar tree.
  - **Goals:**
    - Replace regex-driven `gslLexer` in `parser/internal/parser2/parser.go` with a zero-allocation
      hand-written byte scanner over source slices (with optional trivia retention shared with `L9`
      `cmd/fmt`).
    - Replace Participle reflection grammar parsing (`parser.go`, `grammar.go`, `build_expr.go`)
      with a hand-written recursive-descent declaration parser and Pratt expression parser that
      constructs `*program.PackageFile` and `*program.Operand` nodes directly without intermediate
      CST heap chains, batching `*program.Location` allocations in contiguous slices.
    - Keep the valid-input happy path free of speculative error construction while preserving and
      expanding rich cold-path syntax diagnostics in `ParseError`
      (`parser/internal/parser2/parser.go`, including innermost parse-failure positioning,
      unclosed-delimiter and EOF span anchoring, nested-ternary parentheses hints, missing-`;`
      recovery, and legacy operator hints).

## 5. Medium Improvements

- [ ] **M1. Discrete Option Annotations (`@values`)**
  - Add AST metadata in `program/program.proto` and parser construction in
    `parser/internal/parser2/build_annotation.go` + `parser/internal/annotation/validate.go` for
    `@values(v1, v2, ..., vn)` on scalar assignments, function parameters, and inline scalar
    literals; validate `>= 2` distinct compile-time scalar constants, initial-value membership, and
    mutual exclusion with `@range` in `internal/assembler/propval`; surface discrete controls in UI
    extraction, WGSL uniform lifting, and `examples/viewer/panels/slider-panel.js`; and update
    `program/preamble`, `doc/gsl_lang_spec_current.md`, and `doc/gsl_lang_spec_future.md`.

- [ ] **M2. Phased Retirement of Slider Dot-Properties (`.min`, `.max`, `.period`)**
  - Add a `cmd/fix` rule rewriting `x = v; x.min = a; x.max = b; [x.period = p;]` into
    `@range(a, b [, p]) x = v;`; migrate the 8 `examples/*.gsl` models,
    `internal/testdata/example/*.gsl`, `program/preamble/gsl_description.go`,
    `doc/gsl_lang_spec_current.md`, and `doc/gsl_lang_spec_future.md`; emit compiler deprecation
    warnings when `.min`, `.max`, or `.period` are assigned; and subsequently remove legacy slider
    dot-properties from `internal/scope/pkg_file.go`.

- [ ] **M3. Phased Retirement of Viewport Manipulator Built-ins (`uiPoint`, `uiTransform`,
  `uiSpline`)**
  - Update `examples/viewer/panels/mesh-import.js` to emit `@transform` instead of
    `uiTransform(...)`; add a `cmd/fix` rule rewriting `uiPoint`, `uiTransform`, and `uiSpline`
    calls to `@point`, `@transform`, and `@spline`; emit compiler deprecation warnings on direct
    user calls to `ui*` built-ins; and remove user-callable `ui*` built-ins once `L1` lands.

- [ ] **M4. Complete 3-Argument `reduce(fn, collection, initial)` & Retire Legacy Argument Order**
  - Fix 3-argument `reduce(fn, coll, init)` evaluation, type inference, inlining predicates, and
    lowering across `internal/eval/const.go`, `internal/types/infer.go`,
    `internal/assembler/inline/predicates.go`, `internal/build/expression.go`, and
    `internal/fidget/custom.go`; migrate the 21 legacy `reduce(init, fn, coll)` call sites in
    `stdlib/noise/noise.gsl` and `stdlib/text/text.gsl` (regenerating `stdlib` `.pb` files via
    `./tool/build-stdlib.sh`); add `cmd/fix` rewriting and deprecation diagnostics; remove the
    dual-order fallback (`isFnOperand`); and update `program/program.proto` and `program/preamble`.

- [ ] **M5. Improve Error Messages and Locations**
  - Apply the Fast Happy Path, Rich Cold Error Path principle across the compilation pipeline so
    CLI, viewer, and SDK diagnostics highlight exact source spans and emit actionable guidance for
    both human engineers and AI GSL authors:
    - **Centralize Error Span & Source-Line Enrichment in `internal/error.go`:** Move
      `tokenWidth`/`isWordByte` and `LineFromSource` enrichment from
      `cmd/serve/internal/compiler/compiler.go` (`collectErrors`) into `internal/error.go`, using
      single-line AST `End.Column - Start.Column` spans when available, falling back to token-width
      scanning when `Length <= 1`, and fixing `scope.ErrLoc` (`internal/scope/scope.go`) setting
      `Length: 0` (which renders zero `^` carets in `program.Error.Error()`). Invoke enrichment
      upstream in `parser.ParsePackage`, `assembler.Assemble`, and `internal/build.ErrWithPkgSrc`,
      and populate `FilePath` in `cmd/compile/internal/cli/cli.go` so `gsl compile`, `gsl eval`,
      `cmd/serve`, and SDK callers all receive `file:line:col` headers, source lines, and
      full-token underlines.
    - **Promote `locateErr` to `internal/error.go` & Use `errors.As` Throughout:** Move
      `PackageAnalyzer.locateErr` from `internal/types/infer.go` into `internal/error.go` as a
      shared innermost-AST location and `PackagePath` helper (`LocateErr`) that preserves existing
      inner `*program.Error` spans via `errors.As`. Replace fragile direct `err.(*program.Error)`
      and `err.(*ErrorWithLoc)` type assertions in `internal/build/error.go` (and test runners in
      `internal/srctest/run.go` and `parser/internal/parser2/parser_test.go`) with `errors.As`, and
      resolve dependency source text in `ErrWithPkgSrc` when `PackagePath != ""`.
    - **Attach AST Node Locations Across Unlocated Compiler Passes:** Wrap bare `fmt.Errorf`
      failures with `LocateErr` at the failing AST node instead of defaulting to `1:1` across
      `internal/collect/collect.go` (missing, self-, and cyclic `import` statements at `imp.Start`),
      `internal/types/types.go` and `internal/types/infer.go` (`analyzeTopLevelVar` cycles and
      count mismatches at `assign.Start`, `@annotation` checks at `ann.Start`, and `inferUserCall`
      errors at `fnCall.Start`), `internal/eval/const.go` and `fold.go` (divide-by-zero,
      zero-vector `normalize`, constant cycles, out-of-bounds indexing, and matrix shape errors),
      `internal/lower/lower.go` and `map_reduce.go` (`map`, `reduce`, and `meshDifference`/CSG
      arity or empty-collection errors), `internal/closure/` (`validate.go`, `transform.go`,
      `specialize.go`, `unroll.go`), `internal/assembler/property/` (`property.go`,
      `derivative.go`, `parts.go`, `expand.go`, `companion.go`, `second.go`, `uv.go` for
      `.dp`/`.dtime`/`.d(var)` differentiation and `.box`/`.simple`/`.detail` errors),
      `internal/assembler/inline/` (`inline.go`, `substitute.go` recursion checks),
      `internal/assembler/topofold/topofold.go` (replacing plain `"(line %d, col %d)"` strings with
      `*program.Error`), `internal/assembler/warpfold/warpfold.go`, and `assembler/meshplan/`
      (`slice.go`, `decl.go`).
    - **Fix Parser EOF Spans, Imported-Package Attribution, String Typing & Unit Test Coverage:**
      Anchor unexpected-EOF parse errors in `parser/internal/parser2/parser.go` to the final
      non-empty token/line rather than a trailing blank line, and report the opening delimiter's
      `line:col` for unclosed `(`, `[`, and `{`. Attribute errors originating in imported packages
      (`PackagePath != ""`) to the primary file's `import` declaration or caller call site (without
      slicing `primarySrc` at the dependency's `LineNum` in `cmd/serve`) while retaining the inner
      `pkg:line:col` context. Distinguish string literals from numeric scalars and vectors in
      `parser/internal/parser2/build_value.go` and `internal/types` so invalid string arguments
      such as `sphere("x")` or `box("xyz")` fail type checking instead of lowering to rune float
      code points. Add unit tests covering `ParseError`/`innermostPosition`, `LocateErr`,
      `TokenWidth`, EOF/unclosed-delimiter spans, imported-package attribution, and CLI/viewer
      diagnostic parity.
    - **Lazy Provenance Chains, `"Did You Mean"` Hints & Human/AI Recovery Guidance:** Record
      compact, zero-string-allocation call-site provenance links during inlining in
      `internal/assembler/inline` (`inline.go`, `substitute.go`) and
      `internal/build/keyed_builtin.go` (`pushInlineSite`), formatting `"called from ..."` chains
      lazily on error, and include full cycle paths (`a -> b -> a`) on cyclic imports, variables,
      and recursive functions. Compute cold-path edit-distance `"did you mean ...?"` suggestions for
      unresolved variables, functions, package prefixes, dot-properties, and `@` annotations across
      `internal/link/link.go`, `internal/scope/pkg_file.go` (`invalidPropertyError`),
      `internal/types/infer.go`, and `parser/internal/annotation/validate.go`. Report 1-based
      parameter indices, declared parameter names, idiomatic GSL type names (`vec3`, `4x4 matrix`,
      `scalar`), operator context on `"incompatible types"` (`internal/types/types.go`), natural
      built-in signature/arity hints (`clamp(x, lo, hi)` instead of `"expects 3..3 arguments"` in
      `internal/types/builtin.go`), `.dp`/`.d(...)` differentiation provenance, and concrete
      multi-return destructuring syntax (`a, b = fn(...)`). Add targeted syntax/idiom hints for
      common human and LLM mistakes from GLSL/WGSL/JS/Python (`let`/`var`/`const`,
      `fn`/`function`/`def`, imperative `if`/`else` -> `(cond) ? a : b`, `vec2`/`vec3`/`vec4(...)`
      -> `[x, y, z]`, immutable variable reassignment citing the original declaration's `line:col`
      in `internal/scope/pkg_file.go` and `fn.go`, and bare function identifiers passed to
      `remap(fn, p)` instead of `fn()`).

- [ ] **M6. Strict Lexical Declaration-Order Enforcement in Assembler**
  - Reorder `clearance273` and `shcs4762` above their callers in
    `stdlib/mechanical/catalog/catalog.gsl` and rebuild `package.pb`; enforce strict top-to-bottom
    lexical variable visibility in `internal/link` and `internal/scope` across top-level
    assignments, property assignments, outer variable captures, and function-local statements (while
    keeping named functions order-independent); and turn the late `varInfo == nil` fallback in
    `internal/build/visit.go` into an internal invariant error (resolving the `TODO` in
    `internal/build/visit.go`).

- [ ] **M7. Screen-Space Quad Derivatives (`dpdx`, `dpdy`, `fwidth`) & Viewer Raymarching AA**
  - Register polymorphic scalar/`vec2`..`vec4` derivative built-ins (`dpdx`, `dpdy`, `fwidth`,
    `dpdxFine`, `dpdyFine`, `fwidthFine`) across `program/builtin.go`, `stdlib/builtin`,
    `internal/types`, and `internal/assembler/property`; emit native WGSL derivative calls in
    `backend/wgsl` and deterministic zero stubs in `eval`, `internal/fidget`, `backend/tape`, and
    `backend/audio`; and use screen-space derivatives in `examples/viewer/render/` for procedural
    material and specular anti-aliasing.

- [ ] **M8. Immutable WGSL `CodegenContext` & Upstream Coordinate-Remap Parameterization**
  - Replace mutable scope pointer swapping (`currBScope`/`prevBScope` and `withScopedAgent`) in
    `backend/wgsl/internal/codegen/delegate.go` with an immutable value-passed `CodegenContext`
    carrying the active scope and `ir.Block`, and parameterize coordinate remaps (`with(p = ...)`)
    into explicit `pos: vec3<f32>` parameters upstream in `assembler` (`internal/closure`,
    `internal/assembler/inline`) to eliminate WGSL `gsl_*remap` wrapper synthesis.

- [ ] **M9. Interactive Viewer (`cmd/serve`) Hybrid & Mesh-Only `mesh(...)` Preview Support**
  - Wire `assembler.ExtractMeshPlan` and `assembler/meshplan` preview SDF slicing into
    `cmd/serve/internal/compiler/compiler.go` to populate `CompileResponse.MeshDecls`, enabling
    simultaneous SDF raymarching + Three.js mesh placement previews for hybrid models (such as
    `examples/hybrid_bracket.gsl`) and clean zero-SDF-return compilation for mesh-only models, with
    coverage in `cmd/serve/internal/compiler/compiler_test.go`.

- [ ] **M10. Re-enable 10 Skipped Fidget Conversion Golden Tests**
  - Diagnose and fix the 10 `.glsl` conversion golden tests currently marked `"skip"` in
    `internal/fidget/testdata/` (`builtin`, `car`, `common_type`, `flying_wing`, `func`, `mod`,
    `multi_return`, `overload`, `simple`, `wgsl_keywords`), generate verified goldens, and remove
    the `"skip"` bypass in `internal/fidget/convert_test.go` (resolving its
    `"remove this once all tests are passing"` `TODO`).

- [ ] **M11. Go License Tool Migration & `stdlib/text` `cxf2gsl` Generator Restoration**
  - Replace `tool/third_party_licenses.py` with a pure Go command under
    `tool/third-party-licenses/` (aligning with `AGENTS.md` Go-only developer tooling policy), and
    restore a self-contained Go `cxf2gsl` font-to-GSL generator in `stdlib/text/_tools/cxf2gsl/`
    with updated `stdlib/text/_tools/README.md` instructions (and optional glyph generation above
    `U+3040`).

- [ ] **M12. Complete `cubicBaldwinReinforced` Lattice Function in `stdlib/lattice/honeycomb`**
  - Verify and finalize the commented-out `WIP. TODO.` `cubicBaldwinReinforced` SDF implementation
    at the bottom of `stdlib/lattice/honeycomb/honeycomb.gsl` using certified `eval`
    bounds/clearance checks, add unit tests, uncomment the function, and regenerate
    `stdlib/lattice/honeycomb/package.pb`.

- [ ] **M13. Assembler Fixed-Point Pass Gating, Incremental Re-Linking & Fold/Type Allocation
  Reduction**
  - Eliminate avoidable happy-path compiler passes and per-node heap churn in `assembler.Assemble`:
    - Gate `link.References` and `types.Analyze` inside `optimizeProgram` and Step 4 of
      `assembler/assemble.go` on actual AST mutation booleans (returning
      `(modified bool, err error)` from `lower.Transform` in `internal/lower/lower.go`) instead of
      unconditionally re-linking and re-analyzing 4 times per fixed-point iteration when
      `inline.Functions`, `eval.Fold`, `constspec.Transform`, `lower.Transform`, or
      `unroll.Functions` made no changes.
    - Remove the duplicate back-to-back `deadcode.Remove` invocation between the final
      `optimizeProgram` iteration and Step 7 in `assembler/assemble.go`, and reuse
      `assemblerParamChecker` (`newAssemblerParamChecker`) across non-mutating passes rather than
      re-walking `prog.Primary` twice per iteration.
    - Split fast silent reference re-binding (`link.References`) from cold collision/shadowing
      warning passes (`checkDotImportCollisions`, `checkLocalDeclarationCollisions`,
      `checkBuiltinShadowing`) in `internal/link/link.go`, and skip re-linking unmodified
      `prog.Dependencies` on internal assembler re-links.
    - Reuse `constCtx` cycle-detection maps (`evaluating`, `calls`) on `foldVisitor` and preserve
      matching single-element `op.ConstValues` slices on numeric literals in
      `internal/eval/fold.go` (`VisitValue`, `evalConst`) instead of reallocating per AST node.
    - Share canonical immutable `*program.Type` singletons (`Scalar`, `Vec2`, `Vec3`, `Vec4`) in
      `program/type.go` and add a zero-slice single-type fast path alongside `inferOperandType` in
      `internal/types/infer.go` to eliminate per-operand `[]*program.Type` wrapper allocations.
    - Avoid building and discarding `*program.Package` trees on intermediate reachability rounds,
      replace `O(F^2)` slice prepending (`append([]*program.PackageFile{newFile}, files...)`), and
      prune dependencies with zero reachable declarations in
      `internal/assembler/deadcode/deadcode.go`.

- [ ] **M14. Typed AST Cloning, In-Place Assembler Ownership & Conditional `meshfold`
  Short-Circuiting**
  - Eliminate reflection-heavy `proto.Clone` calls across all 39 production call sites and skip
    redundant whole-program clones when `assembler.Assemble` already owns a fresh `Program`:
    - Replace `proto.Clone` with typed `program.Clone*` helpers (`program/clone.go`, adding
      `CloneProgram`, `CloneAssignment`, and `CloneFunction`) across `program/get_import.go`
      (`GetStdPackage`), `parser/caching.go` (`CachingParser`),
      `internal/fidget/convert/buildcopy.go` (`BuildCopy`), `internal/assembler/property/`
      (`carry.go`, `remap_wrappers.go`, `share.go`), `internal/closure/` (`capture.go`,
      `specialize.go`, `transform.go`, `unroll.go`), `internal/lower/map_reduce.go`,
      `internal/assembler/constspec/constspec.go`, and `assembler/meshplan/inline.go`, and rewrite
      `program.EqualModuloLocations` (`program/program.go`) as a direct structural AST comparator
      without `proto.Clone` + `protoreflect` mutation.
    - Mutate in place (or clone only `Primary.Files`) inside `property.ExpandShared`
      (`internal/assembler/property/expand.go`) and `property.MarkParts`
      (`internal/assembler/property/parts.go`) when operating on the exclusively owned `Program`
      from `collect.Program`, avoiding full `Dependencies` deep copies.
    - Short-circuit `meshfold.FoldProgramWithSlotsChecked` (`internal/meshfold/fold.go`, enabled by
      default via `enableSlotFold: true` in `assembler/assemble.go`) when a program declares no
      mesh/STEP/warp imports and invokes no `mesh`, `bvhSdf`, or `vdbSdf` built-ins, avoiding two
      full-program `FoldMeshIndicesAt` clones plus an extra `eval.Fold` pass on analytical SDFs.

## 6. Small Targeted Cleanups

- [ ] **S1. Generalized Postfix Modifier Chaining (`Modifiers []*Modifier`)**
  - Replace single `Access *Accessor` + `Index *Expression` fields on `Factor` in
    `parser/internal/parser2/grammar.go` and `build_expr.go` with repeatable `Modifiers []*Modifier`
    (`@@*`) folding `.member`, `.dp`/`.dtime`/`.d(var)`, and `[index]` left-to-right (supporting
    expressions such as `m[0][1]`, `pts[0].x`, and `body.box[0]`).

- [ ] **S2. Phased Retirement of Legacy `// Scene Size:` & `// Scene Grid:` Header Comments**
  - Run `bin/gsl fix` across the 72 `.gsl` files in `internal/testdata/example/` and inline test
    fixtures, emit compiler deprecation warnings in `parser/internal/parser2/legacy_scene.go` when
    legacy comments are parsed, and subsequently delete `legacy_scene.go`.

- [ ] **S3. Reject Legacy `not(...)`, `and(...)`, `or(...)`, and `atan2(...)` Calls with Hints**
  - Remove `"and"`, `"or"`, and `"not"` from `primitives` in `program/primitive.go` (preserving
    `Operation_AND`/`OR`/`NOT` for `&&`/`||`/`!`), emit targeted parse diagnostics suggesting `!`,
    `&&`, `||`, and `atan(y, x)` (when `atan2(y, x)` is called), and add a `cmd/fix` rewriter for
    legacy functional logical calls.

- [ ] **S4. Retire Legacy Prelude Import Alias Shims in Parser & Auto-Import**
  - Update legacy test fixtures across `assembler/`, `parser/`, `internal/eval/`, `eval/`,
    `internal/fidget/convert/testdata/`, and `backend/wgsl/internal/codegen/testdata/` that still
    import the legacy prelude package alias, add a `cmd/fix` cleanup rule, and remove the silent
    `vector`/`builtin` import rewrite shims in `parser/internal/parser2/build_imports.go`,
    `internal/scope/pkg_file.go`, and `program/auto_import.go`.

- [ ] **S5. Migrate `cmd/serve` Off `audio.ExtractAudio` & Delete String-Based Audio Splitter**
  - Refactor `cmd/serve/internal/compiler/compiler.go` (and `backend/audio` tests) to run a single
    `assembler.Assemble(..., assembler.WithVariants())` pass with `assembler/variation.Program`
    instead of invoking `audio.ExtractAudio` and assembling twice, then delete `audio.ExtractAudio`
    and `backend/audio/internal/extract/`.

- [x] **S6. Enforce `tool/analyze-deps -rules` in `tool/check.sh` & Remove Dead `internal/pkg`**
  - Resolve the 5 active `tool/analyze-deps -rules` layer/pass-separation violations
    (`internal/assembler/topofold/topofold.go`, `program/range_test.go`,
    `program/returns_traversal_test.go`, and `internal/lower/lower_test.go`), delete the unused
    `internal/pkg/manager.go` package, narrow eligible `internal/*` packages, and wire
    `go run ./tool/analyze-deps -rules` into `tool/check.sh`.

- [ ] **S7. Make `program.Program` Self-Contained & Remove `(*Program).GetImport` Global Fallback**
  - Ensure `internal/collect`, `internal/link`, and lowering passes populate `prog.Dependencies`
    with every referenced standard library package (updating synthetic `Program` structs in unit
    tests), and remove the global `stdPackage(path)` fallback in `program/get_import.go` (resolving
    its `"Remove this when possible"` `TODO`).

- [ ] **S8. Zero-Copy Local Asset Resolution & Allocation-Free `source.Manager` Lookup**
  - Add local file path or `io.ReadCloser` streaming asset resolution on `assembler/source`
    (`internal/source/local.go`) so large `.stl`/`.step`/`.vdb` assets need not be copied into heap
    `[]byte` slices, and refactor `Manager.Sources` and `Manager.Asset` (`internal/source/multi.go`)
    to return `nil` on package/asset miss instead of allocating `ErrPkgNotFound` wrappers (resolving
    both `TODO`s in `internal/source/multi.go`).

- [ ] **S9. Split Material Resolution Out of `property/color.go` & Fix Multi-Return Forwarding**
  - Extract material constants, defaults, and row clamping from the 985-line
    `internal/assembler/property/color.go` into `internal/assembler/property/material.go` (resolving
    the `TODO` in `internal/assembler/property/color.go`), and fix property-gradient differentiation
    when tail-forwarding an undestructured multi-return call (`return fn(...);`, currently worked
    around in `stdlib/material/coat/weathering/weathering.gsl`).

- [ ] **S10. Replace Regex WGSL Type Parsing & Clean Up Post-Unrolling Broadcasting Fallbacks**
  - Replace regex type-name parsing (`arrayTypeRE`, `matrixTypeRE`) in
    `backend/wgsl/internal/codegen/simple.go` (`blendRows`) with structured assembler type metadata
    (resolving the `TODO` in `backend/wgsl/internal/codegen/simple.go`), remove obsolete
    post-unrolling broadcasting fallbacks in `internal/fidget/convert/type.go`
    (`vecType.ExpandTrees`) and `backend/wgsl/internal/wgsl/numeric_op.go`, and flatten single-use
    cross-statement `let` chains for associative accelerated ops in `internal/lower/lower.go`.

- [ ] **S11. Resolve Residual `inline`, `eval`, `meshfold`, and `fidget` Code `TODO`s**
  - Extract the `"gsl_tmp_"` constant in `internal/assembler/inline/inline.go` (`TODO`) and resolve
    the `TODO` on `shouldInlineMultiReturnCall` in `internal/assembler/inline/predicates.go`;
    extract `"matrix"`/`"vector"`/`"multiply"` constants at both `TODO` sites in
    `internal/eval/const.go`; inline trivial `FoldProgram` and document cloning behavior on
    `FoldProgramWithSlots` in `internal/meshfold/fold.go` (2 `TODO`s); and warn or check
    parent-scope variable shadowing in `internal/fidget/convert.go` (`TODO` in `validateName`) via
    `internal/scope`.

- [ ] **S12. Audit `program/meshasset.Registry` vs. `Cache` & Reserve Deprecated Tape Field**
  - Consolidate or deprecate index-keyed `program/meshasset/registry.go` in favor of
    content-addressed `program/meshasset/cache.go`, and reserve deprecated
    `SpinodoidOracle.levelBias` (`field 9`) in `backend/tape/tape.proto` and
    `internal/fidget/internal/rust/src/tape.rs` once no legacy serialized tapes remain.

- [ ] **S13. Self-Contained Code Comments, Tool Test Fixtures & Asset Extensions**
  - Rewrite remaining pre-extraction repository paths and external application/service names in
    comments and test fixtures across `program/`, `stdlib/*.gsl` (rebuilding `package.pb` via
    `./tool/build-stdlib.sh`), `backend/`, `internal/assembler/property/`, `internal/fidget/`,
    `examples/viewer/*.css`, `tool/run-go-tests/internal/*_test.go`, `tool/build-fidget.sh`, and
    `stdlib/pkg_info_test.go`; fix the `@slider` -> `@range` typo in `README.md`; and retire legacy
    non-standard asset extensions in `parser/internal/parser2/build_imports.go` once `L2` lands.

- [x] **S14. Eliminate Redundant Public/Internal API Aliases & Pass-Through Wrappers**
  - Consolidate duplicated type/const aliases (`type X = internalpkg.X`) and thin 1-to-1
    pass-through wrapper functions so each public SDK type and function has a single canonical
    definition without redundant `internal` mirror packages:
    - `parser/parser.go`: `type Result = parser2.Result` (`parser/internal/parser2`).
    - `eval/eval.go` & `eval/types.go`: 7 type aliases (`Evaluator`, `AssetFetcher`, `BoundsResult`,
      `CertifyResult`, `ExtremumResult`, `SurfaceSample`, `VolumeBounds`), 3 const aliases
      (`ReasonBudget`, `ReasonPool`, `ReasonTolerance`), and 4 pass-through functions (`New`,
      `SliderRanges`, `TopLevelNames`, `TopLevelReturnExprs`) wrapping `internal/gsleval`; 5 type
      aliases (`LitView`, `LitOptions`, `LitCamera`, `LitSimple`, `LitLinear`, chained through
      `internal/fidget/lib` to `internal/fidget/internal/mesh`) and 17 pass-through functions
      (`EvalGrad`, `EvalGradGroups`, `EvalInterval`, `EvalIntervalParams`, `EvalIntervalParamBoxes`,
      `Render`, `RenderFramed`, `RenderLit`, `RenderLitLinear`, `RenderLitSized`,
      `RenderLitSizedView`, `RenderLitSizedViewRadiance`, `ComposeSheet`, `ViewPresetNames`,
      `ParseView`, `ParseViews`, `ParseVec3`) wrapping `internal/fidget/lib`; and 12 type aliases
      (`GradSample`, `GradGroup`, `IntervalSample`, `EvalAssetsRequest`, `EvalAssetsResponse`,
      `ImportSlotRef`, `MeshOraclePayload`, `VdbOraclePayload`, `AssetDigest`, `MapRangeDigest`,
      `RenderGlslResponse`, `RenderInput`) wrapping `internal/fidget` (plus
      `type Option = assembler.Option`).
    - `eval/joints/joints.go`: 7 type aliases (`Joint`, `Claim`, `Verdict`, `Prover`, `ClaimResult`,
      `JointReport`, `Report`) and 6 pass-through functions (`Find`, `Others`, `Claims`,
      `RegionExprs`, `Region`, `Run`) mirroring `internal/gsleval/joints` 1-for-1.
    - `backend/mesh/emit.go`: 3 type aliases (`MeshSdfOracle`, `VdbOracle`, `BakedSpinodoid`,
      re-pointed to canonical `meshasset` / `spinodoid` types and marked `Deprecated` for removal in
      `0.2.0`) and 6 pass-through functions (`BakeSpinodoids`, `RegisterMeshOracle`,
      `RegisterVdbOracle`, `BakeAndExtractVdbWithProgress`, `ExtractVdbDenseBytes`,
      `ExtractVdbLevelSet`) wrapping `internal/fidget/lib` (and `internal/fidget/internal/convert`).
    - `backend/tape/emit.go`: 3 type aliases (`MeshSdfOracle`, `VdbOracle`, `SpinodoidOracle`,
      re-pointed to canonical `meshasset` / `spinodoid` types and marked `Deprecated` for removal in
      `0.2.0`) and 2 pass-through functions (`GenerateTapeWithSpinodoids`,
      `GenerateTapeWithAppearanceAndSpinodoids`) wrapping `internal/fidget/lib` (duplicating
      `MeshSdfOracle` and `VdbOracle` with `backend/mesh`).
    - `backend/audio/emit.go`: 2 pass-through functions (`AudioToJs`, `ExtractAudio`) wrapping
      `backend/audio/internal/lib`.
    - `cmd/{compile,fix,gsl}/main.go`: exported `Run` pass-through functions wrapping
      `cmd/*/internal/cli.Run`.
    - Inverted internal-equals-public and internal-to-internal shims: `internal/slider/slider.go`
      (`Extract`, `Lifted` delegating to `assembler/slider`), `internal/ui/ui.go` (`Extract`,
      `ExtractWithOperands` delegating to `assembler/ui`), `internal/source` (`type Manager`
      duplicating `assembler/source.Manager` and `var ErrPkgNotFound = pubsource.ErrPkgNotFound`),
      and `backend/internal/common/build.go` (`Builder`, `Delegate`, `Scope`, `Expression` aliasing
      `internal/build`).

- [ ] **S15. Support Hexadecimal Numeric Literals (`0x...`) & 6-Digit `#RRGGBB` Color Literals**
  - Add hexadecimal integer literal lexing (`0[xX][0-9a-fA-F]+` ordered before `Float` and `Int` in
    `gslLexer`) and `Number` parsing (`strconv.ParseUint` -> `float64`) across
    `parser/internal/parser2/parser.go`, `grammar.go`, `capture.go`, and `parse_scene.go` (with
    clear parse errors on malformed `0x` prefixes).
  - Add strictly 6-hex-digit `#RRGGBB` sRGB color literals in `parser/internal/parser2`
    (`parser.go`, `grammar.go`, `build_value.go`) lowering to `[r/255, g/255, b/255]` (with
    `@color` UI control support in `assembler/ui` and `examples/viewer/`), emitting targeted parse
    errors on non-6-digit `#...` forms (with a CSS digit-doubling hint on 3-digit `#RGB`, e.g.,
    suggesting `#FFFFFF` for `#fff` and `#AABBCC` for `#abc`).
  - Keep `color.fromHex` in `stdlib/color/color.gsl` unchanged for now, add a `cmd/fix` rule
    rewriting `color.fromHex("#RRGGBB")` / `color.fromHex("RRGGBB")` calls to `#RRGGBB`, update
    syntax highlighting in `examples/viewer/editor/editor.js`, and document both literals in
    `doc/gsl_lang_spec_current.md`, `doc/gsl_lang_spec_future.md`, and
    `program/preamble/gsl_description.go`.

- [x] **S16. Tighten `.color` Property Validation & Docs to Strictly 3-Element RGB (`vec3`)**
  - Tighten `validateColorProperty` in `internal/assembler/propval/validate.go` (and unit tests in
    `internal/assembler/propval/validate_test.go`) to require `len(vals) == 3` rather than allowing
    stale 4-element `[r, g, b, a]` vectors that fail downstream in `internal/assembler/property`,
    `backend/wgsl`, and `internal/fidget`, and remove `[r, g, b, a]` from `name.color` in
    `doc/gsl_lang_spec_current.md` and `doc/gsl_lang_spec_future.md`.

- [ ] **S17. Remove `compile eval`, `compile skill` & `compile viewer` in Favor of `gsl` Subcommands Only**
  - Remove the `os.Args[1] == "eval"`, `os.Args[1] == "skill"`, and `os.Args[1] == "viewer"`
    branches from `cmd/compile/main.go` and the `compile eval` usage line from
    `cmd/compile/internal/cli/cli.go` so `cmd/compile` only compiles GSL sources; move
    `cmd/compile/internal/evalcmd` (plus `cmd/compile/render_eval_test.go`) and
    `cmd/compile/internal/skillcmd` into `cmd/gsl/internal/` and dispatch `viewer.Run`
    (`cmd/serve/viewer`) directly from `cmd/gsl/internal/cli/cli.go` (updating usage headers,
    `cmd/gsl/internal/cli/cli_test.go` + `viewer_test.go`, and unifying static wheel linkage in
    `tool/build-wheel.sh` and `python/omega-core` so Fidget is not linked twice).

- [ ] **S18. Pre-Index `wgsl.Render` Span Lookups, Memoize `reachesKeyedBuiltin` & Avoid Double WGSL
  Serialization**
  - Pre-index top-level assignments and functions once in `backend/wgsl/render.go`
    (`newLocationCollector`) and key `visited` by typed structs/pointers instead of re-scanning
    `packageAssignments(c.module)` and `packageFunctions(c.module)` (`O(R * A^2)`) and formatting
    `fmt.Sprintf("assign:%p:%d", ...)` / `fmt.Sprintf("fn:%p", ...)` keys on every reference.
  - Fast-skip `reachesKeyedBuiltin` in `internal/build/keyed_builtin.go` when
    `!prog.GetGlobalTextures().GetUsesSpinodoid()` and memoize reachability per `*program.Function`
    rather than allocating `visited`/`anonFnDecls` maps and `fmt.Sprintf` keys on each `FnCall`.
  - Track `withDisplacement` calls directly during `BuildFuncCall` rather than serializing
    `d.module.WGSL()` mid-build in `backend/wgsl/internal/codegen/build.go` and scanning the full
    module string in `displacementEvident` (`backend/wgsl/internal/codegen/displace.go`);
    synchronize or pre-allocate `vecTypes`/`listVecTypes` in `internal/fidget/convert/type.go`; and
    replace `fmt.Sscanf(name, "gsl_tmp_%d", &n)` in `internal/assembler/inline/inline.go`
    (`initAnonCounter`) with `strings.TrimPrefix` + `strconv.Atoi`.

- [ ] **S19. Audit & Remove Test-Induced `nil` Guards, Unused Parameters & Dead/Test-Only Helpers
  (`clean-go-test-guards`)**
  - Preserve defensive input validation at public SDK boundaries (`parser`, `assembler`, `eval`,
    `backend/*`) while restoring clean internal domain invariants across Go production files:
    - **Delete unreferenced dead helpers:** remove `subdivide8` (`eval/certify.go`), `shadeLit`
      (`eval/litrender.go`), `(*pcg64).uniform` (`program/noise/spinodoid/numpy.go`), `NewSubExpr`
      (`backend/wgsl/internal/wgsl/expression.go`), `VecType` and `NewListVecType`
      (`internal/fidget/convert/type.go`), `HodgeSpinodoidOracleID`, `RegisterHodgeSpinodoid`, and
      `HodgeSpinodoidTree` (`internal/fidget/fidget.go`), and `JSONEq`
      (`internal/test/assert/json.go`).
    - **Relocate or inline test-only internal exports out of production `.go` files:** move
      test-only wrappers/constructors into `*_test.go` (or call production APIs directly) for
      `NewUniformDecl` (`backend/wgsl/internal/ir/decl.go`), `Expand`
      (`internal/assembler/property/expand.go`), `NewEngine`
      (`internal/assembler/property/property.go`), `UFields` and `VFields`
      (`internal/assembler/property/uv.go`), `GraphPackages`
      (`internal/assembler/warpfold/warpfold.go`), `ConstElementsPkg` (`internal/eval/const.go`),
      `ToColorShapes` (`internal/fidget/convert/color.go`), `UnrollReduce`
      (`internal/lower/map_reduce.go`), `FoldMeshIndices` (`internal/meshfold/fold.go`), and
      `IsSupported` (`parser/internal/annotation/annotation.go`).
    - **Drop unused internal parameters & `_`-rename unused interface parameters:** remove unused
      `op` on `(*sliceCtx).sliceMeshCSG` and `(*sliceCtx).sliceMeshLiteral`
      (`assembler/meshplan/slice.go`), `expressions` on `(*Delegate).BuildMeshSdfCall`
      (`backend/wgsl/internal/codegen/mesh_sdf.go`), `currentPkgPath` on
      `(*inliner).shouldInlineCall` and both `calleePkgPath` and `currentPkgPath` on
      `(*inliner).shouldInlineMultiReturnCall` (`internal/assembler/inline/predicates.go`),
      `pkgPath` on `addCompanions` (`internal/assembler/property/carry.go`), `op` on
      `(*warpFolder).rewrite` (`internal/assembler/warpfold/warpfold.go`), `fc` on
      `(*transformer).enclosingNamesForList` (`internal/closure/enclosing.go`), `env` on
      `(*transformer).unrollFunctionFold` (`internal/closure/unroll.go`), `name` on
      `(*Delegate).builtinUnaryScale` (`internal/fidget/convert/builtin.go`), and `fc` on
      `(*FnScope).CallFrom` (`internal/scope/fn.go`), and rename unused interface-mandated
      parameters to `_` across `internal/link`, `internal/eval`, `internal/assembler/deadcode`,
      `internal/assembler/property`, `internal/scope`, `internal/fidget/convert`,
      `backend/wgsl/internal/`, and `backend/audio/internal/codegen`.
    - **Eliminate test-induced internal `nil` skips & update test fixtures:** remove redundant
      `if file == nil`, `if a == nil`, `if v == nil`, `if ft == nil`, `if g == nil`, and
      `if fn == nil` guards inside internal passes (`internal/meshfold/fold.go`,
      `internal/assembler/property/parts.go` and `carry.go`, `internal/assembler/inline/inline.go`
      and `substitute.go`, `internal/eval/const.go`, and `backend/wgsl/render.go`), document
      upstream non-nil AST invariants, and update unit test AST fixtures to satisfy them.

- [ ] **S20. Add `Diagnostics` to `parser.Result` & Warn on Legacy Screen Size Comments**
  - Add `Diagnostics []program.Diagnostic` to `parser.Result` (`parser/parser.go`) and propagate
    parser diagnostics from `parser/internal/parser2` through `parser.ParsePackage` and
    `assembler.Assemble` (`assembler/assemble.go`, alongside existing `Diagnostics`).
  - When `parser/internal/parser2/legacy_scene.go` (`fromLegacyComments`) parses legacy
    `// Scene Size:` / `// Scene Grid:` header comments into a `@screen` annotation (`*program.Scene`),
    emit a `program.SeverityWarning` diagnostic at the legacy comment location advising the user to
    use `@screen` instead and run `gsl fix` to update the source (supporting `S2`).

- [ ] **S21. Display Compiler & Parser Warnings in the Example Viewer When the Code Panel Is Open**
  - Surface `SeverityWarning` diagnostics from `cmd/serve/internal/compiler` (`CompileResponse`)
    into `examples/viewer/` (`examples/viewer/main.js` and `examples/viewer/editor/editor.js` via
    `setEditorDiagnostics` or a dedicated code-panel banner/gutter) so warnings are visible when the
    code editor panel is open without blocking rendering.

## 7. Acceptance Criteria

- [ ] Every backlog item is categorized into `Large`, `Medium`, or `Small` with concrete,
      code-verified `omega-core` context and relative paths.
- [ ] All documentation, terminology, and path references within this specification are completely
      self-contained inside `omega-core`.
- [ ] Compiler, parser, codegen, and diagnostics items uphold the Core Engineering Principles:
      fast, low-allocation happy-path compilation on valid `.gsl` inputs, rich lazy cold-path
      diagnostics on errors/warnings, and strict boundary validation paired with clean internal
      domain invariants (`clean-go-test-guards`).
- [ ] Large and Medium entries provide clear architectural boundaries and sub-goals from which
      standalone Conductor tracks (`spec.md` + `plan.md`) can be created as needed, while Small
      entries define self-contained cleanups verifiable via `./tool/check.sh` and `./tool/test.sh`.

## 8. Out of Scope

- **Native Boolean Type (`bool`):** Introducing a separate `bool` primitive type with implicit
  scalar conversion is intentionally excluded; GSL retains scalar `> 0.0` truthiness.
- **Parser Expression Depth Limit Removal:** Lifting the 127-level `PkgFileScope.IncExprDepth`
  recursion cap is excluded from the planned roadmap.
- **External Service, RPC, or Cloud Infrastructure Concerns:** Microservice dispatchers, remote
  RPC transport layers, authentication, and cloud deployment pipelines lie outside `omega-core`.
