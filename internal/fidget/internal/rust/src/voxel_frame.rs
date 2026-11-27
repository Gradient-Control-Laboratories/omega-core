//! fidget's CPU voxel render, for the lit render's fidget device (eval's
//! litfidget.go): each return voxel-rendered on its own into depth and normal,
//! and fidget's SSAO over the merged image. The browser's Fidget device runs
//! the same two steps on the GPU (fidget-wgpu, internal/wgpu-gbuf); the caller
//! merges the returns, keeping the nearest hit, between them.

use crate::{CANVASES, INPUT_VALUES, TREES};
use core::ffi::c_ulong;
use fidget::{
    eval::Function,
    raster::{
        effects,
        voxel::{GeometryPixel, Image, RenderConfig},
    },
    render::{RenderHints, VoxelSize},
    shape::{Shape, ShapeVars},
    var::Var,
    vm::VmShape,
};
use nalgebra::Matrix4;
use std::os::raw::c_int;

/// canvas_render_voxels voxel-renders one canvas at width x height pixels,
/// depth voxels deep, through `mat`, fidget's world-to-model transform (16
/// floats, column-major). It writes each pixel's voxel depth (0 for a miss) to
/// `out_depth` (width * height) and the field's gradient in fidget's pixel
/// coordinates (x right, y down, z toward the viewer) to `out_normal` (3 per
/// pixel), rows top first. A pixel full up to the camera has depth `depth` and
/// normal (0, 0, 1).
///
/// The document's inputs evaluate at their current values, except the nvars
/// listed in var_trees (tree indices), which take var_values (one each, for
/// the whole frame): the renderer's pixel footprint, which pixelSize() reads.
///
/// Returns 0, or: -60 no such canvas, -61 the shape does not build, -62 an
/// input is unbound, -63 a panic.
#[no_mangle]
pub extern "C" fn canvas_render_voxels(
    canvas_index: c_ulong,
    width: u32,
    height: u32,
    depth: u32,
    mat: *const f32,
    nvars: c_ulong,
    var_trees: *const c_ulong,
    var_values: *const f32,
    out_depth: *mut u32,
    out_normal: *mut f32,
) -> c_int {
    match std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
        let mat = unsafe { std::slice::from_raw_parts(mat, 16) };
        let (var_trees, var_values) = if nvars == 0 {
            (&[][..], &[][..])
        } else {
            unsafe {
                (
                    std::slice::from_raw_parts(var_trees, nvars as usize),
                    std::slice::from_raw_parts(var_values, nvars as usize),
                )
            }
        };
        let px = (width as usize) * (height as usize);
        let out_depth = unsafe { std::slice::from_raw_parts_mut(out_depth, px) };
        let out_normal = unsafe { std::slice::from_raw_parts_mut(out_normal, 3 * px) };
        render_voxels(canvas_index, width, height, depth, mat, var_trees, var_values, out_depth, out_normal)
    })) {
        Ok(code) => code,
        Err(_) => -63,
    }
}

#[allow(clippy::too_many_arguments)]
fn render_voxels(
    canvas_index: c_ulong,
    width: u32,
    height: u32,
    depth: u32,
    mat: &[f32],
    var_trees: &[c_ulong],
    var_values: &[f32],
    out_depth: &mut [u32],
    out_normal: &mut [f32],
) -> c_int {
    let canvases = CANVASES.lock().unwrap_or_else(std::sync::PoisonError::into_inner);
    let canvas = match canvases.get(canvas_index as usize) {
        Some(c) => c,
        None => return -60,
    };
    let (ctx, node) = (&canvas.ctx, canvas.node);

    // Every input at its value, then the listed ones at theirs.
    let trees = TREES.lock().unwrap_or_else(std::sync::PoisonError::into_inner);
    let input_values = INPUT_VALUES.lock().unwrap_or_else(std::sync::PoisonError::into_inner);
    let mut vars = ShapeVars::<f32>::new();
    for (tree_index, value) in input_values.iter() {
        if let Some(index) = trees[*tree_index].var().and_then(|v| v.index()) {
            vars.insert(index, *value as f32);
        }
    }
    for (tree_index, value) in var_trees.iter().zip(var_values) {
        match trees.get(*tree_index as usize).and_then(|t| t.var()).and_then(|v| v.index()) {
            Some(index) => {
                vars.insert(index, *value);
            }
            None => return -62,
        }
    }
    drop(input_values);
    drop(trees);

    let cfg = RenderConfig {
        image_size: VoxelSize::new(width, height, depth),
        world_to_model: Matrix4::from_column_slice(mat),
    };
    // A canvas holding a mesh oracle cannot be JIT-compiled; it renders on the
    // VM, as the meshing paths do.
    let image = if ctx.contains_mesh_oracle(node) {
        match VmShape::new(ctx, node) {
            Ok(shape) => render_shape(shape, &vars, &cfg),
            Err(_) => return -61,
        }
    } else {
        match fidget::jit::JitShape::new(ctx, node) {
            Ok(shape) => render_shape(shape, &vars, &cfg),
            Err(_) => return -61,
        }
    };
    let image = match image {
        Some(image) => image,
        None => return -62,
    };
    for (i, p) in image.iter().enumerate() {
        out_depth[i] = p.depth;
        out_normal[3 * i..3 * i + 3].copy_from_slice(&p.normal);
    }
    0
}

/// render_shape binds the shape's own inputs (an extra one is refused by the
/// evaluator) and renders it; None when one of its inputs has no value.
fn render_shape<F: Function + RenderHints>(
    shape: Shape<F>,
    vars: &ShapeVars<f32>,
    cfg: &RenderConfig,
) -> Option<Image> {
    let mut own = ShapeVars::<f32>::new();
    let needed = shape.inner().vars();
    for (index, value) in vars {
        if needed.get(&Var::V(*index)).is_some() {
            own.insert(*index, *value);
        }
    }
    let bound = shape.bind(&own).ok()?;
    Some(cfg.run(bound))
}

/// voxel_ssao is fidget's screen-space ambient occlusion over a merged frame:
/// `depth_in` and `normal_in` as canvas_render_voxels writes them, `depth`
/// the voxel depth they were rendered at. It writes the blurred occlusion to
/// `out` (width * height): 1 for none, NaN on a miss.
///
/// Returns 0, or -64 for a panic.
#[no_mangle]
pub extern "C" fn voxel_ssao(
    width: u32,
    height: u32,
    depth: u32,
    depth_in: *const u32,
    normal_in: *const f32,
    out: *mut f32,
) -> c_int {
    match std::panic::catch_unwind(std::panic::AssertUnwindSafe(|| {
        let px = (width as usize) * (height as usize);
        let depth_in = unsafe { std::slice::from_raw_parts(depth_in, px) };
        let normal_in = unsafe { std::slice::from_raw_parts(normal_in, 3 * px) };
        let out = unsafe { std::slice::from_raw_parts_mut(out, px) };
        let pixels: Vec<GeometryPixel> = (0..px)
            .map(|i| GeometryPixel {
                depth: depth_in[i],
                normal: [normal_in[3 * i], normal_in[3 * i + 1], normal_in[3 * i + 2]],
            })
            .collect();
        let image = Image::build(pixels, VoxelSize::new(width, height, depth)).expect("pixel count");
        let threads = Some(&fidget::render::ThreadPool::Global);
        let ssao = effects::blur_ssao(&effects::compute_ssao(&image, threads), threads);
        out.copy_from_slice(ssao.as_slice());
        0
    })) {
        Ok(code) => code,
        Err(_) => -64,
    }
}
