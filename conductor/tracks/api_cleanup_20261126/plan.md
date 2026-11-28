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
        (follow-up after #178: omega-dm has the notes from omega-core; its own list was not found)
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
- [x] Goal: Lattices (N6, spec section 6a)
  - [x] A `combine` form on every beam lattice that joins struts (bcc, fcc, octet, truncatedOctahedral,
        truncatedCubic, kelvinCell, fluorite, cubicDouble), written out from the fold domain's pieces and their
        mirror images; folded forms unchanged; a test checks each against its folded form. isotruss, reentrant
        and cubic already took one; the columns join nothing
  - [x] Deprecate `lattice.bcc` (kept for a release) into `beam.bcc(cellSize, combine) - beamRadius`, with a
        `gsl fix -stdlib` rewrite; the blend_lattice fixture went through it
  - [x] Each beam and foam doc names its complex and its dual (checked numerically: which struts lie in which
        walls, and which lie on Voronoi edges)
  - [x] foam.kelvin and foam.rhombicDodecahedral, built from points, each with a combine form; a test checks
        them against brute-force Voronoi walls. The rhombic-dodecahedral strut lattice already existed:
        beam.fluorite is it (the dual of foam.fcc, which sits on the other face-centred lattice of the cell)
  - [x] examples/lattice_duals.gsl pairs each strut lattice with its dual foam; the foam gallery shows the
        Voronoi foams
- [x] Goal: `matrix.scale` becomes `multiplyScalar` (N2), with its rewrite and the lowering updated; hide the lowering targets (R1)
  - [x] `scale` stays as an alias for a release; argument order kept `(s, m)` (`vector.multiplyScalar` is `(v, s)`);
        lower_test's five routed-name checks follow; no golden moved
- [x] Goal: Centred `torus` and `octahedron` (N9); difference operator forms (N10)
  - [x] `torus(center, majorRadius, minorRadius)`, `octahedron(center, size)`; every difference family gets a list
        form `[target, tool, …]` and an operator form; a test per form (constant-folded equality)

## Phase 3: Matrices

- [x] Task: Prototype `determinant` and `inverse` of any size in GSL; check them against the sized versions
      on 2x2, 3x3 and 4x4 and that they constant-fold (pivoted elimination; a zero pivot made the unpivoted
      form NaN)
- [x] Goal: Generic `transpose`, `multiply`, `identity(n)`, `determinant`, `inverse`; sized names as aliases
      for one release; constructors unchanged
  - [x] `list(n)` / `list(n, fn)` builtin (n constant, specialized through parameters: internal/assembler/listlen)
        so `identity(n)` can be written
  - [x] A function reducing over a list parameter with a callback that indexes by the element index is
        specialized per list size and unrolled (internal/assembler/listsize), so run-time matrices work
  - [x] The affine 4x4 inverse is `inverseAffine`; the stdlib's callers use it; sized names deprecated with
        `gsl fix -stdlib` rewrites (`identity3()` becomes `identity(3)`)

## Phase 4: Spinodoids

- [x] Task: Settle what "scale" means in the table decision: only the box changes (spec section 6)
- [ ] Goal: `gaborSpinodoid` is the reference field and `fastGaborSpinodoid` the fast lattice field, in every backend
      (follow-up after #178: needs a GPU parity check)
  - [ ] A WGSL reference field, checked against the CPU one (the reference parity test's values)
  - [ ] WGSL codegen and fidget conversion stop forcing `Fast` for `gaborSpinodoid`
  - [ ] Library: `gabor*` over the reference field, `fastGabor*` over the lattice
- [x] Goal: A graded field's table box comes from the call
  - [x] Graded forms take `box`, [min, max] in the field's own coordinates; the builtins take it as a 13th
        argument (stationary forms pass an ignored zero box). The converter bakes over it: no scene or
        geometry domain, no query-space transform, no outer-remap periodic fallback. SpinodoidBounds stays
        computed (the omega app sends it) but no longer reaches the tables
  - [x] A graded call with no usable box is an error naming box

## Phase 5: Go SDK

- [x] Goal: `With…` options (spec section 7's table); bare names marked `Deprecated` for one release; omega-core's
      callers move (names are `With` + the old name, e.g. `WithDisableDeadcodeRemoval`)
- [x] Goal: Deprecate `RenderLitSized`, `RenderLitSizedView` and `RenderLitSizedViewRadiance` for one release;
      omega-core's callers move to `RenderLit(prog, LitOptions)`; `RenderLitLinear` and `RenderLitWithLinear` stay
      (`LitView.Options(width, height, extent, time)` builds the LitOptions a view used to)
- [ ] Goal: `LitOptions.Device = "fidget"` (follow-up after #178: needs a Rust build)
  - [ ] The cgo layer exposes fidget's CPU voxel render per return: depth, normal, return id
  - [ ] Merge, colour and material tapes at hits, SSAO, then `RenderLit`'s shading; compared with the browser's
        Fidget device on the same frame
  - [ ] `gsl compile --device fidget`; `/api/render-lit` passes `device`
  - [ ] Then retire `eval.Render` and `eval.RenderFramed`, once no repository calls them

## Phase 6: The Language Spec

- [x] Goal: Move what the generated description lacks into it; cut
      [doc/gsl_lang_spec_current.md](../../../doc/gsl_lang_spec_current.md) to the compiler's internals
  - [x] Added to the description: imports, assets and scope; `@duration`, `@caption`; no recursion; the
        one multiple assignment; matrix products and comparisons; name rules and the depth limit. Fixed
        there: the precedence table (no unary `+`, no `=` operator), the component-wise claim, the pitfall
        that recommended recursion, `importMesh`'s forms. The spec is now compiler notes (grammar,
        annotation lowering, types, the specialization passes, audio, mesh tables)
