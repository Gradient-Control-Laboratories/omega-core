# Omega Core SDK (`omega_core`)

`omega_core` is the standalone Go SDK for the Geometric Specification Language (GSL). It provides
pure, modular compiler libraries and the `gslc` command-line compiler for parsing, analyzing, and
compiling GSL models into WebGPU compute shaders (WGSL), 3D meshes (STL/GLTF), bytecode tapes,
and audio synthesizers.

## Architecture & Package Topology

- **`parser/`**: Stateless GSL lexer, parser, and AST generator.
- **`assembler/`**: Multi-package linker, type checker, inliner, constant evaluator, and optimizer.
- **`program/`**: Canonical AST definitions, type systems, and visitor models.
- **`backend/`**:
  - `wgsl/`: WGSL compute shader emitter for real-time GPU SDF raymarching.
  - `mesh/`: Unified 3D mesh generator coordinating analytical dual-contouring (Fidget) and CSG (Manifold).
  - `audio/`: Audio synthesizer compiler emitting WebAudio / JavaScript oscillators.
  - `tape/`: Bytecode tape compiler for compact embedded geometry evaluation.
- **`stdlib/`**: Compile-time embedded GSL standard library packages (`builtin`, `color`, `vector`, `matrix`, etc.).
- **`cmd/gslc/`**: Standalone multi-backend command-line compiler.

## Building

```bash
./tool/build.sh
```
