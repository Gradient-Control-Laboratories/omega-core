# Implementation Plan: Unified Surface Model

## Overview

This plan implements [spec.md](./spec.md) in the order of its section 5. The whole track is one goal
and lands as one PR. The tight-`fract` change prepared on `blake/tight-fract` is brought into this
branch as Phase 1 rather than shipped on its own.

Every phase checks four things before it is done:

1. **Unchanged by default:** a document that sets no `.tolerance` and no build, with the D2 flags off,
   compiles and evaluates exactly as before.
2. **No silent golden moves:** any `testdata/example`, WGSL, tape or mesh golden that moves is shown
   with the reason and regenerated only on the owner's go. Existing tests and their fixtures are not
   edited.
3. **CPU and viewer agree:** where a phase touches what is drawn, the viewer (WGSL) draws what the CPU
   (fidget) evaluates at the same condition. That is the owner's check.
4. **Provable:** the journal claims of spec section 4 still decide as recorded, or better.

Where a phase needs an API the spec does not fix (builtin names, the shape of `.tolerance`, where an
envelope is stored), it opens with a proposal for the owner to approve before any code is written.
New tests are added only once the owner agrees them.

## Execution Metrics & Effort Tracking

| Metric             | Elapsed Time |
| :----------------- | :----------- |
| **Spec Creation**  | `one session, 2026-11-26, with the experiments in spec section 4` |
| **Plan Creation**  | `~10 mins` |
| **Implementation** | `not started` |
| **Code Review**    | `not started` |

## Phase 0: Baseline

- [ ] Task: Build or obtain dylibs that match `origin/main`'s Rust (`marching_cubes.rs` and `tape.rs`
      have moved since the main checkout's copies), on the owner's go, since it is a Rust build
- [ ] Task: Run the `eval`, `internal/fidget/...`, `backend/...` and `cmd/...` tests on unchanged
      `origin/main`, and record what fails there: `TestRun_STL_MarchingCubesIsWatertight` failed and
      `eval` hit the 10-minute timeout on the stale dylibs, with tight `fract`
- [ ] Task: List every reader of `displacementScale()` and of the displaced distance (WGSL codegen,
      `gsl-displace.js`, fidget convert, `litrender.go`, meshers, `eval`), as Phases 2 to 4 change them
- [ ] Task: Phase Verification & Checkpoint

## Phase 1: Tight `fract`

- [ ] Task: Bring in `blake/tight-fract`: `fract` built as fidget's modulo
      ([builtin.go](../../../internal/fidget/convert/builtin.go)), and its CHANGELOG entry
- [ ] Task: Against the Phase 0 baseline, show any test or golden that moves (interval pruning in the
      meshers and the lit renderer reads the same bounds)
- [ ] Task: Phase Verification & Checkpoint

## Phase 2: Build Apart from Relief

- [ ] Task: Proposal: the builtin that scales build (working name `buildScale()`), with
      `displacementScale()` scaling relief alone, and what each is in each context (viewer, CPU
      renderer, meshing, eval)
- [ ] Task: The displaced distance as
      `under(v) - buildScale() * incBuild(v) - displacementScale() * incRelief(v)`, in
      [layers.go](../../../internal/assembler/property/layers.go) and
      [layers_height.go](../../../internal/assembler/property/layers_height.go), with build and
      relief read from the same packed surface read (one REMAP, as #162 requires)
- [ ] Task: WGSL codegen, fidget convert and `litrender.go` follow
- [ ] Task: With both scales at today's values, every golden is unchanged (check 1)
- [ ] Task: Phase Verification & Checkpoint

## Phase 3: Flags for Export and Section Mode (D2)

- [ ] Task: Mesh export: separate build and relief flags on `compile -t stl|glb|3mf`, on `gsl eval`'s
      mesh op and on the server's mesh endpoints, off by default
- [ ] Task: Section mode: separate build and relief toggles in the viewer's section view, whose
      shaders now show the field as built ([gsl-displace.js](../../../examples/viewer/render/gsl-displace.js))
- [ ] Task: A coated block exports with its volume grown by its area times its build, to the mesher's
      tolerance
- [ ] Task: Phase Verification & Checkpoint

## Phase 4: A Condition for Evaluation

- [ ] Task: Proposal: an `eval` option (`With...`) naming the condition (display, nominal, maximum
      material, minimum material; spec 2.3) and the matching request field in `gsl eval`
- [ ] Task: `certify`, `eval`, `bounds`, `volume` and `mesh` read #162's displaced distance at that
      condition, so the journal experiment no longer writes the displacement by hand
- [ ] Task: The journal claims of spec section 4, posed through the condition, decide as recorded
- [ ] Task: Phase Verification & Checkpoint

## Phase 5: Envelopes and the Peak Datum (D1)

- [ ] Task: Proposal: where a material's relief envelope is stored (an appended record row, as
      [material.gsl](../../../stdlib/material/material.gsl) allows, or an accessor computed from the
      process) and how a coat combines the envelope beneath it with its `filmKeeps` share
