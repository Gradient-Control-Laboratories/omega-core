# Changelog

All notable changes to `omega-core` are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
where `### Breaking` lists changes that require updating existing code to the new version,
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html)
(breaking changes increment the minor version while pre-`1.0.0` (`0.y.z`) and the major version
from `1.0.0` onward).

## [Unreleased]

### Breaking

- **The `map` package is renamed `domain`**: `map.rotateXY(...)` is now `domain.rotateXY(...)` and
  `import "map";` is `import "domain";`, so the package no longer shares its name with the
  `map(fn, list)` primitive. A document that still says `map.` gets an error naming `domain`, and
  `gsl fix -domain` rewrites the references and imports. `rotateYZ`, `rotateZX` and `eulerAngles`
  are now documented.
- **Honeycomb lattices drop the repeated name**: `honeycomb.squareHoneycomb`, `squareDiagonalHoneycomb` and
  `squareHoneycombReinforced` are `rectangular`, `rectangularDiagonal` and `rectangularReinforced`;
  `hexagonalHoneycomb`, `triangularHoneycomb` and `octetHoneycomb` are `hexagonal`, `triangular` and
  `octet`. `gsl fix -stdlib` rewrites them.
- **`matrix.inverse` inverts any square matrix**: it was the affine 4×4 inverse, now `matrix.inverseAffine`.
  An affine matrix gets the same result from either; `inverseAffine` is cheaper.
- **Graded spinodoids take a `box`**: `gaborGraded(..., box)` and the other graded forms end with `box`, the
  [min, max] corners, in the field's own coordinates, of the region the field is evaluated over; the table is
  baked there and nowhere else. The scene and imported-geometry bounds and the periodic fallback are no
  longer used for spinodoid tables, and the spinodoid builtins take `box` as a 13th argument.
- **`list` is a builtin**, so a document can no longer declare its own `list` function or variable.
- **Changed Go module path to `github.com/Gradient-Control-Laboratories/omega-core`**: The root
  `go.mod` module declaration and all internal package import paths changed from
  `gradientcontrol.com/omega/core` to `github.com/Gradient-Control-Laboratories/omega-core` so
  downstream Go modules can resolve and vendor `omega-core` directly from GitHub via `go get` and
  `go mod vendor` without local `replace` directives.
- **Exact rounds take field values and a radius (`intersectRoundExact`, `unionRoundExact`)**: The
  two-argument forms and nullary-function (thunk) arguments are removed, and the one-step forms
  `(field1, field2, normal1, normal2)` are now `(field1, field2, radius, normal1, normal2)` taking
  plain field values. Pass `radius = 0` to keep the previous sharp exact union/intersection
  (`unionRoundExact(a, b, 0, a.dp, b.dp)` replaces `unionRoundExact(() { return a; }, () { return b; })`).
  `gsl fix -stdlib` rewrites 2-argument and 4-argument thunk calls to value-based calls with
  `radius = 0`.
- **Removed `sweepLoft` and `bezierTubeDir` from `builtin`**: `sweepLoft`, `sweepLoftScale`,
  `sweepLoftScaleSmooth`, `sweepLoftScaleSmoothSeparateCaps` and `bezierTubeDir` were removed from
  `builtin`. They lofted two sections along a 3D quadratic Bezier, but mixed the sections' distances
  by the curve's parameter and scaled by it, so the field was not a distance bound, and they carried
  no frame, box or material. Existing documents calling them will fail to compile. For constant-section
  planar sweeps (including quadratic Bezier paths), migrate to `sweep`, `sweepOpen`, or
  `sweepOpenEdges` with `edgeBezier`. There is no direct 3D lofting replacement, so `gsl fix -stdlib`
  cannot rewrite them automatically and instead prints manual migration suggestions at each call site;
  documents that need the old 3D lofting behavior can keep a copy in their own library.
- **Removed `.min`, `.max` and `.period` slider properties (replaced by `@range`)**: A document that
  sets `.min`, `.max` or `.period` no longer compiles: the error is at its first such statement and
  names `@range(min, max[, period])` and `gsl fix -range`, which rewrites the document. The parser
  still reads them so the fixer can. The language description served by `gsl skill` and the viewer's
  Help panel teaches `@range` and notes that the properties were removed.

### Added

- **`list(n)` and `list(n, fn)`**: the list `[0, 1, ..., n - 1]`, or `[fn(0), ..., fn(n - 1)]`, with n known when
  the document compiles; a function whose parameter sets a list's length is specialized for each constant
  its calls pass.
- **Matrices of any size**: `matrix.transpose`, `identity(n)`, `determinant` and `inverse` take any size
  (`determinant` and `inverse` pivot, so a zero on the diagonal works), and a matrix that varies at run
  time unrolls them. The sized names (`transpose2/3/4`, `identity3/4`, `determinant2/3/4`, `inverse2/3/4`)
  and `matrix.scale` (now `multiplyScalar`) are deprecated; `gsl fix -stdlib` rewrites them.
