# Material composition: what exists, and the gap now closed

A reassessment of how omega composes a surface from what it is made of and how
it was made, written against the libraries actually in the repo.

**This document replaces an earlier draft that was wrong.** That draft proposed
a four-stage record architecture and a new `material/process` package, having
been written without reading `material/process/machined`, `material/coat/coating` and
`material/process/printed` — three libraries that already implement process-over-
substance. Most of what it proposed already exists. What follows is the honest
picture and the single gap that survives it.

## The architecture already in place

Every process library shares one convention:

```gsl
f(baseMaterial, baseColor, ...parameters) -> material, color, height, slope
```

A process takes the substance it is applied to and returns the surface it
leaves. **"Any substance × any process" therefore already works** — that was
the goal the earlier draft designed an architecture for.

| library | what it covers |
| --- | --- |
| `material` | the record, its accessors and modifiers (`withRoughness`, `withLay`, `withDisplacement`, `blend`), and the scalar composition operators |
| `material/process/machined` | subtractive: `turned`, `faced`, `faceMilled`, `ground`, `brushed`, `blasted`, `knurled` |
| `material/coat/coating` | finishes laid over: `anodized`, `hardAnodized`, `blackOxide`, `phosphate`, powder coat, paint, galvanize |
| `material/process/printed` | additive: fused filament, resin, powder bed, laser-melted metal |
| `material/bulk/wood` | solid wood as a growth field, with its own prep / stain / topcoat pipeline |
| `material/bulk/stone`, `weathering`, `natural`, `animated`, `playful` | further material families on the same convention |

### Conventions that matter, and one trap

**Lobe widths are alpha, not roughness.** `alpha = roughness^2`
(`gsl_layAlphas` in `examples/viewer/render/gsl-shading.js`), so `alpha2 = roughness^4` and
slope variance adds in `alpha^2`. `machined.layFinish(base, direction, along,
across)` takes `along` and `across` as **alphas** and returns
`withRoughness(base, sqrt(sqrt(along * across)))`.

The trap: `texture.reliefRoughness(roughness, slope, size)` has its first
parameter *named* `roughness` but every caller passes an alpha — `calm * calm`,
`0.9025` for 0.95, `0.0004` for 0.02 — and wraps the result in `sqrt`. So the
identity it actually implements is

```
roughness_new^4 = roughness_base^4 + hiddenShare * slope^2
```

which is the same convention as `wood.hiddenSlopeFinish`, whose `base` term is
`rough^4 + slopes` with the factor of two folded into its input. **The two
agree.** An earlier reading of this document claimed a power mismatch between
them; there is none.

**Scale separation already exists and is cited.** `texture.reliefRoughness` is
Toksvig 2005, and its hidden-share term is

```gsl
1 - fadeDetail(size) * displacementDrawn()
```

Relief while it can be resolved and drawn, lobe width when it cannot, never
both. The earlier draft presented this as an invariant nothing enforced. It is
enforced, in `texture`, and `machined.grooveFinish` relies on it: *"as they
fade below a pixel their slopes ... widen the lobe across it instead
(texture.reliefRoughness), so they are never counted twice."*

**The stand-in tier is a property.** `machined.grooveFinish` sets
`finish.simple = ...` for the tier that draws no displacement, which is what
wood spells as its `standIn` argument.

**The coordinate warp is already there too.** `machined.faceMilled` indexes its
arcs along `fract((x - reach) / feedPerRev)` with `reach = sqrt(R^2 - y^2)` —
the cut reaches further along the feed the further the point sits from the
cutter's centre. `machined.nosePitch` is the true circular arc,
`sqrt(r^2 - x^2) - sqrt(r^2 - s^2/4)`, not a parabolic approximation.

## The gap, and how it was closed

**The gap was narrower and sharper than "hand-judged".** Only 5 of `coating`'s
12 presets take `baseMaterial` at all. The four film coatings — `powderCoat`,
`wrinkleCoat`, `sandCoat`, `hammertone` — build their coat from scratch with
`material.make(...)` and **never see the substrate**, so a powder coat over
mirror-polished steel is identical to one over as-cast iron.

And there is a physical split the library already honours, which must not be
flattened:

- **Conversion coatings grow out of the metal and follow it.** `anodized`,
  `hardAnodized` and `blackOxide` keep the lay *correctly*, whatever their
  thickness — *"The oxide follows the surface, so the machining beneath shows
  through."* They need no levelling rule and were left untouched.
- **Flowed films pool in the valleys** and level what is finer than they are
  thick. These are the ones that need `filmKeeps`, and the clear ones are the
  "urethane any metal" case.

So the rule is not thickness against feature size alone; it is thickness
against feature size **for films that flow**.

### What landed

`coating.levelled(coat, baseMaterial, build, underSize)` — a flowed film over a
prepared surface. One share, `material.filmKeeps(underSize, build)`, governs
three things at once: the substrate's displacement is cut by it, its lay's
anisotropy by it, and its lobe's variance by its **square**, variance going as
amplitude squared.

`coating.clearCoat(baseMaterial, baseColor, tint, glossUnits, build,
underSize)` — a clear urethane, lacquer or clear powder over any substance,
keeping the part's own colour, metallic and index.

Measured on a turned aluminium bar (feed 0.15 mm, anisotropy 0.830):

| film | keeps | resulting anisotropy |
| --- | --- | --- |
| 0.08 mm urethane | 0.0155 | **0.0129** — the turning all but vanishes |
| 0.002 mm wax | 1.000 | **0.830** — the turning shows through entire |

And the same 0.08 mm urethane keeps 1.2% of a P120 brushed lay but 16% of a
0.5 mm as-cast skin. One film, two substrates, two outcomes; one substrate, two
films, two outcomes. That is what no preset could express.

It is **purely additive**: no existing call site changed and **no golden
moved**. Unlike `material.gsl`, `coating` already reaches the lay row through
`phosphate` and `galvanized`, so there is no new row-liveness cost either.

### Known limits of what landed

- `underSize` is passed in, not derived. A material record carries roughness,
  lay and displacement but not the **scale** they came from, so the author
  supplies it. The alternative — a fifth return value from every process — is
  a breaking change: callers destructure exactly four
  (`mat, col, h, s = machined.turned(...)` in `Material - Coated.gsl`).
- `filmKeeps`'s coefficient was fitted to wood finishes over sanded, sawn and
  planed surfaces. Carrying it to a urethane over metal is judged, not
  measured; the mechanism is the same film flowing into the same valleys.
- The four opaque film coatings now have substrate-aware overloads (step 3),
  but their three-argument forms remain and still cannot show a substrate
  through. That is deliberate: removing them would break every existing
  caller, and a coat sprayed on nothing is a reasonable thing to ask for.
- **The two gloss models are reconciled** (step 4).
  `coating.glossRoughness` now defers to `material.roughnessFromGloss`. Its own
  closed form, `0.044 sqrt(100/GU - 1)` judged from a GGX lobe's tail, read far
  too rough: integrating the GGX/Smith/Schlick BRDF over ISO 2813's receptor
  (4.4 by 11.7 degrees at 60 degree incidence, against the black glass standard
  of index 1.567) and asking what each roughness would *measure* on that meter
  gives, for a stated 20 / 45 / 70 / 85 gloss units, **16.2 / 31.4 / 48.9 /
  64.0** for the old form against **19.9 / 49.5 / 76.4 / 85.8** for the
  material library's. The old form is off by up to 21 gloss units; the fit that
  was numerically integrated rather than judged is within about 7 at worst and
  within 1 at both ends. The same integration reads **89** for a near-mirror
  surface, which is the number `roughnessFromGloss`'s own comment states — an
  independent reimplementation landing on the documented constant.
  Unlike steps 1 to 3 this **changes appearance**, so it carries its own golden
  regeneration.