- [ ] Task: `machined` declares its envelopes from the process (`[-f^2 / (8 r), 0]` below the cusps for
      turned, faced and face-milled), and interval evaluation over a surface patch checks each one
- [ ] Task: Audit every library relief for its datum. Any that is not measured from its peaks moves by a
      constant to be. That moves surfaces and goldens, so each is shown first
- [ ] Task: Tag each process as measured (heights from the process: `machined`) or judged for appearance
      (wood, `filmKeeps`'s fitted coefficient)
- [ ] Task: Phase Verification & Checkpoint

## Phase 6: Spreads as Parameters (D3)

- [ ] Task: Proposal: how a document names a layer's spread (a coat's build, a machined size) as a
      ranged parameter, and how the condition reads its ends
- [ ] Task: `CertifyParams` / `CertifyOverSliders` prove a claim over every spread at once: the plated
      journal proven for builds in [4, 6] µm machined 6 µm under, and disproven machined 4 µm under
- [ ] Task: Phase Verification & Checkpoint

## Phase 7: `.tolerance` and Zones (D4, D5)

- [ ] Task: Proposal: the `.tolerance` property (its assembler pass, alongside `.material` and
      `.color`), the zone constructors (`tolerance.band(lo, hi)`, a profile `±t/2`, ISO 286 fits as
      bands on features a package knows), and the eval entry point that proves a shape's zone
- [ ] Task: The two claims of spec 2.5, generated from `.tolerance`, with a slack for measurement
      resolution, in the part's frame
- [ ] Task: The edge rule, with the UGF edge exceptions: where a field is not a true distance at a
      convex edge, the zone there is refused or computed from an exact round, never overstated
- [ ] Task: D5: remove the condition arguments from `tolerance.shaftDia` / `holeDia` and
      `bearing.seatDia` / `housingDia`; a `gsl fix -stdlib` rewrite to nominal plus `.tolerance`;
      `examples/shaft_bearings.gsl` migrated (its seats move from maximum material to nominal, 5.5 µm on
      the radius, so its goldens move and are shown first); a CHANGELOG `Breaking` entry
- [ ] Task: Phase Verification & Checkpoint

## Phase 8: `.dp` Through Material Records

- [ ] Task: Find why `.dp` fails through a material record ("collection elements must be all scalars
      or all vectors") and fix it, so the gradient of the surface as made can be taken
- [ ] Task: Phase Verification & Checkpoint

## Phase 9 (optional): Interval-Gradient Search

- [ ] Task: The interval-gradient split (spec section 4) as a `certify` option, off by default, if
      Phases 4 to 8 leave claims the octree cannot decide in budget
- [ ] Task: Phase Verification & Checkpoint

## Phase 10: Verification and Records

- [ ] Task: The journal cases (proven, disproven, undecided at the stated slack; patch and whole seat)
      and the coated-block export volume as tests, once agreed
- [ ] Task: Docs: [material_composition.md](../../../doc/material_composition.md),
      [material_library.md](../../../doc/material_library.md) and a tolerance page
- [ ] Task: CHANGELOG entries for every phase
- [ ] Task: Phase Verification & Checkpoint
