#include "openvdb_glue.h"

#include <atomic>
#include <stdexcept>

#include <openvdb/Types.h>
#include <openvdb/math/Transform.h>
#include <openvdb/tools/MeshToVolume.h>
#include <openvdb/tools/Interpolation.h>
#include <openvdb/util/NullInterrupter.h>
#include <openvdb/version.h>

#include "render/src/openvdb_ffi.rs.h"

namespace omega_vdb {

namespace {

std::atomic<bool> g_initialized{false};

// ProgressInterrupter forwards OpenVDB's wasInterrupted(percent) calls
// (which meshToLevelSet fires at known checkpoints with concrete
// percentages 30/45/50/54/.../94/99) into the Rust-side progress channel
// as *phase-local* progress: completed/total in [0, 1] meaning "fraction
// of the voxelization phase done". The Rust orchestrator owns the global
// bake timeline and remaps this into the overall progress bar — see
// vdb_progress_emit / set_progress_phase. We never return true for
// cancellation here; the host handles cancellation through the
// dispatcher request lifecycle.
struct ProgressInterrupter : public openvdb::util::NullInterrupter {
    void start(const char* /*name*/ = nullptr) override {
        omega_vdb::vdb_progress_emit(0, 100);
    }
    void end() override {
        omega_vdb::vdb_progress_emit(100, 100);
    }
    bool wasInterrupted(int percent = -1) override {
        if (percent >= 0) {
            std::uint32_t p = static_cast<std::uint32_t>(percent);
            if (p > 100) p = 100;
            omega_vdb::vdb_progress_emit(p, 100);
        }
        return false;
    }
};

}  // namespace

void ensure_initialized() {
    bool expected = false;
    if (g_initialized.compare_exchange_strong(expected, true)) {
        openvdb::initialize();
    }
}

std::unique_ptr<FloatGridHandle> mesh_to_level_set(
    rust::Slice<const float> verts,
    rust::Slice<const std::uint32_t> tris,
    float voxel_size,
    float half_band_voxels) {
    ensure_initialized();

    if (verts.size() % 3 != 0) {
        throw std::runtime_error("mesh_to_level_set: verts length not multiple of 3");
    }
    if (tris.size() % 3 != 0) {
        throw std::runtime_error("mesh_to_level_set: tris length not multiple of 3");
    }
    if (voxel_size <= 0.0f) {
        throw std::runtime_error("mesh_to_level_set: voxel_size must be > 0");
    }

    const std::size_t n_verts = verts.size() / 3;
    const std::size_t n_tris = tris.size() / 3;

    std::vector<openvdb::Vec3s> points;
    points.reserve(n_verts);
    for (std::size_t i = 0; i < n_verts; ++i) {
        points.emplace_back(
            verts[i * 3 + 0],
            verts[i * 3 + 1],
            verts[i * 3 + 2]);
    }

    std::vector<openvdb::Vec3I> triangles;
    triangles.reserve(n_tris);
    for (std::size_t i = 0; i < n_tris; ++i) {
        triangles.emplace_back(
            tris[i * 3 + 0],
            tris[i * 3 + 1],
            tris[i * 3 + 2]);
    }

    auto transform = openvdb::math::Transform::createLinearTransform(voxel_size);

    ProgressInterrupter interrupter;
    auto grid = openvdb::tools::meshToLevelSet<openvdb::FloatGrid>(
        interrupter,
        *transform,
        points,
        triangles,
        half_band_voxels);

    auto handle = std::make_unique<FloatGridHandle>();
    handle->grid = grid;
    return handle;
}

float sample_point(const FloatGridHandle& grid, float x, float y, float z) {
    if (!grid.grid) {
        return 0.0f;
    }
    openvdb::tools::GridSampler<openvdb::FloatGrid, openvdb::tools::BoxSampler>
        sampler(*grid.grid);
    return sampler.wsSample(openvdb::Vec3R(x, y, z));
}

BBox world_bbox(const FloatGridHandle& grid) {
    BBox out{0, 0, 0, 0, 0, 0};
    if (!grid.grid) {
        return out;
    }
    auto bbox = grid.grid->evalActiveVoxelBoundingBox();
    auto wmin = grid.grid->indexToWorld(bbox.min());
    auto wmax = grid.grid->indexToWorld(bbox.max());
    out.min_x = static_cast<float>(wmin.x());
    out.min_y = static_cast<float>(wmin.y());
    out.min_z = static_cast<float>(wmin.z());
    out.max_x = static_cast<float>(wmax.x());
    out.max_y = static_cast<float>(wmax.y());
    out.max_z = static_cast<float>(wmax.z());
    return out;
}

float voxel_size_value(const FloatGridHandle& grid) {
    if (!grid.grid) {
        return 0.0f;
    }
    return static_cast<float>(grid.grid->transform().voxelSize().x());
}

rust::Vec<float> extract_dense_grid(
    const FloatGridHandle& grid,
    std::uint32_t dims_x,
    std::uint32_t dims_y,
    std::uint32_t dims_z) {
    rust::Vec<float> out;
    if (!grid.grid || dims_x == 0 || dims_y == 0 || dims_z == 0) {
        return out;
    }
    const std::size_t total = static_cast<std::size_t>(dims_x)
        * static_cast<std::size_t>(dims_y)
        * static_cast<std::size_t>(dims_z);
    out.reserve(total);

    auto bbox = grid.grid->evalActiveVoxelBoundingBox();
    auto wmin = grid.grid->indexToWorld(bbox.min());
    auto wmax = grid.grid->indexToWorld(bbox.max());

    const double inv_x = (dims_x > 1)
        ? 1.0 / static_cast<double>(dims_x - 1)
        : 0.0;
    const double inv_y = (dims_y > 1)
        ? 1.0 / static_cast<double>(dims_y - 1)
        : 0.0;
    const double inv_z = (dims_z > 1)
        ? 1.0 / static_cast<double>(dims_z - 1)
        : 0.0;

    openvdb::tools::GridSampler<openvdb::FloatGrid, openvdb::tools::BoxSampler>
        sampler(*grid.grid);

    // Loop order matches Go's SdfGrid layout: x fastest, then y, then z.
    // After each Z-slice we report *phase-local* progress — the fraction
    // of Z-slices done — and let the Rust orchestrator remap it into the
    // global bake timeline. This phase is exactly linear in wall time:
    // every slice is the same dims_x·dims_y sample count.
    for (std::uint32_t iz = 0; iz < dims_z; ++iz) {
        const double tz = static_cast<double>(iz) * inv_z;
        const double z = wmin.z() + tz * (wmax.z() - wmin.z());
        for (std::uint32_t iy = 0; iy < dims_y; ++iy) {
            const double ty = static_cast<double>(iy) * inv_y;
            const double y = wmin.y() + ty * (wmax.y() - wmin.y());
            for (std::uint32_t ix = 0; ix < dims_x; ++ix) {
                const double tx = static_cast<double>(ix) * inv_x;
                const double x = wmin.x() + tx * (wmax.x() - wmin.x());
                out.push_back(sampler.wsSample(openvdb::Vec3R(x, y, z)));
            }
        }
        omega_vdb::vdb_progress_emit(iz + 1, dims_z);
    }
    return out;
}

namespace {

// The block (8^3 voxels) holding index coordinate v.
int block_of(int v) {
    return v >= 0 ? v / 8 : -((-v + 7) / 8);
}

}  // namespace

LevelSetBlocks extract_level_set(const FloatGridHandle& grid) {
    LevelSetBlocks out{};
    if (!grid.grid) {
        return out;
    }
    const auto& tree = grid.grid->tree();
    out.voxel_size = grid.grid->transform().voxelSize().x();
    out.background = tree.background();
    for (auto it = tree.root().cbeginValueAll(); it; ++it) {
        ++out.root_tiles;
    }

    openvdb::CoordBBox leaves;
    if (!tree.evalLeafBoundingBox(leaves)) {
        return out;
    }
    // One block of margin on each side: the sampler reads the voxel past
    // the one it falls in.
    const int x0 = block_of(leaves.min().x()) - 1;
    const int y0 = block_of(leaves.min().y()) - 1;
    const int z0 = block_of(leaves.min().z()) - 1;
    const int x1 = block_of(leaves.max().x()) + 1;
    const int y1 = block_of(leaves.max().y()) + 1;
    const int z1 = block_of(leaves.max().z()) + 1;
    out.origin_x = x0;
    out.origin_y = y0;
    out.origin_z = z0;
    out.dims_x = static_cast<std::uint32_t>(x1 - x0 + 1);
    out.dims_y = static_cast<std::uint32_t>(y1 - y0 + 1);
    out.dims_z = static_cast<std::uint32_t>(z1 - z0 + 1);

    auto accessor = grid.grid->getConstAccessor();
    const std::size_t blocks = static_cast<std::size_t>(out.dims_x) * out.dims_y * out.dims_z;
    out.block_leaf.reserve(blocks);
    out.block_value.reserve(blocks);
    out.leaf_values.reserve(tree.leafCount() * openvdb::FloatTree::LeafNodeType::SIZE);
    for (int bz = z0; bz <= z1; ++bz) {
        for (int by = y0; by <= y1; ++by) {
            for (int bx = x0; bx <= x1; ++bx) {
                const openvdb::Coord origin(bx * 8, by * 8, bz * 8);
                if (const auto* leaf = tree.probeConstLeaf(origin)) {
                    out.block_leaf.push_back(static_cast<std::int32_t>(
                        out.leaf_values.size() / openvdb::FloatTree::LeafNodeType::SIZE));
                    out.block_value.push_back(0.0f);
                    // Offset order is OpenVDB's: (i & 7) << 6 | (j & 7) << 3 | (k & 7).
                    for (openvdb::Index i = 0; i < openvdb::FloatTree::LeafNodeType::SIZE; ++i) {
                        out.leaf_values.push_back(leaf->getValue(i));
                    }
                } else {
                    // No leaf: the whole block is one tile, or the background.
                    out.block_leaf.push_back(-1);
                    out.block_value.push_back(accessor.getValue(origin));
                }
            }
        }
    }
    return out;
}

}  // namespace omega_vdb
