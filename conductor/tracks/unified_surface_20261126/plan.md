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
- [x] Task: Mesh export (proposal agreed 2026-11-26): `--coatings` and `--relief` on
      `compile -t stl|glb|3mf`, `"coatings"` and `"relief"` on `gsl eval`'s mesh op and the server's mesh
      request, `mesh.WithDisplacement`, off by default (named `--build` / `"build"` until the owner's
      rename to Coatings, 2026-11-27; `buildScale()` and the material's `build` keep their names). `convert.ToDisplacedTreesAtTime` builds each
      return's layered distance with `buildScale()` and `displacementScale()` at the flags. A box worked
      out for the part (`.box`, `--bbox auto`, `@scene`) is grown `mesh.DisplacementMargin`, 3 mm (owner's
      choice): those boxes hold the field as built, and MDC left the coated stick open where its coats
      crossed the box. A box named outright is meshed as given. `gsl eval` meshes displaced take the
      facets' normals. Relief is shipped documented (owner's choice): finer than a cell it aliases, and on
      the oak's noisy relief MDC throws vertices about 11 mm off the surface; `--coatings --relief` on the
      whole stick at subdivision 7 did not finish in 45 CPU-minutes. Both are open, with the MDC work (#170)
- [x] Task: Section mode: Coatings and Relief checkboxes in the Section panel, off by default. The
      field plane draws `gsl_sectionShape(p, time, relief, build)` ([gsl-displace.js](../../../examples/viewer/render/gsl-displace.js)),
      sdShape moved by the toggles through the per-return composite or the winner's material, with no
      Detail fade. `buildScale()` has its own WGSL var (`gsl_buildScaleK`); the march passes its knob to
      both scales. Sixteen WGSL goldens change in text, 0.5 to 3.6% (owner's go to regenerate); no
      program changes composite. Checked in the viewer by the owner (2026-11-26)
- [x] Task: `displacementDrawn()` follows the relief flag (owner's choice): convert keeps it apart from
      `buildScale()` and `displacementScale()`, and a displaced mesh sets it to the relief flag
- [x] Task: New tests (owner's go), `backend/mesh/displacement_test.go`: a 20 mm cube with a 0.1 mm
      build, laid as a coat and as its own material, meshes to the offset cube's volume (8241.94 mm³
      against 8241.89) with `--coatings` and to the bare cube's with relief alone; with both flags off the
      mesh is byte for byte the one meshed without the option. The coated oak stick: flags off byte-
      identical to before; `--coatings` closed, 327,050 mm³ against 317,136 bare
- [x] Task: Viewer follow-ups (owner's requests, 2026-11-26/27). Section panel: Direction, Offset,
      Model (the old Section slider), then Color | Coatings | Relief, then the contour controls, which
      Color disables. Color draws the cut faces in the material's colour (sRGB decoded, so not encoded
      twice). The Model opacity slider is gone in both modes: while the panel is open the field is
      clipped at the plane (`sdShape` wraps `gsl_sdShapeUnclipped`, clip state set at fragment entry)
      and the primary ray starts at the plane, so grazing rays no longer crawl through clipped solids
      (the "ghost objects"). Color samples each visible shape alone and cuts wherever any is solid:
      `gsl_min` keeps the nearest crossing, not the minimum, so a screw in a tapped hole dropped its
      cut. The Export panel has Coatings and Relief checkboxes for the server meshers, and grows the
      measured box 3 mm. Open: the Simple and path-traced tiers do not clip
- [x] Task: Phase Verification & Checkpoint

## Phase 4: A Condition for Evaluation

- [x] Task: Proposal (agreed 2026-11-26, not yet implemented): named conditions on the evaluator,
      `e.SetCondition(c)` (the `SetCertifyBudget` pattern), and a `"condition"` field on `gsl eval`
      requests: `as-built` (default: build 0, relief 0, today's behaviour), `nominal` (build 1,
      relief 0) and `as-made` (build 1, relief 1, what the CPU renderer draws). Maximum and minimum
      material are refused with an error until Phases 5 and 6 give them envelopes and spreads. Every
      query that reads the field takes it: `eval`, `extremum`, `certify` (point, params, sliders),
      `bounds`, `volume`, `surface` and `mesh` (onto Phase 3's flags); `render`, `describe` and `shader`
      do not. Queries are assembled with their expressions as returns, so #162's lowering already moves
      only the shapes inside a claim (`max(rMax - r, seat)` moves `seat`); a query that is one bare
      shape moves as a whole (`convert.DisplacedTree`). The interval, gradient and parameter evaluators
      each get a displaced variant through one helper
- [x] Task: `certify`, `eval`, `bounds`, `volume` and `mesh` read #162's displaced distance at that
      condition, so the journal experiment no longer writes the displacement by hand (2026-11-27).
      `eval.Condition`, `ParseCondition`, `Evaluator.SetCondition`; `gsl eval`'s `"condition"`, set per
      request on the session's evaluator, refused (exit 2) on describe, render and shader.
      `convert.ToSurfaceTrees` is the one helper: a colour/material build with `recordInTree` and both
      scales at the condition, each scalar return `DisplacedTree`, vector returns grouped as
      `ToTreeGroups`, params bound as `ToTreesWithParams`; `eval`'s `scalarTrees` / `groupTrees` /
      `paramTrees` take it only when the condition moves a surface, so as-built is the old path
      exactly. Enclose takes the condition too (it reads the field; `gsl eval`'s bounds op is
      Enclose). Mesh ORs the condition into its Coatings and Relief flags. `displacementDrawn()` is
      the relief scale, as for a displaced mesh. Found and fixed on the owner's go: a coated shape
      returned bare beside a return that layers it was moved by its layer's shading switch (height 0
      within 1 µm of the solid beneath and inside it), leaving a positive shell inside the coat in
      queries, displaced meshes and the viewer; `analyzeLayers` now gives such a return its layered
      distance (second pass over the whole returns; no golden moves). New tests (owner's go):
      `eval/condition_test.go`, `cmd/compile/internal/evalcmd/condition_test.go`
- [x] Task: The journal claims of spec section 4, posed through the condition, decide as recorded
      (2026-11-27; `as-made`, the seat carrying `machined.turned`, plating a coat laid with
      `material.withBuild(fMat, build)`): A proven 169 (as recorded), B disproven 5,305 (5,145),
      no-slack envelope undecided (as recorded), 0.1 µm slack proven 5,257, C disproven 2,217 (as
      recorded), D proven 5,257 (the hand-written field's count on this build), D's least-material
      claim proven 169, whole seat in the profile plane proven 43,689 (as recorded). D was first
      undecided at the 200 k budget: a coat laid on a shape with its own finish was moved by the
      increment, `disp(seat) - (h(coat) - h(seat))`, which reads the relief once at p and twice at the
      closest point; the three cancel in value but interval arithmetic takes them as independent (the
      drawn height's interval 12.8 µm wide over a box where the true range is 3.1 µm). Fixed on the
      owner's choice: such a chain telescopes as an alias chain on an undisplaced shape already did
      (`finishRoot`): the coat is the finished shape's own distance less the top coat's height, read
      once at its closest point. Values unchanged at every point checked, on and off the surface; no
      golden moves. Test (owner's go): `TestCondition_PlatingOverTurningProves`
- [ ] Task: Phase Verification & Checkpoint

## Phase 5: Envelopes and the Peak Datum (D1)

- [x] Task: Proposal (agreed 2026-11-27, not yet implemented). By D1 every relief is measured down from
      its peaks, so its outer bound is 0 and the envelope is one number, its depth (>= 0). The depth
      lives in row 0's fourth component, today "reserved" (default 0, clamped to >= 0, blended with
      the row): `material.withReliefDepth(m, depth)` and `material.reliefDepth(m)`. The record is not
      widened, and a record that sets no depth is unchanged. A coat's depth is its `filmKeeps` share
      of the depth beneath plus its own; `withBuild` keeps the depth of the material it is built
      from. `max-material` and `min-material` are switched on in this phase, builds at nominal:
      maximum material is the builds with the relief at 0 (no envelope needed), minimum material the
      builds with the relief at its depth; Phase 6 moves each build to the end of its spread. The
      measured/judged tag is documentation: a `///` line on each process and a table in
      [material_composition.md](../../../doc/material_composition.md)
- [x] Task: `machined` declares its envelopes from the process (`[-f^2 / (8 r), 0]` below the cusps for
      turned, faced and face-milled), and interval evaluation over a surface patch checks each one
      (2026-11-27). `material.reliefDepth` / `withReliefDepth` (row 0's fourth component);
      `grooveFinish` (turned, faced, face-milled, ball-end) declares the exact groove depth,
      `r - sqrt(r^2 - f^2 / 4)` (`grooveDepth`; f^2 / (8 r) falls a little short, so it would not
      bound the relief), added to any depth beneath, on the finish and its stand-in;
      `coating.levelled` declares the coat's depth plus `keeps` of the depth beneath. Checked over
      the whole 15 mm journal seat: relief <= 0 proven in 1 interval evaluation, relief >= -depth
      proven in 1 with 0.1 µm slack (undecided without, touching at every groove bottom). Three WGSL
      goldens move (owner's go): Material - Coated +0.73%, Layered Paint +0.29%, Machined +0.52%.
- [x] Task: The material conditions (decided 2026-11-27, revising the proposal above): `max-material`
      is switched on, builds on and relief at its peaks, the same surface as `nominal` until Phase 6's
      spreads. `min-material` stays refused as a field condition: a feature's depth is a constant, so
      a least-material claim is its band shifted by `material.reliefDepth`, which Phase 7's generated
      claims do; a claim across surfaces of different depths writes the shift itself. That keeps a
      `reliefDepthScale()` builtin, a TestSanityCheck exception and ~16 goldens out. Two of this
      branch's tests updated to match (owner's go)
- [x] Task: Audit every library relief for its datum. Any that is not measured from its peaks moves by a
      constant to be. That moves surfaces and goldens, so each is shown first. Audit done
      2026-11-27: [datum_audit.md](datum_audit.md). Decided: the process and coat families (cast,
      printed, molded, sheet, coating) move, one family at a time, each declaring its depth; bulk and
      collection materials stay judged appearance, documented as not measured from their peaks. Fix
      in this track: replace-vs-add (shaft.body, fastener male/femaleFinish, @rubberGrip),
      chippedPaint's blended build, the fabric stand-in, the documented amplitudes. Done
      2026-11-27, each relief and its stand-in shifted by the same constant and declaring its depth:
      sheet (checkerPlate 1.2, stuccoEmbossed 0.2; no golden), coats (powderCoat by its sound noise
      bound 0.009 peel, wrinkle 0.0351 size, sand 0.4 grain, hammertone 0.02532 cell; Coated golden),
      cast (sandCast by Rpm 3.1 / GFN, depth Rz 5.6 / GFN; shotPeened declares 0.08 dimple; no golden),
      printed (fdm 0.107301 L, resin 0.04 L, powderBed 0.24 grain, all exact; `layerFinish` takes
      crest and depth; Printed golden), molded (leatherGrain 0.5 depth, exact; no golden). Stochastic
      reliefs (decided): the peak is a profilometer's, Rpm over five sampling lengths (ISO 21920,
      sampling length by ISO 4288), measured with a scratchpad Go tool over twelve profiles, since a
      random surface has no highest point (sandCast's noise bound is 21 / GFN against Rpm 3.1 / GFN).
      Fixes: shaft and fastener finishes add their heights; @rubberGrip's dots sit on its 2 mm build,
      from their tops; chippedPaint keeps the base's build and is measured from the paint's top
      (-0.05 paintGone - 0.03 primerGone; stand-in their mean); the fabric stand-in goes through the
      same resin; water, lava, cloud, reeded and orangePeel docs give sampled amplitudes. Goldens
      (owner's go): Detail Ladder +0.29%, Fastener Cosmetics +0.16%, Coated +0.27%. Threads and knurls
      declare no depth: their tolerance is the pitch diameter (ISO 965), for Phase 7
- [x] Task: Tag each process as measured (heights from the process: `machined`) or judged for appearance
      (wood, `filmKeeps`'s fitted coefficient). Done 2026-11-27: a `/// Relief:` line on each of the
      fifteen process and coat functions, and a table in
      [material_composition.md](../../../doc/material_composition.md)
- [x] Task: Phase Verification & Checkpoint (2026-11-27): backend, internal/assembler, cmd, eval,
      stdlib, program, parser and assembler pass, 58 packages (eval 655 s); lint clean. The stdlib
      protos are regenerated (`./tool/build-stdlib.sh --force`), stale since `79a1e616`. `shaft.body`
      writes its add as `withDisplacement(finish, displacement(finish) + relief)`, the same value as
      `addDisplacement`, which keeps `TestShaftKnurl_FinishedBody`'s check on the shader text as it is

## Phase 6: Tolerances on What Is Made (D3, D8)

- [x] Task: Proposal (agreed and built 2026-11-27): `x.tolerance = tolerance.limits(lo, hi)` (or
      `tolerance.symmetric(w)`, both in the `tolerance` package) on a scalar declaration, offsets from
      its value, limits constant or a field of p; registered as an assignable property
      (`propval`, `scope`) and readable as `.box` is: the linker reads `x.tolerance` as the expression
      assigned to it, so `x.tolerance` reads `[lo, hi]` and `x + x.tolerance[0]` the least value.
      `eval.TolerancedNames`; `CertifyOverTolerances` and `gsl eval`'s `"overTolerances"` (alone, not
      with `overSliders` or `params`): each toleranced declaration is rewritten in place as
      `value + lo + (hi - lo) * u / w`, `u` a slider over `[0, w]`, `w` the tolerance's width at the
      region's centre, and certify ranges the units with `CertifyParams`; results name each
      tolerance's failing share, 0 at its low limit to 1 at its high. Sound for limits that vary
      over the part (pointwise claims), and for declarations that are not literals. A first cut ranged
      `u` over `[0, 1]`: certify resolves a parameter by its leverage and never coarsens one below 1,
      so a unit whose leverage was a micrometre-wide tolerance split ten thousand times finer than
      needed and the journal went undecided; scaling the unit by the width fixed it
- [x] Task: The plated journal and the H7/g6 plated bore, posed with `.tolerance` on their working
      dimensions: certify proves and disproves as the ranged-slider runs did, count for count (journal
      5 µm under disproven in 93,100, zinc share 0.875 to 0.94, 5.75 to 5.88 µm; 6.5 µm under proven in
      272,468; bore clearance at 3.2 µm disproven in 213,126 at boreR's low limit, zinc's and shaftR's
      high). The parser's list of assignable properties gains `tolerance`, and its fixture
      (`parser2/testdata/property.gsl`) with it (owner's go). Tests (owner's go): eval's
      `TestTolerance_ReadsItsLimits`, `_CertifyFindsTheWorstCorner`, `_CertifyProvesThePlatedJournal`,
      `_CertifyNeedsATolerance`; evalcmd's `TestCertify_OverTolerances`
- [ ] Task: Phase Verification & Checkpoint

## Phase 7: Propagation, Display and Requirements (D4, D5, D8)

- [x] Task: Proposal (agreed and built 2026-11-27): a value that declares no tolerance reads its own,
      `y.tolerance = [bias - spread, bias + spread]`, offsets of its value, bias the sum of
      `y.d(x) * (lo + hi) / 2` and spread the sum of `abs(y.d(x)) * (hi - lo) / 2` over the
      package's own toleranced declarations (owner's choices: offsets of the value, so `D + lo` is a
      shape's most-material surface; the document's own sources). The linker rewrites the read into
      that expression of DERIVs (`internal/link/tolerance.go`), expanded as every derivative is, so
      in a query it is taken at the condition. On the plated H7 bore: `boreR - zinc` reads ±4.5 µm;
      the plated block ±4.5 µm on its bore and ±1.5 µm outside as made, ±3 µm and 0 as built; the
      shaft ±1.6 µm; nothing toleranced reads [0, 0]. Test (owner's go): `TestTolerance_Propagates`
- [x] Task: The viewer draws a shape's min and max surfaces (D minus its tolerance's max and min); the
      Section panel gets a Min / Max toggle beside Coatings and Relief. Built 2026-11-27 as the owner
      chose: Render › Detail › Tolerance, Most material / Nominal / Least material, with an Exaggerate
      slider (x1 to x1000, log), and a Band checkbox in the Section panel. Program gains
      `toleranceReturns` (proto field 15, regenerated), lowered by `LowerReturnProperties` only for a
      document that declares a tolerance (`property/tolerance.go`, the returns' propagated tolerance
      as built: coats not counted, the owner's choice) and carried through every pass that walks the
      return slices (clone, refs, link, types, inline, unroll, topofold, warpfold, deadcode, lower,
      meshfold, meshslots, textures, auto-import, eval fold, variation). WGSL emits
      `gsl_toleranceField(p, time, idx)` (`codegen/tolerance.go`, `build.ToleranceReturnsBuilder`).
      The viewer offsets the winning return's distance by `lo` or `hi` times the exaggeration in the
      clipped `sdShape` wrapper and, for the per-return displaced composite, `gsl_tolMoved`; every
      tier (Simple, Color, Standard, Progressive) sets the view from `surfaceKnobs.w`, the sign the
      side and the size the exaggeration. The band shades, in the cut, where each shape's surface
      may lie between its most and least material, at least a pixel wide (tolerances are
      micrometres), its width exaggerated as the view. The CPU device (Fidget only) does not take
      the view. No existing golden moves. Checked in the viewer by the owner. Test (owner's go):
      `TestToleranceField_OnlyForTolerancedDocuments`
- [x] Task: Requirements as claims: ISO 286 and profile constructors (`tolerance.shaftDia` and
      `holeDia` as bands) used inside a certify claim, with a slack for measurement resolution; the
      least-material claim's band shifted by the surface's `material.reliefDepth`. Built 2026-11-27
      (owner's choice: library helpers): `tolerance.mostMaterialClaim(made, reference, hi)`,
      `leastMaterialClaim(made, reference, lo)`, `shaftBand` / `holeBand` (ISO 286 zones as bands
      about a cylinder's radial reference field). A first cut returned both claims from one
      `conform` call; a destructured declaration is not moved by its coats (the layer pass skips
      destructured names), so claims read as built, C proving; each claim is now its own call. The
      journal through the helpers, as made: A proven 169, envelope proven 5,257, C disproven 2,217,
      D proven 5,257, as recorded. Tests (owner's go): `TestRequirement_PlatedJournal`,
      `TestRequirement_Bands`
- [x] Task: The edge rule, with the UGF edge exceptions: where a field is not a true distance at a
      convex edge, a requirement there is refused or computed from an exact round, never overstated.
      Owner's choice: a unit-gradient check, `tolerance.gradientClaim(reference, eps)` certified over
      the region (the journal's patch: proven in 4,681); its doc says a max of planes keeps unit
      gradient to its edges, where a band is overstated, so a band near a convex edge is read from
      an exact round
- [x] Task: D5: remove the condition arguments from `tolerance.shaftDia` / `holeDia` and
      `bearing.seatDia` / `housingDia`; callers draw the size they make, its `.tolerance` on the
      scalar; a `gsl fix -stdlib` rewrite; `examples/shaft_bearings.gsl` migrated (goldens shown
      first); a CHANGELOG `Breaking` entry. Done 2026-11-27 (owner's choices): the two-argument forms
      give the zone's middle, with `shaftZone` / `holeZone` / `seatZone` / `housingZone` as the working
      tolerance about it; `gsl fix -stdlib` rewrites an old call to its exact size, nominal plus the
      condition's deviation (`cmd/fix/internal/fix/tolerance_args.go`), and reports a condition it
      cannot read; the example seats at the zone's middle with `.tolerance`, no golden covering it;
      the tests that used the old forms (`TestBearing_SeatsInAZone`, the journal in
      `condition_test.go`) rewritten to the fixer's exact equivalents, same values (owner's go).
      `../omega-dm` read (owner's go): it has no copies of these functions. Tests (owner's go):
      `TestStdlibFix_ConditionArgumentsBecomeDeviations`, `TestStdlibFix_UnreadableConditionIsReported`
- [ ] Task: Phase Verification & Checkpoint

## Phase 8: `.dp` Through Material Records

- [x] Task: Find why `.dp` fails through a material record ("collection elements must be all scalars
      or all vectors") and fix it, so the gradient of the surface as made can be taken (2026-11-27,
      brought forward ahead of Phases 6-7 on the owner's choice, since tolerance propagation rests on
      it). Cause: a constant row of a record differentiates to the scalar 0 beside vector tangents,
      and the collection mixes them; `derivDomain.Collect` now writes a zero tangent as 0 times its
      element where tangents are mixed. `material.displacement(m).dp` gives 0.001 cos(0.5) for
      0.001 sin(z) at z = 0.5. No golden moves
- [x] Task: Derivatives of the surface as made (added 2026-11-27, owner's choice: queries only).
      `.d(var)` and `.dp` were expanded against each declaration's plain body before the layer pass
      built the coats, so `plated.d(zinc)` was 0 at every condition. `assembler.WithConditionDerivatives`
      (set by eval's queries) keeps a return's DERIV unexpanded (`ExpandSharedKeepingReturnDerivatives`);
      the layer pass carries it into the displaced return (`DERIV(gsl__disp__plated, zinc)`), and
      `LowerReturnProperties` expands the plain returns' before colour and material and the
      displaced returns' after the layers attach. Convert reads a vector displaced return (a `.dp`).
      In-document uses (exact rounds) keep the as-built gradient. On the plated H7 bore: `.d(zinc)` -1
      on every face as made, 0 as built; `.d(boreR)` 1 on the bore and 0 elsewhere. No golden moves.
      The deferred expansion shares nodes as ExpandShared does (a first cut did not, and the
      gradient of eight nested remaps stopped compiling: `TestAPropertyOfADeeplyRemappedField...`,
      back to 72 to 90 ms). Tests (owner's go): `TestCondition_DerivativesSeeTheCoats`,
      `TestDerivative_ThroughAMaterialRecord`
- [x] Task: Phase Verification & Checkpoint (2026-11-27): backend, internal/assembler, cmd, eval,
      stdlib, program, parser, assembler and internal/fidget pass but for gpuparity's two tests that
      fail on origin/main itself; eval rerun after the sharing fix 596 s; lint clean

## Phase 9 (optional): Interval-Gradient Search

- [ ] Task: The interval-gradient split (spec section 4) as a `certify` option, off by default, if
      Phases 4 to 8 leave claims the octree cannot decide in budget
- [ ] Task: Phase Verification & Checkpoint

## Phase 10: Verification and Records

- [x] Task: The journal cases (proven, disproven, undecided at the stated slack; patch and whole seat)
      and the coated-block export volume as tests, once agreed (owner's go, 2026-11-27):
      `TestJournal_Cases` (fine turning least material proven, coarse disproven, the envelope with no
      slack undecided, the whole seat in the profile plane proven); the plated cases in
      `TestCondition_PlatingOverTurningProves`, `TestRequirement_PlatedJournal` and
      `TestTolerance_CertifyProvesThePlatedJournal`; the coated block's export volume in
      `backend/mesh/displacement_test.go`
- [x] Task: Docs: [material_composition.md](../../../doc/material_composition.md) (the relief datum
      and depth table), [material_library.md](../../../doc/material_library.md) (datum and depth in
      "Relief") and a tolerance page, [tolerance.md](../../../doc/tolerance.md) (2026-11-27)
- [x] Task: CHANGELOG entries for every phase
- [x] Task: Phase Verification & Checkpoint (2026-11-27): `go test ./...`, 97 packages pass; gpuparity's two
      tests fail on origin/main itself. `tool/materialswatch`'s `TestTintedCoatIsARender` found the
      committed coat basis stale after Phase 5 moved the coats to their peaks; regenerated with
      `go run ./tool/materialswatch -manifest-only -coat-basis` (owner's go): `coat-basis.bin` and
      `catalog.json` (which also takes the two sheet presets' new doc lines); the swatch sheet and
      finish basis stay as committed. Lint clean. Phase 9 (optional) not done
