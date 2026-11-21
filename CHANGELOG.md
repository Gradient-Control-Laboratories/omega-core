# Changelog

All notable changes to `omega-core` are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `map.bendXZ(v, curvature)` bends space in the XZ plane for `remap`: the plane x = 0 becomes an arc of the
  given curvature, 1 / radius, with lengths along it kept, for a bow, a camber or a warp, as in
  `remap(plate, map.bendXZ(p - origin, 1 / 400) + origin)`. A curvature of 0 is no bend.
- `gsl export part.gsl [more.gsl ...] -o dir` writes the WebGPU viewer and each file's compiled field as a folder
  of static web pages, which any web server can host: the page draws what `gsl viewer` draws, with its sliders,
  gizmos, time, sections and render settings, and compiles nothing, so the source is shown read-only and Save As,
  Imports, Export, Materials and Help are not offered. Several files share the one page and are picked from its
  menu. The folder is about 12 MB, without the material catalog's previews.
- `intersectChamferNormalized`, `unionChamferNormalized` and `differenceChamferNormalized` in builtin:
  chamfers by depth, the bevel the plane `depth` in from the edge, square to the bisector of the angle
  between the faces, with a field of unit gradient at an edge of any angle: the sum of the two fields
  divided by the length of the sum of their normals, offset by the depth. `intersectChamfer`'s bevel is the
  same plane, but its field is divided by sqrt2, which is right only at a square edge (at a 120-degree
  edge it changes 1.22 times as fast as distance and at 150 degrees 1.37), and its parameter is added
  before dividing, so it is a setback along each face only at a square edge. Pass the normals, or omit
  them to take each field's gradient. `intersectChamfer` and its callers are unchanged.
- `chamferDepthFromSetback(setback, normal1, normal2)` and `chamferDepthFromWidth(width, normal1, normal2)`
  in builtin: the depth of a chamfer set back `setback` along each face, s·cos(φ/2), or `width` across,
  (w/2)·cot(φ/2), with φ the angle between the faces through the material, taken from the faces' normals.
- `intersectChamferSetbacks(d1, d2, setback1, setback2, normal1, normal2)` in builtin, with
  `unionChamferSetbacks` and `differenceChamferSetbacks`: a bevel that meets the first face `setback1` from
  the edge and the second `setback2` from it, with a field of unit gradient. Equal setbacks give the
  Normalized chamfer at `chamferDepthFromSetback`'s depth.
- The viewer's Export panel has a `Web page (.zip)` format: it downloads the same folder for the open document,
  as the editor shows it, zipped (`POST /api/export`).
