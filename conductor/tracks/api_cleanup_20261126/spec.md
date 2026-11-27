# Specification: GSL API Cleanup

## 1. Why

A review of the GSL API and the documentation `gsl skill` generates for models (2026-11-26) found names
that say the same thing several ways, leftovers kept beside their replacements, two spinodoid builtins
that draw one field, and a render API split into eight entry points. The documentation and error
fixes it called for are in their own pull requests: #172 (doc fixes and a test that builds every
documentation example), #175 (tiered `gsl skill`, member lookup and search) and #174 (did-you-mean
errors, the remap point check, inline `@range`, and two ways a function silently vanished). This track
covers the rest: changes that break existing documents or add features, each its own pull request.

Every decision below was made item by item in that review. Where a name or an interpretation is still
open, the decision says so (section 9).

## 2. Constraints

- **Breaking changes ship with a rewrite.** A renamed or removed stdlib name gets a `gsl fix` rewrite
  ([cmd/fix](../../../cmd/fix)) in the same pull request, and keeps an alias for one release where the
  decision says so.
- **omega-dm copies stdlib functions.** The omega-dm repository keeps a running list of the stdlib
  functions it copies into its own libraries. Before any rename or removal lands, its uses and copies are
  inventoried (plan, phase 0), and each change says how omega-dm follows: update its copies, or pin the
  old name.
- **Goldens move only with approval.** Stdlib edits move the example goldens under `testdata/`. A change
  that moves one shows what moved and why, and the golden is updated only once that is approved.

## 3. Decisions: Names

| #   | Area              | Today                                                                                                                                                                                                  | Decision                                                                                                                                                                                        |
| :-- | :---------------- | :----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| N1  | `map` package     | Shares its name with the `map(fn, list)` built-in: `map.checker(...)` and `map(f, xs)` sit side by side                                                                                                | Rename the package to `domain`, the SDF term for these operations; `gsl fix` rewrites `map.` member references                                                                                                      |
| N2  | `matrix.scale`    | `matrix.scale(s, m)` multiplies a matrix by a number, `matrix.scaling(factors)` builds a scale matrix, `transform.scale(shape, factor)` scales a shape                                                  | Rename `matrix.scale` to `multiplyScalar`, the name of its twin `vector.multiplyScalar`, so "scale" means scaling geometry. The lowering of `*` and `/` calls it, so the rename updates `internal/lower`; the old name stays as an alias for a release                                                                                                     |
| N3  | Axis rotation     | `transform.rotate(shape, origin, direction, angle)`, `map.rotateAxis(v, axisOrigin, axisDir, angle)`, `matrix.rotation(axis_origin, axis_direction, angle)`; `map.rotateXY(v, ang)`                    | One vocabulary: `(…, axisOrigin, axisDirection, angle)` everywhere, `ang` becomes `angle`. Parameter names only; nothing breaks                                                                 |
| N4  | Lattice size      | The cell size is `sizeIn` (beam), `size` (foam), `cellSize` (lattice, honeycomb), `period` (tpms, spinodoid), `spacing`, `cell`                                                                        | `cellSize` everywhere, except `period` in tpms and spinodoid, where the field's wavelength is the real meaning. Parameter names only                                                             |
| N5  | Honeycomb names   | `honeycomb.squareHoneycomb`, `hexagonalHoneycomb`, … repeat the package name; `foam.cubic` does not                                                                                                    | `rectangular`, `rectangularDiagonal`, `rectangularReinforced`, `hexagonal`, `triangular`, `octet` (`square` is a reserved primitive, and the cells are vec2 rectangles); a hard rename, `gsl fix` rewrites                                                                                                          |
| N6  | Three `bcc`s      | `lattice.bcc`, `beam.bcc`, `foam.bcc`                                                                                                                                                                  | See section 6a: every beam lattice takes a `combine`, `lattice.bcc` is deprecated into `beam.bcc`, each doc names its complex, and the missing duals are added                                                                           |
| N7  | Box corners       | `box(pointA, pointB)`, `cropToBox(shape, low, high)`, `extrude(profile, low, high)`, and `.box = [lo, hi]`                                                                                             | `min`, `max` for every pair of box corners. Parameter names only. `min`/`max` already work as parameter names; #174 makes every built-in name work                                              |
| N8  | Placement words   | `center`, `origin` (cone), `point` (plane), `vertex` (paraboloid)                                                                                                                                      | Keep `origin`, `point` and `vertex` where they are the right geometric word                                                                                                                     |
| N9  | Centred overloads | `torus` and `octahedron` exist only at the origin; `sphere` has both forms                                                                                                                             | Add `torus(center, majorRadius, minorRadius)` and `octahedron(center, size)`                                                                                                                    |
| N10 | Boolean names     | Most booleans take `d1, d2`; the exact ones take `field1, field2`. Unions and intersections can be passed as an operator (`unionRound(2)`); differences cannot                                          | `a, b` across all booleans, as the docs already write them. Add difference operator forms (`differenceRound(r)` and the rest), taking a list `[target, tool, …]`: the target minus the rest      |

