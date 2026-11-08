# Fidget Mesher

The fidget package provides server-side SDF meshing and evaluation. It wraps a Rust crate (`internal/rust/`) that links against a patched [fidget](https://github.com/mkeeter/fidget) plus the C++ [OpenVDB](https://github.com/AcademySoftwareFoundation/openvdb) library for narrow-band level-set sampling.

## Prerequisites
You need:

- **Go** (1.24+) — already required by the rest of the repo.
- **Rust** (stable, via [rustup](https://www.rust-lang.org/tools/install)) — needed for the `internal/rust` crate.
- **CMake** (3.5+ recommended) — the `build.rs` invokes CMake to build OpenVDB and its deps (zlib, blosc, tbb, imath).
  - macOS: `brew install cmake`
  - Debian/Ubuntu: `apt-get install cmake build-essential`
- A C++17 toolchain — clang on macOS, g++/clang on Linux.
- **Git submodules** — the C++ deps are pinned under `vendor/` in the repo root. They must be initialized with:

```sh
# from the repo root
git submodule update --init --recursive
```

## Build
- **[webedit](../../webedit/README.md)** launches the fidget agent for server-side meshing alongside the [manifold](../manifold) package for boolean composition.

The launch scripts for both these agents already invoke cargo builds for you — the manual build steps below are only needed if you're working on the agent itself or you want to verify a fresh setup. The first build of OpenVDB + its transitive dependencies takes several minutes, builds after that are incremental.

```sh
# from the repo root
cd agent/fidget/internal/rust
cargo build --release
```

This produces `omega/target/release/librender.{dylib,so}`, which the Go side of the agent links via `cgo` (`#cgo LDFLAGS: -L../../../../target/release -lrender`). A debug build (`cargo build`) is very slow at `bvhSdf()`, so the launch scripts default to `cargo build --release`.

After the Rust lib is in place, the Go agent compiles normally:

```sh
go build ./agent/fidget/...
```
