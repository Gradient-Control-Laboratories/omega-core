//go:build cgo

package eval

import (
	"context"
	"fmt"
	"io"
	"math"
	"strings"

	"gradientcontrol.com/omega/core/backend/mesh"
	"gradientcontrol.com/omega/core/internal/fidget"
)

// The mesher settings Mesh takes when the caller names none: gsl compile's.
const (
	defaultMeshSubdivision = 9
	defaultMeshAlgorithm   = "mdc"
)

// MeshOptions is how Mesh meshes. The zero value is gsl compile's defaults over the box the field fills.
type MeshOptions struct {
	// Subdivision is the octree depth; 0 takes 9.
	Subdivision int
	// Region is the box meshed, [x0, y0, z0, x1, y1, z1] in mm; nil takes the returned value's .box when it is
	// bounded (ReturnBox), else the box the field fills, found by sampling as gsl compile
	// --bbox auto finds it, with defs applied.
	Region *[6]float64
	// Algorithm is "mdc" or "marching-cubes"; "" takes "mdc".
	Algorithm string
	// Time is the animation time, in seconds, the field is meshed at.
	Time float64
}

// MeshBody is one body of a mesh and what its triangles say of it.
type MeshBody struct {
	// Name and Index are the return's, as mesh.MeshBodies gives them: Index counts from 1.
	Name             string
	Index            int
	Vertices         int
	Triangles        int
	OpenEdges        int
	NonManifoldEdges int
	// VolumeMm3 is the volume the triangles enclose, negative if they face inward. It is the mesh's, not the
	// field's: Volume brackets the field's.
	VolumeMm3 float64
}

// MeshResult is what Mesh wrote: the box it meshed over and each body.
type MeshResult struct {
	Region [6]float64
	Bodies []MeshBody
}

// Mesh meshes the document, with defs applied as the other queries apply them, and writes it to w as one binary
// glTF: a node per top-level return, indexed, with a normal per vertex. A normal is the field's gradient at the
// vertex, so shading follows the surface and not the facets; where the gradient is not finite or is zero, the
// vertex takes the average of its triangles' facet normals instead. At a Time other than zero every normal is that
// average, because the point query that supplies gradients evaluates at time zero.
//
// The result reports, per body, the counts and the enclosed volume of the triangles written, so a caller can tell
// a closed mesh from an open one without reading the file back.
func (e *Evaluator) Mesh(ctx context.Context, manifest map[string]string, code, defs string, w io.Writer, opt MeshOptions) (*MeshResult, error) {
	exprs := TopLevelReturnExprs(code)
	if len(exprs) == 0 {
		return nil, fmt.Errorf("the document returns nothing to mesh")
	}
	subdivision := opt.Subdivision
	if subdivision == 0 {
		subdivision = defaultMeshSubdivision
	}
	if subdivision < 0 {
		return nil, fmt.Errorf("mesh subdivision %d: zero for the default, or more", subdivision)
	}
	algorithm := opt.Algorithm
	if algorithm == "" {
		algorithm = defaultMeshAlgorithm
	}

	var region [6]float64
	padCells := 0
	if opt.Region != nil {
		region = *opt.Region
		for k := 0; k < 3; k++ {
			if math.IsNaN(region[k]) || math.IsNaN(region[k+3]) || math.IsInf(region[k], 0) || math.IsInf(region[k+3], 0) || region[k] >= region[k+3] {
				return nil, fmt.Errorf("mesh region %v: finite, its max past its min on every axis", region)
			}
		}
	} else if declared, ok, err := e.ReturnBox(ctx, manifest, code, defs, exprs); err != nil {
		return nil, err
	} else if ok {
		// The returned value says where it is: mesh that, padded by the mesher by one of its own cells so a face
		// lying on the box is meshed.
		region, padCells = declared, 1
	} else {
		fit, err := e.fitBox(ctx, manifest, code, defs, exprs)
		if err != nil {
			return nil, err
		}
		region = fit
	}
	var box [6]float32
	for k := range box {
		box[k] = float32(region[k])
	}

	bodies, err := e.meshBodies(ctx, manifest, code, defs, exprs,
		mesh.WithSubdivision(subdivision), mesh.WithAlgorithm(algorithm), mesh.WithTime(opt.Time), mesh.WithBbox(box),
		mesh.WithBboxPadCells(padCells))
	if err != nil {
		return nil, err
	}

	// The region reported is the box the mesher meshed: the one given, with the room it padded it by.
	meshed := fidget.PadBbox(box, subdivision, padCells)
	for k := range meshed {
		region[k] = float64(meshed[k])
	}
	result := &MeshResult{Region: region, Bodies: make([]MeshBody, len(bodies))}
	for i, b := range bodies {
		// A body with no triangles is never offered to the normals callback below, so its counts stay zero.
		result.Bodies[i] = MeshBody{Name: b.Name, Index: b.Index}
	}
	byIndex := make(map[int]int, len(bodies))
	for i, b := range bodies {
		byIndex[b.Index] = i
	}
	normals := mesh.GLBNormals(func(body mesh.Body, positions []float32, indices []uint32) ([]float32, error) {
		stats := mesh.MeasureIndexed(positions, indices)
		slot := &result.Bodies[byIndex[body.Index]]
		slot.Vertices, slot.Triangles = stats.Vertices, stats.Triangles
		slot.OpenEdges, slot.NonManifoldEdges = stats.OpenEdges, stats.NonManifoldEdges
		slot.VolumeMm3 = stats.SignedVolume

		var gradients [][3]float32
		if opt.Time == 0 && body.Index >= 1 && body.Index <= len(exprs) {
			points := make([][3]float64, stats.Vertices)
			for v := range points {
				points[v] = [3]float64{float64(positions[3*v]), float64(positions[3*v+1]), float64(positions[3*v+2])}
			}
			groups, err := e.EvalGsl(ctx, manifest, code, defs, []string{exprs[body.Index-1]}, points)
			if err != nil {
				return nil, err
			}
			if len(groups) == 1 && len(groups[0].Samples) == 1 && len(groups[0].Samples[0]) == stats.Vertices {
				gradients = make([][3]float32, stats.Vertices)
				for v, s := range groups[0].Samples[0] {
					gradients[v] = [3]float32{s.Dx, s.Dy, s.Dz}
				}
			}
		}
		return vertexNormals(gradients, positions, indices), nil
	})
	if err := mesh.WriteGLB(w, bodies, nil, normals); err != nil {
		return nil, err
	}
	return result, nil
}

