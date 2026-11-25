# Golden verification — constant-select specialization

*omega · `common/program/constspec`, run in the assembler's optimization loop after each
constant fold. Written 2026-10-22.*

## Summary

The pass clones a function for a call that passes scalar constants for every parameter a
select condition in it reads, when that condition then reads nothing that varies — not `p`,
not time, not the pixel size, not any call other than pure arithmetic. The clone binds those
parameters as locals; the next fold prunes the branch the constants rule out.

Ten WGSL example goldens move; the other 79 rendered examples (`TestBuildExamples`, audio
examples skipped as before) are byte-identical. No other golden in any suite moves.

| Category | Goldens | Status | Notes |
|---|---|---|---|
| Select pruned in a clone | 8 | Identical semantics | each pruned branch is the one the old code always took at that call |
| Constant folding downstream | 1 (Granite) | Identical values | colour conversions of constants folded at build time |
| Temporaries renumbered | 1 (Sand Ripples) | Identical | `gsl_tmp_N` numbering only |

## Select pruned in a clone

Each clone's parameter was a literal at its only call; the select read nothing else that
varies, so the old code evaluated the same branch at every pixel.

- **`wood_poreSlopes(…, standIn)`** — Knotty Planks, Spalted and Wormy, Wood Finishes, Wood
  Planks. Called with `0` (`wood.gsl:2432`) and `1` (`:2434`). `standIn > 0.5` is false in
  `__spec1` (`cover = sizeShare.y * y * y`) and true in `__spec2` (`0.25` folded straight
  into `hidden`), the old select's two branches; `wood_hiddenShare` receives the same
  literal.
- **`stone_veinedMarble(…, sheets, …)`** — Marble, No Displacement, Library. `sheets = 0`:
  `cracked = crackGate`, `wall = crackWall`; the sheet wall, computed and then discarded by
  the select, is no longer computed.
- **`wood_solidWood(…)`** — Library. `knotSpacing = 0` keeps `wood_stemGrowthAt`;
  `bandStrength = 0`, `streak = 0` and `resinCanals = 0` make the band, streak and resin
  terms zero, which their selects already returned; `poreSize = 0.5` folds
  `max(poreSize, 0.01)` to `0.5` and the pore fade's `0.56 * 0.5` to `0.28` and `0.14`.
- **`wood_poreSizeShare(progress, ringPorosity, poreZone)`** — Wood. `ringPorosity = 1`:
  `ringPorosity < 0.5` is false, so `size = semiSize + (ringSize - semiSize) * (2·1 − 1)`
  and the share `0.35 + 0.45 · max(2·1 − 1, 0) = 0.8`.

## Constant folding downstream — Material Test - Granite

One line changes: a colour mix whose endpoints are constant colours is now folded. The old
line computed the Oklab conversion of constant greys at runtime, e.g.
`L = dot([0.2105, 0.7936, −0.0041], cbrt(dot(M₁, [0.21404114]³)))`; for a grey the rows of
M₁ sum to 1, so `L = cbrt(0.21404114) = 0.5981807`, the constant the new line carries.

## Determinism

`TestBuildExamples` run twice gives the same output; clone names are assigned in a fixed
walk order (`__spec1`, `__spec2`, … per function).

## Tests

- `agent/assembler/lib/constspec_test.go`: a select on a constant argument leaves no select
  after assembly (fails without the pass); a condition reading `pixelSize()` gets no clone.
