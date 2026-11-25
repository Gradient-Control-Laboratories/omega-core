# Fold collapse: golden verification

Companion to the change in `common/program/eval/fold.go` that collapses each
scalar constant the fold has evaluated into its literal. Twenty-one goldens
move and one test assertion changes; this records what moved and why each
change is equivalent.

## What changed in the compiler

`eval.Fold` already evaluated every constant subtree and stamped the result in
`ConstValues`, but left the subtree in place. The backends never used it: the
build layer (`common/build/visit.go`) emits any stamped operand from its
`ConstValues` and does not descend. The subtree was dead weight that every
later pass walked and every response shipped. For Fastener Cosmetics the
assembled program was 23.4 MB and 949,466 operands, of which 1,088 read `p`;
396 constant roots held 99.3% of the bytes (catalog lookups such as
`thread.metricPitch(5)` inlined at each use, 45 copies of one 110 KB tree
evaluating to 0.8). After the change: 305 KB and 2,493 operands. Fastener
Matrix 21.1 MB → 219 KB; Driver Bits 18.0 MB → 261 KB.

Under a fold that knows the parameters, `setConst` now replaces a composite
scalar operand with `Operand_Value`, keeping its `Type`, `ConstValues` and
source span. It leaves alone: a ranged assignment's root (the literal the
slider UI rewrites in place), anything inside a UI marker's argument
(`uiPoint`, `uiTransform`, `uiSpline`, whose literals the UI binds by
operand), import calls (resolved by name later), references, vectors and
matrices, and a one-element list written out. A fold without a checker
(`meshfold`, fidget's build) only stamps, as before: its verdicts may be
cleared by the next fold, but a collapse could not be undone.

## Why goldens move at all

The backends emitted stamped constants before, so the collapse itself does
not change emitted values. Three consequences do:

- **A. Temporary renumbering.** A constant whose every use collapsed is no
  longer referenced, so `deadcode.Remove` drops it and the `gsl_tmp_N`
  temporaries after it renumber. 286 of the changed WGSL/js line pairs
  are identical once `gsl_tmp_N` is normalised.
- **B. Constants reached through an import are now literals.** The WGSL
  agent runs its own fold with a checker that treats a reference into another
  package as a parameter (`codegen.Delegate.IsParameter`), so a constant
  built from an imported constant was emitted structurally: `length(vec3<f32>(0.70710677, 0.0, 1.7071068))`
  (Aerospace and Terrain Bracket), `material_plastic[0].y` (CoreXY Printer,
  material_only), `color_oklab2rgb(... dot(...))` of constants (Material
  Library, Granite, No Displacement). The assembler's fold, whose only
  parameters are slider variables, now hands these over as literals:
  `1.847759`, `0.0`, `vec3<f32>(0.94624364, 0.9462436, 0.94624346)`. Each is
  the same expression evaluated in float64 at compile time and printed as an
  f32 literal, where before the GPU evaluated it in f32; results can differ
  in the last bit (the three components of that grey differ in the eighth
  digit where the GPU would have produced one rounding).
- **C. Dead declarations dropped.** js: `const g0 … g9` (barberpole),
  `scale`, `hi`, `low`, `root`, `third` (collection_access; `baseHz` moves
  up as the first surviving constant), `baseCoeff` (vibrating_plate). WGSL:
  lifted closure and remap functions lose parameters that captured constants
  no longer read inside them (`mechanical_drill_pointAngle118` in Drill,
  `material_wood_whiteOakData`, `material_wood_defaultLook`, `color_black` in
  Material Library and Material Test - Wood), and the matching arguments at
  their call sites.
- **D. AST snapshots.** The two `.assembled` files print the program tree, so
  constant expressions print as their values: `((2 * pi) + 0.3)` →
  `6.583185307179586`; `pi = (atan(1) * 4)` is dropped as dead.
  `unresolved.assembled` is unchanged.

Every value in categories B and D was checked against its expression:
`sqrt(0.5² + 1.7071068²) = 1.847759`, `2π + 0.3 = 6.5831853`, the ISO flank
constant `cos(0.26) = 0.96638998`, `5H/8` at pitch 0.8 `= 0.433`.

## Per file

| Golden | + | − | Renumbering-only pairs | Categories |
|---|---|---|---|---|
| `agent/assembler/internal/testdata/simple.assembled` | 1 | 2 | — | D |
| `agent/assembler/internal/testdata/sphere_gears_text.assembled` | 12 | 13 | — | D |
| `agent/js/internal/codegen/testdata/barberpole.js` | 0 | 10 | 0 | C |
| `agent/js/internal/codegen/testdata/collection_access.js` | 1 | 6 | 0 | C |
| `agent/js/internal/codegen/testdata/vibrating_plate.js` | 0 | 1 | 0 | C |
| `agent/wgsl/internal/codegen/testdata/example/Aerospace Bracket.wgsl` | 1 | 1 | 0 | B |
| `agent/wgsl/internal/codegen/testdata/example/Bolted Joint.wgsl` | 3 | 3 | 3 | A |
| `agent/wgsl/internal/codegen/testdata/example/CoreXY Printer.wgsl` | 5 | 5 | 0 | B |
| `agent/wgsl/internal/codegen/testdata/example/Drill.wgsl` | 21 | 25 | 0 | A, C |
| `agent/wgsl/internal/codegen/testdata/example/Driver Bits.wgsl` | 11 | 11 | 11 | A |
| `agent/wgsl/internal/codegen/testdata/example/Fastener Cosmetics.wgsl` | 124 | 124 | 124 | A |
| `agent/wgsl/internal/codegen/testdata/example/Fastener Matrix.wgsl` | 26 | 26 | 26 | A |
| `agent/wgsl/internal/codegen/testdata/example/Material - Library.wgsl` | 58 | 62 | 43 | A, B, C |
| `agent/wgsl/internal/codegen/testdata/example/Material - Sand Ripples.wgsl` | 5 | 5 | 5 | A |
| `agent/wgsl/internal/codegen/testdata/example/Material - Sea and River.wgsl` | 25 | 25 | 25 | A |
| `agent/wgsl/internal/codegen/testdata/example/Material Test - Granite.wgsl` | 1 | 1 | 0 | B |
| `agent/wgsl/internal/codegen/testdata/example/Material Test - No Displacement.wgsl` | 1 | 1 | 0 | B |
| `agent/wgsl/internal/codegen/testdata/example/Material Test - Wood.wgsl` | 44 | 55 | 28 | A, B, C |
| `agent/wgsl/internal/codegen/testdata/example/Terrain Bracket.wgsl` | 1 | 1 | 0 | B |
| `agent/wgsl/internal/codegen/testdata/example/Thread Relief.wgsl` | 21 | 21 | 21 | A |
| `agent/wgsl/internal/codegen/testdata/material_only.wgsl` | 1 | 1 | 0 | B |

## Test assertion

`agent/js/internal/codegen/build_import_test.go` (`TestImportConstants`) no
longer expects `const libk_pi` in the output: `two_pi = pi * 2` in the import
collapses to `6.283…`, so `pi` is dead and dropped. The assertion that
`libk_two_pi` is declared and referenced, which is what the test guards (the
import's constants reach the output and resolve), is unchanged.

## Determinism and sign-off

The goldens were regenerated twice (`UPDATE_GOLDEN=1`), from the first and
final version of the change, with byte-identical hunks in all 21 files. With
the regenerated goldens, `common/program/eval`, `agent/assembler/...`,
`agent/js/...` and `agent/wgsl/...` pass. New tests: `common/program/eval/collapse_test.go`
(collapse, vectors kept, ranged root kept, nil checker, UI marker arguments)
and `agent/assembler/lib/collapse_test.go` (no composite scalar constant
survives assembly, including one an inlined body brings in untyped).
