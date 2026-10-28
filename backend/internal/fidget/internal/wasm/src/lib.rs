//! fidget in the browser: load an omega tape, evaluate its fields at points
//! or over boxes, mesh them by dual contouring, and march the Simple render
//! tier's rays through them. Built for
//! wasm32-unknown-unknown by tool/build-fidget-wasm and driven from
//! common/web/fidget-wasm-worker.js.
//!
//! The interface is plain exported functions over linear memory, no
//! generated glue. The host copies inputs in with `fw_alloc` (freeing them
//! with `fw_free`), calls, and reads results back through the pointer and
//! length getters. A call returns 0 on success; anything else leaves a
//! message at `fw_error_ptr` / `fw_error_len`. One tape is loaded at a time;
//! a worker holds one module.
//!
//! An `output` of -1 means the whole scene (the min of the returns, as the
//! mesher sees it); 0.. is one return.
//!
//! Oracles: the noise tile is built in (tape-core). A vdbSdf resolves when
//! the tape carries the level set the server baked (tape-core's level_set);
//! any other source mesh (a bvhSdf, or a vdbSdf without it) is refused, since
//! that needs render's native BVH or OpenVDB.

use std::cell::RefCell;
use std::sync::atomic::{AtomicU64, Ordering};

use fidget_core::context::{Context, Tree};
use fidget_core::shape::{ShapeBulkEval, ShapeTracingEval};
use fidget_core::types::Interval;
use fidget_core::vm::{VmFloatSliceEval, VmIntervalEval, VmShape};
use fidget_mesh::{Octree, Settings as MeshSettings};
use prost::Message;
use tape_core::{tape_proto, Decoded};

const N: usize = u8::MAX as usize;

// ============ Randomness ==============================================

static RNG_STATE: AtomicU64 = AtomicU64::new(0x9E37_79B9_7F4A_7C15);

/// getrandom's custom backend (see .cargo/config.toml). fidget draws random
/// bits only to give each variable a distinct 64-bit name, so a splitmix64
/// stream is enough: distinct within the process is all that is needed.
#[no_mangle]
unsafe extern "Rust" fn __getrandom_v03_custom(dest: *mut u8, len: usize) -> Result<(), getrandom::Error> {
    for i in 0..len {
        let mut z = RNG_STATE.fetch_add(0x9E37_79B9_7F4A_7C15, Ordering::Relaxed);
        z = (z ^ (z >> 30)).wrapping_mul(0xBF58_476D_1CE4_E5B9);
        z = (z ^ (z >> 27)).wrapping_mul(0x94D0_49BB_1331_11EB);
        unsafe { *dest.add(i) = (z ^ (z >> 31)) as u8 };
    }
    Ok(())
}

// ============ State ===================================================

thread_local! {
    static TAPE: RefCell<Option<Decoded>> = const { RefCell::new(None) };
    static ERROR: RefCell<String> = const { RefCell::new(String::new()) };
    static RESULT: RefCell<Vec<u8>> = const { RefCell::new(Vec::new()) };
    // The Simple tier's march field for the last set of visible returns.
    static MARCH: RefCell<Option<(Vec<usize>, VmShape)>> = const { RefCell::new(None) };
}

fn fail(msg: impl Into<String>) -> i32 {
    ERROR.with(|e| *e.borrow_mut() = msg.into());
    1
}

fn status(r: Result<(), String>) -> i32 {
    match r {
        Ok(()) => 0,
        Err(e) => fail(e),
    }
}

fn with_shape<T>(output: i32, f: impl FnOnce(&VmShape, &Decoded) -> Result<T, String>) -> Result<T, String> {
    TAPE.with(|t| {
        let t = t.borrow();
        let tape = t.as_ref().ok_or("no tape loaded")?;
        let shape = if output < 0 {
            &tape.combined
        } else {
            tape.outputs
                .get(output as usize)
                .ok_or_else(|| format!("output {} out of range (have {})", output, tape.outputs.len()))?
        };
        f(shape, tape)
    })
}

// ============ Memory and errors =======================================

