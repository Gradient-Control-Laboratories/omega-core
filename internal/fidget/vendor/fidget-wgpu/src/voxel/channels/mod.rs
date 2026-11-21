//! Per-pixel channel evaluation
//!
//! The [color pass](super::color) evaluates three channel tapes per shape at
//! each pixel's hit voxel and packs them, clamped, into RGBA8.  This pass
//! evaluates N channel tapes per shape and writes N unclamped floats per
//! pixel, for fields that are not colors (a material record, whose index of
//! refraction or emission strength runs past 1).
use crate::{
    ConfigWorkspace, Gpu, RegPipeline, VoxelSizeMismatch,
    buf::{BufferItemCount, BufferSizeError, FlexBuffer},
    channels::RenderChannels,
    tag,
};
use fidget_core::{
    render::VoxelSize,
    shape::{MissingVar, ShapeVars},
};
use zerocopy::{FromBytes, Immutable, IntoBytes, KnownLayout};

mod shaders {
    pub const CHANNELS: &str = include_str!("shaders/channels.wgsl");
}

fn channels_shader(reg_count: u8) -> String {
    let mut shader_code = crate::shaders::opcode_constants();
    shader_code += &format!("const REG_COUNT: u32 = {reg_count};");
    shader_code
        + shaders::CHANNELS
        + super::shaders::TRANSFORM_INPUT
        + super::shaders::PACKED_VOXEL
        + crate::shaders::COMMON
        + crate::shaders::TAPE_INTERPRETER
        + crate::shaders::DUMMY_STACK
        + crate::shaders::FLOAT_OPS
}

/// Size of a channel image: a voxel image size, and floats written per pixel
#[derive(Copy, Clone, Debug, PartialEq)]
pub struct ChannelSize {
    /// Image size
    pub size: VoxelSize,
    /// Number of channels (floats) per pixel
    pub channels: usize,
}

impl BufferItemCount for ChannelSize {
    fn item_count(&self) -> usize {
        self.size
            .item_count()
            .checked_mul(self.channels.max(1))
            .unwrap()
    }
}

tag!(
    pub ChannelImageTag, f32, ChannelSize, STORAGE | COPY_SRC,
    "Buffer tag for per-pixel channel values (N floats per pixel)"
);

/// Context for evaluating channels at each pixel
pub struct Context {
    gpu: Gpu,
    /// Pipeline for computing per-pixel channels
    pipeline: RegPipeline,
}

/// Arguments for channel evaluation
pub struct Args<'a> {
    /// Incoming image buffer
    pub image: &'a super::merge::Buffers,
    /// Bytecode buffers for channel evaluation
    pub channels: &'a RenderChannels,
}

/// Settings for channel evaluation
pub struct Settings<'a> {
    /// Values for free variables
    pub vars: &'a ShapeVars<f32>,
    /// Transform matrix
    pub world_to_model: nalgebra::Matrix4<f32>,
}

impl Default for Settings<'_> {
    fn default() -> Self {
        Self {
            vars: &crate::EMPTY_VARS,
            world_to_model: nalgebra::Matrix4::identity(),
        }
    }
}

/// Handle to size-dependent buffers used when evaluating channels
///
/// The output is N floats per pixel, row-major, where N is the channel count
/// of the [`ChannelSize`] it was built with (and of the [`RenderChannels`] it
/// is used with).
pub struct Buffers {
    out: FlexBuffer<ChannelImageTag>,
    bind_group: wgpu::BindGroup,
}

impl Buffers {
    pub(crate) fn new(
        gpu: &Gpu,
        size: ChannelSize,
    ) -> Result<Buffers, BufferSizeError> {
        let out = FlexBuffer::new(&gpu.device, "channel output", size)?;
        let bind_group =
            gpu.device.create_bind_group(&wgpu::BindGroupDescriptor {
                label: Some("channel output bind group"),
                layout: &gpu.output_bind_group_layout,
                entries: &[wgpu::BindGroupEntry {
                    binding: 0,
                    resource: out.bind_active(),
                }],
            });
        Ok(Buffers { out, bind_group })
    }

    /// Returns a reference to the output buffer
    pub fn output(&self) -> &FlexBuffer<ChannelImageTag> {
        &self.out
    }

    /// Returns the number of channels (floats) per pixel
    pub fn channels(&self) -> usize {
        self.out.size().channels
    }
}

impl crate::Context for Context {
    type Buffers = Buffers;
    type Args<'a> = Args<'a>;
    type Workspace = Workspace;
    type Settings<'a> = Settings<'a>;
    type Size = ChannelSize;
    type BuffersError = BufferSizeError;
    type EncodeError = ChannelError;

