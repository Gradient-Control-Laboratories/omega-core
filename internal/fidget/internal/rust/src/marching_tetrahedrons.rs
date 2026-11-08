/// Isosurface extraction from a regular 3D grid of SDF values using
/// marching tetrahedra with Freudenthal decomposition.
///
/// Each cube cell is split into 6 tetrahedra using the Freudenthal triangulation,
/// which is automatically face-consistent between adjacent cells. Unlike classic
/// marching cubes, marching tetrahedra has NO ambiguous configurations, so the
/// output mesh is always a closed, orientable manifold.
///
/// Classic marching cubes uses a 256-entry lookup table that contains ambiguous
/// face cases (saddle points on shared faces). Adjacent cells can resolve these
/// ambiguities differently, producing non-manifold edges where 1 or 3 triangles
/// share an edge instead of exactly 2. Marching tetrahedra eliminates this entirely:
/// every tetrahedral sign configuration has a unique, unambiguous triangulation.
///
/// This implementation uses a streaming slice-by-slice approach: only two adjacent
/// Z-slices are kept in memory at any time, reducing memory from O(n^3) to O(n^2).
/// This allows depth 12 (4097^3 grid) to use ~134 MB instead of ~275 GB.

/// The 6 Freudenthal tetrahedra that partition each cube cell.
///
/// Each entry is 4 cube-corner indices [0..7], ordered to form a right-handed
/// (positive-volume) tetrahedron. The Freudenthal decomposition guarantees that
/// shared faces between adjacent cubes always use the same diagonal, eliminating
/// the face-ambiguity problem that causes non-manifold output in marching cubes.

const FREUDENTHAL_TETS: [[usize; 4]; 6] = [
    [0, 1, 2, 6],
    [0, 5, 1, 6],
    [0, 2, 3, 6],
    [0, 3, 7, 6],
    [0, 4, 5, 6],
    [0, 7, 4, 6],
];

/// The 6 edges of a tetrahedron as pairs of local vertex indices (0..3).
const TET_EDGES: [(usize, usize); 6] = [
    (0, 1), (0, 2), (0, 3), (1, 2), (1, 3), (2, 3),
];

/// Canonical ordering rank for each cube corner by global grid position
/// (lexicographic x, then y, then z).
///
/// Corner positions: 0=(0,0,0) 1=(1,0,0) 2=(1,1,0) 3=(0,1,0)
///                   4=(0,0,1) 5=(1,0,1) 6=(1,1,1) 7=(0,1,1)
/// Lex order: 0 < 3 < 4 < 7 < 1 < 2 < 5 < 6
///
/// This ensures edge interpolation always uses a consistent direction
/// so that every tet and every cell sharing the same grid edge computes
/// a bit-identical vertex position. Without this, the three Freudenthal
/// tets whose vertex order was swapped for right-handedness (tets 1, 2, 5)
/// access their (v1,v2) edges in reversed direction, causing floating-point
/// divergence from adjacent cells and producing non-manifold vertices.
const CORNER_CANONICAL: [u8; 8] = [0, 4, 5, 1, 2, 6, 7, 3];

/// Marching tetrahedra triangulation table.
///
/// For each of the 16 sign configurations (4 vertices × inside/outside),
/// lists triangle edges as indices into TET_EDGES, terminated by -1.
/// At most 2 triangles (6 edge indices) per configuration.
///
/// Winding order produces outward normals (from negative/inside SDF toward
/// positive/outside SDF) assuming right-handed tetrahedra.
const MT_TRI_TABLE: [[i8; 7]; 16] = [
    [-1,-1,-1,-1,-1,-1,-1], // 0000: all outside
    [ 0, 1, 2,-1,-1,-1,-1], // 0001: v0 inside
    [ 0, 4, 3,-1,-1,-1,-1], // 0010: v1 inside
    [ 1, 4, 3, 1, 2, 4,-1], // 0011: v0+v1 inside
    [ 1, 3, 5,-1,-1,-1,-1], // 0100: v2 inside
    [ 0, 3, 5, 0, 5, 2,-1], // 0101: v0+v2 inside
    [ 0, 4, 5, 0, 5, 1,-1], // 0110: v1+v2 inside
    [ 2, 4, 5,-1,-1,-1,-1], // 0111: v3 outside
    [ 2, 5, 4,-1,-1,-1,-1], // 1000: v3 inside
    [ 0, 5, 4, 0, 1, 5,-1], // 1001: v0+v3 inside
    [ 0, 5, 3, 0, 2, 5,-1], // 1010: v1+v3 inside
    [ 1, 5, 3,-1,-1,-1,-1], // 1011: v2 outside
    [ 1, 3, 4, 1, 4, 2,-1], // 1100: v2+v3 inside
    [ 0, 3, 4,-1,-1,-1,-1], // 1101: v1 outside
    [ 0, 2, 1,-1,-1,-1,-1], // 1110: v0 outside
    [-1,-1,-1,-1,-1,-1,-1], // 1111: all inside
];

