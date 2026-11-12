# Specification: Backlog

## 1. Overview

This specification serves as the self-contained master improvement backlog for `omega-core`
(`module gradientcontrol.com/omega/core`). It consolidates planned Geometric Specification Language
(GSL) evolution from [README.md](../../../README.md) and
[doc/gsl_lang_spec_future.md](../../../doc/gsl_lang_spec_future.md) alongside compiler, evaluator,
backend, native Rust/C++ FFI, interactive viewer, CLI toolchain, and code-hygiene improvements.

Items are organized by engineering scope into three tiers so future work can be scheduled
predictably:

- **Large Initiatives (`L1`–`L9`):** Multi-pass language, compiler, codegen, runtime, or CLI
  initiatives that warrant spinning off dedicated Conductor tracks (`spec.md` + `plan.md`) when
  scheduled.
- **Medium Improvements (`M1`–`M12`):** Bounded feature completions, phased syntax retirements,
  diagnostics enhancements, viewer capabilities, and developer tooling tasks that can either spin
  off focused tracks or execute as self-contained single-phase tracks.
- **Small Targeted Cleanups (`S1`–`S17`):** Localized refactors, invariant tightening, dead-code
  removals, public/internal API deduplication, and documentation/comment hygiene tasks suitable for
  direct execution or small cleanup batches.

## 2. Phased Language Evolution Lifecycle

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

## 3. Large Initiatives

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

## 4. Medium Improvements

- [ ] **M1. Discrete Option Annotations (`@values`)**
  - Add AST metadata in `program/program.proto` and parser construction in
    `parser/internal/parser2/build_annotation.go` + `parser/internal/annotation/validate.go` for
    `@values(v1, v2, ..., vn)` on scalar assignments, function parameters, and inline scalar
    literals; validate `>= 2` distinct compile-time scalar constants, initial-value membership, and
    mutual exclusion with `@range` in `internal/assembler/propval`; surface discrete controls in UI
    extraction, WGSL uniform lifting, and `examples/viewer/panels/slider-panel.js`; and update
    `program/preamble` and `examples/viewer/docs/gsl_description.md`.

- [ ] **M2. Phased Retirement of Slider Dot-Properties (`.min`, `.max`, `.period`)**
  - Add a `cmd/fix` rule rewriting `x = v; x.min = a; x.max = b; [x.period = p;]` into
    `@range(a, b [, p]) x = v;`; migrate the 8 `examples/*.gsl` models,
    `internal/testdata/example/*.gsl`, `program/preamble/gsl_description.go`, and
    `examples/viewer/docs/gsl_description.md`; emit compiler deprecation warnings when `.min`,
    `.max`, or `.period` are assigned; and subsequently remove legacy slider dot-properties from
    `internal/scope/pkg_file.go`.

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
    `internal/fidget/internal/fidget/custom.go`; migrate the 21 legacy `reduce(init, fn, coll)` call
    sites in `stdlib/noise/noise.gsl` and `stdlib/text/text.gsl` (regenerating `stdlib` `.pb` files
    via `./tool/build-stdlib.sh`); add `cmd/fix` rewriting and deprecation diagnostics; remove the
    dual-order fallback (`isFnOperand`); and update `program/program.proto` and `program/preamble`.

- [ ] **M5. Actionable Compiler Error Provenance, Parameter Attribution & Legacy Construct Hints**
  - Preserve call-site provenance frames across `internal/assembler/inline` so `program.Error`
    attributes errors inside inlined library functions to the caller's source span with
    `"called from ..."` context; include 1-based parameter index and declared parameter name in
    function and built-in argument type mismatch diagnostics (`internal/types`); add built-in
    signature hints; and report targeted hints for common mistakes (such as passing a bare function
    identifier to `remap(fn, p)` instead of `fn()`).

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
    `internal/fidget/internal/fidget/testdata/` (`builtin`, `car`, `common_type`, `flying_wing`,
    `func`, `mod`, `multi_return`, `overload`, `simple`, `wgsl_keywords`), generate verified
    goldens, and remove the `"skip"` bypass in `internal/fidget/internal/fidget/convert_test.go`
    (resolving its `"remove this once all tests are passing"` `TODO`).

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

## 5. Small Targeted Cleanups

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
  - Update legacy test fixtures across `assembler/`, `parser/`, `internal/eval/`,
    `internal/gsleval/`, `internal/fidget/internal/convert/testdata/`, and
    `backend/wgsl/internal/codegen/testdata/` that still import the legacy prelude package alias,
    add a `cmd/fix` cleanup rule, and remove the silent `vector`/`builtin` import rewrite shims in
    `parser/internal/parser2/build_imports.go`, `internal/scope/pkg_file.go`, and
    `program/auto_import.go`.

