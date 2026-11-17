# Implementation Plan: Profile Modeling

## Overview

Implements [spec.md](./spec.md) in the order of its section 4, test first. Phase 0 commits the ground truth
the rest is measured against: #45 checked `profileXY` against a brute-force reference only in scratch, so
no committed test says today's profiles are exact. Each later phase adds one candidate and holds it to the
same reference.

Every phase that adds a function to the stdlib checks four things before it is done:

1. **Exact** (or, where the spec allows a bound, within it): the field against the Go reference.
2. **Provable:** `certify` and `bounds` prove on surface cells.
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

## Phase 0: Ground Truth

- [ ] Task: Gate: the development Rust library (`target/release`) is rebuilt on the owner's go; the Go
      packages that evaluate fields (`eval`, `cmd/compile`) do not link against the current one.
- [ ] Task: A brute-force reference for profiles, in Go
  - [ ] Distance to a densely sampled boundary, sign by containment; one helper the later phases reuse
  - [ ] Today's `profileXY(points)` and `profileXY(points, curvatures)` agree with it to about 1e-6: an L, a
        star, a square with a hole (`max` by hand), the arc L of `profile_cylinder.gsl`, and a loop with a
        collinear point and a tangent-continuous joint
  - [ ] `extrude` and `revolve` of those profiles agree with the reference composed in 3D
- [ ] Task: Provability and size, recorded
  - [ ] `certify` and `bounds` prove on surface cells of each profile and of an extrude with a rounded rim
  - [ ] The WGSL size of each `examples/profile_*.gsl`, recorded as the budget later phases compare against
- [ ] Task: Recheck spec section 5's constraints 1–4 on `main` and record which still hold in the spec
- [ ] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 1: Several Loops in One Profile (C1)

- [ ] Task: Proposal: the signature (a list of loops, each points or points and curvatures), how nesting is
      checked, and what a loop that does not nest reports
- [ ] Task: Write failing tests (Red Phase)
  - [ ] Outline with two holes, and an island inside a hole, against the reference
  - [ ] Nesting check: crossing or touching loops are refused with a message naming them
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

## Phase 5: Gate: `cylinderCapped` Upstream

- [ ] Task: The owner's upstream PR making `cylinderCapped` a cylinder rather than a beam has merged; the caps
      of Phases 6 and 8 are modelled on it

## Phase 6: Partial Revolve (C3)

- [ ] Task: Proposal: the signature (angle range, its origin) and the caps
- [ ] Task: Write failing tests (Red Phase): exact against the reference, caps included; `.u` over the
      range; box
- [ ] Task: Implement (Green & Refactor Phase), within spec section 5.3 as `revolve` is
- [ ] Task: An example; WGSL size; the owner's viewer check
- [ ] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 7: `sweep()` (C4)

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

## Phase 8: Caps for Sweeps (C5)

- [ ] Task: Write failing tests (Red Phase): an open sweep's ends capped as Phase 5's cylinder is, exact
      where the path is, frames across the rim
- [ ] Task: Implement (Green & Refactor Phase)
- [ ] Task: Phase Verification & Checkpoint (Refer to `workflow.md`)

## Phase 9: Verification and Records

- [ ] Task: Full Go suite with `go vet`; `tool/check.sh`
- [ ] Task: The stdlib's doc comments, `gsl skill`'s text and the language spec describe the new functions;
      CHANGELOG
- [ ] Task: The wheel's Python tests on a wheel built from the branch
- [ ] Task: Metrics recorded
