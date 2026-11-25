# `omega-core` Roadmap

This roadmap tracks the evolution of `omega-core` and the Geometric Shape Language (GSL).
For the full, chronological record of shipped features, breaking changes, and fixes, see
[CHANGELOG.md](../CHANGELOG.md).

## Reference Documents

- [Current GSL Language Specification](gsl_lang_spec_current.md) — what the compiler implements today
- [Future GSL Language Specification](gsl_lang_spec_future.md) — target GSL v2 design
- [Changelog](../CHANGELOG.md) — completed releases and unreleased changes on `main`
- [Master Backlog Spec](../conductor/tracks/backlog_20261110/spec.md) — detailed engineering backlog
  (`L1`–`L10`, `M1`–`M18`, `S1`–`S24`)

## Phased Language Evolution Lifecycle

Breaking or surface-level GSL language changes follow a four-stage rollout:

1. **Add Capability:** Introduce the new syntax, AST representation, and semantic validation.
2. **Automate Migration (`gsl fix`):** Add a deterministic `cmd/fix` rule and migrate examples.
3. **Emit Compiler Warnings:** Report actionable `program.Diagnostic` warnings on legacy syntax.
4. **Retire Legacy Syntax:** Remove legacy grammar/desugaring shims with targeted migration errors.

---

## Recently Completed Highlights (See [CHANGELOG.md](../CHANGELOG.md))

- **Go Module & SDK Standalone Cutover:** Consolidated on
  `module github.com/Gradient-Control-Laboratories/omega-core`, added `gsl viewer`, `gsl export`,
  `gsl skill`, `gsl eval` (`shader`, `mesh`, `volume`, `bounds`, `certify`), and shared native
  library provisioning (`tool/ensure-native-libs.sh` + `.github/workflows/build-native-libs.yml`).
- **Annotations & Directives:** Shipped `@scene`, `@duration`, `@caption`, `@range` (and retired
  legacy `.min`, `.max`, `.period` properties via `gsl fix -range`), viewport/parameter annotations
  (`@point`, `@transform`, `@spline`, `@color`, `@direction`, `@frame`), and catalog annotations
  (`@material`, `@finish`, `@surface`).
- **Modeling & Standard Library:** Shipped exact `profileXY` corners, `openProfileXY`, `revolve`,
  `sweep`, `sweepOpen`, quadratic Bezier edges (`edgeBezier`), `nested`, `draft`, exact rolling-ball
  rounds (`unionRoundExact` / `intersectRoundExact`), exact chamfers, conics/quadrics, feature
  pruning (`.featureSize` / `pruneBelow`), stand-ins anywhere (`.simple`), and
  `mechanical/{shaft,bearing,tolerance}`.
- **Diagnostics & Postfix Modifiers:** Generalized left-to-right postfix modifier chaining (`S1`),
  surfaced parser/compiler warnings in `gsl viewer` (`S20`, `S21`), and enforced
  `tool/analyze-deps -rules` in `tool/check.sh` (`S6`).

---

## Planned Work

### 1. GSL Current → Future Language Spec Differences

Items bridging [gsl_lang_spec_current.md](gsl_lang_spec_current.md) to
[gsl_lang_spec_future.md](gsl_lang_spec_future.md):

- **First-Class AST Annotations & Unified UI Control Extraction (`L1`):** Store all annotations
  natively in `program.proto` instead of desugaring viewport annotations into `ui*` calls; register
  dynamic sliders for inline `@range` expressions; unify `assembler/slider` and `assembler/ui`.
- **Discrete Option Annotations (`@values`, `M1`):** Support `@values(v1, v2, ..., vn)` on scalar
  declarations, parameters, and inline literals for discrete UI selection.
- **Unified Asset Import Syntax (`L2`):** Replace numbered-slot and inline `importMesh` /
  `importStep` / `importWarp` forms with `import <name> "<asset>";` and typed asset handles.
- **Phased Syntax Retirements:**
  - Retire `uiPoint`, `uiTransform`, and `uiSpline` in favor of `@point`, `@transform`, `@spline`
    (`M3`).
  - Retire legacy `// Scene Size:` and `// Scene Grid:` header comments (`S2`).
  - Reject legacy `not(...)`, `and(...)`, `or(...)`, `atan2(...)` with migration hints (`S3`) and
    retire legacy `import . "omega"` (`S22`, `S23`) and prelude alias shims (`S4`).
- **Language & Tooling Ergonomics:**
  - Complete 3-argument `reduce(fn, collection, initial)` and retire legacy argument order (`M4`).
  - Strict lexical declaration-order enforcement in the assembler (`M6`).
  - Hexadecimal (`0x...`) and 6-digit `#RRGGBB` color literals (`S15`).
  - Canonical GSL source formatter (`cmd/fmt` / `gsl fmt`, `L9`).
  - Hand-written byte scanner and direct-to-AST parser (`L10`, scheduled last after syntax
    retirements stabilize).

### 2. Compiler, Evaluator, Backend, & Runtime Roadmap

Items tracked in [backlog_20261110](../conductor/tracks/backlog_20261110/spec.md) and active
Conductor tracks:

- **Compilation Cancellation API (`L3`):** Plumb `context.Context` through `assembler`, `eval`,
  `backend/*`, native Fidget/Manifold meshing loops, and `cmd/serve`.
- **WGSL Codegen & Raymarching Performance (`L4`, `M7`, `M8`, `M18`):** Emit GPU branches where
  ternary conditions are spatially invariant, add screen-space derivatives (`dpdx`, `dpdy`,
  `fwidth`) for viewer AA, and parameterize coordinate remaps cleanly.
- **Compiler Architecture & Derivatives (`L5`, `L6`, `L7`, `M13`, `M14`):** Single-pass vectorized
  `.dp` / Jacobian chain rule, unified looped AST lowering (`map` / `reduce` / `reduceCSG`), shared
  backend builder simplification, and typed AST cloning / pass gating.
- **Native Rust FFI Hardening (`L8`, `M10`):** Structured FFI error handling, panic safety, memory
  budgets, and re-enabling skipped Fidget conversion golden tests.
- **Modeling & Evaluator Tracks:**
  - Remaining [Profile Modeling](../conductor/tracks/profile_modeling_20261117/spec.md) operations
    and deferred Phase 7 chained positions from
    [Shaft Modeling](../conductor/tracks/shaft_modeling_20261118/index.md).
  - Evaluator soundness and budget controls (`M16`,
    [eval_answers_20261115](../conductor/tracks/eval_answers_20261115/spec.md),
    [eval_fixes_20261115](../conductor/tracks/eval_fixes_20261115/spec.md)).
  - Complete `cubicBaldwinReinforced` lattice (`M12`) and `stdlib/text` `cxf2gsl` generator (`M11`).
- **Diagnostics & Interactive Viewer (`M5`, `M9`, `S24`):** Richer cold-path error diagnostics with
  provenance chains, hybrid/mesh-only `mesh(...)` preview in `cmd/serve`, and a viewer `Fix` button
  for `gsl fix` warnings.
