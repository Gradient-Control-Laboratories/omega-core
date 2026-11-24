/// Marching cubes algorithm for isosurface extraction from a regular 3D grid of SDF values.
///
/// This implementation uses a streaming slice-by-slice approach: only two adjacent
/// Z-slices are kept in memory at any time, reducing memory from O(n^3) to O(n^2).
/// This allows depth 12 (4097^3 grid) to use ~134 MB instead of ~275 GB.

/// For each of the 256 possible vertex sign configurations, a bitmask indicating
/// which of the 12 edges are intersected by the isosurface.
pub const EDGE_TABLE: [u16; 256] = [
    0x000, 0x109, 0x203, 0x30a, 0x406, 0x50f, 0x605, 0x70c,
    0x80c, 0x905, 0xa0f, 0xb06, 0xc0a, 0xd03, 0xe09, 0xf00,
    0x190, 0x099, 0x393, 0x29a, 0x596, 0x49f, 0x795, 0x69c,
    0x99c, 0x895, 0xb9f, 0xa96, 0xd9a, 0xc93, 0xf99, 0xe90,
    0x230, 0x339, 0x033, 0x13a, 0x636, 0x73f, 0x435, 0x53c,
    0xa3c, 0xb35, 0x83f, 0x936, 0xe3a, 0xf33, 0xc39, 0xd30,
    0x3a0, 0x2a9, 0x1a3, 0x0aa, 0x7a6, 0x6af, 0x5a5, 0x4ac,
    0xbac, 0xaa5, 0x9af, 0x8a6, 0xfaa, 0xea3, 0xda9, 0xca0,
    0x460, 0x569, 0x663, 0x76a, 0x066, 0x16f, 0x265, 0x36c,
    0xc6c, 0xd65, 0xe6f, 0xf66, 0x86a, 0x963, 0xa69, 0xb60,
    0x5f0, 0x4f9, 0x7f3, 0x6fa, 0x1f6, 0x0ff, 0x3f5, 0x2fc,
    0xdfc, 0xcf5, 0xfff, 0xef6, 0x9fa, 0x8f3, 0xbf9, 0xaf0,
    0x650, 0x759, 0x453, 0x55a, 0x256, 0x35f, 0x055, 0x15c,
    0xe5c, 0xf55, 0xc5f, 0xd56, 0xa5a, 0xb53, 0x859, 0x950,
    0x7c0, 0x6c9, 0x5c3, 0x4ca, 0x3c6, 0x2cf, 0x1c5, 0x0cc,
    0xfcc, 0xec5, 0xdcf, 0xcc6, 0xbca, 0xac3, 0x9c9, 0x8c0,
    0x8c0, 0x9c9, 0xac3, 0xbca, 0xcc6, 0xdcf, 0xec5, 0xfcc,
    0x0cc, 0x1c5, 0x2cf, 0x3c6, 0x4ca, 0x5c3, 0x6c9, 0x7c0,
    0x950, 0x859, 0xb53, 0xa5a, 0xd56, 0xc5f, 0xf55, 0xe5c,
    0x15c, 0x055, 0x35f, 0x256, 0x55a, 0x453, 0x759, 0x650,
    0xaf0, 0xbf9, 0x8f3, 0x9fa, 0xef6, 0xfff, 0xcf5, 0xdfc,
    0x2fc, 0x3f5, 0x0ff, 0x1f6, 0x6fa, 0x7f3, 0x4f9, 0x5f0,
    0xb60, 0xa69, 0x963, 0x86a, 0xf66, 0xe6f, 0xd65, 0xc6c,
    0x36c, 0x265, 0x16f, 0x066, 0x76a, 0x663, 0x569, 0x460,
    0xca0, 0xda9, 0xea3, 0xfaa, 0x8a6, 0x9af, 0xaa5, 0xbac,
    0x4ac, 0x5a5, 0x6af, 0x7a6, 0x0aa, 0x1a3, 0x2a9, 0x3a0,
    0xd30, 0xc39, 0xf33, 0xe3a, 0x936, 0x83f, 0xb35, 0xa3c,
    0x53c, 0x435, 0x73f, 0x636, 0x13a, 0x033, 0x339, 0x230,
    0xe90, 0xf99, 0xc93, 0xd9a, 0xa96, 0xb9f, 0x895, 0x99c,
    0x69c, 0x795, 0x49f, 0x596, 0x29a, 0x393, 0x099, 0x190,
    0xf00, 0xe09, 0xd03, 0xc0a, 0xb06, 0xa0f, 0x905, 0x80c,
    0x70c, 0x605, 0x50f, 0x406, 0x30a, 0x203, 0x109, 0x000,
];

