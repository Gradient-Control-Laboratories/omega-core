# Implementation Plan: Profile Modeling

## Overview

Implements [spec.md](./spec.md) in the order of its section 4, test first. Phase 0 commits the ground truth
the rest is measured against: #45 checked `profileXY` against a brute-force reference only in scratch, so
no committed test says today's profiles are exact. Each later phase adds one candidate and holds it to the
same reference.

Every phase that adds a function to the stdlib checks four things before it is done:

1. **Exact** (or, where the spec allows a bound, within it): the field against the Go reference.
2. **Provable:** interval bounds over surface cells are sound and tight, `certify` proves clearance, and
   `bounds` proves the box (`eval/profile_provable_test.go`).
3. **In WGSL:** the example compiles to WGSL, under its size budget, and renders in the viewer as the CPU
   evaluates it (the owner's check).
4. **No silent golden moves:** a stdlib edit can move `testdata/example` goldens; any that move are shown
   with the reason and regenerated only on the owner's go.

Where a phase needs an API the spec does not fix (how a vertex names its blend, how an edge says it is a
Bezier), the phase opens with a proposal for the owner to approve before any code is written.

## Execution Metrics & Effort Tracking

| Metric             | Elapsed Time  |
| :----------------- | :------------ |
| **Spec Creation**  | `~20 mins`    |
| **Plan Creation**  | `in progress` |
| **Implementation** | `pending`     |
| **Code Review**    | `pending`     |
| **Effort**         | `pending`     |

---

## Phase 0: Ground Truth [checkpoint: ac68832]

- [x] Task: Gate: the development Rust library (`target/release`) matches `main`'s Rust; the Go packages
      that evaluate fields (`eval`, `cmd/compile`) did not link against the one there. Met on the owner's go
      by copying a build whose Rust is identical to 6f5462b, rather than rebuilding.
- [x] Task: A brute-force reference for profiles, in Go (`eval/profile_reference_test.go`)
  - [x] Distance to a densely sampled boundary, sign by containment; one helper the later phases reuse
        (`refLoop`, `newRefProfile`, `profileCases`)
  - [x] Today's `profileXY(points)` and `profileXY(points, curvatures)` agree with it to within 2e-5 mm
        (worst 3.6e-6): an L, a star, a square with a hole (`max` by hand), the arc L of
        `profile_cylinder.gsl`, and a slot with a collinear point and tangent-continuous joints. Restoring
        #45's `turn >= 0` corner sign fails the slot (+5 for -5 on the collinear point's normal).
  - [x] `extrude` and `revolve` of those profiles agree with the reference composed in 3D (worst 3.5e-6
        and 8.6e-6)
- [x] Task: Provability and size, recorded (`eval/profile_provable_test.go`)
  - [x] Interval bounds over 0.5 mm cells on each profile's boundary, and on an extrude's rims and cap,
        sharp and with `intersectRound(2)`, contain the field and are at most 4 cell diagonals wide (widest
        2.16). `certify` proves each extrude clear 0.25 mm above its cap and beside it, both rim kinds.
        Restoring #45's gated magnitude (`outside <= inside ? outside : neg(inside)`) fails every profile
        (17 to 78 diagonals).
  - [x] `bounds` (`Evaluator.Enclose`, the interval proof `gsl eval` runs; not `Evaluator.Bounds`, the grid
        fit beneath it) proves each extrude's box, holding the reference's and standing off it by at most
        0.071 mm. Not the star's over the test's region: there it runs out of budget, though the same star
        proves over [-30, -30, -10, 30, 30, 22]. That is `Enclose`'s sensitivity to where its cuts land, a
        finding for the `gsl eval` tracks, and the test logs the star rather than asserting it.
  - [x] The WGSL size of each `examples/profile_*.gsl` on 85a0152, the budget later phases compare against:
        `profile_star` 12,337 bytes, `profile_sweep` 113,847, `profile_arcs` 119,522, `profile_cylinder`
        190,178
- [x] Task: Recheck spec section 5's constraints 1–4 on `main` and record which still hold in the spec:
      1 and 4 are lifted, 2 and 3 hold
- [x] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 1: Several Loops in One Profile (C1) [checkpoint: c94b0d8]

- [x] Task: Proposal: the signature (a list of loops, each points or points and curvatures), how nesting is
      checked, and what a loop that does not nest reports. Decided: a generic `nested(fields)` taking
      each loop as its own `profileXY` field, nearest surface wins; the winding rule documented and pinned
      by a test, not checked (spec section 4, C1)
- [x] Task: Write failing tests (Red Phase) (`eval/profile_nested_test.go`)
  - [x] An outline with two holes and an island inside one, polygon and arc loops mixed, against the
        reference; the same loops in another order; two separate islands (worst 3.6e-6)
  - [x] `extrude` of it against the reference (1.9e-6); surface cells tight (widest 1.93 diagonals);
        `certify` proves the holes and the space above empty; `bounds` proves its box
  - [x] `.u` is the nearest loop's own
  - [x] A 3D solid with a void against an analytic reference (7.1e-7)
  - [x] A mis-wound loop's result pinned: its own distance where it is nearest, so void just outside it in
        the material
- [x] Task: Implement (Green & Refactor Phase); `.box` and `.u` across loops: `nested(fields)` in
      `stdlib/builtin/booleans.gsl`; box and `.u` carry without code of their own. Restoring a gated
      magnitude fails the plate's surface cells (51 diagonals). No golden moved (stdlib, assembler, parser,
      program and WGSL suites)
