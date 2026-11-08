//! fidget-wgpu in the browser: an omega tape rendered by fidget's all-GPU voxel
//! pipeline into a G-buffer (wgpu-gbuf), on the page's WebGPU, for the page to
//! light.
//!
//!   await init();                                  // adapter, device, pipelines
//!   const info = load(tapeBytes, palette);         // {outputs, names, bbox, hasColor, hasMaterial}
//!   const g = await render(w, h, depth, worldToModel);
//!
//! palette: 3 floats (sRGB, 0-1) per return, the colour of a return the tape
//! gives none. g, per pixel from the top left: index (Uint16Array, the return
//! hit, 1-based, 0 for a miss), depth (Uint32Array, voxel depth), normal
//! (Float32Array, 3: the hit return's gradient in fidget's screen space),
//! color (Uint8Array, 4: sRGB), material (Float32Array, 12: the record, rows
//! flattened), occlusion (Float32Array: SSAO, 1 for none).
//!
//! worldToModel is a column-major 4x4 taking fidget's screen cube [-1, 1]^3
//! (x right, y up, z toward the viewer) to world space; its bottom row may be
//! projective, so a perspective camera's inverse projection-view works.
//!
//! One worker, one caller at a time: the state is taken for the length of an
//! await and put back after, so an overlapping call fails instead of racing.

use std::cell::RefCell;

use fidget_core::render::VoxelSize;
use fidget_raster::voxel::RenderConfig;
use fidget_wgpu::Gpu;
use prost::Message;
use tape_core::tape_proto;
use wasm_bindgen::prelude::*;
use wgpu_gbuf::{Renderer, Scene};

struct State {
    renderer: Renderer,
    scene: Option<Scene>,
}

thread_local! {
    static STATE: RefCell<Option<State>> = const { RefCell::new(None) };
}

fn take() -> Result<State, JsError> {
    STATE
        .with(|s| s.borrow_mut().take())
        .ok_or_else(|| JsError::new("fidget-wgpu is not initialized, or a call is already in flight"))
}

fn put(s: State) {
    STATE.with(|c| *c.borrow_mut() = Some(s));
}

fn set(obj: &js_sys::Object, key: &str, value: JsValue) -> Result<(), JsError> {
    js_sys::Reflect::set(obj, &key.into(), &value).map(|_| ()).map_err(|_| JsError::new(key))
}

/// Requests a WebGPU adapter and device and builds fidget's pipelines.
#[wasm_bindgen]
pub async fn init() -> Result<(), JsError> {
    let gpu = Gpu::init().await.map_err(|e| JsError::new(&format!("WebGPU: {e}")))?;
    put(State { renderer: Renderer::new(&gpu), scene: None });
    Ok(())
}

/// Replays an omega tape into the scene the next renders draw, its inputs at
/// their defaults. A tape with mesh-oracle leaves (bvhSdf, vdbSdf, the noise
/// tile) is refused: fidget's GPU path compiles each field to bytecode, which
/// has no encoding for a host callback.
#[wasm_bindgen]
pub fn load(tape: &[u8], palette: &[f32]) -> Result<JsValue, JsError> {
    let mut state = take()?;
    let result = load_into(&mut state, tape, palette);
    put(state);
    result
}

fn load_into(state: &mut State, bytes: &[u8], palette: &[f32]) -> Result<JsValue, JsError> {
    state.scene = None;
    let tape = tape_proto::Tape::decode(bytes).map_err(|e| JsError::new(&format!("tape: {e}")))?;
    let oracles = tape
        .nodes
        .iter()
        .filter(|n| matches!(n.n, Some(tape_proto::node::N::MeshOracle(_))))
        .count();
    if oracles > 0 {
        return Err(JsError::new(&format!(
            "the tape has {oracles} mesh-oracle leaves, which fidget's GPU pipeline cannot run"
        )));
    }
    let decoded = tape_core::decode(&tape, &mut |_| {
        Err("an embedded mesh cannot run on fidget's GPU pipeline".to_string())
    })
    .map_err(|e| JsError::new(&e))?;

    let info = js_sys::Object::new();
    set(&info, "outputs", (decoded.outputs.len() as u32).into())?;
    let names = js_sys::Array::new();
    for n in &decoded.names {
        names.push(&n.into());
    }
    set(&info, "names", names.into())?;
    let bbox = match decoded.bbox {
        Some(b) => js_sys::Float32Array::from(&b[..]).into(),
        None => JsValue::NULL,
    };
    set(&info, "bbox", bbox)?;

    let palette: Vec<[f32; 3]> = palette.chunks_exact(3).map(|c| [c[0], c[1], c[2]]).collect();
    let scene = Scene::new(decoded, &palette).map_err(|e| JsError::new(&e))?;
    let flags = |v: &[bool]| -> JsValue {
        let a = js_sys::Array::new();
        for &b in v {
            a.push(&b.into());
        }
        a.into()
    };
    set(&info, "hasColor", flags(&scene.has_color))?;
    set(&info, "hasMaterial", flags(&scene.has_material))?;
    state.scene = Some(scene);
    Ok(info.into())
}

/// Renders the loaded scene into a width x height G-buffer, `depth` voxels deep.
#[wasm_bindgen]
pub async fn render(width: u32, height: u32, depth: u32, world_to_model: Vec<f32>) -> Result<JsValue, JsError> {
    let mut state = take()?;
    let result = render_with(&mut state, width, height, depth, &world_to_model).await;
    put(state);
    result
}

async fn render_with(state: &mut State, width: u32, height: u32, depth: u32, m: &[f32]) -> Result<JsValue, JsError> {
    if m.len() != 16 {
        return Err(JsError::new("worldToModel must be 16 floats, column-major"));
    }
    let scene = state.scene.as_ref().ok_or_else(|| JsError::new("no tape loaded"))?;
    let cfg = RenderConfig {
        image_size: VoxelSize::new(width, height, depth),
        world_to_model: nalgebra::Matrix4::from_column_slice(m),
    };
    let g = state.renderer.render(scene, cfg).await.map_err(|e| JsError::new(&e))?;
    let out = js_sys::Object::new();
    set(&out, "index", js_sys::Uint16Array::from(&g.index[..]).into())?;
    set(&out, "depth", js_sys::Uint32Array::from(&g.depth[..]).into())?;
    set(&out, "normal", js_sys::Float32Array::from(&g.normal[..]).into())?;
    set(&out, "color", js_sys::Uint8Array::from(&g.color[..]).into())?;
    set(&out, "material", js_sys::Float32Array::from(&g.material[..]).into())?;
    set(&out, "occlusion", js_sys::Float32Array::from(&g.occlusion[..]).into())?;
    Ok(out.into())
}
