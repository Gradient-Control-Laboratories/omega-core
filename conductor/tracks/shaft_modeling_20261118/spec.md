# Specification: Shaft Modeling

> Spec agreed 2026-11-18. Section 2 records the decisions behind this track. Section 3 sets out the
> construction everything depends on, and section 7 what Phase 0 found when it was put through the real
> closure pass. Section 5 gives the work in order.

## 1. Why

Holes and fasteners were built as the first two cases of a **generalized shaft**. A hole is a stack of
coaxial features cut into a part. A screw is a stack of coaxial features standing proud of a head. A
shaft is the general case: a stack of features along an axis, each with its own section, and a hole
at each end. One vocabulary then covers turned parts, gear shafts, splined and threaded shafts,
crankshafts, and the seats that bearings and bushings ride on.

The drill library was written for this. Its features are factories that return a closure of the mouth
plane, which is "what will let the shaft hand its end holes a plane its own fold computed"
([drill.gsl](../../../stdlib/mechanical/drill/drill.gsl), the note above Point Angles).

## 2. Decisions (2026-11-18)

| Topic              | Decision                                                                                                                                                                                                                        |
| :----------------- | :------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Model              | A shaft is three stacks: a list of **segments** along +Z, and a `drill` hole stack at **each end**.                                                                                                                              |
| Transitions        | **The bridge, everywhere** (section 3). Every segment, round or toothed, is an entry in the stack. Nothing waits on the Profile Modeling track.                                                                                    |
| Positions          | Both eventually: **baseline** (each segment gives `[from, to]` from one datum) and **chained** (each gives a length, and the fold computes positions and the end planes). Baseline comes first; spike 0(a) found chained viable.      |
| Threads            | Geometry first: `thread.male` bounded to its segment, a true distance. The relief form (a displacement material, the fastener precedent) comes later.                                                                              |
| Package            | `stdlib/mechanical/shaft/`. Segments are factories returning closures, as drill's features are. The folding function is self-contained (section 6).                                                                              |
| In this track      | Cranks (eccentric and section segments); bearings and bushings; a bounding-box override; a simplified model with no teeth, threads or knurling.                                                                                    |
| Deferred (recorded) | Knurl as geometry. Cross holes and male keyseats as voids attached to a segment. An exact revolved envelope, once profile C2 lands. Relief threads.                                                                                |

## 3. The Bridge

### 3.1 The defect

A min-union of segments that meet exactly at a plane leaves a **zero sheet with no sign change** at each
shoulder: on the plane, inside both radii, both capped segments read 0, and so does their `min`. This
was first seen in omega's area-quadrature work as the phantom shoulder disc of a stepped shaft. Nothing on
the outer surface shows it. It does break interior values, surface-cell proofs and anything that offsets
the field: an inset or a shell opens the sheet into a slit. It is the same coincident-surface defect the
drill library's invariant forbids ("no surface of the void may lie in any face it opens onto").

### 3.2 Why the textbook fix is not enough

The usual fix extends the smaller segment into the larger one, so the abutting faces overlap instead of
coinciding. That needs the **convexity** of the step: which side is larger. A shaft does not always
know. Take a gear whose tip circle lies above the next diameter and whose root circle lies below it: at
some angles the gear is the larger side and at others the smaller. A spline, a hex or a D-flat next to
a round can do the same.

### 3.3 The construction

Let each segment `S_i` give its **uncapped** field `S_i∞` (its section, unbounded along the axis) and
its span `slab_i`. At each joint between adjacent segments A and B:

    bridge = max(A∞, B∞, slab over [A.from, B.to])
    shaft  = min( min_i max(S_i∞, slab_i),  min_joints bridge )

- **It adds no material, whatever the convexity.** Over A's span, `A∞ ∩ B∞` is inside A. Over B's
  span, it is inside B.