- [x] Task: An example, its WGSL size against the budget, and the owner's viewer check:
      `examples/profile_nested.gsl`, a flange of eight loops nested four deep and a block with a void cut
      open; 100,690 bytes of WGSL (its own budget from here); the other examples' sizes unchanged; checked by the owner in the viewer
- [x] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 2: Corner Blends, per Vertex (C2) [checkpoint: 4496b6e]

- [x] Task: Proposal: how a vertex names its blend (none, arc fillet, line chamfer, field blend) and its size,
      and what is refused (a fillet larger than its edges allow). Decided: corner helpers in a `corners`
      list, one edge-blend operator per profile for `cornerBlend()` vertices, `everyCorner()`, clamping
      (spec section 4, C2)
- [x] Task: Part A, corners between straight edges
  - [x] Write failing tests (Red Phase) (`eval/profile_corners_test.go`): exact fillets and chamfers
        against the reference, convex and reflex, mixed with sharp corners (worst 2.1e-6); `everyCorner()`
        (3.3e-6); clamped fillets make the circle they should (2.8e-6); surface cells tight, at cells scaled
        to the smallest fillet (widest 2.89); a corner next to an arc stays sharp. WGSL: the U with its
        eight corners blended is 8,079 bytes, against 62,374 for the plain polygon (each step's work is a
        helper taking values, which the WGSL backend emits as a function instead of unrolling; a helper
        must not index a list by a passed index, which WGSL refuses as not constant)
  - [x] Implement (Green & Refactor Phase): the pools walk a corner's trimmed edges and its arc or chamfer
        (`profileXY(points, curvatures, corners)` and the corner helpers in `stdlib/builtin/2d.gsl`).
        Flipping the fillet's curvature sign fails the reference. No golden moved. Found on the way, not
        fixed: a `map`-built list handed in as `curvatures` fails to assemble once the profile is named,
        with or without corners (its box pass, likely); the test writes its curvatures out
  - [x] Proposal, then tests and implementation: how a `cornerBlend()` corner's field enters the pools,
        where it is a bound, and that the rest of the loop stays exact. A piece between trims at half the
        shorter edge, its field the blend of the two edge lines (the dual at a reflex corner), never less
        than the distance to the triangle of the corner and its trims: without that bound the blend of two
        lines claims less than the corner's distance far off and takes the sign of other parts of the loop
        (`intersectLp`: 139 of 4,809 far points wrong-signed). `eval/profile_corner_blend_test.go`, with
        `intersectRound`, `intersectChamfer` and `intersectLp`: exact outside the pieces (1.9e-6); the
        documented formula inside them; 1-Lipschitz and continuous (steepest 1.0001); surface cells sound
        (widest 3.35); the sign right at every far point. `profileXY(points, curvatures, corners, blend)`;
        WGSL 21,007 bytes for the U with two blend corners
  - [x] Viewer, cost and frame (`examples/profile_corners.gsl`, checked by the owner):
    - The first build hung the GPU. The corner pools now make one pass for all three pools (each feature
      offering a vector, the nearest pool keeping a constant gate of its own), a fillet and a chamfer are
      one edge, and each trim is computed once. `gpu-eval`, 4M points: the example from 124 ms to 61–63
      ms, against 39–43 ms for `profile_arcs.gsl` (whole examples only: that harness's ~25 ms floor makes
      its per-part numbers noise). Folding the per-vertex geometry by inlining was measured and is not to
      build (backlog `M15`): the arcs example 3.7× slower, the corners example 2.3× faster, and slider-
      driven profiles never qualify.
    - A blended corner's `.u` follows its two legs and changes piece where the field does (it had run
      along the chord, diagonally across half of each edge).
    - Found, worked around, not reduced: with a blend made from a slider, the blend form's two-name box
      assignment went undefined (the box now comes from `profileArcBoxXY`).