## Appendix: the gap as first diagnosed

`coating` does address the surface beneath it — carefully, and per preset:

| preset | what it does to what is underneath |
| --- | --- |
| `anodized` | *"The oxide follows the surface, so the machining beneath shows through: its lay and grooves are kept"* — `sqrt(rough^2 + 0.01)` |
| `hardAnodized` | `sqrt(rough^2 + 0.09)`, hazier as it thickens |
| `blackOxide` | *"too thin to hide the machining, so the lay and grooves are kept"* |
| `phosphate` | *"a crust ... that hides the machining"* — `withLay(..., [0,0,0], 0)`, lay zeroed outright |

So the keep-or-hide decision is made **by hand, once per preset**, as a
property of the coating. It is really a property of the coating's thickness
measured against the size of the features under it, and no preset can know
that: the same urethane fills a brushed lay and leaves an as-cast skin
untouched, and a single hand-chosen constant cannot be both.

That rule is `material.filmKeeps(wavelength, build)`, landed in step 1:

```gsl
filmKeeps(wavelength, build) { x = wavelength / (15 * max(build, 0.0001)); return 1 - exp(-x * x); }
```

A 0.08 mm film leaves 1.2% of a P120 brushed lay (0.133 mm features) and
essentially all of a 0.5 mm as-cast skin. One rule, continuous in both
thickness and feature size, replacing a per-preset judgement.

**What blocks it: a material record carries no feature size.** A coating
receives `baseMaterial` and can read its roughness, lay and displacement, but
not the *wavelength* of the relief that produced them — and `filmKeeps` needs
exactly that. Supplying it is the real design question, and the options are:

1. the process returns its feature size as a fifth value alongside material,
   colour, height and slope;
2. the coating takes it as a parameter, which pushes the knowledge onto the
   document author;
3. the record carries it, which costs a slot in every material everywhere.

Option 1 fits the existing convention best and costs nothing to documents that
do not coat anything. Option 3 is what the earlier draft would have implied and
is the worst of the three, for the reason step 1 measured: a row that only some
documents need is paid for by all of them.

## The coat lobe, unchanged and now corroborated

The single-lobe analysis from the earlier draft survives intact, and the
libraries confirm it rather than contradicting it. A dielectric film over a
conductor is two interfaces — a white specular at ior 1.5 plus the substrate's
coloured reflection attenuated beneath it — and one `metallic` scalar cannot
hold both.

`coating` is already working around this by hand: `anodized` sets
`withMetallic(baseMaterial, 1)`, `hardAnodized` sets
`withMetallic(baseMaterial, 0.6 - 0.4 * depth)`, and `blackOxide` and
`phosphate` set it to 0. **That modulation of `metallic` by coat thickness is
precisely the single-lobe approximation of a coat**, arrived at empirically. It
is the strongest evidence that the collapse is workable, and also that a real
coat lobe would let these presets stop guessing.

Deferred, as decided: no second lobe for now. When it lands it is row 4,
`[coatRoughness, coatWeight, coatIor, coatAnisotropy]`, which the record header
already licenses — *"Rows may be appended in later versions"*. The case that
will force it is a **thin** coat over a **coarse** surface, where `filmKeeps`
is near 1 so the substrate's texture survives in full *and* there is a sharp
film highlight: two genuinely different lobe widths. Wax on as-cast iron.

## Duplication worth collapsing, but it is tidy-up

`wood` reimplements what `texture` and `machined` already have:

| wood | equivalent already in the repo |
| --- | --- |
| `hiddenShare(size, standIn)` | the hidden-share term inlined in `texture.reliefRoughness` |
| `hiddenSlopeFinish` tail | `machined.layFinish` — identical roughness `(along·across)^(1/4)` and identical `stretch = (1 - clamp(along/across, 0.1, 1)) / 0.9` |
| `tableSawnHeight`'s arc | `machined.faceMilled`'s `reach`, and `nosePitch` for the profile |

These were arrived at independently and they agree, so converging them is
hygiene rather than a correctness fix. It is worth doing only alongside work
that touches those files anyway.

`wood` goes further than the others in one respect that is not duplication: its
prep, stain and topcoat compose as three layers with a levelling rule between
them, which is the pattern the gap above wants generalised.

## What is dropped from the earlier draft

- **The four-stage record architecture** (substance / prep / treatment / coat).
  The `(material, color, height, slope)` convention already is it.
- **`material/process`, the prep parameter space and the two-primitive
  collapse.** It would be a fourth implementation of `machined`. An unreferenced
  `stdlib/material/process/process.gsl` exists on disk from that
  attempt; it is not registered in `pkg_info.go`, nothing imports it, and it
  should be deleted.
- **The substance record.** A real question, but unrelated to the gap and far
  larger than it.
- **The claim that scale separation needed introducing.** It was already there.

## Done since

- **Retiring the `X`/`XColor` convention.** Listed above as out of scope, and
  then done: metal, glass and physical are 134 zero-argument functions each
  returning `(material, color)`, so a record and the colour that belongs to it
  leave together and a mismatch is unrepresentable rather than merely checked.
  `@material` now requires a function, and `annotation.ValidateReturns` checks
  the pair's shape for the forms type analysis can reach. See
  `doc/material_pairs_golden_verification.md` for the equivalence proof.

## What survives

- **`material.filmKeeps`, `hiddenShare` and `roughnessFromGloss`**, moved out
  of wood in step 1 and verified equivalent — see
  `doc/material_composition_golden_verification.md`. These are the operators a
  levelling rule needs, now reachable from `coating`.
- **The collapse proof**, as a note that wood's four tool-mark fields reduce to
  two primitives, three of them exactly. Useful if wood is ever simplified;
  not a foundation for anything new.
- **The coat-lobe analysis**, corroborated above.

## Order of work

1. **Done.** The scalar operators into `material.gsl`.
2. **Done.** `coating.levelled` and `coating.clearCoat`: a flowed film levels
   the surface beneath it by `filmKeeps`. Additive, no golden moved. The
   conversion coatings were deliberately left alone, being conformal.
3. **Done.** The four opaque film coatings (`powderCoat`, `wrinkleCoat`,
   `sandCoat`, `hammertone`) gained overloads taking `baseMaterial`, `build`
   and `underSize`, routed through `levelled`, each declaring its own levelled
   stand-in. Additive, no golden moved. Measured on a faced plate: the
   three-argument `powderCoat` gives anisotropy 0.000 whatever is beneath it,
   while the overload gives **0.013** over a 0.15 mm feed and **0.632** over a
   1.2 mm one — the same coat, two substrates, two surfaces. `hammertone`
   keeps its metallic 0.5, confirming `levelled` carries the coat's own
   metallic and index rather than the substrate's.
4. **Done.** `coating.glossRoughness` defers to `material.roughnessFromGloss`,
   the two models having been measured against ISO 2813's receptor rather than
   chosen by provenance. The only step so far that changes appearance; one
   golden moved.
5. Only if wanted: converge wood's `hiddenShare` and lobe tail onto
   `texture.reliefRoughness` and `machined.layFinish`. Hygiene, not a
   correctness fix — they were derived independently and agree.