- **Lattices take a boolean**: every beam lattice that joins struts has a `combine` form that blends every
  junction. `foam.kelvin` and `foam.rhombicDodecahedral` are the Voronoi foams, the duals of `beam.bcc`
  and `beam.octet`, and each lattice's docs name its dual. `lattice.bcc` is deprecated for
  `beam.bcc(cellSize, combine) - beamRadius`, which `gsl fix -stdlib` writes.
- **Go options are all `With…`**: `assembler.WithSourcePaths`, `WithParser`, `WithExtractUI` and the rest,
  `mesh.WithMeshOracles`, `tape.WithBbox`, `wgsl.WithStandalone`; the bare names are deprecated for a
  release. `eval.RenderLitSized`, `RenderLitSizedView` and `RenderLitSizedViewRadiance` are deprecated for
  `RenderLit(prog, view.Options(width, height, extent, time))`.
- **Shapes and booleans**: `torus(center, ...)` and `octahedron(center, size)`; every difference takes a list
  `[target, tool, ...]` and has an operator form, as unions and intersections do.

- Migrated GSL compiler architecture, language specification/migration, golden verification, and
  stdlib tolerance-table documentation from `omega-release/doc/` into `doc/`, synced
  `doc/gsl_lang_spec_future.md`, extracted `doc/roadmap.md` from `README.md`, and documented Go
  SDK consumption (`go get` + `tool/ensure-native-libs.sh`) in `README.md`.
- Emitted a `program.SeverityWarning` diagnostic in `parser/internal/parser2` when unnecessary
  `import . "omega"` statements are parsed, advising that `builtin` is implicitly imported, and
  added the `omegaimport` fixer to `gsl fix` (`cmd/fix`) to remove them automatically.
- Shared consumer-reusable native library provisioning (`tool/ensure-native-libs.sh` and reusable
  GitHub Actions workflow `.github/workflows/build-native-libs.yml`): detects host OS/arch,
  downloads and SHA256-verifies prebuilt `librender` and `libmanifold_glue` release tarballs when
  available, and falls back to `tool/build-fidget.sh --no-prompt` when offline, uncached, pinned
  to an untagged commit, or invoked with `--force-build`.
- Stand-ins anywhere: `name.simple = proxy` now works on any named shape anywhere in an expression
  (including intersections and remaps), not only top-level parts, and measures switch size at the
  point drawn (`gsl_featurePixel`) rather than at the eye's distance to the part's box. When paired
  with `name.featureSize = size`, the stand-in switches at `size` (using `builtin.simplified`'s
  5-argument form) even on unbounded shapes or inside library helpers; `gears.gearProfile` and
  `thread.basicForm`/`sinForm` use `.simple` + `.featureSize` so gear teeth and threads switch to
  their pitch cylinder in a single pass.
- Feature pruning: `name.featureSize = size` (mm) marks a value as a geometric feature that the viewer drops
  once it is too small to see, even when its part is large: a hole closes, a bump goes. The size is measured
  against the pixel at the point, from the camera, so it holds inside any rigid placement, and the feature
  fades across the Detail slider's band, as a part does, and is not evaluated once gone. A size of 0 takes
  the smallest extent of the value's box. A feature may stand only in a union or as what a difference takes
  away (`pruneBelow(field, size)` in `builtin`). Gear teeth (`gears.gearProfile`, so every gear and
  `shaft.gear`) give way to the pitch cylinder; threads (`thread.male`, `female`, `maleSin`, `femaleSin`, so
  `shaft.thread`, the threaded fastener shanks and nuts, and `drill.tapThreaded`) give way to the cylinder at
  their pitch diameter, crests and roots each half a pitch; and `drill.hole` voids close, sized by their own
  field at the mouth. Drawn, a gear's and a thread's field is its profile's, as before. CPU evaluation,
  meshing and export are unchanged.
- Layered surfaces: a material's height now moves the shape that sets it, before the booleans that
  use it, by what the material adds over the one the shape inherits. Coats laid as new names over
  the shape they cover (`coat1 = part; coat1.material = coating.levelled(paint, partMat, build,
  underSize)`) stack on the part and on each other; a cut taken after them stays flat and shows
  each coat as a band over what it covers; paint laid over the cut covers it again. Lowered as
  `Program.DisplacedReturns` (`program.proto` field 14), which the viewer and the CPU lit renderer
  march; meshing and export keep the undisplaced returns. Example: `examples/coated_oak_stick.gsl`.
- `material.withBuild(m, thickness)`, `material.build(m)` and `material.relief(m)`: row 2's fourth
  component, reserved until now, is the part of a material's displacement that is film standing on
  the surface beneath, so a film's thickness and the marks on it are kept apart.