/// Process a single layer of cells between two Z-slices using marching tetrahedra,
/// appending triangles to the output buffer.
///
/// Each cube cell is decomposed into 6 Freudenthal tetrahedra. Each tetrahedron
/// has only 16 possible sign configurations, all unambiguous, guaranteeing that
/// the resulting mesh is always manifold.
///
/// slice_lo: SDF values at z = iz, indexed as [iy * n + ix]
/// slice_hi: SDF values at z = iz+1, indexed as [iy * n + ix]
fn process_slice_pair(
    slice_lo: &[f32],
    slice_hi: &[f32],
    n: usize,
    iz: usize,
    origin_x: f32,
    origin_y: f32,
    origin_z: f32,
    cell_size: f32,
    triangles: &mut Vec<f32>,
) {
    let cells = n - 1;

    for iy in 0..cells {
        for ix in 0..cells {
            // Read the 8 corner values from the two slices
            // Corners 0-3 are from slice_lo (z=iz), corners 4-7 from slice_hi (z=iz+1)
            let corner_vals = [
                slice_lo[ iy      * n + ix    ], // 0: (ix,   iy,   iz)
                slice_lo[ iy      * n + ix + 1], // 1: (ix+1, iy,   iz)
                slice_lo[(iy + 1) * n + ix + 1], // 2: (ix+1, iy+1, iz)
                slice_lo[(iy + 1) * n + ix    ], // 3: (ix,   iy+1, iz)
                slice_hi[ iy      * n + ix    ], // 4: (ix,   iy,   iz+1)
                slice_hi[ iy      * n + ix + 1], // 5: (ix+1, iy,   iz+1)
                slice_hi[(iy + 1) * n + ix + 1], // 6: (ix+1, iy+1, iz+1)
                slice_hi[(iy + 1) * n + ix    ], // 7: (ix,   iy+1, iz+1)
            ];

            // Quick reject: skip if all corners are on the same side of the isosurface
            let any_neg = corner_vals.iter().any(|&v| v < 0.0);
            let any_pos = corner_vals.iter().any(|&v| v >= 0.0);
            if !any_neg || !any_pos {
                continue;
            }

            // Pre-compute the 8 corner world-space positions
            let ox = origin_x + ix as f32 * cell_size;
            let oy = origin_y + iy as f32 * cell_size;
            let oz = origin_z + iz as f32 * cell_size;
            let corner_pos = [
                [ox,             oy,             oz            ],
                [ox + cell_size, oy,             oz            ],
                [ox + cell_size, oy + cell_size, oz            ],
                [ox,             oy + cell_size, oz            ],
                [ox,             oy,             oz + cell_size],
                [ox + cell_size, oy,             oz + cell_size],
                [ox + cell_size, oy + cell_size, oz + cell_size],
                [ox,             oy + cell_size, oz + cell_size],
            ];

            // Process each of the 6 Freudenthal tetrahedra
            for tet in &FREUDENTHAL_TETS {
                let tv = [
                    corner_vals[tet[0]], corner_vals[tet[1]],
                    corner_vals[tet[2]], corner_vals[tet[3]],
                ];

                // 4-bit sign configuration
                let mut tet_index: usize = 0;
                for i in 0..4 {
                    if tv[i] < 0.0 {
                        tet_index |= 1 << i;
                    }
                }

                let tri_row = &MT_TRI_TABLE[tet_index];
                if tri_row[0] < 0 {
                    continue;
                }

                let tp = [
                    corner_pos[tet[0]], corner_pos[tet[1]],
                    corner_pos[tet[2]], corner_pos[tet[3]],
                ];

                // Compute intersection points on the 6 tet edges.
                // Always interpolate in canonical direction (lower global grid
                // position first) so that every tet/cell sharing a grid edge
                // produces a bit-identical vertex position.
                let mut edge_verts = [[0.0f32; 3]; 6];
                for e in 0..6 {
                    let (c0, c1) = TET_EDGES[e];
                    let cube_a = tet[c0];
                    let cube_b = tet[c1];
                    let (sv, ev, sp, ep) =
                        if CORNER_CANONICAL[cube_a] < CORNER_CANONICAL[cube_b] {
                            (tv[c0], tv[c1], tp[c0], tp[c1])
                        } else {
                            (tv[c1], tv[c0], tp[c1], tp[c0])
                        };
                    let t = if (ev - sv).abs() > 1e-10 {
                        (-sv / (ev - sv)).clamp(0.0, 1.0)
                    } else {
                        0.5
                    };
                    edge_verts[e] = [
                        sp[0] + t * (ep[0] - sp[0]),
                        sp[1] + t * (ep[1] - sp[1]),
                        sp[2] + t * (ep[2] - sp[2]),
                    ];
                }

                // Emit triangles from the lookup table
                let mut ti = 0;
                while ti < 6 && tri_row[ti] >= 0 {
                    let va = edge_verts[tri_row[ti] as usize];
                    let vb = edge_verts[tri_row[ti + 1] as usize];
                    let vc = edge_verts[tri_row[ti + 2] as usize];

                    // Face normal via cross product
                    let ux = vb[0] - va[0];
                    let uy = vb[1] - va[1];
                    let uz = vb[2] - va[2];
                    let wx = vc[0] - va[0];
                    let wy = vc[1] - va[1];
                    let wz = vc[2] - va[2];
                    let mut nx = uy * wz - uz * wy;
                    let mut ny = uz * wx - ux * wz;
                    let mut nz = ux * wy - uy * wx;
                    let len = (nx * nx + ny * ny + nz * nz).sqrt();

                    // Skip degenerate triangles (zero-area from coincident vertices)
                    if len > 1e-10 {
                        nx /= len;
                        ny /= len;
                        nz /= len;
                        triangles.extend_from_slice(&[nx, ny, nz]);
                        triangles.extend_from_slice(&va);
                        triangles.extend_from_slice(&vb);
                        triangles.extend_from_slice(&vc);
                    }

                    ti += 3;
                }
            }
        }
    }
}

