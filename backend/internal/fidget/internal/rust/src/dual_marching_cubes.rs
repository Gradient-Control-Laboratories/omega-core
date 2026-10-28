/// Dual Marching Cubes (Nielson 2004 / Schaefer & Warren 2004) for manifold isosurface extraction.
///
/// Applies marching cubes to the dual of the primal grid: each primal cell becomes a dual
/// vertex (with SDF = average of 8 corner values), and each interior primal vertex becomes a
/// dual cell (grouping 8 surrounding primal cells). MC lookup tables on the dual cell signs
/// produce guaranteed-manifold connectivity.
///
/// Streaming: O(n^2) memory — only two cell layers are needed at any time.

use crate::marching_cubes::{EDGE_TABLE, TRI_TABLE, CORNER_OFFSETS, EDGE_CORNERS};

/// Compute the cell-center SDF for one layer of primal cells from two adjacent value slices.
///
/// Returns a cells * cells array where cells = n - 1. Entry [iy * cells + ix] is the
/// average of the 8 corner SDF values of primal cell (ix, iy, iz).
fn compute_cell_layer(
    slice_lo: &[f32], // SDF values at z = iz,   indexed [iy * n + ix]
    slice_hi: &[f32], // SDF values at z = iz+1, indexed [iy * n + ix]
    n: usize,
) -> Vec<f32> {
    let cells = n - 1;
    let mut layer = Vec::with_capacity(cells * cells);

    for iy in 0..cells {
        for ix in 0..cells {
            let avg = (slice_lo[iy * n + ix]
                + slice_lo[iy * n + ix + 1]
                + slice_lo[(iy + 1) * n + ix]
                + slice_lo[(iy + 1) * n + ix + 1]
                + slice_hi[iy * n + ix]
                + slice_hi[iy * n + ix + 1]
                + slice_hi[(iy + 1) * n + ix]
                + slice_hi[(iy + 1) * n + ix + 1])
                * 0.125;
            layer.push(avg);
        }
    }

    layer
}

