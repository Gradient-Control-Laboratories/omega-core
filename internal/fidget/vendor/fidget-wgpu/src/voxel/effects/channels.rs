//! Per-pixel channel evaluation (omega fork)
//!
//! The color pass evaluates three channel tapes per shape at each pixel's hit
//! voxel and packs them, clamped, into RGBA8. This pass evaluates N channel
//! tapes per shape and writes N unclamped floats per pixel, for fields that are
//! not colors (a material record, whose index of refraction or emission
//! strength runs past 1).
use super::{
    COMMON_SHADER, ColorError, GenericColorWorkspace, MergeWorkspace, buffer_ro,
    buffer_rw,
};
use crate::{
    CopyVarsError, Gpu, RegPipeline, buf::FlexBuffer, color::ShapeChannelBuffers,
    shaders, tag,
};
use fidget_core::shape::ShapeVars;
use zerocopy::{FromBytes, Immutable, IntoBytes, KnownLayout};

const CHANNELS_SHADER: &str = include_str!("shaders/channels.wgsl");

fn channels_shader(reg_count: u8) -> String {
    let mut shader_code = shaders::opcode_constants();
    shader_code += &format!("const REG_COUNT: u32 = {reg_count};");
    shader_code
        + CHANNELS_SHADER
        + COMMON_SHADER
        + crate::voxel::TRANSFORM_INPUT
        + shaders::COMMON
        + shaders::TAPE_INTERPRETER
        + shaders::DUMMY_STACK
        + shaders::FLOAT_OPS
}

/// Configuration for the channel evaluation pass
#[derive(Copy, Clone, FromBytes, Immutable, IntoBytes, KnownLayout)]
#[repr(C)]
#[doc(hidden)]
pub struct ChannelConfig {
    mat: [f32; 16],
    axes: [u32; 3],
    _pad: u32,
    image_size: [u32; 2],
    channels: u32,
    _pad2: u32,
    // this is followed by a tape_data flexible array member
}

/// Workspace for channel evaluation
pub type ChannelWorkspace = GenericColorWorkspace<ChannelConfig>;

tag!(
    pub ChannelImageTag, f32, usize, STORAGE | COPY_SRC,
    "Buffer tag for per-pixel channel values (N floats per pixel)"
);

/// Output of a channel pass: N floats per pixel, row-major, N = the channel
/// count of the [`ShapeChannelBuffers`] it was run with
pub struct ChannelOutput {
    out: FlexBuffer<ChannelImageTag>,
}

impl ChannelOutput {
    /// Returns a reference to the output buffer
    pub fn output(&self) -> &FlexBuffer<ChannelImageTag> {
        &self.out
    }
}

pub(super) struct ChannelContext {
    config_bind_group_layout: wgpu::BindGroupLayout,
    image_bind_group_layout: wgpu::BindGroupLayout,
    pipeline: RegPipeline,
}

impl ChannelContext {
    pub fn new(device: &wgpu::Device) -> Self {
        let config_bind_group_layout =
            device.create_bind_group_layout(&wgpu::BindGroupLayoutDescriptor {
                label: Some("channel config and shape"),
                entries: &[buffer_ro(0), buffer_ro(1), buffer_ro(2)],
            });
        let image_bind_group_layout =
            device.create_bind_group_layout(&wgpu::BindGroupLayoutDescriptor {
                label: Some("channel images"),
                entries: &[buffer_ro(0), buffer_rw(1)],
            });
        let pipeline_layout =
            device.create_pipeline_layout(&wgpu::PipelineLayoutDescriptor {
                label: Some("channel pipeline"),
                bind_group_layouts: &[
                    Some(&config_bind_group_layout),
                    Some(&image_bind_group_layout),
                ],
                immediate_size: 0u32,
            });
        let device_ = device.clone();
        let pipeline = RegPipeline::build(Box::new(move |reg_count| {
            let shader_module =
                device_.create_shader_module(wgpu::ShaderModuleDescriptor {
                    label: None,
                    source: wgpu::ShaderSource::Wgsl(
                        channels_shader(reg_count).into(),
                    ),
                });
            device_.create_compute_pipeline(&wgpu::ComputePipelineDescriptor {
                label: Some(&format!("channels ({reg_count})")),
                layout: Some(&pipeline_layout),
                module: &shader_module,
                entry_point: Some("channels_main"),
                compilation_options: Default::default(),
                cache: None,
            })
        }));
        Self {
            config_bind_group_layout,
            image_bind_group_layout,
            pipeline,
        }
    }