- `mechanical/shaft`: generalized shafts as a stack of segments along +Z, each placed by its two planes:
  `shaft.diameter`, `taper`, `section` (any 2D field), `hex`, `dFlat`, `doubleD`, `spline` and `gear` (spur or
  helical). `shaft.body(segments)` joins them with no surface inside the solid at any joint, whatever the
  convexity on either side. `shaft.body(segments, bottomHole, topHole)` cuts a `drill` hole stack into each end
  face. Gears and splines can name their ends: `shaft.runout(cutterDia)` runs the gaps on into the neighbour
  along the cutter's arc, and `shaft.chamferEnd(size)` chamfers the tips. `shaft.thread(major, pitch, from,
  to)` cuts an ISO 68-1 thread (`thread.male`) that engages `drill.tap` of the same size; its ends can be
  chamfered or run out as a lathe tool retracts, `shaft.threadRunout(length)`. A joint is blended by an
  entry between its two segments, `shaft.fillet(r)` or `shaft.chamfer(c)`: exact at a square shoulder
  between round segments, and free on a shaft that has none. Cranks: `shaft.eccentric(dia, offset, angle,
  from, to)` is a crank pin or an eccentric journal, and `shaft.web(mainR, pinR, offset, angle, from, to)`
  the exact hull of a circle about the axis and one about the pin.
- `mechanical/bearing`: deep-groove ball bearings (series 60, 62, 63, bores 10 to 50) by designation, as
  `bearing.deepGroove(6205.0, at)`, rings and balls, or `bearing.deepGrooveEnvelope`, the boundary dimensions
  alone; their seat dimensions (`shoulderDia`, `filletMax`); ISO 4379 plain and flanged bushes by series; and
  DIN 471 retaining-ring grooves. Every table cites its sources. Fits are nominal.
- `shaft.body` declares its `.box` (its end faces and its segments' farthest reach), so the renderer skips
  everything outside it; set `.box` on the result to override it. `shaft.envelope(segments[, bottomHole,
  topHole])` is the simplified model: gears at their tip circles, splines and threads at their major
  diameters, no runouts, for proofs, meshing and assemblies.
- A `reduce` over a list of functions can take the element before this one as a fourth callback parameter,
  `reduce((acc, f, i, prev) { ... }, fs, seed)`; the first element is handed itself.
  along the cutter's arc, and `shaft.chamferEnd(size)` chamfers the tips.
- `@caption("text", startTime, endTime [, fade])` is a top-level directive that shows a line of text centred at
  the bottom of the viewer while the animation clock is between the two times, in seconds. It fades in after
  `startTime` and out before `endTime` over `fade` seconds (0.5 by default, at most half the interval; `0` cuts).
  Captions whose times overlap stack. They reach the compile response as `captions`, so a page `gsl export`
  writes shows them too. Only the primary package's captions show.
- `@duration(seconds)` is a top-level directive, written with `@scene` at the top of the first file, that
  gives the length of the document's animation. It is to time what `@scene` is to space: the viewer's Time
  slider and end time span it, and the animation clock starts over on passing it, so an animation and its
  captions loop. It reaches the compile response as `duration`, so a page `gsl export` writes loops too.

### Changed

- The fast chamfers' parameter is named `offset`, not `setback`, in builtin's `intersectChamfer`,
  `unionChamfer`, `differenceChamfer`, `intersectChamferEdges`, `intersectChamferVertices`, their union
  counterparts, and the `ChamferRound` and `ChamferBlend` families (`[edgeOffset, vertexOffset]` for the
  list forms). It is added to the two fields before dividing by sqrt2, so it is a setback along each face
  only at a square edge; `intersectChamferSetbacks` and `chamferDepthFromSetback` take true setbacks.
  Arguments are positional, so callers are unchanged.
- The viewer reaches its server through one module, `examples/viewer/host.js`; `host-static.js` answers the same
  calls from files beside the page, and takes its place in a page `gsl export` writes.

### Removed

- `sweepLoft`, `sweepLoftScale`, `sweepLoftScaleSmooth`, `sweepLoftScaleSmoothSeparateCaps` and
  `bezierTubeDir` from builtin. They lofted two sections along a 3D quadratic Bezier, but mixed the sections'
  distances by the curve's parameter and scaled by it, so the field was not a distance bound, and they carried no
  frame, box or material. A document that needs them can keep a copy in its own library.

### Fixed

- `gears.gearProfile` chamfers its tooth tips with `intersectChamferNormalized`, given the flank's and the
  tip circle's normals at the edge, in place of `intersectChamfer`. The flank meets the tip circle at an
  obtuse edge, where `intersectChamfer`'s bevel, scaled for a right angle, has a gradient above 1; the
  normalized bevel has unit gradient there. The tip chamfer is now `kSmall` deep, measured from the edge to
  the bevel, about 1.7 to 1.8 times as deep as before at these edges (116 to 129 degrees through the
  material); the body, standoff and bore chamfers are unchanged. Gear and bevel-gear WGSL grows by
  582 bytes.

## [0.1.3] - 2026-11-17

### Added

- Surfaced non-blocking `SeverityWarning` diagnostics on `compile.Response` (`warnings`)
  across `gsl serve`'s `/api/compile` and `gsl compile -t field`.
  `gsl viewer` (`examples/viewer/`) now displays inline amber warning underlines, gutter
  highlights, and a code-panel warning banner when the code editor panel is open.
- Added `Diagnostics []program.Diagnostic` to `parser.Result` (preserved by `parser.CachingParser`)
  and propagated parser diagnostics into `assembler.Result.Diagnostics`.
- Emitted a `program.SeverityWarning` diagnostic in `parser/internal/parser2` when legacy
  `// Scene Size:` comments are parsed, advising `@scene(...)` and `gsl fix`.

