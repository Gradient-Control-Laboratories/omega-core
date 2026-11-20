//! An omega tape rendered by fidget-wgpu into a G-buffer.
//!
//! Each return is voxel-rendered on its own and the images merged on the GPU,
//! which keeps the nearest hit and tags it with its return; the colour pass
//! then evaluates each return's colour tapes at its hits, the channel pass
//! (an omega fork addition) its twelve material tapes, and SSAO runs over the
//! merged depth. A return without a colour takes the caller's palette colour,
//! and one without a material the default record, both as constants. Normals
//! are each return's own full-precision ones, not the merged image's packed
//! eight-bit ones.

use fidget_core::{
    context::{Context, Tree},
    render::VoxelSize,
    shape::ShapeVars,
    vm::VmShape,
};
use fidget_raster::voxel::RenderConfig;
use fidget_wgpu::{
    Context as GpuContext, Gpu, RenderShape,
    buf::{BufferTag, FlexBuffer, ReadBuffer},
    channels::RenderChannels,
    color::{RenderColor, ShapeColor},
    voxel::{
        channels::{self, ChannelSize},
        color, merge, render, shade, ssao,
    },
};
use tape_core::Decoded;

/// The material record's default, as the viewer's GSL_MATERIAL_DEFAULT_WGSL:
/// roughness 0.5, metallic 0, ior 1.5, no emission, no volume.
pub const DEFAULT_MATERIAL: [f32; 12] = [0.5, 0.0, 1.5, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0, 0.0];

fn to_shape(tree: &Tree) -> Result<VmShape, String> {
    let mut ctx = Context::new();
    let node = ctx.import(tree);
    VmShape::new(&ctx, node).map_err(|e| format!("shape: {e:?}"))
}

/// A decoded tape made ready to render.
pub struct Scene {
    shapes: Vec<RenderShape>,
    colors: RenderColor,
    // One channel pass per material row, four floats a pixel each: twelve at
    // once is 48 bytes a pixel, which passes WebGPU's 128 MiB storage binding
    // above 2.8 megapixels; 16 matches fidget's own geometry buffers.
    materials: Option<[RenderChannels; 3]>,
    vars: ShapeVars<f32>,
    /// Per return: its colour came from the tape (else the palette).
    pub has_color: Vec<bool>,
    /// Per return: its material came from the tape (else the default).
    pub has_material: Vec<bool>,
}

impl Scene {
    /// `palette` gives the colour (sRGB, 0-1) of each return that has none of
    /// its own, indexed by return, wrapping if short. The scene's tapes live in
    /// GPU buffers, so it is built for one `gpu`.
    pub fn new(gpu: &Gpu, d: Decoded, palette: &[[f32; 3]]) -> Result<Scene, String> {
        let n = d.output_trees.len();
        let mut shapes = Vec::with_capacity(n);
        let mut colors = Vec::with_capacity(n);
        let mut materials = Vec::with_capacity(n);
        let mut has_color = Vec::with_capacity(n);
        let mut has_material = Vec::with_capacity(n);
        for (k, tree) in d.output_trees.iter().enumerate() {
            let app = d.appearance.get(k);
            // A displaced surface is the field less the material's height.
            let field = match app {
                Some(a) if a.displaced && a.material.is_some() => {
                    tree.clone() - a.material.as_ref().unwrap()[10].clone()
                }
                _ => tree.clone(),
            };
            shapes.push(
                gpu.build_render_shape(&to_shape(&field)?)
                    .map_err(|e| format!("return {k}: {e}"))?,
            );

            let color: [Tree; 3] = match app.and_then(|a| a.color.clone()) {
                Some(c) => {
                    has_color.push(true);
                    c
                }
                None => {
                    has_color.push(false);
                    let p = if palette.is_empty() { [0.8, 0.8, 0.8] } else { palette[k % palette.len()] };
                    p.map(Tree::constant)
                }
            };
            let [r, g, b] = color.each_ref().map(|t| to_shape(t));
            colors.push(ShapeColor::Rgb { r: r?, g: g?, b: b? });

            let material: Vec<Tree> = match app.and_then(|a| a.material.clone()) {
                Some(m) => {
                    has_material.push(true);
                    m.to_vec()
                }
                None => {
                    has_material.push(false);
                    DEFAULT_MATERIAL.iter().map(|&v| Tree::constant(v)).collect()
                }
            };
            materials.push(material.iter().map(to_shape).collect::<Result<Vec<_>, _>>()?);
        }
        let colors = gpu.build_render_color(&colors).map_err(|e| format!("colour: {e}"))?;
        // Every return at the default: the constant needs no pass.
        let materials = if has_material.iter().any(|&m| m) {
            let row = |r: usize| -> Result<RenderChannels, String> {
                let rows: Vec<Vec<VmShape>> = materials.iter().map(|m| m[4 * r..4 * r + 4].to_vec()).collect();
                gpu.build_render_channels(&rows, 4).map_err(|e| format!("material: {e}"))
            };
            Some([row(0)?, row(1)?, row(2)?])
        } else {
            None
        };
        Ok(Scene { shapes, colors, materials, vars: d.shape_vars, has_color, has_material })
    }

