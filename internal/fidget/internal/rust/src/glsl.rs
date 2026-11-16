//! WGSL to GLSL ES 3.00, for a host that draws with WebGL2.
//!
//! `backend/wgsl` writes WGSL only. A WebGL2 host needs the same module as a
//! `#version 300 es` fragment shader, and naga's names for what the host must
//! bind: the uniform blocks and the textures. naga renames both, so the names
//! are returned beside the text, keyed by the WGSL global they came from.

use crate::{build_render_result, RenderResult};
use serde::Serialize;
use std::collections::BTreeMap;
use std::os::raw::c_int;

/// The arguments were unusable: a null pointer, or bytes that are not UTF-8.
const ERR_ARGUMENT: c_int = -60;
/// naga refused the module; the result's bytes are its message.
const ERR_SHADER: c_int = -61;
/// naga panicked.
const ERR_PANIC: c_int = -62;

/// A fragment shader in GLSL ES 3.00, and the GLSL name of each WGSL global
/// the host binds.
#[derive(Debug, Serialize)]
pub struct Glsl {
    pub glsl: String,
    /// WGSL `var<uniform>` name to GLSL uniform block name.
    #[serde(rename = "uniformBlocks")]
    pub uniform_blocks: BTreeMap<String, String>,
    /// WGSL texture name to GLSL sampler name.
    pub textures: BTreeMap<String, String>,
}

/// Transpile the fragment entry point `entry` of `wgsl` to GLSL ES 3.00.
///
/// The error is naga's own message, with the source excerpt it points at.
pub fn wgsl_to_glsl_es300(wgsl: &str, entry: &str) -> Result<Glsl, String> {
    use naga::back::glsl;

    let module = naga::front::wgsl::parse_str(wgsl).map_err(|e| e.emit_to_string(wgsl))?;
    // No optional capability: what `backend/wgsl` writes needs none, and a
    // module that does would not run on WebGL2.
    let info = naga::valid::Validator::new(
        naga::valid::ValidationFlags::all(),
        naga::valid::Capabilities::empty(),
    )
    .validate(&module)
    .map_err(|e| e.emit_to_string(wgsl))?;

    let options = glsl::Options {
        version: glsl::Version::Embedded {
            version: 300,
            is_webgl: true,
        },
        ..Default::default()
    };
    let pipeline = glsl::PipelineOptions {
        shader_stage: naga::ShaderStage::Fragment,
        entry_point: entry.to_string(),
        multiview: None,
    };
    let mut text = String::new();
    let mut writer = glsl::Writer::new(
        &mut text,
        &module,
        &info,
        &options,
        &pipeline,
        naga::proc::BoundsCheckPolicies::default(),
    )
    .map_err(|e| e.to_string())?;
    let reflection = writer.write().map_err(|e| e.to_string())?;

    let name =
        |global: naga::Handle<naga::GlobalVariable>| module.global_variables[global].name.clone();
    let mut uniform_blocks = BTreeMap::new();
    for (global, block) in &reflection.uniforms {
        if let Some(name) = name(*global) {
            uniform_blocks.insert(name, block.clone());
        }
    }
    let mut textures = BTreeMap::new();
    for (sampler, mapping) in &reflection.texture_mapping {
        if let Some(name) = name(mapping.texture) {
            textures.insert(name, sampler.clone());
        }
    }
    Ok(Glsl {
        glsl: text,
        uniform_blocks,
        textures,
    })
}

/// The C entry point: `wgsl` and `entry` are UTF-8, not NUL-terminated.
///
/// Code 0 carries the [`Glsl`] as JSON. Any other code carries a message as
/// UTF-8 bytes. Free either with `render_result_free`.
///
/// # Safety
/// `wgsl` must point to `wgsl_len` bytes and `entry` to `entry_len`.
#[no_mangle]
pub unsafe extern "C" fn wgsl_to_glsl_es300_json(
    wgsl: *const u8,
    wgsl_len: usize,
    entry: *const u8,
    entry_len: usize,
) -> RenderResult {
    if wgsl.is_null() || entry.is_null() {
        return message(ERR_ARGUMENT, "wgsl_to_glsl_es300_json: null argument");
    }
    let wgsl = unsafe { std::slice::from_raw_parts(wgsl, wgsl_len) };
    let entry = unsafe { std::slice::from_raw_parts(entry, entry_len) };
    let (Ok(wgsl), Ok(entry)) = (std::str::from_utf8(wgsl), std::str::from_utf8(entry)) else {
        return message(ERR_ARGUMENT, "wgsl_to_glsl_es300_json: not UTF-8");
    };
    match std::panic::catch_unwind(|| wgsl_to_glsl_es300(wgsl, entry)) {
        Ok(Ok(out)) => match serde_json::to_vec(&out) {
            Ok(json) => build_render_result(json),
            Err(e) => message(ERR_PANIC, &e.to_string()),
        },
        Ok(Err(e)) => message(ERR_SHADER, &e),
        Err(_) => message(ERR_PANIC, "naga panicked"),
    }
}