- `wood.prepWavelength(prep)`: how far apart a prepared surface's relief is (2 mm for a bandsaw),
  for a coat laid over the wood to take as `coating.levelled`'s `underSize`.
- `displacementScale()` builtin: how far the renderer is moving surfaces by their materials'
  heights where a distance is evaluated (the viewer's displacement knob in its march, 1 in the CPU
  lit renderer, 0 elsewhere).
- `gsl fix -stdlib`: migrates removed 2-argument and 4-argument thunk calls to `unionRoundExact` and
  `intersectRoundExact` into field-value calls with `radius = 0`, and prints manual migration
  suggestions for calls to removed `builtin` functions (`sweepLoft`, `sweepLoftScale`,
  `sweepLoftScaleSmooth`, `sweepLoftScaleSmoothSeparateCaps`, `bezierTubeDir`).
- `intersectChamferSetbacksSteps(field1, field2, setback1, setback2, steps)` and
  `intersectChamferAngleSteps(field1, field2, distance, angle, steps)` in builtin, with their union and
  difference counterparts: flat chamfers between faces that may be curved, the Steps tier of the chamfer family
  after `intersectChamfer` and the Normalized and Setbacks chamfers, which are exact only where both faces are
  straight in section. For each point, exact round's Newton steps find the nearest point of the edge and the
  faces' normals there; in the plane normal to the edge the chamfer is the straight line through the point of
  face 1 at the first distance from the edge and the point of face 2 at the second (each a straight-line
  distance, found by Newton steps on the face), or from face 1's point at the angle until it meets face 2. Flat
  on a ball as on a plane, as a boundary-representation chamfer is. A sixth argument, `seed`, starts the steps
  elsewhere than p, to select one edge where the faces meet in several. Where the steps miss the edge, or the
  frame degenerates far from it, there is no chamfer and the field stays finite. On brep-bench's 27 chamfer
  cases with a ball face, they meet the grade in 26 at about 30 times the time of a hand-placed cone.
- `intersectRoundExact(field1, field2, radius)` and `unionRoundExact(field1, field2, radius)` in builtin: exact
  rolling-ball rounds for curved faces. They find the seam's point nearest p by four Newton steps (each moves to
  both fields' tangent planes, then slides along the seam to the foot of p) and judge whether p is nearest the
  seam at the point they reach: three steps are right to 1e-6 mm where the seam is a circle, five where it is
  not, where the one-step form is off by up to 0.14 mm on a sphere and a plane rounded by 2.
  `intersectRoundExactSteps(field1, field2, radius, steps)` and `unionRoundExactSteps(...)` take the step count;
  steps not taken are dropped at compile time, so pass it as a constant. Each step evaluates both fields again,
  and fillet/ball-and-post meshes in about 17 to 25 times the CPU of the one-step form.
  `intersectRoundExact(field1, field2, field3, radius)`, `unionRoundExact(field1, field2, field3, radius)` and
  their `Steps` forms round three fields: each edge as the two-field form does, and the corner with the patch
  of the ball touching all three faces, found by Newton steps on the three fields at once; right to 1e-6 mm at
  a box's corner, a pyramid's and a ball cut by two planes. `intersectRoundExact(fields, radius)`,
  `unionRoundExact(fields, radius)` and their `Steps` forms round a list of fields: the largest of every pair's
  edge distance and every three's corner distance, which is exact where the intersection is convex, at any
  number of faces to a corner, among them the apex of fillet/rectangular-pyramid-all-edges, where four faces
  meet and no ball touches all four. `.dp` cannot follow a field through `map`'s functions, so the list forms
  take their gradients by central differences (`differenceGradient`); for three fields they cost about eight
  times the three-field forms, N fields need N(N-1)/2 edges and N(N-1)(N-2)/6 corners, and Fidget's JIT
  cannot build eight. A seam or corner point is accepted only where the fields vanish at it and the faces
  make an edge (normals more than about 0.6 degrees from parallel) or a corner (normals not near coplanar).
  `intersectRoundExact(radius)`, `intersectRoundExactSteps(radius, steps)` and their union duals bind the
  radius and return the operator, to hand a primitive that blends two fields, as `extrude` does.
  `edgeDistance`, `cornerDistance`, `intersectDistance` and their `Differenced` forms, `edgeProject`,
  `edgeNewtonStep`, `cornerNewtonStep` and the `...From` and `...Verdict` arithmetic they share are their
  parts. `examples/ball_and_post.gsl` shows them.