// meshBodies meshes each of the document's returns, defs applied, holding the query slot while the mesher reads
// the oracles the document registered.
func (e *Evaluator) meshBodies(ctx context.Context, manifest map[string]string, code, defs string, exprs []string, opts ...mesh.Option) ([]mesh.Body, error) {
	if err := startQuery(ctx); err != nil {
		return nil, err
	}
	defer endQuery()
	// The returns put back as the query's returns: the program gsl compile meshes, with defs applied.
	prog, err := e.prepareQuery(ctx, manifest, code, defs, exprs)
	if err != nil {
		return nil, err
	}
	return mesh.MeshBodies(prog, opts...)
}

// returnBoxReach is the largest coordinate a box ReturnBox gives may have: builtin's nothing (1e5) is how .box reads
// an axis the analysis cannot bound, so anything at half of it or beyond is taken as unbounded.
const returnBoxReach = 5e4

// ReturnBox is the box the document says its returned value fills, defs applied. A box declared on the name it
// returns, name.box = [lo, hi], is that box exactly as written (DeclaredReturnBox), not grown by the crop's boxPad:
// declaring it is how a document names its box. Otherwise it is the box the analysis carries from the primitives
// through the operations, as .box reads it (property/box.go). ok is false when the box is unbounded on some axis, or
// empty, and the caller must find a box another way. Unlike fitBox it is no fit: a declared box is taken on trust,
// and a derived one is sound.
func (e *Evaluator) ReturnBox(ctx context.Context, manifest map[string]string, code, defs string, exprs []string) (box [6]float64, ok bool, err error) {
	query := ""
	if rhs, declared := DeclaredReturnBox(code); declared && defs == "" {
		query = "(" + rhs + ")"
	} else {
		expr := exprs[0]
		if len(exprs) > 1 {
			expr = "min(" + strings.Join(exprs, ", ") + ")"
		}
		query = "(" + expr + ").box"
	}
	groups, err := e.EvalGsl(ctx, manifest, code, defs, []string{query}, [][3]float64{{0, 0, 0}})
	if err != nil {
		return box, false, err
	}
	if len(groups) != 1 || len(groups[0].Samples) != 6 {
		return box, false, nil
	}
	for k, s := range groups[0].Samples {
		if len(s) != 1 {
			return box, false, nil
		}
		v := float64(s[0].Value)
		if math.IsNaN(v) || math.Abs(v) >= returnBoxReach {
			return box, false, nil
		}
		box[k] = v
	}
	for k := 0; k < 3; k++ {
		if box[k] > box[k+3] {
			return box, false, nil
		}
	}
	return box, true, nil
}

// How far from the origin fitBox looks, and how many times it widens its search: gsl compile --bbox auto's.
const (
	fitReach   = 1024.0
	fitRescans = 8
)

