//! wgpu-render: an omega tape through fidget-wgpu, and through fidget's CPU
//! raster for comparison.
//!
//!   wgpu-render -tape part.tape -out DIR [-size 768] [-yaw 30] [-pitch 25] [-ssao]
//!
//! Writes DIR/gpu.png, DIR/cpu.png and DIR/diff.png, and prints timings and a
//! geometry comparison: pixels one side hits and the other misses, and the
//! depth and normal error where both hit.

use std::time::Instant;

use fidget_core::{
    render::{ThreadPool, VoxelSize},
    shape::ShapeVars,
    vm::VmShape,
};
use fidget_raster::voxel::{Image as VoxelImage, RenderConfig};
use nalgebra::{Matrix4, Rotation3, Vector3};
use prost::Message;
use zerocopy::IntoBytes;
use tape_core::tape_proto;

struct Args {
    tape: String,
    out: String,
    size: u32,
    yaw: f32,
    pitch: f32,
    ssao: bool,
    gbuf: bool,
}

fn parse_args() -> Result<Args, String> {
    let mut a = Args {
        tape: String::new(),
        out: String::new(),
        size: 768,
        yaw: 30.0,
        pitch: 25.0,
        ssao: false,
        gbuf: false,
    };
    let mut it = std::env::args().skip(1);
    while let Some(flag) = it.next() {
        let mut val = || it.next().ok_or(format!("{} needs a value", flag));
        match flag.as_str() {
            "-tape" => a.tape = val()?,
            "-out" => a.out = val()?,
            "-size" => a.size = val()?.parse().map_err(|e| format!("-size: {e}"))?,
            "-yaw" => a.yaw = val()?.parse().map_err(|e| format!("-yaw: {e}"))?,
            "-pitch" => a.pitch = val()?.parse().map_err(|e| format!("-pitch: {e}"))?,
            "-ssao" => a.ssao = true,
            "-gbuf" => a.gbuf = true,
            _ => return Err(format!("unknown flag {flag}")),
        }
    }
    if a.tape.is_empty() || a.out.is_empty() {
        return Err("usage: wgpu-render -tape FILE -out DIR [-size N] [-yaw DEG] [-pitch DEG] [-ssao]".into());
    }
    Ok(a)
}

/// The voxel grid spans [-1, 1]^3; this maps it onto a cube around the scene
/// box, turned by yaw about Z and then pitch about X, sized so the box fits
/// in any orientation.
fn world_to_model(b: [f32; 6], yaw: f32, pitch: f32) -> Matrix4<f32> {
    let center = Vector3::new(0.5 * (b[0] + b[3]), 0.5 * (b[1] + b[4]), 0.5 * (b[2] + b[5]));
    let half = Vector3::new(0.5 * (b[3] - b[0]), 0.5 * (b[4] - b[1]), 0.5 * (b[5] - b[2]));
    let radius = half.norm().max(f32::EPSILON);
    // Look down -Y with Z up: the grid's screen Y is model Z.
    let view = Rotation3::from_axis_angle(&Vector3::x_axis(), std::f32::consts::FRAC_PI_2);
    let rot = Rotation3::from_axis_angle(&Vector3::z_axis(), yaw.to_radians())
        * Rotation3::from_axis_angle(&Vector3::x_axis(), pitch.to_radians())
        * view;
    Matrix4::new_translation(&center) * rot.to_homogeneous() * Matrix4::new_scaling(radius)
}

fn write_png(path: &str, size: u32, rgba: &[u8]) -> Result<(), String> {
    image::save_buffer(path, rgba, size, size, image::ExtendedColorType::Rgba8)
        .map_err(|e| format!("{path}: {e}"))
}

fn main() {
    if let Err(e) = run() {
        eprintln!("wgpu-render: {e}");
        std::process::exit(1);
    }
}

