fn main() {
    // pyo3's `extension-module` feature deliberately does not link libpython —
    // Python symbols are resolved by the host interpreter at import time. On
    // macOS a plain `cargo build` of a cdylib still needs the linker told to
    // allow those undefined symbols (this is what maturin injects for us).
    if std::env::var("CARGO_CFG_TARGET_OS").as_deref() == Ok("macos") {
        println!("cargo:rustc-cdylib-link-arg=-undefined");
        println!("cargo:rustc-cdylib-link-arg=dynamic_lookup");
    }
}