6. **The constant libraries are annotated.** 134 `@material` — metal's 10,
   glass's 8 and physical's 116 — and 8 `@finish` on `material.gsl`'s
   colourless presets. `annotation.Validate` is load-bearing for the first
   time: renaming `brassColor` makes the library stop assembling with its own
   error, raised even for a document that only uses `steel` and never mentions
   `brass`, which also confirms annotations survive deadcoding. Zero goldens
   moved across all 142, so "recorded, not lowered" holds as designed.

   physical's 116 come from the generator, not by hand —
   `tool/physicallybased/gen.go` emits the annotation, and
   `TestGeneratedPackageIsCurrent` guards the two staying in step. The
   snapshot was not refetched.

   What remains is the **function** families: `machined`, `coating`,
   `printed`, `wood`, `stone`, `weathering`, `natural`, `animated`,
   `playful`. They are blocked, and the two blockers turn out to be one.

   `6e52b040` deferred a check that an annotated function returns a material
   *and* a colour, "to the manifest, where types are resolved", because
   `len(fn.Returns)` is 1 for a function forwarding another's result. Types
   alone are not enough. **Return types are inferred per call site from
   argument types**: `types.Analyze`'s `analyzePackage` only *registers*
   functions, and body inference happens in `inferUserCall`, driven from a
   call expression. An uncalled function's parameters and returns are never
   typed — and the uncalled case is exactly the one the check exists for,
   since annotations deliberately survive deadcoding so a browser can list
   materials the document does not draw with.

   So the check cannot be made sound at assembly time without either declared
   parameter types in source (a language change) or assumed ones, which
   `6e52b040` already warned would reject valid presets.

   The two blockers are the same blocker: **you cannot validate an annotated
   function without knowing how to call it, and you cannot preview one
   either.** That points at the consumer rather than the compiler — a browser
   must supply arguments to draw a swatch, and once it has, it has types. The
   compiler cannot know; the thing drawing the preview necessarily does.
   Whatever convention lets a browser call these functions also lets it check
   them, and should be designed once.

   Note for whoever does that: main rebuilt the annotation system, so
   `AnnotationDirective` carries a name and arguments, `@range(min, max[,
   period])` exists, and the grammar admits only the annotations it knows — a
   new kind means a `grammar.go` change as well as a `build_decl` case.

## `.u` and `.v`: a shape's own frame

A shape carries `.u` and `.v`, the coordinates a material reads to know where
it sits on the surface, and a material reads them off the shape by name:

    bar = cylinder(center, axis, 2.0);
    bar.material = material.withLay(steel, bar.u.dp, 0.8);   // streaked round the axis

The frame comes from the shape, so it cannot be handed one that belongs
somewhere else — the drift where one material turned about the origin went on
four bars that were not there. The primitives declare their frames
(`builtin/3d.gsl`); a document can declare or override one with `name.u = ...`.

It is not a surface parameterisation, which is what makes it work: `u` and `v`
stay functions of the 3D point, so a board's `u` field exists everywhere in
space and a cut face still gets the right answer. Nothing about the
cut-consistency these libraries are built on is given up.

### How it is built

`.u` and `.v` are read accessors beside `.dp`, `.second` and `.box`
(`SURFACEU`, `SURFACEV`), expanded by the property pass before backend
lowering, so no backend sees them. A read resolves on its own target with the
colour resolver: a moved shape carries its frame, a union selects each part's
own, and nothing is blended across shapes that do not share one.
`bar.second.u` is the second-closest surface's frame; `bar.u.second` is refused,
since a frame coordinate has no second-closest surface.

A program pays only for the frames it reads. In one that reads none, the carry
pass erases every frame declaration before compiling, and a function that
named its part only to frame it returns the part itself again, so the program
compiles exactly as it did before frames existed.

A library function can read the gradient of a frame coordinate it is handed,
`frame.x.dp` in `machined.turned`'s frame form. Since 2026-10-25 a qualified
call to a function whose body differentiates a parameter is routed to a
companion generated in the library's own package, where its body resolves,
and called by the same prefix with the argument's tangents; companions are
kept by name and arity, so an overloaded name gets one per overload
(`turned__d5`). Before that only an unqualified call, the builtin's
`normalizeField(v)`, reached a companion.

### The conventions

Lengths in mm and angles in radians. `u` is the coordinate that wraps where
one does: a cylinder's angle, a sphere's azimuth. The angle is the ANGLE,
never the arc length — measured on a cylinder about z:

| | error against `cross(axis, offset)` |
| --- | --- |
| `u` = circumferential **angle**, `normalize(u.dp)` | **6e-8**, exact to float precision |
| `u` = circumferential **arc length** (`angle * radius`) | 0.38 to 1.06, wrong |
| `v` = axial position, `normalize(v.dp)` against the axis | **0** |

Making `|grad u|` unit breaks the lay: `grad(theta r) = r grad(theta) + theta
grad(r)`, and the product rule adds a radial component. The lay needs only the
direction, so `u` needs correct level sets, not unit magnitude.

**`grad u` is safe; `u`'s value is not.** `atan2`'s value jumps at its branch
cut but its gradient does not, so a lay from `.u.dp` is smooth away from the
axis, while anything indexing marks along `u`'s value meets the seam. Turning
indexes axially and is fine; knurling is circumferential and is not.

Where faces meet at an edge, their frames turn a right angle rather than
running on: a box cycles each face's in-plane axes, and a capped cylinder's
`v` turns from along the axis on the side to across it on a cap while `u`
runs on.

| primitive | u | v |
| --- | --- | --- |
| `sphere` | azimuth round +Z from +X | polar angle down from +Z |
| `sphere(center, radius, pole, reference)` | azimuth round `pole` from `reference` | polar angle down from `pole` |
| `torus` | azimuth round its axis | angle round the tube from its top — wraps, at the bottom |
| `axis`, `cylinder` and their X/Y/Z forms | angle round the axis | along it from `center` |
| `cone` | angle round the axis | along it from the apex |
| `cylinderCapped` | angle round the axis | side: along it from the middle; cap: from the axis |
| `plane` | in-plane offset from `point` | the other in-plane offset, right-handed with the normal |
| `box`, `boxCentered` | (y, z) / (z, x) / (x, y) on the x / y / z faces | |

### Deferred

- **Blends with their own parameterisation.** `cylinderRounded`,
  `coneRounded` and the like meet their caps along a tangent seam, where a
  frame that runs on may serve better than one that turns a right angle.
  Deferred until a blend's frame is worked out as a whole; they declare none.
- **2D primitives.** Their natural frame is the profile curve's arc length,
  but a length is of little use to a material without the curve's total
  length to scale it by. Deferred until a 2D shape can report that.
- `coneCapped`, `lineSegment`, `octahedron` declare no frame yet.
- **Depth.** The carry pass is one level deep: a frame reaches a part named in
  the document, not one built inside a library function's body.
- **Imported geometry.** STEP faces could supply their B-rep `(u, v)` through
  hodge, so a material on an imported face would inherit the CAD
  parameterisation.
- **Reading `.color` and `.material`.** The same accessor with the colour or
  material resolver would give `.second.color`, the colour of the
  second-closest surface; colour and material stay write-only for now.
- `.second` drops an inline `remap` of a named shape
  (`min(bar, remap(ball, q)).second` reads the ball unmoved). It predates
  frames; a frame read of such a field inherits it.

## Parked: surface operators in the Materials panel

The panel added in this work browses the 134 `@material` surfaces and the 8
`@finish` presets, all of which are zero-argument. The question it leaves open
is what to do about everything else in the libraries that makes a surface and
takes arguments. This is what was measured of it.

### Two kinds get called "a material with parameters"

