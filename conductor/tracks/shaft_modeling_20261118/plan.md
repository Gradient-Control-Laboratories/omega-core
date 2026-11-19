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
| **Implementation** | `pending`    |
| **Code Review**    | `pending`    |
| **Effort**         | `pending`    |

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

## Phase 1: A Coaxial Round Stack

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

## Phase 2: End Holes

- [ ] Task: The closure pass follows a list of functions through a parameter (spec 7, the owner's call)
  - [ ] Proposal: where the list binding is resolved (today `listBindings` binds one frame up), for the
        owner to approve before code
  - [ ] Write failing tests (Red Phase) in `internal/closure`: a library function handing a parameter list
        to another package's fold; `TestShaftPhase0_SpikeHoleListsThroughShaft` flips and is retired
  - [ ] Implement (Green & Refactor Phase); no existing golden moves
- [ ] Task: Proposal: `shaft(segments, topHole, bottomHole)`, and how the bottom stack is flipped
- [ ] Task: Write failing tests (Red Phase): the drill invariant holds at both end faces (no void surface
      in the end plane); depths measured from each end face
- [ ] Task: Implement (Green & Refactor Phase)
- [ ] Task: Example: a shaft center-drilled at both ends with `drill.centerHole`; WGSL size; the owner's
      viewer check
- [ ] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 3: Non-Axisymmetric Segments

- [ ] Task: Proposal: the segment factories (gear, male spline, D-flat, double-D, hex) and their
      parameters, sized as `drill`'s female counterparts are
- [ ] Task: Write failing tests (Red Phase)
  - [ ] Each segment against the reference
  - [ ] Mixed convexity: gear tip > neighbour > root; hex across corners > round > hex across flats
  - [ ] A male spline fits `drill.spline` of the same size (the fit test the driver bits use)
- [ ] Task: Implement (Green & Refactor Phase)
- [ ] Task: An example (a gear shaft); WGSL size; the owner's viewer check
- [ ] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 4: End Treatments

- [ ] Task: Proposal: how a segment names its end (plane, cutter-radius runout, chamfer)
- [ ] Task: Write failing tests (Red Phase): a hobbed gear's runout and a milled spline's runout follow
      the cutter's radius; chamfered tooth ends; plane ends unchanged
- [ ] Task: Implement (Green & Refactor Phase)
- [ ] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 5: Threads

- [ ] Task: Write failing tests (Red Phase): a thread segment from `thread.male` is exact between its
      runouts; it engages a `drill.tap` of the same size (`thread.engaged`)
- [ ] Task: Implement (Green & Refactor Phase)
- [ ] Task: An example; WGSL size; the owner's viewer check
- [ ] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 6: Joint Blends

- [ ] Task: Proposal: how a joint names its blend (none, fillet, chamfer) and its size
- [ ] Task: Write failing tests (Red Phase): a blended joint stays within its stated bound; the rest of
      the shaft is unchanged; no interior zero
- [ ] Task: Implement (Green & Refactor Phase)
- [ ] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 7: Chained Positions

- [ ] Task: Gate: spike 0(a) found the 3-read accumulator within budget, or the spec names another way
- [ ] Task: Write failing tests (Red Phase): a chained shaft equals its baseline twin; end planes computed
- [ ] Task: Implement (Green & Refactor Phase)
- [ ] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 8: Cranks

- [ ] Task: Proposal: the eccentric segment (offset, angle) and the section segment (a 2D field)
- [ ] Task: Write failing tests (Red Phase): crank pin and web against the reference; bridges between
      coaxial and eccentric segments
- [ ] Task: Implement (Green & Refactor Phase)
- [ ] Task: Example: a single-throw crank; WGSL size; the owner's viewer check
- [ ] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 9: Bearings and Bushings

- [ ] Task: Proposal: the `stdlib/mechanical/bearing` package. It covers seat segments (journal,
      locating shoulder, retaining-ring groove) and bodies from ISO tables (ISO 15 deep-groove first),
      with sources for every table
- [ ] Task: Write failing tests (Red Phase): a bearing body seats on its journal under a
      `joint.cylindrical` fit; table values against their sources
- [ ] Task: Implement (Green & Refactor Phase)
- [ ] Task: An example: a shaft on two bearings; WGSL size; the owner's viewer check
- [ ] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 10: Box Override and Simplified Model

- [ ] Task: Proposal: how a caller replaces the shaft's `.box`, and how the simplified model is asked for
      (a second entry point over the same segment list, or a flag)
- [ ] Task: Write failing tests (Red Phase)
  - [ ] The computed box contains the shaft; an override replaces it
  - [ ] The simplified model contains the detailed one (a gear at its tip, a thread or spline at its
        major diameter) and has no teeth, threads or knurling
  - [ ] `certify` on the simplified model proves faster than on the detailed one
- [ ] Task: Implement (Green & Refactor Phase)
- [ ] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 11: Verification and Records

- [ ] Task: Full Go suite with `go vet`; `tool/check.sh`
- [ ] Task: The stdlib's doc comments, `gsl skill`'s text and the language spec describe the new
      packages; CHANGELOG
- [ ] Task: Metrics recorded