    /// The number of returns.
    pub fn returns(&self) -> usize {
        self.shapes.len()
    }
}

/// One frame's G-buffer; pixel i is row i / width, column i % width, rows top
/// first, in fidget's screen conventions.
pub struct GBuffer {
    pub width: u32,
    pub height: u32,
    /// The return hit, 1-based; 0 for a miss.
    pub index: Vec<u16>,
    /// Voxel depth of the hit (0 for a miss), z toward the viewer.
    pub depth: Vec<u32>,
    /// The hit return's gradient in fidget's screen space, 3 per pixel.
    pub normal: Vec<f32>,
    /// sRGB colour, 4 bytes per pixel (alpha 255 on a hit).
    pub color: Vec<u8>,
    /// The material record, 12 per pixel.
    pub material: Vec<f32>,
    /// Blurred SSAO occlusion, 1 for none; NaN on a miss.
    pub occlusion: Vec<f32>,
}

/// The buffers that depend on the frame size, rebuilt when the size changes.
struct Frame {
    size: VoxelSize,
    render: render::Buffers,
    merge: merge::Buffers,
    /// The colour pass writes here.
    shade: shade::Buffers,
    ssao: ssao::Buffers,
    /// One per material row, built the first time a scene with materials is
    /// drawn at this size (they are 16 bytes a pixel each).
    channels: Option<[channels::Buffers; 3]>,
}

/// Copies a GPU buffer into a buffer the host can map.
fn readback<T: BufferTag>(gpu: &Gpu, buf: &FlexBuffer<T>) -> Result<ReadBuffer<T>, String> {
    let mut rb = gpu.read_buffer_for(buf);
    gpu.copy(buf, &mut rb).map_err(|e| format!("copy: {e}"))?;
    Ok(rb)
}

/// GPU state for rendering scenes; one per thread or worker.
pub struct Renderer {
    gpu: Gpu,
    voxel: render::Context,
    merge: merge::Context,
    color: color::Context,
    channels: channels::Context,
    ssao: ssao::Context,
    voxel_ws: render::Workspace,
    merge_ws: merge::Workspace,
    color_ws: color::Workspace,
    channel_ws: [channels::Workspace; 3],
    ssao_ws: ssao::Workspace,
    frame: Option<Frame>,
}

impl Renderer {
    pub fn new(gpu: &Gpu) -> Renderer {
        let voxel = render::Context::new(gpu);
        let merge = merge::Context::new(gpu);
        let color = color::Context::new(gpu);
        let channels = channels::Context::new(gpu);
        let ssao = ssao::Context::new(gpu);
        Renderer {
            voxel_ws: voxel.workspace(),
            merge_ws: merge.workspace(),
            color_ws: color.workspace(),
            channel_ws: [channels.workspace(), channels.workspace(), channels.workspace()],
            ssao_ws: ssao.workspace(),
            voxel,
            merge,
            color,
            channels,
            ssao,
            frame: None,
            gpu: gpu.clone(),
        }
    }