/// Process one z-layer of dual cells at interior primal vertices (vx, vy, vz = iz).
///
/// cell_layer_lo: cell SDFs for the layer at iz - 1 (cells between value slices iz-1 and iz).
/// cell_layer_hi: cell SDFs for the layer at iz (cells between value slices iz and iz+1).
///
/// For each interior primal vertex (vx, vy) where 1 <= vx, vy <= cells - 1, the 8
/// surrounding cells map to MC corners via CORNER_OFFSETS:
///   corner (dx, dy, dz) -> layer_lo if dz==0, layer_hi if dz==1
///                        -> cell (vx - 1 + dx, vy - 1 + dy)
fn process_dual_layer(
    cell_layer_lo: &[f32],
    cell_layer_hi: &[f32],
    cells: usize,
    iz: usize,     // primal vertex z-index (1-based, so dual cell z = iz)
    origin_x: f32,
    origin_y: f32,
    origin_z: f32,
    cell_size: f32,
    triangles: &mut Vec<f32>,
) {
    // Interior primal vertices: 1 <= vx, vy <= cells - 1
    for vy in 1..cells {
        for vx in 1..cells {
            // Gather 8 cell SDFs around primal vertex (vx, vy, iz).
            // Corner ordering matches CORNER_OFFSETS: (dx, dy, dz).
            let mut cell_sdfs = [0.0f32; 8];
            for (ci, &(dx, dy, dz)) in CORNER_OFFSETS.iter().enumerate() {
                let cix = vx - 1 + dx;
                let ciy = vy - 1 + dy;
                let idx = ciy * cells + cix;
                cell_sdfs[ci] = if dz == 0 {
                    cell_layer_lo[idx]
                } else {
                    cell_layer_hi[idx]
                };
            }

            // Build cube_index from cell signs
            let mut cube_index: usize = 0;
            for i in 0..8 {
                if cell_sdfs[i] < 0.0 {
                    cube_index |= 1 << i;
                }
            }

            let edge_bits = EDGE_TABLE[cube_index];
            if edge_bits == 0 {
                continue;
            }

            // Compute edge intersection vertices by interpolating between cell centers.
            // Cell center of cell (cix, ciy, ciz):
            //   x = origin_x + (cix + 0.5) * cell_size
            //   y = origin_y + (ciy + 0.5) * cell_size
            //   z = origin_z + (ciz + 0.5) * cell_size
            // where ciz = iz - 1 for dz==0, iz for dz==1.
            let mut edge_verts = [[0.0f32; 3]; 12];
            for e in 0..12 {
                if edge_bits & (1 << e) != 0 {
                    let (c0, c1) = EDGE_CORNERS[e];
                    let s0 = cell_sdfs[c0];
                    let s1 = cell_sdfs[c1];

                    let t = if (s1 - s0).abs() > 1e-10 {
                        -s0 / (s1 - s0)
                    } else {
                        0.5
                    };
                    let t = t.clamp(0.0, 1.0);

                    let (dx0, dy0, dz0) = CORNER_OFFSETS[c0];
                    let (dx1, dy1, dz1) = CORNER_OFFSETS[c1];

                    // Cell indices for corners c0 and c1
                    let cx0 = (vx - 1 + dx0) as f32 + 0.5;
                    let cy0 = (vy - 1 + dy0) as f32 + 0.5;
                    let cz0 = (iz - 1 + dz0) as f32 + 0.5;
                    let cx1 = (vx - 1 + dx1) as f32 + 0.5;
                    let cy1 = (vy - 1 + dy1) as f32 + 0.5;
                    let cz1 = (iz - 1 + dz1) as f32 + 0.5;

                    edge_verts[e] = [
                        origin_x + (cx0 + t * (cx1 - cx0)) * cell_size,
                        origin_y + (cy0 + t * (cy1 - cy0)) * cell_size,
                        origin_z + (cz0 + t * (cz1 - cz0)) * cell_size,
                    ];
                }
            }

            // Emit triangles
            let tri_row = &TRI_TABLE[cube_index];
            let mut ti = 0;
            while ti < 15 && tri_row[ti] >= 0 {
                let e0 = tri_row[ti] as usize;
                let e1 = tri_row[ti + 1] as usize;
                let e2 = tri_row[ti + 2] as usize;

                let v0 = edge_verts[e0];
                let v1 = edge_verts[e1];
                let v2 = edge_verts[e2];

                // Face normal via cross product
                let u = [v1[0] - v0[0], v1[1] - v0[1], v1[2] - v0[2]];
                let w = [v2[0] - v0[0], v2[1] - v0[1], v2[2] - v0[2]];
                let mut nx = u[1] * w[2] - u[2] * w[1];
                let mut ny = u[2] * w[0] - u[0] * w[2];
                let mut nz = u[0] * w[1] - u[1] * w[0];
                let len = (nx * nx + ny * ny + nz * nz).sqrt();
                if len > 1e-10 {
                    nx /= len;
                    ny /= len;
                    nz /= len;
                }

                triangles.extend_from_slice(&[nx, ny, nz]);
                triangles.extend_from_slice(&v0);
                triangles.extend_from_slice(&v1);
                triangles.extend_from_slice(&v2);

                ti += 3;
            }
        }
    }
}