fn run() -> Result<(), String> {
    let args = parse_args()?;
    std::fs::create_dir_all(&args.out).map_err(|e| format!("{}: {e}", args.out))?;

    let bytes = std::fs::read(&args.tape).map_err(|e| format!("{}: {e}", args.tape))?;
    let tape = tape_proto::Tape::decode(bytes.as_slice()).map_err(|e| format!("decode: {e}"))?;
    // fidget-wgpu compiles the shape to bytecode, which has no encoding for
    // the host-callback oracle (the fork panics on it); say so up front.
    let oracles = tape
        .nodes
        .iter()
        .filter(|n| matches!(n.n, Some(tape_proto::node::N::MeshOracle(_))))
        .count();
    let decoded = tape_core::decode(&tape, &mut |_| {
        Err("mesh oracles are not supported by wgpu-render".to_string())
    })?;
    let bbox = decoded.bbox.ok_or("the tape records no scene box")?;
    if args.gbuf {
        return gbuf(decoded, bbox, &args);
    }
    let shape: &VmShape = &decoded.combined;
    let vars: &ShapeVars<f32> = &decoded.shape_vars;

    let n = args.size;
    let cfg = RenderConfig {
        image_size: VoxelSize::new(n, n, n),
        world_to_model: world_to_model(bbox, args.yaw, args.pitch),
    };
    println!("tape: {} nodes, {} outputs, box {:?}, {}^3 voxels", tape.nodes.len(), decoded.outputs.len(), bbox, n);

    // CPU: fidget's raster and effects.
    let t = Instant::now();
    let bound = shape.bind(vars).map_err(|e| format!("bind: {e}"))?;
    let cpu_img: VoxelImage = cfg.run(bound);
    let cpu_geom = t.elapsed();
    let t = Instant::now();
    let threads = Some(&ThreadPool::Global);
    let denoised = fidget_raster::effects::denoise_normals(&cpu_img, threads);
    let shaded = fidget_raster::effects::apply_shading(&denoised, args.ssao, threads);
    let cpu_rgba: Vec<u8> = denoised
        .iter()
        .zip(shaded.iter())
        .flat_map(|(p, c)| if p.depth > 0 { [c[0], c[1], c[2], 255] } else { [0, 0, 0, 0] })
        .collect();
    let cpu_fx = t.elapsed();
    write_png(&format!("{}/cpu.png", args.out), n, &cpu_rgba)?;
    println!("cpu: geometry {:.1} ms, shading {:.1} ms", ms(cpu_geom), ms(cpu_fx));

    if oracles > 0 {
        println!("gpu: skipped, the tape has {oracles} mesh-oracle leaves, which fidget-wgpu cannot run");
        return Ok(());
    }

    // GPU: fidget-wgpu's voxel pipeline and its effects.
    let gpu = pollster::block_on(fidget_wgpu::Gpu::init()).map_err(|e| format!("gpu init: {e}"))?;
    let rshape = fidget_wgpu::RenderShape::new(shape).map_err(|e| format!("render shape: {e}"))?;
    let ctx = fidget_wgpu::voxel::Context::new(&gpu);
    let mut ws = ctx.workspace();
    let mut out = gpu.read_buffer_for(ws.output());
    // The first frame includes pipeline construction; time a second one too.
    let t = Instant::now();
    let gpu_img = ctx
        .run_with_vars(&rshape, vars, &mut ws, &mut out, cfg)
        .map_err(|e| format!("gpu render: {e}"))?;
    let gpu_first = t.elapsed();
    let t = Instant::now();
    let gpu_img2 = ctx
        .run_with_vars(&rshape, vars, &mut ws, &mut out, cfg)
        .map_err(|e| format!("gpu render: {e}"))?;
    let gpu_geom = t.elapsed();
    let _ = gpu_img2;

    let fx = fidget_wgpu::voxel::effects::Context::new(&gpu);
    let mut merge = fx.merge_workspace();
    let mut ssao = fx.ssao_workspace();
    let mut shade = fx.shade_workspace();
    let t = Instant::now();
    fx.submit_merge(
        ws.output(),
        fidget_wgpu::voxel::effects::MergeSettings { denoise: true, z_scale: 1.0 },
        &mut merge,
    )
    .map_err(|e| format!("merge: {e}"))?;
    if args.ssao {
        fx.submit_ssao(&merge, &mut ssao).map_err(|e| format!("ssao: {e}"))?;
    }
    fx.submit_shade(&merge, if args.ssao { Some(&ssao) } else { None }, &mut shade)
        .map_err(|e| format!("shade: {e}"))?;
    let gpu_rgba: Vec<u8> = gpu.read_vec(shade.output()).as_bytes().to_vec();
    let gpu_fx = t.elapsed();
    write_png(&format!("{}/gpu.png", args.out), n, &gpu_rgba)?;
    println!(
        "gpu: geometry {:.1} ms (first frame {:.1} ms), shading {:.1} ms",
        ms(gpu_geom),
        ms(gpu_first),
        ms(gpu_fx)
    );

    compare(&cpu_img, &gpu_img, n, &format!("{}/diff.png", args.out))
}

