# Omega Core SDK (`omega-core`)

`omega-core` is the standalone Go SDK for the Geometric Specification Language (GSL). It provides
pure, modular compiler libraries and the `gsl` command-line toolchain (`compile`, `serve`, `fix`)
for parsing, analyzing, refactoring, and compiling GSL models into WebGPU compute shaders (WGSL),
3D meshes (STL), bytecode tapes, and audio synthesizers.

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
- **`cmd/gsl/`**: Unified GSL toolchain CLI driver (`gsl <command> [args]`).
- **`cmd/compile/`**: Standalone multi-backend command-line compiler (`gsl compile`).
- **`cmd/serve/`**: Local HTTP server and interactive WebGPU viewer host (`gsl serve`).
- **`cmd/fix/`**: Source refactoring tool for migrating legacy GSL syntax (`gsl fix`).

### Directory Layout

```text
omega-core/
├── go.mod                     # module gradientcontrol.com/omega/core
├── assembler/
│   ├── assemble.go            # Assemble(primarySources, opts...) API
│   ├── options.go             # Options (SourcePath, SourceManager, WithProgress...)
│   └── source/                # Source abstraction
├── backend/                   # Target code generators
│   ├── audio/                 # Web Audio synthesis generator
│   ├── mesh/                  # Unified 3D mesh generator
│   ├── tape/                  # Bytecode tape compiler
│   └── wgsl/                  # WebGPU shader emitter
├── bin/                       # Local build outputs (gsl, compile, serve, fix binaries; gitignored)
├── cmd/                       # Standalone CLI tools
│   ├── compile/               # Unified compiler CLI (emits WGSL, STL, tape, audio)
│   ├── fix/                   # GSL source refactoring CLI (rewrites legacy comments to directives)
│   ├── gsl/                   # Unified GSL toolchain CLI driver (delegates to bin/ tools)
│   └── serve/                 # Local HTTP server and WebGPU viewer host
├── examples/                  # Curated showcase GSL models and assets
│   ├── assets/                # Traditional CAD mesh assets (bracket.stl)
│   └── viewer/                # Interactive WebGPU example viewer web application
│       ├── docs/              # Built-in GSL language reference and interactive help panel
│       ├── editor/            # CodeMirror 6 GSL/WGSL editor, autocomplete, and search
│       ├── fonts/             # Self-hosted UI (PublicSans) and code (IBMPlexMono) fonts
│       ├── icons/             # SVG toolbar and inspector icons
│       ├── materials/         # PBR material catalog, BRDF lookup tables, and swatches
│       └── vendor/            # Vendored browser libraries (Three.js, CodeMirror, Manifold)
├── parser/                    # Lexer and Parser
├── program/                   # Core Program/AST models and definitions
│   ├── program.proto          # Protobuf AST schema definitions
│   └── scene/                 # Scene descriptor construction and extent utilities
├── stdlib/                    # GSL Standard Library
│   ├── pkg_info.go            # Package registry and compile-time embedded source data
│   ├── builtin/
│   ├── color/
│   ├── ...
│   ├── transform/
│   └── vector/
└── tool/                      # Developer build and automation scripts
    ├── build-cli.sh           # Builds compiler CLI executables (gsl, compile, serve, fix) into bin/
    ├── build-fidget.sh        # Builds Fidget and Manifold Rust/C++ native and WASM artifacts
    ├── build-proto.sh         # Compiles program.proto and tape.proto into Go bindings
    ├── build-wgsl-accel-ops.sh # Regenerates WGSL hardware acceleration ops table and inliner registry
    ├── check.sh               # Pre-commit quality gate (gofmt, go vet, tests)
    ├── clean.sh               # Cleans bin/ outputs, test cache, and scratch files
    ├── test-go.sh             # Runs Go unit, integration, and golden test suites
    ├── test-rust.sh           # Runs Rust unit and doc test suites
    ├── test.sh                # Runs both Go and Rust test suites
    ├── update-goldens.sh      # Regenerates reference test goldens
    └── run-go-tests/          # Go test suite runner, timing history, and golden size reporter
```

## Building the Toolchain

Compile the `gsl`, `compile`, `serve`, and `fix` binaries into `bin/`:

```bash
./tool/build-cli.sh
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

## Toolchain Driver (`bin/gsl`) Usage

```text
gsl <command> [arguments]
```

Subcommands delegate to sibling tool binaries in `bin/`:
- `gsl compile` (aliases `build`, `c`): Compile GSL sources to target outputs (`bin/compile`).
- `gsl serve`: Launch the local HTTP server and interactive WebGPU viewer (`bin/serve`).
- `gsl fix`: Refactor GSL source files to current syntax (`bin/fix`).
- `gsl help [command]`: Display help for `gsl` or a specific subcommand.

### Supported Compiler Targets (`-t, --target`)

- `wgsl`: WebGPU signed distance field shader code (default).
- `stl`: 3D printable watertight binary polygon mesh via dual-contouring.
- `tape`: Compact vectorized spatial bytecode evaluation tape.
- `audio`: Web Audio JavaScript oscillator and synthesis code.
- `ast`: Formatted JSON AST dump for structural inspection.

### Companion Metadata (`<output>.json`)

When compiling to output artifacts (such as `.wgsl` or `.stl`), `gsl compile` automatically emits a
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
bin/gsl compile -t wgsl -o model.wgsl model.gsl
```

Compile a standalone WebGPU shader bundle with full-screen raymarcher harness (`vs_main`, `fs_main`):

```bash
bin/gsl compile --standalone -o preview.wgsl model.gsl
```

Compile a model into a watertight binary STL mesh:

```bash
bin/gsl compile -t stl --subdivision 5 --algo marching-cubes -o model.stl model.gsl
```

Compile from standard input (`stdin`) and stream to standard output (`stdout`):

```bash
cat model.gsl | bin/gsl compile -t wgsl -o -
```

Refactor `.gsl` files or directories in-place (or preview with `-diff`):

```bash
bin/gsl fix -diff examples
bin/gsl fix examples
```

## License

The code and data authored for this repository are released under the [MIT License](LICENSE).
Third-party components keep their own licenses: they are listed, with their license texts, in
[THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md), and the vendored directories carry their own
license files (for example `internal/fidget/vendor/`, which is MPL-2.0).
