# Relief Datum Audit (2026-11-27)

Every library site that sets a relief, against D1 (relief is measured down from its peaks: its
maximum over the surface is 0, and it runs from `-depth` to 0). File paths are under
[stdlib/](../../../stdlib/); line numbers are as of commit `158c6fc6`.

- `fd` is `texture.fadeDetail`, in [0, 1] and 1 outside the renderer. "detail" means the term
  sits in `.detail`, so it is 0 where that is not shown.
- Noise bounds ([noise.gsl](../../../stdlib/noise/noise.gsl)). `g3` is `gradientLattice`: its
  theoretical bound is about ±0.87, the doc says about ±1.5, and fbm's measured extremes are
  about ±0.375. `g4` is `gradientLattice4`, ±1. `turbulence2` lies in [0, 1.3] and `turbulence3`
  in [0, 1.52].
- `texture.grainRelief(s, h)` is raised from a base: [0, h], 0 in the gaps. Seven sites negate it,
  which measures from the peaks. Six do not, so it is fixed per call site, not in the helper.

**Decided 2026-11-27.** The process and coat families move to the peaks datum, each family shown
before its goldens move, and each declares its depth. Bulk and collection materials stay as judged
appearance, documented as not measured from their peaks.

## material/process

| site | function | range | datum | shift to peaks; stand-in | basis |
| :--- | :--- | :--- | :--- | :--- | :--- |
| machined.gsl:58 (.simple :59) | grooveFinish | base + [-(r - sqrt(r^2 - f^2/4)) fd, 0] | peaks | 0; stand-in -f^2/12r is the mean | measured |
| cast.gsl:52 | sandCast | base + [-4.4/G - 0.06, ~5.2/G] | mean | ~ -5.2/G; stand-in (0) too | judged fit |
| cast.gsl:195 | shotPeened | base + [-0.08 d, 0] | peaks | 0 | judged |
| printed.gsl:54 | layerFinish (fdm) | base + [-0.393 L, +0.107 L] | mean | -0.107 L; stand-in too | measured geometry |
| printed.gsl:54 | layerFinish (resin) | ±0.04 L | signed | -0.04 L | judged |
| printed.gsl:115 | powderBed | [-0.06 g, +0.24 g] | signed, detail | -0.24 g | judged |
| molded.gsl:52 | leatherGrain | base + depth [-1, +0.5] | signed | -0.5 depth; stand-in (0) is not the mean | judged |
| molded.gsl:214 | @rubberGrip | displacement [0, 0.8] on a build of 2 | base, and broken | -h; the build bug first | judged |
| sheet.gsl:160 | @checkerPlateSteel | [0, barHeight] (1.2) | base | -1.2; stand-in 0 | judged |
| sheet.gsl:170 | @stuccoEmbossedAluminum | [0, depth] (0.2) | base | -0.2; stand-in 0 | judged |

## material/coat (and ceramic)

| site | function | range | datum | shift; stand-in | basis |
| :--- | :--- | :--- | :--- | :--- | :--- |
| coating.gsl:171 | powderCoat | abs <= 0.009 peel | signed, detail | ~ -0.009 peel | amplitude stated |
| coating.gsl:185 | wrinkleCoat | [-0.16 s, +0.035 s] | mean | -0.035 s | judged |
| coating.gsl:197 | sandCoat | grainRelief [0, 0.4 g] | base, detail | -0.4 g | judged |
| coating.gsl:218 | hammertone | [-0.035 c, +0.025 c] | mean | -0.025 c | mean measured |
| coating.gsl:268 (stand-in :385) | chippedPaint | base + [.., 0.05 U0 + 0.03 U1] | mean | depends on wear; stand-in too | judged |
| coating.gsl:484 | levelled | relief(coat) + keeps relief(base) | inherits (keeps D1) | none | filmKeeps judged |
| coating.gsl:509, ceramic.gsl:26 | clearCoat, glazes | film 0, then levelled | inherits | none | as levelled |

## material/bulk and material/collection (left as judged appearance)

| function | datum | shift at the parameters' midpoints |
| :--- | :--- | :--- |
| corrugated, cork, eggshell, openCell, closedCell, glazedTile and its tiles, brick, concreteBlock, hammered, travertine, basalt, wood (fresh, sanded), decayed | peaks | 0 |
| bamboo, wicker, beaded, sand, raisedTerrain, exposedAggregate, asphalt, reeded, crumple, sprinkles, studs, bubbleWrap, scales, gummy, iceFloes, sandRipples, fabric | base | 0.1 to 6 mm |
| orangePeel, bark, coral, leather, water, lava, cloud, boardFormed, plaster, graniteFacet, slate, wood (weathered, sawn, torched), wrinkle, lavaLake, rain, river, mudPot | signed or mean | 0.07 to 6.5 mm |

## mechanical (size-defining)

| site | function | range | datum | basis |
| :--- | :--- | :--- | :--- | :--- |
| thread.gsl:466, :492 | maleRelief, maleSinRelief | [-0.5413 p, 0] | peaks (crest at the major diameter) | ISO 68-1 |
| thread.gsl:482, :505 | femaleRelief, femaleSinRelief | [-0.5413 p, 0] | peaks | ISO 68-1 |
| fastener.gsl:1423, :1451; joint.gsl:126, 135, 238, 246, 290, 298 | thread heights | as above | peaks | ISO 68-1 |
| shaft.gsl:308, 389, 442 | knurl relief | [-pitch/2, 0] | peaks | ISO 13444 |

## Found on the way (fixes decided 2026-11-27)

1. **Replace instead of add.** `shaft.body` (shaft.gsl:389, 442), `fastener.maleFinish` and
   `femaleFinish` (1423, 1451) and `@rubberGrip` (molded.gsl:214) call `withDisplacement`. That
   drops the finish's own relief (a turned finish loses its grooves) and counts a finish's build as
   relief: `@rubberGrip` over a 2 mm build ends up with relief from -2 to -1.2.
2. **`chippedPaint` blends the volume row** (coating.gsl:262): where the paint is intact, a built
   base's build turns into relief.
3. **The composite fabric stand-in** (composite.gsl:45) drops the 0.15 mm resin build and the
   `keeps` share, so it jumps about 0.1 mm at the stand-in switch.
4. **Documented amplitudes differ from the code.** water, lava and cloud (natural.gsl) peak near
   0.1 x amount x size, not at "amount mm". reeded and stuccoEmbossed are raised from a base,
   though their docs say "deep". orangePeel's "0.2 mm swell" is about ±0.075 in practice.
5. **Pass-through finishes.** phosphate, blasted, tumbled, lapped and sparkEroded change roughness
   and lay only, so a base's turned grooves survive under them as displacement. This is noted, not
   decided.
6. **A sum keeps the maximum at or below 0 but not exactly at 0.** machined over a base, levelled
   and decayed add reliefs that are each measured from their own peaks. `reliefDepth` adds the
   depths, which is conservative.
