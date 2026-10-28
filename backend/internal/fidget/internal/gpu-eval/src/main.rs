//! gpu-eval runs one WGSL compute shader over a list of points, headless.
//!
//!   gpu-eval --shader S.wgsl --points P.f32 --stride N --out O.f32 [--noise T.bin] [--repeat R] [--timing 1]
//!
//! The shader's entry point is `main` with @workgroup_size(64). It binds
//!
//!   @group(0) @binding(0) var<storage, read> points: array<vec4<f32>>
//!   @group(0) @binding(1) var<storage, read_write> out: array<f32>
//!   @group(0) @binding(2) var noise: texture_3d<f32>      (only with --noise)
//!
//! P.f32 is little-endian f32 xyzw per point; O.f32 receives `stride` f32 per
//! point. --noise is the 64^3 RGBA8 tile common/noise bakes, bound as
//! rgba8unorm like the viewers bind it. --repeat R dispatches the shader R more
//! times after the one written out, each waited on alone, and prints the
//! milliseconds each took to stdout, one per line: the shader's own time,
//! without starting the GPU, compiling, uploading or reading back. --timing 1
//! prints, to stderr as one JSON line, the milliseconds creating the shader
//! module took (naga's parse and validation) and creating the pipeline took
//! (the backend's translation and the driver's compile, Metal's on a Mac),
//! the compile cost a material's shader pays before it draws anything.
//!
//! Exit codes: 0 ok, 2 the shader failed to compile or validate (the message
//! is on stderr), 3 no GPU adapter (callers skip), 1 anything else.

use std::process::exit;
use wgpu::util::DeviceExt;

const NOISE_TILE: u32 = 64;

struct Args {
    shader: String,
    points: String,
    stride: usize,
    out: String,
    noise: Option<String>,
    repeat: usize,
    timing: bool,
}

fn parse_args() -> Result<Args, String> {
    let mut shader = None;
    let mut points = None;
    let mut stride = None;
    let mut out = None;
    let mut noise = None;
    let mut repeat = 0;
    let mut timing = false;
    let mut it = std::env::args().skip(1);
    while let Some(flag) = it.next() {
        let value = it.next().ok_or_else(|| format!("{flag} needs a value"))?;
        match flag.as_str() {
            "--shader" => shader = Some(value),
            "--points" => points = Some(value),
            "--stride" => stride = Some(value.parse::<usize>().map_err(|e| e.to_string())?),
            "--out" => out = Some(value),
            "--noise" => noise = Some(value),
            "--repeat" => repeat = value.parse::<usize>().map_err(|e| e.to_string())?,
            "--timing" => timing = value != "0",
            _ => return Err(format!("unknown flag {flag}")),
        }
    }
    Ok(Args {
        shader: shader.ok_or("--shader is required")?,
        points: points.ok_or("--points is required")?,
        stride: stride.ok_or("--stride is required")?,
        out: out.ok_or("--out is required")?,
        noise,
        repeat,
        timing,
    })
}

fn fail(code: i32, msg: impl std::fmt::Display) -> ! {
    eprintln!("gpu-eval: {msg}");
    exit(code)
}