- [ ] **S5. Migrate `cmd/serve` Off `audio.ExtractAudio` & Delete String-Based Audio Splitter**
  - Refactor `cmd/serve/internal/compiler/compiler.go` (and `backend/audio` tests) to run a single
    `assembler.Assemble(..., assembler.WithVariants())` pass with `assembler/variation.Program`
    instead of invoking `audio.ExtractAudio` and assembling twice, then delete `audio.ExtractAudio`
    and `backend/audio/internal/extract/`.

- [ ] **S6. Enforce `tool/analyze-deps -rules` in `tool/check.sh` & Remove Dead `internal/pkg`**
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
    post-unrolling broadcasting fallbacks in `internal/fidget/internal/convert/type.go`
    (`vecType.ExpandTrees`) and `backend/wgsl/internal/wgsl/numeric_op.go`, and flatten single-use
    cross-statement `let` chains for associative accelerated ops in `internal/lower/lower.go`.

- [ ] **S11. Resolve Residual `inline`, `eval`, `meshfold`, and `fidget` Code `TODO`s**
  - Extract the `"gsl_tmp_"` constant in `internal/assembler/inline/inline.go` (`TODO`) and remove
    the unused `currentPkgPath` parameter on `shouldInlineMultiReturnCall` in
    `internal/assembler/inline/predicates.go` (`TODO`); extract `"matrix"`/`"vector"`/`"multiply"`
    constants at both `TODO` sites in `internal/eval/const.go`; inline trivial `FoldProgram` and
    document cloning behavior on `FoldProgramWithSlots` in `internal/meshfold/fold.go` (2 `TODO`s);
    pass GSL source/location spans into `GenerateSTLAtTime` in
    `internal/fidget/internal/mesh/mesh.go` (`TODO`); and warn or check parent-scope variable
    shadowing in `internal/fidget/internal/fidget/convert.go` (`TODO`) via `internal/scope`.

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

- [ ] **S16. Tighten `.color` Property Validation & Docs to Strictly 3-Element RGB (`vec3`)**
  - Tighten `validateColorProperty` in `internal/assembler/propval/validate.go` (and unit tests in
    `internal/assembler/propval/validate_test.go`) to require `len(vals) == 3` rather than allowing
    stale 4-element `[r, g, b, a]` vectors that fail downstream in `internal/assembler/property`,
    `backend/wgsl`, and `internal/fidget`, and remove `[r, g, b, a]` from `name.color` in
    `doc/gsl_lang_spec_current.md` and `doc/gsl_lang_spec_future.md`.

- [ ] **S17. Remove `compile eval` & `compile skill` in Favor of `gsl eval` & `gsl skill` Only**
  - Remove the `os.Args[1] == "eval"` and `os.Args[1] == "skill"` branches from
    `cmd/compile/main.go` and the `compile eval` usage line from `cmd/compile/internal/cli/cli.go`
    so `cmd/compile` only compiles GSL sources; move `cmd/compile/internal/evalcmd` (plus
    `cmd/compile/render_eval_test.go`) and `cmd/compile/internal/skillcmd` into `cmd/gsl/internal/`
    for direct `gsl eval` and `gsl skill` execution (updating usage headers and
    `cmd/gsl/internal/cli/cli.go` + `cli_test.go`, and unifying static wheel linkage in
    `tool/build-wheel.sh` and `python/omega-core` so Fidget is not linked twice).

## 6. Acceptance Criteria

- [ ] Every backlog item is categorized into `Large`, `Medium`, or `Small` with concrete,
      code-verified `omega-core` context and relative paths.
- [ ] All documentation, terminology, and path references within this specification are completely
      self-contained inside `omega-core`.
- [ ] Large and Medium entries provide clear architectural boundaries and sub-goals from which
      standalone Conductor tracks (`spec.md` + `plan.md`) can be created as needed, while Small
      entries define self-contained cleanups verifiable via `./tool/check.sh` and `./tool/test.sh`.

## 7. Out of Scope

- **Native Boolean Type (`bool`):** Introducing a separate `bool` primitive type with implicit
  scalar conversion is intentionally excluded; GSL retains scalar `> 0.0` truthiness.
- **Parser Expression Depth Limit Removal:** Lifting the 127-level `PkgFileScope.IncExprDepth`
  recursion cap is excluded from the planned roadmap.
- **External Service, RPC, or Cloud Infrastructure Concerns:** Microservice dispatchers, remote
  RPC transport layers, authentication, and cloud deployment pipelines lie outside `omega-core`.
