# Implementation Plan: `gsl eval`: Answers an Agent Can Lean On

## Overview

Implements [spec.md](./spec.md), test first, one fault per task. Nothing on this track is committed yet.

## Execution Metrics & Effort Tracking

| Metric             | Elapsed Time |
| :----------------- | :----------- |
| **Spec Creation**  | `~15 mins`   |
| **Plan Creation**  | `~5 mins`    |
| **Implementation** | `~150 mins`  |
| **Code Review**    | `pending`    |
| **Effort**         | `pending`    |

---

## Phase 1: The Evaluator's Answers

- [x] Task: A proven box behind the `bounds` op (`Evaluator.Enclose`), sampled and marked where no proof is had
  - [x] `eval/enclose_test.go`, `cmd/compile/internal/evalcmd/answers_test.go`
  - [x] Best-first search per face; a point evaluated with each box says whether it straddles the surface;
        cuts are off the half so boxes do not land on faces at round coordinates
  - [x] Budget 240,000 interval evaluations; a face the budget ended is kept only within 1/64 of the span
- [x] Task: `volume` says whether it converged (`eval/volume_budget_test.go`)
- [x] Task: `eval` defaults `exprs` to the returns
- [x] Task: The sharp blend's intervals (found while timing `bounds`; owner chose the `min` form)
  - [x] `intersectRound(fields, radius)`: `min(sharp, 1e-30 - radius)`
  - [x] 58 text goldens regenerated with the owner's word: 39 example `.wgsl`, 18 codegen documents,
        1 fidget convert tree. 22 convert goldens the update also rewrote (key order only) were restored
- [x] Task: Diagnose `certify`'s null bounds on a normalized gyroid; report
  - [x] `normalizeField(v)` is `v / length(v.dp)`. A gyroid's gradient is zero at points in every cell, so
        the divisor's interval reaches 0 over any box holding one and the quotient is unbounded.
        `unionRound` then spreads that to the whole part. `certify` reports it as `budget` with null bounds
  - [ ] Not changed: `certify` could say `unbounded`; a union could keep the bounded operand's bound
- [x] Task: Reproduce `extremum`'s unconverged plate; report
  - [x] Reproduced: min -3.977 of a true -4, `minConverged: false` after 150,000 evaluations. The minimum
        is the plate's whole mid-plane, which a Lipschitz bound cannot close. It says so; not changed

## Phase 2: The Library

- [x] Task: `shell`'s bias runs as its comment says (`TestShell_BiasRunsAsDocumented`); its golden regenerated
      with the blend's
- [x] Task: The TPMS comments say what the fields are

## Phase 3: Verification and Records

- [x] Task: Full Go suite (97 packages pass), usage text, wheel README and wrapper, CHANGELOG
- [ ] Task: `tool/check.sh`; the wheel's Python tests (need a wheel build); metrics; commit