fn main() {
    let args = parse_args().unwrap_or_else(|e| fail(1, e));
    let source = std::fs::read_to_string(&args.shader).unwrap_or_else(|e| fail(1, e));
    let point_bytes = std::fs::read(&args.points).unwrap_or_else(|e| fail(1, e));
    if point_bytes.len() % 16 != 0 {
        fail(1, "points file is not a whole number of vec4<f32>");
    }
    let count = point_bytes.len() / 16;
    let noise = args
        .noise
        .as_ref()
        .map(|path| std::fs::read(path).unwrap_or_else(|e| fail(1, e)));

    let instance = wgpu::Instance::new(&wgpu::InstanceDescriptor::default());
    let adapter = pollster::block_on(instance.request_adapter(&wgpu::RequestAdapterOptions {
        power_preference: wgpu::PowerPreference::HighPerformance,
        ..Default::default()
    }))
    .unwrap_or_else(|e| fail(3, format!("no GPU adapter: {e}")));
    let (device, queue) = pollster::block_on(adapter.request_device(&wgpu::DeviceDescriptor {
        label: Some("gpu-eval"),
        required_features: wgpu::Features::empty(),
        required_limits: adapter.limits(),
        ..Default::default()
    }))
    .unwrap_or_else(|e| fail(1, format!("request_device: {e}")));

    device.push_error_scope(wgpu::ErrorFilter::Validation);
    let module_start = std::time::Instant::now();
    let module = device.create_shader_module(wgpu::ShaderModuleDescriptor {
        label: Some("gpu-eval shader"),
        source: wgpu::ShaderSource::Wgsl(source.into()),
    });
    let module_ms = module_start.elapsed().as_secs_f64() * 1000.0;
    let pipeline_start = std::time::Instant::now();
    let pipeline = device.create_compute_pipeline(&wgpu::ComputePipelineDescriptor {
        label: Some("gpu-eval pipeline"),
        layout: None,
        module: &module,
        entry_point: Some("main"),
        compilation_options: Default::default(),
        cache: None,
    });
    let pipeline_ms = pipeline_start.elapsed().as_secs_f64() * 1000.0;
    if let Some(err) = pollster::block_on(device.pop_error_scope()) {
        fail(2, err);
    }
    if args.timing {
        eprintln!("{{\"module_ms\": {module_ms:.3}, \"pipeline_ms\": {pipeline_ms:.3}}}");
    }

    let out_size = (count * args.stride * 4) as u64;
    let points_buf = device.create_buffer_init(&wgpu::util::BufferInitDescriptor {
        label: Some("points"),
        contents: &point_bytes,
        usage: wgpu::BufferUsages::STORAGE,
    });
    let out_buf = device.create_buffer(&wgpu::BufferDescriptor {
        label: Some("out"),
        size: out_size,
        usage: wgpu::BufferUsages::STORAGE | wgpu::BufferUsages::COPY_SRC,
        mapped_at_creation: false,
    });
    let read_buf = device.create_buffer(&wgpu::BufferDescriptor {
        label: Some("readback"),
        size: out_size,
        usage: wgpu::BufferUsages::MAP_READ | wgpu::BufferUsages::COPY_DST,
        mapped_at_creation: false,
    });

    let mut entries = vec![
        wgpu::BindGroupEntry { binding: 0, resource: points_buf.as_entire_binding() },
        wgpu::BindGroupEntry { binding: 1, resource: out_buf.as_entire_binding() },
    ];
    let noise_view;
    if let Some(tile) = &noise {
        let expected = (NOISE_TILE * NOISE_TILE * NOISE_TILE * 4) as usize;
        if tile.len() != expected {
            fail(1, format!("noise tile is {} bytes, want {expected}", tile.len()));
        }
        let size = wgpu::Extent3d {
            width: NOISE_TILE,
            height: NOISE_TILE,
            depth_or_array_layers: NOISE_TILE,
        };
        let texture = device.create_texture(&wgpu::TextureDescriptor {
            label: Some("noise"),
            size,
            mip_level_count: 1,
            sample_count: 1,
            dimension: wgpu::TextureDimension::D3,
            format: wgpu::TextureFormat::Rgba8Unorm,
            usage: wgpu::TextureUsages::TEXTURE_BINDING | wgpu::TextureUsages::COPY_DST,
            view_formats: &[],
        });
        queue.write_texture(
            wgpu::TexelCopyTextureInfo {
                texture: &texture,
                mip_level: 0,
                origin: wgpu::Origin3d::ZERO,
                aspect: wgpu::TextureAspect::All,
            },
            tile,
            wgpu::TexelCopyBufferLayout {
                offset: 0,
                bytes_per_row: Some(NOISE_TILE * 4),
                rows_per_image: Some(NOISE_TILE),
            },
            size,
        );
        noise_view = texture.create_view(&Default::default());
        entries.push(wgpu::BindGroupEntry {
            binding: 2,
            resource: wgpu::BindingResource::TextureView(&noise_view),
        });
    }
    let bind_group = device.create_bind_group(&wgpu::BindGroupDescriptor {
        label: Some("gpu-eval bindings"),
        layout: &pipeline.get_bind_group_layout(0),
        entries: &entries,
    });

    let mut encoder = device.create_command_encoder(&Default::default());
    {
        let mut pass = encoder.begin_compute_pass(&Default::default());
        pass.set_pipeline(&pipeline);
        pass.set_bind_group(0, &bind_group, &[]);
        pass.dispatch_workgroups(count.div_ceil(64) as u32, 1, 1);
    }
    encoder.copy_buffer_to_buffer(&out_buf, 0, &read_buf, 0, out_size);
    queue.submit([encoder.finish()]);

    let slice = read_buf.slice(..);
    slice.map_async(wgpu::MapMode::Read, |r| {
        if let Err(e) = r {
            fail(1, format!("map: {e}"));
        }
    });
    device
        .poll(wgpu::PollType::Wait)
        .unwrap_or_else(|e| fail(1, format!("poll: {e}")));
    std::fs::write(&args.out, &*slice.get_mapped_range()).unwrap_or_else(|e| fail(1, e));

    for _ in 0..args.repeat {
        let mut encoder = device.create_command_encoder(&Default::default());
        {
            let mut pass = encoder.begin_compute_pass(&Default::default());
            pass.set_pipeline(&pipeline);
            pass.set_bind_group(0, &bind_group, &[]);
            pass.dispatch_workgroups(count.div_ceil(64) as u32, 1, 1);
        }
        let start = std::time::Instant::now();
        queue.submit([encoder.finish()]);
        device
            .poll(wgpu::PollType::Wait)
            .unwrap_or_else(|e| fail(1, format!("poll: {e}")));
        println!("{:.3}", start.elapsed().as_secs_f64() * 1000.0);
    }
}
