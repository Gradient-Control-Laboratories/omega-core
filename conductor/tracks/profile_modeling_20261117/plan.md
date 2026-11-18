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

## Phase 1: Several Loops in One Profile (C1)

- [x] Task: Proposal: the signature (a list of loops, each points or points and curvatures), how nesting is
      checked, and what a loop that does not nest reports. Decided: a generic `nested(fields)` taking
      each loop as its own `profileXY` field, nearest surface wins; the winding rule documented and pinned
      by a test, not checked (spec section 4, C1)
- [ ] Task: Write failing tests (Red Phase)
  - [ ] An outline with two holes and an island inside one, polygon and arc loops mixed, against the
        reference; the same loops in another order; two separate islands
  - [ ] `extrude` of it against the reference; surface cells tight; `certify` proves clearance
  - [ ] A 3D solid with a void against an analytic reference
  - [ ] A mis-wound loop's result pinned: the sign holds, the magnitude is that loop's own distance
- [ ] Task: Implement (Green & Refactor Phase); `.box` and `.u` across loops
- [ ] Task: An example, its WGSL size against the budget, and the owner's viewer check
- [ ] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 2: Corner Blends, per Vertex (C2)

- [ ] Task: Proposal: how a vertex names its blend (none, arc fillet, line chamfer, field blend) and its size,
      and what is refused (a fillet larger than its edges allow)
- [ ] Task: Write failing tests (Red Phase)
  - [ ] Arc fillets and line chamfers generated as edges are exact against the reference, convex and reflex
  - [ ] A field blend stays within its stated bound and keeps the rest of the loop exact
- [ ] Task: Implement (Green & Refactor Phase)
- [ ] Task: An example mixing the kinds; WGSL size; the owner's viewer check
- [ ] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

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

## Phase 5: Partial Revolve (C3)

- [ ] Task: Proposal: the signature (angle range, its origin) and the caps
- [ ] Task: Write failing tests (Red Phase): exact against the reference, caps included; `.u` over the
      range; box
- [ ] Task: Implement (Green & Refactor Phase), within spec section 5.3 as `revolve` is
- [ ] Task: An example; WGSL size; the owner's viewer check
- [ ] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 6: `sweep()` (C4)

- [ ] Task: Proposal: the signature, starting from the remap in `profile_sweep.gsl`
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