fn ms(d: std::time::Duration) -> f64 {
    d.as_secs_f64() * 1e3
}

/// Pixels one side hits and the other misses, and depth (voxels) and normal
/// (degrees) error where both hit. diff.png: red where only the CPU hits, blue
/// where only the GPU does, grey by depth error elsewhere.
fn compare(cpu: &VoxelImage, gpu: &VoxelImage, n: u32, path: &str) -> Result<(), String> {
    let (mut both, mut cpu_only, mut gpu_only) = (0usize, 0usize, 0usize);
    let mut depth_err: Vec<i64> = Vec::new();
    let mut angle_err: Vec<f32> = Vec::new();
    let mut rgba = Vec::with_capacity((n * n * 4) as usize);
    for (c, g) in cpu.iter().zip(gpu.iter()) {
        match (c.depth > 0, g.depth > 0) {
            (true, true) => {
                both += 1;
                let dd = g.depth as i64 - c.depth as i64;
                depth_err.push(dd);
                let a = Vector3::from(c.normal);
                let b = Vector3::from(g.normal);
                if a.norm() > 0.0 && b.norm() > 0.0 {
                    let cos = a.normalize().dot(&b.normalize()).clamp(-1.0, 1.0);
                    angle_err.push(cos.acos().to_degrees());
                }
                let v = (255 - (dd.unsigned_abs().min(8) * 30)) as u8;
                rgba.extend_from_slice(&[v, v, v, 255]);
            }
            (true, false) => {
                cpu_only += 1;
                rgba.extend_from_slice(&[255, 0, 0, 255]);
            }
            (false, true) => {
                gpu_only += 1;
                rgba.extend_from_slice(&[0, 0, 255, 255]);
            }
            (false, false) => rgba.extend_from_slice(&[0, 0, 0, 0]),
        }
    }
    write_png(path, n, &rgba)?;
    // The two sides count depth from different origins, so compare each
    // pixel's offset with the most common one rather than with zero.
    let mut counts = std::collections::HashMap::new();
    for &d in &depth_err {
        *counts.entry(d).or_insert(0usize) += 1;
    }
    let (mode, _) = counts.iter().max_by_key(|&(_, n)| *n).map(|(d, n)| (*d, *n)).unwrap_or((0, 0));
    let mut off: Vec<u64> = depth_err.iter().map(|d| (d - mode).unsigned_abs()).collect();
    off.sort_unstable();
    angle_err.sort_by(|a, b| a.partial_cmp(b).unwrap_or(std::cmp::Ordering::Equal));
    let pct = |v: &[u64], q: f64| v.get(((v.len() as f64 - 1.0) * q) as usize).copied().unwrap_or(0);
    let pctf = |v: &[f32], q: f64| v.get(((v.len() as f64 - 1.0) * q) as usize).copied().unwrap_or(0.0);
    let far = off.iter().filter(|&&d| d > 1).count();
    println!(
        "compare: {both} pixels both hit, {cpu_only} only CPU, {gpu_only} only GPU; \
         depth offset gpu-cpu {mode} voxels, off it by >1 on {far} pixels (p99 {} max {}); \
         normal p50 {:.3} p99 {:.3} max {:.3} deg over {} pixels",
        pct(&off, 0.99),
        off.last().copied().unwrap_or(0),
        pctf(&angle_err, 0.5),
        pctf(&angle_err, 0.99),
        angle_err.last().copied().unwrap_or(0.0),
        angle_err.len(),
    );
    Ok(())
}

