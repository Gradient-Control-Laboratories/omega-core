# Material library plan: more diverse, more mechanical

A plan for growing the material catalog from 225 entries, 28 of them
operators, into a library that covers what a part is made of and how it was
made, across more processes than it does today. Written 2026-10-27 against the
libraries in `stdlib/material` and the catalog
`tool/materialswatch` renders from them, and revised the same day with the
plans for relief, stand-ins, datums, animation and detail that every new entry
follows. The architecture it builds on is in `material_composition.md`;
nothing here changes that convention.

## What the market list says, and what it does not

The plan was drawn against a competing product's library of about seven
hundred entries. It is inspiration for coverage only: no name, swatch or
editorial set from it is reproduced here, and the product is not referenced by
name anywhere in the repo.

Read for distinct surfaces rather than entries, that list reduces to roughly a
hundred and twenty. Forty-two of its entries are six paint finishes in seven
colours; thirty are clays; fifteen are two-tone swirls. Colour is a parameter
here (`@color`, Adjust), so their forty-two paints are one preset of ours. The
count to match is the surfaces, not the entries.

What that list covers and `omega-core` does not, structurally:

- **Woven and knit cloth**, and leather: nothing in `stdlib/material` is a textile.
- **Tessellations with joints**: brick bonds, tiles with grout, parquet.
- **Stylised relief**: scales, studs, bubble wrap, wrinkles, the sculpting-app
  look, which GSL does better than a texture map can because its relief is
  real displacement on the field.
- **Plated and textured plastics and metals** as finished surfaces rather than
  measured substances.

What `omega-core` has that it cannot: composition. A base under operators under
weathering, each a layer the Materials panel can add and remove, with the relief
applied to the body. So the plan is families of **processes**, each a package
of constructors with a ready-made layer on top for the browser, not a flat
list of adjectives.

## Principles

1. **Name the process or the specification, never the adjective.** `plated.zincYellow`,
   `machined.honed`, `sheet.checkerPlate`. Where a standard grades the surface,
   the grade is the parameter: VDI 3400 for spark-eroded texture, Almen
   intensity for peening, Ra class for finishes, flute class for corrugated
   board.
2. **Substrate × operator before new entries.** A new surface that is some
   base under some process is an operator, not a preset, and gets its presets
   only for discoverability in the grid.
3. **Sized for the swatch.** Every ready-made preset is sized for a 100 mm
   part, which is what the catalog's sphere is. A preset that carries a hidden
   part size says so in its doc line.
4. **Every length is a millimetre.** See "Scale" below. A pattern has one
   characteristic length, its size, a `@range` in millimetres until units are
   built into the system, and is similar under it: the other lengths inside
   it are ratios of the size.
5. **Mechanical first.** Tier 1 is what a customer's part is made of and
   finished with. Tier 2 is coverage. Tier 3 is play.
6. **Every entry declares its scales.** A pattern at its size, relief at the
   millimetre, detail below what the thumbnail shows: an entry says which of
   the three it has, and which it lacks and why. See "Detail".
7. **One name per surface; a tier is a trailing parameter.** Relief and
   motion cost the viewer, so they are chosen by an overload's last argument,
   not by a second name. See "Animation".
8. **Bulk, finish or coat.** Every entry is one of the three, decided by what
   it takes and what it does to the colour beneath, and a document reads as
   that stack. See "Taxonomy".

## Taxonomy: bulk, finish and coat

Today the catalog knows two kinds and the packages know none. The catalog's
`kind` is `finish` for the eight lobe-only presets of `material.gsl`
(`standard` to `chrome`, applied keeping the body's colour) and `material` for
the other 217, which carry a colour and a material; the 28 `@surface`
operators are a third list keyed by the slots they take. The browser lays the
finishes out first and then each package under its own section headings, so
`coating · Ready-made` sits between `animated` and `machined`, and a search
for "steel" finds the metal, its rust and its blast. Inside the packages the
roles mix:

- `coating` holds two families under the same names: full surfaces,
  `powderCoat(coatColor, glossUnits, peel)` returning material, colour, height
  and slope, for the Ready-made presets; and operators over a substrate,
  `powderCoat(baseMaterial, coatColor, ..., build, underSize)`, levelled by
  the film's build.
- `machined`'s operators take the part's own material and pass its colour
  through unchanged; the blast, the brush, the grind and the knurl are all
  the same substance, worked.
- `stone`, `wood`, `natural`, `playful` and `physical`'s metals are
  substances: a colour field and a material, nothing beneath.
- `printed` holds `powderBed`, `fdm` and `resin`, which are substances with
  their process built in, beside `laserMelted`, an operator over a metal.
- `weathering` takes both the material and the colour and alters both.

**The split.** Three roles, decided by the contract an entry already has
rather than by the package it lives in:

| role | takes | does to the colour | examples |
| --- | --- | --- | --- |
| **bulk** | nothing: it is the substance | provides it | the metals, stone, wood, natural, playful, classic, the printed substances, the lobe-only finishes applied to nothing |
| **finish** | the body's material and colour | passes it through | `machined.*`, `printed.laserMelted`, `material.with*`, the sheet mill finishes, casting and molding textures, tumbling, honing, peening |
| **coat** | the body's material, and its colour only to tint it | replaces it (powder, wrinkle, sand, hammertone, black oxide, phosphate, plating, paint) or tints it (anodize dye, clear coat, chipped paint) | `coating.*`, `plated.*`, enamel, primer, e-coat |

Weathering is a coat in the stack's sense, an alteration of whatever is
outermost, and stays its own package because its parameters are age and
exposure rather than a layer's colour and build. The catalog can read the
role off the slots the tool already records: no material slot is bulk; a
colour slot passed through is a finish, which `keepsColor` in the browser
already encodes for the eight lobe presets; a colour slot absent, or taken
only to tint, is a coat.

**Where the split lives: all three places, each doing its own job.**

- **The catalog's `kind`** becomes `bulk`, `finish` or `coat` (today's
  `finish` and `material`, plus the coats pulled out of `material`), derived
  by `tool/materialswatch` from the slots, so the libraries never state a
  role twice. The `section` stays the library author's.
- **The browser** groups by role first, three headings or three tabs, and by
  package and section within each, so the grid reads "what is it made of,
  how was it finished, what is on it". `keepsColor` follows the role: a finish
  keeps the chip, a coat replaces it, a tint shows the tint.
- **The packages** stay named by process, which the naming rules require,
  and may hold more than one role: `printed` is a bulk nylon and a laser
  finish because that is the process. What moves is the overlap: the
  `coating` package's full-surface overloads fold into the operator form
  applied over `material.standard`, and their Ready-made presets become coat
  presets over the default substrate, which is what Wrap builds when a user
  picks one today. `plated` arrives as coats only.

**What a document reads as.** A body's surface is a stack, bulk under finish
under coat under weathering, which is exactly the layer lines Wrap writes:
`boltTurned = machined.turned(bolt.material, ...)`, then
`boltPowderCoat = coating.powderCoat(boltTurned, ...)`. The order is physical:
a coat's build levels the relief finer than it (`filmKeeps`), so a coat over
a turned finish keeps the grooves the film cannot fill and loses the lay,
while a finish over a coat is a different thing, a painted part machined, and
the panel can say so instead of building it. Weathering goes outermost, on
the coat if there is one, and `edgeWear` reads the coat's chips.