- **It removes the sheet, point by point.** On the joint plane, where both sides are solid, the bridge
  is negative. Where only one side is solid (a tooth standing above a smaller neighbour), the bridge is
  positive and the segment's own end face is a real face. This holds angle by angle, so mixed
  convexity needs no special case.
- **It needs no select.** It is built from min and max only, so it stays interval-provable. A ternary
  would lower to a lerp and wreck interval bounds.
- **Adjacency matters.** A bridge across a smaller middle segment would add material, so a joint only
  ever bridges its two neighbours. This is why the fold carries the previous segment's uncapped value.

**Proved in Phase 0** (section 7): the field is exact outside wherever the segments are exact, and a
bound inside. There is no interior zero across any joint plane. `certify` and `bounds` still prove on
surface cells.

**Why it is exact outside, although it is a max.** The extended neighbours `A↑` (A unbounded upward)
and `B↓` (B unbounded downward) are themselves exact semi-infinite extrusions, not a radial field
maxed with a slab. Take a point outside the shaft. If it lies within A's span or below it, then it is
outside `A↑`, and `A↑` reads A's own distance there. The same holds for B within its span or above
it. So the bridge never reads below a segment's own distance, and the min is the min of the exact
capped segments.

**The segment closure.** A segment is a factory returning a closure of two extension lengths:
`seg(0, 0)` is the segment capped to its span, `seg(big, 0)` is unbounded below and `seg(0, big)`
unbounded above. The bridge at a joint is `max(prev(0, big), cur(big, 0))`, so with baseline
positions the fold needs no position at all. Its accumulator is `[d, prevUp]`:

    acc = reduce((a, f) {
      return [min(a.x, f(0, 0), max(a.y, f(big, 0))), f(0, big)];
    }, segments, [big, big]);

**Rejected: gating the bridge to the inside.** A variant counting the bridge only where the plain
union reads ≤ 0 (`max(bridge, m > 0 ? nothing : -nothing)`) is no more exact, since the ungated bridge
is already exact outside. It also costs proofs: on a joint cell the plain union's interval straddles
0, the select stays undecided, and the cell no longer proves solid (0 of 3 against 3 of 3).

### 3.4 Caps

A segment declares how it **ends**. The default is a plane, handled by its slab. The other end
treatments are:

- a **runout** at a cutter's radius, for a hobbed gear or a milled spline;
- a **chamfered end**, for teeth;
- a **thread runout**, after `thread.runout`.

### 3.5 What falls out of the construction

- A retaining-ring groove or a shoulder undercut is a short, smaller segment.
- A crank pin is an **eccentric** segment: its section's axis is offset, and the bridge does not care.
- A crank web is a **section** segment: a 2D field, such as a `profileXY`, read in the shaft's XY.

## 4. What Exists

On `main` (6f5462b):

- `stdlib/mechanical/drill`: `hole(features, zTop)` folds feature closures, with every depth measured
  from the mouth. `centerHole(pilotDia, csDia, depth)` is "the likeliest occupant of the two holes at
  the ends of a shaft". There are female spline, keyway, dFlat and doubleD features, and
  `axisAt(point, axis)` for placement.
- `stdlib/mechanical/thread`: `male(major, pitch)` is the exact ISO 68-1 field, unbounded along z.
  `runout(zTop, zBottom, outerR)` gives the relief envelope.
- `stdlib/mechanical/gears`: `gearProfile(numTeeth, module, pressureAngle, rotOffset, clearance, k)` is
  the 2D involute section. The helical `gear(...)` is a whole gear body, with bore and standoffs.
- `stdlib/mechanical/joint`: `cylindrical(seatPoint, axisDir, nominalDia, fit)`, the fit vocabulary a
  bearing seat uses.
- `stdlib/builtin/3d.gsl`: `cylinderZ`, `cone`, `coneCapped`, `cylinderCapped`. Profile Modeling holds
  `cylinderCapped` as a beam placed by its end points; its upstream fix is that track's Phase 5.
- `stdlib/material/process/machined`: `knurled(...)`, a knurl as a material height (the teeth are real
  shape only as displacement).
