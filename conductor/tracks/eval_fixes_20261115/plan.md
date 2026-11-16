# Implementation Plan: `gsl eval`: Two Faults a Side-by-Side Run Found

## Overview

Implements [spec.md](./spec.md), test first, one fault per task.

## Execution Metrics & Effort Tracking

| Metric             | Elapsed Time |
| :----------------- | :----------- |
| **Spec Creation**  | `~10 mins`   |
| **Plan Creation**  | `~5 mins`    |
| **Implementation** | `~45 mins`   |
| **Code Review**    | `pending`    |
| **Effort**         | `pending`    |

---

## Phase 1: The Two Faults

- [x] Task: A slider read through a closure's `remap` reaches the helper as a parameter
  - [x] `eval/shader_closure_test.go` (`TestShader_SliderInsideAClosuresRemap`): fails on 0.1.2 in `glsl-es300`
  - [x] `Delegate.needSliderParam`, called where a function emits a variable (`BuildContext.EmitVar`)
  - [x] No codegen golden moves
- [x] Task: A box left loose is not answered as proven
  - [x] `eval/enclose_loose_test.go` (`TestEnclose_ABoxLeftLooseIsNotProven`): fails on 0.1.2, proven 9 mm off
  - [x] The 1/64 rule holds every face; `ReasonLoose`
  - [x] The wheel's README and wrapper name the reason

## Phase 2: Verification and Records

- [x] Task: Full Go suite passes with `go vet`; CHANGELOG
- [ ] Task: The wheel's Python tests (need a wheel build); a UI control read through a closure's `remap`
