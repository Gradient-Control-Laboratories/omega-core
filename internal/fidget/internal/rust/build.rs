use std::path::{Path, PathBuf};

fn main() {
    let manifest = PathBuf::from(env!("CARGO_MANIFEST_DIR"));
    let workspace_root = manifest
        .ancestors()
        .nth(4)
        .expect("workspace root above agent/fidget/internal/rust");
    let vendor = workspace_root.join("internal/fidget/vendor");

    if !vendor.join("openvdb").join("CMakeLists.txt").exists() {
        panic!(
            "internal/fidget/vendor submodules not initialized. Run:\n  \
             git submodule update --init --recursive\n\
             from {}",
            workspace_root.display()
        );
    }

    let zlib_install = build_zlib(&vendor);
    let blosc_install = build_blosc(&vendor, &zlib_install);
    let tbb_install = build_tbb(&vendor);
    let imath_install = build_imath(&vendor);
    let openvdb_install = build_openvdb(
        &vendor,
        &zlib_install,
        &blosc_install,
        &tbb_install,
        &imath_install,
    );

    cxx_build::bridge("src/openvdb_ffi.rs")
        .file("src/cpp/openvdb_glue.cpp")
        .include("src/cpp")
        .include(openvdb_install.join("include"))
        .include(imath_install.join("include").join("Imath"))
        .include(imath_install.join("include"))
        .include(tbb_install.join("include"))
        .std("c++17")
        .compile("openvdb_glue");

    emit_static_link(&openvdb_install, &["openvdb"]);
    emit_static_link(&imath_install, &["Imath-3_2"]);
    emit_static_link(&tbb_install, &["tbb"]);
    emit_static_link(&blosc_install, &["blosc"]);
    emit_static_link(&zlib_install, &["z"]);
    let cpp = if cfg!(target_os = "macos") { "c++" } else { "stdc++" };
    println!("cargo:rustc-link-lib={cpp}");

    // The omega tape's protobuf types are generated in tape-core.

    println!("cargo:rerun-if-changed=src/openvdb_ffi.rs");
    println!("cargo:rerun-if-changed=src/cpp/openvdb_glue.cpp");
    println!("cargo:rerun-if-changed=src/cpp/openvdb_glue.h");
    println!("cargo:rerun-if-changed=build.rs");
}

fn build_zlib(vendor: &Path) -> PathBuf {
    let install = cmake::Config::new(vendor.join("zlib"))
        .define("BUILD_SHARED_LIBS", "OFF")
        .define("ZLIB_BUILD_EXAMPLES", "OFF")
        .define("CMAKE_POLICY_VERSION_MINIMUM", "3.5")
        .define("CMAKE_BUILD_TYPE", "Release")
        .build();
    // zlib's CMakeLists builds both static and shared targets even with
    // BUILD_SHARED_LIBS=OFF (it sets up `zlib` and `zlibstatic` targets
    // unconditionally). Delete the dylibs so the macOS linker can't
    // accidentally prefer them over libz.a — without this librender.dylib
    // ends up with @rpath/libz.1.dylib in its load commands.
    let lib_dir = install.join("lib");
    if let Ok(entries) = std::fs::read_dir(&lib_dir) {
        for entry in entries.flatten() {
            let name = entry.file_name();
            let name_str = name.to_string_lossy();
            if name_str.starts_with("libz") && (name_str.ends_with(".dylib") || name_str.ends_with(".so")) {
                let _ = std::fs::remove_file(entry.path());
            }
        }
    }
    install
}