- Exact distances to conics and quadrics in builtin: `ellipseXY(center, radii)`, `parabolaXY(vertex, focalLength)`,
  `hyperbolaXY(center, semiAxes)`, `ellipsoid(center, radii)` and `paraboloid(vertex, focalLength)`, and under them
  `quadricDistance(q, m, b, c, steps)` for any quadric Σ m_i x_i² + 2 Σ b_i x_i + c = 0 in its own frame. The nearest
  point is x(t) = (q - t·b) / (1 + t·m) for one Lagrange multiplier t, the single root of a function that decreases
  strictly over a known interval for every conic and quadric alike; twelve Newton steps from t = 0, each kept within
  the interval, find it, and on an axis of symmetry, where the nearest points are a ring or a pair and the root is at
  the interval's end, the closed form there gives the distance. No step branches on the point. Against brute force
  they are within 2.6e-6 near the surface and 7.5e-6 of the distance far from it, on the axes too.
- A `draft` package: faces tapered from a hinge curve by an angle to a pull direction.
  `draft.horizontalCurve(horizontal, vertical, angle)` is the surface through a curve in a plane normal to the
  pull, `horizontal·cos(angle) + vertical·sin(angle)`, from the signed distance to the curve within planes
  normal to the pull and the height above the curve's plane; the section at height h is the profile offset
  inward by h·tan(angle), with unit gradient. `draft.point(center, pull, angle)` is the cone with its apex at a
  point. A profile with sharp corners is given as the max or min of its sides' half-planes, so that its
  corners stay sharp as draft keeps them.
- `nested(fields)` combines exact fields whose surfaces do not cross, each wound with its solid on the negative
  side (counterclockwise outlines and islands, clockwise holes), into one exact distance in any order: an outline
  with holes in a single profile, or a 3D solid with a void.
- Corners per vertex on `profileXY`: `profileXY(points, curvatures, corners)` and
  `profileXY(points, curvatures, corners, blend)`, with `cornerSharp()`, `cornerFillet(radius)`,
  `cornerChamfer(setback)`, `cornerBlend()` and `everyCorner(points, corner)`. Fillets and chamfers are edges of
  the loop, so the distance stays exact, and are clamped to fit; `cornerBlend()` applies any edge-blend operator
  (`intersectRound(r)`, `intersectChamfer(d)`, …) at that vertex. Corners between two straight edges.
- `revolve(profile, axisPoint, axisDirection, angle)`, a partial revolve with flat caps, exact; `.u` is the angle
  and `.v` the profile's `.u`.
- `sweep(section, path)` sweeps a 2D section along a closed planar path, any exact 2D field: exact while the
  section stays inside the path's tube, never more than the distance anywhere. `.u` runs along the path and `.v`
  round the section.
- `openProfileXY(points)` and `openProfileXY(points, curvatures)`, an open chain of lines and arcs whose ends carry
  on as rays along their tangents: the exact signed distance to that extended chain, positive to the right of
  travel. `.u` runs along the chain.
- `sweepOpen(section, points)` and `sweepOpen(section, points, curvatures)` sweep a section along an open chain,
  capped flat at both ends: exact beside the chain and over the caps; beyond a cap where another part of the sweep
  is nearer, the distance reads that cap's and is too large, though its sign stays right.
- Quadratic Bezier edges: `edgeLine()`, `edgeArc(curvature)` and `edgeBezier(control)` name each edge of
  `profileEdgesXY(points, edges)`, `openProfileEdgesXY(points, edges)` and
  `sweepOpenEdges(section, points, edges)`, so lines, arcs and Beziers mix in one profile, still exact;
  `.u` is the arc length along a curve. Each edge computes only its own kind's step on the GPU (by `choose`);
  for lines and arcs alone `profileXY` is still the faster form.

### Changed

- `coating.levelled` (and every coating laid over a base through it: `clearCoat`, `galvanized`, and
  the `powderCoat`, `wrinkleCoat`, `sandCoat` and `hammertone` overloads) adds the film's `build` to
  the base's build and levels only the base's relief, where it used to level the base's whole
  displacement and add no thickness. A coated surface is drawn `build` mm proud of the one beneath.
- Wood's finishes level a sawn surface's marks by their real spacing, `2 + 2 * noisiness` mm (2 mm
  for a bandsaw), where they used `1 + 3 * noisiness`; a topcoat over bandsawn wood keeps a little
  more of the saw's texture.
- A shape inside a return that sets a material with a height is displaced before the booleans that
  use it, where the return used to be displaced as a whole by its blended material: a face finish
  set on each face of a `max` moves that face, and a cut taken after a displaced material no longer
  moves the cut face. A return that is one shape with its own material, placed by moves or a
  stand-in, compiles exactly as before.
- `mechanical/shaft` segments evaluate their 2D section (gear profile, spline, thread, bevel teeth,
  taper, etc.) once and return all four capped/extended distances together as a `vec4` instead of
  being invoked four times per segment in `shaft.body` and `shaft.envelope`. This eliminates 4×
  section duplication at the source: `shaft_pinion` WGSL shrinks from 87 KB to 35 KB (−60%) and
  `shaft_hobbed` from 51 KB to 20 KB (−60%) with identical fields.