/// Chunked streaming Dual Marching Cubes with the same two-level batching as standard MC.
///
/// Parameters and streaming model are identical to chunked_marching_cubes:
/// - max_buffer_slices: max Z-value-slices kept in memory (RAM limit).
/// - eval_batch_slices: slices per eval call (~1-2s wall-clock for responsive progress).
/// - eval_chunk_fn(iz_start, count): returns count * n * n SDF values.
/// - progress_fn(completed, total): called after each eval sub-batch.
///
/// Internally maintains two cell-layer buffers (O(n^2) each) in addition to the value buffer.
pub fn chunked_dual_marching_cubes<F, P>(
    n: usize,
    origin_x: f32,
    origin_y: f32,
    origin_z: f32,
    cell_size: f32,
    max_buffer_slices: usize,
    eval_batch_slices: usize,
    mut eval_chunk_fn: F,
    mut progress_fn: P,
) -> Vec<f32>
where
    F: FnMut(usize, usize) -> Vec<f32>,
    P: FnMut(u32, u32),
{
    let cells = n - 1;
    let points_per_slice = n * n;
    let mut triangles: Vec<f32> = Vec::new();

    // Value slice buffer (same as MC)
    let mut buffer: Vec<f32> = Vec::new();
    let mut buffer_start_iz: usize = 0;
    let mut buffer_slice_count: usize = 0;

    // Carry-over cell layer from previous chunk
    let mut prev_cell_layer: Option<Vec<f32>> = None;
    // The z-index of the cell layer in prev_cell_layer (layer at iz means cells between slices iz and iz+1)
    let mut prev_cell_layer_iz: usize = 0;

    let mut iz_next_eval: usize = 0;

    while iz_next_eval < n {
        // Phase 1: Fill value buffer with eval sub-batches
        let buffer_target = max_buffer_slices.min(n - buffer_start_iz);

        while buffer_slice_count < buffer_target && iz_next_eval < n {
            let remaining_for_buffer = buffer_target - buffer_slice_count;
            let remaining_in_grid = n - iz_next_eval;
            let batch = eval_batch_slices.min(remaining_for_buffer).min(remaining_in_grid);
            if batch == 0 {
                break;
            }

            let values = eval_chunk_fn(iz_next_eval, batch);
            let expected = batch * points_per_slice;
            buffer.extend_from_slice(&values[..expected]);
            buffer_slice_count += batch;
            iz_next_eval += batch;

            // Report progress
            let last_evaluated_iz = iz_next_eval - 1;
            if last_evaluated_iz > 0 {
                progress_fn(last_evaluated_iz.min(cells) as u32, cells as u32);
            }
        }

        // Phase 2: Compute cell layers and process dual cells
        if buffer_slice_count >= 2 {
            let cell_layers_in_buffer = buffer_slice_count - 1;
            for local_iz in 0..cell_layers_in_buffer {
                let global_cell_iz = buffer_start_iz + local_iz;
                if global_cell_iz >= cells {
                    break;
                }

                let lo_offset = local_iz * points_per_slice;
                let hi_offset = (local_iz + 1) * points_per_slice;
                let slice_lo = &buffer[lo_offset..lo_offset + points_per_slice];
                let slice_hi = &buffer[hi_offset..hi_offset + points_per_slice];

                let curr_cell_layer = compute_cell_layer(slice_lo, slice_hi, n);

                // Process dual cells if we have a previous cell layer
                // Dual cells at primal vertex z = global_cell_iz require:
                //   cell_layer_lo at z = global_cell_iz - 1
                //   cell_layer_hi at z = global_cell_iz
                if let Some(ref prev_layer) = prev_cell_layer {
                    if prev_cell_layer_iz + 1 == global_cell_iz && global_cell_iz >= 1 {
                        process_dual_layer(
                            prev_layer,
                            &curr_cell_layer,
                            cells,
                            global_cell_iz,
                            origin_x,
                            origin_y,
                            origin_z,
                            cell_size,
                            &mut triangles,
                        );
                    }
                }

                prev_cell_layer = Some(curr_cell_layer);
                prev_cell_layer_iz = global_cell_iz;
            }
        }

        // Phase 3: Keep last slice as overlap, reset buffer
        if buffer_slice_count > 0 {
            let last_offset = (buffer_slice_count - 1) * points_per_slice;
            let carry = buffer[last_offset..last_offset + points_per_slice].to_vec();
            buffer.clear();
            buffer.extend_from_slice(&carry);
            buffer_start_iz = iz_next_eval - 1;
            buffer_slice_count = 1;
        } else {
            break;
        }
    }

    triangles
}