    fn new(gpu: &Gpu) -> Self {
        let device = &gpu.device;
        let pipeline_layout =
            device.create_pipeline_layout(&wgpu::PipelineLayoutDescriptor {
                label: Some("channels pipeline"),
                bind_group_layouts: &[
                    Some(&gpu.config_bind_group_layout),
                    Some(&gpu.color_bind_group_layout),
                    Some(&gpu.input_bind_group_layout),
                    Some(&gpu.output_bind_group_layout),
                ],
                immediate_size: 0u32,
            });
        let device_ = device.clone();
        let pipeline = RegPipeline::build(Box::new(move |reg_count| {
            let shader_code = channels_shader(reg_count);
            let shader_module =
                device_.create_shader_module(wgpu::ShaderModuleDescriptor {
                    label: None,
                    source: wgpu::ShaderSource::Wgsl(shader_code.into()),
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
            gpu: gpu.clone(),
            pipeline,
        }
    }

    fn encode(
        &self,
        args: Args<'_>,
        settings: Settings<'_>,
        workspace: &mut Workspace,
        out: &mut Buffers,
        encoder: &mut wgpu::CommandEncoder,
        staging: &mut wgpu::util::StagingBelt,
    ) -> Result<(), ChannelError> {
        if args.image.image_count != args.channels.shape_count() {
            return Err(ChannelError::BadShapeCount {
                merge_count: args.image.image_count,
                shape_count: args.channels.shape_count(),
            });
        }
        if out.channels() != args.channels.channels() {
            return Err(ChannelError::BadChannelCount {
                buffers: out.channels(),
                shapes: args.channels.channels(),
            });
        }
        let size = args.image.out.size();
        if size != out.out.size().size {
            return Err(VoxelSizeMismatch {
                expected: size,
                actual: out.out.size().size,
            }
            .into());
        }
        args.channels.copy_vars(encoder, staging, settings.vars)?;

        let mat = settings.world_to_model * size.screen_to_world();
        let config = ChannelConfig {
            mat: mat.data.as_slice().try_into().unwrap(),
            axes: args.channels.axes(),
            _pad1: 0,
            image_size: [size.width(), size.height()],
            channels: u32::try_from(args.channels.channels()).unwrap(),
            _pad2: 0,
        };
        workspace.encode_copy(encoder, staging, &config);

        let mut compute_pass =
            encoder.begin_compute_pass(&wgpu::ComputePassDescriptor {
                label: None,
                timestamp_writes: None,
            });
        compute_pass.set_bind_group(0, &workspace.bind_group, &[]);
        compute_pass.set_bind_group(1, args.channels.bind_group(), &[]);
        compute_pass.set_bind_group(2, &args.image.ro_bind_group, &[]);
        compute_pass.set_bind_group(3, &out.bind_group, &[]);

        compute_pass.set_pipeline(self.pipeline.get(args.channels.reg_count()));
        compute_pass.dispatch_workgroups(
            size.width().div_ceil(8),
            size.height().div_ceil(8),
            1,
        );
        Ok(())
    }

    fn workspace(&self) -> Workspace {
        self.gpu.config_workspace("channels")
    }

    fn buffers(
        &self,
        size: Self::Size,
    ) -> Result<Self::Buffers, Self::BuffersError> {
        Self::Buffers::new(&self.gpu, size)
    }

    fn gpu(&self) -> &Gpu {
        &self.gpu
    }
}

/// Workspace for evaluating voxel channels
pub type Workspace = ConfigWorkspace<ChannelConfig>;

/// Configuration for the channel evaluation pass
///
/// This is quietly public because it's part of a public API, but is unlikely to
/// be useful for end-users of the library.
#[derive(Copy, Clone, FromBytes, Immutable, IntoBytes, KnownLayout)]
#[cfg_attr(test, derive(facet::Facet))]
#[repr(C)]
#[doc(hidden)]
pub struct ChannelConfig {
    /// Screen-to-model transform matrix
    mat: [f32; 16],

    /// Input index of X, Y, Z axes
    ///
    /// `u32::MAX` is used as a marker if an axis is unused
    axes: [u32; 3],

    // padding for alignment
    _pad1: u32,

    /// Image size (pixels)
    image_size: [u32; 2],

    /// Channel tapes per shape, and floats written per pixel
    channels: u32,

    // padding for alignment
    _pad2: u32,
}

/// Error returned when submitting a channel evaluation operation
#[derive(Debug, thiserror::Error)]
pub enum ChannelError {
    /// A variable is missing from the map
    #[error(transparent)]
    MissingVar(#[from] MissingVar),

    /// Wrong number of shapes
    #[error(
        "merged image is built from {merge_count} shapes,
         but shape channel buffer expected {shape_count} shapes"
    )]
    BadShapeCount {
        /// Number of shapes in the merged image
        merge_count: usize,
        /// Number of shapes in the channel buffer
        shape_count: usize,
    },

    /// The output buffers and the shapes disagree on the channel count
    #[error(
        "output buffers hold {buffers} channels per pixel,
         but shapes have {shapes} channels"
    )]
    BadChannelCount {
        /// Channels per pixel in the output buffers
        buffers: usize,
        /// Channel tapes per shape
        shapes: usize,
    },

    /// The input and output buffers have mismatched sizes
    #[error(transparent)]
    SizeMismatch(#[from] VoxelSizeMismatch),
}

#[cfg(test)]
mod test {
    use super::*;

    #[test]
    fn channel_config_layout() {
        crate::test::compare_struct_layout::<ChannelConfig>(
            &channels_shader(16),
            "Config",
        );
    }

    #[test]
    fn compile_channels_shader() {
        crate::compile_shader(&channels_shader(16), "channels");
    }
}