### Changed

- `gsl eval`'s `volume` stops after 10 seconds unless the request names a longer `timeLimit` (Python
  `Document.volume(time_limit=...)`), and answers with the bracket it has, `"converged": false` and the new reason
  `time`. On a part whose intervals are loose, as a round blend's are, it ran to its evaluation budget for about
  100 seconds. The bracket and the budget are unchanged.

### Fixed

- A field's gradient (`.dp`) did not compile where the field is a `reduce` whose step reads its index, as
  `profileXY` and every extrusion of one do (`undefined identifier: index`). The gradient's evaluation of a
  `reduce` now hands the step its index as lowering does, and an index that reads such a name is evaluated rather
  than copied into the output.

- `gsl eval`'s `bounds` answered `"proven": true` with a box several millimetres off the part where the search
  ended on boxes it could not rule out (an angle folded with `atan`, across its seam). Such a box is now the
  grid's, with `"proven": false` and the new reason `loose`; a proven box stands within 1/64 of the part's span
  of points known to be inside, however its search ended.
- A slider read by a field that a `remap` re-derives inside a function's body (as `unionRoundExact` does with
  its closures) was emitted as a name the WGSL did not define, so `gsl eval`'s `shader` refused the model in
  `glsl-es300` and `gsl compile -t wgsl` wrote a module that does not validate. The helper now takes the slider
  as a parameter.

## [0.1.2] - 2026-11-15

### Added

- Added the `shader` op to `gsl eval` (`eval.Evaluator.Shader`, Python `Document.shader`): a
  document's field as a library a host's own shader calls, in `wgsl` for WebGPU or `glsl-es300`
  for WebGL2. It declares one uniform block for time and the sliders, with each member's byte
  offset in the response, and defines `gsl_field(p)` and `gsl_march(...)`, the front face of the
  viewer's primary march, with no entry point. Textures the field reads (the noise tile, a mesh's
  distance grid) are written as raw texels under `outDir`. For a document with parts the march
  guards them only when its `detail` argument names a rung. Both targets are one WGSL module
  (`backend/wgsl.Host`); the GLSL is its transpile.
- Added the `mesh` op to `gsl eval` (`eval.Evaluator.Mesh`, Python `Document.mesh`): the document,
  with `defs`, as one indexed `.glb` with a node per return and a normal per vertex from the
  field's gradient. The response gives the box meshed and, per body, its vertex and triangle
  counts, open and non-manifold edge counts and enclosed volume.
- `gsl eval`'s `volume` op takes `defs` (`eval.Evaluator.VolumeOf`, Python
  `Document.volume(defs=...)`), so the bracket is of the part a slider was moved to.
- Added `mesh.GLBNormals`, an option of `mesh.WriteGLB` that writes a normal per vertex, and
  `mesh.MeasureIndexed`, the edge counts and signed volume of an indexed mesh.
- The `render` Rust library links naga 30 (`wgsl-in`, `glsl-out`), already resolved in
  `Cargo.lock` for the vendored `fidget-wgpu`, for the `glsl-es300` target
  (`fidget.GLSLES300`). It adds about 1.9 MB.
- Added `gsl compile -t field` (`compile -t field`, default `<source>.field.json`) and shared
  `internal/fieldbundle` compilation so CLI field bundles and `gsl serve`'s `/api/compile` emit
  identical WGSL, shape, part, slider, UI control, `@scene`, mesh preview grid, audio, and
  structured diagnostic JSON.
