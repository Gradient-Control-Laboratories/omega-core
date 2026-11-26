package compiler

import (
	"github.com/Gradient-Control-Laboratories/omega-core/assembler"
	"github.com/Gradient-Control-Laboratories/omega-core/assembler/meshplan"
	"github.com/Gradient-Control-Laboratories/omega-core/assembler/meshslots"
	"github.com/Gradient-Control-Laboratories/omega-core/assembler/variation"
	"github.com/Gradient-Control-Laboratories/omega-core/backend/mesh"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/compile"
	"github.com/Gradient-Control-Laboratories/omega-core/internal/meshimport"
	"github.com/Gradient-Control-Laboratories/omega-core/program"
)

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
	// as on compile.Request.
	Meshes meshimport.Finder `json:"-"`
	// Bodies, when not nil, is the 1-based returns MeshBodies meshes (the
	// viewer's visible bodies); Mesh ignores it.
	Bodies []int `json:"bodies,omitempty"`
}

// Mesh assembles the GSL source and evaluates a binary STL mesh using the Fidget/Manifold mesh backend.
func Mesh(req MeshRequest, sourcePaths []string, progressFn compile.ProgressFunc) ([]byte, error) {
	job, err := meshSetup(req, sourcePaths, progressFn)
	if err != nil {
		return nil, err
	}
	return withImportedMeshes(job.uses, func() ([]byte, error) {
		return mesh.Mesh(job.prog, job.opts...)
	})
}

// MeshBodies is Mesh with each top-level return meshed as its own body (mesh.MeshBodies). A document whose
// returns place imported meshes (mesh) or combine parts with mesh booleans (meshUnion, meshIntersect,
// meshDifference) is meshed by its mesh plan instead, part by part and composed with Manifold
// (mesh.MeshPlanBodies), as the viewer's browser export composes it.
func MeshBodies(req MeshRequest, sourcePaths []string, progressFn compile.ProgressFunc) ([]mesh.Body, error) {
	job, err := meshSetup(req, sourcePaths, progressFn)
	if err != nil {
		return nil, err
	}
	plan, subtrees, err := meshplan.SliceDocument(job.prog, "leaf-", req.Source)
	if err != nil || !mesh.PlanComposes(plan) {
		// A plan the slicer cannot make, or one every body of which is a single field, is meshed as before.
		return withImportedMeshes(job.uses, func() ([]mesh.Body, error) {
			return mesh.MeshBodies(job.prog, job.opts...)
		})
	}
	assets := func(index int) ([]float32, error) {
		asset, err := meshimport.Slot(job.prog, job.slots, req.Meshes, index)
		if err != nil {
			return nil, err
		}
		return asset.Vertices, nil
	}
	return withImportedMeshes(job.uses, func() ([]mesh.Body, error) {
		return mesh.MeshPlanBodies(job.prog, plan, subtrees, assets, job.opts...)
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
	compile.MeshesMu.Lock()
	defer compile.MeshesMu.Unlock()
	if err := meshimport.Register(uses); err != nil {
		var zero T
		return zero, err
	}
	return meshFn()
}

// meshJob is what meshSetup made of a request: the program, the mesher options, the imported meshes its distance
// builtins read, and the assembler's table of expression-form imports.
type meshJob struct {
	prog  *program.Program
	opts  []mesh.Option
	uses  []meshimport.Use
	slots []meshslots.Import
}

// meshSetup assembles req's source, resolves the meshes it imports (from
// req.Meshes), and builds the mesher options it asks for.
func meshSetup(req MeshRequest, sourcePaths []string, progressFn compile.ProgressFunc) (*meshJob, error) {
	asmRes, err := assembler.Assemble([]string{req.Source}, compile.AssemblerOptions(sourcePaths, progressFn)...)
	if err != nil {
		return nil, err
	}
	prog := asmRes.Program
	if len(asmRes.AudioReturns) > 0 {
		prog = variation.PrunePrimary(prog)
	}
	var uses []meshimport.Use
	if meshimport.References(prog) {
		if req.Meshes == nil {
			return nil, compile.ErrNoMeshFolder
		}
		if uses, err = meshimport.Resolve(prog, asmRes.Slots, req.Meshes); err != nil {
			return nil, err
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

	return &meshJob{prog: prog, opts: meshOpts, uses: uses, slots: asmRes.Slots}, nil
}