**Constructors** make a surface out of numbers — `stone.carrara(direction)`,
`natural.terrain(scale, snowBias)`, `coating.powderCoat(coatColor, glossUnits,
peel)`. These really are materials, and their way into the grid is the one
`@finish standard` already took: someone chooses the defaults and names them,
one line each.

**Operators** take a surface and return a modified one. There are **30**, and
in every one the substrate is the first parameter:

| return shape | count | packages |
| --- | --- | --- |
| `material, color, height, slope` | 18 | coating 10, machined 7, printed 1 |
| `material, color` | 4 | playful 2, weathering 2 |
| a material alone | 8 | material 7 (`with…`), coating 1 (`levelled`) |

These are not materials and should not be swatches. `Material - Coated.gsl`
already shows what they are instead — a stack, each one taking the last one's
output as its substrate:

```gsl
turnedMat, turnedColor, turnedHeight, turnedSlope = machined.turned(aluminumMat, aluminumColor, [0,0,0], barAxis, 0.15, 0.8);
clearMat, clearColor, clearHeight, clearSlope     = coating.anodized(turnedMat, aluminumColor, [1,1,1]);
clearBar.material = clearMat;
clearBar.color    = clearColor;
```

So in the panel an operator is a **wrapping** edit, not the replacing edit the
swatch grid performs.

### The preview problem dissolves here

An operator cannot be rendered as a swatch, because an uncalled parameterised
function is never typed — the same reason `ValidateReturns` cannot check one.
But an operator does not want a swatch: "brushed" in the abstract says less
than brushed on *your* part, with your axis and your lighting, which the viewer
shows live the moment the source changes. The viewer is the preview. The limit
that shaped the grid turns out not to bind on the stack.

### Naming cannot find them

Worth recording because it was tried. The family operators can be found by
their first parameter, `baseMaterial` or `coat`. The `material` package's seven
modifiers name theirs `m` — and scanning for that pulls in `matrix.transpose2`,
`matrix.inverse4` and `vector.multiplyPoint`, because `m` is a matrix in half
the libraries and a material in the other half. Neither does arity distinguish
them: `material.roughness(m)` and `material.withRoughness(m, rough)` are both
one parameter in and one value out, and one is an accessor returning a scalar
while the other is an operator returning a material.

Discovery therefore needs an explicit mark, which is the same conclusion the
`@material` annotation reached for the opposite reason: a contract that depends
on spelling is not a contract.

### What an annotation would have to declare

The return arity, because it cannot be inferred for a parameterised function.
The doc comments do currently state it, and state it **correctly in all 22
operators that state one** — but prose is exactly the fragile arrangement that
naming was. So the form would be something like

```gsl
@surface(4) brushed(baseMaterial, baseColor, direction, grit) { … }
```

Annotation arguments already parse (`@range(…)` uses them). Adding a kind means
one alternation in `parser2/grammar.go`, a case in `annotation.ValidateTarget`,
and a validator in `internal/assembler/annotation` — then 30 library annotations,
the generator emitting the operator list, and the stack UI.

### Two routes, smallest first

1. **The seven `material.with…` modifiers only.** Uniform: arity 1, substrate
   first, plain scalar arguments, one package — `withRoughness(m, rough)`,
   `withMetallic`, `withIor`, `withEmissive`, `withTransmission`,
   `withDisplacement`, `withLay`. A roughness or metalness slider on any
   assigned material covers most of what is wanted day to day, needs **no
   language change**, and proves the stack UI before 30 functions and the
   parser are touched.
2. **Annotate all 30.** The complete answer, and the only one that reaches
   `brushed`, `anodized`, `knurled` and the rest. A language-surface change,
   and it wants its own branch and its own review.

### The parameters should not get a UI of their own

Whichever route, a chosen argument should be written into the document as a
top-level variable with `.min` and `.max`, because the **existing Parameters
panel already turns one of those into a live slider**:

```gsl
boltRough = 0.15;
boltRough.min = 0.0;
boltRough.max = 1.0;

boltSurface, boltSurfaceColor = metal.steel();
bolt.material = material.withRoughness(boltSurface, boltRough);
bolt.color = boltSurfaceColor;
```

The Materials panel makes the structural edit once, and its controls under
each effect adjust the value afterwards, persisting it in the document. The
**datum** arguments — `axisOrigin`, `axisDir`, `direction`,
`buildDir`, `flowOrigin` — are not sliders at all: they route to the point and
normal pick tools the viewer already has, since a turned finish's lay follows the
spindle axis and that is a pick, not a number.

## Planned: presets now, parameters later

The constructors that take only numbers and colours now each have one
`@material` preset in a "Ready-made" section at the end of their library, with
the arguments the examples use: `stone.greyGranite`, `stone.whiteMarble`, the
coatings (`redPowderCoat`, `blackWrinkleCoat`, `greenSandCoat`,
`blueHammertone`, `newGalvanized`, `agedGalvanized`), `printed`'s two nylons,
`weathering.rustySteel` and `patinatedBronze`, `playful`'s checker, polka dot
and cell skin, and `natural.rippledWater` and `glowingLava`. A preset is a
starting point that the browser can show, not the end of the design.

What is deliberately left for later:

- **Control over a preset's parameters.** The plan is the one above: the
  browser rewrites the preset as its constructor call, with each chosen
  argument a top-level variable with `.min` and `.max` that the Parameters
  panel already turns into a slider. Until then a preset is edited by hand, by
  replacing it with the constructor it names.
- **Colour as a parameter.** A coating's colour is already separable — the
  Bodies panel's colour chip replaces `.color` and keeps the coat's texture —
  so the presets keep the examples' colours and are named by them. Once
  parameters arrive, a coat's colour is one more of them, and the colour-named
  presets may collapse into one per texture.
- **Datum constructors** (`stone.carrara(direction)`, `printed.fdm(buildDir,
  …)`, the wood species), which wait on frames rather than numbers, and
  constructors taking a function (`playful.confetti`, `rainbowSheen`) or a
  field (`natural.terrain`'s altitude). *Taken up 2026-10-25 for wood:* since
  Adjust lifts a baked datum into a point and an arrow gizmo, each of the 50
  species has a zero-argument `@material` overload at its default look with
  the stem through the origin along +Z, in a Ready-made section of
  `wood.gsl`; `TestEveryWoodSpeciesIsBrowsable` holds the library to it.
  *A second pass the same day:* stone's three marbles (honed, veins along
  +Z), a glinting granite and flagstone; `printed.fdm` and `resin` (build
  along +Z through the origin); `animated.ocean`, `rain`, `mudPot` and
  `sandRipples`; `natural.lava`, `fogMaterial`, `cloudMaterial` and
  `terrain`; `playful.polkaDotDrift`. **What cannot be preset:** a
  constructor that takes a function (`animated.river`, `iceFloes`,
  `lavaLake` with a velocity; `playful.confetti`, `rainbowSheen`,
  `rainbowCycle` with a palette). `types.Analyze` types an annotated
  zero-argument function standalone, before any call site, and a
  function-taking constructor has no generic form by then: the closure pass
  specializes it per call, so the preset's call finds "no function named
  river with 3 parameters". That is a closure-pass gap, not a library one.

The swatches are drawn on a sphere 10 cm across (`tool/materialswatch`), since
a preset's features are in mm: a 4 mm granite grain fills a sphere a
millimetre across with one colour. A material whose colour varies records the
colours that span it in the catalog, and the Bodies panel's chip shows them as
a gradient.

## Design: parameters on materials

Decided 2026-10-24. This takes up the two deferrals above, control over a
preset's parameters and colour as a parameter, and settles what a datum
parameter is. Scalars, colours, points and directions; frames where a material
needs one.

### What the language already has, and does not use

