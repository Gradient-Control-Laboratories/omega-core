# Implementation Plan: GSL API Cleanup

## Overview

Implements [spec.md](./spec.md). Each goal below is one pull request from main, test first. A goal that renames
or removes a stdlib name carries its `gsl fix` rewrite and says how omega-dm follows. A goal that moves an example
golden stops for approval before the golden is updated.

## Execution Metrics & Effort Tracking

| Metric             | Elapsed Time |
| :----------------- | :----------- |
| **Spec Creation**  | `pending`    |
| **Plan Creation**  | `pending`    |
| **Implementation** | `pending`    |
| **Code Review**    | `pending`    |
| **Effort**         | `pending`    |

---

## Phase 0: Before Anything Breaks

- [ ] Task: Inventory omega-dm
  - [ ] Every stdlib name omega-dm calls, and every function it has copied, from its running list and its sources
  - [ ] For each name a goal below renames or removes: update omega-dm's copy, or pin the old name
- [ ] Task: Choose the open names (spec section 9): the `map` package's, `matrix.scale`'s, the booleans' parameter

## Phase 1: Names That Break Nothing

- [ ] Goal: Parameter names (N3, N4, N7, N10's parameter half)
  - [ ] One axis-rotation vocabulary; `cellSize`; `min`/`max` box corners; one boolean parameter name
  - [ ] `TestPackageDocsListEveryDocumentedMember` and the documentation test still pass; no golden moves
- [ ] Goal: Hide and reword (R2, R3, R4)
  - [ ] Drop the solver internals' `///`; reword the `planesEven` lines; remove the preamble `Domain` filter
  - [ ] Check: `gsl skill builtin` no longer lists the internals

## Phase 2: Stdlib Renames and Removals

- [ ] Goal: The `map` package's new name (N1), with its `gsl fix` rewrite
- [ ] Goal: Honeycomb names (N5) and the three `bcc`s (N6), with rewrites
- [ ] Goal: `matrix.scale` (N2) and the vector helpers (R1), with rewrites
- [ ] Goal: Centred `torus` and `octahedron` (N9); difference operator forms (N10)
  - [ ] A test per new form; the documentation examples still build

## Phase 3: Matrices

- [ ] Task: Prototype `determinant` and `inverse` of any size in GSL; check them against the sized versions
      on 2x2, 3x3 and 4x4 and that they constant-fold
- [ ] Goal: Generic `transpose`, `multiply`, `identity(n)`, `determinant`, `inverse`; sized names as aliases
      for one release; constructors unchanged

## Phase 4: Spinodoids

- [ ] Task: Settle what "scale" means in the table decision (spec section 9)
- [ ] Goal: `gaborSpinodoid` is hodge's field and `fastGaborSpinodoid` the lattice field, in every backend
  - [ ] A WGSL hodge field, checked against the CPU one (`hodge_parity_test.go`'s values)
  - [ ] WGSL codegen and fidget conversion stop forcing `Fast` for `gaborSpinodoid`
  - [ ] Library: `gabor*` over hodge's field, `fastGabor*` over the lattice
- [ ] Goal: A graded field's table box and scale come from the call
  - [ ] Remove the `@scene` and geometry fallback (`SceneDomain`) and the periodic fallback for graded fields
  - [ ] A graded call with no box is an error naming the argument

## Phase 5: Go SDK

- [ ] Goal: `With…` options (spec section 7's table); bare names marked `Deprecated` for one release; omega-core's
      callers move
- [ ] Goal: `LitOptions.Device = "fidget"`
  - [ ] The cgo layer exposes fidget's CPU voxel render per return: depth, normal, return id
  - [ ] Merge, colour and material tapes at hits, SSAO, then `RenderLit`'s shading; compared with the browser's
        Fidget device on the same frame
  - [ ] `gsl compile --device fidget`; `/api/render-lit` passes `device`
  - [ ] Then retire `eval.Render` and `eval.RenderFramed`, once no repository calls them

## Phase 6: The Language Spec

- [ ] Goal: Move what the generated description lacks into it; cut
      [doc/gsl_lang_spec_current.md](../../../doc/gsl_lang_spec_current.md) to the compiler's internals