/// A buffer of `len` bytes for the host to fill.
#[no_mangle]
pub extern "C" fn fw_alloc(len: usize) -> *mut u8 {
    let mut v = Vec::<u8>::with_capacity(len.max(1));
    let p = v.as_mut_ptr();
    std::mem::forget(v);
    p
}

/// Frees a buffer from fw_alloc.
///
/// # Safety
/// `ptr` and `len` must be exactly what fw_alloc was given and returned.
#[no_mangle]
pub unsafe extern "C" fn fw_free(ptr: *mut u8, len: usize) {
    drop(unsafe { Vec::from_raw_parts(ptr, 0, len.max(1)) });
}

#[no_mangle]
pub extern "C" fn fw_error_ptr() -> *const u8 {
    ERROR.with(|e| e.borrow().as_ptr())
}

#[no_mangle]
pub extern "C" fn fw_error_len() -> usize {
    ERROR.with(|e| e.borrow().len())
}

/// The bytes the last fw_mesh_dc produced (a binary STL).
#[no_mangle]
pub extern "C" fn fw_result_ptr() -> *const u8 {
    RESULT.with(|r| r.borrow().as_ptr())
}

#[no_mangle]
pub extern "C" fn fw_result_len() -> usize {
    RESULT.with(|r| r.borrow().len())
}

// ============ Tape ====================================================

/// Replays a tape (the bytes of canvas/tape-result) and keeps it.
///
/// # Safety
/// `ptr` must point to `len` readable bytes.
#[no_mangle]
pub unsafe extern "C" fn fw_load_tape(ptr: *const u8, len: usize) -> i32 {
    let bytes = unsafe { std::slice::from_raw_parts(ptr, len) };
    status((|| {
        let tape = tape_proto::Tape::decode(bytes).map_err(|e| format!("decode: {:?}", e))?;
        let decoded = tape_core::decode(&tape, &mut |o| {
            // A vdbSdf whose tape carries the server's level set samples it;
            // anything else would need render's BVH or OpenVDB.
            if tape_core::level_set::register(o)? {
                return Ok(());
            }
            Err("the tape embeds a source mesh (bvhSdf, or a vdbSdf without its level set), which the browser cannot rebuild".into())
        })?;
        TAPE.with(|t| *t.borrow_mut() = Some(decoded));
        MARCH.with(|m| *m.borrow_mut() = None);
        Ok(())
    })())
}

/// The number of returns in the loaded tape, or 0 when none is loaded.
#[no_mangle]
pub extern "C" fn fw_output_count() -> u32 {
    TAPE.with(|t| t.borrow().as_ref().map_or(0, |t| t.outputs.len() as u32))
}

/// Writes the tape's recorded scene box to `out[0..6]` and returns 1, or
/// returns 0 when the tape recorded none.
///
/// # Safety
/// `out` must point to 6 writable f32s.
#[no_mangle]
pub unsafe extern "C" fn fw_bbox(out: *mut f32) -> i32 {
    TAPE.with(|t| match t.borrow().as_ref().and_then(|t| t.bbox) {
        Some(b) => {
            for (i, v) in b.iter().enumerate() {
                unsafe { *out.add(i) = *v };
            }
            1
        }
        None => 0,
    })
}

// ============ Evaluation ==============================================

/// The field at `n` points (`xyz`, interleaved) into `out[0..n]`.
///
/// # Safety
/// `xyz` must point to `3n` readable f32s and `out` to `n` writable ones.
#[no_mangle]
pub unsafe extern "C" fn fw_eval_points(output: i32, xyz: *const f32, n: usize, out: *mut f32) -> i32 {
    let pts = unsafe { std::slice::from_raw_parts(xyz, 3 * n) };
    let dst = unsafe { std::slice::from_raw_parts_mut(out, n) };
    status(with_shape(output, |shape, tape| {
        let xs: Vec<f32> = (0..n).map(|i| pts[3 * i]).collect();
        let ys: Vec<f32> = (0..n).map(|i| pts[3 * i + 1]).collect();
        let zs: Vec<f32> = (0..n).map(|i| pts[3 * i + 2]).collect();
        let t = shape.float_slice_tape(Default::default());
        let mut eval = ShapeBulkEval::<VmFloatSliceEval<N>>::default();
        let values = eval
            .eval_with_vars(&t, &xs, &ys, &zs, &tape.shape_vars)
            .map_err(|e| format!("{:?}", e))?;
        dst.copy_from_slice(&values[..n]);
        Ok(())
    }))
}