- `@range(min, max)` **on a function parameter** parses, validates
  (`internal/assembler/propval/validate.go`, `validateFnType`) and is consumed by
  nothing: `Parameter.range` is cloned and carried and never read. No library
  uses it. Every constructor's range is prose — "`peel` from 0 (flat) to 1",
  "`layerHeight` 0.1 to 0.3", "`glossUnits`: PLA about 40, PETG 70, matte 10".
- `@range`, `@point`, `@transform` and `@spline` **on a top-level variable**
  each give a live control: the slider, the point gizmo, the frame gizmo with
  its rotation arcs. They are lowered to `uiPoint`/`uiTransform`/`uiSpline`
  calls (`parser/internal/annotation.TargetBuiltin`), and the wgsl build
  lifts each component to a shader uniform (`codegen/build.go`, `uiParams`),
  so a drag never recompiles.
- There is no direction gizmo, and no colour control anywhere but the Bodies
  panel's chip, which writes `body.color` alone.
- `@point` and `@transform` on a *parameter* are refused by one case in
  `ValidateTarget` and one literal comparison in `parser2/parse.go` (the
  parameter loop accepts only `range`). Not a structural limit.

### The spine: one vocabulary, two sites

The same annotations do two jobs. On a **constructor parameter** in a library
they declare what kind of control the argument wants. On a **top-level
variable** in a document they give it that control.

| kind | on the parameter (library) | written into the document | control |
| --- | --- | --- | --- |
| scalar | `@range(0, 1) peel` | `boltPeel = 0.3;` with `.min`/`.max` | slider — exists |
| colour | `@color coatColor` | `@color boltCoat = [0.804, 0.098, 0.098];` | chip row in the Parameters panel — new |
| point | `@point center` | `@point boltCenter = [0, 0, 0];` | point gizmo — exists |
| direction | `@direction(axisOrigin) axisDir` | `@direction(boltOrigin) boltAxis = [0, 0, 1];` | arrow gizmo — new |
| frame | `@transform axis` | `@transform boltGrain = [...];` | frame gizmo — exists |

Two kinds are new everywhere, `@color` and `@direction`; `@point` and
`@transform` become legal on parameters. The alternative, inferring a kind from
the argument's value and the parameter's name on the catalog side, was
rejected for the reason the operator section gave: a contract that depends on
spelling is not a contract, and a vec3 does not say whether it is a colour, a
point or a direction.

**A direction has no place of its own, so `@direction` takes one.** Its
argument is the point the arrow is drawn from: a `@point` variable or a
constant in a document, and on a parameter the name of the sibling parameter
that anchors it, `solidWood(@point axisOrigin, @direction(axisOrigin)
axisDir, …)`, recorded as `Parameter.anchor`. Without one the arrow is drawn
from the origin. The anchor is not a handle: it is the point's own gizmo when
it is a `@point`.

**`@color` and `@direction` lower exactly as `@point` does.** Each becomes a
builtin call, `uiColor(vec3)` and `uiDirection(vec3[, anchor])`, whose three components
the build lifts as uniforms like `uiPoint`'s, and which the compile result
reports in `uiControls[]` with its kind and span. So a colour chip and a
direction arrow are both live without a recompile, by the plumbing the point
already has. A colour literal is sRGB-encoded, the `color` library's and the
Bodies chip's convention; the chip row converts the same way.

### How it flows

1. **Library.** Constructor parameters get annotations. The prose ranges
   become `@range`, which the docs generator (`preamble/libraries.go`) can then
   print instead of relying on the sentence. Parameter annotations are
   recorded on the `Parameter`, never lowered, so no golden moves.
2. **Catalog.** `tool/materialswatch` emits, per preset, an `adjust` record:
   the constructors it calls and their annotated parameters — name, kind,
   range, anchor, and the preset's own argument as the default, evaluated
   with the constant evaluator so `color.gray` resolves — plus the preset's
   body as the expansion template, cut from the library source by the parsed
   program's spans with the lifted arguments and the locals replaced by
   placeholders. Only annotated parameters are lifted; a plain argument stays
   the preset's literal, so a metal (whose `finish(rough)` declares nothing)
   has no record and nothing to adjust. The panel never guesses an arity: the
   template already destructures the constructor's returns, or inlines the
   calls a preset returns directly. `-manifest-only` rewrites `catalog.json`
   against the committed sheet without rendering.
3. **Panel.** Applying a surface preset keeps writing the one-line form; an
   effect's layer is opened as it is picked (see "Controls on a layer"). An
   **Adjust** affordance in the readout, shown when the selected body's surface is a
   preset with parameters, rewrites it as the constructor call with one
   top-level variable per parameter, named `<body><Param>`:

   ```gsl
   @color boltCoatColor = [0.62, 0.06, 0.05];
   boltGlossUnits = 85;
   boltGlossUnits.min = 0;
   boltGlossUnits.max = 100;
   boltPeel = 0.7;
   boltPeel.min = 0;
   boltPeel.max = 1;
   boltSurface, boltSurfaceColor, boltHeight, boltSlope = coating.powderCoat(boltCoatColor, boltGlossUnits, boltPeel);
   bolt.material = boltSurface;
   bolt.color = boltSurfaceColor;
   ```

   From here the sliders under the base in the Materials panel and the viewer
   take over; the panel makes the structural edit once. The Parameters panel
   leaves these variables out (`panelVariables`). The
   surface label reads `coating.powderCoat`. For a preset with one `@color`
   parameter (the four coats, the two nylons) the catalog also carries a
   **coat basis**, two renders with that colour at 0 and at 1 like the
   finishes' (`coat.go`, `-coat-basis`), so the row's thumbnail is the preset
   drawn in whatever coat colour the document declares, read from the
   variable's literal (`adjustedColorOf`), the chip shows that colour, and
   picking one on the chip rewrites the declaration rather than replacing
   `.color` with a flat literal. The other parameters stay the preset's in
   that thumbnail. A preset with several colours, a granite's minerals or
   lava's crust and glow, mixes them over the surface and gets no basis; its
   row is marked as coloured by the document. Picking another preset
   afterwards collapses the destructure back to the pair and removes the
   variables only its call read, so trying surfaces after adjusting leaves no
   sliders behind. Expanding every surface pick was rejected: documents stay
   one line per surface until someone wants the knobs. Layers went the other
   way on 2026-11-06, for the reason under "Controls on a layer".
4. **Defaults.** Every parameter's default is the preset's own argument, the
   datums included, so the document draws exactly as it did before the
   expansion. A direction whose parameter names an anchor is written
   `@direction(<body><Anchor>)` when the anchor was lifted too, and is drawn
   from the origin otherwise. The compile result carries no per-body bounds
   today, so nothing defaults to a body's centre; that waits on the bounds.
5. **`material-assign.js`** gains the expansion as another pure string → string
   function, tested under Node like the rest.

### Datums are two different things

Turned, faced and knurled take the *part's own* axis. That is the shape frame,
`bar.u` and `bar.v`, already built above — not a parameter, and the library
should stop asking for it once the carry pass reaches far enough. Since
2026-10-25 `turned` and `knurled` have frame forms that take it as one
argument, `machined.turned(steel, steelColor, [bar.u, bar.v], 0.15, 0.8)`: the
lay is `u`'s gradient and the grooves fall along `v`, so on a capped cylinder
one call turns the side and faces the caps, since a cap's `v` runs out from
the axis. A function reads the gradient of a frame coordinate it is handed
(`frame.x.dp`), which is what makes the form possible. The axis forms stay
for the examples; `faced` has no frame form, since a plane's frame is not
polar. Build
direction, grain axis, runoff flow and marble bedding are *process or stock*
choices extrinsic to the part. Those are the datums that want a gizmo, and the
only ones this design puts one on.