- Shafts in examples are written by hand: the D-cut motor shaft in
  [corexy_printer.gsl](../../../examples/corexy_printer.gsl) and the plain shaft in
  [slider_mechanism.gsl](../../../examples/slider_mechanism.gsl).

Only on the Profile Modeling branches (`modeling-*`), not yet on `main`:

- `nested(fields)` in `stdlib/builtin/booleans.gsl`: one solid from surfaces that do not cross, exact.
- The brute-force profile reference from that track's Phase 0. This track's Phase 0 reuses it if it
  has landed, and otherwise writes its own.

## 5. Scope, in Order

The numbering is the order.

0. **Ground truth and spikes.**
   - A brute-force reference for stepped shafts: the distance to a revolved staircase, and to
     non-round sections sampled densely.
   - The phantom pinned as a failing test: a min-union of abutting capped cylinders reads 0 inside, on
     the joint plane, and an inset opens it.
   - Spike (a): WGSL growth, against the number of segments, of a fold whose accumulator carries
     `[d, prevU]` (baseline, 2 reads) and `[z, d, prevU]` (chained, 3 reads). Profile spec 5.5 saw
     growth that is exponential in the accumulator reads.
   - Spike (b): a list of drill closures passed **through** `shaft` into `drill.hole`. That is two
     frames, and `listBindings` binds a list one frame up.
   - Spike (c): a section segment taking a 2D field without a remap. A remap of a shape that arrives as
     a parameter is a silent no-op.
1. **A coaxial round stack.** Cylinder, taper and groove segments; baseline positions; bridge joints.
2. **End holes.** `shaft(segments, topHole, bottomHole)` passes each end plane to `drill.hole`; the
   bottom stack is flipped. Example: a shaft center-drilled at both ends.
3. **Non-axisymmetric segments.** Gear (from `gearProfile`, with a helix), male spline, D-flat and
   double-D, hex. Mixed-convexity cases: gear tip > neighbour > root, and hex across corners > round >
   hex across flats.
4. **End treatments.** Cutter-radius runouts and chamfered ends (section 3.4).
5. **Threads.** Thread segments from `thread.male`, bounded to the segment, with their runout.
6. **Joint blends.** A fillet or chamfer per joint, as profile C2 is per vertex. A field blend gives a
   bound; its stated limit is tested.
7. **Chained positions.** Lengths in place of `[from, to]`, the fold computing positions and end
   planes. Gated on spike 0(a).
8. **Cranks.** Eccentric segments (crank pins) and section segments (webs). Example: a single-throw
   crank.
9. **Bearings and bushings.** A `stdlib/mechanical/bearing` package. The shaft side is seat segments:
   journal, locating shoulder, retaining-ring groove. The separate bearing and bushing bodies come from
   ISO tables (ISO 15 deep-groove radial bearings first), with `joint.cylindrical` fits.
10. **Box override and simplified model.**
    - A **bounding-box override**: the shaft computes its `.box` from its segments, and the caller can
        replace it.
    - A **simplified model** of the same segment list, with no teeth, threads, knurling or other
        detail. Each segment reduces to its smooth envelope: a gear at its tip diameter, a thread or a
        spline at its major diameter. It is for proofs, meshing and assemblies, following
        `thread.male`'s advice to "prove against the smooth cylinder".

**Later, recorded:** knurl as geometry; cross holes and male keyseats; relief threads; an exact
revolved envelope for the turned diameters, once profile C2 lands.

## 6. Constraints

From the Profile Modeling spec, section 5, rechecked on `main` in Phase 0:

- An accumulator read several times per step unrolls to exponentially growing code (5.5). Not
  reproduced by the shaft folds in Phase 0 (section 7), which grow linearly; recheck for each new fold.
- A stdlib function can be right on the CPU and wrong in WGSL, so WGSL is checked too (5.6).
- A function's internal `remap` carries frame, colour and material only in the narrow
  `return remap(param, coord)` form (5.3).

