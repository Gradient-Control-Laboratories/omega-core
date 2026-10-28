# Omega Core SDK (`omega_core`)

`omega_core` is the standalone Go SDK for the Geometric Specification Language (GSL). It provides
pure, modular compiler libraries and the `gslc` command-line compiler for parsing, analyzing, and
compiling GSL models into WebGPU compute shaders (WGSL), 3D meshes (STL), bytecode tapes,
and audio synthesizers.

## Architecture & Package Topology

- **`parser/`**: Stateless GSL lexer, parser, and AST generator.
- **`assembler/`**: Multi-package linker, type checker, inliner, constant evaluator, and optimizer.
- **`program/`**: Canonical AST definitions, type systems, and visitor models.
- **`backend/`**:
  - `wgsl/`: WGSL compute shader emitter for real-time GPU SDF raymarching.
  - `mesh/`: Unified 3D mesh generator coordinating dual-contouring (Fidget) and CSG (Manifold).
  - `audio/`: Audio synthesizer compiler emitting WebAudio / JavaScript synthesis functions.
  - `tape/`: Bytecode tape compiler for compact embedded geometry evaluation.
- **`stdlib/`**: Compile-time embedded GSL standard library packages (`builtin`, `color`, `vector`, `matrix`, etc.).
- **`cmd/gslc/`**: Standalone multi-backend command-line compiler driver.

## Building the Toolchain

Compile the `gslc` binary into `bin/gslc`:

```bash
./tool/build.sh
```

Run pre-commit quality checks (gofmt, go vet, and tests):

```bash
./tool/check.sh
```

Run unit and integration test suites:

```bash
./tool/test.sh
```

Clean build outputs and test cache:

```bash
./tool/clean.sh
```

## Compiler Driver (`bin/gslc`) Usage

```text
gslc [flags] [source1.gsl source2.gsl ...]
```

### Supported Targets (`-t, --target`)

- `wgsl`: WebGPU signed distance field shader code (default).
- `stl`: 3D printable watertight binary polygon mesh via dual-contouring.
- `tape`: Compact vectorized spatial bytecode evaluation tape.
- `audio`: Web Audio JavaScript oscillator and synthesis code.
- `ast`: Formatted JSON AST dump for structural inspection.

### Companion Metadata (`<output>.json`)

When compiling to output artifacts (such as `.wgsl` or `.stl`), `gslc` automatically emits a
companion `<output>.json` metadata file alongside the output. This document includes:
- Extracted `@slider` definitions (names, min, max, initial values, and precision).
- Extracted UI controls and layout descriptors.
- Scene bounding box dimensions and camera presets.
- Non-fatal compiler diagnostics and warning spans.

To specify a custom metadata destination, pass `-m <file>`. To disable companion metadata emission,
pass `--no-meta`.

### Common Workflow Examples

Compile a model to a WebGPU shader with auto-emitted `model.json` metadata:

```bash
bin/gslc -t wgsl -o model.wgsl model.gsl
```

Compile a standalone WebGPU shader bundle with full-screen raymarcher harness (`vs_main`, `fs_main`):

```bash
bin/gslc --standalone -o preview.wgsl model.gsl
```

Compile a model into a watertight binary STL mesh:

```bash
bin/gslc -t stl --subdivision 5 --algo marching-cubes -o model.stl model.gsl
```

Compile from standard input (`stdin`) and stream to standard output (`stdout`):

```bash
cat model.gsl | bin/gslc -t wgsl -o -
```
