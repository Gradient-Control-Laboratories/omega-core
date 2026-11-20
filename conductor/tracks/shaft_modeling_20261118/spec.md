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
| End treatments (2026-11-19) | **Finished prisms, and runouts cut after the fold** (section 3.4): segments fold as in Phase 3; a cutter runout is a void cut into the neighbour afterwards; a chamfered end cuts the prism. The Phase 3 API is kept, with end overloads. (B1, the blank in the fold, was tried and dropped.) |
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

### 3.3 The construction: a bridge for every run

Let each segment `S_i` be extensible: `S_i↑` is it unbounded above, `S_i↓` unbounded below, `S_i∞`
both ways. Every **run** of consecutive segments `i..j` contributes the intersection of its members
over the run's span:

    bridge(i..j) = max(S_i↑, S_(i+1)∞, …, S_(j-1)∞, S_j↓)
    shaft        = min( min_i S_i,  min over runs bridge(i..j) )

- **It adds no material, whatever the convexity.** Over each member's own span, the intersection lies
  inside that member.
- **It removes the sheet, point by point.** On a joint plane, where both sides are solid, the run of
  the two is negative. Where only one side is solid (a tooth standing above a smaller neighbour), it is
  positive, and that segment's own end face is a real face. This holds angle by angle, so mixed
  convexity needs no special case.
