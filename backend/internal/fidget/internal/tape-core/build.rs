use std::path::PathBuf;

fn main() {
    let manifest = PathBuf::from(env!("CARGO_MANIFEST_DIR"));
    let workspace_root = manifest
        .ancestors()
        .nth(4)
        .expect("workspace root above agent/fidget/internal/tape-core");

    // Generate Rust types for the omega tape interchange format from the same
    // common/tape/tape.proto that the Go side compiles — single source of
    // truth. Requires protoc on PATH (already needed for the Go codegen).
    let tape_proto = workspace_root.join("common").join("tape").join("tape.proto");
    let tape_proto_dir = tape_proto.parent().unwrap().to_path_buf();
    prost_build::Config::new()
        .compile_protos(&[&tape_proto], &[&tape_proto_dir])
        .expect("failed to compile common/tape/tape.proto with prost-build");
    println!("cargo:rerun-if-changed={}", tape_proto.display());
    println!("cargo:rerun-if-changed=build.rs");
}