- Added generalized left-to-right postfix modifier chaining (`Modifiers []*Modifier`) in
  `parser/internal/parser2`, supporting valid `.member`, `.dp`/`.dtime`/`.d(var)`, and
  `[index]` sequences (such as `m[0][1]`, `pts[0].x`, `body.box[0]`, and `v.d(t)[0]`) while
  rejecting invalid modifier sequences at parse time.
- Added `noise.cellularWalls(size)`, the exact distance to the nearest Voronoi wall
  (https://iquilezles.org/articles/voronoilines/), and the builtins it reads, `worleyWall` and
  `worleyWallFeature`. Unlike `cellularEdges` (F2 − F1, about twice the distance beside a wall), it
  is 1-Lipschitz, so a threshold gives walls of one thickness. Native in the cellular oracle and
  the WGSL emitter; shaders that do not call it are unchanged. `natural.lava`'s cracks,
  `ceramic.crackle`'s lines and the cracked-mud example now use it, so their cracks are one width
  (lava's stand-in coverage is 16.3%, crackle's line share 0.59, re-measured), and the cellular
  noise example shows it beside F2 − F1.
- Added `gsl viewer [part.gsl]` (`compile viewer`, in `cmd/serve/viewer`), which serves the
  WebGPU viewer on a file's folder, or on the examples without one, and opens the default browser
  on it (`?file=` in the viewer picks the file it starts on). The omega-core wheel now ships the
  viewer (`omega_core/viewer`) and the examples (`omega_core/examples`) for it.
- Added `gsl skill` (`compile skill`), Python `omega_core.skill(...)`, and `preamble` package-index
  helpers (`GSLSkillIndex`, `PackageIndexText`, `PackageDescriptiveText`, `PackageSummary`, and
  `PackageNames`) to emit GSL language and per-package reference docs directly from the embedded
  standard library (`stdlib/`), and served `/docs/gsl_description.md` and `/docs/gsl_libraries.md`
  dynamically in `cmd/serve`.
- Added canonical pure-Go oracle structs `meshasset.MeshSdfOracle` and `meshasset.VdbOracle` in
  `program/meshasset/oracle.go`, plus `spinodoid.Baked` and `spinodoid.Oracle` in
  `program/noise/spinodoid/oracle.go`.
- Exported `eval.RenderLitWithLinear` and `eval.LitEnvironment` in `eval/litview.go`.
- Separated pure-Go `eval` result types and helpers (`BoundsResult`, `CertifyResult`,
  `ExtremumResult`, `SurfaceSample`, `VolumeBounds`, `LitView`, `LitOptions`, `SliderRanges`,
  `TopLevelNames`, `TopLevelReturnExprs`, `ComposeSheet`, `ParseView`, `ParseViews`,
  `ViewPresetNames`, and `ParseVec3`) plus `eval/joints` from `//go:build cgo` files so pure-Go
  binaries (`CGO_ENABLED=0`, without linking the native Rust/C++ Fidget library) can use them.

### Changed

- `gsl eval`'s `bounds` op answers with a box proven by interval arithmetic to hold the part
  (`eval.Evaluator.Enclose`, `"proven": true`), where it used to fit one on a grid that missed walls
  thinner than a cell and could fall short of the surface. `cellSize` is then how far the box can
  stand off the part. Where a field's intervals prove nothing (a field divided by its own gradient,
  as `normalizeField` of a TPMS) it answers with the grid's box as before, `"proven": false` and a
  `reason`. `eval.Evaluator.Bounds`, the mesher's box fit and `gsl compile --bbox auto` are unchanged.