    pub async fn render(&mut self, scene: &Scene, cfg: RenderConfig) -> Result<GBuffer, String> {
        let gpu = &self.gpu;
        let size = cfg.image_size;
        let (w, h) = (size.width(), size.height());
        let px = (w as usize) * (h as usize);

        // The size-dependent buffers: kept from the last frame if it was this
        // size, else rebuilt (the old ones are dropped first, so the two sets
        // are never alive together).
        if self.frame.as_ref().map(|f| f.size) != Some(size) {
            self.frame = None;
            self.frame = Some(Frame {
                size,
                render: self.voxel.buffers(size).map_err(|e| format!("render buffers: {e}"))?,
                merge: self.merge.buffers(size).map_err(|e| format!("merge buffers: {e}"))?,
                shade: self.color.buffers(size).map_err(|e| format!("colour buffers: {e}"))?,
                ssao: self.ssao.buffers(size.into()).map_err(|e| format!("ssao buffers: {e}"))?,
                channels: None,
            });
        }
        let frame = self.frame.as_mut().unwrap();
        if scene.materials.is_some() && frame.channels.is_none() {
            let cs = ChannelSize { size, channels: 4 };
            let build = || self.channels.buffers(cs).map_err(|e| format!("material buffers: {e}"));
            frame.channels = Some([build()?, build()?, build()?]);
        }

        // Each return on its own, its geometry kept for the full-precision
        // normals, then merged: nearest hit wins, tagged with its return.
        frame.merge.reset();
        let settings = render::Settings { vars: &scene.vars, world_to_model: cfg.world_to_model };
        let mut geoms = Vec::with_capacity(scene.shapes.len());
        for shape in &scene.shapes {
            self.voxel
                .submit(shape, settings, &mut self.voxel_ws, &mut frame.render)
                .map_err(|e| format!("render: {e}"))?;
            geoms.push(readback(gpu, frame.render.output())?);
            self.merge
                .submit(
                    &frame.render,
                    merge::Settings { denoise: false, z_scale: 1.0 },
                    &mut self.merge_ws,
                    &mut frame.merge,
                )
                .map_err(|e| format!("merge: {e}"))?;
        }

        self.color
            .submit(
                color::Args { image: &frame.merge, color: &scene.colors },
                color::Settings { vars: &scene.vars, world_to_model: cfg.world_to_model },
                &mut self.color_ws,
                &mut frame.shade,
            )
            .map_err(|e| format!("colour: {e}"))?;
        if let (Some(rows), Some(outs)) = (&scene.materials, frame.channels.as_mut()) {
            for (r, m) in rows.iter().enumerate() {
                self.channels
                    .submit(
                        channels::Args { image: &frame.merge, channels: m },
                        channels::Settings { vars: &scene.vars, world_to_model: cfg.world_to_model },
                        &mut self.channel_ws[r],
                        &mut outs[r],
                    )
                    .map_err(|e| format!("material: {e}"))?;
            }
        }
        self.ssao
            .submit(&frame.merge, (), &mut self.ssao_ws, &mut frame.ssao)
            .map_err(|e| format!("ssao: {e}"))?;

        let mut merged_rb = readback(gpu, frame.merge.output())?;
        let mut color_rb = readback(gpu, frame.shade.output())?;
        let mut occl_rb = readback(gpu, frame.ssao.blurred_occlusion())?;
        let mut mat_rb = Vec::new();
        if let (Some(_), Some(outs)) = (&scene.materials, &frame.channels) {
            for out in outs {
                mat_rb.push(readback(gpu, out.output())?);
            }
        }

        let merged = gpu.map_async(&mut merged_rb).await.to_vec();
        let color = gpu.map_async(&mut color_rb).await.to_vec();
        let occlusion = gpu.map_async(&mut occl_rb).await.to_vec();
        let mut geom_px = Vec::with_capacity(geoms.len());
        for rb in &mut geoms {
            geom_px.push(gpu.map_async(rb).await.to_vec());
        }
        let mut material_rows: Vec<Vec<f32>> = Vec::new();
        for rb in mat_rb.iter_mut() {
            material_rows.push(gpu.map_async(rb).await.to_vec());
        }

        let mut g = GBuffer {
            width: w,
            height: h,
            index: vec![0; px],
            depth: vec![0; px],
            normal: vec![0.0; 3 * px],
            color: Vec::with_capacity(4 * px),
            material: Vec::with_capacity(12 * px),
            occlusion,
        };
        for i in 0..px {
            let v = merged[i];
            if v.z > 0 && (v.index as usize) < geom_px.len() {
                g.index[i] = v.index + 1;
                g.depth[i] = v.z;
                g.normal[3 * i..3 * i + 3].copy_from_slice(&geom_px[v.index as usize][i].normal);
            }
            g.color.extend_from_slice(&color[i].to_le_bytes());
            if material_rows.is_empty() {
                g.material.extend_from_slice(&DEFAULT_MATERIAL);
            } else {
                for row in &material_rows {
                    g.material.extend_from_slice(&row[4 * i..4 * i + 4]);
                }
            }
        }
        Ok(g)
    }
}