- `gears.gearProfile` now chamfers its tooth tips with `intersectChamferNormalized` (given the flank's
  and tip circle's normals at the edge) in place of `intersectChamfer`. The tip chamfer is now `kSmall`
  deep, measured from the edge to the bevel, about 1.7 to 1.8 times as deep as before at these obtuse
  edges (116 to 129 degrees through the material); the body, standoff and bore chamfers are unchanged.
  Callers compile without changes (no `gsl fix` is needed); callers needing the shallower tip chamfer
  can adjust `kSmall` on `gearProfile`. Gear and bevel-gear WGSL grows by 582 bytes.
- The box analysis folds `&&`, `||` and `!` of constants, so a select on them takes one branch: a shaft
  segment whose field branch has no guard of its own no longer leaves the shaft's box unbounded.
- `shaft.body` and `shaft.envelope` no longer bridge their first segment to nothing, which evaluated it
  twice for no effect: every shaft's WGSL is 1 to 22% smaller, its field unchanged.
- A colour or material built in a local with a stand-in or fine detail of its own (`t.simple`, `t.detail`)
  and set on a part inside a function (`s.color = t`) keeps them out of the function, as one returned from a
  function always did; before, the part's colour lost its stand-in and stayed textured with textures off.
- PNG renders frame a part too thin to be sampled inside beside a thick one (a threaded shaft's thread
  went missing without `--extent`).
- The fast chamfers' parameter is named `offset`, not `setback`, in builtin's `intersectChamfer`,
  `unionChamfer`, `differenceChamfer`, `intersectChamferEdges`, `intersectChamferVertices`, their union
  counterparts, and the `ChamferRound` and `ChamferBlend` families (`[edgeOffset, vertexOffset]` for the
  list forms). It is added to the two fields before dividing by sqrt2, so it is a setback along each face
  only at a square edge; `intersectChamferSetbacks` and `chamferDepthFromSetback` take true setbacks.
  Arguments are positional, so callers are unchanged.
- The viewer reaches its server through one module, `examples/viewer/host.js`; `host-static.js` answers the same
  calls from files beside the page, and takes its place in a page `gsl export` writes.
- A part's box declares what its corners repeat once, as a value before the part, instead of writing every copy
  out in the WGSL, where the viewer evaluates it per point to cull. The box is unchanged; parts whose boxes read
  profile bounds shrink most: `examples/profile_cylinder.gsl` from 190 KB of WGSL to 83 KB, about half its time
  per point on the GPU.
- The viewer's Render panel is regrouped. Resolution comes first, then At rest, Field of view, Depth of field,
  Render image and Frame rate, with two folded sections: Detail (the detail ladder, Pixel filter, and the Relief and
  Fine relief switches) and Advanced (Moving, Environment, Light rotation, and the Shadows, Occlusion and Back faces
  switches). A switch is a chip that wears the label of its state ("Shadows on", "Shadows off"). Occlusion, the loop
  Standard draws that marches the field out along each normal, has a chip of its own and no longer goes off with
  Shadows. The labels of the Render, Scene, Section and Export panels are as light as the Parameters panel's.
- The six `examples/` models that set slider limits with `.min`, `.max` and `.period` (blend, can, gyroid,
  hybrid_bracket, kinetic_sculpture, slider_mechanism) use `@range`, as `gsl fix -range` writes it. Their sliders and
  compiled shaders are unchanged. So are the example documents under `internal/testdata/example` and three codegen
  fixtures, whose WGSL goldens move only in their `gsl_param_<line>_<col>` names (every golden is identical with those
  names normalized), and the tests that used the properties only to declare a slider. Tests about the properties
  themselves keep them.
- The viewer's Materials panel writes `@range(min, max) <body><Param> = value;` for a preset's scalar parameters in place of
  the `.min`/`.max` lines, reads a variable's limits from its `@range` (or from `.min`/`.max` in an older document),
  and prunes either shape. A document already on `@range` used to have its limits read as the catalog's defaults the
  next time the panel adjusted it.
- `profileXY` fills its three distance pools in one pass instead of three, with the same values: 2.6 to 11 times
  faster on the GPU across the profile examples.
- A profile's box is exact: the union of its edges' own boxes, an arc's extremes where they lie on it and a
  Bezier's extremum along each axis, for `profileXY` with arcs, its corner forms (the loop's box without the
  corners, which holds every fillet) and `profileEdgesXY`; it was the points' box grown by the most an edge can
  bulge, 28 mm loose a side on an L with arcs. `sweepOpen` and `sweepOpenEdges` grow their chain's exact box.

### Fixed

