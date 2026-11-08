#pragma once

#include <memory>

#include "rust/cxx.h"

#include <openvdb/openvdb.h>

namespace omega_vdb {

// Shared struct definitions matching the cxx::bridge declaration.
struct BBox;
struct LevelSetBlocks;

// Opaque handle wrapping an openvdb::FloatGrid::Ptr (a shared_ptr).
// cxx exposes this as UniquePtr<FloatGridHandle> on the Rust side.
struct FloatGridHandle {
    openvdb::FloatGrid::Ptr grid;
};

void ensure_initialized();

std::unique_ptr<FloatGridHandle> mesh_to_level_set(
    rust::Slice<const float> verts,
    rust::Slice<const std::uint32_t> tris,
    float voxel_size,
    float half_band_voxels);

float sample_point(const FloatGridHandle& grid, float x, float y, float z);
BBox world_bbox(const FloatGridHandle& grid);
float voxel_size_value(const FloatGridHandle& grid);

rust::Vec<float> extract_dense_grid(
    const FloatGridHandle& grid,
    std::uint32_t dims_x,
    std::uint32_t dims_y,
    std::uint32_t dims_z);

LevelSetBlocks extract_level_set(const FloatGridHandle& grid);

}  // namespace omega_vdb
