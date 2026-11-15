# Specification: `gsl eval`: Answers an Agent Can Lean On

## 1. Why

GSL is offered to agents as a faster loop than a B-rep kernel: edit, ask the field, look. A pilot of six
agents writing ordinary parts (a plate, a bracket, a shell, a slider, two touching blocks, a lattice cube)
found answers that were wrong or uncertain without saying so. Each one costs the loop a turn, or a wrong
part. This track fixes them. It stacks on `eval_mesh_shader_20261113`.

## 2. Faults and Decisions

| # | Fault | Evidence | Decision |
| :-- | :-- | :-- | :-- |
| 1 | `bounds` samples a grid: it misses a wall thinner than a cell and can return a box smaller than the part | An L-bracket with a 6 mm wall 40 mm tall reported 8 mm tall; an 80 mm plate reported 78 mm at resolution 64 | The `bounds` op answers with a box proven by interval arithmetic. `Evaluator.Bounds` (the mesher's fit and `--bbox auto`) is unchanged |
| 2 | Some fields give no usable intervals | A normalized gyroid returns unbounded intervals; an imported mesh's are loose and slow | The op falls back to the sampled box and says `proven: false` with a `reason` |
| 3 | `volume` stops at 64,000,000 evaluations without saying the tolerance was missed | Plate, `relativeTolerance: 0.001`: bracket 2.4% wide, `ok: true` | The answer carries `converged` and, when false, `reason: "budget"`. The algorithm is unchanged |
| 4 | `eval` refuses a request without `exprs` | "eval names its exprs" | `exprs` defaults to the document's returns, as `expr` does for the other ops |
| 5 | `shell(d, thickness, bias)` runs opposite to its comment | `shell(sphere(20), 2, -1)` is a wall from r = 20 to 22 | The comment is right: the code changes. A codegen golden calls it with a bias and is not regenerated without the owner's word |
| 6 | Two comments call TPMS fields unit-normalized | A gyroid's gradient is about 1.5 at its surface; a 1.0 mm sheet measures 0.64 mm | The comments change. The fields do not |
| 7 | `certify` reports `budget` with null bounds on a normalized gyroid | `regionLower: null`, `worstBox` all zero | Diagnose and report before changing anything |
| 8 | `extremum` reported unconverged on a plate | Agent-reported only | Reproduce first |

## 3. The Proven Box

For each of the box's six faces, a best-first search over sub-boxes of the region, ordered by how far they
reach toward that face. A sub-box whose interval lower bound is at least zero holds none of the part and is
dropped. The search ends when the box reaching farthest is a leaf that could not be dropped: every point of
the part then lies on the near side of that leaf's far face. Sub-boxes are kept thin toward the face and up
to 64 times as wide across it, so a flat face perpendicular to an axis costs a few thousand evaluations, not
the square of the resolution.

- **Sound:** the box always contains the part. It can exceed it by `cellSize` on a side.
- **`found`:** false only when every sub-box was dropped, which proves the region holds none of the part.
- **Unproven:** a face whose search spends the evaluation budget ends the proof; the sampled box is
  returned with `proven: false` and `reason` (`budget`, or `unbounded` where intervals were not finite).
- **`resolution`** still names the sampled scan's grid, which also supplies `minValue` and `minAt`.

## 4. Out of Scope

- The mesher's box fit and `gsl compile --bbox auto`.
- Rescaling any TPMS field.
- `volume`'s algorithm or budget.