## 4. Decisions: Removals

| #   | Today                                                                                                                                                                         | Decision                                                                                                                                                   |
| :-- | :---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | :--------------------------------------------------------------------------------------------------------------------------------------------------------- |
| R1  | The vector arithmetic helpers (`addVector`, `addScalar`, `subtractVector`, `subtractScalarFrom`, `subtractFromScalar`, `multiplyScalar`, `multiplyComponents`, `divideByScalar`, `divideScalarBy`, `divideComponents`, `negate`) and `matrix.scale`, `plus`, `minus`, `multiply`, `negate` carry `///` docs. No document calls them, but the compiler lowers `+ - * /` on vectors and matrices to them ([internal/lower/lower.go](../../../internal/lower/lower.go)) | Hide, never remove: drop their `///` docs, as the `ui*` calls the annotations compile to are left out. They stay where the lowering finds them (unary minus lowers to the two `negate`s) |
| R2  | The boolean solver's internals (`edgeProjectFrom`, `edgeSlideFrom`, `edgeVerdict`, `edgeNewtonStep`, `cornerStepFrom`, `cornerVerdict`, `cornerNewtonStep`) carry `///` docs, so they are public API, though only `booleans.gsl` calls them | Drop their `///` docs. No code change; they leave the docs and `gsl skill`                                                                                  |
| R3  | `map.gsl` documents two functions against lattice's "legacy `planesEven`", which no longer exists                                                                             | Reword the two doc lines                                                                                                                                   |
| R4  | `program/preamble` filters libraries by a product `Domain`; no library sets one                                                                                               | Remove the filter and its `includeDomains` parameters. Keep `SourceLibraryDescriptiveText` and `PackageSummary`: another repository may call them           |

## 5. Decisions: Matrices

The matrix package writes each operation once per size (`transpose2/3/4`, `identity3/4`, `determinant2/3/4`,
`inverse2/3/4` and `inverse`, `rotation`/`rotation3`, `scaling`/`scaling3`), with no rule for which size goes
unsuffixed. Collections now hold any number of elements, so the operations can be written once:

- `transpose`, `multiply`, `identity(n)`, `determinant` and `inverse` work on a matrix of any size.
- The sized names stay as aliases for one release, then `gsl fix` rewrites them.
- The constructors (`rotation`, `scaling`, `translation`, `rotationEuler`, `compose`) stay 4x4, with their
  3x3 variants.

Shown in pure GSL during the review, and constant-folded: a transpose of any shape (a `map` over the rows of a
`map` over the columns), a product of any compatible sizes (a `map` over b's columns, each a `reduce` of
`a[k] * x`: 2x3 times 3x2 gives the right 2x2), and a list's length at compile time
(`reduce(0, (acc, x) { return acc + 1; }, list)`). Not shown: determinant and inverse of any size, which need
recursion on the size or an unrolled elimination. The plan prototypes them before anything is renamed.