- `intersectRound(fields, radius)`, which every `box` is built on, writes its sharp term
  `min(sharp, 1e-30 - radius)` where it wrote `sharp - max(sharp + radius, 0)`. Values are the same,
  and so are gradients at radius 0. Interval arithmetic could not cancel the two `sharp`s, so every
  bound outside a box was off by the box's own width: `bounds` of an 80 mm plate is now proven in
  about 13 thousand evaluations where 600 thousand left it half a millimetre loose, and `volume` at
  the default tolerance takes a quarter of the evaluations. At a radius above 0 the gradient is now
  zero on the level set `sharp = -radius`, inside the part.
- `gsl eval`'s `eval` op defaults `exprs` to the document's returns, as `expr` does for the other ops.
- Updated `cmd/serve/internal/compiler` to assemble GSL sources in a single pass using
  `assembler.Result.AudioReturns` and `assembler/variation` instead of splitting source strings
  and assembling twice, preserving sliders, UI controls, and scene metadata on audio-only models.
- Enforced `go run ./tool/analyze-deps -rules` in `tool/check.sh`, decoupled
  `internal/assembler/topofold` from `internal/eval` and `internal/meshfold`, removed unused
  `internal/pkg`, and narrowed `internal/fidget/environment` to `eval/internal/environment`.
- Restricted `.color` validation in `internal/assembler/propval` and GSL specs to `[r, g, b]`,
  along with a `.material` transparency hint when a 4-element RGBA vector is supplied.
- Updated `mesh.MeshOracles`, `mesh.VdbOracles`, and `mesh.BakeSpinodoids` in `backend/mesh/emit.go`
  along with `tape.MeshOracles`, `tape.VdbOracles`, `tape.GenerateTapeWithSpinodoids`,
  `tape.GenerateTapeWithAppearanceAndSpinodoids`, `tape.StampSpinodoids`, and `tape.EmbedLevelSets`
  in `backend/tape/emit.go` to reference the canonical `meshasset` and `spinodoid` oracle types
  (compatible with existing callers via Go type aliases).

### Deprecated

- Deprecated redundant oracle type aliases in `backend/mesh` and `backend/tape` in favor of single
  canonical definitions in `program/meshasset` and `program/noise/spinodoid`; these aliases are
  scheduled for removal in `0.2.0`:
| Deprecated Alias       | Canonical Type            | Canonical Package Import Path                            |
| :--------------------- | :------------------------ | :------------------------------------------------------- |
| `mesh.MeshSdfOracle`   | `meshasset.MeshSdfOracle` | `gradientcontrol.com/omega/core/program/meshasset`       |
| `tape.MeshSdfOracle`   | `meshasset.MeshSdfOracle` | `gradientcontrol.com/omega/core/program/meshasset`       |
| `mesh.VdbOracle`       | `meshasset.VdbOracle`     | `gradientcontrol.com/omega/core/program/meshasset`       |
| `tape.VdbOracle`       | `meshasset.VdbOracle`     | `gradientcontrol.com/omega/core/program/meshasset`       |
| `mesh.BakedSpinodoid`  | `spinodoid.Baked`         | `gradientcontrol.com/omega/core/program/noise/spinodoid` |
| `tape.SpinodoidOracle` | `spinodoid.Oracle`        | `gradientcontrol.com/omega/core/program/noise/spinodoid` |

### Removed

- Removed `audio.ExtractAudio` from `backend/audio` and deleted the string-based
  `backend/audio/internal/extract` splitter package.

### Fixed

- `shell(d, thickness, bias)` moves the wall as documented: -1 inward, +1 outward. It ran the other
  way. A model that passed a bias gets the opposite wall.
- `gsl eval`'s `volume` op says whether its bracket met the tolerance (`converged`) and, where the
  evaluation budget or the subdivision depth ended it first, which (`reason`). It reported neither.
- The lattice library's comments no longer call TPMS fields unit-normalized. A gyroid's gradient is
  about 1.5 at its surface, so a sheet asked to be 1.0 mm is about 0.64 mm unless the field is
  passed through `normalizeField` first.
- Fixed `topofold` constant evaluation for `uiTransform`-derived expressions using column
  indexing and `vector.multiplyScalar` (e.g. scaled part transforms passed to
  `facePointNormal`).