**What it buys.** The grid stops mixing a substance with what is done to it,
so a user looking for steel is not offered its rust; the counts against the
market list are honest, since their forty-two paints are six coats; the
stack's order is checked rather than trusted; and a new entry is placed by
its contract, which a reviewer can read off its signature, rather than by a
judgement about the package. The decision for Blake is whether weathering is
a fourth role in the grid or the last coat, and whether the `coating`
overload fold is worth the golden movement it will cost in the examples that
call the full-surface forms.

## The families

| tier | package | what goes in | needs |
| --- | --- | --- | --- |
| 1 | `printed` | binder jet as the customer builds it: sintered as-built, tumbled, polished, each over the alloys their catalogue names; frosted SLA | nothing new |
| 1 | `plated` (new) | operators over any base: zinc clear, yellow and black chromate, bright and electroless nickel, hard and decorative chrome, tin, cadmium, gold flash, silver | thin film |
| 1 | `machined` | honed crosshatch, engine turned, ball-end scallops, scraped, lapped, spark-eroded by VDI 3400 grade, tumbled | nothing new |
| 1 | `sheet` (new) | hot rolled with mill scale, cold rolled, pickled and oiled, checker plate per ASTM A786, stucco embossed | frame form, relief |
| 1 | `cast` (new) | sand cast by AFS grain fineness, die cast with flow lines, investment, forged scale, shot peened by Almen intensity | nothing new |
| 1 | `molded` (new) | spark-eroded mold grain by VDI grade, leather grain, glass-filled nylon flecks, overmolded grip, silicone, acetal, PTFE, PEEK, UHMW | shares the EDM grain |
| 1 | `composite` (new) | carbon twill and plain weave under clear coat, forged carbon, glass laminate, aramid, phenolic laminate | weave |
| 1 | `coating` | dyed anodize colours, ceramic spray coat, PTFE coat, heat tint by temperature, bluing, e-coat, primer, enamel | thin film |
| 2 | `textile` (new) | canvas, denim, ripstop, ballistic nylon, mesh, jersey, felt, velvet; leather full grain, pebbled, suede, patent; webbing | weave, lay per tow, a sheen lobe |
| 2 | `masonry` (new) | brick bonds with mortar, block, subway, hex and penny tile with grout, herringbone, terrazzo, board-formed and exposed-aggregate and polished concrete, plaster, asphalt, slate, travertine, basalt, obsidian, engineered quartz | tessellation, frame form |
| 2 | `paper`, `foam` (new) | corrugated by flute class, kraft, chipboard, foam core; open cell, closed cell and memory foam, aerogel | nothing new |
| 2 | `ceramic` (new), `glass` | celadon, crackle, speckled and matte glazes; alumina, zirconia, silicon nitride; reeded, hammered, wire and dichroic glass | thin film for dichroic |
| 3 | `organic` (new) | orange peel, bark, cork, bamboo, wicker, coral, eggshell speckle | weave |
| 3 | `playful` | stripes and candy stripes along a direction, barber pole, hazard chevrons, camouflage and animal prints in fixed palettes, tie-dye, halftone, hex grid, sprinkles, studs, bubble wrap, scales, wrinkle, swirl, slime and gummy, neon, holographic film, ombre along a direction | relief, thin film, a frame |
| 3 | `classic` (new) | the textbook procedural textures, each after its paper: Perlin marble and turbulence, Peachey solid wood, Worley cellular and Voronoi mosaic, plasma, metaball lava lamp, Phong plastics, the Cornell-box palette, Truchet tiles, scanlines, pixel art, chrome over a checker floor | nothing new; time |
| 3 | `weathering` | operators: dust on up-facing surfaces, scuffs, fingerprints on polished metal, tarnish, tide lines, efflorescence, mildew, sun fade, chalking, crazing, edge wear | a normal slot; edge distance |

About a hundred and seventy entries over the three tiers, which with the 225
today is a catalog of about 400 against the market list's 700, and more
distinct surfaces than it has.

### Tier 1 in detail

**printed.** The customer's process is binder jetting, and no alloy of theirs
is named in any library today. `sintered(base, color, roughnessRa)` is the
as-built surface, a matte grain finer than `laserMelted`'s with a little
surface porosity, and two post-processes, `tumbled` and `polished`, go in
`machined` since they apply to any metal. The presets are one per alloy the
customer's catalogue lists, over `physical`'s measured metals; verify that
list against the catalogue before naming any. Frosted SLA is `resin` with
clarity and a blasted surface.

**plated.** Every entry is an operator over a base: `zinc(base, color,
chromate)` with the chromate clear, yellow or black, `nickel(base, color,
bright)`, `chrome(base, color, decorative)`, `tin`, `cadmium`, `goldFlash`,
`silver`. Yellow chromate and cadmium are iridescent thin films over the
plate; silver and brass tarnish, which is `weathering`'s job. Plating is thin,
so it keeps the base's relief and replaces its material and colour.

**machined.** `honed(base, color, frame, angle, grit)` is the cylinder-bore
crosshatch, two lays at ±angle. `engineTurned(base, color, frame, spot,
pitch)` is overlapping circles of brushed lay. `ballMilled(base, color,
origin, feed, normal, ballRadius, stepover)` is the scallop field of a 3D
pass. `scraped` is the flaked hand-scraped way. `lapped` is a mirror with no
lay. `sparkEroded(base, color, vdiGrade)` is the EDM grain, Ra = 10^(grade/20)
µm from VDI 3400 reference 12 to 45, shared with `molded`. `tumbled(base,
color, mediaSize)` rounds the micro-relief and dulls the lobe.

**sheet.** Mill finishes are flat materials over steel or aluminium: `hotRolled`
with blue-grey scale and flaking, `coldRolled` with a faint lay along the
rolling direction, `pickledOiled` dull and oiled. `checkerPlate(base, color,
frame, pitch, barLength, barHeight)` is the raised-bar diamond pattern per
ASTM A786, returning the bars as a height the panel applies. `stuccoEmbossed`
is the aluminium pebble.

**cast.** `sandCast(base, color, grainFineness)` is Ra 12 to 25 µm with the
AFS grain fineness number as the parameter. `dieCast` is fine with flow lines
along a direction. `investmentCast` is smooth matte. `forged` is scale over a
die finish. `shotPeened(base, color, almenIntensity)` is a dimple field whose
size and density follow the shot.

**molded.** `textured(base, color, vdiGrade)` is the spark-eroded mold
grain. `leatherGrain(base, color, grainSize)` is the generic leather-pattern
mold texture. `glassFilled(color, fibreShare)` shows short fibres as flecks
under a matte surface. `grip` is overmolded elastomer with a dot or diamond
relief. `silicone` is translucent and matte, and must stay smooth: see the
transmissive rule below. `acetal`, `ptfe`, `peek`, `uhmw`, `abs` are the
engineering polymers as molded surfaces over `physical`'s measured plastics.

**composite.** `carbonTwill(frame, towWidth, gloss)` and `carbonPlain` are a
weave of anisotropic tows, the lay alternating per tow, under a clear coat.
`forgedCarbon(chipSize)` is chopped tow by cellular regions. `glassLaminate`
is the yellow-green translucent weave of G10 and FR4. `aramid` is the yellow
weave. `phenolicLaminate` is canvas, linen or paper base in brown or black.

**coating.** `anodized` already takes a dye; the presets are the colours a
catalogue sells: red, blue, gold, black. `ceramicCoat` is a thin matte spray
with a fine texture. `ptfeCoat` is dark grey at a low sheen. `heatTint(base,
color, celsius)` is the temper colours of steel, pale straw at 200 °C through
brown, purple and blue to grey at 330 °C, as a thin oxide film whose
thickness follows the temperature. `blued` is gun blue. `eCoat` is black
satin, `primer` grey or red oxide matte, `enamel` porcelain gloss.