/// A bound on the field over the box `b[0..6]` (xmin, ymin, zmin, xmax,
/// ymax, zmax) into `out[0..2]` (lower, upper).
///
/// # Safety
/// `b` must point to 6 readable f32s and `out` to 2 writable ones.
#[no_mangle]
pub unsafe extern "C" fn fw_eval_interval(output: i32, b: *const f32, out: *mut f32) -> i32 {
    let b = unsafe { std::slice::from_raw_parts(b, 6) };
    status(with_shape(output, |shape, tape| {
        let t = shape.interval_tape(Default::default());
        let mut eval = ShapeTracingEval::<VmIntervalEval<N>>::default();
        let (v, _) = eval
            .eval_with_vars(
                &t,
                Interval::new(b[0], b[3]),
                Interval::new(b[1], b[4]),
                Interval::new(b[2], b[5]),
                &tape.shape_vars,
            )
            .map_err(|e| format!("{:?}", e))?;
        unsafe {
            *out = v.lower();
            *out.add(1) = v.upper();
        }
        Ok(())
    }))
}

// ============ Meshing =================================================

/// Maps the box onto fidget's [-1, 1] model cube, each axis on its own, as
/// render's world_to_model_from_bbox does.
fn world_to_model(b: &[f32]) -> nalgebra::Matrix4<f32> {
    let mut m = nalgebra::Matrix4::<f32>::identity();
    for k in 0..3 {
        let half = (0.5 * (b[k + 3] - b[k])).max(f32::EPSILON);
        m[(k, k)] = half;
        m[(k, 3)] = 0.5 * (b[k] + b[k + 3]);
    }
    m
}

/// Dual-contours the field over the box `b[0..6]` at octree `depth` (2^depth
/// cells per axis) and keeps the mesh as a binary STL (fw_result_ptr).
/// Single-threaded: wasm has no threads.
///
/// # Safety
/// `b` must point to 6 readable f32s.
#[no_mangle]
pub unsafe extern "C" fn fw_mesh_dc(output: i32, depth: u32, b: *const f32) -> i32 {
    let b = unsafe { std::slice::from_raw_parts(b, 6) };
    status(with_shape(output, |shape, tape| {
        let settings = MeshSettings {
            depth: depth.min(u8::MAX as u32) as u8,
            world_to_model: world_to_model(b),
            threads: None,
            ..Default::default()
        };
        let bound = shape.bind(&tape.shape_vars).map_err(|e| e.to_string())?;
        let octree = Octree::build(&bound, &settings);
        let octree = octree.ok_or("octree build failed")?;
        let mesh = octree.walk_dual();
        let mut stl = Vec::new();
        mesh.write_stl(&mut stl).map_err(|e| format!("{:?}", e))?;
        RESULT.with(|r| *r.borrow_mut() = stl);
        Ok(())
    }))
}

// ============ Simple render ===========================================

fn to_shape(tree: &Tree) -> Result<VmShape, String> {
    let mut ctx = Context::new();
    let node = ctx.import(tree);
    VmShape::new(&ctx, node).map_err(|e| format!("VmShape::new failed: {:?}", e))
}

/// WGSL's mix.
fn mix(a: f32, b: f32, t: f32) -> f32 {
    a * (1.0 - t) + b * t
}