// fitBox is the box the union of exprs fills, by the rule gsl compile --bbox auto follows (cmd/compile's autoBbox),
// with defs applied: Bounds over the default region, then Bounds again around what it found, the box growing until
// it no longer reaches the edge of the region searched, and two cells of room on every side. It is a fit by
// sampling, not a proof.
func (e *Evaluator) fitBox(ctx context.Context, manifest map[string]string, code, defs string, exprs []string) ([6]float64, error) {
	var box [6]float64
	expr := exprs[0]
	if len(exprs) > 1 {
		expr = "min(" + strings.Join(exprs, ", ") + ")"
	}
	found, err := e.Bounds(ctx, manifest, code, defs, expr, [6]float64{}, 64)
	if err != nil {
		return box, err
	}
	if !found.Found {
		return box, fmt.Errorf("the field is nowhere inside within %g mm of the origin (its least value %g, at %v): name the region",
			fitReach, found.MinValue, found.MinAt)
	}
	fit := [6]float64{found.BboxMin[0], found.BboxMin[1], found.BboxMin[2], found.BboxMax[0], found.BboxMax[1], found.BboxMax[2]}
	cell := found.CellSize
	for range fitRescans {
		var region [6]float64
		for k := 0; k < 3; k++ {
			grow := math.Max(fit[k+3]-fit[k], 4*cell[k])
			region[k] = math.Max(fit[k]-grow, -fitReach)
			region[k+3] = math.Min(fit[k+3]+grow, fitReach)
		}
		wider, err := e.Bounds(ctx, manifest, code, defs, expr, region, 64)
		if err != nil {
			return box, err
		}
		if !wider.Found {
			break
		}
		cell = wider.CellSize
		reaches := false
		for k := 0; k < 3; k++ {
			fit[k] = math.Min(fit[k], wider.BboxMin[k])
			fit[k+3] = math.Max(fit[k+3], wider.BboxMax[k])
			if (region[k] > -fitReach && fit[k] < region[k]+cell[k]) || (region[k+3] < fitReach && fit[k+3] > region[k+3]-cell[k]) {
				reaches = true
			}
		}
		if !reaches {
			break
		}
	}
	for k := 0; k < 3; k++ {
		pad := 2 * cell[k]
		box[k] = fit[k] - pad
		box[k+3] = fit[k+3] + pad
	}
	return box, nil
}

// vertexNormals is one unit normal per vertex, xyz: the vertex's gradient made unit, or, where the gradient is
// missing, not finite or zero, the area-weighted average of the facet normals of the triangles at the vertex.
// gradients may be nil, which takes the facet average everywhere. A vertex with neither gets +z.
func vertexNormals(gradients [][3]float32, positions []float32, indices []uint32) []float32 {
	count := len(positions) / 3
	out := make([]float32, 3*count)
	var facets [][3]float64 // built only when some vertex needs it
	facetAverage := func() [][3]float64 {
		sums := make([][3]float64, count)
		at := func(i uint32) [3]float64 {
			return [3]float64{float64(positions[3*i]), float64(positions[3*i+1]), float64(positions[3*i+2])}
		}
		for t := 0; t+2 < len(indices); t += 3 {
			ia, ib, ic := indices[t], indices[t+1], indices[t+2]
			if int(ia) >= count || int(ib) >= count || int(ic) >= count {
				continue
			}
			a, b, c := at(ia), at(ib), at(ic)
			u := [3]float64{b[0] - a[0], b[1] - a[1], b[2] - a[2]}
			v := [3]float64{c[0] - a[0], c[1] - a[1], c[2] - a[2]}
			// The cross product's length is twice the triangle's area, which is the weight.
			n := [3]float64{u[1]*v[2] - u[2]*v[1], u[2]*v[0] - u[0]*v[2], u[0]*v[1] - u[1]*v[0]}
			for _, i := range [3]uint32{ia, ib, ic} {
				for k := 0; k < 3; k++ {
					sums[i][k] += n[k]
				}
			}
		}
		return sums
	}
	unit := func(n [3]float64) ([3]float64, bool) {
		length := math.Sqrt(n[0]*n[0] + n[1]*n[1] + n[2]*n[2])
		if math.IsNaN(length) || math.IsInf(length, 0) || length == 0 {
			return n, false
		}
		return [3]float64{n[0] / length, n[1] / length, n[2] / length}, true
	}
	for v := 0; v < count; v++ {
		var n [3]float64
		ok := false
		if v < len(gradients) {
			n, ok = unit([3]float64{float64(gradients[v][0]), float64(gradients[v][1]), float64(gradients[v][2])})
		}
		if !ok {
			if facets == nil {
				facets = facetAverage()
			}
			if n, ok = unit(facets[v]); !ok {
				n = [3]float64{0, 0, 1}
			}
		}
		out[3*v], out[3*v+1], out[3*v+2] = float32(n[0]), float32(n[1]), float32(n[2])
	}
	return out
}
