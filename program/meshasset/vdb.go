package meshasset

// VdbDefaultResolution sets the bake voxel count along the longest axis of
// the source mesh. 256 gives a meaningful accuracy step up from the
// 128-resolution dense Go-side bake while keeping memory bounded for
// reasonable meshes (256³ × narrow-band-of-3 ≈ 6 MB peak per asset).
const VdbDefaultResolution = 256

// VdbDefaultHalfBandVoxels is the narrow-band half-width OpenVDB carves
// around the surface, expressed in voxels. 3 matches the openvdb tutorial
// default and is wide enough for the trilinear sampler to give smooth
// gradients near the iso-surface.
const VdbDefaultHalfBandVoxels = 3.0

// ResolveVdbParams returns the (voxelSize, halfBandVoxels) pair to use
// for an OpenVDB bake of the given mesh, falling back to defaults when
// the GSL call site didn't supply explicit values. Every path that bakes
// or exports a vdbSdf oracle — the canvas mesh-out and preview paths and
// headless tapegen — resolves through here, so they all key into the same
// baked grid and write the same bake inputs into a tape.
//
// refDivisor is the integer count of voxels along the mesh's longest
// bbox axis (vdbSdf's optional 2nd argument). 0 means "use the default
// resolution"; the world-space voxel size is bbox-longest / divisor.
func ResolveVdbParams(mesh *TriMesh, refDivisor int, refHalfBand float64) (voxelSize, halfBandVoxels float32) {
	divisor := refDivisor
	if divisor <= 0 {
		divisor = VdbDefaultResolution
	}
	size := mesh.Bbox.Size()
	longest := size[0]
	if size[1] > longest {
		longest = size[1]
	}
	if size[2] > longest {
		longest = size[2]
	}
	voxelSize = float32(longest / float64(divisor))
	if voxelSize <= 0 {
		voxelSize = 0.01 // degenerate mesh fallback
	}
	halfBandVoxels = float32(refHalfBand)
	if halfBandVoxels <= 0 {
		halfBandVoxels = VdbDefaultHalfBandVoxels
	}
	return
}