fn build_openvdb(
    vendor: &Path,
    zlib_install: &Path,
    blosc_install: &Path,
    tbb_install: &Path,
    imath_install: &Path,
) -> PathBuf {
    let prefix_path = format!(
        "{};{};{};{}",
        zlib_install.display(),
        blosc_install.display(),
        tbb_install.display(),
        imath_install.display()
    );
    cmake::Config::new(vendor.join("openvdb"))
        .define("OPENVDB_CORE_STATIC", "ON")
        .define("OPENVDB_CORE_SHARED", "OFF")
        .define("USE_BLOSC", "ON")
        .define("USE_ZLIB", "ON")
        .define("USE_EXPLICIT_INSTANTIATION", "OFF")
        .define("OPENVDB_USE_DELAYED_LOADING", "OFF")
        .define("OPENVDB_BUILD_BINARIES", "OFF")
        .define("OPENVDB_BUILD_DOCS", "OFF")
        .define("OPENVDB_BUILD_PYTHON_MODULE", "OFF")
        .define("OPENVDB_BUILD_NANOVDB", "OFF")
        .define("OPENVDB_BUILD_HOUDINI_PLUGIN", "OFF")
        .define("OPENVDB_BUILD_MAYA_PLUGIN", "OFF")
        .define("OPENVDB_BUILD_AX", "OFF")
        .define("OPENVDB_BUILD_VDB_PRINT", "OFF")
        .define("OPENVDB_BUILD_VDB_LOD", "OFF")
        .define("OPENVDB_BUILD_VDB_RENDER", "OFF")
        .define("OPENVDB_BUILD_VDB_VIEW", "OFF")
        .define("OPENVDB_BUILD_VDB_AX", "OFF")
        .define("OPENVDB_BUILD_UNITTESTS", "OFF")
        .define("OPENVDB_INSTALL_CMAKE_MODULES", "OFF")
        .define("OPENVDB_TOPLEVEL_DEPS_ONLY", "OFF")
        .define("CMAKE_PREFIX_PATH", &prefix_path)
        .define("CMAKE_POLICY_VERSION_MINIMUM", "3.5")
        .define("CMAKE_BUILD_TYPE", "Release")
        .build()
}

fn build_imath(vendor: &Path) -> PathBuf {
    cmake::Config::new(vendor.join("imath"))
        .define("BUILD_SHARED_LIBS", "OFF")
        .define("IMATH_INSTALL", "ON")
        .define("BUILD_TESTING", "OFF")
        .define("CMAKE_POLICY_VERSION_MINIMUM", "3.5")
        .define("CMAKE_BUILD_TYPE", "Release")
        .build()
}

fn build_tbb(vendor: &Path) -> PathBuf {
    cmake::Config::new(vendor.join("tbb"))
        .define("BUILD_SHARED_LIBS", "OFF")
        .define("TBB_TEST", "OFF")
        .define("TBB_EXAMPLES", "OFF")
        .define("TBB_STRICT", "OFF")
        .define("CMAKE_POLICY_VERSION_MINIMUM", "3.5")
        .define("CMAKE_BUILD_TYPE", "Release")
        .build()
}

fn build_blosc(vendor: &Path, zlib_install: &Path) -> PathBuf {
    cmake::Config::new(vendor.join("blosc"))
        .define("BUILD_SHARED", "OFF")
        .define("BUILD_STATIC", "ON")
        .define("BUILD_TESTS", "OFF")
        .define("BUILD_FUZZERS", "OFF")
        .define("BUILD_BENCHMARKS", "OFF")
        .define("PREFER_EXTERNAL_ZLIB", "ON")
        .define("DEACTIVATE_ZSTD", "ON")
        .define("DEACTIVATE_LZ4", "ON")
        .define("DEACTIVATE_SNAPPY", "ON")
        .define("CMAKE_PREFIX_PATH", zlib_install)
        .define("CMAKE_POLICY_VERSION_MINIMUM", "3.5")
        .define("CMAKE_BUILD_TYPE", "Release")
        .build()
}

fn emit_static_link(install_prefix: &Path, libs: &[&str]) {
    let lib_dir = install_prefix.join("lib");
    println!("cargo:rustc-link-search=native={}", lib_dir.display());
    // RHEL-family CMake (the manylinux image) installs into lib64.
    let lib64_dir = install_prefix.join("lib64");
    if lib64_dir.exists() {
        println!("cargo:rustc-link-search=native={}", lib64_dir.display());
    }
    for lib in libs {
        println!("cargo:rustc-link-lib=static={lib}");
    }
}
