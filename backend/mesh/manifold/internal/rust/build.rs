use std::path::{Path, PathBuf};

// Build manifold from vendor/manifold as a static lib, then compile our thin
// C-ABI glue (manifold_glue.cpp) and link it all together. The resulting
// dylib exposes a flat C interface that the Go package calls via cgo.
//
// Parallel backend (TBB) is intentionally OFF for the first cut — manifold's
// serial path is already orders of magnitude faster than the WASM build, and
// disabling MANIFOLD_PAR removes the TBB integration headache. Cross-section
// (2D) is also off because the boolean composer never touches 2D paths.
fn main() {
    let manifest = PathBuf::from(env!("CARGO_MANIFEST_DIR"));
    let workspace_root = manifest
        .ancestors()
        .nth(5)
        .expect("workspace root above backend/mesh/manifold/internal/rust");
    let vendor = workspace_root.join("vendor");

    if !vendor.join("manifold").join("CMakeLists.txt").exists() {
        panic!(
            "vendor/manifold submodule not initialized. Run:\n  \
             git submodule update --init --recursive\n\
             from {}",
            workspace_root.display()
        );
    }

    let manifold_install = build_manifold(&vendor);

    let include = manifold_install.join("include");
    cc::Build::new()
        .cpp(true)
        .file("src/cpp/manifold_glue.cpp")
        .include("src/cpp")
        .include(&include)
        .flag_if_supported("-std=c++17")
        .compile("manifold_glue");

    // Manifold's static archive is split into manifold + (optionally) clipper2,
    // tbb, etc. With CROSS_SECTION=OFF and PAR=OFF the only library we need
    // back is `manifold` itself.
    emit_static_link(&manifold_install, &["manifold"]);
    let cpp = if cfg!(target_os = "macos") { "c++" } else { "stdc++" };
    println!("cargo:rustc-link-lib={cpp}");

    println!("cargo:rerun-if-changed=src/cpp/manifold_glue.cpp");
    println!("cargo:rerun-if-changed=src/cpp/manifold_glue.h");
    println!("cargo:rerun-if-changed=src/lib.rs");
    println!("cargo:rerun-if-changed=build.rs");
}

fn build_manifold(vendor: &Path) -> PathBuf {
    cmake::Config::new(vendor.join("manifold"))
        .define("BUILD_SHARED_LIBS", "OFF")
        .define("CMAKE_BUILD_TYPE", "Release")
        .define("CMAKE_POLICY_VERSION_MINIMUM", "3.5")
        // Strip everything we don't use so the link surface stays small and
        // the build doesn't pull in optional dependencies we'd otherwise have
        // to vendor (Clipper2, TBB, nanobind, …).
        .define("MANIFOLD_TEST", "OFF")
        .define("MANIFOLD_PYBIND", "OFF")
        .define("MANIFOLD_CROSS_SECTION", "OFF")
        .define("MANIFOLD_PAR", "OFF")
        .define("MANIFOLD_DOWNLOADS", "OFF")
        .define("MANIFOLD_DEBUG", "OFF")
        .build()
}

fn emit_static_link(install: &Path, libs: &[&str]) {
    for &lib in libs {
        println!("cargo:rustc-link-lib=static={}", lib);
    }
    let lib_dir = install.join("lib");
    let lib64_dir = install.join("lib64");
    if lib_dir.exists() {
        println!("cargo:rustc-link-search=native={}", lib_dir.display());
    }
    if lib64_dir.exists() {
        println!("cargo:rustc-link-search=native={}", lib64_dir.display());
    }
}