fn message(code: c_int, text: &str) -> RenderResult {
    RenderResult {
        code,
        ..build_render_result(text.as_bytes().to_vec())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    /// One uniform block, one 3D texture read by `textureLoad`, one function.
    const SHADER: &str = r#"
struct Host { extent: f32, time: f32 }
@group(0) @binding(0) var<uniform> host: Host;
@group(0) @binding(1) var sdfTex: texture_3d<f32>;

fn field(p: vec3<f32>) -> f32 {
    let texel = textureLoad(sdfTex, vec3<i32>(p * host.extent), 0);
    return texel.r + host.time;
}

@fragment
fn fs_main(@builtin(position) at: vec4<f32>) -> @location(0) vec4<f32> {
    return vec4<f32>(field(at.xyz));
}
"#;

    #[test]
    fn transpiles_to_glsl_es_300() {
        let out = wgsl_to_glsl_es300(SHADER, "fs_main").expect("transpile");
        assert!(out.glsl.starts_with("#version 300 es"), "{}", out.glsl);
        assert!(out.glsl.contains("sampler3D"), "{}", out.glsl);
        assert!(out.glsl.contains("texelFetch"), "{}", out.glsl);
        assert!(out.glsl.contains("float field("), "{}", out.glsl);
    }

    #[test]
    fn names_what_the_host_binds() {
        let out = wgsl_to_glsl_es300(SHADER, "fs_main").expect("transpile");
        assert_eq!(out.uniform_blocks.keys().collect::<Vec<_>>(), ["host"]);
        assert_eq!(out.textures.keys().collect::<Vec<_>>(), ["sdfTex"]);
        let block = &out.uniform_blocks["host"];
        assert!(
            out.glsl.contains(&format!("uniform {block} {{")),
            "{block}\n{}",
            out.glsl
        );
        let sampler = &out.textures["sdfTex"];
        assert!(
            out.glsl.contains(&format!("sampler3D {sampler};")),
            "{sampler}\n{}",
            out.glsl
        );
    }

    #[test]
    fn invalid_wgsl_is_an_error_with_nagas_message() {
        let err = wgsl_to_glsl_es300("fn f( {", "fs_main").unwrap_err();
        assert!(err.contains("expected identifier"), "{err}");
    }

    #[test]
    fn a_module_that_does_not_validate_is_an_error() {
        let err = wgsl_to_glsl_es300(
            "fn f() -> f32 { return 1; }\n@fragment fn fs_main() { let x: i32 = f(); }",
            "fs_main",
        )
        .unwrap_err();
        assert!(!err.is_empty());
    }

    #[test]
    fn a_missing_entry_point_is_an_error() {
        let err = wgsl_to_glsl_es300(SHADER, "nope").unwrap_err();
        assert!(!err.is_empty());
    }

    fn call(wgsl: &[u8], entry: &[u8]) -> (c_int, Vec<u8>) {
        let r = unsafe {
            wgsl_to_glsl_es300_json(wgsl.as_ptr(), wgsl.len(), entry.as_ptr(), entry.len())
        };
        let code = r.code;
        let bytes = unsafe { std::slice::from_raw_parts(r.data, r.len) }.to_vec();
        crate::render_result_free(r);
        (code, bytes)
    }

    #[test]
    fn the_c_entry_point_answers_in_json() {
        let (code, bytes) = call(SHADER.as_bytes(), b"fs_main");
        assert_eq!(code, 0, "{}", String::from_utf8_lossy(&bytes));
        let v: serde_json::Value = serde_json::from_slice(&bytes).expect("json");
        assert!(v["glsl"].as_str().unwrap().starts_with("#version 300 es"));
        assert!(v["uniformBlocks"]["host"].is_string());
        assert!(v["textures"]["sdfTex"].is_string());
    }

    #[test]
    fn the_c_entry_point_carries_the_message_of_a_refusal() {
        let (code, bytes) = call(b"fn f( {", b"fs_main");
        assert_eq!(code, ERR_SHADER);
        assert!(String::from_utf8_lossy(&bytes).contains("expected identifier"));
    }

    #[test]
    fn the_c_entry_point_refuses_bad_arguments() {
        let r = unsafe { wgsl_to_glsl_es300_json(std::ptr::null(), 0, b"fs_main".as_ptr(), 7) };
        assert_eq!(r.code, ERR_ARGUMENT);
        crate::render_result_free(r);
        let (code, _) = call(&[0xff, 0xfe], b"fs_main");
        assert_eq!(code, ERR_ARGUMENT);
    }
}