/// Convert triangle data to binary STL format.
///
/// triangle_data: flat array of floats [nx,ny,nz, v0x,v0y,v0z, v1x,v1y,v1z, v2x,v2y,v2z, ...]
/// Each triangle is 12 floats (48 bytes). STL adds 2 attribute bytes per triangle (50 bytes).
pub fn triangles_to_stl(triangle_data: &[f32]) -> Vec<u8> {
    let num_triangles = triangle_data.len() / 12;

    // Binary STL: 80-byte header + 4-byte triangle count + 50 bytes per triangle
    let stl_size = 80 + 4 + num_triangles * 50;
    let mut stl = vec![0u8; stl_size];

    // Triangle count at offset 80 (u32 little-endian)
    stl[80..84].copy_from_slice(&(num_triangles as u32).to_le_bytes());

    // Reinterpret the f32 slice as bytes (safe on little-endian, which covers x86 + ARM)
    let float_bytes: &[u8] = unsafe {
        std::slice::from_raw_parts(
            triangle_data.as_ptr() as *const u8,
            triangle_data.len() * 4,
        )
    };

    // Each STL triangle = 48 bytes of float data + 2 bytes attribute (zero).
    // The stl buffer is already zero-initialized, so attribute bytes are already 0.
    let mut dst = 84;
    for i in 0..num_triangles {
        let src = i * 48; // 12 floats * 4 bytes
        stl[dst..dst + 48].copy_from_slice(&float_bytes[src..src + 48]);
        dst += 50; // 48 data + 2 attribute
    }

    stl
}

/// Chunked streaming marching cubes: evaluates SDF in batches for speed,
/// while keeping memory usage bounded and progress updates frequent.
///
/// Two-level batching:
/// - max_buffer_slices: max Z-slices kept in memory at once (RAM limit).
/// - eval_batch_slices: slices per eval call (~1-2s wall-clock for responsive progress).
///
/// eval_chunk_fn: given (iz_start, count), returns a flat Vec<f32> of count*n*n SDF values.
/// progress_fn: called with (completed_cell_layers, total_cell_layers) after each eval sub-batch.
pub fn chunked_marching_cubes<F, P>(
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
    let cells = n - 1; // total cell layers in Z
    let points_per_slice = n * n;
    let mut triangles: Vec<f32> = Vec::new();

    // Buffer holds up to max_buffer_slices of evaluated SDF data.
    // We fill it via sub-batches of eval_batch_slices, reporting progress after each.
    // Once the buffer is full (or we've reached the end), we triangulate all cell layers
    // in the buffer, keep the last slice as overlap, and continue.
    let mut buffer: Vec<f32> = Vec::new();
    let mut buffer_start_iz: usize = 0; // global Z-index of first slice in buffer
    let mut buffer_slice_count: usize = 0;

    let mut iz_next_eval: usize = 0; // next Z-index to evaluate

    while iz_next_eval < n {
        // Phase 1: Fill the buffer with sub-batches, reporting progress after each
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

            // Report progress after each sub-batch eval.
            let last_evaluated_iz = iz_next_eval - 1;
            if last_evaluated_iz > 0 {
                progress_fn(last_evaluated_iz.min(cells) as u32, cells as u32);
            }
        }

        // Phase 2: Triangulate all cell layers in the buffer (fast)
        if buffer_slice_count >= 2 {
            let cell_layers_in_buffer = buffer_slice_count - 1;
            for local_iz in 0..cell_layers_in_buffer {
                let global_iz = buffer_start_iz + local_iz;
                if global_iz >= cells {
                    break;
                }
                let lo_offset = local_iz * points_per_slice;
                let hi_offset = (local_iz + 1) * points_per_slice;
                let slice_lo = &buffer[lo_offset..lo_offset + points_per_slice];
                let slice_hi = &buffer[hi_offset..hi_offset + points_per_slice];

                process_slice_pair(
                    slice_lo,
                    slice_hi,
                    n,
                    global_iz,
                    origin_x,
                    origin_y,
                    origin_z,
                    cell_size,
                    &mut triangles,
                );
            }
        }

        // Phase 3: Keep last slice as overlap, reset buffer for next round
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