## 6. Decisions: Spinodoids

**Two versions, both everywhere.** The spinodoid has exactly two algorithms. `gaborSpinodoid` is the reference
field: jittered kernels, the definition the other is measured against, and the expensive one.
`fastGaborSpinodoid` is the fast version: the lattice field, built for the GPU. Each is evaluated by both the
CPU (fidget) and the GPU (WGSL) backends. No other versions exist. In the library, `gabor*` draws the reference
field and `fastGabor*` the fast one.

Each is its own oracle in every backend, with its own table, slots and leaf ids: WGSL codegen and fidget
conversion take the kind from the builtin's name. The reference table is baked in Go
([program/noise/spinodoid](../../../program/noise/spinodoid)), parity-tested against the Python reference
implementation, and evaluated by its Rust port in the fidget tape core and by `gsl_spinodoidRef` in WGSL. A
tape names which oracle each leaf is, and a reader loads it as that one unless it asks for the other.

**The table's box is given, not inferred.** A graded field (scale or orientation varying with `p`) bakes a
bounded kernel table today, whose box comes from the imported geometry's bounds or the `@scene` cube, and falls
back to a periodic table, which repeats the grading, when neither is known. Decision: the box is specified in
the call, never taken from the scene or inferred. The scale is already explicit and stays as it is: the call's
`scale` argument multiplies the wavelength, and the site spacing follows from `period`, `packetWidth` and
`impulses` ([call.go](../../../program/noise/spinodoid/call.go)). The periodic layout of a stationary field
(constant scale and orientation) and the bounded layout stay; they are how a table is stored, not versions of
the field.

## 6a. Decisions: Lattices

**What each lattice is.** A beam lattice is the bond struts of a point lattice; a foam is walls.

| Lattice | `beam.X` | `foam.X` | Relation |
| :------ | :------- | :------- | :------- |
| cubic | struts on the lines through the cell centres | walls on the planes through the cell centres, which hold the struts | the same complex; the dual is `foam.cubicGrid`, whose walls lie halfway between |
| bcc | centre-to-corner bonds | the planes x±y, y±z, z±x = n·s: the Delaunay tetrahedra (tetragonal disphenoids) of the BCC points | the same complex, edges and faces; not duals |
| fcc | corner-to-face-centre bonds (`octet`: all nearest bonds) | the four {111} families: the Delaunay tetrahedra and octahedra of the *other* face-centred lattice of the cell (centres and edge midpoints) | neither; `foam.fcc`'s dual is `beam.fluorite`, and `octet`'s is the rhombic-dodecahedral foam |

The duals of `beam.bcc` and `beam.octet` are the Voronoi foams, Kelvin cells (truncated octahedra) and rhombic
dodecahedra, which `foam` lacked. Checked numerically: which struts lie in which walls, which struts lie on
Voronoi edges, and the foams against brute-force Voronoi walls.

**Decisions.**

- Every lattice that joins parts with a boolean takes it as a callback, `combine`. The folded forms stay the
  default and fast; a mirror fold joins struts sharply where they meet their own images, so a `combine` form
  writes the struts or walls out explicitly to blend every junction (3 to 4 times the generated operations of
  the folded form, opt-in).
- `lattice.bcc(cellSize, beamRadius, combine)` is `beam.bcc(cellSize, combine) - beamRadius`; it is deprecated for
  a release (its body unchanged, so old documents render the same) and `gsl fix -stdlib` rewrites it.
