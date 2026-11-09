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
- **Git submodules** — the C++ deps are pinned under `internal/fidget/vendor/` (beside the vendored fidget crates). They must be initialized with:

```sh
# from the repo root
git submodule update --init --recursive
```

## Build
[`tool/build-fidget.sh`](../../tool/build-fidget.sh) builds this crate and the [manifold](../../backend/mesh/manifold) glue that `backend/mesh` uses for boolean composition. The steps below build the fidget crate on its own, for working on it or verifying a fresh setup. The first build of OpenVDB + its transitive dependencies takes several minutes, builds after that are incremental.

```sh
# from the repo root
cd internal/fidget/internal/rust
cargo build --release
```

This produces `target/release/librender.{dylib,so}` at the repo root, which the Go side links via `cgo` (`#cgo LDFLAGS: -L../../../../target/release -lrender`). A debug build (`cargo build`) is very slow at `bvhSdf()`, so `tool/build-fidget.sh` uses `cargo build --release`.

After the Rust lib is in place, the Go package compiles normally:

```sh
go build ./internal/fidget/...
```