From the closure pass:

- A **folding function's body is cloned into the caller's file**, so it may not name its own package.
  `drill.hole` is written self-contained and says so, and `shaft` must be too.
- A **captured function** is not supported (`internal/closure/transform.go`, `liftIfForeign`), so a
  segment cannot take a closure as an argument (for example, a cam with a periodic function). A
  captured *field* is fine (spike (c)).
- A **list of functions must be visible where it is folded**: it cannot arrive as a parameter of the
  function that folds it (spike (b)).
- A scalar folds as a list of one, so a lone segment needs no list literal.

From the drill library:

- No surface of a void may lie in a face it opens onto. End holes run past the end planes, and the
  shaft's own end faces bound them.

## 7. Phase 0 Findings (2026-11-18)

Run against a throwaway shaft package injected through the mock source manager
([shaft_phase0_test.go](../../../eval/shaft_phase0_test.go)). The reference for round stacks is exact:
the stack's meridian polygon, mirrored across the axis.

| Check                                  | Result                                                                                                                                     |
| :------------------------------------- | :----------------------------------------------------------------------------------------------------------------------------------------- |
| Phantom, plain min                     | Every joint point reads 0 (111–222 per case); a 1 mm inset opens a slit at 69–178 of them.                                                   |
| Phantom, bridge                        | No joint point reads ≥ −0.001; no slit.                                                                                                    |
| Outside, bridge                        | Exact: worst 3.2e-6 mm over 20,000+ points per case (step down, step up, groove, near-equal diameters).                                      |
| Inside, bridge                         | A bound: never deeper than the truth, never the wrong sign. Worst shortfall 0.1–4.4 mm, at the corners of a step.                             |
| Mixed convexity                        | Hex (20 across flats) beside a round of 22, both orders: exact outside (2.4e-6 mm), solid on the joint, real end faces where the corners stand proud. |
| Surface cells                          | Sound, widest 0.81 diagonals; every joint cell proves solid under the bridge, none under plain min.                                          |
| `certify`                              | Proves clearance beside, beyond each end and in the step's corner, in one evaluation each.                                                 |
| Spike (a): accumulator growth          | Linear in the segment count for every fold. WGSL bytes at 2 / 4 / 8 / 16 segments: plain 6283 / 6742 / 7659 / 9550; bridge 7054 / 8562 / 11782 / 18031; chained (3 components, z read four times through a local) 6777 / 8247 / 11312 / 17483. |
| Spike (a): chained positions           | Viable: equal to its baseline twin, and no growth past linear. Profile spec 5.5's blowup is not reproduced by these folds.                  |
| Spike (b): hole lists through `shaft`  | **Refused.** "A list of functions has to be visible where it is folded": a list that arrives as a parameter does not reach `drill.hole`. Pinned by `TestShaftPhase0_SpikeHoleListsThroughShaft`. |
| Spike (b): voids instead of lists      | Works: the document folds each end's `drill.hole` and hands the shaft the voids. A void handed over with its mouth at 0 is placed by the shaft's own `remap` correctly on the CPU, so the "silent no-op" drill warns of did not occur here. WGSL not yet checked. |
| Spike (c): section segment             | Works: a captured `profileXY` field as a segment's section.                                                                                |

**Decided (owner, 2026-11-18): fix the closure pass.** Phase 2 keeps drill's intent,
`shaft(segments, topHole, bottomHole)` taking feature lists, and opens by teaching the closure pass to
follow a list of functions through a parameter into the function that folds it. Voids are not the API.

## 8. Verification

- **Distance:** exact outside wherever the segments are exact, and a bound inside, against the Go
  reference.
- **No sheet:** no interior zero, sampled across every joint plane, including the mixed-convexity cases.
- **Provability:** `certify` and `bounds` prove on surface cells.
- **Backends:** each example compiles to WGSL under a size budget, and the viewer shows what the CPU
  evaluates.
- **Tests:** in Go.