- `topofold` now evaluates general UI-derived constant expressions via a scoped `ConstEvalFunc`
  fallback (`eval.ConstPkgSlotsAt`) in `assembler`, rejecting non-finite results, and Rust Cargo
  build scripts (`internal/fidget/internal/rust/build.rs`, `internal/fidget/internal/tape-core/build.rs`,
  `backend/mesh/manifold/internal/rust/build.rs`) read `CARGO_MANIFEST_DIR` at runtime via
  `std::env::var_os` instead of compile-time `env!`.
- `remap(field, [x, y, z])` with a constant point reads the field at that point in the WGSL too. The viewer
  drew it as the field moved by that point, while CPU evaluation and exports read the point, as the language
  defines it. Such a remap now folds to its value when the field reads nothing but p and constants.
- Viewer: a part whose relief has faded out on the Detail ladder no longer evaluates the displaced
  field when its program displaces per return; it marches its undisplaced sample, which is exact,
  since every height is scaled by the part's weight.
- Keyed `.github/workflows/check.yml` and `.github/workflows/wheels.yml` concurrency groups on
  `github.event.pull_request.number || github.ref` so `pull_request` `closed` events no longer
  collide with and cancel `push` runs on `main`.
- `--algo marching-cubes` meshes are watertight. Each cell worked out the crossing on each of its edges, and
  four of a cell's twelve edges run down their axis, so the cell across such an edge read it from the far end and
  rounded `t = -v0 / (v1 - v0)` differently in float32: the two cells' copies of one vertex differed in the last
  bit, and the mesh was open along the edge (at depth 9, 2,528 unmatched edges on a sphere, 3,128 on a box with
  every edge filleted). Every edge is now read up its axis, so a crossing depends on the grid edge alone. Triangle
  counts, area and volume are unchanged, and `mdc` output is byte-identical.
- Inlining was quadratic in a program's size: before inlining a local it counted every later use of it, over
  bodies whose calls were already inlined, where it needs only to know whether there are two. It now stops at
  the second. The eight-field exact round of fillet/octahedron-all-edges compiles in 8.9 s, from 35.9 s; the
  compiled output is unchanged.
- `.dp` through `map` and `reduce`: a fold over a list of fields whose function applies `.dp` to its element,
  directly or through a function it calls, failed with "the .dtime/.dp derivative operators are only supported
  in top-level expressions". The property pass, which expands `.dp`, runs before folds are unrolled, and it
  differentiates a parameter only at a call whose argument it can see. Such a fold over a list written at its
  site is now unrolled first, as a fold over functions already was, its lambda given what it reads of the
  function it is written in as parameters and moved to the file's own functions; nested folds, such as every
  pair or every three of a list, unroll in turn. A list passed to a function that folds it so, directly or
  through the functions it calls, is bound into that function's specialized clone, as a list of functions is.
- `.dp` through a remap of a function's parameter. A call binds a parameter to its argument's value at the
  caller's p, and a remap inside the function read that value unmoved, so `f(v) { return remap(v, q); }` lost
  the chain rule through q: the value was right and its gradient was not. The argument is now evaluated again
  where the remap reads it, and a parameter counts as varying with p when its argument does.
- A box declared in a function from a `reduce` whose step indexes a list by its `index`, such as
  `points[mod(index + 1, count)]`, now holds: read with `.box` it failed to link ("undefined identifier: index"),
  and in builtin the part it bounds lost its box without a word.
- A box declared in a function written in the document may read locals assigned after the shape it bounds:
  the crop that applies it was written on the shape's own line, so such a local was read before its assignment
  ("undefined identifier"). A cropped body is now put in an order its reads allow; one that compiled before is
  unchanged, and every reader of the shape still reads it cropped.
- A one-element list indexed at run time, such as `curvatures[0]` of `[0.1]` read in a reduce, failed on the CPU
  paths (`gsl eval`, tape and meshing) with "target must be a collection, got num": the fidget backend builds the
  list as one value. Index 0 now reads it there as WGSL already did; past 0 is an error.
