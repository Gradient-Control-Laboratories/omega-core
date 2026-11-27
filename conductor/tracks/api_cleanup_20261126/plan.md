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

- [~] Task: Inventory omega-dm ([spec section 10](./spec.md#10-omega-dm-inventory))
  - [x] Every stdlib name and omega-core Go API this track changes that omega-dm calls
  - [ ] The functions it has copied: its running list was not found in the repository
  - [ ] For each name a goal below renames or removes: update omega-dm's copy, or pin the old name
- [x] Task: Choose the open names: `domain` (N1), `multiplyScalar` (N2), `a, b` (N10)

## Phase 1: Names That Break Nothing

- [x] Goal: Parameter names (N3, N4, N7, N10's parameter half)
  - [x] One axis-rotation vocabulary (`axisDirection` for every `axisDir` in the stdlib); `cellSize`; `min`/`max`
        box corners; `a, b, c` with `normalA/B` and `setbackA/B` for unions, intersections and the exact
        helpers. Differences keep `target`/`tool`; the solver internals and `cornerDistance` (whose body
        names its corner `c`) keep their names
  - [x] The documentation tests pass. Generated WGSL and assembled output keep parameter names, so 33 WGSL
        goldens and the `sphere_gears_text` assembled and rebuild goldens moved; a token-aligned check
        showed every change is one of the renames
- [x] Goal: Hide and reword (R2, R3, R4)
  - [x] Drop the solver internals' `///` (now `//`); reword the `planesEven` lines; remove the preamble
        `Domain` filter and `PkgInfo.Domain`. The four functions keep their variadic argument, ignored and
        documented as going, so omega-dm's calls still build
  - [x] Check: `gsl skill builtin` no longer lists the internals

## Phase 2: Stdlib Renames and Removals

- [x] Goal: Rename the `map` package to `domain` (N1), with its `gsl fix` rewrite
  - [x] A hard rename (no alias): `gsl fix -domain` rewrites references and imports; an old `map.` or
        `import "map"` gets an error naming `domain`; `rotateYZ`, `rotateZX`, `eulerAngles` documented
  - [x] Fixtures and tests that said `map.` were run through the fixer; 30 WGSL goldens moved
        `map_…` → `domain_…` only (token proof)
- [x] Goal: Honeycomb names (N5): the `rectangular` family, `hexagonal`, `triangular`, `octet`, with a rewrite
  - [x] `gsl fix -stdlib` renames members through the package prefix (table in cmd/fix renames.go); no golden
        moved (the functions inline away); three fixtures went through the fixer
- [ ] Goal: Lattices (N6, spec section 6a)
  - [ ] A `combine` form on every beam lattice, struts written out explicitly; folded forms unchanged
  - [ ] Retire `lattice.bcc` into `beam.bcc(cellSize, combine) - beamRadius`, with a rewrite
  - [ ] Each beam and foam doc names its complex and its dual
  - [ ] Kelvin foam, rhombic-dodecahedral foam and strut lattice, built from points
- [ ] Goal: `matrix.scale` becomes `multiplyScalar` (N2), with its rewrite and the lowering updated; hide the lowering targets (R1)
- [ ] Goal: Centred `torus` and `octahedron` (N9); difference operator forms (N10)
  - [ ] A test per new form; the documentation examples still build

## Phase 3: Matrices

- [ ] Task: Prototype `determinant` and `inverse` of any size in GSL; check them against the sized versions
      on 2x2, 3x3 and 4x4 and that they constant-fold
- [ ] Goal: Generic `transpose`, `multiply`, `identity(n)`, `determinant`, `inverse`; sized names as aliases
      for one release; constructors unchanged

## Phase 4: Spinodoids

- [x] Task: Settle what "scale" means in the table decision: only the box changes (spec section 6)
- [ ] Goal: `gaborSpinodoid` is hodge's field and `fastGaborSpinodoid` the lattice field, in every backend
  - [ ] A WGSL hodge field, checked against the CPU one (`hodge_parity_test.go`'s values)
  - [ ] WGSL codegen and fidget conversion stop forcing `Fast` for `gaborSpinodoid`
  - [ ] Library: `gabor*` over hodge's field, `fastGabor*` over the lattice
- [ ] Goal: A graded field's table box comes from the call
  - [ ] Remove the `@scene` and geometry fallback (`SceneDomain`) and the periodic fallback for graded fields
  - [ ] A graded call with no box is an error naming the argument

## Phase 5: Go SDK

- [ ] Goal: `With…` options (spec section 7's table); bare names marked `Deprecated` for one release; omega-core's
      callers move
- [ ] Goal: Deprecate `RenderLitSized`, `RenderLitSizedView` and `RenderLitSizedViewRadiance` for one release;
      omega-core's callers move to `RenderLit(prog, LitOptions)`; `RenderLitLinear` and `RenderLitWithLinear` stay
- [ ] Goal: `LitOptions.Device = "fidget"`
  - [ ] The cgo layer exposes fidget's CPU voxel render per return: depth, normal, return id
  - [ ] Merge, colour and material tapes at hits, SSAO, then `RenderLit`'s shading; compared with the browser's
        Fidget device on the same frame
  - [ ] `gsl compile --device fidget`; `/api/render-lit` passes `device`
  - [ ] Then retire `eval.Render` and `eval.RenderFramed`, once no repository calls them

## Phase 6: The Language Spec

- [ ] Goal: Move what the generated description lacks into it; cut
      [doc/gsl_lang_spec_current.md](../../../doc/gsl_lang_spec_current.md) to the compiler's internals
