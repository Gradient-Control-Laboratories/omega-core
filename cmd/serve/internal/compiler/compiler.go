package compiler

import (
	"gradientcontrol.com/omega/core/assembler"
	"gradientcontrol.com/omega/core/assembler/variation"
	"gradientcontrol.com/omega/core/backend/mesh"
	"gradientcontrol.com/omega/core/internal/fieldbundle"
	"gradientcontrol.com/omega/core/internal/meshimport"
	"gradientcontrol.com/omega/core/program"
)

// The compile half lives in internal/fieldbundle, which gsl compile -t field
// shares, so /api/compile and a field bundle written by the CLI are the same.

// ProgressFunc reports compilation or meshing progress milestones.
type ProgressFunc = fieldbundle.ProgressFunc

// CompileRequest describes an interactive GSL compilation request.
type CompileRequest = fieldbundle.CompileRequest

// AudioResult contains compiled Web Audio synthesis JavaScript and duration.
type AudioResult = fieldbundle.AudioResult

// CompileResponse contains the compiled WGSL shader, UI metadata, optional audio, or structured diagnostics.
type CompileResponse = fieldbundle.CompileResponse

// Compile executes the in-process GSL compilation pipeline (fieldbundle.Compile).
func Compile(req CompileRequest, sourcePaths []string, progressFn ProgressFunc) *CompileResponse {
	return fieldbundle.Compile(req, sourcePaths, progressFn)
}

// UnwrapProgramErrors enriches and flattens an error tree into structured program.Error values.
func UnwrapProgramErrors(src string, err error) []program.Error {
	return fieldbundle.UnwrapProgramErrors(src, err)
}

// MeshRequest describes a server-side binary STL meshing request.
type MeshRequest struct {
	ID          string     `json:"id"`
	ClientID    string     `json:"clientId"`
	Source      string     `json:"source"`
	Subdivision int        `json:"subdivision"`
	Bbox        [6]float32 `json:"bbox"`
	HasBbox     bool       `json:"-"`
	Algo        string     `json:"algo"`
	Time        float64    `json:"time"`
	// Meshes finds the STL files the source's importMesh declarations name,
	// as on CompileRequest.
	Meshes meshimport.Finder `json:"-"`
	// Bodies, when not nil, is the 1-based returns MeshBodies meshes (the
	// viewer's visible bodies); Mesh ignores it.
	Bodies []int `json:"bodies,omitempty"`
}

// Mesh assembles the GSL source and evaluates a binary STL mesh using the Fidget/Manifold mesh backend.
func Mesh(req MeshRequest, sourcePaths []string, progressFn ProgressFunc) ([]byte, error) {
	prog, meshOpts, uses, err := meshSetup(req, sourcePaths, progressFn)
	if err != nil {
		return nil, err
	}
	return withImportedMeshes(uses, func() ([]byte, error) {
		return mesh.Mesh(prog, meshOpts...)
	})
}

// MeshBodies is Mesh with each top-level return meshed as its own body (mesh.MeshBodies).
func MeshBodies(req MeshRequest, sourcePaths []string, progressFn ProgressFunc) ([]mesh.Body, error) {
	prog, meshOpts, uses, err := meshSetup(req, sourcePaths, progressFn)
	if err != nil {
		return nil, err
	}
	return withImportedMeshes(uses, func() ([]mesh.Body, error) {
		return mesh.MeshBodies(prog, meshOpts...)
	})
}

// withImportedMeshes runs meshFn with the program's imported meshes (uses) in
// fidget's oracle table. That table is one per process, so it is held from
// registering this document's meshes to the end of meshing, every body of
// it, so another document's meshes in the same slots cannot replace them
// halfway. With no imported meshes it runs meshFn as it is.
func withImportedMeshes[T any](uses []meshimport.Use, meshFn func() (T, error)) (T, error) {
	if len(uses) == 0 {
		return meshFn()
	}
	fieldbundle.MeshesMu.Lock()
	defer fieldbundle.MeshesMu.Unlock()
	if err := meshimport.Register(uses); err != nil {
		var zero T
		return zero, err
	}
	return meshFn()
}

// meshSetup assembles req's source, resolves the meshes it imports (from
// req.Meshes), and builds the mesher options it asks for.
func meshSetup(req MeshRequest, sourcePaths []string, progressFn ProgressFunc) (*program.Program, []mesh.Option, []meshimport.Use, error) {
	asmRes, err := assembler.Assemble([]string{req.Source}, fieldbundle.AssemblerOptions(sourcePaths, progressFn)...)
	if err != nil {
		return nil, nil, nil, err
	}
	prog := asmRes.Program
	if len(asmRes.AudioReturns) > 0 {
		prog = variation.PrunePrimary(prog)
	}
	var uses []meshimport.Use
	if meshimport.References(prog) {
		if req.Meshes == nil {
			return nil, nil, nil, fieldbundle.ErrNoMeshFolder
		}
		if uses, err = meshimport.Resolve(prog, asmRes.Slots, req.Meshes); err != nil {
			return nil, nil, nil, err
		}
	}

	subdiv := req.Subdivision
	if subdiv <= 0 {
		subdiv = 4
	}
	algo := req.Algo
	if algo == "" {
		algo = "marching-cubes"
	}

	meshOpts := []mesh.Option{
		mesh.WithSubdivision(subdiv),
		mesh.WithAlgorithm(algo),
		mesh.WithTime(req.Time),
	}
	if req.HasBbox {
		meshOpts = append(meshOpts, mesh.WithBbox(req.Bbox))
	}
	if req.Bodies != nil {
		meshOpts = append(meshOpts, mesh.WithBodies(req.Bodies...))
	}
	if progressFn != nil {
		meshOpts = append(meshOpts, mesh.WithProgress(func(stage string, completed, total uint32) {
			progressFn("mesh", stage, completed, total)
		}))
	}

	return prog, meshOpts, uses, nil
}