/// Marches rows [row0, row0 + rows) of a width x height frame by the Simple
/// render tier's spec (common/web/gsl-simple.js), and keeps, per pixel, five
/// f32s at fw_result_ptr: the 1-based index of the return hit, or 0 for a
/// miss; t; and the normal.
///
///   cam[0..17]      the camera position, right, up and forward (3 each), the
///                   tangents of the half-angles (2), 1 for orthographic,
///                   the pixel scale, and the scene box's half-extent
///   visible[0..k]   1 for each of the k returns that shows
///   spec[0..6]      steps, fudge, hit floor, hit pixels, normal floor,
///                   normal pixels
///
/// Each step evaluates every live ray at once through one field, the min of
/// |d| over the visible returns; only the hits evaluate each return, to
/// find the winner and the normal.
///
/// # Safety
/// `cam` must point to 17 readable f32s, `visible` to one per return, and
/// `spec` to 6.
#[no_mangle]
pub unsafe extern "C" fn fw_render_simple(
    cam: *const f32,
    visible: *const f32,
    spec: *const f32,
    width: u32,
    height: u32,
    row0: u32,
    rows: u32,
) -> i32 {
    let c = unsafe { std::slice::from_raw_parts(cam, 17) };
    let spec = unsafe { std::slice::from_raw_parts(spec, 6) };
    status(TAPE.with(|t| {
        let t = t.borrow();
        let tape = t.as_ref().ok_or("no tape loaded")?;
        let visible = unsafe { std::slice::from_raw_parts(visible, tape.outputs.len()) };
        let vis: Vec<usize> = (0..tape.outputs.len()).filter(|&i| visible[i] > 0.5).collect();

        let (w, n) = (width as usize, width as usize * rows as usize);
        let mut out = vec![0f32; 5 * n];
        if vis.is_empty() {
            RESULT.with(|r| *r.borrow_mut() = out.iter().flat_map(|v| v.to_le_bytes()).collect());
            return Ok(());
        }

        let pos = [c[0], c[1], c[2]];
        let right = [c[3], c[4], c[5]];
        let up = [c[6], c[7], c[8]];
        let fwd = [c[9], c[10], c[11]];
        let (tan_x, tan_y, ortho, pixel_scale, extent) = (c[12], c[13], c[14], c[15], c[16]);
        let (steps, fudge, hit_floor, hit_pixel, normal_floor, normal_pixel) =
            (spec[0] as usize, spec[1], spec[2], spec[3], spec[4], spec[5]);
        let pixel_world = |t: f32| pixel_scale * mix(t.max(1.0), 1.0, ortho);

        // The rays, and where each enters and leaves the scene box.
        let mut origin = vec![[0f32; 3]; n];
        let mut dir = vec![[0f32; 3]; n];
        let mut ts = vec![0f32; n];
        let mut t_exit = vec![0f32; n];
        let mut live = Vec::with_capacity(n);
        for i in 0..n {
            let (x, y) = (i % w, row0 as usize + i / w);
            let ndc_x = ((x as f32 + 0.5) / width as f32) * 2.0 - 1.0;
            let ndc_y = 1.0 - ((y as f32 + 0.5) / height as f32) * 2.0;
            let mut d = [0f32; 3];
            for k in 0..3 {
                let off = ndc_x * tan_x * right[k] + ndc_y * tan_y * up[k];
                origin[i][k] = pos[k] + off * ortho;
                d[k] = mix(fwd[k] + off, fwd[k], ortho);
            }
            let len = (d[0] * d[0] + d[1] * d[1] + d[2] * d[2]).sqrt();
            dir[i] = [d[0] / len, d[1] / len, d[2] / len];
            let (mut entry, mut exit) = (f32::NEG_INFINITY, f32::INFINITY);
            for k in 0..3 {
                let inv = 1.0 / dir[i][k];
                let t1 = (-extent - origin[i][k]) * inv;
                let t2 = (extent - origin[i][k]) * inv;
                entry = entry.max(t1.min(t2));
                exit = exit.min(t1.max(t2));
            }
            if entry <= exit && exit > 0.0 {
                ts[i] = entry.max(0.0);
                t_exit[i] = exit;
                live.push(i);
            }
        }

        // The march field, rebuilt only when the visible set changes.
        let march = MARCH.with(|m| -> Result<VmShape, String> {
            let mut m = m.borrow_mut();
            if let Some((v, shape)) = m.as_ref() {
                if *v == vis {
                    return Ok(shape.clone());
                }
            }
            let mut tree = tape.output_trees[vis[0]].abs();
            for &i in &vis[1..] {
                tree = tree.min(tape.output_trees[i].abs());
            }
            let shape = to_shape(&tree)?;
            *m = Some((vis.clone(), shape.clone()));
            Ok(shape)
        })?;
        let march_tape = march.float_slice_tape(Default::default());
        let mut eval = ShapeBulkEval::<VmFloatSliceEval<N>>::default();

        let (mut xs, mut ys, mut zs) = (Vec::new(), Vec::new(), Vec::new());
        let mut hits = Vec::new();
        for _ in 0..steps {
            live.retain(|&i| !(ts[i] > t_exit[i]));
            if live.is_empty() {
                break;
            }
            xs.clear();
            ys.clear();
            zs.clear();
            for &i in &live {
                xs.push(origin[i][0] + dir[i][0] * ts[i]);
                ys.push(origin[i][1] + dir[i][1] * ts[i]);
                zs.push(origin[i][2] + dir[i][2] * ts[i]);
            }
            let values = eval
                .eval_with_vars(&march_tape, &xs, &ys, &zs, &tape.shape_vars)
                .map_err(|e| format!("{:?}", e))?;
            let mut next = Vec::with_capacity(live.len());
            for (j, &i) in live.iter().enumerate() {
                let d = values[j];
                if d < hit_floor.max(pixel_world(ts[i]) * hit_pixel) {
                    hits.push(i);
                } else {
                    ts[i] += d * fudge;
                    next.push(i);
                }
            }
            live = next;
        }

        // At each hit: the point and its four tetrahedral neighbours, through
        // every visible return.
        const K: [[f32; 3]; 4] = [[1.0, -1.0, -1.0], [-1.0, -1.0, 1.0], [-1.0, 1.0, -1.0], [1.0, 1.0, 1.0]];
        xs.clear();
        ys.clear();
        zs.clear();
        for &i in &hits {
            let p = [0, 1, 2].map(|k| origin[i][k] + dir[i][k] * ts[i]);
            let e = normal_floor.max(pixel_world(ts[i]) * normal_pixel);
            xs.push(p[0]);
            ys.push(p[1]);
            zs.push(p[2]);
            for k in &K {
                xs.push(p[0] + k[0] * e);
                ys.push(p[1] + k[1] * e);
                zs.push(p[2] + k[2] * e);
            }
        }
        // gsl_min's composite: the smallest |d|, the later return on a tie.
        let mut best = vec![(0f32, 0usize); xs.len()];
        for (m, &o) in vis.iter().enumerate() {
            let t = tape.outputs[o].float_slice_tape(Default::default());
            let values = eval
                .eval_with_vars(&t, &xs, &ys, &zs, &tape.shape_vars)
                .map_err(|e| format!("{:?}", e))?;
            for (b, &d) in best.iter_mut().zip(values.iter()) {
                if m == 0 || !(b.0.abs() < d.abs()) {
                    *b = (d, o + 1);
                }
            }
        }
        for (h, &i) in hits.iter().enumerate() {
            let s = &best[5 * h..5 * h + 5];
            let mut g = [0f32; 3];
            for (j, k) in K.iter().enumerate() {
                for a in 0..3 {
                    g[a] += k[a] * s[1 + j].0;
                }
            }
            let len = (g[0] * g[0] + g[1] * g[1] + g[2] * g[2]).sqrt();
            out[5 * i] = s[0].1 as f32;
            out[5 * i + 1] = ts[i];
            out[5 * i + 2] = g[0] / len;
            out[5 * i + 3] = g[1] / len;
            out[5 * i + 4] = g[2] / len;
        }

        RESULT.with(|r| *r.borrow_mut() = out.iter().flat_map(|v| v.to_le_bytes()).collect());
        Ok(())
    }))
}
