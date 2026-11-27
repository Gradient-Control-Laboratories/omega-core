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

- [x] Task: Build dylibs that match `origin/main`'s Rust (2026-11-26, owner's go): `render` and
      `gpu-eval` built in this worktree, the fidget vendor submodules cloned from the main checkout's
      module repos; `libmanifold_glue` is the main checkout's, whose sources have not changed
- [x] Task: Baseline on unchanged `origin/main` (`eval`, `internal/fidget/...`, `backend/...`,
      `cmd/...`, `-timeout 30m`): 45 packages pass. `eval` takes 634 s, past `go test`'s 10-minute
      default, which is what the earlier "timeout" was. `gpuparity`'s
      `TestGpuPrune_TeethLeaveThePitchCylinder` and `TestGpuPrune_ThreadLeavesThePitchCylinder` fail on
      `origin/main` itself: on the GPU (WGSL), far from the part, the field reads the teeth or thread,
      not the pitch cylinder. The earlier `TestRun_STL_MarchingCubesIsWatertight` failure was the stale
      dylib
- [x] Task: Readers of the displacement builtins. Fidget convert sets `displacementScale()` and
      `displacementDrawn()` from one field, `dispDrawn`, which only the lit renderer's colour path sets
      to 1 ([color.go](../../../internal/fidget/convert/color.go)); WGSL codegen reads both from the
      viewer's uniforms; the stdlib's appearance switches read `displacementDrawn()` (Phase 3). The
      rest of the 50-odd files only carry `DisplacedReturns` through their passes
- [x] Task: Phase Verification & Checkpoint

## Phase 1: Tight `fract`

- [x] Task: Bring in `blake/tight-fract`: `fract` built as fidget's modulo
      ([builtin.go](../../../internal/fidget/convert/builtin.go)), and its CHANGELOG entry
- [x] Task: Against the Phase 0 baseline, show any test or golden that moves (interval pruning in the
      meshers and the lit renderer reads the same bounds). None moved (2026-11-26, on
      `blake/tight-fract` with the Phase 0 dylibs): the same 45 packages pass, the same two
      `gpuparity` tests fail, and no test wrote a file
- [x] Task: Phase Verification & Checkpoint

## Phase 2: Build Apart from Relief

- [x] Task: Proposal (agreed 2026-11-26): `buildScale()` is new and `displacementScale()` scales
      relief alone. In the viewer's march both are the displacement knob (the build follows the knob,
      as it did); in the CPU renderer both are 1; in meshing, export, the section view and eval both
      are 0 until Phases 3 and 4
- [x] Task: The displaced distance as
      `under(v) - buildScale() * incb(v) - displacementScale() * (inc(v) - incb(v))` in
      [layers.go](../../../internal/assembler/property/layers.go). Each coat's build is read beside
      its height, and reads at a surface point pack four to a remap, so a coat's pair shares one and
      *Material - Layered Paint* keeps its two remaps. The layer switch's thickness and "drawn" tests
      read both scales
- [x] Task: WGSL codegen and fidget convert lower `buildScale()` as they lower `displacementScale()`.
      The whole-return split (the renderers' own `d - h` for a return with no layers:
      `gsl-displace.js`, `litrender.go`) moved to Phase 3, where the scales first differ (owner's
      choice)
- [x] Task: With both scales at today's values, every value is unchanged. Two WGSL goldens change in
      text only, regenerated on the owner's go: *Material - Layered Paint* (+2,401 B) and
      *Material - Machined* (+5,165 B). `TestSanityCheck`'s exception list gains `buildScale`, the
      one-line test edit #162 made for `displacementScale` (owner's go). The full suite otherwise
      matches the Phase 0 baseline
- [x] Task: Phase Verification & Checkpoint

## Phase 3: Flags for Export and Section Mode (D2)

- [x] Task: The whole-return split, moved from Phase 2. Go side: a return with no layers moves by
      `build * b + relief * (h - b)` (`convert.DisplacedTree`), which the mesher and `litrender.go`'s
      `litDisplaced` share; with the scales equal it is `d - s * h` exactly. Viewer: `gsl_heights`
      reads each return's volume row once and returns (height, build), and `gsl_sdShapeDisplaced` moves
      relief by its relief scale and build by its build scale (owner's choice over gating on evidence of
      a build, which missed *Material - Coated*'s builds through `coating.levelled`)
- [x] Task: Mesh export (proposal agreed 2026-11-26): `--build` and `--relief` on
      `compile -t stl|glb|3mf`, `"build"` and `"relief"` on `gsl eval`'s mesh op and the server's mesh
      request, `mesh.WithDisplacement`, off by default. `convert.ToDisplacedTreesAtTime` builds each
      return's layered distance with `buildScale()` and `displacementScale()` at the flags. A box worked
      out for the part (`.box`, `--bbox auto`, `@scene`) is grown `mesh.DisplacementMargin`, 3 mm (owner's
      choice): those boxes hold the field as built, and MDC left the coated stick open where its coats
      crossed the box. A box named outright is meshed as given. `gsl eval` meshes displaced take the
      facets' normals. Relief is shipped documented (owner's choice): finer than a cell it aliases, and on
      the oak's noisy relief MDC throws vertices about 11 mm off the surface; `--build --relief` on the
      whole stick at subdivision 7 did not finish in 45 CPU-minutes. Both are open, with the MDC work (#170)
- [x] Task: Section mode: Build and Relief checkboxes in the Section panel, off by default. The
      field plane draws `gsl_sectionShape(p, time, relief, build)` ([gsl-displace.js](../../../examples/viewer/render/gsl-displace.js)),
      sdShape moved by the toggles through the per-return composite or the winner's material, with no
      Detail fade. `buildScale()` has its own WGSL var (`gsl_buildScaleK`); the march passes its knob to
      both scales. Sixteen WGSL goldens change in text, 0.5 to 3.6% (owner's go to regenerate); no
      program changes composite. Checked in the viewer by the owner (2026-11-26)
- [x] Task: `displacementDrawn()` follows the relief flag (owner's choice): convert keeps it apart from
      `buildScale()` and `displacementScale()`, and a displaced mesh sets it to the relief flag
- [x] Task: New tests (owner's go), `backend/mesh/displacement_test.go`: a 20 mm cube with a 0.1 mm
      build, laid as a coat and as its own material, meshes to the offset cube's volume (8241.94 mm³
      against 8241.89) with `--build` and to the bare cube's with relief alone; with both flags off the
      mesh is byte for byte the one meshed without the option. The coated oak stick: flags off byte-
      identical to before; `--build` closed, 327,050 mm³ against 317,136 bare
- [x] Task: Phase Verification & Checkpoint

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