- Fixed Manifold Dual Contouring (`fidget-mesh`) corner chamfering and octree over-collapse when an
  edge-crossing midpoint lands on an exact `min`/`max` constant tie (`grad == 0`, #38) by evaluating
  the gradient again a hair off the crossing, and marking leaves that still lack a plane normal as
  non-collapsible (`QEF_ERR_INVALID`).
- Fixed `volume` queries without a tolerance (`gsl eval`'s `relativeTolerance`, Python's
  `Document.volume` default) failing with "volume tolerance must be between zero and
  one": `eval.Volume` now reads a tolerance `<= 0` as the default `0.05`, as `surface` and `extremum`
  do.
- Fixed queries that default to the union of a document's returns (`gsl eval` with no `expr`:
  `volume`, `bounds`, `extremum`, `certify`), and `gsl compile --bbox auto`, failing with "failed to
  assemble query" when a comment stood above `return`, or `return` was the first statement after
  `@scene`: `eval.TopLevelReturnExprs` now starts at the `return` keyword (#46).
- Fixed a ternary whose condition varies over space being NaN when the branch it does not take is
  infinite or NaN there (`(p.x > 0) ? p.x : 1 / z`). This affected eval, the meshers, `volume` and
  `certify`, but not the WGSL viewer, and made `unionRoundExact` NaN on the line through both
  shapes' centres. Each branch is now masked with Fidget's `and` instead of being multiplied by its
  0/1 condition (#51).

## [0.1.1] - 2026-11-11

### Added

- Added `gsl eval` (`compile eval`) JSON query CLI (`eval`, `extremum`, `certify`, `bounds`,
  `surface`, `volume`, `describe`, and `render` operations, plus `--stdio` persistent sessions),
  `gsl render` shorthand for `compile -t png`, Python `omega_core.Session` / `omega_core.Document`
  query client (and reference documentation in `python/omega-core/README.md`), and
  `(*eval.Evaluator).RenderViews`.
- Added `STLStats.SignedVolume` in `backend/mesh/measure.go` so `mesh.MeasureClosedSTL` rejects
  inside-out meshes.

### Changed

- Changed `compile -t stl` defaults to `--subdivision 9`, `--algo mdc`, and the `@scene` box
  (`--bbox auto` when no `@scene` is declared), matching the viewer, and bumped default viewer
  export Detail to `9`.

### Fixed

- Fixed marching-cubes triangle winding (`marching_cubes.rs`) so exported meshes and normals face
  outward, and wired the standalone viewer's `Server DC` / `Server MC` export options to `mdc` and
  `marching-cubes` respectively.
- Fixed `compile --bbox auto` clipping thin or tilted geometry by iteratively widening and
  rescanning the search region until the fitted box stops reaching the search boundary.
- Fixed degenerate `0/0` NaN gradients at `sqrt(0)` and NaN QEF vertices from zero-gradient cells
  in Fidget (`fidget-core`, `fidget-jit`, and `fidget-mesh`), keeping shelled exact boxes closed.
- Fixed `eval` query `defs` overrides to match declarations preceded by `@scene(...)` or
  `@range(...)` annotations (replacing slider range annotations so queries can test out-of-range
  values) and ignore `.prop = ...` property statements.

## [0.1.0] - 2026-11-11

- Initial standalone release of `omega-core` (GSL parser, assembler, standard library, evaluator,
  backends, CLI tools, and `omega-core` Python wheel).

[Unreleased]: https://github.com/Gradient-Control-Laboratories/omega-core/compare/v0.1.3...HEAD
[0.1.3]: https://github.com/Gradient-Control-Laboratories/omega-core/compare/v0.1.2...v0.1.3
[0.1.2]: https://github.com/Gradient-Control-Laboratories/omega-core/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/Gradient-Control-Laboratories/omega-core/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/Gradient-Control-Laboratories/omega-core/releases/tag/v0.1.0