- Each beam and foam doc names its complex and its dual.
- The missing duals are added: a Kelvin foam (dual of `beam.bcc`) and a rhombic-dodecahedral foam (dual of
  `beam.octet`). The rhombic-dodecahedral strut lattice already existed as `beam.fluorite`, the dual of
  `foam.fcc` (which sits on the cell's other face-centred lattice, the centres and edge midpoints). For cubic,
  `foam.cubic` holds `beam.cubic`'s struts and `foam.cubicGrid` is its dual. They are built from points: each sublattice's nearest
  point, an `abs` fold, a few planes, and `max` over the sublattices (no select, which would hurt fidget's
  interval bounds). A prototype Kelvin foam generated 34 WGSL operations against `foam.bcc`'s 38.
- A fully generic generator (any point set, every neighbour, no fold) is written up separately in
  [lattice_generator_20261126](../lattice_generator_20261126/spec.md), not built here.

## 7. Decisions: Go SDK

**Options are `With…`.** [AGENTS.md](../../../AGENTS.md) says entry points take `With…` functional options.
Today:

| Package       | Not `With…`                                                                                                                                       | Two names for one option                                                    |
| :------------ | :------------------------------------------------------------------------------------------------------------------------------------------------ | :-------------------------------------------------------------------------- |
| `assembler`   | `Parser`, `SourcePaths`, `SourceManager`, `GeometryResolver`, `WarpResolver`, `FoldWarpFieldToZero`, `DisableMeshSlotFolding`, `DisableDeadcodeRemoval`, `ExtractUI`, `ExtractMeshPlan` |                                                                             |
| `backend/mesh` | `MeshOracles`, `VdbOracles`                                                                                                                      | `Progress`/`WithProgress`, `Subdivision`/`WithSubdivision`, `Bbox`/`WithBbox`, `Algorithm`/`WithAlgorithm`, `Time`/`WithTime` |
| `backend/tape` | `Bbox`, `MeshOracles`, `VdbOracles`, `LevelSets`                                                                                                 | `Progress`/`WithProgress`                                                   |
| `backend/wgsl` | `Standalone`                                                                                                                                     |                                                                             |

Decision: every option is named `With…`. A bare name stays for one release, marked `Deprecated`, then goes.
omega-core's own callers move in the same pull request.

**`RenderLit(prog, LitOptions)` is the entry point.** Three wrappers only fill in `LitOptions` from positional
arguments: `RenderLitSized`, `RenderLitSizedView` and `RenderLitSizedViewRadiance`. Decision: mark them
`Deprecated` for one release, then remove them; omega-core's callers (`gsl compile`, `eval/render_gsl.go`,
`tool/materialswatch`) move to `RenderLit` in the same pull request. `RenderLitLinear` and `RenderLitWithLinear`
stay: they return the linear image, not only the PNG.