- [ ] Task: Part B, corners next to an arc edge (line to arc, arc to arc): **deferred by the owner,
      2026-11-17**, for implicit blends and more pressing work, after Part A's GPU cost. An exact fillet
      there is a line–circle or circle–circle intersection and a shorter arc of the same circle; a fillet
      too big for its edges has no closed-form largest fit. Until then a corner next to an arc stays
      sharp, as documented, and `cornerBlend()` too works between straight edges only
- [x] Task: An example mixing the kinds; WGSL size; the owner's viewer check (Part A's corners)
- [x] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 3: Quadratic Bezier Edges (C6)

- [ ] Task: Proposal: how an edge is marked as a quadratic Bezier (its control point) alongside lines and
      arcs, and how the 2D distance reuses `bezierQuadPointRound`'s closed-form solve
- [ ] Task: Write failing tests (Red Phase): Bezier edges, alone and mixed with lines and arcs, exact against
      the reference, with smooth and sharp joints
- [ ] Task: Implement (Green & Refactor Phase); `.u` along a Bezier edge is its arc length
- [ ] Task: An example; WGSL size (watch the per-step accumulator reads, spec section 5.5); the owner's
      viewer check
- [ ] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 4: Tighter Boxes (C7)

- [ ] Task: Write failing tests (Red Phase): an arc profile's and a Bezier profile's boxes contain the
      reference boundary and are within a stated margin of its true box
- [ ] Task: Implement (Green & Refactor Phase), within whatever Phase 0 found `.box` can hold
- [ ] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 5: Partial Revolve (C3) [checkpoint: 9248b1b]

Taken next after Phase 2, before Phases 3 and 4 (the owner, 2026-11-18).

- [x] Task: Proposal: the signature (angle range, its origin) and the caps. Decided: one angle from the
      sketch plane, flat caps, the exact formula and Go box and frame rules of spec section 4, C3
- [x] Task: Write failing tests (Red Phase) (`eval/profile_partial_revolve_test.go`): exact against the
      reference on a quarter turn, three quarters, a sixth the other way and the arc L (worst 1.3e-5); an
      independent point cloud of the boundary (within its 0.25 mm spacing); surface cells (widest 2.88);
      `certify` behind the sketch plane; a proven box; `.u` and `.v` the full revolve's
- [x] Task: Implement (Green & Refactor Phase): `revolve(profile, axisPoint, axisDirection, angle)` in
      `stdlib/builtin/sweep.gsl`; the frame by the revolve carry rule, now for four arguments too, and the
      full turn's box (`internal/assembler/property`). The sweep is tested by two planes, not an angle:
      `atan2`'s intervals near the axis left the sign undecided and `bounds` could not prove the box.
      Dropping the far cap fails the reference (off by 41 mm). No golden moved
- [x] Task: An example; WGSL size; the owner's viewer check: `examples/profile_revolve.gsl`, a flanged
      sleeve cut open and an arch; 65 KB of WGSL with a slider angle. Checked by the owner: right, and slow.
      Measured (64 evaluations a point, less the floor): the sleeve 56 ms partial against 26 ms full, the
      profile read three times; `profile_cylinder.gsl` 44 ms. Most of the cost is the profile itself, whose
      pools make three passes
- [x] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 6: `sweep()` (C4)

- [x] Task: Proposal: the signature, starting from the remap in `profile_sweep.gsl`. Decided: `sweep(section,
      path)`, the path any exact 2D field; corners as the distance makes them; Go frame and box rules (spec
      section 4, C4)
- [ ] Task: Closed planar paths, exact
  - [ ] Write failing tests (Red Phase): exact against a reference swept section, inside and outside the
        path, at its corners and arcs; frame `.u` along the path
  - [ ] Implement (Green & Refactor Phase); `profile_sweep.gsl` rewritten on it, its picture unchanged
- [ ] Task: Open and non-planar paths, with a stated Lipschitz bound
  - [ ] Proposal: frame transport along the path, and the bound
  - [ ] Write failing tests (Red Phase): the field never exceeds the reference distance and stays within the
        bound; `certify` holds
  - [ ] Implement (Green & Refactor Phase)
- [ ] Task: WGSL size; the owner's viewer check
- [ ] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 7: Caps for Sweeps (C5)

- [ ] Task: Write failing tests (Red Phase): an open sweep's ends capped flat, exact where the path is,
      frames across the rim
- [ ] Task: Implement (Green & Refactor Phase)
- [ ] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 8: Verification and Records

- [ ] Task: Full Go suite with `go vet`; `tool/check.sh`
- [ ] Task: The stdlib's doc comments, `gsl skill`'s text and the language spec describe the new functions;
      CHANGELOG
- [ ] Task: The wheel's Python tests on a wheel built from the branch
- [ ] Task: Metrics recorded