    pub fn output(&self, device: &wgpu::Device) -> ChannelOutput {
        ChannelOutput {
            out: FlexBuffer::new(device, "channel output", 4usize).unwrap(),
        }
    }

    #[allow(clippy::too_many_arguments)]
    pub fn submit(
        &self,
        image: &MergeWorkspace,
        world_to_model: &nalgebra::Matrix4<f32>,
        shape: &ShapeChannelBuffers,
        bufs: &mut ChannelWorkspace,
        vars: &ShapeVars<f32>,
        out: &mut ChannelOutput,
        gpu: &Gpu,
    ) -> Result<(), ColorError> {
        if image.image_count != shape.shape_count() {
            return Err(ColorError::BadShapeCount {
                merge_count: image.image_count,
                shape_count: shape.shape_count(),
            });
        }
        let size = image.out.size();
        let floats = (size.width() as usize)
            * (size.height() as usize)
            * shape.channels().max(1);
        out.out
            .grow_to_fit(&gpu.device, floats)
            .map_err(ColorError::OutputSize)?;

        let mat = world_to_model * size.screen_to_world();
        bufs.copy_vars(gpu, shape.var_map(), vars)
            .map_err(|e| match e {
                CopyVarsError::BufferSize(b) => ColorError::VarBufferSize(b),
                CopyVarsError::MissingVar(v) => ColorError::MissingVar(v),
            })?;
        let config = ChannelConfig {
            mat: mat.data.as_slice().try_into().unwrap(),
            axes: shape.axes(),
            _pad: 0,
            image_size: [size.width(), size.height()],
            channels: u32::try_from(shape.channels()).unwrap(),
            _pad2: 0,
        };
        bufs.copy_config_and_tape(gpu, &config, shape.bytecode())
            .map_err(ColorError::ConfigBufferSize)?;
        bufs.copy_shape_starts(gpu, shape.shape_start())
            .expect("shape starts should always fit if shape bytecode fits");

        let config_bg =
            bufs.config_bind_group(&gpu.device, &self.config_bind_group_layout);
        let mut encoder = gpu.device.create_command_encoder(
            &wgpu::CommandEncoderDescriptor { label: None },
        );
        {
            let mut pass =
                encoder.begin_compute_pass(&wgpu::ComputePassDescriptor {
                    label: None,
                    timestamp_writes: None,
                });
            pass.set_bind_group(0, config_bg, &[]);
            let image_bg =
                gpu.device.create_bind_group(&wgpu::BindGroupDescriptor {
                    label: Some("channel image bind group"),
                    layout: &self.image_bind_group_layout,
                    entries: &[
                        wgpu::BindGroupEntry {
                            binding: 0,
                            resource: image.out.bind_active(),
                        },
                        wgpu::BindGroupEntry {
                            binding: 1,
                            resource: out.out.bind_active(),
                        },
                    ],
                });
            pass.set_bind_group(1, &image_bg, &[]);
            pass.set_pipeline(self.pipeline.get(shape.reg_count()));
            pass.dispatch_workgroups(
                size.width().div_ceil(8),
                size.height().div_ceil(8),
                1,
            );
        }
        gpu.queue.submit(Some(encoder.finish()));
        Ok(())
    }
}