**A CPU fidget render with colour.** The viewer's Fidget device draws a coloured frame in the browser:
fidget-wgpu, compiled to wasm, renders each return as voxels, merges the images keeping the nearest hit
tagged with its return, evaluates each return's colour tapes and its sixteen material tapes at its hits, adds
SSAO, and lights the result with the standard frame's own shading
([examples/viewer/render/fidget-gpu.js](../../../examples/viewer/render/fidget-gpu.js), over
[internal/fidget/internal/wgpu-gbuf](../../../internal/fidget/internal/wgpu-gbuf)). Decision: the same on the CPU,
as an option of the normal render. `LitOptions.Device` is `"trace"` (today's sphere tracer, the default) or
`"fidget"`: fidget's CPU voxel render per return, the same merge and appearance passes, then `RenderLit`'s Go
shading, which already matches the viewer. `gsl compile --device fidget` and `/api/render-lit` pass it.

`eval.Render` and `eval.RenderFramed` are fidget's own image render: a PNG and GLSL of the union of the returns,
with no colour or material, and a debug `log.Printf` of the codegen JSON. Nothing in omega-core outside `eval`
calls them. They are retired once the fidget device lands and no other repository is found calling them.

## 8. Decisions: The Language Spec

The generated description (`program/preamble`) is the language's source of truth; it is what `gsl skill`, the
Help panel and the documentation test read. [doc/gsl_lang_spec_current.md](../../../doc/gsl_lang_spec_current.md)
is hand-written and has drifted from it. What the description lacks moves into it, and the spec is cut down to
the compiler's internals (passes, lowering, protobufs), pointing at `gsl skill` for the language. `@values` stays
reserved and undocumented in `gsl skill`: it is designed in
[doc/gsl_lang_spec_future.md](../../../doc/gsl_lang_spec_future.md) but not built.

## 9. Open

- **Chamfer-blend list forms (found 2026-11-27).** Under a non-sharp blend, `intersectChamferRound` and
  `intersectChamferBlend` on a list differ from their two- and three-field forms: the list form blends every
  edge plane and the vertex planes, the field forms only the binding ones, so the same shapes round
  differently. Sharp blends agree. The new difference list forms inherit it. Not decided: which is right.

Settled on 2026-11-26: the names (N1 `domain`, N2 `multiplyScalar`, N10 `a, b`), the spinodoid's scale (section 6) and the
`RenderLit` wrappers (section 7).

## 10. omega-dm Inventory

Taken read-only from omega-dm at `784bf8937` (2026-11-26): its 101 tracked `.gsl` files and its Go imports of
omega-core. Its running list of copied functions was not found in the repository, so copies, if any, are not
covered here. Only names this track changes are listed.

**GSL** (all in `common/web/example`):

| Change | omega-dm uses | Follows by |
| :----- | :------------ | :--------- |
| N1 `map` package | 12 calls in 9 files: `triangleWave` ×3, `rotateZX` ×3, `checkerXY` ×3, `rotateXY`, `patternRectangularXY`, `patternPolarXY` | The `gsl fix` rewrite, run over its examples |
| N5 honeycomb names | 7 calls, all in `Honeycomb Lattice Gallery.gsl` | The `gsl fix` rewrite |
| N6 three `bcc`s | `beam.bcc` and `foam.bcc`, one gallery each; no `lattice.bcc` | Nothing, unless `beam.bcc` or `foam.bcc` changes |
| N3, N7 parameter names | `transform.rotate`, `map.rotateXY`, called positionally | Nothing |
| N9 centred `torus` | `torus` ×4 and its own `torus(t)` in `Violin.gsl` | Nothing: the new overload takes three arguments |
| R1, R2, matrices, spinodoids | No calls | Nothing |

`map.rotateZX`, which omega-dm calls three times, has no `///` doc, so it is missing from `gsl skill`; so are
`map.rotateYZ` and `map.eulerAngles`. The `map` rename should document them, or say why not.

**Go** (imports of `gradientcontrol.com/omega/core/...`):

| Change | omega-dm uses | Follows by |
| :----- | :------------ | :--------- |
| `With…` options | `assembler.SourcePaths` ×13, `SourceManager` ×5, `ExtractMeshPlan` ×5, `GeometryResolver` ×4, `DisableDeadcodeRemoval` ×2, `ExtractUI` ×1; `mesh.Subdivision`, `mesh.Bbox`, `mesh.Algorithm` ×1 each | The deprecated aliases keep it building for a release; it moves within that release |
| Retire `eval.Render` | One call, `agent/fidget/internal/agent/render.go` | Moves to `RenderLit` with the fidget device before `Render` goes |
| R4 preamble `Domain` | Passes domains: `GSLAuthorSkill(chatCtx.GSLDomain...)`, and `LibrariesDescriptiveText(domain...)` twice | R4 keeps the variadic parameters, ignored and marked `Deprecated`, or omega-dm drops the arguments first |
| Kept exports | `SourceLibraryDescriptiveText` ×2, `CommonPitfalls`, `PrimitiveAndBuiltinDescriptiveText`, `ReservedWordsDescriptiveText`, `GSLDescription` | Nothing; R4 keeps them |

omega's chat builds its system message from `preamble.GSLAuthorSkill(...)`, every library in full (about 550 KB),
then appends `CommonPitfalls`. It does not use the tiered `gsl skill` index (#175).

## 11. Not in This Track

- The documentation and error fixes: #172, #175, #174.
- Lists: a vector added to a list of vectors, and `min`/`max` over a single list. A compiler change, its own
  pull request.
