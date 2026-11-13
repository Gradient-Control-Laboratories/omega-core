# Changelog

All notable changes to `omega-core` are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.2] - 2026-11-12

### Added

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

### Fixed

- Fixed Manifold Dual Contouring (`fidget-mesh`) corner chamfering and octree over-collapse when an
  edge-crossing midpoint lands on an exact `min`/`max` constant tie (`grad == 0`, #38) by evaluating
  the gradient again a hair off the crossing, and marking leaves that still lack a plane normal as
  non-collapsible (`QEF_ERR_INVALID`).
- Fixed `volume` queries without a tolerance (`gsl eval`'s `relativeTolerance`, Python's
  `Document.volume` default) failing with "volume tolerance must be between zero and
  one": `eval.Volume` now reads a tolerance `<= 0` as the default `0.05`, as `surface` and `extremum`
  do.

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

[Unreleased]: https://github.com/Gradient-Control-Laboratories/omega-core/compare/v0.1.2...HEAD
[0.1.2]: https://github.com/Gradient-Control-Laboratories/omega-core/compare/v0.1.1...v0.1.2
[0.1.1]: https://github.com/Gradient-Control-Laboratories/omega-core/compare/v0.1.0...v0.1.1
[0.1.0]: https://github.com/Gradient-Control-Laboratories/omega-core/releases/tag/v0.1.0