- A declared `.box` that reduces over a list a function call returned failed to assemble ("vector subtraction
  dimension mismatch", "collection elements must be all scalars or all vectors"): the box pass handed the
  reduce the whole list as one element. `profileXY(points, curvatures)` with its curvatures made by a function
  (a `map` in a helper) failed this way once the profile was named.
- A declared `.box` that reads names assigned after its shape (`form = ...; low, high = ...; form.box = ...`)
  failed with "undefined identifier" wherever the function's body was inlined into its caller: any document
  function called with constants, or a library function handed a blend made from a slider. The shape's
  assignment now moves after the names its box reads; a shape read in between is refused with a message.
- A function that passes its function parameter on to another only inside a fold's callback
  (`f(blend) { return reduce(..., (best, q) { ... inner(q, blend) ... }) }`) failed to assemble with
  "undefined identifier" once called with a closure: the closure transform pruned the template it called but
  kept it, not seeing that its callback made it a template too.
- A three-way select on a constant vector condition folded to its positive branch whatever the condition's signs:
  `[-1, -2] ? a : b : c` compiled as `a`, where the shader picks `c`. It now folds per component, as the shader
  selects, and is pruned to one branch only when every component picks it.
- A `reduce` whose function takes the index, `(accumulator, element, index)`, now folds to a constant; it was
  left to run per point, so a box written with one did not fold to numbers.
- A constant that folds to infinity or NaN (such as division or `mod` by zero) is now a compile error naming
  it, instead of emitting `+Inf.0` or `NaN.0` into a WGSL shader the browser refuses.
- A bound of an explicit `@range` that reads another top-level variable works: `@range(abs(a - b) - offset, a + b - offset)`
  lost its slider, and its `.period` animation, whenever `offset` was a plain constant. The assembler's passes (link,
  closure, inline, constspec, types and dead-code removal) walked a variable's `.min`/`.max`/`.period` expressions but
  not the bounds written in `@range`, so `offset` was inlined away while the bound still named it and could no longer
  fold. They walk the bounds now. A range synthesized from `.min`/`.max`/`.period` is unchanged.
- `.box` treats a variable declared with `@range` as a slider, as it does one declared with `.min`/`.max`. It found
  sliders by their properties alone, so an `@range` variable read as the constant it starts as: `min(a, r).box` was
  bounded by `a` for an `@range` `r`, a box that is wrong once `r` moves, and unbounded for a `.min`/`.max` one.
- A slider set with `.min`/`.max`/`.period` keeps its slider when a bound reads a computed constant: with
  `hi = 10 + 2;`, `x.max = hi * 2;` lost the slider. The range synthesized from the properties shares their expressions
  until a pass folds a property's expression to a new one; the range kept the old tree, which read `hi`, and `hi` was
  removed as dead. Slider extraction now reads such a range's bounds from the properties as the passes left them
  (`Assignment.LiveRange`).
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
- `shaft.bevel(numTeeth, mateTeeth, moduleVal, pressureAngle, shaftAngle, spiralAngle, faceWidth, backlash,
  apex, from, to)`: a bevel gear as a shaft segment, placed by its pitch apex and its two flat faces, facing
  whichever way the apex lies, with `gears.bevelGear`'s tooth and ends; its toe teeth may stand past its toe
  face and the box holds them. `shaft.gear(..., bottomEnd, topEnd, backlash)` thins the teeth for a running
  pair, and `gears.bevelApexToCrown` gives the distance the faces are placed from. The gear and bevel gear
  examples are now shafts.
- `shaft.knurl(dia, pitch, pattern, from, to)`: an ISO 13444 knurl on a shaft, straight, left- or
  right-hand, diamond with its points raised or sunk, or annular. The geometry is the blank, a plain
  diameter; the teeth are a displacement height, `shaft.relief(segments)`, painted on where they are not
  drawn, `shaft.paint(color, segments)`. `shaft.body(segments, finish)` applies both and the envelope as the
  stand-in, as a fastener's finish does.
- `shaft.undercut(form)`: a thread's end against a shoulder, with the DIN 76-1 form A (`thread.din76A`) or
  ISO 4755 (`thread.iso4755`) undercut a threading tool runs out into: the floor a little under the minor
  diameter, a 30-degree flank cutting the last teeth down, and a radius at the shoulder.
  `thread.undercutDia`, `undercutLength` and `undercutRadius` give its sizes, pitch 0.2 to 6, with sources.