### Tiers 2 and 3 in brief

**textile** needs the weave primitive and a lay per tow; velvet and suede
want a sheen lobe the renderer does not have, so they ship as matte
approximations or wait. Leather is cellular pores under a finish, patent is
leather under `clearCoat`.

**masonry** needs a regular tessellation returning the cell and the distance
to the joint, on the body's frame as `turned` and `knurled` take it. Brick
bonds, tiles in every pattern and parquet are that primitive with a size, a
joint width and two colours. Terrazzo is cellular chips in a palette.

**paper and foam**, **ceramic and glass** need nothing new except thin film
for dichroic glass. Reeded and hammered glass are displaced and transmissive,
so they must stay smooth: see the rule below.

**organic** is the textures `physical` has only as flat measured colours:
orange peel is cellular dimples under a waxy sheen, cork is lenticels, bamboo
is nodes over fine grain, wicker is the weave with round tows.

**playful** is relief, palettes and motion. Camouflage and animal prints take
fixed palettes rather than a palette function, because a constructor that
takes a function cannot be a preset (`material_composition.md`, "Second
pass"). Swirl is domain warp between two colours, and covers all fifteen of the
market list's swirls; tie-dye is the same warp in a radial frame with a
palette. Stripes, candy stripes and ombre run along a `@direction`; a barber
pole is stripes that climb with time. Halftone is `polkaDot` with the dot ratio
driven by a field, on the body's frame. Sprinkles are cellular rods in a
palette on a matte ground. Studs, bubble wrap and scales return a height, as
the knurl does; wrinkle is `coating.wrinkleCoat`'s relief at a larger size.
Slime and gummy are smooth transmissive bodies with a drip along a direction,
so they keep the transmissive rule below. Neon is `withEmissive` in a tube
colour with a slow flicker. Holographic film waits for thin film, and is
rainbow-in-space like `rainbowSheen`: a colour field sees neither the light
nor the eye, so true iridescence stays out of reach.

**classic** is the procedural textures everyone learned from, named after the
work that introduced each and cited in the package header: Perlin's marble and
turbulence (Perlin, "An Image Synthesizer", SIGGRAPH 1985,
https://doi.org/10.1145/325165.325247), Peachey's solid wood rings (Peachey,
"Solid Texturing of Complex Surfaces", SIGGRAPH 1985,
https://doi.org/10.1145/325165.325246), Worley's cellular basis and the
Voronoi mosaic (Worley, "A Cellular Texture Basis Function", SIGGRAPH 1996,
https://doi.org/10.1145/237170.237267), Phong's plastics (Phong, "Illumination
for Computer Generated Pictures", CACM 1975,
https://doi.org/10.1145/360825.360839), the Cornell box's red, green and white
(Goral et al., "Modeling the Interaction of Light Between Diffuse Surfaces",
SIGGRAPH 1984, https://doi.org/10.1145/800031.808601), a metaball lava lamp
(Blinn, "A Generalization of Algebraic Surface Drawing", TOG 1982,
https://doi.org/10.1145/357306.357310), plasma as fractal noise through a
cyclic palette (after Fournier, Fussell and Carpenter, "Computer Rendering of
Stochastic Models", CACM 1982, https://doi.org/10.1145/358523.358553), and
Truchet tiles (Smith and Boucher, "The Tiling Patterns of Sebastien Truchet",
Leonardo 1987, https://doi.org/10.2307/1578535). Scanlines and pixel art are a
pitch in mm along a frame. The chrome ball over a checker floor is a
document, not a material: `metal.chrome` over `playful.checkerboard`. Each
entry is the construction from the paper, not a look copied from any product.
Verify every DOI above against the publisher before copying it into a library
header.

**weathering** operators are `omega-core`'s structural advantage and the deepest
item. Dust, sun fade and mildew settle by how the surface faces, which needs
the body's normal as a slot the panel can pass, as it passes the frame. Edge
wear needs the distance to the nearest edge, which the second-closest
distance property gives for free once its rename lands. Tarnish, tide lines,
chalking and crazing need nothing new.

## Primitives the plan rests on

Four additions carry most of the families; do them first.

1. **One size per pattern**, landed 2026-10-28 as a `@range(min, max)` in
   millimetres on every pattern constructor's characteristic length, which
   Adjust lifts like any range. A separate length kind with a log slider
   labelled in mm was built and withdrawn the same day: units will be built
   into the system, and a control of its own for one unit would be a second
   path to the same place. Before it, fifteen presets baked a size literal the
   panel could not reach (terrain's scale, lava's crackScale, water's
   rippleScale, cloud's puffScale, checker's cellSize, polka's spacing,
   granite's cellSize, marble's scale, flagstone's cellSize, glitter's
   flakeCellSize) and nine lengths were already `@range` (hammertone's
   cellSize, wrinkle's size, sand coat's grain, galvanized's spangleSize, the
   knurl's pitch, the print layer heights, blasting's media and anodizing's
   film in µm). The rule now: a pattern's characteristic length is one
   `@range` size in mm; the lengths inside it are ratios of that size, so the
   pattern is similar under its slider (marble's `veinSpacing`, `veinWidth`
   and `warp`, polka's `dotRatio`, cellSkin's `zBandRatio`); a process length
   that is a setting of a real process keeps its own unit (layer height,
   blast media, anodize film, powder grain, knurl pitch, feeds); a relief
   height is its own `@range` in mm (see "Relief"); a scale model takes the
   length on the part and derives its ratio (`ocean` and `sandRipples` take a
   wavelength in mm). Wood's lengths are a species' real sizes from its table
   and wait for a later pass.
2. **A weave** in `texture`, landed 2026-10-30 as `weave(x, y, pitchU,
   pitchV, towWidth, over, under)` on plane coordinates, as the dot plane
   is, returning a unitless height, whether the warp is on top, and the top
   tow's index; the caller lays it on a frame with a whole count of tows
   round a diameter, or on a face's datums, and knows its own axes, so it
   has the tow direction. Cloth, composites, wicker and mesh are all this
   one function.
3. **A regular tessellation** in `texture`, landed 2026-10-30 as `tiles(x,
   y, sizeU, sizeV, joint, bond)` on plane coordinates, as the weave is,
   returning the cell as a hash seed and the signed distance to the joint's
   edge, with `hexCells(x, y, across)` beside it on map's hexagonal lattice
   for hex and penny tile. Bricks, tiles, parquet and checker plate.
4. **A thin-film colour** from thickness and index, landed 2026-10-28 in
   `color`: `thinFilm(thickness, filmIndex, substrateIndex)` is a single
   layer's reflectance spectrum, straight on, integrated against the CIE
   matching functions (Wyman, Sloan and Shirley's fit) and written as sRGB;
   `thinFilmOver(thickness, filmIndex, baseColor)` tints a metal's colour by
   the film, the metal's reflection standing in as a real index with the
   colour's luminance as its reflectance. Thirty-six wavelengths unroll at
   compile time and the matching functions fold to constants, so a constant
   thickness is a constant colour and a varying one costs a cosine per
   wavelength. Its first users are `coating.heatTint` (the temper colours, a
   thickness table 25 C apart set to the model's own colours) and
   `coating.anodizedTitanium` (2 nm per volt). Still to come on it: yellow
   chromate and cadmium in `plated`, dichroic glass, holographic film. What
   it cannot do: the colour shift with angle, which waits for a film term in
   the renderer.

Two more are deeper and gate only their families:

- **A normal slot** for operators, the way `@frame` passes the body's frame,
  so a weathering operator can read which way the surface faces.
- **A sheen lobe** in the renderer, for velvet, felt and suede.

**Texture primitives, after Blender's texture nodes** (landed 2026-11-03).
Blake set the bar as near-photorealistic and asked what building blocks
were missing against Blender. Six landed in `noise` and `texture`:

- `noise.fbm`, `turbulence`, `ridged` and `hybrid`: fractal noise with
  Blender's detail (1 to 6 octaves, fractional), roughness and lacunarity,
  and Musgrave's ridged and hybrid multifractals for crests that gather
  on crests and valleys that stay smooth.
- `noise.warp(size, amount)`: Blender's distortion, a warped lookup point
  for `remap`, which takes the lattice's grid out of any pattern.
- `noise.voronoi(size, randomness, metric, smoothness)`: Euclidean,
  Manhattan or Chebyshev, jittered by randomness, with a smooth F1; and
  `voronoiEdge(size, randomness)`, the true distance to the cell wall, so a
  thresholded crack or grout line keeps one width (F2 − F1 swells).
- `noise.anisotropic(size, direction, stretch, detail)`: noise stretched
  along a direction, for brushed metal, fibers and wood figure.
- `texture.bands` and `texture.rings`: Blender's Wave, sine, saw or
  triangle, distorted by noise four bands across so lines drift as growth
  rings and strata do.
- `texture.triplanarWeights(normal, sharpness)`: the three projection
  weights that wrap a plane pattern on any part.

What Blender does not do, and a real-time raymarcher must: every fractal
here fades each octave as it shrinks under a pixel, settling to the
octave's mean, because one sample a pixel turns sub-pixel octaves into
sparkle. Off a screen every octave is drawn, so meshing and export see the
full field. The ranges each doc states are measured over 4000 points with
the gsl CLI and pinned in `internal/gsleval/texture_primitives_test.go`,
which also proves the Voronoi scan sits on the cellular builtins' lattice
(Euclidean F1 equals `cellular` to 1e-3). The color ramps Blender's ColorRamp
node gives were already here, as `color`'s colormaps. Next: a realism pass
that rebuilds about ten flagship materials on these, judged against
reference photographs.

## Rules learned, to carry into every new entry

Each is worked out in its own section below.

- **A displaced transmissive surface must be smooth** (Relief).
- **Relief stays under about 3 mm, and its height is a parameter of its own,
  in mm** (Relief).
- **Relief rides in the material unless the operator returns it** (Relief).
- **Every point-dependent texture declares a stand-in, measured or by
  share, that follows its parameters** (Stand-ins).
- **A plane pattern reads the body's frame; a stock or process direction is
  a gizmo** (Datums).
- **One name; relief and motion are a trailing parameter** (Animation).
- **Each entry declares which of pattern, relief and detail it has, and why
  it lacks the rest** (Detail).
- **Each entry is bulk, finish or coat by its contract, and a document is
  that stack** (Taxonomy).

## Relief

Relief is what GSL does that a texture map cannot, and the one place the
renderer sets a hard limit, so every entry decides its relief deliberately.
There are three channels, and an entry may use all three.

**Relief in the material.** `material.withDisplacement(m, height)` carries a
height in mm that the viewer marches as surface. The viewer evaluates the
displaced field only within a band of the undisplaced surface, 4 mm since
2026-10-27 (`gsl_dispNearMargin`, mirrored in `gsl-glass.js` and the CPU lit
renderer), and the band's edge is a jump in the marched field, so a height
has to stay well under it: measured on `natural.terrain`, relief renders clean
to 3.4 mm, slivers at 3.7 and holes from 4.1. Design to 3 mm. This is why a
relief height cannot be a fixed ratio of the pattern's size: a terrain at scale
40 would stand 8 mm tall at its natural proportion. A displaced preset
therefore exposes its height as its own `@range` in mm, read directly against
the band, as `raisedTerrain(scale, peakHeight, snowBias)` does with the field
normalised by its measured summit; lava's `crustDepth`, water's
`rippleAmount` and the cloud's `puffAmount` are the same kind. Once any body in
a program carries a height, every ray in it pays the displaced band; a zero
height does not buy that back, which is one half of the two-names question in
"Animation".

**Relief the operator returns.** A relief that is the part's own shape, teeth,
bars, studs, scales, a weave, is returned as the operator's third value and
marked `relief` in the catalog; the panel applies it to the body, and a
document may subtract it from the geometry instead, as Terrain Planet does
with the terrain's height, when the mesher must see it. Only `knurled` does
this today; checker plate, studs, bubble wrap, scales and the weave follow it.
An operator whose relief rides in its material returns 0 as its height.

**Fine relief.** Relief finer than a pixel is not marched at all: it is the
roughness its hidden slopes make, `texture.reliefRoughness`, and above the
pixel it is drawn as `.detail` with a `.detailSize`, evaluated only where it
shows and faded by `texture.fadeDetail` (the powder coat's two orange-peel
bands, the wrinkle and sand coats, the print layers and powder grains,
granite's seams, wood's fibre and pores). The convention from `printed`'s
`layerFinish` is the model: the coarse look, which is also the stand-in,
carries the slopes in its lobe; the fine detail adds the relief where it is
drawn and narrows the lobe back as it does. New fine relief follows it rather
than adding a second `withDisplacement`.

**A displaced transmissive surface must be smooth.** A rough transmissive
material takes the viewer's dithered glass branch, four refracted rays per
pixel through up to four bounces of 128 steps, and every step evaluates the
displacement and four gradient taps. The cloud preset was transmissive, rough,
displaced and animated at once and hung the GPU on a macOS watchdog; it is
opaque now. Rippled water survives because it is smooth, so it takes the
single-ray branch. Reeded glass, hammered glass, silicone, slime and gummy
follow the same rule: roughness at or under 0.02, or no displacement.

Which families carry which:

| channel | carries it today | in the plan |
| --- | --- | --- |
| in the material, with a height parameter in mm | terrain (raisedTerrain), lava with crustDepth, water, cloud, ocean, sand, rain, mud, ice, lava lake, galvanized, hammertone | leather grain, stucco, shot peen, forged scale, orange peel, bark, cork, wrinkle, swirl relief |
| returned height | knurl | checker plate, studs, bubble wrap, scales, weave, tiles' joints, grip |
| fine relief | coats, prints, granite, wood | sintered grain, spark-eroded grain, sand grain, mill scale, every textile's tow |
| none, by rule | fog, glass, metals, flat colours | plating, heat tint, neon, pixel art, the palette textures |

## Stand-ins

A value's stand-in (`.simple`) is what it draws as below the Detail slider's
Textures rung, and the Bodies chip's colour comes from the same idea: the
flat look a texture settles to. Every point-dependent texture declares one;
a value's stand-in is itself until declared and is not inherited through
arithmetic or function results, so a texture built from another's output
declares its own.

**Two ways to get one, and when each applies.**

- **By share**, when the texture is a two-colour paint whose coverage has a
  closed form: the checker's even mix, polka's pi r^2 over the honeycomb,
  flagstone's 28.9% joints, glitter's flake share, the rust's `corrodedShare`.
  The share is a function of the parameters, so the stand-in follows its
  slider. Prefer this form wherever the geometry gives it.
- **Measured**, when the colour is a nonlinear mix: terrain's Oklab mean over
  the painted globe, the marbles' means, the decayed woods' coverages. Measure
  with the gsl CLI (`gsl eval -doc f.gsl -points <file> expr`) over 4096
  points on the catalog's 50 mm sphere, in Oklab for colours, record the
  mean, the point count and the arguments in a comment beside the declaration,
  and make the result a function of the parameters where the dependence is
  simple, as `terrainSnowShare(snowBias)` fits the snow's share to a measured
  curve. A constant measured at the preset's arguments goes stale the moment
  the user slides a parameter; say so in the comment when that is the case.

**Materials** stand in by the lobe: the roughness that the relief's hidden
slopes make, `sqrt(sqrt(alpha^4 + slope^2))` as the coats and water write it,
with the lay kept and no direction to streak when the features are below a
pixel. A displaced material's stand-in is displaced by its mean height, so the
surface stays where it was.

**Known defect.** `wood.woodStandIn` leaves the sapwood out: a maple round's
measured mean colour is [0.825, 0.740, 0.628] against its stand-in's [0.697,
0.555, 0.392], and every decayed stand-in inherits the bias. Fix it by mixing
the sapwood colour by its share of the round, which `sapwoodDepth` already
knows.

**What the chip needs.** The catalog's `colors` are read off the colour field
unlit over the sphere, one colour for a uniform texture and a spread for a
mottled one (`colors.go`), because the Bodies panel cannot evaluate a field. An
adjusted body's chip reads its `@color` declaration instead (`coat.go`), so a
preset with one colour parameter keeps a live chip and one with none or
several shows the manifest's. A wrapped body still reads document-coloured
even when the operator passes the substrate colour through; that is an open
item in `material_composition.md`. So a new entry's checklist is: a stand-in
for every returned colour and material, by share if the geometry allows,
measured and dated otherwise; one `@color` where the entry has a dominant
colour, so the chip follows it.

## Datums: position, direction and frame

A pattern sits somewhere and faces some way, and the library has two sources
for that: the body's own frame, and a datum the author places.

**The body's frame first.** A shape carries `.u` and `.v`, the coordinates a
material reads to know where it sits (`material_composition.md`, "`.u` and
`.v`"), and `turned` and `knurled` take them as one `@frame` argument, so one
call turns a capped cylinder's side and faces its caps. Every pattern that
lives on a plane wants this form: polka dots, checker, halftone, stripes,
tiles, weave, checker plate, scanlines, pixel art. Dots on a cylinder wrap
round it instead of being projected through it from a fixed XY plane, which is
what `polkaDot` does today. The frame costs the author nothing, and Adjust
passes it from the body as it does for the knurl. Two cautions the frame
conventions already record: `u` is an angle whose value has a seam, so a
pattern that counts along it round a cylinder has to divide the circumference
whole, as the knurl snaps its tooth count; and the frame reaches a part named
in the document, not one built inside a library function.

**A direction where the pattern has one axis.** Stripes, ombre, a barber pole,
rainbow bands, a drip, a lay, a flow, a build direction, a grain axis, a wind.
These are stock or process choices extrinsic to the part, and they are the
datums that want a gizmo: `@point origin, @direction(origin) along` as
`weatheredRunoff` and the wood take them. A single direction is enough when
the pattern is invariant round that axis; `rainbowSheen` and `rainbowCycle`
run along z today and should take one.

**A frame by gizmo where there is no body frame to read.** A plane pattern on
an imported mesh, or inside a library function where the frame does not reach,
takes `@point origin, @direction(origin) normal, @direction(origin) along`,
the three `faceMilled` already takes, and builds its frame from them. This is
the fallback form, not the first; the examples keep it for the cases above.

**None, by nature.** Granite, glitter, terrain, lava, flagstone, the cellular
skins, confetti, foam, every isotropic noise texture. Adding a datum to these
is wrong, not merely unnecessary.

So the answer to "polka dots could use a direction parameter" is: a frame form
first, `polkaDot(spacing, dotRatio, colorA, colorB, frame)`, with the
`(origin, normal, along)` form as the fallback, and the same for checker and
the tiles to come. What landed in the size pass is the smaller step,
`polkaDot(spacing, dotRatio, @direction direction, colorA, colorB)`, the dot
plane's normal through the origin, the identity for +Z; the frame form is the
next overload. Adjust's rule for datums stands: the defaults are the preset's
own arguments, and the frame is read off the body.

## Animation

A material that reads `time` is re-rendered every frame while the clock runs;
one that does not is drawn once and held. The catalog renders every entry at
time zero, so a moving preset's thumbnail is its first frame. Today the
movers are the whole of `animated` (ocean, river, rain, mud pot, lava lake,
ice floes, sand ripples), `natural`'s water, lava and cloud, and `playful`'s
drifting dots, cycling rainbow and twinkling glitter.

**Why there are two of some.** `lava` and `lavaHero` are one look at two
costs: the hero adds a displaced crust, and once a body carries a height the
whole program marches the displaced band. `polkaDot` and `polkaDotDrift` are
a still pattern and the same pattern reading `time`. In neither case does a
zero argument remove the cost: a zero relief still compiles the displaced
march in, and a zero velocity still leaves `time` in the program unless the
compiler folds `p - 0 * time` away, which it does not promise. The second
name is honest about the cost, but it doubles the catalog and reads as two
materials.

**The convention.** One name, with the tier as a trailing parameter of an
arity overload, which the language already uses for datums (`turned`'s axis
and frame forms) and for wood's species: `lava(crackScale, ..., flowTimeScale)`
and `lava(crackScale, ..., flowTimeScale, crustDepth)`; `polkaDot(spacing,
dotRatio, direction, a, b)` and `polkaDot(..., velocity)`. The overload shares
one doc comment, the trailing parameter names the cost it adds, and the cost
is paid only by the call that asks for it. One preset per look, the fuller
tier: `glowingLava` and `polkaDotDrift`, since a preset is a name in the grid
and two zero-argument overloads of one name cannot exist. The plain tier is
reached by writing the shorter call, and Adjust lifts the trailing parameter
(the drift's velocity as a `@point`, the crust's depth as a range) so a
document can turn it down. Both existing pairs are folded; the convention
applies to new entries.

**What moves, and how.** Everything whose real counterpart moves on the scale
of seconds: water and sea, rain, mud, lava, ice, sand, cloud; the barber pole
and candy stripes climb along their direction; neon flickers; the lava lamp's
metaballs rise and merge; plasma cycles its palette; scanlines roll; slime
drips along its direction; the glitter twinkles and the rainbow cycles as they
do. Each takes a rate parameter in real units, mm/s or s per cycle, as
`polkaDotDrift` and `rainbowCycle` do. What should not move inside the
library: weathering. Rust growing or paint fading over time is a document's
story, and the document already has the tool, a slider's `.period`, which
sweeps `amount` without the library reading `time` at all. The same goes for
any parameter an author might want to animate: a slider with a period beats a
baked clock, and it keeps the material still for everyone who does not want
the motion. And nothing that needs the light or the eye can move with them:
a colour field sees `p` and `time` only, so sheen and iridescence are
rainbow-in-space, and `glitterTwinkle` turns a half-way vector of its own.

## Detail

The thumbnail is 180 mm over 96 pixels, so a feature fades below 7.5 mm and
is gone below 3.8 mm, and the 240 pixel preview fades below 3 mm. That is why
a quarter of the catalog is drawn from its stand-in in the grid, and why an
entry judged against its thumbnail looks finished when it is not. There are
three scales an entry can have, and not every entry wants all three: a
polished finish is detail alone, a flat paint has none of them, a
checkerboard is a pattern and nothing else, and each is complete. What every
entry does is declare, in its doc line, which it has and which it leaves out
and why, so that a missing layer is a decision on record and not an
oversight.

- **The pattern**, at its size: the continents, the dots, the
  crystals, the ripples. This is what the grid shows.
- **Relief**, in millimetres and under the band: the ridges, the knurl, the
  ripples' profile, the bars. This is what the preview shows.
- **Detail**, below the pixel at the preview's distance: grain speckle,
  micro-roughness, fine relief as `.detail`, the lobe's hidden slopes. This is
  what the user sees when they lean in, and what the stand-in averages away.

The audit of what exists, in that spirit. "Complete" means the layers it
lacks are ones its real counterpart lacks, or ones the entry's purpose
excludes; "missing" names a layer the real thing has and a user leaning in
would expect.

| entry or family | has | lacks, and why | verdict |
| --- | --- | --- | --- |
| the lobe-only finishes, polished metals, chrome, lapped | detail: the lobe | no pattern or relief: a mirror has none | complete |
| flat paints, plastics, fog, glass, the measured metals | none | a substance with nothing to see at any scale | complete |
| checkerboard, polka dots, cell skin, stripes to come | pattern | no relief or detail: a paint pattern on whatever substance carries it; a plastic micro-texture is the substance's business, as a finish | complete |
| glitter | detail: the flakes | no pattern or relief: the flakes are the look | complete |
| coats: powder, wrinkle, sand, hammertone | pattern, relief, detail (two orange-peel bands, the grain, the ridges, per-dimple shade) | | complete |
| printed layers and powder | pattern, relief, detail (layers as `.detail`, powder grains and specks) | | complete |
| machined | relief (grooves, scallops, knurl), detail (the lay the grooves fold into; the blast's speckle, a shade per dimple of the media's size) | no pattern: a worked surface has none | complete |
| wood | pattern, relief (weathered), detail (fibre, pores) | | complete |
| granite, the marbles | pattern, detail (seams; the fine vein network) | no relief: a polished stone is flat | complete |
| flagstone | pattern, detail (a stone grain per slab, 0.8 mm) | relief by design (the joints are paint, the doc says why) | complete |
| terrain | pattern, relief, detail (the beach's sand grain; the rock's granite seams, which an earlier audit missed) | | complete |
| lava, lava hero | pattern, relief (hero), detail (a 0.9 rough crust) | | complete |
| water, cloud, rain, mud, ocean, ice, lava lake | pattern, relief, and the sea's capillary slopes as roughness | a smooth liquid has no grain | complete |
| sand ripples | pattern, relief, detail (the grains' relief and dark specks, 0.3 mm at any model scale) | | complete |
| weathering | pattern (blotches, streaks), detail (pits at 1.5 mm) | relief: rust's scale is an option | complete; fine rust grain below the pits is an option |

Sand's grain landed first (2026-10-29), because it is the one a customer will
lean into: `texture.grainRelief(size, height)` and `grainSpecks(size, share)`
are the grains' relief and a share of them picked at random, both fine detail;
`natural.sand(grain)` puts them over `physical.sand`'s matte and colour (its
roughness is already 1, so the grains' slopes have nothing to add), the
terrain's beach is that sand, the sand ripples carry the same grain at 0.3 mm
whatever their model scale, and the sand coat's grain is the same relief
helper. The terrain's rock turned out to draw granite's seams already, at
the rock's own cell size, through the colour it mixes; the shader gates them
at the crystal size. One entry is still missing a layer: the flagstone's
slabs take a stone grain. For a new entry the reviewer's question
is not "does it have all three" but "does its doc line say which it has, and
is the layer it leaves out one the real surface leaves out too".

## Scale

The language's unit is the millimetre, and the catalog's sphere is 100 mm
across because a material's features are in millimetres too. The audit of
2026-10-27 found the following; the size pass of 2026-10-28 resolves the items
marked so, and the rest stand.

- **The libraries do not state the unit.** The language docs do, and
  `material.gsl` does for absorption and displacement only. Each library's
  head should say every length is a millimetre, and name the 100 mm swatch
  as the size presets are drawn for. Open.
- **Presets inherited example scales.** The ready-made section says so in its
  own comment. Terrain was 3 mm continents, Terrain Bracket's argument, on a
  100 mm sphere; it is 40 now, two or three continents across the face, and
  with the size a `@range`, a user resizes any preset to their part, so the
  swatch's size is a default rather than a verdict.
- **Three presets carried a hidden part size**: `knurled` computes its teeth
  for a 16 mm diameter, `faceMilled` assumes a 50 mm cutter, `flagstone` is
  250 mm slabs. All three are `@range` parameters now, so the panel shows the
  size and the user sets their own.
- **Two presets were scale models with an unstated ratio.** `ocean` rendered
  real 83 m waves at 0.12 mm per metre, 1:8300; `sandRipples` at 60 mm per
  metre, 1:17. Both now take the wavelength on the part in mm and derive the
  model scale from the wind, and their docs state the ratio. The consequence
  to know: in the Sea and River and Sand Ripples examples the wind slider no
  longer lengthens the waves, since the wavelength is now the input; it still
  changes their speed, height and whitecaps.
- **What the thumbnail can show.** The frame is 180 mm over 96 pixels, so a
  feature fades below 7.5 mm and is gone below 3.8 mm; the 240 pixel preview
  fades below 3 mm. A quarter of the catalog, every machined, printed, powder,
  sand and wrinkle finish, is drawn from its stand-in in the grid by
  construction. Say so in `examples/viewer/materials`' README so nobody tunes a
  preset against the thumbnail.
- **Noise periods are fine.** A cellular field repeats every 64 cells; the
  shortest in the library is glitter's 1 mm flakes, every 64 mm, once across
  the swatch. Everything else repeats above 250 mm.

## Realism: judged against photographs

The goal is surfaces that look nearly photographed. The bar is set
by measurement, not by taste alone. Each flagship surface is sampled flat
with the `gsl` CLI, supersampled at a stated millimetres per pixel. Its
colors are decoded from sRGB, since a GSL color is sRGB-encoded and the
renderers decode it to linear albedo. The sample is then set beside a
reference cropped to the same scale.

**References.** The references are scanned albedo maps with a stated
physical size where they exist: Poly Haven's CC0 scans, such as
rust_coarse_01, the oak and walnut veneers and wood_textured_concrete.
Otherwise they are Wikimedia Commons photographs, cited where used. A
photograph's scale is found from a feature of known size, or assumed and
said so. ambientCG's marble, granite and brushed sets are procedural, not
photographs, and are not references.

**Statistics.** The comparison covers:

- the Oklab mean, spread and percentiles of L;
- chroma and hue;
- the share of L's variance in each octave of wavelength, in millimetres,
  which carries both a pattern's scale and its contrast;
- the anisotropy of the spectrum.

Each comparison ends in a side-by-side judged by eye. Brushed metal and
the cast skins live in relief more than in albedo, so they are judged from
renders instead, on structure rather than tone.

**First result, weathering** (2026-11-03). `weathered` painted one flat
corrosion color through a soft half-width of 0.5, against a field whose
spread is 0.22, so nearly every point sat in the transition. Full rust was
a single color. It now has:

- **Patches** from 5 to 100 mm: a 60 mm field under the 15 mm blotches,
  scaled to the old field's spread.
- **Crisp edges**, at a half-width of 0.08.
- **Thickness:** corrosion is lighter and yellower where thin, as fresh
  ochre rust is over old brown, and stains and dulls the bare surface
  around a patch.
- **Inside a patch**, a mottle at 5 and 40 mm and clustered roundish dark
  pits.

Full rust's color is the scan's mean. Against it, Oklab L is 0.410 against
0.412, L's spread 0.037 against 0.035, chroma 0.077 and hue 52 degrees in
both. The stand-in is the shares' sum in Oklab: the corroded, stained and
deep shares, each a mask fitted to the measured share to 2%.
`weatheredRunoff` keeps its soft field for now.

**Second round** (2026-11-03). Drafted in parallel, each family against its
own references, then ported:

- **Granite** was a mosaic of one size, its dark mineral pure black. It is
  now three sizes:
  - feldspar crystals at `cellSize`;
  - quartz grains at 0.45 of that, clumped and zoned, seen as glass;
  - dark aggregates at 0.15, with a floor at 3% albedo.

  The boundaries are warped and soft. Against Azul Platino, photographed
  with a scale bar, Oklab L is 0.655 against 0.651, its spread 0.133
  against 0.122, and the darkest 5% 0.392 against 0.414; it was 0.000. The
  grey presets take colors fitted to that photograph. `natural.terrain`'s
  rock takes the new mosaic too.
- **Marble.** Statuario is a connected network of vein sheets: a 2D Voronoi
  edge on the face, its cells drifting through depth, so no wall lies in
  the face and smears. Its L spread is 0.114 against 0.121; it was 0.059.
  Carrara is a bright clouded ground with feathered wisps and a sugary
  grain, at L 0.948 against 0.947 and a spread of 0.023 against 0.022; it
  was 0.860 and 0.072.
- **Wood colors.** White oak, red oak and walnut take colors moved toward
  published colorimetry and calibrated veneer scans. Mokume's photographs
  had red oak too dark and red, and walnut too gray. The figure is
  unchanged. The apparent excess of 1 to 3 mm variance was the harness
  aliasing pores at 2x2 supersampling, so sample fine bands at `--ss 4`.
- **Sand-cast skins** (judged; no scanned reference exists). The grain is
  15/GFN mm and Ra about 1000/GFN um. The imprint is a relief of rounded
  bowls, with gas holes. The skin is darker, duller and mottled, with
  burnt-on sand.
- **Terracotta** is new: `ceramic.terracotta(clayColor, grogSize)`. Its
  grog, mottle and inclusions match a tile photograph's L spread (0.010
  against 0.011) and spectrum. The preset is in physical's measured color.
- **Board-formed concrete** keeps its relief. Its albedo gains a tone per
  board, grain streaks, dark seams, pale leaching and bugholes, and it
  matches a scan's spread and anisotropy.

**Third round** (2026-11-04):

- **Leather** is new, as `organic.leather(hideColor, grainSize, depth,
  wrinkle, aniline)`. It is a pebbled grain of domed cells bounded by
  creases, with a wrinkle network, fine pores and an aniline or finished
  tone. It is matched to Poly Haven's brown_leather and fabric_leather_01
  scans: L 0.343 against 0.340 and a spread of 0.038 against 0.034. In
  every scanned grain the creases are paler than the cells, not darker. The
  presets are `brownLeather` and `blackLeather`. The sheen waits for a
  second lobe.
- **Painted metal.**
  - `chippedPaint`'s chips come in five octaves, clustered, with
    notched edges. Their sizes follow a power law, N(>d) ~ d^-1.70 against
    the scans' -1.63 to -1.77. Primer shows in small chips and bare metal
    inside a primer ring in larger ones. The paint is faintly chalked.
  - `powderCoat` takes `glossRoughness` (its own copy of the old formula
    read about 21 GU too rough). Its orange peel is 1.5 and 0.65 mm waves,
    with measured slopes.
- **`weatheredRunoff`** builds on `weathered`'s patches, drawn out down the
  flow. Drips about 3 mm wide run from each patch and fade over the streak
  length. The old linear bias along the flow is gone, so its stand-in is
  the true mean.
- **Stone.**
  - Flagstone's joints keep one width along every wall, with a tone per
    slab, fitted to a scanned floor.
  - Slate's riven layers are terraces of noise whose edges wander and
    nest, with flecks and streaks along the grain.
  - Travertine's pits are small and ragged, drawn out along the bedding
    and gathered in porous layers, matched to a photograph of Classic
    travertine.

`noiseHash3` returns multiples of 1/255, so `texture.grainSpecks` rounds a
share up to whole 255ths; `masonry.sparseSpecks` pairs two channels for
shares under that. Brushed metal is held: its scratches compile, but they
cost about 15 times the frame and break up on curved parts.

## Naming and non-infringement

Names here are process and specification names in camelCase under a package,
which nobody can own. Do not use, from the market list or anywhere:

- **Numbered series**: a name followed by a number and a colour.
- **Adjective-first names**: a finish adjective, a colour and a substance.
- **Trademarks** for generic surfaces. Say spark-eroded texture, not the
  mold-texture vendor's name; ceramic coat, not the brand; PTFE, aramid,
  acetal, solid surface, phenolic laminate, acrylic sheet, hook and loop,
  polystyrene foam, studs. If a trademark is the only name a reader would
  know, the doc line may say "often sold as" once.
- **Their editorial sets**: seasonal and holiday palettes are content, not
  surfaces.
- **Their imagery**: no swatch is modelled on one of theirs.

Two of their families are not materials and are refused as such: perforated
and expanded sheet are holes, which belong to the shape libraries; the clays
are a sculpting app's identity, and `physical.clay` with a fingerprint
operator is as far as `omega-core` goes.

## Order of work

1. One size per pattern (landed), then thin film (landed, with heat tint and
   anodized titanium as its first users): small, and it unblocks the most.
2. The detail pass on what exists (landed): sand's grain and the terrain's
   beach, flagstone's slabs and the blast's speckle, all on `texture`'s
   grain helpers; the terrain's rock had granite's seams all along. Small,
   and it sets the bar the new families are held to.
3. Frame forms for the plane patterns that exist, `polkaDot` and
   `checkerboard`, and a direction for the rainbow bands; the fold of the
   lava and polka pairs into overloads landed with the size pass. Landed
   2026-10-29: `checkerboard(cellSize, a, b, frame, diameter)`, the knurl's
   way, an even count of cells round u nearest the circumference's so the
   board closes across the seam; `rainbowSheen(scale, direction, cmapFn)` and
   `rainbowCycle(scale, period, direction, cmapFn)`, the z forms kept; and
   `polkaDot(spacing, dotRatio, a, b, frame, diameter)`, a whole count of
   honeycomb cells, each sqrt(3) spacing wide, round u. Overloads resolve by
   arity alone, so the frame form took
   the drift's six and the drift moved to seven, gaining the plan's gizmo
   origin: `polkaDot(spacing, dotRatio, origin, direction, a, b, velocity)`.
   The direction form stays at five. Step 3 is landed.
4. `printed` and `plated`: the customer's process and the finishes on the
   fasteners already in the library. `plated` landed 2026-10-29: nine
   operators over a base, each a plated metal's measured colour under a
   judged lobe, the clear, yellow and black chromates and cadmium's yellow
   one as thin films with the chromium compounds' tint, a preset each over
   the metal the plate is usually on. `printed.sintered(base, color,
   roughnessRa)` landed the same day, the binder-jet skin as fine detail,
   grain speckle and sparse pores, with four alloy presets, 17-4PH, 316L,
   304L and copper over physical's stainless and copper, a guess at the
   customer's catalogue and labelled so in the file; `tumbled` and `polished`
   for `machined` remain.
5. `machined` additions, `sheet`, `cast`, `molded`: nothing new needed.
   Landed 2026-10-29: `honed(m, c, frame, angle, grit)` on a bore's frame,
   `engineTurned(m, c, origin, feedDir, normal, spot, stepover)` and
   `ballMilled(m, c, origin, feedDir, normal, ballRadius, stepover)` on face
   datums, each with its Wrap overload; then `scraped(m, c, flakeSize)`,
   `lapped(m, c, roughnessRa)`, `sparkEroded(m, c, vdiGrade)` (Ra = 10^(grade
   / 20) / 10 um, the plan's formula missing the tenth), `tumbled(m, c,
   mediaSize)` and `polished(m, c, mirror)`, the post-processes the printed
   and plated parts take too. machined's additions are landed. `sheet` landed
   the same day: `hotRolled(m, c, flaking)` under its blue-grey scale, flaked
   in patches; `coldRolled(m, c, direction)` with the rolls' lay;
   `pickledOiled(m, c, oiled)`; `checkerPlate(m, c, origin, along, normal,
   pitch, barLength, barHeight)`, the four-way lugs as a returned height,
   which ASTM A786 leaves to the mill so the sizes are a supplier's; and
   `stuccoEmbossed(m, c, pebble, depth)`, its pebbles a height. A preset each
   over steel or aluminium, the heights baked in. `cast` followed:
   `sandCast(m, c, grainFineness)` by AFS number, the sources giving the
   trend and the range but no formula, with the sand's grain as fine detail;
   `dieCast(m, c, direction)` with its flow lines; `investmentCast(m, c,
   roughnessRa)`; `forged(m, c, scale)` under its blotched scale; and
   `shotPeened(m, c, almenIntensity)`, its dimples' relief and shade as fine
   detail, the dimple size judged since the sources tie it to the shot and
   the part's hardness rather than the intensity alone. `molded`'s first
   half: `textured(m, c, vdiGrade)` on machined's spark-eroded skin,
   `leatherGrain(m, c, grainSize, depth)` as relief, `glassFilled(color,
   fibreShare)` with its flecks as fine detail, `grip(m, c, frame, diameter,
   pitch, height, diamond)` and its face form, the dots or diamonds a
   returned height over the base's relief buried by `coating.levelled`, and
   `silicone(color, clarity)`, undisplaced by the rule. Then the engineering
   polymers, `acetal`, `ptfe`, `peek`, `uhmw` and `absPlastic` (abs being
   the absolute value), each a lobe and an index over a colour with a
   natural-shade preset, the indices cited where found and judged otherwise.
   Step 5 is landed.
6. The weave, then `composite` and `textile`. The weave and `composite`
   landed 2026-10-30: `carbonTwill` and `carbonPlain` on a frame with a
   diameter or on face datums, `forgedCarbon(chipSize, gloss)`,
   `glassLaminate`, `aramid` and `phenolicLaminate` on face datums, all
   bulk materials of the one `fabric` helper under `coating.levelled`
   resin, with no operator form, since they have no base. `textile` waits
   on the sheen lobe for velvet, felt and suede; its weaves can follow.
7. The tessellation, then `masonry`. Landed 2026-10-30 in two slices:
   `brick` in its bonds under raked mortar, `subwayTile`, `hexTile` and
   `pennyTile` in grout, bulk materials on face datums with a preset each;
   then `texture.herringbone(x, y, width, aspect, joint)` (blocks a whole
   number of widths long, in bands of staircases along the diagonal) under
   `herringboneTile` and `parquet`, the parquet each block's own flat-sawn
   white oak from the wood package under a polyurethane; `concreteBlock` on
   the brick's tessellation at a 2:1 course; and `terrazzo`, cellular chips
   in three colours with the wall band as matrix, the chip share measured
   for its stand-in. Then the poured and trowelled: `boardFormed` on face
   datums (a level, a pressed grain and a fin per board), `exposedAggregate`
   and `asphalt` with cellular feature points as stones of a radius, so a
   face cuts each at its own depth and they come in every size, standing
   proud as domes (shares and dome means measured), `polishedConcrete` with
   the share cut through rising as the cube of the grind's depth, and
   `plaster` burnished where the trowel pressed. Last, the stones the
   table listed under masonry: riven `slate`, `travertine` with its pores
   cross-cut, vesicular `basalt`, `obsidian` flow-banded or snowflaked,
   and `engineeredQuartz`; a pore, a bubble or a flake is a feature point
   with a radius, and `poreSpacing(radius, share)` inverts the cube fit so
   a share asked for is the share drawn. They were built in `masonry` as
   the table said and then moved to `material/bulk/stone`, Blake's call, so the
   natural stones have one home beside granite, the marbles and flagstone;
   `masonry` keeps the laid, poured and trowelled. Step 7 done.
8. The rest of tier 2, then tier 3 with `classic` first, since it needs
   nothing new; the normal slot and edge distance when `weathering` reaches
   them. Started 2026-10-30 with `paper` (kraft, chipboard, uncoated and
   coated sheets, corrugated board's liner over its flutes on face datums)
   and `foams`, plural because `lattice/foam` already owns the `foam`
   prefix (open cell as pits over half the face, closed cell skinned,
   beaded polystyrene, aerogel as a transmissive block at index 1.03).
   Then `ceramic` (glazes as coats over a body, so they wrap as operators:
   celadon, crackle with its line spacing as the glaze's misfit, speckled,
   matte; alumina, zirconia and silicon nitride as bulk) and `glazing`,
   the patterned architectural sheets, reeded, hammered, wired and
   dichroic, named for the trade since `glass` holds the plain media and
   stays free of imports; the dichroic draws the film's transmitted
   complement as the sheet's tint, the reflection waiting on a film term
   in the renderer. `textile` still waits on the sheen lobe. Tier 3 then
   opened with `classic`: Perlin's marble and turbulence, Peachey's wood,
   Worley's mosaic and crumple, the plasma, the lava lamp, Phong's
   plastic, the Cornell walls, Truchet's tiles, scanlines as a finish and
   operator, pixel art and the chrome ball, each after its paper with the
   paper cited in the file; the stand-ins measured where a mean was not
   closed-form. Then `organic`: orange peel and cork as cellular pores at
   a share from the cube fit, bark as plates over fissures, a bamboo culm
   about a stem, wicker on the weave, coral's cups, eggshell's speckles.
   Then the playful additions, in the one package: stripes and candy
   stripes across a direction, the barber pole on a frame in time,
   chevrons, camouflage in four colours, the leopard, zebra and giraffe
   prints, tie-dye, halftone, hex grid, ombre; and sprinkles, studs,
   bubble wrap, scales, wrinkle, swirl, slime, gummy, neon and the
   holographic film, drawn in place where the real foil turns with the
   view. Left in tier 3: the weathering operators, which wait on the
   normal slot and an edge distance; and `textile` on the sheen lobe.

Each package follows the existing pattern: constructors with annotated
parameters, one `@range` size in mm per pattern with the lengths inside it as
ratios, a relief height as its own `@range` in mm where there is relief, a frame
or a datum as "Datums" decides, a tier overload where there is a cost, a
Ready-made section of zero-argument `@material` presets sized for the 100 mm
swatch, `@surface` overloads for operators, stand-ins by share or measured,
a doc line that says which scales the entry has and which it leaves out, a
role of bulk, finish or coat readable off its signature, citations in the
file for every sourced number, and the catalog regenerated with `go run
./tool/materialswatch` for new entries or `-manifest-only` for signature
changes.
