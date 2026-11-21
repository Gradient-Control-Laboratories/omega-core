//! Data types for evaluating per-shape channels
//!
//! A channel is an arbitrary scalar field, evaluated at each pixel's hit voxel.
//! Unlike [color](crate::color), channels are neither clamped nor packed, so
//! they can carry any per-surface value (e.g. the parameters of a material).
//!
//! [`RenderChannels`] is constructed by
//! [`Gpu::build_render_channels`](crate::Gpu::build_render_channels).
use crate::encode_copy_vars;
use fidget_bytecode::{Bytecode, ReservedRegister};
use fidget_core::{
    eval::Function,
    shape::{MissingVar, ShapeVars},
    var::{Var, VarMap},
    vm::VmShape,
};

use std::{collections::HashMap, sync::Arc};
use zerocopy::IntoBytes;

/// Error type when constructing a [`RenderChannels`]
#[derive(Debug, thiserror::Error)]
pub enum RenderChannelsError {
    /// The shape uses a reserved register
    #[error(transparent)]
    RegisterError(#[from] ReservedRegister),
}

/// GPU buffers for evaluating N channels of each shape
///
/// This is the [`RenderColor`](crate::color::RenderColor) layout, generalized
/// from three clamped color channels to N unclamped floats.
pub struct RenderChannels {
    /// Unified [`VarMap`] object
    vars: Arc<VarMap>,
    /// Number of shapes available
    shape_count: usize,
    /// Number of channel tapes per shape
    channels: usize,
    /// Buffer for storing variable values
    vars_buf: wgpu::Buffer,
    /// Maximum number of registers used by any tape
    reg_count: u8,
    /// Bind group which binds the tape, starts, and variable buffers
    bind_group: wgpu::BindGroup,
}

// Accessor methods
impl RenderChannels {
    pub(crate) fn shape_count(&self) -> usize {
        self.shape_count
    }
    /// Returns the number of channel tapes per shape
    pub fn channels(&self) -> usize {
        self.channels
    }
    pub(crate) fn reg_count(&self) -> u8 {
        self.reg_count
    }
    pub(crate) fn bind_group(&self) -> &wgpu::BindGroup {
        &self.bind_group
    }

    /// Copies variables into a variables buffer
    pub(crate) fn copy_vars(
        &self,
        encoder: &mut wgpu::CommandEncoder,
        staging: &mut wgpu::util::StagingBelt,
        vars: &ShapeVars<f32>,
    ) -> Result<(), MissingVar> {
        encode_copy_vars(encoder, staging, &self.vars, vars, &self.vars_buf)
    }
}

impl RenderChannels {
    /// Builds a new set of buffers for evaluating channels
    ///
    /// Each entry of `shapes` is one shape's list of `channels` channel tapes.
    ///
    /// # Panics
    /// If a shape does not have exactly `channels` tapes
    pub fn new(
        shapes: &[Vec<VmShape>],
        channels: usize,
        device: &wgpu::Device,
        bind_group_layout: &wgpu::BindGroupLayout,
    ) -> Result<Self, RenderChannelsError> {
        // Build a single unified variable map, used across all tapes
        let mut vars = VarMap::new();
        for s in shapes {
            assert_eq!(
                s.len(),
                channels,
                "every shape needs {channels} channel tapes"
            );
            for channel in s {
                let vs = channel.inner().vars();
                for (v, _index) in vs.iter() {
                    vars.insert(v);
                }
            }
        }
        let mut reg_count = 0;
        let mut shape_start = Vec::with_capacity(shapes.len());
        let mut bytecode_data: Vec<u32> = Vec::new();
        let mut local_var_map = HashMap::new();
        for s in shapes {
            // Divide by 2 to convert from `u32` to `TapeWord`
            shape_start.push(u32::try_from(bytecode_data.len() / 2).unwrap());
            for channel in s {
                // Build a local variable remapping array, reusing allocations
                local_var_map.clear();
                local_var_map.extend(channel.inner().vars().iter().map(
                    |(v, i)| {
                        (
                            u32::try_from(i).unwrap(),
                            u32::try_from(vars.get(&v).unwrap()).unwrap(),
                        )
                    },
                ));

                // Generate bytecode for the root tape
                let bytecode = Bytecode::new_with_input_map(
                    channel.inner().data(),
                    &local_var_map,
                )?;
                bytecode_data.extend(bytecode.data());
                reg_count = reg_count.max(bytecode.reg_count());
            }
        }

        let shape_start_buf = device.create_buffer(&wgpu::BufferDescriptor {
            label: Some("channel shape starts"),
            usage: wgpu::BufferUsages::STORAGE,
            size: (shape_start.len() * std::mem::size_of::<u32>()).max(4)
                as u64,
            mapped_at_creation: true,
        });
        if !shape_start.is_empty() {
            shape_start_buf
                .slice(..)
                .get_mapped_range_mut()
                .unwrap()
                .copy_from_slice(shape_start.as_bytes());
        }
        shape_start_buf.unmap();

        let vars_buf = device.create_buffer(&wgpu::BufferDescriptor {
            label: Some("channel vars"),
            size: (vars.len() * std::mem::size_of::<f32>()).max(4) as u64,
            usage: wgpu::BufferUsages::STORAGE | wgpu::BufferUsages::COPY_DST,
            mapped_at_creation: false,
        });

        let tape = device.create_buffer(&wgpu::BufferDescriptor {
            label: Some("channel tape"),
            size: (bytecode_data.len() * std::mem::size_of::<u32>()).max(4)
                as u64,
            usage: wgpu::BufferUsages::STORAGE,
            mapped_at_creation: true,
        });
        if !bytecode_data.is_empty() {
            tape.slice(..)
                .get_mapped_range_mut()
                .unwrap()
                .copy_from_slice(bytecode_data.as_bytes());
        }
        tape.unmap();

        let bind_group = device.create_bind_group(&wgpu::BindGroupDescriptor {
            label: Some("shape channels"),
            layout: bind_group_layout,
            entries: &[
                wgpu::BindGroupEntry {
                    binding: 0,
                    resource: tape.as_entire_binding(),
                },
                wgpu::BindGroupEntry {
                    binding: 1,
                    resource: shape_start_buf.as_entire_binding(),
                },
                wgpu::BindGroupEntry {
                    binding: 2,
                    resource: vars_buf.as_entire_binding(),
                },
            ],
        });

        Ok(Self {
            shape_count: shapes.len(),
            channels,
            vars: vars.into(),
            vars_buf,
            reg_count,
            bind_group,
        })
    }

    /// Helper function to return XYZ variable indices
    pub(crate) fn axes(&self) -> [u32; 3] {
        [Var::X, Var::Y, Var::Z]
            .map(|a| self.vars.get(&a).map(|v| v as u32).unwrap_or(u32::MAX))
    }
}

#[cfg(test)]
mod test {
    use crate::Gpu;

    #[test]
    fn empty_render_channels() {
        crate::require_gpu!();
        let gpu = pollster::block_on(Gpu::init_basic()).unwrap();

        gpu.build_render_channels(&[], 4).unwrap();
    }
}