- Cross holes: `shaft.crossVoid(features, dia, z, angle)` drills a `drill` stack radially into a shaft, its
  depths from the surface, and `shaft.crossHole(void)` cuts it in a shaft's list; `shaft.crossPlace(part,
  dia, z, angle)` puts a part, a set screw say, on the same axis. `fastener.setScrew(nominalDia, length,
  point)`: hexagon socket set screws to ISO 4026, 4027, 4028 and 4029 (flat, cone, dog and cup point),
  M3 to M12, from a sourced table, plain or finished with their thread as relief.
- `mechanical/tolerance`: ISO 286 limits and fits, the grades IT5 to IT7 and the limit deviations of shaft
  zones j5 to p6 and hole zones H6 to M7 over 6 up to 120 mm (`tolerance.shaftUpper`, `shaftLower`,
  `holeUpper`, `holeLower`, `grade`, `shaftDia`, `holeDia`), with sources. `bearing.seatDia` and
  `bearing.housingDia` size a bearing's journal and housing bore in a zone at maximum or minimum material.
- A shaft with a gear, spline, thread, knurl or bevel stands in as its envelope (`.simple`) far off; a plain
  turned shaft has no stand-in and compiles as before. A part whose stand-in is `nothing` is drawn as itself.
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
- `gsl serve` answers `POST /api/render-lit` with the lit renderer's frame of a document from the viewer's camera:
  the request names the source, the frame's size (at most 4096 x 4096 pixels), the camera, the world extent, the
  colours, the tier (standard, color or Schematic) and the back-face brightness; the reply is the frame as raw
  bytes, five little-endian float32s a pixel (the colour, the return hit, 0 for none, and the depth), its size in
  `X-Frame-Width` and `X-Frame-Height`. While it draws, the share done goes to the asking client on
  `/api/progress` as stage `render`, in thousandths.
- The viewer can draw the frame at rest on the server's CPU: Render > Advanced > Draw on > CPU asks `gsl serve` for the
  lit renderer's frame of the document from the viewer's camera (`POST /api/render-lit`) and shows it in the GPU
  frame's place once it arrives, in the tier chosen at rest (Schematic, Color or Standard). The note beside the
  buttons counts the seconds and the percent done; the GPU button is outlined while the GPU's frame is the one on
  screen and the CPU's is still to come, and a change of view takes the screen back to the GPU and asks again. A
  view over 2 megapixels is not drawn, and the note says to lower Resolution. The page `gsl export` writes has no
  server to ask, and does not offer the device. The CPU frame has no axis lines, which the GPU draws in its shape
  pass.
- The viewer can draw the frame at rest with fidget's own GPU pipeline, run in the page: Render > Advanced > Draw on >
  Fidget, or Fidget only, which draws no shapes on the GPU and shows only fidget's frame (the last one stays up until
  the next lands, and hover, outlines and picking, which read the GPU frame, do not see shapes). The page asks
  `gsl serve` for the document's tape, with each return's colour and material fields (`POST /api/tape`), and
  fidget-wgpu, compiled to WebAssembly, finds each pixel's hit, its return, and its colour and material there; the
  standard frame's own shading lights it, with fidget's ambient occlusion and glass drawn opaque. A document that
  reads an imported mesh or noise has no such tape and is refused with a message. The buttons appear only where the
  bundle is built: `tool/build-fidget.sh --wasm-out examples/viewer/vendor` puts it in `examples/viewer/vendor/fidget-wgpu`
  (ignored by git).
- The viewer draws its moving frames smaller when it cannot keep 5 frames a second. While the view moves, four frames
  in a row slower than that halve the size the moving frames are drawn at (Draft draws them at half the still frame's,
  so this takes them to a quarter, and at the least an eighth); it steps back up when the frame time at the larger size,
  predicted from the pixel count, would clear 7.5 frames a second, and waits longer after a step up that turned out
  wrong. The still frame is always the size chosen. The Resolution buttons show it: the one chosen stays filled, the
  size being drawn is outlined, and a note beside them says "moving at 1/4: under 5 fps". The wait that takes the view
  for still, 150 ms without input, now follows the frame time when frames are slower than that (one and a half times the
  slowest of the last four, at most 1.5 s): input reaches the page once a frame, so with half-second frames the wait ran
  out between every two and a standard frame, as slow as the rest, was drawn in each gap.
- The viewer's Render panel has a Normals switch, Numeric or Analytic. Analytic asks `/api/compile`
  (`analyticNormals`) for each return's spatial gradient, taken by the assembler's forward-mode derivative
  (`assembler.WithAnalyticGradient`, `PackageFile.gradientReturns`) and emitted as `gsl_gradField`, which
  `gsl_calcNormal` uses in place of central differences. A return the derivative cannot go through, a program that
  places meshes, and a displaced material keep numeric normals, and the panel says which (`gradientNotes`). Frame
  time is unchanged within noise; the shader is larger and compiles slower (boolean_blends: 37 to 614 KB of WGSL,
  about 2.7 s more per compile).
- `gsl fix` has a `range` fixer: it replaces a top-level variable's `.min`, `.max` and `.period` property statements
  with one `@range(min, max[, period])` in front of the variable, copying the expressions as written, so
  Slider Mechanism's `slidePos`, whose bounds read other variables, converts too. The annotation shares the
  assignment's line when that fits in 100 columns and takes the line above otherwise. A variable is left as it is
  when it has an explicit `@range` already, shares its assignment with other variables, or has a comment inside a
  property's expression; a comment after a removed property stays on its own line. Run on the examples and the
  example documents (30 files), every slider (name, bounds, period, initial value) and every compiled shader is
  unchanged.
- A gradient's interval bound is no longer NaN over a box that holds a length's zero. The derivative of
  `length`, `distance` and the cellular F1, F2 and wall fields divided by the norm with a guard that holds at a
  point but not over an interval, so `.dp` of a circle, a cylinder or a ball read NaN over any box holding its
  axis or centre, and up to 1e16 near it. It is now the clamped unit vector dotted with the tangent, bounded by
  the tangent's length over any box; no value changes. `clamp`'s derivative compares the clamped value with its
  argument, not the argument with the bounds, so it differentiates a vector clamped by scalar bounds.

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
