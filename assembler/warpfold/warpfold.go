package warpfold

// WarpResolver resolves simulation files and bakes or resolves derived warped mesh slots.
type WarpResolver interface {
	ResolveWarpedSlot(meshSlot int, pkgDir, warpPath string, scale float64) (int, error)
}

// WarpedAsset records a derived warped mesh slot resolved during FoldProgram so callers
// can materialize or cache the deformed mesh asset at DerivedSlot.
type WarpedAsset struct {
	DerivedSlot int
	MeshSlot    int
	WarpDir     string
	WarpPath    string
	Scale       float64
}

