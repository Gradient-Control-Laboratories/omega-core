package meshasset

// MeshSdfOracle holds the triangle soup for a bvhSdf(AssetID) leaf.
type MeshSdfOracle struct {
	AssetID   uint32
	Vertices  []float32
	Triangles []uint32
}

// VdbOracle holds the triangle soup and narrow-band level-set parameters for a vdbSdf(AssetID) leaf.
type VdbOracle struct {
	AssetID        uint32
	Vertices       []float32
	Triangles      []uint32
	VoxelSize      float32
	HalfBandVoxels float32
}

