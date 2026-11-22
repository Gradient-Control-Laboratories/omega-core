# Implementation Plan: Shaft Modeling

## Overview

This plan implements [spec.md](./spec.md) in the order of its section 5, test first. Phase 0 proves or
refutes the bridge (spec section 3) before anything is built on it. It also runs three spikes, which
decide whether the closure machinery can carry a shaft as specified.

Every phase that adds a function to the stdlib checks four things before it is done:

1. **Distance:** exact outside wherever the segments are exact, and within the stated bound inside,
   against the Go reference.
2. **No sheet:** no interior zero on any joint plane.
3. **Provable and in WGSL:** `certify` and `bounds` prove on surface cells. The example compiles to
   WGSL under its size budget and renders in the viewer as the CPU evaluates it (the owner's check).
4. **No silent golden moves:** a stdlib edit can move `testdata/example` goldens. Any that move are
   shown with the reason, and regenerated only on the owner's go.

Where a phase needs an API the spec does not fix (a segment's signature, how a segment names its end
treatment), the phase opens with a proposal for the owner to approve before any code is written.

## Execution Metrics & Effort Tracking

| Metric             | Elapsed Time |
| :----------------- | :----------- |
| **Spec Creation**  | `~30 mins`   |
| **Plan Creation**  | `~10 mins`   |
| **Implementation** | `~49.5 h wall-clock, 2026-11-18 20:02 to 2026-11-20 21:35 (first to last commit)` |
| **Code Review**    | `not measured: commits carry no review signal` |
| **Effort**         | `per phase below, wall-clock between checkpoints` |

Recorded from commit times at the owner's choice (2026-11-20), so these are wall-clock, not active
time: they include nights and other work in the same days (Phase 4's span holds the gear PR #94).
Phase 0 has no start commit; Phase 7 was skipped.

| Phase | Checkpoint | Wall-clock since the previous |
| :---- | :--------- | :---------------------------- |
| 0 Ground truth and spikes | `f390b8c` 11-18 20:02 | (no start commit) |
| 1 Coaxial round stack | `516a000` 11-18 22:22 | 2 h 20 |
| 2 End holes | `7259b77` 11-18 23:51 | 1 h 29 |
| 3 Non-axisymmetric segments | `d2634b1` 11-19 00:46 | 0 h 55 |
| 4 End treatments | `2f62bdb` 11-19 23:08 | 22 h 22 |
| 5 Threads | `f690d4a` 11-20 12:11 | 13 h 03 |
| 6 Joint blends | `25ecd77` 11-20 13:32 | 1 h 21 |
| 8 Cranks | `9cf9e41` 11-20 15:11 | 1 h 39 |
| 9 Bearings and bushings | `0e6c8bc` 11-20 18:29 | 3 h 18 |
| 10 Box and simplified model | `c6de608` 11-20 20:32, bearing boxes `44ef859` 21:35 | 3 h 06 |

---

## Phase 0: Ground Truth and Spikes [checkpoint: f390b8c]

- [x] Task: A reference for shafts, in Go (`eval/shaft_phase0_test.go`)
  - [x] The Profile Modeling reference is not on `main`, so the shaft has its own. For round stacks it is
        exact rather than sampled: the meridian polygon mirrored across the axis, all straight edges. For
        the hex beside a round, outside is the min of two exact extrusions and inside is signed by
        containment
  - [x] The fold under test is a throwaway package injected through the mock source manager, so it runs
        the real closure pass
- [x] Task: The phantom, pinned
  - [x] Plain min of capped segments reads 0 at every joint point (asserted, so the check has teeth)
  - [x] A 1 mm inset opens a slit there
  - [x] The bridge passes both, is exact outside and a bound inside, against the reference
  - [x] Mixed convexity, with a hex (20 across flats) in place of a gear, beside a round of 22 (between
        flats and corners), both orders: exact outside, solid on the joint, real end faces where the
        corners stand proud
  - [x] Surface cells sound and tight; joint cells prove solid; `certify` proves clearance
  - [x] A gated bridge measured and rejected (spec 3.3): no more exact, and joint cells stop proving
- [x] Task: Spike (a): accumulator growth. Linear for every fold; chained positions viable (spec 7)
- [x] Task: Spike (b): a hole list through `shaft` is refused (pinned); voids work, placed by the
      caller or by the shaft's own `remap` (CPU; WGSL unchecked). Decision open for the owner (spec 7)
- [x] Task: Spike (c): a section segment from a captured `profileXY` field works
- [x] Task: Recheck spec section 6's constraints on `main` and record which still hold
- [x] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 1: A Coaxial Round Stack [checkpoint: 516a000]

- [x] Task: Proposal (approved 2026-11-18): `stdlib/mechanical/shaft`, entry point `shaft.body(segments)`,
      segments `shaft.diameter(dia, from, to)` and `shaft.taper(diaFrom, diaTo, from, to)` (a groove is a
      diameter), each a `seg(eDown, eUp)` closure; a taper's extensions continue at its end diameters and
      its distance is taken in its half-section (the owner chose this over `revolve(profileXY)`)
- [x] Task: Write failing tests (Red Phase), `eval/shaft_test.go`
  - [x] Diameters, tapers narrowing and widening, a cone shoulder, a taper into a step and into a
        shoulder, and a groove, against the exact meridian reference, outside and inside
  - [x] No interior zero on any joint plane; surface cells tight; joint cells prove solid
- [x] Task: Implement (Green & Refactor Phase); `body` is self-contained as a folding function
  - [x] Exact outside (worst 4.8e-6 mm over nine cases), a bound inside; surface cells within 1.76
        diagonals; every joint cell proves solid
  - [x] A taper's long side and caps are measured by clamping: projecting onto a side that starts 1e5 mm
        away lost 7e-4 mm to float32 (the cone shoulder case caught it)
  - [x] Mutation: `body` without its bridge fails the joint and surface-cell tests in all six
        multi-segment cases
  - [x] Phase 0's throwaway package is now imported as `spike`: an import alias that matches a stdlib
        package's name resolves to the stdlib package (spec 6), and `shaft` now is one
- [x] Task: `examples/shaft_turned.gsl` (eight segments)
- [x] Task: The owner's viewer check of `examples/shaft_turned.gsl`
  - [x] The interior's sharp contours at reflex corners: the stated bound, accepted for now; an exact
        interior recorded for later (spec 5)
  - [x] A collar (groove larger than the journal) showed a shallow sheet under it. Fixed by a bridge for
        every run of consecutive segments in place of neighbours only (spec 3.3, approved)
- [x] Task: The run bridge, test first
  - [x] Red: thin collar, thin groove and a collar on a step; inside never shallower than 1/√2 of the
        true depth (the cost of a square reflex corner), on every grid point, mid-plane and joint
  - [x] Green: `body` folds every run in one pass, accumulator `[shaft, open runs]`
  - [x] Mutation: neighbours only fails twelve subtests (the 3 mm groove and the cone shoulder too)
  - [x] WGSL linear, about 1.5 KB a segment; the example 23112 bytes, the budget
- [x] Task: The owner's viewer check of the run bridge: the collar's sheet is gone; the corner bound shows
      past a groove's rims and at a taper meeting a diameter (spec 3.3), accepted
- [x] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 2: End Holes [checkpoint: 7259b77]

- [x] Task: The closure pass follows a list of functions through a parameter (spec 7, the owner's call)
  - [x] Proposal (approved 2026-11-18): `paramIsFunctionList` counts a parameter handed on whole to a
        callee's function-list parameter, recursively (provisional cache entry ends a cycle)
  - [x] Write failing tests (Red Phase), `internal/closure/fold_forward_test.go`. Lists already forwarded
        within and across packages; what failed was a LONE function forwarded across packages (alone,
        beside a list, under a remap): the spike's `drill.centerHole(...)`
  - [x] Implement (Green & Refactor Phase), `internal/closure/prune.go`; the full suite passes and no
        golden moved
  - [x] `TestShaftPhase0_SpikeHoleListsThroughShaft` flipped and is converted to a positive check (owner's
        call): the holes are cut where the voids spike cuts them
- [x] Task: Proposal (approved 2026-11-18): `shaft.body(segments, bottomHole, topHole)`; the fold finds the
      end faces itself, segments answering `seg(eDown, eUp, want)` (0 the field, 1 and 2 their planes);
      the bottom stack turned a half turn about X
- [x] Task: Write failing tests (Red Phase): holes at both ends, lone and listed, depths from each face;
      the drill invariant at both end planes; outside unchanged clear of the mouths and never deeper
      anywhere; a keyway in the bottom hole stays on +x (rules out a turn about Y, not a mirror in z)
- [x] Task: Implement (Green & Refactor Phase)
  - [x] It met a second closure-pass fault: `nextClosureName` issued `gcl_closure_gsl_anon_2` both as
        gsl_anon's second name and as gsl_anon_2's first, and the later registration replaced the earlier,
        so the field fold called the plane fold's closures (the shaft read 5, its bottom plane, far
        from it). Fixed with the owner's go: every issued name remembered and a taken one skipped;
        `internal/closure/fold_twice_test.go`
- [x] Task: `examples/shaft_end_holes.gsl` (a centre at one end; countersink, tap and drill at the other):
      32998 bytes of WGSL, the budget. The `want` select costs every shaft about 7% (`shaft_turned`
      23112 to 24764 bytes)
- [x] Task: Full suite passes (no golden moved). `internal/link`'s `TestFastExitSingleOrZeroDotImports`
      (1000 links under 2 s) took 2.08 s under the parallel suite once, and 0.69 s alone
- [x] Task: The owner's viewer check of `examples/shaft_end_holes.gsl`
- [x] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 3: Non-Axisymmetric Segments [checkpoint: d2634b1]

- [x] Task: Proposal (approved 2026-11-18)
  - [x] `shaft.section(field2d, from, to)`, public: any 2D section as a prism segment
  - [x] Exact sections from `profileXY` (owner's call, over max/min bounds): `hex(acrossFlats, from, to)`
        with flats on ±x, `dFlat(dia, acrossFlat, from, to)` with its flat on +x, `doubleD(dia,
        acrossFlats, from, to)` with flats on ±x, and the male `spline(count, minor, major, width, from,
        to)`, one tooth (a box with an arc tip) repeated with `map.patternPolarXY` and min'd with the
        minor disc
  - [x] `gear(teeth, module, pressureAngle, helixAngle, from, to)`, spur and helical, from
        `gears.gearProfile`; the helix's twist is divided out by `sqrt(1 + (tipRadius * twistRate)^2)`
        so the field stays a bound (`gears.gear` does not correct it)
- [x] Task: Write failing tests (Red Phase), `eval/shaft_sections_test.go`
  - [x] hex, D-flat, double-D, spline and a generic section of a hex against a reference sampled from
        the same lines and arcs: exact outside (worst 2.9e-6 mm); exact inside for each prism, a bound
        for the spline's union
  - [x] Mixed convexity: hex across corners > round > hex across flats, both orders; a gear between rounds
        between its root and tip; exact outside, solid on every joint
  - [x] The male spline seats in `drill.spline` over a bore at its minor: nearest approach 0.002 mm, no
        interference
  - [x] The spur gear reads `gears.gearProfile` across its span; a helix adds no stretch over the spur
        (full 3D gradient inside the tip circle: helical 1.0002 against the spur's 1.1253; without the
        twist correction 1.19, which fails)
- [x] Task: Implement (Green & Refactor Phase)
  - [x] `gears.gearProfile` is not a bound, so neither is the gear segment. Beyond the tip its involute
        flanks read the string length outside their normal fans (slope 1.56, field 3.51 where the truth
        is about 2.7); inside the tip circle the tip chamfer overstates at the obtuse flank-tip edge
        (1.13). Measured, not the shaft's: fixed apart (owner's call, two PRs off main) as a bounded
        involute flank in `gearProfile` and a normalized chamfer in builtin
- [x] Task: `examples/shaft_pinion.gsl` (nine segments: spline, journals, helical pinion, hex, D, centres
      at both ends): 101596 bytes of WGSL, the budget. Each segment's section is evaluated four times a
      fold step (capped, below, both, above) where its axial term alone differs; computing it once is
      left for later
- [x] Task: Full suite passes; no golden moved
- [x] Task: The owner's viewer check of `examples/shaft_pinion.gsl`
- [x] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 4: End Treatments [checkpoint: 2f62bdb]

- [x] Task: Proposal: how a segment names its end (plane, cutter-radius runout, chamfer). Decided
      2026-11-19: finished prisms in the fold, runouts cut after it, the Phase 3 API kept (spec 3.4;
      B1, the blank in the fold, tried and dropped)
- [x] Task: Write failing tests (Red Phase): a hobbed gear's runout and a milled spline's runout follow
      the cutter's radius, into a larger neighbour, and the runout reads as a void; chamfered tooth
      ends; plane ends unchanged
- [x] Task: Implement (Green & Refactor Phase)
- [x] Task: `examples/shaft_hobbed.gsl`; WGSL size; the owner's viewer check
- [x] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 5: Threads [checkpoint: f690d4a]

- [x] Task: Proposal (approved 2026-11-20): `shaft.thread(major, pitch, from, to[, bottomEnd, topEnd])`,
      `thread.male` as a prism segment; ends `planeEnd`, `chamferEnd`, `runout(cutterDia)` and a new
      `threadRunout(length)`, whose groove floor rises in a straight line from the root to the major
      diameter (a lathe tool or a die retracting), cut after the fold (spec 3.4)
- [x] Task: Write failing tests (Red Phase): a thread segment from `thread.male` is exact between its
      runouts; it engages a `drill.tap` of the same size (`thread.engaged`); a thread runout follows its
      line into a larger neighbour and reads as a void; a chamfered start; no sheet on the joints
- [x] Task: Implement (Green & Refactor Phase): `thread` both ways, `threadRunout`, `threadedSegment`
      (the runout cut is the groove within the triangle the floor's line closes in the half-section:
      past the end plane, inside the major diameter); the thread package imported as `isoThread`,
      since `thread` names the segment in shaft.gsl
- [x] Task: `examples/shaft_threaded.gsl` (M10x1.5, chamfered start, runout into the neck, shoulder,
      journal); WGSL 15,609 B; the owner's viewer check
- [x] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 6: Joint Blends [checkpoint: 25ecd77]

- [x] Task: Proposal (approved 2026-11-20): a joint entry in the segment list, `shaft.fillet(r)` and
      `shaft.chamfer(c)`, between the segments it blends; the union blend of the two capped segments min'd into the
      fold; concave corners only (spec 3.4)
- [x] Task: Write failing tests (Red Phase), `eval/shaft_blends_test.go`: the fillet's surface is the exact arc at a round square
      shoulder; |grad f| = 1 outside near it; no interior zero; the rest of the shaft unchanged; the
      accumulator's WGSL cost measured (the owner chose these over the earlier "within its stated bound")
- [x] Task: Implement (Green & Refactor Phase): the closure pass hands a reduce over functions the
      previous element (4th callback argument; `internal/closure/unroll.go`, the owner's call); joints
      answer their sizes to constant questions; the blend stands in for the segment's term, so plain
      shafts compile as before. Two tests corrected with the owner's approval: the chamfer changes the true
      distance in a strip along the bisector; the 1/sqrt2 depth check holds farther than the blend's size
- [x] Task: `examples/shaft_blended.gsl` (three fillets and a chamfer, slider sizes); WGSL 15,128 B; the
      owner's viewer check
- [x] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 7: Chained Positions

- [ ] Task: Gate: spike 0(a) found the 3-read accumulator within budget, or the spec names another way
- [ ] Task: Write failing tests (Red Phase): a chained shaft equals its baseline twin; end planes computed
- [ ] Task: Implement (Green & Refactor Phase)
- [ ] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 8: Cranks [checkpoint: 9cf9e41]

- [x] Task: Proposal (approved 2026-11-20; Phase 7 skipped for now): `shaft.eccentric(dia, offset, angle,
      from, to)` and `shaft.web(mainR, pinR, offset, angle, from, to)` (`offset` was `throw`, a reserved word), the exact two-circle hull; webs of
      other outlines stay `shaft.section` (spec 3.4)
- [x] Task: Write failing tests (Red Phase), `eval/shaft_cranks_test.go`: the pin and the web exact
      alone; the crank exact outside against the min of its capped prisms; a unit gradient outside; no
      sheet on the four joints (against a lower bound of the true depth); the pin fillet's arc; controls
      for the hull polygon and the reference
- [x] Task: Implement (Green & Refactor Phase): `eccentric` and `web`; the web's section is the
      two-circle hull as the nearest of the discs between them, clamped, so it has no selects; the
      shipped examples' WGSL unchanged
- [x] Task: Example: a single-throw crank (`examples/shaft_crank.gsl`, fillets, centre holes, sliders for
      the offset and its angle); WGSL 22,437 B; the owner's viewer check
- [x] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 9: Bearings and Bushings [checkpoint: 0e6c8bc]

- [x] Task: Proposal (approved 2026-11-20): `stdlib/mechanical/bearing`; deep-groove 60/62/63, d 10-50
      (ISO 15 + a maker's abutment columns); a detailed body (rings, raceways, balls by polar repeat) and an
      envelope; ISO 4379 bushes C and F; DIN 471 grooves; nominal fits (ISO 286 later; `joint.cylindrical`
      is ISO 273 screw clearance, not a bearing fit) (spec 3.4)
- [x] Task: Write failing tests (Red Phase), `eval/bearing_test.go`: the three tables row by row against
      their sources (SKF 2018 + NTN-SNR for ISO 15's columns, SKF for abutment; DIN 471:2011 via Westfield and
      Mädler; ISO 4379:2018 itself); the ball rule; the envelope exact; the detailed body within it, balls
      and gaps where the rule puts them; a 6205 seats on journal and shoulder (proven) and a fillet past the
      chamfer is disproven; bushes (series argument) exact and seated; a control for the rounded ring
- [x] Task: Implement (Green & Refactor Phase): `stdlib/mechanical/bearing` (three sourced tables via
      `catalog.onRow`; `deepGroove` and `deepGrooveEnvelope`; Type C and F bushes and their envelopes, the
      flanged one a `profileXY` turned about the axis), registered in `pkg_info.go` and the mechanical index.
      Two tests corrected with the owner's approval: seat proofs run with a 20M-evaluation certify budget
      (1-4 s each); detailed-within-envelope is sampled, not proven, since intervals cannot separate
      coincident surfaces
- [x] Task: An example: a shaft on two bearings (`examples/shaft_bearings.gsl`: two 6205s on journals
      against shoulders, a DIN 471 groove, a flanged bush, a detail slider); WGSL 54,202 B; the owner's
      viewer check. Its test found a flanged-bush chamfer cutting the tube; fixed, and a mid-length check
      added to the bush test with the owner's approval
- [x] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 10: Box Override and Simplified Model [checkpoint: c6de608]

- [x] Task: Proposal (approved 2026-11-20): `shaft.body` declares a box from its end faces and each
      segment's reach; a caller overrides it with their own `.box`; `shaft.envelope(segments[, bottomHole,
      topHole])` is the simplified model (gear at tip, spline and thread at major, runouts dropped, smooth
      segments, blends and end holes kept) (spec 3.4)
- [x] Task: Write failing tests (Red Phase), `eval/shaft_box_test.go`
  - [x] The computed box contains the shaft (seven stacks, body and envelope, with holes); an override
        replaces it (passes already: a caller's `.box` replaces a library's)
  - [x] The simplified model contains the detailed one (sampled) and is exactly the tip and major
        cylinders outside, so no teeth, threads or splines; it keeps blends and end holes
  - [x] `certify` on the simplified model proves in fewer evaluations than on the detailed one
- [x] Task: Implement (Green & Refactor Phase): every segment answers its planes first, its envelope
      (want 4) and its reach (want 5); body declares the box; `shaft.envelope` both ways. A section leaves the
      box unbounded radially (a library cannot read a parameter's `.box`). One test corrected with the
      owner's approval: the envelope holds the detail by sign, since a helical gear's field is a bound
- [x] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 11: Verification and Records

- [x] Task: Full Go suite with `go vet`; `tool/check.sh` (all passed: stdlib regenerated unchanged, gofmt,
      architecture invariants, vet, tests)
- [x] Task: The stdlib's doc comments, `gsl skill`'s text and the language spec describe the new
      packages; CHANGELOG (shaft.gsl's summary, header and body docs brought up to date; `reduce`'s
      fourth `prev` parameter in the spec and the skill text; `bearing` documented since Phase 9;
      CHANGELOG already had every phase)
- [x] Task: Metrics recorded (from commit times, wall-clock)

## Phase 12: Gears as Shafts [checkpoint: 9ac230a]

- [x] Task: Review of `gears` against the shaft (2026-11-21): the gear and bevel gear examples are
      simple shafts (teeth, a hub after a chamfer joint, a through bore), short only a bevel segment.
      An audit of the bevel against references independent of its construction,
      `eval/gears_bevel_reference_test.go`: the standard equal-addendum straight-bevel formulas (face and
      root angles within 2.3e-4 rad, the arc m/R against atan(m/R); outside diameter +4.7 um; apex to crown
      -2.4 um; tooth thickness exact at both ends), and the profile against the spherical involute, the
      conjugate tooth (2.3 um at module 2, held to 5 um). A transmission-error sweep found none above its
      0.37 um resolution (backlash exactly 2b/cos(alpha)) and was dropped as too slow (47 s)
- [x] Task: Proposal (approved 2026-11-21): `shaft.bevel(numTeeth, mateTeeth, moduleVal, pressureAngle,
      shaftAngle, spiralAngle, faceWidth, backlash, apex, from, to)`, its planes the two flat faces, its
      facing from where the apex lies (a half turn about X when it faces +Z), bevelGear's ends, the blank as
      its envelope; box-extent queries (want 8, 9) so the toe teeth proud of the toe face stay in the box;
      backlash on shaft.bevel and a `shaft.gear` overload; `gears.bevelApexToCrown`; `gears.bevelGear` and
      the other part builders kept (spec 5, item 12)
- [x] Task: Write failing tests (Red Phase), `eval/shaft_bevel_test.go`: equal to gears.bevelGear facing
      -Z (straight, spiral, backlash); facing +Z is the half turn; a pinion on a shaft end (solid across
      the joint, no material added or lost, exact outside); a straddle journal keeps the toe teeth; the box
      holds them; the envelope holds the gear and is the tip cone outside; backlash thins by b/cos(alpha) on
      both segments; bevelApexToCrown
- [x] Task: Implement (Green & Refactor Phase): `shaft.bevel` (its teeth `bevelTeethField`, capped by
      `bevelSegment`, its box answers `bevelBoxAnswer`); a `shaft.gear` overload with a backlash (the
      section offset by half of it, as gears.bevelTeeth does); `gears.bevelApexToCrown`; the box reads
      extents (want 8, 9) through a test against 50000 that folds, so every example's WGSL is unchanged; the
      chamfer joint answers nothing to them. Three test defects corrected with the owner's approval (`ref` is
      a reserved word; bevelGear's zero bore caps its inside, so the comparison skips where it binds;
      backlash was defined twice). Findings: a `.box` sees only names declared before its value; handed the
      raw `want`, a helper that tells the envelope from the field itself leaves the shaft's box unbounded, so
      the bevel is asked with want 0 and a flag, as every segment's envelope is; `tool/check.sh` passed
- [x] Task: Examples: `gears.gsl` and `bevel_gears.gsl` as shafts placed with transform (teeth, a chamfer
      joint, a hub; `drill.thru` chamfered at both ends; phase a rotateZ, gear 2 turned pi (n - 1) / n and,
      for the bevel, swung by rotateY(shaftAngle)); both mesh with a steady clearance across a tooth (0.043,
      0.088); WGSL 11,427 -> 18,961 B and 37,664 -> 43,222 B; the owner's viewer check. The placed gears are
      returned unnamed: named, their part boxes (time and sliders, through rotation) took `bevel_gears` to
      276,697 B. No golden moved: `internal/testdata/example/` holds older documents of its own, built on
      gearPair and bevelPair
- [x] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 13: Fixes [checkpoint: d56bff4]

- [x] Task: Proposal (approved 2026-11-21), each from a diagnosis:
  - 13a: the box analysis folds comparisons of literals but not and/or/not, so every segment's first
    test, `want > 0.5 && want < 2.5`, stays symbolic and the select unions its branches; a field branch
    with no guard of its own then leaves every corner unbounded (Phase 12's "raw want" finding). Fold
    and/or/not in `foldOperation`.
  - 13b: `MarkParts` writes a part's box corners into its vanish guard as one tree, so what a turn shares
    is written again at every use (two nested turns of a slider box: 14 KB a corner). Done by #88's
    `shareBoxRepeats`, which landed on main first; a `sharedLocals` pass written here was dropped.
  - 13c: the fold bridges the first segment to its seed (`max(100000, seg)`), which can never win.
    `i < 0.5 ? ...` in the four fold bodies (measured: bevel_gears 43.2 -> 33.7 KB, every shaft
    example 1-9% smaller).
  - 13d: the PNG framing scan returns only the inside cells once any is found, losing a part thinner
    than the 85 mm coarse cells (the threaded shaft's thread). Union the inside and near bands.
- [x] Task: Write failing tests (Red Phase), `eval/shaft_fixes_test.go`: and/or/not of literals pick a
      branch; a caller's unguarded segment keeps the shaft's box; a turned part's WGSL grows linearly
      with nesting; the first segment is written no more often than a middle one (and the field is
      unchanged, which passes before and after); framing keeps a thin rod beside a block
- [x] Task: Implement (Green & Refactor Phase): and/or/not folded in `foldOperation`; `i < 0.5` in the
      four fold bodies; framing unions the inside and near bands. A part's box sharing is #88's
      `shareBoxRepeats`: a turned part's WGSL by depth is 10,838 / 15,871 / 39,162 -> 10,296 / 11,196 / 12,185 B
      with it (the `sharePartBox` written here, dropped, gave 10,570 / 11,709 / 12,962 B). Shaft examples
      against main: bevel_gears 43,800 -> 34,258 B, gears 19,539 -> 16,573 B, every shaft example 1-9%
      smaller; the bevel pair with its placed gears named is 46,649 B (276,697 before), but they stay
      unnamed (their boxes move). No codegen golden moves: the sharing that moved them is already in main's.
      Found: every segment's field is still written about 20 times into the shader (a later item)
- [x] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 14: Knurls [checkpoint: ceeb5d4]

- [x] Task: Proposal (approved 2026-11-21): `shaft.knurl(dia, pitchVal, pattern, from, to)`, the blank as
      geometry (a plain diameter in the fold, the envelope and the box) and ISO 13444 teeth (90 degrees,
      pitch / 2 deep, a whole number round) as a displacement height bounded as thread.maleRelief is;
      patterns as constants (knurlStraight, knurlLeft, knurlRight, knurlDiamond, knurlDiamondSunk,
      knurlAnnular; helices at 30 degrees); new questions want 10 (height) and want 11 (minus depth
      fraction), which every other segment answers nothing; `shaft.relief(segments)`,
      `shaft.paint(baseColor, segments)` and finished `shaft.body(segments, finish)` /
      `(segments, bottomHole, topHole, finish)`: steel with the paint, the envelope painted as `.simple`
- [x] Task: Write failing tests (Red Phase), `eval/shaft_knurl_test.go`: every pattern's height; the
      invariants (straight along the axis, annular round it, a helix along itself, crest 0, root a tooth
      depth); the bounds (planes, beyond, axis, plain neighbours); the geometry is the blank; paint darkens a
      root and leaves a crest; the finished body is the plain one with displacement and a stand-in in its
      shader; a plain shaft's relief is 0
- [x] Task: Implement (Green & Refactor Phase): `shaft.knurl` (a diameter's closure answering want 10 with
      `knurlHeight` and want 11 with it over the tooth depth), the six pattern constants, `knurlTooth`;
      `shaft.relief`, `shaft.paint`, finished `body(segments, finish)` and `body(segments, bottomHole,
      topHole, finish)` (each written out in full, as body is: the fold, the box, the relief, the depth and
      the envelope); the bevel answers nothing to wants 10 and 11. Every shaft example's WGSL unchanged.
      One test defect corrected with the owner's approval: a lone displaced part's height rides in its
      material record (`gsl_sdShapeDisplaced` is written only for several parts of different heights),
      and its stand-in is `gsl_simple`
- [x] Task: An example (`shaft_knurled.gsl`: the six patterns side by side, each finished in steel, named
      parts with constant boxes): WGSL 71,287 B; the CPU render draws the teeth (a right-hand knurl rises to
      the right on its near face, as a right-hand thread does); the owner's viewer check
- [x] Task: The owner's viewer check found the paint stayed with textures off, where a fastener's goes
      plain. Cause, general and not the shaft's: a property's companion (`companionOf`, carry.go) copied
      its function's assignments with every property stripped, so a colour built in a local with a
      stand-in (`t.simple`) and set as `s.color = t` lost it; only a colour returned from a function kept
      its own. Fixed with the owner's approval: the companion keeps a value's `.simple`, `.detail` and
      `.detailSize`. Red first, `eval/simple_local_test.go` (a document function's local colour keeps its
      stand-in, as a returned one does; the finished knurl's paint carries one). No golden moved
- [x] Task: Phase Verification & Checkpoint (Refer to `workflow.md`); the owner's recheck of the Textures rung in the viewer is pending (away; to be reminded at wrap-up)

## Phase 15: Thread Undercuts [checkpoint: a9798b1]

- [x] Task: Proposal (approved 2026-11-21): `shaft.undercut(form)`, a thread end. The thread keeps its planes
      up to the shoulder and its last g2 becomes the DIN 76 groove: floor at d - I, a 30-degree flank (to the
      axis) to the major diameter on the thread's side, cutting the last teeth down, and radius r at the end
      plane; a revolved half-section overlapping the threaded part, so no sheet. Forms `thread.din76A` (g2
      3.5 P, the table) and `thread.iso4755` (g2 3 P); Form B (2.5 P, one 2004 source) left out. The table
      (P 0.2-6: I, g2, r) in mechanical/thread: `undercutDia`, `undercutLength`, `undercutRadius`. Sources:
      HEIDENHAIN MANUALplus manual's "undercut DIN 76" table (I, K, R, W = 30 deg); ISO 4755:1983 via IS
      1369-2:1993 (turngcode.com). g1 = g2 - sqrt(3) (d - dg) / 2 reproduces ISO's printed g1.
      DIN 76-1:2025-10 corrected some g2 and r values; not included
- [x] Task: Write failing tests (Red Phase), `eval/shaft_undercut_test.go`: the table row by row (DIN g2 and r,
      ISO g2 and r, the undercut diameter); g1 from the flank (passes: a check of the relation); the floor,
      fillet and flank where the table puts them, nothing outside the flank and a tooth reaching it; the
      plain thread below the groove and solid across the overlap; ISO's shorter groove; the bottom end; the
      box
- [x] Task: Implement (Green & Refactor Phase): the table and `thread.din76A` / `thread.iso4755` in
      thread.gsl; `shaft.undercut(form)`, end kind 4, in `threadedSegment`: the threaded part cut by
      `undercutFlank`'s half-space and `undercutGroove` (a floor rectangle and a fillet corner less its disc,
      in the half-section) added over it; the thread's runout cut now keyed to kind 3 alone. A first groove
      built with profileXY was NaN capped (two coincident corners) and cost 52 KB on a thread whose end a
      slider picks (both branches are built), so it is plain fields. Threads with literal ends compile
      identically; the threaded example, whose ends sliders set, carries the table lookups: 11,305 ->
      17,400 B. Two test defects corrected with the owner's approval: the fillet's point was on the wrong
      quarter of its circle, and NaN passed every comparison (undercutEval now fails on it)
- [x] Task: An example (`shaft_undercut.gsl`: an M12 x 1.75 thread chamfered at its end and run up to a
      22 mm shoulder through a DIN 76 undercut, g2 6.1): WGSL 13,319 B; the CPU render shows the groove and
      the flank-cut last teeth; the owner's viewer check
- [x] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 16: Cross Holes and Set Screws [checkpoint: 61e4b57]

- [x] Task: Proposal (approved 2026-11-21): `shaft.crossHole(features, dia, z, angle)`, a list entry the fold
      passes, its drill stack's void cut after the fold (want 3), the mouth the plane tangent to the cylinder
      `dia` across at height z, `angle` from +x, so depths are from the surface; `shaft.crossPlace(part, dia,
      z, angle)` puts a +Z-datum part on the same axis with transform; `fastener.setScrew(nominalDia, length,
      point)`, hex socket set screws to ISO 4026 (flat), 4027 (cone), 4028 (dog), 4029 (cup), M3-M12, from a
      sourced table and the existing set screw blocks
- [x] Task: Write failing tests (Red Phase): `eval/shaft_crosshole_test.go` (blind, through and tapped holes
      where their placement puts them; the shaft unchanged elsewhere, its box too; a fillet before the entry
      still blends; a placed pin on the axis, its box kept) and `eval/fastener_setscrew_test.go` (the table row
      by row: socket width, depth either side of each standard's threshold, point diameters, dog lengths;
      each point's end and the socket on an M6 x 10; the box). Sources for the table (2003 editions): ISO
      4029:2003 and ISO 4026/4027/4028:1993 (iTeh samples), Westfield Fasteners, fasten.it, amesweb.info;
      the 2003 thresholds of 4026 and 4027 rest on amesweb alone; the 2026 editions not consulted
- [x] Task: Implement (Green & Refactor Phase): the approved entry form failed in the fold (an entry cannot
      capture its own drill feature list: "a list of functions has to be visible where it is folded"); the
      owner chose `shaft.crossHole(void)` with `shaft.crossVoid(features, dia, z, angle)` placing the stack,
      and the tests' calls followed. `shaft.crossPlace` with transform. `fastener.setScrew` (plain and finished),
      its table, `setScrewSocket`, `setScrewSocketDepth`, `setScrewPointDia`, `setScrewDogLength` and the point
      constants, built from the existing blocks on a smooth shank. Three test defects corrected with the
      owner's approval: the blind hole's deep point was nearest the drill point, not the wall; the "away"
      point lay on the hole's axis; two "solid" points sat within the threshold's 0.2 of a surface
- [x] Task: An example (`shaft_set_screw.gsl`: a collar on a D-flat held by an M6 x 8 cup point set screw in
      a tapped cross hole, finished in steel with its thread as relief, and a 4 mm through hole for a pin):
      WGSL 65,805 B; the owner's viewer check found the shaft's ends cut like a huge centre drill (the
      example had passed drill.chamfer as end holes, a cone void 17.6 across); its ends are now 0.8 mm
      tapers
- [x] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 17: Stand-ins and Fits [checkpoint: 21e521a]

- [x] Task: Proposal (approved 2026-11-21): (a) shaft.body gives a shaft with detail its envelope as its
      stand-in: a new question, want 12 "has detail" (gear, spline, thread, knurl, bevel), folded to a
      constant; the approved "drop a stand-in identical to its part" proved unworkable (part and stand-in are
      different calls where stand-ins are attached, and the cost was the solid read twice, not the wrapper),
      so, the owner's choice: `.simple = detailed ? envelope : nothing`, and the codegen writes a part whose
      stand-in is the constant nothing as the part alone. (b) A new mechanical/tolerance package: ISO 286
      limit deviations, shaft j5 j6 k5 k6 m5 m6 n6 p6 and hole H6 H7 J6 J7 K6 K7 M6 M7, and IT5-IT7, over 6 to
      120 mm; bearing.seatDia and bearing.housingDia in a zone at maximum or minimum material; a k5 journal in
      the bearing example
- [x] Task: Write failing tests (Red Phase): `eval/shaft_standin_test.go` (a detailed shaft has the switch,
      a plain one none, a part whose stand-in is nothing none) and `eval/tolerance_test.go` (every zone and
      grade in every range, inside and at its upper end; a 6205's k5 journal and H7 housing at both
      conditions). Sources: Simply Bearings' ISO 286 tables and Miki Pulley's JIS B 0401 excerpt (agree but
      K6 6-10, where -7 is right by IT6); Wikipedia's IT table; neither source states 286-2's edition
- [x] Task: Implement (Green & Refactor Phase): want 12 answered 0 by gear, spline, thread, knurl and bevel
      (all nine closures); body and body-with-holes set `.simple = detailed ? envelope : nothing`; the codegen
      (simple.go) writes a part whose stand-in is nothing, or a remap of it (a part placed with transform
      hands its stand-in the same remap), as the part alone. mechanical/tolerance (registered in pkg_info and
      the mechanical index), its table, zone, grade and condition constants, deviations, grade, shaftDia and
      holeDia; bearing.seatDia and housingDia. Every shipped shaft example's WGSL unchanged; no golden moved
- [x] Task: The bearing example's two seats as k5 journals at maximum material (25.011, an interference fit
      with the 25 mm bore as designed): WGSL 36,801 -> 37,134 B; the owner's viewer check
- [x] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 18: Shader CSE [checkpoint: a0c443f]

- [x] Task: Proposal (approved 2026-11-21). Measured: body's fold evaluates every segment four times,
      capped and extended ((0,0), (E,0), (E,E), (0,E)), the stand-in's envelope fold four more, a runout's
      cut one more; the section (gear profile, spline pattern, a helix's cos and sin, length(p.xy)) does not
      depend on the extension and is the same expression in each copy: shaft_pinion's shader calls
      gears_gearProfile 4 times with identical arguments, shaft_hobbed 5, and length(p.xy) appears 15-35
      times per shaft (the "about 20 times" of the earlier notes was wrong). Fidget's context deduplicates,
      so CPU renders, meshing and proofs already evaluate each once; only the WGSL kept them, because the
      inliner names a value only when one variable is reused within one body. The owner's choices: a
      compiler pass, not a segment-protocol rewrite; in the WGSL codegen, not the program IR (there the
      remap's rebound p, fold loops, .simple/carry reference copies, meshplan names and the stand-in's
      guards would all need barriers). Design: `ir.EliminateCommonSubexpressions` on every structured
      function before formatting (raw sources and runtime snippets untouched): a parser for the
      expressions the codegen emits, recording spans so a rewrite only replaces a span with a name; rule 1,
      a let that repeats an earlier one in scope is dropped and read as it, and a subexpression an in-scope
      let names is read as that let; rule 2, a subexpression repeated within one block is named
      `gsl_cse_N` before its first use when that shortens the function, largest first. Never out of a
      guard or a loop body, never what reads a written name (a var, an assignment's or an opaque
      statement's target, a guard header's), never from the right of && or ||; opaque statements verbatim.
      Frame time may not move (the driver may merge these already); WGSL size, naga/Tint and the driver's
      compile will. Verification: golden diffs before any regeneration (owner's authorization), example
      sizes, TestGpuFieldParity and TestShaderCompileCost through main's prebuilt gpu-eval symlinked in
- [x] Task: Write failing tests (Red Phase): `backend/wgsl/internal/ir/cse_test.go` (parse shapes and
      precedence; broken text rejected; every let, var and return expression of all codegen and example
      goldens parses; both rules; a short repeat left; name collision; a name below an opaque early return;
      sibling guards keep their own; a guard reads the let above it; a loop body keeps its own and names
      its own repeat; var reads, opaque writes and a short circuit's right side not merged) and
      `eval/shaft_cse_test.go` (shaft_pinion calls gears_gearProfile once; no function of shaft_turned
      reads length(p.xy) twice). Red: 10 of the unit tests and both integration tests fail (4 calls; 28
      reads in one function); the barrier tests and the rejection test pass against the no-op stub, as they
      should
- [x] Task: Implement (Green & Refactor Phase): `backend/wgsl/internal/ir/cse.go`: the parser; value
      numbers (a let's name reads its value's, a written name a fresh one at every occurrence, so nothing
      reading it merges); rule 1 in statement order with a scope per block; rule 2 per block, largest first,
      a later let of the same value moved up to the first use in place of a new name; the length test
      counts the new line's indentation and newline and an inline guard's reflow (without them one example
      grew 6 bytes). `Module.EliminateCommonSubexpressions` counts the module's private globals as written
      everywhere, and the codegen runs it before the module is written (codegen/build.go). Two tests
      amended with the owner's authorization: the corpus test skips a line holding two statements (a raw
      material function's `let gsl_m = ...; return ...;`), and the turned shaft's assertion became "no
      function holds two lets of one expression" (four length(p.xy) stay: naming them would lengthen
      gsl_remap), scanning the document's functions, not the host's march. Sizes (CLI): shaft_pinion 87,263
      -> 35,728 B, shaft_hobbed 50,963 -> 24,439 (one gear profile, was five), shaft_set_screw 66,057 ->
      41,215, bevel_gears 33,752 -> 24,317, shaft_turned 19,305 -> 14,565; the example corpus 2,912,121 ->
      2,848,000 (-2.2%), 41 unchanged, none larger; compile time +0.06 s at most
- [x] Task: Golden diffs and regeneration (owner's authorization: done, 46 example and 32 codegen goldens;
      TestSliderArithmetic's `sphere(p, radius * 2.0)` became `sphere(p, scale)`, the let in scope holding
      it); field parity and compile cost; the owner's viewer check. TestGpuFieldParity (main's prebuilt
      gpu-eval symlinked in): 61 pass, 0 fail, 28 skipped by the harness's own limits (sliders, textures,
      audio); it checks 26 of the 46 changed example goldens, every one matching fidget; the other 20 (the
      gear examples, Kepler, CoreXY, the slider-driven materials) and the shaft examples, outside its
      corpus, rest on the viewer. TestShaderCompileCost, medians of 3 cold runs, before (HEAD from git) and
      after: the 26 changed examples' naga module 93 -> 94 ms and pipeline 2,830 -> 2,825 ms (-0.2%), the 35
      unchanged ones 2,115 -> 2,118 ms (+0.2%): no measurable change. The driver's pipeline compile is 97% of
      the cost and merges pure duplicates itself; the corpus it can time shrank 0-11%, and the shafts (25-58%
      smaller) are not in it
- [x] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 19: Feature Pruning [checkpoint: 86b42c7]

- [x] Task: Proposal (approved 2026-11-21: direction, opt-in, declared else inferred). Findings: the box
      analysis (property/box.go) is an abstract interpreter over the source, before inlining; every remap
      but the transform library's clears a box, so a polar fold, map.pattern* and drill.hole's own remap
      are unbounded. Vanish (gsl_vanish) measures a part's world box against the eye, which is why
      MarkParts refuses anything read through a remap, and the features to prune nearly all are (holes,
      teeth in their fold, a shaft laid along X). gsl_pw, the detail footprint, is set only at a hit, so
      the march, normals, shadows and AO never see it. Design:
      (a) The footprint: at sdShape's entry, from the world p and gsl_vanishEye,
          gsl_featurePixel = eye.w * mix(max(|eye - p|, 1), 1, ortho), vanish's own pixel, in a private
          global every guard reads; rigid remaps keep distance, so it holds inside a translate, rotate or
          fold, and every kind of ray reads the same one. A host without the eye (w <= 0) prunes nothing.
      (b) The size, the owner's choices: opt-in, and declared, else inferred, else never. A feature opts
          in with a geometric property, `name.featureSize = size` (mm); nothing else is ever pruned, so
          every other model draws as before. The stdlib sets it where it builds a feature (drill
          features their diameter; gear teeth the tooth thickness, pi m / 2; threads and knurls the
          pitch). A size of 0 asks for it to be inferred: the smallest extent of the operand's box,
          where the analysis bounds it, and if it does not, the feature is never pruned.
      (c) The marker: builtin pruneBelow(field, size), the field everywhere but the viewer's WGSL, which
          writes it as a guard that does not evaluate the field once pruned, with vanish's thresholds
          (the Detail slider's) on size / gsl_featurePixel and vanish's fade, field + (1 - w) * size / 2.
          Allowed only as a union operand or a subtracted one (min(a, x), max(a, -x)), where nothing is
          the identity: a bump drops out, a hole closes. Anywhere else (an intersection) it is an error.
      (d) gears.gearProfile re-authored as the pitch disc, with the tips added and the gaps taken away,
          each a pruned operand overlapping the disc across the pitch circle so no sheet of zero lies
          inside; pruned, the teeth leave the pitch cylinder. shaft.gear and gears.gear inherit it.
      Not in this phase: meshing passing its cell as the footprint; scaled features (transform.scale's
      factor); teaching box.go which remaps keep distance.
- [x] Task: Write failing tests (Red Phase), proposal approved 2026-11-21 ("Approve + Red tests"):
      `eval/feature_prune_test.go`: opt-in (no feature, no footprint; a featured hole guarded); the
      footprint set as sdShape is entered, from p and the eye; a feature under an intersection an error
      naming featureSize and the union it may stand in; union and subtracted operands allowed; the CPU
      field unchanged (the hole reads 0.5); drill holes and shaft.gear's teeth opt in; the re-authored
      gearProfile the same solid as the old one, frozen in the test, and as deep at the pitch circle inside
      a tooth as half the old one's (no sheet). `internal/fidget/internal/gpuparity/prune_test.go`, on the
      GPU through gpu-eval with the camera set: the plate's hole open from 50 away, closed from 1500 (the
      plate stays), open with no camera; a 60-tooth module-2 gear from 5000 away reads its pitch cylinder,
      r - 60, at every angle, and from 60 away does not. Red: all fail on the missing property or guard
      (the GPU gear on the undeclared camera); the two profile tests pass, guarding the re-authoring
- [x] Task: Implement (Green & Refactor Phase). Threads deferred (the owner): a thread is one profile,
      pruned whole it would take the shaft; it needs the gear's treatment, a pitch-diameter cylinder plus
      crests less roots. Knurls need nothing (their teeth are a displacement). Built: `featureSize` a
      property (propval, scope); property/feature.go lowers it to builtin pruneBelow and checks every read
      of a feature in its scope (a min's operand, a max's negated one, a union or difference combiner's,
      a remap's or select's carrying, a return; anything else an error naming featureSize and the union);
      size 0 reads builtin featureExtent of the value's box. codegen/prune.go writes the guard (the weight,
      the fade, nothing once gone), sdShape's first statement measures gsl_featurePixel, the vanish
      snippet declares the camera (so both hosts set it, as for parts) and a program with features but no
      parts resets gsl_partNear for the shading. gears.gearProfile re-authored: the pitch disc, min the
      tips (down to the root), max the gaps taken away (out to the tip circle), both a tooth thickness, pi
      m / 2. drill.hole: a library function cannot read a box (".box ... not supported inside a
      differentiated function body"), so, the owner's choice, the void sizes itself by probing its own
      field, twice its depth on its axis 0.01 under the mouth. That found two things, both fixed with the
      owner's go, Red first: the WGSL wrote remap(f, constant point) as a translation, p - c
      (wgsl/func_call.go, since the initial commit), where the language spec, the CPU and exports read the
      field at the point (gpuparity/remap_point_test.go); and the fold could not evaluate a remap to a
      constant point, so each hole evaluated its stack twice: the constant evaluator binds p while it reads
      one (internal/eval, eval/remap_fold_test.go). A drill's size is now a literal (Drill: 12.15, 12.0,
      7.34, ...); inside a fastener function that stays a function (Bolted Joint, Interface Joints) the
      probe still runs, one extra evaluation of that hole. References updated with authorization:
      parser2 property.gsl's error list; sphere_gears_text.assembled and its dependency snapshot; font2.wgsl
- [x] Task: Example and the owner's viewer check: examples/feature_pruning.gsl, a 150 mm gear with 1.6 mm teeth
      and a plate of 1.5 mm holes on a 6 mm grid; the owner: "Looks right". Goldens regenerated with authorization;
      GPU parity 61 pass, 0 fail
- [x] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)
