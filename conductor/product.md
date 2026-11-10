# Product Definition: Omega Core (`omega-core`)

## Vision

`omega-core` (`module gradientcontrol.com/omega/core`) is the standalone Go compiler SDK and CLI
toolchain for the Geometric Specification Language (GSL). It provides pure, modular compiler
libraries and command-line tools (`gsl compile`, `gsl serve`, `gsl fix`) for parsing, linking,
evaluating, certifying, and compiling GSL signed-distance-field (SDF) models into WebGPU compute
shaders (WGSL), watertight 3D meshes (STL), spatial bytecode tapes, and WebAudio synthesizers.

## Core Features

- **GSL Lexer & Parser (`parser`):** Stateless Participle-based GSL parser and thread-safe LRU
  `CachingParser` producing canonical protobuf ASTs (`program.Package`).
- **Multi-Package Assembler (`assembler`):** Linker, lexical scope analyzer, static type checker,
  constant evaluator, closure monomorphizer, inliner, dead-code eliminator, and mesh-plan slicer.
- **Certified Geometry Evaluator (`eval`):** Spatial SDF/gradient evaluator, bounding-box finder,
  Lipschitz extremum solver, and interval-arithmetic formal verifier (`eval`, `eval/joints`).
- **Multi-Target Backends (`backend/*`):**
  - `backend/wgsl`: WebGPU SDF raymarching shader emitter and 3D texture binding helpers.
  - `backend/mesh` & `backend/mesh/manifold`: Dual-contouring SDF mesher (Fidget) and exact CSG
    boolean composer (Manifold) emitting watertight binary STL meshes.
  - `backend/tape`: Compact spatial VM bytecode tape compiler and `.wtb` bundle serializer.
  - `backend/audio`: WebAudio JavaScript synthesis emitter.
- **Embedded Standard Library (`stdlib`):** Compile-time embedded GSL standard library packages
  (`builtin`, `map`, `transform`, `lattice/*`, `mechanical/*`, `material/*`, `vector`, `matrix`).
- **Unified Toolchain CLI (`cmd/gsl`):** Standalone `compile`, `serve` (with interactive WebGPU
  viewer in `examples/viewer/`), and `fix` subcommands.

## Target Users

- Host applications and microservice stacks (such as `omega`) embedding the GSL compiler in-process.
- Computational designers and engineers authoring, compiling, and formally verifying parametric 3D
  models from the command line or local WebGPU viewer.