/// For each of the 256 configurations, up to 5 triangles (15 edge indices), terminated by -1.
pub const TRI_TABLE: [[i8; 16]; 256] = [
    [-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 0, 8, 3,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 0, 1, 9,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 1, 8, 3, 9, 8, 1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 1, 2,10,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 0, 8, 3, 1, 2,10,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 9, 2,10, 0, 2, 9,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 2, 8, 3, 2,10, 8,10, 9, 8,-1,-1,-1,-1,-1,-1,-1],
    [ 3,11, 2,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 0,11, 2, 8,11, 0,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 1, 9, 0, 2, 3,11,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 1,11, 2, 1, 9,11, 9, 8,11,-1,-1,-1,-1,-1,-1,-1],
    [ 3,10, 1,11,10, 3,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 0,10, 1, 0, 8,10, 8,11,10,-1,-1,-1,-1,-1,-1,-1],
    [ 3, 9, 0, 3,11, 9,11,10, 9,-1,-1,-1,-1,-1,-1,-1],
    [ 9, 8,10,10, 8,11,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 4, 7, 8,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 4, 3, 0, 7, 3, 4,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 0, 1, 9, 8, 4, 7,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 4, 1, 9, 4, 7, 1, 7, 3, 1,-1,-1,-1,-1,-1,-1,-1],
    [ 1, 2,10, 8, 4, 7,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 3, 4, 7, 3, 0, 4, 1, 2,10,-1,-1,-1,-1,-1,-1,-1],
    [ 9, 2,10, 9, 0, 2, 8, 4, 7,-1,-1,-1,-1,-1,-1,-1],
    [ 2,10, 9, 2, 9, 7, 2, 7, 3, 7, 9, 4,-1,-1,-1,-1],
    [ 8, 4, 7, 3,11, 2,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [11, 4, 7,11, 2, 4, 2, 0, 4,-1,-1,-1,-1,-1,-1,-1],
    [ 9, 0, 1, 8, 4, 7, 2, 3,11,-1,-1,-1,-1,-1,-1,-1],
    [ 4, 7,11, 9, 4,11, 9,11, 2, 9, 2, 1,-1,-1,-1,-1],
    [ 3,10, 1, 3,11,10, 7, 8, 4,-1,-1,-1,-1,-1,-1,-1],
    [ 1,11,10, 1, 4,11, 1, 0, 4, 7,11, 4,-1,-1,-1,-1],
    [ 4, 7, 8, 9, 0,11, 9,11,10,11, 0, 3,-1,-1,-1,-1],
    [ 4, 7,11, 4,11, 9, 9,11,10,-1,-1,-1,-1,-1,-1,-1],
    [ 9, 5, 4,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 9, 5, 4, 0, 8, 3,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 0, 5, 4, 1, 5, 0,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 8, 5, 4, 8, 3, 5, 3, 1, 5,-1,-1,-1,-1,-1,-1,-1],
    [ 1, 2,10, 9, 5, 4,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 3, 0, 8, 1, 2,10, 4, 9, 5,-1,-1,-1,-1,-1,-1,-1],
    [ 5, 2,10, 5, 4, 2, 4, 0, 2,-1,-1,-1,-1,-1,-1,-1],
    [ 2,10, 5, 3, 2, 5, 3, 5, 4, 3, 4, 8,-1,-1,-1,-1],
    [ 9, 5, 4, 2, 3,11,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 0,11, 2, 0, 8,11, 4, 9, 5,-1,-1,-1,-1,-1,-1,-1],
    [ 0, 5, 4, 0, 1, 5, 2, 3,11,-1,-1,-1,-1,-1,-1,-1],
    [ 2, 1, 5, 2, 5, 8, 2, 8,11, 4, 8, 5,-1,-1,-1,-1],
    [10, 3,11,10, 1, 3, 9, 5, 4,-1,-1,-1,-1,-1,-1,-1],
    [ 4, 9, 5, 0, 8, 1, 8,10, 1, 8,11,10,-1,-1,-1,-1],
    [ 5, 4, 0, 5, 0,11, 5,11,10,11, 0, 3,-1,-1,-1,-1],
    [ 5, 4, 8, 5, 8,10,10, 8,11,-1,-1,-1,-1,-1,-1,-1],
    [ 9, 7, 8, 5, 7, 9,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 9, 3, 0, 9, 5, 3, 5, 7, 3,-1,-1,-1,-1,-1,-1,-1],
    [ 0, 7, 8, 0, 1, 7, 1, 5, 7,-1,-1,-1,-1,-1,-1,-1],
    [ 1, 5, 3, 3, 5, 7,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 9, 7, 8, 9, 5, 7,10, 1, 2,-1,-1,-1,-1,-1,-1,-1],
    [10, 1, 2, 9, 5, 0, 5, 3, 0, 5, 7, 3,-1,-1,-1,-1],
    [ 8, 0, 2, 8, 2, 5, 8, 5, 7,10, 5, 2,-1,-1,-1,-1],
    [ 2,10, 5, 2, 5, 3, 3, 5, 7,-1,-1,-1,-1,-1,-1,-1],
    [ 7, 9, 5, 7, 8, 9, 3,11, 2,-1,-1,-1,-1,-1,-1,-1],
    [ 9, 5, 7, 9, 7, 2, 9, 2, 0, 2, 7,11,-1,-1,-1,-1],
    [ 2, 3,11, 0, 1, 8, 1, 7, 8, 1, 5, 7,-1,-1,-1,-1],
    [11, 2, 1,11, 1, 7, 7, 1, 5,-1,-1,-1,-1,-1,-1,-1],
    [ 9, 5, 8, 8, 5, 7,10, 1, 3,10, 3,11,-1,-1,-1,-1],
    [ 5, 7, 0, 5, 0, 9, 7,11, 0, 1, 0,10,11,10, 0,-1],
    [11,10, 0,11, 0, 3,10, 5, 0, 8, 0, 7, 5, 7, 0,-1],
    [11,10, 5, 7,11, 5,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [10, 6, 5,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 0, 8, 3, 5,10, 6,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 9, 0, 1, 5,10, 6,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 1, 8, 3, 1, 9, 8, 5,10, 6,-1,-1,-1,-1,-1,-1,-1],
    [ 1, 6, 5, 2, 6, 1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 1, 6, 5, 1, 2, 6, 3, 0, 8,-1,-1,-1,-1,-1,-1,-1],
    [ 9, 6, 5, 9, 0, 6, 0, 2, 6,-1,-1,-1,-1,-1,-1,-1],
    [ 5, 9, 8, 5, 8, 2, 5, 2, 6, 3, 2, 8,-1,-1,-1,-1],
    [ 2, 3,11,10, 6, 5,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [11, 0, 8,11, 2, 0,10, 6, 5,-1,-1,-1,-1,-1,-1,-1],
    [ 0, 1, 9, 2, 3,11, 5,10, 6,-1,-1,-1,-1,-1,-1,-1],
    [ 5,10, 6, 1, 9, 2, 9,11, 2, 9, 8,11,-1,-1,-1,-1],
    [ 6, 3,11, 6, 5, 3, 5, 1, 3,-1,-1,-1,-1,-1,-1,-1],
    [ 0, 8,11, 0,11, 5, 0, 5, 1, 5,11, 6,-1,-1,-1,-1],
    [ 3,11, 6, 0, 3, 6, 0, 6, 5, 0, 5, 9,-1,-1,-1,-1],
    [ 6, 5, 9, 6, 9,11,11, 9, 8,-1,-1,-1,-1,-1,-1,-1],
    [ 5,10, 6, 4, 7, 8,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 4, 3, 0, 4, 7, 3, 6, 5,10,-1,-1,-1,-1,-1,-1,-1],
    [ 1, 9, 0, 5,10, 6, 8, 4, 7,-1,-1,-1,-1,-1,-1,-1],
    [10, 6, 5, 1, 9, 7, 1, 7, 3, 7, 9, 4,-1,-1,-1,-1],
    [ 6, 1, 2, 6, 5, 1, 4, 7, 8,-1,-1,-1,-1,-1,-1,-1],
    [ 1, 2, 5, 5, 2, 6, 3, 0, 4, 3, 4, 7,-1,-1,-1,-1],
    [ 8, 4, 7, 9, 0, 5, 0, 6, 5, 0, 2, 6,-1,-1,-1,-1],
    [ 7, 3, 9, 7, 9, 4, 3, 2, 9, 5, 9, 6, 2, 6, 9,-1],
    [ 3,11, 2, 7, 8, 4,10, 6, 5,-1,-1,-1,-1,-1,-1,-1],
    [ 5,10, 6, 4, 7, 2, 4, 2, 0, 2, 7,11,-1,-1,-1,-1],
    [ 0, 1, 9, 4, 7, 8, 2, 3,11, 5,10, 6,-1,-1,-1,-1],
    [ 9, 2, 1, 9,11, 2, 9, 4,11, 7,11, 4, 5,10, 6,-1],
    [ 8, 4, 7, 3,11, 5, 3, 5, 1, 5,11, 6,-1,-1,-1,-1],
    [ 5, 1,11, 5,11, 6, 1, 0,11, 7,11, 4, 0, 4,11,-1],
    [ 0, 5, 9, 0, 6, 5, 0, 3, 6,11, 6, 3, 8, 4, 7,-1],
    [ 6, 5, 9, 6, 9,11, 4, 7, 9, 7,11, 9,-1,-1,-1,-1],
    [10, 4, 9, 6, 4,10,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 4,10, 6, 4, 9,10, 0, 8, 3,-1,-1,-1,-1,-1,-1,-1],
    [10, 0, 1,10, 6, 0, 6, 4, 0,-1,-1,-1,-1,-1,-1,-1],
    [ 8, 3, 1, 8, 1, 6, 8, 6, 4, 6, 1,10,-1,-1,-1,-1],
    [ 1, 4, 9, 1, 2, 4, 2, 6, 4,-1,-1,-1,-1,-1,-1,-1],
    [ 3, 0, 8, 1, 2, 9, 2, 4, 9, 2, 6, 4,-1,-1,-1,-1],
    [ 0, 2, 4, 4, 2, 6,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 8, 3, 2, 8, 2, 4, 4, 2, 6,-1,-1,-1,-1,-1,-1,-1],
    [10, 4, 9,10, 6, 4,11, 2, 3,-1,-1,-1,-1,-1,-1,-1],
    [ 0, 8, 2, 2, 8,11, 4, 9,10, 4,10, 6,-1,-1,-1,-1],
    [ 3,11, 2, 0, 1, 6, 0, 6, 4, 6, 1,10,-1,-1,-1,-1],
    [ 6, 4, 1, 6, 1,10, 4, 8, 1, 2, 1,11, 8,11, 1,-1],
    [ 9, 6, 4, 9, 3, 6, 9, 1, 3,11, 6, 3,-1,-1,-1,-1],
    [ 8,11, 1, 8, 1, 0,11, 6, 1, 9, 1, 4, 6, 4, 1,-1],
    [ 3,11, 6, 3, 6, 0, 0, 6, 4,-1,-1,-1,-1,-1,-1,-1],
    [ 6, 4, 8,11, 6, 8,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 7,10, 6, 7, 8,10, 8, 9,10,-1,-1,-1,-1,-1,-1,-1],
    [ 0, 7, 3, 0,10, 7, 0, 9,10, 6, 7,10,-1,-1,-1,-1],
    [10, 6, 7, 1,10, 7, 1, 7, 8, 1, 8, 0,-1,-1,-1,-1],
    [10, 6, 7,10, 7, 1, 1, 7, 3,-1,-1,-1,-1,-1,-1,-1],
    [ 1, 2, 6, 1, 6, 8, 1, 8, 9, 8, 6, 7,-1,-1,-1,-1],
    [ 2, 6, 9, 2, 9, 1, 6, 7, 9, 0, 9, 3, 7, 3, 9,-1],
    [ 7, 8, 0, 7, 0, 6, 6, 0, 2,-1,-1,-1,-1,-1,-1,-1],
    [ 7, 3, 2, 6, 7, 2,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 2, 3,11,10, 6, 8,10, 8, 9, 8, 6, 7,-1,-1,-1,-1],
    [ 2, 0, 7, 2, 7,11, 0, 9, 7, 6, 7,10, 9,10, 7,-1],
    [ 1, 8, 0, 1, 7, 8, 1,10, 7, 6, 7,10, 2, 3,11,-1],
    [11, 2, 1,11, 1, 7,10, 6, 1, 6, 7, 1,-1,-1,-1,-1],
    [ 8, 9, 6, 8, 6, 7, 9, 1, 6,11, 6, 3, 1, 3, 6,-1],
    [ 0, 9, 1,11, 6, 7,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 7, 8, 0, 7, 0, 6, 3,11, 0,11, 6, 0,-1,-1,-1,-1],
    [ 7,11, 6,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 7, 6,11,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 3, 0, 8,11, 7, 6,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 0, 1, 9,11, 7, 6,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 8, 1, 9, 8, 3, 1,11, 7, 6,-1,-1,-1,-1,-1,-1,-1],
    [10, 1, 2, 6,11, 7,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 1, 2,10, 3, 0, 8, 6,11, 7,-1,-1,-1,-1,-1,-1,-1],
    [ 2, 9, 0, 2,10, 9, 6,11, 7,-1,-1,-1,-1,-1,-1,-1],
    [ 6,11, 7, 2,10, 3,10, 8, 3,10, 9, 8,-1,-1,-1,-1],
    [ 7, 2, 3, 6, 2, 7,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 7, 0, 8, 7, 6, 0, 6, 2, 0,-1,-1,-1,-1,-1,-1,-1],
    [ 2, 7, 6, 2, 3, 7, 0, 1, 9,-1,-1,-1,-1,-1,-1,-1],
    [ 1, 6, 2, 1, 8, 6, 1, 9, 8, 8, 7, 6,-1,-1,-1,-1],
    [10, 7, 6,10, 1, 7, 1, 3, 7,-1,-1,-1,-1,-1,-1,-1],
    [10, 7, 6, 1, 7,10, 1, 8, 7, 1, 0, 8,-1,-1,-1,-1],
    [ 0, 3, 7, 0, 7,10, 0,10, 9, 6,10, 7,-1,-1,-1,-1],
    [ 7, 6,10, 7,10, 8, 8,10, 9,-1,-1,-1,-1,-1,-1,-1],
    [ 6, 8, 4,11, 8, 6,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 3, 6,11, 3, 0, 6, 0, 4, 6,-1,-1,-1,-1,-1,-1,-1],
    [ 8, 6,11, 8, 4, 6, 9, 0, 1,-1,-1,-1,-1,-1,-1,-1],
    [ 9, 4, 6, 9, 6, 3, 9, 3, 1,11, 3, 6,-1,-1,-1,-1],
    [ 6, 8, 4, 6,11, 8, 2,10, 1,-1,-1,-1,-1,-1,-1,-1],
    [ 1, 2,10, 3, 0,11, 0, 6,11, 0, 4, 6,-1,-1,-1,-1],
    [ 4,11, 8, 4, 6,11, 0, 2, 9, 2,10, 9,-1,-1,-1,-1],
    [10, 9, 3,10, 3, 2, 9, 4, 3,11, 3, 6, 4, 6, 3,-1],
    [ 8, 2, 3, 8, 4, 2, 4, 6, 2,-1,-1,-1,-1,-1,-1,-1],
    [ 0, 4, 2, 4, 6, 2,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 1, 9, 0, 2, 3, 4, 2, 4, 6, 4, 3, 8,-1,-1,-1,-1],
    [ 1, 9, 4, 1, 4, 2, 2, 4, 6,-1,-1,-1,-1,-1,-1,-1],
    [ 8, 1, 3, 8, 6, 1, 8, 4, 6, 6,10, 1,-1,-1,-1,-1],
    [10, 1, 0,10, 0, 6, 6, 0, 4,-1,-1,-1,-1,-1,-1,-1],
    [ 4, 6, 3, 4, 3, 8, 6,10, 3, 0, 3, 9,10, 9, 3,-1],
    [10, 9, 4, 6,10, 4,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 4, 9, 5, 7, 6,11,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 0, 8, 3, 4, 9, 5,11, 7, 6,-1,-1,-1,-1,-1,-1,-1],
    [ 5, 0, 1, 5, 4, 0, 7, 6,11,-1,-1,-1,-1,-1,-1,-1],
    [11, 7, 6, 8, 3, 4, 3, 5, 4, 3, 1, 5,-1,-1,-1,-1],
    [ 9, 5, 4,10, 1, 2, 7, 6,11,-1,-1,-1,-1,-1,-1,-1],
    [ 6,11, 7, 1, 2,10, 0, 8, 3, 4, 9, 5,-1,-1,-1,-1],
    [ 7, 6,11, 5, 4,10, 4, 2,10, 4, 0, 2,-1,-1,-1,-1],
    [ 3, 4, 8, 3, 5, 4, 3, 2, 5,10, 5, 2,11, 7, 6,-1],
    [ 7, 2, 3, 7, 6, 2, 5, 4, 9,-1,-1,-1,-1,-1,-1,-1],
    [ 9, 5, 4, 0, 8, 6, 0, 6, 2, 6, 8, 7,-1,-1,-1,-1],
    [ 3, 6, 2, 3, 7, 6, 1, 5, 0, 5, 4, 0,-1,-1,-1,-1],
    [ 6, 2, 8, 6, 8, 7, 2, 1, 8, 4, 8, 5, 1, 5, 8,-1],
    [ 9, 5, 4,10, 1, 6, 1, 7, 6, 1, 3, 7,-1,-1,-1,-1],
    [ 1, 6,10, 1, 7, 6, 1, 0, 7, 8, 7, 0, 9, 5, 4,-1],
    [ 4, 0,10, 4,10, 5, 0, 3,10, 6,10, 7, 3, 7,10,-1],
    [ 7, 6,10, 7,10, 8, 5, 4,10, 4, 8,10,-1,-1,-1,-1],
    [ 6, 9, 5, 6,11, 9,11, 8, 9,-1,-1,-1,-1,-1,-1,-1],
    [ 3, 6,11, 0, 6, 3, 0, 5, 6, 0, 9, 5,-1,-1,-1,-1],
    [ 0,11, 8, 0, 5,11, 0, 1, 5, 5, 6,11,-1,-1,-1,-1],
    [ 6,11, 3, 6, 3, 5, 5, 3, 1,-1,-1,-1,-1,-1,-1,-1],
    [ 1, 2,10, 9, 5,11, 9,11, 8,11, 5, 6,-1,-1,-1,-1],
    [ 0,11, 3, 0, 6,11, 0, 9, 6, 5, 6, 9, 1, 2,10,-1],
    [11, 8, 5,11, 5, 6, 8, 0, 5,10, 5, 2, 0, 2, 5,-1],
    [ 6,11, 3, 6, 3, 5, 2,10, 3,10, 5, 3,-1,-1,-1,-1],
    [ 5, 8, 9, 5, 2, 8, 5, 6, 2, 3, 8, 2,-1,-1,-1,-1],
    [ 9, 5, 6, 9, 6, 0, 0, 6, 2,-1,-1,-1,-1,-1,-1,-1],
    [ 1, 5, 8, 1, 8, 0, 5, 6, 8, 3, 8, 2, 6, 2, 8,-1],
    [ 1, 5, 6, 2, 1, 6,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 1, 3, 6, 1, 6,10, 3, 8, 6, 5, 6, 9, 8, 9, 6,-1],
    [10, 1, 0,10, 0, 6, 9, 5, 0, 5, 6, 0,-1,-1,-1,-1],
    [ 0, 3, 8, 5, 6,10,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [10, 5, 6,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [11, 5,10, 7, 5,11,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [11, 5,10,11, 7, 5, 8, 3, 0,-1,-1,-1,-1,-1,-1,-1],
    [ 5,11, 7, 5,10,11, 1, 9, 0,-1,-1,-1,-1,-1,-1,-1],
    [10, 7, 5,10,11, 7, 9, 8, 1, 8, 3, 1,-1,-1,-1,-1],
    [11, 1, 2,11, 7, 1, 7, 5, 1,-1,-1,-1,-1,-1,-1,-1],
    [ 0, 8, 3, 1, 2, 7, 1, 7, 5, 7, 2,11,-1,-1,-1,-1],
    [ 9, 7, 5, 9, 2, 7, 9, 0, 2, 2,11, 7,-1,-1,-1,-1],
    [ 7, 5, 2, 7, 2,11, 5, 9, 2, 3, 2, 8, 9, 8, 2,-1],
    [ 2, 5,10, 2, 3, 5, 3, 7, 5,-1,-1,-1,-1,-1,-1,-1],
    [ 8, 2, 0, 8, 5, 2, 8, 7, 5,10, 2, 5,-1,-1,-1,-1],
    [ 9, 0, 1, 5,10, 3, 5, 3, 7, 3,10, 2,-1,-1,-1,-1],
    [ 9, 8, 2, 9, 2, 1, 8, 7, 2,10, 2, 5, 7, 5, 2,-1],
    [ 1, 3, 5, 3, 7, 5,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 0, 8, 7, 0, 7, 1, 1, 7, 5,-1,-1,-1,-1,-1,-1,-1],
    [ 9, 0, 3, 9, 3, 5, 5, 3, 7,-1,-1,-1,-1,-1,-1,-1],
    [ 9, 8, 7, 5, 9, 7,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 5, 8, 4, 5,10, 8,10,11, 8,-1,-1,-1,-1,-1,-1,-1],
    [ 5, 0, 4, 5,11, 0, 5,10,11,11, 3, 0,-1,-1,-1,-1],
    [ 0, 1, 9, 8, 4,10, 8,10,11,10, 4, 5,-1,-1,-1,-1],
    [10,11, 4,10, 4, 5,11, 3, 4, 9, 4, 1, 3, 1, 4,-1],
    [ 2, 5, 1, 2, 8, 5, 2,11, 8, 4, 5, 8,-1,-1,-1,-1],
    [ 0, 4,11, 0,11, 3, 4, 5,11, 2,11, 1, 5, 1,11,-1],
    [ 0, 2, 5, 0, 5, 9, 2,11, 5, 4, 5, 8,11, 8, 5,-1],
    [ 9, 4, 5, 2,11, 3,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 2, 5,10, 3, 5, 2, 3, 4, 5, 3, 8, 4,-1,-1,-1,-1],
    [ 5,10, 2, 5, 2, 4, 4, 2, 0,-1,-1,-1,-1,-1,-1,-1],
    [ 3,10, 2, 3, 5,10, 3, 8, 5, 4, 5, 8, 0, 1, 9,-1],
    [ 5,10, 2, 5, 2, 4, 1, 9, 2, 9, 4, 2,-1,-1,-1,-1],
    [ 8, 4, 5, 8, 5, 3, 3, 5, 1,-1,-1,-1,-1,-1,-1,-1],
    [ 0, 4, 5, 1, 0, 5,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 8, 4, 5, 8, 5, 3, 9, 0, 5, 0, 3, 5,-1,-1,-1,-1],
    [ 9, 4, 5,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 4,11, 7, 4, 9,11, 9,10,11,-1,-1,-1,-1,-1,-1,-1],
    [ 0, 8, 3, 4, 9, 7, 9,11, 7, 9,10,11,-1,-1,-1,-1],
    [ 1,10,11, 1,11, 4, 1, 4, 0, 7, 4,11,-1,-1,-1,-1],
    [ 3, 1, 4, 3, 4, 8, 1,10, 4, 7, 4,11,10,11, 4,-1],
    [ 4,11, 7, 9,11, 4, 9, 2,11, 9, 1, 2,-1,-1,-1,-1],
    [ 9, 7, 4, 9,11, 7, 9, 1,11, 2,11, 1, 0, 8, 3,-1],
    [11, 7, 4,11, 4, 2, 2, 4, 0,-1,-1,-1,-1,-1,-1,-1],
    [11, 7, 4,11, 4, 2, 8, 3, 4, 3, 2, 4,-1,-1,-1,-1],
    [ 2, 9,10, 2, 7, 9, 2, 3, 7, 7, 4, 9,-1,-1,-1,-1],
    [ 9,10, 7, 9, 7, 4,10, 2, 7, 8, 7, 0, 2, 0, 7,-1],
    [ 3, 7,10, 3,10, 2, 7, 4,10, 1,10, 0, 4, 0,10,-1],
    [ 1,10, 2, 8, 7, 4,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 4, 9, 1, 4, 1, 7, 7, 1, 3,-1,-1,-1,-1,-1,-1,-1],
    [ 4, 9, 1, 4, 1, 7, 0, 8, 1, 8, 7, 1,-1,-1,-1,-1],
    [ 4, 0, 3, 7, 4, 3,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 4, 8, 7,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 9,10, 8,10,11, 8,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 3, 0, 9, 3, 9,11,11, 9,10,-1,-1,-1,-1,-1,-1,-1],
    [ 0, 1,10, 0,10, 8, 8,10,11,-1,-1,-1,-1,-1,-1,-1],
    [ 3, 1,10,11, 3,10,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 1, 2,11, 1,11, 9, 9,11, 8,-1,-1,-1,-1,-1,-1,-1],
    [ 3, 0, 9, 3, 9,11, 1, 2, 9, 2,11, 9,-1,-1,-1,-1],
    [ 0, 2,11, 8, 0,11,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 3, 2,11,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 2, 3, 8, 2, 8,10,10, 8, 9,-1,-1,-1,-1,-1,-1,-1],
    [ 9,10, 2, 0, 9, 2,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 2, 3, 8, 2, 8,10, 0, 1, 8, 1,10, 8,-1,-1,-1,-1],
    [ 1,10, 2,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 1, 3, 8, 9, 1, 8,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 0, 9, 1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [ 0, 3, 8,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
    [-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1,-1],
];

/// The 8 corner offsets of a cube cell as (dx, dy, dz) in grid units.
pub const CORNER_OFFSETS: [(usize, usize, usize); 8] = [
    (0, 0, 0), // 0
    (1, 0, 0), // 1
    (1, 1, 0), // 2
    (0, 1, 0), // 3
    (0, 0, 1), // 4
    (1, 0, 1), // 5
    (1, 1, 1), // 6
    (0, 1, 1), // 7
];

/// For each of the 12 edges, the two corner indices that form it.
pub const EDGE_CORNERS: [(usize, usize); 12] = [
    (0, 1), // edge  0
    (1, 2), // edge  1
    (2, 3), // edge  2
    (3, 0), // edge  3
    (4, 5), // edge  4
    (5, 6), // edge  5
    (6, 7), // edge  6
    (7, 4), // edge  7
    (0, 4), // edge  8
    (1, 5), // edge  9
    (2, 6), // edge 10
    (3, 7), // edge 11
];

/// `EDGE_CORNERS` with every edge walked up its axis, from the corner nearer the cell's origin to the farther.
///
/// Edges 2, 3, 6 and 7 run down their axis in `EDGE_CORNERS`, so the cell on the other side of one of them reads
/// the same grid edge from its far end. The crossing `t = -v0 / (v1 - v0)` then rounds differently in the two
/// cells, and their copies of the one vertex differ in the last bit: the mesh has unmatched edges wherever such
/// a pair meet. Reading every edge up its axis makes each crossing a function of the grid edge alone.
///
/// `EDGE_CORNERS` itself stays as it is: the dual contouring mesher reads it too.
const UP_EDGE_CORNERS: [(usize, usize); 12] = {
    let mut out = EDGE_CORNERS;
    let mut e = 0;
    while e < 12 {
        let (a, b) = EDGE_CORNERS[e];
        let (ax, ay, az) = CORNER_OFFSETS[a];
        let (bx, by, bz) = CORNER_OFFSETS[b];
        if ax + ay + az > bx + by + bz {
            out[e] = (b, a);
        }
        e += 1;
    }
    out
};

/// Process a single layer of cells between two Z-slices,
/// appending triangles to the output buffer.
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

            // Compute the cube index from corner signs
            let mut cube_index: usize = 0;
            for i in 0..8 {
                if corner_vals[i] < 0.0 {
                    cube_index |= 1 << i;
                }
            }

            let edge_bits = EDGE_TABLE[cube_index];
            if edge_bits == 0 {
                continue;
            }

            // Compute edge intersection points
            let mut edge_verts = [[0.0f32; 3]; 12];
            for e in 0..12 {
                if edge_bits & (1 << e) != 0 {
                    let (c0, c1) = UP_EDGE_CORNERS[e];
                    let v0 = corner_vals[c0];
                    let v1 = corner_vals[c1];

                    let t = if (v1 - v0).abs() > 1e-10 {
                        -v0 / (v1 - v0)
                    } else {
                        0.5
                    };
                    let t = t.clamp(0.0, 1.0);

                    let (dx0, dy0, dz0) = CORNER_OFFSETS[c0];
                    let (dx1, dy1, dz1) = CORNER_OFFSETS[c1];

                    edge_verts[e] = [
                        origin_x + (ix as f32 + dx0 as f32 + t * (dx1 as f32 - dx0 as f32)) * cell_size,
                        origin_y + (iy as f32 + dy0 as f32 + t * (dy1 as f32 - dy0 as f32)) * cell_size,
                        origin_z + (iz as f32 + dz0 as f32 + t * (dz1 as f32 - dz0 as f32)) * cell_size,
                    ];
                }
            }

            // Emit triangles from the lookup table
            let tri_row = &TRI_TABLE[cube_index];
            let mut t = 0;
            while t < 15 && tri_row[t] >= 0 {
                let e0 = tri_row[t] as usize;
                let e1 = tri_row[t + 1] as usize;
                let e2 = tri_row[t + 2] as usize;

                // Bourke's table winds each triangle to face the corners whose bit is set, and a set bit is a
                // corner inside the solid (field < 0): taken as listed, every triangle faces inward and the mesh
                // is inside out. Emitted (e0, e2, e1), each faces outward, up the field's gradient, as the
                // dual-contouring mesher's do and as STL expects.
                let v0 = edge_verts[e0];
                let v1 = edge_verts[e2];
                let v2 = edge_verts[e1];

                // Compute face normal via cross product
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

                t += 3;
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

#[cfg(test)]
mod tests {
    use super::*;

    /// The triangles of `field` marched over the cube [-h, h]^3 at n points per axis.
    fn march(field: impl Fn(f32, f32, f32) -> f32, n: usize, h: f32) -> Vec<f32> {
        let cell = 2.0 * h / (n - 1) as f32;
        chunked_marching_cubes(
            n, -h, -h, -h, cell, n, n,
            |iz0, count| {
                let mut out = Vec::with_capacity(count * n * n);
                for iz in iz0..iz0 + count {
                    for iy in 0..n {
                        for ix in 0..n {
                            out.push(field(-h + ix as f32 * cell, -h + iy as f32 * cell, -h + iz as f32 * cell));
                        }
                    }
                }
                out
            },
            |_, _| {},
        )
    }

    /// The signed volume the triangles enclose, and how many triangles' stored normals point the way their
    /// winding does and away from the origin.
    fn orientation(tris: &[f32]) -> (f64, usize, usize, usize) {
        let (mut vol, mut agree, mut outward) = (0.0f64, 0, 0);
        for t in tris.chunks_exact(12) {
            let n = [t[0], t[1], t[2]];
            let (a, b, c) = ([t[3], t[4], t[5]], [t[6], t[7], t[8]], [t[9], t[10], t[11]]);
            vol += (a[0] * (b[1] * c[2] - b[2] * c[1]) - a[1] * (b[0] * c[2] - b[2] * c[0])
                + a[2] * (b[0] * c[1] - b[1] * c[0])) as f64 / 6.0;
            let u = [b[0] - a[0], b[1] - a[1], b[2] - a[2]];
            let v = [c[0] - a[0], c[1] - a[1], c[2] - a[2]];
            let w = [u[1] * v[2] - u[2] * v[1], u[2] * v[0] - u[0] * v[2], u[0] * v[1] - u[1] * v[0]];
            agree += (w[0] * n[0] + w[1] * n[1] + w[2] * n[2] > 0.0) as usize;
            let centroid = [(a[0] + b[0] + c[0]) / 3.0, (a[1] + b[1] + c[1]) / 3.0, (a[2] + b[2] + c[2]) / 3.0];
            outward += (centroid[0] * n[0] + centroid[1] * n[1] + centroid[2] * n[2] > 0.0) as usize;
        }
        (vol, agree, outward, tris.len() / 12)
    }

    // A solid's mesh faces out: a positive signed volume, each normal along its winding and away from the solid.
    #[test]
    fn a_sphere_marches_outward() {
        let tris = march(|x, y, z| (x * x + y * y + z * z).sqrt() - 1.0, 41, 1.25);
        let (vol, agree, outward, count) = orientation(&tris);
        let exact = 4.0 / 3.0 * std::f64::consts::PI;
        assert!((vol - exact).abs() < 0.02 * exact, "signed volume {vol}, want about {exact}");
        assert_eq!(agree, count, "every normal follows its winding");
        assert_eq!(outward, count, "every normal points away from the centre");
    }

    // A cavity's wall faces into the cavity: the hollow ball's volume is the shell's, positive.
    #[test]
    fn a_hollow_ball_marches_its_cavity_inward() {
        let tris = march(|x, y, z| {
            let r = (x * x + y * y + z * z).sqrt();
            (r - 1.0).max(0.5 - r)
        }, 49, 1.25);
        let (vol, agree, _, count) = orientation(&tris);
        let exact = 4.0 / 3.0 * std::f64::consts::PI * (1.0 - 0.125);
        assert!((vol - exact).abs() < 0.03 * exact, "signed volume {vol}, want about {exact}");
        assert_eq!(agree, count);
    }
}

#[cfg(test)]
mod watertight_tests {
    use super::*;
    use std::collections::{BTreeMap, BTreeSet, HashMap};

    /// The cell's six faces, each as its four corners in cyclic order.
    const FACES: [[usize; 4]; 6] = [[0, 3, 2, 1], [4, 5, 6, 7], [0, 1, 5, 4], [3, 7, 6, 2], [0, 4, 7, 3], [1, 2, 6, 5]];

    /// The cell edge joining corner `a` to corner `b`, whichever way round `EDGE_CORNERS` lists it.
    fn edge_between(a: usize, b: usize) -> usize {
        (0..12)
            .find(|&e| {
                let (c0, c1) = EDGE_CORNERS[e];
                (c0 == a && c1 == b) || (c0 == b && c1 == a)
            })
            .unwrap()
    }

    /// Whether a face's corners alternate in and out around it, in the cube index's bits.
    fn is_ambiguous(index: usize, face: &[usize; 4]) -> bool {
        let s = |k: usize| (index >> face[k]) & 1;
        s(0) == s(2) && s(1) == s(3) && s(0) != s(1)
    }

    /// A repeatable stream of numbers in [-1, 1) (xorshift64).
    struct Noise(u64);

    impl Noise {
        fn next(&mut self) -> f32 {
            self.0 ^= self.0 << 13;
            self.0 ^= self.0 >> 7;
            self.0 ^= self.0 << 17;
            ((self.0 >> 40) as f32 / (1u64 << 24) as f32) * 2.0 - 1.0
        }
    }

    /// The triangles of the n^3 grid `values` (indexed [iz * n * n + iy * n + ix]) over [0, n-1]^3 at unit spacing,
    /// fed through the streaming mesher `buffer` slices at a time and `batch` slices per evaluation.
    fn march_grid(values: &[f32], n: usize, buffer: usize, batch: usize) -> Vec<f32> {
        chunked_marching_cubes(
            n, 0.0, 0.0, 0.0, 1.0, buffer, batch,
            |iz0, count| values[iz0 * n * n..(iz0 + count) * n * n].to_vec(),
            |_, _| {},
        )
    }

    /// The cells of the grid that have a face whose corners alternate in and out.
    fn ambiguous_cells(values: &[f32], n: usize) -> usize {
        let mut count = 0;
        for iz in 0..n - 1 {
            for iy in 0..n - 1 {
                for ix in 0..n - 1 {
                    let mut index = 0;
                    for (i, (dx, dy, dz)) in CORNER_OFFSETS.iter().enumerate() {
                        if values[(iz + dz) * n * n + (iy + dy) * n + ix + dx] < 0.0 {
                            index |= 1 << i;
                        }
                    }
                    count += FACES.iter().any(|face| is_ambiguous(index, face)) as usize;
                }
            }
        }
        count
    }

    /// Asserts every directed edge of the triangles occurs once and its reverse once, matching vertices by their
    /// bits, and returns the signed volume the triangles enclose.
    fn assert_watertight(tris: &[f32], what: &str) -> f64 {
        let bits = |v: &[f32]| [v[0].to_bits(), v[1].to_bits(), v[2].to_bits()];
        let mut edges: HashMap<([u32; 3], [u32; 3]), usize> = HashMap::new();
        let mut volume = 0.0f64;
        for t in tris.chunks_exact(12) {
            let v = [bits(&t[3..6]), bits(&t[6..9]), bits(&t[9..12])];
            for k in 0..3 {
                *edges.entry((v[k], v[(k + 1) % 3])).or_insert(0) += 1;
            }
            let (a, b, c) = ([t[3] as f64, t[4] as f64, t[5] as f64], [t[6] as f64, t[7] as f64, t[8] as f64],
                [t[9] as f64, t[10] as f64, t[11] as f64]);
            volume += (a[0] * (b[1] * c[2] - b[2] * c[1]) - a[1] * (b[0] * c[2] - b[2] * c[0])
                + a[2] * (b[0] * c[1] - b[1] * c[0])) / 6.0;
        }
        let unmatched = edges.iter().filter(|(&(a, b), &count)| count != 1 || edges.get(&(b, a)) != Some(&1)).count();
        assert_eq!(unmatched, 0, "{what}: {unmatched} of {} directed edges lack a single opposite partner", edges.len());
        volume
    }

    // Every edge is read from its near end to its far end, so the two cells that share a grid edge compute its
    // crossing from the same two values in the same order.
    #[test]
    fn every_edge_is_read_up_its_axis() {
        for e in 0..12 {
            let (c0, c1) = UP_EDGE_CORNERS[e];
            let (a, b) = EDGE_CORNERS[e];
            assert!((c0, c1) == (a, b) || (c0, c1) == (b, a), "edge {e} joins the same corners");
            let (x0, y0, z0) = CORNER_OFFSETS[c0];
            let (x1, y1, z1) = CORNER_OFFSETS[c1];
            assert!(x0 + y0 + z0 < x1 + y1 + z1, "edge {e} runs up its axis");
        }
    }

    // What the surface's closure rests on, besides shared vertices: the table joins an ambiguous face's crossings
    // the same way in every cell, keeping the inside corners apart, so the two cells that share the face agree. A
    // table that sometimes joined them would need each face decided from its four values alone.
    #[test]
    fn the_table_keeps_every_ambiguous_face_apart() {
        let mut seen: BTreeMap<(usize, usize), BTreeSet<&str>> = BTreeMap::new();
        for index in 0..256usize {
            // The outline of the table's triangles, emitted (e0, e2, e1): an edge a neighbouring triangle runs
            // back cancels.
            let mut outline: BTreeSet<(usize, usize)> = BTreeSet::new();
            let row = &TRI_TABLE[index];
            let mut t = 0;
            while t < 15 && row[t] >= 0 {
                let tri = [row[t] as usize, row[t + 2] as usize, row[t + 1] as usize];
                for k in 0..3 {
                    let (a, b) = (tri[k], tri[(k + 1) % 3]);
                    if !outline.remove(&(b, a)) {
                        outline.insert((a, b));
                    }
                }
                t += 3;
            }
            for (f, face) in FACES.iter().enumerate() {
                if !is_ambiguous(index, face) {
                    continue;
                }
                let edges: [usize; 4] = std::array::from_fn(|k| edge_between(face[k], face[(k + 1) % 4]));
                // Face corner k sits between edges k - 1 and k, and is cut off by a segment joining them.
                let cut_off = |k: usize| {
                    let (a, b) = (edges[(k + 3) % 4], edges[k]);
                    outline.contains(&(a, b)) || outline.contains(&(b, a))
                };
                let inside: Vec<usize> = (0..4).filter(|&k| (index >> face[k]) & 1 == 1).collect();
                let outside: Vec<usize> = (0..4).filter(|&k| (index >> face[k]) & 1 == 0).collect();
                let way = match (inside.iter().all(|&k| cut_off(k)), outside.iter().all(|&k| cut_off(k))) {
                    (true, false) => "apart",
                    (false, true) => "joined",
                    _ => "neither",
                };
                seen.entry((f, inside[0] % 2)).or_default().insert(way);
            }
        }
        assert_eq!(seen.len(), 12, "all six faces occur, each with either diagonal inside");
        for (key, ways) in &seen {
            assert_eq!(ways, &BTreeSet::from(["apart"]), "face {} with its diagonal {} inside", key.0, key.1);
        }
    }

    // Noise has every kind of face in every arrangement. With the grid's border held outside the solid, its surface
    // is closed, so each of the mesh's edges must be shared by exactly two triangles, one running each way,
    // however the stream is chunked.
    #[test]
    fn a_noise_field_marches_into_a_closed_surface() {
        let mut noise = Noise(0xD1B5_4A32_D192_ED03);
        let mut ambiguous = 0;
        for trial in 0..300 {
            let n = 5 + trial % 6;
            let mut values = vec![0.0f32; n * n * n];
            for iz in 0..n {
                for iy in 0..n {
                    for ix in 0..n {
                        let border = [ix, iy, iz].iter().any(|&i| i == 0 || i == n - 1);
                        values[iz * n * n + iy * n + ix] = if border { 1.0 + noise.next().abs() } else { noise.next() };
                    }
                }
            }
            ambiguous += ambiguous_cells(&values, n);
            for (buffer, batch) in [(n, n), (3, 2), (2, 1)] {
                let tris = march_grid(&values, n, buffer, batch);
                let volume = assert_watertight(&tris, &format!("trial {trial}, n {n}, buffer {buffer}, batch {batch}"));
                assert!(volume >= 0.0, "trial {trial}: signed volume {volume} is inside out");
            }
        }
        assert!(ambiguous > 1000, "the noise reaches ambiguous faces: {ambiguous} cells");
    }

    // Smooth solids at the depths people mesh: a sphere and a hollow ball, their centres off the grid so no
    // vertex lands exactly on a grid point.
    #[test]
    fn smooth_solids_march_into_closed_surfaces() {
        let n = 65;
        let c = 0.5 * (n - 1) as f32 + 0.0137;
        let radius = |x: usize, y: usize, z: usize| {
            ((x as f32 - c).powi(2) + (y as f32 - c + 0.31).powi(2) + (z as f32 - c - 0.17).powi(2)).sqrt()
        };
        for (name, outer, inner) in [("sphere", 28.0f32, 0.0f32), ("hollow ball", 28.0, 16.0)] {
            let mut values = vec![0.0f32; n * n * n];
            for iz in 0..n {
                for iy in 0..n {
                    for ix in 0..n {
                        let r = radius(ix, iy, iz);
                        values[iz * n * n + iy * n + ix] = (r - outer).max(inner - r);
                    }
                }
            }
            let tris = march_grid(&values, n, n, n);
            let volume = assert_watertight(&tris, name);
            let want = 4.0 / 3.0 * std::f64::consts::PI * ((outer as f64).powi(3) - (inner as f64).powi(3));
            assert!((volume - want).abs() < 0.01 * want, "{name}: volume {volume}, want about {want}");
        }
    }
}