- **It needs no select.** It is built from min and max only, so it stays interval-provable.
- **Runs, not just neighbours.** Phase 1 first bridged neighbours only (Phase 0's construction), and
  the owner's viewer check found it wanting. A thin segment, whether a collar or a ring groove, then
  leaves the whole section across it no deeper than half its width. For a collar of 16 on a 12
  journal, the axis read −0.55 where the truth is −6.0, so an inset of 0.55 mm slits the shaft. The
  run of all three carries the section common to them straight across.

**Exact outside, although it is a max.** Each extended member is an exact extrusion, and from within
a member's own span the nearest point of the member extended is the member's own (an extension is only
ever reached at its rim). So outside, no bridge reads below the capped segments' min. Inside it is a
bound, never shallower than 1/√2 of the true depth in every case tested. That is the most a square
reflex corner costs: the field reads `max(a, b)` where the truth is `√(a² + b²)`. The remaining
interior error is those corners, where the true contours are arcs (accepted by the owner, 2026-11-18).

In the viewer's contours this shows two ways. Each is a capped segment's end plane counted as boundary
where it is not:

- **Past a groove's rims**, each contour runs straight on by up to its own depth instead of rounding
  the rim.
- **Where a taper meets a diameter at the same radius**, which has no face at all, contours run
  vertical at a constant offset from the joint. At a 45° cone that run is about 0.6 × the depth long.

**The segment closure and the fold.** A segment is a factory returning a closure of two extension
lengths: `seg(0, 0)` capped, `seg(big, 0)` unbounded below, `seg(0, big)` above, and `seg(big, big)` both
ways. Bridges are pooled by min, and max distributes over min, so all runs fold in one pass. The
accumulator is `[shaft so far, every run still open]`:

    acc = reduce((a, seg) {
      openRuns = a.y;
      return [min(a.x, seg(0, 0), max(openRuns, seg(big, 0))),
              min(max(openRuns, seg(big, big)), seg(0, big))];
    }, segments, [big, big]);

With baseline positions the fold needs no position at all. WGSL grows linearly, by about 1.5 KB a
segment (section 7).

**Rejected: gating the bridge to the inside.** A variant counting the bridge only where the plain
union reads ≤ 0 (`max(bridge, m > 0 ? nothing : -nothing)`) is no more exact, since the ungated bridge
is already exact outside. It also costs proofs: on a joint cell the plain union's interval straddles
0, the select stays undecided, and the cell no longer proves solid (0 of 3 against 3 of 3).

### 3.4 Caps: finished prisms, and runouts cut after the fold (decided 2026-11-19)

A segment declares how it **ends**. The default is a plane. The other end treatments are:

- a **runout** at a cutter's radius, for a hobbed gear or a milled spline;
- a **chamfered end**, for teeth;
- a **thread runout**, after `thread.runout` (Phase 5).

**A runout is a cut, not part of the segment's solid.** A hob or disc cutter of radius Rc cannot
stop at the segment's end plane: where the neighbour is larger than the root, the gaps run on into the
neighbour's material, their floor rising from the root along the cutter's arc, centred at
(root + Rc, the end plane) in the half-section, until it breaks out of the neighbour's diameter. Within
the cutter's reach the neighbour is cut to the segment's profile, tips included.

**The construction (owner's choice: the hybrid).** The fold is Phase 3's: every segment, round or not,
enters it as its finished prism, capped and extended. Only a runout is cut after the fold, as `body`
cuts the end holes:

    shaft = max(fold, -min(cuts), -holes)
    cut   = max(-section, min(bottomDisc, topDisc))     asked only of a segment with a runout end
    disc  = length([r - (root + Rc), z - endPlane]) - Rc, gated off at a plane or chamfered end

- The API stays as Phase 3 left it: `shaft.gear(...)` and `shaft.spline(...)` gain overloads with two
  end arguments, `shaft.planeEnd()`, `shaft.runout(cutterDia)` or `shaft.chamferEnd(size)`, e.g.
  `shaft.gear(z, m, pa, helix, from, to, shaft.planeEnd(), shaft.runout(80.0))`. An end is a value
  `[kind, size]`, not a closure. Segments without named ends run Phase 3's code unchanged.
- The segment closure answers one more question (`want` 3), its cut. Every segment that cannot run out
  answers `nothing` before it evaluates anything: the compiler passes a callee's arguments even when it
  ignores them, so the short circuit has to sit in the closure.
- A chamfered tooth end cuts the prism at 45 degrees by its size, as the blank is chamfered before the
  teeth are cut.

**Rejected: B1, the blank in the fold (tried 2026-11-19).** Non-round segments folded as round blanks,
with their teeth, flats and gaps cut afterwards. The surface was right, but the field gained phantom
faces: the cut's end cap where nothing stands beyond it (contours stepping at a flat's end, up to
1.5–3 mm short), and bridge material that the cut removes (up to 1.4 mm short beside a hex). Removing
them returns the fold to the finished prisms. It was also slower: on the CPU, 94 against 90 ns a point
for the pinion and 37 against 27 for a hobbed runout, with the larger shader (105,041 against 102,557 B).

**Threads (Phase 5, approved 2026-11-20).** `shaft.thread(major, pitch, from, to)` is `thread.male`, the
exact ISO 68-1 field, as a prism segment; it engages `drill.tap` of the same size flank to flank. Its
ends take the same values, and one more: `shaft.threadRunout(length)`, the runout a lathe tool or a die
leaves as it retracts, the groove's floor rising in a straight line from the root to the major diameter
over `length` (ISO 4755 gives about 2.5 pitches). It is cut after the fold as a cutter's runout is, a
line in the half-section in place of the cutter's disc. A thread-milled thread can use
`shaft.runout(cutterDia)`.

As built: the runout's cut is the groove (outside `thread.male`) within the triangle the line closes in
the half-section, past the end plane and inside the major diameter; unbounded, it would cut the groove
back through the whole shaft. A neighbour wider than the major within the runout's length is not
grooved above the major, where a real tool would crash. In `shaft.gsl` the thread package is imported
as `isoThread`, since `thread` names the segment there.

**Joint blends (Phase 6, approved 2026-11-20).** A joint names its blend by an entry in the segment
list, between the two segments it blends: `[shaft.diameter(20, 0, 20), shaft.fillet(1.0),
shaft.diameter(30, 20, 40)]`, and `shaft.chamfer(c)` the same way. The blend is the union blend of the two capped
segments, `unionRound(A, B, r)` or `unionChamfer(A, B, c)`, min'd into the fold, so the bridges still
close every joint. At a square shoulder between round segments the two fields are exact and orthogonal,
so the circular blend is the torus fillet's arc exactly. A union blend fills concave corners only; a
convex edge (a shoulder's rim, a diameter stepping into a narrowing taper) stays an end treatment. Limits:
the size is at most the step's height and either segment's length. Inside, near the blend, the field reads
deeper than the distance: an edge treatment's exception to the unit gradient. The fold carries the
previous capped segment and the pending blend, so its accumulator grows; that cost is measured.

As built: GSL could not carry the pending blend. A number in the accumulator compiles to a runtime
select, an accumulator cannot mix numbers with a closure, and the fold body can neither make a closure
nor see its list. So the closure pass was extended (the owner's call): a reduce over functions passes
the element before this one as an optional fourth callback argument, the first being handed itself. At
each segment the fold asks `prev(0, 0, 6)` for a fillet's radius and `prev(0, 0, 7)` for a chamfer's
size; constant arguments inline, and the constant fold prunes the select. The blend stands in for the
segment's own term, `unionRound(shaft so far, segment, r)`, which is never above the min of the two, so
a plain joint folds back to exactly the old expression: the shipped shafts' WGSL is unchanged but for
the numbering of temporaries. Five fillets on a six-segment shaft cost 1,399 B (12,710 to 14,109). The
accumulator did not grow. Segments with named ends answer nothing past want 3, with the constant test
outermost so it folds even when the ends come from sliders.

**Cranks (Phase 8, approved 2026-11-20; Phase 7 skipped for now by the owner).** A crank pin is
`shaft.eccentric(dia, offset, angle, from, to)`: a round `dia` across whose axis stands `offset` from the
shaft's, at `angle` in radians from +x about +z; exact, a cylinder about a shifted axis, and a multi-throw
crank is its pins at their angles. A web is `shaft.web(mainR, pinR, offset, angle, from, to)`: the convex
hull of a circle of `mainR` about the shaft's axis and one of `pinR` about the pin, exact by the two-circle
hull's closed form; counterweights and other outlines stay `shaft.section`. The bridges join eccentric
neighbours unchanged, and `shaft.fillet` between pin and web is the exact arc, the pin's side being square
to the web's face. Tested as Phase 6: exact outside against the min of the capped prisms' exact
distances, a unit gradient outside, no sheet on the joints, the pin fillet's arc.

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

**Phase 3 depends on gears for the gear segment:** `gears.gearProfile` is not yet a bound (plan, Phase 3), so
neither is `shaft.gear`; the fix is its own PR off main (a bounded involute flank, and the normalized
chamfer for its tip). The shaft adds nothing to it: a helix's twist is divided out.

**Later, recorded:** knurl as geometry; cross holes and male keyseats; relief threads; an exact
revolved envelope for the turned diameters, once profile C2 lands; an **exact interior** (the owner,
2026-11-18). For the last, the true contours at reflex corners are arcs. A candidate is to build the
interior from real boundary pieces only: sides, end caps, and at each joint the shoulder, the
symmetric difference of the neighbours' end sections, whose 2D distance
`max(0, −max(sA, sB), min(sA, sB))` is exact for exact sections. It needs more from a segment than
`seg(eDown, eUp)`.

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
- A **list of functions must be visible where it is folded**: written at the call, bound to a name, or
  handed on through parameters to the fold (a lone function too, since Phase 2). It cannot come back
  from a call.
- **Synthesized names are unique** since Phase 2. Before, a counter per base let one base's numbered
  name equal another's plain one (`gcl_closure_gsl_anon_2`). The later function replaced the earlier,
  silently, whenever one function folded two lists or one list twice.

From the linker (found in Phase 1, predates this track):

- An **import alias that matches a stdlib package's name resolves to the stdlib package**:
  `import gears "geom/gears"` then `gears.f()` reports f "not found in package gears". A
  non-stdlib alias works. Once `mechanical/shaft` existed, Phase 0's throwaway package, imported as
  `shaft`, hit this, so it is imported as `spike` now. Not fixed here.
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

**Phase 1 (2026-11-18): runs, not neighbours.** The owner's viewer check of a collar showed a shallow
sheet under it (section 3.3). With a bridge for every run:

- the collar's axis reads −6.00 (truth −6.03) and a thin groove's −5.00 (truth −5.00);
- outside stays exact;
- inside, the field is at least 1/√2 of the true depth over every grid point of twelve stacks;
- WGSL at 2 / 4 / 8 / 16 segments (with tapers) is 7338 / 10447 / 16323 / 27935 bytes;
- the eight-segment example goes from 19553 to 23112 bytes.

**Decided (owner, 2026-11-18): fix the closure pass.** Phase 2 keeps drill's intent, the shaft taking
feature lists. **Done in Phase 2:** lists already followed a parameter into another function's fold;
what the spike hit was a LONE function (`drill.centerHole(...)`) forwarded across packages, which did
not fold as a list of one. `paramIsFunctionList` now counts a parameter handed on to a callee's
function-list parameter, recursively.

## 8. Verification

- **Distance:** exact outside wherever the segments are exact, and a bound inside, against the Go
  reference.
- **No sheet:** no interior zero, sampled across every joint plane, including the mixed-convexity cases.
- **Provability:** `certify` and `bounds` prove on surface cells.
- **Backends:** each example compiles to WGSL under a size budget, and the viewer shows what the CPU
  evaluates.
- **Tests:** in Go.