/// -gbuf: the tape's G-buffer (wgpu-gbuf), written as colour.png (the colour
/// pass), index.png (the return hit, by hue), roughness.png, metallic.png
/// and occlusion.png, with a summary of what each return's pixels hold.
fn gbuf(decoded: tape_core::Decoded, bbox: [f32; 6], args: &Args) -> Result<(), String> {
    let n = args.size;
    let cfg = RenderConfig {
        image_size: VoxelSize::new(n, n, n),
        world_to_model: world_to_model(bbox, args.yaw, args.pitch),
    };
    let palette = [[0.9, 0.3, 0.3], [0.3, 0.8, 0.4], [0.3, 0.5, 0.9], [0.9, 0.8, 0.3], [0.7, 0.4, 0.9]];
    let scene = wgpu_gbuf::Scene::new(decoded, &palette)?;
    println!("returns {}, colour from the tape {:?}, material from the tape {:?}", scene.returns(), scene.has_color, scene.has_material);
    let gpu = pollster::block_on(fidget_wgpu::Gpu::init()).map_err(|e| format!("gpu init: {e}"))?;
    let mut r = wgpu_gbuf::Renderer::new(&gpu);
    let t = Instant::now();
    let _ = pollster::block_on(r.render(&scene, cfg))?;
    let first = t.elapsed();
    let t = Instant::now();
    let g = pollster::block_on(r.render(&scene, cfg))?;
    println!("gbuf: {:.1} ms (first frame {:.1} ms)", ms(t.elapsed()), ms(first));

    let px = (n * n) as usize;
    let hit = |i: usize| g.index[i] > 0;
    let mut colour = vec![0u8; 4 * px];
    let mut index = vec![0u8; 4 * px];
    let mut rough = vec![0u8; 4 * px];
    let mut metal = vec![0u8; 4 * px];
    let mut occl = vec![0u8; 4 * px];
    let hues: [[u8; 3]; 6] = [[230, 80, 80], [80, 200, 100], [80, 120, 230], [230, 200, 80], [180, 100, 230], [80, 210, 210]];
    let grey = |v: f32| -> [u8; 4] { let b = (v.clamp(0.0, 1.0) * 255.0) as u8; [b, b, b, 255] };
    let mut per_return = vec![(0usize, [0f64; 12]); scene.returns()];
    // Per return: normals that are not finite, and ones that are zero.
    let mut bad_normals = vec![(0usize, 0usize); scene.returns()];
    for i in 0..px {
        if !hit(i) {
            continue;
        }
        let k = g.index[i] as usize - 1;
        let nv = &g.normal[3 * i..3 * i + 3];
        if nv.iter().any(|v| !v.is_finite()) {
            bad_normals[k].0 += 1;
        } else if nv.iter().all(|&v| v == 0.0) {
            bad_normals[k].1 += 1;
        }
        colour[4 * i..4 * i + 4].copy_from_slice(&[g.color[4 * i], g.color[4 * i + 1], g.color[4 * i + 2], 255]);
        let h = hues[k % hues.len()];
        index[4 * i..4 * i + 4].copy_from_slice(&[h[0], h[1], h[2], 255]);
        let m = &g.material[12 * i..12 * i + 12];
        rough[4 * i..4 * i + 4].copy_from_slice(&grey(m[0]));
        metal[4 * i..4 * i + 4].copy_from_slice(&grey(m[1]));
        occl[4 * i..4 * i + 4].copy_from_slice(&grey(g.occlusion[i]));
        per_return[k].0 += 1;
        for c in 0..12 {
            per_return[k].1[c] += m[c] as f64;
        }
    }
    for (name, img) in [("colour", &colour), ("index", &index), ("roughness", &rough), ("metallic", &metal), ("occlusion", &occl)] {
        write_png(&format!("{}/{name}.png", args.out), n, img)?;
    }
    for (k, (count, sums)) in per_return.iter().enumerate() {
        let mean: Vec<String> = sums.iter().map(|s| format!("{:.3}", s / (*count).max(1) as f64)).collect();
        println!(
            "return {}: {} px, mean material [{}], normals non-finite {} zero {}",
            k + 1,
            count,
            mean.join(", "),
            bad_normals[k].0,
            bad_normals[k].1
        );
    }
    Ok(())
}