Wood is the case for the frame. A board is a distance-plus-uv frame and a tree
an axial-plus-uv one; both are origin, axis and roll, which is what the
`@transform` gizmo already edits, and the roll arc sets the grain. For a
direction-only parameter the frame was rejected: a 4×4 literal is sixteen
numbers where an author writes `[0, 0, 1]`, so `@direction` stays its own kind.

### Change sites

- `parser/internal/annotation`: kinds `Color`, `Direction`;
  `ValidateTarget` admits `Point`, `Transform`, `Color`, `Direction` on
  `TargetParameter`; `TargetBuiltin` maps the two new kinds.
- `parser/internal/parser2/parse.go`: the parameter loop accepts the
  five kinds, not `range` alone; testdata gains cases.
- `program/program.proto`: `Parameter` gains a kind beside `range`;
  `scope.SetParamKind` records it.
- `internal/types`: `uiColor` and `uiDirection` take a vec3, as
  `uiPoint` does. No value check: an emissive colour exceeds 1 on purpose
  (`natural.lava`'s `glowColor`), and a zero direction is for the gizmo that
  normalises it to refuse, not the compiler.
- `backend/wgsl/internal/lib`: `uiColor` and `uiDirection` join `uiPoint` in `sanitycheck.go`
  and the `uiParams` extraction; `uiControls[]` carries the kind.
- `examples/viewer/render/gsl-ui.js`: kind `direction` — an arrow, drag on a sphere, snap
  to axes, `formatVec` on commit. `examples/viewer/panels/slider-panel.js`: a chip row for
  kind `color`, sharing the Bodies chip's widget and hex ↔ literal code.
- `tool/materialswatch`: parameters and template per preset, read from the
  assembled program rather than the regex scan.
- `examples/viewer/panels/material-grid.js` and `material-assign.js`: Adjust, and the
  controls under each effect.
- The libraries: annotate the constructors. Colours are the `*Color`
  arguments and `tint`; directions are `direction`, `buildDir`, `axisDir`,
  `feedDir`, `flowDir`, `windDirection`, `up`, wood's `feed`, and `normal` in
  `faceMilled` alone (wood's `normal` is `shape.dp`, a field, not a datum);
  points are `center`, `axisOrigin`, `origin`, `flowOrigin`. A scalar gets
  `@range` only where the prose states a span or the body clamps it; gloss is
  `@range(0, 100)` throughout. The rest (`build`, `underSize`, `chipSize`,
  the water and lava scales, cutter sizes) stay plain until someone states
  their range, rather than carry one invented here. Records (`look`, `prep`,
  `topcoat`, `decay`), species numbers, fields (`wear`), velocities and
  function parameters take no annotation.

### Order of work

1. Language: the two kinds, parameter annotations for all five, proto,
   validators, parser tests. No library or golden changes.
2. Viewer: the direction gizmo and the colour row; `uiControls` kinds.
3. Libraries: annotate constructor parameters. Behaviour unchanged.
4. Catalog and panel: parameters in `catalog.json`, Adjust, unit tests.
5. After: the operators (`@surface(n)`, parked above) use the same parameter
   annotations; swatch thumbnails that follow a changed colour, which needs a
   basis per coat as the finishes have; intrinsic datums read from `.u`/`.v`
   instead of being asked for.

## Design: operators as presets

Decided 2026-10-25, on `blake/material-operators`, stacked on
`blake/material-params`. This takes up the parked operators with the vocabulary
the previous design put in place: the operators' parameters are already
annotated, so what is left is how an operator gets into a document, what values
it carries there, and how a body's stack of them is read back.

### What was measured

30 functions take a surface first and return a modified one, 32 declarations
since `turned` and `knurled` gained their frame forms. Two are helpers no
document calls (`levelled`, and the `layFinish`/`grooveFinish`/`layerFinish`
family), and `withDisplacement` takes a field, which no control can supply. The
28 that remain:

| package | operators | returns |
| --- | --- | --- |
| machined | `turned`, `faced`, `faceMilled`, `ground`, `brushed`, `blasted`, `knurled` | material, color, height, slope |
| coating | `anodized`, `hardAnodized`, `blackOxide`, `phosphate`, `chippedPaint`, `clearCoat`, `powderCoat`, `wrinkleCoat`, `sandCoat`, `hammertone` | material, color, height, slope |
| printed | `laserMelted` | material, color, height, slope |
| weathering | `weathered`, `weatheredRunoff` | material, color |
| playful | `glitter`, `glitterTwinkle` | material, color |
| material | `withRoughness`, `withMetallic`, `withIor`, `withEmissive`, `withTransmission`, `withLay` | material |

No default exists for any of their arguments. The grammar has no default
parameter values, and an operator has no preset to lift arguments from the way
`redPowderCoat` supplies `powderCoat`'s. The constructors answered this with one
named line each, the arguments the examples use, and that is the answer here
too.

### The preset is an overload of the operator's own name

Each operator gets one more overload, in its library's Ready-made section,
taking only what the panel can supply and baking the rest:

```gsl
@surface turned(baseMaterial, @color baseColor, @frame frame) {
  cut, tone, height, slope = turned(baseMaterial, baseColor, frame, 0.15, 0.8);
  return cut, tone, height, slope;
}

@surface withRoughness(m) {
  return withRoughness(m, 0.5);
}
```

- **The name is the operator's.** The Wrap menu lists `turned`, `brushed`,
  `anodized`, which is what someone looks for; a second vocabulary of
  `fineTurned`/`clearAnodized` names would have to be learned. `@surface` marks
  the name, which a recorded annotation already does (`annotatedSymbol`: "an
  annotation marks a name rather than an arity, so every overload of an
  annotated function is covered"; no library has used that yet), and the tool
  picks the preset overload: the one whose every parameter is a slot.
- **The slots** are the three things only the document knows: the substrate
  material (the first parameter, always), the substrate's colour (a `@color`
  parameter, present when the operator reads one), and the part's frame (a
  `@frame` parameter, present where the operator has a frame form). Nothing
  else: an argument the preset bakes is an argument Adjust can lift.
- **The arity is read, not declared.** A preset destructures the operator's
  result and returns every value, so `len(fn.Returns)` states the arity
  syntactically, and the tool refuses a preset whose return is one forwarding
  call. The weathering library already requires that form for another reason:
  tail-forwarding a multi-return call breaks the WGSL property-gradient
  compiler (`weathering.gsl`, the compiler note above `weathered`). So
  `@surface` takes no argument, and the proto's `Annotation` is unchanged.
- **`@frame` is a parameter kind** like `@point`, recorded as
  `Parameter.kind = "frame"` and never lowered. It is refused on a top-level
  variable: a frame is the body's own `[body.u, body.v]`, read, not asked for,
  so there is no control to give it. The panel writes `[bolt.u, bolt.v]` for
  the slot.
- **Datums without a frame form** bake the origin and +Z (`faced`,
  `faceMilled`, `weatheredRunoff`'s `flowOrigin`/`flowDir`) or +X (`ground`,
  `brushed`, `withLay`'s direction), as step 4 chose for the constructors. They
  are `@point`/`@direction` arguments, so Adjust lifts them into gizmos.
- **What stays baked.** A plain argument — `knurled`'s `diameter`,
  `chippedPaint`'s `wear` and `chipSize`, `clearCoat`'s and `powderCoat`'s
  `build`/`underSize`, `faceMilled`'s cutter sizes, `glitter`'s
  `flakeCellSize` — has no annotation and is not lifted. It takes the preset's
  value until its author states a range; a range is not invented here.
- **`material`'s modifiers** need their parameters annotated for Adjust to lift
  them, which the constructor pass did not do: `@range(0, 1) rough`,
  `@range(0, 1) metal`, `@range(0, 1) transmission`, `@range(0, 1) anisotropy`,
  `@color rgb`, `@direction direction`. Those spans are the material record's
  own documented fields, not prose about the modifiers. `index` and
  `absorptionDistance` and `strength` state no span and stay plain.
- **The lay analysis had to learn reachability.** `property.usesCall` decided
  whether a document lays a material by scanning every function of every
  imported package for a call named `withLay`, reached or not. The `withLay(m)`
  preset is such a call inside the material library itself, so every document
  importing `material` would have paid for a blended lay row (seven WGSL
  goldens moved, all with redundant zero-lay blends). It now follows calls
  from the document's top level, through its own and the imported packages'
  functions and constants, so what no path reaches does not count
  (`internal/assembler/property/lay.go`, with `lay_test.go`).

### The defaults

The examples' values where an example calls the operator, the Ready-made
`@material` presets' colours for the coats, and a stated choice otherwise.

| preset | bakes | source |
| --- | --- | --- |
| `machined.turned(m, c, frame)` | feed 0.15, noseRadius 0.8 | Material - Coated bar |
| `machined.faced(m, c)` | origin, +Z, feed 0.15, noseRadius 0.8 | the knob |
| `machined.faceMilled(m, c)` | origin, feed +X, normal +Z, cutter 25, feedPerRev 0.3, corner 0.8 | the plates |
| `machined.ground(m, c)` | +X, Ra 0.8 | the aluminium examples (steel uses 0.3) |
| `machined.brushed(m, c)` | +X, grit 150 | the stainless example (No. 4) |
| `machined.blasted(m, c)` | mediaSize 100, angular 0 | bead blast; the grit example is 300, 1 |
| `machined.knurled(m, c, frame)` | diameter 16, pitch 1.0, diamond 1, raised 1 | the knob |
| `machined.honed(m, c, frame)` | angle 45, grit 220 | the usual bore |
| `machined.engineTurned(m, c)` | origin, feed +X, normal +Z, spot 12, stepover 0.45 | a choice |
| `machined.ballMilled(m, c)` | origin, feed +X, normal +Z, ballRadius 3, stepover 0.5 | a choice |
| `machined.scraped(m, c)` | flakeSize 12 | a choice |
| `machined.lapped(m, c)` | roughnessRa 0.025 | a choice |
| `machined.sparkEroded(m, c)` | vdiGrade 27 | the middle of the scale |
| `machined.tumbled(m, c)` | mediaSize 3 | a choice |
| `machined.polished(m, c)` | mirror 1 | a choice |
| `coating.anodized(m, c)` | dye [1, 1, 1] | clear anodize, the bar |
| `coating.hardAnodized(m, c)` | thickness 50 | the mill example |
| `coating.blackOxide(m)` | oiled 1 | the example |
| `coating.phosphate(m)` | manganese 1 | the example |
| `coating.heatTint(m, c)` | celsius 310 | `blueTemperedSteel`, the dark blue temper |
| `coating.anodizedTitanium(m, c)` | volts 35 | `blueAnodizedTitanium` |
| `coating.chippedPaint(m, c)` | paint [0.14, 0.32, 0.18], primer [0.55, 0.22, 0.14], gloss 60, wear 0.5, chipSize 1.5 | the example; wear is uniform here, an author hands it `edgeWear` |
| `coating.clearCoat(m, c)` | tint [1, 1, 1], gloss 45, build 0.08, underSize 1.2 | the coarse-cut plate |
| `coating.powderCoat(m)` | [0.62, 0.06, 0.05], gloss 85, peel 0.7, build 0.08, underSize 1.2 | `redPowderCoat`, the plate's build |
| `coating.wrinkleCoat(m)` | [0.04, 0.04, 0.045], size 0.6, build 0.08, underSize 1.2 | `blackWrinkleCoat` |
| `coating.sandCoat(m)` | [0.12, 0.22, 0.12], grain 0.2, build 0.08, underSize 1.2 | `greenSandCoat` |
| `coating.hammertone(m)` | [0.18, 0.32, 0.45], cellSize 4, build 0.08, underSize 1.2 | `blueHammertone` |
| `plated.zinc(m)` | bright 1 | `zincPlatedSteel` |
| `plated.zincYellow(m)` | bright 1 | `yellowZincPlatedSteel` |
| `plated.zincBlack(m)` | bright 1 | `blackZincPlatedSteel` |
| `plated.nickel(m)` | bright 1 | `nickelPlatedSteel` |
| `plated.chrome(m)` | decorative 1 | `chromePlatedSteel` |
| `plated.tin(m)` | bright 1 | `tinPlatedCopper` |
| `plated.cadmium(m)` | bright 1 | `cadmiumPlatedSteel` |
| `plated.goldFlash(m)` | microns 0.25 | `goldFlashedBrass` |
| `plated.silver(m)` | bright 1 | `silverPlatedBrass` |
| `sheet.hotRolled(m, c)` | flaking 0.2 | `hotRolledSteel` |
| `sheet.coldRolled(m, c)` | direction +X | `coldRolledSteel` |
| `sheet.pickledOiled(m, c)` | oiled 1 | `pickledOiledSteel` |
| `sheet.checkerPlate(m, c)` | origin, along +X, normal +Z, pitch 25, barLength 1.0, barHeight 1.2 | `checkerPlateSteel` |
| `sheet.stuccoEmbossed(m, c)` | pebble 4, depth 0.2 | `stuccoEmbossedAluminum` |
| `cast.sandCast(m, c)` | grainFineness 60 | `sandCastIron` |
| `cast.dieCast(m, c)` | direction +X | `dieCastAluminum` |
| `cast.investmentCast(m, c)` | roughnessRa 3 | `investmentCastStainless` |
| `cast.forged(m, c)` | scale 0.5 | `forgedSteel` |
| `cast.shotPeened(m, c)` | almenIntensity 8 | `shotPeenedSteel` |
| `molded.textured(m, c)` | vdiGrade 27 | `texturedPlastic` |
| `molded.leatherGrain(m, c)` | grainSize 1.5, depth 0.08 | `leatherGrainBlack` |
| `molded.grip(m, c, frame)` | diameter 30, pitch 3, height 0.8, diamond 0 | `rubberGrip`'s dots on a face |
| `printed.laserMelted(m, c)` | Ra 10 | the titanium examples |
| `printed.sintered(m, c)` | Ra 6 | the binder-jet alloy presets |
| `ceramic.celadon(m, c)` | depth 0.8 | `celadon` |
| `ceramic.crackle(m, c)` | glaze cream, spacing 10, stain 0.6 | `crackleGlaze` |
| `ceramic.speckled(m, c)` | glaze oatmeal, speckSize 1, share 0.04 | `speckledStoneware` |
| `ceramic.matteGlaze(m, c)` | glaze slate blue, opacity 0.95 | `matteGlaze` |
| `classic.scanlines(m, c)` | pitch 1, darkness 0.5, across +Z | `scanlines` |
| `weathering.weathered(m, c)` | rust `hsv2rgb([0.045 tau, 0.75, 0.55])`, amount 0.5 | `rustedSteel`'s colour; amount is a choice |
| `weathering.weatheredRunoff(m, c)` | the same, origin, flow -Z, streakLength 20 | a choice |
| `playful.glitter(m, c)` | flakeCellSize 1, flakes white, density 0.25 | the examples |
| `playful.glitterTwinkle(m, c)` | the same, sparkle 8 | Playful in Time |
| `material.withRoughness(m)` | 0.5 | a choice; the examples run 0.15 to 0.75 |
| `material.withMetallic(m)` | 1 | a choice |
| `material.withIor(m)` | 1.5 | glass, the examples |
| `material.withEmissive(m)` | white, strength 1 | a choice |
| `material.withTransmission(m)` | 0.6, 24 | the water examples |
| `material.withLay(m)` | +X, anisotropy 0.8 | the examples |

The four coats take `powderCoat`'s operator form, which has no substrate colour:
the coat's colour replaces the body's, so those presets have one slot.

### What the catalog carries

`catalog.json` gains `operators`, keyed `package.name`:

- `call`, the qualified preset; `arity`, from the return list; `slots`, the
  positions of the material, colour and frame parameters (`-1` for a slot the
  preset lacks).
- `adjust`, built by the same `adjustFor` as a `@material`'s, with two more
  placeholders: `${layer}` for the layer's binding, and the slot parameters
  written as `${material}`, `${color}`, `${frame}`. The return locals are
  renamed positionally, `${layer}`, `${layer}Color`, `${layer}Height`,
  `${layer}Slope`, which is the operators' contract and the names the examples
  use. A preset that bakes only plain arguments (`withIor`'s index) has no
  record: Wrap applies it and there is nothing for Adjust to open.

The tool proves every preset by a probe compile: applied to `metal.steel()` on
a sphere, with `[s.u, s.v]` for a frame slot, destructured to the recorded
arity. A preset whose arity or slots do not compile fails
`TestCommittedManifestIsCurrent`'s sibling rather than the panel.

### What the panel writes

**Wrap** writes one line and repoints two, above the return:

```gsl
boltSurface, boltSurfaceColor = metal.steel();
boltTurned, boltTurnedColor, boltTurnedHeight, boltTurnedSlope = machined.turned(boltSurface, boltSurfaceColor, [bolt.u, bolt.v]);
bolt.material = boltTurned;
bolt.color = boltTurnedColor;
```

The material slot is whatever `.material` read, as text: the panel's pair, a
finish, an author's expression, a kept wrapper; a body that set no material
wraps `material.standard`, the default it was drawn with. The colour slot is
`.color`'s text; a body without one hands the slot the colour it is drawn in,
as a literal, since the operator needs an expression. A layer is bound `<body><Op>`, fresh
against the document (`boltTurned2`), with `Color`/`Height`/`Slope` after it to
the arity. An arity-1 layer sets `.material` alone and leaves `.color`.

**The stack is read** from `.material`: read through what is written round the
surface by hand (a thread's relief, a lay), which the panel lists as rows of the
stack and keeps through every pick, find the destructure binding the name, and if its call is a catalog operator, by preset
or raw form, it is a layer whose first argument is the substrate; repeat. The
chain ends at a pair, a finish, or text the catalog does not know, which is the
base. The readout shows the base and one row per layer, innermost first.

**Controls on a layer.** Every range and color parameter of a layer has a
control indented under it in the Materials panel, without being asked for. A
pick writes those parameters out (`wrapOpen`, `openLayer`): the variables are
declared above the layer and the preset call becomes the raw call from the
template, the slot expressions carried over. It has to be the pick, because a
preset has no uniform: with the write left to the first drag, nothing moved
until the pointer let go. The cost is lines in the document, and compile time
measured the same. The variables are the panel's alone, so the Parameters panel
leaves them out, recognizing them by the panel's naming: `<body><Param>` on a
layer named `<body><Op>`. An author's slider passed into a layer stays in
Parameters.

**Adjust on a layer** (`adjustLayer`) writes out what a pick leaves in the call:
directions and points, which the viewer's gizmos drive, so a pick drops no gizmo
in the view. The chip and thumbnail rules from step 5 B apply to the layer that
sets `.color`.

**Remove** repoints whatever read the layer (the body's properties, or the
layer above's slots) at the layer's own slot expressions, and drops the layer's
line and its lifted variables when nothing reads them, by the pruning that
already exists.

**A layer's height** (added 2026-10-26). Every operator but the knurl returns
0 as its height, its relief riding in the material; the knurl's teeth are a
height the caller has to apply. The catalog marks such an operator `relief`
(its full form returns something other than the literal 0 third), and Wrap
applies the layer's height as displacement on the body's material, outside
every layer: `bolt.material = material.withDisplacement(boltKnurled,
boltKnurledHeight);`. A layer wrapped over it reads the layer itself, so the
relief stays outside; the stack and the readout read through the wrapping;
Remove takes it with the layer. The displacement replaces the layer's own
displacement row, as `material.withDisplacement` does.

### Order of work

1. Language: `@surface` (recorded, function declarations only, no argument) and
   `@frame` (parameter only); validators, parser testdata, `gsl_description`.
2. Libraries: the 28 presets and `material`'s six parameter annotations.
   Behaviour unchanged; the docs show the new overloads.
3. Catalog: `operators` with `adjust`, the probe compile, `-manifest-only`.
4. Panel: Wrap, the stack rows, controls and Remove per layer; unit tests.

## Parked: the lay at the axis a turned surface turns about

A defect found, fixed, measured and then reverted, because the fix cannot be
seen. Recorded so that the next person does not spend the same afternoon on it.

**The defect.** `machined.grooveFinish` receives the lay as the radius vector,
so at the axis that vector is zero while the anisotropy is not. On
`machined.faced` with a 0.15 mm feed:

| radius | lay | anisotropy |
| --- | --- | --- |
| 4 to 0.15 | `[0, 1, 0]` | 0.8301 |
| 0 | `[0, 0, 0]` | 0.8301 |

So the centre of a faced face asks for full streaking along a direction of no
length, and which way it points depends on the direction the centre is
approached from.

**The fix**, four lines, needing nothing `grooveFinish` does not already hold:

```gsl
grain = clamp(length(direction) / spacing, 0, 1);
```

applied to the anisotropy. A grain exists only where its features are farther
apart than they are wide, and within one feed of the centre the grooves are
rings shorter than their own spacing. It is exactly 1 wherever the length
reaches the spacing, so it changes nothing outside a cylinder of one feed about
the axis. `turned` is inert under it (its length is the part's radius) and so is
`faceMilled`, whose direction has length exactly the cutter radius.

**Why it was reverted.** It cannot be observed, and not because it is subtle.
`layFinish` switches the lay OFF when the grooves are drawn as displacement —
`across` collapses to the lobe's own 0.02, `stretch` goes to zero, and the
anisotropy with it — so the lay exists only in the opposite regime, where the
grooves are too fine to resolve and are folded into the lobe instead. The fade
spans `r < spacing`, one groove spacing. **Anisotropy therefore only exists when
one groove spacing is below a pixel, which is exactly when the whole fade is
below a pixel.** The two conditions exclude each other by construction.

Measured both ways: a close render with the grooves drawn is bit-identical
before and after, and forcing the anisotropy from 0 to 1 on that same scene is
also bit-identical, while the same forcing on a plain plate does change the
picture. Zoomed out, where the lay is live, before and after differ in bytes
and are indistinguishable by eye.

It cost 52 lines of golden movement — no numeric literal changed in either
file — plus two emitted functions and a `length()` per sample in every turned
or faced shader. If the detail thresholds in `texture.fadeDetail` ever move, or
a renderer draws grooves differently, this becomes visible and the fix above is
the whole of it.
